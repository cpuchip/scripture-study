-- =====================================================================
-- i7 — apply_agent_proposal queues pending_file_writes directly with body
--
-- BUG surfaced 2026-05-12 during I.3 smoke (companion to i4 + i6):
--
-- The agent-proposal pipeline sets pipeline.file_content_jsonpath =
-- 'stage_results.validate.output'. This points to the validate stage's
-- output which is a JSON STRING (the normalized proposal):
--
--   { "source_type": "...", "slug": "...", "title": "...",
--     "body": "<the actual file content>", "frontmatter": {...}, ... }
--
-- enqueue_work_item_file calls extract_work_item_file_content which
-- returns this entire JSON string. That JSON then gets written to disk
-- as the file content — i.e., the user's exhibit/study/note/lesson .md
-- file would have been JSON, not markdown. For schema-migration .sql
-- files, this also means the SQL on disk was the wrapping JSON, not the
-- migration body. The bug had not surfaced in I.1 because we never
-- actually materialized to disk (bridge mount is read-only); I.3's
-- validate-sql hook surfaced it by parsing the queued content.
--
-- Fix: apply_agent_proposal now INSERTS into pending_file_writes
-- DIRECTLY with the extracted body as content, then sets
-- work_items.file_enqueued_at = now() so the subsequent enqueue_work_
-- item_file call in on_maturity_verified is a no-op (its IF guard
-- checks file_enqueued_at IS NULL).
--
-- Scoped to agent-proposal pipeline only; other pipelines unchanged.
-- =====================================================================

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
    v_pwid        bigint;
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

    -- i6: schema-migration claude_attested gate
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

    -- Compute file destination per source_type
    v_file_dest := CASE v_source_type
        WHEN 'study'            THEN 'study/' || v_slug || '.md'
        WHEN 'lesson'           THEN 'lessons/' || v_slug || '.md'
        WHEN 'note'             THEN 'becoming/notes/' || v_slug || '.md'
        WHEN 'exhibit'          THEN 'exhibits/' || v_slug || '.md'
        WHEN 'schema-migration' THEN 'projects/pg-ai-stewards/extension/' || v_slug || '.sql'
    END;

    -- Per-source-type DB landing
    IF v_source_type IN ('study','lesson','note','exhibit') THEN
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
        -- No studies row; just the file
        RAISE NOTICE 'apply_agent_proposal: schema-migration; queueing file at %', v_file_dest;
    END IF;

    -- i7: queue pending_file_writes DIRECTLY with the body as content.
    -- Bypasses enqueue_work_item_file's extract_work_item_file_content
    -- which would return the full JSON wrapper (bug from i4).
    INSERT INTO stewards.pending_file_writes
        (requested_by, target_path, write_mode, content, source_id, source_kind)
    VALUES
        ('apply_agent_proposal', v_file_dest, 'create', v_body,
         p_work_item_id::text, 'work_item')
    RETURNING id INTO v_pwid;

    -- Set file_destination AND file_enqueued_at so on_maturity_verified's
    -- subsequent enqueue path becomes a no-op (its IF guard is
    -- 'file_enqueued_at IS NULL').
    UPDATE stewards.work_items
       SET file_destination          = v_file_dest,
           file_enqueued_at          = now(),
           agent_proposal_applied_at = now(),
           updated_at                = now()
     WHERE id = p_work_item_id;

    RAISE NOTICE 'apply_agent_proposal: persisted source_type=% slug=% body_len=% pwid=% file_dest=%',
        v_source_type, v_slug, length(v_body), v_pwid, v_file_dest;
    RETURN true;
END;
$func$;

COMMENT ON FUNCTION stewards.apply_agent_proposal(uuid) IS
'i7 (Batch I.3): queues pending_file_writes DIRECTLY with the validated body as content; bypasses enqueue_work_item_file (which would have returned the full JSON wrapper). Sets work_items.file_enqueued_at = now() so the subsequent on_maturity_verified enqueue path is a no-op. Scoped to agent-proposal pipeline only.';

-- Sanity
SELECT 'i7 verify:' AS check_,
       (SELECT obj_description('stewards.apply_agent_proposal(uuid)'::regprocedure) LIKE '%i7%') AS fn_doc_updated;
