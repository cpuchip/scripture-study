-- Smoke test G.4.5: lesson + resolution → pending_file_writes
DO $$
DECLARE
    v_wi_id uuid;
    v_lesson_id bigint;
    v_council_id uuid;
    v_res_id uuid;
    v_lesson_pw bigint;
    v_res_pw bigint;
BEGIN
    SELECT stewards.work_item_create('study-write',
        '{"binding_question":"smoke G45"}'::jsonb,
        'smoke-g45', 'human', NULL, NULL) INTO v_wi_id;

    INSERT INTO stewards.lessons (work_item_id, kind, content)
    VALUES (v_wi_id, 'principle', 'Triggers should fire when promoted_to is set.')
    RETURNING id INTO v_lesson_id;

    UPDATE stewards.lessons
       SET ratified_at = now(),
           ratified_by = 'smoke-test',
           promoted_to = '.mind/principles.md'
     WHERE id = v_lesson_id;

    SELECT id INTO v_lesson_pw FROM stewards.pending_file_writes
     WHERE source_kind = 'lesson' AND source_id = v_lesson_id::text;
    RAISE NOTICE 'lesson_pw_id=%', v_lesson_pw;

    INSERT INTO stewards.councils (intent_id, binding_question, status, convened_by, bishop)
    VALUES ((SELECT id FROM stewards.intents LIMIT 1),
            'Smoke G.4.5 council?', 'resolved', 'smoke-test', 'smoke-bishop')
    RETURNING id INTO v_council_id;

    INSERT INTO stewards.resolutions (council_id, resolved_by, text)
    VALUES (v_council_id, 'smoke-bishop', 'Synthesized resolution body.')
    RETURNING id INTO v_res_id;

    UPDATE stewards.resolutions SET promoted_to = 'study/smoke-g45.md' WHERE id = v_res_id;

    SELECT id INTO v_res_pw FROM stewards.pending_file_writes
     WHERE source_kind = 'resolution' AND source_id = v_res_id::text;
    RAISE NOTICE 'resolution_pw_id=%', v_res_pw;
END
$$;

SELECT id, source_kind, source_id, target_path, write_mode,
       length(content) AS content_len
  FROM stewards.pending_file_writes
 WHERE source_kind IN ('lesson','resolution')
 ORDER BY id DESC LIMIT 4;
