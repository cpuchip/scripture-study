-- =====================================================================
-- Smoke H.3.6 — first real e2e run of the planning pipeline
--
-- Binding question chosen for:
--   - high prior-work signal (substrate has lots of journals + proposals
--     to surface in context_gather + explore)
--   - actionable outcome (proposed work_items become Michael's next-session
--     todos in pg-ai-stewards project)
--   - testable validation (we know what good output looks like)
--
-- Pause soak already done. Cost cap $0.75 per Q-H3.3.
-- =====================================================================

SELECT stewards.work_item_create(
    'planning',
    jsonb_build_object(
        'binding_question',
            'What is the right order to ship the next three pg-ai-stewards substrate items: '
         || '(a) Batch I — agent write-back rung on the trust ladder (agents propose '
         || 'studies/notes/lessons through gate machinery before persisting); '
         || '(b) the yaml.rs Rust parser refactor (rule of three triggered by scripture-study '
         || '+ general-research + planning-partner intents all needing parsing); '
         || '(c) the still-deferred Phase A pgrx BGW SPI longjmp catch + 60s periodic reaper '
         || 'tick? Plan the order and propose follow-up work_items for the first one to ship. '
         || 'Consult our prior journals (.spec/journal/) and active.md for context on what we '
         || 'just completed (H.1.7, H.2, H.3) and what bugs have surfaced this week.',
        'today', to_char(current_date, 'YYYY-MM-DD')
    ),
    'h3-6-substrate-next-three',
    'human',
    NULL,
    (SELECT id FROM stewards.intents WHERE slug='planning-partner')
);

-- Set cost cap + project_association
UPDATE stewards.work_items
   SET cost_cap_micro = 750000,
       project_association = 'pg-ai-stewards'
 WHERE slug = 'h3-6-substrate-next-three';

-- Dispatch.
SELECT stewards.work_item_dispatch_stage(
    (SELECT id FROM stewards.work_items WHERE slug='h3-6-substrate-next-three')
);

-- Confirm starting state.
SELECT slug, current_stage, status, project_association, cost_cap_micro
  FROM stewards.work_items WHERE slug='h3-6-substrate-next-three';
