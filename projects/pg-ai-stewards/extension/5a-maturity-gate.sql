-- =====================================================================
-- Phase 5a — Maturity ladder + gate machinery (Phase B from
-- full-agentic-substrate.md §IV)
--
-- Builds on:
--   - Phase 4 (cost tracking, escalation chain, steward, dispatch
--     override). Phase B doesn't depend on Phase A semantically but
--     reuses the work_item_dispatch_stage signature for gate calls.
--
-- This file adds the SCHEMA + CORE FUNCTIONS for the gate evaluator.
-- A separate file (5b) will wire scenarios + verify; bgworker auto-fire
-- comes in a follow-up push.
--
-- Per D-B1/D-B2/D-B3/D-B4 ratifications:
--   - Default gate model = OpenCode chain default for `_gate` family
--     (qwen3.6-plus for binary calls, kimi-k2.6 for scenarios)
--   - Revision cap = 2 → surface
--   - Scenarios LLM-generated, human-editable in UI before execute
--   - Maturity-to-stage mapping in config table (NOT naming convention)
--
-- This file adds:
--   1. work_items columns: maturity, scenarios, revision_count, spec,
--      destination_maturity
--   2. stewards.pipeline_stage_maturity — per-(family, stage) → maturity
--   3. stewards.gate_decisions — append-only audit ledger
--   4. stewards.gate_prompts — per-prompt-kind template seed
--   5. stewards.verify_results — per-work_item verify pass/fail records
--   6. stewards.evaluate_gate(work_item_id) — enqueues a gate-eval chat
--   7. stewards.apply_gate_decision(work_item_id, decision_jsonb)
--      — parses + writes audit + applies advance/revise/surface logic
--
-- Maturity ladder: raw → researched → planned → specced → executing → verified
--
-- The gate fires AT THE END of a stage (when a stage completes and that
-- stage produces a maturity). Three possible decisions:
--   advance → bump maturity, dispatch first stage of next maturity
--   revise → revision_count++ (cap at 2 → auto-surface), re-dispatch
--            current stage with feedback prepended
--   surface → set status='awaiting_review', no further auto-action
-- =====================================================================

-- ---------------------------------------------------------------------
-- Section 1: work_items columns
-- ---------------------------------------------------------------------

ALTER TABLE stewards.work_items
    ADD COLUMN IF NOT EXISTS maturity              text NOT NULL DEFAULT 'raw',
    ADD COLUMN IF NOT EXISTS scenarios             jsonb NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS revision_count        int NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS spec                  text,
    ADD COLUMN IF NOT EXISTS destination_maturity  text;

DO $check$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'work_items_maturity_check'
    ) THEN
        ALTER TABLE stewards.work_items
            ADD CONSTRAINT work_items_maturity_check
            CHECK (maturity IN
                ('raw','researched','planned','specced','executing','verified'));
    END IF;
END;
$check$;

DO $check2$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'work_items_destination_maturity_check'
    ) THEN
        ALTER TABLE stewards.work_items
            ADD CONSTRAINT work_items_destination_maturity_check
            CHECK (destination_maturity IS NULL OR destination_maturity IN
                ('researched','planned','specced','executing','verified'));
    END IF;
END;
$check2$;

COMMENT ON COLUMN stewards.work_items.maturity IS
'Phase 5a (Phase B): current maturity of the work_item. Advanced by gate decisions, NOT by stage transitions. Raw → researched → planned → specced → executing → verified.';
COMMENT ON COLUMN stewards.work_items.scenarios IS
'Phase 5a: LLM-generated acceptance criteria as a JSON array of strings. Populated when maturity advances to specced; verify checks against these.';
COMMENT ON COLUMN stewards.work_items.revision_count IS
'Phase 5a: how many times the gate has returned action=revise for this maturity. Capped at 2 → auto-surface (D-B2).';
COMMENT ON COLUMN stewards.work_items.spec IS
'Phase 5a: the canonical spec text for this work_item. Set during the specced maturity.';
COMMENT ON COLUMN stewards.work_items.destination_maturity IS
'Phase 5a: where the human wants this work_item to end. NULL = default (verified, full Ammon-loop). Set lower (e.g. specced) to surface for review before continuing.';

-- ---------------------------------------------------------------------
-- Section 2: pipeline_stage_maturity — per-stage maturity tag
-- ---------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS stewards.pipeline_stage_maturity (
    pipeline_family   text NOT NULL,
    stage_name        text NOT NULL,
    produces_maturity text NOT NULL CHECK (produces_maturity IN
        ('researched','planned','specced','executing','verified')),
    notes             text,
    PRIMARY KEY (pipeline_family, stage_name)
);

COMMENT ON TABLE stewards.pipeline_stage_maturity IS
'Phase 5a: per-(pipeline_family, stage) what maturity that stage produces. Gate fires when a stage completes that has a row here. NULL/missing row = stage doesn''t produce a maturity (intermediate stage).';

-- Seed for known pipelines. study-write goes outline→draft→review which
-- maps to planned→executing→verified (skipping researched + specced
-- because the existing 3-stage pipeline conflates them).
INSERT INTO stewards.pipeline_stage_maturity
    (pipeline_family, stage_name, produces_maturity, notes) VALUES
    ('study-write',      'outline', 'planned',  'planning + source survey'),
    ('study-write',      'draft',   'executing', 'drafting against the outline'),
    ('study-write',      'review',  'verified',  'self-review for voice + verification'),
    ('study-write-qwen', 'outline', 'planned',   'LM Studio variant'),
    ('study-write-qwen', 'draft',   'executing', 'LM Studio variant'),
    ('study-write-qwen', 'review',  'verified',  'LM Studio variant')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE
SET produces_maturity = EXCLUDED.produces_maturity,
    notes             = EXCLUDED.notes;

-- ---------------------------------------------------------------------
-- Section 3: gate_decisions — append-only audit ledger
-- ---------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS stewards.gate_decisions (
    id              bigserial PRIMARY KEY,
    work_item_id    uuid REFERENCES stewards.work_items(id) ON DELETE CASCADE,
    at              timestamptz NOT NULL DEFAULT now(),
    from_maturity   text NOT NULL,
    action          text NOT NULL CHECK (action IN ('advance','revise','surface')),
    reasoning       text,
    feedback        text,
    work_id         bigint,    -- the work_queue row that produced the decision
    revision_count  int NOT NULL DEFAULT 0,  -- snapshot at decision time
    raw_response    jsonb      -- full parsed response for audit
);
CREATE INDEX IF NOT EXISTS gate_decisions_work_item ON stewards.gate_decisions(work_item_id);
CREATE INDEX IF NOT EXISTS gate_decisions_at        ON stewards.gate_decisions(at);

COMMENT ON TABLE stewards.gate_decisions IS
'Phase 5a: append-only audit of every gate decision. Each row captures action (advance|revise|surface), reasoning, feedback, and snapshot of revision_count at decision time.';

-- ---------------------------------------------------------------------
-- Section 4: gate_prompts — per-prompt-kind template seed
-- ---------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS stewards.gate_prompts (
    id        text PRIMARY KEY CHECK (id IN ('evaluate','generate_scenarios','verify')),
    template  text NOT NULL,
    notes     text,
    updated_at timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE stewards.gate_prompts IS
'Phase 5a: per-prompt templates with {{placeholder}} syntax. evaluate_gate composes from these + work_item context.';

INSERT INTO stewards.gate_prompts (id, template, notes) VALUES
    ('evaluate',
$tmpl$You are a gate evaluator for a structured second-brain pipeline. Your job is to decide whether a piece of work has matured enough to advance, needs revision, or needs human steering.

Pipeline: {{pipeline_family}}
Current stage just completed: {{current_stage}}
Current maturity: {{maturity}}
Maturity this stage produces: {{produces_maturity}}
Revision count for this maturity: {{revision_count}}

Binding question / input:
{{input_summary}}

Latest stage output:
{{stage_output}}

Decide ONE of:
- "advance" — the work has clearly satisfied the criteria for this maturity. Move to the next stage / next maturity.
- "revise" — the work is on the right track but needs another pass. Provide specific, actionable feedback for what to improve.
- "surface" — the work needs human steering. Either it's ambiguous, hit a constraint you can't resolve, or the binding question shifted. Provide a brief explanation of what the human needs to decide.

Respond with JSON ONLY (no prose around it):
{
  "action": "advance" | "revise" | "surface",
  "reasoning": "1-3 sentences explaining the decision",
  "feedback": "if revise: what to do differently next pass; if surface: what the human needs to decide; if advance: omit or empty string"
}
$tmpl$,
     'Default gate evaluation prompt. Per D-C4: free-form rather than checklist; trusts the gate model to internalize covenant.'),

    ('generate_scenarios',
$tmpl$You are producing acceptance criteria for a piece of work that has just been spec''d.

Pipeline: {{pipeline_family}}
Binding question: {{input_summary}}
Spec / planning output:
{{spec_or_stage_output}}

Generate 3-7 testable acceptance criteria as a JSON array of strings. Each criterion should be SPECIFIC, VERIFIABLE, and OBSERVABLE in the eventual execution output. Avoid vague criteria like "the work is high quality"; prefer "the output cites at least 3 sources by name" or "the conclusion answers the binding question explicitly."

Respond with JSON ONLY:
{
  "scenarios": [
    "criterion 1 phrased as a checkable statement",
    "criterion 2 ...",
    ...
  ]
}
$tmpl$,
     'Generates acceptance criteria. Output stored in work_items.scenarios; human-editable before execute begins (D-B3).'),

    ('verify',
$tmpl$You are checking whether the execution output meets each acceptance criterion.

Pipeline: {{pipeline_family}}
Binding question: {{input_summary}}

Acceptance criteria:
{{scenarios}}

Execution output:
{{stage_output}}

For each criterion, judge whether the execution output satisfies it. Be strict — if a criterion isn't clearly met, mark it failed.

Respond with JSON ONLY:
{
  "all_passed": true | false,
  "reasoning": "1-2 sentence overall summary",
  "results": [
    {"scenario": "criterion text verbatim", "passed": true, "notes": "where this is evidenced or what's missing"},
    ...
  ]
}
$tmpl$,
     'Verifies execution output against scenarios. all_passed=false drops maturity back to planned with verify feedback.')
ON CONFLICT (id) DO UPDATE
SET template   = EXCLUDED.template,
    notes      = EXCLUDED.notes,
    updated_at = now();

-- ---------------------------------------------------------------------
-- Section 5: verify_results — per-work_item verify outcomes
-- ---------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS stewards.verify_results (
    id            bigserial PRIMARY KEY,
    work_item_id  uuid REFERENCES stewards.work_items(id) ON DELETE CASCADE,
    at            timestamptz NOT NULL DEFAULT now(),
    all_passed    boolean NOT NULL,
    reasoning     text,
    results       jsonb NOT NULL DEFAULT '[]'::jsonb,
    work_id       bigint
);
CREATE INDEX IF NOT EXISTS verify_results_work_item ON stewards.verify_results(work_item_id);

COMMENT ON TABLE stewards.verify_results IS
'Phase 5a: per-work_item verify pass/fail records. all_passed=false → maturity drops back to planned with results as feedback for re-execute.';

-- ---------------------------------------------------------------------
-- Section 6: render_template — minimal {{placeholder}} substitution
-- ---------------------------------------------------------------------

-- Renders a template with placeholder substitution. NOT a full template
-- engine — just replace(template, '{{key}}', value) for each (key, value)
-- in the kv jsonb. NULL values become empty strings.
CREATE OR REPLACE FUNCTION stewards.render_template(
    p_template text,
    p_kv       jsonb
) RETURNS text
LANGUAGE plpgsql IMMUTABLE AS $func$
DECLARE
    v_out text := p_template;
    v_key text;
    v_val text;
BEGIN
    IF p_kv IS NULL THEN
        RETURN v_out;
    END IF;
    FOR v_key, v_val IN
        SELECT key, coalesce(value::text, '')
          FROM jsonb_each_text(p_kv)
    LOOP
        v_out := replace(v_out, '{{' || v_key || '}}', v_val);
    END LOOP;
    RETURN v_out;
END;
$func$;

COMMENT ON FUNCTION stewards.render_template(text, jsonb) IS
'Phase 5a: minimal {{key}} → value substitution for prompt templates. NOT a full template engine.';

-- ---------------------------------------------------------------------
-- Section 7: evaluate_gate(work_item_id) — enqueue gate-eval chat
-- ---------------------------------------------------------------------

-- Enqueues a chat against the _gate/evaluate_gate stage_models default
-- (Qwen3.6 Plus per seed) using the 'evaluate' gate prompt template.
-- The chat work_queue row is marked with payload._gate_eval=true so the
-- response handler (next push) knows to parse + apply automatically.
--
-- Returns the bigint work_id of the enqueued chat. Caller can poll
-- work_queue / messages to see the result, then call
-- apply_gate_decision(work_item_id, decision_jsonb) to act on it.
CREATE OR REPLACE FUNCTION stewards.evaluate_gate(
    p_work_item_id uuid
) RETURNS bigint
LANGUAGE plpgsql AS $func$
DECLARE
    v_wi              stewards.work_items%ROWTYPE;
    v_pipeline_stage  jsonb;
    v_produces_maturity text;
    v_template        text;
    v_input_summary   text;
    v_stage_output    text;
    v_prompt          text;
    v_session_id      text;
    v_payload         jsonb;
    v_work_id         bigint;
    v_gate_model      text := 'qwen3.6-plus';      -- _gate/evaluate_gate default
    v_gate_provider   text := 'opencode_go';
    v_gate_agent      text := 'plan';              -- reuse plan agent for now
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN
        RAISE EXCEPTION 'work_item % not found', p_work_item_id;
    END IF;

    SELECT produces_maturity INTO v_produces_maturity
      FROM stewards.pipeline_stage_maturity
     WHERE pipeline_family = v_wi.pipeline_family
       AND stage_name = v_wi.current_stage;
    -- v_produces_maturity may be NULL (intermediate stage); evaluate_gate
    -- still works, just no maturity advance happens on 'advance' decision.

    SELECT template INTO v_template
      FROM stewards.gate_prompts WHERE id = 'evaluate';
    IF v_template IS NULL THEN
        RAISE EXCEPTION 'gate_prompts.evaluate template missing';
    END IF;

    -- Compose context
    v_input_summary := substring(coalesce(v_wi.input::text, ''), 1, 2000);
    v_stage_output  := substring(
        coalesce(v_wi.stage_results->v_wi.current_stage->>'output', ''),
        1, 8000);

    v_prompt := stewards.render_template(v_template, jsonb_build_object(
        'pipeline_family',   v_wi.pipeline_family,
        'current_stage',     v_wi.current_stage,
        'maturity',          v_wi.maturity,
        'produces_maturity', coalesce(v_produces_maturity, '(none)'),
        'revision_count',    v_wi.revision_count::text,
        'input_summary',     v_input_summary,
        'stage_output',      v_stage_output
    ));

    -- Session for the gate eval — separate from the work_item's main
    -- session so audit stays clean.
    v_session_id := substring(
        'wi--' || substring(v_wi.id::text FROM 1 FOR 8) || '--gate-' ||
        v_wi.maturity || '--' ||
        to_char(extract(epoch from now())::bigint, 'FM9999999999'),
        1, 200);

    INSERT INTO stewards.sessions (id, label, kind)
    VALUES (v_session_id,
            format('gate eval work_item=%s maturity=%s', v_wi.id, v_wi.maturity),
            'gate')
    ON CONFLICT (id) DO NOTHING;

    -- Insert the user message (the gate prompt)
    INSERT INTO stewards.messages (session_id, role, content, model)
    VALUES (v_session_id, 'user', v_prompt, v_gate_model);

    -- Compose chat body via existing dry_run_chat
    v_payload := jsonb_build_object(
        'session_id',         v_session_id,
        'agent_family',       v_gate_agent,
        'requested_model',    v_gate_model,
        'meta',               '{}'::jsonb,
        'body',               (stewards.dry_run_chat(v_gate_agent, v_gate_model, v_session_id, NULL) - '_meta')
                              || jsonb_build_object('user', v_session_id),
        '_work_item_id',      p_work_item_id::text,
        '_stage_name',        v_wi.current_stage,
        '_pipeline_family',   v_wi.pipeline_family,
        '_gate_eval',         true,
        '_gate_from_maturity', v_wi.maturity
    );

    INSERT INTO stewards.work_queue (kind, provider, payload)
    VALUES ('chat', v_gate_provider, v_payload)
    RETURNING id INTO v_work_id;

    RETURN v_work_id;
END;
$func$;

COMMENT ON FUNCTION stewards.evaluate_gate(uuid) IS
'Phase 5a: enqueues a gate-eval chat for a work_item. Returns the work_queue id; caller polls/waits then calls apply_gate_decision with the parsed JSON response. Bgworker auto-fire of apply_gate_decision lands in a follow-up push.';

-- ---------------------------------------------------------------------
-- Section 8: apply_gate_decision(work_item_id, decision_jsonb)
-- ---------------------------------------------------------------------

-- Applies a parsed gate decision: writes audit row, then transitions
-- the work_item per the action.
--
-- decision_jsonb shape:
--   {"action": "advance"|"revise"|"surface",
--    "reasoning": "...",
--    "feedback": "..."}
--
-- Side effects:
--   - INSERT into gate_decisions
--   - On 'advance': bump maturity to next in ladder (if pipeline_stage_maturity
--     row exists), reset revision_count, clear surface flag if any
--   - On 'revise': revision_count++; if >2 force-surface; else set work_item
--     status='failed' + last_failure_reason='gate revise: <feedback>' so
--     the steward picks it up and re-dispatches
--   - On 'surface': status='awaiting_review'
--
-- Returns the new maturity (or current if no transition).
CREATE OR REPLACE FUNCTION stewards.apply_gate_decision(
    p_work_item_id uuid,
    p_decision     jsonb,
    p_work_id      bigint DEFAULT NULL
) RETURNS text
LANGUAGE plpgsql AS $func$
DECLARE
    v_wi             stewards.work_items%ROWTYPE;
    v_action         text;
    v_reasoning      text;
    v_feedback       text;
    v_new_maturity   text;
    v_produces_mat   text;
    v_maturity_order text[] := ARRAY['raw','researched','planned','specced','executing','verified'];
    v_idx            int;
    v_new_revision   int;
BEGIN
    SELECT * INTO v_wi FROM stewards.work_items WHERE id = p_work_item_id;
    IF v_wi.id IS NULL THEN
        RAISE EXCEPTION 'work_item % not found', p_work_item_id;
    END IF;

    v_action    := p_decision->>'action';
    v_reasoning := p_decision->>'reasoning';
    v_feedback  := p_decision->>'feedback';

    IF v_action NOT IN ('advance', 'revise', 'surface') THEN
        RAISE EXCEPTION 'apply_gate_decision: invalid action %', v_action;
    END IF;

    -- Audit first
    INSERT INTO stewards.gate_decisions
        (work_item_id, from_maturity, action, reasoning, feedback,
         work_id, revision_count, raw_response)
    VALUES
        (p_work_item_id, v_wi.maturity, v_action, v_reasoning, v_feedback,
         p_work_id, v_wi.revision_count, p_decision);

    v_new_maturity := v_wi.maturity;

    IF v_action = 'advance' THEN
        -- If this stage produces a maturity, bump to it
        SELECT produces_maturity INTO v_produces_mat
          FROM stewards.pipeline_stage_maturity
         WHERE pipeline_family = v_wi.pipeline_family
           AND stage_name = v_wi.current_stage;

        IF v_produces_mat IS NOT NULL THEN
            v_new_maturity := v_produces_mat;
        ELSE
            -- No mapping: bump to next maturity in the ladder
            v_idx := array_position(v_maturity_order, v_wi.maturity);
            IF v_idx IS NOT NULL AND v_idx < array_length(v_maturity_order, 1) THEN
                v_new_maturity := v_maturity_order[v_idx + 1];
            END IF;
        END IF;

        UPDATE stewards.work_items
           SET maturity       = v_new_maturity,
               revision_count = 0,
               updated_at     = now()
         WHERE id = p_work_item_id;

        -- Note: actually advancing the STAGE is a separate step
        -- (work_item_advance) — gate ratification doesn't auto-advance
        -- stages. Bgworker integration (next push) decides whether
        -- to call work_item_advance based on auto_advance flag.

    ELSIF v_action = 'revise' THEN
        v_new_revision := v_wi.revision_count + 1;

        IF v_new_revision > 2 THEN
            -- Cap exceeded — auto-surface per D-B2
            UPDATE stewards.work_items
               SET status = 'awaiting_review',
                   revision_count = v_new_revision,
                   updated_at = now()
             WHERE id = p_work_item_id;
        ELSE
            -- Mark as failed so the steward picks it up next tick.
            -- last_failure_reason carries the gate's feedback text.
            UPDATE stewards.work_items
               SET status                 = 'failed',
                   revision_count         = v_new_revision,
                   last_failure_reason    = 'gate revise: ' || coalesce(v_feedback, '(no feedback)'),
                   last_failure_diagnosis = 'gate_revise',
                   updated_at             = now()
             WHERE id = p_work_item_id;
        END IF;

    ELSIF v_action = 'surface' THEN
        UPDATE stewards.work_items
           SET status     = 'awaiting_review',
               updated_at = now()
         WHERE id = p_work_item_id;
    END IF;

    RETURN v_new_maturity;
END;
$func$;

COMMENT ON FUNCTION stewards.apply_gate_decision(uuid, jsonb, bigint) IS
'Phase 5a: applies a parsed gate decision (advance|revise|surface) to a work_item. Writes audit row, transitions maturity/status/revision_count per action. Returns new maturity. On revise, sets status=failed so the steward retry path picks it up; cap of 2 revisions per D-B2 → auto-surface.';

-- ---------------------------------------------------------------------
-- Section 9: parse_gate_response_from_message(work_id) — helper
-- ---------------------------------------------------------------------

-- Reads the assistant message produced by a gate-eval chat (work_queue id),
-- extracts the JSON decision, and returns it. Returns NULL if no message
-- exists yet (chat still pending) or if the response isn't valid JSON.
CREATE OR REPLACE FUNCTION stewards.parse_gate_response(
    p_work_id bigint
) RETURNS jsonb
LANGUAGE plpgsql STABLE AS $func$
DECLARE
    v_session_id text;
    v_content    text;
    v_json_start int;
    v_json_end   int;
    v_candidate  text;
    v_parsed     jsonb;
BEGIN
    -- The session_id is in the work_queue payload
    SELECT (payload->>'session_id') INTO v_session_id
      FROM stewards.work_queue
     WHERE id = p_work_id;
    IF v_session_id IS NULL THEN
        RETURN NULL;
    END IF;

    -- Most recent assistant message in that session
    SELECT content INTO v_content
      FROM stewards.messages
     WHERE session_id = v_session_id AND role = 'assistant'
     ORDER BY id DESC LIMIT 1;
    IF v_content IS NULL OR length(trim(v_content)) = 0 THEN
        RETURN NULL;
    END IF;

    -- Try to extract JSON object from the response. Models often
    -- wrap with prose; pull from first { to last } as a heuristic.
    v_json_start := position('{' in v_content);
    v_json_end := length(v_content) - position('}' in reverse(v_content)) + 1;
    IF v_json_start = 0 OR v_json_end < v_json_start THEN
        RETURN NULL;
    END IF;
    v_candidate := substring(v_content FROM v_json_start FOR v_json_end - v_json_start + 1);

    BEGIN
        v_parsed := v_candidate::jsonb;
    EXCEPTION WHEN OTHERS THEN
        RETURN NULL;
    END;

    RETURN v_parsed;
END;
$func$;

COMMENT ON FUNCTION stewards.parse_gate_response(bigint) IS
'Phase 5a: reads the assistant message for a gate-eval work_queue id, extracts the JSON decision (heuristic: first { to last }), returns parsed jsonb or NULL.';

-- =====================================================================
-- Done. Phase 5a maturity + gate machinery is operational.
--
-- Manual flow for testing (bgworker auto-fire is next push):
--   1. SELECT stewards.evaluate_gate('<work_item_uuid>');
--      → returns work_id; bgworker dispatches the chat
--   2. wait for chat to complete (poll stewards.work_queue.status='done')
--   3. SELECT stewards.parse_gate_response(<work_id>);
--      → returns parsed jsonb decision
--   4. SELECT stewards.apply_gate_decision('<work_item_uuid>',
--                                          <jsonb_from_step_3>,
--                                          <work_id>);
--      → writes gate_decisions row + transitions work_item
-- =====================================================================
