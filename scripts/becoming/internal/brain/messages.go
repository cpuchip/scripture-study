// Package brain implements the WebSocket relay hub for the second brain system.
// The hub routes messages between the mobile app and brain.exe agent,
// queuing messages when either side is offline.
package brain

import (
	"encoding/json"
	"fmt"
	"time"
)

// Message types for the WebSocket protocol.
const (
	TypeAuth        = "auth"
	TypeAuthOK      = "auth_ok"
	TypeAuthError   = "auth_error"
	TypeThought     = "thought"
	TypeResult      = "result"
	TypeFix         = "fix"
	TypeFixOK       = "fix_ok"
	TypePresence    = "presence"
	TypeQueued      = "queued"
	TypeStatus      = "status"
	TypePing        = "ping"
	TypePong        = "pong"
	TypeTaskUpdated = "task_updated"
	TypeEntriesSync = "entries_sync"
	TypeEntryUpdate = "entry_update"
	TypeEntryDelete = "entry_delete"
)

// Client roles.
const (
	RoleApp   = "app"
	RoleAgent = "agent"
)

// Envelope is the base message structure — all WS messages have a type.
type Envelope struct {
	Type string `json:"type"`
}

// AuthMessage is sent by clients immediately after connecting.
type AuthMessage struct {
	Type  string `json:"type"`  // "auth"
	Token string `json:"token"` // "bec_..."
	Role  string `json:"role"`  // "app" or "agent"
}

// AuthOKMessage confirms successful authentication.
type AuthOKMessage struct {
	Type   string `json:"type"` // "auth_ok"
	UserID int64  `json:"user_id"`
}

// AuthErrorMessage reports authentication failure.
type AuthErrorMessage struct {
	Type  string `json:"type"` // "auth_error"
	Error string `json:"error"`
}

// ThoughtMessage is a new thought from the app to classify.
type ThoughtMessage struct {
	Type      string `json:"type"`                // "thought"
	ID        string `json:"id"`                  // UUID
	Text      string `json:"text"`                // Raw thought text
	Timestamp string `json:"timestamp"`           // ISO 8601
	Source    string `json:"source,omitempty"`    // "phone", "cli", etc.
	Workspace string `json:"workspace,omitempty"` // workspace/repo context
}

// ResultMessage is the classification result from the agent.
type ResultMessage struct {
	Type        string   `json:"type"`         // "result"
	ThoughtID   string   `json:"thought_id"`   // UUID of the original thought
	Category    string   `json:"category"`     // "people", "projects", etc.
	Title       string   `json:"title"`        // Generated title
	Confidence  float64  `json:"confidence"`   // 0.0-1.0
	Tags        []string `json:"tags"`         // Auto-generated tags
	NeedsReview bool     `json:"needs_review"` // Below confidence threshold
	FilePath    string   `json:"file_path"`    // Where it was stored
}

// FixMessage requests reclassification.
type FixMessage struct {
	Type        string `json:"type"`         // "fix"
	ThoughtID   string `json:"thought_id"`   // UUID of the thought to fix
	NewCategory string `json:"new_category"` // Target category
}

// FixOKMessage confirms reclassification.
type FixOKMessage struct {
	Type      string `json:"type"` // "fix_ok"
	ThoughtID string `json:"thought_id"`
	NewPath   string `json:"new_path"` // New file path after move
}

// PresenceMessage reports agent online/offline status.
type PresenceMessage struct {
	Type        string `json:"type"` // "presence"
	AgentOnline bool   `json:"agent_online"`
}

// QueuedMessage delivers messages that were queued while the recipient was offline.
type QueuedMessage struct {
	Type     string            `json:"type"` // "queued"
	Messages []json.RawMessage `json:"messages"`
}

// StatusMessage is sent by the agent to report its current state.
type StatusMessage struct {
	Type       string         `json:"type"` // "status"
	Model      string         `json:"model"`
	Categories map[string]int `json:"categories"`
}

// TaskUpdatedMessage notifies the agent that a task's status changed.
// Sent server→agent when a task with a brain_entry_id is updated via the REST API.
type TaskUpdatedMessage struct {
	Type         string `json:"type"`           // "task_updated"
	TaskID       int64  `json:"task_id"`        // ibecome task ID
	BrainEntryID string `json:"brain_entry_id"` // brain entry UUID
	Status       string `json:"status"`         // new ibecome status
	Title        string `json:"title"`          // current title
}

// EntriesSyncMessage is sent by the agent with all brain entries.
// The hub stores them in the brain_entries table for web UI access.
type EntriesSyncMessage struct {
	Type    string           `json:"type"` // "entries_sync"
	Entries []SyncEntryPayload `json:"entries"`
}

// SyncEntryPayload is a single brain entry in the sync payload.
type SyncEntryPayload struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	Category   string   `json:"category"`
	Body       string   `json:"body"`
	Status     string   `json:"status,omitempty"`
	ActionDone bool     `json:"action_done,omitempty"`
	DueDate    string   `json:"due_date,omitempty"`
	NextAction string   `json:"next_action,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	Source     string   `json:"source,omitempty"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
}

// EntryUpdateMessage requests the agent update a brain entry.
// Sent ibeco.me→agent via relay when a user edits an entry in the web UI.
type EntryUpdateMessage struct {
	Type    string            `json:"type"` // "entry_update"
	EntryID string            `json:"entry_id"`
	Updates map[string]any    `json:"updates"`
}

// EntryDeleteMessage requests the agent delete a brain entry.
type EntryDeleteMessage struct {
	Type    string `json:"type"` // "entry_delete"
	EntryID string `json:"entry_id"`
}

// Direction indicates which way a queued message should be routed.
type Direction string

const (
	ToAgent Direction = "to_agent"
	ToApp   Direction = "to_app"
)

// QueueEntry is a database record for queued messages.
type QueueEntry struct {
	ID          int64      `json:"id"`
	MessageID   string     `json:"message_id"`
	UserID      int64      `json:"user_id"`
	Direction   Direction  `json:"direction"`
	Payload     string     `json:"payload"`
	Status      string     `json:"status"` // "pending" or "delivered"
	CreatedAt   time.Time  `json:"created_at"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
}

// ParseEnvelope extracts just the type field from a raw JSON message.
func ParseEnvelope(data []byte) (string, error) {
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return "", fmt.Errorf("parsing envelope: %w", err)
	}
	if env.Type == "" {
		return "", fmt.Errorf("message has no type field")
	}
	return env.Type, nil
}

// MarshalJSON helper that wraps json.Marshal with error context.
func MarshalJSON(v any) ([]byte, error) {
	return json.Marshal(v)
}
