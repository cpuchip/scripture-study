# stewards-ui build plan — 2026-05-09 evening

Michael ratified the 3f proposal direction with these decisions:

## Ratified architecture

- **Name:** `stewards-ui` (binary, container `pg-ai-stewards-ui`)
- **Service shape:** separate `ui` compose service alongside `pg` + `bridge`
- **Port:** single 8080, Go serves both `/` (Vue SPA) and `/api/*`
- **Bundling:** Vue dist embedded into Go binary via `embed.FS`
- **Frontend stack:** Vue 3 + Vite + TypeScript + Tailwind 4 + vue-router
  (matches `scripts/becoming/frontend/`)
- **UI components:** shadcn-vue (copy-paste snippets, Tailwind-based)
- **Graph:** Cytoscape.js, studies + citations only in v1
- **Auth:** none (127.0.0.1 binding)
- **Build context:** repo root via `context: ../../..` from compose

## Directory layout

```
scripts/stewards-ui/
├── go.mod                    # API module
├── go.sum
├── main.go                   # HTTP server, embed.FS for dist
├── api/                      # /api/* handlers
│   ├── dashboard.go
│   ├── studies.go
│   ├── work_items.go
│   ├── sessions.go
│   ├── watchman.go
│   ├── bridge.go
│   ├── graph.go
│   └── search.go
├── frontend/                 # Vue/Vite project
│   ├── package.json
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── tailwind.config.js    # if needed (Tailwind 4 may auto-config)
│   ├── index.html
│   ├── src/
│   │   ├── main.ts
│   │   ├── App.vue
│   │   ├── router.ts
│   │   ├── api.ts            # typed fetch wrappers
│   │   ├── style.css
│   │   ├── components/       # shared UI bits
│   │   ├── components/ui/    # shadcn-vue copy-pastes
│   │   ├── composables/
│   │   └── views/
│   │       ├── Dashboard.vue
│   │       ├── Studies.vue
│   │       ├── StudyDetail.vue
│   │       ├── WorkItems.vue
│   │       ├── WorkItemDetail.vue
│   │       ├── Sessions.vue
│   │       ├── Watchman.vue
│   │       ├── BridgeState.vue
│   │       ├── Graph.vue
│   │       └── NewWork.vue
│   └── dist/                 # vite build output (gitignored)
└── README.md

projects/pg-ai-stewards/extension/
├── ui.Dockerfile             # multi-stage: node build frontend, go build, alpine runtime
└── docker-compose.yaml       # +ui service
```

## Build phases (incremental)

### Phase 1 — Foundation (tonight)

- `scripts/stewards-ui/` directory + go.mod + main.go skeleton
- `scripts/stewards-ui/frontend/` Vite-Vue init (matching becoming-app
  package.json)
- Vue router + single placeholder view (`/dashboard`)
- Tailwind 4 configured
- shadcn-vue dir scaffolded (no components yet)
- `ui.Dockerfile` multi-stage: node-build frontend → dist; go-build api
  with embedded dist; alpine runtime
- `docker-compose.yaml` `ui` service entry
- Build, smoke: localhost:8080 returns the placeholder page
- Commit "feat(stewards-ui): v1 phase 1 — foundation scaffold"

### Phase 2 — Dashboard + state API (next session)

- `/api/dashboard` Go handler — health + soak summary + in-flight
  work_items + recent errors
- `Dashboard.vue` consumes the API; renders cards
- shadcn-vue Card, Badge, Skeleton components added
- 5s polling for live state (manual refresh button + auto)

### Phase 3 — Studies browse + global search

- `/api/studies/list?kind=&limit=`, `/api/studies/get/:slug`,
  `/api/studies/search?q=&mode=fts|semantic|combined`
- `Studies.vue` list view with search bar
- `StudyDetail.vue` renders study body via markdown-it; shows
  citations + similar studies
- shadcn-vue Input, Table, Tabs

### Phase 4 — Work items + sessions

- `/api/work-items/list?pipeline=&status=`, `/api/work-items/get/:id`
- `/api/sessions/get/:id` — message timeline
- `WorkItems.vue`, `WorkItemDetail.vue`, `Sessions.vue`
- Token-spend visualization

### Phase 5 — Watchman + bridge state

- `/api/watchman/passes`, `/api/watchman/pass/:id`,
  `/api/watchman/findings/ack`
- `/api/bridge/state` (returns mcp_bridge_state view)
- `/api/bridge/refresh-tools` (POST — triggers refresh)
- Views

### Phase 6 — Graph view

- `/api/graph/studies-citations` returns nodes + edges from substrate
  AGE Cypher
- `Graph.vue` renders Cytoscape.js graph
- Click node → drill into Studies page

### Phase 7 — New work form

- `/api/work-items/create` — pipeline + binding question + budget
- `/api/work-items/dispatch/:id`
- `NewWork.vue` form

## What I will not do without confirming

- Push to remote
- Restart the live `pg` container (soak data preserved but
  mid-pass interruption cancels in-flight Watchman work)
- Touch existing `bridge` or `pg` services in compose beyond
  adding the new `ui` service
- Spawn any chat work that would consume model tokens
- Make architectural decisions Michael's ratified list didn't cover
- Add component libraries beyond shadcn-vue snippets (no Naive UI,
  no Element Plus, no headless UI runtime deps)

## Tonight's scope

Just **Phase 1** — foundation scaffold. ~2 hours. End state: page
loads at localhost:8080, says "stewards-ui v1 phase 1", no real data
yet. Validates the multi-stage build, the embed.FS pattern, and the
docker-compose service all play together cleanly. Phases 2-7 follow
in subsequent sessions.
