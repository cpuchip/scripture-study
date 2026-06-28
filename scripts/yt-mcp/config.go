package main

import (
	"os"
	"strconv"
)

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		YTDir:          envOrDefault("YT_DIR", "./yt"),
		YtDlpPath:      envOrDefault("YT_DLP_PATH", "yt-dlp"),
		StudyDir:       envOrDefault("YT_STUDY_DIR", "./study/yt"),
		CookieFile:     os.Getenv("YT_COOKIE_FILE"),
		FfmpegPath:     envOrDefault("FFMPEG_PATH", "ffmpeg"),
		MaxVideoHeight: envIntOrDefault("YT_MAX_HEIGHT", 720),
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOrDefault(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
