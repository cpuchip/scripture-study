# Gospel Graph Visualization

**Binding problem:** Scripture cross-references and study connections exist in structured data (footnotes, BYU Citation Index, our own studies) but are invisible during reading. The reading experience is linear when the content is a graph. You can't see the web of connections from a passage, explore them interactively, or trace how your own studies fit into the canonical cross-reference network.

**Created:** 2026-03-29
**Research:** [.spec/scratch/gospel-graph/main.md](../../scratch/gospel-graph/main.md)
**Status:** Proposed — awaiting decision

---

## 1. Problem Statement

The standard works contain thousands of explicit cross-references — footnotes, Topical Guide entries, Bible Dictionary links, JST references. The BYU Citation Index adds thousands more: which conference talks cite which verses. Our own study documents contain hundreds of scripture references. And gospel-vec can surface semantic connections that no index makes explicit.

All of this is invisible during reading. The best you can do today is click a footnote and load one target at a time in ibeco.me's ReferencePanel iframe. There's no way to see the full web of connections radiating from a passage, explore that web interactively, or discover unexpected links between your studies and the canonical text.

**Who's affected:** Michael — and potentially anyone using ibeco.me for scripture study.

**How would we know it's fixed:** Open a scripture chapter in the reader. Click a paragraph. See a visual graph showing every cross-reference, citation, and study connection for that passage. Click a node in the graph. Read the connected content. Navigate the graph like a map.

---

## 2. Success Criteria

1. **Graph tab exists** in ibeco.me reader alongside the existing reference panel
2. **Click a paragraph** in the reader → graph renders showing that passage's connections
3. **Cross-reference edges visible** — footnote, TG, BD, GS, JST edges from gospel-mcp
4. **Click a graph node** → loads that content in the reader
5. **Readable layout** — graph is not a hairball; nodes are labeled, edges are typed, layout is force-directed with good defaults
6. **Works locally and deployed** — same deployment path as existing ibeco.me

---

## 3. Constraints & Boundaries

### In scope
- Graph visualization of existing cross-reference data
- Integration into ibeco.me's existing reader
- Backend API for graph queries using gospel-mcp's SQLite database
- Basic navigation: click node to read, expand/collapse neighborhoods

### Explicitly NOT in scope (Phase 1)
- New site (study.ibeco.me) — revisit after core graph proves valuable
- BYU Citation Index integration — Phase 2
- Study document → scripture edges — Phase 3
- LLM-inferred thematic edges — Phase 5 (depends on enriched indexer)
- Semantic similarity edges from gospel-vec — Phase 5
- Full-page graph mode — Phase 6
- Conference talk content in graph — Phase 2

### Conventions
- Go backend, Vue 3 + TypeScript frontend (existing ibeco.me stack)
- Graph library: Cytoscape.js (best layout algorithms for dense reference graphs; vue-cytoscape wrapper available). Fallback: D3.js force if bundle size is a concern.
- gospel-mcp's SQLite database is read directly by ibeco.me backend (read-only). No new service for Phase 1.
- Existing auth model applies — graph is available to authenticated users

---

## 4. Prior Art & Related Work

| Source | Relevance |
|--------|-----------|
| [gospel-mcp cross_references table](../../../scripts/gospel-mcp/internal/db/schema.sql) | The scripture-to-scripture graph already exists in SQLite with bidirectional indexes |
| [Scratch: "Graph edges in SQLite" (Idea 3)](../../scratch/lm-studio-model-experiments/main.md) | Previously identified Gap: multi-hop traversal, talk→scripture edges, visualization all missing |
| [Scratch: "Gospel-comb" (Idea 5)](../../scratch/lm-studio-model-experiments/main.md) | Unified vec + SQLite tool — deferred with condition now met (model experiments complete) |
| [Debug TITSW scratch — 3 consumers](../../scratch/debug-titsw-optimization/main.md) | Identified "knowledge graph" as a consumer needing structured labels and relationships |
| [Enriched indexer proposal](../enriched-indexer.md) | Phase 0 complete. TITSW metadata pipeline — will feed graph node properties in Phase 5 |
| [Enriched search proposal](../enriched-search.md) | Option C (gospel-mcp reads gospel-vec cache) — same read-only pattern this proposal uses |
| [ibeco.me ReaderView](../../../scripts/becoming/frontend/src/views/ReaderView.vue) | Existing reader with scripture detection, reference panel, markdown rendering |
| [ibeco.me ReferencePanel](../../../scripts/becoming/frontend/src/components/ReferencePanel.vue) | Tabbed panel — natural place to add a "Graph" tab |
| [TTS/STT reader proposal (deferred)](../deferred/tts-stt-reader.md) | Described a similar reader surface architecture |

### Key insight from prior work
The "gospel-comb" revisit condition was: *"Revisit when model experiments produce a clear winner AND conference reindex succeeds."* Model experiments have produced a winner (nemotron-3-nano). The graph visualization is the natural realization of the gospel-comb vision — but scoped to visualization first, not a unified query tool.

---

## 5. Proposed Approach

### Architecture: Option C — Extend ibeco.me backend

ibeco.me backend adds graph query endpoints that read gospel-mcp's SQLite database directly (read-only). No new service for Phase 1.

**Rationale:** Simplest integration path. Frontend ↔ backend already works. Auth already handled. ibeco.me's `scripture.go` already parses scripture references. Adding a graph query endpoint is incremental. If graph grows in complexity, can extract to a dedicated service later.

**Data flow:**
```
User clicks paragraph in ReaderView
  → Frontend extracts scripture references from paragraph text
  → POST /api/graph/edges { references: ["John 3:16", "1 Ne 11:33"] }
  → Backend parses references, queries cross_references table
  → Returns: { nodes: [...], edges: [...] }
  → GraphPanel.vue renders with Cytoscape.js
  → User clicks a node
  → Frontend loads that passage in reader
```

### Backend additions (Go)

1. **`internal/graph/graph.go`** — Graph query logic
   - `GetEdges(refs []string, hops int) ([]Node, []Edge)` — returns nodes and edges within N hops
   - Reads gospel-mcp's SQLite database (path configurable, read-only connection)
   - Uses recursive CTE for multi-hop traversal
   - Returns typed edges (footnote, tg, bd, gs, jst)

2. **`internal/api/graph.go`** — HTTP handlers
   - `POST /api/graph/edges` — given references, return graph neighborhood
   - `GET /api/graph/stats` — metadata about graph (edge count, node count per volume)

3. **Configuration** — Gospel-mcp SQLite path added to server config

### Frontend additions (Vue 3 + TypeScript)

1. **`components/GraphPanel.vue`** — Cytoscape.js graph renderer
   - Force-directed layout (CoSE or fcose for compound graphs)
   - Node types: scripture (by volume color-coded), talk, study, topic-guide entry
   - Edge types: footnote, tg, bd, gs, jst (color/style coded)
   - Click node → emit event to load content
   - Hover → show reference label
   - Zoom, pan, fit controls

2. **ReferencePanel modification** — Add "Graph" tab alongside existing tabs
   - When graph tab is active and user clicks a paragraph, graph updates
   - Graph tab shares the reference panel's real estate

3. **Reader integration** — Paragraph click handler
   - On paragraph click/focus, extract scripture references from that paragraph's text
   - Send references to graph API
   - Update graph panel

### Graph data schema

```typescript
interface GraphNode {
  id: string           // "ot/gen/1:1" or "talk/2024/04/holland"
  label: string        // "Genesis 1:1" or "Elder Holland, Apr 2024"
  type: 'scripture' | 'talk' | 'study' | 'tg'
  volume?: string      // "ot", "nt", "bofm", "dc", "pgp"
  metadata?: Record<string, string>  // extensible for TITSW scores later
}

interface GraphEdge {
  source: string       // node ID
  target: string       // node ID
  type: 'footnote' | 'tg' | 'bd' | 'gs' | 'jst' | 'citation' | 'study'
  label?: string       // e.g., "1a" for footnote letter
}
```

---

## 6. Phased Delivery

### Phase 1: Scripture Cross-Reference Graph Tab (1-2 sessions)

**Scope:** Add a graph tab to ibeco.me's reader that visualizes gospel-mcp cross-references.

| Deliverable | Detail |
|-------------|--------|
| `internal/graph/graph.go` | Graph query: given references, return N-hop neighborhood from cross_references table |
| `POST /api/graph/edges` | HTTP endpoint returning nodes + edges |
| `GraphPanel.vue` | Cytoscape.js graph renderer with force-directed layout |
| ReferencePanel update | "Graph" tab alongside existing reference tabs |
| Reader integration | Paragraph click → extract refs → update graph |

**Verification:**
1. Open any Book of Mormon chapter in reader
2. Click a verse paragraph
3. Graph tab shows that verse as center node with all footnote cross-references radiating out
4. Cross-references to other volumes (OT, NT) appear in different colors
5. Click a cross-reference node → reader loads that chapter, scrolls to verse
6. TG/BD entries appear as labeled nodes connecting multiple verses

**Stands alone:** Yes. This delivers the core "make connections visible" value with existing data.

### Phase 2: BYU Citation Edges (1 session)

**Scope:** Cache BYU citation data and add conference talk nodes to the graph.

| Deliverable | Detail |
|-------------|--------|
| `talk_citations` table | Scraped and cached BYU citation data in gospel-mcp SQLite |
| Scrape pipeline | Bulk crawl of byu-citations for all standard work references |
| Graph API update | Include citation edges in graph response |
| GraphPanel update | Talk nodes styled distinctly, clickable |

**Verification:** Click a well-known verse (John 3:16). Graph shows footnote edges AND conference talks that cite it. Click a talk node → see talk metadata.

**Depends on:** Phase 1 working.

### Phase 3: Study Document Edges (1 session)

**Scope:** Parse our study documents for scripture references and add them to the graph.

| Deliverable | Detail |
|-------------|--------|
| Study document parser | Scan `study/` markdown for scripture references |
| `study_references` table | Store study→scripture edges |
| Graph API update | Include study edges |
| GraphPanel update | Study nodes styled distinctly, click → open in reader |

**Verification:** Click a verse we've studied. Graph shows a study document node connected to it. Click study node → opens that study in the reader.

**Depends on:** Phase 1 working.

### Phase 4: Multi-hop Exploration & Navigation (1 session)

**Scope:** Make the graph explorable with progressive disclosure and navigation state.

| Deliverable | Detail |
|-------------|--------|
| Expand/collapse | Click a node to expand its neighborhood, collapse to reduce clutter |
| N-hop slider | Control traversal depth (1-3 hops) |
| Tab history | Back/forward navigation through graph states |
| Node detail panel | Hover or click → see full reference text, metadata, direct link |

**Verification:** Start at one verse. Expand 2 hops. See a web of connections. Navigate back. Forward. Collapse a cluster. The graph remains usable, not a hairball.

**Depends on:** Phase 1 working.

### Phase 5: Thematic Edges & Enriched Metadata (future)

**Scope:** Add non-explicit connections via LLM inference and semantic similarity.

**Depends on:** Enriched indexer Phases 1-3 (separate workstream).

### Phase 6: study.ibeco.me (future — if warranted)

**Scope:** Dedicated study site with full-page graph mode and search→graph→read workflow.

**Depends on:** Phases 1-4 proven, user demand demonstrated.

---

## 7. Verification Strategy

| Phase | Verification |
|-------|-------------|
| 1 | Manual testing: pick 5 chapters across all 5 standard works. Confirm graph shows correct cross-references. Compare against printed scripture footnotes. Edge count should match. |
| 2 | Spot-check 10 well-cited verses against BYU website. Citation count should match (±1 for pagination). |
| 3 | Spot-check 5 study documents. All scripture references in the study should appear as edges in the graph. |
| 4 | UX walkthrough: navigate a 3-hop graph without getting lost. Back button works. Collapse works. |

---

## 8. Costs & Risks

### Costs
- **Development time:** Phase 1 is 1-2 sessions. Full vision (Phases 1-4) is 4-5 sessions.
- **Maintenance:** New Vue component + new backend package. Incremental on existing ibeco.me.
- **Dependencies:** Phase 1 has zero blockers. Phase 5+ depends on enriched indexer.
- **Bundle size:** Cytoscape.js is ~170KB gzipped. Significant but acceptable for the capability it provides.

### Risks
| Risk | Mitigation |
|------|-----------|
| Graph becomes a hairball at high hop counts | Default to 1-hop. Progressive disclosure. LOD. |
| cross_references table has gaps or errors | Spot-check against printed footnotes in Phase 1 verification |
| Performance: large neighborhoods slow to render | Limit to 500 nodes per request. Cytoscape.js handles this well. |
| Scope creep: "just one more edge type" | Strict phasing. Phase 1 ships with ONLY footnote cross-references. |
| ibeco.me complexity growth | Graph is a tab, not a rewrite. If it doesn't prove useful, remove the tab. |

---

## 9. Creation Cycle Review

| Step | Question | This proposal |
|------|----------|--------------|
| **Intent** | Why are we doing this? | Make the invisible connections in scripture visible and explorable. Directly serves the project's root intent: deep, honest scripture study. |
| **Covenant** | Rules of engagement? | ibeco.me conventions (Vue 3/Go), read-before-quoting for verification, source data from gospel-mcp. |
| **Stewardship** | Who owns what? | ibeco.me frontend+backend: dev agent. gospel-mcp data: already indexed. Graph component: new, owned by this workstream. |
| **Spiritual Creation** | Is the spec precise enough? | Phase 1 yes — data flow, API shape, component structure, verification criteria all defined. Phase 5-6 intentionally vague. |
| **Line upon Line** | What's the phasing? | 6 phases. Phase 1 stands alone. Each phase adds one edge type or capability. |
| **Physical Creation** | Who executes? | dev agent builds. Michael reviews and tests. |
| **Review** | How do we know it's right? | Footnote edge count matches printed scripture footnotes. Spot-checks against BYU website. UX walkthrough. |
| **Atonement** | What if it goes wrong? | Graph tab is additive — reader works without it. Worst case: remove the Graph tab. No data is modified. |
| **Sabbath** | When do we stop and reflect? | After Phase 1 — does the graph actually improve study? Is it useful or just pretty? |
| **Consecration** | Who benefits? | Michael first. Other ibeco.me users eventually. The visualization concept could be shared. |
| **Zion** | How does this serve the whole? | Integrates gospel-mcp data, ibeco.me reader, and BYU citations into a unified experience. Moves toward "gospel-comb" vision without premature abstraction. |

---

## 10. Recommendation

**Build — Phase 1 now, remaining phases sequentially.**

Phase 1 has zero blockers, uses existing data, and extends existing infrastructure with minimal new code. The core value — seeing cross-reference connections visually — can be delivered in 1-2 dev sessions. Each subsequent phase adds one capability and stands on the previous phase.

**Sequencing with existing work:**
- Phase 1 can run **now** — it only needs gospel-mcp's existing cross_references data
- Enriched indexer Phase 1 (talk batch enrichment) can run in parallel — they don't conflict
- Phase 5 (thematic edges) should wait for enriched indexer Phases 1-3 — that's where the metadata comes from
- Phase 6 (study.ibeco.me) should wait until Phases 1-4 prove the graph is actually useful in practice

**Phase 1 scope for dev agent:**
1. Add `internal/graph/` package to ibeco.me backend
2. Add `POST /api/graph/edges` endpoint
3. Add `GraphPanel.vue` with Cytoscape.js
4. Add "Graph" tab to ReferencePanel
5. Wire paragraph click → graph update → node click → reader navigation

**Estimated complexity:** Small backend addition (100-200 lines Go), medium frontend addition (300-500 lines Vue/TS), no new infrastructure.
