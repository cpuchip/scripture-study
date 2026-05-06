-- =====================================================================
-- Phase 2.7b.3 — Watchman per-pass token budget enforcement
--
-- Live-DB migration. Folds into extension/src/lib.rs at next intentional
-- rebuild (foldback debt: EIGHTH file — 2-6a/b/c, 2-7a, 3a, 2-7b1,
-- 2-7b2, 2-7b3).
--
-- Builds on:
--   - 2.7b.1 (watchman_pass_start, watchman_passes, completion trigger)
--   - 2.7b.2 (watchman_should_fire, scheduler tick)
--
-- This file adds:
--   1. stewards.watchman_passes.budget_stopped column — true when the
--      pass stopped enqueueing because the projected next-doc estimate
--      would have crossed token_budget.
--   2. stewards.estimate_chat_tokens(slug) function — best-effort
--      per-doc cost estimate based on input length + system prompt
--      overhead + historical avg output.
--   3. Replaces stewards.watchman_pass_start() with a budget-aware
--      version: tracks planned_tokens; stops the loop when the next
--      doc's estimate would exceed the budget; marks budget_stopped.
--
-- Enforcement is at ENQUEUE time only. Chats already enqueued run to
-- completion. Actual spend may slightly exceed budget if a chat outputs
-- much more than its estimate predicted. That's acceptable for v1;
-- mid-pass abort is not implemented.
-- =====================================================================

-- ---------------------------------------------------------------------
-- Add budget_stopped column.
-- ---------------------------------------------------------------------
ALTER TABLE stewards.watchman_passes
    ADD COLUMN IF NOT EXISTS budget_stopped boolean NOT NULL DEFAULT false;

COMMENT ON COLUMN stewards.watchman_passes.budget_stopped IS
'Phase 2.7b.3: true when watchman_pass_start stopped enqueueing because the next doc''s token estimate would have crossed token_budget. Tells the user "budget hit" vs. "queue empty / limit reached" when doc_count_planned < requested limit.';

-- ---------------------------------------------------------------------
-- estimate_chat_tokens(slug) — best-effort per-doc cost estimate.
--
-- Components:
--   input tokens   ≈ chars(watchman_input(slug)) / 4
--   system prompt  ≈ 1500 (compose_system_prompt for watchman is
--                          ~1.0-1.5KB of agent persona + instructions)
--   output tokens  = avg(tokens_out) from recent (30d) verdicts,
--                    or 3500 fallback if cold start
--
-- Returns a single int: total estimated tokens for one chat.
--
-- STABLE because the result is consistent within a single statement.
-- The 30-day window means the fallback can change between calls but
-- not within one watchman_pass_start invocation.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.estimate_chat_tokens(p_slug text)
RETURNS int
LANGUAGE plpgsql STABLE AS $func$
DECLARE
    v_input_chars int;
    v_input_tokens int;
    v_avg_out_tokens numeric;
    v_total int;
BEGIN
    -- Input length. NULL slug → 0 chars.
    v_input_chars := coalesce(length(stewards.watchman_input(p_slug)), 0);
    -- ~4 chars per token, conservative round-up.
    v_input_tokens := ceil(v_input_chars::numeric / 4)::int;

    -- Average output tokens from recent verdicts. Coalesce to 3500
    -- on cold start (first pass with no history). 3500 was the
    -- empirical median across 2.7b.1 + 2.7b.2 verifications.
    SELECT avg(tokens_out)
      INTO v_avg_out_tokens
      FROM stewards.verdicts
     WHERE created_at > now() - interval '30 days'
       AND tokens_out > 0;

    v_total := v_input_tokens
             + 1500                           -- system + persona overhead
             + coalesce(ceil(v_avg_out_tokens)::int, 3500);

    RETURN v_total;
END;
$func$;

COMMENT ON FUNCTION stewards.estimate_chat_tokens(text) IS
'Phase 2.7b.3: best-effort estimate of total tokens (in + out) for one watchman-consolidator chat on the given slug. Used by watchman_pass_start to enforce per-pass token_budget.';

-- ---------------------------------------------------------------------
-- Replace watchman_pass_start() with the budget-aware version.
--
-- Diff from 2.7b.1:
--   1. Track v_planned_tokens.
--   2. Compute per-doc estimate via estimate_chat_tokens(slug).
--   3. If v_planned_tokens + v_estimate > v_budget AND v_planned > 0,
--      stop the loop and mark budget_stopped.
--   4. If the very first doc's estimate alone > v_budget, still
--      enqueue NO docs (don't accidentally enqueue one over-budget
--      chat). The pass will be empty (doc_count_planned = 0,
--      budget_stopped = true). Honest signal to the user that the
--      budget is unworkable.
--
-- Everything else is identical to 2.7b.1.
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
    v_pass_id        text;
    v_provider       text;
    v_model          text;
    v_agent_family   text;
    v_budget         int;
    v_planned        int := 0;
    v_planned_tokens int := 0;
    v_estimate       int;
    v_budget_stopped boolean := false;
    v_slug           text;
    v_session_id     text;
    v_input          text;
    v_body           jsonb;
    v_payload        jsonb;
BEGIN
    -- Resolve defaults from config singleton.
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

    -- Pull dirty docs and enqueue chats, respecting both p_limit
    -- AND v_budget.
    FOR v_slug IN
        SELECT slug FROM stewards.dirty_queue
         ORDER BY coalesce(last_consolidated_at, 'epoch'::timestamptz),
                  updated_at
         LIMIT p_limit
    LOOP
        -- Compute cost estimate for this slug.
        v_estimate := stewards.estimate_chat_tokens(v_slug);

        -- Budget check. If adding this doc would cross v_budget, stop.
        -- We DO allow the very first doc even if its estimate alone
        -- exceeds the budget IFF v_planned is 0... no wait, that
        -- defeats the purpose. Stricter: if the first doc's estimate
        -- exceeds the budget, refuse to enqueue (budget_stopped=true,
        -- doc_count_planned=0). User sees "budget cannot fit even
        -- one doc" and can raise it.
        IF v_planned_tokens + v_estimate > v_budget THEN
            v_budget_stopped := true;
            EXIT;
        END IF;

        v_session_id := substring(v_pass_id || '--' || v_slug FROM 1 FOR 200);

        INSERT INTO stewards.sessions (id, label, kind)
        VALUES (v_session_id,
                'Watchman pass ' || v_pass_id || ' for ' || v_slug,
                'agent')
        ON CONFLICT (id) DO NOTHING;

        v_input := stewards.watchman_input(v_slug);
        IF v_input IS NULL THEN
            CONTINUE;
        END IF;

        INSERT INTO stewards.messages (session_id, role, content, model)
        VALUES (v_session_id, 'user', v_input, v_model);

        v_body := stewards.dry_run_chat(v_agent_family, v_model,
                                         v_session_id, NULL);

        v_payload := jsonb_build_object(
            'session_id',         v_session_id,
            'agent_family',       v_agent_family,
            'requested_model',    v_model,
            'meta',               v_body->'_meta',
            'body',               (v_body - '_meta')
                                  || jsonb_build_object('user', v_session_id),
            '_watchman_pass_id',  v_pass_id,
            '_watchman_slug',     v_slug,
            '_watchman_actor',    p_actor,
            '_watchman_estimate', v_estimate
        );

        INSERT INTO stewards.work_queue (kind, provider, payload)
        VALUES ('chat', v_provider, v_payload);

        v_planned        := v_planned + 1;
        v_planned_tokens := v_planned_tokens + v_estimate;
    END LOOP;

    UPDATE stewards.watchman_passes
       SET doc_count_planned = v_planned,
           budget_stopped    = v_budget_stopped
     WHERE pass_id = v_pass_id;

    -- Empty pass (no docs enqueued) → mark completed immediately.
    IF v_planned = 0 THEN
        UPDATE stewards.watchman_passes
           SET finished_at = now(),
               status      = 'completed'
         WHERE pass_id = v_pass_id;
    END IF;

    UPDATE stewards.watchman_config
       SET last_pass_at = now(),
           updated_at   = now()
     WHERE id = 1;

    RETURN v_pass_id;
END;
$func$;

COMMENT ON FUNCTION stewards.watchman_pass_start(int, text, text, text, text, text, int) IS
'Phase 2.7b.3: budget-aware version. Pulls top-N dirty docs but stops enqueueing if the next doc''s estimate would cross token_budget. Marks budget_stopped=true when that happens. Estimate via stewards.estimate_chat_tokens(slug).';

-- ---------------------------------------------------------------------
-- Update watchman_pass_summary view to expose budget_stopped.
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
    p.actor,
    p.budget_stopped
FROM stewards.watchman_passes p;
