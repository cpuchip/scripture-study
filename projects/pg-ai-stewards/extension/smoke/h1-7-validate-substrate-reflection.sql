-- =====================================================================
-- Smoke H.1.7 — substrate-reflective AI-tooling validation
--
-- Per D-H1.7-A ratification: dispatch the binding question:
--   "Looking over this last week's roundup, what other improvements
--    to our tooling can we make that industry is moving towards, that
--    would be interesting for us to adopt? Self-reflect on the DB +
--    agent substrate we've made and how what's going down in AI now
--    could help us build that better. Consult our prior journals,
--    proposals, mind files, and the 11-cycle guide."
--
-- Expected behavior:
--   - Gather stage issues fs_search / fs_read / study_search /
--     work_item_show calls in the first 2-4 rounds before any
--     external web_search
--   - Brief cites at least one prior journal or proposal by name
--   - Pipeline reaches verified, auto-mat fires
-- =====================================================================

INSERT INTO stewards.work_items (
    slug,
    pipeline_family,
    current_stage,
    cost_cap_micro,
    input,
    actor,
    intent_id
)
VALUES (
    'h1-7-validation-substrate-reflection-2',
    'research-write',
    'gather',
    600000,  -- $0.60 cap — substrate-reflection allows more rounds + tools than vanilla
    jsonb_build_object(
        'binding_question',
            'Looking over last week''s AI tooling roundup (research/ai-tools-weekly-2026-05-11.md), '
         || 'what other improvements to our tooling can we make that industry is moving towards, '
         || 'that would be interesting for us to adopt? Self-reflect on the DB + agent substrate '
         || 'we''ve built (pg-ai-stewards) and how current industry direction in AI infrastructure '
         || 'could help us build that better. Consult our prior journals (.spec/journal/), proposals '
         || '(.spec/proposals/), mind files (.mind/), and the 11-cycle guide '
         || '(.spec/proposals/pg-ai-stewards-11-cycle-review.md) before external search. Specifically: '
         || 'what compounding-knowledge / agent-memory / RAG / context-gathering / multi-agent patterns '
         || 'are emerging in 2026-05 that we should consider for the substrate?'
    ),
    'human',
    (SELECT id FROM stewards.intents WHERE slug='general-research')
)
ON CONFLICT (slug) DO NOTHING
RETURNING id, slug;

-- Dispatch. work_item_dispatch_stage reads current_stage from the
-- work_item; second arg is p_user_input (NULL = render via template).
SELECT stewards.work_item_dispatch_stage(
    (SELECT id FROM stewards.work_items WHERE slug='h1-7-validation-substrate-reflection-2')
);
