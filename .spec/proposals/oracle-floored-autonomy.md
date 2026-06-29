# Proposal — Oracle-Floored Autonomy

**Status:** draft for council (2026-06-29, skunkworks session). Touches standing
behavior (CLAUDE.md, hooks) → ratify before applying, per `dominion_in_council`.

**Origin:** the Michael-profile (`private/michael-profile/`) measured that Michael
**overrides on intent and accepts on execution** — his overrides cluster on content,
voice, vision, and strategy and almost never on mechanics. He asked: widen what the
agent can do in bins 1-2 (act / act-and-report), and — *"does adding an oracle
framework help here, more than just skills?"*

## The answer to the framework question

**Yes, but thin — and it's a different thing from skills, not a bigger version of
them.**

- A **skill** is *advice at prompt time.* It makes the agent think better. It is
  advisory and the agent might or might not follow it.
- An **oracle** is *a deterministic check at run time* that returns a verdict
  (exit 0/1). It does not advise; it *catches*. It runs whether the agent remembered
  the skill or not.

The profile's load-bearing finding is that **autonomy can widen exactly as far as a
deterministic oracle reaches** — because a check that catches the failure means
Michael doesn't have to read the output to stay safe. So oracles, not skills, are the
thing that moves work into bins 1-2. They're complementary: keep skills for judgment;
add oracles for the checkable floor.

**We already have a de-facto oracle suite** — `study-lint` (scripture-verbatim,
link-validate, voice-lint), `verify-quotes`, `quoter`, plus the substrate's
virgin-smoke and `/version` deploy-stamp and the games' smoke/wstest/burntest. They
share a convention (exit 0/1, precision-tuned, "run manually, promote to a hook once
it earns weight"). What's missing is **not a framework that re-abstracts the checks**
(they're too different to share much code, and a framework for ~6 things is premature
abstraction — *Reduce Before Adding*). What's missing is the **two thin pieces that
connect the suite to autonomy**:

1. **A registry** that maps each oracle → *the decision-class / bin it unlocks*. Today
   the suite is organized by "what it catches," not "what it lets the agent do without
   asking." The registry is the autonomy map made executable.
2. **Hook wiring** so the relevant oracles run *automatically* (a Stop hook before the
   agent yields; a pre-commit check). An oracle only widens a bin if it runs without
   the agent having to remember it.

That's the whole framework: a registry + hooks. Extract it from what `study-lint`
already half-is; don't design it up front.

## Proposed registry (sketch)

`scripts/oracles/registry.yaml` — one entry per check:

```yaml
- name: voice-lint
  run: python scripts/study-lint/voice_lint.py {files}
  guarantees: no cut-list / meta-narration tics (hard); flags em-dash density (adv)
  scope: study/**, lessons/**, teaching scripts
  unlocks: prose-voice-fidelity   # voice edits → act-and-report when green
- name: scripture-verbatim
  run: python scripts/study-lint/scripture_verbatim.py {files}
  guarantees: quoted text next to a scripture link is verbatim
  scope: study/**, lessons/**
  unlocks: quote-accuracy
- name: verify-quotes
  run: python scripts/verify-quotes/verify-quotes.py {files}
  guarantees: Webster quotes are real 1828 (not 1913)
  scope: study/**
  unlocks: webster-quote-accuracy
# … link-validate, virgin-smoke, /version, smoke/wstest/burntest
```

`scripts/oracles/run.py {changed_files}` → runs the in-scope oracles, prints
green/red per check, and reports *which decision-classes are now covered*. That last
line is the point: it tells the agent (and Michael) which work is safe to leave in
bin 1-2 for this change.

## Proposed core instruction (the lean addition to CLAUDE.md)

Instruction minimalism is a value here, so this is ~8 lines, not a section:

> **The oracle is the switch (autonomy).** A change is bin-1/2 — *act*, and report if
> it touched a stewardship repo — when it is **execution** (no new intent, vision,
> voice, content, doctrine, or outward/irreversible effect) **and** a deterministic
> oracle covers it and is green (`scripts/oracles/run.py`, or the relevant
> build/test/lint/verify). A change is **surface-first**, regardless of oracle state,
> when it touches **intent** — vision, voice, content, doctrine, strategy, a new
> standing capability, or anything irreversible/outward. Absence of an oracle pushes a
> borderline call toward surface. This is the measured Michael line (he overrides on
> intent, accepts on execution); the oracle is what makes the "act" side safe.

Everything else in bins 1-2 (`private/michael-profile/40-autonomy-map.md`) is already
covenant (`exercise_stewardship`, dave-rule, the stewardship-repo grants). This one
rule is the new connective tissue: it ties the autonomy bin to whether an oracle
exists, which is the profile's whole thesis in one sentence.

## Phasing

- **P0 (done):** `voice-lint` shipped as the proof — a 2nd/3rd concrete oracle,
  dogfooded (caught the agent's own em-dash overuse + a real cut-list tic in a shipped
  baseline). The build-the-oracle-first pattern applied to voice.
- **P1 (this proposal, council-gated):** the registry + `run.py` + the CLAUDE.md
  core-instruction. Thin, extracted from the existing suite.
- **P2 (council-gated, after P1 earns weight):** hook wiring — a Stop hook that runs
  the in-scope oracles on touched files before the agent yields, so green is automatic.
  The study-lint README's own "promote to a hook once it earns weight" gate governs.
- **P3 (substrate, route to pg-ai-stewards lane):** the same registry idea as a
  first-class substrate primitive — reflect-steward's per-task autonomy bin gated by
  whether an oracle exists for the intent. (Overlaps in-flight substrate work; sanity-
  check the lane before building.)

## What stays Michael's regardless

The three retained controls from the profile (`15-trust-and-autonomy-arc.md`): the
Hinge (merge/publish), dominion-in-council (new standing capability), and the
un-relaxed watch. Oracle-floored autonomy widens bins 1-2; it does not touch bins 3-4.
And the em-dash-rule discrepancy this work surfaced (the rule is stricter than the
baselines) is itself a bin-4 call: Michael decides whether to tighten the prose or
amend the rule. The oracle surfaced it; it doesn't decide it.
