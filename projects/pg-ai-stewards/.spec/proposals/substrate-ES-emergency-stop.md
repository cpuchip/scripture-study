---
name: substrate-ES-emergency-stop
title: "ES — Emergency Stop: critical-failure findings, code trace, and remediation plan"
status: ES.0 stabilized (bleed stopped) — ES.1+ awaiting ratification
created: 2026-05-15
trigger: 2026-05-14/15 bacteriopolis fix-bundle retry — runaway DeepSeek churn, bgworker crash loop, ~$20-70 in wasted contextualizer tokens
debug_workflow: .claude/agents/debug.md (Agans' 9 rules)
applicable_research: Nate B Jones "Pinecone Just Demoted Vector Search" (yt lqiwQiDglGk, 2026-05-13)
---

# ES — Emergency Stop

This document supersedes in-flight L.1.1.x verification. Work is organized
as **ES phases**, each with **sessions** under it. ES.0 is done (bleed
stopped). ES.1+ await ratification.

## Naming

- **ES** — Emergency Stop, the umbrella.
- **ES.N** — a phase (stabilize, schema fix, rearchitect, etc.).
- **ES.N.s{n}** — a session within a phase.

---

## ES.0 — What happened (the incident)

Bacteriopolis fix-bundle retry-2 produced a runaway. Timeline:

1. A **cancelled** work_item (`wi--11bc9874`, fixed-retry-1) kept running its
   chat→tool_dispatch→chat loop. Cancelling the work_item set
   `work_items.status='cancelled'` but **did not stop the session's chat
   loop** — the loop runs on `session_id`, independent of work_item status.
2. Each loop iteration produced tool results. Oversized ones (web fetches,
   ~250K-900K chars) tripped the L.1.1.8 intercept.
3. `chunk_and_index` fired synchronously, splitting each into **160-501
   leaves**, enqueueing **one contextualizer chat per leaf** — hundreds of
   DeepSeek-V4-Flash calls per oversized message.
4. Each contextualizer chat sent the **full document** as prefix (~183K-561K
   tokens in). Total: **230M+ DeepSeek input tokens** observed.
5. The embed jobs those leaves spawned hit a **404** (embeddings provider
   down — see ES.0 finding #4) → bgworker tried to stamp `embedding_error`
   → hit `operator does not exist: bigint = text` → **crashed** → restarted
   → picked up the next failed embed → crashed again. **Tight crash loop.**

Bleed stopped manually: killed all non-terminal work_queue rows for the
affected sessions + `leaf-ctx-*` + paused soak. Queue confirmed empty;
bgworker confirmed back in clean poll loop.

---

## Critical failures (code trace — Agans Rule 3: looked, didn't theorize)

### CF-1 — Cancelled/failed work_item does not stop its session's chat loop

**Severity: critical.** A cancelled or failed work_item keeps spending
indefinitely. There is **no `work_item_cancel` function** in the extension
— cancel is a bare `UPDATE work_items SET status='cancelled'`. The
chat→tool_dispatch→chat loop (`chat_post_internal` ←
`tool_dispatch_complete_waiting`) runs purely on `session_id` and never
checks the owning work_item's status.

**Fix:** a `work_item_cancel(uuid)` function that (a) sets status, (b)
marks all non-terminal work_queue rows for the work_item's `session_ids`
as error/cancelled, (c) cancels any `waiting_for_tools` tool_dispatch rows
so `tool_dispatch_complete_waiting` won't resurrect the loop. AND: the
chat/tool_dispatch handlers should check the owning work_item status
before enqueueing a continuation — defense in depth.

### CF-2 — bgworker embed handler: bigint = text crash

**Severity: critical.** `bgworker.rs` ~line 692:
```
UPDATE stewards.{target_table} SET embedding=$2 ... WHERE id = $1
```
`target_id` ($1) is bound as a **string**. Works for text-id tables
(`studies.id`, `engram_embeddings.id` are text). Crashes on
`messages_raw_overflow_leaves` whose `id` is **bigserial**:
`operator does not exist: bigint = text`. The failure-handler path
(~line 1438) does the *same* `WHERE id` lookup to stamp `embedding_error`
— so even error handling crashes. That is the crash loop's engine.

**Root cause: my L.1.1.4 schema bug.** I gave `messages_raw_overflow_leaves`
a `bigserial` PK; L.3's `engram_embeddings.id` is text by design precisely
so the embed handler's string-bound `WHERE id` works.

**Fix:** change `messages_raw_overflow_leaves.id` to text (composite, e.g.
`'leaf-' || parent_id || '-' || leaf_ordinal`), matching L.3's pattern.

### CF-3 — No circuit breaker on the bgworker crash loop

**Severity: high.** When a work_queue row reliably crashes the worker, the
postmaster respawns the worker, which picks up the same class of row and
crashes again. The periodic reaper marks *stale* rows but does not detect
*crash-looping*. Nothing says "this kind of work is failing every worker;
stop dispatching it."

**Fix:** consecutive-failure tracking per `kind` (or per failing row). After
N crashes attributable to one row/kind, quarantine that row/kind and emit a
loud log. The bgworker survives; the poison pill is isolated.

### CF-4 — Embeddings provider is LM Studio (nomic), not OpenCode Go;
no health check; reboot drops the loaded model

**Severity: high.** Embeddings run on a **local LM Studio** instance
serving a **nomic embedding model**. The host rebooted; LM Studio did not
auto-load the nomic model → embed endpoint returns a 404 HTML page. The
substrate enqueues embed work blindly with no provider health check, so
every embed job fails.

**Fix:** (a) a pre-flight health check on the embedding provider before
enqueueing embed work (or before the bgworker claims an embed row);
(b) operational note: after a reboot, load the nomic model in LM Studio;
(c) consider a substrate `provider_health` table the watchman can populate.

### CF-5 — chunk_and_index has no circuit breaker or cost ceiling

**Severity: high.** `chunk_and_index` splits an oversized message into
parents+leaves and fires `contextualize_leaf` for **every leaf**, in one
synchronous trigger call. A 900K-char fetch → ~501 leaves → 501 chat
work_queue rows in a single statement. No max-leaves guard, no projected-
cost gate, no "this is absurd, stop" check.

**Fix:** hard ceiling on leaves per index (and/or on source bytes). Beyond
the ceiling: don't chunk — see ES.3 (rearchitecture). Also: the ratified
L.1.1 cost cap ($0.50/oversized input) was never actually enforced at the
chunk_and_index dispatch point.

### CF-7 — Embed jobs misrouted to the opencode_go provider

**Severity: high.** Embeddings run on **LM Studio** (provider `lm_studio`,
local, port 1234, model `text-embedding-nomic-embed-text-v1.5`). But the
embed-enqueueing SQL I wrote in L.1.1.5 (`l15`), L.1.1.12 (`l26`), and L.3
(`l3`) hardcoded `provider => 'opencode_go'` on the embed work_queue row.
730 embed rows are queued against `opencode_go` — they 404 against
OpenCode Go's API (which has no embeddings endpoint) regardless of LM
Studio being up. Only 409 rows correctly used `lm_studio`.

**Fix:** change the provider literal to `lm_studio` in the embed-enqueue
sites: `l3` lines ~126 and ~241, `l26` line ~161 (live
`apply_contextualize_leaf`). `l15` is superseded by `l26`.

**Resolved operationally 2026-05-15:** the nomic model was reloaded in LM
Studio via `lms load text-embedding-nomic-embed-text-v1.5`; endpoint
verified returning embeddings. The provider-string fix is still needed
(ES.1) — without it, embed work goes to the wrong place even with LM
Studio healthy.

### CF-6 — The whole leaf-chunk-and-embed architecture is the wrong
abstraction (the deepest finding)

**Severity: architectural.** Validated by the Nate B Jones video and by
Michael's own instinct. We built **chatbot-era RAG** (chunk a document into
hundreds of vector leaves, embed them, vector-search) for an **agent-era
job** (a sub-agent on a mission, fetching a page in service of a binding
question).

From the video: *"the retrieval unit needs to match the work you're
doing"*; Page Index's claim — documents whose structure carries meaning
**should never be chunked**; *"better embeddings don't fix this — all they
do is find more relevant text."* And the rediscovery problem: agents burn
~85% of compute re-reading/re-summarizing.

Michael's framing: *"when I think about generating 500 leafs from one web
fetch I feel like we're overdoing it. I see [the sub-agent] parsing through
the document, throwing away any information that isn't needed, summarizing
and quoting the important bits that build on the binding question,
surfacing relevant information... multiple memories or info bits, few calls
to save tokens/dollars."*

**That is the Judge pattern applied correctly.** A web fetch in service of
a binding question doesn't need 500 vector chunks. It needs **one judge
read** — the sub-agent (which already holds the parent binding question)
reads the document once, discards noise, and emits a small compiled bundle:
a handful of engrams (quotes, facts, dates) tied to the binding question.
Few calls. Few tokens. The retrieval unit is a **compiled brief**, not a
vector index.

This supersedes L.1.1.5 (contextualize_leaf), L.1.1.6 (chunk_and_index),
L.1.1.7 (retrieve_with_merge) for the in-flight web-fetch case. Those were
built for a cross-document semantic-search use case that may still have a
place — but NOT as the default path for "agent fetched a page."

---

## ES phase plan

### ES.1 — Stabilize & guardrails (build-ready after ratification)

Fixes that stop this class of incident regardless of the rearchitecture.

- **ES.1.s1 — work_item_cancel cascade (CF-1).** `work_item_cancel(uuid)`
  function: set status + kill non-terminal work_queue rows for all
  session_ids + cancel waiting tool_dispatch rows. Add a status check in
  the chat-continuation path.
- **ES.1.s2 — chunk_and_index circuit breaker (CF-5).** Hard ceiling on
  leaves/source-bytes. Beyond ceiling: refuse to chunk, surface to the
  judge instead (bridges into ES.3).
- **ES.1.s3 — bgworker crash-loop breaker (CF-3).** Consecutive-failure
  detection; quarantine the poison row/kind.
- **ES.1.s4 — embedding provider health check (CF-4).** Pre-flight check
  before enqueueing/claiming embed work.
- **ES.1.s5 — fix embed provider routing (CF-7).** Change `opencode_go` →
  `lm_studio` in the embed-enqueue sites (l3, l26). Re-queue or discard the
  730 misrouted rows.

### ES.2 — Schema fix (CF-2)

- **ES.2.s1 — messages_raw_overflow_leaves.id → text.** Composite text id
  matching L.3's engram_embeddings pattern. Restores the embed handler.
  (May be moot if ES.3 removes leaf embedding entirely — sequence ES.3's
  ratification first.)

### ES.3 — Rearchitect compaction (CF-6, the real fix)

Council + ratify before building. Candidate shape:

- The L.1.1.8 intercept stops chunking-and-embedding by default.
- Instead: an oversized web fetch is handed to a **judge sub-agent** with
  the parent binding question. The sub-agent reads the document (using its
  context window — kimi/qwen both 260K, deepseek-v4-pro 1M), and returns a
  **compiled brief**: a small set of engrams (quote, fact, date, source
  link) selected against the binding question. Few calls.
- "Is the fruit good? What is most precious to save? What to discard?" —
  answered in ONE pass, not 500.
- Cross-document vector search (L.1.1.5-7, L.3) is retained as a SEPARATE,
  opt-in primitive for the genuine "search across the whole corpus" job —
  not the default for in-flight fetches.
- Aligns with the video's "retrieval unit matches the work" + "write down
  the bundle your agent needs" + "don't overbuild."

### ES.4 — Verify

- Bacteriopolis re-run under the rearchitected compaction. Cost target:
  a web fetch should cost a few calls, not hundreds.
- Agans Rule 9: reproduce the original runaway conditions, confirm the
  guardrails hold.

---

## Carry-forward / open questions

- The ~$20-70 in wasted DeepSeek tokens this incident — confirm against the
  OpenCode Go bill; substrate `cost_usd` was unpopulated so it didn't show.
- `cost_usd` not being populated is its own finding — cost discipline can't
  work if cost isn't tracked. Possibly an ES.1 add.
- L.1.1.5/6/7 disposition: keep as opt-in corpus-search primitive or remove
  entirely — decide in ES.3 council.
- The video's "memory accumulates bad conclusions" warning applies to our
  engram system: an agent storing its own inference as confirmed fact.
  Worth a guard — engrams should carry provenance (extracted-from-source
  vs agent-inferred).
