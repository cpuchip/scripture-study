# Phase 6 & 7 Research — Reader UX + Git Integration

**Binding Problem:** The file browser (Phase 2) and file viewer (Phase 1 sidebar) both work in isolation, but they don't connect to each other. File paths in agent messages open a sidebar panel that can't follow internal links. The library reader has no history navigation and no deep linking. Git status is invisible. These gaps prevent the brain from being the single interface for project awareness.

**Session:** 2026-04-06, building on Phases 1-3 shipped earlier today.

---

## Current Architecture Inventory

### File Path Detection — `useMarkdown.ts`

```
FILE_PATH_RE = /(?:^|\s|["'(])((\.spec|study|scripts|...)\/[\w./_-]+...)/g
```

**Bug:** Lookbehind is `(?:^|\s|["'(])` — backtick (`) is NOT included. So `\`scripts/brain/foo.go\`` won't match because the path is preceded by a backtick. Markdown code formatting wraps paths in backticks. This is common in agent output.

**Fix:** Add backtick to character class: `(?:^|\s|["'(\x60])` — one-line change.

### File Viewer — EntryDetailView

`handleMessageClick()` on line 224 handles `.file-link` clicks → calls `openFileViewer(path)` → opens `FileViewer.vue` as a 45vw sidebar panel overlaying from the right.

**Limitation:** FileViewer renders markdown but doesn't handle file-link clicks *within* the rendered content. So if the opened file mentions another file path, that path isn't clickable. And there's no way to navigate to the Library reader from here.

### Library Reader — LibraryView.vue

`openFile(path)` fetches content via `api.readFile(path)`, renders with `renderMarkdown()`.

**Limitations:**
1. No click handler on rendered content — file paths in rendered markdown aren't clickable
2. No navigation history — no back/forward, no breadcrumbs
3. No route query params — can't deep-link to `/library?file=path/to/file.md`
4. No way to navigate here from EntryDetailView file links

### Router

Routes defined in `main.ts`. Library route is just `{ path: '/library', component: LibraryView }` — no query param support, no named route with props.

### Go Server — File Endpoints

- `GET /api/files/read?path=` — returns raw text, multi-layer path validation
- `GET /api/files/tree?root=` — returns JSON tree, skips node_modules/.git/dist/gospel-library/private-brain
- No `/api/git/*` endpoints exist

### TreeNode Component

Emits `toggle-dir` and `open-file`. No slot for status indicators. Would need modification for git status badges.

---

## Feature Analysis

### Item 1: Backtick Link Fix

**Effort:** One regex character class change. Zero risk.

**What:** Add backtick to `FILE_PATH_RE` lookbehind.

**Edge case:** markdown-it converts \`code\` to `<code>foo</code>`. After markdown rendering, the backtick itself is gone — it's now inside a `<code>` tag. So the regex (which runs on the rendered HTML) would need to also handle paths inside `<code>` tags. 

Actually, let me think about this more carefully. `linkifyFilePaths` runs on the HTML *output* of markdown-it. If the markdown source is `` `scripts/brain/foo.go` ``, markdown-it renders it as `<code>scripts/brain/foo.go</code>`. The regex runs on that HTML — the backtick is gone. But `<code>` starts with `<` which isn't in the lookbehind either.

So the fix is: add `>` to the lookbehind (to match after `<code>` closing bracket). The backtick fix handles raw text; the `>` fix handles markdown-rendered code blocks.

Combined lookbehind: `(?:^|\s|["'(\x60>])`

### Item 2: Reader Navigation (Cross-View Linking)

**Two sub-problems:**

A. **EntryDetailView → Library:** When a file link is clicked in entry detail, navigate to `/library?file=path` instead of opening the sidebar. (Or: offer both — sidebar for quick peek, shift+click or icon for full reader.)

B. **Internal links in reader:** When rendered content in the Library reader contains file paths, clicking them should navigate within the reader (update currentFilePath, fetch new content). Same for FileViewer sidebar.

**Recommended approach:**
- Add `file` query param to `/library` route. On mount, if `?file=path` present, auto-open that file.
- Add click handler to Library content area (like EntryDetailView's `handleMessageClick`) that intercepts `.file-link` clicks and calls `openFile(path)`.
- Add same click handler to FileViewer.vue for link following within the sidebar.
- For EntryDetailView: change `openFileViewer` to navigate to `/library?file=path`. This makes the Library the single reading experience. The sidebar FileViewer can be deprecated or kept for quick peeks.

**Decision point for Michael:** Should file links in entries navigate to Library (full reader), or keep the sidebar (quick peek)? Or both (click = sidebar, double-click/icon = library)?

### Item 3: Back Button in Reader

**Implementation:** Navigation history stack in LibraryView.

```typescript
const fileHistory = ref<string[]>([])
const historyIndex = ref(-1)

function openFile(path: string) {
  // Trim forward history when navigating to new file
  fileHistory.value = fileHistory.value.slice(0, historyIndex.value + 1)
  fileHistory.value.push(path)
  historyIndex.value++
  // ... fetch and render
}

function goBack() {
  if (historyIndex.value > 0) {
    historyIndex.value--
    loadFile(fileHistory.value[historyIndex.value])
  }
}

function goForward() {
  if (historyIndex.value < fileHistory.value.length - 1) {
    historyIndex.value++
    loadFile(fileHistory.value[historyIndex.value])
  }
}
```

Add back/forward buttons to the file header bar. Show current path as breadcrumbs.

**Effort:** ~40 lines frontend. No backend changes.

### Item 4: Git Status in File Browser

**Backend:** New endpoint `GET /api/git/status` that runs `git status --porcelain` and returns parsed results.

```go
func (s *Server) handleGitStatus(w http.ResponseWriter, r *http.Request) {
    cmd := exec.Command("git", "status", "--porcelain")
    cmd.Dir = workspaceRoot
    out, err := cmd.Output()
    // Parse porcelain output: "?? file", " M file", "A  file", etc.
    // Return JSON: [{ "path": "study/foo.md", "status": "new" }, ...]
}
```

**Frontend:** TreeNode needs a slot or prop for status indicator. Green dot for new/untracked, yellow for modified, red for deleted.

**Stretch — inline diff:** `GET /api/git/diff?path=study/foo.md` → runs `git diff study/foo.md`, returns unified diff. Display in a split or overlay view. This is a significant scope increase and should probably be a separate phase.

**Security consideration:** Running git commands server-side is safe here (localhost-only server, workspace-scoped). The path validation from handleFileRead should also apply to git endpoints.

**Effort:** ~50 lines backend (endpoint + parsing), ~30 lines frontend (TreeNode badge, status fetch). Diff viewer would be +200 lines.

### Item 5: Auto-Commit After Agent Sessions

Michael explicitly said "plan this work as a later phase." This is a significant feature:

- Detect when an agent session completes (pipeline stage transition)
- Evaluate which files were created/modified (workspace-level git diff)
- Generate a meaningful commit message (summarize the work done)
- Run `git add` for relevant files, `git commit -m "message"`
- Optionally `git push`

**Concerns:**
- Which files to include? Agent may touch scratch files, study docs, code — need filtering rules
- Commit granularity: one commit per pipeline stage? Per entry? Per session?
- Destructive action: commits are permanent. Need confirmation UX or at minimum clear logging
- Push is even more consequential — definitely needs opt-in

**Architecture:** This could live in the pipeline (Go side) as a post-stage hook, or as a scheduled action, or as a button in the UI ("Commit session artifacts"). The pipeline hook is cleanest for automation but the button gives more control.

---

## Phasing Recommendation

### Phase 6: Reader UX (Items 1-3)

All three items are cohesive — they're all about making the reading experience work as a connected system rather than isolated views. Low risk, moderate effort, immediate value.

Sub-phases:
- 6a: Backtick/code link fix (5 min)
- 6b: Internal links in reader content + click handlers
- 6c: Route query param `/library?file=path` + entry-detail navigation
- 6d: Back/forward navigation history

### Phase 7: Git Status Display (Item 4)

Standalone feature. Backend endpoint + TreeNode visual indicator. Defer diff viewer to later.

### Phase 8: Auto-Commit (Item 5)

Deferred. Needs its own proposal — significant scope, destructive actions, configuration decisions.
