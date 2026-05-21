// Package llmproxy implements the /api/llm/* surface — BYOK session
// management + the render endpoint, per llm-proxy.md §VII (D-LP-2
// ratified 2026-05-20).
//
// Storage discipline (the safety property of this design):
//   - Keys never touch disk or DB. Sessions live in an in-memory map
//     keyed by session_id; eviction is by TTL via the janitor goroutine.
//   - Server restart drops all sessions. Readers re-authenticate.
//   - A DB compromise leaks dictionary lookups, not LLM keys.
package llmproxy

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/stuffleberry/i1828/backend/internal/httpx"
)

// Config carries the resolved environment configuration. Built in
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

	// Optional server-side default. Most production deploys leave this
	// empty and require BYOK from readers. When set, anonymous render
	// (no session_id) is allowed subject to D-LP-3 + D-LP-4 caps.
	ServerProvider string
	ServerBaseURL  string
	ServerAPIKey   string
	ServerModel    string

	// Convenience: when ServerProvider=opencode-go and the .env carries
	// OPENCODE_GO_API_KEY, we fall back to that key automatically. Lets
	// Michael smoke-test phase 4 without juggling two key variables.
	OpencodeGoAPIKey string
}

// Service is the HTTP-level LLM bundle.
type Service struct {
	cfg Config

	mu       sync.Mutex
	sessions map[string]*session
	ipBudget map[string]*ipState // per-IP rate-limit + daily cap tracking
	tokens   tokenLedger         // global daily token cap

	httpClient *http.Client
}

type session struct {
	ID        string
	Provider  string
	BaseURL   string
	APIKey    string
	Model     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// ipState tracks per-IP rate-limit windows. The two windows roll
// independently: minute resets every 60s, day resets at UTC midnight.
type ipState struct {
	MinuteWindowStart time.Time
	MinuteCount       int
	DayDate           string // YYYY-MM-DD UTC
	DayCount          int
}

type tokenLedger struct {
	DayDate string // YYYY-MM-DD UTC
	Tokens  int
}

// New constructs the Service.
func New(cfg Config) *Service {
	return &Service{
		cfg:        cfg,
		sessions:   make(map[string]*session),
		ipBudget:   make(map[string]*ipState),
		httpClient: &http.Client{Timeout: cfg.Timeout},
	}
}

// Register attaches all LLM routes to mux.
func (s *Service) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/llm/session", s.handleSessionCreate)
	mux.HandleFunc("DELETE /api/llm/session", s.handleSessionDelete)
	mux.HandleFunc("GET /api/llm/session", s.handleSessionInspect)
	mux.HandleFunc("POST /api/llm/render", s.handleRender)
}

// StartJanitor evicts expired sessions every 60 seconds. Returns
// immediately; the goroutine ends when ctx is cancelled.
func (s *Service) StartJanitor(ctx context.Context) {
	go func() {
		t := time.NewTicker(60 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				s.evictExpired()
			}
		}
	}()
}

func (s *Service) evictExpired() {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, sess := range s.sessions {
		if now.After(sess.ExpiresAt) {
			delete(s.sessions, id)
			log.Printf("[llm] session evicted: id=%s…", safeID(id))
		}
	}
}

// ---- session create -----------------------------------------------

type sessionCreateRequest struct {
	Provider string `json:"provider"`
	BaseURL  string `json:"base_url,omitempty"`
	APIKey   string `json:"api_key"`
	Model    string `json:"model"`
}

type sessionCreateResponse struct {
	SessionID string `json:"session_id"`
	ExpiresAt string `json:"expires_at"`
	Provider  string `json:"provider"`
	Model     string `json:"model"`
}

func (s *Service) handleSessionCreate(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.Enabled {
		httpx.WriteError(w, http.StatusServiceUnavailable, "feature_disabled",
			"LLM proxy is disabled on this deploy")
		return
	}
	if !s.cfg.BYOKEnabled {
		httpx.WriteError(w, http.StatusServiceUnavailable, "byok_disabled",
			"BYOK session minting is disabled on this deploy")
		return
	}
	// Per-IP rate-limit applies to session minting too — defeats
	// session-mint enumeration attacks.
	if denied := s.checkIPRate(httpx.ClientIP(r), w, "session_mint"); denied {
		return
	}

	var req sessionCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "bad_json", err.Error())
		return
	}
	req.Provider = strings.ToLower(strings.TrimSpace(req.Provider))
	req.APIKey = strings.TrimSpace(req.APIKey)
	req.Model = strings.TrimSpace(req.Model)
	req.BaseURL = strings.TrimRight(strings.TrimSpace(req.BaseURL), "/")

	provider, baseURL, err := resolveProvider(req.Provider, req.BaseURL)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_provider", err.Error())
		return
	}
	if req.APIKey == "" {
		httpx.WriteError(w, http.StatusBadRequest, "missing_api_key", "api_key is required")
		return
	}
	if req.Model == "" {
		httpx.WriteError(w, http.StatusBadRequest, "missing_model", "model is required")
		return
	}

	// Probe the key with a minimal /v1/models call. We use /v1/models
	// because all four OpenAI-compat providers support it cheaply. The
	// mock provider skips probe — it's the test-and-demo path that
	// must always mint successfully.
	if provider != "mock" {
		if err := s.probeKey(r.Context(), baseURL, req.APIKey); err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, "key_probe_failed", err.Error())
			return
		}
	}

	id, err := newSessionID()
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "session_mint_failed", err.Error())
		return
	}
	expires := time.Now().Add(s.cfg.SessionTTL)
	sess := &session{
		ID:        id,
		Provider:  provider,
		BaseURL:   baseURL,
		APIKey:    req.APIKey,
		Model:     req.Model,
		ExpiresAt: expires,
		CreatedAt: time.Now(),
	}
	s.mu.Lock()
	s.sessions[id] = sess
	s.mu.Unlock()
	log.Printf("[llm] session minted: id=%s… provider=%s model=%s",
		safeID(id), provider, req.Model)

	// Set the cookie (non-HttpOnly so the frontend can read it).
	http.SetCookie(w, &http.Cookie{
		Name:     "i1828_session",
		Value:    id,
		Path:     "/api/llm",
		Expires:  expires,
		MaxAge:   int(s.cfg.SessionTTL.Seconds()),
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteLaxMode,
		HttpOnly: false,
	})

	httpx.WriteJSON(w, http.StatusOK, sessionCreateResponse{
		SessionID: id,
		ExpiresAt: expires.Format(time.RFC3339),
		Provider:  provider,
		Model:     req.Model,
	})
}

// ---- session delete + inspect ------------------------------------

func (s *Service) handleSessionDelete(w http.ResponseWriter, r *http.Request) {
	id := readSessionID(r)
	if id == "" {
		httpx.WriteError(w, http.StatusBadRequest, "missing_session", "no session cookie")
		return
	}
	s.mu.Lock()
	delete(s.sessions, id)
	s.mu.Unlock()
	// Expire the cookie on the client side too.
	http.SetCookie(w, &http.Cookie{
		Name:     "i1828_session",
		Value:    "",
		Path:     "/api/llm",
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
	})
	log.Printf("[llm] session deleted: id=%s…", safeID(id))
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type sessionInspectResponse struct {
	Active    bool   `json:"active"`
	Provider  string `json:"provider,omitempty"`
	Model     string `json:"model,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

func (s *Service) handleSessionInspect(w http.ResponseWriter, r *http.Request) {
	id := readSessionID(r)
	if id == "" {
		httpx.WriteJSON(w, http.StatusOK, sessionInspectResponse{Active: false})
		return
	}
	s.mu.Lock()
	sess, ok := s.sessions[id]
	s.mu.Unlock()
	if !ok || time.Now().After(sess.ExpiresAt) {
		httpx.WriteJSON(w, http.StatusOK, sessionInspectResponse{Active: false})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, sessionInspectResponse{
		Active:    true,
		Provider:  sess.Provider,
		Model:     sess.Model,
		ExpiresAt: sess.ExpiresAt.Format(time.RFC3339),
	})
}

// ---- render -------------------------------------------------------

type tierWordInput struct {
	Word  string `json:"word"`
	Sense string `json:"sense"`
}

type renderRequest struct {
	VerseText string          `json:"verseText"`
	TierWords []tierWordInput `json:"tierWords"`
	Options   renderOptions   `json:"options"`
}

type renderOptions struct {
	MaxTokens   int     `json:"maxTokens"`
	Temperature float64 `json:"temperature"`
	Stream      bool    `json:"stream"`
}

type renderResponse struct {
	Modernized string         `json:"modernized"`
	PromptUsed string         `json:"promptUsed"`
	Model      string         `json:"model"`
	Provider   string         `json:"provider"`
	DurationMs int64          `json:"durationMs"`
	Usage      map[string]int `json:"usage"`
}

func (s *Service) handleRender(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.Enabled {
		httpx.WriteError(w, http.StatusServiceUnavailable, "feature_disabled",
			"LLM proxy is disabled on this deploy")
		return
	}
	if denied := s.checkIPRate(httpx.ClientIP(r), w, "render"); denied {
		return
	}

	// Resolve the session OR the server-side default.
	provider, baseURL, apiKey, model, err := s.resolveCallerCredentials(r)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "no_active_session",
			err.Error())
		return
	}

	var req renderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "bad_json", err.Error())
		return
	}
	if strings.TrimSpace(req.VerseText) == "" {
		httpx.WriteError(w, http.StatusBadRequest, "missing_verse_text", "verseText is required")
		return
	}
	// Clamp options silently per D-LP-5.
	if req.Options.MaxTokens <= 0 {
		req.Options.MaxTokens = s.cfg.MaxTokensDefault
	}
	if req.Options.MaxTokens > s.cfg.MaxTokensHard {
		req.Options.MaxTokens = s.cfg.MaxTokensHard
	}
	if req.Options.Temperature == 0 {
		req.Options.Temperature = s.cfg.TemperatureDefault
	}
	if req.Options.Temperature > s.cfg.TemperatureHard {
		req.Options.Temperature = s.cfg.TemperatureHard
	}
	if req.Options.Temperature < 0 {
		req.Options.Temperature = 0
	}

	// Global daily token cap (only applies to anonymous / server-default
	// path — BYOK sessions count against the user's own key per D-LP-4).
	isAnon := s.isAnonymousCaller(r)
	if isAnon && !s.checkTokenBudget(req.Options.MaxTokens) {
		s.writeRateLimitBody(w, "global_token_day", 0,
			"1828.ibeco.me throttled this request because the deploy's daily "+
				"token budget for anonymous renders is exhausted. Bring your own key "+
				"in Settings to render with your own provider account.")
		return
	}

	systemPrompt, userPrompt := buildRenderPrompt(req.VerseText, req.TierWords)

	start := time.Now()
	out, usage, err := s.callUpstream(r.Context(), baseURL, apiKey, model, systemPrompt, userPrompt, req.Options)
	if err != nil {
		// Upstream provider errors pass through unchanged with a clear
		// attribution prefix so the reader can distinguish "us" from
		// "their provider" (per D-BE-AUTH).
		httpx.WriteJSON(w, http.StatusBadGateway, map[string]any{
			"error":            "upstream_provider_error",
			"upstream_message": err.Error(),
			"message":          "The upstream LLM provider returned an error. This is not a 1828.ibeco.me throttle.",
		})
		return
	}
	if isAnon {
		total := usage["prompt_tokens"] + usage["completion_tokens"]
		s.consumeTokenBudget(total)
	}

	// Sliding window: extend the session's TTL on use.
	if s.cfg.SessionSliding {
		if id := readSessionID(r); id != "" {
			s.mu.Lock()
			if sess, ok := s.sessions[id]; ok {
				sess.ExpiresAt = time.Now().Add(s.cfg.SessionTTL)
			}
			s.mu.Unlock()
		}
	}

	httpx.WriteJSON(w, http.StatusOK, renderResponse{
		Modernized: out,
		// PromptUsed echoes the rendered user message back so the
		// frontend can show "what we sent" for debugging. The system
		// prompt is the same on every call and not interesting to echo.
		PromptUsed: userPrompt,
		Model:      model,
		Provider:   provider,
		DurationMs: time.Since(start).Milliseconds(),
		Usage:      usage,
	})
}

// resolveCallerCredentials returns the provider/baseURL/key/model the
// render handler should use — either from a valid session, or from the
// server-side default if the deploy configured one and the caller has no
// session, or an error otherwise.
func (s *Service) resolveCallerCredentials(r *http.Request) (provider, baseURL, apiKey, model string, err error) {
	if id := readSessionID(r); id != "" {
		s.mu.Lock()
		sess, ok := s.sessions[id]
		s.mu.Unlock()
		if !ok || time.Now().After(sess.ExpiresAt) {
			return "", "", "", "", errors.New("session expired or unknown; re-authenticate in Settings")
		}
		return sess.Provider, sess.BaseURL, sess.APIKey, sess.Model, nil
	}
	// Anonymous: only allowed if the deploy configured a server-default.
	if s.cfg.ServerProvider == "" {
		return "", "", "", "", errors.New("no active session; mint one at POST /api/llm/session")
	}
	// Mock provider is explicitly anonymous-allowed — it's the canned-
	// response path used for frontend integration tests.
	if s.cfg.ServerProvider == "mock" {
		return "mock", "", "", "mock-model", nil
	}
	provider, baseURL, perr := resolveProvider(s.cfg.ServerProvider, s.cfg.ServerBaseURL)
	if perr != nil {
		return "", "", "", "", perr
	}
	key := s.cfg.ServerAPIKey
	if key == "" && provider == "opencode-go" {
		key = s.cfg.OpencodeGoAPIKey
	}
	if key == "" {
		return "", "", "", "", errors.New("server has no default key; mint a BYOK session at POST /api/llm/session")
	}
	return provider, baseURL, key, s.cfg.ServerModel, nil
}

func (s *Service) isAnonymousCaller(r *http.Request) bool {
	id := readSessionID(r)
	if id == "" {
		return true
	}
	s.mu.Lock()
	_, ok := s.sessions[id]
	s.mu.Unlock()
	return !ok
}

// ---- provider abstraction (OpenAI-compatible) ---------------------

// resolveProvider validates the provider name and returns its canonical
// name + base URL. Reader-supplied base_url wins for providers that need
// it (opencode-go, opencode-zen); the well-known providers (openai,
// openrouter) have fixed URLs the user can override but doesn't need to.
func resolveProvider(name, userBaseURL string) (provider, baseURL string, err error) {
	switch name {
	case "openai":
		base := userBaseURL
		if base == "" {
			base = "https://api.openai.com/v1"
		}
		return "openai", base, nil
	case "openrouter":
		base := userBaseURL
		if base == "" {
			base = "https://openrouter.ai/api/v1"
		}
		return "openrouter", base, nil
	case "opencode-go":
		if userBaseURL == "" {
			return "", "", errors.New("opencode-go requires base_url")
		}
		return "opencode-go", userBaseURL, nil
	case "opencode-zen":
		if userBaseURL == "" {
			return "", "", errors.New("opencode-zen requires base_url")
		}
		return "opencode-zen", userBaseURL, nil
	case "mock":
		return "mock", userBaseURL, nil
	case "":
		return "", "", errors.New("provider is required")
	default:
		return "", "", fmt.Errorf("unknown provider %q (supported: openai, openrouter, opencode-go, opencode-zen, mock)", name)
	}
}

func (s *Service) probeKey(ctx context.Context, baseURL, apiKey string) error {
	// /v1/models is cheap and supported by all four providers we care
	// about. A 200 here means the key is at least syntactically valid.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/models", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("probe call failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	return fmt.Errorf("probe returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
}

func (s *Service) callUpstream(ctx context.Context, baseURL, apiKey, model, systemPrompt, userPrompt string, opts renderOptions) (string, map[string]int, error) {
	if model == "" {
		return "", nil, errors.New("no model configured for session")
	}
	if baseURL == "" {
		// Mock path — mock just echoes the user prompt's first line.
		combined := systemPrompt + "\n\n" + userPrompt
		return mockRender(userPrompt), map[string]int{
			"prompt_tokens":     len(combined) / 4,
			"completion_tokens": 50,
			"total_tokens":      len(combined)/4 + 50,
		}, nil
	}

	// Two-message chat: system carries the "output only" rules so the
	// model stops leaking reasoning into the user-visible content; user
	// carries just the passage + flagged-word table. Empirically this
	// fixes kimi-k2.6's "Let me break this down..." behavior.
	body := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": opts.Temperature,
		"max_tokens":  opts.MaxTokens,
	}
	buf, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/chat/completions", bytes.NewReader(buf))
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", nil, fmt.Errorf("upstream %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	// Usage is parsed into a typed struct (not map[string]int) so that
	// nested-object fields some providers add — `prompt_tokens_details`,
	// `completion_tokens_details`, `reasoning_tokens` as an object, etc. —
	// are tolerated by Go's decoder (unknown struct fields are ignored,
	// but objects can't decode into int map values). Outer signature
	// stays map[string]int so callers don't change.
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", nil, fmt.Errorf("decode upstream: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", nil, errors.New("upstream returned no choices")
	}
	usage := map[string]int{
		"prompt_tokens":     parsed.Usage.PromptTokens,
		"completion_tokens": parsed.Usage.CompletionTokens,
		"total_tokens":      parsed.Usage.TotalTokens,
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), usage, nil
}

func mockRender(prompt string) string {
	return "[mock render] " + strings.TrimSpace(strings.SplitN(prompt, "\n", 2)[0])
}

// ---- prompt -----------------------------------------------------------
//
// Returned as (system, user) so the upstream call can send a two-message
// chat. Splitting instructions out of the user content stops models that
// "think out loud" (e.g. kimi-k2.6) from leaking their reasoning into the
// modernized output. Empirically: a single combined user message produced
// 1500+ tokens of "Let me break this down..." text; the system/user split
// produces just the modernized passage as instructed.

func buildRenderPrompt(verseText string, tier []tierWordInput) (system, user string) {
	wordTable := ""
	for _, t := range tier {
		// Collapse whitespace and clip to 200 chars (same as the frontend).
		sense := strings.Join(strings.Fields(t.Sense), " ")
		if len(sense) > 200 {
			sense = sense[:200]
		}
		wordTable += fmt.Sprintf("- **%s**: %s\n", t.Word, sense)
	}
	if wordTable == "" {
		wordTable = "(no flagged words — render naturally)\n"
	}

	system = `You render scripture passages from KJV / Restoration English into clear modern English, preserving 1828 Webster meanings of words the user flags.

Rules:
1. Render the passage in clear modern English.
2. For each flagged word, replace it with a phrase that captures its 1828 sense as defined by the user. Don't substitute the modern dictionary meaning.
3. Mark each substituted phrase with the original word in square brackets after it, like this: "they tolerated [allowed] their fathers' deeds". This keeps the substitution transparent.
4. Do not add theological interpretation, application, or commentary. Translate the language only.
5. Preserve sentence structure where possible.
6. Output the modernized passage as plain prose. No preamble. No reasoning. No "Let me think..." No explanation of choices. No restating the input. Just the modernized passage.
7. Cap output at 800 tokens. If the passage is longer, modernize until the cap and end with [...continued].`

	user = fmt.Sprintf(`Modernize this passage, preserving the 1828 meanings of the flagged words.

**Passage:**

%s

**Words to preserve in their 1828 sense:**

%s`, verseText, wordTable)
	return system, user
}

// ---- rate limit + token ledger ------------------------------------

func (s *Service) checkIPRate(ip string, w http.ResponseWriter, opType string) (denied bool) {
	if ip == "" {
		ip = "unknown"
	}
	now := time.Now()
	today := now.UTC().Format("2006-01-02")

	s.mu.Lock()
	st, ok := s.ipBudget[ip]
	if !ok {
		st = &ipState{MinuteWindowStart: now, DayDate: today}
		s.ipBudget[ip] = st
	}
	// Reset the minute window if it's elapsed.
	if now.Sub(st.MinuteWindowStart) >= time.Minute {
		st.MinuteWindowStart = now
		st.MinuteCount = 0
	}
	// Reset the day window if the date rolled.
	if st.DayDate != today {
		st.DayDate = today
		st.DayCount = 0
	}
	st.MinuteCount++
	st.DayCount++
	minHit := st.MinuteCount > s.cfg.RatePerIPPerMin
	dayHit := st.DayCount > s.cfg.RatePerIPPerDay
	minRetry := int(time.Minute - now.Sub(st.MinuteWindowStart)) / int(time.Second)
	s.mu.Unlock()

	if dayHit {
		s.writeRateLimitBody(w, "per_ip_day", 86400,
			fmt.Sprintf("1828.ibeco.me throttled this %s request because your IP "+
				"has hit the daily cap (%d). Your provider may still allow more — "+
				"this is our cap, not theirs.", opType, s.cfg.RatePerIPPerDay))
		return true
	}
	if minHit {
		if minRetry < 1 {
			minRetry = 1
		}
		s.writeRateLimitBody(w, "per_ip_minute", minRetry,
			fmt.Sprintf("1828.ibeco.me throttled this %s request because your IP "+
				"has hit the per-minute cap (%d). Retry in %ds.", opType, s.cfg.RatePerIPPerMin, minRetry))
		return true
	}
	return false
}

func (s *Service) writeRateLimitBody(w http.ResponseWriter, limitType string, retryAfter int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if retryAfter > 0 {
		w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
	}
	w.WriteHeader(http.StatusTooManyRequests)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error":               "rate_limited_by_1828",
		"limit_type":          limitType,
		"retry_after_seconds": retryAfter,
		"message":             msg,
	})
}

func (s *Service) checkTokenBudget(reserve int) bool {
	today := time.Now().UTC().Format("2006-01-02")
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.tokens.DayDate != today {
		s.tokens.DayDate = today
		s.tokens.Tokens = 0
	}
	return s.tokens.Tokens+reserve <= s.cfg.GlobalTokenCapPerDay
}

func (s *Service) consumeTokenBudget(actual int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens.Tokens += actual
}

// ---- helpers ------------------------------------------------------

func newSessionID() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	// URL-safe base64; trim padding for short cookie values.
	return strings.TrimRight(base64.RawURLEncoding.EncodeToString(buf), "="), nil
}

// readSessionID prefers the cookie; falls back to Authorization: Bearer
// (so curl-based smoke tests don't need to juggle cookies).
func readSessionID(r *http.Request) string {
	if c, err := r.Cookie("i1828_session"); err == nil && c.Value != "" {
		return c.Value
	}
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}

// safeID returns the first 8 chars of a session id for logging, so we
// never log the full token even at debug level.
func safeID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
