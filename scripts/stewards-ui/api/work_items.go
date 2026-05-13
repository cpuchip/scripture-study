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
	mux.HandleFunc("GET /api/work-items/cost", d.workItemsCostHandler)
	mux.HandleFunc("GET /api/work-items/actions", d.workItemsActionsHandler)
	mux.HandleFunc("GET /api/work-items/gate-decisions", d.workItemsGateDecisionsHandler)
	mux.HandleFunc("POST /api/work-items/set-file-destination", d.workItemsSetFileDestinationHandler)
	mux.HandleFunc("POST /api/work-items/materialize-file", d.workItemsMaterializeFileHandler)
	// H.3-followup: agent_planning proposal actions
	mux.HandleFunc("POST /api/work-items/ratify", d.workItemsRatifyHandler)
	mux.HandleFunc("POST /api/work-items/dispatch", d.workItemsDispatchHandler)
	mux.HandleFunc("POST /api/work-items/cancel-proposal", d.workItemsCancelProposalHandler)
	// Edit + AI-revise (third proposal mode)
	mux.HandleFunc("POST /api/work-items/edit-proposal", d.workItemsEditProposalHandler)
	mux.HandleFunc("POST /api/work-items/revise-with-feedback", d.workItemsReviseWithFeedbackHandler)
	mux.HandleFunc("GET /api/work-items/pending-revisions", d.workItemsPendingRevisionsHandler)
	mux.HandleFunc("POST /api/work-items/apply-revision", d.workItemsApplyRevisionHandler)
	mux.HandleFunc("POST /api/work-items/reject-revision", d.workItemsRejectRevisionHandler)
}

type workItemRow struct {
	ID                  string     `json:"id"`
	Slug                string     `json:"slug"`
	Pipeline            string     `json:"pipeline"`
	CurrentStage        string     `json:"current_stage"`
	Status              string     `json:"status"`
	Actor               string     `json:"actor,omitempty"`
	TokensIn            int        `json:"tokens_in"`
	TokensOut           int        `json:"tokens_out"`
	TokenBudget         *int       `json:"token_budget,omitempty"`
	CreatedAt           *time.Time `json:"created_at,omitempty"`
	UpdatedAt           *time.Time `json:"updated_at,omitempty"`
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
	// H.3 — origin + project + parent linkage
	Origin              string     `json:"origin,omitempty"`
	ProjectAssociation  string     `json:"project_association,omitempty"`
	ParentWorkItemID    string     `json:"parent_work_item_id,omitempty"`
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
	origin := q.Get("origin")
	project := q.Get("project_association")
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
	if origin != "" {
		args = append(args, origin)
		whereClauses = append(whereClauses, "origin = $"+itoa(len(args)))
	}
	if project != "" {
		args = append(args, project)
		whereClauses = append(whereClauses, "project_association = $"+itoa(len(args)))
	}
	where := ""
	if len(whereClauses) > 0 {
		where = " WHERE " + joinAnd(whereClauses)
	}

	resp := workItemsListResp{Items: []workItemRow{}}
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
		        created_at, updated_at, completed_at,
		        coalesce(origin, 'human'),
		        coalesce(project_association, ''),
		        coalesce(parent_work_item_id::text, '')
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
			&w.CreatedAt, &w.UpdatedAt, &w.CompletedAt,
			&w.Origin, &w.ProjectAssociation, &w.ParentWorkItemID); err == nil {
			resp.Items = append(resp.Items, w)
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

type workItemDetail struct {
	workItemRow
	Input                json.RawMessage `json:"input"`
	StageResults         json.RawMessage `json:"stage_results"`
	SessionIDs           []string        `json:"session_ids,omitempty"`
	Error                string          `json:"error,omitempty"`
	// Phase 4j — steward + cost surface on detail view
	FailureCount         int             `json:"failure_count"`
	LastFailureReason    string          `json:"last_failure_reason,omitempty"`
	LastFailureDiagnosis string          `json:"last_failure_diagnosis,omitempty"`
	QuarantinedAt        *time.Time      `json:"quarantined_at,omitempty"`
	QuarantineReason     string          `json:"quarantine_reason,omitempty"`
	ModelOverride        string          `json:"model_override,omitempty"`
	ProviderOverride     string          `json:"provider_override,omitempty"`
	EscalationState      string          `json:"escalation_state"`
	EscalationClaimedBy  string          `json:"escalation_claimed_by,omitempty"`
	EscalationAttempts   int             `json:"escalation_attempts"`
	CostMicroDollars     int64           `json:"cost_micro_dollars"`
	CostCapMicro         *int64          `json:"cost_cap_micro,omitempty"`
	CostCappedAt         *time.Time      `json:"cost_capped_at,omitempty"`
	// Phase 5a (Phase B) — maturity ladder surface
	Maturity             string          `json:"maturity"`
	DestinationMaturity  string          `json:"destination_maturity,omitempty"`
	RevisionCount        int             `json:"revision_count"`
	Scenarios            json.RawMessage `json:"scenarios,omitempty"`
	Spec                 string          `json:"spec,omitempty"`
	// Batch G.4 — file destination + materialization (i3 rename: materialized_at → file_enqueued_at)
	FileDestination          string     `json:"file_destination,omitempty"`
	FileEnqueuedAt           *time.Time `json:"file_enqueued_at,omitempty"`
	PipelineFileTemplate     string     `json:"pipeline_file_template,omitempty"`
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
		`SELECT wi.id::text, wi.slug, wi.pipeline_family, wi.current_stage, wi.status,
		        coalesce(wi.actor, ''),
		        coalesce(wi.tokens_in, 0), coalesce(wi.tokens_out, 0),
		        wi.token_budget,
		        wi.created_at, wi.updated_at, wi.completed_at,
		        wi.input, wi.stage_results,
		        coalesce(wi.session_ids, ARRAY[]::text[]),
		        coalesce(wi.error, ''),
		        coalesce(wi.failure_count, 0),
		        coalesce(wi.last_failure_reason, ''),
		        coalesce(wi.last_failure_diagnosis, ''),
		        wi.quarantined_at,
		        coalesce(wi.quarantine_reason, ''),
		        coalesce(wi.model_override, ''),
		        coalesce(wi.provider_override, ''),
		        coalesce(wi.escalation_state, 'normal'),
		        coalesce(wi.escalation_claimed_by, ''),
		        coalesce(wi.escalation_attempts, 0),
		        coalesce(wi.cost_micro_dollars, 0),
		        wi.cost_cap_micro,
		        wi.cost_capped_at,
		        coalesce(wi.maturity, 'raw'),
		        coalesce(wi.destination_maturity, ''),
		        coalesce(wi.revision_count, 0),
		        coalesce(wi.scenarios, '[]'::jsonb),
		        coalesce(wi.spec, ''),
		        coalesce(wi.file_destination, ''),
		        wi.file_enqueued_at,
		        coalesce(p.file_destination_template, ''),
		        coalesce(wi.origin, 'human'),
		        coalesce(wi.project_association, ''),
		        coalesce(wi.parent_work_item_id::text, '')
		   FROM stewards.work_items wi
		   LEFT JOIN stewards.pipelines p ON p.family = wi.pipeline_family
		  WHERE wi.`+whereSQL,
		whereArg,
	).Scan(&wd.ID, &wd.Slug, &wd.Pipeline, &wd.CurrentStage, &wd.Status,
		&wd.Actor, &wd.TokensIn, &wd.TokensOut, &wd.TokenBudget,
		&wd.CreatedAt, &wd.UpdatedAt, &wd.CompletedAt,
		&wd.Input, &wd.StageResults, &wd.SessionIDs, &wd.Error,
		&wd.FailureCount, &wd.LastFailureReason, &wd.LastFailureDiagnosis,
		&wd.QuarantinedAt, &wd.QuarantineReason,
		&wd.ModelOverride, &wd.ProviderOverride,
		&wd.EscalationState, &wd.EscalationClaimedBy, &wd.EscalationAttempts,
		&wd.CostMicroDollars, &wd.CostCapMicro, &wd.CostCappedAt,
		&wd.Maturity, &wd.DestinationMaturity, &wd.RevisionCount,
		&wd.Scenarios, &wd.Spec,
		&wd.FileDestination, &wd.FileEnqueuedAt, &wd.PipelineFileTemplate,
		&wd.Origin, &wd.ProjectAssociation, &wd.ParentWorkItemID)
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, wd)
}

// =====================================================================
// Phase 4j — cost_events + steward_actions endpoints for WorkItemDetail
// =====================================================================

type costEvent struct {
	ID                  int64      `json:"id"`
	AttemptSeq          int        `json:"attempt_seq"`
	At                  *time.Time `json:"at,omitempty"`
	Provider            string     `json:"provider"`
	Model               string     `json:"model"`
	InputTokens         int        `json:"input_tokens"`
	OutputTokens        int        `json:"output_tokens"`
	CacheWriteTokens    int        `json:"cache_write_tokens"`
	CacheReadTokens     int        `json:"cache_read_tokens"`
	MicroDollars        int64      `json:"micro_dollars"`
	PricingEffectiveAt  *time.Time `json:"pricing_effective_at,omitempty"`
	Notes               string     `json:"notes,omitempty"`
}

type costEventsResp struct {
	Items            []costEvent `json:"items"`
	TotalEvents      int         `json:"total_events"`
	TotalMicro       int64       `json:"total_micro_dollars"`
	WorkItemCostMicro int64      `json:"work_item_cost_micro"`
	CostCapMicro     *int64      `json:"cost_cap_micro,omitempty"`
	CostCappedAt     *time.Time  `json:"cost_capped_at,omitempty"`
}

func (d *Deps) workItemsCostHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id := r.URL.Query().Get("id")
	if id == "" {
		writeErr(w, http.StatusBadRequest, "id query param required")
		return
	}

	resp := costEventsResp{Items: []costEvent{}}

	// Fetch the work_item's denormalized cost summary first
	if err := d.Pool.QueryRow(ctx,
		`SELECT coalesce(cost_micro_dollars, 0), cost_cap_micro, cost_capped_at
		   FROM stewards.work_items WHERE id = $1::uuid`,
		id,
	).Scan(&resp.WorkItemCostMicro, &resp.CostCapMicro, &resp.CostCappedAt); err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}

	rows, err := d.Pool.Query(ctx,
		`SELECT id, attempt_seq, at, provider, model,
		        input_tokens, output_tokens, cache_write_tokens, cache_read_tokens,
		        micro_dollars, pricing_effective_at, coalesce(notes, '')
		   FROM stewards.cost_events
		  WHERE work_item_id = $1::uuid
		  ORDER BY id ASC`,
		id,
	)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		var ev costEvent
		if err := rows.Scan(&ev.ID, &ev.AttemptSeq, &ev.At, &ev.Provider, &ev.Model,
			&ev.InputTokens, &ev.OutputTokens, &ev.CacheWriteTokens, &ev.CacheReadTokens,
			&ev.MicroDollars, &ev.PricingEffectiveAt, &ev.Notes); err == nil {
			resp.Items = append(resp.Items, ev)
			resp.TotalMicro += ev.MicroDollars
		}
	}
	resp.TotalEvents = len(resp.Items)
	writeJSON(w, http.StatusOK, resp)
}

type stewardAction struct {
	ID          int64           `json:"id"`
	At          *time.Time      `json:"at,omitempty"`
	Observation string          `json:"observation"`
	Diagnosis   string          `json:"diagnosis,omitempty"`
	Action      string          `json:"action"`
	Details     json.RawMessage `json:"details,omitempty"`
	ModelUsed   string          `json:"model_used,omitempty"`
	CostMicro   *int64          `json:"cost_micro,omitempty"`
}

type stewardActionsResp struct {
	Items []stewardAction `json:"items"`
	Count int             `json:"count"`
}

func (d *Deps) workItemsActionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id := r.URL.Query().Get("id")
	if id == "" {
		writeErr(w, http.StatusBadRequest, "id query param required")
		return
	}

	resp := stewardActionsResp{Items: []stewardAction{}}
	rows, err := d.Pool.Query(ctx,
		`SELECT id, at, observation, coalesce(diagnosis, ''), action,
		        details, coalesce(model_used, ''), cost_micro
		   FROM stewards.steward_actions
		  WHERE work_item_id = $1::uuid
		  ORDER BY id DESC
		  LIMIT 50`,
		id,
	)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		var a stewardAction
		if err := rows.Scan(&a.ID, &a.At, &a.Observation, &a.Diagnosis, &a.Action,
			&a.Details, &a.ModelUsed, &a.CostMicro); err == nil {
			resp.Items = append(resp.Items, a)
		}
	}
	resp.Count = len(resp.Items)
	writeJSON(w, http.StatusOK, resp)
}

// =====================================================================
// Phase 5a (Phase B) — gate_decisions audit endpoint for WorkItemDetail
// =====================================================================

type gateDecisionRow struct {
	ID             int64           `json:"id"`
	At             *time.Time      `json:"at,omitempty"`
	FromMaturity   string          `json:"from_maturity"`
	Action         string          `json:"action"`
	Reasoning      string          `json:"reasoning,omitempty"`
	Feedback       string          `json:"feedback,omitempty"`
	WorkID         *int64          `json:"work_id,omitempty"`
	RevisionCount  int             `json:"revision_count"`
	RawResponse    json.RawMessage `json:"raw_response,omitempty"`
}

type gateDecisionsResp struct {
	Items []gateDecisionRow `json:"items"`
	Count int               `json:"count"`
}

func (d *Deps) workItemsGateDecisionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id := r.URL.Query().Get("id")
	if id == "" {
		writeErr(w, http.StatusBadRequest, "id query param required")
		return
	}

	resp := gateDecisionsResp{Items: []gateDecisionRow{}}
	rows, err := d.Pool.Query(ctx,
		`SELECT id, at, from_maturity, action,
		        coalesce(reasoning, ''), coalesce(feedback, ''),
		        work_id, revision_count, coalesce(raw_response, '{}'::jsonb)
		   FROM stewards.gate_decisions
		  WHERE work_item_id = $1::uuid
		  ORDER BY at DESC, id DESC
		  LIMIT 50`,
		id,
	)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		var g gateDecisionRow
		if err := rows.Scan(&g.ID, &g.At, &g.FromMaturity, &g.Action,
			&g.Reasoning, &g.Feedback, &g.WorkID, &g.RevisionCount,
			&g.RawResponse); err == nil {
			resp.Items = append(resp.Items, g)
		}
	}
	resp.Count = len(resp.Items)
	writeJSON(w, http.StatusOK, resp)
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

// =====================================================================
// Batch G.4 — file destination + materialize endpoints
// =====================================================================

type setFileDestinationReq struct {
	ID              string `json:"id"`
	FileDestination string `json:"file_destination"` // empty string = DB-only (NULL in DB)
}

type setFileDestinationResp struct {
	ID              string `json:"id"`
	FileDestination string `json:"file_destination"`
}

func (d *Deps) workItemsSetFileDestinationHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req setFileDestinationReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "decode body: "+err.Error())
		return
	}
	if req.ID == "" {
		writeErr(w, http.StatusBadRequest, "id required")
		return
	}

	var arg any
	if req.FileDestination == "" {
		arg = nil
	} else {
		arg = req.FileDestination
	}

	_, err := d.Pool.Exec(ctx,
		`UPDATE stewards.work_items
		    SET file_destination = $1
		  WHERE id = $2::uuid`,
		arg, req.ID)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, setFileDestinationResp{
		ID: req.ID, FileDestination: req.FileDestination,
	})
}

type materializeFileReq struct {
	ID string `json:"id"`
}

type materializeFileResp struct {
	PendingFileWriteID *int64 `json:"pending_file_write_id,omitempty"`
	Skipped            bool   `json:"skipped"` // true when file_destination IS NULL
	SkipReason         string `json:"skip_reason,omitempty"`
}

func (d *Deps) workItemsMaterializeFileHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req materializeFileReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "decode body: "+err.Error())
		return
	}
	if req.ID == "" {
		writeErr(w, http.StatusBadRequest, "id required")
		return
	}

	var pwid *int64
	err := d.Pool.QueryRow(ctx,
		`SELECT stewards.enqueue_work_item_file($1::uuid, 'ui')`,
		req.ID,
	).Scan(&pwid)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	resp := materializeFileResp{PendingFileWriteID: pwid}
	if pwid == nil {
		resp.Skipped = true
		resp.SkipReason = "work_items.file_destination is NULL (DB-only)"
	}
	writeJSON(w, http.StatusOK, resp)
}

// =====================================================================
// H.3-followup-B — agent_planning proposal action handlers
//
// Three actions a human can take on a proposed work_item:
//
//   ratify          — advance maturity from raw to researched. The
//                     work_item stays at status='pending'. Marks
//                     the proposal as accepted; next action is
//                     dispatch.
//   dispatch        — call work_item_dispatch_stage to fire the
//                     current stage. Returns the chat work_queue id.
//   cancel-proposal — set status='cancelled' with a quarantine_reason
//                     note. The row stays in the DB as historical
//                     record but won't surface in active queues.
//
// All three accept a simple { id } JSON body and return a small
// confirmation envelope. UI shows them only when origin='agent_planning'
// AND maturity='raw' (ratify+cancel) or AND maturity='researched'
// (dispatch).
// =====================================================================

type workItemActionReq struct {
	ID     string `json:"id"`
	Reason string `json:"reason,omitempty"` // optional, used by cancel-proposal
}

type workItemActionResp struct {
	ID          string `json:"id"`
	Status      string `json:"status,omitempty"`
	Maturity    string `json:"maturity,omitempty"`
	WorkQueueID *int64 `json:"work_queue_id,omitempty"`
	Message     string `json:"message,omitempty"`
}

// workItemsRatifyHandler advances maturity from 'raw' to 'researched'
// for proposed work_items. Validates that origin='agent_planning' to
// prevent the button from being a general-purpose maturity-advance lever.
func (d *Deps) workItemsRatifyHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req workItemActionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "decode body: "+err.Error())
		return
	}
	if req.ID == "" {
		writeErr(w, http.StatusBadRequest, "id required")
		return
	}

	var resp workItemActionResp
	resp.ID = req.ID
	err := d.Pool.QueryRow(ctx,
		`UPDATE stewards.work_items
		    SET maturity = 'researched',
		        updated_at = now()
		  WHERE id = $1::uuid
		    AND origin = 'agent_planning'
		    AND maturity = 'raw'
		  RETURNING status, maturity`,
		req.ID,
	).Scan(&resp.Status, &resp.Maturity)
	if err != nil {
		writeErr(w, http.StatusBadRequest,
			"ratify failed (must be origin=agent_planning + maturity=raw): "+err.Error())
		return
	}
	resp.Message = "ratified: maturity advanced raw → researched"
	writeJSON(w, http.StatusOK, resp)
}

// workItemsDispatchHandler invokes work_item_dispatch_stage on a
// work_item. The SQL function handles its own validation (pending or
// awaiting_review status, current_stage resolvable, etc.).
func (d *Deps) workItemsDispatchHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req workItemActionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "decode body: "+err.Error())
		return
	}
	if req.ID == "" {
		writeErr(w, http.StatusBadRequest, "id required")
		return
	}

	var resp workItemActionResp
	resp.ID = req.ID
	var wqID int64
	err := d.Pool.QueryRow(ctx,
		`SELECT stewards.work_item_dispatch_stage($1::uuid)`,
		req.ID,
	).Scan(&wqID)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "dispatch failed: "+err.Error())
		return
	}
	resp.WorkQueueID = &wqID
	resp.Message = "dispatched"
	writeJSON(w, http.StatusOK, resp)
}

// workItemsCancelProposalHandler marks a proposed work_item cancelled.
// Stores the reason in quarantine_reason for historical record.
// Restricted to origin='agent_planning' so this isn't a general cancel.
func (d *Deps) workItemsCancelProposalHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req workItemActionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "decode body: "+err.Error())
		return
	}
	if req.ID == "" {
		writeErr(w, http.StatusBadRequest, "id required")
		return
	}
	reason := req.Reason
	if reason == "" {
		reason = "user-cancelled via UI"
	}

	var resp workItemActionResp
	resp.ID = req.ID
	err := d.Pool.QueryRow(ctx,
		`UPDATE stewards.work_items
		    SET status = 'cancelled',
		        quarantine_reason = $2,
		        updated_at = now()
		  WHERE id = $1::uuid
		    AND origin = 'agent_planning'
		  RETURNING status`,
		req.ID, reason,
	).Scan(&resp.Status)
	if err != nil {
		writeErr(w, http.StatusBadRequest,
			"cancel failed (must be origin=agent_planning): "+err.Error())
		return
	}
	resp.Message = "cancelled"
	writeJSON(w, http.StatusOK, resp)
}

// =====================================================================
// Edit + AI-revise — third proposal mode (this session)
//
// Two affordances on top of Ratify/Dispatch/Cancel:
//   1. Direct edit fields (no AI) — UPDATE binding_question / slug /
//      pipeline_family / project_association directly. Restricted to
//      origin=agent_planning AND status != cancelled.
//   2. AI revise-with-feedback — create + dispatch a revise-proposal
//      work_item. When verified, the UI fetches pending-revisions and
//      shows a diff card. User clicks Accept (calls apply-revision SQL)
//      or Reject (UPDATE status=cancelled).
// =====================================================================

type editProposalReq struct {
	ID                  string  `json:"id"`
	BindingQuestion     *string `json:"binding_question,omitempty"`
	Slug                *string `json:"slug,omitempty"`
	PipelineFamilyHint  *string `json:"pipeline_family_hint,omitempty"`
	ProjectAssociation  *string `json:"project_association,omitempty"`
	Rationale           *string `json:"rationale,omitempty"`
}

// workItemsEditProposalHandler does direct UPDATE of editable fields on
// a proposed work_item. No AI involved. Fields are optional; only the
// ones present in the request body get updated. Restricted to
// origin=agent_planning AND status != cancelled so the endpoint can't
// become a general-purpose field-edit lever.
func (d *Deps) workItemsEditProposalHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req editProposalReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "decode body: "+err.Error())
		return
	}
	if req.ID == "" {
		writeErr(w, http.StatusBadRequest, "id required")
		return
	}

	// Validate before touching the DB.
	if req.Slug != nil && *req.Slug != "" {
		// Defensive: same regex the substrate enforces internally.
		matched := true
		for _, c := range *req.Slug {
			if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
				matched = false
				break
			}
		}
		if !matched {
			writeErr(w, http.StatusBadRequest, "slug must match ^[a-z0-9-]+$")
			return
		}
	}
	if req.BindingQuestion != nil && len(*req.BindingQuestion) > 0 && len(*req.BindingQuestion) < 20 {
		writeErr(w, http.StatusBadRequest, "binding_question must be ≥20 chars")
		return
	}

	// Build dynamic UPDATE. We always UPDATE input jsonb for
	// binding_question / rationale; UPDATE top-level columns for slug
	// / pipeline_family / project_association.
	tx, err := d.Pool.Begin(ctx)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback(ctx)

	// Guard: row must be a proposal in editable state.
	var origin, status string
	err = tx.QueryRow(ctx,
		`SELECT origin, status FROM stewards.work_items WHERE id = $1::uuid`,
		req.ID,
	).Scan(&origin, &status)
	if err != nil {
		writeErr(w, http.StatusNotFound, "work_item not found: "+err.Error())
		return
	}
	if origin != "agent_planning" {
		writeErr(w, http.StatusBadRequest, "edit only allowed for origin=agent_planning (got "+origin+")")
		return
	}
	if status == "cancelled" {
		writeErr(w, http.StatusBadRequest, "cannot edit cancelled proposal")
		return
	}

	if req.BindingQuestion != nil {
		if _, err := tx.Exec(ctx,
			`UPDATE stewards.work_items
			    SET input = input || jsonb_build_object('binding_question', $2::text),
			        updated_at = now()
			  WHERE id = $1::uuid`,
			req.ID, *req.BindingQuestion); err != nil {
			writeErr(w, http.StatusBadRequest, "update binding: "+err.Error())
			return
		}
	}
	if req.Rationale != nil {
		if _, err := tx.Exec(ctx,
			`UPDATE stewards.work_items
			    SET input = input || jsonb_build_object('rationale_from_planning', $2::text),
			        updated_at = now()
			  WHERE id = $1::uuid`,
			req.ID, *req.Rationale); err != nil {
			writeErr(w, http.StatusBadRequest, "update rationale: "+err.Error())
			return
		}
	}
	if req.Slug != nil && *req.Slug != "" {
		if _, err := tx.Exec(ctx,
			`UPDATE stewards.work_items SET slug = $2, updated_at = now() WHERE id = $1::uuid`,
			req.ID, *req.Slug); err != nil {
			writeErr(w, http.StatusBadRequest, "update slug: "+err.Error())
			return
		}
	}
	if req.PipelineFamilyHint != nil && *req.PipelineFamilyHint != "" {
		// Validate pipeline_family exists and update current_stage to its first stage.
		if _, err := tx.Exec(ctx,
			`UPDATE stewards.work_items
			    SET pipeline_family = $2,
			        current_stage   = stewards.pipeline_first_stage_name($2),
			        updated_at = now()
			  WHERE id = $1::uuid
			    AND EXISTS (SELECT 1 FROM stewards.pipelines WHERE family = $2)`,
			req.ID, *req.PipelineFamilyHint); err != nil {
			writeErr(w, http.StatusBadRequest, "update pipeline_family: "+err.Error())
			return
		}
	}
	if req.ProjectAssociation != nil {
		// Empty string clears; nil = unchanged (above).
		var pa any = *req.ProjectAssociation
		if *req.ProjectAssociation == "" {
			pa = nil
		}
		if _, err := tx.Exec(ctx,
			`UPDATE stewards.work_items SET project_association = $2, updated_at = now() WHERE id = $1::uuid`,
			req.ID, pa); err != nil {
			writeErr(w, http.StatusBadRequest, "update project: "+err.Error())
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		writeErr(w, http.StatusInternalServerError, "commit: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":      req.ID,
		"message": "edited",
	})
}

type reviseWithFeedbackReq struct {
	ID       string `json:"id"`
	Feedback string `json:"feedback"`
}

type reviseWithFeedbackResp struct {
	ReviseWorkItemID string `json:"revise_work_item_id"`
	WorkQueueID      int64  `json:"work_queue_id"`
	Message          string `json:"message"`
}

// workItemsReviseWithFeedbackHandler creates a revise-proposal work_item
// linked to the proposal being revised, populates its input with the
// original's fields + parent plan excerpt + user feedback, and dispatches
// the revise stage. The UI then polls /pending-revisions.
func (d *Deps) workItemsReviseWithFeedbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var req reviseWithFeedbackReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "decode body: "+err.Error())
		return
	}
	if req.ID == "" || req.Feedback == "" {
		writeErr(w, http.StatusBadRequest, "id and feedback required")
		return
	}
	if len(req.Feedback) < 5 {
		writeErr(w, http.StatusBadRequest, "feedback must be ≥5 chars")
		return
	}

	// Pull the original proposal + parent plan excerpt (synthesize output)
	// to compose the revise stage's input.
	var (
		originSlug, originBinding, originRationale string
		originHint, originProject                  string
		originID, originParentID                   string
		parentPlanExcerpt                          string
	)
	err := d.Pool.QueryRow(ctx, `
		SELECT wi.id::text,
		       wi.slug,
		       coalesce(wi.input->>'binding_question', ''),
		       coalesce(wi.input->>'rationale_from_planning', ''),
		       coalesce(wi.pipeline_family, ''),
		       coalesce(wi.project_association, ''),
		       coalesce(wi.parent_work_item_id::text, ''),
		       coalesce(
		           substring(
		               (SELECT (parent.stage_results -> 'synthesize' -> 'output') #>> '{}'
		                  FROM stewards.work_items parent
		                 WHERE parent.id = wi.parent_work_item_id)
		           FROM 1 FOR 4000),
		           '(no parent plan excerpt available)'
		       )
		  FROM stewards.work_items wi
		 WHERE wi.id = $1::uuid
		   AND wi.origin = 'agent_planning'`,
		req.ID,
	).Scan(&originID, &originSlug, &originBinding, &originRationale,
		&originHint, &originProject, &originParentID, &parentPlanExcerpt)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "proposal lookup failed: "+err.Error())
		return
	}

	// Build the revise work_item's input jsonb.
	input := map[string]any{
		"original_proposal_id":           originID,
		"original_slug":                  originSlug,
		"original_binding_question":      originBinding,
		"original_rationale":             originRationale,
		"original_pipeline_family_hint":  originHint,
		"original_project_association":   originProject,
		"parent_plan_excerpt":            parentPlanExcerpt,
		"feedback":                       req.Feedback,
	}
	inputJSON, _ := json.Marshal(input)

	// work_item_create gets us a fresh row with the right pipeline +
	// intent + first stage. We then set parent_work_item_id +
	// cost_cap_micro before dispatching.
	tx, err := d.Pool.Begin(ctx)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback(ctx)

	var reviseID string
	err = tx.QueryRow(ctx, `
		SELECT stewards.work_item_create(
		    'revise-proposal',
		    $1::jsonb,
		    $2::text,
		    'human',
		    NULL,
		    (SELECT id FROM stewards.intents WHERE slug='planning-partner')
		)::text`,
		inputJSON,
		"revise-"+originSlug+"-"+time.Now().Format("20060102150405"),
	).Scan(&reviseID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "work_item_create: "+err.Error())
		return
	}

	// Link to parent + set cost cap.
	if _, err := tx.Exec(ctx, `
		UPDATE stewards.work_items
		   SET parent_work_item_id = $1::uuid,
		       cost_cap_micro = 100000,
		       project_association = $2
		 WHERE id = $3::uuid`,
		req.ID, originProject, reviseID); err != nil {
		writeErr(w, http.StatusInternalServerError, "link parent: "+err.Error())
		return
	}

	if err := tx.Commit(ctx); err != nil {
		writeErr(w, http.StatusInternalServerError, "commit: "+err.Error())
		return
	}

	// Dispatch the revise stage (outside tx — substrate auto-fire).
	var wqID int64
	if err := d.Pool.QueryRow(ctx,
		`SELECT stewards.work_item_dispatch_stage($1::uuid)`, reviseID,
	).Scan(&wqID); err != nil {
		writeErr(w, http.StatusInternalServerError, "dispatch: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, reviseWithFeedbackResp{
		ReviseWorkItemID: reviseID,
		WorkQueueID:      wqID,
		Message:          "revise dispatched; poll pending-revisions",
	})
}

type pendingRevisionRow struct {
	ID            string     `json:"id"`
	Slug          string     `json:"slug"`
	Status        string     `json:"status"`
	Maturity      string     `json:"maturity"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	CostMicro     int64      `json:"cost_micro"`
	Feedback      string     `json:"feedback,omitempty"`
	RevisionJSON  json.RawMessage `json:"revision_json,omitempty"`
}

type pendingRevisionsResp struct {
	Revisions []pendingRevisionRow `json:"revisions"`
	Count     int                  `json:"count"`
}

// workItemsPendingRevisionsHandler lists revise-proposal work_items
// whose parent_work_item_id matches and which haven't been applied
// or cancelled yet. The frontend polls this every few seconds while
// a revise is in flight, then renders any completed ones as diff cards.
func (d *Deps) workItemsPendingRevisionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id := r.URL.Query().Get("id")
	if id == "" {
		writeErr(w, http.StatusBadRequest, "id required")
		return
	}

	resp := pendingRevisionsResp{Revisions: []pendingRevisionRow{}}
	rows, err := d.Pool.Query(ctx, `
		SELECT id::text, slug, status, maturity, created_at, completed_at,
		       coalesce(cost_micro_dollars, 0),
		       coalesce(input->>'feedback', ''),
		       CASE
		           WHEN status='completed' AND maturity='verified'
		               THEN (stage_results -> 'revise' -> 'output')
		           ELSE NULL
		       END
		  FROM stewards.work_items
		 WHERE parent_work_item_id = $1::uuid
		   AND pipeline_family = 'revise-proposal'
		   AND revision_applied_at IS NULL
		   AND status != 'cancelled'
		 ORDER BY created_at ASC`, id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		var row pendingRevisionRow
		var rawJSON *string
		if err := rows.Scan(&row.ID, &row.Slug, &row.Status, &row.Maturity,
			&row.CreatedAt, &row.CompletedAt, &row.CostMicro,
			&row.Feedback, &rawJSON); err == nil {
			if rawJSON != nil {
				row.RevisionJSON = json.RawMessage(*rawJSON)
			}
			resp.Revisions = append(resp.Revisions, row)
		}
	}
	resp.Count = len(resp.Revisions)
	writeJSON(w, http.StatusOK, resp)
}

// workItemsApplyRevisionHandler accepts a completed revise-proposal
// work_item and UPDATEs the original (its parent_work_item_id). Calls
// the apply_revision SQL function which handles validation + COALESCE
// merge of revision fields.
func (d *Deps) workItemsApplyRevisionHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req workItemActionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "decode body: "+err.Error())
		return
	}
	if req.ID == "" {
		writeErr(w, http.StatusBadRequest, "id required")
		return
	}

	var ok bool
	err := d.Pool.QueryRow(ctx,
		`SELECT stewards.apply_revision($1::uuid)`, req.ID,
	).Scan(&ok)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "apply_revision failed: "+err.Error())
		return
	}
	resp := workItemActionResp{ID: req.ID}
	if ok {
		resp.Message = "revision applied"
	} else {
		resp.Message = "revision NOT applied (already applied, rejected, or validation failed — see substrate NOTICE logs)"
	}
	writeJSON(w, http.StatusOK, resp)
}

// workItemsRejectRevisionHandler marks a revise-proposal work_item
// cancelled so it no longer shows as pending. The original proposal
// is untouched.
func (d *Deps) workItemsRejectRevisionHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req workItemActionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "decode body: "+err.Error())
		return
	}
	if req.ID == "" {
		writeErr(w, http.StatusBadRequest, "id required")
		return
	}
	reason := req.Reason
	if reason == "" {
		reason = "revision rejected via UI"
	}

	var resp workItemActionResp
	resp.ID = req.ID
	err := d.Pool.QueryRow(ctx,
		`UPDATE stewards.work_items
		    SET status = 'cancelled',
		        quarantine_reason = $2,
		        updated_at = now()
		  WHERE id = $1::uuid
		    AND pipeline_family = 'revise-proposal'
		  RETURNING status`,
		req.ID, reason,
	).Scan(&resp.Status)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "reject failed: "+err.Error())
		return
	}
	resp.Message = "revision rejected"
	writeJSON(w, http.StatusOK, resp)
}
