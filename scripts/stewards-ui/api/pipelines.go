// pipelines endpoints — Batch G.4.4.
// Surfaces stewards.pipelines for the NewWork form's dynamic dropdown
// + file_destination_template prefill.

package api

import (
	"context"
	"net/http"
	"time"
)

func (d *Deps) registerPipelines(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/pipelines/list", d.pipelinesListHandler)
}

type pipelineRow struct {
	Family                  string `json:"family"`
	Description             string `json:"description"`
	SabbathEnabled          bool   `json:"sabbath_enabled"`
	AtonementEnabled        bool   `json:"atonement_enabled"`
	FileDestinationTemplate string `json:"file_destination_template,omitempty"`
	FileContentJsonpath     string `json:"file_content_jsonpath,omitempty"`
}

type pipelinesListResp struct {
	Items []pipelineRow `json:"items"`
	Total int           `json:"total"`
}

func (d *Deps) pipelinesListHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rows, err := d.Pool.Query(ctx, `
		SELECT family, coalesce(description, ''),
		       coalesce(sabbath_enabled, false),
		       coalesce(atonement_enabled, false),
		       coalesce(file_destination_template, ''),
		       coalesce(file_content_jsonpath, '')
		  FROM stewards.pipelines
		  ORDER BY family`)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	resp := pipelinesListResp{Items: []pipelineRow{}}
	for rows.Next() {
		var p pipelineRow
		if err := rows.Scan(&p.Family, &p.Description,
			&p.SabbathEnabled, &p.AtonementEnabled,
			&p.FileDestinationTemplate, &p.FileContentJsonpath); err == nil {
			resp.Items = append(resp.Items, p)
		}
	}
	resp.Total = len(resp.Items)
	writeJSON(w, http.StatusOK, resp)
}
