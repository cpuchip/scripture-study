# Brain Audit — 2026-04-22

## Landscape

| Metric | Count |
|--------|-------|
| Total entries | 115 |
| In inbox (no project) | 39 |
| `status IS NULL` (lifecycle never closed) | 96 |
| `action_done=1` but status not closed | 12 |
| `maturity=raw` | 67 |
| `maturity=verified/specced/planned/complete` | 48 |

## How brain entries map to workstreams

| WS | Entries with workspace proposal already | Entries needing a workspace home | Notes |
|----|------------------------------------------|----------------------------------|-------|
| WS1 Brain Core | classifier-qwen-fix, classify-bench, data-safety, embedding-comparison (specced/planned, all in inbox by mistake) | — | All have proposals; entries just need project=6 ("2nd Brain") + status=specced |
| WS2 Brain UX | brain-project-kanban (verified, in inbox), brain-windows-service, ~6 done brain-app bug reports | "Use Kanban tools instead of building" idea, "Brain app task structure" idea, "Fix captured widgets" bug | Mostly closure work; few decisions |
| WS3 Gospel Engine | gospel-engine FTS+vec (verified, inbox), gospel-engine 1.5 ergonomics (planned, inbox), gospel-graph (specced, inbox) | "Multi-step agents for gospel RAG", "LightRAG investigate/continue", "MetaClaw memory" | Inbox entries duplicate proposals; raw research links could mostly archive |
| WS4 study.ibeco.me | — | "Implement Gateway Auth for Projects" (ibeco.me, verified) | Real proposal exists for UI; gateway auth is a separate idea worth its own one-liner proposal |
| WS5 Memory & Process | memory-architecture, sabbath-agent, claude-code, brain-workspace-aware, session-journal, brain-relay (all specced/verified, in inbox) | "Brain research pre-step pipeline", "AI personal assistant beyond second brain", "Brain as agent OS" | Lots of fertile ideas — some should become proposals, most can be tagged `someday` |
| WS6 Studies | study-workstream (planned, inbox) | "Intelligence study article ref" (link only) | Just needs project assignment |
| WS7 Teaching | teaching-workstream (planned, inbox) | — | Just needs project assignment |
| WS8 Sunday School | — | "Reflect back what wife says" (this is personal/relational, not SS) | Inbox has 1 SS entry waiting; no major proposals |
| WS9 Other (Space Center, cpuchip.net, Budget, Notebook) | LCARS theme done, Marshfield science center verified | "Pull cpuchip.net from wayback", "Build PC hardware", "Custom Desk Project", grocery/temple visit personal notes | Personal items not workstream-tracked — fine as is |

## Categories of cleanup

### A. Mechanical closure (no decisions, ~25 entries)

Entries where `action_done=1` or the description literally says "Shipped." Mark `status=done`, set project, archive.

- All 8 brain-app bug reports in 2nd Brain with `done=1` → `status=done`
- 7 "verified" entries in inbox tagged "Shipped/Complete" (Session Journal Tool, Becoming App, Desktop Notifications, TITSW Enrichment, Gospel Engine FTS+vec, Multi-Agent Routing, Brain Relay, Brain Project-Kanban) → project=6, status=done
- Personal done items (ATX case, cold meds, keda) → status=archived

### B. Inbox → project move (no decision, ~12 entries)

Specced/planned proposals that are in inbox by classifier mistake. Move to correct project, set status=`active`:

- 9 specced inbox entries → all point to existing proposal files. project=6 (2nd Brain) or 3 (Workspace).
- 4 planned inbox entries (gospel-engine 1.5, teaching, study, classify-bench) → likewise.

### C. Raw research links — mass-archive candidates (~25 entries)

The "Workspace improvements" raw bucket is mostly youtube links + "should we look at X?" research items. Given raised costs, most of these are deferral candidates. One-line rationale per item, then `status=someday` or `status=archived`.

Examples: AI Dungeons YT, Trinity 57B model, LightRAG, MetaClaw, Stripe minions, Brave MCP, AutoAgent, Hormozi delegation video, etc.

A few may be worth promoting to proposals (LightRAG specifically — it's been mentioned twice).

### D. Genuine inspiration — needs decision (~10 entries)

Ideas that don't have proposals but might deserve them:

| Entry | Workstream | Recommendation |
|-------|------------|----------------|
| "Mount DB as filesystem for AI collaboration" | WS1/WS5 | Worth a one-liner proposal as "explore" — connects to brain-vscode-bridge |
| "Brain as agent OS platform" | WS1 | Already covered by current pipeline direction; `someday` |
| "Use existing tools like Kanban" | WS2 | Push-back on our own build. Either capture as decision or archive (we chose to build) |
| "AI personal assistant beyond second brain" | WS1 | `someday` — too broad |
| "Brain research pre-step pipeline" | WS1 | Worth a proposal — concrete and aligned |
| "Johari Windows AI agent" | WS5 (?) | Interesting; capture as research note |
| "Gospel engine v3 proxy-pointer" (raw) | WS3 | Worth reading the full body before deciding |
| "Pull cpuchip.net from wayback" | WS9 | Personal project; fine as `someday` |
| "Implement Gateway Auth for Projects" | WS4 | Real infrastructure need; should become proposal or merge into ibeco.me security audit |
| "Marshfield science center" (Space Center) | WS9 | Already in Space Center project; just needs status |

### E. Personal notes (~5 entries)

Grocery, temple visit, KISS reminder, wife-reflection, custom desk. Not workstream-tracked. Keep in inbox or move to a "Personal" project. **Question for Michael: do we want a project=11 "Personal" bucket, or leave these in inbox?**

## Harness recommendations

The user is asking the deeper question: **how do we keep brain ↔ proposals ↔ active.md in sync going forward?** This is exactly what `brain-vscode-bridge` was written to solve. The proposal already exists but is unbuilt. Three concrete moves:

1. **Add `workstream` field to brain `entries` table** (migration: `workstream TEXT` referencing the WS1-WS9 enum). Backfill from project_id with a default mapping (project 6 → WS1/WS2 split by triage). This gives brain the same vocabulary as proposals.

2. **Build brain-vscode-bridge Phase 1** (read-only): a CLI command `brain sync workstream <ws>` that lists entries + proposal frontmatter side-by-side. Doesn't write anything yet — just exposes the drift.

3. **Add a "needs-proposal" flag to brain entries.** When an entry is `maturity=planned` or `specced` and has no `proposal_path`, it surfaces as a TODO. Closes the gap where ideas escape into research without a workspace artifact.

For inspiration — **two unworked entries point at this exact need:**
- "Use existing tools like Kanban" — pushes back on our build-everything tendency. The harness should be opinionated about *not* building things we can buy.
- "Mount DB as filesystem for AI collaboration" — natural fit for the bridge architecture. The brain DB becomes a first-class file the agent reads, not a tool API.

## Recommended phased execution

**Phase A — mechanical (this session, ~30 min):** Categories A + B. ~37 entries closed/relocated with one batched UPDATE. No decisions needed beyond approval to run.

**Phase B — research-backlog triage (this session or next, ~30 min):** Category C. Walk the ~25 raw research entries with Michael, decide `someday | archived | promote-to-proposal` per item. Optionally batched as 5-row prompts.

**Phase C — inspiration triage (next session):** Category D. ~10 entries. Each gets a clear verdict; the 3-4 that survive become real proposals.

**Phase D — personal bucket (5 min):** Category E. One question: do we want a "Personal" project?

**Phase E — harness build (separate workstream):** Promote `brain-vscode-bridge` to `building`. Add the `workstream` field migration. Ship Phase 1 (read-only sync inspector) before adding write-back.

## Postgres-future alignment

When the framework moves to PG, the `entries` table needs a `workstream TEXT NOT NULL` column with the same CHECK constraint as `proposal.workstream` (per `.mind/workstreams.md`). The `proposal_path` column makes the brain entry → proposal file link queryable. Both are pure additions; current SQLite schema can absorb them via goose without disruption.
