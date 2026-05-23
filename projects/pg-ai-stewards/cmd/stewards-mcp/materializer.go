// Autonomous materializer (am1, 2026-05-22).
//
// Drains stewards.pending_file_writes from inside the bridge daemon so
// scheduled and triggered pipelines actually land their files on disk
// without waiting for a human `git commit` to fire the pre-commit hook.
//
// Architecture per .spec/proposals/autonomous-materializer.md:
//
//   - LISTEN stewards_pending_file_write (fired by am1 trigger on INSERT)
//   - 60s safety poll tick in case NOTIFY is missed
//   - On either signal: exec `stewards-cli materialize-writes` against
//     the configured repo root. The CLI is the source-of-truth algorithm
//     (path validation, write_mode dispatch, mark-materialized). Bridge
//     just invokes it.
//
// Disabled when STEWARDS_MATERIALIZE_DISABLED=1 (so dev workflows that
// prefer the git-hook materializer can opt out).
//
// Repo root resolved from STEWARDS_REPO_ROOT env (default /workspace,
// which is where docker-compose mounts the repo root inside the bridge
// container; the compose mount was bumped to :rw concurrent with this
// goroutine landing).
//
// Companion SQL: extension/am1-pending-file-writes-notify.sql

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	materializerNotifyChannel = "stewards_pending_file_write"
	materializerPollInterval  = 60 * time.Second
	materializerCLIPath       = "stewards-cli"
)

// runMaterializer is started from runBridgeRun in a goroutine. It owns
// its own pgxConn for LISTEN (separate from the bridge's mcp_proxy
// listener) so the two never block each other.
//
// Returns only when ctx is done — runs for the lifetime of the bridge.
func runMaterializer(ctx context.Context, pool *pgxpool.Pool) {
	if os.Getenv("STEWARDS_MATERIALIZE_DISABLED") == "1" {
		log.Printf("materializer: disabled via STEWARDS_MATERIALIZE_DISABLED=1")
		return
	}

	repoRoot := os.Getenv("STEWARDS_REPO_ROOT")
	if repoRoot == "" {
		repoRoot = "/workspace"
	}

	// Verify the CLI exists before entering the loop — log clearly if not
	// so the root cause is obvious in container logs.
	if _, err := exec.LookPath(materializerCLIPath); err != nil {
		log.Printf("materializer: %s not found in PATH (%v) — autonomous draining disabled. "+
			"Pre-commit hook remains the materializer.", materializerCLIPath, err)
		return
	}

	// Verify the repo root exists + is writable. The bridge needs
	// :rw mount to actually flush files. Without it, exec'ing the CLI
	// will succeed but every write will fail with permission denied.
	if err := assertRepoRootWritable(repoRoot); err != nil {
		log.Printf("materializer: repo root %s not writable (%v) — autonomous draining disabled. "+
			"Check docker-compose mount (should be :rw).", repoRoot, err)
		return
	}

	// Dedicated LISTEN conn. Pinned via Hijack so the pool doesn't
	// recycle it; same pattern as the main mcp_proxy listener.
	listenAcq, err := pool.Acquire(ctx)
	if err != nil {
		log.Printf("materializer: acquire listen conn failed: %v — disabled", err)
		return
	}
	pgxConn := listenAcq.Hijack()
	defer pgxConn.Close(context.Background())

	if _, err := pgxConn.Exec(ctx, "LISTEN "+materializerNotifyChannel); err != nil {
		log.Printf("materializer: LISTEN %s failed: %v — disabled", materializerNotifyChannel, err)
		return
	}
	log.Printf("materializer: LISTENing on %s + %s safety poll (repo-root=%s)",
		materializerNotifyChannel, materializerPollInterval, repoRoot)

	// Drain once at startup to clear anything pending from a previous
	// bridge crash or pre-feature buildup.
	drainNow(ctx, repoRoot, "startup")

	for {
		waitCtx, cancel := context.WithTimeout(ctx, materializerPollInterval)
		_, err := pgxConn.WaitForNotification(waitCtx)
		cancel()

		if ctx.Err() != nil {
			log.Printf("materializer: shutting down (ctx done)")
			return
		}
		if err != nil {
			if waitCtx.Err() == context.DeadlineExceeded {
				// Tick — drain as safety net even without a NOTIFY.
				drainNow(ctx, repoRoot, "tick")
				continue
			}
			log.Printf("materializer: WaitForNotification: %v (sleeping 5s)", err)
			time.Sleep(5 * time.Second)
			continue
		}
		// NOTIFY arrived — drain.
		drainNow(ctx, repoRoot, "notify")
	}
}

// drainNow exec's `stewards-cli materialize-writes --repo-root <root>`,
// captures combined stdout+stderr, and logs the result. Errors are
// non-fatal — the next NOTIFY or tick retries.
//
// `trigger` is one of "startup" | "tick" | "notify" — surfaced in the
// log line for easy correlation with substrate events.
func drainNow(ctx context.Context, repoRoot, trigger string) {
	cmdCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, materializerCLIPath,
		"materialize-writes", "--repo-root", repoRoot)
	out, err := cmd.CombinedOutput()

	if err != nil {
		// Non-zero exit = at least one row failed. The CLI logs details
		// to stderr; surface a one-line summary here for log grepping
		// without flooding on big batches.
		log.Printf("materializer: drain trigger=%s FAILED (%v): %s",
			trigger, err, summarizeOutput(out))
		return
	}

	// CLI prints "materialize-writes: nothing pending" when there's
	// nothing to do — that's the dominant case. Quiet log it.
	if isQuietOutput(out) {
		return
	}

	log.Printf("materializer: drain trigger=%s ok: %s", trigger, summarizeOutput(out))
}

// assertRepoRootWritable creates and removes a probe file in the repo
// root to confirm the mount is rw. Cheap, runs once at startup.
func assertRepoRootWritable(root string) error {
	probe := root + "/.stewards-materializer-probe"
	if err := os.WriteFile(probe, []byte("ok\n"), 0o644); err != nil {
		return fmt.Errorf("write probe: %w", err)
	}
	if err := os.Remove(probe); err != nil {
		return fmt.Errorf("remove probe: %w", err)
	}
	return nil
}

func isQuietOutput(b []byte) bool {
	s := string(b)
	return s == "" || s == "materialize-writes: nothing pending\n"
}

// summarizeOutput trims the CLI's output to the last non-empty line
// for log brevity. The full output is captured in error context.
func summarizeOutput(b []byte) string {
	s := string(b)
	if len(s) == 0 {
		return "(no output)"
	}
	// Take the last 240 chars to capture the final summary line + a bit
	// of preceding error context if any.
	if len(s) > 240 {
		s = "..." + s[len(s)-240:]
	}
	// Collapse newlines for single-line log readability.
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' || s[i] == '\r' {
			if len(out) > 0 && out[len(out)-1] != ' ' {
				out = append(out, ' ')
			}
			continue
		}
		out = append(out, s[i])
	}
	return string(out)
}
