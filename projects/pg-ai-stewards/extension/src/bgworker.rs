//! Bgworker entry point, tick loop, and provider dispatch helpers.
//!
//! Owns:
//! - `_PG_init` — registers the background worker at postmaster startup
//! - `check_watchman_schedule` — 60s scheduler tick decisions
//! - `process_one_pending` — claim + run + write loop body
//! - `dispatch` / `embed` / `chat` — per-kind work_queue handlers
//!
//! Per the pgrx-rust skill, `_PG_init` works in any submodule — Postgres
//! finds the symbol at `dlopen` time via C linkage. plain `mod bgworker;`
//! in lib.rs is enough.
//!
//! Extracted from lib.rs as Phase 3c.3.6 v4 (2026-05-08).

use crate::providers::{
    GospelEngineConfig, ProviderRegistry, GOSPEL_ENGINE_CONFIG, PROVIDER_REGISTRY,
};
use crate::tools::{resolve_ref, tool_dispatch};
use crate::types::WorkOutcome;
use pgrx::bgworkers::*;
use pgrx::prelude::*;
use std::time::{Duration, Instant};

// ---------------------------------------------------------------------------
// Bgworker registration
// ---------------------------------------------------------------------------

#[pg_guard]
pub extern "C-unwind" fn _PG_init() {
    // Only register the bgworker when we are actually being preloaded
    // via shared_preload_libraries. Otherwise `CREATE EXTENSION` in a
    // database that doesn't preload us would fail.
    if unsafe { !pgrx::pg_sys::process_shared_preload_libraries_in_progress } {
        return;
    }

    // Parse provider registry once, in the postmaster. All backends
    // (and the bgworker) inherit it via fork() copy-on-write, so
    // `stewards.providers_loaded()` works from any psql session and
    // the worker doesn't need to re-parse.
    let registry = ProviderRegistry::from_env();
    pgrx::log!(
        "stewards: postmaster loaded {} provider(s) from env",
        registry.providers.len()
    );
    for p in &registry.providers {
        pgrx::log!(
            "stewards:   provider '{}' kind={} base_url={} model={} api_key={}",
            p.name,
            p.kind,
            p.base_url,
            p.default_model,
            if p.api_key.is_some() { "yes" } else { "no" }
        );
    }
    let _ = PROVIDER_REGISTRY.set(registry);

    // Phase 2.2 — read gospel-engine config from env. Trim trailing
    // slashes from URL so {url}/api/get?... composes cleanly.
    let ge_cfg = GospelEngineConfig {
        url: std::env::var("GOSPEL_ENGINE_URL")
            .ok()
            .map(|s| s.trim_end_matches('/').to_string())
            .filter(|s| !s.is_empty()),
        token: std::env::var("GOSPEL_ENGINE_TOKEN")
            .ok()
            .filter(|s| !s.is_empty()),
    };
    pgrx::log!(
        "stewards: gospel-engine url={} token={}",
        ge_cfg.url.as_deref().unwrap_or("<unset>"),
        if ge_cfg.token.is_some() { "yes" } else { "no" }
    );
    let _ = GOSPEL_ENGINE_CONFIG.set(ge_cfg);

    // Phase 3e.2.a — register N dispatcher workers. Each worker runs
    // the same tick loop but claims rows independently via FOR UPDATE
    // SKIP LOCKED, so concurrent draining is safe. The first worker
    // (index 0) is also responsible for once-per-postmaster startup
    // chores (stale-claim reaper) and the periodic Watchman scheduler
    // tick — those would race or duplicate work if all N ran them.
    //
    // Worker count is configurable via STEWARDS_DISPATCHER_WORKERS,
    // defaulting to 4. Cap at 16 to keep postmaster registration tidy.
    let worker_count: usize = std::env::var("STEWARDS_DISPATCHER_WORKERS")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(4)
        .min(16)
        .max(1);
    pgrx::log!("stewards: registering {} dispatcher worker(s)", worker_count);
    for i in 0..worker_count {
        BackgroundWorkerBuilder::new(&format!("pg_ai_stewards dispatcher #{}", i))
            .set_function("stewards_dispatcher_main")
            .set_library("pg_ai_stewards")
            .enable_spi_access()
            .set_restart_time(Some(Duration::from_secs(5)))
            .set_argument(Some(pg_sys::Datum::from(i as u64)))
            .load();
    }
}

/// Worker entry point. Polls `stewards.work_queue` every 500ms,
/// claims one row, runs the stub provider, writes the result back.
///
/// `arg` carries the worker index assigned at registration time
/// (0..N). Worker 0 is the "leader" — it owns the stale-claim reaper
/// and the Watchman scheduler tick, both of which must not run from
/// every worker simultaneously. All workers share the claim loop
/// (FOR UPDATE SKIP LOCKED makes that safe).
#[pg_guard]
#[unsafe(no_mangle)]
pub extern "C-unwind" fn stewards_dispatcher_main(arg: pg_sys::Datum) {
    let worker_index: usize = arg.value() as usize;
    let is_leader: bool = worker_index == 0;

    BackgroundWorker::attach_signal_handlers(
        SignalWakeFlags::SIGHUP | SignalWakeFlags::SIGTERM,
    );

    let dbname = std::env::var("POSTGRES_DB").unwrap_or_else(|_| "stewards".to_string());
    BackgroundWorker::connect_worker_to_spi(Some(&dbname), None);

    let provider_count = PROVIDER_REGISTRY.get().map(|r| r.providers.len()).unwrap_or(0);
    pgrx::log!(
        "stewards: bgworker #{} entering poll loop (500ms tick); leader={}; {} provider(s) inherited from postmaster",
        worker_index, is_leader, provider_count
    );

    // Stale-claim reaper: any row left in 'in_progress' by a previous
    // bgworker crash is unreachable \u2014 we never reclaim our own
    // claims (that would risk double-side-effects). Mark them errored
    // at startup with a clear message so the caller knows what
    // happened and can decide whether to re-enqueue.
    //
    // For tool_dispatch rows specifically, also call
    // synthesize_tool_failure: write the missing role='tool' replies
    // and enqueue a continuation chat. Otherwise the parent chat's
    // loop stalls forever waiting for tool replies that will never
    // come (Phase 1.6.1).
    //
    // Leader-only (worker 0): otherwise N workers race to reap and
    // synthesize, producing duplicate continuation chats.
    if is_leader {
    let _ = BackgroundWorker::transaction(|| {
        Spi::connect_mut(|client| {
            // Pull the rows we're about to reap so we can synthesize
            // continuations for tool_dispatch ones.
            //
            // Phase 3e.2.b: skip kind='mcp_proxy'. Those rows belong
            // to the bridge daemon's lifecycle, not the bgworker's.
            // The bridge has its own startup reaper for stale
            // mcp_proxy rows it left in_progress at last shutdown.
            let stale_rows: Vec<(i64, String, String, serde_json::Value)> = {
                let rows = client.select(
                    "SELECT id, kind, provider, payload \
                     FROM stewards.work_queue \
                     WHERE status = 'in_progress' \
                       AND kind <> 'mcp_proxy'",
                    None, &[],
                )?;
                rows.into_iter().filter_map(|r| {
                    let id: i64 = r.get(1).ok()??;
                    let kind: String = r.get(2).ok()??;
                    let provider: String = r.get(3).ok()??;
                    let payload: pgrx::JsonB = r.get(4).ok()??;
                    Some((id, kind, provider, payload.0))
                }).collect()
            };

            for (id, kind, provider, payload) in &stale_rows {
                if kind == "tool_dispatch" {
                    if let (Some(parent), Some(session), Some(family), Some(model)) = (
                        payload.get("parent_work_id").and_then(|v| v.as_i64()),
                        payload.get("session_id").and_then(|v| v.as_str()),
                        payload.get("agent_family").and_then(|v| v.as_str()),
                        payload.get("model").and_then(|v| v.as_str()),
                    ) {
                        let synth = client.select(
                            "SELECT stewards.synthesize_tool_failure($1, $2, $3, $4, $5, $6)",
                            Some(1),
                            &[
                                parent.into(),
                                family.to_string().into(),
                                model.to_string().into(),
                                session.to_string().into(),
                                provider.to_string().into(),
                                format!(
                                    "bgworker crashed mid-dispatch on work_item id={}; loop continued via reaper",
                                    id
                                ).into(),
                            ],
                        );
                        if let Err(e) = synth {
                            pgrx::log!(
                                "stewards: reaper synthesize failed for id={}: {}",
                                id, e
                            );
                        } else {
                            pgrx::log!(
                                "stewards: reaper synthesized tool failure for tool_dispatch id={} (parent={})",
                                id, parent
                            );
                        }
                    }
                }
            }

            client.update(
                "UPDATE stewards.work_queue \
                 SET status = 'error', \
                     error  = coalesce(error, '') \
                              || 'bgworker crashed before completion (stale in_progress reaped at startup)', \
                     done_at = now() \
                 WHERE status = 'in_progress' \
                   AND kind <> 'mcp_proxy'",
                None, &[]
            )?;
            Ok::<(), pgrx::spi::Error>(())
        })
    });
    }

    // Phase 2.7b.2 — Watchman scheduler tick.
    //
    // The bgworker drains the work_queue every 500ms. Independently
    // (and much more rarely), it checks whether a Watchman pass should
    // fire. Decision logic lives entirely in SQL (stewards.watchman_
    // should_fire); Rust just calls it on a 60s tick and dispatches.
    //
    // last_sched=None on entry forces an immediate check on first tick,
    // useful when a fresh bgworker comes up after being down for a
    // while (don't make the user wait 60s for the first decision).
    let mut last_sched: Option<Instant> = None;
    const SCHED_INTERVAL: Duration = Duration::from_secs(60);

    // Phase 4d — Steward tick.
    //
    // Same pattern as the Watchman scheduler tick: independent of the
    // 500ms work_queue drain, the leader periodically calls
    // stewards.steward_tick() which walks failed work_items applying
    // cost-cap + breaker + diagnosis + escalation logic and dispatching
    // retries. 30s tick is chosen to balance retry latency against
    // log noise (the function returns 0 most of the time).
    //
    // Leader-only because steward_tick uses FOR UPDATE SKIP LOCKED
    // internally — multiple workers calling it would be SAFE but would
    // double the SQL traffic without throughput gain (the lock-skip
    // means each item is processed once anyway).
    let mut last_steward: Option<Instant> = None;
    const STEWARD_INTERVAL: Duration = Duration::from_secs(30);

    while BackgroundWorker::wait_latch(Some(Duration::from_millis(500))) {
        if BackgroundWorker::sighup_received() {
            pgrx::log!("stewards: SIGHUP received");
        }

        // Drain whatever's pending. process_one_pending() returns
        // false when the queue is empty, so the loop bounds itself.
        let mut processed = 0u32;
        while process_one_pending() {
            processed += 1;
            // Cap a single tick to avoid starving signal handling.
            if processed >= 16 {
                break;
            }
        }

        // Phase 3e.2.b — async-fan-out completion pass. Promotes any
        // tool_dispatch row whose mcp_proxy children have all
        // resolved out of 'waiting_for_tools' into 'done', writing
        // tool messages and enqueueing the continuation chat. All
        // workers run this (FOR UPDATE SKIP LOCKED inside the SQL
        // function keeps them from racing) so tool reply latency
        // doesn't hinge on a single leader.
        complete_waiting_tool_dispatches();

        // Watchman scheduler tick. Cheap when no trigger is hot
        // (single SPI call returning NULL). Two SPI calls when a
        // trigger fires (decide → enqueue chats). Leader-only —
        // running it from every worker would multiply the firing
        // decisions and risk duplicate passes despite cooldown logic.
        if is_leader && last_sched.map_or(true, |t| t.elapsed() >= SCHED_INTERVAL) {
            last_sched = Some(Instant::now());
            check_watchman_schedule();
        }

        // Phase 4d — Steward tick. Walks failed work_items that need
        // retry decisions. Returns count of actions taken (cost-cap
        // quarantine, breaker defer, queue-for-opus, retry dispatch,
        // or tick_error). Leader-only.
        if is_leader && last_steward.map_or(true, |t| t.elapsed() >= STEWARD_INTERVAL) {
            last_steward = Some(Instant::now());
            check_steward_tick();
        }
    }

    pgrx::log!("stewards: bgworker #{} received SIGTERM, exiting", worker_index);
}

/// Phase 3e.2.b — completion pass for waiting tool_dispatch rows.
///
/// Calls `stewards.tool_dispatch_complete_waiting()` which scans
/// `kind='tool_dispatch' AND status='waiting_for_tools'` rows, joins
/// each one's pending children to check whether they've all resolved,
/// and (if so) inserts the tool messages, enqueues the continuation
/// chat, and promotes the parent to status='done'. Concurrent-safe
/// via FOR UPDATE SKIP LOCKED inside the function.
///
/// Errors are logged but never propagated — a transient SPI failure
/// shouldn't kill the bgworker. The next tick retries.
fn complete_waiting_tool_dispatches() {
    let result: Result<Option<i32>, pgrx::spi::Error> =
        BackgroundWorker::transaction(|| {
            Spi::connect_mut(|client| {
                let row = client.select(
                    "SELECT stewards.tool_dispatch_complete_waiting()",
                    Some(1), &[],
                )?;
                let n: Option<i32> = row.into_iter().next()
                    .and_then(|r| r.get(1).ok().flatten());
                Ok::<Option<i32>, pgrx::spi::Error>(n)
            })
        });

    match result {
        Ok(Some(n)) if n > 0 => {
            pgrx::log!("stewards: completed {} waiting tool_dispatch row(s)", n);
        }
        Ok(_) => {
            // Silent on zero — runs every tick, would flood the log.
        }
        Err(e) => {
            pgrx::log!("stewards: tool_dispatch completion pass errored: {}", e);
        }
    }
}

/// Phase 2.7b.2 — invoke the Watchman scheduler decision function.
///
/// Calls `stewards.watchman_scheduler_fire()` which itself calls
/// `watchman_should_fire()` and (if non-NULL) `watchman_pass_start()`.
/// Always logs the outcome:
///   - `pass_id != NULL` → a pass was started
///   - `pass_id == NULL` → either disabled, in cooldown, or no trigger
///
/// Errors here are swallowed (logged only) so a transient SPI failure
/// doesn't take down the bgworker. The next tick will try again.
fn check_watchman_schedule() {
    // Use connect_mut even though our SPI client only does a SELECT —
    // the SQL function it invokes (watchman_scheduler_fire) does
    // INSERTs/UPDATEs internally, and a read-only SPI context would
    // block those. Mirrors process_one_pending() and the reaper.
    let result: Result<Option<String>, pgrx::spi::Error> =
        BackgroundWorker::transaction(|| {
            Spi::connect_mut(|client| {
                let row = client.select(
                    "SELECT stewards.watchman_scheduler_fire()",
                    Some(1), &[],
                )?;
                let pass_id: Option<String> = row.into_iter().next()
                    .and_then(|r| r.get(1).ok().flatten());
                Ok::<Option<String>, pgrx::spi::Error>(pass_id)
            })
        });

    match result {
        Ok(Some(pass_id)) => {
            pgrx::log!(
                "stewards: scheduler fired Watchman pass: {}",
                pass_id
            );
        }
        Ok(None) => {
            // No-op (no trigger, disabled, in cooldown). Don't log
            // every 60 seconds — that floods the postmaster log.
        }
        Err(e) => {
            pgrx::log!("stewards: scheduler check errored: {}", e);
        }
    }
}

/// Phase 4d — invoke the steward tick.
///
/// Calls `stewards.steward_tick()` which walks failed work_items and
/// applies cost-cap + breaker + diagnosis + escalation logic, then
/// dispatches retries via work_item_dispatch_stage. Returns count of
/// actions taken in this tick (0 = no failed work_items needed
/// attention). Errors swallowed — next tick retries.
fn check_steward_tick() {
    let result: Result<Option<i32>, pgrx::spi::Error> =
        BackgroundWorker::transaction(|| {
            Spi::connect_mut(|client| {
                let row = client.select(
                    "SELECT stewards.steward_tick()",
                    Some(1), &[],
                )?;
                let n: Option<i32> = row.into_iter().next()
                    .and_then(|r| r.get(1).ok().flatten());
                Ok::<Option<i32>, pgrx::spi::Error>(n)
            })
        });

    match result {
        Ok(Some(n)) if n > 0 => {
            pgrx::log!("stewards: steward_tick processed {} action(s)", n);
        }
        Ok(_) => {
            // Silent on zero — runs every 30s, would flood the log.
        }
        Err(e) => {
            pgrx::log!("stewards: steward_tick errored: {}", e);
        }
    }
}

/// Try to claim and process exactly one pending row. Returns true if
/// a row was processed (caller may want to immediately try again),
/// false if the queue was empty.
///
/// The work happens in three phases so we don't hold a row lock
/// across a slow HTTP call (LM Studio first-request model load can
/// be 30s+):
///
///   1. Tx A: claim oldest pending row, mark `in_progress`. Commit.
///   2. No tx: dispatch by kind, possibly making HTTP calls.
///   3. Tx B: write result or error, `NOTIFY stewards_done`. Commit.
fn process_one_pending() -> bool {
    // ----- Phase 1: claim -----
    let claim: Result<Option<(i64, String, String, serde_json::Value)>, pgrx::spi::Error> =
        BackgroundWorker::transaction(|| {
            Spi::connect_mut(|client| {
                // Phase 3e.2.b: bgworker explicitly skips kind='mcp_proxy'
                // rows. Those are owned by the bridge daemon
                // (cmd/stewards-mcp/bridge.go `bridge run`), which uses
                // the same SKIP LOCKED claim against this queue but
                // filters TO kind='mcp_proxy'. The two sides partition
                // by kind without coordinating beyond the row lock.
                let claimed = client.update(
                    "WITH next AS ( \
                         SELECT id FROM stewards.work_queue \
                         WHERE status = 'pending' AND kind <> 'mcp_proxy' \
                         ORDER BY created_at \
                         FOR UPDATE SKIP LOCKED \
                         LIMIT 1 \
                     ) \
                     UPDATE stewards.work_queue q \
                     SET status = 'in_progress', claimed_at = now() \
                     FROM next \
                     WHERE q.id = next.id \
                     RETURNING q.id, q.kind, q.provider, q.payload",
                    Some(1),
                    &[],
                )?;

                let mut iter = claimed.into_iter();
                let Some(row) = iter.next() else {
                    return Ok(None);
                };

                let id: i64 = row.get(1)?.expect("id non-null");
                let kind: String = row.get(2)?.expect("kind non-null");
                let provider: String = row.get(3)?.expect("provider non-null");
                let payload: pgrx::JsonB = row.get(4)?.expect("payload non-null");
                Ok(Some((id, kind, provider, payload.0)))
            })
        });

    let Some((id, kind, provider, payload)) = (match claim {
        Ok(opt) => opt,
        Err(e) => {
            pgrx::log!("stewards: claim phase errored: {}", e);
            return false;
        }
    }) else {
        return false;
    };

    pgrx::log!(
        "stewards: claimed work_item id={} kind={} provider={}",
        id,
        kind,
        provider
    );

    // ----- Phase 2: dispatch (no tx; HTTP allowed) -----
    let outcome = dispatch(&kind, &provider, &payload);

    // ----- Phase 3: write result -----
    let write: Result<(), pgrx::spi::Error> = BackgroundWorker::transaction(|| {
        Spi::connect_mut(|client| {
            match &outcome {
                Ok(WorkOutcome::Embedded {
                    target_table,
                    target_id,
                    model,
                    embedding_text,
                    dimensions,
                }) => {
                    // Write vector back to the target row. We hard-code
                    // brain_entries for now; messages comes when chat
                    // step lands. The cast to vector(N) validates
                    // dimensions; mismatch raises a Postgres error
                    // that the outer match converts to row error.
                    let update_target = format!(
                        "UPDATE stewards.{} \
                         SET embedding = $2::vector({}), \
                             embedded_at = now(), \
                             embedded_model = $3, \
                             embedding_error = NULL \
                         WHERE id = $1",
                        target_table, dimensions
                    );
                    client.update(
                        &update_target,
                        None,
                        &[
                            target_id.clone().into(),
                            embedding_text.clone().into(),
                            model.clone().into(),
                        ],
                    )?;

                    let result_jsonb = pgrx::JsonB(serde_json::json!({
                        "kind": "embed",
                        "provider": provider,
                        "model": model,
                        "dimensions": dimensions,
                        "target": format!("{}#{}", target_table, target_id),
                    }));
                    client.update(
                        "UPDATE stewards.work_queue \
                         SET status = 'done', result = $2, done_at = now() \
                         WHERE id = $1",
                        None,
                        &[id.into(), result_jsonb.into()],
                    )?;
                }
                Ok(WorkOutcome::Echo(value)) => {
                    let result_jsonb = pgrx::JsonB(value.clone());
                    client.update(
                        "UPDATE stewards.work_queue \
                         SET status = 'done', result = $2, done_at = now() \
                         WHERE id = $1",
                        None,
                        &[id.into(), result_jsonb.into()],
                    )?;
                }
                Ok(WorkOutcome::Chatted {
                    response,
                    session_id,
                    model,
                    agent_family,
                    requested_model,
                    assistant_content,
                    assistant_tool_calls,
                    reasoning_content,
                    reasoning_details,
                    finish_reason,
                    tokens_in,
                    tokens_out,
                    reasoning_tokens,
                }) => {
                    // Insert the assistant turn. tool_calls and the
                    // reasoning fields are stored verbatim so the
                    // next compose_messages call can echo them back
                    // (required by Moonshot when thinking is enabled).
                    // parent_work_id ties this message back to THIS
                    // work item so tool_dispatch can find it.
                    let tool_calls_jsonb = assistant_tool_calls
                        .clone()
                        .map(pgrx::JsonB);
                    let reasoning_details_jsonb = reasoning_details
                        .clone()
                        .map(pgrx::JsonB);
                    client.update(
                        "INSERT INTO stewards.messages \
                            (session_id, role, content, model, \
                             tool_calls, finish_reason, \
                             tokens_in, tokens_out, reasoning_tokens, \
                             reasoning_content, reasoning_details, \
                             parent_work_id) \
                         VALUES ($1, 'assistant', $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
                        None,
                        &[
                            session_id.clone().into(),
                            assistant_content.clone().into(),
                            model.clone().into(),
                            tool_calls_jsonb.into(),
                            finish_reason.clone().into(),
                            (*tokens_in).into(),
                            (*tokens_out).into(),
                            (*reasoning_tokens).into(),
                            reasoning_content.clone().into(),
                            reasoning_details_jsonb.into(),
                            id.into(),
                        ],
                    )?;

                    // Phase 4f — Record a cost_event so cost-cap discipline
                    // and bucket tracking actually receive token data.
                    // Only fires when the chat is tied to a work_item (the
                    // payload's _work_item_id field, set by work_item_
                    // dispatch_stage). Ad-hoc chats without a work_item
                    // (e.g., watchman passes) are not cost-tracked yet —
                    // cost_events.work_item_id has a NOT NULL FK so they
                    // can't be recorded without a schema change.
                    //
                    // IMPORTANT: use `requested_model` (the substrate's
                    // canonical short name like 'kimi-k2.6'), NOT `model`
                    // (the provider's full versioned identifier like
                    // 'moonshotai/kimi-k2.6-20260420'). model_pricing is
                    // keyed on the canonical name; the provider response
                    // echoes back its own internal versioned name which
                    // doesn't exist in our pricing table → micro_dollars=0.
                    //
                    // Cache token parsing (cache_creation_input_tokens,
                    // cache_read_input_tokens from Anthropic-style
                    // providers) is deferred — we pass 0/0 for now.
                    //
                    // Errors logged, never propagated — a missing pricing
                    // row or transient SPI failure should not poison the
                    // chat write.
                    if let Some(wi_str) = payload
                        .get("_work_item_id")
                        .and_then(|v| v.as_str())
                    {
                        let in_tok = tokens_in.unwrap_or(0);
                        let out_tok = tokens_out.unwrap_or(0);
                        if in_tok > 0 || out_tok > 0 {
                            let cost_result = client.update(
                                "SELECT stewards.record_cost_event( \
                                    $1::uuid, \
                                    (SELECT count(*)::int + 1 FROM stewards.cost_events WHERE work_item_id = $1::uuid), \
                                    $2, $3, $4, $5, 0, 0, $6)",
                                Some(1),
                                &[
                                    wi_str.into(),
                                    provider.to_string().into(),
                                    requested_model.clone().into(),
                                    in_tok.into(),
                                    out_tok.into(),
                                    format!(
                                        "work_id={} session={} response_model={}",
                                        id, session_id, model
                                    ).into(),
                                ],
                            );
                            if let Err(e) = cost_result {
                                pgrx::log!(
                                    "stewards: record_cost_event failed for work_item {} (work_id {}): {}",
                                    wi_str, id, e
                                );
                            }
                        }
                    }

                    // Loop continuation: if assistant returned
                    // tool_calls AND we haven't exhausted agent.steps,
                    // enqueue a tool_dispatch row. The bgworker will
                    // pick it up on the next poll (~500ms).
                    let has_tool_calls = assistant_tool_calls
                        .as_ref()
                        .and_then(|v| v.as_array())
                        .map(|a| !a.is_empty())
                        .unwrap_or(false);
                    let mut continuation_enqueued: Option<i64> = None;
                    let mut stop_reason: Option<&'static str> = None;
                    if has_tool_calls && finish_reason.as_deref() == Some("tool_calls") {
                        // Pull iteration count and agent.steps in one
                        // round-trip. Default steps to 8 if the agent
                        // row's steps column is somehow NULL.
                        let iter_row = client.select(
                            "SELECT \
                                stewards.iteration_count($1) AS iter, \
                                coalesce((stewards.resolve_agent($2, $3)).steps, 8) AS max_steps",
                            Some(1),
                            &[
                                session_id.clone().into(),
                                agent_family.clone().into(),
                                requested_model.clone().into(),
                            ],
                        )?;
                        let mut iter_iter = iter_row.into_iter();
                        let iter_r = iter_iter.next().expect("iter row");
                        let iter_count: i32 = iter_r.get(1)?.unwrap_or(0);
                        let max_steps: i32 = iter_r.get(2)?.unwrap_or(8);

                        if iter_count < max_steps {
                            let enq_row = client.select(
                                "SELECT stewards.tool_dispatch_enqueue($1, $2, $3, $4, $5)",
                                Some(1),
                                &[
                                    id.into(),
                                    agent_family.clone().into(),
                                    requested_model.clone().into(),
                                    session_id.clone().into(),
                                    provider.to_string().into(),
                                ],
                            )?;
                            let mut e_iter = enq_row.into_iter();
                            let e_r = e_iter.next().expect("enqueue returns id");
                            continuation_enqueued = Some(e_r.get(1)?.unwrap_or(0));
                        } else {
                            pgrx::log!(
                                "stewards: agent step budget exhausted ({} >= {}); not continuing",
                                iter_count, max_steps
                            );
                            stop_reason = Some("steps_exhausted");
                        }
                    } else if has_tool_calls {
                        // Provider returned tool_calls but with a
                        // finish_reason other than 'tool_calls'
                        // (e.g., 'length' truncation mid-call). Don't
                        // try to continue — the call list may be
                        // incomplete and dispatching it would corrupt
                        // the conversation.
                        stop_reason = Some("truncated_tool_calls");
                    }

                    let result_jsonb = pgrx::JsonB(serde_json::json!({
                        "kind": "chat",
                        "provider": provider,
                        "model": model,
                        "session_id": session_id,
                        "finish_reason": finish_reason,
                        "tokens_in": tokens_in,
                        "tokens_out": tokens_out,
                        "reasoning_tokens": reasoning_tokens,
                        "billable_output":
                            tokens_out.unwrap_or(0)
                            + reasoning_tokens.unwrap_or(0),
                        "tool_call_count":
                            assistant_tool_calls.as_ref()
                                .and_then(|v| v.as_array())
                                .map(|a| a.len())
                                .unwrap_or(0),
                        "continuation_enqueued": continuation_enqueued,
                        "loop_stop_reason": stop_reason,
                        "response": response,
                    }));
                    client.update(
                        "UPDATE stewards.work_queue \
                         SET status = 'done', result = $2, done_at = now() \
                         WHERE id = $1",
                        None,
                        &[id.into(), result_jsonb.into()],
                    )?;
                }
                Ok(WorkOutcome::ToolsDispatched {
                    parent_work_id,
                    session_id,
                    agent_family,
                    model,
                    tool_messages,
                }) => {
                    // Insert one role='tool' message per dispatched
                    // call, with tool_call_id echoing the assistant's
                    // tool_call.id (provider requirement: each tool
                    // reply must reference its call). parent_work_id
                    // points at THIS tool_dispatch row for trace.
                    for (tc_id, _name, content) in tool_messages.iter() {
                        client.update(
                            "INSERT INTO stewards.messages \
                                (session_id, role, content, \
                                 tool_call_id, parent_work_id) \
                             VALUES ($1, 'tool', $2, $3, $4)",
                            None,
                            &[
                                session_id.clone().into(),
                                content.clone().into(),
                                tc_id.clone().into(),
                                id.into(),
                            ],
                        )?;
                    }

                    // Enqueue the next chat round. compose_messages
                    // will pick up the new tool messages automatically
                    // because they're now in the session history.
                    let next_row = client.select(
                        "SELECT stewards.chat_post_internal($1, $2, $3, $4)",
                        Some(1),
                        &[
                            agent_family.clone().into(),
                            model.clone().into(),
                            session_id.clone().into(),
                            provider.to_string().into(),
                        ],
                    )?;
                    let mut n_iter = next_row.into_iter();
                    let next_chat_work_id: i64 = n_iter
                        .next()
                        .and_then(|r| r.get(1).ok().flatten())
                        .unwrap_or(0);

                    let result_jsonb = pgrx::JsonB(serde_json::json!({
                        "kind": "tool_dispatch",
                        "parent_work_id": parent_work_id,
                        "session_id": session_id,
                        "tool_count": tool_messages.len(),
                        "tools": tool_messages.iter()
                            .map(|(tc_id, name, _)| serde_json::json!({
                                "tool_call_id": tc_id,
                                "name": name,
                            }))
                            .collect::<Vec<_>>(),
                        "next_chat_work_id": next_chat_work_id,
                    }));
                    client.update(
                        "UPDATE stewards.work_queue \
                         SET status = 'done', result = $2, done_at = now() \
                         WHERE id = $1",
                        None,
                        &[id.into(), result_jsonb.into()],
                    )?;
                }
                Ok(WorkOutcome::WaitingForTools {
                    parent_work_id,
                    session_id,
                    agent_family,
                    model,
                    resolved,
                    pending,
                }) => {
                    // Phase 3e.2.b — async fan-out. The dispatch
                    // emitted at least one mcp_proxy child; we
                    // pause this tool_dispatch row in
                    // 'waiting_for_tools' and store enough state
                    // for the SQL completion pass to finish the
                    // job once children resolve. NO message inserts
                    // and NO continuation chat enqueue here — both
                    // happen inside tool_dispatch_complete_waiting().
                    let resolved_json: Vec<serde_json::Value> = resolved
                        .iter()
                        .map(|(tc_id, name, content)| serde_json::json!({
                            "tc_id":   tc_id,
                            "name":    name,
                            "content": content,
                        }))
                        .collect();
                    let pending_json: Vec<serde_json::Value> = pending
                        .iter()
                        .map(|(tc_id, name, child_id)| serde_json::json!({
                            "tc_id":         tc_id,
                            "name":          name,
                            "child_work_id": child_id,
                        }))
                        .collect();
                    let result_jsonb = pgrx::JsonB(serde_json::json!({
                        "kind": "tool_dispatch_waiting",
                        "parent_work_id": parent_work_id,
                        "session_id": session_id,
                        "agent_family": agent_family,
                        "model": model,
                        "provider": provider,
                        "resolved": resolved_json,
                        "pending":  pending_json,
                        "started_waiting_at": format!("{:?}", std::time::SystemTime::now()),
                    }));
                    client.update(
                        "UPDATE stewards.work_queue \
                         SET status = 'waiting_for_tools', result = $2 \
                         WHERE id = $1",
                        None,
                        &[id.into(), result_jsonb.into()],
                    )?;
                    pgrx::log!(
                        "stewards: tool_dispatch id={} waiting on {} mcp_proxy child(ren)",
                        id, pending.len()
                    );
                }
                Ok(WorkOutcome::Resolved {
                    ref_str,
                    content,
                    error,
                }) => {
                    // UPSERT the cache row. attempt_count increments
                    // on conflict so we can see how many tries a
                    // flaky ref has taken.
                    let content_jsonb = content.clone().map(pgrx::JsonB);
                    client.update(
                        "INSERT INTO stewards.resolved_refs \
                            (ref, content, error, fetched_at, attempt_count) \
                         VALUES ($1, $2, $3, now(), 1) \
                         ON CONFLICT (ref) DO UPDATE \
                         SET content = EXCLUDED.content, \
                             error   = EXCLUDED.error, \
                             fetched_at = now(), \
                             attempt_count = stewards.resolved_refs.attempt_count + 1",
                        None,
                        &[
                            ref_str.clone().into(),
                            content_jsonb.into(),
                            error.clone().into(),
                        ],
                    )?;
                    let result_jsonb = pgrx::JsonB(serde_json::json!({
                        "kind": "resolve_ref",
                        "ref":  ref_str,
                        "cached": content.is_some(),
                        "error": error,
                    }));
                    client.update(
                        "UPDATE stewards.work_queue \
                         SET status = 'done', result = $2, done_at = now() \
                         WHERE id = $1",
                        None,
                        &[id.into(), result_jsonb.into()],
                    )?;
                }
                Err(msg) => {
                    pgrx::log!("stewards: work_item id={} failed: {}", id, msg);
                    // Best-effort: also stamp the brain row's
                    // embedding_error if this was an embed job, so
                    // the failure surfaces in app queries.
                    if kind == "embed" {
                        if let (Some(table), Some(target_id)) = (
                            payload.get("target_table").and_then(|v| v.as_str()),
                            payload.get("target_id").and_then(|v| v.as_str()),
                        ) {
                            let stamp = format!(
                                "UPDATE stewards.{} SET embedding_error = $2 WHERE id = $1",
                                table
                            );
                            // Ignore secondary errors (e.g., table
                            // we don't know about) — primary error
                            // is already on its way to the queue.
                            let _ = client.update(
                                &stamp,
                                None,
                                &[target_id.to_string().into(), msg.clone().into()],
                            );
                        }
                    }
                    // tool_dispatch failures: write synthetic
                    // role='tool' replies + enqueue continuation so
                    // the loop never stalls. Phase 1.6 left this
                    // gap open. Phase 1.6.1 closes it.
                    let mut continuation: Option<i64> = None;
                    if kind == "tool_dispatch" {
                        if let (Some(parent), Some(session), Some(family), Some(model_str)) = (
                            payload.get("parent_work_id").and_then(|v| v.as_i64()),
                            payload.get("session_id").and_then(|v| v.as_str()),
                            payload.get("agent_family").and_then(|v| v.as_str()),
                            payload.get("model").and_then(|v| v.as_str()),
                        ) {
                            let synth = client.select(
                                "SELECT stewards.synthesize_tool_failure($1, $2, $3, $4, $5, $6)",
                                Some(1),
                                &[
                                    parent.into(),
                                    family.to_string().into(),
                                    model_str.to_string().into(),
                                    session.to_string().into(),
                                    provider.to_string().into(),
                                    msg.clone().into(),
                                ],
                            );
                            match synth {
                                Ok(rows) => {
                                    continuation = rows.into_iter().next()
                                        .and_then(|r| r.get(1).ok().flatten());
                                    pgrx::log!(
                                        "stewards: synthesized tool failure for parent={}; continuation={:?}",
                                        parent, continuation
                                    );
                                }
                                Err(e) => {
                                    pgrx::log!(
                                        "stewards: synthesize_tool_failure SPI failed: {} (loop will stall)",
                                        e
                                    );
                                }
                            }
                        }
                    }
                    let err_result = pgrx::JsonB(serde_json::json!({
                        "error": msg,
                        "continuation_after_failure": continuation,
                    }));
                    client.update(
                        "UPDATE stewards.work_queue \
                         SET status = 'error', error = $2, result = $3, \
                             done_at = now() \
                         WHERE id = $1",
                        None,
                        &[id.into(), msg.clone().into(), err_result.into()],
                    )?;
                }
            }

            // NOTIFY listeners with the row id as payload.
            let notify_sql = format!("NOTIFY stewards_done, '{}'", id);
            client.update(&notify_sql, None, &[])?;
            Ok(())
        })
    });

    if let Err(e) = write {
        pgrx::log!("stewards: write phase errored for id={}: {}", id, e);
    }
    true
}

// `WorkOutcome` enum moved to types.rs (Phase 3c.3.6 v2 module split).

/// Dispatch a work item by `kind`. Returns `Ok(WorkOutcome)` on
/// success, `Err(message)` on failure (the message is stored in
/// `work_queue.error` and surfaces to callers).
fn dispatch(
    kind: &str,
    provider: &str,
    payload: &serde_json::Value,
) -> Result<WorkOutcome, String> {
    match kind {
        "echo" => Ok(WorkOutcome::Echo(serde_json::json!({
            "echo": payload,
            "kind": kind,
            "provider": provider,
            "stub": "pg_ai_stewards echo",
        }))),
        "embed" => embed(provider, payload),
        "chat"  => chat(provider, payload),
        "tool_dispatch" => tool_dispatch(payload),
        "resolve_ref"   => resolve_ref(payload),
        other => Err(format!("unknown work kind: {}", other)),
    }
}

/// Call an OpenAI-compatible /v1/embeddings endpoint and format the
/// response as a Postgres `vector` text literal (e.g. "[0.1,0.2,...]").
fn embed(provider_name: &str, payload: &serde_json::Value) -> Result<WorkOutcome, String> {
    let provider = PROVIDER_REGISTRY
        .get()
        .ok_or_else(|| "provider registry not initialized".to_string())?
        .providers
        .iter()
        .find(|p| p.name == provider_name)
        .ok_or_else(|| format!("unknown provider: {}", provider_name))?;

    let text = payload
        .get("text")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "payload.text missing".to_string())?;
    let model = payload
        .get("model")
        .and_then(|v| v.as_str())
        .unwrap_or(&provider.default_model);
    let target_table = payload
        .get("target_table")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "payload.target_table missing".to_string())?
        .to_string();
    let target_id = payload
        .get("target_id")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "payload.target_id missing".to_string())?
        .to_string();
    let expected_dim = payload
        .get("dimensions")
        .and_then(|v| v.as_i64())
        .unwrap_or(768) as i32;

    let url = format!(
        "{}/embeddings",
        provider.base_url.trim_end_matches('/')
    );
    let body = serde_json::json!({
        "model": model,
        "input": text,
    });

    // 120s timeout: LM Studio's first request after a cold start
    // can take that long while it loads the model into memory.
    let client = reqwest::blocking::Client::builder()
        .timeout(std::time::Duration::from_secs(120))
        .build()
        .map_err(|e| format!("http client build: {}", e))?;

    let mut req = client.post(&url).json(&body);
    if let Some(key) = &provider.api_key {
        req = req.bearer_auth(key);
    }

    let resp = req
        .send()
        .map_err(|e| format!("POST {}: {}", url, e))?;
    let status = resp.status();
    if !status.is_success() {
        let body = resp.text().unwrap_or_default();
        return Err(format!("embeddings HTTP {}: {}", status, body));
    }

    let parsed: serde_json::Value = resp
        .json()
        .map_err(|e| format!("decode embeddings response: {}", e))?;

    let arr = parsed
        .get("data")
        .and_then(|d| d.as_array())
        .and_then(|a| a.first())
        .and_then(|d| d.get("embedding"))
        .and_then(|e| e.as_array())
        .ok_or_else(|| {
            format!(
                "unexpected embeddings response shape: {}",
                parsed
            )
        })?;

    if arr.len() as i32 != expected_dim {
        return Err(format!(
            "embedding dimension mismatch: got {}, expected {}",
            arr.len(),
            expected_dim
        ));
    }

    // Build pgvector's text format: "[v1,v2,...]". No spaces; floats
    // formatted with full f32 precision.
    let mut s = String::with_capacity(arr.len() * 12);
    s.push('[');
    for (i, v) in arr.iter().enumerate() {
        if i > 0 {
            s.push(',');
        }
        let f = v
            .as_f64()
            .ok_or_else(|| format!("embedding[{}] not a number", i))?;
        // f32 max precision is ~9 digits; pgvector stores f32 anyway.
        s.push_str(&format!("{}", f));
    }
    s.push(']');

    Ok(WorkOutcome::Embedded {
        target_table,
        target_id,
        model: model.to_string(),
        embedding_text: s,
        dimensions: expected_dim,
    })
}

/// Call an OpenAI-compatible /v1/chat/completions endpoint.
///
/// Payload shape (built by stewards.chat_enqueue):
///   {
///     "session_id":      "<id>",
///     "agent_family":    "<family>",
///     "requested_model": "<model>",
///     "meta":            { ... audit only, not sent ... },
///     "body":            { "model":..., "messages":[...], "tools":[...], ... }
///   }
///
/// On success, returns Chatted with the parsed assistant message
/// extracted into top-level fields. Phase 3 inserts that message
/// into stewards.messages and stamps usage.
fn chat(provider_name: &str, payload: &serde_json::Value) -> Result<WorkOutcome, String> {
    let provider = PROVIDER_REGISTRY
        .get()
        .ok_or_else(|| "provider registry not initialized".to_string())?
        .providers
        .iter()
        .find(|p| p.name == provider_name)
        .ok_or_else(|| format!("unknown provider: {}", provider_name))?;

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
    let requested_model = payload
        .get("requested_model")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "payload.requested_model missing".to_string())?
        .to_string();
    let body = payload
        .get("body")
        .ok_or_else(|| "payload.body missing".to_string())?;

    let url = format!(
        "{}/chat/completions",
        provider.base_url.trim_end_matches('/')
    );

    // Chat timeout. 120s was the original (matched embeddings) but
    // reasoning models on big inputs blow past that — the proposal
    // doc + ~50KB scratch files timed out during Phase 3a Watchman
    // smoke. Default raised to 600s; override via STEWARDS_CHAT_TIMEOUT_SECONDS
    // for ops tuning without a binary rebuild. The bgworker is
    // single-threaded per process, so a long chat blocks the queue —
    // the right Phase 3b move is also CLI-side input trimming, not
    // unbounded server time.
    let timeout_secs: u64 = std::env::var("STEWARDS_CHAT_TIMEOUT_SECONDS")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(600);
    let client = reqwest::blocking::Client::builder()
        .timeout(std::time::Duration::from_secs(timeout_secs))
        .build()
        .map_err(|e| format!("http client build: {}", e))?;

    let mut req = client.post(&url).json(body);
    if let Some(key) = &provider.api_key {
        req = req.bearer_auth(key);
    }

    let resp = req
        .send()
        .map_err(|e| format!("POST {}: {}", url, e))?;
    let status = resp.status();
    if !status.is_success() {
        let resp_body = resp.text().unwrap_or_default();
        return Err(format!("chat HTTP {}: {}", status, resp_body));
    }

    let parsed: serde_json::Value = resp
        .json()
        .map_err(|e| format!("decode chat response: {}", e))?;

    // Standard OpenAI shape: { choices: [{ message: { role, content,
    // tool_calls? }, finish_reason }], usage: { prompt_tokens,
    // completion_tokens } }
    let choice = parsed
        .get("choices")
        .and_then(|c| c.as_array())
        .and_then(|a| a.first())
        .ok_or_else(|| format!("no choices[0] in response: {}", parsed))?;
    let message = choice
        .get("message")
        .ok_or_else(|| format!("no choices[0].message: {}", parsed))?;

    // OpenAI returns content as either a string OR null (when only
    // tool_calls are present). NOT NULL on messages.content with
    // default '' handles both — we coerce to "".
    let assistant_content = message
        .get("content")
        .and_then(|v| v.as_str())
        .unwrap_or("")
        .to_string();
    let assistant_tool_calls = message.get("tool_calls").cloned();
    // Reasoning capture. Field names vary by gateway:
    //   OpenRouter / OpenCode Go: `reasoning` (string), `reasoning_details` (array)
    //   Moonshot direct:          `reasoning_content` (string)
    // Coalesce both string forms; keep details verbatim for fidelity.
    let reasoning_content = message
        .get("reasoning_content")
        .or_else(|| message.get("reasoning"))
        .and_then(|v| v.as_str())
        .map(String::from);
    let reasoning_details = message.get("reasoning_details").cloned();
    let finish_reason = choice
        .get("finish_reason")
        .and_then(|v| v.as_str())
        .map(String::from);

    let model = parsed
        .get("model")
        .and_then(|v| v.as_str())
        .unwrap_or_else(|| {
            body.get("model").and_then(|v| v.as_str()).unwrap_or("?")
        })
        .to_string();

    let usage = parsed.get("usage");
    let tokens_in = usage
        .and_then(|u| u.get("prompt_tokens"))
        .and_then(|v| v.as_i64())
        .map(|v| v as i32);
    let tokens_out = usage
        .and_then(|u| u.get("completion_tokens"))
        .and_then(|v| v.as_i64())
        .map(|v| v as i32);
    // OpenAI's newer usage shape:
    //   usage.completion_tokens_details.reasoning_tokens
    // Reasoning tokens are NOT a subset of completion_tokens for kimi/
    // o1-class models — they're billed separately. The OpenCode Go
    // dashboard's "OUTPUT" column sums both; we record them apart so
    // cost math stays honest.
    let reasoning_tokens = usage
        .and_then(|u| u.get("completion_tokens_details"))
        .and_then(|d| d.get("reasoning_tokens"))
        .and_then(|v| v.as_i64())
        .map(|v| v as i32);

    Ok(WorkOutcome::Chatted {
        response: parsed,
        session_id,
        model,
        agent_family,
        requested_model,
        assistant_content,
        assistant_tool_calls,
        reasoning_content,
        reasoning_details,
        finish_reason,
        tokens_in,
        tokens_out,
        reasoning_tokens,
    })
}
