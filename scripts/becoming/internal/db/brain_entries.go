package db

import (
	"encoding/json"
	"fmt"
	"time"
)

// BrainEntry is a cached copy of a brain.exe entry, stored on ibeco.me
// so the web UI can display all brain data even when the agent is offline.
type BrainEntry struct {
	ID         string         `json:"id"`
	Title      string         `json:"title"`
	Category   string         `json:"category"`
	Body       string         `json:"body"`
	Status     string         `json:"status,omitempty"`
	ActionDone bool           `json:"action_done,omitempty"`
	DueDate    string         `json:"due_date,omitempty"`
	NextAction string         `json:"next_action,omitempty"`
	Tags       []string       `json:"tags,omitempty"`
	SubTasks   []BrainSubTask `json:"subtasks,omitempty"`
	Source     string         `json:"source,omitempty"`
	CreatedAt  string         `json:"created_at"`
	UpdatedAt  string         `json:"updated_at"`
	SyncedAt   string         `json:"synced_at"`
}

// BrainSubTask is a cached subtask from brain.exe.
type BrainSubTask struct {
	ID        string `json:"id"`
	EntryID   string `json:"entry_id"`
	Text      string `json:"text"`
	Done      bool   `json:"done"`
	SortOrder int    `json:"sort_order"`
}

// UpsertBrainEntry inserts or updates a cached brain entry.
func (db *DB) UpsertBrainEntry(userID int64, e *BrainEntry) error {
	now := time.Now().UTC().Format(time.RFC3339)
	tagsJSON, _ := json.Marshal(e.Tags)
	subtasksJSON, _ := json.Marshal(e.SubTasks)

	_, err := db.Exec(`
		INSERT INTO brain_entries (id, user_id, title, category, body, status, action_done, due_date, next_action, tags, subtasks, source, created_at, updated_at, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id, user_id) DO UPDATE SET
			title = EXCLUDED.title,
			category = EXCLUDED.category,
			body = EXCLUDED.body,
			status = EXCLUDED.status,
			action_done = EXCLUDED.action_done,
			due_date = EXCLUDED.due_date,
			next_action = EXCLUDED.next_action,
			tags = EXCLUDED.tags,
			subtasks = EXCLUDED.subtasks,
			source = EXCLUDED.source,
			updated_at = EXCLUDED.updated_at,
			synced_at = EXCLUDED.synced_at`,
		e.ID, userID, e.Title, e.Category, e.Body, e.Status, e.ActionDone,
		e.DueDate, e.NextAction, string(tagsJSON), string(subtasksJSON), e.Source,
		e.CreatedAt, e.UpdatedAt, now,
	)
	return err
}

// BulkUpsertBrainEntries upserts many entries in a single transaction and
// removes entries that are no longer present in the sync payload.
// Conflict-aware: only overwrites cached entries when the incoming data
// is at least as recent (by updated_at) to avoid clobbering web edits.
func (db *DB) BulkUpsertBrainEntries(userID int64, entries []*BrainEntry) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().UTC().Format(time.RFC3339)

	// Track IDs we received
	receivedIDs := make(map[string]bool, len(entries))

	for _, e := range entries {
		receivedIDs[e.ID] = true
		tagsJSON, _ := json.Marshal(e.Tags)
		subtasksJSON, _ := json.Marshal(e.SubTasks)

		_, err = tx.Exec(`
			INSERT INTO brain_entries (id, user_id, title, category, body, status, action_done, due_date, next_action, tags, subtasks, source, created_at, updated_at, synced_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT (id, user_id) DO UPDATE SET
				title = EXCLUDED.title,
				category = EXCLUDED.category,
				body = EXCLUDED.body,
				status = EXCLUDED.status,
				action_done = EXCLUDED.action_done,
				due_date = EXCLUDED.due_date,
				next_action = EXCLUDED.next_action,
				tags = EXCLUDED.tags,
				subtasks = EXCLUDED.subtasks,
				source = EXCLUDED.source,
				updated_at = EXCLUDED.updated_at,
				synced_at = EXCLUDED.synced_at
			WHERE EXCLUDED.updated_at >= brain_entries.updated_at`,
			e.ID, userID, e.Title, e.Category, e.Body, e.Status, e.ActionDone,
			e.DueDate, e.NextAction, string(tagsJSON), string(subtasksJSON), e.Source,
			e.CreatedAt, e.UpdatedAt, now,
		)
		if err != nil {
			return fmt.Errorf("upserting entry %s: %w", e.ID, err)
		}
	}

	// Delete entries that were removed from the brain
	// (only if we received at least one entry — empty sync means agent has no entries, not a broken sync)
	if len(entries) > 0 {
		// Get existing IDs
		rows, err := tx.Query(`SELECT id FROM brain_entries WHERE user_id = ?`, userID)
		if err != nil {
			return fmt.Errorf("listing existing entries: %w", err)
		}
		defer rows.Close()

		var toDelete []string
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return fmt.Errorf("scanning entry id: %w", err)
			}
			if !receivedIDs[id] {
				toDelete = append(toDelete, id)
			}
		}

		for _, id := range toDelete {
			if _, err := tx.Exec(`DELETE FROM brain_entries WHERE id = ? AND user_id = ?`, id, userID); err != nil {
				return fmt.Errorf("deleting stale entry %s: %w", id, err)
			}
		}
	}

	return tx.Commit()
}

// ListBrainEntries returns all cached brain entries for a user, newest first.
// When includeParked is false, entries with status "someday" or "archived" are
// hidden — mirroring brain.exe's /api/entries default (shipped 04-23).
func (db *DB) ListBrainEntries(userID int64, category string, includeParked bool) ([]*BrainEntry, error) {
	query := `SELECT id, title, category, body, status, action_done, due_date, next_action, tags, subtasks, source, created_at, updated_at, synced_at
		FROM brain_entries WHERE user_id = ?`
	args := []any{userID}
	if category != "" {
		query += ` AND category = ?`
		args = append(args, category)
	}
	if !includeParked {
		query += ` AND (status IS NULL OR status NOT IN ('someday','archived'))`
	}
	query += ` ORDER BY updated_at DESC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing brain entries: %w", err)
	}
	defer rows.Close()

	var entries []*BrainEntry
	for rows.Next() {
		e := &BrainEntry{}
		var tagsJSON, subtasksJSON string
		if err := rows.Scan(&e.ID, &e.Title, &e.Category, &e.Body, &e.Status, &e.ActionDone,
			&e.DueDate, &e.NextAction, &tagsJSON, &subtasksJSON, &e.Source, &e.CreatedAt, &e.UpdatedAt, &e.SyncedAt); err != nil {
			return nil, fmt.Errorf("scanning brain entry: %w", err)
		}
		if tagsJSON != "" && tagsJSON != "null" {
			json.Unmarshal([]byte(tagsJSON), &e.Tags)
		}
		if subtasksJSON != "" && subtasksJSON != "null" {
			json.Unmarshal([]byte(subtasksJSON), &e.SubTasks)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// GetBrainEntry returns a single cached brain entry.
func (db *DB) GetBrainEntry(userID int64, entryID string) (*BrainEntry, error) {
	e := &BrainEntry{}
	var tagsJSON, subtasksJSON string
	err := db.QueryRow(`
		SELECT id, title, category, body, status, action_done, due_date, next_action, tags, subtasks, source, created_at, updated_at, synced_at
		FROM brain_entries WHERE id = ? AND user_id = ?`, entryID, userID).
		Scan(&e.ID, &e.Title, &e.Category, &e.Body, &e.Status, &e.ActionDone,
			&e.DueDate, &e.NextAction, &tagsJSON, &subtasksJSON, &e.Source, &e.CreatedAt, &e.UpdatedAt, &e.SyncedAt)
	if err != nil {
		return nil, fmt.Errorf("getting brain entry: %w", err)
	}
	if tagsJSON != "" && tagsJSON != "null" {
		json.Unmarshal([]byte(tagsJSON), &e.Tags)
	}
	if subtasksJSON != "" && subtasksJSON != "null" {
		json.Unmarshal([]byte(subtasksJSON), &e.SubTasks)
	}
	return e, nil
}

// EnsureBrainEntriesTable is a no-op — goose migrations handle the schema.
func (db *DB) EnsureBrainEntriesTable() error {
	return nil
}

// DeleteBrainEntry removes a cached brain entry.
func (db *DB) DeleteBrainEntry(userID int64, entryID string) error {
	_, err := db.Exec(`DELETE FROM brain_entries WHERE id = ? AND user_id = ?`, entryID, userID)
	return err
}
