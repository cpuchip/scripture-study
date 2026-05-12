-- =====================================================================
-- H.3-followup #3 — revise-proposal pipeline + apply_revision helper
--
-- New pipeline family for AI-revising an existing proposed work_item.
-- Each revise creates its own work_item (origin=human, dispatched by
-- the UI's "Revise with feedback" button) whose:
--   - parent_work_item_id = the proposal being revised
--   - input.feedback      = user's text describing what to change
--
-- One stage `revise` reads:
--   - The original proposal (via parent_work_item_id link)
--   - The grandparent's planning context (for project + plan alignment)
--   - input.feedback
-- Emits strict JSON: { binding_question, rationale,
--                      slug?, pipeline_family_hint?, project_association? }
-- Maturity → verified at completion. No sabbath or auto-materialize.
--
-- After verified: the UI fetches /api/work-items/{id}/pending-revisions
-- to show a diff card. apply_revision SQL function UPDATEs the original
-- proposal when user clicks Accept.
--
-- Q-revise-3: rejected revises get status='cancelled' (existing column);
-- accepted revises get revision_applied_at = now() (new column).
-- =====================================================================

-- New column on work_items for tracking accepted revises.
ALTER TABLE stewards.work_items
    ADD COLUMN IF NOT EXISTS revision_applied_at timestamp with time zone;

-- Pipeline definition.
DO $$
DECLARE
    v_revise_template text;
    v_stages          jsonb;
BEGIN

v_revise_template :=
$T$You are revising a proposed work_item based on user feedback.

## ORIGINAL PROPOSAL (the work_item being revised)

- slug: {{input.original_slug}}
- binding_question: {{input.original_binding_question}}
- rationale: {{input.original_rationale}}
- pipeline_family_hint: {{input.original_pipeline_family_hint}}
- project_association: {{input.original_project_association}}

## PARENT PLANNING CONTEXT (excerpt)

{{input.parent_plan_excerpt}}

## USER FEEDBACK

{{input.feedback}}

## YOUR TASK — emit a JSON revision

Read the original + parent context + user feedback. Emit a JSON object with the REVISED fields. Only include fields you're changing — omit fields that stay the same. The substrate will merge your output into the original.

## SCHEMA

```json
{
  "binding_question":     "Revised question text (optional)",
  "rationale":            "Revised rationale, one sentence (optional)",
  "slug":                 "revised-kebab-case-slug (optional)",
  "pipeline_family_hint": "research-write | planning | null (optional)",
  "project_association":  "string or null (optional)"
}
```

## HARD CONSTRAINTS

- **Output ONLY the JSON object.** No prose intro/outro. No markdown fences.
- **Honor the user's feedback as the primary signal.** If they say "scope tighter," tighten. If they say "rephrase," rephrase. Don't second-guess.
- **Preserve fields the user didn't ask to change.** Omit them from your output.
- **slug regex: ^[a-z0-9-]+$** if you're changing it.
- **binding_question must be a complete question** ending in `?` and ≥20 chars.

## EXAMPLE

User feedback: "scope this tighter — just validate the laptop webcams, not the full ML stack"

Original binding_question: "Do all five repurposed laptops support Chrome kiosk mode and offline TensorFlow.js webcam inference without driver conflicts or privacy blocks?"

Revision:
```json
{
  "binding_question": "Do all five repurposed laptops have functional built-in webcams accessible to Chrome under a kiosk-mode user profile, ignoring ML stack validation for a later work_item?",
  "rationale": "Splits hardware compatibility from software validation so the cheaper hardware test runs first."
}
```

Your turn. Output ONLY the JSON.$T$;

v_stages := jsonb_build_array(
    jsonb_build_object(
        'name', 'revise',
        'next', NULL,
        'model', 'qwen3.6-plus',
        'provider', 'opencode_go',
        'agent_family', 'research',
        'auto_advance', true,
        'tools_disabled', true,
        'input_template', v_revise_template
    )
);

INSERT INTO stewards.pipelines (
    family, description, stages, metadata,
    sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath,
    maturity_ladder, auto_materialize_on_verified
)
VALUES (
    'revise-proposal',
    'AI revision of an existing proposed work_item. Reads the original + parent plan + user feedback; emits a JSON partial revision; the UI shows a diff card with Accept/Reject. Used by the "Revise with feedback" button on WorkItemDetail.vue when the proposal is close but needs a tweak. parent_work_item_id MUST be set to the proposal being revised.',
    v_stages,
    jsonb_build_object(
        'cost_cap_default_micro', 100000,
        'cost_cap_default_dollars', 0.10,
        'note', 'Single stage, qwen3.6-plus, tools off; typical cost $0.02-0.05'
    ),
    false,  -- sabbath_enabled: no, revise is a small interactive call
    false,  -- atonement_enabled: no, revisions don't quarantine
    NULL,   -- file_destination_template: no file artifact
    'stage_results.revise.output',  -- where the revision JSON lives
    -- maturity_ladder: just goes raw → verified
    '["raw","verified"]'::jsonb,
    false   -- auto_materialize_on_verified: no file write
)
ON CONFLICT (family) DO UPDATE
   SET description                  = EXCLUDED.description,
       stages                       = EXCLUDED.stages,
       metadata                     = EXCLUDED.metadata,
       sabbath_enabled              = EXCLUDED.sabbath_enabled,
       atonement_enabled            = EXCLUDED.atonement_enabled,
       file_destination_template    = EXCLUDED.file_destination_template,
       file_content_jsonpath        = EXCLUDED.file_content_jsonpath,
       maturity_ladder              = EXCLUDED.maturity_ladder,
       auto_materialize_on_verified = EXCLUDED.auto_materialize_on_verified,
       updated_at                   = now();

END $$;

-- Stage maturity: revise → verified
INSERT INTO stewards.pipeline_stage_maturity (pipeline_family, stage_name, produces_maturity)
VALUES ('revise-proposal', 'revise', 'verified')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE
   SET produces_maturity = EXCLUDED.produces_maturity;

INSERT INTO stewards.stage_models (pipeline_family, stage_name, default_model)
VALUES ('revise-proposal', 'revise', 'qwen3.6-plus')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE
   SET default_model = EXCLUDED.default_model;

-- =====================================================================
-- stewards.apply_revision — UPDATE the original proposal with the
-- revision JSON. Called by the /api/work-items/apply-revision endpoint.
--
-- Returns true if revision applied; false if validation failed or
-- already applied. NOTICE-logs reasons for skip (not RAISE EXCEPTION
-- so the caller gets clean feedback via the bool).
-- =====================================================================

CREATE OR REPLACE FUNCTION stewards.apply_revision(p_revise_work_item_id uuid)
RETURNS boolean
LANGUAGE plpgsql
AS $func$
DECLARE
    v_revise   stewards.work_items%ROWTYPE;
    v_original stewards.work_items%ROWTYPE;
    v_raw      text;
    v_clean    text;
    v_json     jsonb;
    v_new_slug text;
    v_new_binding text;
    v_new_rationale text;
    v_new_hint text;
    v_new_project text;
BEGIN
    SELECT * INTO v_revise FROM stewards.work_items WHERE id = p_revise_work_item_id;
    IF v_revise.id IS NULL THEN
        RAISE NOTICE 'apply_revision: revise work_item % not found', p_revise_work_item_id;
        RETURN false;
    END IF;
    IF v_revise.pipeline_family <> 'revise-proposal' THEN
        RAISE NOTICE 'apply_revision: work_item % is not a revise-proposal (family=%)',
            p_revise_work_item_id, v_revise.pipeline_family;
        RETURN false;
    END IF;
    IF v_revise.revision_applied_at IS NOT NULL THEN
        RAISE NOTICE 'apply_revision: revision already applied at %', v_revise.revision_applied_at;
        RETURN false;
    END IF;
    IF v_revise.status = 'cancelled' THEN
        RAISE NOTICE 'apply_revision: revision was rejected (status=cancelled)';
        RETURN false;
    END IF;
    IF v_revise.parent_work_item_id IS NULL THEN
        RAISE NOTICE 'apply_revision: revision % has no parent_work_item_id', p_revise_work_item_id;
        RETURN false;
    END IF;

    SELECT * INTO v_original FROM stewards.work_items WHERE id = v_revise.parent_work_item_id;
    IF v_original.id IS NULL THEN
        RAISE NOTICE 'apply_revision: parent (original) work_item % not found', v_revise.parent_work_item_id;
        RETURN false;
    END IF;

    -- Parse the revision JSON.
    v_raw := (v_revise.stage_results -> 'revise' -> 'output') #>> '{}';
    IF v_raw IS NULL OR length(trim(v_raw)) = 0 THEN
        RAISE NOTICE 'apply_revision: revise.output is empty';
        RETURN false;
    END IF;

    -- Strip markdown fences defensively.
    v_clean := regexp_replace(v_raw, E'^\\s*```(?:json)?\\s*\\n?|\\n?```\\s*$', '', 'g');
    v_clean := trim(v_clean);

    BEGIN
        v_json := v_clean::jsonb;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'apply_revision: JSON parse failed: %', SQLERRM;
        RETURN false;
    END;

    -- Extract revision fields (optional — only present ones get applied).
    v_new_slug      := v_json ->> 'slug';
    v_new_binding   := v_json ->> 'binding_question';
    v_new_rationale := v_json ->> 'rationale';
    v_new_hint      := v_json ->> 'pipeline_family_hint';
    v_new_project   := v_json ->> 'project_association';

    -- Validate any fields that were provided.
    IF v_new_slug IS NOT NULL THEN
        IF v_new_slug !~ '^[a-z0-9-]+$' THEN
            RAISE NOTICE 'apply_revision: invalid slug %', v_new_slug;
            RETURN false;
        END IF;
        IF EXISTS (
            SELECT 1 FROM stewards.work_items
             WHERE slug = v_new_slug AND id <> v_original.id
        ) THEN
            RAISE NOTICE 'apply_revision: slug % already in use', v_new_slug;
            RETURN false;
        END IF;
    END IF;
    IF v_new_binding IS NOT NULL AND length(trim(v_new_binding)) < 20 THEN
        RAISE NOTICE 'apply_revision: binding_question too short';
        RETURN false;
    END IF;

    -- Resolve pipeline_family_hint to actual pipeline_family if changed.
    -- "null" string from the model = explicit clear-hint; treat as NULL.
    IF v_new_hint IS NOT NULL AND v_new_hint = 'null' THEN
        v_new_hint := NULL;
    END IF;

    -- UPDATE the original. COALESCE keeps existing values for fields
    -- the revision didn't touch.
    UPDATE stewards.work_items
       SET slug            = COALESCE(v_new_slug, slug),
           input           = input
                          || COALESCE(
                               CASE WHEN v_new_binding IS NOT NULL
                                    THEN jsonb_build_object('binding_question', v_new_binding)
                                    ELSE NULL END,
                               '{}'::jsonb)
                          || COALESCE(
                               CASE WHEN v_new_rationale IS NOT NULL
                                    THEN jsonb_build_object('rationale_from_planning', v_new_rationale)
                                    ELSE NULL END,
                               '{}'::jsonb),
           pipeline_family = CASE
                                WHEN v_new_hint IS NOT NULL AND EXISTS (
                                    SELECT 1 FROM stewards.pipelines WHERE family = v_new_hint
                                ) THEN v_new_hint
                                ELSE pipeline_family
                             END,
           current_stage   = CASE
                                WHEN v_new_hint IS NOT NULL AND EXISTS (
                                    SELECT 1 FROM stewards.pipelines WHERE family = v_new_hint
                                ) THEN stewards.pipeline_first_stage_name(v_new_hint)
                                ELSE current_stage
                             END,
           project_association = CASE
                                    WHEN v_json ? 'project_association'
                                        THEN v_new_project
                                    ELSE project_association
                                END,
           updated_at      = now()
     WHERE id = v_original.id;

    -- Mark the revise work_item as applied.
    UPDATE stewards.work_items
       SET revision_applied_at = now(),
           updated_at = now()
     WHERE id = p_revise_work_item_id;

    RAISE NOTICE 'apply_revision: applied revision % to original %',
        p_revise_work_item_id, v_original.id;
    RETURN true;
END;
$func$;

COMMENT ON FUNCTION stewards.apply_revision(uuid) IS
'H.3-followup-3: applies a completed revise-proposal work_item to its parent (the original proposal). Validates schema, UPDATEs original with non-null revision fields (using COALESCE to preserve unchanged values), marks the revise work_item with revision_applied_at. Idempotent — re-call after applied returns false.';

-- Sanity check.
SELECT family,
       jsonb_array_length(stages) AS n_stages,
       maturity_ladder,
       (stages->0)->>'tools_disabled' AS tools_off
  FROM stewards.pipelines WHERE family='revise-proposal';
