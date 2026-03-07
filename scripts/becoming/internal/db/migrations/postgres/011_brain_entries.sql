-- +goose Up
CREATE TABLE brain_entries (
    id          TEXT NOT NULL,
    user_id     BIGINT NOT NULL REFERENCES users(id),
    title       TEXT NOT NULL,
    category    TEXT NOT NULL,
    body        TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT '',
    action_done BOOLEAN NOT NULL DEFAULT FALSE,
    due_date    TEXT NOT NULL DEFAULT '',
    next_action TEXT NOT NULL DEFAULT '',
    tags        TEXT NOT NULL DEFAULT '[]',
    source      TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    synced_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, user_id)
);

CREATE INDEX idx_brain_entries_user ON brain_entries(user_id);
CREATE INDEX idx_brain_entries_category ON brain_entries(user_id, category);

-- +goose Down
DROP TABLE IF EXISTS brain_entries;
