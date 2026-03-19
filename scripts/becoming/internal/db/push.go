package db

import (
	"encoding/json"
	"fmt"
	"time"
)

// PushSubscription represents a browser push subscription for a user.
type PushSubscription struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Endpoint  string `json:"endpoint"`
	KeyP256DH string `json:"keys_p256dh"`
	KeyAuth   string `json:"keys_auth"`
	UserAgent string `json:"user_agent,omitempty"`
	CreatedAt string `json:"created_at"`
}

// NotificationLog records that a notification was sent, preventing duplicates.
type NotificationLog struct {
	ID         int64  `json:"id"`
	UserID     int64  `json:"user_id"`
	PracticeID int64  `json:"practice_id"`
	SentAt     string `json:"sent_at"`
	Date       string `json:"date"`
}

// SavePushSubscription stores a push subscription for a user.
// If the endpoint already exists for this user, it updates the keys.
func (db *DB) SavePushSubscription(userID int64, endpoint, keyP256DH, keyAuth, userAgent string) error {
	_, err := db.Exec(`
		INSERT INTO push_subscriptions (user_id, endpoint, keys_p256dh, keys_auth, user_agent)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (user_id, endpoint) DO UPDATE SET keys_p256dh = ?, keys_auth = ?, user_agent = ?`,
		userID, endpoint, keyP256DH, keyAuth, userAgent,
		keyP256DH, keyAuth, userAgent)
	return err
}

// DeletePushSubscription removes a push subscription by endpoint.
func (db *DB) DeletePushSubscription(userID int64, endpoint string) error {
	_, err := db.Exec(`DELETE FROM push_subscriptions WHERE user_id = ? AND endpoint = ?`, userID, endpoint)
	return err
}

// DeletePushSubscriptionByEndpoint removes a subscription by endpoint only (for cleanup on 410 Gone).
func (db *DB) DeletePushSubscriptionByEndpoint(endpoint string) error {
	_, err := db.Exec(`DELETE FROM push_subscriptions WHERE endpoint = ?`, endpoint)
	return err
}

// GetPushSubscriptions returns all push subscriptions for a user.
func (db *DB) GetPushSubscriptions(userID int64) ([]*PushSubscription, error) {
	rows, err := db.Query(`
		SELECT id, user_id, endpoint, keys_p256dh, keys_auth, user_agent, created_at
		FROM push_subscriptions WHERE user_id = ?`, userID)
	if err != nil {
		return nil, fmt.Errorf("listing push subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []*PushSubscription
	for rows.Next() {
		s := &PushSubscription{}
		if err := rows.Scan(&s.ID, &s.UserID, &s.Endpoint, &s.KeyP256DH, &s.KeyAuth, &s.UserAgent, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning push subscription: %w", err)
		}
		subs = append(subs, s)
	}
	return subs, rows.Err()
}

// UsersWithPushSubscriptions returns user IDs that have at least one push subscription.
func (db *DB) UsersWithPushSubscriptions() ([]int64, error) {
	rows, err := db.Query(`SELECT DISTINCT user_id FROM push_subscriptions`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// LogNotification records that a notification was sent for a practice today.
func (db *DB) LogNotification(userID, practiceID int64, date string) error {
	_, err := db.Exec(`
		INSERT INTO notification_log (user_id, practice_id, date)
		VALUES (?, ?, ?)`, userID, practiceID, date)
	return err
}

// NotificationsSentToday returns practice IDs that already received a notification today.
func (db *DB) NotificationsSentToday(userID int64, date string) (map[int64]bool, error) {
	rows, err := db.Query(`
		SELECT practice_id FROM notification_log
		WHERE user_id = ? AND date = ?`, userID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sent := make(map[int64]bool)
	for rows.Next() {
		var pid int64
		if err := rows.Scan(&pid); err != nil {
			return nil, err
		}
		sent[pid] = true
	}
	return sent, rows.Err()
}

// DuePracticesForNotification returns active scheduled practices that are due for a user on a given date.
// It reuses the same scheduling logic as the daily summary, then filters to practices
// with the per-practice notify flag enabled in their config JSON.
func (db *DB) DuePracticesForNotification(userID int64, date string) ([]*DailySummary, error) {
	summaries, err := db.GetDailySummary(userID, date)
	if err != nil {
		return nil, err
	}

	var due []*DailySummary
	for _, s := range summaries {
		// Only include scheduled practices that are actually due and not yet logged
		if s.IsDue != nil && *s.IsDue && s.LogCount == 0 {
			// Check per-practice notify flag in config JSON
			if !practiceNotifyEnabled(s.Config) {
				continue
			}
			due = append(due, s)
		}
	}
	return due, nil
}

// practiceNotifyEnabled checks the "notify" field in a practice's config JSON.
// Returns false if the field is missing or explicitly false.
func practiceNotifyEnabled(configJSON string) bool {
	if configJSON == "" || configJSON == "{}" {
		return false
	}
	var cfg map[string]any
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return false
	}
	notify, ok := cfg["notify"]
	if !ok {
		return false
	}
	b, ok := notify.(bool)
	return ok && b
}

// GetUserNotificationsEnabled checks if a user has notifications enabled.
// Returns false if no setting exists (opt-in by default).
func (db *DB) GetUserNotificationsEnabled(userID int64) bool {
	var enabled bool
	err := db.QueryRow(`
		SELECT notifications_enabled FROM user_settings WHERE user_id = ?`, userID).Scan(&enabled)
	if err != nil {
		return false
	}
	return enabled
}

// SetUserNotificationsEnabled sets the notifications_enabled flag for a user.
func (db *DB) SetUserNotificationsEnabled(userID int64, enabled bool) error {
	_, err := db.Exec(`
		INSERT INTO user_settings (user_id, notifications_enabled)
		VALUES (?, ?)
		ON CONFLICT (user_id) DO UPDATE SET notifications_enabled = ?`,
		userID, enabled, enabled)
	return err
}

// GetUserSettings returns notification settings for a user.
type UserSettings struct {
	NotificationsEnabled    bool    `json:"notifications_enabled"`
	NotifyByDefault         bool    `json:"notify_practices_by_default"`
	QuietHoursStart         *string `json:"quiet_hours_start"`
	QuietHoursEnd           *string `json:"quiet_hours_end"`
	DefaultTiming           string  `json:"default_timing"`
}

func (db *DB) GetUserSettings(userID int64) (*UserSettings, error) {
	s := &UserSettings{DefaultTiming: "at_time"}
	err := db.QueryRow(`
		SELECT notifications_enabled, notify_practices_by_default, quiet_hours_start, quiet_hours_end, default_timing
		FROM user_settings WHERE user_id = ?`, userID).Scan(
		&s.NotificationsEnabled, &s.NotifyByDefault, &s.QuietHoursStart, &s.QuietHoursEnd, &s.DefaultTiming)
	if err != nil {
		// No settings row yet — return defaults
		return &UserSettings{DefaultTiming: "at_time"}, nil
	}
	return s, nil
}

// SaveUserSettings upserts all notification settings for a user.
func (db *DB) SaveUserSettings(userID int64, s *UserSettings) error {
	_, err := db.Exec(`
		INSERT INTO user_settings (user_id, notifications_enabled, notify_practices_by_default, quiet_hours_start, quiet_hours_end, default_timing)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (user_id) DO UPDATE SET
			notifications_enabled = ?,
			notify_practices_by_default = ?,
			quiet_hours_start = ?,
			quiet_hours_end = ?,
			default_timing = ?`,
		userID, s.NotificationsEnabled, s.NotifyByDefault, s.QuietHoursStart, s.QuietHoursEnd, s.DefaultTiming,
		s.NotificationsEnabled, s.NotifyByDefault, s.QuietHoursStart, s.QuietHoursEnd, s.DefaultTiming)
	return err
}

// CleanupOldNotificationLogs removes notification logs older than 7 days.
func (db *DB) CleanupOldNotificationLogs() error {
	cutoff := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	_, err := db.Exec(`DELETE FROM notification_log WHERE date < ?`, cutoff)
	return err
}

// EnableNotifyForScheduledPractices sets notify:true in config JSON for all of a user's
// scheduled practices that don't already have it. Returns the count of updated practices.
func (db *DB) EnableNotifyForScheduledPractices(userID int64) (int64, error) {
	res, err := db.Exec(`
		UPDATE practices
		SET config = jsonb_set(COALESCE(config::jsonb, '{}'), '{notify}', 'true')
		WHERE user_id = ? AND type = 'scheduled' AND status = 'active'
		  AND (config::jsonb->>'notify' IS NULL OR config::jsonb->>'notify' = 'false')`,
		userID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
