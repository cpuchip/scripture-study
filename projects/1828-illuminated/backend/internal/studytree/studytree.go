package studytree

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stuffleberry/i1828/backend/internal/auth"
	"github.com/stuffleberry/i1828/backend/internal/httpx"
)

type Service struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

type SavedTree struct {
	ID             string          `json:"id"`
	BecomingUserID int64           `json:"becoming_user_id"`
	Title          string          `json:"title"`
	TreeData       json.RawMessage `json:"tree_data"`
	UpdatedAt      string          `json:"updated_at"`
}

type SaveRequest struct {
	ID       *string         `json:"id,omitempty"`
	Title    string          `json:"title"`
	TreeData json.RawMessage `json:"tree_data"`
}

func (s *Service) Register(mux *http.ServeMux, authSvc *auth.Auth) {
	mux.Handle("GET /api/study-tree", authSvc.RequireAuth(http.HandlerFunc(s.handleList)))
	mux.Handle("POST /api/study-tree", authSvc.RequireAuth(http.HandlerFunc(s.handleSave)))
	mux.Handle("GET /api/study-tree/{id}", authSvc.RequireAuth(http.HandlerFunc(s.handleGet)))
	mux.Handle("DELETE /api/study-tree/{id}", authSvc.RequireAuth(http.HandlerFunc(s.handleDelete)))
}

func (s *Service) handleList(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "sign in first")
		return
	}

	rows, err := s.pool.Query(r.Context(), `
		SELECT id, becoming_user_id, title, tree_data, updated_at
		FROM saved_trees
		WHERE becoming_user_id = $1
		ORDER BY updated_at DESC
	`, user.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	defer rows.Close()

	var trees []SavedTree
	for rows.Next() {
		var t SavedTree
		var tTime time.Time
		err := rows.Scan(&t.ID, &t.BecomingUserID, &t.Title, &t.TreeData, &tTime)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "db_error", err.Error())
			return
		}
		t.UpdatedAt = tTime.Format(time.RFC3339)
		trees = append(trees, t)
	}

	// pgx returns nil slice when empty, let's render empty array [] instead of null
	if trees == nil {
		trees = []SavedTree{}
	}

	httpx.WriteJSON(w, http.StatusOK, trees)
}

func (s *Service) handleSave(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "sign in first")
		return
	}

	var req SaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	if req.Title == "" {
		httpx.WriteError(w, http.StatusBadRequest, "missing_title", "title is required")
		return
	}

	if req.ID != nil && *req.ID != "" {
		// Update
		_, err := s.pool.Exec(r.Context(), `
			UPDATE saved_trees
			SET title = $1, tree_data = $2, updated_at = now()
			WHERE id = $3 AND becoming_user_id = $4
		`, req.Title, req.TreeData, *req.ID, user.ID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "db_error", err.Error())
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"id": *req.ID, "status": "updated"})
	} else {
		// Insert
		var newID string
		err := s.pool.QueryRow(r.Context(), `
			INSERT INTO saved_trees (becoming_user_id, title, tree_data)
			VALUES ($1, $2, $3)
			RETURNING id
		`, user.ID, req.Title, req.TreeData).Scan(&newID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "db_error", err.Error())
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"id": newID, "status": "created"})
	}
}

func (s *Service) handleGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	user := auth.GetUser(r.Context())
	if user == nil {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "sign in first")
		return
	}

	var t SavedTree
	var tTime time.Time
	err := s.pool.QueryRow(r.Context(), `
		SELECT id, becoming_user_id, title, tree_data, updated_at
		FROM saved_trees
		WHERE id = $1 AND becoming_user_id = $2
	`, id, user.ID).Scan(&t.ID, &t.BecomingUserID, &t.Title, &t.TreeData, &tTime)
	if err == pgx.ErrNoRows {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "tree not found")
		return
	} else if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	t.UpdatedAt = tTime.Format(time.RFC3339)
	httpx.WriteJSON(w, http.StatusOK, t)
}

func (s *Service) handleDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	user := auth.GetUser(r.Context())
	if user == nil {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "sign in first")
		return
	}

	_, err := s.pool.Exec(r.Context(), `
		DELETE FROM saved_trees
		WHERE id = $1 AND becoming_user_id = $2
	`, id, user.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
