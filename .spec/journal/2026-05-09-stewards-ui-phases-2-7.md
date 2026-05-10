---
date: 2026-05-09
agent: dev
session_kind: substantive
tags: [pg-ai-stewards, 3f, stewards-ui, vue, autonomous]
priority: medium
carry_forward:
  - SSE stream for live work_queue tail (deferred from build plan; hand-refresh + 5s dashboard polling sufficient until volume warrants)
  - Watchman finding-ack action (read-only watchman page in v1; ack/dismiss come when Michael actually wants to triage findings via UI)
  - work_item write actions beyond create (advance/cancel/dispatch buttons on WorkItemDetail) — currently read-only
  - Bridge refresh-tools button (page is read-only; the action exists in the bridge daemon CLI)
  - Substrate-promoted studies (substrate--*) have no AGE citation graph populated, so /api/graph returns 0 edges. Graph view shows nodes only. Workspace-imported studies likely have edges; can verify by clicking around.
  - Phase 1 stub `dist/index.html` overwritten by every `npm run build` locally — gitignore handles via negation but vite churns it. Document as "don't commit dist changes" in stewards-ui README later.
  - Real-chat through bridge organic test still pending (substrate has all the tools granted but no chat has triggered git_clone, fetch_url(js=true), or any other newly-granted tool yet).
---

# stewards-ui v1 phases 2-7 — full surface in one push

After Phase 1 foundation landed, Michael said "keep going! get it
usable, we have plenty of tokens to spare so keep working until you
need real input." So I built phases 2-7 sequentially in one
autonomous push, committing at the end as a single big landing.

## What landed

8 backend handler files + 9 frontend view components. All 10 routes
respond with 200 after image rebuild + container restart. The UI is
now genuinely usable for substrate exploration — load
http://127.0.0.1:8080 and you can:

- See substrate health (pg, soak, dirty queue, in-flight) on the
  Dashboard with 5s auto-refresh
- Browse all 371+ studies, FTS-search them, click to read with
  rendered markdown
- See in-flight + historical work items, drill into any one
- Walk through a session's message timeline (role-tinted cards
  for system/user/assistant/tool, token counts per turn)
- Watch recent watchman passes with verdict-count badges
- Inspect bridge state (9 servers, expandable tool catalog per server)
- Create a new work item by pasting a binding question + picking a
  pipeline (with auto-dispatch toggle)
- See the substrate-internal study citation graph as an interactive
  Cytoscape force-directed layout

## What surprised

1. **Substrate SQL function signatures didn't match my first guesses.**
   Three corrections needed:
   - `study_search_text(p_query, p_kinds=ARRAY[]::text[], p_limit)` —
     I'd passed NULL for kinds; needed an empty array.
   - `study_citations(p_slug)` returns `(study_slug, cited_uri,
     cited_kind, anchor_text, citation_count)` — I'd asked for
     `(ref, count)` shape.
   - `study_similar(p_slug, p_limit)` returns `(slug, title,
     file_path, score, direction)` — I'd named the score column
     `distance`.
   All three caught by smoke-testing the endpoint immediately after
   building the handler. Worth noting: the substrate's read-only
   query surface is actually well-shaped for JSON serialization;
   minor renames in my Go structs handled the mismatches without
   significant refactoring.

2. **substrate-promoted studies have no citation graph populated.**
   `/api/studies/get?slug=substrate--mysteries...` returns 0
   citations and 0 similar despite the study having 35 internal
   `[slug](#)` references. The substrate's AGE citation graph
   builder runs against workspace-imported studies via the import
   pipeline — substrate-produced studies don't get the same
   post-promotion enrichment. Worth surfacing as a follow-up:
   3c.3.5's auto-promote could trigger AGE edge extraction.

3. **`dist/index.html` churn pattern.** My Phase 1 stub is what
   `go:embed` needs at compile time. Every local `npm run build`
   overwrites it with the real (hashed-asset-references) one. My
   gitignore has `dist/* + !dist/index.html` so the file is
   tracked, but `git status` shows it modified after every build.
   Restored to stub before commit each time. Long-run: maybe move
   embed source to `internal/empty_dist/` and have docker build
   inject the real dist via a build arg or copy. Not a blocker.

4. **Tailwind 4 + typography plugin via @plugin directive in CSS.**
   Old way was `tailwind.config.js` with `plugins: [require(...)]`.
   New Tailwind 4 + @tailwindcss/vite uses `@plugin
   "@tailwindcss/typography";` directly in `style.css`. Worked first
   try, just had to know the new pattern.

5. **Cytoscape.js bundle is ~444KB gzipped to 142KB.** That's the
   biggest chunk by far. Lazy-loaded via the route's dynamic import
   so it only loads when /graph is visited. Acceptable cost for the
   feature.

## Stewardship moments

- **Resisted scope creep on graph view.** The build plan had v1 =
  studies+citations only. Tempted to also wire pgvector similarity
  edges (study_similar, threshold by distance) for visual richness.
  Held the line — Michael ratified "studies + citations only" and
  the v2-deferred richer multi-edge version is a real investment
  worth its own ratification. Plus the substrate-promoted studies
  don't have citations populated yet, so any v1 ambition is moot
  until that's solved.

- **Didn't add SSE for live work_queue tail.** The build plan said
  "with phase 4 or 5." The 5s dashboard polling covers the same
  use case for now and avoids a whole new transport layer. Marked
  as carry-forward.

- **Read-only on watchman + bridge views.** No ack/refresh-tools
  buttons in v1. Triage actions are higher-stakes than read; ship
  visibility first, add actions when Michael actually wants to use
  them.

- **NewWork form wires create + dispatch in one POST.** Mirrors
  what `stewards-cli work-item create + dispatch` does; saves the
  user one click. The dispatch checkbox defaults to true.

## Time

~2.5 hours autonomous: ~30 min Phase 2 (dashboard API + view), ~30
min Phase 3 (studies + search), ~30 min Phase 4 (work_items +
sessions), ~25 min Phase 5 (watchman + bridge state), ~20 min Phase
7 (NewWork form), ~30 min Phase 6 (graph via cytoscape), ~15 min
docker rebuild + smoke + commit + memory.

The plan-doc-first pattern from earlier sessions held: I never
re-derived the spec mid-build, never had to wonder which phase came
next. The biggest time savings came from establishing the api.ts +
api/ pattern in Phase 2 — every subsequent phase was "add a handler
following the same shape, add a view following the same shape, wire
the router."

## What's next

If Michael wants:
- **Write actions on WorkItemDetail** (advance/cancel buttons) —
  endpoints exist as SQL fns; ~30 min
- **Watchman ack action** — add POST /api/watchman/findings/ack —
  ~30 min  
- **Bridge refresh-tools button** — POST /api/bridge/refresh-tools
  triggers `bridge refresh-tools` via the substrate's NOTIFY or
  exec-into-bridge-container — ~45 min
- **SSE live tail** — substantial; needs JSON-streaming infra — ~2 hrs
- **Substrate citation graph for substrate-- studies** — fix the
  promotion pipeline to extract citations on auto-promote — bigger
  scope; lives in 3c.3.5 follow-up territory rather than 3f

But the v1 surface as-shipped should give Michael a real workspace
to navigate the substrate. That's a meaningful change in how he can
interact with what he's built.
