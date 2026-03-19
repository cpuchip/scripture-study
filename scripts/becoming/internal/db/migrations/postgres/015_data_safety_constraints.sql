-- +goose Up

-- practices: enforce NOT NULL on active, and valid status/type values
ALTER TABLE practices ALTER COLUMN active SET NOT NULL;
ALTER TABLE practices ADD CONSTRAINT practices_status_check
    CHECK (status IN ('active', 'paused', 'completed', 'archived'));
ALTER TABLE practices ADD CONSTRAINT practices_type_check
    CHECK (type IN ('memorize', 'tracker', 'habit', 'task', 'scheduled'));

-- tasks: enforce valid status values (type is loosely defined, no constraint)
ALTER TABLE tasks ADD CONSTRAINT tasks_status_check
    CHECK (status IN ('active', 'completed', 'deferred', 'paused', 'archived'));

-- practice_logs: enforce quality range (0-5, SM-2 scale)
ALTER TABLE practice_logs ADD CONSTRAINT practice_logs_quality_check
    CHECK (quality IS NULL OR (quality >= 0 AND quality <= 5));

-- +goose Down
ALTER TABLE practices ALTER COLUMN active DROP NOT NULL;
ALTER TABLE practices DROP CONSTRAINT IF EXISTS practices_status_check;
ALTER TABLE practices DROP CONSTRAINT IF EXISTS practices_type_check;
ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_status_check;
ALTER TABLE practice_logs DROP CONSTRAINT IF EXISTS practice_logs_quality_check;
