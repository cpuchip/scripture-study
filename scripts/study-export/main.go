// Command study-export pulls a study row from the pg-ai-stewards
// substrate and writes it as a polished /study/{slug}.md file.
//
// Substrate-produced studies (slug prefix `substrate--`) emit the
// agent's review-stage marker as the first line and use [slug](#)
// placeholders for cross-references — neither belongs in a published
// /study/ file. This tool fixes both:
//
//   - Strips the leading stage marker (REVIEW:, OUTLINE:, DRAFT:)
//   - Resolves [slug](#) → [slug](<workspace-path>) by scanning
//     /study/ recursively for matching .md files
//   - Resolves [Scripture Ref](#) → [Ref](../gospel-library/...)
//     using the convention from .github/skills/scripture-linking/
//   - Strips the `substrate--` prefix from the output slug
//
// Usage:
//   study-export <substrate-slug> [--out path] [--study-dir path] [--gl-dir path]
//
// Defaults: writes to ../../study/{stripped-slug}.md, scans
// ../../study/ recursively for slug→path index, scripture base is
// ../gospel-library/eng/scriptures/.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5"
)

func main() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("study-export: ")
	log.SetFlags(0)

	var (
		dsn      = flag.String("dsn", "", "Postgres DSN (default: $STEWARDS_DSN, then localhost compose port 55433)")
		outPath  = flag.String("out", "", "Output file path (default: study/{stripped-slug}.md relative to repo root)")
		studyDir = flag.String("study-dir", "", "Path to /study/ for slug→path index (default: auto-detect)")
		glPrefix = flag.String("gl-prefix", "../gospel-library/eng/scriptures", "Prefix for resolved scripture links (relative to output file)")
		dryRun   = flag.Bool("dry-run", false, "Print resolved body to stdout instead of writing a file")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: study-export <substrate-slug> [flags]\n")
		flag.PrintDefaults()
	}

	// Permissive parsing: pull the first non-flag token as the
	// substrate slug, then run flag.Parse on the rest. Go's stdlib
	// flag package stops at the first positional, which makes
	// `study-export <slug> --out <path>` silently drop the --out.
	rawArgs := os.Args[1:]
	var slug string
	var flagArgs []string
	for _, a := range rawArgs {
		if slug == "" && !strings.HasPrefix(a, "-") {
			slug = a
			continue
		}
		flagArgs = append(flagArgs, a)
	}
	if err := flag.CommandLine.Parse(flagArgs); err != nil {
		os.Exit(2)
	}
	if slug == "" {
		flag.Usage()
		os.Exit(2)
	}

	if *dsn == "" {
		*dsn = os.Getenv("STEWARDS_DSN")
	}
	if *dsn == "" {
		*dsn = "postgres://stewards:stewards@localhost:55433/stewards?sslmode=disable"
	}

	repoRoot := findRepoRoot()
	if *studyDir == "" {
		*studyDir = filepath.Join(repoRoot, "study")
	}
	outSlug := strings.TrimPrefix(slug, "substrate--")
	if *outPath == "" {
		*outPath = filepath.Join(*studyDir, outSlug+".md")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, *dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close(ctx)

	var body string
	err = conn.QueryRow(ctx,
		"SELECT body FROM stewards.studies WHERE slug = $1",
		slug,
	).Scan(&body)
	if err != nil {
		log.Fatalf("fetch %s: %v", slug, err)
	}
	log.Printf("read substrate study %s (%d chars)", slug, len(body))

	body = stripStageMarkers(body)

	idx, err := buildStudyIndex(*studyDir)
	if err != nil {
		log.Fatalf("build study index: %v", err)
	}
	log.Printf("indexed %d /study/ files for slug resolution", len(idx))

	body, slugRes := resolveSlugLinks(body, idx, *studyDir, *outPath)
	log.Printf("resolved %d slug links (%d unresolved)", slugRes.resolved, slugRes.unresolved)

	body, scriptRes := resolveScriptureLinks(body, *glPrefix)
	log.Printf("resolved %d scripture links (%d unresolved)", scriptRes.resolved, scriptRes.unresolved)

	// Ensure trailing newline for consistency with hand-edited files.
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}

	if *dryRun {
		fmt.Print(body)
		return
	}

	if err := os.WriteFile(*outPath, []byte(body), 0o644); err != nil {
		log.Fatalf("write %s: %v", *outPath, err)
	}
	log.Printf("wrote %s (%d chars)", *outPath, len(body))
}

// findRepoRoot walks up from the executable / cwd looking for a
// `go.work` file. Falls back to "." if not found.
func findRepoRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	dir := cwd
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return cwd
}
