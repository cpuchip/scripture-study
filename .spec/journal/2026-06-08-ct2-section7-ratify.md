---
date: 2026-06-08
title: §7 ratified — and the faceted-audience model that emerged from "it might get messy"
workstream: pg-ai-stewards
mode: plan/ratify
tags: [ct2, ratification, durable-memory, faceted-audience, collaboration]
---

# Ratifying §7 — a design that got better by being questioned

Michael: "lets ratify 7." I walked the open decisions rather than flip a flag. The
first pass settled the easy ones (build all of §7 incl. the gated §7.3; generous
notes cap ~40/~4k tokens; default audience = the authoring persona; §7.3 greenlit
as a gated, later-pass design). But the scope question opened something better.

## The turn

The spec's §7.1 scoped notes by `session | persona | global`. Michael pushed back
in two moves:
1. "We may also need a kind/type + purpose filter so roleplay personas don't get
   notes meant for code-style personas." → global is too blunt; reach ≠ audience.
2. "I want it to translate to all of pg-ai-stewards work — a webfetch agent noting
   a site needs a parser tool, for all webfetch… the tag/location makes sense but
   **it might get messy.**"

That "might get messy" was the real signal. The mess comes from *adding tiers*
(kind, then room, then work-type, …). The fix isn't a better tier list — it's to
stop having tiers.

## What we ratified: a faceted-audience model

A note carries **audience selectors** (`persona`/`kind`/`room`/`pipeline`/`global`/
`session`, AND-combinable). Every dispatch exposes its own **facets** (computed in
`compose_system_prompt`). A note renders iff its selectors match the dispatch's
facets — **one match rule, no tiers.** Every case Michael raised becomes a selector:
`{kind:code}` (his B / shared per-kind pool), `{room:10-forward}` (location),
`{pipeline:webfetch}` (work-type), `{persona:starlet}`, `{global:true}`. A new
dimension later = a new facet key, no schema change. Plus free-form `tags` for
search/curation (they don't gate delivery — a dispatch has no "purpose" facet, so
purpose is organization, not routing). Default audience `{persona:self}`.

This is the anti-mess move: the generalization is what *prevents* the sprawl, and
it makes the feature substrate-wide (not just chat personas) for free.

## A correction worth keeping

Michael's gut mapped "tell Starlet in 10-Forward, carries to Holodeck-3" to
*global*. It's actually *persona* — the substrate runs a session per (persona,
room), so cross-room-same-character is persona scope; global is cross-*persona*.
Naming that cleanly is what let the faceted model click.

## Reflections

- This is the council pattern working in miniature: the human surfaced the right
  worry ("messy"), the AI found the unifying abstraction. Neither alone gets the
  faceted model — his concern without my abstraction is a pile of tiers; my
  abstraction without his concern never gets built (the spec would've shipped
  session/persona/global).
- It also vindicates *walking* a ratification instead of flipping a flag. Three of
  four answers were quick; the fourth (scope) was where the design actually
  improved. The AskUserQuestion "clarify" path is where the value was.

## Carry-forward
- Build tasks #137 (§7 core — faceted self-notes + caps + kind + facets + prompt
  split + working tags) and #138 (§7.3 self-editable base prompt, gated, own pass).
- Mostly SQL like CT2.1/2.2, but the facet wiring touches `compose_system_prompt`
  again — use the same before/after backward-compat hash check (the k2/l13 lesson).
- Spec §7.6 is the as-ratified build doc. CT2 core (§§1-6) already shipped.
