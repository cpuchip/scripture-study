-- Smoke: start_brainstorm pre-flight refuses cleanly when a lens routes to
-- an over-cap provider, BEFORE creating the parent work_item.
UPDATE stewards.provider_spend_caps SET cap_micro = 1 WHERE provider = 'google_gemini';

DO $$
BEGIN
    PERFORM stewards.start_brainstorm(
        p_binding_question := 'preflight cap test',
        p_destination      := 'study/.scratch/pf.md',
        p_slug             := 'pf-smoke',
        p_lenses           := ARRAY['scamper'],
        p_models           := '{"scamper":{"model":"gemini-2.5-flash-lite","provider":"google_gemini"}}'::jsonb
    );
    RAISE NOTICE 'UNEXPECTED: pre-flight did not raise';
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'PREFLIGHT (expected): %', SQLERRM;
END $$;

SELECT count(*) AS pf_parents_created FROM stewards.work_items WHERE slug = 'pf-smoke';

-- restore
UPDATE stewards.provider_spend_caps SET cap_micro = 18000000 WHERE provider = 'google_gemini';
