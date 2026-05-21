# Active Context

> **2026-05-19/20 — substrate: Council ① CLOSED + YT-T batch shipped same day.** WS5/pg-ai-stewards. ~15 commits across PE-A + PE-B + PE-C + end-to-end smoke + YT-T, zero rollbacks. **YT-T (substrate-yt-transcripts)** wedged in after council ① closed because PE-final's yt-gospel-evaluate against Morgan Philpot produced an honest refusal — bridge had no yt-dlp, workspace yt/ was read-only mounted. YT-T fixed all three layers: yt-dlp in bridge Dockerfile (pip-upgraded to 2026.03.17 from broken Alpine 2024.12.03 pin), workspace yt/ rw-mounted at /opt/yt/yt (bridge writes; pg also gets ro-mount for SQL function reads), and new substrate primitives `yt_transcripts` + `yt_transcript_segments` tables with `import_yt_transcript()` populating them from cues.json + metadata.json. The Morgan Philpot rerun completed at $0.46 with a real substantive 16k-char evaluation citing 8 verbatim quotes with timestamps from the 1516-segment substrate transcript, caught two factual citation errors in the talk, and pushed back theologically on the priesthood-gender segment. Soak resumed. Next: council ② substrate-scheduled-workflows. Canonical navigation: [`projects/pg-ai-stewards/.spec/open-items.md`](../projects/pg-ai-stewards/.spec/open-items.md) §0.

*Hardware: Dual 4090s desktop. NOCIX server live.*

> **Edit rule:** Rewrite this file directly. Do NOT cat/append the old content first — its archive snapshot lives under `.mind/archive/`. Appending duplicates the document and doubles every memory load. (Bug: 2026-04-20.)

> **Workstream taxonomy:** [.mind/workstreams.md](workstreams.md) is canonical: WS1–WS9, status enum, frontmatter convention. Every active proposal carries `workstream:` frontmatter. Read that file before tagging anything new.

> **Model:** GitHub Copilot now runs Claude Opus 4.7 (was 4.6 through April). 4.7 is more literal — see Foresight & Adjacent Surfaces section in [.github/copilot-instructions.md](../.github/copilot-instructions.md). Adjacent Surface Audit is now standard before declaring dev/debug work complete.

---

## Priorities

1. ★ **[WS6] Study** — "It keeps me in the spirit." Latest: [last-supper-four-cups.md](../study/last-supper-four-cups.md) (May 17 — Passover ↔ the Lord's Supper, the four cups, the bitter cup begun in Gethsemane and finished on the cross). Prior presidency arc: art-of-presidency → art-of-delegation → zion-in-a-presidency → broken-heart-and-contrite-spirit. Next: TBD → [study-workstream.md](../.spec/proposals/study-workstream.md)
2. ★ **[WS7] Teaching** — 11-episode arc (Option C). Agent + repo scaffolded. Content not started → [teaching-workstream.md](../.spec/proposals/teaching-workstream.md)
3. **[WS5] Token efficiency** — Apr 16 proposal awaiting refresh review → [token-efficiency.md](../.spec/proposals/token-efficiency.md)
4. **[WS2] Brain Inline Panel + Kanban 4c** — next active build → [brain-inline-panel.md](../.spec/proposals/brain-inline-panel.md) · [brain-project-kanban.md](../.spec/proposals/brain-project-kanban.md)
5. **[WS5] pg-ai-stewards** — Postgres-as-AI-substrate; an autonomous agentic creation cycle. **Phases A–F + Batches G/H/I/J/K/L/L.1.1 + ES arc + Council ① (PE-A + PE-B + PE-C + final smoke) all shipped (2026-05-04 → 2026-05-19).** Next: council ② `substrate-scheduled-workflows` (cron jobs, D-SW1–7; inherits PE-B's scheduled_pipelines machinery) → ③ `stewards-ui-evolution` (D-UI1–12; PE-C was a preview slice). Directional lens: **hybrid work-together modes** — the substrate as a workspace human + agent co-manage. Soak RUNNING. Deferred: pg rebuild + replay-86-post-G-SQL ritual (touched zero times in council ①). → [open-items.md](../projects/pg-ai-stewards/.spec/open-items.md) §0 · [pipelines-expansion §X+§XI](../projects/pg-ai-stewards/.spec/proposals/substrate-pipelines-expansion.md)

## Key Facts

- Copilot: 1500 premium/mo ($40 Pro+). Haiku 4.5=0.33x, Sonnet 4.6=1.0x, Opus 4.7=7.5x, GPT-5/5-mini/4.1/4o=0
- Brain default model: gpt-5-mini (0x). Pipeline big = claude-opus-4.7 (7.5x)
- **Brain stage defaults (Apr 23 cost discipline):** research=sonnet, plan=opus, spec=sonnet, execute=sonnet, verify=haiku (hard-pinned), revise=sonnet. Commission `Model` field = steward judgment only (gate eval). Revise loop capped at 2 → surface.
- Claude Code: Pro $20/mo. 200K context. Project caching
- Pipeline cost: research=0.33 + plan=1.0 = 1.33/entry  *(stale — see stage defaults above)*
- Active MCP servers: gospel-engine-v2 (engine.ibeco.me), webster, yt, byu-citations, becoming, exa-search
- gospel-engine v2 hosted is the single canonical search backend; gospel-mcp + gospel-vec retired as fallback

---

## In Flight

| WS | Item | Status | Ref |
|----|------|--------|-----|
| WS7 | cpuchip.net — personal site revival (AI + Gospel) | 🔨 revival arc shipped — deploy, LCARS visual arc, scripture-panel component. Own repo + own `.mind`/`.spec`/`CLAUDE.md`. Next: republish two studies. | [projects/cpuchip.net/.mind/active.md](../projects/cpuchip.net/.mind/active.md) |
| WS2 | Brain Inline Panel + Nudge Bot Controls | ▶ P1 next | [brain-inline-panel.md](../.spec/proposals/brain-inline-panel.md) |
| WS2 | Brain Project-Kanban Phase 4c | ▶ next | [brain-project-kanban.md](../.spec/proposals/brain-project-kanban.md) |
| WS5 | Token Efficiency & Memory v2 | ⏸ awaiting refresh | [token-efficiency.md](../.spec/proposals/token-efficiency.md) |
| WS5 | Brain ↔ VS Code Bridge | 🔨 building (Phase 0 shipped Apr 22) | [brain-vscode-bridge/main.md](../.spec/proposals/brain-vscode-bridge/main.md) |
| WS2 | Johari window agent mode | 📝 proposed Apr 22 | [johari-window-agent-mode.md](../.spec/proposals/johari-window-agent-mode.md) |
| WS5 | pg-ai-stewards (Postgres substrate for agent state, memory, work, model calls) | ✅ **Council ① + YT-T CLOSED 2026-05-19/20**. YT-T added yt-dlp to bridge, workspace yt/ rw mount, native `yt_transcripts` table. Morgan Philpot rerun produced a real 16k-char evaluation. Next: ② scheduled-workflows → ③ ui-evolution. Direction: hybrid work-together modes. | [open-items.md](../projects/pg-ai-stewards/.spec/open-items.md) §0 |
| WS7 | 1828-illuminated scriptures (1828.ibeco.me) | 🔨 **Backend pivot under planning 2026-05-20.** MVP + 3 stretch goals + connectivity work + route-reactivity fix shipped. Thummim batch RUNNING in background (14 verified + 7 in_progress, $5.76 / $9 hard-stop, zero failures — `charity` self-corrected Moroni 7:46 → 7:48). Michael authorized backend + Postgres + public-domain scripture corpus + class-E word reach (every 1828 entry queryable). **Five proposals written by plan agent at `projects/1828-illuminated/.spec/proposals/`**: `backend-pivot.md` (umbrella) + `scripture-corpus.md` + `dictionary-backend.md` + `llm-proxy.md` + `deployment-shape.md`. **BLOCKING DECISION D-BE-COPYRIGHT** before any building: bcbooks/scriptures-json says "2013 LDS edition" — not unambiguously PD. Options A (accept), B (strip to 1769/1830/1835/1851 sources), or C (keep iframe for canon, backend for dict + LLM + pasted-text only). ~30 other decisions batched downstream. Commits ac687b9 (route fix) + b66444d → 153ff43 (overnight). | [proposals](../projects/1828-illuminated/.spec/proposals/) · [project README](../projects/1828-illuminated/README.md) · [Thummim proposal](../.spec/proposals/thummim-restoration-dictionary.md) |
| WS2 | Motivation coach agent mode | 📝 proposed Apr 22 | [motivation-coach-agent-mode.md](../.spec/proposals/motivation-coach-agent-mode.md) |
| WS3 | LightRAG investigation | 📝 proposed Apr 22 | [lightrag-investigation.md](../.spec/proposals/lightrag-investigation.md) |
| WS3 | Gospel engine v3 proxy-pointer | 📝 proposed Apr 22 | [gospel-engine-v3-proxy-pointer.md](../.spec/proposals/gospel-engine-v3-proxy-pointer.md) |
| WS5 | VS Code agent hooks integration | 📝 proposed Apr 22 | [vscode-agent-hooks-integration.md](../.spec/proposals/vscode-agent-hooks-integration.md) |
| WS5 | Memory & context research bundle | 📝 proposed Apr 22 | [memory-research-bundle.md](../.spec/proposals/memory-research-bundle.md) |
| WS6 | Study: Nate B Jones on delegation | 📝 proposed Apr 22 | [study-nate-jones-delegation.md](../.spec/proposals/study-nate-jones-delegation.md) |
| WS7 | AI presentation site tool | 📝 proposed Apr 22 | [ai-presentation-site-tool.md](../.spec/proposals/ai-presentation-site-tool.md) |
| WS7 | Launch YouTube channel | 📝 proposed Apr 22 | [launch-youtube-channel.md](../.spec/proposals/launch-youtube-channel.md) |
| WS5 | Sabbath agent | ready to build | [sabbath-agent.md](../.spec/proposals/sabbath-agent.md) |
| WS5 | Debug agent: layer triage | proposed | [debug-layer-triage.md](../.spec/proposals/debug-layer-triage.md) |
| WS5 | Claude Code Integration | researched | [claude-code-integration.md](../.spec/proposals/claude-code-integration.md) |
| WS5 | tokenomics-2026 (research) | placeholder | [tokenomics-2026/main.md](../.spec/proposals/tokenomics-2026/main.md) |
| WS1 | Classifier qwen fix | ready to build | [classifier-qwen-fix.md](../.spec/proposals/classifier-qwen-fix.md) |
| WS1 | Classification quality benchmark | proposed | [classify-bench.md](../.spec/proposals/classify-bench.md) |
| WS1 | Data safety: dev hardening + audit log | proposed | [data-safety/main.md](../.spec/proposals/data-safety/main.md) |
| WS2 | Brain Windows Service (systray) | proposed | [brain-windows-service.md](../.spec/proposals/brain-windows-service.md) |
| WS3 | Gospel Engine v1.5 + research rollup | ratified 2026-05-13 — 6 phase files at `scripts/gospel-engine-v2/.spec/proposals/` (1.5a docs → 1.5b–1.5e code+indexer → single reindex → 3-research). Phase 2 TITSW deferred. | [rollup README](../scripts/gospel-engine-v2/.spec/proposals/README.md) · [parent (superseded)](../.spec/proposals/gospel-engine/phase1.5-ergonomics.md) |
| WS3 | Gospel Graph Visualization | proposed (blocked on AGE/PG18) | [gospel-graph/main.md](../.spec/proposals/gospel-graph/main.md) |
| WS4 | study.ibeco.me UI | proposed | [study-ibeco-me/main.md](../.spec/proposals/study-ibeco-me/main.md) |
| WS4 | ibeco.me Security Audit | not started | — |

---

## Recently Shipped (last ~30 days)

| WS | Item | Shipped | Notes |
|----|------|---------|-------|
| WS5 | pg-ai-stewards — YT-T batch (substrate-yt-transcripts) | May 19/20 | Post-council-① wedge after PE-final's yt-gospel-evaluate produced a refusal (bridge had no yt-dlp). Five sub-steps: (YT-T.1) yt-dlp in bridge.Dockerfile via pip install --upgrade — Alpine apk pin to 2024.12.03 was broken on every YouTube video, upgraded to 2026.03.17; (YT-T.2) workspace yt/ rw-mounted into bridge + ro into pg so the SQL function can read; (YT-T.3) `yt_transcripts` + `yt_transcript_segments` schema (D-YTT3 separate FK table for time-range queries; D-YTT4 no pgvector, engrams handle that); (YT-T.4) `stewards.import_yt_transcript(video_id)` reads /opt/yt/yt cache via pg_read_file, populates both tables idempotently; (YT-T.5) Morgan Philpot rerun completed at $0.46 with substantive 16k-char evaluation — 8 verbatim quotes with timestamps, caught 2 factual errors in talk, pushed back on priesthood-gender segment citing Oaks. Bridge crash-loop sub-incident (scratch file in extension/ broke migration discovery) caught + cleaned. Soak resumed. |
| WS5 | pg-ai-stewards — Council ① CLOSED (PE-A + PE-B + PE-C + final smoke) | May 19 | One calendar day, 13+ commits, zero rollbacks. PE-A: 3 new pipelines + studies/AGE promotion + backfill. PE-B: scheduled_pipelines table + plpgsql cron parser + dispatcher wired into watchman tick + ai-news-7am seed (reframed from Rust cron crate to plpgsql to avoid the 86-file replay event). PE-C: full-CRUD /scheduled Vue route + Dashboard scheduled-runs card + NewWork.vue per-pipeline forms. PE-final: all three pipelines (research-summary, yt-secular-digest with Nate B Jones, yt-gospel-evaluate with Morgan Philpot) dispatched + completed + verified + promoted to studies + AGE. Total spend $0.36. One bug surfaced + fixed mid-smoke (yt-* missing auto_materialize_on_verified). Three reframes: D-PE1' Option B, D-PE2' reuse general-research, D-PE7' build missing promotion path. |
| WS5 | pg-ai-stewards — ES emergency-stop arc (ES.1/3/4/5/6/3.s5) | May 15–17 | Bacteriopolis runaway diagnosed + closed: bleed-stoppers, judge-compiled-brief, streaming chat dispatch, fs/PDF/consult follow-ups, gateway upstream-cost capture. ~95 commits, zero rollbacks; the substrate runs a research pipeline clean to a verified artifact (~$0.33). |
| WS5 | pg-ai-stewards — Batches G/H/I/J/K/L/L.1.1 | May 11–14 | File-write mechanism (G), research/planning pipelines (H), agent write-back (I), fan-out + brainstorm (J), engram context compaction (K), Context Engine v2 + v2.1 Judges pattern (L/L.1.1). |
| WS5 | pg-ai-stewards — Phases A–F (agentic creation cycle) | May 10–11 | Watch→Diagnose→Act→Account loop, maturity ladder + gates, intent/covenant, atonement/sabbath/consecration, trust ladder, multi-agent council. |
| WS6 | Study — last-supper-four-cups | May 17 | Passover ↔ the Lord's Supper; the four cups; the bitter cup begun in Gethsemane, finished on the cross. |
| WS2 | brain-steward-cost-discipline | Apr 23 | Three-defect fix: (1) commission `Model` no longer overrides every stage — `modelForStage` helper routes through `config.StageDefaults`; `c.Model` reserved for `EvaluateGate` only. (2) Verify hard-pinned to haiku regardless of catalog. (3) Revise loop capped at 2 → surface `loop_limit_exceeded`. `RevisionCount` field on Commission with DB migration; "Revised X/2" badge on EntryDetailView. Same fix applied to `commissionWaitForExecution` (the loop that actually burned 105 credits). Best-case opus commission ~25 credits (was ~52); worst case ~28-35 (was unbounded, hit 105). Followup Apr 23: research bumped haiku→sonnet per Michael ("stronger model researching is good") — chain still escalates from sonnet, just one fewer step before quarantine. |
| WS2 | brain-model-catalog-sot | Apr 23 | Single source of truth at `internal/config/models.go`: `Catalog` slice + `StageDefaults` map. Two pre-existing drifted maps (`modelCosts`, `AvailableModels`) and `steward.EscalationChain` now derive from it. New `GET /api/models` endpoint; frontend composable + dynamic dropdowns in CommissionDialog and ProjectDetailView inline commission. Default is now Claude Opus 4.7 (7.5×) instead of stale Opus 4.6 (3.0×). Stewardship sweep: same-bug-same-fix on `feedbackDialog` and `executeDialog` modals (UA-stylesheet dark-theme bug from doneDialog fix earlier same day). Note: `/api/models` path collided with legacy LM Studio profiles handler — moved to `/api/models/profiles` (no consumers found). |
| WS2 | brain non-pipeline kanban flow (Phases 1-3 + classify gate) | Apr 23 | Status vocab gained `working` (in-progress lane). `boardColumns` branches on `pipeline_enabled`: manual path uses literal status keys (active/working/done) instead of route-status pipeline. 5-button manual row (▶ Start / ✓ Done / ↩ Reopen / ⏸ Someday / 🗄 Archive) on board + list views. Optional reason dialog on ✓ Done appends a `_Closed YYYY-MM-DD: reason_` line to body. Native HTML5 drag-and-drop between columns (no library dep). Auto-classify gated for non-pipeline projects in relay client. Done-dialog theming fix (centered, larger, dark-theme textarea). `handleCreateEntry` now accepts `project_id` — single-call POST round-trips correctly. brain-app P4 mirror skipped per user direction (no Project surface). |
| WS2 | brain non-pipeline projects (Phases 1+2) | Apr 23 | `pipeline_enabled` column on `projects` (default true). Notebook auto-flipped to false. Single primary gate in `routeEntry`; defense-in-depth in `BuildProjectContext`. UI checkbox + 📓 badge on web. Inverse-hypothesis verified via PUT-set project_id + explicit `/api/agent/route`. |
| WS2 | brain manual stage transitions (Phase 1 mobile) | Apr 23 | brain-app: status dropdown in edit screen, long-press quick-actions sheet (done/park/waiting/archive/reactivate) + undo, semantic chip colors. Phase 2 web: pre-existing — `EntryDetailView.vue` already had full STATUS_OPTIONS dropdown. |
| WS2 | brain status-aware-views ecosystem parity (Phases 1+2) | Apr 23 | ibeco.me `ListBrainEntries` honors `?include_parked=1` (default off). brain-app history `_showParked` covers both someday+archived. Phase 3 cross-surface check pending Dokploy deploy. |
| WS5 | Opus 4.7 harness tuning | Apr 23 | Foresight & Adjacent Surfaces section in copilot-instructions.md, dev.agent.md update, Council Moment extended to dev/debug/ux. Diagnosis + fix for literalism failures. |
| WS2 | brain-status-aware-views (Phases 1-3 + Dashboard) | Apr 23 | All planned phases shipped. Phase 4 (Capture semantic) deferred. Server-side `/api/entries` filter (`?include_parked=1` opts in). Project board toggle relocated to visible header checkbox. Dashboard agent surfaces (routable/review/your-turn) filter parked server-side. |
| WS2 | brain-status-field-on-list-queries (full SELECT audit) | Apr 23 | All 6 list queries now expose `status`: ListAll, ListCategory, ListEntriesByProject, ListByRouteStatus, ListUnassigned, ListPipeline. Future filter UIs no longer guess data layer contracts. |
| WS2 | Brain audit 2026-04-22 | Apr 22 | 69 entries triaged. Personal merged into Notebook. inbox 39→10, status=NULL 96→5 |
| WS5 | Brain↔VS Code bridge Phase 0 | Apr 22 | Schema migration (workstream + proposal_path columns). Read-only inspector at scripts/harness/harness_inspect.py |
| WS5 | cleanup-2026-04 + part2 (all phases) | Apr 21-22 | Spec dedup, 19 proposals archived, workstream taxonomy + frontmatter convention, active.md rewritten |
| WS5 | Voice/bias harness v2 | Apr 21 | em-dash budget, three-beat pivot, refrains, stats cite-count |
| WS3 | engine.ibeco.me Phase 1-3 | Apr 20 | Hosted gospel search. Token UI. First study used it |

*Older items (Apr 4–15: Commission UX, Orchestrator Steward, Commission UI, Brain UX QoL, Brain Pipeline Evolution, Brain Project-Kanban P1–4b) rolled off the 30-day window — recorded in git history + `.spec/journal/`.*

---

## Deferred / Paused

| Item | Revisit When |
|------|--------------|
| Brain UX QoL Phase 8 (auto-commit) | After human-in-loop signal stabilizes |
| Plan 17: Proactive Surfacing | WS2 Phase 3 |
| Plan 18: Widget Overhaul (Ph 3-4) | Agent infra proves out |
| Plan 19: Brain App Ideas | Natural pause |
