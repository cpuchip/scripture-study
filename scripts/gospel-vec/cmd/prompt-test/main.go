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

	// Extract just the verse content (simplified)
	chapter := string(content)

	fmt.Println("=== Testing Different Summary Prompts ===")
	fmt.Println()

	// Prompt 1: Original (simple)
	prompt1System := `You are a scripture study assistant. Create concise, searchable summaries of scripture chapters.

Guidelines:
- Focus on key doctrines, principles, and events
- Include important characters and their actions
- Note any significant prophecies or promises
- Keep the summary between 100-200 words
- Use clear, searchable language
- Do not add interpretation beyond what's in the text`

	prompt1User := fmt.Sprintf("Summarize 1 Nephi chapter 3:\n\n%s", chapter)

	fmt.Println("--- PROMPT 1: Simple Summary ---")
	result1, err := chat(ctx, baseURL, model, prompt1System, prompt1User)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println(result1)
	}
	fmt.Println()

	// Prompt 2: Structured output
	prompt2System := `You are a scripture study assistant. Summarize scripture chapters in a structured format.

Output format:
SETTING: [Time/place context]
CHARACTERS: [Key people involved]
EVENTS: [What happens, in order]
DOCTRINES: [Principles taught]
KEY VERSE: [Most important verse reference and why]

Be concise. Each section should be 1-2 sentences.`

	prompt2User := fmt.Sprintf("Summarize 1 Nephi chapter 3:\n\n%s", chapter)

	fmt.Println("--- PROMPT 2: Structured Output ---")
	result2, err := chat(ctx, baseURL, model, prompt2System, prompt2User)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println(result2)
	}
	fmt.Println()

	// Prompt 3: Study-focused
	prompt3System := `You are helping someone study the scriptures. For each chapter, create a summary that helps with:
1. Quick recall - what happens in this chapter?
2. Cross-referencing - what topics/themes are covered?
3. Application - what principles can be applied today?

Keep each section brief (1-2 sentences). Use natural language that would match search queries.`

	prompt3User := fmt.Sprintf("Create a study summary for 1 Nephi chapter 3:\n\n%s", chapter)

	fmt.Println("--- PROMPT 3: Study-Focused ---")
	result3, err := chat(ctx, baseURL, model, prompt3System, prompt3User)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println(result3)
	}
	fmt.Println()

	// Prompt 4: Search-optimized
	prompt4System := `Generate a semantic search index entry for this scripture chapter. Include:
- Main topics and themes (comma-separated keywords)
- One-paragraph narrative summary
- Key phrases that someone might search for

Focus on making this findable via semantic search queries like "faith despite obstacles" or "obtaining sacred records".`

	prompt4User := fmt.Sprintf("Create a search index entry for 1 Nephi chapter 3:\n\n%s", chapter)

	fmt.Println("--- PROMPT 4: Search-Optimized ---")
	result4, err := chat(ctx, baseURL, model, prompt4System, prompt4User)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println(result4)
	}
	fmt.Println()

	// Prompt 5: Concise with verse refs
	prompt5System := `Summarize this scripture chapter in exactly 3-4 sentences. Include specific verse references for the most important points. Write in a way that helps someone quickly understand what this chapter covers.`

	prompt5User := fmt.Sprintf("1 Nephi chapter 3:\n\n%s", chapter)

	fmt.Println("--- PROMPT 5: Concise with Verse Refs ---")
	result5, err := chat(ctx, baseURL, model, prompt5System, prompt5User)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println(result5)
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
		"temperature": 0.3,
		"max_tokens":  500,
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
