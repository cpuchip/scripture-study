# WS3: Brain UX Quality-of-Life Improvements

**Workstream:** WS3 (Brain UX)
**Status:** in-progress (Phase 1 ✅, Phase 5 ✅, Phase 2 ✅, Phase 3 ✅, Phase 4 ✅, Phase 6 ✅, Phase 7 ✅, Phase 7a ✅, Phase 7b specced, Phase 8 deferred)
**Binding problem:** After an agent auto-advances an entry, the user can't see what was generated, can't tell what the agent needs next, and has to switch to VS Code to read the output. The reply textbox is too small for substantive responses. There's no real-time feedback. The brain app pipeline works mechanically but is opaque to the user.

**Discovered:** 2026-04-05, during first real review cycle on "Build Physical Display Dashboard" entry.

---

## Success Criteria

1. User can type multi-paragraph replies without fighting a tiny textbox
2. Agent messages render as markdown (headings, links, code blocks, lists)
3. File paths in agent messages are clickable — clicking opens the rendered file inline
4. User can browse workspace files from within the brain UI
5. UI updates in real-time when agents post messages or change entry status
6. User can see cost (premium requests) consumed per entry

## Constraints

- Each phase must deliver value independently — no "infrastructure now, value later"
- File serving endpoint must have path traversal protection (OWASP)
- Use markdown-it (matches ibeco.me, plugin ecosystem)
- Port components from ibeco.me's ReaderView where applicable — don't rebuild
- One session per phase — keep scope tight

## Prior Art

- **ibeco.me ReaderView** (`becoming/frontend/src/views/ReaderView.vue`) — file tree sidebar, markdown rendering, bookmarks, search filter. Uses GitHub API for data, but the UI components (TreeNode.vue, markdown rendering) are portable.
- **Brain inline panel spec** (`.spec/proposals/brain-inline-panel.md`) — already specked textarea for slide-out panel. Same `rows="2"` problem applies there too.
- **Relay WebSocket** (`brain/internal/relay/client.go`) — gorilla/websocket v1.5.3 already in go.mod. Only outbound to ibeco.me currently.
- **Token tracking** — `tokens_used INT64` on entries table, set once at route completion. No per-interaction granularity.

---

## Phase 1: See What the Agent Did ✅ COMPLETE (Apr 5)

*Shipped: auto-expanding textarea, markdown-it rendering, clickable file paths with backslash normalization, FileViewer sidebar panel (not modal), content shift via shared reactive useFilePanel, external links target=_blank, onUnmounted cleanup.*

### 1a. Auto-Expanding Textarea

Replace fixed `rows="2"` + `resize-none` with an auto-expanding textarea.

**Implementation:**
```typescript
// composable: useAutoExpand.ts
function useAutoExpand(maxHeight = 300) {
  const el = ref<HTMLTextAreaElement | null>(null)
  const resize = () => {
    if (!el.value) return
    el.value.style.height = 'auto'
    el.value.style.height = Math.min(el.value.scrollHeight, maxHeight) + 'px'
  }
  return { el, resize }
}
```

**Apply to:**
- EntryDetailView.vue (line 428) — existing reply textarea
- ProjectDetailView.vue slide-out panel — when inline reply is built (Phase 1 of inline panel spec)

**Change:** Remove `resize-none`, add `@input="resize"`, bind `ref="el"`, set `min-rows="2"`.

**Effort:** ~20 lines (composable + template changes)

### 1b. Markdown Rendering in Messages

**Add dependency:** `npm install markdown-it`

**Create:** `src/composables/useMarkdown.ts`
```typescript
import MarkdownIt from 'markdown-it'

const md = new MarkdownIt({
  html: false,        // no raw HTML (security)
  linkify: true,      // auto-link URLs
  breaks: true,       // newlines become <br>
})

export function renderMarkdown(text: string): string {
  return md.render(text)
}
```

**Change message rendering in EntryDetailView.vue:**

From:
```html
<div class="whitespace-pre-wrap text-gray-300">{{ msg.content }}</div>
```

To:
```html
<div class="prose prose-invert prose-sm max-w-none text-gray-300" v-html="renderMarkdown(msg.content)" />
```

**Security:** `html: false` prevents XSS. `linkify: true` auto-links URLs. Content is from our own agent (not user-generated external content), but defense-in-depth matters.

**Effort:** ~30 lines (composable + template changes + Tailwind prose plugin)

### 1c. Clickable File Paths + Inline Viewer

**Backend — new endpoint:**

`GET /api/files/read?path=.spec/scratch/build-physical-display-dashboard/main.md`

```go
func (s *Server) handleFileRead(w http.ResponseWriter, r *http.Request) {
    relPath := r.URL.Query().Get("path")
    if relPath == "" {
        http.Error(w, "path required", 400)
        return
    }

    // Security: clean and jail to workspace root
    cleaned := filepath.Clean(relPath)
    if filepath.IsAbs(cleaned) || strings.Contains(cleaned, "..") {
        http.Error(w, "invalid path", 403)
        return
    }

    fullPath := filepath.Join(s.workspaceRoot, cleaned)
    // Verify resolved path is under workspace root
    if !strings.HasPrefix(filepath.Clean(fullPath), filepath.Clean(s.workspaceRoot)) {
        http.Error(w, "access denied", 403)
        return
    }

    data, err := os.ReadFile(fullPath)
    if err != nil {
        http.Error(w, "not found", 404)
        return
    }

    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.Write(data)
}
```

**Frontend — file path detection:**

Add a markdown-it plugin or post-process the rendered HTML to detect workspace-relative paths (`.spec/...`, `study/...`, `scripts/...`) and convert them to clickable links that open a viewer modal.

**Frontend — FileViewer modal component:**

```vue
<!-- FileViewer.vue -->
<template>
  <div v-if="open" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60">
    <div class="bg-gray-900 border border-gray-700 rounded-xl w-[80vw] h-[80vh] flex flex-col">
      <div class="flex items-center justify-between px-4 py-3 border-b border-gray-700">
        <span class="text-sm text-gray-300 font-mono">{{ filePath }}</span>
        <button @click="$emit('close')" class="text-gray-500 hover:text-gray-300">✕</button>
      </div>
      <div class="flex-1 overflow-auto p-6 prose prose-invert prose-sm max-w-none"
           v-html="renderedContent" />
    </div>
  </div>
</template>
```

**Effort:** ~60 lines backend, ~80 lines frontend (endpoint + detection + modal + rendering)

### Phase 1 Verification ✅

- [x] Type a 5-paragraph reply — textarea grows to accommodate, caps at max height, scrolls after
- [x] Agent message with markdown headings/lists renders properly (not raw `##` text)
- [x] Agent message "Findings at .spec/scratch/foo/main.md" → "Findings at" is text, path is a clickable link
- [x] Click file path → sidebar panel opens with rendered markdown content
- [x] Path traversal attempt (`../../etc/passwd`) returns 403
- [x] Panel has close button, scrolls for long files

---

## Phase 2: File Browser ✅ COMPLETE (Apr 6)

*Shipped: GET /api/files/tree endpoint with security validation, TreeNode.vue recursive component, Library files tab with search filter, wide layout (max-w-6xl) when browsing files, .spec auto-expanded on load.*

### 2a. Backend — File Tree Endpoint

`GET /api/files/tree?root=.spec`

Returns:
```json
[
  { "name": "scratch/", "path": ".spec/scratch/", "children": [
    { "name": "brain-simplification/", "path": ".spec/scratch/brain-simplification/", "children": [...] },
    { "name": "build-physical-display-dashboard/", "path": "...", "children": [...] }
  ]},
  { "name": "proposals/", "path": ".spec/proposals/", "children": [...] }
]
```

**Security:** Same path validation as file read. Only serve under workspace root.

### 2b. Frontend — Library View

New route: `/library` (or `/files`)

Port from ibeco.me:
- `TreeNode.vue` — directory tree with expand/collapse
- Search/filter input
- Content area with markdown rendering (reuse from Phase 1)

**Not porting:** Bookmarks, share modal, practice creation (ibeco.me-specific)

### Phase 2 Verification

- [x] Navigate to `/library` → see tree of workspace files
- [x] Click a directory → expands to show children
- [x] Click a `.md` file → renders in content area
- [x] Search filter narrows tree results
- [x] Nav header shows "Library" as a top-level section

---

## Phase 3: Real-Time Updates (WebSockets) ✅ COMPLETE (Apr 6)

*Shipped: Hub with gorilla/websocket, GET /ws endpoint with localhost-only origin check, Event broadcasts from HTTP handlers (create/update/reply/complete/advance) and pipeline background goroutines (research/plan/execute/verify/nudge). Frontend useWebSocket composable with auto-reconnect + exponential backoff. Dashboard, EntryDetailView, and ProjectDetailView all receive live updates.*

### 3a. Backend — WebSocket Server

```go
// internal/web/hub.go
type Hub struct {
    clients    map[*websocket.Conn]bool
    broadcast  chan []byte
    register   chan *websocket.Conn
    unregister chan *websocket.Conn
}
```

**Endpoint:** `GET /ws` — upgrades to WebSocket

**Events broadcast:**
- `{ "type": "message.new", "entry_id": "...", "message": {...} }`
- `{ "type": "entry.updated", "entry": {...} }`
- `{ "type": "agent.started", "entry_id": "..." }`
- `{ "type": "agent.completed", "entry_id": "...", "result": "..." }`

**Integration points:**
- `AddSessionMessage()` → broadcasts `message.new`
- `UpdateEntry()` → broadcasts `entry.updated`
- Pipeline `Advance()` start/end → broadcasts agent events

### 3b. Frontend — WebSocket Composable

```typescript
// composables/useWebSocket.ts
export function useWebSocket() {
  const ws = ref<WebSocket | null>(null)
  const connect = () => { /* auto-reconnect logic */ }
  const on = (type: string, handler: (data: any) => void) => { /* subscribe */ }
  return { connect, on }
}
```

**Usage in views:**
- EntryDetailView: auto-append new messages
- ProjectDetailView: live-update entry badges
- DashboardView: replace polling with WebSocket events

### Phase 3 Verification

- [x] Open entry detail → agent runs in background → new message appears without refresh
- [x] Open project board → agent advances entry → badge updates live
- [x] Close browser tab → reopen → WebSocket reconnects
- [x] Multiple browser tabs → all receive updates

---

## Phase 4: Cost Tracking ✅ COMPLETE (Apr 6)

*Know what the pipeline is spending.*

*Shipped: `premium_requests_used REAL` column on entries table, atomic IncrementPremiumRequests after every pipeline agent call (research=0.33, plan=1.0, execute=1.0, nudge=0.33), badge in EntryDetailView metadata row, aggregate total in ProjectDetailView header.*

### 4a. Track Premium Requests Per Entry

Add to entries table (or new table):
```sql
ALTER TABLE entries ADD COLUMN premium_requests_used REAL DEFAULT 0;
```

**Instrument:** After each agent call (research, plan, nudge), increment by the model's multiplier:
- Haiku (research, nudge) = 0.33
- Sonnet (plan) = 1.0
- Opus = 3.0

### 4b. Display in UI

Entry detail view — small badge:
```
🎟️ 1.33 premium requests used
```

Project rollup:
```
💰 Total: 8.6 premium requests across 12 entries
```

### Phase 4 Verification

- [x] Research pass increments premium_requests_used by 0.33
- [x] Plan pass increments by 1.0
- [x] Entry detail shows the running total
- [x] Project detail shows aggregate across entries

---

## Phase 6: Reader UX — Links, Navigation, History ✅ COMPLETE (Apr 6)

*Shipped: FILE_PATH_RE lookbehind fix (added `>` and backtick for code spans), click handlers on Library and FileViewer content areas for `.file-link` navigation, `/library?file=path` deep linking with route watcher for same-page navigation, "Open in Reader →" button in FileViewer header, full back/forward navigation history with `fileHistory[]` + `historyIndex`, `openFileFromQuery()` helper that expands parent dirs and switches to files tab.*

**Origin:** Backtick-wrapped file paths aren't detected as links. File links in entry detail open a sidebar but can't follow internal links. The Library reader has no navigation history and no deep linking. These gaps prevent the brain from being the single interface for project awareness.

### 6a. Fix File Path Detection in Code Spans

**Bug:** `FILE_PATH_RE` lookbehind is `(?:^|\s|["'(])`. Markdown-it converts `` `scripts/brain/foo.go` `` to `<code>scripts/brain/foo.go</code>`. The regex runs on rendered HTML, so:
- The backtick itself is gone (consumed by markdown-it)
- The path is now preceded by `>` (from `<code>`)
- `>` is not in the lookbehind character class

**Fix:** Add `>` to the lookbehind: `(?:^|\s|["'(>])`. This catches paths inside `<code>` tags. Also add backtick for edge cases where raw backticks survive (non-markdown contexts): `(?:^|\s|["'(>\x60])`.

**Effort:** One character class change in `useMarkdown.ts`.

### 6b. Internal Link Following in Reader Content

Add click handlers to rendered content in both `LibraryView.vue` and `FileViewer.vue`:

```typescript
// LibraryView — content area click handler
function handleContentClick(e: MouseEvent) {
  const target = e.target as HTMLElement
  if (target.classList.contains('file-link') && target.dataset.filePath) {
    e.preventDefault()
    openFile(target.dataset.filePath)
  }
}
```

**Apply to:**
- Library reader content area (`@click="handleContentClick"`)
- FileViewer.vue rendered content (same pattern — clicking a link in a file opens the next file)

### 6c. Route-Based Deep Linking

Add `file` query parameter to the `/library` route:

```typescript
// Router: /library supports ?file=path
{ path: '/library', component: LibraryView }

// LibraryView — on mount, check for query param
onMounted(() => {
  const filePath = route.query.file as string
  if (filePath) openFile(filePath)
})
```

**Cross-view navigation:** Keep the FileViewer sidebar for quick peeks. Add an "Open in Reader" link (upper-right of the sidebar header, next to the close button) that navigates to `/library?file=path` for full-context reading. This gives both: fast inline access AND the full reader when you want it.

### 6d. Navigation History (Back/Forward)

```typescript
const fileHistory = ref<string[]>([])
const historyIndex = ref(-1)

function navigateToFile(path: string) {
  // Trim forward history
  fileHistory.value = fileHistory.value.slice(0, historyIndex.value + 1)
  fileHistory.value.push(path)
  historyIndex.value++
  loadFile(path)
}

const canGoBack = computed(() => historyIndex.value > 0)
const canGoForward = computed(() => historyIndex.value < fileHistory.value.length - 1)
```

**UI:** Back/forward arrows in the file content header bar, next to the file path. Disabled state when at ends of history. Breadcrumb trail optional (could show last 3-5 files).

### Phase 6 Verification

- [x] `` `scripts/brain/foo.go` `` renders as a clickable link (code span detection)
- [x] Click a file path in Library reader content → opens that file in reader
- [x] Navigate to `/library?file=.spec/proposals/brain-ux-quality-of-life.md` → file opens directly
- [x] Click file link in entry detail → navigates to Library reader
- [x] Open file A → click link to file B → click link to file C → back button returns to B → forward returns to C
- [x] Back button disabled when at start of history

---

## Phase 7: Git Status in File Browser ✅ COMPLETE (Apr 6)

*Shipped: `GET /api/git/status` endpoint using `git status --porcelain`, parsed into `{path, status}` JSON. Frontend: `gitStatusMap` loaded on mount and tab activation, passed to TreeNode. TreeNode shows colored dots (green=new, yellow=modified, red=deleted) for files and inherits most severe status for directories. Summary bar above file tree shows counts (`3 new, 2 modified`) and toggles a "show only changed" filter. No polling — refreshes when files tab activates.*

### 7a. Backend — Git Status Endpoint

`GET /api/git/status` — runs `git status --porcelain` in the workspace root, parses output, returns JSON.

```go
type GitFileStatus struct {
    Path   string `json:"path"`
    Status string `json:"status"` // "new", "modified", "deleted", "renamed"
}

func (s *Server) handleGitStatus(w http.ResponseWriter, r *http.Request) {
    cmd := exec.Command("git", "status", "--porcelain")
    cmd.Dir = workspaceRoot
    out, err := cmd.Output()
    // Parse: "?? file" → new, " M file" → modified, " D file" → deleted
    // Return: [{ "path": "study/foo.md", "status": "new" }, ...]
}
```

**Security:** Localhost-only server. Workspace-scoped. No user input in the git command (no path injection). Safe.

### 7b. Frontend — Status Indicators in TreeNode

Fetch git status on Library mount. Pass status map as prop to TreeNode.

**Visual indicators:**
- 🟢 Green dot — new/untracked file
- 🟡 Yellow dot — modified
- 🔴 Red dot — deleted
- Directory inherits the "most severe" status of its children

```vue
<!-- TreeNode.vue addition -->
<span v-if="statusIndicator" class="ml-1 text-xs" :class="statusClass">●</span>
```

### 7c. Stretch — Summary Bar

Above the file tree, show a summary line: "3 new, 2 modified" with counts. Clickable to filter tree to only changed files.

### Phase 7 Verification

- [x] Create a new file → green dot appears in file tree (after refresh)
- [x] Modify an existing file → yellow dot appears
- [x] Summary bar shows correct counts
- [x] Git status refreshes on Library tab activation (not continuously polling)

---

## Phase 7a: Inline Diff Viewer ✅ COMPLETE (Apr 6)

*See what changed without leaving the brain. Toggle between rendered content and diff view for any file with git changes.*

**Origin:** Phase 7 shows which files changed (dots + summary bar), but not what changed. To review agent work or your own edits, you currently have to switch to VS Code or a terminal. The diff should live where the reading already happens.

### 7a-1. Backend — Git Diff Endpoint

`GET /api/git/diff?path=<file>` — returns the unified diff for a single file.

```go
func (s *Server) handleGitDiff(w http.ResponseWriter, r *http.Request) {
    // Same workspace root derivation as handleGitStatus
    pathParam := r.URL.Query().Get("path")
    // Path traversal protection (same as handleFileRead)
    
    // For tracked modified files: git diff HEAD -- <path>
    // For untracked new files: git diff --no-index /dev/null <path>
    // Return raw unified diff text (Content-Type: text/plain)
}
```

**Design decisions:**
- `git diff HEAD -- <path>` shows both staged and unstaged changes vs last commit. Simpler than separate staged/unstaged views.
- For new (untracked) files: `git diff --no-index /dev/null <path>` creates a proper unified diff showing all lines as additions. Alternatively, fabricate a minimal diff header + all-green output.
- Path safety: reuse the same traversal protection from `handleFileRead` (resolve absolute, check prefix).
- Returns raw text, not JSON — the frontend passes it directly to diff2html.

**Route:** `s.mux.HandleFunc("GET /api/git/diff", s.cors(s.handleGitDiff))`

### 7a-2. Frontend — diff2html Integration

**New dependency:** `diff2html` (npm, MIT, 434K weekly downloads, 2 deps). Supports `git diff` output directly.

```bash
npm install diff2html
```

**Usage in LibraryView:**
```typescript
import { html as diff2html } from 'diff2html'
import 'diff2html/bundles/css/diff2html.min.css'

const showDiff = ref(false)
const diffContent = ref('')
const diffLoading = ref(false)
const diffMode = ref<'line-by-line' | 'side-by-side'>('line-by-line')

// Computed: is current file changed?
const currentFileChanged = computed(() => gitStatusMap.value.has(currentFilePath.value))

async function loadDiff() {
  diffLoading.value = true
  try {
    diffContent.value = await api.gitDiff(currentFilePath.value)
  } catch (e: any) {
    diffContent.value = ''
  } finally {
    diffLoading.value = false
  }
}

const renderedDiff = computed(() => {
  if (!diffContent.value) return ''
  return diff2html(diffContent.value, {
    outputFormat: diffMode.value,
    drawFileList: false,
    matching: 'lines',
    colorScheme: 'dark',
  })
})
```

**API addition:**
```typescript
async gitDiff(path: string): Promise<string> {
  const res = await fetch(`/api/git/diff?path=${encodeURIComponent(path)}`)
  if (!res.ok) throw new Error(`${res.status}: ${res.statusText}`)
  return res.text()
}
```

### 7a-3. UI — Toggle Button + Diff Rendering

**Header bar** (where ← → and file path already live):

```vue
<!-- Only show when file has git changes -->
<button
  v-if="currentFileChanged"
  @click="toggleDiff"
  class="text-xs px-2 py-0.5 rounded transition-colors"
  :class="showDiff
    ? 'bg-yellow-900/50 text-yellow-300'
    : 'text-gray-500 hover:text-gray-300'"
>
  {{ showDiff ? '✕ Diff' : 'Δ Diff' }}
</button>

<!-- Mode toggle (only when diff is showing) -->
<button
  v-if="showDiff"
  @click="diffMode = diffMode === 'line-by-line' ? 'side-by-side' : 'line-by-line'"
  class="text-xs text-gray-500 hover:text-gray-300 px-1"
>
  {{ diffMode === 'line-by-line' ? '⇔' : '⇕' }}
</button>
```

**Content area** — swap between rendered markdown and diff:

```vue
<div class="flex-1 overflow-auto p-6" @click="handleContentClick">
  <!-- Normal view -->
  <div v-if="!showDiff" ...>
    <!-- existing markdown rendering -->
  </div>
  <!-- Diff view -->
  <div v-else>
    <div v-if="diffLoading" class="text-gray-500 text-sm">Loading diff...</div>
    <div v-else-if="!diffContent" class="text-gray-600 text-sm">No changes</div>
    <div v-else v-html="renderedDiff" />
  </div>
</div>
```

**Behavior:**
- Toggle button is only visible when `currentFileChanged` is true (file appears in `gitStatusMap`)
- Clicking "Δ Diff" fetches the diff (lazy — not pre-fetched for every file) and shows it
- Clicking "✕ Diff" returns to the rendered markdown view
- `showDiff` resets to `false` when navigating to a different file
- Mode toggle switches between line-by-line (unified) and side-by-side (split)

### 7a-4. CSS — Dark Theme Alignment

diff2html ships its own CSS. The dark color scheme (`colorScheme: 'dark'`) handles most of it, but may need minor overrides to blend with the gray-900 background:

```css
/* If needed — scope to our container */
.d2h-wrapper {
  --d2h-dark-bg: theme('colors.gray.900');
}
```

Test in browser first before adding overrides. The `colorScheme: 'dark'` option may be sufficient.

### Phase 7a Verification

- [x] modified file → "Δ Diff" button appears in header bar
- [x] click "Δ Diff" → unified diff renders with green/red lines
- [x] toggle to side-by-side → two-column diff view
- [x] navigate to different file → diff view resets to normal view
- [x] untracked (new) file → diff shows all lines as additions
- [x] file with no changes → no diff button shown

**Shipped:** Apr 6, 2026. Backend: `handleGitDiff` handler (~65 lines, full path safety). Frontend: `diff2html` with `ColorSchemeType.DARK`, toggle buttons, lazy loading, mode switch. LibraryView JS chunk grew from 11KB to 54KB (includes diff2html).

---

## Phase 7b: Nested Git Repo Awareness

*Make the brain aware of all 13+ nested git repos in the workspace. Show diffs for subrepo files, and mark repo root directories in the file tree.*

**Origin:** The workspace contains nested repos (`scripts/brain/`, `teaching/`, `private/`, `external_context/*`, etc.). Phases 7 and 7a only query the workspace root's git, so all subrepo changes are invisible. The most actively developed repo (brain itself) has no git status or diff support in the UI.

### 7b-1. Backend — Repo Discovery + Aggregated Git Status

**Repo discovery function** — walks workspace root (depth ≤ 4) looking for `.git` directories. Returns a list of repo paths relative to workspace root. Cached per request (or per `handleGitStatus` call).

```go
func discoverGitRepos(workspaceRoot string) []string {
    // Walk looking for .git dirs. Returns relative paths like:
    // "." (workspace root), "scripts/brain", "teaching", "private", etc.
    // Skip dirs already in skipDirs that won't appear in tree (node_modules, .venv, etc.)
}
```

**`handleGitStatus` changes:**
- Call `discoverGitRepos` to find all repos
- For each repo, run `git status --porcelain` in that directory
- Prefix each file's path with the repo's relative path (for non-root repos)
- Add `repo` field to response: `{"path": "internal/web/server.go", "status": "modified", "repo": "scripts/brain"}`
- Root repo files have `repo: "."` 
- Return unified list across all repos

Updated response type:
```go
type GitFileStatus struct {
    Path   string `json:"path"`
    Status string `json:"status"`
    Repo   string `json:"repo"`  // relative path to repo root, "." for workspace root
}
```

### 7b-2. Backend — Repo-Aware Git Diff

**`handleGitDiff` changes:**
- Given a file path (e.g., `scripts/brain/internal/web/server.go`), determine which repo owns it
- Walk up from the file path to find which discovered repo prefix matches
- Run `git diff` in that repo's root, with the path *relative to that repo* (e.g., `internal/web/server.go`)
- Same `/dev/null` fallback for untracked files

```go
func findRepoForPath(repos []string, filePath string) (repoRoot string, relPath string) {
    // Find the longest matching repo prefix
    // e.g., "scripts/brain/internal/web/server.go" → repo="scripts/brain", rel="internal/web/server.go"
}
```

### 7b-3. Backend — File Tree Git Repo Indicator

**`handleFileTree` changes:**
- When building tree nodes for directories, check if directory contains a `.git` subdirectory
- Add `IsGitRepo bool` to TreeNode struct

```go
type TreeNode struct {
    Name      string      `json:"name"`
    Path      string      `json:"path"`
    IsDir     bool        `json:"is_dir"`
    IsGitRepo bool        `json:"is_git_repo,omitempty"`
    Children  []*TreeNode `json:"children,omitempty"`
}
```

The check is simple: `os.Stat(filepath.Join(dir, name, ".git"))` succeeds → `IsGitRepo: true`. No recursive walk needed here since we're already iterating children.

### 7b-4. Frontend — FileTreeNode + TreeNode.vue

**`FileTreeNode` interface update:**
```typescript
export interface FileTreeNode {
  name: string
  path: string
  is_dir: boolean
  is_git_repo?: boolean
  children?: FileTreeNode[]
}
```

**`GitFileStatus` interface update:**
```typescript
export interface GitFileStatus {
  path: string
  status: 'new' | 'modified' | 'deleted' | 'renamed'
  repo: string
}
```

**`TreeNode.vue` changes:**
- Directory entries with `is_git_repo` get a repo indicator badge (small icon or label)
- Something like: `<span v-if="node.is_git_repo" class="text-[9px] text-violet-400 ml-1">⎇</span>` (branch symbol) or a small "git" badge
- Git status still works the same — the Map keys are still full workspace-relative paths

**`LibraryView.vue` changes:**
- `loadGitStatus()` — the response now includes `repo` field, but `gitStatusMap` is still keyed by full path (same as before since backend returns full workspace-relative paths)
- No other changes needed — the aggregation happens on the backend

### 7b-5. Git Status Path Convention

The backend needs a clear path convention:

- **Git status paths** are always workspace-relative: `scripts/brain/internal/web/server.go`
- **Git diff** receives workspace-relative path, backend resolves to correct repo + relative path
- **Tree node paths** are already workspace-relative

For the root repo, `git status --porcelain` already returns workspace-relative paths. For subrepos, the backend must prepend the repo prefix: if `scripts/brain/` reports `internal/web/server.go`, the unified path becomes `scripts/brain/internal/web/server.go`.

### Phase 7b Verification

- [ ] `scripts/brain/` directory shows git repo indicator (⎇) in file tree
- [ ] files changed in brain subrepo appear in git status summary bar
- [ ] clicking a changed brain file → "Δ Diff" button works, shows correct diff
- [ ] workspace root changes still work as before
- [ ] multiple repos can show changes simultaneously
- [ ] repo indicator visible for `teaching/`, `private/`, `external_context/*` directories

### Effort Estimate

- Backend: ~80 lines (repo discovery ~25, status aggregation ~25, diff routing ~15, tree flag ~5, helper ~10)
- Frontend: ~10 lines (FileTreeNode field, TreeNode badge, GitFileStatus repo field)
- Scope: one session. No database changes. No new dependencies.

---

## Phase 8: Auto-Commit After Agent Sessions (Deferred)

*Plan only — not building yet. Needs its own proposal when the time comes.*

**Concept:** When an agent session completes (pipeline stage transition), the brain evaluates what files were created/modified, generates a meaningful commit message, and commits the work.

**Key design questions (to resolve in a future proposal):**
1. **Trigger:** Automatic on stage transition? Button in UI? Both?
2. **Scope:** Which files? All changes since last commit? Only files the agent touched? Need a way to track agent-modified files.
3. **Message:** AI-generated summary of the work? Template with entry title + stage? Human-editable before commit?
4. **Push:** Opt-in per commit? Global setting? Never auto-push?
5. **Safety:** Commits are permanent (in reflog). Push is more consequential. Need confirmation or at minimum clear audit trail.
6. **Architecture:** Pipeline post-stage hook (Go) vs. UI button + backend endpoint. Pipeline hook is cleaner for automation, button gives more control. Could start with button, add automation later.

**Prerequisite:** Phase 7 (git status) gives us the UI foundation. Auto-commit builds on knowing what changed.

**Revisit when:** After Phase 7 is shipped and in daily use. The git status display will surface whether auto-commit is actually needed or whether manual commits are fine.

---

## Phase 5: Smarter Auto-Advance Messages ✅ COMPLETE (Apr 5)

*Shipped: extractQuestionSummary() in research.go parses scratch file, counts numbered questions under "Open Questions" heading, lists category sub-headings. Appended to auto-advance message.*

### Current Message

> Auto-advanced: raw → researched. Research pass complete. Findings at .spec\scratch\build-physical-display-dashboard\main.md

### Improved Message

> Auto-advanced: raw → researched. Research pass complete.
>
> **What I found:** Hardware options (LVGL on ESP32-S3), LCARS design patterns, integration approaches for ibeco.me and weather.
>
> **What I need from you:** 15 open questions about hardware drivers, design fidelity, feature scope, and connectivity. [View full research →](.spec/scratch/build-physical-display-dashboard/main.md)
>
> Your answers will drive the planning phase.

**Implementation:** After research agent completes, read the scratch file, extract the "Open Questions" section header count and category summary. Include in the auto-advance message.

**Effort:** ~30 lines in pipeline/research.go

### Phase 5 Verification ✅

- [x] Auto-advance message includes question count and categories
- [x] File link in message is clickable (Phase 1 makes this work)
- [x] Message doesn't bloat — summary is 3-4 lines max

---

## Costs & Risks

| Phase | Backend | Frontend | New Deps | Risk |
|-------|---------|----------|----------|------|
| 1: See What Agent Did | ~60 lines | ~130 lines | markdown-it | Low — one new endpoint + rendering |
| 2: File Browser | ~50 lines | ~150 lines | None new | Low — proven pattern from ibeco.me |
| 3: WebSockets | ~120 lines | ~80 lines | None (gorilla already in go.mod) | Medium — connection lifecycle management |
| 4: Cost Tracking | ~30 lines | ~20 lines | None | Low — DB schema + display |
| 5: Smarter Messages | ~30 lines | 0 | None | Low — backend-only change |
| 6: Reader UX | 0 | ~80 lines | None | Low — frontend-only, builds on Phase 1+2 |
| 7: Git Status | ~50 lines | ~30 lines | None (os/exec) | Low — read-only git commands |
| 8: Auto-Commit | TBD | TBD | TBD | Medium — destructive actions need safety |

**Total Phases 1-7:** ~340 backend, ~490 frontend. Delivered incrementally — each phase stands alone.
**Phase 8:** Deferred. Needs its own proposal.

**Biggest risk:** Phase 3 (WebSockets) changes the communication model. If buggy, could cause connection leaks or missed events. Mitigate with timeouts, heartbeats, and graceful degradation (fall back to polling).

---

## Creation Cycle Review

| Step | Question | Answer |
|------|----------|--------|
| Intent | Why? | The pipeline is opaque. User can't see agent output without switching tools. Reader exists but doesn't connect. |
| Covenant | Rules? | OWASP for file serving. Match ibeco.me patterns. One session per phase. |
| Stewardship | Who owns what? | dev agent executes each phase against this spec. |
| Spiritual Creation | Spec precise enough? | Phases 1-3 ✅ shipped. Phase 4-7 specced. Phase 8 deferred to own proposal. |
| Line upon Line | Phasing? | 8 phases, each independent. Phases 6-7 build on 1+2 but ship standalone value. |
| Physical Creation | Who executes? | dev agent with brain.exe running locally. |
| Review | How do we know it's right? | Verification checklists per phase. |
| Atonement | What if it goes wrong? | File serving has security checks. WebSocket has fallback. Git endpoints are read-only. Auto-commit (Phase 8) deferred until safety model is designed. |
| Sabbath | When do we stop? | After Phase 7. Phase 8 gets its own proposal and review cycle. |
| Consecration | Who benefits? | Michael directly. Model for brain-as-tool UX. |
| Zion | Integration? | Phase 3 benefits ibeco.me relay. Phase 6 makes reader a connected system. |

## Recommendation

**Phase 6 next after Phase 4.** Phase 6 is the cohesive reader experience — link fix, internal navigation, deep linking, history. It makes Phases 1+2 fully useful. Phase 7 (git status) is nice-to-have but not blocking daily use. Phase 8 is explicitly deferred.
