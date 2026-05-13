-- Real end-to-end smoke: COMMITs an exhibit agent-proposal so we can
-- verify materialize-writes lands the file on disk. Cleanup at the end
-- removes all artifacts.

\set ON_ERROR_STOP on
\echo '=== real e2e: exhibit through full materialize ==='

DO $$
DECLARE
    v_wi_id uuid;
    v_out jsonb;
BEGIN
    v_out := jsonb_build_object(
        'source_type', 'exhibit',
        'slug', 'i4-real-e2e-smoke-marker',
        'title', 'i4 real-e2e smoke marker (will be deleted)',
        'body', E'# i4 Real-E2E Smoke\n\nThis file exists only to verify the agent-proposal pipeline materializes content to disk. It will be deleted at the end of the smoke run.\n\nIf you see this file in git status, the smoke cleanup failed.',
        'frontmatter', jsonb_build_object('smoke', true),
        'project_association', NULL,
        'rationale', 'Verifies i4 agent-proposal pipeline end-to-end: studies INSERT + file enqueue + materialize-writes lands file on disk.'
    );

    INSERT INTO stewards.work_items
        (slug, pipeline_family, status, maturity, current_stage,
         input, stage_results, actor, origin, intent_id)
    VALUES
        ('i4-real-e2e-smoke-' || substring(gen_random_uuid()::text from 1 for 8),
         'agent-proposal', 'completed', 'raw', 'validate',
         jsonb_build_object('draft', v_out),
         jsonb_build_object('validate', jsonb_build_object('output', v_out::text)),
         'agent', 'agent_proposal',
         (SELECT id FROM stewards.intents WHERE slug = 'scripture-study'))
    RETURNING id INTO v_wi_id;

    UPDATE stewards.work_items SET maturity = 'verified' WHERE id = v_wi_id;
    RAISE NOTICE 'real e2e: work_item % advanced', v_wi_id;
END $$;

\echo '--- pre-materialize pending_file_writes ---'
SELECT id, target_path, source_id, length(content) AS content_len, materialized_at
  FROM stewards.pending_file_writes
 WHERE target_path = 'exhibits/i4-real-e2e-smoke-marker.md';
