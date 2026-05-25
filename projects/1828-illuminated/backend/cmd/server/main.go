// 1828-illuminated backend.
//
// Phase 1: HTTP skeleton with healthz, migration runner, and a single
// entrypoint binary. Phases 2-4 add scripture / dictionary / LLM routes
// behind the same router.
//
// Subcommands:
//
//	(no args)   — run the HTTP server (default)
//	healthcheck — run a one-shot health probe for the docker HEALTHCHECK
//	              directive. Exits 0 if healthy, non-zero otherwise.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/stuffleberry/i1828/backend/internal/auth"
	"github.com/stuffleberry/i1828/backend/internal/dict"
	"github.com/stuffleberry/i1828/backend/internal/httpx"
	"github.com/stuffleberry/i1828/backend/internal/llmproxy"
	"github.com/stuffleberry/i1828/backend/internal/mcp"
	"github.com/stuffleberry/i1828/backend/internal/migrate"
	"github.com/stuffleberry/i1828/backend/internal/scripture"
	"github.com/stuffleberry/i1828/backend/internal/seed"
	"github.com/stuffleberry/i1828/backend/internal/studytree"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "healthcheck":
			runHealthcheck()
			return
		case "help", "-h", "--help":
			fmt.Println("usage: i1828 [healthcheck]")
			return
		}
	}

	if err := runServer(); err != nil {
		log.Fatalf("server exited with error: %v", err)
	}
}

func runServer() error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Printf("i1828 backend starting (listen=%s db_host=%s)", cfg.ListenAddr, redactedHost(cfg.DatabaseURL))

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}
	defer pool.Close()

	// Wait for the DB to come up. docker compose's depends_on:condition:service_healthy
	// usually handles this, but if a sibling Postgres restarts after our backend
	// is already up we want to ride out the gap.
	if err := waitForDB(ctx, pool, 30*time.Second); err != nil {
		return fmt.Errorf("db not reachable: %w", err)
	}

	if err := migrate.Run(ctx, pool); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	// Seeders are idempotent — they no-op when their target tables already
	// hold data, and the heavy 98k 1828 ingest only runs once per fresh DB.
	if err := seed.RunAll(ctx, pool); err != nil {
		return fmt.Errorf("seed: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		// Cheap DB ping so the healthcheck actually proves end-to-end reach.
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			httpx.WriteError(w, http.StatusServiceUnavailable, "db_unreachable", err.Error())
			return
		}
		httpx.WriteText(w, http.StatusOK, "ok\n")
	})

	// Auth service
	authSvc := auth.New(pool, cfg.BecomingURL)

	// Wire user session endpoint
	mux.HandleFunc("GET /api/auth/session", func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUser(r.Context())
		if user == nil {
			httpx.WriteJSON(w, http.StatusOK, map[string]any{"authenticated": false})
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{
			"authenticated": true,
			"user":          user,
		})
	})

	mux.HandleFunc("POST /api/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		cookieDomain := os.Getenv("COOKIE_DOMAIN")
		if cookieDomain == "" {
			cookieDomain = ".ibeco.me"
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "becoming_session",
			Value:    "",
			Path:     "/",
			Domain:   cookieDomain,
			HttpOnly: true,
			MaxAge:   -1,
		})
		// Also clear on host-local
		http.SetCookie(w, &http.Cookie{
			Name:     "becoming_session",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			MaxAge:   -1,
		})
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"authenticated": false})
	})

	// Phase 2 — scripture.
	scriptureSvc := scripture.New(pool)
	scriptureSvc.Register(mux)

	// Phase 3 — dictionary.
	dictSvc := dict.New(pool, cfg.ModernFetchDailyCap)
	dictSvc.Register(mux)

	// Study Tree
	studytreeSvc := studytree.New(pool)
	studytreeSvc.Register(mux, authSvc)

	// MCP Proxy
	mcpSvc := mcp.New(cfg.GospelEngineURL, cfg.GospelEngineToken)
	mcpSvc.Register(mux)

	// Phase 4 — LLM proxy + BYOK sessions.
	llmSvc := llmproxy.New(llmproxy.Config{
		Enabled:              cfg.LLMProxyEnabled,
		BYOKEnabled:          cfg.LLMBYOKEnabled,
		SessionTTL:           cfg.LLMSessionTTL,
		SessionSliding:       cfg.LLMSessionSliding,
		RatePerIPPerMin:      cfg.LLMRatePerIPPerMin,
		RatePerIPPerDay:      cfg.LLMRatePerIPPerDay,
		GlobalTokenCapPerDay: cfg.LLMGlobalTokenCapPerDay,
		MaxTokensDefault:     cfg.LLMMaxTokensDefault,
		MaxTokensHard:        cfg.LLMMaxTokensHard,
		TemperatureDefault:   cfg.LLMTemperatureDefault,
		TemperatureHard:      cfg.LLMTemperatureHard,
		Timeout:              cfg.LLMTimeout,
		ServerProvider:       cfg.LLMProvider,
		ServerBaseURL:        cfg.LLMBaseURL,
		ServerAPIKey:         cfg.LLMAPIKey,
		ServerModel:          cfg.LLMModel,
		OpencodeGoAPIKey:     cfg.OpencodeGoAPIKey,
	})
	llmSvc.Register(mux)
	llmSvc.StartJanitor(ctx)

	srv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           httpx.LoggingMiddleware(authSvc.Middleware(mux)),
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      90 * time.Second, // > LLM_TIMEOUT_SECONDS so proxy responses fit
		IdleTimeout:       120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Printf("shutdown signal received")
	case err := <-errCh:
		return err
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	return srv.Shutdown(shutdownCtx)
}

func waitForDB(ctx context.Context, pool *pgxpool.Pool, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		pingCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
		err := pool.Ping(pingCtx)
		cancel()
		if err == nil {
			return nil
		}
		lastErr = err
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
	return fmt.Errorf("db ping failed after %s: %w", timeout, lastErr)
}

func runHealthcheck() {
	// Used by the Docker HEALTHCHECK directive on the distroless image where
	// wget / curl aren't available. Hits our own listener over loopback.
	url := os.Getenv("HEALTHCHECK_URL")
	if url == "" {
		url = "http://127.0.0.1:8080/api/healthz"
	}
	client := &http.Client{Timeout: 4 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "healthcheck status %d\n", resp.StatusCode)
		os.Exit(1)
	}
	os.Exit(0)
}

// redactedHost extracts host:port from a postgres DSN for logging. Avoids
// echoing the password to stdout.
func redactedHost(dsn string) string {
	// e.g. postgres://i1828:PASS@db:5432/i1828?sslmode=disable
	// We don't bother fully parsing; just find @ and / after it.
	at := -1
	for i, r := range dsn {
		if r == '@' {
			at = i
			break
		}
	}
	if at < 0 {
		return "(unparsed)"
	}
	rest := dsn[at+1:]
	for i, r := range rest {
		if r == '/' || r == '?' {
			return rest[:i]
		}
	}
	return rest
}
