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

### April 20, 2026

**Gospel MCP servers + Webster `define` were disabled in a study-mode session (Behavior)**

During the inaugural study using the new ibeco.me-issued engine token (`study/give-away-all-my-sins.md`), every gospel-engine and gospel-vec tool returned "Tool currently disabled by the user." Webster `webster_search` worked but `webster_define` and `mcp_webster_define` were both disabled. Workarounds that landed cleanly: `grep_search` over `gospel-library/` with `includeIgnoredFiles: true` for phrase singularity checks, and a one-line Python script reading `scripts/webster-mcp/data/webster1828.json.gz` directly for definitions.

Worth checking the model's tool-allowlist for study mode — the `tools: [...]` frontmatter in `study.chatmode.md` may be filtering them out unintentionally. The chatmode declares `gospel/*` and `webster/*` permission, but the prefix may not be matching against the actual tool names (`mcp_gospel-engine_*`, `mcp_webster_*`). Worth one debugging pass.

Net effect: the study still produced a strong result via direct file reads, but lost the semantic-search ability to discover unexpected cross-references. The phrase-singularity finding ("give away all" appears nowhere else in scripture) was only possible because grep was available.



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

### March 31, 2026

**Gospel-engine semantic search excels for talk discovery (Win)**
During the Art of Delegation study, `gospel_search` with `mode: "semantic"` and `sources: ["conference"]` found 4 relevant talks including Ballard's "O Be Wise" (2006), Holland's "A Robe, a Ring, and a Fatted Calf" (1985), and Christofferson's "Moral Agency" (2009). Source filtering works well — narrowing to conference talks gave focused, relevant results. This is exactly the tool's sweet spot: discovering content you don't have a reference for.

**`read_file` still preferred for known-reference deep reading (Pattern)**
For 9 of 11 scripture sources in the study, the agent bypassed gospel-engine entirely and used `read_file` on gospel-library markdown files. When you already know the reference (e.g., Exodus 18, D&C 84:33-42), `read_file` gives full context — surrounding verses, footnotes, formatting — that a retrieval tool truncates. **Design principle:** MCP tools are for *discovery*; `read_file` is for *understanding*. Don't use MCP snippets as a crutch to avoid reading the full source in context. Good for direct quotes, bad for total context understanding. Michael has flagged this pattern before.

**`gospel_get` has no verse-level retrieval (Regression from gospel-mcp)**
Gospel-engine's `handleGet` fetches *entire chapters* from the `chapters` table — no verse-level granularity at all. The old gospel-mcp had verse-range support (fixed Feb 15) using the `scriptures` table. Gospel-engine has the `scriptures` table indexed with individual verses, but `handleGet` doesn't query it. During the study, requesting D&C 84 returned the full 48KB chapter. Fell back to `grep_search` + `read_file` (3 calls instead of 1). The proposal's Phase 1 tool spec envisions a `reference` parameter with scripture reference parsing, but it wasn't implemented.

**`/gospel-library/` in `.gitignore` breaks VS Code search tools (Root cause identified)**
`grep_search` and `file_search` silently exclude `/gospel-library/` because it's gitignored (line 4 of `.gitignore`). Setting `includeIgnoredFiles: true` works around this, but the agent doesn't use it by default. This creates a pattern where the agent tries to search gospel-library content with workspace tools, gets empty results, and falls back to MCP or blind `read_file` guesses. Three solutions: (1) `.vscode/settings.json` to override `search.exclude`, (2) agent instructions to always pass `includeIgnoredFiles: true`, (3) rely on gospel-engine MCP for retrieval. Option 2 is cheapest. Option 3 is most robust.

**Cross-references exist in the DB but no tool exposes them (Gap)**
The `cross_references` table has 85,590 entries. The `edges` table holds graph relationships. Neither is accessible through any MCP tool. During the study, tracing delegation patterns across standard works required manual cross-referencing — opening each chapter's markdown and scanning footnotes. A tool that returns "what scriptures are cross-referenced from this verse?" would have accelerated the study significantly.

---

## Ideas for New Tools / Improvements

| Idea | Priority | Status | Notes |
|------|----------|--------|-------|
| ~~Verse range support in `gospel_get`~~ | ~~High~~ | **Done** | Parse `24-30` in `chapter:verse-endverse` format. Returns multiple verses in one call |
| ~~Fix cross-reference scoping~~ | ~~Medium~~ | **Done** | Indexer now scopes footnotes per-verse using `fn-{verse}` anchor IDs. Requires re-index |
| ~~Fix `" 0"` artifact on TG entries~~ | ~~Medium~~ | **Done** | `formatScriptureRef` now handles study aids (chapter=0, verse=0) |
| ~~Remove `get_chapter` from gospel-vec~~ | ~~Medium~~ | **Done** | Redundant with `gospel_get` + `read_file`. `search_scriptures` description updated to point to `gospel_get` |
| ~~Default context=0 in `gospel_get`~~ | ~~Medium~~ | **Done** | Lean by default for document building. Pass `context=N` when you want surrounding verses |
| Verse-level retrieval in gospel-engine `gospel_get` | **High** | Open | Phase 1 only has volume/book/chapter. Needs `reference` param with scripture parsing (e.g., "D&C 93:24-30") querying the `scriptures` table. Regression from gospel-mcp. |
| Cross-reference retrieval tool or parameter | **High** | Open | 85,590 cross-refs in DB, no tool exposes them. Add `cross_references: true` param to `gospel_get` or a dedicated tool. |
| Agent instruction for `includeIgnoredFiles` | Medium | Open | `/gospel-library/` is gitignored → `grep_search` / `file_search` return nothing. Either update copilot-instructions or create `.vscode/settings.json`. |
| Study document index | Low | Open | Search across `/study/` files by topic, date, or connected scriptures |
| BYU Speeches downloader / MCP tool | Medium | Open | BYU devotionals (speeches.byu.edu) are not in gospel-library but are frequently cited primary sources. Maxwell's "Meekly Drenched in Destiny" was a gap during the [Mormon YouTube evaluation](../study/yt/UjzeDUBMaUA-problem-with-mormon-youtube.md). A `byu-speeches/` directory + download tool (similar to yt-mcp) would cover Holland, Eyring, Maxwell devotionals, etc. Site has clean HTML with full text, audio, and video. |

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

---

### February 28, 2026

**No automated quote verification tool exists (Missing tool / Critical gap)**
During verification of the 7-part Working with AI guide series, we discovered 3 wrong scripture quotes, 2 fabricated YouTube composites, and multiple minor wording errors — all generated from training-data memory during the writing phase. The source-verification skill has the rules to prevent this, but they're manual and self-reported. There is no tool that:
1. Extracts all blockquote attributions from a document
2. Resolves each attribution to a source file path
3. Compares the quoted text against the actual source
4. Reports mismatches

**Impact:** 45 manual corrections across 7 files. Every one was preventable if verified during writing.

**Potential solutions:**
- A `verify-quotes` CLI tool that parses markdown blockquotes, resolves scripture references to file paths, and diffs quoted text against source files
- A pre-commit hook or publish-step that runs automated quote verification
- A `verify` slash command in the study agent that scans the current document for unverified citations
- Integration with `gospel_get` to check verse text without full file reads

**Priority:** High — this is a trust/integrity issue, not a convenience feature.

**Source-verification skill scope widened (Process improvement)**
The skill description and checklist were updated to apply to ALL document types, not just studies/lessons/evaluations. Added "Quote Hygiene" section distinguishing direct quotes (verbatim, verified), paraphrases (indirect speech, no quotes), and references (see Source). Added explicit confabulation warning. See [source-verification SKILL.md](../.github/skills/source-verification/SKILL.md).

**New bias pattern added: Memory Confabulation (#8)**
Documented in [biases.md](biases.md) with three sub-patterns: wording drift, phantom attribution, and fabricated composites. Includes detection heuristic: "If you wrote a direct quote without having called `read_file` on its source during this session, the quote is suspect."
