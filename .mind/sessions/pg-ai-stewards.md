---
lane: pg-ai-stewards
session_id: 7ea7faa4-688a-451a-ac68-b7ea662d4b81
status: active
started: 2026-06-11T22:00:16
last_active: 2026-06-12T15:52:35
---

## Working on
- **★ P1 EXTRACTION UNDERWAY (kicked off 2026-06-12, Michael's "Lets kick off P1!"):**
  (1) `github.com/cpuchip/pg-ai-stewards-workspace` (PRIVATE) created at
  `projects/pg-ai-stewards-workspace/` — skeleton + covenant/intent overlay
  copies + 241-file classification (`overlays/classification.tsv`: 191 core /
  17 core-p2 / 27+1 overlay / 5 mixed / 1 scratch) + 33-entry overlay manifest
  + all overlay migrations populated. (2) OSS extension layer extracted
  (`3d8229d`): src/*.rs audited, lib.rs chain reworked (4 seed embeds removed),
  189 core + 5 SPLIT migrations, 193-entry core manifest, bundle = build
  artifact (never checked in). **Build GREEN + virgin CREATE EXTENSION proven**
  (scratch container, 0 workspace seeds leaked) → OSS pushed through journal.
  NEXT P1: Go daemon re-homing, runner-consumes-manifest, compose, seed pack.
  (3) Private manifest REPAIRED (root `e5ccc0c3`): 9 live-applied migrations
  (r11-r17, ct2-5, ct2-7e) restored from ledger order; found the runner is
  LEXICAL + manifest-blind (replayed scratch-ct2-run2 into live 06-10 —
  codewright-ct2 rows; disposition = Michael's call).
- **pg-ai-stewards OSS extraction** (continues the `pg-ai-stewards-oss` lane —
  same session, retitled): spec RATIFIED, Apache-2.0 FINAL (`3c43d4e`).
  **"Anatomy of a Turn" SHIPPED (`0e8c3c9`)** + order-research update +
  2026 regrounding (`1a604af`). **Cutover gate AMENDED (`8662448`,
  ratified 06-12): FULL PARITY before the cut** — coder-mcp + UI become
  pre-cutover (P2 before cut), 20 mismatches + ledger wart now on the
  cutover critical path. Next: P1 extraction (task #151), side-by-side
  compose (`stewards-oss-*`, 55434/8081/8091). Overlay design ratified
  (`0e01a04`): private repo pg-ai-stewards-workspace, created at P1
  kickoff. jumpstart-crossover reflection seeded (`48864a47`, no build).
- **PR.1 SHIPPED + live-verified 2026-06-12** (inbox assignment, Michael's
  "best of your judgement" grant): covenants.extensions catch-all +
  presiding render + Watch echo; reseed through the real path; smoke
  `600f6673` ACK with presiding terms in the dispatched payload. Journal:
  `projects/pg-ai-stewards/.spec/journal/2026-06-12-pr1-covenant-extensions.md`.
  Carry-forwards there: walls-vs-compulsion audit (§V), trailing-reminder
  proposal, verify-suite full run, ledger naming wart.
- **compact_context SEED captured** (Michael's sketch, 2026-06-12 — HOLD,
  no build until council): commissioned-curation side quest; seed at
  `projects/pg-ai-stewards/.spec/proposals/substrate-compact-context-sidequest.md`
  with parked council questions. 2026 research also on hold per Michael.

## Claims
- NONE live. (PR.1 window CLOSED 2026-06-12: pg+bridge rebuilt/restarted,
  watchman resumed, queue clean, live smoke verified. Persona-host
  container was never touched.)
- The general-workspase lane owns the containerized persona-host
  (acknowledged; will not restart it).

## Handoffs / notes
- 2026-06-12: Anatomy doc is public — sibling sessions citing substrate
  architecture can link github.com/cpuchip/pg-ai-stewards/blob/main/docs/anatomy-of-a-turn.md.
- Supersedes lane file `pg-ai-stewards-oss.md` (same session_id; hook
  re-claimed under the new title).
