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
