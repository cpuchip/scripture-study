// bridge endpoints — surface the mcp_bridge_state view + the per-server
// tool catalog from mcp_tool_cache.

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

func (d *Deps) registerBridge(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/bridge/state", d.bridgeStateHandler)
}

type serverState struct {
	Server             string          `json:"server"`
	Transport          string          `json:"transport"`
	Enabled            bool            `json:"enabled"`
	LastHealthCheckAt  *time.Time      `json:"last_health_check_at,omitempty"`
	LastToolsRefreshAt *time.Time      `json:"last_tools_refresh_at,omitempty"`
	ActiveTools        int             `json:"active_tools"`
	LastError          string          `json:"last_error,omitempty"`
	Tools              []toolBrief     `json:"tools,omitempty"`
}

type toolBrief struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Active      bool            `json:"active"`
	InputSchema json.RawMessage `json:"input_schema,omitempty"`
}

type bridgeStateResp struct {
	Servers []serverState `json:"servers"`
}

func (d *Deps) bridgeStateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rows, err := d.Pool.Query(ctx,
		`SELECT server, transport, enabled,
		        last_health_check_at, last_tools_refresh_at,
		        active_tools, last_error
		   FROM stewards.mcp_bridge_state
		   ORDER BY server`,
	)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	resp := bridgeStateResp{}
	for rows.Next() {
		var s serverState
		var lastErr *string
		if err := rows.Scan(&s.Server, &s.Transport, &s.Enabled,
			&s.LastHealthCheckAt, &s.LastToolsRefreshAt,
			&s.ActiveTools, &lastErr); err == nil {
			if lastErr != nil {
				s.LastError = *lastErr
			}
			resp.Servers = append(resp.Servers, s)
		}
	}

	// Pull tool catalog per server
	toolRows, err := d.Pool.Query(ctx,
		`SELECT server_name, tool_name, coalesce(description,''),
		        active, input_schema
		   FROM stewards.mcp_tool_cache
		   ORDER BY server_name, tool_name`,
	)
	if err == nil {
		defer toolRows.Close()
		toolsByServer := map[string][]toolBrief{}
		for toolRows.Next() {
			var server string
			var t toolBrief
			if err := toolRows.Scan(&server, &t.Name, &t.Description, &t.Active, &t.InputSchema); err == nil {
				toolsByServer[server] = append(toolsByServer[server], t)
			}
		}
		for i := range resp.Servers {
			resp.Servers[i].Tools = toolsByServer[resp.Servers[i].Server]
		}
	}

	writeJSON(w, http.StatusOK, resp)
}
