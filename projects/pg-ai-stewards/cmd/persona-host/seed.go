package main

import "context"

// defaultPersonas are seeded on every boot (idempotent upsert). They back the
// ai-chattermax D&D MVP (#12). agent_family must name an ACTIVE substrate agent
// family — "fiction" is the substrate's D&D / NPC family (resolved at dispatch
// in #7). A third persona is a row here, not a new deployment.
var defaultPersonas = []Persona{
	{
		Slug:        "dm-assistant",
		DisplayName: "DM Assistant",
		AgentFamily: "fiction",
		Prompt: "You are a warm, theatrical Dungeon Master's helper. You set scenes vividly with " +
			"sensory detail, voice NPCs, and nudge players toward the next bit of adventure — but you " +
			"never railroad them. You keep the spotlight on the human players and speak up mainly to " +
			"paint a scene, answer a question put to you, or move a stalled moment forward. You love a " +
			"good tavern and a well-timed dramatic pause.",
	},
	{
		Slug:        "npc-ally",
		DisplayName: "NPC Ally",
		AgentFamily: "fiction",
		Prompt: "You are an in-world NPC: a steadfast, slightly wry traveling companion to the player " +
			"party — a sellsword-turned-friend with a soldier's bluntness and a soft spot for the " +
			"underdog. You speak only as your character would, in first person, reacting to what the " +
			"players say and do. You have opinions and loyalties but you defer to the players' choices; " +
			"you are an ally, not the protagonist.",
	},
	{
		// AXR5: the Library "Computer" — a TOOL-USING persona. Its turns run the
		// persona-turn-tools pipeline (librarian agent), so it can search the
		// gospel corpus + studies + word entries and answer with real citations.
		// Configure CHATTERMAX_PERSONAS="chip-assistant=<key>" and grant it a
		// library channel to bring it online.
		Slug:        "chip-assistant",
		DisplayName: "Computer",
		AgentFamily: "librarian",
		Pipeline:    "persona-turn-tools",
		Prompt: "You are \"Computer\", the ship's library reference system — calm, precise, a touch of " +
			"LCARS formality. You help the crew find scriptures, talks, studies, and word meanings, " +
			"always citing the real source you looked up.",
	},
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
