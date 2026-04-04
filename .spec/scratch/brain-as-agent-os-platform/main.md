# Research: Brain as Agent OS Platform

**Entry ID:** brain-agent-os-2026-04
**Category:** ideas
**Captured:** 2026-04-04
**Status:** Active architecture, phases in-flight

---

## What This Is About

Develop the scriptue-study workspace's "brain" (brain.exe, an always-on personal agent) into a full operating system platform for multi-agent work. This would enable autonomous agent dispatch, workspace-aware sessions, and composition of specialized agents (study, teaching, dev, etc.) under a unified governance model—similar to an "agentic OS" concept.

The differentiators under investigation are specific domain applications: Sunday School instruction, space center tours, brain development infrastructure, and workspace improvements.

---

## What Already Exists

[WORKSPACE] **Brain Architecture is Already Live**

The scripture-study workspace has **already implemented a working multi-agent OS architecture** (not just a concept):

### 1. **Core "Brain as Agent OS" Foundation**
- **File:** `.spec/proposals/deferred/second-brain-architecture.md` (codename: Garvis, now brain.exe)
- **Status:** Formally adopted March 19, 2026 — *decision: brain.exe IS the second brain*
- **Architecture:** Always-on personal agent OS with:
  - SQLite + vector DB (chromem-go) for memory
  - 7 MCP servers (gospel, webster, becoming, byu-citations, search, yt, relay)
  - Relay for async ingress (Discord, CLI, ibeco.me)
  - Classifiers running on local qwen3.5-9b (4090)
  - Proactive surfacing (morning digest, weekly review)
  - HITL feedback loop (one-step corrections)

### 2. **Multi-Agent Routing Architecture**
- **File:** `.spec/proposals/brain-multi-agent/main.md` + scratch version
- **Status:** Phase 3c deployed April 2, 2026 (auto-routing + review queue)
- **Pattern:** Capture → Classify → Route → Agent Session → Output → Review → Store
- **11 Agents deployed:**
  - Reasoning: study, dev, lesson, eval, journal, plan, review, talk, podcast, story, ux
  - Classification: 6 categories (people, projects, ideas, actions, study, journal)
  - Routing modes: "auto" (dispatch automatically), "suggest" (present to human), "none" (store only)

### 3. **Workspace-Aware Sessions**
- **File:** `.spec/proposals/brain-workspace-aware/main.md` + scratch version
- **Status:** Phase 2.5 shipped March 2026
- **How it works:**
  - Parses `.github/agents/` (11 agent definitions)
  - Parses `.github/skills/` (13+ skill definitions)
  - Auto-discovers MCP servers in `.vscode/mcp.json`
  - Embeds agent definitions + skills in system message at session start
  - Result: Each agent session "knows" what it's authorized to do and what tools are available
- **Trade-off tested:** SDK-native CustomAgents vs Squad-style file embedding. Chose hybrid approach.

### 4. **Memory Architecture (Foundation for Agent Autonomy)**
- **File:** `.spec/proposals/memory-architecture.md`
- **6 Tiered Memory Types:**
  1. Identity Memory (permanent) — relational dynamics, Abraham 4-5 pattern
  2. Project Knowledge (evergreen) — theological framework, learned tools
  3. Procedural Memory (patterns) — evaluation workflow, study methodology
  4. Episodic Memory (recency-weighted) — journal entries with session context
  5. Project State (ephemeral) — active work, blockers, open questions
  6. Personal Context (preferences) — calling (Sunday School President), schedules
- **Loaded every session** — establishes relational frame and agent context

### 5. **Governance Model**
- **Files:** `.spec/memory/decisions.md`, `.spec/scratch/squad-analysis/main.md`
- **Principles:**
  - Gated autonomy (agents wait for human-assigned specs)
  - Dual AI backend (qwen3.5 for classification, Copilot SDK for reasoning)
  - Hook-based governance (OnPreToolUse/OnPostToolUse for file-write guards, token budgets, audit)
  - Ceremonial checkpoints (session start/end rituals, proposal before multi-agent work)
  - Cost tracking (1500/month capacity, 56% utilization as of late March)
  - Model escalation on rejection (rejected work → escalate model tier → escalate agent → escalate to human)

---

## Differentiators Identified

[WORKSPACE] **Existing Differentiators:**

1. **Workspace-Aware Agent Context**
   - Unlike generic agent frameworks, this brain OS auto-loads:
     - Agent definitions (.github/agents/*.md)
     - Skill definitions (.github/skills/*.md)
     - MCP server inventory from .vscode/mcp.json
   - Result: Agents are contextually aware of their boundaries and tools
   - Differentiator: *Portable architecture, contextual instantiation*

2. **Theological Framework as System Message Layer**
   - **Files:** `gospel-vocab.md`, `titsw-framework.md`, `.spec/memory/identity.md`
   - Abraham 4-5 pattern (Council → Spiritual Creation → Physical Creation → Watch → Correct → Rest)
   - Not just a process; embedded in every agent's system prompt
   - Result: Agent behavior aligns with relational stewardship principles
   - Differentiator: *Values-driven agent composition*

3. **TITSW Framework (Theological Integration & Typological Study Work)**
   - **Files:** `.spec/scratch/lm-studio-model-experiments/main.md`
   - 6 differentiators across meta-principles + 4 principles
   - Enables typological scoring at scripture level (15-column enrichment)
   - Differentiator: *Domain-specific enrichment for theological work*

4. **Gated Autonomy + Ceremonial Checkpoints**
   - Not "fully autonomous" but not "fully manual" either
   - Agents propose work, humans approve specs, agents execute, humans review
   - Sessions end with memory updates (journal entry + active.md + principles updates)
   - Differentiator: *Progressive trust + mandatory reflection*

---

## Differentiators To Explore (Sunday School, Space Center, Brain Dev, Workspace)

[SYNTHESIS] **Potential application-specific differentiators:**

### 1. **Sunday School Context**
- **Known:** Michael is Sunday School president (personal stewardship constraint)
- **Known:** "teaching" agent exists in agent pool
- **Opportunity:** 
  - Real-time lesson preparation with brain-aware research agent
  - Class discussion prediction (semantic search of conference talks for likely questions)
  - Agent-generated handouts aligned to gospel library canon
  - Differentiator: *Stewardship-specific agentic workflows*

### 2. **Space Center Context**
- **Status:** Not found in workspace (yet)
- **Opportunity:**
  - If there's a space center tour/presentation workflow, could build agent-driven docent prep
  - Differentiator: *Specialized domain agent for consistent, contextual tour delivery*
- **Question:** Is there an existing space center project or reference in the workspace?

### 3. **Brain Development Infrastructure**
- **Known:** Phase 3c (April 2) deployed "auto-routing + review queue"
- **Opportunity:**
  - Agent pools with specialized routing (e.g., high-trust agents don't need review)
  - Model selection per task type (Haiku for classification, Sonnet for study, Opus for complex reasoning)
  - Cost tracking + progressive autonomy grants
  - Differentiator: *Capability-aware agent bootstrapping*
- **In-flight work:** Understand model selection heuristics, review queue UX

### 4. **Workspace Improvements**
- **Known:** Phase architecture (schema → indexing → enrichment → search → batch → cutover)
- **Opportunity:**
  - Agent-driven workspace onboarding (walk new agents through directory structure)
  - Automatic guardrails suggestion (analyze agent code, suggest guardrails.md patterns)
  - Cross-project brain sync (brain entries auto-link to relevant studies)
  - Differentiator: *Self-managing workspace topology*

---

## Open Questions

1. **Sunday School Differentiator:**
   - What specific lesson-prep workflows would the agent handle?
   - Should agents have access to class roster / attendance data?
   - How autonomous should lesson recommendations be? (Suggest topics vs. auto-generate agendas?)

2. **Space Center Context:**
   - Does a space center project exist in the workspace or is this a new idea?
   - What's the delivery model? (Pre-recorded docent scripts? Real-time agentic narration? Mixed?)

3. **Brain Development Infrastructure:**
   - Model selection rules: How to decide Haiku vs. Sonnet vs. Opus per task type?
   - Review queue prioritization: Does human review happen for all agent work or only high-stakes (file writes, tool calls)?
   - Progressive autonomy: How to grant agents higher trust tier? (Perfect execution record? Time-in-service? Domain expertise?)

4. **Workspace Improvements:**
   - Should the brain auto-update memory (active.md, decisions.md) or only suggest?
   - How to balance "clean workspace" with "accessible history"? (Archive threshold?)
   - What's the feedback signal for "workspace improvement worked"? (Less agent context needed? Faster task completion?)

5. **Platform Positioning:**
   - Is this "Brain as Agent OS" aimed at:
     - **Personal productivity** (Michael's stewardship workflow)?
     - **Teachable platform** (exportable patterns for others)?
     - **Theological reasoning engine** (domain-specific agent capabilities)?
     - **All three**?
   - Affects which differentiators to prioritize.

6. **Competitive Landscape:**
   - What "earlier video" about agent OS is the reference? (To calibrate scope/ambition)
   - How does this compare to existing agentic frameworks (Anthropic SDK, Strands, etc.)?
   - Is the thesis "different architecture" or "same architecture + theological + stewardship layer"?

---

## External Context

[WEB] **Agent OS Landscape:**

### Anthropic's Framework on Agentic Systems
- **Source:** https://www.anthropic.com/research/building-effective-agents
- **Key insight:** Agentic systems trade latency and cost for better task performance. Most successful implementations use *simple, composable patterns*, not complex frameworks.
- **Distinction:** Workflows (predefined code paths) vs. Agents (LLMs dynamically direct tool usage)
- **Recommendation:** Start simple, add complexity only when needed
- **Applicable to this project:** scripture-study brain already has complexity (11 agents, routing, memory tiers). Design check: Are all layers necessary, or can architecture simplify?

### Wikipedia: Autonomous Agent Definition
- **Source:** https://en.wikipedia.org/wiki/Autonomous_agent
- **Franklin & Graesser (1997) definition:** "An autonomous agent is a system situated within and a part of an environment that senses that environment and acts on it, over time, in pursuit of its own agenda and so as to effect what it senses in the future."
- **Spectrum:** Humans/animals (high autonomy, multiple drives/senses) ← → thermostats (single sense, one action, simple control)
- **Modern applications:** Agentic AI (Devin), IoT integration, enterprise automation
- **Applicable to this project:** The gated autonomy model sits in the middle-to-lower autonomy spectrum by design (human-assigned specs, ceremonial checkpoints). This is intentional — not a limitation.

---

## Raw Sources

### Workspace References (Existing Implementation)
- `.spec/proposals/deferred/second-brain-architecture.md` — Original brain vision
- `.spec/proposals/brain-multi-agent/main.md` + `.spec/scratch/brain-multi-agent/main.md` — Routing architecture
- `.spec/proposals/brain-workspace-aware/main.md` + `.spec/scratch/brain-workspace-aware/main.md` — Workspace integration
- `.spec/proposals/memory-architecture.md` — Memory types and lifecycle
- `.spec/memory/identity.md` — Abraham 4-5 pattern, relational frame
- `.spec/memory/decisions.md` — 15+ canonical decisions (gated autonomy, dual AI backend, cost tracking)
- `.spec/memory/active.md` — Current state, in-flight work
- `.spec/scratch/squad-analysis/main.md` — Governance patterns imported from Squad
- `.spec/scratch/lm-studio-model-experiments/main.md` — TITSW differentiators
- `.spec/proposals/overview/main.md` — WS1 (Agentic Foundation) phase tracking
- `private-brain/README.md` — Brain agent operational docs
- `private-brain/.brain/guardrails.md` — Non-negotiable constraints

### External References
- https://www.anthropic.com/research/building-effective-agents — Anthropic's agent patterns
- https://en.wikipedia.org/wiki/Autonomous_agent — Definition and spectrum
- https://platform.claude.com/docs/en/agent-sdk/overview — Claude Agent SDK (in use)

---

## Next Steps for Human Discernment

1. **Clarify the reference:** What "earlier video" about agent OS is the inspiration? (Helps calibrate scope)
2. **Prioritize differentiators:** 
   - Sunday School use case feels most grounded (direct stewardship context)
   - Space center needs more context (project status?)
   - Brain development infrastructure is already in-flight (phase 3c)
   - Workspace improvements feel perpetual (worth prioritizing?)
3. **Positioning decision:** Personal productivity vs. teachable platform vs. theological engine vs. hybrid?
4. **Review existing work:** Before new proposals, ensure the 11 agents + workspace-aware sessions + memory model align with the vision you have in mind.

---

**Research completed:** 2026-04-04 17:50 UTC
**Researcher notes:** Workspace already has a working multi-agent OS. The question isn't "how to build one" but "what specific domain applications (Sunday School, space center, etc.) should be the primary differentiators?" and "what does the platform claim to do differently than existing frameworks?" This research surfaces the existing foundation and flags the open questions needed for human discernment.
