-- +goose Up
CREATE TABLE shared_links (
    id          SERIAL PRIMARY KEY,
    code        TEXT NOT NULL UNIQUE,
    user_id     INTEGER REFERENCES users(id),
    source_id   INTEGER REFERENCES document_sources(id) ON DELETE SET NULL,
    provider    TEXT NOT NULL DEFAULT 'gh',
    repo        TEXT NOT NULL,
    branch      TEXT NOT NULL DEFAULT 'main',
    doc_filter  TEXT NOT NULL DEFAULT '**/*.md',
    file_path   TEXT,
    hits        INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_shared_links_code ON shared_links(code);

-- +goose Down
DROP TABLE IF EXISTS shared_links;
