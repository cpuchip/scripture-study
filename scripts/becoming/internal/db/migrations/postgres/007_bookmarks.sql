-- +goose Up
CREATE TABLE bookmarks (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id),
    source_id   BIGINT NOT NULL REFERENCES document_sources(id),
    file_path   TEXT NOT NULL,
    anchor      TEXT NOT NULL DEFAULT '',
    excerpt     TEXT NOT NULL DEFAULT '',
    note        TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bookmarks_user_id ON bookmarks(user_id);
CREATE INDEX idx_bookmarks_user_source ON bookmarks(user_id, source_id);

-- +goose Down
DROP TABLE IF EXISTS bookmarks;
