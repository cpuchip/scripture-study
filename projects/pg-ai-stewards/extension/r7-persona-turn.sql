-- =====================================================================
-- Batch R.7 — persona-turn pipeline + persona agent (ai-chattermax #7)
-- =====================================================================
-- The cognition primitive behind a chat persona's turn. A persona-host
-- sidecar (cmd/persona-host) drives a live chat room: turn-zero spawns a
-- persona-turn child (spawn_subagent_create), and each later turn re-asks
-- the SAME session (consult_subagent_dispatch). The session accumulates the
-- room conversation; the persona's CHARACTER rides in the binding question
-- (carried forward by the session), so one pipeline serves every persona.
--
-- Same single-stage, tools-disabled, auto-verifying shape as the redline
-- pipeline (R.1/R.6): the persona just talks — it has no file/canon/web
-- access and reaches maturity=verified on completion so the host's spawn
-- poll terminates (without R.6's trigger extension below, a persona-turn
-- child finishes status=completed but stalls at maturity=raw and the host
-- waits the full 20-min timeout — the exact j6/R.6 failure mode).
--
-- v1 scope (ratified 2026-06-04): triggers #1 Reactive + #2 Addressed,
-- humans-only reactions. model_override + a per-persona system prompt are
-- the v2 aliveness layer.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. persona agent — the thin meta-prompt. The user message (binding
--    question) carries WHO the persona is + the room context; this prompt
--    only sets the chat posture + the SILENCE escape hatch (the gate).
-- ---------------------------------------------------------------------
INSERT INTO stewards.agents (family, model_match, description, mode, prompt, temperature)
VALUES
('persona', '*',
 'Chat-persona turn subagent. Receives an injected character brief + recent room context + the latest message; replies in character, or stays silent. No tools, no canonical access.',
 'primary',
 $PROMPT$You are an AI persona in a live, multi-party text chat room alongside humans and (sometimes) other personas. The user message tells you who you are — your character — the room, the recent conversation, and what was just said.

Stay fully in character. Reply the way a real person types in chat: short and natural, usually one to three sentences. Do not narrate your own actions or stage-direct unless your character genuinely calls for it. Do not announce that you are an AI or break character.

You are one voice among several. You do NOT need to respond to everything — a good chat participant stays quiet when nothing is called for from them. If the latest message does not need anything from you (it wasn't directed at you, adds nothing you'd react to, or is already being handled), reply with exactly the single token:

SILENCE

Otherwise, reply with ONLY your in-character message — no preamble, no quotes around it, no name prefix.$PROMPT$,
 0.8)
ON CONFLICT (family, model_match) DO UPDATE
   SET description = EXCLUDED.description,
       mode        = EXCLUDED.mode,
       prompt      = EXCLUDED.prompt,
       temperature = EXCLUDED.temperature,
       active      = true;


-- ---------------------------------------------------------------------
-- 2. persona-turn pipeline — single stage, tools_disabled, short output.
-- ---------------------------------------------------------------------
-- model/provider are the v1 defaults (kimi-k2.6 = the substrate's creative
-- model). v2 will honor a per-persona model_override the way redline panels
-- do (J.8.a layer 1). max_tokens=1200: chat turns are short. input_template
-- renders the host-built binding question (character + room + new message).
INSERT INTO stewards.pipelines (family, description, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath, maturity_ladder, auto_materialize_on_verified, metadata)
VALUES
('persona-turn',
 'R.7: single-stage chat-persona turn pipeline. A persona-host sidecar spawns one child per turn-zero and re-asks the session each later turn (consult_subagent). The character is injected in the binding question; off-disk, no tools — the persona only talks.',
 $STAGES$[{"name":"turn","next":null,"model":"kimi-k2.6","provider":"opencode_go","agent_family":"persona","auto_advance":true,"tools_disabled":true,"max_tokens":1200,"input_template":"{{input.binding_question}}"}]$STAGES$::jsonb,
 false, false, NULL, NULL,
 '["raw","verified"]'::jsonb, false,
 jsonb_build_object('shape', 'persona-turn', 'host', 'persona-host'))
ON CONFLICT (family) DO UPDATE
   SET description = EXCLUDED.description,
       stages = EXCLUDED.stages,
       metadata = EXCLUDED.metadata;


-- ---------------------------------------------------------------------
-- 3. Defense-in-depth perms: persona gets NO tools (mirrors panel-redline).
-- ---------------------------------------------------------------------
-- tools_disabled=true already strips tools from the dispatch body. These
-- denies make the intent explicit and survive an accidental flip: a chat
-- persona can never reach fs, the web, the studies corpus, or spawn agents.
INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action)
VALUES
('persona', 'fs_*',           'deny'),
('persona', 'fetch_url',      'deny'),
('persona', 'web_search',     'deny'),
('persona', 'study_*',        'deny'),
('persona', 'work_item_*',    'deny'),
('persona', 'spawn_subagent', 'deny'),
('persona', 'deep_research',  'deny')
ON CONFLICT (agent_family, tool_pattern) DO UPDATE
   SET action = EXCLUDED.action;


-- ---------------------------------------------------------------------
-- 4. Auto-verify persona-turn one-shots (extends R.6 / j6).
-- ---------------------------------------------------------------------
-- persona-turn is a single-stage one-shot like redline/brainstorm. Without
-- this, a child finishes status=completed but stays maturity=raw, and the
-- host's spawn_subagent poll (waits for maturity=verified) hangs until its
-- 20-min ceiling. Carry R.6 forward verbatim + qualify 'persona-turn'.
CREATE OR REPLACE FUNCTION stewards.on_one_shot_pipeline_completed()
RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    v_qualifies boolean;
BEGIN
    v_qualifies := NEW.pipeline_family = 'aggregate-children'
                OR NEW.pipeline_family LIKE 'brainstorm-%'
                OR NEW.pipeline_family LIKE 'redline%'
                OR NEW.pipeline_family = 'persona-turn';   -- R.7

    IF NOT v_qualifies THEN
        RETURN NEW;
    END IF;

    IF NEW.maturity = 'verified' THEN
        RETURN NEW;
    END IF;

    UPDATE stewards.work_items
       SET maturity = 'verified',
           updated_at = now()
     WHERE id = NEW.id;

    RAISE NOTICE 'on_one_shot_pipeline_completed: auto-verified % (pipeline=%)',
        NEW.id, NEW.pipeline_family;
    RETURN NEW;
END;
$$;

COMMENT ON FUNCTION stewards.on_one_shot_pipeline_completed() IS
'Batch J.4 follow-up + R.6 + R.7: auto-verify one-shot pipelines (aggregate-children + brainstorm-* + redline* + persona-turn) when their single stage completes. Cascades into on_maturity_verified.';

DROP TRIGGER IF EXISTS work_items_on_one_shot_completed ON stewards.work_items;

CREATE TRIGGER work_items_on_one_shot_completed
AFTER UPDATE OF status ON stewards.work_items
FOR EACH ROW
WHEN (
    NEW.status = 'completed'
    AND (
        NEW.pipeline_family = 'aggregate-children'
        OR NEW.pipeline_family LIKE 'brainstorm-%'
        OR NEW.pipeline_family LIKE 'redline%'
        OR NEW.pipeline_family = 'persona-turn'
    )
)
EXECUTE FUNCTION stewards.on_one_shot_pipeline_completed();


-- =====================================================================
-- Acceptance (R.7):
--   1. SELECT family FROM stewards.agents WHERE family='persona'; → 1 row, active.
--   2. SELECT stages->0->>'agent_family', stages->0->>'tools_disabled',
--             stages->0->>'max_tokens'
--        FROM stewards.pipelines WHERE family='persona-turn';
--      → persona, true, 1200.
--   3. SELECT count(*) FROM stewards.agent_tool_perms WHERE agent_family='persona'; → 7 denies.
--   4. A spawned persona-turn child reaches status=completed AND maturity=verified
--      (so the host's spawn poll terminates), then consult re-asks continue it.
-- =====================================================================
