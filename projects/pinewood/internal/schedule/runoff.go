package schedule

import "fmt"

// RunoffOptions describes the per-tier run-off format.
type RunoffOptions struct {
	Lanes      int
	RunsPerCar int
}

// Runoff returns options + adjusted lane count for a given car-count.
//
// Rules per proposal §5.3:
//   2 cars  → 2 lanes (lane 3 idle), 1 run per lane each = 2 heats, 2 runs/car
//   3 cars  → 3 lanes, 1 run per lane each = 3 heats, 3 runs/car
//   4+ cars → 3 lanes, 3 runs/car (solver)
func RunoffPlan(numCars int) (lanes, runs int, err error) {
	switch {
	case numCars < 2:
		return 0, 0, fmt.Errorf("need >= 2 cars for runoff")
	case numCars == 2:
		return 2, 2, nil
	case numCars == 3:
		return 3, 3, nil
	default:
		return 3, 3, nil
	}
}

// GenerateRunoff builds a schedule for the supplied car list using the
// run-off plan rules.
func GenerateRunoff(cars []int) (*Chart, error) {
	lanes, runs, err := RunoffPlan(len(cars))
	if err != nil {
		return nil, err
	}
	return Generate(cars, Options{Lanes: lanes, RunsPerCar: runs, MinGap: 0, Seed: 7})
}
