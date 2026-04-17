# Token Efficiency & Memory Architecture v2 — Research

*Scratch file for .spec/proposals/token-efficiency.md*
*Created: 2026-04-16*

---

## Binding Problem

Our memory/context system consumes ~25K tokens at session start — ~12-15% of context window before any actual work begins. Shane Murphy's research confirms LLMs are most effective at ≤20% context utilization, with steep decline past 60%. Every token spent on memory is a token unavailable for study, reading, and reasoning. This affects interactive sessions AND the brain pipeline (where agents pay per-call).

---

## Measurements (2026-04-16)

### Memory files loaded every session

| File | Chars | ~Tokens | Notes |
|------|-------|---------|-------|
| active.md | 30,087 | ~8,600 | Largest. Contains full phase histories of completed work |
| principles.md | 25,149 | ~7,200 | Theological + methodology insights. Rarely changes |
| decisions.md | 12,465 | ~3,600 | Architecture decisions. Semi-stable |
| identity.md | 3,419 | ~1,000 | Who we are. Rarely changes |
| preferences.yaml | 2,475 | ~700 | YAML already. Compact |
| copilot-instructions.md | 12,817 | ~3,700 | Always injected by VS Code |
| **Total (before work starts)** | **~86K** | **~25K** | |

### Agent files (loaded per-mode)

| Agent | Chars | ~Tokens |
|-------|-------|---------|
| story.agent.md | 18,832 | ~5,400 |
| dev.agent.md | 13,801 | ~3,900 |
| debug.agent.md | 13,216 | ~3,800 |
| plan.agent.md | 12,230 | ~3,500 |
| lesson.agent.md | 11,331 | ~3,200 |
| teaching.agent.md | 10,942 | ~3,100 |
| sabbath.agent.md | 10,656 | ~3,000 |
| study.agent.md | 10,314 | ~2,900 |

### Biggest offenders in active.md

- Orchestrator Steward phases 1-6: ~2,000 tokens of per-phase detail (all completed)
- Brain UX QoL phases 1-7: ~1,800 tokens of per-phase detail (all completed)
- Pipeline Evolution phases 1-9: ~2,500 tokens of per-phase detail (all completed)
- Commission UI: ~500 tokens of phase detail (all completed)
- Each completed phase includes: types added, test counts, methods, what shipped, date

**Key insight:** Completed phases are *history*, not *state*. They belong in the proposal or an archive — not in the always-loaded active context.

---

## Source: Shane Murphy Videos (2026-04-16/17)

### Video 1: "Markdown is not the answer for AI Agents"
https://www.youtube.com/watch?v=y1NBC1iXiL4

Key claims:
- Markdown wins accuracy benchmarks over JSON, XML, YAML. XML distant second
- Token *reduction* doesn't improve accuracy — just makes room
- JSON is verbose (repeated keys/schema). TUNE project reduces JSON tokens but doesn't beat MD accuracy
- YAML more compact than JSON but repeated characters still waste tokens
- LLMs most effective at ~20% context utilization, steep decline past 60%
- Bigger context windows (1M, 2M) don't help — same 20% rule applies
- The `@` symbol as reference/mention mechanism is underutilized
- Real problem: "intent, purpose, and productivity" — not just format optimization

### Video 2: "The Context That Costs Zero Tokens"
https://www.youtube.com/watch?v=qM4CtPhNQa8

"Inherent context" — things the model understands from pre-training without explicit injection:
1. **Monorepo structure** — seeing `package.json` or `go.mod` tells the model the project type for free
2. **File names as meaning** — `_archive`, `_temp`, `v3` in filenames encode status without instructions
3. **Path as meaning** — where a file sits creates semantic relationships (ad-copy in /videos/ vs /project/)
4. **Path as policy** — hooks/scripts that enforce rules with minimal context injection (read/write locks, line limits)
5. **Virtual trees** — file-system-based data stores the agent navigates with bash (vs. needing a tool)
6. **agents.md as changelog** — temporal context (what changed, when) is more valuable than instruction repetition

Rules for inherent context:
- Can't require the agent to learn something new (not a tool)
- Must be git-trackable
- Must be minimally invasive to context window
- Relies on pre-training knowledge, not prompt injection

### What we already do well (inherent context)
- `.spec/memory/` — path communicates "this is memory" without explanation
- `.spec/proposals/` vs `.spec/scratch/` — path communicates document type
- `study/.scratch/` — working drafts
- `scripts/{name}/` — code for named tools
- YAML for preferences — structured, compact, AI-native

---

## Tokenizer Facts

- Claude uses BPE tokenizer trained heavily on English
- English is the most efficiently tokenized language (most BPE merges)
- Japanese: semantically dense (kanji = concepts) but tokenizer-inefficient (fewer merges, 1-3 tokens per character)
- Made-up language: catastrophically bad (zero BPE merges, every character boundary costs)
- Base64: ~33% MORE tokens than plaintext (demonstrates robustness, not efficiency)
- Mathematical/scientific symbols are well-trained: Δ, →, ∴, ≈, @, ✓, ✗ — all 1 token with rich pre-trained meaning

### Symbols that carry pre-trained meaning (1 token each)

| Symbol | Meaning | Training domain |
|--------|---------|-----------------|
| Δ | change, delta, difference | Physics, math, programming |
| → | implies, leads to, maps to | Logic, programming, math |
| ∴ | therefore | Logic, math proofs |
| ≈ | approximately | Math, physics |
| @ | reference, mention | Programming, social media |
| ✓ / ✗ | pass / fail | Testing, checklists |
| ★ | priority, important | Ratings, documentation |
| ⊘ | blocked, null, empty | Math, programming |
| ▶ | in progress, active | UI, media |
| § | section reference | Legal, academic |
| ⚠ | warning, caution | Software, documentation |
| λ | function, abstraction | Programming, math |

---

## Michael's Ideas (session conversation)

1. **CLI/MCP tool for tiered memory loading** — reads our file structure, strips unneeded parts, flag for tier level. Files remain durable markdown, tool delivers compressed context.
2. **Rename `.spec` to something more meaningful** — or create `.spec_v2/` alongside during transition.
3. **PostgreSQL as backing store** — robust JSON, FTS, graph capabilities. Already used in ibeco.me production.
4. **Symbol notation** — Δ, @, →, Greek letters for common concepts.
5. **Lookup table** — pre-defined symbol→meaning mappings.

### On PostgreSQL
Pros:
- Already running in ibeco.me production (Dokploy VPS)
- JSONB: structured queries on semi-structured data
- FTS: full-text search on memory content
- Graph-like queries: ltree extension, recursive CTEs for relationship traversal
- Temporal queries: "what was active on date X?" is trivial with timestamps
- Shared between brain pipeline agents and interactive sessions
- Concurrent access (multiple agents, multiple sessions)

Cons:
- Memory files stop being git-trackable (Shane's rule: must be git-trackable)
- Breaks the "files are durable" principle that governs our scratch/proposal workflow
- Another service dependency (though already running)
- Editing memory requires a tool/UI instead of just opening a file

Hybrid option: PostgreSQL as the *query layer* over git-tracked files. Files remain canonical. DB indexes them for fast tiered retrieval. CLI/MCP reads from either source. Best of both worlds — files for durability/editability, DB for smart loading.

### On renaming .spec

Ideas discussed or considered:
- `.mind` — the AI's working memory
- `.cortex` — neural metaphor
- `.ctx` — context (short, technical)
- `.core` — the core of the collaboration
- `.covenant` — theological, fits our framework
- `.council` — Abraham 4:26, "took counsel among themselves"
- `.mem` — memory (very direct)

Constraints: should be short, meaningful to both human and AI, not conflict with existing conventions (`.spec` conflicts with testing conventions in some ecosystems).

---

## Analysis: Where Are the Real Savings?

### Not in markdown syntax
Headers, bullets, bold — these cost very few tokens. `## Heading\n` is 3-4 tokens. The format overhead of markdown is <5% of file size.

### In prose duplication
active.md repeats what proposals already say. Every completed phase is summarized inline at full fidelity when it could be a one-line pointer.

**Example — current (Orchestrator Steward Phase 2, ~180 tokens):**
> **Phase 2 (DONE):** Model escalation — EscalationChain (Haiku→Sonnet→Opus→Human), `pickModel()` decides escalation based on diagnosis + failure count, `ModelOverride` threaded through Pipeline (AdvanceRequest, ExecuteRequest, runResearch, runPlan, runExecute), `config.ModelCost()` for cost lookup, `MaxCostPerEntry` cost guardrail (10.0 default), `quarantineCostLimit()` for budget-exceeded entries, escalation tracking in Status/Action, 33 tests all passing. Shipped Apr 11.

**Compressed version (~20 tokens):**
> ✓ P2: Model escalation (Haiku→Sonnet→Opus→Human) → [proposal]

Same information retrieval: if I need the details, I read the proposal. The inline summary is wasted tokens 95% of sessions.

### In temporal retention
Completed items persist at full fidelity forever. After a Sabbath, completed phases should collapse. Archives exist but aren't used aggressively enough.

### In narrative where tables suffice
Phase lists are paragraphs. Should be table rows.

---

## Experiment Designs

### Experiment 1: Compressed Active State
**Hypothesis:** active.md can be rewritten with symbols + tables + pointers at 40-50% token reduction without degrading session quality.
**Method:** Create `active-v2.md` alongside current. Use compressed version for 3 sessions. Evaluate: Did I miss context? Did I need to read proposals more? Was reasoning quality affected?
**Measure:** Token count (automated), session quality (Michael's judgment), number of "go read the proposal" roundtrips.

### Experiment 2: Reference-Not-Repeat
**Hypothesis:** Inline phase summaries in active.md can be replaced with pointers (`→ proposal#section`) without information loss.
**Method:** Strip all completed phase descriptions. Replace with `✓ Phase: one-liner → [file]`. Track whether any session needed the detail.
**Measure:** Frequency of needing to read proposals mid-session.

### Experiment 3: Tiered Memory CLI/MCP
**Hypothesis:** A tool that reads .spec files and serves tiered context will reduce session-start token load by 50-70% while keeping files durable and human-editable.
**Design:**
```
ctx load --tier 0     # Identity + priorities + active decisions (~2K tokens)
ctx load --tier 1     # + principles + recent decisions (~8K tokens) 
ctx load --tier 2     # + full history, archives (~25K tokens)
ctx load --focus dev  # Tier 0 + dev-relevant principles + brain architecture
ctx load --focus study # Tier 0 + study methodology + theological framework
```
Files remain markdown on disk. Tool reads, strips, compresses, serves.
Could be: Go CLI (like session-journal), MCP tool, or both.
**Implementation options:**
- Go CLI — `scripts/ctx/` — invoked at session start like session-journal
- MCP server — always available, agents can call `ctx_load` with tier/focus
- Both — CLI for manual sessions, MCP for pipeline agents

### Experiment 4: Symbol-Dense Scratch Files
**Hypothesis:** Working/scratch files using symbol notation are readable by AI at equal or better quality with 40-60% fewer tokens.
**Method:** Rewrite one scratch file in compressed notation:
```
## Findings
- @D&C-93:29 → intelligence=eternal ∴ not created
- Δ from KJV: "light" → φῶς (phōs) ≈ illumination, not brightness  
- ★ tension: v33 vs v36 re: agency timeline
- ✗ claim "spirits existed before" — not supported by text
```
Compare: token count, reasoning quality when used in a study session.
**Risk:** Michael can't read compressed scratch files as easily. Dual-audience problem.
**Mitigation:** Only for AI-primary files (session scratch, pipeline working docs). Human-facing studies stay prose.

### Experiment 5: Inherent Context Audit
**Hypothesis:** Some information currently in memory files is already communicated by our file structure (inherent context) and can be removed.
**Method:** Audit each section of active.md, decisions.md, principles.md. Mark each item:
- `inherent` — path/structure already communicates this
- `essential` — must be loaded, no other source
- `on-demand` — valuable but can be loaded when needed
- `archive` — historical, move to archive
**Measure:** Items categorized, tokens freed.

### Experiment 6: PostgreSQL Hybrid Memory
**Hypothesis:** A PostgreSQL layer indexing git-tracked files enables smart queries ("what decisions relate to the brain?") without losing file durability.
**Method:** 
- Files remain canonical (markdown in .spec/)
- PG indexes: file path, front matter, headings, content FTS, timestamps
- CLI/MCP queries PG for retrieval, reads files for content
- Changes sync: file watcher or git hook re-indexes on change
**Prerequisite:** Experiments 1-3 first. This is infrastructure for infrastructure. Only build if simpler approaches hit limits.
**Risk:** Mosiah 4:27. This is exciting but potentially over-engineered for current scale (~7 memory files).

---

## Critical Analysis Notes

### Is this the RIGHT thing to build?
Yes — but with scope discipline. The token budget problem is real and measurable. It affects every session and every pipeline agent call. The danger is over-engineering: building a PostgreSQL-backed tiered loading system for 7 files.

### Simplest useful version?
Experiment 1 (compressed active.md) + Experiment 2 (reference-not-repeat). Zero new code. Just rewrite the file. If that saves 40% of active.md tokens, that's ~3,400 tokens freed — roughly a full chapter of scripture or an extra proposal read.

### What gets worse?
- Compressed formats are harder for Michael to scan visually
- Pointer-only summaries require more tool calls to get full context
- If the CLI/MCP adds another tool to manage, that's complexity

### Does this duplicate something?
The session-journal CLI already does session-aware context loading. The ctx tool would be a generalization of that pattern. Could potentially merge.

### Mosiah 4:27 check
Michael has brain pipeline, teaching workstream, commission UX, space center, study, and tool improvements all in flight. However: experiments 1-2 cost nothing but file editing. The CLI (exp 3) is a small Go tool. The dangerous one is PostgreSQL (exp 6) — that's a real project.

### Ordering
1. Experiment 5 (audit) — understand what we have before changing format
2. Experiment 1 + 2 (compress + reference) — zero cost, immediate savings
3. Experiment 4 (symbol notation) — test on one file, evaluate
4. Experiment 3 (CLI/MCP) — build if manual compression proves the concept
5. Experiment 6 (PostgreSQL) — only if we hit limits that files + CLI can't handle

---

**Fix plan extracted to proposal:** [.spec/proposals/token-efficiency.md](../proposals/token-efficiency.md)
