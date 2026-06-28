package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Slide is a frame aligned to the transcript narration spoken over it (and a
// chapter title, when the description carries chapter markers).
type Slide struct {
	Sec       int    `json:"sec"`
	File      string `json:"file"`
	TLink     string `json:"t_link"`
	Title     string `json:"title,omitempty"`     // chapter title covering this moment
	Narration string `json:"narration,omitempty"` // transcript spoken until the next slide
}

// Chapter is a timestamped section parsed from a video description.
type Chapter struct {
	Sec   int
	Title string
}

// chapterLineRe matches "1:16 Title", "12:03 Title", or "1:02:55 Title" at the
// start of a line (YouTube's chapter convention).
var chapterLineRe = regexp.MustCompile(`^\s*(?:(\d{1,2}):)?(\d{1,2}):(\d{2})\s+(\S.*?)\s*$`)

// parseChapters extracts chapter markers (timestamp + title) from a description.
// Returns nil unless there are at least 3 (so a stray timestamp isn't mistaken
// for a chapter list).
func parseChapters(description string) []Chapter {
	var chapters []Chapter
	for line := range strings.SplitSeq(description, "\n") {
		m := chapterLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		h, _ := strconv.Atoi(m[1]) // empty hours group → 0
		minutes, _ := strconv.Atoi(m[2])
		secs, _ := strconv.Atoi(m[3])
		chapters = append(chapters, Chapter{Sec: h*3600 + minutes*60 + secs, Title: m[4]})
	}
	if len(chapters) < 3 {
		return nil
	}
	sort.Slice(chapters, func(i, j int) bool { return chapters[i].Sec < chapters[j].Sec })
	return chapters
}

// LoadCues reads cues.json from a video dir (nil if absent or unparseable).
func LoadCues(videoDir string) []Cue {
	b, err := os.ReadFile(filepath.Join(videoDir, "cues.json"))
	if err != nil {
		return nil
	}
	var cues []Cue
	if json.Unmarshal(b, &cues) != nil {
		return nil
	}
	return cues
}

// SlideOptions configures BuildSlides. Capture mode is chosen automatically:
// chapter markers → scene-change → interval fallback.
type SlideOptions struct {
	SceneThreshold float64
	MaxFrames      int
}

// BuildSlides is the one-shot study path: pick the best capture strategy for the
// video, extract frames, and align each to the transcript narration (and chapter
// titles). Returns the slides plus the capture mode used. The video must already
// be downloaded.
func BuildSlides(cfg *Config, videoDir, webURL string, meta *VideoMetadata, opts SlideOptions) ([]Slide, string, error) {
	if opts.MaxFrames <= 0 {
		opts.MaxFrames = 60
	}
	// Start the one-shot from a clean frames dir so a prior capture (or a
	// scene→interval fallback) doesn't leave orphan PNGs behind.
	removeGlob(filepath.Join(videoDir, "frames", "*.png"))

	chapters := parseChapters(meta.Description)
	var frames []Frame
	var err error
	var mode string

	if len(chapters) >= 3 {
		// Chapters ARE the slide boundaries — the best signal when present.
		mode = "chapters"
		ts := make([]int, 0, len(chapters))
		for _, c := range chapters {
			ts = append(ts, c.Sec)
		}
		frames, err = ExtractFrames(cfg, videoDir, webURL,
			FrameOptions{Mode: "timestamps", Timestamps: ts, MaxFrames: opts.MaxFrames})
	} else {
		mode = "scene"
		frames, err = ExtractFrames(cfg, videoDir, webURL,
			FrameOptions{Mode: "scene", SceneThreshold: opts.SceneThreshold, MaxFrames: opts.MaxFrames})
		// Smooth-scroll screen-shares (Excalidraw, scrolling a page) defeat
		// scene-change detection. If it came back sparse (<1 frame / 2 min on a
		// video over 4 min), fall back to even interval sampling.
		if err == nil && meta.Duration > 240 && len(frames) < meta.Duration/120 {
			mode = "interval"
			frames, err = ExtractFrames(cfg, videoDir, webURL,
				FrameOptions{Mode: "interval", EverySec: pickInterval(meta.Duration, opts.MaxFrames), MaxFrames: opts.MaxFrames})
		}
	}
	if err != nil {
		return nil, mode, err
	}

	return alignSlides(frames, chapters, LoadCues(videoDir)), mode, nil
}

// pickInterval chooses an interval (seconds) to land roughly maxFrames frames,
// never finer than 30s.
func pickInterval(duration, maxFrames int) int {
	if maxFrames <= 0 {
		maxFrames = 60
	}
	return max(duration/maxFrames, 30)
}

// alignSlides attaches to each frame the narration spoken until the next frame
// and the chapter title covering its timestamp.
func alignSlides(frames []Frame, chapters []Chapter, cues []Cue) []Slide {
	slides := make([]Slide, 0, len(frames))
	for i, f := range frames {
		end := 1 << 30
		if i+1 < len(frames) {
			end = frames[i+1].Sec
		}
		slides = append(slides, Slide{
			Sec:       f.Sec,
			File:      f.File,
			TLink:     f.TLink,
			Title:     chapterTitleAt(chapters, f.Sec),
			Narration: narrationBetween(cues, f.Sec, end),
		})
	}
	return slides
}

// chapterTitleAt returns the title of the last chapter starting at or before sec.
func chapterTitleAt(chapters []Chapter, sec int) string {
	title := ""
	for _, c := range chapters {
		if c.Sec <= sec {
			title = c.Title
		} else {
			break
		}
	}
	return title
}

// narrationBetween concatenates the text of every cue starting in [start, end).
func narrationBetween(cues []Cue, start, end int) string {
	var b strings.Builder
	for _, c := range cues {
		if int(c.Begin) >= start && int(c.Begin) < end {
			if b.Len() > 0 {
				b.WriteByte(' ')
			}
			b.WriteString(c.Text)
		}
	}
	return strings.TrimSpace(b.String())
}

// WriteSlidesDoc writes a readable slides.md interleaving each slide image with
// the narration spoken over it. Returns the path written.
func WriteSlidesDoc(videoDir string, meta *VideoMetadata, slides []Slide, mode string) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "# Slides — %s\n\n", meta.Title)
	fmt.Fprintf(&b, "%s · %s · [watch](%s)\n\ncapture mode: %s · %d slides\n\n",
		meta.Channel, formatDuration(meta.Duration), meta.URL, mode, len(slides))
	for i, sl := range slides {
		head := formatDuration(sl.Sec)
		if sl.Title != "" {
			head = fmt.Sprintf("%s — %s", head, sl.Title)
		}
		fmt.Fprintf(&b, "## %d. [%s](%s)\n\n![slide %d](%s)\n\n", i+1, head, sl.TLink, i+1, sl.File)
		if sl.Narration != "" {
			fmt.Fprintf(&b, "> %s\n\n", sl.Narration)
		}
	}
	path := filepath.Join(videoDir, "slides.md")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", err
	}
	return path, nil
}
