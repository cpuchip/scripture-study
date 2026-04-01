package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL string

// DB wraps the SQLite database connection.
type DB struct {
	*sql.DB
	path string
}

// Open opens or creates the database at the specified path.
func Open(path string) (*DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating database directory: %w", err)
	}

	dsn := fmt.Sprintf("%s?_foreign_keys=on&_journal_mode=WAL", path)
	sqlDB, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	db := &DB{DB: sqlDB, path: path}

	if err := db.initSchema(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("initializing schema: %w", err)
	}

	return db, nil
}

func (db *DB) initSchema() error {
	_, err := db.Exec(schemaSQL)
	if err != nil {
		return err
	}
	return db.migrate()
}

// migrate applies incremental schema changes to existing databases.
func (db *DB) migrate() error {
	// Add TITSW analysis columns (reasoning, raw_output, model) if missing.
	// These were added after the initial schema.
	migrations := []string{
		"ALTER TABLE talks ADD COLUMN titsw_reasoning TEXT",
		"ALTER TABLE talks ADD COLUMN titsw_raw_output TEXT",
		"ALTER TABLE talks ADD COLUMN titsw_model TEXT",
		// Phase 3: Scripture enrichment columns on chapters
		"ALTER TABLE chapters ADD COLUMN enrichment_summary TEXT",
		"ALTER TABLE chapters ADD COLUMN enrichment_keywords TEXT",
		"ALTER TABLE chapters ADD COLUMN enrichment_key_verse TEXT",
		"ALTER TABLE chapters ADD COLUMN enrichment_christ_types TEXT",
		"ALTER TABLE chapters ADD COLUMN enrichment_connections TEXT",
		"ALTER TABLE chapters ADD COLUMN enrichment_model TEXT",
		"ALTER TABLE chapters ADD COLUMN enrichment_raw_output TEXT",
	}
	for _, m := range migrations {
		// ALTER TABLE ADD COLUMN fails if column exists; ignore that error.
		db.Exec(m)
	}

	// Rebuild talks FTS5 if it doesn't have titsw columns.
	// Check by looking at the column count in the FTS table.
	var colCount int
	err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('talks_fts')`).Scan(&colCount)
	if err == nil && colCount < 7 {
		// Old FTS has 3 columns (title, speaker, content). New has 7 (+titsw_dominant, titsw_mode, titsw_keywords, titsw_summary).
		db.Exec("DROP TRIGGER IF EXISTS talks_ai")
		db.Exec("DROP TRIGGER IF EXISTS talks_ad")
		db.Exec("DROP TRIGGER IF EXISTS talks_au")
		db.Exec("DROP TABLE IF EXISTS talks_fts")
		// Re-create from schema (which now has the extended FTS + triggers)
		_, err = db.Exec(schemaSQL)
		if err != nil {
			return fmt.Errorf("rebuilding FTS: %w", err)
		}
		// Rebuild FTS content from existing talks
		db.Exec(`INSERT INTO talks_fts(talks_fts) VALUES('rebuild')`)
	}

	// Create chapters FTS if it doesn't exist yet (Phase 3).
	var chaptersFtsExists int
	err = db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='chapters_fts'`).Scan(&chaptersFtsExists)
	if err == nil && chaptersFtsExists == 0 {
		// Re-run schema to create chapters_fts + triggers
		db.Exec(schemaSQL)
		db.Exec(`INSERT INTO chapters_fts(chapters_fts) VALUES('rebuild')`)
	}

	return nil
}

// Path returns the database file path.
func (db *DB) Path() string {
	return db.path
}

// Reset drops all data and recreates the schema.
func (db *DB) Reset() error {
	tables := []string{
		"index_metadata",
		"edges",
		"cross_references",
		"books_fts", "books",
		"manuals_fts", "manuals",
		"talks_fts", "talks",
		"chapters",
		"scriptures_fts", "scriptures",
		"schema_version",
	}
	for _, table := range tables {
		if _, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table)); err != nil {
			return fmt.Errorf("dropping table %s: %w", table, err)
		}
	}
	return db.initSchema()
}

// Stats holds counts of indexed content.
type Stats struct {
	Scriptures int64
	Chapters   int64
	Talks      int64
	Manuals    int64
	Books      int64
	CrossRefs  int64
	Edges      int64
}

// GetStats returns counts of indexed content.
func (db *DB) GetStats() (*Stats, error) {
	stats := &Stats{}
	queries := []struct {
		table string
		dest  *int64
	}{
		{"scriptures", &stats.Scriptures},
		{"chapters", &stats.Chapters},
		{"talks", &stats.Talks},
		{"manuals", &stats.Manuals},
		{"books", &stats.Books},
		{"cross_references", &stats.CrossRefs},
		{"edges", &stats.Edges},
	}
	for _, q := range queries {
		if err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", q.table)).Scan(q.dest); err != nil {
			return nil, fmt.Errorf("counting %s: %w", q.table, err)
		}
	}
	return stats, nil
}

// FileMetadata tracks indexed files for incremental updates.
type FileMetadata struct {
	FilePath    string
	ContentType string
	IndexedAt   time.Time
	FileMtime   time.Time
	FileSize    int64
	RecordCount int
}

// NeedsReindex checks if a file needs to be reindexed based on mtime and size.
func (db *DB) NeedsReindex(filePath string, mtime time.Time, size int64) (bool, error) {
	var storedMtime time.Time
	var storedSize int64
	err := db.QueryRow(`SELECT file_mtime, file_size FROM index_metadata WHERE file_path = ?`, filePath).
		Scan(&storedMtime, &storedSize)
	if err == sql.ErrNoRows {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return !storedMtime.Equal(mtime) || storedSize != size, nil
}

// SetFileMetadata inserts or updates metadata for a file.
func (db *DB) SetFileMetadata(m *FileMetadata) error {
	_, err := db.Exec(`
		INSERT OR REPLACE INTO index_metadata 
			(file_path, content_type, indexed_at, file_mtime, file_size, record_count)
		VALUES (?, ?, ?, ?, ?, ?)
	`, m.FilePath, m.ContentType, m.IndexedAt, m.FileMtime, m.FileSize, m.RecordCount)
	return err
}

// InsertEdge inserts a graph edge.
func (db *DB) InsertEdge(sourceType, sourceID, targetType, targetID, edgeType string, weight float64, metadata string) error {
	_, err := db.Exec(`
		INSERT INTO edges (source_type, source_id, target_type, target_id, edge_type, weight, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, sourceType, sourceID, targetType, targetID, edgeType, weight, metadata)
	return err
}
