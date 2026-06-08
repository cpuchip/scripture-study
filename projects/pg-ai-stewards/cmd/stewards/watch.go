package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// terminalStatuses are the work-item statuses at which --follow stops polling.
var terminalStatuses = map[string]bool{
	"completed": true,
	"cancelled": true,
	"failed":    true,
}

type wiSnapshot struct {
	id          string
	slug        string
	pipeline    string
	stage       string
	status      string
	maturity    string
	project     string
	micro       int64
	tokensIn    int64
	tokensOut   int64
	escalation  string
	errText     string
	createdAt   time.Time
	updatedAt   time.Time
	completedAt *time.Time
}

// runWatch shows one work item: stage, status, cost, and recent activity. With
// --follow it polls (read-only) until the item reaches a terminal status.
func runWatch(ctx context.Context, args []string) {
	fs := flag.NewFlagSet("watch", flag.ExitOnError)
	follow := fs.Bool("follow", false, "poll until the item reaches a terminal status")
	interval := fs.Int("interval", 3, "poll interval in seconds (with --follow)")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "watch: <id-or-slug> required")
		os.Exit(1)
	}
	ref := fs.Arg(0)

	pool := mustConnect(ctx)
	defer pool.Close()

	for {
		snap, err := loadWorkItem(ctx, pool, ref)
		if err != nil {
			fail("watch", err)
		}
		fmt.Print("\033[H\033[2J") // best-effort clear; harmless if unsupported
		printSnapshot(ctx, pool, snap)

		if !*follow || terminalStatuses[snap.status] {
			if *follow {
				fmt.Printf("\n(terminal: %s)\n", snap.status)
			}
			return
		}
		time.Sleep(time.Duration(*interval) * time.Second)
	}
}

// loadWorkItem resolves a work item by exact slug, exact id, or id prefix.
func loadWorkItem(ctx context.Context, pool *pgxpool.Pool, ref string) (*wiSnapshot, error) {
	rows, err := pool.Query(ctx, `
		SELECT id::text, coalesce(slug,''), pipeline_family, current_stage, status, maturity,
		       coalesce(project_association,'(none)'), cost_micro_dollars,
		       tokens_in, tokens_out, escalation_state, coalesce(error,''),
		       created_at, updated_at, completed_at
		FROM stewards.work_items
		WHERE slug = $1 OR id::text = $1 OR id::text LIKE $1 || '%'
		ORDER BY updated_at DESC
		LIMIT 2`, ref)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var found []*wiSnapshot
	for rows.Next() {
		var s wiSnapshot
		if err := rows.Scan(&s.id, &s.slug, &s.pipeline, &s.stage, &s.status, &s.maturity,
			&s.project, &s.micro, &s.tokensIn, &s.tokensOut, &s.escalation, &s.errText,
			&s.createdAt, &s.updatedAt, &s.completedAt); err != nil {
			return nil, err
		}
		found = append(found, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	switch len(found) {
	case 0:
		return nil, fmt.Errorf("no work item matching %q", ref)
	case 1:
		return found[0], nil
	default:
		return nil, fmt.Errorf("ambiguous ref %q — matched %s… and %s…; use a longer prefix",
			ref, found[0].id[:8], found[1].id[:8])
	}
}

func printSnapshot(ctx context.Context, pool *pgxpool.Pool, s *wiSnapshot) {
	name := s.slug
	if name == "" {
		name = s.id[:8]
	}
	fmt.Printf("%s  [%s]\n", name, s.id)
	fmt.Printf("  project   %s\n", s.project)
	fmt.Printf("  pipeline  %s\n", s.pipeline)
	fmt.Printf("  stage     %s\n", s.stage)
	fmt.Printf("  status    %s   maturity %s   escalation %s\n", s.status, s.maturity, s.escalation)
	fmt.Printf("  cost      %s   tokens %s in / %s out\n",
		fmtMicro(s.micro), fmtTokens(s.tokensIn), fmtTokens(s.tokensOut))
	fmt.Printf("  created   %s   updated %s", relTime(s.createdAt), relTime(s.updatedAt))
	if s.completedAt != nil {
		fmt.Printf("   completed %s", relTime(*s.completedAt))
	}
	fmt.Println()
	if s.errText != "" {
		fmt.Printf("  error     %s\n", s.errText)
	}

	printRecentEvents(ctx, pool, s.id)
}

func printRecentEvents(ctx context.Context, pool *pgxpool.Pool, id string) {
	rows, err := pool.Query(ctx, `
		SELECT at, provider, model, input_tokens, output_tokens, micro_dollars, coalesce(notes,'')
		FROM stewards.cost_events
		WHERE work_item_id = $1
		ORDER BY at DESC
		LIMIT 8`, id)
	if err != nil {
		// non-fatal: the snapshot is still useful without the event tail
		fmt.Fprintf(os.Stderr, "  (cost events unavailable: %v)\n", err)
		return
	}
	defer rows.Close()

	var table [][]string
	for rows.Next() {
		var at time.Time
		var provider, model, notes string
		var in, out int
		var micro int64
		if err := rows.Scan(&at, &provider, &model, &in, &out, &micro, &notes); err != nil {
			fail("scan cost event", err)
		}
		if len(notes) > 40 {
			notes = notes[:39] + "…"
		}
		table = append(table, []string{
			relTime(at), provider + "/" + model,
			fmtTokens(int64(in)), fmtTokens(int64(out)), fmtMicro(micro), notes,
		})
	}
	if err := rows.Err(); err != nil && err != pgx.ErrNoRows {
		fmt.Fprintf(os.Stderr, "  (cost events: %v)\n", err)
	}
	if len(table) == 0 {
		fmt.Println("\n  no cost events yet")
		return
	}
	fmt.Println("\n  recent cost events:")
	printTable(
		[]string{"WHEN", "MODEL", "IN", "OUT", "COST", "NOTES"}, table,
		[]align{alignLeft, alignLeft, alignRight, alignRight, alignRight, alignLeft},
	)
}
