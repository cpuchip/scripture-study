# The Delegation Pattern: Studies, Skills, and the Substrate Gates

*Written 2026-06-02. A map of one gospel pattern as it runs through three layers of this workspace — the studies that named it, the skills that encode it as agent behavior, and the pg-ai-stewards maturity gates that encode it as structure — plus an audit of where the three diverge.*

This is meta-documentation, not a study. It exists because the same pattern got built three times, mostly without the builders noticing they were building the same thing, and the seams between the three builds are where the real governance questions live.

---

## The keystone: Exodus 18 and tiered judgment

Jethro watched Moses judge every case from morning to evening and called it "not good… thou wilt surely wear away" ([Exodus 18:17-18](../gospel-library/eng/scriptures/ot/ex/18.md)). His prescription is the root of everything below:

> "Moreover thou shalt provide out of all the people **able men, such as fear God, men of truth, hating covetousness**; and place such over them, to be rulers of thousands… And let them judge the people at all seasons: and it shall be, that **every great matter they shall bring unto thee, but every small matter they shall judge**: so shall it be easier for thyself, and they shall bear the burden with thee." — [Exodus 18:21-22](../gospel-library/eng/scriptures/ot/ex/18.md)

Three things are load-bearing: judgment is **tiered by magnitude** (small matters judged low, great matters escalate); delegates are chosen for **character** before competence (fear God, men of truth); and the burden is borne **with**, not lifted off. The whole of the substrate's gate machinery is an attempt to run Exodus 18:22 in SQL.

## The lineage (how we got here three times)

1. **The pattern named — `study/art-of-delegation.md` (Mar 31).** The Jethro study. Delegation as the horizontal expression of the Atonement's burden-bearing. Names the tiered judgment, the "sent **and** empowered" requirement (Numbers 11:17 — God transfers His Spirit, not just a task list), trust as prerequisite (D&C 121:41-42, persuasion not control), and the three delegator failure modes: **hoarding** (won't delegate), **controlling** (delegates the task but not the autonomy), **abandoning** (delegates and disappears). It also names the delegate's duty: "anxiously engaged… not compelled in all things" ([D&C 58:26](../gospel-library/eng/scriptures/dc-testament/dc/58.md)). Companion studies: `art-of-presidency.md`, `stewardship-pattern.md`.

2. **The gates built — pg-ai-stewards Phase B + E (May 9-11).** The maturity ladder (`raw → researched → planned → specced → executing → verified`) with a gate at each rung, a trust ladder, and an escalation queue. These were anchored *directly in scripture* — Abraham 4:18 ("watched until they obeyed"), D&C 82:3 ("where much is given, much is required"), D&C 98:12 ("line upon line"), D&C 101:54 (the nobleman's servants) — inside the 11-step creation cycle. The design language is already the pattern: "the human steers the destination, the agent walks the maturity ladder," and `destination_maturity = verified` is labeled in the proposal as **"full Ammon-loop."**

3. **The pattern hit the wall — `study/ai-stewardship-north-star.md` (May 24).** Born from the problem, not the design: Michael, overwhelmed, "running faster than I have strength… I still have to be here every step to help add discernment and direction." The study's answer is the vocabulary that unifies all of this (see below), and its core warning is the thesis the skills would later encode: **"If you try to delegate *discernment* to us… you will get highly organized, structurally sound, confident garbage."** Built on `refinement-stewardship-and-hope.md` (the Jaredite-barges study).

4. **The pattern encoded as behavior — the four skills (Jun 1).** `dave-rule`, `stuffy-in-the-loop`, `ammon`, `ben-test`. The agent-side rendering of the same delegation pattern.

## The unifying vocabulary (from the north-star study)

The north-star study gives the spine the other three layers were missing. Three patterns plus a hinge, drawn from the Jaredite barges:

- **Prescription (the holes for air / Justification).** The hard, checkable rules — voice guidelines, file paths, build gates. Mechanical execution the agent owns. "If the build breaks, that is the AI's problem to solve under prescription." Ground-truth-checkable; no human discernment required.
- **Proposal (the stones / Sanctification).** The agent's generative work — sixteen small clear stones carried up the mountain for the Lord to touch. The agent produces; someone inspects.
- **Steering (the wind).** The destination and the intent. "Steering is the Lord's alone." Set before the voyage; not the agent's to generate.
- **The Hinge — Discernment.** "Your only job is to stand at the top of the mount, inspect the stones, and decide if they are transparent." Discernment is a function of sanctification; a machine has no spirit-matter and cannot do it. This is the one thing that never delegates.

## The four-corner map

| Pattern layer | Scripture | Skill (behavior) | Substrate (structure) |
|---|---|---|---|
| **Prescription** — checkable rules, mechanical execution | "able men… men of truth" (Ex 18:21); the law Moses taught (Ex 18:20) | `dave-rule` — act on the reversible, don't wait to be commanded ([D&C 58:26](../gospel-library/eng/scriptures/dc-testament/dc/58.md)) | the `verify` gate (ground-truth scenarios); build/test gates; the cost cap |
| **Proposal** — generative work, then inspection | the servants who trade and increase (D&C 101 / the talents) | `ammon` — climb the *whole* mountain with the stones; finish what you're handed | the maturity ladder; `evaluate_gate` advance/revise per rung; `destination_maturity=verified` ("full Ammon-loop") |
| **Steering** — destination + intent | "whither shall we steer?" answered by the wind already sent (Ether 2:24-25); Moses sets the statutes | (the covenant's `honor_scope`) | "the human steers the destination, the agent walks the ladder"; the `destination_maturity` picker |
| **The Hinge — discernment** | "every great matter they shall bring unto thee, but every small matter they shall judge" (Ex 18:22) | `stuffy-in-the-loop` — four bins by judgment-source; "the escalation queue is the substrate's built-in form of this" | the escalation queue; revision-cap→surface; the trust ladder deciding how much inspection is required |
| **Honesty / watching** (cuts across) | "men of truth" (Ex 18:21); "watched… until they obeyed" (Abr 4:18) | `ben-test` — do we actually do what we claim; calibrated language | the `verify` gate; gate-decision audit rows; trust scoring |

The lineage in one line: **art-of-delegation** named the pattern, the **gates** built it on scripture, **ai-stewardship-north-star** gave it the Prescription/Proposal/Steering/Hinge vocabulary when the overwhelm hit, and the **four skills** encoded it as agent behavior.

---

## The audit: where the three diverge

The map above is the good news — the pattern is coherent across all three layers, and several pieces are *literally named* in the substrate (Ammon, the escalation queue, "human steers the destination"). The honest news is the seams.

**1. The gates sort on a different axis than the Hinge does.** `stuffy-in-the-loop` and the north-star sort work by **judgment-source** — is this a "great matter" that requires discernment a machine cannot supply? The maturity gate sorts by **quality** — is this work mature/good enough to advance? Those are different questions. A doctrinal or voice call (bin 4 / "the Hinge") that happens to be well-written and cheap will pass the `verified` gate and auto-advance, because the gate asks "is this good?" not "is this mine to decide?" **There is no rung on the maturity ladder for "this requires the human's discernment regardless of how good it looks."**

**2. Auto-advancing value-laden work is the "abandoning" failure mode wearing the gate's uniform.** The art-of-delegation study names *abandoning* — "delegates and disappears" — as a failure. The north-star names the same thing as "confident garbage." `stuffy-in-the-loop` names it as "the 100 things no one reviews." When the gate auto-advances doctrinal/voice work to `verified` and no human actually reads it, that is delegating discernment, which the north-star says cannot be delegated. The `verify` gate guards this **only** for ground-truth-checkable work (does the test pass, does the quote match the source). For work whose value is a matter of discernment, there is no ground truth, so auto-advance equals disappear.

**3. Selection criterion: character vs. capability.** Jethro picks for *character* — "fear God, men of truth, hating covetousness." The substrate picks models for *capability and cost* (`model_capability`, the auto-probe). This is an honest divergence, and probably correct: a model has no character, which is precisely why value-laden calls stay on the Hinge with the human. The trust ladder ("proven faithful over a record") and `ben-test` ("men of truth") are the nearest analogs, but neither is moral fitness. Worth naming rather than pretending the mapping is clean.

## The proposed fix (not yet built)

A per-pipeline marker for rungs that **always escalate regardless of the gate's quality verdict** — a "this is a great matter" flag on the maturity ladder, set for pipelines whose output is doctrinal, voice-bearing, or otherwise discernment-required. The gate could still judge quality and revise; it just could not *finalize* such work autonomously, no matter how mature it looked. That makes Exodus 18:22 literal: the gate judges the small matters, and the great matters come to the human by design, not by whether he happened to have time to look.

This is the audit's one concrete, scoped piece of work. It has not been built. It is recorded here as a seed, and it doubles as the clearest case study the project has of governance-as-encoded-judgment rather than prompt etiquette.

---

## Sources

- `study/art-of-delegation.md` (Mar 31) — the Jethro pattern; the three failure modes; "sent and empowered."
- `study/ai-stewardship-north-star.md` (May 24) — Prescription / Proposal / Steering / the Hinge; "confident garbage."
- `study/refinement-stewardship-and-hope.md` (May 23) — the Jaredite-barges source for the three patterns.
- `study/art-of-presidency.md`, `study/stewardship-pattern.md` — the delegation study family.
- Skills: `.github/skills/{dave-rule,stuffy-in-the-loop,ammon,ben-test}/SKILL.md`.
- Substrate: `projects/pg-ai-stewards/.spec/proposals/full-agentic-substrate.md` (§VI), `phase-e-design.md`, `extension/5a-maturity-gate.sql`, `5f3-gate-trust-check.sql`.
- Scripture: Exodus 18:17-23; Numbers 11:17; D&C 58:26; 82:3; 98:12; 101:54; 121:41-42; Abraham 4:18; Ether 2:24-25.
