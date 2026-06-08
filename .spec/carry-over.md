# Carry-over backlog — ai-chattermax + pg-ai-stewards

Living list of next-actions so nothing gets lost between sessions. Sorted by
**what it needs from Michael**, which is the useful axis. Last updated 2026-06-07.

> Companion to `.mind/active.md` (narrative state) — this file is the flat
> checklist. When an item ships, move it to "Done recently" then trim.

## ★ RATIFIED — build queue (build through the week)

Ratified 2026-06-07 (9pm Sunday, budget reset). Michael chose **"ratify the
backlog, then I build"** — these execute async through the week; time-with-Claude
stays for planning/council/ratify. Order: cockpit P1 → code persona P1.

- **① stewards cockpit P1 — read-only Go CLI** (was ★ FOCUS).
  Spec: `projects/pg-ai-stewards/.spec/proposals/stewards-cockpit-cli.md` →
  **RATIFIED**. Ratified decisions: verbs = `project / board / do / council /
  ratify / watch / review / cost`; planning-state ladder =
  `idea → spec → ratified → building → blocked → done`; connection = **direct
  pgxpool** (port 55433, like persona-host); cards = **un-dispatched work_item**
  (one table, no separate `tracked_items`). **Build P1 = read-only
  `project / board / watch / cost`** (pure SQL reads, zero risk). Then P2 adds
  `planning_state` + project dims (and `carry-over.md` becomes a generated view).
  New code lives at `projects/pg-ai-stewards/cmd/stewards/` (Go, pgxpool).

- **② code persona P1 — `research_codebase` agentic tool + read-only ai-chattermax persona.**
  Spec: `projects/pg-ai-stewards/.spec/proposals/agentic-tools-model-cascade.md` →
  **RATIFIED**. Ratified decisions: flagship repo = **ai-chattermax**; scope =
  **read-only first** (`research_codebase` returns findings + `file:line`
  citations; no edits/PRs until it earns it). Build P1 = a `researcher-flash`
  agent_family (deepseek-v4-flash + read-only repo tools: grep/glob/read) + the
  `research_codebase` MCP tool (a `consult_subagent` preset + return contract);
  smoke it on a real ai-chattermax question. P2 = wire it into a read-only code
  persona pipeline in ai-chattermax, live-test in a room.

## Needs Michael's decision first (hard gate)

- **CT2 — substrate self-context management** (task #118). Spec complete:
  `projects/pg-ai-stewards/.spec/proposals/substrate-self-context-management.md`.
  §§1–6 (agent-callable compress/expand/mute/pin + addressable handles +
  pressure line + circuit-breaker lock) plus the §7 expansion driven 2026-06-05:
  - §7.1/§7.2 — durable, **removable** self-notes (Hermes self-curation) +
    system prompt split into immutable-base + model-curated notes block.
  - §7.3 — model edits its own BASE prompt; gated propose→ratify, off by default.
  - §7.4 — working tags: `context_set_tag(tag)` auto-stamps subsequent
    messages/tool calls; `fold_tag` sweeps a finished task out in one call.
  Held because the build restarts the live substrate Starlet + the Computer ride
  on. **Action: Michael reads → ratifies → I build.**

- **claude-worker dispatch** — spec complete:
  `projects/pg-ai-stewards/.spec/proposals/claude-worker-dispatch.md`. Hand more
  autonomous work to Claude (+ gpt-5.5/gemini) as **CLI workers** dispatched from the
  substrate, so it draws the new **`claude -p` Agent-SDK credit pool** (separate from
  interactive; ~$200/mo on Max-20x, live 2026-06-15) instead of wasting it. Dumb host
  poller → `claude -p` on demand (zero idle tokens); bins-1/2-only unattended; spend
  guard. **Engine decided 2026-06-07: Model A = `claude -p` to start** (Michael's lean);
  Model B = the reverse-dispatcher (a long-lived *normal* interactive session draws the
  generous interactive pool — most bang for the buck, but reserved for councilled queues,
  never an automation farm). **Still pending:** (a) the **second connector** for substrate
  redundancy/capacity — opencode_go is live; choose Atlas / GLM / Ollama; (b) the agent-SDK
  pool itself doesn't exist until **2026-06-15**, so Model-A async can't spend it before
  then. Build (host poller = non-Claude code) is cheap and can start any time.

## I can do now (no ratification, low Claude-token cost)

- **Delete-message endpoint** (ai-chattermax) — closes the "demo message lingers
  in 10-Forward" gap; small backend route + UI affordance.
- **Gemini reference client** in `projects/ai-chattermax/examples/` — mirrors the
  LM Studio one; the substrate `persona-turn-gemini` pipeline already exists.
- **Restore the per-message rate ceiling** (ai-chattermax) — re-assert the hard
  room-enforced cap that the platform rebuild dropped.

## Design pass with Michael (~5 min), then I build

- **Engineering / repo-reader persona** — **RATIFIED 2026-06-07** as code persona
  (build-queue item ②): repo = **ai-chattermax**, **read-only first**, backed by the
  agentic `research_codebase` tool (not raw coder tools in the persona). P2 of item ②
  wires the pipeline + live-tests it in a room. A propose-changes engineering persona
  stays a later, gated step (drafts a PR, never merges — the Hinge).
- **D&D MVP** — the original target: DM-assistant + NPCs + in/out-of-character
  side channels (sub-personas). The sub-persona model is a new surface.
- **Moderation (#11)** — policy + tools. Rate-ceiling restore (above) is the
  mechanical half; moderation policy needs Michael's input.

## Token-budget strategy (the lens for all of the above)

- **Claude tokens are the scarce premium resource** — spend on judgment, design,
  discernment, voice-sensitive prose (the book/studies), the Hinge, hard
  cross-layer debugging, and **shepherding the substrate's output**.
- **Route long-horizon VOLUME work through pg-ai-stewards** (kimi/qwen/glm/
  deepseek/minimax/local LM Studio — NOT Claude tokens; cost-tracked at cents/run).
  Canon walks, batch evaluations, research gathering, transcript ingestion,
  code-pr builds. Claude orchestrates + reviews; the substrate executes.
- Lesson from the BoM walk (2026-06): a long-horizon walk done directly on Claude
  ≈ 2 days of weekly Max budget. The same class of task run through the substrate
  costs a fraction of Claude budget. **Migrate that work to the substrate.**

## Seeds / ideas (not yet decided)

- **Agentic tools / model-cascade in pg-ai-stewards** — **RATIFIED 2026-06-07, now in
  the build queue above (item ②).** (Kept here for the discriminator note.) The big
  model orchestrates; some "tools" are themselves cheap-model agentic calls that do the
  heavy lifting and return curated results (like WebFetch/`summarize_url` already do for
  web). Mirror of [[claude-worker-dispatch]] — that escalates UP to Claude; this delegates
  DOWN to cheap models. **Discriminator (= the gospel line "delegate execution, not
  discernment"):** agentic-wrap a subtask ONLY if it's language/judgment, large vs. its
  instruction, with cheap deterministic verification (tests/compile/exact-match). Don't
  wrap mechanical/exact ops (raw grep, precise edits) — loses determinism, often a token
  wash. Open Qs still live: cheap model tier (deepseek-v4-flash vs a stronger flash); sync
  vs async for a chat turn (lean: sync, tight timeout); whether read-only graduates to a
  gated PR-proposing persona later.

- **Harness-leveling experiment for cheap/local models** (2026-06-06, Michael). Can
  curated instructions + context (memory/intent/examples/tight specs/critic/ground-truth)
  level up free local models (qwen3.6-27b, gemma-12b via LM Studio — 100M tokens/day,
  $0 to Michael) toward usable autonomy? Hypothesis: harness **externalizes direction** so
  weak models infer less → converts them from "needs micromanagement" to "competent
  EXECUTOR of well-specified, verifiable tasks." It will NOT grant unspecified-direction
  inference (that stays with Claude). Test rig = the substrate (it already injects harness
  per agent_family): A/B bare-prompt vs full-harness on a bounded task, measure quality
  delta. Gospel/book frame: the harness is how you raise up the less-capable (teach +
  context + verify) — the substrate as a school. Pairs with [[agentic-tools-model-cascade]].

- **Steal-list from Dave's workflow framework** (2026-06-06; see
  `docs/ai-utilization-landscape-2026.md` §7 — Dave = the `dave-rule` Dave;
  `external_context/workflow`). Independent-convergence peer system; borrow: (1) explicit
  **"AI Freedom"** spec section (name what's intentionally unconstrained — we over-specify);
  (2) **invariant-traceability** rule in the critical-analysis gate (each constraint traces
  to a requirement — anti-invented-constraint, twin of the cite-count rule); (3) a named
  **SideQuest** lightweight lane (we have the bins, not the lane); (4) fold **ODD/SRE depth**
  into our debug agent; (5) **"file-first / avoid double-dipping"** discipline. Plus an
  **audit**: which substrate bgworker gate-evals are grounded-in-artifact vs cold-start
  role-pattern (Dave's "agent task suitability" caution on our dispatched critic, cv6).

## Done recently (trim periodically)

- 2026-06-06 — **md-mcp** (markdown planning tools): Michael forked happydave/md-mcp to
  `projects/md-mcp`; built the MCP **server wiring** (the repo was library-only — no
  main.go) via the official go-sdk + added 3 tools (`md-section-append`,
  `md-section-move`, `md-frontmatter-set`). go test/vet green + live stdio smoke (13
  tools, append + frontmatter persist). Registered in `.mcp.json` (**restart to load the
  tools**). Upstream PR → [happydave/md-mcp#1](https://github.com/happydave/md-mcp/pull/1). ✅
- 2026-06-06 — gospel-engine `web_url` (#4) confirmed fully released + live on
  engine.ibeco.me (c8f3c79 on origin/main). ✅
- 2026-06-06 — ibeco.me deploy break fixed: an apostrophe in a root commit
  subject broke `scripts/becoming/Dockerfile` ldflags (`-X 'main.ReleaseNotes=…'`);
  sanitized (`2b98b4c`), pushed root, rebuilt green + verified. dokploy skill +
  `reference_ibeco_deploy_topology` memory updated. ✅
- 2026-06-05 — ai-chattermax AXR roadmap 1–6 shipped (multi-room, grant mgmt,
  DMs, Library Computer, docs+examples, markdown + scripture panel). ✅
