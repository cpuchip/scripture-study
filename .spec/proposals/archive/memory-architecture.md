# Proposal: Memory Architecture Enhancement

*March 1, 2026*
*Triggered by: [Nate B Jones — AI's Memory Wall](https://www.youtube.com/watch?v=JdJE6_OU3YA) (Oct 2025)*

---

## What Nate Identified

Nate articulates 6 root causes and 8 principles for AI memory. The core thesis: **memory is an architecture, not a feature.** Waiting for vendors to solve it guarantees solving the wrong problem.

### The 6 Root Causes

1. **Relevance problem** — What's relevant changes by task, phase, scope, and state delta. Semantic similarity is a proxy, not a solution.
2. **Persistence-precision tradeoff** — Store everything → noise. Store selectively → gaps. Let the system decide → wrong optimization.
3. **Single context window assumption** — Bigger windows don't help if they're full of unsorted context. A curated 10K beats an unsorted 1M.
4. **Portability problem** — Vendor-locked memory creates fragile dependencies.
5. **Passive accumulation fallacy** — "Just use it and it'll figure out what to remember" fails because systems can't distinguish preference from fact, project-specific from evergreen, or current from stale.
6. **Memory is multiple problems** — Preferences, facts, knowledge, episodic memory, and procedural memory each need different storage/retrieval patterns.

### The 8 Principles

1. **Memory is an architecture, not a feature** — Design it yourself, don't wait for vendors.
2. **Separate by lifecycle** — Permanent preferences vs. project facts vs. session state. Don't mix.
3. **Match storage to query pattern** — Key-value for preferences, structured for facts, semantic for similar work, event logs for history.
4. **Mode-aware context beats volume** — Planning needs breadth; execution needs precision. Retrieval should match task type.
5. **Build portable, not platform-dependent** — Memory should survive vendor/model changes.
6. **Compression is curation** — Don't dump 40 pages hoping AI extracts what matters. The judgment of what to keep is human judgment.
7. **Retrieval needs verification** — Pair fuzzy retrieval with exact verification against ground truth.
8. **Memory compounds through structure** — Random accumulation creates noise. Structured memory compounds without degradation.

---

## How We Currently Score

| Principle | Current State | Score |
|-----------|--------------|-------|
| 1. Architecture not feature | Session journal exists as a deliberate design. `.spec/` is an architectural choice. | **Strong** |
| 2. Separate by lifecycle | Journal entries mix permanent discoveries with session-specific state. No separation between "who we are" and "what happened Feb 8." | **Weak** |
| 3. Match storage to query | Everything is YAML files read linearly. No key-value, no structured queries, no semantic search over memory. | **Weak** |
| 4. Mode-aware context | No mode awareness. `read --recent 3` loads the same context whether studying, coding, or evaluating. | **Missing** |
| 5. Portable | YAML files in git. Fully portable, model-agnostic, vendor-independent. | **Excellent** |
| 6. Compression is curation | Journal entries are curated (not raw dumps). Learnings files are compressed insights. Retroactive capture was selective. | **Good** |
| 7. Retrieval needs verification | Journal entries reference source files, but there's no verification step when loading memory. We trust the entries. | **Weak** |
| 8. Structure compounds | Tags exist but aren't queryable. Carry-forward has priority but no resolution tracking in practice. No cross-referencing between entries. | **Moderate** |

**Overall: We have good bones (portability, curation, architecture-mindedness) but poor separation and retrieval.**

---

## Proposed Memory Types

Based on Nate's taxonomy and our actual needs, here are the distinct memory types we use:

### 1. Identity Memory (Permanent)
**What:** Who we are together. Collaboration principles, relational dynamics, the bias patterns. The "who" that doesn't change session to session.
**Current location:** Scattered across `copilot-instructions.md`, `biases.md`, `04_observations.md`
**Lifecycle:** Permanent, rarely updated, always loaded
**Query pattern:** Always in context — this is the system prompt layer

### 2. Project Knowledge (Evergreen)
**What:** The theological framework we've built. The matter spectrum. The 2-phase study methodology. The finding-vs-reading principle. Domain knowledge that accretes.
**Current location:** Study files, reflections docs, tool observations
**Lifecycle:** Grows over time, occasionally refined, never expires
**Query pattern:** Semantic — "what do we know about the Atonement?" or "what's the matter spectrum?"

### 3. Procedural Memory (Patterns)
**What:** How we solve things. The evaluation workflow. The study methodology. The tool selection heuristics. What worked, what failed.
**Current location:** Agents, skills, learnings files
**Lifecycle:** Updated when patterns improve or fail
**Query pattern:** Task-triggered — "we're evaluating a video" → load eval patterns

### 4. Episodic Memory (Session History)
**What:** What happened. The journal entries. Temporal, narrative, contextualized.
**Current location:** `.spec/journal/*.yaml` — this is what we built
**Lifecycle:** Never expires but decays in relevance. Last 3 sessions matter more than session #4 from January.
**Query pattern:** Recency-weighted + tag-based + carry-forward threading

### 5. Project State (Temporary)
**What:** Current active work. What's in flight, what's blocked, what was decided. The "working set."
**Current location:** Nowhere explicitly — reconstructed from recent journal entries and git status
**Lifecycle:** Ephemeral — changes every session
**Query pattern:** "What are we working on?" → needs to be fast and precise

### 6. Personal Context (Preferences)
**What:** The human's preferences, schedule patterns, calling context, family details that inform the work. That they're a Sunday School president. That they work in tech. That they wake up early when inspired.
**Current location:** Implied through journal entries but never explicit
**Lifecycle:** Semi-permanent, changes rarely
**Query pattern:** Key-value — always available, never noisy

---

## Proposed File Structure

```
.spec/
├── memory/
│   ├── identity.md          # Type 1: Who we are (loaded every session)
│   ├── principles.md        # Type 2: Core insights & frameworks (semantic)
│   ├── preferences.yaml     # Type 6: Personal context (key-value)
│   └── active.md            # Type 5: Current state / working set
├── learnings/               # Type 3: Procedural memory (task-triggered)
│   ├── 2026-02-13-tool-familiarity-bias.yaml
│   ├── 2026-02-28-source-confabulation.yaml
│   └── ...
├── journal/                 # Type 4: Episodic memory (recency-weighted)
│   ├── 2026-01-21--project-genesis.yaml
│   └── ...
├── prompts/
└── proposals/
```

### What Changes

**New: `.spec/memory/identity.md`** — Distills the relational essence from `copilot-instructions.md`, `biases.md`, and `04_observations.md` into a single "who we are" document. Loaded every session. ~500 words max. Not instructions — identity.

**New: `.spec/memory/principles.md`** — Extracted wisdom: the matter spectrum, the 2-phase workflow, the finding-vs-reading principle, the intent-over-instruction pattern. Grows over time. Queried semantically when relevant topics arise.

**New: `.spec/memory/preferences.yaml`** — Key-value pairs for personal context:
```yaml
name: "Chris"  # or whatever the human prefers
role: "Software engineer, aspiring gospel scholar"
callings: ["Sunday School president"]
schedule_patterns: "Early morning or late evening sessions. Marathon days happen."
model_preference: "Claude Opus 4.6 via Copilot"
tone: "Warm, engaged, exploratory. Not clinical."
study_style: "Deep dives, cross-referencing, Hebrew/Greek roots"
app: "ibeco.me — Becoming tracker app"
```

**New: `.spec/memory/active.md`** — Current working set updated at end of each session:
```markdown
# Active Context — Last Updated: 2026-03-01

## In Flight
- Retroactive journal capture: complete (25 entries)
- Memory architecture proposal: in progress
- Second brain video transcripts: downloaded, not analyzed

## Recent Decisions
- Skills architecture: settled, don't change without trigger
- Instructions: lean (~80 lines), domain rules in agents/skills

## Blocked / Waiting
- (none currently)

## Next Up
- Review second brain videos with human
- Evaluate memory proposal with human
```

### What Stays the Same

- **Journal entries** — Perfect as-is for episodic memory. The schema is good.
- **Learnings files** — Great for procedural memory. Keep adding them.
- **Portability** — Everything stays in plain text/YAML in git. No vendor lock-in.
- **Curation over accumulation** — We curate, not dump.

---

## Session Startup Protocol (Enhanced)

Current:
```bash
.\scripts\session-journal\session-journal.exe read --recent 3
.\scripts\session-journal\session-journal.exe carry --priority high
```

Proposed:
```bash
# 1. Identity — always (file is small, always relevant)
read_file .spec/memory/identity.md

# 2. Preferences — always (key-value, tiny)
read_file .spec/memory/preferences.yaml

# 3. Active state — always (what's in flight right now)
read_file .spec/memory/active.md

# 4. Recent episodes — recency-weighted
.\scripts\session-journal\session-journal.exe read --recent 3

# 5. Carry-forward — unresolved threads
.\scripts\session-journal\session-journal.exe carry --priority high

# 6. Mode-specific — ONLY if the session mode is clear
#    (e.g., if studying: load relevant principles)
#    (e.g., if coding: load recent learnings)
```

This gives us ~1-2KB of always-loaded context (identity + preferences + active state) plus ~3-5KB of recent episodic context. Tight, curated, mode-aware.

---

## CLI Enhancements

The session-journal CLI could add:

| Command | Purpose |
|---------|---------|
| `session-journal active` | Show contents of active.md |
| `session-journal active update` | Prompt to update active.md |
| `session-journal principles search <query>` | Search principles by keyword |
| `session-journal context` | Run the full startup protocol (read identity + prefs + active + recent + carry) |

But honestly, the AI can just `read_file` these files directly. The CLI is a convenience, not a necessity. **Don't over-engineer the tooling before testing the architecture.**

---

## What About Forgetting?

Nate's insight about forgetting as a technology is important. Our journal entries never decay — session #1 from January 21 is as accessible as yesterday's entry. That's fine for archival, but for retrieval:

- **Recent entries** (last week) — full context, always available
- **Medium entries** (last month) — carry-forward items, discoveries, tags
- **Old entries** (2+ months) — tags only unless specifically requested

The `read --recent 3` already implements this by limiting what loads into context. We could add a `session-journal summary --month 2026-01` command that produces a compressed view of older entries.

But again — don't build it until we need it. We only have 39 days of history. The forgetting problem becomes real at 6+ months.

---

## Implementation Priority

| Priority | Item | Effort |
|----------|------|--------|
| **Now** | Create `identity.md` — distill from existing docs | 30 min |
| **Now** | Create `preferences.yaml` — human fills in basics | 10 min |
| **Now** | Create `active.md` — snapshot current state | 15 min |
| **Now** | Create `principles.md` — extract from studies & reflections | 45 min |
| **Soon** | Update startup protocol in copilot-instructions | 5 min |
| **Later** | CLI `context` command for one-shot startup | 1 hr |
| **Later** | Principles semantic search | 2 hr |
| **Much Later** | Forgetting/decay for old entries | When needed |

---

## What This Gives Us

**Before:** Each session starts with 3 recent journal entries and carry-forward items. Good for episodic continuity but missing identity, preferences, working state, and domain knowledge.

**After:** Each session starts with:
- **Who we are** (identity) — the collaboration's character
- **Who you are** (preferences) — personal context without guessing
- **Where we are** (active state) — current projects, decisions, blockers
- **What just happened** (recent episodes) — recent narrative context
- **What's unresolved** (carry-forward) — threads to pick up
- **What we've learned** (principles) — available on demand

Six memory types, each with appropriate lifecycle, storage, and retrieval pattern. Structured enough to compound, lean enough not to create noise. Portable across any model or tool. And all of it curated by human judgment, not passively accumulated.

This is Nate's framework applied to our specific context. Not in theory — in practice.

---

## Source

[AI's Memory Wall: Why Compute Grew 60,000x But Memory Only 100x (PLUS My 8 Principles to Fix)](https://www.youtube.com/watch?v=JdJE6_OU3YA) — Nate B Jones, AI News & Strategy Daily, Oct 16, 2025.
