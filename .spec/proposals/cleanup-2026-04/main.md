# Cleanup 2026-04 — Spec, Mind, and Plan Reorganization

**Binding problem:** The `.spec/` and `.mind/` directories have drifted out of alignment with each other and with the actual state of the work. Memory lives in two places. Two gospel-engine proposals describe overlapping pieces of one system without a clean split. `active.md` lists workstreams that have shipped. Skills and proposals still point at gospel-vec and gospel-mcp by name even though gospel-engine replaced them. Every session pays a context-budget tax to load and reconcile the staleness.

This proposal sorts the plans, separates the engine-vs-interface concerns, retires stale references, and produces a deliverable: cleanly organized `.spec/`, `.mind/`, and `.github/` folders.

**Created:** 2026-04-20
**Scope:** Six related cleanup threads pulled into one coherent execution plan.
**Status:** Proposed — awaiting Michael's review before any destructive action.

---

## 1. The Six Threads

| # | Thread | What's broken |
|---|--------|---------------|
| 1 | `.spec/memory/` ↔ `.mind/` duplication | Six files duplicated. `.mind/` is canonical (Apr 20). `.spec/memory/` is stale (Apr 16). `.spec/README.md` still documents the old layout. |
| 2 | Gospel-engine plans entangled with study.ibeco.me | Engine (data/tooling) and interface (study site) are conflated under one proposal. Graph DB work is split across three places. |
| 3 | Active-vs-shipped drift | `.mind/active.md` and `.spec/proposals/overview/main.md` both list items as "in flight" that shipped weeks ago. |
| 4 | Opus 4.6 missing + tokenomics shift | `raptor-mini` and `claude-sonnet-4.6` references in brain config and proposals will start breaking. GitHub is moving toward token-based billing; the `PremiumRequestCost=0.33/1.0/3.0` mental model is about to change. |
| 5 | Stale `gospel-vec` / `gospel-mcp` references | Skills, proposals, scratch files still teach the old three-server model. Real state: one MCP (`gospel-engine`) plus the hosted `engine.ibeco.me`. |
| 6 | Context-budget bloat | Always-loaded instructions + auto-loaded memory ≈ 25K+ tokens before the user types anything. The token-efficiency proposal exists but hasn't been executed. |

---

## 2. Inventory — What Exists Right Now

### `.mind/` (canonical, Apr 20)
- `identity.md`, `preferences.yaml`, `principles.md`, `decisions.md`, `active.md`
- `archive/active-*.md` — periodic archives of active.md

### `.spec/memory/` (stale duplicate, Apr 16)
- Same six files. Same names. Older content. Last journal-driven update was the token-efficiency proposal entry; nothing since.

### `.spec/proposals/` (33 files / dirs)
The active set, grouped by what they're really about:

| Group | Proposals |
|-------|-----------|
| **Brain — pipeline & orchestration** | `brain-phase4-pipeline.md`, `brain-pipeline-evolution.md`, `brain-pipeline-fixes.md`, `brain-pipeline-fixes-phase4.md`, `brain-project-kanban.md`, `brain-project-ux.md`, `brain-ux-quality-of-life.md`, `brain-inline-panel.md`, `brain-windows-service.md`, `brain-workspace-aware/`, `orchestrator-steward/`, `commission-ui.md`, `commission-ux-fixes.md`, `classify-bench.md`, `classifier-qwen-fix.md`, `project-aware-pipeline.md` |
| **Brain — integration** | `brain-ibecome-layer2.md`, `claude-code-integration.md` |
| **Gospel data layer** | `gospel-engine/` (main + phase1.5), `gospel-engine-postgresql/`, `study-ibeco-me/`, `gospel-graph/`, `enriched-indexer.md`, `embedding-comparison.md` |
| **Studies / content workstreams** | `study-workstream.md`, `teaching-workstream.md`, `sabbath-agent.md` |
| **Memory & process** | `memory-architecture.md`, `token-efficiency.md`, `spec-housekeeping-archive.md`, `debug-layer-triage.md`, `data-safety/`, `lm-studio-model-experiments/` |
| **Cross-cutting** | `overview/` (master plan) |

### `.spec/proposals/overview/main.md`
The "single source of truth" that *should* reconcile everything. Hasn't been updated since the gospel-engine and brain Phase 4 work shipped. References the original 7-MCP architecture in spots.

### `.github/` (16 agents, 14 skills, 1 copilot-instructions)
Voice rules just refreshed today. Cite-count rule extended. But:
- `wide-search` skill still describes the gospel-mcp + gospel-vec split as the canonical search flow
- Agent files inherit voice rules through copilot-instructions but most don't reference the new gospel-engine model
- copilot-instructions is ~14.7KB (per brain's load log) — fully loaded every session

---

## 3. Recommendation Matrix

For each thread: what to do, in what order, and what to defer.

### Thread 1 — Memory Deduplication

**Recommendation: Delete `.spec/memory/` after one verification pass. Update `.spec/README.md`.**

Order:
1. Diff `.spec/memory/active.md` vs `.mind/active.md` to confirm `.spec/memory/` has nothing newer than `.mind/` (already confirmed by git log: Apr 16 vs Apr 20).
2. Diff each of the other five files for any content in `.spec/memory/` not present in `.mind/`. If anything is found, port it.
3. `git rm -r .spec/memory/`.
4. Rewrite `.spec/README.md` to point at `.mind/` and remove the "Memory Files" table. Replace with a one-line pointer: "Memory is in `/.mind/`. See `docs/work-with-ai/` for the rationale."

Risk: low. Git history preserves everything if a port was missed.

### Thread 2 — Gospel-Engine vs study.ibeco.me Split

**Recommendation: Rename and reorganize. Make the engine and the interface clearly two products, with a third proposal for the graph layer.**

Currently:
- `.spec/proposals/gospel-engine/` describes the local single-binary v1. Phase 1-5 shipped.
- `.spec/proposals/gospel-engine-postgresql/` describes the PG migration of v1.
- `.spec/proposals/study-ibeco-me/` is *actually* "host the engine at engine.ibeco.me + serve it via study.ibeco.me UI." It contains both engine-server work and UI work.
- `.spec/proposals/gospel-graph/` is the graph viz, deferred until the engine stabilizes.

The clean split should be:

| New proposal | Owns | Status |
|--------------|------|--------|
| `gospel-engine/` (kept) | Local single-binary engine, MCP tooling, indexing, embedding pipeline. v1 = local. v2 = hosted backend at `engine.ibeco.me`. Includes the PG migration as a phase. | Engine v1 SHIPPED. v2 (engine.ibeco.me) Phase 1-3 SHIPPED today. Phase 1.5 ergonomics open. |
| `study-ibeco-me/` (refocused) | Web interface at `study.ibeco.me` only — search UI, study histories, notes, annotations, frontend. *No backend infra.* Consumes `engine.ibeco.me`. | All Phase 4 ("user features deferred") items become Phase 1+ here. UI scope only. |
| `gospel-graph/` (kept) | Graph layer — Apache AGE on PG, semantic+graph traversal, viz frontend. Consumes `engine.ibeco.me`. Lives in the engine repo as a server module + viz client. | Specced. Waiting on engine v2 stability + AGE on PG18. |
| `gospel-engine-postgresql/` | **Archive.** Folded into `gospel-engine/` as the v1→v2 migration phase. | Mark superseded. |

Concrete rename steps:
1. In `gospel-engine/main.md`, add a Phase Map header: v1 (local, SHIPPED) → v1.5 (ergonomics, OPEN) → v2 (hosted backend at engine.ibeco.me, SHIPPED today) → v3 (graph layer, see gospel-graph proposal).
2. Move the engine-server portions of `study-ibeco-me/main.md` (Phase 1, 2, 3 — DB, embeddings, MCP client, auth delegation) into `gospel-engine/v2-hosted.md`.
3. Rewrite `study-ibeco-me/main.md` to be UI-only: search interface, user history, notes. The "Phase 4 deferred" content becomes the new Phase 1.
4. Mark `gospel-engine-postgresql/main.md` as superseded with a one-line redirect to `gospel-engine/v2-hosted.md`.
5. In `gospel-graph/main.md`, add a "Depends on" header pointing at gospel-engine v2 and AGE PG18 readiness.

This keeps the engine.ibeco.me/study.ibeco.me distinction Michael described: engine = tooling, study = interface.

### Thread 3 — Active-vs-Shipped Reconciliation

**Recommendation: One pass on `.mind/active.md` and `.spec/proposals/overview/main.md`. Then archive what's done.**

In `.mind/active.md`:
- Move all "✓ ALL COMPLETE" items from "In Flight" into a "Recently Shipped" section, then aim to drop them next archive cycle.
- Items currently shipped but listed as in-flight: Brain Project-Kanban (all phases), Orchestrator Steward (P1-6), Commission UI (P1-3), WS3 Brain UX QoL (P1-7b), WS4 Brain Pipeline Evolution (P1-9), engine.ibeco.me Phase 1-3.
- Re-confirm what's actually open: brain inline panel, commission UX fixes, token efficiency, gospel-engine v1.5 ergonomics, gospel-graph (specced), study.ibeco.me UI (post-split).

In `.spec/proposals/overview/main.md`:
- Mark Workstream 1 Phase 4d (REST API + Execution) — confirm shipped or still next.
- Workstream 3 Phase 3 ("gospel-vec experiments") is obsolete; the engine consolidation made it moot.
- Add a new Workstream G — Gospel Engine — that points at the now-clean engine/study/graph split.

After this pass, archive the current `active.md` to `.mind/archive/active-2026-04-20.md` and start a new cycle.

### Thread 4 — Opus 4.6 → 4.7 + Tokenomics

**Recommendation: One audit pass + a new short proposal for the tokenomics rework.** Do NOT bake in numbers until GitHub publishes the new pricing.

Audit targets (grep for these strings):
- `claude-opus-4.6`, `claude-opus-4-6`, `opus 4.6`, `Opus 4.6`
- `claude-sonnet-4.6`, `sonnet-4.6`
- `raptor-mini` (brain's current default, no longer available — confirmed by today's runtime error)
- `PremiumRequestCost=0.33`, `=1.0`, `=3.0`
- `1500 premium/mo`, `$40 Pro+`

For each: replace model identifiers with the new ones (Opus 4.7, current Sonnet, current Haiku), and flag any cost numbers as "subject to revision pending GitHub's token-cost rollout."

Create `.spec/proposals/tokenomics-2026/main.md` with these open questions:
- What does the new GitHub billing model look like (per-token vs per-request)?
- How do brain's `PremiumRequestCost` settings need to change?
- Does the Haiku→Sonnet→Opus→Human escalation ladder still make economic sense?
- What's the right model for orchestrator-steward, commission, classification?

This is a research task, not an implementation task. The implementation comes after GitHub publishes details.

### Thread 5 — Stale Gospel-vec / Gospel-mcp References

**Recommendation: Targeted sweep. Update what teaches; leave what historically documents.**

Two categories:
- **Teaching documents** (instruct the agent how to work today): UPDATE.
  - `.github/skills/wide-search/SKILL.md` — describes 3-step flow using gospel-mcp + gospel-vec. Rewrite to use gospel-engine.
  - `.github/copilot-instructions.md` (already mostly clean per the previous edit).
  - Any agent file that names the old MCPs.
- **Historical / proposal documents** (record what we considered or did): LEAVE.
  - Scratch files like `lm-studio-model-experiments/main.md` reference the old split because that was the state when written. Rewriting them rewrites history.
  - Old proposal bodies (`enriched-indexer.md`, `embedding-comparison.md`) reference the old split because they pre-date the merge.
  - Acceptable change: add a one-line "Note (Apr 2026): superseded by gospel-engine consolidation" header where it's useful to readers.

The grep showed ~60+ matches. Most are in scratch files and old proposals. Maybe 5-10 are in teaching docs that need real updates.

### Thread 6 — Context Budget Bloat

**Recommendation: Measure first. Then act on the existing token-efficiency proposal.**

Today's brain startup log says copilot-instructions is 14,696 bytes. Memory files (`.mind/`) load on top of that. Estimate:
- copilot-instructions.md: ~14.7KB → ~3.7K tokens
- `.mind/identity.md` + `preferences.yaml` + `active.md` (always-load): ~estimate 8-12K tokens
- Skills loaded on demand (not always-on)
- Agent file (~the active mode): 5-10K tokens

Rough session-start budget: 20-30K tokens before the user types anything. The token-efficiency proposal targets ≤10K.

What to do now:
1. Run an actual measurement: word-count `.mind/active.md`, copilot-instructions.md, the typical agent files. Get a real number rather than my estimate.
2. Identify the bloat sources. Suspect: `active.md` carries too much milestone history (the "Recently Shipped" list will be huge).
3. The fix is already specced in `.spec/proposals/token-efficiency.md`. Execute that proposal rather than re-planning it.

If we want a quick win: archive completed-and-stable workstreams out of `active.md` aggressively. The "In Flight" list should only contain things that are genuinely in flight this week.

---

## 4. Phased Delivery

### Phase 1 — Safe Cleanup (this session or next)
Low-risk, immediately useful, all reversible via git.

1. Delete `.spec/memory/` after diff verification (Thread 1).
2. Rewrite `.spec/README.md` to match the new architecture (Thread 1).
3. Update `.github/skills/wide-search/SKILL.md` to teach the gospel-engine flow (Thread 5, teaching docs only).
4. Sweep agent files for `gospel-vec` / `gospel-mcp` (Thread 5, teaching docs only).
5. Reconcile `.mind/active.md`: move shipped items to "Recently Shipped"; archive (Thread 3).

**Deliverable:** `.spec/` and `.mind/` no longer duplicate each other. Skills teach current tools. Active.md reflects reality.

### Phase 2 — Gospel-Engine Reorg
Documentation reorg. No code changes. Reversible.

1. Add Phase Map to `gospel-engine/main.md` (v1→v1.5→v2→v3) (Thread 2).
2. Create `gospel-engine/v2-hosted.md` from the engine-server portions of `study-ibeco-me/main.md` (Thread 2).
3. Rewrite `study-ibeco-me/main.md` as UI-only (Thread 2).
4. Mark `gospel-engine-postgresql/` superseded (Thread 2).
5. Add "Depends on" header to `gospel-graph/main.md` (Thread 2).

**Deliverable:** Engine, interface, and graph are three clearly bounded proposals. Each names one thing.

### Phase 3 — Tokenomics + Model Rotation
Research-led, partly external.

1. Audit grep for Opus 4.6 / sonnet 4.6 / raptor-mini / PremiumRequestCost references (Thread 4).
2. Replace model identifiers with current names (where safe).
3. Create `.spec/proposals/tokenomics-2026/main.md` as a research placeholder for the GitHub billing change (Thread 4).
4. Fix brain's classifier model so auto-classify works again (Thread 4 — this is blocking real work right now).

**Deliverable:** No stale model names in instructions. Brain's auto-classify is unblocked. Tokenomics rework is captured as a research task with open questions.

### Phase 4 — Token Efficiency Execution
Execute the existing proposal rather than re-plan.

1. Measure actual session-start token cost.
2. Run `.spec/proposals/token-efficiency.md` Phase 1 (compress active.md).
3. Run Phase 2 (tiered loading) if measurement says it's worth it.
4. Decide on Phase 3-6 based on what Phase 1-2 freed up.

**Deliverable:** Session-start token cost measured and reduced.

### Phase 5 (deferred) — Brain Pipeline State Audit
After cleanups, query brain itself for what's actually in flight per the entries.

Brain has 112 entries; the relay just synced and surfaced three new gospel-engine-adjacent entries from the inbox today (memory-systems research, GPU training questions, proxy-pointer RAG as gospel-engine v3 candidate). Worth folding these into the engine roadmap.

---

## 5. Verification Criteria

### Phase 1
- [ ] `git status` shows `.spec/memory/` deleted, `.spec/README.md` modified
- [ ] `grep -r "gospel-vec\|gospel-mcp" .github/skills/` returns only historical mentions or zero
- [ ] `.mind/active.md` "In Flight" section ≤ 5 items
- [ ] Session-start memory load is unchanged in size (no regressions)

### Phase 2
- [ ] `gospel-engine/main.md` opens with a Phase Map showing v1/v1.5/v2/v3
- [ ] `study-ibeco-me/main.md` contains zero references to backend infrastructure
- [ ] `gospel-graph/main.md` opens with a "Depends on engine v2 + AGE PG18" line
- [ ] No proposal owns the same scope as another

### Phase 3
- [ ] No `claude-opus-4.6` / `claude-sonnet-4.6` / `raptor-mini` strings in `.github/` or active proposals
- [ ] Brain auto-classify runs without "Model not available" error
- [ ] `.spec/proposals/tokenomics-2026/main.md` exists with open questions

### Phase 4
- [ ] Measured session-start token cost recorded in the token-efficiency proposal
- [ ] active.md size reduction documented

---

## 6. Costs & Risks

### Costs
- **Time:** Phase 1 is one focused session. Phase 2 is one session. Phase 3 is one session of audit + a research placeholder. Phase 4 depends on token-efficiency execution scope.
- **Risk of mis-archive:** Moving items from "In Flight" to "Shipped" requires accuracy. Phase 1 step 5 should be done with brain or a sanity check, not from memory.

### Risks
- **Dropping context the agent needs.** Cutting `.spec/memory/` and trimming `active.md` could lose narrative the next session relies on. Mitigation: archive, don't delete. Git keeps everything.
- **Renaming proposals breaks links.** The `study-ibeco-me/` directory is referenced from many places (active.md, journals, the give-away study). Mitigation: keep the directory name; just refocus the content. Add a "Renamed scope" note at top.
- **Tokenomics moves underneath us.** GitHub hasn't published the new billing yet. Capturing it as a research placeholder is the safest move. Don't bake numbers in.

---

## 7. Creation-Cycle Map

| Step | This proposal |
|------|---------------|
| **Intent** | Stop paying the bloat tax. Make the spec match reality. |
| **Covenant** | No destructive action without verification. Git revisioning is the safety net. |
| **Stewardship** | Plan agent (this proposal). Dev agent or Michael executes. Phases independent. |
| **Spiritual Creation** | Each phase has a deliverable named in §4 + verification in §5. |
| **Line upon Line** | Phase 1 stands alone. Phase 2 is independent. Phase 3 partly external (GitHub billing). Phase 4 depends on token-efficiency proposal already specced. |
| **Physical Creation** | Dev agent or manual session. Each phase ≤ one session. |
| **Review** | Verification checklist per phase. |
| **Atonement** | Git revert. No data loss possible. |
| **Sabbath** | After Phase 1+2 ship, natural pause point for a Sabbath reflection on what the cleanup revealed. |
| **Consecration** | Cleaner spec serves every future session. The bloat tax compounds; so does the cleanup. |
| **Zion** | The whole project benefits — agent context budget, human cognitive load, future contributors (if any). |

---

## 8. Decision Surface

**Recommendation: Build Phase 1 in the next session. Phase 2 in the session after.**

What I need from Michael before executing:
1. **Phase 1 step 1 — delete `.spec/memory/`?** Confirm `.mind/` is canonical and `.spec/memory/` should be removed. (My read: yes, but want explicit approval before destructive action.)
2. **Phase 2 — keep `study-ibeco-me/` as the directory name?** It's referenced everywhere. I'd refocus its scope rather than rename it. Sound right?
3. **Phase 3 — fix brain's `raptor-mini` issue this session, or roll into Phase 3 batch?** It's blocking auto-classify right now. Quick fix is to set the model to `claude-haiku-4.7` or whatever the current Haiku is.
4. **Phase 4 — execute token-efficiency proposal, or re-spec it first?** That proposal is from Apr 16; it may need a refresh.

---

## 9. What This Proposal Deliberately Does NOT Cover

- **Brain pipeline architecture changes.** Brain is healthy; the entries-in-flight problem is a tracking problem, not a brain problem.
- **New gospel-engine features.** The graph proposal, the proxy-pointer RAG idea, the LoRA fine-tuning question — all interesting, all out of scope for cleanup. Each goes in its own proposal.
- **Voice / instruction-harness changes.** Just shipped today. Stable.
- **Becoming app or teaching workstream.** Not part of this cleanup.

If new ideas surface mid-cleanup, file them in `.spec/scratch/` or as brain entries. Don't expand the cleanup scope.
