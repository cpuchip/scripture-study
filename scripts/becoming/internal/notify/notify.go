// Package notify handles Web Push notifications for due practices.
package notify

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/cpuchip/scripture-study/scripts/becoming/internal/db"
)

// Scheduler checks for due practices and sends push notifications.
type Scheduler struct {
	db             *db.DB
	vapidPublicKey string
	vapidPrivate   string
	vapidContact   string
	stop           chan struct{}
}

// NewScheduler creates a notification scheduler.
func NewScheduler(database *db.DB, vapidPublic, vapidPrivate, vapidContact string) *Scheduler {
	return &Scheduler{
		db:             database,
		vapidPublicKey: vapidPublic,
		vapidPrivate:   vapidPrivate,
		vapidContact:   vapidContact,
		stop:           make(chan struct{}),
	}
}

// Start begins the notification loop, ticking every minute.
func (s *Scheduler) Start() {
	go s.run()
}

// Stop signals the scheduler to stop.
func (s *Scheduler) Stop() {
	close(s.stop)
}

func (s *Scheduler) run() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Println("Notification scheduler started (1-minute interval)")

	// Run once at startup
	s.tick()

	for {
		select {
		case <-ticker.C:
			s.tick()
		case <-s.stop:
			log.Println("Notification scheduler stopped")
			return
		}
	}
}

func (s *Scheduler) tick() {
	now := time.Now()
	date := now.Format("2006-01-02")

	// Periodically clean old logs (once per hour-ish, keyed off minute == 0)
	if now.Minute() == 0 {
		if err := s.db.CleanupOldNotificationLogs(); err != nil {
			log.Printf("notification: cleanup error: %v", err)
		}
	}

	userIDs, err := s.db.UsersWithPushSubscriptions()
	if err != nil {
		log.Printf("notification: error listing users: %v", err)
		return
	}

	for _, userID := range userIDs {
		if !s.db.GetUserNotificationsEnabled(userID) {
			continue
		}

		due, err := s.db.DuePracticesForNotification(userID, date)
		if err != nil {
			log.Printf("notification: error getting due practices for user %d: %v", userID, err)
			continue
		}
		if len(due) == 0 {
			continue
		}

		// Filter out already-sent notifications
		sent, err := s.db.NotificationsSentToday(userID, date)
		if err != nil {
			log.Printf("notification: error checking sent notifications for user %d: %v", userID, err)
			continue
		}

		var unsent []*db.DailySummary
		for _, d := range due {
			if !sent[d.PracticeID] {
				unsent = append(unsent, d)
			}
		}
		if len(unsent) == 0 {
			continue
		}

		payload := buildPayload(unsent)

		subs, err := s.db.GetPushSubscriptions(userID)
		if err != nil {
			log.Printf("notification: error getting subscriptions for user %d: %v", userID, err)
			continue
		}

		for _, sub := range subs {
			s.sendPush(sub, payload)
		}

		// Log the notifications
		for _, d := range unsent {
			if err := s.db.LogNotification(userID, d.PracticeID, date); err != nil {
				log.Printf("notification: error logging notification for practice %d: %v", d.PracticeID, err)
			}
		}
	}
}

// SendTestNotification sends a test notification to all of a user's subscriptions.
func (s *Scheduler) SendTestNotification(userID int64) error {
	subs, err := s.db.GetPushSubscriptions(userID)
	if err != nil {
		return fmt.Errorf("getting subscriptions: %w", err)
	}
	if len(subs) == 0 {
		return fmt.Errorf("no push subscriptions found")
	}

	payload := NotificationPayload{
		Title: "I Become",
		Body:  "Notifications are working!",
		URL:   "/today",
		Tag:   "test",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	for _, sub := range subs {
		s.sendPushBytes(sub, data)
	}
	return nil
}

func (s *Scheduler) sendPush(sub *db.PushSubscription, payload []byte) {
	s.sendPushBytes(sub, payload)
}

func (s *Scheduler) sendPushBytes(sub *db.PushSubscription, payload []byte) {
	resp, err := webpush.SendNotification(payload, &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			P256dh: sub.KeyP256DH,
			Auth:   sub.KeyAuth,
		},
	}, &webpush.Options{
		Subscriber:      s.vapidContact,
		VAPIDPublicKey:  s.vapidPublicKey,
		VAPIDPrivateKey: s.vapidPrivate,
		TTL:             3600,
	})
	if err != nil {
		log.Printf("notification: send error to %s: %v", sub.Endpoint[:40], err)
		return
	}
	defer resp.Body.Close()

	// Handle stale subscriptions
	if resp.StatusCode == http.StatusGone || resp.StatusCode == http.StatusNotFound {
		log.Printf("notification: subscription expired (HTTP %d), removing: %s", resp.StatusCode, sub.Endpoint[:40])
		if err := s.db.DeletePushSubscriptionByEndpoint(sub.Endpoint); err != nil {
			log.Printf("notification: error removing stale subscription: %v", err)
		}
		return
	}

	if resp.StatusCode >= 400 {
		log.Printf("notification: push service returned HTTP %d for %s", resp.StatusCode, sub.Endpoint[:40])
	}
}

// NotificationPayload is the JSON payload sent to the service worker.
type NotificationPayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url"`
	Tag   string `json:"tag"`
}

func buildPayload(practices []*db.DailySummary) []byte {
	var payload NotificationPayload

	if len(practices) == 1 {
		p := practices[0]
		payload = NotificationPayload{
			Title: "Practice Due",
			Body:  p.PracticeName,
			URL:   "/today",
			Tag:   fmt.Sprintf("practice-%d", p.PracticeID),
		}
	} else {
		names := ""
		for i, p := range practices {
			if i > 0 {
				names += ", "
			}
			if i >= 3 {
				names += fmt.Sprintf("and %d more", len(practices)-3)
				break
			}
			names += p.PracticeName
		}
		payload = NotificationPayload{
			Title: fmt.Sprintf("%d Practices Due", len(practices)),
			Body:  names,
			URL:   "/today",
			Tag:   "practices-due",
		}
	}

	data, _ := json.Marshal(payload)
	return data
}

// VAPIDPublicKey returns the public VAPID key (needed by frontend).
func (s *Scheduler) VAPIDPublicKey() string {
	return s.vapidPublicKey
}
