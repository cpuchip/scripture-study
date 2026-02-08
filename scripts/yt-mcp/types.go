package main

// Cue represents a single timestamped subtitle cue from a TTML file.
type Cue struct {
	Begin float64 `json:"begin"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

// Paragraph represents a group of merged cues forming a logical paragraph.
type Paragraph struct {
	Begin float64 // Start time (seconds) of first cue in the paragraph
	Text  string  // Merged, cleaned text
}

// VideoMetadata holds the structured metadata extracted from yt-dlp --dump-json.
type VideoMetadata struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Channel     string `json:"channel"`
	ChannelID   string `json:"channel_id"`
	UploadDate  string `json:"upload_date"` // YYYYMMDD
	Duration    int    `json:"duration"`    // seconds
	Description string `json:"description"`
	URL         string `json:"url"`          // canonical watch URL
	ChannelSlug string `json:"channel_slug"` // derived: lowercase-hyphenated channel name
}

// Config holds runtime configuration for the yt-mcp server.
type Config struct {
	YTDir     string // Base directory for downloads (default: "./yt")
	YtDlpPath string // Path to yt-dlp executable (default: "yt-dlp")
	StudyDir  string // Where evaluation docs go (default: "./study/yt")
}
