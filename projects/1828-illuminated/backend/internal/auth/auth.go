package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type contextKey string

const UserContextKey contextKey = "user"

type User struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Auth struct {
	pool        *pgxpool.Pool
	becomingURL string
	client      *http.Client
}

func New(pool *pgxpool.Pool, becomingURL string) *Auth {
	return &Auth{
		pool:        pool,
		becomingURL: strings.TrimRight(becomingURL, "/"),
		client:      &http.Client{Timeout: 10 * time.Second},
	}
}

// Middleware verifies becoming_session cookie, syncs user info to DB, and puts User in context.
func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("becoming_session")
		if err != nil || cookie.Value == "" {
			next.ServeHTTP(w, r)
			return
		}

		user, err := a.verifySession(r.Context(), cookie.Value)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		// Sync user to 1828 postgres users table
		err = a.syncUser(r.Context(), user)
		if err != nil {
			fmt.Printf("[auth] failed to sync user %d: %v\n", user.ID, err)
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth guards endpoints requiring a valid session.
func (a *Auth) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r.Context())
		if user == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"unauthorized","message":"you must sign in with becoming first"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func GetUser(ctx context.Context) *User {
	if val := ctx.Value(UserContextKey); val != nil {
		if u, ok := val.(*User); ok {
			return u
		}
	}
	return nil
}

func (a *Auth) verifySession(ctx context.Context, sessionToken string) (*User, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", a.becomingURL+"/api/me", nil)
	if err != nil {
		return nil, err
	}
	req.AddCookie(&http.Cookie{
		Name:  "becoming_session",
		Value: sessionToken,
	})

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("becoming returned status %d", resp.StatusCode)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (a *Auth) syncUser(ctx context.Context, u *User) error {
	_, err := a.pool.Exec(ctx, `
		INSERT INTO users (becoming_user_id, email)
		VALUES ($1, $2)
		ON CONFLICT (becoming_user_id) DO UPDATE
		SET email = EXCLUDED.email
	`, u.ID, u.Email)
	return err
}
