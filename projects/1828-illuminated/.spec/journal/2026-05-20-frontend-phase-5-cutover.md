---
date: 2026-05-20
session_type: dev
project: 1828-illuminated
workstream: WS7
status: phase-5-shipped
commits:
  - bbc1466 — phase 5.A+B (composables + Settings BYOK UI + Vite env)
  - 07bf9f3 — phase 5.C (canon-browse mode + WordSearch class-E reach)
parent: 2026-05-20-backend-phase-1234-build.md
---

# Frontend cutover — phase 5

Picked up where the phase-1-4 session left off. The four backend
phases shipped earlier in the day put every API the frontend needed in
place; this session swapped the Vue SPA from "static-import every JSON
bundle" to "fetch from `/api/*` and let the backend own the
heavy-lifting." Two commits on `main`. The legacy single-container `1828`
on 8082 never moved.

## What shipped, by commit

**bbc1466 — phase 5.A+B: composables + Settings BYOK UI + Vite env.**
The split between 5.A and 5.B that the orchestrator hinted at collapsed
to one commit because the `useLLMSettings` reshape forced
`VerseExplorer` + `Settings` to re-import `sessionActive` at the same
time. Splitting would have shipped a half-working state.

- `useApiBase.ts` (new): resolves `VITE_API_BASE_URL` with inline fallback
  to `/api` so production builds never hardcode `localhost`. `.env.example`
  documents the variable; `.env.development` (force-added past the
  workspace `.env.*` gitignore) overrides to `http://localhost:8083/api`
  for `npm run dev` against the compose stack. `vite-env.d.ts` adds the
  `ImportMetaEnv` types so `vue-tsc` is happy without `skipLibCheck`
  hiding bugs.
- `useWordData.ts`: dropped the static imports of `definitions-1828.json`
  (1.7MB) and `definitions-modern.json` (876KB). `tier-words.json` +
  `manual-additions.json` stay statically imported because they drive
  the synchronous tokenize() highlight pass + the curated WordSearch
  section. `get1828()` / `getModern()` are now async, returning the
  backend's `Def1828Response` / `ModernResponse` shapes (with
  `stem_matched` surfaced). Server-side stem fallback (D-DICT-2) means
  the client-side `ARCHAIC_SUFFIXES` block was deleted from the
  definition path; the tier-side stem matcher (a smaller copy) stays
  for tokenize so HighlightedText doesn't have to await. A 200-entry
  per-endpoint LRU coalesces repeat lookups in a session.
- `useLLMSettings.ts`: localStorage shape v1→v2. The `apiKey` field
  was DROPPED from persistence per D-LP-2 — keys live only in the
  Settings form input ref, never reactive-bound to llmSettings, never
  serialized. v1 readers get a one-time `migrationNeeded` banner via
  the `MIGRATION_BANNER_KEY` flag; v2 readers see nothing. Provider
  matrix expanded to seven (`openai`, `openrouter`, `opencode-go`,
  `opencode-zen`, `lm-studio`, `custom`, `mock`); LM Studio stays
  reserved for embeddings per D-LP-1 but is in the preset list for
  local dev convenience.
- `useLLMSession.ts` (new): `startSession()` / `endSession()` /
  `refreshSession()` / `isSessionActive()` + a `sessionActive`
  reactive ref other components observe. `credentials: 'include'` on
  every call; `Authorization: Bearer` mirror so curl smoke tests and
  cookie-stripped browsers both work. Module-load bootstraps
  `sessionActive` from localStorage + expiry check; Settings.vue
  calls `refreshSession()` on mount to confirm against the server
  (catches the janitor-evicted-while-idle case).
- `useLLMRender.ts`: no longer talks to the provider directly. POSTs
  to `/api/llm/render` with the session cookie + Bearer mirror.
  `RenderError` is a typed `{kind, message, retryAfterSeconds?}`
  shape now, not a free-form string. 401 surfaces a reauth state
  AND triggers `refreshSession()` to sync the mirror. 429 surfaces
  `rate_limited_by_1828` with the backend's message verbatim so the
  D-BE-AUTH attribution ("this is our cap, not theirs") is visible
  to the reader. 502 surfaces `upstream_provider_error` passed
  through (different body shape; the spec covers this in phase-2
  streaming). 503 surfaces `feature_disabled` for kill-switched
  deploys.
- `Settings.vue` rebuilt: top-of-page session-state card showing
  "active until …" with Sign out, OR "no active session" prompt.
  Transient API-key input lives in a local component ref ONLY
  (`apiKeyInput`); never bound to llmSettings; cleared after Start
  session regardless of success. The Start session button probes
  server-side and renders the backend's error message verbatim
  (`key_probe_failed`, `missing_model`, etc.). CORS section
  reframed as "now handled server-side" — the old browser→LM Studio
  dance is obsolete. The v1 migration banner sits above everything
  with a Dismiss action.
- `WordCard.vue` + `Present.vue` patched to the async pattern:
  `watch(props.word) → Promise.all([get1828, getModern]) → ref
  updates`. Loading states per section so "No 1828 entry on file"
  doesn't flash before the fetch lands. Both surfaces now render
  `stem_matched` ("showing entry for `obtain`") which the static
  bundle path never could — the backend's authoritative stem
  fallback finally has a user-facing surface.
- `VerseExplorer.vue` patched alongside the others to import
  `sessionActive` instead of the dropped `isConfigured`, call
  `refreshSession()` on mount, and render the new `RenderError`
  shape with kind-specific headlines + a "Re-authenticate in
  Settings" link for the reauth case.

**07bf9f3 — phase 5.C: canon-browse mode + WordSearch class-E reach.**
The two headline backend wins, surfaced at the UX layer.

- `VerseExplorer.vue` got a third mode tab: "Browse canon." Volume,
  book, and chapter selectors drive a `GET
  /api/scripture/chapter/{abbr}/{chapter}` call; the result renders
  inline through the existing `HighlightedText` component so tier-word
  highlights + click-to-open-WordCard work the same way the demo and
  paste modes do. Per D-BE-COPYRIGHT option D, the rendered chapter
  pairs with a "Full chapter at churchofjesuschrist.org ↗" breakout
  so the canonical apparatus (footnotes, study aids) is one click
  away. The mode switcher reads `demo / canon / paste`; existing
  demo + paste modes intact.
- `WordSearch.vue` got a secondary section: "Other 1828 words
  matching `{query}`." Activates when the search has ≥2 characters,
  calls `/api/dict/search`, surfaces `all_1828_results` that aren't
  already in the curated `tier_results`. Words deep-link to
  `/word/{word}` the same way curated words do — WordCard fetches
  the 1828 entry async via the dictionary backend. The "98,828
  headwords on file" framing is in the section's subtitle so the
  class-E reach property is visible, not just functional. Stale-query
  guard (`lastQueryRequested`) discards in-flight fetches when the
  user types past them.
- `data/canon-books.ts` (new): 80 books across 5 volumes (OT, NT,
  BoM, D&C, PGP) with `abbr` / `urlPath` / `chapters`, plus
  `buildChurchUrl()`. Hand-maintained mirror of the backend's abbr
  conventions; one-line PRs to keep in sync. When the backend
  grows `/api/scripture/books`, this becomes a fetch.

## What fought me

- TypeScript strict mode rejected `results[i]` indexing (possibly
  undefined). Rewrote as `forEach` with optional-chain.
- The workspace `.env.*` gitignore caught `.env.development` — needed
  `git add -f` to check it in. `.env` (with the real opencode key)
  stayed gitignored as it should.
- `vue-tsc` flagged `import.meta.env` missing types — added
  `vite-env.d.ts` with the `/// <reference types="vite/client" />`
  shim and an `ImportMetaEnv` interface.
- WordCard's `defModern === null` check in the template was a v1
  semantics leak — replaced with explicit `modernSource` /
  `modernError` reactive refs that distinguish loading from
  fetched-but-empty from rate-limited from network-error.
- Bash heredoc with backtick-rich commit message ate itself. Used a
  scratch file + `git commit -F` for both commits.

## Decisions I made outside the spec

- **`.env.development` checked in** (the project root's `.env.*`
  gitignore covers secrets; the dev URL is not a secret). Without
  this, every new developer has to recreate it from `.env.example`
  before `npm run dev` works.
- **Books list hand-maintained in the frontend** rather than fetched
  from the backend. The spec said "Don't touch backend code"; adding
  a `/api/scripture/books` endpoint would have been backend work.
  The 80-book CANON table is 100 lines and stable across decades.
- **`buildChurchUrl()` lives in `canon-books.ts`** rather than
  reusing `selectedVerse.church_url` from `demo-verses.json` because
  demo URLs are hand-curated with verse ranges; canon-mode URLs are
  generated from `{volumeUrl, bookUrl, chapter}` and don't yet have
  verse-range refinement (5.B carry-forward).
- **Migration banner pattern via `migrationNeeded` reactive ref +
  `MIGRATION_BANNER_KEY` localStorage flag**, dismissible per-browser.
  The spec said "one-time banner" but didn't specify mechanism.
- **`sessionActive` lives in `useLLMSession.ts` as a module-scoped
  ref** rather than a composable factory. Multiple components
  (VerseExplorer + Settings) need to observe the same value; a
  module ref is the cleanest shared-reactive pattern.
- **VerseExplorer canon mode does single-chapter render only** for
  v1. Verse-range refinement (e.g. `?id=22-23#22` after Abr 3:22-23
  is selected) is 5.B carry-forward.
- **Class-E reach gated at `query.length >= 2`** to avoid hammering
  the search endpoint on every keystroke from cold. The tier section
  shows immediately on first character.

## Steward-mode fixes folded in

- **`Present.vue` migrated to the async word-data pattern.** Wasn't
  in the orchestrator's spec, but Present.vue uses the same
  `get1828`/`getModern` API as WordCard — leaving it sync would have
  shipped a broken Presentation Mode page. Same fix, same shape,
  no behavior change. Plus it gained the `stem_matched` surface
  (showing "obtain" when the reader clicked "obtaineth") which was
  invisible in the static-bundle path.
- **Stale-query guard in WordSearch's class-E fetch** so fast
  typing doesn't render results from a query the user has moved
  past. Wasn't called out in spec; it's the "don't ship a flickering
  UI" baseline.
- **CORS section in Settings.vue reframed as "now handled
  server-side"** rather than deleting it. Returning readers who
  remember the LM Studio CORS dance will look for it and find an
  explanation of why it's no longer needed.
- **`Sign out` button on Settings labelled "Sign out (drop the held
  key)"** rather than just "Sign out." The user-mental-model gain
  matters — they're actively releasing a key from server memory, not
  just toggling a UI flag.

## What I deliberately didn't touch

- The legacy `1828` container on 8082 — restart/rebuild count: 0.
- `Dockerfile.legacy` — rollback escape hatch, untouched.
- Backend code (`backend/`) — phase-1-4's contracts are what the
  frontend depends on.
- `docker compose down -v` — never run.
- The `OPENCODE_GO_API_KEY` value — used it for smoke-test session
  mint via Bearer header without echoing to stdout.

## Phase 5 carry-forward (5.B refinements + phase 6)

- **5.B verse-range refinement** in canon mode. Adding a verse-start
  / verse-end pair of inputs that build
  `GET /api/scripture/{abbr}/{chapter}:{start}-{end}` and update
  `buildChurchUrl()` to carry `id={start}-{end}#{start}`.
- **5.B standalone WordSearch class-E page**? Currently class-E
  surfaces under the curated section in WordSearch. A dedicated
  `/dictionary/search` route for browsing the full 98k corpus
  alphabetically might be worth it — TBD whether that's more
  reading-frame help or more decoder-ring.
- **5.B `/api/scripture/books` backend endpoint** so the CANON table
  can become a fetch + cache, eliminating the hand-maintained
  duplication.
- **5.B word-study reverse lookup surface.** The backend's
  `/api/scripture/word-study/{word}` endpoint exists but no UI
  reaches it yet. A "see this word's verses" link on WordCard
  would be the natural mounting point.
- **Phase 6 Thummim cache sync.** Out of this session's scope; the
  backend's `thummim_entries_cache` table is empty awaiting the
  nightly snapshot job.
- **Phase 7 backup polish + production deploy** to Dokploy at
  1828.ibeco.me. The current stack on `localhost:8083` is ready;
  production rollout is gated on Michael's go-ahead.

## Final state of the stack

```
NAME             STATUS                   PORTS
i1828-backend    Up 35m (healthy)         8080/tcp (internal)
i1828-db         Up 54m (healthy)         5432/tcp (internal)
i1828-frontend   Up 1m  (healthy)         0.0.0.0:8083->80/tcp

legacy:
1828             Up 5h  (unhealthy*)      0.0.0.0:8082->80/tcp
                 *legacy healthcheck pre-dates this session
```

Frontend bundle: `index-DgUfsWZ3.js` is served live. `useWordData`
chunk dropped from ~1.5MB (with the static dict bundles) to 247KB.
The dictionary corpus is now reached on-demand at sub-100ms latency
from cache.

Verified end-to-end:
- `/healthz` returns ok on both 8083 (new) and 8082 (legacy)
- `/api/dict/1828/intelligence`, `/api/dict/1828/gainsay` (class-E),
  `/api/scripture/chapter/john/3?highlight=1`,
  `/api/scripture/chapter/dc/93` all return through nginx proxy
- `/api/dict/search?q=peradv` returns `peradventure` in
  `all_1828_results` with empty `tier_results`
- Session mint+render+inspect round-trip works with the `mock`
  provider; `Authorization: Bearer fake-session-id` returns
  `{active:false}` (inspect rejects unknown ids cleanly)

No `docker compose down -v` was ever run.
