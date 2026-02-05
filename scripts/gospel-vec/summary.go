package main

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

// Summarizer generates LLM summaries using LM Studio
type Summarizer struct {
	baseURL string
	model   string
}

// NewSummarizer creates a new summarizer
func NewSummarizer(baseURL, model string) *Summarizer {
	return &Summarizer{
		baseURL: baseURL,
		model:   model,
	}
}

// ChatMessage represents a message in a chat completion request
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChapterSummary contains parsed summary components
type ChapterSummary struct {
	Keywords []string `json:"keywords"`
	Summary  string   `json:"summary"`
	KeyVerse string   `json:"key_verse"`
	Raw      string   `json:"-"`
}

// SummarizeChapter generates a summary of a chapter optimized for semantic search
func (s *Summarizer) SummarizeChapter(ctx context.Context, book string, chapter int, content string) (*ChapterSummary, error) {
	systemPrompt := `Create a summary optimized for semantic search indexing.

Format your response EXACTLY like this:
KEYWORDS: [10-15 comma-separated searchable terms including people, places, concepts, events]
SUMMARY: [50-75 word narrative covering main events and principles, present tense]
KEY_VERSE: [Most memorable verse with reference]

Keep output under 200 words total. No other text.`

	userPrompt := fmt.Sprintf(`Summarize %s chapter %d:

%s`, book, chapter, content)

	response, err := s.chat(ctx, systemPrompt, userPrompt, 300)
	if err != nil {
		return nil, err
	}

	// Parse the structured response
	summary := &ChapterSummary{Raw: response}

	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "KEYWORDS:") {
			kwStr := strings.TrimPrefix(line, "KEYWORDS:")
			kwStr = strings.TrimSpace(kwStr)
			keywords := strings.Split(kwStr, ",")
			for _, kw := range keywords {
				kw = strings.TrimSpace(kw)
				if kw != "" {
					summary.Keywords = append(summary.Keywords, kw)
				}
			}
		} else if strings.HasPrefix(line, "SUMMARY:") {
			summary.Summary = strings.TrimSpace(strings.TrimPrefix(line, "SUMMARY:"))
		} else if strings.HasPrefix(line, "KEY_VERSE:") {
			summary.KeyVerse = strings.TrimSpace(strings.TrimPrefix(line, "KEY_VERSE:"))
		}
	}

	// Deduplicate keywords (case-insensitive)
	summary.Keywords = deduplicateKeywords(summary.Keywords)

	return summary, nil
}

// deduplicateKeywords removes duplicate keywords (case-insensitive)
// Keeps the first occurrence's casing
func deduplicateKeywords(keywords []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(keywords))

	for _, kw := range keywords {
		lower := strings.ToLower(kw)
		if !seen[lower] {
			seen[lower] = true
			result = append(result, kw)
		}
	}

	return result
}

// ShortSummary generates a brief 50-75 word summary for paragraph-level indexing
func (s *Summarizer) ShortSummary(ctx context.Context, book string, chapter int, content string) (string, error) {
	systemPrompt := `Create a one-paragraph summary (50-75 words) optimized for semantic search. Include key events, people, and principles. Write in present tense. No headers or labels, just the summary text.`

	userPrompt := fmt.Sprintf(`Summarize %s chapter %d:

%s`, book, chapter, content)

	return s.chat(ctx, systemPrompt, userPrompt, 150)
}

// DetectThemes identifies narrative themes/sections within a chapter
func (s *Summarizer) DetectThemes(ctx context.Context, book string, chapter int, verses []string) ([]ThemeRange, error) {
	systemPrompt := `Identify narrative sections in this scripture chapter. Return ONLY valid JSON array.

Format: [{"range": "1-5", "theme": "Brief description"}]

Rules:
- Identify 2-5 natural narrative sections
- Use verse numbers for ranges
- Keep descriptions under 15 words
- No explanation, just the JSON array`

	// Format verses with numbers
	var versesText strings.Builder
	for i, verse := range verses {
		versesText.WriteString(fmt.Sprintf("%d. %s\n", i+1, verse))
	}

	userPrompt := fmt.Sprintf(`%s chapter %d:

%s`, book, chapter, versesText.String())

	response, err := s.chat(ctx, systemPrompt, userPrompt, 300)
	if err != nil {
		return nil, err
	}

	// Parse JSON response
	var themes []ThemeRange
	if err := json.Unmarshal([]byte(response), &themes); err != nil {
		// Try to extract JSON from response if it's wrapped in text
		start := strings.Index(response, "[")
		end := strings.LastIndex(response, "]")
		if start >= 0 && end > start {
			jsonStr := response[start : end+1]
			if err := json.Unmarshal([]byte(jsonStr), &themes); err != nil {
				return nil, fmt.Errorf("parsing theme response: %w (response: %s)", err, response)
			}
		} else {
			return nil, fmt.Errorf("parsing theme response: %w (response: %s)", err, response)
		}
	}

	return themes, nil
}

// ThemeRange represents a detected theme with verse range
type ThemeRange struct {
	Range string `json:"range"`
	Theme string `json:"theme"`
}

// chat performs a chat completion request
func (s *Summarizer) chat(ctx context.Context, systemPrompt, userPrompt string, maxTokens int) (string, error) {
	if maxTokens <= 0 {
		maxTokens = 500 // default
	}

	reqBody := map[string]any{
		"model": s.model,
		"messages": []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		"temperature": 0.2, // Low temperature for consistency
		"max_tokens":  maxTokens,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second} // Longer timeout for generation
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("chat API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from model")
	}

	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}

// GetAvailableModels queries LM Studio for available models
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

// TestChat tests if the chat endpoint is working
func TestChat(ctx context.Context, baseURL, model string) error {
	s := NewSummarizer(baseURL, model)
	_, err := s.chat(ctx, "You are a test assistant.", "Say 'hello' in one word.", 50)
	return err
}
