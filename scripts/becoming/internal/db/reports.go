package db

import (
	"fmt"
	"time"
)

// ReportEntry holds aggregated stats for a single practice over a date range.
type ReportEntry struct {
	PracticeID     int64            `json:"practice_id"`
	PracticeName   string           `json:"practice_name"`
	PracticeType   string           `json:"practice_type"`
	Category       string           `json:"category"`
	Config         string           `json:"config"`
	TotalLogs      int              `json:"total_logs"`
	TotalSets      int              `json:"total_sets"`
	TotalReps      int              `json:"total_reps"`
	DaysActive     int              `json:"days_active"`
	DaysInRange    int              `json:"days_in_range"`
	CompletionRate float64          `json:"completion_rate"` // days_complete / days_in_range
	CurrentStreak  int              `json:"current_streak"`
	DailyData      []DailyDataPoint `json:"daily_data"`
}

// DailyDataPoint holds per-day log aggregation for a single practice.
type DailyDataPoint struct {
	Date string `json:"date"`
	Logs int    `json:"logs"`
	Sets int    `json:"sets"`
	Reps int    `json:"reps"`
}

// GetReport returns aggregated stats per active practice for a date range.
func (db *DB) GetReport(startDate, endDate string) ([]*ReportEntry, error) {
	// Query: group logs by practice and date within the range.
	rows, err := db.Query(`
		SELECT
			p.id, p.name, p.type, p.category, p.config,
			l.date,
			COUNT(l.id) as logs,
			COALESCE(SUM(l.sets), 0) as sets,
			COALESCE(SUM(l.reps), 0) as reps
		FROM practices p
		LEFT JOIN practice_logs l ON l.practice_id = p.id AND l.date >= ? AND l.date <= ?
		WHERE p.active = 1
		GROUP BY p.id, l.date
		ORDER BY p.sort_order, p.type, p.name, l.date`,
		startDate, endDate,
	)
	if err != nil {
		return nil, fmt.Errorf("getting report: %w", err)
	}
	defer rows.Close()

	// Collect raw rows into a map of practice_id → ReportEntry.
	entries := make(map[int64]*ReportEntry)
	var order []int64 // preserve query order

	for rows.Next() {
		var (
			pid      int64
			name     string
			ptype    string
			category string
			config   string
			date     *string
			logs     int
			sets     int
			reps     int
		)
		if err := rows.Scan(&pid, &name, &ptype, &category, &config, &date, &logs, &sets, &reps); err != nil {
			return nil, fmt.Errorf("scanning report row: %w", err)
		}

		entry, exists := entries[pid]
		if !exists {
			entry = &ReportEntry{
				PracticeID:   pid,
				PracticeName: name,
				PracticeType: ptype,
				Category:     category,
				Config:       config,
			}
			entries[pid] = entry
			order = append(order, pid)
		}

		if date != nil && *date != "" {
			entry.DailyData = append(entry.DailyData, DailyDataPoint{
				Date: *date,
				Logs: logs,
				Sets: sets,
				Reps: reps,
			})
			entry.TotalLogs += logs
			entry.TotalSets += sets
			entry.TotalReps += reps
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Calculate derived stats.
	start := mustParseDate(startDate)
	end := mustParseDate(endDate)
	daysInRange := int(end.Sub(start).Hours()/24) + 1
	today := time.Now().Format("2006-01-02")

	for _, entry := range entries {
		entry.DaysInRange = daysInRange

		// Build a set of active dates.
		activeDates := make(map[string]bool, len(entry.DailyData))
		for _, dp := range entry.DailyData {
			if dp.Logs > 0 {
				activeDates[dp.Date] = true
			}
		}
		entry.DaysActive = len(activeDates)

		// Completion rate = days with any activity / days in range.
		if daysInRange > 0 {
			entry.CompletionRate = float64(entry.DaysActive) / float64(daysInRange)
		}

		// Current streak: consecutive days ending at today (or yesterday).
		entry.CurrentStreak = calcStreak(activeDates, today)
	}

	// Return in query order.
	result := make([]*ReportEntry, 0, len(order))
	for _, pid := range order {
		result = append(result, entries[pid])
	}
	return result, nil
}

// calcStreak counts consecutive days with activity ending at or before refDate.
func calcStreak(activeDates map[string]bool, refDate string) int {
	d := mustParseDate(refDate)
	streak := 0
	// Allow starting from today or yesterday (if today isn't done yet).
	if !activeDates[d.Format("2006-01-02")] {
		d = d.AddDate(0, 0, -1)
	}
	for i := 0; i < 365; i++ {
		ds := d.AddDate(0, 0, -i).Format("2006-01-02")
		if activeDates[ds] {
			streak++
		} else {
			break
		}
	}
	return streak
}
