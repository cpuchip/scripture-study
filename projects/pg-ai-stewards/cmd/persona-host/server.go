package main

import (
	"encoding/json"
	"net/http"
)

// Server is the persona-host HTTP surface that ai-chattermax and personas talk
// to: /pubkey (token verification key) and, from PS.5, /join (handshake).
type Server struct {
	store *Store
	key   *KeyMaterial
}

func NewServer(store *Store, key *KeyMaterial) *Server {
	return &Server{store: store, key: key}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /pubkey", s.handlePubkey)
	return mux
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
