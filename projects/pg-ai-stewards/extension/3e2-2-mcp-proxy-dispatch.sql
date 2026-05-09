-- =====================================================================
-- Phase 3e.2.b — MCP proxy dispatch substrate
--
-- This migration wires the substrate side of execute_target='mcp_proxy'.
-- The Rust bgworker emits child mcp_proxy work_queue rows when it sees
-- a tool_call whose tool_def routes to mcp_proxy. The Go bridge daemon
-- (Phase 3e.2.c, see cmd/stewards-mcp/bridge.go `bridge run`) claims
-- those child rows, calls the underlying MCP server, and writes the
-- result back. The bgworker's tool_dispatch row sits in a new
-- 'waiting_for_tools' status until all its mcp_proxy children resolve;
-- a completion pass in the bgworker tick loop then synthesizes the
-- final tool_messages and continues the chat.
--
-- This async fan-out pattern was chosen over block-poll (2026-05-08)
-- to avoid stalling bgworkers on outstanding bridge calls. With 4
-- workers and parallel children, a chat with 3 mcp_proxy tool_calls
-- can resolve them concurrently rather than serially.
-- =====================================================================

-- ---------------------------------------------------------------------
-- 1. Extend work_queue.status CHECK to include 'waiting_for_tools'
--
-- Idempotent: drop and re-add the constraint with the expanded list.
-- The constraint name pgrx assigned at create-time was the implicit
-- "work_queue_status_check"; we look it up rather than assume.
-- ---------------------------------------------------------------------
DO $$
DECLARE
    cn text;
BEGIN
    SELECT c.conname INTO cn
    FROM pg_constraint c
    JOIN pg_class t ON c.conrelid = t.oid
    JOIN pg_namespace n ON t.relnamespace = n.oid
    WHERE n.nspname = 'stewards'
      AND t.relname = 'work_queue'
      AND c.contype = 'c'
      AND pg_get_constraintdef(c.oid) ILIKE '%status%pending%';

    IF cn IS NOT NULL THEN
        EXECUTE format('ALTER TABLE stewards.work_queue DROP CONSTRAINT %I', cn);
    END IF;
END $$;

ALTER TABLE stewards.work_queue
    ADD CONSTRAINT work_queue_status_check
    CHECK (status IN ('pending', 'in_progress', 'waiting_for_tools', 'done', 'error'));

-- ---------------------------------------------------------------------
-- 2. mcp_proxy_enqueue — substrate-internal API
--
-- Inserts a child work_queue row of kind='mcp_proxy' with payload
-- describing which MCP server + tool to call and the tool's args.
-- The provider column carries the server name so operators can grep
-- the queue easily. NOTIFY wakes the bridge daemon immediately
-- (otherwise it would tick on its 1s default).
--
-- Returns the new row's id. The caller (Rust exec_one_tool::mcp_proxy
-- arm) records this id in its parent tool_dispatch row's result jsonb
-- so the completion handler knows which child belongs to which
-- tool_call_id.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.mcp_proxy_enqueue(
    p_server   text,
    p_tool     text,
    p_args     jsonb,
    p_parent_tool_dispatch_id bigint  -- nullable; for synthetic tests
) RETURNS bigint
LANGUAGE plpgsql
AS $$
DECLARE
    new_id bigint;
BEGIN
    -- Defensive: refuse to enqueue against a server that doesn't exist
    -- or isn't enabled. The bridge would refuse anyway, but failing
    -- here gives the parent tool_dispatch immediate feedback rather
    -- than a 30s timeout.
    IF NOT EXISTS (
        SELECT 1 FROM stewards.mcp_servers
        WHERE name = p_server AND enabled
    ) THEN
        RAISE EXCEPTION 'mcp_proxy_enqueue: server % is not registered or not enabled', p_server;
    END IF;

    INSERT INTO stewards.work_queue (kind, provider, payload)
    VALUES (
        'mcp_proxy',
        p_server,
        jsonb_build_object(
            'server',                  p_server,
            'tool',                    p_tool,
            'args',                    p_args,
            'parent_tool_dispatch_id', p_parent_tool_dispatch_id
        )
    )
    RETURNING id INTO new_id;

    -- NOTIFY payload is the row id as text. Bridge LISTENs and uses
    -- it as a hint to immediately try claiming (it claims the OLDEST
    -- pending mcp_proxy regardless, so race-safe under concurrent
    -- producers).
    PERFORM pg_notify('stewards_mcp_proxy', new_id::text);

    RETURN new_id;
END
$$;

COMMENT ON FUNCTION stewards.mcp_proxy_enqueue IS
  'Enqueue a child work_queue row of kind=mcp_proxy and notify the '
  'bridge daemon. Used by the Rust bgworker''s exec_one_tool::mcp_proxy '
  'arm during async tool fan-out. Synthetic callers (verify scripts) '
  'can pass NULL for p_parent_tool_dispatch_id.';

-- ---------------------------------------------------------------------
-- 3. tool_dispatch_complete_waiting — completion pass
--
-- Scans for tool_dispatch rows in 'waiting_for_tools' status, checks
-- whether all their mcp_proxy children are done/errored, and if so
-- collects the children's results, builds tool_messages, runs the
-- original Phase 3 logic (insert role='tool' messages, enqueue
-- continuation chat), and transitions the parent to 'done'.
--
-- Implemented in SQL because the per-row logic is SPI-heavy and
-- duplicating it in Rust would require porting compose_messages /
-- chat_post_internal call sites that already exist as SQL fns.
-- The Rust tick loop just calls this on each tick.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.tool_dispatch_complete_waiting()
RETURNS integer
LANGUAGE plpgsql
AS $$
DECLARE
    parent_row    record;
    child_row     record;
    resolved_arr  jsonb;
    pending_arr   jsonb;
    pending_elem  jsonb;
    all_done      boolean;
    final_msgs    jsonb := '[]'::jsonb;
    completed_n   integer := 0;
    chat_work_id  bigint;
    parent_chat_id bigint;
    parent_session text;
    parent_family  text;
    parent_model   text;
    parent_provider text;
BEGIN
    -- Iterate waiting tool_dispatch rows. SKIP LOCKED so concurrent
    -- workers running this same function don't block each other.
    FOR parent_row IN
        SELECT id, payload, result, provider
          FROM stewards.work_queue
         WHERE kind = 'tool_dispatch'
           AND status = 'waiting_for_tools'
         ORDER BY created_at
         FOR UPDATE SKIP LOCKED
    LOOP
        resolved_arr := coalesce(parent_row.result -> 'resolved', '[]'::jsonb);
        pending_arr  := coalesce(parent_row.result -> 'pending',  '[]'::jsonb);
        all_done := true;
        final_msgs := '[]'::jsonb;

        -- Re-merge resolved (sync) replies first.
        final_msgs := resolved_arr;

        -- For each pending entry, look up the child's status.
        FOR pending_elem IN SELECT * FROM jsonb_array_elements(pending_arr)
        LOOP
            SELECT id, status, result, error
              INTO child_row
              FROM stewards.work_queue
             WHERE id = (pending_elem ->> 'child_work_id')::bigint;

            IF child_row.status NOT IN ('done', 'error') THEN
                all_done := false;
                EXIT;
            END IF;

            -- Pull the tool reply content. Bridge stores
            -- result.content (string) on success, error column on
            -- failure. We hand the model whichever surfaced.
            DECLARE
                content_text text;
            BEGIN
                IF child_row.status = 'done' THEN
                    content_text := child_row.result ->> 'content';
                    IF content_text IS NULL THEN
                        content_text := child_row.result::text;
                    END IF;
                ELSE
                    content_text := jsonb_build_object(
                        'error', child_row.error
                    )::text;
                END IF;

                final_msgs := final_msgs || jsonb_build_array(
                    jsonb_build_object(
                        'tc_id',   pending_elem ->> 'tc_id',
                        'name',    pending_elem ->> 'name',
                        'content', content_text
                    )
                );
            END;
        END LOOP;

        IF NOT all_done THEN
            CONTINUE;
        END IF;

        -- All children resolved; promote to done. Run the equivalent
        -- of the Rust ToolsDispatched write phase: insert tool
        -- messages, enqueue the continuation chat.
        parent_chat_id  := (parent_row.payload ->> 'parent_work_id')::bigint;
        parent_session  := parent_row.payload ->> 'session_id';
        parent_family   := parent_row.payload ->> 'agent_family';
        parent_model    := parent_row.payload ->> 'model';
        parent_provider := parent_row.provider;

        FOR pending_elem IN SELECT * FROM jsonb_array_elements(final_msgs)
        LOOP
            INSERT INTO stewards.messages
                (session_id, role, content, tool_call_id, parent_work_id)
            VALUES (
                parent_session,
                'tool',
                pending_elem ->> 'content',
                pending_elem ->> 'tc_id',
                parent_row.id
            );
        END LOOP;

        -- Enqueue continuation chat. chat_post_internal returns the
        -- new chat work_queue id.
        SELECT stewards.chat_post_internal(
            parent_family, parent_model, parent_session, parent_provider
        ) INTO chat_work_id;

        UPDATE stewards.work_queue
           SET status = 'done',
               result = parent_row.result || jsonb_build_object(
                   'completed_at',     now()::text,
                   'next_chat_work_id', chat_work_id,
                   'final_tool_count',  jsonb_array_length(final_msgs)
               ),
               done_at = now()
         WHERE id = parent_row.id;

        completed_n := completed_n + 1;
    END LOOP;

    RETURN completed_n;
END
$$;

COMMENT ON FUNCTION stewards.tool_dispatch_complete_waiting IS
  'Completion pass for async-fan-out tool_dispatch. Bgworker calls '
  'this on each tick; rows whose mcp_proxy children have all '
  'resolved are promoted from waiting_for_tools to done with the '
  'usual Phase 3 side effects (insert tool messages + enqueue '
  'continuation chat).';

-- ---------------------------------------------------------------------
-- 4. Example tool_defs pointing at mcp_proxy
--
-- One per MCP server that has a clean primary tool. Operators can
-- always insert more rows pointing at additional cached tools.
-- Phase 3e.2.d will auto-promote mcp_tool_cache rows into tool_defs;
-- for now, a small hand-curated set is enough for verification.
--
-- args_schema is intentionally permissive — the underlying server
-- enforces its own input schema, and forwarding the model's args
-- jsonb verbatim lets the operator avoid duplicating that contract.
-- ---------------------------------------------------------------------
INSERT INTO stewards.tool_defs (name, description, args_schema, execute_target) VALUES
  ('gospel_search',
   'Search scriptures, talks, and manuals via the gospel-engine MCP server. '
   'Pass query plus optional kinds[] and limit. Routed through the bridge daemon.',
   '{"type":"object","properties":{"query":{"type":"string"},"kinds":{"type":"array","items":{"type":"string"}},"limit":{"type":"integer"},"mode":{"type":"string"}},"required":["query"]}'::jsonb,
   '{"kind":"mcp_proxy","server":"gospel-engine-v2","tool":"gospel_search"}'::jsonb),

  ('gospel_get',
   'Fetch a specific scripture, talk, or manual section by reference (e.g. "Mosiah 18:8"). '
   'Routed through the bridge to gospel-engine.',
   '{"type":"object","properties":{"ref":{"type":"string"}},"required":["ref"]}'::jsonb,
   '{"kind":"mcp_proxy","server":"gospel-engine-v2","tool":"gospel_get"}'::jsonb),

  ('webster_define',
   '1828 Webster dictionary lookup. Routed through the bridge to webster-mcp.',
   '{"type":"object","properties":{"word":{"type":"string"}},"required":["word"]}'::jsonb,
   '{"kind":"mcp_proxy","server":"webster","tool":"webster_define"}'::jsonb)
ON CONFLICT (name) DO UPDATE
   SET description    = EXCLUDED.description,
       args_schema    = EXCLUDED.args_schema,
       execute_target = EXCLUDED.execute_target;

-- agent_tool_perms intentionally NOT granted here. Operators must
-- explicitly allow these on a per-agent basis. This preserves the
-- soak's current behavior (it stays on its existing tool surface)
-- and keeps the new mcp_proxy surface deny-by-default for safety.
