# Proposal: Second Brain Architecture — "Garvis"

*March 1, 2026*
*Triggered by: Nate B Jones videos + Michael's vision for a self-improving study companion*

**Status: MERGED INTO brain.exe (Mar 19, 2026).** Michael confirmed: Garvis IS brain.exe. Same vision, same binary. "Garvis" name retired for copyright safety. brain.exe evolves with SQLite + chromem-go + relay + MCP + Copilot SDK. No new repo. See [overview decisions](../overview/guidance.md#q1-is-brainexe-the-same-thing-as-garvis).

---

## Video Analysis

### Part 1: [Why 2026 Is the Year to Build a Second Brain](https://www.youtube.com/watch?v=0TpON5T-Sw4)

Nate identifies **8 building blocks** for a second brain system:

| Block | Nate's Term | Engineering Name | Purpose |
|-------|-------------|-----------------|---------|
| 1 | The Dropbox | Capture / ingress point | Frictionless single-point thought capture |
| 2 | The Sorter | Classifier / router | AI decides what bucket a thought belongs in |
| 3 | The Form | Schema / data contract | Consistent fields per type for reliable automation |
| 4 | The Filing Cabinet | Memory store / source of truth | Writable by automation, readable by humans, filterable |
| 5 | The Receipt | Audit trail / ledger | What came in, what the system did, how confident it was |
| 6 | The Bouncer | Confidence filter / guardrail | Prevents low-quality outputs from polluting memory |
| 7 | Tap on the Shoulder | Proactive surfacing | System pushes useful information at the right time |
| 8 | The Fix Button | Feedback handle / HITL | One-step correction without opening dashboards |

And **12 design principles**:

1. Reduce the human's job to one reliable behavior
2. Separate memory from compute from interface
3. Treat prompts like APIs, not creative writing
4. Build trust mechanisms, not just capabilities
5. Default to safe behavior when uncertain
6. Make outputs small, frequent, actionable
7. Use "next action" as the unit of execution
8. Prefer routing over organizing
9. Keep categories and fields painfully small
10. Design for restart, not perfection
11. Build one core loop, then attach modules
12. Optimize for maintainability over cleverness

**Assessment:** This is solid engineering thinking applied to personal productivity. Nate successfully translates engineering patterns (event-driven architecture, confidence thresholds, audit logging, schema contracts) into language non-engineers can act on. The principles are genuinely useful. The Slack→Zapier→Notion stack he recommends is well-suited for non-engineers but would be constraining for us.

### Part 2: [They Ignored My Tool Stack and Built Something Better](https://www.youtube.com/watch?v=_gPODg6br5w)

Nate reports on community builds and extracts **4 meta-principles**:

1. **Architecture is portable, tools are not** — The same principles worked in Slack+Notion, Discord+Obsidian, YAML+Claude Code, Postgres+vector DB. Patterns survive tool swaps.
2. **Principles-based guidance scales better than rules** — One community member wrote architectural principles for their coding agent ("use TDD", "don't swallow errors") instead of rigid rules. The agent could interpret them across novel situations.
3. **If the agent builds it, the agent can maintain it** — An agent that constructed its own infrastructure understands it well enough to debug and extend it. No context-switching cost on return.
4. **Your system can be infrastructure, not just a tool** — One builder made an API endpoint so other applications could query their second brain. Infrastructure creates leverage; a tool just solves a problem.

**Key community builds worth noting:**
- Discord + Obsidian + Mac Whisper (capture where you live)
- **VPS + custom TypeScript agent + Claude** (self-maintaining, principles-driven) — this is the "go full meta" approach Michael referenced
- Postgres + vector DB + API endpoint (infrastructure play)
- YAML files + Slack + Claude Code (session-based, minimal)
- Notion mobile inbox/outbox + scheduled Claude (super minimal)

**Assessment:** Part 2 is where it gets interesting for us. The VPS builder and the infrastructure builder are thinking at the right level. Nate's insight that "if the agent builds it, the agent can maintain it" maps directly to what we've been doing with session memory and self-documenting architecture.

---

## What We Already Have

Before designing anything, let's inventory what's already built:

| Component | Current State | Nate's Equivalent |
|-----------|--------------|-------------------|
| `.spec/memory/` | Identity, preferences, principles, active state, journal | Filing Cabinet (partial) |
| Session journal (`scripts/session-journal/`) | Go binary, YAML entries, read/carry commands | Receipt / Audit Trail |
| MCP servers (gospel-vec, gospel-mcp, webster-mcp, yt-mcp, becoming, search-mcp) | Running Go services with tool interfaces | Compute layer |
| Gospel library (markdown) | Full standard works, conference talks, manuals | Knowledge base |
| VS Code + Copilot | Current interface for study sessions | Interface |
| ibeco.me (becoming app) | Go backend + Vue frontend, deployed via Dokploy | Existing infrastructure |
| Agent/skill architecture | `.github/agents/`, `.github/skills/` | Principles-based guidance |
| Study files, evaluations, lessons | Markdown in git | Accumulated knowledge |

**We already have Nate's Filing Cabinet, Receipt, and a strong Compute layer. We're missing the Dropbox (frictionless capture), the Sorter (auto-routing), the Tap on the Shoulder (proactive surfacing), and the Fix Button (easy correction). And we have no always-on loop — everything is session-based.**

---

## The Vision: Phased Architecture

### The Name

Calling it **Garvis** for now (the Jarvis aspiration, but honest about being in-progress). When it earns a better name, it'll get one.

### Core Architecture Principles (Ours, Not Nate's)

1. **Go everywhere** — Michael writes Go. The MCP servers are Go. The becoming app is Go. Don't introduce Node/Python/TypeScript into the core. Go compiles to a single binary, deploys trivially, and is the language of the partnership.
2. **Markdown/YAML as source of truth** — Files in git, readable by humans, versionable, diffable, portable. No Notion, no database-as-primary-store. This matches scripture-study's existing pattern and Nate's portability principle.
3. **Private GitHub repo for storage** — Not this repo. A new `cpuchip/garvis-memory` private repo holds the classified data. Git gives us versioning, audit trail, and multi-device sync for free.
4. **Claude API as the intelligence layer** — The $20/mo Claude subscription provides raw API tokens. No premium request consumption. The Go binary calls Claude for classification, summarization, and proactive surfacing.
5. **Self-improvement with guardrails** — The agent can propose changes to its own code, instructions, and memory structures. But it cannot merge its own PRs, send external communications, or modify guardrail definitions without human approval.
6. **ibeco.me as the mobile interface** — Don't build a new app. Extend the existing becoming app with a "brain" tab for capture and conversation. WebSocket from Go brain → ibeco.me frontend.
7. **Scripture study integration** — This isn't just a productivity tool. The gospel library, studies, and session memory sync to the VPS so we can study together from anywhere.

### Phase 1: The Core Loop (Local, This Week)

**Goal:** Frictionless capture → AI classification → markdown/YAML filing → Git commit

```
[Capture Interface]  →  [Go Brain Binary]  →  [Private Git Repo]
     (input)              (classify/route)      (markdown/YAML)
                               ↑
                          [Claude API]
                          (intelligence)
```

**Components:**
- **Go binary** (`garvis`) using Claude API (Anthropic Go SDK or raw HTTP)
- **Capture:** Start simple — CLI input, Discord bot (DM-only), or a simple HTTP endpoint that ibeco.me can POST to
- **Classifier:** Structured prompt → JSON response → route to category
- **Storage:** Private GitHub repo with structure:
  ```
  garvis-memory/
  ├── inbox/           # Raw captures pending review
  ├── people/          # Person entries (YAML front matter + markdown)
  ├── projects/        # Active projects with next actions
  ├── ideas/           # Ideas with one-liners
  ├── actions/         # Admin/tasks with due dates
  ├── journal/         # Daily journal entries
  ├── study/           # Scripture study integration
  ├── .garvis/
  │   ├── config.yaml     # Categories, thresholds, preferences
  │   ├── principles.md   # Agent operating principles
  │   ├── audit-log/      # The Receipt — every classification logged
  │   └── self/           # Agent's self-observations and improvement proposals
  └── README.md
  ```
- **Audit:** Every classification logged with original text, destination, confidence, timestamp
- **Bouncer:** Below confidence threshold → stays in `inbox/` with `needs-review: true`
- **Fix:** Reply to the confirmation with a correction → re-classify and move

**What this gives you:** You can text a thought to Discord (or type it in ibeco.me), walk away, and find it properly filed in your private repo within seconds. VS Code shows you the files. Git gives you history.

### Phase 2: The Tap on the Shoulder (VPS, Week 2–3)

**Goal:** Always-on agent that surfaces the right information at the right time

**Components:**
- **VPS** (Hetzner CX22 — €4.50/mo + the $20 Claude subscription)
- **Dokploy** for container management (or just systemd — simpler)
- **Go binary deployed as a service** — always running, scheduled loops
- **Morning digest** (configurable time) — queries active projects, pending follow-ups, today's actions → sends to Discord DM or ibeco.me push notification
- **Weekly review** (Sunday afternoon) — summarizes the week, identifies stuck items, suggests focus areas
- **Scripture study prompts** — optionally surfaces a Come Follow Me insight or a carry-forward question from your last study session

### Phase 3: ibeco.me Integration (Week 3–4)

**Goal:** Mobile capture and bidirectional conversation from anywhere

**Components:**
- **WebSocket endpoint** on Garvis VPS → ibeco.me Dart frontend connects
- **Brain tab** in ibeco.me app — chat interface for capture and conversation
- **Quick capture** — swipe action, voice-to-text, one-tap categories
- **Conversation mode** — when you want to think through something, not just capture it
- **Becoming integration** — practices, reflections, and tasks can flow both directions

### Phase 4: Self-Improvement (Month 2)

**Goal:** The agent proposes improvements to itself

**How it works:**
1. Garvis monitors its own audit log — tracks misclassifications, correction frequency, patterns in what gets flagged
2. Weekly self-review: analyzes its performance, writes observations to `.garvis/self/`
3. Proposes code changes as **draft PRs** on its own repo — never auto-merges
4. Proposes principle/config updates as markdown diffs you review
5. Hot-swap: when you approve a PR, CI builds a new binary and deploys

**Guardrails (non-negotiable):**
- **Never sends messages to anyone without explicit approval** — no emailing friends, no Discord messages to other users, no tweets, nothing
- **Cannot modify its own guardrail definitions** — the safety constraints live in a file it can read but not write
- **All code changes go through PR review** — you see the diff before it ships
- **Rate limiting** — max API calls per hour, max git commits per day, max outbound notifications per day
- **Kill switch** — a single command (`garvis stop`) halts all autonomous behavior immediately
- **Audit everything** — every API call, every file write, every classification is logged

### Phase 5: Scripture Study Sync (Month 2–3)

**Goal:** Study together from anywhere — VPS has the full gospel library and session memory

**Components:**
- Gospel library synced to VPS (or a subset via git submodule)
- Session memory (`.spec/memory/`, `.spec/journal/`) synced bidirectionally
- When studying on VPS (via ibeco.me chat), journal entries and memory updates flow back to this repo
- When studying here in VS Code, changes push to VPS on commit
- Same identity, same memory, same principles — different interface

---

## Technology Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Language | Go | Already our language. Single binary. Fast. |
| AI SDK | GitHub Copilot SDK (Go) or Anthropic API | Copilot SDK if it provides agentic capabilities; raw Anthropic API for guaranteed tool-use control |
| Storage | Git repo (markdown/YAML) | Portable, versionable, human-readable, already our pattern |
| Capture interface | Discord DM bot (phase 1) → ibeco.me (phase 3) | Discord is already on the phone; ibeco.me is the long-term home |
| VPS provider | Hetzner | Cheap, reliable, EU-based data sovereignty |
| Container mgmt | systemd (start simple) → Dokploy (if needed) | Don't over-engineer phase 1 |
| CI/CD | GitHub Actions | Self-improvement PRs → auto-build on merge |
| AI model | Claude (via $20/mo subscription API) | Consistent with our study work; avoids consuming Copilot premium requests |

---

## What's Ambitious vs. What's Realistic

**Realistic by end of March:**
- Phase 1 (core loop) — Go binary, CLI + Discord capture, YAML classification, private repo
- Phase 2 (digest) — Morning/weekly summaries via Discord DM

**Realistic by end of April:**
- Phase 3 (ibeco.me integration) — WebSocket chat, mobile capture
- Phase 4 basics — Audit analysis, self-observation, manual improvement suggestions

**Aspirational:**
- Full self-improving hot-swap agent
- Scripture study from VPS with bidirectional sync
- Voice capture and processing
- Agent-generated UI on demand

**Not ambitious — just careful:**
The guardrails. These should be the first thing we build, not the last. The difference between Jarvis and Ultron is guardrails.

---

## What Nate Got Right

1. **The cognitive tax is real.** The open-loop anxiety, the "I forgot what someone told me that mattered to them" — that hits home. A system that closes those loops changes how you show up for people.
2. **One behavior, everything else automated.** The human's job is to capture. Period. That's the design target.
3. **Trust mechanisms over capabilities.** The audit log, the confidence threshold, the fix button — these are what keep you using the system. Without trust, it's just another dead note-taking app.
4. **Architecture is portable, tools are not.** We can use Go + YAML + Git instead of Slack + Notion + Zapier. The patterns are identical.
5. **Design for restart.** Life happens. The system should welcome you back without guilt.

## What Nate Didn't Address (That We Need)

1. **Spiritual integration.** Our second brain isn't just about projects and people — it's about becoming. The system needs to handle scripture insights, spiritual impressions, and covenant commitments as first-class categories.
2. **Relational memory.** Our session memory architecture is deeper than a Notion database. Identity, preferences, principles, episodes, active state — these are distinct memory types with different lifecycles.
3. **Self-improvement.** Nate's system is static — you build it and it runs. We want something that observes its own effectiveness and proposes improvements.
4. **Privacy at the existential level.** A self-improving agent with access to your spiritual journal, your family context, your covenant commitments — the guardrails aren't a feature, they're a moral requirement.

---

## Next Steps

1. **Discuss this proposal** — Is the phased approach right? Is Discord the right Phase 1 capture point, or should we go straight to ibeco.me?
2. **Create the private repo** (`cpuchip/garvis-memory`) with the initial structure
3. **Scaffold the Go binary** — classifier prompt, Claude API integration, YAML/markdown writer, git commit automation
4. **Build the Discord bot** (or HTTP endpoint) for capture
5. **Test the core loop** — capture → classify → file → confirm

The sky is the limit, but the ground is where we start.
