// Package db provides database operations for the Gospel MCP server.
package db

import (
	"database/sql"
	"time"
)

// FileMetadata tracks indexed files for incremental updates.
type FileMetadata struct {
	FilePath    string
	ContentType string
	IndexedAt   time.Time
	FileMtime   time.Time
	FileSize    int64
	RecordCount int
}

// GetFileMetadata retrieves metadata for a file if it exists.
func (db *DB) GetFileMetadata(filePath string) (*FileMetadata, error) {
	var m FileMetadata
	err := db.QueryRow(`
		SELECT file_path, content_type, indexed_at, file_mtime, file_size, record_count
		FROM index_metadata
		WHERE file_path = ?
	`, filePath).Scan(&m.FilePath, &m.ContentType, &m.IndexedAt, &m.FileMtime, &m.FileSize, &m.RecordCount)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
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

// DeleteFileMetadata removes metadata for a file.
func (db *DB) DeleteFileMetadata(filePath string) error {
	_, err := db.Exec(`DELETE FROM index_metadata WHERE file_path = ?`, filePath)
	return err
}

// NeedsReindex checks if a file needs to be reindexed based on mtime and size.
func (db *DB) NeedsReindex(filePath string, mtime time.Time, size int64) (bool, error) {
	meta, err := db.GetFileMetadata(filePath)
	if err != nil {
		return false, err
	}

	// New file, needs indexing
	if meta == nil {
		return true, nil
	}

	// Check if file has been modified
	if !meta.FileMtime.Equal(mtime) || meta.FileSize != size {
		return true, nil
	}

	return false, nil
}

// GetAllMetadata returns all file metadata, optionally filtered by content type.
func (db *DB) GetAllMetadata(contentType string) ([]FileMetadata, error) {
	var query string
	var args []interface{}

	if contentType != "" {
		query = `SELECT file_path, content_type, indexed_at, file_mtime, file_size, record_count 
				 FROM index_metadata WHERE content_type = ? ORDER BY file_path`
		args = []interface{}{contentType}
	} else {
		query = `SELECT file_path, content_type, indexed_at, file_mtime, file_size, record_count 
				 FROM index_metadata ORDER BY file_path`
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []FileMetadata
	for rows.Next() {
		var m FileMetadata
		if err := rows.Scan(&m.FilePath, &m.ContentType, &m.IndexedAt, &m.FileMtime, &m.FileSize, &m.RecordCount); err != nil {
			return nil, err
		}
		results = append(results, m)
	}

	return results, rows.Err()
}

// ClearMetadataByType removes all metadata for a content type.
func (db *DB) ClearMetadataByType(contentType string) error {
	_, err := db.Exec(`DELETE FROM index_metadata WHERE content_type = ?`, contentType)
	return err
}
