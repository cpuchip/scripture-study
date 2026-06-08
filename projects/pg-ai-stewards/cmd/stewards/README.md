# stewards — the human cockpit (P1, read-only)

A terminal front-end over the substrate Postgres so Michael can drive
pg-ai-stewards directly — see the board, watch a pipeline, read the cost
dashboard — without going through Claude. Implements **Option A** of the cockpit
fork in `docs/ai-utilization-landscape-2026.md`, spec'd in
`.spec/proposals/stewards-cockpit-cli.md` (RATIFIED 2026-06-07).

It connects with a `pgxpool` to the substrate Postgres (the same pattern
`cmd/persona-host` uses) and calls only existing tables — **no new engine, no
writes.** Every P1 verb is a read.

## Build

```sh
cd projects/pg-ai-stewards/cmd/stewards
GOWORK=off go build -o stewards.exe .
```

## Verbs (P1)

| Verb | Does |
|---|---|
| `stewards project [<slug>]` | List projects + open/total work-item counts (`*` = active), or switch the sticky **active project** (validated against the DB). `--clear` to unset. |
| `stewards board [--all] [--project S] [--status S] [--limit N]` | The work-item board for the active project (or `--all` to span). Columns: ref, pipeline, stage, status, maturity, cost, tokens, updated. Trailing line summarizes counts by status. |
| `stewards watch <id-or-slug> [--follow] [--interval N]` | One work item: stage, status, maturity, cost, tokens, escalation, error, plus its recent cost events. `--follow` polls (read-only) until a terminal status. |
| `stewards cost [--by project\|model\|day] [--project S] [--all] [--days N]` | Spend dashboard from `cost_events`, grouped by project / model / day over the last N days (default 30). Honors the active project unless `--all`/`--project`. |

## Active project

A sticky context (like a `kubectl` context or the current git branch) that scopes
the work-item verbs. Resolution order:

1. `STEWARDS_PROJECT` env var (override)
2. `~/.stewards.json` (written by `stewards project <slug>`)
3. none → verbs span all projects (with a hint)

`--project X` overrides per-command; `--all` spans every project.

## Environment

| Var | Default | Meaning |
|---|---|---|
| `STEWARDS_DSN` | `postgres://stewards:stewards@localhost:55433/stewards?sslmode=disable` | Substrate Postgres DSN (host-mapped port 55433 → the dev container). |
| `STEWARDS_PROJECT` | — | Active-project override. |
| `STEWARDS_CONFIG` | `~/.stewards.json` | Config file path override. |

## Roadmap (from the spec)

- **P1 (this)** — read-only `project / board / watch / cost`.
- **P2** — `planning_state` ladder (idea → spec → ratified → building → blocked →
  done) on work items; `board` renders it; `carry-over.md` becomes a generated view.
- **P3** — `do` (create + dispatch a work item).
- **P4** — Hinge verbs: `council` (pre-ratify critical pass), `ratify` (input
  Hinge), `review` (output Hinge — approve escalations / PRs).
- **P5** — `personas` / `chat`, `brain`.
- **P6** — the same cost + board queries power a `stewards-ui` panel (Option B).

P1 is deliberately read-only: immediately useful, zero risk, and it proves the
pgxpool + verb surface before any write path lands.
