// watchman endpoints — list passes, pass detail with verdicts/findings.

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

func (d *Deps) registerWatchman(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/watchman/passes",   d.watchmanPassesHandler)
	mux.HandleFunc("GET /api/watchman/pass",     d.watchmanPassHandler)
}

type passRow struct {
	PassID            string          `json:"pass_id"`
	Status            string          `json:"status"`
	Trigger           string          `json:"trigger,omitempty"`
	StartedAt         *time.Time      `json:"started_at,omitempty"`
	FinishedAt        *time.Time      `json:"finished_at,omitempty"`
	Provider          string          `json:"provider,omitempty"`
	Model             string          `json:"model,omitempty"`
	AgentFamily       string          `json:"agent_family,omitempty"`
	DocCountPlanned   int             `json:"doc_count_planned"`
	DocCountDone      int             `json:"doc_count_done"`
	TokensIn          int             `json:"tokens_in"`
	TokensOut         int             `json:"tokens_out"`
	TokenBudget       *int            `json:"token_budget,omitempty"`
	BudgetStopped     bool            `json:"budget_stopped"`
	VerdictCounts     json.RawMessage `json:"verdict_counts,omitempty"`
}

type passesResp struct {
	Items []passRow `json:"items"`
}

func (d *Deps) watchmanPassesHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	limit := atoiDefault(r.URL.Query().Get("limit"), 25, 1, 200)

	rows, err := d.Pool.Query(ctx,
		`SELECT pass_id, status, coalesce(trigger,''), started_at, finished_at,
		        coalesce(provider,''), coalesce(model,''), coalesce(agent_family,''),
		        coalesce(doc_count_planned,0), coalesce(doc_count_done,0),
		        coalesce(tokens_in,0), coalesce(tokens_out,0), token_budget,
		        coalesce(budget_stopped, false),
		        verdict_counts
		   FROM stewards.watchman_passes
		   ORDER BY started_at DESC
		   LIMIT $1`,
		limit,
	)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	resp := passesResp{Items: []passRow{}}
	for rows.Next() {
		var p passRow
		if err := rows.Scan(&p.PassID, &p.Status, &p.Trigger, &p.StartedAt, &p.FinishedAt,
			&p.Provider, &p.Model, &p.AgentFamily,
			&p.DocCountPlanned, &p.DocCountDone,
			&p.TokensIn, &p.TokensOut, &p.TokenBudget,
			&p.BudgetStopped, &p.VerdictCounts); err == nil {
			resp.Items = append(resp.Items, p)
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

type verdictRow struct {
	StudyID   string     `json:"study_id"`
	Verdict   string     `json:"verdict"`
	Reasoning string     `json:"reasoning,omitempty"`
	Model     string     `json:"model,omitempty"`
	Tokens    int        `json:"tokens,omitempty"`
	Actor     string     `json:"actor,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

type passDetailResp struct {
	Pass     passRow      `json:"pass"`
	Verdicts []verdictRow `json:"verdicts"`
}

func (d *Deps) watchmanPassHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	passID := r.URL.Query().Get("id")
	if passID == "" {
		writeErr(w, http.StatusBadRequest, "id (pass_id) required")
		return
	}
	resp := passDetailResp{Pass: passRow{PassID: passID}}

	err := d.Pool.QueryRow(ctx,
		`SELECT pass_id, status, coalesce(trigger,''), started_at, finished_at,
		        coalesce(provider,''), coalesce(model,''), coalesce(agent_family,''),
		        coalesce(doc_count_planned,0), coalesce(doc_count_done,0),
		        coalesce(tokens_in,0), coalesce(tokens_out,0), token_budget,
		        coalesce(budget_stopped, false), verdict_counts
		   FROM stewards.watchman_passes WHERE pass_id=$1`,
		passID,
	).Scan(&resp.Pass.PassID, &resp.Pass.Status, &resp.Pass.Trigger,
		&resp.Pass.StartedAt, &resp.Pass.FinishedAt,
		&resp.Pass.Provider, &resp.Pass.Model, &resp.Pass.AgentFamily,
		&resp.Pass.DocCountPlanned, &resp.Pass.DocCountDone,
		&resp.Pass.TokensIn, &resp.Pass.TokensOut, &resp.Pass.TokenBudget,
		&resp.Pass.BudgetStopped, &resp.Pass.VerdictCounts)
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}

	rows, err := d.Pool.Query(ctx,
		`SELECT study_id::text, verdict, coalesce(reasoning,''),
		        coalesce(model,''), coalesce(tokens_in + tokens_out, 0),
		        coalesce(actor,''), created_at
		   FROM stewards.verdicts
		  WHERE pass_id=$1
		  ORDER BY created_at`,
		passID,
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var v verdictRow
			if err := rows.Scan(&v.StudyID, &v.Verdict, &v.Reasoning,
				&v.Model, &v.Tokens, &v.Actor, &v.CreatedAt); err == nil {
				resp.Verdicts = append(resp.Verdicts, v)
			}
		}
	}

	writeJSON(w, http.StatusOK, resp)
}
