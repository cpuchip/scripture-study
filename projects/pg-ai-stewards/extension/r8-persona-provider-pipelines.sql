-- =====================================================================
-- Batch R.8 — persona pipelines on alternate model providers (AXR6 examples)
-- =====================================================================
-- ai-chattermax personas default to persona-turn (kimi-k2.6 / opencode_go).
-- For the docs/examples that show a coworker how to back a persona with their
-- own model, add provider variants: LM Studio (local) + Google Gemini. Same
-- thin `persona` agent + tools-disabled + single-stage shape as persona-turn;
-- only the model+provider differ. persona-host selects a persona's pipeline via
-- the new persona_host.personas.pipeline column.
--
-- Also: generalize the R.6/R.7 one-shot auto-verify to LIKE 'persona-%' so every
-- persona* pipeline (incl. these + a future persona-tool-turn) auto-verifies on
-- completion (else the host's spawn poll hangs).
-- =====================================================================

-- LM Studio (local, OpenAI-compatible; reached via host.docker.internal:1234).
INSERT INTO stewards.pipelines (family, description, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath, maturity_ladder, auto_materialize_on_verified, metadata)
VALUES
('persona-turn-lmstudio',
 'R.8: persona turn on a local LM Studio model (qwen3.6-27b). Same as persona-turn, different provider — an example backend for a self-hosted persona.',
 $STAGES$[{"name":"turn","next":null,"model":"qwen/qwen3.6-27b","provider":"lm_studio","agent_family":"persona","auto_advance":true,"tools_disabled":true,"max_tokens":1200,"input_template":"{{input.binding_question}}"}]$STAGES$::jsonb,
 false, false, NULL, NULL,
 '["raw","verified"]'::jsonb, false,
 jsonb_build_object('shape','persona-turn','host','persona-host','provider','lm_studio'))
ON CONFLICT (family) DO UPDATE SET description=EXCLUDED.description, stages=EXCLUDED.stages, metadata=EXCLUDED.metadata;

-- Google Gemini (gemini-3.5-flash).
INSERT INTO stewards.pipelines (family, description, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath, maturity_ladder, auto_materialize_on_verified, metadata)
VALUES
('persona-turn-gemini',
 'R.8: persona turn on Google Gemini (gemini-3.5-flash). Same as persona-turn, different provider — an example backend for a persona on a hosted API.',
 $STAGES$[{"name":"turn","next":null,"model":"gemini-3.5-flash","provider":"google_gemini","agent_family":"persona","auto_advance":true,"tools_disabled":true,"max_tokens":1200,"input_template":"{{input.binding_question}}"}]$STAGES$::jsonb,
 false, false, NULL, NULL,
 '["raw","verified"]'::jsonb, false,
 jsonb_build_object('shape','persona-turn','host','persona-host','provider','google_gemini'))
ON CONFLICT (family) DO UPDATE SET description=EXCLUDED.description, stages=EXCLUDED.stages, metadata=EXCLUDED.metadata;

-- The `persona` agent must send NO tools: it was allow-by-default minus r7's
-- specific denies, so compose_tools('persona') returned 93 tools every turn.
-- kimi tolerated the (malformed) schemas; LM Studio's stricter validation 400'd.
-- A deny-* makes a CHARACTER persona tool-free (tool-USING personas get their own
-- agent with explicit allows). Also trims every persona turn's payload.
INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action)
VALUES ('persona', '*', 'deny')
ON CONFLICT (agent_family, tool_pattern) DO UPDATE SET action = 'deny';

-- Generalize one-shot auto-verify to all persona-* pipelines (carry R.6/R.7 fwd).
CREATE OR REPLACE FUNCTION stewards.on_one_shot_pipeline_completed()
RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    v_qualifies boolean;
BEGIN
    v_qualifies := NEW.pipeline_family = 'aggregate-children'
                OR NEW.pipeline_family LIKE 'brainstorm-%'
                OR NEW.pipeline_family LIKE 'redline%'
                OR NEW.pipeline_family LIKE 'persona-%';   -- R.8: any persona-* pipeline
    IF NOT v_qualifies THEN
        RETURN NEW;
    END IF;
    IF NEW.maturity = 'verified' THEN
        RETURN NEW;
    END IF;
    UPDATE stewards.work_items SET maturity = 'verified', updated_at = now() WHERE id = NEW.id;
    RAISE NOTICE 'on_one_shot_pipeline_completed: auto-verified % (pipeline=%)', NEW.id, NEW.pipeline_family;
    RETURN NEW;
END;
$$;

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
        OR NEW.pipeline_family LIKE 'persona-%'
    )
)
EXECUTE FUNCTION stewards.on_one_shot_pipeline_completed();

-- =====================================================================
-- Acceptance: spawn_subagent_create('persona-turn-lmstudio', <framing>) reaches
-- completed/verified with an in-character reply (proves LM Studio chat dispatch).
-- =====================================================================
