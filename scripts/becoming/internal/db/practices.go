package db

import (
	"database/sql"
	"fmt"
	"time"
)

// Practice lifecycle status values.
const (
	StatusActive    = "active"
	StatusPaused    = "paused"
	StatusCompleted = "completed"
	StatusArchived  = "archived"
)

// Practice represents a trackable item (memorization, exercise, habit, etc.)
type Practice struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Type        string     `json:"type"`     // memorize | tracker | habit | task
	Category    string     `json:"category"` // scripture, pt, spiritual, fitness, etc.
	SourceDoc   string     `json:"source_doc,omitempty"`
	SourcePath  string     `json:"source_path,omitempty"`
	Config      string     `json:"config"`
	SortOrder   int        `json:"sort_order"`
	Active      bool       `json:"active"` // legacy — use Status instead
	Status      string     `json:"status"` // active | paused | completed | archived
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	ArchivedAt  *time.Time `json:"archived_at,omitempty"`
	EndDate     *string    `json:"end_date,omitempty"` // target end date (YYYY-MM-DD)
}

// practiceColumns is the standard SELECT column list for practices.
const practiceColumns = `id, name, description, type, category, source_doc, source_path, config, sort_order, active, status, created_at, completed_at, archived_at, end_date`

// scanPractice scans a row into a Practice struct. Column order must match practiceColumns.
func scanPractice(scanner interface{ Scan(...any) error }) (*Practice, error) {
	p := &Practice{}
	if err := scanner.Scan(&p.ID, &p.Name, &p.Description, &p.Type, &p.Category, &p.SourceDoc, &p.SourcePath, &p.Config, &p.SortOrder, &p.Active, &p.Status, &p.CreatedAt, &p.CompletedAt, &p.ArchivedAt, &p.EndDate); err != nil {
		return nil, err
	}
	return p, nil
}

// CreatePractice inserts a new practice.
func (db *DB) CreatePractice(userID int64, p *Practice) error {
	if p.Status == "" {
		p.Status = StatusActive
	}
	id, err := db.InsertReturningID(`
		INSERT INTO practices (user_id, name, description, type, category, source_doc, source_path, config, sort_order, active, status, end_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, p.Name, p.Description, p.Type, p.Category, p.SourceDoc, p.SourcePath, p.Config, p.SortOrder, p.Active, p.Status, p.EndDate,
	)
	if err != nil {
		return fmt.Errorf("inserting practice: %w", err)
	}
	p.ID = id
	p.CreatedAt = time.Now()
	return nil
}

// GetPractice returns a single practice by ID, scoped to the given user.
func (db *DB) GetPractice(userID, id int64) (*Practice, error) {
	p, err := scanPractice(db.QueryRow(`
		SELECT `+practiceColumns+`
		FROM practices WHERE id = ? AND user_id = ?`, id, userID))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting practice: %w", err)
	}
	return p, nil
}

// ListPractices returns practices, optionally filtered by type and/or status.
// For backward compatibility: activeOnly=true maps to status='active'.
func (db *DB) ListPractices(userID int64, practiceType string, activeOnly bool) ([]*Practice, error) {
	return db.ListPracticesByStatus(userID, practiceType, "", activeOnly)
}

// ListPracticesByStatus returns practices filtered by type and/or explicit status.
// If status is non-empty, it takes precedence over activeOnly. If status is empty,
// activeOnly=true filters to status='active' (legacy behavior).
func (db *DB) ListPracticesByStatus(userID int64, practiceType, status string, activeOnly bool) ([]*Practice, error) {
	query := `SELECT ` + practiceColumns + ` FROM practices WHERE user_id = ?`
	args := []any{userID}

	if practiceType != "" {
		query += ` AND type = ?`
		args = append(args, practiceType)
	}
	if status != "" {
		query += ` AND status = ?`
		args = append(args, status)
	} else if activeOnly {
		query += ` AND status = 'active'`
	}
	query += ` ORDER BY sort_order, name`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing practices: %w", err)
	}
	defer rows.Close()

	var practices []*Practice
	for rows.Next() {
		p, err := scanPractice(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning practice: %w", err)
		}
		practices = append(practices, p)
	}
	return practices, rows.Err()
}

// UpdatePractice updates an existing practice, scoped to the given user.
func (db *DB) UpdatePractice(userID int64, p *Practice) error {
	_, err := db.Exec(`
		UPDATE practices SET name=?, description=?, type=?, category=?, source_doc=?, source_path=?,
			config=?, sort_order=?, active=?, status=?, completed_at=?, archived_at=?, end_date=?
		WHERE id=? AND user_id=?`,
		p.Name, p.Description, p.Type, p.Category, p.SourceDoc, p.SourcePath,
		p.Config, p.SortOrder, p.Active, p.Status, p.CompletedAt, p.ArchivedAt, p.EndDate,
		p.ID, userID,
	)
	if err != nil {
		return fmt.Errorf("updating practice: %w", err)
	}
	return nil
}

// CompletePractice marks a practice as completed.
func (db *DB) CompletePractice(userID, id int64) error {
	_, err := db.Exec(`
		UPDATE practices SET status = 'completed', active = FALSE, completed_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return fmt.Errorf("completing practice: %w", err)
	}
	return nil
}

// ArchivePractice marks a practice as archived.
func (db *DB) ArchivePractice(userID, id int64) error {
	_, err := db.Exec(`
		UPDATE practices SET status = 'archived', active = FALSE, archived_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return fmt.Errorf("archiving practice: %w", err)
	}
	return nil
}

// PausePractice marks a practice as paused.
func (db *DB) PausePractice(userID, id int64) error {
	_, err := db.Exec(`
		UPDATE practices SET status = 'paused', active = FALSE
		WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return fmt.Errorf("pausing practice: %w", err)
	}
	return nil
}

// RestorePractice restores a practice to active status.
func (db *DB) RestorePractice(userID, id int64) error {
	_, err := db.Exec(`
		UPDATE practices SET status = 'active', active = TRUE, completed_at = NULL, archived_at = NULL
		WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return fmt.Errorf("restoring practice: %w", err)
	}
	return nil
}

// DeletePractice removes a practice and its logs (cascade), scoped to the given user.
func (db *DB) DeletePractice(userID, id int64) error {
	_, err := db.Exec(`DELETE FROM practices WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return fmt.Errorf("deleting practice: %w", err)
	}
	return nil
}
