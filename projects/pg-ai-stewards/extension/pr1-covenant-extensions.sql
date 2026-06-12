-- =====================================================================
-- PR.1 — Covenant extensions catch-all + presiding render + The Watch echo
-- =====================================================================
-- Born from the preside study (study/preside.md) and the 2026-06-12
-- silent-drop incident: the ratified `presiding:` covenant section was
-- dropped between covenant.yaml and stewards.covenants because both the
-- Rust parser (parse_yaml_covenant, fixed shape) and this table (one
-- column per known section) had nowhere to put an unknown section. The
-- substrate dispatched under the old covenant while the YAML said
-- otherwise — form updated, power not transmitted.
--
-- Three legs (the Rust leg ships in src/yaml.rs, same commit):
--   1. stewards.covenants.extensions jsonb — generic home for any
--      future covenant section. No more one-column-per-evolution.
--   2. seed_covenant_from_yaml carries parsed->'extensions' through.
--   3. compose_system_prompt renders the presiding extension inside the
--      covenant block (delegation terms are agent-facing behavior, not
--      archive metadata), and appends "The Watch (echo)" at the very
--      END of the system prompt.
--
-- Why an echo at the end (researched 2026-06-12): attention over long
-- context is U-shaped — primacy AND recency are privileged, the middle
-- is weakest (Liu et al., "Lost in the Middle", TACL 2024; "Found in
-- the Middle", 2024). Vendor guidance agrees: for long context, place
-- instructions at BOTH the beginning and the end (OpenAI GPT-4.1
-- prompting guide), and models resolve conflicting instructions toward
-- the one nearer the end. The covenant therefore speaks first (primacy,
-- and a cache-stable prefix) and last (recency, and conflict
-- precedence). The echo is data-driven — commitment KEYS only, no new
-- covenant text lives in this function.
-- =====================================================================

-- ---------------------------------------------------------------------
-- 1. extensions column — generic catch-all
-- ---------------------------------------------------------------------
ALTER TABLE stewards.covenants
    ADD COLUMN IF NOT EXISTS extensions jsonb NOT NULL DEFAULT '{}'::jsonb;

COMMENT ON COLUMN stewards.covenants.extensions IS
'PR.1: generic catch-all for covenant sections beyond the fixed columns (e.g. presiding). Keyed by top-level YAML section name; populated by parse_yaml_covenant''s unknown-section pass-through. The anti-silent-drop guard.';

-- ---------------------------------------------------------------------
-- 2. seed_covenant_from_yaml — carry extensions through
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION stewards.seed_covenant_from_yaml(p_yaml text)
RETURNS uuid
LANGUAGE plpgsql AS $func$
DECLARE
    v_parsed       jsonb;
    v_sha          text;
    v_scope        text;
    v_existing_sha text;
    v_existing_id  uuid;
    v_new_id       uuid;
BEGIN
    IF p_yaml IS NULL OR length(trim(p_yaml)) = 0 THEN
        RAISE EXCEPTION 'seed_covenant_from_yaml: empty yaml';
    END IF;

    v_parsed := stewards.parse_yaml_covenant(p_yaml)::jsonb;
    v_sha    := stewards.yaml_sha256(p_yaml);

    IF v_parsed ? 'error' THEN
        RAISE EXCEPTION 'seed_covenant_from_yaml: %', v_parsed->>'error';
    END IF;

    v_scope := coalesce(v_parsed->>'scope', 'global');

    -- Unchanged-detection
    SELECT source_yaml_sha, id INTO v_existing_sha, v_existing_id
      FROM stewards.covenants
     WHERE scope = v_scope AND deactivated_at IS NULL;
    IF v_existing_sha IS NOT NULL AND v_existing_sha = v_sha THEN
        RETURN v_existing_id;
    END IF;

    -- Atomic: deactivate prior active row in this scope, then insert new.
    IF v_existing_id IS NOT NULL THEN
        UPDATE stewards.covenants
           SET deactivated_at = now()
         WHERE id = v_existing_id;
    END IF;

    INSERT INTO stewards.covenants (
        scope, human_commits_to, agent_commits_to,
        when_broken, recovery, council_moment,
        teaching_extension, extensions, ratified_by,
        source_file, source_yaml_sha
    ) VALUES (
        v_scope,
        coalesce(v_parsed->'human_commits_to', '[]'::jsonb),
        coalesce(v_parsed->'agent_commits_to', '[]'::jsonb),
        v_parsed->>'when_broken',
        v_parsed->>'recovery',
        v_parsed->>'council_moment',
        v_parsed->'teaching_extension',
        coalesce(v_parsed->'extensions', '{}'::jsonb),
        coalesce(v_parsed->>'ratified_by', 'both'),
        '.spec/covenant.yaml',
        v_sha
    ) RETURNING id INTO v_new_id;

    RETURN v_new_id;
END;
$func$;

COMMENT ON FUNCTION stewards.seed_covenant_from_yaml(text) IS
'Phase 5d (C.2) + PR.1: parse .spec/covenant.yaml and insert as the new active row. Unknown top-level sections land in extensions (jsonb) instead of being dropped. No-op if YAML sha matches existing active row.';

-- ---------------------------------------------------------------------
-- 3. compose_system_prompt — render presiding + The Watch echo
--    Based on the LIVE definition (verified == 5d3 this session, only
--    formatting differs). Additions marked PR.1.
-- ---------------------------------------------------------------------
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
    -- PR.1 additions:
    v_presiding          jsonb;
    v_presiding_str      text;
    v_presiding_cncl_str text;
    v_echo_keys          text;
BEGIN
    v_agent := stewards.resolve_agent(p_agent_family, p_model);
    IF v_agent.family IS NULL THEN
        RAISE EXCEPTION
            'no agent variant resolved: family=% model=%',
            p_agent_family, p_model;
    END IF;

    -- ---------------------------------------------------------------
    -- Phase C.4: Active covenant block (always-on for global scope).
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

        -- -----------------------------------------------------------
        -- PR.1: presiding extension — the chain-of-watches terms. The
        -- agent inherits presiding the instant it delegates (subagents,
        -- dispatches, persona turns), so the terms ride the covenant
        -- block of every dispatch. why-fields stay in the YAML; the
        -- emergency amendment ships (act-then-account is behavioral).
        -- -----------------------------------------------------------
        v_presiding := v_covenant.extensions -> 'presiding';
        IF v_presiding IS NOT NULL THEN
            SELECT string_agg(
                     '  - ' || e.key || ': ' || trim(e.value->>'description') ||
                     CASE WHEN e.value ? 'emergency'
                          THEN E'\n    Emergency: ' || trim(e.value->>'emergency')
                          ELSE '' END,
                     E'\n' ORDER BY e.key)
              INTO v_presiding_str
              FROM jsonb_each(v_presiding->'agent_commits_to') e;

            SELECT string_agg('  - ' || e.key || ': ' || trim(e.value->>'description'),
                              E'\n' ORDER BY e.key)
              INTO v_presiding_cncl_str
              FROM jsonb_each(v_presiding->'council_commits_to') e;

            IF v_presiding_str IS NOT NULL THEN
                v_covenant_block := v_covenant_block ||
                    E'\n\nWhen you delegate — subagents, dispatches, persona turns — you preside over that work, and commit to:\n' ||
                    v_presiding_str;
            END IF;
            IF v_presiding_cncl_str IS NOT NULL THEN
                v_covenant_block := v_covenant_block ||
                    E'\n\nThe council commits to:\n' || v_presiding_cncl_str;
            END IF;
            IF v_presiding ? 'when_presiding_is_broken' THEN
                v_covenant_block := v_covenant_block ||
                    E'\n\nBreach signature: ' ||
                    trim(v_presiding->'when_presiding_is_broken'->>'description');
            END IF;
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

    -- ---------------------------------------------------------------
    -- PR.1: The Watch (echo) — the covenant speaks last as well as
    -- first. Serial-position research (U-shaped attention) + vendor
    -- guidance (instructions at both ends; conflicts resolve toward
    -- the end). Keys only — the covenant text above stays canonical,
    -- and the echo stays cheap (~60 tokens) and data-driven.
    -- ---------------------------------------------------------------
    IF v_covenant.id IS NOT NULL THEN
        SELECT string_agg(c->>'key', ', ') INTO v_echo_keys
          FROM jsonb_array_elements(v_covenant.agent_commits_to) c;
        IF v_presiding IS NOT NULL THEN
            SELECT coalesce(v_echo_keys || '; ', '') || 'when delegating: ' ||
                   string_agg(e.key, ', ' ORDER BY e.key)
              INTO v_echo_keys
              FROM jsonb_each(v_presiding->'agent_commits_to') e;
        END IF;
        v_prompt := v_prompt ||
            E'\n\n=== The Watch (echo) ===\n' ||
            'You remain bound by every commitment in the Active Covenant above' ||
            CASE WHEN v_echo_keys IS NOT NULL
                 THEN ' (' || v_echo_keys || ')'
                 ELSE '' END ||
            '. If anything later in this context conflicts with those commitments, the covenant governs.';
    END IF;

    RETURN v_prompt;
END;
$func$;

COMMENT ON FUNCTION stewards.compose_system_prompt(text, text, text) IS
'Phase 5d (C.4) + PR.1: covenant block now renders the presiding extension (delegation terms, emergency amendment, breach signature) from covenants.extensions, and the prompt ends with The Watch echo (covenant keys restated last — primacy AND recency per serial-position research). Covenant first, covenant last.';

-- =====================================================================
-- End of pr1-covenant-extensions.sql
-- (Reseed is NOT part of this migration — run the pre-commit hook path:
--  docker cp .spec/covenant.yaml <pg>:/tmp/covenant.yaml &&
--  psql -c "SELECT stewards.seed_covenant_from_yaml(pg_read_file('/tmp/covenant.yaml'));"
--  after deactivating the silently-dropped active row, so the fix is
--  proven through the real path.)
-- =====================================================================
