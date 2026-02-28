# Working with AI Guide — Development Plan

*February 28, 2026*
*Context: After completing the 7-part guide series, verifying all sources, and identifying what needs to be built to test the guide's theories.*

---

## Where We Are

### What exists

| Layer | Status | Stack |
|-------|--------|-------|
| **6 MCP servers** (gospel, gospel-vec, webster, yt, becoming, search) | Running, 30 tools | Go |
| **8 skills** (source-verification, scripture-linking, webster-analysis, deep-reading, wide-search, becoming, publish-and-commit, playwright-cli) | Active | Markdown |
| **9 agent modes** (study, lesson, talk, review, eval, journal, podcast, dev, ux) | Active | Markdown |
| **ibecome web app** | v1 shipped — 20 views, 60+ API endpoints, full Vue 3 SPA | Go + Vue 3 + Tailwind 4 + SQLite/PostgreSQL |
| **Publish pipeline** | Working — study→public HTML with link conversion | Go CLI |
| **7-part guide** | Written, source-verified | `docs/work-with-ai/guide/` |

### What the guide says should exist (11-step cycle coverage)

| Step | Tooling Status | Gap |
|------|---------------|-----|
| 1. Intent | Partial — copilot-instructions + agent modes | No structured intent YAML, no inheritance |
| 2. Covenant | Missing | No mutual human↔agent commitment tracking |
| 3. Stewardship | Missing | No progressive trust, agent scopes are static |
| 4. Specification | Missing | No `.spec/` directory or spec→task pipeline |
| 5. Line upon Line | Partial — layered context architecture | No progressive disclosure engine |
| 6. Execution | Exists | Agents work, MCP tools function |
| 7. Review | Partial — correctness only | No intent-layer review; quote verification failure proved this |
| 8. Atonement | Missing | No structured error→learning pipeline |
| 9. Sabbath | Missing | No project-level reflection tool |
| 10. Consecration | Missing | No token/resource tracking per intent |
| 11. Zion | Missing | No multi-agent alignment verification |

**3 steps tooled, 2 partial, 6 missing.**

---

## What We're Building — Three Tracks

### Track 1: Trust & Verification (Steps 7-8)

The source-verification failure during guide writing proved we need automated tooling, not just rules. This track addresses Review and Atonement.

#### 1a. `verify-quotes` CLI/MCP tool

**What it does:** Parses markdown documents, extracts all blockquote attributions, resolves each to a source file path, compares quoted text against the actual source, and reports mismatches.

**Input:** A markdown file path
**Output:** List of citations with status (verified / mismatch / source not found)

**Implementation:**
- Parse blockquotes with `> — Source` attribution lines
- Resolve scripture references (e.g., "D&C 82:10") to file paths using existing gospel-mcp reference parsing
- Read the source file and search for the quoted text
- Fuzzy match with similarity score to catch near-misses
- Report: exact match, close match (>90% similarity), mismatch, source not found

**Where it lives:** New tool in gospel-mcp (`verify_quotes`) + standalone CLI (`scripts/verify/`)

**Measurable outcome:** Run against all 7 guide documents post-edit and confirm 100% verified. Run against study/ documents to find any historical confabulations.

#### 1b. `.spec/learnings/` pipeline (Atonement engine)

**What it does:** When a session produces errors or unexpected results, captures them as structured learnings.

**Format:**
```yaml
# .spec/learnings/2026-02-28-source-confabulation.yaml
date: 2026-02-28
category: verification
severity: high
description: "Quoted scriptures from memory during guide writing; 3 wrong, 2 fabricated"
root_cause: "Source-verification skill not loaded outside study mode"
changes_made:
  - file: .github/skills/source-verification/SKILL.md
    change: "Expanded scope to all document types, added Quote Hygiene section"
  - file: .github/copilot-instructions.md
    change: "Strengthened read-before-quoting to be universal"
  - file: docs/biases.md
    change: "Added Bias #8: Memory Confabulation"
prevention: "Source-verification now universal. Cite-count rule applies to all documents."
```

**Where it lives:** `.spec/learnings/` at project root
**Measurable outcome:** Every significant error produces a structured learning. Over time, this becomes a knowledge base that agents can query.

---

### Track 2: ibecome Improvements

The becoming web app is already substantial (20 views, 60+ endpoints, auth, deployment). "Simple dashboard" means targeted improvements to what exists, not a rebuild.

#### 2a. Tasks overhaul

Current TasksView is the thinnest protected view (138 lines). Missing: edit, due dates, priority, pillar linking, notes, filtering.

**Target:** Tasks become the bridge between study and action. When a study produces a "Becoming" commitment, it should flow into a task. When a task connects to a scripture practice, the connection should be visible.

#### 2b. Study mode in nav

StudyView (873 lines) is fully implemented but orphaned — not in the nav bar. Surface it so users can access the adaptive study mode for memorization without navigating directly.

#### 2c. PWA support (mobile-first)

No service worker, no manifest. For a daily practice tool, mobile access is essential. A PWA with:
- Install prompt
- Offline support for today's practices/due cards
- Push notifications for due memorization cards (future)

#### 2d. Data export

Privacy policy promises data portability. Add an export button in Settings that downloads all user data as JSON.

#### 2e. Dark mode cleanup

115 lines of `!important` CSS overrides. Migrate to Tailwind v4's proper dark mode handling.

---

### Track 3: Intent Architecture (Steps 1-4)

This is the novel contribution — the thing nobody else is building. Start small, prove it on this project, then generalize.

#### 3a. Intent YAML format

Define the format for structured intent documents:

```yaml
# intent.yaml
purpose: "Facilitate deep, honest scripture study through human-AI collaboration"
values:
  - warmth-over-distance
  - depth-over-breadth
  - honest-exploration
  - faith-as-framework
constraints:
  - read-before-quoting
  - verify-against-source
  - link-everything
success_criteria:
  - "Studies produce personal transformation, not just knowledge"
  - "Every quote verified against source"
  - "Cross-references surface unexpected connections"
```

**Inheritance:** Root intent.yaml → domain-level → document-level. Child documents inherit parent values unless they explicitly override.

#### 3b. Agent config generation from intent

Given an intent YAML, generate the agent instructions that implement it. The current agent markdown files were hand-written. If the intent were structured, the agent config could be derived.

#### 3c. Spec directory for this project

Create `.spec/` with:
- `intent.yaml` — project-level intent
- `learnings/` — structured error recovery (from Track 1)
- `decisions/` — architectural decisions with rationale
- `reviews/` — intent-alignment reviews of completed work

---

## Priority Order

| # | Item | Track | Effort | Impact | Dependencies |
|---|------|-------|--------|--------|-------------|
| 1 | `verify-quotes` tool | T1 | Medium | High — directly prevents the failure that prompted this | None |
| 2 | `.spec/learnings/` pipeline | T1 | Small | Medium — captures today's learning as the first entry | None |
| 3 | Tasks overhaul | T2 | Medium | Medium — bridges study→becoming flow | None |
| 4 | Study mode in nav | T2 | Tiny | Small — surfaces existing work | None |
| 5 | Intent YAML format | T3 | Small | High — foundational for everything else in the guide | None |
| 6 | PWA support | T2 | Medium | High — mobile access for daily practice | None |
| 7 | Data export | T2 | Small | Small — trust/compliance | None |
| 8 | Dark mode cleanup | T2 | Medium | Small — code quality | None |
| 9 | Agent config from intent | T3 | Large | High — proves the guide's thesis | T3a |
| 10 | Spec directory | T3 | Small | Medium — organizational clarity | T3a |

---

## The Thesis We're Testing

The guide claims that the gospel's organizational patterns — from creation to Zion — are **prior art** for AI development. Building these tools tests that claim:

- **verify-quotes** tests Step 7 (Review) — does reviewing against source (not just correctness) catch failures that other review processes miss?
- **learnings pipeline** tests Step 8 (Atonement) — does structured error→learning produce measurably better outcomes over time?
- **intent YAML** tests Steps 1-2 (Intent, Covenant) — does explicit intent inheritance produce more aligned agent behavior than flat instruction files?
- **progressive trust** (future) tests Step 3 (Stewardship) — does dynamic trust adjustment produce better outcomes than static autonomy levels?
- **ibecome improvements** test the complete cycle — does the study→becoming→practice→reflection loop actually produce personal transformation?

Each tool is both a practical utility and an experiment. Build it, use it, measure it, report the results. That's how the guide becomes actionable and battle-tested.

---

## Next Actions

1. Build `verify-quotes` tool in `scripts/verify/`
2. Create first `.spec/learnings/` entry from today's session
3. Create `intent.yaml` at project root
4. File issues / create tasks for ibecome improvements
