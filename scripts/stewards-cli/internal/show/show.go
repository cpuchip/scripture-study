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
