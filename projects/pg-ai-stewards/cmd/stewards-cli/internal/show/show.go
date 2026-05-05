// Package show wraps the SQL functions that render or list documents.
package show

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

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
