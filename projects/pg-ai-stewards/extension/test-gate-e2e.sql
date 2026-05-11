-- =====================================================================
-- End-to-end test: gate auto-fire path
--
-- 1. Create synthetic work_item at study-write/outline with mock output
-- 2. Call evaluate_gate → enqueues chat with _gate_eval=true
-- 3. Bgworker dispatches qwen3.6-plus
-- 4. On chat completion bgworker auto-fires parse_gate_response
-- 5. apply_gate_decision routes to advance/revise/surface
-- 6. Check maturity transition + gate_decisions audit row
-- =====================================================================

\set wi_slug 'gate-test-e2e-1'

-- Clean any prior runs
DELETE FROM stewards.work_items WHERE slug = :'wi_slug';

INSERT INTO stewards.work_items (
    id, slug, pipeline_family, current_stage, status,
    input, stage_results, maturity, revision_count, actor
)
VALUES (
    gen_random_uuid(),
    :'wi_slug',
    'study-write',
    'outline',
    'in_progress',
    '{"topic":"D&C 130:18-19 — intelligence in the resurrection"}'::jsonb,
    jsonb_build_object(
        'outline', jsonb_build_object(
            'output', E'I. Opening hook — what does it mean for intelligence to "rise with us"?\n  - Brigham Young: knowledge is the only thing we take with us\n  - Joseph Smith Lectures on Faith 5: God''s perfections rest on intelligence\n\nII. Hebrew/Greek word work\n  - Hebrew: binah (insight, discernment) vs daat (knowledge by acquaintance)\n  - Greek: gnosis vs sophia\n\nIII. Connection to the resurrection itself\n  - 1 Cor 15:42-44 — sown in corruption, raised in incorruption\n  - Alma 11:43 — same body restored to perfect frame\n  - The "us" that rises includes accumulated intelligence\n\nIV. Pastoral application\n  - Why study now: not for credit, but for stature\n  - Daily knowledge-getting as resurrection prep\n  - The covenant connection: D&C 88:118 → seek learning by study and faith\n\nV. Closing'
        )
    ),
    'raw',
    0,
    'gate-test'
)
RETURNING id, slug, current_stage, maturity, revision_count;

-- Show maturity ladder mapping for this stage
SELECT pipeline_family, stage_name, produces_maturity
  FROM stewards.pipeline_stage_maturity
 WHERE pipeline_family = 'study-write' AND stage_name = 'outline';

-- Now call evaluate_gate. Capture work_id (work_queue row).
SELECT stewards.evaluate_gate(id) AS gate_work_id
  FROM stewards.work_items WHERE slug = :'wi_slug';
