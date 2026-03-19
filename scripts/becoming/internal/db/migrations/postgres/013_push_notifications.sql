-- +goose Up
CREATE TABLE push_subscriptions (
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    endpoint    TEXT NOT NULL,
    keys_p256dh TEXT NOT NULL,
    keys_auth   TEXT NOT NULL,
    user_agent  TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, endpoint)
);

CREATE TABLE notification_log (
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    practice_id INTEGER NOT NULL REFERENCES practices(id) ON DELETE CASCADE,
    date        DATE NOT NULL,
    sent_at     TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_notification_log_user_date ON notification_log(user_id, date);

CREATE TABLE user_settings (
    user_id                INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    notifications_enabled  BOOLEAN NOT NULL DEFAULT FALSE
);

-- +goose Down
DROP TABLE IF EXISTS user_settings;
DROP TABLE IF EXISTS notification_log;
DROP TABLE IF EXISTS push_subscriptions;
