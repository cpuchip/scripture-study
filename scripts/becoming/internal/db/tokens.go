package db

import (
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// APIToken represents a programmatic access token.
type APIToken struct {
	ID           int64   `json:"id"`
	UserID       int64   `json:"user_id"`
	Name         string  `json:"name"`
	Prefix       string  `json:"prefix"`
	CreatedAt    string  `json:"created_at"`
	LastUsed     *string `json:"last_used"`
	ExpiresAt    *string `json:"expires_at"`
	BrainEnabled bool    `json:"brain_enabled"`
}

const tokenPrefix = "bec_"

// CreateAPIToken generates a new API token and stores its bcrypt hash.
// Returns the APIToken metadata AND the raw token string (shown once to the user).
func (db *DB) CreateAPIToken(userID int64, name string) (*APIToken, string, error) {
	raw, err := generateToken(32) // 64 hex chars
	if err != nil {
		return nil, "", err
	}
	fullToken := tokenPrefix + raw
	prefix := fullToken[:12] // "bec_" + 8 chars

	hash, err := bcrypt.GenerateFromPassword([]byte(fullToken), bcryptCost)
	if err != nil {
		return nil, "", fmt.Errorf("hashing api token: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	id, err := db.InsertReturningID(
		`INSERT INTO api_tokens (user_id, name, token_hash, prefix, created_at) VALUES (?, ?, ?, ?, ?)`,
		userID, name, string(hash), prefix, now,
	)
	if err != nil {
		return nil, "", fmt.Errorf("creating api token: %w", err)
	}

	return &APIToken{
		ID:        id,
		UserID:    userID,
		Name:      name,
		Prefix:    prefix,
		CreatedAt: now,
	}, fullToken, nil
}

// ValidateAPIToken checks a raw Bearer token against stored hashes.
// Returns the matching APIToken if valid, nil otherwise.
func (db *DB) ValidateAPIToken(rawToken string) (*APIToken, error) {
	// Use the prefix to narrow the search (avoids comparing against every hash)
	if len(rawToken) < 12 {
		return nil, nil
	}
	prefix := rawToken[:12]

	rows, err := db.Query(
		`SELECT id, user_id, name, prefix, token_hash, created_at, last_used, expires_at, brain_enabled
		 FROM api_tokens WHERE prefix = ?`, prefix,
	)
	if err != nil {
		return nil, fmt.Errorf("querying api tokens: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var t APIToken
		var hash string
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Prefix, &hash, &t.CreatedAt, &t.LastUsed, &t.ExpiresAt, &t.BrainEnabled); err != nil {
			return nil, err
		}
		// Check expiry
		if t.ExpiresAt != nil {
			exp, err := time.Parse(time.RFC3339, *t.ExpiresAt)
			if err == nil && time.Now().UTC().After(exp) {
				continue // expired
			}
		}
		// Compare hash
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(rawToken)) == nil {
			return &t, nil
		}
	}
	return nil, nil
}

// TouchAPIToken updates the last_used timestamp.
func (db *DB) TouchAPIToken(tokenID int64) {
	db.Exec(`UPDATE api_tokens SET last_used = ? WHERE id = ?`,
		time.Now().UTC().Format(time.RFC3339), tokenID)
}

// ListAPITokens returns all tokens for a user (never exposes the hash).
func (db *DB) ListAPITokens(userID int64) ([]*APIToken, error) {
	rows, err := db.Query(
		`SELECT id, user_id, name, prefix, created_at, last_used, expires_at, brain_enabled
		 FROM api_tokens WHERE user_id = ? ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing api tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*APIToken
	for rows.Next() {
		t := &APIToken{}
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Prefix, &t.CreatedAt, &t.LastUsed, &t.ExpiresAt, &t.BrainEnabled); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, nil
}

// DeleteAPIToken revokes a specific token.
func (db *DB) DeleteAPIToken(userID, tokenID int64) error {
	res, err := db.Exec(`DELETE FROM api_tokens WHERE id = ? AND user_id = ?`, tokenID, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// UpdateTokenBrainEnabled toggles brain_enabled for a specific token.
func (db *DB) UpdateTokenBrainEnabled(userID, tokenID int64, enabled bool) error {
	res, err := db.Exec(
		`UPDATE api_tokens SET brain_enabled = ? WHERE id = ? AND user_id = ?`,
		enabled, tokenID, userID,
	)
	if err != nil {
		return fmt.Errorf("updating brain_enabled: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// hasBrainToken returns true if the user has at least one token with brain_enabled.
func (db *DB) hasBrainToken(userID int64) bool {
	var count int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM api_tokens WHERE user_id = ? AND brain_enabled = ?`, userID, true,
	).Scan(&count)
	return err == nil && count > 0
}
