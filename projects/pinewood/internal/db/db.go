// Package db wraps SQLite + migrations + queries for the derby app.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func Open(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		return nil, err
	}
	d := &DB{conn}
	if err := d.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return d, nil
}

const schema = `
CREATE TABLE IF NOT EXISTS race (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  name          TEXT NOT NULL,
  created_at    TEXT NOT NULL,
  status        TEXT NOT NULL,
  parent_id     INTEGER REFERENCES race(id) ON DELETE SET NULL,
  lane_count    INTEGER NOT NULL DEFAULT 3,
  finalized_at  TEXT
);

CREATE TABLE IF NOT EXISTS car (
  id        INTEGER PRIMARY KEY AUTOINCREMENT,
  race_id   INTEGER NOT NULL REFERENCES race(id) ON DELETE CASCADE,
  number    INTEGER NOT NULL,
  name      TEXT,
  UNIQUE(race_id, number)
);

CREATE TABLE IF NOT EXISTS heat (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  race_id     INTEGER NOT NULL REFERENCES race(id) ON DELETE CASCADE,
  heat_number INTEGER NOT NULL,
  status      TEXT NOT NULL DEFAULT 'pending',
  UNIQUE(race_id, heat_number)
);

CREATE TABLE IF NOT EXISTS heat_slot (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  heat_id      INTEGER NOT NULL REFERENCES heat(id) ON DELETE CASCADE,
  lane         INTEGER NOT NULL,
  car_id       INTEGER REFERENCES car(id) ON DELETE CASCADE,
  place        INTEGER,
  recorded_at  TEXT,
  UNIQUE(heat_id, lane)
);

CREATE INDEX IF NOT EXISTS idx_car_race ON car(race_id);
CREATE INDEX IF NOT EXISTS idx_heat_race ON heat(race_id);
CREATE INDEX IF NOT EXISTS idx_slot_heat ON heat_slot(heat_id);
CREATE INDEX IF NOT EXISTS idx_slot_car ON heat_slot(car_id);
`

func (d *DB) migrate() error {
	_, err := d.Exec(schema)
	return err
}

// ---- Models ----

type Race struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	CreatedAt   time.Time  `json:"created_at"`
	Status      string     `json:"status"`
	ParentID    *int64     `json:"parent_id,omitempty"`
	LaneCount   int        `json:"lane_count"`
	FinalizedAt *time.Time `json:"finalized_at,omitempty"`
}

type Car struct {
	ID     int64  `json:"id"`
	RaceID int64  `json:"race_id"`
	Number int    `json:"number"`
	Name   string `json:"name,omitempty"`
}

type Heat struct {
	ID         int64      `json:"id"`
	RaceID     int64      `json:"race_id"`
	HeatNumber int        `json:"heat_number"`
	Status     string     `json:"status"`
	Slots      []HeatSlot `json:"slots"`
}

type HeatSlot struct {
	ID         int64   `json:"id"`
	HeatID     int64   `json:"heat_id"`
	Lane       int     `json:"lane"`
	CarID      *int64  `json:"car_id"`
	CarNumber  int     `json:"car_number,omitempty"`
	CarName    string  `json:"car_name,omitempty"`
	Place      *int    `json:"place"`
	RecordedAt *string `json:"recorded_at,omitempty"`
}

// ---- Race CRUD ----

func (d *DB) CreateRace(ctx context.Context, name string, laneCount int, parentID *int64) (*Race, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := d.ExecContext(ctx,
		`INSERT INTO race(name, created_at, status, parent_id, lane_count) VALUES(?,?,?,?,?)`,
		name, now, "registration", parentID, laneCount)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return d.GetRace(ctx, id)
}

func (d *DB) GetRace(ctx context.Context, id int64) (*Race, error) {
	r := &Race{}
	var created, finalized sql.NullString
	var parent sql.NullInt64
	err := d.QueryRowContext(ctx,
		`SELECT id, name, created_at, status, parent_id, lane_count, finalized_at FROM race WHERE id=?`, id).
		Scan(&r.ID, &r.Name, &created, &r.Status, &parent, &r.LaneCount, &finalized)
	if err != nil {
		return nil, err
	}
	if t, e := time.Parse(time.RFC3339, created.String); e == nil {
		r.CreatedAt = t
	}
	if parent.Valid {
		r.ParentID = &parent.Int64
	}
	if finalized.Valid {
		if t, e := time.Parse(time.RFC3339, finalized.String); e == nil {
			r.FinalizedAt = &t
		}
	}
	return r, nil
}

func (d *DB) ListRaces(ctx context.Context) ([]Race, error) {
	rows, err := d.QueryContext(ctx,
		`SELECT id, name, created_at, status, parent_id, lane_count, finalized_at FROM race ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Race
	for rows.Next() {
		r := Race{}
		var created, finalized sql.NullString
		var parent sql.NullInt64
		if err := rows.Scan(&r.ID, &r.Name, &created, &r.Status, &parent, &r.LaneCount, &finalized); err != nil {
			return nil, err
		}
		if t, e := time.Parse(time.RFC3339, created.String); e == nil {
			r.CreatedAt = t
		}
		if parent.Valid {
			r.ParentID = &parent.Int64
		}
		if finalized.Valid {
			if t, e := time.Parse(time.RFC3339, finalized.String); e == nil {
				r.FinalizedAt = &t
			}
		}
		out = append(out, r)
	}
	return out, nil
}

func (d *DB) UpdateRaceStatus(ctx context.Context, id int64, status string) error {
	_, err := d.ExecContext(ctx, `UPDATE race SET status=? WHERE id=?`, status, id)
	return err
}

func (d *DB) MarkFinalized(ctx context.Context, id int64) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := d.ExecContext(ctx, `UPDATE race SET finalized_at=?, status=? WHERE id=?`, now, "racing", id)
	return err
}

func (d *DB) DeleteRace(ctx context.Context, id int64) error {
	_, err := d.ExecContext(ctx, `DELETE FROM race WHERE id=?`, id)
	return err
}

// ---- Car CRUD ----

func (d *DB) AddCar(ctx context.Context, raceID int64, number int, name string) (*Car, error) {
	res, err := d.ExecContext(ctx,
		`INSERT INTO car(race_id, number, name) VALUES(?,?,?)`, raceID, number, nullable(name))
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &Car{ID: id, RaceID: raceID, Number: number, Name: name}, nil
}

func (d *DB) UpdateCar(ctx context.Context, raceID, carID int64, number int, name string) error {
	_, err := d.ExecContext(ctx,
		`UPDATE car SET number=?, name=? WHERE id=? AND race_id=?`,
		number, nullable(name), carID, raceID)
	return err
}

func (d *DB) DeleteCar(ctx context.Context, raceID, carID int64) error {
	_, err := d.ExecContext(ctx, `DELETE FROM car WHERE id=? AND race_id=?`, carID, raceID)
	return err
}

func (d *DB) ListCars(ctx context.Context, raceID int64) ([]Car, error) {
	rows, err := d.QueryContext(ctx,
		`SELECT id, race_id, number, COALESCE(name,'') FROM car WHERE race_id=? ORDER BY number`, raceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Car
	for rows.Next() {
		c := Car{}
		if err := rows.Scan(&c.ID, &c.RaceID, &c.Number, &c.Name); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

// ---- Heat / slot ----

// ReplacePendingSchedule deletes pending heats and inserts the supplied chart
// starting after the highest existing completed heat number.
func (d *DB) ReplacePendingSchedule(ctx context.Context, raceID int64, chart [][]int) error {
	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete all heats whose slots have no recorded place. Use cascade.
	_, err = tx.ExecContext(ctx, `
		DELETE FROM heat
		WHERE race_id=?
		  AND id NOT IN (SELECT DISTINCT heat_id FROM heat_slot WHERE place IS NOT NULL)
	`, raceID)
	if err != nil {
		return err
	}

	// Find max heat_number among remaining (completed-or-partial) heats.
	var maxN sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT MAX(heat_number) FROM heat WHERE race_id=?`, raceID).Scan(&maxN); err != nil {
		return err
	}
	startN := int(maxN.Int64) + 1

	// Build car number → car id map.
	rows, err := tx.QueryContext(ctx, `SELECT id, number FROM car WHERE race_id=?`, raceID)
	if err != nil {
		return err
	}
	carIDByNum := map[int]int64{}
	for rows.Next() {
		var id int64
		var num int
		if err := rows.Scan(&id, &num); err != nil {
			rows.Close()
			return err
		}
		carIDByNum[num] = id
	}
	rows.Close()

	for offset, heat := range chart {
		hn := startN + offset
		res, err := tx.ExecContext(ctx,
			`INSERT INTO heat(race_id, heat_number, status) VALUES(?,?,?)`,
			raceID, hn, "pending")
		if err != nil {
			return err
		}
		hid, _ := res.LastInsertId()
		for laneIdx, carNum := range heat {
			lane := laneIdx + 1
			var cid sql.NullInt64
			if carNum != 0 {
				if v, ok := carIDByNum[carNum]; ok {
					cid = sql.NullInt64{Int64: v, Valid: true}
				}
			}
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO heat_slot(heat_id, lane, car_id) VALUES(?,?,?)`,
				hid, lane, cid); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (d *DB) ListHeats(ctx context.Context, raceID int64) ([]Heat, error) {
	rows, err := d.QueryContext(ctx, `
		SELECT h.id, h.race_id, h.heat_number, h.status,
		       s.id, s.lane, s.car_id, s.place, s.recorded_at,
		       COALESCE(c.number, 0), COALESCE(c.name, '')
		FROM heat h
		LEFT JOIN heat_slot s ON s.heat_id=h.id
		LEFT JOIN car c ON c.id=s.car_id
		WHERE h.race_id=?
		ORDER BY h.heat_number, s.lane`, raceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	heatMap := map[int64]*Heat{}
	var order []int64
	for rows.Next() {
		var hid, rid int64
		var hn int
		var status string
		var sid sql.NullInt64
		var lane sql.NullInt64
		var cid sql.NullInt64
		var place sql.NullInt64
		var rec sql.NullString
		var num int
		var nm string
		if err := rows.Scan(&hid, &rid, &hn, &status, &sid, &lane, &cid, &place, &rec, &num, &nm); err != nil {
			return nil, err
		}
		h, ok := heatMap[hid]
		if !ok {
			h = &Heat{ID: hid, RaceID: rid, HeatNumber: hn, Status: status}
			heatMap[hid] = h
			order = append(order, hid)
		}
		if sid.Valid {
			slot := HeatSlot{ID: sid.Int64, HeatID: hid, Lane: int(lane.Int64), CarNumber: num, CarName: nm}
			if cid.Valid {
				v := cid.Int64
				slot.CarID = &v
			}
			if place.Valid {
				p := int(place.Int64)
				slot.Place = &p
			}
			if rec.Valid {
				v := rec.String
				slot.RecordedAt = &v
			}
			h.Slots = append(h.Slots, slot)
		}
	}
	out := make([]Heat, 0, len(order))
	for _, id := range order {
		out = append(out, *heatMap[id])
	}
	return out, nil
}

func (d *DB) GetHeat(ctx context.Context, raceID int64, heatNumber int) (*Heat, error) {
	heats, err := d.ListHeats(ctx, raceID)
	if err != nil {
		return nil, err
	}
	for _, h := range heats {
		if h.HeatNumber == heatNumber {
			h := h
			return &h, nil
		}
	}
	return nil, sql.ErrNoRows
}

// SetSlotPlace records or updates a place for a heat slot.
func (d *DB) SetSlotPlace(ctx context.Context, raceID int64, heatNumber, lane int, place *int) error {
	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var heatID int64
	if err := tx.QueryRowContext(ctx,
		`SELECT id FROM heat WHERE race_id=? AND heat_number=?`, raceID, heatNumber).Scan(&heatID); err != nil {
		return err
	}
	now := sql.NullString{String: time.Now().UTC().Format(time.RFC3339), Valid: place != nil}
	var p sql.NullInt64
	if place != nil {
		p = sql.NullInt64{Int64: int64(*place), Valid: true}
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE heat_slot SET place=?, recorded_at=? WHERE heat_id=? AND lane=?`,
		p, now, heatID, lane); err != nil {
		return err
	}
	// Update heat status.
	var total, scored int
	if err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*), SUM(CASE WHEN place IS NOT NULL THEN 1 ELSE 0 END)
		   FROM heat_slot WHERE heat_id=? AND car_id IS NOT NULL`, heatID).
		Scan(&total, &scored); err != nil {
		return err
	}
	status := "pending"
	switch {
	case scored == total && total > 0:
		status = "complete"
	case scored > 0:
		status = "running"
	}
	if _, err := tx.ExecContext(ctx, `UPDATE heat SET status=? WHERE id=?`, status, heatID); err != nil {
		return err
	}
	return tx.Commit()
}

// Standing is one row of the results table.
type Standing struct {
	CarID     int64  `json:"car_id"`
	CarNumber int    `json:"car_number"`
	CarName   string `json:"car_name,omitempty"`
	Total     int    `json:"total"`
	Heats     int    `json:"heats"`
	Rank      int    `json:"rank"`
}

// Standings computes the per-car totals + rank for a race.
func (d *DB) Standings(ctx context.Context, raceID int64) ([]Standing, error) {
	rows, err := d.QueryContext(ctx, `
		SELECT c.id, c.number, COALESCE(c.name, ''),
		       COALESCE(SUM(s.place), 0),
		       COUNT(s.place)
		FROM car c
		LEFT JOIN heat_slot s ON s.car_id=c.id
		WHERE c.race_id=?
		GROUP BY c.id
		ORDER BY c.number`, raceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Standing
	for rows.Next() {
		s := Standing{}
		if err := rows.Scan(&s.CarID, &s.CarNumber, &s.CarName, &s.Total, &s.Heats); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	// Rank by total ascending; ties get same rank.
	sortStandings(out)
	rank := 0
	prev := -1
	for i := range out {
		if out[i].Total != prev {
			rank = i + 1
			prev = out[i].Total
		}
		out[i].Rank = rank
	}
	return out, nil
}

func sortStandings(s []Standing) {
	// Simple insertion sort by Total asc, then car number asc.
	for i := 1; i < len(s); i++ {
		j := i
		for j > 0 && (s[j].Total < s[j-1].Total ||
			(s[j].Total == s[j-1].Total && s[j].CarNumber < s[j-1].CarNumber)) {
			s[j], s[j-1] = s[j-1], s[j]
			j--
		}
	}
}

// CurrentAndOnDeck returns the current (first non-complete) heat and the next pending one.
func (d *DB) CurrentAndOnDeck(ctx context.Context, raceID int64) (current, onDeck *Heat, err error) {
	heats, err := d.ListHeats(ctx, raceID)
	if err != nil {
		return nil, nil, err
	}
	for i := range heats {
		if heats[i].Status != "complete" {
			h := heats[i]
			current = &h
			if i+1 < len(heats) {
				next := heats[i+1]
				onDeck = &next
			}
			return current, onDeck, nil
		}
	}
	return nil, nil, nil
}

func nullable(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// CarNumbers returns the car numbers for a race, ordered by number.
func (d *DB) CarNumbers(ctx context.Context, raceID int64) ([]int, error) {
	cars, err := d.ListCars(ctx, raceID)
	if err != nil {
		return nil, err
	}
	out := make([]int, len(cars))
	for i, c := range cars {
		out[i] = c.Number
	}
	return out, nil
}

// Ties returns groups of car numbers tied for the top N places.
func Ties(standings []Standing, topN int) [][]int {
	if topN < 1 {
		topN = 3
	}
	groups := map[int][]int{}
	var rankOrder []int
	for _, s := range standings {
		if s.Rank > topN {
			continue
		}
		if _, ok := groups[s.Rank]; !ok {
			rankOrder = append(rankOrder, s.Rank)
		}
		groups[s.Rank] = append(groups[s.Rank], s.CarNumber)
	}
	var out [][]int
	for _, r := range rankOrder {
		if len(groups[r]) > 1 {
			out = append(out, groups[r])
		}
	}
	_ = fmt.Sprintf
	return out
}
