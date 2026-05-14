-- =====================================================================
-- Batch L.1.1.2 — Agent-aware extraction threshold
-- =====================================================================
-- Replaces the constant 60000 chars in the K.1 INSERT trigger with a
-- dynamic threshold that scales with the consuming agent's budget.
--
-- Formula: threshold_chars = effective_budget(session, tokens) * 3.5 / N
--   where 3.5 chars/token is a reasonable approximation and N=16
--   means "1/16th of the agent's budget" triggers engram extraction.
--
-- For provider context_window 262144 (kimi-k2.6): threshold ~= 57K chars
--   (matches the prior 60K constant — kimi was the implicit assumption).
-- For working_budget 64000 (a smaller agent): threshold ~= 14K chars
--   (drops bacteriopolis's 10K-15K medium messages into the extraction
--    path — the main carry-forward from L gets fixed).
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. effective_extraction_threshold helper.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.effective_extraction_threshold(
    p_session_id text,
    p_stage_name text DEFAULT NULL
) RETURNS integer LANGUAGE plpgsql STABLE AS $FN$
DECLARE
    v_budget_tokens int;
    -- Ratified L.1.1: 1/16th of budget triggers extraction.
    v_ratio_n         constant int  := 16;
    -- Approximation: avg chars per token for English text.
    v_chars_per_token constant numeric := 3.5;
    v_threshold     int;
BEGIN
    v_budget_tokens := stewards.effective_budget(p_session_id, p_stage_name);
    IF v_budget_tokens IS NULL OR v_budget_tokens <= 0 THEN
        -- Conservative floor: same as the prior K.1 constant.
        RETURN 60000;
    END IF;

    v_threshold := ((v_budget_tokens::numeric * v_chars_per_token) / v_ratio_n)::int;

    -- Floor and ceiling so we don't drop below a sensible minimum
    -- (would extract too aggressively) or above the prior K.1 default.
    RETURN GREATEST(LEAST(v_threshold, 60000), 5000);
END;
$FN$;

COMMENT ON FUNCTION stewards.effective_extraction_threshold(text, text) IS
'Batch L.1.1.2: chars-threshold above which engram extraction fires for a role=tool message in this session. Scales with effective_budget (tokens * 3.5 chars/tok / 16). Floored at 5000 chars, ceilinged at 60000 chars (prior K.1 constant). 5K floor prevents over-extraction; 60K ceiling matches existing behavior for the kimi-class default.';


-- ---------------------------------------------------------------------
-- 2. Rewrite trigger function to use the dynamic threshold.
-- ---------------------------------------------------------------------
-- The trigger WHERE clause stays permissive (length > 5000) so the fn
-- is called for medium-sized inputs; the fn does the dynamic check.

CREATE OR REPLACE FUNCTION stewards.trigger_extract_engrams_on_large_tool()
RETURNS trigger LANGUAGE plpgsql AS $FN$
DECLARE
    v_threshold int;
BEGIN
    v_threshold := stewards.effective_extraction_threshold(NEW.session_id);

    IF length(NEW.content) <= v_threshold THEN
        RETURN NEW;
    END IF;

    BEGIN
        PERFORM stewards.extract_engrams(NEW.id);
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'trigger_extract_engrams_on_large_tool: enqueue failed for msg=%: %',
            NEW.id, SQLERRM;
    END;

    RETURN NEW;
END;
$FN$;

-- Replace the trigger to broaden the WHERE clause floor.
DROP TRIGGER IF EXISTS messages_extract_engrams_on_large_tool ON stewards.messages;

CREATE TRIGGER messages_extract_engrams_on_large_tool
AFTER INSERT ON stewards.messages
FOR EACH ROW
WHEN (NEW.role = 'tool' AND length(NEW.content) > 5000 AND NEW.engrams IS NULL)
EXECUTE FUNCTION stewards.trigger_extract_engrams_on_large_tool();

COMMENT ON FUNCTION stewards.trigger_extract_engrams_on_large_tool() IS
'Batch L.1.1.2: AFTER INSERT trigger handler. Calls effective_extraction_threshold(session_id) to compute the agent-aware threshold dynamically; only enqueues extraction if NEW.content exceeds it. The trigger WHERE clause uses a permissive 5000-char floor to avoid invoking the fn on tiny tool results.';


-- =====================================================================
-- End of l12-agent-aware-extraction-threshold.sql
-- =====================================================================
