// Bridge mode (Phase 3e.2.a, 2026-05-08): outbound MCP-client capability.
//
// The substrate's `stewards.mcp_servers` table registers external MCP
// servers we may consume (gospel-engine-v2, webster, exa-search, ...).
// The bridge connects to those servers using the official Go MCP SDK
// as a client, calls tools/list, and upserts results into
// `stewards.mcp_tool_cache`. Substrate-internal AI agents (running
// inside pipeline work_items) can then route tool calls through the
// bridge instead of being trapped inside Postgres.
//
// Tonight's scope is just `bridge refresh-tools` — a one-shot
// discovery pass. The long-running daemon mode (`bridge run`) that
// listens on `stewards_mcp_proxy` notifications and dispatches calls
// is deferred to 3e.2.b/c.
//
// Stderr discipline does NOT apply here — bridge mode does not own
// stdout for protocol traffic, so normal logging is fine.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// resolveSecret expands `$env:VAR` indirection. The seed rows in
// 3e2-1-mcp-bridge-schemas.sql store secret values as
// `$env:GOSPEL_ENGINE_TOKEN`, `$env:BECOMING_TOKEN`, etc. The bridge
// resolves them against its own process environment at connect time.
// Anything that doesn't start with `$env:` passes through unchanged.
func resolveSecret(value string) string {
	const prefix = "$env:"
	if !strings.HasPrefix(value, prefix) {
		return value
	}
	name := strings.TrimPrefix(value, prefix)
	resolved := os.Getenv(name)
	if resolved == "" {
		// Leave the placeholder visible — the call will likely fail,
		// but the failure mode is clearer than silently sending the
		// literal "$env:..." as a header.
		return value
	}
	return resolved
}

// mcpServerRow mirrors the relevant columns from stewards.mcp_servers.
type mcpServerRow struct {
	Name      string
	Transport string
	Command   *string
	Args      []string
	URL       *string
	Env       map[string]string
}

// runBridge dispatches `bridge <subcommand> [args]`. Today we only
// implement `refresh-tools`; future actions will include `run` for
// the long-lived daemon and `health` for ad-hoc connectivity checks.
func runBridge(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: stewards-mcp bridge <action> [flags]\n  actions: refresh-tools")
	}
	action, rest := args[0], args[1:]
	switch action {
	case "refresh-tools":
		return runBridgeRefreshTools(rest)
	default:
		return fmt.Errorf("unknown bridge action: %s", action)
	}
}

func runBridgeRefreshTools(args []string) error {
	fs := flag.NewFlagSet("bridge refresh-tools", flag.ContinueOnError)
	dsn := fs.String("dsn", "",
		"Postgres DSN (default: $STEWARDS_DSN, then localhost compose port 55433)")
	timeoutSecs := fs.Int("timeout", 30,
		"Per-server connect+list timeout in seconds")
	includeAll := fs.Bool("all", false,
		"Refresh all servers, including those with enabled=false")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *dsn == "" {
		*dsn = os.Getenv("STEWARDS_DSN")
	}
	if *dsn == "" {
		*dsn = "postgres://stewards:stewards@localhost:55433/stewards?sslmode=disable"
	}

	rootCtx, cancel := context.WithTimeout(context.Background(),
		time.Duration(*timeoutSecs)*time.Duration(8)*time.Second)
	defer cancel()

	pool, err := pgxpool.New(rootCtx, *dsn)
	if err != nil {
		return fmt.Errorf("pgxpool.New: %w", err)
	}
	defer pool.Close()
	if err := pool.Ping(rootCtx); err != nil {
		return fmt.Errorf("pool.Ping: %w", err)
	}

	servers, err := loadMcpServers(rootCtx, pool, *includeAll)
	if err != nil {
		return fmt.Errorf("load servers: %w", err)
	}
	if len(servers) == 0 {
		fmt.Printf("No servers to refresh (use --all to include disabled rows)\n")
		return nil
	}
	fmt.Printf("Refreshing %d MCP server(s)\n", len(servers))

	successCount := 0
	for _, srv := range servers {
		ctx, srvCancel := context.WithTimeout(rootCtx,
			time.Duration(*timeoutSecs)*time.Second)
		err := refreshOneServer(ctx, pool, srv)
		srvCancel()

		if err != nil {
			fmt.Printf("  [FAIL] %-20s %s\n", srv.Name, err)
			recordHealthFailure(rootCtx, pool, srv.Name, err)
			continue
		}
		successCount++
	}

	fmt.Printf("\nRefresh complete: %d/%d successful\n",
		successCount, len(servers))
	if successCount < len(servers) {
		return fmt.Errorf("%d server(s) failed", len(servers)-successCount)
	}
	return nil
}

// loadMcpServers reads the registry. By default returns only enabled
// rows so a `refresh-tools` run on a freshly-seeded substrate (where
// all rows are enabled=false) requires `--all` to do anything — that
// matches how the seed comment phrased it ("operator flips them to
// true once they verify the bridge can reach them").
func loadMcpServers(ctx context.Context, pool *pgxpool.Pool, includeAll bool) ([]mcpServerRow, error) {
	q := "SELECT name, transport, command, args, url, env FROM stewards.mcp_servers"
	if !includeAll {
		q += " WHERE enabled"
	}
	q += " ORDER BY name"

	rows, err := pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []mcpServerRow
	for rows.Next() {
		var r mcpServerRow
		var envJSON []byte
		if err := rows.Scan(&r.Name, &r.Transport, &r.Command, &r.Args, &r.URL, &envJSON); err != nil {
			return nil, err
		}
		if len(envJSON) > 0 {
			if err := json.Unmarshal(envJSON, &r.Env); err != nil {
				return nil, fmt.Errorf("server %s: env unmarshal: %w", r.Name, err)
			}
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// refreshOneServer connects to a single MCP server, calls tools/list,
// and upserts the results. It also stamps last_health_check_at and
// last_tools_refresh_at on success, clearing last_error.
func refreshOneServer(ctx context.Context, pool *pgxpool.Pool, srv mcpServerRow) error {
	transport, err := buildClientTransport(srv)
	if err != nil {
		return err
	}

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "pg-ai-stewards-bridge",
		Version: version,
	}, nil)

	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer session.Close()

	result, err := session.ListTools(ctx, nil)
	if err != nil {
		return fmt.Errorf("tools/list: %w", err)
	}

	toolNames := make([]string, 0, len(result.Tools))
	if err := upsertToolCache(ctx, pool, srv.Name, result.Tools, &toolNames); err != nil {
		return fmt.Errorf("upsert tool cache: %w", err)
	}

	if _, err := pool.Exec(ctx,
		`UPDATE stewards.mcp_servers
		    SET last_health_check_at  = now(),
		        last_tools_refresh_at = now(),
		        last_error            = NULL,
		        updated_at            = now()
		  WHERE name = $1`,
		srv.Name,
	); err != nil {
		return fmt.Errorf("update health: %w", err)
	}

	fmt.Printf("  [ OK ] %-20s %d tool(s): %s\n",
		srv.Name, len(toolNames), strings.Join(toolNames, ", "))
	return nil
}

// buildClientTransport wires up either CommandTransport (stdio) or
// StreamableClientTransport (http) based on the row's transport
// column. Secrets in env are resolved via $env: indirection.
func buildClientTransport(srv mcpServerRow) (mcp.Transport, error) {
	switch srv.Transport {
	case "stdio":
		if srv.Command == nil || *srv.Command == "" {
			return nil, fmt.Errorf("stdio transport requires command")
		}
		cmd := exec.Command(*srv.Command, srv.Args...)
		cmd.Env = os.Environ()
		for k, v := range srv.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, resolveSecret(v)))
		}
		// Pipe child stderr to ours so spawned-server diagnostics
		// land in the bridge's log instead of being silently dropped.
		cmd.Stderr = os.Stderr
		return &mcp.CommandTransport{Command: cmd}, nil

	case "http":
		if srv.URL == nil || *srv.URL == "" {
			return nil, fmt.Errorf("http transport requires url")
		}
		// Note: the StreamableClientTransport doesn't expose a Headers
		// field directly — auth has to flow through the http.Client
		// (e.g., a custom RoundTripper that injects bearer tokens). For
		// servers that need bearer auth (none of our seed rows for
		// http transport do today — exa-search uses ?token= in URL),
		// extend this branch to wrap http.DefaultTransport. Keep it
		// minimal until we hit a server that needs it.
		return &mcp.StreamableClientTransport{Endpoint: *srv.URL}, nil

	default:
		return nil, fmt.Errorf("unknown transport: %s", srv.Transport)
	}
}

// upsertToolCache writes one row per tool. Existing rows for tools no
// longer present on the server are marked active=false (not deleted)
// so historical schema is preserved.
func upsertToolCache(ctx context.Context, pool *pgxpool.Pool, serverName string, tools []*mcp.Tool, names *[]string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	currentNames := make([]string, 0, len(tools))
	for _, t := range tools {
		currentNames = append(currentNames, t.Name)
		*names = append(*names, t.Name)

		inputSchemaJSON, err := jsonOrEmpty(t.InputSchema)
		if err != nil {
			return fmt.Errorf("tool %s input_schema: %w", t.Name, err)
		}
		var outputSchemaJSON *string
		if t.OutputSchema != nil {
			s, err := jsonOrEmpty(t.OutputSchema)
			if err != nil {
				return fmt.Errorf("tool %s output_schema: %w", t.Name, err)
			}
			outputSchemaJSON = &s
		}
		var titlePtr *string
		if t.Title != "" {
			titlePtr = &t.Title
		}

		if _, err := tx.Exec(ctx,
			`INSERT INTO stewards.mcp_tool_cache
			   (server_name, tool_name, description, title,
			    input_schema, output_schema, last_refreshed_at, active)
			 VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb, now(), true)
			 ON CONFLICT (server_name, tool_name) DO UPDATE
			   SET description       = EXCLUDED.description,
			       title             = EXCLUDED.title,
			       input_schema      = EXCLUDED.input_schema,
			       output_schema     = EXCLUDED.output_schema,
			       last_refreshed_at = now(),
			       active            = true`,
			serverName, t.Name, t.Description, titlePtr,
			inputSchemaJSON, outputSchemaJSON,
		); err != nil {
			return fmt.Errorf("upsert tool %s: %w", t.Name, err)
		}
	}

	// Soft-deactivate cached tools the server no longer reports. The
	// `= ANY(...)` filter wants a non-empty list; pad with a sentinel
	// when the server returned zero tools (degenerate but legal).
	deactivateNames := currentNames
	if len(deactivateNames) == 0 {
		deactivateNames = []string{""}
	}
	if _, err := tx.Exec(ctx,
		`UPDATE stewards.mcp_tool_cache
		    SET active = false
		  WHERE server_name = $1
		    AND active = true
		    AND tool_name <> ALL($2)`,
		serverName, deactivateNames,
	); err != nil {
		return fmt.Errorf("deactivate stale tools: %w", err)
	}

	return tx.Commit(ctx)
}

func jsonOrEmpty(v any) (string, error) {
	if v == nil {
		return "{}", nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// recordHealthFailure stamps last_health_check_at and last_error so
// the substrate's mcp_bridge_state view reflects the most recent
// diagnosis. Best-effort — we don't fail the whole refresh on this.
func recordHealthFailure(ctx context.Context, pool *pgxpool.Pool, name string, err error) {
	_, _ = pool.Exec(ctx,
		`UPDATE stewards.mcp_servers
		    SET last_health_check_at = now(),
		        last_error           = $2,
		        updated_at           = now()
		  WHERE name = $1`,
		name, err.Error(),
	)
}
