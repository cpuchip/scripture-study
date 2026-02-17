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
func (db *DB) ReviewCard(userID, practiceID int64, quality int, date string) (*Practice, error) {
	p, err := db.GetPractice(userID, practiceID)
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
	if err := db.UpdatePractice(userID, p); err != nil {
		return nil, fmt.Errorf("updating practice SM-2 config: %w", err)
	}

	// Log the review
	log := &PracticeLog{
		PracticeID: practiceID,
		Date:       date,
		Quality:    &quality,
		NextReview: &cfg.NextReview,
	}
	if err := db.CreateLog(userID, log); err != nil {
		return nil, fmt.Errorf("logging review: %w", err)
	}

	return p, nil
}

// GetDueCards returns all memorize-type practices due for review on or before the given date,
// excluding cards that have already been reviewed today, scoped to user.
func (db *DB) GetDueCards(userID int64, date string) ([]*Practice, error) {
	nextReview := db.JSONExtract("config", "next_review")
	repetitions := db.JSONExtract("config", "repetitions")

	query := fmt.Sprintf(`
		SELECT `+practiceColumns+`
		FROM practices
		WHERE type = 'memorize' AND status = 'active' AND user_id = ?
		  AND (
		    %s <= ?
		    OR %s IS NULL
		    OR %s = '0'
		  )
		  AND id NOT IN (
		    SELECT DISTINCT practice_id FROM practice_logs
		    WHERE date = ? AND quality IS NOT NULL
		  )
		ORDER BY %s, name`,
		nextReview, nextReview, repetitions, nextReview,
	)

	rows, err := db.Query(query, userID, date, date)
	if err != nil {
		return nil, fmt.Errorf("getting due cards: %w", err)
	}
	defer rows.Close()

	var practices []*Practice
	for rows.Next() {
		p, err := scanPractice(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning due card: %w", err)
		}
		practices = append(practices, p)
	}
	return practices, rows.Err()
}

// GetAllMemorizeCards returns all memorize-type practices with their SM-2 state.
func (db *DB) GetAllMemorizeCards(userID int64) ([]*Practice, error) {
	return db.ListPractices(userID, "memorize", true)
}

// MemorizeCardStatus represents a memorize card with today's review progress.
type MemorizeCardStatus struct {
	Practice        *Practice           `json:"practice"`
	ReviewsToday    int                 `json:"reviews_today"`
	TodayQualities  []int               `json:"today_qualities"`
	IsDue           bool                `json:"is_due"`
	TargetReps      int                 `json:"target_daily_reps"`
	Aptitudes       []*MemorizeAptitude `json:"aptitudes"`
	OverallAptitude float64             `json:"overall_aptitude"`
	IsMastered      bool                `json:"is_mastered"`
	DaysUntilEnd    *int                `json:"days_until_end"`
}

// GetMemorizeCardStatuses returns all active memorize cards with today's review status, scoped to user.
func (db *DB) GetMemorizeCardStatuses(userID int64, date string) ([]*MemorizeCardStatus, error) {
	practices, err := db.ListPractices(userID, "memorize", true)
	if err != nil {
		return nil, fmt.Errorf("listing memorize practices: %w", err)
	}

	// Batch-fetch today's quality scores for all memorize practices
	qualityMap := make(map[int64][]int)
	rows, err := db.Query(`
		SELECT practice_id, quality FROM practice_logs
		WHERE date = ? AND quality IS NOT NULL
		  AND practice_id IN (SELECT id FROM practices WHERE type = 'memorize' AND status = 'active' AND user_id = ?)
		ORDER BY practice_id, logged_at`, date, userID)
	if err != nil {
		return nil, fmt.Errorf("getting today's memorize qualities: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pid int64
		var q int
		if err := rows.Scan(&pid, &q); err != nil {
			return nil, fmt.Errorf("scanning quality row: %w", err)
		}
		qualityMap[pid] = append(qualityMap[pid], q)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Batch-fetch aptitudes for all cards
	aptitudeMap, err := db.GetAllUserAptitudes(userID)
	if err != nil {
		return nil, fmt.Errorf("getting aptitudes: %w", err)
	}

	// Parse today for days-until-end calculation
	todayTime, _ := time.Parse("2006-01-02", date)
	if todayTime.IsZero() {
		todayTime = time.Now()
	}

	var statuses []*MemorizeCardStatus
	for _, p := range practices {
		var cfg SM2Config
		if err := json.Unmarshal([]byte(p.Config), &cfg); err != nil {
			cfg = DefaultSM2Config()
		}

		qualities := qualityMap[p.ID]
		if qualities == nil {
			qualities = []int{}
		}

		isDue := cfg.NextReview <= date || cfg.Repetitions == 0
		targetReps := cfg.TargetDailyReps
		if targetReps < 1 {
			targetReps = 1
		}

		// Aptitude data
		apts := aptitudeMap[p.ID]
		if apts == nil {
			apts = []*MemorizeAptitude{}
		}
		overall := OverallAptitude(apts)

		// Mastery detection: interval >= 21d, overall aptitude >= 0.8, level >= 4, at least 3 modes sampled
		sampledModes := 0
		for _, a := range apts {
			if a.SampleCount > 0 {
				sampledModes++
			}
		}
		isMastered := cfg.Interval >= 21 && overall >= 0.8 && p.MemorizeLevel >= 4 && sampledModes >= 3

		// Days until end date
		var daysUntilEnd *int
		if p.EndDate != nil && *p.EndDate != "" {
			endStr := *p.EndDate
			// Handle both "2006-01-02" and "2006-01-02T15:04:05Z" formats
			if len(endStr) > 10 {
				endStr = endStr[:10]
			}
			if endTime, err := time.Parse("2006-01-02", endStr); err == nil {
				days := int(endTime.Sub(todayTime).Hours() / 24)
				daysUntilEnd = &days
			}
		}

		statuses = append(statuses, &MemorizeCardStatus{
			Practice:        p,
			ReviewsToday:    len(qualities),
			TodayQualities:  qualities,
			IsDue:           isDue,
			TargetReps:      targetReps,
			Aptitudes:       apts,
			OverallAptitude: overall,
			IsMastered:      isMastered,
			DaysUntilEnd:    daysUntilEnd,
		})
	}

	return statuses, nil
}
