package main

import "os"

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		YTDir:     envOrDefault("YT_DIR", "./yt"),
		YtDlpPath: envOrDefault("YT_DLP_PATH", "yt-dlp"),
		StudyDir:  envOrDefault("YT_STUDY_DIR", "./study/yt"),
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
