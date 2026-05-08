// Command stewards-mcp is the MCP (Model Context Protocol) sidecar for
// the pg-ai-stewards substrate. It connects to Postgres via pgxpool and
// exposes substrate tools to MCP clients (Claude Code, etc.) over stdio.
//
// Phase 3e.1 (2026-05-08): initial version exposes two read-only tools
// over the studies corpus:
//   - study_search — full-text + kinds-filter search (wraps stewards.study_search_text)
//   - study_get    — read a study by slug with line-range pagination (wraps stewards.study_get)
//
// Future phases will add stewards_brain, stewards_work_item, gospel_passthrough,
// and outbound MCP-client capability for consuming gospel-engine-v2.
//
// Critical discipline (per .github/skills/mcp-server-go/SKILL.md):
//   - All logging MUST go to stderr. Stdout is reserved for the JSON-RPC
//     protocol stream — any stray println there corrupts the wire.
//   - The MCP SDK handles the initialize handshake, capability negotiation,
//     newline-delimited JSON-RPC framing, and notification ordering.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Version reported in the initialize handshake's serverInfo.
const version = "0.1.0"

func main() {
	// CRITICAL: pin the default logger to stderr. Anything to stdout
	// (including library logs we don't control) corrupts the protocol
	// stream. We override the package-level default so transitive deps
	// that call log.Print* land on stderr.
	log.SetOutput(os.Stderr)
	log.SetPrefix("stewards-mcp: ")
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	// CLI flags. DSN can also come from STEWARDS_DSN env var (same as
	// stewards-cli) so the .mcp.json config can stay terse.
	var dsn string
	flag.StringVar(&dsn, "dsn", "",
		"Postgres DSN (default: $STEWARDS_DSN, then localhost compose port 55433)")
	flag.Parse()

	if dsn == "" {
		dsn = os.Getenv("STEWARDS_DSN")
	}
	if dsn == "" {
		dsn = "postgres://stewards:stewards@localhost:55433/stewards?sslmode=disable"
	}

	// Root context cancelled on SIGINT/SIGTERM so the server shuts down
	// cleanly when Claude Code closes the stdio pipe.
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Open the pool. ParseConfig + NewWithConfig would let us tune
	// MaxConns etc., but the default (4 * NumCPU) is fine for a
	// read-mostly tool surface.
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("pgxpool.New: %v", err)
	}
	defer pool.Close()

	// Quick connectivity check before declaring ready. Fail-fast on
	// startup is friendlier than the first tool call returning a cryptic
	// connection error.
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("pool.Ping: %v", err)
	}
	log.Printf("connected to substrate (dsn=%s)", redactDSN(dsn))

	// Build the MCP server. Capabilities are auto-declared by the SDK
	// based on what tools/resources/prompts we register.
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "pg-ai-stewards",
		Version: version,
	}, nil)

	// Register tools. Each handler closes over the pool so it can run
	// queries; the pool is already context-aware and goroutine-safe.
	registerStudyTools(srv, pool)
	registerInspectionTools(srv, pool)

	log.Printf("server starting on stdio (mcp protocol)")
	if err := srv.Run(ctx, &mcp.StdioTransport{}); err != nil {
		// Run returns nil on graceful shutdown (ctx cancellation), so
		// any non-nil err here is a real failure.
		log.Fatalf("server.Run: %v", err)
	}
	log.Printf("server stopped cleanly")
}

// redactDSN strips the password component from a Postgres URL so we can
// log the connection target without leaking the secret. Best-effort —
// if the DSN isn't a URL form (e.g. key=value pair list), returns it
// unchanged.
func redactDSN(dsn string) string {
	// postgres://user:password@host:port/db?args
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
			break // hit scheme://, no password component
		}
	}
	if colon < 0 {
		return dsn
	}
	return dsn[:colon+1] + "***" + dsn[at:]
}
