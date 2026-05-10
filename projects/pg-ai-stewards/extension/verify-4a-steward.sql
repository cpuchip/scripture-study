-- =====================================================================
-- Phase 4a-steward smoke test
-- =====================================================================

\echo '=== A. Schema presence ==='
SELECT 'steward_actions'      AS object, count(*) AS rows FROM stewards.steward_actions
UNION ALL SELECT 'retry_guidance_text', count(*) FROM stewards.retry_guidance_text
UNION ALL SELECT 'pipeline_breakers',   count(*) FROM stewards.pipeline_breakers;

\echo ''
\echo '=== B. work_items new columns ==='
SELECT column_name, data_type
  FROM information_schema.columns
 WHERE table_schema='stewards' AND table_name='work_items'
   AND column_name IN ('failure_count','last_failure_reason','last_failure_diagnosis',
                       'quarantined_at','quarantine_reason')
 ORDER BY column_name;

\echo ''
\echo '=== C. diagnose_failure: timeout patterns ==='
\echo '    Expected: timeout for all'
SELECT
    stewards.diagnose_failure('connection timeout', 0)              AS r1,
    stewards.diagnose_failure('context deadline exceeded', 0)       AS r2,
    stewards.diagnose_failure('Request timed out after 30s', 0)     AS r3,
    stewards.diagnose_failure('inactivity exceeded threshold', 0)   AS r4;

\echo ''
\echo '=== D. diagnose_failure: transient patterns ==='
\echo '    Expected: transient for all'
SELECT
    stewards.diagnose_failure('429 too many requests', 0)              AS r1,
    stewards.diagnose_failure('rate limit hit', 0)                     AS r2,
    stewards.diagnose_failure('upstream returned 503', 0)              AS r3,
    stewards.diagnose_failure('connection refused by provider', 0)     AS r4,
    stewards.diagnose_failure('Service Unavailable', 0)                AS r5;

\echo ''
\echo '=== E. diagnose_failure: tool_error patterns ==='
\echo '    Expected: tool_error for all'
SELECT
    stewards.diagnose_failure('tool not found: gospel_search', 0)        AS r1,
    stewards.diagnose_failure('schema validation failed', 0)             AS r2,
    stewards.diagnose_failure('function arguments missing', 0)           AS r3;

\echo ''
\echo '=== F. diagnose_failure: model_limit fallback ==='
\echo '    Expected: model_limit when failure_count >= 2 and no pattern match'
SELECT
    stewards.diagnose_failure('weird unparseable error', 2)  AS r1_model_limit,
    stewards.diagnose_failure('weird unparseable error', 1)  AS r2_unknown,
    stewards.diagnose_failure(NULL, 3)                        AS r3_model_limit_no_reason,
    stewards.diagnose_failure('', 0)                          AS r4_unknown_empty;

\echo ''
\echo '=== G. retry_guidance: each diagnosis returns text with attempt substituted ==='
SELECT diagnosis, length(stewards.retry_guidance(diagnosis, 3)) AS chars,
       substring(stewards.retry_guidance(diagnosis, 3) from 1 for 80) AS preview
  FROM stewards.retry_guidance_text
 ORDER BY diagnosis;

\echo ''
\echo '=== H. retry_guidance: attempt substitution ==='
\echo '    Expected: text contains "(attempt 7)"'
SELECT stewards.retry_guidance('timeout', 7) ~ '\(attempt 7\)' AS substitution_works;

\echo ''
\echo '=== I. retry_guidance: unknown diagnosis returns NULL ==='
SELECT stewards.retry_guidance('not_a_real_diagnosis', 1) IS NULL AS returns_null;

\echo ''
\echo '=== J. breaker_check on fresh state: lazy-creates closed breaker ==='
SELECT stewards.breaker_check('test_pipeline','test_stage') AS allowed_first_call;
SELECT state, failure_count FROM stewards.pipeline_breakers
 WHERE pipeline_family='test_pipeline' AND stage_name='test_stage';

\echo ''
\echo '=== K. breaker_record_failure x4: stays closed (threshold=5) ==='
SELECT stewards.breaker_record_failure('test_pipeline','test_stage');
SELECT stewards.breaker_record_failure('test_pipeline','test_stage');
SELECT stewards.breaker_record_failure('test_pipeline','test_stage');
SELECT stewards.breaker_record_failure('test_pipeline','test_stage');
SELECT state, failure_count FROM stewards.pipeline_breakers
 WHERE pipeline_family='test_pipeline' AND stage_name='test_stage';

\echo ''
\echo '=== L. 5th failure trips breaker open ==='
SELECT stewards.breaker_record_failure('test_pipeline','test_stage');
SELECT state, failure_count, opened_at IS NOT NULL AS has_opened_at
  FROM stewards.pipeline_breakers
 WHERE pipeline_family='test_pipeline' AND stage_name='test_stage';

\echo ''
\echo '=== M. breaker_check returns false while open + within cooldown ==='
SELECT stewards.breaker_check('test_pipeline','test_stage') AS allowed_during_cooldown;

\echo ''
\echo '=== N. breaker_record_success closes the breaker ==='
SELECT stewards.breaker_record_success('test_pipeline','test_stage');
SELECT state, failure_count FROM stewards.pipeline_breakers
 WHERE pipeline_family='test_pipeline' AND stage_name='test_stage';

\echo ''
\echo '=== O. steward_tick on empty queue returns 0 ==='
SELECT stewards.steward_tick() AS actions_taken;

\echo ''
\echo '=== P. work_items_steward_status view exists and queryable ==='
SELECT column_name FROM information_schema.columns
 WHERE table_schema='stewards' AND table_name='work_items_steward_status'
 ORDER BY ordinal_position;

\echo ''
\echo '=== Q. Cleanup test breaker ==='
DELETE FROM stewards.pipeline_breakers
 WHERE pipeline_family='test_pipeline' AND stage_name='test_stage';

\echo ''
\echo '=== Phase 4a-steward smoke test complete ==='
