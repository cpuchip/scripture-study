// new_work endpoint — POST /api/work-items/create. Wraps the substrate's
// stewards.work_item_create() + work_item_dispatch_stage() so the UI's
// New Work form can submit a binding question and immediately kick off
// the first stage.

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

func (d *Deps) registerNewWork(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/work-items/create", d.workItemCreateHandler)
}

type workItemCreateReq struct {
	Pipeline            string          `json:"pipeline"`
	Slug                string          `json:"slug,omitempty"`
	Input               json.RawMessage `json:"input,omitempty"`
	UserInput           string          `json:"user_input,omitempty"`
	Actor               string          `json:"actor,omitempty"`
	TokenBudget         *int            `json:"token_budget,omitempty"`
	Dispatch            bool            `json:"dispatch,omitempty"`
	DestinationMaturity string          `json:"destination_maturity,omitempty"`
	IntentID            string          `json:"intent_id,omitempty"`
	FileDestination     string          `json:"file_destination,omitempty"`
}

type workItemCreateResp struct {
	ID           string `json:"id"`
	WorkQueueID  *int64 `json:"work_queue_id,omitempty"`
	Dispatched   bool   `json:"dispatched"`
}

func (d *Deps) workItemCreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var req workItemCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "decode body: "+err.Error())
		return
	}
	if req.Pipeline == "" {
		writeErr(w, http.StatusBadRequest, "pipeline is required")
		return
	}
	if req.Actor == "" {
		req.Actor = "human"
	}
	if len(req.Input) == 0 {
		req.Input = json.RawMessage(`{}`)
	}

	var slugArg any = nil
	if req.Slug != "" {
		slugArg = req.Slug
	}
	var budgetArg any = nil
	if req.TokenBudget != nil {
		budgetArg = *req.TokenBudget
	}

	var intentArg any = nil
	if req.IntentID != "" {
		intentArg = req.IntentID
	}
	var newID string
	err := d.Pool.QueryRow(ctx,
		`SELECT stewards.work_item_create($1, $2::jsonb, $3, $4, $5, $6::uuid)::text`,
		req.Pipeline, string(req.Input), slugArg, req.Actor, budgetArg, intentArg,
	).Scan(&newID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "create: "+err.Error())
		return
	}

	// Phase 5a: optional destination_maturity sets the human's ceiling.
	// NULL = default = full Ammon-loop to verified; set lower (e.g.
	// specced) to surface for review before continuing.
	if req.DestinationMaturity != "" {
		_, err := d.Pool.Exec(ctx,
			`UPDATE stewards.work_items SET destination_maturity = $1 WHERE id = $2::uuid`,
			req.DestinationMaturity, newID,
		)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "set destination_maturity: "+err.Error())
			return
		}
	}

	// Batch G.4: optional file_destination prefilled from pipeline
	// template or human-edited. NULL = DB-only (default).
	if req.FileDestination != "" {
		_, err := d.Pool.Exec(ctx,
			`UPDATE stewards.work_items SET file_destination = $1 WHERE id = $2::uuid`,
			req.FileDestination, newID,
		)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "set file_destination: "+err.Error())
			return
		}
	}

	resp := workItemCreateResp{ID: newID}

	if req.Dispatch {
		var wqID int64
		var userInputArg any = nil
		if req.UserInput != "" {
			userInputArg = req.UserInput
		}
		err := d.Pool.QueryRow(ctx,
			`SELECT stewards.work_item_dispatch_stage($1::uuid, $2)`,
			newID, userInputArg,
		).Scan(&wqID)
		if err != nil {
			// Created OK but dispatch failed — surface partial success
			writeJSON(w, http.StatusOK, map[string]any{
				"id":         newID,
				"dispatched": false,
				"error":      "create OK but dispatch failed: " + err.Error(),
			})
			return
		}
		resp.WorkQueueID = &wqID
		resp.Dispatched = true
	}

	writeJSON(w, http.StatusOK, resp)
}
