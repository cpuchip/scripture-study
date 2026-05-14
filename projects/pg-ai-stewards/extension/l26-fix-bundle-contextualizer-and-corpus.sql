-- =====================================================================
-- L.1.1.12 — Fix bundle: contextualizer breathing room + corpus access
-- =====================================================================
-- Three fixes:
-- 1. Remove max_tokens=200 from contextualize_leaf body. Per Michael:
--    direct API calls need breathing room; cost discipline lives in
--    cost_cap_micro and intercept-budget patterns, not a tight cap.
-- 2. Update apply_contextualize_leaf to fall back to reasoning_content
--    when content is empty (defensive — captures DeepSeek-V4-flash's
--    output even when reasoning eats most of the budget).
-- 3. SQL helpers + tool_def for read_corpus_parents — the agent's
--    actual retrieval path on a [CORPUS-INDEXED] surface. Replaces
--    the 'retrieve_from_corpus' the judge template promised but
--    didn't ship. Pure paginated parent reads (no vector search yet
--    — that ships with L.3 search_engrams Go wrapper).
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. contextualize_leaf — drop max_tokens cap.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.contextualize_leaf(p_leaf_id bigint)
RETURNS bigint LANGUAGE plpgsql AS $FN$
DECLARE
    v_leaf      stewards.messages_raw_overflow_leaves%ROWTYPE;
    v_message   stewards.messages%ROWTYPE;
    v_agent     stewards.agents%ROWTYPE;
    v_body      jsonb;
    v_user_msg  text;
    v_wq_id     bigint;
    v_doc       text;
BEGIN
    SELECT * INTO v_leaf FROM stewards.messages_raw_overflow_leaves WHERE id = p_leaf_id;
    IF v_leaf.id IS NULL THEN
        RAISE EXCEPTION 'contextualize_leaf: leaf % not found', p_leaf_id;
    END IF;

    IF v_leaf.context_prefix IS NOT NULL THEN
        RAISE NOTICE 'contextualize_leaf: leaf % already contextualized; skipping', p_leaf_id;
        RETURN NULL;
    END IF;

    SELECT * INTO v_message FROM stewards.messages WHERE id = v_leaf.message_id;
    IF v_message.id IS NULL THEN
        RAISE EXCEPTION 'contextualize_leaf: message % not found for leaf %', v_leaf.message_id, p_leaf_id;
    END IF;

    SELECT * INTO v_agent FROM stewards.agents
     WHERE family = 'leaf-contextualizer' AND active LIMIT 1;
    IF v_agent.family IS NULL THEN
        RAISE EXCEPTION 'contextualize_leaf: leaf-contextualizer agent missing';
    END IF;

    v_doc := v_message.content;

    v_user_msg :=
        E'<document>\n' || v_doc || E'\n</document>\n\n' ||
        E'Here is the chunk we want to situate within the whole document:\n' ||
        E'<chunk>\n' || v_leaf.content || E'\n</chunk>\n\n' ||
        E'Please give a short succinct context to situate this chunk within the overall document for the purposes of improving search retrieval of the chunk. Answer only with the succinct context and nothing else.';

    -- L.1.1.12: NO max_tokens. Direct API calls need breathing room.
    -- DeepSeek V4 Flash is a reasoning model; capping at 200 produced
    -- empty content because reasoning ate the entire budget.
    v_body := jsonb_build_object(
        'model', 'deepseek-v4-flash',
        'messages', jsonb_build_array(
            jsonb_build_object('role', 'system', 'content', v_agent.prompt),
            jsonb_build_object('role', 'user',   'content', v_user_msg)
        ),
        'temperature', v_agent.temperature
    );

    INSERT INTO stewards.work_queue (kind, provider, payload, status)
    VALUES (
        'chat',
        'opencode_go',
        jsonb_build_object(
            'session_id', 'leaf-ctx-' || v_leaf.message_id::text,
            'agent_family', 'leaf-contextualizer',
            'requested_model', 'deepseek-v4-flash',
            'body', v_body,
            'tools_disabled', true,
            '_contextualize_leaf_id', p_leaf_id
        ),
        'pending'
    )
    RETURNING id INTO v_wq_id;

    INSERT INTO stewards.sessions (id, kind, label)
    VALUES ('leaf-ctx-' || v_leaf.message_id::text, 'tool',
            'leaf contextualization for message ' || v_leaf.message_id::text)
    ON CONFLICT (id) DO NOTHING;

    RETURN v_wq_id;
END;
$FN$;


-- ---------------------------------------------------------------------
-- 2. apply_contextualize_leaf — reasoning_content fallback.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.apply_contextualize_leaf(
    p_work_queue_id bigint
) RETURNS void LANGUAGE plpgsql AS $FN$
DECLARE
    v_wq         stewards.work_queue%ROWTYPE;
    v_leaf_id    bigint;
    v_content    text;
    v_reasoning  text;
    v_leaf       stewards.messages_raw_overflow_leaves%ROWTYPE;
    v_embed_text text;
BEGIN
    SELECT * INTO v_wq FROM stewards.work_queue WHERE id = p_work_queue_id;
    IF v_wq.id IS NULL THEN
        RAISE EXCEPTION 'apply_contextualize_leaf: wq % not found', p_work_queue_id;
    END IF;

    v_leaf_id := (v_wq.payload ->> '_contextualize_leaf_id')::bigint;
    IF v_leaf_id IS NULL THEN
        RAISE EXCEPTION 'apply_contextualize_leaf: missing _contextualize_leaf_id on wq %', p_work_queue_id;
    END IF;

    SELECT m.content, m.reasoning_content INTO v_content, v_reasoning
      FROM stewards.messages m
     WHERE m.parent_work_id = p_work_queue_id
       AND m.role = 'assistant'
     ORDER BY m.id DESC LIMIT 1;

    -- L.1.1.12: fall back to reasoning_content when content is empty.
    -- DeepSeek-class reasoning models put their working in reasoning_
    -- content. With max_tokens removed (L.1.1.12 fix #1) this is rare
    -- but we keep it as defensive insurance.
    IF v_content IS NULL OR length(v_content) = 0 THEN
        IF v_reasoning IS NOT NULL AND length(v_reasoning) > 0 THEN
            v_content := v_reasoning;
            RAISE NOTICE 'apply_contextualize_leaf: leaf=% empty content; using reasoning_content (% chars)',
                v_leaf_id, length(v_reasoning);
        ELSE
            RAISE NOTICE 'apply_contextualize_leaf: no content for wq=%; leaving leaf=% uncontextualized',
                p_work_queue_id, v_leaf_id;
            RETURN;
        END IF;
    END IF;

    IF length(v_content) > 500 THEN
        v_content := substring(v_content FROM 1 FOR 500);
    END IF;

    UPDATE stewards.messages_raw_overflow_leaves
       SET context_prefix = v_content
     WHERE id = v_leaf_id
    RETURNING * INTO v_leaf;

    v_embed_text := v_content || E'\n\n' || v_leaf.content;

    INSERT INTO stewards.work_queue (kind, provider, payload, status)
    VALUES (
        'embed',
        'opencode_go',
        jsonb_build_object(
            'target_table', 'messages_raw_overflow_leaves',
            'target_id', v_leaf_id::text,
            'text', v_embed_text
        ),
        'pending'
    );

    RAISE NOTICE 'apply_contextualize_leaf: leaf=% prefix written (% chars); embed enqueued',
        v_leaf_id, length(v_content);
END;
$FN$;


-- ---------------------------------------------------------------------
-- 3. read_corpus_parents — paginated read of overflow parents.
-- ---------------------------------------------------------------------
-- Provides agent-callable access to indexed corpus content. Keep simple
-- (paginated by parent_ordinal); vector search ships when the L.3 Go
-- wrapper for synchronous query embedding lands.

CREATE OR REPLACE FUNCTION stewards.read_corpus_parents(
    p_message_id          bigint,
    p_parent_ord_start    int  DEFAULT 0,
    p_count               int  DEFAULT 4,
    p_max_chars_per_part  int  DEFAULT 14000
) RETURNS TABLE (
    parent_ordinal int,
    byte_size      int,
    content        text,
    has_more       boolean
) LANGUAGE sql STABLE AS $$
    WITH page AS (
        SELECT p.parent_ordinal, p.byte_size,
               substring(p.content FROM 1 FOR p_max_chars_per_part) AS content,
               row_number() OVER (ORDER BY p.parent_ordinal) AS rn
          FROM stewards.messages_raw_overflow p
         WHERE p.message_id = p_message_id
           AND p.parent_ordinal >= p_parent_ord_start
         ORDER BY p.parent_ordinal
         LIMIT p_count
    ),
    total AS (
        SELECT count(*) AS n
          FROM stewards.messages_raw_overflow
         WHERE message_id = p_message_id
    )
    SELECT page.parent_ordinal,
           page.byte_size,
           page.content,
           (p_parent_ord_start + p_count) < total.n AS has_more
      FROM page CROSS JOIN total
     ORDER BY page.parent_ordinal
$$;

COMMENT ON FUNCTION stewards.read_corpus_parents(bigint, int, int, int) IS
'Batch L.1.1.12: paginated read of overflow parent chunks for an indexed corpus message. Agent-facing read path until L.3 vector search ships. p_parent_ord_start = first parent to return; p_count = how many parents; p_max_chars_per_part = char cap per parent (avoid bloating).';


-- ---------------------------------------------------------------------
-- 4. tool_def for read_corpus_parents.
-- ---------------------------------------------------------------------

INSERT INTO stewards.tool_defs (name, description, args_schema, execute_target, active)
VALUES (
    'read_corpus_parents',
    'Read parent chunks from an indexed corpus on a [CORPUS-INDEXED] tool message. ' ||
    'Use after the L.1.1.8 judge surface presents you with a corpus — paginate through parents ' ||
    'with parent_ord_start + count. Mark anything precious with mark_engram_important once you find it.',
    $JSON$
    {
      "type": "object",
      "required": ["message_id"],
      "additionalProperties": false,
      "properties": {
        "message_id":         {"type": "integer", "description": "The message id from the [CORPUS-INDEXED] surface header."},
        "parent_ord_start":   {"type": "integer", "default": 0, "description": "First parent ordinal to return."},
        "count":              {"type": "integer", "default": 4, "description": "How many parents to return this call."},
        "max_chars_per_part": {"type": "integer", "default": 14000, "description": "Char cap per parent in the response."}
      }
    }
    $JSON$::jsonb,
    jsonb_build_object('kind', 'mcp_proxy', 'server', 'pg-ai-stewards', 'tool', 'read_corpus_parents'),
    true
)
ON CONFLICT (name) DO UPDATE
   SET description = EXCLUDED.description,
       args_schema = EXCLUDED.args_schema,
       execute_target = EXCLUDED.execute_target,
       active = true;


-- ---------------------------------------------------------------------
-- 5. Update the canonical judge template — point at the real tool.
-- ---------------------------------------------------------------------

UPDATE stewards.judge_templates
   SET template_text = $TMPL$You have been delivered an oversized tool result from **{{tool_name}}** — {{source_bytes}} bytes, indexed into a per-message mini-corpus of {{parent_count}} parent chunks and {{leaf_count}} leaf chunks (~512 tokens each).

Your binding question: **{{binding_question}}**

Top-level overview of the source:
> {{top_overview}}

Within your stewardship over the binding question, judge:

1. **Is the fruit good?** Is this content credible, on-topic, and worth preserving? If not, you may discard the corpus.
2. **What is most precious to save?** Use `read_corpus_parents(message_id={{message_id}}, parent_ord_start=N, count=K)` to scan the parents in order. When you find specific engrams worth preserving for citation, call `mark_engram_important(message_id=..., engram_id=...)`.
3. **What should be discarded?** Anything noise / off-topic / suspect. You can simply not pull it.

You have full agency here. Surface only what matters; pass on what doesn't.$TMPL$,
       updated_at = now()
 WHERE scope = 'canonical';


-- =====================================================================
-- End of l26-fix-bundle-contextualizer-and-corpus.sql
-- =====================================================================
