// Dashboard endpoint — single-call snapshot of substrate health for
// the home view. Aggregates four pieces:
//   - pg health (db ping)
//   - soak status (schedule_enabled, last pass, dirty_queue depth)
//   - in-flight work_items (status IN pending/in_progress/dispatched)
//   - recent errors (work_queue.status='error' in the last 24h)
//
// All four queried in parallel via goroutines so dashboard load stays
// snappy even under heavy substrate load.

package api

import (
	"context"
	"net/http"
	"sync"
	"time"
)

type dashboardResponse struct {
	PG          pgHealth        `json:"pg"`
	Soak        soakStatus      `json:"soak"`
	InFlight    []workItemBrief `json:"in_flight"`
	RecentError []errorBrief    `json:"recent_errors"`
	FetchedAtMs int64           `json:"fetched_at_ms"`
}

type pgHealth struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

type soakStatus struct {
	ScheduleEnabled    bool       `json:"schedule_enabled"`
	LastPassID         string     `json:"last_pass_id,omitempty"`
	LastPassStatus     string     `json:"last_pass_status,omitempty"`
	LastPassStartedAt  *time.Time `json:"last_pass_started_at,omitempty"`
	LastPassFinishedAt *time.Time `json:"last_pass_finished_at,omitempty"`
	DirtyQueueDepth    int        `json:"dirty_queue_depth"`
}

type workItemBrief struct {
	ID            string     `json:"id"`
	Slug          string     `json:"slug"`
	Pipeline      string     `json:"pipeline"`
	CurrentStage  string     `json:"current_stage"`
	Status        string     `json:"status"`
	TokensIn      int        `json:"tokens_in"`
	TokensOut     int        `json:"tokens_out"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
}

type errorBrief struct {
	ID       int64      `json:"id"`
	Kind     string     `json:"kind"`
	Provider string     `json:"provider"`
	Error    string     `json:"error"`
	DoneAt   *time.Time `json:"done_at,omitempty"`
}

func (d *Deps) dashboardHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp := dashboardResponse{
		FetchedAtMs: time.Now().UnixMilli(),
		InFlight:    []workItemBrief{},
		RecentError: []errorBrief{},
	}

	var wg sync.WaitGroup
	wg.Add(4)

	// pg ping
	go func() {
		defer wg.Done()
		if err := d.Pool.Ping(ctx); err != nil {
			resp.PG = pgHealth{OK: false, Error: err.Error()}
		} else {
			resp.PG = pgHealth{OK: true}
		}
	}()

	// soak status
	go func() {
		defer wg.Done()
		var s soakStatus
		err := d.Pool.QueryRow(ctx,
			`SELECT schedule_enabled FROM stewards.watchman_config WHERE id=1`,
		).Scan(&s.ScheduleEnabled)
		if err != nil {
			resp.Soak = s
			return
		}
		// last pass
		_ = d.Pool.QueryRow(ctx,
			`SELECT pass_id, status, started_at, finished_at
			   FROM stewards.watchman_passes
			   ORDER BY started_at DESC LIMIT 1`,
		).Scan(&s.LastPassID, &s.LastPassStatus, &s.LastPassStartedAt, &s.LastPassFinishedAt)
		// dirty queue depth
		_ = d.Pool.QueryRow(ctx,
			`SELECT count(*) FROM stewards.dirty_queue`,
		).Scan(&s.DirtyQueueDepth)
		resp.Soak = s
	}()

	// in-flight work_items
	go func() {
		defer wg.Done()
		rows, err := d.Pool.Query(ctx,
			`SELECT id::text, slug, pipeline_family, current_stage, status,
			        coalesce(tokens_in, 0), coalesce(tokens_out, 0), updated_at
			   FROM stewards.work_items
			  WHERE status IN ('pending', 'in_progress', 'dispatched')
			  ORDER BY updated_at DESC NULLS LAST
			  LIMIT 20`,
		)
		if err != nil {
			return
		}
		defer rows.Close()
		for rows.Next() {
			var w workItemBrief
			if err := rows.Scan(&w.ID, &w.Slug, &w.Pipeline, &w.CurrentStage,
				&w.Status, &w.TokensIn, &w.TokensOut, &w.UpdatedAt); err == nil {
				resp.InFlight = append(resp.InFlight, w)
			}
		}
	}()

	// recent errors (last 24h)
	go func() {
		defer wg.Done()
		rows, err := d.Pool.Query(ctx,
			`SELECT id, kind, provider, coalesce(error, ''), done_at
			   FROM stewards.work_queue
			  WHERE status='error' AND done_at > now() - interval '24 hours'
			  ORDER BY done_at DESC
			  LIMIT 10`,
		)
		if err != nil {
			return
		}
		defer rows.Close()
		for rows.Next() {
			var e errorBrief
			if err := rows.Scan(&e.ID, &e.Kind, &e.Provider, &e.Error, &e.DoneAt); err == nil {
				resp.RecentError = append(resp.RecentError, e)
			}
		}
	}()

	wg.Wait()
	writeJSON(w, http.StatusOK, resp)
}
