# Carry-over backlog — ai-chattermax + pg-ai-stewards

Living list of next-actions so nothing gets lost between sessions. Sorted by
**what it needs from Michael**, which is the useful axis. Last updated 2026-06-06.

> Companion to `.mind/active.md` (narrative state) — this file is the flat
> checklist. When an item ships, move it to "Done recently" then trim.

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
  guard. Also maps the **sub-with-API connector** options (opencode_go + Atlas/GLM/
  Ollama) for substrate redundancy + capacity. **Action: Michael reads → ratifies →
  build P1 after Sunday reset.** Build is cheap (host poller = non-Claude code).

- **★ FOCUS — stewards cockpit CLI (Option A)** — spec complete:
  `projects/pg-ai-stewards/.spec/proposals/stewards-cockpit-cli.md`. A `stewards` Go CLI
  (pgxpool to the substrate, like persona-host) so Michael drives the substrate himself:
  `project / board / do / council / ratify / watch / review / cost` — `project` selects a
  sticky **active-project** context (like a kubectl context) that scopes the work-item
  verbs; **ratify = input Hinge**
  (approve a plan to build) and **review = output Hinge** (approve finished work);
  **council** = pre-ratify critical-analysis pass (surfaces tensions; the human decides).
  Closes the #1 gap (useful-to-agent-not-to-Michael).
  Includes the **project board** (work_items gain `project` + `planning_state`;
  `carry-over.md` becomes a generated view) and the **token dashboard by project × model**
  (shared backend for the CLI `cost` verb + a stewards-ui panel — Michael's add). P1 =
  read-only (board/watch/cost). **Action: Michael confirms verbs + planning-state ladder →
  ratify → build.** Chosen focus among the 3 cockpit shapes (CLI now, stewards-ui next, ai-chattermax-as-cockpit later).

## I can do now (no ratification, low Claude-token cost)

- **Delete-message endpoint** (ai-chattermax) — closes the "demo message lingers
  in 10-Forward" gap; small backend route + UI affordance.
- **Gemini reference client** in `projects/ai-chattermax/examples/` — mirrors the
  LM Studio one; the substrate `persona-turn-gemini` pipeline already exists.
- **Restore the per-message rate ceiling** (ai-chattermax) — re-assert the hard
  room-enforced cap that the platform rebuild dropped.

## Design pass with Michael (~5 min), then I build

- **Engineering / repo-reader persona** — a chat persona backed by a real repo
  (ai-chattermax / pg-ai-stewards) that reads its own codebase and answers code
  questions. Needs a NEW tool-scoped substrate pipeline (kept coder/repo tools
  OUT of the librarian on purpose). Decide: which tools, which repos, read-only
  vs propose-changes. The "next app."
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

- **Agentic tools / model-cascade in pg-ai-stewards** (2026-06-06, Michael). The big
  model orchestrates; some "tools" are themselves cheap-model agentic calls that do the
  heavy lifting and return curated results (like WebFetch/`summarize_url` already do for
  web). Mirror of [[claude-worker-dispatch]] — that escalates UP to Claude; this delegates
  DOWN to cheap models. Together = the full stewardship tree (cheap ← Claude ← Michael).
  **Discriminator (= the gospel line "delegate execution, not discernment"):** agentic-wrap
  a subtask ONLY if it's language/judgment, large vs. its instruction, with cheap
  deterministic verification (tests/compile/exact-match). Don't wrap mechanical/exact ops
  (raw grep, precise edits) — loses determinism, often a token wash. First candidate:
  read-only `research_codebase` (deepseek-v4-flash explores + curates). Edits later, gated
  on ground-truth. Built on existing `spawn_subagent`/`consult_subagent`. A/B the savings
  (only real on large subtasks). **Spec written 2026-06-06:**
  `projects/pg-ai-stewards/.spec/proposals/agentic-tools-model-cascade.md` — flagship
  = the **ai-chattermax code/repo-reader persona** (its repo tools become cheap-model
  agentic calls: `research_codebase` via deepseek-flash). Awaiting ratification.

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
