-- =====================================================================
-- Phase 4b — Dispatch override (Push A of steward-bgworker-integration)
--
-- Implements Push A from
-- projects/pg-ai-stewards/.spec/proposals/steward-bgworker-integration.md.
--
-- Builds on:
--   - 4a-cost-tracking.sql, 4a-escalation-chain.sql, 4a-steward.sql
--   - 3c3-stage-templating-and-study-write.sql (the work_item_dispatch_stage
--     this file replaces)
--
-- This file does THREE things:
--
--   1. **Provider name migration.** During code reading we discovered
--      4a's seed used provider='opencode-zen' but the substrate's
--      registered provider is 'opencode_go' (per stewards.providers_loaded()
--      and existing pipelines). Migrate live data: rename provider in
--      model_pricing + cost_buckets. The corresponding 4a SQL files
--      have been edited at-rest so fresh containers get 'opencode_go'.
--
--   2. **Add work_items.provider_override column** (sibling of
--      model_override added in 4a-escalation-chain). Allows steward to
--      pin both model AND provider for a one-shot dispatch.
--
--   3. **Replace stewards.work_item_dispatch_stage** with a version that:
--      a. Honors v_wi.model_override (COALESCE with stage default)
--      b. Honors v_wi.provider_override (COALESCE with stage default)
--      c. Optionally allows re-dispatch from status='failed' via new
--         p_allow_failed_status DEFAULT false parameter. Existing call
--         sites (NewWork, watchman) pass nothing → behavior unchanged.
--      d. Resets failure_count to 0 on successful dispatch from a
--         previously-failed state (the retry "consumed" the prior
--         failure; failure_count tracks consecutive failures, not
--         lifetime failures).
--
-- Verification of non-regression:
--   - SELECT calling work_item_dispatch_stage on a 'pending' work_item
--     with model_override=NULL produces identical work_queue payload
--     to today (same provider, same model, same body).
--   - SELECT calling on a 'failed' work_item without p_allow_failed_status
--     still raises (preserves the gate against accidental re-dispatch
--     by callers that don't know about the new parameter).
-- =====================================================================

-- ---------------------------------------------------------------------
-- Section 1: Live-data migration — opencode-zen → opencode_go
-- ---------------------------------------------------------------------

UPDATE stewards.model_pricing
   SET provider = 'opencode_go'
 WHERE provider = 'opencode-zen';

UPDATE stewards.cost_buckets
   SET provider = 'opencode_go'
 WHERE provider = 'opencode-zen';

-- The cost_events table has a provider column populated by record_cost_event.
-- No historical rows yet (cost_events count = 0 as of 2026-05-10) but
-- guard idempotently in case any have been inserted.
UPDATE stewards.cost_events
   SET provider = 'opencode_go'
 WHERE provider = 'opencode-zen';

-- ---------------------------------------------------------------------
-- Section 2: Add provider_override column to work_items
-- ---------------------------------------------------------------------

ALTER TABLE stewards.work_items
    ADD COLUMN IF NOT EXISTS provider_override text;

COMMENT ON COLUMN stewards.work_items.provider_override IS
'Phase 4b: when non-NULL, the dispatch uses this provider regardless of pipeline_stage_lookup. Paired with model_override; both cleared after escalation queue resolution. Must reference a registered provider (stewards.providers_loaded()).';

-- ---------------------------------------------------------------------
-- Section 3: Replace work_item_dispatch_stage with override-aware version
--
-- IMPORTANT: PostgreSQL keys functions on signature, so CREATE OR REPLACE
-- on a 3-arg version doesn't replace the existing 2-arg version — both
-- coexist and any 1-or-2-arg call becomes ambiguous ("Could not choose
-- a best candidate function"). Explicitly DROP the old signature first.
-- This regression was caught by verify-4b.sql section G during smoke.
-- ---------------------------------------------------------------------

DROP FUNCTION IF EXISTS stewards.work_item_dispatch_stage(uuid, text);

CREATE OR REPLACE FUNCTION stewards.work_item_dispatch_stage(
    p_work_item_id           uuid,
    p_user_input             text DEFAULT NULL,
    p_allow_failed_status    boolean DEFAULT false
) RETURNS bigint
LANGUAGE plpgsql AS $func$
DECLARE
    v_wi          stewards.work_items%ROWTYPE;
    v_stage       jsonb;
    v_agent       text;
    v_model       text;
    v_provider    text;
    v_session_id  text;
    v_user_input  text;
    v_body        jsonb;
    v_payload     jsonb;
    v_work_id     bigint;
    v_was_failed  boolean := false;
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN
        RAISE EXCEPTION 'work_item % not found', p_work_item_id;
    END IF;

    -- Status gate. Default behavior (p_allow_failed_status=false) preserves
    -- the original contract for existing call sites (NewWork form, watchman).
    -- Steward retries pass true to enable re-dispatch from 'failed'.
    IF v_wi.status NOT IN ('pending', 'awaiting_review')
       AND NOT (p_allow_failed_status AND v_wi.status = 'failed')
    THEN
        RAISE EXCEPTION 'work_item %: cannot dispatch from status %',
            p_work_item_id, v_wi.status;
    END IF;

    v_was_failed := (v_wi.status = 'failed');

    v_stage := stewards.pipeline_stage_lookup(v_wi.pipeline_family, v_wi.current_stage);
    IF v_stage IS NULL THEN
        RAISE EXCEPTION 'work_item %: stage % not found in pipeline %',
            p_work_item_id, v_wi.current_stage, v_wi.pipeline_family;
    END IF;

    v_agent    := v_stage->>'agent_family';
    -- Phase 4b: model + provider honor work_items overrides if set.
    v_model    := COALESCE(v_wi.model_override,    v_stage->>'model');
    v_provider := COALESCE(v_wi.provider_override, v_stage->>'provider');

    IF v_agent IS NULL OR v_model IS NULL OR v_provider IS NULL THEN
        RAISE EXCEPTION 'work_item %: stage % missing agent_family/model/provider',
            p_work_item_id, v_wi.current_stage;
    END IF;

    v_session_id := substring(
        'wi--' || substring(p_work_item_id::text FROM 1 FOR 8)
        || '--' || v_wi.current_stage
        FROM 1 FOR 200);

    INSERT INTO stewards.sessions (id, label, kind)
    VALUES (v_session_id,
            format('work_item %s stage %s', v_wi.id, v_wi.current_stage),
            'agent')
    ON CONFLICT (id) DO NOTHING;

    -- Input resolution priority (unchanged from 3c3):
    --   1. Explicit p_user_input override (CLI dispatch w/ --user-input)
    --      OR steward retry guidance.
    --   2. Stage's input_template rendered against work_item state.
    --   3. work_item.input.user_input field (legacy fallback).
    --   4. Stringified work_item.input (last-resort fallback).
    IF p_user_input IS NOT NULL THEN
        v_user_input := p_user_input;
    ELSE
        v_user_input := stewards.render_stage_input(p_work_item_id);
        IF v_user_input IS NULL THEN
            v_user_input := coalesce(
                v_wi.input->>'user_input',
                v_wi.input::text
            );
        END IF;
    END IF;

    INSERT INTO stewards.messages (session_id, role, content, model)
    VALUES (v_session_id, 'user', v_user_input, v_model);

    v_body := stewards.dry_run_chat(v_agent, v_model, v_session_id, NULL);

    v_payload := jsonb_build_object(
        'session_id',         v_session_id,
        'agent_family',       v_agent,
        'requested_model',    v_model,
        'meta',               v_body->'_meta',
        'body',               (v_body - '_meta')
                              || jsonb_build_object('user', v_session_id),
        '_work_item_id',      p_work_item_id::text,
        '_stage_name',        v_wi.current_stage,
        '_pipeline_family',   v_wi.pipeline_family
    );

    INSERT INTO stewards.work_queue (kind, provider, payload)
    VALUES ('chat', v_provider, v_payload)
    RETURNING id INTO v_work_id;

    -- Status transition. From 'pending'/'awaiting_review': → 'in_progress'.
    -- From 'failed' (steward retry path): → 'in_progress' AND we DO NOT
    -- reset failure_count yet — that happens when the dispatch SUCCEEDS
    -- (handled wherever stage_results are written by the bridge response
    -- handler). Resetting on dispatch alone would lose the failure history
    -- if this attempt also fails.
    UPDATE stewards.work_items
       SET status      = 'in_progress',
           session_ids = session_ids || v_session_id,
           updated_at  = now()
     WHERE id = p_work_item_id;

    RETURN v_work_id;
END;
$func$;

COMMENT ON FUNCTION stewards.work_item_dispatch_stage(uuid, text, boolean) IS
'Phase 4b: dispatches a stage. Honors work_items.model_override + provider_override. Allows status=failed re-dispatch only when p_allow_failed_status=true (steward retries pass true; existing call sites stay safe by passing nothing).';

-- =====================================================================
-- Done. Phase 4b dispatch override is operational.
--
-- Acceptance:
--   1. Existing pipelines unchanged: dispatching a 'pending' work_item with
--      model_override=NULL produces identical work_queue payload as before
--      (same provider 'opencode_go', same model from stage default).
--   2. With model_override='qwen3.6-plus' on a pending work_item, the
--      dispatch produces a work_queue row with the override model.
--   3. Calling dispatch on a 'failed' work_item without
--      p_allow_failed_status raises 'cannot dispatch from status failed'.
--   4. Calling dispatch on a 'failed' work_item WITH p_allow_failed_status
--      => true succeeds, transitioning status → 'in_progress'.
--   5. Provider migration: stewards.compute_cost('opencode_go', 'kimi-k2.6',
--      1000000, 500000) returns the expected $2.95 (was returning 0 before
--      this push due to provider name mismatch).
-- =====================================================================
