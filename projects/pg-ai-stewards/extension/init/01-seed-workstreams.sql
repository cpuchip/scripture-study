-- 01-seed-workstreams.sql
-- First-boot seed: create the WS1-WS9 workstream vertices.
--
-- Separated from 2-6a-workstreams.sql at v0.2.0 because the seed
-- calls stewards.import_workstream() -> AGE cypher() which requires
-- ag_catalog on search_path. Running it during CREATE EXTENSION
-- (which is what extension_sql_file! does) corrupted the install
-- transaction's search_path and pushed every later pgrx-emitted
-- pg_extern (version, enqueue, providers_loaded, ...) into ag_catalog
-- instead of stewards. See the head comment of 2-6a-workstreams.sql.
--
-- AGE has been LOADed by 00-extensions.sql by the time this runs.
-- Read from .mind/workstreams.md; do not invent.

LOAD 'age';
SET search_path TO ag_catalog, "$user", public;

SELECT stewards.import_workstream('WS1', 'Brain Core',
    'Pipeline, steward, commissions, classifier, retry/escalation, model selection, data safety',
    'active');
SELECT stewards.import_workstream('WS2', 'Brain UX',
    'UI panels, dialogs, kanban, file viewer, inline panel, Windows service/systray',
    'active');
SELECT stewards.import_workstream('WS3', 'Gospel Engine',
    'engine.ibeco.me, gospel-engine MCP, search/index, graph, hosted backend',
    'active');
SELECT stewards.import_workstream('WS4', 'study.ibeco.me',
    'Web UI for studies, notes, reader, public study pages',
    'active');
SELECT stewards.import_workstream('WS5', 'Memory & Process',
    '.mind/, agents, skills, voice/bias, cleanup passes, tokenomics, brain<->VS Code bridge, debug agent, Claude Code integration, Sabbath agent, pg-ai-stewards',
    'active');
SELECT stewards.import_workstream('WS6', 'Studies',
    'Scripture study output (study/, becoming/)',
    'active');
SELECT stewards.import_workstream('WS7', 'Teaching',
    'YouTube content arc, talks, public-facing teaching',
    'active');
SELECT stewards.import_workstream('WS8', 'Sunday School',
    'Calling — lesson prep, ward council',
    'active');
SELECT stewards.import_workstream('WS9', 'Other Apps',
    'Budget app, cpuchip.net rebuild, Space Center',
    'active');

SELECT 'workstreams seeded: ' || count(*)::text AS ok
    FROM stewards.workstreams;
