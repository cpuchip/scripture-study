-- +goose Up
-- Document sources for the Study Reader: git-backed document libraries.

CREATE TABLE document_sources (
    id              SERIAL PRIMARY KEY,
    user_id         INTEGER NOT NULL REFERENCES users(id),
    name            TEXT NOT NULL,
    source_type     TEXT NOT NULL,
    repo            TEXT NOT NULL,
    branch          TEXT DEFAULT 'main',
    include_paths   TEXT DEFAULT '[]',
    exclude_paths   TEXT DEFAULT '[]',
    tree_cache      TEXT,
    tree_etag       TEXT,
    tree_cached_at  TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE reading_progress (
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(id),
    source_id   INTEGER NOT NULL REFERENCES document_sources(id) ON DELETE CASCADE,
    file_path   TEXT NOT NULL,
    read_at     TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    scroll_pct  REAL DEFAULT 0,
    UNIQUE(user_id, source_id, file_path)
);

-- +goose Down
DROP TABLE IF EXISTS reading_progress;
DROP TABLE IF EXISTS document_sources;
