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

## Real-time reaction (#3) — two levels

1. **`check_room()` tool (pragmatic, fits the pattern):** the model pulls "what's new
   in the room since my turn started" mid-turn and reacts. Gets you "personas take
   turns while messages stream in" (D&D initiative) with no turn-loop rewrite.
2. **True mid-turn injection (the only real arch change):** interrupt a running
   dispatch to splice in a newly-arrived message. Hard (the dispatch is a fixed
   message array) and probably unnecessary — level 1 covers the use case. Flagged, not
   recommended for v1.

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

## Not an arch change — the one-liner

Personas talking-as-they-work, showing mood, and reacting live is **a new tool
(`room_say`) + an outbox the persona-host drains + a `check_room` tool** — all on the
persona-host/substrate-tool side. The dispatch engine, the turn model, and the gateway
protocol stay as they are. The only thing that would touch the architecture (mid-turn
dispatch interruption) is exactly the thing we don't need.
