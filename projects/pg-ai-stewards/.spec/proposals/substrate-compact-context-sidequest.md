# compact_context — the commissioned-curation side quest

**Status:** RATIFIED TO BUILD — 2026-06-12 evening council (same day as
the seed). Michael lifted the hold: "I think compact_context is a good
one to do, and might as well pull that one in." Builds as a P1-adjacent
leg in **OSS core** (it is core context machinery), exercised on the
side-by-side stack. The parked council questions below get settled in a
quick ratification batch when the leg starts. The OTHER research-derived
items (trailing-reminder, broader 2026 adoption) remain held — "the rest
need experiments and more research."

*(Original seed status: Michael's sketch, 2026-06-12 — "lets not act on
the research yet." This file existed so the shape wasn't lost.)*

**Binding question:** when an agent's own context grows past usefulness,
can it commission a reviewable side quest to curate that context — judge
pattern, not executor pattern — and then continue its work lighter?

## The sketch (Michael, verbatim intent)

> we've given the models a lot of tools and maybe these new covenants and
> enabling freer use of the context tools might help here — with an
> indicator and instructions to keep the context below 50% we could enable
> a compact_context tool with instructions that will spin off a side quest
> to look at what's turned on in the gathered context and see what can be
> muted or engrammed; once finished the original loop can continue with a
> compaction entry and the session that spawned it. (fully reviewable.)

## Why this lands (the 2026 evidence it answers, without acting on it yet)

- Context rot: every frontier model degrades with input length;
  "instruction weight loss" is a named failure mode of long agent loops.
- Reasoning shift (arXiv 2604.01161, Apr 2026): long/distracting context
  silently halves reasoning tokens and self-verification — the model gets
  tired. Michael has felt the same in himself.
- Degradation cliffs measured at 40–50% of window — converging with the
  substrate's existing 50% shedding threshold. "Keep below 50%" is not a
  style preference; it is where the cliff is.

## What makes the design strong

1. **It converts compaction from wall to judgment.** Today's pressure
   shedding (50/70/85/95%) is automatic — an executor pattern, and that's
   correct as the safety floor (a wall around the field). compact_context
   adds the judgment layer above it: the agent *notices* its watch
   fragmenting and *commissions* curation. Judges-not-executors, applied
   to the agent's own mind. (Same split as preside §V walls-vs-compulsion.)
2. **It is the presiding covenant, recursive.** The parent presides over
   its compactor: the compaction entry is `watch_what_you_order`'s
   accounting, "fully reviewable" is the Ezekiel clause, and
   `keep_the_watch_whole` becomes *actionable* — the agent's duty to
   reground at the next boundary now has a tool that creates the boundary.
3. **Safe by construction with existing primitives.** mute = recoverable
   tombstone (handle preserved), compress = engram render (originals
   never destroyed), pinned = exempt. A wrong compaction call is an
   unmute away from undone. No deletion anywhere in the loop.

## Existing parts it composes from (nearly no new machinery)

| Sketch element | Already exists |
|---|---|
| indicator | `context_pressure_line` (§5) — could state "keep below 50%" explicitly |
| instructions | a row in `stewards.instructions` — data, not code |
| side quest | `spawn_subagent_create` / consult machinery (es8) |
| look at what's on | judge-brief / `investigate_session` surface; `[ctx:handle]` addressing |
| mute / engram | CT2 context tools (`context_mute`, `context_compress`, engrams from Batch K/L) |
| original loop continues | `waiting_for_tools` suspension pattern (tool_dispatch already does async-resume) |
| compaction entry | messages/work-item rows — the flight recorder is free |
| fully reviewable | the side quest IS a work item with its own session |

## Questions for the ratification council (parked, not asked)

- **Mid-turn or between turns?** Mid-turn = the waiting_for_tools pattern
  (tool call suspends, side quest runs, continuation resumes recomposed —
  the recomposition automatically renders the new mutes/engrams). Between
  turns = simpler, but the relief arrives a turn late.
- **Who is the compactor?** Fixed cheap model vs family. Curation is a
  judgment task but a narrow one; note the council-review-beats-
  gift-matching datapoint (n=1) before assuming a fancy lineup.
- **What does the compactor see?** The parent's full message list with
  handles + engram states, or the judge-brief condensed form?
- **Trigger discipline:** agent-initiated only, or also suggested by the
  pressure line at ≥50% ("consider compact_context")? The covenant's
  dominion_in_council says a new standing capability needs a council —
  this file is the pre-read.
- Relation to the trailing-reminder proposal (journal 2026-06-12): both
  are responses to the same evidence; they may ship together or the
  compactor may make the trailing echo less necessary (a compacted
  context keeps the covenant proportionally closer to the end).

## Anti-scope (named now so the seed doesn't bloat)

- Not automatic compaction-by-rule — that already exists as pressure
  shedding and stays the floor.
- Not deletion. Nothing in this loop destroys content.
- Not a replacement for engram extraction at ingest — this curates what
  ingest-time judgment let through.
