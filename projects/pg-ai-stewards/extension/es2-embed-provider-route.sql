-- =====================================================================
-- ES.1.s5 — Embed provider routing fix (CF-7)
-- =====================================================================
-- Embeddings run on LM Studio (provider 'lm_studio', local, port 1234,
-- model text-embedding-nomic-embed-text-v1.5). But the embed-enqueue
-- SQL in l3 and l26 hardcoded provider='opencode_go' — OpenCode Go has
-- no embeddings endpoint, so those jobs 404.
--
-- Rather than chase every enqueue site (present and future), enforce
-- the invariant in one place: a BEFORE INSERT trigger on work_queue
-- rewrites kind='embed' rows to provider='lm_studio'. Embeds ALWAYS
-- route to LM Studio; future code cannot misroute.
-- =====================================================================

CREATE OR REPLACE FUNCTION stewards.trigger_embed_provider_route()
RETURNS trigger LANGUAGE plpgsql AS $FN$
BEGIN
    IF NEW.kind = 'embed' AND COALESCE(NEW.provider, '') <> 'lm_studio' THEN
        RAISE NOTICE 'embed provider route: rewrote % -> lm_studio (wq pending insert)',
            COALESCE(NEW.provider, '(null)');
        NEW.provider := 'lm_studio';
    END IF;
    RETURN NEW;
END;
$FN$;

DROP TRIGGER IF EXISTS work_queue_embed_provider_route ON stewards.work_queue;

CREATE TRIGGER work_queue_embed_provider_route
BEFORE INSERT ON stewards.work_queue
FOR EACH ROW
WHEN (NEW.kind = 'embed')
EXECUTE FUNCTION stewards.trigger_embed_provider_route();

COMMENT ON FUNCTION stewards.trigger_embed_provider_route() IS
'ES.1.s5 (CF-7): BEFORE INSERT trigger on work_queue. Forces every kind=embed row to provider=lm_studio. Embeddings run on local LM Studio (nomic-embed-text-v1.5); OpenCode Go has no embeddings endpoint. Enforces the routing invariant in one place so no enqueue site can misroute.';

-- Discard the 730 already-misrouted pending embed rows — they would
-- 404. Re-embedding can be re-triggered cleanly once the pipeline is
-- sound (and ES.3 may remove leaf embedding entirely).
WITH discarded AS (
    UPDATE stewards.work_queue
       SET status = 'error'
     WHERE kind = 'embed'
       AND status = 'pending'
       AND COALESCE(provider, '') <> 'lm_studio'
    RETURNING 1
)
SELECT count(*) AS discarded_misrouted FROM discarded;

-- =====================================================================
-- End of es2-embed-provider-route.sql
-- =====================================================================
