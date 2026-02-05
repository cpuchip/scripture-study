package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	ctx := context.Background()
	baseURL := "http://localhost:1234/v1"
	model := "qwen/qwen3-vl-8b"

	// Load 1 Nephi 3 as test content
	content, err := os.ReadFile("../../../../gospel-library/eng/scriptures/bofm/1-ne/3.md")
	if err != nil {
		fmt.Printf("Error loading file: %v\n", err)
		return
	}

	chapter := string(content)

	fmt.Println("=== Testing Theme Detection Prompts ===")
	fmt.Println()

	// Theme Detection - JSON format
	prompt1System := `You identify narrative sections in scripture chapters. Return ONLY valid JSON array.

Format:
[{"range": "1-5", "theme": "Brief description"}]

Rules:
- Identify 2-5 natural narrative sections
- Use verse numbers for ranges
- Keep descriptions under 15 words
- No explanation, just JSON`

	prompt1User := fmt.Sprintf("1 Nephi chapter 3:\n\n%s", chapter)

	fmt.Println("--- Theme Detection (Strict JSON) ---")
	result1, err := chat(ctx, baseURL, model, prompt1System, prompt1User)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println(result1)
		// Try to parse
		var themes []map[string]string
		if err := json.Unmarshal([]byte(result1), &themes); err != nil {
			// Try extracting JSON
			start := strings.Index(result1, "[")
			end := strings.LastIndex(result1, "]")
			if start >= 0 && end > start {
				jsonStr := result1[start : end+1]
				if err := json.Unmarshal([]byte(jsonStr), &themes); err != nil {
					fmt.Printf("Parse error: %v\n", err)
				} else {
					fmt.Printf("\n✅ Parsed %d themes\n", len(themes))
				}
			}
		} else {
			fmt.Printf("\n✅ Parsed %d themes\n", len(themes))
		}
	}
	fmt.Println()

	// Short summary for embedding
	prompt2System := `Create a one-paragraph summary (50-75 words) optimized for semantic search. Include key events, people, and principles. Write in present tense.`

	prompt2User := fmt.Sprintf("Summarize 1 Nephi 3:\n\n%s", chapter)

	fmt.Println("--- Short Summary (50-75 words) ---")
	result2, err := chat(ctx, baseURL, model, prompt2System, prompt2User)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println(result2)
		words := len(strings.Fields(result2))
		fmt.Printf("\n(Word count: %d)\n", words)
	}
	fmt.Println()

	// Keywords extraction
	prompt3System := `Extract 10-15 searchable keywords/phrases from this chapter. Return as comma-separated list. Include: people, places, concepts, themes, events.`

	prompt3User := fmt.Sprintf("1 Nephi 3:\n\n%s", chapter)

	fmt.Println("--- Keywords Extraction ---")
	result3, err := chat(ctx, baseURL, model, prompt3System, prompt3User)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println(result3)
	}
}

func chat(ctx context.Context, baseURL, model, systemPrompt, userPrompt string) (string, error) {
	messages := []map[string]string{
		{"role": "system", "content": systemPrompt},
		{"role": "user", "content": userPrompt},
	}

	reqBody := map[string]any{
		"model":       model,
		"messages":    messages,
		"temperature": 0.2,
		"max_tokens":  400,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
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
