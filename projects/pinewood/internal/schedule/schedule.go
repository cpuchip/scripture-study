// Package schedule generates fair pinewood derby heat charts.
//
// Goals (priority order):
//  1. Each car runs RunsPerCar heats total.
//  2. Each car runs each lane an equal number of times.
//  3. No car races back-to-back when avoidable (MinGap).
//  4. Opponent matchups spread as evenly as possible.
//
// For N=25 / runs=6 / lanes=3 the chart matches the lane and opponent-balance
// properties verified against the 2025 Marshfield Excel.
package schedule

import (
	"fmt"
	"math/rand"
	"sort"
)

// Heat is a single heat: lane index → car number (0 = empty lane).
type Heat []int

// Chart is the full schedule for a race.
type Chart struct {
	Lanes      int
	RunsPerCar int
	Heats      []Heat
}

// Options controls generation.
type Options struct {
	Lanes      int
	RunsPerCar int
	MinGap     int   // minimum heats between same car (0 = back-to-back ok)
	Seed       int64 // 0 ⇒ default 42
}

// Default returns the standard main-race options.
func Default() Options {
	return Options{Lanes: 3, RunsPerCar: 6, MinGap: 1, Seed: 42}
}

// Generate produces a chart for the given car numbers.
func Generate(cars []int, opt Options) (*Chart, error) {
	if opt.Lanes < 2 {
		return nil, fmt.Errorf("lanes must be >= 2 (got %d)", opt.Lanes)
	}
	if opt.RunsPerCar < 1 {
		return nil, fmt.Errorf("runsPerCar must be >= 1")
	}
	n := len(cars)
	if n < opt.Lanes {
		return nil, fmt.Errorf("need at least %d cars for %d lanes (got %d)", opt.Lanes, opt.Lanes, n)
	}
	if opt.Seed == 0 {
		opt.Seed = 42
	}
	totalSlots := n * opt.RunsPerCar
	totalHeats := (totalSlots + opt.Lanes - 1) / opt.Lanes

	for gap := opt.MinGap; gap >= 0; gap-- {
		o := opt
		o.MinGap = gap
		if ch, err := tryGenerate(cars, o, totalHeats); err == nil {
			return ch, nil
		}
	}
	return nil, fmt.Errorf("schedule generator failed for %d cars / %d lanes / %d runs",
		n, opt.Lanes, opt.RunsPerCar)
}

func tryGenerate(cars []int, opt Options, totalHeats int) (*Chart, error) {
	rng := rand.New(rand.NewSource(opt.Seed))
	n := len(cars)

	idx := make(map[int]int, n)
	for i, c := range cars {
		idx[c] = i
	}

	runs := make([]int, n)
	laneCount := make([][]int, n)
	opponent := make([][]int, n)
	lastHeat := make([]int, n)
	for i := range laneCount {
		laneCount[i] = make([]int, opt.Lanes)
		opponent[i] = make([]int, n)
		lastHeat[i] = -1000
	}

	heats := make([]Heat, 0, totalHeats)

	for h := 0; h < totalHeats; h++ {
		heat := make(Heat, opt.Lanes)
		taken := make(map[int]bool)

		laneOrder := make([]int, opt.Lanes)
		for i := range laneOrder {
			laneOrder[i] = (i + h) % opt.Lanes
		}

		for _, lane := range laneOrder {
			ci := pickCar(n, lane, h, opt, taken, runs, laneCount, opponent, lastHeat, heat, idx, rng)
			if ci < 0 {
				heat[lane] = 0
				continue
			}
			heat[lane] = cars[ci]
			taken[ci] = true
		}

		for lane, carNum := range heat {
			if carNum == 0 {
				continue
			}
			ci := idx[carNum]
			runs[ci]++
			laneCount[ci][lane]++
			lastHeat[ci] = h
			for olane, ocarNum := range heat {
				if olane == lane || ocarNum == 0 {
					continue
				}
				opponent[ci][idx[ocarNum]]++
			}
		}

		heats = append(heats, heat)
	}

	for i, c := range cars {
		if runs[i] != opt.RunsPerCar {
			return nil, fmt.Errorf("car %d got %d runs, expected %d", c, runs[i], opt.RunsPerCar)
		}
	}

	return &Chart{Lanes: opt.Lanes, RunsPerCar: opt.RunsPerCar, Heats: heats}, nil
}

func pickCar(
	n, lane, h int,
	opt Options,
	taken map[int]bool,
	runs []int,
	laneCount [][]int,
	opponent [][]int,
	lastHeat []int,
	heat Heat,
	idx map[int]int,
	rng *rand.Rand,
) int {
	type cand struct {
		i     int
		score int
	}
	var pool []cand
	for i := 0; i < n; i++ {
		if taken[i] || runs[i] >= opt.RunsPerCar {
			continue
		}
		if h-lastHeat[i] <= opt.MinGap {
			continue
		}
		score := 0
		score += laneCount[i][lane] * 1000
		score += runs[i] * 50
		for olane, ocarNum := range heat {
			if olane == lane || ocarNum == 0 {
				continue
			}
			score += opponent[i][idx[ocarNum]] * 10
		}
		score -= (h - lastHeat[i])
		pool = append(pool, cand{i, score})
	}
	if len(pool) == 0 {
		return -1
	}
	sort.SliceStable(pool, func(a, b int) bool { return pool[a].score < pool[b].score })
	bestScore := pool[0].score
	tieEnd := 0
	for tieEnd < len(pool) && pool[tieEnd].score == bestScore {
		tieEnd++
	}
	return pool[rng.Intn(tieEnd)].i
}

// Stats summarizes fairness properties of a chart.
type Stats struct {
	TotalHeats    int
	RunsPerCar    map[int]int
	LaneCounts    map[int][]int
	OpponentPairs map[[2]int]int
	UniquePairs   int
	PairCountDist map[int]int
	MinGap        int
	MaxGap        int
	AvgGap        float64
}

// Analyze returns fairness stats for a chart.
func Analyze(ch *Chart) Stats {
	s := Stats{
		TotalHeats:    len(ch.Heats),
		RunsPerCar:    map[int]int{},
		LaneCounts:    map[int][]int{},
		OpponentPairs: map[[2]int]int{},
		PairCountDist: map[int]int{},
		MinGap:        1 << 30,
	}
	last := map[int]int{}
	gaps := []int{}
	for h, heat := range ch.Heats {
		for lane, car := range heat {
			if car == 0 {
				continue
			}
			s.RunsPerCar[car]++
			if _, ok := s.LaneCounts[car]; !ok {
				s.LaneCounts[car] = make([]int, ch.Lanes)
			}
			s.LaneCounts[car][lane]++
			if prev, ok := last[car]; ok {
				gaps = append(gaps, h-prev)
			}
			last[car] = h
		}
		for i := 0; i < len(heat); i++ {
			for j := i + 1; j < len(heat); j++ {
				if heat[i] == 0 || heat[j] == 0 {
					continue
				}
				a, b := heat[i], heat[j]
				if a > b {
					a, b = b, a
				}
				s.OpponentPairs[[2]int{a, b}]++
			}
		}
	}
	s.UniquePairs = len(s.OpponentPairs)
	for _, c := range s.OpponentPairs {
		s.PairCountDist[c]++
	}
	if len(gaps) == 0 {
		s.MinGap, s.MaxGap = 0, 0
	} else {
		sum := 0
		s.MaxGap = 0
		for _, g := range gaps {
			if g < s.MinGap {
				s.MinGap = g
			}
			if g > s.MaxGap {
				s.MaxGap = g
			}
			sum += g
		}
		s.AvgGap = float64(sum) / float64(len(gaps))
	}
	return s
}
