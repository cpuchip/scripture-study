package main

import (
	"context"
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
