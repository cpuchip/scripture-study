package brain

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/cpuchip/scripture-study/scripts/becoming/internal/auth"
	"github.com/cpuchip/scripture-study/scripts/becoming/internal/db"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Auth happens after upgrade via token message
	},
}

// conn represents an authenticated WebSocket connection.
type conn struct {
	ws     *websocket.Conn
	userID int64
	role   string // "app" or "agent"
	send   chan []byte
}

// Hub manages WebSocket connections and message routing for the brain relay.
type Hub struct {
	db    *db.DB
	queue *Queue

	mu    sync.RWMutex
	conns map[*conn]struct{} // all active connections

	// Per-user connection tracking
	agents map[int64]*conn // userID -> agent connection
	apps   map[int64]*conn // userID -> app connection (most recent)
}

// NewHub creates a new brain relay hub.
func NewHub(database *db.DB) *Hub {
	q := NewQueue(database)
	if err := q.EnsureTable(); err != nil {
		log.Printf("[brain] warning: could not ensure brain_messages table: %v", err)
	}
	if err := database.EnsureBrainEntriesTable(); err != nil {
		log.Printf("[brain] warning: could not ensure brain_entries table: %v", err)
	}

	return &Hub{
		db:     database,
		queue:  q,
		conns:  make(map[*conn]struct{}),
		agents: make(map[int64]*conn),
		apps:   make(map[int64]*conn),
	}
}

// HandleWebSocket upgrades an HTTP connection to WebSocket and manages the lifecycle.
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[brain] websocket upgrade failed: %v", err)
		return
	}

	c := &conn{
		ws:   ws,
		send: make(chan []byte, 64),
	}

	// Give the client 10 seconds to authenticate
	ws.SetReadDeadline(time.Now().Add(10 * time.Second))

	_, data, err := ws.ReadMessage()
	if err != nil {
		log.Printf("[brain] failed to read auth message: %v", err)
		ws.Close()
		return
	}

	var auth AuthMessage
	if err := json.Unmarshal(data, &auth); err != nil || auth.Type != TypeAuth {
		h.sendError(ws, "first message must be auth")
		ws.Close()
		return
	}

	if auth.Role != RoleApp && auth.Role != RoleAgent {
		h.sendError(ws, "role must be 'app' or 'agent'")
		ws.Close()
		return
	}

	// Validate the bearer token
	token, err := h.db.ValidateAPIToken(auth.Token)
	if err != nil {
		log.Printf("[brain] token validation error: %v", err)
		h.sendError(ws, "token validation failed")
		ws.Close()
		return
	}
	if token == nil {
		h.sendError(ws, "invalid token")
		ws.Close()
		return
	}

	// Touch the token's last_used timestamp
	h.db.TouchAPIToken(token.ID)

	c.userID = token.UserID
	c.role = auth.Role

	// Send auth_ok
	okMsg, _ := json.Marshal(AuthOKMessage{Type: TypeAuthOK, UserID: token.UserID})
	ws.WriteMessage(websocket.TextMessage, okMsg)

	// Register connection
	h.register(c)
	defer h.unregister(c)

	log.Printf("[brain] %s connected (user %d)", c.role, c.userID)

	// Deliver queued messages
	h.deliverQueued(c)

	// Notify about presence
	if c.role == RoleAgent {
		// Agent came online — tell the app
		h.notifyPresence(c.userID, true)
	} else if c.role == RoleApp {
		// App connected — tell it whether agent is online
		h.mu.RLock()
		_, agentOnline := h.agents[c.userID]
		h.mu.RUnlock()
		h.notifyPresence(c.userID, agentOnline)
	}

	// Reset read deadline and configure keepalive
	ws.SetReadDeadline(time.Time{}) // no deadline
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Start writer goroutine
	go h.writePump(c)

	// Start ping ticker
	go h.pingPump(c)

	// Read loop
	h.readPump(c)
}

// register adds a connection to the hub.
func (h *Hub) register(c *conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.conns[c] = struct{}{}

	switch c.role {
	case RoleAgent:
		// If there's an existing agent connection, close it
		if old, ok := h.agents[c.userID]; ok {
			close(old.send)
			old.ws.Close()
			delete(h.conns, old)
		}
		h.agents[c.userID] = c
	case RoleApp:
		// Allow newest app connection (previous stays connected for read-only)
		h.apps[c.userID] = c
	}
}

// unregister removes a connection from the hub.
func (h *Hub) unregister(c *conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.conns, c)

	switch c.role {
	case RoleAgent:
		if h.agents[c.userID] == c {
			delete(h.agents, c.userID)
			log.Printf("[brain] agent disconnected (user %d)", c.userID)
			// Notify app that agent went offline
			go h.notifyPresence(c.userID, false)
		}
	case RoleApp:
		if h.apps[c.userID] == c {
			delete(h.apps, c.userID)
			log.Printf("[brain] app disconnected (user %d)", c.userID)
		}
	}

	c.ws.Close()
}

// readPump reads messages from the WebSocket and routes them.
func (h *Hub) readPump(c *conn) {
	defer func() {
		// unregister is called by deferred in HandleWebSocket
	}()

	for {
		_, data, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("[brain] read error (%s, user %d): %v", c.role, c.userID, err)
			}
			return
		}

		msgType, err := ParseEnvelope(data)
		if err != nil {
			log.Printf("[brain] invalid message from %s (user %d): %v", c.role, c.userID, err)
			continue
		}

		h.routeMessage(c, msgType, data)
	}
}

// writePump sends messages from the send channel to the WebSocket.
func (h *Hub) writePump(c *conn) {
	for msg := range c.send {
		if err := c.ws.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Printf("[brain] write error (%s, user %d): %v", c.role, c.userID, err)
			return
		}
	}
}

// pingPump sends periodic pings to keep the connection alive.
func (h *Hub) pingPump(c *conn) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		h.mu.RLock()
		_, exists := h.conns[c]
		h.mu.RUnlock()

		if !exists {
			return
		}

		if err := c.ws.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil {
			return
		}
	}
}

// routeMessage handles an incoming message based on its type and the sender's role.
func (h *Hub) routeMessage(sender *conn, msgType string, data []byte) {
	switch msgType {
	case TypeThought:
		// App -> Agent
		if sender.role != RoleApp {
			log.Printf("[brain] ignoring thought from non-app role")
			return
		}
		// Extract message ID for queue tracking
		var thought ThoughtMessage
		if err := json.Unmarshal(data, &thought); err != nil {
			log.Printf("[brain] invalid thought message: %v", err)
			return
		}
		h.routeToAgent(sender.userID, thought.ID, data)

	case TypeResult, TypeFixOK:
		// Agent -> App
		if sender.role != RoleAgent {
			log.Printf("[brain] ignoring result/fix_ok from non-agent role")
			return
		}
		// Extract thought_id for queue tracking
		var result struct {
			ThoughtID string `json:"thought_id"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			log.Printf("[brain] invalid result message: %v", err)
			return
		}
		h.routeToApp(sender.userID, result.ThoughtID, data)

	case TypeFix:
		// App -> Agent
		if sender.role != RoleApp {
			return
		}
		var fix FixMessage
		if err := json.Unmarshal(data, &fix); err != nil {
			log.Printf("[brain] invalid fix message: %v", err)
			return
		}
		h.routeToAgent(sender.userID, fix.ThoughtID, data)

	case TypeStatus:
		// Agent -> broadcast to app
		if sender.role != RoleAgent {
			return
		}
		h.sendToApp(sender.userID, data)

	case TypeEntriesSync:
		// Agent -> server: store all entries in brain_entries table
		if sender.role != RoleAgent {
			return
		}
		go h.handleEntriesSync(sender.userID, data)

	case TypeEntryCreated:
		// Agent -> server: a new entry was created (real-time push)
		if sender.role != RoleAgent {
			return
		}
		go h.handleEntryCreated(sender.userID, data)

	case TypeEntryUpdated:
		// Agent -> server: an entry was updated (real-time push)
		if sender.role != RoleAgent {
			return
		}
		go h.handleEntryUpdated(sender.userID, data)

	case TypePing:
		// Respond with pong
		pong, _ := json.Marshal(Envelope{Type: TypePong})
		sender.send <- pong

	case TypeSubTaskCreate, TypeSubTaskUpdate, TypeSubTaskDelete:
		// These are relayed from the REST handlers, not from WS clients.
		// If an agent sends entry_updated after processing, it goes through TypeEntryUpdated.
		log.Printf("[brain] unexpected subtask message on WS: %s", msgType)

	default:
		log.Printf("[brain] unknown message type: %s", msgType)
	}
}

// routeToAgent sends a message to the user's agent, or queues it.
func (h *Hub) routeToAgent(userID int64, messageID string, data []byte) {
	h.mu.RLock()
	agent, online := h.agents[userID]
	h.mu.RUnlock()

	if online {
		select {
		case agent.send <- data:
			log.Printf("[brain] relayed to agent (user %d, msg %s)", userID, messageID)
		default:
			log.Printf("[brain] agent send buffer full, queuing (user %d)", userID)
			h.enqueue(messageID, userID, ToAgent, data)
		}
	} else {
		log.Printf("[brain] agent offline, queuing (user %d, msg %s)", userID, messageID)
		h.enqueue(messageID, userID, ToAgent, data)

		// Let the app know the agent is offline
		h.notifyPresence(userID, false)
	}
}

// routeToApp sends a message to the user's app, or queues it.
func (h *Hub) routeToApp(userID int64, messageID string, data []byte) {
	h.mu.RLock()
	app, online := h.apps[userID]
	h.mu.RUnlock()

	if online {
		select {
		case app.send <- data:
			log.Printf("[brain] relayed to app (user %d, msg %s)", userID, messageID)
		default:
			log.Printf("[brain] app send buffer full, queuing (user %d)", userID)
			h.enqueue(messageID, userID, ToApp, data)
		}
	} else {
		log.Printf("[brain] app offline, queuing (user %d, msg %s)", userID, messageID)
		h.enqueue(messageID, userID, ToApp, data)
	}
}

// sendToApp sends data directly to the app without queuing (for transient messages like status).
func (h *Hub) sendToApp(userID int64, data []byte) {
	h.mu.RLock()
	app, online := h.apps[userID]
	h.mu.RUnlock()

	if online {
		select {
		case app.send <- data:
		default:
		}
	}
}

// enqueue stores a message in the persistent queue.
func (h *Hub) enqueue(messageID string, userID int64, dir Direction, data []byte) {
	if err := h.queue.Enqueue(messageID, userID, dir, data); err != nil {
		log.Printf("[brain] queue error: %v", err)
	}
}

// deliverQueued sends all pending queued messages to a newly connected client.
func (h *Hub) deliverQueued(c *conn) {
	var dir Direction
	switch c.role {
	case RoleAgent:
		dir = ToAgent
	case RoleApp:
		dir = ToApp
	default:
		return
	}

	payloads, err := h.queue.DequeueAll(c.userID, dir)
	if err != nil {
		log.Printf("[brain] dequeue error (user %d, %s): %v", c.userID, dir, err)
		return
	}

	if len(payloads) == 0 {
		return
	}

	log.Printf("[brain] delivering %d queued messages to %s (user %d)", len(payloads), c.role, c.userID)

	// Send as a queued bundle
	rawMessages := make([]json.RawMessage, len(payloads))
	for i, p := range payloads {
		rawMessages[i] = json.RawMessage(p)
	}

	queuedMsg, _ := json.Marshal(QueuedMessage{
		Type:     TypeQueued,
		Messages: rawMessages,
	})

	select {
	case c.send <- queuedMsg:
	default:
		log.Printf("[brain] warning: could not deliver queued messages, send buffer full")
	}
}

// notifyPresence sends a presence message to the user's app connection.
func (h *Hub) notifyPresence(userID int64, agentOnline bool) {
	msg, _ := json.Marshal(PresenceMessage{
		Type:        TypePresence,
		AgentOnline: agentOnline,
	})
	h.sendToApp(userID, msg)
}

// sendError sends an auth_error message directly on the websocket (pre-registration).
func (h *Hub) sendError(ws *websocket.Conn, errMsg string) {
	data, _ := json.Marshal(AuthErrorMessage{
		Type:  TypeAuthError,
		Error: errMsg,
	})
	ws.WriteMessage(websocket.TextMessage, data)
}

// IsAgentOnline checks if a user's agent is currently connected.
func (h *Hub) IsAgentOnline(userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.agents[userID]
	return ok
}

// NotifyAgent sends a message to a user's agent via the relay.
// If the agent is offline, the message is queued for delivery on reconnect.
// messageID is used for queue deduplication.
func (h *Hub) NotifyAgent(userID int64, messageID string, data []byte) {
	h.routeToAgent(userID, messageID, data)
}

// HandleHistory returns recent brain messages as JSON.
// Unpacks payloads into the format the Flutter brain-app expects:
// {"messages": [{"id": "...", "text": "...", "category": "...", "title": "...", "confidence": 0.9, "created_at": "...", "processed": true}]}
func (h *Hub) HandleHistory(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserID(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}

	entries, err := h.queue.History(userID, limit)
	if err != nil {
		log.Printf("[brain] history error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Build a map of thought_id → result so we can merge them
	type resultInfo struct {
		Category   string  `json:"category"`
		Title      string  `json:"title"`
		Confidence float64 `json:"confidence"`
	}
	results := make(map[string]resultInfo)
	for _, e := range entries {
		var envelope struct {
			Type       string  `json:"type"`
			ThoughtID  string  `json:"thought_id"`
			Category   string  `json:"category"`
			Title      string  `json:"title"`
			Confidence float64 `json:"confidence"`
		}
		if err := json.Unmarshal([]byte(e.Payload), &envelope); err != nil {
			continue
		}
		if envelope.Type == "result" && envelope.ThoughtID != "" {
			results[envelope.ThoughtID] = resultInfo{
				Category:   envelope.Category,
				Title:      envelope.Title,
				Confidence: envelope.Confidence,
			}
		}
	}

	// Build output: one entry per thought message
	type historyMsg struct {
		ID         string  `json:"id"`
		Text       string  `json:"text"`
		Category   string  `json:"category"`
		Title      string  `json:"title"`
		Confidence float64 `json:"confidence"`
		CreatedAt  string  `json:"created_at"`
		Processed  bool    `json:"processed"`
	}
	var messages []historyMsg
	for _, e := range entries {
		var envelope struct {
			Type string `json:"type"`
			ID   string `json:"id"`
			Text string `json:"text"`
		}
		if err := json.Unmarshal([]byte(e.Payload), &envelope); err != nil {
			continue
		}
		if envelope.Type != "thought" {
			continue
		}
		msg := historyMsg{
			ID:        envelope.ID,
			Text:      envelope.Text,
			CreatedAt: e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if ri, ok := results[envelope.ID]; ok {
			msg.Category = ri.Category
			msg.Title = ri.Title
			msg.Confidence = ri.Confidence
			msg.Processed = true
		}
		messages = append(messages, msg)
	}
	if messages == nil {
		messages = []historyMsg{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"messages": messages,
	})
}

// HandleStatus returns brain relay status (agent online, queue counts).
func (h *Hub) HandleStatus(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserID(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	toAgent, toApp, err := h.queue.PendingCount(userID)
	if err != nil {
		log.Printf("[brain] status error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"agent_online":     h.IsAgentOnline(userID),
		"pending_to_agent": toAgent,
		"pending_to_app":   toApp,
	})
}

// handleEntriesSync processes an entries_sync message from the agent,
// storing all brain entries in the database for web UI access.
func (h *Hub) handleEntriesSync(userID int64, data []byte) {
	var msg EntriesSyncMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("[brain] invalid entries_sync: %v", err)
		return
	}

	entries := make([]*db.BrainEntry, len(msg.Entries))
	for i, e := range msg.Entries {
		entries[i] = &db.BrainEntry{
			ID:         e.ID,
			Title:      e.Title,
			Category:   e.Category,
			Body:       e.Body,
			Status:     e.Status,
			ActionDone: e.ActionDone,
			DueDate:    e.DueDate,
			NextAction: e.NextAction,
			Tags:       e.Tags,
			SubTasks:   syncSubTasksToDB(e.SubTasks),
			Source:     e.Source,
			CreatedAt:  e.CreatedAt,
			UpdatedAt:  e.UpdatedAt,
		}
	}

	if err := h.db.BulkUpsertBrainEntries(userID, entries); err != nil {
		log.Printf("[brain] entries_sync error (user %d): %v", userID, err)
		return
	}

	log.Printf("[brain] synced %d entries from agent (user %d)", len(entries), userID)
}

// HandleBrainEntries returns cached brain entries as JSON.
func (h *Hub) HandleBrainEntries(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserID(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	category := r.URL.Query().Get("category")
	entries, err := h.db.ListBrainEntries(userID, category)
	if err != nil {
		log.Printf("[brain] brain entries error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if entries == nil {
		entries = []*db.BrainEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"entries":      entries,
		"agent_online": h.IsAgentOnline(userID),
	})
}

// HandleBrainEntryUpdate proxies an entry update through the relay to the agent.
func (h *Hub) HandleBrainEntryUpdate(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserID(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	entryID := r.URL.Query().Get("id")
	if entryID == "" {
		http.Error(w, "missing entry id", http.StatusBadRequest)
		return
	}

	var updates map[string]any
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Update local cache immediately for responsiveness
	entry, err := h.db.GetBrainEntry(userID, entryID)
	if err != nil {
		http.Error(w, "entry not found", http.StatusNotFound)
		return
	}

	if v, ok := updates["title"].(string); ok {
		entry.Title = v
	}
	if v, ok := updates["status"].(string); ok {
		entry.Status = v
	}
	if v, ok := updates["action_done"].(bool); ok {
		entry.ActionDone = v
	}
	if v, ok := updates["due_date"].(string); ok {
		entry.DueDate = v
	}
	if v, ok := updates["category"].(string); ok {
		entry.Category = v
	}
	if v, ok := updates["body"].(string); ok {
		entry.Body = v
	}

	if err := h.db.UpsertBrainEntry(userID, entry); err != nil {
		log.Printf("[brain] cache update error: %v", err)
	}

	// Send update to agent via relay
	msg, _ := json.Marshal(EntryUpdateMessage{
		Type:    TypeEntryUpdate,
		EntryID: entryID,
		Updates: updates,
	})
	msgID := "entry_update_" + entryID + "_" + time.Now().Format("150405")
	h.routeToAgent(userID, msgID, msg)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

// HandleBrainEntryDelete proxies an entry deletion through the relay to the agent
// and removes it from the local cache.
func (h *Hub) HandleBrainEntryDelete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserID(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	entryID := r.URL.Query().Get("id")
	if entryID == "" {
		http.Error(w, "missing entry id", http.StatusBadRequest)
		return
	}

	// Remove from local cache immediately
	if err := h.db.DeleteBrainEntry(userID, entryID); err != nil {
		log.Printf("[brain] cache delete error: %v", err)
	}

	// Send delete to agent via relay
	msg, _ := json.Marshal(EntryDeleteMessage{
		Type:    TypeEntryDelete,
		EntryID: entryID,
	})
	msgID := "entry_delete_" + entryID + "_" + time.Now().Format("150405")
	h.routeToAgent(userID, msgID, msg)

	w.WriteHeader(http.StatusNoContent)
}

// HandleBrainEntryClassify asks the agent to run AI classification on an existing entry.
// The agent will classify and send back an entry_updated message.
func (h *Hub) HandleBrainEntryClassify(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserID(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	entryID := r.URL.Query().Get("id")
	if entryID == "" {
		http.Error(w, "missing id parameter", http.StatusBadRequest)
		return
	}

	msg, _ := json.Marshal(EntryClassifyMessage{
		Type:    TypeEntryClassify,
		EntryID: entryID,
	})
	msgID := "entry_classify_" + entryID
	h.routeToAgent(userID, msgID, msg)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "queued",
		"entry_id": entryID,
	})
}

// HandleBrainEntryCreate creates a new brain entry optimistically in the cache
// and sends an entry_create message to the agent via relay.
func (h *Hub) HandleBrainEntryCreate(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserID(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var fields map[string]any
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	title, _ := fields["title"].(string)
	if title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	// Generate entry ID server-side
	entryID := fmt.Sprintf("%d-%d", time.Now().UnixNano(), userID)
	// Use proper UUID if available via the agent; this is a temporary ID
	// that will be confirmed by entry_created from the agent
	now := time.Now().UTC().Format(time.RFC3339)

	entry := &db.BrainEntry{
		ID:        entryID,
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if v, ok := fields["category"].(string); ok && v != "" {
		entry.Category = v
	} else {
		entry.Category = "inbox"
	}
	if v, ok := fields["body"].(string); ok {
		entry.Body = v
	}
	if v, ok := fields["status"].(string); ok {
		entry.Status = v
	}
	if v, ok := fields["due_date"].(string); ok {
		entry.DueDate = v
	}
	if v, ok := fields["next_action"].(string); ok {
		entry.NextAction = v
	}
	if v, ok := fields["source"].(string); ok {
		entry.Source = v
	} else {
		entry.Source = "web"
	}
	if tags, ok := fields["tags"].([]any); ok {
		for _, t := range tags {
			if s, ok := t.(string); ok {
				entry.Tags = append(entry.Tags, s)
			}
		}
	}

	// Store optimistically in cache
	if err := h.db.UpsertBrainEntry(userID, entry); err != nil {
		log.Printf("[brain] cache create error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Send create to agent via relay
	msg, _ := json.Marshal(EntryCreateMessage{
		Type:    TypeEntryCreate,
		EntryID: entryID,
		Fields:  fields,
	})
	msgID := "entry_create_" + entryID
	h.routeToAgent(userID, msgID, msg)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

// handleEntryCreated processes an entry_created message from the agent,
// storing the new entry in the brain_entries cache.
func (h *Hub) handleEntryCreated(userID int64, data []byte) {
	var msg EntryCreatedMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("[brain] invalid entry_created: %v", err)
		return
	}

	entry := &db.BrainEntry{
		ID:         msg.Entry.ID,
		Title:      msg.Entry.Title,
		Category:   msg.Entry.Category,
		Body:       msg.Entry.Body,
		Status:     msg.Entry.Status,
		ActionDone: msg.Entry.ActionDone,
		DueDate:    msg.Entry.DueDate,
		NextAction: msg.Entry.NextAction,
		Tags:       msg.Entry.Tags,
		SubTasks:   syncSubTasksToDB(msg.Entry.SubTasks),
		Source:     msg.Entry.Source,
		CreatedAt:  msg.Entry.CreatedAt,
		UpdatedAt:  msg.Entry.UpdatedAt,
	}

	if err := h.db.UpsertBrainEntry(userID, entry); err != nil {
		log.Printf("[brain] entry_created cache error (user %d): %v", userID, err)
		return
	}

	log.Printf("[brain] cached new entry %s from agent (user %d)", entry.ID, userID)
}

// handleEntryUpdated processes an entry_updated message from the agent,
// updating the entry in the brain_entries cache.
func (h *Hub) handleEntryUpdated(userID int64, data []byte) {
	var msg EntryUpdatedMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("[brain] invalid entry_updated: %v", err)
		return
	}

	entry := &db.BrainEntry{
		ID:         msg.Entry.ID,
		Title:      msg.Entry.Title,
		Category:   msg.Entry.Category,
		Body:       msg.Entry.Body,
		Status:     msg.Entry.Status,
		ActionDone: msg.Entry.ActionDone,
		DueDate:    msg.Entry.DueDate,
		NextAction: msg.Entry.NextAction,
		Tags:       msg.Entry.Tags,
		SubTasks:   syncSubTasksToDB(msg.Entry.SubTasks),
		Source:     msg.Entry.Source,
		CreatedAt:  msg.Entry.CreatedAt,
		UpdatedAt:  msg.Entry.UpdatedAt,
	}

	if err := h.db.UpsertBrainEntry(userID, entry); err != nil {
		log.Printf("[brain] entry_updated cache error (user %d): %v", userID, err)
		return
	}

	log.Printf("[brain] updated cached entry %s from agent (user %d)", entry.ID, userID)
}

// syncSubTasksToDB converts SyncSubTask slice to db.BrainSubTask slice.
func syncSubTasksToDB(sts []SyncSubTask) []db.BrainSubTask {
	if len(sts) == 0 {
		return nil
	}
	out := make([]db.BrainSubTask, len(sts))
	for i, st := range sts {
		out[i] = db.BrainSubTask{
			ID:        st.ID,
			EntryID:   st.EntryID,
			Text:      st.Text,
			Done:      st.Done,
			SortOrder: st.SortOrder,
		}
	}
	return out
}

// HandleBrainSubTaskCreate proxies a subtask creation through the relay to the agent.
func (h *Hub) HandleBrainSubTaskCreate(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserID(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		EntryID string `json:"entry_id"`
		Text    string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.EntryID == "" || req.Text == "" {
		http.Error(w, "entry_id and text are required", http.StatusBadRequest)
		return
	}

	msg, _ := json.Marshal(SubTaskCreateMessage{
		Type:    TypeSubTaskCreate,
		EntryID: req.EntryID,
		Text:    req.Text,
	})
	msgID := fmt.Sprintf("subtask_create_%s_%s", req.EntryID, time.Now().Format("150405"))
	h.routeToAgent(userID, msgID, msg)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "queued"})
}

// HandleBrainSubTaskUpdate proxies a subtask update through the relay to the agent.
func (h *Hub) HandleBrainSubTaskUpdate(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserID(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req SubTaskUpdateMessage
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.SubTaskID == "" || req.EntryID == "" {
		http.Error(w, "subtask_id and entry_id are required", http.StatusBadRequest)
		return
	}

	req.Type = TypeSubTaskUpdate
	msg, _ := json.Marshal(req)
	msgID := fmt.Sprintf("subtask_update_%s_%s", req.SubTaskID, time.Now().Format("150405"))
	h.routeToAgent(userID, msgID, msg)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "queued"})
}

// HandleBrainSubTaskDelete proxies a subtask deletion through the relay to the agent.
func (h *Hub) HandleBrainSubTaskDelete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserID(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	subtaskID := r.URL.Query().Get("subtask_id")
	entryID := r.URL.Query().Get("entry_id")
	if subtaskID == "" || entryID == "" {
		http.Error(w, "subtask_id and entry_id are required", http.StatusBadRequest)
		return
	}

	msg, _ := json.Marshal(SubTaskDeleteMessage{
		Type:      TypeSubTaskDelete,
		SubTaskID: subtaskID,
		EntryID:   entryID,
	})
	msgID := fmt.Sprintf("subtask_delete_%s_%s", subtaskID, time.Now().Format("150405"))
	h.routeToAgent(userID, msgID, msg)

	w.WriteHeader(http.StatusNoContent)
}
