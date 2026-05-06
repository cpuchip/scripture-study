-- =====================================================================
-- Phase 2.7b.1 — Watchman automation (SQL substrate + completion trigger)
--
-- Live-DB migration. Folds into extension/src/lib.rs at next intentional
-- rebuild (foldback debt: SIXTH file — 2-6a/b/c, 2-7a, 3a, 2-7b1).
--
-- Builds on:
--   - Phase 2.7a (stewards.verdicts, findings, dirty_queue,
--                 record_verdict, record_finding)
--   - Phase 3a   (watchman-consolidator agent, watchman_input)
--   - Phase 1.5/1.6 (chat_enqueue, dry_run_chat, message loop)
--
-- This file adds:
--   1. stewards.watchman_passes — one row per pass run.
--   2. stewards.watchman_config — singleton (id=1) with defaults.
--   3. stewards.watchman_pass_start() — pulls top-N dirty docs and
--      enqueues kind='chat' rows tagged with _watchman_pass_id.
--   4. stewards.advance_watchman_pass_counters() — helper used by
--      the completion trigger to roll up per-pass stats.
--   5. AFTER UPDATE OF status trigger on stewards.work_queue that
--      harvests verdicts/findings from completed watchman chats.
--   6. stewards.watchman_pass_summary view — convenient pass listing.
--
-- The bgworker stays generic. All Watchman semantics live in this SQL
-- and the trigger. The scheduler that wakes a pass automatically is
-- 2.7b.2 (not in this file).
-- =====================================================================

-- ---------------------------------------------------------------------
-- watchman_passes — one row per pass run
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS stewards.watchman_passes (
    pass_id            text PRIMARY KEY,
    started_at         timestamptz NOT NULL DEFAULT now(),
    finished_at        timestamptz,
    trigger            text NOT NULL DEFAULT 'manual'
                       CHECK (trigger IN ('manual','cron','pressure',
                                          'idle','api')),
    provider           text NOT NULL,
    model              text NOT NULL,
    agent_family       text NOT NULL DEFAULT 'watchman-consolidator',
    token_budget       int  NOT NULL DEFAULT 50000,
    actor              text NOT NULL DEFAULT 'watchman',
    -- Counters: planned at start, advanced by completion trigger.
    doc_count_planned  int  NOT NULL DEFAULT 0,
    doc_count_done     int  NOT NULL DEFAULT 0,
    tokens_in          int  NOT NULL DEFAULT 0,
    tokens_out         int  NOT NULL DEFAULT 0,
    verdict_counts     jsonb NOT NULL DEFAULT '{}'::jsonb,
    status             text NOT NULL DEFAULT 'in_progress'
                       CHECK (status IN ('in_progress','completed',
                                         'errored'))
);

CREATE INDEX IF NOT EXISTS watchman_passes_started_idx
    ON stewards.watchman_passes (started_at DESC);
CREATE INDEX IF NOT EXISTS watchman_passes_status_idx
    ON stewards.watchman_passes (status, started_at DESC);

COMMENT ON TABLE stewards.watchman_passes IS
'Phase 2.7b.1: one row per Watchman consolidation pass. doc_count_done, tokens_*, and verdict_counts are advanced by the AFTER UPDATE trigger on work_queue as each chat completes. Pass auto-completes when doc_count_done >= doc_count_planned.';

-- ---------------------------------------------------------------------
-- watchman_config — singleton (id=1)
-- 2.7b.2's scheduler will read this; 2.7b.1 just creates it with sane
-- defaults so the table is queryable from day one.
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS stewards.watchman_config (
    id                    int PRIMARY KEY DEFAULT 1
                          CHECK (id = 1),
    schedule_cron         text NOT NULL DEFAULT 'weekly@sun-03:00',
    default_provider      text NOT NULL DEFAULT 'opencode_go',
    default_model         text NOT NULL DEFAULT 'kimi-k2.6',
    default_agent_family  text NOT NULL DEFAULT 'watchman-consolidator',
    token_budget          int  NOT NULL DEFAULT 50000,
    -- 2.7b.2 triggers a pass when dirty_queue exceeds this.
    dirty_threshold       int  NOT NULL DEFAULT 50,
    idle_threshold_hours  int  NOT NULL DEFAULT 48,
    last_pass_at          timestamptz,
    updated_at            timestamptz NOT NULL DEFAULT now()
);

INSERT INTO stewards.watchman_config (id) VALUES (1)
ON CONFLICT (id) DO NOTHING;

COMMENT ON TABLE stewards.watchman_config IS
'Phase 2.7b.1: singleton config row (id=1) with Watchman defaults. 2.7b.2 reads schedule_cron + dirty_threshold + idle_threshold_hours to decide when to fire a pass automatically. 2.7b.1 just creates the row.';

-- ---------------------------------------------------------------------
-- advance_watchman_pass_counters(pass_id, verdict, tokens_in, tokens_out)
--
-- Called from the completion trigger. Increments doc_count_done,
-- adds tokens, increments the verdict_counts jsonb counter for the
-- specific verdict. When doc_count_done catches up to doc_count_planned,
-- marks the pass completed and stamps finished_at.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.advance_watchman_pass_counters(
    p_pass_id    text,
    p_verdict    text,
    p_tokens_in  int,
    p_tokens_out int
) RETURNS void
LANGUAGE plpgsql AS $func$
DECLARE
    v_planned int;
    v_done    int;
BEGIN
    UPDATE stewards.watchman_passes
       SET doc_count_done = doc_count_done + 1,
           tokens_in      = tokens_in + coalesce(p_tokens_in, 0),
           tokens_out     = tokens_out + coalesce(p_tokens_out, 0),
           verdict_counts = jsonb_set(
               coalesce(verdict_counts, '{}'::jsonb),
               ARRAY[p_verdict],
               to_jsonb(coalesce(
                   (verdict_counts->>p_verdict)::int, 0) + 1)
           )
     WHERE pass_id = p_pass_id
     RETURNING doc_count_planned, doc_count_done
        INTO v_planned, v_done;

    IF v_planned IS NOT NULL
       AND v_planned > 0
       AND v_done >= v_planned THEN
        UPDATE stewards.watchman_passes
           SET finished_at = now(),
               status      = 'completed'
         WHERE pass_id = p_pass_id
           AND status = 'in_progress';
    END IF;
END;
$func$;

-- ---------------------------------------------------------------------
-- watchman_pass_start(...)
--
-- Inserts the watchman_passes row, pulls top-N dirty docs, for each:
--   - Composes user input via watchman_input(slug).
--   - Creates a deterministic session id (pass_id--slug, capped at 200).
--   - Inserts the user message + composes body via dry_run_chat.
--   - Builds a payload with both the standard chat shape AND
--     _watchman_pass_id / _watchman_slug / _watchman_actor extras.
--   - Inserts work_queue row (kind='chat'). Bgworker dispatches normally.
-- Returns the new pass_id.
--
-- Runs in a single transaction. The work_queue rows become visible to
-- the bgworker only after this function returns and the caller commits.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.watchman_pass_start(
    p_limit         int  DEFAULT 5,
    p_provider      text DEFAULT NULL,
    p_model         text DEFAULT NULL,
    p_agent_family  text DEFAULT NULL,
    p_actor         text DEFAULT 'watchman',
    p_trigger       text DEFAULT 'manual',
    p_token_budget  int  DEFAULT NULL
) RETURNS text
LANGUAGE plpgsql AS $func$
DECLARE
    v_pass_id      text;
    v_provider     text;
    v_model        text;
    v_agent_family text;
    v_budget       int;
    v_planned      int := 0;
    v_slug         text;
    v_session_id   text;
    v_input        text;
    v_body         jsonb;
    v_payload      jsonb;
BEGIN
    -- Resolve defaults from config singleton (with hard fallbacks if
    -- the row was deleted or never created).
    SELECT coalesce(p_provider,     default_provider,     'opencode_go'),
           coalesce(p_model,        default_model,        'kimi-k2.6'),
           coalesce(p_agent_family, default_agent_family, 'watchman-consolidator'),
           coalesce(p_token_budget, token_budget,         50000)
      INTO v_provider, v_model, v_agent_family, v_budget
      FROM stewards.watchman_config
     WHERE id = 1;

    IF v_provider IS NULL THEN
        v_provider     := coalesce(p_provider,     'opencode_go');
        v_model        := coalesce(p_model,        'kimi-k2.6');
        v_agent_family := coalesce(p_agent_family, 'watchman-consolidator');
        v_budget       := coalesce(p_token_budget, 50000);
    END IF;

    -- pass_id: timestamp + short uuid suffix to disambiguate same-second
    -- pass_now invocations from CLI/API.
    v_pass_id := 'watchman-'
                 || to_char(now() AT TIME ZONE 'UTC',
                            'YYYYMMDD"T"HH24MISS"Z"')
                 || '-'
                 || substring(replace(gen_random_uuid()::text, '-', '')
                              FROM 1 FOR 6);

    INSERT INTO stewards.watchman_passes
        (pass_id, started_at, trigger, provider, model, agent_family,
         token_budget, actor, status)
    VALUES
        (v_pass_id, now(), p_trigger, v_provider, v_model,
         v_agent_family, v_budget, p_actor, 'in_progress');

    -- Pull dirty docs and enqueue chats. Order matches dirty_queue's
    -- own ordering (oldest-consolidated first, then oldest-touched).
    FOR v_slug IN
        SELECT slug FROM stewards.dirty_queue
         ORDER BY coalesce(last_consolidated_at, 'epoch'::timestamptz),
                  updated_at
         LIMIT p_limit
    LOOP
        v_session_id := substring(v_pass_id || '--' || v_slug FROM 1 FOR 200);

        INSERT INTO stewards.sessions (id, label, kind)
        VALUES (v_session_id,
                'Watchman pass ' || v_pass_id || ' for ' || v_slug,
                'agent')
        ON CONFLICT (id) DO NOTHING;

        v_input := stewards.watchman_input(v_slug);
        IF v_input IS NULL THEN
            -- Doc disappeared between dirty_queue read and now. Skip.
            CONTINUE;
        END IF;

        -- Persist user message (mirrors chat_enqueue's behavior).
        INSERT INTO stewards.messages (session_id, role, content, model)
        VALUES (v_session_id, 'user', v_input, v_model);

        -- Compose body via dry_run_chat with NULL user_input — the
        -- history already carries everything. Same shape as
        -- chat_post_internal's enqueue path.
        v_body := stewards.dry_run_chat(v_agent_family, v_model,
                                         v_session_id, NULL);

        v_payload := jsonb_build_object(
            'session_id',         v_session_id,
            'agent_family',       v_agent_family,
            'requested_model',    v_model,
            'meta',               v_body->'_meta',
            'body',               (v_body - '_meta')
                                  || jsonb_build_object('user', v_session_id),
            -- Watchman-specific extras read by the completion trigger:
            '_watchman_pass_id',  v_pass_id,
            '_watchman_slug',     v_slug,
            '_watchman_actor',    p_actor
        );

        INSERT INTO stewards.work_queue (kind, provider, payload)
        VALUES ('chat', v_provider, v_payload);

        v_planned := v_planned + 1;
    END LOOP;

    UPDATE stewards.watchman_passes
       SET doc_count_planned = v_planned
     WHERE pass_id = v_pass_id;

    -- Empty dirty queue → mark completed immediately so callers polling
    -- on status see a clean terminal state.
    IF v_planned = 0 THEN
        UPDATE stewards.watchman_passes
           SET finished_at = now(),
               status      = 'completed'
         WHERE pass_id = v_pass_id;
    END IF;

    -- Stamp last_pass_at for the 2.7b.2 scheduler.
    UPDATE stewards.watchman_config
       SET last_pass_at = now(),
           updated_at   = now()
     WHERE id = 1;

    RETURN v_pass_id;
END;
$func$;

COMMENT ON FUNCTION stewards.watchman_pass_start(int, text, text, text, text, text, int) IS
'Phase 2.7b.1: enqueues N watchman chats from the dirty_queue, tagging each work_queue payload with _watchman_pass_id/_watchman_slug. Returns the new pass_id. Result harvesting happens in the completion trigger.';

-- ---------------------------------------------------------------------
-- handle_watchman_chat_completion()
--
-- Trigger function. Fires AFTER UPDATE OF status on stewards.work_queue
-- with a WHEN guard limiting it to chat rows tagged with
-- _watchman_pass_id. When a watchman chat row transitions to
-- 'done' or 'error':
--
--   1. Read the latest assistant message for the session.
--   2. Strip optional ```json fences.
--   3. Cast content to jsonb. Bad JSON → record verdict='skipped'.
--   4. Validate verdict against the 5-element enum. Invalid → 'skipped'.
--   5. Call record_verdict; if non-clean and finding present, call
--      record_finding.
--   6. Advance watchman_passes counters via
--      advance_watchman_pass_counters.
--
-- Defensive: every record_verdict / record_finding call is wrapped in
-- BEGIN/EXCEPTION so a bug in the harvester never breaks the bgworker's
-- work_queue UPDATE.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.handle_watchman_chat_completion()
RETURNS trigger
LANGUAGE plpgsql AS $func$
DECLARE
    v_pass_id    text;
    v_slug       text;
    v_session_id text;
    v_actor      text;
    v_content    text;
    v_tokens_in  int;
    v_tokens_out int;
    v_model      text;
    v_parsed     jsonb;
    v_verdict    text;
    v_reasoning  text;
    v_finding    jsonb;
    v_skipped_reason text;
BEGIN
    -- Defensive (the WHEN clause already filters; this catches updates
    -- to old rows whose payload didn't have the markers when WHEN was
    -- evaluated, e.g. payload got rewritten mid-flight).
    IF NEW.kind <> 'chat'
       OR (NEW.payload->>'_watchman_pass_id') IS NULL THEN
        RETURN NEW;
    END IF;

    -- Only fire on completion transitions.
    IF NEW.status NOT IN ('done', 'error') THEN
        RETURN NEW;
    END IF;
    IF OLD.status = NEW.status THEN
        RETURN NEW;
    END IF;

    v_pass_id    := NEW.payload->>'_watchman_pass_id';
    v_slug       := NEW.payload->>'_watchman_slug';
    v_session_id := NEW.payload->>'session_id';
    v_actor      := coalesce(NEW.payload->>'_watchman_actor', 'watchman');

    -- ----- error path: record skipped verdict with the chat error -----
    IF NEW.status = 'error' THEN
        v_skipped_reason := 'watchman chat errored: '
                            || coalesce(NEW.error, '(no error msg)');
        BEGIN
            PERFORM stewards.record_verdict(
                v_slug, 'skipped', v_skipped_reason,
                NULL, 0, 0, v_pass_id, v_actor);
        EXCEPTION WHEN OTHERS THEN
            RAISE WARNING
                'watchman trigger record_verdict failed for %: %',
                v_slug, SQLERRM;
        END;
        BEGIN
            PERFORM stewards.advance_watchman_pass_counters(
                v_pass_id, 'skipped', 0, 0);
        EXCEPTION WHEN OTHERS THEN
            RAISE WARNING
                'watchman trigger advance_counters failed for pass %: %',
                v_pass_id, SQLERRM;
        END;
        RETURN NEW;
    END IF;

    -- ----- done path: read assistant message, parse, record -----
    SELECT m.content, m.tokens_in, m.tokens_out, m.model
      INTO v_content, v_tokens_in, v_tokens_out, v_model
      FROM stewards.messages m
     WHERE m.session_id = v_session_id
       AND m.role = 'assistant'
     ORDER BY m.id DESC
     LIMIT 1;

    IF v_content IS NULL OR length(trim(v_content)) = 0 THEN
        v_skipped_reason := 'watchman: no assistant message for session '
                            || v_session_id;
        BEGIN
            PERFORM stewards.record_verdict(
                v_slug, 'skipped', v_skipped_reason,
                v_model, 0, 0, v_pass_id, v_actor);
        EXCEPTION WHEN OTHERS THEN NULL;
        END;
        BEGIN
            PERFORM stewards.advance_watchman_pass_counters(
                v_pass_id, 'skipped', 0, 0);
        EXCEPTION WHEN OTHERS THEN NULL;
        END;
        RETURN NEW;
    END IF;

    -- Strip optional code-fence wrapper. kimi/qwen sometimes wrap JSON
    -- in ```json ... ``` even when response_format demands raw JSON.
    v_content := regexp_replace(v_content,
        '^\s*```(?:json|JSON)?\s*\n', '');
    v_content := regexp_replace(v_content, '\n```\s*$', '');
    v_content := trim(v_content);

    -- Try to parse JSON.
    BEGIN
        v_parsed := v_content::jsonb;
    EXCEPTION WHEN OTHERS THEN
        v_skipped_reason := 'watchman: failed to parse assistant JSON: '
                            || SQLERRM;
        BEGIN
            PERFORM stewards.record_verdict(
                v_slug, 'skipped', v_skipped_reason,
                v_model,
                coalesce(v_tokens_in, 0),
                coalesce(v_tokens_out, 0),
                v_pass_id, v_actor);
        EXCEPTION WHEN OTHERS THEN NULL;
        END;
        BEGIN
            PERFORM stewards.advance_watchman_pass_counters(
                v_pass_id, 'skipped',
                coalesce(v_tokens_in, 0),
                coalesce(v_tokens_out, 0));
        EXCEPTION WHEN OTHERS THEN NULL;
        END;
        RETURN NEW;
    END;

    v_verdict   := v_parsed->>'verdict';
    v_reasoning := coalesce(v_parsed->>'reasoning', '');
    v_finding   := v_parsed->'finding';

    IF v_verdict IS NULL
       OR v_verdict NOT IN ('clean','drift','done','superseded','skipped') THEN
        v_skipped_reason := 'watchman: invalid or missing verdict: '
                            || coalesce(v_verdict, '(null)');
        BEGIN
            PERFORM stewards.record_verdict(
                v_slug, 'skipped', v_skipped_reason,
                v_model,
                coalesce(v_tokens_in, 0),
                coalesce(v_tokens_out, 0),
                v_pass_id, v_actor);
        EXCEPTION WHEN OTHERS THEN NULL;
        END;
        BEGIN
            PERFORM stewards.advance_watchman_pass_counters(
                v_pass_id, 'skipped',
                coalesce(v_tokens_in, 0),
                coalesce(v_tokens_out, 0));
        EXCEPTION WHEN OTHERS THEN NULL;
        END;
        RETURN NEW;
    END IF;

    -- Happy path. Record verdict, then optionally finding, then advance.
    BEGIN
        PERFORM stewards.record_verdict(
            v_slug, v_verdict, v_reasoning,
            v_model,
            coalesce(v_tokens_in, 0),
            coalesce(v_tokens_out, 0),
            v_pass_id, v_actor);
    EXCEPTION WHEN OTHERS THEN
        RAISE WARNING
            'watchman trigger record_verdict failed for %: %',
            v_slug, SQLERRM;
    END;

    IF v_finding IS NOT NULL
       AND jsonb_typeof(v_finding) = 'object'
       AND v_verdict <> 'clean' THEN
        BEGIN
            PERFORM stewards.record_finding(
                v_slug,
                coalesce(v_finding->>'kind', 'drift'),
                coalesce(v_finding->>'message', '(no message)'),
                coalesce(v_finding->>'severity', 'medium'),
                v_finding->>'suggested_action',
                ARRAY[]::text[],
                v_pass_id, v_actor);
        EXCEPTION WHEN OTHERS THEN
            RAISE WARNING
                'watchman trigger record_finding failed for %: %',
                v_slug, SQLERRM;
        END;
    END IF;

    BEGIN
        PERFORM stewards.advance_watchman_pass_counters(
            v_pass_id, v_verdict,
            coalesce(v_tokens_in, 0),
            coalesce(v_tokens_out, 0));
    EXCEPTION WHEN OTHERS THEN
        RAISE WARNING
            'watchman trigger advance_counters failed for pass %: %',
            v_pass_id, SQLERRM;
    END;

    RETURN NEW;
END;
$func$;

-- Drop and recreate the trigger so re-applying this file is idempotent
-- across DB rebuilds.
DROP TRIGGER IF EXISTS watchman_harvest_completion ON stewards.work_queue;

CREATE TRIGGER watchman_harvest_completion
    AFTER UPDATE OF status ON stewards.work_queue
    FOR EACH ROW
    WHEN ((NEW.kind = 'chat')
          AND (NEW.payload ? '_watchman_pass_id')
          AND (NEW.status IN ('done', 'error'))
          AND (OLD.status IS DISTINCT FROM NEW.status))
    EXECUTE FUNCTION stewards.handle_watchman_chat_completion();

COMMENT ON FUNCTION stewards.handle_watchman_chat_completion() IS
'Phase 2.7b.1: AFTER UPDATE trigger function on work_queue. Harvests verdict + finding from a completed watchman chat, records them, and advances watchman_passes counters. All side effects in the same tx as the work_queue status flip.';

-- ---------------------------------------------------------------------
-- watchman_pass_summary view — convenient pass listing
-- ---------------------------------------------------------------------
CREATE OR REPLACE VIEW stewards.watchman_pass_summary AS
SELECT
    p.pass_id,
    p.started_at,
    p.finished_at,
    (p.finished_at - p.started_at) AS elapsed,
    p.trigger,
    p.provider,
    p.model,
    p.status,
    p.doc_count_planned,
    p.doc_count_done,
    p.tokens_in,
    p.tokens_out,
    coalesce((p.verdict_counts->>'clean')::int,      0) AS n_clean,
    coalesce((p.verdict_counts->>'drift')::int,      0) AS n_drift,
    coalesce((p.verdict_counts->>'done')::int,       0) AS n_done,
    coalesce((p.verdict_counts->>'superseded')::int, 0) AS n_superseded,
    coalesce((p.verdict_counts->>'skipped')::int,    0) AS n_skipped,
    p.token_budget,
    p.actor
FROM stewards.watchman_passes p;

COMMENT ON VIEW stewards.watchman_pass_summary IS
'Phase 2.7b.1: per-pass summary with verdict_counts unpacked into named columns. CLI watchman passes reads from here.';
