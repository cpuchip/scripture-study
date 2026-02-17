-- +goose Up
-- Add start_date column to practices for deferred practice visibility.
-- Practices with start_date > today are hidden from the daily view until that date.

ALTER TABLE practices ADD COLUMN start_date DATE;

-- Backfill: set start_date to the date portion of created_at
UPDATE practices SET start_date = created_at::date;

-- +goose Down
ALTER TABLE practices DROP COLUMN IF EXISTS start_date;
