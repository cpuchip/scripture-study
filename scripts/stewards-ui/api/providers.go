// providers endpoint — list provider+model pairs the substrate has
// loaded. Used by the NewWork form's model picker. Backed by
// stewards.providers_loaded() which the bgworker exposes.

package api

import (
	"context"
	"net/http"
	"time"
)

func (d *Deps) registerProviders(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/providers", d.providersHandler)
}

type providerRow struct {
	Name         string `json:"name"`
	BaseURL      string `json:"base_url"`
	DefaultModel string `json:"default_model"`
	Kind         string `json:"kind"`
	HasAPIKey    bool   `json:"has_api_key"`
}

type providersResp struct {
	Items []providerRow `json:"items"`
}

func (d *Deps) providersHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp := providersResp{Items: []providerRow{}}
	rows, err := d.Pool.Query(ctx,
		`SELECT name, base_url, default_model, kind, has_api_key
		   FROM stewards.providers_loaded()
		   ORDER BY name`,
	)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		var p providerRow
		if err := rows.Scan(&p.Name, &p.BaseURL, &p.DefaultModel, &p.Kind, &p.HasAPIKey); err == nil {
			resp.Items = append(resp.Items, p)
		}
	}
	writeJSON(w, http.StatusOK, resp)
}
