// Package sandbox manages per-work_item coding sandboxes: ephemeral Docker
// containers (image coder-runtime) the substrate's coder writes/builds/tests
// inside. It shells out to the `docker` CLI against the host daemon (the
// bridge mounts /var/run/docker.sock) — the "trusted-tool" isolation tier
// ratified in substrate-coding-capability D-CC2 (medium-safe; shared host
// kernel accepted for our own code).
//
// Lifecycle (D-CC8 — owned here, keyed by work_item id):
//
//	Provision(wi)  docker run -d --name coder-sb-<wi> <hardening> <net> coder-runtime sleep infinity
//	Exec(wi, cmd)  docker exec coder-sb-<wi> bash -lc '<cmd>'
//	Teardown(wi)   docker rm -f coder-sb-<wi>
//
// The worktree lives inside the container's own (ephemeral) filesystem and is
// discarded on teardown — the ephemeral-per-task posture from the research.
// The coder never touches the live /workspace mount (proposal §4).
package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Network controls the sandbox's egress. Default is On (D-CC5: open,
// default-on, switchable offline — the agent must pull go mod / npm / pip).
type Network string

const (
	NetOn  Network = "on"  // host daemon default network: egress allowed
	NetOff Network = "off" // --network none: fully offline
)

// Manager provisions and drives coding sandboxes.
type Manager struct {
	Image     string // coder-runtime image (CODER_RUNTIME_IMAGE or default)
	MemLimit  string // --memory (e.g. "2g")
	CPULimit  string // --cpus  (e.g. "2")
	PidsLimit string // --pids-limit
}

// New returns a Manager with the ratified defaults.
func New() *Manager {
	img := os.Getenv("CODER_RUNTIME_IMAGE")
	if img == "" {
		img = "coder-runtime:latest"
	}
	// CV2.2: git/gh run as root (bridge) over coder-uid-owned worktrees; disable
	// git's dubious-ownership guard for our own worktrees so commit/push/gh work
	// (the worktrees are ours — the guard is a multi-user safety net we don't need).
	_ = exec.Command("git", "config", "--global", "--add", "safe.directory", "*").Run()
	return &Manager{Image: img, MemLimit: "2g", CPULimit: "2", PidsLimit: "512"}
}

// containerName is the deterministic per-work_item container name.
func containerName(wi string) string {
	return "coder-sb-" + sanitize(wi)
}

// sanitize keeps the work_item id docker-name-safe ([a-zA-Z0-9_.-]).
func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9',
			r == '_', r == '.', r == '-':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	out := b.String()
	if out == "" {
		out = "wi"
	}
	return out
}

// docker runs a docker subcommand, returning combined output.
func docker(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return buf.String(), err
}

// Provision starts an idle sandbox container for wi. Idempotent-ish: it tears
// down any pre-existing container of the same name first. When worktree is
// true, the shared coder-worktrees volume (subpath wi) is mounted at /work —
// so the coder tools operate on a repo the bridge cloned there (CV2.1). The
// caller must CloneRepo first (the subpath must exist).
func (m *Manager) Provision(ctx context.Context, wi string, net Network, worktree bool) error {
	_ = m.Teardown(ctx, wi) // clear any leftover; ignore "not found"
	args := []string{
		"run", "-d", "--name", containerName(wi),
		// Hardening (defense-in-depth; the container is the real boundary).
		"--cap-drop=ALL",
		"--security-opt=no-new-privileges",
		"--memory=" + m.MemLimit,
		"--cpus=" + m.CPULimit,
		"--pids-limit=" + m.PidsLimit,
		"--label=stewards.coder=1",
		"--label=stewards.work_item=" + sanitize(wi),
	}
	if worktree {
		args = append(args, "--mount",
			fmt.Sprintf("type=volume,source=%s,target=/work,volume-subpath=%s", worktreeVol, sanitize(wi)))
	}
	if net == NetOff {
		args = append(args, "--network=none")
	}
	args = append(args, m.Image, "sleep", "infinity")
	if out, err := docker(ctx, args...); err != nil {
		return fmt.Errorf("provision %s: %w\n%s", wi, err, out)
	}
	return nil
}

// --- coder-v2: repo worktrees (CV2.1) ---

const (
	worktreeVol  = "coder-worktrees" // shared volume; bridge + sandbox both mount it
	worktreeRoot = "/worktrees"      // the bridge's mount point of worktreeVol
)

// repoAllowed reports whether repo matches CODER_REPO_ALLOWLIST (comma-separated
// substrings; default: ai-chattermax only). The tool-layer guard from D-CV2.2 —
// even with a token, the coder only touches whitelisted repos.
func repoAllowed(repo string) bool {
	list := os.Getenv("CODER_REPO_ALLOWLIST")
	if list == "" {
		list = "github.com/cpuchip/ai-chattermax"
	}
	for _, pat := range strings.Split(list, ",") {
		if pat = strings.TrimSpace(pat); pat != "" && strings.Contains(repo, pat) {
			return true
		}
	}
	return false
}

// CloneRepo clones an allow-listed repo into the per-work_item worktree
// (/worktrees/<wi> on the shared volume) and chowns it to the sandbox's coder
// uid (1000). Runs in the bridge — the GitHub token (CV2.2) lives here, never
// in the sandbox.
func (m *Manager) CloneRepo(ctx context.Context, wi, repo, branch string) error {
	if !repoAllowed(repo) {
		return fmt.Errorf("repo %q not in the coder allow-list (CODER_REPO_ALLOWLIST)", repo)
	}
	dir := worktreeRoot + "/" + sanitize(wi)
	_ = exec.CommandContext(ctx, "rm", "-rf", dir).Run() // fresh clone
	args := []string{"clone", "--depth", "50"}
	if branch != "" {
		args = append(args, "--branch", branch)
	}
	args = append(args, repo, dir)
	cmd := exec.CommandContext(ctx, "git", args...)
	var buf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &buf, &buf
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("clone %s: %w\n%s", repo, err, buf.String())
	}
	if out, err := exec.CommandContext(ctx, "chown", "-R", "1000:1000", dir).CombinedOutput(); err != nil {
		return fmt.Errorf("chown worktree %s: %w\n%s", dir, err, out)
	}
	return nil
}

// WorktreePath is the bridge-side path of wi's repo worktree.
func (m *Manager) WorktreePath(wi string) string { return worktreeRoot + "/" + sanitize(wi) }

func (m *Manager) HasWorktree(wi string) bool {
	_, err := os.Stat(m.WorktreePath(wi) + "/.git")
	return err == nil
}

// gitC runs `git -C dir args...` (combined output). Inherits coder-mcp's env,
// which carries GITHUB_TOKEN — these run bridge-side, never in the sandbox.
func gitC(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", dir, "-c", "safe.directory=*"}, args...)...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func protectedBranch(b string) bool {
	return b == "main" || b == "master" || strings.HasPrefix(b, "release/")
}

// Commit stages all changes in wi's worktree onto `branch` (created if absent)
// and commits. Local op — no token. Returns the new SHA + the branch.
func (m *Manager) Commit(ctx context.Context, wi, message, branch string) (sha, br string, err error) {
	dir := m.WorktreePath(wi)
	if !m.HasWorktree(wi) {
		return "", "", fmt.Errorf("no repo worktree for %q — start the sandbox with repo=", wi)
	}
	if branch == "" {
		branch = "agent/coder/" + sanitize(wi)
	}
	if protectedBranch(branch) {
		return "", "", fmt.Errorf("refusing to commit onto protected branch %q", branch)
	}
	if out, e := gitC(ctx, dir, "checkout", "-B", branch); e != nil {
		return "", "", fmt.Errorf("checkout %s: %w\n%s", branch, e, out)
	}
	if out, e := gitC(ctx, dir, "add", "-A"); e != nil {
		return "", "", fmt.Errorf("add: %w\n%s", e, out)
	}
	msg := message + "\n\nCo-Authored-By: pg-ai-stewards-coder <coder@cpuchip.net>\n"
	if out, e := gitC(ctx, dir, "-c", "user.name=pg-ai-stewards coder",
		"-c", "user.email=coder@cpuchip.net", "commit", "-m", msg); e != nil {
		return "", "", fmt.Errorf("commit: %w\n%s", e, out)
	}
	out, _ := gitC(ctx, dir, "rev-parse", "HEAD")
	return strings.TrimSpace(out), branch, nil
}

// Push pushes branch to origin. The GitHub token (coder-mcp's env) is supplied
// via a one-shot credential helper — never persisted in .git/config or the
// worktree (so the sandbox can't read it). Runs bridge-side.
func (m *Manager) Push(ctx context.Context, wi, branch string) (string, error) {
	if branch == "" || protectedBranch(branch) {
		return "", fmt.Errorf("refusing to push protected/empty branch %q", branch)
	}
	const helper = `!f() { echo username=x-access-token; echo "password=$GITHUB_TOKEN"; }; f`
	out, err := gitC(ctx, m.WorktreePath(wi),
		"-c", "credential.helper=", "-c", "credential.helper="+helper,
		"push", "--set-upstream", "origin", branch)
	if err != nil {
		return "", fmt.Errorf("push %s: %w\n%s", branch, err, out)
	}
	return out, nil
}

// OpenPR opens a pull request via gh (uses GITHUB_TOKEN from env). Bridge-side.
func (m *Manager) OpenPR(ctx context.Context, wi, title, body, base string, draft bool) (string, error) {
	if base == "" {
		base = "main"
	}
	// Pass --head explicitly. gh's "current branch" auto-detect unreliably
	// reports "you must first push the current branch" in this bridge-side
	// worktree setup even after a successful push; resolving the checked-out
	// branch and passing it as --head sidesteps that detection.
	head, herr := gitC(ctx, m.WorktreePath(wi), "rev-parse", "--abbrev-ref", "HEAD")
	if herr != nil {
		return "", fmt.Errorf("resolve head branch: %w\n%s", herr, head)
	}
	head = strings.TrimSpace(head)
	args := []string{"pr", "create", "--base", base, "--head", head, "--title", title, "--body", body}
	if draft {
		args = append(args, "--draft")
	}
	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = m.WorktreePath(wi)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gh pr create: %w\n%s", err, out)
	}
	return strings.TrimSpace(string(out)), nil
}

// ExecResult is the outcome of a sandbox command.
type ExecResult struct {
	Output   string
	ExitCode int
}

// Exec runs a shell command inside wi's sandbox (login shell, so PATH carries
// go/node/python). Returns the command's exit code separately from a docker
// transport error.
func (m *Manager) Exec(ctx context.Context, wi, command string) (ExecResult, error) {
	cmd := exec.CommandContext(ctx, "docker", "exec", containerName(wi), "bash", "-lc", command)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	res := ExecResult{Output: buf.String()}
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			res.ExitCode = ee.ExitCode()
			return res, nil // command failed inside the box — not a transport error
		}
		return res, fmt.Errorf("exec %s: %w\n%s", wi, err, buf.String())
	}
	return res, nil
}

// WriteFile writes content to an absolute path inside wi's sandbox, creating
// parent directories. Uses stdin so content needs no shell escaping.
func (m *Manager) WriteFile(ctx context.Context, wi, path, content string) error {
	cmd := exec.CommandContext(ctx, "docker", "exec", "-i", containerName(wi),
		"sh", "-c", `mkdir -p "$(dirname "$0")" && cat > "$0"`, path)
	cmd.Stdin = strings.NewReader(content)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("write %s: %w\n%s", path, err, buf.String())
	}
	return nil
}

// ReadFile reads an absolute path from wi's sandbox (argv form — no shell, so
// the path needs no quoting).
func (m *Manager) ReadFile(ctx context.Context, wi, path string) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "exec", containerName(wi), "cat", "--", path)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("read %s: %s", path, strings.TrimSpace(errBuf.String()))
		}
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	return out.String(), nil
}

// Exists reports whether wi's sandbox container is present.
func (m *Manager) Exists(ctx context.Context, wi string) (bool, error) {
	out, err := docker(ctx, "ps", "-aq", "--filter", "name=^"+containerName(wi)+"$")
	if err != nil {
		return false, fmt.Errorf("exists %s: %w\n%s", wi, err, out)
	}
	return strings.TrimSpace(out) != "", nil
}

// Teardown removes wi's sandbox container (force, ignores not-found).
func (m *Manager) Teardown(ctx context.Context, wi string) error {
	out, err := docker(ctx, "rm", "-f", containerName(wi))
	if err != nil && !strings.Contains(out, "No such container") {
		return fmt.Errorf("teardown %s: %w\n%s", wi, err, out)
	}
	return nil
}

// SandboxInfo describes a coder sandbox container.
type SandboxInfo struct {
	Name     string    `json:"name"`
	WorkItem string    `json:"work_item,omitempty"`
	Created  time.Time `json:"created"`
	AgeMin   int       `json:"age_minutes"`
}

// ListSandboxes lists all coder sandboxes (label stewards.coder=1) with age.
func (m *Manager) ListSandboxes(ctx context.Context) ([]SandboxInfo, error) {
	out, err := docker(ctx, "ps", "-a", "--filter", "label=stewards.coder=1",
		"--format", "{{.Names}}\t{{.CreatedAt}}\t{{.Label \"stewards.work_item\"}}")
	if err != nil {
		return nil, fmt.Errorf("list sandboxes: %w\n%s", err, out)
	}
	var infos []SandboxInfo
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		info := SandboxInfo{Name: parts[0]}
		if len(parts) == 3 {
			info.WorkItem = parts[2]
		}
		// docker's CreatedAt format, e.g. "2026-06-03 22:20:01 +0000 UTC".
		if len(parts) >= 2 {
			if t, perr := time.Parse("2006-01-02 15:04:05 -0700 MST", parts[1]); perr == nil {
				info.Created = t
				info.AgeMin = int(time.Since(t).Minutes())
			}
		}
		infos = append(infos, info)
	}
	return infos, nil
}

// ReapSandboxes force-removes sandboxes older than maxAge (the reaper for
// leaked/abandoned sandboxes). Returns the names removed.
func (m *Manager) ReapSandboxes(ctx context.Context, maxAge time.Duration) ([]string, error) {
	infos, err := m.ListSandboxes(ctx)
	if err != nil {
		return nil, err
	}
	var removed []string
	for _, info := range infos {
		if info.Created.IsZero() || time.Since(info.Created) <= maxAge {
			continue
		}
		if out, derr := docker(ctx, "rm", "-f", info.Name); derr == nil ||
			strings.Contains(out, "No such container") {
			removed = append(removed, info.Name)
		}
	}
	return removed, nil
}
