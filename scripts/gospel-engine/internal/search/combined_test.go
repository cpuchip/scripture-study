package search

import (
	"math"
	"testing"
)

func TestRRFScore(t *testing.T) {
	// Document ranked #1 in both lists should score highest
	both1 := rrfScore(1, 1)
	// Document ranked #1 in keyword only
	kw1 := rrfScore(1, 0)
	// Document ranked #1 in semantic only
	vec1 := rrfScore(0, 1)
	// Document ranked #10 in both
	both10 := rrfScore(10, 10)

	if both1 <= kw1 {
		t.Errorf("both-#1 (%.6f) should beat kw-only-#1 (%.6f)", both1, kw1)
	}
	if both1 <= vec1 {
		t.Errorf("both-#1 (%.6f) should beat vec-only-#1 (%.6f)", both1, vec1)
	}
	if kw1 != vec1 {
		t.Errorf("kw-only-#1 (%.6f) should equal vec-only-#1 (%.6f) — symmetric", kw1, vec1)
	}
	if both10 >= both1 {
		t.Errorf("both-#10 (%.6f) should be less than both-#1 (%.6f)", both10, both1)
	}

	// Check expected values with k=60
	expectedBoth1 := 2.0 / 61.0 // 1/(60+1) + 1/(60+1)
	if math.Abs(both1-expectedBoth1) > 1e-9 {
		t.Errorf("both-#1 = %.8f, want %.8f", both1, expectedBoth1)
	}
	expectedKW1 := 1.0 / 61.0
	if math.Abs(kw1-expectedKW1) > 1e-9 {
		t.Errorf("kw-only-#1 = %.8f, want %.8f", kw1, expectedKW1)
	}

	// Absent from both lists should be 0
	absent := rrfScore(0, 0)
	if absent != 0 {
		t.Errorf("absent should be 0, got %.6f", absent)
	}
}

func TestBuildRankMap(t *testing.T) {
	results := []Result{
		{FilePath: "a.md", Reference: "ref1", Type: "paragraph"},
		{FilePath: "b.md", Reference: "ref2", Type: "paragraph"},
		{FilePath: "c.md", Reference: "ref3", Type: "paragraph"},
	}

	ranks := buildRankMap(results)
	if ranks["a.md|ref1"] != 1 {
		t.Errorf("first item should be rank 1, got %d", ranks["a.md|ref1"])
	}
	if ranks["b.md|ref2"] != 2 {
		t.Errorf("second item should be rank 2, got %d", ranks["b.md|ref2"])
	}
	if ranks["c.md|ref3"] != 3 {
		t.Errorf("third item should be rank 3, got %d", ranks["c.md|ref3"])
	}
}

func TestDocKey(t *testing.T) {
	tests := []struct {
		name string
		r    Result
		want string
	}{
		{
			name: "verse dedup extracts verse number",
			r:    Result{FilePath: "/path/nt/1-cor/13.md", Reference: "Alma 32:27", Type: "verse"},
			want: "/path/nt/1-cor/13.md|v27",
		},
		{
			name: "verse dedup works with slug format",
			r:    Result{FilePath: "/path/nt/1-cor/13.md", Reference: "1-cor 13:27", Type: "verse"},
			want: "/path/nt/1-cor/13.md|v27",
		},
		{
			name: "paragraph includes full reference",
			r:    Result{FilePath: "/path/talk.md", Reference: "Elder Oaks ¶5", Type: "paragraph"},
			want: "/path/talk.md|Elder Oaks ¶5",
		},
		{
			name: "file-level uses path only",
			r:    Result{FilePath: "/path/chapter.md", Reference: "Genesis 22", Type: "chapter"},
			want: "/path/chapter.md",
		},
		{
			name: "no filepath falls back to reference",
			r:    Result{Reference: "some ref", Type: "verse"},
			want: "some ref",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := docKey(tt.r)
			if got != tt.want {
				t.Errorf("docKey() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	results := []Result{
		{Reference: "a"},
		{Reference: "b"},
		{Reference: "c"},
	}

	// Truncate to 2
	out := truncate(results, 2)
	if len(out) != 2 {
		t.Errorf("truncate(3, 2) = %d items, want 2", len(out))
	}

	// Truncate to more than length
	out = truncate(results, 10)
	if len(out) != 3 {
		t.Errorf("truncate(3, 10) = %d items, want 3", len(out))
	}

	// Truncate nil
	out = truncate(nil, 5)
	if len(out) != 0 {
		t.Errorf("truncate(nil, 5) = %d items, want 0", len(out))
	}
}

func TestRRFRankingOrder(t *testing.T) {
	// Simulate: doc A is #1 in both, doc B is #1 in keyword #30 in semantic,
	// doc C is #5 in semantic only
	scoreA := rrfScore(1, 1)   // high in both
	scoreB := rrfScore(1, 30)  // high keyword, moderate semantic
	scoreC := rrfScore(0, 5)   // semantic only

	if scoreA <= scoreB {
		t.Errorf("A (both #1, %.6f) should beat B (kw#1+vec#30, %.6f)", scoreA, scoreB)
	}
	if scoreB <= scoreC {
		t.Errorf("B (kw#1+vec#30, %.6f) should beat C (vec-only#5, %.6f)", scoreB, scoreC)
	}
}
