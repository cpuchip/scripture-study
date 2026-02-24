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

## Project Architecture

This workspace contains several Go MCP servers and utility scripts:

| Server | Location | Purpose |
|--------|----------|---------|
| gospel-mcp | `scripts/gospel-mcp/` | FTS5 full-text search of gospel library |
| gospel-vec | `scripts/gospel-vec/` | Semantic vector search with chromem-go |
| webster-mcp | `scripts/webster-mcp/` | Webster 1828 + modern dictionary |
| becoming-mcp | `scripts/becoming/` | Practice tracking, journal, memorization |
| yt-mcp | `scripts/yt-mcp/` | YouTube transcript download and search |
| search-mcp | `scripts/search-mcp/` | DuckDuckGo web search |

Additional scripts:
- `scripts/publish/` — Converts study/lesson/talk documents to public HTML
- `scripts/convert/` — Various conversion utilities
- `scripts/gospel-library/` — Gospel Library content download

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
