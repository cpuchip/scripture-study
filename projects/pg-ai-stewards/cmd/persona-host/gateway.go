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
		ID         string `json:"id"`
		Sender     string `json:"sender"`
		SenderKind string `json:"senderKind"`
		Body       string `json:"body"`
	} `json:"message"`
	Messages []struct {
		ID         string `json:"id"`
		Sender     string `json:"sender"`
		SenderKind string `json:"senderKind"`
		Body       string `json:"body"`
	} `json:"messages"`
	// On "ready" frames: who the platform says we are. The platform display
	// name is the name humans actually type in chat — addressing must match it.
	Session struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"session"`
	// On "cast" frames (DH-2): the room's cast members. We keep OUR OWN cast
	// names per channel so "@Grimble" addresses the persona voicing Grimble.
	Cast []struct {
		PersonaID   string `json:"personaId"`
		DisplayName string `json:"displayName"`
	} `json:"cast"`
}

// channelState is one room's accumulating conversation, owned by the worker.
type channelState struct {
	sessionID string
	recent    []wireMessage
	label     string
	// Async turn loop: a turn runs in its own goroutine so the select loop
	// keeps draining room_say + handling other channels. `busy` = a turn is in
	// flight for this channel (one at a time, so a persona doesn't talk over
	// itself). `pending` = the latest human message that arrived mid-turn; one
	// coalesced follow-up turn fires when the current finishes.
	busy    bool
	pending *wireMessage
	// castNames = OUR cast members in this channel (from "cast" frames).
	// Addressing one of them addresses us (DH-2).
	castNames []string
	// charSessions maps a promoted character's session → its name, for the
	// drainer (route + attribute its room_say beats) and answer attribution.
	charSessions map[string]string
	// eyedID = the message currently carrying this persona's 👀 reaction (the
	// turn's trigger). Added at turn start, removed when the turn finishes — so
	// the room sees WHICH question the persona is working, not just that it's
	// busy (typing covers that). Hops naturally on a coalesced follow-up.
	eyedID string
	// hops counts consecutive persona-triggered turns (DH-1/D1). Each
	// persona→persona turn spends one; any human message resets it. At
	// personaHopBudget the chain stops and the table waits for a human —
	// the ping-pong guard.
	hops int
}

// personaHopBudget bounds persona→persona turn chains per channel (DM narrates
// → PC reacts → DM resolves = 3 hops, then it's the humans' table again).
const personaHopBudget = 3

// turnResult is what a turn goroutine reports back to the worker loop (which is
// the sole owner of gc.channels). Carrying `gen` lets the loop discard results
// from a connection that has since reconnected (the channels map was reset).
type turnResult struct {
	gen       uint64
	channel   string
	kind      turnResultKind
	sessionID string // for kindSession (early) and kindDone (the turn's session)
	answer    string // for kindDone
	err       error  // for kindDone
	// Promoted-character turns (DH-2): `as` attributes the answer to the
	// character; charID persists the character's session on first spawn.
	as     string
	charID string
}

type turnResultKind int

const (
	kindSession turnResultKind = iota // the session id became known (mid-turn)
	kindDone                          // the turn finished (answer or err)
)

// cognition is the slice of *Cognition the gateway uses — an interface so the
// async turn loop can be unit-tested with a fake (no substrate, no socket).
type cognition interface {
	SpawnTurn(ctx context.Context, pipeline, slug, bindingQuestion string, onSession func(string)) (sessionID, answer string, err error)
	ConsultTurn(ctx context.Context, sessionID, question string) (answer string, err error)
	ClaimOutboxForSessions(ctx context.Context, sessionIDs []string) ([]OutboxRow, error)
	EnsureCharacter(ctx context.Context, personaSlug, name string, promote bool) (Character, error)
	SaveCharacterSession(ctx context.Context, id, sessionID string) error
}

// GatewayConn drives one persona across all the rooms its key grants.
type GatewayConn struct {
	persona Persona
	key     string
	wsBase  string
	apiBase string
	cog     cognition

	conn    *websocket.Conn
	writeMu sync.Mutex
	httpc   *http.Client

	// emitFn, when set, intercepts room posts (tests). Default = the real
	// websocket send (gc.emit).
	emitFn func(channel, body string) error
	// rawFn, when set, intercepts ALL raw frames — typing pulses, reactions
	// (tests). Default = the real websocket send.
	rawFn func(v any) error

	// Worker-owned (single goroutine) — no locks needed.
	channels    map[string]*channelState
	frames      chan gwOutbound
	generation  uint64          // bumped per connection; guards stale turn results
	turnResults chan turnResult // turn goroutines → the worker loop
	selfName    string          // platform display name (from the ready frame)
	selfID      string          // platform persona id (from the ready frame)
	// respondPolicy (REM-3): all | mentioned | judgment — refreshed on the rooms
	// poll. "mentioned" skips the turn entirely (no dispatch cost) unless the
	// message names the persona; "judgment" licenses unaddressed chiming-in.
	respondPolicy string
}

// emit posts a message to a channel — overridable in tests via emitFn.
func (gc *GatewayConn) emit(channel, body string) error {
	if gc.emitFn != nil {
		return gc.emitFn(channel, body)
	}
	return gc.sendRaw(map[string]any{"type": "message", "channel": channel, "body": body})
}

// emitAs posts a message spoken by a named cast member (DH-2): the platform
// attributes the line to subPersona, auto-creating the character on first use.
func (gc *GatewayConn) emitAs(channel, subPersona, body string) error {
	if subPersona == "" {
		return gc.emit(channel, body)
	}
	if gc.emitFn != nil {
		return gc.emitFn(channel, "["+subPersona+"] "+body) // test seam: visible attribution
	}
	return gc.sendRaw(map[string]any{"type": "message", "channel": channel, "body": body, "subPersona": subPersona})
}

const roomRefreshInterval = 30 * time.Second

// room_say beats should feel near-real-time ("hang on…" before a slow tool),
// so drain often. The query is a cheap partial-index scan over unposted rows.
const roomSayDrainInterval = 1 * time.Second

// frameBufferSize bounds the inbound gateway-frame channel. Recreated per
// connection in Run (see the reconnect-panic fix).
const frameBufferSize = 128

// roomTypingInterval refreshes the "Codewright is typing…" indicator while a
// turn is in flight. Typing indicators auto-expire client-side after a few
// seconds, so we re-send periodically; this is only possible because the async
// turn loop keeps the worker loop free during a turn.
const roomTypingInterval = 3 * time.Second

// turnResultsBuffer bounds the turn-goroutine → loop channel. In-flight turns
// are capped by the number of channels (one turn per channel at a time), so
// this is generous; it also absorbs results from a just-reconnected old
// generation without blocking those goroutines.
const turnResultsBuffer = 64

// NewGatewayConn builds a multi-room connection for a persona.
func NewGatewayConn(p Persona, key, wsBase string, cog *Cognition) *GatewayConn {
	api := strings.Replace(strings.Replace(wsBase, "wss://", "https://", 1), "ws://", "http://", 1)
	return &GatewayConn{
		persona: p, key: key, wsBase: strings.TrimRight(wsBase, "/"),
		apiBase: strings.TrimRight(api, "/"), cog: cog,
		httpc:       &http.Client{Timeout: 10 * time.Second},
		channels:    map[string]*channelState{},
		frames:      make(chan gwOutbound, frameBufferSize),
		turnResults: make(chan turnResult, turnResultsBuffer),
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

	// Fresh per-connection frame channel. readPump closes gc.frames on
	// disconnect, so a reconnect MUST NOT reuse the closed one — doing so
	// double-closes it (panic), which crash-loops the whole host on every
	// chat.ibeco.me redeploy (the connection drops → reconnect → panic).
	// `defer close(gc.frames)` captures this channel value, so an old readPump
	// closes the old channel while the new connection uses this fresh one.
	gc.frames = make(chan gwOutbound, frameBufferSize)

	// New connection generation. Turn goroutines started under this generation
	// tag their results with it; results from an older generation (a turn still
	// finishing after a reconnect reset gc.channels) are discarded by the loop.
	gc.generation++

	go func() { <-ctx.Done(); _ = conn.Close() }()
	go gc.readPump(ctx)

	refresh := time.NewTicker(roomRefreshInterval)
	defer refresh.Stop()
	gc.refreshRooms(ctx) // subscribe to current grants immediately

	// room_say delivery: drain mid-turn messages this persona emitted and post
	// them to the right channel. Runs in THIS worker goroutine so it can read
	// gc.channels lock-free, same as refresh.
	drain := time.NewTicker(roomSayDrainInterval)
	defer drain.Stop()

	// "X is typing…" while a turn runs — refreshed because it auto-expires
	// client-side. Possible only now that the loop is free during a turn.
	typing := time.NewTicker(roomTypingInterval)
	defer typing.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-refresh.C:
			gc.refreshRooms(ctx)
		case <-drain.C:
			gc.drainOutbox(ctx)
		case <-typing.C:
			gc.pulseTyping()
		case tr := <-gc.turnResults:
			gc.applyTurnResult(ctx, tr)
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
		Persona struct {
			RespondPolicy string `json:"respondPolicy"`
		} `json:"persona"`
		Rooms []personaRoom `json:"rooms"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	// respond_policy rides the rooms poll, so a Settings change applies within
	// one refresh interval — no host restart.
	if p := out.Persona.RespondPolicy; p != "" && p != gc.respondPolicy {
		gc.respondPolicy = p
		log.Printf("[%s] respond_policy: %s", gc.persona.Slug, p)
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
	if f.Type == "ready" {
		if f.Session.Name != "" {
			gc.selfName = f.Session.Name
		}
		if f.Session.ID != "" {
			gc.selfID = f.Session.ID
		}
		return
	}
	cs := gc.channels[f.Channel]
	if cs == nil {
		cs = &channelState{}
		gc.channels[f.Channel] = cs
	}
	switch f.Type {
	case "cast":
		// Track OUR cast members in this channel: addressing one of them
		// (DH-2: "@Grimble, how much?") addresses us.
		cs.castNames = cs.castNames[:0]
		for _, m := range f.Cast {
			if m.PersonaID == gc.selfID {
				cs.castNames = append(cs.castNames, m.DisplayName)
			}
		}
	case "history":
		for _, m := range f.Messages {
			gc.note(cs, wireMessage{Sender: m.Sender, Body: m.Body})
		}
	case "message":
		wm := wireMessage{ID: f.Message.ID, Sender: f.Message.Sender, Body: f.Message.Body}
		gc.note(cs, wm)
		switch {
		case f.Message.SenderKind == "human":
			cs.hops = 0 // a human at the table resets the persona-chain budget
			if strings.TrimSpace(wm.Body) != "" {
				gc.maybeStartTurn(ctx, f.Channel, cs, wm)
			}
		case f.Message.SenderKind == "persona" && wm.Sender != gc.selfName && wm.Sender != gc.persona.DisplayName && !isOwnCast(cs, wm.Sender):
			// Persona→persona (DH-1/D1): another persona's message starts a
			// turn ONLY when it names us, and only within the hop budget —
			// that's what lets the DM hand off to a PC ("@party-bard, your
			// move") without two models ping-ponging forever. Our own cast
			// members' lines never trigger us (we spoke them).
			if !isAddressed(wm.Body, gc.addressNames(cs)...) {
				return
			}
			if cs.hops >= personaHopBudget {
				log.Printf("[%s] hop budget reached in %s — waiting for a human", gc.persona.Slug, f.Channel)
				return
			}
			cs.hops++
			gc.maybeStartTurn(ctx, f.Channel, cs, wm)
		}
	}
}

// maybeStartTurn kicks off a turn in its own goroutine so the worker loop keeps
// running (draining room_say, serving other channels) while the model works.
// One turn at a time per channel: if a turn is already running, the trigger is
// held as `pending` and a single coalesced follow-up fires when it finishes.
// Runs in the worker goroutine, so it owns cs and prepares all turn inputs as
// plain values — the goroutine never touches gc.channels.
// addressNames is every name that counts as "us" for addressing: slug, host
// display name, platform display name, and our cast members in this channel.
// Cast members also answer to their FIRST name — people say "Grimble, how
// much?", not "Grimble the shopkeep, how much?".
func (gc *GatewayConn) addressNames(cs *channelState) []string {
	names := []string{gc.persona.Slug, gc.persona.DisplayName, gc.selfName}
	for _, cn := range cs.castNames {
		names = append(names, cn)
		if first := castFirstName(cn); first != "" && !strings.EqualFold(first, cn) {
			names = append(names, first)
		}
	}
	return names
}

// castFirstName extracts a usable first name from a cast display name —
// "Vex, guard captain" → "Vex". Short or article-like first words ("The
// Magistrate") return "" rather than matching half the dictionary.
func castFirstName(name string) string {
	fields := strings.Fields(name)
	if len(fields) == 0 {
		return ""
	}
	first := strings.TrimFunc(fields[0], func(r rune) bool {
		return !('a' <= r && r <= 'z' || 'A' <= r && r <= 'Z' || '0' <= r && r <= '9')
	})
	switch strings.ToLower(first) {
	case "", "the", "a", "an", "sir", "old", "mr", "mrs", "ms", "dr", "lady", "lord", "captain":
		return ""
	}
	if len(first) < 3 {
		return ""
	}
	return first
}

// isOwnCast reports whether a sender name is one of OUR cast members here.
func isOwnCast(cs *channelState, sender string) bool {
	for _, n := range cs.castNames {
		if strings.EqualFold(n, sender) {
			return true
		}
	}
	return false
}

// matchCast returns the canonical cast-member name a body addresses ("" if
// none) — full name or first name, same variants addressing uses.
func matchCast(cs *channelState, body string) string {
	for _, cn := range cs.castNames {
		names := []string{cn}
		if first := castFirstName(cn); first != "" {
			names = append(names, first)
		}
		if isAddressed(body, names...) {
			return cn
		}
	}
	return ""
}

func (gc *GatewayConn) maybeStartTurn(ctx context.Context, channel string, cs *channelState, trigger wireMessage) {
	addressed := isAddressed(trigger.Body, gc.addressNames(cs)...)
	// respond_policy "mentioned": unaddressed messages cost nothing — no turn,
	// no typing, no eyes. They're already note()'d into recent, so the persona
	// still sees them as context when it IS addressed.
	if gc.respondPolicy == "mentioned" && !addressed {
		return
	}
	if cs.busy {
		t := trigger
		cs.pending = &t // coalesce: keep the latest; intervening msgs are in recent
		return
	}
	cs.busy = true
	// Show "typing…" immediately, not on the next 3s tick.
	_ = gc.sendRaw(map[string]any{"type": "typing", "channel": channel})
	// 👀 on the message we're working — message-scoped, where typing is
	// channel-scoped. Removed in applyTurnResult when the turn finishes.
	if trigger.ID != "" {
		_ = gc.sendRaw(map[string]any{"type": "reaction", "channel": channel, "messageId": trigger.ID, "emoji": "👀", "op": "add"})
		cs.eyedID = trigger.ID
	}

	gen := gc.generation
	pipeline := gc.persona.Pipeline

	// Promotion routing (DH-2): a message addressed to one of our PROMOTED
	// characters runs on THAT character's own session — its own mind — and the
	// answer posts under the character's name. Facet characters fall through
	// to the owner's turn (one mind, many voices).
	if castName := matchCast(cs, trigger.Body); castName != "" {
		ch, err := gc.cog.EnsureCharacter(ctx, gc.persona.Slug, castName, gc.persona.DefaultPromote)
		if err != nil {
			log.Printf("[%s] ensure character %q: %v", gc.persona.Slug, castName, err)
		} else if ch.Promoted {
			label := cs.label
			if label == "" {
				label = "the chat room"
			}
			var bq string
			if ch.SessionID == "" {
				bq = buildCharacterFraming(ch, gc.persona.DisplayName, label, cs.recent, trigger)
			} else {
				bq = buildConsultFraming(trigger, true)
			}
			slug := gc.persona.Slug + "-" + short(channel) + "-" + short(strings.ToLower(strings.ReplaceAll(ch.Name, " ", "")))
			go gc.runTurn(ctx, gen, channel, ch.SessionID, pipeline, slug, bq, ch.Name, ch.ID)
			return
		}
	}

	sessionID := cs.sessionID
	slug := gc.persona.Slug + "-" + short(channel)

	var bq string
	if sessionID == "" {
		label := cs.label
		if label == "" {
			label = "the chat room"
		}
		bq = buildTurnZeroFraming(gc.persona, label, cs.recent, trigger, addressed, gc.selfName)
	} else {
		bq = buildConsultFraming(trigger, addressed)
	}
	if gc.respondPolicy == "judgment" && !addressed {
		bq += "\n(You weren't addressed directly — feel free to chime in if you have something genuinely worth adding; otherwise SILENCE.)"
	}

	go gc.runTurn(ctx, gen, channel, sessionID, pipeline, slug, bq, "", "")
}

// runTurn does ONLY the cognition (off the worker loop) and reports back over
// gc.turnResults. It never touches gc.channels. For turn-zero it reports the
// session id early (via SpawnTurn's onSession callback) so the loop can route
// room_say beats while the model is still working. asName/charID are set for
// promoted-character turns (DH-2): the answer is attributed to the character
// and the spawned session is persisted as the character's own mind.
func (gc *GatewayConn) runTurn(ctx context.Context, gen uint64, channel, sessionID, pipeline, slug, framing, asName, charID string) {
	send := func(tr turnResult) {
		select {
		case gc.turnResults <- tr:
		case <-ctx.Done():
		}
	}
	var answer string
	var err error
	if sessionID == "" {
		sessionID, answer, err = gc.cog.SpawnTurn(ctx, pipeline, slug, framing, func(sid string) {
			send(turnResult{gen: gen, channel: channel, kind: kindSession, sessionID: sid, as: asName, charID: charID})
		})
	} else {
		answer, err = gc.cog.ConsultTurn(ctx, sessionID, framing)
	}
	// Truncated-stream retry (2026-06-10: a Fireworks stream died mid-turn —
	// reasoning arrived, content never did, and the turn "completed" empty).
	// An EMPTY answer with no error is never legitimate (silence is the
	// explicit token SILENCE), so re-ask once on the now-known session.
	if err == nil && strings.TrimSpace(answer) == "" && sessionID != "" {
		log.Printf("[%s] empty answer in %s (truncated stream?) — retrying once", gc.persona.Slug, channel)
		answer, err = gc.cog.ConsultTurn(ctx, sessionID, "(Your previous reply was lost in transmission and never reached the room — please send it again, briefly. If you truly have nothing to say, reply with exactly SILENCE.)")
	}
	send(turnResult{gen: gen, channel: channel, kind: kindDone, sessionID: sessionID, answer: answer, err: err, as: asName, charID: charID})
}

// applyTurnResult runs in the worker loop (sole owner of gc.channels) and
// applies a turn goroutine's report: set the session early, or post the answer
// (after flushing any room_say beats so they precede it), free the channel, and
// fire a coalesced follow-up if a message arrived mid-turn. Results from an old
// connection generation are discarded.
func (gc *GatewayConn) applyTurnResult(ctx context.Context, tr turnResult) {
	if tr.gen != gc.generation {
		return // stale: the connection reconnected and reset its channels
	}
	cs := gc.channels[tr.channel]
	if cs == nil {
		return
	}

	switch tr.kind {
	case kindSession:
		if tr.charID != "" {
			// A promoted character's own session (DH-2): register it for the
			// drainer (its beats post as the character) and persist it — the
			// character's mind is durable and room-agnostic.
			if cs.charSessions == nil {
				cs.charSessions = map[string]string{}
			}
			cs.charSessions[tr.sessionID] = tr.as
			if err := gc.cog.SaveCharacterSession(ctx, tr.charID, tr.sessionID); err != nil {
				log.Printf("[%s] save character session (%s): %v", gc.persona.Slug, tr.as, err)
			}
			return
		}
		if cs.sessionID == "" {
			cs.sessionID = tr.sessionID // early — the drainer can now route beats
		}
	case kindDone:
		if tr.charID != "" && tr.sessionID != "" {
			if cs.charSessions == nil {
				cs.charSessions = map[string]string{}
			}
			cs.charSessions[tr.sessionID] = tr.as
		} else if tr.sessionID != "" {
			cs.sessionID = tr.sessionID
		}
		if tr.err != nil {
			if ctx.Err() == nil {
				log.Printf("[%s] turn error in %s: %v", gc.persona.Slug, tr.channel, tr.err)
			}
		} else if !IsSilence(tr.answer) {
			// Flush any pending room_say beats FIRST so "🔍 …" precedes the answer.
			gc.drainOutbox(ctx)
			if err := gc.emitAs(tr.channel, tr.as, tr.answer); err != nil {
				if ctx.Err() == nil {
					log.Printf("[%s] post reply in %s: %v", gc.persona.Slug, tr.channel, err)
				}
			} else {
				sender := gc.persona.DisplayName
				if tr.as != "" {
					sender = tr.as
				}
				gc.note(cs, wireMessage{Sender: sender, Body: tr.answer})
			}
		}
		// Take the 👀 off the trigger message — the turn is over whether it
		// answered, stayed silent, or errored.
		if cs.eyedID != "" {
			_ = gc.sendRaw(map[string]any{"type": "reaction", "channel": tr.channel, "messageId": cs.eyedID, "emoji": "👀", "op": "remove"})
			cs.eyedID = ""
		}
		// Channel free; fire one coalesced follow-up if a message arrived mid-turn.
		cs.busy = false
		if cs.pending != nil {
			next := *cs.pending
			cs.pending = nil
			gc.maybeStartTurn(ctx, tr.channel, cs, next)
		}
	}
}

// pulseTyping sends a "typing" frame for every channel with a turn in flight,
// so the room shows "<persona> is typing…" between the human's message and the
// reply. The gateway stamps the persona's name + broadcasts to others; it
// auto-expires client-side, hence the periodic refresh. Worker-goroutine only.
func (gc *GatewayConn) pulseTyping() {
	for chID, cs := range gc.channels {
		if cs.busy {
			_ = gc.sendRaw(map[string]any{"type": "typing", "channel": chID})
		}
	}
}

// drainOutbox posts this persona's pending room_say messages. It maps each
// claimed row's session back to the channel currently holding that session and
// posts there (with the optional mood emoji prefixed). Runs in the worker
// goroutine, so reading gc.channels is lock-free.
func (gc *GatewayConn) drainOutbox(ctx context.Context) {
	// session_id → channel for this connection's live channels — the persona's
	// room sessions AND its promoted characters' sessions (their beats post
	// attributed to the character).
	sessToChan := make(map[string]string, len(gc.channels))
	sessChar := make(map[string]string)
	sessions := make([]string, 0, len(gc.channels))
	for chID, cs := range gc.channels {
		if cs.sessionID != "" {
			sessToChan[cs.sessionID] = chID
			sessions = append(sessions, cs.sessionID)
		}
		for sid, name := range cs.charSessions {
			sessToChan[sid] = chID
			sessChar[sid] = name
			sessions = append(sessions, sid)
		}
	}
	if len(sessions) == 0 {
		return
	}
	msgs, err := gc.cog.ClaimOutboxForSessions(ctx, sessions)
	if err != nil {
		log.Printf("[%s] drain outbox: %v", gc.persona.Slug, err)
		return
	}
	for _, m := range msgs {
		chID := sessToChan[m.SessionID]
		if chID == "" {
			continue // session no longer mapped to a live channel
		}
		body := m.Body
		// Prefix the mood emoji unless the model already led with it ("🔍 🔍 …").
		if m.Mood != "" && !strings.HasPrefix(strings.TrimSpace(body), m.Mood) {
			body = m.Mood + " " + body
		}
		// Attribution: an explicit as_character wins; otherwise a beat from a
		// promoted character's own session speaks as that character.
		as := m.SubPersona
		if as == "" {
			as = sessChar[m.SessionID]
		}
		if err := gc.emitAs(chID, as, body); err != nil {
			log.Printf("[%s] post room_say: %v", gc.persona.Slug, err)
			continue
		}
		// Record our own mid-turn post so the persona doesn't re-react to it.
		sender := gc.persona.DisplayName
		if as != "" {
			sender = as
		}
		if cs := gc.channels[chID]; cs != nil {
			gc.note(cs, wireMessage{Sender: sender, Body: body})
		}
	}
}

func (gc *GatewayConn) note(cs *channelState, wm wireMessage) {
	cs.recent = append(cs.recent, wm)
	if len(cs.recent) > recentBufferSize {
		cs.recent = cs.recent[len(cs.recent)-recentBufferSize:]
	}
}

func (gc *GatewayConn) sendRaw(v any) error {
	if gc.rawFn != nil {
		return gc.rawFn(v) // test seam — capture frames without a socket
	}
	if gc.conn == nil {
		return nil // no live connection (or a test) — nothing to send
	}
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
