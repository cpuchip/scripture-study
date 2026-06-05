// turnloop.go — a persona's live presence in one ai-chattermax room (#7 v1).
//
// RoomConn dials the room as a WebSocket client (kind=persona so presence shows
// it as an agent, not Human), reads the attributed envelope AX3-2 broadcasts
// ({sender, body, ts}), and for each HUMAN message runs a substrate turn:
//   - turn zero spawns the persona's session (character injected here);
//   - each later turn re-asks that session.
// If the turn returns SILENCE, the persona stays quiet; otherwise it posts the
// reply as plain text and the room attributes it to the persona's display name.
//
// v1 scope (ratified): triggers #1 Reactive + #2 Addressed, HUMANS-ONLY — the
// persona ignores other personas' messages entirely (zero ping-pong risk). The
// aliveness layer (pacing, persona↔persona, cron) is v2.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// wireMessage mirrors ai-chattermax's AX3-2 envelope (cmd/server/main.go). The
// room broadcasts every message as this JSON; we decode it to attribute senders.
type wireMessage struct {
	Sender string `json:"sender"`
	Body   string `json:"body"`
	TS     string `json:"ts"`
}

const recentBufferSize = 12 // room turns kept for turn-zero context framing

// RoomConn is one persona's connection to one room.
type RoomConn struct {
	persona     Persona
	room        string
	wsURL       string
	cog         *Cognition
	isPersona   func(sender string) bool // humans-only gate (sender is a known persona)

	// Worker-owned mutable state (single goroutine — no locks needed).
	sessionID string
	recent    []wireMessage

	conn    *websocket.Conn
	writeMu sync.Mutex // gorilla forbids concurrent writes
	incoming chan wireMessage
}

// NewRoomConn builds a connection for persona p into room. wsBase is the room
// server origin (e.g. ws://localhost:8080); isPersona reports whether a sender
// name belongs to a persona (so the loop reacts to humans only).
func NewRoomConn(p Persona, room, wsBase string, cog *Cognition, isPersona func(string) bool) *RoomConn {
	// id = display name so the room attributes the persona's posts to it;
	// kind=persona so presence tags it as an agent rather than Human.
	q := url.Values{}
	q.Set("id", p.DisplayName)
	q.Set("kind", "persona")
	wsURL := strings.TrimRight(wsBase, "/") + "/ws/" + url.PathEscape(room) + "?" + q.Encode()
	return &RoomConn{
		persona:   p,
		room:      room,
		wsURL:     wsURL,
		cog:       cog,
		isPersona: isPersona,
		incoming:  make(chan wireMessage, 64),
	}
}

// Run dials the room and processes turns until ctx is cancelled or the socket
// closes. Read pump and turn worker are split so a slow turn (a blocking
// substrate dispatch) never stalls reading — messages buffer meanwhile.
func (rc *RoomConn) Run(ctx context.Context) error {
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, rc.wsURL, nil)
	if err != nil {
		return fmt.Errorf("dial %s: %w", rc.room, err)
	}
	rc.conn = conn
	defer conn.Close()
	log.Printf("[%s@%s] connected", rc.persona.Slug, rc.room)

	// Close the socket when ctx is cancelled so the read pump unblocks.
	go func() {
		<-ctx.Done()
		_ = conn.Close()
	}()

	go rc.worker(ctx)

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			if ctx.Err() != nil {
				return nil // clean shutdown
			}
			return fmt.Errorf("[%s@%s] read: %w", rc.persona.Slug, rc.room, err)
		}
		var wm wireMessage
		if jerr := json.Unmarshal(raw, &wm); jerr != nil {
			// Pre-AX3-2 raw bytes, or a non-envelope frame — ignore.
			continue
		}
		select {
		case rc.incoming <- wm:
		default:
			log.Printf("[%s@%s] turn buffer full, dropping message from %s", rc.persona.Slug, rc.room, wm.Sender)
		}
	}
}

// worker drains the incoming channel and runs one turn at a time. It owns
// rc.recent + rc.sessionID, so no locking is needed on them.
func (rc *RoomConn) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case wm := <-rc.incoming:
			rc.note(wm)
			if !rc.shouldConsider(wm) {
				continue
			}
			if err := rc.takeTurn(ctx, wm); err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("[%s@%s] turn error: %v", rc.persona.Slug, rc.room, err)
			}
		}
	}
}

// note appends a message to the bounded recent-context buffer.
func (rc *RoomConn) note(wm wireMessage) {
	rc.recent = append(rc.recent, wm)
	if len(rc.recent) > recentBufferSize {
		rc.recent = rc.recent[len(rc.recent)-recentBufferSize:]
	}
}

// shouldConsider applies the v1 trigger gate: ignore the persona's own posts and
// (humans-only) any other persona's messages; everything else is a human worth
// considering (the model then judges whether to actually speak).
func (rc *RoomConn) shouldConsider(wm wireMessage) bool {
	if wm.Sender == rc.persona.DisplayName {
		return false
	}
	if rc.isPersona != nil && rc.isPersona(wm.Sender) {
		return false
	}
	return strings.TrimSpace(wm.Body) != ""
}

// takeTurn runs the substrate turn for a triggering message and posts the reply
// unless the persona judged SILENCE.
func (rc *RoomConn) takeTurn(ctx context.Context, trigger wireMessage) error {
	addressed := isAddressed(trigger.Body, rc.persona.Slug, rc.persona.DisplayName)

	var answer string
	var err error
	if rc.sessionID == "" {
		bq := buildTurnZeroFraming(rc.persona, rc.room, rc.recent, trigger, addressed)
		var sess string
		sess, answer, err = rc.cog.SpawnTurn(ctx, rc.persona.Slug+"-"+rc.room, bq)
		if err != nil {
			return err
		}
		rc.sessionID = sess
	} else {
		answer, err = rc.cog.ConsultTurn(ctx, rc.sessionID, buildConsultFraming(trigger, addressed))
		if err != nil {
			return err
		}
	}

	if IsSilence(answer) {
		return nil
	}
	if err := rc.post(answer); err != nil {
		return fmt.Errorf("post reply: %w", err)
	}
	// Record our own turn locally so later turn-zero framing (if we ever
	// re-establish) reflects it; the session already holds it for consults.
	rc.note(wireMessage{Sender: rc.persona.DisplayName, Body: answer})
	return nil
}

// post sends a plain-text message; the room wraps it with sender = our id.
func (rc *RoomConn) post(body string) error {
	rc.writeMu.Lock()
	defer rc.writeMu.Unlock()
	return rc.conn.WriteMessage(websocket.TextMessage, []byte(body))
}

// isAddressed reports whether the body directs attention at this persona — an
// @slug / @display-name mention, or the display name in passing. It only
// strengthens the framing hint; the model still makes the final call.
func isAddressed(body, slug, displayName string) bool {
	b := strings.ToLower(body)
	if slug != "" && strings.Contains(b, "@"+strings.ToLower(slug)) {
		return true
	}
	if displayName != "" {
		dn := strings.ToLower(displayName)
		if strings.Contains(b, "@"+dn) || strings.Contains(b, "@"+strings.ReplaceAll(dn, " ", "")) {
			return true
		}
		if strings.Contains(b, dn) {
			return true
		}
	}
	return false
}

// buildTurnZeroFraming composes the binding question that establishes a persona's
// session: who it is (character), the room, the recent conversation, and the
// triggering message. This is the only place the character enters; the session
// carries it forward on every later turn.
func buildTurnZeroFraming(p Persona, room string, recent []wireMessage, trigger wireMessage, addressed bool) string {
	var b strings.Builder
	fmt.Fprintf(&b, "You are %q, a persona in the live chat room %q.\n\n", p.DisplayName, room)
	b.WriteString("YOUR CHARACTER:\n")
	if strings.TrimSpace(p.Prompt) != "" {
		b.WriteString(strings.TrimSpace(p.Prompt))
	} else {
		fmt.Fprintf(&b, "You are %s.", p.DisplayName)
	}
	b.WriteString("\n\nRECENT ROOM CONVERSATION:\n")
	b.WriteString(formatRecent(recent, trigger))
	b.WriteString("\n\nA new message just arrived:\n")
	fmt.Fprintf(&b, "%s: %s\n", trigger.Sender, trigger.Body)
	if addressed {
		b.WriteString("(You were directly addressed.)\n")
	}
	b.WriteString("\nReply in character as ")
	b.WriteString(p.DisplayName)
	b.WriteString(", in 1-3 short conversational sentences. If nothing is called for from you right now, reply with exactly the single token: SILENCE")
	return b.String()
}

// buildConsultFraming composes the short re-ask for an established session — the
// session already holds the persona's character and prior turns.
func buildConsultFraming(trigger wireMessage, addressed bool) string {
	var b strings.Builder
	b.WriteString("A new message arrived in the room:\n")
	fmt.Fprintf(&b, "%s: %s\n", trigger.Sender, trigger.Body)
	if addressed {
		b.WriteString("(You were directly addressed.)\n")
	}
	b.WriteString("\nReply in character, or reply with exactly SILENCE if nothing is called for from you.")
	return b.String()
}

// formatRecent renders the recent buffer, excluding the trigger itself (it's
// shown separately as the new message). Returns a placeholder when empty.
func formatRecent(recent []wireMessage, trigger wireMessage) string {
	var lines []string
	for _, m := range recent {
		if m.Sender == trigger.Sender && m.Body == trigger.Body {
			continue // don't duplicate the trigger
		}
		lines = append(lines, fmt.Sprintf("%s: %s", m.Sender, m.Body))
	}
	if len(lines) == 0 {
		return "(you just joined — no prior conversation)"
	}
	return strings.Join(lines, "\n")
}

// roomLoopRetryDelay is the pause before re-dialing a dropped room connection.
const roomLoopRetryDelay = 5 * time.Second
