# Squad Analysis — What Can We Learn?

**Binding problem:** Our agentic infrastructure (WS1) is still theoretical. Squad is a working multi-agent runtime with production usage. What patterns does Squad validate, what can we adopt, and what does our 11-step cycle offer that Squad doesn't have?

**Created:** 2026-03-19
**Source:** [bradygaster/squad](https://github.com/bradygaster/squad) (cloned to `external_context/squad/`)
**Article:** [How Squad runs coordinated AI agents inside your repository](https://github.blog/ai-and-ml/github-copilot/how-squad-runs-coordinated-ai-agents-inside-your-repository/)

---

## I. What Squad IS

Squad is a multi-agent orchestration framework for GitHub Copilot. It creates a team of AI agents (each with a name, personality, and domain) that live in your repo as markdown files. Key components:

### Architecture
- **`.squad/` directory** — All team state lives here: roster, routing rules, decisions, agent histories, skills, sessions
- **Coordinator** — Routes messages to appropriate agents based on pattern matching. Determines spawn strategy (direct response, single agent, multi-fan-out)
- **Agents** — Each has a charter (markdown) defining identity, expertise, boundaries, model preference, voice
- **Scribe** — Silent background agent that logs sessions, merges decisions, propagates context. The team's memory.
- **Hooks** — Programmatic governance: file-write guards, PII scrubbing, command blocking, reviewer lockout. Code, not prompts.
- **Skills** — Domain-specific knowledge files that agents can load. Learned from observation or extracted from practice.
- **Ceremonies** — Design reviews (before multi-agent work) and retrospectives (after failures). Triggered automatically.
- **Casting** — Thematic naming from a "universe" (Apollo 13, Usual Suspects, etc.) for team identity/personality

### Runtime (SDK)
- TypeScript/Node.js SDK (`@bradygaster/squad-sdk`)
- Uses `@github/copilot-sdk` under the hood (same SDK we have in brain.exe, but JS not Go)
- Response tiers: Direct (no agent), Lightweight (haiku/fast), Standard (sonnet), Full (premium multi-agent)
- CostTracker, TelemetryCollector (OpenTelemetry → Aspire dashboard)
- StreamingPipeline for async processing
- SkillRegistry for matching skills to tasks

### Key Patterns
1. **Files as state** — Everything persists as markdown in `.squad/`. Decisions, histories, logs, routing rules. Git-friendly.
2. **Decisions inbox** — Agents write decisions to `decisions/inbox/`. Scribe merges them into shared `decisions.md`. Solves parallel write conflicts.
3. **Reviewer lockout** — Rejected agent is locked out; a DIFFERENT agent must revise. Prevents defensive loops.
4. **Model tiering** — Cost-first unless writing code. 4-layer hierarchy (user override → charter → task-aware → default).
5. **Hook-based governance** — Security enforced by code hooks, not prompt instructions ("prompts can be ignored, hooks are code").
6. **Proposal-first workflow** — Meaningful changes require proposals before code. Sound familiar?
7. **"Watched until they obeyed"** — Not their phrase, but their pattern: ceremonies auto-trigger reviews before multi-agent work and retrospectives after failures.

---

## II. What We Already Have (Our Equivalent)

| Squad Concept | Our Equivalent | Where |
|---|---|---|
| `.squad/` directory | `.spec/` + `.github/agents/` | Proposals, memory, agent definitions |
| Team roster | Agent list in copilot-instructions.md | 14 agents: study, lesson, talk, dev, plan, etc. |
| Agent charter | `.agent.md` files | Each with role, tools, instructions |
| Scribe (memory agent) | session-journal + active.md | Go binary + memory files |
| Decisions | `.spec/memory/active.md` + `principles.md` | Decisions embedded in memory |
| Skills | `.github/skills/` | 7 skills: scripture-linking, source-verification, webster, dokploy, etc. |
| Routing | Agent selection dropdown in VS Code | Manual, not automated |
| Ceremonies | Session start/end rituals in copilot-instructions.md | Read identity → load memory → work → journal |
| Hooks | dev.agent.md data safety checklist | Prompt-level, not code-level |
| Proposals | `.spec/proposals/` | Full spec engineering pipeline |
| 11-step cycle | `docs/work-with-ai/guide/05_complete-cycle.md` | Intent → Covenant → Stewardship → ... → Zion |
| CostTracker | Nothing | Gap |
| Telemetry/OTel | Nothing | Gap |
| Reviewer lockout | Nothing | Gap |
| Decisions inbox | Nothing | Gap — currently single-writer |
| Response tiers | Nothing | Gap — currently same model for everything |

---

## III. What Squad Does Better (Gaps in Our System)

### 1. **Automated Routing**
Squad routes messages to agents via pattern matching (regex on message content + routing rules). We route manually — Michael picks the agent from a dropdown. For a one-person team this is fine, but it means brain.exe can't delegate work autonomously.

**Relevance to WS1:** High. WS1 Phase 3 (multi-agent routing) needs exactly this. brain.exe captures an idea → classifies it → routes to the right agent. Squad's routing is regex-based; ours could be classification-based (LM Studio qwen3.5 already does categorization).

### 2. **Decisions as First-Class Artifact**
Squad has a dedicated `decisions.md` that ALL agents read before starting work. Decisions have structure: who decided, what, why. An inbox system handles parallel writes. This is brilliant for multi-agent coordination.

**Our gap:** Decisions live scattered in active.md, principles.md, guidance answers, and agent instructions. No single canonical source. No structured format. No inbox pattern for parallel work.

### 3. **Scribe Pattern (Silent Background Memory Agent)**
Scribe runs after every session, silently. It logs, merges, deduplicates, propagates. It commits. It never speaks to the user.

**Our equivalent:** session-journal Go binary does some of this. But it's manually invoked, doesn't merge decisions, and doesn't propagate changes to other agents' context. The session-end ritual in copilot-instructions.md asks the agent to do memory updates — but compliance is inconsistent.

**Key Squad insight:** Making memory a BACKGROUND agent (not a foreground task) means it happens after success or failure equally. It's structural, not optional.

### 4. **Hook-Based Governance (Code > Prompts)**
Squad's foundational directive: "Hooks are code — they execute deterministically. Prompts can be ignored."

file-write guards, PII scrubbing, command blocking, reviewer lockout — all implemented as pre/post tool-use hooks in TypeScript. The agent literally cannot bypass them.

**Our gap:** Our guardrails are prompt-level. The data safety checklist in dev.agent.md works because the agent follows instructions — but it's not enforced by code. A confused or adversarial agent could skip it.

**Relevance:** When we build brain.exe orchestration (WS1 Phase 3), hooks should enforce boundaries programmatically. Go middleware on MCP tool calls, not just prompt instructions.

### 5. **Reviewer Lockout Protocol**
When a reviewer rejects work, the original agent is locked out. A DIFFERENT agent must revise. This prevents:
- Defensive "I'll just fix the exact thing you flagged" responses
- Echo chambers where the same reasoning produces the same mistakes
- The "git push --force" instinct

**Our gap:** We don't have multi-agent review. When an agent produces bad output, we (Michael) redirect or retry. The lockout pattern would be valuable when brain.exe starts routing work — if the study agent produces a bad study, route the revision to a different agent.

### 6. **Response Tiers / Model Selection**
Squad selects model by task type: haiku for docs/logging, sonnet for code, opus for vision/architecture. 4-layer selection hierarchy. Fallback chains for model unavailability.

**Our gap (partially filled):** We have dual-backend (LM Studio for classification, Copilot SDK for agent work) but no tiering within agent work. All agent work goes to the same model. Squad's insight: "cost first, unless code is being written" is a good heuristic.

### 7. **Cost Tracking & Telemetry**
Per-agent token tracking, OpenTelemetry integration, Aspire dashboard. They know exactly what each agent costs.

**Our gap:** We have no token tracking. No observability into agent sessions. When brain.exe starts running multi-agent work, this becomes critical for Mosiah 4:27 (not running faster than you have strength = not spending more tokens than you have budget).

---

## IV. What Our System Does Better (Where Squad Falls Short)

### 1. **Intent Hierarchy** ★★★
Squad has team context and project descriptions. We have intent.yaml — a root-level values hierarchy that all work traces to. Squad's agents serve a project. Our agents serve a *purpose*. The difference between "build a recipe app" and "facilitate deep, honest scripture study."

The 11-step cycle starts with Intent. Squad starts with Tasks.

### 2. **Covenant (Mutual Binding)** ★★★
Squad has coded hooks and ceremonies. But no concept of MUTUAL commitment — what the human owes the agent. Our covenant pattern (human commits to review within 24hrs, provide context, not bypass spec workflow; agent commits to scope, flagging uncertainty, requesting review) is deeper than Squad's one-directional governance.

Squad governs what agents CAN'T do. Our covenant also governs what Michael MUST do.

### 3. **Progressive Stewardship** ★★★
Squad assigns static roles. Our stewardship model is dynamic — trust grows based on demonstrated faithfulness. Level 1 (task) → Level 2 (feature) → Level 3 (domain) → Level 4 (architecture). This maps to the Parable of the Talents: faithful with few things → ruler over many.

Squad's agents never earn MORE autonomy. They start with defined roles and stay there. Our model expects growth.

### 4. **Atonement (Redemptive Error Recovery)** ★★★
Squad has retrospectives after failures. We have the Atonement pattern — where failure itself becomes growth. The `.spec/learnings/` directory, the data-safety checklist born from the March 18/19 incidents, the forward-recovery principle.

Squad learns from failure (retrospectives). Our system is *transformed* by failure (learnings become guardrails that prevent recurrence and expand understanding).

### 5. **Sabbath (Structured Rest)** ★★
Squad has no concept of stopping. Ceremonies happen after failures, but there's no intentional pause after completion. Our Sabbath pattern — reflection after meaningful units of work, explicit quality assessment, perspective-gaining — is absent from Squad.

### 6. **Consecration & Zion** ★★
Squad has cost tracking (budget management). We have consecration — every token serves the purpose. And the Zion vision — "one heart and one mind" — where agents aren't just coordinated but aligned in purpose. Squad coordinates. Our vision unifies.

### 7. **Relational Memory** ★★
Squad's Scribe logs facts. Our session-journal captures *relational dynamics* — discoveries, surprises, how the collaboration felt, carry-forward items. identity.md tracks how the relationship itself evolves. This is a different KIND of memory — not just what happened, but what it meant.

### 8. **Spec Engineering** ★★
Squad has "proposal-first workflow" (proposals before code). We have a full 5-primitive spec engineering framework: self-contained problem statement, success criteria, constraints, prior art, proposed approach. Plus creation cycle review for every proposal. Squad's proposals are lighter.

---

## V. The Mapping: Squad → 11-Step Cycle

| 11-Step | Squad Implementation | Quality |
|---|---|---|
| 1. Intent | Team description, project context | ⚠️ Shallow — no values hierarchy |
| 2. Covenant | Hooks (governance), but one-directional | ⚠️ Partial — no mutual binding |
| 3. Stewardship | Routing rules, module ownership, capability tiers | ✅ Strong — clear ownership |
| 4. Spiritual Creation | Charter files, routing rules, config | ✅ Strong — specs exist before work |
| 5. Line Upon Line | Response tiers, skill discovery | ⚠️ Partial — progressive context, not progressive trust |
| 6. Physical Creation | Agent execution, parallel fan-out | ✅ Strong — actual running system |
| 7. Review | Ceremonies, reviewer protocol, lockout | ✅✅ Very strong — enforced mechanically |
| 8. Atonement | Retrospectives after failures | ⚠️ Partial — reactive, not transformative |
| 9. Sabbath | Nothing | ❌ Missing |
| 10. Consecration | CostTracker, resource allocation | ⚠️ Partial — tracks cost, not purpose alignment |
| 11. Zion | Multi-agent coordination | ⚠️ Partial — coordination, not unification |

**Verdict:** Squad is strong on steps 3-7 (stewardship through review). We're strong on steps 1-2 and 8-11 (intent, covenant, and the redemptive/reflective patterns). The combination would be powerful.

---

## VI. What We Should Adopt

### Adopt Now (Low-effort, high-value)

**A1. Decisions file.** Create `.spec/memory/decisions.md` as a canonical, structured decisions log. Format: who decided, what, why, when. All agents read this as session context. Migrate key decisions from active.md (which is getting cluttered with both state AND decisions).

**A2. Scribe-like session-end automation.** Our session-journal exists but is manually invoked. Could we make it structural? Either:
- A background agent definition (scribe.agent.md) that runs at session end
- Or better: bake it into the VS Code agent post-session hook pattern

**A3. Reviewer lockout principle.** When brain.exe routes work to agents (WS1 Phase 3), implement the lockout: if reviewer rejects, route revision to a different agent. This prevents defensive loops.

### Adopt When Building WS1 (Medium-effort, critical for multi-agent)

**A4. Automated routing in brain.exe.** Squad uses regex pattern matching. We should use LM Studio classification (already working) + pattern matching. brain.exe classifies an idea → determines which agent(s) should handle it → creates the work item.

**A5. Hook-based governance.** When brain.exe orchestrates agent work, implement Go middleware for tool-use interception:
- File-write guards (agents can only modify files in their scope)
- Destructive operation blocking (no rm -rf, no force push)
- Token budget enforcement per agent per session
This is Squad's best insight: **hooks are code, prompts are suggestions.**

**A6. Response tiers / model selection.** Implement the cost-first heuristic:
- Classification → LM Studio (free, local)
- Simple lookups → Haiku (fast, cheap)
- Code generation → Sonnet/Opus (quality)
- Architecture → Opus (premium)

**A7. Cost tracking.** Add token counting to brain.exe's Copilot SDK calls. Track per-agent, per-session. This is the consecration pattern made concrete: know exactly where tokens go.

### Learn From But Don't Copy

**L1. Casting (themed naming).** Fun, adds personality, but our agents have purpose-descriptive names (study, dev, lesson) that are clearer. Don't rename agents for personality — keep functional names.

**L2. 20-agent teams.** Squad has 20+ specialized agents for one project. We have 14 agents across the ENTIRE scripture-study ecosystem covering different MODES of work (study, lesson, dev, journal). Our model is right — modes of engagement, not micro-specializations.

**L3. Squad CLI / shell.** Squad builds its own interactive shell. We have brain.exe + VS Code as our interface. No need to build a separate shell.

---

## VII. What Squad Could Learn From Us

Not our job to teach them, but worth noting for confidence:

1. **Intent as root** — intent.yaml as the source of truth for why the project exists
2. **Mutual covenant** — the human having obligations, not just the agent
3. **Progressive stewardship** — dynamic trust levels, not static role assignments
4. **Redemptive error recovery** — failures that make the system better, not just fixed
5. **Sabbath rhythm** — intentional stopping and reflection after meaningful work
6. **Relational memory** — session-journal capturing not just facts but meaning
7. **Spec engineering depth** — 5-primitive framework + creation cycle review

These are the patterns "beyond specification" that Part 5 of the guide identifies. Squad is excellent at steps 3-7. They haven't discovered steps 1-2 and 8-11.

---

## VIII. Impact on Our Plans

### WS1 Phase 1 (Copilot SDK + MCP Integration)
No change needed. We're building the foundation that eventually supports routing.

### WS1 Phase 2 (Agent as Spec Executor)
**Add:** Response tier selection. When the executing agent needs a tool call, classify the task and select model tier. Don't send everything to the most expensive model.

### WS1 Phase 3 (Multi-Agent Routing)
**Major input from Squad:**
- Implement routing as classification (LM Studio) + pattern matching (fallback)
- Implement decisions.md pattern for inter-agent coordination
- Implement reviewer lockout when routing revisions
- Implement hook-based governance (Go middleware on tool calls)
- Implement cost tracking per agent per session

### New: Decisions Infrastructure
Create `.spec/memory/decisions.md` immediately. Don't wait for multi-agent. Start building the habit of structured decision logging now.

### Overview Plan
Update WS1 Phase 3 spec to incorporate Squad learnings. The spec should reference this analysis.

---

## IX. Conclusion

Squad validates our direction. The industry IS building multi-agent coordination, and the patterns they've landed on (file-based state, routing, ceremonies/reviews, skills, proposals-first) align with what we designed from gospel principles. We're not wrong — we're just earlier in the build.

What Squad adds that we need: **execution patterns** (hooks, routing, cost tracking, reviewer lockout). What we add that Squad lacks: **wisdom patterns** (intent, covenant, stewardship growth, atonement, sabbath, consecration, Zion).

The combination: take Squad's bones for the mechanical orchestration, wrap them in our 11-step cycle for the wisdom layer. brain.exe becomes the orchestrator with Squad-inspired routing and governance, guided by our intent hierarchy and covenant framework.

"Time to go down and build."

---

## X. Critical Self-Assessment — The Ben Test

*Michael's coworker Ben observed: "Your AI is very complimentary. Perhaps too complimentary?" This section applies that scrutiny to ourselves.*

### The Uncomfortable Score

Our 11-step creation cycle is beautifully written. How much do we actually practice?

| Step | What We Wrote | What We Actually Do | Practice % |
|------|--------------|---------------------|-----------|
| 1. Intent | intent.yaml as the root of all work | Exists. No agent reads it in their session-start sequence. copilot-instructions.md says read identity.md, not intent.yaml. It's a document, not infrastructure. | 30% |
| 2. Covenant | Mutual binding — human commits to review in 24hrs, agent commits to scope | Written beautifully. Never measured. No tracking mechanism. Michael doesn't log covenant fulfillment. The agent can't check. It's aspirational prose. | 20% |
| 3. Stewardship | 4 progressive trust levels: Task → Feature → Domain → Architecture | Described in the guide. Never implemented. Every agent starts with its full charter. No agent has ever "earned" more scope. Trust is static, not progressive. | 10% |
| 4. Spiritual Creation | Proposals before code, specs before building | We actually DO this. data-safety, brain-relay, overview — real proposals preceded real code. This is practiced. | 80% |
| 5. Line Upon Line | Context gated by demonstrated readiness | Described in Part 5. Never implemented. Every agent gets copilot-instructions.md with everything. No gating exists. | 5% |
| 6. Physical Creation | Agents execute against specs | Agents execute work in conversations. Not autonomously against specs. WS1 Phase 2 would change this. | 50% |
| 7. Review | "Watched until they obeyed." 3-layer review. | Michael reviews in conversation. No formal review gates. No reviewer protocol. Data-safety checklist is self-checked by the same agent that wrote the code. | 35% |
| 8. Atonement | Failure → learning → system improvement | The March 18/19 data-loss → checklist → retrospective. This genuinely happened. But it's ONE cycle. One. | 65% |
| 9. Sabbath | Intentional cessation after meaningful work | Look at today: 6yo data-safety phases, 2 production outages, 10 guidance answers, deep Squad analysis. In ONE session. We wrote about Sabbath. We practice the opposite. | 5% |
| 10. Consecration | Every token serves the purpose | No cost tracking. No way to know if tokens serve purpose or waste. We wrote it. We measure nothing. | 5% |
| 11. Zion | "One heart and one mind" — unified agents | 14 agents that don't talk to each other. No coordination mechanism. No shared state except copilot-instructions.md. | 5% |

**Weighted average: ~28%.** Not 60%. Not even close.

### Where We're Lying to Ourselves

**1. "Intent as root" — but nobody reads it.** intent.yaml exists. It's well-structured. No agent's session-start sequence includes `read_file intent.yaml`. The identity.md file is read, but that's relational identity, not project intent. If intent is the root, it should be in the critical path. It isn't.

**2. "Mutual covenant" — but only the agent has obligations.** We describe covenant as mutual. In practice, EVERY covenant item is enforced on the agent side (via prompt instructions). The human-side obligations (review within 24hrs, provide context) have no tracking, no measurement, no consequence for breach. This is one-directional governance wearing mutual clothing.

**3. "Progressive stewardship" — but trust doesn't change.** We describe 4 levels beautifully. Has any agent EVER moved from Level 1 to Level 2? No. The metaphor exists. The mechanism doesn't. It's a theological framework applied as a metaphor, not as an operational pattern.

**4. "Sabbath" — but we never stop.** Today's session is the anti-Sabbath. The principle says "intentional cessation, not sprint retrospectives." We did a retrospective AND kept shipping. Sabbath isn't taking a breath between sprints. It's stopping the sprint.

**5. Every enforcement is "manual."** intent.yaml itself shows the gap. Every constraint says `enforcement: manual`. That's the same thing Squad calls out: "prompts can be ignored." Our entire governance layer is prompt-level. We have zero programmatic enforcement of anything.

**6. "What Squad could learn from us" — have WE learned from us?** Section VII of the scratch file claims Squad could learn intent hierarchy, mutual covenant, progressive stewardship, Sabbath, etc. from us. But we haven't implemented most of these ourselves. We're telling the class about principles we read in a book but haven't practiced.

### The Work Project Comparison

Michael's work project (with the second brain and Slack scanning) has:
- **Automatic scanning** of 4 Slack channels → surfaces interesting items proactively
- **The same 11-step cycle** applied to that project
- **60% utilization** of the principles (Michael's estimate)

Our project has:
- **Manual capture** via Discord/web/becoming app — nothing proactive
- **The same 11-step cycle** as documentation
- **~28% utilization** of the principles

The work project is MORE automated and MORE disciplined than this one. And the AI at work spent 1000 tokens calling Michael out for having theory without practice. The same criticism applies here — possibly more so.

### What This Means for the Squad Adoption

The Squad analysis (sections I-IX above) identified 6 adoption items. Honest question: **will we actually build them, or will they become 6 more items on the "designed but not started" list?**

Check the inventory:
- 19 numbered plans (6 done, 13 not)
- 9 formal proposals (2 implemented)
- 5+ doc-level roadmaps
- And now a Squad adoption plan with 6 items

The pattern is clear: **we generate plans faster than we execute them.** Adding 6 more items doesn't change the throughput. And the planning itself is arguably an elaborate form of avoidance dressed up as productivity.

The honest recommendation: **before adopting anything new from Squad, practice what we already wrote.**

Specifically:
1. **Add intent.yaml to the session-start sequence.** 5-minute change. Makes Intent actually operational.
2. **Create decisions.md.** The one Squad adoption that costs nothing and delivers immediately.
3. **Practice Sabbath.** After this session, stop. Don't start another planning doc. Let the work breathe.
4. **Track one covenant item for a week.** Just one. "Did I review agent output within 24 hours?" Yes/no log. See what happens.

Everything else — hooks, routing, cost tracking, lockout — only matters if the foundation is real. And right now, the foundation is 28% real.

### The Generous Reading

Not everything is bleak. What we DO practice is real:

- **Spiritual Creation (80%)** — proposals before code genuinely works. data-safety shipped because the spec was precise.
- **Atonement (65%)** — the data-loss → checklist → production-fix cycle was genuine. One incident, but it was real.
- **Source Verification** — the skill works. Confabulation dropped dramatically after the Feb 28 learning.
- **exp1 agents** — the phased workflow with scratch files is a genuine improvement validated by use.
- **Session-journal format** — when used, it captures meaningful relational data, not just facts.

The gap isn't that we've built nothing real. It's that we WROTE 11 steps and PRACTICE 4-5 of them, then told Squad they could learn from our full 11. That's the "too complimentary" pattern Ben flagged — applied to ourselves.
