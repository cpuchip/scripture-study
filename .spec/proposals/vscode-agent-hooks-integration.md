---
title: VS Code Agent Hooks — Workflow Integration
status: proposed
workstream: WS5 Memory & Process
created: 2026-04-22
source_brain_entry: 17735267
binding_problem: VS Code 1.111+ exposes agent hooks. We have a chained agent workflow that today is implicit (run study → run review → run publish manually). Agent hooks could let one mode hand off to the next automatically without losing context.
---

# VS Code Agent Hooks — Workflow Integration

## Binding Problem

Right now our multi-mode workflow (study → review → podcast → publish, or plan → dev → publish) requires Michael to manually invoke each mode. Context is lost between hand-offs unless externalized to scratch files. VS Code added agent hooks in 1.111 that could automate this transition while preserving context.

## Success Criteria

- A study agent finishing a session can trigger a review-pass agent automatically.
- Hand-off carries enough context that the next agent doesn't have to re-read 200 files.
- Hooks are opt-in per-mode (not every study should auto-publish).

## Open Questions

- What's the actual hooks API surface? Read https://code.visualstudio.com/updates/v1_111 first.
- Does it work with our custom `.github/agents/` mode definitions?
- Does it integrate with our session-journal system?

## Costs / Risks

- Could create runaway chains if not gated.
- New VS Code feature → API may shift.

## Phase 1

Read the API docs. One-page evaluation: can our existing modes use this, or does it need a different harness shape?
