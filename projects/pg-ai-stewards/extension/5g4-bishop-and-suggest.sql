-- =====================================================================
-- Phase 5g.4 (Phase F.5) — bishop_eligible + watchman convening hints
--
-- Two pieces:
--   1. stewards.bishop_eligible(bishop_token, intent_id) — checks
--      whether a bishop can preside per D-F2 (humans always; agents
--      only if master-tier on the intent's pipeline AND the intent is
--      "low-stakes" per the 2026-05-11 ratification: scripture_anchor
--      IS NULL AND values_hierarchy doesn't contain doctrinal /
--      spiritual / discernment).
--   2. stewards.suggest_councils() — scans stewards.lessons for
--      clusters of 5+ ratified lessons on the same (pipeline_family,
--      current_stage); returns suggestions watchman pass surfaces.
--      Per D-F4: surfaces in BOTH watchman pass output AND dashboard
--      banner (UI surfacing in F.7).
-- =====================================================================

-- ---------------------------------------------------------------------
-- (1) bishop_eligible
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.bishop_eligible(
    p_bishop    text,
    p_intent_id uuid
) RETURNS boolean
LANGUAGE plpgsql STABLE AS $func$
DECLARE
    v_intent      stewards.intents%ROWTYPE;
    v_parts       text[];
    v_agent       text;
    v_pipeline    text;
    v_required_tier text;
    v_actual_level text;
    v_low_stakes  boolean;
BEGIN
    -- Humans always eligible
    IF p_bishop LIKE 'human:%' THEN
        RETURN true;
    END IF;

    SELECT * INTO v_intent FROM stewards.intents WHERE id = p_intent_id;
    IF v_intent.id IS NULL THEN
        RETURN false;
    END IF;

    -- Low-stakes check (per 2026-05-11 ratification)
    -- Doctrinal/spiritual/discernment intents always require human bishop.
    v_low_stakes := (
        v_intent.scripture_anchor IS NULL
        AND v_intent.values_hierarchy::text !~* '(doctrinal|spiritual|discernment)'
    );

    IF NOT v_low_stakes THEN
        RETURN false;
    END IF;

    -- Parse 'agent:<family>:<pipeline>:master' (or :journeyman, etc.)
    v_parts := string_to_array(p_bishop, ':');
    IF array_length(v_parts, 1) < 4 OR v_parts[1] <> 'agent' THEN
        RETURN false;
    END IF;
    v_agent         := v_parts[2];
    v_pipeline      := v_parts[3];
    v_required_tier := v_parts[4];

    IF v_required_tier <> 'master' THEN
        RETURN false;  -- F1 requires master per D-F2
    END IF;

    -- Look up actual trust level for any model on this (agent, pipeline)
    -- — F1 keying is master-on-pipeline regardless of model. If ANY
    -- (agent, pipeline, model) cell is master, the agent is eligible.
    SELECT trust_level INTO v_actual_level
      FROM stewards.trust_scores
     WHERE agent_family = v_agent
       AND pipeline_family = v_pipeline
       AND trust_level = 'master'
     LIMIT 1;

    RETURN v_actual_level IS NOT NULL;
END;
$func$;

COMMENT ON FUNCTION stewards.bishop_eligible(text, uuid) IS
'Phase 5g (F.5): bishop eligibility per D-F2. Humans (bishop=human:<name>) always eligible. Agents (bishop=agent:<family>:<pipeline>:master) only on low-stakes intents (no scripture_anchor + values_hierarchy lacks doctrinal/spiritual/discernment) AND must be master-tier on at least one (agent, pipeline, model) cell. F2 future evolution path: introduce council_authority as separate trust dimension; debug agent as candidate first cultivator (per Michael''s 2026-05-11 nuance).';

-- ---------------------------------------------------------------------
-- (2) suggest_councils
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.suggest_councils(
    p_min_lessons int DEFAULT 5
) RETURNS TABLE (
    pipeline_family text,
    current_stage   text,
    lesson_count    bigint,
    sample_content  text
)
LANGUAGE sql STABLE AS $func$
SELECT
    pipeline_family,
    current_stage,
    count(*) AS lesson_count,
    string_agg('  - ' || left(content, 100), E'\n' ORDER BY at DESC) FILTER (WHERE rn <= 3) AS sample_content
  FROM (
    SELECT
        l.id,
        l.content,
        l.at,
        wi.pipeline_family,
        wi.current_stage,
        row_number() OVER (PARTITION BY wi.pipeline_family, wi.current_stage ORDER BY l.at DESC) AS rn
      FROM stewards.lessons l
      JOIN stewards.work_items wi ON wi.id = l.work_item_id
     WHERE l.ratified_at IS NOT NULL
       AND l.kind IN ('lesson', 'principle')
       -- Avoid re-suggesting clusters already addressed by an open or
       -- recently-resolved council. Heuristic: the lesson must be
       -- newer than the most recent council on this pipeline.
       AND l.at > COALESCE((
           SELECT max(c.convened_at)
             FROM stewards.councils c
             JOIN stewards.intents i ON i.id = c.intent_id
            WHERE i.purpose ILIKE '%' || wi.pipeline_family || '%'
       ), '-infinity'::timestamptz)
  ) clustered
 GROUP BY pipeline_family, current_stage
HAVING count(*) >= p_min_lessons
 ORDER BY lesson_count DESC, pipeline_family, current_stage;
$func$;

COMMENT ON FUNCTION stewards.suggest_councils(int) IS
'Phase 5g (F.5): scan ratified lessons for clusters by (pipeline_family, current_stage). Default threshold 5+. Returns rows watchman pass + dashboard banner surface. Heuristic dedupe: skip clusters where a council on this pipeline has already been convened more recently than the lessons.';
