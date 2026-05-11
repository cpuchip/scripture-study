-- =====================================================================
-- Batch G.4.5 — Lesson + resolution file-write producer hooks
--
-- Two AFTER UPDATE triggers + two producer functions wire lessons and
-- resolutions into stewards.pending_file_writes when their promoted_to
-- column transitions from NULL to a path.
--
-- Lessons (Phase D.7 ratify flow): "Approve & promote → .mind/principles.md"
--   button in /lessons UI calls /api/lessons/ratify which UPDATE
--   stewards.lessons SET ratified_at + ratified_by + promoted_to. The new
--   trigger detects the promoted_to transition and INSERTs pending_file_writes
--   with append mode + dated section header.
--
-- Resolutions (Phase F.7 accept flow): resolve_council('accept', text,
--   destination='study'|'decisions', ...) UPDATEs stewards.resolutions
--   SET resolved_by + text + promoted_to. The trigger detects the
--   transition and INSERTs pending_file_writes with create mode
--   (resolutions land in their own file, not appended).
--
-- Both producer functions use existing schema; this is pure wiring.
-- =====================================================================

-- ---------------------------------------------------------------------
-- (1) enqueue_lesson_file — fired by the lessons AFTER UPDATE trigger
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.enqueue_lesson_file(p_lesson_id bigint)
RETURNS bigint
LANGUAGE plpgsql AS $func$
DECLARE
    v_lesson stewards.lessons%ROWTYPE;
    v_wi     stewards.work_items%ROWTYPE;
    v_pwid   bigint;
    v_header text;
    v_content text;
BEGIN
    SELECT * INTO v_lesson FROM stewards.lessons WHERE id = p_lesson_id;
    IF v_lesson.id IS NULL OR v_lesson.promoted_to IS NULL THEN
        RETURN NULL;
    END IF;

    SELECT * INTO v_wi FROM stewards.work_items WHERE id = v_lesson.work_item_id;

    -- Build a dated section header so the .mind file stays browsable
    -- when entries accumulate.
    v_header := format(E'\n\n## %s — %s (%s)\n',
        to_char(coalesce(v_lesson.ratified_at, now()), 'YYYY-MM-DD'),
        v_lesson.kind,
        coalesce(v_wi.slug, v_lesson.work_item_id::text));

    v_content := v_header || v_lesson.content || E'\n';

    INSERT INTO stewards.pending_file_writes
        (requested_by, target_path, write_mode, content, source_id, source_kind)
    VALUES
        ('lesson_promote', v_lesson.promoted_to, 'append', v_content,
         v_lesson.id::text, 'lesson')
    RETURNING id INTO v_pwid;

    RETURN v_pwid;
END;
$func$;

COMMENT ON FUNCTION stewards.enqueue_lesson_file(bigint) IS
'Batch G.4.5: queue a pending_file_writes row (append mode) for a ratified+promoted lesson. Dated section header keeps .mind/principles.md + .mind/decisions.md browsable as entries accumulate.';

-- ---------------------------------------------------------------------
-- (2) enqueue_resolution_file — fired by the resolutions AFTER UPDATE
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.enqueue_resolution_file(p_resolution_id uuid)
RETURNS bigint
LANGUAGE plpgsql AS $func$
DECLARE
    v_res stewards.resolutions%ROWTYPE;
    v_council stewards.councils%ROWTYPE;
    v_pwid bigint;
    v_content text;
    v_write_mode text;
BEGIN
    SELECT * INTO v_res FROM stewards.resolutions WHERE id = p_resolution_id;
    IF v_res.id IS NULL OR v_res.promoted_to IS NULL THEN
        RETURN NULL;
    END IF;

    SELECT * INTO v_council FROM stewards.councils WHERE id = v_res.council_id;

    -- For 'study/<id>.md' style paths: create mode (one file per resolution).
    -- For '.mind/decisions.md': append mode (decisions accumulate).
    IF v_res.promoted_to LIKE '.mind/%' THEN
        v_write_mode := 'append';
        v_content := format(
            E'\n\n## %s — Council resolution: %s\n\n%s\n',
            to_char(coalesce(v_res.resolved_at, now()), 'YYYY-MM-DD'),
            coalesce(v_council.binding_question, '(no binding question)'),
            v_res.text);
    ELSE
        v_write_mode := 'create';
        v_content := format(
            E'# Council resolution\n\n**Binding question:** %s\n**Resolved by:** %s\n**Resolved at:** %s\n\n---\n\n%s\n',
            coalesce(v_council.binding_question, '(no binding question)'),
            v_res.resolved_by,
            to_char(coalesce(v_res.resolved_at, now()), 'YYYY-MM-DD HH24:MI'),
            v_res.text);
    END IF;

    INSERT INTO stewards.pending_file_writes
        (requested_by, target_path, write_mode, content, source_id, source_kind)
    VALUES
        ('council_resolve', v_res.promoted_to, v_write_mode, v_content,
         v_res.id::text, 'resolution')
    RETURNING id INTO v_pwid;

    RETURN v_pwid;
END;
$func$;

COMMENT ON FUNCTION stewards.enqueue_resolution_file(uuid) IS
'Batch G.4.5: queue a pending_file_writes row for an accepted council resolution. Paths under .mind/ use append mode + dated header; study/<id>.md paths use create mode + full document frontmatter.';

-- ---------------------------------------------------------------------
-- (3) AFTER UPDATE trigger on lessons — fires when promoted_to transitions
--     from NULL to a path
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.lessons_promoted_to_trigger()
RETURNS trigger
LANGUAGE plpgsql AS $func$
BEGIN
    IF NEW.promoted_to IS NOT NULL
       AND (OLD.promoted_to IS NULL OR OLD.promoted_to <> NEW.promoted_to) THEN
        BEGIN
            PERFORM stewards.enqueue_lesson_file(NEW.id);
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'enqueue_lesson_file failed for lesson %: %', NEW.id, SQLERRM;
        END;
    END IF;
    RETURN NEW;
END;
$func$;

DROP TRIGGER IF EXISTS lessons_promoted_to_au ON stewards.lessons;
CREATE TRIGGER lessons_promoted_to_au
    AFTER UPDATE OF promoted_to ON stewards.lessons
    FOR EACH ROW
    EXECUTE FUNCTION stewards.lessons_promoted_to_trigger();

COMMENT ON FUNCTION stewards.lessons_promoted_to_trigger() IS
'Batch G.4.5: fires enqueue_lesson_file when a lesson''s promoted_to column transitions from NULL to a path (i.e., a human clicked "Approve & promote → .mind/X.md"). Errors swallowed via NOTICE so the original ratify UPDATE still succeeds.';

-- ---------------------------------------------------------------------
-- (4) AFTER UPDATE trigger on resolutions — same pattern
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.resolutions_promoted_to_trigger()
RETURNS trigger
LANGUAGE plpgsql AS $func$
BEGIN
    IF NEW.promoted_to IS NOT NULL
       AND (OLD.promoted_to IS NULL OR OLD.promoted_to <> NEW.promoted_to) THEN
        BEGIN
            PERFORM stewards.enqueue_resolution_file(NEW.id);
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'enqueue_resolution_file failed for resolution %: %', NEW.id, SQLERRM;
        END;
    END IF;
    RETURN NEW;
END;
$func$;

DROP TRIGGER IF EXISTS resolutions_promoted_to_au ON stewards.resolutions;
CREATE TRIGGER resolutions_promoted_to_au
    AFTER UPDATE OF promoted_to ON stewards.resolutions
    FOR EACH ROW
    EXECUTE FUNCTION stewards.resolutions_promoted_to_trigger();

COMMENT ON FUNCTION stewards.resolutions_promoted_to_trigger() IS
'Batch G.4.5: fires enqueue_resolution_file when a council resolution''s promoted_to transitions from NULL to a path (i.e., bishop accepted with destination=study|decisions). Errors swallowed via NOTICE.';
