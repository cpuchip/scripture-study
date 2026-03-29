# Gospel Graph Visualization — Research & Scratch

*Created: 2026-03-29*
*Proposal: [.spec/proposals/gospel-graph/main.md](../../proposals/gospel-graph/main.md)*

---

## Binding Problem

Scripture cross-references and study connections exist in structured data (footnotes, BYU Citation Index, our own studies) but are invisible during reading. The reading experience is linear when the content is a graph. You can't see the web of connections from the passage you're in, explore them interactively, or trace how your own studies relate to the canonical cross-references.

---

## Phase 2: Research Findings

### A. ibeco.me Architecture (Existing Reader)

**Stack:** Vue 3 + Vite frontend, Go + chi router backend, PostgreSQL + SQLite
**Deployment:** NOCIX server via Dokploy, Google OAuth + session auth

**Relevant existing components:**
- `ReaderView.vue` — Full markdown reader with file tree sidebar, search, reading progress, bookmarks
- `ReferencePanel.vue` — Tabbed reference panel with iframe embedding (church URLs) + markdown rendering
- `TreeNode.vue` — File tree navigation
- `github.ts` — Fetches content from GitHub repos (our study files)
- `scripture.go` — Scripture reference parsing (verse ranges, multi-references)
- `api.ts` — Full API client (~200 lines, typed)
- `/read/:id` — Reader route with source selection, file tree, content pane, reference panel
- `/sources` — GitHub repo sources with include/exclude globs
- `/s/:code` — Public shared reader links

**Key observations:**
- Reader already detects scripture URLs (`isScriptureUrl`) and opens them in ReferencePanel iframes
- ReferencePanel already handles tabs, loading states, error states, markdown-it rendering
- Reader already has heading anchors, bookmarks, "become" task creation from content
- The reference panel is the natural attachment point for a graph view tab
- File tree navigation + content rendering + reference handling already works end-to-end

### B. Gospel Data Infrastructure

#### gospel-mcp (SQLite + FTS5)
- **Tables:** scriptures (verse-level), chapters, talks, manuals, books, cross_references
- **cross_references table:** source_volume/book/chapter/verse → target_volume/book/chapter/verse, reference_type (footnote/tg/bd/jst)
- **Indexes:** Bidirectional (can query "what references THIS verse" not just "what does this verse reference")
- **FTS5:** On scriptures, talks, manuals, books
- **Tools:** gospel_search, gospel_get, gospel_list
- **NOTE:** cross_references table IS the scripture-to-scripture graph with bidirectional indexes

#### gospel-vec (chromem-go vector search)
- **Layers:** verse, paragraph, summary, theme
- **Storage:** Per-source gob.gz files (NOT SQLite — pure vector DB)
- **Metadata:** Source, Layer, Book, Chapter, Reference, Range, FilePath, Generated, Model, Speaker, Year, Month, Session, TalkTitle
- **Tools:** search_scriptures, list_books, get_talk, search_talks
- **NOTE:** No relational data, no edges — semantic similarity only

#### byu-citations (HTTP scraper)
- **Source:** scriptures.byu.edu HTML scraping
- **Returns:** Citation structs (Reference, Speaker, Title, TalkID, RefID)
- **Tools:** byu_citations, byu_citations_bulk, byu_citations_books
- **NOTE:** Live scraping — not cached locally. Each call is an HTTP request.

#### Scripture markdown format
- Bold verse numbers, superscript footnote anchors, cross-reference links as relative markdown paths
- ~1,500 chapter files across 5 standard works
- TG/BD/GS in study aids directories

### C. Prior Art & Related Proposals

#### Enriched Indexer (.spec/proposals/enriched-indexer.md)
- Phase 0 COMPLETE (18 runs, T4 calibration won at MAE 1.83)
- Adds TEACHING_PROFILE to gospel-vec talk summaries (modes, patterns, 6 TITSW scores)
- Phase 1 not yet started — batch enrichment of all talks
- **Relevance:** Enriched metadata would be graph node properties (teaching mode, dominant principle, etc.)

#### Enriched Search (.spec/proposals/enriched-search.md)
- Option C chosen: gospel-mcp reads gospel-vec cache
- Not started. Depends on enriched-indexer Phase 1.
- **Relevance:** Would add TITSW columns to gospel-mcp's SQLite → queryable from graph API

#### TTS/STT Reader (.spec/proposals/deferred/tts-stt-reader.md)
- DEFERRED. Contains relevant web reader architecture (audio player, /read route)
- **Relevance:** Its "Web Reader" phase describes a similar reader surface

#### Scratch notes — Ideas 3 & 5 (lm-studio-model-experiments)
- **Idea 3 "Graph edges in SQLite":** ALREADY IDENTIFIED. Notes: cross_references exists but MISSING multi-hop traversal, talk→scripture edges, LLM-inferred thematic edges, study→scripture edges, graph visualization.
- **Idea 5 "Gospel-comb":** Unified vec + SQLite tool. Deferred. Options: add SQLite to gospel-vec, add vectors to gospel-mcp, or new tool wrapping both.
- **Deferred with condition:** "Revisit when model experiments produce a clear winner AND conference reindex succeeds" — model experiments ARE now complete.

#### Debug TITSW scratch
- Identified three consumers: semantic search, structured queries, knowledge graph
- Knowledge graph consumer described as needing "structured labels and relationships"

---

## Phase 3: Gap Analysis

### What EXISTS (assets we can build on)

| Asset | What it provides for the graph |
|-------|-------------------------------|
| cross_references table | Scripture↔scripture graph edges (footnotes, TG, BD, GS, JST). Bidirectional indexes. ~thousands of edges across all 5 standard works. |
| gospel-mcp FTS5 | Full-text search of all scripture/talk/manual/book content |
| gospel-vec semantic search | "Find similar" capability — thematic similarity without explicit edges |
| byu-citations | Conference talk → scripture citation edges (live scraping) |
| ibeco.me ReaderView | Full reader with markdown rendering, file tree, reference panel, scripture detection |
| ReferencePanel | Tabbed viewer with iframe + markdown — pattern for additional graph tab |
| ibeco.me Go backend | Auth, API, deployment pipeline, PostgreSQL, session management |
| Scripture markdown files | ~1,500 chapter files with structured footnote anchors |
| Study documents | Our own studies with scripture references embedded in text |
| Enriched indexer (Phase 0) | TITSW scoring framework validated — node metadata pipeline |

### What's MISSING (gaps to fill)

| Gap | Impact | Difficulty |
|-----|--------|------------|
| **Graph visualization** | Core feature — nothing exists | HIGH — need JS graph library, data API, interactive UX |
| **Graph query API** | No way to request "all edges for verse X" | MEDIUM — extend gospel-mcp or new service |
| **Talk→scripture citation edges** | Can't see which TALKS cite a verse (only via live byu-citations scraping) | MEDIUM — need to cache/index BYU citation data |
| **Multi-hop traversal** | Can't ask "what's 2 hops from John 3:16" | MEDIUM — recursive CTE in SQLite or app-level BFS |
| **Study→scripture edges** | Can't see which of OUR studies reference a passage | MEDIUM — parse study markdown for scripture refs |
| **LLM-inferred thematic edges** | Only explicit footnote edges exist, not thematic relationships | HIGH — needs LLM pipeline, quality control, cost |
| **Unified query layer** | gospel-mcp (structured) and gospel-vec (semantic) are separate services | MEDIUM — facade or merged service |
| **Graph data format** | No standardized node/edge schema for visualization | LOW — design decision |
| **Frontend graph component** | No graph visualization in ibeco.me | HIGH — need library selection, integration, interactivity |
| **"Explode from paragraph" interaction** | Click paragraph → show connections in graph | MEDIUM — needs paragraph→scripture reference extraction at render time |
| **Navigation: tabs, history, back** | BYU citation index style browsing | MEDIUM — state management in Vue |

### Under-researched areas

1. **Graph visualization libraries** — Need to evaluate: D3.js force-directed, vis.js, Cytoscape.js, Sigma.js, react-force-graph. Key criteria: Vue 3 compatibility, large graph performance (10K+ nodes), interactive layout, click/hover/focus, mobile-friendly.
2. **BYU citation data volume** — How many talk→scripture edges exist? Is bulk scraping feasible or do we need a crawl+cache strategy?
3. **cross_references table edge count** — How many edges are currently indexed? Determines graph density.
4. **Gospel Library path→content resolution** — Can we render gospel-library markdown in the graph panel, or do we need church website iframes?
5. **Performance at scale** — A fully-connected graph of all scripture cross-references could be thousands of nodes. Need LOD (level-of-detail) or progressive disclosure.

---

## Phase 3a: Critical Analysis

### 1. Is this the RIGHT thing to build, or just the EXCITING thing?

**Assessment: Both, but needs scoping.**

The vision is genuinely exciting AND solves a real problem. Scripture cross-references ARE a graph. Reading IS linear. The BYU Citation Index IS useful but terrible to browse. These are real pain points.

BUT — the full vision (study.ibeco.me, graph visualization, BYU citation browsing, talk connections, thematic edges, semantic search integration, deployable web service) is MASSIVE. This touches:
- Frontend: New Vue component or page (graph visualization library)
- Backend: New or extended API (graph queries)
- Data: New edge types (talk→scripture, study→scripture, thematic)
- Infrastructure: New subdomain, deployment
- Dependencies: Enriched indexer (for node metadata), gospel-mcp schema changes

This is not a weekend project. Full vision is weeks of work across multiple codebases.

### 2. Does this solve the binding problem, or a different one?

**It solves the stated binding problem** — "can't see the web of connections from the passage you're in." The graph visualization directly addresses this.

BUT the vision also includes elements that are SEPARATE problems:
- BYU Citation Index browsing → this is a UX/data access problem, not strictly a graph problem
- "New site study.ibeco.me" → deployment architecture, not core feature
- Tabs/history/back button → standard navigation, not graph-specific

The binding problem is: **make the invisible connections visible and explorable.** The rest is UX sugar (important, but not the core).

### 3. What's the simplest version that would be useful?

**Minimum viable graph: Add a "Graph" tab to ibeco.me's existing ReferencePanel.**

When reading a passage:
1. Parse the visible content for scripture references
2. Query gospel-mcp's cross_references table for all edges from those references
3. Render a small force-directed graph showing: current verse (center) → footnote targets → their footnote targets (2 hops)
4. Click a node → load that passage in the reader

That's it. No new service. No new site. No BYU citations. No thematic edges. Just: **show me what the footnotes connect to, visually.**

This could be built in 1-2 sessions by extending existing infrastructure:
- Add a `/api/graph/edges` endpoint to ibeco.me backend (queries gospel-mcp SQLite directly or via MCP)
- Add a `GraphPanel.vue` component using a lightweight graph library
- Add a "Graph" tab to ReferencePanel
- Wire the click handler to load content

### 4. What gets WORSE if we build this?

- **Cognitive load:** Another feature in ibeco.me to maintain
- **Performance:** Graph rendering on every paragraph click could be slow if not debounced/cached
- **Data dependency:** Graph quality depends on cross_references table completeness — if footnote extraction is buggy, graph shows wrong connections
- **Scope creep magnet:** Once the basic graph works, the pull to add BYU citations, thematic edges, semantic search, etc. will be very strong. Need discipline to stop.
- **ibeco.me complexity:** The becoming app is already substantial. Adding graph visualization moves it further from "simple practice tracker" toward "comprehensive gospel study platform."

### 5. Does this duplicate something we already have?

**Partially.** The ReferencePanel already shows cross-references when you click a scripture link — it loads the target in an iframe or markdown tab. The graph doesn't REPLACE this; it EXTENDS it by showing the full web of connections rather than one-at-a-time.

gospel-mcp's cross_references table already stores the graph data. We're not recreating data — we're visualizing it.

### 6. Is this the right time?

**Mixed signals.**

FOR now:
- Model experiments COMPLETE — that long-running workstream is done
- Enriched indexer Phase 0 COMPLETE — related data pipeline validated
- ibeco.me reader is mature — good foundation to build on
- The "gospel-comb" revisit condition ("model experiments produce a winner") is now met
- Michael was explicitly inspired during church — spiritual timing matters

AGAINST now:
- 7 priorities already in active.md (study, teaching, model experiments, debugging, WS1, desktop, server)
- Enriched indexer Phase 1 not yet started — this feeds graph node metadata
- Teaching workstream not started — Spirit-driven priority from Mar 23
- The overview shows 19 plans, 9 proposals, and multiple blocked/waiting items
- Adding a new workstream when existing ones have pending work

### 7. Mosiah 4:27 check

**Michael is stretched.** The active.md shows 7 priorities with substantial work remaining. The recent pace (TITSW experiments, context engineering, debug audit, phase 0 experiments — all in the Mar 28-29 window) suggests he's in a high-energy phase, but the teaching and enriched indexer work haven't started.

**Recommendation:** This idea deserves a SPEC, not immediate execution. The minimum viable version (graph tab in ReferencePanel) is small enough to fit, but ONLY if it doesn't displace the enriched indexer Phase 1 or teaching work. Sequence it AFTER enriched indexer Phase 1 — that way the graph has richer node data from the start.

### 8. Creation Cycle alignment

This is at the **Spiritual Creation** step — Michael received the vision, it needs to be specified precisely before physical creation begins. That's exactly what we're doing. The question is where it falls in the **Line upon Line** step — what's the phasing that delivers value incrementally without swallowing capacity?

---

## Key Architecture Decision: New Service vs. Extension

### Option A: New "gospel-graph" service
- Separate Go service wrapping gospel-mcp + gospel-vec + byu-citations
- Own SQLite database with graph-specific tables
- Deployed alongside ibeco.me on NOCIX
- PRO: Clean separation, purpose-built graph queries, independent evolution
- CON: Another service to maintain, deploy, monitor. Data duplication.

### Option B: Extend gospel-mcp
- Add graph query endpoints to gospel-mcp
- Add talk→scripture edges, multi-hop traversal to existing schema
- ibeco.me backend calls gospel-mcp (already in the ecosystem)
- PRO: Builds on existing infrastructure. Single source of truth for structured gospel data.
- CON: gospel-mcp is an MCP server, not an HTTP API. ibeco.me would need to either embed it or call it via MCP protocol.

### Option C: Extend ibeco.me backend directly
- Add graph query endpoints to ibeco.me Go backend
- Import gospel-mcp's SQLite database or connect to it
- Frontend talks to its own backend (existing pattern)
- PRO: Simplest integration path. Frontend ↔ backend already works. Auth already handled.
- CON: Couples gospel data logic into the becoming app. Harder to share with other consumers.

### Option D: gospel-mcp HTTP mode + ibeco.me frontend
- gospel-mcp adds HTTP API alongside MCP (like becoming server does)
- ibeco.me frontend or backend calls gospel-mcp HTTP endpoints
- PRO: gospel-mcp becomes the canonical gospel data API for all consumers. Clean separation.
- CON: gospel-mcp needs HTTP layer added. More work upfront. Two servers to deploy.

**Emerging recommendation: Option C for Phase 1.** Simplest path. ibeco.me backend adds a `/api/graph/edges` endpoint that reads gospel-mcp's SQLite DB file directly (read-only). No new service, no new deployment. If graph grows in complexity later, can extract to Option D.

This matches the existing pattern: ibeco.me's scripture.go already parses scripture references. The backend already has a database layer. Adding a graph query is incremental.

---

## Graph Visualization Library Evaluation (Quick)

| Library | Vue 3? | Force-directed? | Interactive? | Performance (10K nodes) | Bundle size |
|---------|--------|-----------------|--------------|------------------------|-------------|
| D3.js force | Yes (manual) | Yes | Yes (SVG events) | Good (canvas mode) | ~50KB |
| vis-network | Community wrapper | Yes | Excellent (built-in) | Good | ~180KB |
| Cytoscape.js | vue-cytoscape | Yes (CoSE) | Excellent | Very good | ~170KB |
| Sigma.js v2 | Manual | Yes (ForceAtlas2) | Good | Excellent (WebGL) | ~50KB |
| vue-force-graph | Native Vue 3 | Yes (d3-force) | Good | Good (3D with three.js) | ~30KB |
| @antv/g6 | Yes | Yes | Excellent | Excellent | ~200KB |

**Quick assessment:** For a graph tab in a panel, we need: small bundle, good interactivity, Vue 3 compatible, handles ~100-500 nodes (2-hop neighborhood). **D3.js force** or **vue-force-graph** are the lightweight choices. **Cytoscape.js** is the "batteries included" choice with better layout algorithms.

Recommendation: **Cytoscape.js** — it has the richest layout algorithms (important for readable scripture graphs), excellent interaction model, well-maintained, and the vue-cytoscape wrapper exists. If bundle size matters, D3.js force is the fallback.

---

## Phasing Sketch

### Phase 1: Scripture Cross-Reference Graph Tab (1-2 sessions)
- Add `/api/graph/edges` to ibeco.me backend (reads gospel-mcp SQLite)
- Add `GraphPanel.vue` component (Cytoscape.js or similar)
- Add "Graph" tab to ReferencePanel
- Click paragraph → detect scripture references → request edges → render graph
- Click node → load that scripture in reader
- **Value:** See the web of footnote connections for any passage, interactively

### Phase 2: BYU Citation Edges (1 session)
- Bulk-scrape and cache BYU citation data into gospel-mcp SQLite (new table: `talk_citations`)
- Add talk→scripture edges to graph API
- Graph nodes now include: scriptures + conference talks
- Click talk node → show talk metadata, link to content
- **Value:** See which talks cite a passage, from the graph

### Phase 3: Study Document Edges (1 session)
- Parse study/ directory markdown for scripture references
- Store study→scripture edges in SQLite
- Add study nodes to graph
- Click study node → open study in reader
- **Value:** See your own work connected to the scripture graph

### Phase 4: Multi-hop & Exploration (1 session)
- Recursive CTE for N-hop traversal
- Progressive disclosure (expand/collapse neighborhoods)
- Tab navigation with history (back button)
- Node detail panel (summary, metadata, links)
- **Value:** Explore the graph like a map, not just a static snapshot

### Phase 5: Thematic Edges (future — needs enriched indexer)
- LLM-inferred thematic connections
- Semantic similarity edges from gospel-vec
- Enriched node metadata (TITSW scores, teaching modes)
- **Value:** See connections the footnotes don't make explicit

### Phase 6: study.ibeco.me (future — if warranted)
- Separate subdomain/route for study-focused experience
- Full-page graph mode
- Search → graph → read workflow
- **Value:** Dedicated study tool, not a tab in a practice tracker

---

## Dependencies

```
Phase 1 → nothing (existing data suffices)
Phase 2 → BYU citation bulk scraping (new work)
Phase 3 → study document parsing (new work, small)
Phase 4 → Phase 1 working
Phase 5 → enriched indexer Phases 1-3 (separate workstream)
Phase 6 → Phases 1-4 proven, demand demonstrated
```

Phase 1 has NO blockers. It can start now with existing gospel-mcp data.
