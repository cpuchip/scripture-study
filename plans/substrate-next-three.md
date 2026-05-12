# Shipping Order: Batch I Write-Back, yaml.rs Refactor, and Phase A Reaper

**Binding question:** What is the right order to ship the next three pg-ai-stewards substrate items: (a) Batch I — agent write-back rung on the trust ladder (agents propose studies/notes/lessons through gate machinery before persisting); (b) the yaml.rs Rust parser refactor (rule of three triggered by scripture-study + general-research + planning-partner intents all needing parsing); (c) the still-deferred Phase A pgrx BGW SPI longjmp catch + 60s periodic reaper tick?

**Project:** pg-ai-stewards

**Date:** 2026-05-12

---

## The plan

Batch I ships first. It is the explicit compounding step the last three sessions built toward: H.1.7 gave the substrate memory, H.2 gave it context-gather, and H.3 designed planning. Batch I is the substrate noticing it has hands — the first rung where agents propose studies, notes, or lessons that route through existing gate machinery before persisting.

We do not wait for the full H.3 build to unblock Batch I. Instead we pull the studies generalization schema migration forward as its opening work_item. The D-H3-B ratification in H.1.7 already approved adding `tags`, `source_type`, and `project_association` to the studies table as nullable columns. That migration runs first so agent-proposed studies have a clean place to land.

Once the table accepts non-scripture rows, the second work_item wires agent proposals through the same `evaluate_gate` and `apply_gate_decision` patterns that Phase B already uses. No new gate logic is invented; the existing machinery simply receives a new source of proposals.

yaml.rs ships second. By the time Batch I is closing, H.3 ratification will have answered its five open questions, making the third intent — `planning-partner` or `professional-awareness` — concrete and triggering the rule-of-three. The refactor replaces the SQL workaround with native Rust `serde_yaml` parsing in one focused session.

Phase A ships last. The NOTICE+NULL sidestep in `mcp_proxy_enqueue` has held stable since H.1.5a. Neither Batch I gate wiring nor yaml.rs build-time parsing introduces new bgworker SPI exception paths. The longjmp catch and 60s periodic reaper tick remain necessary polish but do not block the compounding work ahead.

## Assumptions

- Batch I "agent write-back" means proposals route through the existing `evaluate_gate` → `apply_gate_decision` flow before persisting.
- The studies generalization (`tags`, `source_type`, `project_association`) ratified as D-H3-B in H.1.7 is a prerequisite for agent-proposed studies. We pull this schema migration forward rather than waiting for full H.3 build.
- yaml.rs is rule-of-three gated until H.3 ratifies the `planning-partner` / `professional-awareness` intent. H.3 ratification happens during or immediately after Batch I, making the refactor timely but not a hard prerequisite for Batch I itself.
- Phase A's NOTICE+NULL sidestep is stable. Batch I and yaml.rs do not add new bgworker SPI exception surface, so Phase A urgency does not increase.
- The H.2 bridge deadline-handler carry-forward bug is small but may need a patch before Batch I end-to-end testing if that testing exercises fs-read context gathering.

## Risks

- The studies generalization migration may need revision if Q-H3.1–Q-H3.5 constrain `source_type` to an enum or `project_association` to a foreign key before the schema can be considered final. Keep the first cut permissive — text and jsonb, no strict constraints — and tighten later.
- Batch I scope could bloat if treated as a single deliverable spanning schema, SQL, backend, and UI. Split it into discrete work_items, each bounded to roughly two hours.
- The H.2 bridge stale-session bug could leave queue rows stuck during Batch I validation. Patch `bridge_run.go` as Batch I prep if end-to-end testing is required immediately.
- The yaml.rs refactor might reveal `serde_yaml` shape mismatches deeper than `src/yaml.rs`, cascading into intent or covenant parsing. Scope the refactor to `parse_yaml_intent` only, leaving covenant parsing untouched unless it is also rule-of-three triggered.

## Next steps

Run the studies generalization schema migration first. Then wire agent-proposed rows through the existing gate SQL functions. The first deliverable therefore comprises four work_items: (1) schema migration D-H3-B, (2) gate SQL for agent proposals, (3) backend endpoint to accept agent proposal payloads, and (4) Vue gate-review queue surface for agent-authored items. yaml.rs follows once H.3 ratifies its third intent. Phase A closes the sequence.