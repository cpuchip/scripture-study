// gateway.go — persona-host client for the ai-chattermax PLATFORM gateway.
//
// MULTI-ROOM: a persona authenticates with its platform-minted KEY and is present
// in EVERY room its key is granted to. It discovers them via the platform's
// `GET /api/persona/rooms` (persona-key auth), subscribes to all, re-polls so new
// grants are picked up, and holds a SEPARATE substrate session per channel (each
// room's conversation accumulates independently). Humans-only is exact (the
// envelope carries senderKind). Cognition (SpawnTurn/ConsultTurn) is unchanged.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Platform envelope (mirrors ai-chattermax/internal/gateway/envelope.go).
type gwOutbound struct {
	Type    string `json:"type"`
	Channel string `json:"channel"`
	Message struct {
		Sender     string `json:"sender"`
		SenderKind string `json:"senderKind"`
		Body       string `json:"body"`
	} `json:"message"`
	Messages []struct {
		Sender     string `json:"sender"`
		SenderKind string `json:"senderKind"`
		Body       string `json:"body"`
	} `json:"messages"`
}

// channelState is one room's accumulating conversation, owned by the worker.
type channelState struct {
	sessionID string
	recent    []wireMessage
	label     string
}

// GatewayConn drives one persona across all the rooms its key grants.
type GatewayConn struct {
	persona Persona
	key     string
	wsBase  string
	apiBase string
	cog     *Cognition

	conn    *websocket.Conn
	writeMu sync.Mutex
	httpc   *http.Client

	// Worker-owned (single goroutine) — no locks needed.
	channels map[string]*channelState
	frames   chan gwOutbound
}

const roomRefreshInterval = 30 * time.Second

// NewGatewayConn builds a multi-room connection for a persona.
func NewGatewayConn(p Persona, key, wsBase string, cog *Cognition) *GatewayConn {
	api := strings.Replace(strings.Replace(wsBase, "wss://", "https://", 1), "ws://", "http://", 1)
	return &GatewayConn{
		persona: p, key: key, wsBase: strings.TrimRight(wsBase, "/"),
		apiBase: strings.TrimRight(api, "/"), cog: cog,
		httpc:    &http.Client{Timeout: 10 * time.Second},
		channels: map[string]*channelState{},
		frames:   make(chan gwOutbound, 128),
	}
}

// Run dials the gateway and serves all granted rooms until ctx ends.
func (gc *GatewayConn) Run(ctx context.Context) error {
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, gc.wsBase+"/gateway?key="+gc.key, nil)
	if err != nil {
		return fmt.Errorf("dial gateway (%s): %w", gc.persona.Slug, err)
	}
	gc.conn = conn
	defer conn.Close()
	log.Printf("[%s] gateway connected", gc.persona.Slug)

	go func() { <-ctx.Done(); _ = conn.Close() }()
	go gc.readPump(ctx)

	refresh := time.NewTicker(roomRefreshInterval)
	defer refresh.Stop()
	gc.refreshRooms(ctx) // subscribe to current grants immediately

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-refresh.C:
			gc.refreshRooms(ctx)
		case f, ok := <-gc.frames:
			if !ok {
				if ctx.Err() != nil {
					return nil
				}
				return fmt.Errorf("[%s] gateway read closed", gc.persona.Slug)
			}
			gc.handle(ctx, f)
		}
	}
}

// readPump reads frames into the worker channel until the socket closes.
func (gc *GatewayConn) readPump(ctx context.Context) {
	defer close(gc.frames)
	for {
		_, raw, err := gc.conn.ReadMessage()
		if err != nil {
			return
		}
		var f gwOutbound
		if json.Unmarshal(raw, &f) != nil {
			continue
		}
		select {
		case gc.frames <- f:
		case <-ctx.Done():
			return
		}
	}
}

// refreshRooms fetches the persona's granted rooms AND its DM threads, and
// subscribes to any new ones. A persona reacts in DMs exactly as in rooms
// (humans-only), so once subscribed the existing turn loop handles them.
func (gc *GatewayConn) refreshRooms(ctx context.Context) {
	rooms, err := gc.fetchRooms(ctx)
	if err != nil {
		log.Printf("[%s] fetch rooms: %v", gc.persona.Slug, err)
		return
	}
	for _, r := range rooms {
		gc.subscribeNew(r.ID, r.Name)
	}
	dms, err := gc.fetchDMs(ctx)
	if err != nil {
		log.Printf("[%s] fetch dms: %v", gc.persona.Slug, err)
		return
	}
	for _, d := range dms {
		gc.subscribeNew(d.ID, "DM:"+d.OtherName)
	}
}

// subscribeNew subscribes to a channel (room or DM) if not already joined.
func (gc *GatewayConn) subscribeNew(id, label string) {
	if gc.channels[id] != nil {
		return
	}
	gc.channels[id] = &channelState{label: label}
	if err := gc.sendRaw(map[string]any{"type": "subscribe", "channels": []string{id}}); err != nil {
		return
	}
	log.Printf("[%s] joined %s (%s)", gc.persona.Slug, label, id)
}

type personaRoom struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (gc *GatewayConn) fetchRooms(ctx context.Context) ([]personaRoom, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, gc.apiBase+"/api/persona/rooms?key="+gc.key, nil)
	if err != nil {
		return nil, err
	}
	resp, err := gc.httpc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rooms api returned %d", resp.StatusCode)
	}
	var out struct {
		Rooms []personaRoom `json:"rooms"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Rooms, nil
}

type personaDM struct {
	ID        string `json:"id"`
	OtherName string `json:"otherName"`
}

// fetchDMs returns the persona's DM threads (GET /api/persona/dms, persona-key auth).
func (gc *GatewayConn) fetchDMs(ctx context.Context) ([]personaDM, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, gc.apiBase+"/api/persona/dms?key="+gc.key, nil)
	if err != nil {
		return nil, err
	}
	resp, err := gc.httpc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dms api returned %d", resp.StatusCode)
	}
	var out struct {
		DMs []personaDM `json:"dms"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.DMs, nil
}

func (gc *GatewayConn) handle(ctx context.Context, f gwOutbound) {
	cs := gc.channels[f.Channel]
	if cs == nil {
		cs = &channelState{}
		gc.channels[f.Channel] = cs
	}
	switch f.Type {
	case "history":
		for _, m := range f.Messages {
			gc.note(cs, wireMessage{Sender: m.Sender, Body: m.Body})
		}
	case "message":
		wm := wireMessage{Sender: f.Message.Sender, Body: f.Message.Body}
		gc.note(cs, wm)
		if f.Message.SenderKind == "human" && strings.TrimSpace(wm.Body) != "" {
			if err := gc.takeTurn(ctx, f.Channel, cs, wm); err != nil && ctx.Err() == nil {
				log.Printf("[%s] turn error in %s: %v", gc.persona.Slug, f.Channel, err)
			}
		}
	}
}

func (gc *GatewayConn) takeTurn(ctx context.Context, channel string, cs *channelState, trigger wireMessage) error {
	addressed := isAddressed(trigger.Body, gc.persona.Slug, gc.persona.DisplayName)
	var answer string
	var err error
	if cs.sessionID == "" {
		label := cs.label
		if label == "" {
			label = "the chat room"
		}
		bq := buildTurnZeroFraming(gc.persona, label, cs.recent, trigger, addressed)
		var sess string
		sess, answer, err = gc.cog.SpawnTurn(ctx, gc.persona.Pipeline, gc.persona.Slug+"-"+short(channel), bq)
		if err != nil {
			return err
		}
		cs.sessionID = sess
	} else {
		answer, err = gc.cog.ConsultTurn(ctx, cs.sessionID, buildConsultFraming(trigger, addressed))
		if err != nil {
			return err
		}
	}
	if IsSilence(answer) {
		return nil
	}
	if err := gc.sendRaw(map[string]any{"type": "message", "channel": channel, "body": answer}); err != nil {
		return fmt.Errorf("post reply: %w", err)
	}
	gc.note(cs, wireMessage{Sender: gc.persona.DisplayName, Body: answer})
	return nil
}

func (gc *GatewayConn) note(cs *channelState, wm wireMessage) {
	cs.recent = append(cs.recent, wm)
	if len(cs.recent) > recentBufferSize {
		cs.recent = cs.recent[len(cs.recent)-recentBufferSize:]
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

// StartGatewayPersonas parses CHATTERMAX_PERSONAS ("localSlug=key[,...]") and
// dials each local persona into the platform; each is present in all rooms its
// key grants. wsBase is CHATTERMAX_GATEWAY (e.g. wss://chat.ibeco.me). A trailing
// "@room" (legacy single-room form) is tolerated and ignored.
func StartGatewayPersonas(ctx context.Context, store *Store, cog *Cognition, wsBase, spec string) error {
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		slug, rest, ok := strings.Cut(part, "=")
		if !ok {
			log.Printf("gateway personas: skip malformed %q (want slug=key)", part)
			continue
		}
		key, _, _ := strings.Cut(rest, "@") // tolerate + drop a legacy @room suffix
		slug, key = strings.TrimSpace(slug), strings.TrimSpace(key)
		if slug == "" || key == "" {
			log.Printf("gateway personas: skip malformed %q (want slug=key)", part)
			continue
		}
		p, err := store.PersonaBySlug(ctx, slug)
		if err != nil {
			log.Printf("gateway personas: persona %q not found locally — skipping", slug)
			continue
		}
		go superviseGateway(ctx, NewGatewayConn(*p, key, wsBase, cog))
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
		gc.channels = map[string]*channelState{} // re-establish sessions on reconnect
		if err := sleepCtx(ctx, roomLoopRetryDelay); err != nil {
			return
		}
	}
}
