---
workstream: WS3
status: superseded
brain_project: 3
created: 2026-04-10
last_updated: 2026-05-13
superseded_by: scripts/gospel-engine-v2/.spec/proposals/README.md
---

# gospel-engine Phase 1.5 — Agent-Ergonomic Improvements

> **Superseded 2026-05-13.** This proposal has been ratified into per-phase
> execution specs that live in the gospel-engine-v2 repo (where the code
> lives):
>
> - **Rollup overview:** [scripts/gospel-engine-v2/.spec/proposals/README.md](../../../scripts/gospel-engine-v2/.spec/proposals/README.md)
> - Phase 1.5a — mode enum docs fix
> - Phase 1.5b — `gospel_get` handler rewrite (ref/reference, verse-range, chapter-level)
> - Phase 1.5c — cross-references (opt-in)
> - Phase 1.5d — speaker indexer fix + parse-failures log
> - Phase 1.5e — study aids indexed (TG, BD, GS, JST) — new `study_aids` table
> - Phase 3-research — v3 architecture spike (proxy-pointer + LightRAG)
>
> **Phase 2 (TITSW migration to v2)** is deferred past this rollup; revisit
> after 1.5 + research land. Original Phase 2 scope retained in [main.md](main.md).
>
> The body below is preserved as historical context for what was originally
> proposed and how it evolved.

---

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

## Enhancements 4–6 (added 2026-05-13)

Reproduced live against the running v2 hosted engine in this planning session.

### Enhancement 4 — Mode enum drift (docs vs server)

**Symptom.** Calling `gospel_search { mode: "combined" }` returns `ERROR: Your input to the tool was invalid (must be equal to one of the allowed values)`.

**Root cause.** Server enum at `scripts/gospel-engine-v2/cmd/gospel-mcp/main.go:~280` is `["keyword", "semantic", "hybrid"]`. Documentation in `.github/copilot-instructions.md` says `"keyword" | "semantic" | "combined"`. CLAUDE.md likely echoes the same. Pure documentation drift.

**Fix.** Update `.github/copilot-instructions.md` and any agent files that document the mode parameter to say `hybrid`, not `combined`. No code change.

**Estimated effort.** 5 minutes.

### Enhancement 5 — `ref` vs `reference` param-name leak

**Symptom.** When the agent calls `gospel_get` wrong, the HTTP error message says `provide either ref= or (type= and id=)`. The MCP schema actually accepts `reference`, not `ref`. The agent reasonably retries with `ref:`, which the MCP layer silently drops, and the call fails again. Two passes lost to the mismatch.

**Root cause.** Schema layer (MCP) and HTTP layer have different param names. Error message names the HTTP one.

**Fix (recommended).** In `internal/api/server.go:handleGet`, accept BOTH `ref` and `reference` as query params. Five extra lines, eliminates the foot-gun forever. Update the error message to say `reference=` so future drift doesn't recur.

**Estimated effort.** 10 minutes.

### Enhancement 6 — Speaker field corrupted by indexer

**Symptom.** `gospel_search` results include `"speaker": "🎧 Listen to Audio"` for talks instead of the actual speaker name (e.g., "Elder Wan-Liang Wu"). Reproduced 2026-05-13 on Wu's 2026-04 talk and on Benson's 1983-10 talk.

**Root cause.** Indexer in `scripts/gospel-engine-v2/internal/indexer/` is grabbing the audio-link button text instead of the speaker line in the talk markdown.

**Fix.** Locate the speaker-extraction selector in the indexer; correct it to read the actual speaker line. **Requires reindex of the `talks` table** to take effect on already-indexed content.

**Estimated effort.** Half session including reindex.

---

## Implementation Priority

| # | Enhancement | Priority | Effort | Impact |
|---|-------------|----------|--------|--------|
| 1 | Verse-level `gospel_get` (range support) | **High** | Small (port from gospel-mcp) | Eliminates most common friction |
| 2 | Cross-reference retrieval | **High** | Small (SQL on existing table) | Unlocks 85K cross-refs for study |
| 3 | Search filtering fix (`includeIgnoredFiles`) | ~~Medium~~ | ~~Trivial~~ | **Done — in copilot-instructions.md** |
| 4 | Mode enum docs fix (`combined` → `hybrid`) | **High** | 5 min, docs only | Stops every session burning a call on wrong enum |
| 5 | Dual-accept `ref`/`reference` + error message fix | **High** | 10 min Go | Stops 2-pass loss when calls fail |
| 6 | Speaker indexer fix + reindex | Medium | Half session | Search results show real speaker names |

**Phasing recommendation (2026-05-13).**
- **Phase 1.5a — docs only (15 min):** ship #4 today. No build, no rebuild. Eliminates the cheapest, most-frequent papercut.
- **Phase 1.5b — small Go (half session):** ship #1 + #5 together. Both touch `handleGet` + the MCP wrapper schema. Same rebuild.
- **Phase 1.5c — indexer + reindex (half session):** ship #6. Decide reindex timing.
- **Phase 1.5d — cross-refs (half-to-full session):** ship #2 once Phase 1.5b is stable.

Enhancement #3 is closed; mark it done in the table above.

Code paths to touch (v2):
- `scripts/gospel-engine-v2/internal/api/server.go` — `handleGet`, `getByReference`, error messages
- `scripts/gospel-engine-v2/cmd/gospel-mcp/main.go` — `gospel_get` schema (already takes `reference`; may add range docs)
- `scripts/gospel-engine-v2/internal/indexer/` — speaker extraction
- `.github/copilot-instructions.md` + `CLAUDE.md` — mode-enum line

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
