-- =====================================================================
-- Phase 4a smoke test — runs after CREATE EXTENSION on an ephemeral container
-- to verify both 4a-cost-tracking.sql and 4a-escalation-chain.sql applied
-- cleanly. Plain SELECTs; no DDL. Output is human-readable.
-- =====================================================================

\echo '=== A. Schema presence ==='
SELECT 'model_pricing'         AS object, count(*) AS rows FROM stewards.model_pricing
UNION ALL SELECT 'cost_events',           count(*) FROM stewards.cost_events
UNION ALL SELECT 'cost_buckets',          count(*) FROM stewards.cost_buckets
UNION ALL SELECT 'stage_models',          count(*) FROM stewards.stage_models
UNION ALL SELECT 'model_escalation',      count(*) FROM stewards.model_escalation;

\echo ''
\echo '=== B. work_items new columns ==='
SELECT column_name, data_type, is_nullable
  FROM information_schema.columns
 WHERE table_schema='stewards' AND table_name='work_items'
   AND column_name IN ('cost_micro_dollars','cost_cap_micro','cost_capped_at',
                       'model_override','escalation_state','escalation_claimed_by',
                       'escalation_claimed_at','escalation_completed_at','escalation_attempts')
 ORDER BY column_name;

\echo ''
\echo '=== C. Pricing seed (Chinese models) ==='
SELECT model, input_micro_per_mtok, output_micro_per_mtok,
       cache_write_micro_per_mtok, cache_read_micro_per_mtok
  FROM stewards.model_pricing
 WHERE provider='opencode-zen' AND model IN ('kimi-k2.6','glm-5.1','minimax-m2.7','qwen3.6-plus')
 ORDER BY input_micro_per_mtok;

\echo ''
\echo '=== D. Pricing seed (Anthropic via OpenCode Zen) ==='
SELECT model, input_micro_per_mtok, output_micro_per_mtok,
       cache_write_micro_per_mtok, cache_read_micro_per_mtok
  FROM stewards.model_pricing
 WHERE provider='opencode-zen' AND model LIKE 'claude-%'
 ORDER BY input_micro_per_mtok;

\echo ''
\echo '=== E. Bucket seed + caps ==='
SELECT bucket_kind, period_start::date AS p_start, period_end::date AS p_end,
       bucket_limit_micro, notes
  FROM stewards.cost_buckets
 WHERE provider='opencode-zen'
 ORDER BY array_position(ARRAY['session_5h','daily','weekly','monthly']::text[], bucket_kind);

\echo ''
\echo '=== F. compute_cost: Kimi 1M input + 500K output ==='
\echo '    Expected micro_dollars = 950000 + 2000000 = 2950000 ($2.95)'
SELECT * FROM stewards.compute_cost('opencode-zen','kimi-k2.6', 1000000, 500000);

\echo ''
\echo '=== G. compute_cost: cache-aware MiniMax (input+cache_write+cache_read) ==='
\echo '    Expected: 100000 input * 300000/1M = 30000'
\echo '            + 50000 cache_write * 375000/1M = 18750'
\echo '            + 200000 cache_read * 60000/1M = 12000'
\echo '            + 50000 output * 1200000/1M = 60000'
\echo '            = 120750 micro_dollars ($0.121)'
SELECT * FROM stewards.compute_cost('opencode-zen','minimax-m2.7',
    100000, 50000, 50000, 200000);

\echo ''
\echo '=== H. compute_cost: provider/model that does not exist ==='
\echo '    Expected micro_dollars=0, pricing_effective_at=-infinity'
SELECT * FROM stewards.compute_cost('nonexistent','xyz-1.0', 1000, 1000);

\echo ''
\echo '=== I. pick_model: study/research, attempt 1 ==='
\echo '    Expected: kimi-k2.6 (stage default)'
SELECT stewards.pick_model('study','research', 1, 'initial') AS model;

\echo ''
\echo '=== J. pick_model: study/research, attempt 2 model_limit ==='
\echo '    Expected: glm-5.1 (Kimi escalates immediately on model_limit)'
SELECT stewards.pick_model('study','research', 2, 'model_limit') AS model;

\echo ''
\echo '=== K. pick_model: study/research, attempt 3 model_limit ==='
\echo '    Expected: __queue_for_opus__ (GLM escalates to queue on its attempt 2)'
SELECT stewards.pick_model('study','research', 3, 'model_limit') AS model;

\echo ''
\echo '=== L. pick_model: study/research, attempt 5 transient ==='
\echo '    Expected: kimi-k2.6 (transient stays on same model regardless of attempts)'
SELECT stewards.pick_model('study','research', 5, 'transient') AS model;

\echo ''
\echo '=== M. pick_model: dev/plan starts at top of chain ==='
\echo '    Expected: glm-5.1 (dev/plan stage default)'
SELECT stewards.pick_model('dev','plan', 1, 'initial') AS model;

\echo ''
\echo '=== N. pick_model: nonexistent pipeline raises ==='
\echo '    Expected: ERROR'
\set ON_ERROR_STOP off
SELECT stewards.pick_model('nonexistent','x', 1, 'initial');
\set ON_ERROR_STOP on

\echo ''
\echo '=== O. Escalation matrix coverage check ==='
\echo '    Every (current_model, diagnosis) for our 4 chain models should have a row'
SELECT
    array_agg(DISTINCT current_model) FILTER (WHERE next_model IS NOT NULL) AS escalating_models,
    count(*) FILTER (WHERE next_model = '__queue_for_opus__') AS queue_sentinels,
    count(*) FILTER (WHERE next_model IS NULL) AS stay_on_current,
    count(*) AS total_rules
  FROM stewards.model_escalation;

\echo ''
\echo '=== P. Stage-models coverage ==='
SELECT pipeline_family, count(*) AS stages, array_agg(stage_name ORDER BY stage_name) AS stage_list
  FROM stewards.stage_models
 GROUP BY pipeline_family
 ORDER BY pipeline_family;

\echo ''
\echo '=== Q. bucket_period_for sanity ==='
SELECT 'daily', stewards.bucket_period_for('daily', '2026-05-10 14:30:00+00'::timestamptz)
UNION ALL
SELECT 'weekly', stewards.bucket_period_for('weekly', '2026-05-10 14:30:00+00'::timestamptz)
UNION ALL
SELECT 'monthly', stewards.bucket_period_for('monthly', '2026-05-10 14:30:00+00'::timestamptz)
UNION ALL
SELECT 'session_5h', stewards.bucket_period_for('session_5h', '2026-05-10 14:30:00+00'::timestamptz);

\echo ''
\echo '=== Phase 4a smoke test complete ==='
