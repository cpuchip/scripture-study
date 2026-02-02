-- Gospel MCP Server Database Schema
-- SQLite with FTS5 for full-text search

-- ============================================================================
-- SCRIPTURES
-- ============================================================================

-- Individual verses with full structural metadata
CREATE TABLE IF NOT EXISTS scriptures (
    id INTEGER PRIMARY KEY,
    volume TEXT NOT NULL,        -- 'ot', 'nt', 'bofm', 'dc-testament', 'pgp'
    book TEXT NOT NULL,          -- 'gen', 'matt', '1-ne', 'dc', 'moses'
    chapter INTEGER NOT NULL,
    verse INTEGER NOT NULL,
    text TEXT NOT NULL,          -- Verse text (markdown)
    file_path TEXT NOT NULL,     -- Relative path: 'gospel-library/eng/scriptures/ot/gen/1.md'
    source_url TEXT NOT NULL,    -- https://www.churchofjesuschrist.org/study/scriptures/ot/gen/1?lang=eng
    
    UNIQUE(volume, book, chapter, verse)
);

CREATE INDEX IF NOT EXISTS idx_scriptures_volume ON scriptures(volume);
CREATE INDEX IF NOT EXISTS idx_scriptures_book ON scriptures(volume, book);
CREATE INDEX IF NOT EXISTS idx_scriptures_chapter ON scriptures(volume, book, chapter);
CREATE INDEX IF NOT EXISTS idx_scriptures_file ON scriptures(file_path);

-- Full-text search on verse content
CREATE VIRTUAL TABLE IF NOT EXISTS scriptures_fts USING fts5(
    text,
    content='scriptures',
    content_rowid='id'
);

-- Triggers to keep FTS index in sync
CREATE TRIGGER IF NOT EXISTS scriptures_ai AFTER INSERT ON scriptures BEGIN
    INSERT INTO scriptures_fts(rowid, text) VALUES (new.id, new.text);
END;

CREATE TRIGGER IF NOT EXISTS scriptures_ad AFTER DELETE ON scriptures BEGIN
    INSERT INTO scriptures_fts(scriptures_fts, rowid, text) VALUES('delete', old.id, old.text);
END;

CREATE TRIGGER IF NOT EXISTS scriptures_au AFTER UPDATE ON scriptures BEGIN
    INSERT INTO scriptures_fts(scriptures_fts, rowid, text) VALUES('delete', old.id, old.text);
    INSERT INTO scriptures_fts(rowid, text) VALUES (new.id, new.text);
END;

-- ============================================================================
-- CHAPTERS
-- ============================================================================

-- Chapter-level content for context retrieval
CREATE TABLE IF NOT EXISTS chapters (
    id INTEGER PRIMARY KEY,
    volume TEXT NOT NULL,
    book TEXT NOT NULL,
    chapter INTEGER NOT NULL,
    title TEXT,                  -- Chapter heading/summary if available
    full_content TEXT NOT NULL,  -- Full chapter markdown
    file_path TEXT NOT NULL,
    source_url TEXT,             -- Chapter URL for linking
    
    UNIQUE(volume, book, chapter)
);

CREATE INDEX IF NOT EXISTS idx_chapters_volume ON chapters(volume);
CREATE INDEX IF NOT EXISTS idx_chapters_book ON chapters(volume, book);
CREATE INDEX IF NOT EXISTS idx_chapters_file ON chapters(file_path);

-- ============================================================================
-- TALKS
-- ============================================================================

-- General conference talks and other addresses
CREATE TABLE IF NOT EXISTS talks (
    id INTEGER PRIMARY KEY,
    year INTEGER NOT NULL,
    month INTEGER NOT NULL,      -- 4 or 10
    session TEXT,                -- 'saturday-morning', 'priesthood', etc.
    speaker TEXT NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,       -- Full talk markdown
    file_path TEXT NOT NULL,
    source_url TEXT NOT NULL,    -- https://www.churchofjesuschrist.org/study/...
    
    UNIQUE(file_path)
);

CREATE INDEX IF NOT EXISTS idx_talks_year ON talks(year);
CREATE INDEX IF NOT EXISTS idx_talks_year_month ON talks(year, month);
CREATE INDEX IF NOT EXISTS idx_talks_speaker ON talks(speaker);

-- Full-text search on talks
CREATE VIRTUAL TABLE IF NOT EXISTS talks_fts USING fts5(
    title,
    speaker,
    content,
    content='talks',
    content_rowid='id'
);

-- Triggers to keep FTS index in sync
CREATE TRIGGER IF NOT EXISTS talks_ai AFTER INSERT ON talks BEGIN
    INSERT INTO talks_fts(rowid, title, speaker, content) VALUES (new.id, new.title, new.speaker, new.content);
END;

CREATE TRIGGER IF NOT EXISTS talks_ad AFTER DELETE ON talks BEGIN
    INSERT INTO talks_fts(talks_fts, rowid, title, speaker, content) VALUES('delete', old.id, old.title, old.speaker, old.content);
END;

CREATE TRIGGER IF NOT EXISTS talks_au AFTER UPDATE ON talks BEGIN
    INSERT INTO talks_fts(talks_fts, rowid, title, speaker, content) VALUES('delete', old.id, old.title, old.speaker, old.content);
    INSERT INTO talks_fts(rowid, title, speaker, content) VALUES (new.id, new.title, new.speaker, new.content);
END;

-- ============================================================================
-- MANUALS
-- ============================================================================

-- Come Follow Me, handbooks, teaching guides, magazines, etc.
CREATE TABLE IF NOT EXISTS manuals (
    id INTEGER PRIMARY KEY,
    content_type TEXT NOT NULL,  -- 'manual', 'magazine', 'handbook'
    collection_id TEXT NOT NULL, -- 'come-follow-me-for-home-and-church-old-testament-2026'
    section TEXT,                -- Lesson number or section identifier
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    file_path TEXT NOT NULL,
    source_url TEXT NOT NULL,    -- https://www.churchofjesuschrist.org/study/...
    
    UNIQUE(file_path)
);

CREATE INDEX IF NOT EXISTS idx_manuals_type ON manuals(content_type);
CREATE INDEX IF NOT EXISTS idx_manuals_collection ON manuals(collection_id);

-- Full-text search on manuals
CREATE VIRTUAL TABLE IF NOT EXISTS manuals_fts USING fts5(
    title,
    content,
    content='manuals',
    content_rowid='id'
);

-- Triggers to keep FTS index in sync
CREATE TRIGGER IF NOT EXISTS manuals_ai AFTER INSERT ON manuals BEGIN
    INSERT INTO manuals_fts(rowid, title, content) VALUES (new.id, new.title, new.content);
END;

CREATE TRIGGER IF NOT EXISTS manuals_ad AFTER DELETE ON manuals BEGIN
    INSERT INTO manuals_fts(manuals_fts, rowid, title, content) VALUES('delete', old.id, old.title, old.content);
END;

CREATE TRIGGER IF NOT EXISTS manuals_au AFTER UPDATE ON manuals BEGIN
    INSERT INTO manuals_fts(manuals_fts, rowid, title, content) VALUES('delete', old.id, old.title, old.content);
    INSERT INTO manuals_fts(rowid, title, content) VALUES (new.id, new.title, new.content);
END;

-- ============================================================================
-- CROSS REFERENCES
-- ============================================================================

-- Links between scriptures (from footnotes)
CREATE TABLE IF NOT EXISTS cross_references (
    id INTEGER PRIMARY KEY,
    source_volume TEXT NOT NULL,
    source_book TEXT NOT NULL,
    source_chapter INTEGER NOT NULL,
    source_verse INTEGER NOT NULL,
    target_volume TEXT NOT NULL,
    target_book TEXT NOT NULL,
    target_chapter INTEGER NOT NULL,
    target_verse INTEGER,        -- NULL if whole chapter reference
    reference_type TEXT,         -- 'footnote', 'tg', 'bd', 'jst'
    
    FOREIGN KEY (source_volume, source_book, source_chapter, source_verse)
        REFERENCES scriptures(volume, book, chapter, verse)
);

CREATE INDEX IF NOT EXISTS idx_cross_ref_source ON cross_references(source_volume, source_book, source_chapter, source_verse);
CREATE INDEX IF NOT EXISTS idx_cross_ref_target ON cross_references(target_volume, target_book, target_chapter, target_verse);

-- ============================================================================
-- INDEX METADATA
-- ============================================================================

-- Tracks indexed files for incremental updates
CREATE TABLE IF NOT EXISTS index_metadata (
    file_path TEXT PRIMARY KEY,
    content_type TEXT NOT NULL,  -- 'scripture', 'talk', 'manual'
    indexed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    file_mtime DATETIME NOT NULL,
    file_size INTEGER NOT NULL,
    record_count INTEGER NOT NULL  -- verses/sections indexed from this file
);

CREATE INDEX IF NOT EXISTS idx_metadata_mtime ON index_metadata(file_mtime);
CREATE INDEX IF NOT EXISTS idx_metadata_type ON index_metadata(content_type);

-- ============================================================================
-- SCHEMA VERSION
-- ============================================================================

-- Track schema version for migrations
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY,
    applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Insert initial version if not exists
INSERT OR IGNORE INTO schema_version (version) VALUES (1);
