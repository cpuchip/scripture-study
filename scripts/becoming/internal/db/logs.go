package db

import (
	"fmt"
	"time"
)

// PracticeLog represents a single logged instance of doing a practice.
type PracticeLog struct {
	ID         int64     `json:"id"`
	PracticeID int64     `json:"practice_id"`
	LoggedAt   time.Time `json:"logged_at"`
	Date       string    `json:"date"` // YYYY-MM-DD

	Quality    *int    `json:"quality,omitempty"`     // SM-2 quality (0-5)
	Value      string  `json:"value,omitempty"`       // freeform: "25 min", "3 miles"
	Sets       *int    `json:"sets,omitempty"`        // number of sets
	Reps       *int    `json:"reps,omitempty"`        // reps per set
	DurationS  *int    `json:"duration_s,omitempty"`  // seconds
	Notes      string  `json:"notes,omitempty"`
	NextReview *string `json:"next_review,omitempty"` // date string
}

// CreateLog inserts a new practice log entry.
func (db *DB) CreateLog(l *PracticeLog) error {
	if l.Date == "" {
		l.Date = time.Now().Format("2006-01-02")
	}
	result, err := db.Exec(`
		INSERT INTO practice_logs (practice_id, date, quality, value, sets, reps, duration_s, notes, next_review)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		l.PracticeID, l.Date, l.Quality, l.Value, l.Sets, l.Reps, l.DurationS, l.Notes, l.NextReview,
	)
	if err != nil {
		return fmt.Errorf("inserting log: %w", err)
	}
	l.ID, _ = result.LastInsertId()
	l.LoggedAt = time.Now()
	return nil
}

// ListLogsByDate returns all logs for a specific date.
func (db *DB) ListLogsByDate(date string) ([]*PracticeLog, error) {
	rows, err := db.Query(`
		SELECT id, practice_id, logged_at, date, quality, value, sets, reps, duration_s, notes, next_review
		FROM practice_logs WHERE date = ?
		ORDER BY logged_at`, date,
	)
	if err != nil {
		return nil, fmt.Errorf("listing logs by date: %w", err)
	}
	defer rows.Close()
	return scanLogs(rows)
}

// ListLogsByPractice returns logs for a specific practice, ordered newest first.
func (db *DB) ListLogsByPractice(practiceID int64, limit int) ([]*PracticeLog, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := db.Query(`
		SELECT id, practice_id, logged_at, date, quality, value, sets, reps, duration_s, notes, next_review
		FROM practice_logs WHERE practice_id = ?
		ORDER BY logged_at DESC LIMIT ?`, practiceID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("listing logs by practice: %w", err)
	}
	defer rows.Close()
	return scanLogs(rows)
}

// ListLogsByPracticeRange returns logs for a practice between two dates (inclusive).
func (db *DB) ListLogsByPracticeRange(practiceID int64, startDate, endDate string) ([]*PracticeLog, error) {
	rows, err := db.Query(`
		SELECT id, practice_id, logged_at, date, quality, value, sets, reps, duration_s, notes, next_review
		FROM practice_logs WHERE practice_id = ? AND date >= ? AND date <= ?
		ORDER BY date`, practiceID, startDate, endDate,
	)
	if err != nil {
		return nil, fmt.Errorf("listing logs by range: %w", err)
	}
	defer rows.Close()
	return scanLogs(rows)
}

// DeleteLog removes a log entry.
func (db *DB) DeleteLog(id int64) error {
	_, err := db.Exec(`DELETE FROM practice_logs WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleting log: %w", err)
	}
	return nil
}

// DailySummary represents practice completion status for a single date.
type DailySummary struct {
	PracticeID   int64  `json:"practice_id"`
	PracticeName string `json:"practice_name"`
	PracticeType string `json:"practice_type"`
	Category     string `json:"category"`
	Config       string `json:"config"`
	LogCount     int    `json:"log_count"`
	TotalSets    *int   `json:"total_sets,omitempty"`
	TotalReps    *int   `json:"total_reps,omitempty"`
	LastValue    string `json:"last_value,omitempty"`
	LastNotes    string `json:"last_notes,omitempty"`
}

// GetDailySummary returns all active practices with their log status for a date.
func (db *DB) GetDailySummary(date string) ([]*DailySummary, error) {
	rows, err := db.Query(`
		SELECT
			p.id, p.name, p.type, p.category, p.config,
			COALESCE(COUNT(l.id), 0) as log_count,
			SUM(l.sets) as total_sets,
			SUM(l.reps) as total_reps,
			COALESCE((SELECT value FROM practice_logs WHERE practice_id = p.id AND date = ? ORDER BY logged_at DESC LIMIT 1), '') as last_value,
			COALESCE((SELECT notes FROM practice_logs WHERE practice_id = p.id AND date = ? ORDER BY logged_at DESC LIMIT 1), '') as last_notes
		FROM practices p
		LEFT JOIN practice_logs l ON l.practice_id = p.id AND l.date = ?
		WHERE p.active = 1
		GROUP BY p.id
		ORDER BY p.sort_order, p.type, p.name`, date, date, date,
	)
	if err != nil {
		return nil, fmt.Errorf("getting daily summary: %w", err)
	}
	defer rows.Close()

	var summaries []*DailySummary
	for rows.Next() {
		s := &DailySummary{}
		if err := rows.Scan(&s.PracticeID, &s.PracticeName, &s.PracticeType, &s.Category, &s.Config, &s.LogCount, &s.TotalSets, &s.TotalReps, &s.LastValue, &s.LastNotes); err != nil {
			return nil, fmt.Errorf("scanning daily summary: %w", err)
		}
		summaries = append(summaries, s)
	}
	return summaries, rows.Err()
}

func scanLogs(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]*PracticeLog, error) {
	var logs []*PracticeLog
	for rows.Next() {
		l := &PracticeLog{}
		if err := rows.Scan(&l.ID, &l.PracticeID, &l.LoggedAt, &l.Date, &l.Quality, &l.Value, &l.Sets, &l.Reps, &l.DurationS, &l.Notes, &l.NextReview); err != nil {
			return nil, fmt.Errorf("scanning log: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}
