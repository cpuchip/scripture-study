-- =====================================================================
-- Batch M.4 — Model auto-probe (keeps model_capability current)
-- =====================================================================
-- The M.1 capability signal is only as good as the verdicts in it. This
-- batch lets the substrate test a model itself, over the EXACT streaming
-- path it dispatches with — so it catches the subtle failure (GLM streams
-- empty) as well as the hard one (qwen3.7-max is gateway-rejected).
--
-- Pure SQL, no bgworker change, no rebuild:
--   - enqueue_model_probe(provider, model) inserts a tiny chat DIRECTLY into
--     work_queue (bypassing work_item_dispatch_stage, so the M.2 substitution
--     does NOT swap the very model we are testing), tagged _probe.
--   - the bgworker runs it on its normal chat path (writes the assistant
--     message, marks work_queue done/error).
--   - a trigger on work_queue's terminal transition reads the outcome and
--     records the verdict into model_capability (probed_via='auto-probe').
--
-- Three outcomes, all handled:
--   done + non-empty assistant content -> usable=true,  streaming=true
--   done + empty assistant content     -> usable=false, streaming=false  (GLM)
--   error (HTTP 4xx etc.)              -> usable=false, streaming=false  (qwen3.7-max)
--
-- M.5 adds the scheduler (enqueue_due_model_probes) wired to the watchman.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. enqueue_model_probe(provider, model) -> work_queue id
-- ---------------------------------------------------------------------
-- A minimal OpenAI chat request. The bgworker (chat()) POSTs payload.body
-- verbatim + stream:true, so a {model, messages, max_tokens} body is all
-- that's needed — no dry_run_chat / agent composition (leaner + cheaper:
-- a probe is ~one short reply, free on the $0 models). agent_family is an
-- attribution label only; it need not be a registered agent.
CREATE OR REPLACE FUNCTION stewards.enqueue_model_probe(
    p_provider text,
    p_model    text
) RETURNS bigint
LANGUAGE plpgsql AS $func$
DECLARE
    v_session  text;
    v_payload  jsonb;
    v_work_id  bigint;
BEGIN
    v_session := substring(
        'probe--' || p_provider || '--' || p_model || '--'
        || to_char(clock_timestamp(), 'YYYYMMDDHH24MISSUS')
        FROM 1 FOR 200);

    -- The session must exist so the bgworker's assistant-message INSERT lands.
    INSERT INTO stewards.sessions (id, label, kind)
    VALUES (v_session, format('model probe %s/%s', p_provider, p_model), 'agent')
    ON CONFLICT (id) DO NOTHING;

    v_payload := jsonb_build_object(
        'session_id',      v_session,
        'agent_family',    'model-probe',
        'requested_model', p_model,
        'tools_disabled',  true,
        'body', jsonb_build_object(
            'model',      p_model,
            'max_tokens', 256,
            'messages',   jsonb_build_array(
                jsonb_build_object(
                    'role', 'user',
                    'content', 'Reply with exactly: OK'
                )
            )
        ),
        '_probe', jsonb_build_object('provider', p_provider, 'model', p_model)
    );

    -- Direct work_queue insert — NOT work_item_dispatch_stage — so the M.2
    -- capability substitution does not swap the model under test.
    INSERT INTO stewards.work_queue (kind, provider, payload)
    VALUES ('chat', p_provider, v_payload)
    RETURNING id INTO v_work_id;

    RETURN v_work_id;
END;
$func$;

COMMENT ON FUNCTION stewards.enqueue_model_probe(text, text) IS
'Batch M.4: enqueue a tiny streaming chat to test whether (provider, model) is dispatchable. Direct work_queue insert (bypasses the M.2 substitution gate). The work_queue terminal-transition trigger records the verdict into model_capability.';


-- ---------------------------------------------------------------------
-- 2. trigger_resolve_model_probe — record the verdict on terminal status.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.trigger_resolve_model_probe()
RETURNS trigger LANGUAGE plpgsql AS $FN$
DECLARE
    v_provider text;
    v_model    text;
    v_session  text;
    v_content  text;
    v_finish   text;
    v_usable   boolean;
    v_detail   text;
BEGIN
    v_provider := NEW.payload -> '_probe' ->> 'provider';
    v_model    := NEW.payload -> '_probe' ->> 'model';
    v_session  := NEW.payload ->> 'session_id';

    IF NEW.status = 'error' THEN
        v_usable := false;
        v_detail := 'auto-probe: dispatch error: '
                    || left(COALESCE(NEW.error, '(no error text)'), 240);
    ELSE
        -- done: did content arrive over the streaming path?
        SELECT content, finish_reason INTO v_content, v_finish
          FROM stewards.messages
         WHERE session_id = v_session AND role = 'assistant'
         ORDER BY id DESC LIMIT 1;

        v_usable := length(trim(COALESCE(v_content, ''))) > 0;
        IF v_usable THEN
            v_detail := format('auto-probe: ok — %s content chars, finish=%s',
                               length(v_content), COALESCE(v_finish, '(null)'));
        ELSE
            v_detail := format('auto-probe: streaming returned empty content (0 chars), finish=%s',
                               COALESCE(v_finish, '(null)'));
        END IF;
    END IF;

    INSERT INTO stewards.model_capability
        (provider, model, usable, supports_streaming, last_probed_at, probe_detail, probed_via)
    VALUES
        (v_provider, v_model, v_usable, v_usable, now(), v_detail, 'auto-probe')
    ON CONFLICT (provider, model) DO UPDATE
    SET usable             = EXCLUDED.usable,
        supports_streaming = EXCLUDED.supports_streaming,
        last_probed_at     = now(),
        probe_detail       = EXCLUDED.probe_detail,
        probed_via         = 'auto-probe',
        updated_at         = now();

    RAISE NOTICE 'auto-probe verdict: %/% usable=% (%)',
        v_provider, v_model, v_usable, v_detail;

    RETURN NEW;
END;
$FN$;

DROP TRIGGER IF EXISTS work_queue_resolve_model_probe ON stewards.work_queue;

CREATE TRIGGER work_queue_resolve_model_probe
AFTER UPDATE ON stewards.work_queue
FOR EACH ROW
WHEN (NEW.status IN ('done', 'error')
      AND OLD.status IS DISTINCT FROM NEW.status
      AND NEW.payload -> '_probe' IS NOT NULL)
EXECUTE FUNCTION stewards.trigger_resolve_model_probe();

COMMENT ON FUNCTION stewards.trigger_resolve_model_probe() IS
'Batch M.4: on a probe work_queue row reaching done/error, records the verdict into model_capability. error -> unusable; done+empty content -> unusable (streaming-empty, e.g. GLM); done+content -> usable. probed_via=auto-probe.';


-- =====================================================================
-- Acceptance (live, costs ~$0 — free models + errors generate no tokens):
--   1. SELECT stewards.enqueue_model_probe('opencode_go','deepseek-v4-flash');
--      after the bgworker runs it: model_capability shows usable=true,
--      probed_via='auto-probe', probe_detail LIKE 'auto-probe: ok%'.
--   2. SELECT stewards.enqueue_model_probe('opencode_go','glm-5');
--      -> usable=false, probe_detail LIKE '%empty content%'.
--   3. SELECT stewards.enqueue_model_probe('opencode_go','qwen3.7-max');
--      -> usable=false, probe_detail LIKE '%dispatch error%oa-compat%'.
-- =====================================================================
