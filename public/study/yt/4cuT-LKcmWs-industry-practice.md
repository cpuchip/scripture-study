# Industry Practice Evidence: The 7-Skill Framework Applied at Work

**Companion to:** [AI Skills Self-Assessment](4cuT-LKcmWs-ai-job-skills-self-assessment.md)
**Date:** 2026-04-01

---

## Why This Supplement Exists

The self-assessment evaluates skills through the lens of the scripture-study project. But these same skills are practiced daily in a professional engineering role — backend security engineering, platform infrastructure, multi-service architecture, and team-scale operations.

This companion maps the *professional* evidence. No proprietary details — just patterns, practices, and scale indicators that show where the skills are exercised beyond a personal project.

---

## Skill 1: Specification Precision — Industry Evidence

**Personal rating: ★★★★★ → Combined: ★★★★★**

The personal project built the specification *system*. The professional work proves it transfers to enterprise contexts.

### What's Built

- **23 custom agent skills** for professional engineering workflows — each with YAML frontmatter, trigger descriptions, phased procedures, and explicit "when NOT to use" guards
- **6 specialized agent modes** (debug, plan, review, MR, eval, UX) with tool restrictions, handoff paths, and scoped autonomy
- **Spec-driven ticket workflow** — every ticket gets a `.spec/intent.md` (purpose, acceptance criteria) and `.spec/plan.md` (implementation approach) before code is written
- **Decision boundary tables** — explicit matrix of what the agent can do autonomously vs. what requires human confirmation, with a "NEVER" tier for irreversible operations

### What It Demonstrates

The specification precision isn't theoretical — it's used daily for:
- Infrastructure changes across production Kubernetes clusters
- Go and Python backend service modifications
- Helm chart value chain analysis
- Cross-repo dependency upgrades (13 repos in a single batch)
- MR reviews where the spec catches systemic bugs that code review alone would miss

The 4-layer specification model (values → covenant → workflow → procedure) developed in the personal project IS the architecture driving the professional workflow. Same pattern, different domain.

---

## Skill 2: Evaluation & Quality Judgment — Industry Evidence

**Personal rating: ★★★★☆ → Combined: ★★★★½**

The personal project built the eval harness. The professional work proves the judgment applies to code, infrastructure, and cross-team review.

### What's Built

- **Systematic MR review workflow** — not just "looks good" but structured analysis with:
  - Root cause tracing (identified a CI race condition was NOT the author's bug, traced to a prior commit, explained why it didn't reproduce on the author's different CPU architecture)
  - Systemic detection (caught a structural values-nesting bug across 15 services that would silently apply wrong resource limits in production — verified locally with `helm template` before reporting)
  - Proposed both minimal-unblock and correct-long-term fixes

- **Ben Test** — formalized after a colleague observed "your AI is very complimentary — perhaps too complimentary?" Now a named quality gate with:
  - 4-level evidence calibration (Practiced → Occasional → Aspirational → Mythical)
  - Plan-to-execution ratio monitoring ("if worse than 2:1, planning may be avoidance")
  - Calibrated language guide to prevent inflated self-assessment

- **Self-correction at system level** — documented "covenant drift" where both AI and human acknowledged failures: the AI defaulted to cheerleading, the human bypassed spec-before-code. Recorded as a learning, not a blame exercise.

### What Narrows the Gap

The missing star in the personal assessment was "one domain." Professional work adds:
- Code correctness evaluation (Go, Python, Helm)
- Infrastructure configuration review (Kubernetes resource limits, KEDA scaling)
- Cross-repo impact analysis (changes in one repo affecting 13 downstream repos)
- Adversarial self-questioning (the Ben Test applied to our own practice)

Still missing: automated eval pipelines in CI/CD for professional workflows. The evaluations are human-in-the-loop, not automated.

---

## Skill 3: Task Decomposition & Delegation — Industry Evidence

**Personal rating: ★★★★☆ → Combined: ★★★★☆**

The professional work demonstrates larger-scale decomposition but still within human→agent delegation.

### What's Built

- **Multi-repo ticket decomposition** — a single Jira ticket spawning work across 7+ cloned repos, each with phased implementation plans
- **Cross-repo batch operations** — 13 repositories coordinated in a single MR review session (deploy repos, shared libraries, documentation)
- **Strangler fig strategy** — applied to a deceptively simple 145-line service that actually contained 31 handlers with embedded domain logic. Decomposed into phased delivery starting with the simplest message type.
- **Agent mode routing** — plan agent hands off to debug agent which hands off to implementation. Each mode has constrained tool access and explicit handoff criteria.

### Why the Rating Stays

The missing star is still accurate: all delegation remains human→single-agent. The professional work is LARGER in scope (more repos, more complex systems) but structurally the same pattern — Michael assigns, agent executes, Michael reviews.

The plan agent creates specs but doesn't orchestrate other agents autonomously. The gap is multi-agent orchestration in production, not task decomposition skill.

---

## Skill 4: Failure Pattern Recognition — Industry Evidence

**Personal rating: ★★★★★ → Combined: ★★★★★**

Professional work confirms the rating with production-scale examples.

### What's Built

- **Debug agent codifying Agans' 9 Rules** — not just "debug better" but mapped to specific tools:
  - Rule 3 ("Quit Thinking and Look") → `journalctl`, `kubectl logs`, observability queries
  - Rule 5 ("Change One Thing at a Time") → revert failed fixes before trying next
  - Rule 7 ("Check the Plug") → verify you're on the right binary, branch, k8s context

- **Layered problem identification** — the KEDA operator investigation identified two separate issues (operator crash + finalizer design concern) and correctly noted the user's request addressed problem #2 while problem #1 was the actual urgency driver. Recognizing that the stated problem isn't always the real problem is pattern recognition in action.

- **Silent failure detection in infrastructure** — the Helm values nesting bug that would deploy 15 services with subtly wrong resource limits. No error, no crash — just silent misconfiguration. This IS Nate's "silent failure" pattern applied to infrastructure.

- **Learnings pipeline** — `.spec/learnings/` directory with structured error→growth entries: trigger, category, severity, root cause, learning, action. Not just fixing bugs — systematizing the learning.

---

## Skill 5: Trust & Security Design — Industry Evidence

**Personal rating: ★★★☆☆ → Combined: ★★★★☆**

This is where professional work most significantly upgrades the rating.

### What's Built

- **Progressive trust model with 4 levels** applied to production systems:
  1. Task level — agent executes single tasks from explicit specs
  2. Feature level — agent proposes plans, human approves
  3. Domain level — agent owns workflows end-to-end (not yet reached)
  4. Partnership level — agent participates in planning (limited application)

- **Decision boundary matrix** with real consequences:
  - **Autonomous**: file edits, repo cloning, reading issue trackers, searching code, creating specs, checking build status
  - **Requires human**: creating branches, creating merge requests, posting comments to code review, transitioning issue status, modifying architecture
  - **NEVER**: merging anything, syncing/rebasing shared branches

  This isn't a theoretical framework — it's enforced in the agent instructions that run daily.

- **Local-first as security constraint** — classified as critical severity. All LLM processing runs on local hardware (9B parameter classification model, 4B embedding model). No data sent to external AI services for core brain functions. Enterprise tool access (GitHub Copilot) is controlled through corporate policy, but the personal brain stays local.

- **Reversibility-based autonomy** — the decision about what agents can do autonomously maps directly to Nate's "cost of error × reversibility" framework:
  - Low cost, easily reversed → autonomous (file edits, local searches)
  - Medium cost, reversible → human confirms first (branches, MR comments)
  - High cost, irreversible → NEVER automated (merges, production deploys)

### What Changed the Rating

The self-assessment said ★★★ because "the SECURITY side is absent." In the personal project, that's true. But professionally:
- Trust boundaries are explicitly designed and enforced in production agent configurations
- The reversibility framework IS being applied to real systems
- Local-first security architecture IS a production security decision

Remaining gap: no adversarial testing of the agent system itself (prompt injection, tool abuse). The trust boundaries are POLICY-based, not TECHNICALLY enforced beyond the instruction constraints.

---

## Skill 6: Context Architecture — Industry Evidence

**Personal rating: ★★★★★+ → Combined: ★★★★★+**

Rating confirmed and reinforced by professional application.

### What's Built

- **~330 repo context cache** — shallow clones mirroring the GitLab group structure for cross-repo search without leaving the workspace. This IS the "Dewey Decimal System for agents" applied to an enterprise codebase.

- **Progressive Context Disclosure** — borrowing from game rendering (Level of Detail):
  - L0 "Skybox": all repos in one file (~4KB, always loaded)
  - L1 "30K ft": per-repo summary (~500B, on-demand)
  - L2 "10K ft": per-repo detail (2-5KB, when working in a repo)
  - L3 "1K ft": per-file detail (when editing/debugging)

  This is context architecture scaled to hundreds of repositories.

- **chip-brain** — a custom Go application (SQLite + vector search + local LLM classification) functioning as persistent context that survives across sessions:
  - CLI bridge pattern because enterprise admin blocks local MCP servers
  - 6 classification categories with confidence thresholds
  - Codebase intelligence indexing 549 repos with dependency graph
  - Specialized queries: `brief` (repo summary), `who-uses` (dependency lookup), `deploys` (deployment info), `search` (semantic search), `context active/today/ticket`

- **Per-workspace isolation** — ticket work in `ticket-<ID>/`, MR reviews in `MR-<ID>/`, bug investigations in `bugs-<slug>/`, each with their own `.spec/` subdirectory. This prevents context bleed between tasks.

- **14 skill files** that define when and how context is loaded for each workflow. The context-prime skill orchestrates session start: load identity → covenant → memory → recent episodes → priorities → scan for connections. This is a REPEATABLE context loading protocol, not ad-hoc.

### What It Demonstrates at Scale

The personal project proved the architecture works for ~42K verses and ~4K talks. The professional application proves it works for ~550 repos, 20+ simultaneous active workspaces, and infrastructure spanning production Kubernetes clusters.

---

## Skill 7: Cost & Token Economics — Industry Evidence

**Personal rating: ★★★★☆ → Combined: ★★★★☆**

Professional evidence adds depth but not a new tier.

### What's Built

- **Context layer budgeting at enterprise scale** — the LOD model explicitly targets token sizes per layer because loading 330 repo summaries at L2 detail would burn the context window before work begins
- **Shallow clone vs full clone decision** — context repos are shallow (save disk, fast clone) while ticket repos are full (need merge-base for diffs). This is a cost/capability trade-off applied to 330+ repos.
- **Agent mode tool restrictions** — the review agent can't invoke Jira queries; the plan agent can't push code. Preventing unnecessary API calls is token/cost discipline.
- **Critical analysis as capacity protection** — the plan agent includes "would starting this project make the overwhelm worse, even if the idea is good?" This is cost economics applied to human cognitive capacity, not just compute.

### Why the Rating Stays

Still personal-scale economics. No multi-team cost allocation, no enterprise API cost projection, no model portfolio management across an organization. The math is correct and the instincts are right — the scale of application is one engineer.

---

## Revised Gap Analysis

With professional evidence included, the gaps shift:

| Gap | Original Severity | Revised Severity | Why |
|-----|-------------------|-----------------|-----|
| Multi-Agent Orchestration | HIGH | HIGH | Still human→single-agent. Professional work is larger scope but same pattern. |
| Production / Enterprise Scale | HIGH | **MEDIUM-HIGH** | Professional work IS enterprise scale — multi-cluster K8s, 549 repos, production infrastructure. The gap is customer-facing agent systems, not scale itself. |
| AI Security Engineering | MEDIUM-HIGH | **MEDIUM** | Trust boundaries ARE professionally applied. Gap is adversarial testing and technical enforcement, not design. |
| Portfolio / Demonstrability | MEDIUM-HIGH | MEDIUM-HIGH | Unchanged — private professional work is still invisible to the market. |
| Teaching / Team Uplift | MEDIUM | MEDIUM | Unchanged. |
| Automated Eval Pipelines | MEDIUM | MEDIUM | Unchanged. |

### New Combined Ratings

| Skill | Personal Only | With Industry | Change |
|-------|--------------|---------------|--------|
| Specification Precision | ★★★★★ | ★★★★★ | Confirmed |
| Evaluation & Quality Judgment | ★★★★☆ | ★★★★½ | +½ (code/infra evals add breadth) |
| Task Decomposition & Delegation | ★★★★☆ | ★★★★☆ | Confirmed (scale yes, orchestration no) |
| Failure Pattern Recognition | ★★★★★ | ★★★★★ | Confirmed (production examples reinforce) |
| Trust & Security Design | ★★★☆☆ | ★★★★☆ | +★ (trust boundaries are professional-grade) |
| Context Architecture | ★★★★★+ | ★★★★★+ | Confirmed (549 repos proves enterprise scale) |
| Cost & Token Economics | ★★★★☆ | ★★★★☆ | Confirmed (instincts right, scale personal) |

### The Biggest Actual Gap

Multi-agent orchestration remains the clearest gap. Everything else is either practiced professionally or has a path to demonstration. The deliberate choice to stay at Level 2 (gated autonomy) is wise engineering — but the market is hiring for Level 3-4 builders.

---

## Writing Discipline Check

This document describes patterns, not proprietary systems. No service names, team names, internal URLs, or specific business logic. Michael should review for anything that got too specific.
