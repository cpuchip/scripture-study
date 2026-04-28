// Package ws is a tiny WebSocket broadcast hub.
package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
)

type Hub struct {
	mu      sync.Mutex
	clients map[*client]struct{}
}

type client struct {
	conn *websocket.Conn
	send chan []byte
}

func New() *Hub {
	return &Hub{clients: make(map[*client]struct{})}
}

func (h *Hub) Handle(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		log.Printf("ws accept: %v", err)
		return
	}
	cl := &client{conn: c, send: make(chan []byte, 32)}
	h.mu.Lock()
	h.clients[cl] = struct{}{}
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, cl)
		h.mu.Unlock()
		c.Close(websocket.StatusNormalClosure, "")
	}()

	ctx := r.Context()
	go func() {
		// drain incoming so the conn stays alive; we don't act on client messages.
		for {
			_, _, err := c.Read(ctx)
			if err != nil {
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-cl.send:
			if !ok {
				return
			}
			ctxW, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := c.Write(ctxW, websocket.MessageText, msg)
			cancel()
			if err != nil {
				return
			}
		}
	}
}

// Broadcast sends a JSON-encoded message of the given type to all clients.
func (h *Hub) Broadcast(msgType string, payload interface{}) {
	body := map[string]interface{}{"type": msgType, "payload": payload, "ts": time.Now().Unix()}
	b, err := json.Marshal(body)
	if err != nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		select {
		case c.send <- b:
		default:
			// drop if slow
		}
	}
}
