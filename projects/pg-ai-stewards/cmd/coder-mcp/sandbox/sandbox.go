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
// down any pre-existing container of the same name first.
func (m *Manager) Provision(ctx context.Context, wi string, net Network) error {
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
	if net == NetOff {
		args = append(args, "--network=none")
	}
	args = append(args, m.Image, "sleep", "infinity")
	if out, err := docker(ctx, args...); err != nil {
		return fmt.Errorf("provision %s: %w\n%s", wi, err, out)
	}
	return nil
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
