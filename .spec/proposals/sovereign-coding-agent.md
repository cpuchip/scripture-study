# Garrison — A Lean, Sovereign, Local-First Coding Agent

**Date:** 2026-06-13
**Status:** Proposed — awaiting council (`dominion_in_council`; nothing built until ratified)
**Origin:** general-workspace session, out of the Euclid digestion and the "can pg-ai-stewards become our own opinionated CLI coding tool?" question.
**Working name:** *Garrison* — from the preside study (Webster 1828 *praesidium*, the fortified position held when the field is threatened). Naming is an open question; the name is doing real work here, so it leads.

---

## Binding question

What is the **leanest stack that lets Michael keep coding productively on his own hardware, with a weak local model (qwen3.6-27B class), owning the whole thing — if frontier-model and Claude Code access were gone tomorrow?** And the prior question underneath it: how do pg-ai-stewards' principles make a model that weak *trustworthy enough to ship code*?

## The heart (why this exists)

In Michael's words: *"if I lose access to claude code and frontier models, and all I am left with is something like qwen3.6-27B then I'd want a lean stack that enables me to code with my local hardware without the fuss of dealing with something I don't have full control over."*

This is not a market play and not a feature race. It is a **go-bag** — a fortified fallback position. The values are resilience, sovereignty, and control: a coding agent Michael fully owns, runs on hardware he owns, against models he runs, that survives the loss of everything rented. He doesn't love how opencode or "pi agents" are put together and hasn't tried hermes, so there is room for an opinionated alternative built on principles already proven here.

## What already exists (we are not starting from zero)

- **`stewards-cli`** — ~1,160 LOC, twelve subcommands including **`materialize-writes`** (DB → working tree). The separation Garrison needs — *think somewhere, write to local files* — already exists in the substrate.
- **`coder-mcp` + the `code-pr` pipeline** — plan → implement → verify (`go test`) → commit → push → PR, proven end-to-end on OSS (M2). Today it runs in a sandbox clone and opens a PR. Garrison brings that capability **home**: the working tree, interactive, no clone.
- **This workspace** — `covenant.yaml`, session lanes, grounding hooks, skills, `verify-quotes`, the study-linter, the reground counter. That *is* an opinionated, principled coding harness; it just happens to be layered on Claude Code rather than shipped as a binary. It is the client-side prototype of Garrison.
- **Spin** — Michael's local-model voice front (qwen on LM Studio). The local-model gotchas are already mapped (thinking-budget behavior, non-thinking instruct models for tool loops).
- **`principles.md` → "Harness > Intelligence."** The whole bet, already written.

## The thesis: Harness > Intelligence is the enabling bet

The opinionated harness is what makes a weak local model produce code worth shipping. Source-verification cut confabulation more than any model upgrade did; phased workflows beat single-pass prompts regardless of model. On a 27B local model the governance is not decoration — it is **load-bearing**. That is simultaneously why Garrison can work on local hardware *and* why it is differentiated from every other CLI agent. State it plainly: Garrison's value is **highest exactly in the survival scenario it is built for**, because the weaker the model, the more the harness matters.

## Architecture: B + C, corrected for sovereignty

The earlier sketch of option (b) put "Claude Code / the Agent SDK as the hands." The sovereignty scenario forbids that — Garrison must survive their loss, so the executor cannot depend on them. The corrected synthesis:

- **The executor is a lean loop Michael owns** (Go), driving a **local model** (LM Studio / Ollama / llama.cpp). This is the (c) spine: a thin local loop.
- **The substrate's governance is the engine** — council, verify, compact, work-item, watch — available via MCP and/or vendored in-process. This is the (b) relationship: the substrate *presides*, the lean loop *labors*, mapped onto the presiding covenant (Garrison presides over its own sub-steps under D&C 121; force only at the walls — cost caps, sandboxes, deny-lists).
- **DB-optional, graceful degradation.** Floor: principles plus local memory files, exactly how this workspace runs with no database. Ceiling: substrate-backed engrams, work-item ledger, cost caps, council records. Better-with, never requires. In pure survival mode Garrison runs with **no Postgres at all**.
- **Frontier-as-luxury, never as dependency.** When Claude Code or a frontier API *is* available, Garrison may dispatch heavy steps to it as an optional stronger pair of hands. It must never need it.

The single most important commitment: **floor mode needs nothing but the Go binary and a local model.** Everything else is enrichment. That is what makes it a real go-bag rather than a thin client to a server that might also be gone.

## What Garrison is deliberately NOT

- **Not a frontier-feature competitor.** It will not out-edit Claude Code or aider, and trying would be the losing game. The niche is governance, not edit quality.
- **Not opencode-complexity.** Lean is a hard requirement, not a preference — Michael named the dislike directly.
- **Not a standalone-agent maximalist rewrite.** Reinventing a full frontier-grade agent loop betrays the substrate's identity (presider, not executor) and drowns in tool-protocol churn. Rejected.
- **Not DB-required**, and **not a replacement for `stewards-cli`** (a separate thing that may share libraries).

## The lean core loop

`read working tree → plan (council-lite) → edit (local model) → verify (the oracle) → watch / repeat`, in small steps, with the oracle as the safety net under a weak model. Each substrate principle has a concrete job in that loop:

- **Build-the-oracle-first** → the verify gate: build + tests must pass, plus code detectors in the study-linter spirit ("cite the warrant" for code = every change carries a passing test or a named reason). The deterministic floor is what lets a 27B model be *trusted* rather than *believed*.
- **Judges, not executors** → surface decisions to Michael instead of burying them in an opaque path; the weaker the model, the more it should ask.
- **Council / D&C 88:122** → one local doer plus a critic pass (even the same model, a second adversarial look) catches what the tired doer missed. The workspace already learned that the critic loop beats per-stage gift-matching.
- **Inverse hypothesis** → after a fix: reproduce the failure, apply, confirm gone, remove, confirm it returns. "Tests pass" is not verification.
- **Gated autonomy** → human-in-the-loop by default; tighten the gate as the model weakens.
- **Presiding / watch** → Garrison watches its own sub-steps to *intent*, not just to completion.

## Why governance is load-bearing here, not luxury

A 27B model hallucinates more, plans worse over long horizons, and drifts faster. The harness compensates, mechanically: decompose into steps small enough that a weak model rarely goes wrong; gate every step behind a hard oracle (build/test/lint) it cannot talk its way past; add a critic pass to catch the doer's misses; keep the autonomy gate tight so a human confirms the consequential moves. Strip those four away and a local-model coding agent is a liability. Keep them and it becomes usable. They are the product.

## Local-model design constraints (from Spin + memory)

- qwen3.6-27B on LM Studio always reasons; a small `max_tokens` yields empty `content` with `finish_reason=length`. Give it ≥2000 tokens; the answer is in `content`, the reasoning in `reasoning_content`. Tool-calling on local models is weaker and inconsistent.
- Design for that reality: structured output with a forgiving parser, retries, and possibly a split — non-thinking instruct models for the tool loop, reasoners reserved for planning. Distrust a negative result from a parser written in haste (the verify-via-real-path lesson applies double here).

## Tensions and risks (honest)

- **Capability floor.** It will not match Claude Code, full stop. The bet is "good enough, fully owned, always available," not "best." Name it so no one is surprised.
- **The yet-another-agent trap.** Mitigated by staying lean, owning the governance niche, and the dogfood test: in survival mode Michael uses it by necessity; in luxury mode he would only reach for it *for the governance*. If the honest answer in luxury mode is "I wouldn't," that argues for keeping Garrison small and the substrate-as-MCP path primary.
- **Maintenance.** Owning the whole stack is a real ongoing cost; lean and library-reuse are the only defenses.
- **Effort vs. the parity roadmap.** Garrison is post-cut. Spec now, build later. Do not fork the parity push.

## Phasing (post-parity / post-cut)

- **P0** — this spec + council ratification (`dominion_in_council`).
- **P1** — the lean local loop MVP: read / plan / edit / verify on the working tree, one local model, no DB. The pure go-bag floor. Dogfood on a small real task.
- **P2** — the code oracle suite: a build/test wrapper plus code detectors reusing the `verify-quotes` / study-linter patterns (precision-tuned, oracle-first).
- **P3** — the council/critic pass (the D&C 88:122 lever).
- **P4** — substrate-backed enrichment over MCP (engrams, work-item ledger, cost caps); the DB-optional ceiling.
- **P5** — package and share; ties to `plugin-someday` and ai-jumpstart / *Beyond the Prompt* — Garrison is the tool that practices what the book preaches.

## Relationship to existing assets

`stewards-cli` (sibling; shares libraries; `materialize-writes` is the seed of the local-write path) · `coder-mcp` (the capability brought home from the sandbox) · the Claude Code workspace layer (the client-side prototype) · `plugin-someday` (the P5 packaging) · ai-jumpstart / *Beyond the Prompt* (Garrison as the embodied companion) · Spin (the local-model sibling that already pathfound the runtime gotchas).

## Open questions for council

1. **Name** — Garrison, or something else?
2. **How lean is lean** — single Go binary? What is the irreducible core?
3. **Primary local runtime** — LM Studio vs. Ollama vs. llama.cpp.
4. **Tool-calling strategy for weak models** — structured output, constrained decoding, ReAct-text, or non-thinking-only for the loop.
5. **Substrate coupling (the crux of sovereignty)** — does floor mode MCP-call the substrate, or vendor a minimal subset of its logic so Garrison needs *nothing but itself and a local model*? True sovereignty likely means floor mode cannot depend on Postgres at all.
6. **Plugin relationship** — is the `plugin-someday` Claude Code plugin simply Garrison's luxury-mode client?
7. **P1 dogfood target.**

## Recommendation

Pursue it, as specced: B + C, executor owned-and-local, DB-optional, governance as the safety net, built after the cut. Hold the one commitment above all others — **floor mode runs on nothing but the binary and a local model** — because that is the difference between a sovereignty tool and a thin client to a server that might also be gone.
