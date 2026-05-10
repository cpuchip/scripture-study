// work_items endpoints — list, get. Sessions live in their own file.

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

func (d *Deps) registerWorkItems(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/work-items/list", d.workItemsListHandler)
	mux.HandleFunc("GET /api/work-items/get",  d.workItemsGetHandler)
}

type workItemRow struct {
	ID            string     `json:"id"`
	Slug          string     `json:"slug"`
	Pipeline      string     `json:"pipeline"`
	CurrentStage  string     `json:"current_stage"`
	Status        string     `json:"status"`
	Actor         string     `json:"actor,omitempty"`
	TokensIn      int        `json:"tokens_in"`
	TokensOut     int        `json:"tokens_out"`
	TokenBudget   *int       `json:"token_budget,omitempty"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}

type workItemsListResp struct {
	Items []workItemRow `json:"items"`
	Total int           `json:"total"`
}

func (d *Deps) workItemsListHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	q := r.URL.Query()
	pipeline := q.Get("pipeline")
	status := q.Get("status")
	limit := atoiDefault(q.Get("limit"), 100, 1, 500)

	whereClauses := []string{}
	args := []any{}
	if pipeline != "" {
		args = append(args, pipeline)
		whereClauses = append(whereClauses, "pipeline_family = $"+itoa(len(args)))
	}
	if status != "" {
		args = append(args, status)
		whereClauses = append(whereClauses, "status = $"+itoa(len(args)))
	}
	where := ""
	if len(whereClauses) > 0 {
		where = " WHERE " + joinAnd(whereClauses)
	}

	resp := workItemsListResp{}
	if err := d.Pool.QueryRow(ctx,
		"SELECT count(*) FROM stewards.work_items"+where, args...,
	).Scan(&resp.Total); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	args = append(args, limit)
	rows, err := d.Pool.Query(ctx,
		`SELECT id::text, slug, pipeline_family, current_stage, status,
		        coalesce(actor, ''),
		        coalesce(tokens_in, 0), coalesce(tokens_out, 0),
		        token_budget,
		        created_at, updated_at, completed_at
		   FROM stewards.work_items`+where+
			` ORDER BY updated_at DESC NULLS LAST LIMIT $`+itoa(len(args)),
		args...,
	)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		var w workItemRow
		if err := rows.Scan(&w.ID, &w.Slug, &w.Pipeline, &w.CurrentStage, &w.Status,
			&w.Actor, &w.TokensIn, &w.TokensOut, &w.TokenBudget,
			&w.CreatedAt, &w.UpdatedAt, &w.CompletedAt); err == nil {
			resp.Items = append(resp.Items, w)
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

type workItemDetail struct {
	workItemRow
	Input        json.RawMessage `json:"input"`
	StageResults json.RawMessage `json:"stage_results"`
	SessionIDs   []string        `json:"session_ids,omitempty"`
	Error        string          `json:"error,omitempty"`
}

func (d *Deps) workItemsGetHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id := r.URL.Query().Get("id")
	slug := r.URL.Query().Get("slug")
	if id == "" && slug == "" {
		writeErr(w, http.StatusBadRequest, "id or slug query param required")
		return
	}

	var (
		wd       workItemDetail
		whereSQL = "id::text = $1"
		whereArg any = id
	)
	if id == "" {
		whereSQL = "slug = $1"
		whereArg = slug
	}
	err := d.Pool.QueryRow(ctx,
		`SELECT id::text, slug, pipeline_family, current_stage, status,
		        coalesce(actor, ''),
		        coalesce(tokens_in, 0), coalesce(tokens_out, 0),
		        token_budget,
		        created_at, updated_at, completed_at,
		        input, stage_results,
		        coalesce(session_ids, ARRAY[]::text[]),
		        coalesce(error, '')
		   FROM stewards.work_items
		  WHERE `+whereSQL,
		whereArg,
	).Scan(&wd.ID, &wd.Slug, &wd.Pipeline, &wd.CurrentStage, &wd.Status,
		&wd.Actor, &wd.TokensIn, &wd.TokensOut, &wd.TokenBudget,
		&wd.CreatedAt, &wd.UpdatedAt, &wd.CompletedAt,
		&wd.Input, &wd.StageResults, &wd.SessionIDs, &wd.Error)
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, wd)
}

// Tiny string helpers — keep local to avoid pulling strings/strconv
// imports in a handler file.
func itoa(n int) string {
	const digits = "0123456789"
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := []byte{}
	for n > 0 {
		buf = append([]byte{digits[n%10]}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}
func joinAnd(ss []string) string {
	if len(ss) == 0 {
		return ""
	}
	out := ss[0]
	for _, s := range ss[1:] {
		out += " AND " + s
	}
	return out
}
