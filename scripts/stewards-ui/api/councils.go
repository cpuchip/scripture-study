// councils endpoints — Phase 5g (F.6).
// Backs Stewards-UI /councils route + dashboard suggestion banner.

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

func (d *Deps) registerCouncils(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/councils/list", d.councilsListHandler)
	mux.HandleFunc("GET /api/councils/get", d.councilsGetHandler)
	mux.HandleFunc("POST /api/councils/convene", d.councilsConveneHandler)
	mux.HandleFunc("POST /api/councils/resolve", d.councilsResolveHandler)
	mux.HandleFunc("GET /api/councils/suggestions", d.councilsSuggestHandler)
}

type councilRow struct {
	ID              string     `json:"id"`
	IntentID        string     `json:"intent_id"`
	IntentSlug      string     `json:"intent_slug,omitempty"`
	BindingQuestion string     `json:"binding_question"`
	ConvenedAt      *time.Time `json:"convened_at,omitempty"`
	ConvenedBy      string     `json:"convened_by"`
	Bishop          string     `json:"bishop"`
	Status          string     `json:"status"`
	ResolutionID    *string    `json:"resolution_id,omitempty"`
	DissolvedReason string     `json:"dissolved_reason,omitempty"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
}

type councilsListResp struct {
	Items []councilRow `json:"items"`
	Total int          `json:"total"`
}

func (d *Deps) councilsListHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	q := r.URL.Query()
	limit := atoiDefault(q.Get("limit"), 50, 1, 200)

	rows, err := d.Pool.Query(ctx, `
		SELECT c.id::text, c.intent_id::text, coalesce(i.slug, ''),
		       c.binding_question, c.convened_at, c.convened_by,
		       c.bishop, c.status, c.resolution_id::text,
		       coalesce(c.dissolved_reason, ''), c.resolved_at
		  FROM stewards.councils c
		  LEFT JOIN stewards.intents i ON i.id = c.intent_id
		  ORDER BY c.convened_at DESC
		  LIMIT $1`, limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	resp := councilsListResp{Items: []councilRow{}}
	for rows.Next() {
		var c councilRow
		var resID *string
		if err := rows.Scan(&c.ID, &c.IntentID, &c.IntentSlug,
			&c.BindingQuestion, &c.ConvenedAt, &c.ConvenedBy,
			&c.Bishop, &c.Status, &resID,
			&c.DissolvedReason, &c.ResolvedAt); err == nil {
			if resID != nil && *resID != "" {
				c.ResolutionID = resID
			}
			resp.Items = append(resp.Items, c)
		}
	}
	resp.Total = len(resp.Items)
	writeJSON(w, http.StatusOK, resp)
}

type councilMember struct {
	AgentFamily string     `json:"agent_family"`
	Role        string     `json:"role"`
	WorkID      *int64     `json:"work_id,omitempty"`
	Response    string     `json:"response,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type councilResolution struct {
	ID         string          `json:"id"`
	ResolvedBy string          `json:"resolved_by"`
	Text       string          `json:"text"`
	PromotedTo string          `json:"promoted_to,omitempty"`
	PromotedAt *time.Time      `json:"promoted_at,omitempty"`
	RawProposal json.RawMessage `json:"raw_proposal,omitempty"`
	ResolvedAt *time.Time      `json:"resolved_at,omitempty"`
}

type councilDetail struct {
	councilRow
	IntentPurpose string             `json:"intent_purpose,omitempty"`
	Members       []councilMember    `json:"members"`
	Resolution    *councilResolution `json:"resolution,omitempty"`
}

func (d *Deps) councilsGetHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id := r.URL.Query().Get("id")
	if id == "" {
		writeErr(w, http.StatusBadRequest, "id required")
		return
	}

	var (
		c     councilDetail
		resID *string
	)
	err := d.Pool.QueryRow(ctx, `
		SELECT c.id::text, c.intent_id::text, coalesce(i.slug, ''),
		       coalesce(i.purpose, ''),
		       c.binding_question, c.convened_at, c.convened_by,
		       c.bishop, c.status, c.resolution_id::text,
		       coalesce(c.dissolved_reason, ''), c.resolved_at
		  FROM stewards.councils c
		  LEFT JOIN stewards.intents i ON i.id = c.intent_id
		 WHERE c.id::text = $1`,
		id,
	).Scan(&c.ID, &c.IntentID, &c.IntentSlug, &c.IntentPurpose,
		&c.BindingQuestion, &c.ConvenedAt, &c.ConvenedBy,
		&c.Bishop, &c.Status, &resID,
		&c.DissolvedReason, &c.ResolvedAt)
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	if resID != nil && *resID != "" {
		c.ResolutionID = resID
	}

	// Members
	c.Members = []councilMember{}
	memRows, err := d.Pool.Query(ctx, `
		SELECT agent_family, role, work_id,
		       coalesce(response, ''), completed_at
		  FROM stewards.council_members
		 WHERE council_id::text = $1
		 ORDER BY role, agent_family`, id)
	if err == nil {
		defer memRows.Close()
		for memRows.Next() {
			var m councilMember
			if err := memRows.Scan(&m.AgentFamily, &m.Role, &m.WorkID, &m.Response, &m.CompletedAt); err == nil {
				c.Members = append(c.Members, m)
			}
		}
	}

	// Resolution (if any)
	if c.ResolutionID != nil {
		var res councilResolution
		err := d.Pool.QueryRow(ctx, `
			SELECT id::text, resolved_by, text,
			       coalesce(promoted_to, ''), promoted_at,
			       coalesce(raw_proposal, '{}'::jsonb), resolved_at
			  FROM stewards.resolutions
			 WHERE id::text = $1`,
			*c.ResolutionID,
		).Scan(&res.ID, &res.ResolvedBy, &res.Text,
			&res.PromotedTo, &res.PromotedAt, &res.RawProposal, &res.ResolvedAt)
		if err == nil {
			c.Resolution = &res
		}
	}

	writeJSON(w, http.StatusOK, c)
}

type councilConveneReq struct {
	IntentID        string          `json:"intent_id"`
	BindingQuestion string          `json:"binding_question"`
	Members         json.RawMessage `json:"members"`
	Bishop          string          `json:"bishop"`
	ConvenedBy      string          `json:"convened_by,omitempty"`
}

type councilConveneResp struct {
	ID string `json:"id"`
}

func (d *Deps) councilsConveneHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	var req councilConveneReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "decode body: "+err.Error())
		return
	}
	if req.IntentID == "" || req.BindingQuestion == "" || req.Bishop == "" {
		writeErr(w, http.StatusBadRequest, "intent_id, binding_question, bishop required")
		return
	}
	if len(req.Members) == 0 {
		writeErr(w, http.StatusBadRequest, "members array required")
		return
	}
	if req.ConvenedBy == "" {
		req.ConvenedBy = "human"
	}

	var newID string
	err := d.Pool.QueryRow(ctx,
		`SELECT stewards.convene_council($1::uuid, $2, $3::jsonb, $4, $5)::text`,
		req.IntentID, req.BindingQuestion, string(req.Members), req.Bishop, req.ConvenedBy,
	).Scan(&newID)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, councilConveneResp{ID: newID})
}

type councilResolveReq struct {
	CouncilID        string `json:"council_id"`
	Action           string `json:"action"`             // accept | request_revision | dissolve
	ResolutionText   string `json:"resolution_text,omitempty"`
	Destination      string `json:"destination,omitempty"`  // study | decisions | NULL
	ResolvedBy       string `json:"resolved_by,omitempty"`
	DissolvedReason  string `json:"dissolved_reason,omitempty"`
}

type councilResolveResp struct {
	ResolutionID string `json:"resolution_id"`
}

func (d *Deps) councilsResolveHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var req councilResolveReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "decode body: "+err.Error())
		return
	}
	if req.CouncilID == "" || req.Action == "" {
		writeErr(w, http.StatusBadRequest, "council_id and action required")
		return
	}

	var destArg any = nil
	if req.Destination != "" {
		destArg = req.Destination
	}
	var dissolvedArg any = nil
	if req.DissolvedReason != "" {
		dissolvedArg = req.DissolvedReason
	}

	var resID *string
	err := d.Pool.QueryRow(ctx,
		`SELECT stewards.resolve_council($1::uuid, $2, $3, $4, $5, $6)::text`,
		req.CouncilID, req.Action, req.ResolutionText, destArg, req.ResolvedBy, dissolvedArg,
	).Scan(&resID)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	out := councilResolveResp{}
	if resID != nil {
		out.ResolutionID = *resID
	}
	writeJSON(w, http.StatusOK, out)
}

type councilSuggestion struct {
	PipelineFamily string `json:"pipeline_family"`
	CurrentStage   string `json:"current_stage"`
	LessonCount    int64  `json:"lesson_count"`
	SampleContent  string `json:"sample_content"`
}

type councilSuggestResp struct {
	Items []councilSuggestion `json:"items"`
	Total int                 `json:"total"`
}

func (d *Deps) councilsSuggestHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	minLessons := atoiDefault(r.URL.Query().Get("min_lessons"), 5, 1, 100)

	rows, err := d.Pool.Query(ctx,
		`SELECT pipeline_family, current_stage, lesson_count,
		        coalesce(sample_content, '')
		   FROM stewards.suggest_councils($1)`, minLessons)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	resp := councilSuggestResp{Items: []councilSuggestion{}}
	for rows.Next() {
		var s councilSuggestion
		if err := rows.Scan(&s.PipelineFamily, &s.CurrentStage,
			&s.LessonCount, &s.SampleContent); err == nil {
			resp.Items = append(resp.Items, s)
		}
	}
	resp.Total = len(resp.Items)
	writeJSON(w, http.StatusOK, resp)
}
