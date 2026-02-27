# AI Hardware & Infrastructure Proposal: From Individual Tools to Team-Scale Agentic Development

**Part of:** [Intent-Driven Development Research](00_index.md)
**Date:** February 2026
**Status:** Proposal Draft
**Audience:** Engineering Leadership

---

## Executive Summary

We've invested in the right tools — GitHub Copilot Business, Cursor licenses, talented engineers hungry to go deeper. But we're hitting a ceiling that more cloud API calls can't solve. Our brownfield codebase is too large to fit in any single context window. Our laptops sleep, lose VPN, and kill long-running agent sessions. Our Python→Go migration requires sustained, cross-referencing analysis that current hardware can't maintain.

**This proposal recommends a phased experiment** — starting with what we can do *today* on current hardware, then investing in shared team infrastructure that transforms individual AI assistance into **team-scale agentic engineering**.

The ask: **$5,000–$8,000** for a shared Mac Studio or equivalent that serves the entire team. Not per-developer hardware upgrades — one shared resource that multiplies what every engineer can do.

---

## The Landscape: What Changed in 2025–2026

The industry shifted from "AI-assisted coding" to "agentic engineering" — agents that reason, plan, act across files, and maintain context over hours, not seconds.

Key signals:

- **90% of Fortune 100** now use GitHub Copilot ([Will Hackett, Dec 2025](https://www.willhackett.uk/agentic-coding-workforce/))
- **Gartner projects 90%** of enterprise devs will use AI code assistants by 2028, up from 14% in early 2024
- **84% of developers** use AI tools; **41% of all code** is now AI-generated ([Index.dev, 2026 stats](https://www.index.dev/blog/developer-productivity-statistics-with-ai-tools))
- **McKinsey found 2x speed gains** on coding tasks, with documentation tasks completing in half the time ([McKinsey Digital, June 2023](https://www.mckinsey.com/capabilities/mckinsey-digital/our-insights/unleashing-developer-productivity-with-generative-ai))
- **ZoomInfo study (400+ devs):** 33% acceptance rate for completions, **72% developer satisfaction** ([arXiv, Jan 2025](https://arxiv.org/abs/2501.13282))

But the *real* shift is organizational. Will Hackett frames it precisely:

> "A team of 14 that used to own a product vertical might now need only 3 to maintain the same output. That doesn't mean you fire 11. It means you suddenly have 11 engineers available for priorities that were stuck in the backlog. **The constraint on what your company can build just changed.**"

We're not trying to make individual developers faster. We're trying to **change what the team can take on.**

---

## Current State: What We Have

| Resource | Specs | AI Capability |
|----------|-------|---------------|
| M3/M4 MacBooks | 64GB unified memory | Run 14–27B models locally (comfortable); Copilot + Cursor daily drivers |
| Dell Linux workstation | 12GB Nvidia GPU | Can serve a 7–14B model (Phi-4, Gemma 3 12B); limited for larger models |
| GitHub Copilot Business | Per-seat license | Code completions, chat, agent mode in VS Code |
| Cursor licenses | Per-seat license | AI-native editor, background agents, tab completions |
| Cloud APIs | Pay-per-token | Claude, GPT-4, etc. — no data locality guarantees |

### What We Can Do Today (No New Hardware)

These require only configuration and workflow changes:

1. **Local coding assistants via Ollama/LM Studio** — Run a coding model (Phi-4 14B, Codestral 22B, or Gemma 3 12B) on existing Macs for offline completions and faster iteration loops. ~8–14GB RAM at Q4 quantization leaves headroom for IDE and browser.

2. **MCP-based tool integration** — Model Context Protocol lets local or cloud models call into project-specific tools (database queries, API exploration, test runners). Works today with Cursor and VS Code + Copilot.

3. **Spec-driven development workflow** — A `.spec/` directory per project with intent preambles, task files, and learning logs. Zero hardware cost, immediate improvement in agent output quality. (See [synthesis doc](04_synthesis.md) for the full pattern.)

4. **Codebase indexing with embeddings** — Generate vector embeddings of our codebase using a small local model. On 64GB Macs, we can index a medium codebase and query it semantically. This helps agents find relevant code across our brownfield repos.

5. **Cloud API orchestration** — Use cloud models (Claude, GPT-5.3) as reasoning engines while local models handle code search, linting, and repetitive tasks. This hybrid approach keeps costs reasonable while accessing frontier capabilities.

**Limitations at this tier:**
- Laptops sleep after inactivity → kills multi-hour agent sessions
- VPN drops → breaks agent connections to internal resources
- 64GB limits model size to ~27B quantized — adequate for coding (Codestral 22B, Gemma 3 27B) but not for complex reasoning or large codebase analysis
- No shared resource → each developer maintains separate local model setups
- Cannot run long-running background agents (indexing, migration analysis, test generation) without tying up a developer's primary machine

---

## The Gap: Why Current Hardware Hits a Ceiling

### Problem 1: Brownfield Complexity

Our codebase is large, interconnected, and being ported from Python to Go. This is precisely the scenario where AI agents provide the most value — and where they need the most context. A 14B model on a laptop can autocomplete a function. It cannot:

- Hold architectural understanding of how 200 Python modules map to 50 Go packages
- Cross-reference type signatures, error handling patterns, and test coverage across the migration boundary
- Run sustained analysis that generates migration plans spanning multiple sessions

### Problem 2: Session Persistence

Agentic workflows need to run for hours — indexing a codebase, running test suites, generating migration scaffolds, validating spec alignment. Laptop limitations:

- **Sleep/lock policies** interrupt long-running tasks
- **VPN timeouts** break connections to internal git, artifact stores, and APIs
- **Thermal throttling** degrades model performance on sustained loads
- **Developer blocked** — can't use their machine while an agent occupies it

### Problem 3: Fragmented Setup

Every developer running their own local model setup means:
- N different model versions, quantizations, and configurations
- No shared embeddings database — each person re-indexes
- No consistent baseline for evaluating agent output quality
- Duplicated effort in model tuning and prompt engineering

---

## Proposed Investment: Shared Team Infrastructure

### Option A: Mac Studio M4 Ultra (Recommended)

| Configuration | Price | Capability |
|---------------|-------|------------|
| Mac Studio M4 Ultra, 96GB | ~$4,000 | Runs Llama 4 Scout (109B MoE) at Q4; serves 3–5 concurrent developers |
| Mac Studio M4 Ultra, 192GB | ~$5,500 | Runs Llama 4 Maverick (400B MoE) or Mistral Large 3 (675B MoE) at Q4; comfortable headroom for concurrent users and embeddings databases |

**Why Mac Studio:**
- **Unified memory architecture** — GPU and CPU share the same memory pool. No PCIe bandwidth bottleneck. A Llama 4 Maverick model (~110GB at Q4) loads directly into GPU-accessible memory — something that would require multiple discrete GPUs otherwise.
- **Always-on** — Desktop form factor, no sleep/lid/battery issues. Runs 24/7 on a shelf or under a desk.
- **Silent operation** — Can live in an office or server closet.
- **Apple Silicon inference speed** — MLX framework, llama.cpp, and Ollama are highly optimized for Apple Silicon. Competitive tokens/second with dedicated Nvidia cards for models that fit in memory.
- **macOS ecosystem** — Familiar to the team. Easy to set up Ollama, LM Studio, or vLLM.

### Option B: Ryzen AI 395+ Max Workstation

| Configuration | Price | Capability |
|---------------|-------|------------|
| Ryzen AI 395+ Max, 128GB unified | ~$3,000–$5,000 | 128GB unified memory for large models; newer architecture; competitive inference |

**Why Ryzen AI 395+:**
- **128GB unified memory** at a lower price point than equivalent Mac Studio
- **x86 ecosystem** — easier to run Linux, Docker, and existing server tooling
- **AMD ROCm** — improving GPU compute support for AI workloads
- **Emerging platform** — newer, less battle-tested but rapidly maturing

### Option C: Dedicated GPU Server (Dell/Custom Linux)

| Configuration | Price | Capability |
|---------------|-------|------------|
| Dual RTX 5090 (32GB VRAM each) | ~$5,000–$7,000 | 64GB total VRAM; matches H100 performance for 70B models at 25% of the cost per [Introl blog](https://introl.com/blog/local-llm-hardware-pricing-guide-2025); faster inference than unified memory |

**Why dedicated GPU:**
- **Fastest inference** — dedicated VRAM is ~3–5x faster than unified memory for model serving
- **Proven Linux ecosystem** — vLLM, TGI, Ollama all optimized for Nvidia CUDA
- **Scalable** — can add GPUs over time

**Tradeoff:** Higher price for the full system. More complex setup. Louder.

### My Recommendation: Option A (Mac Studio 192GB)

The sweet spot for our team size and use cases. Fast enough for real-time assistance, enough memory for large models plus embeddings, always-on, silent, familiar ecosystem. We can always add a GPU server later if demand justifies it.

---

## What Becomes Possible

### Tier 1: Shared Model Server (Week 1)

Set up Ollama or LM Studio on the Mac Studio. Expose an OpenAI-compatible API endpoint on the local network.

**Every developer gets:**
- Access to frontier-class MoE models (Llama 4 Maverick, Mistral Large 3) for complex reasoning tasks — without using cloud API tokens
- A shared embeddings database of the entire codebase, updated nightly
- A persistent codebase assistant that knows the architecture, not just the current file
- Zero marginal cost per query after initial setup

**Model options at 192GB (non-Chinese models only per company policy):**

| Model | Origin | Quantization | RAM Required | Use Case |
|-------|--------|-------------|--------------|----------|
| Llama 4 Maverick (400B MoE, 17B active) | Meta (US) | Q4 | ~110GB | Frontier-class reasoning, architecture decisions, migration analysis |
| Mistral Large 3 (675B MoE, 41B active) | Mistral (France) | Q4 | ~95GB | Complex reasoning, code review, multi-step planning |
| Llama 4 Scout (109B MoE, 17B active) | Meta (US) | Q4 | ~60GB | Long-context analysis (10M token window), document-level tasks |
| Codestral (22B) | Mistral (France) | Q8 | ~24GB | Code generation, completion, 80+ languages, fill-in-the-middle |
| Gemma 3 (27B) | Google (US) | Q8 | ~29GB | Strong general + coding, good benchmark scores |
| Phi-4 (14B) | Microsoft (US) | Q8 | ~15GB | Fast coding assistant, lightweight reasoning |
| Nomic-embed-text | Nomic AI (US) | FP16 | ~0.5GB | Codebase vector embeddings |
| Codestral Embed | Mistral (France) | FP16 | ~1GB | Code-specific embeddings (outperforms Voyage Code 3, Cohere v4) |

The MoE (Mixture-of-Experts) models are the game changer here — Llama 4 Maverick has 400B total parameters but only activates 17B per token, meaning it delivers frontier-class performance while fitting in memory that a traditional 400B model never could.

Multiple models can run concurrently. A Codestral (22B) for fast code tasks + Codestral Embed for search + Llama 4 Scout for deep reasoning can all coexist in 192GB with room to spare.

### Tier 2: Agentic Workspace Platform (Month 1)

Deploy [Coder](https://coder.com/) — an open-source, self-hosted development infrastructure platform. Coder launched "AI Development Infrastructure for Hybrid Human and Agent Teams" in December 2025, with three key capabilities:

- **AI Bridge** — Safely routes AI tool traffic through governed infrastructure
- **Agent Boundaries** — Sandboxed execution environments for coding agents
- **Mux** — Run parallel AI coding agents in governed workspaces

**What this enables:**
- Long-running agent sessions that survive laptop sleep/VPN drops
- Parallel agents working on different parts of the migration simultaneously
- Shared workspace state — agents can access the same embeddings, specs, and context
- Governance and audit trail — leadership can see what agents are doing
- Works with VS Code, Cursor, JetBrains — developers keep their preferred tools

Coder is **open-source** (115K+ GitHub stars) and self-hosts on any infrastructure. It integrates with the Mac Studio as the compute backend.

### Tier 3: Intent-Driven Agentic Workflows (Month 2+)

This is where our investment in spec-driven and intent-driven development patterns pays off. With shared infrastructure, we can:

1. **Spec → Agent → Review cycles** — Write a spec for a migration task. Agent works overnight on the Mac Studio. Review the PR in the morning.

2. **Progressive trust architecture** — Start agents with read-only access. As they demonstrate reliability on simple tasks, expand their stewardship to include code changes, test generation, and eventually PR creation.

3. **Shared codebase intelligence** — One agent continuously indexes and understands the entire codebase. Other agents query it. New team members (human or agent) get up to speed by consulting the shared knowledge base.

4. **Brownfield migration acceleration** — The Python→Go migration is exactly where large context + sustained reasoning shines:
   - Agent analyzes Python module → generates Go equivalent → writes tests → validates behavior parity
   - Another agent reviews the output against architectural specs
   - A third agent updates documentation and migration tracking

5. **Cost optimization** — Local models for routine tasks (code search, completions, embeddings, initial drafts). Cloud APIs reserved for frontier reasoning (complex architecture decisions, novel problem-solving). This balances capability with cost.

---

## Integration with Existing Tools

### GitHub Copilot Business + Copilot Extensions SDK

We're already paying for Copilot Business. The [GitHub Copilot Extensions SDK](https://github.com/features/copilot/extensions) lets us build custom agents that:

- Tap into our local model server for codebase-aware responses
- Query our embeddings database for semantic code search
- Execute against our `.spec/` task files for guided implementation
- Report back through familiar Copilot UI in VS Code

This means developers don't need to learn new tools. Copilot becomes the interface; our local infrastructure becomes the brain.

### Cursor + Local Models

Cursor already supports custom API endpoints. Point it at the Mac Studio's Ollama endpoint:

- Tab completions from a fast local model like Codestral 22B (zero latency, zero cost)
- Background agents powered by Llama 4 Maverick or Mistral Large 3 for complex tasks
- MCP integration with project-specific tools
- All data stays on our network

### Hybrid Cloud + Local Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Developer Laptops                     │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐              │
│  │  Cursor   │  │ VS Code  │  │ JetBrains│              │
│  │  + Copilot│  │ + Copilot│  │  + AI    │              │
│  └─────┬─────┘  └────┬─────┘  └────┬─────┘              │
│        │              │              │                    │
└────────┼──────────────┼──────────────┼───────────────────┘
         │              │              │
    ┌────▼──────────────▼──────────────▼────┐
    │           Local Network                │
    │  ┌─────────────────────────────────┐  │
    │  │    Mac Studio (Shared Server)    │  │
    │  │                                  │  │
    │  │  ┌──────────┐  ┌─────────────┐  │  │
    │  │  │  Ollama   │  │  Coder      │  │  │
    │  │  │  Maverick │  │  Workspaces │  │  │
    │  │  │  Codestral│  │  + Agents   │  │  │
    │  │  │  Embeddings│ │  + Mux      │  │  │
    │  │  └──────────┘  └─────────────┘  │  │
    │  │                                  │  │
    │  │  ┌──────────────────────────┐   │  │
    │  │  │  Vector DB (Codebase)    │   │  │
    │  │  │  .spec/ Task Engine      │   │  │
    │  │  │  Migration Tracker       │   │  │
    │  │  └──────────────────────────┘   │  │
    │  └─────────────────────────────────┘  │
    └───────────────────┬───────────────────┘
                        │
              ┌─────────▼─────────┐
              │   Cloud APIs      │
              │   (Frontier only) │
              │   Claude, GPT-4   │
              │   Complex tasks   │
              └───────────────────┘
```

---

## ROI Analysis

### Direct Cost Savings

| Category | Current Monthly Cost | With Shared Server |
|----------|--------------------|--------------------|
| Cloud API tokens (team) | ~$500–$2,000/mo (growing) | ~$100–$400/mo (frontier-only) |
| Developer time on setup | ~4 hrs/dev/month maintaining local models | ~1 hr/dev/month (shared config) |
| Context switching (model issues) | Hard to quantify, real | Eliminated for routine tasks |

**Conservative estimate:** $300–$1,500/month in API savings alone. Hardware pays for itself in **3–12 months**.

### Productivity Multipliers

These are harder to quantify but more valuable:

| Capability | Without Shared Server | With Shared Server |
|------------|----------------------|-------------------|
| Overnight agent runs | Impossible (laptops sleep) | Standard workflow |
| Codebase-wide analysis | Partial (laptop RAM limits) | Complete (192GB + embeddings) |
| Parallel agent tasks | Not practical | 3–5 concurrent agents |
| New dev onboarding | Weeks of codebase archaeology | "Ask the codebase" day one |
| Migration velocity | Manual file-by-file | Agent-assisted batch processing |

### The Real ROI: Backlog Velocity

If the Will Hackett framing is even *partially* true — that agentic workflows let a smaller team maintain current output — then the freed-up engineering time against the backlog is where the real return lives. A $5,500 Mac Studio that frees even 10 hours/week of engineering time across the team pays for itself in **2 weeks** at typical engineering costs.

---

## Risk Mitigation

| Risk | Mitigation |
|------|-----------|
| Hardware underutilized | Start with Tier 1 (model server only); expand only as usage justifies |
| Models not good enough | Local models supplement, not replace, cloud APIs; always have fallback |
| Security concerns | All data stays on-premises; Coder provides governance layer; no external model training on our code |
| Team adoption resistance | Start with volunteers; let results sell the approach |
| Rapid hardware obsolescence | Apple Silicon + Ollama ecosystem is stable; 192GB is generous headroom for 2–3 years of model growth |

---

## Proposed Experiment Plan

### Phase 1: Current Hardware (Now — 2 weeks)

**Cost: $0**

- [ ] Set up Ollama on the Dell Linux box with Phi-4 (14B) or Gemma 3 (12B)
- [ ] Expose OpenAI-compatible API on the internal network
- [ ] Configure one volunteer's Cursor to use the local endpoint for tab completions
- [ ] Create a `.spec/` directory in the migration project with intent preambles
- [ ] Measure: token savings, developer feedback, completion quality

**Success criteria:** At least one developer prefers the local model for routine coding tasks. The `.spec/` workflow reduces rework on at least one migration task.

### Phase 2: Shared Server (Week 3–6)

**Cost: ~$5,500 (Mac Studio M4 Ultra 192GB)**

- [ ] Purchase Mac Studio; set up Ollama with Llama 4 Maverick + Codestral + Codestral Embed
- [ ] Generate and host codebase vector embeddings
- [ ] Connect all team Cursor/VS Code instances to shared server
- [ ] Deploy Coder for persistent workspace management
- [ ] Run first overnight agent task (e.g., generate test coverage report for 10 modules)
- [ ] Measure: queries/day, model utilization, API cost reduction, developer satisfaction

**Success criteria:** Team uses shared server daily. At least one meaningful task completed by an overnight agent. API costs visibly reduced.

### Phase 3: Agentic Workflows (Month 2–3)

**Cost: $0 additional (Coder is open-source)**

- [ ] Implement spec→agent→review cycle for migration tasks
- [ ] Set up progressive trust levels for agent permissions
- [ ] Build Copilot Extension that queries local embeddings + runs against specs
- [ ] Run parallel agents on different migration modules
- [ ] Measure: migration velocity, code quality, review turnaround

**Success criteria:** Migration velocity measurably increases. Agents produce code that passes review with fewer iterations than early experiments.

### Phase 4: Evaluate & Expand (Month 4+)

- [ ] Assess whether to add GPU compute (dual RTX 5090 for faster inference)
- [ ] Evaluate emerging models (new releases every few weeks)
- [ ] Consider extending to other teams or projects
- [ ] Document patterns and share across the organization

---

## Connection to Our Methodology

This proposal isn't just about hardware. It's the infrastructure that enables the [intent-driven development patterns](04_synthesis.md) we've been studying:

| Pattern | How Infrastructure Enables It |
|---------|------------------------------|
| **Intent preambles** | Agents on shared server can read `.spec/intent.md` before every task |
| **Spec-driven workflow** | Persistent agents maintain spec state across sessions |
| **Progressive trust** | Coder's Agent Boundaries enforce trust levels architecturally |
| **Covenant-based work** | Shared infrastructure formalizes the mutual commitment — the team invests in the machine; the machine serves the team's intent, not just individual queries |
| **Sabbath cycles** | Scheduled reflection — agent runs nightly analysis, humans review mornings |
| **Stewardship expansion** | Agents earn broader access to the codebase as they demonstrate reliable output |

The hardware is the body. The methodology is the spirit. Neither works alone.

---

## Summary

| | Now (Free) | Phase 2 ($5.5K) | Phase 3 ($0 more) |
|---|---|---|---|
| **Models** | Phi-4/Gemma 3 on laptops/Dell | Maverick + Codestral + embeddings | Same + specialized |
| **Sessions** | Limited by laptop | 24/7 persistent | Agent-managed |
| **Codebase understanding** | Current file only | Full codebase indexed | Continuously updated |
| **Agentic workflows** | Manual, fragmented | Centralized, shared | Spec-driven, automated |
| **Migration support** | One file at a time | Batch analysis | Parallel agent teams |
| **Cost model** | Growing cloud spend | Fixed hardware + minimal cloud | Optimized hybrid |

**The ask:** Approve a ~$5,500 experiment with a Mac Studio M4 Ultra (192GB) as shared team infrastructure. We start with what we have today, prove the concept on current hardware, then invest in the shared server to unlock agentic workflows that change what this team can build.

We're already investing in the tools. This gives them a body to work in.

---

## Appendix: Sources

- [Will Hackett — "Agentic coding is changing the engineering workforce"](https://www.willhackett.uk/agentic-coding-workforce/) (Dec 2025)
- [McKinsey — "Unleashing developer productivity with generative AI"](https://www.mckinsey.com/capabilities/mckinsey-digital/our-insights/unleashing-developer-productivity-with-generative-ai) (June 2023)
- [Index.dev — "Top 100 Developer Productivity Statistics with AI Tools (2026)"](https://www.index.dev/blog/developer-productivity-statistics-with-ai-tools)
- [ZoomInfo — "Experience with GitHub Copilot for Developer Productivity"](https://arxiv.org/abs/2501.13282) (Jan 2025)
- [Coder — "AI Development Infrastructure for Hybrid Human and Agent Teams"](https://coder.com/blog/ai-development-infrastructure-for-hybrid-human-and-agent-teams) (Dec 2025)
- [Coder — "AI Maturity Self-Assessment"](https://markets.businessinsider.com/news/stocks/coder-launches-ai-maturity-self-assessment-to-help-enterprises-benchmark-agentic-ai-adoption-in-software-development-1035761267) (Jan 2026)
- [Introl — "Local LLM Hardware Guide 2025: GPU Specs & Pricing"](https://introl.com/blog/local-llm-hardware-pricing-guide-2025) (Aug 2025)
- [SitePoint — "Guide to Local LLMs in 2026: Privacy, Tools & Hardware"](https://www.sitepoint.com/definitive-guide-local-llms-2026-privacy-tools-hardware/)
- [SitePoint — "The Complete Stack for Local Autonomous Agents"](https://www.sitepoint.com/the-complete-stack-for-local-autonomous-agents--from-ggml-to-orchestration/) (Feb 2026)
- [SelfHostLLM — Mac LLM Compatibility Calculator](https://selfhostllm.org/mac)
- [Cline vs Cursor 2026](https://dev.to/tan_genie_6a51065da7b63b6/cline-vs-cursor-2026-open-source-vs-proprietary-ai-coding-2aen) (Feb 2026)
- [GitHub Copilot Extensions SDK](https://github.com/features/copilot/extensions)
