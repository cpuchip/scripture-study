# gospel-engine Phase 1.5 — Agent-Ergonomic Improvements

**Binding problem:** Gospel-engine Phase 1 ships 3 working MCP tools, but the agent's actual study workflow during the Art of Delegation study revealed three friction points: (1) `gospel_get` fetches entire chapters with no verse-level retrieval — a regression from gospel-mcp's Feb 15 verse-range fix, (2) 85,590 cross-references sit in the database with no tool to query them, and (3) `/gospel-library/` being gitignored makes VS Code's `grep_search` and `file_search` silently return nothing for scripture files.

These aren't Phase 2 (TITSW enrichment) items — they're foundational retrieval ergonomics that make the existing tool surface useful for study.

**Created:** 2026-03-31
**Parent proposal:** [main.md](main.md)
**Observed during:** Art of Delegation study (March 31, 2026)
**Logged in:** [docs/06_tool-use-observance.md](../../../docs/06_tool-use-observance.md)

---

## Design Principle: Discovery vs. Understanding

Michael's caution, based on prior experience:

> "I will caution about using it as a crutch to get text. We had issues before where you would just use it to get snippets out of context (good for direct quotes, bad for total context understanding)."

The tool should support **two distinct workflows**:

| Workflow | Tool | Purpose |
|----------|------|---------|
| **Discovery** | `gospel_search` (semantic/keyword) | Find content you don't have a reference for |
| **Quoting** | `gospel_get` with verse reference | Get exact text for a known reference (direct quotes, citation) |
| **Understanding** | `read_file` on gospel-library markdown | Full chapter with footnotes, formatting, surrounding context |

Gospel-engine should excel at discovery and quoting. It should **not** try to replace `read_file` for deep reading. The tool description should actively guide the agent toward `read_file` when full-context understanding is the goal.

---

## Enhancement 1: Verse-Level Retrieval in `gospel_get`

**Current state:** `handleGet` queries the `chapters` table (full chapter text). No way to get individual verses or verse ranges. The `scriptures` table has 41,995 individual verses indexed and ready.

**Target state:** Add a `reference` parameter that accepts human-readable scripture references. Port the `parseReference` logic from gospel-mcp (already tested, handles edge cases).

### Supported reference formats
```
"1 Nephi 3:7"           → single verse
"D&C 93:24-30"          → verse range
"Moses 6:57"            → single verse
"Mosiah 4"              → full chapter (current behavior)
"John 3:16-17"          → verse range
```

### Implementation

1. Add `reference` string parameter to `handleGet`
2. Port `parseReference` from `scripts/gospel-mcp/internal/tools/get.go` — it already handles:
   - Book name normalization ("1 Nephi" → "1-ne", "D&C" → "dc")
   - Chapter:verse parsing
   - Verse range parsing (`24-30`)
   - Multi-word book names
3. For single verse: `SELECT text, file_path, source_url FROM scriptures WHERE book = ? AND chapter = ? AND verse = ?`
4. For verse range: `SELECT verse, text FROM scriptures WHERE book = ? AND chapter = ? AND verse >= ? AND verse <= ? ORDER BY verse`
5. For chapter only: current behavior (query `chapters` table)

### Response format
```
Reference: D&C 93:24-30
File: gospel-library/eng/scriptures/dc-testament/dc/93.md

24. And truth is knowledge of things as they are, and as they were, and as they are to come;
25. And whatsoever is more or less than this is the spirit of that wicked one who was a liar from the beginning.
...
30. All truth is independent in that sphere in which God has placed it, to act for itself, as all intelligence also; otherwise there is no existence.
```

Lean output. No 48KB chapter dumps. Use `read_file` for full context.

### Scope
- **Port:** `parseReference`, `getScripture`, `getScriptureRange` from gospel-mcp
- **Add:** `reference` parameter to tool definition's InputSchema
- **Modify:** `handleGet` to dispatch on `reference` before checking volume/book/chapter
- **Keep:** volume/book/chapter path as fallback (some agents may use structured params)

### Estimated effort
Small — the logic already exists in gospel-mcp. Port + adapt to gospel-engine's DB wrapper.

---

## Enhancement 2: Cross-Reference Retrieval

**Current state:** `cross_references` table has 85,590 entries, indexed on both source and target. The `edges` table holds graph relationships. Neither is queryable through any MCP tool.

**Target state:** When `gospel_get` returns a verse or verse range, optionally include cross-references for those verses.

### Implementation

Add `cross_refs` boolean parameter to `gospel_get` (default: false). When true, query:

```sql
SELECT target_volume, target_book, target_chapter, target_verse, reference_type
FROM cross_references
WHERE source_book = ? AND source_chapter = ? AND source_verse = ?
```

For verse ranges, query all verses in the range and deduplicate.

### Response format addition
```
Cross-references for D&C 84:33:
  - Numbers 25:13 (footnote)
  - Hebrews 7:11-12 (footnote)
  - D&C 107:40 (footnote)
  - TG: Priesthood, Aaronic (topical guide)
```

### Reverse lookup consideration

A separate parameter or mode — "what scriptures reference *this* verse?" — queries the target index:

```sql
SELECT source_volume, source_book, source_chapter, source_verse
FROM cross_references
WHERE target_book = ? AND target_chapter = ? AND target_verse = ?
```

This is powerful for study but adds complexity. Consider as a follow-up or make it a `direction` parameter: `"from"` (default, what does this verse reference?) or `"to"` (what references this verse?).

### Estimated effort
Small — straightforward SQL queries on an already-indexed table. The gospel-mcp `getCrossReferences` function is a direct reference.

---

## Enhancement 3: Workspace Search Filtering Fix

**Current state:** `/gospel-library/` is in `.gitignore` (line 4), which causes VS Code's `grep_search` and `file_search` to silently exclude all gospel-library files. The agent tries to search, gets empty results, and falls back to blind `read_file` guesses or MCP calls.

**Root cause:** `.gitignore` drives both git tracking AND VS Code's default search exclusion.

### Options

| Option | Mechanism | Effort | Trade-off |
|--------|-----------|--------|-----------|
| A. `.vscode/settings.json` | `"search.exclude": { "**/gospel-library": false }` to re-include | Low | Adds a settings file; only affects VS Code, not agent instructions |
| B. Agent instructions | Add to copilot-instructions.md: "For gospel-library searches, always pass `includeIgnoredFiles: true`" | Low | Relies on agent following instructions consistently |
| C. Rely on gospel-engine | Stop using `grep_search`/`file_search` for scripture content; use `gospel_search` for discovery and `gospel_get` for retrieval | Zero (already works) | Doesn't help when agent needs to find a *file path* to `read_file` |
| **D. B + C combined** | Agent instructions + gospel-engine as primary | Low | Belt and suspenders — gospel-engine for discovery/quoting, `includeIgnoredFiles` for the cases where `file_search`/`grep_search` is genuinely needed |

**Recommendation: Option D.** Two lines in copilot-instructions.md + rely on gospel-engine as the primary scripture retrieval path. The agent instruction catches the cases where workspace search is genuinely needed (e.g., finding file paths, checking file existence before linking).

### Implementation
1. Add a `## Gospel Library Search` section to `.github/copilot-instructions.md`:
   > The `/gospel-library/` directory is gitignored (too large for git). When using `grep_search` or `file_search` on gospel-library content, always pass `includeIgnoredFiles: true`. Prefer `gospel_search` and `gospel_get` for scripture/talk retrieval.
2. No `.vscode/settings.json` needed — the instruction is sufficient and doesn't add configuration debt.

---

## Implementation Priority

| Enhancement | Priority | Effort | Impact |
|-------------|----------|--------|--------|
| 1. Verse-level `gospel_get` | **High** | Small (port from gospel-mcp) | Eliminates the most common friction point |
| 2. Cross-reference retrieval | **High** | Small (SQL on existing table) | Unlocks 85K cross-refs for study |
| 3. Search filtering fix | Medium | Trivial (2 lines in instructions) | Prevents silent search failures |

All three can ship in one session. Enhancement 3 can be done immediately (no code changes needed). Enhancements 1 and 2 are code changes to `scripts/gospel-engine/internal/mcp/tools.go`.

---

## Verification

1. **Verse retrieval:** `gospel_get` with reference "D&C 93:24-30" returns 7 verses, not the full chapter
2. **Single verse:** `gospel_get` with reference "1 Nephi 3:7" returns one verse with file path
3. **Cross-references:** `gospel_get` with reference "Mosiah 4:9" and `cross_refs: true` returns verse-scoped footnote references
4. **Search filtering:** `grep_search` with `includeIgnoredFiles: true` on gospel-library content returns results
5. **Snippet crutch check:** Tool description on `gospel_get` says "Use `read_file` for full chapter context with footnotes and formatting"

---

## Relationship to Phase 2 (TITSW Enrichment)

These changes are **orthogonal** to Phase 2. They improve the retrieval ergonomics of the tool surface that already exists. Phase 2 (TITSW scoring, enriched summaries, calibrated prompts) builds on top of a working retrieval layer — and that layer needs verse-level access to be useful.

Build order: Phase 1.5 (this) → Phase 2 (TITSW) → Phase 3 (graph queries, thematic edges).
