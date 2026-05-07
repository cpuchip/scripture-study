-- =====================================================================
-- Phase 3c.2 — Auto-advance work_items on chat completion
--
-- Live-DB migration. Folds into extension/src/lib.rs at next
-- intentional rebuild (foldback debt: 11th file).
--
-- Builds on:
--   - Phase 3c.1 (work_items, transition functions, payload markers)
--   - Phase 2.7b.1 (the AFTER UPDATE trigger pattern this mirrors)
--   - Phase 1.6 (chat → tool_dispatch → continuation chat loop)
--
-- This file adds an AFTER UPDATE trigger on stewards.work_queue that
-- harvests completed work_item chats. When a chat dispatched by
-- stewards.work_item_dispatch_stage() lands done/error, the trigger:
--   1. Rolls up tokens_in/tokens_out into the parent work_item
--      (always — even on intermediate tool_call iterations).
--   2. Detects whether the chat was final (clean stop) or
--      intermediate (still has tool_calls pending in the loop).
--   3. On final: calls work_item_advance with structured stage_output.
--      If the next stage exists AND auto_advance=true AND the
--      work_item's token_budget hasn't been exceeded, calls
--      work_item_dispatch_stage to enqueue the next chat. Otherwise
--      leaves status=awaiting_review for human review.
--   4. On error: calls work_item_fail.
--
-- Same pattern + WHEN-clause prefilter as Phase 2.7b.1's
-- handle_watchman_chat_completion. Keeps the bgworker generic.
-- =====================================================================

-- ---------------------------------------------------------------------
-- handle_work_item_chat_completion — trigger function
--
-- Defensive everywhere (every record_*/advance/dispatch call wrapped
-- in BEGIN/EXCEPTION) so a bug in the harvester never breaks the
-- bgworker's status flip on the underlying work_queue row.
-- ---------------------------------------------------------------------
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
    v_has_tool_calls  bool;
    v_is_final        bool;
    v_stage_output    jsonb;
    v_next_stage      text;
    v_wi_after        stewards.work_items%ROWTYPE;
    v_msg_tokens_in   int;
    v_msg_tokens_out  int;
BEGIN
    -- The WHEN clause prefilters; this is belt-and-suspenders.
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

    -- ----- error path: fail the work_item, no rollup needed -----
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

    -- ----- done path: read the latest assistant message -----
    SELECT * INTO v_assistant
      FROM stewards.messages
     WHERE session_id = v_session_id AND role = 'assistant'
     ORDER BY id DESC LIMIT 1;

    IF v_assistant.id IS NULL THEN
        -- No assistant message? Can't advance. Mark failed.
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
    -- The work_item's tokens_in/out tracks ALL tokens spent
    -- (including tool-loop iterations).
    v_msg_tokens_in  := coalesce(v_assistant.tokens_in,  0);
    v_msg_tokens_out := coalesce(v_assistant.tokens_out, 0)
                      + coalesce(v_assistant.reasoning_tokens, 0);

    UPDATE stewards.work_items
       SET tokens_in  = tokens_in  + v_msg_tokens_in,
           tokens_out = tokens_out + v_msg_tokens_out,
           updated_at = now()
     WHERE id = v_work_item_id;

    -- Decide: is this the FINAL chat in this stage, or an intermediate
    -- chat that will continue via tool_dispatch?
    v_finish_reason  := v_assistant.finish_reason;
    v_loop_stop      := NEW.result->>'loop_stop_reason';
    v_has_tool_calls := v_assistant.tool_calls IS NOT NULL
                        AND jsonb_typeof(v_assistant.tool_calls) = 'array'
                        AND jsonb_array_length(v_assistant.tool_calls) > 0;

    -- A chat is "final" (work_item should advance) when:
    --   (a) clean stop: no tool_calls AND finish_reason ∈ stop|length|content_filter
    --   (b) loop stopped intentionally: loop_stop_reason ∈
    --       (steps_exhausted, truncated_tool_calls)
    -- Intermediate (don't advance) when finish_reason='tool_calls'
    -- and the chat handler enqueued a tool_dispatch continuation.
    v_is_final := (NOT v_has_tool_calls
                   AND v_finish_reason IN ('stop', 'length', 'content_filter'))
                OR v_loop_stop IN ('steps_exhausted', 'truncated_tool_calls');

    IF NOT v_is_final THEN
        -- Intermediate. Token rollup already happened above.
        -- Wait for the next chat (after tool_dispatch + continuation).
        RETURN NEW;
    END IF;

    -- Build the stage output. Includes loop_stop_reason when present
    -- so downstream stages can see "the prior stage hit step budget."
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

    -- Advance: records output in stage_results, transitions current_stage,
    -- sets status='pending' (auto_advance) or 'awaiting_review'.
    -- Returns next stage name or NULL when terminal.
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

    -- Terminal stage → work_item is now status=completed. Done.
    IF v_next_stage IS NULL THEN
        RETURN NEW;
    END IF;

    -- Re-fetch to check status. work_item_advance set it to 'pending'
    -- (auto_advance=true on the previous stage) or 'awaiting_review'
    -- (auto_advance=false). Only auto-dispatch when 'pending'.
    SELECT * INTO v_wi_after FROM stewards.work_items WHERE id = v_work_item_id;
    IF v_wi_after.status <> 'pending' THEN
        RETURN NEW;
    END IF;

    -- Token budget gate (cost guard). If the work_item carried a budget
    -- AND we've already hit it, don't dispatch the next stage. Mark
    -- awaiting_review with an explanatory error so the human can decide
    -- whether to bump the budget or cancel.
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

    -- Auto-dispatch next stage. If dispatch fails, mark awaiting_review
    -- (don't fail the whole work_item — the prior stage's results are
    -- valid; the human just needs to decide what to do).
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

COMMENT ON FUNCTION stewards.handle_work_item_chat_completion() IS
'Phase 3c.2: AFTER UPDATE trigger function on work_queue. When a chat row dispatched by stewards.work_item_dispatch_stage() lands done/error, advances the parent work_item: rolls up tokens, detects intermediate-vs-final, calls work_item_advance, and auto-dispatches the next stage (subject to token_budget + auto_advance gates). All side effects in the same tx as the work_queue status flip.';

-- Drop and recreate the trigger so re-applying this file is idempotent.
DROP TRIGGER IF EXISTS work_item_advance_completion ON stewards.work_queue;

CREATE TRIGGER work_item_advance_completion
    AFTER UPDATE OF status ON stewards.work_queue
    FOR EACH ROW
    WHEN ((NEW.kind = 'chat')
          AND (NEW.payload ? '_work_item_id')
          AND (NEW.status IN ('done', 'error'))
          AND (OLD.status IS DISTINCT FROM NEW.status))
    EXECUTE FUNCTION stewards.handle_work_item_chat_completion();
