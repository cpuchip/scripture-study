-- Smoke 1: with the bgworker stream_options fix, a Gemini dispatch should
-- now record a cost_event with tokens > 0 and micro_dollars > 0.
SELECT stewards.start_brainstorm(
    p_binding_question := 'SMOKE J.11: one sentence on why prepaid caps matter.',
    p_destination      := 'study/.scratch/smoke-j11-gemini.md',
    p_slug             := 'smoke-j11-cost',
    p_lenses           := ARRAY['scamper'],
    p_models           := '{"scamper":{"model":"gemini-2.5-flash-lite","provider":"google_gemini"}}'::jsonb,
    p_cost_cap_per_lens_micro := 100000
) AS parent_id;
