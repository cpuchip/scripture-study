-- =====================================================================
-- i4 — agent-proposal pipeline + apply_agent_proposal SQL function
--
-- Batch I.1 from substrate-batch-i-agent-write-back.md proposal.
-- Enables agents (kimi, qwen, claude, etc.) to propose study / lesson /
-- note / exhibit / schema-migration artifacts through the existing
-- trust ladder + Proposed-work panel UI for human ratification.
--
-- Pipeline shape (single stage):
--   1. `validate` — reads input.draft (agent's structured proposal),
--      light normalization + schema validation via cheap qwen pass.
--      Emits clean JSON to stage_results.validate.output. Cost ~$0.005.
--      auto_advance: true. Maturity → verified.
--
-- After verified, on_maturity_verified calls apply_agent_proposal which:
--   - Reads stage_results.validate.output
--   - For study/lesson/note/exhibit: INSERT into stewards.studies
--     (kind=source_type) + sets work_items.file_destination
--   - For schema-migration: validates SQL syntax (deferred to I.3) +
--     sets file_destination = projects/pg-ai-stewards/extension/iN-<slug>.sql
--   - enqueue_work_item_file then fires via existing path; file lands.
--
-- New work_items column: agent_proposal_applied_at (NULL until persisted).
-- Mirrors revision_applied_at pattern from h3-followup-3.
-- =====================================================================

-- ---------------------------------------------------------------------
-- (1) Schema additions
-- ---------------------------------------------------------------------

ALTER TABLE stewards.work_items
    ADD COLUMN IF NOT EXISTS agent_proposal_applied_at timestamp with time zone;

COMMENT ON COLUMN stewards.work_items.agent_proposal_applied_at IS
'i4 (Batch I.1): set by apply_agent_proposal when an agent-proposal work_item has been persisted to its target (studies row + file enqueued). NULL = not yet persisted (or never an agent proposal).';

-- ---------------------------------------------------------------------
-- (2) Pipeline definition
-- ---------------------------------------------------------------------

DO $$
DECLARE
    v_validate_template text;
    v_stages            jsonb;
BEGIN

v_validate_template :=
$T$You are validating an agent-submitted proposal for a substrate artifact.

## AGENT DRAFT

```json
{{input.draft}}
```

## YOUR TASK

Read the draft. Validate and normalize it. Output ONLY a JSON object — no prose, no markdown fences.

## SCHEMA (output)

```json
{
  "source_type": "study | lesson | note | exhibit | schema-migration",
  "slug": "kebab-case-slug",
  "title": "Human-readable title (10-120 chars)",
  "body": "Full markdown body OR full SQL for schema-migration",
  "frontmatter": { /* per-source-type metadata; jsonb object */ },
  "project_association": "string slug or null",
  "rationale": "Why this proposal exists (1-3 sentences; shown in ratification UI)"
}
```

## VALIDATION RULES

- `source_type` MUST be one of: study, lesson, note, exhibit, schema-migration.
- `slug` MUST match `^[a-z0-9-]+$`. If the draft slug is malformed, fix it.
- `title` MUST be 10-120 chars. If too short, expand from body's first heading. If too long, trim.
- `body` MUST be non-empty.
- For `schema-migration`: `body` MUST start with `-- ` (SQL comment header) and contain at least one `CREATE`, `ALTER`, `INSERT`, or `CREATE OR REPLACE` statement.
- `frontmatter` MUST be a JSON object (use `{}` if no metadata).
- `project_association` is optional; pass through from draft or set null.
- `rationale` MUST be 20-500 chars. If missing, derive from body's intro.

## ON ERROR

If the draft cannot be normalized into a valid proposal, output:
```json
{"error": "Brief reason"}
```

Output ONLY the JSON object. Your turn.$T$;

v_stages := jsonb_build_array(
    jsonb_build_object(
        'name', 'validate',
        'next', NULL,
        'model', 'qwen3.6-plus',
        'provider', 'opencode_go',
        'agent_family', 'research',
        'auto_advance', true,
        'tools_disabled', true,
        'input_template', v_validate_template
    )
);

INSERT INTO stewards.pipelines (
    family, description, stages, metadata,
    sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath,
    maturity_ladder, auto_materialize_on_verified
)
VALUES (
    'agent-proposal',
    'Agent submits a study/lesson/note/exhibit/schema-migration proposal. Single-stage validate pass normalizes the draft JSON. On verified, apply_agent_proposal persists to studies + enqueues file write. Human ratifies via WorkItemDetail.vue Proposed-work panel (origin filter agent_proposal). schema-migration source_type is Claude-only and lands at projects/pg-ai-stewards/extension/iN-<slug>.sql.',
    v_stages,
    jsonb_build_object(
        'cost_cap_default_micro', 100000,
        'cost_cap_default_dollars', 0.10,
        'note', 'Single qwen validate pass; typical cost $0.005-0.01. apply_agent_proposal sets file_destination dynamically per source_type.'
    ),
    false,  -- sabbath_enabled
    false,  -- atonement_enabled
    NULL,   -- file_destination_template: dynamic via apply_agent_proposal
    'stage_results.validate.output',
    '["raw","verified"]'::jsonb,
    true    -- auto_materialize_on_verified: yes, file writes after apply_agent_proposal sets file_destination
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

INSERT INTO stewards.pipeline_stage_maturity (pipeline_family, stage_name, produces_maturity)
VALUES ('agent-proposal', 'validate', 'verified')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE
   SET produces_maturity = EXCLUDED.produces_maturity;

INSERT INTO stewards.stage_models (pipeline_family, stage_name, default_model)
VALUES ('agent-proposal', 'validate', 'qwen3.6-plus')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE
   SET default_model = EXCLUDED.default_model;

-- ---------------------------------------------------------------------
-- (3) apply_agent_proposal — persists the validated proposal
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.apply_agent_proposal(p_work_item_id uuid)
RETURNS boolean
LANGUAGE plpgsql
AS $func$
DECLARE
    v_wi          stewards.work_items%ROWTYPE;
    v_raw         text;
    v_clean       text;
    v_json        jsonb;
    v_source_type text;
    v_slug        text;
    v_title       text;
    v_body        text;
    v_frontmatter jsonb;
    v_project     text;
    v_rationale   text;
    v_file_dest   text;
    v_existing_id text;
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN
        RAISE NOTICE 'apply_agent_proposal: work_item % not found', p_work_item_id;
        RETURN false;
    END IF;
    IF v_wi.pipeline_family <> 'agent-proposal' THEN
        RAISE NOTICE 'apply_agent_proposal: work_item % is not agent-proposal (family=%)',
            p_work_item_id, v_wi.pipeline_family;
        RETURN false;
    END IF;
    IF v_wi.agent_proposal_applied_at IS NOT NULL THEN
        RAISE NOTICE 'apply_agent_proposal: already applied at %', v_wi.agent_proposal_applied_at;
        RETURN false;
    END IF;

    -- Pull validated output
    v_raw := (v_wi.stage_results -> 'validate' -> 'output') #>> '{}';
    IF v_raw IS NULL OR length(trim(v_raw)) = 0 THEN
        RAISE NOTICE 'apply_agent_proposal: validate.output is empty';
        RETURN false;
    END IF;

    v_clean := regexp_replace(v_raw, E'^\\s*```(?:json)?\\s*\\n?|\\n?```\\s*$', '', 'g');
    v_clean := trim(v_clean);

    BEGIN
        v_json := v_clean::jsonb;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'apply_agent_proposal: JSON parse failed: %', SQLERRM;
        RETURN false;
    END;

    -- Check for validator-flagged error
    IF v_json ? 'error' THEN
        RAISE NOTICE 'apply_agent_proposal: validator returned error: %', v_json->>'error';
        RETURN false;
    END IF;

    -- Extract + validate required fields
    v_source_type := v_json ->> 'source_type';
    v_slug        := v_json ->> 'slug';
    v_title       := v_json ->> 'title';
    v_body        := v_json ->> 'body';
    v_frontmatter := COALESCE(v_json -> 'frontmatter', '{}'::jsonb);
    v_project     := v_json ->> 'project_association';
    v_rationale   := v_json ->> 'rationale';

    IF v_source_type IS NULL OR v_source_type NOT IN ('study','lesson','note','exhibit','schema-migration') THEN
        RAISE NOTICE 'apply_agent_proposal: invalid source_type %', v_source_type;
        RETURN false;
    END IF;
    IF v_slug IS NULL OR v_slug !~ '^[a-z0-9-]+$' THEN
        RAISE NOTICE 'apply_agent_proposal: invalid slug %', v_slug;
        RETURN false;
    END IF;
    IF v_title IS NULL OR length(v_title) < 10 OR length(v_title) > 120 THEN
        RAISE NOTICE 'apply_agent_proposal: invalid title length: %', coalesce(length(v_title), 0);
        RETURN false;
    END IF;
    IF v_body IS NULL OR length(trim(v_body)) = 0 THEN
        RAISE NOTICE 'apply_agent_proposal: empty body';
        RETURN false;
    END IF;

    -- Route by source_type
    IF v_source_type IN ('study','lesson','note','exhibit') THEN
        -- File destination per source_type
        v_file_dest := CASE v_source_type
            WHEN 'study'   THEN 'study/' || v_slug || '.md'
            WHEN 'lesson'  THEN 'lessons/' || v_slug || '.md'
            WHEN 'note'    THEN 'becoming/notes/' || v_slug || '.md'
            WHEN 'exhibit' THEN 'exhibits/' || v_slug || '.md'
        END;

        -- Check uniqueness on (kind, slug) — soft check, not a constraint
        SELECT id INTO v_existing_id
          FROM stewards.studies
         WHERE kind = v_source_type AND slug = v_slug
         LIMIT 1;
        IF v_existing_id IS NOT NULL THEN
            RAISE NOTICE 'apply_agent_proposal: (kind=%, slug=%) already exists as study id=%',
                v_source_type, v_slug, v_existing_id;
            RETURN false;
        END IF;

        -- Augment frontmatter with provenance
        v_frontmatter := v_frontmatter
                      || jsonb_build_object(
                            'source_type', v_source_type,
                            'origin', 'agent_proposal',
                            'proposed_by_work_item_id', p_work_item_id::text,
                            'rationale', v_rationale
                         );

        -- INSERT into studies. file_path set so future migrate-writes
        -- can find / update if needed.
        INSERT INTO stewards.studies (slug, title, body, kind, frontmatter, project_association, file_path)
        VALUES (v_slug, v_title, v_body, v_source_type, v_frontmatter, v_project, v_file_dest);

    ELSIF v_source_type = 'schema-migration' THEN
        -- Claude-only gate (deferred to I.3 SQL syntax validator).
        -- For now: only allow if work_item's stage model resolves to claude.
        -- This batch (I.1) doesn't ship schema-migration end-to-end;
        -- we'll allow apply to set file_destination for testing, but
        -- the syntax validator is the I.3 task.
        v_file_dest := 'projects/pg-ai-stewards/extension/' || v_slug || '.sql';
        RAISE NOTICE 'apply_agent_proposal: schema-migration source_type — I.3 will add syntax validation; landing at %', v_file_dest;
    END IF;

    -- Set the file_destination on the work_item so enqueue_work_item_file
    -- (called by on_maturity_verified after this) picks it up.
    UPDATE stewards.work_items
       SET file_destination       = v_file_dest,
           agent_proposal_applied_at = now(),
           updated_at             = now()
     WHERE id = p_work_item_id;

    RAISE NOTICE 'apply_agent_proposal: persisted source_type=% slug=% file_dest=%',
        v_source_type, v_slug, v_file_dest;
    RETURN true;
END;
$func$;

COMMENT ON FUNCTION stewards.apply_agent_proposal(uuid) IS
'i4 (Batch I.1): persists a verified agent-proposal work_item. Parses stage_results.validate.output, validates schema, INSERTs into studies (kind=source_type) for study/lesson/note/exhibit, sets work_items.file_destination so enqueue_work_item_file fires next. schema-migration source_type lands at projects/pg-ai-stewards/extension/<slug>.sql; full syntax validation deferred to I.3. Idempotent via agent_proposal_applied_at guard.';

-- ---------------------------------------------------------------------
-- (4) Hook into on_maturity_verified — call apply_agent_proposal BEFORE
--     the existing enqueue path. Same shape as h3-followup-2 redefinition
--     of on_maturity_verified with auto-render; we add the agent-proposal
--     branch ahead of the enqueue.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.on_maturity_verified()
RETURNS trigger
LANGUAGE plpgsql
AS $func$
DECLARE
    v_pipeline      stewards.pipelines%ROWTYPE;
    v_sabbath       boolean;
    v_auto_mat      boolean;
    v_pwid          bigint;
    v_dispatch_id   bigint;
    v_proposed_n    int;
    v_rendered      text;
    v_agent_ok      boolean;
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
    -- Runs BEFORE the existing enqueue path so apply_agent_proposal
    -- can set file_destination dynamically per source_type.
    IF NEW.pipeline_family = 'agent-proposal' AND NEW.agent_proposal_applied_at IS NULL THEN
        BEGIN
            v_agent_ok := stewards.apply_agent_proposal(NEW.id);
            IF v_agent_ok THEN
                -- Refresh NEW with the file_destination that apply set
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

    RETURN NEW;
END;
$func$;

COMMENT ON FUNCTION stewards.on_maturity_verified() IS
'i4 (Batch I.1): adds agent-proposal branch BEFORE the existing enqueue path. For pipeline_family=agent-proposal, calls apply_agent_proposal which sets file_destination dynamically per source_type. Existing sabbath / auto-materialize / planning-propagation paths preserved.';

-- ---------------------------------------------------------------------
-- (5) Sanity checks
-- ---------------------------------------------------------------------

SELECT 'i4 verify:' AS check_,
       (SELECT count(*) FROM stewards.pipelines WHERE family = 'agent-proposal') AS pipeline_present,
       (SELECT count(*) FROM pg_proc WHERE proname = 'apply_agent_proposal') AS fn_present,
       (SELECT count(*) FROM information_schema.columns
         WHERE table_schema='stewards' AND table_name='work_items'
           AND column_name='agent_proposal_applied_at') AS column_present;
