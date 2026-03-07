// Package main implements an MCP (Model Context Protocol) server for the Becoming app.
// This lets AI assistants (like GitHub Copilot) interact with practices, tasks,
// notes, and memorization cards during study sessions.
//
// Usage:
//
//	becoming-mcp -token bec_... [-url http://localhost:8080]
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var (
	apiURL   string
	apiToken string
)

func main() {
	urlFlag := flag.String("url", envOrDefault("BECOMING_URL", "http://localhost:8080"), "Becoming API URL")
	tokenFlag := flag.String("token", os.Getenv("BECOMING_TOKEN"), "API token (bec_...)")
	flag.Parse()

	apiURL = strings.TrimRight(*urlFlag, "/")
	apiToken = *tokenFlag

	if apiToken == "" {
		fmt.Fprintln(os.Stderr, "Error: API token required. Use -token flag or BECOMING_TOKEN env var.")
		os.Exit(1)
	}

	// Log to file (stdout/stderr is the MCP transport)
	logFile, err := os.OpenFile("becoming-mcp.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	}

	s := server.NewMCPServer("Becoming", "1.0.0",
		server.WithToolCapabilities(true),
	)

	registerTools(s)

	log.Println("Becoming MCP server starting...")
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// --- API Client ---

func apiRequest(method, path string, body any) (json.RawMessage, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encoding body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, apiURL+"/api"+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}
	if resp.StatusCode == 204 {
		return json.RawMessage(`{"status":"ok"}`), nil
	}
	return respBody, nil
}

// --- Tool Registration ---

func registerTools(s *server.MCPServer) {
	// Read-only tools
	s.AddTool(toolGetToday(), handleGetToday)
	s.AddTool(toolListPractices(), handleListPractices)
	s.AddTool(toolGetDueCards(), handleGetDueCards)
	s.AddTool(toolListTasks(), handleListTasks)
	s.AddTool(toolListNotes(), handleListNotes)
	s.AddTool(toolGetReport(), handleGetReport)
	s.AddTool(toolGetReflection(), handleGetReflection)
	s.AddTool(toolGetTodayPrompt(), handleGetTodayPrompt)

	// Write tools
	s.AddTool(toolLogPractice(), handleLogPractice)
	s.AddTool(toolCreatePractice(), handleCreatePractice)
	s.AddTool(toolReviewCard(), handleReviewCard)
	s.AddTool(toolCreateTask(), handleCreateTask)
	s.AddTool(toolUpdateTask(), handleUpdateTask)
	s.AddTool(toolCreateNote(), handleCreateNote)
	s.AddTool(toolUpsertReflection(), handleUpsertReflection)

	// Brain tools — read
	s.AddTool(toolBrainSearch(), handleBrainSearch)
	s.AddTool(toolBrainRecent(), handleBrainRecent)
	s.AddTool(toolBrainGet(), handleBrainGet)
	s.AddTool(toolBrainStats(), handleBrainStats)
	s.AddTool(toolBrainTags(), handleBrainTags)

	// Brain tools — write
	s.AddTool(toolBrainCreate(), handleBrainCreate)
	s.AddTool(toolBrainUpdate(), handleBrainUpdate)
	s.AddTool(toolBrainDelete(), handleBrainDelete)
}

// --- Tool Definitions ---

func toolGetToday() mcp.Tool {
	return mcp.NewTool("get_today",
		mcp.WithDescription("Get today's daily summary — all practices with their log counts, notes, and schedule status. Use this as the starting point for any session to understand current progress."),
		mcp.WithString("date", mcp.Description("Date in YYYY-MM-DD format (default: today)")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func toolListPractices() mcp.Tool {
	return mcp.NewTool("list_practices",
		mcp.WithDescription("List all practices with their type, category, and active status."),
		mcp.WithString("type", mcp.Description("Filter by type: memorize, tracker, habit, scheduled")),
		mcp.WithBoolean("active_only", mcp.Description("Only return active practices (default: true)")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func toolGetDueCards() mcp.Tool {
	return mcp.NewTool("get_due_cards",
		mcp.WithDescription("Get memorization cards due for review today. Returns scripture references and their spaced repetition status."),
		mcp.WithString("date", mcp.Description("Date in YYYY-MM-DD format (default: today)")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func toolListTasks() mcp.Tool {
	return mcp.NewTool("list_tasks",
		mcp.WithDescription("List all tasks with their status, scripture references, and descriptions."),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func toolListNotes() mcp.Tool {
	return mcp.NewTool("list_notes",
		mcp.WithDescription("List notes, optionally filtered by practice, task, or pinned status."),
		mcp.WithBoolean("pinned_only", mcp.Description("Only return pinned notes")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func toolGetReport() mcp.Tool {
	return mcp.NewTool("get_report",
		mcp.WithDescription("Get a progress report for a date range. Shows practice completion data."),
		mcp.WithString("start", mcp.Required(), mcp.Description("Start date (YYYY-MM-DD)")),
		mcp.WithString("end", mcp.Required(), mcp.Description("End date (YYYY-MM-DD)")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func toolGetReflection() mcp.Tool {
	return mcp.NewTool("get_reflection",
		mcp.WithDescription("Get the daily reflection journal entry for a specific date."),
		mcp.WithString("date", mcp.Description("Date in YYYY-MM-DD format (default: today)")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func toolGetTodayPrompt() mcp.Tool {
	return mcp.NewTool("get_today_prompt",
		mcp.WithDescription("Get the daily reflection prompt for today. This is a rotating question to guide daily reflection."),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func toolLogPractice() mcp.Tool {
	return mcp.NewTool("log_practice",
		mcp.WithDescription("Log a practice for today. Use after completing a practice (e.g., scripture reading, exercise, study)."),
		mcp.WithNumber("practice_id", mcp.Required(), mcp.Description("ID of the practice to log")),
		mcp.WithString("date", mcp.Description("Date in YYYY-MM-DD format (default: today)")),
		mcp.WithString("notes", mcp.Description("Optional notes about this log entry")),
		mcp.WithString("value", mcp.Description("Freeform value (e.g., passage read, weight lifted)")),
		mcp.WithNumber("quality", mcp.Description("Quality rating 1-5 (for memorize type)")),
		mcp.WithNumber("sets", mcp.Description("Number of sets (for exercise type)")),
		mcp.WithNumber("reps", mcp.Description("Number of reps (for exercise type)")),
		mcp.WithNumber("duration_s", mcp.Description("Duration in seconds")),
	)
}

func toolCreatePractice() mcp.Tool {
	return mcp.NewTool("create_practice",
		mcp.WithDescription("Create a new practice to track. Types: 'memorize' (scripture memorization with spaced repetition), 'tracker' (simple counter), 'habit' (daily habit), 'scheduled' (time-based)."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Practice name")),
		mcp.WithString("type", mcp.Required(), mcp.Description("Practice type: memorize, tracker, habit, scheduled"),
			mcp.Enum("memorize", "tracker", "habit", "scheduled")),
		mcp.WithString("description", mcp.Description("Description of the practice")),
		mcp.WithString("category", mcp.Description("Category for grouping (e.g., 'spiritual', 'physical')")),
		mcp.WithString("config", mcp.Description("JSON config (for memorize: scripture reference, interval, etc.)")),
	)
}

func toolReviewCard() mcp.Tool {
	return mcp.NewTool("review_card",
		mcp.WithDescription("Submit a memorization card review. The quality score determines the next review interval using spaced repetition."),
		mcp.WithNumber("practice_id", mcp.Required(), mcp.Description("ID of the memorize practice")),
		mcp.WithNumber("quality", mcp.Required(), mcp.Description("Review quality: 0 (forgot), 1 (hard), 2 (ok), 3 (good), 4 (easy), 5 (perfect)")),
		mcp.WithString("date", mcp.Description("Date in YYYY-MM-DD format (default: today)")),
	)
}

func toolCreateTask() mcp.Tool {
	return mcp.NewTool("create_task",
		mcp.WithDescription("Create a task — an actionable to-do derived from scripture study or a prompting."),
		mcp.WithString("title", mcp.Required(), mcp.Description("Task title")),
		mcp.WithString("description", mcp.Description("Detailed description")),
		mcp.WithString("scripture", mcp.Description("Related scripture reference")),
		mcp.WithString("source_doc", mcp.Description("Source document or study topic")),
		mcp.WithString("type", mcp.Description("Task type: 'action', 'ongoing', 'reflection'"),
			mcp.Enum("action", "ongoing", "reflection")),
	)
}

func toolUpdateTask() mcp.Tool {
	return mcp.NewTool("update_task",
		mcp.WithDescription("Update a task's status or details."),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("Task ID")),
		mcp.WithString("status", mcp.Description("New status: active, completed, deferred"),
			mcp.Enum("active", "completed", "deferred")),
		mcp.WithString("title", mcp.Description("Updated title")),
		mcp.WithString("description", mcp.Description("Updated description")),
	)
}

func toolCreateNote() mcp.Tool {
	return mcp.NewTool("create_note",
		mcp.WithDescription("Create a note — a quick insight, cross-reference, or thought. Can optionally link to a practice, task, or pillar."),
		mcp.WithString("content", mcp.Required(), mcp.Description("Note content (supports markdown)")),
		mcp.WithNumber("practice_id", mcp.Description("Link to a practice")),
		mcp.WithNumber("task_id", mcp.Description("Link to a task")),
		mcp.WithNumber("pillar_id", mcp.Description("Link to a pillar")),
		mcp.WithBoolean("pinned", mcp.Description("Pin this note for quick access")),
	)
}

func toolUpsertReflection() mcp.Tool {
	return mcp.NewTool("upsert_reflection",
		mcp.WithDescription("Create or update today's daily reflection journal entry. If an entry exists for the date, it will be updated."),
		mcp.WithString("content", mcp.Required(), mcp.Description("Reflection content (supports markdown)")),
		mcp.WithString("date", mcp.Description("Date in YYYY-MM-DD format (default: today)")),
		mcp.WithNumber("mood", mcp.Description("Mood rating 1-5")),
		mcp.WithNumber("prompt_id", mcp.Description("ID of the prompt that inspired this reflection")),
		mcp.WithString("prompt_text", mcp.Description("Text of the prompt (stored with reflection)")),
	)
}

// --- Tool Handlers ---

func today() string {
	return time.Now().Format("2006-01-02")
}

func handleGetToday(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	date := req.GetString("date", today())
	data, err := apiRequest("GET", "/daily/"+date, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

func handleListPractices(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	params := url.Values{}
	if t := req.GetString("type", ""); t != "" {
		params.Set("type", t)
	}
	if req.GetBool("active_only", true) {
		params.Set("active", "true")
	}
	path := "/practices"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}
	data, err := apiRequest("GET", path, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

func handleGetDueCards(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	date := req.GetString("date", today())
	data, err := apiRequest("GET", "/memorize/due/"+date, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

func handleListTasks(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := apiRequest("GET", "/tasks", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

func handleListNotes(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	params := url.Values{}
	if req.GetBool("pinned_only", false) {
		params.Set("pinned", "true")
	}
	path := "/notes"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}
	data, err := apiRequest("GET", path, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

func handleGetReport(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := req.GetString("start", "")
	end := req.GetString("end", "")
	data, err := apiRequest("GET", "/reports?start="+start+"&end="+end, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

func handleGetReflection(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	date := req.GetString("date", today())
	data, err := apiRequest("GET", "/reflections/"+date, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

func handleGetTodayPrompt(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := apiRequest("GET", "/prompts/today", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

func handleLogPractice(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	body := map[string]any{
		"practice_id": int(req.GetFloat("practice_id", 0)),
		"date":        req.GetString("date", today()),
	}
	if v := req.GetString("notes", ""); v != "" {
		body["notes"] = v
	}
	if v := req.GetString("value", ""); v != "" {
		body["value"] = v
	}
	if v := req.GetFloat("quality", 0); v > 0 {
		body["quality"] = int(v)
	}
	if v := req.GetFloat("sets", 0); v > 0 {
		body["sets"] = int(v)
	}
	if v := req.GetFloat("reps", 0); v > 0 {
		body["reps"] = int(v)
	}
	if v := req.GetFloat("duration_s", 0); v > 0 {
		body["duration_s"] = int(v)
	}

	data, err := apiRequest("POST", "/logs", body)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

func handleCreatePractice(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	body := map[string]any{
		"name":   req.GetString("name", ""),
		"type":   req.GetString("type", "tracker"),
		"active": true,
	}
	if v := req.GetString("description", ""); v != "" {
		body["description"] = v
	}
	if v := req.GetString("category", ""); v != "" {
		body["category"] = v
	}
	if v := req.GetString("config", ""); v != "" {
		body["config"] = v
	}

	data, err := apiRequest("POST", "/practices", body)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

func handleReviewCard(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	body := map[string]any{
		"practice_id": int(req.GetFloat("practice_id", 0)),
		"quality":     int(req.GetFloat("quality", 0)),
		"date":        req.GetString("date", today()),
	}

	data, err := apiRequest("POST", "/memorize/review", body)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

func handleCreateTask(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	body := map[string]any{
		"title":  req.GetString("title", ""),
		"status": "active",
	}
	if v := req.GetString("description", ""); v != "" {
		body["description"] = v
	}
	if v := req.GetString("scripture", ""); v != "" {
		body["scripture"] = v
	}
	if v := req.GetString("source_doc", ""); v != "" {
		body["source_doc"] = v
	}
	if v := req.GetString("type", ""); v != "" {
		body["type"] = v
	}

	data, err := apiRequest("POST", "/tasks", body)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

func handleUpdateTask(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id := int(req.GetFloat("id", 0))
	if id == 0 {
		return mcp.NewToolResultError("task id is required"), nil
	}

	body := map[string]any{}
	if v := req.GetString("status", ""); v != "" {
		body["status"] = v
	}
	if v := req.GetString("title", ""); v != "" {
		body["title"] = v
	}
	if v := req.GetString("description", ""); v != "" {
		body["description"] = v
	}

	data, err := apiRequest("PUT", fmt.Sprintf("/tasks/%d", id), body)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

func handleCreateNote(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	body := map[string]any{
		"content": req.GetString("content", ""),
	}
	if v := req.GetFloat("practice_id", 0); v > 0 {
		body["practice_id"] = int(v)
	}
	if v := req.GetFloat("task_id", 0); v > 0 {
		body["task_id"] = int(v)
	}
	if v := req.GetFloat("pillar_id", 0); v > 0 {
		body["pillar_id"] = int(v)
	}
	if req.GetBool("pinned", false) {
		body["pinned"] = true
	}

	data, err := apiRequest("POST", "/notes", body)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

func handleUpsertReflection(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	body := map[string]any{
		"content": req.GetString("content", ""),
		"date":    req.GetString("date", today()),
	}
	if v := req.GetFloat("mood", 0); v > 0 {
		body["mood"] = int(v)
	}
	if v := req.GetFloat("prompt_id", 0); v > 0 {
		body["prompt_id"] = int(v)
	}
	if v := req.GetString("prompt_text", ""); v != "" {
		body["prompt_text"] = v
	}

	data, err := apiRequest("POST", "/reflections", body)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// --- Brain Tool Definitions ---

func toolBrainSearch() mcp.Tool {
	return mcp.NewTool("brain_search",
		mcp.WithDescription("Search brain entries by text. Returns matching entries from the synced brain cache. Use to find entries by keyword, topic, or content."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search text to match against entry titles and bodies")),
		mcp.WithString("category", mcp.Description("Filter by category: inbox, actions, projects, ideas, people, study, journal")),
		mcp.WithNumber("limit", mcp.Description("Max results to return (default: 20)")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func toolBrainRecent() mcp.Tool {
	return mcp.NewTool("brain_recent",
		mcp.WithDescription("Get recent brain entries, newest first. Optionally filter by category."),
		mcp.WithString("category", mcp.Description("Filter by category: inbox, actions, projects, ideas, people, study, journal")),
		mcp.WithNumber("limit", mcp.Description("Max entries to return (default: 20)")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func toolBrainGet() mcp.Tool {
	return mcp.NewTool("brain_get",
		mcp.WithDescription("Get a single brain entry by ID. Returns all fields including body, tags, status, due date, and next action."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Entry UUID")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func toolBrainStats() mcp.Tool {
	return mcp.NewTool("brain_stats",
		mcp.WithDescription("Get brain relay status — whether the agent is online, and message queue depths."),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func toolBrainTags() mcp.Tool {
	return mcp.NewTool("brain_tags",
		mcp.WithDescription("List all tags used across brain entries with usage counts."),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func toolBrainCreate() mcp.Tool {
	return mcp.NewTool("brain_create",
		mcp.WithDescription("Create a new brain entry. The entry is saved to the relay cache and forwarded to brain.exe for classification when the agent is online."),
		mcp.WithString("title", mcp.Required(), mcp.Description("Entry title")),
		mcp.WithString("body", mcp.Description("Entry body (supports markdown)")),
		mcp.WithString("category", mcp.Description("Category: inbox, actions, projects, ideas, people, study, journal (default: inbox)")),
		mcp.WithString("status", mcp.Description("Status: active, done, waiting, blocked, someday")),
		mcp.WithString("due_date", mcp.Description("Due date in YYYY-MM-DD format")),
		mcp.WithString("next_action", mcp.Description("Next action text")),
		mcp.WithString("tags", mcp.Description("Comma-separated tags")),
	)
}

func toolBrainUpdate() mcp.Tool {
	return mcp.NewTool("brain_update",
		mcp.WithDescription("Update an existing brain entry. Only provided fields are changed. The update is saved locally and forwarded to brain.exe."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Entry UUID to update")),
		mcp.WithString("title", mcp.Description("New title")),
		mcp.WithString("body", mcp.Description("New body")),
		mcp.WithString("category", mcp.Description("New category")),
		mcp.WithString("status", mcp.Description("New status")),
		mcp.WithBoolean("action_done", mcp.Description("Mark as done (true) or not done (false)")),
		mcp.WithString("due_date", mcp.Description("New due date (YYYY-MM-DD)")),
		mcp.WithString("tags", mcp.Description("Comma-separated tags (replaces existing)")),
	)
}

func toolBrainDelete() mcp.Tool {
	return mcp.NewTool("brain_delete",
		mcp.WithDescription("Delete a brain entry by ID. Removes from relay cache and forwards delete to brain.exe."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Entry UUID to delete")),
	)
}

// --- Brain Tool Handlers ---

func handleBrainSearch(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := req.GetString("query", "")
	if query == "" {
		return mcp.NewToolResultError("query is required"), nil
	}

	// Fetch all entries (with optional category filter) and search client-side.
	// TODO: Add server-side search endpoint to ibeco.me for better performance.
	params := url.Values{}
	if cat := req.GetString("category", ""); cat != "" {
		params.Set("category", cat)
	}
	path := "/brain/entries"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	data, err := apiRequest("GET", path, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Parse entries and filter by query
	var resp struct {
		Entries     []json.RawMessage `json:"entries"`
		AgentOnline bool              `json:"agent_online"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("parsing response: %v", err)), nil
	}

	limit := int(req.GetFloat("limit", 20))
	queryLower := strings.ToLower(query)
	var matches []json.RawMessage

	for _, raw := range resp.Entries {
		var entry struct {
			Title string   `json:"title"`
			Body  string   `json:"body"`
			Tags  []string `json:"tags"`
		}
		if err := json.Unmarshal(raw, &entry); err != nil {
			continue
		}
		if strings.Contains(strings.ToLower(entry.Title), queryLower) ||
			strings.Contains(strings.ToLower(entry.Body), queryLower) ||
			tagsContain(entry.Tags, queryLower) {
			matches = append(matches, raw)
			if len(matches) >= limit {
				break
			}
		}
	}

	result := map[string]any{
		"entries":      matches,
		"total":        len(matches),
		"agent_online": resp.AgentOnline,
		"search_type":  "text",
	}
	out, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(out)), nil
}

func tagsContain(tags []string, query string) bool {
	for _, t := range tags {
		if strings.Contains(strings.ToLower(t), query) {
			return true
		}
	}
	return false
}

func handleBrainRecent(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	params := url.Values{}
	if cat := req.GetString("category", ""); cat != "" {
		params.Set("category", cat)
	}
	path := "/brain/entries"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	data, err := apiRequest("GET", path, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Trim to limit
	limit := int(req.GetFloat("limit", 20))
	var resp struct {
		Entries     []json.RawMessage `json:"entries"`
		AgentOnline bool              `json:"agent_online"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return mcp.NewToolResultText(string(data)), nil
	}
	if len(resp.Entries) > limit {
		resp.Entries = resp.Entries[:limit]
	}
	out, _ := json.Marshal(resp)
	return mcp.NewToolResultText(string(out)), nil
}

func handleBrainGet(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id := req.GetString("id", "")
	if id == "" {
		return mcp.NewToolResultError("id is required"), nil
	}

	// Fetch all entries and find the one with matching ID
	data, err := apiRequest("GET", "/brain/entries", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var resp struct {
		Entries []json.RawMessage `json:"entries"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("parsing response: %v", err)), nil
	}

	for _, raw := range resp.Entries {
		var entry struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(raw, &entry); err != nil {
			continue
		}
		if entry.ID == id {
			return mcp.NewToolResultText(string(raw)), nil
		}
	}

	return mcp.NewToolResultError(fmt.Sprintf("entry %s not found", id)), nil
}

func handleBrainStats(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := apiRequest("GET", "/brain/status", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

func handleBrainTags(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := apiRequest("GET", "/brain/entries", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var resp struct {
		Entries []struct {
			Tags []string `json:"tags"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("parsing response: %v", err)), nil
	}

	counts := map[string]int{}
	for _, e := range resp.Entries {
		for _, t := range e.Tags {
			if t != "" {
				counts[t]++
			}
		}
	}

	out, _ := json.Marshal(counts)
	return mcp.NewToolResultText(string(out)), nil
}

func handleBrainCreate(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	body := map[string]any{
		"title": req.GetString("title", ""),
	}
	if v := req.GetString("body", ""); v != "" {
		body["body"] = v
	}
	if v := req.GetString("category", ""); v != "" {
		body["category"] = v
	}
	if v := req.GetString("status", ""); v != "" {
		body["status"] = v
	}
	if v := req.GetString("due_date", ""); v != "" {
		body["due_date"] = v
	}
	if v := req.GetString("next_action", ""); v != "" {
		body["next_action"] = v
	}
	if v := req.GetString("tags", ""); v != "" {
		tags := strings.Split(v, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
		body["tags"] = tags
	}

	data, err := apiRequest("POST", "/brain/entries", body)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

func handleBrainUpdate(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id := req.GetString("id", "")
	if id == "" {
		return mcp.NewToolResultError("id is required"), nil
	}

	body := map[string]any{}
	if v := req.GetString("title", ""); v != "" {
		body["title"] = v
	}
	if v := req.GetString("body", ""); v != "" {
		body["body"] = v
	}
	if v := req.GetString("category", ""); v != "" {
		body["category"] = v
	}
	if v := req.GetString("status", ""); v != "" {
		body["status"] = v
	}
	if v := req.GetString("due_date", ""); v != "" {
		body["due_date"] = v
	}
	if v := req.GetString("tags", ""); v != "" {
		tags := strings.Split(v, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
		body["tags"] = tags
	}
	// action_done is a bool — check if it was explicitly set
	if argsMap, ok := req.Params.Arguments.(map[string]any); ok {
		if done, ok := argsMap["action_done"]; ok {
			if b, ok := done.(bool); ok {
				body["action_done"] = b
			}
		}
	}

	if len(body) == 0 {
		return mcp.NewToolResultError("no fields to update"), nil
	}

	data, err := apiRequest("PUT", "/brain/entries?id="+url.QueryEscape(id), body)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

func handleBrainDelete(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id := req.GetString("id", "")
	if id == "" {
		return mcp.NewToolResultError("id is required"), nil
	}

	data, err := apiRequest("DELETE", "/brain/entries?id="+url.QueryEscape(id), nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}
