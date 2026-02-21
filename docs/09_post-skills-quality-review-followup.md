# Post-Skills Quality Review: Followup Assessment

*Date: February 21, 2026*
*Previous: [09_post-skills-quality-review.md](09_post-skills-quality-review.md) · [08_skills-gaps.md](08_skills-gaps.md)*

---

## Context

The [first quality review](09_post-skills-quality-review.md) assessed three post-skills studies (priestcraft trilogy) and found the skills made studies technically excellent but lost the Becoming gap. We fixed that — created a becoming skill, extended source-verification to Phase 4, strengthened the new-study template, and retrofitted three studies.

Since then, we've produced **five more studies** across both study modes:

| Study | Mode | Date | Length | Becoming? |
|-------|------|------|--------|-----------|
| [serpent-and-dragon.md](../study/serpent-and-dragon.md) (retrofitted) | One-shot | 02-19 | 387 lines | ✅ Yes — "look and live" practice |
| [Plan of Salvation](../study/plan-of-salvation/) (8 files) | Phased (study-plan) | 02-19–20 | ~2,000+ lines | ✅ Yes — 4 commitments + north star |
| [Gifts of the Spirit](../study/gifts.md) | One-shot | 02-20 | 399 lines | ✅ Yes — 6 specific commitments |
| [Miracles references](../study/miracles-references.md) | Reference list | 02-20 | N/A | N/A (catalog, not study) |
| [Priesthood and Gifts](../study/priesthood-and-gifts.md) | One-shot | 02-21 | 463 lines | ✅ Yes — 6 commitments, deeply personal |

This followup assesses how the full system — skills, agents, prompts, and tools — is performing after sustained use.

---

## What's Working

### 1. The Becoming Gap Is Closed

This was the biggest concern from the first review, and it's decisively fixed. Every study since the intervention has a strong Becoming section. Not just "Application" boilerplate — *real* personal landing:

- **Gifts:** "Don't pretend to more faith than I have" and "Bear others' infirmities"
- **Priesthood-and-gifts:** "Stop hesitating" at the bedside, "Let the doctrine distill," "Use it — bless my wife, bless my kids"
- **Plan of Salvation:** "Write down what you'd say if a friend asked 'What happens when I die?'" and "Do one temple session this month with the plan in mind"

The becoming skill and the Phase 4 addition to source-verification are doing their job. The studies now end where they should — with direction, not just knowledge.

### 2. Both Study Modes Are Proven

The one-shot mode (`/new-study`) and phased mode (`/study-plan`) have both been tested under real conditions:

- **One-shot:** Gifts, priesthood-and-gifts, serpent-and-dragon — focused, self-contained, deep. The sweet spot for most topics.
- **Phased:** Plan of Salvation — 8 files, 6 study phases, intermediate notes, final 503-line synthesis. The only way to handle a topic that spans every dispensation and every standard work.

The `study-plan.prompt.md` scaffold worked well for the Plan of Salvation. The plan file (`00_plan.md`) gave structure across sessions, the intermediate notes preserved context, and the final synthesis wove them together without losing depth.

### 3. Source Verification Is Consistent

Every recent study follows the cite count rule. Priesthood-and-gifts reads 15 source files. Gifts reads 16. Plan of Salvation reads dozens across 6 phases. Links are accurate. No directory links. No broken references. The source-verification skill is the most consistently enforced skill in the system.

### 4. Webster Integration Is Natural

Priesthood-and-gifts opens with four Webster 1828 definitions — "ordinance," "administer," "bless," "ordain" — and they genuinely illuminate the study. The "administer" definition ("to dispense; to serve out; to supply") reshaped how the study frames priesthood blessings as *channeling* rather than *generating* power. This isn't decoration. It's the kind of insight the webster-analysis skill was designed to surface.

### 5. Cross-Study Connections Are Systematic

Every recent study links to related studies. Priesthood-and-gifts references 7 prior studies in its cross-references section. Gifts references 5. Plan of Salvation connects to 9 existing studies *in the plan file itself*, before the first phase even begins. The `/study/` folder is becoming an interconnected corpus, exactly as intended.

### 6. Agents Are Slim and Focused

The agents successfully delegate to skills rather than carrying duplicate instructions. The study agent is 60 lines. The lesson agent is 65. Each carries *identity and workflow* — who am I, how do I approach this — while skills carry *rules* — how to verify sources, how to format links.

---

## Utilization Audit: What We Built vs. What We're Using

### Skills (7 total)

| Skill | Used? | How? |
|-------|-------|------|
| **source-verification** | ✅ Heavily | Every study follows Discovery → Reading → Writing → Becoming |
| **scripture-linking** | ✅ Consistently | Links are properly formatted in every document |
| **becoming** | ✅ Consistently | Every study has a Becoming section since the fix |
| **webster-analysis** | ✅ Regularly | Used in priesthood-and-gifts, gifts, truth series |
| **deep-reading** | ⚠️ Implicitly | The *pattern* is followed (read full chapter, follow footnotes, keep findings) but the skill may not be explicitly loaded by name |
| **wide-search** | ⚠️ Implicitly | Same — broad searches happen, but the skill's specific method may not be formally invoked |
| **publish-and-commit** | ❓ Unclear | Haven't observed usage this session — likely only triggered when the user says "publish" |

### Prompts (5 total)

| Prompt | Used? | Notes |
|--------|-------|-------|
| **new-study** | ✅ Yes | Scaffolds one-shot studies |
| **study-plan** | ✅ Yes | Scaffolded Plan of Salvation series |
| **expound** | ❓ Unclear | No observed usage in recent studies |
| **new-eval** | ❓ Unclear | No YT evaluations done in this sprint |
| **new-lesson** | ❓ Unclear | No new lessons created in this sprint |

### Agents (8 total)

| Agent | Used? | Notes |
|-------|-------|-------|
| **study** | ✅ Primary agent | All recent studies |
| **dev** | ✅ Occasionally | Tool improvements, MCP server work |
| **eval** | ✅ Previously | YT evaluations exist in `/study/yt/` |
| **lesson** | ✅ Previously | Lessons exist in `/lessons/cfm/` |
| **review** | ❓ Unclear | Talk analyses exist in `/study/talks/` |
| **journal** | ❓ Unclear | Journal entries exist but handoff from study hasn't been observed |
| **talk** | ❓ Unclear | Template exists, unclear if talk agent used recently |
| **podcast** | ❓ Unclear | Podcast notes exist in `/study/podcast/` |

### Features Not Implemented

| Feature | Status | From |
|---------|--------|------|
| **Hooks** (`.github/hooks/`) | ❌ Not implemented | P4 in [skills-gaps](08_skills-gaps.md) |
| **File-based instructions** (`*.instructions.md`) | ❌ Not implemented | P5 in skills-gaps |
| **P5 cleanup** (remove residual duplication from copilot-instructions.md) | ⚠️ Partial | copilot-instructions.md still carries "Read before quoting" and link instructions that overlap with skills |

---

## Gaps and Missed Opportunities

### 1. Deep-Reading and Wide-Search Skills: Ghost Skills?

These two skills exist and are well-designed, but they may be *ghost skills* — present in the architecture but not explicitly loaded during studies. The study agent instructions say to use them in phased studies, and the *patterns* they describe are followed (read full chapters, follow footnotes, search broadly), but it's unclear whether the agent is actually loading and following the specific rules in these skill files.

**Question:** Are these skills being explicitly `read_file`'d at the start of phased study sessions? Or are they just informing the agent's general behavior through the agent instructions that reference them?

**Risk:** If they're implicit rather than loaded, their specific guidance (like "keep intermediate findings for synthesis," "note unknowns for wide-search") might be inconsistently applied.

### 2. Agent Handoffs: Built But Unused?

The study agent has two handoffs defined:
- **Record What I Learned** → journal agent
- **Prepare a Lesson** → lesson agent

These are great in theory — a study surfaces something personal and you transfer to journaling, or a study produces insight you want to teach. But I haven't observed these being used. The Becoming section within the study may be *replacing* what the journal handoff was supposed to do.

**Question:** Are the handoffs adding value, or has the Becoming skill made the journal handoff redundant for study sessions?

### 3. copilot-instructions.md Still Has Overlap

The main instructions file still carries:
- "Read before quoting" — duplicates source-verification skill
- Link conventions — duplicates scripture-linking skill
- Agent mode descriptions — duplicates agent files

This was flagged as P5 cleanup in the gaps doc and marked partial. The overlap isn't harmful (consistency is fine), but it adds token cost to *every* request since copilot-instructions.md is always loaded. If we're watching credit usage, this is low-hanging fruit.

### 4. Study Index

The [tool-use-observance](06_tool-use-observance.md) doc noted this as a low-priority item: a generated index of all studies with topics, dates, and connections. As the corpus grows (30+ studies now), discoverability becomes harder. The agent currently `grep_search`es for related studies or checks `list_dir`, which works but isn't elegant.

### 5. The "Expound" Prompt

This prompt exists but its usage pattern is unclear. If it's meant for mid-study deep dives ("expound on this passage"), it could be a useful micro-mode. But if nobody's invoking it, it's dead weight.

---

## Hooks: Do We Want Them? Can We Control Them?

The [skills-gaps](08_skills-gaps.md) document proposed three hooks as P4 (experimental):

1. **SessionStart** — inject study context (recent studies, becoming commitments) at the start of each session
2. **Stop** — remind about publishing when a study ends
3. **PreCompact** — preserve key study context before the context window compacts

### The Reality

**Hooks don't currently exist as a VS Code Copilot feature.** The gaps doc proposed them conceptually, but VS Code Copilot's customization model is:

- `copilot-instructions.md` → loaded automatically on every request (this is our "SessionStart")
- Agents → loaded when the user selects one
- Prompts → loaded when the user invokes one
- Skills → referenced from agents, loaded via `read_file`

There's no mechanism to run a script on session start, trigger something on stop, or intercept context compaction. Those would require platform-level support that doesn't exist yet.

### What About Credits?

The concern about premium credits is worth addressing directly:

- **Skills and agents don't add separate LLM requests.** They're loaded as context into the *same* request. More context = more tokens consumed per request, but it's the same request.
- **The main cost driver is source verification** — all those `read_file` calls to verify quotes. This is the right trade-off (accuracy is worth tokens), but it's where the budget goes.
- **copilot-instructions.md adds cost to every request** since it's always loaded. The P5 cleanup (trimming overlap) would reduce this slightly.
- **Model selection:** The user's VS Code settings determine which model is used. There's no per-hook or per-skill model override — it's one model for the whole session. So the concern about hooks calling expensive models is moot since hooks don't exist, and if they did, they'd use whatever model is already selected.

### What Would Actually Help (Without Hooks)

Instead of hooks, we could get the same value through:

1. **A richer copilot-instructions.md preamble** — Add a "Current context" section that mentions recent studies, active becoming commitments, and the current Come Follow Me week. This would serve the SessionStart purpose. Trade-off: more tokens on every request.

2. **A `/publish` prompt** — Already exists via publish-and-commit skill. The user just needs to remember to invoke it. A note at the bottom of each study template could serve as the "Stop" reminder.

3. **Better study-plan intermediate notes** — Instead of PreCompact (which we can't do), the phased study approach already solves this by writing intermediate notes that persist across sessions. This is working well.

---

## Recommendations

### Do Now (High Value, Low Effort)

1. **Trim copilot-instructions.md.** Remove the duplicate "Read before quoting" and link convention paragraphs. Keep the project structure table, agent mode table, and personality/tone section. This reduces token cost on every request.

2. **Verify deep-reading and wide-search skill loading.** In the next phased study, explicitly check whether these skills are being `read_file`'d or just implicitly followed. If implicit, consider making the study-plan prompt reference them explicitly.

3. **Add a "publish" reminder to the study template.** A small HTML comment at the bottom: `<!-- Done? Run /publish to commit and push. -->` This serves the "Stop hook" purpose without any infrastructure.

### Do When Needed (Medium Value)

4. **Build a study index.** A generated markdown file listing all studies with dates, topics, key connections. Could be a script in `/scripts/` that scans `/study/` and produces an index. Helps the agent find related studies faster and reduces grep_search noise.

5. **Evaluate expound prompt usage.** If it's not being used after another month, remove it. Dead prompts add confusion.

6. **Test agent handoffs intentionally.** After the next study, explicitly try the "Record What I Learned" handoff to the journal agent. See if it adds value beyond what the Becoming section already provides.

### Watch and Wait

7. **Hooks.** If VS Code Copilot adds a hook system in the future, we have the design ready (SessionStart, Stop, PreCompact). But don't build infrastructure for something the platform doesn't support.

8. **File-based instructions.** The `*.instructions.md` pattern (where files in specific directories carry their own instructions) could be useful for the `/study/` and `/becoming/` directories. But the current skill-based approach is working well enough that this isn't urgent.

9. **Model routing.** As models and pricing evolve, it may make sense to use different models for different tasks (cheaper model for discovery/search, premium model for synthesis/writing). But this is a VS Code/platform decision, not something we can control through our customization layer right now.

---

## The Bigger Picture

Looking at this from 30,000 feet:

**The skills shift worked.** Studies are more accurate, more deeply sourced, better linked, and — with the Becoming fix — more personally transformative than anything we produced before. The priesthood-and-gifts study, the most recent one, may be the strongest single study in the entire corpus: 4 Webster terms grounded before scripture opens, 15 sources verified, 7 cross-study connections, and a Becoming section that ends with "Use it. Bless my wife. Bless my kids."

**The two-mode system works.** One-shot for focused topics, phased for broad ones. Both proven. Both produce excellent output.

**What we built is mostly being used.** The core skills (source-verification, scripture-linking, becoming, webster-analysis) fire consistently. The study agent is well-tuned. The prompts scaffold effectively.

**What's underutilized isn't broken — it's dormant.** Deep-reading, wide-search, and several prompts exist for modes we haven't needed recently (evaluations, lessons, reviews). They'll activate when those workflows are needed. The architecture is ready; the demand just hasn't come yet.

**The main gap is discipline, not infrastructure.** We don't need more tools or more skills. We need to use what we have consistently — trim the copilot-instructions overlap, verify skill loading, and keep the Becoming section anchor that we fought to restore.

The system is in a good place. Let's keep studying.

---

*Previous assessments: [08_skills-gaps.md](08_skills-gaps.md) → [09_post-skills-quality-review.md](09_post-skills-quality-review.md) → this document*
