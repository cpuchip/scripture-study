// sessions endpoint — get session detail (messages timeline).

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

func (d *Deps) registerSessions(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/sessions/get", d.sessionsGetHandler)
}

type messageRow struct {
	ID              int64           `json:"id"`
	Role            string          `json:"role"`
	Content         string          `json:"content"`
	Model           string          `json:"model,omitempty"`
	ToolCallID      string          `json:"tool_call_id,omitempty"`
	ToolCalls       json.RawMessage `json:"tool_calls,omitempty"`
	FinishReason    string          `json:"finish_reason,omitempty"`
	TokensIn        *int            `json:"tokens_in,omitempty"`
	TokensOut       *int            `json:"tokens_out,omitempty"`
	ReasoningTokens *int            `json:"reasoning_tokens,omitempty"`
	ParentWorkID    *int64          `json:"parent_work_id,omitempty"`
	CreatedAt       *time.Time      `json:"created_at,omitempty"`
}

type sessionDetail struct {
	SessionID string       `json:"session_id"`
	Messages  []messageRow `json:"messages"`
	TokensIn  int          `json:"tokens_in"`
	TokensOut int          `json:"tokens_out"`
}

func (d *Deps) sessionsGetHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	sid := r.URL.Query().Get("id")
	if sid == "" {
		writeErr(w, http.StatusBadRequest, "id (session_id) query param required")
		return
	}

	resp := sessionDetail{SessionID: sid}

	rows, err := d.Pool.Query(ctx,
		`SELECT id, role, coalesce(content, ''), coalesce(model, ''),
		        coalesce(tool_call_id, ''),
		        tool_calls,
		        coalesce(finish_reason, ''),
		        tokens_in, tokens_out, reasoning_tokens,
		        parent_work_id, created_at
		   FROM stewards.messages
		  WHERE session_id = $1
		  ORDER BY id`,
		sid,
	)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		var m messageRow
		if err := rows.Scan(&m.ID, &m.Role, &m.Content, &m.Model,
			&m.ToolCallID, &m.ToolCalls, &m.FinishReason,
			&m.TokensIn, &m.TokensOut, &m.ReasoningTokens,
			&m.ParentWorkID, &m.CreatedAt); err == nil {
			resp.Messages = append(resp.Messages, m)
			if m.TokensIn != nil {
				resp.TokensIn += *m.TokensIn
			}
			if m.TokensOut != nil {
				resp.TokensOut += *m.TokensOut
			}
		}
	}
	writeJSON(w, http.StatusOK, resp)
}
