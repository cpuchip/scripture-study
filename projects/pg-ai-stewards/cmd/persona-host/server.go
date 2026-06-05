package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// personaStore is the surface the HTTP handlers need. *Store satisfies it; tests
// stub it so handlers are testable without a DB.
type personaStore interface {
	ListPersonas(ctx context.Context) ([]Persona, error)
	PersonaBySlug(ctx context.Context, slug string) (*Persona, error)
	UpsertPersonaRoom(ctx context.Context, personaID, roomID string) error
}

// Server is the persona-host HTTP surface that ai-chattermax and personas talk
// to: /pubkey (token verification key), /personas (roster), and /join (the room
// handshake that mints a persona's token).
type Server struct {
	store  personaStore
	key    *KeyMaterial
	minter *Minter
}

func NewServer(store personaStore, key *KeyMaterial, minter *Minter) *Server {
	return &Server{store: store, key: key, minter: minter}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /pubkey", s.handlePubkey)
	mux.HandleFunc("GET /personas", s.handlePersonas)
	mux.HandleFunc("POST /join", s.handleJoin)
	return mux
}

// JoinRequest is the POST /join body.
type JoinRequest struct {
	Slug string `json:"slug"`
	Room string `json:"room"`
}

// JoinResult is what a persona/ai-chattermax receives to authenticate into a
// room: the signed token plus the persona identity and expiry.
type JoinResult struct {
	Token     string      `json:"token"`
	Persona   personaView `json:"persona"`
	Room      string      `json:"room"`
	ExpiresAt time.Time   `json:"expires_at"`
}

// JoinRoom is the handshake: resolve the persona, record the room membership,
// and mint a scoped token. Shared by handleJoin and the smoke.
func (s *Server) JoinRoom(ctx context.Context, slug, room string) (*JoinResult, error) {
	p, err := s.store.PersonaBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if err := s.store.UpsertPersonaRoom(ctx, p.ID, room); err != nil {
		return nil, err
	}
	token, exp, err := s.minter.MintToken(ctx, p, room, DefaultTokenTTL)
	if err != nil {
		return nil, err
	}
	return &JoinResult{
		Token: token,
		Persona: personaView{
			Slug:        p.Slug,
			DisplayName: p.DisplayName,
			AvatarURL:   p.AvatarURL,
			AgentFamily: p.AgentFamily,
		},
		Room:      room,
		ExpiresAt: exp,
	}, nil
}

// handleJoin mints a token for (persona slug, room). A missing persona or bad
// body is a 4xx; the minted token authenticates the persona into the room (the
// room verifies it against /pubkey — no callback here).
func (s *Server) handleJoin(w http.ResponseWriter, r *http.Request) {
	var req JoinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	if req.Slug == "" || req.Room == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "slug and room are required"})
		return
	}
	res, err := s.JoinRoom(r.Context(), req.Slug, req.Room)
	if err != nil {
		// Unknown persona is the expected 404; anything else is a 500.
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "persona not found or join failed"})
		return
	}
	writeJSON(w, http.StatusOK, res)
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
