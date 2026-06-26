# Force-final-at-cap — a done-rail under every standalone agent

**Status:** proposed (2026-06-26). **NOT for build before the demo** — this is a
bgworker/dispatch-loop change and deserves a calm window + a council moment. Specced now
so it's ready when the demo's behind us.

## The problem

A `dispatch_chat_turn` agent runs a tool loop bounded only by `steps` (its max tool
rounds). When it hits that cap mid-search, the loop simply **stops** — the last turn was a
tool call, so there is no final answer, just an empty bubble or a half-built artifact. The
model never got the chance to "answer now with what you have."

This is the structural floor under the `bounded-gather` failure family. We have closed it
**per-agent** twice — `work-item-chat` (the COMMIT clause, file 45) and `world-build` (the
walk + COMMIT, file 61) — and **per-pipeline-stage** once (`max_tool_rounds_hard`,
file 35, which forces a final tools-disabled turn when a gather stage hits its hard cap).

The gap: **standalone agent dispatches have neither.** Any agent invoked outside a capped
pipeline stage — subagents (`subagent-doc-investigate`, `subagent-docs-audit`), personas,
`loremaster`, `compactor`, and every future one — can still run to its step cap and die
mid-tool-call. We keep patching prompts; the loop itself should guarantee a final turn.

## The fix (one place)

In the bgworker dispatch loop (the same loop that already honors `max_tool_rounds_hard`
for pipeline stages), apply the **force-final** behavior to *every* agent run, keyed off
the agent's own `steps`:

- Track the tool-round count for the dispatch (already tracked for the step cap).
- When the NEXT turn would exceed `steps` (or a configurable `force_final_at = steps - 1`),
  dispatch **one final turn with tools disabled** and a system nudge:
  *"You have reached your tool budget. Do NOT call another tool. Answer now / write your
  result now using only what you have already gathered."*
- That turn's text becomes the final output instead of an empty/cut-off result.

This is exactly what `max_tool_rounds_hard` does for a stage; the change is to make it the
**default for the standalone path** too, not only the pipeline-stage path. Reuse the stage
implementation — don't fork a second mechanism.

## Design choices to settle (the council moment)

1. **Default on, or opt-in?** Lean **default on** for any agent with `steps > 1` (a 1-step
   judge like `trajectory-critic` is already single-shot and needs no final turn). A
   forced final turn is strictly better than an empty bubble; it should be the floor.
2. **Config knob.** A per-agent / global `force_final_at` (default `steps - 1`) so an agent
   can reserve N turns for the final write, not just one — world-build's relationship pass,
   for instance, wants more than one trailing turn. Config row, no rebuild to tune.
3. **Graders are exempt by intent, not by edit.** A judge/critic that hits its cap should
   still emit a verdict on what it saw (a forced final turn is fine — it just says "answer
   now"), but we do NOT add "stop early / be brief" pressure to evaluators. The force-final
   nudge is about *producing a result*, not *gathering less* — keep that line clean so the
   rail never reads as "cut corners."
4. **Interaction with the walk.** For a walk-driven agent that legitimately needs many runs
   (a >150-chunk corpus), force-final ends *this run* cleanly; the persisted coverage makes
   the next run resume. Force-final and the walk compose — the walk makes the cap
   *resumable*, force-final makes the cap *graceful*.

## Why it's the highest-leverage version

The COMMIT clause is advisory (a prompt the model can ignore). The walk is mechanical but
only fits a finite enumerable corpus. **Force-final-at-cap is enforced and universal** —
one loop change protects every standalone agent that exists now or is added later,
including ones the substrate's own self-improver might spawn. It turns "ran out of turns →
nothing" into "ran out of turns → best answer so far" everywhere, for free.

## Scope / cost

- One change in the bgworker dispatch loop (Rust), reusing the existing
  `max_tool_rounds_hard` final-turn path.
- A config row (`force_final_at`) + a virgin-smoke assert (an agent at its cap emits a
  non-empty final turn).
- No new tables, no per-agent prompt edits. The per-agent COMMIT clauses become
  belt-and-suspenders, not the only rail.

## Related

- `bounded-gather` skill — the principle and the three fixes; this is fix #3 (the floor).
- `45-work-item-chat` (COMMIT), `61-world-build-worklist` (walk), `35-research-doc-construction`
  (`max_tool_rounds_hard`, the stage-level version to reuse).
- `.spec/journal/2026-06-26-world-build-worklist.md` — where the family was named.
