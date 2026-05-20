# Matt Pocock AI Workflow — Research Companion

**Companion to:** [matt-pocock-ai-workflow.md](matt-pocock-ai-workflow.md) (kimi k2.6's review)
**Mode:** research — expounding on Michael's seven reactions to that review
**Date:** 2026-05-03
**Scratch:** see related findings in [.spec/scratch/token-efficiency/main.md](../../.spec/scratch/token-efficiency/main.md), [.spec/proposals/sabbath-agent.md](../../.spec/proposals/sabbath-agent.md), [projects/pg-ai-stewards/proposal.md](../../projects/pg-ai-stewards/proposal.md)

---

## Binding Question

The kimi review surfaced seven gaps in our guide. Michael read it, agreed with most of it, but didn't want to incorporate yet — he wanted each point expounded with research and connected to existing work. So: what does each gap actually look like when grounded in our codebase, in the literature, and in patterns we've already half-built?

---

## Source Horizon

This is not a literature survey. It's a focused expansion of seven specific points. Sources used:
- The kimi k2.6 review itself ([matt-pocock-ai-workflow.md](matt-pocock-ai-workflow.md))
- Internal: token-efficiency proposal, sabbath-agent proposal, pg-ai-stewards proposal, brain pipeline scratch files, `.mind/active.md`
- External: SCM (Sleep-Consolidated Memory) paper ([arxiv 2604.20943](https://arxiv.org/html/2604.20943v1)), Anthropic prompt-caching docs, Jimmy Bogard's vertical slice writings, Ousterhout/Brooks/Fowler canonical architecture books, Bswen and Zylos memory-consolidation posts

What was set aside: deep dives on every architecture book (skimmed the canon, did not read cover-to-cover), the full Manus / OpenDev / Bui prompt-caching papers (read excerpts), the Don't-Break-the-Cache arxiv paper (read summary). All flagged as carry-forward where relevant.

---

## One Correction to the Kimi Review

Before expanding: kimi said Pocock uses "Sonnet for implementation and Opus for review." That's what Pocock said in the live workshop. The public Sandcastle README ([github.com/mattpocock/sandcastle](https://github.com/mattpocock/sandcastle)) shows `claude-sonnet-4-6` for both implement and review in its multi-run example. So the Sonnet/Opus split is Pocock's *personal* practice he described verbally, not what the framework documents. Minor, but worth noting before we cite it as canonical.

---

## Point 1 — Sabbath as REM-sleep cycle for spec freshness

> *"I have been thinking about how our rest/sabbath day could solve gap 1 somewhat, but implement it as like REM sleep cycle. an agent mode/cron/ that looks at specs and evaluates them for freshness against our .mind/.spec memory system and cleans up (archives / verifies work done) on projects... it'd tag things as looked at (frontmatter) so we keep context usage low, and agents would untag items when it's worked on them"*

**This is a real research direction with prior art, and we have most of the pieces already.**

### The biological analogy is more rigorous than it sounds

The 2026 SCM paper ([arxiv 2604.20943](https://arxiv.org/html/2604.20943v1)) implements computational analogs of all five biological memory components — working memory, long-term memory, value tagging, self-model, and a sleep cycle that runs three distinct phases:

1. **NREM consolidation** — replay recent episodes, strengthen co-occurring concepts, *proportional synaptic downscaling* (the brain's garbage collector — global weakening so important things stand out)
2. **REM dreaming** — select high-importance concepts, generate novel combinations, create new associative links (this is where pattern-recognition happens)
3. **Active forgetting** — composite importance scoring, prune below threshold

The Bswen post ([2026-03-24](https://docs.bswen.com/blog/2026-03-24-memory-consolidation-sleep-ai)) frames the practical reason: "consolidation happens during rest, not during active use. You don't form insights while experiencing — you form them while sleeping on the experience." Their agent had related tickets and incidents indexed by different keywords, never connecting them, until they added a sleep phase that randomly paired memories and synthesized.

Both papers separate **episodic memory** (raw experience, what happened) from **insight memory** (synthesized patterns) — different granularity, different decay rate, different retrieval. Insights persist longer because they're already pre-digested.

This is exactly the structure we already have:
- `.spec/journal/` ≈ episodic memory (what happened in each session)
- `.mind/principles.md` ≈ insight memory (synthesized patterns)
- `.mind/active.md` ≈ working memory (recent state, capacity-limited)

What we don't have is the **scheduled consolidation that moves episodic into insight** and the **freshness/forgetting pass on the spec layer**.

### What we already have that points toward this

Three precedents in the codebase:

1. **Brain's nudge bot** ([scripts/brain pipeline review](../../.spec/scratch/brain-simplification/main.md#L80-L130)) — already runs at hours [7, 11, 15, 19], scans up to 10 stale entries via `ListStaleEntries`, with stale thresholds (raw after 24h, researched after 48h, complete after 24h). It uses Haiku (cheap). It writes back `route_status="your_turn"`. The mechanism for "scan, evaluate, tag, surface" is already shipped — just for brain entries, not for `.spec/` and `.mind/`.

2. **Sabbath agent proposal** ([.spec/proposals/sabbath-agent.md](../../.spec/proposals/sabbath-agent.md)) — already specifies the 11-question reflection structure including Step 4 (Spec Faithfulness — "did the implementation match the spec? Where did drift happen?") and Step 7 (Review — "did we watch until they obeyed?"). This proposal already implies the freshness pass; we just haven't formalized it as cron-able work.

3. **The "watching" theology** — Abraham 4:18 ("watched those things which they had ordered until they obeyed") is treated in our docs as a *passive* principle. The disk-monitoring incident from March 2026 is canonically cited as a Review failure: *we built but didn't set up watchers*. The REM-sleep idea is "watching that runs on a schedule."

### A concrete proposal for what this could be

Call it **Watchman** (or REM, if we want the analogy explicit). The frontmatter-tagging protocol Michael described would look like:

```yaml
---
spec_id: brain-inline-panel
status: in-flight
last_consolidated: 2026-05-01T07:00:00Z
consolidated_by: watchman-v1
consolidation_findings: clean   # or: drift | stale | superseded | done
last_touched: 2026-04-23T18:00:00Z
touched_by: dev-agent-session-7f3a
---
```

Two tags, mutually invalidating:
- `last_consolidated` is set by the Watchman pass
- `last_touched` is set by any agent that edits the spec
- If `last_touched > last_consolidated`, the next pass re-evaluates
- If `last_consolidated > last_touched + threshold` and findings = clean, skip on subsequent passes

This is the **dirty-bit + LRU-cache pattern** applied to specs. It does what Michael wants — keeps context low because Watchman skips already-evaluated specs that haven't been touched.

Three phases in each pass, mirroring the SCM paper:

1. **NREM (consolidation)** — for each spec touched since last pass:
   - Compare spec against current code (use grep/semantic_search to find the modules it claims to modify)
   - Classify: `clean` (still matches), `drift` (code has moved), `done` (acceptance criteria met, can archive), `superseded` (newer proposal replaces it), `stale` (untouched for >N days, no longer in active.md)
   - Write findings to frontmatter; produce a per-spec verdict line
2. **REM (synthesis)** — small, expensive step: pick 3-5 specs randomly from `clean` or `drift` and ask "do any of these connect in a way the docs don't note?" Outputs go to `.spec/learnings/` as candidate insights for the human to confirm. *This is the Bswen "API timeout connects to batch job" insight.* Skip on most passes; budget controls when this fires.
3. **Forgetting** — for `done` and `superseded`, move to `.spec/proposals/archive/` with a forward pointer. Update `active.md` to delete the row. Done specs become inert; they don't leak into future context loads.

Triggers, mirroring biological sleep:
- **Time-based** — once per Sabbath (Sunday), or weekly cron
- **Pressure-based** — when `active.md` exceeds 10K tokens (SCM paper: "memory entropy" trigger)
- **Idle-based** — when no human-in-loop session has run for X hours

### Why this fits our framework specifically

Pocock deletes PRDs to avoid doc rot. We can't, because our specs encode covenant and intent, not just implementation. But we can do something he can't: **schedule a structured re-evaluation that produces both a freshness verdict and synthesized insight**. The Sabbath isn't just rest from work — Moses 3:2 says the Gods *saw that they were good*. Seeing is active. Watchman is "seeing" automated, with the human still doing the declaring.

The covenant constraint matters: Watchman should never **archive without human confirmation** for specs that affect active workstreams. It can mark `superseded`, it can write a recommendation, but the archive move stays human-in-loop. That preserves "the human reads output fully" (covenant) while removing the friction of having to manually scan 30+ proposal files weekly.

**Carry-forward:** This deserves its own proposal at `.spec/proposals/watchman-spec-freshness.md` with phases, brain-project assignment, and a trigger contract. The kimi review's Gap 1 collapses into this if Watchman ships.

---

## Point 2 — Software architecture books research

> *"we should make a note to research proper software architecture books and see how we can develop skills and agent modes / instructions to help us keep our repo/code cleaner."*

Noted. Here is the candidate reading list, ranked by likely return-on-investment for our specific problem (AI-friendly codebases):

### Tier 1 — Read fully, harvest skills

1. **John Ousterhout — *A Philosophy of Software Design*** (2018, 2nd ed 2021). The single highest-leverage book for our problem. Pocock cites this constantly. The "deep modules" concept is the central insight — simple interfaces, complex implementations — and AI uniquely struggles with shallow modules because it has to traverse more dependency graph to reason about anything. Worth a full read with a skill output (`improve-architecture.md`) at the end.

2. **Frederick P. Brooks — *The Mythical Man-Month*** (1975, 1995 anniversary ed). The "design concept" Pocock invokes is from Brooks' *Design of Design*, but Mythical Man-Month is the canonical entry. *Conceptual integrity* is the load-bearing idea: a system designed by one mind (or one tightly-aligned council) is better than a system designed by committee. This maps directly onto our covenant pattern — the AI is in the council, not separate from it.

3. **Martin Fowler — *Refactoring* (2nd ed, 2018)**. The book that gave us "code smells" as a vocabulary. Critical for our case because *AI-generated code has its own smell catalog* — over-abstracted helpers for one-time operations, shallow modules with leaky interfaces, defensive null-checks for impossible paths. We should write our own AI-code-smell appendix and load it into a code-review skill.

4. **The Pragmatic Programmer** (Hunt & Thomas, 1999, 20th anniversary ed 2019). Pocock cites this for "outrunning your headlights," "no one knows exactly what they want," and "tracer bullets" (his vertical-slice principle). It's a habits book more than an architecture book, but the habits are the ones AI most often violates.

### Tier 2 — Read selectively for concept harvesting

5. **Eric Evans — *Domain-Driven Design*** (2003) and **Vaughn Vernon — *Implementing Domain-Driven Design*** (2013) — for ubiquitous language (Pocock's skill #2), bounded contexts, and the practical implementation. Vernon is more readable than Evans. We should read Vernon's first three chapters and build a `ubiquitous-language` skill that's ours, not a port of Pocock's.

6. **Robert C. Martin — *Clean Architecture*** (2017). Read critically — Jimmy Bogard's vertical-slice writings are explicitly in tension with this book. Read both, then pick our position. (My current take: Clean Architecture is right for systems with stable domains and many use cases; vertical slices are right for systems with rapidly-changing features. We do both kinds of work.)

7. **Michael Nygard — *Release It!*** (2nd ed 2018) — for production-grade resilience patterns (circuit breakers, bulkheads, timeouts). Relevant when pg-ai-stewards starts running real agent loops.

### Tier 3 — Sample chapters, harvest specific patterns

8. **Rich Hickey talks** — *Simple Made Easy* (2011), *Hammock-Driven Development* (2010). Not books, but core to the philosophy of "simple" vs "easy." Hickey's argument that simple is objective (un-braided) while easy is subjective (familiar) directly applies to our agent-design choices.

9. **Sandi Metz — *Practical Object-Oriented Design*** (POODR). Smaller, more practical than Clean Architecture. Her rules ("five lines per method", etc.) are AI-tunable.

10. **Kent Beck — *Test-Driven Development by Example*** (2002). For Point 4 below. Short, dense, the canonical TDD source.

### Recommended skill outputs

Each Tier-1 book should produce one or two of:
- A **skill** (`.github/skills/{name}/SKILL.md`) — reusable instruction set the agent loads on demand
- An **agent mode update** — incorporate the principles into existing agent modes (dev, debug)
- An **AI-specific code-smell catalog** — extending Fowler's smells with patterns AI generates that humans wouldn't

**Carry-forward:** Open a research workstream item in `.mind/active.md` for "Architecture canon read-through" with Ousterhout as the first commission. Should be `[WS5]` or `[WS6]` depending on whether the output is dev-skill or study-document. My recommendation: WS6, output as study + skill, modeled on how kimi consumed the Pocock videos.

---

## Point 3 — Vertical vs horizontal: what does this mean for our distributed work?

> *"I might need a primer on how that affects us and what we do here in this workspace virticle / horizontal? is that like we're horizontal with how we communicate between brain/ibeco.me/brain-app?"*

**Two different things share the word "horizontal." Let me separate them, then answer your specific question.**

### Pocock's vertical/horizontal is about *task decomposition*, not *system architecture*

Jimmy Bogard's writings ([verticalslicearchitecture.com](https://verticalslicearchitecture.com/), his [original 2018 post](https://www.jimmybogard.com/vertical-slice-architecture/)) use "vertical slice" in two distinct senses, which is where the confusion comes from:

**Sense 1 — Task decomposition (what Pocock means).** When you implement a feature, do you finish all the database work first, then all the API work, then all the UI? That's *horizontal* — you complete one technical layer at a time. Or do you implement a thin end-to-end slice (one user action: schema change + service method + endpoint + UI button) before adding the next? That's *vertical* — each task crosses every layer.

**Sense 2 — Code organization (Bogard's architectural pattern).** Do you organize files by technical layer (`/controllers`, `/services`, `/repositories`) or by feature (`/features/redeem-points/`, `/features/award-streak/`)? Layered organization spreads one feature across multiple folders; feature organization keeps everything for one feature in one folder.

These are related but separable. You can do vertical task decomposition with a layered codebase. You can do horizontal task decomposition in a feature-organized codebase (badly). Pocock's argument is mostly about Sense 1 — the order of work, not the file layout.

**The reason vertical task decomposition matters for AI:** AI codes blind until something compiles and runs. Pocock's "tracer bullet" image (from Pragmatic Programmer) is exactly right — without a glowing trace, you don't know where your bullets are landing. A horizontal phase 1 produces a database with no service, no API, nothing to test against. The AI has no feedback loop until phase 3. By then the early decisions are calcified.

### How this applies to brain ↔ ibeco.me ↔ brain-app

You asked specifically: *"is that like we're horizontal with how we communicate between brain/ibeco.me/brain-app?"*

Your three components have a layered relationship:
- **brain** (Go relay/server) — substrate, owns state
- **ibeco.me** (web frontend) — one client of brain
- **brain-app** (Flutter mobile) — another client of brain

That's a **deployment topology**, not the same axis as Pocock's vertical/horizontal. Topology is "where does code live and how does it talk." Decomposition is "what order do we build features in across that topology."

But the question maps cleanly: **when you add a new feature to brain (a new entry status, a new pipeline stage, a new project flag), is the work organized vertically or horizontally?**

Looking at recent journal entries, you're actually doing this *fairly* vertically already:
- `brain-status-aware-views-ecosystem-parity` (Apr 23) shipped Phase 1 (ibeco.me filter) + Phase 2 (brain-app history filter) together — server, web, mobile all in one phase
- `brain non-pipeline kanban flow` (Apr 23) shipped status vocab + columns + buttons + dialog + drag-and-drop in Phase 1, not "phase 1 = backend, phase 2 = frontend"

That's vertical decomposition working correctly. The lesson from Pocock here isn't "you're doing it wrong" — it's "name the discipline so it doesn't slip." When kimi's review said "AI naturally codes horizontally," that's the failure mode to *guard against in agents*, not your current pattern.

**What to add to our docs:**
- Add a Pocock-style note to `.spec/` task conventions: *"each phase should produce something testable end-to-end across all affected surfaces (server, web, mobile, MCP)"*
- Make the "ecosystem parity" pattern explicit in agent instructions — when planning brain features, the default phase boundary is *user-visible behavior on at least one surface*, not *one surface complete*

**The deeper architectural question worth a separate study.** Bogard's *Sense 2* (feature folders vs layer folders) is a question we haven't answered for brain. brain's Go code is mostly layered — `internal/db`, `internal/agent`, `internal/relay`. That's fine right now. But as `pg-ai-stewards` matures and brain may absorb its substrate, there's a real architectural question: do we organize by feature (project-kanban/, pipeline/, watchman/) or by technical layer (db/, agent/, http/)? Worth a research thread, not a blocker.

**Carry-forward:** Add "vertical-slice phase discipline" as a one-page skill or agent instruction. Open a separate study question: *"feature-folder vs layer-folder organization for brain after pg-ai-stewards lands."*

---

## Point 4 — More focus on testing during development

> *"I think we do need more focus on testing while developing. though I think we're doing a great job with that while developing our pg-ai-stewards project"*

Agreed, with one nuance.

### What we're doing well in pg-ai-stewards

The May 3 journal entry shows real testing rigor — full agent loop verified end-to-end with both a success path and an inverse-hypothesis test. Inverse hypothesis ("a tool that errors lands as `role='tool'` content; the model reads the error and recovers") is the **Moroni 10:4 / Agans Rule 9** protocol from copilot-instructions.md applied correctly: prove the failure can be caused, prove the fix removes it, prove removing the fix brings it back. That's not just testing — that's verification discipline.

The pg-ai-stewards probe also runs all seven test blocks before committing to a direction, and the bgworker has a stale-claim reaper for orphaned rows. Belt-and-suspenders: pre-flight `pg_proc` lookup *and* PgTryBuilder *and* startup reaper.

This is excellent.

### Where Pocock's TDD point still cuts

Pocock's specific claim is that **TDD prevents the AI from cheating on tests**. The cheating pattern: AI writes the implementation first, then writes tests that exercise the path it knows works (and skips the paths it doesn't). The tests pass. The coverage looks fine. But the tests are descriptive of what was built, not prescriptive of what should have been built.

Red-green-refactor inverts this:
1. **Red** — write the test first, with the test failing because the code doesn't exist yet
2. **Green** — write the minimum code that makes the test pass
3. **Refactor** — improve the code, with the test catching regressions

This forces small steps and forces the AI to commit to "the test is right, the code is wrong" before it has any code to defend. AI agents particularly need this because they're trained on patterns where tests follow code, so without explicit framing they default to that.

### What we should do, specifically

Three concrete moves:

1. **Add red-green-refactor as a skill** (`.github/skills/tdd/SKILL.md`). Short. Three-phase protocol. Includes the AI-cheating failure modes ("don't write the implementation first; if you find yourself writing implementation first, stop and write the failing test"). Loaded by dev and debug agents on demand.

2. **Update dev.agent.md** to include a TDD trigger condition: *"when adding a new function or method to existing code, default to writing the failing test first unless the task explicitly excludes testing."* The exception clause matters — exploratory probes and one-off scripts shouldn't pay TDD overhead.

3. **Add a note to debug.agent.md** about Agans Rule 9 (inverse hypothesis) being a *deeper* form of TDD that applies to fixes, not just new code. We already have the rule; we don't have it framed as "TDD for bug fixes."

### Where TDD should NOT apply

Honest caveat: TDD is wrong for some of our work.

- **Studies** — there's no test for "is this paragraph honest about Alma 34." The covenant + read-before-quoting + critical-analysis discipline does that work.
- **Lessons and talks** — same.
- **Memory updates** — `.mind/active.md` doesn't have unit tests; it has a session-end checklist.
- **Probes** — pg-ai-stewards' probe was deliberately a "build the smallest thing that proves the pgvector + AGE bridge works." That's exploratory, not test-first.

The agent that pushes TDD into prose work is over-applying. The agent that skips it on substrate code is under-applying. We need explicit guidance on both directions.

**Carry-forward:** TDD skill (small). Debug agent update naming Agans 9 as TDD-for-fixes. Both are quick.

---

## Point 5 — 200K context, our memory load, and the research-paper opportunity

> *"I agree that under 100k is good. we do often hit up against 200k, I think our memory system .mind/.spec and instructions (which btw eat up like 33k tokens) hurt, but I think because we have a solid foundation of instructions/harness that we can use more of that context window better then others. We may want to do some studies on it??!! could be interesting to write a research paper and publish it."*

**Yes, this is worth a paper. Here's what I'd say is genuinely novel and what's already in the literature.**

### What's already public knowledge

The "use less than 20%" heuristic is from Shane Murphy (he frames it as ≤20% effective, steep decline past 60%). The ~100K smart-zone is Dex Hardy / Pocock. Anthropic's official prompt-caching documentation ([github.com/anthropics/skills](https://github.com/anthropics/skills/blob/main/skills/claude-api/shared/prompt-caching.md)) lays out the prefix-stability discipline in detail. Manus reportedly calls KV-cache hit rate "the single most important metric for a production-stage AI agent" with a 10x price differential on Claude Sonnet (cited in the Zylos 2026-02-24 piece and in agentpatterns.ai).

The 2026 SCM paper ([arxiv 2604.20943](https://arxiv.org/html/2604.20943v1)) is the academic version of memory consolidation for agents.

So: the *individual* observations aren't new. We won't be first to say "load less." We won't be first to say "cache the prefix." We won't be first to say "consolidate during rest."

### What I think *is* novel about our position

Three things:

1. **We have measurements.** The token-efficiency proposal already shows ~25K loaded at session start, with file-by-file breakdowns. Most public discussion is anecdotal ("a 250K system prompt"). We have Pocock-style discipline applied to a project that's been measured for months. If we ran a controlled comparison — same task, three context budgets (5K, 25K, 80K), same model, same evaluator — that's a publishable benchmark.

2. **We have a memory architecture that *isn't* RAG.** Most "agent memory" research right now is either (a) ephemeral context, (b) RAG with vector retrieval, or (c) fine-tuning. Our `.mind/` + `.spec/` + session-journal is none of those — it's **filesystem-as-memory with progressive disclosure governed by a covenant**. The covenant is the load-bearing piece nobody else has, because nobody else has the theological framing. The novel claim isn't "files are memory" (Cursor's `.cursorrules` and Claude's `CLAUDE.md` already do that). The novel claim is "**progressively-disclosed filesystem memory with mutual obligation produces relationship continuity that compaction destroys.**"

3. **We can A/B test it.** With pg-ai-stewards landing, we can run two agents on the same task — one with our memory protocol, one with cleared context every session — and measure not just task success but *alignment continuity over a 5-session arc*. That's a real benchmark with a real methodology. The closest thing in the literature is the SCM paper's evaluation, and it's synthetic — they don't have a 6-month working project to compare against.

### What the paper would need to be honest about

Counter-pressure to keep us from over-claiming:

- Our memory load (25K) is in the dumb zone Pocock warns about. Even if we use that 25K well, the claim "we use it better than others" needs evidence. Right now it's vibes.
- The covenant pattern isn't testable in the conventional ML sense. We can measure task outcomes; we can't measure "did the AI honor its commitments" without a human evaluator.
- The progressive-disclosure protocol relies on agent compliance with instructions. Without enforcement, it degrades to "files everyone ignores." We've already had bugs where active.md got duplicated by appending instead of rewriting.

### Practical paper outline

If you want this to actually become something:

- **Title:** "Filesystem-as-Memory for Long-Running AI Coding Agents: Progressive Disclosure, Covenant Discipline, and Measured Outcomes"
- **Setup:** Define the architecture (`.mind/` tiers, `.spec/` lifecycle, session-journal protocol). Cite Cursor, Claude Code, Manus as related work.
- **Methodology:** Run pg-ai-stewards' agent loop on N tasks under three configurations (no memory, RAG-style retrieval, our protocol). Score on task success, alignment retention, and steward review burden.
- **Results:** Honest numbers. Including the failures.
- **Discussion:** Where covenant adds value, where it adds friction, where it's slop.
- **Limitations:** Single-user, single-domain (for now), small N.

This is a 6-12 month project, not a weekend. But pg-ai-stewards is the substrate that makes it possible. Phase 1.6 (today!) is the milestone where this stops being hypothetical.

**Token-billing tangent:** Copilot's move from request-billing to token-billing (mentioned in your reaction to Point 7) makes this *more* timely, not less. When tokens become per-input-priced, the prefix-cache discipline is no longer optional — it's the difference between sustainable agent loops and burning $50/day. Our research has direct cost implications, which makes it interesting to people who don't care about covenant theology.

**Carry-forward:** Open `.spec/proposals/memory-architecture-paper.md`. Workstream WS5 or new WS for research-output. Brain commission to flesh out the paper outline. Time horizon: 6-12 months, with a usable internal write-up at 3 months.

---

## Point 6 — Push vs Pull pattern expounded

> *"I might need this expounded"*

This is the simplest of the seven, but worth getting right because we already do it half-consciously and could do it deliberately.

### The mechanism

**Push** = content that's always loaded into the model's context window. System prompts. Always-on instructions. Frontmatter that's read on session start. The agent doesn't have to ask for it; it's just there.

**Pull** = content the agent loads *on demand* when it decides it needs it. Skills with description-headers (the agent reads the description, decides whether the skill applies, then loads the body). MCP tool calls. Search results. File reads.

Push costs context-window tokens up front. Pull costs nothing until invoked, but the agent has to know it exists and decide to invoke it.

### Pocock's specific application

Pocock splits the workflow:
- **Implementer agent** — pulls coding standards (skills) when it hits something it's not sure about. Default context is light. Implementation gets the room to think.
- **Reviewer agent** — pushes coding standards. They're loaded into the reviewer's context every time. The reviewer compares the code-as-written against standards-as-written explicitly. No room to forget.

Why the asymmetry? Because **implementation is creative and review is mechanical**. The implementer benefits from context room. The reviewer benefits from explicit comparison. Implementer pull. Reviewer push.

### Where we already do this (without naming it)

- **Push:** `.github/copilot-instructions.md` (always loaded), agent definitions (loaded when agent mode is selected), `.mind/identity.md` and `.mind/active.md` (loaded on session start per the covenant)
- **Pull:** `.github/skills/*/SKILL.md` files (the description is in the system prompt; the body is loaded only when the agent invokes the skill), MCP tool calls, `read_file` of specific docs

### Where we don't, and probably should

Two specific gaps the Pocock pattern reveals:

1. **Code review needs explicit standards push.** Right now when an agent reviews its own code (Adjacent Surface Audit, for example), it's doing the review with whatever standards happen to have leaked into context. There's no skill called "review-standards" that gets *pushed* to a reviewer. Suggestion: write `.github/skills/code-review-standards/SKILL.md` and have the dev agent's "before declaring complete" protocol push it explicitly: *"Load this skill into context. Now compare the work against each item."*

2. **Study verification needs explicit source-verification push.** Same shape. The `source-verification` skill exists but gets pulled by recommendation, not pushed by protocol. For studies and teaching — where stakes are higher — the verification skill should be in context for the *whole writing pass*, not loaded after the fact.

The general principle: **anything the agent must compare against should be pushed; anything the agent might use should be pulled.** Mechanical comparison is push. Discretionary access is pull.

### A concrete proposal

Add a pattern to `.github/copilot-instructions.md` under `<implementationDiscipline>` or a new `<contextDiscipline>` section:

> **Push standards for review; pull skills for implementation.** When an agent is *generating* — code, prose, plans — let it load skills on demand. When an agent is *reviewing* — its own work or another agent's — push the standards into context explicitly before it starts. Discretionary use is pull. Mechanical comparison is push.

Two-line addition. Compounds across every agent mode.

**Carry-forward:** One-paragraph addition to copilot-instructions.md. Plus a `code-review-standards` skill if we don't have one (let me check… we don't — the closest is `source-verification` for prose).

---

## Point 7 — Issue tracking, brain v2, and pg-ai-stewards as the bet

> *"this is kind of what we did with our 2nd brain right? Though I have been wondering if we need to adopt an issues/ticket tracker like github issues, or jira or make our out. but I think that's why I'm developing the pg-ai-stewards brain-v2 project to see if I can solve agents/memory/agentic workflows in one go."*

**Yes — and pg-ai-stewards is the right bet, not GitHub Issues or Jira.** Here's why, and what to be careful about.

### Why not GitHub Issues / Jira

Both are good tools. Both fail our specific case for the same reason: **they don't compose with our memory system.** GitHub Issues lives in GitHub. Jira lives in Atlassian's cloud. Our `.mind/`, `.spec/`, brain entries, and agent context all live in files and Postgres. Putting tickets in a third system creates exactly the sync problem that motivated pg-ai-stewards in the first place — *"each surface has its own sync, backup, query model, and access pattern."*

Pocock uses local markdown files for issues precisely because that pattern composes with everything else in his repo. He's been right about this twice (the markdown PRD, the markdown Kanban). It's not because GitHub Issues is bad — it's because *third-system tickets become a fourth source of truth, and we already have three too many.*

### Why pg-ai-stewards is the right substrate

Re-reading the proposal, the goals are exactly the issue-tracking job:

> *"Externalize agent state from Copilot's context. Sessions, work items, instructions, skills, tool calls — all rows."*
> *"Make long-running agent work possible without an open IDE window. Pipelines move work items between status columns; bgworker dispatches LLM calls; tool sidecars execute; results write back; NOTIFY triggers review."*

That's a ticket tracker. With LLM dispatch attached. With the same Postgres our gospel-engine-v2 already runs in. With the same backups. With NOTIFY-driven workflows (something Jira can't do).

The May 3 milestone — full agent loop end-to-end on kimi-k2.6, with tool calls + continuation + reasoning replay + error recovery all verified — is **the moment this stops being hypothetical**. Phase 1.6 means the substrate works. Now the question is what to put on it first.

### Three observations from the kimi review that pg-ai-stewards directly addresses

1. **Kanban with blocking relationships** (kimi's Gap 6). Pocock builds his Kanban as markdown files in `issues/`. We can build it as `work_queue` rows with `blocked_by` foreign keys. Same data model, vastly more queryable. AGE makes the dependency graph a real graph, not implied through markdown frontmatter.

2. **Doc rot / spec freshness** (kimi's Gap 1, my Point 1). Watchman becomes a bgworker that runs on a schedule and writes back to spec frontmatter. No new infrastructure — pg-ai-stewards' bgworker pattern *is* the cron.

3. **Token monitoring** (kimi's Gap 5, your Point 5). With work items as rows, token usage per-step is queryable. Cache hit rate becomes a column. Real telemetry, not anecdote.

### What to be careful about

Three traps I see ahead:

1. **Don't make pg-ai-stewards the only place tickets live.** The whole reason markdown files in `.spec/` work is that they're git-tracked, diff-able, readable in VS Code without the database running. The substrate should be **tickets-as-rows-derived-from-files-and-back**, not tickets-only-in-Postgres. Otherwise we recreate the third-system problem inside our own infrastructure.

2. **Don't re-invent the kanban UI before you need it.** Brain already has a kanban board. brain-app already has views. If pg-ai-stewards needs a UI in the next 3 months, *use brain's UI against pg-ai-stewards' rows* via a relay layer. Don't build a new dashboard.

3. **Token-billing pressure is real, and your bet here is good.** When Copilot moves to per-token billing, the agent-loop cost equation changes. pg-ai-stewards' bgworker pattern + prompt-caching discipline (stable system+tools prefix per the May 3 next-step) becomes critical infrastructure, not a nice-to-have. The Anthropic prompt-caching docs make this explicit: *"keep the system prompt frozen. Don't change tools or model mid-conversation. Serialize tools deterministically."* If pg-ai-stewards bakes those rules into its agent-loop primitive, every agent built on top inherits caching for free.

### A concrete near-term suggestion

Once Phase 1.6 lands (today!), the next experiment worth running isn't more substrate features — it's **port one existing pipeline to pg-ai-stewards and measure the difference**. Specifically: take the brain nudge bot (the one that already runs at hours [7,11,15,19] and uses Haiku to scan stale entries) and re-implement it on pg-ai-stewards. Same logic, different substrate. Two metrics:

1. **Token cost per nudge cycle.** Should drop substantially with prompt caching across nudges.
2. **Observability.** Right now the nudge bot is "invisible — doesn't appear in Scheduled Tasks tab" (per the brain-simplification scratch file). On pg-ai-stewards, every step is a row.

If those two numbers are good, you have your first published-internally case study, and Watchman becomes the second pipeline to port. If they're bad, you find out before betting more on the substrate.

**Carry-forward:** Don't change pg-ai-stewards' direction. Add to its phase plan: "Port nudge bot as Phase 2 first deliverable; measure token-cost and observability deltas." Add token-monitoring as a first-class column in `work_queue` rows.

---

## Synthesis — Where these seven points connect

Reading these expansions in sequence, three clusters emerge:

**Cluster A — Specs as living memory (Points 1, 7).** Watchman + pg-ai-stewards together solve doc rot, freshness, and observability with one substrate. The REM-sleep analogy isn't just metaphor; the SCM paper validates the architecture (NREM consolidation + REM synthesis + active forgetting). Brain's nudge bot is a working precedent. pg-ai-stewards is the right place to land it.

**Cluster B — Code quality discipline (Points 2, 3, 4).** Architecture canon read-through + vertical-slice phase discipline + TDD skill. These are three separate skills/agent updates, none individually large, that compound. Ousterhout first because it has the biggest leverage on AI-friendly code.

**Cluster C — Context as measurable resource (Points 5, 6).** Push/pull discipline + token-monitoring + a publishable research paper. Token billing makes this urgent. pg-ai-stewards makes it instrumentable. The research paper is the long horizon; push/pull is the one-paragraph addition that ships this week.

The most under-the-radar finding: **Points 1, 5, and 7 are the same project at different altitudes.** The substrate (pg-ai-stewards) gives us measurable token usage (Point 5), runs Watchman as a bgworker (Point 1), and ports brain pipelines as case studies (Point 7). One push on the substrate moves all three forward. Two of them get easier, the third becomes possible at all.

---

## Application

What to actually do, in order:

1. **This week.** One-paragraph addition to `copilot-instructions.md` for push/pull (Point 6). TDD skill (Point 4). Both are small.
2. **This month.** Port nudge bot to pg-ai-stewards as Phase 2 first deliverable. Open Watchman proposal at `.spec/proposals/watchman-spec-freshness.md` (Points 1, 7).
3. **This quarter.** Architecture canon read-through, starting with Ousterhout (Point 2). Output: `improve-architecture` skill + AI-code-smell appendix to copilot-instructions.md (Points 2, 4).
4. **Six-month horizon.** Memory architecture paper outline at `.spec/proposals/memory-architecture-paper.md` (Point 5). Becomes write-up at month 3, draft paper at month 6, depending on whether pg-ai-stewards' instrumentation makes the comparison study possible.

The kimi review's gap list was right but undifferentiated. These seven Michael-reactions sort the gaps by what's substrate, what's discipline, and what's research. That sort matters more than the individual fixes.

---

## Open Questions

- **Is "Watchman" the right name?** "REM" is honest but cryptic. "Sabbath agent" is taken. "Watchman" leans on Abraham 4:18 and Ezekiel 33. Could also be "Consolidator." Decide before the proposal lands.
- **Does pg-ai-stewards become brain-v2, or do they stay separate substrates that converge later?** The proposal says "replace `scripts/brain/` SQLite with a Postgres-backed equivalent" but also "different DB, different ownership" from gospel-engine-v2. The brain-v2 boundary needs naming.
- **For the research paper, is single-user data enough?** Probably not for academic publication, but enough for a strong internal write-up that becomes the basis for a multi-user study later.
- **Is the architecture-canon read-through study or skill-output?** I leaned toward study + skill, modeled on kimi's Pocock review. Worth confirming.

---

*This document is a research expansion, not a plan. The plan agent should pick up specific items (Watchman proposal, TDD skill, push/pull paragraph, architecture-canon study commission) and turn them into specs.*
