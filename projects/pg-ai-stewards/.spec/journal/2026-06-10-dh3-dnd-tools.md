# DH-3: dnd-tools — the PC trinity's third leg (built + proven same night)

**Date:** 2026-06-10 → 11 (overnight, Fable 5, same marathon day) · **Mode:** dev
**Trigger:** Michael: "lets pick up dh-3" — right after the DH-2 compact.

## What shipped

**The repo (github.com/cpuchip/dnd-tools, public, MIT, `2481fff`):** a Go MCP
server, strongs-mcp's twin, holding what nobody else offers — OUR campaign and
character state. Pure-Go SQLite (modernc, CGO-free so it cross-compiles into
the bridge); 11 `dnd_`-prefixed tools (one allow pattern grants the family):
campaigns + auto-numbered session log (the future /archive record), sheets on
the SRD 5.2 ruleset (class-derived HP and save proficiencies, standard-array /
CSV / JSON abilities, inventory with `x3 | notes` parsing, spell slots, level
up at the class average), Open5e v2 reference lookups (`srd-2024` default,
`srd-2014` option) with a read-through cache in the same SQLite file, and an
optional read-only HTTP sheet API. CC-BY-4.0 SRD attribution in the README.

**The dice line held:** `dnd_char_check` returns the modifier, the breakdown,
and the exact `/roll 1d20+5 [Name — Check]` string — the server NEVER rolls.
One dice implementation in the world (chattermax's crypto/rand), and now the
sheets feed it instead of competing with it.

**Substrate wiring (dnd1-mcp-seed, ledgered + pre-applied):** `dnd` mcp_server
(bridge image rebuilt; state at `/workspace/projects/dnd-tools/.data/dnd.db`
on the rw mount — inspectable from the host, gitignored); `gamemaster` agent
(R.9 librarian pattern: deny * / allow dnd_* + room_say, persona chat frame +
"dice are sacred"); `persona-turn-dnd` pipeline (tools on, 16k, kimi).
refresh-tools cataloged 11/11; `tool_permission('gamemaster', …)` shows
exactly dnd_* + room_say.

**persona-host:** dm-assistant + party seeded onto gamemaster/persona-turn-dnd
(party's `default_promote=true` is now seed-owned — UpsertPersona carries the
column, so DR rebuilds get it free); promoted-character turn-zero framing
points the character at its own sheet (dnd_char_get/check/update).

## Proven (three layers)

1. **Stdio smoke:** full tool walk — campaign → Thorin sheet → checks →
   damage/inventory → levelup → log → Open5e goblin (srd-2024) — all green.
   (Gotcha logged: mcp-go handles tool calls CONCURRENTLY; an all-at-once
   piped smoke races. Pace request→response.)
2. **Substrate e2e:** one persona-turn-dnd work item created the campaign +
   Vexa Nightbloom (halfling rogue: Stealth +5 = DEX +3 / prof +2, HP 9) and
   answered with the suggested roll — 15s, $0.017, verified+completed, rows
   confirmed in the SQLite file from the host side.
3. **Live Holodeck-3:** "@Party Vexa wants to slip past the goblin sentry —
   what does she roll?" → Party (sheet-backed): "Vexa's Stealth is **+5** —
   DEX +3, proficiency +2" with the `/roll` posted and inline-rolled in the
   open. ~12s.

## The bonus bug (live testing earns its keep again)

Run 1 woke the **DM** on a question meant for the Party: cast-name addressing
used plain substring matching, so cast member "Vex" matched inside "Vexa
Nightbloom". Fixed with word-boundary matching in `isAddressed` (one fix
covers the wake gate AND matchCast's promotion routing — they share it);
regression test added; verified live by work items, the ground truth: run 2
fired exactly ONE turn (Party's), zero for the DM.

Run 2's diagnosis also surfaced that a **stale DH-2-era host instance** had
been running alongside the new one (run 1 produced FOUR work items — each
persona fired on both its old and new pipeline). The taskkill swept both;
one instance confirmed by PID now. Lesson re-learned: after rebuilding a
host-side daemon, verify the OLD process count, not just the new start.

**Watch:** one phantom 👀 add/remove frame pair at subscribe+2s in run 2 with
no work item behind it — display-layer only, possibly an echo tied to the
duplicate-instance mess; look again next session if it recurs.

## Session declaration

It was good — DH-3 ratified-to-complete in one evening: a public repo, the
bridge wiring, the personas on their new pipeline, three verification layers,
and a real addressing bug found and closed by testing in the live room. The
holodeck now has sheets; the dice still belong to the room.

## Carry-forward
- **DH-4 (#150):** prep-room flow, program-ready chime, `/char` slash command
  on chattermax (needs a deployed dnd-tools HTTP surface or a persona detour),
  /archive + /resume wiring (dnd_campaign_log is ready for it), first campaign.
- `/init` with no modifier pulling DEX from the sheet (spec D8 tie-in).
- Spell slots are free-form (no auto seed by class) — fine for v1, revisit if
  a caster PC joins.
- Phantom-eyes watch item above.

## Set down
- D&D Beyond import (no API exists; flat character model leaves the door open).
- Form-based character builder (conversational-first per spec; HTTP viewer
  exists).
- Per-character model routing (still waiting on a spawn model param).
