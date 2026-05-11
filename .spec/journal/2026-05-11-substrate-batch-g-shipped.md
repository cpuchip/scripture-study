---
date: 2026-05-11
mode: build
workstream: WS5
project: pg-ai-stewards
title: "Batch G shipped — substrate lands in real files"
status: shipped
carry_forward:
  - "Bridge restart + soak re-enable at session end"
  - "First real materialization run on a verified study (validate G.4 end-to-end with real content)"
  - "Watch whether human reviewers actually USE the per-work_item file-destination UI vs leave it default"
links:
  - "../proposals/substrate-completion-batch-g.md"
  - "../../projects/pg-ai-stewards/extension/6e-lesson-resolution-file-producers.sql"
  - "../../projects/pg-ai-stewards/CLAUDE.md"
---

# Batch G — substrate lands in real files (2026-05-11)

## What shipped

Eight commits in one session closed Batch G — the "make the substrate land in real files" bridge between Postgres-as-substrate and the repo's filesystem. Substrate-side state stays the source of truth; files are now an optional materialization, not a precondition.

**G.1 (9f5ce97)** — `studies.file_path` NOT NULL was blocking promote_to_study. Dropped the constraint; studies can exist with file_path NULL and be materialized later (or never).

**G.2 (9b400bf)** — `steward_tick` now calls `retry_guidance_with_lessons` instead of `retry_guidance`, so retries pull the last 3 ratified lessons for the (pipeline, stage) cell into the prompt context. Closes the Phase E loop where lessons existed but the steward wasn't reading them.

**G.3 (1ab531a)** — Quarantine path in `steward_tick` now fires `maybe_enqueue_atonement(work_item_id)` at cost-cap quarantine. Atonement no longer requires a human to remember to trigger it. Surfaced one follow-up gap: `failure_count_limit` quarantine doesn't currently fire atonement either — left as Phase 4a follow-up.

**G.4.1 (a4dae0a)** — Schema: `pipelines.file_destination_template` + `file_content_jsonpath`; `work_items.file_destination` + `materialized_at`; `stewards.pending_file_writes` table (target_path, write_mode='create'|'append', content, source_kind, source_id, materialized_at, materialized_by). `enqueue_work_item_file(p_work_item_id, p_requested_by)` returns NULL when file_destination is NULL (DB-only mode) — opt-in is the default. Seeded study-write + study-write-qwen with `'study/substrate--<slug>.md'`.

**G.4.2 (89470ad)** — `stewards-cli materialize-writes` subcommand. Flags: --dry-run, --limit, --repo-root. Queries unmaterialized rows; path-escape protection (rejects paths starting with `/`, containing `..`, or absolute); create mode refuses overwrite (skip + log); append mode creates if missing. On success: sets `materialized_at + materialized_by='cli'`; for source_kind='work_item' also UPDATEs `studies.file_path` so the path lives in two consistent places.

**G.4.3 (2312be1)** — Pre-commit hook integration. `materialize_writes()` function added to `scripts/git-hooks/pre-commit` after the existing intent/covenant reseed functions. Locates binary at `projects/pg-ai-stewards/bin/stewards-cli.exe` → `bin/stewards-cli` → PATH; quick count check before invoking CLI; STEWARDS_DSN env or default. `bin/*.exe` is gitignored so the hook script ships but the binary is built locally.

**G.4.4 (6bd41ca)** — UI: `/api/pipelines/list` returns templates; `/api/work-items/set-file-destination` + `/api/work-items/materialize-file` endpoints. NewWork.vue: dynamic pipeline dropdown (replaces the old hardcoded list); writeFile checkbox + fileDestination input that prefills from the pipeline's template with `<slug>` substitution. WorkItemDetail.vue: new file destination panel between Steward status and Scenarios — edit mode with Save/Cancel/"Use pipeline default" + "Materialize now" button (only shown when destination set and not yet materialized).

**G.4.5 (a26b63b)** — Lesson + resolution producer hooks. Two AFTER UPDATE OF promoted_to triggers + producer functions wire the existing ratify / accept buttons into pending_file_writes. Lessons → append mode + dated section header (`## YYYY-MM-DD — <kind> (<slug>)`); resolutions under `.mind/` → append, under `study/` → create with frontmatter (binding question + resolved_by + resolved_at). Trigger bodies wrap producer calls in EXCEPTION → NOTICE so the original UPDATE always succeeds (file-write failure doesn't block ratification).

## What this batch unlocks

The substrate now has end-to-end file-write capability without losing its substrate-first discipline:

- **DB-default, opt-in** — work_items with `file_destination = NULL` (which is most of them, since pipelines suggest but don't require) never produce files. The substrate stays FS-stateless.
- **Three-layer decision** — pipeline UI suggestion → per-work_item explicit set (or unset) → explicit materialization gesture. Each layer can defer or override the previous.
- **Generalized** — not study-specific. Any pipeline that wants a file deliverable can declare `file_destination_template` and the same machinery serves it. Lessons promote to `.mind/principles.md`; resolutions promote to `study/<id>.md` or `.mind/decisions.md`; future pipelines (research, video evals, episode prep) plug in the same way.
- **Pre-commit-integrated** — files land in the working tree *just before* `git commit`, so the working tree is always either materialized + committed OR not materialized at all. No drift between substrate and repo.

## What this means for Michael's three larger proposals

The three substantial directions Michael added to the plan (research/yt pipelines, UI authoring for intents+covenants, substrate-aware chat) all benefit from G's file-write capability:

- **More pipelines** — each new pipeline (research, yt-gospel, yt-secular, scheduled news) gets `file_destination_template` for free; output flows to repo without per-pipeline plumbing.
- **UI authoring of intents/covenants** — same producer pattern: edit YAML in the UI → INSERT pending_file_writes → pre-commit hook materializes → existing YAML reseed hook picks it up. The intent/covenant round-trip closes.
- **Substrate-aware chat (read-only/write modes)** — write mode is now well-defined: chat can INSERT into pending_file_writes (with `requested_by='chat'`) and the same materialization path applies. No need to invent a new file-write surface.

## Discoveries / lessons

**1. Producer pattern is the right architecture for "DB → file" wiring.** Three different sources (work_item materialize button, lesson ratify button, council accept button) all converge on `INSERT INTO pending_file_writes`. The consumer (CLI + pre-commit hook) doesn't need to know about the sources. This is the substrate equivalent of dependency inversion.

**2. Triggers > Go-handler duplication.** G.4.5 could have been "extend lessonsRatifyHandler in Go + extend resolve_council in SQL." Doing both in SQL via AFTER UPDATE OF promoted_to triggers means the file-write hook fires regardless of *how* promoted_to gets set (CLI, MCP tool, future API endpoint, manual psql). The substrate stays the single source of truth.

**3. Swallowing trigger errors via EXCEPTION → NOTICE.** The standard Postgres trigger pattern would let a pending_file_writes INSERT failure roll back the ratify UPDATE. Wrong: a human ratifying a lesson shouldn't be blocked by a file-write hiccup. Following Phase D's pattern (sabbath_dispatch errors swallowed via NOTICE so the parent gate_decision still applies) gives us the right semantics — ratification succeeds, file-write surfaces as a NOTICE for later inspection.

**4. Smoke test surfaced real schema friction in councils.** The G.4.5 smoke caught three sequential NOT NULL / CHECK violations on `councils` (intent_id, convened_by, bishop, status). Not a bug — the constraints are correct — but a reminder that synthetic test fixtures need to match production shape or they don't exercise the real path. Used `'resolved'` for status to bypass the one_active_council partial unique index.

**5. The bin/*.exe gitignore caught the hook integration.** Pre-commit hook needs the binary, but the binary is gitignored. Solved by documenting rebuild instructions in the hook script and the commit message rather than checking the binary in. Trade-off: fresh clones need a `go build` before the hook works.

## Substrate state at session end

All six substrate phases (A-F) + Batch G feature-complete on dev. The substrate now has:

- **Watch → Diagnose → Act → Account loop** (Phase A) — bgworker tick + steward + cost tracking + escalation chain
- **Maturity ladder + gate machinery** (Phase B) — raw → researched → planned → specced → executing → verified with gate prompts at each rung
- **Intent + Covenant as first-class state** (Phase C) — YAML-seeded, prepended to every compose_system_prompt, tools-disabled gates
- **Post-completion ritual** (Phase D) — Sabbath reflection + Atonement triggered on quarantine + Lessons as audit ledger
- **Trust ladder + line-upon-line** (Phase E) — per-(agent, pipeline, model) trust scores, gate overrides with justification, retry pulls last 3 ratified lessons
- **Multi-agent council** (Phase F) — proposer/critic/synthesizer roles, master-tier bishop eligibility, accept/revise/dissolve with optional file destination
- **File materialization** (Batch G) — pending_file_writes producer pattern, three-layer per-pipeline-per-item-per-gesture decision, pre-commit hook integration

**Substrate proposal at `full-agentic-substrate.md` is fully reality.** Next moves are about USE, not BUILD.

## What's next (carry-forward to next session)

1. **Bridge restart + soak re-enable** (this session end)
2. **First real materialization run** — convene a real council on a real intent or run a study end-to-end with a real file_destination, watch a `.md` actually land in the working tree
3. **UI surface check** — does the per-work_item file-destination UI feel right in practice, or does the human want to set destinations after seeing the output? (G.4 already supports post-hoc setting via WorkItemDetail; this is about whether the default flow surfaces it at the right time.)
4. **The three big proposals Michael added** — more pipelines (research, yt-gospel/secular, scheduled news), UI authoring for intents+covenants, substrate-aware chat with read-only/write modes. All three are unblocked by Batch G.

Cost this session: ~$0 — entirely smoke-test and schema work, no real LLM dispatches.

## Covenant moment

Mid-G.4 design, Michael flagged: *"because this is the first time pg-ai-stewards will have file access right? will this be generalized for any pipeline... I'd want to make sure this is configurables, or even settable after we finish a study all the way through as I'm reviewing them."*

That intervention reshaped G.4 from a study-specific shortcut into the three-layer generalized pattern that actually shipped. Initial design was: study work_items always produce `study/<slug>.md`. Final design: pipelines suggest, work_items decide (or defer), humans materialize on demand. The difference is the difference between a feature and a substrate primitive.

This is the `flag_when_wrong` covenant commitment doing exactly what it's supposed to do — catching an over-fit design before it became infrastructure. Recording it here because next time a "first time the substrate gets X capability" question comes up, the pattern is clear: ask whether the capability is generalizable, configurable, and settable post-hoc *before* building the specific case.
