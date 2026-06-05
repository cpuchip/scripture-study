package main

import (
	"context"
	"encoding/json"
	"net/http"
)

// personaStore is the read surface the HTTP handlers need. *Store satisfies it;
// tests stub it so handlers are testable without a DB.
type personaStore interface {
	ListPersonas(ctx context.Context) ([]Persona, error)
	PersonaBySlug(ctx context.Context, slug string) (*Persona, error)
}

// Server is the persona-host HTTP surface that ai-chattermax and personas talk
// to: /pubkey (token verification key), /personas (roster), and, from PS.5,
// /join (handshake).
type Server struct {
	store personaStore
	key   *KeyMaterial
}

func NewServer(store personaStore, key *KeyMaterial) *Server {
	return &Server{store: store, key: key}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /pubkey", s.handlePubkey)
	mux.HandleFunc("GET /personas", s.handlePersonas)
	return mux
}

// personaView is the public JSON shape of a persona (no internal ids).
type personaView struct {
	Slug        string `json:"slug"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	AgentFamily string `json:"agent_family"`
}

// handlePersonas lists active personas — the roster ai-chattermax can show.
func (s *Server) handlePersonas(w http.ResponseWriter, r *http.Request) {
	personas, err := s.store.ListPersonas(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list personas failed"})
		return
	}
	out := make([]personaView, 0, len(personas))
	for _, p := range personas {
		out = append(out, personaView{
			Slug:        p.Slug,
			DisplayName: p.DisplayName,
			AvatarURL:   p.AvatarURL,
			AgentFamily: p.AgentFamily,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handlePubkey publishes the Ed25519 PUBLIC key (PEM) so ai-chattermax can
// verify persona tokens against it. The private key is never exposed.
func (s *Server) handlePubkey(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/x-pem-file")
	_, _ = w.Write([]byte(s.key.PublicPEM))
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
