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
	CreatedAt     string `json:"created_at"`
	CompletedAt   string `json:"completed_at,omitempty"`
}

// CreateTask inserts a new task.
func (db *DB) CreateTask(t *Task) error {
	result, err := db.Exec(`
		INSERT INTO tasks (title, description, source_doc, source_section, scripture, type, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		t.Title, t.Description, t.SourceDoc, t.SourceSection, t.Scripture, t.Type, t.Status,
	)
	if err != nil {
		return fmt.Errorf("inserting task: %w", err)
	}
	t.ID, _ = result.LastInsertId()
	return nil
}

// ListTasks returns tasks, optionally filtered by status.
func (db *DB) ListTasks(status string) ([]*Task, error) {
	query := `SELECT id, title, description, source_doc, source_section, scripture, type, status, created_at, COALESCE(completed_at, '') FROM tasks WHERE 1=1`
	args := []any{}
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
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.SourceDoc, &t.SourceSection, &t.Scripture, &t.Type, &t.Status, &t.CreatedAt, &t.CompletedAt); err != nil {
			return nil, fmt.Errorf("scanning task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// UpdateTask updates a task.
func (db *DB) UpdateTask(t *Task) error {
	_, err := db.Exec(`
		UPDATE tasks SET title=?, description=?, source_doc=?, source_section=?, scripture=?, type=?, status=?, completed_at=?
		WHERE id=?`,
		t.Title, t.Description, t.SourceDoc, t.SourceSection, t.Scripture, t.Type, t.Status, t.CompletedAt, t.ID,
	)
	return err
}

// DeleteTask removes a task.
func (db *DB) DeleteTask(id int64) error {
	_, err := db.Exec(`DELETE FROM tasks WHERE id = ?`, id)
	return err
}
