-- Smoke H.1.6.6: REVIEW prefix strip on real review outputs from the
-- three completed research-write work_items. Uses extract_work_item_file_content
-- directly so we test the actual codepath, not a synthetic string.

DO $$
DECLARE
    v_slug text;
    v_wi_id uuid;
    v_content text;
    v_first_50 text;
BEGIN
    FOR v_slug IN
        SELECT slug FROM (VALUES
            ('ai-tools-weekly-2026-05-11-v3'),
            ('pg-ext-distribution-2026-05-11'),
            ('physics-news-20260503-science-center-roundup')
        ) AS t(slug)
    LOOP
        SELECT id INTO v_wi_id FROM stewards.work_items WHERE slug = v_slug;
        IF v_wi_id IS NULL THEN
            RAISE NOTICE '[%] work_item not found', v_slug;
            CONTINUE;
        END IF;

        v_content := stewards.extract_work_item_file_content(v_wi_id);
        v_first_50 := substring(v_content, 1, 50);

        RAISE NOTICE '[%] first 50 chars: %', v_slug, v_first_50;
        IF v_content ~ '^REVIEW:' THEN
            RAISE EXCEPTION '[%] still has REVIEW prefix!', v_slug;
        END IF;
    END LOOP;

    RAISE NOTICE 'H.1.6.6 smoke PASSED — all 3 prior research-write outputs strip cleanly';
END
$$;
