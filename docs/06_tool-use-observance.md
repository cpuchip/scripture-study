# Tool Use Observance

*A running log of tool behavior, gaps, and improvement ideas. Not a complaint box — a collaboration improvement tracker.*

*Started: February 15, 2026*

---

## How to Use This Document

When something about a tool stands out during a session — good or bad — note it here. Patterns matter more than individual incidents. Over time, this log helps us decide what to build, what to fix, and what to work around.

Categories:
- **Context pressure** — tools that return too much, pushing the context window
- **Missing tools** — gaps where a tool *should* exist but doesn't
- **Behavior** — unexpected results, formatting issues, timeouts
- **Wins** — tools working especially well, worth noting what makes them effective

---

## Observations

### February 15, 2026

**`gospel_get` is already a verse-level retrieval tool (Win)**
Tested with `Mosiah 4:9`, `Moses 6:57`, and `D&C 93:24`. For single verses, this tool is exactly what we need — it returns the verse text, configurable surrounding context (via `context` param), the file path for follow-up, cross-references, and source URL. Clean, focused, and lightweight. We don't need a new verse-retrieval tool — we need to *improve this one*.

**Verse ranges don't work (Behavior / Bug)**
Asked for `D&C 93:24-30` — only got verse 24 back. The `parseReference` function in [get.go](../scripts/gospel-mcp/internal/tools/get.go) uses `fmt.Sscanf(lastPart, "%d:%d", &chapter, &verse)` which parses `93:24-30` as chapter 93, verse 24. The `-30` is silently dropped. **Fix needed:** support range syntax (`24-30`) to return multiple verses in one call. This would eliminate the main reason to use `read_file` or `get_chapter` during document construction.

**Cross-references are chapter-scoped, not verse-scoped (Behavior)**
The `related_references` returned by `gospel_get` for Mosiah 4:9 include footnotes from across the chapter section (1 Chronicles 21, Nehemiah 8, etc.), not just verse 9's specific footnotes. Also, Topical Guide entries have a trailing ` 0` artifact (e.g., `"Reverence 0"`, `"Poor-In-Spirit 0"`). Worth investigating whether the database indexes footnotes per-verse or per-chapter-block.

**`get_chapter` (gospel-vec) returns full chapter — context pressure (Context pressure)**
Tested with Mosiah 4. Returns all 30 verses (~8KB). This is appropriate for *study* (you need the full chapter with context) but expensive when you just need a quote. The distinction matters:
- **Study mode:** Use `read_file` on the full chapter markdown (includes footnotes, formatting, cross-references in the source)
- **Document building:** Use `gospel_get` for specific verses
- **`get_chapter`'s role:** Somewhere in between — full text but without the markdown footnote formatting. Useful for gospel-vec's semantic operations but not clearly better than either `read_file` or `gospel_get` for our workflows.

**Verse-level retrieval (Revised assessment — Missing tool → Existing tool needs improvement)**
~~A dedicated tool to fetch a specific verse or range of verses would save significant context window space during document construction.~~ `gospel_get` already does this for single verses. The improvement needed is **verse range support** in the parser, not a new tool. Once ranges work, the workflow becomes:
- Study: `read_file` the full chapter markdown (footnotes, cross-refs, surrounding context)
- Building documents: `gospel_get` with verse or verse range (just the text you need to cite)

**Context window pressure (Pattern)**
Some MCP tool responses are verbose — full search results or transcript chunks can fill the context window quickly, especially in longer sessions. This is exacerbated by the current model context limits in GitHub Copilot. Worth tracking which tools are the biggest offenders and whether response truncation or summarization options would help.

**gospel-mcp search (Behavior)**
Full-text search works well for exact phrases. Semantic search via gospel-vec complements it for concept-level queries. The two together cover most needs. Worth watching for cases where neither finds what we need.

---

## Ideas for New Tools / Improvements

| Idea | Priority | Status | Notes |
|------|----------|--------|-------|
| ~~Verse range support in `gospel_get`~~ | ~~High~~ | **Done** | Parse `24-30` in `chapter:verse-endverse` format. Returns multiple verses in one call |
| ~~Fix cross-reference scoping~~ | ~~Medium~~ | **Done** | Indexer now scopes footnotes per-verse using `fn-{verse}` anchor IDs. Requires re-index |
| ~~Fix `" 0"` artifact on TG entries~~ | ~~Medium~~ | **Done** | `formatScriptureRef` now handles study aids (chapter=0, verse=0) |
| ~~Remove `get_chapter` from gospel-vec~~ | ~~Medium~~ | **Done** | Redundant with `gospel_get` + `read_file`. `search_scriptures` description updated to point to `gospel_get` |
| ~~Default context=0 in `gospel_get`~~ | ~~Medium~~ | **Done** | Lean by default for document building. Pass `context=N` when you want surrounding verses |
| Study document index | Low | Open | Search across `/study/` files by topic, date, or connected scriptures |

---

### February 15, 2026 — Fixes Applied

**Verse range parsing — FIXED** ([get.go](../scripts/gospel-mcp/internal/tools/get.go))
`parseReference` now splits the verse portion on `-` to extract `EndVerse`. New `getScriptureRange` method queries `WHERE verse >= ? AND verse <= ?`, builds numbered output, collects cross-references for all verses in the range. `D&C 93:24-30` now returns all 7 verses.

**Cross-reference scoping — FIXED** ([scripture.go](../scripts/gospel-mcp/internal/indexer/scripture.go))
`extractCrossReferences` previously ran the cross-ref regex on the *entire* footnotes section regardless of `sourceVerse`. Each footnote has an anchor like `<a id="fn-9a">` where `9` is the verse number. The function now parses each line's anchor ID, compares to `sourceVerse`, and only extracts references from matching footnotes. **Requires re-indexing** to take effect.

**TG " 0" artifact — FIXED** ([search.go](../scripts/gospel-mcp/internal/tools/search.go))
`formatScriptureRef` now detects study aids (chapter=0, verse=0) and formats them as title-cased topic names instead of appending ` 0`.

**`get_chapter` removed from gospel-vec** ([mcp.go](../scripts/gospel-vec/mcp.go))
Tool definition, handler case, and implementation removed. `search_scriptures` and truncation messages updated to reference `gospel_get` instead.

**Default context=0** ([get.go](../scripts/gospel-mcp/internal/tools/get.go))
`gospel_get` no longer defaults to `context=3`. Returns just the requested verse(s) by default — lean for document building. Pass `context=N` explicitly when studying.

---

*This is a living document. Add observations as they arise during any session.*
