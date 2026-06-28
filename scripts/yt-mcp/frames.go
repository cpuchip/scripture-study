package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

// ── Video file download ──────────────────────────────────────────────────────

// DownloadVideoFile downloads the actual video (resolution-capped) into
// yt/{channelSlug}/{videoID}/video.<ext>. Big and optional — this is a separate,
// explicitly-invoked path; the transcript-only yt_download never calls it.
// Returns the path to the downloaded video file and the video metadata.
func DownloadVideoFile(cfg *Config, rawURL string, force bool, maxHeight int, cookieOverride string) (string, *VideoMetadata, error) {
	videoID, err := ExtractVideoID(rawURL)
	if err != nil {
		return "", nil, err
	}
	videoURL := CanonicalURL(videoID)

	meta, err := FetchMetadata(cfg, videoURL, cookieOverride)
	if err != nil {
		return "", nil, fmt.Errorf("fetching metadata: %w", err)
	}

	outputDir := filepath.Join(cfg.YTDir, meta.ChannelSlug, videoID)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", nil, fmt.Errorf("creating output dir: %w", err)
	}

	if !force {
		if existing := findVideoFile(outputDir); existing != "" {
			writeMetadataIfAbsent(outputDir, meta)
			return existing, meta, nil
		}
	}

	if maxHeight <= 0 {
		maxHeight = cfg.MaxVideoHeight
	}
	if maxHeight <= 0 {
		maxHeight = 720
	}

	format := fmt.Sprintf("bestvideo[height<=%d]+bestaudio/best[height<=%d]/best", maxHeight, maxHeight)
	outTemplate := filepath.Join(outputDir, "video.%(ext)s")
	args := []string{
		"-f", format,
		"--merge-output-format", "mp4",
		"-o", outTemplate,
	}
	// Point yt-dlp at the real ffmpeg (overriding any stale --ffmpeg-location in
	// the user's yt-dlp config) so the video+audio merge actually runs.
	args = append(args, ffmpegLocationArgs(cfg)...)
	args = append(args, cookieArgs(cfg, cookieOverride)...)
	args = append(args, videoURL)

	cmd := exec.Command(cfg.YtDlpPath, args...)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", nil, fmt.Errorf("yt-dlp video download failed: %w", err)
	}

	videoPath := findVideoFile(outputDir)
	if videoPath == "" {
		return "", nil, fmt.Errorf("video download produced no video file in %s", outputDir)
	}
	writeMetadataIfAbsent(outputDir, meta)
	return videoPath, meta, nil
}

// ffmpegLocationArgs tells yt-dlp where ffmpeg lives, overriding any stale
// --ffmpeg-location in the user's yt-dlp config (a common breakage that fails the
// video+audio merge even when ffmpeg is on PATH).
func ffmpegLocationArgs(cfg *Config) []string {
	if p, err := exec.LookPath(cfg.FfmpegPath); err == nil {
		return []string{"--ffmpeg-location", filepath.Dir(p)}
	}
	return nil
}

// findVideoFile returns the path to a video.* file in dir (mp4/mkv/webm), or "".
func findVideoFile(dir string) string {
	for _, ext := range []string{"mp4", "mkv", "webm"} {
		p := filepath.Join(dir, "video."+ext)
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p
		}
	}
	return ""
}

func writeMetadataIfAbsent(dir string, meta *VideoMetadata) {
	metaPath := filepath.Join(dir, "metadata.json")
	if _, err := os.Stat(metaPath); err == nil {
		return
	}
	if b, err := json.MarshalIndent(meta, "", "  "); err == nil {
		_ = os.WriteFile(metaPath, b, 0o644)
	}
}

// ── Frame extraction ─────────────────────────────────────────────────────────

// FrameOptions configures ExtractFrames.
type FrameOptions struct {
	Mode           string  // "scene" (default) | "interval" | "timestamps"
	SceneThreshold float64 // for "scene"; default 0.4
	EverySec       int     // for "interval"; default 30
	Timestamps     []int   // for "timestamps"; seconds
	MaxFrames      int     // cap; default 200
}

// ExtractFrames extracts slide frames from a previously-downloaded video into
// {videoDir}/frames/ and writes a timestamp-aligned frames.json manifest. It
// returns the manifest — NOT the image bytes; the caller reads the specific PNGs
// it wants (or a vision stage pages them in by handle).
func ExtractFrames(cfg *Config, videoDir, webURL string, opts FrameOptions) ([]Frame, error) {
	videoPath := findVideoFile(videoDir)
	if videoPath == "" {
		return nil, fmt.Errorf("no video file in %s — run yt_download_video first", videoDir)
	}
	framesDir := filepath.Join(videoDir, "frames")
	if err := os.MkdirAll(framesDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating frames dir: %w", err)
	}
	if opts.MaxFrames <= 0 {
		opts.MaxFrames = 200
	}

	var frames []Frame
	var err error
	switch opts.Mode {
	case "", "scene":
		thr := opts.SceneThreshold
		if thr <= 0 {
			thr = 0.4
		}
		frames, err = extractByFilter(cfg, videoPath, framesDir, webURL, "scene",
			fmt.Sprintf("select='gt(scene,%g)'", thr))
	case "interval":
		n := opts.EverySec
		if n <= 0 {
			n = 30
		}
		frames, err = extractByFilter(cfg, videoPath, framesDir, webURL, "frame",
			fmt.Sprintf("fps=1/%d", n))
	case "timestamps":
		frames, err = extractAtTimestamps(cfg, videoPath, framesDir, webURL, opts.Timestamps)
	default:
		return nil, fmt.Errorf("unknown frame mode %q (use scene | interval | timestamps)", opts.Mode)
	}
	if err != nil {
		return nil, err
	}

	frames = capFrames(frames, videoDir, opts.MaxFrames)

	if err := writeFramesManifest(videoDir, frames); err != nil {
		return frames, fmt.Errorf("writing frames.json: %w", err)
	}
	return frames, nil
}

// extractByFilter runs one ffmpeg pass with a select/fps filter + showinfo, then
// aligns the emitted PNGs to the pts_time the filter reported for each (in order).
// We parse showinfo from stderr rather than the metadata=print filter, to dodge
// ffmpeg's filter-arg path escaping (Windows drive colons / backslashes).
func extractByFilter(cfg *Config, videoPath, framesDir, webURL, prefix, selectFilter string) ([]Frame, error) {
	removeGlob(filepath.Join(framesDir, prefix+"-*.png"))
	pattern := filepath.Join(framesDir, prefix+"-%04d.png")

	vf := selectFilter + ",showinfo"
	args := []string{"-i", videoPath, "-vf", vf, "-fps_mode", "vfr", "-y", pattern}

	cmd := exec.Command(cfg.FfmpegPath, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg frame extraction failed: %w\n%s", err, tail(stderr.String(), 600))
	}

	times := parsePtsTimes(stderr.String())
	pngs := sortedGlob(filepath.Join(framesDir, prefix+"-*.png"))

	n := min(len(pngs), len(times))
	frames := make([]Frame, 0, n)
	for i := range n {
		sec := int(times[i] + 0.5)
		frames = append(frames, Frame{
			Sec:   sec,
			File:  "frames/" + filepath.Base(pngs[i]),
			TLink: fmt.Sprintf("%s&t=%d", webURL, sec),
		})
	}
	return frames, nil
}

func extractAtTimestamps(cfg *Config, videoPath, framesDir, webURL string, ts []int) ([]Frame, error) {
	removeGlob(filepath.Join(framesDir, "ts-*.png"))
	var frames []Frame
	for _, t := range ts {
		if t < 0 {
			continue
		}
		out := filepath.Join(framesDir, fmt.Sprintf("ts-%06d.png", t))
		// -ss before -i = fast seek; -frames:v 1 = a single frame.
		args := []string{"-ss", strconv.Itoa(t), "-i", videoPath, "-frames:v", "1", "-y", out}
		cmd := exec.Command(cfg.FfmpegPath, args...)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return frames, fmt.Errorf("ffmpeg at %ds failed: %w\n%s", t, err, tail(stderr.String(), 400))
		}
		frames = append(frames, Frame{
			Sec:   t,
			File:  "frames/" + filepath.Base(out),
			TLink: fmt.Sprintf("%s&t=%d", webURL, t),
		})
	}
	return frames, nil
}

// ── Manifest + helpers ───────────────────────────────────────────────────────

var ptsTimeRe = regexp.MustCompile(`pts_time:([0-9]+(?:\.[0-9]+)?)`)

// parsePtsTimes pulls the pts_time of each emitted frame from ffmpeg showinfo
// output, in order.
func parsePtsTimes(showinfo string) []float64 {
	matches := ptsTimeRe.FindAllStringSubmatch(showinfo, -1)
	out := make([]float64, 0, len(matches))
	for _, m := range matches {
		if v, err := strconv.ParseFloat(m[1], 64); err == nil {
			out = append(out, v)
		}
	}
	return out
}

// capFrames evenly samples frames down to max (so coverage spans the whole
// video), deleting the dropped PNGs from disk. Returns the kept manifest.
func capFrames(frames []Frame, videoDir string, max int) []Frame {
	if max <= 0 || len(frames) <= max {
		return frames
	}
	keep := make([]bool, len(frames))
	stride := float64(len(frames)) / float64(max)
	for i := range max {
		idx := int(float64(i) * stride)
		if idx >= len(frames) {
			idx = len(frames) - 1
		}
		keep[idx] = true
	}
	kept := make([]Frame, 0, max)
	for i, f := range frames {
		if keep[i] {
			kept = append(kept, f)
		} else {
			_ = os.Remove(filepath.Join(videoDir, filepath.FromSlash(f.File)))
		}
	}
	return kept
}

func writeFramesManifest(videoDir string, frames []Frame) error {
	b, err := json.MarshalIndent(frames, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(videoDir, "frames.json"), b, 0o644)
}

// LoadFrames reads frames.json from a video directory (empty if none).
func LoadFrames(videoDir string) []Frame {
	b, err := os.ReadFile(filepath.Join(videoDir, "frames.json"))
	if err != nil {
		return nil
	}
	var frames []Frame
	if json.Unmarshal(b, &frames) != nil {
		return nil
	}
	return frames
}

func sortedGlob(pattern string) []string {
	m, _ := filepath.Glob(pattern)
	sort.Strings(m)
	return m
}

func removeGlob(pattern string) {
	if m, err := filepath.Glob(pattern); err == nil {
		for _, p := range m {
			_ = os.Remove(p)
		}
	}
}

func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return "…" + s[len(s)-n:]
}
