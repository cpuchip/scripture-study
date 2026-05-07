// Phase 3c.1 — pipeline + work_item CLI surface.
//
// Pipelines are read-only here (defined via SQL migration). work_items
// are created/inspected/transitioned through the functions in
// 3c1-pipelines-work-items.sql.
package show

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PipelineList prints all registered pipelines + stage counts.
func PipelineList(ctx context.Context, pool *pgxpool.Pool) error {
	rows, err := pool.Query(ctx, `
        SELECT family, coalesce(description, ''),
               jsonb_array_length(stages) AS n_stages,
               created_at, updated_at
          FROM stewards.pipelines
         ORDER BY family
    `)
	if err != nil {
		return err
	}
	defer rows.Close()

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "FAMILY\tSTAGES\tCREATED\tDESCRIPTION")
	count := 0
	for rows.Next() {
		var family, description string
		var nStages int
		var created, updated time.Time
		if err := rows.Scan(&family, &description, &nStages, &created, &updated); err != nil {
			return err
		}
		short := description
		if len(short) > 80 {
			short = short[:77] + "..."
		}
		fmt.Fprintf(tw, "%s\t%d\t%s\t%s\n",
			family, nStages, created.Format("2006-01-02"), short)
		count++
	}
	tw.Flush()
	fmt.Printf("\n%d pipeline(s)\n", count)
	return rows.Err()
}

// PipelineShow prints one pipeline's full stage list.
func PipelineShow(ctx context.Context, pool *pgxpool.Pool, family string) error {
	var description string
	var stagesJSON []byte
	var created, updated time.Time
	err := pool.QueryRow(ctx, `
        SELECT coalesce(description, ''), stages, created_at, updated_at
          FROM stewards.pipelines WHERE family = $1
    `, family).Scan(&description, &stagesJSON, &created, &updated)
	if err != nil {
		return fmt.Errorf("pipeline not found: %w", err)
	}

	fmt.Printf("\n## %s\n", family)
	if description != "" {
		fmt.Printf("%s\n\n", description)
	}
	fmt.Printf("created: %s\n", created.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("updated: %s\n\n", updated.Format("2006-01-02 15:04:05 MST"))

	var stages []map[string]any
	if err := json.Unmarshal(stagesJSON, &stages); err != nil {
		return fmt.Errorf("decode stages: %w", err)
	}
	fmt.Println("Stages:")
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "  #\tNAME\tAGENT\tMODEL\tPROVIDER\tNEXT\tAUTO_ADVANCE")
	for i, s := range stages {
		next, _ := s["next"].(string)
		if next == "" {
			next = "(terminal)"
		}
		auto := "true"
		if v, ok := s["auto_advance"].(bool); ok && !v {
			auto = "false"
		}
		fmt.Fprintf(tw, "  %d\t%s\t%s\t%s\t%s\t%s\t%s\n",
			i, s["name"], s["agent_family"], s["model"],
			s["provider"], next, auto)
	}
	tw.Flush()
	return nil
}

// WorkItemCreate calls stewards.work_item_create and prints the new id.
func WorkItemCreate(ctx context.Context, pool *pgxpool.Pool,
	pipeline, slug, actor string, input []byte, budget int,
) error {
	args := []any{pipeline, input, nullable(slug), actor}
	if budget > 0 {
		args = append(args, budget)
	} else {
		args = append(args, nil)
	}
	var id string
	if err := pool.QueryRow(ctx,
		`SELECT stewards.work_item_create($1, $2::jsonb, $3, $4, $5)::text`,
		args...,
	).Scan(&id); err != nil {
		return err
	}
	fmt.Printf("created work_item %s (pipeline %s)\n", id, pipeline)
	return printWorkItemDetail(ctx, pool, id)
}

// WorkItemList prints all work_items, optionally filtered.
func WorkItemList(ctx context.Context, pool *pgxpool.Pool,
	pipeline, status string,
) error {
	q := `
        SELECT id::text, coalesce(slug, ''), pipeline_family,
               current_stage, status, stages_completed, stages_total,
               tokens_in, tokens_out,
               coalesce(token_budget, 0) AS budget,
               created_at
          FROM stewards.work_items_summary
         WHERE ($1 = '' OR pipeline_family = $1)
           AND ($2 = '' OR status = $2)
         ORDER BY created_at DESC
    `
	rows, err := pool.Query(ctx, q, pipeline, status)
	if err != nil {
		return err
	}
	defer rows.Close()

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tSLUG\tPIPELINE\tSTAGE\tSTATUS\tDONE/TOTAL\tTOK_IN\tTOK_OUT\tBUDGET\tCREATED")
	count := 0
	for rows.Next() {
		var id, slug, pl, stage, st string
		var done, total, tIn, tOut, budget int
		var created time.Time
		if err := rows.Scan(&id, &slug, &pl, &stage, &st,
			&done, &total, &tIn, &tOut, &budget, &created); err != nil {
			return err
		}
		short := id
		if len(short) > 8 {
			short = short[:8]
		}
		display := short
		if slug != "" {
			display = slug + " (" + short + ")"
		}
		bDisp := "-"
		if budget > 0 {
			bDisp = fmt.Sprintf("%d", budget)
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%d/%d\t%d\t%d\t%s\t%s\n",
			display, slug, pl, stage, st,
			done, total, tIn, tOut, bDisp,
			created.Format("2006-01-02 15:04"))
		_ = display
		count++
	}
	tw.Flush()
	fmt.Printf("\n%d work_item(s)\n", count)
	return rows.Err()
}

// WorkItemShow prints one work_item with full detail.
func WorkItemShow(ctx context.Context, pool *pgxpool.Pool, ref string) error {
	id, err := resolveWorkItemRef(ctx, pool, ref)
	if err != nil {
		return err
	}
	return printWorkItemDetail(ctx, pool, id)
}

// WorkItemDispatch dispatches the current stage by enqueueing a chat.
func WorkItemDispatch(ctx context.Context, pool *pgxpool.Pool, ref, userInput string) error {
	id, err := resolveWorkItemRef(ctx, pool, ref)
	if err != nil {
		return err
	}
	var workQueueID int64
	if err := pool.QueryRow(ctx,
		`SELECT stewards.work_item_dispatch_stage($1::uuid, $2)`,
		id, nullable(userInput),
	).Scan(&workQueueID); err != nil {
		return err
	}
	fmt.Printf("dispatched work_item %s — work_queue id=%d (status=in_progress)\n", id, workQueueID)
	return nil
}

// WorkItemAdvance records stage output + transitions to the next stage
// (or marks completed if terminal).
func WorkItemAdvance(ctx context.Context, pool *pgxpool.Pool,
	ref string, output []byte,
) error {
	id, err := resolveWorkItemRef(ctx, pool, ref)
	if err != nil {
		return err
	}
	if len(output) == 0 {
		output = []byte(`{}`)
	}
	var nextStage *string
	if err := pool.QueryRow(ctx,
		`SELECT stewards.work_item_advance($1::uuid, $2::jsonb)`,
		id, output,
	).Scan(&nextStage); err != nil {
		return err
	}
	if nextStage == nil {
		fmt.Printf("work_item %s: completed (no further stages)\n", id)
	} else {
		fmt.Printf("work_item %s: advanced to stage %s\n", id, *nextStage)
	}
	return printWorkItemDetail(ctx, pool, id)
}

// WorkItemCancel marks a work_item cancelled.
func WorkItemCancel(ctx context.Context, pool *pgxpool.Pool, ref, reason string) error {
	id, err := resolveWorkItemRef(ctx, pool, ref)
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx,
		`SELECT stewards.work_item_cancel($1::uuid, $2)`,
		id, nullable(reason),
	); err != nil {
		return err
	}
	fmt.Printf("work_item %s: cancelled\n", id)
	return nil
}

// resolveWorkItemRef accepts either a uuid or a slug, returns the uuid.
func resolveWorkItemRef(ctx context.Context, pool *pgxpool.Pool, ref string) (string, error) {
	// Try uuid first.
	var id string
	err := pool.QueryRow(ctx,
		`SELECT id::text FROM stewards.work_items WHERE id::text = $1 OR slug = $1`,
		ref,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("work_item %q not found: %w", ref, err)
	}
	return id, nil
}

// printWorkItemDetail is the shared printer used by show/create/advance.
func printWorkItemDetail(ctx context.Context, pool *pgxpool.Pool, id string) error {
	var (
		slug, pipeline, currentStage, status, actor string
		input, stageResults                         []byte
		sessionIDs                                  []string
		tokensIn, tokensOut                         int
		tokenBudget                                 *int
		errStr                                      *string
		created, updated                            time.Time
		completed                                   *time.Time
	)
	err := pool.QueryRow(ctx, `
        SELECT coalesce(slug, ''), pipeline_family, current_stage, status,
               actor, input, stage_results, session_ids,
               tokens_in, tokens_out, token_budget, error,
               created_at, updated_at, completed_at
          FROM stewards.work_items WHERE id = $1
    `, id).Scan(&slug, &pipeline, &currentStage, &status,
		&actor, &input, &stageResults, &sessionIDs,
		&tokensIn, &tokensOut, &tokenBudget, &errStr,
		&created, &updated, &completed)
	if err != nil {
		return err
	}

	fmt.Printf("\n## work_item %s\n", id)
	if slug != "" {
		fmt.Printf("slug:           %s\n", slug)
	}
	fmt.Printf("pipeline:       %s\n", pipeline)
	fmt.Printf("current_stage:  %s\n", currentStage)
	fmt.Printf("status:         %s\n", status)
	fmt.Printf("actor:          %s\n", actor)
	fmt.Printf("created:        %s\n", created.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("updated:        %s\n", updated.Format("2006-01-02 15:04:05 MST"))
	if completed != nil {
		fmt.Printf("completed:      %s  (elapsed %s)\n",
			completed.Format("2006-01-02 15:04:05 MST"),
			completed.Sub(created).Round(time.Second))
	}
	budgetStr := "(none)"
	if tokenBudget != nil {
		budgetStr = fmt.Sprintf("%d", *tokenBudget)
	}
	fmt.Printf("tokens:         %d in / %d out  (budget %s)\n", tokensIn, tokensOut, budgetStr)
	if errStr != nil && *errStr != "" {
		fmt.Printf("error:          %s\n", *errStr)
	}
	if len(sessionIDs) > 0 {
		fmt.Printf("sessions:       %s\n", strings.Join(sessionIDs, ", "))
	}
	fmt.Printf("input:          %s\n", string(input))
	if len(stageResults) > 0 && string(stageResults) != "{}" {
		fmt.Println("\nstage_results:")
		var pretty map[string]any
		if err := json.Unmarshal(stageResults, &pretty); err == nil {
			for stage, result := range pretty {
				rb, _ := json.MarshalIndent(result, "  ", "  ")
				fmt.Printf("  %s: %s\n", stage, string(rb))
			}
		} else {
			fmt.Printf("  (raw) %s\n", string(stageResults))
		}
	}
	return nil
}
