-- +goose Up
ALTER TABLE brain_entries ADD COLUMN subtasks TEXT NOT NULL DEFAULT '[]';

-- +goose Down
ALTER TABLE brain_entries DROP COLUMN subtasks;
