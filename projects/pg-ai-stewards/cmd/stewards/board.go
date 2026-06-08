package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"
)

// runBoard renders the work-item board for the active project (or --all).
//
// P1 groups by status/maturity — the existing work-item dims. P2 will add the
// planning_state ladder (idea → spec → ratified → building → blocked → done)
// and render that instead; the query here is the seam that grows.
func runBoard(ctx context.Context, args []string) {
	fs := flag.NewFlagSet("board", flag.ExitOnError)
	all := fs.Bool("all", false, "span every project")
	projectFlag := fs.String("project", "", "scope to this project (overrides active)")
	status := fs.String("status", "", "filter by status")
	limit := fs.Int("limit", 100, "max rows")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	pool := mustConnect(ctx)
	defer pool.Close()

	scope, label := boardScope(*all, *projectFlag)
	showProject := scope == "" // spanning → include a project column

	rows, err := pool.Query(ctx, `
		SELECT coalesce(slug, left(id::text, 8))      AS ref,
		       coalesce(project_association, '(none)') AS project,
		       pipeline_family, current_stage, status, maturity,
		       cost_micro_dollars,
		       (tokens_in + tokens_out)::bigint        AS tokens,
		       updated_at,
		       (status NOT IN ('completed','cancelled')) AS open
		FROM stewards.work_items
		WHERE ($1 = '' OR project_association = $1)
		  AND ($2 = '' OR status = $2)
		ORDER BY open DESC, updated_at DESC
		LIMIT $3`, scope, *status, *limit)
	if err != nil {
		fail("query board", err)
	}
	defer rows.Close()

	var table [][]string
	counts := map[string]int{}
	for rows.Next() {
		var ref, project, pipeline, stage, st, maturity string
		var micro, tokens int64
		var updated time.Time
		var open bool
		if err := rows.Scan(&ref, &project, &pipeline, &stage, &st, &maturity,
			&micro, &tokens, &updated, &open); err != nil {
			fail("scan board", err)
		}
		counts[st]++
		row := []string{ref}
		if showProject {
			row = append(row, project)
		}
		row = append(row, pipeline, stage, st, maturity,
			fmtMicro(micro), fmtTokens(tokens), relTime(updated))
		table = append(table, row)
	}
	if err := rows.Err(); err != nil {
		fail("iterate board", err)
	}

	fmt.Printf("Board — %s\n\n", label)
	if len(table) == 0 {
		fmt.Println("no work items")
		return
	}

	headers := []string{"REF"}
	aligns := []align{alignLeft}
	if showProject {
		headers = append(headers, "PROJECT")
		aligns = append(aligns, alignLeft)
	}
	headers = append(headers, "PIPELINE", "STAGE", "STATUS", "MATURITY", "COST", "TOKENS", "UPDATED")
	aligns = append(aligns, alignLeft, alignLeft, alignLeft, alignLeft, alignRight, alignRight, alignRight)

	printTable(headers, table, aligns)
	fmt.Printf("\n%s\n", summarizeCounts(counts))
}

// boardScope resolves the project filter and a human label. An empty scope
// means "every project" (the spanning case, which adds a project column).
func boardScope(all bool, projectFlag string) (scope, label string) {
	switch {
	case all:
		return "", "all projects"
	case projectFlag != "":
		return projectFlag, projectFlag
	}
	if ap := activeProject(); ap != "" {
		return ap, ap
	}
	return "", "all projects (no active project — 'stewards project <slug>' to scope)"
}
