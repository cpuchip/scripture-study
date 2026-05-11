-- H.1.5d materialize: extract the verified study from the work_item and
-- queue it as a pending_file_write. Then `stewards-cli materialize-writes`
-- (or the pre-commit hook) lands it at research/ai-tools-weekly-2026-05-11.md.
--
-- Stripping the "REVIEW: passes\n\n" prefix from the review output so
-- the materialized file is the clean study, not the review verdict.

DO $$
DECLARE
    v_wi_id      uuid;
    v_review_out text;
    v_clean_md   text;
    v_pwid       bigint;
BEGIN
    SELECT id INTO v_wi_id FROM stewards.work_items
     WHERE slug = 'ai-tools-weekly-2026-05-11-v3';

    SELECT stage_results->'review'->>'output' INTO v_review_out
      FROM stewards.work_items WHERE id = v_wi_id;

    -- Strip the leading "REVIEW: passes\n\n" if present (verbatim per template).
    IF v_review_out LIKE 'REVIEW: passes%' THEN
        v_clean_md := regexp_replace(v_review_out, E'^REVIEW: passes\\s*\\n\\s*', '');
    ELSIF v_review_out LIKE 'REVIEW: revised%' THEN
        -- Strip "REVIEW: revised\n\n" and any trailing notes block
        v_clean_md := regexp_replace(v_review_out, E'^REVIEW: revised\\s*\\n\\s*', '');
    ELSE
        v_clean_md := v_review_out;
    END IF;

    INSERT INTO stewards.pending_file_writes
        (requested_by, target_path, write_mode, content, source_id, source_kind)
    VALUES (
        'h1-5d-manual',
        'research/ai-tools-weekly-2026-05-11.md',
        'create',
        v_clean_md,
        v_wi_id::text,
        'work_item'
    )
    RETURNING id INTO v_pwid;

    -- Mark the work_item's materialized_at + studies.file_path (mirrors
    -- what stewards-cli materialize-writes would do for source_kind=work_item)
    UPDATE stewards.work_items SET materialized_at = now() WHERE id = v_wi_id;

    RAISE NOTICE 'pending_file_write id=% queued; content length=%; target=%',
        v_pwid, length(v_clean_md), 'research/ai-tools-weekly-2026-05-11.md';
END
$$;

SELECT id, target_path, write_mode, length(content) AS content_len, source_kind, requested_by, materialized_at
  FROM stewards.pending_file_writes
 WHERE target_path = 'research/ai-tools-weekly-2026-05-11.md'
 ORDER BY id DESC LIMIT 3;
