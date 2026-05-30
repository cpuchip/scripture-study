---
date: 2026-05-29
title: Batch M — model capability awareness + auto-probe (and the misdiagnosis it caught)
workstream: WS5
session_type: dev (decisions-up-front, gated phased commits)
status: M.1–M.5 SHIPPED + verified; auto-probe live on the watchman cadence
---

# Batch M — the substrate stops choosing models it can't use

## Where it started

After the `start_brainstorm` schema fix (rebuilt the MCP binary this session),
Michael asked me to check on qwen3.7-max and what glm-5 versions exist, and to
test each. The first-real-brainstorm-run empties (qwen3.7-max + glm-5 on the
disney lens) were the seed. He then asked: should we add a tool that displays
available models and connectors? — and chose the full scope twice: **substitute
+ log** on a dead model, and **build the auto-probe now**.

## What shipped (5 gated phases, all live-apply SQL + one Go build, NO pg rebuild)

A pleasant surprise: the whole thing avoided a pg rebuild. The dispatch
chokepoint is one SQL function; the bgworker already records completions to
`stewards.messages` and terminal status to `work_queue`, so the auto-probe is a
trigger, not a Rust change.

- **M.1** `model_capability` table + `model_usable()` + `model_catalog` view +
  `first_usable_model()`. usable=false is the ONLY gate; unrowed defaults usable
  (mirrors the J.11 cap gate — nothing working breaks).
- **M.2** capability substitution in `work_item_dispatch_stage` (J.11 body
  carried forward verbatim). An unusable resolved model is swapped for a usable
  same-provider one (catalog default → cheapest usable → raise) and logged via
  the l29 trigger, now the single writer with a `reason` column. Usable dispatch
  is byte-identical (no marker added when no swap).
- **M.3** `list_models` + `list_connectors` MCP tools (Go). Read DB-resident
  state, NOT `providers_loaded()` (its in-memory registry only exists in the
  bgworker process). Binary rebuilt — also carries the brainstorm schema fix.
- **M.4** `enqueue_model_probe` (direct work_queue insert, bypasses the M.2
  substitution so the real model is tested) + `trigger_resolve_model_probe`
  (on done/error, records the verdict). Tests the EXACT streaming path.
- **M.5** `enqueue_due_model_probes` + a guarded `AFTER INSERT` trigger on
  `watchman_passes` — probing rides the watchman cadence, pauses with the soak,
  never aborts a pass (errors swallowed). Self-throttling via staleness + dedup.

## The misdiagnosis the auto-probe caught (the real story)

Earlier this session I diagnosed glm-5/glm-5.1 as "streams empty via the
substrate" using a **shell-grep** SSE probe (`test-glm-qwen-models.sh`). I
committed that, annotated `model_pricing`, wrote it into the scripture-book
workflow doc, and reported it to Michael as fact.

M.4's first run overturned it. The substrate's real parser (`parse_chat_sse`)
extracts GLM's content fine — a substantive auto-probe returned **385 chars,
finish=stop**. My shell grep was the thing that failed to parse the stream, not
the substrate. So glm-5/glm-5.1 are **usable**; the brainstorm emptiness was a
per-lens token-budget/transient issue (a reasoning model can exhaust a tight
`max_tokens` before producing content — the non-streaming `finish=length`,
empty, 80-token case I'd also seen). qwen3.7-max is the one genuinely-unusable
opencode model (HTTP 401 whose body is "not supported for format oa-compat" —
the 401 and the oa-compat message my curl saw are the SAME rejection).

Corrected everywhere: m1 seed, model_pricing notes (j10 + 4a + live DB), the m2
smoke (now uses qwen3.7-max as the unusable example), and the scripture-book
workflow doc. The live `model_capability` was already corrected by the probe
itself.

This is the whole point of M.4, demonstrated on its first run: **build the
verification into the system and let it test the real path — ad-hoc diagnostic
tools lie.** Moroni 10:4 / Agans rule 9 (inverse hypothesis) caught a false
diagnosis I had already shipped.

## Verification

- M.1: 6 acceptance checks (model_usable true/false/default, first_usable picks
  free deepseek, unusable count).
- M.2: transactional smoke (ROLLBACK, $0) — usable path unchanged; unusable →
  substituted + logged once with reason; clean rollback.
- M.3: `go vet` + build clean; both queries verified against live DB (3 dead
  models sort last; gemini $17.96/$18 remaining).
- M.4: live probes ($0) — deepseek usable, glm-5 usable (385-char substantive
  probe), qwen3.7-max unusable (401 ×2).
- M.5: transactional dedup test — call1=3, call2=0 (dedup), no model
  double-probed, clean rollback; trigger installed + guarded.

## Carry-forward

- **Confirm the live auto-probe fires end-to-end on the next real watchman
  pass** — the trigger is installed + guarded and `enqueue_due_model_probes` is
  verified, but I have not yet watched a scheduled pass enqueue+resolve probes
  for the unprobed models (claude-*, gemini, minimax-m2.5, qwen3.5-plus).
  schedule_enabled=true, so it will. Watch cost on the first gemini probes
  (tiny, and capped).
- **M-batch SQL is live-only** (post-am1 pattern, like j8a–l29) — it persists in
  the data volume but is not in the `extension_sql_file!` chain / Dockerfile. The
  `stewards-cli migrate` ledger (the working system I'd missed earlier) covers
  fresh rebuilds. No action unless a fresh-volume extension-binary build is ever
  needed.
- A dedicated `model_probe_enabled` toggle (independent of the soak) if probing
  ever needs to pause without pausing the watchman. Not built — drop the trigger
  to disable.

## Pace note

This rode the tail of an already-enormous session (J.8–J.12, models catalog, 3
projects, Opus 4.8 review, gemini cap + budget surfacing, docs pass, competitive
research, 1828 UX, brainstorm schema fix, the glm/qwen diagnostic — and now the
full M batch + correction cycle). Michael drove the M batch with two explicit
scope choices, so it was responsive, not self-generated scope. But "less but
better" remains the live Sabbath question. Closing here deliberately — no M.6.
