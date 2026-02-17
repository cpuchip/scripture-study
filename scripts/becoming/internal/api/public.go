package api

import (
	"encoding/json"
	"net/http"

	"github.com/cpuchip/scripture-study/scripts/becoming/internal/auth"
	"github.com/cpuchip/scripture-study/scripts/becoming/internal/db"
	"github.com/go-chi/chi/v5"
)

// PublicRouter creates routes for unauthenticated/public access.
// These are mounted outside the auth.Required middleware.
func PublicRouter(database *db.DB) chi.Router {
	r := chi.NewRouter()

	r.Get("/share/{code}", resolveSharedLink(database))
	r.Post("/share", createSharedLink(database))

	return r
}

// resolveSharedLink looks up a short code and returns the full reader parameters.
func resolveSharedLink(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := chi.URLParam(r, "code")
		if code == "" {
			writeError(w, http.StatusBadRequest, "missing code")
			return
		}

		link, err := database.ResolveSharedLink(code)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to resolve link")
			return
		}
		if link == nil {
			writeError(w, http.StatusNotFound, "link not found")
			return
		}

		writeJSON(w, http.StatusOK, link)
	}
}

// createSharedLink creates a new short link. Auth is optional — logged-in users
// get their user_id recorded; anonymous users can still create links.
func createSharedLink(database *db.DB) http.HandlerFunc {
	type createReq struct {
		Repo      string  `json:"repo"`
		Branch    string  `json:"branch"`
		DocFilter string  `json:"doc_filter"`
		FilePath  *string `json:"file_path"`
		SourceID  *int64  `json:"source_id"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req createReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		if req.Repo == "" {
			writeError(w, http.StatusBadRequest, "repo is required")
			return
		}

		// Check for existing link with same params
		fp := ""
		if req.FilePath != nil {
			fp = *req.FilePath
		}
		existing, err := database.GetSharedLinkByParams(req.Repo, branchOrDefault(req.Branch), filterOrDefault(req.DocFilter), fp)
		if err == nil && existing != nil {
			writeJSON(w, http.StatusOK, existing)
			return
		}

		link := &db.SharedLink{
			Provider:  "gh",
			Repo:      req.Repo,
			Branch:    branchOrDefault(req.Branch),
			DocFilter: filterOrDefault(req.DocFilter),
			FilePath:  req.FilePath,
			SourceID:  req.SourceID,
		}

		// Optional auth: if user is logged in, record their user_id
		userID := auth.UserID(r)
		if userID > 0 {
			link.UserID = &userID
		}

		if err := database.CreateSharedLink(link); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create link")
			return
		}

		writeJSON(w, http.StatusCreated, link)
	}
}

func branchOrDefault(b string) string {
	if b == "" {
		return "main"
	}
	return b
}

func filterOrDefault(f string) string {
	if f == "" {
		return "**/*.md"
	}
	return f
}
