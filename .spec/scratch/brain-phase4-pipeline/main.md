# WS1 Phase 4: Brain Pipeline Maturity — Scratch

**Binding problem:** brain.exe classifies *what* an entry is (study, idea, project, etc.) but not *how ready* it is to act on. Every entry lands at the same readiness level regardless of whether it's a half-formed thought or a fully specced workstream. The human ends up being the maturity router — manually deciding what needs research, what needs planning, and what's ready for execution. This is the manual overhead the brain is supposed to eliminate.

**Created:** 2026-04-03

---

## Research Findings (from council session)

### Current Pipeline
```
Capture (Discord/Relay/Web) → Classify (LM Studio 9b, category only) → Store → Route (suggest/auto) → Agent → Review
```

### What Exists
- Classification: 6 categories (people, projects, ideas, actions, study, journal)
- Routing: category → agent mapping with modes (suggest/auto/none)
- Agent pool: named sessions, lazy creation, Copilot SDK
- Review queue: API endpoints for accept/reject/dismiss
- Token budgets: warning + hard cap per session
- Custom agents: study, journal, plan mapped with Infer:true

### What's Missing
- **Readiness dimension**: no assessment of maturity level
- **Multi-step pipeline**: no way for an entry to go through research → plan → spec → execute
- **Transition logic**: no way to advance an entry from one stage to the next
- **Human interaction UI**: no way to review, refine, approve stage transitions
- **Scenarios / success criteria**: no structured way to define "done" for spec validation

### Prompt Injection Note
The classifier is vulnerable to content that reads like instructions (Michael's own brain entries triggered research behavior instead of classification). Structural fix needed: entry text as delimited opaque data, not raw user message injection.

---

## Key Design Questions (in progress)

### Q1: What does the maturity ladder look like?
Initial proposal:
```
raw → researched → planned → specced → executing → verified → done
```
TBD: Is this too many stages? Should some collapse?

### Q2: Where does the human interact?
- Brain-app (Flutter) — existing but limited
- ibeco.me (web) — richer UI, already has auth
- VS Code chat — existing workflow
- All of the above?

### Q3: What does "review" look like at each stage?
- Raw → Researched: "Here's what I found. Worth pursuing?"
- Researched → Planned: "Here's the plan shape. Approve/revise?"
- Planned → Specced: "Here's the full spec with scenarios. Ready to build?"
- Specced → Executing: automatic if approved
- Executing → Verified: "Scenarios passed/failed. Accept?"

### Q4: Which models at which stages?
- Classification: LM Studio (free, local)
- Research: Haiku/Flash (cheap, fast)
- Planning: Sonnet (mid-tier reasoning)
- Execution: Opus/Sonnet (full capability)
- Verification: Haiku (structured checking)

### Q5: What's the interaction pattern?
- Chat-based? (like current brain-app)
- Form-based? (structured fields)
- Hybrid? (chat with structured output)

---

## Q&A Pass 1 — Where Michael Lives

### Answers
- **Where do you go after brain-app?** VS Code chat — "I end up here anyway"
- **What kills momentum?** "I forget what's queued" — out of sight, out of mind
- **Review cadence?** Daily morning review — part of a routine
- **Interaction depth?** Both — quick triage for simple, conversational for complex

### Implications
1. **VS Code is the primary surface.** The pipeline must surface well in VS Code chat, not just ibeco.me or brain-app. This means MCP tools or a chat-first interaction pattern.
2. **Forgetting is the enemy, not complexity.** The core problem isn't "too hard to work on items" — it's "I don't know they exist." This means proactive surfacing matters more than sophisticated UI.
3. **Daily morning review is the ritual.** This is the natural checkpoint. The pipeline needs a "morning brief" — here's what's queued, what moved overnight, what needs your input.
4. **Two modes needed:** Quick approve/reject for clear items, conversational refinement for ones that need shaping.

## Workflow Sketches

*(to be filled after Pass 2)*

---

## Q&A Pass 2 — Workflow Shape

### Answers
- **Morning brief:** Pull-based — "show me my queue" via MCP tool in VS Code chat
- **Research outputs:** Scratch files in workspace (study/.scratch/ or .spec/scratch/)
- **Conversational shaping:** VS Code chat with MCP tools — no new UI surfaces
- **Maturity stages:** raw → researched → planned → specced → executing → verified — confirmed as right
- **Auto-execute boundary:** Always gate before execution — nothing runs without approval

### Implications
1. **MCP-first interaction.** The primary interface is brain MCP tools called from VS Code chat. New tools needed:
   - `brain_queue` — show pipeline status grouped by maturity stage
   - `brain_advance` — advance an item to next stage (triggers the appropriate agent pass)
   - `brain_review` — show an item's research/plan output for approval
   - `brain_refine` — conversational mode for shaping an item
2. **Scratch files are the artifact.** Research → scratch file. Plan → scratch file. Spec → proposal file. This matches the existing convention perfectly. Files are durable, context is not.
3. **Human gate before every execution.** No auto-execute even for fully specced items. The gated autonomy decision holds. The pipeline can PREPARE everything, but Michael pulls the trigger.
4. **VS Code chat IS the pipeline UI.** No need to build new brain-app screens for this (yet). brain-app stays as the capture surface. VS Code chat is the refinement + review surface.

### Emerging Workflow

#### The Daily Loop
```
Morning: "show me my queue" (brain_queue MCP tool)
  → See items grouped by stage: 3 raw, 2 researched, 1 planned
  → Pick one: "advance item X" (brain_advance)
  → Brain runs research pass → writes scratch file → marks "researched"
  → Next morning (or same session): "review item X" (brain_review)
  → Read scratch file, approve/revise/reject
  → "advance item X" → plan pass runs → writes to scratch file
  → Repeat until specced
  → "execute item X" → human-gated → agent runs → output → verify
```

#### The Capture-to-Queue Flow
```
Brain-app (phone, Discord, relay) → capture → classify (category + maturity=raw) → store
  → Shows up in next "brain_queue" call
```

---

## Decisions Made

### From Q&A Pass 3

- **Research pass scope:** Smart mix — internal (existing studies, proposals, brain entries, prior art) AND external (web search, YouTube, articles) weighted by what the item needs.
- **Pipeline categories:** Ideas + Projects + Study. Actions and journal don't need maturity pipeline (actions have a clear "done" state, journal is reflective not actionable).
- **Spec template — what makes "specced" ready?** Scenarios. Everything else (binding problem, success criteria, scope, phasing) is important but scenarios are the gating field. If you can't write a testable scenario for what "done" looks like, the spec isn't ready.
- **ibeco.me role:** (Michael stepped away — agent decision) ibeco.me serves as the **dashboard/glance surface** — pipeline status at a glance, accessible from phone or browser. VS Code chat is the **work surface** where refinement happens via MCP tools. brain-app remains the **capture surface**. Three surfaces, three roles, no overlap. **CONFIRMED by Michael (Apr 3).** Additional nuance:
  - ibeco.me is not just read-only — Michael should be able to kick things off, iterate through progressive entries on a topic
  - brain-app = view status/progress/items + capture, intentionally lightweight
  - VS Code = full work surface, everything including triage, iterate, study, build
- **Prompt injection:** Fix in Phase 4 as part of pipeline integrity. Classifier needs structural protection (delimiters, input validation) not just prompt hardening.

### From Q&A Passes 1-2

- VS Code chat is the primary interaction surface (MCP tools)
- Pull-based morning brief ("show me my queue")
- Scratch files are the artifact at every stage
- Human gate before every execution — no auto-execute
- Maturity stages confirmed: raw → researched → planned → specced → executing → verified
- Quick triage for simple items, conversational refinement for complex ones

### From Pass 4 — Governance + Testing (Apr 3)

- **Option A confirmed** for classifier enhancement: post-classification maturity assessment, keep classifier simple
- **Per-layer governance documents added.** 5 docs mapping to 11-step creation cycle:
  - `classifier-stewardship.md` (Steps 1-3): never act on content
  - `maturity-stewardship.md` (Steps 1-3): assess readiness, not quality
  - `research-covenant.md` (Steps 1-5): internal first, write everything, never decide
  - `plan-covenant.md` (Steps 1-7): produce binding problem + scenarios, flag when idea fails analysis
  - `execution-covenant.md` (Full cycle): build per spec, stay within boundaries
- **Governance gap confirmed:** all brain governance currently hardcoded in Go — functional but invisible, not auditable, agents can't reflect on their own boundaries
- **Documents loaded into system messages at runtime** via BuildSystemMessage — same pattern as .agent.md files but for boundary docs
- **Test harness: sandbox approach, not Docker.**
  - `.gitignored` `test-sandbox/` directory with fresh brain.db per run
  - `pipeline-bench` CLI command (extends classify-bench pattern)
  - Seeds test entries including prompt injection attacks
  - `--research` and `--plan` flags to test deeper pipeline stages
  - `--clean` flag to wipe sandbox
  - Brain's env-var config (`BRAIN_DATA_DIR`) makes isolation trivial
  - Docker deferred to CI/CD phase — Copilot SDK auth forwarding is possible but painful for iteration loop
- **Phase 4a expanded:** now includes governance docs + test harness (1-2 sessions instead of 1)
- **Each subsequent phase includes its layer's governance doc as a deliverable**

### From Pass 5 — Surface Roles + Docker Prior Art (Apr 3)

- **Surface roles confirmed.** ibeco.me = dashboard + interaction, VS Code = primary work, brain-app = capture + status view.
- **ibeco.me capabilites refined:** not just glance — should support kicking off pipeline stages, iterating on entries. The REST API (§5.9) enables this.
- **Docker prior art saved for CI/CD phase:**
  - [copilot_here](https://github.com/GordonBeeming/copilot_here) — Copilot CLI in Docker, `gh` auth forwarding solved, Go image variant available. The auth problem we deferred is already solved here.
  - [Docker Sandboxes (sbx)](https://www.docker.com/products/docker-sandboxes/) — microVM isolation for AI agents. `winget install Docker.sbx` on Windows. Supports Copilot CLI natively. MicroVM > container for agent isolation.
- **Emergency stop added (§5.12).** Michael wants a kill switch accessible from brain-app and ibeco.me — if an agent goes runaway, kill it from phone/browser without being at the desk. `POST /api/emergency-stop` (all) or `/api/emergency-stop/{id}` (specific). Also `brain_stop` MCP tool for VS Code. `GET /api/status` shows what's running. Implementation: cancel context on running goroutines, SIGTERM Copilot CLI subprocesses. Added to Phase 4d deliverables.
