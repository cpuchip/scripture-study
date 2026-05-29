// Brainstorm endpoints — list lenses + dispatch a brainstorm work_item.
// Backs the Stewards-UI /brainstorm route. Wraps the J.8 + J.9 SQL work
// landed 2026-05-29 (4-layer dispatch fallback chain, p_models per-lens
// override, p_lenses subset selection, 12 lens library).

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func (d *Deps) registerBrainstorm(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/brainstorm/lenses", d.brainstormLensesHandler)
	mux.HandleFunc("POST /api/brainstorm/start", d.brainstormStartHandler)
}

// ---------------------------------------------------------------------
// GET /api/brainstorm/lenses
// ---------------------------------------------------------------------

type brainstormLensRow struct {
	ShortName         string `json:"short_name"`
	PipelineFamily    string `json:"pipeline_family"`
	Description       string `json:"description"`
	DefaultModel      string `json:"default_model,omitempty"`
	SuggestedModel    string `json:"suggested_model,omitempty"`
	DefaultProvider   string `json:"default_provider,omitempty"`
	SuggestedProvider string `json:"suggested_provider,omitempty"`
	IsOriginal        bool   `json:"is_original"`
}

type brainstormLensesResp struct {
	Items []brainstormLensRow `json:"items"`
	Total int                 `json:"total"`
}

// Original 4 lenses (per 2026-05-13 J.4 ratification); used to mark the
// pre-checked subset in the UI.
var originalLensSet = map[string]bool{
	"scamper":  true,
	"six-hats": true,
	"crazy8s":  true,
	"reverse":  true,
}

func (d *Deps) brainstormLensesHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rows, err := d.Pool.Query(ctx, `
		SELECT regexp_replace(family, '^brainstorm-', ''),
		       family,
		       description,
		       COALESCE(metadata->>'default_model',      ''),
		       COALESCE(metadata->>'suggested_model',    ''),
		       COALESCE(metadata->>'default_provider',   ''),
		       COALESCE(metadata->>'suggested_provider', '')
		  FROM stewards.pipelines
		 WHERE family LIKE 'brainstorm-%'
		 ORDER BY family`)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	resp := brainstormLensesResp{Items: []brainstormLensRow{}}
	for rows.Next() {
		var row brainstormLensRow
		if err := rows.Scan(&row.ShortName, &row.PipelineFamily, &row.Description,
			&row.DefaultModel, &row.SuggestedModel,
			&row.DefaultProvider, &row.SuggestedProvider); err != nil {
			continue
		}
		row.IsOriginal = originalLensSet[row.ShortName]
		resp.Items = append(resp.Items, row)
	}
	resp.Total = len(resp.Items)
	writeJSON(w, http.StatusOK, resp)
}

// ---------------------------------------------------------------------
// POST /api/brainstorm/start
// ---------------------------------------------------------------------

type brainstormStartReq struct {
	BindingQuestion       string            `json:"binding_question"`
	Destination           string            `json:"destination,omitempty"`
	Slug                  string            `json:"slug,omitempty"`
	Lenses                []string          `json:"lenses,omitempty"`
	Models                map[string]string `json:"models,omitempty"`
	ProjectAssociation    string            `json:"project_association,omitempty"`
	Actor                 string            `json:"actor,omitempty"`
	CostCapPerLensMicro   int64             `json:"cost_cap_per_lens_micro,omitempty"`
}

type brainstormChildRow struct {
	ID             string `json:"id"`
	Slug           string `json:"slug"`
	PipelineFamily string `json:"pipeline_family"`
	ModelOverride  string `json:"model_override,omitempty"`
}

type brainstormStartResp struct {
	ParentID     string               `json:"parent_id"`
	Slug         string               `json:"slug"`
	Destination  string               `json:"destination"`
	Lenses       []string             `json:"lenses"`
	Children     []brainstormChildRow `json:"children"`
	AggregatorID string               `json:"aggregator_id,omitempty"`
}

func (d *Deps) brainstormStartHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var req brainstormStartReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}

	if strings.TrimSpace(req.BindingQuestion) == "" {
		writeErr(w, http.StatusBadRequest, "binding_question is required")
		return
	}

	if req.Actor == "" {
		req.Actor = "michael"
	}

	// Default slug — mirror the SQL function's pattern but resolve here
	// so the response carries the same slug used for destination defaulting.
	slug := req.Slug
	if slug == "" {
		slug = "brainstorm-" + time.Now().UTC().Format("20060102-150405")
	}

	destination := req.Destination
	if destination == "" {
		destination = "study/.scratch/" + slug + ".md"
	}

	// Build the SQL call. NULL-pass any field the caller left unset so
	// the SQL function's DEFAULTs take effect (single source of truth).
	var lensesArg any
	if len(req.Lenses) > 0 {
		lensesArg = req.Lenses
	}

	var modelsArg any
	if len(req.Models) > 0 {
		modelsJSON, err := json.Marshal(req.Models)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "models marshal: "+err.Error())
			return
		}
		modelsArg = string(modelsJSON)
	}

	var costCapArg any
	if req.CostCapPerLensMicro > 0 {
		costCapArg = req.CostCapPerLensMicro
	}

	var projectArg any
	if req.ProjectAssociation != "" {
		projectArg = req.ProjectAssociation
	}

	var parentID string
	err := d.Pool.QueryRow(ctx, `
		SELECT stewards.start_brainstorm(
			p_binding_question        := $1,
			p_destination             := $2,
			p_project_association     := $3,
			p_actor                   := $4,
			p_slug                    := $5,
			p_cost_cap_per_lens_micro := $6,
			p_models                  := $7::jsonb,
			p_lenses                  := $8::text[]
		)::text`,
		req.BindingQuestion,
		destination,
		projectArg,
		req.Actor,
		slug,
		costCapArg,
		modelsArg,
		lensesArg,
	).Scan(&parentID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "start_brainstorm: "+err.Error())
		return
	}

	// Read back children + aggregator for the response.
	resp := brainstormStartResp{
		ParentID:    parentID,
		Slug:        slug,
		Destination: destination,
		Lenses:      []string{},
		Children:    []brainstormChildRow{},
	}

	childRows, err := d.Pool.Query(ctx, `
		SELECT id::text, slug, pipeline_family, COALESCE(model_override, '')
		  FROM stewards.work_items
		 WHERE parent_work_item_id = $1::uuid
		   AND pipeline_family LIKE 'brainstorm-%'
		 ORDER BY slug`, parentID)
	if err == nil {
		defer childRows.Close()
		for childRows.Next() {
			var c brainstormChildRow
			if err := childRows.Scan(&c.ID, &c.Slug, &c.PipelineFamily, &c.ModelOverride); err == nil {
				resp.Children = append(resp.Children, c)
				resp.Lenses = append(resp.Lenses, strings.TrimPrefix(c.PipelineFamily, "brainstorm-"))
			}
		}
	}

	var aggID string
	_ = d.Pool.QueryRow(ctx, `
		SELECT id::text
		  FROM stewards.work_items
		 WHERE parent_work_item_id = $1::uuid
		   AND pipeline_family = 'aggregate-children'
		 LIMIT 1`, parentID).Scan(&aggID)
	resp.AggregatorID = aggID

	writeJSON(w, http.StatusCreated, resp)
}
