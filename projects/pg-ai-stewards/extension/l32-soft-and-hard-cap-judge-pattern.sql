-- =====================================================================
-- L.1.1.17 — Soft cap (Judge pattern) + hard cap safety net
-- =====================================================================
-- Replaces L.1.1.16's single hard cap with a two-tier model:
--
--   Soft (max_tool_rounds, default 5):
--     At threshold, INSERT a role='system' STEWARD NOTICE message into
--     the session before composing the next chat. Tools stay available;
--     the agent decides whether to continue. Notice content suggests
--     including a one-sentence justification in the next response so a
--     future audit can review the decision.
--
--   Hard (max_tool_rounds_hard, default 50):
--     Cost safety net. At threshold, force tools_disabled=true +
--     tool_choice="none" (the L.1.1.16 mechanism). Real ceiling.
--
-- Per Michael (2026-05-14): "I really want to see what the system can
-- produce at this early stage, and only limit it to protect costs."
-- The OpenCode Go subscription cap is the wallet-level safety net.
--
-- Honors the Judges principle (Exodus 18:21-22 + go-and-do return-and-
-- report 1 Nephi 3:7 → 4:6): substrate funds the going, doesn't
-- interrupt thoroughness mid-mission.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. Helper: stage_max_tool_rounds_hard.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.stage_max_tool_rounds_hard(
    p_pipeline_family text,
    p_stage_name      text
) RETURNS int LANGUAGE plpgsql STABLE AS $FN$
DECLARE
    v_stage  jsonb;
    v_rounds int;
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

    BEGIN
        v_rounds := (v_stage ->> 'max_tool_rounds_hard')::int;
    EXCEPTION WHEN invalid_text_representation THEN
        v_rounds := NULL;
    END;
    RETURN v_rounds;
END;
$FN$;

COMMENT ON FUNCTION stewards.stage_max_tool_rounds_hard(text, text) IS
'Batch L.1.1.17: read the hard cap from a pipeline stage. Hard cap is the safety-net ceiling — at-or-above this round count, tools_disabled+tool_choice=none are forced.';


-- ---------------------------------------------------------------------
-- 2. Helper: build the soft-cap STEWARD NOTICE text.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.build_soft_cap_notice(
    p_rounds_so_far int,
    p_soft_cap      int,
    p_hard_cap      int,
    p_stage_name    text
) RETURNS text LANGUAGE sql IMMUTABLE AS $$
    SELECT
        '[STEWARD NOTICE — soft cap reached]' || E'\n\n' ||
        'You have used ' || p_rounds_so_far::text || ' tool calls in the ' ||
        COALESCE(p_stage_name, 'current') || ' stage. The soft cap for this stage is ' ||
        p_soft_cap::text || '; the hard cap (where tools will be removed entirely) is ' ||
        p_hard_cap::text || '.' || E'\n\n' ||
        'If you can answer the binding question now from what you have already gathered, ' ||
        'finalize your response. If you genuinely need another tool call, include a ' ||
        'one-sentence justification in your next response so future review can audit ' ||
        'the decision.' || E'\n\n' ||
        'You retain full agency. The substrate is funding your mission, not micromanaging it.'
$$;


-- ---------------------------------------------------------------------
-- 3. Rewrite chat_post_internal — two-tier cap.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.chat_post_internal(
    p_agent_family text,
    p_model        text,
    p_session_id   text,
    p_provider     text
) RETURNS bigint LANGUAGE plpgsql AS $FN$
DECLARE
    v_body                  jsonb;
    v_payload               jsonb;
    v_work_id               bigint;
    v_inherited_markers     jsonb;
    v_stage_name            text;
    v_pipeline_family       text;
    v_soft_cap              int;
    v_hard_cap              int;
    v_rounds_so_far         int;
    v_force_tools_disabled  boolean := false;
    v_inject_soft_notice    boolean := false;
    v_already_soft_notified boolean := false;
    v_notice_text           text;
BEGIN
    -- Pull inherited markers FIRST so we can use them for cap lookup
    -- BEFORE composing the body.
    SELECT jsonb_object_agg(je.key, je.value)
      INTO v_inherited_markers
      FROM stewards.work_queue wq
      CROSS JOIN LATERAL jsonb_each(wq.payload) je
     WHERE wq.payload->>'session_id' = p_session_id
       AND wq.kind = 'chat'
       AND wq.id = (
           SELECT max(id) FROM stewards.work_queue
            WHERE payload->>'session_id' = p_session_id
              AND kind = 'chat'
       )
       AND je.key LIKE '\_%' ESCAPE '\';

    v_pipeline_family := v_inherited_markers ->> '_pipeline_family';
    v_stage_name      := v_inherited_markers ->> '_stage_name';

    -- Has soft notice already been injected this stage? Avoid spam.
    v_already_soft_notified := COALESCE(
        (v_inherited_markers ->> '_soft_cap_notified')::boolean, false);

    IF v_pipeline_family IS NOT NULL AND v_stage_name IS NOT NULL THEN
        v_soft_cap := COALESCE(
            stewards.stage_max_tool_rounds(v_pipeline_family, v_stage_name),
            5
        );
        v_hard_cap := COALESCE(
            stewards.stage_max_tool_rounds_hard(v_pipeline_family, v_stage_name),
            50
        );

        SELECT count(*) INTO v_rounds_so_far
          FROM stewards.messages
         WHERE session_id = p_session_id
           AND role = 'assistant';

        -- Hard cap takes precedence.
        IF v_rounds_so_far >= v_hard_cap THEN
            v_force_tools_disabled := true;
            RAISE NOTICE 'chat_post_internal: session=% rounds=%/HARD-cap-% — forcing tools_disabled+tool_choice=none',
                p_session_id, v_rounds_so_far, v_hard_cap;
        ELSIF v_rounds_so_far >= v_soft_cap AND NOT v_already_soft_notified THEN
            v_inject_soft_notice := true;
            RAISE NOTICE 'chat_post_internal: session=% rounds=%/soft-cap-% — injecting STEWARD NOTICE',
                p_session_id, v_rounds_so_far, v_soft_cap;
        END IF;
    END IF;

    -- L.1.1.17: inject soft notice as role='system' message BEFORE composing.
    IF v_inject_soft_notice THEN
        v_notice_text := stewards.build_soft_cap_notice(
            v_rounds_so_far, v_soft_cap, v_hard_cap, v_stage_name);
        INSERT INTO stewards.messages (session_id, role, content, model)
        VALUES (p_session_id, 'system', v_notice_text, p_model);
    END IF;

    -- Now compose the body (will pick up the just-inserted system notice).
    v_body := stewards.dry_run_chat(p_agent_family, p_model, p_session_id, NULL, p_provider);
    v_body := v_body - '_meta';

    IF v_force_tools_disabled THEN
        v_body := v_body || jsonb_build_object('tool_choice', 'none');
    END IF;

    v_payload := jsonb_build_object(
        'session_id',      p_session_id,
        'agent_family',    p_agent_family,
        'requested_model', p_model,
        'body',            v_body
    );

    IF v_force_tools_disabled THEN
        v_payload := v_payload || jsonb_build_object('tools_disabled', true);
    END IF;

    -- Mark soft-notice injection so we don't spam every continuation.
    IF v_inject_soft_notice THEN
        v_payload := v_payload || jsonb_build_object(
            '_soft_cap_notified', true,
            '_soft_cap_injected_at_round', v_rounds_so_far
        );
    END IF;

    IF v_inherited_markers IS NOT NULL THEN
        -- Don't overwrite the just-set soft markers.
        v_payload := (v_inherited_markers - '_soft_cap_notified' - '_soft_cap_injected_at_round') || v_payload;
    END IF;

    INSERT INTO stewards.work_queue (kind, provider, payload, status)
    VALUES ('chat', p_provider, v_payload, 'pending')
    RETURNING id INTO v_work_id;

    RETURN v_work_id;
END;
$FN$;


-- ---------------------------------------------------------------------
-- 4. Update research-write defaults: soft + hard caps per stage.
-- ---------------------------------------------------------------------
-- Soft (existing): context_gather=5, gather=5, synthesize=3, review=1
-- Hard (NEW):       context_gather=50, gather=50, synthesize=15, review=3

UPDATE stewards.pipelines
   SET stages = jsonb_set(stages, '{0,max_tool_rounds_hard}', '50'::jsonb)
 WHERE family = 'research-write';

UPDATE stewards.pipelines
   SET stages = jsonb_set(stages, '{1,max_tool_rounds_hard}', '50'::jsonb)
 WHERE family = 'research-write';

UPDATE stewards.pipelines
   SET stages = jsonb_set(stages, '{2,max_tool_rounds_hard}', '15'::jsonb)
 WHERE family = 'research-write';

UPDATE stewards.pipelines
   SET stages = jsonb_set(stages, '{3,max_tool_rounds_hard}', '3'::jsonb)
 WHERE family = 'research-write';


-- =====================================================================
-- End of l32-soft-and-hard-cap-judge-pattern.sql
-- =====================================================================
