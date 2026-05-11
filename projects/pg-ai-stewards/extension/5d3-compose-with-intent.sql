-- =====================================================================
-- Phase 5d.3 (Phase C.4) — compose_system_prompt extended to inject
-- active covenant + work_item intent
--
-- Extends the existing compose_system_prompt (src/schema.rs:686) by
-- prepending two structured blocks before the agent prompt:
--
--   === Active Covenant ===
--   <agent commitments rendered as bullets>
--
--   === Intent ===
--   Purpose: ...
--   Values: ...
--
-- Per ratifications:
--   D-C1/C2: covenant + intent are loaded from the substrate (which
--            mirrors the canonical YAML files)
--   D-C4: free-form covenant judgment — the gate prompt does the
--         honors-the-covenant evaluation; compose_system_prompt
--         just makes it visible to the dispatched chat
--
-- Sessions without a work_item (ad-hoc chats, watchman) get only the
-- global covenant block. The intent block is omitted when no
-- work_item resolves.
-- =====================================================================

CREATE OR REPLACE FUNCTION stewards.compose_system_prompt(
    p_agent_family text, p_model text, p_session_id text
) RETURNS text
LANGUAGE plpgsql STABLE AS $func$
DECLARE
    v_agent          stewards.agents;
    v_prompt         text := '';
    v_instructions   text;
    v_skills_block   text;
    v_covenant       stewards.covenants;
    v_intent         stewards.intents;
    v_covenant_block text := '';
    v_intent_block   text := '';
    v_human_str      text;
    v_agent_str      text;
    v_values_str     text;
    v_non_goals_str  text;
BEGIN
    v_agent := stewards.resolve_agent(p_agent_family, p_model);
    IF v_agent.family IS NULL THEN
        RAISE EXCEPTION
            'no agent variant resolved: family=% model=%',
            p_agent_family, p_model;
    END IF;

    -- ---------------------------------------------------------------
    -- Phase C.4: Active covenant block (always-on for global scope).
    -- Most-specific scope match wins; for now, just global.
    -- ---------------------------------------------------------------
    SELECT * INTO v_covenant
      FROM stewards.covenants
     WHERE scope = 'global' AND deactivated_at IS NULL
     ORDER BY activated_at DESC
     LIMIT 1;

    IF v_covenant.id IS NOT NULL THEN
        SELECT string_agg('  - ' || (c->>'key') || ': ' || (c->>'description'), E'\n')
          INTO v_human_str
          FROM jsonb_array_elements(v_covenant.human_commits_to) c;

        SELECT string_agg('  - ' || (c->>'key') || ': ' || (c->>'description'), E'\n')
          INTO v_agent_str
          FROM jsonb_array_elements(v_covenant.agent_commits_to) c;

        v_covenant_block :=
            E'=== Active Covenant ===\n' ||
            E'The human commits to:\n' || coalesce(v_human_str, '  (none)') || E'\n\n' ||
            E'The agent (you) commits to:\n' || coalesce(v_agent_str, '  (none)');

        IF v_covenant.council_moment IS NOT NULL AND length(v_covenant.council_moment) > 0 THEN
            v_covenant_block := v_covenant_block || E'\n\nCouncil moment:\n  ' || v_covenant.council_moment;
        END IF;
    END IF;

    -- ---------------------------------------------------------------
    -- Phase C.4: Intent block (only when session resolves to a work_item
    -- with an intent_id).
    -- ---------------------------------------------------------------
    SELECT i.* INTO v_intent
      FROM stewards.intents i
      JOIN stewards.work_items wi ON wi.intent_id = i.id
     WHERE p_session_id = ANY(coalesce(wi.session_ids, ARRAY[]::text[]))
     LIMIT 1;

    IF v_intent.id IS NOT NULL THEN
        SELECT string_agg(
                 '  - ' || (v->>'key') ||
                 CASE WHEN v ? 'kind' AND v->>'kind' = 'constraint'
                      THEN ' [constraint, severity=' || coalesce(v->>'severity','?') || ']'
                      ELSE ''
                 END ||
                 ': ' || (v->>'description'),
                 E'\n'
               )
          INTO v_values_str
          FROM jsonb_array_elements(v_intent.values_hierarchy) v;

        v_non_goals_str := array_to_string(v_intent.non_goals, E'\n  - ', '');

        v_intent_block :=
            E'=== Intent ===\n' ||
            E'Slug: ' || v_intent.slug || E'\n' ||
            E'Purpose: ' || v_intent.purpose || E'\n';

        IF v_intent.beneficiary IS NOT NULL THEN
            v_intent_block := v_intent_block || E'Beneficiary: ' || v_intent.beneficiary || E'\n';
        END IF;

        v_intent_block := v_intent_block || E'\nValues (in order of priority):\n' ||
            coalesce(v_values_str, '  (none)');

        IF v_intent.non_goals IS NOT NULL AND array_length(v_intent.non_goals, 1) > 0 THEN
            v_intent_block := v_intent_block || E'\n\nNon-goals:\n  - ' || v_non_goals_str;
        END IF;

        IF v_intent.scripture_anchor IS NOT NULL THEN
            v_intent_block := v_intent_block || E'\n\nScripture anchor: ' || v_intent.scripture_anchor;
        END IF;
    END IF;

    -- Compose the final prompt. Phase C blocks first (covenant, intent),
    -- then === Agent === marker, then existing agent + instructions + skills.
    IF length(v_covenant_block) > 0 THEN
        v_prompt := v_covenant_block || E'\n\n';
    END IF;
    IF length(v_intent_block) > 0 THEN
        v_prompt := v_prompt || v_intent_block || E'\n\n';
    END IF;
    IF length(v_prompt) > 0 THEN
        v_prompt := v_prompt || E'=== Agent ===\n';
    END IF;

    v_prompt := v_prompt || v_agent.prompt;

    -- ---------------------------------------------------------------
    -- Existing logic: instructions + skills
    -- ---------------------------------------------------------------
    SELECT string_agg(body, E'\n\n' ORDER BY ord, family)
    INTO v_instructions
    FROM (
        SELECT DISTINCT ON (family)
            family, body, ord
        FROM stewards.instructions
        WHERE active
          AND scope IN ('global', 'agent:' || p_agent_family)
          AND stewards.glob_match(model_match, p_model)
        ORDER BY family, length(model_match) DESC, model_match
    ) t;
    IF v_instructions IS NOT NULL THEN
        v_prompt := v_prompt || E'\n\n' || v_instructions;
    END IF;

    IF stewards.tool_permission(p_agent_family, 'skill') <> 'deny' THEN
        SELECT E'\n\n<available_skills>\n' || string_agg(
            '  <skill>' || E'\n'
            || '    <name>' || family || '</name>' || E'\n'
            || '    <description>' || description || '</description>' || E'\n'
            || '  </skill>',
            E'\n'
            ORDER BY family
        ) || E'\n</available_skills>'
        INTO v_skills_block
        FROM (
            SELECT DISTINCT ON (family) family, description
            FROM stewards.skills
            WHERE active
              AND stewards.glob_match(model_match, p_model)
              AND stewards.skill_permission(p_agent_family, family) <> 'deny'
            ORDER BY family, length(model_match) DESC, model_match
        ) s;
        IF v_skills_block IS NOT NULL THEN
            v_prompt := v_prompt || v_skills_block;
        END IF;
    END IF;

    RETURN v_prompt;
END;
$func$;

COMMENT ON FUNCTION stewards.compose_system_prompt(text, text, text) IS
'Phase 5d (C.4): now prepends active covenant + work_item intent (when session resolves to a work_item with intent_id) before the agent block. Sessions without a work_item get only the global covenant.';
