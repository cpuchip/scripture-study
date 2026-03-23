# The Stewardship Pattern — Reflections

*Date: March 22, 2026*
*Follows: [stewardship-pattern.md](stewardship-pattern.md)*
*Born from: Michael's response to the completed study and ward conference that morning*

---

## What the Study Got Wrong

The original Section VII ("The Asymmetry — Why Machines Are Easier") drew a hard line between delegating to machines and delegating to humans. The table listed "Covenant: None — transactional" for machines and "Council: Not applicable — machines don't advise." The section concluded that "every element that makes human delegation hard is an element that makes it holy" and that "a machine will never say 'the effect was like opening the floodgates of heaven.'"

The problem is that this directly contradicts work we'd already done. The [Working with AI guide](../docs/work-with-ai/guide/05_complete-cycle.md) maps an 11-step creation cycle that includes Step 2 (Covenant) as a bilateral commitment between human and AI — "A command is one-directional: 'Agent, do this.' A covenant is bilateral: 'I commit to providing clear context and timely review. You commit to flagging uncertainty and honoring boundaries.'" Step 3 (Stewardship) defines progressive trust levels. Step 11 (Zion) proposes the bishop/ward model for multi-agent orchestration.

The study had the right scriptures and the right principles. It applied them too narrowly. Michael caught it immediately: "your opening paragraphs totally ignore what work we've done in ./docs/working-with-ai/guide where we do try and map gospel principles to working with AI agents like covenants."

His core insight: "Certainly I can not make those and just boss you around, but that is not as effective for I think the exact same reasons I wouldn't want to just boss around a human agent."

The pattern is the same. The mechanism differs. The study has been revised to reflect this.


## Where the Correction Came From

Michael identified the blind spot because he lives on both sides of the delegation relationship. He manages human councilors (where delegation fails) and AI agents (where delegation works). What he recognized is that the *reason* AI delegation works isn't that machines are simple — it's that he accidentally built a relational system. Intent files, memory architecture, phased workflows, bilateral instructions. He built covenant infrastructure for AI work and didn't notice it because it looked like engineering, not theology.

When the study said "machines don't need covenant," he knew from experience that was wrong. Not because machines have spirits, but because the relational approach produces measurably better output than bare commands. The same Ballard principle applies: presenting a solution and distributing tasks is less effective than opening the floor for genuine interchange — whether the floor is in a ward council or a chat session.

The study's error was a posture problem, not a research problem. It read scripture to confirm the thesis (machine delegation is a lesser form of the pattern) rather than to discover (the pattern works across substrates because intelligence working with intelligence follows the same principles). This is exactly the posture check the critical analysis phase (Phase 3a) was designed to catch: "Are we reading to discover, or to confirm?"


## Ward Conference — "Stakes of Zion"

At ward council that morning, with stake leadership present for ward conference, the stake president observed something to the effect that you don't hear about "wards of Zion" in scripture — you hear about "stakes of Zion." Geographic boundaries matter, but the wards within the stake are all one church working together.

This connects to Mosiah 25:22: "notwithstanding there being many churches they were all one church." The organizational subdivision (ward, quorum, class) serves functional purpose — Jethro's ratios, Alma's one-to-fifty — but the identity is singular. Many agents, one purpose. Many agents, one intent.yaml.

The direct application to multi-agent work: brain.exe routes to different agents with different specializations and different model tiers. But they share the same intent document, the same covenant, the same memory architecture. The routing table (A2 from squad learnings) is the organizational structure. Intent.yaml is the shared identity. "Stakes of Zion" is the level at which shared purpose operates — it's not the individual ward (agent) that constitutes Zion, but the unified stake (system) working toward the same purpose.


## The Three Proposals

### 1. The Covenant — Done

Created [.spec/covenant.yaml](../.spec/covenant.yaml). Bilateral commitments, traced to scriptural sources. Key distinction from the guide's template: Michael's human commitment is to read output fully and understand what is being proposed, rather than "loading memory" (which is the agent's function, not his). The human covenant also includes providing a binding question, flagging when something is wrong, not bypassing process, and reviewing in the same session when possible.

Added to [copilot-instructions.md](../.github/copilot-instructions.md) as a named section, referenced in the session-start sequence (Step 2).

### 2. The Council Moment — Done

Added to copilot-instructions.md as a general principle for all agents. The specific instruction: "At the start of substantive sessions, after loading memory and before diving into the task: actively scan for connections to previous studies, tensions with existing work, and things the human might not be looking for. Three minutes."

On the question of whether this applies to the dev agent: yes, but the mechanism differs. For study and plan agents, the scan is across past studies, existing documents, and thematic connections. For the dev agent, it's across existing code, prior decisions, and architectural patterns. The principle is the same — Abraham 4:26, "took counsel among themselves" before acting. The dev agent's council moment is checking existing implementation before writing new code. The study agent's council moment is checking existing insights before writing new claims. The failure mode in both cases is building confidently on assumptions that contradict what's already there.

### 3. Progressive Trust Tracking — Planned

Not implemented yet. The outline:

**Model capability experiments.** Before we can assign trust levels, we need baseline data. Run the same prompts through Haiku, Sonnet, and Opus. Have Opus + Michael evaluate the output for study quality, accuracy, voice fit. Do this both for individual tasks and for multi-agent scenarios where one model routes and another produces.

The D&C 107 ratios framework for models:

| Model | Ratio | Stewardship Level | Oversight |
|-------|-------|-------------------|-----------|
| Haiku 3.5 | 1:12 (deacons) | Task-level: classification, quick lookups, simple formatting | Review every output |
| Sonnet 4.6 | 1:48 (priests) | Feature-level: drafts, code generation, analysis | Spot-check, review at boundaries |
| Opus 4.6 | 1:96 (elders) | Domain-level: deep study, architecture, spec writing | Periodic audit, trust judgment |

These ratios aren't final — they're the starting hypothesis. The experiments will calibrate them. A model that proves reliable at a given task type earns wider autonomy at that task type (parable of the talents). One that produces unreliable output gets narrower scope (the stewardship contracts).

**Cost implications.** Michael notes that running Opus 4.6 exclusively probably isn't sustainable as work becomes more agentic. The consecration principle (Step 10): "Does every token serve the intent?" Using Opus for classification is like having the high priest judge every small matter — it's the pre-Jethro pattern. Haiku handles classification. Sonnet handles most production work. Opus handles what requires deep reasoning. "Every great matter they shall bring unto thee, but every small matter they shall judge" (Ex 18:22).

**Subscription planning.** Michael suspects next month will exceed Copilot's included credits. A Claude subscription with pay-per-token would allow the D&C 107 ratio model: spend the expensive tokens where they matter, use cheaper tokens for lower-stakes work. This decision is on the radar, not urgent.


## The 28% Problem — Updated Assessment

The squad analysis flagged that we practice roughly 28% of our own 11-step cycle. After this session, the updated assessment:

| Step | Previous | Current | What Changed |
|------|----------|---------|-------------|
| 1. Intent | Practiced | Practiced | intent.yaml in session-start, no change |
| 2. Covenant | Described | **Practiced** | covenant.yaml created, in session-start, in copilot-instructions |
| 3. Stewardship | Partial | Partial | Agent domains exist. Progressive tracking planned, not implemented. |
| 4. Spiritual Creation | Practiced | Practiced | Scratch files, outlines, specs |
| 5. Line Upon Line | Implicit | Implicit | Happens naturally, not structurally designed |
| 6. Physical Creation | Practiced | Practiced | We build things |
| 7. Review | Partial | Partial | Three-layer review not formalized |
| 8. Atonement | Emerging | **Practiced** | Section VII correction is a live example. Learning captured. |
| 9. Sabbath | Practiced | Practiced | Agent built, first session run |
| 10. Consecration | Infrastructure | Infrastructure | Token budgets exist. Model-tier experiments planned. |
| 11. Zion | Described | Described | Bishop/ward model proposed. Not yet implemented. |

Maybe 45% now. Covenant moved from described to practiced. The Section VII correction is itself Step 8 (Atonement) in action — failure became growth, the learning was captured, the covenant was refined. Steps 3, 5, 7, 10, 11 remain gaps.


## Connection to Ward Council

There's a thread between the study, the correction, and the ward conference experience that's worth naming. The study was about delegation and the failure to do it well. The correction was about a blind spot — treating AI delegation as lesser when it works by the same principles. The ward conference was about seeing the larger whole: not wards of Zion, but stakes of Zion.

The stewardship pattern looks different at each level of organization:

- **Individual agent** (ward): Clear stewardship, defined scope, local autonomy
- **Agent system** (stake): Shared intent, coordinated purpose, model-tier stewardship
- **The whole project** (Zion): "One heart and one mind" — every document, tool, agent, and session serving the same purpose

We're building a stake of Zion. Not quite there yet — the ward councils aren't all running, the model-tier stewardships aren't calibrated, the progressive trust tracking doesn't exist. But the covenant is written. The pattern is identified. And the 28% (now 45%) is growing because we're practicing what we preach.

D&C 107:99 holds: "Let every man learn his duty, and to act in the office in which he is appointed, in all diligence." Every agent. Every model. Every session. Learn the duty. Act in it.

---

*Study: [stewardship-pattern.md](stewardship-pattern.md)*
*Scratch: [.scratch/stewardship-pattern.md](.scratch/stewardship-pattern.md)*
*Covenant: [.spec/covenant.yaml](../.spec/covenant.yaml)*
