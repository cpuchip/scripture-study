---
title: claude-worker dispatch — premium models as CLI workers + the model-access map
date: 2026-06-06
status: DESIGN-ONLY — awaiting Michael's ratification. Engine lean (2026-06-07): `claude -p` (model A) to start; see §3a.
binding_question: >
  How do we hand more autonomous work to Claude (and other premium models)
  without burning Michael's interactive budget — and how does pg-ai-stewards
  orchestrate the mix of premium + cheap model access cost-effectively?
---

# claude-worker dispatch

> All dollar figures are as researched 2026-06-06 and move fast. **Confirm exact
> credit amounts in the Anthropic / OpenAI / provider dashboards before spending.**

## 1. Why now — the token economics changed

**Claude Code splits billing into two buckets on 2026-06-15** ([Anthropic
help](https://support.claude.com/en/articles/15036540-use-the-claude-agent-sdk-with-your-claude-plan)):

- **Interactive pool** (unchanged) — claude.ai, Claude Code *interactive* in the
  terminal, Cowork. **Our together-work (book, study, design sessions) lives here.**
- **Agent SDK credit pool** (new) — `claude -p` headless, the Agent SDK, GitHub
  Actions, third-party SDK apps. A **separate monthly dollar credit, billed at API
  rates, no rollover**: reported **$20 Pro / $100 Max-5x / $200 Max-20x**.

So a $200 Max plan yields a **dedicated ~$200/mo headless budget that does not
touch interactive sessions.** If we don't build a dispatcher, that ~$100–$200 is
**wasted every month**. This spec is how we spend it well.

⚠ Cautionary tale: `claude -p` has run up [$1,800 in 2 days](https://github.com/anthropics/claude-code/issues/37686)
when it spilled past the pool to API-key billing. **Spend-guarding is mandatory.**

## 2. The organizing principle — two access patterns, two roles

| Pattern | Who | How the substrate uses it | Role |
|---|---|---|---|
| **Sub WITHOUT a clean API** | Claude Max (agent-SDK pool), ChatGPT/Codex (gpt-5.5) | dispatch to its **CLI harness** (`claude -p`, `codex exec`, `agy -p`) | **premium CLI workers** — judgment, full-context, voice |
| **Sub/plan WITH a real OpenAI/Anthropic-compatible API** | opencode_go, GLM/Z.ai, Atlas, Ollama, … | add as a **model connector** inside pg-ai-stewards | **rank-and-file** — volume, mechanical, verifiable |

The mistake to avoid: trying to make a premium *sub* behave like a raw API
connector (the undocumented-backend route). It's gray/fragile/ToS-risky and, for
Codex, doesn't even carry gpt-5.5 yet (see §7).

## 3. Architecture — dumb poller + `claude -p` on demand

```
pg-ai-stewards (Docker)                host (Windows, Michael's logged-in Claude)
  work_items assigned to "claude"  ←──  [dumb host poller]  (plain PS/Go loop, NO AI, free)
        ▲                                      │  on new claude-item:
        │  writeback (output / status)         ▼
        └──────────────────────────────  claude -p "<binding question + context>"
                                               │   in the repo dir (Agent-SDK pool $)
                                               ▼
                                          capture output → mark done / escalate
```

- The bridge is in Docker and **can't reach the host `claude` CLI**, so a host-side
  component is required regardless.
- The poller is **non-AI and free** → **zero Claude tokens at rest**; the agent-SDK
  pool is only touched when real work exists. (Avoids the BoM-walk "idle burn" trap.)
- Reuses the existing `work_item_*` machinery — no substrate schema change needed for
  a v1, just an assignee/label convention (`assignee = claude-code`).

## 3a. Execution engine — three models (`claude -p` · long-lived session · Agent SDK)

The harness can be driven three ways. They differ on *who drives*, *which budget*, and
*warm vs cold*. **Michael leans (A) `claude -p` to start.**

**A. `claude -p` per task** *(the §3 architecture; the lean for P1).* The host poller
shells out a fresh headless run per work item. Full harness auto-loads (CLAUDE.md,
skills, `.mcp.json`, memory, subagents); **stateless/cold each task** (no cross-task
continuity); **draws the Agent-SDK credit pool** (the dedicated ~$200). Best for
independent, bounded jobs. Zero new code. ← **P1.**

**B. Long-lived interactive session — Michael's "node in the brain."** A *normal*
(non-`-p`) `claude` session living at the workspace root or a subproject, kept awake by a
timer / the `/loop` skill / a terminal script. Each tick it polls pg-ai-stewards
(`work_item_list`), claims an item, works it **in-session**, and pushes the result back.
**Warm + continuous** — accumulated context + memory; it "takes better account of you in
the brain," reasons across tasks, feels like a persistent collaborator. **Cost caveat
(important): a normal interactive session draws the INTERACTIVE pool, NOT the Agent-SDK
pool** — so it competes with together-time, not the dedicated $200. Also serial (one task
at a time) and needs compaction discipline (context rot over a long life). Buildable
*today* on `/loop` + the `work_item_*` MCP tools, no new code — but spend it knowingly.

**C. Agent SDK (programmatic).** A library (`@anthropic-ai/claude-agent-sdk` /
`claude-agent-sdk`) where *an application* drives the same engine: structured tool-use
events, session capture/resume/fork, in-process hooks, error handling, dynamic context.
For the **production control plane** — when pg-ai-stewards needs to (1) capture structured
tool-events for the cockpit's `cost`/`watch`, (2) enforce bin-1/2 guardrails *in code*
(intercept tool calls, not just trust the prompt), (3) manage per-work-item session
lifecycle. **Auth trap:** the SDK defaults to `ANTHROPIC_API_KEY` (Platform pay-per-token);
to draw the Agent-SDK pool it must be **plan-authed**, or you're back to the $1,800
overage. This is Anthropic's own end-of-path (*explore in Code → prototype → extract
prompts → embed in the SDK*).

**Billing summary:** A → Agent-SDK pool. **B → interactive pool** (the surprise). C →
Agent-SDK pool *if plan-authed*, else Platform API. And the Agent-SDK credit is
**per-user, not pooled** — a capacity ceiling if parallel workers ever run off one account.

**Design principle — engine-agnostic contract.** Keep the dispatcher contract
`{work_item, repo, binding_question} → {result | PR}` independent of the engine, so it
swaps under us. **Start A (`claude -p`).** Keep **B** as the warm-continuity option for
work that benefits from a persistent in-the-brain collaborator (spending its interactive
cost knowingly). Reserve **C** for the control plane once the cockpit needs structured
observability + programmatic guardrails. (Refs: [Augment — Code vs Agent SDK](https://www.augmentcode.com/tools/claude-code-vs-claude-agent-sdk),
[Agent SDK overview](https://platform.claude.com/docs/en/agent-sdk/overview).)

## 4. Work-item contract

1. A work item destined for Claude carries `{binding_question, repo/dir, context,
   acceptance, bin}` and `assignee=claude-code`. Created by a human or proposed by
   the substrate (the "bishop moment").
2. Poller claims it (atomic), shells out: `claude -p "<prompt>" --output-format json`
   in the target dir (loads CLAUDE.md + skills + memory).
3. Worker writes its result back: marks the item `done` with output attached, **or**
   raises an escalation for Michael's review (bin-dependent — see §5).
4. Failures degrade loud (re-probe cadence per [[feedback_unattended_run_resilience]]),
   never silently retry-spin.

## 5. Guardrails — what makes unattended-Claude safe

The autonomy governance already exists as skills — `ammon`, `dave-rule`,
`stuffy-in-the-loop`. A `claude-worker` skill activates them and enforces:

- **Unattended = bins 1–2 ONLY** (gather / verify / draft / build-reversible).
  **Never the Hinge.** Anything bin-3+ → produce a draft/PR and **escalate for
  Michael**, do not commit/merge/deploy.
- **Model tiering:** run the dispatcher on **Sonnet/Haiku for volume**, escalate to
  **Opus only for judgment-heavy items** (mirrors the substrate's own kimi-vs-critic
  tiering; conserves the capped pool).
- **Spend guard:** track agent-SDK pool burn; hard-stop at a configurable ceiling
  well under the monthly credit; **never fall back to API-key billing** (the $1,800
  trap). Log spend loudly (the BoM-walk "ate 25% before I noticed" lesson).
- **Scope-fence:** allow-list dirs/repos + tools per work item, like the coder sandbox.

## 6. Budget model (the dual-bucket, applied)

- **Interactive pool** → reserved for *together-work*: study, the book, design,
  ratifications, hard debugging with Michael present.
- **Agent-SDK pool ($200 on Max-20x)** → autonomous Claude missions via the
  dispatcher. Finite + no rollover, so treat it as a monthly allowance to *spend
  down deliberately* (e.g. a canon-walk shepherding budget, overnight cross-repo
  refactors with tests as ground truth, study drafts for later review).
- **Substrate model budget** (opencode_go + any added connectors, §8) → the rank-
  and-file volume that never needs Claude at all.

Three wallets, three jobs. The dispatcher is what converts wallet #2 from "wasted"
into "autonomous Claude."

## 7. gpt-5.5 / Codex — the clean routes (avoid the gray one)

- ChatGPT/Codex sub **does NOT cover raw API calls** ([OpenAI](https://help.openai.com/en/articles/11369540-using-codex-with-your-chatgpt-plan),
  [ToolColumn](https://www.toolcolumn.com/learn/chatgpt-subscription-vs-openai-api-pricing)).
- The `opencode-openai-codex-auth` plugin uses sub-OAuth against an **undocumented
  Codex backend**, is **"personal use only — no commercial/multi-user,"** is
  **Node-only** (not opencode_go), and **does not support gpt-5.5** (only 5.2/5.1).
  **Do not wire it into the substrate** — account-ban risk, and it doesn't even do 5.5.
- **Clean route A** — gpt-5.5 as a substrate connector via a real **OpenAI API key**
  (Platform billing, ~$5/1M in, $30/1M out). Official, stable, drops into the
  connector model; just separate $.
- **Clean route B** — gpt-5.5 on the **Codex sub** via **`codex exec` headless as a
  dispatched CLI worker** (same shape as `claude -p`). Official, uses the sub's
  included limits. Sequence *after* the Claude dispatcher proves the contract.

## 8. Model-access map — subs WITH a real API (substrate connector candidates)

Research answer to "another provider like opencode_go": **yes, a whole category**
("AI coding plans" / "API gateway plans") — flat-rate subs exposing OpenAI/Anthropic-
compatible endpoints, drop-in as pg-ai-stewards connectors. Adding a 2nd provider buys
**redundancy** (when opencode_go 529s — the resilience lesson), **more daily capacity**
(each sub has its own quota), **per-key budget isolation** (parallel agents), and
**model variety**.

| Provider | Price | Models | API | Note |
|---|---|---|---|---|
| **opencode Go** (current) | $10/mo | GLM/Kimi/Qwen/MiniMax/DeepSeek (broadest list) | OpenAI-compat | already in use |
| **GLM Coding Plan (Z.ai)** | $10 / $30 / $80 | GLM-5.1 (~94% Opus on coding) | **Anthropic-** + OpenAI-compat | strong single-vendor value |
| **Atlas Cloud Coding Plan** | $10 (800K cr/day) / $20 (1.8M) | DeepSeek/Kimi/GLM/MiniMax/Qwen | OpenAI-compat | **per-key isolation**, opencode-friendly |
| **CheapestInference** | ~$10+ | Qwen/Kimi/GLM/MiniMax/DeepSeek | OpenAI-compat | no waitlist, per-key budgets |
| **Ollama Cloud Pro** | **$20/mo** (Max $100) | gpt-oss + open weights | OpenAI-compat | **metered by GPU-time, not tokens**; 3 concurrent |
| **Featherless / Awan / NaN** | $25 / flat / €70 | thousands / open weights | OpenAI-compat | true flat-rate, no token meter |

Caveats: all have rate caps (req/5h, req/day, concurrency) — "generous," not infinite;
tiers can close to new users (Qwen Lite closed Mar 2026); several are China-hosted
(data-residency) — Featherless/NaN emphasize no-logging. **Recommendation:** add **one**
second connector for redundancy + capacity — **Atlas Cloud** (per-key isolation suits
multi-agent) or **GLM/Z.ai** (Anthropic-compat + strongest single model). **Ollama Pro
$20** is the open-weights/GPU-time option if we want a different cost model for long-context.

## 9. Build phases (cheap-first)

- **P1 — host-poller proof** (host poller is non-Claude code; near-free): poll a
  `claude-code`-assigned work_item, run `claude -p`, write result back. Prove the loop.
- **P2 — work-item contract + `claude-worker` skill** (bins, scope-fence, escalation).
- **P3 — spend guard + tiering** (pool ceiling, Sonnet-volume/Opus-judgment, loud logging).
- **P4 — generalize to premium CLI workers** (`codex exec`, `agy -p`) + add **one**
  sub-with-API connector (§8) to the substrate for redundancy.
- **P5 (later)** — substrate proposes its own claude-work (the autonomous bishop moment).

## 10. Open questions for Michael

1. Execution engine (§3a): **(A) `claude -p` push** vs **(B) long-lived interactive
   `/loop` session** (warm, but interactive-pool) vs **(C) Agent SDK** (control plane).
   **Decided lean (Michael, 2026-06-07): A to start; B as the warm option; C later.**
2. Which second sub-with-API connector to add first (Atlas / GLM / Ollama)?
3. Upgrade to Max-20x **before or after** P1 proves the loop? (Lean: prove with the
   current plan's pool first, then upgrade once the loop demonstrably uses it.)
4. Codex sub now, or defer gpt-5.5 to P4?

## Sources
- Anthropic Agent SDK billing — https://support.claude.com/en/articles/15036540-use-the-claude-agent-sdk-with-your-claude-plan
- Claude Code vs Agent SDK (when to use each) — https://www.augmentcode.com/tools/claude-code-vs-claude-agent-sdk
- Agent SDK overview (capabilities, auth) — https://platform.claude.com/docs/en/agent-sdk/overview
- claude -p headless mode — https://www.mindstudio.ai/blog/claude-code-headless-mode-autonomous-agents
- June 15 change explainer — https://codersera.com/blog/anthropic-june-2026-billing-change-claude-code/
- `claude -p` $1,800 overage — https://github.com/anthropics/claude-code/issues/37686
- Codex with ChatGPT plan — https://help.openai.com/en/articles/11369540-using-codex-with-your-chatgpt-plan
- ChatGPT sub vs API billing — https://www.toolcolumn.com/learn/chatgpt-subscription-vs-openai-api-pricing
- opencode-codex-auth plugin — https://numman-ali.github.io/opencode-openai-codex-auth/
- opencode Go — https://opencode.ai/go
- AI coding plans comparison — https://codingplan.run/ · https://www.atlascloud.ai/blog/guides/coding-plan-best-ai-coding-subscription-under-20
- Ollama Cloud pricing — https://ollama.com/pricing
- Featherless / Awan / NaN — https://featherless.ai · https://www.awanllm.com · https://justuse.nan.builders
