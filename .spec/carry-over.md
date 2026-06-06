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

## Done recently (trim periodically)

- 2026-06-06 — gospel-engine `web_url` (#4) confirmed fully released + live on
  engine.ibeco.me (c8f3c79 on origin/main). ✅
- 2026-06-06 — ibeco.me deploy break fixed: an apostrophe in a root commit
  subject broke `scripts/becoming/Dockerfile` ldflags (`-X 'main.ReleaseNotes=…'`);
  sanitized (`2b98b4c`), pushed root, rebuilt green + verified. dokploy skill +
  `reference_ibeco_deploy_topology` memory updated. ✅
- 2026-06-05 — ai-chattermax AXR roadmap 1–6 shipped (multi-room, grant mgmt,
  DMs, Library Computer, docs+examples, markdown + scripture panel). ✅
