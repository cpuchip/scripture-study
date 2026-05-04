// Package importer parses heterogeneous documents into the
// (slug, file_path, title, body, frontmatter, kind) shape that
// stewards.import_study() consumes.
//
// Each kind has a different "natural shape" (markdown w/ italic
// metadata, markdown w/ YAML frontmatter, structured YAML). Parsers
// live in per-kind files; the dispatcher in importer.go picks one
// based on Source.Kind.
package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Source is a (kind, path) pair from the --source flag. Path may be
// a directory (recursive scan for matching files) or a single file.
type Source struct {
	Kind string
	Path string
}

// Doc is the parsed shape ready for stewards.import_study().
type Doc struct {
	Slug        string
	FilePath    string // workspace-relative
	Title       string
	Body        string
	Frontmatter map[string]any
	Kind        string
}

// Parser turns one file into one Doc. The sourceRoot is the absolute
// path of the --source root, used to compute the file's subpath
// (everything between sourceRoot and the basename) so the slug can be
// disambiguated when basenames collide across subdirs. (Studies were
// silently overwriting each other in the old PowerShell importer
// when `study/x.md` and `study/talks/x.md` existed in parallel.)
type Parser func(absPath, relPath, sourceRoot string) (*Doc, error)

func parserFor(kind string) (Parser, string, error) {
	switch kind {
	case "study":
		return parseMarkdownStudy, ".md", nil
	case "doc":
		return parseMarkdownDoc, ".md", nil
	case "proposal":
		return parseMarkdownProposal, ".md", nil
	case "phase-doc":
		return parseMarkdownPhaseDoc, ".md", nil
	case "journal":
		return parseJournalYAML, ".yaml", nil
	default:
		return nil, "", fmt.Errorf("unknown kind: %s (want study|doc|proposal|phase-doc|journal)", kind)
	}
}

// ImportSource walks src.Path (file or dir), parses each matching
// file, and inserts via stewards.import_study(). Returns (ok, fail).
func ImportSource(ctx context.Context, pool *pgxpool.Pool, src Source, limit int, verbose bool) (int, int) {
	parser, ext, err := parserFor(src.Kind)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", src.Kind, err)
		return 0, 1
	}

	// Resolve the source root relative to the workspace root so the
	// stored file_path is portable. We assume the CLI is invoked from
	// the workspace root; if not, the user's --source path is taken
	// as-is and we just store it.
	absRoot, err := filepath.Abs(src.Path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: resolve %s: %v\n", src.Kind, src.Path, err)
		return 0, 1
	}
	workspaceRoot, _ := os.Getwd()

	// Single-file source (e.g. phases.md).
	info, err := os.Stat(absRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: stat %s: %v\n", src.Kind, absRoot, err)
		return 0, 1
	}

	var files []string
	if !info.IsDir() {
		files = []string{absRoot}
	} else {
		err = filepath.WalkDir(absRoot, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if strings.EqualFold(filepath.Ext(p), ext) {
				files = append(files, p)
			}
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: walk %s: %v\n", src.Kind, absRoot, err)
			return 0, 1
		}
	}

	if limit > 0 && len(files) > limit {
		files = files[:limit]
	}

	ok, fail := 0, 0
	for _, abs := range files {
		rel, err := filepath.Rel(workspaceRoot, abs)
		if err != nil {
			rel = abs
		}
		// Normalize to forward slashes for cross-platform consistency
		// of stored file_path values.
		rel = filepath.ToSlash(rel)

		doc, err := parser(abs, rel, absRoot)
		if err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "  PARSE FAIL: %s: %v\n", rel, err)
			}
			fail++
			continue
		}
		doc.Kind = src.Kind

		if err := upsert(ctx, pool, doc); err != nil {
			fmt.Fprintf(os.Stderr, "  IMPORT FAIL: %s: %v\n", rel, err)
			fail++
			continue
		}
		if verbose {
			fmt.Printf("  ok: %s (%s)\n", doc.Slug, rel)
		}
		ok++
	}
	return ok, fail
}

// upsert calls stewards.import_study via pgx — fully parameterized,
// no SQL string building, no apostrophe-escape issues.
func upsert(ctx context.Context, pool *pgxpool.Pool, doc *Doc) error {
	fmJSON, err := json.Marshal(doc.Frontmatter)
	if err != nil {
		return fmt.Errorf("frontmatter marshal: %w", err)
	}
	if len(fmJSON) == 0 || string(fmJSON) == "null" {
		fmJSON = []byte("{}")
	}
	_, err = pool.Exec(ctx,
		`SELECT stewards.import_study($1, $2, $3, $4, $5::jsonb, $6)`,
		doc.Slug, doc.FilePath, doc.Title, doc.Body, string(fmJSON), doc.Kind,
	)
	return err
}
