-- =====================================================================
-- ES.3.s4 — Drop the leaf index (destructive — RATIFIED, decision 3)
-- =====================================================================
-- The judge-compiled-brief (ES.3.s2) replaced the leaf-chunk-and-embed
-- compaction. The leaf machinery is now dead code — verified before
-- this drop:
--   - chunk_and_index            — 0 callers (es7 intercept replaced it)
--   - render_judge_surface       — 0 callers (es7 intercept replaced it)
--   - retrieve_with_merge        — 0 callers, not a registered tool
--   - retrieve_with_merge_like_leaf — 0 callers
--   - contextualize_leaf         — only called by chunk_and_index
--   - apply_contextualize_leaf   — only fired by the trigger below
--   - list_overflow_parents      — 0 callers, not a registered tool
--   - messages_raw_overflow_leaves — 972 dead rows, no FK references it
--
-- KEPT: messages_raw_overflow (raw parent recovery — read_overflow_raw,
-- read_corpus_parents both touch only parents), L.3 engram_embeddings
-- (opt-in corpus search), map_reduce_extract_engrams (unattended cases).
--
-- The user's vote on the ES.3 council (decision 3, 2026-05-15) is the
-- explicit ratification this destructive SQL requires.
-- =====================================================================

-- 1. Completion trigger for the leaf contextualizer.
DROP TRIGGER IF EXISTS work_queue_apply_contextualize_leaf ON stewards.work_queue;

-- 2. Leaf-machinery functions (dependency order: callers before callees,
--    though all are dead so order is cosmetic).
DROP FUNCTION IF EXISTS stewards.trigger_apply_contextualize_leaf();
DROP FUNCTION IF EXISTS stewards.apply_contextualize_leaf(bigint);
DROP FUNCTION IF EXISTS stewards.contextualize_leaf(bigint);
DROP FUNCTION IF EXISTS stewards.chunk_and_index(bigint, text, integer, integer, integer, integer);
DROP FUNCTION IF EXISTS stewards.retrieve_with_merge(vector, bigint, integer, integer);
DROP FUNCTION IF EXISTS stewards.retrieve_with_merge_like_leaf(bigint, bigint, integer, integer);
DROP FUNCTION IF EXISTS stewards.render_judge_surface(bigint, text);
DROP FUNCTION IF EXISTS stewards.list_overflow_parents(bigint);

-- 3. The leaf table itself. No FK references it (verified); DROP is clean.
DROP TABLE IF EXISTS stewards.messages_raw_overflow_leaves;

-- =====================================================================
-- End of es9-drop-leaf-index.sql
-- =====================================================================
