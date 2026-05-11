-- =====================================================================
-- Phase 5d — Phase C.1 schema: intents + covenants tables
--
-- Phase C of the full agentic substrate makes intent and covenant
-- first-class substrate state. Today they live in YAML files
-- (intent.yaml, .spec/covenant.yaml) and the dispatched LLM never
-- sees them directly. C.1 lays the schema; subsequent C.x phases
-- add the YAML parser, seed SQL functions, prompt composition
-- extension, and UI surfaces.
--
-- Per ratifications:
--   D-C1: YAML canonical at repo root; substrate is a runtime mirror
--   D-C2: Same for .spec/covenant.yaml
--   D-C3: intent_id required at work_item creation (NOT NULL added in C.5
--         after a default 'scripture-study' intent is seeded + existing rows
--         backfilled)
--
-- Idempotent — IF NOT EXISTS / DO blocks throughout. Safe to re-apply.
-- =====================================================================

-- ---------------------------------------------------------------------
-- Section 1: stewards.intents — reusable across work_items + pipelines
-- ---------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS stewards.intents (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug                text UNIQUE NOT NULL,
    purpose             text NOT NULL,
    beneficiary         text,
    values_hierarchy    jsonb NOT NULL DEFAULT '[]'::jsonb,
    non_goals           text[] DEFAULT ARRAY[]::text[],
    scripture_anchor    text,
    source_file         text,
    source_yaml_sha     text,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE stewards.intents IS
'Phase 5d (C.1): the why behind a work_item. YAML canonical; substrate is the runtime mirror per D-C1.';
COMMENT ON COLUMN stewards.intents.values_hierarchy IS
'Ordered list of trade-off priorities. Today seeded as [{key, description, source}] preserving order via array semantics from intent.yaml values: map.';
COMMENT ON COLUMN stewards.intents.source_file IS
'Relative path to the YAML this intent was seeded from. NULL for substrate-native intents created via /api/intents/create.';
COMMENT ON COLUMN stewards.intents.source_yaml_sha IS
'sha256 hex of the YAML at last seed. Skip re-seeding if unchanged.';

-- ---------------------------------------------------------------------
-- Section 2: stewards.covenants — bilateral commitments
-- ---------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS stewards.covenants (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    scope               text NOT NULL,
    human_commits_to    jsonb NOT NULL,
    agent_commits_to    jsonb NOT NULL,
    when_broken         text,
    recovery            text,
    council_moment      text,
    teaching_extension  jsonb,
    activated_at        timestamptz NOT NULL DEFAULT now(),
    deactivated_at      timestamptz,
    ratified_by         text NOT NULL,
    source_file         text,
    source_yaml_sha     text,
    created_at          timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE stewards.covenants IS
'Phase 5d (C.1): bilateral commitments. Typically one active row scoped global. YAML canonical per D-C2.';
COMMENT ON COLUMN stewards.covenants.scope IS
'global | pipeline:<family> | work_item:<id>. Most-specific active row wins at compose_system_prompt time.';

-- Partial index — at most one active covenant per scope at a time
DO $idx$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes WHERE indexname = 'covenants_active_scope'
    ) THEN
        CREATE UNIQUE INDEX covenants_active_scope
            ON stewards.covenants (scope) WHERE deactivated_at IS NULL;
    END IF;
END;
$idx$;

-- ---------------------------------------------------------------------
-- Section 3: work_items.intent_id FK
-- ---------------------------------------------------------------------

ALTER TABLE stewards.work_items
    ADD COLUMN IF NOT EXISTS intent_id uuid REFERENCES stewards.intents(id);

CREATE INDEX IF NOT EXISTS work_items_intent_id ON stewards.work_items (intent_id);

COMMENT ON COLUMN stewards.work_items.intent_id IS
'Phase 5d (C.1): FK to stewards.intents. Nullable in C.1; NOT NULL added in C.5 after backfill.';
