-- +goose Up
-- Per-exercise score history for adaptive difficulty
CREATE TABLE memorize_scores (
    id          SERIAL PRIMARY KEY,
    practice_id INTEGER NOT NULL REFERENCES practices(id) ON DELETE CASCADE,
    user_id     INTEGER NOT NULL REFERENCES users(id),
    mode        TEXT NOT NULL,       -- Forward: 'reveal_whole', 'reveal_words', 'type_words', 'arrange', 'type_full'
                                     -- Reverse: 'reverse_full', 'reverse_partial', 'reverse_fragment'
    score       REAL NOT NULL,       -- 0.0 to 1.0 (accuracy)
    quality     INTEGER,             -- SM-2 quality 0-5 (user's self-rating)
    duration_s  INTEGER,             -- how long the exercise took
    date        DATE NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_memorize_scores_practice ON memorize_scores(practice_id);
CREATE INDEX idx_memorize_scores_user_date ON memorize_scores(user_id, date);

-- Per-card per-mode aptitude cache (denormalized, recalculated on each score)
CREATE TABLE memorize_aptitude (
    id            SERIAL PRIMARY KEY,
    practice_id   INTEGER NOT NULL REFERENCES practices(id) ON DELETE CASCADE,
    user_id       INTEGER NOT NULL REFERENCES users(id),
    mode          TEXT NOT NULL,
    aptitude      REAL NOT NULL DEFAULT 0.0,  -- rolling average, 0.0 to 1.0
    sample_count  INTEGER DEFAULT 0,
    last_score_at TIMESTAMPTZ,
    UNIQUE(practice_id, user_id, mode)
);

-- Overall card difficulty level (what level the algorithm currently targets)
ALTER TABLE practices ADD COLUMN memorize_level INTEGER DEFAULT 1;

-- +goose Down
DROP TABLE IF EXISTS memorize_scores;
DROP TABLE IF EXISTS memorize_aptitude;
ALTER TABLE practices DROP COLUMN IF EXISTS memorize_level;
