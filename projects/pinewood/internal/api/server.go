// Package api wires HTTP + WebSocket handlers around DB and schedule.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cpuchip/pinewood/internal/audit"
	"github.com/cpuchip/pinewood/internal/db"
	"github.com/cpuchip/pinewood/internal/excel"
	"github.com/cpuchip/pinewood/internal/schedule"
	"github.com/cpuchip/pinewood/internal/ws"
)

type Server struct {
	DB     *db.DB
	Audit  *audit.Logger
	Hub    *ws.Hub
	Static fs.FS
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/api/races", s.handleRaces)
	mux.HandleFunc("/api/races/", s.handleRaceSub)
	mux.HandleFunc("/api/import", s.handleImport)

	mux.HandleFunc("/ws", s.Hub.Handle)

	// SPA static
	if s.Static != nil {
		mux.Handle("/", spaHandler{fs: s.Static})
	}

	return mux
}

// ---- Race CRUD ----

func (s *Server) handleRaces(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		races, err := s.DB.ListRaces(r.Context())
		if err != nil {
			httpErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, races)
	case http.MethodPost:
		var body struct {
			Name      string `json:"name"`
			LaneCount int    `json:"lane_count"`
			ParentID  *int64 `json:"parent_id,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httpErr(w, err)
			return
		}
		if body.LaneCount == 0 {
			body.LaneCount = 3
		}
		if body.Name == "" {
			body.Name = "Pinewood Derby " + time.Now().Format("2006-01-02")
		}
		race, err := s.DB.CreateRace(r.Context(), body.Name, body.LaneCount, body.ParentID)
		if err != nil {
			httpErr(w, err)
			return
		}
		s.Audit.Log("race_created", map[string]interface{}{"race_id": race.ID, "name": race.Name})
		writeJSON(w, http.StatusCreated, race)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleRaceSub(w http.ResponseWriter, r *http.Request) {
	// /api/races/{id}/...
	path := strings.TrimPrefix(r.URL.Path, "/api/races/")
	parts := strings.Split(path, "/")
	if len(parts) < 1 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if len(parts) == 1 {
		s.handleRace(w, r, id)
		return
	}
	switch parts[1] {
	case "cars":
		if len(parts) == 2 {
			s.handleCars(w, r, id)
		} else {
			s.handleCar(w, r, id, parts[2])
		}
	case "finalize":
		s.handleFinalize(w, r, id)
	case "schedule":
		s.handleSchedule(w, r, id)
	case "heats":
		if len(parts) == 2 {
			s.handleHeats(w, r, id)
		} else {
			s.handleHeatSub(w, r, id, parts[2:])
		}
	case "standings":
		s.handleStandings(w, r, id)
	case "state":
		s.handleState(w, r, id)
	case "runoff":
		s.handleRunoff(w, r, id)
	case "export":
		s.handleExport(w, r, id)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleRace(w http.ResponseWriter, r *http.Request, id int64) {
	switch r.Method {
	case http.MethodGet:
		race, err := s.DB.GetRace(r.Context(), id)
		if err != nil {
			httpErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, race)
	case http.MethodDelete:
		if err := s.DB.DeleteRace(r.Context(), id); err != nil {
			httpErr(w, err)
			return
		}
		s.Audit.Log("race_deleted", map[string]interface{}{"race_id": id})
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// ---- Cars ----

func (s *Server) handleCars(w http.ResponseWriter, r *http.Request, raceID int64) {
	switch r.Method {
	case http.MethodGet:
		cars, err := s.DB.ListCars(r.Context(), raceID)
		if err != nil {
			httpErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, cars)
	case http.MethodPost:
		var body struct {
			Number int    `json:"number"`
			Name   string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httpErr(w, err)
			return
		}
		car, err := s.DB.AddCar(r.Context(), raceID, body.Number, body.Name)
		if err != nil {
			httpErr(w, err)
			return
		}
		s.Audit.Log("car_added", map[string]interface{}{"race_id": raceID, "number": body.Number})
		// If race already finalized, regenerate schedule.
		s.maybeRegenerate(r.Context(), raceID, "late_add")
		s.Hub.Broadcast("cars_changed", map[string]int64{"race_id": raceID})
		writeJSON(w, http.StatusCreated, car)
	}
}

func (s *Server) handleCar(w http.ResponseWriter, r *http.Request, raceID int64, idStr string) {
	carID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodPut:
		var body struct {
			Number int    `json:"number"`
			Name   string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httpErr(w, err)
			return
		}
		if err := s.DB.UpdateCar(r.Context(), raceID, carID, body.Number, body.Name); err != nil {
			httpErr(w, err)
			return
		}
		s.Audit.Log("car_updated", map[string]interface{}{"race_id": raceID, "car_id": carID})
		s.Hub.Broadcast("cars_changed", map[string]int64{"race_id": raceID})
		w.WriteHeader(http.StatusNoContent)
	case http.MethodDelete:
		if err := s.DB.DeleteCar(r.Context(), raceID, carID); err != nil {
			httpErr(w, err)
			return
		}
		s.Audit.Log("car_deleted", map[string]interface{}{"race_id": raceID, "car_id": carID})
		s.maybeRegenerate(r.Context(), raceID, "car_removed")
		s.Hub.Broadcast("cars_changed", map[string]int64{"race_id": raceID})
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// ---- Finalize / schedule ----

func (s *Server) handleFinalize(w http.ResponseWriter, r *http.Request, raceID int64) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := s.regenerateSchedule(r.Context(), raceID); err != nil {
		httpErr(w, err)
		return
	}
	if err := s.DB.MarkFinalized(r.Context(), raceID); err != nil {
		httpErr(w, err)
		return
	}
	s.Audit.Log("race_finalized", map[string]interface{}{"race_id": raceID})
	s.Hub.Broadcast("schedule_changed", map[string]interface{}{"race_id": raceID, "reason": "finalize"})
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleSchedule(w http.ResponseWriter, r *http.Request, raceID int64) {
	heats, err := s.DB.ListHeats(r.Context(), raceID)
	if err != nil {
		httpErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, heats)
}

func (s *Server) handleHeats(w http.ResponseWriter, r *http.Request, raceID int64) {
	heats, err := s.DB.ListHeats(r.Context(), raceID)
	if err != nil {
		httpErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, heats)
}

func (s *Server) handleHeatSub(w http.ResponseWriter, r *http.Request, raceID int64, parts []string) {
	heatNum, err := strconv.Atoi(parts[0])
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if len(parts) == 1 {
		// GET heat
		h, err := s.DB.GetHeat(r.Context(), raceID, heatNum)
		if err != nil {
			httpErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, h)
		return
	}
	if parts[1] != "score" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Lane  int  `json:"lane"`
		Place *int `json:"place"` // null clears
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpErr(w, err)
		return
	}
	if body.Place != nil && (*body.Place < 1 || *body.Place > 3) {
		http.Error(w, "place must be 1, 2 or 3", http.StatusBadRequest)
		return
	}
	if err := s.DB.SetSlotPlace(r.Context(), raceID, heatNum, body.Lane, body.Place); err != nil {
		httpErr(w, err)
		return
	}
	s.Audit.Log("score", map[string]interface{}{
		"race_id": raceID, "heat": heatNum, "lane": body.Lane, "place": body.Place,
	})
	// Broadcast updated state.
	s.broadcastState(r.Context(), raceID)
	w.WriteHeader(http.StatusNoContent)
}

// ---- Standings, run-off, export ----

func (s *Server) handleStandings(w http.ResponseWriter, r *http.Request, raceID int64) {
	st, err := s.DB.Standings(r.Context(), raceID)
	if err != nil {
		httpErr(w, err)
		return
	}
	ties := db.Ties(st, 3)
	writeJSON(w, http.StatusOK, map[string]interface{}{"standings": st, "ties": ties})
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request, raceID int64) {
	state, err := s.collectState(r.Context(), raceID)
	if err != nil {
		httpErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (s *Server) handleRunoff(w http.ResponseWriter, r *http.Request, raceID int64) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Cars []int  `json:"cars"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpErr(w, err)
		return
	}
	if len(body.Cars) < 2 {
		http.Error(w, "need >=2 cars", http.StatusBadRequest)
		return
	}
	parent, err := s.DB.GetRace(r.Context(), raceID)
	if err != nil {
		httpErr(w, err)
		return
	}
	if body.Name == "" {
		body.Name = parent.Name + " — Run-off"
	}
	lanes, _, _ := schedule.RunoffPlan(len(body.Cars))
	parentID := parent.ID
	newRace, err := s.DB.CreateRace(r.Context(), body.Name, lanes, &parentID)
	if err != nil {
		httpErr(w, err)
		return
	}
	// Add cars (copy name from parent if known).
	parentCars, _ := s.DB.ListCars(r.Context(), raceID)
	nameByNum := map[int]string{}
	for _, c := range parentCars {
		nameByNum[c.Number] = c.Name
	}
	for _, num := range body.Cars {
		if _, err := s.DB.AddCar(r.Context(), newRace.ID, num, nameByNum[num]); err != nil {
			httpErr(w, err)
			return
		}
	}
	// Generate schedule.
	chart, err := schedule.GenerateRunoff(body.Cars)
	if err != nil {
		httpErr(w, err)
		return
	}
	heats := make([][]int, len(chart.Heats))
	for i, h := range chart.Heats {
		heats[i] = h
	}
	if err := s.DB.ReplacePendingSchedule(r.Context(), newRace.ID, heats); err != nil {
		httpErr(w, err)
		return
	}
	if err := s.DB.MarkFinalized(r.Context(), newRace.ID); err != nil {
		httpErr(w, err)
		return
	}
	s.Audit.Log("runoff_created", map[string]interface{}{
		"parent_id": raceID, "race_id": newRace.ID, "cars": body.Cars,
	})
	s.Hub.Broadcast("schedule_changed", map[string]interface{}{"race_id": newRace.ID, "reason": "runoff"})
	writeJSON(w, http.StatusCreated, newRace)
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request, raceID int64) {
	race, err := s.DB.GetRace(r.Context(), raceID)
	if err != nil {
		httpErr(w, err)
		return
	}
	cars, err := s.DB.ListCars(r.Context(), raceID)
	if err != nil {
		httpErr(w, err)
		return
	}
	heats, err := s.DB.ListHeats(r.Context(), raceID)
	if err != nil {
		httpErr(w, err)
		return
	}
	standings, err := s.DB.Standings(r.Context(), raceID)
	if err != nil {
		httpErr(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.xlsx"`, sanitize(race.Name)))
	if err := excel.Export(w, race, cars, heats, standings); err != nil {
		httpErr(w, err)
	}
}

func (s *Server) handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		httpErr(w, err)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		httpErr(w, err)
		return
	}
	defer file.Close()
	name := strings.TrimSuffix(header.Filename, ".xlsx")
	data, err := excel.Parse(file, name)
	if err != nil {
		httpErr(w, err)
		return
	}
	ctx := r.Context()
	race, err := s.DB.CreateRace(ctx, data.RaceName, data.LaneCount, nil)
	if err != nil {
		httpErr(w, err)
		return
	}
	for _, c := range data.Cars {
		s.DB.AddCar(ctx, race.ID, c.Number, c.Name)
	}
	// Insert heats with car assignments.
	heatChart := make([][]int, len(data.Heats))
	for i, h := range data.Heats {
		heatChart[i] = h
	}
	if err := s.DB.ReplacePendingSchedule(ctx, race.ID, heatChart); err != nil {
		httpErr(w, err)
		return
	}
	// Apply scores.
	for i, scores := range data.Scores {
		heatNum := i + 1
		for laneIdx, p := range scores {
			if p == nil {
				continue
			}
			s.DB.SetSlotPlace(ctx, race.ID, heatNum, laneIdx+1, p)
		}
	}
	s.DB.UpdateRaceStatus(ctx, race.ID, "complete")
	s.Audit.Log("race_imported", map[string]interface{}{"race_id": race.ID, "warnings": data.Warnings})

	resp := map[string]interface{}{"race": race, "warnings": data.Warnings}
	// Compare standings.
	if len(data.Standings) > 0 {
		got, _ := s.DB.Standings(ctx, race.ID)
		mismatches := compareStandings(data.Standings, got)
		if len(mismatches) > 0 {
			resp["mismatches"] = mismatches
		}
	}
	writeJSON(w, http.StatusCreated, resp)
}

// ---- helpers ----

func (s *Server) regenerateSchedule(ctx context.Context, raceID int64) error {
	cars, err := s.DB.CarNumbers(ctx, raceID)
	if err != nil {
		return err
	}
	if len(cars) < 3 {
		return fmt.Errorf("need at least 3 cars to schedule (got %d)", len(cars))
	}
	chart, err := schedule.Generate(cars, schedule.Default())
	if err != nil {
		return err
	}
	heats := make([][]int, len(chart.Heats))
	for i, h := range chart.Heats {
		heats[i] = h
	}
	return s.DB.ReplacePendingSchedule(ctx, raceID, heats)
}

func (s *Server) maybeRegenerate(ctx context.Context, raceID int64, reason string) {
	race, err := s.DB.GetRace(ctx, raceID)
	if err != nil {
		return
	}
	if race.FinalizedAt == nil {
		return
	}
	if err := s.regenerateSchedule(ctx, raceID); err != nil {
		s.Audit.Log("regenerate_failed", map[string]interface{}{"race_id": raceID, "error": err.Error()})
		return
	}
	s.Audit.Log("schedule_regenerated", map[string]interface{}{"race_id": raceID, "reason": reason})
	s.Hub.Broadcast("schedule_changed", map[string]interface{}{"race_id": raceID, "reason": reason})
}

func (s *Server) broadcastState(ctx context.Context, raceID int64) {
	state, err := s.collectState(ctx, raceID)
	if err != nil {
		return
	}
	s.Hub.Broadcast("state", state)
}

func (s *Server) collectState(ctx context.Context, raceID int64) (map[string]interface{}, error) {
	race, err := s.DB.GetRace(ctx, raceID)
	if err != nil {
		return nil, err
	}
	cars, _ := s.DB.ListCars(ctx, raceID)
	heats, _ := s.DB.ListHeats(ctx, raceID)
	standings, _ := s.DB.Standings(ctx, raceID)
	current, onDeck, _ := s.DB.CurrentAndOnDeck(ctx, raceID)
	return map[string]interface{}{
		"race":      race,
		"cars":      cars,
		"heats":     heats,
		"standings": standings,
		"current":   current,
		"on_deck":   onDeck,
		"ties":      db.Ties(standings, 3),
	}, nil
}

func compareStandings(want, got []db.Standing) []string {
	wantByCar := map[int]db.Standing{}
	for _, s := range want {
		wantByCar[s.CarNumber] = s
	}
	var msgs []string
	for _, g := range got {
		w, ok := wantByCar[g.CarNumber]
		if !ok {
			continue
		}
		if w.Total != g.Total {
			msgs = append(msgs, fmt.Sprintf("car %d: file total=%d, computed=%d", g.CarNumber, w.Total, g.Total))
		}
	}
	return msgs
}

func sanitize(name string) string {
	out := strings.Builder{}
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == ' ' {
			out.WriteRune(r)
		} else {
			out.WriteRune('_')
		}
	}
	return out.String()
}

func writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func httpErr(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// spaHandler serves files from fs.FS, falling back to index.html for SPA routing.
type spaHandler struct {
	fs fs.FS
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := strings.TrimPrefix(r.URL.Path, "/")
	if p == "" {
		p = "index.html"
	}
	if _, err := fs.Stat(h.fs, p); err != nil {
		// fall back to index.html for SPA routes
		p = "index.html"
	}
	f, err := h.fs.Open(p)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	stat, _ := f.Stat()
	rs, ok := f.(interface {
		Read([]byte) (int, error)
		Seek(int64, int) (int64, error)
	})
	if !ok {
		// Fallback: read all
		http.ServeContent(w, r, p, stat.ModTime(), readSeekerWrap(f))
		return
	}
	http.ServeContent(w, r, p, stat.ModTime(), rs)
}

// readSeekerWrap defensively handles fs.File that don't implement io.Seeker.
type readSeekerStub struct {
	data []byte
	pos  int64
}

func readSeekerWrap(f fs.File) *readSeekerStub {
	stat, _ := f.Stat()
	buf := make([]byte, stat.Size())
	io_ReadFull(f, buf)
	return &readSeekerStub{data: buf}
}

func io_ReadFull(f fs.File, buf []byte) {
	off := 0
	for off < len(buf) {
		n, err := f.Read(buf[off:])
		off += n
		if err != nil {
			return
		}
	}
}

func (r *readSeekerStub) Read(p []byte) (int, error) {
	if r.pos >= int64(len(r.data)) {
		return 0, fmt.Errorf("EOF")
	}
	n := copy(p, r.data[r.pos:])
	r.pos += int64(n)
	return n, nil
}

func (r *readSeekerStub) Seek(off int64, whence int) (int64, error) {
	switch whence {
	case 0:
		r.pos = off
	case 1:
		r.pos += off
	case 2:
		r.pos = int64(len(r.data)) + off
	}
	return r.pos, nil
}
