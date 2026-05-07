package importer

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Journal entries evolved over time:
//   - early entries: `session` + `title`, discoveries as []string,
//     `relational_dynamics` instead of `relationship`,
//     carry_forward items have `topic` field
//   - later entries: `session_id`, discoveries as [{title, detail}],
//     `relationship` as [{quality, detail}], no `topic` on carry_forward
//
// We parse into a fully-loose map[string]any then synthesize the body
// by inspecting whatever shape we got. This is the parser pain Michael
// asked us to surface — and absorbing it here means the substrate
// accepts every entry rather than half of them.
//
// The whole raw structure is preserved as the frontmatter so Phase
// 2.6 typed edges can still walk it.

func parseJournalYAML(absPath, relPath, sourceRoot string) (*Doc, error) {
	raw, err := readUTF8(absPath)
	if err != nil {
		return nil, err
	}
	var entry map[string]any
	if err := yaml.Unmarshal([]byte(raw), &entry); err != nil {
		// Resilience: malformed YAML still gets indexed as raw text
		// so no journal entry is ever silently lost. The frontmatter
		// flags it so a future cleanup pass can find these.
		slug := buildSlug(absPath, sourceRoot, "journal")
		return &Doc{
			Slug:        slug,
			FilePath:    relPath,
			Title:       slug + " (raw — yaml parse failed)",
			Body:        raw,
			Frontmatter: map[string]any{"parse_error": err.Error()},
		}, nil
	}
	entry = normalizeYAML(entry).(map[string]any)

	slug := buildSlug(absPath, sourceRoot, "journal")

	// Title preference: explicit `title` > session_id > session > truncated intent > slug.
	title := firstString(entry, "title")
	if title == "" {
		title = firstString(entry, "session_id", "session")
	}
	if title == "" {
		title = truncate(strings.TrimSpace(firstString(entry, "intent")), 80)
	}
	if title == "" {
		title = slug
	}
	if d := firstString(entry, "date"); d != "" {
		title = fmt.Sprintf("%s — %s", d, title)
	}

	var b strings.Builder

	if intent := strings.TrimSpace(firstString(entry, "intent")); intent != "" {
		b.WriteString("## Intent\n\n")
		b.WriteString(intent)
		b.WriteString("\n\n")
	}

	writeSection(&b, "Discoveries", entry["discoveries"], "title")
	writeSection(&b, "Surprises", entry["surprises"], "")
	writeSection(&b, "Relationship",
		firstNonNil(entry, "relationship", "relational_dynamics", "relational"),
		"quality")
	writeCarry(&b, entry["carry_forward"])
	writeSection(&b, "Open questions", entry["questions"], "")

	body := b.String()
	if strings.TrimSpace(body) == "" {
		body = raw
	}

	fm := map[string]any{
		"date":       firstString(entry, "date"),
		"session_id": firstString(entry, "session_id", "session"),
		"tags":       entry["tags"],
		// Phase 2.7b.4: pass through `watchman` so the dirty_queue
		// exemption (e.g., `watchman: skip` to keep journals out of
		// Watchman passes) actually reaches the substrate.
		"watchman":   firstString(entry, "watchman"),
	}
	for k, v := range fm {
		if v == nil || v == "" {
			delete(fm, k)
		}
	}

	return &Doc{
		Slug:        slug,
		FilePath:    relPath,
		Title:       title,
		Body:        body,
		Frontmatter: fm,
	}, nil
}

// writeSection accepts a value that may be:
//   - nil or empty
//   - []any of strings (just bullet them)
//   - []any of map[string]any (use namedField for the heading,
//     "detail" / "note" for the body)
func writeSection(b *strings.Builder, heading string, v any, namedField string) {
	items := toSlice(v)
	if len(items) == 0 {
		return
	}
	fmt.Fprintf(b, "## %s\n\n", heading)
	for _, it := range items {
		switch x := it.(type) {
		case string:
			fmt.Fprintf(b, "- %s\n", strings.TrimSpace(x))
		case map[string]any:
			head := ""
			if namedField != "" {
				head = firstString(x, namedField, "title", "name")
			}
			detail := strings.TrimSpace(firstString(x, "detail", "note", "description"))
			if head != "" {
				fmt.Fprintf(b, "### %s\n\n", head)
				if detail != "" {
					b.WriteString(detail)
					b.WriteString("\n\n")
				}
			} else if detail != "" {
				fmt.Fprintf(b, "- %s\n", detail)
			} else {
				fmt.Fprintf(b, "- %v\n", x)
			}
		default:
			fmt.Fprintf(b, "- %v\n", x)
		}
	}
	b.WriteString("\n")
}

// writeCarry handles carry_forward, which has its own consistent shape
// across versions: list of {priority, note, [topic]}.
func writeCarry(b *strings.Builder, v any) {
	items := toSlice(v)
	if len(items) == 0 {
		return
	}
	b.WriteString("## Carry forward\n\n")
	for _, it := range items {
		m, ok := it.(map[string]any)
		if !ok {
			fmt.Fprintf(b, "- %v\n", it)
			continue
		}
		pri := firstString(m, "priority")
		topic := firstString(m, "topic")
		note := firstString(m, "note")
		switch {
		case topic != "" && note != "":
			fmt.Fprintf(b, "- (%s) **%s** — %s\n", pri, topic, note)
		case note != "":
			fmt.Fprintf(b, "- (%s) %s\n", pri, note)
		default:
			fmt.Fprintf(b, "- %v\n", m)
		}
	}
	b.WriteString("\n")
}

// ---------------- helpers ----------------

func toSlice(v any) []any {
	switch x := v.(type) {
	case nil:
		return nil
	case []any:
		return x
	default:
		// Single value smuggled in where a list was expected.
		return []any{x}
	}
}

func firstString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}

func firstNonNil(m map[string]any, keys ...string) any {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			return v
		}
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
