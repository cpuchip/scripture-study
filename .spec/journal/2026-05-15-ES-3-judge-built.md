---
date: 2026-05-15
mode: build + verify (Emergency Stop, ES.3)
workstream: WS5 (substrate)
project: pg-ai-stewards
title: "ES.3 built — the judge-compiled-brief replaces leaf-chunk-and-embed; s1-s4 shipped + verified, soak left PAUSED for Michael's resume"
status: ES.3 s1-s4 SHIPPED + verified (real judge call confirmed). s5 deferred. Soak PAUSED — resume is Michael's call. ES.4 (full bacteriopolis re-run) pending.
carry_forward:
  - "RESUME THE SOAK — left paused deliberately. ES.3's judge path is verified component-by-component + one real LLM call, but a full oversized-fetch inside a live multi-stage pipeline has not run unattended. Michael should resume (UPDATE stewards.watchman_config SET schedule_enabled=true WHERE id=1) after a look."
  - "consult_subagent is built + deployed (Go handler, bridge rebuilt) but NOT granted to any agent — agent_tool_perms is deny-by-default. It is inert until granted. Granting it (which agents get re-engagement power) is a deliberate decision. The judge brief surface text already advertises consult_subagent, so granting it to research/study agents closes that loop."
  - "ES.3.s5 — model-name normalization (kimi-k2.6's 3 gateway identifiers) — deferred, optional, carry-forward."
  - "ES.4 — full bacteriopolis re-run under the judge path (the inverse-hypothesis verification). The component verification is strong; ES.4 is the live-pipeline confirmation."
  - "cost_usd still unpopulated on work_queue rows (judge chat showed cost=(unset), tokens_in=18645). Pre-existing finding, not an ES.3 regression — but cost discipline needs it."
  - "Orphan engram_embeddings rows from the dropped leaf path may remain — low-priority corpus hygiene."
links:
  - "../../projects/pg-ai-stewards/.spec/proposals/substrate-ES-emergency-stop.md"
  - "../../projects/pg-ai-stewards/extension/es6-engram-provenance.sql"
  - "../../projects/pg-ai-stewards/extension/es7-judge-brief-dispatch.sql"
  - "../../projects/pg-ai-stewards/extension/es8-consult-subagent.sql"
  - "../../projects/pg-ai-stewards/extension/es9-drop-leaf-index.sql"
---

# 2026-05-15 — ES.3 built: the judge-compiled-brief

Michael ratified ES.3 in a two-round council, then: "open the build
window now. I'm heading to bed — finish the block of work, then return
and report." This is the build.

## What shipped

| Commit | Phase | What |
|---|---|---|
| `cc8fde9` | ES.3.s1 | engram `provenance` field (extracted\|inferred) |
| `84209ea` | ES.3.s2 | judge-compiled-brief — dispatch + intercept rewrite |
| `2f6c25a` | ES.3.s3 | consult_subagent — re-engage a persistent sub-agent |
| `c44ddbd` | ES.3.s4 | drop the leaf index (destructive — ratified) |

The judge replaces the leaf-chunk-and-embed compaction (CF-6). An
oversized tool result is preserved whole in `messages_raw_overflow` and
handed to ONE deepseek-v4-flash chat that reads it against the binding
question and returns a compiled brief — ≤7 provenance-tagged engrams +
state + a discarded note. The L.1.1.8 intercept no longer chunks; it
spawns the judge and replaces the content with a `[JUDGE-PENDING]`
placeholder. `tool_dispatch_complete_waiting` gains one additive branch:
when a tool message is `[JUDGE-PENDING]` the parent's continuation is
NOT enqueued — `apply_judge_brief` resumes the gated parent once the
brief is ready (decision 2: always sync). One LLM call where the old
path fired hundreds.

## The design call worth recording

The ratified proposal said the judge would route through
`spawn_subagent_create` as a work_item. Building it, I deviated: the
judge runs as a **bare chat** (the proven K.1 extract_engrams pattern —
own session, manual body, completion trigger) rather than a full
work_item + pipeline. Reason: the intercept is a trigger; threading a
work_item + maturity ladder + the spawn_subagent Go polling handler
through a trigger context is heavyweight and risk-prone. The bare-chat
pattern is pure SQL, no bgworker.rs change, and was already proven by
K.1. The re-engagement generalization is preserved — `consult_subagent`
keys on the **session** (the judge's `judge-<msgid>` session), and a
session is the unifying handle whether the sub-agent is a judge or a
spawn_subagent child. This is a faithful realization of the intent
(re-engage any sub-agent) by a different mechanism than the literal
proposal text. Flagging it for Michael's review.

## The bridge crash-loop (caught + fixed)

`smoke-es7.sql` was committed into `extension/` — and the migration
ledger (run by the bridge on every startup) executes every `.sql`
there. A smoke file's psql `\` meta-commands are not valid SQL → the
bridge crash-looped on startup. Caught it in the logs, moved the smoke
file to `extension/smoke/` (the ledger skips subdirectories — and there
was already a `smoke/` dir, the established convention I should have
used). Bridge recovered: `migrate: substrate is current`. Lesson:
anything in `extension/*.sql` is a migration. Smoke/test SQL goes in
`extension/smoke/`.

## Verification

Four passes:
- **es7 synthetic** (rolled back) — intercept → judge dispatch → brief
  write → parent resume; stray-K.1 skip + provenance confirmed.
- **es8 synthetic** (rolled back) — consult_subagent_dispatch rebuilds
  the document into the re-engagement body; soft cap fires on the 6th.
- **gating synthetic** (rolled back) — `tool_dispatch_complete_waiting`
  itself: the normal path enqueues the continuation, the judge path
  withholds it. The hot-path change verified on both branches.
- **REAL judge call** — a real 72,656-char document (the iron-rod
  study) → real intercept → real deepseek-v4-flash call → a real brief:
  7 well-formed engrams, verbatim quotes in `preserved`, all
  `provenance: extracted`, an honest `discarded` note ("pastoral
  applications, personal addresses to Michael, supporting citations").
  18,645 tokens in, one call. The 7 engram embeds all routed to
  `lm_studio`. Test artifacts cleaned up; queue empty.

## Why the soak is paused

The judge path is verified component-by-component and by one real LLM
call. What has NOT run: a full oversized fetch inside a live
multi-stage pipeline, unattended. We just came out of an emergency stop
caused by exactly that class of thing. The cost of leaving the soak
paused a few hours is near-zero; the cost of resuming onto a subtle bug
is real. So the soak stays paused — Michael resumes it after the report
and an optional look. This matches the action's reversibility to the
verification level; it is not timidity.

## The arc

Two days, one continuous thread: Batch L → L.1.1 → L.1.1.x post-mortem
→ ES (emergency stop) → ES.1 (stabilize, verified) → ES.3 (the
rearchitecture). ~70 commits, zero rollbacks. The bleed class is closed
AND the substrate is now thoughtful — an oversized fetch produces a
compiled brief, not 500 embedded leaves. The judge sits down with the
net and sorts (Matthew 13:48).
