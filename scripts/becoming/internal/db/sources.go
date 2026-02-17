package db

import (
	"database/sql"
	"fmt"
	"time"
)

// DocumentSource represents a git-based document source for the Study Reader.
type DocumentSource struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"user_id"`
	Name         string     `json:"name"`
	SourceType   string     `json:"source_type"` // github_public | github_private
	Repo         string     `json:"repo"`         // "owner/repo"
	Branch       string     `json:"branch"`
	IncludePaths string     `json:"include_paths"` // JSON array of glob patterns
	ExcludePaths string     `json:"exclude_paths"` // JSON array of glob patterns
	TreeCache    *string    `json:"tree_cache,omitempty"`
	TreeEtag     *string    `json:"tree_etag,omitempty"`
	TreeCachedAt *time.Time `json:"tree_cached_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// ReadingProgress tracks which documents a user has read.
type ReadingProgress struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	SourceID  int64     `json:"source_id"`
	FilePath  string    `json:"file_path"`
	ReadAt    time.Time `json:"read_at"`
	ScrollPct float64   `json:"scroll_pct"`
}

const sourceColumns = `id, user_id, name, source_type, repo, branch, include_paths, exclude_paths, tree_cache, tree_etag, tree_cached_at, created_at`

func scanSource(scanner interface{ Scan(...any) error }) (*DocumentSource, error) {
	s := &DocumentSource{}
	if err := scanner.Scan(&s.ID, &s.UserID, &s.Name, &s.SourceType, &s.Repo, &s.Branch, &s.IncludePaths, &s.ExcludePaths, &s.TreeCache, &s.TreeEtag, &s.TreeCachedAt, &s.CreatedAt); err != nil {
		return nil, err
	}
	return s, nil
}

// CreateSource inserts a new document source.
func (db *DB) CreateSource(userID int64, s *DocumentSource) error {
	if s.Branch == "" {
		s.Branch = "main"
	}
	if s.IncludePaths == "" {
		s.IncludePaths = "[]"
	}
	if s.ExcludePaths == "" {
		s.ExcludePaths = "[]"
	}
	id, err := db.InsertReturningID(`
		INSERT INTO document_sources (user_id, name, source_type, repo, branch, include_paths, exclude_paths)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID, s.Name, s.SourceType, s.Repo, s.Branch, s.IncludePaths, s.ExcludePaths,
	)
	if err != nil {
		return fmt.Errorf("inserting source: %w", err)
	}
	s.ID = id
	s.UserID = userID
	s.CreatedAt = time.Now()
	return nil
}

// GetSource returns a single source by ID, scoped to user.
func (db *DB) GetSource(userID, id int64) (*DocumentSource, error) {
	s, err := scanSource(db.QueryRow(`
		SELECT `+sourceColumns+`
		FROM document_sources WHERE id = ? AND user_id = ?`, id, userID))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting source: %w", err)
	}
	return s, nil
}

// ListSources returns all sources for a user.
func (db *DB) ListSources(userID int64) ([]*DocumentSource, error) {
	rows, err := db.Query(`
		SELECT `+sourceColumns+`
		FROM document_sources WHERE user_id = ?
		ORDER BY name`, userID)
	if err != nil {
		return nil, fmt.Errorf("listing sources: %w", err)
	}
	defer rows.Close()

	var sources []*DocumentSource
	for rows.Next() {
		s, err := scanSource(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning source: %w", err)
		}
		sources = append(sources, s)
	}
	return sources, nil
}

// UpdateSource updates a document source.
func (db *DB) UpdateSource(userID int64, s *DocumentSource) error {
	res, err := db.Exec(`
		UPDATE document_sources
		SET name = ?, source_type = ?, repo = ?, branch = ?, include_paths = ?, exclude_paths = ?
		WHERE id = ? AND user_id = ?`,
		s.Name, s.SourceType, s.Repo, s.Branch, s.IncludePaths, s.ExcludePaths, s.ID, userID)
	if err != nil {
		return fmt.Errorf("updating source: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("source not found")
	}
	return nil
}

// UpdateSourceTreeCache updates the cached tree data for a source.
func (db *DB) UpdateSourceTreeCache(userID, id int64, treeJSON, etag string) error {
	_, err := db.Exec(`
		UPDATE document_sources
		SET tree_cache = ?, tree_etag = ?, tree_cached_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ?`,
		treeJSON, etag, id, userID)
	if err != nil {
		return fmt.Errorf("updating tree cache: %w", err)
	}
	return nil
}

// DeleteSource removes a document source and its reading progress.
func (db *DB) DeleteSource(userID, id int64) error {
	res, err := db.Exec(`DELETE FROM document_sources WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return fmt.Errorf("deleting source: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("source not found")
	}
	return nil
}

// UpsertReadingProgress records or updates reading progress for a file.
func (db *DB) UpsertReadingProgress(userID, sourceID int64, filePath string, scrollPct float64) error {
	if db.driver == "postgres" {
		_, err := db.Exec(`
			INSERT INTO reading_progress (user_id, source_id, file_path, scroll_pct, read_at)
			VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
			ON CONFLICT (user_id, source_id, file_path)
			DO UPDATE SET scroll_pct = EXCLUDED.scroll_pct, read_at = CURRENT_TIMESTAMP`,
			userID, sourceID, filePath, scrollPct)
		return err
	}
	// SQLite
	_, err := db.Exec(`
		INSERT INTO reading_progress (user_id, source_id, file_path, scroll_pct, read_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, source_id, file_path)
		DO UPDATE SET scroll_pct = excluded.scroll_pct, read_at = CURRENT_TIMESTAMP`,
		userID, sourceID, filePath, scrollPct)
	return err
}

// ListReadingProgress returns reading progress for a source.
func (db *DB) ListReadingProgress(userID, sourceID int64) ([]*ReadingProgress, error) {
	rows, err := db.Query(`
		SELECT id, user_id, source_id, file_path, read_at, scroll_pct
		FROM reading_progress
		WHERE user_id = ? AND source_id = ?
		ORDER BY read_at DESC`, userID, sourceID)
	if err != nil {
		return nil, fmt.Errorf("listing reading progress: %w", err)
	}
	defer rows.Close()

	var progress []*ReadingProgress
	for rows.Next() {
		rp := &ReadingProgress{}
		if err := rows.Scan(&rp.ID, &rp.UserID, &rp.SourceID, &rp.FilePath, &rp.ReadAt, &rp.ScrollPct); err != nil {
			return nil, fmt.Errorf("scanning reading progress: %w", err)
		}
		progress = append(progress, rp)
	}
	return progress, nil
}
