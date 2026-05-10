// Command stewards-ui serves the local web UI for pg-ai-stewards
// substrate observability and interaction (Phase 3f v1, 2026-05-09).
//
// Architecture: single Go binary serves both the Vue SPA (embedded
// from frontend/dist/ via embed.FS) and the JSON API at /api/*.
// pgxpool connects to the substrate using STEWARDS_DSN. Single port
// (default 8080); 127.0.0.1 binding by default for local-only safety.
//
// Phase 1 (this commit): foundation only. /healthz returns 200; /
// returns the placeholder Vue page; /api/* returns 501 (handlers come
// in Phase 2+). Validates the multi-stage build, embed.FS pattern,
// and docker-compose service shape.
package main

import (
	"context"
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cpuchip/scripture-study/scripts/stewards-ui/api"
)

//go:embed all:frontend/dist
var distFS embed.FS

func main() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("stewards-ui: ")
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	var (
		dsn  = flag.String("dsn", "", "Postgres DSN (default: $STEWARDS_DSN, then localhost compose port 55433)")
		addr = flag.String("addr", "127.0.0.1:8080", "HTTP listen address")
	)
	flag.Parse()

	if *dsn == "" {
		*dsn = os.Getenv("STEWARDS_DSN")
	}
	if *dsn == "" {
		*dsn = "postgres://stewards:stewards@localhost:55433/stewards?sslmode=disable"
	}

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, *dsn)
	if err != nil {
		log.Fatalf("pgxpool.New: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("pool.Ping: %v", err)
	}
	log.Printf("connected to substrate (dsn=%s)", redactDSN(*dsn))

	mux := http.NewServeMux()

	// /healthz — liveness + DB ping
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			http.Error(w, "db: "+err.Error(), http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("ok"))
	})

	// /api/* handlers — registered by api.Register(). Each endpoint
	// owns its own file under api/. Unknown /api/* paths fall through
	// to mux's default-not-found.
	api.Register(mux, &api.Deps{Pool: pool})

	// SPA static files. Strip the embed prefix so / maps to dist/.
	distSub, err := fs.Sub(distFS, "frontend/dist")
	if err != nil {
		log.Fatalf("fs.Sub frontend/dist: %v", err)
	}
	spa := http.FileServer(http.FS(distSub))

	// Catch-all: serve dist files; for paths that don't exist as
	// files, fall back to index.html so the Vue router can handle
	// client-side routes.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Try the file. If 404, serve index.html.
		f, err := distSub.Open(r.URL.Path[1:])
		if err == nil {
			_ = f.Close()
			spa.ServeHTTP(w, r)
			return
		}
		// Fall back to index.html (root)
		r2 := r.Clone(r.Context())
		r2.URL.Path = "/"
		spa.ServeHTTP(w, r2)
	})

	srv := &http.Server{
		Addr:              *addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()

	log.Printf("listening on http://%s/", *addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe: %v", err)
	}
	log.Printf("server stopped cleanly")
}

func redactDSN(dsn string) string {
	at := -1
	for i := len(dsn) - 1; i >= 0; i-- {
		if dsn[i] == '@' {
			at = i
			break
		}
	}
	if at < 0 {
		return dsn
	}
	colon := -1
	for i := at - 1; i >= 0; i-- {
		if dsn[i] == ':' {
			colon = i
			break
		}
		if dsn[i] == '/' {
			break
		}
	}
	if colon < 0 {
		return dsn
	}
	return dsn[:colon+1] + "***" + dsn[at:]
}
