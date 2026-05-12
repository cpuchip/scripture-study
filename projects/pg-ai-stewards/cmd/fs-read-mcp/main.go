// Command fs-read-mcp is a path-scoped filesystem read MCP server.
//
// Built for Substrate Batch H.1.7 (2026-05-11). The substrate's research
// agent needs to consult prior work — journals, proposals, mind files,
// docs — before doing external research. This server exposes read-only
// filesystem access tightly scoped via --allowed-paths so a misbehaving
// agent can't reach config, secrets, or arbitrary files the bridge
// container happens to have visible.
//
// Discipline:
//   - All logging goes to stderr. Stdout is the JSON-RPC stream.
//   - Every tool call validates the requested path against the
//     allow-list BEFORE any filesystem syscall.
//   - Symlink resolution happens via filepath.EvalSymlinks; after
//     resolution we re-check the allow-list. Symlink escape = reject.
//   - Per-call size cap on fs_read prevents one tool call from
//     blowing the model's context budget.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const version = "0.1.0"

// Sandbox config is read at startup. Mutating it later would require
// re-running the allow-list check on in-flight requests; instead we
// require a restart for scope changes.
type sandbox struct {
	repoRoot     string   // absolute path on the container fs
	allowedGlobs []string // patterns relative to repoRoot (e.g., ".spec/journal/*")
	maxReadBytes int      // per-call cap for fs_read
}

func main() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("fs-read-mcp: ")
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	var (
		repoRoot     string
		allowedFlag  string
		maxReadBytes int
	)
	flag.StringVar(&repoRoot, "repo-root", "/workspace",
		"Absolute path the server treats as the root for all relative paths in tool args.")
	flag.StringVar(&allowedFlag, "allowed-paths", "",
		"Comma-separated glob patterns (repo-root-relative) the server is allowed to read.\nExample: .spec/journal/*,.spec/proposals/*,.mind/*,docs/**")
	flag.IntVar(&maxReadBytes, "max-read-bytes", 50*1024,
		"Per-call cap on fs_read response size in bytes (default 50KB).")
	flag.Parse()

	if allowedFlag == "" {
		log.Fatalf("--allowed-paths is required (got empty)")
	}
	var allowed []string
	for _, p := range strings.Split(allowedFlag, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		// Normalize leading "./" — repo-root-relative patterns don't
		// need it and filepath.Match doesn't strip it.
		p = strings.TrimPrefix(p, "./")
		allowed = append(allowed, p)
	}
	if len(allowed) == 0 {
		log.Fatalf("--allowed-paths parsed to zero non-empty patterns")
	}

	// Resolve repo-root absolute + check it exists. Fail-fast.
	rootInfo, err := os.Stat(repoRoot)
	if err != nil {
		log.Fatalf("repo-root %q does not exist or is unreadable: %v", repoRoot, err)
	}
	if !rootInfo.IsDir() {
		log.Fatalf("repo-root %q is not a directory", repoRoot)
	}

	sb := &sandbox{
		repoRoot:     repoRoot,
		allowedGlobs: allowed,
		maxReadBytes: maxReadBytes,
	}
	log.Printf("sandbox: repo-root=%s allowed=%v max-read=%d", repoRoot, allowed, maxReadBytes)

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "fs-read",
		Version: version,
	}, nil)

	registerTools(srv, sb)

	log.Printf("server starting on stdio (mcp protocol)")
	if err := srv.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatalf("server.Run: %v", err)
	}
	log.Printf("server stopped cleanly")
}
