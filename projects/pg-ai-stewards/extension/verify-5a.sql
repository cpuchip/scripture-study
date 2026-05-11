-- Phase 5a smoke test — apply_gate_decision + render_template
-- (no LLM cost; tests SQL paths only)

\echo '=== A. render_template substitutes {{key}} → value ==='
SELECT stewards.render_template(
    'Hello {{name}}, you are {{role}}!',
    '{"name":"steward","role":"gate"}'::jsonb
) AS rendered;

\echo ''
\echo '=== B. Setup synthetic work_item with simulated stage output ==='
DO $x$
DECLARE
    v_wi_id uuid;
BEGIN
    v_wi_id := stewards.work_item_create(
        'study-write',
        jsonb_build_object('binding_question', 'phase 5a synthetic test'),
        'phase5a-' || (extract(epoch FROM now())::int)::text,
        'verify-5a'
    );
    UPDATE stewards.work_items
       SET status        = 'completed',
           stage_results = jsonb_build_object(
               'outline', jsonb_build_object(
                   'output', 'Mock outline text for the gate to evaluate.',
                   'agent', 'plan'))
     WHERE id = v_wi_id;
    PERFORM set_config('verify5a.wi_id', v_wi_id::text, false);
    RAISE NOTICE 'work_item % at maturity=raw, current_stage=outline', v_wi_id;
END;
$x$;

\echo ''
\echo '=== C. apply_gate_decision: ADVANCE path ==='
SELECT stewards.apply_gate_decision(
    current_setting('verify5a.wi_id')::uuid,
    '{"action":"advance","reasoning":"outline looks solid","feedback":""}'::jsonb
) AS new_maturity;

\echo ''
\echo '=== D. State after advance ==='
SELECT id::text, maturity, revision_count, status
  FROM stewards.work_items
 WHERE id = current_setting('verify5a.wi_id')::uuid;

\echo '=== Audit row from advance ==='
SELECT from_maturity, action, reasoning, revision_count
  FROM stewards.gate_decisions
 WHERE work_item_id = current_setting('verify5a.wi_id')::uuid
 ORDER BY at DESC LIMIT 1;

\echo ''
\echo '=== E. apply_gate_decision: REVISE path (revision 1) ==='
SELECT stewards.apply_gate_decision(
    current_setting('verify5a.wi_id')::uuid,
    '{"action":"revise","reasoning":"missing structure","feedback":"add a clearer therefore/but chain"}'::jsonb
);

\echo '=== State after revise: status=failed (steward will pick up) ==='
SELECT id::text, maturity, revision_count, status, last_failure_diagnosis,
       substr(last_failure_reason, 1, 80) AS reason_preview
  FROM stewards.work_items
 WHERE id = current_setting('verify5a.wi_id')::uuid;

\echo ''
\echo '=== F. apply_gate_decision: REVISE again (revision 2) ==='
SELECT stewards.apply_gate_decision(
    current_setting('verify5a.wi_id')::uuid,
    '{"action":"revise","reasoning":"still rough","feedback":"keep going"}'::jsonb
);

\echo '=== State after revision 2: still status=failed, count=2 ==='
SELECT id::text, revision_count, status FROM stewards.work_items
 WHERE id = current_setting('verify5a.wi_id')::uuid;

\echo ''
\echo '=== G. apply_gate_decision: REVISE 3rd time → cap exceeded → auto-surface ==='
SELECT stewards.apply_gate_decision(
    current_setting('verify5a.wi_id')::uuid,
    '{"action":"revise","reasoning":"third pass","feedback":"giving up"}'::jsonb
);

\echo '=== State: status=awaiting_review, revision_count=3 ==='
SELECT id::text, revision_count, status FROM stewards.work_items
 WHERE id = current_setting('verify5a.wi_id')::uuid;

\echo ''
\echo '=== H. apply_gate_decision: SURFACE path (sanity) ==='
SELECT stewards.apply_gate_decision(
    current_setting('verify5a.wi_id')::uuid,
    '{"action":"surface","reasoning":"needs human","feedback":"binding question changed"}'::jsonb
);
SELECT id::text, status FROM stewards.work_items
 WHERE id = current_setting('verify5a.wi_id')::uuid;

\echo ''
\echo '=== I. Full gate_decisions audit trail for this work_item ==='
SELECT id, from_maturity, action, revision_count,
       substr(reasoning, 1, 50) AS reasoning_preview
  FROM stewards.gate_decisions
 WHERE work_item_id = current_setting('verify5a.wi_id')::uuid
 ORDER BY id ASC;

\echo ''
\echo '=== J. Cleanup ==='
DELETE FROM stewards.gate_decisions
 WHERE work_item_id = current_setting('verify5a.wi_id')::uuid;
DELETE FROM stewards.work_items
 WHERE id = current_setting('verify5a.wi_id')::uuid;

\echo ''
\echo '=== Phase 5a smoke test complete ==='
