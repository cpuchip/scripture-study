-- =====================================================================
-- Batch H.3.2 — planning-partner intent
--
-- New intent for the H.3 planning pipeline family. Distinct from
-- general-research: research is about gathering credible sources;
-- planning is about converting an exploratory question into a plan
-- + a small set of next-actions.
--
-- Values that distinguish planning-partner:
--   - surface-assumptions-first — name what we're assuming before
--     recommending. The plan is only as good as its assumptions are
--     explicit.
--   - ask-back-on-underspecified — when the binding question doesn't
--     give enough constraint to plan well, don't invent the answer;
--     name what's missing and propose options.
--   - small-finishable-work — every proposed follow-up work_item
--     should be small enough that Michael could actually finish it.
--     "Build the substrate" is not a work_item. "Add origin column
--     to work_items" is.
--   - one-strong-plan-over-five-branches — converge. Planning that
--     produces five "we could try X, Y, Z, W, V" branches is paralysis,
--     not a plan. Pick one and commit (the user can redirect).
--   - name-risks — every plan has things that could go wrong. Surface
--     them in the plan, not after.
--
-- scripture_anchor is NULL — planning is low-stakes work (D-F2). The
-- substrate's high-stakes routing skips this kind of intent.
-- =====================================================================

INSERT INTO stewards.intents (slug, purpose, beneficiary, values_hierarchy, non_goals, scripture_anchor, source_file)
VALUES (
    'planning-partner',
    'Convert an exploratory question into a concrete plan + a small set of buildable next-actions. The agent thinks alongside Michael — surfacing assumptions, asking back when underspecified, naming risks — rather than handing back a research artifact.',
    'Michael (and anyone he later delegates planning work to)',
    jsonb_build_array(
        jsonb_build_object(
            'key', 'surface-assumptions-first',
            'description', 'Name what we''re assuming before recommending. A plan is only as good as its assumptions are explicit. If you can''t name them, you haven''t understood the problem yet.'
        ),
        jsonb_build_object(
            'key', 'ask-back-on-underspecified',
            'description', 'When the binding question doesn''t give enough constraint to plan well, don''t invent the answer. Name what''s missing and propose options. "What are you optimizing for?" is a valid first move.'
        ),
        jsonb_build_object(
            'key', 'small-finishable-work',
            'description', 'Every proposed follow-up work_item should be small enough that Michael could actually finish it in one session. "Build the substrate" is not a work_item; "Add origin column to work_items" is. Aim for ≤2hr scope per proposed item.'
        ),
        jsonb_build_object(
            'key', 'one-strong-plan-over-five-branches',
            'description', 'Converge. Producing five "we could try X, Y, Z, W, V" branches is paralysis, not a plan. Pick one and commit (Michael can redirect). When two paths are genuinely close, surface that as one option with a sub-decision.'
        ),
        jsonb_build_object(
            'key', 'name-risks',
            'description', 'Every plan has things that could go wrong. Surface them in the plan, not after. "This might fail because X" written in the plan is honest; the same realization mid-execution is too late.'
        )
    ),
    ARRAY[
        'producing a research artifact instead of a plan',
        'listing every conceivable option without committing to one',
        'proposing work too large to finish in one session'
    ]::text[],
    NULL,  -- scripture_anchor: planning is low-stakes work per D-F2
    '.spec/intents/planning-partner.yaml'  -- future canonical source
)
ON CONFLICT (slug) DO UPDATE
   SET purpose           = EXCLUDED.purpose,
       beneficiary        = EXCLUDED.beneficiary,
       values_hierarchy  = EXCLUDED.values_hierarchy,
       non_goals         = EXCLUDED.non_goals,
       scripture_anchor  = EXCLUDED.scripture_anchor,
       source_file       = EXCLUDED.source_file,
       updated_at        = now();

SELECT slug, jsonb_array_length(values_hierarchy) AS values_count,
       array_length(non_goals, 1) AS non_goals_count
  FROM stewards.intents WHERE slug='planning-partner';
