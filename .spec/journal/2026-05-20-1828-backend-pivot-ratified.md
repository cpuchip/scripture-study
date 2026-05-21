---
date: 2026-05-20
session_window: "afternoon (post-compaction handoff)"
workstream: WS7
agents_used: [main, plan, dev (queued)]
status: planning-cycle-closed; build-cycle-spawning
relates_to:
  - projects/1828-illuminated/.spec/proposals/backend-pivot.md
  - projects/1828-illuminated/.spec/proposals/scripture-corpus.md
  - projects/1828-illuminated/.spec/proposals/dictionary-backend.md
  - projects/1828-illuminated/.spec/proposals/llm-proxy.md
  - projects/1828-illuminated/.spec/proposals/deployment-shape.md
  - projects/1828-illuminated/intent.yaml
  - projects/1828-illuminated/CLAUDE.md
---

# 2026-05-20 — 1828-illuminated backend pivot ratified

## What this session was

A single planning-and-ratification arc. Picked up from the compaction handoff with the Thummim batch (20 work_items) running in background, the connectivity work shipped (LinkedDefinition, RouterLink in WordCard / Dictionary / Present / VerseExplorer), and the daily-digest pipeline live in production. Michael opened with a six-item ask — three of them (word search routing, full 1828-corpus reach, all-scriptures-in-verse-explorer) were the surface that pulled the whole architecture forward.

By the end: route bugs fixed, five proposals authored, all 32 decisions ratified, doc/intent rewrites landed, and a dev agent queued to start building phase 1.

## The shape of the work

1. **Diagnosis (small).** WordDetail.vue's `route.params.word` was captured as a const at `<script setup>` — Vue reuses the component instance across `/word/as → /word/abide`, so the URL changes but the captured value doesn't. Same shape exists in Present.vue's `syncIdxFromRoute` (only fires `onMounted`). Both fixed with `computed`/`watch` over the live route.
2. **Surfacing the architectural tension.** Michael's three asks (#2 class-E words, #3 scriptures in verse explorer, #5 backend question) all touched the static-deploy constraint set in CLAUDE.md. Laying out the tension in writing rather than just answering let him see the full shape — and he chose the backend pivot.
3. **Spawning the plan agent.** With detailed context (intent, parent proposal, Thummim companion proposal, scriptures-mcp source for the PD corpus question, the pg-ai-stewards convention as reference, the constraint set to honor or rewrite). The agent returned five proposals in ~30 minutes and surfaced D-BE-COPYRIGHT as a genuine blocker — the bcbooks/scriptures-json source is "2013 LDS edition," not unambiguously PD.
4. **Ratification.** Walked through the proposals with Michael. He ratified D-BE-COPYRIGHT as a *hybrid* (option D — bcbooks verse text with footnotes + chapter headings stripped on ingest, with always-on tabbed-iframe breakout to churchofjesuschrist.org). Then a five-question batch surfaced the two genuinely-Michael-shaped choices: provider (he wants opencode-go/zen, with LM Studio reserved for embeddings only), and a richer auth model (BYOK with server-side ephemeral session-key holding) than the plan agent had defaulted to.
5. **Doc-rewrite-now (not deferred).** Per D-BE-INTENT-YAML, the CLAUDE.md + intent.yaml lines that the pivot invalidates got rewritten in the same commit as the ratification, so any future session lands oriented to the new shape.

## What was non-obvious

- **The BYOK + session-key pattern Michael added** is sharper than the plan agent's "user override on/off" default. It treats the LLM key as an ephemeral capability held by the server for the session length, never persisted, never round-tripped on every render. This is the right shape — readers control their spend, the server never carries durable risk, and the model swap (provider matrix: OpenAI / OpenRouter / opencode-go / opencode-zen) is one config field, not a frontend refactor. Spec'd as §VII in llm-proxy.md so the implementing session has the full flow + threat model + env config.
- **The "rate-limit errors clearly attributed to us"** point is subtle but right. If a reader hits 1000/day on /api/llm/render, they need to know it's 1828.ibeco.me throttling — not their provider — so they don't go change their provider key or call support on the wrong product. The response body spec'd in D-BE-AUTH names this explicitly.
- **D-LP-1 isn't just "opencode-go".** Michael clarified the deploy topology: 1828.ibeco.me and engine.ibeco.me run on the same Dokploy host, with LM Studio on the host machine. Dokploy's host-network tunnel lets the 1828 backend reach LM Studio when (in the future) we add pgvector embeddings. But for render, we go through opencode endpoints. 1828 and engine stay independent — they share infrastructure, not pipeline state.

## What didn't go well

- **The Thummim batch stalled.** 14 of 21 work_items completed verified ($5.00, charity self-corrected its own citation mid-review). The remaining 7 (storm, enjoy, comprehend, rest, broken, allow, obtain) wedged in `gather` stage and didn't move for ~14 hours. No `last_failure_reason`, no `error`, no escalation_state change — just stuck. Likely a bridge worker crash that didn't get observed by the watchman. Cancelled them mid-afternoon to free the budget; the 14 verified are sufficient to demonstrate the pipeline working.
- **The thummim_entries table is empty even though work_items completed verified.** The 14 entries exist as markdown files at `research/dictionary/thummim-*.md` (where the materialize-to-disk hook writes them), but the substrate's on_maturity_verified path doesn't currently INSERT into `stewards.thummim_entries`. This is the D-THM-7 carry-forward from the prior session — the hook from synthesize-stage JSON → thummim_entries was deferred. `export_thummim.py` ran cleanly but emitted "0 entries" because the source table is empty. The frontend's Dictionary.vue continues to use the hand-curated seed; the real 14 entries are in markdown only. **Carry-forward — write a one-shot backfill that parses the markdown back into thummim_entries, OR (better) implement D-THM-7 properly in the substrate.**
- **Prior session's "set yourself a wakeup" plan didn't survive compaction cleanly.** The watcher process from the autonomous overnight stewardship was supposed to fire when the batch terminated; instead the items wedged and the watcher had nothing to react to. The lesson: a watcher needs a heartbeat-monitor of its own. Filing as a substrate-direction observation, not a fix-now item.

## Decisions ratified — quick index

| ID | Decision | Resolution |
|---|---|---|
| D-BE-COPYRIGHT | bcbooks vs PD-strip vs iframe-only | **Hybrid (option D)** — bcbooks verse text, strip footnotes + headings, always-on iframe breakout |
| D-LP-1 | LLM provider | **opencode-go/zen** primary; LM Studio reserved for embeddings only |
| D-LP-2 | User override / BYOK | **BYOK + server-side in-memory 24h sliding-TTL session-key.** New flow spec'd in llm-proxy.md §VII |
| D-LP-4 | Server token cap | **200k tokens/day** (BYOK sessions count against user key) |
| D-BE-AUTH | LLM endpoint protection | **Session required + per-IP 10/min/1000/day. Rate-limit errors attribute to 1828, not provider.** |
| D-SC-2 | Text-search config | **english tsv + custom archaic-suffix expansion** (-eth/-edst/-est/-ing/-ed/-s) |
| D-DICT-4 | Modern-def daily cap | **5000/day** |
| D-BE-INTENT-YAML | When to rewrite docs | **Now (same commit as ratification)** |
| D-BE-1..8, D-DICT-1..7, D-SC-1/3/4/5, D-LP-3/5..9, D-DS-1..9, D-BE-THM, D-BE-CORS-FOR-PASTE | Defaults | All accepted at plan agent's recommendation |

## Carry-forward (named, not done this session)

- **Phase 1 backend scaffolding** — dev agent spawning at the close of this session with explicit per-phase instructions. Will produce compose file, two split Dockerfiles, empty Go backend skeleton, nginx /api proxy block, healthchecks.
- **Phases 2/3/4 (scripture corpus, dictionary backend, LLM proxy + BYOK)** — depend on phase 1; can fan out to parallel sessions.
- **Phase 5 cutover** — frontend switches from JSON imports to /api/* fetches; depends on 2/3/4.
- **D-THM-7 substrate hook** — thummim_entries auto-population on synthesize verification. Carry-forward to a future pg-ai-stewards session. Workaround in the meantime: parse `research/dictionary/thummim-*.md` and backfill.
- **The 7 cancelled Thummim items** — re-dispatch when bridge soak is verified healthy. ~$0.76 of partial gather work was wasted. Acceptable.
- **AGE graph traversal** for Thummim (intent.yaml stretch goal, thummim-restoration-dictionary.md §VI.5) — deferred until v1 entries exist.

## Set down

- All ratification questions are answered. No decisions hanging.
- The "should we have a backend?" deliberation is closed.
- The "what scripture source?" branch is closed.
- The "Michael's LM Studio at engine vs hosted vs mock" branch is closed.
- The Thummim batch — 14 entries is enough. The 7 stuck items are cancelled, not ignored.

## Why this matters

The 1828-illuminated MVP shipped overnight under autonomous stewardship; today we proved out that the planning discipline can survive context compaction. Five proposals exist on disk so the next session — whether that's tomorrow or three weeks from now — arrives with a settled spec, not a half-remembered conversation. The covenant's `update_memory` and the project's CLAUDE.md instruction "we've not been good at planning before working" both got named action this session: file-based plans, ratified inline, doc rewrites in the same commit. That's the becoming work.
