-- Becoming App Schema
-- Generalized practice tracking: memorization, exercises, habits, tasks

-- Practices: anything you do repeatedly and want to track
-- Types: memorize, tracker, habit, task
CREATE TABLE IF NOT EXISTS practices (
    id          INTEGER PRIMARY KEY,
    user_id     INTEGER NOT NULL DEFAULT 1 REFERENCES users(id),
    name        TEXT NOT NULL,           -- "D&C 93:29" or "Clamshell" or "Morning prayer"
    description TEXT,                    -- Full verse text, exercise instructions, etc.
    type        TEXT NOT NULL,           -- memorize | exercise | habit | task
    category    TEXT,                    -- "scripture", "pt", "spiritual", "fitness", custom grouping
    source_doc  TEXT,                    -- link to study doc that generated this
    source_path TEXT,                    -- path to source file (scripture, talk, etc.)

    -- Type-specific config stored as JSON
    config      TEXT DEFAULT '{}',

    sort_order  INTEGER DEFAULT 0,
    active      BOOLEAN DEFAULT 1,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME               -- for tasks that finish
);

-- Practice logs: each time you do a practice
CREATE TABLE IF NOT EXISTS practice_logs (
    id          INTEGER PRIMARY KEY,
    practice_id INTEGER NOT NULL REFERENCES practices(id) ON DELETE CASCADE,
    logged_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    date        DATE NOT NULL,           -- the calendar date (for grouping)

    quality     INTEGER,                 -- SM-2 quality rating (0-5)
    value       TEXT,                    -- freeform value ("25 min", "3 miles")
    sets        INTEGER,                 -- number of sets completed
    reps        INTEGER,                 -- reps per set (or total)
    duration_s  INTEGER,                 -- duration in seconds
    notes       TEXT,

    next_review DATE
);

-- Tasks/commitments from study documents (separate from practices for clarity)
CREATE TABLE IF NOT EXISTS tasks (
    id          INTEGER PRIMARY KEY,
    user_id     INTEGER NOT NULL DEFAULT 1 REFERENCES users(id),
    title       TEXT NOT NULL,
    description TEXT,
    source_doc  TEXT,
    source_section TEXT,
    scripture   TEXT,
    type        TEXT NOT NULL DEFAULT 'ongoing', -- once | daily | weekly | ongoing
    status      TEXT NOT NULL DEFAULT 'active',  -- active | completed | paused | archived
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_practices_type ON practices(type);
CREATE INDEX IF NOT EXISTS idx_practices_active ON practices(active);
CREATE INDEX IF NOT EXISTS idx_practice_logs_practice ON practice_logs(practice_id);
CREATE INDEX IF NOT EXISTS idx_practice_logs_date ON practice_logs(date);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);

-- Notes: lightweight notes optionally linked to a practice, task, or pillar
CREATE TABLE IF NOT EXISTS notes (
    id          INTEGER PRIMARY KEY,
    user_id     INTEGER NOT NULL DEFAULT 1 REFERENCES users(id),
    content     TEXT NOT NULL,
    practice_id INTEGER REFERENCES practices(id) ON DELETE SET NULL,
    task_id     INTEGER REFERENCES tasks(id) ON DELETE SET NULL,
    pillar_id   INTEGER,
    pinned      BOOLEAN DEFAULT 0,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notes_practice ON notes(practice_id);
CREATE INDEX IF NOT EXISTS idx_notes_task ON notes(task_id);
CREATE INDEX IF NOT EXISTS idx_notes_pinned ON notes(pinned);

-- Prompts: reflection questions (DB-stored, user-editable)
CREATE TABLE IF NOT EXISTS prompts (
    id          INTEGER PRIMARY KEY,
    user_id     INTEGER NOT NULL DEFAULT 1 REFERENCES users(id),
    text        TEXT NOT NULL,
    active      BOOLEAN DEFAULT 1,
    sort_order  INTEGER DEFAULT 0,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Reflections: one per user per day, inward-facing daily journal
CREATE TABLE IF NOT EXISTS reflections (
    id          INTEGER PRIMARY KEY,
    user_id     INTEGER NOT NULL DEFAULT 1 REFERENCES users(id),
    date        DATE NOT NULL,
    prompt_id   INTEGER REFERENCES prompts(id) ON DELETE SET NULL,
    prompt_text TEXT,
    content     TEXT NOT NULL,
    mood        INTEGER,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, date)
);

CREATE INDEX IF NOT EXISTS idx_reflections_date ON reflections(date);

-- Pillars: growth areas (vision layer)
CREATE TABLE IF NOT EXISTS pillars (
    id          INTEGER PRIMARY KEY,
    user_id     INTEGER NOT NULL DEFAULT 1 REFERENCES users(id),
    name        TEXT NOT NULL,
    description TEXT,
    icon        TEXT,
    parent_id   INTEGER REFERENCES pillars(id) ON DELETE CASCADE,
    sort_order  INTEGER DEFAULT 0,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Junction: practices ↔ pillars (many-to-many)
CREATE TABLE IF NOT EXISTS practice_pillars (
    practice_id INTEGER NOT NULL REFERENCES practices(id) ON DELETE CASCADE,
    pillar_id   INTEGER NOT NULL REFERENCES pillars(id) ON DELETE CASCADE,
    PRIMARY KEY (practice_id, pillar_id)
);

-- Junction: tasks ↔ pillars (many-to-many)
CREATE TABLE IF NOT EXISTS task_pillars (
    task_id     INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    pillar_id   INTEGER NOT NULL REFERENCES pillars(id) ON DELETE CASCADE,
    PRIMARY KEY (task_id, pillar_id)
);

CREATE INDEX IF NOT EXISTS idx_pillars_parent ON pillars(parent_id);
CREATE INDEX IF NOT EXISTS idx_practice_pillars_practice ON practice_pillars(practice_id);
CREATE INDEX IF NOT EXISTS idx_practice_pillars_pillar ON practice_pillars(pillar_id);
CREATE INDEX IF NOT EXISTS idx_task_pillars_task ON task_pillars(task_id);
CREATE INDEX IF NOT EXISTS idx_task_pillars_pillar ON task_pillars(pillar_id);
