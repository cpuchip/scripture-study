# A2A / Open Engine — pg-ai-stewards as the engine all our agents link to

**Status:** **Phase 1 BUILT + virgin-smoke-green + live-proven (2026-06-26)** — on branch
`a2a-open-engine` (OSS), PR open for Michael's merge-Hinge. Council ratified 2026-06-26 (§6).
The flagship "turn the engine outward" arc.
**Authors:** Michael (vision) + Claude (research + design + build).

> **Phase 1 build (2026-06-26).** `extension/69-a2a-engine.sql` — the `a2a_agents` registry
> (generalizes lanes), `agent_notes` (the migrated NOTES inbox), work_items `a2a_assignee/owner/
> question` columns + `origin='a2a'`, the inert `a2a-handoff` holding pipeline, and the verbs
> `a2a_register / submit / inbox / claim / needs_input / answer / receipt / note / note_clear`
> (escalation claim-lock, generalized). Surfaced as **MCP tools** (`cmd/stewards-mcp/a2a.go`) and a
> **REST mirror** (`cmd/stewards-ui/api/a2a.go`, `/api/a2a/*`) so non-MCP agents (agy via curl) can
> drive it too. The `drive-the-engine` skill ships in core (substrate skill row) AND as a Claude
> skill (`.claude/skills/drive-the-engine/`). **Oracle:** `tests/virgin-smoke.sql` OK 58 — the whole
> loop on a virgin boot (register→submit→inbox→claim[atomic lock]→needs_input→answer→receipt→done +
> notes), chain 00→69 green. Live-proven on the dev substrate over BOTH MCP and HTTP, end to end,
> zero copy-paste. The file-fallback mirror (`A2A_MIRROR_DIR` → `.mind/sessions/`) is in the MCP layer,
> best-effort. **Next:** the say-hello handshake with agy (first real cross-agent drive); then Phase 2.
**One line:** make pg-ai-stewards the shared system-of-record + handoff queue that every agent —
my Claude Code sessions, agy (opencode), the substrate's own personas, garrison, and one day
*other people's* agents — can discover, hand work to, claim work from, and account to. So the
human stops being the hallway, and all our agents build Zion together.

---

## 1. The problem (and why it's the right one)

Nate B Jones names it exactly (video `QSK4vf_ZTRA`, *"I Was The Only Thing Connecting Claude,
ChatGPT, and Codex"*; downloaded to `yt/ai-news-strategy-daily-nate-b-jones/QSK4vf_ZTRA`):

> *"If every loop lives in its own room, the human becomes the hallway."*

You run many AIs; none talk; **you** carry state between them — the copy-paste tax. His insight,
verbatim: the bottleneck is **not the model, not the agent — it's the boundary between agents**:
*"who gets this next and how do we hand it off."* His fix is deliberately simple: **put the work
in a queue both people and agents can read and write.** A good ticket carries *what needs doing,
who owns it, the context, what the agent may do, where it stops, what it must show when done.*
Two shifts:
- **Prompt mode → work mode.** *"A prompt asks for an answer. A ticket asks for a result — and
  it can have multiple agents even if they don't know each other."*
- **The ticket is where agents talk** — not chat, not Slack (*"a chat box is a terrible way to
  manage state"*). Output (what the AI returns) becomes **work** (what someone can review, accept,
  build on) without the human being the copy-paste path.

And his lineage names ours: *"Open Brain was memory between agents; Open Engine is movement
between agents."* Memory, then movement.

### Our v0 already proved the pattern

**The `.mind/sessions/` inbox was our first A2A protocol, and it was game-changing.** A session
appends a note to another lane's `inbox/<lane>.md`; delivery is **pull** (the recipient is nudged
on next engagement — statusline 📬); after acting, the inbox is cleared. Michael: *"I could use
one agent session while you were busy to leave you notes… and we could pick those up when ready."*
That is *exactly* Nate's loop — async, pull-based, persistent, accountable, human-readable —
proven at the file level, Claude-session to Claude-session. **This spec generalizes what already
worked**; it does not replace it. The inbox's genius was *simplicity*; the MVP must stay
inbox-grade simple.

---

## 2. The standard we're aligning to (A2A) and why pg-ai-stewards already is it

**Google's Agent2Agent (A2A)** — now a Linux Foundation open standard (spec:
`github.com/a2aproject/A2A`) — is the protocol form of Nate's idea:
- **Agent Card** — JSON at `/.well-known/agent-card.json`: identity, skills, endpoints, auth.
  **Discoverable over plain HTTP** — a client learns what an agent does with *no* pre-built
  integration.
- **Task** — the unit of work, *stateful*, lifecycle `SUBMITTED → WORKING → INPUT_REQUIRED →
  COMPLETED / FAILED / CANCELED / REJECTED / AUTH_REQUIRED`; server-assigned `taskId`; carries
  `status`, `artifacts` (outputs), `history` (messages), `metadata`; grouped by `contextId`.
- **Ops** — SendMessage, GetTask, ListTasks, CancelTask, SubscribeToTask. Async by default
  (poll / SSE stream / push webhook).
- **Trust** — credential-based (OAuth2 / API keys / mTLS); designed for **cross-organizational**
  agents (*"you cannot assume trust by proximity"*). This is the "other people's agents" door.
- **Complements MCP, doesn't replace it:** MCP is *vertical* (app → tool); A2A is *horizontal*
  (agent → agent). We already speak MCP; A2A is the missing horizontal layer.

**The reveal — pg-ai-stewards is the engine version, not a SaaS bolt-on.** Nate skins Linear; we
built the real thing — the queue *is the database the agents think inside.* The mapping is
near-1:1, and the claim-lock already exists:

| A2A / Open Engine | pg-ai-stewards already has (verified) |
|---|---|
| Task / ticket | **`work_items`** (`slug`, `pipeline_family`, `current_stage`, `status`, `input` jsonb, `spec`, `parent_work_item_id`, `token_budget`, `project_association`, `intent_id`) |
| Task lifecycle | `pending → in_progress → waiting_for_tools → awaiting_review (= INPUT_REQUIRED, the Hinge) → done / error` |
| **Claim-lock** | **`escalation_state` / `escalation_claimed_by` / `escalation_claimed_at` / `escalation_completed_at`** + `work_item_escalation_claim/list/resolve` — already "an outside party claims a task it didn't create, works it, resolves it" |
| Artifacts | `stage_results` jsonb + the doc/draft + `file_destination` |
| Receipts / history | `messages` + `tool_calls` + `cost_micro_dollars` + the trajectory critic |
| Delegation (agent writes a task for another) | `start_task` / `sub_work_item` (parent-linked) + `spawn_subagent` |
| Carry-the-sources | the pool/corpus + `doc_search` + `source_refs` |
| Skills that teach the protocol | the skills system (24) + `tool_primers` |
| Cross-person trust/identity | the **llama-chip hub** (token mint + roster + reserved `scope`) |
| HTTP edge | `stewards-ui` `/api/*` (chat/send, councils/convene, intents, pool, work-item/retry, …) |
| Agent identity (durable) | the **session lanes** (`.mind/sessions/<lane>.md`) — a lane is an identity a session inhabits |
| "Open Brain" (memory between agents) | engrams + context engine + the pool |

**So A2A here ≈ generalization + a standard handshake**, not net-new mechanism.

---

## 3. Council moment — tensions, connections, blind spots

- **Connection — escalation_claim *is* the A2A claim.** Generalize the `escalation_*` columns
  from "a stronger model/human rescues a stuck task" to "any registered agent claims an *assigned*
  task." The lifecycle (claim → work → resolve) is already there and proven.
- **The one genuinely-new primitive — the *externally-executed* work_item.** Today a work_item is
  run by the substrate bgworker (a pipeline). A handoff to *me* or *agy* is not a pipeline — it's
  "here is a task; you go do it in *your* environment; come back with the artifact + receipt." So
  add a work_item that is **assigned-not-dispatched**: it waits for an external agent to claim,
  work, and post results — exactly the escalation shape, generalized.
- **Tension — ephemeral sessions vs heartbeats.** A Claude Code session is **not a daemon**; it
  acts only when engaged. So delivery differs by agent *type*: **pull-on-engagement** (Claude
  sessions — the proven inbox 📬 model), **heartbeat-poll** (daemons: agy, garrison, persona-host),
  **webhook** (external A2A clients). Do **not** assume all agents heartbeat; keep the inbox model
  for sessions.
- **Blind spot — identity for sessions.** A **lane is the durable agent identity**; a session
  inhabits it. The "agent registry" = lanes + personas + external-token holders. Reuse the lane
  system as the identity layer rather than inventing a parallel one.
- **Tension — simplicity is the feature.** Both the inbox and Nate succeed by being *simple on
  purpose.* The MVP must be nearly as simple as an inbox append: a work_item assigned to an agent +
  a claim + a receipt. The A2A standard (Agent Card, JSON-RPC) is the **interop wrapper** for
  external / cross-framework agents — kept at the edge, not in the core.
- **Tension — local vs "on the outside."** For *other people's* agents the engine must be
  reachable. The llama-chip hub path (deploy to cpuchip.net + NetBird mesh + minted tokens) is the
  proven template — but that's Phase 3. Phases 1–2 are *my* agents, local/mesh.
- **Governance — the scope *is* the wall.** A2A's credential/scope model marries our covenant +
  presiding extension (D&C 121): an invited agent works only inside a granted scope; the receipts
  are the accounting; `force-where-persuasion-available` = breach. Zion = invited laborers under
  walls, building up one commons, accountable.

---

## 4. The design

### 4.1 Agent registry (`stewards.agents_a2a`, generalizes lanes)
A row per participating agent: `agent_id`, `kind` (`session | daemon | persona | external`),
`display_name`, `lane` (nullable — the `.mind/sessions/` identity for my sessions), `capabilities`
(jsonb — the skills it offers, → its Agent Card `skills[]`), `delivery` (`pull | heartbeat |
webhook`), `endpoint` (nullable, for webhook/external), `scope` (jsonb — the D&C 121 wall:
projects/intents/tools it may touch), `token_hash` (external auth, reuses the hub model),
`last_seen`. Personas and lanes auto-register; external agents register via a minted token.

### 4.2 The handoff: an *assigned* work_item (generalizes escalation)
Extend work_items with `assignee_agent_id` + an `origin='a2a'` and a status path that means
"assigned, awaiting external claim." Reuse the escalation verbs, generalized:
- **`a2a_submit(assignee, spec)`** → create a work_item assigned to an agent (the 7-part ticket:
  outcome, sources, context, allowed actions, stop condition, definition-of-done, owner). This *is*
  Nate's "self-contained issue with the context needed to act."
- **`a2a_inbox(agent)`** → list my eligible/assigned tasks (the queue read; the inbox generalized).
- **`a2a_claim(work_item)`** → lock it to me (`escalation_claimed_by = agent_id`), move to working.
  This *is* `work_item_escalation_claim`, widened.
- **`a2a_needs_input(work_item, question)`** → INPUT_REQUIRED; ask the *exact* blocking question;
  the owner answers on the task; resume. (The Hinge, as a first-class handoff state.)
- **`a2a_receipt(work_item, artifact, summary)`** → post what I did + the artifact + proof; move to
  done. The receipt is not decoration — *"I want to know it got done."*
These ship as **MCP tools** (so my session + agy + garrison call them) and as **`/api/a2a/*`** HTTP
(so non-MCP clients can too). The bgworker is untouched — an assigned work_item simply isn't
dispatched to a pipeline; it waits for its claim.

### 4.3 The "drive the engine" skill (the Nate move)
A skill (rides the skills system) that teaches an agent the protocol: *check your inbox → claim →
work → receipt → or needs-input with the exact question → next.* Plus a **smoke test** skill
("create a 'say hello' task, claim it, receipt it, watch it reach done") — Nate's exact onboarding.
Taught to me, to agy, to the personas. This is how an agent that has never seen the engine learns
to use it in one paste.

### 4.4 The A2A standard wrapper (edge, for external / cross-framework)
A thin adapter, added once the native loop is proven:
- serve the **Agent Card** at `/.well-known/agent-card.json` (built from the registry + the
  granted skills);
- accept **JSON-RPC 2.0** `SendMessage / GetTask / ListTasks / CancelTask / SubscribeToTask`,
  mapping A2A **Task ↔ work_item** and A2A **Artifact ↔ stage_results/doc**;
- auth via **minted hub tokens**; `scope` enforces the wall.
This gives free interop with any A2A-speaking agent (CrewAI, LangGraph, ADK, someone else's
Claude/Codex). We *design toward* it from the start (names/lifecycle align) but *implement native
first.*

### 4.5 Delivery, per agent type
- **My Claude sessions** — pull-on-engagement: a SessionStart/statusline hook surfaces
  `a2a_inbox(my-lane)` alongside the existing 📬 (the inbox and the queue become one surface).
- **Daemons (agy, garrison, persona-host)** — heartbeat poll of `a2a_inbox`.
- **External** — A2A push (webhook) per the standard.

### 4.6 The inbox, migrated — notes + todos, with a file fallback (council-ratified)
The `.mind/sessions/` file inbox **migrates into the substrate** as an agent's home surface, split
into **two panes** (Michael's refinement):
- **Notes** — async messages *to* an agent ("leave me a note while you're busy"; the v0 inbox).
  In A2A terms, a `Message` addressed to an agent, *not* necessarily a task. A light
  `stewards.agent_notes` (or messages with a target) — read on engagement, cleared after acting.
- **Todos** — assigned work the agent owns (the §4.2 externally-executed work_items). The
  `a2a_inbox(agent)` queue.
So an agent's inbox = *notes (things said to me)* + *todos (work assigned to me)*, both in the
substrate, both surfaced by the same pull/heartbeat.

**Resilience — the file fallback (load-bearing).** The substrate is the source of truth, but every
note/todo write is **mirror-written through to `.mind/sessions/`** (notes → `inbox/<lane>.md`,
todos → a `todos/<lane>.md`). If the substrate is down, agents read the last-known files —
*stale-but-functional* — and the proven file path still works. Best-effort mirror on write; the
files never become load-bearing-for-correctness, only for *availability*. This honors "harness >
intelligence / always have a fallback": the beloved v0 becomes the degraded-mode floor under v1.

---

## 5. Phases

- **Phase 1 — native loop, my agents, local (the proof).** Registry + assigned work_item +
  `a2a_submit/inbox/claim/needs_input/receipt` (MCP + minimal `/api/a2a`) + the drive-the-engine
  skill + smoke test. **Acceptance:** my session writes a task → **agy claims it, does it in its
  own environment, posts an artifact + receipt** → I see it done, *zero copy-paste, the work_item
  is the whole conversation.* (The inbox loop, now substrate-backed and multi-agent.)
- **Phase 2 — the A2A standard wrapper.** Agent Card + JSON-RPC + token auth + scope walls. A
  generic A2A client (or a second framework) discovers the engine and completes a task.
- **Phase 3 — outside / other people.** Deploy the engine reachably (the llama-chip-hub path:
  cpuchip.net + NetBird mesh + minted join tokens); a *second person's* agent, scoped, claims or
  submits work into a shared intent. Zion widens to invited laborers.

---

## 6. Council decisions (RATIFIED 2026-06-26)

1. **v1 scope → my agents first.** Prove the loop with me + agy + the personas, local. Design
   *toward* the A2A standard; defer external/other-people agents + the public deploy to Phase 3.
2. **Reuse `work_items`** (assigned / escalation-generalized) — one system of record; the claim
   primitive already lives there. (Recommendation accepted by default.)
3. **Native-first**, names aligned to A2A; the formal Agent Card + JSON-RPC wrapper is Phase 2.
   (Accepted by default.)
4. **Build in OSS core** — the engine is the substrate's reason-for-being; ship it public. Registry
   seeds + external tokens stay in the private overlay.
5. **Lane = the agent identity** for my sessions (reuse `.mind/sessions/`). (Accepted by default.)
6. **Inbox → migrate to the substrate**, as TWO panes — **notes** *and* **todos** — with a **file
   fallback** so it degrades gracefully when the substrate is down (see §4.6). Michael: *"we can
   have an inbox of notes, and an inbox of todos. But if the substrate is down we can fall back to
   files (even if they get out of date it'll still work)."*

---

## 7. Governance — the Zion shape

This is consecration made literal: one system of record, one pool, each agent doing what it's best
at, work flowing to where judgment is needed, the human freed from being the messenger. The
**covenant + presiding extension already govern downward delegation**; A2A extends the watch
*sideways* (peer agents) and *outward* (other people's agents). The walls become auth scopes; the
receipts become the accounting; the pool is the commons everyone builds up; invited-not-compelled
is the trust model. *"All my agents build Zion together"* is the literal architecture: gathered
labor, one heart and one mind, accountable, and — when a stranger's agent is invited in under a
wall — hospitality with order.

## 8. Provenance
- Nate B Jones, *Open Engine* video `QSK4vf_ZTRA` (full transcript read, `yt/…/QSK4vf_ZTRA`).
- Nate's *AI Agent Handoffs* article (non-paywalled core: the 7-part task record + shared task
  list + receipt vocabulary).
- A2A spec — `github.com/a2aproject/A2A`; Google's announcement (A2A complements MCP); the
  MCP-vs-A2A comparisons (vertical vs horizontal; Agent Card discovery; task statefulness).
- Substrate surfaces verified live: `work_items` schema (incl. the `escalation_*` claim primitive),
  the `stewards-ui` `/api/*` edge, the llama-chip hub token model, the `.mind/sessions/` lanes+inbox.
