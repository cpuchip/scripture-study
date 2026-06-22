---
date: 2026-06-22
topic: pg-ai-stewards — vivint pipeline (A + B + organize keystone, council → build → cutover)
lane: pg-ai-stewards
---

# The vivint knowledge assembly line — built, proven, cut over

One long, high-momentum session. Michael opened with "lets finish building this!" (the
A/B primitive pair from the prior arc), then "lets open a council and build vivint… promote
B into OSS too, and review the overlay," then "keep going! I love this momentum!" — and we
took the whole arc from two half-built primitives to a live, proven, autonomous vivint loop.

## What shipped (all gated, all green)

**Three primitives into public OSS core (chain 00→44):**
- **A — `route_on`** (42, prior arc): data-driven conditional/loop-back stage edge; the
  coder's hardcoded loop-backs migrated to it. `OK 31`.
- **B — `request_research` + `gather-feedback`** (43, this arc): the analyze→gather feedback
  loop, *promoted from the private overlay* so it's a public primitive alongside A. The
  `gather-feedback` tool_group is the opt-in per-stage scope (the tool-side dual of A: A loops
  within a run, B feeds the pool for a later cycle). `OK 32`. Overlay trimmed to just the Vera
  grant (no clobber).
- **Organize keystone — `graph_node` / `graph_supersede` / freshness** (44, this arc): the
  genuinely-missing capability §6 predicted — nothing let a stage *create a node* (graph_link
  only auto-upserts edge endpoints). Adds the node-maker, `observed_at`/`status` recency,
  `SUPERSEDES` aging, an opt-in `fresh_only` recall, and the `graph-organize`/`graph-read`
  scopes. `OK 33`. **Proven e2e** on *Tao Teh King* — qwen built 14 freshness-stamped nodes +
  25 typed edges (incl the ineffability paradox + a wu-wei tension) in ~48s.

**The vivint pipeline (private overlay, file_private, all-local, $0):**
gather → organize → analyze[A + B] → draft → refine. **Proven e2e on live** (vivint-proof-1,
~10 min, all 5 stages): a 22-node fault graph → analyze `VERDICT: COMPLETE` → a grounded
7-proposal improvement doc (`vivint-improvement-proposals`). **Cut over**: vivint-cron enabled,
the coupled vivint-reflect/planning loop (the backlog source) disabled.

## The council's real findings

- **The overlay audit's honest verdict:** the cut classification held. B was the *one*
  generic-but-overlay capability worth pulling public; everything else (personas, MCP seeds,
  provider keys, local config, his personal intents) is correctly private, and the mechanisms
  one might want to promote (corpus-pools, model/role-aliases) are *already* core with
  overlay-only config. I resisted manufacturing a bigger promotion list to look productive.
- **vivint is the answer to the backlog**, not a parallel thing. The old loop gathered AND
  proposed in one coupled pass; separating knowing (gather+organize→graph) from concluding
  (analyze→proposals) is the Phase-K two-loop that dissolves the coupling.

## Bugs the discipline caught (inverse-hypothesis paid off three times)

1. The book-proof sat `pending` 10 minutes — *not* a broken keystone: `work_item_create` does
   NOT auto-dispatch; the caller must call `work_item_dispatch_stage(id)` (the scheduler does
   it automatically). Diagnosed by checking the dispatch path, not by retrying.
2. The vivint gather template referenced `{{input.gather_gaps}}`, which only exists after a
   route_on loop-back — so the first dispatch failed (the template engine errors loudly on
   missing paths, by design). Seed it empty.
3. The cutover "succeeded" (`UPDATE 1` ×2) but reverted — a trailing ambiguous-`slug` error in
   the same psql `-c` batch rolled back the whole implicit transaction. Caught it by *verifying
   the state*, not trusting the UPDATE count. Redid the UPDATEs as separate statements.

## Presiding / accounting

The cutover changed live autonomous behavior (retired one scheduled loop, started another).
It was ratified in council (dominion_in_council satisfied — all four forks: build+cutover,
prompt-good, recency-now, just-B), proven e2e before the flip, and is accounted here, in the
lane, and to Michael in-session. Not emergency force — a deliberate, watched cutover. The new
loop's first *scheduled* fire is 20:00 (the proof used manual dispatch); worth watching that
the scheduler path runs clean.

## Carry-forward

- Watch the 20:00 scheduled vivint-cron fire (confirms the scheduler path end-to-end).
- The ~66 old research-write proposals in the backlog are separate — Michael's triage, or let
  the disciplined new loop supersede the need.
- **Dashboard introspection** (Michael's "if possible"): a per-GPU "now working: pipeline/stage
  · work_item" panel — join the substrate's in-flight dispatches (by model) with the rig's
  model→GPU map (1:1 in dance-moe). Its own UI/dev arc.
- Backfill (§6 step 3): point the book/video/article digesters at the shared gather+organize so
  all ingestion compounds into one graph. Deferred.
