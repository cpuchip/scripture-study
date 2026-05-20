// scheduled endpoints — list, get, create, update, toggle, delete, and
// recent-runs for stewards.scheduled_pipelines. Backs Stewards-UI
// /scheduled route + Dashboard "Last 7 scheduled runs" card. Phase PE-C.1.

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

func (d *Deps) registerScheduled(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/scheduled/list", d.scheduledListHandler)
	mux.HandleFunc("GET /api/scheduled/get", d.scheduledGetHandler)
	mux.HandleFunc("POST /api/scheduled/create", d.scheduledCreateHandler)
	mux.HandleFunc("PUT /api/scheduled/update", d.scheduledUpdateHandler)
	mux.HandleFunc("POST /api/scheduled/toggle", d.scheduledToggleHandler)
	mux.HandleFunc("DELETE /api/scheduled/delete", d.scheduledDeleteHandler)
	mux.HandleFunc("GET /api/scheduled/recent-runs", d.scheduledRecentRunsHandler)
}

type scheduledRow struct {
	ID                 string          `json:"id"`
	Slug               string          `json:"slug"`
	PipelineFamily     string          `json:"pipeline_family"`
	IntentID           string          `json:"intent_id"`
	IntentSlug         string          `json:"intent_slug,omitempty"`
	CronPattern        string          `json:"cron_pattern"`
	InputTemplate      json.RawMessage `json:"input_template"`
	Enabled            bool            `json:"enabled"`
	MissedWindowHours  int             `json:"missed_window_hours"`
	LastDispatchedAt   *time.Time      `json:"last_dispatched_at,omitempty"`
	NextDueAt          *time.Time      `json:"next_due_at,omitempty"`
	CreatedAt          *time.Time      `json:"created_at,omitempty"`
	UpdatedAt          *time.Time      `json:"updated_at,omitempty"`
	Notes              string          `json:"notes,omitempty"`
}

type scheduledListResp struct {
	Items []scheduledRow `json:"items"`
	Total int            `json:"total"`
}

func scanScheduledRow(scanner interface {
	Scan(...any) error
}) (scheduledRow, error) {
	var sr scheduledRow
	err := scanner.Scan(
		&sr.ID, &sr.Slug, &sr.PipelineFamily, &sr.IntentID, &sr.IntentSlug,
		&sr.CronPattern, &sr.InputTemplate, &sr.Enabled, &sr.MissedWindowHours,
		&sr.LastDispatchedAt, &sr.NextDueAt, &sr.CreatedAt, &sr.UpdatedAt, &sr.Notes,
	)
	return sr, err
}

const scheduledSelectSQL = `
	SELECT sp.id::text, sp.slug, sp.pipeline_family,
	       sp.intent_id::text, coalesce(i.slug, ''),
	       sp.cron_pattern, sp.input_template, sp.enabled,
	       sp.missed_window_hours, sp.last_dispatched_at, sp.next_due_at,
	       sp.created_at, sp.updated_at, coalesce(sp.notes, '')
	  FROM stewards.scheduled_pipelines sp
	  LEFT JOIN stewards.intents i ON i.id = sp.intent_id
`

func (d *Deps) scheduledListHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp := scheduledListResp{Items: []scheduledRow{}}
	rows, err := d.Pool.Query(ctx, scheduledSelectSQL+" ORDER BY sp.slug")
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		sr, err := scanScheduledRow(rows)
		if err == nil {
			resp.Items = append(resp.Items, sr)
		}
	}
	resp.Total = len(resp.Items)
	writeJSON(w, http.StatusOK, resp)
}

func (d *Deps) scheduledGetHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id := r.URL.Query().Get("id")
	slug := r.URL.Query().Get("slug")
	if id == "" && slug == "" {
		writeErr(w, http.StatusBadRequest, "id or slug query param required")
		return
	}
	where := "WHERE sp.id::text = $1"
	arg := any(id)
	if id == "" {
		where = "WHERE sp.slug = $1"
		arg = slug
	}

	sr, err := scanScheduledRow(d.Pool.QueryRow(ctx, scheduledSelectSQL+" "+where, arg))
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sr)
}

type scheduledCreateReq struct {
	Slug              string          `json:"slug"`
	PipelineFamily    string          `json:"pipeline_family"`
	IntentSlug        string          `json:"intent_slug"`
	CronPattern       string          `json:"cron_pattern"`
	InputTemplate     json.RawMessage `json:"input_template"`
	Enabled           *bool           `json:"enabled,omitempty"`
	MissedWindowHours *int            `json:"missed_window_hours,omitempty"`
	Notes             string          `json:"notes,omitempty"`
}

type scheduledCreateResp struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
}

func (d *Deps) scheduledCreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req scheduledCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "decode body: "+err.Error())
		return
	}
	if req.Slug == "" || req.PipelineFamily == "" || req.IntentSlug == "" || req.CronPattern == "" {
		writeErr(w, http.StatusBadRequest, "slug, pipeline_family, intent_slug, cron_pattern are required")
		return
	}
	if len(req.InputTemplate) == 0 {
		req.InputTemplate = json.RawMessage(`{}`)
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	missed := 24
	if req.MissedWindowHours != nil {
		missed = *req.MissedWindowHours
	}

	var newID string
	err := d.Pool.QueryRow(ctx, `
		INSERT INTO stewards.scheduled_pipelines
		    (slug, pipeline_family, intent_id, cron_pattern, input_template,
		     enabled, missed_window_hours, notes)
		VALUES (
		    $1, $2,
		    (SELECT id FROM stewards.intents WHERE slug = $3),
		    $4, $5, $6, $7, NULLIF($8, '')
		)
		RETURNING id::text
	`, req.Slug, req.PipelineFamily, req.IntentSlug, req.CronPattern,
		req.InputTemplate, enabled, missed, req.Notes).Scan(&newID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "create: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, scheduledCreateResp{ID: newID, Slug: req.Slug})
}

type scheduledUpdateReq struct {
	CronPattern       *string         `json:"cron_pattern,omitempty"`
	InputTemplate     json.RawMessage `json:"input_template,omitempty"`
	Enabled           *bool           `json:"enabled,omitempty"`
	MissedWindowHours *int            `json:"missed_window_hours,omitempty"`
	Notes             *string         `json:"notes,omitempty"`
}

func (d *Deps) scheduledUpdateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id := r.URL.Query().Get("id")
	if id == "" {
		writeErr(w, http.StatusBadRequest, "id query param required")
		return
	}

	var req scheduledUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "decode body: "+err.Error())
		return
	}

	// COALESCE-style update: pass NULL for fields the client didn't send.
	// next_due_at is recomputed by the trigger when cron_pattern changes.
	_, err := d.Pool.Exec(ctx, `
		UPDATE stewards.scheduled_pipelines SET
		    cron_pattern        = COALESCE($2, cron_pattern),
		    input_template      = COALESCE($3, input_template),
		    enabled             = COALESCE($4, enabled),
		    missed_window_hours = COALESCE($5, missed_window_hours),
		    notes               = COALESCE($6, notes)
		 WHERE id::text = $1
	`, id, req.CronPattern, req.InputTemplate, req.Enabled, req.MissedWindowHours, req.Notes)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "update: "+err.Error())
		return
	}

	// Return the refreshed row
	sr, err := scanScheduledRow(d.Pool.QueryRow(ctx, scheduledSelectSQL+" WHERE sp.id::text = $1", id))
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sr)
}

func (d *Deps) scheduledToggleHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id := r.URL.Query().Get("id")
	if id == "" {
		writeErr(w, http.StatusBadRequest, "id query param required")
		return
	}

	var enabled bool
	err := d.Pool.QueryRow(ctx, `
		UPDATE stewards.scheduled_pipelines
		   SET enabled = NOT enabled
		 WHERE id::text = $1
		 RETURNING enabled
	`, id).Scan(&enabled)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "toggle: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"enabled": enabled})
}

func (d *Deps) scheduledDeleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id := r.URL.Query().Get("id")
	if id == "" {
		writeErr(w, http.StatusBadRequest, "id query param required")
		return
	}
	_, err := d.Pool.Exec(ctx, `DELETE FROM stewards.scheduled_pipelines WHERE id::text = $1`, id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "delete: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"deleted": id})
}

// scheduledRecentRunsHandler returns the most recent N work_items
// spawned by scheduled_pipelines (identified by actor='scheduler').
// Backs the Dashboard "Last 7 scheduled runs" card.
type scheduledRunRow struct {
	WorkItemID     string     `json:"work_item_id"`
	Slug           string     `json:"slug"`
	ScheduleSlug   string     `json:"schedule_slug,omitempty"`
	PipelineFamily string     `json:"pipeline_family"`
	Status         string     `json:"status"`
	CurrentStage   string     `json:"current_stage,omitempty"`
	CreatedAt      *time.Time `json:"created_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	FilePath       string     `json:"file_path,omitempty"`
}

type scheduledRunsResp struct {
	Items []scheduledRunRow `json:"items"`
	Total int               `json:"total"`
}

func (d *Deps) scheduledRecentRunsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	limit := 7
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	rows, err := d.Pool.Query(ctx, `
		SELECT wi.id::text,
		       coalesce(wi.slug, ''),
		       coalesce(split_part(wi.slug, '--', 1), '') AS schedule_slug,
		       wi.pipeline_family,
		       wi.status,
		       coalesce(wi.current_stage, ''),
		       wi.created_at,
		       wi.completed_at,
		       coalesce(wi.file_destination, '')
		  FROM stewards.work_items wi
		 WHERE wi.actor = 'scheduler'
		 ORDER BY wi.created_at DESC
		 LIMIT $1
	`, limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	resp := scheduledRunsResp{Items: []scheduledRunRow{}}
	for rows.Next() {
		var rr scheduledRunRow
		if err := rows.Scan(&rr.WorkItemID, &rr.Slug, &rr.ScheduleSlug,
			&rr.PipelineFamily, &rr.Status, &rr.CurrentStage,
			&rr.CreatedAt, &rr.CompletedAt, &rr.FilePath); err == nil {
			resp.Items = append(resp.Items, rr)
		}
	}
	resp.Total = len(resp.Items)
	writeJSON(w, http.StatusOK, resp)
}
