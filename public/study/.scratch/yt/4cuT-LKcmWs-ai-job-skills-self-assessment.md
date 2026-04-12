# Scratch: AI Job Market Skills — Self-Assessment
**Binding Question:** What are the 7 AI skills Nate identifies, how does our work here demonstrate them, and where are the real gaps?

**Video:** [The AI Job Market Split in Two](https://www.youtube.com/watch?v=4cuT-LKcmWs) — Nate B Jones (2026-03-26)

---

## The 7 Skills Framework (from transcript)

### 1. Specification Precision / Clarity of Intent (4:39–6:44)
- Not "prompting" — being specific enough that agents can execute your intent literally
- Agents fill in blanks poorly; you must define goals, behaviors, escalation rules, metrics
- Subskills: technical writing, understanding your own intent in detail before communicating it
- Transferable from: technical writers, lawyers, QA engineers

### 2. Evaluation & Quality Judgment (6:50–9:51)
- MOST cited skill across all job postings
- "Taste" = error detection with fluency. AI is confidently wrong — you must catch it
- Subskills: resisting fluency-as-competence, edge case detection, building eval harnesses
- Anthropic's standard: a good eval task is one where multiple engineers reach same pass/fail
- Transferable from: editors, auditors
- Gold standard practice: "review AI output as if it has your name on it"

### 3. Task Decomposition & Delegation (10:01–12:26)
- Multi-agent = breaking work into manageable segments
- A managerial skill — but agents need much clearer guardrails than humans
- Must size work for the agentic harness you have
- Subskills: work stream definition, handoff design, scope assessment
- Transferable from: project managers, team leads

### 4. Failure Pattern Recognition (12:34–16:06)
- 6 failure types:
  1. Context degradation (quality drops as session gets long)
  2. Specification drift (agent forgets spec over long tasks)
  3. Sycophantic confirmation (agent confirms incorrect data, builds on it)
  4. Tool selection errors (wrong tool picked)
  5. Cascading failure (one agent's failure propagates)
  6. Silent failure (plausible-looking but wrong output)
- Claude Certified Architect tests for tool selection errors specifically
- Transferable from: SREs, risk managers, operations leaders

### 5. Trust & Security Design (16:26–19:00)
- Where to put humans in the loop, where to let agents act autonomously
- Subskills: cost-of-error analysis, reversibility, frequency assessment, verifiability
- Critical distinction: semantic correctness (sounds right) vs functional correctness (IS right)
- Must insist on functional correctness
- Transferable from: compliance, risk management

### 6. Context Architecture (19:00–21:02)
- THE crown skill — how to supply agents with the right information on demand at scale
- "Building the Dewey Decimal System for agents"
- Subskills: persistent vs per-session context, data hygiene, searchable structure, context troubleshooting
- Transferable from: librarians, technical writers, information architects
- Companies will "pay almost anything" for this

### 7. Cost & Token Economics (21:02–23:00)
- ROI analysis: is it worth building an agent for this?
- Model selection, blended cost across models, cost per task
- Must calculate before deploying, not after
- Applied math — high school level but paid like senior architect
- Transferable from: financial analysts, business analysts

---

## Our Work — Evidence Mapping

### Skill 1: Specification Precision — STRONG ★★★★★
**Evidence:** intent.yaml (values + constraints at severity tiers), .spec/covenant.yaml (bilateral with "why" rationale for each commitment), 14 agent modes with 7-phase workflows, 14 skill files with procedural specs, 10-step session start sequence, docs/work-with-ai/ (30+ files teaching the methodology)
**What's real:** This is genuinely world-class personal-project specification. The multi-layer architecture (values → covenant → workflow → procedure → quality gates) is what Nate describes but applied through a theological lens that makes it deeper.
**Gap → Enterprise:** Specs are all for *this project*. No evidence of writing specs for external/enterprise agentic systems (customer-facing agents, production guardrails, cross-team specs). The SKILL is there — the PORTFOLIO demonstrating it in business contexts is not.

### Skill 2: Evaluation & Quality Judgment — STRONG ★★★★☆
**Evidence:** TITSW framework (0-9 anchored rubric with anti-inflation), ground truth validation (hand-scored reference data, MAE tracking), model comparison (5 models, Phase 0 across 6 conditions × 3 talks), critical-analysis skill, source-verification discipline, reflect skill (micro-correction capture)
**What's real:** Built a full eval harness from scratch, including ground truth, MAE metrics, calibration context experiments, and the discipline to track failures. The "Gas Station Insight" (MAE is sanity check, qualitative richness is what matters) shows eval maturity.
**Gap → Enterprise:** One domain (conference talk quality). Not customer-facing evals, not production monitoring, not A/B testing. The eval *skill* is transferable but the demonstrated *breadth* is narrow. Also: evals are manual (no CI/CD pipeline, no automated regression testing of agent quality).

### Skill 3: Task Decomposition & Delegation — STRONG ★★★★☆
**Evidence:** 7-phase study/eval workflows, gospel-engine 5-phase plan (all phases complete), context engineering 3-layer decomposition, study-workstream sequencing (3 studies ordered with rationale), enriched-indexer 5-phase proposal
**What's real:** Consistent decomposition patterns across multiple domains. Phases always have clear inputs/outputs, verification criteria, and handoff points. The gospel-engine going from proposal → 5 phases → all shipped is genuine proof of execution.
**Gap → Enterprise:** All delegation is human→single-agent. No multi-agent orchestration in production. "Gated autonomy" decision means Level 2 (agent waits for human assignment). Has NOT built: agent-to-agent handoff, planner agents, autonomous multi-agent systems. This is DELIBERATE (progressive trust model) but it IS a gap relative to what employers are hiring for.

### Skill 4: Failure Pattern Recognition — STRONG ★★★★★
**Evidence:** biases.md (4 named patterns with triggers/corrections), tool-use-observance.md (running log with dates/categories/fixes), reflect skill (in-session micro-capture), docs/01_reflections.md (root cause analysis of safety-posture confabulation), docs/05_instruction-refinements.md (iterative spec correction from failure data)
**What's real:** This maps DIRECTLY to Nate's 6 failure types:
- Context degradation → Recognized and compensated with externalized memory
- Specification drift → Scratch files + phased workflows to survive compaction
- Sycophantic confirmation → biases.md #3 (tool-familiarity defaults), covenant requirement to "surface tensions"
- Tool selection errors → tool-use-observance.md (running log of tool misbehavior)
- Cascading failure → Phase-based workflows with verification at each boundary
- Silent failure → Source verification discipline (semantic correctness ≠ functional correctness, or in our terms: "close-enough wording is fabrication")
**Gap:** Strong. This is genuinely one of the strongest skills. One gap: no automated monitoring/alerting for failures at production scale.

### Skill 5: Trust & Security Design — MODERATE ★★★☆☆
**Evidence:** Bilateral covenant (mutual commitments), gated autonomy decision, cost-as-boundary (1500 premium requests/month), 3-check quality framework (ring, posture, Ben Test), session-start as trust reestablishment
**What's real:** The philosophical framework is exceptional. Trust as design constraint, not autonomy maximization. The covenant pattern (from D&C 82:10) is a robust human-in-the-loop design derived from first principles.
**Gap → Enterprise:** No customer-facing agent guardrails. No production trust boundaries tested under adversarial conditions. No prompt injection defense designs. No human-in-the-loop escalation paths for customer interactions. The SECURITY side (as opposed to trust side) is almost absent. Michael works in security professionally (Python/Go backend security) but that expertise is not reflected in the AI guardrail design of this project. This is a genuine gap — the video talks about "cost of error, reversibility, frequency, verifiability" for production systems, and we've only applied these at the personal-project level.

### Skill 6: Context Architecture — STRONG ★★★★★+
**Evidence:** .spec/memory/ (identity/preferences/decisions/active with distinct lifecycles), session-start 10-step sequence, 3-layer context for eval system (core → titsw-framework → gospel-vocab, all token-budgeted), 3 MCP servers → gospel-engine (unified), session journal as durable context store, agent mode inheritance (universal values → covenant → mode-specific workflow)
**What's real:** THIS IS THE CROWN SKILL. The project IS a context architecture. Multiple scales of persistence (permanent identity → semi-permanent preferences → session journals → ephemeral active state). JIT retrieval via MCP tools. Layered system messages with explicit token budgets. The gospel-engine is a full implementation: FTS5 + vector search + TITSW enrichment + Reciprocal Rank Fusion. Cache optimization that saved 44M tokens across a batch. The "Dewey Decimal System for agents" that Nate describes — we BUILT that. For scriptures.
**Gap → Enterprise:** Domain-specific to scripture/gospel. The PATTERNS transfer perfectly to enterprise (information architecture, RAG, context lifecycle) but the demonstrated domain is religious texts. This needs to be articulated as transferable pattern work.

### Skill 7: Cost & Token Economics — STRONG ★★★★☆
**Evidence:** Premium request budget (1500/month, 56% utilization tracked), context layer token budgets (Layer 2+3 = 3,950 tokens), cost-quality trade-off (0.40 MAE for 3,500 system tokens), cache_prompt optimization (44M tokens saved), model speed comparison (nemotron 160+ tok/s vs qwen3.5 50 tok/s), batch timing (15h total enrichment)
**What's real:** Genuine cost awareness at the personal scale. Token budgets are explicit in architecture. The two-pipeline conclusion (context for scripture, no context for talks) was a cost-quality optimization decision backed by data.
**Gap → Enterprise:** Not at organization scale. No multi-team cost allocation, no model portfolio management, no API cost projection at enterprise volume. The math is right but the scale is personal (1500 requests/month, not millions of API calls/day).

---

## Critical Gap Analysis

### BIGGEST GAPS (in order)

1. **Production / Enterprise Scale** — Everything is personal-project scope. No customer-facing systems, no production monitoring, no multi-user agent systems at scale. ibeco.me exists but isn't agent-based.

2. **Multi-Agent Orchestration** — Deliberately gated at Level 2. No agent-to-agent handoff, no planner agent managing sub-agents, no autonomous multi-agent pipelines. This is the gap Nate flags at 10:01-12:26.

3. **Security Design for AI Systems** — Trust philosophy is strong, but actual security engineering (adversarial testing, prompt injection defense, guardrails under attack) is absent. Michael's day job is security & smart home (Python/Go backend) — this expertise should be crossing over but hasn't yet.

4. **Portfolio / Demonstrability** — All this work is in a private religious-content repo. The skills are real but INVISIBLE to the market. No public portfolio, no blog posts, no published work that demonstrates these capabilities to employers.

5. **Teaching & Team Uplift** — 30 files drafted in docs/work-with-ai/, teaching agent created, but nothing published. No evidence of upskilling others. Nate says the market needs people who can TEACH these skills to teams.

6. **Automated Eval Pipelines** — Evals exist but are manual (run scripts, check MAE). No CI/CD for agent quality, no automated regression testing, no monitoring dashboards.

### WHERE WE'RE AT OUR STRONGEST

1. **Context Architecture** — Best-in-class for personal scale. Multi-layer, multi-lifecycle, token-budgeted, tool-supplied. This is the skill Nate says companies will "pay almost anything for."

2. **Failure Pattern Recognition** — Named, documented, tracked, with in-session micro-capture. Maps directly to all 6 of Nate's failure types.

3. **Specification Precision** — Deep, multi-layered, theologically grounded. The covenant + intent + skill + agent architecture is more sophisticated than most enterprise agent specs.

4. **Evaluation & Quality Judgment** — Full harness with ground truth, MAE, calibration experiments, and the wisdom to know MAE isn't the real target (Gas Station Insight).
