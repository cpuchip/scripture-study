package importer

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/yaml.v3"
)

// SkillDoc is the parsed shape of one .github/skills/<name>/SKILL.md.
type SkillDoc struct {
	Family       string // skill name
	Description  string
	Body         string // SKILL.md body (becomes skills.body)
	UserInvokable bool  // skip-by-default for non-user-invokable; informational
}

type skillFrontmatter struct {
	Name          string `yaml:"name"`
	Description   string `yaml:"description"`
	UserInvokable bool   `yaml:"user-invokable"`
}

// parseSkillMarkdown reads one SKILL.md and returns a SkillDoc.
// The skill's family is derived from the parent directory name to
// preserve `.github/skills/<dir>/SKILL.md → family=<dir>` even when
// the frontmatter `name:` field disagrees (warning logged in that case).
func parseSkillMarkdown(absPath string) (*SkillDoc, error) {
	raw, err := readUTF8(absPath)
	if err != nil {
		return nil, err
	}

	m := yamlFrontRe.FindStringSubmatchIndex(raw)
	if m == nil {
		return nil, fmt.Errorf("no YAML frontmatter found")
	}
	yamlBlock := raw[m[2]:m[3]]
	body := raw[m[1]:]

	var fm skillFrontmatter
	if err := yaml.Unmarshal([]byte(yamlBlock), &fm); err != nil {
		return nil, fmt.Errorf("yaml parse: %w", err)
	}

	// Family from parent directory (e.g., source-verification/SKILL.md
	// → source-verification). Falls back to frontmatter name if the
	// file isn't in a directory matching the convention.
	dirName := filepath.Base(filepath.Dir(absPath))
	family := strings.TrimSpace(fm.Name)
	if dirName != "" && dirName != "." && dirName != "skills" {
		if family != "" && family != dirName {
			fmt.Fprintf(os.Stderr,
				"  WARN skill %s: dir name (%s) and frontmatter name (%s) disagree; using dir\n",
				absPath, dirName, family)
		}
		family = dirName
	}
	if family == "" {
		return nil, fmt.Errorf("no family (no parent dir + no frontmatter name)")
	}

	return &SkillDoc{
		Family:        family,
		Description:   strings.TrimSpace(fm.Description),
		Body:          strings.TrimSpace(body),
		UserInvokable: fm.UserInvokable,
	}, nil
}

// upsertSkill inserts/updates a skill. The substrate's skills table
// has CHECK constraints on family format and description length;
// callers should pass through any check-violation errors so the user
// sees what's wrong (e.g., a skill family with uppercase letters).
func upsertSkill(ctx context.Context, pool *pgxpool.Pool, s *SkillDoc) error {
	// Description must be 1-1024 chars per the Phase 1.5 schema check.
	desc := s.Description
	if desc == "" {
		desc = "(no description provided)"
	}
	if len(desc) > 1024 {
		desc = desc[:1021] + "..."
	}

	if _, err := pool.Exec(ctx,
		`INSERT INTO stewards.skills
		    (family, model_match, description, body)
		 VALUES ($1, '*', $2, $3)
		 ON CONFLICT (family, model_match) DO UPDATE
		 SET description = EXCLUDED.description,
		     body        = EXCLUDED.body`,
		s.Family, desc, s.Body,
	); err != nil {
		return fmt.Errorf("upsert skill %s: %w", s.Family, err)
	}
	return nil
}

// ImportSkills walks a directory of <name>/SKILL.md files, parses
// each, and upserts into stewards.skills. Returns (ok, fail) counts.
func ImportSkills(ctx context.Context, pool *pgxpool.Pool,
	src Source, limit int, verbose bool,
) (int, int) {
	absRoot, err := filepath.Abs(src.Path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "skills: resolve %s: %v\n", src.Path, err)
		return 0, 1
	}
	info, err := os.Stat(absRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "skills: stat %s: %v\n", absRoot, err)
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
			// We want SKILL.md specifically (case-insensitive). Other
			// files in skill dirs (templates, helpers) are out of scope.
			if strings.EqualFold(filepath.Base(p), "SKILL.md") {
				files = append(files, p)
			}
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "skills: walk %s: %v\n", absRoot, err)
			return 0, 1
		}
	}

	if limit > 0 && len(files) > limit {
		files = files[:limit]
	}

	ok, fail := 0, 0
	for _, abs := range files {
		s, err := parseSkillMarkdown(abs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  PARSE FAIL: %s: %v\n",
				filepath.Base(filepath.Dir(abs)), err)
			fail++
			continue
		}
		if err := upsertSkill(ctx, pool, s); err != nil {
			fmt.Fprintf(os.Stderr, "  IMPORT FAIL: %s: %v\n", s.Family, err)
			fail++
			continue
		}
		if verbose {
			fmt.Printf("  ok: skill %s\n", s.Family)
		}
		ok++
	}
	return ok, fail
}
