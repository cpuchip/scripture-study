package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

// Session represents an active browser session.
type Session struct {
	ID         string `json:"id"`
	UserID     int64  `json:"user_id"`
	CreatedAt  string `json:"created_at"`
	ExpiresAt  string `json:"expires_at"`
	LastActive string `json:"last_active"`
	UserAgent  string `json:"user_agent"`
	IPAddress  string `json:"ip_address"`
}

const sessionDuration = 30 * 24 * time.Hour // 30 days

// generateToken creates a cryptographically random hex token.
func generateToken(bytes int) (string, error) {
	b := make([]byte, bytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating random token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// CreateSession creates a new session for the given user.
func (db *DB) CreateSession(userID int64, userAgent, ipAddress string) (*Session, error) {
	token, err := generateToken(32)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	expires := now.Add(sessionDuration)

	_, err = db.Exec(
		`INSERT INTO sessions (id, user_id, created_at, expires_at, last_active, user_agent, ip_address)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		token, userID,
		now.Format(time.RFC3339),
		expires.Format(time.RFC3339),
		now.Format(time.RFC3339),
		userAgent, ipAddress,
	)
	if err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}
	return &Session{
		ID:         token,
		UserID:     userID,
		CreatedAt:  now.Format(time.RFC3339),
		ExpiresAt:  expires.Format(time.RFC3339),
		LastActive: now.Format(time.RFC3339),
		UserAgent:  userAgent,
		IPAddress:  ipAddress,
	}, nil
}

// GetSession retrieves a session by its token ID.
func (db *DB) GetSession(token string) (*Session, error) {
	s := &Session{}
	err := db.QueryRow(
		`SELECT id, user_id, created_at, expires_at, last_active, user_agent, ip_address
		 FROM sessions WHERE id = ?`, token,
	).Scan(&s.ID, &s.UserID, &s.CreatedAt, &s.ExpiresAt, &s.LastActive, &s.UserAgent, &s.IPAddress)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting session: %w", err)
	}
	return s, nil
}

// IsExpired checks if the session has expired.
func (s *Session) IsExpired() bool {
	t, err := time.Parse(time.RFC3339, s.ExpiresAt)
	if err != nil {
		return true
	}
	return time.Now().UTC().After(t)
}

// TouchSession updates the last_active and extends expiry (sliding window).
func (db *DB) TouchSession(token string) {
	now := time.Now().UTC()
	expires := now.Add(sessionDuration)
	db.Exec(
		`UPDATE sessions SET last_active = ?, expires_at = ? WHERE id = ?`,
		now.Format(time.RFC3339), expires.Format(time.RFC3339), token,
	)
}

// DeleteSession removes a session (logout).
func (db *DB) DeleteSession(token string) error {
	_, err := db.Exec(`DELETE FROM sessions WHERE id = ?`, token)
	return err
}

// DeleteUserSessions removes all sessions for a user.
func (db *DB) DeleteUserSessions(userID int64) error {
	_, err := db.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID)
	return err
}

// DeleteUserSessionsExcept removes all sessions for a user except the specified one.
func (db *DB) DeleteUserSessionsExcept(userID int64, exceptToken string) error {
	_, err := db.Exec(`DELETE FROM sessions WHERE user_id = ? AND id != ?`, userID, exceptToken)
	return err
}

// ListUserSessions returns all active sessions for a user.
func (db *DB) ListUserSessions(userID int64) ([]*Session, error) {
	rows, err := db.Query(
		`SELECT id, user_id, created_at, expires_at, last_active, user_agent, ip_address
		 FROM sessions WHERE user_id = ? ORDER BY last_active DESC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		s := &Session{}
		if err := rows.Scan(&s.ID, &s.UserID, &s.CreatedAt, &s.ExpiresAt, &s.LastActive, &s.UserAgent, &s.IPAddress); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

// CleanExpiredSessions removes expired sessions.
func (db *DB) CleanExpiredSessions() (int64, error) {
	res, err := db.Exec(`DELETE FROM sessions WHERE expires_at < ?`, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
