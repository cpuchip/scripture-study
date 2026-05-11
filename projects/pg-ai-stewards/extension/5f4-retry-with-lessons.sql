-- =====================================================================
-- Phase 5f.4 (Phase E.4) — retry composer pulls last 3 ratified lessons
--
-- Extends Phase A's retry_guidance(diagnosis, attempt) with a new
-- variant that also appends "Recent lessons from this pipeline + stage"
-- pulling from stewards.lessons_recent_ratified (Phase D view).
--
-- Per ratification:
--   - Last 3 ratified lessons for (pipeline, stage)
--   - Only ratified lessons (D-D3 said human curates)
--   - Stage-specific avoids polluting outline retries with draft lessons
-- =====================================================================

CREATE OR REPLACE FUNCTION stewards.retry_guidance_with_lessons(
    p_diagnosis       text,
    p_attempt         integer,
    p_pipeline_family text,
    p_stage_name      text
) RETURNS text
LANGUAGE plpgsql STABLE AS $func$
DECLARE
    v_base    text;
    v_lessons text;
BEGIN
    v_base := stewards.retry_guidance(p_diagnosis, p_attempt);

    SELECT string_agg('  - ' || content, E'\n')
      INTO v_lessons
      FROM (
        SELECT content
          FROM stewards.lessons_recent_ratified
         WHERE pipeline_family = p_pipeline_family
           AND current_stage   = p_stage_name
         ORDER BY at DESC
         LIMIT 3
      ) recent;

    IF v_lessons IS NOT NULL THEN
        v_base := coalesce(v_base, '') ||
                  E'\n\nRecent lessons from this pipeline + stage:\n' ||
                  v_lessons;
    END IF;

    RETURN v_base;
END;
$func$;

COMMENT ON FUNCTION stewards.retry_guidance_with_lessons(text, integer, text, text) IS
'Phase 5f (E.4): wraps retry_guidance() and appends last 3 ratified lessons for the (pipeline_family, current_stage) cell from lessons_recent_ratified view. Only ratified content (kind in lesson|principle) influences retry context — proposed-but-unratified lessons stay out per D-D3.';
