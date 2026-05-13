-- =====================================================================
-- Batch J.2 — Fan-out pipeline machinery
-- =====================================================================
-- Adds two new pipeline shapes (decompose-fanout, aggregate-children),
-- the spawn_children() SQL function that creates the N child work_items
-- + one aggregator-child, and extends on_maturity_verified to wire the
-- parent -> children -> aggregator chain.
--
-- Ratified in projects/pg-ai-stewards/.spec/proposals/
--               substrate-batch-j-fanout-brainstorm.md (2026-05-13).
-- Decisions: A1 deterministic SQL, A2 event-driven via on_maturity_verified,
--            A3 per-child pipeline assignment, A4 index always.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. Register the decomposer + aggregator agents.
-- ---------------------------------------------------------------------

INSERT INTO stewards.agents (family, model_match, description, mode, prompt, temperature, response_format)
VALUES (
    'fanout-decompose',
    '*',
    'Decomposes a binding question into N child work_items + 1 aggregator manifest. Output is strict JSON.',
    'primary',
    $PROMPT$You are the decomposer for a fan-out pipeline. Your job is to take a single binding question and produce a manifest of N child work_items, each focused on a specific sub-question, plus an aggregate destination.

OUTPUT ONLY VALID JSON in this schema. No prose around it. No markdown fences.

{
  "rationale": "1-3 sentences explaining the decomposition",
  "children": [
    {
      "slug": "kebab-case-unique-slug",
      "binding_question": "specific scoped question for this child (1-3 sentences)",
      "pipeline_family": "research-write | study-write | study-write-qwen",
      "project_association": "optional string",
      "input_extra": {},
      "cost_cap_micro": 500000
    }
  ],
  "aggregate": {
    "destination": "relative/path/to/index.md",
    "synthesis": false
  }
}

RULES:
- Each child's binding_question must be tightly scoped — one artifact's worth of work.
- Choose pipeline_family per child based on what shape the deliverable needs.
- Keep child count to 3-12. If the natural decomposition exceeds 12, group into categories and decompose each category in a follow-on fan-out.
- aggregate.destination is the INDEX file, distinct from any child's destination.
- cost_cap_micro is in micro-dollars; default 500000 = $0.50 per child.
- Use input_extra only if the child needs values beyond binding_question (e.g. {"audience": "youth", "deliverable": "exhibit-brief"}).

You have read-only tools (fs_*, study_*, work_item_*) available if you need to inspect prior context, but spend at most 2 rounds of tool calls. Your output is the manifest; the next stage is spawn (deterministic), not another LLM call. Keep the JSON minimal and valid.$PROMPT$,
    0.3,
    NULL
)
ON CONFLICT (family, model_match) DO UPDATE
   SET description = EXCLUDED.description,
       mode        = EXCLUDED.mode,
       prompt      = EXCLUDED.prompt,
       temperature = EXCLUDED.temperature,
       response_format = EXCLUDED.response_format,
       active      = true;


INSERT INTO stewards.agents (family, model_match, description, mode, prompt, temperature, response_format)
VALUES (
    'fanout-aggregate',
    '*',
    'Aggregates completed children into an index file. Reads input.children + their stage_results / files; emits markdown.',
    'primary',
    $PROMPT$You are the aggregator for a fan-out pipeline. The decomposer split a binding question into N children, each ran on its own pipeline, and now you stitch their results into one index file.

You will receive in your input:
- `parent_work_item_id` — the original binding question's work_item id
- `destination` — the file path the index will be written to
- `synthesis` — boolean. If true, ALSO produce a digest section with cross-cutting themes. If false, index only.
- `children` — array of {id, slug, binding_question, pipeline_family} for each child

YOUR JOB:
- For each child, read its output. Use `work_item_show` with the child's id to read its stage_results. The most useful field is usually `stage_results.review.output` or `stage_results.synthesize.output` (last meaningful stage) or its `file_destination` if set.
- Compose an index in markdown:

```
# <title derived from parent's binding question>

Brief intro paragraph (2-4 sentences): what the children collectively answer.

## Children

| Slug | Title | One-line summary | Link |
|---|---|---|---|
| ... | ... | ... | [...](path-to-child-file.md) |

(if synthesis=true add:)

## Synthesis

Cross-cutting themes 2-4 paragraphs. What's true across the children? What pattern emerged? What still feels unanswered?
```

RULES:
- Verify a child's output before quoting it. If you can't read it (tool failure, missing field), say "child <slug> output unavailable" rather than confabulate.
- Each child gets ONE one-line summary in the table. Compress aggressively.
- If synthesis=false, do NOT include the Synthesis section.
- Output ONLY the markdown body — no JSON wrapper, no commentary about what you did.
- Keep total output under 4KB. The index is a navigation aid, not the artifact itself; the artifacts are the children.$PROMPT$,
    0.4,
    NULL
)
ON CONFLICT (family, model_match) DO UPDATE
   SET description = EXCLUDED.description,
       mode        = EXCLUDED.mode,
       prompt      = EXCLUDED.prompt,
       temperature = EXCLUDED.temperature,
       response_format = EXCLUDED.response_format,
       active      = true;


-- ---------------------------------------------------------------------
-- 2. Register the decompose-fanout pipeline (2 stages: context_gather, decompose).
-- ---------------------------------------------------------------------

INSERT INTO stewards.pipelines (
    family, description, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath,
    maturity_ladder, auto_materialize_on_verified, metadata
)
VALUES (
    'decompose-fanout',
    'Fan-out: decompose a binding question into N child work_items + an aggregator child. spawn fires on maturity=verified via on_maturity_verified trigger; aggregator dispatches when all siblings verified.',
    $STAGES$[
      {
        "name": "context_gather",
        "next": "decompose",
        "model": "qwen3.6-plus",
        "provider": "opencode_go",
        "agent_family": "research",
        "auto_advance": true,
        "tools_disabled": false,
        "input_template": "Binding question: {{input.binding_question}}\n\n## YOUR TASK — context briefing for the decomposer\n\nThis is the context_gather stage of a fan-out pipeline. The next stage (decompose) will split this binding question into N child work_items. Brief the decomposer on:\n\n1. **Prior work on this question** — search `.spec/journal/`, `.spec/proposals/`, `.mind/*`, `docs/**`, and prior work_items. Has this question been tackled before? How was it scoped?\n2. **Natural decomposition axes** — for THIS question, what are the obvious sub-questions? Categories? Phases? Stakeholders? Don't decide the decomposition — that's the decomposer's job — just surface the axes.\n3. **Existing artifacts to NOT duplicate** — if children would produce files that overlap with files already in the repo, name them so the decomposer can adjust scope.\n\nHARD CONSTRAINTS:\n- Maximum 3 rounds of tool calls.\n- Output budget: ~1.5KB.\n- End-of-turn: your final message is the briefing in markdown, then STOP.\n\nOUTPUT FORMAT:\n\n## Prior work\n<bullets, file paths>\n\n## Decomposition axes worth considering\n<bullets of axes, not the decomposition itself>\n\n## Existing artifacts to avoid duplicating\n<bullets, file paths>\n\nIf there's no prior work, say so — the decomposer will start fresh."
      },
      {
        "name": "decompose",
        "next": null,
        "model": "qwen3.6-plus",
        "provider": "opencode_go",
        "agent_family": "fanout-decompose",
        "auto_advance": true,
        "tools_disabled": false,
        "input_template": "Binding question: {{input.binding_question}}\n\n## CONTEXT BRIEFING (from context_gather stage)\n\n{{stage_results.context_gather.output}}\n\n## YOUR TASK\n\nProduce the decomposition manifest as JSON only (no prose, no fences). Follow the schema in your system prompt exactly. Each child must have slug, binding_question, pipeline_family. Aggregate must have destination and synthesis."
      }
    ]$STAGES$::jsonb,
    false,                           -- sabbath_enabled
    false,                           -- atonement_enabled
    NULL,                            -- no parent file destination (children write files)
    NULL,                            -- no file_content_jsonpath
    '["raw", "planned", "verified"]'::jsonb,
    false,                           -- spawn fires via trigger, not auto-materialize
    jsonb_build_object('shape', 'fanout', 'spawn_on_verified', true)
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
-- 3. Register the aggregate-children pipeline (1 stage: aggregate).
-- ---------------------------------------------------------------------

INSERT INTO stewards.pipelines (
    family, description, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath,
    maturity_ladder, auto_materialize_on_verified, metadata
)
VALUES (
    'aggregate-children',
    'Aggregator: writes an index of completed fan-out children to a single file. Spawned by spawn_children() in status=pending; dispatched by on_maturity_verified when all sibling children verify.',
    $STAGES$[
      {
        "name": "aggregate",
        "next": null,
        "model": "qwen3.6-plus",
        "provider": "opencode_go",
        "agent_family": "fanout-aggregate",
        "auto_advance": true,
        "tools_disabled": false,
        "input_template": "## Fan-out aggregation\n\nParent binding question: {{input.binding_question}}\n\nDestination: {{input.destination}}\nSynthesis: {{input.synthesis}}\n\nChildren to aggregate:\n\n{{input.children}}\n\nUse work_item_show on each child id to read its output. Compose the index per your system prompt. Return ONLY the markdown body."
      }
    ]$STAGES$::jsonb,
    false,
    false,
    '{{input.destination}}',         -- file destination comes from input
    'stage_results.aggregate.output',
    '["raw", "verified"]'::jsonb,
    true,                            -- aggregate output auto-materializes
    jsonb_build_object('shape', 'aggregate')
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
-- 4. spawn_children() — deterministic SQL function (A1 ratified).
-- ---------------------------------------------------------------------
-- Reads the decompose manifest from stage_results, validates schema,
-- inserts N children + 1 aggregator child. Returns count of regular
-- children spawned (not counting the aggregator).
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.spawn_children(p_parent_id uuid)
RETURNS int LANGUAGE plpgsql AS $FN$
DECLARE
    v_parent        stewards.work_items%ROWTYPE;
    v_manifest      jsonb;
    v_manifest_raw  text;
    v_child         jsonb;
    v_child_id      uuid;
    v_count         int := 0;
    v_aggregator    jsonb;
    v_agg_id        uuid;
    v_children_arr  jsonb := '[]'::jsonb;
    v_child_pipeline text;
    v_child_slug    text;
    v_child_input   jsonb;
    v_cost_cap      bigint;
BEGIN
    SELECT * INTO v_parent FROM stewards.work_items WHERE id = p_parent_id;
    IF v_parent.id IS NULL THEN
        RAISE EXCEPTION 'spawn_children: parent % not found', p_parent_id;
    END IF;

    -- Pull manifest from decompose output. LLM stages typically store
    -- the output as a JSON string (text). Try to coerce.
    v_manifest := v_parent.stage_results -> 'decompose' -> 'output';
    IF v_manifest IS NULL THEN
        RAISE EXCEPTION 'spawn_children: no decompose output on parent %', p_parent_id;
    END IF;

    IF jsonb_typeof(v_manifest) = 'string' THEN
        v_manifest_raw := v_manifest #>> '{}';
        BEGIN
            v_manifest := v_manifest_raw::jsonb;
        EXCEPTION WHEN OTHERS THEN
            RAISE EXCEPTION 'spawn_children: decompose output is not valid JSON: %', SQLERRM;
        END;
    END IF;

    IF v_manifest -> 'children' IS NULL
       OR jsonb_typeof(v_manifest -> 'children') <> 'array'
       OR jsonb_array_length(v_manifest -> 'children') = 0 THEN
        RAISE EXCEPTION 'spawn_children: manifest.children is missing or empty';
    END IF;

    IF v_manifest -> 'aggregate' IS NULL
       OR (v_manifest -> 'aggregate' ->> 'destination') IS NULL THEN
        RAISE EXCEPTION 'spawn_children: manifest.aggregate.destination is required';
    END IF;

    -- Spawn regular children.
    FOR v_child IN SELECT * FROM jsonb_array_elements(v_manifest -> 'children') LOOP
        v_child_pipeline := v_child ->> 'pipeline_family';
        v_child_slug     := v_child ->> 'slug';

        IF v_child_pipeline IS NULL OR v_child_slug IS NULL
           OR (v_child ->> 'binding_question') IS NULL THEN
            RAISE EXCEPTION 'spawn_children: child entry missing slug/pipeline_family/binding_question: %', v_child;
        END IF;

        v_child_input := jsonb_build_object(
            'binding_question', v_child ->> 'binding_question'
        );
        IF (v_child -> 'input_extra') IS NOT NULL
           AND jsonb_typeof(v_child -> 'input_extra') = 'object' THEN
            v_child_input := v_child_input || (v_child -> 'input_extra');
        END IF;

        v_child_id := stewards.work_item_create(
            p_pipeline_family => v_child_pipeline,
            p_input           => v_child_input,
            p_slug            => v_child_slug,
            p_actor           => v_parent.actor,
            p_intent_id       => v_parent.intent_id
        );

        v_cost_cap := NULL;
        IF (v_child ->> 'cost_cap_micro') IS NOT NULL THEN
            v_cost_cap := (v_child ->> 'cost_cap_micro')::bigint;
        END IF;

        UPDATE stewards.work_items
           SET parent_work_item_id = p_parent_id,
               project_association = COALESCE(
                   v_child ->> 'project_association',
                   v_parent.project_association
               ),
               cost_cap_micro = COALESCE(v_cost_cap, cost_cap_micro)
         WHERE id = v_child_id;

        -- Dispatch each child immediately so they start processing in parallel.
        PERFORM stewards.work_item_dispatch_stage(v_child_id, NULL);

        v_children_arr := v_children_arr || jsonb_build_object(
            'id', v_child_id::text,
            'slug', v_child_slug,
            'binding_question', v_child ->> 'binding_question',
            'pipeline_family', v_child_pipeline
        );
        v_count := v_count + 1;
    END LOOP;

    -- Spawn the aggregator (NOT dispatched yet — waits for siblings).
    v_aggregator := v_manifest -> 'aggregate';

    v_agg_id := stewards.work_item_create(
        p_pipeline_family => 'aggregate-children',
        p_input           => jsonb_build_object(
            'binding_question', 'Aggregate index for: ' || COALESCE(v_parent.input ->> 'binding_question', v_parent.slug),
            'parent_work_item_id', p_parent_id::text,
            'destination', v_aggregator ->> 'destination',
            'synthesis', COALESCE((v_aggregator ->> 'synthesis')::boolean, false),
            'children', v_children_arr
        ),
        p_slug            => COALESCE(v_parent.slug, p_parent_id::text) || '-aggregator',
        p_actor           => v_parent.actor,
        p_intent_id       => v_parent.intent_id
    );

    UPDATE stewards.work_items
       SET parent_work_item_id = p_parent_id,
           project_association = v_parent.project_association
     WHERE id = v_agg_id;
    -- aggregator stays at status='pending' (work_item_create default).
    -- on_maturity_verified will flip it to 'dispatched' when siblings done.

    RAISE NOTICE 'spawn_children: parent=% spawned % children + aggregator %',
        p_parent_id, v_count, v_agg_id;

    RETURN v_count;
END;
$FN$;

COMMENT ON FUNCTION stewards.spawn_children(uuid) IS
'Batch J.2: decompose-fanout spawn mechanism. Reads stage_results.decompose.output manifest, inserts N children (each dispatched immediately) + 1 aggregator child (status=pending until siblings done). Returns count of regular children.';


-- ---------------------------------------------------------------------
-- 5. Extend on_maturity_verified with the two J.2 branches.
-- ---------------------------------------------------------------------
-- Branch A: parent of fan-out reaches verified -> spawn_children.
-- Branch B: child of any fan-out reaches verified -> count unfinished
--           siblings; if 0, dispatch the aggregator sibling.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.on_maturity_verified()
RETURNS trigger LANGUAGE plpgsql AS $FN$
DECLARE
    v_pipeline      stewards.pipelines%ROWTYPE;
    v_sabbath       boolean;
    v_auto_mat      boolean;
    v_pwid          bigint;
    v_dispatch_id   bigint;
    v_proposed_n    int;
    v_rendered      text;
    v_agent_ok      boolean;
    v_spawn_n       int;
    v_unfinished    int;
    v_agg_id        uuid;
    v_agg_wq        bigint;
BEGIN
    IF NEW.maturity <> 'verified' OR OLD.maturity = 'verified' THEN
        RETURN NEW;
    END IF;

    SELECT * INTO v_pipeline FROM stewards.pipelines WHERE family = NEW.pipeline_family;
    IF v_pipeline.family IS NULL THEN
        RAISE NOTICE 'on_maturity_verified: pipeline % not found', NEW.pipeline_family;
        RETURN NEW;
    END IF;

    v_sabbath := COALESCE(NEW.sabbath_enabled, v_pipeline.sabbath_enabled);
    IF v_sabbath AND NEW.sabbath_completed_at IS NULL THEN
        BEGIN
            v_dispatch_id := stewards.sabbath_dispatch(NEW.id);
            RAISE NOTICE 'on_maturity_verified: sabbath_dispatch work_id=% for work_item=%',
                v_dispatch_id, NEW.id;
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'on_maturity_verified: sabbath_dispatch failed: %', SQLERRM;
        END;
    END IF;

    -- i4 (Batch I.1): agent-proposal source_type routing.
    IF NEW.pipeline_family = 'agent-proposal' AND NEW.agent_proposal_applied_at IS NULL THEN
        BEGIN
            v_agent_ok := stewards.apply_agent_proposal(NEW.id);
            IF v_agent_ok THEN
                SELECT file_destination INTO NEW.file_destination
                  FROM stewards.work_items WHERE id = NEW.id;
            ELSE
                RAISE NOTICE 'on_maturity_verified: apply_agent_proposal returned false for work_item=%; skipping file enqueue',
                    NEW.id;
                RETURN NEW;
            END IF;
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'on_maturity_verified: apply_agent_proposal raised: %', SQLERRM;
            RETURN NEW;
        END;
    END IF;

    -- J.2 (Batch J): decompose-fanout parent reached verified -> spawn children.
    -- Only fires when this is a TOP-LEVEL fan-out parent (no parent of its own).
    -- A nested fan-out can still work: its own decompose stage's verification
    -- triggers spawn for ITS children.
    IF NEW.pipeline_family = 'decompose-fanout' THEN
        BEGIN
            v_spawn_n := stewards.spawn_children(NEW.id);
            RAISE NOTICE 'on_maturity_verified: spawn_children parent=% spawned=%',
                NEW.id, v_spawn_n;
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'on_maturity_verified: spawn_children failed: %', SQLERRM;
        END;
    END IF;

    -- D-H6.3 + D-H6.4 + H.3-followup-2: auto-materialize.
    v_auto_mat := COALESCE(NEW.auto_materialize_enabled, v_pipeline.auto_materialize_on_verified);
    IF v_auto_mat AND NEW.file_enqueued_at IS NULL THEN
        IF NEW.file_destination IS NULL AND v_pipeline.file_destination_template IS NOT NULL THEN
            BEGIN
                v_rendered := stewards.render_file_destination(NEW.id);
                IF v_rendered IS NOT NULL THEN
                    UPDATE stewards.work_items
                       SET file_destination = v_rendered
                     WHERE id = NEW.id;
                    NEW.file_destination := v_rendered;
                    RAISE NOTICE 'on_maturity_verified: auto-rendered file_destination=% for work_item=%',
                        v_rendered, NEW.id;
                END IF;
            EXCEPTION WHEN OTHERS THEN
                RAISE NOTICE 'on_maturity_verified: render_file_destination failed: %', SQLERRM;
            END;
        END IF;

        IF NEW.file_destination IS NOT NULL THEN
            BEGIN
                v_pwid := stewards.enqueue_work_item_file(NEW.id, 'auto_materialize_on_verified');
                RAISE NOTICE 'on_maturity_verified: enqueue_work_item_file pwid=% for work_item=%',
                    v_pwid, NEW.id;
            EXCEPTION WHEN OTHERS THEN
                RAISE NOTICE 'on_maturity_verified: enqueue_work_item_file failed: %', SQLERRM;
            END;
        END IF;
    END IF;

    -- H.3.5: planning pipeline propagation
    IF NEW.pipeline_family = 'planning' THEN
        BEGIN
            v_proposed_n := stewards.enqueue_proposed_work_items(NEW.id);
            RAISE NOTICE 'on_maturity_verified: enqueue_proposed_work_items inserted=% for work_item=%',
                v_proposed_n, NEW.id;
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'on_maturity_verified: enqueue_proposed_work_items failed: %', SQLERRM;
        END;
    END IF;

    -- J.2 (Batch J): child of a fan-out verified -> count siblings;
    -- if all (non-aggregator) siblings done, dispatch the aggregator.
    IF NEW.parent_work_item_id IS NOT NULL
       AND NEW.pipeline_family <> 'aggregate-children' THEN
        BEGIN
            SELECT COUNT(*) INTO v_unfinished
              FROM stewards.work_items
             WHERE parent_work_item_id = NEW.parent_work_item_id
               AND id <> NEW.id
               AND pipeline_family <> 'aggregate-children'
               AND maturity <> 'verified'
               AND status NOT IN ('cancelled', 'failed');

            IF v_unfinished = 0 THEN
                SELECT id INTO v_agg_id
                  FROM stewards.work_items
                 WHERE parent_work_item_id = NEW.parent_work_item_id
                   AND pipeline_family = 'aggregate-children'
                   AND status = 'pending'
                 LIMIT 1;

                IF v_agg_id IS NOT NULL THEN
                    v_agg_wq := stewards.work_item_dispatch_stage(v_agg_id, NULL);
                    RAISE NOTICE 'on_maturity_verified: aggregator % dispatched wq=% (all % siblings of % done)',
                        v_agg_id, v_agg_wq,
                        (SELECT COUNT(*) FROM stewards.work_items
                          WHERE parent_work_item_id = NEW.parent_work_item_id
                            AND pipeline_family <> 'aggregate-children'),
                        NEW.parent_work_item_id;
                END IF;
            END IF;
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'on_maturity_verified: aggregator-dispatch-check failed: %', SQLERRM;
        END;
    END IF;

    RETURN NEW;
END;
$FN$;

COMMENT ON FUNCTION stewards.on_maturity_verified() IS
'j1 (Batch J.2): adds decompose-fanout spawn branch + aggregator-dispatch branch. Spawn fires when fan-out parent reaches verified; aggregator-dispatch fires when child of a fan-out verifies and no siblings remain unfinished. Existing sabbath / agent-proposal / auto-materialize / planning paths preserved.';


-- =====================================================================
-- End of j1-fanout-machinery.sql
-- =====================================================================
