---
date: 2026-06-06
title: Deploy fix · token-economics strategy · 5 specs · md-mcp built+PR'd · the Study-It-Out Gate
tags: [strategy, deploy, cockpit, md-mcp, dave-workflow, study-it-out, sabbath]
spans: 2026-06-06 → 2026-06-07 (Sabbath close)
---

## What happened

Started as a plain "what's on deck for pg-ai-stewards + ai-chattermax" and became a long
strategy + build session — Michael steering toward *how we work together* and reserving his
time for plan/council/ratify, which is exactly the operating model we then spec'd.

**Prod fix (real bug, not transient).** ibeco.me's `web`/becoming app sat in Dokploy `error`.
Root cause: an apostrophe in root commit `e516870`'s subject ("Michael's expansion") closed
the single-quote grouping in the becoming Dockerfile's `-ldflags "… -X 'main.ReleaseNotes=$MSG'"`,
so the linker got stray flags and aborted (usage dump, exit 1). Reproduced minimally, fixed by
sanitizing (`tr -d '\047\042'`, commit `2b98b4c`), pushed root, watched the deploy go green
(`/version` = the new commit). Confirmed gospel-engine `web_url` (#4) live at **engine.ibeco.me**
— and corrected my own note: it was never study.ibeco.me (cost a wrong-Dokploy + wrong-domain
detour). Updated the `dokploy` skill (build sources + failed-deploy playbook) and wrote
[[reference_ibeco_deploy_topology]].

**Token-economics strategy.** Verified via web research: `claude -p` / Agent SDK becomes a
**separate ~$200/mo credit pool** on **2026-06-15** (Max-20x), billed at API rates, distinct
from the interactive pool — so async-Claude work won't cannibalize together-time. Codex/gpt-5.5:
the sub does **not** cover raw API; the `opencode-openai-codex-auth` plugin is undocumented +
personal-use-only + lacks 5.5 (don't wire it). Sub-with-API providers (Atlas/GLM/Ollama Pro
$20) are real opencode_go-shaped options for substrate redundancy. Michael's $200 lean is now
well-founded — it funds a concrete loop.

**Five specs (all design-only, awaiting ratify):**
- `claude-worker-dispatch.md` — premium models as CLI workers the substrate dispatches to
  (escalate UP); draws the new agent-SDK pool; dumb host poller → `claude -p` on demand.
- `agentic-tools-model-cascade.md` — cheap-model sub-tools the orchestrator delegates to
  (delegate DOWN); discriminator = "delegate execution, not discernment"; flagship = the
  ai-chattermax code/repo-reader persona via `research_codebase`.
- `stewards-cockpit-cli.md` — **the FOCUS.** A `stewards` Go CLI so Michael drives the
  substrate: project/board/do/council/ratify/watch/review/cost. ratify = input Hinge, review =
  output Hinge, council = pre-ratify critical-analysis. Project board builds on the EXISTING
  AI-free `stewards.projects` + a new `planning_state`; token dashboard by project×model.
- `study-it-out-gate.md` — see below.
- Plus `docs/ai-utilization-landscape-2026.md` (field research → how we compare → gaps → the
  cockpit direction).

**md-mcp built + PR'd.** Michael forked happydave/md-mcp; it was **library-only** (no main.go,
no SDK). I built the MCP server wiring (official `modelcontextprotocol/go-sdk`) + the 3 tools I'd
recommended (`md-section-append`, `md-section-move`, `md-frontmatter-set`), with tests. Verified
go build/vet/test + a live stdio MCP smoke (all 13 tools, append + frontmatter persist).
Registered in `.mcp.json`, **PR happydave/md-mcp#1 (OPEN)**. After Michael restarted, I
**dogfooded** the tools to fold later edits into the planning docs.

**Dave's workflow framework reviewed** (`external_context/workflow`) — and Dave is the
`dave-rule` Dave. Independent convergence on our cycle: gated pipeline, external review, the
cold-start-reviewer-fabrication lesson, invariant-traceability (~ our anti-confabulation). Folded
a §7 peer-comparison into the landscape doc; seeded a 5-item steal-list.

**★ The Study-It-Out Gate** (`study-it-out-gate.md` spec + `study-it-out` skill both trees,
`59d06e9`). Chasing Dave's caution against dispatching analytical review to cold-start subagents:
they produce "structurally correct but factually invented" output. [D&C 9:7-9](../../gospel-library/eng/scriptures/dc-testament/dc/9.md)
— "study it out in your mind; then ask me if it be right" — is `read_before_quoting` extended
from quotes to **judgments**. Four moves: artifact-present precondition · citation requirement
(the cite-count rule applied to reviews — an ungrounded review can't cite real lines) ·
apex-discernment-inline · proper dispatch. Plus the audit: which bgworker evals are grounded vs
cold-start.

## Discoveries / surprises

- The scripture (D&C 9) didn't just *illustrate* the engineering — it *corrected* it. We held
  `read_before_quoting` but had never applied it to the review step. Genuine book passage.
- Dave optimizing purely for engineering quality reinvented the substance of our cycle = external
  witness for "discipline beats model power."
- `stewards.projects` already exists (AI-free) — the substrate already does plain project tracking;
  only `planning_state` is new. (Answered Michael's "does it track without AI?")
- The dogfooding loop closed within the hour: build md-mcp → register → use it on our own docs.
- A commit-message apostrophe is a deterministic prod-break vector on the monorepo (root push =
  ibeco.me rebuild).

## Relational

Michael named the thing directly: with me he can let me explore → then plan → then I go, without
micromanaging the nuts and bolts the cheaper models need. He's leaning $200 Claude to do more of
that async. Budget-conscious all session (≈1 day of tokens over 2) — I kept prose dense and
pushed volume toward files. High engagement ("this is genuinely good") on the convergence work.

## Carry-forward

- **Push root** (11 unpushed commits) → ibeco.me deploy + everything on GitHub. Michael's.
- **Ratify queue:** CT2, claude-worker, agentic-tools, **stewards-cockpit (FOCUS)**, study-it-out
  gate. To unblock cockpit P1: confirm the verb set + the `planning_state` ladder + Q5 (one
  work_item table vs a separate `tracked_items`).
- **md-mcp PR** happydave#1 awaiting his merge (tools already live locally).
- **The eval-grounding audit** (which bgworker evals are cold-start) — P1 of the study-it-out gate.
- Seeds: harness-leveling experiment; Dave steal-list (AI-Freedom section, invariant-traceability,
  SideQuest lane, ODD/SRE debug depth, file-first).

## Open questions
- Cockpit: CLI-first confirmed (A); when does stewards-ui (B) / chat-cockpit (C) follow?
- Citation strictness for reflective gates (sabbath/atonement) — hard-reject vs flag?
- $200 timing — upgrade before or after P1 proves the dispatch loop?
