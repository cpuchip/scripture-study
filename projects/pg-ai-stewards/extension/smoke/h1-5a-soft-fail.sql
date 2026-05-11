-- Smoke H.1.5a: disabled server returns NULL (not EXCEPTION)
DO $$
DECLARE v_id bigint;
BEGIN
    SELECT stewards.mcp_proxy_enqueue('webster', 'define', '{"word":"test"}'::jsonb, NULL) INTO v_id;
    RAISE NOTICE 'disabled-server enqueue returned: %', COALESCE(v_id::text, 'NULL');
    IF v_id IS NOT NULL THEN
        RAISE EXCEPTION 'expected NULL, got %', v_id;
    END IF;

    -- Sanity: enabled server still enqueues normally
    SELECT stewards.mcp_proxy_enqueue('fetch-md', 'fetch_url', '{"url":"https://example.com"}'::jsonb, NULL) INTO v_id;
    RAISE NOTICE 'enabled-server enqueue returned: %', v_id;
    IF v_id IS NULL THEN
        RAISE EXCEPTION 'expected bigint, got NULL';
    END IF;

    -- Cleanup: kill the synth fetch-md row before bridge picks it up
    UPDATE stewards.work_queue SET status='error', error='h1-5a smoke cleanup' WHERE id = v_id AND status='pending';
    RAISE NOTICE 'H.1.5a smoke PASSED';
END
$$;
