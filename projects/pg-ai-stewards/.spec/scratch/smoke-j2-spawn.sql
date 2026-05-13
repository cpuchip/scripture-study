-- Smoke A: hand-craft a fanout parent and trigger spawn.

INSERT INTO stewards.work_items (
    pipeline_family, current_stage, slug, input, intent_id, actor,
    stage_results, maturity, status
) VALUES (
    'decompose-fanout',
    'decompose',
    'smoke-j2-fanout',
    '{"binding_question": "Smoke test the J.2 fan-out machinery"}'::jsonb,
    (SELECT id FROM stewards.intents WHERE slug = 'scripture-study'),
    'michael',
    jsonb_build_object(
        'context_gather', jsonb_build_object('output', 'smoke gather output'),
        'decompose', jsonb_build_object('output',
            '{"rationale":"smoke test of fan-out with 2 children",'
            '"children":['
              '{"slug":"smoke-j2-child-1","binding_question":"Echo hello from child 1","pipeline_family":"echo-test"},'
              '{"slug":"smoke-j2-child-2","binding_question":"Echo hello from child 2","pipeline_family":"echo-test"}'
            '],'
            '"aggregate":{"destination":"projects/pg-ai-stewards/.spec/scratch/smoke-j2-index.md","synthesis":false}'
            '}'
        )
    ),
    'planned',
    'completed'
)
RETURNING id, slug, pipeline_family, current_stage, status, maturity;

-- Now trigger the spawn by flipping maturity to verified.
UPDATE stewards.work_items
   SET maturity = 'verified'
 WHERE slug = 'smoke-j2-fanout'
RETURNING id, slug, maturity;

-- Inspect what got spawned.
SELECT id, slug, pipeline_family, current_stage, status, maturity, parent_work_item_id IS NOT NULL AS is_child
  FROM stewards.work_items
 WHERE slug LIKE 'smoke-j2-%'
 ORDER BY created_at;
