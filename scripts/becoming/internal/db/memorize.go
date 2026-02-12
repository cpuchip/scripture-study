package db

import (
	"encoding/json"
	"fmt"
	"math"
	"time"
)

// SM2Config holds the spaced-repetition state stored in practice.config for memorize-type practices.
type SM2Config struct {
	EaseFactor      float64 `json:"ease_factor"`
	Interval        int     `json:"interval"`          // days until next review
	Repetitions     int     `json:"repetitions"`       // consecutive correct reps
	NextReview      string  `json:"next_review"`       // YYYY-MM-DD
	TargetDailyReps int     `json:"target_daily_reps"` // daily practice goal (0 = use SM-2 scheduling only)
}

// DefaultSM2Config returns initial SM-2 parameters for a new card.
func DefaultSM2Config() SM2Config {
	return SM2Config{
		EaseFactor:      2.5,
		Interval:        0,
		Repetitions:     0,
		NextReview:      time.Now().Format("2006-01-02"), // due immediately
		TargetDailyReps: 1,
	}
}

// SM2Review applies the SM-2 algorithm given a quality rating (0-5).
// Returns updated config with new interval, ease factor, repetitions, and next review date.
//
// SM-2 Algorithm (Piotr Wozniak, 1987):
//
//	quality 0: complete blackout
//	quality 1: incorrect, but remembered upon seeing answer
//	quality 2: incorrect, but answer seemed easy to recall
//	quality 3: correct with serious difficulty
//	quality 4: correct after hesitation
//	quality 5: perfect response
//
// If quality < 3: reset repetitions to 0, interval to 1
// If quality >= 3: advance interval based on repetitions
// Ease factor adjusted: EF' = EF + (0.1 - (5-q) * (0.08 + (5-q)*0.02))
// Minimum ease factor: 1.3
func SM2Review(cfg SM2Config, quality int) SM2Config {
	if quality < 0 {
		quality = 0
	}
	if quality > 5 {
		quality = 5
	}

	// Update ease factor
	q := float64(quality)
	cfg.EaseFactor = cfg.EaseFactor + (0.1 - (5-q)*(0.08+(5-q)*0.02))
	if cfg.EaseFactor < 1.3 {
		cfg.EaseFactor = 1.3
	}

	if quality < 3 {
		// Failed: reset
		cfg.Repetitions = 0
		cfg.Interval = 1
	} else {
		// Passed: advance
		cfg.Repetitions++
		switch cfg.Repetitions {
		case 1:
			cfg.Interval = 1
		case 2:
			cfg.Interval = 6
		default:
			cfg.Interval = int(math.Round(float64(cfg.Interval) * cfg.EaseFactor))
		}
	}

	// Calculate next review date
	next := time.Now().AddDate(0, 0, cfg.Interval)
	cfg.NextReview = next.Format("2006-01-02")

	return cfg
}

// ReviewCard processes a memorization review: updates SM-2 state on the practice,
// logs the review, and returns the updated practice.
func (db *DB) ReviewCard(practiceID int64, quality int, date string) (*Practice, error) {
	p, err := db.GetPractice(practiceID)
	if err != nil {
		return nil, fmt.Errorf("getting practice for review: %w", err)
	}
	if p == nil {
		return nil, fmt.Errorf("practice %d not found", practiceID)
	}
	if p.Type != "memorize" {
		return nil, fmt.Errorf("practice %d is type %q, not memorize", practiceID, p.Type)
	}

	// Parse current SM-2 config
	var cfg SM2Config
	if err := json.Unmarshal([]byte(p.Config), &cfg); err != nil {
		cfg = DefaultSM2Config()
	}

	// Apply SM-2
	cfg = SM2Review(cfg, quality)

	// Save updated config back to practice
	cfgJSON, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshaling SM-2 config: %w", err)
	}
	p.Config = string(cfgJSON)
	if err := db.UpdatePractice(p); err != nil {
		return nil, fmt.Errorf("updating practice SM-2 config: %w", err)
	}

	// Log the review
	log := &PracticeLog{
		PracticeID: practiceID,
		Date:       date,
		Quality:    &quality,
		NextReview: &cfg.NextReview,
	}
	if err := db.CreateLog(log); err != nil {
		return nil, fmt.Errorf("logging review: %w", err)
	}

	return p, nil
}

// GetDueCards returns all memorize-type practices due for review on or before the given date,
// excluding cards that have already been reviewed today.
func (db *DB) GetDueCards(date string) ([]*Practice, error) {
	rows, err := db.Query(`
		SELECT id, name, description, type, category, source_doc, source_path, config, sort_order, active, created_at, completed_at
		FROM practices
		WHERE type = 'memorize' AND active = 1
		  AND (
		    json_extract(config, '$.next_review') <= ?
		    OR json_extract(config, '$.next_review') IS NULL
		    OR json_extract(config, '$.repetitions') = 0
		  )
		  AND id NOT IN (
		    SELECT DISTINCT practice_id FROM practice_logs
		    WHERE date = ? AND quality IS NOT NULL
		  )
		ORDER BY json_extract(config, '$.next_review'), name`,
		date, date,
	)
	if err != nil {
		return nil, fmt.Errorf("getting due cards: %w", err)
	}
	defer rows.Close()

	var practices []*Practice
	for rows.Next() {
		p := &Practice{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Type, &p.Category, &p.SourceDoc, &p.SourcePath, &p.Config, &p.SortOrder, &p.Active, &p.CreatedAt, &p.CompletedAt); err != nil {
			return nil, fmt.Errorf("scanning due card: %w", err)
		}
		practices = append(practices, p)
	}
	return practices, rows.Err()
}

// GetAllMemorizeCards returns all memorize-type practices with their SM-2 state.
func (db *DB) GetAllMemorizeCards() ([]*Practice, error) {
	return db.ListPractices("memorize", true)
}
