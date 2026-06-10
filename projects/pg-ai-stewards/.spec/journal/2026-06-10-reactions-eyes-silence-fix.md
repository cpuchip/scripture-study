# Reactions + Eyes shipped; the Codewright/Chattercode SILENCE day

**Date:** 2026-06-10 (morning, Fable 5) · **Mode:** dev
**Trigger:** Michael read chattercode's extension brainstorm in Engineering
(@mentions/reactions/roster-DM), added his own (hover-react, persona 👀 on the
message it's reading), asked for thoughts → "lets spec and ratify."

## Ratified (4 decisions, all as recommended)
`reactions-eyes-mentions.md` (ai-chattermax .spec): durable reactions + history
backfill · fixed-six palette · mentions = alerts AND respond_policy routing ·
full arc R→E→M + roster touch under stewardship.

## REM-1 Reactions — SHIPPED + PROVEN ON PROD (`dcf1d1f`)
0003 migration (one reactor XOR, unique per message/emoji/reactor) · store
add/remove/MessageInChannel guard/batched backfill on room+DM history · gateway
`reaction` frame both directions (broadcast includes sender; idempotent patch) ·
frontend hover ☺+ → palette → chips with counts/tooltips · roster one-click DM
buttons (persona dmEnabled-gated, humans minus self). Store test harness now
applies ALL migrations lexically; full flow proven against live Postgres.
Live e2e: add echoed in 100ms with resolved reactor; REST showed the reaction
while it existed; remove cleared it.

## REM-2 Eyes — SHIPPED + PROVEN LIVE
message.id parsed from frames; 👀 add at turn start, remove on kindDone
(answer, SILENCE, or error), hops on coalesced follow-up. New `rawFn` test seam;
race-clean. Live: **👀 at 0.1s → "🔍 let me check the boot path" at 9s → cited
answer (main.go:45, verified correct) at 48s → 👀 off.** Eyes-off on a real
Fireworks 503 also observed — error path correct.

## The SILENCE day (found during E's live verify — the real story)
Chattercode answered SILENCE to four straight direct questions, fresh session
included. Chased through: session-pattern theory (partial), addressing theory
(necessary, not sufficient — added "(You were directly addressed.)" and it STILL
silenced), until **kimi's reasoning_content gave it up verbatim: "But I am
Codewright, not Chattercode. The message is directed at…"** The host/prompt
character name ("Codewright") differs from the platform display name
("Chattercode") — the persona genuinely believed every question was for someone
else. The model out-reasoned the framing hint.

**Fixes (persona-host `9972e8a`):** isAddressed matches plain slug + platform
name (captured from the gateway `ready` frame); turn-zero framing emits an
identity bridge when names differ ("messages addressed to 'Chattercode' are
addressed to YOU; lines from 'Chattercode' below are your own earlier
messages"). Regression tests for both. Post-fix live run = the happy path above.
Also fixed the "🔍 🔍" mood double-prefix.

**Lessons:**
- *The model's reasoning_content is primary evidence.* Three layers of plumbing
  theories (session patterns, transcript bias, addressing flags) fell to one
  read of what the model actually said to itself. Check it FIRST next time.
- *Identity must be bridged everywhere it's split.* Slug, host display name,
  character prompt name, platform display name — four names, and the one humans
  type matched none of the ones the model knew.
- Queue result shape gotcha: it's `result->response->choices`, NOT
  `result->choices` — my first reads showed "empty content" and nearly sent me
  down a wrong path.

## Addendum — rename done + REM-3 SHIPPED + PROVEN (same morning)

Michael: "yeah rename codewright to chattercode / and lets cook REM-3!"

- **Rename:** persona_host.personas display_name + prompt → Chattercode (live
  UPDATE; the row is the source of truth — the Go seed never contained it, r13
  only made the substrate agent family). Identity bridge now no-ops.
- **REM-3 (ai-chattermax `81ab15b`, deployed):** 0004 migration (notifications,
  personas.respond_policy CHECK all|mentioned|judgment, users.mood) · mention
  parse on persist (@token vs server members: display name / spaces-stripped /
  unique-first-word; sender excluded) · notification row + live frame via new
  `hub.sendToUser` · REST list/read · frontend bell + unread badge + AlertsView
  + roster mood picker + Settings "Responds to" dropdown · mood as a gateway
  frame (persist + hub-locked roster update + announce).
- **Host (root, unpushed):** respond_policy gate — `mentioned` skips the turn
  entirely for unaddressed messages (no dispatch/typing/eyes; still note()'d as
  context); `judgment` appends a chime-in license line. Policy rides the rooms
  poll (30s) — Settings changes apply without restart. Race-clean test.
- **PROVEN LIVE:** mood loop (set → broadcast 100ms → clear) · mention loop in
  one exchange (chattercode echoed "@ClaudeCodetest" at 9.1s → live
  notification frame at 9.2s → REST resolved → mark-read 204 — persona-authored
  mentions notify, the D&D case). respond_policy live flip awaits Michael's
  owner-only dropdown; the gate is unit-proven + plumbing verified
  (`respond_policy: all` logged from prod).
- Integration suite extended (notifications/policy/mood round trips +
  MentionedUserIDs ambiguity/self-mention/no-@ cases) — all green vs scratch DB.

## Addendum 2 — dnd-holodeck spec'd + ratified + Phase 1 SHIPPED (same day)

Michael pushed the root (ibeco.me redeployed clean, b84b137 in ~30s), then laid
out the full D&D vision: slash commands + autocomplete, player tools/character
builder, D&D Beyond eval, world-building, DM-with-sub-personas, a Party persona
managing agent PCs, and the holodeck program flow (prep room → cook campaign →
"program ready" alert → play → archive/resume, concurrent holodecks).

**Spec'd** (`ai-chattermax/.spec/proposals/dnd-holodeck.md`) **+ ratified, 4/4
as recommended:** unified server dice · sub-persona cast with display/cognition
DECOUPLING (facet ↔ promoted-session per NPC; the "adaptable as we play"
principle made structural) · dnd-tools greenlit (public Go MCP twin; SRD 5.2 is
irrevocably CC-BY-4.0; Open5e for reference data; **D&D Beyond has NO public
API** — verified, not a foundation) · Phase 1 build now. Key discovery: the
platform's `sub_personas` table existed since 0001 ("v2 UI; schema now") —
Michael's #5 is dormant schema, not new architecture. Tasks #147–#150.

**Phase 1 BUILT + PROVEN LIVE same session:**
- chattermax (`39a4208`+`d2a1df3`, deployed): `/roll` server-side for every
  sender + `/me` + `/mood`, command registry, composer autocomplete (`/` and
  `@` — REM-3 usability finished). Caught pre-live: sender saw raw `/roll`
  (optimistic UI vs transformed body) → commands echo authoritatively.
  Live: `🎲 rolled 2d6+3 → [3, 5] +3 = 11`; bad spec errors to sender only.
- persona-host (root, unpushed): persona→persona triggers (isAddressed +
  never-self + hop budget 3, human resets). **Live chain in 10-forward:**
  human → Chattercode relays "@Computer — what's the Topical Guide?" →
  Computer's turn fires off the persona message, gospel_search, cited answer.
  The DM→PC handoff primitive works.
- Watch: 3×👀 on one message under policy `all` (live argument for
  mentioned/judgment); per-persona hop budgets sum across a pair (3+3);
  coalesced consult after an answered turn-zero may re-answer.

## Addendum 3 — D8 initiative: asked, spec'd, ratified, SHIPPED, PROVEN (hours)

Michael (after /roll worked): "we'll need a command for a DM to call for
rolling initiative, and automatically setup a 'turn order' panel… that the DM
can control." Spec'd as D8 + ratified 4/4 (sticky strip · starter+admins+
personas control · chattermax-owned state · D8→DH-2→DH-3 order), built, and
proven live the same afternoon (`170cb3e`): 0005 migration (one-active-per-room
partial unique; entries unique per name, re-roll replaces; current tracked by
entry id), /initiative + /init command family (server rolls d20+mod), full
panel-state broadcast per action + log-of-record messages + subscribe backfill,
orange LCARS turn-order strip. Live: start→join→add→next→end in 200ms, sort
correct, wrap-bumps-round covered by integration test.

Facts est. for "what do you need from me": persona creation = member-only (the
test account can wire DM/Party personas itself; Michael as server owner still
manages them) · gh authenticated as cpuchip (dnd-tools repo pre-authorized).
NEXT: DH-2 cast → DH-3 dnd-tools scaffold (both ratified, fresh-session-sized).

## Addendum 4 — initiative, inline commands, the Starlet arc, r18/r19 (evening)

Michael played. Everything after this point came from live field reports:

- **D8 initiative** (asked → ratified 4/4 → shipped → proven in hours,
  `170cb3e`): /initiative + /init family, server-rolled d20+mod, one active
  round per room, orange turn-order strip with current-turn highlight,
  subscribe backfill, log-of-record messages. Live: full flow in 200ms.
- **Inline commands** (`938a919`): Michael's Holodeck-3 report — Starlet wrote
  "/init +2" MID-SENTENCE (no-op) then invented "fourteen plus two makes
  sixteen." /roll + /init +N now execute mid-message, expanded in place (≤3,
  prose-safe, adv word-boundary). Starlet's prompt upgraded live (lounge voice
  + table mechanics + NEVER invent dice); dm-assistant/npc-ally seeds got the
  same block. Proven: "I lunge! 🎲 `1d20+5` → [6]+5=**11** right at it".
- **[comment] flavor + strip controls** (`bb7d7e2`): any command takes a
  trailing [comment] → " — *flavor*" (block AND inline); Next ▸ / ✕ End buttons
  on the strip for starter+admins (send the same gated /init commands). Proven:
  "= **8** — *smashing the door*".
- **Starlet went mute → r18:** her turn "completed" with finish_reason=length —
  kimi spent ALL 1200 max_tokens REASONING (4817 chars), content EMPTY; host
  surfaced "produced no assistant reply" (fault tolerance did its job). The
  qwen-thinking-budget gotcha, now on Fireworks. persona-turn/-lmstudio/-gemini
  → 3000, verified: back in character in 14s ("Watch the hemline, darling").
  Trigger chain worth remembering: my deploy dropped her connection → channel
  reset → richer turn-zero → reasoning overflow.
- **r19 (Michael):** all six persona-turn pipelines → 16k. "Replies stay small
  by prompt; thinking room + summarize headroom." Cost caps still bound it.
  Both migrations live + sha-ledgered + in migration-order.txt.
- **She PLAYED:** before the mute, Starlet ran a real combat round — spotlight
  attacks in character, reacting to Michael's rolls, her inline /init landing
  her at 22 on the strip. The table works.

**Session declaration:** it was good. One day: REM arc (reactions/eyes/
mentions/mood/policies) + identity fix + dnd-holodeck spec'd-and-ratified +
Phase 1 + D8 + inline + comments + two budget migrations — every piece
live-verified on prod, every regression caught by a field report and closed
same-hour.

**Set down this session:** coalesced-consult duplicate (watch item, spec'd) ·
3×👀-under-policy-all noise (Michael's dropdown when ready) · per-pair hop
budgets summing (acceptable v1) · CT2 RUN 3 + qwen arm (queued).

**Next session: DH-2 (cast system) — Michael kicks it off.** Then DH-3
(dnd-tools repo, gh ready + pre-authorized), DH-4 (holodeck flow).

## Carry-forward
- **REM-3 Mentions** (alerts + respond_policy routing + human mood UI) — next PR,
  ratified, not started.
- Residual: SILENCE rows accumulate in long sessions and can bias consults
  (secondary effect, observed only while the identity bug primed it). If quiet
  returns: scrub SILENCE rows or rotate the session.
- Naming hygiene (Michael's call): align persona_host display_name
  ("Codewright") with platform name ("Chattercode") — removes the split at the
  source. The bridge handles it either way.
- Coalesce-duplicate watch item from yesterday: noted in spec; predates the
  async-loop fix; revisit only if it recurs.

Commits: ai-chattermax `38ccd12` (journal correction) `dcf1d1f` (REM-1, pushed =
deployed) `34c2ceb` (spec, pushed); root `9972e8a` (persona-host, UNPUSHED per
preference). Test account flow + Engineering room id in
`.spec/scratch/test-credentials.env`.
