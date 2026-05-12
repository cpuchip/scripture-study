-- =====================================================================
-- Smoke H.2 — context_gather stage runs and briefs the gather stage
--
-- A different binding question from H.1.7's so we test the substrate
-- generalization, not the same self-reflection topic. Picking a science-
-- center-adjacent research question that should benefit from the
-- physics-news work_item already in the substrate.
-- =====================================================================

SELECT stewards.work_item_create(
    'research-write',
    jsonb_build_object(
        'binding_question',
            'What are practical mid-2026 examples of interactive science museum exhibits '
         || 'that translate cutting-edge physics or astronomy results into tangible visitor '
         || 'experiences? Looking for buildable patterns: what makes the exhibit work, what '
         || 'physical interaction is at the core, and what does the visitor walk away knowing? '
         || 'Build on whatever we''ve already gathered.'
    ),
    'h2-science-museum-exhibit-patterns',  -- slug
    'human',
    NULL,                                    -- token_budget
    (SELECT id FROM stewards.intents WHERE slug='general-research')
);

-- Dispatch.
SELECT stewards.work_item_dispatch_stage(
    (SELECT id FROM stewards.work_items WHERE slug='h2-science-museum-exhibit-patterns')
);

-- Confirm current_stage is context_gather (new first stage from H.2).
SELECT slug, current_stage, status FROM stewards.work_items
 WHERE slug='h2-science-museum-exhibit-patterns';
