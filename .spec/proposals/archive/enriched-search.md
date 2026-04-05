# Enriched Search: TITSW Metadata in gospel-mcp

*Proposal — 2026-03-29*
*Depends on: [enriched-indexer.md](enriched-indexer.md) (Phase 1+ must be complete)*
*Scratch: [.spec/scratch/enriched-indexer/main.md](../scratch/enriched-indexer/main.md)*

---

## ⚠️ ARCHITECTURAL PIVOT (Mar 29)

**This proposal is SUPERSEDED in its current form.** Michael decided (Mar 29) to build a **new combined gospel tool** that merges gospel-mcp (SQLite/FTS) and gospel-vec (vector/chromem-go) into one application, rather than modifying the originals.

**Rationale:** Keep gospel-mcp and gospel-vec unchanged and available for study work during reindexing. The new combined tool shares one SQLite DB and one vector DB, avoiding the two-tool import pipeline described below.

**What changes:**
- Option C below (ALTER TABLE in gospel-mcp, read gospel-vec cache) → replaced by a single app that owns both databases
- The schema design (TITSW columns, FTS enhancement) is still valid — just lives in the new tool
- The search/get tool enhancements are still valid — just exposed by the new tool's MCP server
- No import pipeline needed between separate tools — the combined tool indexes and queries in one codebase

**Decision recorded in:** [decisions.md](../memory/decisions.md) — "Combined gospel tool for enriched pipeline"

**Next step:** Write a new proposal for the combined tool (working name: `gospel-study` or `gospel-combined`). This proposal remains as historical reference for the schema design and tool enhancement specs.

---

## Binding Problem

gospel-vec will produce enriched summaries with TITSW teaching profiles (dominant dimensions, teaching modes, scores). But gospel-vec only supports *semantic* search — you can search by meaning but not by exact field values. There's no way to ask "find all talks where titsw_mode = enacted" or "all talks with teach_score >= 7" efficiently.

gospel-mcp has the opposite strength: SQLite + FTS5 gives fast keyword and structured search, but its `talks` table has only basic metadata (year, month, speaker, title, content). No teaching dimensions.

The two systems are complementary. The question is: how should they share the enriched data?

**Who is affected:** Michael doing lesson prep ("find me 3 talks that enact love"), brain app search (filtered semantic queries), any MCP consumer that wants structured + semantic search combined.

**How would you know it's fixed:** `gospel_search` can filter by teaching mode, dimension scores, and dominant dimensions. `gospel-vec` semantic search can be pre-filtered by structured metadata. A single query like "show me talks about the Atonement that are primarily experiential" combines both.

---

## Success Criteria

1. TITSW fields are searchable via `gospel_search` (structured filters)
2. FTS on enriched keywords/summaries captures TITSW vocabulary
3. Clear data flow: gospel-vec generates → export → gospel-mcp imports
4. No duplicate LLM calls — gospel-mcp consumes gospel-vec output, doesn't regenerate
5. Existing `gospel_search`, `gospel_get`, `gospel_list` tools continue working unchanged
6. New fields surface in search results without breaking existing consumers

---

## Constraints

- **gospel-vec has no SQLite.** Pure chromem-go. Adding SQLite to gospel-vec would be a major architectural change.
- **gospel-mcp has no LLM.** Zero. It shouldn't need one — it consumes structured data.
- **chromem-go metadata is `map[string]string`.** Flat fields only. Already handled by enriched-indexer proposal.
- **Schema migrations.** gospel-mcp uses embedded `schema.sql` with a `schema_version` table. New columns = schema version bump.
- **NOT in scope:** New MCP tools. We're extending existing tools, not creating new endpoints.

---

## Architecture Options

### Option A: gospel-vec Exports → gospel-mcp Imports

**gospel-vec** generates enriched summaries (it already has the LLM pipeline). After indexing, it exports TITSW metadata to a file (JSON or SQLite). **gospel-mcp** imports that file during its own indexing pass and populates new columns in the `talks` table.

```
gospel-vec index-talks → cached summaries (JSON) → TITSW metadata
                                                          ↓
gospel-mcp index → reads markdown files + reads TITSW export → talks table with TITSW columns
```

**Pros:** Clean separation. Each system does what it's good at. No new dependencies.
**Cons:** Two-step indexing required. Must keep exports in sync.

### Option B: Shared SQLite Database

Add a lightweight SQLite database to gospel-vec for TITSW metadata (not replacing chromem-go — supplementing it). gospel-mcp reads from this database during search to enrich results.

```
gospel-vec index-talks → chromem-go (vectors) + sqlite (TITSW metadata)
                                                          ↓
gospel-mcp search → own SQLite (FTS) + gospel-vec SQLite (TITSW) → combined results
```

**Pros:** Single source of truth. No export/import step.
**Cons:** Cross-process SQLite access (WAL mode helps). gospel-vec gets a new dependency. Two databases to coordinate.

### Option C: gospel-mcp Gets TITSW Columns Directly

Add TITSW columns to gospel-mcp's `talks` table. During gospel-mcp's `index` command, it reads the gospel-vec summary cache (JSON files on disk) and populates the columns.

```
gospel-vec index-talks → summary cache (JSON files in data/summaries/)
                                            ↓
gospel-mcp index → reads markdown files + reads summary cache → talks table with TITSW columns + FTS
```

**Pros:** Simplest. gospel-mcp already reads files during indexing. Summary cache is just more files to read. No new tools, no new databases, no runtime coordination.
**Cons:** Tight coupling to gospel-vec's cache format. gospel-mcp needs to know where gospel-vec stores its data.

---

## Recommended Approach: Option C

Option C wins on simplicity. The summary cache already exists as JSON files on disk. gospel-mcp already reads files during indexing. Adding "also read the summary cache and extract TITSW fields" is a small change to the indexer, not an architectural shift.

The tight coupling concern is manageable: the cache directory is configurable, and the JSON format is simple (see cache format in enriched-indexer proposal). If the format changes, it's one parser to update.

### Schema Changes

Add TITSW columns to the `talks` table:

```sql
-- Schema version 2: Add TITSW teaching profile columns
ALTER TABLE talks ADD COLUMN titsw_dominant TEXT;      -- "teach_about_christ,invite"
ALTER TABLE talks ADD COLUMN titsw_mode TEXT;           -- "enacted"
ALTER TABLE talks ADD COLUMN titsw_pattern TEXT;        -- "story→doctrine→invitation"
ALTER TABLE talks ADD COLUMN titsw_teach INTEGER;       -- 0-9
ALTER TABLE talks ADD COLUMN titsw_help INTEGER;        -- 0-9
ALTER TABLE talks ADD COLUMN titsw_love INTEGER;        -- 0-9
ALTER TABLE talks ADD COLUMN titsw_spirit INTEGER;      -- 0-9
ALTER TABLE talks ADD COLUMN titsw_doctrine INTEGER;    -- 0-9
ALTER TABLE talks ADD COLUMN titsw_invite INTEGER;      -- 0-9
ALTER TABLE talks ADD COLUMN titsw_summary TEXT;        -- enriched summary from gospel-vec
ALTER TABLE talks ADD COLUMN titsw_key_quote TEXT;      -- key quote from gospel-vec
ALTER TABLE talks ADD COLUMN titsw_keywords TEXT;       -- comma-separated enriched keywords
```

Scores are INTEGER in gospel-mcp (actual values), vs STRING in gospel-vec chromem-go (flat metadata). Conversion at import time.

### FTS Enhancement

The `talks_fts` table currently indexes: `title, speaker, content`. Extend to include TITSW fields:

```sql
-- Drop and recreate FTS with additional columns
CREATE VIRTUAL TABLE IF NOT EXISTS talks_fts USING fts5(
    title,
    speaker,
    content,
    titsw_dominant,
    titsw_mode,
    titsw_keywords,
    titsw_summary,
    content='talks',
    content_rowid='id'
);
```

This enables queries like `gospel_search("enacted love")` to match talks where mode is enacted AND dominant includes love — through FTS, not through `WHERE` clauses. Both structured (`WHERE titsw_teach >= 7`) and full-text (`MATCH 'enacted love'`) queries work.

### Import Pipeline

During `gospel-mcp index`:
1. Index talk markdown files as before (year, month, speaker, title, content)
2. For each talk, look up the corresponding summary cache file: `{cache_dir}/talk-{year}-{month}-{filename}.json`
3. If found AND `prompt_version >= "v2"`, parse the `teaching_profile` fields
4. Populate TITSW columns on the talks row
5. If no cache file found, leave TITSW columns NULL (graceful degradation)

### Search Tool Enhancement

Add optional TITSW filter parameters to `gospel_search`:

```json
{
  "query": "atonement",
  "source": "conference",
  "titsw_mode": "enacted",
  "titsw_min_teach": 6,
  "titsw_dominant": "love"
}
```

The tool builds a combined query: FTS on `query`, `WHERE titsw_mode = ?` if provided, `WHERE titsw_teach >= ?` if min score provided.

### Get Tool Enhancement

`gospel_get` returns TITSW fields when retrieving a talk:

```json
{
  "reference": "Patrick Kearon, April 2024",
  "title": "Welcome to the Church of Joy",
  "content": "...",
  "titsw": {
    "dominant": ["help_come_to_christ", "love"],
    "mode": "enacted",
    "pattern": "invitation→doctrine→testimony",
    "scores": { "teach": 5, "help": 7, "love": 7, "spirit": 5, "doctrine": 4, "invite": 7 },
    "summary": "Elder Kearon invites members to rediscover joy...",
    "key_quote": "Welcome to the church of joy!"
  }
}
```

---

## Phased Delivery

### Phase 1: Schema + Import (1 session)

1. Bump schema version to 2 with migration SQL
2. Write `importTITSWFromCache()` — reads gospel-vec summary cache JSON files
3. Add TITSW columns to talks table
4. Re-run `gospel-mcp index` and verify TITSW fields populated
5. **Test:** Query `SELECT speaker, title, titsw_mode, titsw_teach FROM talks WHERE titsw_teach >= 7` and verify results

### Phase 2: Search Enhancement (1 session)

1. Recreate `talks_fts` with TITSW columns
2. Add `titsw_mode`, `titsw_min_*`, `titsw_dominant` parameters to `SearchParams`
3. Update `searchTalks()` to apply TITSW filters
4. Update MCP tool definition with new parameters
5. **Test:** `gospel_search("atonement", source="conference", titsw_mode="enacted")` returns only enacted talks

### Phase 3: Get Enhancement (1 session)

1. Update `GetResponse` struct with TITSW fields
2. Populate TITSW in `getTalk()` response
3. Update `gospel_get` MCP tool definition
4. **Test:** `gospel_get(reference="Patrick Kearon, April 2024")` includes teaching profile

---

## Verification Strategy

| Phase | Verification | Criteria |
|---|---|---|
| 1 | Import validation | TITSW fields populated for 90%+ of talks (some old talks may not have cache) |
| 1 | NULL handling | Talks without cache files have NULL TITSW fields, no errors |
| 2 | Filtered search | Mode/score filters produce correct subsets |
| 2 | FTS + TITSW | Combined text query + TITSW filter works |
| 2 | Backward compat | Existing queries without TITSW params still work |
| 3 | Get response | TITSW fields present in talk responses |
| All | No regression | `gospel_search`, `gospel_get`, `gospel_list` all pass existing test patterns |

---

## Costs and Risks

| Cost | Impact | Mitigation |
|------|--------|------------|
| Schema migration | One-time `ALTER TABLE` + full FTS rebuild | Simple SQL migration. No data loss risk. |
| Cache coupling | gospel-mcp depends on gospel-vec cache format | JSON format is simple and stable. Document the contract. |
| FTS rebuild | Adding columns to FTS requires drop-and-recreate | One-time cost during `index` command. Automatic. |
| Parameter expansion | Search tool gets more parameters | All new params optional. Existing callers unaffected. |

**What gets worse:** gospel-mcp indexer becomes slightly more complex (reads cache files in addition to markdown). This is the right trade-off — complexity in the indexer rather than in the runtime query path.

**What if it goes wrong:** TITSW columns are nullable. If import fails for any talk, it gracefully degrades to NULL fields. Search without TITSW filters continues working exactly as before.

---

## Creation Cycle Review

| Step | Question | This Proposal |
|------|----------|---------------|
| Intent | Why? | Structured search on teaching dimensions. |
| Covenant | Rules? | gospel-mcp conventions. No LLM in gospel-mcp. |
| Stewardship | Who? | dev agent. gospel-mcp codebase. |
| Spiritual Creation | Spec precise? | Yes — schema, import path, tool params all specified. |
| Line upon Line | Phasing? | 3 phases. Phase 1 stands alone. |
| Physical Creation | Who executes? | dev agent after enriched-indexer Phase 1 ships. |
| Review | How to verify? | SQL queries + MCP tool calls. |
| Atonement | If wrong? | NULL fields. Backward compatible. |
| Sabbath | When to pause? | After Phase 1 — check data quality before wiring into search. |
| Consecration | Who benefits? | Michael + brain app + any MCP consumer. |
| Zion | Integration? | Completes the loop: gospel-vec generates → gospel-mcp searches. |

---

## Recommendation

**Build after enriched-indexer Phase 1 ships.** The import depends on the enriched summary cache existing. Once we have TITSW-enriched talk summaries in gospel-vec, this is a clean 3-session build. Phase 1 (schema + import) is the critical path — it proves the data flows correctly. Phases 2-3 are pure tool enhancement.

**Architectural note:** gospel-vec remains the LLM-powered indexer. gospel-mcp remains the structured search engine. This proposal extends their natural roles rather than blurring them. The cache file acts as the contract between them.

**Hand off to:** dev agent, after enriched-indexer Phase 1 validation.
