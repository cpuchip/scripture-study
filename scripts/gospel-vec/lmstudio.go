package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// EnsureLMStudio checks if LM Studio server is running and starts it if not,
// then verifies the required embedding model is loaded and loads it if not.
func EnsureLMStudio(ctx context.Context, baseURL, embeddingModel string) error {
	if err := ensureLMStudioServer(ctx, baseURL); err != nil {
		return err
	}
	if err := ensureLMStudioModel(ctx, baseURL, embeddingModel); err != nil {
		return err
	}
	return nil
}

func ensureLMStudioServer(ctx context.Context, baseURL string) error {
	if isLMStudioRunning(ctx, baseURL) {
		return nil
	}

	log.Printf("LM Studio server not running, starting with 'lms server start'...")

	if _, err := exec.LookPath("lms"); err != nil {
		return fmt.Errorf("lms CLI not found in PATH: %w", err)
	}

	cmd := exec.CommandContext(ctx, "lms", "server", "start")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("starting LM Studio server: %w\nOutput: %s", err, string(output))
	}
	log.Printf("lms server start: %s", strings.TrimSpace(string(output)))

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if isLMStudioRunning(ctx, baseURL) {
			log.Printf("LM Studio server is ready")
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
	return fmt.Errorf("LM Studio server did not become ready within 30 seconds")
}

func ensureLMStudioModel(ctx context.Context, baseURL, modelID string) error {
	loaded, err := listLMStudioModels(ctx, baseURL)
	if err != nil {
		return fmt.Errorf("listing models: %w", err)
	}

	for _, m := range loaded {
		if m == modelID || strings.Contains(m, modelID) {
			return nil
		}
	}

	log.Printf("Model %q not loaded, loading with 'lms load %s'...", modelID, modelID)

	cmd := exec.CommandContext(ctx, "lms", "load", modelID, "--yes")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("loading model %q: %w\nOutput: %s", modelID, err, string(output))
	}
	log.Printf("lms load: %s", strings.TrimSpace(string(output)))

	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		loaded, err := listLMStudioModels(ctx, baseURL)
		if err == nil {
			for _, m := range loaded {
				if m == modelID || strings.Contains(m, modelID) {
					log.Printf("Model %q loaded successfully", modelID)
					return nil
				}
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return fmt.Errorf("model %q did not appear in loaded models within 60s", modelID)
}

func listLMStudioModels(ctx context.Context, baseURL string) ([]string, error) {
	url := strings.TrimRight(baseURL, "/") + "/models"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GET %s returned %d: %s", url, resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	ids := make([]string, len(result.Data))
	for i, m := range result.Data {
		ids[i] = m.ID
	}
	return ids, nil
}

func isLMStudioRunning(ctx context.Context, baseURL string) bool {
	url := strings.TrimRight(baseURL, "/") + "/models"
	reqCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "GET", url, nil)
	if err != nil {
		return false
	}

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
