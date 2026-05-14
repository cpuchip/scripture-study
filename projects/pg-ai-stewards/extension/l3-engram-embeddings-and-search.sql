-- =====================================================================
-- Batch L.3 — engram_embeddings table + search_engrams SQL fn + tool
-- =====================================================================
-- Cross-message engram search via pgvector. Foundation for Batch M's
-- cross-session memory tool.
--
-- Schema: stewards.engram_embeddings table keyed by "<msg_id>:<engram_id>"
-- (single-column text id matches bgworker's embed handler's UPDATE WHERE
-- id = $1 pattern). Vector dimension 768 (matches stewards.studies for
-- shared semantic space).
--
-- Population: AFTER UPDATE trigger on stewards.messages.engrams parses
-- the engram items, INSERTs/UPSERTs rows into engram_embeddings, then
-- enqueues embed work_queue jobs.
--
-- Search: search_engrams(query_embedding, session_id?, project?, limit)
-- SQL fn. The Go MCP tool wrapper (search_engrams.go) does the
-- query-side embedding call + invokes the SQL.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. engram_embeddings table.
-- ---------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS stewards.engram_embeddings (
    id                  text PRIMARY KEY,                 -- "<message_id>:<engram_id>"
    message_id          bigint NOT NULL,
    engram_id           text NOT NULL,
    tier                text NOT NULL CHECK (tier IN ('hot','medium','cold')),
    topic               text NOT NULL DEFAULT '',
    content_preview     text NOT NULL DEFAULT '',         -- first ~200 chars for cheap snippet
    embedding           vector(768),
    embedded_at         timestamptz,
    embedded_model      text,
    embedding_error     text,
    session_id          text,                              -- denormalized for cheap filter
    project_association text,                              -- denormalized for cheap filter
    created_at          timestamptz NOT NULL DEFAULT now(),
    FOREIGN KEY (message_id) REFERENCES stewards.messages(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS engram_embeddings_vec
    ON stewards.engram_embeddings
    USING hnsw (embedding vector_cosine_ops);

CREATE INDEX IF NOT EXISTS engram_embeddings_message_id
    ON stewards.engram_embeddings (message_id);

CREATE INDEX IF NOT EXISTS engram_embeddings_session
    ON stewards.engram_embeddings (session_id);

CREATE INDEX IF NOT EXISTS engram_embeddings_project
    ON stewards.engram_embeddings (project_association);

COMMENT ON TABLE stewards.engram_embeddings IS
'Batch L.3: per-engram embeddings for cross-message semantic search. Populated via AFTER UPDATE trigger on stewards.messages.engrams. id = "<message_id>:<engram_id>" matches bgworker embed handler''s UPDATE WHERE id = $1 pattern. session_id and project_association denormalized from work_items for cheap filtering.';


-- ---------------------------------------------------------------------
-- 2. Trigger: AFTER UPDATE OF engrams populates engram_embeddings.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.trigger_populate_engram_embeddings()
RETURNS trigger LANGUAGE plpgsql AS $FN$
DECLARE
    v_item           jsonb;
    v_engram_id      text;
    v_tier           text;
    v_topic          text;
    v_content        text;
    v_preview        text;
    v_session        text;
    v_project        text;
    v_work_item      stewards.work_items%ROWTYPE;
    v_composite_id   text;
    v_wq_id          bigint;
BEGIN
    -- Only fire when engrams.items changed (or first populated). Skip
    -- if items array is empty.
    IF NEW.engrams IS NULL
       OR jsonb_typeof(NEW.engrams -> 'items') <> 'array'
       OR jsonb_array_length(NEW.engrams -> 'items') = 0 THEN
        RETURN NEW;
    END IF;

    v_session := NEW.session_id;

    -- Denormalize project_association from the owning work_item.
    SELECT * INTO v_work_item FROM stewards.work_items
     WHERE v_session = ANY(session_ids) LIMIT 1;
    v_project := v_work_item.project_association;

    FOR v_item IN
        SELECT i FROM jsonb_array_elements(NEW.engrams -> 'items') i
    LOOP
        v_engram_id    := v_item ->> 'id';
        v_tier         := lower(COALESCE(v_item ->> 'tier', 'cold'));
        v_topic        := COALESCE(v_item ->> 'topic', '');
        v_content      := COALESCE(v_item ->> 'content', '');
        v_preview      := substring(v_content FROM 1 FOR 200);
        v_composite_id := NEW.id::text || ':' || v_engram_id;

        -- Upsert the row (will be embedded async via bgworker embed kind).
        INSERT INTO stewards.engram_embeddings
            (id, message_id, engram_id, tier, topic, content_preview, session_id, project_association)
        VALUES
            (v_composite_id, NEW.id, v_engram_id, v_tier, v_topic, v_preview, v_session, v_project)
        ON CONFLICT (id) DO UPDATE
           SET tier = EXCLUDED.tier,
               topic = EXCLUDED.topic,
               content_preview = EXCLUDED.content_preview,
               session_id = EXCLUDED.session_id,
               project_association = EXCLUDED.project_association,
               -- Reset embedded_at to force re-embedding if content changed
               embedded_at = CASE WHEN stewards.engram_embeddings.content_preview <> EXCLUDED.content_preview
                                  THEN NULL ELSE stewards.engram_embeddings.embedded_at END;

        -- Enqueue embed job (only if not already embedded).
        IF NOT EXISTS (
            SELECT 1 FROM stewards.engram_embeddings
             WHERE id = v_composite_id AND embedded_at IS NOT NULL
        ) THEN
            INSERT INTO stewards.work_queue (kind, provider, payload, status)
            VALUES (
                'embed',
                'opencode_go',
                jsonb_build_object(
                    'target_table', 'engram_embeddings',
                    'target_id', v_composite_id,
                    'text', COALESCE(v_topic || E'\n\n' || v_content, v_content, '')
                ),
                'pending'
            )
            RETURNING id INTO v_wq_id;
        END IF;
    END LOOP;

    RETURN NEW;
END;
$FN$;

DROP TRIGGER IF EXISTS messages_populate_engram_embeddings ON stewards.messages;

CREATE TRIGGER messages_populate_engram_embeddings
AFTER UPDATE OF engrams ON stewards.messages
FOR EACH ROW
WHEN (NEW.engrams IS DISTINCT FROM OLD.engrams)
EXECUTE FUNCTION stewards.trigger_populate_engram_embeddings();

COMMENT ON FUNCTION stewards.trigger_populate_engram_embeddings() IS
'Batch L.3: AFTER UPDATE OF engrams trigger. Upserts one engram_embeddings row per engram item, denormalizing tier/topic/session_id/project_association. Enqueues embed work_queue jobs for rows lacking embeddings.';


-- ---------------------------------------------------------------------
-- 3. search_engrams SQL fn — cosine similarity over the embedding index.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.search_engrams_by_vector(
    p_query_embedding vector,
    p_session_id text DEFAULT NULL,
    p_project_association text DEFAULT NULL,
    p_limit int DEFAULT 10
) RETURNS TABLE (
    id text,
    message_id bigint,
    engram_id text,
    tier text,
    topic text,
    content_preview text,
    session_id text,
    project_association text,
    similarity float
) LANGUAGE sql STABLE AS $$
    SELECT
        e.id,
        e.message_id,
        e.engram_id,
        e.tier,
        e.topic,
        e.content_preview,
        e.session_id,
        e.project_association,
        1 - (e.embedding <=> p_query_embedding) AS similarity
      FROM stewards.engram_embeddings e
     WHERE e.embedding IS NOT NULL
       AND (p_session_id IS NULL OR e.session_id = p_session_id)
       AND (p_project_association IS NULL OR e.project_association = p_project_association)
     ORDER BY e.embedding <=> p_query_embedding
     LIMIT GREATEST(p_limit, 1)
$$;

COMMENT ON FUNCTION stewards.search_engrams_by_vector(vector, text, text, int) IS
'Batch L.3: cosine-similarity search over engram_embeddings. Substrate-wide by default; optional session_id / project_association filters. Returns top-K results ordered by similarity. The Go MCP tool wrapper (search_engrams) handles the query-side embedding call before invoking this.';


-- ---------------------------------------------------------------------
-- 4. Backfill: populate engram_embeddings from existing messages.engrams.
-- ---------------------------------------------------------------------
-- One-shot fill so existing engrams become searchable immediately.

DO $$
DECLARE
    v_msg stewards.messages%ROWTYPE;
    v_item jsonb;
    v_engram_id text;
    v_composite_id text;
    v_work_item stewards.work_items%ROWTYPE;
    v_project text;
BEGIN
    FOR v_msg IN
        SELECT * FROM stewards.messages
         WHERE engrams IS NOT NULL
           AND jsonb_typeof(engrams -> 'items') = 'array'
           AND jsonb_array_length(engrams -> 'items') > 0
    LOOP
        SELECT * INTO v_work_item FROM stewards.work_items
         WHERE v_msg.session_id = ANY(session_ids) LIMIT 1;
        v_project := v_work_item.project_association;

        FOR v_item IN
            SELECT i FROM jsonb_array_elements(v_msg.engrams -> 'items') i
        LOOP
            v_engram_id := v_item ->> 'id';
            v_composite_id := v_msg.id::text || ':' || v_engram_id;

            INSERT INTO stewards.engram_embeddings
                (id, message_id, engram_id, tier, topic, content_preview, session_id, project_association)
            VALUES (
                v_composite_id, v_msg.id, v_engram_id,
                lower(COALESCE(v_item ->> 'tier', 'cold')),
                COALESCE(v_item ->> 'topic', ''),
                substring(COALESCE(v_item ->> 'content', '') FROM 1 FOR 200),
                v_msg.session_id, v_project
            )
            ON CONFLICT (id) DO NOTHING;

            -- Enqueue embed job.
            INSERT INTO stewards.work_queue (kind, provider, payload, status)
            VALUES (
                'embed', 'opencode_go',
                jsonb_build_object(
                    'target_table', 'engram_embeddings',
                    'target_id', v_composite_id,
                    'text', COALESCE((v_item ->> 'topic') || E'\n\n' || (v_item ->> 'content'),
                                     v_item ->> 'content', '')
                ),
                'pending'
            );
        END LOOP;
    END LOOP;

    RAISE NOTICE 'L.3 backfill: engram_embeddings populated from existing messages.engrams; embed jobs queued';
END$$;


-- =====================================================================
-- End of l3-engram-embeddings-and-search.sql
-- =====================================================================
