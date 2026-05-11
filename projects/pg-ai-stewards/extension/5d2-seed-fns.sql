-- =====================================================================
-- Phase 5d.2 (Phase C.2) — seed SQL functions for intents + covenants
--
-- Consumes the Rust YAML parser helpers (yaml.rs):
--   stewards.parse_yaml_intent(yaml text)   RETURNS jsonb
--   stewards.parse_yaml_covenant(yaml text) RETURNS jsonb
--   stewards.yaml_sha256(yaml text)         RETURNS text
--
-- Idempotent: if source_yaml_sha matches the current YAML's sha,
-- the seed is a no-op (cheap unchanged-detection).
-- =====================================================================

-- ---------------------------------------------------------------------
-- seed_intents_from_yaml: parse + upsert by slug
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.seed_intents_from_yaml(p_yaml text)
RETURNS uuid
LANGUAGE plpgsql AS $func$
DECLARE
    v_parsed jsonb;
    v_sha    text;
    v_slug   text;
    v_existing_sha text;
    v_id     uuid;
BEGIN
    IF p_yaml IS NULL OR length(trim(p_yaml)) = 0 THEN
        RAISE EXCEPTION 'seed_intents_from_yaml: empty yaml';
    END IF;

    v_parsed := stewards.parse_yaml_intent(p_yaml)::jsonb;
    v_sha    := stewards.yaml_sha256(p_yaml);

    IF v_parsed ? 'error' THEN
        RAISE EXCEPTION 'seed_intents_from_yaml: %', v_parsed->>'error';
    END IF;

    v_slug := v_parsed->>'slug';
    IF v_slug IS NULL OR length(v_slug) = 0 THEN
        RAISE EXCEPTION 'seed_intents_from_yaml: parsed intent has no slug';
    END IF;

    -- Unchanged-detection: skip if sha matches
    SELECT source_yaml_sha, id INTO v_existing_sha, v_id
      FROM stewards.intents WHERE slug = v_slug;
    IF v_existing_sha IS NOT NULL AND v_existing_sha = v_sha THEN
        RETURN v_id;
    END IF;

    INSERT INTO stewards.intents (
        slug, purpose, beneficiary, values_hierarchy, non_goals,
        scripture_anchor, source_file, source_yaml_sha, updated_at
    ) VALUES (
        v_slug,
        v_parsed->>'purpose',
        v_parsed->>'beneficiary',
        coalesce(v_parsed->'values_hierarchy', '[]'::jsonb),
        coalesce(
            ARRAY(SELECT jsonb_array_elements_text(v_parsed->'non_goals')),
            ARRAY[]::text[]
        ),
        v_parsed->>'scripture_anchor',
        'intent.yaml',
        v_sha,
        now()
    )
    ON CONFLICT (slug) DO UPDATE SET
        purpose          = EXCLUDED.purpose,
        beneficiary      = EXCLUDED.beneficiary,
        values_hierarchy = EXCLUDED.values_hierarchy,
        non_goals        = EXCLUDED.non_goals,
        scripture_anchor = EXCLUDED.scripture_anchor,
        source_yaml_sha  = EXCLUDED.source_yaml_sha,
        updated_at       = now()
    RETURNING id INTO v_id;

    RETURN v_id;
END;
$func$;

COMMENT ON FUNCTION stewards.seed_intents_from_yaml(text) IS
'Phase 5d (C.2): parse intent.yaml and upsert into stewards.intents by slug. Returns the intent id. No-op if YAML sha matches existing row.';

-- ---------------------------------------------------------------------
-- seed_covenant_from_yaml: deactivate prior global, insert new
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
    -- The covenants_active_scope partial unique index would otherwise reject
    -- the insert until the prior row is deactivated.
    IF v_existing_id IS NOT NULL THEN
        UPDATE stewards.covenants
           SET deactivated_at = now()
         WHERE id = v_existing_id;
    END IF;

    INSERT INTO stewards.covenants (
        scope, human_commits_to, agent_commits_to,
        when_broken, recovery, council_moment,
        teaching_extension, ratified_by,
        source_file, source_yaml_sha
    ) VALUES (
        v_scope,
        coalesce(v_parsed->'human_commits_to', '[]'::jsonb),
        coalesce(v_parsed->'agent_commits_to', '[]'::jsonb),
        v_parsed->>'when_broken',
        v_parsed->>'recovery',
        v_parsed->>'council_moment',
        v_parsed->'teaching_extension',
        coalesce(v_parsed->>'ratified_by', 'both'),
        '.spec/covenant.yaml',
        v_sha
    ) RETURNING id INTO v_new_id;

    RETURN v_new_id;
END;
$func$;

COMMENT ON FUNCTION stewards.seed_covenant_from_yaml(text) IS
'Phase 5d (C.2): parse .spec/covenant.yaml and insert as the new active row in stewards.covenants. Deactivates any prior active row in the same scope. Returns new covenant id. No-op if YAML sha matches existing active row.';
