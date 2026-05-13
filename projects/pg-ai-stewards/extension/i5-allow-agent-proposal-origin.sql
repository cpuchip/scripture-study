-- =====================================================================
-- i5 — allow 'agent_proposal' in work_items.origin CHECK constraint
--
-- Companion to i4. The agent-proposal pipeline creates work_items with
-- origin='agent_proposal'. The existing work_items_origin_check did not
-- include that value. Discovered during i4 smoke.
--
-- Existing allowed values (preserved): human, scheduled, watchman,
-- steward, council, agent_planning. Adding: agent_proposal.
-- =====================================================================

ALTER TABLE stewards.work_items
    DROP CONSTRAINT IF EXISTS work_items_origin_check;

ALTER TABLE stewards.work_items
    ADD CONSTRAINT work_items_origin_check
    CHECK (origin = ANY (ARRAY[
        'human'::text,
        'scheduled'::text,
        'watchman'::text,
        'steward'::text,
        'council'::text,
        'agent_planning'::text,
        'agent_proposal'::text
    ]));

COMMENT ON CONSTRAINT work_items_origin_check ON stewards.work_items IS
'i5 (Batch I.1, 2026-05-12): added agent_proposal to allowed origins. Used by the agent-proposal pipeline_family from i4.';
