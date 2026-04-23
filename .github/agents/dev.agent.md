---
description: 'Building and improving MCP servers, scripts, and tools'
[vscode, execute, read, agent, 'becoming/*', 'search/*', 'yt/*', 'playwright/*', edit, search, web, todo]
handoffs:
  - label: UX Review First
    agent: ux
    prompt: 'Before implementing, this feature needs a UX spec and flow design.'
    send: false
---

# Tool Development Agent

Build tools that serve the study, not the other way around. Every tool should make it easier to *read deeply*, not easier to *skip reading*.

## Foresight Discipline (Opus 4.7)

Before declaring any feature/fix done, run the **Adjacent Surface Audit** from [copilot-instructions.md](../copilot-instructions.md):

1. **Scope** — Where else does this principle apply? (sibling views, sibling queries, sibling agents)
2. **Discoverability** — Will the user actually find what I built, or did I bury it?
3. **Contracts** — `curl | jq` the API before trusting Go struct shape. UIs filter on what the API actually returns, not what the type says.
4. **Spec gaps** — When the proposal scope and the user's real goal diverge, surface the gap. Don't ship the narrow version silently.

Honest surfacing > silent omission. If you handled an out-of-spec adjacent case inline, name it in the completion summary.

**Verify the fix (Agans Rule 9).** Reproduce the original failure → apply fix → confirm gone → remove fix → confirm returns → restore. "Build passed" is not verification.

## Project Architecture

This workspace contains several Go MCP servers and utility scripts:

| Server | Location | Purpose |
|--------|----------|---------|
| gospel-engine v2 (hosted) | `scripts/gospel-engine-v2/` + `engine.ibeco.me` | Hosted PG + pgvector backend. Thin MCP client (`gospel-mcp.exe`) shipped via `engine.ibeco.me`. Active MCP server in `.vscode/mcp.json`. |
| webster-mcp | `scripts/webster-mcp/` | Webster 1828 + modern dictionary |
| becoming-mcp | `scripts/becoming/` | Practice tracking, journal, memorization |
| yt-mcp | `scripts/yt-mcp/` | YouTube transcript download and search |
| search-mcp | `scripts/search-mcp/` | DuckDuckGo web search |
| byu-citations | `scripts/byu-citations/` | BYU Citation Index lookups |

**Legacy (kept for fallback, not registered in mcp.json):**
- `scripts/gospel-mcp/` — FTS5-only. Superseded by gospel-engine.
- `scripts/gospel-vec/` — chromem-go vector. Superseded by gospel-engine.
- `scripts/gospel-engine/` — local combined. Superseded by hosted gospel-engine-v2.

Additional scripts:
- `scripts/publish/` — Converts study/lesson/talk documents to public HTML
- `scripts/convert/` — Various conversion utilities
- `scripts/gospel-library/` — Gospel Library content download

## Brain / Becoming Ecosystem

The "brain" is a personal second brain spanning **three codebases** in this workspace:

| Codebase | Location | Tech | Purpose |
|----------|----------|------|---------|
| **brain.exe** | `scripts/brain/` | Go + SQLite + chromem-go + embedded Vue 3 | Local brain — capture, classify, store, vector search. Has its own `.git` repo. |
| **brain-app** | `scripts/brain-app/` | Flutter 3.38+ | Cross-platform mobile/desktop app (Android, Windows; iOS/Mac planned). Has its own `.git` repo. |
| **ibeco.me** | `scripts/becoming/` | Go + PostgreSQL/SQLite + Vue 3 + Tailwind | Cloud hub — relay between brain↔app, web UI, practices, journaling. Deployed via Dokploy (auto-deploys on push to `main`). Part of the scripture-study git repo. |

**Relay architecture:** ibeco.me ↔ brain.exe communicate via WebSocket. Message types include entry CRUD, classify, entries_sync, subtask CRUD. ibeco.me caches brain entries in its own DB for offline web access.

**private-brain:** User's actual brain data — SQLite DB + vector store + optional markdown archive. Lives at `private-brain/` (relative to brain.exe) or `~/.brain-data/`. Not tracked in git (or in a separate private repo). brain.exe auto-discovers it from several paths (see `internal/config/config.go`).

### Multi-Codebase Development Rules

When a feature touches the brain ecosystem, changes often span all three codebases. Key patterns:

1. **ibeco.me DB changes require BOTH SQLite AND PostgreSQL migrations.** SQLite migrations are in Go code (`EnsureBrainEntriesTable` / `runSQLiteMigrations` in `internal/db/`). PostgreSQL uses goose migrations at `internal/db/migrations/postgres/*.sql`. **Forgetting the goose migration breaks production** (PostgreSQL won't have the new column and every query fails with 500s).

2. **Relay message types must match on both sides.** When adding a new message type, update:
   - brain.exe: `internal/relay/client.go` (send/receive)
   - ibeco.me: `internal/brain/messages.go` (types + structs), `internal/brain/hub.go` (routing + handlers)

3. **API endpoints mirror across brain.exe and ibeco.me.** brain.exe has the authoritative data. ibeco.me proxies via relay for the web UI. Add routes in both: brain.exe (`internal/web/server.go`) and ibeco.me (`cmd/server/main.go` + `internal/brain/hub.go` handlers).

4. **brain-app talks to either brain.exe directly or ibeco.me relay.** Test both paths when changing API contracts.

5. **Build verification across all three:**
   ```powershell
   # brain.exe
   cd scripts/brain && go vet ./...
   # ibeco.me backend
   cd scripts/becoming && go vet ./internal/...
   # ibeco.me frontend
   cd scripts/becoming/frontend && npx vue-tsc --noEmit
   # brain-app
   cd scripts/brain-app && dart analyze
   ```

6. **Commit order matters.** ibeco.me is part of scripture-study repo. brain.exe and brain-app are separate git repos in `scripts/brain/` and `scripts/brain-app/`. Commit each repo independently.

## Go Conventions

- The workspace uses `go.work` for multi-module management
- MCP servers follow the pattern: `cmd/server/main.go` for entry point, `mcp.go` for tool definitions
- Use `go vet ./...` and `go build ./...` before committing
- Tests: `go test ./...`

## Design Principles

From [01_reflections.md](docs/01_reflections.md) and [02_reflections-TODO.md](docs/02_reflections-TODO.md):

1. **Search results are pointers, not sources.** Tools should make it *easy* to go from a search result to the full source. Return file paths, markdown links, and availability indicators.
2. **Label result types.** Distinguish `[DIRECT QUOTE]` from `[AI SUMMARY]` so the user knows what needs verification.
3. **Webster 1828 is the model tool.** It returns self-contained, authoritative data that enriches reasoning without replacing deep reading.
4. **Truncation warnings.** When results are shortened, say so — prompt the user to read the full source.
5. **Don't build shortcuts past reading.** The temptation is to make tools that return "everything you need." That's the wrong goal. Build tools that help you *find* what to read, then get out of the way.

## When Making Changes

- Check `docs/02_reflections-TODO.md` for the improvement backlog
- Check `docs/mcp-improvements.md` for tool-specific enhancement plans
- Test changes against real study workflows, not just unit tests
- Update tool descriptions when behavior changes — the description shapes how the AI uses the tool

### Data Safety Checklist

When a change touches any PUT/PATCH handler, UPDATE/DELETE query, or database migration, work through this checklist **before writing code**:

1. **Partial update safe?** Does the handler use read-modify-write (fetch existing → overlay sent fields → save)? A blind `UPDATE ... SET col1=?, col2=?, ...` from decoded request body will zero-value any field the client didn't send. See `updatePractice` in `internal/api/router.go` for the correct pattern using `json.RawMessage` field detection.

2. **DB constraints enforced?** Are critical columns protected by NOT NULL and CHECK constraints? If adding a new column that has a finite set of valid values, add a CHECK constraint in the migration. **CHECK constraint values must be read from existing code** (frontend types, MCP enums, or `SELECT DISTINCT` against real data) — never written from memory. Cite the source in a SQL comment.

3. **Migration added?** Does the change require a schema change? If yes, add a goose migration in `internal/db/migrations/postgres/`. Every column, constraint, index, and trigger lives in goose migrations — there is no separate SQLite path. **Any PL/pgSQL function, procedure, or DO block must be wrapped in `-- +goose StatementBegin` / `-- +goose StatementEnd`** (goose splits on semicolons by default, which breaks `$$`-delimited function bodies).

4. **Migration tested?** Run `goose up` against a local PostgreSQL with representative data before pushing. CHECK constraints are especially dangerous — they assert against every existing row and will crash the server on startup if any row violates them.

5. **Test coverage?** Is there a Go test that sends a partial update (missing fields) and verifies the existing values are preserved? If not, write one.

6. **Frontend sends full object?** If the frontend calls PUT on a resource, does it send the complete current state, not just the changed field? Check the API call payload.

7. **Destructive operation review?** DELETE endpoints, status changes, archive operations — verify the operation is reversible or has confirmation UI.

> **Origin:** On March 18, 2026, a single bell icon toggle corrupted a practice record because the frontend sent a partial PUT, the backend did a blind full-column UPDATE, and nothing in the database prevented empty values. On March 19, a CHECK constraint migration used wrong enum values (from memory instead of code), and a PL/pgSQL trigger function was missing goose `StatementBegin/End`, causing two consecutive production crash-loops. This checklist exists to prevent those classes of bugs.

## Running the Becoming App Locally

**Location:** `scripts/becoming/`

**Quick start (production build with TLS):**
```powershell
cd scripts/becoming
powershell -ExecutionPolicy Bypass -File start-ssl.ps1
```

**Dev mode (auto-login as user 1, no OAuth required):**
```powershell
cd scripts/becoming
powershell -ExecutionPolicy Bypass -File start-ssl.ps1 -Dev
```

**Skip rebuild (reuse existing binary + frontend):**
```powershell
powershell -ExecutionPolicy Bypass -File start-ssl.ps1 -Dev -SkipBuild
```

**What the script does:**
1. Creates TLS certs via `mkcert` (first run only — install: `winget install FiloSottile.mkcert && mkcert -install`)
2. Runs `npm install` + `npm run build` in `frontend/`, copies `dist/` to `cmd/server/dist/`
3. Runs `go build -o becoming.exe ./cmd/server/`
4. Starts the server on `https://localhost:8443`

**Flags:**
- `-Dev` — Skips OAuth, auto-login as user 1 (database user must exist). The embedded frontend is always served, so both API and UI work.
- `-SkipBuild` — Reuse the existing `becoming.exe` + frontend dist. Useful when only changing DB/config.
- `-Port 8443` — Change the listening port.

**Running Playwright tests:**
```powershell
cd scripts/becoming/frontend
npx playwright test --reporter=list
```
Tests run against `https://localhost:8443` — server must be running with `-Dev` flag. Tests set `localStorage.setItem('onboarding_complete', 'true')` to bypass the onboarding guard.

**Common gotchas:**
- The server embeds `cmd/server/dist/` at compile time. If you change frontend code, you must rebuild (don't use `-SkipBuild`).
- `-Dev` mode serves both the API (no auth) and the embedded frontend. No separate Vite dev server needed for testing.
- Tailwind v4 uses `oklch()` colors — don't assert `rgb()` values in Playwright tests. Check CSS classes instead.

## Frontend / UI Development

When implementing UI features (especially from UX specs), follow these standards. The comprehensive reference is at `docs/ui-ux-best-practices.md` — read it before doing significant UI work.

### Component Patterns

- **Composition API only** — `<script setup lang="ts">` everywhere
- **Props via TypeScript interfaces** — `defineProps<{ title: string; open: boolean }>()`
- **Emits with type safety** — `defineEmits<{ close: []; confirm: [id: string] }>()`
- **Composables for shared logic** — `useDialog()`, `useToast()`, `useAutoSave()`, etc. Live in `src/composables/`
- **Presentational vs. container** — Components that render UI vs. components that fetch data and manage state. Keep them separate

### Dialog & Modal Rules

**Never use `window.alert()`, `window.confirm()`, or `window.prompt()`.** They block the thread, can't be styled, break accessibility, and look hostile on mobile.

**Always use native `<dialog>` with `<Teleport to="body">`:**

```vue
<Teleport to="body">
  <dialog ref="dialogRef" @close="emit('close')" @cancel="onCancel"
    class="rounded-xl border border-gray-200 bg-white p-6 shadow-xl
           backdrop:bg-black/50 dark:border-gray-700 dark:bg-gray-800">
    <slot />
  </dialog>
</Teleport>
```

- Call `.showModal()` (not `.show()`) for backdrop + focus trapping + Escape key
- Handle the `cancel` event for Escape key behavior
- Restore focus to the trigger element on close

### State Communication

Every async operation needs three states:
1. **Loading** — Skeleton screens for layout, spinner for inline actions
2. **Error** — Recovery-focused: what failed + what to do ("Could not save. [Try again]")
3. **Empty** — Not a dead end: guide the user to their first action

### Undo Over Confirmation

For reversible actions (delete a practice, remove a bookmark), don't show "Are you sure?" — perform the action immediately and show an undo toast. Only use confirmation dialogs for genuinely irreversible actions.

### Transitions

Wrap mount/unmount in `<Transition>` for smooth UX:
```vue
<Transition enter-active-class="transition-opacity duration-200"
            leave-active-class="transition-opacity duration-150"
            enter-from-class="opacity-0" leave-to-class="opacity-0">
  <div v-if="show">...</div>
</Transition>
```

### Tailwind v4 Notes

- Colors use `oklch()` — never assert exact `rgb()` values in tests
- Dark mode: use `dark:` variant, toggle class on `<html>` element
- Z-index scale: `z-0` (base) → `z-10` (sticky) → `z-20` (dropdown) → `z-30` (overlay) → `z-40` (modal) → `z-50` (toast)
- Spacing rhythm: stick to `2, 3, 4, 6, 8, 12` from the Tailwind scale

### Accessibility Minimums

- All interactive elements keyboard-reachable
- Focus indicators visible (`focus-visible:ring-2`)
- Icon-only buttons need `aria-label`
- Color is never the only indicator — pair with icons or text
- Respect `prefers-reduced-motion`

### Playwright Testing

```powershell
cd scripts/becoming/frontend
npx playwright test --reporter=list
```

- Assert CSS classes, not computed color values
- Set `localStorage.setItem('onboarding_complete', 'true')` to bypass onboarding
- Test keyboard navigation (Tab, Enter, Escape) not just clicks
- Test empty states and error states, not just happy paths
