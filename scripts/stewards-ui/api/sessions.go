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

type chatDispatch struct {
	WorkID         int64           `json:"work_id"`
	Provider       string          `json:"provider"`
	Model          string          `json:"model,omitempty"`
	AgentFamily    string          `json:"agent_family,omitempty"`
	SystemPrompt   string          `json:"system_prompt,omitempty"`
	Tools          json.RawMessage `json:"tools,omitempty"`
	MessagesCount  int             `json:"messages_count"`
	BodyMessages   json.RawMessage `json:"body_messages,omitempty"`
	Status         string          `json:"status"`
	CreatedAt      *time.Time      `json:"created_at,omitempty"`
	DoneAt         *time.Time      `json:"done_at,omitempty"`
}

type sessionDetail struct {
	SessionID  string         `json:"session_id"`
	Messages   []messageRow   `json:"messages"`
	Dispatches []chatDispatch `json:"dispatches"`
	TokensIn   int            `json:"tokens_in"`
	TokensOut  int            `json:"tokens_out"`
}

func (d *Deps) sessionsGetHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	sid := r.URL.Query().Get("id")
	if sid == "" {
		writeErr(w, http.StatusBadRequest, "id (session_id) query param required")
		return
	}

	resp := sessionDetail{
		SessionID:  sid,
		Messages:   []messageRow{},
		Dispatches: []chatDispatch{},
	}

	// Pull chat dispatches for this session — each holds the full
	// payload.body that was sent to the model: system prompt,
	// composed tools array, accumulated message history. The
	// messages table only persists assistant + tool replies; the
	// system prompt and tools array live here.
	dispRows, err := d.Pool.Query(ctx,
		`SELECT id, provider,
		        coalesce(payload->'body'->>'model', ''),
		        coalesce(payload->>'agent_family', ''),
		        coalesce(payload->'body'->'messages'->0->>'content', ''),
		        payload->'body'->'tools',
		        coalesce(jsonb_array_length(payload->'body'->'messages'), 0),
		        payload->'body'->'messages',
		        status, created_at, done_at
		   FROM stewards.work_queue
		  WHERE kind = 'chat'
		    AND payload->>'session_id' = $1
		  ORDER BY id`,
		sid,
	)
	if err == nil {
		defer dispRows.Close()
		for dispRows.Next() {
			var di chatDispatch
			if err := dispRows.Scan(&di.WorkID, &di.Provider, &di.Model,
				&di.AgentFamily, &di.SystemPrompt, &di.Tools,
				&di.MessagesCount, &di.BodyMessages,
				&di.Status, &di.CreatedAt, &di.DoneAt); err == nil {
				resp.Dispatches = append(resp.Dispatches, di)
			}
		}
	}

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
