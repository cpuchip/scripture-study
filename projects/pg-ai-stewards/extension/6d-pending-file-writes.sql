-- =====================================================================
-- Batch G.4.1 — pending_file_writes schema + producer + columns
--
-- First substrate-side file-write mechanism. Per D-G ratifications:
--   D-G1: pipelines.file_destination_template is UI suggestion only
--   D-G2: materialization via explicit "Materialize now" button
--   D-G3: lessons + resolutions keep their separate promoted_to
--
-- Four schema pieces:
--   1. pipelines.file_destination_template (text NULL) — UI prefill
--   2. pipelines.file_content_jsonpath (text NULL) — override the
--      default convention (stage_results.<final_stage>.output)
--   3. work_items.file_destination (text NULL) — per-work_item decision
--   4. work_items.materialized_at (timestamptz NULL) — set when queued
--   5. stewards.pending_file_writes table — the queue
--
-- One function:
--   stewards.enqueue_work_item_file(work_item_id) — checks
--     work_items.file_destination; if non-NULL, INSERTs a pending row
--     with rendered path + extracted content; returns the pending row id
--     or NULL (DB-only, no-op).
-- =====================================================================

-- ---------------------------------------------------------------------
-- (1) pipelines.file_destination_template + file_content_jsonpath
-- ---------------------------------------------------------------------

ALTER TABLE stewards.pipelines
    ADD COLUMN IF NOT EXISTS file_destination_template text,
    ADD COLUMN IF NOT EXISTS file_content_jsonpath     text;

COMMENT ON COLUMN stewards.pipelines.file_destination_template IS
'Batch G.4 (2026-05-11): optional UI prefill suggestion for work_items file destination. Template supports <slug> and <id>. UI prefills the field on NewWork; human can change or unset. NOT enforced (D-G1).';
COMMENT ON COLUMN stewards.pipelines.file_content_jsonpath IS
'Batch G.4: jsonpath override for extracting file content from work_items.stage_results. NULL = convention (stage_results.<final_stage>.output via pipeline_first_stage_name + walk). Set per-pipeline when the output lives elsewhere (e.g. a `final` aggregator stage).';

-- Seed suggestion for study-write — substrate-promoted studies typically
-- want to land at study/substrate--<slug>.md (matches the existing
-- work_item_promote_to_study slug convention).
UPDATE stewards.pipelines
   SET file_destination_template = 'study/substrate--<slug>.md'
 WHERE family IN ('study-write', 'study-write-qwen')
   AND file_destination_template IS NULL;

-- ---------------------------------------------------------------------
-- (2) work_items.file_destination + materialized_at
-- ---------------------------------------------------------------------

ALTER TABLE stewards.work_items
    ADD COLUMN IF NOT EXISTS file_destination text,
    ADD COLUMN IF NOT EXISTS materialized_at  timestamptz;

COMMENT ON COLUMN stewards.work_items.file_destination IS
'Batch G.4 (2026-05-11): NULL = DB-only (default). Path = will materialize there on explicit "Materialize now" gesture (D-G2). Settable at create time (NewWork prefills from pipeline template) or after the fact via WorkItemDetail.';
COMMENT ON COLUMN stewards.work_items.materialized_at IS
'Batch G.4: timestamp when enqueue_work_item_file successfully queued a pending_file_writes row. NULL = not yet materialized (or DB-only).';

-- ---------------------------------------------------------------------
-- (3) stewards.pending_file_writes table
-- ---------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS stewards.pending_file_writes (
    id              bigserial PRIMARY KEY,
    requested_at    timestamptz NOT NULL DEFAULT now(),
    requested_by    text NOT NULL,
    target_path     text NOT NULL,
    write_mode      text NOT NULL CHECK (write_mode IN ('append', 'create')),
    content         text NOT NULL,
    source_id       text,
    source_kind     text,
    materialized_at timestamptz,
    materialized_by text
);

CREATE INDEX IF NOT EXISTS pending_file_writes_unmaterialized
    ON stewards.pending_file_writes (requested_at)
    WHERE materialized_at IS NULL;

CREATE INDEX IF NOT EXISTS pending_file_writes_source
    ON stewards.pending_file_writes (source_kind, source_id);

COMMENT ON TABLE stewards.pending_file_writes IS
'Batch G.4 (2026-05-11): first substrate-side file-write mechanism. Producer hooks (enqueue_work_item_file, apply_lesson_ratify, resolve_council) INSERT rows; stewards-cli materialize-writes consumes them. Substrate stays FS-stateless — actual file I/O happens in the Go CLI.';

-- ---------------------------------------------------------------------
-- (4) Helper: render template placeholders in a path
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.render_file_path_template(
    p_template text,
    p_slug     text,
    p_id       uuid
) RETURNS text
LANGUAGE plpgsql IMMUTABLE AS $func$
BEGIN
    IF p_template IS NULL THEN
        RETURN NULL;
    END IF;
    RETURN replace(replace(p_template,
        '<slug>', coalesce(p_slug, p_id::text)),
        '<id>',   p_id::text);
END;
$func$;

COMMENT ON FUNCTION stewards.render_file_path_template(text, text, uuid) IS
'Batch G.4: substitute <slug> and <id> placeholders in a file path template. Used by NewWork prefill + enqueue_work_item_file.';

-- ---------------------------------------------------------------------
-- (5) Helper: extract file content from work_item stage_results
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.extract_work_item_file_content(
    p_work_item_id uuid
) RETURNS text
LANGUAGE plpgsql STABLE AS $func$
DECLARE
    v_wi       stewards.work_items%ROWTYPE;
    v_pipeline stewards.pipelines%ROWTYPE;
    v_path     text;
    v_content  text;
    v_final_stage text;
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN RETURN NULL; END IF;

    SELECT * INTO v_pipeline FROM stewards.pipelines WHERE family = v_wi.pipeline_family;
    IF v_pipeline.family IS NULL THEN RETURN NULL; END IF;

    IF v_pipeline.file_content_jsonpath IS NOT NULL THEN
        -- Pipeline overrides the convention
        v_path := v_pipeline.file_content_jsonpath;
    ELSE
        -- Convention: stage_results.<final_stage>.output. Final stage =
        -- last stage in pipelines.stages where next IS NULL.
        SELECT s->>'name' INTO v_final_stage
          FROM jsonb_array_elements(v_pipeline.stages) s
         WHERE s->>'next' IS NULL OR s->'next' = 'null'::jsonb
         LIMIT 1;
        IF v_final_stage IS NULL THEN RETURN NULL; END IF;
        v_path := format('stage_results.%s.output', v_final_stage);
    END IF;

    -- v_path is now a jsonpath-ish dotted path. Resolve via #>>.
    -- For convention "stage_results.review.output": split, then walk.
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
        -- Final step: cast to text
        IF jsonb_typeof(v_traversed) = 'string' THEN
            v_content := v_traversed #>> '{}';
        ELSE
            v_content := v_traversed::text;
        END IF;
    END;

    RETURN v_content;
END;
$func$;

COMMENT ON FUNCTION stewards.extract_work_item_file_content(uuid) IS
'Batch G.4: pull content for the file write from work_items.stage_results. Honors pipelines.file_content_jsonpath override; defaults to stage_results.<final_stage>.output.';

-- ---------------------------------------------------------------------
-- (6) enqueue_work_item_file — the main producer hook
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.enqueue_work_item_file(
    p_work_item_id uuid,
    p_requested_by text DEFAULT 'work_item'
) RETURNS bigint
LANGUAGE plpgsql AS $func$
DECLARE
    v_wi      stewards.work_items%ROWTYPE;
    v_path    text;
    v_content text;
    v_pwid    bigint;
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN
        RAISE EXCEPTION 'enqueue_work_item_file: work_item % not found', p_work_item_id;
    END IF;

    -- DB-only by default — no destination set means no file write.
    IF v_wi.file_destination IS NULL THEN
        RETURN NULL;
    END IF;

    -- Render <slug> / <id> placeholders in case the destination is
    -- still a template (NewWork may store templates verbatim).
    v_path := stewards.render_file_path_template(
        v_wi.file_destination, v_wi.slug, v_wi.id);
    IF v_path IS NULL OR length(trim(v_path)) = 0 THEN
        RAISE EXCEPTION 'enqueue_work_item_file: rendered path is empty for work_item %', p_work_item_id;
    END IF;

    v_content := stewards.extract_work_item_file_content(p_work_item_id);
    IF v_content IS NULL OR length(v_content) = 0 THEN
        RAISE EXCEPTION 'enqueue_work_item_file: extracted content is empty for work_item % (file path %)',
            p_work_item_id, v_path;
    END IF;

    INSERT INTO stewards.pending_file_writes
        (requested_by, target_path, write_mode, content, source_id, source_kind)
    VALUES
        (p_requested_by, v_path, 'create', v_content,
         p_work_item_id::text, 'work_item')
    RETURNING id INTO v_pwid;

    UPDATE stewards.work_items
       SET materialized_at = now()
     WHERE id = p_work_item_id;

    RETURN v_pwid;
END;
$func$;

COMMENT ON FUNCTION stewards.enqueue_work_item_file(uuid, text) IS
'Batch G.4 (2026-05-11): the universal work_item file-write producer. Checks file_destination; if NULL returns NULL (no-op). Otherwise renders the path + extracts content via extract_work_item_file_content + INSERTs pending_file_writes + sets work_items.materialized_at. Idempotent only via the materialized_at guard in callers (this function does NOT check materialized_at — callers may re-enqueue intentionally).';
