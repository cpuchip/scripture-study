// Resolution logic: stage-marker stripping, slug-link resolution,
// scripture-link resolution.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// stripStageMarkers removes leading agent stage markers like
//   REVIEW: revised
//
//   # Title...
// produced by the substrate's study-write pipeline. Pattern: a line
// matching ^(REVIEW|OUTLINE|DRAFT)(:.*)?$ followed by a blank line.
// Idempotent — body without a marker passes through unchanged.
func stripStageMarkers(body string) string {
	re := regexp.MustCompile(`(?m)^(REVIEW|OUTLINE|DRAFT)(:.*)?\n\n`)
	return re.ReplaceAllString(body, "")
}

// buildStudyIndex walks studyDir recursively and returns a map of
// stem (filename without .md) → relative path within studyDir.
// Multiple matches for the same stem keep the first one found
// (sorted alphabetically by full path) — collisions are rare.
func buildStudyIndex(studyDir string) (map[string]string, error) {
	idx := map[string]string{}
	err := filepath.Walk(studyDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip on permission etc.
		}
		if info.IsDir() {
			// Skip hidden / scratch directories.
			name := info.Name()
			if name == ".scratch" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}
		rel, err := filepath.Rel(studyDir, path)
		if err != nil {
			return nil
		}
		stem := strings.TrimSuffix(filepath.Base(path), ".md")
		// Hyphen-flatten variant: nested files like
		// plan-of-salvation/notes-03-spirit-world.md have a flat
		// alias under the agent's slug convention.
		flatKey := strings.ReplaceAll(strings.TrimSuffix(rel, ".md"), string(filepath.Separator), "-")
		flatKey = strings.ReplaceAll(flatKey, "/", "-")
		// Stem alone (one entry per file, may collide with another
		// of the same name in a different dir — first-wins).
		if _, ok := idx[stem]; !ok {
			idx[stem] = rel
		}
		// Flat-key alias only if it differs from stem.
		if flatKey != stem {
			if _, ok := idx[flatKey]; !ok {
				idx[flatKey] = rel
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return idx, nil
}

type resolveCounts struct {
	resolved   int
	unresolved int
}

// resolveSlugLinks rewrites [slug](#) → [slug](rel-path) using the
// index. Skips links whose text doesn't look like a slug (anything
// containing space, capital letter, or non-ASCII). The relative path
// is computed from the output file's directory.
//
// Pattern matched: [slug](#) where slug is [a-z0-9_-]+
func resolveSlugLinks(body string, idx map[string]string, studyDir, outPath string) (string, resolveCounts) {
	outDir := filepath.Dir(outPath)
	re := regexp.MustCompile(`\[([a-z0-9][a-z0-9_-]*)\]\(#\)`)
	c := resolveCounts{}
	out := re.ReplaceAllStringFunc(body, func(match string) string {
		sub := re.FindStringSubmatch(match)
		slug := sub[1]
		target, ok := idx[slug]
		if !ok {
			c.unresolved++
			return match
		}
		// Build relative path from outDir to studyDir/target.
		absTarget := filepath.Join(studyDir, target)
		rel, err := filepath.Rel(outDir, absTarget)
		if err != nil {
			c.unresolved++
			return match
		}
		// Normalize separators to forward slashes for markdown.
		rel = filepath.ToSlash(rel)
		c.resolved++
		return fmt.Sprintf("[%s](%s)", slug, rel)
	})
	return out, c
}

// resolveScriptureLinks rewrites [Scripture Ref](#) → [Ref](path).
// Uses bookMap to translate book names to gospel-library paths.
// Format: "<Book> <Chapter>:<Verse[-Verse]>" or "<Book> <Chapter>".
func resolveScriptureLinks(body, glPrefix string) (string, resolveCounts) {
	// Pattern: [Book Chapter[:Verse[-Verse]]](#)
	// Books may include numerical prefixes ("1 Nephi"), ampersands
	// ("D&C"), and em-dashes/hyphens. Keep the regex permissive then
	// validate with bookMap.
	re := regexp.MustCompile(`\[([0-9A-Z][^]]{0,40})\]\(#\)`)
	c := resolveCounts{}
	out := re.ReplaceAllStringFunc(body, func(match string) string {
		sub := re.FindStringSubmatch(match)
		ref := sub[1]
		path, ok := scriptureRefToPath(ref, glPrefix)
		if !ok {
			c.unresolved++
			return match
		}
		c.resolved++
		return fmt.Sprintf("[%s](%s)", ref, path)
	})
	return out, c
}

// scriptureRefToPath turns "Alma 40:3" → "<glPrefix>/bofm/alma/40.md".
// Returns ok=false for unrecognized book names — caller leaves the
// link unchanged in that case.
func scriptureRefToPath(ref, glPrefix string) (string, bool) {
	// Normalize: collapse whitespace, strip surrounding spaces.
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", false
	}

	// Find the chapter:verse split. Book = everything before the
	// last digit-led token. Chapter = the digit-led token (before
	// the colon if present).
	parts := strings.Fields(ref)
	if len(parts) < 2 {
		return "", false
	}
	chapVerse := parts[len(parts)-1]
	bookName := strings.Join(parts[:len(parts)-1], " ")

	chapter := chapVerse
	if i := strings.IndexByte(chapVerse, ':'); i >= 0 {
		chapter = chapVerse[:i]
	}
	if chapter == "" {
		return "", false
	}
	// Validate chapter is numeric.
	for _, r := range chapter {
		if r < '0' || r > '9' {
			return "", false
		}
	}

	loc, ok := bookMap[strings.ToLower(bookName)]
	if !ok {
		return "", false
	}
	return fmt.Sprintf("%s/%s/%s.md", glPrefix, loc, chapter), true
}

// bookMap is the lowercase book name → "<volume>/<book-dir>" lookup.
// Built from .github/skills/scripture-linking/SKILL.md conventions
// and the structure of /gospel-library/eng/scriptures/. Aliases
// include common abbreviations (D&C, JS-H, etc).
var bookMap = map[string]string{
	// Old Testament
	"genesis":       "ot/gen",
	"gen":           "ot/gen",
	"exodus":        "ot/ex",
	"ex":            "ot/ex",
	"leviticus":     "ot/lev",
	"lev":           "ot/lev",
	"numbers":       "ot/num",
	"num":           "ot/num",
	"deuteronomy":   "ot/deut",
	"deut":          "ot/deut",
	"joshua":        "ot/josh",
	"josh":          "ot/josh",
	"judges":        "ot/judg",
	"judg":          "ot/judg",
	"ruth":          "ot/ruth",
	"1 samuel":      "ot/1-sam",
	"2 samuel":      "ot/2-sam",
	"1 kings":       "ot/1-kgs",
	"2 kings":       "ot/2-kgs",
	"1 chronicles":  "ot/1-chr",
	"2 chronicles":  "ot/2-chr",
	"ezra":          "ot/ezra",
	"nehemiah":      "ot/neh",
	"neh":           "ot/neh",
	"esther":        "ot/esth",
	"job":           "ot/job",
	"psalm":         "ot/ps",
	"psalms":        "ot/ps",
	"ps":            "ot/ps",
	"proverbs":      "ot/prov",
	"prov":          "ot/prov",
	"ecclesiastes":  "ot/eccl",
	"eccl":          "ot/eccl",
	"isaiah":        "ot/isa",
	"isa":           "ot/isa",
	"jeremiah":      "ot/jer",
	"jer":           "ot/jer",
	"ezekiel":       "ot/ezek",
	"ezek":          "ot/ezek",
	"daniel":        "ot/dan",
	"dan":           "ot/dan",
	"hosea":         "ot/hosea",
	"joel":          "ot/joel",
	"amos":          "ot/amos",
	"obadiah":       "ot/obad",
	"jonah":         "ot/jonah",
	"micah":         "ot/micah",
	"nahum":         "ot/nahum",
	"habakkuk":      "ot/hab",
	"zephaniah":     "ot/zeph",
	"haggai":        "ot/hag",
	"zechariah":     "ot/zech",
	"malachi":       "ot/mal",
	"mal":           "ot/mal",
	// New Testament
	"matthew":       "nt/matt",
	"matt":          "nt/matt",
	"mark":          "nt/mark",
	"luke":          "nt/luke",
	"john":          "nt/john",
	"acts":          "nt/acts",
	"romans":        "nt/rom",
	"rom":           "nt/rom",
	"1 corinthians": "nt/1-cor",
	"2 corinthians": "nt/2-cor",
	"galatians":     "nt/gal",
	"gal":           "nt/gal",
	"ephesians":     "nt/eph",
	"eph":           "nt/eph",
	"philippians":   "nt/philip",
	"philip":        "nt/philip",
	"colossians":    "nt/col",
	"col":           "nt/col",
	"1 thessalonians": "nt/1-thes",
	"2 thessalonians": "nt/2-thes",
	"1 timothy":     "nt/1-tim",
	"2 timothy":     "nt/2-tim",
	"titus":         "nt/titus",
	"philemon":      "nt/philem",
	"hebrews":       "nt/heb",
	"heb":           "nt/heb",
	"james":         "nt/james",
	"1 peter":       "nt/1-pet",
	"2 peter":       "nt/2-pet",
	"1 john":        "nt/1-jn",
	"2 john":        "nt/2-jn",
	"3 john":        "nt/3-jn",
	"jude":          "nt/jude",
	"revelation":    "nt/rev",
	"rev":           "nt/rev",
	// Book of Mormon
	"1 nephi":       "bofm/1-ne",
	"1 ne":          "bofm/1-ne",
	"2 nephi":       "bofm/2-ne",
	"2 ne":          "bofm/2-ne",
	"jacob":         "bofm/jacob",
	"enos":          "bofm/enos",
	"jarom":         "bofm/jarom",
	"omni":          "bofm/omni",
	"words of mormon": "bofm/w-of-m",
	"w of m":        "bofm/w-of-m",
	"mosiah":        "bofm/mosiah",
	"alma":          "bofm/alma",
	"helaman":       "bofm/hel",
	"hel":           "bofm/hel",
	"3 nephi":       "bofm/3-ne",
	"3 ne":          "bofm/3-ne",
	"4 nephi":       "bofm/4-ne",
	"4 ne":          "bofm/4-ne",
	"mormon":        "bofm/morm",
	"morm":          "bofm/morm",
	"ether":         "bofm/ether",
	"moroni":        "bofm/moro",
	"moro":          "bofm/moro",
	// Doctrine and Covenants
	"d&c":                       "dc-testament/dc",
	"dc":                        "dc-testament/dc",
	"doctrine and covenants":    "dc-testament/dc",
	// Pearl of Great Price
	"moses":                "pgp/moses",
	"abraham":              "pgp/abr",
	"abr":                  "pgp/abr",
	"joseph smith—matthew": "pgp/js-m",
	"js-m":                 "pgp/js-m",
	"joseph smith—history": "pgp/js-h",
	"js-h":                 "pgp/js-h",
	"articles of faith":    "pgp/a-of-f",
	"a of f":               "pgp/a-of-f",
}
