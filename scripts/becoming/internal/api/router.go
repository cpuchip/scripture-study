// Package api provides HTTP handlers for the Becoming app.
package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cpuchip/scripture-study/scripts/becoming/internal/auth"
	"github.com/cpuchip/scripture-study/scripts/becoming/internal/db"
	"github.com/cpuchip/scripture-study/scripts/becoming/internal/scripture"
	"github.com/go-chi/chi/v5"
)

// Router creates the API router with all routes.
func Router(database *db.DB, scripturesRoot string) chi.Router {
	r := chi.NewRouter()

	// Practices
	r.Route("/practices", func(r chi.Router) {
		r.Get("/", listPractices(database))
		r.Post("/", createPractice(database))
		r.Get("/{id}", getPractice(database))
		r.Put("/{id}", updatePractice(database))
		r.Delete("/{id}", deletePractice(database))
		r.Get("/{id}/logs", listPracticeLogs(database))
		r.Post("/{id}/complete", completePractice(database))
		r.Post("/{id}/archive", archivePractice(database))
		r.Post("/{id}/pause", pausePractice(database))
		r.Post("/{id}/restore", restorePractice(database))
	})

	// Practice logs
	r.Route("/logs", func(r chi.Router) {
		r.Post("/", createLog(database))
		r.Delete("/{id}", deleteLog(database))
		r.Delete("/latest", deleteLatestLog(database))
	})

	// Daily summary
	r.Get("/daily/{date}", getDailySummary(database))

	// Reports
	r.Get("/reports", getReport(database))

	// Tasks
	r.Route("/tasks", func(r chi.Router) {
		r.Get("/", listTasks(database))
		r.Post("/", createTask(database))
		r.Put("/{id}", updateTask(database))
		r.Delete("/{id}", deleteTask(database))
	})

	// Memorization / spaced repetition
	r.Route("/memorize", func(r chi.Router) {
		r.Get("/due/{date}", getDueCards(database))
		r.Get("/cards/{date}", getMemorizeCards(database))
		r.Post("/review", reviewCard(database))

		// Study mode (adaptive difficulty)
		r.Get("/study/next", studyNext(database))
		r.Post("/study/score", studyScore(database))
		r.Get("/study/aptitudes/{practiceId}", studyAptitudes(database))
		r.Post("/study/seed", studySeed(database))
	})

	// Scripture lookup
	r.Route("/scriptures", func(r chi.Router) {
		r.Get("/lookup", lookupScripture(scripturesRoot))
		r.Get("/books", listScriptureBooks())
		r.Get("/search", searchScriptureBooks())
	})

	// Notes
	r.Route("/notes", func(r chi.Router) {
		r.Get("/", listNotes(database))
		r.Post("/", createNote(database))
		r.Put("/{id}", updateNote(database))
		r.Delete("/{id}", deleteNote(database))
	})

	// Prompts
	r.Route("/prompts", func(r chi.Router) {
		r.Get("/", listPrompts(database))
		r.Get("/today", getTodayPrompt(database))
		r.Post("/", createPrompt(database))
		r.Put("/{id}", updatePrompt(database))
		r.Delete("/{id}", deletePrompt(database))
	})

	// Reflections
	r.Route("/reflections", func(r chi.Router) {
		r.Get("/", listReflections(database))
		r.Get("/{date}", getReflection(database))
		r.Post("/", upsertReflection(database))
		r.Delete("/{id}", deleteReflection(database))
	})

	// Pillars
	r.Route("/pillars", func(r chi.Router) {
		r.Get("/", listPillarsTree(database))
		r.Get("/flat", listPillarsFlat(database))
		r.Get("/suggestions", getPillarSuggestions())
		r.Get("/has-pillars", hasPillars(database))
		r.Post("/", createPillar(database))
		r.Get("/{id}", getPillar(database))
		r.Put("/{id}", updatePillar(database))
		r.Delete("/{id}", deletePillar(database))
		r.Put("/{id}/practices", setPracticePillars(database))
		r.Get("/{id}/practices", getPracticePillars(database))
	})

	// Practice ↔ Pillar links
	r.Put("/practices/{id}/pillars", setPracticePillarsForPractice(database))
	r.Get("/practices/{id}/pillars", getPracticePillarsForPractice(database))

	return r
}

// --- Practices ---

func listPractices(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		pType := r.URL.Query().Get("type")
		status := r.URL.Query().Get("status")

		// Legacy compat: ?active=false shows all; otherwise default to active only
		activeOnly := r.URL.Query().Get("active") != "false"

		practices, err := database.ListPracticesByStatus(userID, pType, status, activeOnly)
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
		userID := auth.UserID(r)
		var p db.Practice
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		if p.Name == "" || p.Type == "" {
			writeError(w, http.StatusBadRequest, "name and type are required")
			return
		}
		if p.Config == "" || p.Config == "{}" {
			if p.Type == "memorize" {
				cfg := db.DefaultSM2Config()
				cfgJSON, _ := json.Marshal(cfg)
				p.Config = string(cfgJSON)
			} else if p.Type == "tracker" {
				p.Config = `{"target_sets":2,"target_reps":15,"unit":"reps"}`
			} else if p.Type == "scheduled" {
				writeError(w, http.StatusBadRequest, "schedule config is required for scheduled type")
				return
			} else {
				p.Config = "{}"
			}
		} else if p.Type == "memorize" {
			// Merge user-provided config (e.g., target_daily_reps) with SM-2 defaults
			var userCfg map[string]any
			if err := json.Unmarshal([]byte(p.Config), &userCfg); err == nil {
				// Check if this has SM-2 state already
				if _, ok := userCfg["ease_factor"]; !ok {
					cfg := db.DefaultSM2Config()
					if v, ok := userCfg["target_daily_reps"]; ok {
						if n, ok := v.(float64); ok {
							cfg.TargetDailyReps = int(n)
						}
					}
					cfgJSON, _ := json.Marshal(cfg)
					p.Config = string(cfgJSON)
				}
			}
		} else if p.Type == "scheduled" {
			// Validate that the config contains a valid schedule.
			if _, err := db.ParseScheduleConfig(p.Config); err != nil {
				writeError(w, http.StatusBadRequest, "invalid schedule config: "+err.Error())
				return
			}
		}
		p.Active = true

		if err := database.CreatePractice(userID, &p); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, p)
	}
}

func getPractice(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		id, err := parseID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		p, err := database.GetPractice(userID, id)
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
		userID := auth.UserID(r)
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
		if err := database.UpdatePractice(userID, &p); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, p)
	}
}

func deletePractice(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		id, err := parseID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		if err := database.DeletePractice(userID, id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// --- Practice Lifecycle Actions ---

func completePractice(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		id, err := parseID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		if err := database.CompletePractice(userID, id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		p, _ := database.GetPractice(userID, id)
		writeJSON(w, http.StatusOK, p)
	}
}

func archivePractice(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		id, err := parseID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		if err := database.ArchivePractice(userID, id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		p, _ := database.GetPractice(userID, id)
		writeJSON(w, http.StatusOK, p)
	}
}

func pausePractice(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		id, err := parseID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		if err := database.PausePractice(userID, id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		p, _ := database.GetPractice(userID, id)
		writeJSON(w, http.StatusOK, p)
	}
}

func restorePractice(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		id, err := parseID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		if err := database.RestorePractice(userID, id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		p, _ := database.GetPractice(userID, id)
		writeJSON(w, http.StatusOK, p)
	}
}

func listPracticeLogs(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
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
			logs, err = database.ListLogsByPracticeRange(userID, id, startDate, endDate)
		} else {
			logs, err = database.ListLogsByPractice(userID, id, limit)
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
		userID := auth.UserID(r)
		var l db.PracticeLog
		if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		if l.PracticeID == 0 {
			writeError(w, http.StatusBadRequest, "practice_id is required")
			return
		}
		if err := database.CreateLog(userID, &l); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, l)
	}
}

func deleteLog(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		id, err := parseID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		if err := database.DeleteLog(userID, id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func deleteLatestLog(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		practiceID := r.URL.Query().Get("practice_id")
		date := r.URL.Query().Get("date")
		if practiceID == "" || date == "" {
			writeError(w, http.StatusBadRequest, "practice_id and date are required")
			return
		}
		pid, err := strconv.ParseInt(practiceID, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid practice_id")
			return
		}
		ok, err := database.DeleteLatestLog(userID, pid, date)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !ok {
			writeError(w, http.StatusNotFound, "no log found for that practice and date")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// --- Daily ---

func getDailySummary(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		date := chi.URLParam(r, "date")
		if date == "" {
			writeError(w, http.StatusBadRequest, "date is required (YYYY-MM-DD)")
			return
		}
		summaries, err := database.GetDailySummary(userID, date)
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

// --- Reports ---

func getReport(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		start := r.URL.Query().Get("start")
		end := r.URL.Query().Get("end")
		if start == "" || end == "" {
			writeError(w, http.StatusBadRequest, "start and end query params are required (YYYY-MM-DD)")
			return
		}
		entries, err := database.GetReport(userID, start, end)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if entries == nil {
			entries = []*db.ReportEntry{}
		}
		writeJSON(w, http.StatusOK, entries)
	}
}

// --- Tasks ---

func listTasks(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		status := r.URL.Query().Get("status")
		tasks, err := database.ListTasks(userID, status)
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
		userID := auth.UserID(r)
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
		if err := database.CreateTask(userID, &t); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, t)
	}
}

func updateTask(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
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
		if err := database.UpdateTask(userID, &t); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, t)
	}
}

func deleteTask(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		id, err := parseID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		if err := database.DeleteTask(userID, id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// --- Memorize ---

func getDueCards(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		date := chi.URLParam(r, "date")
		if date == "" {
			writeError(w, http.StatusBadRequest, "date is required (YYYY-MM-DD)")
			return
		}
		cards, err := database.GetDueCards(userID, date)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if cards == nil {
			cards = []*db.Practice{}
		}
		writeJSON(w, http.StatusOK, cards)
	}
}

func getMemorizeCards(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		date := chi.URLParam(r, "date")
		if date == "" {
			writeError(w, http.StatusBadRequest, "date is required (YYYY-MM-DD)")
			return
		}
		cards, err := database.GetMemorizeCardStatuses(userID, date)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if cards == nil {
			cards = []*db.MemorizeCardStatus{}
		}
		writeJSON(w, http.StatusOK, cards)
	}
}

func reviewCard(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		var req struct {
			PracticeID int64  `json:"practice_id"`
			Quality    int    `json:"quality"`
			Date       string `json:"date"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		if req.PracticeID == 0 {
			writeError(w, http.StatusBadRequest, "practice_id is required")
			return
		}
		if req.Quality < 0 || req.Quality > 5 {
			writeError(w, http.StatusBadRequest, "quality must be 0-5")
			return
		}
		if req.Date == "" {
			req.Date = r.URL.Query().Get("date")
		}

		p, err := database.ReviewCard(userID, req.PracticeID, req.Quality, req.Date)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, p)
	}
}

// --- Study Mode (Adaptive Difficulty) ---

func studyNext(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		date := r.URL.Query().Get("date")
		if date == "" {
			date = time.Now().Format("2006-01-02")
		}
		category := r.URL.Query().Get("category")
		pillarIDsStr := r.URL.Query().Get("pillar_ids")
		lastCardIDStr := r.URL.Query().Get("last_card_id")
		var lastCardID int64
		if lastCardIDStr != "" {
			lastCardID, _ = strconv.ParseInt(lastCardIDStr, 10, 64)
		}

		// Parse session state from query params
		momentumStr := r.URL.Query().Get("momentum")
		recentScoresStr := r.URL.Query().Get("recent_scores")
		session := db.NewStudySession()
		if momentumStr != "" {
			session.Momentum = db.SessionMomentum(momentumStr)
		}
		if recentScoresStr != "" {
			for _, s := range splitComma(recentScoresStr) {
				if v, err := strconv.ParseFloat(s, 64); err == nil {
					session.RecentScores = append(session.RecentScores, v)
				}
			}
		}

		// Get due cards (or all active memorize cards for "keep studying")
		mode := r.URL.Query().Get("mode") // "due" (default) or "all"
		var cards []*db.Practice
		var err error
		if mode == "all" {
			cards, err = database.GetAllMemorizeCards(userID)
		} else {
			cards, err = database.GetDueCards(userID, date)
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Filter by category if specified (supports comma-separated)
		if category != "" {
			cats := splitComma(category)
			catSet := make(map[string]bool, len(cats))
			for _, c := range cats {
				catSet[c] = true
			}
			var filtered []*db.Practice
			for _, c := range cards {
				if catSet[c.Category] {
					filtered = append(filtered, c)
				}
			}
			cards = filtered
		}

		// Filter by pillar IDs if specified (comma-separated)
		if pillarIDsStr != "" {
			pillarIDs := make(map[int64]bool)
			for _, s := range splitComma(pillarIDsStr) {
				if id, err := strconv.ParseInt(s, 10, 64); err == nil {
					pillarIDs[id] = true
				}
			}
			if len(pillarIDs) > 0 {
				var filtered []*db.Practice
				for _, c := range cards {
					links, _ := database.GetPracticePillars(userID, c.ID)
					for _, link := range links {
						if pillarIDs[link.PillarID] {
							filtered = append(filtered, c)
							break
						}
					}
				}
				cards = filtered
			}
		}

		if len(cards) == 0 {
			writeJSON(w, http.StatusOK, map[string]any{
				"done":    true,
				"message": "No cards available for study",
			})
			return
		}

		// Get aptitudes for all cards
		aptitudeMap, err := database.GetAllUserAptitudes(userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Select next exercise
		exercise := db.SelectNextExercise(cards, aptitudeMap, session, lastCardID)
		if exercise == nil {
			writeJSON(w, http.StatusOK, map[string]any{
				"done":    true,
				"message": "No exercise available",
			})
			return
		}

		// Include all card names so reverse mode can use real references as distractors
		allNames := make([]string, 0, len(cards))
		for _, c := range cards {
			allNames = append(allNames, c.Name)
		}
		exercise.AllCardNames = allNames

		writeJSON(w, http.StatusOK, exercise)
	}
}

func studyScore(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		var req struct {
			PracticeID int64   `json:"practice_id"`
			Mode       string  `json:"mode"`
			Score      float64 `json:"score"`
			Quality    *int    `json:"quality"`
			DurationS  *int    `json:"duration_s"`
			Date       string  `json:"date"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		if req.PracticeID == 0 {
			writeError(w, http.StatusBadRequest, "practice_id is required")
			return
		}
		if req.Mode == "" {
			writeError(w, http.StatusBadRequest, "mode is required")
			return
		}
		if req.Score < 0 || req.Score > 1 {
			writeError(w, http.StatusBadRequest, "score must be 0.0-1.0")
			return
		}
		if req.Date == "" {
			req.Date = time.Now().Format("2006-01-02")
		}

		score := &db.MemorizeScore{
			PracticeID: req.PracticeID,
			UserID:     userID,
			Mode:       req.Mode,
			Score:      req.Score,
			Quality:    req.Quality,
			DurationS:  req.DurationS,
			Date:       req.Date,
		}

		if err := database.RecordScore(userID, score); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Also update memorize_level based on new aptitude
		if err := database.UpdateMemorizeLevel(userID, req.PracticeID); err != nil {
			// Non-fatal — log but don't fail the request
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Always log a practice_log entry so this counts toward daily reps.
		// If quality is provided (level 3+), use ReviewCard which also advances SM-2.
		// If quality is nil (level 1-2 exposure), log a basic entry with a default quality
		// so it counts as a rep but doesn't aggressively advance SM-2 scheduling.
		if req.Quality != nil {
			if _, err := database.ReviewCard(userID, req.PracticeID, *req.Quality, req.Date); err != nil {
				// Non-fatal — the score was already recorded
				// Log warning but return success
			}
		} else {
			// Level 1-2 exposure: log with quality 3 ("correct with difficulty")
			// to count as a daily rep without heavily advancing SM-2 interval
			defaultQuality := 3
			if _, err := database.ReviewCard(userID, req.PracticeID, defaultQuality, req.Date); err != nil {
				// Non-fatal
			}
		}

		// Return updated aptitudes for this card
		aptitudes, err := database.GetAptitudes(userID, req.PracticeID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"score":     score,
			"aptitudes": aptitudes,
			"overall":   db.OverallAptitude(aptitudes),
		})
	}
}

func studyAptitudes(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		practiceID, err := strconv.ParseInt(chi.URLParam(r, "practiceId"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid practice ID")
			return
		}

		aptitudes, err := database.GetAptitudes(userID, practiceID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if aptitudes == nil {
			aptitudes = []*db.MemorizeAptitude{}
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"aptitudes": aptitudes,
			"overall":   db.OverallAptitude(aptitudes),
		})
	}
}

func studySeed(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		if err := database.SeedAptitudesFromSM2(userID); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

// splitComma splits a comma-separated string into parts.
func splitComma(s string) []string {
	if s == "" {
		return nil
	}
	var parts []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

// --- Helpers ---

func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
}

// --- Notes ---

func listNotes(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		var practiceID, taskID, pillarID *int64
		if v := r.URL.Query().Get("practice_id"); v != "" {
			id, err := strconv.ParseInt(v, 10, 64)
			if err == nil {
				practiceID = &id
			}
		}
		if v := r.URL.Query().Get("task_id"); v != "" {
			id, err := strconv.ParseInt(v, 10, 64)
			if err == nil {
				taskID = &id
			}
		}
		if v := r.URL.Query().Get("pillar_id"); v != "" {
			id, err := strconv.ParseInt(v, 10, 64)
			if err == nil {
				pillarID = &id
			}
		}
		pinnedOnly := r.URL.Query().Get("pinned") == "true"

		notes, err := database.ListNotes(userID, practiceID, taskID, pillarID, pinnedOnly)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if notes == nil {
			notes = []*db.Note{}
		}
		writeJSON(w, http.StatusOK, notes)
	}
}

func createNote(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		var n db.Note
		if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		if n.Content == "" {
			writeError(w, http.StatusBadRequest, "content is required")
			return
		}
		if err := database.CreateNote(userID, &n); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, n)
	}
}

func updateNote(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		id, err := parseID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		var n db.Note
		if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		n.ID = id
		if err := database.UpdateNote(userID, &n); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, n)
	}
}

func deleteNote(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		id, err := parseID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		if err := database.DeleteNote(userID, id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// --- Prompts ---

func listPrompts(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		activeOnly := r.URL.Query().Get("active") != "false"
		prompts, err := database.ListPrompts(userID, activeOnly)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if prompts == nil {
			prompts = []*db.Prompt{}
		}
		writeJSON(w, http.StatusOK, prompts)
	}
}

func getTodayPrompt(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		dayOfYear := time.Now().YearDay()
		prompt, err := database.GetTodayPrompt(userID, dayOfYear)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if prompt == nil {
			writeJSON(w, http.StatusOK, map[string]any{"text": ""})
			return
		}
		writeJSON(w, http.StatusOK, prompt)
	}
}

func createPrompt(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		var p db.Prompt
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		if p.Text == "" {
			writeError(w, http.StatusBadRequest, "text is required")
			return
		}
		p.Active = true
		if err := database.CreatePrompt(userID, &p); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, p)
	}
}

func updatePrompt(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		id, err := parseID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		var p db.Prompt
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		p.ID = id
		if err := database.UpdatePrompt(userID, &p); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, p)
	}
}

func deletePrompt(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		id, err := parseID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		if err := database.DeletePrompt(userID, id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// --- Reflections ---

func listReflections(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")
		reflections, err := database.ListReflections(userID, from, to)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if reflections == nil {
			reflections = []*db.Reflection{}
		}
		writeJSON(w, http.StatusOK, reflections)
	}
}

func getReflection(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		date := chi.URLParam(r, "date")
		ref, err := database.GetReflection(userID, date)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if ref == nil {
			writeJSON(w, http.StatusOK, nil)
			return
		}
		writeJSON(w, http.StatusOK, ref)
	}
}

func upsertReflection(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		var ref db.Reflection
		if err := json.NewDecoder(r.Body).Decode(&ref); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		if ref.Date == "" || ref.Content == "" {
			writeError(w, http.StatusBadRequest, "date and content are required")
			return
		}
		if ref.Mood != nil && (*ref.Mood < 1 || *ref.Mood > 5) {
			writeError(w, http.StatusBadRequest, "mood must be 1-5")
			return
		}
		if err := database.UpsertReflection(userID, &ref); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, ref)
	}
}

func deleteReflection(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		id, err := parseID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		if err := database.DeleteReflection(userID, id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// --- Pillars ---

func listPillarsTree(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		pillars, err := database.ListPillarsTree(userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if pillars == nil {
			pillars = []*db.Pillar{}
		}
		writeJSON(w, http.StatusOK, pillars)
	}
}

func listPillarsFlat(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		pillars, err := database.ListPillars(userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if pillars == nil {
			pillars = []*db.Pillar{}
		}
		writeJSON(w, http.StatusOK, pillars)
	}
}

func getPillarSuggestions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, db.GetDefaultPillarSuggestions())
	}
}

func hasPillars(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		has, err := database.HasPillars(userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"has_pillars": has})
	}
}

func createPillar(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		var p db.Pillar
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		if p.Name == "" {
			writeError(w, http.StatusBadRequest, "name is required")
			return
		}
		if err := database.CreatePillar(userID, &p); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, p)
	}
}

func getPillar(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		id, err := parseID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		p, err := database.GetPillar(userID, id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, p)
	}
}

func updatePillar(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		id, err := parseID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		var p db.Pillar
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		p.ID = id
		if err := database.UpdatePillar(userID, &p); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, p)
	}
}

func deletePillar(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		id, err := parseID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		if err := database.DeletePillar(userID, id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func setPracticePillars(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		id, err := parseID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		var body struct {
			PracticeIDs []int64 `json:"practice_ids"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		// Set all given practice_ids to link to this pillar
		for _, pid := range body.PracticeIDs {
			if err := database.LinkPracticePillar(userID, pid, id); err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func getPracticePillars(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		// This endpoint isn't very useful by itself, stub for future use
		writeJSON(w, http.StatusOK, []any{})
	}
}

func setPracticePillarsForPractice(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		id, err := parseID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		var body struct {
			PillarIDs []int64 `json:"pillar_ids"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if err := database.SetPracticePillars(userID, id, body.PillarIDs); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func getPracticePillarsForPractice(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		id, err := parseID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		links, err := database.GetPracticePillars(userID, id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if links == nil {
			links = []db.PillarLink{}
		}
		writeJSON(w, http.StatusOK, links)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// --- Scriptures ---

func lookupScripture(root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ref := r.URL.Query().Get("ref")
		if ref == "" {
			writeError(w, http.StatusBadRequest, "ref query parameter is required (e.g., ?ref=D%26C+93:29)")
			return
		}

		result, err := scripture.Lookup(root, ref)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, result)
	}
}

func listScriptureBooks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, scripture.ListBooks())
	}
}

func searchScriptureBooks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		if q == "" {
			writeJSON(w, http.StatusOK, []any{})
			return
		}
		writeJSON(w, http.StatusOK, scripture.SearchBooks(q))
	}
}
