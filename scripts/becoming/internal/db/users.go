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
	res, err := db.Exec(
		`INSERT INTO users (email, password_hash, name, provider) VALUES (?, ?, ?, 'email')`,
		email, hash, name,
	)
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}
	id, _ := res.LastInsertId()
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
		// Update provider info if they're linking a new provider
		if existing.Provider != provider || existing.ProviderID != providerID {
			_, err = db.Exec(
				`UPDATE users SET provider = ?, provider_id = ?, avatar_url = ?, last_login = CURRENT_TIMESTAMP WHERE id = ?`,
				provider, providerID, avatarURL, existing.ID,
			)
			if err != nil {
				return nil, fmt.Errorf("updating user provider: %w", err)
			}
		} else {
			_, _ = db.Exec(`UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = ?`, existing.ID)
		}
		return db.GetUserByID(existing.ID)
	}

	// Create new user
	res, err := db.Exec(
		`INSERT INTO users (email, name, avatar_url, provider, provider_id) VALUES (?, ?, ?, ?, ?)`,
		email, name, avatarURL, provider, providerID,
	)
	if err != nil {
		return nil, fmt.Errorf("creating oauth user: %w", err)
	}
	id, _ := res.LastInsertId()
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
	return u, nil
}

// UpdateUserName updates a user's display name.
func (db *DB) UpdateUserName(userID int64, name string) error {
	_, err := db.Exec(`UPDATE users SET name = ? WHERE id = ?`, name, userID)
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
