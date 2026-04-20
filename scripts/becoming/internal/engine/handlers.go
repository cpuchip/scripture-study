// HTTP handlers for /api/engine-tokens. Mints, lists, and revokes
// gospel-engine tokens on behalf of the authenticated ibeco.me user.
package engine

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/cpuchip/scripture-study/scripts/becoming/internal/auth"
	"github.com/go-chi/chi/v5"
)

// Handlers wires up engine-token endpoints. If client is nil or unconfigured
// every endpoint returns 503 so the UI can show a helpful message instead of
// crashing.
type Handlers struct {
	Client *Client
}

// externalUserFor returns the canonical engine `external_user` value for an
// ibeco.me user. Pattern: "ibeco:<id>".
func externalUserFor(userID int64) string {
	return fmt.Sprintf("ibeco:%d", userID)
}

// EngineToken is the shape returned to the frontend. We deliberately do not
// expose ExternalUser since it's an implementation detail.
type EngineToken struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Prefix    string `json:"prefix"`
	CreatedAt string `json:"created_at"`
	LastUsed  string `json:"last_used,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"`
	RateLimit int    `json:"rate_limit"`
}

func toEngineToken(t Token) EngineToken {
	out := EngineToken{
		ID:        t.ID,
		Name:      t.Name,
		Prefix:    t.Prefix,
		CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		RateLimit: t.RateLimit,
	}
	if t.LastUsed != nil {
		out.LastUsed = t.LastUsed.Format("2006-01-02T15:04:05Z")
	}
	if t.ExpiresAt != nil {
		out.ExpiresAt = t.ExpiresAt.Format("2006-01-02T15:04:05Z")
	}
	return out
}

// List handles GET /api/engine-tokens.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	if !h.Client.Configured() {
		http.Error(w, "engine integration not configured", http.StatusServiceUnavailable)
		return
	}
	userID := auth.UserID(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	all, err := h.Client.ListTokens()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	want := externalUserFor(userID)
	mine := make([]EngineToken, 0)
	for _, t := range all {
		if t.ExternalUser == want && !t.Revoked {
			mine = append(mine, toEngineToken(t))
		}
	}
	writeJSON(w, http.StatusOK, mine)
}

// Create handles POST /api/engine-tokens. Body: {"name": "..."}.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	if !h.Client.Configured() {
		http.Error(w, "engine integration not configured", http.StatusServiceUnavailable)
		return
	}
	userID := auth.UserID(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	resp, err := h.Client.CreateToken(CreateTokenRequest{
		ExternalUser: externalUserFor(userID),
		Name:         req.Name,
		RateLimit:    600, // sensible per-user default
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"token": toEngineToken(resp.Token),
		"raw":   resp.Raw,
	})
}

// Revoke handles DELETE /api/engine-tokens/{id}. Verifies the token belongs
// to this ibeco user before revoking.
func (h *Handlers) Revoke(w http.ResponseWriter, r *http.Request) {
	if !h.Client.Configured() {
		http.Error(w, "engine integration not configured", http.StatusServiceUnavailable)
		return
	}
	userID := auth.UserID(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	// Verify ownership before revoking — we don't want a malicious user to
	// be able to revoke arbitrary engine tokens by guessing IDs.
	all, err := h.Client.ListTokens()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	want := externalUserFor(userID)
	owned := false
	for _, t := range all {
		if t.ID == id && t.ExternalUser == want {
			owned = true
			break
		}
	}
	if !owned {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if err := h.Client.RevokeToken(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Status handles GET /api/engine-tokens/status. Returns whether the engine
// integration is configured. Used by the UI to decide whether to show the
// section.
func (h *Handlers) Status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"configured": h.Client.Configured(),
		"engine_url": func() string {
			if h.Client == nil {
				return ""
			}
			return h.Client.BaseURL
		}(),
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
