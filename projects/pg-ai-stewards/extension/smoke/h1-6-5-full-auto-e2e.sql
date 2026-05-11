-- H.1.6.5: FULL AUTO e2e — proves the three substrate completion
-- patches close the loop. NEW binding question (scenario #3) to test
-- on different domain content.

-- Enable auto-materialize on research-write so the trigger fires it
UPDATE stewards.pipelines
   SET auto_materialize_on_verified = true
 WHERE family = 'research-write';

DO $$
DECLARE
    v_wi uuid;
    v_intent_id uuid;
    v_work_id bigint;
BEGIN
    SELECT id INTO v_intent_id FROM stewards.intents WHERE slug = 'general-research';

    SELECT stewards.work_item_create(
        'research-write',
        '{"binding_question":"What is the state of Postgres extension distribution in 2026? How do mature extensions (pgvector, paradedb / pg_search, Citus, TimescaleDB) ship, version, and handle upgrades? Focus on packaging, compatibility, and what an extension author should know in 2026."}'::jsonb,
        'pg-ext-distribution-2026-05-11',
        'human',
        NULL,
        v_intent_id
    ) INTO v_wi;

    -- Cost cap + opt-in file destination. auto_materialize inherits from pipeline.
    UPDATE stewards.work_items
       SET file_destination = 'research/pg-ext-distribution-2026-05-11.md',
           cost_cap_micro   = 400000
     WHERE id = v_wi;

    v_work_id := stewards.work_item_dispatch_stage(v_wi);
    RAISE NOTICE 'H.1.6.5 full-auto e2e: work_item=% gather work_id=% cap=$0.40 auto_mat=ON',
        v_wi, v_work_id;
END
$$;

SELECT id, slug, current_stage, status, maturity, file_destination,
       auto_materialize_enabled AS wi_override,
       (SELECT auto_materialize_on_verified FROM stewards.pipelines WHERE family='research-write') AS pipeline_default
  FROM stewards.work_items
 WHERE slug='pg-ext-distribution-2026-05-11';
