package db

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"time"
)

// SharedLink represents a short-code link that resolves to public reader parameters.
type SharedLink struct {
	ID        int64     `json:"id"`
	Code      string    `json:"code"`
	UserID    *int64    `json:"user_id,omitempty"`  // nullable — anonymous shares allowed
	SourceID  *int64    `json:"source_id,omitempty"` // optional link to user's source
	Provider  string    `json:"provider"`            // "gh" for GitHub
	Repo      string    `json:"repo"`
	Branch    string    `json:"branch"`
	DocFilter string    `json:"doc_filter"` // include glob
	FilePath  *string   `json:"file_path,omitempty"`
	Hits      int64     `json:"hits"`
	CreatedAt time.Time `json:"created_at"`
}

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// GenerateShortCode creates a cryptographically random base62 code of the given length.
func GenerateShortCode(length int) (string, error) {
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(base62Chars))))
		if err != nil {
			return "", fmt.Errorf("generating random code: %w", err)
		}
		b[i] = base62Chars[n.Int64()]
	}
	return string(b), nil
}

const sharedLinkColumns = `id, code, user_id, source_id, provider, repo, branch, doc_filter, file_path, hits, created_at`

func scanSharedLink(scanner interface{ Scan(...any) error }) (*SharedLink, error) {
	s := &SharedLink{}
	if err := scanner.Scan(&s.ID, &s.Code, &s.UserID, &s.SourceID, &s.Provider, &s.Repo, &s.Branch, &s.DocFilter, &s.FilePath, &s.Hits, &s.CreatedAt); err != nil {
		return nil, err
	}
	return s, nil
}

// CreateSharedLink inserts a new shared link with a generated short code.
// Retries on code collision (unlikely but possible).
func (db *DB) CreateSharedLink(s *SharedLink) error {
	if s.Provider == "" {
		s.Provider = "gh"
	}
	if s.Branch == "" {
		s.Branch = "main"
	}
	if s.DocFilter == "" {
		s.DocFilter = "**/*.md"
	}

	// Try up to 5 times in case of code collision
	for attempt := 0; attempt < 5; attempt++ {
		code, err := GenerateShortCode(7)
		if err != nil {
			return err
		}
		s.Code = code

		id, err := db.InsertReturningID(`
			INSERT INTO shared_links (code, user_id, source_id, provider, repo, branch, doc_filter, file_path)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			s.Code, s.UserID, s.SourceID, s.Provider, s.Repo, s.Branch, s.DocFilter, s.FilePath,
		)
		if err != nil {
			// Unique constraint violation on code — retry with new code
			if attempt < 4 {
				continue
			}
			return fmt.Errorf("inserting shared link after retries: %w", err)
		}
		s.ID = id
		s.CreatedAt = time.Now()
		return nil
	}
	return fmt.Errorf("failed to generate unique code")
}

// ResolveSharedLink looks up a shared link by code and increments the hit counter.
func (db *DB) ResolveSharedLink(code string) (*SharedLink, error) {
	s, err := scanSharedLink(db.QueryRow(`
		SELECT `+sharedLinkColumns+`
		FROM shared_links WHERE code = ?`, code))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("resolving shared link: %w", err)
	}

	// Increment hits asynchronously (fire and forget)
	go func() {
		db.Exec(`UPDATE shared_links SET hits = hits + 1 WHERE id = ?`, s.ID)
	}()

	return s, nil
}

// GetSharedLinkByParams finds an existing shared link matching the parameters.
// Used to avoid duplicate short codes for the same content.
func (db *DB) GetSharedLinkByParams(repo, branch, docFilter, filePath string) (*SharedLink, error) {
	var row *sql.Row
	if filePath == "" {
		row = db.QueryRow(`
			SELECT `+sharedLinkColumns+`
			FROM shared_links
			WHERE repo = ? AND branch = ? AND doc_filter = ? AND file_path IS NULL
			LIMIT 1`, repo, branch, docFilter)
	} else {
		row = db.QueryRow(`
			SELECT `+sharedLinkColumns+`
			FROM shared_links
			WHERE repo = ? AND branch = ? AND doc_filter = ? AND file_path = ?
			LIMIT 1`, repo, branch, docFilter, filePath)
	}

	s, err := scanSharedLink(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("finding shared link: %w", err)
	}
	return s, nil
}
