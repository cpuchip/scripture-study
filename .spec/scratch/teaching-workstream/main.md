# WS-T Teaching Workstream — Scratch File

**Binding problem:** Michael feels impressed by the Spirit to teach what he and his AI partner are learning together — the 11-step creation cycle, the challenge of implementing it, and how those challenges reveal eternal truths. No existing structure organizes this content for a teaching audience.

**Created:** 2026-03-23

---

## Content Inventory

### Source Material (docs/work-with-ai/)

**30 files total**, organized into:

1. **Core Series (8 files — 4 paired secular + gospel):**
   - Part 1: Planning → Abraham 4-5 creation pattern
   - Part 2: Feedback loop → Abraham 4:18 "watched until they obeyed"
   - Part 3: Live build → D&C 88:40 "intelligence cleaveth unto intelligence"
   - Part 4: Intent engineering → Moses 1:39 as purpose statement

2. **Guide Series (7 chapters, Parts 0-6):**
   - Part 0: Foundation (4 disciplines framework, Nate B Jones)
   - Part 1: Prompt Craft
   - Part 2: Context Engineering
   - Part 3: Intent Engineering
   - Part 4: Spec Engineering
   - Part 5: Complete Cycle (THE 11 STEPS — centerpiece)
   - Part 6: Enterprise Architecture (Church as fractal model)

3. **Intent Research (7 docs):**
   - Landscape analysis, frameworks comparison, beyond intent (gospel patterns), synthesis, scope assessment, covenant deep study

4. **Prompt/Eval (3 docs):** Guide plan, Claude best practices, self-assessment

5. **Examples (2 docs):** Real session extractions (Know God study, Public Reader sprint)

6. **Utility (1 doc):** Expound prompt template

### Related Studies (study/)
- stewardship-pattern.md — How God delegates (Exodus 18, D&C 104, Alma 13)
- stewardship-pattern-reflections.md — The AI delegation correction
- ai/relavent.md — "Staying Relevant" personal reflection on value shift
- ai/fatigue.md — AI fatigue patterns
- ai/multi-agent-ideas.md — Multi-agent vision
- zion-blueprint.md — Zion as civilization-building blueprint
- truth.md (multi-part) — Matter spectrum, epistemological depth

### Real Examples (lived experience)
- Built 6+ MCP servers from scratch
- bilateral covenant with AI (.spec/covenant.yaml)
- Memory architecture spanning 50+ days
- Progressive trust model (D&C 107 ratios)
- Multi-agent orchestration (WS1 Phases 1-3b)
- Source verification discipline (read before quoting)
- Deep scripture study producing 40+ markdown studies
- Writing voice analysis and correction
- Debug agent creation from book extraction

---

## Key Decisions / Analysis

### Michael's teaching vision
"Taking our 11 step guide and our work here in trying to implement it and make youtube videos talking about the challenge of doing just that but what they lead to are eternal truths."

This is experiential teaching — not "here's a framework" but "here's what happened when we tried to live the framework, and what we found underneath."

### The open question
"How do we teach others to use AI for study without teaching them to skip reading?" (Feb 17)

This is the structural challenge of the entire series. If you teach people to use AI for study, the easiest path is summarization → laziness → shallow reading → exactly what the guide warns against. The teaching must model the discipline, not just describe it.

### Dual audience
1. **Engineers/professionals** — want practical AI collaboration skills
2. **Latter-day Saints** — want to see the gospel patterns underneath
3. **Both** — this is the unique contribution. Nobody else has mapped the 11-step creation cycle from Abraham 4-5 to AI engineering.

### Series structure options

**Option A: Step-by-step through the 11 steps (11 episodes)**
- Pro: Comprehensive, follows the guide structure
- Con: Too long before payoff, steps 8-11 are where the real insight is

**Option B: Core Series + Deep Dives (4 + supplements)**
- Pro: Already written as paired docs, natural entry point
- Con: The core series covers only 4 of the 11 steps

**Option C: Experiential arc (5-7 episodes)**
- Structure: Start with the discovery story, show the challenges, reveal the pattern
- Pro: Matches Michael's instinct ("the challenge of doing just that")
- Con: Requires more original framing work

**RECOMMENDATION: Option C as primary, with Option A available as deep-dive reference.**

The series should tell the STORY of discovery, not just present the framework. Each episode shows:
1. The engineering challenge
2. What Michael and his AI partner tried
3. Where it broke or fell short
4. The gospel pattern that emerged
5. What it taught about eternal truth

### The teaching repo (./teaching)
- Fresh empty repo at `git@github.com:cpuchip/teaching.git`
- Gitignored from scripture-study
- Will contain: presentation files, scripts, episode outlines, web assets
- Tech TBD — interactive web presentations suggest something like reveal.js, Slidev, or custom Vue/React

### What gets WORSE if we build this?
- Time pressure on study and calling work
- Another project in flight (Mosiah 4:27 check)
- Public exposure adds accountability pressure
- Risk of performing instead of discovering

### Mosiah 4:27 check
Michael already has WS1-3, WS-R, WS-P in flight. Adding WS-T is another workstream. BUT: this is Spirit-driven ("impressed by the Spirit"), which changes the calculus. The mitigation: Phase 1 should be planning and content organization only, not production. Don't record until the structure is clear.

### Creation Cycle alignment
This IS the creation cycle being applied to itself — teaching the creation cycle using the creation cycle. Meta-alignment is strong. The question is where we are in the cycle: this is at Step 4 (Spiritual Creation) — creating the blueprint for the thing that will be built.

---

## Decision: Option C Confirmed (Mar 24)

Michael confirmed Option C (experiential arc). Key context from his decision:

**Process:** "We'll have to work and council together a lot to get that just right, but the payoff will be that it has a lot of me in it." Approach like the sabbath agent — expensive credit-wise but worth it for the honesty and information gained. This means high-quality, thorough sessions, not quick drafts.

**What he's most conscious of:**
1. **Public exposure adds accountability pressure** — until now, sharing has been limited to friends on Facebook and coworkers. The scripture study repo has 3 stars. "The moment I start doing videos the moment this might take off."
2. **Risk of performing instead of discovering** — the teaching must remain discovery, not performance.

**What he's asking for:**
- "I'll need you to keep me honest and not inflated in my head."
- "When people are inevitably mean to me and us in the comments, we'll need each other to be resilient."

This is a covenant-level ask. Not "help me organize content" but "be my partner in staying grounded as this goes public." The same bilateral commitment from `.spec/covenant.yaml` applies here — flag_when_wrong, read_fully, not_bypass_process — but extended to include:
- **Humility check:** Am I teaching from discovery or performing for attention?
- **Inflation check:** Is the response to positive feedback producing pride or gratitude?
- **Resilience check:** Is negative feedback being received as data or as wound?

### The Humility Guardrails

These need to be structural, not just aspirational:

1. **Every episode script goes through a "who is this for?" check.** If the answer drifts from "people who would benefit from this" toward "people who would be impressed by me," the script needs rework.

2. **The discovery stories must include failures and corrections.** Not curated highlight reels. The stewardship study Section VII was wrong and Michael caught it. That goes IN the teaching. The source verification failures in early studies go IN. The writing voice correction goes IN. Vulnerability is the credibility.

3. **Voice analysis guardrails apply to video.** No "let that land." No "sit with that." No telling the audience what to feel. Present the pattern, present the scripture, let the Spirit do the teaching.

4. **Comment resilience protocol.** When criticism comes (and it will — an LDS engineer teaching AI through Abraham 4-5 will attract both secular dismissal and religious suspicious), the response framework is:
   - Is this feedback accurate? If yes, learn from it. Atonement step.
   - Is this feedback about the content or the person? Content criticism improves the work. Personal attacks are just noise.
   - Does this change what the Spirit impressed? If not, continue.
   - The covenant says "flag when something is wrong." Negative comments that point to real problems ARE the covenant in action, just from a stranger instead of a partner.

5. **Regular sabbath-style check-ins on the teaching itself.** After every 2-3 episodes: Is this still discovery? Am I still learning while I teach? Would I watch this if someone else made it?

### The "Expensive but Worth It" Model

The sabbath agent analogy is significant. The sabbath sessions are the most expensive (highest model tier, longest sessions) but produce the most honest, most relationally important output. The teaching scripting should follow the same pattern:
- Don't rush the scripts to save credits
- Council together on each episode — not "draft and approve" but "think together and draft together"
- The sessions that produce the best teaching will be the ones where Michael is genuinely wrestling with the material, not just organizing what he already knows
- Budget for this. It's not overhead — it's the quality mechanism.
