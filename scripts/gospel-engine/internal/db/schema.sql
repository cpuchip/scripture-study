-- Gospel Engine Database Schema
-- SQLite with FTS5 for full-text search + graph edges

-- ============================================================================
-- SCRIPTURES
-- ============================================================================

CREATE TABLE IF NOT EXISTS scriptures (
    id INTEGER PRIMARY KEY,
    volume TEXT NOT NULL,        -- 'ot', 'nt', 'bofm', 'dc-testament', 'pgp'
    book TEXT NOT NULL,          -- 'gen', 'matt', '1-ne', 'dc', 'moses'
    chapter INTEGER NOT NULL,
    verse INTEGER NOT NULL,
    text TEXT NOT NULL,
    file_path TEXT NOT NULL,
    source_url TEXT NOT NULL,
    UNIQUE(volume, book, chapter, verse)
);

CREATE INDEX IF NOT EXISTS idx_scriptures_volume ON scriptures(volume);
CREATE INDEX IF NOT EXISTS idx_scriptures_book ON scriptures(volume, book);
CREATE INDEX IF NOT EXISTS idx_scriptures_chapter ON scriptures(volume, book, chapter);
CREATE INDEX IF NOT EXISTS idx_scriptures_file ON scriptures(file_path);

CREATE VIRTUAL TABLE IF NOT EXISTS scriptures_fts USING fts5(
    text,
    content='scriptures',
    content_rowid='id'
);

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

CREATE TABLE IF NOT EXISTS chapters (
    id INTEGER PRIMARY KEY,
    volume TEXT NOT NULL,
    book TEXT NOT NULL,
    chapter INTEGER NOT NULL,
    title TEXT,
    full_content TEXT NOT NULL,
    file_path TEXT NOT NULL,
    source_url TEXT,
    UNIQUE(volume, book, chapter)
);

CREATE INDEX IF NOT EXISTS idx_chapters_volume ON chapters(volume);
CREATE INDEX IF NOT EXISTS idx_chapters_book ON chapters(volume, book);
CREATE INDEX IF NOT EXISTS idx_chapters_file ON chapters(file_path);

-- ============================================================================
-- TALKS
-- ============================================================================

CREATE TABLE IF NOT EXISTS talks (
    id INTEGER PRIMARY KEY,
    year INTEGER NOT NULL,
    month INTEGER NOT NULL,
    session TEXT,
    speaker TEXT NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    file_path TEXT NOT NULL,
    source_url TEXT NOT NULL,
    -- TITSW enrichment columns (Phase 2)
    titsw_dominant TEXT,
    titsw_mode TEXT,
    titsw_pattern TEXT,
    titsw_teach INTEGER,
    titsw_help INTEGER,
    titsw_love INTEGER,
    titsw_spirit INTEGER,
    titsw_doctrine INTEGER,
    titsw_invite INTEGER,
    titsw_summary TEXT,
    titsw_key_quote TEXT,
    titsw_keywords TEXT,
    UNIQUE(file_path)
);

CREATE INDEX IF NOT EXISTS idx_talks_year ON talks(year);
CREATE INDEX IF NOT EXISTS idx_talks_year_month ON talks(year, month);
CREATE INDEX IF NOT EXISTS idx_talks_speaker ON talks(speaker);

CREATE VIRTUAL TABLE IF NOT EXISTS talks_fts USING fts5(
    title,
    speaker,
    content,
    content='talks',
    content_rowid='id'
);

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

CREATE TABLE IF NOT EXISTS manuals (
    id INTEGER PRIMARY KEY,
    content_type TEXT NOT NULL,
    collection_id TEXT NOT NULL,
    section TEXT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    file_path TEXT NOT NULL,
    source_url TEXT NOT NULL,
    UNIQUE(file_path)
);

CREATE INDEX IF NOT EXISTS idx_manuals_type ON manuals(content_type);
CREATE INDEX IF NOT EXISTS idx_manuals_collection ON manuals(collection_id);

CREATE VIRTUAL TABLE IF NOT EXISTS manuals_fts USING fts5(
    title,
    content,
    content='manuals',
    content_rowid='id'
);

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
-- BOOKS
-- ============================================================================

CREATE TABLE IF NOT EXISTS books (
    id INTEGER PRIMARY KEY,
    collection TEXT NOT NULL,
    section TEXT NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    file_path TEXT NOT NULL,
    UNIQUE(file_path)
);

CREATE INDEX IF NOT EXISTS idx_books_collection ON books(collection);

CREATE VIRTUAL TABLE IF NOT EXISTS books_fts USING fts5(
    title,
    content,
    content='books',
    content_rowid='id'
);

CREATE TRIGGER IF NOT EXISTS books_ai AFTER INSERT ON books BEGIN
    INSERT INTO books_fts(rowid, title, content) VALUES (new.id, new.title, new.content);
END;
CREATE TRIGGER IF NOT EXISTS books_ad AFTER DELETE ON books BEGIN
    INSERT INTO books_fts(books_fts, rowid, title, content) VALUES('delete', old.id, old.title, old.content);
END;
CREATE TRIGGER IF NOT EXISTS books_au AFTER UPDATE ON books BEGIN
    INSERT INTO books_fts(books_fts, rowid, title, content) VALUES('delete', old.id, old.title, old.content);
    INSERT INTO books_fts(rowid, title, content) VALUES (new.id, new.title, new.content);
END;

-- ============================================================================
-- CROSS REFERENCES
-- ============================================================================

CREATE TABLE IF NOT EXISTS cross_references (
    id INTEGER PRIMARY KEY,
    source_volume TEXT NOT NULL,
    source_book TEXT NOT NULL,
    source_chapter INTEGER NOT NULL,
    source_verse INTEGER NOT NULL,
    target_volume TEXT NOT NULL,
    target_book TEXT NOT NULL,
    target_chapter INTEGER NOT NULL,
    target_verse INTEGER,
    reference_type TEXT
);

CREATE INDEX IF NOT EXISTS idx_cross_ref_source ON cross_references(source_volume, source_book, source_chapter, source_verse);
CREATE INDEX IF NOT EXISTS idx_cross_ref_target ON cross_references(target_volume, target_book, target_chapter, target_verse);

-- ============================================================================
-- GRAPH EDGES
-- ============================================================================

CREATE TABLE IF NOT EXISTS edges (
    id INTEGER PRIMARY KEY,
    source_type TEXT NOT NULL,
    source_id TEXT NOT NULL,
    target_type TEXT NOT NULL,
    target_id TEXT NOT NULL,
    edge_type TEXT NOT NULL,     -- 'cross_reference', 'thematic', 'semantic', 'typological'
    weight REAL DEFAULT 1.0,
    metadata TEXT,               -- JSON
    created_at TEXT DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_edges_source ON edges(source_type, source_id);
CREATE INDEX IF NOT EXISTS idx_edges_target ON edges(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_edges_type ON edges(edge_type);

-- ============================================================================
-- INDEX METADATA
-- ============================================================================

CREATE TABLE IF NOT EXISTS index_metadata (
    file_path TEXT PRIMARY KEY,
    content_type TEXT NOT NULL,
    indexed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    file_mtime DATETIME NOT NULL,
    file_size INTEGER NOT NULL,
    record_count INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_metadata_type ON index_metadata(content_type);

-- ============================================================================
-- SCHEMA VERSION
-- ============================================================================

CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY,
    applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO schema_version (version) VALUES (1);

-- ============================================================================
-- VECTOR DOCUMENT METADATA (for mmap-backed vector search)
-- ============================================================================

CREATE TABLE IF NOT EXISTS vec_docs (
    id INTEGER PRIMARY KEY,
    collection TEXT NOT NULL,       -- e.g. "scriptures-verse", "conference-paragraph"
    vec_idx INTEGER NOT NULL,       -- position in the .vecf file (0-indexed)
    doc_id TEXT NOT NULL,           -- original chromem document ID
    content TEXT NOT NULL,          -- the indexed text content
    source TEXT NOT NULL,           -- scriptures, conference, manual, music
    layer TEXT NOT NULL,            -- verse, paragraph, summary, theme
    book TEXT DEFAULT '',
    chapter INTEGER DEFAULT 0,
    reference TEXT DEFAULT '',
    range_text TEXT DEFAULT '',
    file_path TEXT DEFAULT '',
    speaker TEXT DEFAULT '',
    position TEXT DEFAULT '',
    year INTEGER DEFAULT 0,
    month TEXT DEFAULT '',
    session TEXT DEFAULT '',
    talk_title TEXT DEFAULT '',
    UNIQUE(collection, vec_idx)
);

CREATE INDEX IF NOT EXISTS idx_vec_docs_coll ON vec_docs(collection);
CREATE INDEX IF NOT EXISTS idx_vec_docs_source ON vec_docs(source);
CREATE INDEX IF NOT EXISTS idx_vec_docs_lookup ON vec_docs(collection, vec_idx);
