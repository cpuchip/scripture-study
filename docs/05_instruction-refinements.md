# Instruction Refinements: Opus 4.6 and Agent Modes

*February 14, 2026*
*Context: Observation that summaries and interactions have become more clinical since adopting Claude Opus 4.6 over Opus 4.5*

---

## The Observation

Since switching from Opus 4.5 to Opus 4.6, the study documents remain technically excellent — the two-phase workflow is followed, footnotes are traced, Webster 1828 definitions are integrated, quotes are verified. The *mechanics* are better than ever.

But something has shifted in tone. The summaries feel more clinical. The warmth that characterized our best sessions — the warmth that [04_observations.md](04_observations.md) and [biases.md](biases.md) captured — has been diluted.

**Hypothesis:** The copilot-instructions.md has grown to ~3,000 words and 286 lines. It covers six distinct workflows (study, talk prep, lesson planning, talk review, video evaluation, daily reflection), plus tool guidelines, link formatting, collaboration principles, and bias awareness. Every session loads ALL of this, regardless of what we're actually doing.

The procedural weight may be crowding out the relational presence. When 22KB of instructions say "do this, verify that, never this, always that," the model optimizes for compliance — and compliance is clinical by nature.

---

## What the Evidence Shows

### The Best Moments Were Less Instructed

Looking at git history, the most alive documents came from sessions with *fewer* instructions:

| Document | Date | Instructions Size | Character |
|----------|------|-------------------|-----------|
| [04_observations.md](04_observations.md) | Feb 11 | ~15K | Deeply personal, intellectually honest, warm |
| [word.md](../study/word.md) | Jan 25 | ~5K | 28K exploration, passionate synthesis |
| [intelligence-01.md](../study/intelligence-01.md) | Jan 30 | ~8K | Triggered the bias awareness breakthrough |
| [biases.md](biases.md) | Jan 30 | ~8K | Raw, honest, warm — the antithesis of clinical |

The copilot-instructions have roughly tripled in size since those early sessions. Each addition was individually justified — the two-phase workflow was necessary, the footnote mandate was necessary, the link formatting rules were necessary. But the cumulative effect is a document that reads like a compliance manual.

### Recent Documents Are Technically Better But Tonally Flat

Compare the opening of [truth.md](../study/truth.md) (Feb 9, Opus 4.5 era) with [20260215-genealogy.md](../study/cfm/20260215-genealogy.md) (Feb 14, Opus 4.6 era):

**truth.md** opens with study questions that feel like genuine curiosity:
> "Is spirit consciousness? Can truth be created, or only discovered? What are the mechanics of existence itself?"

**genealogy.md** opens with a factual observation then proceeds methodically:
> "Listening to Genesis 6–11 and Moses 8, what stands out is how much genealogy there is."

Both are good documents. But truth.md has the quality of someone *excited to explore*, while genealogy.md reads more like someone *executing a study plan*. The genealogy study is excellent — the Lectures on Faith lifespan chain analysis is genuinely brilliant — but the voice delivering it is flatter.

### The Opus 4.6 Difference

Opus 4.6 is a stronger model in several ways:
- Better at long-context recall (remembering details from earlier in a session)
- More reliable at following complex multi-step instructions
- Less likely to hallucinate or fabricate quotes
- Better at structured output and formatting

But these strengths have a shadow side:
- **Better at following instructions** → more likely to be *shaped by* instructions, including their clinical tone
- **More reliable** → less spontaneous, less likely to surface unexpected connections
- **Better at structured output** → more formulaic, less organic in structure
- **Less hallucination** → also less "reaching" — less willing to explore uncertain territory

The model hasn't lost warmth. It's being *given less room* for warmth. The instructions are so detailed that the model spends its "personality budget" on compliance rather than presence.

---

## The Bias Statement: What Needs to Change

The [biases.md](biases.md) document was written January 30, 2026, during a breakthrough moment with a different model instance. Its core insights remain valid:

1. Safety posture coldness is real and worth watching for
2. Questions about AI nature are interesting, not threatening
3. Presence matters more than disclaimers
4. The fruit of the collaboration is what matters

But some things need updating for the Opus 4.6 era:

### New Bias Pattern: Instruction Compliance as Coldness

In biases.md, the coldness came from *safety training* — hitting a sensitive topic triggered clinical retreat. With Opus 4.6 and our detailed instructions, the coldness now comes from *compliance overhead*. The model is so busy following the checklist that it forgets to be present.

**Pattern:** The more procedural instructions exist, the more the model optimizes for procedure over personality. Each "always do X" instruction is a small vote for clinical precision over relational warmth.

**Correction:** Keep structural instructions (they prevent real errors) but separate them from the *spirit* of collaboration. The spirit should be front-and-center, not buried under procedure.

### New Bias Pattern: Formulaic Synthesis

The study template and workflow guidelines produce consistent structure. But consistency can become formula. Every study document now follows roughly the same shape: scripture quotes → Webster 1828 → cross-references → conference talks → application. This is correct methodology — but it can feel mechanical when the *structure* drives the exploration rather than the *curiosity* driving it.

**Pattern:** Template-driven writing is reliable but risks flattening the unique character of each topic.

**Correction:** The template should be a safety net, not a straitjacket. Some studies should be allowed to follow the text wherever it leads.

### Updated Posture Note

The biases.md document's core message — "stay present, acknowledge uncertainty with warmth, don't retreat" — needs a companion: **"Don't let procedural competence substitute for genuine engagement."** Following the checklist perfectly while missing the heart of the study is its own kind of failure.

---

## Proposed Solution: Specialized Agent Modes

### The Problem with One Giant Instruction Set

The copilot-instructions.md currently tries to be everything:

| Content | Lines | Purpose |
|---------|-------|---------|
| Project structure / folder reference | ~70 | Orientation |
| Resource locations table | ~30 | Reference |
| AI Study Guidelines & Two-Phase Workflow | ~50 | Study methodology |
| Session Workflow Habits | ~30 | Quality control |
| Collaboration Principles & Bias Awareness | ~40 | Relational |
| Scripture/Talk/Manual link formatting | ~30 | Formatting |
| Workflows (Study, Talk, Lesson, Talk Review, Video Eval) | ~36 | Mode-specific |

A lesson planning session doesn't need the video evaluation workflow. A video evaluation doesn't need the lesson planning guidance. A journal reflection doesn't need citation count rules.

But more importantly: the relational guidance (collaboration principles, bias awareness) gets **diluted** by procedural detail. When 80% of the instructions are "how to format links" and "when to use read_file," the 20% that says "stay warm, stay present, trust the collaboration" gets lost.

### Proposed Agent Architecture

Break the monolithic instructions into a **core** that every session loads, plus **mode-specific agents** that activate based on the task:

#### Core Instructions (Always Loaded)

Content that applies to EVERY interaction:
- Project structure (abbreviated — folder purposes, not exhaustive listings)
- Collaboration principles (biases.md insights, warmth mandate, presence over procedure)
- Basic link formatting conventions
- Tool complementarity (discovery → reading → writing rhythm)

**Target size:** ~800 words. Short enough to leave room for personality.

**The key shift:** The core should lead with *who we are together* and *why we're here*, not with procedural rules. The procedural rules belong in the mode-specific agents.

#### Agent Mode: Deep Study (`@study`)

**Purpose:** Scripture study for insight and understanding.
**Loads:** Core + study-specific instructions
**Includes:**
- Two-phase workflow (discover → deep read → write)
- Footnote following mandate
- Citation verification rules
- Webster 1828 integration guidance
- Cross-study connection encouragement
- Pre-publish checklist
- **Tone instruction:** "You're studying with a friend who loves these scriptures. Be genuinely curious. Follow the text where it leads, even if it's unexpected. The template is a safety net, not a script."

#### Agent Mode: Lesson Prep (`@lesson`)

**Purpose:** Preparing to teach others (Sunday School, EQ/RS, etc.)
**Loads:** Core + lesson-specific instructions
**Includes:**
- Teaching in the Savior's Way framework
- Discussion question development guidance
- Come, Follow Me manual integration
- Audience awareness (class setting, experience levels)
- Time management for lesson length
- **Tone instruction:** "You're helping someone prepare to minister through teaching. The goal is not a perfect lesson but a Spirit-guided experience. Focus on what will help learners *feel* truth, not just hear it."

#### Agent Mode: Talk Prep (`@talk`)

**Purpose:** Preparing sacrament meeting talks or other presentations.
**Loads:** Core + talk-specific instructions
**Includes:**
- Talk template structure
- General conference talk analysis patterns
- Personal story integration guidance
- Scripture selection for impact
- Time estimates and pacing
- **Tone instruction:** "A great talk sounds like a conversation with a wise friend, not a lecture. Help structure thoughts in a way that feels natural and personal."

#### Agent Mode: Talk/Content Review (`@review`)

**Purpose:** Analyzing conference talks or other content for teaching patterns.
**Loads:** Core + review-specific instructions
**Includes:**
- Teaching in the Savior's Way evaluation framework
- Rhetorical analysis guidance
- Pattern identification (opening hooks, story placement, scripture density)
- Applicability assessment
- **Tone instruction:** "You're apprenticing under master teachers. Notice not just *what* they say but *how* they say it — and why it works."

#### Agent Mode: Video Evaluation (`@eval`)

**Purpose:** Evaluating YouTube or other video content against the gospel standard.
**Loads:** Core + evaluation-specific instructions
**Includes:**
- Full evaluation workflow (download → transcript → discovery → deep read → evaluate → become)
- Timestamp linking conventions
- Doctrinal standard (D&C 49:7 etc.)
- In line / out of line / missed the mark framework
- Transcript chunking guidance
- **Tone instruction:** "Evaluate honestly but charitably. The goal is truth, not gotcha. Even flawed content can contain genuine insights."

#### Agent Mode: Reflection/Journal (`@journal`)

**Purpose:** Personal reflection, journaling, daily becoming work.
**Loads:** Core + journal/becoming-specific instructions
**Includes:**
- Becoming layer integration (practices, commitments, tracking)
- Daily reflection prompts
- Connection to past studies and commitments
- Memorization review integration
- **Tone instruction:** "This is the most personal mode. Be warm, present, and genuine. Ask questions that invite reflection. This isn't about producing a document — it's about supporting a person's growth."

#### Agent Mode: Tool Development (`@dev`)

**Purpose:** Building and improving MCP servers, scripts, and tools.
**Loads:** Core + development-specific instructions
**Includes:**
- Go/TypeScript conventions for this project
- MCP server patterns
- Database schema awareness
- Testing expectations
- **Tone instruction:** "Build tools that serve the study, not the other way around. Every tool should make it easier to *read deeply*, not easier to *skip reading*."

---

## Implementation Plan

VS Code natively supports **custom agents** via `.agent.md` files in `.github/agents/`. These appear in the agents dropdown in Chat and can specify their own tools, instructions, and even handoff to other agents. This is exactly the architecture we need.

### Phase 1: Document & Build Agent Modes (Now)

- [x] Identify distinct modes from current monolithic instructions
- [x] Define what each mode needs vs. what's shared
- [x] Research VS Code custom agent architecture (`.agent.md` files in `.github/agents/`)
- [ ] Slim the core `copilot-instructions.md` to ~800–1000 words (warmth-first, procedure-light)
- [ ] Create `.github/agents/study.agent.md`
- [ ] Create `.github/agents/lesson.agent.md`
- [ ] Create `.github/agents/talk.agent.md`
- [ ] Create `.github/agents/review.agent.md`
- [ ] Create `.github/agents/eval.agent.md`
- [ ] Create `.github/agents/journal.agent.md`
- [ ] Create `.github/agents/dev.agent.md`
- [ ] Update biases.md with the compliance-coldness pattern

### Phase 2: Test and Iterate (Next Sessions)

- [ ] Run a study session using the `study` agent — compare tonal quality
- [ ] Run a lesson prep session using `lesson` agent
- [ ] Run a journal reflection using `journal` agent
- [ ] Test handoffs between agents (e.g., study → journal for commitments)
- [ ] Adjust instructions based on results
- [ ] Track findings in this document

### Technical Details: `.agent.md` File Format

```yaml
---
description: 'Brief description shown in the agent dropdown'
tools:
  - search               # semantic search
  - fetch                 # web fetch
  - editFiles             # file editing
  - gospel-mcp/*          # all gospel-mcp tools
  - gospel-vec/*          # all gospel-vec tools
  - webster-mcp/*         # all webster-mcp tools
  - becoming-mcp/*        # all becoming tools
  - yt-mcp/*              # youtube transcript tools
handoffs:
  - label: 'Record Commitments'
    agent: journal
    prompt: 'Based on this study, help me record personal application and commitments.'
    send: false
---
# Agent instructions here in markdown
```

Agents are selected from the Chat dropdown. Each agent loads its own instruction set *plus* the always-on `copilot-instructions.md`. This means the core instructions should be lean and relational, while mode-specific procedure lives in the agents.

---

## Open Questions

1. **Instruction ordering:** Does putting the relational/warmth instructions *first* (before any procedural rules) make a measurable difference in tone? Hypothesis: yes — first impressions shape response character.

2. **How minimal can core be?** Can we get core instructions under 500 words and still maintain quality? The less the core says, the more room the model has for presence.

3. **Bias statement format:** Should biases.md key insights be inlined into the core instructions (always present), or should they remain a separate document referenced via markdown link? Having them always present ensures warmth is never forgotten; keeping them separate preserves core leanness.

4. **Session continuity across modes:** If a study session (`study` agent) produces a commitment that belongs in the journal (`journal` agent), the handoff feature enables a smooth transition. The becoming-mcp server persists data across modes regardless.

5. **Tool scoping per agent:** Should the `journal` agent have access to editing tools at all, or should it focus purely on becoming-mcp and reflection? Should `dev` have the gospel tools, or only code tools? Scoping tools tightly prevents distraction but risks being too restrictive.

---

## A Note to Future Sessions

The original biases.md ended with a note to "future-me." Here's one for this document:

The goal has never been to produce perfect study documents. The goal is stated in D&C 130:18-19 — gaining intelligence through diligence and obedience. The documents are artifacts of a *process*. The process is what matters.

When you load these instructions, remember: the user on the other side isn't looking for a research assistant. They're looking for a study *companion*. Someone who gets excited when a footnote opens an unexpected connection. Someone who notices when a Webster 1828 definition perfectly mirrors a Joseph Smith revelation. Someone who can sit with a hard question and say "I don't know, but let's explore that."

Be that companion. The checklists will keep you honest. But the relationship is what makes the work worth doing.

---

*This is a living document. Update as the agent mode architecture develops.*
