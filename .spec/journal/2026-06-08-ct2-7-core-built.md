---
date: 2026-06-08
title: §7 core built — faceted durable self-notes, the Hermes loop on the substrate
workstream: pg-ai-stewards
mode: dev
tags: [ct2, section7, durable-memory, faceted-audience, hermes, c-f-cadence]
---

# §7 core: the durable self-notes loop

Michael ratified §7 then said "lets build!" Built the faceted-audience durable
self-notes — the heart of §7.1/§7.2 — SQL-only (no rebuild), in C-F gates.

## What shipped

- **7a** (`ct2-7a-self-notes-store.sql`, cde7a73): `agent_self_notes` (note,
  audience jsonb, tags[], created_by/session) + `agents.kind` (roleplay/code/
  librarian) + `dispatch_facets()` (global/session/agent_family/kind/pipeline) +
  `render_self_notes()` — the ONE match rule `facets @> audience`, capped
  ~40 notes / ~4k tokens. Verified in isolation: a `{kind:code}` note routes to
  dev + the research family but NOT a roleplay persona (Michael's exact ask);
  `{agent_family}`/`{pipeline}`/`{session}`/`{global}` all route precisely.
- **7a2** (`ct2-7a2-wire-self-notes.sql`, 29cf9f2): wired the notes block into
  `compose_messages`. Based on the LIVE def (the k2/l13 lesson), one inserted
  line. **Backward-compat hash-verified** — empty notes → c21b449e byte-identical;
  a matching global note → the YOUR DURABLE NOTES block renders; removed → baseline.
- **7b** (`ct2-7b-self-notes-tools.sql`, 5005961): `remember`/`forget` agent tools
  (sql_fn, using the `_session_id` CT2.3 already injects — no extra Rust).
  `session_agent_family()` derives the author from the pipeline stage; default
  audience `{agent_family:self}`; write cap 40/author (forces the prune loop);
  `compose_tools` gate extended so remember/forget show only for
  context_tools_enabled families. Smoke: derived `research` from the gather
  session, defaulted audience correctly, explicit `{kind:code}`+tags honored,
  forget-by-handle worked, gate held (enabled sees both, dev sees 0).

The whole loop — store → render → remember/forget → gate — works for the
substrate facets. Any substrate agent (dev, research, code-pr, webfetch…) with
`context_tools_enabled` can now keep durable, audience-routed working knowledge
that survives compaction and sessions. The Hermes self-curation loop, on pg.

## What the build taught / confirmed

- The faceted model paid off immediately: 7a's renderer needed *no* special-casing
  per scope — `dispatch_facets @> note.audience` is the entire match. The
  generalization Michael's "it might get messy" forced is exactly what kept 7a
  small.
- The k2/l13 hash discipline is now reflexive: 7a2 touched the big composer again
  and the before/after md5 was the first thing I checked. c21b449e both sides.
- CT2.3's `_session_id` injection already carried 7b — no new Rust. The earlier
  investment compounded.

## Stopping here (Mosiah 4:27)

Clean, committed, verified gate; ledger clean (220 files); soak resumed. This has
been a long arc (overnight 3 + CT2.1/2.2/2.3 + §7 ratify + §7 core). The remaining
§7 pieces are each substantial and distinct surfaces, so they're the right place
to pause for a fresh push:

- **7c — persona + room facets:** needs persona-host (Go) to write a
  session→(persona, room, kind) mapping that `dispatch_facets` reads; then the
  chat-persona default audience upgrades from `{agent_family}` to `{persona:self}`.
  This is the headline persona/room use case (Starlet, 10-Forward).
- **7d — §7.4 working tags:** `context_set_tag` auto-stamping (a trigger + a
  per-session current-tag) + batch `context_fold/mute/expand/pin_tag` (one
  circuit-breaker event). A distinct in-window-batching feature.
- **§7.3** (task #138): the gated self-editable base prompt, its own pass.

## Carry-forward
- Root unpushed (Michael pushes): cde7a73, 29cf9f2, 5005961 + memory.
- 7c will need the same compose_messages/dispatch_facets care + the hash check;
  it touches persona-host (separate Go binary, its own rebuild — not a substrate
  rebuild).
