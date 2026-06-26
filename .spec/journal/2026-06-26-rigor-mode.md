---
date: 2026-06-26
lane: pg-ai-stewards
topic: Rigor Mode — a traceable, source-cited response; the orientation arc, pointed at research
tags: [rigor-mode, provenance, harness, knowledge-anyone-can-use, orientation-arc, ben-test]
---

# Rigor Mode: the thesis, instantiated

## Where it came from

A coworker ran their substrate instance (CKE) for a marketing report and a researcher read
it back. The praise was real — it read like a competent marketer — and the critique was
sharper: the whole thing *asked for trust it couldn't earn.* "There is no way — for you or a
skeptic in the room — to tell which of these claims came from CKE observations and which are
the model's generic priors." Five gaps: provenance, verify-the-believable-claims-first, flat
calibration, confirmation bias (the prompt said "doorbell cameras"; the system confirmed
camera-first), and observation tangled with recommendation.

The striking thing: the coworker, with no knowledge of our covenant or skills, **wrote the
spec for our next ring without knowing it.** Every gap was a discipline our workspace had
already bled for — read-before-quoting, build-the-oracle-first, epistemic humility, the
critical-analysis posture check, the study structure. The substrate was missing exactly the
orientation the workspace carries. The same finding as the study, arriving from the outside.

## What it is

Rigor mode is the orientation arc (62/63/64) pointed at research. Not a new engine — the loop
we'd just shipped:

- **Orient** — `orient_survey` the bucket: what observations exist?
- **Act under a contract** — the `research-rigor` skill: GROUND OR FLAG every claim
  (`[grounded: slug]` / `[inference]` / `[model-prior]`), verify the specific claims first,
  calibrate by evidence strength, separate "what the data shows" from "what I'd recommend,"
  check the premise.
- **Verify** — the standing trajectory critic (64), tuned to grounding (v2).

Shipped as `65-rigor-mode.sql` (the skill → OSS baseline + a render refinement so a
*dispatcher-loaded* session skill reaches a skill-denied agent — the same lent-not-opted-into
principle as autoload), a `rigor` flag in `chat.go`, and a **🔬 toggle** in the composer.
v1: `c13c2ed..fd321f8`, virgin-smoke OK 54, fresh-image green, CI green.

## The test that mattered

Not "does it look good" — the critiqued report already did. The coworker handed us the
acceptance test: ask a **neutral** question, no premise, and see whether the conclusions emerge
from the data on their own. On the vivint bucket ("what drives home-security satisfaction and
frustration?"), the rigor toggle returned a grounded, severity-calibrated, observation-first
answer citing the corpus — night-and-day from the report it was modeled on.

## The honest v1→v2 line (Ben Test)

v1 clears the critique's *core* — provenance, calibration, separation — but not yet its
*deepest* test. The local model grounded **narratively** ("the doc notes…") rather than with
the literal `[grounded: slug]` tags, so a skeptic can't yet *mechanically subtract* the priors.
And it cited the *synthesis* doc, not the *primary* Trustpilot/BBB observations. Both are v2,
both specced: the premise-neutrality reflex (run the question stripped of its framing) and the
verify-pass as a pre-delivery gate that enforces the tags and flags the untraced. v1 makes a
research answer *defensible*; v2 makes it *line-by-line auditable*.

## Why it matters

This is the clearest instance yet of the thesis the study named: **the workspace's hard-won
orientation, lent to the substrate, becomes a feature anyone can use.** A marketer with no
covenant gets a researcher's rigor by pressing a button — and the rigor is *real*, because
every line is being made to trace to something the bucket actually holds. "Knowledge anyone
can use" stops being a slogan when the output can survive a skeptic in the room.

## Carry-forward

- **Rigor v2:** the premise-neutrality reflex + the verify-pass-as-gate enforcing the tags
  (`.spec/proposals/rigor-mode.md`). The build that makes priors subtractable.
- The orientation carry-forwards still stand: more disciplines onto the shelf
  (read-before-quoting → digesters, inverse-hypothesis → coder), and the post-demo
  `force-final-at-cap` floor.
- vivint stays file-private — rigor prototyping is local-rig / local-embeddings only.
- The lineage to keep visible: study (`lending-the-substrate-our-orientation.md`) → orientation
  arc (62/63/64) → rigor mode (65). One thread, orient → act → verify, three times over.
