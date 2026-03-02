-- +goose Up
CREATE TABLE brain_messages (
    id           BIGSERIAL PRIMARY KEY,
    message_id   TEXT NOT NULL UNIQUE,
    user_id      BIGINT NOT NULL REFERENCES users(id),
    direction    TEXT NOT NULL,
    payload      TEXT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'pending',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    delivered_at TIMESTAMPTZ
);

CREATE INDEX idx_brain_messages_pending ON brain_messages(user_id, status, direction);
CREATE INDEX idx_brain_messages_user_created ON brain_messages(user_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS brain_messages;
