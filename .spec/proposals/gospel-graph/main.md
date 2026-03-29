# study.ibeco.me — Gospel Graph Visualization

**Binding problem:** Scripture cross-references and study connections exist in structured data (footnotes, BYU Citation Index, our own studies) but are invisible during reading. The reading experience is linear when the content is a graph. You can't see the web of connections from a passage, explore them interactively, or trace how your own studies fit into the canonical cross-reference network.

**Created:** 2026-03-29
**Reworked:** 2026-03-29 — changed from ibeco.me extension to standalone site
**Research:** [.spec/scratch/gospel-graph/main.md](../../scratch/gospel-graph/main.md)
**Depends on:** Enriched indexer + enriched search completion
**Status:** Proposed — new workstream, build after enriched indexer/vector pipeline ships

---

## 1. Problem Statement

The standard works contain thousands of explicit cross-references — footnotes, Topical Guide entries, Bible Dictionary links, JST references. The BYU Citation Index adds thousands more: which conference talks cite which verses. Our own study documents contain hundreds of scripture references. And gospel-vec can surface semantic connections that no index makes explicit.

All of this is invisible during reading. You can't see the web of connections radiating from a passage, explore that web interactively, or discover unexpected links between your studies and the canonical text.

This is a new site — **study.ibeco.me** — purpose-built for interactive gospel study with graph visualization. Not an extension of ibeco.me's becoming/practice app. Separate codebase, separate database, separate deployment. ibeco.me stays focused on becoming and daily practices. study.ibeco.me is for deep reading and exploration.

**Who's affected:** Michael — and potentially anyone interested in visual scripture exploration.

**How would we know it's fixed:** Open a scripture chapter in the study reader. Click a paragraph. See a visual graph showing every cross-reference, citation, and study connection for that passage. Click a node in the graph. Read the connected content. Navigate the graph like a map. Browse the BYU Citation Index with tabs, history, and back button.

---

## 2. Success Criteria

1. **study.ibeco.me exists** as a standalone site with its own reader and graph visualization
2. **Reader pane** shows scripture chapters, conference talks, and our study documents in markdown
3. **Graph pane** shows an interactive visualization of connections for the current passage
4. **Click a paragraph** in the reader → graph focuses on that passage's connections and "explodes out" to related content
5. **Click a graph node** → loads that content in the reader
6. **BYU Citation Index browsable** — with tabs, history (back button), details/summaries, links to full content
7. **Enriched metadata visible** — TITSW scores, teaching modes, and thematic edges appear as graph node properties
8. **Works locally and deployed** — runs on NOCIX alongside ibeco.me, separate container

---

## 3. Constraints & Boundaries

### In scope
- New standalone site: study.ibeco.me
- New Go backend with PostgreSQL (graph-optimized schema)
- New Vue 3 + TypeScript frontend (reader + graph visualization)
- Data imported from gospel-mcp (cross-references, scriptures, talks), gospel-vec (enriched metadata, semantic edges), and BYU citations (talk→scripture edges)
- Graph visualization library (Cytoscape.js or similar)
- Scripture reader with markdown rendering
- BYU Citation Index browsing with navigation state

### Explicitly NOT in scope
- Modifying ibeco.me — it stays focused on becoming/practices
- Auth system for study.ibeco.me (start without auth; add later if needed)
- LLM inference at query time — all edges are pre-computed at index time
- Mobile app — web only
- Real-time collaboration or multi-user features

### Conventions
- Go backend (chi router, same patterns as ibeco.me and MCP servers)
- Vue 3 + Vite + TypeScript frontend
- PostgreSQL for graph data (native recursive CTEs, array types, JSONB for metadata)
- Graph library: Cytoscape.js (best layout algorithms for dense reference graphs; vue-cytoscape wrapper). Fallback: D3.js force.
- Data flows one direction: gospel-mcp/gospel-vec/byu-citations → study.ibeco.me PostgreSQL (import pipeline, not live queries)
- Deployed via Dokploy on NOCIX alongside ibeco.me

---

## 4. Prior Art & Related Work

| Source | Relevance |
|--------|-----------|
| [gospel-mcp cross_references table](../../../scripts/gospel-mcp/internal/db/schema.sql) | Scripture-to-scripture graph with bidirectional indexes — primary data source |
| [Enriched indexer proposal](../enriched-indexer.md) | TITSW metadata pipeline — feeds graph node properties (modes, scores, patterns) |
| [Enriched search proposal](../enriched-search.md) | Option C: gospel-mcp imports gospel-vec cache. Same data flows into study.ibeco.me |
| [Scratch: "Gospel-comb" (Idea 5)](../../scratch/lm-studio-model-experiments/main.md) | Unified vec + SQLite tool — this proposal IS the realization, as a web app instead of an MCP tool |
| [Debug TITSW scratch — 3 consumers](../../scratch/debug-titsw-optimization/main.md) | Identified "knowledge graph" as a consumer needing structured labels and relationships |
| [ibeco.me ReaderView](../../../scripts/becoming/frontend/src/views/ReaderView.vue) | Patterns to learn from: scripture detection, markdown rendering, reference panel tabs |
| [TTS/STT reader proposal (deferred)](../deferred/tts-stt-reader.md) | Reader surface architecture concepts |
| [byu-citations MCP](../../../scripts/byu-citations/internal/citations/client.go) | BYU Citation Index scraping — data source for talk→scripture edges |

### Key insight
The "gospel-comb" revisit condition was: *"Revisit when model experiments produce a clear winner AND conference reindex succeeds."* Model experiments have produced a winner (nemotron-3-nano). study.ibeco.me is the gospel-comb vision realized as a web application — unifying structured search (gospel-mcp), semantic search (gospel-vec), citation data (byu-citations), and enriched metadata (TITSW) into one explorable interface.

---

## 5. Proposed Approach

### Architecture: Standalone site with PostgreSQL

A new Go + Vue 3 application at `scripts/study-site/`. Own PostgreSQL database on NOCIX. Data imported from existing tools via an import pipeline.

**Why PostgreSQL (not SQLite):**
- Native recursive CTEs for multi-hop graph traversal
- JSONB columns for flexible metadata (TITSW scores, enriched fields) without schema migrations per new field
- Array types for multi-value fields (dominant dimensions, keywords)
- Same database Michael already runs for ibeco.me — no new infrastructure, just a new database
- Better suited for a deployed web app than SQLite file locking

**Why separate from ibeco.me:**
- ibeco.me is a becoming/practice tracker. study.ibeco.me is a study/exploration tool. Different purposes, different data models.
- Avoids verification burden of modifying a working application
- Can evolve independently — different release cadence, different feature priorities
- Cleaner mental model: "ibeco.me is where I track my life, study.ibeco.me is where I explore scripture"

**Data flow:**
```
Import pipeline (offline, periodic):
  gospel-mcp SQLite → cross_references, scriptures, talks, chapters → study PostgreSQL
  gospel-vec cache  → enriched summaries, TITSW metadata            → study PostgreSQL
  byu-citations     → talk→scripture citation edges                  → study PostgreSQL
  study/ markdown   → study→scripture reference edges                → study PostgreSQL

Runtime:
  User opens study.ibeco.me
    → Browses/searches for a chapter or talk
    → Reader pane renders the content (markdown)
    → Graph pane shows connections for the current passage
    → Click paragraph → graph focuses, explodes connections
    → Click graph node → reader navigates to that content
    → Browse BYU citations with tabs, history, back button
```

### Backend (Go)

**`scripts/study-site/`** — new Go module

1. **`cmd/server/main.go`** — HTTP server (chi router, embedded frontend, PostgreSQL)
2. **`cmd/import/main.go`** — Import pipeline (reads gospel-mcp SQLite + gospel-vec cache + byu-citations → writes PostgreSQL)
3. **`internal/graph/`** — Graph query logic
   - `GetNeighborhood(ref string, hops int, edgeTypes []string) Graph` — N-hop traversal with edge type filtering
   - `GetCitations(ref string) []Citation` — BYU citation index for a reference
   - `Search(query string, filters GraphFilters) []Node` — full-text + metadata search
4. **`internal/api/`** — HTTP handlers
   - `GET /api/content/:path` — scripture/talk/study content (markdown)
   - `POST /api/graph/neighborhood` — graph neighborhood for references
   - `GET /api/citations/:ref` — BYU citation data
   - `GET /api/search` — combined search
   - `GET /api/stats` — graph metadata

### PostgreSQL Schema (core tables)

```sql
-- Nodes: every addressable piece of content
CREATE TABLE nodes (
    id          TEXT PRIMARY KEY,       -- "ot/gen/1:1", "talk/2024/04/kearon-joy", "study/charity"
    type        TEXT NOT NULL,          -- 'scripture', 'talk', 'study', 'tg', 'bd'
    label       TEXT NOT NULL,          -- "Genesis 1:1", "Elder Kearon, Apr 2024"
    volume      TEXT,                   -- "ot", "nt", "bofm", "dc", "pgp"
    content     TEXT,                   -- full markdown content
    metadata    JSONB DEFAULT '{}',     -- extensible: titsw scores, keywords, summaries
    file_path   TEXT                    -- path in gospel-library or study/
);

-- Edges: every connection between nodes
CREATE TABLE edges (
    id          SERIAL PRIMARY KEY,
    source_id   TEXT NOT NULL REFERENCES nodes(id),
    target_id   TEXT NOT NULL REFERENCES nodes(id),
    type        TEXT NOT NULL,          -- 'footnote', 'tg', 'bd', 'gs', 'jst', 'citation', 'study', 'semantic'
    label       TEXT,                   -- "1a", speaker name, study title
    weight      REAL DEFAULT 1.0,      -- edge strength (semantic similarity, citation count)
    metadata    JSONB DEFAULT '{}'
);

CREATE INDEX idx_edges_source ON edges(source_id);
CREATE INDEX idx_edges_target ON edges(target_id);
CREATE INDEX idx_edges_type ON edges(type);
CREATE INDEX idx_nodes_type ON nodes(type);
CREATE INDEX idx_nodes_volume ON nodes(volume);

-- Full-text search
CREATE INDEX idx_nodes_content_fts ON nodes USING gin(to_tsvector('english', content));
```

### Frontend (Vue 3 + TypeScript)

1. **`views/ReaderView.vue`** — Split-pane: reader (left) + graph (right)
   - Markdown rendering (markdown-it)
   - Scripture reference detection in rendered content
   - Paragraph click → graph focus
2. **`components/GraphPanel.vue`** — Cytoscape.js interactive graph
   - Force-directed layout (fcose)
   - Nodes color-coded by volume/type
   - Edges styled by type (footnote, citation, study, semantic)
   - Click node → navigate reader
   - Expand/collapse neighborhoods
   - Zoom, pan, fit controls
3. **`components/CitationBrowser.vue`** — BYU Citation Index interface
   - Tabbed browsing with history stack (back/forward)
   - Citation details: speaker, title, year, summary
   - Links to full content in reader
4. **`components/SearchPanel.vue`** — Combined search
   - Full-text search + TITSW filters
   - Results show in graph or list view
5. **`views/GraphExplorer.vue`** — Full-page graph mode
   - Open-ended exploration without starting from a specific passage

### Graph data types

```typescript
interface GraphNode {
  id: string           // "ot/gen/1:1" or "talk/2024/04/kearon-joy"
  label: string        // "Genesis 1:1" or "Elder Kearon, Apr 2024"
  type: 'scripture' | 'talk' | 'study' | 'tg' | 'bd'
  volume?: string      // "ot", "nt", "bofm", "dc", "pgp"
  metadata?: {
    titsw_mode?: string
    titsw_dominant?: string[]
    titsw_teach?: number
    summary?: string
    keywords?: string[]
    [key: string]: unknown
  }
}

interface GraphEdge {
  source: string
  target: string
  type: 'footnote' | 'tg' | 'bd' | 'gs' | 'jst' | 'citation' | 'study' | 'semantic'
  label?: string
  weight?: number
}
```

---

## 6. Phased Delivery

**This workstream starts AFTER the enriched indexer + enriched search pipeline is complete.** That pipeline produces the data this site consumes. Building before the data is ready means building with one hand tied behind our back.

### Prerequisite: Enriched Indexer + Search (separate workstream)

| Step | What it produces for study.ibeco.me |
|------|-------------------------------------|
| Enriched indexer Phase 1 (talks) | TITSW metadata on all 5,500 talks → graph node properties |
| Enriched indexer Phase 2 (scripture) | Deeper scripture keywords, typological connections → richer nodes |
| Enriched indexer Phase 3 (manuals + themes) | Manual summaries, talk theme detection → more edge types |
| Enriched search Phase 1 (schema + import) | gospel-mcp gets TITSW columns → import pipeline reads structured data |

### Phase 1: Foundation — Reader + Basic Graph (2-3 sessions)

**Scope:** Standalone site with scripture reader and cross-reference graph visualization.

| Deliverable | Detail |
|-------------|--------|
| `scripts/study-site/` | New Go module with server + import commands |
| PostgreSQL schema | nodes + edges tables with indexes |
| Import pipeline | Reads gospel-mcp SQLite → populates nodes and edges |
| Reader pane | Markdown rendering of scripture chapters |
| Graph pane | Cytoscape.js with cross-reference edges, color-coded by volume |
| Split-pane layout | Reader left, graph right |
| Click interactions | Paragraph → graph focus. Node → reader navigation. |

**Verification:**
1. Run import against gospel-mcp SQLite
2. Open study.ibeco.me locally
3. Browse to a Book of Mormon chapter
4. Click a verse paragraph → graph shows cross-references
5. Click a cross-reference node → reader loads that chapter
6. TG/BD entries appear as labeled nodes

**Stands alone:** Yes. Reader + graph with existing cross-reference data.

### Phase 2: BYU Citations + Talk Nodes (1-2 sessions)

**Scope:** Import BYU citation data. Add conference talk nodes and citation edges to the graph.

| Deliverable | Detail |
|-------------|--------|
| Citation import | Bulk scrape + cache BYU citation data into PostgreSQL |
| Talk nodes | Conference talks as graph nodes with metadata |
| Citation edges | talk→scripture edges in graph |
| CitationBrowser component | Tabbed browsing with history, details, links |
| Graph update | Talk nodes styled distinctly, clickable |

**Verification:** Click John 3:16. See footnote edges AND conference talks that cite it. Open CitationBrowser. Browse tabs, go back, go forward. Click a talk → read it.

### Phase 3: Study Documents + Search (1-2 sessions)

**Scope:** Import our study documents. Add full-text + metadata search.

| Deliverable | Detail |
|-------------|--------|
| Study document import | Parse `study/` markdown for scripture references, store as nodes and edges |
| Search panel | Full-text search across all content |
| Search → graph | Search results appear as graph nodes or list |
| Study edges in graph | Our studies connected to the passages they reference |

**Verification:** Search for "charity." See study documents AND scriptures. Click a study → read it. Graph shows study↔scripture connections.

### Phase 4: Enriched Metadata + Thematic Edges (1-2 sessions)

**Scope:** Import TITSW enriched data. Add semantic similarity edges. Make metadata visible.

| Deliverable | Detail |
|-------------|--------|
| TITSW metadata import | Read gospel-mcp TITSW columns → populate node metadata JSONB |
| TITSW filters | Filter graph/search by teaching mode, score ranges, dominant dimensions |
| Semantic edges | Import gospel-vec similarity data as weighted edges |
| Node detail panel | Show TITSW scores, summary, keywords when hovering/clicking |

**Depends on:** Enriched indexer + enriched search complete.

**Verification:** Filter for "enacted love" talks. See only talks where mode=enacted and dominant includes love. Hover a talk node → see TITSW scores and summary.

### Phase 5: Multi-hop Exploration + Deploy (1-2 sessions)

**Scope:** Make the graph deeply explorable. Production deployment.

| Deliverable | Detail |
|-------------|--------|
| N-hop traversal | Expand neighborhoods progressively (1-3 hops) |
| Expand/collapse | Click to expand a node's connections, collapse to reduce clutter |
| Navigation history | Back/forward through graph states |
| Full-page graph mode | GraphExplorer view for open-ended exploration |
| Deploy to NOCIX | Dokploy container, study.ibeco.me subdomain |

**Verification:** Start at one verse. Expand 2 hops. Navigate back. Forward. Collapse a cluster. Full-page graph mode works. Deployed at study.ibeco.me.

---

## 7. Enriched Indexer Compatibility Checklist

These are the specific things to verify as the enriched indexer ships, to ensure study.ibeco.me can consume the data:

| Enriched Indexer Output | study.ibeco.me Needs | Compatible? |
|------------------------|---------------------|-------------|
| gospel-vec summary cache JSON (TEACHING_PROFILE fields) | Import into nodes.metadata JSONB | **Yes** — JSON → JSONB is direct |
| gospel-mcp titsw_* columns (via enriched search Phase 1) | Import from gospel-mcp SQLite | **Yes** — structured columns map to JSONB fields |
| gospel-mcp cross_references table | Import as edges | **Yes** — already the primary edge source |
| gospel-vec semantic search (chromem-go) | Pre-compute similarity edges at import time | **Needs work** — gospel-vec has no "export all similarities" API. Options: (1) add `export-similarities` command to gospel-vec, (2) read gob.gz files directly and compute cosine similarity, (3) skip semantic edges initially. Recommend option 1. |
| BYU citation data | Import as citation edges | **Yes** — bulk scrape via byu-citations MCP |
| Enriched keywords (deeper scripture keywords from Phase 2) | Import into nodes.metadata | **Yes** — same JSON format |
| Talk theme detection (Phase 3) | Import as edge metadata or node metadata | **Yes** — additional JSONB fields |

**Action item for enriched indexer:** No changes needed to the existing plan. The cache format (JSON with prompt_version, TEACHING_PROFILE fields) and gospel-mcp schema (titsw_* columns) are directly importable. One addition to consider during enriched indexer Phase 2: a bulk similarity export tool for gospel-vec.

---

## 8. Verification Strategy

| Phase | Verification |
|-------|-------------|
| 1 | Import edge count matches gospel-mcp cross_references count. 5 chapters across all standard works: graph edges match printed footnotes. |
| 2 | 10 well-cited verses: citation count matches BYU website (±1). CitationBrowser back/forward works. |
| 3 | 5 study documents: all scripture references appear as edges. Search returns relevant results. |
| 4 | TITSW filters produce correct subsets. Metadata visible on node hover. |
| 5 | 3-hop traversal stays performant. Deployed site accessible at study.ibeco.me. |

---

## 9. Costs & Risks

### Costs
- **Development time:** Full build is 7-11 sessions across 5 phases.
- **New codebase:** New Go module, new Vue app, new PostgreSQL database.
- **Infrastructure:** New Dokploy container on NOCIX. PostgreSQL already runs there (shared with ibeco.me).
- **Bundle size:** Cytoscape.js ~170KB gzipped.

### Risks
| Risk | Mitigation |
|------|-----------|
| Scope: standalone site is a big build | Strict phasing. Phase 1 is reader + basic graph only. |
| Data sync: gospel-mcp changes → must re-import | Import is idempotent. Run after any reindex. |
| Graph hairball at scale | Progressive disclosure, LOD, neighborhood-based loading |
| Delayed start (waits for enriched indexer) | Planning is done now — ready to build when data is ready. |
| PostgreSQL complexity vs. SQLite simplicity | PostgreSQL already runs on NOCIX. JSONB + recursive CTEs justify it. |

### What gets worse
- One more application to deploy and maintain
- Import pipeline needs to run after every gospel-mcp or gospel-vec reindex

### What gets better
- ibeco.me stays clean and focused — no verification burden
- study.ibeco.me can evolve independently
- Clean data model designed for graph queries from the start
- No risk of breaking ibeco.me

---

## 10. Creation Cycle Review

| Step | Question | This proposal |
|------|----------|--------------|
| **Intent** | Why are we doing this? | Make the invisible connections in scripture visible and explorable. Directly serves the project's root intent: deep, honest scripture study. |
| **Covenant** | Rules of engagement? | Go + Vue 3 conventions. Read-before-quoting for verification. Source data from gospel-mcp + gospel-vec. |
| **Stewardship** | Who owns what? | New codebase: `scripts/study-site/`. dev agent builds. Michael reviews and tests. gospel-mcp/gospel-vec data: owned by existing tools, consumed read-only. |
| **Spiritual Creation** | Is the spec precise enough? | Schema, API shape, data flow, component structure, verification criteria all defined. Phase 1 is buildable. |
| **Line upon Line** | What's the phasing? | 5 phases after enriched indexer prerequisite. Phase 1 stands alone. |
| **Physical Creation** | Who executes? | dev agent builds. Sequenced AFTER enriched indexer + enriched search. |
| **Review** | How do we know it's right? | Edge counts match source data. BYU citations match website. TITSW filters produce correct subsets. |
| **Atonement** | What if it goes wrong? | Separate site — ibeco.me unaffected. Worst case: delete the container. Source data is never modified. |
| **Sabbath** | When do we stop and reflect? | After Phase 1 — does the reader + graph actually improve study? After Phase 4 — does enriched metadata make the graph meaningfully better? |
| **Consecration** | Who benefits? | Michael first. Potentially anyone interested in visual scripture study. |
| **Zion** | How does this serve the whole? | Unifies gospel-mcp, gospel-vec, and byu-citations into one explorable interface. The "gospel-comb" vision realized as a web application. |

---

## 11. Recommendation

**Build as a new workstream, sequenced after enriched indexer + enriched search.**

The enriched indexer pipeline produces the data that makes this site sing — TITSW metadata, deeper keywords, thematic edges. Building before that data exists means building with one hand tied behind our back.

**Sequencing:**
1. **Now:** Enriched indexer Phase 1 (talk batch enrichment) — already next in queue
2. **Then:** Enriched indexer Phases 2-3 + enriched search Phase 1
3. **Then:** study.ibeco.me Phase 1 (reader + basic graph with fully enriched data)
4. **Then:** Phases 2-5 sequentially

**Add as Workstream G (Graph) in overview:**

| Phase | Scope | Depends on |
|-------|-------|------------|
| Phase 1 | Reader + cross-reference graph | Enriched search Phase 1 complete |
| Phase 2 | BYU citations + talk nodes | Phase 1 |
| Phase 3 | Study documents + search | Phase 1 |
| Phase 4 | Enriched metadata + thematic edges | Enriched indexer Phases 1-3 |
| Phase 5 | Multi-hop exploration + deploy | Phases 1-4 |

**What to watch during enriched indexer work:**
- Verify the compatibility checklist (Section 7) as each enriched indexer phase ships
- If gospel-vec needs a bulk similarity export, spec it during enriched indexer Phase 2
- Keep the import pipeline in mind when making data format decisions — study.ibeco.me is a downstream consumer
