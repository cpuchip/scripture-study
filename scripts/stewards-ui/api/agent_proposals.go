// agent_proposals — POST /api/agent-proposals/create
//
// Batch I.2 (2026-05-12): sibling pattern to /api/work-items/create.
// Accepts an agent's structured proposal payload, creates a work_item
// with pipeline_family='agent-proposal' + origin='agent_proposal' +
// input.draft = the payload. The bgworker's validate stage then
// normalizes the draft; on advance to verified, apply_agent_proposal
// persists to studies or queues a schema-migration .sql file.
//
// Payload schema (mirrors apply_agent_proposal's expected JSON):
//
//	{
//	  "source_type": "study | lesson | note | exhibit | schema-migration",
//	  "slug":        "kebab-case-slug",
//	  "title":       "Human-readable title (10-120 chars)",
//	  "body":        "Full markdown OR SQL body",
//	  "frontmatter": { /* per-source-type metadata */ },
//	  "project_association": "string or null",
//	  "rationale":   "Why this proposal exists (20-500 chars)",
//	  "claude_attested": true  // REQUIRED only for schema-migration
//	}
//
// Auth/trust: any caller can hit this endpoint today. The kimi-trust
// gate for schema-migration is enforced inside apply_agent_proposal
// via claude_attested. Tightening (auth tokens, model attestation)
// is out of scope for I.2.

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"
)

func (d *Deps) registerAgentProposals(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/agent-proposals/create", d.agentProposalCreateHandler)
}

var (
	agentProposalSlugRegex = regexp.MustCompile(`^[a-z0-9-]+$`)
	agentProposalSourceTypes = map[string]bool{
		"study": true, "lesson": true, "note": true, "exhibit": true, "schema-migration": true,
	}
)

type agentProposalCreateReq struct {
	SourceType         string          `json:"source_type"`
	Slug               string          `json:"slug"`
	Title              string          `json:"title"`
	Body               string          `json:"body"`
	Frontmatter        json.RawMessage `json:"frontmatter,omitempty"`
	ProjectAssociation string          `json:"project_association,omitempty"`
	Rationale          string          `json:"rationale"`
	ClaudeAttested     bool            `json:"claude_attested,omitempty"`
	Actor              string          `json:"actor,omitempty"`
	Dispatch           bool            `json:"dispatch,omitempty"`
	IntentSlug         string          `json:"intent_slug,omitempty"`
}

type agentProposalCreateResp struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	WorkQueueID *int64 `json:"work_queue_id,omitempty"`
	Dispatched  bool   `json:"dispatched"`
}

func (d *Deps) agentProposalCreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var req agentProposalCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "decode body: "+err.Error())
		return
	}

	// Validation — keep close to apply_agent_proposal's checks so we fail
	// fast at the boundary instead of letting bad payloads round-trip
	// through the bgworker validate stage.
	if !agentProposalSourceTypes[req.SourceType] {
		writeErr(w, http.StatusBadRequest, "source_type must be one of: study, lesson, note, exhibit, schema-migration")
		return
	}
	if !agentProposalSlugRegex.MatchString(req.Slug) {
		writeErr(w, http.StatusBadRequest, "slug must match ^[a-z0-9-]+$")
		return
	}
	if l := len(req.Title); l < 10 || l > 120 {
		writeErr(w, http.StatusBadRequest, "title must be 10-120 chars")
		return
	}
	if len(req.Body) == 0 {
		writeErr(w, http.StatusBadRequest, "body is required")
		return
	}
	if l := len(req.Rationale); l < 20 || l > 500 {
		writeErr(w, http.StatusBadRequest, "rationale must be 20-500 chars")
		return
	}
	if req.SourceType == "schema-migration" && !req.ClaudeAttested {
		writeErr(w, http.StatusForbidden,
			"schema-migration requires claude_attested=true per kimi-trust ratification 2026-05-11")
		return
	}
	if req.Actor == "" {
		req.Actor = "agent"
	}
	if len(req.Frontmatter) == 0 {
		req.Frontmatter = json.RawMessage(`{}`)
	}

	// Resolve intent_id (default: scripture-study)
	intentSlug := req.IntentSlug
	if intentSlug == "" {
		intentSlug = "scripture-study"
	}
	var intentID string
	if err := d.Pool.QueryRow(ctx,
		`SELECT id::text FROM stewards.intents WHERE slug = $1`, intentSlug,
	).Scan(&intentID); err != nil {
		writeErr(w, http.StatusBadRequest, "intent not found: "+intentSlug)
		return
	}

	// Build input.draft jsonb. The draft IS the payload (mirroring the
	// shape apply_agent_proposal will read on advance to verified). We
	// include claude_attested even when false so the kimi-trust gate has
	// something to read for non-schema-migration types.
	draft := map[string]any{
		"source_type":         req.SourceType,
		"slug":                req.Slug,
		"title":               req.Title,
		"body":                req.Body,
		"frontmatter":         json.RawMessage(req.Frontmatter),
		"project_association": req.ProjectAssociation,
		"rationale":           req.Rationale,
		"claude_attested":     req.ClaudeAttested,
	}
	draftJSON, err := json.Marshal(draft)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "encode draft: "+err.Error())
		return
	}
	inputJSON, err := json.Marshal(map[string]json.RawMessage{
		"draft": json.RawMessage(draftJSON),
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "encode input: "+err.Error())
		return
	}

	// Create the work_item via the existing SQL function. work_item_create
	// defaults origin to 'human'; we UPDATE to 'agent_proposal' below.
	// Slug uniqueness on work_items table — agent-proposal slugs are
	// independent of studies slugs.
	wiSlug := req.Slug + "-proposal-" + fmt.Sprintf("%d", time.Now().Unix())
	var newID string
	err = d.Pool.QueryRow(ctx,
		`SELECT stewards.work_item_create($1, $2::jsonb, $3, $4, $5, $6::uuid)::text`,
		"agent-proposal", string(inputJSON), wiSlug, req.Actor, nil, intentID,
	).Scan(&newID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "create: "+err.Error())
		return
	}

	// origin='agent_proposal' + project_association passthrough.
	if _, err := d.Pool.Exec(ctx,
		`UPDATE stewards.work_items
		    SET origin = 'agent_proposal',
		        project_association = NULLIF($2, '')
		  WHERE id = $1::uuid`,
		newID, req.ProjectAssociation,
	); err != nil {
		writeErr(w, http.StatusInternalServerError, "set origin: "+err.Error())
		return
	}

	resp := agentProposalCreateResp{ID: newID, Slug: wiSlug}

	if req.Dispatch {
		var wqID int64
		if err := d.Pool.QueryRow(ctx,
			`SELECT stewards.work_item_dispatch_stage($1::uuid, NULL)`, newID,
		).Scan(&wqID); err != nil {
			writeJSON(w, http.StatusOK, map[string]any{
				"id":         newID,
				"slug":       wiSlug,
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
