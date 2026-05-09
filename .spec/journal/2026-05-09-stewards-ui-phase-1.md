---
date: 2026-05-09
agent: dev
session_kind: substantive
tags: [pg-ai-stewards, 3f, stewards-ui, foundation, autonomous]
priority: medium
carry_forward:
  - 3f phase 2 — real /api/dashboard endpoint + Dashboard.vue card grid (substrate state, soak summary, in-flight work_items)
  - 3f phase 3 — studies browse + global search bar (backend FTS + semantic merge; Studies.vue + StudyDetail.vue with markdown-it rendering)
  - 3f phase 4 — work_items + sessions views (timeline, token-spend per turn, tool-call inspector)
  - 3f phase 5 — watchman + bridge state pages (refresh-tools button, finding-ack action)
  - 3f phase 6 — graph view via Cytoscape.js (studies + citations only in v1)
  - 3f phase 7 — new-work form (pipeline + binding question + budget)
  - shadcn-vue components: copy-paste snippets when first needed (Button, Card, Input, Table, Badge, Skeleton, Tabs)
  - SSE stream for live work_queue tail (deferred from phase 1; comes with phase 4 or 5)
---

# stewards-ui v1 phase 1 — foundation scaffold

Michael ratified the 3f local-UI design in two AskUserQuestion batches
and asked me to "kick that off" with him. Phase 1 = foundation only
per the build plan I authored before starting.

## What landed

Brand-new `scripts/stewards-ui/` directory:

```
scripts/stewards-ui/
├── go.mod, go.sum, main.go         # API binary with embed.FS
├── api/                             # /api/* handlers (empty in p1)
└── frontend/
    ├── package.json, vite.config.ts
    ├── tsconfig.json + tsconfig.app.json + tsconfig.node.json
    ├── index.html
    ├── dist/index.html              # stub for go:embed (committed)
    └── src/
        ├── main.ts, App.vue, router.ts, style.css
        ├── components/{,ui}/        # shadcn-vue dest, empty in p1
        ├── composables/, services/, assets/
        └── views/
            ├── Dashboard.vue        # real (calls /healthz)
            └── Placeholder.vue      # 8 routes share this
```

Plus:
- `projects/pg-ai-stewards/extension/ui.Dockerfile` — node→go→alpine
  3-stage build
- `docker-compose.yaml` gains `ui` service alongside `pg` and
  `bridge`; depends on pg.healthy; binds 127.0.0.1:8080 only

Single binary contains embedded Vite dist + the API surface. Phase 2+
fills in /api/* handlers and view content; the foundation is now
solid enough to iterate.

## What surprised

1. **`go:embed all:frontend/dist` needs files at compile time.**
   Stubbed `dist/index.html` with a placeholder note explaining the
   pattern. Vite's `npm run build` overwrites it during real builds.
   Gitignored via negation: `dist/*` ignored except `dist/index.html`.

2. **`@vue/tsconfig/tsconfig.node.json` doesn't exist** in current
   versions of `@vue/tsconfig`. My initial 2-file tsconfig (with
   tsconfig.json extending dom.json + tsconfig.node.json extending
   node.json) failed inside the Vite TS check. Fix: matched
   becoming-app's 3-file split exactly — root `tsconfig.json` is
   pure references, `tsconfig.app.json` extends dom, `tsconfig.node.json`
   stands alone.

3. **`scripts/yt-mcp/`, `lectures-on-faith/`, `publish/` lack go.sum**
   (no external deps beyond stdlib). My ui.Dockerfile assumed they
   had go.sum like everyone else. Removed those COPY lines.
   Same shape gotcha I'd already hit on bridge.Dockerfile and
   forgotten about.

4. **Compose context relative depth.** Already learned this on
   bridge: `context: ../../..` not `../..`. ui.Dockerfile inherits
   the same pattern correctly the first time.

## Stewardship moments

- **Resisted Phase 2 momentum.** After Phase 1 worked, the
  natural pull was "the build pipeline is hot, let's add the
  /api/dashboard endpoint and the Card components now." Stuck to
  the plan-doc scope cut. Phase 2 follows in a future session.
  Reason: Michael ratified Phase 1 only for tonight; expanding scope
  unilaterally would burn his ratification budget.

- **Stub committed, real outputs gitignored.** A handful of obvious
  patterns failed first (gitignore everything in dist/ → go:embed
  fails on fresh clone; commit everything → hashed filenames churn
  in git). Settled on negation pattern that ships the exact stub
  needed for clean compile, ignores the rest.

- **No npm-create-vue interactive setup.** That CLI prompts for
  TypeScript? router? pinia? testing? Each prompt is an interaction
  point I'd have had to surface back. Manually scaffolded
  package.json + tsconfigs + main.ts to match becoming-app's
  conventions exactly. Reproducible, no prompts.

## Live verification

```
docker compose up -d ui
→ Container pg-ai-stewards-ui Started

curl http://127.0.0.1:8080/healthz
→ ok                  (200, db ping passes)

curl http://127.0.0.1:8080/
→ <real Vue HTML with /assets/index-DLgl3fZQ.js script tag>

curl http://127.0.0.1:8080/studies
→ <same index.html, 200>      (SPA fallback works)

curl http://127.0.0.1:8080/api/dashboard
→ stewards-ui phase 1: api endpoints not yet implemented   (501)
```

`docker ps` shows pg + bridge + ui all running healthy. Soak still
cadencing through it all.

## Time

~1.5 hours: 30 min plan + scaffolding decisions, 15 min Go API,
30 min Vue/Vite scaffold, 15 min Dockerfile + compose service,
20 min build/smoke iteration (tsconfig + go.sum gotchas), 10 min
commit + memory.

## What's next

Phases 2-7 per the build plan. Each phase is roughly 1-2 sessions
of work. Phase 2 (Dashboard + state API) is the obvious next thing
and gets the UI from "working scaffold" to "actually useful."
