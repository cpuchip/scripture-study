// Command fetch-md-mcp is an MCP server that fetches web pages and
// returns clean markdown for AI agents.
//
// Tools:
//   - fetch_url(url, max_chars?)        — single URL → readability + markdown
//   - fetch_urls(urls[], max_chars?)    — concurrent batch
//   - extract_links(url)                — categorized links from a page
//   - fetch_url_raw(url, max_chars?)    — raw HTML (no conversion)
//
// Phase 1 scope (2026-05-09):
//   - Plain Go HTTP client. No JS rendering. Most docs sites, blogs,
//     READMEs, and Wikipedia work fine. JS-rendered SPAs return whatever
//     the initial HTML payload contains.
//   - Conversion via JohannesKaufmann/html-to-markdown.
//   - Content extraction via go-shiori/go-readability (Mozilla
//     Readability port).
//
// Critical discipline (per .github/skills/mcp-server-go/SKILL.md):
//   - All logging MUST go to stderr. Stdout is reserved for the JSON-RPC
//     protocol stream.
//   - The MCP SDK handles the initialize handshake, capability negotiation,
//     newline-delimited JSON-RPC framing, and notification ordering.
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const version = "0.1.0"

func main() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("fetch-md-mcp: ")
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	var (
		userAgent  = flag.String("user-agent", defaultUserAgent, "HTTP User-Agent header")
		timeoutSec = flag.Int("timeout", 30, "Per-request HTTP timeout in seconds")
		maxBytes   = flag.Int64("max-bytes", 5*1024*1024, "Max response body bytes (0 = unlimited)")
	)
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop()

	client := &http.Client{
		Timeout: time.Duration(*timeoutSec) * time.Second,
	}

	cfg := &fetchConfig{
		HTTPClient: client,
		UserAgent:  *userAgent,
		MaxBytes:   *maxBytes,
	}

	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "fetch-md",
		Version: version,
	}, nil)

	registerFetchTools(srv, cfg)

	log.Printf("server starting on stdio (mcp protocol); ua=%q timeout=%ds max-bytes=%d",
		*userAgent, *timeoutSec, *maxBytes)
	if err := srv.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatalf("server.Run: %v", err)
	}
	log.Printf("server stopped cleanly")
}

const defaultUserAgent = "fetch-md-mcp/0.1 (+https://github.com/cpuchip/scripture-study)"
