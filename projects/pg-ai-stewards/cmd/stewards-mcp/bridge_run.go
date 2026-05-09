// Bridge `run` mode (Phase 3e.2.c, 2026-05-08): the long-running
// daemon that services kind='mcp_proxy' rows in stewards.work_queue.
//
// Architecture:
//
//   - On startup: reap stale in_progress mcp_proxy rows (those left
//     by a previous bridge crash). Mark them errored so the bgworker's
//     completion pass can release waiting tool_dispatch parents.
//
//   - Main loop: LISTEN stewards_mcp_proxy on a dedicated pgx Conn,
//     plus a 1s poll tick as a safety net. On either signal: claim
//     oldest pending mcp_proxy row (FOR UPDATE SKIP LOCKED), dispatch
//     to a worker goroutine, repeat until the queue is empty.
//
//   - Per-row: look up the row's payload (server, tool, args), get
//     or lazy-init the MCP client session for that server, call
//     CallTool, write result back to work_queue, NOTIFY stewards_done.
//
//   - Sessions are cached by server name in a sync.Map; lazy-init on
//     first call. Sessions persist for the lifetime of the bridge
//     process. v1 does NOT detect or recover from server-side
//     crashes — that's documented as a known limitation. Restart the
//     bridge to flush.
//
// Concurrency: configurable worker count (default 4) running in
// parallel. Each worker handles one row at a time, but multiple
// workers can call into different MCP sessions concurrently. Same-
// session concurrent calls are also safe — the SDK serializes
// requests internally on its JSON-RPC pipe.

package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// runBridgeRun is the entry point for `stewards-mcp bridge run`.
// Long-running; exits on SIGINT/SIGTERM with a graceful drain.
func runBridgeRun(args []string) error {
	fs := flag.NewFlagSet("bridge run", flag.ContinueOnError)
	dsn := fs.String("dsn", "",
		"Postgres DSN (default: $STEWARDS_DSN, then localhost compose port 55433)")
	workers := fs.Int("workers", 4, "Number of concurrent worker goroutines")
	tickMs := fs.Int("tick-ms", 1000, "Poll interval safety net in ms (LISTEN is primary)")
	callTimeoutSecs := fs.Int("call-timeout", 60, "Per-tool-call timeout in seconds")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *dsn == "" {
		*dsn = os.Getenv("STEWARDS_DSN")
	}
	if *dsn == "" {
		*dsn = "postgres://stewards:stewards@localhost:55433/stewards?sslmode=disable"
	}

	rootCtx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(rootCtx, *dsn)
	if err != nil {
		return fmt.Errorf("pgxpool.New: %w", err)
	}
	defer pool.Close()
	if err := pool.Ping(rootCtx); err != nil {
		return fmt.Errorf("pool.Ping: %w", err)
	}
	log.Printf("bridge run: connected to substrate (%s)", redactDSN(*dsn))

	if err := reapStaleMcpProxyRows(rootCtx, pool); err != nil {
		// Non-fatal: log and continue. Reaper failures shouldn't
		// stop the daemon from servicing new traffic.
		log.Printf("bridge run: reaper failed (non-fatal): %v", err)
	}

	cache := newSessionCache()
	defer cache.closeAll()

	jobCh := make(chan int64, *workers*2)
	var wg sync.WaitGroup
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			runWorker(rootCtx, workerID, pool, cache, jobCh,
				time.Duration(*callTimeoutSecs)*time.Second)
		}(i)
	}
	log.Printf("bridge run: spawned %d worker(s); call-timeout=%ds",
		*workers, *callTimeoutSecs)

	// LISTEN on a dedicated pgx connection. Acquire from the pool's
	// underlying conn pool but pin it (Hijack) for the duration so
	// pgx's pool doesn't recycle it under us.
	listenConn, err := pool.Acquire(rootCtx)
	if err != nil {
		return fmt.Errorf("acquire listen conn: %w", err)
	}
	pgxConn := listenConn.Hijack()
	defer pgxConn.Close(context.Background())

	if _, err := pgxConn.Exec(rootCtx, "LISTEN stewards_mcp_proxy"); err != nil {
		return fmt.Errorf("LISTEN stewards_mcp_proxy: %w", err)
	}
	log.Printf("bridge run: LISTENing on stewards_mcp_proxy")

	tick := time.Duration(*tickMs) * time.Millisecond

	// Main loop: wait for notification or tick, then drain.
	for {
		// Drain whatever's already pending (may include rows enqueued
		// before LISTEN took effect, or rows the bridge missed during
		// transient disconnect). We loop until claim returns nothing.
		for drainOne(rootCtx, pool, jobCh) {
		}

		// Block waiting for a notification or the tick deadline.
		waitCtx, cancel := context.WithTimeout(rootCtx, tick)
		_, err := pgxConn.WaitForNotification(waitCtx)
		cancel()

		if rootCtx.Err() != nil {
			break
		}
		if err != nil {
			// Timeout is the expected fast path (no notification in
			// the tick window). Anything else is real.
			if waitCtx.Err() == context.DeadlineExceeded {
				continue
			}
			log.Printf("bridge run: WaitForNotification: %v (sleeping 1s)", err)
			time.Sleep(time.Second)
		}
	}

	log.Printf("bridge run: shutting down (closing job channel + waiting on workers)")
	close(jobCh)
	wg.Wait()
	log.Printf("bridge run: clean shutdown complete")
	return nil
}

// drainOne tries to claim one mcp_proxy row and hands it to a worker.
// Returns true if a row was claimed (caller should keep draining).
func drainOne(ctx context.Context, pool *pgxpool.Pool, jobCh chan<- int64) bool {
	var claimedID int64
	err := pool.QueryRow(ctx,
		`WITH next AS (
			SELECT id FROM stewards.work_queue
			WHERE status = 'pending' AND kind = 'mcp_proxy'
			ORDER BY created_at
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		UPDATE stewards.work_queue q
		SET status = 'in_progress', claimed_at = now()
		FROM next
		WHERE q.id = next.id
		RETURNING q.id`,
	).Scan(&claimedID)

	if errors.Is(err, pgx.ErrNoRows) {
		return false
	}
	if err != nil {
		log.Printf("bridge run: claim error: %v", err)
		return false
	}

	select {
	case jobCh <- claimedID:
		return true
	case <-ctx.Done():
		return false
	}
}

// runWorker pulls claimed row ids from jobCh, dispatches each one,
// writes the result back to the substrate.
func runWorker(ctx context.Context, workerID int, pool *pgxpool.Pool,
	cache *sessionCache, jobCh <-chan int64, callTimeout time.Duration) {
	for jobID := range jobCh {
		callCtx, cancel := context.WithTimeout(ctx, callTimeout)
		dispatchOne(callCtx, workerID, pool, cache, jobID)
		cancel()
	}
}

// dispatchOne handles a single claimed mcp_proxy row end-to-end.
// On success: writes result + NOTIFY. On failure: writes error +
// NOTIFY (so the parent tool_dispatch's completion pass releases
// the waiting tool reply with an error message).
func dispatchOne(ctx context.Context, workerID int, pool *pgxpool.Pool,
	cache *sessionCache, jobID int64) {
	var payloadJSON []byte
	if err := pool.QueryRow(ctx,
		"SELECT payload FROM stewards.work_queue WHERE id = $1",
		jobID,
	).Scan(&payloadJSON); err != nil {
		writeError(ctx, pool, jobID, fmt.Errorf("read payload: %w", err))
		return
	}

	var payload struct {
		Server string          `json:"server"`
		Tool   string          `json:"tool"`
		Args   json.RawMessage `json:"args"`
	}
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		writeError(ctx, pool, jobID, fmt.Errorf("decode payload: %w", err))
		return
	}

	// Decode args as map[string]any — that's what the SDK's CallTool
	// wants for Arguments. Models almost always emit objects; if they
	// emit something exotic the call will fail clearly.
	var args map[string]any
	if len(payload.Args) > 0 {
		if err := json.Unmarshal(payload.Args, &args); err != nil {
			writeError(ctx, pool, jobID, fmt.Errorf("decode args: %w", err))
			return
		}
	}

	session, err := cache.get(ctx, pool, payload.Server)
	if err != nil {
		writeError(ctx, pool, jobID, fmt.Errorf("session(%s): %w", payload.Server, err))
		return
	}

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      payload.Tool,
		Arguments: args,
	})
	if err != nil {
		writeError(ctx, pool, jobID, fmt.Errorf("CallTool(%s/%s): %w",
			payload.Server, payload.Tool, err))
		return
	}

	// Build the content the model will see. MCP CallToolResult has
	// IsError + Content[]. For the substrate's purposes we serialize
	// the whole thing so the completion pass can extract whichever
	// piece the agent needs.
	resultJSON, err := json.Marshal(map[string]any{
		"content":           contentToText(result.Content),
		"structuredContent": result.StructuredContent,
		"isError":           result.IsError,
	})
	if err != nil {
		writeError(ctx, pool, jobID, fmt.Errorf("encode result: %w", err))
		return
	}

	if result.IsError {
		// Tool-level errors (model called wrong tool, bad args)
		// are NOT bridge errors. Write status='done' so the parent
		// tool_dispatch sees the failure as a tool reply, not an
		// infrastructure failure.
		log.Printf("bridge run [%d] id=%d: %s/%s tool-error",
			workerID, jobID, payload.Server, payload.Tool)
	} else {
		log.Printf("bridge run [%d] id=%d: %s/%s ok",
			workerID, jobID, payload.Server, payload.Tool)
	}

	if _, err := pool.Exec(ctx,
		`UPDATE stewards.work_queue
		    SET status = 'done', result = $2::jsonb, done_at = now()
		  WHERE id = $1`,
		jobID, string(resultJSON),
	); err != nil {
		log.Printf("bridge run [%d] id=%d: write result failed: %v", workerID, jobID, err)
		return
	}
	// Hint the bgworker tick loop. Optional — tick still picks it up
	// within 500ms — but reduces tail latency.
	_, _ = pool.Exec(ctx, fmt.Sprintf("NOTIFY stewards_done, '%d'", jobID))
}

// contentToText concatenates MCP Content into a single string the
// model can read. Most tools return one TextContent; we handle the
// degenerate cases (empty, multiple parts, non-text) gracefully.
func contentToText(content []mcp.Content) string {
	if len(content) == 0 {
		return ""
	}
	if len(content) == 1 {
		if tc, ok := content[0].(*mcp.TextContent); ok {
			return tc.Text
		}
	}
	// Multi-part or non-text: marshal everything as JSON so nothing
	// is silently dropped.
	b, err := json.Marshal(content)
	if err != nil {
		return fmt.Sprintf("<encode content: %v>", err)
	}
	return string(b)
}

// writeError records a bridge-side failure (anything that isn't a
// normal tool reply). Sets status='error' and leaves a structured
// error column so the completion pass can synthesize a meaningful
// tool reply for the model.
func writeError(ctx context.Context, pool *pgxpool.Pool, id int64, e error) {
	log.Printf("bridge run id=%d: %v", id, e)
	_, _ = pool.Exec(ctx,
		`UPDATE stewards.work_queue
		    SET status = 'error', error = $2, done_at = now()
		  WHERE id = $1`,
		id, e.Error(),
	)
	_, _ = pool.Exec(ctx, fmt.Sprintf("NOTIFY stewards_done, '%d'", id))
}

// reapStaleMcpProxyRows handles the bridge's startup recovery pass.
// Any in_progress mcp_proxy rows from a previous bridge run are
// orphans — there's no way to know if their underlying call
// completed. Mark errored so waiting tool_dispatch parents can
// proceed (with an error reply, but at least they unblock).
func reapStaleMcpProxyRows(ctx context.Context, pool *pgxpool.Pool) error {
	tag, err := pool.Exec(ctx,
		`UPDATE stewards.work_queue
		    SET status  = 'error',
		        error   = coalesce(error,'') || 'bridge crashed before completion (stale in_progress reaped)',
		        done_at = now()
		  WHERE kind = 'mcp_proxy' AND status = 'in_progress'`,
	)
	if err != nil {
		return err
	}
	if n := tag.RowsAffected(); n > 0 {
		log.Printf("bridge run: reaped %d stale mcp_proxy row(s)", n)
	}
	return nil
}

// ---------------------------------------------------------------------
// Session cache
// ---------------------------------------------------------------------

type sessionCache struct {
	mu       sync.Mutex
	sessions map[string]*mcp.ClientSession
}

func newSessionCache() *sessionCache {
	return &sessionCache{
		sessions: make(map[string]*mcp.ClientSession),
	}
}

// get returns a connected session for the given server, lazily
// initializing one if needed. The double-check pattern with a single
// mutex is fine here — we don't expect to thrash, and avoiding a
// per-server mutex keeps the type simple.
func (c *sessionCache) get(ctx context.Context, pool *pgxpool.Pool, name string) (*mcp.ClientSession, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if s, ok := c.sessions[name]; ok {
		return s, nil
	}

	row, err := loadOneServer(ctx, pool, name)
	if err != nil {
		return nil, fmt.Errorf("load server %s: %w", name, err)
	}
	if row == nil {
		return nil, fmt.Errorf("server %s not in registry", name)
	}

	transport, err := buildClientTransport(*row)
	if err != nil {
		return nil, err
	}

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "pg-ai-stewards-bridge",
		Version: version,
	}, nil)
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, fmt.Errorf("connect %s: %w", name, err)
	}
	c.sessions[name] = session
	log.Printf("bridge run: session opened for %s", name)
	return session, nil
}

func (c *sessionCache) closeAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for name, s := range c.sessions {
		_ = s.Close()
		log.Printf("bridge run: session closed for %s", name)
	}
	c.sessions = nil
}

// loadOneServer pulls a single mcp_servers row by name. Returns nil
// if the row doesn't exist; error only on actual DB failure.
func loadOneServer(ctx context.Context, pool *pgxpool.Pool, name string) (*mcpServerRow, error) {
	var r mcpServerRow
	var envJSON []byte
	err := pool.QueryRow(ctx,
		"SELECT name, transport, command, args, url, env FROM stewards.mcp_servers WHERE name = $1 AND enabled",
		name,
	).Scan(&r.Name, &r.Transport, &r.Command, &r.Args, &r.URL, &envJSON)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(envJSON) > 0 {
		if err := json.Unmarshal(envJSON, &r.Env); err != nil {
			return nil, fmt.Errorf("env unmarshal: %w", err)
		}
	}
	return &r, nil
}

