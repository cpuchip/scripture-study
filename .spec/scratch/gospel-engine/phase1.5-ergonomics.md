# gospel-engine Phase 1.5 — Scratch

**Binding problem:** Agent-ergonomic gaps discovered during Art of Delegation study (Mar 31, 2026).

---

## Research Provenance

### Source: Art of Delegation study (Mar 31)
- Used gospel-engine for 4 conference talk discoveries via `gospel_search` (semantic, sources: conference)
- Bypassed gospel-engine for 9 of 11 scripture sources → used `read_file` directly on gospel-library markdown
- `gospel_get` with volume=dc-testament, book=dc, chapter=84 returned full 48KB chapter. Fell back to grep_search + read_file (3 calls)
- No way to get cross-references for any verse through MCP tools

### Source: gospel-engine codebase audit (Mar 31)
- `tools.go` handleGet: queries `chapters` table only for scripture. No verse-level access despite `scriptures` table having 41,995 individual verses.
- `schema.sql`: `cross_references` table has source/target columns with indexes on both. `edges` table for graph relationships. Neither exposed through MCP.
- `search.go`: Search returns individual verse results (queries `scriptures` table), but `get` doesn't offer the same granularity.

### Source: gospel-mcp comparison (Mar 31)
- `get.go` in gospel-mcp has: `parseReference`, `getScripture`, `getScriptureRange`, `getCrossReferences`
- parseReference handles: book name normalization, chapter:verse, verse ranges (24-30), multi-word book names
- getScriptureRange: `WHERE book = ? AND chapter = ? AND verse >= ? AND verse <= ?`
- getCrossReferences: queries `cross_references` table by source_book, source_chapter, source_verse
- All this was working as of Feb 15 fixes. Gospel-engine just didn't port it.

### Source: .gitignore (Mar 31)
- Line 4: `/gospel-library/` — causes grep_search and file_search to silently exclude
- No `.vscode/settings.json` exists
- `includeIgnoredFiles: true` works around it but agent doesn't use it by default

### Michael's snippet crutch warning (Mar 31)
> "I will caution about using it as a crutch to get text. We had issues before where you would just use it to get snippets out of context (good for direct quotes, bad for total context understand). So work best around that please."

Design principle: discovery (gospel_search) + quoting (gospel_get) + understanding (read_file). Don't collapse the last two.

---

## Decisions

- **Phase 1.5 scoped to 3 enhancements:** verse-level get, cross-reference retrieval, search filtering fix
- **Port from gospel-mcp, don't reinvent:** parseReference and getCrossReferences already tested and working
- **Option D for search filtering:** agent instruction + gospel-engine as primary. No .vscode/settings.json needed.
- **cross_refs as opt-in parameter:** default false to keep responses lean
- **Reverse lookup deferred:** "what references this verse" is a Phase 3 graph query feature
