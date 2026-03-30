package vec

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// GetAvailableModels queries LM Studio for loaded models.
func GetAvailableModels(ctx context.Context, baseURL string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	models := make([]string, len(result.Data))
	for i, m := range result.Data {
		models[i] = m.ID
	}
	return models, nil
}

// EnsureModelLoaded checks whether the embedding model is already loaded in
// LM Studio (via the OpenAI-compat /v1/models endpoint). If not, it loads the
// model with the requested context length via the management API. This avoids
// duplicate model instances that waste VRAM/RAM.
func EnsureModelLoaded(ctx context.Context, embeddingURL, model string, contextLength int) error {
	// Check if the model is already loaded.
	loaded, err := GetAvailableModels(ctx, embeddingURL)
	if err == nil {
		for _, id := range loaded {
			if id == model {
				return nil // already loaded — nothing to do
			}
		}
	}
	// Not loaded (or couldn't check) — load it now.
	serverBase := strings.TrimSuffix(embeddingURL, "/v1")
	loadURL := serverBase + "/api/v1/models/load"

	reqBody, err := json.Marshal(map[string]any{
		"model":          model,
		"context_length": contextLength,
	})
	if err != nil {
		return fmt.Errorf("marshaling load request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", loadURL, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("creating load request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("calling model load API: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("model load API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Status       string  `json:"status"`
		Type         string  `json:"type"`
		LoadTimeSecs float64 `json:"load_time_seconds"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("decoding load response: %w", err)
	}

	if result.Status != "loaded" {
		return fmt.Errorf("unexpected load status: %s", result.Status)
	}

	return nil
}
