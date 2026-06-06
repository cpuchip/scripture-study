---
title: Agentic tools / model cascade — cheap-model sub-tools the orchestrator delegates to
date: 2026-06-06
status: DESIGN-ONLY — awaiting Michael's ratification
flagship: the ai-chattermax code / repo-reader persona
binding_question: >
  Can pg-ai-stewards let an orchestrating model treat *other, cheaper models* as
  tools — delegating the grind (search, read, compile, edit) downward and keeping
  only judgment — so heavy models don't burn tokens on light work, and a chat
  "code persona" can answer questions about a repo affordably?
---

# Agentic tools / model cascade

## 1. The idea, and why now

Some "tools" should be **cheap-model agents**, not deterministic code. The
orchestrator says *"find everywhere we handle auth and summarize the patterns"* or
*"update the changelog from these commits"*; a cheap model (e.g. deepseek-v4-flash)
does the legwork with real tools and returns a **curated result**. The expensive
model spends tokens on *deciding and judging*, not on grinding through 40 files.

**Flagship use — the ai-chattermax code / repo-reader persona** (carry-over item).
A chat persona that answers "how does auth work in ai-chattermax?" should NOT have
its orchestrator read the whole repo every turn. It should call an agentic
`research_codebase` tool → a cheap model explores + curates → the orchestrator
composes the answer. That's the repo-reader persona done right *and* cheaply.

This is the **mirror** of [[claude-worker-dispatch]]: that escalates work UP to
Claude; this delegates work DOWN to cheap models. Together they form the full
stewardship tree — **cheap models ← orchestrator ← Claude ← Michael** — each
delegating execution downward and escalating discernment upward.

## 2. What already exists (extension, not invention)

The orchestrator-worker / model-cascade pattern is established (Claude Code's
Explore subagent; WebFetch using a small model to read+digest a page). pg-ai-stewards
already has the primitives:

- `spawn_subagent` (async) / `consult_subagent` (sync) — orchestrator spawns a
  scoped child with its own agent_family (model + tools + system prompt via
  `compose_*`).
- `summarize_url`, `deep_research`, `panel_redline` — tool-calls that ARE model calls.
- The coder sandbox + per-wi worktrees (CV2.1) — repo-mounted deterministic tools
  (grep/glob/read/apply_patch/build/test).

So an **agentic tool = a named, purpose-built `consult_subagent` preset** exposed
as a first-class MCP tool: a fixed cheap-model agent_family + scoped inner tools +
a return contract + a clean signature. The engine is built; this adds presets and
thin wrappers.

## 3. The discriminator — when to agentic-wrap (the design heart)

Wrap a subtask as an agentic tool **only when ALL three hold**:

1. **Language/judgment task** (not mechanical/exact).
2. **Large relative to its instruction** (the delegation handoff must be cheaper
   than doing it inline).
3. **Cheap deterministic verification exists** (so the orchestrator can *trust the
   result without fully re-reading it* — else the verification tax eats the saving).

| Candidate | Verdict | Why |
|---|---|---|
| "find all `X`" (exact) | **raw tool** | deterministic + instant; wrapping loses determinism, often a token wash |
| "find where we handle auth and **compile curated findings**" | **agentic** ✓ | language task, big payoff, read-only (low risk) — the Explore pattern |
| "replace this exact line" | **raw `apply_patch`** | orchestrator already knows the exact change; instruction ≈ the edit |
| "update the changelog from these 5 commits" | **agentic** ✓ (gated) | needs drafting; verify with a deterministic check, not a re-read |
| "read these 50 files and summarize the data model" | **agentic** ✓ | huge size:instruction ratio |

**Anti-pattern:** turning *every* tool into a model call. Mechanical, exact, or
tiny ops stay deterministic. Over-delegation adds latency, nondeterminism, and
cost with no payoff.

## 4. Why it's gospel (and the gospel sharpens the engineering)

This is **Jethro at a smaller fractal** — Exodus 18:22, "every small matter they
shall judge," Moses keeps the hard ones. The substrate already embodies delegation
(gates, trust ladder, the Hinge); this pushes it down into the tool layer. The
convergence that makes it trustworthy: the gospel line **"delegate execution, not
discernment"** (`study/ai-stewardship-north-star.md`) is the *same boundary* as the
§3 discriminator — execution/mechanical → delegate; discernment/judgment → keep.
When the doctrine and the engineering draw the identical line, the line is right.
The failure mode is also the same: delegating *discernment* to a cheap model and
trusting it blindly = the "confident garbage" / abandoning failure. Guard it with
deterministic verification (§6), not faith.

## 5. The abstraction — an "agentic tool" registry row

Each agentic tool = a definition:

| Field | Meaning |
|---|---|
| `name` | tool name the orchestrator sees (e.g. `research_codebase`) |
| `description` | the use contract: what / when / returns — **plus a cost hint** ("expensive agentic search; for exact matches use grep") to steer when-to-delegate |
| `model` | cheap default (deepseek-v4-flash); per-call escalatable |
| `inner_tools` | the deterministic tools the sub-agent gets (research: grep/glob/read; edit: apply_patch/build/test) |
| `system_prompt` | sub-agent instructions + the return contract |
| `input_schema` | what the orchestrator passes (question/intent + scope, e.g. repo/dir) |
| `return_contract` | structured curated output (§6) |
| `budget` / `max_turns` / `timeout` | bounded — cost guard |
| `verification` | optional deterministic gate the sub-agent must pass before returning |

Realized as a purpose-built agent_family + a thin MCP wrapper that calls
`consult_subagent` against it and shapes the return. No new engine.

## 6. Return contract + verification (the trust layer)

The point is the orchestrator **trusts without re-reading**, so returns are
**curated + checkable**:

- **research_codebase** → `{ summary, findings[], citations:[file:line], confidence,
  caveats }`. Citations let the orchestrator spot-check cheaply (open one file)
  instead of re-doing the search.
- **edit_file** (later) → `{ diff, files_changed, verification:{build:pass,
  tests:pass}, notes }`. The orchestrator trusts the **deterministic gate**
  (build/tests pass) rather than reading the whole diff. If no gate exists, it must
  re-read — and then the tool wasn't worth it (so don't offer agentic edits without
  ground-truth).

## 7. Flagship: the ai-chattermax code / repo-reader persona

Ties two carry-over items together. The code persona = a tool-using persona (like
the librarian "Computer", but repo-scoped):

- **Cognition pipeline:** a `persona-turn-tools` variant with repo tools instead of
  gospel tools — specifically the **agentic** `research_codebase` (scoped to one
  repo, e.g. ai-chattermax or pg-ai-stewards), **read-only** (no edits from chat).
- **Flow:** human asks "how does the gateway auth a persona?" → persona orchestrator
  (kimi) calls `research_codebase(question, repo=ai-chattermax)` → deepseek-flash
  greps/reads/curates in a repo-mounted sandbox → returns findings+citations →
  orchestrator composes the chat answer.
- **Why it's the right design:** the orchestrator never reads the whole repo per
  turn; the cheap model does the grind; the persona stays affordable to run in a
  busy room. Citations make answers verifiable (links to file:line).
- **Safety:** read-only inner tools for the chat persona (no apply_patch). Repo
  access via the existing allow-listed worktree mount. An *engineering* persona that
  proposes changes is a later, gated step (drafts a PR, never merges — the Hinge).

This also generalizes "tool-using persona": the persona's tools can be cheap-model
agents, not just deterministic APIs. (librarian = deterministic gospel tools;
code persona = agentic repo tools.)

## 8. Cost model + measurement

- **Cost hints in tool descriptions** steer the orchestrator to delegate only when
  worth it (frontier models honor this well).
- **Measure, don't assume** (the council-vs-gift-matching lesson, [[project_council_review_beats_gift_matching]]):
  A/B `research_codebase` vs orchestrator-reads-directly on real questions — the
  saving is real only on large subtasks. Track tokens both sides + answer quality.
- The cheap model's tokens come from a substrate connector (opencode_go, or a
  second one per [[claude-worker-dispatch]] §8) — not Claude tokens at all.

## 9. Risks / anti-patterns

- **Over-delegation** (wrapping mechanical/tiny ops) — mitigate via §3 + cost hints.
- **Verification tax** — agentic edits without deterministic ground-truth force a
  re-read and erase the saving. Don't ship edit-tools without a check.
- **Nondeterminism / missed matches** — for completeness-critical exact search, keep
  raw grep available alongside; the agentic version is for curate-and-judge, not
  guaranteed-exhaustive match.
- **Repo-write safety** — chat personas get read-only; write-capable agentic tools
  only inside the coder pipeline behind the existing Hinge.
- **Latency** — each agentic call adds seconds; fine for chat/orchestration, wrong
  for tight inner loops.

## 10. Build phases (cheap-first)

- **P1** — `researcher-flash` agent_family (deepseek-v4-flash + read-only repo tools)
  + `research_codebase` MCP tool (consult_subagent preset) + return contract.
  Smoke: ask it a real ai-chattermax question, check curated+cited output.
- **P2** — wire `research_codebase` into a **code persona** pipeline in ai-chattermax
  (read-only, repo-scoped). Live test in a room.
- **P3** — A/B vs orchestrator-direct; tune cost hints; decide default.
- **P4** — `edit_file` agentic tool **gated on deterministic verification**
  (build/tests), inside the coder pipeline only (not chat).
- **P5** — generalize the registry so new agentic tools are declarative
  (family + wrapper + contract), and let the substrate orchestrator (not just
  Claude) use them.

## 11. Open questions for Michael

1. First flagship repo for the code persona — **ai-chattermax** or **pg-ai-stewards**?
2. Cheap model for the researcher family — deepseek-v4-flash (cheapest) vs a slightly
   stronger flash tier for better curation?
3. Should the code persona be read-only forever, or graduate to a PR-proposing
   engineering persona (gated)? (Lean: read-only first, prove value, then gate writes.)
4. Sync (`consult_subagent`, blocks the turn) vs async for `research_codebase` in a
   chat turn? (Lean: sync with a tight timeout — a chat turn needs the answer inline.)

## 12. Relation to other specs

- [[claude-worker-dispatch]] — the inverse direction (escalate up to Claude). Same tree.
- Persona pipelines (`r9-persona-tools.sql`, librarian) — the code persona is the next
  tool-using persona; this makes its tools agentic.
- Coder sandbox / CV2.1 worktrees — supplies the repo-mounted inner tools.
- CT2 self-context — orthogonal, but agentic tools *reduce* orchestrator context
  pressure (curated returns instead of raw file dumps), so they compose well.
