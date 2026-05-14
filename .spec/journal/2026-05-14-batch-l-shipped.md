---
date: 2026-05-14
mode: build (single-session sprint)
workstream: WS5 (substrate)
project: pg-ai-stewards
title: "Batch L shipped — Context Engine v2 (graduated rendering + provider rules + cross-message search + heavyweight wrapper set + defense in depth)"
status: shipped — L.1 through L.9 complete (SQL + Go), bin/stewards-mcp.exe rebuilt
carry_forward:
  - "L.1 alone doesn't fix bacteriopolis: session has 15 medium non-engram tool messages totaling ~170K bytes that dominate even crisis-mode engram rendering. Need L.1.1 — lower extraction threshold below 60K bytes OR truncate raw tool messages in crisis mode."
  - "End-to-end smoke of L.6 delegation (actual roundtrip through Claude Code → tool → spawn_subagent_create → child completes → digest returned) deferred. Needs Claude Code restart + a real workload to drive any of the 6 wrappers. SQL side fully smoked (agents/pipelines/perms/tool_defs all registered); Go side compiled and bin rebuilt."
  - "L.3 search_engrams Go MCP tool — synchronous query-side embedding call. SQL search_engrams_by_vector(vector, ...) callable directly. The Go wrapper that embeds a text query then calls the SQL needs its own pulse — deferred (likely Batch M cross-session memory tool ships this)."
  - "Batch M (deferred per ratification): cross-session memory tool, streaming, unified trace, skill cards. Council on these before building."
  - "Resume soak: UPDATE stewards.watchman_config SET schedule_enabled = true (paused during the K build; verify state)."
links:
  - "../proposals/substrate-batch-l-context-engine-v2.md"
  - "../../projects/pg-ai-stewards/extension/l1-provider-rules-and-graduated-rendering.sql"
  - "../../projects/pg-ai-stewards/extension/l3-engram-embeddings-and-search.sql"
  - "../../projects/pg-ai-stewards/extension/l4-mark-engram-important.sql"
  - "../../projects/pg-ai-stewards/extension/l5-re-extract-engrams.sql"
  - "../../projects/pg-ai-stewards/extension/l6-heavyweight-wrappers.sql"
  - "../../projects/pg-ai-stewards/extension/l7-suspect-sources-blocklist.sql"
  - "../../projects/pg-ai-stewards/extension/l8-untrusted-data-wrap.sql"
  - "../../projects/pg-ai-stewards/extension/l9-subagent-depth-cap.sql"
  - "../../projects/pg-ai-stewards/cmd/stewards-mcp/heavyweight_tools.go"
---

# 2026-05-14 — Batch L shipped (one-session sprint)

Michael returned from being on the road, ratified Batch L (the K v2 + all 5 K-deferred carry-forwards), and asked me to ship it all in one go with commits at good boundaries. Nine sub-phases, ten commits.

## What shipped

| Commit | Sub-phase | Contents |
|---|---|---|
| `d8b1276` | L.1 + L.2 + L.4-read-side | provider_rules table (5 providers seeded with message_field_rules + context_window) + graduated rendering under pressure (50/70/85/95% thresholds, drops MEDIUM → COLD → HOT-truncate) + render_engrams_under_pressure + rewrite of compose_messages to self-look-up provider |
| `64692b4` | L.3 | engram_embeddings table (composite id, vector(768) matching studies, HNSW + btree indexes) + AFTER UPDATE trigger populating from messages.engrams + search_engrams_by_vector SQL fn + DO-block backfill (40 rows from existing engrams) |
| `98cbae0` | L.4 | mark_engram_important SQL fn + tool_def registration (write-side; read-side already in L.1's compose_messages) |
| `1e5f21b` | L.5 | re_extract_engrams SQL fn (archives to engrams._history, clears items[], enqueues fresh extraction with new binding) + tool_def |
| `2b76c58` | L.6 | 6 heavyweight wrappers: summarize_url / audit_files / investigate_session / summarize_study / investigate_study / audit_studies — each gets agent + single-stage pipeline + tool_subset enforcement via 38 agent_tool_perms deny rules + tool_def |
| `4939e29` | L.7 | suspect_sources blocklist (3 seed paste-site domains) + suspect_source_approvals + extract_domains_from_jsonb + is_suspect_domain (walks parent chain) + AFTER INSERT/UPDATE OF result on tool_calls trigger that annotates web tool results |
| `c829c23` | L.8 | BEFORE INSERT on messages wraps web-tool results with [BEGIN/END UNTRUSTED EXTERNAL DATA] markers. tool_name_for_tool_call_id resolves via prior assistant tool_calls. is_web_tool covers 7 known web tools |
| `e80a5bd` | L.9 | subagent_depth_of(uuid) walks parent chain (cycle-guard at 64 hops) + check_subagent_depth helper + BEFORE INSERT/UPDATE OF parent_work_item_id trigger raises if a new child would exceed depth 2 |
| `ce63f8a` | Go handlers | heavyweight_tools.go now registers 8 new MCP tools (mark_engram_important, re_extract_engrams, summarize_url, audit_files, investigate_session, summarize_study, investigate_study, audit_studies). bin/stewards-mcp.exe rebuilt |

## Decisions made under pressure

**L.1 4-arg signature retained.** Couldn't drop compose_messages(text,text,text,text) because pg_ai_stewards extension declares dependency on it. Couldn't add a 5-arg overload with all-DEFAULT params because that creates ambiguous dispatch. Resolution: keep the 4-arg signature; have it self-look-up provider via `provider_for_session(session_id)` from work_queue.payload. Cleaner from the caller's perspective anyway.

**L.5 session row first.** Direct enqueue of a chat work_queue row hit messages.session_id FK during the bgworker dispatch (the inserted tool message FKs to a session). Pattern lifted from K.1: create session row before enqueuing.

**L.6 simplest viable wrappers.** Each wrapper is ~30 lines Go (input struct + handler + spawn delegation). Each SQL pipeline is 1 stage. The agent's system prompt teaches what to do; the tool subset enforced via agent_tool_perms denies. Resisted the urge to add per-wrapper output schemas — the parent agent gets prose digest just like spawn_subagent.

**L.7 parent-chain walk for domains.** Flagging "pastebin.com" should also flag "foo.pastebin.com". is_suspect_domain walks foo.bar.example.com → bar.example.com → example.com.

**L.8 tool-name resolution via tool_call_id.** Messages don't carry the producing tool name directly; resolved via prior assistant message's tool_calls array. Supports both `tc.name` (anthropic shape) and `tc.function.name` (openai shape).

**L.9 depth-3 raise smoke.** Direct INSERT INTO work_items for the smoke (bypassing spawn_subagent_create) confirmed the trigger fires. The DO block bubbled the exception out before cleanup ran — but the smoke also showed the inserts rolled back, so no orphan rows.

## What's verified

- All 9 SQL files applied to live container, no errors
- L.1: provider_rules seeded, compose_messages rewrite passes 4-arg dispatch
- L.3: 40 engram_embeddings rows backfilled, embed jobs queued
- L.4: mark_engram_important set is_important=true on msg 2381 e1
- L.5: re_extract_engrams on msg 2381 archived 1 set to _history, cleared items[], enqueued wq=2490 (in_progress)
- L.6: 6 agents + 6 pipelines + 38 deny rules + 6 tool_defs registered cleanly
- L.7: tool_call with pastebin URL annotated with _suspect_severity=warn + warnings array
- L.8: tool message after assistant tool_call wraps correctly with visible markers
- L.9: 3-level chain accepted; depth-3 INSERT correctly raised "subagent depth cap exceeded"
- Go build: `go vet` + `go build -o bin/stewards-mcp.exe` clean

## What's NOT verified (carry-forward)

The end-to-end delegation roundtrip through any of the 6 L.6 wrappers — needs Claude Code restart to pick up the new binary, plus a real workload. SQL side fully smoked; Go side compiled and binary rebuilt; bridge routing is the same `mcp_proxy → server=pg-ai-stewards` pattern that K.5's deep_research already uses.

L.1's bacteriopolis verification target is honest carry-forward: graduated rendering of engrams alone doesn't crush bacteriopolis because the session has 15 medium non-engram tool messages (~170K bytes) that dominate the budget even when MEDIUM tier is dropped. L.1.1 (extraction threshold drop OR raw-message crisis truncation) is needed to actually close that exhibit.

## The pattern that worked

- Decisions ratified in one batch (per the user's Batch L ratification round, recorded in `substrate-batch-l-context-engine-v2.md` §Decisions)
- Phased commits per sub-step, smoke between each
- Single Go file accumulates 8 new tool registrations (heavyweight_tools.go), one bridge rebuild at the end
- Carry-forwards captured in the journal rather than blocking the build

Ten commits. Zero rollbacks. The C–F discipline still holds for L.
