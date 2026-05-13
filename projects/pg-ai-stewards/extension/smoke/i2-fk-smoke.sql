\set ON_ERROR_STOP off

-- Pick a stable work_item row to UPDATE-test against
\echo '--- test 1: UPDATE to bogus project_association must fail ---'
BEGIN;
UPDATE stewards.work_items
   SET project_association = 'no-such-project-xyz'
 WHERE id = (SELECT id FROM stewards.work_items ORDER BY created_at DESC LIMIT 1);
ROLLBACK;

\echo '--- test 2: UPDATE to NULL must succeed ---'
BEGIN;
UPDATE stewards.work_items
   SET project_association = NULL
 WHERE id = (SELECT id FROM stewards.work_items ORDER BY created_at DESC LIMIT 1)
RETURNING id, project_association;
ROLLBACK;

\echo '--- test 3: UPDATE to known slug must succeed ---'
BEGIN;
UPDATE stewards.work_items
   SET project_association = 'pg-ai-stewards'
 WHERE id = (SELECT id FROM stewards.work_items ORDER BY created_at DESC LIMIT 1)
RETURNING id, project_association;
ROLLBACK;

\echo '--- test 4: DELETE project with work_items must be blocked ---'
BEGIN;
DELETE FROM stewards.projects WHERE slug = 'pg-ai-stewards';
ROLLBACK;

\echo '--- test 5: ON UPDATE CASCADE — rename slug propagates ---'
BEGIN;
-- temp project nobody references
INSERT INTO stewards.projects (slug, name) VALUES ('cascade-smoke-from', 'cascade smoke');
UPDATE stewards.work_items
   SET project_association = 'cascade-smoke-from'
 WHERE id = (SELECT id FROM stewards.work_items ORDER BY created_at DESC LIMIT 1);
UPDATE stewards.projects SET slug = 'cascade-smoke-to' WHERE slug = 'cascade-smoke-from';
SELECT 'cascade result:' AS check_,
       (SELECT project_association FROM stewards.work_items
        WHERE id = (SELECT id FROM stewards.work_items ORDER BY created_at DESC LIMIT 1)) AS wi_assoc;
ROLLBACK;
