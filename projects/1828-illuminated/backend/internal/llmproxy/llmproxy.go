// Package llmproxy implements the /api/llm/* surface — BYOK session
// management plus the render endpoint. Phase 1 ships the Config + Service
// shells so the cmd/server build compiles; phase 4 wires the in-memory
// session store, OpenAI-compatible provider, rate-limiting, and the
// rate_limited_by_1828 error shape.
package llmproxy

import (
	"context"
	"net/http"
	"time"
)

// Config carries the resolved environment for the LLM surface. Built in
// cmd/server/config.go from environment variables.
type Config struct {
	Enabled        bool
	BYOKEnabled    bool
	SessionTTL     time.Duration
	SessionSliding bool

	RatePerIPPerMin      int
	RatePerIPPerDay      int
	GlobalTokenCapPerDay int

	MaxTokensDefault   int
	MaxTokensHard      int
	TemperatureDefault float64
	TemperatureHard    float64
	Timeout            time.Duration

	// Optional server-side default. Most production deploys leave these
	// empty and require BYOK from readers.
	ServerProvider string
	ServerBaseURL  string
	ServerAPIKey   string
	ServerModel    string

	// Convenience: when LLM_PROVIDER=opencode-go and the .env carries
	// OPENCODE_GO_API_KEY, we fall back to that key automatically. Lets
	// Michael smoke-test phase 4 without juggling two key variables.
	OpencodeGoAPIKey string
}

type Service struct {
	cfg Config
}

func New(cfg Config) *Service {
	return &Service{cfg: cfg}
}

func (s *Service) Register(mux *http.ServeMux) {
	// Phase 4 will populate: POST/DELETE /api/llm/session, POST /api/llm/render.
	_ = s
	_ = mux
}

// StartJanitor kicks off the background goroutine that evicts expired
// sessions. Phase 1 is a no-op; phase 4 wires the real session map.
func (s *Service) StartJanitor(ctx context.Context) {
	_ = ctx
}
