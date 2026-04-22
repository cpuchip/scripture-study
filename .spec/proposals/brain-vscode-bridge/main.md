---
workstream: WS5
status: building
brain_project: 6
created: 2026-04-21
last_updated: 2026-04-22
phase_status: "Phase 0 shipped 2026-04-22 (workstream + proposal_path columns added to brain.db, harness_inspect.py CLI). Phase 1+ Go-side work pending."
---

# Brain ↔ VS Code Bridge — Bidirectional Sync Between Plans and Work Items

**Binding problem:** `.spec/proposals/` is the long-form planning view. The brain's `entries` table is the actionable kanban. They describe the same work but have no link between them. Work in VS Code updates the proposal — the brain doesn't know. Work in the brain advances entries — the proposal stays stale. Every reconciliation pass is a manual cross-reference (see [cleanup-2026-04-part2/main.md](../cleanup-2026-04-part2/main.md) for the latest one). This burns tokens and Michael's attention.

This proposal designs a structural fix: make proposals and brain entries explicitly linkable, give each side hooks to update the other, and add a brain-side cleanup pass that closes its own loops.

**Created:** 2026-04-21
**Type:** Architecture / system design
**Status:** Building — Phase 0 (schema + read-only Python inspector) shipped 2026-04-22. Go-side bridge pending.
**Sibling:** [cleanup-2026-04-part2](../cleanup-2026-04-part2/main.md) (the immediate reconciliation work this is meant to make obsolete)

---

## Phase 0 — Schema beachhead (SHIPPED 2026-04-22)

Before the Go-side bridge can sync anything, brain needs the vocabulary. Phase 0 added it:

- `entries.workstream TEXT` — backfilled from project_id with per-entry overrides. 98/115 entries tagged.
- `entries.proposal_path TEXT` — backfilled by scraping body for `.spec/proposals/...` references. 14 entries linked.
- `scripts/harness/harness_inspect.py` — read-only CLI showing per-workstream proposal+entry alignment, gaps (mature entries without proposal_path), and orphans (proposals with no brain entry).

Migration script: `.spec/scratch/brain-audit-2026-04-22/harness_phase1_migration.py` (idempotent).

This is the structural prerequisite for everything below. The Go bridge work can now reference `workstream` and `proposal_path` as first-class columns instead of inferring them.

---

## 1. The Drift Mechanism

```
  Proposal author (Claude in VS Code)              Brain runtime (commission, pipeline)
       |                                                     |
       | writes/edits .spec/proposals/X.md                   | advances entry maturity
       | updates .mind/active.md                             | runs commission goroutines
       | NO write to brain DB                                | NO write to .spec/ or .mind/
       v                                                     v
  Static markdown files                                  brain.db
  (git-tracked, human-readable)                       (sqlite, structured)

       \\__________________ disconnected __________________//
```

Both sides do good work. Neither side notifies the other. The result: 96 brain entries with `status=None`, several proposals stuck on "Proposed" while the work shipped, an active.md that lists shipped work as in-flight.

## 2. Design Principles

1. **Both views stay first-class.** Proposals are for rationale, phasing, and human reading. Brain entries are for actionable units, kanban, and AI runtime. Neither replaces the other.
2. **Linkage is opt-in but enforced where it exists.** Not every proposal needs entries. Not every entry needs a proposal. But when a link exists, both sides honor it.
3. **One direction of "truth" per field.** Proposal status is owned by the proposal file. Entry maturity is owned by brain. The bridge synchronizes references, not data ownership.
4. **Brain cleans up after itself.** Every commission run, every steward action, every status change writes a one-line journal entry to a known location. Reconciliation passes can read that journal instead of reverse-engineering state.

## 3. Architecture

### 3a. Schema additions

**New `entries` columns:**
- `proposal_path TEXT` — workspace-relative path to the proposal file (e.g. `.spec/proposals/brain-inline-panel.md`)
- `proposal_phase TEXT` — phase identifier within the proposal (e.g. `P2`, `Phase 4c`)
- `workstream TEXT` — WS1–WS9 tag (per cleanup-2026-04-part2 §4 Phase C)

**New `projects` columns:**
- `workstream TEXT` — primary workstream
- `proposal_path TEXT` — pointer to the umbrella proposal if any

**Proposal frontmatter convention** (already markdown, no schema change):

```yaml
---
workstream: WS1
brain_entries:
  - { id: "abc-123", phase: "P1" }
  - { id: "def-456", phase: "P2" }
brain_project: 6
status: building
last_synced: 2026-04-21
---
```

### 3b. Sync surfaces

**Three sync surfaces, each unidirectional and explicit:**

1. **Proposal → Brain (`brain link` command):**
   - From VS Code: agent runs `brain link <proposal-path> [--phase X] [--entry-id ID]`
   - Effect: writes `proposal_path` + `proposal_phase` to the entry, updates the proposal frontmatter `brain_entries` array
   - Idempotent

2. **Brain → Proposal (status hook):**
   - When brain advances an entry to `verified` or `done`, it checks for `proposal_path`
   - If set: writes a one-line markdown entry to a `## Brain Sync Log` section at the bottom of the proposal: `- 2026-04-21 entry abc-123 (P1) → verified`
   - Does NOT auto-edit the Status: line — Michael reviews and updates that himself

3. **Brain → Memory (`active.md` candidates):**
   - Brain emits a daily JSON snapshot to `.mind/brain-state.json` with: open commissions, entries advanced today, top 10 stale `your_turn` entries
   - The agent reads this at session start and uses it to suggest active.md updates (doesn't apply automatically)

### 3c. Brain self-cleanup

**Steward-side hook (in `scripts/brain/internal/steward/`):**

After every commission completion, retry exhaustion, or quarantine, the steward writes:
1. The commission decision (already happens)
2. A one-line journal entry to `private-brain/journal/{date}.jsonl` with: action, entry_id, proposal_path (if any), result, cost
3. If `proposal_path` set: the brain → proposal sync hook (3b.2)

**Daily nudge bot extension:**

Add a "stale entries" pass to the existing nudge cycle. Definition of stale:
- `route_status = your_turn` for > 7 days AND
- `category != journal` AND
- no session_messages added in the last 7 days

Action: surface a single nudge entry: "N stale entries — review or close" with a link to the kanban filtered view. Don't auto-close.

### 3d. VS Code side

A small MCP tool group on the brain MCP server (already running):

| Tool | Purpose |
|------|---------|
| `brain_link_proposal` | Write proposal_path/phase to an entry |
| `brain_proposal_entries` | List entries linked to a given proposal_path |
| `brain_workstream_status` | Snapshot of all entries in a workstream — what's open, what shipped this week |
| `brain_stale_entries` | Top-N entries that have been "your_turn" longest |

These let the planning agent (this agent) ask "what's the brain say about WS3?" without grepping markdown.

## 4. Phased Plan

### Phase 1 — Schema + Frontmatter Convention (small)

- Add the four new columns (entries.proposal_path, entries.proposal_phase, entries.workstream, projects.workstream, projects.proposal_path) via brain migration
- Document the proposal frontmatter convention in `.spec/conventions.md` (new file)
- No code changes outside the migration. **Deliverable: schema is ready to be populated.**

### Phase 2 — `brain link` command + 4 MCP tools

- CLI: `brain link <proposal> [--phase X] [--entry ID]`
- MCP: the four tools in §3d above
- The tools are the cheap layer that makes Phase 1 useful immediately, even before the auto-sync hooks

### Phase 3 — Brain → Proposal sync hook

- After commission/steward terminal events, write the Sync Log line if `proposal_path` is set
- Test against a real shipped commission

### Phase 4 — Daily snapshot + stale nudges

- `.mind/brain-state.json` written on a daily cron (or when brain.exe starts each day)
- Stale-entry pass added to the nudge bot
- Agent picks up `brain-state.json` at session start as a tier-1 read (cheap, structured)

### Phase 5 — Backfill (one-time pass)

- Walk every active proposal, identify the brain entries it covers (manual + agent-assisted)
- Run `brain link` on each
- After this, the bridge is *retroactively* useful for the existing 35 active proposals + 107 open entries

## 5. What This Does NOT Do

- It does not auto-edit the Status: line on a proposal. Michael keeps that authority.
- It does not auto-archive proposals when their entries all complete. Same reason.
- It does not move work between brain projects. Project assignment stays manual.
- It does not deduplicate the 96 statusless entries. That's a one-time review Michael needs to do, not something the bridge can do correctly.

## 6. Verification Criteria

| Phase | Done when |
|-------|-----------|
| 1 | Migration applies cleanly. New columns exist. Frontmatter convention doc in `.spec/conventions.md`. |
| 2 | `brain link` works. The 4 MCP tools return real data for at least one linked proposal. |
| 3 | A shipped commission writes a `## Brain Sync Log` line to its linked proposal. Manually verified. |
| 4 | `.mind/brain-state.json` updates daily. Stale nudge fires for at least one real stale entry. |
| 5 | At least the 5 most active proposals (cleanup-2026-04, brain-inline-panel, gospel-engine, study-ibeco-me, teaching-workstream) have linked brain entries. |

## 7. Costs and Risks

- **Cost:** Phase 1 is small (~1 day). Phase 2 is medium (~2-3 days). Phase 3-4 each ~1-2 days. Phase 5 is mostly Michael time, agent-assisted.
- **Risk:** New columns + sync logic = new surface area for bugs. Mitigation: every sync action is auditable in `private-brain/journal/`.
- **Risk:** Backfill could mis-link entries. Mitigation: every `brain link` is reversible (NULL the columns); agent should propose links and let Michael confirm in batches.
- **Won't help:** Doesn't address the brain UI's commission/kanban UX issues (separate proposals).

## 8. Open Questions

1. Should `brain_entries` frontmatter live in the markdown file or in a sidecar `.spec/proposals/X.brain.json`? Frontmatter is human-readable but mutable; sidecar is cleaner but invisible.
2. Stale threshold — 7 days reasonable, or should it scale by category (actions=3d, ideas=30d)?
3. Should the daily snapshot include cost summaries (premium requests used per workstream)? Useful for tokenomics-2026 but adds scope.
4. Should this proposal absorb the brain-side cleanup ideas, or stay focused on the bridge and let cleanup live separately?

## 9. Recommendation

**Build Phase 1-2 first** (~3-4 days). They give immediate value (link + query) without changing brain runtime behavior. Then evaluate whether Phase 3-4 are worth building or if the manual `brain link` workflow is enough.

Don't build all five phases up front. The first two are the load-bearing ones.

## 10. Creation Cycle Check

| Step | Notes |
|------|-------|
| Intent | Stop manually reconciling proposal ↔ brain every cleanup cycle |
| Covenant | Linkage is opt-in. Brain never auto-edits proposal Status. Michael keeps authority. |
| Stewardship | Brain dev agent owns implementation. Michael owns the workstream taxonomy. |
| Spiritual Creation | Schema + frontmatter shape defined here. Phasing keeps each step independently shippable. |
| Line upon Line | Phase 1 alone makes future cleanup faster. Each subsequent phase is additive. |
| Physical Creation | brain Go code + minor MCP additions. |
| Review | Per-phase verification criteria above. |
| Atonement | Migrations are reversible. Sync log is append-only. |
| Sabbath | Natural pause after Phase 2 — evaluate before continuing. |
| Consecration | Reduces every future session's overhead permanently. |
| Zion | Brings the engineering plan view and the runtime kanban view into one coherent system without merging them. |
