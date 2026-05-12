-- =====================================================================
-- Smoke H.3-followup #3 — first science-center planning question
--
-- Per Q4 ratification (with Michael's constraint corrections):
--   - $500 actual budget (not the $3K placeholder in the question option)
--   - 5 laptops repurposed from the Bridge Simulator project
--   - 1x 10" ESP32 panel (see waveshare-esp32-s3-specs.md in
--     projects/space-center/docs/)
--   - Existing business plans + research in projects/space-center/docs/
--     that "his last brain" developed but hasn't worked through yet
--
-- fs-read scope has been expanded (h3-followup-1) to include the
-- space-center docs + .spec, so context_gather can consult them.
--
-- Project association: 'space-center'. Cost cap: $0.75 per Q-H3.3.
-- =====================================================================

SELECT stewards.work_item_create(
    'planning',
    jsonb_build_object(
        'binding_question',
            'What is the minimum-viable AI-literacy exhibit we could build for the '
         || 'Marsfield science center in 8 weeks with one staffer, ~$500 in NEW '
         || 'materials, and the existing hardware Michael already has: 5 laptops '
         || 'repurposed from the Bridge Simulator project + one 10" ESP32 panel '
         || '(see projects/space-center/docs/waveshare-esp32-s3-specs.md)? '
         || 'CONSULT projects/space-center/docs/ for prior planning notes — '
         || 'diy-science-exhibits-research.md, marshfield-research.md, '
         || 'financial-model.md, exhibits/, opening-timeline.md, and any other '
         || 'docs that bear. The business plans there were developed by an '
         || 'earlier agent that Michael hasn''t worked through yet, so trust '
         || 'them as ratified context. Plan 3-5 follow-up work_items that get '
         || 'us from "scoped MVP" to "demoable exhibit." Be ruthless about '
         || 'the $500 budget — that''s 5x tighter than typical references.',
        'today', to_char(current_date, 'YYYY-MM-DD')
    ),
    'h3-followup-sc-ai-literacy-mvp',  -- slug
    'human',
    NULL,                                -- token_budget
    (SELECT id FROM stewards.intents WHERE slug='planning-partner')
);

-- Set cost cap + project_association. The render_file_destination helper
-- from h3-followup-2 will auto-render plans/<slug>.md at trigger time,
-- so we don't have to set file_destination manually anymore.
UPDATE stewards.work_items
   SET cost_cap_micro = 750000,
       project_association = 'space-center'
 WHERE slug = 'h3-followup-sc-ai-literacy-mvp';

-- Dispatch.
SELECT stewards.work_item_dispatch_stage(
    (SELECT id FROM stewards.work_items WHERE slug='h3-followup-sc-ai-literacy-mvp')
);

SELECT slug, current_stage, status, project_association, cost_cap_micro
  FROM stewards.work_items WHERE slug='h3-followup-sc-ai-literacy-mvp';
