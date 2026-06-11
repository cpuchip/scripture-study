# DH-4: the holodeck flow — sheets behind every slash command (built + deployed + proven overnight)

**Date:** 2026-06-11 (overnight, Fable 5, the marathon's fourth arc) · **Mode:** dev
**Trigger:** Michael: "lets do dh-4. we're getting close!" + his asks: /attack
with weapon resolution, skill/stat slash checks, world-building tools, full
sheet depth (equipment/spells/levels), /char as an editable panel with HP
chips. Ratified 4/4: deploy dnd-tools as a service · /attack+/check+/save+
/cast+/hp + the /char panel · attacks+spells+conditions · dnd_lore.

## The architecture move that unlocked it

chattermax (NOCIX) couldn't reach sheets living in the local bridge's SQLite.
**dnd-tools became a deployed service** (dnd.ibeco.me, sidecar in the
chattermax compose, built straight from the public repo, SQLite on a volume,
bearer-key gated): chattermax's slash commands call its JSON API server-side;
the LOCAL substrate bridge dials the SAME service as remote MCP over
streamable HTTP (the exa-search transport; bridge now resolves embedded
`$env:` placeholders in http URLs so the key lives in extension/.env, not the
DB row). One database under every surface — the dice stay chattermax's.

## What shipped

- **dnd-tools v0.2** (`7c51d0f`): structured attacks (ability+prof+magic
  derivation), known spells (cast spends real slots; Open5e damage_roll
  surfaced), conditions; dnd_char_attack/cast; dnd_lore_* (dm_secret rows
  NEVER served on the player HTTP surface — tested); campaign↔room binding;
  HTTP API v2 (auth, by-player resolution, resolve/cast/hp/PATCH endpoints);
  /mcp streamable endpoint; -serve mode + Dockerfile. Tests across rules/
  store/httpapi/mcpserver incl. the auth gate and secret-exclusion.
- **chattermax** (`fec69ba`, deployed): the command family — /attack parses
  "target with weapon", rolls to-hit from the sheet, hands back the damage
  roll for after the DM's call (Michael's exact flow); /check + /save;
  /cast (rolls known damage dice inline); /hp [name]. **/char opens an
  editable character panel** (ScripturePanel mold; PATCH through a chattermax
  proxy so the API key never reaches the browser; edit gate = the sheet's
  player or a room admin). HP chips on roster cast + player names (click =
  open sheet). /archive + /resume → `program` frame (admins + personas).
  Registry-driven autocomplete picked up all 8 commands with zero frontend
  changes — D3's design paying out exactly as drawn.
- **persona-host:** on `program` archive — one closing turn ("write the
  session recap with dnd_campaign_log"), then session rotation; promoted
  character sessions persist (a mind is room-agnostic — DH-2's principle
  held at the boundary). gamemaster prompt v2: lore tools + "when a session
  resumes, read dnd_campaign_get + dnd_lore_list FIRST."
- **Deploy plumbing via the NOCIX Dokploy API:** compose env key set,
  dnd.ibeco.me domain created (wildcard DNS already pointed home),
  letsencrypt issued; dnd2 migration ledgered (transport flip + prompt v2 —
  one re-apply: stewards.agents has no updated_at column).

## Live proofs (all on prod)

1. **Table setup through the remote:** one Party turn created The Brine Cave
   Run, BOUND it to Holodeck-3, built Vexa Nightbloom (Dagger + Fire Bolt,
   player "Claude Codetest"), wrote a lore entry — 10s / $0.024.
2. **The command family at ~0.1s each:** /check stealth → 🎲 [4]+5=9 with
   breakdown · "/attack the goblin sentry with dagger" → [18]+5=**23** to hit
   + `/roll 1d4+3` handed back · /cast fire bolt · /hp -3 → 6/9.
3. **/char proxy** served the full sheet to a member session.
4. **★ State unification:** the human's /hp -3 was read back by the PARTY
   PERSONA through its own substrate tools as **6/9 HP** — chat command and
   persona cognition, one truth. (Party also free-styled an in-character
   Vexa beat via room_say — the cast machinery riding along unprompted.)

## Bugs the build caught

- **The public repo had no main package.** An unanchored `dnd-mcp` gitignore
  line matched the `cmd/dnd-mcp` SOURCE directory; every local build masked
  it (workspace COPYs, not git). The FIRST build from the git context failed
  in seconds. Lesson: building from the published artifact is itself a test —
  root-anchor binary ignores (`/dnd-mcp`).
- mcp-go handles tool calls CONCURRENTLY over stdio — a piped smoke races;
  pace request→response (caught in DH-3, re-confirmed designing the client).

## Declaration

It was good — the holodeck's machinery is COMPLETE: prep room (gamemaster v2
+ lore tools), program-ready chime (mentions), play (sheets behind every
command, dice in the open), archive/resume (program frames + log + rotation).
Four arcs in ~one day: REM → DH-1 → DH-2 → DH-3 → DH-4, every layer deployed
and live-verified.

## Addendum — room gating: the binding is the switch (same night)

Michael: "does every room have the dnd machinery? should rooms have on/off
settings?" Honest answer: half-gated — the binding already gated *function*,
but autocomplete advertised the commands everywhere and /archive//resume
leaked program frames into any room. Ratified: no separate feature flag —
**the campaign binding IS the switch**, surfaced three ways: `/dnd enable
[name]` (bare form auto-names a campaign after the room) / `/dnd disable`;
`/campaign [bind|unbind]`; a Settings "D&D Campaign" row. Autocomplete shows
the sheet-command family only in bound rooms (registry grew a `group` field —
the D3 registry design absorbing its first metadata without breaking
anything); 🎲 campaign chip in the room header; archive/resume refuse unbound
rooms; a `program: state` frame refreshes clients on bind/unbind.

dnd-tools 0.2.1 (`cf461ab`): POST /api/campaigns + PUT /api/rooms/{id}/campaign
(bind-creating-if-needed / unbind). chattermax `eca5e76`. Both deployed,
freshness proven by version markers (the stale-build discipline). **Live gate
proof at 0.2s on the test account's own server:** enable → 🎲 bound →
/check functional (needs-a-sheet error) → 🗺 → disable → 🚪 → /check refused.
Generic /roll + /init stay global on purpose — dice belong to every room.

## Carry-forward (Michael's table)
- **/archive + /resume live-proof** — needs a room admin; the gate correctly
  refuses members. Natural first use: the end of campaign session #1.
- **The first real campaign** — prep room ritual → chime → play. Everything
  is loaded for it.
- Browser eyeball of the /char panel + HP chips (data path verified; the Vue
  render awaits human eyes).
- Watch: /init with no modifier pulling DEX from the sheet (D8 tie-in, small).

## Set down
- Per-character model routing (still waiting on a spawn model param).
- Persona-side lore secrecy enforcement (prompt-level v1; tool-level needs
  caller identity in MCP).
- Local .data/dnd.db (DH-3 test data) — retired; the deployed volume is the
  table's memory now.
