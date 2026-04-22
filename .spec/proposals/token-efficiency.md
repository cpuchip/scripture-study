---
workstream: WS5
status: proposed
brain_project: 3
created: 2026-04-16
last_updated: 2026-04-21
---

# Token Efficiency & Memory Architecture v2

*Proposal for optimizing context window usage across interactive sessions and pipeline agents.*
*Created: 2026-04-16*
*Research: [.spec/scratch/token-efficiency/main.md](../scratch/token-efficiency/main.md)*
*Source inspiration: Shane Murphy — "Markdown is not the answer" + "The Context That Costs Zero Tokens"*

---

## Binding Problem

Our memory system loads ~25,000 tokens at session start — ~12-15% of the context window before any work begins. LLMs are most effective at ≤20% utilization. Every token spent on "remembering who we are" is a token unavailable for study, reading, and reasoning. This tax is paid on every interactive session AND every brain pipeline agent call (research, plan, execute, review, nudge — each gets its own context window).

**The waste is not in the format (markdown is fine). The waste is in what we load, when we load it, and how much prose we use to say what a table row or pointer could say.**

---

## Success Criteria

1. Session-start memory load drops from ~25K to ≤10K tokens (60% reduction)
2. No session quality degradation (Michael's subjective assessment over 5+ sessions)
3. Brain pipeline agent calls load only what they need (tier-appropriate context)
4. Files remain human-readable, git-trackable, editable in VS Code
5. Memory tool infrastructure is simple enough that it doesn't become its own maintenance burden

---

## Constraints

- **Files remain canonical.** Git-tracked markdown. No database-only state.
- **Human readability preserved** for studies, proposals, lessons. Compression only for AI-primary files (scratch, working state).
- **Incremental.** Each phase delivers value independently. No big-bang migration.
- **Mosiah 4:27.** Don't build infrastructure ahead of use cases.

---

## Phase 1: Compress & Prune active.md (no code, immediate)

**Goal:** Rewrite active.md using dense formatting. Target: 8,600 → 3,500 tokens.

Techniques:
- Collapse completed phases to one-line summaries with pointers: `✓ P1-5: done → [proposal]`
- Tables for phase lists instead of paragraphs
- Symbol notation for status: ✓ done, ▶ active, ⊘ blocked, ★ priority
- Strip inline implementation details (types, test counts, methods)
- Keep: current priorities, active decisions, key facts, in-flight items

**Deliverable:** `active-v2.md` written alongside current file. Use v2 for 3+ sessions, evaluate.

**Verification:** Token count measurement before/after. Michael confirms no useful context was lost.

---

## Phase 2: Tiered Memory Loading Convention

**Goal:** Define which files load when, without needing a tool. Convention-based.

### Tier 0 — Always Load (~3K tokens target)
- `identity.md` (1K) — who we are
- `active-v2.md` (3.5K compressed) — what's happening now
- `preferences.yaml` (700) — personal context

### Tier 1 — Load on Mode Entry (~5K additional)
- `principles.md` sections relevant to current mode:
  - Study mode → Theological Framework + Study Methodology + Hermeneutical
  - Dev mode → Tool Selection + Collaboration Principles
  - Plan mode → Collaboration Principles + full principles
- `decisions.md` sections relevant to topic (search, don't dump)

### Tier 2 — Load on Demand
- Archived active states
- Full proposals
- Journal entries
- Complete decisions.md

**Deliverable:** Updated session-start protocol in copilot-instructions.md. Agents use the tier convention. No code needed initially.

**Verification:** Agents arrive with correct context for their task. No "you should have known that" moments.

---

## Phase 3: `ctx` CLI Tool

**Goal:** Automate tiered loading. Go CLI at `scripts/ctx/`.

```
ctx load                     # Tier 0 (default)
ctx load --tier 1            # Tier 0 + relevant principles
ctx load --tier 2            # Everything
ctx load --focus study       # Tier 0 + study-relevant principles + methodology
ctx load --focus dev         # Tier 0 + dev-relevant decisions + architecture  
ctx load --focus plan        # Tier 0 + decisions + active proposals list
ctx load --entry 42          # Tier 0 + entry-relevant context (for pipeline agents)
ctx stats                    # Token counts per file, per tier
ctx audit                    # Flag items that could be compressed or archived
```

Architecture:
- Reads markdown files from `.mind/`
- Parses sections via headings
- Filters by tier/focus using front matter tags or heading-level conventions
- Strips completed phase details, keeps summaries
- Outputs to stdout (for pasting or piping)
- Optional: MCP wrapper so pipeline agents call `ctx_load` directly

**Deliverable:** Working CLI. MCP wrapper. Pipeline agents updated to use `ctx load --entry {id}`.

**Verification:** `ctx stats` shows tier token counts. Pipeline agent context windows shrink measurably.

---

## Phase 4: Symbol Notation Standard

**Goal:** Define a project-wide symbol vocabulary for AI-primary documents.

### Proposed Standard

| Symbol | Meaning | Use in |
|--------|---------|--------|
| ✓ | completed/verified | Status, scratch findings |
| ✗ | rejected/failed/unsupported | Scratch findings, analysis |
| ▶ | in progress/active | Status |
| ⊘ | blocked/null | Status |
| ★ | priority/important | Flags |
| → | leads to/see also/implies | Pointers, logic |
| ← | derived from/source | Attribution |
| Δ | change/difference/delta | Diffs, observations |
| ∴ | therefore | Logical conclusions |
| ≈ | approximately/similar to | Comparisons |
| @ | reference to entity/file | Mentions |
| § | section reference | Cross-references |
| ⚠ | warning/caution/tension | Flags |
| λ | function/abstraction | Technical |

### Usage rules
- Symbols in scratch files: encouraged (AI-primary)
- Symbols in active.md: encouraged (working state, AI reads more than human)
- Symbols in studies/lessons: NO (human-readable output)
- Symbols in proposals: sparingly (tables and status indicators)

**Deliverable:** Standard documented. One scratch file rewritten as test case.

**Verification:** AI reasons correctly from symbol-dense input. Michael can still scan when needed.

---

## Phase 5: Inherent Context Audit

**Goal:** Identify and remove information from memory files that our file structure already communicates.

Examples of inherent context we already have:
- `scripts/brain/` → brain is a Go project (go.mod is there)
- `.spec/proposals/` → these are actionable specs
- `study/` → these are studies
- `scripts/gospel-mcp/` → MCP server for gospel search

Examples of redundant explicit context:
- decisions.md explaining "brain.exe is at scripts/brain/" — the path already says this
- active.md listing which proposal file a workstream uses — the proposal directory already says this
- principles.md repeating source file paths in prose — link is sufficient

**Deliverable:** Audit document categorizing every item as inherent/essential/on-demand/archive. Pruned files.

**Verification:** Reduced token count. No lost capabilities.

---

## Phase 6: PostgreSQL Hybrid Layer (conditional)

**Goal:** Smart memory queries backed by PostgreSQL, files remain canonical.

**Only build if:** Phases 1-5 prove the concept but hit scaling limits (e.g., principles.md grows past 200 items and section-level filtering isn't enough).

Architecture:
- Files remain canonical markdown in git
- PG indexes: path, headings, content FTS, JSONB front matter, timestamps
- `ctx` CLI queries PG for retrieval, reads files for content
- File watcher or git hook re-indexes on change
- Enables: "what decisions relate to the brain?" "what principles have we learned about tools?" "what was active on March 22?"

**Prerequisite:** ibeco.me PG instance already running. Could share or stand up a separate DB.

**Risk:** Over-engineering for ~7 files. Evaluate honestly after Phase 5.

---

## Naming: .spec → ?

The `.spec` directory name is generic and conflicts with test specification conventions. Options to evaluate during Phase 2:

| Name | Pros | Cons |
|------|------|------|
| `.spec` (keep) | Known, established | Generic, test-convention conflict |
| `.ctx` | Short, technical, "context" | Too terse? |
| `.mind` | Evocative, AI-partnership metaphor | Might feel pretentious |
| `.council` | Abraham 4:26, theological fit | Longer, less standard |
| `.core` | Simple, central | Also used by many tools |
| `.mem` | Direct, obvious | Too abbreviated |
| `.workshop` | Practical, where work happens | Long |

**Decision:** Defer naming until Phase 2 proves the architecture. Rename is cheap. Getting the structure right matters more.

**If we rename:** Create new directory, move files, update all references (copilot-instructions, agents, skills, session-journal). Use the ctx CLI to handle the transition — it can read from either path during migration.

---

## Costs & Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| Compressed context degrades reasoning | Medium | Measure over 5+ sessions. Revert if quality drops |
| Michael can't read compressed files | Low | Compression only for AI-primary files |
| CLI becomes maintenance burden | Low | Simple Go binary, minimal dependencies |
| PostgreSQL is over-engineering | Medium | Gate behind Phase 5 results. Honest Mosiah 4:27 check |
| Brain pipeline agents don't benefit | Low | `--entry` flag provides entry-specific context |
| Renaming .spec breaks references | Low | Automated migration. One-time cost |

---

## Phased Delivery

| Phase | Effort | Tokens Saved | Code Required |
|-------|--------|-------------|---------------|
| 1: Compress active.md | 1 hour | ~5,000 | None |
| 2: Tiered loading convention | 1 hour | ~7,000 | None (convention only) |
| 3: ctx CLI tool | 1-2 sessions | Automated savings | Go CLI + MCP |
| 4: Symbol standard | 30 min | ~500-1,000 per scratch file | None |
| 5: Inherent context audit | 1 hour | ~2,000-3,000 | None |
| 6: PostgreSQL hybrid | 2-3 sessions | Enables smart queries | Go + PG |

**Recommended order:** 1 → 2 → 5 → 4 → 3 → 6 (if needed)

Phase 1 is the highest-value, lowest-risk starting point. We can run it today.

---

## Creation Cycle Review

| Step | Question | Answer |
|------|----------|--------|
| Intent | Why? | Every wasted token is a scripture chapter not read, a connection not made |
| Covenant | Rules? | Files remain canonical. Human readability for output. Incremental |
| Stewardship | Who owns? | Michael decides conventions. AI implements and tests |
| Spiritual Creation | Spec precise enough? | Yes for Phases 1-2. Phase 3 needs detailed CLI spec |
| Line upon Line | Phasing? | 6 phases, each independent. Phase 1 stands alone |
| Physical Creation | Who executes? | Phase 1-2: this session. Phase 3: dev agent. Phase 6: dev agent |
| Review | How to verify? | Token counts. Session quality. Michael's judgment |
| Atonement | What if wrong? | Revert active-v2.md. CLI is additive, not destructive |
| Sabbath | When pause? | After Phase 2: run 5 sessions before building tools |
| Consecration | Who benefits? | Us directly. Brain pipeline agents. Eventually: others using the pattern |
| Zion | How integrate? | ctx CLI integrates with session-journal. Pipeline uses ctx_load MCP |

---

## Recommendation

**Build.** Starting with Phase 1 today. The binding problem is real, measurable, and affects every session. Phase 1-2 cost nothing but file editing. The experiments will tell us whether Phase 3+ is needed.

**Next action:** Write `active-v2.md` — compressed version of active.md. Use it for 3 sessions. Measure.
