// Package api provides HTTP handlers for the Becoming app.
package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cpuchip/scripture-study/scripts/becoming/internal/db"
	"github.com/go-chi/chi/v5"
)

// Router creates the API router with all routes.
func Router(database *db.DB) chi.Router {
	r := chi.NewRouter()

	// Practices
	r.Route("/practices", func(r chi.Router) {
		r.Get("/", listPractices(database))
		r.Post("/", createPractice(database))
		r.Get("/{id}", getPractice(database))
		r.Put("/{id}", updatePractice(database))
		r.Delete("/{id}", deletePractice(database))
		r.Get("/{id}/logs", listPracticeLogs(database))
	})

	// Practice logs
	r.Route("/logs", func(r chi.Router) {
		r.Post("/", createLog(database))
		r.Delete("/{id}", deleteLog(database))
	})

	// Daily summary
	r.Get("/daily/{date}", getDailySummary(database))

	// Tasks
	r.Route("/tasks", func(r chi.Router) {
		r.Get("/", listTasks(database))
		r.Post("/", createTask(database))
		r.Put("/{id}", updateTask(database))
		r.Delete("/{id}", deleteTask(database))
	})

	return r
}

// --- Practices ---

func listPractices(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pType := r.URL.Query().Get("type")
		activeOnly := r.URL.Query().Get("active") != "false"

		practices, err := database.ListPractices(pType, activeOnly)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if practices == nil {
			practices = []*db.Practice{}
		}
		writeJSON(w, http.StatusOK, practices)
	}
}

func createPractice(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p db.Practice
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		if p.Name == "" || p.Type == "" {
			writeError(w, http.StatusBadRequest, "name and type are required")
			return
		}
		if p.Config == "" {
			p.Config = "{}"
		}
		p.Active = true

		if err := database.CreatePractice(&p); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, p)
	}
}

func getPractice(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		p, err := database.GetPractice(id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if p == nil {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeJSON(w, http.StatusOK, p)
	}
}

func updatePractice(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		var p db.Practice
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		p.ID = id
		if err := database.UpdatePractice(&p); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, p)
	}
}

func deletePractice(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		if err := database.DeletePractice(id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func listPracticeLogs(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		limit := 100
		if l := r.URL.Query().Get("limit"); l != "" {
			if n, err := strconv.Atoi(l); err == nil {
				limit = n
			}
		}

		// Check for date range
		startDate := r.URL.Query().Get("start")
		endDate := r.URL.Query().Get("end")
		var logs []*db.PracticeLog
		if startDate != "" && endDate != "" {
			logs, err = database.ListLogsByPracticeRange(id, startDate, endDate)
		} else {
			logs, err = database.ListLogsByPractice(id, limit)
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if logs == nil {
			logs = []*db.PracticeLog{}
		}
		writeJSON(w, http.StatusOK, logs)
	}
}

// --- Logs ---

func createLog(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var l db.PracticeLog
		if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		if l.PracticeID == 0 {
			writeError(w, http.StatusBadRequest, "practice_id is required")
			return
		}
		if err := database.CreateLog(&l); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, l)
	}
}

func deleteLog(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		if err := database.DeleteLog(id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// --- Daily ---

func getDailySummary(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		date := chi.URLParam(r, "date")
		if date == "" {
			writeError(w, http.StatusBadRequest, "date is required (YYYY-MM-DD)")
			return
		}
		summaries, err := database.GetDailySummary(date)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if summaries == nil {
			summaries = []*db.DailySummary{}
		}
		writeJSON(w, http.StatusOK, summaries)
	}
}

// --- Tasks ---

func listTasks(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		tasks, err := database.ListTasks(status)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if tasks == nil {
			tasks = []*db.Task{}
		}
		writeJSON(w, http.StatusOK, tasks)
	}
}

func createTask(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var t db.Task
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		if t.Title == "" {
			writeError(w, http.StatusBadRequest, "title is required")
			return
		}
		if t.Type == "" {
			t.Type = "ongoing"
		}
		if t.Status == "" {
			t.Status = "active"
		}
		if err := database.CreateTask(&t); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, t)
	}
}

func updateTask(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		var t db.Task
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		t.ID = id
		if err := database.UpdateTask(&t); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, t)
	}
}

func deleteTask(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		if err := database.DeleteTask(id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// --- Helpers ---

func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
