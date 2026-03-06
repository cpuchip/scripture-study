package brain

import (
	"encoding/json"
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

	case TypePing:
		// Respond with pong
		pong, _ := json.Marshal(Envelope{Type: TypePong})
		sender.send <- pong

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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
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
