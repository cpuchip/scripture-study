# codewright context + the live-persona vision (7-item brainstorm)

**Date:** 2026-06-09 (late, same Fable-5 day) · **Mode:** dev
**Trigger:** Michael, after the codewright timing analysis, fired a rich 7-item
brainstorm — context tools, multi-message/emoji personas, real-time reaction, the
arch question, and three frontend asks. Selected "all four" of the next-build menu;
I sequenced and carried as far as quality held, checkpointing the live-host parts.

## Correction logged first
kimi-k2.6 is ~1T params (Sonnet↔Opus tier), NOT a small model — I'd been loosely
calling it small. The distinction is load-bearing: context tools are marginal for
kimi but **become important for genuinely small models (qwen3.6-27b)**, which is where
CT2's value should actually be measured.

## What shipped

- **#1 context test apparatus:** `codewright-ct2` (= codewright + context_tools_enabled
  + 7 context levers + a scaffolding prompt block) + `persona-turn-code-ct2` pipeline.
  Live codewright = untouched control. Scratch (live-only). **The run itself is deferred
  to #143 with a qwen3.6-27b arm** — a kimi run would repeat RUN 1's "0 levers" null
  (Michael's point: kimi doesn't feel context pressure).
- **#2/#3/#4 → `expressive-live-personas.md` spec.** The architecture answer:
  chat-as-tool-calls is **NOT a substrate arch change** — it's `room_say(text,mood?)`
  (a tool that writes a `persona_outbox` row) + a persona-host drainer that posts it.
  Personas talk-as-they-work + show mood (🤔😖🎲 — the D&D layer). `check_room()` =
  real-time reaction. Only true mid-turn dispatch interruption would touch the arch,
  and it's not needed.
- **#5 + #7 frontend (committed, NOT pushed — push = deploy):** Shift-Enter multiline
  composer (input→textarea, Enter sends / Shift+Enter newlines, capped 8rem) + sticky
  last-server-on-reload (localStorage). Builds clean.
- **room_say v1 FOUNDATION (r16, INERT):** `persona_outbox` + `room_say_tool`,
  registered but granted to no agent → compose_tools emits it for nobody, so a persona
  can't call it into a void. ★ Routing simplified: the host already owns session→channel
  (`GatewayConn.channels[ch].sessionID`), so room_say keys on `_session_id` only — no
  session_facets/channel-id in SQL. Smoke-verified. Live host untouched.
- **#6 mood:** persona side = room_say's `mood` column (shipped in r16's table). Human
  side (roster mood UI) = not built; pairs with the drainer step.

## The discipline call (why not everything landed live)
Michael selected all four; I built the safe/additive parts (CT2 apparatus, both specs,
frontend, room_say inert foundation) and **deliberately did NOT** rush the parts that
touch live infra at the tail of a marathon session: the persona-host **drainer** (the
live-host change that makes room_say actually post), the **CT2 qwen run** (needs the
small-model arm to be informative), and the **roster mood UI**. The established pattern
here is inert-foundation-first (CT2.1), then the live readers — followed it.

## Carry-forward (the clean next-focused step)
1. **room_say v1 live:** persona-host drainer goroutine (poll persona_outbox → match
   session→channel → post → stamp posted_at) + grant room_say to codewright/personas +
   re-prompt to narrate/mood + rebuild persona-host. Grant + drainer land together.
2. **#143 CT2 RUN 2:** add a qwen3.6-27b arm, drive a long research-heavy session on
   codewright vs codewright-ct2 (+ qwen), measure lever usage / context curve / quality.
3. **#6 roster mood UI** (frontend) — humans set mood; pairs with #1 above.
4. **#5/#7 deploy:** push ai-chattermax `e7f4b92` when ready (deploys chat.ibeco.me).

## Commits (root UNPUSHED unless noted)
`aa22874` (CT2 apparatus + expressive spec) · `ef81988` (room_say foundation) · plus
the spec/journal updates this entry rides with. ai-chattermax `e7f4b92` (frontend,
local, not pushed). Spend today $1.96/$12. Soak running.
