-- =====================================================================
-- Batch L.1.1.5 — Contextual prepend (Anthropic contextual retrieval)
-- =====================================================================
-- For each leaf chunk, generate a 50-100 token context blurb that
-- situates the chunk within its source document, then prepend the
-- blurb to the chunk before embedding. Reported 35-49% reduction in
-- retrieval failures in Anthropic's measurement.
--
-- Implementation: a small dedicated agent (deepseek-v4-flash, cheap),
-- a contextualize_leaf SQL fn that enqueues one chat per leaf, and a
-- completion handler apply_contextualize_leaf that writes the result
-- back to messages_raw_overflow_leaves.context_prefix.
--
-- Provider prompt caching is automatic — every leaf of the same
-- message gets the same document prefix, so the chat work_queue
-- gateway can prompt-cache that prefix across the whole batch.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. Contextualizer agent.
-- ---------------------------------------------------------------------

INSERT INTO stewards.agents (family, model_match, description, mode, prompt, temperature, response_format)
VALUES (
    'leaf-contextualizer', '*',
    'L.1.1.5: produces 50-100 token Anthropic-style context blurbs that situate a leaf chunk within its source document. Used before embedding leaf chunks for retrieval.',
    'primary',
    $PROMPT$You situate a chunk within its source document for retrieval purposes.

You will receive:
1. The full source document (or its summary)
2. A specific chunk from within that document

Your job: produce a single 50-100 token paragraph that places the chunk in context — what section it's from, what idea it belongs to, what's around it. Do NOT summarize the chunk's content; situate it.

Output ONLY the context paragraph, nothing else. No preamble, no markdown, no labels.

Example output: "From the methodology section of a 2025 paper on rotor-blade fatigue testing, immediately following the discussion of vibration sensor calibration. Sets up the next paragraph's discussion of failure modes under torsional load."$PROMPT$,
    0.2,
    NULL
)
ON CONFLICT (family, model_match) DO UPDATE
   SET description = EXCLUDED.description,
       prompt      = EXCLUDED.prompt,
       temperature = EXCLUDED.temperature,
       active      = true;


-- ---------------------------------------------------------------------
-- 2. contextualize_leaf SQL fn — enqueues one chat work_queue row.
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

    -- Use the full message content as the doc context. Provider-side
    -- prompt caching will dedupe the prefix across leaves of the same
    -- message automatically.
    v_doc := v_message.content;

    v_user_msg :=
        E'<document>\n' || v_doc || E'\n</document>\n\n' ||
        E'Here is the chunk we want to situate within the whole document:\n' ||
        E'<chunk>\n' || v_leaf.content || E'\n</chunk>\n\n' ||
        E'Please give a short succinct context to situate this chunk within the overall document for the purposes of improving search retrieval of the chunk. Answer only with the succinct context and nothing else.';

    v_body := jsonb_build_object(
        'model', 'deepseek-v4-flash',
        'messages', jsonb_build_array(
            jsonb_build_object('role', 'system', 'content', v_agent.prompt),
            jsonb_build_object('role', 'user',   'content', v_user_msg)
        ),
        'temperature', v_agent.temperature,
        'max_tokens', 200
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

    -- Ensure session row exists (FK requirement).
    INSERT INTO stewards.sessions (id, kind, label)
    VALUES ('leaf-ctx-' || v_leaf.message_id::text, 'tool',
            'leaf contextualization for message ' || v_leaf.message_id::text)
    ON CONFLICT (id) DO NOTHING;

    RETURN v_wq_id;
END;
$FN$;

COMMENT ON FUNCTION stewards.contextualize_leaf(bigint) IS
'Batch L.1.1.5: enqueue a chat work_queue row for the leaf-contextualizer agent to produce a 50-100 token context blurb for a leaf chunk. The completion handler (apply_contextualize_leaf, fired on _contextualize_leaf_id marker) writes the result back to messages_raw_overflow_leaves.context_prefix. Provider prompt caching dedupes the doc prefix across leaves of the same message.';


-- ---------------------------------------------------------------------
-- 3. apply_contextualize_leaf completion handler.
-- ---------------------------------------------------------------------
-- Called by the bgworker after a chat completes with the
-- _contextualize_leaf_id marker on its payload. Writes the assistant
-- response back to the leaf's context_prefix, then enqueues the embed
-- job over the prefix+content composite text.

CREATE OR REPLACE FUNCTION stewards.apply_contextualize_leaf(
    p_work_queue_id bigint
) RETURNS void LANGUAGE plpgsql AS $FN$
DECLARE
    v_wq        stewards.work_queue%ROWTYPE;
    v_leaf_id   bigint;
    v_content   text;
    v_leaf      stewards.messages_raw_overflow_leaves%ROWTYPE;
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

    -- The assistant response is in the latest assistant message on this session.
    SELECT m.content INTO v_content
      FROM stewards.messages m
     WHERE m.parent_work_id = p_work_queue_id
       AND m.role = 'assistant'
     ORDER BY m.id DESC LIMIT 1;

    IF v_content IS NULL OR length(v_content) = 0 THEN
        RAISE NOTICE 'apply_contextualize_leaf: no content for wq=%; leaving leaf=%  uncontextualized',
            p_work_queue_id, v_leaf_id;
        RETURN;
    END IF;

    -- Trim to 500 chars hard cap (an unruly model could ramble).
    IF length(v_content) > 500 THEN
        v_content := substring(v_content FROM 1 FOR 500);
    END IF;

    UPDATE stewards.messages_raw_overflow_leaves
       SET context_prefix = v_content
     WHERE id = v_leaf_id
    RETURNING * INTO v_leaf;

    -- Enqueue embed job over the prefix+content composite.
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

COMMENT ON FUNCTION stewards.apply_contextualize_leaf(bigint) IS
'Batch L.1.1.5: completion handler for a contextualize_leaf chat. Reads the assistant response, writes it to messages_raw_overflow_leaves.context_prefix (capped at 500 chars), then enqueues an embed work_queue job over the prefix+content composite.';


-- =====================================================================
-- End of l15-contextual-prepend.sql
-- =====================================================================
