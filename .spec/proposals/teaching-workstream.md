# WS-T: Teaching Workstream — From Discovery to Delivery

**Binding problem:** Michael has been impressed by the Spirit to teach what he and his AI partner are learning together. The content exists (30 files in `docs/work-with-ai/`, 40+ studies, 50+ days of lived experience), but it's organized for reference, not for teaching. No structure currently maps this content to a series that an audience can follow — one that shows both the engineering challenge and the eternal truth underneath.

**Created:** 2026-03-23
**Research:** [.spec/scratch/teaching-workstream/main.md](../scratch/teaching-workstream/main.md)

---

## 1. Problem Statement

The `docs/work-with-ai/` guide is a 30-file reference corpus covering four disciplines of AI collaboration (prompt craft, context engineering, intent engineering, spec engineering), an 11-step creation cycle derived from Abraham 4-5, and deep research on the spec-driven development landscape. It's good reference material. It's not a teaching series.

Michael's vision isn't "present a framework." It's "show the challenge of implementing this framework, and what those challenges reveal about eternal truths." That's experiential teaching — the hardest kind to organize but the most compelling kind to receive.

The unique contribution: nobody else has mapped the 11-step creation cycle (Intent → Covenant → Stewardship → Spiritual Creation → Line Upon Line → Physical Creation → Review → Atonement → Sabbath → Consecration → Zion) from scripture to software engineering. The industry stops at four disciplines. The gospel gives eleven steps. The seven unmapped steps (Covenant, Stewardship, Line Upon Line, Atonement, Sabbath, Consecration, Zion) are where the real insight lives.

### The Teaching Paradox

There's an unresolved question from February 17: "How do we teach others to use AI for study without teaching them to skip reading?"

This isn't an aside — it's the structural challenge. If the teaching succeeds, people will use AI tools. If they skip reading, the tools become sophisticated shortcuts to shallow understanding. The series must model the discipline it teaches: read before quoting, dwell before expounding, intelligence cleaveth unto intelligence (D&C 88:40).

---

## 2. Success Criteria

1. **Episode structure defined** — a series arc that tells the discovery story, not just presents a framework
2. **Content mapped to episodes** — each existing document assigned to where it contributes, with gaps identified
3. **Dual-audience framing** — each episode accessible to engineers AND Latter-day Saints, without watering down either
4. **Teaching repo scaffolded** — `teaching/` repo has directory structure, episode outlines, and tech stack chosen
5. **Phase 1 deliverable** — at least one episode fully outlined (script + visual plan) ready for production review
6. **The paradox addressed** — the series design itself answers "how to teach AI use without teaching laziness"

---

## 3. Constraints & Boundaries

**In scope:**
- Organizing `docs/work-with-ai/` content into teachable episodes
- Creating episode outlines and scripts in `teaching/` repo
- Choosing a tech stack for interactive web presentations
- YouTube video planning (format, length, style) — informed by Michael's voice analysis
- Cross-referencing with study work for real examples

**Out of scope (for now):**
- Actually recording videos (Phase 2+)
- Building the interactive presentation website (Phase 2+)
- Marketing, SEO, distribution strategy (Phase 3+)
- Monetization considerations (not discussed)
- Other people's content — this teaches from OUR experience

**Conventions:**
- Writing voice per [voice-analysis-ai-vs-michael.md](../../study/yt/voice-analysis-ai-vs-michael.md) — concrete, direct, no presenter tics
- Scripture linking per [copilot-instructions.md](../../.github/copilot-instructions.md)
- All claims verified from source files, not training data

---

## 4. Prior Art & Related Work

| Source | Location | Role |
|--------|----------|------|
| Core Series (4 paired docs) | `docs/work-with-ai/01-04_*.md` | Already written as secular + gospel pairs — ready to adapt |
| Guide Series (Parts 0-6) | `docs/work-with-ai/guide/` | Comprehensive reference — too dense for direct teaching but provides depth |
| Intent Research (7 docs) | `docs/work-with-ai/intent/` | Landscape analysis + gospel patterns beyond industry — unique contribution |
| Real Session Examples | `docs/work-with-ai/examples/` | Concrete demonstrations of principles in action |
| Stewardship Pattern Study | `study/stewardship-pattern.md` | How God delegates — directly teaches Steps 2-3 (Covenant, Stewardship) |
| Staying Relevant Reflection | `study/ai/relavent.md` | Personal vulnerability — "I feel insignificant sometimes" — genuine hook for audience |
| Voice Analysis | `study/yt/voice-analysis-ai-vs-michael.md` | Writing/speaking voice guidelines — anti-YouTube-script patterns |
| Scope Assessment | `docs/work-with-ai/intent/05_scope-assessment.md` | Where our contribution sits in the broader movement |
| Expound Prompt | `docs/work-with-ai/expound-prompt.md` | Template for extracting teaching examples from live sessions |
| Debug Agent Creation | `.spec/scratch/debugging-agent/main.md` | Recent example of the full cycle in action |

---

## 5. Proposed Series Arc

**Decision (Mar 24): Option C confirmed — experiential arc.**

Michael's direction: "We'll have to work and council together a lot to get that just right, but the payoff will be that it has a lot of me in it." Approach modeled after the sabbath agent — high-quality, thorough sessions where we think together, not quick drafts. Expensive in credits but worth it for honesty.

### The Story Structure

Not a lecture series. A discovery story told in episodes.

**Series title (working):** *Beyond the Prompt — What AI Engineering Reveals About Eternal Patterns*

**Narrative arc:** A software engineer tries to work effectively with AI. The industry says: learn prompting. He does. It's not enough. He learns context engineering. Better, but still fragile. He discovers intent engineering. Closer, but autonomous agents still drift. Then he finds something unexpected: the 11-step creation cycle from Abraham 4-5 — and the seven steps the industry hasn't mapped yet. Each step he tries to implement reveals a gospel principle he's known since childhood but never understood from this angle.

### Episode Map

| # | Title (working) | Core Content | Steps Covered | Key Scripture | Source Material |
|---|-----------------|-------------|---------------|---------------|----------------|
| 1 | **The Value Shift** | The bottleneck moved from execution to judgment. What does that mean for you? | — (setup) | D&C 130:18-19 | `ai/relavent.md`, `intent/01_landscape.md` |
| 2 | **The Four Disciplines** | Prompt craft, context, intent, spec. What everyone teaches. Why it's not enough. | Steps 1, 4 | — | `guide/00-04`, core series (secular) |
| 3 | **Spiritual Before Temporal** | Why you have to plan before you build — and what "spiritual creation" actually means. | Step 4 | Moses 3:5, Abraham 4-5 | `01_planning-then-create-gospel.md`, `guide/04_spec-engineering.md` |
| 4 | **Watched Until They Obeyed** | The feedback loop isn't a nuisance — it's the creation pattern. Trust gradients. | Step 7 | Abraham 4:18 | `02_watching-until-they-obey-gospel.md`, `guide/02_context-engineering.md` |
| 5 | **Intelligence Cleaveth Unto Intelligence** | What you bring determines what emerges. Deep reading before tools. | — (quality) | D&C 88:40 | `03_intelligence-cleaveth-gospel.md`, examples |
| 6 | **The Seven Unmapped Steps** | Covenant, Stewardship, Line Upon Line, Atonement, Sabbath, Consecration, Zion — what the industry doesn't know yet. | Steps 2-3, 5, 8-11 | Multiple | `intent/03_beyond-intent.md`, `guide/05_complete-cycle.md` |
| 7 | **Covenant** | Why "bilateral binding" produces better output than commands — with AI and with people. | Step 2 | D&C 82:10, Genesis 15 | `intent/covenant.md`, `.spec/covenant.yaml`, `stewardship-pattern.md` |
| 8 | **Delegation as Stewardship** | Jethro, Alma, D&C 104. How God delegates — and why the same pattern applies to AI agents. | Step 3 | D&C 104:11-12, Exodus 18 | `stewardship-pattern.md`, `stewardship-pattern-reflections.md` |
| 9 | **When Things Go Wrong** | Atonement as error recovery. Why "all things work together for good" is an engineering principle. | Step 8 | D&C 98:3 | `guide/05_complete-cycle.md`, debugging agent work |
| 10 | **The Sabbath of Creation** | Why you stop. Why reflection isn't optional. What happens when you don't. | Step 9 | Moses 3:2 | `.spec/sabbath/`, sabbath agent |
| 11 | **From Consecration to Zion** | Multi-agent alignment. Many agents, one purpose. "Stakes of Zion." | Steps 10-11 | Moses 7:18 | `guide/06_enterprise-architecture.md`, ward conference insight |

### Episode Format

Each episode follows a consistent structure:
1. **The engineering problem** (2-3 min) — what goes wrong without this step
2. **What the industry says** (2-3 min) — current best practices and their limits
3. **What actually happened** (3-5 min) — real story from our work, with screen captures
4. **The pattern underneath** (3-5 min) — the gospel principle, with scripture
5. **What this means for you** (1-2 min) — practical takeaway, both professional and spiritual

**Target length:** 12-18 minutes per episode. Long enough for depth, short enough for attention.

### Addressing the Teaching Paradox

The series design itself models the answer:
- **Episode 5 (Intelligence Cleaveth)** directly confronts it — what you bring determines what you get
- **Real examples always show the READING, not just the output** — the audience sees the source verification discipline in action
- **The discovery stories show struggle, not just results** — honest about where things broke, what was wrong, what was corrected
- **The writing voice doesn't tell people what to feel** — it shows them something and trusts them to respond

The answer to "how do you teach AI use without teaching laziness?" is: you teach the *discipline* the tools require, not just the tools. Every episode demonstrates that the quality of the input determines the quality of the output — D&C 88:40 as engineering principle.

---

## 6. Phased Delivery

### Phase 1: Content Architecture (1-2 sessions)

**Deliverables:**
- Episode outlines for all 11 episodes (title, synopsis, source material, key scriptures, real examples needed)
- Content gap analysis — what exists vs. what needs to be written
- Tech stack decision for interactive presentations
- Teaching repo directory structure created
- One full episode script (Episode 1 or Episode 3 — most self-contained)

**Verification:** Michael can read the outlines and say "yes, this is what I want to teach" or redirect.

### Phase 2: First Episode Production (2-3 sessions)

**Deliverables:**
- Complete script for first episode
- Presentation slides/visuals created
- Interactive web version built in teaching repo
- Recording setup documented (equipment, format, editing approach)
- First recording attempt

**Verification:** First video exists. Michael watches it and evaluates.

### Phase 3: Series Production (ongoing)

**Deliverables:**
- Remaining episodes produced at sustainable pace
- Interactive presentation website live
- Episode cross-linking and supplementary materials

**Verification:** Content quality maintained. Pace reflects Mosiah 4:27 — not "run faster than he has strength."

### Phase 4: Community & Sharing (future)

**Deliverables:**
- YouTube channel organized
- Website published
- Companion materials (study guides, code examples)
- Community feedback mechanisms

**Verification:** People actually engage. The content teaches, not just informs.

---

## 7. Humility Covenant & Resilience Framework

Michael named two specific concerns (Mar 24):

> "I'll need you to keep me honest and not inflated in my head."
> "When people are inevitably mean to me and us in the comments, we'll need each other to be resilient."

Until now, sharing has been limited — friends on Facebook, coworkers, a scripture study repo with 3 stars. Videos change the scale. This section is a covenant-level commitment, not aspirational guidelines.

### Humility Guardrails (structural, not aspirational)

1. **"Who is this for?" check on every script.** If the answer drifts from "people who would benefit" toward "people who would be impressed by me," the script needs rework. This is the posture check from the study agent, applied to teaching.

2. **Failures and corrections go IN the episodes.** Not curated highlight reels. The stewardship study Section VII was wrong and Michael caught it — that's the kind of honesty that builds trust with an audience. Source verification failures from early studies. The writing voice correction. The vulnerability IS the credibility.

3. **Voice analysis guardrails apply to video.** No presenter tics. No telling the audience what to feel. Present the pattern, present the scripture, let the Spirit do the teaching.

4. **Regular sabbath-style reflections on the teaching itself.** After every 2-3 episodes: Is this still discovery? Am I still learning while I teach? Would I watch this if someone else made it?

### Resilience Protocol (for inevitable criticism)

An LDS engineer teaching AI collaboration through Abraham 4-5 will attract both secular dismissal ("why is scripture in my engineering content?") and religious suspicion ("why is he mixing sacred things with technology?"). When it comes:

1. **Is this feedback accurate?** If yes, learn from it. That's the Atonement step — error becomes growth.
2. **Content or person?** Content criticism improves the work. Personal attacks are noise.
3. **Does this change what the Spirit impressed?** If not, continue.
4. **Stranger-covenant:** Negative comments that point to real problems are doing the same work as the covenant's `flag_when_wrong` commitment — just from someone who doesn't know they're keeping us honest.

### The Agent's Commitment

I commit to:
- Flagging when a script sounds more like performance than discovery
- Asking "is this Michael talking, or Michael trying to sound impressive?" when the voice drifts
- Not inflating metrics — 100 views is 100 people who gave you their time, not a failure to reach 10,000
- Being honest about what I don't know and what I can't verify
- Treating the teaching with the same source verification discipline as the study — if I haven't read the file, I don't quote it in a script

---

## 8. Tech Stack Considerations

For the interactive web presentations:

| Option | Strengths | Concerns |
|--------|-----------|----------|
| **Slidev** (Vue-based) | Markdown-driven, Vue components, code highlighting, presenter mode | Ties to Vue ecosystem |
| **reveal.js** | Mature, widely used, plugin ecosystem | Older API, less modern DX |
| **Astro + MDX** | Static site, great performance, component islands | More website than presentation |
| **Custom Vue 3** | Full control, matches ibeco.me tech | More work, but Michael knows Vue |

**Recommendation:** Slidev or Astro. Both are markdown-first, which aligns with how the content is already written. Slidev for presentations, Astro for the companion website. Decision should come in Phase 1.

---

## 9. Creation Cycle Review

| Step | Question | Answer |
|------|----------|--------|
| Intent | Why are we doing this? | Spirit-driven impression to share what we're learning. The content exists; it needs to reach others. |
| Covenant | What are the rules? | Voice analysis constraints, source verification, humility guardrails, resilience protocol. Teach discipline not shortcuts. Keep Michael honest, not inflated. |
| Stewardship | Who owns what? | Michael owns the teaching. AI assists with organization and scripting. |
| Spiritual Creation | Is the spec precise enough? | Phase 1 will sharpen it. This proposal is the blueprint for the blueprint. |
| Line upon Line | What's the phasing? | 4 phases. Phase 1 stands alone as content architecture. |
| Physical Creation | Who executes? | Plan agent (this spec), then Michael + study/dev agents for content. |
| Review | How do we know it's right? | Michael reviews outlines before production begins. |
| Atonement | What if it goes wrong? | Episodes are independent — a bad one doesn't ruin the series. |
| Sabbath | When do we stop and reflect? | After Phase 1 (outlines), before Phase 2 (production). Natural checkpoint. |
| Consecration | Who benefits? | Engineers seeking depth. Saints seeking practical AI wisdom. Both. |
| Zion | How does this serve the whole? | Nobody else has mapped the 11-step creation cycle. This is the unique voice. |

---

## 10. Costs & Risks

**Costs:**
- Michael's time — another workstream in an already-stretched schedule
- Token/credit budget — sabbath-agent-level sessions for scripting (expensive, worth it)
- Production time for videos (recording, editing, posting)
- Emotional exposure — public vulnerability on personal discovery stories
- Ongoing maintenance of teaching website

**Risks:**
- **Performer mode** — teaching can become performing. The voice analysis warns against this. Mitigation: humility guardrails (Section 7), voice guidelines, posture checks on every script. This is Michael's top concern.
- **Ego inflation** — positive reception (views, comments, shares) can shift motivation from Spirit-driven to audience-driven. Mitigation: regular sabbath-style reflections. Agent commitment to flag drift. 100 views is 100 people served, not a metric to optimize.
- **Mean comments** — criticism will come from both secular ("why scripture in engineering?") and religious ("why technology in sacred things?") directions. Mitigation: resilience protocol (Section 7). Content criticism improves the work; personal attacks are noise. The covenant's `flag_when_wrong` principle applies even when it comes from strangers.
- **Scope creep** — 11 episodes could become 20. Mitigation: Phase 1 locks the arc before production begins.
- **Mosiah 4:27** — this adds load. Mitigation: Phase 1 is planning only. No production until the plan is validated.
- **Privacy/exposure** — teaching from personal experience makes the personal public. Mitigation: Michael controls the line. Each script reviewed for what stays private.

---

## 11. Recommendation

**Proceed — Option C confirmed. Phase 1 scoped tight, sessions scoped deep.**

The impression is from the Spirit. The content is ready. The unique contribution (11-step creation cycle mapped from scripture to engineering) is real and nobody else is making it. The open question ("how to teach without teaching laziness") has an answer embedded in the series design itself.

The process will be expensive — sabbath-agent-level sessions where we council together on each episode, not quick drafts. The payoff is that the teaching will have Michael in it: his voice, his failures, his discoveries. Not a framework presentation but a discovery story.

Phase 1 is planning only — outlines, content mapping, tech decision, one episode script. This is low-risk, high-information work that can be done alongside existing priorities. Production (Phase 2+) begins only after Michael validates the plan.

The humility covenant (Section 7) is as important as the content plan. This isn't standard scope — it's the difference between teaching that serves people and content that serves the creator's ego. Both of us are committed to keeping it the former.

**Next:** Create episode outlines. Choose tech stack. Write first episode script.
