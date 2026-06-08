-- =====================================================================
-- CT2.7b — Durable self-notes: the agent-callable remember/forget tools
-- =====================================================================
-- spec §7. The WRITE path (the Hermes self-curation loop). Uses the
-- _session_id the CT2.3 dispatcher already injects into sql_fn args; derives
-- the authoring agent_family from the session (no extra Rust). Pure SQL.
--
-- Default audience = the authoring agent_family (Gate-1 default; 7c upgrades
-- chat personas to {persona:self} once the persona facet is threaded). Write
-- cap = 40 notes per author (forces the prune loop). compose_tools gate
-- extended so remember/forget are exposed only to context_tools_enabled
-- families (like the context_* tools).
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. Derive the authoring agent_family from a session (for default audience
--    + created_by). The current stage's agent_family in the pipeline.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.session_agent_family(p_session_id text)
RETURNS text LANGUAGE sql STABLE AS $$
    SELECT s.elem ->> 'agent_family'
      FROM stewards.work_items w
      JOIN stewards.pipelines p ON p.family = w.pipeline_family
      CROSS JOIN LATERAL jsonb_array_elements(p.stages) AS s(elem)
     WHERE p_session_id = ANY(w.session_ids)
       AND s.elem ->> 'name' = w.current_stage
     ORDER BY w.id DESC
     LIMIT 1;
$$;


-- ---------------------------------------------------------------------
-- 2. remember(note, audience?, tags?) — add a durable self-note.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.remember_tool(p_args jsonb)
RETURNS jsonb LANGUAGE plpgsql AS $FN$
DECLARE
    v_sess  text := p_args ->> '_session_id';
    v_note  text := p_args ->> 'note';
    v_aud   jsonb := p_args -> 'audience';
    v_tags  text[];
    v_fam   text := stewards.session_agent_family(v_sess);
    v_owner text := COALESCE(v_fam, v_sess);
    v_count int;
    v_id    bigint;
    v_cap   int := 40;
BEGIN
    IF v_note IS NULL OR length(btrim(v_note)) = 0 THEN
        RETURN jsonb_build_object('error', 'note text required');
    END IF;

    -- default audience = the authoring agent_family (durable + scoped); fall
    -- back to session if the family couldn't be derived.
    IF v_aud IS NULL OR jsonb_typeof(v_aud) <> 'object' OR v_aud = '{}'::jsonb THEN
        v_aud := CASE WHEN v_fam IS NOT NULL
                      THEN jsonb_build_object('agent_family', v_fam)
                      ELSE jsonb_build_object('session', v_sess) END;
    END IF;

    IF p_args ? 'tags' AND jsonb_typeof(p_args -> 'tags') = 'array' THEN
        SELECT array_agg(t) INTO v_tags FROM jsonb_array_elements_text(p_args -> 'tags') t;
    END IF;

    SELECT count(*) INTO v_count FROM stewards.agent_self_notes WHERE created_by = v_owner;
    IF v_count >= v_cap THEN
        RETURN jsonb_build_object('error',
            format('note budget full (%s/%s for %s) — forget() an integrated one first', v_count, v_cap, v_owner));
    END IF;

    INSERT INTO stewards.agent_self_notes (note, audience, tags, created_by, created_session)
    VALUES (v_note, v_aud, COALESCE(v_tags, '{}'), v_owner, v_sess)
    RETURNING id INTO v_id;

    RETURN jsonb_build_object('ok', true,
        'handle', stewards.context_note_handle(v_id), 'audience', v_aud, 'note_id', v_id);
END;
$FN$;


-- ---------------------------------------------------------------------
-- 3. forget(handle) — drop a durable note this dispatch can see / authored.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.forget_tool(p_args jsonb)
RETURNS jsonb LANGUAGE plpgsql AS $FN$
DECLARE
    v_sess   text := p_args ->> '_session_id';
    v_handle text := lower(substring(COALESCE(p_args ->> 'handle', '') FROM '([0-9a-fA-F]{4})'));
    v_fam    text := stewards.session_agent_family(v_sess);
    v_owner  text := COALESCE(v_fam, v_sess);
    v_facets jsonb := stewards.dispatch_facets(COALESCE(v_fam, '~none~'), v_sess);
    v_deleted int;
BEGIN
    IF v_handle IS NULL THEN
        RETURN jsonb_build_object('error', 'handle required (the [note:xxxx] of the note to drop)');
    END IF;
    WITH del AS (
        DELETE FROM stewards.agent_self_notes n
         WHERE stewards.context_note_handle(n.id) = v_handle
           AND (v_facets @> n.audience OR n.created_by = v_owner)
        RETURNING n.id
    )
    SELECT count(*) INTO v_deleted FROM del;
    IF v_deleted = 0 THEN
        RETURN jsonb_build_object('error', 'no note [note:' || v_handle || '] you can forget in this context');
    END IF;
    RETURN jsonb_build_object('ok', true, 'forgotten', v_handle, 'count', v_deleted);
END;
$FN$;


-- ---------------------------------------------------------------------
-- 4. tool_defs (sql_fn). _session_id is injected by the dispatcher (CT2.3).
-- ---------------------------------------------------------------------
INSERT INTO stewards.tool_defs (name, description, args_schema, execute_target, active)
VALUES
('remember',
 'Save a durable note to your FUTURE self — it survives context compaction AND session boundaries, rendered back to you in YOUR DURABLE NOTES. Use it to park a fact you''ll need later or a self-tuning reminder. audience routes WHO sees it: default = your own agent family; {global:true} = everyone; {kind:"code"} = all code-kind agents; etc. Keep notes few and curated — forget() them once integrated (you have a budget).',
 '{"type":"object","required":["note"],"additionalProperties":false,"properties":{"note":{"type":"string","description":"The durable note text."},"audience":{"type":"object","description":"Optional routing selectors, e.g. {\"global\":true} or {\"kind\":\"code\"}. Default: your own agent family."},"tags":{"type":"array","items":{"type":"string"},"description":"Optional free-form labels for search/organization."}}}'::jsonb,
 jsonb_build_object('kind','sql_fn','name','remember_tool','schema','stewards'), true),
('forget',
 'Drop one of YOUR durable notes by its [note:xxxx] handle — do this once you''ve integrated the fact elsewhere (the self-curation loop; your note budget is finite).',
 '{"type":"object","required":["handle"],"additionalProperties":false,"properties":{"handle":{"type":"string","description":"The 4-char handle, e.g. f139 or [note:f139]."}}}'::jsonb,
 jsonb_build_object('kind','sql_fn','name','forget_tool','schema','stewards'), true)
ON CONFLICT (name) DO UPDATE
   SET description = EXCLUDED.description,
       args_schema = EXCLUDED.args_schema,
       execute_target = EXCLUDED.execute_target,
       active = true;


-- ---------------------------------------------------------------------
-- 5. compose_tools gate — also gate remember/forget on context_tools_enabled.
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.compose_tools(p_agent_family text)
RETURNS jsonb LANGUAGE sql STABLE AS $function$
    SELECT coalesce(jsonb_agg(
        jsonb_build_object(
            'type', 'function',
            'function', jsonb_build_object(
                'name', t.name,
                'description', t.description,
                'parameters', t.args_schema
            )
        )
        ORDER BY t.name
    ), '[]'::jsonb)
    FROM stewards.tool_defs t
    WHERE t.active
      AND stewards.tool_permission(p_agent_family, t.name) <> 'deny'
      AND ( (t.name NOT LIKE 'context\_%' ESCAPE '\' AND t.name NOT IN ('remember','forget'))
            OR stewards.context_tools_on(p_agent_family) )
$function$;

COMMENT ON FUNCTION stewards.compose_tools(text) IS
'Active tool_defs not denied for the family. CT2.3/§7: context_* tools AND remember/forget are gated — included only when the family has context_tools_enabled.';


-- =====================================================================
-- End of ct2-7b-self-notes-tools.sql
-- =====================================================================
