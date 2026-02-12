package db

import (
	"database/sql"
	"fmt"
	"time"
)

// Practice represents a trackable item (memorization, exercise, habit, etc.)
type Practice struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Type        string     `json:"type"`     // memorize | exercise | habit | task
	Category    string     `json:"category"` // scripture, pt, spiritual, fitness, etc.
	SourceDoc   string     `json:"source_doc,omitempty"`
	SourcePath  string     `json:"source_path,omitempty"`
	Config      string     `json:"config"`
	SortOrder   int        `json:"sort_order"`
	Active      bool       `json:"active"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// CreatePractice inserts a new practice.
func (db *DB) CreatePractice(p *Practice) error {
	result, err := db.Exec(`
		INSERT INTO practices (name, description, type, category, source_doc, source_path, config, sort_order, active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.Name, p.Description, p.Type, p.Category, p.SourceDoc, p.SourcePath, p.Config, p.SortOrder, p.Active,
	)
	if err != nil {
		return fmt.Errorf("inserting practice: %w", err)
	}
	p.ID, _ = result.LastInsertId()
	p.CreatedAt = time.Now()
	return nil
}

// GetPractice returns a single practice by ID.
func (db *DB) GetPractice(id int64) (*Practice, error) {
	p := &Practice{}
	err := db.QueryRow(`
		SELECT id, name, description, type, category, source_doc, source_path, config, sort_order, active, created_at, completed_at
		FROM practices WHERE id = ?`, id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.Type, &p.Category, &p.SourceDoc, &p.SourcePath, &p.Config, &p.SortOrder, &p.Active, &p.CreatedAt, &p.CompletedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting practice: %w", err)
	}
	return p, nil
}

// ListPractices returns practices, optionally filtered by type and/or active status.
func (db *DB) ListPractices(practiceType string, activeOnly bool) ([]*Practice, error) {
	query := `SELECT id, name, description, type, category, source_doc, source_path, config, sort_order, active, created_at, completed_at FROM practices WHERE 1=1`
	args := []any{}

	if practiceType != "" {
		query += ` AND type = ?`
		args = append(args, practiceType)
	}
	if activeOnly {
		query += ` AND active = 1`
	}
	query += ` ORDER BY sort_order, name`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing practices: %w", err)
	}
	defer rows.Close()

	var practices []*Practice
	for rows.Next() {
		p := &Practice{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Type, &p.Category, &p.SourceDoc, &p.SourcePath, &p.Config, &p.SortOrder, &p.Active, &p.CreatedAt, &p.CompletedAt); err != nil {
			return nil, fmt.Errorf("scanning practice: %w", err)
		}
		practices = append(practices, p)
	}
	return practices, rows.Err()
}

// UpdatePractice updates an existing practice.
func (db *DB) UpdatePractice(p *Practice) error {
	_, err := db.Exec(`
		UPDATE practices SET name=?, description=?, type=?, category=?, source_doc=?, source_path=?, config=?, sort_order=?, active=?, completed_at=?
		WHERE id=?`,
		p.Name, p.Description, p.Type, p.Category, p.SourceDoc, p.SourcePath, p.Config, p.SortOrder, p.Active, p.CompletedAt, p.ID,
	)
	if err != nil {
		return fmt.Errorf("updating practice: %w", err)
	}
	return nil
}

// DeletePractice removes a practice and its logs (cascade).
func (db *DB) DeletePractice(id int64) error {
	_, err := db.Exec(`DELETE FROM practices WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleting practice: %w", err)
	}
	return nil
}
