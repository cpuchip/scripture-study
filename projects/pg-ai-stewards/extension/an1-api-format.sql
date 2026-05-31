-- =====================================================================
-- Batch AN.1 — per-model api_format + work_queue stamp trigger
-- =====================================================================
-- opencode serves some models ONLY in Anthropic format (POST /messages),
-- not the OpenAI-compat /chat/completions the substrate speaks. This batch
-- records which format each model needs and stamps it onto every chat
-- work_queue row, so the bgworker (AN.2) can pick the right dispatch path.
--
-- The stamp lives in a BEFORE INSERT trigger on work_queue (not in
-- work_item_dispatch_stage) so it ALSO covers direct inserters — notably
-- enqueue_model_probe, which bypasses the dispatcher. That's required: to
-- probe qwen3.7-max through the new path, the probe row needs the stamp too.
--
-- Inert until AN.2 rebuilds the bgworker (which reads payload.api_format);
-- 'openai' (the default) is the existing path, so stamping it changes nothing.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. model_capability.api_format
-- ---------------------------------------------------------------------
ALTER TABLE stewards.model_capability
    ADD COLUMN IF NOT EXISTS api_format text NOT NULL DEFAULT 'openai';

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'model_capability_api_format_chk') THEN
        ALTER TABLE stewards.model_capability
            ADD CONSTRAINT model_capability_api_format_chk CHECK (api_format IN ('openai','anthropic'));
    END IF;
END $$;

COMMENT ON COLUMN stewards.model_capability.api_format IS
'AN.1: which gateway API shape the model needs — openai (/chat/completions, default) or anthropic (/messages, x-api-key + anthropic-version). Stamped onto chat work_queue payloads; the bgworker branches on it.';


-- ---------------------------------------------------------------------
-- 2. model_api_format(provider, model) -> 'openai' | 'anthropic'
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.model_api_format(p_provider text, p_model text)
RETURNS text LANGUAGE sql STABLE AS $$
    SELECT COALESCE(
        (SELECT api_format FROM stewards.model_capability
          WHERE provider = p_provider AND model = p_model),
        'openai'
    );
$$;

COMMENT ON FUNCTION stewards.model_api_format(text, text) IS
'AN.1: the dispatch API format for a model — defaults to openai for unrowed models.';


-- ---------------------------------------------------------------------
-- 3. Stamp payload.api_format on every chat work_queue insert.
-- ---------------------------------------------------------------------
-- BEFORE INSERT so it covers work_item_dispatch_stage AND direct inserters
-- (enqueue_model_probe, gate dispatches). Idempotent: leaves a pre-set
-- api_format alone. Looks up by the row's provider + requested_model.
CREATE OR REPLACE FUNCTION stewards.trigger_stamp_api_format()
RETURNS trigger LANGUAGE plpgsql AS $FN$
DECLARE
    v_model text;
    v_fmt   text;
BEGIN
    IF NEW.payload ? 'api_format' THEN
        RETURN NEW;  -- caller already specified
    END IF;
    v_model := COALESCE(NEW.payload ->> 'requested_model', NEW.payload -> 'body' ->> 'model');
    IF v_model IS NULL THEN
        RETURN NEW;
    END IF;
    v_fmt := stewards.model_api_format(NEW.provider, v_model);
    NEW.payload := NEW.payload || jsonb_build_object('api_format', v_fmt);
    RETURN NEW;
END;
$FN$;

DROP TRIGGER IF EXISTS work_queue_stamp_api_format ON stewards.work_queue;

CREATE TRIGGER work_queue_stamp_api_format
BEFORE INSERT ON stewards.work_queue
FOR EACH ROW
WHEN (NEW.kind = 'chat')
EXECUTE FUNCTION stewards.trigger_stamp_api_format();

COMMENT ON FUNCTION stewards.trigger_stamp_api_format() IS
'AN.1: BEFORE INSERT on chat work_queue rows — stamps payload.api_format from model_api_format(provider, requested_model) unless already set. Covers dispatch + the direct-insert probe path.';


-- ---------------------------------------------------------------------
-- 4. Seed: qwen3.7-max + minimax-m2.7 are Anthropic-format.
-- ---------------------------------------------------------------------
-- usable stays false until AN.3 verifies the path live (before AN.2 they
-- genuinely cannot be dispatched). ON CONFLICT only updates api_format so
-- qwen3.7-max keeps its existing (accurate) detail + usable=false.
INSERT INTO stewards.model_capability
    (provider, model, api_format, usable, supports_streaming, probe_detail, probed_via)
VALUES
    ('opencode_go', 'qwen3.7-max',  'anthropic', false, false,
     'Anthropic-format (needs /messages); usability pending the AN.2 dispatch path.', 'seed'),
    ('opencode_go', 'minimax-m2.7', 'anthropic', false, false,
     'Anthropic-format (needs /messages); usability pending the AN.2 dispatch path.', 'seed')
ON CONFLICT (provider, model) DO UPDATE
    SET api_format = EXCLUDED.api_format;


-- =====================================================================
-- Acceptance (AN.1):
--   1. model_api_format('opencode_go','qwen3.7-max')  = 'anthropic'
--      model_api_format('opencode_go','kimi-k2.6')    = 'openai'
--      model_api_format('opencode_go','never-heard')  = 'openai'
--   2. A chat work_queue insert (transactional) for qwen3.7-max gets
--      payload.api_format='anthropic'; for kimi-k2.6 gets 'openai'. ROLLBACK.
-- =====================================================================
