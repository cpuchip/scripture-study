package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// ── Video ID Extraction ──────────────────────────────────────────────────────

var videoIDPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?:youtube\.com/watch\?.*v=|youtu\.be/|youtube\.com/shorts/|youtube\.com/embed/|youtube\.com/v/)([a-zA-Z0-9_-]{11})`),
}

// ExtractVideoID extracts the 11-character YouTube video ID from various URL formats.
func ExtractVideoID(rawURL string) (string, error) {
	for _, re := range videoIDPatterns {
		matches := re.FindStringSubmatch(rawURL)
		if len(matches) >= 2 {
			return matches[1], nil
		}
	}
	// Maybe it's already a bare video ID
	if len(rawURL) == 11 && regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(rawURL) {
		return rawURL, nil
	}
	return "", fmt.Errorf("could not extract video ID from URL: %s", rawURL)
}

// CanonicalURL returns the canonical YouTube watch URL for a video ID.
func CanonicalURL(videoID string) string {
	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
}

// ── Channel Slug ─────────────────────────────────────────────────────────────

var slugReplacer = regexp.MustCompile(`[^a-z0-9-]+`)

// ChannelSlug generates a filesystem-safe slug from a channel name.
// e.g., "Book of Mormon Central" → "book-of-mormon-central"
func ChannelSlug(channel string) string {
	s := strings.ToLower(strings.TrimSpace(channel))
	s = strings.ReplaceAll(s, " ", "-")
	s = slugReplacer.ReplaceAllString(s, "")
	// Collapse multiple hyphens
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")
	if s == "" {
		return "unknown-channel"
	}
	return s
}

// ── yt-dlp Wrapper ───────────────────────────────────────────────────────────

// ytdlpMetadata is the subset of yt-dlp --dump-json output we care about.
type ytdlpMetadata struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Channel     string `json:"channel"`
	ChannelID   string `json:"channel_id"`
	UploadDate  string `json:"upload_date"`
	Duration    int    `json:"duration"`
	Description string `json:"description"`
	WebpageURL  string `json:"webpage_url"`
}

// FetchMetadata runs yt-dlp --dump-json to get video metadata without downloading.
func FetchMetadata(cfg *Config, videoURL string) (*VideoMetadata, error) {
	cmd := exec.Command(cfg.YtDlpPath, "--dump-json", "--skip-download", videoURL)
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp --dump-json failed: %w", err)
	}

	var raw ytdlpMetadata
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("parsing yt-dlp metadata: %w", err)
	}

	webURL := raw.WebpageURL
	if webURL == "" {
		webURL = CanonicalURL(raw.ID)
	}

	return &VideoMetadata{
		ID:          raw.ID,
		Title:       raw.Title,
		Channel:     raw.Channel,
		ChannelID:   raw.ChannelID,
		UploadDate:  raw.UploadDate,
		Duration:    raw.Duration,
		Description: raw.Description,
		URL:         webURL,
		ChannelSlug: ChannelSlug(raw.Channel),
	}, nil
}

// DownloadSubtitles runs yt-dlp to download TTML subtitles to a temp directory.
// Returns the path to the downloaded TTML file.
func DownloadSubtitles(cfg *Config, videoURL string, videoID string) (string, error) {
	// Create a temp directory for the download
	tmpDir, err := os.MkdirTemp("", "yt-mcp-*")
	if err != nil {
		return "", fmt.Errorf("creating temp dir: %w", err)
	}

	outputTemplate := filepath.Join(tmpDir, "%(id)s.%(ext)s")

	cmd := exec.Command(cfg.YtDlpPath,
		"--write-subs",
		"--write-auto-subs",
		"--sub-langs", "en.*,en",
		"--sub-format", "ttml",
		"--skip-download",
		"-o", outputTemplate,
		videoURL,
	)
	cmd.Stderr = os.Stderr

	// Run yt-dlp — it may exit non-zero even if some subs downloaded
	// (e.g., 429 on one language variant while another succeeded)
	cmd.Run()

	// Check if any TTML file was actually created
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("reading temp dir: %w", err)
	}

	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".ttml") {
			return filepath.Join(tmpDir, e.Name()), nil
		}
	}

	os.RemoveAll(tmpDir)
	return "", fmt.Errorf("no TTML subtitle file found — this video may not have English subtitles")
}

// ── Full Download Pipeline ───────────────────────────────────────────────────

// DownloadResult contains the output of a successful download.
type DownloadResult struct {
	Metadata   *VideoMetadata
	Transcript string // the transcript.md content
	OutputDir  string // path to the output directory
}

// DownloadVideo is the full pipeline: fetch metadata → download subs → parse → write files.
func DownloadVideo(cfg *Config, rawURL string, force bool) (*DownloadResult, error) {
	// 1. Extract video ID
	videoID, err := ExtractVideoID(rawURL)
	if err != nil {
		return nil, err
	}

	videoURL := CanonicalURL(videoID)

	// 2. Fetch metadata
	meta, err := FetchMetadata(cfg, videoURL)
	if err != nil {
		return nil, fmt.Errorf("fetching metadata: %w", err)
	}

	// 3. Determine output directory
	outputDir := filepath.Join(cfg.YTDir, meta.ChannelSlug, videoID)

	// 4. Check if already downloaded
	transcriptPath := filepath.Join(outputDir, "transcript.md")
	if !force {
		if _, err := os.Stat(transcriptPath); err == nil {
			// Already exists — read and return existing content
			content, err := os.ReadFile(transcriptPath)
			if err != nil {
				return nil, fmt.Errorf("reading existing transcript: %w", err)
			}
			return &DownloadResult{
				Metadata:   meta,
				Transcript: string(content),
				OutputDir:  outputDir,
			}, nil
		}
	}

	// 5. Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("creating output dir: %w", err)
	}

	// 6. Download subtitles
	ttmlPath, err := DownloadSubtitles(cfg, videoURL, videoID)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(filepath.Dir(ttmlPath)) // clean up temp dir

	// 7. Process TTML → markdown + cues.json
	transcript, err := ProcessTTML(ttmlPath, meta, outputDir)
	if err != nil {
		return nil, fmt.Errorf("processing transcript: %w", err)
	}

	// 8. Write metadata.json
	metaJSON, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling metadata: %w", err)
	}
	metaPath := filepath.Join(outputDir, "metadata.json")
	if err := os.WriteFile(metaPath, metaJSON, 0644); err != nil {
		return nil, fmt.Errorf("writing metadata.json: %w", err)
	}

	return &DownloadResult{
		Metadata:   meta,
		Transcript: transcript,
		OutputDir:  outputDir,
	}, nil
}

// ── Helpers for yt_get ───────────────────────────────────────────────────────

// FindVideoDir searches the yt/ directory tree for a video ID folder.
func FindVideoDir(ytDir string, videoID string) (string, error) {
	// Walk channel directories looking for the video ID
	channels, err := os.ReadDir(ytDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no downloaded videos found (yt/ directory doesn't exist)")
		}
		return "", fmt.Errorf("reading yt dir: %w", err)
	}

	for _, ch := range channels {
		if !ch.IsDir() {
			continue
		}
		candidate := filepath.Join(ytDir, ch.Name(), videoID)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("video %s not found in %s", videoID, ytDir)
}

// LoadVideoData reads metadata.json and transcript.md from a video directory.
func LoadVideoData(dir string) (*VideoMetadata, string, error) {
	// Read metadata
	metaPath := filepath.Join(dir, "metadata.json")
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, "", fmt.Errorf("reading metadata.json: %w", err)
	}
	var meta VideoMetadata
	if err := json.Unmarshal(metaData, &meta); err != nil {
		return nil, "", fmt.Errorf("parsing metadata.json: %w", err)
	}

	// Read transcript
	transcriptPath := filepath.Join(dir, "transcript.md")
	transcriptData, err := os.ReadFile(transcriptPath)
	if err != nil {
		return nil, "", fmt.Errorf("reading transcript.md: %w", err)
	}

	return &meta, string(transcriptData), nil
}

// ListVideos walks the yt/ directory and returns metadata for all downloaded videos.
func ListVideos(ytDir string, channelFilter string, limit int) ([]VideoMetadata, error) {
	if limit <= 0 {
		limit = 20
	}

	channels, err := os.ReadDir(ytDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // empty is fine
		}
		return nil, fmt.Errorf("reading yt dir: %w", err)
	}

	var results []VideoMetadata
	for _, ch := range channels {
		if !ch.IsDir() {
			continue
		}
		if channelFilter != "" && ch.Name() != channelFilter {
			continue
		}

		videos, err := os.ReadDir(filepath.Join(ytDir, ch.Name()))
		if err != nil {
			continue
		}
		for _, v := range videos {
			if !v.IsDir() {
				continue
			}
			metaPath := filepath.Join(ytDir, ch.Name(), v.Name(), "metadata.json")
			data, err := os.ReadFile(metaPath)
			if err != nil {
				continue
			}
			var meta VideoMetadata
			if err := json.Unmarshal(data, &meta); err != nil {
				continue
			}
			results = append(results, meta)
			if len(results) >= limit {
				return results, nil
			}
		}
	}
	return results, nil
}

// SearchTranscripts searches across all downloaded transcripts for a query string.
func SearchTranscripts(ytDir string, query string, channelFilter string, limit int) ([]SearchHit, error) {
	if limit <= 0 {
		limit = 10
	}
	queryLower := strings.ToLower(query)

	channels, err := os.ReadDir(ytDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading yt dir: %w", err)
	}

	var hits []SearchHit
	for _, ch := range channels {
		if !ch.IsDir() {
			continue
		}
		if channelFilter != "" && ch.Name() != channelFilter {
			continue
		}

		videos, err := os.ReadDir(filepath.Join(ytDir, ch.Name()))
		if err != nil {
			continue
		}
		for _, v := range videos {
			if !v.IsDir() {
				continue
			}
			dir := filepath.Join(ytDir, ch.Name(), v.Name())

			meta, transcript, err := LoadVideoData(dir)
			if err != nil {
				continue
			}

			// Search paragraphs
			paragraphs := strings.Split(transcript, "\n\n")
			for _, p := range paragraphs {
				if strings.Contains(strings.ToLower(p), queryLower) {
					// Try to extract timestamp from paragraph start: [M:SS](url&t=N)
					tsLink := extractTimestampLink(p)
					hits = append(hits, SearchHit{
						VideoID:   meta.ID,
						Title:     meta.Title,
						Channel:   meta.Channel,
						Date:      meta.UploadDate,
						URL:       meta.URL,
						Excerpt:   strings.TrimSpace(p),
						Timestamp: tsLink,
					})
					if len(hits) >= limit {
						return hits, nil
					}
				}
			}
		}
	}
	return hits, nil
}

// SearchHit represents a search result from transcript search.
type SearchHit struct {
	VideoID   string `json:"video_id"`
	Title     string `json:"title"`
	Channel   string `json:"channel"`
	Date      string `json:"date"`
	URL       string `json:"url"`
	Excerpt   string `json:"excerpt"`
	Timestamp string `json:"timestamp"` // clickable ?t= link if found
}

// extractTimestampLink pulls the first markdown link from a paragraph that looks like a timestamp.
// e.g., "[0:30](https://www.youtube.com/watch?v=abc&t=30)"
func extractTimestampLink(paragraph string) string {
	// Look for [M:SS](url) or [H:MM:SS](url) pattern at start
	if len(paragraph) > 0 && paragraph[0] == '[' {
		endBracket := strings.Index(paragraph, "](")
		if endBracket > 0 {
			endParen := strings.Index(paragraph[endBracket:], ")")
			if endParen > 0 {
				return paragraph[:endBracket+endParen+1]
			}
		}
	}
	return ""
}

// VideoOutputDir returns the output directory path for a video.
func VideoOutputDir(cfg *Config, channelSlug, videoID string) string {
	return filepath.Join(cfg.YTDir, channelSlug, videoID)
}

// NormalizeURL cleans up a YouTube URL, handling mobile links, extra params, etc.
func NormalizeURL(rawURL string) string {
	// Parse and rebuild to strip tracking params
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// Convert mobile URLs
	if u.Host == "m.youtube.com" {
		u.Host = "www.youtube.com"
	}

	return u.String()
}
