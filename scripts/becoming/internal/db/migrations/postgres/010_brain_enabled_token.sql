-- +goose Up
ALTER TABLE api_tokens ADD COLUMN brain_enabled BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE api_tokens DROP COLUMN brain_enabled;
