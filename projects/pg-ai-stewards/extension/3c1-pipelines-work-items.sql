-- =====================================================================
-- Phase 3c.1 — Pipelines + work_items: orchestration above work_queue
--
-- Live-DB migration. Folds into extension/src/lib.rs at next intentional
-- rebuild (foldback debt: TENTH file).
--
-- Builds on:
--   - Phase 1.5/1.6 (agents, dry_run_chat, chat dispatch via work_queue)
--   - Phase 2.7b.1 (the AFTER UPDATE trigger pattern this phase mirrors)
--   - Phase 3a.1 (the agent corpus that pipelines orchestrate)
--
-- Architecture choice (option B from the conversation): pipelines are
-- an orchestration layer ABOVE work_queue. A pipeline stage = "enqueue
-- a chat with this agent." When the chat completes, the work_item
-- advances (3c.2 — auto-advance trigger; 3c.1 ships manual advance only).
--
-- Same pattern as Phase 2.7b.1's Watchman: payload markers
-- (`_work_item_id`, `_stage_name`) connect work_queue rows back to
-- their parent work_item. The bgworker stays generic.
--
-- This file adds:
--   1. stewards.pipelines — pipeline definitions (immutable templates).
--   2. stewards.work_items — instances flowing through stages.
--   3. Transition functions (work_item_create, work_item_dispatch_stage,
--      work_item_advance, work_item_fail, work_item_cancel).
--   4. work_items_active / work_items_summary views for inspection.
--   5. Seed pipeline `echo-test` — single-stage smoke test that proves
--      the wiring (one stewards-explore chat dispatch, then complete).
-- =====================================================================

-- ---------------------------------------------------------------------
-- pipelines — definitions
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS stewards.pipelines (
    family       text PRIMARY KEY
                 CHECK (family ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    description  text NOT NULL DEFAULT '',
    -- stages: jsonb array. Each element is an object with:
    --   name           text  required, unique within the pipeline
    --   agent_family   text  required, refs stewards.agents
    --   model          text  required (the requested model)
    --   provider       text  required (e.g., 'opencode_go', 'lm_studio')
    --   next           text  next stage name; NULL/missing for terminal
    --   auto_advance   bool  default true; false = stop at awaiting_review
    -- More fields can be added later (input_template, gate_predicate, etc.)
    stages       jsonb NOT NULL
                 CHECK (jsonb_typeof(stages) = 'array'
                        AND jsonb_array_length(stages) >= 1),
    metadata     jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE stewards.pipelines IS
'Phase 3c.1: pipeline definitions. Each row is an immutable template describing the stages of a multi-step agent flow. work_items are instances that traverse a pipeline''s stages.';

-- ---------------------------------------------------------------------
-- work_items — instances flowing through pipeline stages
-- ---------------------------------------------------------------------
-- Status lifecycle:
--   pending          — created, current_stage not yet dispatched
--   in_progress      — current_stage's chat dispatched (work_queue row pending/in_progress)
--   awaiting_review  — current_stage completed; pipeline says auto_advance=false; human ack needed
--   completed        — all stages done; terminal
--   failed           — error encountered; recoverable via human (cancel or retry)
--   cancelled        — terminal, intentional stop
CREATE TABLE IF NOT EXISTS stewards.work_items (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug            text UNIQUE,
    pipeline_family text NOT NULL
                    REFERENCES stewards.pipelines(family) ON DELETE RESTRICT,
    current_stage   text NOT NULL,
    status          text NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'in_progress',
                                       'awaiting_review', 'completed',
                                       'failed', 'cancelled')),
    -- Opening inputs — the user-supplied data the first stage works on.
    input           jsonb NOT NULL DEFAULT '{}'::jsonb,
    -- Per-stage outputs accumulate here keyed by stage name:
    --   {"outline": {"session_id": "...", "completed_at": "...",
    --                "tokens_in": N, "tokens_out": N, "output": "..."},
    --    ...}
    stage_results   jsonb NOT NULL DEFAULT '{}'::jsonb,
    -- All chat session ids spawned by this work_item (one per stage).
    -- Useful for `SELECT * FROM stewards.messages WHERE session_id = ANY(...)`.
    session_ids     text[] NOT NULL DEFAULT ARRAY[]::text[],
    -- Cost guards
    token_budget    int,
    tokens_in       int NOT NULL DEFAULT 0,
    tokens_out      int NOT NULL DEFAULT 0,
    -- Audit
    actor           text NOT NULL DEFAULT 'human',
    error           text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    completed_at    timestamptz
);

CREATE INDEX IF NOT EXISTS work_items_status_idx
    ON stewards.work_items (status, created_at DESC);
CREATE INDEX IF NOT EXISTS work_items_pipeline_idx
    ON stewards.work_items (pipeline_family);
CREATE INDEX IF NOT EXISTS work_items_active_idx
    ON stewards.work_items (created_at DESC)
    WHERE status NOT IN ('completed', 'cancelled');

COMMENT ON TABLE stewards.work_items IS
'Phase 3c.1: instances flowing through a pipeline''s stages. Each stage''s output is recorded in stage_results keyed by stage name. session_ids carries the chat session id per dispatched stage so the full message history is reachable via `SELECT * FROM messages WHERE session_id = ANY(work_item.session_ids)`.';

-- ---------------------------------------------------------------------
-- Helper: stewards.pipeline_stage_lookup(family, stage_name)
-- Returns the stage's jsonb object or NULL if not found.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.pipeline_stage_lookup(
    p_family     text,
    p_stage_name text
) RETURNS jsonb
LANGUAGE sql STABLE AS $func$
    SELECT s
      FROM stewards.pipelines p,
           jsonb_array_elements(p.stages) AS s
     WHERE p.family = p_family
       AND s->>'name' = p_stage_name
     LIMIT 1;
$func$;

-- ---------------------------------------------------------------------
-- Helper: stewards.pipeline_first_stage_name(family)
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.pipeline_first_stage_name(p_family text)
RETURNS text
LANGUAGE sql STABLE AS $func$
    SELECT (stages->0)->>'name'
      FROM stewards.pipelines
     WHERE family = p_family;
$func$;

-- ---------------------------------------------------------------------
-- work_item_create(pipeline, input, slug?, actor?, token_budget?)
--
-- Creates a new work_item with status='pending', current_stage =
-- pipeline's first stage. Does NOT auto-dispatch; caller decides when
-- via work_item_dispatch_stage() (or 3c.2's auto-advance trigger).
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.work_item_create(
    p_pipeline_family text,
    p_input           jsonb DEFAULT '{}'::jsonb,
    p_slug            text  DEFAULT NULL,
    p_actor           text  DEFAULT 'human',
    p_token_budget    int   DEFAULT NULL
) RETURNS uuid
LANGUAGE plpgsql AS $func$
DECLARE
    v_first_stage text;
    v_id          uuid;
BEGIN
    SELECT stewards.pipeline_first_stage_name(p_pipeline_family)
      INTO v_first_stage;
    IF v_first_stage IS NULL THEN
        RAISE EXCEPTION
            'work_item_create: pipeline % not found or has no stages',
            p_pipeline_family;
    END IF;

    INSERT INTO stewards.work_items
        (pipeline_family, current_stage, slug, input, actor, token_budget)
    VALUES
        (p_pipeline_family, v_first_stage, p_slug, p_input, p_actor, p_token_budget)
    RETURNING id INTO v_id;

    RETURN v_id;
END;
$func$;

COMMENT ON FUNCTION stewards.work_item_create(text, jsonb, text, text, int) IS
'Phase 3c.1: create a new work_item bound to a pipeline. Status starts ''pending'' with current_stage = first stage in the pipeline definition. Caller dispatches with work_item_dispatch_stage().';

-- ---------------------------------------------------------------------
-- work_item_dispatch_stage(work_item_id)
--
-- Composes input + payload + enqueues a chat work_queue row for the
-- work_item's current_stage. Sets status='in_progress'.
--
-- Mirrors the Phase 2.7b.1 pattern: builds payload directly (not via
-- chat_enqueue) so we can inject `_work_item_id` / `_stage_name`
-- markers that 3c.2's auto-advance trigger will read.
--
-- Returns the new work_queue id.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.work_item_dispatch_stage(
    p_work_item_id uuid,
    p_user_input   text DEFAULT NULL  -- override for first stage; later stages
                                       -- typically derive from prior outputs
) RETURNS bigint
LANGUAGE plpgsql AS $func$
DECLARE
    v_wi          stewards.work_items%ROWTYPE;
    v_stage       jsonb;
    v_agent       text;
    v_model       text;
    v_provider    text;
    v_session_id  text;
    v_user_input  text;
    v_body        jsonb;
    v_payload     jsonb;
    v_work_id     bigint;
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN
        RAISE EXCEPTION 'work_item % not found', p_work_item_id;
    END IF;
    IF v_wi.status NOT IN ('pending', 'awaiting_review') THEN
        RAISE EXCEPTION 'work_item %: cannot dispatch from status %',
            p_work_item_id, v_wi.status;
    END IF;

    v_stage := stewards.pipeline_stage_lookup(v_wi.pipeline_family, v_wi.current_stage);
    IF v_stage IS NULL THEN
        RAISE EXCEPTION 'work_item %: stage % not found in pipeline %',
            p_work_item_id, v_wi.current_stage, v_wi.pipeline_family;
    END IF;

    v_agent    := v_stage->>'agent_family';
    v_model    := v_stage->>'model';
    v_provider := v_stage->>'provider';
    IF v_agent IS NULL OR v_model IS NULL OR v_provider IS NULL THEN
        RAISE EXCEPTION 'work_item %: stage % missing agent_family/model/provider',
            p_work_item_id, v_wi.current_stage;
    END IF;

    -- Session id pattern: wi--<short-uuid>--<stage>, capped at 200.
    v_session_id := substring(
        'wi--' || substring(p_work_item_id::text FROM 1 FOR 8)
        || '--' || v_wi.current_stage
        FROM 1 FOR 200);

    INSERT INTO stewards.sessions (id, label, kind)
    VALUES (v_session_id,
            format('work_item %s stage %s', v_wi.id, v_wi.current_stage),
            'agent')
    ON CONFLICT (id) DO NOTHING;

    -- Resolve user input. For 3c.1 we accept an explicit override, or
    -- fall back to the work_item.input.user_input field, or stringify
    -- the whole input jsonb. Templating (3c.3) will replace this.
    v_user_input := coalesce(
        p_user_input,
        v_wi.input->>'user_input',
        v_wi.input::text
    );

    INSERT INTO stewards.messages (session_id, role, content, model)
    VALUES (v_session_id, 'user', v_user_input, v_model);

    v_body := stewards.dry_run_chat(v_agent, v_model, v_session_id, NULL);

    v_payload := jsonb_build_object(
        'session_id',         v_session_id,
        'agent_family',       v_agent,
        'requested_model',    v_model,
        'meta',               v_body->'_meta',
        'body',               (v_body - '_meta')
                              || jsonb_build_object('user', v_session_id),
        -- 3c.1 markers — read by the 3c.2 auto-advance trigger:
        '_work_item_id',      p_work_item_id::text,
        '_stage_name',        v_wi.current_stage,
        '_pipeline_family',   v_wi.pipeline_family
    );

    INSERT INTO stewards.work_queue (kind, provider, payload)
    VALUES ('chat', v_provider, v_payload)
    RETURNING id INTO v_work_id;

    -- Update work_item: status, append session_id, bump updated_at.
    UPDATE stewards.work_items
       SET status      = 'in_progress',
           session_ids = session_ids || v_session_id,
           updated_at  = now()
     WHERE id = p_work_item_id;

    RETURN v_work_id;
END;
$func$;

COMMENT ON FUNCTION stewards.work_item_dispatch_stage(uuid, text) IS
'Phase 3c.1: dispatch the current stage. Composes the chat body via dry_run_chat, enqueues a kind=chat work_queue row with _work_item_id/_stage_name markers, and sets status=in_progress. Caller (or 3c.2 trigger) advances after completion.';

-- ---------------------------------------------------------------------
-- work_item_advance(work_item_id, stage_output)
--
-- Records the current stage's output, finds the next stage, and either
-- (a) advances current_stage and resets status to pending (caller can
--     re-dispatch), OR
-- (b) marks the work_item completed if there's no next stage.
--
-- Returns the next stage name, or NULL if completed.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.work_item_advance(
    p_work_item_id uuid,
    p_stage_output jsonb DEFAULT '{}'::jsonb
) RETURNS text
LANGUAGE plpgsql AS $func$
DECLARE
    v_wi          stewards.work_items%ROWTYPE;
    v_stage       jsonb;
    v_next_name   text;
    v_auto_advance bool;
    v_results     jsonb;
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN
        RAISE EXCEPTION 'work_item % not found', p_work_item_id;
    END IF;
    IF v_wi.status NOT IN ('in_progress', 'awaiting_review', 'pending') THEN
        RAISE EXCEPTION 'work_item %: cannot advance from status %',
            p_work_item_id, v_wi.status;
    END IF;

    v_stage := stewards.pipeline_stage_lookup(v_wi.pipeline_family, v_wi.current_stage);
    IF v_stage IS NULL THEN
        RAISE EXCEPTION 'work_item %: stage % not found in pipeline %',
            p_work_item_id, v_wi.current_stage, v_wi.pipeline_family;
    END IF;

    v_next_name := v_stage->>'next';
    -- coalesce missing/null auto_advance to true
    v_auto_advance := coalesce((v_stage->>'auto_advance')::bool, true);

    -- Record this stage's output keyed by stage name.
    v_results := v_wi.stage_results
              || jsonb_build_object(v_wi.current_stage,
                     p_stage_output
                     || jsonb_build_object('completed_at', now()));

    IF v_next_name IS NULL OR v_next_name = '' THEN
        -- Terminal: no next stage.
        UPDATE stewards.work_items
           SET stage_results = v_results,
               status        = 'completed',
               completed_at  = now(),
               updated_at    = now()
         WHERE id = p_work_item_id;
        RETURN NULL;
    END IF;

    -- Validate next stage exists in the pipeline.
    IF stewards.pipeline_stage_lookup(v_wi.pipeline_family, v_next_name) IS NULL THEN
        RAISE EXCEPTION
            'work_item %: stage %s `next` references missing stage %',
            p_work_item_id, v_wi.current_stage, v_next_name;
    END IF;

    -- Advance. If auto_advance=false on the COMPLETING stage, set
    -- status=awaiting_review (human must call advance again or
    -- dispatch_stage to proceed). Otherwise pending → caller dispatches.
    UPDATE stewards.work_items
       SET stage_results = v_results,
           current_stage = v_next_name,
           status        = CASE WHEN v_auto_advance THEN 'pending'
                                ELSE 'awaiting_review' END,
           updated_at    = now()
     WHERE id = p_work_item_id;

    RETURN v_next_name;
END;
$func$;

COMMENT ON FUNCTION stewards.work_item_advance(uuid, jsonb) IS
'Phase 3c.1: record the current stage''s output and transition to the next stage (or mark completed if terminal). Returns next stage name or NULL.';

-- ---------------------------------------------------------------------
-- work_item_fail / work_item_cancel
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.work_item_fail(
    p_work_item_id uuid,
    p_error        text
) RETURNS void
LANGUAGE plpgsql AS $func$
BEGIN
    UPDATE stewards.work_items
       SET status     = 'failed',
           error      = p_error,
           updated_at = now()
     WHERE id = p_work_item_id
       AND status NOT IN ('completed', 'cancelled');
    IF NOT FOUND THEN
        RAISE EXCEPTION
            'work_item_fail: % not found or already in terminal status',
            p_work_item_id;
    END IF;
END;
$func$;

CREATE OR REPLACE FUNCTION stewards.work_item_cancel(
    p_work_item_id uuid,
    p_reason       text DEFAULT NULL
) RETURNS void
LANGUAGE plpgsql AS $func$
BEGIN
    UPDATE stewards.work_items
       SET status       = 'cancelled',
           error        = coalesce(p_reason, error),
           updated_at   = now(),
           completed_at = now()
     WHERE id = p_work_item_id
       AND status NOT IN ('completed', 'cancelled');
    IF NOT FOUND THEN
        RAISE EXCEPTION
            'work_item_cancel: % not found or already in terminal status',
            p_work_item_id;
    END IF;
END;
$func$;

-- ---------------------------------------------------------------------
-- Views
-- ---------------------------------------------------------------------
CREATE OR REPLACE VIEW stewards.work_items_active AS
SELECT id, slug, pipeline_family, current_stage, status,
       jsonb_object_keys(stage_results) AS completed_stage,
       cardinality(session_ids) AS sessions_dispatched,
       tokens_in, tokens_out, token_budget, actor,
       created_at, updated_at
  FROM stewards.work_items
 WHERE status NOT IN ('completed', 'cancelled');

CREATE OR REPLACE VIEW stewards.work_items_summary AS
SELECT wi.id,
       wi.slug,
       wi.pipeline_family,
       wi.current_stage,
       wi.status,
       wi.created_at,
       wi.updated_at,
       wi.completed_at,
       (wi.completed_at - wi.created_at) AS elapsed,
       wi.tokens_in,
       wi.tokens_out,
       wi.token_budget,
       cardinality(wi.session_ids)            AS stages_dispatched,
       (SELECT count(*) FROM jsonb_object_keys(wi.stage_results)) AS stages_completed,
       (SELECT jsonb_array_length(p.stages) FROM stewards.pipelines p
         WHERE p.family = wi.pipeline_family) AS stages_total,
       wi.actor,
       wi.error
  FROM stewards.work_items wi;

-- ---------------------------------------------------------------------
-- Seed: echo-test pipeline (1 stage, smoke-test wiring)
-- ---------------------------------------------------------------------
INSERT INTO stewards.pipelines (family, description, stages)
VALUES (
    'echo-test',
    'Single-stage smoke test for Phase 3c.1. Dispatches one chat to stewards-explore to verify the pipeline → work_item → chat → completion wiring.',
    jsonb_build_array(
        jsonb_build_object(
            'name',         'echo',
            'agent_family', 'stewards-explore',
            'model',        'kimi-k2.6',
            'provider',     'opencode_go',
            'next',         null,
            'auto_advance', true
        )
    )
)
ON CONFLICT (family) DO UPDATE
   SET description = EXCLUDED.description,
       stages      = EXCLUDED.stages,
       updated_at  = now();
