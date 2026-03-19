package db

import (
	"encoding/json"
	"fmt"
	"time"
)

// AuditEntry represents a row from the audit_log table.
type AuditEntry struct {
	ID        int64           `json:"id"`
	TableName string          `json:"table_name"`
	RowID     int64           `json:"row_id"`
	Operation string          `json:"operation"`
	OldData   json.RawMessage `json:"old_data"`
	ChangedBy *int64          `json:"changed_by,omitempty"`
	ChangedAt time.Time       `json:"changed_at"`
}

// ListAuditLog returns audit entries, optionally filtered by table and row ID.
func (db *DB) ListAuditLog(tableName string, rowID int64, limit int) ([]*AuditEntry, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	query := `SELECT id, table_name, row_id, operation, old_data, changed_by, changed_at FROM audit_log WHERE 1=1`
	args := []any{}
	n := 0

	if tableName != "" {
		n++
		query += fmt.Sprintf(` AND table_name = $%d`, n)
		args = append(args, tableName)
	}
	if rowID > 0 {
		n++
		query += fmt.Sprintf(` AND row_id = $%d`, n)
		args = append(args, rowID)
	}

	n++
	query += fmt.Sprintf(` ORDER BY changed_at DESC LIMIT $%d`, n)
	args = append(args, limit)

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying audit log: %w", err)
	}
	defer rows.Close()

	var entries []*AuditEntry
	for rows.Next() {
		e := &AuditEntry{}
		if err := rows.Scan(&e.ID, &e.TableName, &e.RowID, &e.Operation, &e.OldData, &e.ChangedBy, &e.ChangedAt); err != nil {
			return nil, fmt.Errorf("scanning audit entry: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
