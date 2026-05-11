// lessons + sabbath endpoints — Phase 5e (D.6).
// Backs Stewards-UI /lessons (review + ratify) and /sabbath
// (reflection log) routes.

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

func (d *Deps) registerLessons(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/lessons/list", d.lessonsListHandler)
	mux.HandleFunc("POST /api/lessons/ratify", d.lessonsRatifyHandler)
	mux.HandleFunc("GET /api/sabbath/list", d.sabbathListHandler)
}

type lessonRow struct {
	ID            int64           `json:"id"`
	WorkItemID    string          `json:"work_item_id,omitempty"`
	WorkItemSlug  string          `json:"work_item_slug,omitempty"`
	At            *time.Time      `json:"at,omitempty"`
	Kind          string          `json:"kind"`
	Content       string          `json:"content"`
	RawResponse   json.RawMessage `json:"raw_response,omitempty"`
	RatifiedAt    *time.Time      `json:"ratified_at,omitempty"`
	RatifiedBy    string          `json:"ratified_by,omitempty"`
	PromotedTo    string          `json:"promoted_to,omitempty"`
	WorkID        *int64          `json:"work_id,omitempty"`
	PipelineFamily string         `json:"pipeline_family,omitempty"`
	CurrentStage   string         `json:"current_stage,omitempty"`
}

type lessonsListResp struct {
	Items []lessonRow `json:"items"`
	Total int         `json:"total"`
}

func (d *Deps) lessonsListHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	q := r.URL.Query()
	kind := q.Get("kind")           // optional: principle | decision | lesson | sabbath_reflection
	ratified := q.Get("ratified")   // optional: 'true' | 'false' | '' (any)
	limit := atoiDefault(q.Get("limit"), 100, 1, 500)

	whereClauses := []string{}
	args := []any{}
	if kind != "" {
		args = append(args, kind)
		whereClauses = append(whereClauses, "l.kind = $"+itoa(len(args)))
	}
	if ratified == "true" {
		whereClauses = append(whereClauses, "l.ratified_at IS NOT NULL")
	} else if ratified == "false" {
		whereClauses = append(whereClauses, "l.ratified_at IS NULL")
	}

	where := ""
	if len(whereClauses) > 0 {
		where = " WHERE " + joinAnd(whereClauses)
	}
	args = append(args, limit)

	rows, err := d.Pool.Query(ctx, `
		SELECT l.id, l.work_item_id::text, coalesce(wi.slug, ''), l.at, l.kind,
		       l.content, coalesce(l.raw_response, '{}'::jsonb),
		       l.ratified_at, coalesce(l.ratified_by, ''),
		       coalesce(l.promoted_to, ''), l.work_id,
		       coalesce(wi.pipeline_family, ''), coalesce(wi.current_stage, '')
		  FROM stewards.lessons l
		  LEFT JOIN stewards.work_items wi ON wi.id = l.work_item_id`+where+`
		  ORDER BY l.at DESC
		  LIMIT $`+itoa(len(args)),
		args...)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	resp := lessonsListResp{Items: []lessonRow{}}
	for rows.Next() {
		var l lessonRow
		if err := rows.Scan(&l.ID, &l.WorkItemID, &l.WorkItemSlug, &l.At, &l.Kind,
			&l.Content, &l.RawResponse, &l.RatifiedAt, &l.RatifiedBy,
			&l.PromotedTo, &l.WorkID, &l.PipelineFamily, &l.CurrentStage); err == nil {
			resp.Items = append(resp.Items, l)
		}
	}
	resp.Total = len(resp.Items)
	writeJSON(w, http.StatusOK, resp)
}

type lessonRatifyReq struct {
	ID         int64  `json:"id"`
	RatifiedBy string `json:"ratified_by"`
	PromotedTo string `json:"promoted_to,omitempty"`  // optional: ".mind/principles.md" | ".mind/decisions.md"
}

type lessonRatifyResp struct {
	ID         int64  `json:"id"`
	RatifiedAt string `json:"ratified_at"`
}

func (d *Deps) lessonsRatifyHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req lessonRatifyReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "decode body: "+err.Error())
		return
	}
	if req.ID == 0 || req.RatifiedBy == "" {
		writeErr(w, http.StatusBadRequest, "id and ratified_by required")
		return
	}

	var ratifiedAt time.Time
	err := d.Pool.QueryRow(ctx, `
		UPDATE stewards.lessons
		   SET ratified_at = now(),
		       ratified_by = $1,
		       promoted_to = NULLIF($2, '')
		 WHERE id = $3
		 RETURNING ratified_at`,
		req.RatifiedBy, req.PromotedTo, req.ID,
	).Scan(&ratifiedAt)
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, lessonRatifyResp{ID: req.ID, RatifiedAt: ratifiedAt.Format(time.RFC3339)})
}

// /api/sabbath/list — chronological list of recent sabbath_reflection rows.
type sabbathRow struct {
	ID            int64      `json:"id"`
	WorkItemID    string     `json:"work_item_id"`
	WorkItemSlug  string     `json:"work_item_slug"`
	PipelineFamily string    `json:"pipeline_family"`
	At            *time.Time `json:"at,omitempty"`
	Reflection    string     `json:"reflection"`
	CarryForward  string     `json:"carry_forward,omitempty"`
	Surprise      string     `json:"surprise,omitempty"`
}

type sabbathListResp struct {
	Items []sabbathRow `json:"items"`
	Total int          `json:"total"`
}

func (d *Deps) sabbathListHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	q := r.URL.Query()
	pipeline := q.Get("pipeline")
	limit := atoiDefault(q.Get("limit"), 50, 1, 200)

	whereClauses := []string{"l.kind = 'sabbath_reflection'"}
	args := []any{}
	if pipeline != "" {
		args = append(args, pipeline)
		whereClauses = append(whereClauses, "wi.pipeline_family = $"+itoa(len(args)))
	}
	where := " WHERE " + joinAnd(whereClauses)
	args = append(args, limit)

	rows, err := d.Pool.Query(ctx, `
		SELECT l.id, l.work_item_id::text, coalesce(wi.slug, ''),
		       coalesce(wi.pipeline_family, ''), l.at,
		       coalesce(l.raw_response->>'reflection', ''),
		       coalesce(l.raw_response->>'carry_forward', ''),
		       coalesce(l.raw_response->>'surprise', '')
		  FROM stewards.lessons l
		  LEFT JOIN stewards.work_items wi ON wi.id = l.work_item_id`+where+`
		  ORDER BY l.at DESC
		  LIMIT $`+itoa(len(args)),
		args...)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	resp := sabbathListResp{Items: []sabbathRow{}}
	for rows.Next() {
		var s sabbathRow
		if err := rows.Scan(&s.ID, &s.WorkItemID, &s.WorkItemSlug,
			&s.PipelineFamily, &s.At, &s.Reflection, &s.CarryForward, &s.Surprise); err == nil {
			resp.Items = append(resp.Items, s)
		}
	}
	resp.Total = len(resp.Items)
	writeJSON(w, http.StatusOK, resp)
}

// Local atoiDefault helper (exists in helpers.go but re-declared here
// would conflict; using the existing one from helpers.go).
