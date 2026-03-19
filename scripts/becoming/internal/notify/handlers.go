package notify

import (
	"encoding/json"
	"net/http"

	"github.com/cpuchip/scripture-study/scripts/becoming/internal/auth"
	"github.com/cpuchip/scripture-study/scripts/becoming/internal/db"
	"github.com/go-chi/chi/v5"
)

// Router creates the push notification API routes.
func Router(database *db.DB, scheduler *Scheduler) chi.Router {
	r := chi.NewRouter()

	r.Get("/vapid-key", handleVAPIDKey(scheduler))
	r.Post("/subscribe", handleSubscribe(database))
	r.Delete("/unsubscribe", handleUnsubscribe(database))
	r.Post("/test", handleTest(scheduler))
	r.Get("/settings", handleGetSettings(database))
	r.Put("/settings", handleUpdateSettings(database))

	return r
}

func handleVAPIDKey(scheduler *Scheduler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"public_key": scheduler.VAPIDPublicKey(),
		})
	}
}

type subscribeRequest struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256DH string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

func handleSubscribe(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)

		var req subscribeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if req.Endpoint == "" || req.Keys.P256DH == "" || req.Keys.Auth == "" {
			writeError(w, http.StatusBadRequest, "endpoint, keys.p256dh, and keys.auth are required")
			return
		}

		userAgent := r.Header.Get("User-Agent")
		if err := database.SavePushSubscription(userID, req.Endpoint, req.Keys.P256DH, req.Keys.Auth, userAgent); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save subscription")
			return
		}

		// Auto-enable notifications when first subscription is created
		if err := database.SetUserNotificationsEnabled(userID, true); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to update settings")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "subscribed"})
	}
}

func handleUnsubscribe(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)

		var req struct {
			Endpoint string `json:"endpoint"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if req.Endpoint == "" {
			writeError(w, http.StatusBadRequest, "endpoint is required")
			return
		}

		if err := database.DeletePushSubscription(userID, req.Endpoint); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to remove subscription")
			return
		}

		// Check if user has any remaining subscriptions
		subs, _ := database.GetPushSubscriptions(userID)
		if len(subs) == 0 {
			// No more subscriptions — disable notifications
			database.SetUserNotificationsEnabled(userID, false)
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func handleTest(scheduler *Scheduler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		if err := scheduler.SendTestNotification(userID); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "sent"})
	}
}

func handleGetSettings(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)
		settings, err := database.GetUserSettings(userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load settings")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(settings)
	}
}

func handleUpdateSettings(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r)

		var req struct {
			NotificationsEnabled bool `json:"notifications_enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		if err := database.SetUserNotificationsEnabled(userID, req.NotificationsEnabled); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to update settings")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"notifications_enabled": req.NotificationsEnabled})
	}
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
