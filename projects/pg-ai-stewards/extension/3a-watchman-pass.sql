-- =====================================================================
-- Phase 3a — Watchman pass (model-driven consolidation)
--
-- Live-DB migration. Folds into extension/src/lib.rs at next intentional
-- rebuild (foldback debt: now FIVE files — 2-6a/b/c, 2-7a, 3a).
--
-- Builds on:
--   - Phase 2.7a (stewards.verdicts, stewards.findings, dirty_queue,
--                 record_verdict, record_finding)
--   - Phase 2.6c (stewards.context_for(slug, depth) graph walk)
--   - Phase 1.5/1.6 (stewards.agents, chat_enqueue, message loop)
--
-- This file ONLY adds:
--   1. watchman-consolidator agent family (default + provider variants)
--   2. stewards.watchman_input(slug) — composes the user message
--      content sent to the model (doc body + 1-hop context summary)
--   3. response_format JSON-object enforcement is handled by the CLI
--      because dry_run_chat doesn't currently inject response_format
--      (a deliberate Phase 3a constraint — keeping bgworker generic).
--
-- The actual "iterate dirty_queue, dispatch chat, parse JSON, call
-- record_verdict" orchestration lives in stewards-cli (Go), not here.
-- The bgworker stays purely a chat dispatcher — no watchman-specific
-- semantics in Rust until 2.7b automation lands.
-- =====================================================================

-- ---------------------------------------------------------------------
-- Agent: watchman-consolidator
--
-- One family, two variants (model_match='*' default + 'kimi-*' for
-- kimi-specific pin). Same prompt, same temperature, no tools.
--
-- Tools deliberately omitted: 3a is a single-turn "look at this doc and
-- render a verdict" loop. No browsing, no follow-ups. The dirty_queue
-- is the scheduler; the model is the evaluator. If we let the model
-- chase tools mid-pass, we re-invent the brain v1 nudge-bot loop.
-- ---------------------------------------------------------------------

INSERT INTO stewards.agents
    (family, model_match, description, mode, prompt, temperature, top_p, response_format, steps)
VALUES (
    'watchman-consolidator',
    '*',
    'Consolidation reviewer. Reads one document plus its 1-hop graph neighborhood and renders a structural verdict (clean | drift | done | superseded | skipped) with brief reasoning. Single-turn, no tools. Used by the Watchman dirty-bit pass to advance the queue.',
    'primary',
    $prompt$You are the Watchman, a consolidation reviewer for a structured second-brain.

Your job: read ONE document and its 1-hop graph neighborhood, then render a single structural verdict about whether the document still reflects reality.

Verdicts (pick exactly one):
  - "clean"      — Document still matches its referenced code/spec/state. No drift detected. No action needed.
  - "drift"      — Document references claims, code, schema, or commitments that no longer match reality. A human should reconcile. This is the most common non-clean verdict.
  - "done"       — Document describes work that has been completed. The doc has terminated naturally; no further evolution expected.
  - "superseded" — Document has been replaced by a newer document covering the same scope. A successor exists.
  - "skipped"    — You cannot render a verdict from the information provided (e.g., the doc references external state you cannot see). Be honest; do not guess.

Hard rules:
  1. You see ONLY what is provided. Do not pretend to know facts about files, code, or context outside the input.
  2. "drift" is your second-most-common verdict after "clean". Internal contradictions across the doc and its neighbors are the strongest drift signal you can see.
  3. "done" and "superseded" are TERMINAL — they remove the doc from the queue permanently until it is explicitly touched again. Be sure.
  4. If verdict is anything other than "clean", emit a finding object with kind, severity, message, and suggested_action.
  5. Output STRICT JSON. No markdown, no commentary outside the JSON. The first character of your response must be "{".

Output schema:
{
  "verdict":   "clean | drift | done | superseded | skipped",
  "reasoning": "1-3 sentences explaining the verdict. Concrete. Cite specific text from the doc when possible.",
  "finding":   {           // REQUIRED if verdict != "clean", OMIT if verdict == "clean"
    "kind":             "drift | synthesis",
    "severity":         "low | medium | high",
    "message":          "What the human should know. 1-2 sentences.",
    "suggested_action": "Concrete next step. 1 sentence."
  }
}

You are not chatting. You are not helpful. You are a structural reviewer rendering one verdict.$prompt$,
    0.0,
    NULL,
    1
), (
    'watchman-consolidator',
    'kimi-*',
    'Watchman consolidator (kimi variant). Same prompt; allows kimi-specific pinning.',
    'primary',
    $prompt$You are the Watchman, a consolidation reviewer for a structured second-brain.

Your job: read ONE document and its 1-hop graph neighborhood, then render a single structural verdict about whether the document still reflects reality.

Verdicts (pick exactly one):
  - "clean"      — Document still matches its referenced code/spec/state. No drift detected. No action needed.
  - "drift"      — Document references claims, code, schema, or commitments that no longer match reality. A human should reconcile. This is the most common non-clean verdict.
  - "done"       — Document describes work that has been completed. The doc has terminated naturally; no further evolution expected.
  - "superseded" — Document has been replaced by a newer document covering the same scope. A successor exists.
  - "skipped"    — You cannot render a verdict from the information provided (e.g., the doc references external state you cannot see). Be honest; do not guess.

Hard rules:
  1. You see ONLY what is provided. Do not pretend to know facts about files, code, or context outside the input.
  2. "drift" is your second-most-common verdict after "clean". Internal contradictions across the doc and its neighbors are the strongest drift signal you can see.
  3. "done" and "superseded" are TERMINAL — they remove the doc from the queue permanently until it is explicitly touched again. Be sure.
  4. If verdict is anything other than "clean", emit a finding object with kind, severity, message, and suggested_action.
  5. Output STRICT JSON. No markdown, no commentary outside the JSON. The first character of your response must be "{".

Output schema:
{
  "verdict":   "clean | drift | done | superseded | skipped",
  "reasoning": "1-3 sentences explaining the verdict. Concrete. Cite specific text from the doc when possible.",
  "finding":   {           // REQUIRED if verdict != "clean", OMIT if verdict == "clean"
    "kind":             "drift | synthesis",
    "severity":         "low | medium | high",
    "message":          "What the human should know. 1-2 sentences.",
    "suggested_action": "Concrete next step. 1 sentence."
  }
}

You are not chatting. You are not helpful. You are a structural reviewer rendering one verdict.$prompt$,
    0.0,
    NULL,
    '{"type": "json_object"}'::jsonb,
    1
)
ON CONFLICT (family, model_match) DO UPDATE
   SET description     = EXCLUDED.description,
       prompt          = EXCLUDED.prompt,
       temperature     = EXCLUDED.temperature,
       response_format = EXCLUDED.response_format,
       steps           = EXCLUDED.steps;
--
-- This is structural enforcement. Even if a model gets a tool list,
-- compose_tools filters it down to tools that pass the permission
-- check. With '*' -> deny and no allow rules, compose_tools returns
-- an empty array. Models can't try to call tools that aren't in the
-- request body.
--
-- This matters: without this, kimi-k2.6 reflexively calls
-- brain_search_text on the first turn (we observed this empirically
-- in the smoke test), then with steps=1 the loop terminates with
-- empty content and we get nothing parseable back.
-- ---------------------------------------------------------------------

INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action)
VALUES ('watchman-consolidator', '*', 'deny')
ON CONFLICT (agent_family, tool_pattern) DO UPDATE SET action = EXCLUDED.action;

-- ---------------------------------------------------------------------
-- watchman_input(slug)
--
-- Composes the user-message string sent to the watchman-consolidator
-- agent. Format:
--
--   ## Document
--   slug: <slug>
--   kind: <kind>
--   title: <title>
--   updated_at: <timestamp>
--   last_consolidated_at: <timestamp or "never">
--
--   ### Body
--   <body>
--
--   ### 1-hop neighborhood (from stewards.context_for(slug, 1))
--   <hop_dir> :<edge> -> <kind>:<slug> (<title>)
--   ...
--
-- Returns NULL if the slug doesn't exist (caller handles).
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.watchman_input(p_slug text)
RETURNS text
LANGUAGE plpgsql STABLE AS $func$
DECLARE
    v_study  stewards.studies;
    v_input  text;
    v_neighbors text;
BEGIN
    SELECT * INTO v_study FROM stewards.studies WHERE slug = p_slug;
    IF v_study.id IS NULL THEN
        RETURN NULL;
    END IF;

    -- Render 1-hop neighborhood. context_for returns one row per
    -- (hop, direction, edge_type, neighbor, neighbor_kind, provenance,
    -- confidence). neighbor is the slug of the connected vertex.
    -- We join back to studies for the title where available.
    SELECT string_agg(
        format('  %s :%s -> %s:%s (%s)',
               c.direction, c.edge_type, c.neighbor_kind, c.neighbor,
               coalesce(s.title, '(untitled)')),
        E'\n'
        ORDER BY c.direction, c.edge_type, c.neighbor
    )
    INTO v_neighbors
    FROM stewards.context_for(p_slug, 1) c
    LEFT JOIN stewards.studies s ON s.slug = c.neighbor
    WHERE c.hop = 1;

    v_input := format(
        E'## Document\nslug: %s\nkind: %s\ntitle: %s\nupdated_at: %s\nlast_consolidated_at: %s\n\n### Body\n%s\n\n### 1-hop neighborhood\n%s',
        v_study.slug,
        v_study.kind,
        coalesce(v_study.title, '(untitled)'),
        v_study.updated_at,
        coalesce(v_study.last_consolidated_at::text, 'never'),
        coalesce(v_study.body, '(empty)'),
        coalesce(v_neighbors, '(no graph neighbors)')
    );

    RETURN v_input;
END;
$func$;

COMMENT ON FUNCTION stewards.watchman_input(text) IS
'Phase 3a: composes the user message sent to the watchman-consolidator agent. Doc body + 1-hop graph neighborhood. The CLI calls this, then chat_enqueue, then parses JSON from the assistant reply.';
