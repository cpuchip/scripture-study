package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestParsePtsTimes(t *testing.T) {
	// Representative ffmpeg showinfo output (one line per emitted frame).
	showinfo := `frame=    1 fps=0.0 q=-0.0 size=N/A time=00:00:05.12
[Parsed_showinfo_1 @ 0x55] n:0 pts:128 pts_time:5.12 pos:1 fmt:rgb24
[Parsed_showinfo_1 @ 0x55] n:1 pts:256 pts_time:10.24 pos:2 fmt:rgb24
[Parsed_showinfo_1 @ 0x55] n:2 pts:512 pts_time:20 pos:3 fmt:rgb24
`
	got := parsePtsTimes(showinfo)
	want := []float64{5.12, 10.24, 20}
	if len(got) != len(want) {
		t.Fatalf("parsePtsTimes returned %d, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("time[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestHumanSize(t *testing.T) {
	cases := map[int64]string{
		512:        "512 B",
		1024:       "1.0 KB",
		1536:       "1.5 KB",
		1048576:    "1.0 MB",
		1572864:    "1.5 MB",
		1073741824: "1.0 GB",
	}
	for n, want := range cases {
		if got := humanSize(n); got != want {
			t.Errorf("humanSize(%d) = %q, want %q", n, got, want)
		}
	}
}

// capFrames should evenly sample down to the cap AND delete the dropped PNGs.
func TestCapFramesEvenSampleAndDelete(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "frames"), 0o755); err != nil {
		t.Fatal(err)
	}
	var frames []Frame
	for i := range 10 {
		name := "frames/" + fmt.Sprintf("scene-%04d.png", i)
		if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(name)), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		frames = append(frames, Frame{Sec: i, File: name})
	}

	kept := capFrames(frames, dir, 4)
	if len(kept) != 4 {
		t.Fatalf("kept %d frames, want 4", len(kept))
	}

	keptSet := map[string]bool{}
	for _, f := range kept {
		keptSet[f.File] = true
	}
	for _, f := range frames {
		_, err := os.Stat(filepath.Join(dir, filepath.FromSlash(f.File)))
		if keptSet[f.File] && err != nil {
			t.Errorf("kept frame %s should still exist on disk", f.File)
		}
		if !keptSet[f.File] && err == nil {
			t.Errorf("dropped frame %s should have been deleted", f.File)
		}
	}
}

// Under the cap, capFrames is a no-op (keeps everything).
func TestCapFramesNoOp(t *testing.T) {
	frames := []Frame{{Sec: 1}, {Sec: 2}, {Sec: 3}}
	if got := capFrames(frames, t.TempDir(), 200); len(got) != 3 {
		t.Fatalf("capFrames under cap changed length to %d, want 3", len(got))
	}
}
