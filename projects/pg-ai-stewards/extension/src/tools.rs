//! Tool dispatch + reference resolution for the bgworker.
//!
//! Two `pub(crate)` entry points called from the bgworker dispatch loop:
//! - `resolve_ref` — Phase 2.2 gospel-engine reference resolver
//! - `tool_dispatch` — Phase 1.6 tool_calls executor (sql_fn + http kinds)
//!
//! Extracted from lib.rs as Phase 3c.3.6 v3 (2026-05-08). Per the
//! pgrx-rust skill, plain `mod tools;` in lib.rs is sufficient — no
//! `pub use` re-export needed for `pub(crate)` items.

use crate::providers::GOSPEL_ENGINE_CONFIG;
use crate::types::WorkOutcome;
use pgrx::bgworkers::*;
use pgrx::prelude::*;

// ---------------------------------------------------------------------------
// Phase 1.6: tool_dispatch — execute the tool_calls from a parent
// assistant message and produce N role='tool' replies for phase 3.
//
// Three execute_target kinds are wired up:
//   sql_fn:    SELECT <schema>.<name>($1::jsonb)::text  (sync)
//   http:      POST args as JSON body, response.text() returned as-is  (sync)
//   mcp_proxy: enqueue child mcp_proxy work_queue row, bridge resolves  (async — Phase 3e.2.b)
//
// Tool errors are NOT returned as Err(). Each per-call failure is
// captured into the tool reply content as {"error": "..."}, so the
// model sees what went wrong and can recover. Only structural
// failures (no parent message, malformed payload) raise Err.
//
// Phase 3e.2.b: when at least one tool routes to mcp_proxy, the
// dispatcher returns WorkOutcome::WaitingForTools instead of
// ToolsDispatched. The bgworker write phase then transitions the
// tool_dispatch row to 'waiting_for_tools' status and stores the
// (resolved, pending) split in result jsonb. The SQL completion
// function `stewards.tool_dispatch_complete_waiting()` (called from
// the bgworker tick loop) promotes the row to 'done' once all
// pending children resolve, then runs the equivalent of Phase 3.
// ---------------------------------------------------------------------------

/// Per-call result inside `tool_dispatch`. Sync tools (sql_fn / http)
/// return their content directly; mcp_proxy returns a child work_queue
/// id that the completion pass will join on later.
pub(crate) enum ToolReply {
    Sync(String),
    Async {
        child_work_id: i64,
    },
}

/// Phase 2.2 — resolve a single gospel-engine reference like
/// "Mosiah 18:8". GETs {GOSPEL_ENGINE_URL}/api/get?ref=<urlencoded>
/// with the bearer token from GOSPEL_ENGINE_TOKEN.
///
/// Soft-error semantics: a 404 from gospel-engine becomes a Resolved
/// row with content=NULL and error="not found", NOT an Err. This way
/// the work item completes successfully (no retry storms on
/// genuinely-missing refs) and the cache row records the negative
/// result. Only network failures and 5xx responses raise Err so the
/// bgworker's retry policy can take over.
pub(crate) fn resolve_ref(payload: &serde_json::Value) -> Result<WorkOutcome, String> {
    let ref_str = payload
        .get("ref")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "payload.ref missing".to_string())?
        .to_string();

    let cfg = GOSPEL_ENGINE_CONFIG
        .get()
        .cloned()
        .unwrap_or_default();

    let Some(base) = cfg.url else {
        // Cache the failure so we don't keep retrying with no config.
        return Ok(WorkOutcome::Resolved {
            ref_str,
            content: None,
            error: Some("GOSPEL_ENGINE_URL not set".to_string()),
        });
    };

    // Build URL with manual encoding of the ref (spaces -> %20, colon
    // is fine in a query string but we percent-encode '&' which
    // appears in "D&C 88:67"). reqwest's Client.get(url) does NOT
    // re-encode the path/query, so we encode here.
    let encoded = url_encode_query_value(&ref_str);
    let url = format!("{}/api/get?ref={}", base, encoded);

    let client = reqwest::blocking::Client::builder()
        .timeout(std::time::Duration::from_secs(30))
        .build()
        .map_err(|e| format!("http client build: {}", e))?;

    let mut req = client.get(&url);
    if let Some(tok) = &cfg.token {
        req = req.bearer_auth(tok);
    }
    let resp = req
        .send()
        .map_err(|e| format!("GET {}: {}", url, e))?;
    let status = resp.status();
    let body = resp.text().unwrap_or_default();

    if status == reqwest::StatusCode::NOT_FOUND {
        return Ok(WorkOutcome::Resolved {
            ref_str,
            content: None,
            error: Some(format!("not found: {}", body.trim())),
        });
    }
    if !status.is_success() {
        // 5xx, 401, etc. — surface as Err so the work_queue marks
        // 'error' and ops can see the failure mode.
        return Err(format!("gospel-engine HTTP {}: {}", status, body));
    }

    let parsed: serde_json::Value = serde_json::from_str(&body)
        .map_err(|e| format!("decode gospel-engine response: {} (body={})", e, body))?;

    Ok(WorkOutcome::Resolved {
        ref_str,
        content: Some(parsed),
        error: None,
    })
}

/// Minimal RFC 3986 query-value percent encoder. Encodes everything
/// outside ALPHA / DIGIT / "-._~" plus a few we know are safe in
/// gospel-engine refs (':' is safe in a query). Avoids pulling
/// percent-encoding crate for one call site.
fn url_encode_query_value(s: &str) -> String {
    let mut out = String::with_capacity(s.len() + 8);
    for b in s.bytes() {
        let safe = b.is_ascii_alphanumeric()
            || matches!(b, b'-' | b'.' | b'_' | b'~' | b':');
        if safe {
            out.push(b as char);
        } else if b == b' ' {
            out.push_str("%20");
        } else {
            out.push_str(&format!("%{:02X}", b));
        }
    }
    out
}

pub(crate) fn tool_dispatch(payload: &serde_json::Value) -> Result<WorkOutcome, String> {
    let parent_work_id = payload
        .get("parent_work_id")
        .and_then(|v| v.as_i64())
        .ok_or_else(|| "payload.parent_work_id missing".to_string())?;
    let session_id = payload
        .get("session_id")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "payload.session_id missing".to_string())?
        .to_string();
    let agent_family = payload
        .get("agent_family")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "payload.agent_family missing".to_string())?
        .to_string();
    let model = payload
        .get("model")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "payload.model missing".to_string())?
        .to_string();

    // Tx A: read the parent assistant message's tool_calls and grab
    // the (already fetched) tool_def execute_target for each name.
    // We do this in one tx so the dispatch loop below sees a coherent
    // snapshot. The dispatch itself runs OUTSIDE this tx so HTTP
    // tools don't hold row locks.
    type Prep = Vec<(String, String, serde_json::Value, serde_json::Value)>;
    let prep: Result<Option<Prep>, pgrx::spi::Error> =
        BackgroundWorker::transaction(|| {
            Spi::connect(|client| {
                // Find the assistant message produced by the parent
                // chat work item. We look it up by parent_work_id
                // (set in phase 3 of chat).
                let rows = client.select(
                    "SELECT tool_calls FROM stewards.messages \
                     WHERE parent_work_id = $1 AND role = 'assistant' \
                     ORDER BY id DESC LIMIT 1",
                    Some(1),
                    &[parent_work_id.into()],
                )?;
                let mut iter = rows.into_iter();
                let Some(row) = iter.next() else {
                    return Ok(None);
                };
                let tool_calls: pgrx::JsonB = row.get(1)?.expect("tool_calls non-null");
                let tcs = tool_calls.0.as_array().cloned().unwrap_or_default();

                // For each tool_call, look up the tool_def. Build
                // (tool_call_id, name, args_jsonb, target_jsonb).
                let mut prepped: Prep = Vec::with_capacity(tcs.len());
                for tc in tcs {
                    let tc_id: String = tc.get("id")
                        .and_then(|v| v.as_str())
                        .unwrap_or("unknown")
                        .to_string();
                    let name: String = tc.get("function")
                        .and_then(|f| f.get("name"))
                        .and_then(|v| v.as_str())
                        .unwrap_or("")
                        .to_string();
                    // OpenAI returns function.arguments as a STRING
                    // (JSON-encoded). Decode here so dispatch sees a
                    // jsonb. If the model emits malformed JSON, fall
                    // back to a sentinel so dispatch can still run
                    // and the tool can complain meaningfully.
                    let args_str = tc.get("function")
                        .and_then(|f| f.get("arguments"))
                        .and_then(|v| v.as_str())
                        .unwrap_or("{}");
                    let args = serde_json::from_str::<serde_json::Value>(args_str)
                        .unwrap_or_else(|_| serde_json::json!({
                            "_decode_error": "function.arguments was not valid JSON",
                            "_raw": args_str,
                        }));

                    // Look up tool_def. If absent, store a sentinel
                    // target so dispatch can return a structured
                    // error reply (the model needs to know).
                    let target_rows = client.select(
                        "SELECT execute_target FROM stewards.tool_defs \
                         WHERE name = $1 AND active",
                        Some(1),
                        &[name.clone().into()],
                    )?;
                    let target = match target_rows.into_iter().next() {
                        Some(r) => r.get::<pgrx::JsonB>(1)?.map(|j| j.0)
                            .unwrap_or(serde_json::json!({"kind":"missing"})),
                        None => serde_json::json!({"kind":"missing"}),
                    };
                    prepped.push((tc_id, name, args, target));
                }
                Ok(Some(prepped))
            })
        });

    let prepped = prep
        .map_err(|e| format!("tool_dispatch prep tx: {}", e))?
        .ok_or_else(|| format!(
            "no assistant message found for parent_work_id={}", parent_work_id
        ))?;

    // Phase 2 (no tx): execute each tool. Sync tools resolve here;
    // async (mcp_proxy) tools enqueue a child row and we collect the
    // child id for the completion pass.
    let mut resolved: Vec<(String, String, String)> = Vec::new();
    let mut pending: Vec<(String, String, i64)> = Vec::new();
    for (tc_id, name, args, target) in prepped.into_iter() {
        match exec_one_tool(&name, &args, &target, &session_id) {
            Ok(ToolReply::Sync(content)) => {
                resolved.push((tc_id, name, content));
            }
            Ok(ToolReply::Async { child_work_id }) => {
                pending.push((tc_id, name, child_work_id));
            }
            Err(e) => {
                pgrx::log!("stewards: tool '{}' failed: {}", name, e);
                resolved.push((
                    tc_id,
                    name,
                    serde_json::json!({"error": e}).to_string(),
                ));
            }
        }
    }

    if pending.is_empty() {
        // All-sync path: behaves identically to pre-3e.2.b.
        Ok(WorkOutcome::ToolsDispatched {
            parent_work_id,
            session_id,
            agent_family,
            model,
            tool_messages: resolved,
        })
    } else {
        Ok(WorkOutcome::WaitingForTools {
            parent_work_id,
            session_id,
            agent_family,
            model,
            resolved,
            pending,
        })
    }
}

/// Dispatch a single tool call. Returns either a sync content string
/// (model sees it directly) or an async child work_queue id (the
/// completion pass will join on it once the bridge writes the
/// result back). SHOULD be JSON-parseable for sync, but freeform
/// strings work since the LLM is told this is a tool reply.
fn exec_one_tool(
    name: &str,
    args: &serde_json::Value,
    target: &serde_json::Value,
    session_id: &str,
) -> Result<ToolReply, String> {
    let kind = target.get("kind")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "tool execute_target.kind missing".to_string())?;
    match kind {
        "sql_fn"    => exec_sql_fn_tool(target, args, session_id).map(ToolReply::Sync),
        "http"      => exec_http_tool(target, args).map(ToolReply::Sync),
        "mcp_proxy" => exec_mcp_proxy_tool(target, args),
        "missing"   => Err(format!("tool '{}' is not registered or inactive", name)),
        other       => Err(format!("unsupported tool kind: {}", other)),
    }
}

/// MCP proxy tool. Target shape:
///   {"kind":"mcp_proxy", "server":"<name>", "tool":"<tool_name>"}
/// Calls `stewards.mcp_proxy_enqueue(server, tool, args, NULL)` which
/// inserts a child work_queue row of kind='mcp_proxy' and NOTIFY's
/// `stewards_mcp_proxy`. Bridge daemon claims the row and writes the
/// result back.
///
/// Errors here (server not registered, server not enabled, SPI
/// failure) raise Err — the dispatcher captures them into a
/// {"error": "..."} sync reply rather than blocking.
fn exec_mcp_proxy_tool(
    target: &serde_json::Value,
    args: &serde_json::Value,
) -> Result<ToolReply, String> {
    let server = target.get("server")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "mcp_proxy target.server missing".to_string())?;
    let tool = target.get("tool")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "mcp_proxy target.tool missing".to_string())?;

    let args_value = args.clone();

    let result: Result<Option<i64>, pgrx::spi::Error> =
        BackgroundWorker::transaction(|| {
            Spi::connect(|client| {
                let row = client.select(
                    "SELECT stewards.mcp_proxy_enqueue($1, $2, $3, NULL)",
                    Some(1),
                    &[
                        server.into(),
                        tool.into(),
                        pgrx::JsonB(args_value).into(),
                    ],
                )?;
                Ok(row.into_iter().next()
                    .and_then(|r| r.get::<i64>(1).ok().flatten()))
            })
        });

    match result {
        Ok(Some(id)) => Ok(ToolReply::Async { child_work_id: id }),
        Ok(None) => Err(format!(
            "mcp_proxy_enqueue({}, {}) returned NULL", server, tool
        )),
        Err(e) => Err(format!(
            "mcp_proxy_enqueue({}, {}) SPI: {}", server, tool, e
        )),
    }
}

/// SQL function tool. Convention: the target SQL fn has signature
///   fn(p_args jsonb) RETURNS jsonb
/// Wrapped by stewards.<name>_tool functions for this convention.
fn exec_sql_fn_tool(
    target: &serde_json::Value,
    args: &serde_json::Value,
    session_id: &str,
) -> Result<String, String> {
    let schema = target.get("schema")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "sql_fn target.schema missing".to_string())?;
    let fn_name = target.get("name")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "sql_fn target.name missing".to_string())?;

    // Identifier-safe-ish guard: schema and name must match a
    // simple identifier pattern. This is a defense-in-depth measure
    // because we're string-formatting these into SQL. The CHECK
    // constraint on tool_defs.name already enforces this at insert,
    // but the schema field is free-form.
    let safe = |s: &str| s.chars().all(|c| c.is_ascii_alphanumeric() || c == '_');
    if !safe(schema) || !safe(fn_name) {
        return Err(format!("unsafe identifier in sql_fn target: {}.{}", schema, fn_name));
    }

    let sql = format!("SELECT {}.{}($1)::text", schema, fn_name);
    // JsonB doesn't impl Clone in pgrx 0.18; build the value once and
    // wrap it in Rc so multiple PgTryBuilder retries (if any) could
    // share it. We currently only call it once, so a fresh build per
    // entry is fine — just don't try to .clone() it later.
    //
    // CT2.3: inject the dispatch session_id so session-scoped sql_fn tools
    // (the context-management levers) can resolve [ctx:handle] → message_id
    // within the agent's own session. Existing sql_fn tools ignore the extra
    // key (they read their named args), so this is backward-compatible. Only
    // injected when args is an object and the key is absent.
    let mut args_value = args.clone();
    if let Some(obj) = args_value.as_object_mut() {
        obj.entry("_session_id".to_string())
            .or_insert_with(|| serde_json::Value::String(session_id.to_string()));
    }

    // Pre-flight: does the function exist with a jsonb signature?
    // PG ereports on missing function would otherwise reach the
    // bgworker via longjmp; PgTryBuilder is supposed to catch them
    // but in pgrx 0.18 + BackgroundWorker::transaction the longjmp
    // path empirically still kills the worker (verified in
    // testing — see verify-loop.sql). The cheapest defense is to
    // check pg_proc first and never trigger the ereport.
    let exists: Result<bool, pgrx::spi::Error> =
        BackgroundWorker::transaction(|| {
            Spi::connect(|client| {
                let row = client.select(
                    "SELECT EXISTS ( \
                        SELECT 1 FROM pg_proc p \
                        JOIN pg_namespace n ON p.pronamespace = n.oid \
                        WHERE n.nspname = $1 AND p.proname = $2 \
                          AND pg_get_function_arguments(p.oid) ILIKE '%jsonb%' \
                     )",
                    Some(1),
                    &[schema.into(), fn_name.into()],
                )?;
                Ok(row.into_iter().next()
                    .and_then(|r| r.get::<bool>(1).ok().flatten())
                    .unwrap_or(false))
            })
        });
    match exists {
        Ok(true) => { /* fall through */ }
        Ok(false) => return Err(format!(
            "sql_fn target {}.{}(jsonb) does not exist", schema, fn_name)),
        Err(e) => return Err(format!("sql_fn pre-flight: {}", e)),
    }

    use pgrx::PgTryBuilder;

    // PgTryBuilder wraps PG_TRY/PG_CATCH. It catches the ereport,
    // unwinds the implicit subtx pgrx opened around BackgroundWorker
    // ::transaction, and returns a recoverable Err we match on here.
    // The bgworker survives.
    //
    // We do NOT use SAVEPOINT here. SAVEPOINT requires an explicit
    // BEGIN, but BackgroundWorker::transaction opens an implicit one
    // — trying to issue SAVEPOINT inside it errors with
    // "SAVEPOINT can only be used in transaction blocks", which
    // ironically broke even the success path. PG_TRY handles the
    // unwind without our help.
    let result: Result<Option<String>, String> = PgTryBuilder::new(|| {
        let outer: Result<Option<String>, pgrx::spi::Error> =
            BackgroundWorker::transaction(|| {
                Spi::connect(|client| {
                    let row = client.select(
                        &sql, Some(1),
                        &[pgrx::JsonB(args_value.clone()).into()]
                    )?;
                    let mut iter = row.into_iter();
                    match iter.next() {
                        Some(r) => r.get::<String>(1),
                        None => Ok(None),
                    }
                })
            });
        outer.map_err(|e| format!("spi: {}", e))
    })
    .catch_others(|cause| {
        Err(format!("postgres error: {:?}", cause))
    })
    .execute();

    match result {
        Ok(Some(s)) => Ok(s),
        Ok(None) => Ok("null".to_string()),
        Err(e) => Err(format!("sql_fn {}.{}: {}", schema, fn_name, e)),
    }
}

/// HTTP tool. Target shape:
///   {"kind":"http", "method":"POST", "url":"...", "headers":{...}}
/// Method defaults to POST. Args are sent as JSON body. Response
/// body is returned as-is (assumed JSON; freeform strings also OK).
fn exec_http_tool(
    target: &serde_json::Value,
    args: &serde_json::Value,
) -> Result<String, String> {
    let url = target.get("url")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "http target.url missing".to_string())?;
    let method = target.get("method")
        .and_then(|v| v.as_str())
        .unwrap_or("POST")
        .to_uppercase();

    let client = reqwest::blocking::Client::builder()
        .timeout(std::time::Duration::from_secs(60))
        .build()
        .map_err(|e| format!("http client build: {}", e))?;

    let mut req = match method.as_str() {
        "POST" => client.post(url).json(args),
        "GET"  => client.get(url),
        other  => return Err(format!("unsupported http method: {}", other)),
    };

    if let Some(headers) = target.get("headers").and_then(|v| v.as_object()) {
        for (k, v) in headers {
            if let Some(vs) = v.as_str() {
                req = req.header(k.as_str(), vs);
            }
        }
    }

    let resp = req.send().map_err(|e| format!("POST {}: {}", url, e))?;
    let status = resp.status();
    let body = resp.text().unwrap_or_default();
    if !status.is_success() {
        return Err(format!("http {} {}: {}", method, status, body));
    }
    Ok(body)
}
