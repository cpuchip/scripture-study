// Command git-mcp is a sandboxed git/gh MCP server for substrate-driven
// repo operations (Phase 3d v1, 2026-05-09).
//
// Each tool wraps a narrow, vetted git or gh invocation. Forbidden ops
// (force-push, reset --hard, branch -D, rebase, tag, raw subcommand
// passthrough) do not exist as tools, so an agent cannot reach them.
// Branch names are constrained to the agent/<pipeline>/<work-item-id>-<slug>
// namespace by an anchored regex; protected branches (main, master,
// release/*) are refused at the tool layer.
//
// Token handling: GITHUB_TOKEN is read from env at startup and never
// passed through tool args. The agent's tool-call context never sees it.
// gh CLI inherits it via env when this process spawns subprocesses.
//
// Workdir: /tmp/stewards-git/<work-item-id>/ per pipeline. Persists
// for inspection — no auto-cleanup. The bridge container's tmp is
// container-local; restart wipes everything (intentional).
//
// Critical discipline (per .github/skills/mcp-server-go/SKILL.md):
//   - All logging MUST go to stderr. Stdout is reserved for JSON-RPC.
//   - The MCP SDK handles handshake, framing, notifications.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const version = "0.1.0"

func main() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("git-mcp: ")
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	var (
		workdirRoot = flag.String("workdir-root", "/tmp/stewards-git",
			"Root directory for per-work-item git workdirs")
		ghBin = flag.String("gh", "gh",
			"Path to gh CLI binary (resolved via $PATH if bare name)")
		gitBin = flag.String("git", "git",
			"Path to git binary (resolved via $PATH if bare name)")
		coAuthorEmail = flag.String("co-author-email", "agents@cpuchip.net",
			"Email used in Co-Authored-By trailer on agent commits")
	)
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := &gitConfig{
		WorkdirRoot:   *workdirRoot,
		GitBin:        *gitBin,
		GhBin:         *ghBin,
		CoAuthorEmail: *coAuthorEmail,
		// Token is read at exec time from os.Getenv so a token rotation
		// without process restart still picks up the new value.
	}

	// Ensure workdir root exists. Per-work-item subdirs are created
	// lazily by git_clone.
	if err := os.MkdirAll(cfg.WorkdirRoot, 0o755); err != nil {
		log.Fatalf("workdir-root mkdir %s: %v", cfg.WorkdirRoot, err)
	}

	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "git-mcp",
		Version: version,
	}, nil)

	registerGitTools(srv, cfg)

	tokenStatus := "unset"
	if os.Getenv("GITHUB_TOKEN") != "" {
		tokenStatus = "set"
	}
	log.Printf("server starting on stdio (mcp protocol); workdir-root=%s gh=%s git=%s GITHUB_TOKEN=%s",
		cfg.WorkdirRoot, cfg.GhBin, cfg.GitBin, tokenStatus)
	if err := srv.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatalf("server.Run: %v", err)
	}
	log.Printf("server stopped cleanly")
}
