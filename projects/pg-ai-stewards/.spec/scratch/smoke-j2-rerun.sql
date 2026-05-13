-- Clean up prior smoke, then rerun the fan-out chain end-to-end.

-- 1. Clean up.
DELETE FROM stewards.work_items WHERE slug LIKE 'smoke-j2-%' OR slug LIKE 'smoke-j2b-%';

-- 2. Create fresh fanout parent.
INSERT INTO stewards.work_items (
    pipeline_family, current_stage, slug, input, intent_id, actor,
    stage_results, maturity, status
) VALUES (
    'decompose-fanout',
    'decompose',
    'smoke-j2b-fanout',
    '{"binding_question": "Smoke J2 rerun with file-destination fix"}'::jsonb,
    (SELECT id FROM stewards.intents WHERE slug = 'scripture-study'),
    'michael',
    jsonb_build_object(
        'context_gather', jsonb_build_object('output', 'smoke gather output'),
        'decompose', jsonb_build_object('output',
            '{"rationale":"rerun smoke with auto-verify + direct destination",'
            '"children":['
              '{"slug":"smoke-j2b-child-1","binding_question":"Echo hello again 1","pipeline_family":"echo-test"},'
              '{"slug":"smoke-j2b-child-2","binding_question":"Echo hello again 2","pipeline_family":"echo-test"}'
            '],'
            '"aggregate":{"destination":"projects/pg-ai-stewards/.spec/scratch/smoke-j2b-index.md","synthesis":false}'
            '}'
        )
    ),
    'planned',
    'completed'
)
RETURNING id, slug;

-- 3. Trigger spawn.
UPDATE stewards.work_items
   SET maturity = 'verified'
 WHERE slug = 'smoke-j2b-fanout';

-- 4. Inspect.
SELECT slug, pipeline_family, status, maturity, file_destination
  FROM stewards.work_items
 WHERE slug LIKE 'smoke-j2b-%'
 ORDER BY created_at;
