---
workstream: WS5
status: design pass — awaiting Michael's direction on stack + scope
created: 2026-05-09
supersedes:
  - the original 3f spec ("a.ibeco.me web UI surface" in projects/pg-ai-stewards/phases.md)
related:
  - 3e-mcp-findings.md (MCP wave that produced the data this UI surfaces)
  - pg-ai-stewards-3d-sandboxed-git.md (parallel agent-capability work)
---

# Phase 3f — Local pg-ai-stewards UI

> Design doc; **no implementation yet**. Michael's pivoted vision (2026-05-09):
>
> > "My idea for a.ibeco.me probably needs to be more flushed out since
> > our current setup doesn't allow for more than 1 user. For now I want
> > an elegant UI that allows me to see the state, explore the data,
> > kick off AI work and interact with it. And see results. To search
> > it, and see the graph connections between studies/gospel-library/etc.
> > This will be hosted locally, probably in the bridge or a new service
> > that runs beside the other two."
>
> This is a **significant departure from the original 3f spec**. The
> original (`a.ibeco.me`) was cloud-hosted, multi-user, OAuth via
> becoming. This is local-first, single-user, co-located with the
> docker stack. Naming TBD; this doc uses "stewards-ui" as a
> placeholder.

## Why this matters now

The producer side of pg-ai-stewards is real. Phase 3e shipped the
bridge, the substrate ran the mysteries-of-god study end-to-end with
organic gospel_get verification, and there are now 370+ studies, 33+
watchman passes, 800+ work_queue rows, and a graph of citations
sitting in the substrate.

Today the only ways to see any of that are:
- `docker exec ... psql ...` (developer-grade, raw SQL)
- The MCP inbound tools through Claude Code (text-only,
  query-at-a-time)
- Hand-running `stewards-cli` commands

None of those compose well into "show me what the substrate is
thinking about right now." We've reached the threshold where Michael
benefits from a real visual surface.

## What the UI is for

From Michael's vision, decomposed:

1. **See the state** — health, soak status, in-flight work_queue,
   recent passes, watchman dirty queue depth, recent errors
2. **Explore the data** — browse studies, work_items, sessions,
   messages, verdicts, findings; jump from one to its citations or
   neighbors
3. **Kick off AI work** — submit a binding question + pipeline +
   token budget; see it move through stages
4. **Interact with it** — pause, cancel, advance work_items;
   maybe ack findings; maybe broadcast a message into a session
5. **See results** — read the published study; see what tools were
   called; trace the agent's reasoning
6. **Search** — across studies (FTS + semantic), maybe across
   gospel-library too via gospel-engine
7. **Graph connections** — node-link view of studies +
   gospel-library citations; click to traverse

What this UI is **NOT** for (out of scope for v1):
- Authoring scripts or templates (use the IDE)
- Editing tool_defs / agents / pipelines (use psql or stewards-cli)
- Cloud hosting / multi-user (local single-user only)
- Mobile (desktop browser only)
- Authentication (local-only; IP binds to 127.0.0.1)

## Threat model

For a local-only, single-user UI on a developer workstation:

- The threat is mostly **accidental misclick** (advance a wrong
  work_item, cancel an in-flight pass), not adversarial input.
- The substrate already has guardrails (token budgets per work_item,
  deny-by-default tool grants, soak cooldowns).
- The UI is one more way to drive those existing safeguards; it
  doesn't introduce new privilege.

What the UI must not do:
- Bind to anything other than 127.0.0.1
- Expose model-call or token authentication material in the DOM
- Cache auth tokens in browser localStorage (none needed; no auth)
- Allow file system access beyond what the substrate already exposes

## Architecture options

### Option A: Go-served single binary (Recommended starting point)

A new `scripts/stewards-ui/` Go binary (or fold into bridge) that
serves an embedded HTML/JS bundle and a JSON API backed by pgxpool.
HTMX or vanilla JS for interactivity; no build step beyond `go build`
+ asset embedding via `embed.FS`.

**Pros:**
- Matches the existing pattern (gospel-engine, webster-mcp, fetch-md-
  mcp, bridge are all Go binaries). One process to deploy. Single
  binary to ship.
- No node_modules, no Vite, no separate frontend build step.
- HTMX gives surprisingly rich UX with very little JS.
- Easy to fold into the bridge container if we want fewer services
  (one more port exposure).

**Cons:**
- Limited to what HTMX + small JS can do gracefully. Graph view is
  the hardest case (Cytoscape.js is JS-heavy and works best in a
  rich frontend).
- Less vibrant ecosystem than Vue/Vite for complex interactions.

### Option B: Vue/Vite + thin Go API split

Frontend in `scripts/stewards-ui-frontend/` (Vue 3 + Vite + Pinia),
backend in `scripts/stewards-ui-api/` (Go pgxpool + HTTP). Frontend
served as a static bundle (could be embedded into the Go binary at
build time, or served by Vite in dev).

**Pros:**
- Matches the becoming app's stack (Vue/Vite). Familiar shape.
- Better at rich interactions (graph view, drag-rearrange,
  multi-pane).
- Frontend can iterate independently.

**Cons:**
- Two processes / two build pipelines. More moving parts.
- npm install + node_modules bloat (~100MB) for a single-user UI is
  heavy.
- Couple-day build instead of half-day for v1.

### Option C: Fold the UI into the bridge container

Same Go binary as Option A, but ship inside the bridge image. One
fewer service. The bridge already has pgxpool, env, network access
to pg.

**Pros:**
- Zero new compose service. `docker compose up -d` brings up everything.
- Reuses bridge's existing pgxpool config.

**Cons:**
- Mixes concerns: bridge handles spawn/RPC; UI handles HTTP. Different
  failure modes; harder to restart one without the other.
- Bridge restart for a UI-only bug is annoying.
- Port mapping gets messy.

## Recommendation

**Option A as v1, served by a new `ui` compose service alongside `pg`
and `bridge`.** Reasons:

- Single Go binary fits the existing pattern.
- HTMX + small JS handles 80% of the surface (state, browse, search,
  pipeline kick-off, work_item interaction).
- Defer the graph view to v2 — it's the only piece that wants a
  richer JS ecosystem. When we get there, embed Cytoscape.js
  directly via CDN or vendor; still no Vue/Vite needed.
- New compose service rather than folding into bridge keeps restart
  semantics clean.

Long-run, if the graph view + interactive timeline + multi-pane data
exploration starts to feel cramped, **revisit Option B** without
discarding Option A's bones — the JSON API surface is the same.

## v1 scope (slim)

Read-only state browser + search + pipeline kick-off. No graph view
(that's v2). No write tools beyond pipeline create/dispatch/cancel.

**Pages:**

1. **Dashboard** — health card (pg + bridge status), soak summary
   (last pass, dirty queue depth, schedule_enabled), in-flight
   work_items, recent errors. Auto-refresh on a 5s tick.
2. **Studies** — list view with FTS search bar. Click into a study
   to see body, citations, similar studies (uses
   `stewards.study_search_text`, `study_get`, `study_citations`,
   `study_similar`).
3. **Work items** — list filtered by pipeline + status. Click into
   one to see stage_results, session_ids, token usage, error.
   Action buttons: dispatch, advance, cancel.
4. **Sessions** — drill into a session to see message timeline
   (system → user → assistant → tool → continuation). Token-spend
   per turn. Tool-call inspector.
5. **Watchman** — recent passes with verdict counts; click in to
   see per-doc verdicts and findings. Ack action on findings.
6. **Bridge state** — `mcp_bridge_state` view rendered as a table.
   Per-server tool catalog. Refresh-tools button.
7. **New work** — form: pick pipeline, paste binding question,
   optional slug, optional token budget. Submits via the same
   `work-item create` + `dispatch` shape as `stewards-cli`.

**Search:**

A global search bar at the top. Submits to a backend endpoint that
runs `study_search_text` (FTS) and `study_similar` (semantic) in
parallel and merges. Results link to the Studies page entry.

**In-flight observability:**

Server-Sent Events stream of `work_queue` row transitions
(NOTIFY-driven, cheap), rendered as a timeline. Per-row tap-out
to see the full row JSON.

## v2 scope (deferred)

- Graph view via Cytoscape.js. Queries:
  - Studies → citations → studies (substrate AGE :CITES edges)
  - Studies → similar → studies (pgvector cosine over a threshold)
  - Studies → workstreams → studies (frontmatter declared edges)
- Drag-rearrange, zoom, filter by edge type
- Click a node to drill into Studies page

## Open design questions for Michael

1. **Naming** — "stewards-ui"? "substrate-ui"? "becoming-local"?
   Something else? The placeholder in this doc is `stewards-ui`.
2. **Service shape** — separate `ui` compose service (recommended)
   or fold into bridge?
3. **Port** — 5174? 8080? Keep 8080-ish to leave room for future
   becoming-local? Bind 127.0.0.1 only.
4. **Auth** — confirmed none for local single-user. If we ever
   share with another machine on the LAN, revisit.
5. **Scope cut** — okay to defer graph view to v2? It's the most
   visible feature but also the most code. Slim v1 ships faster.
6. **Tech for the embedded JS** — HTMX? Alpine.js? Vanilla?
   HTMX + Alpine is a common combo for "Go-rendered HTML +
   reactive bits."
7. **Theme / styling** — shadcn-ui vibes (CSS framework like
   Tailwind, but built into the Go-served HTML)? Or just plain
   stylesheet?

## Done criteria (v1)

- `docker compose up -d` brings up pg + bridge + ui
- `http://localhost:<port>/` shows the dashboard with live state
- Search bar finds a study; click renders body
- Work items list shows in-flight pipelines; pipeline kick-off
  works (creates a work_item, dispatches it, returns to dashboard)
- Watchman page shows recent passes; ack a finding works
- Bridge state page shows the 9 MCP servers with their tool
  catalogs; refresh-tools button works

## Phase boundaries

- **3f.1 — UI v1 (this proposal).** Slim scope above. ~1-2 days.
- **3f.2 — Graph view.** Cytoscape.js + AGE/Cypher queries. ~1-2
  days.
- **3f.3 — Write actions beyond pipeline.** Edit agents, edit
  tool_defs, broadcast a message into a session. Triggers when
  Michael wants to manage substrate state visually instead of
  via psql.
- **3f.4 — Multi-user / cloud.** When/if we want
  shared substrate or `a.ibeco.me`-style hosting. The original
  cloud-hosted spec lives at the top of this doc as the
  long-run destination if needed.

## Why not just build it now

- Stack choice (HTMX vs Alpine vs vanilla) is reversible but
  shapes the next several weeks of UI work.
- Service-shape choice (ui as separate vs fold into bridge) is
  also reversible but worth Michael's preference.
- Port choice + naming is a small thing but impossible to undo
  cleanly later (every doc + bookmark + screenshot ends up
  pointing at the wrong place).
- The slim v1 scope is conservative; Michael may want to expand
  (or further trim) before any code lands.

The proposal is ready. Implementation triggers on Michael's
direction.
