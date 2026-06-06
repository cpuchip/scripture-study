---
title: stewards cockpit — a human CLI for pg-ai-stewards (Option A)
date: 2026-06-06
status: DESIGN-ONLY — awaiting Michael's ratification (focus item, per 2026-06-06)
binding_question: >
  How does Michael drive the substrate directly — dispatch work, watch it run, approve
  at the Hinge, see cost — without going through Claude, becoming the manager-of-agents
  himself?
---

# stewards cockpit CLI (Option A)

> Closes the #1 gap from `docs/ai-utilization-landscape-2026.md`: pg-ai-stewards is
> useful to the agent, less to Michael. The field's winning tools are **human cockpits**
> for supervising the fleet. This is ours, terminal-first.

## 1. Why this, why now

Michael's operating model (ratified direction 2026-06-06): **spec + ratify a backlog
ahead → Claude executes async while he's at work / asleep → his time-with-Claude is
spent planning, counciling, ratifying.** That model needs a cockpit so he can *drive*
between planning sessions. It composes exactly with the billing reality:

- **Interactive pool** → time *with* Claude (plan / council / ratify).
- **Agent-SDK pool ($200 on Max-20x)** → async Claude work he kicks off from the cockpit
  ([[claude-worker-dispatch]]).
- **Substrate model budget** → the rank-and-file volume.

The cockpit is the human entry point to the whole stewardship tree. The $200 lean is
now justified by a concrete plan, not a guess — see §6.

## 2. Shape

A `stewards` **Go CLI** (matches the workspace; opencode/claude-code-style ergonomics)
that talks to the substrate via a **pgxpool to the substrate Postgres** (host
port 55433) — the same pattern `cmd/persona-host` already uses. It calls the *existing*
`work_item_*` SQL functions, cost tables, and brain tables. **No new engine; a human
front-end over what's there.** (Read paths are pure SQL; dispatch reuses
`work_item_dispatch_stage` etc.)

## 3. Verbs (starter set — Michael to confirm the daily 4–6)

| Verb | Does | Reuses |
|---|---|---|
| `stewards project [name]` | show / switch the **active project** (ai-chattermax, pg-ai-stewards, scripture-book, book, study, …); a sticky context that scopes the work-item verbs | work_items.project |
| `stewards board` / `ls` | the project board for the active project — items by planning state (`--all` spans every project) | work_items + new planning dims |
| `stewards do "<binding question>" [--repo --pipeline]` | create + dispatch a work item | `work_item_*` dispatch |
| `stewards council <item>` | convene a critical-analysis pass on a plan/spec — cheap models + a critic surface tensions, connections, blind spots *before* you ratify (the Abraham 4:26 council moment, automated) | `consult_subagent` / `panel_redline` / critic |
| `stewards ratify <item>` | the **input Hinge** — approve a spec to build (`planning_state` spec→ratified), making it eligible for async execution | work_items planning_state |
| `stewards watch [id]` | live tail of a pipeline / item (stages, status, cost) | work_item status + cost |
| `stewards review [id]` | the **output Hinge** — escalations / finished PRs awaiting approval; approve / reject (building→done) | `work_item_escalation_*` |
| `stewards cost [--by project\|model\|day]` | token + $ spend dashboard (§5) | cost tables |
| `stewards personas` / `chat <persona>` | list + talk to a persona | persona pipelines |
| `stewards brain <query>` | search the brain | brain tables |

**Two Hinges, two ends.** `ratify` is the *input* Hinge — a plan approved to build.
`review` is the *output* Hinge — finished work approved as good. `council` is the
deliberation that *informs* ratify (it surfaces; the human decides — Webster 1828
"counsel": interchange of opinions, mutual advising).

**Active project** is a sticky context (stored in a local `.stewards` config — like a
`kubectl` context or the current git branch) that scopes the work-item verbs
(`board`/`do`/`cost`/`watch`) to one project. `--project X` overrides per-command;
`--all` spans every project. Daily set: **project / board / do / council / ratify /
watch / review / cost.**

## 4. Project board (the "keep track of everything" ask)

Michael wants pg-ai-stewards to be **a larger project board** that tracks everything
we're working on *and its planning state*. Today `.spec/carry-over.md` is that board,
maintained by hand. Elevate it into the substrate:

- Add two dimensions to work items: **`project`** (ai-chattermax, pg-ai-stewards,
  scripture-book, book, study, …) and **`planning_state`** (`idea → spec → ratified →
  building → blocked → done`).
- `stewards board` renders it (group by project, filter by state). `carry-over.md`
  becomes a **generated view**, not hand-maintained — single source of truth in the DB.
- Claude (and the cockpit) read/write it through the same work_item surface, so the
  backlog, the dispatch queue, and the planning board are *one* system.

This is what turns "a pile of specs" into a board Michael can actually steer from.

## 5. Token dashboard by project + model (Michael's add)

A shared cost aggregation — **spend grouped by project × model × day** — surfaced two
ways: `stewards cost --by project|model` (CLI, Option A) **and** a panel in
`stewards-ui` (Option B). Backend is one query over the existing per-run cost tracking;
both front-ends read it. Makes the BoM-walk "ate 25% before I noticed" problem visible
*before* it happens. (Should also fold in the Claude agent-SDK pool burn once
[[claude-worker-dispatch]] lands, so all three wallets are in one view.)

## 6. How the $200 / async model coheres

The pieces now line up into one loop:
1. **Together (interactive pool):** explore → plan → council → ratify a backlog (the
   project board fills with `ratified` items).
2. **Apart (agent-SDK pool + substrate):** Michael kicks ratified items from the cockpit
   (`stewards do` / assign-to-claude); Claude works async while he's at work/asleep;
   the substrate's cheap models do volume.
3. **Back together:** `stewards review` surfaces the Hinge queue; he approves/merges; the
   board advances to `done`.

The cockpit is what makes steps 2–3 his, not mine. Without it, he can only drive via me.

## 7. Build phases (cheap-first)

- **P1 — read-only cockpit:** `board` / `watch` / `cost` (pure SQL reads). Immediately
  useful, zero risk. Proves the pgxpool + verbs.
- **P2 — project-board dims:** add `project` + `planning_state`; generate `carry-over.md`
  from the DB.
- **P3 — dispatch:** `do` (create + dispatch a work item from the terminal).
- **P4 — Hinge:** `review` (approve/reject escalations + PRs).
- **P5 — personas / brain:** `chat`, `brain`.
- **P6 — stewards-ui** gets the same cost dashboard + board (Option B), reusing P1/P2
  queries.

## 8. Open questions for Michael
1. Confirm the daily verb set (do / watch / review / cost / board — add personas/brain?).
2. Planning-state vocabulary — is `idea → spec → ratified → building → blocked → done`
   the right ladder, or do you want a `seed`/`backlog` rung?
3. CLI talks to the DB directly (pgxpool, simplest) vs through a small substrate HTTP
   API (cleaner boundary, more work)? (Lean: direct pgxpool for P1, like persona-host.)
4. Should `stewards do` be able to target a **Claude** work item (agent-SDK pool) from
   day one, or substrate-only first?

## 9. Relation to other specs
- [[claude-worker-dispatch]] — the cockpit's `do --assignee claude` kicks the agent pool.
- [[agentic-tools-model-cascade]] — `stewards do` can invoke agentic tools / the code persona.
- `docs/ai-utilization-landscape-2026.md` — this is Option A of the cockpit fork.
