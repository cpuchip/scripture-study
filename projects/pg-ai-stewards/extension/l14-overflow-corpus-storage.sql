-- =====================================================================
-- Batch L.1.1.4 — Overflow corpus storage (parents + embedded leaves)
-- =====================================================================
-- When the bridge intercepts an oversized tool result, the content is
-- recursively split into:
--   - parent chunks (~4K tokens / 14K chars, no embedding) for context
--     scaffolding when the agent retrieves
--   - leaf chunks  (~512 tokens / 1800 chars, embedded into pgvector)
--     for fine-grained vector search
--
-- Auto-merging retrieval (L.1.1.7) hits leaves and promotes to parents
-- when 3+ leaves under the same parent score (LlamaIndex pattern).
--
-- Pure schema sub-phase. Tables empty until L.1.1.6 (chunk_and_index)
-- populates them.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. messages_raw_overflow (parents — storage only).
-- ---------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS stewards.messages_raw_overflow (
    id              bigserial PRIMARY KEY,
    message_id      bigint NOT NULL REFERENCES stewards.messages(id) ON DELETE CASCADE,
    parent_ordinal  int    NOT NULL,
    content         text   NOT NULL,
    byte_size       int    NOT NULL,
    tool_name       text,                              -- denormalized for filter
    binding_question text,                             -- the binding-at-time-of-index
    created_at      timestamptz NOT NULL DEFAULT now(),
    UNIQUE (message_id, parent_ordinal)
);

CREATE INDEX IF NOT EXISTS messages_raw_overflow_message_id
    ON stewards.messages_raw_overflow (message_id);

COMMENT ON TABLE stewards.messages_raw_overflow IS
'Batch L.1.1.4: parent chunks (~4K tokens each) of an oversized tool message. Storage-only — no embeddings. Used by auto-merging retrieval (L.1.1.7) when multiple leaves under the same parent score in vector search. tool_name and binding_question denormalized for cheap filtering during retrieval.';


-- ---------------------------------------------------------------------
-- 2. messages_raw_overflow_leaves (embedded chunks).
-- ---------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS stewards.messages_raw_overflow_leaves (
    id                  bigserial PRIMARY KEY,
    message_id          bigint NOT NULL REFERENCES stewards.messages(id) ON DELETE CASCADE,
    parent_id           bigint NOT NULL REFERENCES stewards.messages_raw_overflow(id) ON DELETE CASCADE,
    leaf_ordinal        int    NOT NULL,        -- ordinal within the parent
    content             text   NOT NULL,         -- the 512-token leaf body
    context_prefix      text,                    -- L.1.1.5 contextual blurb (50-100 tok) prepended for embedding
    byte_size           int    NOT NULL,
    embedding           vector(768),
    embedded_at         timestamptz,
    embedded_model      text,
    embedding_error     text,
    created_at          timestamptz NOT NULL DEFAULT now(),
    UNIQUE (parent_id, leaf_ordinal)
);

CREATE INDEX IF NOT EXISTS messages_raw_overflow_leaves_vec
    ON stewards.messages_raw_overflow_leaves
    USING hnsw (embedding vector_cosine_ops);

CREATE INDEX IF NOT EXISTS messages_raw_overflow_leaves_message_id
    ON stewards.messages_raw_overflow_leaves (message_id);

CREATE INDEX IF NOT EXISTS messages_raw_overflow_leaves_parent_id
    ON stewards.messages_raw_overflow_leaves (parent_id);

CREATE INDEX IF NOT EXISTS messages_raw_overflow_leaves_pending_embed
    ON stewards.messages_raw_overflow_leaves (id)
    WHERE embedded_at IS NULL AND embedding_error IS NULL;

COMMENT ON TABLE stewards.messages_raw_overflow_leaves IS
'Batch L.1.1.4: leaf chunks (~512 tokens each) of an oversized tool message. Embedded into pgvector for cross-message retrieval. context_prefix holds the Anthropic-style 50-100 token contextual blurb added by L.1.1.5 before embedding. Embedded async via bgworker embed kind.';


-- ---------------------------------------------------------------------
-- 3. Helper: list_overflow_parents — for the judge surface.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.list_overflow_parents(p_message_id bigint)
RETURNS TABLE (
    parent_id bigint,
    parent_ordinal int,
    byte_size int,
    leaf_count bigint,
    embedded_count bigint
) LANGUAGE sql STABLE AS $$
    SELECT
        p.id AS parent_id,
        p.parent_ordinal,
        p.byte_size,
        count(l.id) AS leaf_count,
        count(l.embedded_at) AS embedded_count
      FROM stewards.messages_raw_overflow p
      LEFT JOIN stewards.messages_raw_overflow_leaves l
        ON l.parent_id = p.id
     WHERE p.message_id = p_message_id
     GROUP BY p.id, p.parent_ordinal, p.byte_size
     ORDER BY p.parent_ordinal
$$;

COMMENT ON FUNCTION stewards.list_overflow_parents(bigint) IS
'Batch L.1.1.4: enumerate the parent chunks indexed for a given message, with leaf counts and embedding-progress counts. Used by the judge surface (L.1.1.8) to present the corpus overview to the consuming agent.';


-- =====================================================================
-- End of l14-overflow-corpus-storage.sql
-- =====================================================================
