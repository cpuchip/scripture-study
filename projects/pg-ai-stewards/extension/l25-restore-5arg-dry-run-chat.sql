-- =====================================================================
-- L.1.1 follow-up — RESTORE the 5-arg dry_run_chat overload
-- =====================================================================
-- L24 dropped this thinking it was a duplicate. It wasn't — the live
-- chat_post_internal (which the bgworker's tool_dispatch_complete_
-- waiting → continuation chat path uses) was updated at some earlier
-- point to call dry_run_chat with 5 args (agent, model, session, NULL,
-- provider). The 5-arg form wasn't in any tracked migration file, so
-- dropping it broke every continuation chat across the substrate.
--
-- Restored as a thin wrapper that delegates to the 4-arg form. The
-- provider argument is currently a no-op for body composition because
-- compose_messages looks up provider_for_session(session_id) internally
-- (L.1's pattern). The wrapper exists for signature compatibility.
-- =====================================================================

CREATE OR REPLACE FUNCTION stewards.dry_run_chat(
    p_agent_family text,
    p_model        text,
    p_session_id   text,
    p_user_input   text,
    p_provider     text
) RETURNS jsonb
LANGUAGE plpgsql STABLE AS $func$
BEGIN
    -- Provider is informational; compose_messages looks it up via
    -- provider_for_session(session_id) internally per L.1.
    RETURN stewards.dry_run_chat(p_agent_family, p_model, p_session_id, p_user_input);
END;
$func$;

COMMENT ON FUNCTION stewards.dry_run_chat(text, text, text, text, text) IS
'L.1.1 restoration: 5-arg wrapper around the canonical 4-arg dry_run_chat. Exists for signature compatibility with chat_post_internal''s 5-arg call form. Provider arg currently informational — compose_messages does its own provider lookup.';

-- =====================================================================
-- End of l25-restore-5arg-dry-run-chat.sql
-- =====================================================================
