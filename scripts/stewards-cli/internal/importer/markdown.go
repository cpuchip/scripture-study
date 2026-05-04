package importer

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	h1Re        = regexp.MustCompile(`(?m)^#\s+(.+)$`)
	italicMetaRe = regexp.MustCompile(`(?m)^\*([A-Za-z][A-Za-z _-]+):\s*(.+?)\*\s*$`)
	yamlFrontRe = regexp.MustCompile(`(?s)\A---\s*\n(.*?)\n---\s*\n`)
)

// readUTF8 reads the file as UTF-8 (no BOM). Mirrors the PowerShell
// importer's explicit UTF8Encoding(false) to keep em-dashes intact.
func readUTF8(absPath string) (string, error) {
	b, err := os.ReadFile(absPath)
	if err != nil {
		return "", err
	}
	// Strip UTF-8 BOM if present.
	if len(b) >= 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
		b = b[3:]
	}
	return string(b), nil
}

func slugFromPath(absPath string) string {
	base := filepath.Base(absPath)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// buildSlug produces a corpus-unique, human-readable slug:
//   - root-level study files keep their bare basename (so existing
//     similarity edges, citations, and memory references stay valid);
//   - everything else gets a kind prefix (or "" for study) plus the
//     joined subdir path.
//
// Examples:
//
//	study/charity.md                     → charity
//	study/talks/art-of-delegation.md     → talks-art-of-delegation
//	study/yt/foo-bar.md                  → yt-foo-bar
//	docs/work-with-ai/01_planning.md     → doc-01_planning
//	.spec/proposals/dev-loop.md          → proposal-dev-loop
//	.spec/journal/2026-05-04--x.yaml     → journal-2026-05-04--x
//
func buildSlug(absPath, sourceRoot, kind string) string {
	base := strings.TrimSuffix(filepath.Base(absPath), filepath.Ext(absPath))

	// Compute the subdir between sourceRoot and the file (excluding the
	// basename). Errors fall back to bare basename.
	var subParts []string
	if sourceRoot != "" {
		if rel, err := filepath.Rel(sourceRoot, filepath.Dir(absPath)); err == nil &&
			rel != "." && rel != "" {
			rel = filepath.ToSlash(rel)
			for _, p := range strings.Split(rel, "/") {
				if p != "" {
					subParts = append(subParts, p)
				}
			}
		}
	}

	parts := make([]string, 0, len(subParts)+2)
	// kind prefix: drop for 'study' so root-level study slugs stay bare
	// and match every existing reference in the corpus.
	if kind != "" && kind != "study" {
		parts = append(parts, kind)
	}
	parts = append(parts, subParts...)
	parts = append(parts, base)
	return strings.Join(parts, "-")
}

func extractTitle(body, fallback string) string {
	if m := h1Re.FindStringSubmatch(body); m != nil {
		return strings.TrimSpace(m[1])
	}
	return fallback
}

// parseMarkdownStudy: pre-2.5 import_study() shape. Title is first H1;
// frontmatter is collected from `*key: value*` italic lines in the
// first 20 lines (legacy convention from existing study/ files).
func parseMarkdownStudy(absPath, relPath, sourceRoot string) (*Doc, error) {
	body, err := readUTF8(absPath)
	if err != nil {
		return nil, err
	}
	slug := buildSlug(absPath, sourceRoot, "study")
	title := extractTitle(body, slug)
	fm := map[string]any{}
	headLines := strings.Split(body, "\n")
	if len(headLines) > 20 {
		headLines = headLines[:20]
	}
	for _, line := range headLines {
		if m := italicMetaRe.FindStringSubmatch(line); m != nil {
			key := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(m[1]), " ", "_"))
			fm[key] = strings.TrimSpace(m[2])
		}
	}
	return &Doc{Slug: slug, FilePath: relPath, Title: title, Body: body, Frontmatter: fm}, nil
}

// parseMarkdownDoc: docs/work-with-ai/ files. Title is first H1.
// Frontmatter is YAML if present, otherwise empty.
func parseMarkdownDoc(absPath, relPath, sourceRoot string) (*Doc, error) {
	body, err := readUTF8(absPath)
	if err != nil {
		return nil, err
	}
	slug := buildSlug(absPath, sourceRoot, "doc")
	body, fm := splitYAMLFrontmatter(body)
	title := extractTitle(body, slug)
	return &Doc{Slug: slug, FilePath: relPath, Title: title, Body: body, Frontmatter: fm}, nil
}

// parseMarkdownProposal: .spec/proposals/ files. Same shape as doc
// but YAML frontmatter is expected (workstream, status, created).
// We don't enforce it — a missing frontmatter just yields {}.
func parseMarkdownProposal(absPath, relPath, sourceRoot string) (*Doc, error) {
	body, err := readUTF8(absPath)
	if err != nil {
		return nil, err
	}
	slug := buildSlug(absPath, sourceRoot, "proposal")
	body, fm := splitYAMLFrontmatter(body)
	title := extractTitle(body, slug)
	return &Doc{Slug: slug, FilePath: relPath, Title: title, Body: body, Frontmatter: fm}, nil
}

// parseMarkdownPhaseDoc: a single (often very large) phases.md file.
// Currently treated as one document; Phase 2.6 will split per-phase.
// Slug uses the parent project directory name for uniqueness:
// projects/pg-ai-stewards/phases.md → "phase-doc-pg-ai-stewards-phases".
func parseMarkdownPhaseDoc(absPath, relPath, sourceRoot string) (*Doc, error) {
	body, err := readUTF8(absPath)
	if err != nil {
		return nil, err
	}
	parent := filepath.Base(filepath.Dir(absPath))
	slug := fmt.Sprintf("phase-doc-%s-phases", parent)
	body, fm := splitYAMLFrontmatter(body)
	title := extractTitle(body, fmt.Sprintf("%s — phases", parent))
	return &Doc{Slug: slug, FilePath: relPath, Title: title, Body: body, Frontmatter: fm}, nil
}

// splitYAMLFrontmatter peels a leading --- ... --- block off the
// document and parses it as YAML into a map. Returns the body without
// the frontmatter and the parsed map (empty if absent or malformed).
// Malformed frontmatter is logged but not fatal — we'd rather index
// the body than reject the file.
func splitYAMLFrontmatter(text string) (string, map[string]any) {
	fm := map[string]any{}
	m := yamlFrontRe.FindStringSubmatchIndex(text)
	if m == nil {
		return text, fm
	}
	yamlText := text[m[2]:m[3]]
	body := text[m[1]:]
	parsed := map[string]any{}
	if err := yaml.Unmarshal([]byte(yamlText), &parsed); err == nil {
		// Normalize values to JSON-friendly types so json.Marshal
		// doesn't choke on map[interface{}]interface{}.
		fm = normalizeYAML(parsed).(map[string]any)
	}
	return body, fm
}

// normalizeYAML walks YAML output and converts map[any]any to
// map[string]any so the result round-trips cleanly through json.Marshal.
// gopkg.in/yaml.v3 mostly produces map[string]any already, but nested
// values from arbitrary YAML can still surprise us.
func normalizeYAML(v any) any {
	switch x := v.(type) {
	case map[any]any:
		m := make(map[string]any, len(x))
		for k, val := range x {
			m[fmt.Sprint(k)] = normalizeYAML(val)
		}
		return m
	case map[string]any:
		for k, val := range x {
			x[k] = normalizeYAML(val)
		}
		return x
	case []any:
		for i, val := range x {
			x[i] = normalizeYAML(val)
		}
		return x
	default:
		return v
	}
}
