-- =====================================================================
-- Batch PE.B.2 — scheduled_pipelines_fire + watchman tick integration + seed
--
-- Builds the dispatcher side of PE-B. The watchman tick (60s, leader-only)
-- already runs via the bgworker; here we extend stewards.watchman_scheduler_fire
-- to also call stewards.scheduled_pipelines_fire(). No bgworker.rs change.
--
-- Scheduled-pipelines dispatch fires regardless of watchman_config.schedule_enabled
-- (the two schedulers are independent). The watchman early-return only short-
-- circuits the watchman pass logic, not the pipelines logic.
--
-- D-PE4 (fire-one-missed) semantics:
--   - now() <= next_due_at + missed_window_hours → dispatch one run
--   - now() >  next_due_at + missed_window_hours → skip missed runs,
--     advance next_due_at to the next future match without dispatching
-- =====================================================================

-- ---------------------------------------------------------------------
-- PE.B.2.a — scheduled_pipelines_fire(): the dispatcher
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.scheduled_pipelines_fire()
RETURNS int
LANGUAGE plpgsql AS $func$
DECLARE
    v_row             stewards.scheduled_pipelines%ROWTYPE;
    v_child_slug      text;
    v_work_item_id    uuid;
    v_now             timestamptz := now();
    v_missed_cutoff   timestamptz;
    v_dispatched      int := 0;
    v_skipped_missed  int := 0;
    v_next_due        timestamptz;
BEGIN
    -- FOR UPDATE SKIP LOCKED keeps multiple leader candidates / multi-
    -- worker invocations from racing. With one leader today the lock
    -- just prevents accidental re-entry mid-tick.
    FOR v_row IN
        SELECT *
          FROM stewards.scheduled_pipelines
         WHERE enabled = true
           AND next_due_at IS NOT NULL
           AND next_due_at <= v_now
         ORDER BY next_due_at
         FOR UPDATE SKIP LOCKED
    LOOP
        -- D-PE4 missed-window check. If the scheduled time is older
        -- than the window allows, we advance next_due_at without
        -- dispatching. This prevents a flood after a long outage.
        v_missed_cutoff := v_row.next_due_at + (v_row.missed_window_hours || ' hours')::interval;

        IF v_now > v_missed_cutoff THEN
            v_next_due := stewards.cron_next_after(v_row.cron_pattern, v_now);
            UPDATE stewards.scheduled_pipelines
               SET next_due_at = v_next_due,
                   updated_at  = v_now
             WHERE id = v_row.id;
            RAISE NOTICE 'scheduled_pipelines_fire: skipping missed run for % (due % was older than % hours); advanced next_due_at to %',
                v_row.slug, v_row.next_due_at, v_row.missed_window_hours, v_next_due;
            v_skipped_missed := v_skipped_missed + 1;
            CONTINUE;
        END IF;

        -- Compose a child work_item slug. Append YYYY-MM-DD-HHMM in UTC
        -- so daily, sub-daily, and weekly schedules all produce
        -- non-colliding slugs without any ambiguity.
        v_child_slug := v_row.slug || '--' ||
            to_char(v_row.next_due_at AT TIME ZONE 'UTC', 'YYYY-MM-DD-HH24MI');

        -- Dispatch. work_item_create returns the new uuid; we then
        -- dispatch the first stage immediately so the work_queue picks
        -- it up next tick.
        BEGIN
            v_work_item_id := stewards.work_item_create(
                p_pipeline_family => v_row.pipeline_family,
                p_input           => v_row.input_template,
                p_slug            => v_child_slug,
                p_actor           => 'scheduler',
                p_token_budget    => NULL,
                p_intent_id       => v_row.intent_id
            );
            PERFORM stewards.work_item_dispatch_stage(v_work_item_id);

            -- Advance the schedule
            v_next_due := stewards.cron_next_after(v_row.cron_pattern, v_now);
            UPDATE stewards.scheduled_pipelines
               SET last_dispatched_at = v_now,
                   next_due_at        = v_next_due,
                   updated_at         = v_now
             WHERE id = v_row.id;

            RAISE NOTICE 'scheduled_pipelines_fire: dispatched %/% as work_item %; next_due_at=%',
                v_row.slug, v_child_slug, v_work_item_id, v_next_due;
            v_dispatched := v_dispatched + 1;

        EXCEPTION WHEN OTHERS THEN
            -- Don't kill the whole tick on one bad row. Log + leave
            -- the row alone (its next_due_at stays in the past so we
            -- retry next tick — unless missed-window kicks in).
            RAISE NOTICE 'scheduled_pipelines_fire: dispatch failed for %: % (next tick will retry)',
                v_row.slug, SQLERRM;
        END;
    END LOOP;

    IF v_dispatched > 0 OR v_skipped_missed > 0 THEN
        RAISE NOTICE 'scheduled_pipelines_fire: dispatched=% missed_skipped=%',
            v_dispatched, v_skipped_missed;
    END IF;

    RETURN v_dispatched;
END;
$func$;

COMMENT ON FUNCTION stewards.scheduled_pipelines_fire() IS
'PE-B: scan scheduled_pipelines for due rows, dispatch work_items via work_item_create + work_item_dispatch_stage, honor D-PE4 fire-one-missed (skip missed runs older than missed_window_hours). Returns count dispatched. Called from watchman_scheduler_fire on the 60s leader tick.';

-- ---------------------------------------------------------------------
-- PE.B.2.b — extend watchman_scheduler_fire to also tick pipelines
--
-- Verbatim of the live body with two added blocks at the top:
--   1. Call scheduled_pipelines_fire() (independent of watchman state)
--   2. The rest of the existing body unchanged
--
-- We tick scheduled_pipelines FIRST so even when watchman is disabled
-- (soak paused), scheduled jobs still fire.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.watchman_scheduler_fire()
RETURNS text
LANGUAGE plpgsql AS $func$
DECLARE
    v_reason             text;
    v_cfg                stewards.watchman_config%ROWTYPE;
    v_pass_id            text;
    v_pipelines_fired    int;
BEGIN
    -- PE-B: dispatch any scheduled pipelines that are due. Independent
    -- of watchman pass logic — runs every tick even when the watchman
    -- soak is paused. EXCEPTION wrapper keeps a bad row from killing
    -- the watchman tick.
    BEGIN
        v_pipelines_fired := stewards.scheduled_pipelines_fire();
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'watchman_scheduler_fire: scheduled_pipelines_fire raised: %', SQLERRM;
    END;

    -- Original watchman logic below (verbatim).
    v_reason := stewards.watchman_should_fire();
    IF v_reason IS NULL THEN
        RETURN NULL;
    END IF;

    SELECT * INTO v_cfg FROM stewards.watchman_config WHERE id = 1;

    v_pass_id := stewards.watchman_pass_start(
        p_limit        => v_cfg.schedule_pass_limit,
        p_provider     => NULL,
        p_model        => NULL,
        p_agent_family => NULL,
        p_actor        => 'scheduler',
        p_trigger      => v_reason,
        p_token_budget => NULL
    );

    RAISE NOTICE 'watchman scheduler fired (%): pass_id=%', v_reason, v_pass_id;
    RETURN v_pass_id;
END;
$func$;

COMMENT ON FUNCTION stewards.watchman_scheduler_fire() IS
'PE-B update (2026-05-19): now also calls scheduled_pipelines_fire() at the top of each tick (independent of watchman state). Otherwise verbatim from Phase 2.7b.2.';

-- ---------------------------------------------------------------------
-- PE.B.2.c — seed the canonical ai-news-7am schedule
--
-- 0 13 * * 1-5 = 13:00 UTC weekdays = 7:00 MDT (MT in summer)
-- Idempotent via ON CONFLICT (slug) DO UPDATE.
-- ---------------------------------------------------------------------

INSERT INTO stewards.scheduled_pipelines (
    slug, pipeline_family, intent_id, cron_pattern, input_template,
    enabled, missed_window_hours, notes
)
VALUES (
    'ai-news-7am',
    'research-summary',
    (SELECT id FROM stewards.intents WHERE slug = 'general-research'),
    '0 13 * * 1-5',
    jsonb_build_object(
        'binding_question', 'What shipped in AI today that I should know about?',
        'sources_spec', jsonb_build_object(
            'queries', jsonb_build_array(
                'AI news today',
                'claude release',
                'openai update',
                'anthropic announcement',
                'GPT model release',
                'AI tooling launch'
            ),
            'max_per_query', 10,
            'since', '24h'
        ),
        'output_kind', 'daily-digest'
    ),
    true,
    24,
    'Daily AI news digest. Weekdays at 7am MT (13:00 UTC). Per D-PE4 fire-one-missed window = 24h. Output materializes to study/daily-digest/ via PE.5 promote_to_study path; joins AGE graph per D-PE7.'
)
ON CONFLICT (slug) DO UPDATE SET
    pipeline_family     = EXCLUDED.pipeline_family,
    intent_id           = EXCLUDED.intent_id,
    cron_pattern        = EXCLUDED.cron_pattern,
    input_template      = EXCLUDED.input_template,
    enabled             = EXCLUDED.enabled,
    missed_window_hours = EXCLUDED.missed_window_hours,
    notes               = EXCLUDED.notes,
    updated_at          = now();
