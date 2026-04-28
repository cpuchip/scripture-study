// Package excel handles xlsx export/import in the legacy 4-tab format.
package excel

import (
	"fmt"
	"io"
	"sort"

	"github.com/cpuchip/pinewood/internal/db"
	"github.com/xuri/excelize/v2"
)

// Export writes the race to an xlsx with tabs: Heats, Scores, Results, Cars.
func Export(w io.Writer, race *db.Race, cars []db.Car, heats []db.Heat, standings []db.Standing) error {
	f := excelize.NewFile()
	defer f.Close()

	// Sort heats by number for export.
	sort.Slice(heats, func(i, j int) bool { return heats[i].HeatNumber < heats[j].HeatNumber })

	// Heats sheet
	if _, err := f.NewSheet("Heats"); err != nil {
		return err
	}
	f.SetCellValue("Heats", "A1", "Heat")
	f.SetCellValue("Heats", "B1", "Lane 1")
	f.SetCellValue("Heats", "C1", "Lane 2")
	f.SetCellValue("Heats", "D1", "Lane 3")
	for i, h := range heats {
		row := i + 2
		f.SetCellValue("Heats", fmt.Sprintf("A%d", row), h.HeatNumber)
		for _, s := range h.Slots {
			col := string(rune('A' + s.Lane))
			f.SetCellValue("Heats", fmt.Sprintf("%s%d", col, row), s.CarNumber)
		}
	}

	// Scores sheet
	if _, err := f.NewSheet("Scores"); err != nil {
		return err
	}
	cols := []string{"Heat", "Lane 1", "Lane 2", "Lane 3", "Score 1", "Score 2", "Score 3"}
	for i, c := range cols {
		f.SetCellValue("Scores", fmt.Sprintf("%s1", string(rune('A'+i))), c)
	}
	for i, h := range heats {
		row := i + 2
		f.SetCellValue("Scores", fmt.Sprintf("A%d", row), h.HeatNumber)
		for _, s := range h.Slots {
			carCol := string(rune('A' + s.Lane))
			scoreCol := string(rune('A' + 3 + s.Lane))
			f.SetCellValue("Scores", fmt.Sprintf("%s%d", carCol, row), s.CarNumber)
			if s.Place != nil {
				f.SetCellValue("Scores", fmt.Sprintf("%s%d", scoreCol, row), *s.Place)
			}
		}
	}

	// Results sheet
	if _, err := f.NewSheet("Results"); err != nil {
		return err
	}
	f.SetCellValue("Results", "A1", "Car Number")
	f.SetCellValue("Results", "B1", "Total Score")
	f.SetCellValue("Results", "C1", "Rank")
	for i, s := range standings {
		row := i + 2
		f.SetCellValue("Results", fmt.Sprintf("A%d", row), s.CarNumber)
		f.SetCellValue("Results", fmt.Sprintf("B%d", row), s.Total)
		f.SetCellValue("Results", fmt.Sprintf("C%d", row), s.Rank)
	}

	// Cars sheet (round-trip)
	if _, err := f.NewSheet("Cars"); err != nil {
		return err
	}
	f.SetCellValue("Cars", "A1", "Number")
	f.SetCellValue("Cars", "B1", "Name")
	for i, c := range cars {
		row := i + 2
		f.SetCellValue("Cars", fmt.Sprintf("A%d", row), c.Number)
		f.SetCellValue("Cars", fmt.Sprintf("B%d", row), c.Name)
	}

	// Drop default sheet
	if err := f.DeleteSheet("Sheet1"); err != nil {
		// ignore if not present
	}
	f.SetActiveSheet(0)

	return f.Write(w)
}

// ImportData is the parsed payload ready for DB insertion.
type ImportData struct {
	RaceName  string
	LaneCount int
	Cars      []db.Car // RaceID unset
	Heats     [][]int  // heat → lane → carNumber
	Scores    [][]*int // heat → lane → place pointer (nil = not scored)
	Standings []db.Standing
	Warnings  []string
}

// Parse reads an xlsx produced by Export and returns ImportData.
// Tolerant: missing tabs are best-effort.
func Parse(r io.Reader, name string) (*ImportData, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out := &ImportData{RaceName: name, LaneCount: 3}

	// Cars
	if rows, err := f.GetRows("Cars"); err == nil && len(rows) > 1 {
		for _, row := range rows[1:] {
			if len(row) == 0 || row[0] == "" {
				continue
			}
			var num int
			fmt.Sscanf(row[0], "%d", &num)
			nm := ""
			if len(row) > 1 {
				nm = row[1]
			}
			out.Cars = append(out.Cars, db.Car{Number: num, Name: nm})
		}
	}

	// Scores (preferred over Heats since it has place data too)
	scoresRows, err := f.GetRows("Scores")
	if err != nil {
		return nil, fmt.Errorf("Scores tab missing: %w", err)
	}
	if len(scoresRows) < 2 {
		return nil, fmt.Errorf("Scores tab empty")
	}
	for _, row := range scoresRows[1:] {
		if len(row) == 0 || row[0] == "" {
			continue
		}
		heat := make([]int, 3)
		score := make([]*int, 3)
		for i := 0; i < 3; i++ {
			if len(row) > 1+i && row[1+i] != "" {
				var n int
				fmt.Sscanf(row[1+i], "%d", &n)
				heat[i] = n
			}
			if len(row) > 4+i && row[4+i] != "" {
				var p int
				if _, e := fmt.Sscanf(row[4+i], "%d", &p); e == nil && p >= 1 && p <= 3 {
					score[i] = &p
				} else {
					out.Warnings = append(out.Warnings,
						fmt.Sprintf("heat %s lane %d: place '%s' is not 1/2/3, skipped", row[0], i+1, row[4+i]))
				}
			}
		}
		out.Heats = append(out.Heats, heat)
		out.Scores = append(out.Scores, score)
	}

	// If Cars tab was missing, derive from Heats.
	if len(out.Cars) == 0 {
		seen := map[int]bool{}
		var nums []int
		for _, h := range out.Heats {
			for _, c := range h {
				if c != 0 && !seen[c] {
					seen[c] = true
					nums = append(nums, c)
				}
			}
		}
		sort.Ints(nums)
		for _, n := range nums {
			out.Cars = append(out.Cars, db.Car{Number: n})
		}
	}

	// Optional Results tab for warning checks
	if rows, err := f.GetRows("Results"); err == nil && len(rows) > 1 {
		for _, row := range rows[1:] {
			if len(row) < 2 || row[0] == "" {
				continue
			}
			s := db.Standing{}
			fmt.Sscanf(row[0], "%d", &s.CarNumber)
			fmt.Sscanf(row[1], "%d", &s.Total)
			if len(row) > 2 {
				fmt.Sscanf(row[2], "%d", &s.Rank)
			}
			out.Standings = append(out.Standings, s)
		}
	}

	return out, nil
}
