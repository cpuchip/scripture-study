-- +goose Up
ALTER TABLE tasks ADD COLUMN brain_entry_id TEXT;

-- +goose Down
ALTER TABLE tasks DROP COLUMN brain_entry_id;
