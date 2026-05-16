-- =====================================================================
-- ES.3.s3 — consult_subagent: re-engage a persistent sub-agent
-- =====================================================================
-- The companion to spawn_subagent (K.4). spawn_subagent creates a child
-- and returns its digest; the sub-agent's session then PERSISTS. This
-- phase adds the missing half: send that sub-agent a new question, in
-- the context it already built.
--
-- A report you file once becomes a steward you can send back
-- (D&C 104; Matthew 13:52 — the householder's treasure yields things
-- new and old).
--
-- consult_subagent_dispatch enqueues the re-engagement chat:
--   - judge session ('judge-<msgid>'): a manual body rebuilds the
--     document (from messages_raw_overflow) + the prior brief + the
--     new question — the judge re-reads the SAME document on a new
--     angle. Prompt caching keeps the re-ask cheap.
--   - any other sub-agent session: chat_post_internal continuation —
--     the session already holds its full message history.
--
-- Soft cap (decision 5): after 5 re-asks on one session a STEWARD
-- NOTICE is prepended (the L.1.1.17 judge pattern — surface the cost,
-- let the agent judge). The work_item cost cap is the hard backstop.
--
-- The Go handler (cmd/stewards-mcp/consult_subagent.go) does the sync
-- poll + answer extraction, mirroring spawn_subagent.go.
-- =====================================================================


CREATE OR REPLACE FUNCTION stewards.consult_subagent_dispatch(
    p_session_id text,
    p_question   text
) RETURNS bigint LANGUAGE plpgsql AS $FN$
DECLARE
    v_soft_cap   constant int := 5;
    v_prior      int;
    v_question   text;
    v_judge_msgid bigint;
    v_document   text;
    v_binding    text;
    v_prior_ans  text;
    v_agent      stewards.agents;
    v_body       jsonb;
    v_payload    jsonb;
    v_wq_id      bigint;
    v_family     text;
    v_model      text;
    v_provider   text;
BEGIN
    IF NOT EXISTS (SELECT 1 FROM stewards.sessions WHERE id = p_session_id) THEN
        RAISE EXCEPTION 'consult_subagent_dispatch: session % not found', p_session_id;
    END IF;
    IF COALESCE(trim(p_question), '') = '' THEN
        RAISE EXCEPTION 'consult_subagent_dispatch: question is empty';
    END IF;

    -- Prior re-asks on this session (the [CONSULT] user messages).
    SELECT count(*) INTO v_prior
      FROM stewards.messages
     WHERE session_id = p_session_id
       AND role = 'user'
       AND content LIKE '[CONSULT]%';

    v_question := p_question;
    IF v_prior >= v_soft_cap THEN
        v_question :=
            E'[STEWARD NOTICE — soft cap reached]\n'
         || E'You have re-engaged this sub-agent ' || v_prior::text
         || E' times (soft cap ' || v_soft_cap::text || E'). Each re-ask spends real budget. '
         || E'If you can answer your binding question from what you already hold, do that '
         || E'instead. If this consult is genuinely needed, proceed and it will be honored.'
         || E'\n\n' || p_question;
    END IF;

    -- Record the consult question (counting + audit trail).
    INSERT INTO stewards.messages (session_id, role, content)
    VALUES (p_session_id, 'user', '[CONSULT] ' || v_question);

    -- ---- Judge session: rebuild the document context manually -------
    IF p_session_id LIKE 'judge-%' THEN
        v_judge_msgid := NULLIF(substring(p_session_id FROM 7), '')::bigint;

        SELECT content, binding_question
          INTO v_document, v_binding
          FROM stewards.messages_raw_overflow
         WHERE message_id = v_judge_msgid
         ORDER BY parent_ordinal ASC
         LIMIT 1;

        IF v_document IS NULL THEN
            RAISE EXCEPTION 'consult_subagent_dispatch: no preserved document for judge session % (msg %)',
                p_session_id, v_judge_msgid;
        END IF;

        -- The judge's most recent answer in this session, for continuity.
        SELECT content INTO v_prior_ans
          FROM stewards.messages
         WHERE session_id = p_session_id AND role = 'assistant'
         ORDER BY id DESC LIMIT 1;

        SELECT * INTO v_agent
          FROM stewards.agents WHERE family = 'judge-brief' AND active LIMIT 1;
        IF v_agent.family IS NULL THEN
            RAISE EXCEPTION 'consult_subagent_dispatch: judge-brief agent not registered';
        END IF;

        v_body := jsonb_build_object(
            'model', 'deepseek-v4-flash',
            'messages', jsonb_build_array(
                jsonb_build_object('role','system','content', v_agent.prompt),
                jsonb_build_object('role','user','content',
                    E'BINDING QUESTION:\n' || COALESCE(v_binding,'(none)') ||
                    E'\n\nDOCUMENT (' || length(v_document)::text || E' chars):\n---\n' ||
                    v_document || E'\n---'),
                jsonb_build_object('role','assistant','content',
                    COALESCE(v_prior_ans, '(prior brief unavailable)')),
                jsonb_build_object('role','user','content',
                    E'FOLLOW-UP — re-judge the SAME document for this new question:\n'
                    || v_question ||
                    E'\n\nOutput ONLY the JSON brief, scoped to this follow-up.')
            ),
            'temperature', v_agent.temperature
        );
        IF v_agent.response_format IS NOT NULL THEN
            v_body := v_body || jsonb_build_object('response_format', v_agent.response_format);
        END IF;

        v_payload := jsonb_build_object(
            'session_id', p_session_id,
            'agent_family', 'judge-brief',
            'requested_model', 'deepseek-v4-flash',
            'body', v_body,
            'tools_disabled', true,
            '_consult_subagent_session', p_session_id,
            '_consult_reask_index', v_prior + 1
        );

        INSERT INTO stewards.work_queue (kind, provider, payload, status)
        VALUES ('chat', 'opencode_go', v_payload, 'pending')
        RETURNING id INTO v_wq_id;

        RAISE NOTICE 'consult_subagent_dispatch: judge session % re-engaged, chat wq=% (re-ask #%)',
            p_session_id, v_wq_id, v_prior + 1;
        RETURN v_wq_id;
    END IF;

    -- ---- Any other sub-agent session: normal continuation -----------
    -- The session holds its own message history; chat_post_internal
    -- composes it (including the [CONSULT] user message just inserted).
    SELECT payload ->> 'agent_family', payload ->> 'requested_model', provider
      INTO v_family, v_model, v_provider
      FROM stewards.work_queue
     WHERE kind = 'chat'
       AND payload ->> 'session_id' = p_session_id
     ORDER BY id DESC LIMIT 1;

    IF v_family IS NULL THEN
        RAISE EXCEPTION 'consult_subagent_dispatch: cannot resolve agent for session % (no prior chat)',
            p_session_id;
    END IF;

    SELECT stewards.chat_post_internal(v_family, v_model, p_session_id, v_provider)
      INTO v_wq_id;

    -- Tag the freshly-enqueued chat so the Go handler can poll it.
    UPDATE stewards.work_queue
       SET payload = payload || jsonb_build_object(
               '_consult_subagent_session', p_session_id,
               '_consult_reask_index', v_prior + 1)
     WHERE id = v_wq_id;

    RAISE NOTICE 'consult_subagent_dispatch: session % re-engaged via chat_post_internal, chat wq=% (re-ask #%)',
        p_session_id, v_wq_id, v_prior + 1;
    RETURN v_wq_id;
END;
$FN$;

COMMENT ON FUNCTION stewards.consult_subagent_dispatch(text, text) IS
'ES.3.s3: enqueues a re-engagement chat into an existing sub-agent session. Judge sessions rebuild the document context manually; other sessions continue via chat_post_internal. Soft cap 5 re-asks (STEWARD NOTICE prepended past it). The Go handler consult_subagent.go does the sync wait.';


-- ---------------------------------------------------------------------
-- tool_def — consult_subagent (routes to the pg-ai-stewards MCP server).
-- ---------------------------------------------------------------------

INSERT INTO stewards.tool_defs (name, description, args_schema, execute_target, active)
VALUES (
    'consult_subagent',
    'Re-engage a sub-agent you (or the substrate) already spawned — send it a NEW question in the context it already built. '
    || 'For a judge that compiled a brief from an oversized fetch, this re-reads the SAME document on a new angle without re-fetching it. '
    || 'The sub-agent answers from its own context window; you only see its answer. '
    || 'Use when a prior brief or sub-agent digest did not cover something you now need. '
    || 'Pass session_id (e.g. a judge brief names its session as judge-<id>) or a sub-agent work_item id.',
    $JSON$
    {
      "type": "object",
      "required": ["target", "question"],
      "additionalProperties": false,
      "properties": {
        "target": {
          "type": "string",
          "description": "The sub-agent to re-engage: a session_id (e.g. 'judge-5636') or a spawned work_item uuid."
        },
        "question": {
          "type": "string",
          "description": "The new question for the sub-agent. Tightly scoped — it answers from the context it already holds."
        }
      }
    }
    $JSON$::jsonb,
    jsonb_build_object('kind', 'mcp_proxy', 'server', 'pg-ai-stewards', 'tool', 'consult_subagent'),
    true
)
ON CONFLICT (name) DO UPDATE
   SET description    = EXCLUDED.description,
       args_schema    = EXCLUDED.args_schema,
       execute_target = EXCLUDED.execute_target,
       active         = true;


-- =====================================================================
-- End of es8-consult-subagent.sql
-- =====================================================================
