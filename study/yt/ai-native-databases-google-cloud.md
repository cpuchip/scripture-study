---
title: "The database that thinks, said by Google — AI-native databases and the agentic Data Cloud"
source: "Power intelligent agents with AI-native databases (Google Cloud Tech, 2026-06-25)"
url: https://www.youtube.com/watch?v=7awKinJhGPo
speakers: Amit Ganesh (Google Cloud, ex-Oracle), Yannis Papaconstantinou (Google Cloud databases), David Soria (Anthropic — creator of MCP)
date_digested: 2026-06-27
tags: [pg-ai-stewards, ai-native-database, agentic, mcp, postgres, alloydb, substrate, convergence]
relevance: high — the enterprise-vendor restatement of the pg-ai-stewards thesis
transcript: yt/google-cloud-tech/7awKinJhGPo/
---

# The database that thinks, said by Google

A 50-minute Google Cloud keynote-session, June 2026, with Amit Ganesh (who came from Oracle), Yannis Papaconstantinou, and — in a fireside — **David Soria of Anthropic, the creator of the Model Context Protocol.** The whole thing is the enterprise-vendor version of the bet pg-ai-stewards has been built on across four brain-generations: *the database is where the intelligence should live.* It is worth digesting not because it teaches us something new, but because it tells us we have been standing on the right ground — and, read carefully, it tells us exactly where our ground is still ours alone.

## The thesis is ours

Amit opens by naming legacy data as "walled gardens of disconnected warehouses and siloed operational databases," and frames the move:

> "What if instead you could move AI to data and break down those walls." — [3:27]

Then he gives the category its name:

> "We are introducing a whole new category for the next decade, and we are calling it the agentic Data Cloud. This is a system of action for agents, not just a system of insights." — [4:31]

Yannis sharpens it to a definition that could be the first line of our own README:

> "An AI native database doesn't just store data, it natively processes and understands the data using built in AI primitives." — [9:28]

That is "the database thinks," with a marketing budget. And the substrate underneath it is the one we chose: AlloyDB is "100% PostgreSQL compatibility" with Google's additions on top, and AlloyDB Omni "runs anywhere on premise, on any Cloud, or even on your local laptop" [10:28, 11:03]. The clinching detail comes from Anthropic's own MCP creator:

> "we are at Anthropic use AlloyDB for example… and ground our workflows in the data that we have at these database systems." — [39:42]

So the pgrx-on-Postgres bet is not a niche eccentricity. Google launched a *category* on it; Anthropic runs on it. Independent convergence on the same noun — the same signal we got from the Google/Kaggle "substrate" whitepapers, now from the largest vendor on the planet.

## But the line that matters is the one they didn't cross

Read closely, Google puts the *AI primitives* in the database — vector search, hybrid search, AI functions callable from SQL, natural-language-to-SQL — and keeps the *agent runtime outside* it. Their agents (ADK, the Query Data tool, the prebuilt observability/testing agents) call *into* the database. The database is a very smart tool the agent reaches for.

pg-ai-stewards put the runtime *in* too. The dispatch loop, the work queue, the engrams, the personas, the A2A handoffs, the governance bounds — those are not services calling Postgres; they *are* Postgres objects. Their "agentic Data Cloud," read for what it actually describes, is request/response tool-serving plus some single-purpose prebuilt agents. There is no in-database autonomous work queue, no self-dispatching loop, no memory-as-tables, no agent-to-agent inside the database in anything they showed.

So the validation is real *and* it leaves the hard part — the autonomy layer — untouched. They put the primitives in the database; we put the runtime in too. That is the moat, and the keynote draws its own border around it for us.

## What to take

The talk's most useful gift is concrete technique, and one of them lands squarely on the security question we had been circling (OpenClaw, send-only, deterministic bounds):

- **Parameterized secure views as a deterministic guardrail.** The agent never names the base table; a per-principal view hands it *only the current user's rows*, so a prompt-injected query cannot widen access. Yannis: *"This is a deterministic security guardrail because the view provides to the agent only the current end user's data"* [27:50]. This is an *oracle-style* control — deterministic, not model-judged — which is precisely our "build the oracle first" doctrine applied to data access. For a substrate whose coder agents generate and run SQL with real grants, a per-principal secure-view layer is a deterministic wall worth adding.
- **Skills-over-MCP and long-running tasks** — straight from MCP's creator: *"we're going to make it possible to serve skills over MCP servers so that people that offer MCP servers can also offer domain specific knowledge… in the form of skills. In addition, we're going to work on Long running tasks, which is super interesting to really enable more agentic behavior"* [41:43]. We already have skills and long-running work_items as first-class database state. When that protocol surface lands, `stewards-mcp` is unusually well-positioned to *be the server* for it — a direct tie to the A2A standard-wrapper work. Track the spec.
- **Hybrid (vector + BM25) + a rerank pass** as the default retrieval recipe, not vector-alone [14:21, 16:10] — if our engram/doc search is vector-only, this is the validated production upgrade.
- **A cheap deterministic pre-filter before the expensive LLM judge** — their cost-optimized `AI.IF` predicate (claimed ~1000× cheaper than a baseline LLM call, [18:08]). The *pattern* maps onto our reaper/judge economics: push as much filtering as possible to a cheap in-database predicate before spending a dispatch.

## What to leave

The Ben-test reading, because a keynote is a sales document:

- **Lock-in dressed as openness.** "Use Postgres, bring your own model, open standards" sits right next to ScaNN (proprietary), TimesFM, Gemini-only AI functions, TPU inference, and Model Armor — and *the differentiating numbers all depend on the proprietary pieces.* The "100% Postgres compatible" surface is real; the differentiators are not portable. This is the exact trap the owned substrate exists to avoid.
- **Unverifiable superlatives.** "#1 on BIRD," "100x faster than standard Postgres," "near 100% accuracy," "1000x cost reduction," a "20% relevance gain" at Target — all vendor-asserted, methodology "in upcoming white papers." Directional marketing, not benchmarks.
- **A governance plane shaped like centralization.** Their security story (Model Armor, managed-server visibility, "which MCP servers have been activated") is genuinely strong, but it *presumes Google sits in the middle.* Adopt the technique (the secure view), never the topology (the surveillance plane).
- **"Agentic" inflation.** Much of what is branded agentic is natural-language-to-SQL plus SQL-callable AI functions. Don't let the framing imply they built an autonomous in-database runtime. They didn't describe one — and that absence is exactly our differentiator.

## What it means for us

The industry spent fifty minutes confirming the foundation we have been building on, and the part they are still missing is the part we actually built. The right response is neither triumph nor anxiety. It is two concrete follow-ups — **add the parameterized-secure-view wall** (a deterministic data bound that fits both our coder agents and the OpenClaw-shaped security posture) and **track the MCP skills/long-running-task surface** so the substrate is ready to serve it — and one steady fact to hold: when the biggest vendor alive launches a category on your thesis and the protocol's own creator runs on the same substrate you chose, the question is no longer whether the ground is real. It is whether you keep the part of it that is still only yours.
