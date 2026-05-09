// Tool implementations for git-mcp. Each tool wraps a narrow git or
// gh invocation. Inputs validated against allow-list before any
// subprocess spawn. Outputs returned as structured types so the agent
// (and the calling work_queue row) sees clean JSON.

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerGitTools(srv *mcp.Server, cfg *gitConfig) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "git_clone",
		Description: "Clone a git repo into the per-work-item workdir " +
			"(/tmp/stewards-git/<work-item-id>/). Refuses if the workdir " +
			"already exists. Uses GITHUB_TOKEN from env for auth on " +
			"github.com remotes.",
	}, makeGitClone(cfg))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "git_status",
		Description: "Run `git status --porcelain=v1 --branch` inside " +
			"the work-item's workdir. Returns the porcelain output as text " +
			"plus the current branch name.",
	}, makeGitStatus(cfg))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "git_branch_create",
		Description: "Create and check out a new branch under the agent " +
			"namespace: agent/<pipeline>/<work-item-id>-<slug>. Refuses " +
			"protected branches and any name outside the namespace regex. " +
			"Slug truncated to ~40 chars.",
	}, makeGitBranchCreate(cfg))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "git_add",
		Description: "Run `git add` on a list of paths inside the workdir. " +
			"Paths are validated to stay within the workdir (no .. " +
			"traversal). Refuses if no paths supplied.",
	}, makeGitAdd(cfg))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "git_commit",
		Description: "Commit staged changes with the supplied message. " +
			"Auto-appends Co-Authored-By: <agent-family>-via-pg-ai-stewards " +
			"<configured-email> trailer. Refuses empty message; refuses " +
			"--amend (not exposed). Returns the new commit SHA.",
	}, makeGitCommit(cfg))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "git_push",
		Description: "Push the current branch to origin. Refuses if " +
			"branch is protected (main, master, release/*). --force is " +
			"never sent. Returns the push output.",
	}, makeGitPush(cfg))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "gh_pr_create",
		Description: "Create a GitHub pull request via `gh pr create`. " +
			"head must be in the agent/* namespace. Defaults to " +
			"ready-for-review (not draft). Returns the PR URL.",
	}, makeGhPrCreate(cfg))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "gh_issue_create",
		Description: "Create a GitHub issue via `gh issue create`. " +
			"Title and body required. Returns the issue URL.",
	}, makeGhIssueCreate(cfg))
}

// ---------------------------------------------------------------------
// git_clone
// ---------------------------------------------------------------------

type gitCloneInput struct {
	RepoURL    string `json:"repo_url" jsonschema:"Git remote URL (https or ssh)"`
	WorkItemID string `json:"work_item_id" jsonschema:"Work item id; used as workdir name (UUID or short id)"`
}

type gitCloneOutput struct {
	Workdir string `json:"workdir"`
	Stdout  string `json:"stdout,omitempty"`
}

func makeGitClone(cfg *gitConfig) func(context.Context, *mcp.CallToolRequest, gitCloneInput) (*mcp.CallToolResult, gitCloneOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in gitCloneInput) (*mcp.CallToolResult, gitCloneOutput, error) {
		if strings.TrimSpace(in.RepoURL) == "" {
			return toolError("repo_url is required"), gitCloneOutput{}, nil
		}
		if err := validateWorkdirID(in.WorkItemID); err != nil {
			return toolError("%v", err), gitCloneOutput{}, nil
		}
		workdir := filepath.Join(cfg.WorkdirRoot, in.WorkItemID)
		if _, err := os.Stat(workdir); err == nil {
			return toolError("workdir %s already exists; refusing to clone over it", workdir), gitCloneOutput{}, nil
		}

		// Token is injected via env, not via the URL, so it never
		// appears in process arglist or stdout. gh CLI / git's
		// credential helper picks it up.
		out, err := runGit(ctx, cfg, cfg.WorkdirRoot, "clone", in.RepoURL, workdir)
		if err != nil {
			return toolError("git clone: %v\n%s", err, out), gitCloneOutput{}, nil
		}
		return nil, gitCloneOutput{Workdir: workdir, Stdout: out}, nil
	}
}

// ---------------------------------------------------------------------
// git_status
// ---------------------------------------------------------------------

type gitStatusInput struct {
	WorkItemID string `json:"work_item_id" jsonschema:"Work item id (workdir name)"`
}

type gitStatusOutput struct {
	Branch string `json:"branch,omitempty"`
	Status string `json:"status"`
}

func makeGitStatus(cfg *gitConfig) func(context.Context, *mcp.CallToolRequest, gitStatusInput) (*mcp.CallToolResult, gitStatusOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in gitStatusInput) (*mcp.CallToolResult, gitStatusOutput, error) {
		workdir, err := workdirFor(cfg, in.WorkItemID)
		if err != nil {
			return toolError("%v", err), gitStatusOutput{}, nil
		}
		out, err := runGit(ctx, cfg, workdir, "status", "--porcelain=v1", "--branch")
		if err != nil {
			return toolError("git status: %v\n%s", err, out), gitStatusOutput{}, nil
		}
		// First line of `--branch` output is `## <branch>...`.
		branch := ""
		if first, _, _ := strings.Cut(out, "\n"); strings.HasPrefix(first, "## ") {
			rest := strings.TrimPrefix(first, "## ")
			branch, _, _ = strings.Cut(rest, "...")
		}
		return nil, gitStatusOutput{Branch: branch, Status: out}, nil
	}
}

// ---------------------------------------------------------------------
// git_branch_create
// ---------------------------------------------------------------------

type gitBranchCreateInput struct {
	WorkItemID string `json:"work_item_id" jsonschema:"Work item id (workdir name); also used in the branch name"`
	Pipeline   string `json:"pipeline"     jsonschema:"Pipeline family name (e.g. study-write)"`
	Slug       string `json:"slug,omitempty" jsonschema:"Optional short slug for human readability (truncated to ~40 chars)"`
}

type gitBranchCreateOutput struct {
	Branch string `json:"branch"`
}

func makeGitBranchCreate(cfg *gitConfig) func(context.Context, *mcp.CallToolRequest, gitBranchCreateInput) (*mcp.CallToolResult, gitBranchCreateOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in gitBranchCreateInput) (*mcp.CallToolResult, gitBranchCreateOutput, error) {
		workdir, err := workdirFor(cfg, in.WorkItemID)
		if err != nil {
			return toolError("%v", err), gitBranchCreateOutput{}, nil
		}
		branch, err := buildAgentBranchName(in.Pipeline, shortenWorkItemID(in.WorkItemID), in.Slug)
		if err != nil {
			return toolError("branch name: %v", err), gitBranchCreateOutput{}, nil
		}
		out, err := runGit(ctx, cfg, workdir, "checkout", "-b", branch)
		if err != nil {
			return toolError("git checkout -b %s: %v\n%s", branch, err, out), gitBranchCreateOutput{}, nil
		}
		return nil, gitBranchCreateOutput{Branch: branch}, nil
	}
}

// ---------------------------------------------------------------------
// git_add
// ---------------------------------------------------------------------

type gitAddInput struct {
	WorkItemID string   `json:"work_item_id" jsonschema:"Work item id (workdir name)"`
	Paths      []string `json:"paths"        jsonschema:"Paths to add (relative to workdir; no .. traversal)"`
}

type gitAddOutput struct {
	AddedPaths []string `json:"added_paths"`
}

func makeGitAdd(cfg *gitConfig) func(context.Context, *mcp.CallToolRequest, gitAddInput) (*mcp.CallToolResult, gitAddOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in gitAddInput) (*mcp.CallToolResult, gitAddOutput, error) {
		workdir, err := workdirFor(cfg, in.WorkItemID)
		if err != nil {
			return toolError("%v", err), gitAddOutput{}, nil
		}
		if len(in.Paths) == 0 {
			return toolError("paths is required"), gitAddOutput{}, nil
		}
		for _, p := range in.Paths {
			if strings.Contains(p, "..") || filepath.IsAbs(p) {
				return toolError("path %q contains .. or is absolute; refusing", p), gitAddOutput{}, nil
			}
		}
		args := append([]string{"add", "--"}, in.Paths...)
		out, err := runGit(ctx, cfg, workdir, args...)
		if err != nil {
			return toolError("git add: %v\n%s", err, out), gitAddOutput{}, nil
		}
		return nil, gitAddOutput{AddedPaths: in.Paths}, nil
	}
}

// ---------------------------------------------------------------------
// git_commit
// ---------------------------------------------------------------------

type gitCommitInput struct {
	WorkItemID  string `json:"work_item_id" jsonschema:"Work item id (workdir name)"`
	Message     string `json:"message"      jsonschema:"Commit message body (subject + optional body)"`
	AgentFamily string `json:"agent_family" jsonschema:"Agent family that authored the commit (used in Co-Authored-By trailer)"`
}

type gitCommitOutput struct {
	Sha     string `json:"sha"`
	Message string `json:"message"`
}

func makeGitCommit(cfg *gitConfig) func(context.Context, *mcp.CallToolRequest, gitCommitInput) (*mcp.CallToolResult, gitCommitOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in gitCommitInput) (*mcp.CallToolResult, gitCommitOutput, error) {
		workdir, err := workdirFor(cfg, in.WorkItemID)
		if err != nil {
			return toolError("%v", err), gitCommitOutput{}, nil
		}
		if strings.TrimSpace(in.Message) == "" {
			return toolError("message is required"), gitCommitOutput{}, nil
		}
		family := normalizeSegment(in.AgentFamily)
		if family == "" {
			return toolError("agent_family is required"), gitCommitOutput{}, nil
		}
		fullMsg := in.Message + "\n\nCo-Authored-By: " +
			family + "-via-pg-ai-stewards <" + cfg.CoAuthorEmail + ">\n"
		out, err := runGit(ctx, cfg, workdir, "commit", "-m", fullMsg)
		if err != nil {
			return toolError("git commit: %v\n%s", err, out), gitCommitOutput{}, nil
		}
		// Resolve the new SHA
		shaOut, err := runGit(ctx, cfg, workdir, "rev-parse", "HEAD")
		if err != nil {
			return toolError("git rev-parse HEAD after commit: %v", err), gitCommitOutput{}, nil
		}
		return nil, gitCommitOutput{
			Sha:     strings.TrimSpace(shaOut),
			Message: fullMsg,
		}, nil
	}
}

// ---------------------------------------------------------------------
// git_push
// ---------------------------------------------------------------------

type gitPushInput struct {
	WorkItemID string `json:"work_item_id" jsonschema:"Work item id (workdir name)"`
	Branch     string `json:"branch"       jsonschema:"Branch to push (must be in agent/* namespace)"`
}

type gitPushOutput struct {
	Stdout string `json:"stdout,omitempty"`
}

func makeGitPush(cfg *gitConfig) func(context.Context, *mcp.CallToolRequest, gitPushInput) (*mcp.CallToolResult, gitPushOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in gitPushInput) (*mcp.CallToolResult, gitPushOutput, error) {
		workdir, err := workdirFor(cfg, in.WorkItemID)
		if err != nil {
			return toolError("%v", err), gitPushOutput{}, nil
		}
		if err := validateAgentBranch(in.Branch); err != nil {
			return toolError("%v", err), gitPushOutput{}, nil
		}
		// Set upstream on first push so subsequent ops are simpler.
		out, err := runGit(ctx, cfg, workdir, "push", "--set-upstream", "origin", in.Branch)
		if err != nil {
			return toolError("git push: %v\n%s", err, out), gitPushOutput{}, nil
		}
		return nil, gitPushOutput{Stdout: out}, nil
	}
}

// ---------------------------------------------------------------------
// gh_pr_create
// ---------------------------------------------------------------------

type ghPrCreateInput struct {
	Repo  string `json:"repo,omitempty" jsonschema:"Repo in OWNER/REPO form (default: inferred from workdir if work_item_id supplied)"`
	WorkItemID string `json:"work_item_id,omitempty" jsonschema:"Optional workdir id; gh infers repo from local clone"`
	Head  string `json:"head"  jsonschema:"Source branch (must be in agent/* namespace)"`
	Base  string `json:"base"  jsonschema:"Target branch (typically main)"`
	Title string `json:"title" jsonschema:"PR title"`
	Body  string `json:"body"  jsonschema:"PR body (markdown)"`
	Draft bool   `json:"draft,omitempty" jsonschema:"Open as draft PR (default false; ready-for-review)"`
}

type ghPrCreateOutput struct {
	URL string `json:"url"`
}

func makeGhPrCreate(cfg *gitConfig) func(context.Context, *mcp.CallToolRequest, ghPrCreateInput) (*mcp.CallToolResult, ghPrCreateOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in ghPrCreateInput) (*mcp.CallToolResult, ghPrCreateOutput, error) {
		if strings.TrimSpace(in.Title) == "" {
			return toolError("title is required"), ghPrCreateOutput{}, nil
		}
		if err := validateAgentBranch(in.Head); err != nil {
			return toolError("head: %v", err), ghPrCreateOutput{}, nil
		}
		if strings.TrimSpace(in.Base) == "" {
			in.Base = "main"
		}
		// gh needs a workdir context (it reads .git/config to find the repo)
		// unless --repo is supplied.
		workdir := cfg.WorkdirRoot
		if in.WorkItemID != "" {
			wd, err := workdirFor(cfg, in.WorkItemID)
			if err == nil {
				workdir = wd
			}
		}
		args := []string{"pr", "create",
			"--head", in.Head,
			"--base", in.Base,
			"--title", in.Title,
			"--body", in.Body,
		}
		if in.Repo != "" {
			args = append(args, "--repo", in.Repo)
		}
		if in.Draft {
			args = append(args, "--draft")
		}
		out, err := runGh(ctx, cfg, workdir, args...)
		if err != nil {
			return toolError("gh pr create: %v\n%s", err, out), ghPrCreateOutput{}, nil
		}
		// gh prints the PR URL on stdout
		url := strings.TrimSpace(out)
		return nil, ghPrCreateOutput{URL: url}, nil
	}
}

// ---------------------------------------------------------------------
// gh_issue_create
// ---------------------------------------------------------------------

type ghIssueCreateInput struct {
	Repo  string `json:"repo,omitempty" jsonschema:"Repo in OWNER/REPO form (required if not run from a workdir)"`
	WorkItemID string `json:"work_item_id,omitempty" jsonschema:"Optional workdir id for repo inference"`
	Title string `json:"title" jsonschema:"Issue title"`
	Body  string `json:"body"  jsonschema:"Issue body (markdown)"`
}

type ghIssueCreateOutput struct {
	URL string `json:"url"`
}

func makeGhIssueCreate(cfg *gitConfig) func(context.Context, *mcp.CallToolRequest, ghIssueCreateInput) (*mcp.CallToolResult, ghIssueCreateOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in ghIssueCreateInput) (*mcp.CallToolResult, ghIssueCreateOutput, error) {
		if strings.TrimSpace(in.Title) == "" {
			return toolError("title is required"), ghIssueCreateOutput{}, nil
		}
		workdir := cfg.WorkdirRoot
		if in.WorkItemID != "" {
			wd, err := workdirFor(cfg, in.WorkItemID)
			if err == nil {
				workdir = wd
			}
		}
		args := []string{"issue", "create",
			"--title", in.Title,
			"--body", in.Body,
		}
		if in.Repo != "" {
			args = append(args, "--repo", in.Repo)
		}
		out, err := runGh(ctx, cfg, workdir, args...)
		if err != nil {
			return toolError("gh issue create: %v\n%s", err, out), ghIssueCreateOutput{}, nil
		}
		return nil, ghIssueCreateOutput{URL: strings.TrimSpace(out)}, nil
	}
}

// ---------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------

// workdirFor resolves and validates the per-work-item workdir path.
// Refuses if the workdir doesn't exist (caller must clone first).
func workdirFor(cfg *gitConfig, workItemID string) (string, error) {
	if err := validateWorkdirID(workItemID); err != nil {
		return "", err
	}
	wd := filepath.Join(cfg.WorkdirRoot, workItemID)
	st, err := os.Stat(wd)
	if err != nil {
		return "", fmt.Errorf("workdir %s not found (clone first?): %w", wd, err)
	}
	if !st.IsDir() {
		return "", fmt.Errorf("workdir path %s exists but is not a directory", wd)
	}
	return wd, nil
}

// runGit executes git in the given workdir with arglist `args`.
// Returns combined stdout+stderr. GITHUB_TOKEN inherits via env.
func runGit(ctx context.Context, cfg *gitConfig, dir string, args ...string) (string, error) {
	return runCmdAt(ctx, dir, cfg.GitBin, args...)
}

// runGh executes gh in the given workdir with arglist `args`.
// gh inherits GITHUB_TOKEN via env.
func runGh(ctx context.Context, cfg *gitConfig, dir string, args ...string) (string, error) {
	return runCmdAt(ctx, dir, cfg.GhBin, args...)
}

// runCmdAt is the shared subprocess runner. 60s timeout per call.
func runCmdAt(ctx context.Context, dir, bin string, args ...string) (string, error) {
	callCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(callCtx, bin, args...)
	cmd.Dir = dir
	cmd.Env = os.Environ() // inherit GITHUB_TOKEN, PATH, etc.
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func toolError(format string, args ...any) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(format, args...)}},
	}
}
