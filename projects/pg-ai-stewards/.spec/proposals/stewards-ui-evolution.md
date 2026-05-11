---
title: Stewards-UI evolution — authoring, chat, navigation, write actions
date: 2026-05-11
status: design proposal — needs ratification per section before build
parent: open-items.md (Sections IV.2, IV.7-IV.10, X.11; Michael's 2026-05-11 new asks)
purpose: >
  Stewards-UI shipped read-mostly in 3f v1 (2026-05-09), then gained per-phase
  surfaces (Intents/Covenants in C.8, Sabbath/Lessons in D.7, Trust in E.7,
  Councils in F.7). 14 routes now. This proposal scopes the next layer:
  authoring (no more "edit YAML and commit"), conversational interaction
  (substrate-aware chat), navigation cleanup (sidebar grouping), and the
  remaining read-mostly→write-mostly conversions.
---

# Stewards-UI evolution

## I. Binding problem

Today's UI is read-mostly with a few action islands: NewWork form, Sabbath/Lessons ratify buttons, Trust adjust modal, Councils convene/resolve. Several friction points:

1. **Intent + covenant authoring requires editing YAML files + committing.** That's the right pattern for repo-tracked canonical state, but it makes one-off / experimental intents heavy. Michael flagged wanting to add intents from the UI.
2. **No conversational interface to the substrate.** Today Michael talks to the substrate via Claude Code (here) or via the UI's discrete action buttons. A substrate-aware chat that can read/write the substrate would be a third surface — and one with read-only/write-mode safety controls. Michael's exact framing: *"with k2.5 or qwen or glm etc... with read only (ask questions) or write mode (effect change)"*.
3. **14 routes in horizontal nav is genuinely cluttered.** Visual scan time has crept up.
4. **Several views are read-only where actions are obviously wanted.** Watchman finding ack, Bridge refresh-tools, WorkItemDetail advance/cancel/dispatch. Each is a small endpoint + button.

## II. Success criteria

Sectioned by sub-feature; each is independently mergeable.

### II.A — Intent + Covenant authoring (UI-side)
1. A new intent can be created from the UI without editing YAML. Created intents are substrate-native (no `source_file`).
2. The UI distinguishes substrate-native intents from YAML-canonical ones so Michael never confuses "this is repo-tracked" with "I clicked-to-make this last Tuesday."
3. Existing YAML-seeded intents are read-only in the UI; an "Edit YAML" button links to the file path with a tooltip ("editing here means committing intent.yaml").

### II.B — Substrate-aware chat
1. A new `/chat` route + dashboard widget host a multi-turn conversation with a substrate-aware agent.
2. Two modes selected per-chat: **read-only** (the agent can query substrate state, can NOT modify) and **write** (the agent can also create intents, dispatch work_items, convene councils, etc.). Mode is visible in the UI and immutable mid-conversation.
3. The agent has model choice — kimi-k2.6, qwen3.6-plus, glm-5.1 default options.
4. Tool surface is minimal + scoped:
   - read-only: `substrate_query(sql, limit=100)`, `substrate_view(view_name, where)`, the existing `study_*` MCP tools
   - write (in addition): `intent_create`, `dispatch_work_item`, `convene_council`, `enqueue_pending_file_write`, plus a generic `substrate_mutate(sql)` gated by a confirmation step (the chat shows the SQL and asks "execute?")

### II.C — Navigation cleanup
1. Horizontal nav replaced by left sidebar with 4 groups: **Substrate** (Intents, Covenant, Sabbath, Lessons, Trust, Councils, Scheduled), **Workspace** (Work items, Sessions, Watchman, Bridge, New work), **Records** (Studies, Graph), **Action** (Chat, Dashboard).
2. Active route highlighted.
3. Mobile-friendly fallback: sidebar collapses to hamburger menu.

### II.D — Write actions on read-mostly views
1. Watchman page: per-finding "Ack" button → POST /api/watchman/findings/ack
2. Bridge page: top-right "Refresh tools" button → POST /api/bridge/refresh-tools (triggers `bridge refresh-tools` via NOTIFY or exec-into-container)
3. WorkItemDetail: "Advance" / "Cancel" / "Re-dispatch current stage" buttons in a new actions panel
4. All write actions show a quick toast confirming success/failure

## III. Constraints and boundaries

**In scope (per section):**
- II.A: new modal in `/intents` for create-intent (already partially built — Phase C.8 added inline-create from NewWork); UI distinguishes source_file IS NULL vs NOT NULL
- II.B: new pipeline_family `steward-chat` + new agent + chat-UI Vue components + chat dispatch path (multi-turn, NOT a maturity ladder)
- II.C: structural Vue change to `App.vue` + new sidebar component + router-link active class
- II.D: 3 small backend endpoints + 3 UI button additions

**Out of scope (explicitly):**
- Covenant authoring in UI. Covenants are bilateral commitments to the project as a whole; "click to add a covenant commitment" is too low-friction. Keep YAML-only.
- Plugin / extension architecture for chat tools. v1 has a fixed tool surface.
- Chat history search / threading across sessions. v1 is one session = one continuous thread.
- Mobile-first rebuild. Sidebar collapses for mobile but the UI stays primarily desktop.
- File upload to chat (e.g. attach a transcript). v1 is text-only conversation.

## IV. Prior art

- **Phase C.7/C.8** Intent inline-create modal in NewWork.vue + `/api/intents/create` endpoint. Already works for new intents; II.A mostly surfaces it on `/intents`.
- **Phase F's `/councils/:id`** — proves that multi-turn substrate UI pattern (auto-refresh while in flight, role-tinted message cards) works. Chat can borrow the same Vue components.
- **stewards-mcp Go binary** — already exposes substrate read tools via MCP (`work_item_list`, `watchman_passes_list`, etc.) for Claude Code. Chat can use SUB-set in read-only mode and a wider set in write mode.
- **bgworker tool_dispatch path (Phase 3e.2)** — chat agent's tool calls route through the existing mcp_proxy + bridge mechanism. No new infrastructure for tool invocation.
- **`substrate_query(sql)` doesn't exist yet** as an MCP tool. New addition; bounded by READ_ONLY transaction wrapper + row limit + 5s timeout.

## V. Proposed approach

### V.A Intent + Covenant authoring

**Intent.vue gains:**
- "+ New intent" button (top-right). Same modal Phase C.8 added to NewWork, lifted up.
- Per-row badge: 📄 for YAML-sourced (clickable opens VS Code via `vscode://` link if Michael wants), ✨ for substrate-native
- "Promote to YAML" action on substrate-native intents — generates the YAML snippet, copies to clipboard, suggests appending to intent.yaml + commit. Doesn't write the file directly (manual git is still the source of truth).

**Covenants.vue — read-only kept.** The covenant is project-scope; UI-only edits are wrong shape. The page surfaces the "Edit .spec/covenant.yaml" instruction.

### V.B Substrate-aware chat

#### V.B.1 Architecture

Chat is a NEW pipeline pattern. Existing patterns:
- **Pipeline** = one agent walking a maturity ladder, one-shot per work_item
- **Council** = multiple agents deliberating on one binding question

**Chat** = one agent in an open-ended multi-turn conversation. Not goal-directed; not gated; doesn't accumulate trust the same way.

**Schema additions:**
```sql
ALTER TABLE stewards.sessions
    ADD CONSTRAINT sessions_kind_check
    CHECK (kind = ANY (ARRAY['chat','agent','tool','study','dev','gate','sabbath','atonement','council','steward_chat']));

CREATE TABLE stewards.chat_threads (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id      text NOT NULL UNIQUE,         -- same session_id used by stewards.sessions
    mode            text NOT NULL CHECK (mode IN ('read_only', 'write')),
    model           text NOT NULL,                -- kimi-k2.6 | qwen3.6-plus | glm-5.1 etc.
    intent_id       uuid REFERENCES stewards.intents(id),
    title           text,                         -- auto-generated from first user message or human-edited
    created_at      timestamptz NOT NULL DEFAULT now(),
    last_message_at timestamptz,
    archived_at     timestamptz
);
```

#### V.B.2 Agent + tools

New agent: `steward-chat` (lives in `.stewards/steward-chat.agent.md` or via the existing extension_sql! seed path).

System prompt frame:
- Active covenant + (optionally) intent injected by `compose_system_prompt` (already wired for any work_item-tied session; need to extend for chat_threads with intent_id)
- "You are a substrate-aware assistant. The user is talking to you in `{mode}` mode. In read-only mode, you can query but not change anything. In write mode, you may also create + modify substrate state — but always confirm before destructive operations."
- Tool list inline in system prompt

**Tools (read-only mode):**
- `substrate_query(sql text, limit int = 100)` — wrapped in `SET LOCAL TRANSACTION READ ONLY` + 5s timeout + 100-row default cap. Returns rows as JSON.
- `substrate_view(view_name text, where text = NULL, limit int = 100)` — convenience for known views (work_items_steward_status, lessons_recent_ratified, etc.)
- All existing `study_*` MCP tools from Phase 3c.2.5
- `current_substrate_summary()` — returns a structured summary: in-flight work_items, active council if any, recent sabbath reflections, dirty queue depth. Useful for "what's happening in the substrate right now?"

**Tools (write mode, in addition):**
- `intent_create(slug, purpose, ...)` — same as `/api/intents/create`
- `work_item_create_and_dispatch(pipeline_family, input)` — convenience wrapper
- `convene_council(intent_id, binding_question, members, bishop)` — wraps SQL fn
- `pending_file_write_enqueue(path, content, mode)` — for asking the substrate to write a file (Batch G's mechanism)
- `substrate_mutate(sql text)` — generic escape hatch. Requires the agent to call `confirm_mutation(description)` first; the chat UI shows the SQL + description and asks the human to approve. Without approval the actual SQL doesn't execute.

#### V.B.3 Chat dispatch path

Multi-turn chats are NOT work_items. They're sessions only.

**On user message:**
1. UI POSTs to `/api/chat/send` with `{thread_id, content}`
2. Backend INSERTs into `stewards.messages` (role='user', model=current chat model)
3. Backend enqueues a `chat` work_queue row with `_steward_chat=true` marker (new marker; joins the 7 existing in bgworker auto-fire)
4. Bgworker dispatches the chat with the steward-chat agent + chat-thread mode-appropriate tools
5. Assistant response inserted into stewards.messages by the existing chat path
6. UI polls `/api/chat/get?thread_id=` every 2s (or via SSE if Batch J ships it) and renders new messages

**On confirmation needed (write mode `substrate_mutate`):**
1. Agent calls `confirm_mutation(description, sql)` tool
2. Tool inserts a row into a new `stewards.chat_confirmations` table; returns a synthetic tool_call_id
3. UI sees the pending confirmation, shows SQL + description, "Approve" / "Reject" buttons
4. Human clicks Approve → UI POSTs `/api/chat/confirm/<id>` → row marks approved → next user message + agent dispatch can execute the SQL (assistant gets the row's state via `check_confirmation(id)` tool)

**Why this confirmation pattern, not just direct execution:** mistakes in write mode have real blast radius. The substrate's destructive operations need a human-in-the-loop checkpoint that matches Claude Code's tool-call permission flow. The chat agent must ASK before it acts; the human always sees the SQL.

#### V.B.4 UI

**`/chat` route:** list of threads + active thread panel.
- Sidebar: thread list (title, mode badge, model badge, last activity), "+ New chat" button
- Main: message history (user/assistant cards, tool calls rendered as collapsible blocks with results), pending confirmations highlighted, input textarea at bottom + send button + model selector + mode toggle (disabled if any messages exist — mode is immutable mid-thread)
- Auto-refresh every 2s while waiting for assistant response

**Dashboard widget:** "Most recent chat" with last 3 messages + click-through.

#### V.B.5 Trust + cost integration

Chat work_queue rows record cost_events normally. Per-thread cost surfaced in the thread header. Trust state for the steward-chat agent doesn't pass through the normal pipeline → verified path (chats don't reach verified maturity); decide:
- **Option A:** chat agent doesn't participate in trust at all. Its dispatches are categorized; no trust transitions.
- **Option B:** chat earns trust via thumbs-up / thumbs-down per message. Lighter signal than verified completions.

**Recommendation: A** for v1. Trust was designed for goal-directed work; chat is exploratory.

### V.C Navigation cleanup

Replace App.vue's horizontal nav with a left sidebar component. Tailwind classes for the active state. Group structure:

```
SUBSTRATE
  Intents
  Covenant
  Sabbath
  Lessons
  Trust
  Councils
  Scheduled                  [new in pipelines-expansion]

WORKSPACE
  Dashboard
  Work items
  Sessions
  Watchman
  Bridge
  New work

RECORDS
  Studies
  Graph

ACTION
  Chat                       [new in V.B]
```

Mobile: sidebar collapses to a hamburger menu icon top-left; current page name in header.

### V.D Write actions on read-mostly views

#### Watchman finding ack
- Backend: POST `/api/watchman/findings/ack` body `{finding_id, acked_by}` → UPDATE finding SET acked_at = now(), acked_by = ...
- UI: per-finding "Ack" button. Disabled if already acked.

#### Bridge refresh-tools
- Backend: POST `/api/bridge/refresh-tools` → NOTIFY 'bridge_refresh' OR `docker exec pg-ai-stewards-bridge stewards-mcp bridge refresh-tools`
- UI: top-right "Refresh tools" button with last-refreshed timestamp

#### WorkItemDetail actions panel
- Backend: POST `/api/work-items/advance` `/cancel` `/redispatch`
- UI: new actions card with three buttons, each with confirmation modal

## VI. Decision points for Michael

### VI.A Intent authoring
- **D-UI1:** Promote-to-YAML helper that generates clipboard text + suggests git diff? Or full automation (UI writes intent.yaml + auto-commits)? Recommend: clipboard helper only. Git commits are human-stewarded.
- **D-UI2:** Should substrate-native intents ever be promoted to YAML automatically? Or is the manual-promote path the only one? Recommend: manual only.

### VI.B Chat
- **D-UI3: Read-only/write mode-switch policy.** Mode is set at thread creation and immutable mid-thread (recommended), OR can be toggled per-message with a confirmation gate, OR write-mode requires re-auth each time? Recommend: mode immutable per thread. New thread for a different mode.
- **D-UI4: Default model for chat.** kimi-k2.6 (default everywhere) / qwen3.6-plus (cheaper) / glm-5.1 (heaviest)? Recommend: kimi-k2.6 default; user can switch per-thread.
- **D-UI5: `substrate_mutate` confirmation flow.** Always confirm in write mode (recommended), or only for specific patterns (DELETE, DROP, TRUNCATE)? Recommend: always for the generic mutator; the targeted tools (`intent_create`, etc.) execute without confirmation since their shape is bounded.
- **D-UI6: Cost cap per chat thread.** Should each thread have a token budget? Recommend: yes, default $0.50 per thread (configurable per-mode). The chat agent can warn when 80% of budget is reached.
- **D-UI7: Should chat have intent injection?** Per V.B.1, optional `intent_id` on chat_threads. Recommend: optional. If set, intent purpose appears in system prompt; if NULL, covenant only (like ad-hoc chats).
- **D-UI8: Trust for chat agent (Option A vs B).** Recommend A (chat doesn't participate in trust).
- **D-UI9: What if the agent makes a mistake in write mode?** No rollback for already-executed mutations. Recommend: chat agent's system prompt explicitly encourages dry-run-then-execute pattern; `substrate_mutate` confirmation gate is the safety net.

### VI.C Navigation
- **D-UI10: Sidebar group names** as proposed (Substrate / Workspace / Records / Action), or different? Recommend: as proposed.
- **D-UI11: Default-collapsed groups?** All four expanded on first load (recommended), or only the group containing the current route?

### VI.D Write actions
- **D-UI12: Watchman finding ack — what UPDATE shape?** Add `acked_at + acked_by` to `watchman_findings` (simple) vs. a separate `finding_acks` table (audit-trail-friendly). Recommend: simple columns; the substrate already has good audit elsewhere.

## VII. Estimated programming time

| Section | Sessions |
|---|---|
| V.A Intent authoring (mostly already built; surface promotion) | 0.5 |
| V.B Substrate-aware chat (backend pipeline + bgworker marker + tools + UI components + confirmation flow) | 3–4 (largest) |
| V.C Sidebar grouping | 0.5 |
| V.D Write actions (Watchman ack + Bridge refresh + WorkItem actions) | 1 |

**Total: 5–6 sessions.** Realistically two batches:
- **UI Polish batch:** V.A + V.C + V.D (~2 sessions)
- **Chat batch:** V.B alone (~3-4 sessions; biggest piece, ratifications first)

## VIII. Acceptance scenarios

### Authoring
1. From `/intents`, click "+ New intent". Modal: slug, purpose, scripture_anchor. Submit. New intent appears in list with ✨ badge. NewWork form picks it up.
2. Click "Promote to YAML" on a substrate-native intent. Modal shows YAML snippet to paste into intent.yaml + a one-line instruction.

### Chat (read-only)
3. From `/chat`, click "+ New chat". Pick read-only + kimi-k2.6. Ask: "How many work_items have sabbath_completed_at set?" Agent uses `substrate_query` tool, returns count + slug list. Cost recorded; ~$0.005.

### Chat (write)
4. Same as above but write mode. Ask: "Create a new intent for tracking my daily AI news pipeline." Agent calls `intent_create(slug='professional-awareness', ...)`. New intent appears in `/intents`. Cost ~$0.01.
5. Ask: "Cancel the in-flight work_item with slug 'foo-bar'." Agent calls `substrate_mutate("UPDATE stewards.work_items SET status='canceled' WHERE slug='foo-bar'")` → confirmation modal shows the SQL → human clicks Approve → executes → assistant confirms.

### Navigation + write actions
6. Sidebar collapses on mobile to hamburger. Click hamburger → sidebar slides in.
7. From Watchman page, click "Ack" on a finding. Toast confirms. Reload — finding shows "acked by michael, <timestamp>" badge.
8. From Bridge page, click "Refresh tools". Toast confirms. Last-refreshed timestamp updates.
9. From WorkItemDetail of an in-flight work_item, click "Advance" → confirmation → status transitions.

## IX. Why this is one proposal, not four

All four sub-features (authoring / chat / navigation / write actions) compound on each other:
- Chat in write mode wants the same `intent_create` flow the authoring UI uses
- Sidebar grouping anticipates `/chat` and `/scheduled` (from pipelines-expansion) routes
- Write actions normalize the "POST endpoint → toast" pattern that chat's confirmation flow also uses

One proposal, two batches (Polish + Chat). Chat is the big lift; everything else is small.

## X. Open architectural concern

**Chat agent is a brand new pattern.** Pipeline = ladder; Council = parallel deliberation; Chat = open-ended dialogue. The substrate has been intentionally structured for the first two. Adding the third risks pulling the substrate toward a general agent surface (where the actual work happens) rather than the orchestration surface (where work is governed). Worth ratifying explicitly that this is the right move.

Counter-argument for building it: today Michael talks to Claude Code (here) to interact with the substrate. That's a hard dependency on the Claude Code session for substrate observability. A substrate-internal chat is a steward of the substrate, not a competing general agent — its tool surface is bounded to substrate operations. The framing is "Bishop chat" or "Steward chat" — talking to the entity that watches the substrate, not to a separate AI assistant.

Either ratify-or-defer this section explicitly before building.
