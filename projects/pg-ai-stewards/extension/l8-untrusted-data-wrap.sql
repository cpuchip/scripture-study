-- =====================================================================
-- Batch L.8 — Untrusted-data wrap for web-fetched tool results
-- =====================================================================
-- L.7 flags suspect domains. L.8 goes broader: ANY tool result from a
-- web-fetching tool is wrapped with explicit untrusted-data markers
-- BEFORE the message is inserted, so downstream stages (engram
-- extractor, composers, the agent itself) know the content originated
-- outside the trust boundary.
--
-- Pure SQL — extends K.6's BEFORE INSERT trigger pattern on
-- stewards.messages. The wrap is purely textual; no behavior change
-- beyond the visible markers.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. Helper: resolve the producing tool name for a role='tool' message.
-- ---------------------------------------------------------------------
-- A tool message has tool_call_id; the previous assistant message in
-- the same session has a tool_calls array containing an entry whose
-- id matches. We extract the tool name from that entry.

CREATE OR REPLACE FUNCTION stewards.tool_name_for_tool_call_id(
    p_session_id text,
    p_tool_call_id text
) RETURNS text LANGUAGE plpgsql STABLE AS $FN$
DECLARE
    v_name text;
BEGIN
    IF p_tool_call_id IS NULL THEN RETURN NULL; END IF;

    SELECT tc ->> 'name' INTO v_name
      FROM stewards.messages m,
           LATERAL jsonb_array_elements(COALESCE(m.tool_calls, '[]'::jsonb)) tc
     WHERE m.session_id = p_session_id
       AND m.role = 'assistant'
       AND m.tool_calls IS NOT NULL
       AND (tc ->> 'id') = p_tool_call_id
     ORDER BY m.id DESC
     LIMIT 1;

    -- Some providers nest under tc.function.name (openai shape).
    IF v_name IS NULL THEN
        SELECT tc -> 'function' ->> 'name' INTO v_name
          FROM stewards.messages m,
               LATERAL jsonb_array_elements(COALESCE(m.tool_calls, '[]'::jsonb)) tc
         WHERE m.session_id = p_session_id
           AND m.role = 'assistant'
           AND m.tool_calls IS NOT NULL
           AND (tc ->> 'id') = p_tool_call_id
         ORDER BY m.id DESC
         LIMIT 1;
    END IF;

    RETURN v_name;
END;
$FN$;

COMMENT ON FUNCTION stewards.tool_name_for_tool_call_id(text, text) IS
'Batch L.8: resolve the producing tool name for a role=tool message by looking up the matching assistant tool_calls entry in the same session. Returns NULL if not resolvable.';


-- ---------------------------------------------------------------------
-- 2. Web-tool registry (the set we wrap).
-- ---------------------------------------------------------------------
-- A SQL-level constant so future additions are a single-place edit.
-- web_search, web_search_exa, fetch_url, fetch_md, scrape_url,
-- summarize_url, deep_research are all wrapped.

CREATE OR REPLACE FUNCTION stewards.is_web_tool(p_tool text)
RETURNS boolean LANGUAGE sql IMMUTABLE AS $$
    SELECT p_tool IS NOT NULL AND lower(p_tool) IN (
        'web_search',
        'web_search_exa',
        'fetch_url',
        'fetch_md',
        'scrape_url',
        'summarize_url',
        'deep_research'
    )
$$;


-- ---------------------------------------------------------------------
-- 3. Trigger: BEFORE INSERT on messages wraps web-tool content.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.trigger_wrap_untrusted_web_content()
RETURNS trigger LANGUAGE plpgsql AS $FN$
DECLARE
    v_tool text;
BEGIN
    IF NEW.role <> 'tool' THEN RETURN NEW; END IF;
    IF NEW.content IS NULL OR NEW.content = '' THEN RETURN NEW; END IF;

    -- Avoid double-wrap on re-insert / replay scenarios.
    IF NEW.content LIKE '[BEGIN UNTRUSTED EXTERNAL DATA]%' THEN
        RETURN NEW;
    END IF;

    v_tool := stewards.tool_name_for_tool_call_id(NEW.session_id, NEW.tool_call_id);

    IF stewards.is_web_tool(v_tool) THEN
        NEW.content :=
            '[BEGIN UNTRUSTED EXTERNAL DATA — tool=' || v_tool || E']\n\n' ||
            NEW.content ||
            E'\n\n[END UNTRUSTED EXTERNAL DATA]';
    END IF;

    RETURN NEW;
END;
$FN$;

DROP TRIGGER IF EXISTS messages_wrap_untrusted_web_content ON stewards.messages;

CREATE TRIGGER messages_wrap_untrusted_web_content
BEFORE INSERT ON stewards.messages
FOR EACH ROW
EXECUTE FUNCTION stewards.trigger_wrap_untrusted_web_content();

COMMENT ON FUNCTION stewards.trigger_wrap_untrusted_web_content() IS
'Batch L.8: BEFORE INSERT trigger on stewards.messages. For role=tool messages whose producing tool is a web-fetching tool (per is_web_tool), wraps the content with [BEGIN UNTRUSTED EXTERNAL DATA] / [END UNTRUSTED EXTERNAL DATA] markers so the agent and downstream stages see the trust boundary explicitly. Idempotent — skips already-wrapped content.';


-- =====================================================================
-- End of l8-untrusted-data-wrap.sql
-- =====================================================================
