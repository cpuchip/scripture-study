---
date: 2026-06-04
title: R.7 persona-turn pipeline + the persona-host turn loop (#7 v1) — a persona talks
---

## What happened

Built ai-chattermax #7 v1: a persona now holds a live presence in a chat room, backed by real substrate cognition. The payoff moment of the whole arc — the first time an AI persona actually talks in the room, not a scripted bot reply but a cost-tracked, gated substrate dispatch.

**The path-check that opened it (the thing I flagged at last handoff):** can persona-host drive a turn straight from its pgxpool, or does the re-ask need a Go/MCP path? Answer: pure SQL. Both `spawn_subagent` and `consult_subagent` MCP tools are thin sync-poll wrappers over `stewards.spawn_subagent_create` / `stewards.consult_subagent_dispatch`. persona-host already holds a pgxpool, so it replicates exactly that — no new plumbing, the substrate stays general. Exactly what the proposal predicted.

**R.7 substrate side** (`extension/r7-persona-turn.sql`, live-applied):
- a `persona` agent — a thin chat-posture meta-prompt with a `SILENCE` escape hatch (the model judges whether to speak), 7 tool-perm denies.
- a `persona-turn` pipeline — single-stage clone of `redline`: tools-disabled, kimi-k2.6, max_tokens 1200, `input_template={{input.binding_question}}`. The persona's character rides in the binding question (carried forward by the session), so one pipeline serves every persona.
- **the adjacent-surface catch:** `on_one_shot_pipeline_completed` only auto-verified `aggregate-children`/`brainstorm-%`/`redline%`. A `persona-turn` child would finish `completed` but stall at `maturity=raw`, and the host's spawn poll (waits for `verified`) would hang the full 20-min timeout on *every* turn-zero. Extended the trigger to qualify `persona-turn` — the same j6/R.6 pattern applied once more. Without the foresight pass this would have silently broken the live test.

**Host side** (`cmd/persona-host/`):
- `dispatch.go` — `Cognition.SpawnTurn` (turn zero: spawn → poll work_items to verified → reply + session id) and `ConsultTurn` (re-ask → poll work_queue to done → reply). `$0.10`/turn cap; `IsSilence` gates a stay-quiet reply.
- `turnloop.go` — `RoomConn` dials `?id=<display name>&kind=persona`, reads the AX3-2 envelope, runs a turn per HUMAN message (own + other-persona messages filtered — humans-only), parses @mentions as a strong hint, posts the reply or stays silent. Read pump + worker split so a blocking dispatch never stalls reading.
- `autojoin.go` — env-driven (`CHATTERMAX_WS_BASE` + `PERSONA_AUTOJOIN=slug@room`) with a reconnect supervisor.

## Verification

**SQL smoke (proved cognition before writing Go):** spawn a persona-turn for `dm-assistant` → `completed/verified`, in-character tavern scene, ~$0.014. consult the same session with the player's next action → continued coherently (remembered the bard, the warmth, introduced a barkeep + quest hook — context accumulated). An unaddressed message → `SILENCE`. All three gates of the turn model proven in raw SQL first.

**Live e2e (the real thing):** chattermax-server + persona-host + a throwaway human WS client, all against the live substrate. Human posted "DM Assistant, set the scene…" → DM Assistant replied in character, attributed to its display name. Roster confirmed `Kind=persona`. And a gift I didn't script: the human's message landed *before* the persona finished connecting (startup race → "bad handshake" → the reconnect supervisor re-dialed 5s later) — yet the persona still answered, because **AX3-2 replay-on-join** delivered the backlog on connect. The protocol fix from earlier this session served #7 exactly as designed.

## Decisions / deviations (none reverse a ratified call)

- **Character in the binding question, not a per-persona system message** (v1). Cleaner than spawn could do anyway (spawn can't inject persona_prompt); a proper system-message + `model_override` honoring is the v2 upgrade. `dry_run_chat` composes `role='system'` session messages, so the v2 path is known.
- **r7 lives in the core extension, not persona-host.** The pipeline/agent are substrate cognition config (general primitives); persona-host *calls* `stewards.*` (a blessed API) but doesn't define substrate config — keeps the sidecar boundary clean.
- **Live-applied, pending fold-back** into lib.rs/Dockerfile at the next rebuild — consistent with siblings r1–r6 (all live-only right now).

## Carry-forward

- **v2 aliveness layer** (deferred, ratified): #3 timed-delay/pacing (shy vs quick), #4 ambient cron ("thinking out loud" in a quiet room), #5 pipeline-state hooks, real pileup arbitration, persona↔persona reactions.
- **v2 cognition polish:** per-persona system prompt + `model_override` at dispatch; batch consecutive messages into one turn (cost lever); an envelope `kind` field so humans-only is robust without name-matching.
- **Hardening (not MVP-blocking):** persona-host deploy container + scoped `persona_host` DB role; token-verify on WS connect (the `/pubkey` is ready).
- **Soak was paused for the build, resumed at close.** Dev DB carries a few verified `persona-turn` smoke/e2e work_items (harmless evidence).
