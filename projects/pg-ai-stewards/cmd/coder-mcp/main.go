// coder-mcp is the substrate's coding capability (substrate-coding-capability
// proposal). It runs as a stdio MCP server (CC.2) exposing the sandbox + coding
// tool surface; the sandbox-manager (CC.1) spawns hardened coder-runtime
// containers against the host docker daemon.
//
// Critical discipline (.github/skills/mcp-server-go): all logging to stderr;
// stdout is reserved for JSON-RPC.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/cpuchip/scripture-study/projects/pg-ai-stewards/cmd/coder-mcp/sandbox"
)

const version = "0.2.0"

func main() {
	smoke := flag.Bool("smoke", false, "Provision a sandbox, print toolchain versions, tear down, and exit (CC.1 smoke).")
	flag.Parse()

	log.SetOutput(os.Stderr)
	log.SetPrefix("coder-mcp: ")
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	if *smoke {
		if err := runSmoke(); err != nil {
			log.Fatalf("smoke FAILED: %v", err)
		}
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mgr := sandbox.New()
	srv := mcp.NewServer(&mcp.Implementation{Name: "coder-mcp", Version: version}, nil)
	registerCoderTools(srv, mgr)

	// CC.6: best-effort reap of stale sandboxes (>2h) on startup. The bridge
	// spawns coder-mcp periodically, so this sweeps leaked/abandoned sandboxes
	// without a long-lived daemon. In-use sandboxes (<2h) are untouched.
	if removed, rerr := mgr.ReapSandboxes(ctx, 2*time.Hour); rerr != nil {
		log.Printf("startup reap: %v", rerr)
	} else if len(removed) > 0 {
		log.Printf("startup reap: removed %d stale sandbox(es): %v", len(removed), removed)
	}

	log.Printf("server starting on stdio (mcp protocol); runtime image=%s", mgr.Image)
	if err := srv.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatalf("server.Run: %v", err)
	}
	log.Printf("server stopped cleanly")
}

// runSmoke proves the sandbox end to end: provision → exec toolchain checks → teardown.
func runSmoke() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	m := sandbox.New()
	const wi = "smoke"

	fmt.Printf("coder-mcp smoke: provisioning sandbox (image=%s, network=on)…\n", m.Image)
	if err := m.Provision(ctx, wi, sandbox.NetOn); err != nil {
		return err
	}
	defer func() {
		if err := m.Teardown(ctx, wi); err != nil {
			log.Printf("warning: teardown: %v", err)
		} else {
			fmt.Println("coder-mcp smoke: torn down.")
		}
	}()

	checks := "echo '--- toolchains ---'; go version; node --version; npm --version; python3 --version; " +
		"echo '--- LSP servers ---'; gopls version; typescript-language-server --version; pyright --version; " +
		"echo '--- write+build smoke ---'; mkdir -p /tmp/t && cd /tmp/t && " +
		"printf 'package main\\nimport \"fmt\"\\nfunc main(){fmt.Println(\"hello from the sandbox\")}\\n' > main.go && " +
		"go mod init t >/dev/null 2>&1 && go run main.go"
	res, err := m.Exec(ctx, wi, checks)
	if err != nil {
		return err
	}
	fmt.Print(res.Output)
	if res.ExitCode != 0 {
		return fmt.Errorf("smoke checks exited %d", res.ExitCode)
	}
	fmt.Println("coder-mcp smoke: PASS")
	return nil
}
