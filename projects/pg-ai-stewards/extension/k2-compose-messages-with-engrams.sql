-- =====================================================================
-- Batch K.2 — compose_messages emits engrams (head/torso/tail compaction)
-- =====================================================================
-- Rewrites stewards.compose_messages to apply head/torso/tail logic:
--   - HEAD: the first user message + system prompt (zone 1) — never compacted
--   - TAIL: the last 8 messages — always raw (preserve agent rhythm,
--           matches LangChain Deep Agents' "raw recent turns" rule)
--   - TORSO: messages older than the tail — eligible for compaction:
--       * tool message with engrams populated → emit HOT-tier engrams
--         rendered as markdown via render_engrams_markdown helper
--       * assistant message → drop reasoning_content / reasoning_details
--         (rarely re-read; per Anthropic context engineering)
--       * tool message WITHOUT engrams → emit raw (graceful fallback)
--   - Error-trace messages (regex match) → emit raw regardless of
--     position (LangChain rule: preserve errors so the agent doesn't
--     repeat the failure)
--   - User messages → never compacted (binding context)
--
-- Ratified in projects/pg-ai-stewards/.spec/proposals/
--               substrate-batch-k-engram-context.md (2026-05-13).
--
-- Decisions wired in here:
--   - Tail size: 8 messages (~3 assistant turns + their tool calls)
--   - Emit HOT engrams in active context; MEDIUM/COLD retrievable via
--     expand_message (K.3)
--   - Backwards-compatible: messages without engrams populated render
--     as raw, identical to pre-K.2 behavior
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. Helper: render engrams as markdown for compose_messages output.
-- ---------------------------------------------------------------------
-- Renders only HOT-tier engrams. MEDIUM/COLD live on the engrams
-- jsonb and are retrievable via expand_message (K.3). Sets a header
-- with raw size + engram count, the injection-suspected banner if
-- applicable (K.6), HOT engram bodies with preserved URLs/quotes,
-- and the expand_message footer.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.render_engrams_markdown(
    p_message_id bigint,
    p_engrams jsonb
) RETURNS text LANGUAGE plpgsql STABLE AS $FN$
DECLARE
    v_md          text := '';
    v_n_total     int;
    v_n_hot       int;
    v_raw_chars   int;
    v_injection   boolean;
    v_evidence    text;
    v_item        jsonb;
    v_urls_str    text;
    v_quotes_str  text;
    v_dates_str   text;
    v_names_str   text;
BEGIN
    v_n_total   := jsonb_array_length(COALESCE(p_engrams -> 'items', '[]'::jsonb));
    v_raw_chars := COALESCE((p_engrams ->> 'raw_chars')::int, 0);
    v_injection := COALESCE((p_engrams ->> 'injection_suspected')::boolean, false);
    v_evidence  := p_engrams ->> 'injection_evidence';

    SELECT COUNT(*) INTO v_n_hot
      FROM jsonb_array_elements(COALESCE(p_engrams -> 'items', '[]'::jsonb)) i
     WHERE i ->> 'tier' = 'hot';

    -- Header.
    v_md := '[Engrams from msg #' || p_message_id::text
         || ', raw ' || v_raw_chars::text || ' chars, '
         || v_n_total::text || ' total engrams ('
         || v_n_hot::text || ' hot shown below)]'
         || E'\n\n';

    -- Injection banner (K.6 L1 — banner-only injection defense).
    IF v_injection THEN
        v_md := v_md ||
            E'⚠️ Source content showed signs of prompt injection. Engrams have been filtered by the extractor. ' ||
            E'Raw available via expand_message(id=' || p_message_id::text ||
            E', tier=''raw'', confirm_inspect_raw=true) — operator awareness required.';
        IF v_evidence IS NOT NULL AND v_evidence <> '' THEN
            v_md := v_md || E'\nEvidence: ' || v_evidence;
        END IF;
        v_md := v_md || E'\n\n';
    END IF;

    -- HOT engrams rendered as markdown sections.
    FOR v_item IN
        SELECT i FROM jsonb_array_elements(COALESCE(p_engrams -> 'items', '[]'::jsonb)) i
         WHERE i ->> 'tier' = 'hot'
         ORDER BY (i ->> 'id')
    LOOP
        -- Section heading from topic (or content preview if topic empty).
        v_md := v_md || '## ';
        IF (v_item ->> 'topic') IS NOT NULL AND length(v_item ->> 'topic') > 0 THEN
            v_md := v_md || (v_item ->> 'topic');
        ELSE
            v_md := v_md || substring(COALESCE(v_item ->> 'content', '(empty)') FROM 1 FOR 80);
        END IF;
        v_md := v_md || E'\n';

        -- Body.
        v_md := v_md || COALESCE(v_item ->> 'content', '') || E'\n';

        -- Preserved entities as markdown footer fields.
        SELECT string_agg(u, ', ' ORDER BY u) INTO v_urls_str
          FROM jsonb_array_elements_text(COALESCE(v_item -> 'preserved' -> 'urls', '[]'::jsonb)) u;
        IF v_urls_str IS NOT NULL AND v_urls_str <> '' THEN
            v_md := v_md || 'Sources: ' || v_urls_str || E'\n';
        END IF;

        SELECT string_agg(d, ', ' ORDER BY d) INTO v_dates_str
          FROM jsonb_array_elements_text(COALESCE(v_item -> 'preserved' -> 'dates', '[]'::jsonb)) d;
        IF v_dates_str IS NOT NULL AND v_dates_str <> '' THEN
            v_md := v_md || 'Dates: ' || v_dates_str || E'\n';
        END IF;

        SELECT string_agg(n, ', ' ORDER BY n) INTO v_names_str
          FROM jsonb_array_elements_text(COALESCE(v_item -> 'preserved' -> 'names', '[]'::jsonb)) n;
        IF v_names_str IS NOT NULL AND v_names_str <> '' THEN
            v_md := v_md || 'Names: ' || v_names_str || E'\n';
        END IF;

        SELECT string_agg('"' || q || '"', ' ' ORDER BY q) INTO v_quotes_str
          FROM jsonb_array_elements_text(COALESCE(v_item -> 'preserved' -> 'quotes', '[]'::jsonb)) q;
        IF v_quotes_str IS NOT NULL AND v_quotes_str <> '' THEN
            v_md := v_md || 'Quotes: ' || v_quotes_str || E'\n';
        END IF;

        v_md := v_md || E'\n';
    END LOOP;

    -- Footer with expand_message affordance.
    v_md := v_md ||
        '(MEDIUM/COLD engrams + raw content retrievable via expand_message(id=' ||
        p_message_id::text || ', tier=''hot''|''medium''|''cold''|''raw''))';

    RETURN v_md;
END;
$FN$;

COMMENT ON FUNCTION stewards.render_engrams_markdown(bigint, jsonb) IS
'Batch K.2: renders HOT-tier engrams as markdown for emission via compose_messages. Includes header (raw size + engram count), injection banner if suspected, per-engram sections with topic/content/preserved entities, and expand_message footer.';


-- ---------------------------------------------------------------------
-- 2. compose_messages rewrite with head/torso/tail logic.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.compose_messages(
    p_agent_family text,
    p_model text,
    p_session_id text,
    p_user_input text DEFAULT NULL
) RETURNS jsonb
LANGUAGE plpgsql STABLE AS $FN$
DECLARE
    v_system    text;
    v_history   jsonb;
    v_result    jsonb;
    v_tail_size int := 8;   -- last N messages preserved as raw
BEGIN
    v_system := stewards.compose_system_prompt(p_agent_family, p_model, p_session_id);

    WITH ordered AS (
        SELECT m.id, m.role, m.content, m.tool_call_id, m.tool_calls,
               m.reasoning_content, m.reasoning_details, m.engrams,
               ROW_NUMBER() OVER (ORDER BY m.created_at ASC, m.id ASC) AS pos,
               ROW_NUMBER() OVER (ORDER BY m.created_at DESC, m.id DESC) AS rn_from_end,
               -- Error-trace detection (LangChain rule: never compact errors).
               (m.content ~* '(traceback|exception|stack trace|panic:|HTTP [45]\d{2}|error from provider|error:)') AS is_error_trace
          FROM stewards.messages m
         WHERE m.session_id = p_session_id
    ),
    decided AS (
        SELECT *,
               -- Tail: most recent v_tail_size messages — always raw.
               -- Errors: always raw regardless of position.
               -- User/system: always raw (binding context).
               (rn_from_end <= v_tail_size
                OR is_error_trace
                OR role IN ('user', 'system')) AS preserve_raw,
               -- Eligible for engram emission: torso tool message with
               -- engrams populated (items array non-empty).
               (role = 'tool'
                AND engrams IS NOT NULL
                AND COALESCE(jsonb_array_length(engrams -> 'items'), 0) > 0
                AND rn_from_end > v_tail_size
                AND NOT is_error_trace) AS use_engrams
          FROM ordered
    )
    SELECT coalesce(jsonb_agg(
        CASE
            -- Torso tool message with engrams → emit HOT engrams as markdown.
            WHEN use_engrams THEN
                jsonb_build_object(
                    'role', 'tool',
                    'tool_call_id', coalesce(tool_call_id, ''),
                    'content', stewards.render_engrams_markdown(id, engrams)
                )
            -- Raw tool message (tail, no engrams, or error-trace).
            WHEN role = 'tool' THEN
                jsonb_build_object(
                    'role', 'tool',
                    'tool_call_id', coalesce(tool_call_id, ''),
                    'content', content
                )
            -- Tail assistant or error-tagged assistant: full fidelity.
            WHEN role = 'assistant' AND preserve_raw THEN
                jsonb_build_object('role', 'assistant', 'content', content)
                || (CASE WHEN tool_calls IS NOT NULL
                         THEN jsonb_build_object('tool_calls', tool_calls)
                         ELSE '{}'::jsonb END)
                || (CASE WHEN reasoning_content IS NOT NULL
                         THEN jsonb_build_object('reasoning_content', reasoning_content)
                         ELSE '{}'::jsonb END)
                || (CASE WHEN reasoning_details IS NOT NULL
                         THEN jsonb_build_object('reasoning_details', reasoning_details)
                         ELSE '{}'::jsonb END)
            -- Torso assistant: drop reasoning_content / reasoning_details
            -- (Anthropic pattern — agents rarely re-read their own old thinking).
            WHEN role = 'assistant' THEN
                jsonb_build_object('role', 'assistant', 'content', content)
                || (CASE WHEN tool_calls IS NOT NULL
                         THEN jsonb_build_object('tool_calls', tool_calls)
                         ELSE '{}'::jsonb END)
            -- User / system / fallback.
            ELSE
                jsonb_build_object('role', role, 'content', content)
        END
        ORDER BY pos
    ), '[]'::jsonb)
    INTO v_history
    FROM decided;

    v_result := jsonb_build_array(
        jsonb_build_object('role', 'system', 'content', v_system)
    ) || v_history;

    IF p_user_input IS NOT NULL THEN
        v_result := v_result || jsonb_build_array(
            jsonb_build_object('role', 'user', 'content', p_user_input)
        );
    END IF;

    RETURN v_result;
END;
$FN$;

COMMENT ON FUNCTION stewards.compose_messages(text, text, text, text) IS
'Batch K.2: rewrite with head/torso/tail compaction. Tail (last 8 messages) + error-trace messages + user/system messages always emit raw. Torso tool messages with engrams populated emit HOT engrams as markdown (via render_engrams_markdown). Torso assistant messages drop reasoning_content/reasoning_details. Backwards-compatible: messages without engrams emit raw, identical to pre-K.2 behavior.';


-- =====================================================================
-- End of k2-compose-messages-with-engrams.sql
-- =====================================================================
