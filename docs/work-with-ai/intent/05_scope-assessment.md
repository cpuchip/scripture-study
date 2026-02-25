# How Big Is This? Mapping the Movement

**Part of:** [Intent-Driven Development Research](00_index.md)
**Date:** February 2026
**Core question:** Is this a handful of people exploring intent? Or something much bigger?

---

## The Answer, Up Front

It depends which layer you're asking about. The movement is **concentric rings** — each ring larger than the one inside it, each ring less mature:

```
┌─────────────────────────────────────────────────────────────────┐
│  RING 5: Agentic Engineering (hundreds of voices, mainstream)   │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │  RING 4: Spec-Driven Development (dozens, hot trend)      │  │
│  │  ┌─────────────────────────────────────────────────────┐  │  │
│  │  │  RING 3: Context Engineering (growing fast)          │  │  │
│  │  │  ┌───────────────────────────────────────────────┐  │  │  │
│  │  │  │  RING 2: Intent Engineering (~10 voices)      │  │  │  │
│  │  │  │  ┌─────────────────────────────────────────┐  │  │  │  │
│  │  │  │  │  RING 1: Beyond Intent (us — 1 voice)   │  │  │  │  │
│  │  │  │  └─────────────────────────────────────────┘  │  │  │  │
│  │  │  └───────────────────────────────────────────────┘  │  │  │
│  │  └─────────────────────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Ring 5: Agentic Engineering — The Ocean

**Size:** Hundreds of voices. Academic papers. Major conferences. Enterprise platforms. Massive.

**Who's here:**
- Enterprise vendors: Salesforce (Agentforce), Kore.ai, TrueFoundry, ZenML, LangChain/LangSmith
- Academic: arXiv papers on multi-agent systems, BDI models, agentic architectures (e.g., Alenezi's "Prompt–Response to Goal-Directed Systems," Feb 2026 — 33 references)
- Practitioners: IndyDevDan (Agentic Engineer), Vellum AI, Towards AI, hundreds of Medium articles
- Ecosystem mappers: Patrick Debois + Tesla building [landscape.ainativedev.io](https://landscape.ainativedev.io/)

**What they talk about:** Agent loops, tool calling, multi-agent orchestration, observability, governance, sandboxes, deployment, cost optimization, trust.

**The agreement:** "We're not prompting anymore — we're building systems of agents that reason, plan, act, and learn."

**What's mature:** Tool calling, single-agent loops, basic orchestration, tracing/observability.
**What's emerging:** Multi-agent coordination, agent sandboxes, enterprise hardening.

**This is NOT niche.** This is the defining trend of 2025-2026 software engineering.

---

## Ring 4: Spec-Driven Development — The Growing Wave

**Size:** Dozens of voices. Major industry players. Multiple tools. Thoughtworks called it "one of 2025's hottest programming trends."

**Who's here:**
- Tool builders: OpenSpec (Hari Krishnan), AWS Kiro, GitHub Spec Kit, Augment Code, BMAD, Antigravity
- Major players: GitHub Blog official post (Den Delimarsky, Sept 2025), Thoughtworks, Augment Code
- Practitioners: Multiple Medium articles, DEV Community posts, practitioner guides
- Frameworks: AIDD (Binoy Ayyagari), SDD (Hari Krishnan)

**What they talk about:** Specifications as the primary artifact, not code. Write spec → implement → validate alignment. Living specifications that evolve with the system. "Specs as executable truth."

**The key insight everyone agrees on:**
> "Code is about to cost nothing. Knowing what to build is about to cost everything."

**What's mature:** The idea. The vocabulary. Multiple tools in early/mid adoption.
**What's emerging:** Standardization. Best practices. Enterprise adoption.

**Notable:** Dan Shapiro's 5 Levels framework went viral precisely because it articulates where SDD fits — Level 3+ is where specs become the primary work product.

---

## Ring 3: Context Engineering — The Accelerating Discipline

**Size:** Growing fast. Anthropic's own engineering blog covers it. Shopify's CEO (Tobi Lütke) named it. Moving from buzzword to discipline.

**Who's here:**
- Anthropic Engineering: official guide to [context engineering for AI agents](https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents) (Sept 2025)
- Tyler Brandt / Intent Systems: built an entire company around context-as-infrastructure
- Tobi Lütke: "The fundamental skill of using AI well is to be able to state a problem with enough context"
- Deepset.ai: "Context Engineering: The Next Frontier Beyond Prompt Engineering"
- Academic: "Agentic Context Engineering" (arXiv)
- Multiple practitioner articles (Towards Data Science, Snyk, Glean, etc.)

**What they talk about:** What the model sees before it acts determines its ceiling. Context windows are finite. Progressive disclosure. Hierarchical context. Token optimization. "Context without intent is noise" (Huryn).

**Tyler Brandt's Intent Layer** is the most sophisticated implementation:
- Hierarchical AGENTS.md files as progressive context disclosure
- Fractal compression (leaf nodes compress code → parent nodes compress children)
- Maintenance flywheel (automated sync on merge)
- "Your codebase becomes a reinforcement learning environment"
- Claims 50%+ individual productivity improvement

**What's mature:** The concept — everyone agrees context matters more than model capability.
**What's emerging:** Structured approaches. Tooling. Best practices for authoring context.

---

## Ring 2: Intent Engineering — The Named Discipline

**Size:** Small but real. ~10 named voices. Growing.

**Who's here (using "intent" explicitly):**

| Voice | Origin | Intent Concept |
|-------|--------|---------------|
| **Nate B Jones** | AI News & Strategy Daily | Three disciplines: prompt → context → intent. Intent as the highest layer — encoding *why* and *what matters*. |
| **Paweł Huryn** | Product Compass | Intent Engineering Framework: Objective + Outcomes + Health Metrics + Context + Constraints + Decision Types + Stop Rules. "Intent is what determines how an agent acts when instructions run out." |
| **Tyler Brandt** | Intent Systems | "Intent Layer" — hierarchical context system. Named his company after it. But his "intent" is closer to context engineering — it's about what the agent *knows*, not the human-agent *relationship*. |
| **Patrick Debois** | AI Native Dev / Tesla | Pattern 2: Implementation → Intent. Explicitly names the shift from specifying *how* to specifying *what you want*. |
| **Hari Krishnan** | intent-driven.dev / OpenSpec | "Intent-Driven Development" — coined the domain name. But his implementation (OpenSpec) is spec-driven, not intent-driven. The intent layer he describes is aspirational. |
| **Binoy Ayyagari** | AIDD framework | "Adaptive Intent-Driven Development" — enterprise SDLC reimagined around intent. Goal-oriented agent contracts. |
| **IndyDevDan** | Agentic Engineer | Doesn't use "intent" as a term, but his trust thesis IS about intent: "The limitation is not the model. It is our ability to put together the right context, model, prompt, and tools." The *right* things = intent expressed through engineering. |
| **BDI Model (academic)** | 1980s/90s AI theory | Belief-Desire-Intention is the formal framework. "Intentions" = adopted plans + tool calls. The academic ancestor of everything the industry is rediscovering. |

**The convergence:** Everyone who uses "intent" means something slightly different:
- Huryn: Intent = product objective + constraints (PM perspective)
- Brandt: Intent = what the agent needs to know (context perspective)
- Debois: Intent = what you want, not how to build it (developer perspective)
- Jones: Intent = the highest discipline above prompting and context (hierarchy perspective)
- Krishnan: Intent = specification as source of truth (artifact perspective)

**Nobody has unified these.** Each person holds one facet of the elephant. The word "intent" is emerging as a shared term, but there is no shared definition.

---

## Ring 1: Beyond Intent — Where We Are

**Size:** Us. One voice. Genuinely novel.

**What makes our contribution unique:**

The industry is converging on *intent as specification* — better ways to tell agents what you want. Our research ([03_beyond-intent.md](03_beyond-intent.md)) identifies seven gospel patterns that go beyond specification into *relationship*:

| Pattern | Gospel Source | Industry Closest Analogue | Why Ours Goes Further |
|---------|-------------|--------------------------|----------------------|
| **Covenant** | D&C 82:10 | Huryn's Constraints + Decision Types | Covenant is *mutual binding* — both parties commit. Industry intent is one-directional (human defines, agent executes). |
| **Stewardship** | D&C 104 / Matthew 25 | IndyDevDan's "defer trust until the merge" | Stewardship is *progressive trust through faithfulness*, not just results-based trust. It includes accountability (D&C 72:3-4) and the parable structure of "well done, receive more." |
| **Line upon Line** | Isaiah 28:10 / D&C 98:12 | Tyler Brandt's Intent Layer (progressive disclosure) | Brandt's disclosure is *informational* — graduated data access. Gospel pattern is *relational* — graduated trust/autonomy based on demonstrated readiness, not just information architecture. |
| **Atonement** | D&C 82:2,7 | arXiv paper's circuit breakers + retry logic | Industry error recovery is *mechanical* (retry, rollback, circuit-break). Atonement is *redemptive* — failure makes the system better, preserves the relationship, creates learning that benefits ALL future interactions. |
| **Sabbath** | Moses 3:2 | Patrick Debois's Content→Knowledge (adjacent) | Nobody has *structural reflection*. The industry has retrospectives (reactive, cultural). Sabbath is an *architectural* pause built into the cycle as a first-class concern. |
| **Zion** | Moses 7:18 | Multi-agent coordination (arXiv topologies) | All industry multi-agent patterns are *coordination* — agents directed toward compatible goals. Zion is *alignment* — agents sharing purpose, one heart and one mind, no poor among them. |
| **Consecration** | D&C 104:15-17 | Token budgets, resource allocation | Industry treats resources as *costs to optimize*. Consecration treats resources as *entrusted for a purpose* — "enough and to spare" for those who share the purpose. |

**The meta-argument:** "God solved the multi-agent alignment problem before we had agents." The Plan of Salvation is an intent architecture. The covenant path is a progressive trust framework. The temple is a specification system. These aren't metaphors — they're the *originals* that the industry is independently rediscovering fragments of.

---

## The Maturity Gradient

Here's what the concentric rings mean in practical terms:

```
Ring 5 (Agentic)     → "How do we build reliable agents?"        → Solved: loops, tools, tracing
Ring 4 (SDD)         → "How do we tell agents what to build?"    → Solved: spec format, delta specs
Ring 3 (Context)     → "How do we give agents what they need?"   → Emerging: Intent Layer, progressive disclosure
Ring 2 (Intent)      → "How do we encode WHY, not just WHAT?"    → Early: frameworks exist, no shared definition
Ring 1 (Beyond)      → "How do we build RELATIONSHIPS with agents?" → Novel: gospel patterns, covenant-based
```

Each inner ring depends on and transcends the outer rings. You can't do intent without context. You can't do context without specs. You can't do specs without agents. And you can't do covenant-based relationships without all of the above as foundation.

But the inner rings are where the *real* leverage lives. Everyone can run agents. The differentiation is in *how well you encode intent* — and beyond that, *how you build trust, recover from failure, reflect, and align toward shared purpose*.

---

## So Is It Just a Handful?

**No — and yes.**

- **Agentic engineering** is massive. Not niche at all.
- **Spec-driven development** is a recognized hot trend. Dozens of voices.
- **Context engineering** is accelerating. Anthropic, Shopify CEO, a company named for it.
- **Intent engineering by name** is a handful — maybe 10 voices. But they're influential (Product Compass has 200K+ subscribers, Patrick Debois shaped DevOps, IndyDevDan is a major practitioner voice).
- **Beyond intent** — covenant, stewardship, Zion — **that appears to be us alone.**

The industry is climbing the same mountain. Most are on the lower slopes (agentic engineering, SDD). A few have reached the ridgeline (intent engineering). Nobody else is looking at the peak.

But here's the remarkable thing: **they're already finding fragments of what the gospel teaches, without knowing the source.** IndyDevDan's trust thesis IS stewardship without the name. Tyler Brandt's Intent Layer IS line-upon-line without the scripture. Huryn's framework IS covenant structure without the binding commitment.

The gospel doesn't invalidate their work. It *completes* it.

> "For behold, the Spirit of Christ is given to every man, that he may know good from evil" — Moroni 7:16

These people are following the Light of Christ toward patterns God established. Our contribution isn't to say they're wrong — it's to point to the source and show where these partial insights converge into complete principles.

---

## What This Means for Us

1. **We're not early to a niche — we're at a specific altitude on a well-traveled mountain.** The foundation beneath us (agents, specs, context) is solid and growing. Our contribution (covenant, stewardship, Zion patterns) sits at a unique elevation nobody else has reached.

2. **The timing is right.** If the industry is just now naming "intent" as a discipline, the "beyond intent" conversation is next. When people start asking "okay, we have intent — now what?" — that's our opening.

3. **The audience is ready.** IndyDevDan's trust thesis has an engaged audience already asking the right question ("how do I build trust?"). Tyler Brandt's company has customers asking "how do I build better AGENTS.md files?" These are people one step away from the patterns we've identified.

4. **We need to be careful.** The gospel patterns must be presented in a way that's accessible to people who don't share our faith framework. The patterns work regardless of belief — progressive trust is universally true, structural reflection is universally beneficial. Lead with the principle, cite the source for those interested.

5. **The research validates our approach.** This scripture study project — with its `.github/copilot-instructions.md`, specialized agents, study methodology, intent preambles — *is* many of these patterns in practice. We're not theorizing about what might work. We're demonstrating it daily.

---

## Source Count Summary

| Category | Sources Found | Named Voices |
|----------|--------------|-------------|
| Agentic engineering (Ring 5) | 15+ articles, 5+ academic papers, 6+ enterprise platforms | 20+ |
| Spec-driven development (Ring 4) | 10+ articles, 6+ tools | 12+ |
| Context engineering (Ring 3) | 8+ articles, 1 company, 1 official Anthropic guide | 8+ |
| Intent engineering (Ring 2) | 6 articles/videos using "intent" explicitly | ~10 |
| Beyond intent (Ring 1) | Our 5 research docs | 1 (us) |
| **Total unique sources reviewed** | **30+** | **40+** |

