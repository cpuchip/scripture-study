package db

import (
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a registered user.
type User struct {
	ID           int64  `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"` // never serialize
	Name         string `json:"name"`
	AvatarURL    string `json:"avatar_url"`
	Provider     string `json:"provider"`
	ProviderID   string `json:"-"`
	CreatedAt    string `json:"created_at"`
	LastLogin    string `json:"last_login"`

	// Computed fields — tell the frontend what auth methods are available
	HasPassword  bool `json:"has_password"`
	GoogleLinked bool `json:"google_linked"`
}

// computeAuthFields sets has_password and google_linked from internal state.
func (u *User) computeAuthFields() {
	u.HasPassword = u.PasswordHash != ""
	u.GoogleLinked = u.ProviderID != ""
}

const bcryptCost = 12

// HashPassword hashes a plaintext password with bcrypt.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}
	return string(hash), nil
}

// CheckPassword compares a plaintext password against a bcrypt hash.
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// CreateUser creates a new user with email/password.
func (db *DB) CreateUser(email, password, name string) (*User, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	id, err := db.InsertReturningID(
		`INSERT INTO users (email, password_hash, name, provider) VALUES (?, ?, ?, 'email')`,
		email, hash, name,
	)
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}
	return db.GetUserByID(id)
}

// CreateOAuthUser creates or finds a user from an OAuth provider.
func (db *DB) CreateOAuthUser(email, name, avatarURL, provider, providerID string) (*User, error) {
	// Check if user already exists with this email
	existing, err := db.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		// Link Google to existing account — preserve original provider and password_hash.
		// Only update provider_id and avatar_url (don't overwrite provider).
		if existing.ProviderID != providerID {
			_, err = db.Exec(
				`UPDATE users SET provider_id = ?, avatar_url = ?, last_login = CURRENT_TIMESTAMP WHERE id = ?`,
				providerID, avatarURL, existing.ID,
			)
			if err != nil {
				return nil, fmt.Errorf("linking google account: %w", err)
			}
		} else {
			_, _ = db.Exec(`UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = ?`, existing.ID)
		}
		return db.GetUserByID(existing.ID)
	}

	// Create new user
	id, err := db.InsertReturningID(
		`INSERT INTO users (email, name, avatar_url, provider, provider_id) VALUES (?, ?, ?, ?, ?)`,
		email, name, avatarURL, provider, providerID,
	)
	if err != nil {
		return nil, fmt.Errorf("creating oauth user: %w", err)
	}
	return db.GetUserByID(id)
}

// GetUserByID returns a user by ID.
func (db *DB) GetUserByID(id int64) (*User, error) {
	u := &User{}
	err := db.QueryRow(
		`SELECT id, email, password_hash, name, avatar_url, provider, provider_id, created_at, last_login FROM users WHERE id = ?`,
		id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.AvatarURL, &u.Provider, &u.ProviderID, &u.CreatedAt, &u.LastLogin)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting user by id: %w", err)
	}
	u.computeAuthFields()
	return u, nil
}

// GetUserByEmail returns a user by email.
func (db *DB) GetUserByEmail(email string) (*User, error) {
	u := &User{}
	err := db.QueryRow(
		`SELECT id, email, password_hash, name, avatar_url, provider, provider_id, created_at, last_login FROM users WHERE email = ?`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.AvatarURL, &u.Provider, &u.ProviderID, &u.CreatedAt, &u.LastLogin)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting user by email: %w", err)
	}
	u.computeAuthFields()
	return u, nil
}

// UpdateUserName updates a user's display name.
func (db *DB) UpdateUserName(userID int64, name string) error {
	_, err := db.Exec(`UPDATE users SET name = ? WHERE id = ?`, name, userID)
	return err
}

// UpdateUserPassword updates a user's password hash.
func (db *DB) UpdateUserPassword(userID int64, newHash string) error {
	_, err := db.Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, newHash, userID)
	return err
}

// SetPassword sets a password for a user who doesn't have one (e.g., Google-only user).
func (db *DB) SetPassword(userID int64, newHash string) error {
	_, err := db.Exec(
		`UPDATE users SET password_hash = ?, provider = 'email' WHERE id = ? AND password_hash = ''`,
		newHash, userID,
	)
	return err
}

// UnlinkGoogle removes the Google provider link from a user account.
// Only allowed if user has a password (so they can still log in).
func (db *DB) UnlinkGoogle(userID int64) error {
	_, err := db.Exec(
		`UPDATE users SET provider_id = '', avatar_url = '' WHERE id = ? AND password_hash != ''`,
		userID,
	)
	return err
}

// TouchUserLogin updates the last_login timestamp.
func (db *DB) TouchUserLogin(userID int64) {
	db.Exec(`UPDATE users SET last_login = ? WHERE id = ?`, time.Now().UTC().Format(time.RFC3339), userID)
}

// DeleteUser deletes a user and all their data (cascades via ON DELETE CASCADE for sessions/tokens).
func (db *DB) DeleteUser(userID int64) error {
	_, err := db.Exec(`DELETE FROM users WHERE id = ?`, userID)
	return err
}

// DeleteUserAndData deletes a user and ALL associated data in the correct order
// to satisfy foreign key constraints. This handles tables that lack ON DELETE CASCADE.
func (db *DB) DeleteUserAndData(userID int64) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback()

	queries := []string{
		// Junction tables first (reference practices/tasks which reference user)
		`DELETE FROM practice_pillars WHERE practice_id IN (SELECT id FROM practices WHERE user_id = ?)`,
		`DELETE FROM task_pillars WHERE task_id IN (SELECT id FROM tasks WHERE user_id = ?)`,
		// Logs cascade from practices, but delete explicitly to be safe
		`DELETE FROM practice_logs WHERE practice_id IN (SELECT id FROM practices WHERE user_id = ?)`,
		// Tables with direct user_id
		`DELETE FROM notes WHERE user_id = ?`,
		`DELETE FROM reflections WHERE user_id = ?`,
		`DELETE FROM prompts WHERE user_id = ?`,
		`DELETE FROM pillars WHERE user_id = ?`,
		`DELETE FROM practices WHERE user_id = ?`,
		`DELETE FROM tasks WHERE user_id = ?`,
		// User record last (sessions + api_tokens cascade)
		`DELETE FROM users WHERE id = ?`,
	}

	for _, q := range queries {
		if _, err := tx.Exec(q, userID); err != nil {
			return fmt.Errorf("deleting user data: %w", err)
		}
	}

	return tx.Commit()
}

// EnsureDefaultUser creates user_id=1 if it doesn't exist (for dev mode / migration).
func (db *DB) EnsureDefaultUser() (*User, error) {
	u, err := db.GetUserByID(1)
	if err != nil {
		return nil, err
	}
	if u != nil {
		return u, nil
	}
	// Create a default user for development / migration
	_, err = db.Exec(
		`INSERT INTO users (id, email, name, provider) VALUES (1, 'dev@becoming.local', 'Developer', 'email')`,
	)
	if err != nil {
		return nil, fmt.Errorf("creating default user: %w", err)
	}
	return db.GetUserByID(1)
}

// UserExport contains all user data for export.
type UserExport struct {
	ExportedAt  string         `json:"exported_at"`
	User        *UserProfile   `json:"user"`
	Practices   []*Practice    `json:"practices"`
	Logs        []*PracticeLog `json:"logs"`
	Tasks       []*Task        `json:"tasks"`
	Notes       []*Note        `json:"notes"`
	Prompts     []*Prompt      `json:"prompts"`
	Reflections []*Reflection  `json:"reflections"`
	Pillars     []*Pillar      `json:"pillars"`
}

// UserProfile is a safe-for-export user profile (no password hash).
type UserProfile struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Provider  string `json:"provider"`
	CreatedAt string `json:"created_at"`
	LastLogin string `json:"last_login"`
}

// ExportUserData returns all data belonging to a user.
func (db *DB) ExportUserData(userID int64) (*UserExport, error) {
	user, err := db.GetUserByID(userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found")
	}

	export := &UserExport{
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		User: &UserProfile{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			Provider:  user.Provider,
			CreatedAt: user.CreatedAt,
			LastLogin: user.LastLogin,
		},
	}

	// Practices (all, including inactive)
	export.Practices, err = db.ListPractices(userID, "", false)
	if err != nil {
		return nil, fmt.Errorf("exporting practices: %w", err)
	}

	// All logs for all practices
	rows, err := db.Query(`
		SELECT pl.id, pl.practice_id, pl.logged_at, pl.date, pl.quality, pl.value,
		       pl.sets, pl.reps, pl.duration_s, pl.notes, pl.next_review
		FROM practice_logs pl
		JOIN practices p ON pl.practice_id = p.id
		WHERE p.user_id = ?
		ORDER BY pl.logged_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("exporting logs: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		l := &PracticeLog{}
		if err := rows.Scan(&l.ID, &l.PracticeID, &l.LoggedAt, &l.Date, &l.Quality,
			&l.Value, &l.Sets, &l.Reps, &l.DurationS, &l.Notes, &l.NextReview); err != nil {
			return nil, err
		}
		export.Logs = append(export.Logs, l)
	}

	// Tasks (all statuses)
	export.Tasks, err = db.ListTasks(userID, "")
	if err != nil {
		return nil, fmt.Errorf("exporting tasks: %w", err)
	}

	// Notes (all)
	export.Notes, err = db.ListNotes(userID, nil, nil, nil, false)
	if err != nil {
		return nil, fmt.Errorf("exporting notes: %w", err)
	}

	// Prompts (all)
	export.Prompts, err = db.ListPrompts(userID, false)
	if err != nil {
		return nil, fmt.Errorf("exporting prompts: %w", err)
	}

	// Reflections (all)
	export.Reflections, err = db.ListReflections(userID, "", "")
	if err != nil {
		return nil, fmt.Errorf("exporting reflections: %w", err)
	}

	// Pillars (flat)
	export.Pillars, err = db.ListPillars(userID)
	if err != nil {
		return nil, fmt.Errorf("exporting pillars: %w", err)
	}

	return export, nil
}
