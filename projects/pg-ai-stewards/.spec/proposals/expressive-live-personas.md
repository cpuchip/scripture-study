# Expressive, live personas — "chat as tool calls"

**Status:** DESIGN-ONLY (Michael's vision, 2026-06-09). Captures the answer to
"would chat-as-tool-calls change the arch?" + the multi-message / emoji-state /
real-time-reaction ideas. Big D&D payoff.
**Binding question:** How does a persona go from "one incoming message → one reply"
to a live participant that talks as it works, shows mood, and reacts in near-real-time —
and how much of that is substrate vs. the Go persona-host?

## The key architectural finding

**It's mostly on our side (persona-host + new substrate tools), NOT a substrate
arch change.** The substrate already runs a tool loop: the model emits tool calls
mid-turn, the bgworker executes them, feeds results back, the model continues. So
"the model talks as it works" = **make 'post to the room' itself a tool.**

```
today:   room msg → persona-host → ONE dispatch → ONE final reply → post
                                   (research_codebase etc. invisible to room)

wanted:  room msg → persona-host → dispatch that can room_say() mid-turn:
            room_say("🤔 hang on, searching…")   → posts NOW
            research_codebase(...)                → works
            room_say("found it: …  handler.go:71") → posts NOW
         (exactly how Claude Code emits text between tool calls)
```

## The mechanism (one tool + one outbox)

- **`room_say(text, mood?)`** — a substrate tool. The model calls it mid-turn to post
  to the room immediately. Implemented as: write a row to a `persona_outbox` table
  (room_id, persona, text, mood, ts). The persona-host already holds the gateway WS;
  it **drains the outbox and posts** (poll or LISTEN/NOTIFY). The substrate's
  dispatch/tool model is untouched — this is a new SQL fn + a persona-host consumer.
- **`mood`** is an optional emoji/state on `room_say` (or a standalone `set_mood`).
  "🤔 thinking", "😖 frustrated at the code", "😀". Frontend shows it next to the
  persona (and it just reads naturally inline in chat too).
- The **final** turn reply stays as today (the persona-host posts it), OR — cleaner —
  the persona-host stops auto-posting the final message and the model is told to
  `room_say` everything it wants the room to see. Decide at build: dual-path (auto-post
  final + optional mid-turn says) is the gentlest migration.

## Real-time reaction (#3) — the "pivot mid-turn" idea (Michael)

Michael's framing: while a model is mid-turn, newly-arrived room messages get added to
its context so it can pivot. The key realization: **a turn is already a LOOP of rounds**
— model emits tool calls → bgworker runs them → appends results → re-dispatches the model
with the updated message array → repeat until it stops. So there's a natural injection
point at every **round boundary** (not true token-level interruption, which we don't
need — the model re-evaluates each round anyway).

Three levels, cheapest first:

1. **`check_room()` tool (pull, no core change):** the model calls it to see "what's new
   since my turn started" and reacts. Fits the room_say pattern exactly; zero dispatch-
   loop change. Good for "let me check the table before I act."
2. **★ Auto-inject at the round boundary (push — the real "pivot"):** when the bgworker
   assembles the next round's messages, it also appends any room messages that arrived
   since the last round (as user turns). The model literally sees "[alice]: wait, I
   attack instead" mid-thought and pivots on its next round. This is a bounded change to
   the chat/tool_dispatch loop's message assembly — moderate, touches the core loop, but
   NOT a rewrite (the loop already rebuilds the array each round). **This is the D&D
   magic** — everyone rolls initiative, messages land while the model is working, and it
   weaves them in.
3. **True token-level interruption:** abort the in-flight LLM call to splice a message.
   Not needed; the round boundary is fine-grained enough. Don't build.

**So #3 is feasible and bounded:** level 2 is "the model pivots mid-turn," and it's a
contained change to the dispatch loop's per-round message build — the most invasive of
the expressive features (it touches the core loop, unlike room_say which is host-side),
but far from an arch rewrite. Recommend: room_say + check_room (level 1) first; level 2
when D&D wants true live-reaction.

## Why this is the D&D layer

- NPCs that post "🎲 rolls… a 17!" then narrate, express mood (😏 the smug merchant,
  😱 the spooked guard), and "think out loud" before acting. Initiative order = each
  persona takes a turn (check_room → act → room_say), the table sees it unfold live.
- The persona-host already buffers recent room messages and runs a turn per human
  message; room_say + check_room are additive to that loop.

## Cost / behavior guards

- room_say is cheap (a row write) but each one the model emits is still inside one
  dispatch — the turn's token budget + the stage soft-cap (5 tools) already bound it.
  A chatty persona spamming room_say is bounded by the same caps; add a per-turn
  room_say ceiling if needed.
- Mood/multi-message adds tokens to the persona's job — fine for kimi-tier, watch it
  for small models (qwen3.6-27b) where the budget is tighter.

## Build progress

- **✅ v1 foundation SHIPPED (r16, 2026-06-09, INERT):** `persona_outbox` table +
  `room_say(body, mood?)` tool, registered but granted to no agent. Smoke-verified.
  ★ Routing simplified: room_say keys on `_session_id` only — the persona-host already
  owns the session→channel map (`GatewayConn.channels[ch].sessionID`), so the drainer
  matches session→channel; no session_facets / channel-id in SQL. Live host untouched.
- **⏳ v1 NEXT (the live-host step):** persona-host drainer goroutine — poll
  `persona_outbox WHERE posted_at IS NULL`, for each row find the channel whose
  `sessionID` matches, `sendRaw({type:"message", channel, body})`, stamp `posted_at`.
  Then grant `room_say` to codewright/personas + re-prompt to narrate + mood. Rebuild
  persona-host (touches the live chat.ibeco.me personas — do with care, it reconnects
  clean). Grant + drainer must land together (inert until both).

## Build split (recommendation)

- **v1 (substrate + host):** `room_say(text, mood?)` tool + `persona_outbox` table +
  persona-host outbox drainer. Re-prompt personas (esp. roleplay/D&D) to narrate +
  mood. Dual-path (keep auto-final-post). This alone delivers #2 and the expressive layer.
- **v2:** `check_room()` for real-time reaction (#3 level 1) + frontend mood display
  (overlaps the people-mood frontend ask).
- **v3 (only if proven needed):** true mid-turn injection. Likely never.

## ★ v1.1 PLAN — the async turn loop (fixes "room_say posts late")

**The bug (diagnosed 2026-06-10, live):** a room_say beat created at 03:04:31 was posted
at 03:05:22 — 51s late, coincident with the final answer. Root cause: `handle()` calls
`takeTurn()` **synchronously inside the Run select loop**, and takeTurn blocks on
SpawnTurn/ConsultTurn for the whole ~51s turn. So the select loop is FROZEN during a
turn — the `<-drain.C` tick that posts room_say can't fire until the turn returns.
Compounding: on turn-zero, `cs.sessionID` isn't set until SpawnTurn returns, so even an
unblocked drainer couldn't route the message. (Same family as the Spin filler: a sync
step starving an async update.)

**The fix — run the turn OFF the loop, preserve single-goroutine ownership of
`gc.channels` via a results channel (no mutexes):**

1. `handle()` on a human message spawns a **turn goroutine** that does ONLY the cognition
   (SpawnTurn/ConsultTurn). It never touches `gc.channels`.
2. The goroutine reports back over a new `turnResults chan turnResult` that the Run
   select loop reads (a new `case`). The **loop stays the sole owner of `gc.channels`** —
   it applies results: set `cs.sessionID`, post the answer, `note()` it. No locks; the
   existing single-goroutine invariant is preserved, not fought.
3. **Early session id:** the goroutine sends `turnResult{kind:session, sessionID}` as soon
   as `spawn_subagent_create` returns the child id (session = `wi--<short>--turn`,
   derivable immediately) — BEFORE polling to completion. The loop sets `cs.sessionID`
   right away → the drainer can route room_say from the first beat.
4. **Per-channel serialization:** one turn at a time per channel (`cs.busy`). A human
   message arriving mid-turn is `note()`d (so it's in context) and a single follow-up
   turn fires when the current finishes if a human message went unanswered. The loop
   stays free for OTHER channels + the drainer throughout.
5. **Reconnect safety:** a per-connection generation counter; turn results whose
   generation != current are discarded (cs was reset on reconnect — the bug we just
   fixed makes this path real).

Result: room_say posts within ~1s of the model calling it ("🔍 let me check" → ~50s →
the answer), on turn-zero AND consults. Scope: persona-host gateway.go + dispatch.go
(an `onSession` callback on SpawnTurn). Moderate, concurrency-careful, fresh-eyes build.

## Not an arch change — the one-liner

Personas talking-as-they-work, showing mood, and reacting live is **a new tool
(`room_say`) + an outbox the persona-host drains + a `check_room` tool** — all on the
persona-host/substrate-tool side. The dispatch engine, the turn model, and the gateway
protocol stay as they are. The only thing that would touch the architecture (mid-turn
dispatch interruption) is exactly the thing we don't need.
