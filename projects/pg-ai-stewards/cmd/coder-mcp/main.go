// coder-mcp is the substrate's coding capability (substrate-coding-capability
// proposal). CC.1 seeds it with the sandbox-manager + a -smoke mode; CC.2
// grows the MCP tool surface (write/edit/apply_patch/read/glob/grep/shell/lsp)
// onto the same module.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cpuchip/scripture-study/projects/pg-ai-stewards/cmd/coder-mcp/sandbox"
)

func main() {
	smoke := flag.Bool("smoke", false, "Provision a sandbox, print toolchain versions, tear down, and exit (CC.1 smoke).")
	flag.Parse()

	log.SetOutput(os.Stderr)

	if *smoke {
		if err := runSmoke(); err != nil {
			log.Fatalf("coder-mcp smoke FAILED: %v", err)
		}
		return
	}

	// CC.2 will start the MCP server here.
	log.Println("coder-mcp: CC.1 stub (sandbox-manager only). Run with -smoke to test the sandbox.")
}

// runSmoke proves CC.1 end to end: provision → exec toolchain checks → teardown.
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
