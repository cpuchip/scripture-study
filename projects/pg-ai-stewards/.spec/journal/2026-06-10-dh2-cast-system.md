# DH-2: the cast system — one persona, many voices (built + proven same evening)

**Date:** 2026-06-10 (evening, Fable 5, same marathon day) · **Mode:** dev
**Trigger:** Michael: "lets good!" — kicking off DH-2 right after the session
close (the compact didn't happen; we rode the same context).

## What shipped

**Platform (`0e2f0b4`, deployed):** 0006 migration (one-alias UNIQUE → unique
by (persona, room, lower(name))); ResolveSubPersona **auto-creates on first
use** — the cast UX is declarative, "Grimble exists because the DM spoke as
him" (cap 50/room); message frames carry `subPersona`, attribution falls back
to the persona's own name rather than dropping a line; RoomCast + retire;
`cast` frame + subscribe backfill; roster nests cast under the voicing persona.
Integration-tested against real Postgres (auto-create, case-insensitive reuse,
attributed message renders the cast name, listing, retire).

**Substrate + host (r20, live + ledgered; host rebuilt):**
`room_say(as_character)` writes persona_outbox.sub_persona; drainer posts via
`emitAs` → platform subPersona; recent-notes record the cast name so the
persona sees its own characters in context. Race-clean attribution test.
dm-assistant seed prompt: voice NPCs via as_character, narration stays DM-voice.

**Wiring (test account, autonomous):** dm-assistant + NPC Ally created as
platform personas — display names MATCHED to host identities (the
Codewright/Chattercode lesson applied at create-time), respond_policy
`mentioned` from birth (no talking over Michael's game), granted into
Holodeck-3, keys minted + appended to the host .env without echoing.

## The live proof (one DM turn, three voices, ~15s)

> Grimble the shopkeep: "Back again, captain? The usual pickled herring…"
> Vex, guard captain: "Keep your wit in your pocket, Grimble…"
> DM Assistant: "The cramped shop smells of brine and mothballs…"

Auto-created, attributed, narration under the DM's own name. Starlet (policy
`all`) eyed the message and stayed silent; the DM woke only because it was
@-addressed. The decoupling principle held: the room saw characters, the
cognition was one kimi turn making three room_say calls.

## Addendum — Michael's first table session: four findings, four fixes (same night)

He played; the cast system met reality. Findings decoded from the transcript:
(1) Starlet (policy `all`) hijacked Grimble — minted HER OWN "Grimble the
shopkeep" (cast names were persona-scoped) and the customer got two Grimbles;
(2) her duplicate message = the coalesce-repeat watch item, live again;
(3) no way to @ a cast member; (4) "Grimble I need herring" couldn't wake the
DM (not his name).

**Fixes (chattermax `dc0e088` deployed + host rebuilt):** 0007 room-unique
cast names (dedupe keeps oldest claim → the DM's Grimble survived, Starlet's
deleted at boot); cast-name addressing (host parses cast frames via selfID
from ready, matches full + first names with stop-word guard, own-cast lines
never self-trigger) + cast in the @ popup; consult framing "never repeat
yourself"; **Starlet → Party** (Michael: "she was really only a test") — host
persona `party` (PCs as cast via as_character, dice honesty), platform persona
test-account-owned, policy `judgment`, pg-starlet removed from the host env.

**Proven live:** "Grimble, I'll take those two pickled herring" → DM 👀 at
0.1s (first-name routing!) → "Grimble the shopkeep: 😏 Six coppers? You're a
scholar and a saint" at 11.4s → Party eyed + correctly silent. First attempt
hit a Fireworks stream truncation (reasoning present, content/finish null,
4m47s) — host posted nothing, fault-tolerant; watch: substrate
retry-on-empty-stream. Typing stays persona-level (can't know the speaker
before the line lands) — explained, accepted.

## Carry-forward
- **Party persona** — same machinery; waits on dnd-tools sheets (DH-3).
- **Promotion to own-session** (villain with private memory) — when a campaign
  needs it; the display layer won't change.
- **DH-3 next:** dnd-tools repo scaffold (gh ready, pre-authorized).
- Root unpushed: 2 commits (host cast half + this journal's commit).

Michael's table now seats: Starlet (player, 22 on the strip), DM Assistant
(with a cast), NPC Ally — and the shop already smells of brine.
