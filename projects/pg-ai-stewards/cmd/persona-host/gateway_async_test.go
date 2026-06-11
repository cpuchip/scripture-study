package main

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeCog stands in for *Cognition. SpawnTurn blocks for `dur`, reporting the
// session after `sessionAfter` (so we can prove the session is known mid-turn).
type fakeCog struct {
	dur          time.Duration
	sessionAfter time.Duration
	session      string
	answer       string

	mu       sync.Mutex
	outbox   []OutboxRow // returned once, for the session, on the next claim
	claimed  bool
	spawns   int
	consults int

	// Promotion (DH-2): characters EnsureCharacter hands back.
	promote       bool
	chars         map[string]Character
	savedSessions map[string]string // character id → saved session
}

func (f *fakeCog) EnsureCharacter(_ context.Context, slug, name string, promote bool) (Character, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.chars == nil {
		f.chars = map[string]Character{}
	}
	key := strings.ToLower(name)
	ch, ok := f.chars[key]
	if !ok {
		ch = Character{ID: "c-" + key, PersonaSlug: slug, Name: name, Promoted: promote || f.promote}
		f.chars[key] = ch
	}
	return ch, nil
}

func (f *fakeCog) SaveCharacterSession(_ context.Context, id, sessionID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.savedSessions == nil {
		f.savedSessions = map[string]string{}
	}
	f.savedSessions[id] = sessionID
	return nil
}

func (f *fakeCog) SpawnTurn(ctx context.Context, pipeline, slug, bq string, onSession func(string)) (string, string, error) {
	f.mu.Lock()
	f.spawns++
	f.mu.Unlock()
	if onSession != nil && f.sessionAfter >= 0 {
		go func() {
			time.Sleep(f.sessionAfter)
			onSession(f.session)
		}()
	}
	select {
	case <-time.After(f.dur):
	case <-ctx.Done():
		return "", "", ctx.Err()
	}
	return f.session, f.answer, nil
}

func (f *fakeCog) ConsultTurn(ctx context.Context, sessionID, q string) (string, error) {
	f.mu.Lock()
	f.consults++
	f.mu.Unlock()
	select {
	case <-time.After(f.dur):
	case <-ctx.Done():
		return "", ctx.Err()
	}
	return f.answer, nil
}

func (f *fakeCog) ClaimOutboxForSessions(ctx context.Context, sessions []string) ([]OutboxRow, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.claimed || len(f.outbox) == 0 {
		return nil, nil
	}
	f.claimed = true
	return f.outbox, nil
}

func newTestConn(cog cognition) (*GatewayConn, *[]string, *sync.Mutex) {
	var mu sync.Mutex
	var posts []string
	gc := &GatewayConn{
		persona:     Persona{Slug: "tester", DisplayName: "Tester", Pipeline: "persona-turn-code"},
		cog:         cog,
		channels:    map[string]*channelState{},
		turnResults: make(chan turnResult, turnResultsBuffer),
		generation:  1,
	}
	gc.emitFn = func(channel, body string) error {
		mu.Lock()
		posts = append(posts, body)
		mu.Unlock()
		return nil
	}
	return gc, &posts, &mu
}

// drain applyTurnResult for every result currently queued (mimics the Run loop
// pumping the turnResults case), with a short wait for async sends to land.
func pump(gc *GatewayConn, ctx context.Context, d time.Duration) {
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		select {
		case tr := <-gc.turnResults:
			gc.applyTurnResult(ctx, tr)
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}

// The core fix: a turn runs OFF the loop (maybeStartTurn returns immediately
// though cognition blocks), the session is known mid-turn (early), room_say
// beats post BEFORE the answer, and the channel frees for a follow-up.
func TestAsyncTurn_NonBlocking_EarlySession_Ordering(t *testing.T) {
	ctx := context.Background()
	cog := &fakeCog{
		dur:          300 * time.Millisecond,
		sessionAfter: 30 * time.Millisecond,
		session:      "wi--test--turn",
		answer:       "here is the answer",
		outbox:       []OutboxRow{{ID: 1, SessionID: "wi--test--turn", Body: "looking that up", Mood: "🔍"}},
	}
	gc, posts, pmu := newTestConn(cog)
	cs := &channelState{label: "Engineering"}
	gc.channels["eng"] = cs

	// maybeStartTurn must return immediately even though SpawnTurn blocks 300ms.
	t0 := time.Now()
	gc.maybeStartTurn(ctx, "eng", cs, wireMessage{Sender: "human", Body: "how does X work?"})
	if elapsed := time.Since(t0); elapsed > 50*time.Millisecond {
		t.Fatalf("maybeStartTurn blocked %v — turn is not async", elapsed)
	}
	if !cs.busy {
		t.Fatal("channel should be busy while the turn runs")
	}

	// Pump results for longer than the turn. The early session lands first,
	// then the done result drains the beat and posts the answer.
	pump(gc, ctx, 600*time.Millisecond)

	if cs.sessionID != "wi--test--turn" {
		t.Fatalf("session not set (got %q)", cs.sessionID)
	}
	pmu.Lock()
	defer pmu.Unlock()
	if len(*posts) != 2 {
		t.Fatalf("want 2 posts (beat + answer), got %d: %v", len(*posts), *posts)
	}
	if (*posts)[0] != "🔍 looking that up" {
		t.Fatalf("beat should post FIRST, got %q", (*posts)[0])
	}
	if (*posts)[1] != "here is the answer" {
		t.Fatalf("answer should post second, got %q", (*posts)[1])
	}
	if cs.busy {
		t.Fatal("channel should be free after the turn")
	}
}

// A turn result from an older connection generation is discarded (the reconnect
// guard) — no post, no panic.
func TestAsyncTurn_StaleGenerationDiscarded(t *testing.T) {
	ctx := context.Background()
	gc, posts, pmu := newTestConn(&fakeCog{})
	cs := &channelState{busy: true}
	gc.channels["eng"] = cs

	gc.applyTurnResult(ctx, turnResult{gen: 0, channel: "eng", kind: kindDone, answer: "stale answer"})

	pmu.Lock()
	defer pmu.Unlock()
	if len(*posts) != 0 {
		t.Fatalf("stale-generation result must not post, got %v", *posts)
	}
}

// Promotion (DH-2): addressing a promoted character runs the turn on the
// CHARACTER's own session — the spawn is persisted as the character's mind,
// the answer posts attributed to the character, and the character's room_say
// beats attribute to it too.
func TestPromotedCharacterTurn(t *testing.T) {
	ctx := context.Background()
	cog := &fakeCog{
		dur: 30 * time.Millisecond, sessionAfter: 5 * time.Millisecond,
		session: "wi--thorin--turn", answer: "I raise my axe.",
		promote: true,
		outbox:  []OutboxRow{{ID: 1, SessionID: "wi--thorin--turn", Body: "tightens his grip"}},
	}
	gc, posts, pmu := newTestConn(cog)
	gc.persona.DefaultPromote = true
	gc.selfID = "p-party"
	cs := &channelState{sessionID: "wi--party-room--turn", label: "Holodeck-3"}
	gc.channels["dnd"] = cs

	var cf gwOutbound
	cf.Type = "cast"
	cf.Channel = "dnd"
	cf.Cast = []struct {
		PersonaID   string `json:"personaId"`
		DisplayName string `json:"displayName"`
	}{{"p-party", "Thorin Oakenshield"}}
	gc.handle(ctx, cf)

	var mf gwOutbound
	mf.Type = "message"
	mf.Channel = "dnd"
	mf.Message.ID = "m1"
	mf.Message.Sender = "michael"
	mf.Message.SenderKind = "human"
	mf.Message.Body = "Thorin, the goblin lunges at you — what do you do?"
	gc.handle(ctx, mf)
	if !cs.busy {
		t.Fatal("promoted-character trigger should start a turn")
	}
	pump(gc, ctx, 300*time.Millisecond)

	// The character's session was persisted and registered for the drainer.
	cog.mu.Lock()
	saved := cog.savedSessions["c-thorin oakenshield"]
	spawns := cog.spawns
	cog.mu.Unlock()
	if saved != "wi--thorin--turn" {
		t.Fatalf("character session not saved: %q", saved)
	}
	if spawns != 1 {
		t.Fatalf("want a SPAWN on the character's own session, got %d spawns", spawns)
	}
	if cs.charSessions["wi--thorin--turn"] != "Thorin Oakenshield" {
		t.Fatalf("charSessions = %v", cs.charSessions)
	}
	// The owner's room session is untouched.
	if cs.sessionID != "wi--party-room--turn" {
		t.Fatalf("owner session clobbered: %q", cs.sessionID)
	}
	pmu.Lock()
	defer pmu.Unlock()
	if len(*posts) != 2 {
		t.Fatalf("want beat + answer, got %v", *posts)
	}
	if (*posts)[0] != "[Thorin Oakenshield] tightens his grip" {
		t.Fatalf("character beat attribution wrong: %q", (*posts)[0])
	}
	if (*posts)[1] != "[Thorin Oakenshield] I raise my axe." {
		t.Fatalf("character answer attribution wrong: %q", (*posts)[1])
	}
}

// Cast addressing (DH-2): naming one of OUR cast members addresses us; our own
// cast members' lines never trigger us.
func TestCastAddressing(t *testing.T) {
	ctx := context.Background()
	cog := &fakeCog{dur: 10 * time.Millisecond, sessionAfter: -1, session: "wi--ca--turn", answer: "three coppers"}
	gc, _, _ := newTestConn(cog)
	gc.respondPolicy = "mentioned"
	gc.selfID = "p-dm"
	cs := &channelState{sessionID: "wi--ca--turn"}
	gc.channels["dnd"] = cs

	// Cast frame: Grimble is ours, Lady Vex belongs to another persona.
	var cf gwOutbound
	cf.Type = "cast"
	cf.Channel = "dnd"
	cf.Cast = []struct {
		PersonaID   string `json:"personaId"`
		DisplayName string `json:"displayName"`
	}{{"p-dm", "Grimble the shopkeep"}, {"p-other", "Lady Vex"}}
	gc.handle(ctx, cf)
	if len(cs.castNames) != 1 || cs.castNames[0] != "Grimble the shopkeep" {
		t.Fatalf("castNames = %v", cs.castNames)
	}

	mkMsg := func(kind, sender, body string) gwOutbound {
		var f gwOutbound
		f.Type = "message"
		f.Channel = "dnd"
		f.Message.ID = "m"
		f.Message.Sender = sender
		f.Message.SenderKind = kind
		f.Message.Body = body
		return f
	}
	// Under policy=mentioned, "Grimble ..." wakes us (our cast member's name).
	gc.handle(ctx, mkMsg("human", "michael", "Grimble I need some pickled herring! how much?"))
	if !cs.busy {
		t.Fatal("naming our cast member must address us")
	}
	pump(gc, ctx, 100*time.Millisecond)
	// "Lady Vex ..." does not (someone else's character).
	gc.handle(ctx, mkMsg("human", "michael", "Lady Vex, arrest him!"))
	if cs.busy {
		t.Fatal("someone else's cast member must not address us")
	}
	// Our own cast member's line never triggers us.
	gc.handle(ctx, mkMsg("persona", "Grimble the shopkeep", "tester, want herring?"))
	if cs.busy {
		t.Fatal("our own cast lines must not trigger us")
	}
}

// Cast (DH-2): an outbox row with a SubPersona posts attributed to the cast
// member (the emitFn seam renders it as a [Name] prefix), and the persona's
// recent-notes record the cast name as the sender.
func TestDrainOutbox_CastAttribution(t *testing.T) {
	ctx := context.Background()
	cog := &fakeCog{outbox: []OutboxRow{
		{ID: 1, SessionID: "wi--cast--turn", Body: "Best prices in the realm!", Mood: "🎲", SubPersona: "Grimble the shopkeep"},
		{ID: 2, SessionID: "wi--cast--turn", Body: "plain narration"},
	}}
	gc, posts, pmu := newTestConn(cog)
	gc.channels["dnd"] = &channelState{sessionID: "wi--cast--turn"}

	gc.drainOutbox(ctx)

	pmu.Lock()
	defer pmu.Unlock()
	if len(*posts) != 2 {
		t.Fatalf("want 2 posts, got %v", *posts)
	}
	if (*posts)[0] != "[Grimble the shopkeep] 🎲 Best prices in the realm!" {
		t.Fatalf("cast attribution wrong: %q", (*posts)[0])
	}
	if (*posts)[1] != "plain narration" {
		t.Fatalf("plain beat should stay unattributed: %q", (*posts)[1])
	}
	cs := gc.channels["dnd"]
	if len(cs.recent) != 2 || cs.recent[0].Sender != "Grimble the shopkeep" {
		t.Fatalf("recent should note the cast name: %+v", cs.recent)
	}
}

// Eyes (REM-2): 👀 lands on the trigger message at turn start, comes off when
// the turn finishes, and hops to the coalesced follow-up's message.
func TestAsyncTurn_EyesFollowTheWork(t *testing.T) {
	ctx := context.Background()
	cog := &fakeCog{dur: 100 * time.Millisecond, sessionAfter: -1, session: "wi--e--turn", answer: "ok"}
	gc, _, _ := newTestConn(cog)

	var rmu sync.Mutex
	type rx struct{ msgID, op string }
	var reactions []rx
	gc.rawFn = func(v any) error {
		m, ok := v.(map[string]any)
		if !ok || m["type"] != "reaction" {
			return nil // ignore typing pulses
		}
		rmu.Lock()
		reactions = append(reactions, rx{m["messageId"].(string), m["op"].(string)})
		rmu.Unlock()
		return nil
	}

	cs := &channelState{sessionID: "wi--e--turn"}
	gc.channels["eng"] = cs

	gc.maybeStartTurn(ctx, "eng", cs, wireMessage{ID: "msg-1", Sender: "human", Body: "first"})
	if cs.eyedID != "msg-1" {
		t.Fatalf("eyes should be on msg-1, got %q", cs.eyedID)
	}
	// Arrives mid-turn → pending; eyes must hop to it when the follow-up fires.
	gc.maybeStartTurn(ctx, "eng", cs, wireMessage{ID: "msg-2", Sender: "human", Body: "second"})
	pump(gc, ctx, 400*time.Millisecond)

	rmu.Lock()
	defer rmu.Unlock()
	want := []rx{{"msg-1", "add"}, {"msg-1", "remove"}, {"msg-2", "add"}, {"msg-2", "remove"}}
	if len(reactions) != len(want) {
		t.Fatalf("want %d reaction frames %v, got %d: %v", len(want), want, len(reactions), reactions)
	}
	for i, w := range want {
		if reactions[i] != w {
			t.Fatalf("frame %d = %v, want %v (all: %v)", i, reactions[i], w, reactions)
		}
	}
	if cs.eyedID != "" {
		t.Fatalf("eyes should be cleared after the turns, got %q", cs.eyedID)
	}
}

// Persona↔persona triggers (DH-1/D1): another persona's message starts a turn
// only when it names us; the hop budget caps the chain; a human message resets.
func TestPersonaTrigger_MentionGateAndHopBudget(t *testing.T) {
	ctx := context.Background()
	cog := &fakeCog{dur: 10 * time.Millisecond, sessionAfter: -1, session: "wi--pp--turn", answer: "noted"}
	gc, _, _ := newTestConn(cog)
	cs := &channelState{sessionID: "wi--pp--turn"}
	gc.channels["dnd"] = cs

	frame := func(kind, sender, body string) gwOutbound {
		var f gwOutbound
		f.Type = "message"
		f.Channel = "dnd"
		f.Message.ID = "m"
		f.Message.Sender = sender
		f.Message.SenderKind = kind
		f.Message.Body = body
		return f
	}

	// Unaddressed persona message: no turn.
	gc.handle(ctx, frame("persona", "DM Assistant", "the tavern hums with chatter"))
	if cs.busy || cs.hops != 0 {
		t.Fatalf("unaddressed persona message must not trigger (busy=%v hops=%d)", cs.busy, cs.hops)
	}
	// Addressed persona messages: turns fire, spending hops, up to the budget.
	for i := 1; i <= personaHopBudget; i++ {
		gc.handle(ctx, frame("persona", "DM Assistant", "@tester your move"))
		pump(gc, ctx, 80*time.Millisecond) // let the turn finish so busy frees
		if cs.hops != i {
			t.Fatalf("hop %d: hops=%d", i, cs.hops)
		}
	}
	gc.handle(ctx, frame("persona", "DM Assistant", "@tester again"))
	if cs.busy {
		t.Fatal("turn past the hop budget must not start")
	}
	cog.mu.Lock()
	consults := cog.consults
	cog.mu.Unlock()
	if consults != personaHopBudget {
		t.Fatalf("want %d consults, got %d", personaHopBudget, consults)
	}
	// A human message resets the budget and triggers normally.
	gc.handle(ctx, frame("human", "michael", "tester, what do you see?"))
	if cs.hops != 0 || !cs.busy {
		t.Fatalf("human message should reset hops (=%d) and start a turn (busy=%v)", cs.hops, cs.busy)
	}
	pump(gc, ctx, 80*time.Millisecond)
	// Our own messages (platform name) never trigger us.
	gc.selfName = "Tester Prime"
	gc.handle(ctx, frame("persona", "Tester Prime", "@tester echo of myself"))
	if cs.busy {
		t.Fatal("a persona must not trigger on its own messages")
	}
}

// respond_policy "mentioned" (REM-3): unaddressed messages start NO turn (no
// dispatch, no typing, no eyes); addressed ones run normally.
func TestRespondPolicy_MentionedGatesTurns(t *testing.T) {
	ctx := context.Background()
	cog := &fakeCog{dur: 50 * time.Millisecond, sessionAfter: -1, session: "wi--p--turn", answer: "yes?"}
	gc, _, _ := newTestConn(cog)
	gc.respondPolicy = "mentioned"
	cs := &channelState{sessionID: "wi--p--turn"}
	gc.channels["eng"] = cs

	gc.maybeStartTurn(ctx, "eng", cs, wireMessage{ID: "m1", Sender: "human", Body: "talking amongst ourselves"})
	if cs.busy || cs.eyedID != "" {
		t.Fatal("unaddressed message must not start a turn under policy=mentioned")
	}
	gc.maybeStartTurn(ctx, "eng", cs, wireMessage{ID: "m2", Sender: "human", Body: "@tester what do you think?"})
	if !cs.busy {
		t.Fatal("addressed message must start a turn under policy=mentioned")
	}
	pump(gc, ctx, 300*time.Millisecond)
	cog.mu.Lock()
	defer cog.mu.Unlock()
	if cog.consults != 1 {
		t.Fatalf("want exactly 1 consult (the addressed one), got %d", cog.consults)
	}
}

// A human message arriving mid-turn is coalesced into one follow-up turn.
func TestAsyncTurn_MidTurnMessageCoalesced(t *testing.T) {
	ctx := context.Background()
	cog := &fakeCog{dur: 150 * time.Millisecond, sessionAfter: -1, session: "wi--c--turn", answer: "ok"}
	gc, _, _ := newTestConn(cog)
	cs := &channelState{sessionID: "wi--c--turn"}
	gc.channels["eng"] = cs

	gc.maybeStartTurn(ctx, "eng", cs, wireMessage{Sender: "human", Body: "first"})
	// second arrives while busy → should be held as pending, not a 2nd turn.
	gc.maybeStartTurn(ctx, "eng", cs, wireMessage{Sender: "human", Body: "second"})
	if cs.pending == nil || cs.pending.Body != "second" {
		t.Fatalf("mid-turn message should be pending, got %+v", cs.pending)
	}
	pump(gc, ctx, 500*time.Millisecond)

	cog.mu.Lock()
	defer cog.mu.Unlock()
	if cog.consults != 2 {
		t.Fatalf("want exactly 2 consults (initial + one coalesced follow-up), got %d", cog.consults)
	}
}
