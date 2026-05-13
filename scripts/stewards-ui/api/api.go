// Package api hosts the HTTP handlers under /api/* for the stewards-ui
// service. Each file in this package owns one logical surface
// (dashboard, studies, work_items, sessions, watchman, bridge, graph,
// search). All handlers receive a shared *Deps with the pgxpool and
// any other shared state.
//
// Convention: handlers return JSON. Errors are wrapped via writeErr()
// with appropriate HTTP status. Successful responses go through
// writeJSON(). Both helpers set Content-Type and never panic on
// encoding failures (encoding errors are logged to stderr).

package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Deps is the shared dependency bundle handed to every handler.
type Deps struct {
	Pool *pgxpool.Pool
}

// Register wires every handler into the supplied mux. main.go calls
// this once at startup.
func Register(mux *http.ServeMux, deps *Deps) {
	mux.HandleFunc("GET /api/dashboard", deps.dashboardHandler)
	deps.registerStudies(mux)
	deps.registerWorkItems(mux)
	deps.registerSessions(mux)
	deps.registerWatchman(mux)
	deps.registerBridge(mux)
	deps.registerNewWork(mux)
	deps.registerGraph(mux)
	deps.registerProviders(mux)
	deps.registerIntents(mux)
	deps.registerCovenants(mux)
	deps.registerLessons(mux)
	deps.registerTrust(mux)
	deps.registerCouncils(mux)
	deps.registerPipelines(mux)
	deps.registerProjects(mux)
	deps.registerAgentProposals(mux)
}

// writeJSON marshals v to JSON, sets the Content-Type header, and
// writes status. Encoding errors are logged but not surfaced to the
// caller — by the time the writer has been touched it's too late to
// change status anyway.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		log.Printf("api: writeJSON encode: %v", err)
	}
}

// writeErr returns a structured JSON error: {"error": "<msg>"}.
func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
