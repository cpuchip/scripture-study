-- =====================================================================
-- Batch H.1.5a — mcp_proxy_enqueue: soft-fail on disabled server
--
-- Surfaced 2026-05-11 during H.1.4's first real e2e run:
-- When kimi-k2.6 called web_search_exa against the disabled exa-search
-- MCP server, mcp_proxy_enqueue RAISE EXCEPTION'd. The pgrx SPI longjmp
-- exited the bgworker dispatcher process (exit code 1) before it could
-- catch the error into a structured spi::Error and write a synthetic
-- tool reply. The in_progress tool_dispatch row stayed stuck for 5+ min
-- because the reaper only sweeps at worker startup, not periodically.
--
-- Fix at the SQL layer: RAISE NOTICE + RETURN NULL instead of EXCEPTION.
-- The Rust code path at tools.rs:355 already has an Ok(None) branch
-- that converts to "mcp_proxy_enqueue(...) returned NULL" error string,
-- which becomes a synthetic tool reply for the model to see and route
-- around. No Rust rebuild needed; sidesteps the longjmp issue entirely
-- for the disabled-server case.
--
-- Deeper followups still open:
--   - Harden BGW SPI longjmp catch generally (so ANY future RAISE
--     EXCEPTION in a substrate fn doesn't crash workers)
--   - Periodic reaper tick (60s) — not just startup-only
-- Both intentionally deferred; this fix is the minimum needed to
-- finish H.1's first real e2e.
-- =====================================================================

CREATE OR REPLACE FUNCTION stewards.mcp_proxy_enqueue(
    p_server text,
    p_tool text,
    p_args jsonb,
    p_parent_tool_dispatch_id bigint
)
RETURNS bigint
LANGUAGE plpgsql AS $func$
DECLARE
    new_id bigint;
BEGIN
    -- H.1.5a (2026-05-11): soft-fail when the target server is not
    -- registered or not enabled. Previously raised EXCEPTION; that
    -- crashed the bgworker dispatcher via pgrx SPI longjmp. NOTICE +
    -- NULL lets the Rust caller (tools.rs::exec_mcp_proxy_tool) emit a
    -- structured tool reply ("mcp_proxy_enqueue returned NULL") that
    -- the model sees as a tool error — same shape it'd see for any
    -- tool failure.
    IF NOT EXISTS (
        SELECT 1 FROM stewards.mcp_servers
        WHERE name = p_server AND enabled
    ) THEN
        RAISE NOTICE 'mcp_proxy_enqueue: server % is not registered or not enabled — returning NULL', p_server;
        RETURN NULL;
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

    PERFORM pg_notify('stewards_mcp_proxy', new_id::text);

    RETURN new_id;
END;
$func$;

COMMENT ON FUNCTION stewards.mcp_proxy_enqueue(text, text, jsonb, bigint) IS
'H.1.5a (Batch H): RAISE NOTICE + RETURN NULL on disabled/unregistered server (was RAISE EXCEPTION). Prevents bgworker crash via pgrx SPI longjmp; lets caller emit a structured tool-failure reply that the model can read and route around.';
