-- =====================================================================
-- Batch J.9.b — 8 new brainstorm lens pipelines
-- =====================================================================
-- Single-stage pipelines pointing at the agents seeded in j9a. Per the
-- J.8 generalization:
--   * stages[0].model AND stages[0].provider = NULL (use fallback chain)
--   * metadata.default_model + default_provider = the lens's preferred
--     defaults (preserves the B1 "each lens declares its provider"
--     spirit while leaving caller override fully available)
--   * metadata.suggested_model = UI hint, identical to default for these
--
-- The j6 auto-verify trigger already matches `pipeline_family LIKE
-- 'brainstorm-%'` so these 8 inherit the one-shot completion behavior
-- without further trigger changes.
-- =====================================================================


-- Mind Mapping ---------------------------------------------------------
INSERT INTO stewards.pipelines (
    family, description, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath,
    maturity_ladder, auto_materialize_on_verified, metadata
)
VALUES (
    'brainstorm-mind-mapping',
    'Brainstorm lens: Mind Mapping. Hierarchical idea tree, 3-4 angular branches × 3-5 children, optional cross-branch links.',
    $STAGES$[
      {
        "name": "lens",
        "next": null,
        "model": null,
        "provider": null,
        "agent_family": "brainstorm-mind-mapping",
        "auto_advance": true,
        "tools_disabled": false,
        "input_template": "Binding question: {{input.binding_question}}\n\nProduce a mind map: 3-4 angular sub-themes, 3-5 children each. Mark cross-branch links inline. End your turn after the last leaf."
      }
    ]$STAGES$::jsonb,
    false, false, NULL, NULL,
    '["raw", "verified"]'::jsonb,
    false,
    jsonb_build_object(
        'shape', 'brainstorm-lens',
        'lens', 'mind-mapping',
        'default_model', 'qwen3.6-plus',
        'default_provider', 'opencode_go',
        'suggested_model', 'qwen3.6-plus',
        'suggested_provider', 'opencode_go'
    )
)
ON CONFLICT (family) DO UPDATE
   SET description                  = EXCLUDED.description,
       stages                       = EXCLUDED.stages,
       sabbath_enabled              = EXCLUDED.sabbath_enabled,
       atonement_enabled            = EXCLUDED.atonement_enabled,
       file_destination_template    = EXCLUDED.file_destination_template,
       file_content_jsonpath        = EXCLUDED.file_content_jsonpath,
       maturity_ladder              = EXCLUDED.maturity_ladder,
       auto_materialize_on_verified = EXCLUDED.auto_materialize_on_verified,
       metadata                     = EXCLUDED.metadata;


-- Brainwriting --------------------------------------------------------
INSERT INTO stewards.pipelines (
    family, description, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath,
    maturity_ladder, auto_materialize_on_verified, metadata
)
VALUES (
    'brainstorm-brainwriting',
    'Brainstorm lens: Brainwriting (6-3-5). 6 seed ideas, 3 builds per seed (extend / vary / counter). 24 items total.',
    $STAGES$[
      {
        "name": "lens",
        "next": null,
        "model": null,
        "provider": null,
        "agent_family": "brainstorm-brainwriting",
        "auto_advance": true,
        "tools_disabled": false,
        "input_template": "Binding question: {{input.binding_question}}\n\nProduce 6 seed ideas, then 3 builds (Extend / Vary / Counter) per seed. End your turn after seed 6's Counter build."
      }
    ]$STAGES$::jsonb,
    false, false, NULL, NULL,
    '["raw", "verified"]'::jsonb,
    false,
    jsonb_build_object(
        'shape', 'brainstorm-lens',
        'lens', 'brainwriting',
        'default_model', 'kimi-k2.6',
        'default_provider', 'opencode_go',
        'suggested_model', 'kimi-k2.6',
        'suggested_provider', 'opencode_go'
    )
)
ON CONFLICT (family) DO UPDATE
   SET description                  = EXCLUDED.description,
       stages                       = EXCLUDED.stages,
       sabbath_enabled              = EXCLUDED.sabbath_enabled,
       atonement_enabled            = EXCLUDED.atonement_enabled,
       file_destination_template    = EXCLUDED.file_destination_template,
       file_content_jsonpath        = EXCLUDED.file_content_jsonpath,
       maturity_ladder              = EXCLUDED.maturity_ladder,
       auto_materialize_on_verified = EXCLUDED.auto_materialize_on_verified,
       metadata                     = EXCLUDED.metadata;


-- Starbursting (5W1H) --------------------------------------------------
INSERT INTO stewards.pipelines (
    family, description, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath,
    maturity_ladder, auto_materialize_on_verified, metadata
)
VALUES (
    'brainstorm-starbursting',
    'Brainstorm lens: Starbursting (5W1H). Question-generation, not answer-generation. 4-6 questions per Who/What/When/Where/Why/How.',
    $STAGES$[
      {
        "name": "lens",
        "next": null,
        "model": null,
        "provider": null,
        "agent_family": "brainstorm-starbursting",
        "auto_advance": true,
        "tools_disabled": false,
        "input_template": "Binding question: {{input.binding_question}}\n\nProduce 4-6 specific actionable questions in each of the six categories. Do NOT answer them. End your turn after the last HOW question."
      }
    ]$STAGES$::jsonb,
    false, false, NULL, NULL,
    '["raw", "verified"]'::jsonb,
    false,
    jsonb_build_object(
        'shape', 'brainstorm-lens',
        'lens', 'starbursting',
        'default_model', 'kimi-k2.6',
        'default_provider', 'opencode_go',
        'suggested_model', 'kimi-k2.6',
        'suggested_provider', 'opencode_go'
    )
)
ON CONFLICT (family) DO UPDATE
   SET description                  = EXCLUDED.description,
       stages                       = EXCLUDED.stages,
       sabbath_enabled              = EXCLUDED.sabbath_enabled,
       atonement_enabled            = EXCLUDED.atonement_enabled,
       file_destination_template    = EXCLUDED.file_destination_template,
       file_content_jsonpath        = EXCLUDED.file_content_jsonpath,
       maturity_ladder              = EXCLUDED.maturity_ladder,
       auto_materialize_on_verified = EXCLUDED.auto_materialize_on_verified,
       metadata                     = EXCLUDED.metadata;


-- Disney Method -------------------------------------------------------
INSERT INTO stewards.pipelines (
    family, description, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath,
    maturity_ladder, auto_materialize_on_verified, metadata
)
VALUES (
    'brainstorm-disney',
    'Brainstorm lens: Disney Method. Three voices in sequence — Dreamer (no constraints), Realist (concrete execution), Critic (risks). Later voices reference earlier.',
    $STAGES$[
      {
        "name": "lens",
        "next": null,
        "model": null,
        "provider": null,
        "agent_family": "brainstorm-disney",
        "auto_advance": true,
        "tools_disabled": false,
        "input_template": "Binding question: {{input.binding_question}}\n\nApply Disney Method: Dreamer (5-7 visions) → Realist (execution per dream) → Critic (3-5 critiques with watch-out principles). End your turn after the last Critic bullet."
      }
    ]$STAGES$::jsonb,
    false, false, NULL, NULL,
    '["raw", "verified"]'::jsonb,
    false,
    jsonb_build_object(
        'shape', 'brainstorm-lens',
        'lens', 'disney',
        'default_model', 'kimi-k2.6',
        'default_provider', 'opencode_go',
        'suggested_model', 'kimi-k2.6',
        'suggested_provider', 'opencode_go'
    )
)
ON CONFLICT (family) DO UPDATE
   SET description                  = EXCLUDED.description,
       stages                       = EXCLUDED.stages,
       sabbath_enabled              = EXCLUDED.sabbath_enabled,
       atonement_enabled            = EXCLUDED.atonement_enabled,
       file_destination_template    = EXCLUDED.file_destination_template,
       file_content_jsonpath        = EXCLUDED.file_content_jsonpath,
       maturity_ladder              = EXCLUDED.maturity_ladder,
       auto_materialize_on_verified = EXCLUDED.auto_materialize_on_verified,
       metadata                     = EXCLUDED.metadata;


-- Storyboarding -------------------------------------------------------
INSERT INTO stewards.pipelines (
    family, description, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath,
    maturity_ladder, auto_materialize_on_verified, metadata
)
VALUES (
    'brainstorm-storyboarding',
    'Brainstorm lens: Storyboarding. 5-7 narrative scenes with a single protagonist; each scene seeds one design idea. Surfaces temporal / contextual angles flat lists miss.',
    $STAGES$[
      {
        "name": "lens",
        "next": null,
        "model": null,
        "provider": null,
        "agent_family": "brainstorm-storyboarding",
        "auto_advance": true,
        "tools_disabled": false,
        "input_template": "Binding question: {{input.binding_question}}\n\nWrite 5-7 scenes following one protagonist through baseline → complication → midpoint → resolution. Each scene ends with an Idea seed. End your turn after the final Idea."
      }
    ]$STAGES$::jsonb,
    false, false, NULL, NULL,
    '["raw", "verified"]'::jsonb,
    false,
    jsonb_build_object(
        'shape', 'brainstorm-lens',
        'lens', 'storyboarding',
        'default_model', 'qwen3.6-plus',
        'default_provider', 'opencode_go',
        'suggested_model', 'qwen3.6-plus',
        'suggested_provider', 'opencode_go'
    )
)
ON CONFLICT (family) DO UPDATE
   SET description                  = EXCLUDED.description,
       stages                       = EXCLUDED.stages,
       sabbath_enabled              = EXCLUDED.sabbath_enabled,
       atonement_enabled            = EXCLUDED.atonement_enabled,
       file_destination_template    = EXCLUDED.file_destination_template,
       file_content_jsonpath        = EXCLUDED.file_content_jsonpath,
       maturity_ladder              = EXCLUDED.maturity_ladder,
       auto_materialize_on_verified = EXCLUDED.auto_materialize_on_verified,
       metadata                     = EXCLUDED.metadata;


-- TRIZ ---------------------------------------------------------------
INSERT INTO stewards.pipelines (
    family, description, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath,
    maturity_ladder, auto_materialize_on_verified, metadata
)
VALUES (
    'brainstorm-triz',
    'Brainstorm lens: TRIZ. Identify contradictions in the binding question; map to 3-5 of TRIZ''s 40 inventive principles; concrete solution sketch per principle.',
    $STAGES$[
      {
        "name": "lens",
        "next": null,
        "model": null,
        "provider": null,
        "agent_family": "brainstorm-triz",
        "auto_advance": true,
        "tools_disabled": false,
        "input_template": "Binding question: {{input.binding_question}}\n\nApply TRIZ: name 2-3 contradictions, map each to 2-3 of the 40 inventive principles, write a concrete solution sketch per cited principle. End your turn after the last solution sketch."
      }
    ]$STAGES$::jsonb,
    false, false, NULL, NULL,
    '["raw", "verified"]'::jsonb,
    false,
    jsonb_build_object(
        'shape', 'brainstorm-lens',
        'lens', 'triz',
        'default_model', 'kimi-k2.6',
        'default_provider', 'opencode_go',
        'suggested_model', 'kimi-k2.6',
        'suggested_provider', 'opencode_go'
    )
)
ON CONFLICT (family) DO UPDATE
   SET description                  = EXCLUDED.description,
       stages                       = EXCLUDED.stages,
       sabbath_enabled              = EXCLUDED.sabbath_enabled,
       atonement_enabled            = EXCLUDED.atonement_enabled,
       file_destination_template    = EXCLUDED.file_destination_template,
       file_content_jsonpath        = EXCLUDED.file_content_jsonpath,
       maturity_ladder              = EXCLUDED.maturity_ladder,
       auto_materialize_on_verified = EXCLUDED.auto_materialize_on_verified,
       metadata                     = EXCLUDED.metadata;


-- Forced Analogy ------------------------------------------------------
INSERT INTO stewards.pipelines (
    family, description, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath,
    maturity_ladder, auto_materialize_on_verified, metadata
)
VALUES (
    'brainstorm-forced-analogy',
    'Brainstorm lens: Forced Analogy. 3 random unrelated domains × restate-generate-port. Plus one standout port that surfaces something the home domain''s clichés missed.',
    $STAGES$[
      {
        "name": "lens",
        "next": null,
        "model": null,
        "provider": null,
        "agent_family": "brainstorm-forced-analogy",
        "auto_advance": true,
        "tools_disabled": false,
        "input_template": "Binding question: {{input.binding_question}}\n\nApply Forced Analogy: pick 3 random unrelated domains, restate the question in each, generate 3-4 in-domain ideas, port each back. Close with one STANDOUT port. End your turn after the STANDOUT."
      }
    ]$STAGES$::jsonb,
    false, false, NULL, NULL,
    '["raw", "verified"]'::jsonb,
    false,
    jsonb_build_object(
        'shape', 'brainstorm-lens',
        'lens', 'forced-analogy',
        'default_model', 'qwen3.6-plus',
        'default_provider', 'opencode_go',
        'suggested_model', 'qwen3.6-plus',
        'suggested_provider', 'opencode_go'
    )
)
ON CONFLICT (family) DO UPDATE
   SET description                  = EXCLUDED.description,
       stages                       = EXCLUDED.stages,
       sabbath_enabled              = EXCLUDED.sabbath_enabled,
       atonement_enabled            = EXCLUDED.atonement_enabled,
       file_destination_template    = EXCLUDED.file_destination_template,
       file_content_jsonpath        = EXCLUDED.file_content_jsonpath,
       maturity_ladder              = EXCLUDED.maturity_ladder,
       auto_materialize_on_verified = EXCLUDED.auto_materialize_on_verified,
       metadata                     = EXCLUDED.metadata;


-- Worst Possible Idea -------------------------------------------------
INSERT INTO stewards.pipelines (
    family, description, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath,
    maturity_ladder, auto_materialize_on_verified, metadata
)
VALUES (
    'brainstorm-worst-idea',
    'Brainstorm lens: Worst Possible Idea. 5-7 intentionally terrible solutions → name the violated principle each embodies → invert into a positive design constraint.',
    $STAGES$[
      {
        "name": "lens",
        "next": null,
        "model": null,
        "provider": null,
        "agent_family": "brainstorm-worst-idea",
        "auto_advance": true,
        "tools_disabled": false,
        "input_template": "Binding question: {{input.binding_question}}\n\nApply Worst Possible Idea: 5-7 terrible solutions, each with violated-principle diagnosis, each inverted into a positive constraint. End your turn after the last Constraint."
      }
    ]$STAGES$::jsonb,
    false, false, NULL, NULL,
    '["raw", "verified"]'::jsonb,
    false,
    jsonb_build_object(
        'shape', 'brainstorm-lens',
        'lens', 'worst-idea',
        'default_model', 'qwen3.6-plus',
        'default_provider', 'opencode_go',
        'suggested_model', 'qwen3.6-plus',
        'suggested_provider', 'opencode_go'
    )
)
ON CONFLICT (family) DO UPDATE
   SET description                  = EXCLUDED.description,
       stages                       = EXCLUDED.stages,
       sabbath_enabled              = EXCLUDED.sabbath_enabled,
       atonement_enabled            = EXCLUDED.atonement_enabled,
       file_destination_template    = EXCLUDED.file_destination_template,
       file_content_jsonpath        = EXCLUDED.file_content_jsonpath,
       maturity_ladder              = EXCLUDED.maturity_ladder,
       auto_materialize_on_verified = EXCLUDED.auto_materialize_on_verified,
       metadata                     = EXCLUDED.metadata;


-- =====================================================================
-- Acceptance:
--   SELECT family,
--          stages->0->>'model' AS stage_model,
--          metadata->>'default_model' AS default_model,
--          metadata->>'suggested_model' AS suggested_model
--     FROM stewards.pipelines
--    WHERE family LIKE 'brainstorm-%'
--    ORDER BY family;
--   → 12 rows. All 12 stage_model = NULL. All 12 have default_model
--     populated. Suggested = default for new 8.
-- =====================================================================
