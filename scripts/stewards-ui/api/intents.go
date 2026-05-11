// intents endpoints — list, get, create-inline. Phase 5d (C.7).
// Backs Stewards-UI /intents route + NewWork's intent picker.

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

func (d *Deps) registerIntents(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/intents/list", d.intentsListHandler)
	mux.HandleFunc("GET /api/intents/get", d.intentsGetHandler)
	mux.HandleFunc("POST /api/intents/create", d.intentsCreateHandler)
}

type intentRow struct {
	ID              string          `json:"id"`
	Slug            string          `json:"slug"`
	Purpose         string          `json:"purpose"`
	Beneficiary     string          `json:"beneficiary,omitempty"`
	ValuesHierarchy json.RawMessage `json:"values_hierarchy"`
	NonGoals        []string        `json:"non_goals,omitempty"`
	ScriptureAnchor string          `json:"scripture_anchor,omitempty"`
	SourceFile      string          `json:"source_file,omitempty"`
	WorkItemCount   int             `json:"work_item_count"`
	CreatedAt       *time.Time      `json:"created_at,omitempty"`
	UpdatedAt       *time.Time      `json:"updated_at,omitempty"`
}

type intentsListResp struct {
	Items []intentRow `json:"items"`
	Total int         `json:"total"`
}

func (d *Deps) intentsListHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp := intentsListResp{Items: []intentRow{}}
	rows, err := d.Pool.Query(ctx, `
		SELECT i.id::text, i.slug, i.purpose,
		       coalesce(i.beneficiary, ''),
		       coalesce(i.values_hierarchy, '[]'::jsonb),
		       coalesce(i.non_goals, ARRAY[]::text[]),
		       coalesce(i.scripture_anchor, ''),
		       coalesce(i.source_file, ''),
		       (SELECT count(*)::int FROM stewards.work_items wi WHERE wi.intent_id = i.id),
		       i.created_at, i.updated_at
		  FROM stewards.intents i
		  ORDER BY i.slug
	`)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		var ir intentRow
		if err := rows.Scan(&ir.ID, &ir.Slug, &ir.Purpose, &ir.Beneficiary,
			&ir.ValuesHierarchy, &ir.NonGoals, &ir.ScriptureAnchor,
			&ir.SourceFile, &ir.WorkItemCount, &ir.CreatedAt, &ir.UpdatedAt); err == nil {
			resp.Items = append(resp.Items, ir)
		}
	}
	resp.Total = len(resp.Items)
	writeJSON(w, http.StatusOK, resp)
}

func (d *Deps) intentsGetHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id := r.URL.Query().Get("id")
	slug := r.URL.Query().Get("slug")
	if id == "" && slug == "" {
		writeErr(w, http.StatusBadRequest, "id or slug query param required")
		return
	}

	var (
		ir       intentRow
		whereSQL = "i.id::text = $1"
		whereArg any = id
	)
	if id == "" {
		whereSQL = "i.slug = $1"
		whereArg = slug
	}

	err := d.Pool.QueryRow(ctx, `
		SELECT i.id::text, i.slug, i.purpose,
		       coalesce(i.beneficiary, ''),
		       coalesce(i.values_hierarchy, '[]'::jsonb),
		       coalesce(i.non_goals, ARRAY[]::text[]),
		       coalesce(i.scripture_anchor, ''),
		       coalesce(i.source_file, ''),
		       (SELECT count(*)::int FROM stewards.work_items wi WHERE wi.intent_id = i.id),
		       i.created_at, i.updated_at
		  FROM stewards.intents i
		 WHERE `+whereSQL,
		whereArg,
	).Scan(&ir.ID, &ir.Slug, &ir.Purpose, &ir.Beneficiary,
		&ir.ValuesHierarchy, &ir.NonGoals, &ir.ScriptureAnchor,
		&ir.SourceFile, &ir.WorkItemCount, &ir.CreatedAt, &ir.UpdatedAt)
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ir)
}

type intentCreateReq struct {
	Slug            string   `json:"slug"`
	Purpose         string   `json:"purpose"`
	Beneficiary     string   `json:"beneficiary,omitempty"`
	NonGoals        []string `json:"non_goals,omitempty"`
	ScriptureAnchor string   `json:"scripture_anchor,omitempty"`
}

type intentCreateResp struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
}

// intentsCreateHandler — inline-create from NewWork's "create new intent…" flow.
// Substrate-native intents (no source_file) created here are NOT in YAML.
// Per D-C1: YAML is canonical for repo-tracked intents; substrate-native
// intents land here for one-off work.
func (d *Deps) intentsCreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req intentCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "decode body: "+err.Error())
		return
	}
	if req.Slug == "" || req.Purpose == "" {
		writeErr(w, http.StatusBadRequest, "slug and purpose required")
		return
	}

	var (
		newID    string
		nonGoals []string = req.NonGoals
	)
	if nonGoals == nil {
		nonGoals = []string{}
	}
	err := d.Pool.QueryRow(ctx, `
		INSERT INTO stewards.intents
		    (slug, purpose, beneficiary, non_goals, scripture_anchor)
		VALUES ($1, $2, NULLIF($3, ''), $4, NULLIF($5, ''))
		RETURNING id::text
	`, req.Slug, req.Purpose, req.Beneficiary, nonGoals, req.ScriptureAnchor,
	).Scan(&newID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "create: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, intentCreateResp{ID: newID, Slug: req.Slug})
}
