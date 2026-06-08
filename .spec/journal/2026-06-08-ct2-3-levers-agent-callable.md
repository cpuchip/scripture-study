---
date: 2026-06-08
title: CT2.3 — the context levers become agent-callable (the loop closes)
workstream: pg-ai-stewards
mode: dev
tags: [ct2, context-management, rust, sql_fn, rebuild, inverse-hypothesis]
---

# CT2.3: a dispatched agent can now manage its own context

Michael chose "build CT2.3 now" + had cleared the restart. CT2.3 closes the loop
CT2.1 (state model) → CT2.2 (render honors it + emits handles) → **CT2.3 (agent
folds/pins by handle)**.

## The one Rust change

`exec_sql_fn_tool` (tools.rs) ran `SELECT fn($1)` with only the model's args, so a
lever couldn't know which session a `[ctx:handle]` belongs to. Fix: thread the
dispatch `session_id` (already in `tool_dispatch`) down through `exec_one_tool` →
`exec_sql_fn_tool`, which injects `_session_id` into the args jsonb. Backward-
compatible — the 4 existing sql_fn tools ignore the extra key. ~4 lines.

## The SQL half

`ct2-3-context-tools.sql`: `context_resolve_handle(session, handle)` (session-
scoped, lenient input); five wrappers (`context_compress/mute/expand/pin/unpin`,
`p_args jsonb`) that resolve the handle, call the CT1 lever, and **catch the §4
lock RAISE as a structured `{"error":…}`** so a locked re-toggle informs the model
instead of crashing the dispatch; tool_defs(sql_fn) registration; and a
`compose_tools` **gate** — `context_*` tools appear only for
`context_tools_enabled` families (decision #6), additive so every existing
family's tool list is unchanged.

## Verification (the discipline that's been paying off)

- **SQL wrappers (direct):** compress-by-handle resolves + sets state+lock; re-call
  within cooldown → `{"error":"context lock … cannot re-toggle yet"}`; bad handle /
  missing session → friendly errors.
- **Gate:** persona + dev (not enabled) → 0 context tools; an enabled family → all 5.
- **Rebuild + restart:** pg image built (Rust compiled); `down`→`up` (no `-v`);
  the bridge entrypoint's startup migrate reported **"current (217 files, 0 drift)"**
  → daemon started. The morning's drift fix is what made this clean — without it
  the bridge (`set -e`, no `|| true`) would have exit-2'd and never started.
- **★ Injection end-to-end:** a real kimi dispatch called a temp `echo` sql_fn
  tool; the reply was `{"ping":"hello","_session_id":"wi--48e80c00--turn"}` — the
  Rust injected the live dispatch session. The whole chain works.

## State

CT2 core (§§1–6) is complete and usable. Soak paused for the build, resumed after.
Ledger clean. Leftover harmless smoke records: the `ct2-echo-pipe` pipeline + its
completed work_item (FK-linked, so the pipeline delete was declined — left as a
record, like the research_codebase smokes).

## Carry-forward
- CT2.4 (task #136): the A/B — does agent-driven context management beat
  automatic-only on a long judgment-heavy run? Enable per-stage via
  `agents.context_tools_enabled`. Measure tokens/cost/quality + thrash.
- §7 (durable self-notes / self-editable base prompt / working tags): UNRATIFIED.
- Root unpushed (Michael pushes): the overnight 4 + morning CT2.2/CT2.3 + memory.
