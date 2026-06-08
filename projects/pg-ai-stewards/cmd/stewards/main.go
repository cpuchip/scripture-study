// stewards — the human cockpit for pg-ai-stewards (read-only P1).
//
// A terminal front-end over the substrate Postgres (pgxpool, like
// cmd/persona-host), so Michael can drive the substrate directly — see the
// board, watch a pipeline, read the cost dashboard — without going through
// Claude. This is P1: read-only verbs only. Dispatch / council / ratify /
// review (the write + Hinge verbs) come in later phases.
//
// Verbs (P1):
//
//	stewards project [<slug>]          show / switch the sticky active project
//	stewards board   [--all] [...]     the work-item board for the active project
//	stewards watch   <id-or-slug>      one work item: stage, status, cost, recent activity
//	stewards cost    [--by ...] [...]  spend dashboard (project × model × day)
//
// Connection: STEWARDS_DSN (default host port 55433 → the dev substrate).
// Active project: ~/.stewards.json (STEWARDS_PROJECT overrides; --project per
// command; --all spans every project).
package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	ctx := context.Background()
	switch os.Args[1] {
	case "project", "proj":
		runProject(ctx, os.Args[2:])
	case "board", "ls":
		runBoard(ctx, os.Args[2:])
	case "watch":
		runWatch(ctx, os.Args[2:])
	case "cost":
		runCost(ctx, os.Args[2:])
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `stewards — the human cockpit for pg-ai-stewards (read-only P1)

Verbs:
  project [<slug>] [--clear]
      No arg: list projects + open/total work-item counts (* = active).
      <slug>: switch the sticky active project (validated against the DB).
      --clear: unset the active project.

  board [--all] [--project <slug>] [--status <s>] [--limit N]
      The work-item board for the active project (or --all to span every
      project). Columns: ref, pipeline, stage, status, maturity, cost,
      tokens, updated. Trailing line summarizes counts by status.

  watch <id-or-slug> [--follow] [--interval N]
      One work item: stage, status, maturity, cost, tokens, escalation,
      error, plus its most recent cost events. --follow polls every N
      seconds (default 3) until the item reaches a terminal status.

  cost [--by project|model|day] [--project <slug>] [--all] [--days N]
      Spend dashboard from cost_events. Grouped by project (default),
      model, or day, over the last N days (default 30). Honors the active
      project unless --all or --project is given.

Environment:
  STEWARDS_DSN      Postgres DSN (default: postgres://stewards:stewards@localhost:55433/stewards?sslmode=disable)
  STEWARDS_PROJECT  Active-project override (beats ~/.stewards.json)
  STEWARDS_CONFIG   Config file path override (default: ~/.stewards.json)
`)
}
