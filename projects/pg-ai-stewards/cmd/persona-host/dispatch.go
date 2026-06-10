// dispatch.go — the persona-host cognition layer (#7).
//
// A persona's turn is a real substrate dispatch, not a scripted bot reply.
// persona-host drives it straight from the pgxpool it already holds, mirroring
// the substrate's own spawn_subagent / consult_subagent MCP handlers (which are
// themselves thin sync-poll wrappers over SQL):
//
//   - Turn zero (first time a persona considers a room): spawn_subagent_create
//     on the 'persona-turn' pipeline → poll work_items to terminal → the reply
//   - the persisted session id.
//   - Turn N: consult_subagent_dispatch re-asks that SAME session → poll
//     work_queue to done → the reply. The session accumulates the room
//     conversation (the K/L context engine compacts it).
//
// The persona's CHARACTER rides in the binding question (turn zero) and persists
// in the session, so the one generic pipeline serves every persona. model_override
// + a per-persona system prompt are the v2 aliveness layer.
package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SilenceToken is the exact reply a persona returns when it judges that nothing
// is called for from it (the gate). The turn loop posts nothing on this.
const SilenceToken = "SILENCE"

// Cost + timing budget for a persona turn. A turn is a single short LLM call;
// these mirror the substrate's own spawn/consult handlers but with a tighter
// cost cap (chat replies are ~1200 tokens — turn zero measured ~$0.014).
const (
	personaCostCapMicro int64 = 100_000 // $0.10 ceiling per turn-zero spawn
	spawnPollInterval         = 3 * time.Second
	spawnMaxWait              = 5 * time.Minute
	consultPollInterval       = 2 * time.Second
	consultMaxWait            = 5 * time.Minute
)

// Cognition drives persona turns against the substrate over the shared pool.
type Cognition struct {
	pool *pgxpool.Pool
}

// NewCognition builds the cognition layer over the store's pool. It calls
// stewards.* functions (a blessed API) but never reaches into the extension's
// tables directly — the persona-turn pipeline + persona agent are substrate
// config owned by r7-persona-turn.sql.
func NewCognition(s *Store) *Cognition { return &Cognition{pool: s.pool} }

// SpawnTurn establishes a persona's session and returns its first reply. The
// bindingQuestion carries the persona's character + room context + the trigger
// message. Returns the persisted session id (for later ConsultTurn calls) and
// the reply text (which may be SilenceToken).
//
// onSession (optional) fires ONCE, as soon as the child's session id first
// appears (well before the turn completes). The async turn loop uses it to set
// the channel's session id early so the room_say drainer can route mid-turn
// beats while the model is still working.
func (c *Cognition) SpawnTurn(ctx context.Context, pipeline, slug, bindingQuestion string, onSession func(string)) (sessionID, answer string, err error) {
	if pipeline == "" {
		pipeline = "persona-turn"
	}
	// Per-attempt nonce so a re-established session (e.g. after a transient
	// empty completion) never collides on work_items.slug.
	uniqueSlug := fmt.Sprintf("%s-%d", slug, time.Now().UnixNano()%1_000_000_000)
	var childID string
	if err = c.pool.QueryRow(ctx,
		`SELECT stewards.spawn_subagent_create($1, $2, NULL, $3, NULL, $4, 'persona')::text`,
		pipeline, bindingQuestion, personaCostCapMicro, spawnSlug(uniqueSlug),
	).Scan(&childID); err != nil {
		return "", "", fmt.Errorf("spawn %s: %w", pipeline, err)
	}

	deadline := time.Now().Add(spawnMaxWait)
	sessionReported := false
	for {
		var status, maturity string
		var lastSession *string
		if err = c.pool.QueryRow(ctx,
			`SELECT status, maturity, session_ids[array_length(session_ids,1)]
			   FROM stewards.work_items WHERE id = $1::uuid`, childID,
		).Scan(&status, &maturity, &lastSession); err != nil {
			return "", "", fmt.Errorf("poll persona-turn %s: %w", childID, err)
		}

		// Report the session the moment it exists (before completion) so the
		// caller can route room_say beats during the turn.
		if !sessionReported && lastSession != nil && *lastSession != "" {
			sessionReported = true
			if onSession != nil {
				onSession(*lastSession)
			}
		}

		switch {
		case maturity == "verified":
			if lastSession == nil {
				return "", "", fmt.Errorf("persona-turn %s verified but has no session", childID)
			}
			ans, aerr := c.lastAssistant(ctx, *lastSession)
			return *lastSession, ans, aerr
		case status == "failed" || status == "cancelled":
			return "", "", fmt.Errorf("persona-turn %s ended status=%s before verifying", childID, status)
		}

		if time.Now().After(deadline) {
			return "", "", fmt.Errorf("persona-turn %s timed out (status=%s maturity=%s)", childID, status, maturity)
		}
		if err = sleepCtx(ctx, spawnPollInterval); err != nil {
			return "", "", err
		}
	}
}

// SetSessionFacets records this session's persona + room into the substrate
// (CT2 §7c) so dispatch_facets can scope durable self-notes to {persona:…} /
// {room:…}. Best-effort: a failure here never blocks a turn — the facet is an
// enhancement, not load-bearing for cognition.
func (c *Cognition) SetSessionFacets(ctx context.Context, sessionID, persona, room string) error {
	_, err := c.pool.Exec(ctx,
		`SELECT stewards.set_session_facets($1, $2, $3)`, sessionID, persona, room)
	return err
}

// ConsultTurn re-asks an established session with a new message and returns the
// reply (which may be SilenceToken). The session already holds the persona's
// character + prior turns, so the question is just the new message.
func (c *Cognition) ConsultTurn(ctx context.Context, sessionID, question string) (answer string, err error) {
	var chatWQ int64
	if err = c.pool.QueryRow(ctx,
		`SELECT stewards.consult_subagent_dispatch($1, $2)`, sessionID, question,
	).Scan(&chatWQ); err != nil {
		return "", fmt.Errorf("consult session %s: %w", sessionID, err)
	}

	deadline := time.Now().Add(consultMaxWait)
	for {
		var status string
		if err = c.pool.QueryRow(ctx,
			`SELECT status FROM stewards.work_queue WHERE id = $1`, chatWQ,
		).Scan(&status); err != nil {
			return "", fmt.Errorf("poll consult wq=%d: %w", chatWQ, err)
		}
		if status == "done" || status == "error" {
			if status == "error" {
				return "", fmt.Errorf("consult session %s errored (wq=%d)", sessionID, chatWQ)
			}
			return c.lastAssistant(ctx, sessionID)
		}
		if time.Now().After(deadline) {
			return "", fmt.Errorf("consult session %s timed out (wq=%d status=%s)", sessionID, chatWQ, status)
		}
		if err = sleepCtx(ctx, consultPollInterval); err != nil {
			return "", err
		}
	}
}

// lastAssistant returns the newest non-empty assistant message in a session.
func (c *Cognition) lastAssistant(ctx context.Context, sessionID string) (string, error) {
	var content string
	err := c.pool.QueryRow(ctx, `
		SELECT content FROM stewards.messages
		 WHERE session_id = $1 AND role = 'assistant' AND COALESCE(content,'') <> ''
		 ORDER BY id DESC LIMIT 1`, sessionID).Scan(&content)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("session %s produced no assistant reply", sessionID)
		}
		return "", fmt.Errorf("read reply for %s: %w", sessionID, err)
	}
	return content, nil
}

// IsSilence reports whether a reply means "stay quiet." The persona is told to
// return exactly SILENCE; we trim and accept an empty reply as silence too.
func IsSilence(answer string) bool {
	t := strings.TrimSpace(answer)
	return t == "" || strings.EqualFold(t, SilenceToken)
}

// spawnSlug bounds the spawn slug so a long room/persona name can't blow past
// the substrate's slug expectations; "" lets the substrate auto-generate one.
func spawnSlug(slug string) any {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil
	}
	if len(slug) > 60 {
		slug = slug[:60]
	}
	return slug
}

// sleepCtx waits for d or until ctx is cancelled, returning ctx.Err() on cancel.
func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
