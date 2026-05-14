-- =====================================================================
-- L.1.1.14 — REVIEW prefix verify gate
-- =====================================================================
-- Auto-advance to maturity=verified currently fires when the review
-- stage's chat completes — regardless of content. Bacteriopolis hit
-- this: review's output was 'I've carefully read your message... the
-- draft does not appear in this message.' Maturity advanced anyway.
--
-- Fix: BEFORE UPDATE trigger on work_items. If maturity is being set
-- to 'verified' AND the work_item is on a review-style stage, the
-- stage's output text must start with 'REVIEW: passes' or 'REVIEW:
-- revised'. Otherwise the maturity advance is suppressed (NEW.maturity
-- := OLD.maturity) and quarantine_reason captures why.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. Helper: does the review output text pass the gate?
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.review_output_passes_gate(p_output_text text)
RETURNS boolean LANGUAGE sql IMMUTABLE AS $$
    -- Note: PostgreSQL POSIX regex uses \y for word boundary, not \b.
    -- We use ~* (case-insensitive) and skip the trailing word-boundary
    -- since (passes|revised) followed by whitespace/punctuation is the
    -- only valid shape anyway.
    SELECT p_output_text IS NOT NULL
       AND p_output_text ~* '^\s*REVIEW:\s*(passes|revised)';
$$;

COMMENT ON FUNCTION stewards.review_output_passes_gate(text) IS
'Batch L.1.1.14: returns true if the review-stage output text starts with the explicit verdict prefix REVIEW: passes or REVIEW: revised. Anything else (including the bacteriopolis "where''s the draft" message) fails the gate.';


-- ---------------------------------------------------------------------
-- 2. BEFORE UPDATE trigger — gate maturity transitions to verified.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.trigger_review_prefix_verify_gate()
RETURNS trigger LANGUAGE plpgsql AS $FN$
DECLARE
    v_review_stages constant text[] := ARRAY['review','review_plan','revise','validate'];
    v_completing    text;
    v_stage_output  text;
    v_passes        boolean;
BEGIN
    -- Only act on transitions INTO verified.
    IF NEW.maturity IS DISTINCT FROM OLD.maturity AND NEW.maturity = 'verified' THEN
        v_completing := COALESCE(NEW.current_stage, OLD.current_stage);

        -- Skip the gate for stages we don't know how to validate.
        IF v_completing IS NULL OR NOT (v_completing = ANY(v_review_stages)) THEN
            RETURN NEW;
        END IF;

        v_stage_output := NEW.stage_results -> v_completing ->> 'output';

        v_passes := stewards.review_output_passes_gate(v_stage_output);

        IF NOT v_passes THEN
            RAISE NOTICE 'review verify gate FAILED: work_item=% stage=% output_head=%',
                NEW.id, v_completing,
                substring(COALESCE(v_stage_output, '(null)') FROM 1 FOR 80);

            NEW.maturity         := OLD.maturity;
            NEW.quarantine_reason := COALESCE(
                NEW.quarantine_reason,
                'verify gate (L.1.1.14): review-stage output did not start with REVIEW: passes or REVIEW: revised. ' ||
                'Output head: ' || substring(COALESCE(v_stage_output, '(null)') FROM 1 FOR 200)
            );
        END IF;
    END IF;

    RETURN NEW;
END;
$FN$;

DROP TRIGGER IF EXISTS work_items_review_verify_gate ON stewards.work_items;

CREATE TRIGGER work_items_review_verify_gate
BEFORE UPDATE OF maturity ON stewards.work_items
FOR EACH ROW
EXECUTE FUNCTION stewards.trigger_review_prefix_verify_gate();

COMMENT ON FUNCTION stewards.trigger_review_prefix_verify_gate() IS
'Batch L.1.1.14: BEFORE UPDATE trigger. When maturity is being set to verified on a review-style stage (review, review_plan, revise, validate), the stage_results[stage].output must start with REVIEW: passes or REVIEW: revised. Otherwise maturity stays at OLD value and quarantine_reason captures why.';


-- =====================================================================
-- End of l28-review-prefix-verify-gate.sql
-- =====================================================================
