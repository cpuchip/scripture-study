package db

import (
	"fmt"
)

// Note represents a user note optionally linked to a practice, task, or pillar.
type Note struct {
	ID         int64  `json:"id"`
	Content    string `json:"content"`
	PracticeID *int64 `json:"practice_id,omitempty"`
	TaskID     *int64 `json:"task_id,omitempty"`
	PillarID   *int64 `json:"pillar_id,omitempty"`
	Pinned     bool   `json:"pinned"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`

	// Joined fields (read-only, populated on list)
	PracticeName string `json:"practice_name,omitempty"`
	TaskTitle    string `json:"task_title,omitempty"`
}

// CreateNote inserts a new note.
func (db *DB) CreateNote(n *Note) error {
	result, err := db.Exec(`
		INSERT INTO notes (content, practice_id, task_id, pillar_id, pinned)
		VALUES (?, ?, ?, ?, ?)`,
		n.Content, n.PracticeID, n.TaskID, n.PillarID, n.Pinned,
	)
	if err != nil {
		return fmt.Errorf("inserting note: %w", err)
	}
	n.ID, _ = result.LastInsertId()
	// Read back timestamps
	row := db.QueryRow(`SELECT created_at, updated_at FROM notes WHERE id = ?`, n.ID)
	_ = row.Scan(&n.CreatedAt, &n.UpdatedAt)
	return nil
}

// ListNotes returns notes with optional filters.
func (db *DB) ListNotes(practiceID, taskID, pillarID *int64, pinnedOnly bool) ([]*Note, error) {
	query := `
		SELECT n.id, n.content, n.practice_id, n.task_id, n.pillar_id, n.pinned,
		       n.created_at, n.updated_at,
		       COALESCE(p.name, ''), COALESCE(t.title, '')
		FROM notes n
		LEFT JOIN practices p ON n.practice_id = p.id
		LEFT JOIN tasks t ON n.task_id = t.id
		WHERE 1=1`
	args := []any{}

	if practiceID != nil {
		query += ` AND n.practice_id = ?`
		args = append(args, *practiceID)
	}
	if taskID != nil {
		query += ` AND n.task_id = ?`
		args = append(args, *taskID)
	}
	if pillarID != nil {
		query += ` AND n.pillar_id = ?`
		args = append(args, *pillarID)
	}
	if pinnedOnly {
		query += ` AND n.pinned = 1`
	}

	query += ` ORDER BY n.pinned DESC, n.created_at DESC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing notes: %w", err)
	}
	defer rows.Close()

	var notes []*Note
	for rows.Next() {
		n := &Note{}
		if err := rows.Scan(
			&n.ID, &n.Content, &n.PracticeID, &n.TaskID, &n.PillarID,
			&n.Pinned, &n.CreatedAt, &n.UpdatedAt,
			&n.PracticeName, &n.TaskTitle,
		); err != nil {
			return nil, fmt.Errorf("scanning note: %w", err)
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

// GetNote returns a single note by ID.
func (db *DB) GetNote(id int64) (*Note, error) {
	n := &Note{}
	err := db.QueryRow(`
		SELECT n.id, n.content, n.practice_id, n.task_id, n.pillar_id, n.pinned,
		       n.created_at, n.updated_at,
		       COALESCE(p.name, ''), COALESCE(t.title, '')
		FROM notes n
		LEFT JOIN practices p ON n.practice_id = p.id
		LEFT JOIN tasks t ON n.task_id = t.id
		WHERE n.id = ?`, id,
	).Scan(
		&n.ID, &n.Content, &n.PracticeID, &n.TaskID, &n.PillarID,
		&n.Pinned, &n.CreatedAt, &n.UpdatedAt,
		&n.PracticeName, &n.TaskTitle,
	)
	if err != nil {
		return nil, fmt.Errorf("getting note: %w", err)
	}
	return n, nil
}

// UpdateNote updates an existing note.
func (db *DB) UpdateNote(n *Note) error {
	_, err := db.Exec(`
		UPDATE notes SET content=?, practice_id=?, task_id=?, pillar_id=?, pinned=?, updated_at=CURRENT_TIMESTAMP
		WHERE id=?`,
		n.Content, n.PracticeID, n.TaskID, n.PillarID, n.Pinned, n.ID,
	)
	if err != nil {
		return fmt.Errorf("updating note: %w", err)
	}
	// Read back updated_at
	row := db.QueryRow(`SELECT updated_at FROM notes WHERE id = ?`, n.ID)
	_ = row.Scan(&n.UpdatedAt)
	return nil
}

// DeleteNote removes a note by ID.
func (db *DB) DeleteNote(id int64) error {
	_, err := db.Exec(`DELETE FROM notes WHERE id = ?`, id)
	return err
}
