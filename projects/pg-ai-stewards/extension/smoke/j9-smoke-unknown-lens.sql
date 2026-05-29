-- Smoke 1: unknown lens raises (no LLM cost)
DO $$
BEGIN
    PERFORM stewards.start_brainstorm(
        p_binding_question := 'will-not-run',
        p_destination      := '/tmp/nope.md',
        p_lenses           := ARRAY['scamper', 'typo-mispelled-lens']
    );
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'CAUGHT (expected): %', SQLERRM;
END $$;
