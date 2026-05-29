// models endpoint — list of provider+model pairs the substrate knows
// about, with prices and provider-default flag. Used by:
//   - Brainstorm.vue per-lens model override <datalist> autocompletion
//   - Models.vue dedicated catalog browse view
// Pulls from model_pricing (the authoritative cost table) and joins
// providers_loaded() to mark which model is each provider's default.

package api

import (
	"context"
	"net/http"
	"time"
)

func (d *Deps) registerModels(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/models", d.modelsHandler)
}

type modelRow struct {
	Provider             string `json:"provider"`
	Model                string `json:"model"`
	InputMicroPerMtok    int64  `json:"input_micro_per_mtok"`
	OutputMicroPerMtok   int64  `json:"output_micro_per_mtok"`
	CacheWriteMicroPerMtok *int64 `json:"cache_write_micro_per_mtok,omitempty"`
	CacheReadMicroPerMtok  *int64 `json:"cache_read_micro_per_mtok,omitempty"`
	IsProviderDefault    bool   `json:"is_provider_default"`
	Notes                string `json:"notes,omitempty"`
}

type modelsResp struct {
	Items []modelRow `json:"items"`
	Total int        `json:"total"`
}

func (d *Deps) modelsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// LEFT JOIN providers_loaded so a model without a matching live
	// provider still appears (catalog rows can outlive provider state).
	// Latest pricing only — model_pricing PK is (provider, model,
	// effective_at) so DISTINCT ON pulls the most recent row per pair.
	rows, err := d.Pool.Query(ctx, `
		SELECT DISTINCT ON (mp.provider, mp.model)
		       mp.provider, mp.model,
		       mp.input_micro_per_mtok, mp.output_micro_per_mtok,
		       mp.cache_write_micro_per_mtok, mp.cache_read_micro_per_mtok,
		       (pl.default_model IS NOT NULL
		         AND pl.default_model = mp.model) AS is_provider_default,
		       COALESCE(mp.notes, '')
		  FROM stewards.model_pricing mp
		  LEFT JOIN stewards.providers_loaded() pl
		    ON pl.name = mp.provider
		 ORDER BY mp.provider, mp.model, mp.effective_at DESC`)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	resp := modelsResp{Items: []modelRow{}}
	for rows.Next() {
		var m modelRow
		if err := rows.Scan(&m.Provider, &m.Model,
			&m.InputMicroPerMtok, &m.OutputMicroPerMtok,
			&m.CacheWriteMicroPerMtok, &m.CacheReadMicroPerMtok,
			&m.IsProviderDefault, &m.Notes); err == nil {
			resp.Items = append(resp.Items, m)
		}
	}
	resp.Total = len(resp.Items)
	writeJSON(w, http.StatusOK, resp)
}
