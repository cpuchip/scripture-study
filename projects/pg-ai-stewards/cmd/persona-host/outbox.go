// outbox.go — room_say delivery (expressive-live-personas v1).
//
// A persona calls room_say(body, mood?) mid-turn (substrate tool, r16); it
// writes a stewards.persona_outbox row keyed by the dispatch session. The
// persona-host owns the session→channel map (each channelState holds its
// sessionID), so it drains unposted rows for ITS sessions and posts them to
// the right channel, then stamps posted_at.
//
// The drain runs inside the GatewayConn worker goroutine (see gateway.go Run),
// so it touches gc.channels without locks — same single-goroutine discipline
// as the rest of the connection.
package main

import (
	"context"
	"sort"
)

// OutboxRow is one mid-turn message a persona emitted.
type OutboxRow struct {
	ID         int64
	SessionID  string
	Body       string
	Mood       string // optional emoji; "" if none
	SubPersona string // optional cast member speaking this line (DH-2); "" = the persona itself
	ReactEmoji string // room_react (R21): non-empty = a REACTION on the turn's trigger message, not a post
}

// ClaimOutboxForSessions atomically claims (marks posted) the unposted outbox
// rows for the given sessions and returns them in creation order. Claim-then-post
// (vs post-then-mark) is deliberate: losing a "hang on" beat to a crash between
// claim and send is acceptable; double-posting the same message is not.
func (c *Cognition) ClaimOutboxForSessions(ctx context.Context, sessionIDs []string) ([]OutboxRow, error) {
	if len(sessionIDs) == 0 {
		return nil, nil
	}
	rows, err := c.pool.Query(ctx, `
		UPDATE stewards.persona_outbox o
		   SET posted_at = now()
		  FROM (
		     SELECT id FROM stewards.persona_outbox
		      WHERE posted_at IS NULL AND session_id = ANY($1)
		      ORDER BY created_at
		      FOR UPDATE SKIP LOCKED
		  ) pick
		 WHERE o.id = pick.id
		 RETURNING o.id, o.session_id, o.body, COALESCE(o.mood, ''), COALESCE(o.sub_persona, ''), COALESCE(o.react_emoji, '')`,
		sessionIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []OutboxRow
	for rows.Next() {
		var r OutboxRow
		if err := rows.Scan(&r.ID, &r.SessionID, &r.Body, &r.Mood, &r.SubPersona, &r.ReactEmoji); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// RETURNING order is unspecified; id is monotonic with creation, so sort by
	// it to post a turn's beats in the order the persona emitted them.
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
