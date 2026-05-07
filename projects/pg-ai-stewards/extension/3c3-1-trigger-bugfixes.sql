-- =====================================================================
-- Phase 3c.3.1 — Trigger bug fixes from 3c.3 v1
--
-- Live-DB migration. Folds into extension/src/lib.rs at next intentional
-- rebuild (foldback debt: 14th file).
--
-- Builds on:
--   - Phase 1.5/1.6 (chat_post_internal — the function we're patching)
--   - Phase 3c.2 (handle_work_item_chat_completion trigger — bug fix)
--   - Phase 3c.3 v1 (the run that surfaced these bugs;
--     2026-05-07-pg-ai-stewards-3c3-v1-instructive-failure.md)
--
-- Fixes:
--   1. chat_post_internal — propagate `_*` payload markers from the
--      session's most recent chat row. Without this, continuation
--      chats (after tool_dispatch) lose the _watchman_pass_id /
--      _work_item_id markers, so the corresponding triggers only see
--      the FIRST chat per stage and miss the actual final chat.
--   2. handle_work_item_chat_completion — `v_is_final` was evaluating
--      to NULL when `loop_stop_reason IS NULL`, because `NULL IN (...)`
--      returns NULL not FALSE. PL/pgSQL `IF NOT NULL` doesn't take
--      the early-return branch, so the trigger advanced on
--      intermediate chats incorrectly. Fix: coalesce to false / guard
--      with IS NOT NULL.
--   3. agents.steps bump to 50 for non-watchman agents. The default 8
--      from 3a.1's import is too tight for real tool-using research
--      (agent step-exhausts before reaching synthesis). Watchman
--      agents stay at steps=1 (single-shot, no tools).
--
-- Token rollup undercounting (bug 3 in the v1 journal) is downstream
-- of fix 1: once continuation chats carry markers, the trigger fires
-- per-iteration and rolls up tokens correctly.
-- =====================================================================

-- ---------------------------------------------------------------------
-- Fix 1: chat_post_internal — inherit `_*` markers from session
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.chat_post_internal(
    p_agent_family text,
    p_model        text,
    p_session_id   text,
    p_provider     text
) RETURNS bigint
LANGUAGE plpgsql AS $func$
DECLARE
    v_body              jsonb;
    v_payload           jsonb;
    v_work_id           bigint;
    v_inherited_markers jsonb;
BEGIN
    v_body := stewards.dry_run_chat(
        p_agent_family, p_model, p_session_id, NULL);

    -- Phase 3c.3.1 fix: copy any payload keys starting with `_` from
    -- the most recent chat work_queue row in the same session.
    -- Continuation chats now inherit _watchman_pass_id, _work_item_id,
    -- _stage_name, _pipeline_family, _watchman_estimate, etc.
    --
    -- This is generic — works for any current or future marker scheme
    -- as long as the convention "marker keys start with underscore"
    -- holds. Watchman, work_items, and any future system that injects
    -- markers via the FIRST-chat-in-session pattern get correct
    -- propagation for free.
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

    v_payload := jsonb_build_object(
        'session_id',      p_session_id,
        'agent_family',    p_agent_family,
        'requested_model', p_model,
        'meta',            v_body->'_meta',
        'body',            (v_body - '_meta')
                           || jsonb_build_object('user', p_session_id)
    ) || coalesce(v_inherited_markers, '{}'::jsonb);

    INSERT INTO stewards.work_queue (kind, provider, payload)
    VALUES ('chat', p_provider, v_payload)
    RETURNING id INTO v_work_id;

    RETURN v_work_id;
END;
$func$;

COMMENT ON FUNCTION stewards.chat_post_internal(text, text, text, text) IS
'Phase 3c.3.1: continuation chat enqueuer. Inherits any `_*` payload markers from the most recent chat work_queue row in the same session, so triggers like handle_watchman_chat_completion and handle_work_item_chat_completion see ALL chats in their loop, not just the first.';

-- ---------------------------------------------------------------------
-- Fix 2: handle_work_item_chat_completion — coalesce v_is_final
-- ---------------------------------------------------------------------
-- The whole trigger function is rewritten to make the coalesce
-- behavior explicit and to add a defense for v_finish_reason being
-- NULL too (theoretical, but cheap to guard).
CREATE OR REPLACE FUNCTION stewards.handle_work_item_chat_completion()
RETURNS trigger
LANGUAGE plpgsql AS $func$
DECLARE
    v_work_item_id    uuid;
    v_stage_name      text;
    v_session_id      text;
    v_assistant       stewards.messages%ROWTYPE;
    v_finish_reason   text;
    v_loop_stop       text;
    v_has_tool_calls  boolean;
    v_is_final        boolean;
    v_stage_output    jsonb;
    v_next_stage      text;
    v_wi_after        stewards.work_items%ROWTYPE;
    v_msg_tokens_in   int;
    v_msg_tokens_out  int;
BEGIN
    -- WHEN clause prefilters; this is belt-and-suspenders.
    IF NEW.kind <> 'chat'
       OR (NEW.payload->>'_work_item_id') IS NULL THEN
        RETURN NEW;
    END IF;
    IF NEW.status NOT IN ('done', 'error') THEN
        RETURN NEW;
    END IF;
    IF OLD.status = NEW.status THEN
        RETURN NEW;
    END IF;

    v_work_item_id := (NEW.payload->>'_work_item_id')::uuid;
    v_stage_name   := NEW.payload->>'_stage_name';
    v_session_id   := NEW.payload->>'session_id';

    -- Error path: fail the work_item.
    IF NEW.status = 'error' THEN
        BEGIN
            PERFORM stewards.work_item_fail(
                v_work_item_id,
                format('chat dispatch failed at stage %s: %s',
                       v_stage_name,
                       coalesce(NEW.error, '(no error msg)')));
        EXCEPTION WHEN OTHERS THEN
            RAISE WARNING
                'work_item trigger work_item_fail() failed for %: %',
                v_work_item_id, SQLERRM;
        END;
        RETURN NEW;
    END IF;

    -- Done path: read the latest assistant message.
    SELECT * INTO v_assistant
      FROM stewards.messages
     WHERE session_id = v_session_id AND role = 'assistant'
     ORDER BY id DESC LIMIT 1;

    IF v_assistant.id IS NULL THEN
        BEGIN
            PERFORM stewards.work_item_fail(
                v_work_item_id,
                format('no assistant message for stage %s session %s',
                       v_stage_name, v_session_id));
        EXCEPTION WHEN OTHERS THEN NULL;
        END;
        RETURN NEW;
    END IF;

    -- Token rollup — applies to BOTH intermediate and final chats.
    v_msg_tokens_in  := coalesce(v_assistant.tokens_in,  0);
    v_msg_tokens_out := coalesce(v_assistant.tokens_out, 0)
                      + coalesce(v_assistant.reasoning_tokens, 0);

    UPDATE stewards.work_items
       SET tokens_in  = tokens_in  + v_msg_tokens_in,
           tokens_out = tokens_out + v_msg_tokens_out,
           updated_at = now()
     WHERE id = v_work_item_id;

    -- Final-vs-intermediate detection.
    --
    -- Phase 3c.3.1 fix: every clause guarded against NULL so the
    -- whole expression collapses to a true boolean (never NULL).
    -- The original 3c.2 version had:
    --   v_is_final := (NOT v_has_tool_calls AND finish_reason IN (...))
    --              OR loop_stop IN (...);
    -- When loop_stop was NULL, `NULL IN (...)` returned NULL, then
    -- `FALSE OR NULL` was NULL, then `IF NOT NULL` skipped the
    -- early-return — and the trigger advanced on intermediate chats.
    v_finish_reason  := v_assistant.finish_reason;
    v_loop_stop      := NEW.result->>'loop_stop_reason';
    v_has_tool_calls := v_assistant.tool_calls IS NOT NULL
                        AND jsonb_typeof(v_assistant.tool_calls) = 'array'
                        AND jsonb_array_length(v_assistant.tool_calls) > 0;

    v_is_final := coalesce(
        (NOT v_has_tool_calls
         AND v_finish_reason IS NOT NULL
         AND v_finish_reason IN ('stop', 'length', 'content_filter'))
        OR (v_loop_stop IS NOT NULL
            AND v_loop_stop IN ('steps_exhausted', 'truncated_tool_calls')),
        false
    );

    IF NOT v_is_final THEN
        RETURN NEW;
    END IF;

    -- Build stage output.
    v_stage_output := jsonb_build_object(
        'output',           v_assistant.content,
        'model',            v_assistant.model,
        'tokens_in',        v_msg_tokens_in,
        'tokens_out',       v_msg_tokens_out,
        'finish_reason',    v_finish_reason
    );
    IF v_loop_stop IS NOT NULL THEN
        v_stage_output := v_stage_output
            || jsonb_build_object('loop_stop_reason', v_loop_stop);
    END IF;

    BEGIN
        v_next_stage := stewards.work_item_advance(v_work_item_id, v_stage_output);
    EXCEPTION WHEN OTHERS THEN
        RAISE WARNING
            'work_item trigger work_item_advance() failed for %: %',
            v_work_item_id, SQLERRM;
        BEGIN
            PERFORM stewards.work_item_fail(v_work_item_id,
                'auto-advance failed: ' || SQLERRM);
        EXCEPTION WHEN OTHERS THEN NULL;
        END;
        RETURN NEW;
    END;

    IF v_next_stage IS NULL THEN
        RETURN NEW;
    END IF;

    SELECT * INTO v_wi_after FROM stewards.work_items WHERE id = v_work_item_id;
    IF v_wi_after.status <> 'pending' THEN
        RETURN NEW;
    END IF;

    -- Token budget gate (cost guard).
    IF v_wi_after.token_budget IS NOT NULL
       AND (v_wi_after.tokens_in + v_wi_after.tokens_out)
            >= v_wi_after.token_budget THEN
        UPDATE stewards.work_items
           SET status     = 'awaiting_review',
               error      = format(
                   'token budget exhausted at stage %s (%s/%s); '
                   || 'next stage %s not auto-dispatched',
                   v_stage_name,
                   v_wi_after.tokens_in + v_wi_after.tokens_out,
                   v_wi_after.token_budget,
                   v_next_stage),
               updated_at = now()
         WHERE id = v_work_item_id;
        RETURN NEW;
    END IF;

    -- Auto-dispatch next stage.
    BEGIN
        PERFORM stewards.work_item_dispatch_stage(v_work_item_id);
    EXCEPTION WHEN OTHERS THEN
        RAISE WARNING
            'work_item trigger dispatch_stage() failed for %: %',
            v_work_item_id, SQLERRM;
        UPDATE stewards.work_items
           SET status     = 'awaiting_review',
               error      = format('auto-dispatch of stage %s failed: %s',
                                   v_next_stage, SQLERRM),
               updated_at = now()
         WHERE id = v_work_item_id;
    END;

    RETURN NEW;
END;
$func$;

-- The trigger declaration itself is unchanged, but Postgres requires
-- the trigger be re-bound to pick up the new function body. CREATE
-- OR REPLACE FUNCTION already updates the body in place; the
-- existing trigger now uses the new function. No need to drop+recreate.

-- ---------------------------------------------------------------------
-- Fix 3: bump agents.steps for non-watchman agents
--
-- The 3a.1 import set steps=8 as a placeholder default. Real
-- tool-using research (study, plan, etc.) routinely needs 20+
-- iterations. Bumping to 50 is generous but safe — the agent stops
-- early when it produces a `finish_reason='stop'` message; the limit
-- only kicks in if the agent never reaches a synthesis answer.
--
-- Watchman agents stay at steps=1 (single-shot, no tools by design;
-- structural enforcement of the consolidator's "render one verdict"
-- contract).
-- ---------------------------------------------------------------------
UPDATE stewards.agents
   SET steps = 50
 WHERE family NOT LIKE 'watchman%'
   AND steps < 50;
