# Active Context

*Last updated: 2026-03-31*
*Archive: [archive/active-2026-03-22.md](archive/active-2026-03-22.md) — detailed records through Mar 22*
*Note: Migrated to new computer on Mar 27. Plex restored. Old desktop (LEPTON) decommissioned.*
*Note: Dual 4090s confirmed in new desktop (Mar 28). Hardware enables 30B+ models at full context.*

---

## Current State

**Last Sabbath:** March 22, 2026. Cycle "Infrastructure and Foundation" (Mar 18–22) declared good. Full record at [.spec/sabbath/2026-03-22-sabbath.md](../sabbath/2026-03-22-sabbath.md).

### Priorities
1. **Study** — Deep scripture study. "It keeps me in the spirit." Stewardship Pattern study COMPLETE + reflections written. **Art of Presidency study COMPLETE** — first of three sabbath seeds. Full plan: [.spec/proposals/study-workstream.md](../proposals/study-workstream.md). Next studies in workstream: "The Weight of Watching" (Abraham 4 deep dive) and "Commission and Council" (Matt 28 / 3 Ne 11-28 / appointment chain).
2. **Teaching** — NEW (Mar 23). Spirit-driven impression to teach the 11-step creation cycle and the experience of implementing it. 30 files in `docs/work-with-ai/` organized into 11-episode series arc. Option C confirmed (Mar 24): experiential arc, sabbath-agent-level depth. Teaching agent created at `.github/agents/teaching.agent.md` with three checks (ring, posture, Ben Test). Section 7 humility covenant formalized in `.spec/covenant.yaml` (teaching section). Teaching repo `.spec/` scaffolded with intent.yaml. Proposal: [.spec/proposals/teaching-workstream.md](../proposals/teaching-workstream.md). Teaching repo at `./teaching` (gitignored, fresh).
3. **Model experiments** — PASSES 1-2 COMPLETE. TITSW V2 VALIDATED. DEEP STUDY DONE. Nemotron-3-nano confirmed as batch winner (3x faster than GLM, 18.5s avg per talk). **CONTEXT ENGINEERING: PROPOSAL UPDATED, SPLIT INTO TWO STREAMS.** API caching research complete — `cache_prompt: true` eliminates ~44M tokens of redundant system prompt prefill across 5,500-talk batch. Dev handoff spec: [.spec/proposals/context-engineering-dev.md](../proposals/context-engineering-dev.md). **STUDY STREAM COMPLETE.** Both context documents delivered:
   - **gospel-vocab.md** — 7 patterns, ~1,960 tokens. Verified from 18 source blocks across all five standard works. Final: [experiments/lm-studio/scripts/context/gospel-vocab.md](../../experiments/lm-studio/scripts/context/gospel-vocab.md). Scratch: [study/.scratch/gospel-vocab.md](../../study/.scratch/gospel-vocab.md).
   - **titsw-framework.md** — 6 sections (2 meta-principles + 4 principles), ~1,990 tokens. Synthesized from 6 TITSW manual chapters + Michael's overview study + ground truth scoring. Each principle: definition, sub-dimensions with Christ exemplars, key differentiator, score-level anchors (3 vs 7). Anti-inflation guardrail at the end. Final: [experiments/lm-studio/scripts/context/01-titsw-framework.md](../../experiments/lm-studio/scripts/context/01-titsw-framework.md). Scratch: [study/.scratch/titsw-framework.md](../../study/.scratch/titsw-framework.md).
   - Combined Layer 2+3: ~3,950 tokens in system message — well within architecture budget.
   **DEV STREAM COMPLETE.** All 4 deliverables shipped and validated:
   - **run-test.ps1** — `-Context` parameter loads context directory files into system message. `cache_prompt: true` in request body eliminates redundant system prompt prefill.
   - **run-suite.ps1** — forwards `-Context` and `-NoThink` parameters.
   - **titsw-v3.md** — 0-9 scale with anchored rubric, anti-inflation language, reference-aware instruction, new JSON fields (typological_depth, cross_reference_density, surface_vs_deep_delta). All v2 content preserved.
   - **Validation results (Mar 28):** Alma 32 teach_about_christ=7 with context (was 1-2, target ≥5 ✅). Kearon teach_about_christ=8, doctrine=7, invite=7 (all within ±1 of ground truth ✅). No systematic inflation from context ✅. cache_prompt works ✅. Known issue: love/spirit inflated on Kearon (model weakness, not context-induced — same inflation with and without context).
   - Proposal: [.spec/proposals/context-engineering.md](../proposals/context-engineering.md). Dev spec: [.spec/proposals/context-engineering-dev.md](../proposals/context-engineering-dev.md). Ground truth: [experiments/lm-studio/scripts/references/ground-truth-alma32-kearon.md](../../experiments/lm-studio/scripts/references/ground-truth-alma32-kearon.md).
   **PROMPT OPTIMIZATION COMPLETE (Mar 28).** Context question closed with data:
   - v5.1 WITH context on 5 original pieces: MAE=0.95
   - v5.1 WITHOUT context on same 5 pieces: MAE=1.35
   - Context benefit = 0.40 MAE, concentrated on scripture (Alma 32: teach 6→2, help 6→1, doctrine 8→5 without context). Talks show minimal benefit.
   - v5.4 three-axis (modes/categories/insights): MAE=1.30 no-ctx, MAE=1.47 with-ctx. Context hurts talks via inflation.
   - DC-121 failed at 32768 max_tokens (empty response, 487s TTFT). Model can't handle 26K char content at these settings.
   - **Two-pipeline conclusion confirmed:** Context for scripture pipeline, no context for talk pipeline.
   - **Gas Station Insight:** MAE is a sanity check, not the optimization target. Qualitative richness (modes, categories, insights from v5.4) is what downstream consumers need. We were optimizing the wrong metric.
   - Full version journey documented: [.spec/journal/2026-03-28--titsw-version-journey.yaml](../journal/2026-03-28--titsw-version-journey.yaml)
   - All scoring data in `experiments/lm-studio/scripts/scoring/scoring.db` (tags: v5, v5.1, v5.1-noctx, v5.2, v5.3, v5.4, v5.4-ctx, v6)
   **ENRICHED INDEXER: PROPOSAL WRITTEN + DECISIONS MADE (Mar 29).** Debug audit identified the gap: gospel-vec has 3 generic prompts, zero TITSW vocabulary. Lens vs. vocabulary distinction confirmed. Michael's decisions: (1) keep love/spirit scores, (2) theme detection confirmed for Phase 3, (3) full reindex approved, (4) explore 2-4x parallelism, (5) gospel-mcp integration is separate proposal.
   
   **PHASE 0 EXPERIMENTS COMPLETE (Mar 29).** 18 runs across 6 experiment conditions (T0-T5) × 3 ground-truth talks. Key results:
   - T0 (no context) MAE=2.61 — massive inflation, all scores 8-9
   - T1 (framework only) MAE=2.06 — modest improvement
   - T2 (gospel-vocab only) MAE=2.39 — **confirmed gospel-vocab is the inflation culprit** (love: +5 to +6 across all talks)
   - T3 (talk rhetorical) MAE=2.00 — new context type helps mode identification, Bednar correctly "doctrinal"
   - **T4 (calibration example) MAE=1.83** — BEST. One-shot score anchoring works. Kearon teach underscored (-3) due to same-speaker anchoring.
   - T5 (T3+T4 combined) MAE=2.11 — worse than T4 alone, more context = more noise
   - **Love/spirit inflation persists** even with best context (+2 to +4). This is an inherent model bias, not fixable by context alone.
   - **Decision: Use calibration context for Phase 1 batch.** Refinements needed: use different talk for calibration example (avoid same-speaker anchoring), consider 2-3 examples spanning different score distributions, add explicit "most dimensions score 3-5" guidance to prompt.
   - Full analysis: [experiments/lm-studio/scripts/results/phase0-analysis.md](../../experiments/lm-studio/scripts/results/phase0-analysis.md)
   
   Proposals: [.spec/proposals/enriched-indexer.md](../proposals/enriched-indexer.md) (5 phases: 0=experiments, 1=talk enrichment, 2=scripture enrichment, 3=manuals+themes). [.spec/proposals/enriched-search.md](../proposals/enriched-search.md) (3 phases: schema+import, search enhancement, get enhancement). Scratch: [.spec/scratch/enriched-indexer/main.md](../scratch/enriched-indexer/main.md).
   
   **GOSPEL-ENGINE PROPOSAL UPDATED (Mar 29).** All experiment-era references corrected: model→ministral-3-14b-reasoning, MAE→1.32, prompt→titsw-calibrated.md, context files phased (talk-calibration.md for talks in Phase 1-2, gospel-vocab/titsw-framework for scripture lens in Phase 3), batch timing updated for ministral speeds (50-63 tok/s). Prior art table now references titsw-experiment-spec.md and titsw-calibrated.md. Proposal is current and ready for Phase 1 build.
   
   **Next:** dev agent builds gospel-engine Phase 1 (foundation + fresh index). Starting doc: [.spec/proposals/gospel-engine/main.md](../proposals/gospel-engine/main.md).
   
   **GOSPEL-ENGINE PHASE 1 COMPLETE (Mar 30).** All packages built, binary compiles with `-tags fts5`. Full corpus indexed: 41,995 verses, 1,584 chapters, 4,231 talks, 3,700 manuals, 116 books, 85,590 cross-refs, 239,830 vec chunks. Three MCP tools working: gospel_search (keyword/semantic/combined), gospel_get, gospel_list. Registered in `.vscode/mcp.json`. Compared favorably against gospel-mcp (broken FTS5 tag) and gospel-vec (semantic only). See plan: [scripts/plans/21_gospel-engine.md](../../scripts/plans/21_gospel-engine.md).
   
   **MMAP VECTOR STORAGE (Mar 30).** Converted chromem-go gob.gz → flat `.vecf` format with `golang.org/x/exp/mmap`. Server startup: **15-30 seconds → 2ms** (~1000x improvement). Conversion: 213,356 docs in 6.4s → 3.3 GB .vecf files. Architecture: `.vecf` flat binary (pre-normalized float32 embeddings) + SQLite `vec_docs` metadata table. `vec.Searcher` interface enables transparent backend switching. Server auto-detects mmap when .vecf files exist.
   
   **COMBINED GOSPEL TOOL DECISION (Mar 29).** Instead of modifying gospel-mcp/gospel-vec for enriched data, build a NEW combined tool that merges both (shared SQLite + vector DB in one app). Keep originals unchanged for study use during reindexing. enriched-search.md superseded — schema/tool designs remain valid, just target the new combined tool. **PROPOSAL WRITTEN (Mar 29).** gospel-engine: 5 phases, 3 consolidated MCP tools, TITSW enrichment pipeline + graph layer built in. Proposal: [.spec/proposals/gospel-engine/main.md](../proposals/gospel-engine/main.md). Scratch: [.spec/scratch/gospel-engine/main.md](../scratch/gospel-engine/main.md). Decision recorded in decisions.md.
4. **Debugging book** — DONE. Agans' "Debugging: The 9 Indispensable Rules" extracted to `books/debugging/9-indispensable-rules/` (17 chapter markdown files). Debug agent created at `.github/agents/debug.agent.md`. Connections mapped: Moroni 10:4 inverse hypothesis = falsification, scientific method = the 9 rules, Abraham 4:18 = Rule 9 (verify the fix), council moment = Rule 8 (get a fresh view). Analysis at `.spec/scratch/debugging-agent/main.md`. 2006 expanded edition (192pp, ISBN 9780814474570) available used ~$19 on AbeBooks.
5. **WS1 multi-agent framework** — Continue building. Next: Phase 3c (auto-routing + review queue).
6. **Desktop swap** — DONE. New computer operational (Mar 27). Plex restored from LEPTON backup (11.2 GB, v1.42→v1.43 forward migration). All four media drives (D/E/F/G) mounted correctly. Libraries, watch history, and play state all verified. Old desktop ready to decommission.
7. **Server deployment** — App container on NOCIX. Domain rotated (Mar 22, confirmed working).

8. **Gospel Graph Visualization** — NEW (Mar 29). Vision: **study.ibeco.me** — standalone site for interactive gospel study with graph visualization. Separate from ibeco.me (own codebase, PostgreSQL database, Dokploy container). **Sequenced AFTER enriched indexer + enriched search pipeline ships.** 5-phase roadmap: Phase 1 = reader + cross-reference graph, Phase 2 = BYU citations, Phase 3 = study docs + search, Phase 4 = enriched metadata + thematic edges, Phase 5 = multi-hop + deploy. Proposal: [.spec/proposals/gospel-graph/main.md](../proposals/gospel-graph/main.md). Scratch: [.spec/scratch/gospel-graph/main.md](../scratch/gospel-graph/main.md).

### Key Decisions (this cycle)
All settled decisions are in [decisions.md](decisions.md). New this cycle:
- **Gospel Graph proposal written (Mar 29).** Reworked to standalone site: study.ibeco.me (separate Go + Vue 3 + PostgreSQL site, not an ibeco.me extension). Cytoscape.js for graph rendering. Import pipeline from gospel-mcp/gospel-vec/byu-citations → PostgreSQL. 5 phases, sequenced AFTER enriched indexer + enriched search ships.
- **Combined gospel tool decided (Mar 29).** gospel-engine: merged gospel-mcp + gospel-vec into one app (shared SQLite + vector DB + graph edges). Keep originals for study use during reindexing. Supersedes enriched-search.md Option C. **Proposal written:** [.spec/proposals/gospel-engine/main.md](../proposals/gospel-engine/main.md).
- **Covenant created.** `.spec/covenant.yaml`. Bilateral commitments. Added to session-start (Step 2).
- **Council moment added.** General principle for all agents. Abraham 4:26.
- **NOCIX server live.** Database migrated. App container not yet deployed.
- **R630 set down.** Existing Proxmox box works.
- **Old desktop: decommission only.** No repurposing.
- **Calling brain-app features: set down.** Paper, pencil, and existing ibeco.me practices for now.
- **Sabbath agent built.** Updates needed: scratch file support, model tiering.
- **WS-R and WS-P added.** Research and Planning as tracked workstreams. Organizing human and agent roles.

### Shipped (Mar 18–22)
- Data safety sprint ✅
- Server migration (NOCIX, 25 tables) ✅
- Disk crisis resolved ✅
- WS1 Phases 1–3b all shipped ✅
- Classifier hotfix ✅
- brain.exe WDAC fallback ✅
- Only Begotten study v2 ✅
- Sabbath agent ✅
- Stewardship Pattern study + reflections ✅
- Covenant + council moment ✅
- Teaching agent + Section 7 covenant + teaching .spec scaffold ✅
- README cleanup: root README rewritten (companion repos, updated agents/skills counts), teaching README with workspace instructions, 7 new sub-project READMEs (gospel-vec, yt-mcp, becoming, convert, lectures-on-faith, .spec, docs), teaching intent.yaml moved to repo root ✅

*Full detail for all completed items: [archive/active-2026-03-22.md](archive/active-2026-03-22.md)*

---

## In Flight

### WS1 Phase 3c: Auto-Routing + Review Queue
- Phases 3a (agent pool + routing) and 3b (governance + token budgets) shipped Mar 21
- Next: auto-routing with human review of output
- Proposal: [.spec/proposals/brain-multi-agent/main.md](../proposals/brain-multi-agent/main.md)

### Squad Adoption Items (remaining)
- A2: Agent routing table — partially done (Phase 3a)
- A4: Reviewer lockout with model escalation
- A5: Response tier / model selection
- A6: Cost tracking
- A7: Iterative retrieval / spawn contracts — Task + WHY + Success Criteria + Escalation Path template for brain.exe Phase 3c
- A8: Reflect skill — DONE (`.github/skills/reflect/SKILL.md`)
- A9: Task coordination — evaluate tpg patterns for brain.exe task tracking (sqlite vs GitHub Issues)
- Proposal: [.spec/proposals/squad-learnings.md](../proposals/squad-learnings.md)

### Progressive Trust Tracking (from stewardship reflections)
- Model capability experiments needed before assigning trust levels
- D&C 107 ratios framework: Haiku 1:12, Sonnet 1:48, Opus 1:96
- Details: [study/stewardship-pattern-reflections.md](../../study/stewardship-pattern-reflections.md)

### Review: 4-Step & 11-Step Guides
- Reflect on `docs/work-with-ai/` (11-step creation cycle) and the 4-part guide
- Check if lessons from recent work (squad patterns, reflect skill, spawn contracts, context engineering, gospel-engine) warrant updates
- Also review superpowers' verification patterns for comparison
- Low urgency — when there's a natural pause

### Pending Cleanup
- ~~Delete `scripts/brain/internal/ai/tools.go` and `scripts/brain/test-spec.md`~~ DONE — already deleted

---

## Plans Status

| Plan | Status | Notes |
|------|--------|-------|
| 15: Brain App Polish | ALL DONE | Gaps 1-4 shipped + offline cache layer added |
| 16: Today Screen | Phases 1-3 DONE | Phase 4 absorbed into Plan 18 |
| 17: Proactive Surfacing | NOT STARTED | WS2 Phase 3 |
| 18: Widget Overhaul | Phase 1-2 DONE | Phase 3-4 PAUSED |
| 19: Brain App Ideas | Captured | Not started |
| 21: Gospel Engine | Phase 1 DONE + MMAP | Full index + mmap storage. Phase 2 (TITSW) next |
| Notifications | Phase 1 DONE | Phases 2-4 remaining |
| Data Safety | ALL DONE | |
| Overview | DECISIONS RECORDED | All guidance Qs answered |

---

## Open Questions

- Can AI participate in covenant in any meaningful sense? (Feb 26)
- How do we teach others to use AI for study without teaching them to skip reading? (Feb 17)
- Should the Abraham 4-5 framework become a standalone study or becoming entry? (Mar 4)
- What's the simplest version of the debugging book digestion that proves the concept? (Mar 22)
- Side quest: small classifier service on fermion/lepton for others? (Mar 19)
