package config

import (
	"os"
	"path/filepath"
	"strconv"
)

// Config holds all configuration for gospel-engine.
type Config struct {
	// Data directory for SQLite + vector + summaries
	DataDir string

	// SQLite database path
	DBPath string

	// LM Studio embedding endpoint
	EmbeddingURL           string
	EmbeddingModel         string
	EmbeddingContextLength int

	// LM Studio chat endpoint (for summaries/enrichment)
	ChatURL   string
	ChatModel string

	// Content paths (relative to workspace root)
	ScripturesPath string
	ConferencePath string
	ManualsPath    string
	BooksPath      string
	MusicPath      string

	// Workspace root
	Root string
}

// Default returns sensible defaults, auto-detecting workspace root.
func Default() *Config {
	cfg := &Config{
		EmbeddingURL:           "http://localhost:1234/v1",
		EmbeddingModel:         "text-embedding-qwen3-embedding-8b",
		EmbeddingContextLength: 16384,
		ChatURL:                "http://localhost:1234/v1",
		ChatModel:              "",
	}

	// Detect workspace root
	if _, err := os.Stat("gospel-library"); err == nil {
		// Running from repo root
		cfg.Root = "."
		cfg.DataDir = filepath.Join("scripts", "gospel-engine", "data")
		cfg.DBPath = filepath.Join("scripts", "gospel-engine", "data", "gospel.db")
		cfg.ScripturesPath = filepath.Join("gospel-library", "eng", "scriptures")
		cfg.ConferencePath = filepath.Join("gospel-library", "eng", "general-conference")
		cfg.ManualsPath = filepath.Join("gospel-library", "eng", "manual")
		cfg.BooksPath = "books"
		cfg.MusicPath = filepath.Join("gospel-library", "eng", "music")
	} else {
		// Running from scripts/gospel-engine/
		cfg.Root = filepath.Join("..", "..")
		cfg.DataDir = "data"
		cfg.DBPath = filepath.Join("data", "gospel.db")
		cfg.ScripturesPath = filepath.Join("..", "..", "gospel-library", "eng", "scriptures")
		cfg.ConferencePath = filepath.Join("..", "..", "gospel-library", "eng", "general-conference")
		cfg.ManualsPath = filepath.Join("..", "..", "gospel-library", "eng", "manual")
		cfg.BooksPath = filepath.Join("..", "..", "books")
		cfg.MusicPath = filepath.Join("..", "..", "gospel-library", "eng", "music")
	}

	// Environment variable overrides
	if v := os.Getenv("GOSPEL_ENGINE_DATA_DIR"); v != "" {
		cfg.DataDir = v
		cfg.DBPath = filepath.Join(v, "gospel.db")
	}
	if v := os.Getenv("GOSPEL_ENGINE_DB"); v != "" {
		cfg.DBPath = v
	}
	if v := os.Getenv("GOSPEL_ENGINE_EMBEDDING_URL"); v != "" {
		cfg.EmbeddingURL = v
	}
	if v := os.Getenv("GOSPEL_ENGINE_EMBEDDING_MODEL"); v != "" {
		cfg.EmbeddingModel = v
	}
	if v := os.Getenv("GOSPEL_ENGINE_EMBEDDING_CONTEXT_LENGTH"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.EmbeddingContextLength = n
		}
	}
	if v := os.Getenv("GOSPEL_ENGINE_CHAT_URL"); v != "" {
		cfg.ChatURL = v
	}
	if v := os.Getenv("GOSPEL_ENGINE_CHAT_MODEL"); v != "" {
		cfg.ChatModel = v
	}
	if v := os.Getenv("GOSPEL_ENGINE_ROOT"); v != "" {
		cfg.Root = v
		cfg.ScripturesPath = filepath.Join(v, "gospel-library", "eng", "scriptures")
		cfg.ConferencePath = filepath.Join(v, "gospel-library", "eng", "general-conference")
		cfg.ManualsPath = filepath.Join(v, "gospel-library", "eng", "manual")
		cfg.BooksPath = filepath.Join(v, "books")
		cfg.MusicPath = filepath.Join(v, "gospel-library", "eng", "music")
	}

	return cfg
}

// SummariesDir returns the path to the summaries cache directory.
func (c *Config) SummariesDir() string {
	return filepath.Join(c.DataDir, "summaries")
}
