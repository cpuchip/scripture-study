package main

import (
	"context"
	"flag"
	"fmt"
	"os"
)

// runCost renders the spend dashboard from cost_events, grouped by project,
// model, or day. This is the CLI half of the shared "token dashboard by project
// × model" — the same aggregation a stewards-ui panel will read (Option B).
func runCost(ctx context.Context, args []string) {
	fs := flag.NewFlagSet("cost", flag.ExitOnError)
	by := fs.String("by", "project", "group by: project | model | day")
	days := fs.Int("days", 30, "look back this many days")
	projectFlag := fs.String("project", "", "scope to this project")
	all := fs.Bool("all", false, "span every project (ignore the active project)")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	// cost is lenient about scope: an explicit --project wins; else the active
	// project unless --all; else everything (a money overview shouldn't error
	// just because no project is selected).
	scope := *projectFlag
	scopeLabel := "all projects"
	if scope != "" {
		scopeLabel = scope
	} else if !*all {
		if ap := activeProject(); ap != "" {
			scope = ap
			scopeLabel = ap
		}
	}

	var bucketExpr, bucketHeader, orderBy string
	switch *by {
	case "project":
		bucketExpr = "coalesce(w.project_association,'(none)')"
		bucketHeader = "PROJECT"
		orderBy = "micro DESC"
	case "model":
		bucketExpr = "ce.provider || '/' || ce.model"
		bucketHeader = "MODEL"
		orderBy = "micro DESC"
	case "day":
		bucketExpr = "to_char(date_trunc('day', ce.at), 'YYYY-MM-DD')"
		bucketHeader = "DAY"
		orderBy = "bucket DESC"
	default:
		fmt.Fprintf(os.Stderr, "cost --by must be project|model|day, got %q\n", *by)
		os.Exit(1)
	}

	pool := mustConnect(ctx)
	defer pool.Close()

	q := fmt.Sprintf(`
		SELECT %s AS bucket,
		       sum(ce.input_tokens + ce.output_tokens)::bigint AS tokens,
		       sum(ce.micro_dollars)::bigint                    AS micro,
		       count(*)::bigint                                 AS events
		FROM stewards.cost_events ce
		LEFT JOIN stewards.work_items w ON w.id = ce.work_item_id
		WHERE ce.at >= now() - make_interval(days => $1)
		  AND ($2 = '' OR w.project_association = $2)
		GROUP BY 1
		ORDER BY %s`, bucketExpr, orderBy)

	rows, err := pool.Query(ctx, q, *days, scope)
	if err != nil {
		fail("query cost", err)
	}
	defer rows.Close()

	var table [][]string
	var totMicro, totTokens, totEvents int64
	for rows.Next() {
		var bucket string
		var tokens, micro, events int64
		if err := rows.Scan(&bucket, &tokens, &micro, &events); err != nil {
			fail("scan cost", err)
		}
		totMicro += micro
		totTokens += tokens
		totEvents += events
		table = append(table, []string{
			bucket, fmtMicro(micro), fmtTokens(tokens), fmt.Sprintf("%d", events),
		})
	}
	if err := rows.Err(); err != nil {
		fail("iterate cost", err)
	}

	fmt.Printf("Cost — %s, last %d days, by %s\n\n", scopeLabel, *days, *by)
	if len(table) == 0 {
		fmt.Println("no cost events in range")
		return
	}
	// total row
	table = append(table, []string{
		"TOTAL", fmtMicro(totMicro), fmtTokens(totTokens), fmt.Sprintf("%d", totEvents),
	})
	printTable(
		[]string{bucketHeader, "COST", "TOKENS", "EVENTS"}, table,
		[]align{alignLeft, alignRight, alignRight, alignRight},
	)
}
