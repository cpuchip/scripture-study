// projects endpoints — list, get, create, update, archive.
// Batch I.1: formalizes work_items.project_association.

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"time"
)

func (d *Deps) registerProjects(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/projects/list", d.projectsListHandler)
	mux.HandleFunc("GET /api/projects/get", d.projectsGetHandler)
	mux.HandleFunc("POST /api/projects/create", d.projectsCreateHandler)
	mux.HandleFunc("POST /api/projects/update", d.projectsUpdateHandler)
	mux.HandleFunc("POST /api/projects/archive", d.projectsArchiveHandler)
}

type projectRow struct {
	Slug          string     `json:"slug"`
	Name          string     `json:"name"`
	Description   string     `json:"description,omitempty"`
	RootDirectory string     `json:"root_directory,omitempty"`
	Archived      bool       `json:"archived"`
	WorkItemCount int        `json:"work_item_count"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
}

type projectsListResp struct {
	Items []projectRow `json:"items"`
	Total int          `json:"total"`
}

// slugRegex matches the same shape work_items.slug uses.
var projectSlugRegex = regexp.MustCompile(`^[a-z0-9-]+$`)

func (d *Deps) projectsListHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	includeArchived := r.URL.Query().Get("include_archived") == "true"

	whereSQL := "WHERE NOT archived"
	if includeArchived {
		whereSQL = ""
	}

	resp := projectsListResp{Items: []projectRow{}}
	rows, err := d.Pool.Query(ctx, `
		SELECT p.slug, p.name,
		       coalesce(p.description, ''),
		       coalesce(p.root_directory, ''),
		       p.archived,
		       (SELECT count(*)::int FROM stewards.work_items wi
		         WHERE wi.project_association = p.slug),
		       p.created_at, p.updated_at
		  FROM stewards.projects p
		  `+whereSQL+`
		 ORDER BY p.archived ASC, p.slug ASC`)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		var p projectRow
		if err := rows.Scan(&p.Slug, &p.Name, &p.Description, &p.RootDirectory,
			&p.Archived, &p.WorkItemCount, &p.CreatedAt, &p.UpdatedAt); err == nil {
			resp.Items = append(resp.Items, p)
		}
	}
	resp.Total = len(resp.Items)
	writeJSON(w, http.StatusOK, resp)
}

func (d *Deps) projectsGetHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	slug := r.URL.Query().Get("slug")
	if slug == "" {
		writeErr(w, http.StatusBadRequest, "slug query param required")
		return
	}

	var p projectRow
	err := d.Pool.QueryRow(ctx, `
		SELECT p.slug, p.name,
		       coalesce(p.description, ''),
		       coalesce(p.root_directory, ''),
		       p.archived,
		       (SELECT count(*)::int FROM stewards.work_items wi
		         WHERE wi.project_association = p.slug),
		       p.created_at, p.updated_at
		  FROM stewards.projects p
		 WHERE p.slug = $1`, slug).Scan(
		&p.Slug, &p.Name, &p.Description, &p.RootDirectory,
		&p.Archived, &p.WorkItemCount, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		writeErr(w, http.StatusNotFound, "project not found: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, p)
}

type projectCreateReq struct {
	Slug          string `json:"slug"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	RootDirectory string `json:"root_directory,omitempty"`
}

func (d *Deps) projectsCreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req projectCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "decode body: "+err.Error())
		return
	}
	if req.Slug == "" || req.Name == "" {
		writeErr(w, http.StatusBadRequest, "slug and name required")
		return
	}
	if !projectSlugRegex.MatchString(req.Slug) {
		writeErr(w, http.StatusBadRequest, "slug must match ^[a-z0-9-]+$")
		return
	}

	// Empty-string → NULL for nullable columns.
	var desc, root any
	if req.Description != "" {
		desc = req.Description
	}
	if req.RootDirectory != "" {
		root = req.RootDirectory
	}

	_, err := d.Pool.Exec(ctx, `
		INSERT INTO stewards.projects (slug, name, description, root_directory)
		VALUES ($1, $2, $3, $4)`,
		req.Slug, req.Name, desc, root)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "create: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"slug":    req.Slug,
		"message": "created",
	})
}

type projectUpdateReq struct {
	Slug          string  `json:"slug"`
	Name          *string `json:"name,omitempty"`
	Description   *string `json:"description,omitempty"`
	RootDirectory *string `json:"root_directory,omitempty"`
}

// projectsUpdateHandler edits name/description/root_directory on an
// existing project. Each field is optional; only the ones present in
// the request body get updated. Slug is the immutable identifier.
func (d *Deps) projectsUpdateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req projectUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "decode body: "+err.Error())
		return
	}
	if req.Slug == "" {
		writeErr(w, http.StatusBadRequest, "slug required")
		return
	}

	tx, err := d.Pool.Begin(ctx)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback(ctx)

	if req.Name != nil && *req.Name != "" {
		if _, err := tx.Exec(ctx,
			`UPDATE stewards.projects SET name=$2, updated_at=now() WHERE slug=$1`,
			req.Slug, *req.Name); err != nil {
			writeErr(w, http.StatusBadRequest, "update name: "+err.Error())
			return
		}
	}
	if req.Description != nil {
		var v any
		if *req.Description == "" {
			v = nil
		} else {
			v = *req.Description
		}
		if _, err := tx.Exec(ctx,
			`UPDATE stewards.projects SET description=$2, updated_at=now() WHERE slug=$1`,
			req.Slug, v); err != nil {
			writeErr(w, http.StatusBadRequest, "update description: "+err.Error())
			return
		}
	}
	if req.RootDirectory != nil {
		var v any
		if *req.RootDirectory == "" {
			v = nil
		} else {
			v = *req.RootDirectory
		}
		if _, err := tx.Exec(ctx,
			`UPDATE stewards.projects SET root_directory=$2, updated_at=now() WHERE slug=$1`,
			req.Slug, v); err != nil {
			writeErr(w, http.StatusBadRequest, "update root_directory: "+err.Error())
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		writeErr(w, http.StatusInternalServerError, "commit: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"slug":    req.Slug,
		"message": "updated",
	})
}

type projectArchiveReq struct {
	Slug     string `json:"slug"`
	Archived bool   `json:"archived"` // toggle: true to archive, false to unarchive
}

func (d *Deps) projectsArchiveHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req projectArchiveReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "decode body: "+err.Error())
		return
	}
	if req.Slug == "" {
		writeErr(w, http.StatusBadRequest, "slug required")
		return
	}

	_, err := d.Pool.Exec(ctx,
		`UPDATE stewards.projects SET archived=$2, updated_at=now() WHERE slug=$1`,
		req.Slug, req.Archived)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "archive: "+err.Error())
		return
	}
	msg := "unarchived"
	if req.Archived {
		msg = "archived"
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"slug":     req.Slug,
		"archived": req.Archived,
		"message":  msg,
	})
}
