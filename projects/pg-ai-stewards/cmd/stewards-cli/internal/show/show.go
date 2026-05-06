// Package show wraps the SQL functions that render or list documents.
package show

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Study calls stewards.study_show and prints the returned text.
func Study(ctx context.Context, pool *pgxpool.Pool, slug string, sim, cites, verseChars int) error {
	var out string
	err := pool.QueryRow(ctx,
		`SELECT stewards.study_show($1, $2, $3, $4)`,
		slug, sim, cites, verseChars,
	).Scan(&out)
	if err != nil {
		return err
	}
	fmt.Print(out)
	return nil
}

// List prints every (slug, kind, title, embedded date), optionally
// filtered by kind. Tab-aligned for readability.
func List(ctx context.Context, pool *pgxpool.Pool, kind string) error {
	const baseQ = `
SELECT slug, kind, title,
       coalesce(to_char(embedded_at, 'YYYY-MM-DD'), '(unembedded)') AS embedded
  FROM stewards.studies
`
	q := baseQ
	args := []any{}
	if kind != "" {
		q += " WHERE kind = $1\n"
		args = append(args, kind)
	}
	q += " ORDER BY kind ASC, title ASC"

	rows, err := pool.Query(ctx, q, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "KIND\tSLUG\tEMBEDDED\tTITLE")
	count := 0
	for rows.Next() {
		var slug, k, title, embedded string
		if err := rows.Scan(&slug, &k, &title, &embedded); err != nil {
			return err
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", k, slug, embedded, title)
		count++
	}
	if err := rows.Err(); err != nil {
		return err
	}
	tw.Flush()
	fmt.Printf("\n%d row(s)\n", count)
	return nil
}

// Refresh re-resolves citations + similarity. With slug, scoped to
// that slug; without, corpus-wide.
func Refresh(ctx context.Context, pool *pgxpool.Pool, slug string) error {
	var resolves, sim int
	if slug == "" {
		err := pool.QueryRow(ctx,
			`SELECT stewards.refresh_all_study_refs(),
                    stewards.refresh_all_study_similarity()`,
		).Scan(&resolves, &sim)
		if err != nil {
			return err
		}
		fmt.Printf("corpus refresh: resolves_enqueued=%d  similarity_edges=%d\n", resolves, sim)
		return nil
	}
	err := pool.QueryRow(ctx,
		`SELECT stewards.refresh_study_refs($1),
                stewards.refresh_study_similarity($1)`,
		slug,
	).Scan(&resolves, &sim)
	if err != nil {
		return err
	}
	fmt.Printf("%s: resolves_enqueued=%d  similarity_edges=%d\n", slug, resolves, sim)
	return nil
}

// ============================================================
// Phase 2.6a — Workstream views
// ============================================================

// WorkstreamList prints all rows from stewards.workstreams.
func WorkstreamList(ctx context.Context, pool *pgxpool.Pool) error {
	rows, err := pool.Query(ctx, `
        SELECT w.id, w.name, w.status,
               coalesce(p.cnt, 0)::int AS proposal_count
          FROM stewards.workstreams w
          LEFT JOIN (
              -- Count via SQL view of the graph by joining frontmatter.
              -- Simpler: count studies whose frontmatter->>'workstream' = w.id.
              SELECT frontmatter->>'workstream' AS ws, COUNT(*) AS cnt
                FROM stewards.studies
               WHERE frontmatter ? 'workstream'
               GROUP BY frontmatter->>'workstream'
          ) p ON p.ws = w.id
         ORDER BY w.id
    `)
	if err != nil {
		return err
	}
	defer rows.Close()

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tNAME\tSTATUS\tPROPOSALS")
	for rows.Next() {
		var id, name, status string
		var count int
		if err := rows.Scan(&id, &name, &status, &count); err != nil {
			return err
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d\n", id, name, status, count)
	}
	tw.Flush()
	return rows.Err()
}

// WorkstreamShow prints one workstream + its declared proposals from
// the graph (via stewards.workstream_proposals).
func WorkstreamShow(ctx context.Context, pool *pgxpool.Pool, id string) error {
	var name, description, status string
	err := pool.QueryRow(ctx,
		`SELECT name, description, status FROM stewards.workstreams WHERE id = $1`,
		id,
	).Scan(&name, &description, &status)
	if err != nil {
		return fmt.Errorf("workstream %s not found: %w", id, err)
	}
	fmt.Printf("# %s — %s\n", id, name)
	fmt.Printf("Status: %s\n", status)
	if description != "" {
		fmt.Printf("\n%s\n", description)
	}
	fmt.Printf("\n## Declared proposals (from graph)\n\n")

	rows, err := pool.Query(ctx, `SELECT slug, kind, title, file_path FROM stewards.workstream_proposals($1)`, id)
	if err != nil {
		return err
	}
	defer rows.Close()
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "KIND\tSLUG\tTITLE")
	count := 0
	for rows.Next() {
		var slug, kind, title, file string
		if err := rows.Scan(&slug, &kind, &title, &file); err != nil {
			return err
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\n", kind, slug, title)
		count++
	}
	tw.Flush()
	fmt.Printf("\n%d proposal(s) declared in graph\n", count)
	return rows.Err()
}

// DeclaredEdges prints the outbound declared-provenance edges for a slug.
func DeclaredEdges(ctx context.Context, pool *pgxpool.Pool, slug string) error {
	rows, err := pool.Query(ctx, `SELECT from_slug, edge_type, to_slug, provenance, confidence, source FROM stewards.declared_edges($1)`, slug)
	if err != nil {
		return err
	}
	defer rows.Close()
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "TYPE\tTARGET\tPROVENANCE\tCONFIDENCE\tSOURCE")
	count := 0
	for rows.Next() {
		var from, etype, to, prov, src string
		var conf float64
		if err := rows.Scan(&from, &etype, &to, &prov, &conf, &src); err != nil {
			return err
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%.2f\t%s\n", etype, to, prov, conf, src)
		count++
	}
	tw.Flush()
	fmt.Printf("\n%d edge(s) from %s\n", count, slug)
	return rows.Err()
}

// ============================================================
// Phase 2.6b — Todos
// ============================================================

// TodoCreate calls stewards.create_todo and prints the new uuid.
func TodoCreate(ctx context.Context, pool *pgxpool.Pool, parentKind, parentSlug, title, body, slug, session string) error {
	var id string
	err := pool.QueryRow(ctx,
		`SELECT stewards.create_todo($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''))::text`,
		parentKind, parentSlug, title, body, slug, session,
	).Scan(&id)
	if err != nil {
		return err
	}
	fmt.Printf("created %s\n  parent: %s/%s\n  title:  %s\n", id, parentKind, parentSlug, title)
	return nil
}

// TodoComplete marks a todo done (or other terminal status).
func TodoComplete(ctx context.Context, pool *pgxpool.Pool, ref, session, status string) error {
	var id string
	err := pool.QueryRow(ctx,
		`SELECT stewards.complete_todo($1, NULLIF($2, ''), $3)::text`,
		ref, session, status,
	).Scan(&id)
	if err != nil {
		return err
	}
	fmt.Printf("%s -> %s (%s)\n", ref, status, id)
	return nil
}

// TodoList prints todos, optionally filtered by parent and status.
func TodoList(ctx context.Context, pool *pgxpool.Pool, parentKind, parentSlug, status string) error {
	rows, err := pool.Query(ctx,
		`SELECT id::text, coalesce(slug,''), title, status,
                coalesce(parent_kind,''), coalesce(parent_slug,''),
                to_char(created_at, 'YYYY-MM-DD HH24:MI'),
                coalesce(to_char(completed_at, 'YYYY-MM-DD HH24:MI'), '')
           FROM stewards.list_todos(NULLIF($1,''), NULLIF($2,''), NULLIF($3,''))`,
		parentKind, parentSlug, status,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "STATUS\tPARENT\tTITLE\tCREATED\tCOMPLETED\tID")
	count := 0
	for rows.Next() {
		var id, slug, title, st, pk, ps, created, completed string
		if err := rows.Scan(&id, &slug, &title, &st, &pk, &ps, &created, &completed); err != nil {
			return err
		}
		parent := pk + "/" + ps
		short := id
		if len(short) > 8 {
			short = short[:8]
		}
		display := title
		if slug != "" {
			display = slug + " — " + title
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n", st, parent, display, created, completed, short)
		count++
	}
	tw.Flush()
	fmt.Printf("\n%d todo(s)\n", count)
	return rows.Err()
}

// TodoAudit prints rows from stewards.todo_rollup_audit().
func TodoAudit(ctx context.Context, pool *pgxpool.Pool) error {
	rows, err := pool.Query(ctx, `SELECT finding, parent_kind, parent_slug,
        coalesce(parent_title,''), todo_count, open_count, done_count
        FROM stewards.todo_rollup_audit()`)
	if err != nil {
		return err
	}
	defer rows.Close()
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "FINDING\tPARENT\tTITLE\tTOTAL\tOPEN\tDONE")
	count := 0
	for rows.Next() {
		var finding, pk, ps, title string
		var total, open, done int
		if err := rows.Scan(&finding, &pk, &ps, &title, &total, &open, &done); err != nil {
			return err
		}
		fmt.Fprintf(tw, "%s\t%s/%s\t%s\t%d\t%d\t%d\n", finding, pk, ps, title, total, open, done)
		count++
	}
	tw.Flush()
	if count == 0 {
		fmt.Println("audit: clean (no findings)")
	} else {
		fmt.Printf("\n%d finding(s)\n", count)
	}
	return rows.Err()
}

// ============================================================
// Phase 2.6c — context_for graph walk
// ============================================================

// Context prints the typed graph neighborhood of a slug up to depth.
func Context(ctx context.Context, pool *pgxpool.Pool, slug string, depth int) error {
	rows, err := pool.Query(ctx,
		`SELECT hop, direction, edge_type, neighbor, neighbor_kind, provenance, confidence
           FROM stewards.context_for($1, $2)`,
		slug, depth,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "HOP\tDIR\tEDGE\tNEIGHBOR\tKIND\tPROV\tCONF")
	count := 0
	for rows.Next() {
		var hop int
		var dir, etype, neighbor, kind, prov string
		var conf float64
		if err := rows.Scan(&hop, &dir, &etype, &neighbor, &kind, &prov, &conf); err != nil {
			return err
		}
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\t%s\t%.2f\n", hop, dir, etype, neighbor, kind, prov, conf)
		count++
	}
	tw.Flush()
	fmt.Printf("\n%d neighbor(s) of %s within depth %d\n", count, slug, depth)
	return rows.Err()
}

// ============================================================
// Phase 2.7a — Watchman substrate
// ============================================================

// WatchmanQueue prints the dirty queue (oldest-touched first).
func WatchmanQueue(ctx context.Context, pool *pgxpool.Pool, limit int) error {
	rows, err := pool.Query(ctx,
		`SELECT slug, kind, title, updated_at, last_consolidated_at, dirty_for
           FROM stewards.dirty_queue
          LIMIT $1`,
		limit,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "SLUG\tKIND\tTOUCHED\tLAST_CONSOLIDATED\tDIRTY_FOR\tTITLE")
	count := 0
	for rows.Next() {
		var slug, kind, title string
		var touched time.Time
		var lastCons *time.Time
		var dirtyFor time.Duration
		if err := rows.Scan(&slug, &kind, &title, &touched, &lastCons, &dirtyFor); err != nil {
			return err
		}
		lastConsStr := "(never)"
		if lastCons != nil {
			lastConsStr = lastCons.Format("2006-01-02 15:04")
		}
		// Truncate title for table readability.
		if len(title) > 60 {
			title = title[:57] + "..."
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			slug, kind, touched.Format("2006-01-02 15:04"),
			lastConsStr, dirtyFor.Truncate(time.Minute), title)
		count++
	}
	tw.Flush()
	fmt.Printf("\n%d doc(s) in dirty queue (limit %d)\n", count, limit)
	return rows.Err()
}

// WatchmanVerdict records a verdict for a doc, bumping
// last_consolidated_at in the same transaction.
func WatchmanVerdict(ctx context.Context, pool *pgxpool.Pool,
	slug, verdict, reasoning, model, passID, actor string,
	tokensIn, tokensOut int) error {
	var id int64
	err := pool.QueryRow(ctx,
		`SELECT stewards.record_verdict($1, $2, $3, $4, $5, $6, $7, $8)`,
		slug, verdict, reasoning,
		nullable(model), tokensIn, tokensOut,
		nullable(passID), actor,
	).Scan(&id)
	if err != nil {
		return err
	}
	fmt.Printf("verdict %s recorded for %s (id=%d, dirty-bit reset)\n", verdict, slug, id)
	return nil
}

// WatchmanFinding writes a finding row.
func WatchmanFinding(ctx context.Context, pool *pgxpool.Pool,
	slug, kind, message, severity, suggestedAction, passID, actor string,
	related []string) error {
	var id int64
	err := pool.QueryRow(ctx,
		`SELECT stewards.record_finding($1, $2, $3, $4, $5, $6, $7, $8)`,
		slug, kind, message, severity,
		nullable(suggestedAction), related,
		nullable(passID), actor,
	).Scan(&id)
	if err != nil {
		return err
	}
	fmt.Printf("finding %s/%s recorded for %s (id=%d)\n", kind, severity, slug, id)
	return nil
}

// WatchmanAcknowledge marks a finding acknowledged.
func WatchmanAcknowledge(ctx context.Context, pool *pgxpool.Pool,
	id int64, resolution, actor string) error {
	if _, err := pool.Exec(ctx,
		`SELECT stewards.acknowledge_finding($1, $2, $3)`,
		id, resolution, actor,
	); err != nil {
		return err
	}
	fmt.Printf("finding %d acknowledged (%s)\n", id, resolution)
	return nil
}

// WatchmanHistory prints the verdict + finding timeline for a doc.
func WatchmanHistory(ctx context.Context, pool *pgxpool.Pool, slug string) error {
	rows, err := pool.Query(ctx,
		`SELECT event_at, event_type, detail, actor, extra
           FROM stewards.study_history($1)`,
		slug,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "WHEN\tTYPE\tACTOR\tDETAIL")
	count := 0
	for rows.Next() {
		var at time.Time
		var etype, detail, actor string
		var extra []byte
		if err := rows.Scan(&at, &etype, &detail, &actor, &extra); err != nil {
			return err
		}
		_ = extra // available for verbose mode later
		if len(detail) > 80 {
			detail = detail[:77] + "..."
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			at.Format("2006-01-02 15:04"), etype, actor, detail)
		count++
	}
	tw.Flush()
	fmt.Printf("\n%d event(s) for %s\n", count, slug)
	return rows.Err()
}

// nullable returns nil for empty strings so they go to SQL as NULL.
func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// =====================================================================
// Phase 3a — model-driven Watchman pass.
//
// Orchestrates: dirty_queue -> chat_enqueue -> poll -> parse JSON ->
// record_verdict (+ record_finding when verdict != clean).
//
// The bgworker stays generic. All watchman-specific semantics live
// here in the CLI orchestration. When 2.7b lands, this same logic
// moves into the bgworker as a scheduled pass.
// =====================================================================

// WatchmanPassResult is the parsed JSON shape the watchman-consolidator
// agent is asked to emit. See extension/3a-watchman-pass.sql for the
// authoritative schema description in the system prompt.
type WatchmanPassResult struct {
	Verdict   string                  `json:"verdict"`
	Reasoning string                  `json:"reasoning"`
	Finding   *WatchmanPassFinding    `json:"finding,omitempty"`
}

type WatchmanPassFinding struct {
	Kind             string `json:"kind"`
	Severity         string `json:"severity"`
	Message          string `json:"message"`
	SuggestedAction  string `json:"suggested_action"`
}

// WatchmanPass runs ONE consolidation pass over the dirty queue.
// Up to `limit` docs (oldest-touched first). Each doc gets one chat
// dispatch through the watchman-consolidator agent. The pass terminates
// cleanly when the queue is empty or the limit is hit.
//
// Provider/model defaults: opencode_go + kimi-k2.6 (the cheap, proven
// path from Phase 1.6). LM Studio + qwen3.6-27b is the local
// alternative. No remote API key required for the local path.
func WatchmanPass(ctx context.Context, pool *pgxpool.Pool,
	provider, model, agentFamily string,
	limit int, perItemTimeout time.Duration,
	dryRun bool, slugFilter string,
	maxInputChars int,
) error {
	passID := fmt.Sprintf("watchman-%s", time.Now().UTC().Format("20060102T150405Z"))

	// Build the queue. If a slug filter is given, run only that one
	// doc (useful for repro tests). Otherwise pull from dirty_queue.
	var slugs []string
	if slugFilter != "" {
		slugs = []string{slugFilter}
	} else {
		rows, err := pool.Query(ctx,
			`SELECT slug FROM stewards.dirty_queue ORDER BY updated_at ASC NULLS FIRST LIMIT $1`,
			limit,
		)
		if err != nil {
			return fmt.Errorf("dirty_queue: %w", err)
		}
		for rows.Next() {
			var s string
			if err := rows.Scan(&s); err != nil {
				rows.Close()
				return err
			}
			slugs = append(slugs, s)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}
	}

	if len(slugs) == 0 {
		fmt.Printf("watchman pass %s: dirty queue empty, nothing to do\n", passID)
		return nil
	}

	fmt.Printf("watchman pass %s: %d doc(s), provider=%s, model=%s\n",
		passID, len(slugs), provider, model)
	if dryRun {
		fmt.Println("  (dry-run: will print verdicts but NOT call record_verdict/record_finding)")
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "SLUG\tVERDICT\tELAPSED\tTOK_IN\tTOK_OUT\tFINDING?")

	var nClean, nDrift, nDone, nSuperseded, nSkipped, nErr int
	totalIn, totalOut := 0, 0

	for _, slug := range slugs {
		res, elapsed, tIn, tOut, err := watchmanPassOne(
			ctx, pool, slug, provider, model, agentFamily, passID, perItemTimeout, maxInputChars)
		if err != nil {
			fmt.Fprintf(tw, "%s\tERROR\t%s\t-\t-\t%v\n", slug, elapsed.Round(time.Millisecond), err)
			nErr++
			continue
		}
		totalIn += tIn
		totalOut += tOut
		findingStr := "no"
		if res.Finding != nil {
			findingStr = fmt.Sprintf("yes:%s/%s", res.Finding.Kind, res.Finding.Severity)
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%d\t%s\n",
			slug, res.Verdict, elapsed.Round(time.Millisecond), tIn, tOut, findingStr)

		switch res.Verdict {
		case "clean":
			nClean++
		case "drift":
			nDrift++
		case "done":
			nDone++
		case "superseded":
			nSuperseded++
		case "skipped":
			nSkipped++
		}

		if dryRun {
			continue
		}

		// Record verdict (advances last_consolidated_at as a side effect).
		// Signature: (slug, verdict, reasoning, model, tokens_in, tokens_out, pass_id, actor)
		if _, err := pool.Exec(ctx,
			`SELECT stewards.record_verdict($1, $2, $3, $4, $5, $6, $7, $8)`,
			slug, res.Verdict, res.Reasoning, model, tIn, tOut, passID, "watchman",
		); err != nil {
			fmt.Fprintf(os.Stderr, "  record_verdict(%s): %v\n", slug, err)
			continue
		}

		// Record finding when present.
		// Signature: (slug, kind, message, severity, suggested_action, related_slugs[], pass_id, actor)
		if res.Finding != nil {
			if _, err := pool.Exec(ctx,
				`SELECT stewards.record_finding($1, $2, $3, $4, $5, $6, $7, $8)`,
				slug,
				res.Finding.Kind,
				res.Finding.Message,
				res.Finding.Severity,
				res.Finding.SuggestedAction,
				[]string{},
				passID,
				"watchman",
			); err != nil {
				fmt.Fprintf(os.Stderr, "  record_finding(%s): %v\n", slug, err)
			}
		}
	}

	tw.Flush()
	fmt.Printf("\npass %s done: clean=%d drift=%d done=%d superseded=%d skipped=%d err=%d  tokens=%d in / %d out\n",
		passID, nClean, nDrift, nDone, nSuperseded, nSkipped, nErr, totalIn, totalOut)
	return nil
}

// watchmanPassOne runs the loop for a single doc.
//
// Steps:
//   1. Build session id (deterministic per pass+slug for replay).
//   2. INSERT INTO sessions.
//   3. Compose user input via stewards.watchman_input(slug).
//   4. chat_enqueue → returns work_queue.id.
//   5. Poll work_queue.id until status='done' or 'error'.
//   6. Read assistant message from messages WHERE session_id=... AND role='assistant'.
//   7. Parse JSON (defensive — strip ``` fences if present).
func watchmanPassOne(ctx context.Context, pool *pgxpool.Pool,
	slug, provider, model, agentFamily, passID string,
	timeout time.Duration, maxInputChars int,
) (WatchmanPassResult, time.Duration, int, int, error) {
	start := time.Now()
	var zero WatchmanPassResult

	sessionID := fmt.Sprintf("%s--%s", passID, slug)
	if len(sessionID) > 200 {
		sessionID = sessionID[:200]
	}

	// 1+2. Session.
	if _, err := pool.Exec(ctx,
		`INSERT INTO stewards.sessions (id, label, kind) VALUES ($1, $2, 'agent')
         ON CONFLICT (id) DO NOTHING`,
		sessionID, fmt.Sprintf("Watchman pass %s for %s", passID, slug),
	); err != nil {
		return zero, time.Since(start), 0, 0, fmt.Errorf("session insert: %w", err)
	}

	// 3. Compose input. Apply truncation AFTER the SQL composer so we
	// keep the doc-header + body + neighborhood structure intact and
	// only reach in to truncate the body if needed.
	var userInput string
	if err := pool.QueryRow(ctx,
		`SELECT stewards.watchman_input($1)`,
		slug,
	).Scan(&userInput); err != nil {
		return zero, time.Since(start), 0, 0, fmt.Errorf("watchman_input: %w", err)
	}
	if userInput == "" {
		return zero, time.Since(start), 0, 0, fmt.Errorf("study not found: %s", slug)
	}
	if maxInputChars > 0 && len(userInput) > maxInputChars {
		userInput = truncateMiddle(userInput, maxInputChars)
	}

	// 4. chat_enqueue.
	var workID int64
	if err := pool.QueryRow(ctx,
		`SELECT stewards.chat_enqueue($1, $2, $3, $4, $5)`,
		agentFamily, model, sessionID, userInput, provider,
	).Scan(&workID); err != nil {
		return zero, time.Since(start), 0, 0, fmt.Errorf("chat_enqueue: %w", err)
	}

	// 5. Poll. Polling is fine at this scale; LISTEN/NOTIFY can come
	// later if it becomes a bottleneck. 250ms tick keeps the local
	// latency low; total bounded by `timeout`.
	deadline := time.Now().Add(timeout)
	tick := time.NewTicker(250 * time.Millisecond)
	defer tick.Stop()
	var status, errStr string
	for {
		if err := pool.QueryRow(ctx,
			`SELECT status, COALESCE(error, '') FROM stewards.work_queue WHERE id = $1`,
			workID,
		).Scan(&status, &errStr); err != nil {
			return zero, time.Since(start), 0, 0, fmt.Errorf("poll: %w", err)
		}
		if status == "done" {
			break
		}
		if status == "error" {
			return zero, time.Since(start), 0, 0, fmt.Errorf("work_queue error: %s", errStr)
		}
		if time.Now().After(deadline) {
			return zero, time.Since(start), 0, 0, fmt.Errorf("timeout (status=%s)", status)
		}
		<-tick.C
	}

	// 6. Read assistant message. There may be multiple if the agent
	// mistakenly tried tools (we gave it none, but be defensive). Take
	// the most recent.
	var content string
	var tIn, tOut sql.NullInt32
	if err := pool.QueryRow(ctx,
		`SELECT content, tokens_in, tokens_out
           FROM stewards.messages
          WHERE session_id = $1 AND role = 'assistant'
          ORDER BY id DESC LIMIT 1`,
		sessionID,
	).Scan(&content, &tIn, &tOut); err != nil {
		return zero, time.Since(start), 0, 0, fmt.Errorf("read assistant msg: %w", err)
	}

	// 7. Parse JSON, defensively.
	parsed, err := parseWatchmanJSON(content)
	if err != nil {
		return zero, time.Since(start), int(tIn.Int32), int(tOut.Int32),
			fmt.Errorf("parse JSON: %w (raw: %s)", err, truncate(content, 200))
	}

	return parsed, time.Since(start), int(tIn.Int32), int(tOut.Int32), nil
}

// parseWatchmanJSON extracts the JSON object from the model's reply.
// Strips markdown fences if present (defensive — kimi/qwen sometimes
// wrap JSON in ```json ... ``` even when told not to).
func parseWatchmanJSON(s string) (WatchmanPassResult, error) {
	var zero WatchmanPassResult
	t := strings.TrimSpace(s)
	// Strip ```json or ``` fences.
	if strings.HasPrefix(t, "```") {
		// drop first line (the fence with optional language tag)
		if nl := strings.Index(t, "\n"); nl > 0 {
			t = t[nl+1:]
		}
		// drop trailing fence
		if i := strings.LastIndex(t, "```"); i >= 0 {
			t = t[:i]
		}
		t = strings.TrimSpace(t)
	}
	// Last-resort: find first { and last }.
	if i := strings.Index(t, "{"); i > 0 {
		t = t[i:]
	}
	if i := strings.LastIndex(t, "}"); i >= 0 && i < len(t)-1 {
		t = t[:i+1]
	}

	var r WatchmanPassResult
	if err := json.Unmarshal([]byte(t), &r); err != nil {
		return zero, err
	}
	if r.Verdict == "" {
		return zero, fmt.Errorf("missing verdict field")
	}
	return r, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// truncateMiddle keeps the head and tail of s and replaces the middle
// with an explicit elision marker. Used for big watchman inputs that
// would otherwise blow past the bgworker chat timeout. Head/tail split
// is 60/40 — the document header (slug, kind, title, updated_at) lives
// in the first ~200 chars and we prefer to lose body-middle over
// neighborhood-tail because the neighborhood lists are short and the
// 1-hop edges are often the strongest drift signal.
//
// The elision marker is human-readable AND tells the model what was
// dropped so it can render an honest 'skipped' verdict if the missing
// content matters.
func truncateMiddle(s string, max int) string {
	if len(s) <= max {
		return s
	}
	dropped := len(s) - max
	marker := fmt.Sprintf(
		"\n\n[... %d characters elided by stewards-cli --max-input-chars to fit bgworker chat timeout. If the missing middle is load-bearing, render verdict='skipped' with reasoning that calls out the truncation. ...]\n\n",
		dropped,
	)
	budget := max - len(marker)
	if budget < 200 {
		// Marker is bigger than budget; just hard-truncate.
		return s[:max]
	}
	head := budget * 6 / 10
	tail := budget - head
	return s[:head] + marker + s[len(s)-tail:]
}

