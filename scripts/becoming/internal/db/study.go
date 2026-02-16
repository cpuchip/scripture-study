package db

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"
)

// Study mode constants
const (
	// Forward track modes (recall — "what does this scripture say?")
	ModeRevealWhole = "reveal_whole" // Level 1: show full text, user reads
	ModeRevealWords = "reveal_words" // Level 2: ~35% blanked, tap to reveal
	ModeTypeWords   = "type_words"   // Level 3: ~35% blanked, type missing
	ModeArrange     = "arrange"      // Level 3: all words shuffled, rearrange
	ModeTypeFull    = "type_full"    // Level 4: blank canvas, type entire text

	// Reverse track modes (recognition — "where is this from?")
	ModeReverseFull     = "reverse_full"     // R1: full text shown → identify reference
	ModeReversePartial  = "reverse_partial"  // R2: ~35% missing → identify reference
	ModeReverseFragment = "reverse_fragment" // R3: 3-5 key words → identify reference
)

// ForwardLevel maps forward mode to difficulty level (1-4).
var ForwardLevel = map[string]int{
	ModeRevealWhole: 1,
	ModeRevealWords: 2,
	ModeTypeWords:   3,
	ModeArrange:     3,
	ModeTypeFull:    4,
}

// ReverseLevel maps reverse mode to difficulty level (R1-R3, stored as 1-3).
var ReverseLevel = map[string]int{
	ModeReverseFull:     1,
	ModeReversePartial:  2,
	ModeReverseFragment: 3,
}

// AllModes is the complete list of study modes.
var AllModes = []string{
	ModeRevealWhole, ModeRevealWords, ModeTypeWords, ModeArrange, ModeTypeFull,
	ModeReverseFull, ModeReversePartial, ModeReverseFragment,
}

// Aptitude rolling window size
const aptitudeWindow = 5

// SessionMomentum tracks the feel of the current session.
type SessionMomentum string

const (
	MomentumStruggling SessionMomentum = "struggling" // 2+ poor scores in a row
	MomentumSteady     SessionMomentum = "steady"     // mixed results
	MomentumCruising   SessionMomentum = "cruising"   // 3+ good scores in a row
)

// MemorizeScore represents a single exercise score.
type MemorizeScore struct {
	ID         int64     `json:"id"`
	PracticeID int64     `json:"practice_id"`
	UserID     int64     `json:"user_id"`
	Mode       string    `json:"mode"`
	Score      float64   `json:"score"`      // 0.0 to 1.0
	Quality    *int      `json:"quality"`    // SM-2 quality 0-5
	DurationS  *int      `json:"duration_s"` // seconds
	Date       string    `json:"date"`
	CreatedAt  time.Time `json:"created_at"`
}

// MemorizeAptitude represents the cached per-card per-mode aptitude.
type MemorizeAptitude struct {
	ID          int64      `json:"id"`
	PracticeID  int64      `json:"practice_id"`
	UserID      int64      `json:"user_id"`
	Mode        string     `json:"mode"`
	Aptitude    float64    `json:"aptitude"` // 0.0 to 1.0
	SampleCount int        `json:"sample_count"`
	LastScoreAt *time.Time `json:"last_score_at"`
}

// StudyExercise represents the next exercise the algorithm has selected.
type StudyExercise struct {
	Practice  *Practice       `json:"practice"`
	Mode      string          `json:"mode"`
	IsReverse bool            `json:"is_reverse"`
	Level     int             `json:"level"`
	Momentum  SessionMomentum `json:"momentum"`
	CardType  string          `json:"card_type"` // "goldilocks", "stretch", "confidence", "fresh"
}

// StudySessionState tracks the current study session's momentum.
type StudySessionState struct {
	RecentScores  []float64       `json:"recent_scores"` // last N scores (most recent last)
	Momentum      SessionMomentum `json:"momentum"`
	ExercisesDone int             `json:"exercises_done"`
	TotalScore    float64         `json:"total_score"`
}

// NewStudySession creates a fresh session state.
func NewStudySession() *StudySessionState {
	return &StudySessionState{
		RecentScores: []float64{},
		Momentum:     MomentumSteady,
	}
}

// UpdateMomentum recalculates momentum after a new score.
func (s *StudySessionState) UpdateMomentum(score float64) {
	s.RecentScores = append(s.RecentScores, score)
	s.ExercisesDone++
	s.TotalScore += score

	// Keep only last 5 scores for momentum calculation
	if len(s.RecentScores) > 5 {
		s.RecentScores = s.RecentScores[len(s.RecentScores)-5:]
	}

	// Determine momentum from recent trajectory
	if len(s.RecentScores) >= 2 {
		recentPoor := 0
		recentGood := 0
		for i := len(s.RecentScores) - 1; i >= max(0, len(s.RecentScores)-3); i-- {
			if s.RecentScores[i] < 0.6 {
				recentPoor++
			} else if s.RecentScores[i] >= 0.8 {
				recentGood++
			}
		}
		if recentPoor >= 2 {
			s.Momentum = MomentumStruggling
		} else if recentGood >= 3 {
			s.Momentum = MomentumCruising
		} else {
			s.Momentum = MomentumSteady
		}
	}
}

// --- Database operations ---

// RecordScore saves an exercise score and updates the aptitude cache.
func (db *DB) RecordScore(userID int64, score *MemorizeScore) error {
	_, err := db.InsertReturningID(`
		INSERT INTO memorize_scores (practice_id, user_id, mode, score, quality, duration_s, date)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		score.PracticeID, userID, score.Mode, score.Score, score.Quality, score.DurationS, score.Date,
	)
	if err != nil {
		return fmt.Errorf("inserting memorize score: %w", err)
	}

	// Update aptitude cache for this card+mode
	if err := db.updateAptitude(userID, score.PracticeID, score.Mode); err != nil {
		return fmt.Errorf("updating aptitude: %w", err)
	}

	return nil
}

// updateAptitude recalculates the rolling average aptitude for a card+mode.
func (db *DB) updateAptitude(userID, practiceID int64, mode string) error {
	// Get last N scores for this card+mode
	rows, err := db.Query(`
		SELECT score FROM memorize_scores
		WHERE practice_id = ? AND user_id = ? AND mode = ?
		ORDER BY created_at DESC
		LIMIT ?`,
		practiceID, userID, mode, aptitudeWindow,
	)
	if err != nil {
		return fmt.Errorf("fetching recent scores: %w", err)
	}
	defer rows.Close()

	var scores []float64
	for rows.Next() {
		var s float64
		if err := rows.Scan(&s); err != nil {
			return fmt.Errorf("scanning score: %w", err)
		}
		scores = append(scores, s)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if len(scores) == 0 {
		return nil
	}

	// Calculate rolling average
	sum := 0.0
	for _, s := range scores {
		sum += s
	}
	aptitude := sum / float64(len(scores))

	// Upsert aptitude cache
	_, err = db.Exec(`
		INSERT INTO memorize_aptitude (practice_id, user_id, mode, aptitude, sample_count, last_score_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT (practice_id, user_id, mode)
		DO UPDATE SET aptitude = ?, sample_count = ?, last_score_at = CURRENT_TIMESTAMP`,
		practiceID, userID, mode, aptitude, len(scores),
		aptitude, len(scores),
	)
	if err != nil {
		return fmt.Errorf("upserting aptitude: %w", err)
	}

	return nil
}

// GetAptitudes returns all aptitudes for a given card.
func (db *DB) GetAptitudes(userID, practiceID int64) ([]*MemorizeAptitude, error) {
	rows, err := db.Query(`
		SELECT id, practice_id, user_id, mode, aptitude, sample_count, last_score_at
		FROM memorize_aptitude
		WHERE practice_id = ? AND user_id = ?`,
		practiceID, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("getting aptitudes: %w", err)
	}
	defer rows.Close()

	var aptitudes []*MemorizeAptitude
	for rows.Next() {
		a := &MemorizeAptitude{}
		if err := rows.Scan(&a.ID, &a.PracticeID, &a.UserID, &a.Mode, &a.Aptitude, &a.SampleCount, &a.LastScoreAt); err != nil {
			return nil, fmt.Errorf("scanning aptitude: %w", err)
		}
		aptitudes = append(aptitudes, a)
	}
	return aptitudes, rows.Err()
}

// GetAllUserAptitudes returns all aptitudes for a user across all cards.
func (db *DB) GetAllUserAptitudes(userID int64) (map[int64][]*MemorizeAptitude, error) {
	rows, err := db.Query(`
		SELECT id, practice_id, user_id, mode, aptitude, sample_count, last_score_at
		FROM memorize_aptitude
		WHERE user_id = ?`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("getting all aptitudes: %w", err)
	}
	defer rows.Close()

	result := make(map[int64][]*MemorizeAptitude)
	for rows.Next() {
		a := &MemorizeAptitude{}
		if err := rows.Scan(&a.ID, &a.PracticeID, &a.UserID, &a.Mode, &a.Aptitude, &a.SampleCount, &a.LastScoreAt); err != nil {
			return nil, fmt.Errorf("scanning aptitude: %w", err)
		}
		result[a.PracticeID] = append(result[a.PracticeID], a)
	}
	return result, rows.Err()
}

// OverallAptitude calculates the weighted overall aptitude for a card.
// Higher-level modes are weighted more heavily.
func OverallAptitude(aptitudes []*MemorizeAptitude) float64 {
	if len(aptitudes) == 0 {
		return 0.0
	}

	weights := map[string]float64{
		ModeRevealWhole:     1.0,
		ModeRevealWords:     2.0,
		ModeTypeWords:       3.0,
		ModeArrange:         3.0,
		ModeTypeFull:        5.0,
		ModeReverseFull:     1.0,
		ModeReversePartial:  2.5,
		ModeReverseFragment: 4.0,
	}

	totalWeight := 0.0
	weightedSum := 0.0
	for _, a := range aptitudes {
		if a.SampleCount == 0 {
			continue
		}
		w := weights[a.Mode]
		if w == 0 {
			w = 1.0
		}
		totalWeight += w
		weightedSum += a.Aptitude * w
	}

	if totalWeight == 0 {
		return 0.0
	}
	return weightedSum / totalWeight
}

// cardScore pairs a card with a selection weight.
type cardScore struct {
	card  *Practice
	score float64
}

// --- Adaptive Algorithm: Goldilocks Selection ---

// SelectNextExercise picks the next exercise for a study session.
// It considers: card aptitudes, session momentum, and the Goldilocks balance.
func SelectNextExercise(
	cards []*Practice,
	aptitudeMap map[int64][]*MemorizeAptitude,
	session *StudySessionState,
	lastCardID int64,
) *StudyExercise {
	if len(cards) == 0 {
		return nil
	}

	// Determine card type distribution based on momentum
	roll := rand.Float64()
	var cardType string

	switch session.Momentum {
	case MomentumStruggling:
		// Heavy confidence bias: 50% confidence, 10% stretch, 40% goldilocks
		if roll < 0.50 {
			cardType = "confidence"
		} else if roll < 0.60 {
			cardType = "stretch"
		} else {
			cardType = "goldilocks"
		}
	case MomentumCruising:
		// Push harder: 10% confidence, 30% stretch, 60% goldilocks
		if roll < 0.10 {
			cardType = "confidence"
		} else if roll < 0.40 {
			cardType = "stretch"
		} else {
			cardType = "goldilocks"
		}
	default: // MomentumSteady
		// Balanced: 15% confidence, 15% stretch, 70% goldilocks
		if roll < 0.15 {
			cardType = "confidence"
		} else if roll < 0.30 {
			cardType = "stretch"
		} else {
			cardType = "goldilocks"
		}
	}

	// After a miss (last score < 0.6), next should lean confidence
	if len(session.RecentScores) > 0 && session.RecentScores[len(session.RecentScores)-1] < 0.6 {
		if rand.Float64() < 0.7 { // 70% chance of confidence after a miss
			cardType = "confidence"
		}
	}

	// Score each card for the selected card type
	var candidates []cardScore

	for _, card := range cards {
		// Avoid repeating the same card back-to-back (unless only 1 card)
		if card.ID == lastCardID && len(cards) > 1 {
			continue
		}

		apts := aptitudeMap[card.ID]
		overall := OverallAptitude(apts)
		hasSamples := len(apts) > 0 && apts[0].SampleCount > 0

		var s float64
		switch cardType {
		case "confidence":
			// Prefer high-aptitude cards (things they know well)
			s = overall
			if !hasSamples {
				s = 0.1 // fresh cards aren't confident
			}
		case "stretch":
			// Prefer low-to-mid aptitude cards that need growth
			if !hasSamples {
				s = 0.8 // fresh cards are inherently stretchy
			} else {
				s = 1.0 - overall // lower aptitude = better stretch candidate
			}
		case "goldilocks":
			// Prefer mid-range aptitude (0.4-0.8 sweet spot)
			if !hasSamples {
				s = 0.5 // fresh cards are decent goldilocks candidates
			} else {
				// Bell curve peaking at 0.6
				s = math.Exp(-math.Pow(overall-0.6, 2) / 0.08)
			}
		}

		candidates = append(candidates, cardScore{card: card, score: s})
	}

	if len(candidates) == 0 {
		// Fallback: just use the first card
		candidates = append(candidates, cardScore{card: cards[0], score: 1.0})
	}

	// Weighted random selection among candidates
	selected := weightedRandomSelect(candidates)

	// Determine mode for the selected card
	apts := aptitudeMap[selected.ID]
	aptMap := make(map[string]float64)
	for _, a := range apts {
		aptMap[a.Mode] = a.Aptitude
	}

	mode, isReverse, level := selectMode(aptMap, cardType, selected.MemorizeLevel)

	return &StudyExercise{
		Practice:  selected,
		Mode:      mode,
		IsReverse: isReverse,
		Level:     level,
		Momentum:  session.Momentum,
		CardType:  cardType,
	}
}

// selectMode picks a mode based on card aptitudes and the intended card type.
func selectMode(aptMap map[string]float64, cardType string, currentLevel int) (string, bool, int) {
	// Decide forward vs reverse (50/50 base, adjusted by relative aptitude)
	isReverse := rand.Float64() < 0.35 // 35% reverse, 65% forward (forward is primary goal)

	if isReverse {
		return selectReverseMode(aptMap, cardType)
	}
	return selectForwardMode(aptMap, cardType, currentLevel)
}

// selectForwardMode picks a forward mode based on aptitude and card type.
func selectForwardMode(aptMap map[string]float64, cardType string, currentLevel int) (string, bool, int) {
	var targetLevel int

	switch cardType {
	case "confidence":
		// Easy — drop to level 1 or 2
		targetLevel = max(1, currentLevel-2)
		if rand.Float64() < 0.5 {
			targetLevel = 1
		}
	case "stretch":
		// Push up — go 1 level above current
		targetLevel = min(4, currentLevel+1)
	default: // goldilocks
		// Stay at current level, occasionally test adjacent
		roll := rand.Float64()
		if roll < 0.15 {
			targetLevel = max(1, currentLevel-1) // occasionally easier
		} else if roll < 0.30 {
			targetLevel = min(4, currentLevel+1) // occasionally harder
		} else {
			targetLevel = currentLevel
		}
	}

	// Map level to mode
	switch targetLevel {
	case 1:
		return ModeRevealWhole, false, 1
	case 2:
		return ModeRevealWords, false, 2
	case 3:
		// Randomly choose between Type Words and Arrange
		if rand.Float64() < 0.5 {
			return ModeTypeWords, false, 3
		}
		return ModeArrange, false, 3
	case 4:
		return ModeTypeFull, false, 4
	default:
		return ModeRevealWhole, false, 1
	}
}

// selectReverseMode picks a reverse mode based on aptitude and card type.
func selectReverseMode(aptMap map[string]float64, cardType string) (string, bool, int) {
	var targetLevel int

	switch cardType {
	case "confidence":
		targetLevel = 1
	case "stretch":
		targetLevel = 3
	default: // goldilocks
		roll := rand.Float64()
		if roll < 0.3 {
			targetLevel = 1
		} else if roll < 0.7 {
			targetLevel = 2
		} else {
			targetLevel = 3
		}
	}

	switch targetLevel {
	case 1:
		return ModeReverseFull, true, 1
	case 2:
		return ModeReversePartial, true, 2
	case 3:
		return ModeReverseFragment, true, 3
	default:
		return ModeReverseFull, true, 1
	}
}

// weightedRandomSelect does weighted random selection from scored candidates.
func weightedRandomSelect(candidates []cardScore) *Practice {
	// Normalize scores to prevent zero-sum
	totalScore := 0.0
	for _, c := range candidates {
		totalScore += c.score + 0.01 // small epsilon to avoid zero weights
	}

	roll := rand.Float64() * totalScore
	cumulative := 0.0
	for _, c := range candidates {
		cumulative += c.score + 0.01
		if roll <= cumulative {
			return c.card
		}
	}
	return candidates[len(candidates)-1].card
}

// SeedAptitudesFromSM2 seeds initial aptitude data for users with existing SM-2 history.
// This is a one-time operation — cards that already have aptitude data are skipped.
func (db *DB) SeedAptitudesFromSM2(userID int64) error {
	// Get all memorize practices
	practices, err := db.ListPractices(userID, "memorize", true)
	if err != nil {
		return fmt.Errorf("listing memorize practices: %w", err)
	}

	// Get existing aptitudes to skip already-seeded cards
	existingApts, err := db.GetAllUserAptitudes(userID)
	if err != nil {
		return fmt.Errorf("getting existing aptitudes: %w", err)
	}

	seeded := 0
	for _, p := range practices {
		// Skip if this card already has aptitude data
		if len(existingApts[p.ID]) > 0 {
			continue
		}

		var cfg SM2Config
		if err := parseJSON(p.Config, &cfg); err != nil {
			continue
		}

		// Determine seed level based on SM-2 state
		seedLevel := 1
		if cfg.Interval > 30 {
			seedLevel = 4
		} else if cfg.Interval > 14 {
			seedLevel = 3
		} else if cfg.Interval > 3 || cfg.Repetitions >= 3 {
			seedLevel = 2
		}

		// Seed aptitude based on ease factor and interval
		baseAptitude := 0.5
		if cfg.EaseFactor >= 2.5 && cfg.Interval > 7 {
			baseAptitude = 0.8
		} else if cfg.EaseFactor >= 2.0 && cfg.Interval > 3 {
			baseAptitude = 0.65
		} else if cfg.Repetitions == 0 {
			baseAptitude = 0.0
		}

		// Update the practice's memorize_level
		if _, err := db.Exec(`UPDATE practices SET memorize_level = ? WHERE id = ? AND user_id = ?`,
			seedLevel, p.ID, userID); err != nil {
			log.Printf("Warning: failed to seed memorize_level for practice %d: %v", p.ID, err)
			continue
		}

		// Seed forward aptitudes up to the seed level
		forwardModes := []struct {
			mode  string
			level int
		}{
			{ModeRevealWhole, 1},
			{ModeRevealWords, 2},
			{ModeTypeWords, 3},
			{ModeTypeFull, 4},
		}
		for _, fm := range forwardModes {
			if fm.level > seedLevel {
				break
			}
			apt := baseAptitude
			if fm.level < seedLevel {
				apt = min(1.0, baseAptitude+0.15) // higher aptitude at easier levels
			}
			if _, err := db.Exec(`
				INSERT INTO memorize_aptitude (practice_id, user_id, mode, aptitude, sample_count, last_score_at)
				VALUES (?, ?, ?, ?, 1, CURRENT_TIMESTAMP)
				ON CONFLICT (practice_id, user_id, mode) DO NOTHING`,
				p.ID, userID, fm.mode, apt); err != nil {
				log.Printf("Warning: failed to seed aptitude for practice %d mode %s: %v", p.ID, fm.mode, err)
			}
		}

		seeded++
	}

	if seeded > 0 {
		log.Printf("Seeded aptitudes for %d cards from SM-2 history", seeded)
	}
	return nil
}

// UpdateMemorizeLevel updates the target difficulty level for a card based on aptitude.
func (db *DB) UpdateMemorizeLevel(userID, practiceID int64) error {
	apts, err := db.GetAptitudes(userID, practiceID)
	if err != nil {
		return err
	}

	// Find the highest level with aptitude >= 0.7
	newLevel := 1
	for _, a := range apts {
		level, isForward := ForwardLevel[a.Mode]
		if !isForward {
			continue
		}
		if a.Aptitude >= 0.7 && a.SampleCount >= 2 && level >= newLevel {
			newLevel = level
		}
	}

	_, err = db.Exec(`UPDATE practices SET memorize_level = ? WHERE id = ? AND user_id = ?`,
		newLevel, practiceID, userID)
	return err
}

// GetStudySessionScores returns scores from today's study session for a user.
func (db *DB) GetStudySessionScores(userID int64, date string) ([]*MemorizeScore, error) {
	rows, err := db.Query(`
		SELECT id, practice_id, user_id, mode, score, quality, duration_s, date, created_at
		FROM memorize_scores
		WHERE user_id = ? AND date = ?
		ORDER BY created_at`,
		userID, date,
	)
	if err != nil {
		return nil, fmt.Errorf("getting session scores: %w", err)
	}
	defer rows.Close()

	var scores []*MemorizeScore
	for rows.Next() {
		s := &MemorizeScore{}
		if err := rows.Scan(&s.ID, &s.PracticeID, &s.UserID, &s.Mode, &s.Score, &s.Quality, &s.DurationS, &s.Date, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning score: %w", err)
		}
		scores = append(scores, s)
	}
	return scores, rows.Err()
}

// helper to parse JSON config
func parseJSON(data string, v any) error {
	if data == "" || data == "{}" {
		return fmt.Errorf("empty config")
	}
	return json.Unmarshal([]byte(data), v)
}
