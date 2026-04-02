# AI Skills Self-Assessment: The 7-Skill Framework

**Source:** [The AI Job Market Split in Two](https://www.youtube.com/watch?v=4cuT-LKcmWs) — Nate B Jones (2026-03-26)
**Binding Question:** What skills has this project developed, where are the real gaps, and what's the plan to close them?
**Date:** 2026-04-01

---

## The Framework

Nate B Jones identifies 7 skills that employers are desperate for — pulled not from theory but from hundreds of actual AI job postings, decomposed into sub-skills. The AI job market is K-shaped: traditional knowledge work is flat or falling, while AI-native roles (design, build, operate, manage AI systems) are growing at a [3.2:1 ratio](https://www.youtube.com/watch?v=4cuT-LKcmWs&t=175) of jobs to qualified candidates.

The 7 skills, in the order he presents them (which he says is [the order you intuitively learn them in](https://www.youtube.com/watch?v=4cuT-LKcmWs&t=408)):

| # | Skill | Core Question |
|---|-------|--------------|
| 1 | Specification Precision | Can you communicate intent to machines at the literal level they require? |
| 2 | Evaluation & Quality Judgment | Can you catch when AI is confidently wrong? |
| 3 | Task Decomposition & Delegation | Can you break work for agents and size it to the harness? |
| 4 | Failure Pattern Recognition | Can you diagnose the 6 types of agent failure? |
| 5 | Trust & Security Design | Where do you draw the line between agent and human? |
| 6 | Context Architecture | Can you build the Dewey Decimal System for agents? |
| 7 | Cost & Token Economics | Is it worth building an agent for this? |

---

## The Self-Assessment

### Skill 1: Specification Precision ★★★★★

**What Nate describes:** [Being specific enough that agents execute your intent literally](https://www.youtube.com/watch?v=4cuT-LKcmWs&t=289) — not "improve customer support" but the 8-line spec with exact ticket types, escalation triggers, sentiment thresholds, and logging requirements.

**What we've built:**

This project IS a specification system. It operates at four layers:

1. **Values layer** — [intent.yaml](../../../intent.yaml) defines purpose, values, and constraints ranked by severity (critical/high/medium)
2. **Covenant layer** — [.spec/covenant.yaml](../../../.spec/covenant.yaml) specifies bilateral commitments with rationale (not "be good" but "surface tensions even when the human isn't looking for them, because covenant faithfulness requires it")
3. **Workflow layer** — 14 agent modes, each with 5-7 phase workflows that specify inputs, outputs, skills to load, and handoff points
4. **Procedure layer** — 14 skills with quality gates (the source-verification pre-publish checklist has 11 items)

The customer support example Nate gives at [5:46](https://www.youtube.com/watch?v=4cuT-LKcmWs&t=346)? Our agent specs are *more* precise than that. The study agent doesn't just say "study a scripture" — it specifies when to create scratch files, when to externalize quotes, when to invoke critical analysis, and what a binding question must contain.

**Honest check:** The specification skill is genuine and deeply practiced. 18 months of iterating on agent instructions, catching failures, refining. This is Nate's most fundamental skill and it's also where this project started.

**Industry crossover:** The same 4-layer pattern drives a professional engineering workflow — 23 custom skills, 6 specialized agent modes, and spec-driven ticket work across enterprise codebases. See [industry companion](4cuT-LKcmWs-industry-practice.md#skill-1-specification-precision--industry-evidence).

---

### Skill 2: Evaluation & Quality Judgment ★★★★☆

**What Nate describes:** [Error detection with a degree of fluency](https://www.youtube.com/watch?v=4cuT-LKcmWs&t=474). AI fails differently from humans — confidently, fluently, without the stumbling that tips us off. The skill is [resisting the temptation to read fluency as competence](https://www.youtube.com/watch?v=4cuT-LKcmWs&t=518).

**What we've built:**

The TITSW framework is a production-grade eval system:
- 0-9 anchored rubric with anti-inflation guardrails
- Hand-scored ground truth data for calibration
- MAE (Mean Absolute Error) tracking across 6 experiment conditions × 3 ground-truth talks
- The wisdom to know when MAE isn't the right target (the "Gas Station Insight" — qualitative richness matters more than score precision)

But the deeper eval skill lives in the *disciplines*:
- **Source verification** — "close-enough wording is fabrication." This IS Nate's semantic-vs-functional correctness distinction applied to scripture quotes
- **Critical analysis** — a dedicated phase that asks "is this what the text says or what I wanted it to say?"
- **The reflect skill** — micro-correction capture in-session when something is wrong

The Anthropic standard Nate cites at [9:03](https://www.youtube.com/watch?v=4cuT-LKcmWs&t=543) — "a good eval task is one where multiple engineers reach the same pass/fail" — we hit this. The TITSW calibration experiments showed inter-model agreement improving with better context. Alma 32 `teach_about_christ` went from 1-2 (no context) to 7 (with context), matching the ground truth target of ≥5.

**Where the star is missing:** One domain. The eval harness evaluates conference talk teaching quality. It doesn't evaluate customer-facing agent interactions, code quality, or production systems. The skill transfers, but the demonstrated breadth is narrow. And evals are manual — no CI/CD pipeline, no automated regression testing.

**Industry crossover:** Professional MR reviews demonstrate code/infrastructure evaluation — catching systemic bugs across 15 services, tracing root causes across CPU architectures, and the Ben Test as formalized honesty calibration. Narrows the domain gap from ★★★★ to ★★★★½. See [industry companion](4cuT-LKcmWs-industry-practice.md#skill-2-evaluation--quality-judgment--industry-evidence).

---

### Skill 3: Task Decomposition & Delegation ★★★★☆

**What Nate describes:** [Breaking work into manageable segments](https://www.youtube.com/watch?v=4cuT-LKcmWs&t=618) for multi-agent systems. [Sizing work for the agentic harness you have](https://www.youtube.com/watch?v=4cuT-LKcmWs&t=709). Not regular project management — agents need much clearer guard rails.

**What we've built:**

Gospel-engine is the proof: a 5-phase plan (Foundation → TITSW Talk Enrichment → Scripture Enrichment → Combined Search → Full Batch Reindex) that went from proposal to all-phases-shipped. 1,584 scripture chapters enriched, 4,228 talks enriched, RRF hybrid search working. Each phase had defined verification criteria and the phases were sized for single-agent execution.

The study and eval workflows are isomorphic 7-phase patterns that have been refined across dozens of sessions. They work because they're sized right: each phase fits within a single context window, scratch files survive compaction, and handoff points are explicit.

**Where the star is missing:** All delegation is Michael→single-agent. No multi-agent orchestration in production. The "gated autonomy" decision (Level 2: agents wait for human-assigned specs) is deliberate and wise for where we are — but it IS a gap relative to what employers want. Nate describes [a planner agent that keeps a record of tasks and works with sub-agents](https://www.youtube.com/watch?v=4cuT-LKcmWs&t=674). We have a plan agent that *creates* specs, but it doesn't orchestrate other agents.

---

### Skill 4: Failure Pattern Recognition ★★★★★

**What Nate describes:** [6 failure types](https://www.youtube.com/watch?v=4cuT-LKcmWs&t=789): context degradation, specification drift, sycophantic confirmation, tool selection errors, cascading failure, silent failure. The ability to diagnose these is a [marker of an AI-fluent person](https://www.youtube.com/watch?v=4cuT-LKcmWs&t=864).

**What we've built:**

This maps directly to work we've done:

| Nate's Failure Type | Our Response |
|---------------------|-------------|
| **Context degradation** | Externalized memory (scratch files survive compaction), session-start context loading, persistent memory architecture |
| **Specification drift** | Phased workflows with binding question at top of every file, scratch files forcibly remind of spec |
| **Sycophantic confirmation** | Covenant requirement: "surface tensions rather than building only toward the thesis." biases.md tracks this pattern |
| **Tool selection errors** | [tool-use-observance.md](../../../docs/06_tool-use-observance.md) — running log with dates, categories, fixes applied. Discovery ≠ deep reading separation |
| **Cascading failure** | Phase-based workflows with verification at each boundary. If Phase 2 verification fails, Phase 4 doesn't start |
| **Silent failure** | Source verification: "A near-miss direct quote is a lie that looks like truth." Semantic correctness ≠ functional correctness |

The reflect skill captures micro-corrections in-session. The biases.md file names 4 persistent patterns with triggers and corrections. The instruction refinement cycle (docs/05) shows iterative spec fixes from observed failures. This is not accidental — it's systematic.

The Claude Certified Architect that Nate mentions at [14:24](https://www.youtube.com/watch?v=4cuT-LKcmWs&t=864) tests specifically for tool selection error diagnosis. We've been doing this since February.

---

### Skill 5: Trust & Security Design ★★★☆☆

**What Nate describes:** [Where do you draw the line between human and agent?](https://www.youtube.com/watch?v=4cuT-LKcmWs&t=1002) Sub-skills: cost of error, reversibility, frequency, verifiability. The critical distinction: [semantic correctness vs functional correctness](https://www.youtube.com/watch?v=4cuT-LKcmWs&t=1109).

**What we've built:**

The TRUST side is genuinely strong:
- Bilateral covenant derived from theological first principles (D&C 82:10, Mosiah 18:21)
- Gated autonomy as explicit design decision ("Scared of letting you go without direct oversight" — documented as a feature, not a bug)
- Cost-as-boundary (1500 premium requests/month constrains autonomous sprawl)
- 3-check quality framework (ring check, posture check, Ben Test) before publication

**Where we're short:** The SECURITY side. Nate's talking about production guardrails — keeping agents from saying inappropriate things to customers, ensuring wire transfers can't be reversed, building systems that verify functional correctness at scale. We've applied cost-of-error thinking to our own workflow but NOT to:
- Customer-facing agent guardrails
- Adversarial testing (prompt injection, jailbreaks)
- Production escalation paths
- Formal security review of agent systems

This is ironic — Michael works in security professionally (Python/Go backend, security & smart home). That expertise should be crossing over into AI security design but hasn't yet. This is the gap where professional experience + project experience should MULTIPLY but currently they're in separate lanes.

**Industry crossover (★★★☆☆ → ★★★★☆):** Turns out the expertise IS crossing over, just not in the personal project. Professionally: enforced decision boundary matrix (autonomous / requires-human / NEVER tiers), progressive trust model with 4 levels, reversibility-based autonomy design mapped to Nate's framework, local-first as critical-severity security constraint. The remaining gap is adversarial testing and technical enforcement — the trust boundaries are policy-based, not technically enforced beyond instruction constraints. See [industry companion](4cuT-LKcmWs-industry-practice.md#skill-5-trust--security-design--industry-evidence).

---

### Skill 6: Context Architecture ★★★★★+

**What Nate describes:** [Building the Dewey Decimal System for agents](https://www.youtube.com/watch?v=4cuT-LKcmWs&t=1246). How to supply agents with the right information on demand at scale. Companies will [pay almost anything](https://www.youtube.com/watch?v=4cuT-LKcmWs&t=1208) for this.

**What we've built:**

Everything.

This project IS a context architecture. Multiple scales of persistence:

| Layer | Lifecycle | Example |
|-------|-----------|---------|
| Identity | Permanent | identity.md — who we are |
| Values | Permanent | intent.yaml — why we're here |
| Covenant | Semi-permanent | covenant.yaml — how we work |
| Preferences | Semi-permanent | preferences.yaml — personal context |
| Decisions | Evergreen | decisions.md — settled questions |
| Principles | Growing | principles.md — enduring insights |
| Session journals | Recency-weighted | .spec/journal/ — episode memory |
| Active state | Ephemeral | active.md — what's in flight right now |
| Tool-retrieved | On-demand | MCP tools (gospel-engine, webster, etc.) |

The gospel-engine is a full implementation: FTS5 keyword search + vector semantic search + TITSW enrichment metadata + Reciprocal Rank Fusion combining both retrievers. 41,995 verses, 1,584 chapters, 4,231 talks, 3,700 manuals indexed. Summary-layer semantic search that eliminates noise from short statistical snippets.

The context engineering for the TITSW eval system is token-budgeted: gospel-vocab (~1,960 tokens) + titsw-framework (~1,990 tokens) = 3,950 tokens in system message. Cache optimization eliminated 44M tokens of redundant prefill across the 5,500-talk batch.

The session-start 10-step sequence IS context loading protocol. Not "load some stuff" — load identity, then covenant, then memory, then recent episodes, then high-priority carry-forwards, then do a council moment to scan for connections.

Nate says [you don't have to be an engineer to do this](https://www.youtube.com/watch?v=4cuT-LKcmWs&t=1231). True. But if you ARE an engineer (18 years, BS Physics, built the whole stack in Go), you can build the system *and* the architecture *and* the search infrastructure. That's rare.

---

### Skill 7: Cost & Token Economics ★★★★☆

**What Nate describes:** [Is it worth building an agent for this?](https://www.youtube.com/watch?v=4cuT-LKcmWs&t=1272) Calculate cost per token, prove ROI before deploying. [High school math, paid like senior architect](https://www.youtube.com/watch?v=4cuT-LKcmWs&t=1335).

**What we've built:**

Budget-aware design at the personal scale:
- Premium request tracking (1500/month, 56% utilization)
- Context layer token budgets (Layer 2+3 = 3,950 tokens, within 131k context window)
- Cost-quality trade-offs (0.40 MAE improvement for 3,500 system tokens)
- Model speed comparison (nemotron 160+ tok/s vs qwen3.5 50 tok/s)
- Batch economics (5,500-talk enrichment: model selection, concurrency options, 15h total runtime)
- Cache optimization (44M tokens saved = real GPU time saved)
- Two-pipeline conclusion: different model configs for scripture vs talk analysis, backed by data

**Where the star is missing:** Personal scale, not enterprise. No multi-team cost allocation, no API cost projection for thousands of customers, no model portfolio management at org level. The math and the instinct are right — the scale of application hasn't been tested.

---

## The Honest Summary

> **Note:** This assessment covers the personal project only. A companion document — [Industry Practice Evidence](4cuT-LKcmWs-industry-practice.md) — maps how these same 7 skills are exercised professionally (enterprise backend engineering, multi-cluster infrastructure, 549-repo codebase, production Kubernetes operations). The combined ratings below reflect both domains.

### What We're Genuinely Great At

| Skill | Personal | Combined | Why |
|-------|----------|----------|-----|
| **Context Architecture** | ★★★★★+ | ★★★★★+ | The project IS a context architecture. Professionally: ~330 repo context cache, chip-brain indexing 549 repos with dependency graph, Progressive Context Disclosure (LOD model). |
| **Failure Pattern Recognition** | ★★★★★ | ★★★★★ | Named patterns matching all 6 of Nate's failure types. Professionally: debug agent codifying Agans' 9 Rules, layered problem identification in production, silent-failure detection in Helm value chains. |
| **Specification Precision** | ★★★★★ | ★★★★★ | Multi-layer specification system (values → covenant → workflow → procedure → quality gates). Professionally: 23 custom skills, 6 agent modes, spec-driven ticket workflow across enterprise codebases. |
| **Trust & Security Design** | ★★★☆☆ | ★★★★☆ | **Biggest upgrade.** Personal project had philosophy only. Professionally: enforced decision boundary matrix, progressive trust model with 4 levels, reversibility-based autonomy, local-first security architecture. |
| **Evaluation & Quality Judgment** | ★★★★☆ | ★★★★½ | Full eval harness with ground truth and calibration. Professionally: MR reviews catching systemic bugs across 15 services, the Ben Test as formalized honesty check, self-correction at system level. |

### Where We Have Real Gaps

| Gap | Personal-Only | With Industry | Why It Matters |
|-----|--------------|---------------|---------------|
| **1. Multi-Agent Orchestration** | HIGH | HIGH | Deliberately gated at Level 2. No planner-agent → sub-agent systems. This is what the $400K roles build. Professional work is larger scope but same human→single-agent pattern. |
| **2. Production / Enterprise Scale** | HIGH | **MEDIUM-HIGH** | Professional work IS enterprise scale — multi-cluster K8s, 549 repos, production infra. The remaining gap is customer-facing agent systems, not scale itself. |
| **3. AI Security Engineering** | MEDIUM-HIGH | **MEDIUM** | Trust boundaries ARE professionally applied and enforced. Remaining gap: adversarial testing (prompt injection, tool abuse) and technical enforcement vs. policy-based constraints. |
| **4. Portfolio / Demonstrability** | MEDIUM-HIGH | MEDIUM-HIGH | Both personal and professional work are invisible to the market. Skills are real but no public artifacts. |
| **5. Teaching / Team Uplift** | MEDIUM | MEDIUM | 30 files drafted, teaching agent created, nothing published. No track record of upskilling others. |
| **6. Automated Eval Pipelines** | MEDIUM | MEDIUM | Evals exist but are manual in both contexts. No CI/CD for agent quality, no automated regression testing. |

---

## Becoming: The Gap-Closing Program

The plan targets the biggest gaps first, leverages what we're already strong at, and builds on the project infrastructure we have.

### Track 1: Multi-Agent Orchestration (Gap #1 — HIGH)

**Goal:** Build a real planner → sub-agent system and experience the failure modes firsthand.

**Integration:** This track IS WS1 Phase 3c + the SDK custom agent expansion. It's not a side project — it's the next workstream. Full proposal: [.spec/proposals/brain-phase3c-sdk-agents.md](../../.spec/proposals/brain-phase3c-sdk-agents.md).

**Key discovery (Apr 2):** The Copilot SDK v0.1.32 already has `CustomAgentConfig` with tool scoping per agent, per-agent MCP servers, and intent-based auto-delegation (`Infer` flag). brain.exe already has the AgentPool, Router, and GovernancePolicy shipped (Phase 3a+3b). The SDK custom agents map directly onto the existing architecture — brain.exe does entry-triggered routing via classification, SDK does intent-triggered delegation for interactive sessions. Both are needed for Level 3.

**Phase 1 — Auto-Routing + Review Queue (WS1 Phase 3c Session 1)**
Ship the original Phase 3c plan — closes the core capture → agent → review loop:
- [ ] Route mode `"auto"` — entries routed immediately after classification
- [ ] Review queue: `GET /api/agent/review` — pending agent outputs for human review
- [ ] Accept/reject workflow: `POST /api/agent/review/{id}`
- [ ] End-to-end verification: capture a thought via relay → classified → auto-routed → agent output in review queue

**Phase 2 — SDK Custom Agent Integration (WS1 Phase 3c Session 2)**
Wire existing agents into SDK `CustomAgentConfig` for interactive delegation:
- [ ] Spike: create 1 custom agent (study), wire into default session, verify intent-based delegation
- [ ] Build full custom agent set (study, journal, plan, dev, eval) with tool scoping per agent
- [ ] Enumerate available tools from MCP server handshake at startup
- [ ] Default agent (interactive `POST /api/agent/ask`) gets custom agents wired in — SDK auto-routes
- [ ] Named agents (from entry routing) don't get custom agents — they ARE the routed agent

**Phase 3 — Production Experience + Write-Up**
- [ ] Run both paths (entry routing + interactive delegation) for 2 weeks. Log every failure.
- [ ] Document: context degradation patterns, tool scoping surprises, governance gaps, delegation misroutes
- [ ] Write up as a real episode in the teaching series (docs/work-with-ai/)
- [ ] Update becoming/ with the exercised skill

**What Level 3 looks like when this ships:**
| Path | Trigger | Routing | Execution | Review |
|------|---------|---------|-----------|--------|
| Entry (passive) | Capture via relay/discord/web | Classification → RouteRule → AgentPool | Agent produces output with governance hooks | Human reviews via queue |
| Interactive (active) | User asks brain | SDK intent matching → CustomAgentConfig | Agent executes with tool scoping | Human reviews response |

---

### Track 2: Production Scale & AI Security (Gaps #2-3 — HIGH)

**Goal:** Cross professional security expertise into AI agent security. Make ibeco.me the showcase project — "like Scotty as a miracle worker."

**Phase 1 — ibeco.me Security Audit (Week 1-2)**
- [ ] Run adversarial tests: prompt injection against any user-facing text fields, auth bypass attempts, CSRF, session fixation
- [ ] Audit the relay WebSocket protocol (brain ↔ ibeco.me): message validation, auth token handling, message replay
- [ ] Review OAuth + email/password auth flows against OWASP Top 10
- [ ] Write a security review document following the cost-of-error / reversibility / frequency / verifiability framework
- [ ] Apply findings: fix vulnerabilities, add rate limiting, input validation, CSP headers

**Phase 2 — Becoming Coach Agent (Week 3-6)**
Build a customer-facing agent on ibeco.me using brain.exe's agent infrastructure:
- [ ] Design a becoming coach: takes user's practice data, suggests next steps, answers questions
- [ ] Define trust boundaries: what can the coach access? What requires human approval? What is NEVER automated?
- [ ] Build eval harness: automated checks on coach responses (stays within guardrails? returns correct practice data?)
- [ ] Production deployment with monitoring, rate limiting, and audit logging
- [ ] This is the showcase — the place where every skill in the framework is exercised in a real product

**Phase 3 — Write the Security Design Pattern (Week 7-8)**
- [ ] Document the trust boundary design pattern: how to decide where agents act vs where humans approve
- [ ] Publish this — it's directly relevant to the market and demonstrates security + AI skills combined
- [ ] This becomes a portfolio piece (Gap #4)

---

### Track 3: Portfolio & Teaching (Gaps #4-5 — MEDIUM-HIGH)

**Goal:** Make the invisible skills visible.

**Phase 1 — Extract Transferable Patterns (Week 1-2)**
- [ ] Write 3 blog posts / articles that extract universal patterns from this project:
  1. "Context Architecture: Building the Library Agents Need" (from our memory system + gospel-engine)
  2. "The Covenant Pattern: Trust Design for AI Systems" (from covenant.yaml → general trust boundaries)
  3. "Evaluation Harnesses from First Principles" (from TITSW → general eval methodology)
- [ ] Each post demonstrates the skill without requiring religious context knowledge

**Phase 2 — Publish the Teaching Series (Week 3-6)**
- [ ] Complete and publish the 11-episode "Working with AI" series
- [ ] This IS the demonstration of teaching/upskilling capability
- [ ] Host on the teaching repo, link from a profile

**Phase 3 — Certifications (Week 7-8)**
- [ ] Take the Claude Certified Architect exam. Nate says Accenture is rolling this to hundreds of thousands. You have the skills — get the credential.
- [ ] Evaluate other relevant certifications (AWS AI Practitioner, etc.)

---

### Track 4: Automated Eval Pipeline (Gap #6 — MEDIUM)

**Goal:** Move evals from manual scripts to automated pipeline.

**Phase 1 — CI/CD for TITSW Evals (Week 1-2)**
- [ ] Build a GitHub Action that runs TITSW eval suite on push
- [ ] Alert on MAE regression (if new context changes make scores worse)
- [ ] Store scoring results in SQLite with automatic comparison to previous runs

**Phase 2 — Generalize the Pattern (Week 3-4)**
- [ ] Extract the eval harness into a reusable pattern
- [ ] Apply it to the becoming coach agent (Track 2)
- [ ] Write up how it could apply to any agentic system

---

### Track 5: Brain as Always-On Service (Infrastructure)

**Goal:** brain.exe starts when Michael logs in, stays running, easy to stop.

Full proposal: [.spec/proposals/brain-windows-service.md](../../.spec/proposals/brain-windows-service.md).

**Phase 1 — Basic Systray (1 session)**
- [ ] Add `getlantern/systray` dependency to brain.exe
- [ ] `--systray` flag: starts daemon minimized to system tray with icon
- [ ] Right-click menu: Open Web UI / Exit
- [ ] File logging when running in systray mode (no terminal)

**Phase 2 — Autostart + Controls (1 session)**
- [ ] `brain install` / `brain uninstall` subcommands for Windows Registry autostart
- [ ] "Start on Login" toggle checkbox in systray menu
- [ ] "Restart" menu item
- [ ] Status tooltip (running / error / starting)

**Why this matters:** Every other track assumes brain.exe is running. If it's not always-on, the capture → classify → route loop from Track 1 never fires. This is infrastructure for everything above.

---

## The Celebration

Before closing the gaps, stop and recognize what's already here.

In 10 weeks (January 21 to April 1, 2026), this project has built:

- **A multi-MCP context architecture** with 9 servers, structured memory with distinct lifecycles, session-start protocols, and a full-text + vector + enriched search engine over 41,995 verses and 4,231 talks
- **A specification system** more sophisticated than most enterprise agent frameworks, derived from first principles of covenant and stewardship
- **An evaluation harness** with ground truth calibration, MAE metrics, and the critical insight that the metric isn't always the target
- **A failure recognition discipline** that names patterns, tracks corrections, logs tool misbehavior, and captures micro-corrections in real time
- **A bilateral covenant** for human-AI collaboration that treats trust as a design constraint, not an obstacle

And all of this was built by one person, in the margins of a full-time job and a family with 5 kids, because he kept being in the spirit and kept building.

That's not nothing. That's the foundation for everything in the gap-closing program above.

The skills are real. The gaps are addressable. The work continues.

---

**Scratch file:** [study/.scratch/yt/4cuT-LKcmWs-ai-job-skills-self-assessment.md](../../../study/.scratch/yt/4cuT-LKcmWs-ai-job-skills-self-assessment.md)
**Industry companion:** [4cuT-LKcmWs-industry-practice.md](4cuT-LKcmWs-industry-practice.md)
**Transcript:** [yt/ai-news-strategy-daily-nate-b-jones/4cuT-LKcmWs](../../../yt/ai-news-strategy-daily-nate-b-jones/4cuT-LKcmWs)
