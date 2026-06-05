// autojoin.go — env-driven room joining for #7 v1.
//
// Set CHATTERMAX_WS_BASE (e.g. ws://localhost:8080) and PERSONA_AUTOJOIN
// (e.g. "dm-assistant@tavern,npc-ally@tavern") and persona-host dials each
// persona into its room on boot and runs the turn loop. This is the simplest
// path to a live test; a /activate HTTP endpoint is a later add.
package main

import (
	"context"
	"log"
	"strings"
)

// autojoinSpec is one persona-into-room directive.
type autojoinSpec struct {
	Slug string
	Room string
}

// parseAutojoin parses "slug@room,slug@room" into directives, skipping malformed
// or empty entries.
func parseAutojoin(spec string) []autojoinSpec {
	var out []autojoinSpec
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		slug, room, ok := strings.Cut(part, "@")
		slug, room = strings.TrimSpace(slug), strings.TrimSpace(room)
		if !ok || slug == "" || room == "" {
			log.Printf("autojoin: skipping malformed entry %q (want slug@room)", part)
			continue
		}
		out = append(out, autojoinSpec{Slug: slug, Room: room})
	}
	return out
}

// personaPredicate returns a func reporting whether a sender name is a known
// persona (by display name or slug) — the humans-only gate's "is a persona"
// test. Built from the live persona registry.
func personaPredicate(personas []Persona) func(string) bool {
	names := make(map[string]bool, len(personas)*2)
	for _, p := range personas {
		names[p.DisplayName] = true
		names[p.Slug] = true
	}
	return func(sender string) bool { return names[sender] }
}

// StartAutojoin dials each configured persona into its room and supervises the
// connection (reconnecting on drop) until ctx is cancelled. It returns
// immediately; loops run in the background.
func StartAutojoin(ctx context.Context, store *Store, cog *Cognition, wsBase, spec string) error {
	specs := parseAutojoin(spec)
	if len(specs) == 0 {
		return nil
	}

	personas, err := store.ListPersonas(ctx)
	if err != nil {
		return err
	}
	isPersona := personaPredicate(personas)
	bySlug := make(map[string]Persona, len(personas))
	for _, p := range personas {
		bySlug[p.Slug] = p
	}

	for _, s := range specs {
		p, ok := bySlug[s.Slug]
		if !ok {
			log.Printf("autojoin: persona %q not registered — skipping room %q", s.Slug, s.Room)
			continue
		}
		rc := NewRoomConn(p, s.Room, wsBase, cog, isPersona)
		go superviseRoom(ctx, rc)
	}
	return nil
}

// superviseRoom runs a room connection, re-dialing after a drop until ctx ends.
func superviseRoom(ctx context.Context, rc *RoomConn) {
	for {
		if ctx.Err() != nil {
			return
		}
		if err := rc.Run(ctx); err != nil {
			log.Printf("[%s@%s] disconnected: %v", rc.persona.Slug, rc.room, err)
		}
		if ctx.Err() != nil {
			return
		}
		log.Printf("[%s@%s] reconnecting in %s", rc.persona.Slug, rc.room, roomLoopRetryDelay)
		if err := sleepCtx(ctx, roomLoopRetryDelay); err != nil {
			return
		}
	}
}
