package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config is the cockpit's local, human-side state — chiefly the sticky active
// project (like a kubectl context or the current git branch). It lives in a
// small JSON file in the user's home dir; nothing here is substrate state.
type Config struct {
	ActiveProject string `json:"active_project"`
}

// configPath resolves the on-disk config location. STEWARDS_CONFIG overrides
// (used by tests); otherwise ~/.stewards.json, falling back to the cwd if the
// home dir can't be determined.
func configPath() string {
	if p := os.Getenv("STEWARDS_CONFIG"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ".stewards.json"
	}
	return filepath.Join(home, ".stewards.json")
}

// loadConfig reads the config file, returning a zero Config if it is absent or
// unreadable (a missing config is normal on first run, not an error).
func loadConfig() Config {
	var c Config
	b, err := os.ReadFile(configPath())
	if err != nil {
		return c
	}
	_ = json.Unmarshal(b, &c)
	return c
}

// saveConfig writes the config file (pretty-printed, 0644).
func saveConfig(c Config) error {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), b, 0o644)
}

// activeProject resolves the effective active project: the STEWARDS_PROJECT env
// override wins, else the saved config. Empty means "no active project."
func activeProject() string {
	if p := os.Getenv("STEWARDS_PROJECT"); p != "" {
		return p
	}
	return loadConfig().ActiveProject
}
