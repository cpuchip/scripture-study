package db

import (
	"fmt"
)

// Task represents a commitment from a study document.
type Task struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description,omitempty"`
	SourceDoc     string `json:"source_doc,omitempty"`
	SourceSection string `json:"source_section,omitempty"`
	Scripture     string `json:"scripture,omitempty"`
	Type          string `json:"type"`   // once | daily | weekly | ongoing
	Status        string `json:"status"` // active | completed | paused | archived
	BrainEntryID  string `json:"brain_entry_id,omitempty"`
	CreatedAt     string `json:"created_at"`
	CompletedAt   string `json:"completed_at,omitempty"`
}

// CreateTask inserts a new task.
func (db *DB) CreateTask(userID int64, t *Task) error {
	id, err := db.InsertReturningID(`
		INSERT INTO tasks (user_id, title, description, source_doc, source_section, scripture, type, status, brain_entry_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, t.Title, t.Description, t.SourceDoc, t.SourceSection, t.Scripture, t.Type, t.Status, t.BrainEntryID,
	)
	if err != nil {
		return fmt.Errorf("inserting task: %w", err)
	}
	t.ID = id
	return nil
}

// ListTasks returns tasks, optionally filtered by status, scoped to user.
func (db *DB) ListTasks(userID int64, status string) ([]*Task, error) {
	// Postgres needs a cast to coalesce TIMESTAMPTZ with text; SQLite is fine without.
	completedExpr := "COALESCE(completed_at, '')"
	if db.IsPostgres() {
		completedExpr = "COALESCE(completed_at::text, '')"
	}
	query := `SELECT id, title, description, source_doc, source_section, scripture, type, status, COALESCE(brain_entry_id, ''), created_at, ` + completedExpr + ` FROM tasks WHERE user_id = ?`
	args := []any{userID}
	if status != "" {
		query += ` AND status = ?`
		args = append(args, status)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		t := &Task{}
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.SourceDoc, &t.SourceSection, &t.Scripture, &t.Type, &t.Status, &t.BrainEntryID, &t.CreatedAt, &t.CompletedAt); err != nil {
			return nil, fmt.Errorf("scanning task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// UpdateTask updates a task, scoped to user.
func (db *DB) UpdateTask(userID int64, t *Task) error {
	_, err := db.Exec(`
		UPDATE tasks SET title=?, description=?, source_doc=?, source_section=?, scripture=?, type=?, status=?, completed_at=?, brain_entry_id=?
		WHERE id=? AND user_id=?`,
		t.Title, t.Description, t.SourceDoc, t.SourceSection, t.Scripture, t.Type, t.Status, t.CompletedAt, t.BrainEntryID, t.ID, userID,
	)
	return err
}

// GetTask returns a single task by ID, scoped to user.
func (db *DB) GetTask(userID, id int64) (*Task, error) {
	t := &Task{}
	err := db.QueryRow(`SELECT id, title, description, source_doc, source_section, scripture, type, status, COALESCE(brain_entry_id, ''), created_at, COALESCE(completed_at, '')
		FROM tasks WHERE id = ? AND user_id = ?`, id, userID).
		Scan(&t.ID, &t.Title, &t.Description, &t.SourceDoc, &t.SourceSection, &t.Scripture, &t.Type, &t.Status, &t.BrainEntryID, &t.CreatedAt, &t.CompletedAt)
	if err != nil {
		return nil, fmt.Errorf("getting task: %w", err)
	}
	return t, nil
}

// DeleteTask removes a task, scoped to user.
func (db *DB) DeleteTask(userID, id int64) error {
	_, err := db.Exec(`DELETE FROM tasks WHERE id = ? AND user_id = ?`, id, userID)
	return err
}
