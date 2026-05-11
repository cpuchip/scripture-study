-- =====================================================================
-- Phase 5c — Tiny fix: allow 'gate' as a valid sessions.kind
--
-- 5a's evaluate_gate (and 5b's generate_scenarios + verify_work_item)
-- create sessions with kind='gate' for clean audit separation between
-- pipeline-stage chats and gate-style evaluation chats. The existing
-- sessions_kind_check (chat|agent|tool|study|dev) doesn't include
-- 'gate' so the INSERT raised. Add 'gate' to the allowed set.
--
-- Caught during smoke testing of 5a/5b on dev.
-- =====================================================================

ALTER TABLE stewards.sessions DROP CONSTRAINT IF EXISTS sessions_kind_check;

ALTER TABLE stewards.sessions
    ADD CONSTRAINT sessions_kind_check
    CHECK (kind = ANY (ARRAY['chat','agent','tool','study','dev','gate']));
