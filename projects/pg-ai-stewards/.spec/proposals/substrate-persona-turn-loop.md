# Substrate proposal — Persona Turn Loop (ai-chattermax #7)

**Status:** v1 **FULLY RATIFIED** (Michael, 2026-06-04) — triggers **#1 Reactive + #2 Addressed**, cognition = persistent session per (persona,room), runaway scope = **humans-only reactions**. **#3/#4/#5 + persona↔persona deferred to v2+.** Build-ready.
**Parent:** `substrate-persona-concept.md` (#6, BUILT — `cmd/persona-host`). Build lane: mine.
**Principle (Michael):** ship the simplest trigger set, get quick real results, let the rest earn their way in from what we learn.

---

## Cognition model (ratified)

A **persistent substrate session per (persona, room)**: spawn once when the persona joins, then `consult_subagent` each turn so the session's context accumulates (the Batch K/L context engine compacts it). persona-host drives this through the pgxpool it already holds — `spawn_subagent` is `SELECT stewards.spawn_subagent_create(...)` + poll `work_items`; `consult_subagent` re-asks the live session. No new MCP plumbing; the substrate stays general.

## The five triggers (what makes a persona *consider* speaking)

| # | Trigger | v1? | Notes |
|---|---|---|---|
| 1 | **Reactive** | ✅ v1 | a message posts → the persona's session judges "do I want/need to respond, or stay silent?" |
| 2 | **Addressed** | ✅ v1 | @room / @persona / @group mention → a strong "consider" signal |
| 3 | **Timed delay** | v2+ | per-persona pacing trait — a shy character waits ~20s, a quick one ~3s |
| 4 | **Ambient cron** | v2+ | periodic "thinking out loud" so a quiet room doesn't go dead |
| 5 | **Pipeline hook** | v2+ | fires on substrate work_item / state events |

## The gate (whether it *actually* speaks, given a trigger)

1. **Model judgment** — the persona's session decides respond-or-stay-silent (so #1 doesn't mean "reply to everything").
2. **Room hard rate ceiling** — the existing scheduler (`scheduler.New`, currently 10/min) is the v1 spam/runaway backstop.
3. **Pileup-avoidance** — v1 leans on the ceiling + judgment with two personas; real arbitration (whose turn when 4 agents all want to speak) is v2.

A per-persona **pacing profile** (response delay, talkativeness, quiet-period behavior) is what will eventually make one persona feel shy and another chatty — that's the v2 aliveness layer (#3/#4).

## v1 build scope (#7 v1 — simplest path to a live test)

- persona-host opens a **WebSocket client** to the room and speaks the AX3-2 envelope (`{sender, body}`); it connects as **`kind=agent`** (presence should not tag it Human).
- On each incoming message that isn't its own: feed it to the persona's session (`consult_subagent`) → if the session returns a message, post it; if it declines, stay silent.
- **#2 Addressed** = parse `@slug` / `@display-name` in the body → strong trigger. **#1 Reactive** = any other message → judgment.
- v1 e2e test: run ai-chattermax + persona-host locally, join `dm-assistant` to a room, post as a human, watch it respond in character and appear attributed (AX3-2). Then point at `chat.ibeco.me`.

## Runaway scope — RATIFIED: humans-only (v1)

**v1 personas react to HUMAN messages only** (Michael, 2026-06-04) — they ignore other personas' messages entirely. Zero ping-pong/runaway risk, simplest code (no exchange counting), fastest to a working test. The D&D magic of personas riffing off **each other** is a clean **v2** add with proper arbitration. So v1 trigger logic: incoming message → if sender is a persona, ignore; if a human, judge (#1) — and an @mention by a human (#2) is a strong "consider."

## v2+ (recorded, deferred — design later)

- **#3** per-persona timed-delay / pacing profile (shy vs quick).
- **#4** ambient cron — "thinking out loud" in quiet rooms; the substrate's quiet-period work (memory parse / intent refine / work-item propose) maps here.
- **#5** pipeline-state hooks — a persona reacts to substrate events.
- Real **pileup arbitration** for many agents.
- **Token-verify on connect** — the room verifies the persona JWT (#6 `/pubkey` is ready); a hardening pass, not MVP-blocking.

## Creation-cycle framing (book)
v1 is **Prescription/ground-truth made conversational**: the persona's turn is a real substrate dispatch, cost-tracked and gated, not a scripted bot reply. The aliveness layer (v2) is where **pacing becomes character** — the same self-governance the substrate already does (Sabbath/quiet-period), surfaced as personality.

---

*Written 2026-06-04. v1 triggers #1+#2 ratified; #3/#4/#5 deferred. Next: ratify the v1 runaway scope, then build persona-host's turn loop.*
