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
	// Batch G.4 — file destination + materialization
	FileDestination          string     `json:"file_destination,omitempty"`
	MaterializedAt           *time.Time `json:"materialized_at,omitempty"`
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
		        wi.materialized_at,
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
		&wd.FileDestination, &wd.MaterializedAt, &wd.PipelineFileTemplate,
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
