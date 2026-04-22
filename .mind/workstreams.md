# Workstreams — Canonical Taxonomy

*Created: 2026-04-21 · Replaces: [.spec/proposals/archive/overview/main.md](../.spec/proposals/archive/overview/main.md)*

This file is the single source of truth for **how work is grouped**. Every active proposal carries a `workstream:` field in its frontmatter. Every In-Flight row in [active.md](active.md) has a WS column. Brain projects map to workstreams (one project may be split across workstreams; multiple workstreams may share a project).

## Why this exists

Three views — `.spec/proposals/`, `.mind/active.md`, brain DB — describe the same work in different shapes. Workstreams are the shared vocabulary that lets us pivot between them. Without WS tags, every cleanup pass has to reconstruct the grouping from scratch (we just did this on 2026-04-21; it took a session).

## Forward compatibility (Postgres)

This taxonomy is designed to drop into a relational schema. Everything below maps cleanly:

```sql
CREATE TABLE workstream (
  id          TEXT PRIMARY KEY,            -- 'WS1', 'WS2', ...
  name        TEXT NOT NULL,
  description TEXT,
  status      TEXT NOT NULL DEFAULT 'active'
                CHECK (status IN ('active', 'paused', 'retired')),
  brain_project_ids INTEGER[]              -- references brain.projects.id
);

CREATE TABLE proposal (
  path           TEXT PRIMARY KEY,         -- '.spec/proposals/...'
  title          TEXT NOT NULL,
  workstream     TEXT REFERENCES workstream(id),
  status         TEXT NOT NULL
                   CHECK (status IN ('proposed','building','shipped',
                                     'deferred','superseded','archived')),
  brain_project  INTEGER,                  -- brain.projects.id
  brain_entries  TEXT[],                   -- brain.entries.id values
  created        DATE,
  last_updated   DATE,
  superseded_by  TEXT REFERENCES proposal(path)
);
```

Every proposal frontmatter field maps to a column. Status enum is the closed CHECK set — values come from this file, not from memory (per data-safety checklist).

## The Workstreams

| ID | Name | Owns | Brain project(s) | Status |
|----|------|------|------------------|--------|
| **WS1** | Brain Core | Pipeline, steward, commissions, classifier, retry/escalation, model selection, data safety | `6` (2nd Brain) | active |
| **WS2** | Brain UX | UI panels, dialogs, kanban, file viewer, inline panel, Windows service/systray | `6` (2nd Brain) | active |
| **WS3** | Gospel Engine | engine.ibeco.me, gospel-engine MCP, search/index, graph, hosted backend | `3` (Workspace improvements) | active |
| **WS4** | study.ibeco.me | Web UI for studies, notes, reader, public study pages | `5` (ibeco.me) | active |
| **WS5** | Memory & Process | `.mind/`, agents, skills, voice/bias, cleanup passes, tokenomics, brain↔VS Code bridge, debug agent, Claude Code integration, Sabbath agent | `3` (Workspace improvements) | active |
| **WS6** | Studies | Scripture study output (study/, becoming/) | `1` (study) | active |
| **WS7** | Teaching | YouTube content arc, talks, public-facing teaching | `7` (YouTube/Content) | active |
| **WS8** | Sunday School | Calling — lesson prep, ward council | `2` (Sunday School) | active |
| **WS9** | Other Apps | Budget app, cpuchip.net rebuild, Space Center | `4`, `8`, `10` | active |

### Notes on the boundaries

- **WS1 vs WS2.** If a proposal touches data flow, model choice, or pipeline behavior → WS1. If it's about what the human sees and clicks → WS2. Frontend-only changes that don't change backend behavior are WS2.
- **WS3 vs WS4.** Engine = the search/index/graph backend. study.ibeco.me = the user-facing reading and study UI that consumes the engine. They will diverge more over time.
- **WS5 catch-all rule.** If a proposal is about *how we work* rather than *what we ship* — process, memory, agents, voice rules, cleanup — it's WS5. WS5 is the workstream that maintains the other workstreams.

## Status enum (closed set)

Use these exact strings in frontmatter `status:` and in the future Postgres CHECK constraint. **Read from this file; never write from memory.**

| Status | Meaning |
|--------|---------|
| `proposed` | Written, not started. May or may not be approved. |
| `building` | Phase ≥1 in flight. |
| `shipped` | All planned phases done. May still have follow-on but the original scope is closed. |
| `deferred` | Paused intentionally. Has a `revisit_when` condition. |
| `superseded` | Replaced by another proposal. `superseded_by:` points to it. |
| `archived` | Lives under `.spec/proposals/archive/`. Terminal state. |

## Frontmatter convention

Every proposal under `.spec/proposals/` (top level, not `archive/`) carries this YAML frontmatter as the first thing in the file:

```yaml
---
workstream: WS1
status: building
brain_project: 6
created: 2026-04-21
last_updated: 2026-04-21
---
```

Optional fields (add only when meaningful):

```yaml
brain_entries: [abc-123, def-456]   # brain entry UUIDs this proposal owns
phases: 4                           # total planned phases
phase_status: "1-3 shipped, 4 next" # human-readable phase rollup
superseded_by: .spec/proposals/foo/main.md
revisit_when: "After WS1 P4 ships"  # for status=deferred
```

The first prose line below the closing `---` should be the `# Title` of the proposal. The existing `**Status:**` line in the body becomes redundant once frontmatter is in place; keep it for human reading or remove it. Don't have both contradicting each other.

## Mapping (active proposals as of 2026-04-21 Phase C)

This is the snapshot the rest of Phase C edits against. Add a row when a proposal is created; archive removes it.

| Path | WS | Status | Brain project | Notes |
|------|----|--------|---------------|-------|
| `brain-inline-panel.md` | WS2 | building | 6 | P1 reply textarea, P2 nudge bot |
| `brain-project-kanban.md` | WS2 | building | 6 | P1-3 + P4a-4b shipped, P4c next |
| `brain-ux-qol-phase8-autocommit.md` | WS2 | deferred | 6 | After P7 in daily use |
| `brain-vscode-bridge/main.md` | WS5 | proposed | 6 | Sibling of cleanup-2026-04-part2 |
| `brain-windows-service.md` | WS2 | proposed | 6 | Systray |
| `classifier-qwen-fix.md` | WS1 | proposed | 6 | Ready to build |
| `classify-bench.md` | WS1 | proposed | 6 | Bench harness |
| `claude-code-integration.md` | WS5 | proposed | 3 | Researched |
| `cleanup-2026-04/main.md` | WS5 | building | 3 | P1-3 done; P4 → tokenomics |
| `cleanup-2026-04-part2/main.md` | WS5 | building | 3 | This pass — A & B done; C in progress |
| `data-safety/main.md` | WS1 | proposed | 6 | Dev agent hardening + audit log |
| `debug-layer-triage.md` | WS5 | proposed | 3 | Debug agent enhancement |
| `gospel-engine/main.md` | WS3 | building | 3 | v1 shipped, v1.5 next |
| `gospel-engine/phase1.5-ergonomics.md` | WS3 | proposed | 3 | Ergonomic improvements |
| `gospel-engine/v2-hosted.md` | WS3 | shipped | 3 | engine.ibeco.me Apr 20 |
| `gospel-graph/main.md` | WS3 | proposed | 3 | After enriched indexer/PG18 + AGE |
| `sabbath-agent.md` | WS5 | proposed | 3 | Ready to build |
| `study-ibeco-me/main.md` | WS4 | proposed | 5 | UI-only Phase 1+ |
| `study-workstream.md` | WS6 | building | 1 | Rolling parent for study output |
| `teaching-workstream.md` | WS7 | building | 7 | 11-episode arc, content not started |
| `token-efficiency.md` | WS5 | proposed | 3 | Apr 16 — needs refresh |
| `tokenomics-2026/main.md` | WS5 | proposed | 3 | Research placeholder |

22 active proposals across 5 workstreams. WS6/WS7/WS8/WS9 currently have less proposal coverage because their work lives mostly outside `.spec/proposals/` (in `study/`, `teaching/`, `lessons/`, `projects/`).

## When to update this file

- **New workstream:** add a row to "The Workstreams" table.
- **New proposal:** add a row to the mapping table.
- **Proposal moved to archive/:** remove from mapping table.
- **Status enum change:** update the table AND search for the old value across `.spec/proposals/` frontmatter.

When this file changes, also update [active.md](active.md) priorities/in-flight section if a workstream's status shifted.
