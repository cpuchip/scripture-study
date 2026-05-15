-- =====================================================================
-- L.1.1.15 — Model resolver substitution log
-- =====================================================================
-- Bacteriopolis silently ran on qwen3.6-plus for 75 chats when the
-- pipeline (research-write.gather/synthesize) said kimi-k2.6. The
-- substitution can come from steward retry logic, model_override,
-- escalation, or a path I haven't fully traced.
--
-- Rather than fix every path, this logs the symptom: every chat
-- work_queue row at INSERT time has its payload.requested_model
-- compared to the pipeline stage's declared model. Mismatch → log to
-- stewards.model_substitutions so humans see WHEN this happens.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. model_substitutions log table.
-- ---------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS stewards.model_substitutions (
    id                bigserial PRIMARY KEY,
    work_queue_id     bigint REFERENCES stewards.work_queue(id) ON DELETE CASCADE,
    work_item_id      uuid,
    pipeline_family   text,
    stage_name        text,
    pipeline_model    text,    -- what the pipeline declared
    requested_model   text,    -- what the dispatch actually requested
    session_id        text,
    detected_at       timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS model_substitutions_recent
    ON stewards.model_substitutions (detected_at DESC);

CREATE INDEX IF NOT EXISTS model_substitutions_work_item
    ON stewards.model_substitutions (work_item_id);

COMMENT ON TABLE stewards.model_substitutions IS
'Batch L.1.1.15: log of every chat dispatch where the requested_model differs from the pipeline-declared stage model. Surfaces silent model swapping (steward retries, escalation, model_override, etc.) so humans can audit.';


-- ---------------------------------------------------------------------
-- 2. Trigger function: detect substitution at chat enqueue time.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.trigger_log_model_substitution()
RETURNS trigger LANGUAGE plpgsql AS $FN$
DECLARE
    v_pipeline_family text;
    v_stage_name      text;
    v_pipeline_model  text;
    v_requested       text;
    v_work_item_id    text;
    v_session_id      text;
BEGIN
    v_requested := NEW.payload ->> 'requested_model';
    IF v_requested IS NULL THEN RETURN NEW; END IF;

    v_pipeline_family := NEW.payload ->> '_pipeline_family';
    v_stage_name      := NEW.payload ->> '_stage_name';
    v_work_item_id    := NEW.payload ->> '_work_item_id';
    v_session_id      := NEW.payload ->> 'session_id';

    -- We can only check substitution when the chat carries pipeline+stage markers.
    IF v_pipeline_family IS NULL OR v_stage_name IS NULL THEN RETURN NEW; END IF;

    SELECT s ->> 'model' INTO v_pipeline_model
      FROM stewards.pipelines p,
           LATERAL jsonb_array_elements(p.stages) s
     WHERE p.family = v_pipeline_family
       AND (s ->> 'name') = v_stage_name
     LIMIT 1;

    IF v_pipeline_model IS NULL OR v_pipeline_model = v_requested THEN
        RETURN NEW;
    END IF;

    INSERT INTO stewards.model_substitutions
        (work_queue_id, work_item_id, pipeline_family, stage_name,
         pipeline_model, requested_model, session_id)
    VALUES
        (NEW.id,
         CASE WHEN v_work_item_id ~ '^[0-9a-f-]{36}$' THEN v_work_item_id::uuid ELSE NULL END,
         v_pipeline_family, v_stage_name,
         v_pipeline_model, v_requested, v_session_id);

    RAISE NOTICE 'model substitution: pipeline=%/% declared=% but requested=% (wq=%)',
        v_pipeline_family, v_stage_name, v_pipeline_model, v_requested, NEW.id;

    RETURN NEW;
END;
$FN$;

DROP TRIGGER IF EXISTS work_queue_log_model_substitution ON stewards.work_queue;

CREATE TRIGGER work_queue_log_model_substitution
AFTER INSERT ON stewards.work_queue
FOR EACH ROW
WHEN (NEW.kind = 'chat')
EXECUTE FUNCTION stewards.trigger_log_model_substitution();

COMMENT ON FUNCTION stewards.trigger_log_model_substitution() IS
'Batch L.1.1.15: AFTER INSERT trigger on chat work_queue rows. Compares payload.requested_model to the pipeline-stage declared model. Mismatch logged to model_substitutions and emits a NOTICE.';


-- =====================================================================
-- End of l29-model-substitution-log.sql
-- =====================================================================
