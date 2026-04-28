package schedule

import "testing"

func TestGenerate25(t *testing.T) {
	cars := make([]int, 25)
	for i := range cars {
		cars[i] = i + 1
	}
	ch, err := Generate(cars, Default())
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(ch.Heats) != 50 {
		t.Errorf("expected 50 heats, got %d", len(ch.Heats))
	}
	st := Analyze(ch)
	for c, r := range st.RunsPerCar {
		if r != 6 {
			t.Errorf("car %d ran %d times, want 6", c, r)
		}
	}
	for c, lanes := range st.LaneCounts {
		for l, count := range lanes {
			if count != 2 {
				t.Errorf("car %d lane %d: %d runs, want 2", c, l+1, count)
			}
		}
	}
	if st.MinGap < 1 {
		t.Errorf("min gap was %d (want >=1)", st.MinGap)
	}
	t.Logf("N=25 stats: heats=%d uniquePairs=%d pairDist=%v gap min/max/avg=%d/%d/%.2f",
		st.TotalHeats, st.UniquePairs, st.PairCountDist, st.MinGap, st.MaxGap, st.AvgGap)
}

func TestGenerateSmall(t *testing.T) {
	for n := 4; n <= 12; n++ {
		cars := make([]int, n)
		for i := range cars {
			cars[i] = i + 1
		}
		ch, err := Generate(cars, Default())
		if err != nil {
			t.Errorf("N=%d: %v", n, err)
			continue
		}
		st := Analyze(ch)
		for c, r := range st.RunsPerCar {
			if r != 6 {
				t.Errorf("N=%d car %d ran %d times", n, c, r)
			}
		}
	}
}

func TestRunoff(t *testing.T) {
	cases := []struct {
		cars       []int
		wantLanes  int
		wantHeats  int
		wantPerCar int
	}{
		{[]int{5, 17}, 2, 2, 2},
		{[]int{5, 17, 22}, 3, 3, 3},
		{[]int{5, 17, 22, 1}, 3, 4, 3},
	}
	for _, c := range cases {
		ch, err := GenerateRunoff(c.cars)
		if err != nil {
			t.Errorf("cars=%v: %v", c.cars, err)
			continue
		}
		if ch.Lanes != c.wantLanes {
			t.Errorf("cars=%v lanes=%d want %d", c.cars, ch.Lanes, c.wantLanes)
		}
		if len(ch.Heats) != c.wantHeats {
			t.Errorf("cars=%v heats=%d want %d", c.cars, len(ch.Heats), c.wantHeats)
		}
		st := Analyze(ch)
		for car, runs := range st.RunsPerCar {
			if runs != c.wantPerCar {
				t.Errorf("cars=%v car %d runs=%d want %d", c.cars, car, runs, c.wantPerCar)
			}
		}
	}
}
