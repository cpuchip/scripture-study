# compact_context — the commissioned-curation side quest

**Status:** ★ BUILT + PROVEN E2E — 2026-06-14 (OSS `a8d5cc5`). Council held
2026-06-14 (4 questions ratified, see §"RATIFIED BUILD SPEC"); built same
session in OSS core (`extension/21-compact-context.sql` +
`cmd/stewards-mcp/compact_context.go`, wired into lib.rs/Dockerfile/main.go).
Proven through the real substrate path: agent → compact_context tool → spawn
tools-off compactor (deepseek-v4-flash) → judge verdict → substrate applies →
[COMPACTED] marker, reversible. On a 14-message migration session clogged with
spent grep/schema dumps, the compactor correctly compressed the two dumps and
kept the plan (compressed=2, 25s, $0 free-tier). The trailing-reminder + broader
2026 items remain held. Follow-ups: a tests/virgin-smoke.sql assertion; tune the
compactor model via the curate stage (Michael's experiments).

*(Prior status: RATIFIED TO BUILD — 2026-06-12 evening council. Michael lifted
the hold: "I think compact_context is a good one to do.")*

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

---

## RATIFIED BUILD SPEC — M5 council 2026-06-14

Council convened with Michael (he chose "convene the M5 council now" at the
parity roadmap's M5 brake). `dominion_in_council` satisfied. All four parked
questions answered:

1. **Timing → mid-turn** (waiting_for_tools). The tool call suspends the
   parent, the curation side quest runs, the parent resumes recomposed the
   same turn.
2. **Compactor model → a fixed cheap model, but TUNABLE.** Michael: "start
   with fixed cheap model (fast with large 1M context window) … we'll want to
   tune this model, run experiments, etc. to find a good compactor counselor."
   → the model is a **config key**, not hardcoded. Generic default in core;
   operator sets the real model in overlay.
3. **What it sees → the judge-brief condensed surface** (with handles to
   expand specific items on demand). **CORRECTION found during grounding:**
   `render_judge_brief_surface(p_message_id, p_brief)` is **per-message**
   (oversized-tool-result judging), NOT a whole-session view. So the
   session-level "condensed surface" the compactor sees is the
   **`context_pressure(session)` foldable list** — `{handle, est_tokens}` per
   message — plus the ability to read/expand specific messages by handle.
   OPEN REFINEMENT for the build: decide whether to add a thin
   `compact_context_surface(session)` that renders the foldable list + a
   one-line gist per foldable message (cheap), or just hand the compactor the
   raw foldable list + `context_resolve_handle`/read tools.
4. **Trigger → agent-initiated + a ≥threshold pressure-line nudge.** The agent
   decides (judges-not-executors); the system nudges so a foggy parent still
   notices. Persuasion, not compulsion. NOT auto-fired (pressure-shedding is
   the floor/wall).

### Verified mechanics (all primitives confirmed present in OSS core)

- **Mute is by message-id, globally:** `context_mute(message_id, cooldown)` /
  `context_compress(message_id, cooldown)` / `context_pin(message_id)`. The
  compactor targets the PARENT's messages via
  `context_resolve_handle(parent_session_id, handle) → message_id`, then
  mute/compress. From its own sub-session it can act on the parent because the
  ops are message-id-keyed and resolve takes an explicit session.
- **Reversible (blind spot RESOLVED):** `context_expand` / `context_expand_tool`
  is the unmute/restore path. Mute is a recoverable tombstone — "safe by
  construction, an unmute away" holds. The compactor may mute freely; for
  anything uncertain it can still prefer compress (engram; originals never
  destroyed).
- **Mid-turn rails already exist:** `tool_dispatch_complete_waiting()` resumes
  any `tool_dispatch` row in `waiting_for_tools` once its
  `result.pending=[{child_work_id}]` children are `done`/`error`, then
  re-dispatches the parent chat (which recomposes → renders the new
  mutes/engrams). `spawn_subagent` already rides these rails — compact_context
  reuses them rather than inventing suspend/resume.
- **Agent shape:** `stewards.agents` has `family, mode, model_pin, prompt,
  temperature, top_p, response_format, steps, active, working_budget,
  **context_tools_enabled**, kind, allow_self_base_prompt` (NO model/provider
  cols — model comes from model_pin/model_match + dispatch). The compactor is
  an agent with `context_tools_enabled=true` and the three-judge prompt.
- `context_pressure(session)` returns `{foldable[], est_tokens, current_turn,
  message_count}` — no pct. The ≥threshold nudge keys off a **config token
  threshold** (tunable), not a window fraction.
- Side-quest spawn: `spawn_subagent_create(pipeline_family, binding_question,
  …)` is pipeline-based; or enqueue a chat for a `compactor` agent directly
  (chat_post_internal pattern, as consult does). Build decides which.

### Components for `extension/21-compact-context.sql` (the remaining build)

1. **Config seeds** (`stewards.config`): `compact_context_model` (generic
   default, tunable) + `compact_context_suggest_tokens` (nudge threshold).
2. **Pressure-line nudge:** edit `context_pressure_line` to append
   "consider compact_context to curate this window" when
   `est_tokens >= compact_context_suggest_tokens`.
3. **`compactor` agent:** `context_tools_enabled=true`, prompt = the three
   judge questions (is the fruit good? what is most precious to keep? what is
   curatable?) + "resolve handles against session '<parent>' and
   mute/compress the curatable; pin the precious; expand to undo." Grants:
   context_mute/compress/pin/expand/resolve_handle (+ read).
4. **`compact_context(p_session_id)`** dispatch fn: render the foldable
   surface, spawn the compactor sub-session targeting the parent, set the
   calling parent `tool_dispatch` to `waiting_for_tools` with
   `pending=[{child_work_id: <compactor wq>}]` (mirror spawn_subagent so
   `tool_dispatch_complete_waiting` auto-resumes the parent recomposed).
5. **`compact_context_tool(p_args)`** wrapper + `tool_defs` row + grant to the
   general agent families (so any agent can self-commission).
6. **Compaction entry:** the side quest IS a work item / sub-session
   (reviewable); write a parent-session marker message
   `[COMPACTED] muted N, compressed M, ~X tokens freed` (the accounting —
   watch_what_you_order).
7. **Chain + smoke:** add `21-compact-context.sql` to `lib.rs`
   `extension_sql_file!` chain (after 20; deps all earlier) + bundle; add a
   `tests/virgin-smoke.sql` assertion (compact_context exists, compactor agent
   ships with context_tools_enabled, nudge fires past threshold, and an e2e:
   bloat a session → call compact_context → compactor mutes parent handles →
   `context_pressure(parent).est_tokens` drops + a `[COMPACTED]` marker lands).

### Why the build paused here (2026-06-14, past midnight, Sabbath next day)

The council was the gated part and is done. The build touches the most
delicate core (the tool_dispatch waiting-tool integration) and surfaced a real
design correction (per-message vs session brief). Per keep_the_watch_whole /
Mosiah 4:27 / not_bypass_process, the SQL build is left for a fresh watch with
this spec locked — not rushed under fatigue into the public substrate. Estimate
once fresh: one focused session (the spec is execution-ready; the only open
choice is component #3's surface granularity).
