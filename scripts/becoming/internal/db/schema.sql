-- Becoming App Schema
-- Generalized practice tracking: memorization, exercises, habits, tasks

-- Practices: anything you do repeatedly and want to track
-- Types: memorize, tracker, habit, task
CREATE TABLE IF NOT EXISTS practices (
    id          INTEGER PRIMARY KEY,
    name        TEXT NOT NULL,           -- "D&C 93:29" or "Clamshell" or "Morning prayer"
    description TEXT,                    -- Full verse text, exercise instructions, etc.
    type        TEXT NOT NULL,           -- memorize | exercise | habit | task
    category    TEXT,                    -- "scripture", "pt", "spiritual", "fitness", custom grouping
    source_doc  TEXT,                    -- link to study doc that generated this
    source_path TEXT,                    -- path to source file (scripture, talk, etc.)

    -- Type-specific config stored as JSON
    -- memorize: {"ease_factor": 2.5, "interval": 1, "repetitions": 0, "target_daily_reps": 1}
    -- tracker:  {"target_sets": 2, "target_reps": 15, "unit": "reps"}
    -- habit:    {"frequency": "daily"}
    -- task:     {"due_date": "2026-03-01"}
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

    -- Flexible value fields — meaning depends on practice type
    -- memorize: quality=0-5 (SM-2), value=null, sets/reps=null
    -- tracker:  quality=null, value=null, sets=2, reps=15
    -- habit:    quality=null, value="25 min", sets/reps=null
    -- task:     quality=null, value="completed milestone X"
    quality     INTEGER,                 -- SM-2 quality rating (0-5)
    value       TEXT,                    -- freeform value ("25 min", "3 miles")
    sets        INTEGER,                 -- number of sets completed
    reps        INTEGER,                 -- reps per set (or total)
    duration_s  INTEGER,                 -- duration in seconds
    notes       TEXT,

    -- Spaced repetition: next review date (updated after each memorize log)
    -- Stored here for historical tracking; current schedule lives on practice.config
    next_review DATE
);

-- Tasks/commitments from study documents (separate from practices for clarity)
CREATE TABLE IF NOT EXISTS tasks (
    id          INTEGER PRIMARY KEY,
    title       TEXT NOT NULL,
    description TEXT,
    source_doc  TEXT,                    -- e.g., "study/truth-atonement.md"
    source_section TEXT,                 -- e.g., "Become"
    scripture   TEXT,                    -- e.g., "D&C 93:29"
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
