// trust + gate-override endpoints — Phase 5f (E.6).
// Backs Stewards-UI /trust route + override button on WorkItemDetail.

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

func (d *Deps) registerTrust(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/trust/scores", d.trustScoresHandler)
	mux.HandleFunc("GET /api/trust/transitions", d.trustTransitionsHandler)
	mux.HandleFunc("POST /api/trust/adjust", d.trustAdjustHandler)
	mux.HandleFunc("POST /api/gate-overrides/apply", d.gateOverrideApplyHandler)
}

type trustScoreRow struct {
	AgentFamily            string     `json:"agent_family"`
	PipelineFamily         string     `json:"pipeline_family"`
	Model                  string     `json:"model"`
	SuccessfulCompletions  int        `json:"successful_completions"`
	FailedCompletions      int        `json:"failed_completions"`
	HumanOverrides         int        `json:"human_overrides"`
	TrustLevel             string     `json:"trust_level"`
	LastEvaluatedAt        *time.Time `json:"last_evaluated_at,omitempty"`
	LastCompletionAt       *time.Time `json:"last_completion_at,omitempty"`
}

type trustScoresResp struct {
	Items []trustScoreRow `json:"items"`
	Total int             `json:"total"`
}

func (d *Deps) trustScoresHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rows, err := d.Pool.Query(ctx, `
		SELECT agent_family, pipeline_family, model,
		       successful_completions, failed_completions, human_overrides,
		       trust_level, last_evaluated_at, last_completion_at
		  FROM stewards.trust_scores
		  ORDER BY pipeline_family, agent_family, model`)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	resp := trustScoresResp{Items: []trustScoreRow{}}
	for rows.Next() {
		var t trustScoreRow
		if err := rows.Scan(&t.AgentFamily, &t.PipelineFamily, &t.Model,
			&t.SuccessfulCompletions, &t.FailedCompletions, &t.HumanOverrides,
			&t.TrustLevel, &t.LastEvaluatedAt, &t.LastCompletionAt); err == nil {
			resp.Items = append(resp.Items, t)
		}
	}
	resp.Total = len(resp.Items)
	writeJSON(w, http.StatusOK, resp)
}

type trustTransitionRow struct {
	ID              int64           `json:"id"`
	At              *time.Time      `json:"at,omitempty"`
	AgentFamily     string          `json:"agent_family"`
	PipelineFamily  string          `json:"pipeline_family"`
	Model           string          `json:"model"`
	FromLevel       string          `json:"from_level"`
	ToLevel         string          `json:"to_level"`
	TransitionKind  string          `json:"transition_kind"`
	Actor           string          `json:"actor"`
	Justification   string          `json:"justification,omitempty"`
	Metrics         json.RawMessage `json:"metrics,omitempty"`
}

type trustTransitionsResp struct {
	Items []trustTransitionRow `json:"items"`
	Total int                  `json:"total"`
}

func (d *Deps) trustTransitionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	q := r.URL.Query()
	whereClauses := []string{}
	args := []any{}
	if v := q.Get("agent"); v != "" {
		args = append(args, v)
		whereClauses = append(whereClauses, "agent_family = $"+itoa(len(args)))
	}
	if v := q.Get("pipeline"); v != "" {
		args = append(args, v)
		whereClauses = append(whereClauses, "pipeline_family = $"+itoa(len(args)))
	}
	if v := q.Get("model"); v != "" {
		args = append(args, v)
		whereClauses = append(whereClauses, "model = $"+itoa(len(args)))
	}
	where := ""
	if len(whereClauses) > 0 {
		where = " WHERE " + joinAnd(whereClauses)
	}
	limit := atoiDefault(q.Get("limit"), 100, 1, 500)
	args = append(args, limit)

	rows, err := d.Pool.Query(ctx, `
		SELECT id, at, agent_family, pipeline_family, model,
		       from_level, to_level, transition_kind, actor,
		       coalesce(justification, ''),
		       coalesce(metrics, '{}'::jsonb)
		  FROM stewards.trust_transitions`+where+`
		  ORDER BY at DESC
		  LIMIT $`+itoa(len(args)),
		args...)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	resp := trustTransitionsResp{Items: []trustTransitionRow{}}
	for rows.Next() {
		var t trustTransitionRow
		if err := rows.Scan(&t.ID, &t.At, &t.AgentFamily, &t.PipelineFamily, &t.Model,
			&t.FromLevel, &t.ToLevel, &t.TransitionKind, &t.Actor,
			&t.Justification, &t.Metrics); err == nil {
			resp.Items = append(resp.Items, t)
		}
	}
	resp.Total = len(resp.Items)
	writeJSON(w, http.StatusOK, resp)
}

type trustAdjustReq struct {
	AgentFamily    string `json:"agent_family"`
	PipelineFamily string `json:"pipeline_family"`
	Model          string `json:"model"`
	NewLevel       string `json:"new_level"`
	Actor          string `json:"actor"`
	Justification  string `json:"justification"`
}

type trustAdjustResp struct {
	NewLevel string `json:"new_level"`
}

func (d *Deps) trustAdjustHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req trustAdjustReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "decode body: "+err.Error())
		return
	}
	if req.AgentFamily == "" || req.PipelineFamily == "" || req.Model == "" {
		writeErr(w, http.StatusBadRequest, "agent_family, pipeline_family, model required")
		return
	}

	var newLevel string
	err := d.Pool.QueryRow(ctx,
		`SELECT stewards.trust_adjust($1, $2, $3, $4, $5, $6)`,
		req.AgentFamily, req.PipelineFamily, req.Model,
		req.NewLevel, req.Actor, req.Justification,
	).Scan(&newLevel)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, trustAdjustResp{NewLevel: newLevel})
}

type gateOverrideApplyReq struct {
	GateDecisionID int64  `json:"gate_decision_id"`
	OverriddenBy   string `json:"overridden_by"`
	NewAction      string `json:"new_action"`
	Justification  string `json:"justification"`
}

type gateOverrideApplyResp struct {
	NewMaturity string `json:"new_maturity"`
}

func (d *Deps) gateOverrideApplyHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req gateOverrideApplyReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "decode body: "+err.Error())
		return
	}

	var newMaturity string
	err := d.Pool.QueryRow(ctx,
		`SELECT stewards.apply_gate_override($1, $2, $3, $4)`,
		req.GateDecisionID, req.OverriddenBy, req.NewAction, req.Justification,
	).Scan(&newMaturity)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, gateOverrideApplyResp{NewMaturity: newMaturity})
}
