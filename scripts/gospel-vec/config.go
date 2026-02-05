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

// DefaultConfig returns sensible defaults for local development
func DefaultConfig() *Config {
	return &Config{
		EmbeddingURL:   "http://localhost:1234/v1",
		EmbeddingModel: "text-embedding-qwen3-embedding-4b",
		ChatURL:        "http://localhost:1234/v1",
		ChatModel:      "", // Will be auto-detected or set by user

		DataDir: "./data",
		DBFile:  "gospel-vec.gob.gz",

		ScripturesPath: "../../gospel-library/eng/scriptures",
		ConferencePath: "../../gospel-library/eng/general-conference",
	}
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
