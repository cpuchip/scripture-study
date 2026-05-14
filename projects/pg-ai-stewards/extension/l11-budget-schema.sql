-- =====================================================================
-- Batch L.1.1.1 — Budget schema + effective_budget cascade
-- =====================================================================
-- Adds working_budget to agents (column) and to pipelines.stages[]
-- (jsonb field — no schema change). Helper effective_budget walks the
-- ratified cascade: pipeline-stage > agent > provider.context_window.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. working_budget column on agents.
-- ---------------------------------------------------------------------

ALTER TABLE stewards.agents
    ADD COLUMN IF NOT EXISTS working_budget integer;

COMMENT ON COLUMN stewards.agents.working_budget IS
'Batch L.1.1: the agent''s declared working context budget in tokens. NULL means inherit from provider.context_window. Pipeline stage working_budget takes precedence over this when set.';


-- ---------------------------------------------------------------------
-- 2. Helper: extract working_budget from a stage definition.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.stage_working_budget(
    p_pipeline_family text,
    p_stage_name text
) RETURNS integer LANGUAGE plpgsql STABLE AS $FN$
DECLARE
    v_stage jsonb;
    v_budget int;
BEGIN
    IF p_pipeline_family IS NULL OR p_stage_name IS NULL THEN
        RETURN NULL;
    END IF;

    SELECT s INTO v_stage
      FROM stewards.pipelines p,
           LATERAL jsonb_array_elements(p.stages) s
     WHERE p.family = p_pipeline_family
       AND (s ->> 'name') = p_stage_name
     LIMIT 1;

    IF v_stage IS NULL THEN
        RETURN NULL;
    END IF;

    v_budget := (v_stage ->> 'working_budget')::int;
    RETURN v_budget;
EXCEPTION WHEN invalid_text_representation THEN
    RETURN NULL;
END;
$FN$;

COMMENT ON FUNCTION stewards.stage_working_budget(text, text) IS
'Batch L.1.1: read the working_budget field from a specific stage in a pipeline.stages[] array. Returns NULL if not declared.';


-- ---------------------------------------------------------------------
-- 3. Helper: resolve effective_budget(session_id, stage_name)
--    walking the ratified cascade.
-- ---------------------------------------------------------------------
-- pipeline-stage > agent > provider.context_window
--
-- The "agent" here is determined by the active session's most-recent
-- payload's agent_family (since one session can run multiple stages
-- across its lifetime). For lookup, we prefer the most recent chat
-- work_queue row for the session.

CREATE OR REPLACE FUNCTION stewards.effective_budget(
    p_session_id text,
    p_stage_name text DEFAULT NULL
) RETURNS integer LANGUAGE plpgsql STABLE AS $FN$
DECLARE
    v_work_item    stewards.work_items%ROWTYPE;
    v_stage_name   text := p_stage_name;
    v_agent_family text;
    v_budget       int;
    v_provider     text;
    v_context_win  int;
BEGIN
    -- Locate the parent work_item via session_ids.
    SELECT * INTO v_work_item
      FROM stewards.work_items
     WHERE p_session_id = ANY(session_ids)
     LIMIT 1;

    -- Stage defaults to the work_item's current_stage.
    IF v_stage_name IS NULL THEN
        v_stage_name := v_work_item.current_stage;
    END IF;

    -- Layer 1: pipeline-stage.
    v_budget := stewards.stage_working_budget(v_work_item.pipeline_family, v_stage_name);
    IF v_budget IS NOT NULL AND v_budget > 0 THEN
        RETURN v_budget;
    END IF;

    -- Layer 2: agent — resolve from the most-recent chat payload on this session.
    SELECT payload ->> 'agent_family' INTO v_agent_family
      FROM stewards.work_queue
     WHERE payload ->> 'session_id' = p_session_id
       AND kind = 'chat'
     ORDER BY id DESC
     LIMIT 1;

    IF v_agent_family IS NOT NULL THEN
        SELECT working_budget INTO v_budget
          FROM stewards.agents
         WHERE family = v_agent_family
           AND active
         ORDER BY model_match = '*' ASC  -- prefer specific match
         LIMIT 1;
        IF v_budget IS NOT NULL AND v_budget > 0 THEN
            RETURN v_budget;
        END IF;
    END IF;

    -- Layer 3: provider.context_window via provider_for_session (L.1).
    v_provider := stewards.provider_for_session(p_session_id);
    IF v_provider IS NOT NULL THEN
        SELECT context_window INTO v_context_win
          FROM stewards.provider_rules
         WHERE name = v_provider;
        IF v_context_win IS NOT NULL AND v_context_win > 0 THEN
            RETURN v_context_win;
        END IF;
    END IF;

    -- Final fallback: a conservative default so callers never get NULL.
    RETURN 64000;
END;
$FN$;

COMMENT ON FUNCTION stewards.effective_budget(text, text) IS
'Batch L.1.1: resolve the effective working budget (tokens) for a session+stage. Cascade: pipeline-stage.working_budget > agent.working_budget > provider.context_window. Final fallback 64000. Used by L.1.1.2 (extraction threshold), L.1.1.8 (intercept threshold), and any caller that needs to size budgets against the consuming agent''s actual capacity.';


-- =====================================================================
-- End of l11-budget-schema.sql
-- =====================================================================
