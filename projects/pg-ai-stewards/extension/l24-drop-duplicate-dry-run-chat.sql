-- =====================================================================
-- L.1.1 follow-up — drop duplicate dry_run_chat(5-arg) overload
-- =====================================================================
-- Surfaced during bacteriopolis L.1.1 verification dispatch. The
-- function work_item_dispatch_stage calls:
--   stewards.dry_run_chat(v_agent, v_model, v_session_id, NULL)
-- which Postgres can't disambiguate when BOTH the 4-arg and 5-arg
-- overloads exist (NULL's type is unknown, and both signatures match).
--
-- The 5-arg variant added p_provider as a positional arg; everywhere
-- it was called, the 4-arg form sufficed (provider is resolved inside
-- the fn from work_queue payload). Drop the 5-arg overload to remove
-- ambiguity. Stewardship fix — affected every new dispatch.
-- =====================================================================

DROP FUNCTION IF EXISTS stewards.dry_run_chat(text, text, text, text, text);

-- =====================================================================
-- End of l24-drop-duplicate-dry-run-chat.sql
-- =====================================================================
