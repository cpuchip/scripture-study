package main

import "context"

// defaultPersonas are seeded on every boot (idempotent upsert). They back the
// ai-chattermax D&D MVP (#12). agent_family must name an ACTIVE substrate agent
// family — "fiction" is the substrate's D&D / NPC family (resolved at dispatch
// in #7). A third persona is a row here, not a new deployment.
var defaultPersonas = []Persona{
	{Slug: "dm-assistant", DisplayName: "DM Assistant", AgentFamily: "fiction"},
	{Slug: "npc-ally", DisplayName: "NPC Ally", AgentFamily: "fiction"},
}

// SeedDefaultPersonas upserts the built-in personas. Idempotent: re-running
// updates display/family but never duplicates (slug is unique).
func SeedDefaultPersonas(ctx context.Context, s *Store) error {
	for _, p := range defaultPersonas {
		if _, err := s.UpsertPersona(ctx, p); err != nil {
			return err
		}
	}
	return nil
}
