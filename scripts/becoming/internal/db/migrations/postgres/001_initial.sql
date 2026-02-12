-- Becoming App — PostgreSQL Schema
-- Full initial schema (equivalent to SQLite schema.sql + auth_schema.sql)
-- +goose Up

-- Users (identity)
CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL DEFAULT '',
    name          TEXT NOT NULL DEFAULT '',
    avatar_url    TEXT NOT NULL DEFAULT '',
    provider      TEXT NOT NULL DEFAULT 'email',
    provider_id   TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    last_login    TIMESTAMPTZ DEFAULT NOW()
);

-- Sessions (browser auth via HttpOnly cookie)
CREATE TABLE IF NOT EXISTS sessions (
    id          TEXT PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    expires_at  TIMESTAMPTZ NOT NULL,
    last_active TIMESTAMPTZ DEFAULT NOW(),
    user_agent  TEXT NOT NULL DEFAULT '',
    ip_address  TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

-- API tokens (programmatic auth via Bearer header)
CREATE TABLE IF NOT EXISTS api_tokens (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL DEFAULT '',
    token_hash  TEXT NOT NULL,
    prefix      TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    last_used   TIMESTAMPTZ,
    expires_at  TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_api_tokens_user ON api_tokens(user_id);

-- Practices: anything you do repeatedly and want to track
CREATE TABLE IF NOT EXISTS practices (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL REFERENCES users(id),
    name         TEXT NOT NULL,
    description  TEXT,
    type         TEXT NOT NULL,
    category     TEXT,
    source_doc   TEXT,
    source_path  TEXT,
    config       TEXT DEFAULT '{}',
    sort_order   INTEGER DEFAULT 0,
    active       BOOLEAN DEFAULT TRUE,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_practices_type ON practices(type);
CREATE INDEX IF NOT EXISTS idx_practices_active ON practices(active);
CREATE INDEX IF NOT EXISTS idx_practices_user ON practices(user_id);

-- Practice logs: each time you do a practice
CREATE TABLE IF NOT EXISTS practice_logs (
    id          BIGSERIAL PRIMARY KEY,
    practice_id BIGINT NOT NULL REFERENCES practices(id) ON DELETE CASCADE,
    logged_at   TIMESTAMPTZ DEFAULT NOW(),
    date        DATE NOT NULL,
    quality     INTEGER,
    value       TEXT,
    sets        INTEGER,
    reps        INTEGER,
    duration_s  INTEGER,
    notes       TEXT,
    next_review DATE
);
CREATE INDEX IF NOT EXISTS idx_practice_logs_practice ON practice_logs(practice_id);
CREATE INDEX IF NOT EXISTS idx_practice_logs_date ON practice_logs(date);

-- Tasks/commitments
CREATE TABLE IF NOT EXISTS tasks (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id),
    title           TEXT NOT NULL,
    description     TEXT,
    source_doc      TEXT,
    source_section  TEXT,
    scripture       TEXT,
    type            TEXT NOT NULL DEFAULT 'ongoing',
    status          TEXT NOT NULL DEFAULT 'active',
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    completed_at    TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_user ON tasks(user_id);

-- Notes
CREATE TABLE IF NOT EXISTS notes (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id),
    content     TEXT NOT NULL,
    practice_id BIGINT REFERENCES practices(id) ON DELETE SET NULL,
    task_id     BIGINT REFERENCES tasks(id) ON DELETE SET NULL,
    pillar_id   BIGINT,
    pinned      BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_notes_practice ON notes(practice_id);
CREATE INDEX IF NOT EXISTS idx_notes_task ON notes(task_id);
CREATE INDEX IF NOT EXISTS idx_notes_pinned ON notes(pinned);
CREATE INDEX IF NOT EXISTS idx_notes_user ON notes(user_id);

-- Prompts: reflection questions
CREATE TABLE IF NOT EXISTS prompts (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id),
    text       TEXT NOT NULL,
    active     BOOLEAN DEFAULT TRUE,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_prompts_user ON prompts(user_id);

-- Reflections: one per user per day
CREATE TABLE IF NOT EXISTS reflections (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id),
    date        DATE NOT NULL,
    prompt_id   BIGINT REFERENCES prompts(id) ON DELETE SET NULL,
    prompt_text TEXT,
    content     TEXT NOT NULL,
    mood        INTEGER,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, date)
);
CREATE INDEX IF NOT EXISTS idx_reflections_date ON reflections(date);
CREATE INDEX IF NOT EXISTS idx_reflections_user ON reflections(user_id);

-- Pillars: growth areas (vision layer)
CREATE TABLE IF NOT EXISTS pillars (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id),
    name        TEXT NOT NULL,
    description TEXT,
    icon        TEXT,
    parent_id   BIGINT REFERENCES pillars(id) ON DELETE CASCADE,
    sort_order  INTEGER DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_pillars_parent ON pillars(parent_id);
CREATE INDEX IF NOT EXISTS idx_pillars_user ON pillars(user_id);

-- Junction: practices ↔ pillars (many-to-many)
CREATE TABLE IF NOT EXISTS practice_pillars (
    practice_id BIGINT NOT NULL REFERENCES practices(id) ON DELETE CASCADE,
    pillar_id   BIGINT NOT NULL REFERENCES pillars(id) ON DELETE CASCADE,
    PRIMARY KEY (practice_id, pillar_id)
);
CREATE INDEX IF NOT EXISTS idx_practice_pillars_practice ON practice_pillars(practice_id);
CREATE INDEX IF NOT EXISTS idx_practice_pillars_pillar ON practice_pillars(pillar_id);

-- Junction: tasks ↔ pillars (many-to-many)
CREATE TABLE IF NOT EXISTS task_pillars (
    task_id   BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    pillar_id BIGINT NOT NULL REFERENCES pillars(id) ON DELETE CASCADE,
    PRIMARY KEY (task_id, pillar_id)
);
CREATE INDEX IF NOT EXISTS idx_task_pillars_task ON task_pillars(task_id);
CREATE INDEX IF NOT EXISTS idx_task_pillars_pillar ON task_pillars(pillar_id);

-- +goose Down
DROP TABLE IF EXISTS task_pillars;
DROP TABLE IF EXISTS practice_pillars;
DROP TABLE IF EXISTS pillars;
DROP TABLE IF EXISTS reflections;
DROP TABLE IF EXISTS prompts;
DROP TABLE IF EXISTS notes;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS practice_logs;
DROP TABLE IF EXISTS practices;
DROP TABLE IF EXISTS api_tokens;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
