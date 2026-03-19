-- +goose Up
ALTER TABLE user_settings ADD COLUMN notify_practices_by_default BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE user_settings ADD COLUMN quiet_hours_start TEXT;
ALTER TABLE user_settings ADD COLUMN quiet_hours_end TEXT;
ALTER TABLE user_settings ADD COLUMN default_timing TEXT NOT NULL DEFAULT 'at_time';

-- +goose Down
ALTER TABLE user_settings DROP COLUMN IF EXISTS notify_practices_by_default;
ALTER TABLE user_settings DROP COLUMN IF EXISTS quiet_hours_start;
ALTER TABLE user_settings DROP COLUMN IF EXISTS quiet_hours_end;
ALTER TABLE user_settings DROP COLUMN IF EXISTS default_timing;
