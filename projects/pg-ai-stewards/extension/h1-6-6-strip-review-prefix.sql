-- =====================================================================
-- Batch H.1.6.6 — strip the "REVIEW: <verdict>" prefix in materialized output
--
-- Three runs have surfaced the same polish issue: research-write's
-- review stage template asks the model to start its output with
-- "REVIEW: passes\n\n<draft>" or "REVIEW: revised\n\n<draft>\n\n
-- Notes on revisions:\n...". When extract_work_item_file_content
-- walks the convention path stage_results.review.output, it returns
-- the verdict line as part of the file. Manual strip on three files
-- now (ai-tools-weekly, pg-ext-distribution, physics-news).
--
-- Fix: regexp_replace the leading verdict line + any blank lines that
-- follow it. Only applied when content comes from the convention
-- (stage_results.<final>.output) path; pipelines that explicitly set
-- file_content_jsonpath own their own conventions and shouldn't be
-- touched.
--
-- The "Notes on revisions" footer on revised drafts is intentionally
-- preserved — it's useful provenance.
-- =====================================================================

CREATE OR REPLACE FUNCTION stewards.extract_work_item_file_content(p_work_item_id uuid)
RETURNS text
LANGUAGE plpgsql
STABLE
AS $func$
DECLARE
    v_wi          stewards.work_items%ROWTYPE;
    v_pipeline    stewards.pipelines%ROWTYPE;
    v_path        text;
    v_content     text;
    v_final_stage text;
    v_used_convention boolean := false;
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN RETURN NULL; END IF;

    SELECT * INTO v_pipeline FROM stewards.pipelines WHERE family = v_wi.pipeline_family;
    IF v_pipeline.family IS NULL THEN RETURN NULL; END IF;

    IF v_pipeline.file_content_jsonpath IS NOT NULL THEN
        v_path := v_pipeline.file_content_jsonpath;
    ELSE
        SELECT s->>'name' INTO v_final_stage
          FROM jsonb_array_elements(v_pipeline.stages) s
         WHERE s->>'next' IS NULL OR s->'next' = 'null'::jsonb
         LIMIT 1;
        IF v_final_stage IS NULL THEN RETURN NULL; END IF;
        v_path := format('stage_results.%s.output', v_final_stage);
        v_used_convention := true;
    END IF;

    DECLARE
        v_parts text[];
        v_traversed jsonb := to_jsonb(v_wi);
    BEGIN
        v_parts := string_to_array(v_path, '.');
        FOR i IN 1..array_length(v_parts, 1) LOOP
            IF v_traversed IS NULL THEN RETURN NULL; END IF;
            v_traversed := v_traversed -> v_parts[i];
        END LOOP;
        IF v_traversed IS NULL THEN RETURN NULL; END IF;
        IF jsonb_typeof(v_traversed) = 'string' THEN
            v_content := v_traversed #>> '{}';
        ELSE
            v_content := v_traversed::text;
        END IF;
    END;

    -- H.1.6.6: strip the substrate-convention review verdict prefix
    -- when content came through the convention path. The review stage
    -- template asks the model to emit "REVIEW: <verdict>\n\n<draft>";
    -- the verdict line is a substrate sentinel, not part of the
    -- published artifact. regexp_replace is a no-op when the pattern
    -- doesn't match, so we apply unconditionally on the convention
    -- path. The leading-anchor + word-character pattern is specific
    -- enough to avoid false positives on legitimate content. Any
    -- trailing "Notes on revisions" section is preserved untouched.
    IF v_used_convention THEN
        v_content := regexp_replace(v_content, E'^REVIEW:\\s+\\w+\\s*\\n+', '');
    END IF;

    RETURN v_content;
END;
$func$;

COMMENT ON FUNCTION stewards.extract_work_item_file_content(uuid) IS
'H.1.6.6 (Batch H): when content comes through the convention path (stage_results.<final>.output, no explicit file_content_jsonpath) and the first line matches the substrate REVIEW verdict pattern, strip that line + following blank line(s). Pipelines that explicitly set file_content_jsonpath own their own conventions and are not affected.';
