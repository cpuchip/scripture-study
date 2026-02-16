package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds all configuration for gospel-vec
type Config struct {
	// LM Studio embedding endpoint
	EmbeddingURL   string `json:"embedding_url"`
	EmbeddingModel string `json:"embedding_model"`

	// LM Studio chat endpoint (for summaries)
	ChatURL   string `json:"chat_url"`
	ChatModel string `json:"chat_model"`

	// Storage paths
	DataDir string `json:"data_dir"`
	DBFile  string `json:"db_file"`

	// Content paths (relative to workspace root)
	ScripturesPath string `json:"scriptures_path"`
	ConferencePath string `json:"conference_path"`
}

// DefaultConfig returns sensible defaults for local development.
// Environment variables override defaults:
//   - GOSPEL_VEC_DATA_DIR: override data directory
//   - GOSPEL_VEC_EMBEDDING_MODEL: override embedding model name
//   - GOSPEL_VEC_CHAT_MODEL: override chat model name
//   - GOSPEL_VEC_EMBEDDING_URL: override embedding endpoint URL
//   - GOSPEL_VEC_CHAT_URL: override chat endpoint URL
func DefaultConfig() *Config {
	// Detect if running from repo root or from scripts/gospel-vec/
	scripturesPath := "../../gospel-library/eng/scriptures"
	conferencePath := "../../gospel-library/eng/general-conference"
	dataDir := "./data"

	// Check if gospel-library exists in current directory (repo root)
	if _, err := os.Stat("gospel-library"); err == nil {
		scripturesPath = "gospel-library/eng/scriptures"
		conferencePath = "gospel-library/eng/general-conference"
		dataDir = "scripts/gospel-vec/data"
	}

	cfg := &Config{
		EmbeddingURL:   "http://localhost:1234/v1",
		EmbeddingModel: "text-embedding-qwen3-embedding-4b",
		ChatURL:        "http://localhost:1234/v1",
		ChatModel:      "", // Will be auto-detected or set by user

		DataDir: dataDir,
		DBFile:  "gospel-vec.gob.gz",

		ScripturesPath: scripturesPath,
		ConferencePath: conferencePath,
	}

	// Apply environment variable overrides
	if v := os.Getenv("GOSPEL_VEC_DATA_DIR"); v != "" {
		cfg.DataDir = v
	}
	if v := os.Getenv("GOSPEL_VEC_EMBEDDING_MODEL"); v != "" {
		cfg.EmbeddingModel = v
	}
	if v := os.Getenv("GOSPEL_VEC_CHAT_MODEL"); v != "" {
		cfg.ChatModel = v
	}
	if v := os.Getenv("GOSPEL_VEC_EMBEDDING_URL"); v != "" {
		cfg.EmbeddingURL = v
	}
	if v := os.Getenv("GOSPEL_VEC_CHAT_URL"); v != "" {
		cfg.ChatURL = v
	}

	return cfg
}

// DBPath returns the full path to the database file
func (c *Config) DBPath() string {
	return filepath.Join(c.DataDir, c.DBFile)
}

// LoadConfig loads config from a JSON file, falling back to defaults
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // Use defaults if no config file
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return cfg, nil
}

// SaveConfig saves config to a JSON file
func (c *Config) SaveConfig(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}
