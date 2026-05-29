-- Smoke 2: the cap gate refuses a Gemini dispatch once spend >= cap.
-- Temporarily set the cap below current spend, prove the gate fires +
-- opencode is unaffected, then restore the $18 cap (since untouched so
-- the real smoke spend keeps counting).

-- a) drop cap below current gemini spend (2990 micro)
UPDATE stewards.provider_spend_caps SET cap_micro = 1 WHERE provider = 'google_gemini';
SELECT stewards.provider_cap_exceeded('google_gemini') AS gemini_exceeded,
       stewards.provider_cap_exceeded('opencode_go')   AS opencode_exceeded;

-- b) a gemini-routed dispatch must RAISE (refused before enqueue)
DO $$
DECLARE v_id uuid; v_wq bigint;
BEGIN
    v_id := stewards.work_item_create(
        p_pipeline_family => 'brainstorm-scamper',
        p_input           => jsonb_build_object('binding_question','gate test'),
        p_slug            => 'smoke-j11-gate-' || to_char(now(),'HH24MISS'),
        p_actor           => 'smoke'
    );
    UPDATE stewards.work_items
       SET model_override='gemini-2.5-flash-lite', provider_override='google_gemini'
     WHERE id = v_id;
    BEGIN
        v_wq := stewards.work_item_dispatch_stage(v_id, NULL);
        RAISE NOTICE 'UNEXPECTED: dispatch succeeded wq=% (gate did NOT fire)', v_wq;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'EXPECTED gate refusal: %', SQLERRM;
    END;
    -- clean up the test work_item
    UPDATE stewards.work_items SET status='cancelled' WHERE id = v_id;
END $$;

-- c) opencode dispatch is NOT gated (sanity) — create + dispatch + cancel
DO $$
DECLARE v_id uuid; v_wq bigint;
BEGIN
    v_id := stewards.work_item_create(
        p_pipeline_family => 'brainstorm-scamper',
        p_input           => jsonb_build_object('binding_question','opencode not gated'),
        p_slug            => 'smoke-j11-oc-' || to_char(now(),'HH24MISS'),
        p_actor           => 'smoke'
    );
    UPDATE stewards.work_items SET model_override='deepseek-v4-flash' WHERE id = v_id;
    v_wq := stewards.work_item_dispatch_stage(v_id, NULL);
    RAISE NOTICE 'opencode dispatch OK (not gated) wq=%', v_wq;
    -- delete the queued chat (work_queue has no 'cancelled' status) so it
    -- doesn't actually run, then cancel the work_item.
    DELETE FROM stewards.work_queue WHERE id = v_wq AND status = 'pending';
    UPDATE stewards.work_items SET status='cancelled' WHERE id = v_id;
END $$;

-- d) restore the $18 cap (UPDATE, not refill — keeps `since` so the real
--    smoke spend stays counted against the balance)
UPDATE stewards.provider_spend_caps SET cap_micro = 18000000 WHERE provider = 'google_gemini';
SELECT provider, cap_micro, enforced,
       stewards.provider_spend_since(provider) AS spent_micro,
       stewards.provider_cap_exceeded(provider) AS exceeded
  FROM stewards.provider_spend_caps;
