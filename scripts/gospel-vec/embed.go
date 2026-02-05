package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// NewLMStudioEmbedder creates an embedding function for LM Studio
func NewLMStudioEmbedder(baseURL, model string) func(ctx context.Context, text string) ([]float32, error) {
	return func(ctx context.Context, text string) ([]float32, error) {
		return getEmbedding(ctx, baseURL, model, text)
	}
}

// getEmbedding calls LM Studio's embedding endpoint
func getEmbedding(ctx context.Context, baseURL, model, text string) ([]float32, error) {
	reqBody := map[string]any{
		"model": model,
		"input": text,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/embeddings", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return result.Data[0].Embedding, nil
}

// BatchEmbedder handles batch embedding with progress reporting
type BatchEmbedder struct {
	baseURL string
	model   string
}

// NewBatchEmbedder creates a new batch embedder
func NewBatchEmbedder(baseURL, model string) *BatchEmbedder {
	return &BatchEmbedder{
		baseURL: baseURL,
		model:   model,
	}
}

// EmbedBatch embeds multiple texts with progress callback
func (b *BatchEmbedder) EmbedBatch(ctx context.Context, texts []string, onProgress func(current, total int)) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))

	for i, text := range texts {
		if onProgress != nil {
			onProgress(i+1, len(texts))
		}

		embedding, err := getEmbedding(ctx, b.baseURL, b.model, text)
		if err != nil {
			return nil, fmt.Errorf("embedding text %d: %w", i, err)
		}

		embeddings[i] = embedding
	}

	return embeddings, nil
}

// TestEmbedding tests if the embedding endpoint is working
func TestEmbedding(ctx context.Context, baseURL, model string) error {
	_, err := getEmbedding(ctx, baseURL, model, "test")
	return err
}
