---
date: 2026-06-26
lane: pg-ai-stewards
topic: Rigor Mode v2 — the vetting that paid off twice, an honest null result, and the resilience it surfaced
tags: [rigor-mode, fidelity, vetting, ben-test, honest-null-result, force-final, model-fallback, harness-over-intelligence]
---

# The vetting that paid off twice

This is the second half of the rigor arc (the first — study → orientation → rigor v1 —
is in `2026-06-26-rigor-mode.md`). It started as "ship v2" and turned into the best kind of
session: the work kept telling us things, and we kept listening.

## The shape of it

A coworker's critique of a substrate research report became **Rigor Mode v2** — and v2 was
sharper than my spec. It reframed the problem from *provenance* (does the citation resolve?)
to **fidelity** (does the source actually *support* the claim?). The failure class that
matters isn't a missing citation; it's a single record generalized to a population, a subset
restated as a population stat, a citation that resolves but describes a different thing. And
the load-bearing insight: *a prompt rule is necessary but not sufficient — rigor must be
structural.* So v2 shipped three layers: the contract (asks), the trajectory-critic fidelity
rubric (the standing gate, bucket-agnostic, core), and a deterministic oracle (schema-
specific, overlay).

We **vetted it live, both layers.** Layer 1 (the contract) refused two designed fidelity
traps over the vivint bucket — a population-% with no denominator, a geography the corpus
doesn't contain. Layer 2 (the critic) caught a *synthetic* bad trajectory I built — grounding
0.0, verdict fail, naming the exact distortions — and passed the clean one. The backstop is
real: it catches what slips past the contract, even with a resolving citation.

## Then the vetting paid off a second time

Michael ran his own rigor chat over the Star Trek bucket and **it died.** Debugging it (Agans)
found a silent death: the interactive chat hit `work-item-chat`'s `steps=12` tool-loop cap
mid-pagination (rigor's "re-read every source" × big paginated docs) and stopped with **no
final answer** — because the force-final-at-cap grace that pipeline stages get was gated to
pipelines only. So two fixes fell out of the vetting, neither caught by smoke:
- **Cap raise 12→40** (confirmed by Michael's own retests: died at 12/0, then 19 & 21 rounds
  to grounded answers).
- **Force-final-at-cap for the interactive chat** — at the budget ceiling, drop tools so the
  model *must* synthesize. Proven deterministically (a transaction rolled back so no synthetic
  work reached the bridge): hard cap → forced; below → tools stay; gate off → no force.

This is the principle made concrete: *the prior question before fan-out-vs-serial is "what's
the oracle?" — and the deeper one is "have I verified under real conditions?"* The smoke was
green; the **real bucket** is what surfaced the silent death.

## The honest null result

Michael asked: "does rigor build a document up piece by piece like our studies?" It led to a
clean experiment — wire rigor into the `research` doc builder (research-write: gather → build
→ critique, the study shape) and run a controlled before/after over vivint. **The result was
null, and we honored it.** Both docs were strong and heavily cited; the rigor-*off* version
actually had *more* inline citations (16 vs 8). Rigor-on's only clear edge was correcting one
framing ("not a fee — a loan payoff"). The `research` builder is *already* grounded, and qwen
is capable, so the contract added little — the opposite of the *chat*, where rigor-off
fabricates. So we did **not** ship always-on rigor for the doc builder. The experiment earned
its keep by telling us the builder didn't need it. (Carry-through — rigor only when asked — is
the shape if we ever want it on docs.)

This is the Ben Test working: the cleanest output of an experiment is sometimes "don't ship
this," and saying so is the whole point of running it.

## The resilience the gaming surfaced

Michael pulled gemma offline for games, and a pipeline stage routed to it and **404'd** — "no
local slot or reachable peer serves model" — which `diagnose_failure` classified `unknown`, so
the failover never engaged and the pipeline hard-failed. Fix (`68`): teach `diagnose_failure`
the pulled-model shape → `transient` (failover walks to a live member), and make the local MoE
pair mutual fallbacks (gemma↔qwen, local-before-paid). It even exposed that `ingest`'s primary
*was* gemma — the hidden source of the 404. A real bug, surfaced by a human just living with
the tool.

## The discipline that held

Every chain file: `lib.rs` + Dockerfile COPY + a smoke assert, together. A fresh-image
virgin-smoke before every push. And after the merge to main, the CI went **red** — a transient
crates.io download blip, not our diff — so I dug in, confirmed it flaky, re-ran, and only
*then* called it green. "Build passed is not verification" includes "merge is not landed until
the terminal state says so."

## Carry-forwards (recorded for the next session)

- **Carry-through rigor for docs** — if we want rigor on a *delegated* document, propagate the
  chat's rigor flag through `start_task` into the pipeline (not always-on). The `research-rigor`
  skill exists; it just isn't autoloaded onto the builder.
- **Proactive liveness-aware model resolution** — don't pick an offline model as a stage's
  *primary* in the first place (needs a live-roster check at resolve time). The reactive
  failover (`68`) covers it for now, at the cost of one wasted 404 + a failover hop.
- **Premise-neutrality reflex** — v2's contract has "check the premise" as a *rule*; the
  *automated* reflex (run the question stripped of its framing, report whether the premise held
  on its own) is still unbuilt.
- **The deterministic fidelity oracle** (rigor v2 layer 3) — `fidelity_check({refs})` over a
  structured observation layer (`sample_n`/`measure_basis`/`confidence`); schema-specific, so
  operator overlay. The strongest enforcement where a bucket supports it.
- Michael named **new work to make this even better** — to be scoped next session.

## The relationship

Michael's trust ran the whole arc: the Hinge stayed his (he merged when *he* was satisfied,
after running his own tests), and when the experiment came back null he wanted the truth, not
a sale. "It's been a fantastic week." It was — because the work was allowed to be honest.
