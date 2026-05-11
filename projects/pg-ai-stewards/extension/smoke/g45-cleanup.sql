-- Cleanup synthetic rows from G.4.5 smoke
BEGIN;

DELETE FROM stewards.pending_file_writes
 WHERE source_kind IN ('lesson','resolution')
   AND (target_path = '.mind/principles.md' OR target_path = 'study/smoke-g45.md')
   AND requested_by IN ('lesson_promote','council_resolve');

DELETE FROM stewards.lessons
 WHERE ratified_by = 'smoke-test';

-- Resolution side: break the circular FK then cascade
UPDATE stewards.councils
   SET resolution_id = NULL
 WHERE resolution_id IN (SELECT id FROM stewards.resolutions WHERE resolved_by = 'smoke-bishop');

DELETE FROM stewards.resolutions
 WHERE resolved_by = 'smoke-bishop';

DELETE FROM stewards.councils
 WHERE convened_by = 'smoke-test';

DELETE FROM stewards.work_items
 WHERE slug = 'smoke-g45';

SELECT
  (SELECT count(*) FROM stewards.lessons WHERE ratified_by = 'smoke-test') AS leftover_lessons,
  (SELECT count(*) FROM stewards.resolutions WHERE resolved_by = 'smoke-bishop') AS leftover_resolutions,
  (SELECT count(*) FROM stewards.councils WHERE convened_by = 'smoke-test') AS leftover_councils,
  (SELECT count(*) FROM stewards.work_items WHERE slug = 'smoke-g45') AS leftover_work_items,
  (SELECT count(*) FROM stewards.pending_file_writes WHERE requested_by IN ('lesson_promote','council_resolve')) AS leftover_pw;

COMMIT;
