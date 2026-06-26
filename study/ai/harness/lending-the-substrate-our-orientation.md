# Lending the Substrate Our Orientation

*An audit of where pg-ai-stewards is thin — and what the workspace's battle-tested
knowledge is for. The operational sequel to [the-harness-is-orientation.md](the-harness-is-orientation.md).*

> Gathered 2026-06-26 by three parallel readers (a workspace-discipline audit, a substrate-
> agent audit, and a distillation of Google's *New SDLC* papers), then synthesized by one
> mind. Quotes are file-cited; the three load-bearing ones (the two Google lines and the
> substrate's own "council moment" comment) were re-read from source before quoting. Sources
> listed at the foot. **Draft — its conclusions are candidates for the overlay, then OSS.**

**Binding question.** We are deliberately *thin and light* in pg-ai-stewards' agent
instructions, and that is good. But the workspace holds years of battle-tested orientation —
the council moment, read-before-quoting, the inverse hypothesis, the adjacent-surface audit.
Michael's instinct: *we're missing something.* What is it? What orientation does our workspace
harness carry that the substrate's does not — and is the missing thing the **Orient** move
itself?

---

## I. The audit in one picture

The workspace encodes how-we-work in three tiers: always-on instructions (`copilot-instructions.md`,
~80 lean lines), the bilateral `covenant.yaml`, and ~50 loadable skills the agents reach for as
needed. Keeping the always-on layer lean and pushing procedure to skills is itself a named
discipline (`.mind/principles.md`). The substrate is built the *same* way — a thin core plus a
3-tier skills machinery (`24-skills.sql`) that, by deliberate design, **seeds no skill content
at all.** Same architecture. The difference is what fills it.

The workspace's shelf is full. Its five most battle-tested disciplines, each reinforced across
the covenant *and* the instructions *and* a skill *and* multiple agents — the four-tier presence
that is the best proxy for "this one burned us":

1. **Read before quoting** (`source-verification`; covenant, critical severity) — memory
   confabulates, so every quote is verified against the source this session.
2. **The Council Moment** (`council-moment`; covenant `council_moment:`) — a three-minute scan
   for connections, tensions, and blind spots *before* substantive work (Abraham 4:26).
3. **Watch what you order** (covenant `presiding.watch_what_you_order`) — every delegation
   carries the duty to watch until it obeys the *intent*.
4. **Stewardship over surfacing** (covenant `exercise_stewardship`) — fix the sibling bug and
   report; surfacing-without-acting when action is obvious is offloading dressed as humility.
5. **Verify the fix / build the oracle first** (`debug` Rule 9; `feedback_build_the_oracle_first`)
   — reproduce the failure, confirm the fix kills it; "build passed" is not verification.

The substrate's shelf, by the audit of its 43 agents and its pipelines, holds: agents that
**ground** well (world-build: *"GROUND EVERYTHING… Do not invent lore,"* `61-world-build-worklist.sql`),
that now **commit** well (the COMMIT clause, `45-work-item-chat.sql`; the walk's done-signal,
`61`), and that hand **verification** to a separate layer of judges and gates. What it does *not*
carry, on almost every agent, is the move at the **front** of the loop — the orient.

---

## II. Three witnesses point at one node

The frame study established that the harness is orientation, and that Boyd's OODA loop has one
irreducible node: **Orient** is the interpreting, the judgment, the seat of meaning, and it is
the part that does not automate. Hold that beside the two new witnesses and they say the same
thing from three directions.

**The workspace** made orientation its second-most-battle-tested discipline. The Council Moment
is a hard covenant clause that applies to *every* agent — study, plan, dev, debug — because a
blind spot once shipped a study that contradicted existing work, and the `check_existing_work`
clause was written to prevent the next one. Orientation, here, is not optional and not local. It
is the universal first move.

**Google's SDLC** defines an agent as something that *perceives a goal, plans steps, acts,
observes, and iterates* (day1) — perceive and plan *before* act, by definition. Day 4 makes
"plan quality" and "context handling — did the agent use prior information effectively?" first-
class evaluation dimensions. An agent that does not survey what already exists fails context-
handling on its face. And the Factory Model line names the whole posture: *"Success comes from
giving agents success criteria rather than step-by-step instructions, then letting them iterate"*
(day1) — which only works if the agent first *orients* to those criteria.

**The substrate** built the act leg beautifully and the orient leg almost not at all. That is the
finding, and it is the thing we were missing.

---

## III. The honest correction

The lazy version of this thesis — *"the substrate has no council moment"* — is false, and the
Ben Test requires saying so plainly before building on it.

The substrate has a real, named, scripture-cited council moment. The reflect-steward's
`intent_work_survey` tool carries this comment in its own source: *"This is the substrate's own
Council Moment (Abraham 4:26 — 'took counsel among themselves' before acting)"* (`22-reflect-steward.sql`),
and the tool tells the agent *"Call this FIRST, before proposing anything… This is your council
moment."* It surfaces what is already proposed, in flight, and recently done, with gists, so the
planner does not duplicate. It is exactly the scan-for-connections-and-tensions move — and it
exists because a cold start once accrued thirteen near-duplicate proposals.

The substrate has also already built half of what Google's Day 4 prescribes for the *other* gap:
a **trajectory critic** (`56-trajectory-critic.sql`) that scores a run's process — its tool
selection, its grounding, its error handling — and even carries Day 4's exact warning in its own
prompt: *"A fluent final answer that skipped its verification steps is a MORE dangerous failure."*
The Google paper, in our own words, months early.

So the substrate is not orientation-blind. The accurate, narrower claim is the one the audit
landed on: **orientation in the substrate is surgical, not universal.** Exactly one agent family
orients first, and only because a failure forced it. The transactional workers — research,
work-item-chat, doc-build, the personas, the digesters — have no orientation pass. They act
immediately. The walk and the COMMIT clause arrived the same way: retrofitted onto one agent
after an observed failure, not designed in as a first move every agent makes.

---

## IV. The shape of the gap

Set the two harnesses side by side and the difference is not *amount* — both are thin by design —
but *direction*.

The workspace front-loads orientation as a **universal discipline.** Every skill, every covenant
clause, is crystallized orientation: a past failure turned into a rule the next agent inherits
before it acts. The library grew the honest way, failure by failure, but it grew into a *standing*
first move.

The substrate retrofits orientation **per-failure, per-agent.** It is reactive where the workspace
became proactive. The council moment lives in one agent; trajectory verification exists but runs
as a separate judge, not as a habit every agent carries. And the skills shelf — the very mechanism
built to hold loadable orientation — *ships empty by design.* The machinery is there. The content
is not.

That is the gap, named precisely. Not "the substrate is too thin" — Michael is right that thin is
good. The gap is that **the orientation we have battle-tested over years has not been lent to it.**

---

## V. The reframe: the workspace is the substrate's orientation library

Here is the thing that was missing, said as a claim.

The harness is how a human's orientation gets into the machine — that was the frame study's
thesis. Follow it one step: then **our workspace is the orientation, in storage.** Fifty skills
and a covenant are not documentation. They are years of judgment, each one a failure metabolized
into a standing rule. Read-before-quoting is the mazzaroth miss made permanent. The council moment
is the Section VII blind spot made permanent. Build-the-oracle-first is the eight contaminations
the walk missed, made permanent. The workspace is a reservoir of orientation, and the substrate
has been running on almost none of it — not because thin is wrong, but because the reservoir was
never piped in.

This resolves the tension in Michael's own instinct — *thin is good, but we're missing something.*
Both are true. The substrate's **core** should stay thin; that is the maneuver-warfare bet, the
lean always-on layer. But orientation does not belong in the core. It belongs on the **skills
shelf** — loadable, paid for only when a task reaches for it, the same context-lever the workspace
already uses. Keep the engine thin; fill the shelf. The bounded-gather skill shipped today
(`examples/bounded-gather-skill.sql`) is instance one: a battle-tested workspace discipline — commit,
treat empty as absent, walk a finite set — ported into the substrate as a loadable orientation
module. There are roughly two dozen more on the shelf behind it.

---

## VI. Becoming — what we're missing, and what to do about it

Three moves, smallest to largest. Each ports a battle-tested workspace discipline, answers a named
Google gap, and keeps the core thin.

- **Give every worker the one universal orient move.** Generalize the reflect-steward's
  `intent_work_survey` (`22`) from "the autonomous planner" to "the first move of any agent that
  builds over an existing corpus." This is the Council Moment made universal — the workspace's
  posture, the substrate's existing mechanism, and Google's *context-handling* judge dimension,
  all the same thing. Cheapest where it can be a prompt clause; a tool where the agent needs to
  *see* what already exists.

- **Fill the skills shelf with the battle-tested disciplines.** Port the reservoir as loadable
  orientation skills, a few at a time, oracle-tested like bounded-gather was: read-before-quoting
  for the digesters, the inverse hypothesis for the coder, the adjacent-surface audit for any
  builder. The core stays empty by design; the operator loads what the work needs. The workspace
  becomes the substrate's skill library — its lent orientation.

- **Run the trajectory critic as the trajectory *half* of verification, not a side judge.** The
  substrate built `56`; Google says (day1) that *both* output and trajectory eval are necessary,
  and that the skipped-verification failure is the *more dangerous* one — the 404 read as success,
  the tests reported but never run, the loop that never converged. The data is already in the
  ledger; every tool dispatch is recorded. The missing piece is making the critic a standing
  pre-completion pass, the orient-and-verify bracket closing the loop the way Abraham 4:18 closes
  it: *watch the thing you ordered until it obeys.*

The frame study said orientation is the part God left for us to paint, on purpose, that we might
know the joy of the making. This is the next brushstroke: not a smarter model — Google is explicit
that the intuition to reach for one is wrong — but our own hard-won orientation, lent to the work
a skill at a time. We have been carrying the reservoir all along. The work is to pipe it in.

---

*Sources. Workspace audit: `copilot-instructions.md`, `CLAUDE.md`, `.mind/principles.md`,
`docs/biases.md`, `.spec/covenant.yaml`, and the `.claude/skills/` shelf (council-moment,
source-verification, fan-out, ben-test, debug, bounded-gather, et al.). Substrate audit:
`projects/pg-ai-stewards-oss/extension/` — `22-reflect-steward.sql` (the council moment),
`45-work-item-chat.sql` (COMMIT), `56-trajectory-critic.sql` (Glass-Box), `61-world-build-worklist.sql`
(the walk), `24-skills.sql` (the empty-by-design shelf), `35-research-doc-construction.sql`
(`max_tool_rounds_hard`), `59-self-improvement.sql` (the eval-gaming gate). Google: `external_context/
google-new-sdlc/` days 1, 4, 5 (+ `NOTES.md`), the two quoted lines re-read from `day1-new-sdlc.txt`.
Frame: [the-harness-is-orientation.md](the-harness-is-orientation.md) and its `provenance.md`.*
