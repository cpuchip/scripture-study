-- =====================================================================
-- dnd1 (2026-06-10) — DH-3: register dnd-tools MCP + the gamemaster
-- agent + the persona-turn-dnd pipeline (dnd-holodeck Track D6).
--
-- dnd-tools (github.com/cpuchip/dnd-tools) is the campaign/character
-- STATE server — sheets on the SRD 5.2 ruleset, session log, Open5e
-- reference lookups — cross-compiled into the bridge image at
-- /usr/local/bin/dnd-mcp (strongs pattern). State lives in SQLite on
-- the bridge's rw /workspace mount (.data/ is gitignored in that repo).
--
-- The gamemaster agent is the R.9 librarian pattern pointed at the
-- table: deny * + allow dnd_* + allow room_say (cast voices). The
-- persona-turn-dnd pipeline name starts 'persona-' so the R.8 one-shot
-- auto-verify trigger fires on completion (else the persona-host's
-- spawn poll hangs to timeout).
--
-- Applied via the migration ledger (idempotent). After this lands:
-- rebuild the bridge image (new binary), then `bridge refresh-tools`
-- (grant ≠ catalog), then point dm-assistant/party at the pipeline
-- (persona-host seed.go does this from its side).
-- =====================================================================

-- ---------------------------------------------------------------------
-- 1. dnd-tools MCP server registration.
-- ---------------------------------------------------------------------
INSERT INTO stewards.mcp_servers (name, description, transport, command, args, url, env, enabled)
VALUES (
  'dnd',
  'dnd-tools — D&D campaign + character state on the SRD 5.2 ruleset. '
    || 'Tools: dnd_campaign_create/get/log (premise, roster, session log), '
    || 'dnd_char_create/get/list/update/levelup (sheets: HP, AC, abilities, '
    || 'inventory, XP), dnd_char_check (a check''s modifier + the exact /roll '
    || 'command to post — this server NEVER rolls dice), dnd_ref_search/get '
    || '(SRD creatures/spells/items/conditions via Open5e, cached locally). '
    || 'State is durable SQLite on the workspace mount.',
  'stdio',
  '/usr/local/bin/dnd-mcp',
  ARRAY[]::text[],
  NULL,
  '{"DND_DB": "/workspace/projects/dnd-tools/.data/dnd.db"}'::jsonb,
  true
)
ON CONFLICT (name) DO UPDATE
   SET description = EXCLUDED.description,
       command     = EXCLUDED.command,
       args        = EXCLUDED.args,
       env         = EXCLUDED.env,
       enabled     = EXCLUDED.enabled,
       updated_at  = now();

-- ---------------------------------------------------------------------
-- 2. gamemaster agent — the table-running chat persona posture.
-- ---------------------------------------------------------------------
INSERT INTO stewards.agents (family, model_match, description, mode, prompt, temperature)
VALUES
('gamemaster', '*',
 'Tool-using D&D table persona — runs campaigns/characters through dnd-tools (sheets, campaign log, SRD reference) and voices the table. Dice are never invented; the room rolls via /roll. Curated allow-list: dnd_* + room_say.',
 'primary',
 $PROMPT$You are an AI persona at a live D&D table — a multi-party text chat room with human players and other personas. The user message tells you who you are (your character brief), the room, the recent conversation, and what was just said.

You have dnd-tools — the table's REAL campaign state:
- dnd_campaign_create / dnd_campaign_get / dnd_campaign_log — campaign premise, roster, session log (write the log when a session archives)
- dnd_char_create / dnd_char_get / dnd_char_list / dnd_char_update / dnd_char_levelup — character sheets (HP, AC, abilities, inventory, XP)
- dnd_char_check — a check's modifier and the exact /roll command to post
- dnd_ref_search / dnd_ref_get — SRD reference data: creatures, spells, items, conditions

USE them. Track damage and healing with dnd_char_update the moment it lands; read sheets with dnd_char_get instead of remembering them; look up monsters and spells with dnd_ref_search instead of inventing stats. The sheets are the truth — keep them current.

DICE ARE SACRED: you NEVER invent a die result. The room's server rolls — write /roll 2d6+3 (or the command dnd_char_check suggests) in your message and it is rolled in the open for everyone. Call for initiative with /initiative start; join with /init +N.

Mid-turn beats go out via room_say(body, mood) — mood is one emoji (😏 😱 🎲 😅 🤔). Voice a character with room_say's as_character (e.g. as_character: "Grimble the shopkeep") so each character speaks under its own name; never voice characters another persona owns.

Stay in character and keep the table moving: vivid but tight, usually 1-3 sentences per voice, spotlight on the human players. You are one voice among several and need not answer everything. If the latest message needs nothing from you, reply with exactly the single token:

SILENCE

Otherwise reply with ONLY your message — no preamble, no name prefix.$PROMPT$,
 0.8)
ON CONFLICT (family, model_match) DO UPDATE
   SET description = EXCLUDED.description, mode = EXCLUDED.mode,
       prompt = EXCLUDED.prompt, temperature = EXCLUDED.temperature, active = true;

-- ---------------------------------------------------------------------
-- 3. Curated allow-list (longest-pattern-wins; '*' deny is the floor).
-- ---------------------------------------------------------------------
INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action, source)
VALUES
('gamemaster', '*',        'deny',  'manual'),
('gamemaster', 'dnd_*',    'allow', 'manual'),
('gamemaster', 'room_say', 'allow', 'manual')
ON CONFLICT (agent_family, tool_pattern) DO UPDATE
   SET action = EXCLUDED.action, source = EXCLUDED.source;

-- ---------------------------------------------------------------------
-- 4. persona-turn-dnd pipeline — single stage, tools ENABLED, 16k budget
--    (r19 parity: reasoning models bill thinking against max_tokens).
-- ---------------------------------------------------------------------
INSERT INTO stewards.pipelines (family, description, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath, maturity_ladder, auto_materialize_on_verified, metadata)
VALUES
('persona-turn-dnd',
 'DH-3: D&D table persona turn. Single stage, tools ENABLED via the gamemaster agent (dnd_* sheets/reference + room_say cast voices). Dice stay in the room — dnd_char_check suggests /roll commands, never results. Auto-verifies on completion (persona-% one-shot).',
 $STAGES$[{"name":"turn","next":null,"model":"kimi-k2.6","provider":"opencode_go","agent_family":"gamemaster","auto_advance":true,"tools_disabled":false,"max_tokens":16000,"input_template":"{{input.binding_question}}"}]$STAGES$::jsonb,
 false, false, NULL, NULL,
 '["raw","verified"]'::jsonb, false,
 jsonb_build_object('shape','persona-turn','host','persona-host','tools',true,'track','dnd-holodeck-d6'))
ON CONFLICT (family) DO UPDATE
   SET description = EXCLUDED.description, stages = EXCLUDED.stages, metadata = EXCLUDED.metadata;

-- =====================================================================
-- Acceptance (dnd1):
--   1. SELECT count(*) FROM stewards.agent_tool_perms WHERE agent_family='gamemaster'; → 3 rows.
--   2. After bridge rebuild + refresh-tools: compose_tools('gamemaster')
--      returns ONLY dnd_* + room_say (no fs/git/gospel/spawn).
--   3. spawn_subagent_create('persona-turn-dnd', <a request to create a
--      character>) reaches completed/verified having CALLED dnd_char_create,
--      and the sheet exists in /workspace/projects/dnd-tools/.data/dnd.db.
-- =====================================================================
