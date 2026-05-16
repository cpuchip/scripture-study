---
name: substrate-ES-emergency-stop
title: "ES — Emergency Stop: critical-failure findings, code trace, and remediation plan"
status: ES.1 COMPLETE + verified. ES.3 s1-s4 SHIPPED + verified 2026-05-15 (real judge call confirmed); s5 deferred; soak PAUSED pending Michael's resume. ES.4 (full bacteriopolis re-run) pending.
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

### ES.1 — Stabilize & guardrails

Fixes that stop this class of incident regardless of the rearchitecture.

- **ES.1.s1 — work_item_cancel cascade (CF-1). SHIPPED `b6ac127`.**
  work_item_cancel now hard-stops every non-terminal work_queue row
  (pending/in_progress/waiting_for_tools) for the work_item's session_ids.
- **ES.1.s2 — chunk_and_index circuit breaker (CF-5). SHIPPED `61f56d1`.**
  40-leaf ceiling; over it, chunk_and_index returns {skipped:true} and the
  intercept leaves the message raw + flags for ES.3.
- **ES.1.s3 — bgworker crash-loop breaker (CF-3). NOT SHIPPED.** Needs a
  bgworker.rs change + docker rebuild — fresh-context session.
- **ES.1.s4 — embedding provider health check (CF-4). NOT SHIPPED.**
  LM Studio nomic model reloaded operationally; the substrate-side
  pre-flight health check is still to build.
- **ES.1.s5 — embed provider routing (CF-7). SHIPPED `149a783`.** BEFORE
  INSERT trigger forces every kind=embed row to provider=lm_studio.
- **ES.2/CF-2 — disable leaf embed enqueue. SHIPPED `60cd8d2`.** Option B
  (ratified): removed the embed INSERT from apply_contextualize_leaf
  rather than building a text-id cascade ES.3 may discard.

- **ES.1.s3 — bgworker crash-loop breaker (CF-3). SHIPPED `0dcdf75`.**
  kind_circuit_breaker table; reaper records one crash per distinct kind;
  5 consecutive crashes → 10-min pause; claim query skips paused kinds;
  success resets. bgworker rebuilt.
- **ES.1.s4 — embedding health check. SUBSUMED by s3.** The per-kind
  breaker covers the LM-Studio-down case (embed fails → kind pauses →
  cooldown → retries). CF-2 Option B already removed the embed-404 crash,
  so embed failures fail gracefully now. A dedicated retry-on-transient
  would be nicer — carry-forward, not ES.1-blocking.

**ES.1 COMPLETE. Verified by a clean pipeline smoke test 2026-05-15:**
a small research-write run (es-smoke-nomic-embed-compare) reached
verified at $0.205. Every guardrail confirmed in production — no
runaway (caps held: context_gather 4 rounds, synthesize 2), zero
bgworker crashes, all 10 embed jobs routed to lm_studio and completed,
the REVIEW: passes gate satisfied honestly, real 6377-char artifact
produced. One cosmetic finding: kimi-k2.6 reported under 3 gateway
identifiers (Fireworks / Moonshot / canonical) — model-name
normalization is a low-priority ES.3-era cleanup.

### ES.2 — Schema fix (CF-2)

- **ES.2.s1 — messages_raw_overflow_leaves.id → text.** Composite text id
  matching L.3's engram_embeddings pattern. Restores the embed handler.
  (May be moot if ES.3 removes leaf embedding entirely — sequence ES.3's
  ratification first.)

### ES.3 — Rearchitect compaction: the judge-compiled-brief (CF-6, the real fix)

**RATIFIED 2026-05-15.** Council held with identity + ES journal loaded,
gospel research (Matthew 13:47-52 — the net), and six 2026 context-
engineering sources. 7 decisions by user vote.

#### Ratified decisions

| # | Decision | Outcome |
|---|---|---|
| 1 | Judge model | deepseek-v4-flash — 1M context; **no `max_tokens` set** (the L.1.1.12 lesson — never restrict the reasoning budget) |
| 2 | Judge timing | Always sync — every judge call returns a real brief in-turn |
| 3 | Leaf index | DROP in ES.3 — the user's vote is the explicit ratification for the destructive SQL |
| 4 | Re-engagement | **Generalized** (council 2026-05-15 round 2) — `consult_subagent` re-engages ANY spawned sub-agent, not just fetch-judges. Ships in ES.3. |
| 5 | Re-ask cap | Soft STEWARD NOTICE after ~5 re-asks/document + work_item cost cap as hard backstop |
| 6 | Re-ask engrams | Yes — provenance-tagged (`extracted` vs `inferred`) |
| 7 | Empty verdict | The judge may return a brief with zero engrams + a reason — "cast the bad away" is a valid judgment |

#### The architecture

An oversized tool result is the **net** (Matt 13:47 — "gathered of every
kind"). The net is not the sorting. The sort is a separate, deliberate act
by a judge who *sits down* with the catch (v.48) and gathers the good into
vessels. Our bug was making the net sort itself — 500 embedded leaves.

Flow: oversized result → L.1.1.8 intercept → hand the **whole document** +
the **binding question** to a judge sub-agent (deepseek-v4-flash, fresh
isolated session) → judge reads once → returns a **compiled brief**: a
handful of engrams (quote / fact / date + source pointer), each tagged to
the binding question, plus a state line and an explicit *discarded* note.
The brief replaces the result body in the consuming agent's context. The
raw document stays in `messages_raw_overflow` for `expand_message
tier='raw'` — **recoverable** summarization, not lossy (the 2026 field
consensus: keep raw reachable; summarize last).

#### The judge is not a special case — it is a `spawn_subagent`

The substrate already has a general sub-agent primitive: `spawn_subagent`
(Batch K.4) creates a child work_item running ANY registered pipeline
(`research-write`, `study-write`, …), isolated context, own cost cap,
depth-capped. The fetch-judge is just an instance: the L.1.1.8 intercept
calls `spawn_subagent_create` internally with a `judge-brief` pipeline,
the oversized document seeded as the child's context.

Today the sub-agent lifecycle is **spawn → digest → done** — the child
completes and its context is gone. ES.3 adds the missing half: a spawned
sub-agent's session **persists addressable** after it returns its digest,
for the parent work_item's lifetime (the ES.1.s1 cancel cascade still
tears it down on cancel — no orphan spend).

The brief returns with the child's `work_item_id` as the handle. The
parent can later call `consult_subagent(work_item_id, question)` — a sync
chat into the *same* child session, its context still resident
(prompt-cached prefix → cheap re-ask). This works for ANY spawned
sub-agent — a fetch-judge, a delegated scripture study, a transcript
review — not just fetches. A report you file once becomes a steward you
can send back (D&C 104; the householder of Matt 13:52 — the treasure
yields "things new and old").

#### Sub-phases

**SHIPPED 2026-05-15** — s1-s4 built, smoked, committed (`cc8fde9`,
`84209ea`, `2f6c25a`, `c44ddbd`) and verified by a real deepseek-v4-flash
judge call on a 72K-char document (7 well-formed engrams, correct
provenance, embeds routed to lm_studio). Build note: the judge runs as
a bare chat (the K.1 extract_engrams pattern) rather than a
`spawn_subagent` work_item — a trigger context can't cleanly thread a
work_item + pipeline + the spawn_subagent Go handler. `consult_subagent`
keys on the **session** instead, which is the unifying handle for any
sub-agent. s5 deferred (carry-forward). The soak is PAUSED — a full
oversized-fetch inside a live multi-stage pipeline has not run
unattended; resuming is Michael's call (ES.4 territory).

- **ES.3.s1 — Engram provenance + brief schema.** Add `provenance`
  (`extracted` | `inferred`) to the engram shape. Define the compiled-brief
  structure: ≤7 engrams, state line (done / partial / empty), discarded
  note. Provenance answers the Nate Jones warning — memory must not record
  agent inference as sourced fact.
- **ES.3.s2 — Judge dispatch + intercept rewrite.** L.1.1.8 intercept stops
  calling `chunk_and_index`; instead calls `spawn_subagent_create` with a
  `judge-brief` pipeline (deepseek-v4-flash, no max_tokens), the oversized
  document as seeded context. Spawned sub-agent sessions persist addressable
  after digest. Brief replaces the result body; raw → `messages_raw_overflow`;
  consuming agent receives brief + the child `work_item_id` handle.
- **ES.3.s3 — `consult_subagent` tool (general).** Sync dispatch into any
  spawned sub-agent's persistent session — fetch-judge, delegated study,
  transcript review. Companion to the existing `spawn_subagent`. Soft
  STEWARD NOTICE after ~5 re-asks per child (the L.1.1.17 pattern). Re-ask
  answers may mint provenance-tagged engrams.
- **ES.3.s4 — Drop the leaf index (destructive — ratified, decision 3).**
  `DROP TABLE messages_raw_overflow_leaves`; remove `contextualize_leaf`,
  the leaf path of `chunk_and_index`, and `retrieve_with_merge`. One
  migration, gated behind a smoke confirming the judge path works first.
- **ES.3.s5 — (optional) model-name normalization.** Canonical mapping for
  kimi-k2.6's three gateway identifiers. The judge is a new spend path;
  clean cost attribution matters. Low-priority — may slip to carry-forward.

#### Kept (not dropped)

- `messages_raw_overflow` — raw parent recovery; the "recoverable" half of
  recoverable summarization.
- L.3 `engram_embeddings` — corpus-wide semantic search, retained as a
  **separate opt-in** primitive (the dual-index lesson — retrieval-side and
  agent-side chunking are different jobs). Not the default for in-flight
  fetches.
- `map_reduce_extract_engrams` — for unattended cases (sabbath reflection
  over an archive, oversized study inputs) where no live judge holds a
  binding question.

#### Deferred (named, not in scope)

- **Documents larger than the judge's window (>1M tokens).** The brief
  schema's `state: partial` value is the seam — a judge past its window
  returns an honest partial brief, and the parent can spawn another judge
  on the remainder. `map_reduce_extract_engrams` (kept) is the natural home
  for a windowed pre-pass that feeds the judge. Not designed here — cross
  the bridge when a real >1M document forces it.
- **Cross-work-item document reuse** — a shared document cache any mission
  can address. Real and bigger; ES.3 sub-agent sessions are work_item-
  scoped, and the ES.1.s1 cancel cascade tears them down — no orphan spend.
- **Multi-witness extraction** (D&C 6:28) — two judges on load-bearing
  engrams. Phase-F-adjacent; not ES.3.

### ES.4 — Verify

- Bacteriopolis re-run under the rearchitected compaction. Cost target:
  a web fetch costs a few calls, not hundreds.
- Verify a re-ask: parent calls `ask_document` on a judged document, gets a
  coherent answer; the soft STEWARD NOTICE fires on the 6th re-ask.
- Agans Rule 9 (inverse hypothesis): reproduce the original runaway
  conditions — oversized fetch on a cancelled work_item — confirm no leaf
  explosion, guardrails hold, judge path engages instead.

---

## Model-name normalization (new, low priority)

The smoke test surfaced that kimi-k2.6 is reported under three gateway
identifiers — `kimi-k2.6`, `accounts/fireworks/models/kimi-k2p6`,
`moonshotai/kimi-k2.6-20260420` — and the gather stage failed over
between Fireworks and Moonshot routes mid-stage. Same logical model;
not a bug. But cost attribution is split three ways and the L.1.1.15
substitution detector can't see gateway-route changes. A canonical
model-name mapping would fix both. ES.3-era cleanup, not urgent.

## Carry-forward / open questions

- The ~$20-70 in wasted DeepSeek tokens this incident — confirm against the
  OpenCode Go bill; substrate `cost_usd` was unpopulated so it didn't show.
- `cost_usd` not being populated is its own finding — cost discipline can't
  work if cost isn't tracked. Possibly an ES.1 add.
- L.1.1.5/6/7 disposition: DECIDED in the 2026-05-15 council. Leaf index
  (`messages_raw_overflow_leaves`, `contextualize_leaf`, `chunk_and_index`
  leaf path, `retrieve_with_merge`) drops in ES.3.s4. `engram_embeddings`
  (L.3) survives as an opt-in corpus-search primitive.
- The video's "memory accumulates bad conclusions" warning: ADDRESSED by
  ES.3.s1 — engrams gain a `provenance` field (`extracted` | `inferred`).
