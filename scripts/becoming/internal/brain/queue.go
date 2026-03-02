package brain

import (
	"fmt"
	"time"

	"github.com/cpuchip/scripture-study/scripts/becoming/internal/db"
)

// Queue manages the persistent message queue in SQLite.
type Queue struct {
	db *db.DB
}

// NewQueue creates a queue backed by the given database.
func NewQueue(database *db.DB) *Queue {
	return &Queue{db: database}
}

// EnsureTable creates the brain_messages table if it doesn't exist.
// For SQLite, this runs CREATE TABLE IF NOT EXISTS directly.
// For PostgreSQL, the table is created by goose migration 008_brain_messages.sql.
func (q *Queue) EnsureTable() error {
	if q.db.IsPostgres() {
		// Goose migrations handle PostgreSQL schema
		return nil
	}

	_, err := q.db.Exec(`
		CREATE TABLE IF NOT EXISTS brain_messages (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			message_id   TEXT NOT NULL UNIQUE,
			user_id      INTEGER NOT NULL,
			direction    TEXT NOT NULL,
			payload      TEXT NOT NULL,
			status       TEXT NOT NULL DEFAULT 'pending',
			created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			delivered_at DATETIME,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("creating brain_messages table: %w", err)
	}

	_, err = q.db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_brain_messages_pending
		ON brain_messages(user_id, status, direction)
	`)
	if err != nil {
		return fmt.Errorf("creating brain_messages index: %w", err)
	}

	return nil
}

// Enqueue stores a message for later delivery.
func (q *Queue) Enqueue(messageID string, userID int64, direction Direction, payload []byte) error {
	_, err := q.db.Exec(
		`INSERT INTO brain_messages (message_id, user_id, direction, payload, status, created_at)
		 VALUES (?, ?, ?, ?, 'pending', ?)`,
		messageID, userID, string(direction), string(payload), time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("enqueueing message %s: %w", messageID, err)
	}
	return nil
}

// DequeueAll retrieves and marks as delivered all pending messages for a user+direction.
func (q *Queue) DequeueAll(userID int64, direction Direction) ([][]byte, error) {
	rows, err := q.db.Query(
		`SELECT id, payload FROM brain_messages
		 WHERE user_id = ? AND direction = ? AND status = 'pending'
		 ORDER BY created_at ASC`,
		userID, string(direction),
	)
	if err != nil {
		return nil, fmt.Errorf("querying pending messages: %w", err)
	}
	defer rows.Close()

	var ids []int64
	var payloads [][]byte
	for rows.Next() {
		var id int64
		var payload string
		if err := rows.Scan(&id, &payload); err != nil {
			return nil, fmt.Errorf("scanning pending message: %w", err)
		}
		ids = append(ids, id)
		payloads = append(payloads, []byte(payload))
	}

	// Mark all as delivered
	now := time.Now().UTC().Format(time.RFC3339)
	for _, id := range ids {
		_, _ = q.db.Exec(
			`UPDATE brain_messages SET status = 'delivered', delivered_at = ? WHERE id = ?`,
			now, id,
		)
	}

	return payloads, nil
}

// History returns recent messages for a user (both directions), newest first.
func (q *Queue) History(userID int64, limit int) ([]QueueEntry, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := q.db.Query(
		`SELECT id, message_id, user_id, direction, payload, status, created_at, delivered_at
		 FROM brain_messages
		 WHERE user_id = ?
		 ORDER BY created_at DESC
		 LIMIT ?`,
		userID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("querying message history: %w", err)
	}
	defer rows.Close()

	var entries []QueueEntry
	for rows.Next() {
		var e QueueEntry
		var dir string
		var createdStr string
		var deliveredStr *string
		if err := rows.Scan(&e.ID, &e.MessageID, &e.UserID, &dir, &e.Payload, &e.Status, &createdStr, &deliveredStr); err != nil {
			return nil, fmt.Errorf("scanning history entry: %w", err)
		}
		e.Direction = Direction(dir)
		e.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
		if deliveredStr != nil {
			t, _ := time.Parse(time.RFC3339, *deliveredStr)
			e.DeliveredAt = &t
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// PendingCount returns the number of pending messages in each direction.
func (q *Queue) PendingCount(userID int64) (toAgent int, toApp int, err error) {
	row := q.db.QueryRow(
		`SELECT COALESCE(SUM(CASE WHEN direction = 'to_agent' THEN 1 ELSE 0 END), 0),
		        COALESCE(SUM(CASE WHEN direction = 'to_app' THEN 1 ELSE 0 END), 0)
		 FROM brain_messages
		 WHERE user_id = ? AND status = 'pending'`, userID,
	)
	err = row.Scan(&toAgent, &toApp)
	return
}

// Cleanup removes delivered messages older than the given duration.
func (q *Queue) Cleanup(olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan).UTC().Format(time.RFC3339)
	res, err := q.db.Exec(
		`DELETE FROM brain_messages WHERE status = 'delivered' AND delivered_at < ?`,
		cutoff,
	)
	if err != nil {
		return 0, fmt.Errorf("cleaning up old messages: %w", err)
	}
	return res.RowsAffected()
}
