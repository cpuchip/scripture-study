package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cpuchip/scripture-study/scripts/becoming/internal/auth"
	"github.com/cpuchip/scripture-study/scripts/becoming/internal/db"
	"github.com/go-chi/chi/v5"
)

// testDB connects to a PostgreSQL test database.
// Set BECOMING_TEST_DB to a PostgreSQL connection string to run these tests.
func testDB(t *testing.T) *db.DB {
	t.Helper()
	dsn := os.Getenv("BECOMING_TEST_DB")
	if dsn == "" {
		t.Skip("skipping: set BECOMING_TEST_DB to a PostgreSQL connection string to run")
	}

	database, err := db.Open(dsn)
	if err != nil {
		t.Fatalf("opening test database: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	// Create a test user (ignore conflict if already exists)
	if _, err := database.Exec(`INSERT INTO users (id, email, name) VALUES (1, 'test@test.com', 'Test') ON CONFLICT (id) DO NOTHING`); err != nil {
		t.Fatal(err)
	}
	return database
}

// withUser sets the auth context for a request.
func withUser(r *http.Request, userID int64) *http.Request {
	return r.WithContext(auth.WithUserID(r.Context(), userID))
}

func TestUpdateTask_PartialUpdate(t *testing.T) {
	database := testDB(t)

	// Create a task with all fields populated
	task := &db.Task{
		Title:       "Read Alma 32",
		Description: "Study faith chapter",
		Type:        "once",
		Status:      "active",
		Scripture:   "Alma 32:21",
	}
	if err := database.CreateTask(1, task); err != nil {
		t.Fatal(err)
	}

	r := chi.NewRouter()
	r.Put("/tasks/{id}", updateTask(database, nil))

	// Send partial update — only change status
	body, _ := json.Marshal(map[string]any{"status": "completed"})
	req := httptest.NewRequest("PUT", fmt.Sprintf("/tasks/%d", task.ID), bytes.NewReader(body))
	req = withUser(req, 1)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify: status changed, but all other fields preserved
	got, err := database.GetTask(1, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "completed" {
		t.Errorf("status: want %q, got %q", "completed", got.Status)
	}
	if got.Title != "Read Alma 32" {
		t.Errorf("title lost: want %q, got %q", "Read Alma 32", got.Title)
	}
	if got.Description != "Study faith chapter" {
		t.Errorf("description lost: want %q, got %q", "Study faith chapter", got.Description)
	}
	if got.Scripture != "Alma 32:21" {
		t.Errorf("scripture lost: want %q, got %q", "Alma 32:21", got.Scripture)
	}
	if got.Type != "once" {
		t.Errorf("type lost: want %q, got %q", "once", got.Type)
	}
}

func TestUpdateNote_PartialUpdate(t *testing.T) {
	database := testDB(t)

	note := &db.Note{
		Content: "Original note content",
		Pinned:  true,
	}
	if err := database.CreateNote(1, note); err != nil {
		t.Fatal(err)
	}

	r := chi.NewRouter()
	r.Put("/notes/{id}", updateNote(database))

	// Send partial update — only change content
	body, _ := json.Marshal(map[string]any{"content": "Updated content"})
	req := httptest.NewRequest("PUT", fmt.Sprintf("/notes/%d", note.ID), bytes.NewReader(body))
	req = withUser(req, 1)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	got, err := database.GetNote(1, note.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Content != "Updated content" {
		t.Errorf("content: want %q, got %q", "Updated content", got.Content)
	}
	if !got.Pinned {
		t.Error("pinned was lost: should still be true")
	}
}

func TestUpdatePrompt_PartialUpdate(t *testing.T) {
	database := testDB(t)

	prompt := &db.Prompt{
		Text:      "What did I learn today?",
		Active:    true,
		SortOrder: 5,
	}
	if err := database.CreatePrompt(1, prompt); err != nil {
		t.Fatal(err)
	}

	r := chi.NewRouter()
	r.Put("/prompts/{id}", updatePrompt(database))

	// Send partial update — only toggle active
	body, _ := json.Marshal(map[string]any{"active": false})
	req := httptest.NewRequest("PUT", fmt.Sprintf("/prompts/%d", prompt.ID), bytes.NewReader(body))
	req = withUser(req, 1)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	got, err := database.GetPrompt(1, prompt.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Active {
		t.Error("active should be false after update")
	}
	if got.Text != "What did I learn today?" {
		t.Errorf("text lost: want %q, got %q", "What did I learn today?", got.Text)
	}
	if got.SortOrder != 5 {
		t.Errorf("sort_order lost: want 5, got %d", got.SortOrder)
	}
}
