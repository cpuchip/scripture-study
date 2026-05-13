---
title: Batch K — Progressive Context Disclosure for the Dispatch Substrate
status: research / pre-proposal
date: 2026-05-13
project: pg-ai-stewards
related:
  - .spec/proposals/substrate-batch-j-fanout-brainstorm.md (Batch J — shipped today)
  - .spec/journal/2026-05-13-batch-j-shipped.md (J.3 token-limit failures)
  - .spec/scratch/brainstorm-context-management-candidates.md (J.4 brainstorm output)
  - extension/src/bgworker.rs (chat dispatch loop)
  - extension/pg_ai_stewards--0.2.0.sql (compose_messages)
binding_question: |
  How should the substrate manage growing context across multi-turn tool-using
  sessions? When a single fetch_url tool result lands a 426K-character body in
  stewards.messages, every subsequent dispatch turn re-includes it in full —
  blowing the model's input limit. The fix must keep full content retrievable
  (a quote in a study still requires the verbatim source) while keeping the
  active context lean.
---

# Batch K — Progressive Context Disclosure

## 1. The problem, concretely

J.3 (science-center exhibits fanout) shipped 2 of 6 children today. The other 4 all failed with the same shape — Moonshot's Kimi K2.6 rejecting the gather-stage chat call because the request body exceeded 262K input tokens (374K and 376K requested in the worst cases).

**The smoking gun.** Querying the session for the worst failure (`wi--b0b1185f--gather`, 14 messages, 496KB total content):

| Row | Role | Content chars | Tool-calls chars |
|----:|---|---:|---:|
| 1 | user | 7,360 | 0 |
| 2 | assistant | 0 | 935 |
| 3 | tool | 4,845 | 0 |
| 4 | tool | **28,876** | 0 |
| 5 | tool | 3,366 | 0 |
| 6 | tool | **21,183** | 0 |
| 7 | tool | 3,545 | 0 |
| 8 | assistant | 0 | 703 |
| 9 | tool | 1,793 | 0 |
| 10 | tool | 1,710 | 0 |
| 11 | tool | 6,740 | 0 |
| 12 | tool | **426,651** | 0 |
| 13 | user | 207 | 0 |
| 14 | user | 207 | 0 |

Row 12 — **one tool message holds 426,651 characters (~142K tokens) of fetched content**. The substrate retrieved a large web page (probably a research-paper PDF, an arXiv HTML mirror, or a vendor data sheet) and stored the body verbatim. Every subsequent `compose_messages()` call replays this row in full. Once it lands, the session is over: each new dispatch will fail.

This is not a Kimi-specific bug. The same pattern would 429 GPT-4o, blow Anthropic's 200K window, and waste cache on Gemini. **The substrate has no concept of context discipline.**

## 2. Current implementation (what we have)

`compose_messages(agent_family, model, session_id, user_input?)` lives in `pg_ai_stewards--0.2.0.sql:725`. Body:

```sql
SELECT coalesce(jsonb_agg(... ORDER BY m.created_at, m.id), '[]'::jsonb)
  INTO v_history
  FROM stewards.messages m
 WHERE m.session_id = p_session_id;

v_result := jsonb_build_array(system_message) || v_history;
IF p_user_input IS NOT NULL THEN
    v_result := v_result || jsonb_build_array(user_message);
END IF;
RETURN v_result;
```

**No truncation. No summarization. No size guard.** Every turn re-includes the full history. The 426K row gets sent on every dispatch from that session onward.

The bgworker (`bgworker.rs:1678 chat()`) accepts the body unchanged and POSTs it to the provider. There's no client-side context check before sending. The error comes back from Moonshot as HTTP 400 with an explicit message.

## 3. Prior art (verified via web search 2026-05-13)

### Anthropic — context editing + memory tool (Sonnet 4.5, public 2025)

Two coupled capabilities:
- **Context editing** automatically clears stale tool calls and results from the context window when approaching token limits. In a 100-turn web-search evaluation, context editing enabled agents to complete workflows that would otherwise fail due to context exhaustion **while reducing token consumption by 84%**.
- **Memory tool** is client-side: Claude makes a tool call to create/read/update/delete files; the application executes locally. Memory persists across sessions.

Anthropic's [effective context engineering guide](https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents) recommends three patterns:
1. **Compaction** — summarize history near the limit, preserving "architectural decisions, unresolved bugs, and implementation details while discarding redundant tool outputs."
2. **Structured note-taking** — agents maintain external NOTES.md / to-do files.
3. **Multi-agent isolation** — sub-agents return condensed summaries (typically 1,000–2,000 tokens) to a coordinator. (We already do this — our research-write `synthesize` stage receives the `gather` stage's brief, not gather's full history.)

> *"identify the smallest set of high-signal tokens that maximize the likelihood of your desired outcome"*

### LangChain Deep Agents (2026)

Concrete numeric thresholds:
- **85% of context window** = compression trigger. "When the context size crosses a threshold, we offload old write/edit arguments from tool calls to the filesystem."
- **20,000 tokens** = single-tool-result offload threshold. "When Deep Agents detects a tool response exceeding 20,000 tokens, it offloads the response to the filesystem and substitutes it with a file path reference and a preview of the first 10 lines."
- Agents can re-read or search the offloaded content as needed.
- "**Keep the most recent tool calls in raw format** so the model maintains its rhythm and formatting style"
- "**Do not compress away error traces** — when a tool call fails, leaving the error and stack trace helps the model avoid repeating the same mistake."

### MemGPT (Packer et al, 2023, arXiv 2310.08560) — virtual context paging

OS-style hierarchical memory. Key technique: **recursive summarization** — the first system message in every dispatch holds a running summary of evicted history. The queue manager generates a *new* recursive summary using the *existing* recursive summary plus the newly-evicted messages.

This is conceptually beautiful (the summary itself is bounded) and operationally tricky (the summary becomes a single point of failure — if it loses an entity, that entity is gone unless re-derived from the offloaded full history).

### LogRocket / state-of-context-engineering 2026

> *"the field has converged on sliding window plus summarisation hybrids as the dominant approach: keep recent turns in full detail, compress older context through LLM-based summarisation."*

> *"successful production systems treat context engineering as a first-class discipline and deliberately filter, rank, prune, summarize, and isolate information."*

## 4. Brainstorm output (J.4 self-application)

Ran `start_brainstorm()` on this exact question — 4 lenses in parallel, ~$0.06 total. Synthesis at `projects/pg-ai-stewards/.spec/scratch/brainstorm-context-management-candidates.md`. Highlights:

- **All 4 lenses converged on the same architectural skeleton**: full results stay in `stewards.messages`, summaries replace them in active context, retrieval tool fetches full on demand. The structural answer is settled.
- **Disagreement on triggers**: turn count vs token budget vs single-result size vs cumulative byte count. Brainstorm synthesis recommended **cumulative tool-result byte count**, with hard ceiling at ~70% of model's context window. This matches LangChain's 85% threshold roughly (70% gives more headroom for the response).
- **Black Hat surfaced the covenant risk**: a cheap summarizer that drops URLs, dates, or verbatim quotes directly violates the `read_before_quoting` covenant. **Source-verification is non-negotiable**. The summarizer prompt must explicitly preserve concrete entities (URLs, filenames, dates, exact numbers, direct-quoted text).
- **Reverse lens insight**: design the summarizer prompt by inverting failure modes. "How would I make this summarizer DESTROY the cite chain?" → "Don't preserve URLs." → invert → "MUST preserve every URL verbatim."
- **Async is tempting but dangerous** — race condition: a fast follow-up dispatch could call `compose_messages()` before the async summary commits. **Synchronous compression inside `compose_messages()` is the safe first implementation.** Async is a later optimization.
- **Open question**: how does the main agent know which summarized result to fetch? Brainstorm offered checksum-addressable (Six Hats) and embedding-based retrieval (SCAMPER). Both work; checksum is simpler.

## 5. Proposed design

The convergence across prior art + brainstorm produces a coherent design. Phases ordered by dependency.

### 5.1 Storage model (no migration on read)

`stewards.messages` already has the right shape. Add **two columns** (idempotent migration):

```sql
ALTER TABLE stewards.messages ADD COLUMN IF NOT EXISTS
    content_summary text,           -- short summary, NULL if not compressed
    content_offloaded_at timestamptz; -- when summary replaced content in active context
```

The `content` column stays untouched — full text always retrievable. `compose_messages()` will choose between `content` and `content_summary` based on policy (next section).

### 5.2 Single-result offload (LangChain pattern)

**Trigger:** any single tool message with `length(content) > 60_000` characters (~20K tokens).

**Action at message insert time** (in the bridge / `complete_waiting_tool_dispatches` path):
1. Insert the full content into `stewards.messages.content` (unchanged).
2. Compute a sha256 prefix of the content (8 chars) as a stable id.
3. Generate a short summary via a cheap model: **qwen3.6-air** (cloud) or **a locally-hosted llama / mistral via Ollama** (free). Use the summarizer prompt below.
4. Write summary to `content_summary`. Set `content_offloaded_at = now()`.

**Summarizer prompt (preserves source-verification):**

```
You are a tool-result summarizer. Given a tool result (web page, file content,
search hits, etc.), produce a 100-200 token summary for an LLM's working context.

MUST PRESERVE VERBATIM:
- Every URL (markdown links to the source).
- Every date and exact number.
- Every direct-quoted passage the calling agent might want to cite.
- Author names, file paths, organization names.

MAY COMPRESS:
- Prose framing, repeated phrasings, marketing language.
- HTML/markdown structural noise.

OUTPUT FORMAT:
- 100-200 tokens.
- End with: "(full content retrievable via fetch_message_content(<id=...>))"
- The <id> is provided in the system message at compression time.
```

Wrap the cheap summarizer's output with the verbatim id so the main agent knows the affordance.

### 5.3 `compose_messages` reads the summary when offloaded

Tiny change to existing function:

```sql
WHEN m.role = 'tool' THEN jsonb_build_object(
    'role', 'tool',
    'tool_call_id', coalesce(m.tool_call_id, ''),
    'content',
        CASE WHEN m.content_summary IS NOT NULL AND m.content_offloaded_at IS NOT NULL
             THEN m.content_summary
             ELSE m.content
        END
)
```

Backwards-compatible: rows without `content_summary` use `content` as today. The transition is incremental — new large tool results offload; existing rows in old sessions stay raw until a session is replayed.

### 5.4 `fetch_message_content` tool

Expose a single MCP tool the main agent can call:

```jsonc
{
  "name": "fetch_message_content",
  "description": "Retrieve the full content of a previously-summarized tool result. Use when the summary mentions something specific you need verbatim — a quote, a URL, a number, a code block.",
  "input_schema": {
    "type": "object",
    "properties": {
      "message_id": { "type": "string", "description": "The id mentioned in '(full content retrievable via fetch_message_content(id=...))'" }
    },
    "required": ["message_id"]
  }
}
```

When called, the bridge returns the full `content` from `stewards.messages WHERE id = message_id`. This inserts the full content as a new tool message — bounded to ONE per fetch, replacing the summary in active context only for that turn's response. The summary stays in place for subsequent turns.

(Open question: should fetch_message_content's returned content also be subject to the same offload threshold? Probably yes, but with a slight bump — if the agent explicitly asked for it, give it a turn or two of full content before re-summarizing.)

### 5.5 Session-level cumulative cap (safety net)

Per-message offload solves the 426KB-tool-result case. But many medium-size results can still add up to >70% of the context window. Add a session-level cumulative check:

**Trigger:** before each dispatch, compute total bytes that would be sent. If > 70% of target model context window, **summarize older turns** (LRU, oldest first), skipping:
- The system prompt (never compress).
- The user's current binding question (never compress).
- The last N=3 assistant + tool turns (recent rhythm, per LangChain).
- Any message with `role='user'` (binding clarifications, never compress).
- Any **error trace** (per LangChain — preserve so the agent doesn't repeat the failure).

Older `tool` and `assistant` messages get the same summarizer treatment as 5.2, but at a coarser granularity (the assistant turn including its tool_calls is summarized as "made 3 tool calls to investigate X, found Y").

### 5.6 Cost model

Per-turn cost increase:
- **Best case**: 0 (no offload triggered).
- **Typical**: 1 cheap summarizer call per dispatch (~$0.0005 with qwen3.6-air; $0 with local Ollama).
- **Worst case**: N cheap calls when summarizing N older turns at session-cumulative threshold (~$0.005 if N=10).

Net effect on a session like the failed `wi--b0b1185f--gather`:
- Without K: dispatch fails at 262K tokens. Cost: $0.10 of wasted gather, 0 deliverable.
- With K: 426K-char tool result summarized to ~200 tokens at insert time. Subsequent dispatches succeed. Cost: 1 summarizer call (~$0.0005) at offload, plus normal dispatch costs.

**Net positive on cost and reliability.**

## 6. Phasing

Five sub-phases. Each commitable in one pulse. Estimated total: ~2-3 build sessions.

| Sub-phase | What | LLM cost | Risk |
|---|---|---|---|
| **K.1** | `messages.content_summary` + `content_offloaded_at` columns. Backfill NULL. Update `compose_messages()` to read summary when present. | $0 | low |
| **K.2** | Single-result offload trigger on `stewards.messages` INSERT. Cheap-summarizer SQL function `summarize_tool_result(message_id)`. Heuristic fallback (first 200 + last 200 chars) when summarizer unavailable. | ~$0.005/dispatch where it triggers | medium — summarizer prompt design |
| **K.3** | `fetch_message_content` MCP tool. Bridge tool registration. Permission: granted to all agents. Smoke test by manually offloading + retrieving. | $0 smoke | low |
| **K.4** | Session-cumulative cap. Pre-dispatch sizing check. LRU summarize-older for messages exceeding 70%. Protected-classes rules (system / user / recent N / error traces). | ~$0.001-0.005/over-cap dispatch | medium |
| **K.5** | Replay-test on the 4 J.3 failed children. Re-run with K applied; verify they now produce briefs. | ~$2-3 LLM (real research-write retries) | low (passes or fails cleanly) |

## 7. Open questions

1. **Embedding-based vs checksum-based retrieval.** Brainstorm SCAMPER proposed embedding-based: the main agent's *next message* gets embedded, top-K nearest summarized messages re-inflate. Brainstorm Six Hats proposed simpler checksum-addressable. Checksum is simpler and adequate for the failure modes I've seen — embedding is a v2 enhancement.

2. **What's the local summarizer?** Brainstorm Black Hat warned the cheap summarizer must preserve URLs/dates/quotes. Options:
   - qwen3.6-air (cloud, ~$0.0005/call, fast)
   - llama-3.3-8b via local Ollama (free, slightly slower)
   - mistral-small via LM Studio (free, slightly slower)
   - Custom fine-tune on a "preserve verbatim entities" dataset (later)
   First implementation: pick qwen3.6-air as default (already configured in pipelines), with a config knob to switch to local.

3. **Is `compose_messages()` the right place?** The function is `STABLE` and called from many places. Summary-substitution is read-only and idempotent (the summary already exists). Inserting offload at INSERT time (trigger) makes `compose_messages` a pure read. This is cleaner than computing on every call.

4. **What about `tool_calls` JSON?** Currently small (typically <1KB per assistant turn) but if a multi-tool dispatch makes 10 calls, could grow. Not yet a failure mode; defer.

5. **Streaming vs blocking summarization at INSERT.** If the bridge inserts a 426KB tool message and synchronously calls a 2-second summarizer, the dispatch loop blocks. Options:
   - Block (simple; adds 1-3s to tool-call landing).
   - Insert + queue an async work_queue job; mark `content_offloaded_at = NULL until processed`. `compose_messages` returns `content` if `content_offloaded_at IS NULL`, summary otherwise.
   - First implementation: block. The race condition in async is the killer Black Hat caught. Tune later.

6. **Should error messages bypass offload entirely?** Per LangChain, error traces help the agent not repeat. Add a heuristic: `if content matches /error|traceback|exception|stderr/i, skip offload regardless of size`. Defer to K.2 review.

## 8. What this doesn't address (and why that's OK)

- **System prompt growth.** Our `compose_system_prompt()` already includes intent + covenant + agent prompt (~600 tokens overhead). Not yet a failure mode; defer until covenant blocks exceed ~2K tokens.
- **Cross-session memory.** Anthropic's memory tool is a persistent file store across sessions. We have something similar via the `studies` and `messages` tables, but no agent-facing "remember this for next time" surface. Different problem; defer to a future Batch L.
- **Tool count growth.** As we add MCP servers, the `tools[]` array grows. Each tool definition is ~100-500 tokens. Tracked as carry-forward; not in K.

## 9. Verification target

K is done when:
1. A session with a 426KB tool result no longer fails on subsequent dispatches.
2. The summarizer preserves URLs and dates verbatim — auditable by running summarize + diff against the original to confirm no link is missing.
3. `fetch_message_content` round-trips: agent gets summary, calls fetch, gets the verbatim original.
4. K.5 retest of the 4 failed J.3 children produces at least 2 more verified briefs (i.e., the failure mode is fixed for the dominant case).

Cost ceiling for K.5 retest: $5. Total Batch K budget (build + verify): ~$10.

## 10. Decision points to ratify

Before building, ratify:

1. **Offload threshold per single tool result.** Default 60K chars (~20K tokens). Alternatives: 30K (more aggressive, more summary calls), 100K (less aggressive, larger risk of cumulative-cap hits).
2. **Session cumulative threshold.** Default 70% of model context. Alternatives: 80% (LangChain), 60% (more headroom).
3. **Summarizer model default.** qwen3.6-air vs local Ollama llama-3.3.
4. **Block vs async at INSERT time.** Block (simple) vs queued (faster insert path but adds the race condition Black Hat flagged).
5. **Error-trace bypass.** Skip offload for error-shaped content vs treat uniformly.
6. **fetch_message_content re-offload.** When the main agent fetches a previously-summarized full result, does the returned content also get offloaded on the next turn? Yes-but-with-delay (recent fetch grace period) vs uniform.

---

## Appendix A — Brainstorm artifact

Full brainstorm output (4 lenses + synthesis aggregator) lives at `projects/pg-ai-stewards/.spec/scratch/brainstorm-context-management-candidates.md`. Total cost ~$0.065. Surface bug: aggregator's table linked to per-lens file_destinations that don't exist on disk (lens children don't write files). Same carry-forward as J.3 aggregator; deferred.

## Appendix B — Real data from J.3 failure

`wi--b0b1185f--gather` session (Crystal Radio exhibit, failed at gather): 14 messages, 496 KB total content, with row 12 holding a single tool result of 426,651 characters. Moonshot returned `Invalid request: Your request exceeded model token limit: 262144 (requested: 376671)`.

`wi--60b16bf7--gather` (Bacteriopolis, failed): 12 messages, 480 KB. Similar shape.

`wi--5a31f9d0--gather` (CS Unplugged, failed): 23 messages, 290 KB. More turns, smaller individual messages, but still over.

These are the test cases for K.5 verification.

## Appendix C — Prior art summary

| Source | Threshold | Pattern | Notes |
|---|---|---|---|
| Anthropic Memory Tool (2025) | adaptive | Client-side file store; `view/create/str_replace/delete` ops | Cross-session persistence; complementary to in-session compaction |
| Anthropic Context Editing | adaptive | Auto-clear stale tool calls + results near limit | 84% token reduction on 100-turn web-search eval |
| LangChain Deep Agents | 85% context / 20K per result | Filesystem offload + 10-line preview; raw recent + raw errors | Concrete numbers we can lift |
| MemGPT (2023) | OS-style paging | Recursive summary in first system message | Beautiful but single-summary failure mode |
| 2026 consensus (LogRocket / SWIRL) | sliding window + summarization | "Information discipline as first-class" | The field has converged |
