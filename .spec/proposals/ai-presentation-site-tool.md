---
title: AI-Generated Presentation Site Tool (for Teaching)
status: proposed
workstream: WS7 Teaching
created: 2026-04-22
source_brain_entry: c168be5c
binding_problem: For the planned teaching site (teaching AI / scripture study), we need a tool that generates beautiful presentation websites using Creative Commons art and scripture content with minimal manual effort. The tool itself isn't gospel content — it's infrastructure for creating it.
---

# AI-Generated Presentation Site Tool

## Binding Problem

Building a teaching site (WS7) means producing a lot of visually-rich pages — scripture quotes with art, concept walkthroughs, before/after comparisons. Doing this manually for every page is a bottleneck. An AI agent that takes a topic + scripture references and generates a Copilot-powered, CC-licensed presentation site would unblock the teaching project.

## Success Criteria

- A reusable tool (CLI or VS Code mode) that:
  1. Takes a study/topic input.
  2. Selects appropriate CC art (Met Open Access, Wikimedia Commons, Smithsonian, etc.).
  3. Generates a static presentation site using our existing publish pipeline aesthetic.
- First production use: generate a presentation page for one of the existing studies in study/.
- Output works on mobile and desktop.

## Constraints

- All art must be CC-licensed and properly attributed.
- Output must integrate with the existing publish pipeline (or replace it cleanly).
- Don't reinvent revealjs / similar — wrap an existing presentation framework.

## Related

- Pairs with `launch-youtube-channel` — these are the WS7 infrastructure pieces.
- Pairs with the teaching agent mode (already exists in `.github/agents/teaching.chatmode.md`).

## Phase 1

Survey existing presentation generators (revealjs, slidev, etc.) + CC art APIs. One-page evaluation. Decide buy-vs-build for each layer.
