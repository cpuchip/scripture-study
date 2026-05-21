// Package httpx provides small response helpers and middleware shared
// across the i1828 backend. Centralizing the JSON error envelope here
// keeps every endpoint consistent — the rate-limited-by-1828 shape from
// D-BE-AUTH lives in this same envelope, just with extra fields.
package httpx

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

// ErrorBody is the canonical error envelope. Endpoints that need extra
// context (rate-limit attribution, upstream provider error pass-through)
// add fields by writing the envelope manually rather than calling WriteError.
type ErrorBody struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// WriteJSON marshals v and writes it with the given status. Sets
// Content-Type. Failures log; the response is best-effort by that point.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("[httpx] write json failed: %v", err)
	}
}

// WriteText is the small twin of WriteJSON for plain-text responses
// (/healthz, etc.).
func WriteText(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

// WriteError renders the canonical error envelope at the given status.
func WriteError(w http.ResponseWriter, status int, code, msg string) {
	WriteJSON(w, status, ErrorBody{Error: code, Message: msg})
}

// LoggingMiddleware writes a one-line access log per request. Bare
// stdlib — Dokploy collects stdout, no need for a structured logger yet.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		log.Printf("%s %s %d %s %s",
			r.Method, r.URL.Path, rw.status, time.Since(start), clientIP(r))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

// ClientIP returns the remote IP, honoring X-Forwarded-For when nginx
// proxies to us. We trust nginx because it's a sibling container; users
// can't bypass it to spoof the header.
func ClientIP(r *http.Request) string {
	return clientIP(r)
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if comma := strings.Index(xff, ","); comma >= 0 {
			return strings.TrimSpace(xff[:comma])
		}
		return strings.TrimSpace(xff)
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	// RemoteAddr is host:port; strip the port.
	addr := r.RemoteAddr
	if colon := strings.LastIndex(addr, ":"); colon >= 0 {
		return addr[:colon]
	}
	return addr
}
