-- +goose Up
-- Practice lifecycle: add status, archived_at, and end_date columns.
-- Replaces the boolean "active" field with a proper status enum.

ALTER TABLE practices ADD COLUMN status TEXT NOT NULL DEFAULT 'active';
ALTER TABLE practices ADD COLUMN archived_at TIMESTAMPTZ;
ALTER TABLE practices ADD COLUMN end_date DATE;

-- Backfill status from existing active/completed_at columns
UPDATE practices SET status = CASE
    WHEN completed_at IS NOT NULL THEN 'completed'
    WHEN active = FALSE THEN 'paused'
    ELSE 'active'
END;

CREATE INDEX IF NOT EXISTS idx_practices_status ON practices(status);

-- +goose Down
DROP INDEX IF EXISTS idx_practices_status;
ALTER TABLE practices DROP COLUMN IF EXISTS end_date;
ALTER TABLE practices DROP COLUMN IF EXISTS archived_at;
ALTER TABLE practices DROP COLUMN IF EXISTS status;
