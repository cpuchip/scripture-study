package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// config is the parsed environment configuration. Only DATABASE_URL is
// required for phase 1; LLM/Thummim variables are read defensively and
// validated only when the relevant feature is enabled.
type config struct {
	ListenAddr  string
	DatabaseURL string

	// Phase 3
	ModernFetchDailyCap int

	// Phase 4 — server-side default (optional; BYOK is the primary path)
	LLMProxyEnabled bool
	LLMBYOKEnabled  bool
	LLMProvider     string
	LLMBaseURL      string
	LLMAPIKey       string
	LLMModel        string

	LLMSessionTTL    time.Duration
	LLMSessionSliding bool

	LLMRatePerIPPerMin      int
	LLMRatePerIPPerDay      int
	LLMGlobalTokenCapPerDay int

	LLMMaxTokensDefault   int
	LLMMaxTokensHard      int
	LLMTemperatureDefault float64
	LLMTemperatureHard    float64
	LLMTimeout            time.Duration

	OpencodeGoAPIKey string
}

func loadConfig() (*config, error) {
	cfg := &config{
		ListenAddr:              envOrDefault("LISTEN_ADDR", ":8080"),
		DatabaseURL:             os.Getenv("DATABASE_URL"),
		ModernFetchDailyCap:     envInt("MODERN_FETCH_DAILY_CAP", 5000),
		LLMProxyEnabled:         envBool("LLM_PROXY_ENABLED", true),
		LLMBYOKEnabled:          envBool("LLM_BYOK_ENABLED", true),
		LLMProvider:             envOrDefault("LLM_PROVIDER", "mock"),
		LLMBaseURL:              os.Getenv("LLM_BASE_URL"),
		LLMAPIKey:               os.Getenv("LLM_API_KEY"),
		LLMModel:                os.Getenv("LLM_MODEL"),
		LLMSessionTTL:           time.Duration(envInt("LLM_SESSION_TTL_HOURS", 24)) * time.Hour,
		LLMSessionSliding:       envBool("LLM_SESSION_SLIDING_WINDOW", true),
		LLMRatePerIPPerMin:      envInt("LLM_RATE_PER_IP_PER_MIN", 10),
		LLMRatePerIPPerDay:      envInt("LLM_RATE_PER_IP_PER_DAY", 1000),
		LLMGlobalTokenCapPerDay: envInt("LLM_GLOBAL_TOKEN_CAP_PER_DAY", 200000),
		LLMMaxTokensDefault:     envInt("LLM_MAX_TOKENS_DEFAULT", 800),
		LLMMaxTokensHard:        envInt("LLM_MAX_TOKENS_HARD", 1500),
		LLMTemperatureDefault:   envFloat("LLM_TEMPERATURE_DEFAULT", 0.3),
		LLMTemperatureHard:      envFloat("LLM_TEMPERATURE_HARD", 0.7),
		LLMTimeout:              time.Duration(envInt("LLM_TIMEOUT_SECONDS", 60)) * time.Second,
		OpencodeGoAPIKey:        os.Getenv("OPENCODE_GO_API_KEY"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	return cfg, nil
}

func envOrDefault(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envInt(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func envFloat(k string, def float64) float64 {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return f
}

func envBool(k string, def bool) bool {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	switch v {
	case "true", "TRUE", "True", "1", "yes", "YES", "on", "ON":
		return true
	case "false", "FALSE", "False", "0", "no", "NO", "off", "OFF":
		return false
	}
	return def
}
