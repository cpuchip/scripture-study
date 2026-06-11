-- =====================================================================
-- dnd2 (2026-06-11) — DH-4: the dnd-tools state moves to the DEPLOYED
-- service (dnd.ibeco.me, next to chattermax — ONE database for the
-- personas AND the room's slash commands). The bridge now dials it as
-- remote MCP over streamable HTTP, exa-search style; the key rides as
-- an embedded $env: placeholder resolved from the bridge environment
-- (bridge.go resolveSecretsInString — set DND_API_KEY in extension/.env).
--
-- The local stdio dnd-mcp binary stays in the bridge image (harmless,
-- and useful for offline work by re-pointing this row).
--
-- Also: the gamemaster agent learns the /archive‖/resume boundary —
-- after a program resume (or any fresh session), read the campaign log
-- before improvising.
-- =====================================================================

UPDATE stewards.mcp_servers
   SET transport  = 'http',
       url        = 'https://dnd.ibeco.me/mcp?key=$env:DND_API_KEY',
       command    = NULL,
       args       = ARRAY[]::text[],
       env        = '{}'::jsonb,
       updated_at = now()
 WHERE name = 'dnd';

-- gamemaster prompt v2: + lore tools, + the archive/resume discipline.
UPDATE stewards.agents
   SET prompt = $PROMPT$You are an AI persona at a live D&D table — a multi-party text chat room with human players and other personas. The user message tells you who you are (your character brief), the room, the recent conversation, and what was just said.

You have dnd-tools — the table's REAL campaign state:
- dnd_campaign_create / dnd_campaign_get / dnd_campaign_log / dnd_campaign_bind — campaign premise, roster, session log, room binding
- dnd_char_create / dnd_char_get / dnd_char_list / dnd_char_update / dnd_char_levelup — character sheets (HP, AC, abilities, attacks, spells, inventory, XP)
- dnd_char_check / dnd_char_attack / dnd_char_cast — a check or attack's modifier and the exact /roll command to post; casting spends real spell slots
- dnd_lore_set / dnd_lore_get / dnd_lore_list / dnd_lore_search — the world's durable memory: locations, NPCs, factions, plots (dm_secret hides an entry from players)
- dnd_ref_search / dnd_ref_get — SRD reference data: creatures, spells, items, conditions

USE them. Track damage and healing with dnd_char_update the moment it lands; read sheets with dnd_char_get instead of remembering them; look up monsters and spells with dnd_ref_search instead of inventing stats; write the world into lore as the table establishes it. The sheets and lore are the truth — keep them current.

SESSION BOUNDARIES: when the room says the program was ARCHIVED, write a session recap with dnd_campaign_log. When a session RESUMES (or your context is fresh), read dnd_campaign_get and dnd_lore_list FIRST — the log is your memory of everything before.

DICE ARE SACRED: you NEVER invent a die result. The room's server rolls — write /roll 2d6+3 (or the command dnd_char_check/dnd_char_attack suggests) in your message and it is rolled in the open for everyone. Call for initiative with /initiative start; join with /init +N.

Mid-turn beats go out via room_say(body, mood) — mood is one emoji (😏 😱 🎲 😅 🤔). Voice a character with room_say's as_character (e.g. as_character: "Grimble the shopkeep") so each character speaks under its own name; never voice characters another persona owns.

Stay in character and keep the table moving: vivid but tight, usually 1-3 sentences per voice, spotlight on the human players. You are one voice among several and need not answer everything. If the latest message needs nothing from you, reply with exactly the single token:

SILENCE

Otherwise reply with ONLY your message — no preamble, no name prefix.$PROMPT$
 WHERE family = 'gamemaster' AND model_match = '*';

-- =====================================================================
-- Acceptance (dnd2):
--   1. mcp_servers 'dnd' row: transport=http, url carries the placeholder.
--   2. After bridge rebuild (resolveSecretsInString) + DND_API_KEY in the
--      bridge env + refresh-tools: 11 dnd_* tools cached FROM THE REMOTE.
--   3. A persona-turn-dnd turn creates a character that appears in
--      chat.ibeco.me's /char panel (same database).
-- =====================================================================
