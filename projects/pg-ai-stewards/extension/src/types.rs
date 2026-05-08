//! Shared types used across `lib.rs`, `bgworker` dispatch helpers,
//! and `tools` dispatchers. Extracted from lib.rs as Phase 3c.3.6 v2
//! (2026-05-08) so `tools.rs` and `bgworker.rs` can split next without
//! circular module dependencies.

/// Result of running a single work item, before it's written back.
pub(crate) enum WorkOutcome {
    Echo(serde_json::Value),
    Embedded {
        target_table: String,
        target_id: String,
        model: String,
        embedding_text: String,
        dimensions: i32,
    },
    Chatted {
        // Raw provider response (full JSON), for the work_queue audit trail.
        response: serde_json::Value,
        // Echo back so phase 3 can persist the assistant message.
        session_id: String,
        // Model the provider actually used (echo from response.model).
        model: String,
        // Continuation context: needed if assistant returned tool_calls
        // and we want to enqueue a tool_dispatch for the next loop step.
        // These mirror what was in the original payload.
        agent_family: String,
        requested_model: String,
        // Extracted bits we want to write into stewards.messages.
        assistant_content: String,
        assistant_tool_calls: Option<serde_json::Value>,
        // Captured reasoning fields. Required to echo back on the
        // next request when thinking is enabled (kimi-k2.6, o1).
        reasoning_content: Option<String>,
        reasoning_details: Option<serde_json::Value>,
        finish_reason: Option<String>,
        tokens_in: Option<i32>,
        tokens_out: Option<i32>,
        // OpenAI usage.completion_tokens_details.reasoning_tokens.
        // Billed separately from tokens_out by kimi/o1-style models;
        // store so cost computation can sum both. None when absent.
        reasoning_tokens: Option<i32>,
    },
    /// Result of executing one or more tool calls. Phase 3 inserts
    /// each (tool_call_id, content) as a `role='tool'` message and
    /// then enqueues the next chat round to continue the loop.
    ToolsDispatched {
        parent_work_id: i64,
        session_id: String,
        agent_family: String,
        model: String,
        // Per call: (tool_call_id, tool_name, content_jsonb_string).
        // content is what the model will see in the next turn. It's
        // either the JSON-stringified tool result or {"error": "..."}.
        tool_messages: Vec<(String, String, String)>,
    },
    /// Result of resolving a single gospel-engine reference. Write
    /// phase UPSERTs into stewards.resolved_refs. content is the
    /// raw JSON returned by /api/get?ref=... (or null if errored).
    Resolved {
        ref_str: String,
        content: Option<serde_json::Value>,
        error: Option<String>,
    },
}
