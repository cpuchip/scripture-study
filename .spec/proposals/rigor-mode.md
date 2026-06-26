# Rigor Mode — a study-disciplined response over a curated bucket

**Status:** proposed (2026-06-26). The capstone of the orientation arc (62/63/64): it
*uses* the autoload shelf, `orient_survey`, and the standing trajectory critique to turn a
fluent-but-untrustable research answer into one where every line traces to an observation.

## The problem (a real critique, verbatim in spirit)

A coworker ran their substrate instance (CKE) for a marketing report and a researcher read it
back. The output read like a competent marketer — and was *unfalsifiable*: "There is no way —
for you or a skeptic in the room — to tell which of these claims came from CKE observations
and which are the model's generic priors." Five gaps, each one a discipline our workspace
already hardened:

1. **Provenance.** No line ties to an observation; "camera-first," "all-in-one ecosystem" are
   indistinguishable from stock marketing priors until something cites the bucket.
2. **The believable claims are the dangerous ones.** "$50/$100/mo ceilings," "military-
   affiliated customers," "quote-form → SMS → callback" — specific enough to sound
   authoritative, which is exactly where a confident hallucination hides. Verify those first.
3. **Calibration is flat.** A thin finding and a robust one look identical on the page.
4. **Confirmation bias.** The prompt said "doorbell cameras"; the system confirmed a camera-
   first worldview. Did it find that, or mirror the premise?
5. **Observation vs recommendation are interwoven** — a reader can't separate "what customers
   said" from "what the AI suggests."

The bar the researcher set: *"once references land, the test isn't whether the output looks
good — it already does — it's whether every confident line actually traces to something real."*

## The output contract

Rigor mode changes what a response is *allowed to say*. Every factual claim must be one of:

- **`[grounded: <doc-slug>]`** — traces to a retrieved bucket document (with the quote/span).
- **`[inference]`** — a reasoning step built on grounded claims, labeled as such.
- **`[model-prior]`** — general knowledge the bucket does not support; allowed but flagged, so
  a skeptic can subtract it.

And the response is *structured* so the line stays visible:

- **What the data shows** — grounded claims only, each carrying its source and a **calibration
  tag**: `[multiple sources]` / `[single source]` / `[weak/anecdotal]`. (The vivint
  `sentiment-synthesis-taxonomy` is *source-weighted* already — rigor mode reads that weight
  through, it doesn't invent one.)
- **What we'd recommend** — clearly inference, building on the grounded section.
- **Premise check** — see below.
- **Provenance distinction:** a primary observation (sentiment data, a source-weighted finding)
  outranks the substrate's *own* prior synthesis. Citing a `vivint-reflect--*` doc is citing a
  prior opinion, not an observation — rigor mode says which it is.

## The mechanism (rides the orientation arc)

Rigor mode is not a new engine — it's the orientation loop, pointed at research rigor:

1. **Orient** — `orient_survey` (63) + `doc_search` over the bucket: what observations exist?
2. **Act under the rigor contract** — a loadable **`research-rigor` skill** (the autoload shelf
   from 62) carries the contract above; the response cites as it drafts (the world-graph already
   paints `source_refs`, so the rendering exists).
3. **Premise-neutrality reflex (the anti-confirmation-bias move).** When the prompt embeds a
   premise ("doorbell cameras"), rigor mode first runs the **neutral** version of the question
   ("what drives security purchases?") with the framing stripped, and reports whether the
   premise emerged from the data on its own or only when primed. This is the researcher's own
   test, automated — it's the critical-analysis posture check (discovering vs confirming).
4. **Verify** — the standing trajectory/grounding critic (64) re-reads the answer against the
   retrieved sources and **flags every claim that does not trace** — pressure-testing the
   believable-but-specific lines first (gap #2). Untraced claims are dropped or demoted to
   `[model-prior]` before delivery.

Delivery as a **mode toggle** in the Stewdio chat (like Fast/Smart): "Rigor" loads the skill
bundle + the grounded contract + the verify pass. Default is the normal fast response;
rigor mode is the deliberate, defensible one.

## Acceptance test (the coworker handed it to us)

Not "does it look good" — it already did. The test, on the **vivint bucket** (curated real
research; no CKE data here, same shape):

- **Neutral-premise test:** ask *"What drives home-security customers' satisfaction and
  frustration?"* — no product, no feature framing. PASS if the answer reaches its conclusions
  grounded in the vivint docs, **without** being primed toward them.
- **Trace test:** every line in "what the data shows" carries a `[grounded: slug]`; pull three
  at random and confirm the cited doc actually says it (the read-before-quoting check).
- **Calibration test:** a single-source claim and a multi-source one are visibly different.
- **Subtraction test:** a skeptic can delete every `[model-prior]` / `[inference]` line and be
  left with only what the bucket supports.

A before/after against this exact critique is the deliverable.

## Scope / privacy / OSS

- The **mechanism** (the `research-rigor` skill, the premise-neutrality reflex, the grounding-
  verify pass) is generic — OSS-core eligible, like the orientation baseline.
- The **vivint data** is file-private: all prototyping stays on the local rig + local
  embeddings; the bucket content is never pushed and never routed to a train-on-data provider.
- Pairs with: the orientation arc (62/63/64), the workspace `study` agent + `source-verification`
  + `critical-analysis` + `ben-test` disciplines (this is them, instantiated in the substrate),
  and `study/ai/harness/lending-the-substrate-our-orientation.md` (rigor mode is the next ring).

## Build order

1. This spec.
2. Prototype the grounding-rigor pass on vivint with the neutral question — proves the output
   contract is achievable on the local rig before any UI toggle.
3. If the prototype clears the bar: the `research-rigor` skill + the premise reflex + wiring the
   verify pass as a pre-delivery gate; then the Stewdio "Rigor" toggle.
