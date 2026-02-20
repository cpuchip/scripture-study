---
name: source-verification
description: "Verify scripture and talk quotes against actual source files. Enforces read-before-quoting, the cite count rule, and the pre-publish checklist. Load when writing studies, evaluations, lessons, or any document that cites scriptures or conference talks."
user-invokable: false
---

# Source Verification

## The Core Rule

**Search results are pointers, not sources.** Use search tools to *find* what to study. Use `read_file` to *actually study it*. Use webster-mcp to *understand the language*. Each tool has a role; none replaces the others.

## Two-Phase Workflow

### Phase 1 — Discovery (search freely)
- `gospel_search` for keyword/phrase search across scriptures, conference, manuals
- `search_scriptures` / `search_talks` for semantic/concept search
- `define` / `webster_define` for historical word meanings
- `web_search` for external scholarly context
- **Note file paths and references** — these are *pointers* to read in Phase 2, not sources to quote

### Phase 2 — Deep Reading (verify everything)
- For EVERY scripture you plan to quote, `read_file` the actual chapter markdown from `gospel-library/`
- For EVERY conference talk you plan to cite, `read_file` the actual talk file
- **Follow the footnotes.** Scripture markdown files contain superscript footnote markers and cross-references. These are insights handed to us on a silver platter — read them, follow them, use them.
- Pull real quotes from source files, never from search excerpts

## The Cite Count Rule

For **N** citations in your output, perform at least **N** `read_file` calls.

If you cite 8 scriptures and 2 conference talks, you should have called `read_file` at least 10 times on actual source files.

## What Counts as a Source

| ✅ Source (quote from these) | ❌ Not a Source (use as pointers only) |
|------------------------------|---------------------------------------|
| `read_file` on a gospel-library markdown file | `gospel_search` excerpts |
| `read_file` on a conference talk file | `search_scriptures` summaries |
| `read_file` on a manual chapter | `search_talks` summaries |
| | Vector search results labeled `[AI SUMMARY]` |
| | Your own memory of a scripture's wording |

## Phase 3 — Writing (synthesize)

After discovery and deep reading, write the study. Weave real quotes with analysis. Follow the text where it leads.

## Phase 4 — Becoming (bridge to life)

After writing the study, include a **Becoming** section. See the `becoming` skill for full guidance. Every study should land somewhere personal — not just knowledge, but direction.

- What did this study reveal about how I should live?
- What specific practice or commitment does this point toward?
- Is there an existing `becoming/` document this connects to?
- What would it look like to apply this next week?

## Quality Rhythm

**Discovery → Reading → Writing → Becoming.** Start broad (search), go deep (read full sources with footnotes), synthesize (write), then land it personally (become). Never write from search results. Never end at synthesis.

## Pre-Publish Checklist

Before finalizing any study, lesson, talk, or evaluation, verify:

- [ ] Every quoted passage verified against the actual source file (not search excerpts or vector summaries)
- [ ] Every conference talk reference links to a **specific talk file**, not a conference directory
- [ ] Every scripture reference links to the specific chapter file with correct path
- [ ] Files claimed to exist are verified with `file_search` or `list_dir`
- [ ] Webster 1828 definitions used where historical meaning differs from modern usage
- [ ] All markdown links are relative and follow project conventions
- [ ] `read_file` was used at least once per cited source
- [ ] Footnotes were followed and incorporated where they add insight
- [ ] Study includes a "Becoming" section with specific personal questions and/or commitments
- [ ] If a related `becoming/` document exists, it's linked
