// Package auth provides authentication middleware and helpers.
package auth

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/cpuchip/scripture-study/scripts/becoming/internal/db"
)

type contextKey string

const userIDKey contextKey = "userID"

// UserID extracts the authenticated user ID from the request context.
// Returns 0 if not authenticated (should not happen behind AuthRequired).
func UserID(r *http.Request) int64 {
	if v, ok := r.Context().Value(userIDKey).(int64); ok {
		return v
	}
	return 0
}

// Required returns middleware that enforces authentication.
// It checks (in order): session cookie, Bearer token, dev mode fallback.
func Required(database *db.DB, devMode bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var userID int64

			// 1. Session cookie (browser)
			if cookie, err := r.Cookie("becoming_session"); err == nil && cookie.Value != "" {
				if session, err := database.GetSession(cookie.Value); err == nil && session != nil && !session.IsExpired() {
					database.TouchSession(session.ID)
					userID = session.UserID
				}
			}

			// 2. Bearer token (API / MCP)
			if userID == 0 {
				if token := extractBearerToken(r); token != "" {
					if apiToken, err := database.ValidateAPIToken(token); err == nil && apiToken != nil {
						database.TouchAPIToken(apiToken.ID)
						userID = apiToken.UserID
					}
				}
			}

			// 3. Dev mode fallback
			if userID == 0 && devMode {
				userID = 1
			}

			if userID == 0 {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractBearerToken pulls the token from "Authorization: Bearer <token>".
func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return ""
	}
	return strings.TrimSpace(auth[len(prefix):])
}

// Optional returns middleware that sets user context if authenticated,
// but allows unauthenticated requests to proceed (userID = 0).
func Optional(database *db.DB, devMode bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var userID int64

			// 1. Session cookie (browser)
			if cookie, err := r.Cookie("becoming_session"); err == nil && cookie.Value != "" {
				if session, err := database.GetSession(cookie.Value); err == nil && session != nil && !session.IsExpired() {
					database.TouchSession(session.ID)
					userID = session.UserID
				}
			}

			// 2. Bearer token (API / MCP)
			if userID == 0 {
				if token := extractBearerToken(r); token != "" {
					if apiToken, err := database.ValidateAPIToken(token); err == nil && apiToken != nil {
						database.TouchAPIToken(apiToken.ID)
						userID = apiToken.UserID
					}
				}
			}

			// 3. Dev mode fallback
			if userID == 0 && devMode {
				userID = 1
			}

			// Set context even if userID is 0 (anonymous)
			if userID > 0 {
				ctx := context.WithValue(r.Context(), userIDKey, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}

// AdminRequired returns middleware that restricts access to admin users.
// Admin emails are read from the ADMIN_EMAILS environment variable
// (comma-separated, case-insensitive).
// Must be used after Required — assumes userID is already in context.
func AdminRequired(database *db.DB) func(http.Handler) http.Handler {
	raw := os.Getenv("ADMIN_EMAILS")
	allowed := make(map[string]bool)
	for _, email := range strings.Split(raw, ",") {
		email = strings.TrimSpace(strings.ToLower(email))
		if email != "" {
			allowed[email] = true
		}
	}
	if len(allowed) == 0 {
		log.Println("Warning: ADMIN_EMAILS not set — admin endpoints will reject all requests")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := UserID(r)
			if userID == 0 {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			user, err := database.GetUserByID(userID)
			if err != nil || user == nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			if !allowed[strings.ToLower(user.Email)] {
				http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
