---
date: 2026-06-23
topic: Stewdio — the work-item cockpit (idea → research → spec → council → P0 → P1 → P2, all live)
lane: pg-ai-stewards
---

# Stewdio: "what was missing from the beginning"

A second arc the same day as the llama-chip remote-management work. Michael, after the
federation arc, asked to refocus on pg-ai-stewards — and out of an AskUserQuestion menu he
redirected to a *better* idea than any of my options: a UI to **work with the built artifacts**,
with a chat window where, with a book study open, he could chat against "the full context of the
book and the agent sessions that went into it." When it landed live he said: "that is amazing and
it is what was missing from the beginning. holy cow!"

## The reframe (his)

The doc, its source corpus, and the agent sessions that produced it are all **facets of one
work_item**. So the unit isn't "chat with a document" — it's **"chat with a work item."** And he
widened it: a polished, VS-Code-like cockpit — project-filtered work-item browser on the left, a
multi-session model-switchable chat on the right that *talks to* a work item AND *kicks off* new
pipelines (book / video / a task) with live progress.

## Research, not guess (his ask: "do some research… there has to be a UI project we could borrow")

Web search + a 4-agent read-only code sweep of the substrate. Findings:
- **Borrow `dockview-vue`** (MIT, VS-Code docking, first-class Vue 3) for the shell — drops into the
  existing Vue 3 + Vite stewards-ui without a rewrite. The one real "borrow."
- **Don't adopt a chat app wholesale** (LibreChat/Open WebUI/LobeChat) — they're their own
  stacks/DBs and only know "chat with an LLM," not our work_items/pipelines/sessions. Mine the UX,
  build the panel.
- **The agent-cockpit field converged** (Cursor 3 / Devin Desktop / GitHub Mission Control / Claude
  Code) on **editor-vs-manager surface separation** and **"plan surface = progress stream."** That
  second one maps 1:1 onto our pipeline stages — a work_item's stages ARE the plan, lighting up as
  it runs. We got it nearly for free.
- **Key substrate finding:** YouTube digests persist their source (`yt_transcripts` keyed by
  `video_id`, linked from doc frontmatter); **book digests discard theirs** (fetched at build,
  lives only in that run's messages). So "chat with the book's passages" needs a persisted book
  corpus — deferred to P3; P1/P2 ship the doc + sessions + citations facets (universal).

## Spec → council → ratify

Wrote `.spec/proposals/stewards-studio.md` (grounded, with exact functions + file:line). Surfaced
the real tensions (the book-corpus gap honesty; chat-turn-as-work-item vs persistent-session;
streaming-can-only-be-a-DB-poll-relay; scope/pacing). Michael ratified 8 decisions: name
**Stewdio** (`/stewdio`), generic in OSS core · local model · **SSE from P1** · **Pinia** ·
book-corpus = P3 · order P0→P1→P2→P3→P4 · **persistent chat sessions** (not work-item-per-turn).
`dominion_in_council` satisfied — his nod was the ratification.

## Built, all live on his `:8081/stewdio`

- **P0** (`fc32c14`) — dockview-vue shell: lazy `/stewdio`, 3-zone dock (Work items | Artifact |
  Chat), abyss theme, full-bleed `App.vue` special-case, Pinia. dockview ships as its own ~76KB-gzip
  lazy chunk, off every other route.
- **P1a** (`630a41e`) — the backend: `extension/45-work-item-chat.sql` = the `work-item-chat` agent
  (read-only allow-list = deny `*` + allow the retrieval tools) + **`dispatch_chat_turn`** (a thin
  wrapper over the existing bare-session `chat_enqueue` — an Explore agent found it; marker-free
  `kind='chat'` → bgworker tool loop → reply in messages by session_id; no `_work_item_id` so the
  work_item/watchman triggers never fire). Chain 00→45, virgin-smoke OK 34.
- **P1b+P1c** (`3570055`) — Go `/api/chat/send` + `/api/chat/stream` SSE relay (DB-poll, so
  cost/trust accounting stays intact) + the three Vue panels wired (project-filtered browser · doc /
  work-item artifact view · streaming chat with a model switcher) + the Pinia store.
- **P2** (`3783333`) — `GET /api/pipelines/get` (the ordered stage plan) + a **"＋ New" launcher**
  (pick pipeline + binding question → `workItemCreate(dispatch:true)`) + **plan=progress** in the
  center (stages ✓/▸/○ from `stage_results`+`current_stage`, polled while running = "· live").

## Proofs (live container + local rig, all cleaned up after)

- P1: clicked the On Liberty study → it rendered → asked "Mill's central argument? quote the doc" →
  the agent did `doc_get` + `result_read` and **streamed a grounded answer with a verbatim quote**
  into the chat panel, local qwen, ~9s.
- P2: ＋New → book-digest → Launch → center rendered **▸ read · ○ build · ○ critique**, in_progress,
  "· live." Then `work_item_cancel`'d to free the rig.

## Bugs the e2e caught (verify-under-real-conditions earned its keep again)

- **NULL-slug 404 (P2):** `workItemsGetHandler` scanned `wi.slug` into a `*string` without coalesce
  → 404 "cannot scan NULL into *string" for any NULL-slug work item (which the substrate allows + a
  launched one has). The launcher hit it immediately. Fixed: `coalesce(wi.slug,'')` — a latent bug
  that would've hit any NULL-slug item.
- **A transient 500 on `/chat/send`** once, right after a container restart (rig warmup); not
  reproduced in 3 later sends. Added a server log on the dispatch-error path for next time.
- **`go:embed` stub:** `npm run build` overwrites the committed `dist/index.html` stub; must restore
  before each commit (bit me ~3×; now a reflex).
- **Playwright `fill`+`press` across separate CLI calls** didn't fire Vue's v-model/keydown → drove
  the chat send via one eval (set value + input event + Enter keydown).

## Presiding / accounting

Built P1+P2 autonomously while Michael was away ("continue with p1", "lets keep going, p2!") — an
Ammon-style delegated build. Kept each phase a tested commit + a live e2e + a screenshot, checkpointed
between phases, cleaned every test artifact (sessions, work_queue rows, work items, screenshots),
rebuilt the `ui` container 3× (P0 view, P1, P2), and left the rig untouched ($0, local model). All
the SQL is additive/reversible (a new agent + functions). Autonomy stayed paused throughout (his
call — GPUs free for innovation week); the chat tests are interactive, not the autonomous loop.

## Carry-forward

- **P3** — persist a book source corpus (mirror `yt_transcripts`/`segments` keyed by `book_slug` +
  a frontmatter backlink + the book digester writing it + a backfill) so a book chat can quote the
  book's actual passages. The bigger substrate change of the remaining arc.
- **P4** — polish: dockview layout serialization (toJSON/fromJSON to localStorage), a multi-session
  chat history sidebar, per-message provenance chips (which facet a claim came from).
- A fully NL-chat-driven launcher (the chat agent itself decides to kick off a pipeline) is a
  future enhancement beyond P2's deterministic launcher.
- Michael will compact, then we resume at P3/P4.
