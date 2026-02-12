package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/cpuchip/scripture-study/scripts/becoming/internal/db"
	"github.com/go-chi/chi/v5"
)

// Handlers holds auth-related HTTP handlers.
type Handlers struct {
	DB      *db.DB
	DevMode bool
	Secure  bool         // true = set Secure flag on cookies (HTTPS)
	OAuth   *OAuthConfig // nil = Google sign-in disabled
}

// --- Registration & Login ---

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	User *db.User `json:"user"`
}

// Register handles POST /auth/register.
func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Name = strings.TrimSpace(req.Name)

	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		return
	}
	if len(req.Password) < 8 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters"})
		return
	}

	// Check if email already exists
	existing, err := h.DB.GetUserByEmail(req.Email)
	if err != nil {
		log.Printf("register: error checking email: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if existing != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "email already registered"})
		return
	}

	if req.Name == "" {
		req.Name = strings.Split(req.Email, "@")[0]
	}

	user, err := h.DB.CreateUser(req.Email, req.Password, req.Name)
	if err != nil {
		log.Printf("register: error creating user: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create account"})
		return
	}

	// Create session
	session, err := h.DB.CreateSession(user.ID, r.UserAgent(), clientIP(r))
	if err != nil {
		log.Printf("register: error creating session: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create session"})
		return
	}

	h.setSessionCookie(w, session.ID)
	writeJSON(w, http.StatusCreated, authResponse{User: user})
}

// Login handles POST /auth/login.
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		return
	}

	user, err := h.DB.GetUserByEmail(req.Email)
	if err != nil {
		log.Printf("login: error looking up user: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if user == nil || !db.CheckPassword(user.PasswordHash, req.Password) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
		return
	}

	h.DB.TouchUserLogin(user.ID)

	session, err := h.DB.CreateSession(user.ID, r.UserAgent(), clientIP(r))
	if err != nil {
		log.Printf("login: error creating session: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create session"})
		return
	}

	h.setSessionCookie(w, session.ID)
	writeJSON(w, http.StatusOK, authResponse{User: user})
}

// Logout handles POST /auth/logout.
func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("becoming_session")
	if err == nil && cookie.Value != "" {
		h.DB.DeleteSession(cookie.Value)
	}
	h.clearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

// --- User Profile ---

// Me handles GET /api/me.
func (h *Handlers) Me(w http.ResponseWriter, r *http.Request) {
	userID := UserID(r)
	user, err := h.DB.GetUserByID(userID)
	if err != nil || user == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	writeJSON(w, http.StatusOK, user)
}

// UpdateMe handles PUT /api/me.
func (h *Handlers) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID := UserID(r)
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}
	if err := h.DB.UpdateUserName(userID, req.Name); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update"})
		return
	}
	user, _ := h.DB.GetUserByID(userID)
	writeJSON(w, http.StatusOK, user)
}

// --- API Tokens ---

// ListTokens handles GET /api/tokens.
func (h *Handlers) ListTokens(w http.ResponseWriter, r *http.Request) {
	userID := UserID(r)
	tokens, err := h.DB.ListAPITokens(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list tokens"})
		return
	}
	if tokens == nil {
		tokens = []*db.APIToken{}
	}
	writeJSON(w, http.StatusOK, tokens)
}

// CreateToken handles POST /api/tokens.
func (h *Handlers) CreateToken(w http.ResponseWriter, r *http.Request) {
	userID := UserID(r)
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		req.Name = "API Token"
	}

	token, rawToken, err := h.DB.CreateAPIToken(userID, req.Name)
	if err != nil {
		log.Printf("create token: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create token"})
		return
	}

	// Return the raw token ONCE — user must copy it now
	writeJSON(w, http.StatusCreated, map[string]any{
		"token":     rawToken,
		"id":        token.ID,
		"name":      token.Name,
		"prefix":    token.Prefix,
		"created_at": token.CreatedAt,
	})
}

// DeleteToken handles DELETE /api/tokens/{id}.
func (h *Handlers) DeleteToken(w http.ResponseWriter, r *http.Request) {
	userID := UserID(r)
	idStr := chi.URLParam(r, "id")
	tokenID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid token id"})
		return
	}
	if err := h.DB.DeleteAPIToken(userID, tokenID); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "token not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}

// --- Providers ---

// Providers handles GET /api/auth/providers — tells the frontend which sign-in methods are available.
func (h *Handlers) Providers(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{
		"email":  true,
		"google": h.OAuth != nil,
	})
}

// --- Helpers ---

func (h *Handlers) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "becoming_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.Secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   30 * 24 * 60 * 60, // 30 days
	})
}

func (h *Handlers) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "becoming_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.Secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}
