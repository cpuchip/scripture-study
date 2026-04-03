package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// ---- Data types ----

// Query is a test query with its category.
type Query struct {
	Text     string `json:"text"`
	Category string `json:"category"`
}

// DocEmbedding holds a document's text, metadata, and its embedding vector.
type DocEmbedding struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	Reference string    `json:"reference"`
	Layer     string    `json:"layer"` // "summary", "verse"
	Embedding []float32 `json:"embedding"`
}

// QueryEmbedding holds a query and its embedding.
type QueryEmbedding struct {
	Query     Query     `json:"query"`
	Embedding []float32 `json:"embedding"`
}

// EmbedResult is the full output of an embed run — saved to JSON.
type EmbedResult struct {
	Tag       string           `json:"tag"`
	Model     string           `json:"model"`
	Dims      int              `json:"dims"`
	ForceDims int              `json:"force_dims"` // 0 = native
	Timestamp string           `json:"timestamp"`
	EmbedTime string           `json:"embed_time"`
	Docs      []DocEmbedding   `json:"docs"`
	Queries   []QueryEmbedding `json:"queries"`
}

// ---- Embed command ----

func runEmbed(tag, dbPath, baseURL string, forceDims int) error {
	ctx := context.Background()

	// Test connection to LM Studio
	fmt.Printf("Testing LM Studio connection at %s ...\n", baseURL)
	modelName, err := detectModel(ctx, baseURL)
	if err != nil {
		return fmt.Errorf("LM Studio not reachable: %w", err)
	}
	fmt.Printf("  Model: %s\n", modelName)

	// Test embedding to get native dimension
	testVec, err := getEmbedding(ctx, baseURL, modelName, "test", forceDims)
	if err != nil {
		return fmt.Errorf("test embedding failed: %w", err)
	}
	dims := len(testVec)
	fmt.Printf("  Dimensions: %d", dims)
	if forceDims > 0 {
		fmt.Printf(" (forced to %d)", forceDims)
	}
	fmt.Println()

	// Open gospel.db
	fmt.Printf("Opening %s ...\n", dbPath)
	db, err := sql.Open("sqlite3", dbPath+"?mode=ro")
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer db.Close()

	// Load 1 Nephi data
	summaries, err := loadSummaries(db)
	if err != nil {
		return fmt.Errorf("loading summaries: %w", err)
	}
	fmt.Printf("  Loaded %d chapter summaries\n", len(summaries))

	verses, err := loadVerses(db)
	if err != nil {
		return fmt.Errorf("loading verses: %w", err)
	}
	fmt.Printf("  Loaded %d verses\n", len(verses))

	allDocs := append(summaries, verses...)

	// Embed all documents
	fmt.Printf("\nEmbedding %d documents ...\n", len(allDocs))
	start := time.Now()
	for i := range allDocs {
		vec, err := getEmbedding(ctx, baseURL, modelName, allDocs[i].Text, forceDims)
		if err != nil {
			return fmt.Errorf("embedding doc %d (%s): %w", i, allDocs[i].ID, err)
		}
		allDocs[i].Embedding = vec
		if (i+1)%50 == 0 || i == len(allDocs)-1 {
			fmt.Printf("  %d/%d (%.1fs)\n", i+1, len(allDocs), time.Since(start).Seconds())
		}
	}
	docTime := time.Since(start)

	// Embed all queries
	queries := testQueries()
	fmt.Printf("\nEmbedding %d queries ...\n", len(queries))
	qStart := time.Now()
	queryEmbeddings := make([]QueryEmbedding, len(queries))
	for i, q := range queries {
		vec, err := getEmbedding(ctx, baseURL, modelName, q.Text, forceDims)
		if err != nil {
			return fmt.Errorf("embedding query %d (%s): %w", i, q.Text, err)
		}
		queryEmbeddings[i] = QueryEmbedding{Query: q, Embedding: vec}
	}
	queryTime := time.Since(qStart)

	totalTime := docTime + queryTime
	fmt.Printf("\nDone! %d docs in %s, %d queries in %s (total: %s)\n",
		len(allDocs), docTime.Round(time.Millisecond),
		len(queries), queryTime.Round(time.Millisecond),
		totalTime.Round(time.Millisecond))

	// Save results
	result := EmbedResult{
		Tag:       tag,
		Model:     modelName,
		Dims:      dims,
		ForceDims: forceDims,
		Timestamp: time.Now().Format(time.RFC3339),
		EmbedTime: totalTime.Round(time.Millisecond).String(),
		Docs:      allDocs,
		Queries:   queryEmbeddings,
	}

	outDir := filepath.Join("data", tag)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	outPath := filepath.Join(outDir, "embeddings.json")
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		return fmt.Errorf("writing JSON: %w", err)
	}

	fmt.Printf("Saved to %s\n", outPath)
	return nil
}

// ---- LM Studio helpers ----

// detectModel queries LM Studio /v1/models to find the loaded embedding model.
func detectModel(ctx context.Context, baseURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/models", nil)
	if err != nil {
		return "", err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding models response: %w", err)
	}

	// Find embedding model (prefer explicit embedding model, fall back to first)
	for _, m := range result.Data {
		if strings.Contains(strings.ToLower(m.ID), "embedding") {
			return m.ID, nil
		}
	}
	if len(result.Data) > 0 {
		return result.Data[0].ID, nil
	}
	return "", fmt.Errorf("no models loaded in LM Studio")
}

// getEmbedding calls LM Studio's /v1/embeddings endpoint.
func getEmbedding(ctx context.Context, baseURL, model, text string, forceDims int) ([]float32, error) {
	reqBody := map[string]any{
		"model": model,
		"input": text,
	}
	if forceDims > 0 {
		reqBody["dimensions"] = forceDims
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

// ---- Data loading ----

// loadSummaries reads 1 Nephi enrichment summaries from gospel.db.
func loadSummaries(db *sql.DB) ([]DocEmbedding, error) {
	rows, err := db.Query(`
		SELECT id, volume, book, chapter, enrichment_summary, enrichment_keywords, enrichment_christ_types
		FROM chapters
		WHERE book = '1-ne' AND enrichment_summary IS NOT NULL
		ORDER BY chapter
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []DocEmbedding
	for rows.Next() {
		var id int
		var volume, book string
		var chapter int
		var summary, keywords string
		var christTypes sql.NullString
		if err := rows.Scan(&id, &volume, &book, &chapter, &summary, &keywords, &christTypes); err != nil {
			return nil, err
		}

		ref := fmt.Sprintf("1 Nephi %d", chapter)
		// Match gospel-engine's embed format
		content := fmt.Sprintf("%s: %s\nKeywords: %s", ref, summary, keywords)
		if christTypes.Valid && christTypes.String != "" && strings.ToLower(christTypes.String) != "none" {
			content += fmt.Sprintf("\nChrist types: %s", christTypes.String)
		}

		docs = append(docs, DocEmbedding{
			ID:        fmt.Sprintf("1ne-%d-summary", chapter),
			Text:      content,
			Reference: ref,
			Layer:     "summary",
		})
	}
	return docs, rows.Err()
}

// loadVerses reads all 1 Nephi verses from gospel.db.
func loadVerses(db *sql.DB) ([]DocEmbedding, error) {
	rows, err := db.Query(`
		SELECT id, chapter, verse, text
		FROM scriptures
		WHERE book = '1-ne'
		ORDER BY chapter, verse
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []DocEmbedding
	for rows.Next() {
		var id, chapter, verse int
		var text string
		if err := rows.Scan(&id, &chapter, &verse, &text); err != nil {
			return nil, err
		}

		ref := fmt.Sprintf("1 Nephi %d:%d", chapter, verse)
		docs = append(docs, DocEmbedding{
			ID:        fmt.Sprintf("1ne-%d-%d-verse", chapter, verse),
			Text:      text,
			Reference: ref,
			Layer:     "verse",
		})
	}
	return docs, rows.Err()
}
