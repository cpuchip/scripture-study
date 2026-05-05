# stewards-cli

Cross-platform Go CLI for the `pg_ai_stewards` extension. Replaces the
PowerShell-based `import-studies.ps1` and `stewards.ps1` with one
binary that compiles cleanly to `windows/amd64` and `linux/amd64`.

Built for Phase 2.5 ("Generic Document Substrate") so the Linux
hosting server and the Windows dev box use the same tool.

## Build

```powershell
# Native
go build -o stewards-cli.exe .

# Linux deploy target
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o stewards-cli-linux .
```

## Connection

```
STEWARDS_DSN=postgres://stewards:stewards@localhost:55433/stewards?sslmode=disable
```

Defaults match `projects/pg-ai-stewards/extension/docker-compose.yaml`
(port 55433 mapped to the container's 5432).

## Commands

### `import --source <kind>:<dir-or-file> [--source ...]`

Repeatable. Supported kinds: `study`, `doc`, `proposal`, `phase-doc`,
`journal`. Each kind has its own parser (see
[internal/importer](internal/importer)).

Slug strategy (Phase 2.5):
- `study/charity.md` → `charity` (root-level study, bare basename)
- `study/talks/art-of-delegation.md` → `talks-art-of-delegation`
- `study/yt/foo.md` → `yt-foo`
- `docs/work-with-ai/01_x.md` → `doc-01_x`
- `.spec/proposals/foo.md` → `proposal-foo`
- `.spec/journal/2026-05-04--x.yaml` → `journal-2026-05-04--x`

Why bare basenames for root-level studies: every existing reference to
study slugs in the corpus (and in similarity edges, citations, memory)
stays valid. Non-study kinds and subdir-nested studies get prefixes to
prevent the silent overwrites the old PowerShell importer suffered
(`art-of-delegation.md` existed in both `study/` and `study/talks/` —
the second one always clobbered the first).

Full corpus import:

```
stewards-cli import \
    --source study:study \
    --source doc:docs/work-with-ai \
    --source proposal:.spec/proposals \
    --source phase-doc:projects/pg-ai-stewards/phases.md \
    --source journal:.spec/journal
```

### `study show <slug> [--sim N --cites N --verse-chars N]`

Calls `stewards.study_show()` and prints the formatted blob. Works on
any kind, not just studies — the function signature stayed `study_show`
because renaming would break every existing caller.

**Note flag ordering:** Go's stdlib `flag` package stops at the first
non-flag arg, so put `--sim`/`--cites` BEFORE the slug:

```
stewards-cli study show --sim 5 --cites 10 my-slug    # ✓ works
stewards-cli study show my-slug --sim 5               # ✗ flags ignored
```

### `study list [--kind <kind>]`

Tab-aligned (kind, slug, embedded date, title). Filter by kind to
narrow.

### `study refresh [<slug>]`

Re-runs `refresh_study_refs` and `refresh_study_similarity` for one
slug, or corpus-wide when omitted.

## Phase 2.5 verification (corpus state as of 2026-05-04)

| Kind | Count |
|------|-------|
| study | 188 |
| journal | 65 |
| proposal | 73 |
| doc | 32 |
| phase-doc | 1 |
| **total** | **359** |

Similarity edges: 1,795 across all 359 documents. Resolves: 2,625.

The cross-kind bridge prediction landed clearly:

- `creation` (study) — top mutual neighbor at 0.915 is
  `doc-01_planning-then-create-gospel`. #2 at 0.868 is
  `journal-2026-01-21--project-genesis` (the day the project began).
- `doc-01_planning-then-create-gospel` — its top-3 are other gospel/AI
  docs in the same series, then `creation` itself, then a proposal.
- `stewardship-pattern` (study) — connects to `art-of-delegation`,
  the journal entry from the day it was written, AND
  `proposal-archive-orchestrator-steward-main` /
  `proposal-study-nate-jones-delegation`. The graph rendered a
  workstream from pure embedding similarity.
- `proposal-pg-ai-stewards-phase-2-5-generic-substrate` — top three
  mutual neighbors are the journal entries from Phase 2.2, 2.3, 2.4.
  The substrate just rendered the timeline of its own creation.

## Known issues

- **Em-dash mojibake on Windows PowerShell stdout.** Stored data is
  correct UTF-8 (verified via `encode(...::bytea, 'hex')`); the
  corruption only happens in the PowerShell console's output decoder.
  Use `Out-File -Encoding utf8` and read the file, or run from WSL/Linux.
  Not a code bug — a Windows console issue. `chcp 65001` does not fix it.
- **`phases.md` is one document for now.** Per-phase splitting is
  Phase 2.6.
