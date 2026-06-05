---
date: 2026-06-05
title: R.9 — tool-using chat persona (the Library "Computer", AXR5)
tags: [persona, tools, ai-chattermax, librarian, gospel-search]
---

## What happened

Built the tool-USING chat persona — Michael's Library "Computer" (AXR5). The
character personas (R.7/R.8) send zero tools (`persona` agent, deny *). R.9 is the
deliberate inverse: a persona that can SEARCH real sources and answer in chat with
citations.

**r9-persona-tools.sql** (applied live; pipelines/agents are DB data, not in the
lib.rs build chain — same as r7/r8):
- `librarian` agent: tool-using reference posture, read-before-quoting discipline
  baked into the prompt ("search first, never invent a reference"), SILENCE escape
  hatch like `persona`.
- Curated allow-list via the **longest-pattern-wins** perm model (confirmed in
  schema.rs `effective_tool_action`: ORDER BY length(tool_pattern) DESC, default
  allow): `deny *` + allow `gospel_*`, `study_*`, `strongs_*`, `define`,
  `modern_define`, `byu_citations*`, `read_corpus_parents`. The built-in
  `stewards-explore` agent was the precedent (deny *, allow brain_*/skill).
  `compose_tools('librarian')` verified = exactly 18 reference tools, no
  fs/git/coder/spawn/brain.
- `persona-turn-tools` pipeline: single stage, **tools ENABLED**
  (tools_disabled:false), kimi-k2.6/opencode_go, max_tokens 3000 (tool loop needs
  headroom). Name starts `persona-` so the R.8 one-shot auto-verify trigger fires.

**seed.go**: added the `chip-assistant` Computer persona (display "Computer",
pipeline persona-turn-tools) so it's durable across a fresh host rebuild.

## Verification (prod, end-to-end)

1. **Direct spawn**: `spawn_subagent_create('persona-turn-tools', <Alma 32 faith
   question>)` → called `gospel_search` then `gospel_get`, returned a correct
   citation of **Alma 32:21** + the seed metaphor (vv. 27/41/42). ~5s,
   completed/verified.
2. **Through the gateway**: ran a local persona-host (isolated from the in-compose
   Starlet host) driving a throwaway "Computer" persona (my test creds, granted
   #Library on prod). Asked about D&C 121 → it answered with the exact **121:41**
   text ("No power or influence can or ought to be maintained by virtue of the
   priesthood, only by persuasion...") in ~12s. Cleaned up the throwaway after
   (revoked grant + key via the new AXR2 endpoints).

Read-before-quoting held: both quotes came from the tools, not memory.

## Gotchas

- The substrate's live `tool_defs` is richer than the docs claimed — `gospel_*`
  IS wired (a study-stage prompt said "don't try gospel_search," but it's active).
- `dispatch.go`'s 7th arg to `spawn_subagent_create` is `p_actor` (provenance
  label), NOT agent_family — the agent comes from the pipeline stage. So a tools
  pipeline pointing at `librarian` works without touching dispatch.go.
- persona-host go.mod is at `cmd/persona-host/` (each cmd/* is its own module);
  build with `go build -C cmd/persona-host`.

## Carry-forward

- **To bring the Library Computer online**: `chip-assistant=<its key>` in
  `extension/.env` CHATTERMAX_PERSONAS + restart persona-host, and grant it a
  library channel. The persona row + pipeline + agent are all ready. Michael holds
  the chip-assistant key (his to wire — or hand it over and I'll do the .env+restart).
- The in-compose persona-host binary predates the seed.go change; it'll pick up
  chip-assistant from the existing DB row regardless, and re-seed correctly on its
  next image rebuild.
- "Engineering bots" (code-writing personas) from the original AXR5 idea are a
  separate, larger step — they'd need coder_*/git_* tools, deliberately EXCLUDED
  from the read-only librarian. Future work, not this.
