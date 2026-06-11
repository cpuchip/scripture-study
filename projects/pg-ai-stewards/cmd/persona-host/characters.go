// characters.go — promoted characters (DH-2 promotion, ratified 2026-06-10).
//
// A cast member is display identity (a platform sub-persona). A PROMOTED
// character is also a MIND: its own substrate session — own memory, own LLM
// loop — room-agnostic, because the mind belongs to the character, not the
// room. The owning persona's connection still posts every line (attribution
// via subPersona), so the platform sees no difference between a facet and a
// promoted character. Party's cast promotes by default (every PC is its own
// mind); the DM's scene NPCs stay facets unless explicitly promoted.
package main

import (
	"context"
	"fmt"
)

// Character is one named cast member, host-side.
type Character struct {
	ID            string
	PersonaSlug   string
	Name          string
	SessionID     string // the character's own substrate session ("" until first turn)
	ModelOverride string // stored now; applied when per-character model routing lands
	Prompt        string // personality/sheet seed for the character's turn-zero
	Promoted      bool
}

// EnsureCharacter finds-or-creates the character record for (persona, name).
// New characters take the persona's default_promote.
func (c *Cognition) EnsureCharacter(ctx context.Context, personaSlug, name string, promote bool) (Character, error) {
	var ch Character
	err := c.pool.QueryRow(ctx, `
		INSERT INTO persona_host.characters (persona_slug, name, promoted)
		VALUES ($1, $2, $3)
		ON CONFLICT (persona_slug, lower(name)) DO UPDATE SET updated_at = now()
		RETURNING id, persona_slug, name, COALESCE(session_id,''), COALESCE(model_override,''), COALESCE(prompt,''), promoted`,
		personaSlug, name, promote,
	).Scan(&ch.ID, &ch.PersonaSlug, &ch.Name, &ch.SessionID, &ch.ModelOverride, &ch.Prompt, &ch.Promoted)
	if err != nil {
		return Character{}, fmt.Errorf("ensure character %q: %w", name, err)
	}
	return ch, nil
}

// SaveCharacterSession persists a character's newly-spawned session id.
func (c *Cognition) SaveCharacterSession(ctx context.Context, id, sessionID string) error {
	_, err := c.pool.Exec(ctx, `
		UPDATE persona_host.characters SET session_id = $2, updated_at = now()
		WHERE id = $1 AND COALESCE(session_id,'') = ''`, id, sessionID)
	return err
}
