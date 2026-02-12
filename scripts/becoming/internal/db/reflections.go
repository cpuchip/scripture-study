package db

import (
	"database/sql"
	"fmt"
)

// Prompt represents a reflection question.
type Prompt struct {
	ID        int64  `json:"id"`
	Text      string `json:"text"`
	Active    bool   `json:"active"`
	SortOrder int    `json:"sort_order"`
	CreatedAt string `json:"created_at"`
}

// Reflection represents a daily journal entry.
type Reflection struct {
	ID         int64  `json:"id"`
	Date       string `json:"date"`
	PromptID   *int64 `json:"prompt_id,omitempty"`
	PromptText string `json:"prompt_text,omitempty"`
	Content    string `json:"content"`
	Mood       *int   `json:"mood,omitempty"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// --- Prompts ---

// SeedPrompts inserts default prompts for a user if they have none.
func (db *DB) SeedPrompts(userID int64) error {
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM prompts WHERE user_id = ?`, userID).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil // Already seeded
	}

	defaults := []string{
		"What did I learn today?",
		"What am I grateful for?",
		"How did I see God's hand today?",
		"What's one thing I did well?",
		"What do I want to do better tomorrow?",
		"What scripture spoke to me today?",
		"How did I serve someone today?",
	}
	for i, text := range defaults {
		if _, err := db.Exec(`INSERT INTO prompts (user_id, text, active, sort_order) VALUES (?, ?, 1, ?)`, userID, text, i); err != nil {
			return fmt.Errorf("seeding prompt %d: %w", i, err)
		}
	}
	return nil
}

// ListPrompts returns all prompts ordered by sort_order, scoped to user.
func (db *DB) ListPrompts(userID int64, activeOnly bool) ([]*Prompt, error) {
	query := `SELECT id, text, active, sort_order, created_at FROM prompts WHERE user_id = ?`
	args := []any{userID}
	if activeOnly {
		query += ` AND active = 1`
	}
	query += ` ORDER BY sort_order, id`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing prompts: %w", err)
	}
	defer rows.Close()

	var prompts []*Prompt
	for rows.Next() {
		p := &Prompt{}
		if err := rows.Scan(&p.ID, &p.Text, &p.Active, &p.SortOrder, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning prompt: %w", err)
		}
		prompts = append(prompts, p)
	}
	return prompts, rows.Err()
}

// GetTodayPrompt returns the prompt for the given day based on day-of-year % active prompt count.
func (db *DB) GetTodayPrompt(userID int64, dayOfYear int) (*Prompt, error) {
	prompts, err := db.ListPrompts(userID, true)
	if err != nil {
		return nil, err
	}
	if len(prompts) == 0 {
		return nil, nil
	}
	return prompts[dayOfYear%len(prompts)], nil
}

// CreatePrompt inserts a new prompt, scoped to user.
func (db *DB) CreatePrompt(userID int64, p *Prompt) error {
	// Default sort_order to max+1
	var maxOrder int
	_ = db.QueryRow(`SELECT COALESCE(MAX(sort_order), -1) FROM prompts WHERE user_id = ?`, userID).Scan(&maxOrder)
	if p.SortOrder == 0 {
		p.SortOrder = maxOrder + 1
	}

	result, err := db.Exec(`INSERT INTO prompts (user_id, text, active, sort_order) VALUES (?, ?, ?, ?)`,
		userID, p.Text, p.Active, p.SortOrder)
	if err != nil {
		return fmt.Errorf("inserting prompt: %w", err)
	}
	p.ID, _ = result.LastInsertId()
	row := db.QueryRow(`SELECT created_at FROM prompts WHERE id = ?`, p.ID)
	_ = row.Scan(&p.CreatedAt)
	return nil
}

// UpdatePrompt updates an existing prompt, scoped to user.
func (db *DB) UpdatePrompt(userID int64, p *Prompt) error {
	_, err := db.Exec(`UPDATE prompts SET text=?, active=?, sort_order=? WHERE id=? AND user_id=?`,
		p.Text, p.Active, p.SortOrder, p.ID, userID)
	return err
}

// DeletePrompt removes a prompt by ID, scoped to user.
func (db *DB) DeletePrompt(userID, id int64) error {
	_, err := db.Exec(`DELETE FROM prompts WHERE id = ? AND user_id = ?`, id, userID)
	return err
}

// --- Reflections ---

// ListReflections returns reflections ordered by date, optionally filtered by date range, scoped to user.
func (db *DB) ListReflections(userID int64, from, to string) ([]*Reflection, error) {
	query := `SELECT id, date, prompt_id, COALESCE(prompt_text, ''), content, mood, created_at, updated_at FROM reflections WHERE user_id = ?`
	args := []any{userID}
	if from != "" {
		query += ` AND date >= ?`
		args = append(args, from)
	}
	if to != "" {
		query += ` AND date <= ?`
		args = append(args, to)
	}
	query += ` ORDER BY date DESC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing reflections: %w", err)
	}
	defer rows.Close()

	var reflections []*Reflection
	for rows.Next() {
		r := &Reflection{}
		if err := rows.Scan(&r.ID, &r.Date, &r.PromptID, &r.PromptText, &r.Content, &r.Mood, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning reflection: %w", err)
		}
		reflections = append(reflections, r)
	}
	return reflections, rows.Err()
}

// GetReflection returns the reflection for a specific date, scoped to user.
func (db *DB) GetReflection(userID int64, date string) (*Reflection, error) {
	r := &Reflection{}
	err := db.QueryRow(`
		SELECT id, date, prompt_id, COALESCE(prompt_text, ''), content, mood, created_at, updated_at
		FROM reflections WHERE date = ? AND user_id = ?`, date, userID,
	).Scan(&r.ID, &r.Date, &r.PromptID, &r.PromptText, &r.Content, &r.Mood, &r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting reflection: %w", err)
	}
	return r, nil
}

// UpsertReflection creates or updates the reflection for a user+date.
func (db *DB) UpsertReflection(userID int64, r *Reflection) error {
	// Snapshot prompt text if prompt_id is set
	if r.PromptID != nil && r.PromptText == "" {
		var text string
		err := db.QueryRow(`SELECT text FROM prompts WHERE id = ? AND user_id = ?`, *r.PromptID, userID).Scan(&text)
		if err == nil {
			r.PromptText = text
		}
	}

	result, err := db.Exec(`
		INSERT INTO reflections (user_id, date, prompt_id, prompt_text, content, mood)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, date) DO UPDATE SET
			prompt_id = excluded.prompt_id,
			prompt_text = excluded.prompt_text,
			content = excluded.content,
			mood = excluded.mood,
			updated_at = CURRENT_TIMESTAMP`,
		userID, r.Date, r.PromptID, r.PromptText, r.Content, r.Mood,
	)
	if err != nil {
		return fmt.Errorf("upserting reflection: %w", err)
	}
	if r.ID == 0 {
		r.ID, _ = result.LastInsertId()
	}
	// Read back
	row := db.QueryRow(`SELECT id, created_at, updated_at FROM reflections WHERE date = ? AND user_id = ?`, r.Date, userID)
	_ = row.Scan(&r.ID, &r.CreatedAt, &r.UpdatedAt)
	return nil
}

// DeleteReflection removes a reflection by ID, scoped to user.
func (db *DB) DeleteReflection(userID, id int64) error {
	_, err := db.Exec(`DELETE FROM reflections WHERE id = ? AND user_id = ?`, id, userID)
	return err
}
