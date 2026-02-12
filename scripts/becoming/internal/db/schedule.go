package db

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ScheduleConfig defines when a scheduled practice is due.
type ScheduleConfig struct {
	Type         string   `json:"type"`                     // interval | daily_slots | weekly | monthly | once
	IntervalDays int      `json:"interval_days,omitempty"`  // for "interval": every N days
	AnchorDate   string   `json:"anchor_date,omitempty"`    // for "interval": starting reference date (YYYY-MM-DD)
	ShiftOnEarly bool     `json:"shift_on_early,omitempty"` // for "interval": shift schedule when done early
	Slots        []string `json:"slots,omitempty"`          // for "daily_slots": ["morning","lunch","night"]
	Days         []string `json:"days,omitempty"`           // for "weekly": ["mon","wed","fri"]
	DayOfMonth   int      `json:"day_of_month,omitempty"`   // for "monthly": 1-31
	DueDate      string   `json:"due_date,omitempty"`       // for "once": YYYY-MM-DD
}

// ScheduleStatus is the computed due state for a scheduled practice on a given date.
type ScheduleStatus struct {
	IsDue       bool     `json:"is_due"`
	NextDue     string   `json:"next_due,omitempty"`     // YYYY-MM-DD
	DaysOverdue int      `json:"days_overdue,omitempty"` // 0 if on time
	SlotsDue    []string `json:"slots_due,omitempty"`    // remaining slots for daily_slots type
}

// ParseScheduleConfig extracts the schedule config from a practice's JSON config.
func ParseScheduleConfig(configJSON string) (*ScheduleConfig, error) {
	var wrapper struct {
		Schedule ScheduleConfig `json:"schedule"`
	}
	if err := json.Unmarshal([]byte(configJSON), &wrapper); err != nil {
		return nil, fmt.Errorf("parsing schedule config: %w", err)
	}
	if wrapper.Schedule.Type == "" {
		return nil, fmt.Errorf("schedule.type is required")
	}
	return &wrapper.Schedule, nil
}

// IsScheduledDue computes whether a scheduled practice is due on the given date.
// lastLogDate is the most recent log date for this practice (empty string if never logged).
// completedSlots are the slot values from today's logs (for daily_slots type).
func IsScheduledDue(sched *ScheduleConfig, date string, lastLogDate string, completedSlots []string) ScheduleStatus {
	switch sched.Type {
	case "interval":
		return intervalDue(sched, date, lastLogDate)
	case "daily_slots":
		return dailySlotsDue(sched, date, completedSlots)
	case "weekly":
		return weeklyDue(sched, date)
	case "monthly":
		return monthlyDue(sched, date)
	case "once":
		return onceDue(sched, date, lastLogDate)
	default:
		return ScheduleStatus{IsDue: false, NextDue: date}
	}
}

// intervalDue: every N days from anchor or last completion.
func intervalDue(sched *ScheduleConfig, date string, lastLogDate string) ScheduleStatus {
	d := mustParseDate(date)
	interval := sched.IntervalDays
	if interval <= 0 {
		interval = 1
	}

	// Determine the reference point.
	var ref time.Time
	if lastLogDate != "" && sched.ShiftOnEarly {
		// Shift mode: next due is relative to last completion.
		ref = mustParseDate(lastLogDate)
	} else if lastLogDate != "" {
		// Fixed mode: next due is relative to last completion (same logic, keeps it simple).
		ref = mustParseDate(lastLogDate)
	} else if sched.AnchorDate != "" {
		ref = mustParseDate(sched.AnchorDate)
		// If anchor is in the future, not yet due.
		if ref.After(d) {
			return ScheduleStatus{IsDue: false, NextDue: sched.AnchorDate}
		}
		// Check if today is a valid interval day from anchor.
		daysSince := int(d.Sub(ref).Hours() / 24)
		isDue := daysSince%interval == 0
		nextDue := ref.AddDate(0, 0, ((daysSince/interval)+1)*interval)
		if isDue {
			nextDue = ref.AddDate(0, 0, ((daysSince/interval)+1)*interval)
		} else {
			nextDue = ref.AddDate(0, 0, ((daysSince/interval)+1)*interval)
		}
		return ScheduleStatus{
			IsDue:   isDue,
			NextDue: nextDue.Format("2006-01-02"),
		}
	} else {
		// No anchor, no logs — due today.
		return ScheduleStatus{IsDue: true, NextDue: d.AddDate(0, 0, interval).Format("2006-01-02")}
	}

	// With a reference date (last log), compute next due.
	nextDue := ref.AddDate(0, 0, interval)
	daysOverdue := 0

	if d.Before(nextDue) {
		// Not yet due.
		return ScheduleStatus{IsDue: false, NextDue: nextDue.Format("2006-01-02")}
	}

	// Due or overdue.
	if d.After(nextDue) {
		daysOverdue = int(d.Sub(nextDue).Hours() / 24)
	}
	return ScheduleStatus{
		IsDue:       true,
		NextDue:     nextDue.Format("2006-01-02"),
		DaysOverdue: daysOverdue,
	}
}

// dailySlotsDue: multiple completions per day (morning/lunch/night).
func dailySlotsDue(sched *ScheduleConfig, date string, completedSlots []string) ScheduleStatus {
	done := make(map[string]bool, len(completedSlots))
	for _, s := range completedSlots {
		done[strings.ToLower(strings.TrimSpace(s))] = true
	}

	var remaining []string
	for _, slot := range sched.Slots {
		if !done[strings.ToLower(strings.TrimSpace(slot))] {
			remaining = append(remaining, slot)
		}
	}

	return ScheduleStatus{
		IsDue:    len(remaining) > 0,
		SlotsDue: remaining,
		NextDue:  date, // Always due today (slots reset daily).
	}
}

// weeklyDue: specific days of the week.
func weeklyDue(sched *ScheduleConfig, date string) ScheduleStatus {
	d := mustParseDate(date)
	dayName := strings.ToLower(d.Weekday().String()[:3]) // "mon","tue",...

	isDue := false
	for _, day := range sched.Days {
		if strings.ToLower(day) == dayName {
			isDue = true
			break
		}
	}

	// Compute next due day.
	nextDue := nextWeeklyDue(d, sched.Days)

	return ScheduleStatus{
		IsDue:   isDue,
		NextDue: nextDue.Format("2006-01-02"),
	}
}

// nextWeeklyDue finds the next occurrence of one of the given days starting from the day after d.
func nextWeeklyDue(d time.Time, days []string) time.Time {
	daySet := make(map[string]bool, len(days))
	for _, day := range days {
		daySet[strings.ToLower(day)] = true
	}
	for i := 1; i <= 7; i++ {
		next := d.AddDate(0, 0, i)
		name := strings.ToLower(next.Weekday().String()[:3])
		if daySet[name] {
			return next
		}
	}
	return d.AddDate(0, 0, 7) // fallback
}

// monthlyDue: specific day of the month.
func monthlyDue(sched *ScheduleConfig, date string) ScheduleStatus {
	d := mustParseDate(date)
	isDue := d.Day() == sched.DayOfMonth

	// Compute next due.
	var nextDue time.Time
	if isDue || d.Day() > sched.DayOfMonth {
		// Next month.
		nextMonth := d.AddDate(0, 1, 0)
		nextDue = time.Date(nextMonth.Year(), nextMonth.Month(), sched.DayOfMonth, 0, 0, 0, 0, time.UTC)
	} else {
		// This month still.
		nextDue = time.Date(d.Year(), d.Month(), sched.DayOfMonth, 0, 0, 0, 0, time.UTC)
	}

	return ScheduleStatus{
		IsDue:   isDue,
		NextDue: nextDue.Format("2006-01-02"),
	}
}

// onceDue: one-time task with a due date.
func onceDue(sched *ScheduleConfig, date string, lastLogDate string) ScheduleStatus {
	if lastLogDate != "" {
		// Already completed.
		return ScheduleStatus{IsDue: false}
	}
	d := mustParseDate(date)
	due := mustParseDate(sched.DueDate)

	daysOverdue := 0
	if d.After(due) {
		daysOverdue = int(d.Sub(due).Hours() / 24)
	}

	return ScheduleStatus{
		IsDue:       !d.Before(due),
		NextDue:     sched.DueDate,
		DaysOverdue: daysOverdue,
	}
}

func mustParseDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Now().UTC().Truncate(24 * time.Hour)
	}
	return t
}

// GetLastLogDate returns the most recent log date for a practice, or empty string if none.
func (db *DB) GetLastLogDate(practiceID int64) (string, error) {
	var date *string
	err := db.QueryRow(`
		SELECT MAX(date) FROM practice_logs WHERE practice_id = ?`,
		practiceID,
	).Scan(&date)
	if err != nil || date == nil {
		return "", err
	}
	return *date, nil
}

// GetTodaySlots returns the log values for a practice on a given date (for daily_slots tracking).
func (db *DB) GetTodaySlots(practiceID int64, date string) ([]string, error) {
	rows, err := db.Query(`
		SELECT COALESCE(value, '') FROM practice_logs
		WHERE practice_id = ? AND date = ?
		ORDER BY logged_at`, practiceID, date,
	)
	if err != nil {
		return nil, fmt.Errorf("getting today slots: %w", err)
	}
	defer rows.Close()

	var slots []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		if v != "" {
			slots = append(slots, v)
		}
	}
	return slots, rows.Err()
}
