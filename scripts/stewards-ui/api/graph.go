// graph endpoint — studies + citations subgraph for the v1 graph view.
//
// Phase 1 of graph: nodes are studies; edges are derived from
// stewards.study_citations() expansion. Each cited_uri that maps to a
// known study slug becomes an edge. Refs that don't resolve to a
// substrate study (gospel-library URIs, etc.) are dropped — the
// substrate is the graph here, not the corpus at large.

package api

import (
	"context"
	"net/http"
	"time"
)

func (d *Deps) registerGraph(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/graph/studies-citations", d.graphStudiesCitationsHandler)
}

type graphNode struct {
	ID    string `json:"id"`    // = slug
	Label string `json:"label"`
	Kind  string `json:"kind,omitempty"`
}

type graphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Weight int    `json:"weight,omitempty"`
}

type graphResp struct {
	Nodes []graphNode `json:"nodes"`
	Edges []graphEdge `json:"edges"`
}

func (d *Deps) graphStudiesCitationsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	limitNodes := atoiDefault(r.URL.Query().Get("limit"), 200, 10, 1000)

	// Initialize as empty slices (not nil) so JSON marshals as
	// `[]` not `null` — frontend consumers do .length safely.
	resp := graphResp{Nodes: []graphNode{}, Edges: []graphEdge{}}
	idx := map[string]bool{} // slug → present

	// Get most-recently-updated studies as nodes (deterministic + bounded)
	rows, err := d.Pool.Query(ctx,
		`SELECT slug, kind, coalesce(frontmatter->>'title', slug) AS title
		   FROM stewards.studies
		   ORDER BY updated_at DESC NULLS LAST
		   LIMIT $1`,
		limitNodes,
	)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	for rows.Next() {
		var n graphNode
		if err := rows.Scan(&n.ID, &n.Kind, &n.Label); err == nil {
			resp.Nodes = append(resp.Nodes, n)
			idx[n.ID] = true
		}
	}
	rows.Close()

	// For each node, pull its citations and emit edges to other in-graph slugs.
	// Substrate's study_citations returns cited_uri (often a path); we extract
	// the basename without .md as a candidate slug.
	for _, n := range resp.Nodes {
		erows, err := d.Pool.Query(ctx,
			`SELECT cited_uri, citation_count FROM stewards.study_citations($1)`,
			n.ID,
		)
		if err != nil {
			continue
		}
		for erows.Next() {
			var uri string
			var count int
			if err := erows.Scan(&uri, &count); err != nil {
				continue
			}
			target := slugFromURI(uri)
			if target == "" || target == n.ID || !idx[target] {
				continue
			}
			resp.Edges = append(resp.Edges, graphEdge{
				Source: n.ID,
				Target: target,
				Weight: count,
			})
		}
		erows.Close()
	}

	writeJSON(w, http.StatusOK, resp)
}

// slugFromURI: strip leading paths and trailing .md.
//   "/study/order-of-god.md" → "order-of-god"
//   "study/order-of-god.md"  → "order-of-god"
//   "order-of-god"           → "order-of-god"
//   "/gospel-library/eng/scriptures/bofm/alma/40.md" → "" (not a study)
//
// We only count it as a study slug if the path looks like .../study/<slug>.md
// or it has no path prefix at all (raw slug). gospel-library refs are dropped.
func slugFromURI(uri string) string {
	if uri == "" {
		return ""
	}
	// Strip query/fragment if any
	for _, sep := range []byte{'?', '#'} {
		for i := len(uri) - 1; i >= 0; i-- {
			if uri[i] == sep {
				uri = uri[:i]
				break
			}
		}
	}
	// Find last "/" — basename
	last := -1
	for i := len(uri) - 1; i >= 0; i-- {
		if uri[i] == '/' {
			last = i
			break
		}
	}
	prefix := ""
	base := uri
	if last >= 0 {
		prefix = uri[:last]
		base = uri[last+1:]
	}
	// If there's a prefix, only accept "study" or "../study" containers
	if prefix != "" {
		ok := false
		for _, want := range []string{"study", "/study", "../study", "./study"} {
			if endsWith(prefix, want) {
				ok = true
				break
			}
		}
		if !ok {
			return ""
		}
	}
	// Strip trailing .md
	if endsWith(base, ".md") {
		base = base[:len(base)-3]
	}
	return base
}

func endsWith(s, suffix string) bool {
	if len(s) < len(suffix) {
		return false
	}
	return s[len(s)-len(suffix):] == suffix
}
