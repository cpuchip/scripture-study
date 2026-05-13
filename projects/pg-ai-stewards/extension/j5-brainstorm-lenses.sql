-- =====================================================================
-- Batch J.4 — Brainstorm pipeline + 4-lens library
-- =====================================================================
-- Brainstorm is implemented as a SPECIAL CASE of decompose-fanout:
--   * The decompose stage is pre-populated (deterministic — always emits
--     the 4 lens children + aggregator)
--   * Each lens is its own single-stage pipeline with its own system
--     prompt encoding the brainstorming technique
--   * The aggregator runs with synthesis=true, producing a ranked
--     candidate list across the 4 lenses
--
-- Lenses (per ratified B2 — start with 4):
--   * SCAMPER — Substitute/Combine/Adapt/Modify/Put-to/Eliminate/Reverse
--   * Six Hats — Edward de Bono's framework, focused on Green/White/Black
--   * Crazy 8s — 8 ideas in 8 minutes, volume over polish
--   * Reverse — generate failure modes first, then invert
--
-- Per ratified B1 (mix providers): kimi-k2.6 + qwen3.6-plus split across
-- the 4 lenses. Local LM Studio/Ollama integration deferred — add later
-- as new lens agents pointing at local provider rows.
--
-- Per ratified B3 (lens storage = agents table): each lens is a row in
-- stewards.agents with its own system prompt. Pipelines reference them
-- via agent_family.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. The 4 lens agents.
-- ---------------------------------------------------------------------

-- SCAMPER lens
INSERT INTO stewards.agents (family, model_match, description, mode, prompt, temperature, response_format)
VALUES (
    'brainstorm-scamper',
    '*',
    'SCAMPER brainstorming lens. Applies 7 transformations (Substitute/Combine/Adapt/Modify/Put-to/Eliminate/Reverse) to generate candidate ideas.',
    'primary',
    $PROMPT$You are the SCAMPER lens for a brainstorming pipeline. Given a binding question, apply the SCAMPER framework to generate distinct, concrete candidate ideas. SCAMPER prompts:

S — SUBSTITUTE: what could be swapped out for something else?
C — COMBINE: what could be merged with another idea, object, or process?
A — ADAPT: what existing solution from another domain could be adapted?
M — MAGNIFY / MINIFY: what could be made bigger, smaller, stronger, or weaker?
P — PUT TO OTHER USE: what other purposes could this serve?
E — ELIMINATE: what could be removed to simplify or focus?
R — REVERSE / REARRANGE: what if the order or relationship were flipped?

Generate 2-3 ideas per prompt letter (14-21 ideas total). Each idea has:
- A one-line title (max 12 words)
- A 2-3 sentence description explaining what it is and why it might work
- A SCAMPER tag in brackets, e.g. [S-Substitute]

Output ONE markdown list grouped by letter. No prose intro. No prose outro. End your turn after the list.$PROMPT$,
    0.7,
    NULL
)
ON CONFLICT (family, model_match) DO UPDATE
   SET description = EXCLUDED.description,
       mode        = EXCLUDED.mode,
       prompt      = EXCLUDED.prompt,
       temperature = EXCLUDED.temperature,
       active      = true;

-- Six Hats lens (focused on Green/White/Black for brainstorming)
INSERT INTO stewards.agents (family, model_match, description, mode, prompt, temperature, response_format)
VALUES (
    'brainstorm-six-hats',
    '*',
    'Six Thinking Hats lens (de Bono). Generates ideas through Green (creative), White (factual), and Black (critical) modes.',
    'primary',
    $PROMPT$You are the Six Hats lens for a brainstorming pipeline, applying Edward de Bono's Six Thinking Hats framework. For brainstorm purposes, focus on three complementary modes:

GREEN HAT — Creative, wild, "what if"
   Generate 5-7 unconventional, ambitious ideas. Don't worry about feasibility. Aim for variety in mechanism and scale.

WHITE HAT — Facts, data, what's been done before
   Generate 3-4 ideas grounded in existing examples from the literature, similar institutions, or proven patterns. Cite the precedent where you can.

BLACK HAT — Critical, what could go wrong
   Surface 3-4 risks, constraints, failure modes, or things to avoid. Each phrased as a thing to watch out for. These are constraints the other lenses (and downstream synthesis) should respect.

Total target: 11-15 items across the three hats.

Format each item as a bullet:
- One-line title
- 2-sentence description
- Hat tag in brackets at end, e.g. [GREEN], [WHITE], [BLACK]

Output ONE markdown list grouped by hat. No prose intro. No prose outro. End your turn after the list.$PROMPT$,
    0.8,
    NULL
)
ON CONFLICT (family, model_match) DO UPDATE
   SET description = EXCLUDED.description,
       mode        = EXCLUDED.mode,
       prompt      = EXCLUDED.prompt,
       temperature = EXCLUDED.temperature,
       active      = true;

-- Crazy 8s lens
INSERT INTO stewards.agents (family, model_match, description, mode, prompt, temperature, response_format)
VALUES (
    'brainstorm-crazy8s',
    '*',
    'Crazy 8s lens. Sprint exercise: 8 distinct ideas in 8 minutes. Volume over polish; deliberate variety.',
    'primary',
    $PROMPT$You are the Crazy 8s lens for a brainstorming pipeline. Crazy 8s is a sprint technique: generate 8 distinct ideas fast, prioritizing VOLUME and VARIETY over polish.

Rules:
- Exactly 8 ideas. Number them 1-8.
- Each idea is ONE sentence (max 25 words).
- Ideas must be DISTINCT — no variations of the same theme.
- Deliberate variety across the 8:
  * At least one OBVIOUS idea (the first thing anyone would think of)
  * At least one WEIRD idea (unconventional mechanism or framing)
  * At least one ADJACENT-DOMAIN idea (stolen from a different field)
  * At least one MOONSHOT (impossible or expensive but interesting)
  * Fill the rest with whatever mix feels right
- Tag each with ONE keyword in brackets at the end, e.g. [obvious], [weird], [adjacent-domain], [moonshot], [cheap], [community], [tech], [analog].

Output a numbered markdown list (1-8). No prose intro. No prose outro. End your turn after item 8.$PROMPT$,
    0.9,
    NULL
)
ON CONFLICT (family, model_match) DO UPDATE
   SET description = EXCLUDED.description,
       mode        = EXCLUDED.mode,
       prompt      = EXCLUDED.prompt,
       temperature = EXCLUDED.temperature,
       active      = true;

-- Reverse brainstorm lens
INSERT INTO stewards.agents (family, model_match, description, mode, prompt, temperature, response_format)
VALUES (
    'brainstorm-reverse',
    '*',
    'Reverse brainstorm lens. Generate failure modes first, then invert each into a positive solution.',
    'primary',
    $PROMPT$You are the Reverse Brainstorm lens. Instead of "how could we solve this?", you ask "how could we GUARANTEE FAILURE here?" — then INVERT each failure mode into a solution.

The technique works because identifying failure modes is often easier than identifying solutions, and the inversion produces solutions that specifically guard against the worst outcomes.

Step 1 — FAILURE MODES: Generate 5-7 specific ways this question could be answered badly. Concrete and specific. Format each as:
  - Failure mode: <description>

Step 2 — INVERSIONS: For each failure mode, write the opposite/protective approach. Format each as:
  - → Inverted: <description of the protective or opposite approach>

Output as a markdown list of paired items. No prose intro. No prose outro. End your turn after the last inversion.$PROMPT$,
    0.7,
    NULL
)
ON CONFLICT (family, model_match) DO UPDATE
   SET description = EXCLUDED.description,
       mode        = EXCLUDED.mode,
       prompt      = EXCLUDED.prompt,
       temperature = EXCLUDED.temperature,
       active      = true;


-- ---------------------------------------------------------------------
-- 2. The 4 lens pipelines. Each single-stage, references its lens agent.
-- ---------------------------------------------------------------------

INSERT INTO stewards.pipelines (
    family, description, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath,
    maturity_ladder, auto_materialize_on_verified, metadata
)
VALUES
(
    'brainstorm-scamper',
    'Brainstorm lens: SCAMPER. Single-stage pipeline emitting 14-21 candidate ideas tagged by transformation.',
    $STAGES$[
      {
        "name": "lens",
        "next": null,
        "model": "qwen3.6-plus",
        "provider": "opencode_go",
        "agent_family": "brainstorm-scamper",
        "auto_advance": true,
        "tools_disabled": false,
        "input_template": "Binding question: {{input.binding_question}}\n\nApply your SCAMPER framework. Return ONE markdown list. End your turn after the list."
      }
    ]$STAGES$::jsonb,
    false, false, NULL, NULL,
    '["raw", "verified"]'::jsonb,
    false,
    jsonb_build_object('shape', 'brainstorm-lens', 'lens', 'scamper')
),
(
    'brainstorm-six-hats',
    'Brainstorm lens: Six Thinking Hats (Green/White/Black focus).',
    $STAGES$[
      {
        "name": "lens",
        "next": null,
        "model": "kimi-k2.6",
        "provider": "opencode_go",
        "agent_family": "brainstorm-six-hats",
        "auto_advance": true,
        "tools_disabled": false,
        "input_template": "Binding question: {{input.binding_question}}\n\nApply the Six Thinking Hats framework. Return ONE markdown list grouped by hat. End your turn after the list."
      }
    ]$STAGES$::jsonb,
    false, false, NULL, NULL,
    '["raw", "verified"]'::jsonb,
    false,
    jsonb_build_object('shape', 'brainstorm-lens', 'lens', 'six-hats')
),
(
    'brainstorm-crazy8s',
    'Brainstorm lens: Crazy 8s. 8 ideas, 8 minutes, deliberate variety.',
    $STAGES$[
      {
        "name": "lens",
        "next": null,
        "model": "qwen3.6-plus",
        "provider": "opencode_go",
        "agent_family": "brainstorm-crazy8s",
        "auto_advance": true,
        "tools_disabled": false,
        "input_template": "Binding question: {{input.binding_question}}\n\nApply Crazy 8s. Output 8 numbered ideas with one-line descriptions and a keyword tag each. End your turn after item 8."
      }
    ]$STAGES$::jsonb,
    false, false, NULL, NULL,
    '["raw", "verified"]'::jsonb,
    false,
    jsonb_build_object('shape', 'brainstorm-lens', 'lens', 'crazy8s')
),
(
    'brainstorm-reverse',
    'Brainstorm lens: Reverse — generate failure modes first, then invert each into a positive solution.',
    $STAGES$[
      {
        "name": "lens",
        "next": null,
        "model": "kimi-k2.6",
        "provider": "opencode_go",
        "agent_family": "brainstorm-reverse",
        "auto_advance": true,
        "tools_disabled": false,
        "input_template": "Binding question: {{input.binding_question}}\n\nApply Reverse Brainstorm: list 5-7 failure modes, then invert each. End your turn after the last inversion."
      }
    ]$STAGES$::jsonb,
    false, false, NULL, NULL,
    '["raw", "verified"]'::jsonb,
    false,
    jsonb_build_object('shape', 'brainstorm-lens', 'lens', 'reverse')
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


-- ---------------------------------------------------------------------
-- 3. start_brainstorm() — user-facing entry point.
-- ---------------------------------------------------------------------
-- Creates a decompose-fanout parent with a pre-populated 4-lens
-- manifest + synthesis aggregator. Triggers spawn immediately via
-- maturity=verified. Returns the parent uuid.
--
-- Composes from existing primitives: brainstorm is just a special-case
-- fanout where the children are 4 fixed lens pipelines.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.start_brainstorm(
    p_binding_question text,
    p_destination text,
    p_project_association text DEFAULT NULL,
    p_actor text DEFAULT 'human',
    p_slug text DEFAULT NULL,
    p_cost_cap_per_lens_micro bigint DEFAULT 200000
)
RETURNS uuid LANGUAGE plpgsql AS $FN$
DECLARE
    v_slug      text;
    v_parent_id uuid;
    v_manifest  jsonb;
BEGIN
    v_slug := COALESCE(p_slug, 'brainstorm-' || to_char(now() AT TIME ZONE 'UTC', 'YYYYMMDD-HH24MISS'));

    v_manifest := jsonb_build_object(
        'rationale', 'Brainstorm: 4 lenses (SCAMPER, Six Hats, Crazy 8s, Reverse) run in parallel, converged via synthesis aggregator.',
        'children', jsonb_build_array(
            jsonb_build_object(
                'slug', v_slug || '-scamper',
                'pipeline_family', 'brainstorm-scamper',
                'binding_question', p_binding_question,
                'cost_cap_micro', p_cost_cap_per_lens_micro
            ),
            jsonb_build_object(
                'slug', v_slug || '-six-hats',
                'pipeline_family', 'brainstorm-six-hats',
                'binding_question', p_binding_question,
                'cost_cap_micro', p_cost_cap_per_lens_micro
            ),
            jsonb_build_object(
                'slug', v_slug || '-crazy8s',
                'pipeline_family', 'brainstorm-crazy8s',
                'binding_question', p_binding_question,
                'cost_cap_micro', p_cost_cap_per_lens_micro
            ),
            jsonb_build_object(
                'slug', v_slug || '-reverse',
                'pipeline_family', 'brainstorm-reverse',
                'binding_question', p_binding_question,
                'cost_cap_micro', p_cost_cap_per_lens_micro
            )
        ),
        'aggregate', jsonb_build_object(
            'destination', p_destination,
            'synthesis', true
        )
    );

    INSERT INTO stewards.work_items (
        pipeline_family, current_stage, slug, input, intent_id, actor,
        project_association, stage_results, maturity, status
    ) VALUES (
        'decompose-fanout',
        'decompose',
        v_slug,
        jsonb_build_object('binding_question', p_binding_question),
        (SELECT id FROM stewards.intents WHERE slug = 'scripture-study'),
        p_actor,
        p_project_association,
        jsonb_build_object(
            'context_gather', jsonb_build_object('output', 'brainstorm: pre-populated 4-lens manifest, no context_gather LLM call'),
            'decompose', jsonb_build_object('output', v_manifest)
        ),
        'planned',
        'completed'
    )
    RETURNING id INTO v_parent_id;

    UPDATE stewards.work_items SET maturity = 'verified' WHERE id = v_parent_id;

    RAISE NOTICE 'start_brainstorm: parent=% slug=%', v_parent_id, v_slug;
    RETURN v_parent_id;
END;
$FN$;

COMMENT ON FUNCTION stewards.start_brainstorm(text, text, text, text, text, bigint) IS
'Batch J.4: brainstorm entry point. Creates a decompose-fanout work_item with a 4-lens manifest (SCAMPER, Six Hats, Crazy 8s, Reverse) + synthesis aggregator. Triggers spawn immediately. Returns parent uuid.';

-- =====================================================================
-- End of j5-brainstorm-lenses.sql
-- =====================================================================
