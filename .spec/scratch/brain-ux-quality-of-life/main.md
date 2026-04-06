# Brain UX Quality-of-Life — Research & Findings

**Binding Problem:** The brain app's pipeline works mechanically but is opaque to the user. After an agent auto-advances an entry, the user can't see what was generated, can't tell what input is needed, and has to switch to VS Code to read the output. The reply textbox is too small for substantive responses. There's no real-time feedback. These aren't feature requests — they're friction points discovered by actually using the review flow on the Space Center display entry.

**Discovered:** 2026-04-05, during first real review cycle on entry "Build Physical Display Dashboard"

---

## The Six Pain Points

### 1. Reply Textarea Too Small

**Current state:** `rows="2"`, `resize-none` in both EntryDetailView.vue (line 428) and the inline panel spec.

**Problem:** Works fine for "yes, do it" responses. Falls apart for multi-paragraph answers like the ESP32/LCARS description Michael gave. The text scrolls inside a tiny box — hard to read, hard to edit.

**Fix:** Auto-expanding textarea that grows with content. Common pattern: measure scrollHeight, set height to match. Cap at some max (e.g., 300px) then scroll. Remove `resize-none`.

**Effort:** ~15 lines of code. One composable or inline watcher.

---

### 2. No Access to Generated Files

**Current state:** Agent posts "Auto-advanced: raw → researched. Research pass complete. Findings at .spec\scratch\build-physical-display-dashboard\main.md" as a plain-text session message. No markdown rendering. No link detection. No file serving.

**The disconnect:** The agent writes a detailed 100+ line research document with structured sections, open questions, and source citations. The user sees a one-line message and has no way to access it from the brain UI. Has to open VS Code.

**Root cause:** 
- Messages rendered with `{{ msg.content }}` + `whitespace-pre-wrap` — no markdown, no link parsing
- Brain serves no workspace files — only the embedded SPA (`http.FS(s.frontendFS)`)
- No file reader component exists (ibeco.me has one, brain doesn't)

---

### 3. No Visibility Into What Agent Needs

**Current state:** The research agent writes open questions into the scratch file (the display entry had 15 excellent questions). But the auto-advance message just says "Research pass complete." The user has to dig into the file to discover what the agent needs from them.

**Possible approaches:**
- A. Include a summary of open questions in the auto-advance message itself
- B. Make the file clickable so the user can read it immediately
- C. Both — brief summary in message + link to full document

Option C is best. The summary gives immediate context ("I have 15 open questions about hardware, design, and integration"), the link gives full access.

---

### 4. Cost/Token Tracking Per Entry

**Current state:** `tokens_used INT64` exists on the entries table but is only set once when routing completes (db.go line 308). No per-message tracking. No premium request counting.

**What's missing:**
- Token counts per agent interaction (research pass, plan pass, nudge)
- Premium request cost per entry (how many requests did this entry consume?)
- Running total visible in the UI

**Challenge:** The Copilot SDK may not expose token counts back to the caller. Need to check what `ai.NewAgent` returns after completion. If not available, we can at least count the number of agent invocations (each = 0.33 or 1.0 premium requests depending on model).

---

### 5. File Browser / Library Reader

**Current state:** Brain has no file browser. ibeco.me has a full one:
- `becoming/frontend/src/views/ReaderView.vue` — tree sidebar + markdown content area
- `becoming/frontend/src/components/TreeNode.vue` — expandable directory tree
- Uses markdown-it for rendering
- Has bookmarks, heading anchors, search filtering
- Data source: GitHub API (not local filesystem)

**For brain:** We need a workspace file browser. Two levels:

**Level 1 — Inline file viewer:** Detect file paths in agent messages, make them clickable, serve the file content through a new API endpoint, render in a modal or slide-out panel with markdown rendering. This solves the immediate problem.

**Level 2 — Full file browser sidebar:** Port the ibeco.me ReaderView pattern. File tree on the left, rendered content on the right. New route `/library` or integrated into the entry detail view.

**Backend needed:**
- `GET /api/files/tree?root=.spec` — returns directory listing
- `GET /api/files/read?path=.spec/scratch/build-physical-display-dashboard/main.md` — returns file content
- **CRITICAL: path traversal protection.** Must jail to workspace root. Reject `..`, absolute paths, symlinks outside root.

---

### 6. Real-Time Updates (WebSockets)

**Current state:** 
- gorilla/websocket v1.5.3 already in go.mod
- Used ONLY for outbound relay to ibeco.me (relay/client.go)
- Frontend has NO WebSocket connection to brain.exe
- Dashboard/ReviewView poll with setInterval; EntryDetailView doesn't poll at all
- After sending a reply that triggers auto-advance, user has to manually refresh to see the agent's response

**What's needed:** A local WebSocket server endpoint so the frontend gets push notifications:
- New messages on an entry
- Entry status changes (maturity, route_status)
- Agent progress (started, in-progress, complete)
- Nudge bot activity

**Architecture:** 
- `GET /ws` endpoint — upgrades to WebSocket
- Hub pattern: one goroutine manages connected clients, broadcasts events
- Frontend: composable `useWebSocket()` — auto-reconnect, subscribe to entry/project events
- Events: `entry.updated`, `message.new`, `agent.started`, `agent.completed`

---

## ibeco.me Reader — What to Port

The ReaderView.vue from ibeco.me has these components worth studying:

| Component | Port? | Notes |
|-----------|-------|-------|
| TreeNode.vue | Yes | Directory tree with expand/collapse. Reusable. |
| File tree loading | Adapt | ibeco.me uses GitHub API. Brain needs local filesystem API. |
| Markdown rendering | Yes | markdown-it with custom heading anchors. |
| Bookmarks | Later | Nice-to-have, not urgent |
| Search filter | Yes | Filter tree by name. Very useful. |
| Share modal | No | Brain is local, not public |

---

## Critical Analysis

**Is this the right thing to build?** Yes. These are discovered-from-use problems, not speculative features. Michael literally couldn't complete a review cycle without switching to VS Code. That's a broken workflow.

**Mosiah 4:27 check:** These are small, focused improvements. Each one can be built independently in a single session. The risk is trying to build all six at once instead of one at a time.

**What gets worse?** 
- File serving adds security surface area (path traversal). Must be careful.
- WebSockets add connection management complexity. Must handle reconnects, cleanup.
- Markdown rendering adds a dependency (markdown-it or marked). Small but real.
- More code = more maintenance. But these are UX essentials, not nice-to-haves.

**Does this duplicate something we already have?** The file browser component exists in ibeco.me. Port it instead of rebuilding. The markdown rendering library choice should match ibeco.me's (markdown-it).

**Simplest version that's useful:**
1. Auto-expanding textarea (5 minutes)
2. Markdown rendering in messages (1 hour — add lib, render messages)
3. Parse file paths in messages → make clickable → serve file in modal (1 session)

That trio solves 80% of the pain. File browser, WebSockets, and cost tracking are Phase 2+.

---

## Recommended Phasing

### Phase 1: See What the Agent Did (one session)
- Auto-expanding textarea
- Add markdown-it to frontend dependencies
- Render message content as markdown (with sanitization)
- File path detection in messages → clickable links
- `GET /api/files/read` endpoint with path traversal protection
- Click link → modal/panel shows rendered file content

**Why this first:** This is the critical path. Without it, the review flow is broken — user can't see agent output without switching tools.

### Phase 2: File Browser (one session)
- `GET /api/files/tree` endpoint
- Port TreeNode.vue from ibeco.me
- New route or panel: file browser with tree sidebar + content area
- Search/filter on tree

### Phase 3: Real-Time Updates (one session)
- WebSocket server endpoint (`/ws`)
- Hub pattern for broadcasting events
- Frontend `useWebSocket()` composable
- Auto-update messages when agent posts new ones
- Live status badges on entries

### Phase 4: Cost Tracking (one session)
- Instrument agent calls to count invocations per entry
- Track premium request cost (0.33 for Haiku, 1.0 for Sonnet, etc.)
- Display cost badge on entry detail view
- Maybe: total cost per project rollup

### Phase 5: Smarter Auto-Advance Messages (small)
- Research agent includes "Questions for you:" summary in the advance message
- Or: auto-advance handler reads the scratch file and extracts the "Open Questions" section
- Include in the session message so the user sees questions without opening the file

---

## Key Decision: markdown-it vs marked

ibeco.me uses markdown-it. Brain frontend has zero markdown deps.

| Library | Size | Features | ibeco.me uses? |
|---------|------|----------|---------------|
| markdown-it | 39kB min | Plugin ecosystem, security rules, lazy loading | Yes |
| marked | 38kB min | Fast, simple, less extensible | No |

**Recommendation:** markdown-it. Matches ibeco.me, has plugins for heading anchors, code highlighting. One library across both apps.

---

## Security Notes

**File serving endpoint MUST:**
1. Resolve the requested path relative to workspace root
2. `filepath.Clean()` the path
3. Verify the resolved path is UNDER the workspace root (prevent `../../etc/passwd`)
4. Reject absolute paths
5. Reject symlinks that resolve outside workspace
6. Consider a whitelist of allowed extensions (`.md`, `.yaml`, `.json`, `.txt`, `.go`, `.vue`, `.ts`)
7. Set Content-Type properly
8. Consider rate limiting (though this is local, not public)

**WebSocket endpoint:**
- Local-only (same as all brain endpoints)
- No auth needed (brain runs on localhost)
- Handle client disconnect gracefully
- Limit broadcast message size

---

## Raw Evidence

**Textarea:** EntryDetailView.vue line 428 — `rows="2"` + `resize-none`
**Message rendering:** EntryDetailView.vue line 443 — `{{ msg.content }}` with `whitespace-pre-wrap`
**File serving:** web/server.go line 970 — only `http.FS(s.frontendFS)`
**WebSocket dep:** go.mod — `gorilla/websocket v1.5.3`
**Token tracking:** db.go line 182/308 — `tokens_used INT64`, set once at routing completion
**ibeco.me reader:** becoming/frontend/src/views/ReaderView.vue
**Auto-advance message:** web/server.go line 1719 — `fmt.Sprintf("Auto-advanced: %s → %s. %s", ...)`
**Research scratch path:** pipeline/research.go line 226 — `filepath.Join(".spec", "scratch", slug, "main.md")`
