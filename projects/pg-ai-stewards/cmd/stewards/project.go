package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// runProject lists projects, or switches/clears the sticky active project.
func runProject(ctx context.Context, args []string) {
	fs := flag.NewFlagSet("project", flag.ExitOnError)
	clear := fs.Bool("clear", false, "clear the active project")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	pool := mustConnect(ctx)
	defer pool.Close()

	if *clear {
		cfg := loadConfig()
		cfg.ActiveProject = ""
		if err := saveConfig(cfg); err != nil {
			fail("save config", err)
		}
		fmt.Println("active project cleared")
		return
	}

	if fs.NArg() >= 1 {
		slug := fs.Arg(0)
		var exists bool
		if err := pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM stewards.projects WHERE slug = $1)`, slug).Scan(&exists); err != nil {
			fail("look up project", err)
		}
		if !exists {
			fmt.Fprintf(os.Stderr, "no such project: %q (run 'stewards project' to list)\n", slug)
			os.Exit(1)
		}
		cfg := loadConfig()
		cfg.ActiveProject = slug
		if err := saveConfig(cfg); err != nil {
			fail("save config", err)
		}
		fmt.Printf("active project → %s\n", slug)
		return
	}

	listProjects(ctx, pool)
}

func listProjects(ctx context.Context, pool *pgxpool.Pool) {
	active := activeProject()
	rows, err := pool.Query(ctx, `
		SELECT p.slug, p.name, p.archived,
		       count(w.id) FILTER (WHERE w.status NOT IN ('completed','cancelled')) AS open_items,
		       count(w.id) AS total_items
		FROM stewards.projects p
		LEFT JOIN stewards.work_items w ON w.project_association = p.slug
		GROUP BY p.slug, p.name, p.archived
		ORDER BY p.archived, p.slug`)
	if err != nil {
		fail("query projects", err)
	}
	defer rows.Close()

	var table [][]string
	for rows.Next() {
		var slug, name string
		var archived bool
		var open, total int64
		if err := rows.Scan(&slug, &name, &archived, &open, &total); err != nil {
			fail("scan project", err)
		}
		marker := "  "
		if slug == active {
			marker = "* "
		}
		if archived {
			name += " (archived)"
		}
		table = append(table, []string{
			marker + slug, name,
			fmt.Sprintf("%d", open), fmt.Sprintf("%d", total),
		})
	}
	if err := rows.Err(); err != nil {
		fail("iterate projects", err)
	}
	if len(table) == 0 {
		fmt.Println("no projects yet")
		return
	}
	printTable(
		[]string{"PROJECT", "NAME", "OPEN", "TOTAL"}, table,
		[]align{alignLeft, alignLeft, alignRight, alignRight},
	)
	if active == "" {
		fmt.Println("\n(no active project — 'stewards project <slug>' to pick one)")
	}
}
