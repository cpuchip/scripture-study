-- =====================================================================
-- Batch K.4 — spawn_subagent primitive (substrate side)
-- =====================================================================
-- SQL function spawn_subagent_create() creates a child work_item with
-- parent_work_item_id, dispatches the first stage, and returns the
-- child id. The Go handler in cmd/stewards-mcp/spawn_subagent.go does
-- the synchronous polling and digest extraction.
--
-- Separation of concerns:
--   - SQL: create + dispatch (transactional, fast)
--   - Go:  wait + digest (long-running, can be polled)
--
-- Cost-cap protection: every spawned child gets a default
-- cost_cap_micro of $0.50 unless caller overrides. Combined with the
-- pipeline-level cost tracking, this bounds runaway cost.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. spawn_subagent_create(pipeline_family, binding_question,
--                          parent_work_item_id?, cost_cap_micro?,
--                          project_association?)
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.spawn_subagent_create(
    p_pipeline_family    text,
    p_binding_question   text,
    p_parent_work_item_id uuid DEFAULT NULL,
    p_cost_cap_micro     bigint DEFAULT 500000,
    p_project_association text DEFAULT NULL,
    p_slug               text DEFAULT NULL,
    p_actor              text DEFAULT 'subagent'
) RETURNS uuid LANGUAGE plpgsql AS $FN$
DECLARE
    v_parent       stewards.work_items%ROWTYPE;
    v_child_id     uuid;
    v_intent_id    uuid;
    v_actor        text;
    v_project      text;
    v_slug         text;
BEGIN
    -- Inherit intent + actor + project from parent if given; otherwise
    -- fall back to scripture-study intent + the supplied actor.
    IF p_parent_work_item_id IS NOT NULL THEN
        SELECT * INTO v_parent FROM stewards.work_items WHERE id = p_parent_work_item_id;
        IF v_parent.id IS NULL THEN
            RAISE EXCEPTION 'spawn_subagent_create: parent % not found', p_parent_work_item_id;
        END IF;
        v_intent_id := v_parent.intent_id;
        v_actor     := COALESCE(p_actor, v_parent.actor);
        v_project   := COALESCE(p_project_association, v_parent.project_association);
    ELSE
        SELECT id INTO v_intent_id FROM stewards.intents WHERE slug='scripture-study' LIMIT 1;
        v_actor   := COALESCE(p_actor, 'subagent');
        v_project := p_project_association;
    END IF;

    v_slug := COALESCE(p_slug, 'subagent-' || to_char(now() AT TIME ZONE 'UTC', 'YYYYMMDD-HH24MISS-MS'));

    -- Create the child via the standard primitive.
    v_child_id := stewards.work_item_create(
        p_pipeline_family => p_pipeline_family,
        p_input           => jsonb_build_object('binding_question', p_binding_question),
        p_slug            => v_slug,
        p_actor           => v_actor,
        p_intent_id       => v_intent_id
    );

    UPDATE stewards.work_items
       SET parent_work_item_id = p_parent_work_item_id,
           project_association = v_project,
           cost_cap_micro      = COALESCE(p_cost_cap_micro, cost_cap_micro),
           origin              = 'agent_planning'   -- treated as agent-spawned work
     WHERE id = v_child_id;

    -- Dispatch the first stage. work_item_dispatch_stage handles the
    -- session_id allocation and the actual chat work_queue enqueue.
    PERFORM stewards.work_item_dispatch_stage(v_child_id, NULL);

    RAISE NOTICE 'spawn_subagent_create: parent=% child=% pipeline=% slug=% cost_cap=%',
        p_parent_work_item_id, v_child_id, p_pipeline_family, v_slug,
        COALESCE(p_cost_cap_micro, 0);

    RETURN v_child_id;
END;
$FN$;

COMMENT ON FUNCTION stewards.spawn_subagent_create(text, text, uuid, bigint, text, text, text) IS
'Batch K.4: substrate-side primitive — creates a child work_item with parent linkage and dispatches its first stage. Returns the child uuid. The Go handler in stewards-mcp does the synchronous wait + digest extraction.';


-- ---------------------------------------------------------------------
-- 2. Tool definition for spawn_subagent.
-- ---------------------------------------------------------------------

INSERT INTO stewards.tool_defs (name, description, args_schema, execute_target, active)
VALUES (
    'spawn_subagent',
    'Delegate verbose / multi-turn work to a child agent that runs in its own isolated context. ' ||
    'The child uses up to its own 200K-token context exploring the binding_question; you only see the digest it returns. ' ||
    'Use for: deep research across multiple sources, audits over many files, surveys of related sessions. ' ||
    'DO NOT use for: a single cheap tool call (overhead exceeds savings), or work that needs to read/write your active state.',
    $JSON$
    {
      "type": "object",
      "required": ["pipeline_family", "binding_question"],
      "additionalProperties": false,
      "properties": {
        "pipeline_family": {
          "type": "string",
          "description": "Which pipeline the sub-agent runs. Common: 'research-write' (broad sourced research), 'study-write' (scripture study), or any other registered pipeline."
        },
        "binding_question": {
          "type": "string",
          "description": "The specific question the sub-agent should answer. Be tightly scoped — the sub-agent's whole context is built around this."
        },
        "cost_cap_micro": {
          "type": "integer",
          "default": 500000,
          "description": "Max micro-dollars the sub-agent may spend (default 500000 = $0.50). Higher caps for genuinely heavy work."
        },
        "project_association": {
          "type": "string",
          "description": "Optional project slug; inherits from parent if not set."
        },
        "slug": {
          "type": "string",
          "description": "Optional slug for the spawned work_item; auto-generated if not provided."
        }
      }
    }
    $JSON$::jsonb,
    jsonb_build_object('kind', 'mcp_proxy', 'server', 'pg-ai-stewards', 'tool', 'spawn_subagent'),
    true
)
ON CONFLICT (name) DO UPDATE
   SET description = EXCLUDED.description,
       args_schema = EXCLUDED.args_schema,
       execute_target = EXCLUDED.execute_target,
       active = true;


-- =====================================================================
-- End of k4-spawn-subagent.sql
-- =====================================================================
