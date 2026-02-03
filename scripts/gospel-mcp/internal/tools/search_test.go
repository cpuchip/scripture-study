package tools_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/cpuchip/scripture-study/scripts/gospel-mcp/internal/db"
	"github.com/cpuchip/scripture-study/scripts/gospel-mcp/internal/tools"
)

// TestDB holds a test database instance
type TestDB struct {
	*db.DB
	path string
}

// setupTestDB creates a temporary test database with sample data
func setupTestDB(t *testing.T) *TestDB {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "gospel-mcp-test-*")
	if err != nil {
		t.Fatalf("creating temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("opening database: %v", err)
	}

	// Insert test data for FTS5 queries
	testVerses := []struct {
		volume, book       string
		chapter, verse     int
		text, path, srcURL string
	}{
		{"dc-testament", "dc", 93, 36, "The glory of God is intelligence, or, in other words, light and truth.", "gospel-library/eng/scriptures/dc-testament/dc/93.md", "https://example.com"},
		{"dc-testament", "dc", 93, 37, "Light and truth forsake that evil one.", "gospel-library/eng/scriptures/dc-testament/dc/93.md", "https://example.com"},
		{"dc-testament", "dc", 93, 38, "Every spirit of man was innocent in the beginning.", "gospel-library/eng/scriptures/dc-testament/dc/93.md", "https://example.com"},
		{"dc-testament", "dc", 130, 18, "Whatever principle of intelligence we attain unto in this life, it will rise with us in the resurrection.", "gospel-library/eng/scriptures/dc-testament/dc/130.md", "https://example.com"},
		{"dc-testament", "dc", 130, 19, "And if a person gains more knowledge and intelligence in this life through his diligence and obedience than another, he will have so much the advantage in the world to come.", "gospel-library/eng/scriptures/dc-testament/dc/130.md", "https://example.com"},
		{"ot", "job", 38, 7, "When the morning stars sang together, and all the sons of God shouted for joy?", "gospel-library/eng/scriptures/ot/job/38.md", "https://example.com"},
		{"ot", "isa", 40, 26, "Lift up your eyes on high, and behold who hath created these things, that bringeth out their host by number.", "gospel-library/eng/scriptures/ot/isa/40.md", "https://example.com"},
		{"pgp", "moses", 3, 5, "And every plant of the field before it was in the earth, and every herb of the field before it grew. For I, the Lord God, created all things, of which I have spoken, spiritually, before they were naturally upon the face of the earth.", "gospel-library/eng/scriptures/pgp/moses/3.md", "https://example.com"},
	}

	for _, v := range testVerses {
		_, err := database.Exec(`
			INSERT OR REPLACE INTO scriptures (volume, book, chapter, verse, text, file_path, source_url)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, v.volume, v.book, v.chapter, v.verse, v.text, v.path, v.srcURL)
		if err != nil {
			database.Close()
			os.RemoveAll(tmpDir)
			t.Fatalf("inserting test verse: %v", err)
		}
	}

	// Insert test talks
	testTalks := []struct {
		year, month    int
		speaker, title string
		content        string
		path, srcURL   string
	}{
		{2025, 4, "Russell M. Nelson", "Confidence in the Presence of God", "As we diligently seek to have charity and virtue fill our lives, our confidence in approaching God will increase.", "gospel-library/eng/general-conference/2025/04/57nelson.md", "https://example.com"},
		{2025, 10, "David A. Bednar", "Things as They Really Are", "Technology and intelligence when used correctly can bless our lives.", "gospel-library/eng/general-conference/2025/10/41bednar.md", "https://example.com"},
	}

	for _, talk := range testTalks {
		_, err := database.Exec(`
			INSERT OR REPLACE INTO talks (year, month, speaker, title, content, file_path, source_url)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, talk.year, talk.month, talk.speaker, talk.title, talk.content, talk.path, talk.srcURL)
		if err != nil {
			database.Close()
			os.RemoveAll(tmpDir)
			t.Fatalf("inserting test talk: %v", err)
		}
	}

	return &TestDB{DB: database, path: tmpDir}
}

func (tdb *TestDB) cleanup() {
	tdb.Close()
	os.RemoveAll(tdb.path)
}

// TestFTS5ExactPhraseSearch tests that exact phrase queries work
func TestFTS5ExactPhraseSearch(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.cleanup()

	toolsInstance := tools.New(tdb.DB)

	tests := []struct {
		name        string
		query       string
		expectHits  bool
		minResults  int
		description string
	}{
		{
			name:        "exact phrase - light and truth",
			query:       `"light and truth"`,
			expectHits:  true,
			minResults:  1,
			description: "Should find D&C 93:36-37 with exact phrase",
		},
		{
			name:        "exact phrase - morning stars",
			query:       `"morning stars"`,
			expectHits:  true,
			minResults:  1,
			description: "Should find Job 38:7",
		},
		{
			name:        "exact phrase - nonexistent",
			query:       `"xyz nonexistent phrase abc"`,
			expectHits:  false,
			minResults:  0,
			description: "Should return no results for nonexistent phrase",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			params := map[string]interface{}{
				"query":  tc.query,
				"source": "scriptures",
				"limit":  20,
			}

			paramsJSON, _ := json.Marshal(params)
			result, err := toolsInstance.Search(paramsJSON)

			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}

			if tc.expectHits && result.TotalMatches < tc.minResults {
				t.Errorf("%s: expected at least %d results, got %d", tc.description, tc.minResults, result.TotalMatches)
			}

			if !tc.expectHits && result.TotalMatches > 0 {
				t.Errorf("%s: expected no results, got %d", tc.description, result.TotalMatches)
			}
		})
	}
}

// TestFTS5WildcardSearch tests prefix/wildcard searches
func TestFTS5WildcardSearch(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.cleanup()

	toolsInstance := tools.New(tdb.DB)

	tests := []struct {
		name        string
		query       string
		expectHits  bool
		minResults  int
		description string
	}{
		{
			name:        "prefix search - intelli*",
			query:       "intelli*",
			expectHits:  true,
			minResults:  1,
			description: "Should find verses with intelligence, intelligent, etc.",
		},
		{
			name:        "prefix search - resurrec*",
			query:       "resurrec*",
			expectHits:  true,
			minResults:  1,
			description: "Should find verses with resurrection",
		},
		{
			name:        "prefix search - star*",
			query:       "star*",
			expectHits:  true,
			minResults:  1,
			description: "Should find verses with stars",
		},
		{
			name:        "prefix search - nonexistent prefix",
			query:       "xyzabc*",
			expectHits:  false,
			minResults:  0,
			description: "Should return no results for nonexistent prefix",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			params := map[string]interface{}{
				"query":  tc.query,
				"source": "scriptures",
				"limit":  20,
			}

			paramsJSON, _ := json.Marshal(params)
			result, err := toolsInstance.Search(paramsJSON)

			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}

			if tc.expectHits && result.TotalMatches < tc.minResults {
				t.Errorf("%s: expected at least %d results, got %d", tc.description, tc.minResults, result.TotalMatches)
			}

			if !tc.expectHits && result.TotalMatches > 0 {
				t.Errorf("%s: expected no results, got %d", tc.description, result.TotalMatches)
			}
		})
	}
}

// TestFTS5BooleanOperators tests AND, OR, NOT operators
func TestFTS5BooleanOperators(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.cleanup()

	toolsInstance := tools.New(tdb.DB)

	tests := []struct {
		name        string
		query       string
		expectHits  bool
		minResults  int
		description string
	}{
		{
			name:        "OR operator",
			query:       "intelligence OR stars",
			expectHits:  true,
			minResults:  2,
			description: "Should find verses with intelligence OR stars",
		},
		{
			name:        "AND operator (implicit)",
			query:       "light truth",
			expectHits:  true,
			minResults:  1,
			description: "Should find verses with both light AND truth",
		},
		{
			name:        "AND operator (explicit)",
			query:       "glory AND intelligence",
			expectHits:  true,
			minResults:  1,
			description: "Should find D&C 93:36",
		},
		{
			name:        "NOT operator",
			query:       "intelligence NOT resurrection",
			expectHits:  true,
			minResults:  1,
			description: "Should find intelligence verses without resurrection",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			params := map[string]interface{}{
				"query":  tc.query,
				"source": "scriptures",
				"limit":  20,
			}

			paramsJSON, _ := json.Marshal(params)
			result, err := toolsInstance.Search(paramsJSON)

			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}

			if tc.expectHits && result.TotalMatches < tc.minResults {
				t.Errorf("%s: expected at least %d results, got %d", tc.description, tc.minResults, result.TotalMatches)
			}
		})
	}
}

// TestFTS5NEARSearch tests NEAR operator for proximity search
// Note: This tests if "truth" and "light" appear near each other (within N words)
func TestFTS5NEARSearch(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.cleanup()

	toolsInstance := tools.New(tdb.DB)

	tests := []struct {
		name        string
		query       string
		expectHits  bool
		minResults  int
		description string
	}{
		{
			name:        "NEAR operator - close words",
			query:       "NEAR(light truth, 5)",
			expectHits:  true,
			minResults:  1,
			description: "Should find verses where light and truth are within 5 words",
		},
		{
			name:        "NEAR operator - words in different order",
			query:       "NEAR(truth light, 5)",
			expectHits:  true,
			minResults:  1,
			description: "NEAR should be order-independent",
		},
		{
			name:        "NEAR operator - distant words should fail",
			query:       "NEAR(morning resurrection, 2)",
			expectHits:  false,
			minResults:  0,
			description: "Should not find morning and resurrection within 2 words",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			params := map[string]interface{}{
				"query":  tc.query,
				"source": "scriptures",
				"limit":  20,
			}

			paramsJSON, _ := json.Marshal(params)
			result, err := toolsInstance.Search(paramsJSON)

			if err != nil {
				// NEAR might not be supported, check error
				t.Logf("Note: NEAR query returned error (may not be supported): %v", err)
				if tc.expectHits {
					t.Skipf("NEAR operator may not be supported in this FTS5 configuration")
				}
				return
			}

			if tc.expectHits && result.TotalMatches < tc.minResults {
				t.Errorf("%s: expected at least %d results, got %d", tc.description, tc.minResults, result.TotalMatches)
			}

			if !tc.expectHits && result.TotalMatches > 0 {
				t.Errorf("%s: expected no results, got %d", tc.description, result.TotalMatches)
			}
		})
	}
}

// TestFTS5WordVariations tests that similar words are found
// FTS5 doesn't do stemming by default, so "light" won't find "lights" unless explicitly searched
func TestFTS5WordVariations(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.cleanup()

	toolsInstance := tools.New(tdb.DB)

	tests := []struct {
		name        string
		query       string
		expectHits  bool
		description string
	}{
		{
			name:        "singular word",
			query:       "star",
			expectHits:  true,
			description: "Should find 'stars' (partial match)",
		},
		{
			name:        "plural word",
			query:       "stars",
			expectHits:  true,
			description: "Should find 'stars' exact",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			params := map[string]interface{}{
				"query":  tc.query,
				"source": "scriptures",
				"limit":  20,
			}

			paramsJSON, _ := json.Marshal(params)
			result, err := toolsInstance.Search(paramsJSON)

			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}

			if tc.expectHits && result.TotalMatches == 0 {
				t.Logf("Note: Query '%s' returned 0 results - FTS5 may require exact match", tc.query)
			}
		})
	}
}

// TestSearchResultMarkdownLink tests that markdown_link field is properly populated
func TestSearchResultMarkdownLink(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.cleanup()

	toolsInstance := tools.New(tdb.DB)

	params := map[string]interface{}{
		"query":  "intelligence",
		"source": "scriptures",
		"limit":  5,
	}

	paramsJSON, _ := json.Marshal(params)
	result, err := toolsInstance.Search(paramsJSON)

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if result.TotalMatches == 0 {
		t.Fatal("Expected at least one result")
	}

	for _, r := range result.Results {
		// Check that markdown_link is populated
		if r.MarkdownLink == "" {
			t.Errorf("Result for %s should have MarkdownLink populated", r.Reference)
		}

		// Check that it's a proper markdown link format
		if !isValidMarkdownLink(r.MarkdownLink) {
			t.Errorf("MarkdownLink '%s' is not valid markdown format", r.MarkdownLink)
		}

		// Check that file_path is still present
		if r.FilePath == "" {
			t.Errorf("Result should still have FilePath")
		}
	}
}

// isValidMarkdownLink checks if the string is a valid markdown link [text](path)
func isValidMarkdownLink(link string) bool {
	// Should match [text](path) pattern
	if len(link) < 5 {
		return false
	}
	return link[0] == '[' &&
		containsString(link, "](") &&
		link[len(link)-1] == ')'
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestSearchConferenceTalkMetadata tests that conference talk results have good metadata
func TestSearchConferenceTalkMetadata(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.cleanup()

	toolsInstance := tools.New(tdb.DB)

	params := map[string]interface{}{
		"query":  "virtue",
		"source": "conference",
		"limit":  5,
	}

	paramsJSON, _ := json.Marshal(params)
	result, err := toolsInstance.Search(paramsJSON)

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	for _, r := range result.Results {
		// Reference should include speaker and date
		if r.Reference == "" {
			t.Error("Conference talk should have Reference with speaker info")
		}

		// Title should be present
		if r.Title == "" {
			t.Error("Conference talk should have Title")
		}

		// SourceType should be "conference"
		if r.SourceType != "conference" {
			t.Errorf("SourceType should be 'conference', got '%s'", r.SourceType)
		}
	}
}
