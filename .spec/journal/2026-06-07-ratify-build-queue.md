---
date: 2026-06-07
title: Ratify the backlog, then build — the week's build queue set
workstream: pg-ai-stewards / operating-model
mode: plan/ratify (Sabbath-light)
tags: [ratification, cockpit, code-persona, operating-model, sabbath]
---

# Ratify the backlog, then build

Short session. Budget reset, ~9pm Sunday. Michael came in with "we're reset — what
should we ratify and work on?" and chose **"ratify the backlog, then I build."** The
shape of the night was deliberately Sabbath-light: **ratify tonight, build async through
the week.** No build marathon on a Sunday evening.

## What got ratified

The 2026-06-06 strategy session left five specs designed but un-ratified. Tonight two of
them flipped to **RATIFIED + build-ready**, with the open questions answered in an
AskUserQuestion batch:

**① stewards cockpit P1** (`stewards-cockpit-cli.md`):
- Verbs = `project / board / do / council / ratify / watch / review / cost`.
- Planning-state ladder = `idea → spec → ratified → building → blocked → done` (no
  separate `seed`/`backlog` rung — the early states already mean "tracked, no AI yet").
- Connection = **direct pgxpool** to the substrate Postgres (port 55433, like
  persona-host) for P1. No HTTP API boundary yet.
- Cards = **un-dispatched work_item**, one table. We explicitly did NOT add a separate
  `tracked_items` table — an item graduates to AI work in place rather than reconciling
  two systems.
- **Build P1 = read-only `project / board / watch / cost`** (pure SQL reads, zero risk).

**② code persona P1** (`agentic-tools-model-cascade.md`):
- Flagship repo = **ai-chattermax**.
- Scope = **read-only first** — `research_codebase` returns findings + `file:line`
  citations; no edits/PRs until it earns the trust. A propose-changes engineering persona
  is a later, gated step (drafts a PR, never merges — the Hinge).
- **Build P1** = a `researcher-flash` agent_family (deepseek-v4-flash + read-only repo
  tools) + the `research_codebase` MCP tool (a `consult_subagent` preset + return
  contract). Then P2 wires it into a read-only code persona in a room.

## Still pending (not ratified tonight)

- **claude-worker** — engine decided (**Model A = `claude -p`** to start), but the
  agent-SDK credit pool doesn't exist until **2026-06-15**, so Model-A async can't spend
  it before then; and the **second connector** (Atlas / GLM / Ollama for substrate
  redundancy) is still an open pick.
- **CT2** — read-then-ratify; it restarts the live substrate Starlet, so Michael reads
  the spec before I build.
- **Dave steal-list** — which of the five borrows (AI-Freedom section, invariant
  traceability, SideQuest lane, ODD/SRE debug depth, file-first discipline).
- **Harness-leveling experiment** — yes/no on running the A/B.

## How it was recorded

The discipline this session was *durability*, not building: the ratifications had to
survive into the build sessions that happen later in the week, when this context is gone.

- Spec frontmatter `status:` flipped to RATIFIED on both files (via the dogfooded
  `md-frontmatter-set` tool), plus an appended ratified-decisions block in each.
- `.spec/carry-over.md` restructured: a new **★ RATIFIED — build queue** section at the
  top (items ① and ②), the cockpit moved out of "needs decision," the agentic-tools seed
  marked ratified, the engineering-persona item marked ratified.
- Tasks **#133** (cockpit P1) and **#134** (code persona P1) created.
- Committed to root as `0e42269` — **not pushed** (root rule: Michael pushes the
  workspace root himself).

## Reflections / carry-forward

- The operating model we spec'd last session is now *live in its own shape*: tonight was
  pure plan/ratify on the interactive pool; the building it authorized runs cheap and
  async later. That's the whole point of the cockpit + claude-worker arc — Michael's
  time-with-Claude spent on judgment, the volume spent elsewhere.
- The natural first build is **cockpit P1** (read-only, zero risk, proves the pgxpool +
  verbs against the real substrate) before the code persona, which touches dispatch.
- Watch the Mosiah 4:27 line on the pending pile — four un-ratified items is fine as a
  *queue*; it's only a problem if it becomes four things in flight at once.
