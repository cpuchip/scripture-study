-- =====================================================================
-- i6 — claude_attested gate for source_type='schema-migration'
--
-- Batch I.3, ratification 2026-05-12: schema-migration source_type
-- requires input.draft.claude_attested = true. Self-claimed (honor
-- system); defense-in-depth for the kimi-trust ratification that says
-- substrate-internal SQL stays Claude-only.
--
-- apply_agent_proposal redefined to check this gate BEFORE persisting.
-- A non-Claude actor that sets claude_attested=true is committing a
-- covenant violation, not bypassing a technical control. We can tighten
-- to model-resolution-based enforcement later if needed.
--
-- Pipeline validate stage prompt updated to document claude_attested
-- field; validate output is preserved but the GATE reads input.draft
-- directly (validate stage cannot promote attestation).
-- =====================================================================

-- ---------------------------------------------------------------------
-- (1) Redefine apply_agent_proposal — adds claude_attested gate
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
    v_claude_attested boolean;
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

    IF v_json ? 'error' THEN
        RAISE NOTICE 'apply_agent_proposal: validator returned error: %', v_json->>'error';
        RETURN false;
    END IF;

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

    -- i6: schema-migration requires claude_attested=true on input.draft
    -- (read from input, not validate.output — validate stage cannot
    -- promote attestation).
    IF v_source_type = 'schema-migration' THEN
        v_claude_attested := COALESCE(
            (v_wi.input -> 'draft' ->> 'claude_attested')::boolean,
            false
        );
        IF v_claude_attested <> true THEN
            RAISE NOTICE 'apply_agent_proposal: schema-migration requires input.draft.claude_attested=true per kimi-trust ratification 2026-05-11; got %',
                v_wi.input -> 'draft' ->> 'claude_attested';
            RETURN false;
        END IF;
    END IF;

    IF v_source_type IN ('study','lesson','note','exhibit') THEN
        v_file_dest := CASE v_source_type
            WHEN 'study'   THEN 'study/' || v_slug || '.md'
            WHEN 'lesson'  THEN 'lessons/' || v_slug || '.md'
            WHEN 'note'    THEN 'becoming/notes/' || v_slug || '.md'
            WHEN 'exhibit' THEN 'exhibits/' || v_slug || '.md'
        END;

        SELECT id INTO v_existing_id
          FROM stewards.studies
         WHERE kind = v_source_type AND slug = v_slug
         LIMIT 1;
        IF v_existing_id IS NOT NULL THEN
            RAISE NOTICE 'apply_agent_proposal: (kind=%, slug=%) already exists as study id=%',
                v_source_type, v_slug, v_existing_id;
            RETURN false;
        END IF;

        v_frontmatter := v_frontmatter
                      || jsonb_build_object(
                            'source_type', v_source_type,
                            'origin', 'agent_proposal',
                            'proposed_by_work_item_id', p_work_item_id::text,
                            'rationale', v_rationale
                         );

        INSERT INTO stewards.studies (slug, title, body, kind, frontmatter, project_association, file_path)
        VALUES (v_slug, v_title, v_body, v_source_type, v_frontmatter, v_project, v_file_dest);

    ELSIF v_source_type = 'schema-migration' THEN
        -- File destination naming: caller's slug is the migration name
        -- (e.g., "iN-add-foo-table"). Land in extension/.
        v_file_dest := 'projects/pg-ai-stewards/extension/' || v_slug || '.sql';

        -- i6: claude_attested already validated above. The CLI-side
        -- validate-sql will do BEGIN/ROLLBACK syntax check before the
        -- file lands on disk (Batch I.3 stewards-cli companion).
        RAISE NOTICE 'apply_agent_proposal: schema-migration with claude_attested=true; file_dest=% (validate-sql will run at materialize time)',
            v_file_dest;
    END IF;

    UPDATE stewards.work_items
       SET file_destination          = v_file_dest,
           agent_proposal_applied_at = now(),
           updated_at                = now()
     WHERE id = p_work_item_id;

    RAISE NOTICE 'apply_agent_proposal: persisted source_type=% slug=% file_dest=%',
        v_source_type, v_slug, v_file_dest;
    RETURN true;
END;
$func$;

COMMENT ON FUNCTION stewards.apply_agent_proposal(uuid) IS
'i6 (Batch I.3): adds claude_attested gate for source_type=schema-migration. Reads input.draft.claude_attested directly (validate stage cannot promote attestation). schema-migration files land at projects/pg-ai-stewards/extension/<slug>.sql; stewards-cli validate-sql runs at materialize time for syntax validation.';

-- ---------------------------------------------------------------------
-- (2) Update validate stage prompt to document claude_attested
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

## SCHEMA-MIGRATION KIMI-TRUST GATE (i6, 2026-05-12)

For `source_type=schema-migration`, the substrate enforces a `claude_attested=true` gate at apply time. This is the kimi-trust ratification: substrate-internal SQL stays Claude-only. The attestation lives on `input.draft.claude_attested` and is NOT promoted by this validate stage. Your output should preserve any draft.claude_attested value verbatim alongside the normalized fields, but the gate check reads from input.draft directly.

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

UPDATE stewards.pipelines
   SET stages = v_stages,
       updated_at = now()
 WHERE family = 'agent-proposal';

END $$;

-- ---------------------------------------------------------------------
-- (3) Sanity check
-- ---------------------------------------------------------------------

SELECT 'i6 verify:' AS check_,
       (SELECT obj_description('stewards.apply_agent_proposal(uuid)'::regprocedure) LIKE '%i6%')
           AS fn_doc_updated,
       (SELECT (stages->0->>'input_template') LIKE '%KIMI-TRUST GATE%'
          FROM stewards.pipelines WHERE family='agent-proposal')
           AS prompt_updated;
