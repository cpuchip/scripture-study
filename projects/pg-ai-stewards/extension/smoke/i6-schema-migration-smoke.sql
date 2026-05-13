-- =====================================================================
-- i6/I.3 smoke — schema-migration source_type end-to-end
--
-- Tests three paths:
--   1. schema-migration WITHOUT claude_attested → apply_agent_proposal
--      returns false; no studies row; file_destination not set
--   2. schema-migration WITH claude_attested AND valid SQL body →
--      apply_agent_proposal returns true; file_destination set to
--      projects/pg-ai-stewards/extension/<slug>.sql; pending_file_writes
--      row queued. (materialize-writes --dry-run will then exercise
--      validate-sql on the queued content.)
--   3. schema-migration WITH claude_attested AND BAD SQL body →
--      apply_agent_proposal still succeeds (it doesn't validate SQL
--      itself); pending_file_writes queued; materialize-writes will
--      catch syntax error and mark error:syntax.
-- =====================================================================

\set ON_ERROR_STOP off

BEGIN;

\echo '=== smoke 1: schema-migration WITHOUT claude_attested ==='

DO $$
DECLARE
    v_wi_id uuid;
    v_out jsonb;
    v_result boolean;
BEGIN
    v_out := jsonb_build_object(
        'source_type', 'schema-migration',
        'slug', 'iz-smoke-test-without-attest',
        'title', 'Smoke without claude_attested - should reject',
        'body', E'-- iz smoke\nCREATE TABLE stewards.smoke_iz_test (id int);',
        'frontmatter', '{}'::jsonb,
        'project_association', 'pg-ai-stewards',
        'rationale', 'Should reject: claude_attested missing from input.draft.'
    );

    INSERT INTO stewards.work_items
        (slug, pipeline_family, status, maturity, current_stage,
         input, stage_results, actor, origin, intent_id)
    VALUES
        ('smoke-sm-no-attest-' || substring(gen_random_uuid()::text from 1 for 8),
         'agent-proposal', 'completed', 'raw', 'validate',
         jsonb_build_object('draft', v_out),  -- no claude_attested key
         jsonb_build_object('validate', jsonb_build_object('output', v_out::text)),
         'agent', 'agent_proposal',
         (SELECT id FROM stewards.intents WHERE slug = 'scripture-study'))
    RETURNING id INTO v_wi_id;

    -- Manually advance maturity (no trigger needed since we want to call apply directly)
    v_result := stewards.apply_agent_proposal(v_wi_id);
    RAISE NOTICE 'smoke 1: apply_agent_proposal returned %', v_result;
END $$;

\echo '--- state after smoke 1 (expect: applied=f, file_dest=NULL) ---'
SELECT slug, file_destination, agent_proposal_applied_at IS NOT NULL AS applied
  FROM stewards.work_items
 WHERE slug LIKE 'smoke-sm-no-attest-%'
 ORDER BY created_at DESC LIMIT 1;


\echo ''
\echo '=== smoke 2: schema-migration WITH claude_attested (valid SQL) ==='

DO $$
DECLARE
    v_wi_id uuid;
    v_out jsonb;
    v_input_draft jsonb;
    v_result boolean;
BEGIN
    v_out := jsonb_build_object(
        'source_type', 'schema-migration',
        'slug', 'iz-smoke-test-with-attest-valid',
        'title', 'Smoke with claude_attested + valid SQL',
        'body', E'-- iz smoke with attestation\nCREATE TABLE stewards.smoke_iz_valid (id int);',
        'frontmatter', '{}'::jsonb,
        'project_association', 'pg-ai-stewards',
        'rationale', 'Should accept: claude_attested=true on input.draft; SQL is valid.'
    );
    -- input.draft includes claude_attested=true
    v_input_draft := v_out || jsonb_build_object('claude_attested', true);

    INSERT INTO stewards.work_items
        (slug, pipeline_family, status, maturity, current_stage,
         input, stage_results, actor, origin, intent_id)
    VALUES
        ('smoke-sm-attest-' || substring(gen_random_uuid()::text from 1 for 8),
         'agent-proposal', 'completed', 'raw', 'validate',
         jsonb_build_object('draft', v_input_draft),  -- attested
         jsonb_build_object('validate', jsonb_build_object('output', v_out::text)),
         'agent', 'agent_proposal',
         (SELECT id FROM stewards.intents WHERE slug = 'scripture-study'))
    RETURNING id INTO v_wi_id;

    v_result := stewards.apply_agent_proposal(v_wi_id);
    RAISE NOTICE 'smoke 2: apply_agent_proposal returned %', v_result;

    -- Manually trigger enqueue (since we bypassed the maturity trigger)
    PERFORM stewards.enqueue_work_item_file(v_wi_id, 'smoke-i6');
END $$;

\echo '--- state after smoke 2 (expect: applied=t, file_dest set) ---'
SELECT slug, file_destination, agent_proposal_applied_at IS NOT NULL AS applied,
       file_enqueued_at IS NOT NULL AS enqueued
  FROM stewards.work_items
 WHERE slug LIKE 'smoke-sm-attest-%'
 ORDER BY created_at DESC LIMIT 1;

\echo '--- pending_file_writes row for smoke 2 (expect queued, not materialized) ---'
SELECT id, target_path, length(content) AS content_len,
       materialized_at, materialized_by
  FROM stewards.pending_file_writes
 WHERE target_path = 'projects/pg-ai-stewards/extension/iz-smoke-test-with-attest-valid.sql';


\echo ''
\echo '=== smoke 3: schema-migration WITH claude_attested but BAD SQL ==='

DO $$
DECLARE
    v_wi_id uuid;
    v_out jsonb;
    v_input_draft jsonb;
    v_result boolean;
BEGIN
    v_out := jsonb_build_object(
        'source_type', 'schema-migration',
        'slug', 'iz-smoke-test-bad-sql',
        'title', 'Smoke with claude_attested but bad SQL syntax',
        'body', E'-- iz smoke bad sql\nCREATE TABL stewards.broken (id;',  -- intentionally malformed
        'frontmatter', '{}'::jsonb,
        'project_association', 'pg-ai-stewards',
        'rationale', 'apply succeeds (does not validate SQL); materialize-writes catches via validate-sql.'
    );
    v_input_draft := v_out || jsonb_build_object('claude_attested', true);

    INSERT INTO stewards.work_items
        (slug, pipeline_family, status, maturity, current_stage,
         input, stage_results, actor, origin, intent_id)
    VALUES
        ('smoke-sm-bad-' || substring(gen_random_uuid()::text from 1 for 8),
         'agent-proposal', 'completed', 'raw', 'validate',
         jsonb_build_object('draft', v_input_draft),
         jsonb_build_object('validate', jsonb_build_object('output', v_out::text)),
         'agent', 'agent_proposal',
         (SELECT id FROM stewards.intents WHERE slug = 'scripture-study'))
    RETURNING id INTO v_wi_id;

    v_result := stewards.apply_agent_proposal(v_wi_id);
    RAISE NOTICE 'smoke 3: apply_agent_proposal returned % (apply does not check SQL)', v_result;

    PERFORM stewards.enqueue_work_item_file(v_wi_id, 'smoke-i6');
END $$;

\echo '--- pending_file_writes row for smoke 3 (expect queued; materialize-writes will catch) ---'
SELECT id, target_path, length(content) AS content_len, materialized_at, materialized_by
  FROM stewards.pending_file_writes
 WHERE target_path = 'projects/pg-ai-stewards/extension/iz-smoke-test-bad-sql.sql';

-- COMMIT so materialize-writes can see the rows in the next step
COMMIT;

\echo ''
\echo '=== smoke complete (committed for materialize-writes test) ==='
\echo '   Next: run materialize-writes --dry-run from bridge container'
\echo '   to exercise validate-sql hook. Then clean up.'
