// MCP tool surface for coder-mcp (substrate-coding-capability CC.2). Each tool
// operates on a named sandbox (the work_item id; the bridge dispatch carries no
// implicit context, so the sandbox id is an explicit argument — the code-write
// pipeline keys it to the work_item). Tools modeled on opencode's surface
// (write/edit/apply_patch/read/glob/grep/shell), plus sandbox lifecycle.
//
// File paths are resolved relative to /work (the project root in the sandbox);
// absolute paths are allowed but ".." escape above /work is refused.
package main

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/cpuchip/scripture-study/projects/pg-ai-stewards/cmd/coder-mcp/sandbox"
)

const workRoot = "/work"

func registerCoderTools(srv *mcp.Server, mgr *sandbox.Manager) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "coder_sandbox_start",
		Description: "Start an isolated, hardened sandbox container for a work_item " +
			"(Go + Node/TS + Python + LSP). Idempotent — replaces any existing sandbox " +
			"of the same id. Network is on by default (for go mod / npm / pip); set " +
			"offline=true to cut egress. Call coder_sandbox_stop when done.",
	}, makeSandboxStart(mgr))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "coder_sandbox_stop",
		Description: "Stop and remove a work_item's sandbox (discards its filesystem).",
	}, makeSandboxStop(mgr))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "coder_write",
		Description: "Write a file in the sandbox (creates parent dirs, overwrites if present). Path is relative to /work.",
	}, makeWrite(mgr))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "coder_read",
		Description: "Read a file from the sandbox. Path is relative to /work.",
	}, makeRead(mgr))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "coder_edit",
		Description: "Replace an exact string in a sandbox file. old_string must appear " +
			"exactly once (unless replace_all=true). Path is relative to /work.",
	}, makeEdit(mgr))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "coder_apply_patch",
		Description: "Apply a unified diff to the sandbox working tree via `git apply` (run from /work).",
	}, makeApplyPatch(mgr))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "coder_shell",
		Description: "Run a shell command in the sandbox (login bash, cwd /work) — build, test, run, " +
			"install packages. Returns combined output + exit code. This is the ground-truth gate: " +
			"`go build`, `go test`, `npm test`, etc.",
	}, makeShell(mgr))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "coder_glob",
		Description: "List files in the sandbox matching a glob (e.g. **/*.go), relative to /work.",
	}, makeGlob(mgr))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "coder_grep",
		Description: "Search file contents in the sandbox (grep -rn). Optional path scopes the search (relative to /work).",
	}, makeGrep(mgr))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "coder_lsp",
		Description: "Get type/compile diagnostics for a file using the language's checker " +
			"(gopls for Go, tsc for TS/JS, pyright for Python — detected by extension). " +
			"`clean=true` means no diagnostics. Faster feedback than a full build for catching errors mid-edit.",
	}, makeLsp(mgr))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "coder_deploy",
		Description: "Deploy a built artifact: run run_command as a background service in the sandbox " +
			"(the sandbox IS its docker sidecar), wait, then healthcheck http://localhost:<port><health_path>. " +
			"Returns healthy + the healthcheck result + the service log tail. The actual deploy step is gated " +
			"by the always-escalate Hinge in the code-deploy pipeline — a human ratifies before this runs.",
	}, makeDeploy(mgr))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "coder_sandbox_list",
		Description: "List all coder sandboxes (name, work_item, age in minutes) — visibility into what's running.",
	}, makeSandboxList(mgr))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "coder_sandbox_reap",
		Description: "Remove coder sandboxes older than max_age_minutes (default 120) — the reaper for leaked/abandoned sandboxes. Returns the names removed.",
	}, makeSandboxReap(mgr))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "coder_commit",
		Description: "Stage all changes in the sandbox's repo worktree onto a branch (created if absent) and commit. " +
			"Local op — no token. Returns the SHA + branch. Repo-mode sandboxes only (start with repo=).",
	}, makeCommit(mgr))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "coder_push",
		Description: "Push the branch to origin. Runs bridge-side with the GitHub token (never in the sandbox); " +
			"refuses protected branches (main/master/release/*).",
	}, makePush(mgr))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "coder_open_pr",
		Description: "Open a pull request (gh) for the pushed branch — base defaults to main. Returns the PR URL. " +
			"Bridge-side, token from env. The human reviews + merges (the Hinge).",
	}, makeOpenPR(mgr))
}

// resolvePath joins a user path onto /work, refusing escapes above it.
func resolvePath(p string) (string, error) {
	if p == "" {
		return "", fmt.Errorf("path is required")
	}
	var full string
	if path.IsAbs(p) {
		full = path.Clean(p)
	} else {
		full = path.Clean(path.Join(workRoot, p))
	}
	if full != workRoot && !strings.HasPrefix(full, workRoot+"/") {
		return "", fmt.Errorf("path %q escapes the sandbox work root", p)
	}
	return full, nil
}

func errResult(format string, a ...any) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(format, a...)}},
	}
}

// --- sandbox lifecycle ---

type startInput struct {
	Sandbox string `json:"sandbox" jsonschema:"Sandbox id (the work_item id)"`
	Offline bool   `json:"offline,omitempty" jsonschema:"Cut network egress (default false = on, for package pulls)"`
	Repo    string `json:"repo,omitempty" jsonschema:"Optional git repo URL to clone into the sandbox (repo-mode, CV2.1). Must be allow-listed. The repo lands at /work."`
	Branch  string `json:"branch,omitempty" jsonschema:"Optional branch to clone (repo-mode)"`
}
type startOutput struct {
	Sandbox string `json:"sandbox"`
	Network string `json:"network"`
	Repo    string `json:"repo,omitempty"`
}

func makeSandboxStart(mgr *sandbox.Manager) func(context.Context, *mcp.CallToolRequest, startInput) (*mcp.CallToolResult, startOutput, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in startInput) (*mcp.CallToolResult, startOutput, error) {
		if strings.TrimSpace(in.Sandbox) == "" {
			return errResult("sandbox is required"), startOutput{}, nil
		}
		// Reuse an existing sandbox rather than wiping it — the revise loop
		// (verify-fail → re-implement) must keep the in-progress work.
		if exists, err := mgr.Exists(ctx, in.Sandbox); err != nil {
			return errResult("%v", err), startOutput{}, nil
		} else if exists {
			return nil, startOutput{Sandbox: in.Sandbox, Network: "reused"}, nil
		}
		net := sandbox.NetOn
		if in.Offline {
			net = sandbox.NetOff
		}
		// Repo-mode (CV2.1): the bridge clones the allow-listed repo into the
		// shared worktree first (token never in the sandbox), then the sandbox
		// mounts that worktree at /work.
		if strings.TrimSpace(in.Repo) != "" {
			if err := mgr.CloneRepo(ctx, in.Sandbox, in.Repo, in.Branch); err != nil {
				return errResult("%v", err), startOutput{}, nil
			}
			if err := mgr.Provision(ctx, in.Sandbox, net, true); err != nil {
				return errResult("%v", err), startOutput{}, nil
			}
			return nil, startOutput{Sandbox: in.Sandbox, Network: string(net), Repo: in.Repo}, nil
		}
		// No repo passed. If a worktree already exists for this sandbox (e.g. an
		// earlier clone stage cloned it, then the container was reaped or the
		// bridge restarted mid-pipeline), re-mount it so implement/verify/pr keep
		// operating on the cloned repo — never silently fall back to an ephemeral
		// /work that drops the clone. A fresh sandbox with no worktree stays
		// ephemeral (v1 code-write behavior).
		if err := mgr.Provision(ctx, in.Sandbox, net, mgr.HasWorktree(in.Sandbox)); err != nil {
			return errResult("%v", err), startOutput{}, nil
		}
		return nil, startOutput{Sandbox: in.Sandbox, Network: string(net)}, nil
	}
}

type stopInput struct {
	Sandbox string `json:"sandbox" jsonschema:"Sandbox id (the work_item id)"`
}
type stopOutput struct {
	Stopped string `json:"stopped"`
}

func makeSandboxStop(mgr *sandbox.Manager) func(context.Context, *mcp.CallToolRequest, stopInput) (*mcp.CallToolResult, stopOutput, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in stopInput) (*mcp.CallToolResult, stopOutput, error) {
		if err := mgr.Teardown(ctx, in.Sandbox); err != nil {
			return errResult("%v", err), stopOutput{}, nil
		}
		return nil, stopOutput{Stopped: in.Sandbox}, nil
	}
}

// --- file ops ---

type writeInput struct {
	Sandbox string `json:"sandbox" jsonschema:"Sandbox id"`
	Path    string `json:"path"    jsonschema:"File path relative to /work"`
	Content string `json:"content" jsonschema:"File content"`
}
type writeOutput struct {
	Path  string `json:"path"`
	Bytes int    `json:"bytes"`
}

func makeWrite(mgr *sandbox.Manager) func(context.Context, *mcp.CallToolRequest, writeInput) (*mcp.CallToolResult, writeOutput, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in writeInput) (*mcp.CallToolResult, writeOutput, error) {
		full, err := resolvePath(in.Path)
		if err != nil {
			return errResult("%v", err), writeOutput{}, nil
		}
		if err := mgr.WriteFile(ctx, in.Sandbox, full, in.Content); err != nil {
			return errResult("%v", err), writeOutput{}, nil
		}
		return nil, writeOutput{Path: full, Bytes: len(in.Content)}, nil
	}
}

type readInput struct {
	Sandbox string `json:"sandbox" jsonschema:"Sandbox id"`
	Path    string `json:"path"    jsonschema:"File path relative to /work"`
}
type readOutput struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func makeRead(mgr *sandbox.Manager) func(context.Context, *mcp.CallToolRequest, readInput) (*mcp.CallToolResult, readOutput, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in readInput) (*mcp.CallToolResult, readOutput, error) {
		full, err := resolvePath(in.Path)
		if err != nil {
			return errResult("%v", err), readOutput{}, nil
		}
		content, err := mgr.ReadFile(ctx, in.Sandbox, full)
		if err != nil {
			return errResult("%v", err), readOutput{}, nil
		}
		return nil, readOutput{Path: full, Content: content}, nil
	}
}

type editInput struct {
	Sandbox    string `json:"sandbox"     jsonschema:"Sandbox id"`
	Path       string `json:"path"        jsonschema:"File path relative to /work"`
	OldString  string `json:"old_string"  jsonschema:"Exact text to replace"`
	NewString  string `json:"new_string"  jsonschema:"Replacement text"`
	ReplaceAll bool   `json:"replace_all,omitempty" jsonschema:"Replace all occurrences (default false = require exactly one)"`
}
type editOutput struct {
	Path        string `json:"path"`
	Replacements int   `json:"replacements"`
}

func makeEdit(mgr *sandbox.Manager) func(context.Context, *mcp.CallToolRequest, editInput) (*mcp.CallToolResult, editOutput, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in editInput) (*mcp.CallToolResult, editOutput, error) {
		full, err := resolvePath(in.Path)
		if err != nil {
			return errResult("%v", err), editOutput{}, nil
		}
		if in.OldString == "" {
			return errResult("old_string is required"), editOutput{}, nil
		}
		content, err := mgr.ReadFile(ctx, in.Sandbox, full)
		if err != nil {
			return errResult("%v", err), editOutput{}, nil
		}
		n := strings.Count(content, in.OldString)
		if n == 0 {
			return errResult("old_string not found in %s", full), editOutput{}, nil
		}
		if n > 1 && !in.ReplaceAll {
			return errResult("old_string appears %d times in %s; set replace_all=true or make it unique", n, full), editOutput{}, nil
		}
		var updated string
		if in.ReplaceAll {
			updated = strings.ReplaceAll(content, in.OldString, in.NewString)
		} else {
			updated = strings.Replace(content, in.OldString, in.NewString, 1)
			n = 1
		}
		if err := mgr.WriteFile(ctx, in.Sandbox, full, updated); err != nil {
			return errResult("%v", err), editOutput{}, nil
		}
		return nil, editOutput{Path: full, Replacements: n}, nil
	}
}

type patchInput struct {
	Sandbox string `json:"sandbox" jsonschema:"Sandbox id"`
	Diff    string `json:"diff"    jsonschema:"Unified diff to apply with git apply, from /work"`
}
type patchOutput struct {
	Applied bool   `json:"applied"`
	Output  string `json:"output,omitempty"`
}

func makeApplyPatch(mgr *sandbox.Manager) func(context.Context, *mcp.CallToolRequest, patchInput) (*mcp.CallToolResult, patchOutput, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in patchInput) (*mcp.CallToolResult, patchOutput, error) {
		if strings.TrimSpace(in.Diff) == "" {
			return errResult("diff is required"), patchOutput{}, nil
		}
		const patchPath = "/tmp/coder-mcp.patch"
		if err := mgr.WriteFile(ctx, in.Sandbox, patchPath, in.Diff); err != nil {
			return errResult("%v", err), patchOutput{}, nil
		}
		res, err := mgr.Exec(ctx, in.Sandbox, "cd "+workRoot+" && git apply "+patchPath)
		if err != nil {
			return errResult("%v", err), patchOutput{}, nil
		}
		if res.ExitCode != 0 {
			return errResult("git apply failed (exit %d):\n%s", res.ExitCode, res.Output), patchOutput{}, nil
		}
		return nil, patchOutput{Applied: true, Output: res.Output}, nil
	}
}

// --- shell + search ---

type shellInput struct {
	Sandbox string `json:"sandbox" jsonschema:"Sandbox id"`
	Command string `json:"command" jsonschema:"Shell command (login bash, cwd /work)"`
}
type shellOutput struct {
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code"`
}

func makeShell(mgr *sandbox.Manager) func(context.Context, *mcp.CallToolRequest, shellInput) (*mcp.CallToolResult, shellOutput, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in shellInput) (*mcp.CallToolResult, shellOutput, error) {
		if strings.TrimSpace(in.Command) == "" {
			return errResult("command is required"), shellOutput{}, nil
		}
		res, err := mgr.Exec(ctx, in.Sandbox, in.Command)
		if err != nil {
			return errResult("%v", err), shellOutput{}, nil
		}
		// A non-zero exit is a normal result (the agent inspects it), not a tool error.
		return nil, shellOutput{Output: res.Output, ExitCode: res.ExitCode}, nil
	}
}

type globInput struct {
	Sandbox string `json:"sandbox" jsonschema:"Sandbox id"`
	Pattern string `json:"pattern" jsonschema:"Glob pattern (globstar enabled), relative to /work"`
}
type globOutput struct {
	Matches []string `json:"matches"`
}

func makeGlob(mgr *sandbox.Manager) func(context.Context, *mcp.CallToolRequest, globInput) (*mcp.CallToolResult, globOutput, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in globInput) (*mcp.CallToolResult, globOutput, error) {
		if strings.TrimSpace(in.Pattern) == "" {
			return errResult("pattern is required"), globOutput{}, nil
		}
		// globstar + nullglob so **/*.go works and no-match yields nothing.
		cmd := "cd " + workRoot + " && shopt -s globstar nullglob && printf '%s\\n' " + in.Pattern
		res, err := mgr.Exec(ctx, in.Sandbox, cmd)
		if err != nil {
			return errResult("%v", err), globOutput{}, nil
		}
		var matches []string
		for _, line := range strings.Split(strings.TrimSpace(res.Output), "\n") {
			if line != "" {
				matches = append(matches, line)
			}
		}
		return nil, globOutput{Matches: matches}, nil
	}
}

type grepInput struct {
	Sandbox string `json:"sandbox" jsonschema:"Sandbox id"`
	Pattern string `json:"pattern" jsonschema:"Search pattern (extended regex)"`
	Path    string `json:"path,omitempty" jsonschema:"Optional path to scope the search (relative to /work; default whole tree)"`
}
type grepOutput struct {
	Output string `json:"output"`
}

func makeGrep(mgr *sandbox.Manager) func(context.Context, *mcp.CallToolRequest, grepInput) (*mcp.CallToolResult, grepOutput, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in grepInput) (*mcp.CallToolResult, grepOutput, error) {
		if strings.TrimSpace(in.Pattern) == "" {
			return errResult("pattern is required"), grepOutput{}, nil
		}
		target := "."
		if in.Path != "" {
			full, err := resolvePath(in.Path)
			if err != nil {
				return errResult("%v", err), grepOutput{}, nil
			}
			target = full
		}
		// grep returns exit 1 when no matches — not an error for us.
		res, err := mgr.Exec(ctx, in.Sandbox, fmt.Sprintf("cd %s && grep -rnE -- %s %s || true", workRoot, shQuote(in.Pattern), shQuote(target)))
		if err != nil {
			return errResult("%v", err), grepOutput{}, nil
		}
		return nil, grepOutput{Output: res.Output}, nil
	}
}

type lspInput struct {
	Sandbox string `json:"sandbox" jsonschema:"Sandbox id"`
	Path    string `json:"path"    jsonschema:"File to check, relative to /work"`
}
type lspOutput struct {
	Path        string `json:"path"`
	Checker     string `json:"checker"`
	Clean       bool   `json:"clean"`
	Diagnostics string `json:"diagnostics"`
}

func makeLsp(mgr *sandbox.Manager) func(context.Context, *mcp.CallToolRequest, lspInput) (*mcp.CallToolResult, lspOutput, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in lspInput) (*mcp.CallToolResult, lspOutput, error) {
		full, err := resolvePath(in.Path)
		if err != nil {
			return errResult("%v", err), lspOutput{}, nil
		}
		var checker, cmd string
		switch path.Ext(full) {
		case ".go":
			checker, cmd = "gopls", "cd "+workRoot+" && gopls check "+shQuote(full)
		case ".ts", ".tsx", ".js", ".jsx":
			checker, cmd = "tsc", "cd "+workRoot+" && tsc --noEmit --skipLibCheck "+shQuote(full)
		case ".py":
			checker, cmd = "pyright", "cd "+workRoot+" && pyright "+shQuote(full)
		default:
			return errResult("no diagnostics checker for %q (supported: .go, .ts/.tsx/.js/.jsx, .py)", path.Ext(full)), lspOutput{}, nil
		}
		res, err := mgr.Exec(ctx, in.Sandbox, cmd)
		if err != nil {
			return errResult("%v", err), lspOutput{}, nil
		}
		// Each checker exits 0 with no diagnostics; non-zero (or output) means issues.
		clean := res.ExitCode == 0 && strings.TrimSpace(res.Output) == ""
		return nil, lspOutput{Path: full, Checker: checker, Clean: clean, Diagnostics: res.Output}, nil
	}
}

type deployInput struct {
	Sandbox     string `json:"sandbox"      jsonschema:"Sandbox id"`
	RunCommand  string `json:"run_command"  jsonschema:"Command to start the service, run from /work (e.g. 'go run .' or './app' or 'node server.js')"`
	Port        int    `json:"port"         jsonschema:"TCP port the service listens on"`
	HealthPath  string `json:"health_path,omitempty"  jsonschema:"HTTP path to healthcheck (default /)"`
	WaitSeconds int    `json:"wait_seconds,omitempty" jsonschema:"Seconds to wait for startup before healthcheck (default 3)"`
}
type deployOutput struct {
	Healthy      bool   `json:"healthy"`
	Pid          string `json:"pid,omitempty"`
	HealthOutput string `json:"health_output"`
	Log          string `json:"log,omitempty"`
}

func makeDeploy(mgr *sandbox.Manager) func(context.Context, *mcp.CallToolRequest, deployInput) (*mcp.CallToolResult, deployOutput, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in deployInput) (*mcp.CallToolResult, deployOutput, error) {
		if strings.TrimSpace(in.Sandbox) == "" || strings.TrimSpace(in.RunCommand) == "" {
			return errResult("sandbox and run_command are required"), deployOutput{}, nil
		}
		if in.Port <= 0 {
			return errResult("port is required (the TCP port the service listens on)"), deployOutput{}, nil
		}
		hp := in.HealthPath
		if hp == "" {
			hp = "/"
		}
		wait := in.WaitSeconds
		if wait <= 0 {
			wait = 3
		}
		// Start the service detached (setsid + redirected stdio) so it survives
		// the exec and keeps running in the sandbox container.
		start := fmt.Sprintf("cd %s && setsid bash -lc %s >/tmp/deploy.log 2>&1 </dev/null & PID=$!; sleep %d; echo $PID",
			workRoot, shQuote(in.RunCommand), wait)
		startRes, err := mgr.Exec(ctx, in.Sandbox, start)
		if err != nil {
			return errResult("%v", err), deployOutput{}, nil
		}
		pid := strings.TrimSpace(startRes.Output)
		// Healthcheck: curl exit 0 = healthy.
		health, err := mgr.Exec(ctx, in.Sandbox,
			fmt.Sprintf("curl -fsS -m 5 http://localhost:%d%s", in.Port, hp))
		if err != nil {
			return errResult("%v", err), deployOutput{}, nil
		}
		logRes, _ := mgr.Exec(ctx, in.Sandbox, "tail -n 20 /tmp/deploy.log")
		return nil, deployOutput{
			Healthy:      health.ExitCode == 0,
			Pid:          pid,
			HealthOutput: health.Output,
			Log:          logRes.Output,
		}, nil
	}
}

// --- sandbox visibility + reaper (CC.6) ---

type listSbOutput struct {
	Sandboxes []sandbox.SandboxInfo `json:"sandboxes"`
}

func makeSandboxList(mgr *sandbox.Manager) func(context.Context, *mcp.CallToolRequest, struct{}) (*mcp.CallToolResult, listSbOutput, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, listSbOutput, error) {
		infos, err := mgr.ListSandboxes(ctx)
		if err != nil {
			return errResult("%v", err), listSbOutput{}, nil
		}
		return nil, listSbOutput{Sandboxes: infos}, nil
	}
}

type reapInput struct {
	MaxAgeMinutes int `json:"max_age_minutes,omitempty" jsonschema:"Remove sandboxes older than this many minutes (0/omitted = default 120; negative = flush ALL)"`
}
type reapOutput struct {
	Removed []string `json:"removed"`
}

func makeSandboxReap(mgr *sandbox.Manager) func(context.Context, *mcp.CallToolRequest, reapInput) (*mcp.CallToolResult, reapOutput, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in reapInput) (*mcp.CallToolResult, reapOutput, error) {
		age := in.MaxAgeMinutes
		if age == 0 { // 0/omitted = default; negative = flush all
			age = 120
		}
		removed, err := mgr.ReapSandboxes(ctx, time.Duration(age)*time.Minute)
		if err != nil {
			return errResult("%v", err), reapOutput{}, nil
		}
		return nil, reapOutput{Removed: removed}, nil
	}
}

// --- git: commit (local) / push / open PR (bridge-side, token never in sandbox) ---

type commitInput struct {
	Sandbox string `json:"sandbox" jsonschema:"Sandbox id (repo-mode)"`
	Message string `json:"message" jsonschema:"Commit message"`
	Branch  string `json:"branch,omitempty" jsonschema:"Branch to commit onto (default agent/coder/<sandbox>; protected branches refused)"`
}
type commitOutput struct {
	Sha    string `json:"sha"`
	Branch string `json:"branch"`
}

func makeCommit(mgr *sandbox.Manager) func(context.Context, *mcp.CallToolRequest, commitInput) (*mcp.CallToolResult, commitOutput, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in commitInput) (*mcp.CallToolResult, commitOutput, error) {
		if strings.TrimSpace(in.Message) == "" {
			return errResult("message is required"), commitOutput{}, nil
		}
		sha, br, err := mgr.Commit(ctx, in.Sandbox, in.Message, in.Branch)
		if err != nil {
			return errResult("%v", err), commitOutput{}, nil
		}
		return nil, commitOutput{Sha: sha, Branch: br}, nil
	}
}

type pushInput struct {
	Sandbox string `json:"sandbox" jsonschema:"Sandbox id (repo-mode)"`
	Branch  string `json:"branch"  jsonschema:"Branch to push (protected branches refused)"`
}
type pushOutput struct {
	Output string `json:"output"`
}

func makePush(mgr *sandbox.Manager) func(context.Context, *mcp.CallToolRequest, pushInput) (*mcp.CallToolResult, pushOutput, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in pushInput) (*mcp.CallToolResult, pushOutput, error) {
		out, err := mgr.Push(ctx, in.Sandbox, in.Branch)
		if err != nil {
			return errResult("%v", err), pushOutput{}, nil
		}
		return nil, pushOutput{Output: out}, nil
	}
}

type prInput struct {
	Sandbox string `json:"sandbox" jsonschema:"Sandbox id (repo-mode)"`
	Title   string `json:"title"   jsonschema:"PR title"`
	Body    string `json:"body"    jsonschema:"PR body (markdown)"`
	Base    string `json:"base,omitempty"  jsonschema:"Base branch (default main)"`
	Draft   bool   `json:"draft,omitempty" jsonschema:"Open as a draft PR"`
}
type prOutput struct {
	URL string `json:"url"`
}

func makeOpenPR(mgr *sandbox.Manager) func(context.Context, *mcp.CallToolRequest, prInput) (*mcp.CallToolResult, prOutput, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in prInput) (*mcp.CallToolResult, prOutput, error) {
		if strings.TrimSpace(in.Title) == "" {
			return errResult("title is required"), prOutput{}, nil
		}
		url, err := mgr.OpenPR(ctx, in.Sandbox, in.Title, in.Body, in.Base, in.Draft)
		if err != nil {
			return errResult("%v", err), prOutput{}, nil
		}
		return nil, prOutput{URL: url}, nil
	}
}

// shQuote single-quotes a string for safe inclusion in a shell command.
func shQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
