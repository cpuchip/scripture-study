# Brain UX Quality-of-Life Improvements

**Status:** in-progress (Phase 1 ✅, Phase 5 ✅, Phase 2 ✅, Phase 3 ✅, Phase 4 next)
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

## Phase 4: Cost Tracking

*Know what the pipeline is spending.*

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

- [ ] Research pass increments premium_requests_used by 0.33
- [ ] Plan pass increments by 1.0
- [ ] Entry detail shows the running total
- [ ] Project detail shows aggregate across entries

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

**Total across all phases:** ~290 backend, ~380 frontend. But delivered incrementally — each phase stands alone.

**Biggest risk:** Phase 3 (WebSockets) changes the communication model. If buggy, could cause connection leaks or missed events. Mitigate with timeouts, heartbeats, and graceful degradation (fall back to polling).

---

## Creation Cycle Review

| Step | Question | Answer |
|------|----------|--------|
| Intent | Why? | The pipeline is opaque. User can't see agent output without switching tools. |
| Covenant | Rules? | OWASP for file serving. Match ibeco.me patterns. One session per phase. |
| Stewardship | Who owns what? | dev agent executes each phase against this spec. |
| Spiritual Creation | Spec precise enough? | Phase 1 yes. Phases 2-5 need detail when they're next up. |
| Line upon Line | Phasing? | 5 phases, each independent. Phase 1 is highest priority. |
| Physical Creation | Who executes? | dev agent with brain.exe running locally. |
| Review | How do we know it's right? | Verification checklists per phase. |
| Atonement | What if it goes wrong? | File serving has security checks. WebSocket has fallback to polling. |
| Sabbath | When do we stop? | After Phase 1 — use it, then decide what's next. |
| Consecration | Who benefits? | Michael directly. Also: model for brain-as-tool UX. |
| Zion | Integration? | Phase 3 benefits ibeco.me relay too. Phase 2 is portable. |

## Recommendation

**Build Phase 1 first.** It solves 80% of the pain (can't see output, can't type replies, can't access files). One focused session. Use it on the Space Center entries, then decide if Phase 2-5 are needed or if the inline viewer is sufficient.

Phase 5 (smarter messages) could also be done quickly and independently since it's backend-only.
