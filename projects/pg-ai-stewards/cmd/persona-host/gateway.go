// gateway.go — persona-host client for the ai-chattermax PLATFORM gateway.
//
// The platform (see projects/ai-chattermax) replaced the per-room socket with a
// single multiplexed /gateway speaking a typed envelope. A persona authenticates
// with a platform-minted KEY (not a display-name), subscribes to its granted
// room, and the envelope carries senderKind — so the humans-only gate is exact
// (no name-matching). Cognition is unchanged: SpawnTurn / ConsultTurn against the
// substrate's persona-turn pipeline.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Platform envelope (mirrors ai-chattermax/internal/gateway/envelope.go).
type gwOutbound struct {
	Type    string `json:"type"`
	Channel string `json:"channel,omitempty"`
	Message struct {
		ID         string `json:"id"`
		Sender     string `json:"sender"`
		SenderKind string `json:"senderKind"`
		Body       string `json:"body"`
	} `json:"message,omitempty"`
	Messages []struct {
		Sender     string `json:"sender"`
		SenderKind string `json:"senderKind"`
		Body       string `json:"body"`
	} `json:"messages,omitempty"`
	Session struct {
		Name string `json:"name"`
	} `json:"session,omitempty"`
}

// GatewayConn drives one persona on the platform gateway.
type GatewayConn struct {
	persona   Persona // the local substrate persona (character + agent_family)
	key       string  // platform-minted persona key
	roomID    string  // the granted room (channel) id
	roomLabel string
	wsBase    string
	cog       *Cognition

	conn      *websocket.Conn
	writeMu   sync.Mutex
	sessionID string
	recent    []wireMessage
	incoming  chan wireMessage
}

// NewGatewayConn builds a platform gateway connection for a persona.
func NewGatewayConn(p Persona, key, roomID, roomLabel, wsBase string, cog *Cognition) *GatewayConn {
	if roomLabel == "" {
		roomLabel = "the chat room"
	}
	return &GatewayConn{
		persona: p, key: key, roomID: roomID, roomLabel: roomLabel,
		wsBase: wsBase, cog: cog, incoming: make(chan wireMessage, 64),
	}
}

// Run dials the gateway, subscribes to the room, and drives turns until ctx ends.
func (gc *GatewayConn) Run(ctx context.Context) error {
	url := strings.TrimRight(gc.wsBase, "/") + "/gateway?key=" + gc.key
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("dial gateway (%s): %w", gc.persona.Slug, err)
	}
	gc.conn = conn
	defer conn.Close()
	log.Printf("[%s] gateway connected (room=%s)", gc.persona.Slug, gc.roomID)

	go func() { <-ctx.Done(); _ = conn.Close() }()

	// Subscribe to the granted room.
	if err := gc.sendRaw(map[string]any{"type": "subscribe", "channels": []string{gc.roomID}}); err != nil {
		return err
	}
	// Re-subscribe periodically so a grant made AFTER connect takes effect
	// without a restart — the gateway silently drops an ungranted subscribe, so
	// the first attempt is a no-op until the persona is granted to the room.
	go func() {
		t := time.NewTicker(12 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if err := gc.sendRaw(map[string]any{"type": "subscribe", "channels": []string{gc.roomID}}); err != nil {
					return
				}
			}
		}
	}()
	go gc.worker(ctx)

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("[%s] gateway read: %w", gc.persona.Slug, err)
		}
		var f gwOutbound
		if json.Unmarshal(raw, &f) != nil {
			continue
		}
		switch f.Type {
		case "history":
			for _, m := range f.Messages {
				gc.note(wireMessage{Sender: m.Sender, Body: m.Body})
			}
		case "message":
			if f.Channel != gc.roomID {
				continue
			}
			wm := wireMessage{Sender: f.Message.Sender, Body: f.Message.Body}
			gc.note(wm)
			// HUMANS-ONLY (v1): the envelope tells us the kind exactly.
			if f.Message.SenderKind == "human" && strings.TrimSpace(wm.Body) != "" {
				select {
				case gc.incoming <- wm:
				default:
					log.Printf("[%s] turn buffer full, dropping", gc.persona.Slug)
				}
			}
		}
	}
}

func (gc *GatewayConn) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case wm := <-gc.incoming:
			if err := gc.takeTurn(ctx, wm); err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("[%s] turn error: %v", gc.persona.Slug, err)
			}
		}
	}
}

func (gc *GatewayConn) takeTurn(ctx context.Context, trigger wireMessage) error {
	addressed := isAddressed(trigger.Body, gc.persona.Slug, gc.persona.DisplayName)
	var answer string
	var err error
	if gc.sessionID == "" {
		bq := buildTurnZeroFraming(gc.persona, gc.roomLabel, gc.recent, trigger, addressed)
		var sess string
		sess, answer, err = gc.cog.SpawnTurn(ctx, gc.persona.Slug+"-"+short(gc.roomID), bq)
		if err != nil {
			return err
		}
		gc.sessionID = sess
	} else {
		answer, err = gc.cog.ConsultTurn(ctx, gc.sessionID, buildConsultFraming(trigger, addressed))
		if err != nil {
			return err
		}
	}
	if IsSilence(answer) {
		return nil
	}
	if err := gc.sendRaw(map[string]any{"type": "message", "channel": gc.roomID, "body": answer}); err != nil {
		return fmt.Errorf("post reply: %w", err)
	}
	gc.note(wireMessage{Sender: gc.persona.DisplayName, Body: answer})
	return nil
}

func (gc *GatewayConn) note(wm wireMessage) {
	gc.recent = append(gc.recent, wm)
	if len(gc.recent) > recentBufferSize {
		gc.recent = gc.recent[len(gc.recent)-recentBufferSize:]
	}
}

func (gc *GatewayConn) sendRaw(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	gc.writeMu.Lock()
	defer gc.writeMu.Unlock()
	return gc.conn.WriteMessage(websocket.TextMessage, b)
}

func short(s string) string {
	if len(s) > 8 {
		return s[:8]
	}
	return s
}

// StartGatewayPersonas parses CHATTERMAX_PERSONAS ("localSlug=key@roomId,...")
// and dials each local persona into its platform room over the gateway. wsBase is
// CHATTERMAX_GATEWAY (e.g. ws://localhost:8090). Returns immediately; loops run
// in the background.
func StartGatewayPersonas(ctx context.Context, store *Store, cog *Cognition, wsBase, spec string) error {
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		slug, rest, ok := strings.Cut(part, "=")
		if !ok {
			log.Printf("gateway personas: skip malformed %q (want slug=key@roomId)", part)
			continue
		}
		key, roomID, ok := strings.Cut(rest, "@")
		slug, key, roomID = strings.TrimSpace(slug), strings.TrimSpace(key), strings.TrimSpace(roomID)
		if !ok || slug == "" || key == "" || roomID == "" {
			log.Printf("gateway personas: skip malformed %q (want slug=key@roomId)", part)
			continue
		}
		p, err := store.PersonaBySlug(ctx, slug)
		if err != nil {
			log.Printf("gateway personas: persona %q not found locally — skipping", slug)
			continue
		}
		go superviseGateway(ctx, NewGatewayConn(*p, key, roomID, "", wsBase, cog))
	}
	return nil
}

// superviseGateway runs a gateway connection with reconnect until ctx ends.
func superviseGateway(ctx context.Context, gc *GatewayConn) {
	for {
		if ctx.Err() != nil {
			return
		}
		if err := gc.Run(ctx); err != nil {
			log.Printf("[%s] gateway disconnected: %v", gc.persona.Slug, err)
		}
		if ctx.Err() != nil {
			return
		}
		gc.sessionID = "" // re-establish cognition on reconnect
		if err := sleepCtx(ctx, roomLoopRetryDelay); err != nil {
			return
		}
	}
}
