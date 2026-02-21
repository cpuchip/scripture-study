package db

import "fmt"

// Bookmark represents a saved location in the reader.
type Bookmark struct {
	ID        int64  `json:"id"`
	SourceID  int64  `json:"source_id"`
	FilePath  string `json:"file_path"`
	Anchor    string `json:"anchor,omitempty"`
	Excerpt   string `json:"excerpt,omitempty"`
	Note      string `json:"note,omitempty"`
	CreatedAt string `json:"created_at"`

	// Joined fields (read-only)
	SourceName string `json:"source_name,omitempty"`
}

// CreateBookmark inserts a new bookmark.
func (db *DB) CreateBookmark(userID int64, b *Bookmark) error {
	id, err := db.InsertReturningID(`
		INSERT INTO bookmarks (user_id, source_id, file_path, anchor, excerpt, note)
		VALUES (?, ?, ?, ?, ?, ?)`,
		userID, b.SourceID, b.FilePath, b.Anchor, b.Excerpt, b.Note,
	)
	if err != nil {
		return fmt.Errorf("inserting bookmark: %w", err)
	}
	b.ID = id
	row := db.QueryRow(`SELECT created_at FROM bookmarks WHERE id = ?`, b.ID)
	_ = row.Scan(&b.CreatedAt)
	return nil
}

// ListBookmarks returns all bookmarks for a user, optionally filtered by source_id.
func (db *DB) ListBookmarks(userID int64, sourceID *int64) ([]*Bookmark, error) {
	query := `
		SELECT b.id, b.source_id, b.file_path, b.anchor, b.excerpt, b.note, b.created_at,
		       COALESCE(ds.name, '')
		FROM bookmarks b
		LEFT JOIN document_sources ds ON b.source_id = ds.id
		WHERE b.user_id = ?`
	args := []any{userID}

	if sourceID != nil {
		query += ` AND b.source_id = ?`
		args = append(args, *sourceID)
	}

	query += ` ORDER BY b.created_at DESC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing bookmarks: %w", err)
	}
	defer rows.Close()

	var bookmarks []*Bookmark
	for rows.Next() {
		b := &Bookmark{}
		if err := rows.Scan(
			&b.ID, &b.SourceID, &b.FilePath, &b.Anchor, &b.Excerpt, &b.Note, &b.CreatedAt,
			&b.SourceName,
		); err != nil {
			return nil, fmt.Errorf("scanning bookmark: %w", err)
		}
		bookmarks = append(bookmarks, b)
	}
	return bookmarks, rows.Err()
}

// ListBookmarksForFile returns bookmarks for a specific file.
func (db *DB) ListBookmarksForFile(userID, sourceID int64, filePath string) ([]*Bookmark, error) {
	rows, err := db.Query(`
		SELECT id, source_id, file_path, anchor, excerpt, note, created_at
		FROM bookmarks
		WHERE user_id = ? AND source_id = ? AND file_path = ?
		ORDER BY created_at ASC`,
		userID, sourceID, filePath,
	)
	if err != nil {
		return nil, fmt.Errorf("listing file bookmarks: %w", err)
	}
	defer rows.Close()

	var bookmarks []*Bookmark
	for rows.Next() {
		b := &Bookmark{}
		if err := rows.Scan(&b.ID, &b.SourceID, &b.FilePath, &b.Anchor, &b.Excerpt, &b.Note, &b.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning bookmark: %w", err)
		}
		bookmarks = append(bookmarks, b)
	}
	return bookmarks, rows.Err()
}

// UpdateBookmarkNote updates the note on a bookmark.
func (db *DB) UpdateBookmarkNote(userID, id int64, note string) error {
	res, err := db.Exec(`UPDATE bookmarks SET note = ? WHERE id = ? AND user_id = ?`, note, id, userID)
	if err != nil {
		return fmt.Errorf("updating bookmark note: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("bookmark not found")
	}
	return nil
}

// DeleteBookmark removes a bookmark by ID, scoped to user.
func (db *DB) DeleteBookmark(userID, id int64) error {
	_, err := db.Exec(`DELETE FROM bookmarks WHERE id = ? AND user_id = ?`, id, userID)
	return err
}
