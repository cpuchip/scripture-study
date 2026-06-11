package main

import (
	"strings"
	"testing"
)

func TestIsSilence(t *testing.T) {
	cases := map[string]bool{
		"SILENCE":          true,
		"  silence  ":      true,
		"Silence":          true,
		"":                 true,
		"   ":              true,
		"hello there":      false,
		"SILENCE is great": false, // a real reply that merely contains the word
	}
	for in, want := range cases {
		if got := IsSilence(in); got != want {
			t.Errorf("IsSilence(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestMentionsAnother(t *testing.T) {
	// Callie's names after the 2026-06-11 rename (slug, display, platform).
	mine := []string{"callie", "Callie", "Callie"}
	cases := []struct {
		body string
		want bool
	}{
		// The 2026-06-11 incident: a sheet request explicitly @-addressed to the
		// DM — Callie (judgment policy) must treat it as the DM's spotlight.
		{"@DMAssistant can you create me a night elf character for me to use? make him a monk", true},
		{"@DM-Assistant set the scene", true},  // dashed mention form
		{"@ClaudeCodetest look at this", true}, // humans get spotlights too
		{"@Callie make me a sheet", false},     // it's ours — not another's
		{"I attack the goblin", false},         // no mentions at all
		{"@everyone roll initiative", false},   // broadcast is nobody's spotlight
		{"@all gather round", false},           // broadcast alias
		{"mail me at dm@example.com", false},   // emails are not mentions
	}
	for _, c := range cases {
		if got := mentionsAnother(c.body, mine); got != c.want {
			t.Errorf("mentionsAnother(%q) = %v, want %v", c.body, got, c.want)
		}
	}
	// "@callie and also @DMAssistant" names another persona AND us. The helper
	// reports the foreign mention, but the maybeStartTurn gate never reaches it
	// when we're addressed — assert both halves so the contract stays visible.
	both := "@callie and also @DMAssistant"
	if !mentionsAnother(both, mine) {
		t.Errorf("mentionsAnother(%q) should see the foreign mention", both)
	}
	if !isAddressed(both, mine...) {
		t.Errorf("isAddressed(%q) must be true — the gate yields to addressing first", both)
	}
}

func TestIsAddressed(t *testing.T) {
	const slug, name = "dm-assistant", "DM Assistant"
	cases := []struct {
		body string
		want bool
	}{
		{"@dm-assistant set the scene", true},          // @slug
		{"@DM Assistant help", true},                    // @display name
		{"@DMAssistant help", true},                     // @display name, spaces stripped
		{"DM Assistant, what do I see?", true},          // name in passing
		{"hey dm assistant can you describe it", true},  // case-insensitive name
		{"I attack the goblin", false},                  // no mention
		{"the npc waves at bob", false},                 // unrelated
	}
	for _, c := range cases {
		if got := isAddressed(c.body, slug, name); got != c.want {
			t.Errorf("isAddressed(%q) = %v, want %v", c.body, got, c.want)
		}
	}

	// The 2026-06-10 SILENCE bug: the room shows a platform name ("Chattercode")
	// that differs from both the slug and the host display name ("Codewright").
	// Plain slug and platform-name addressing must both count.
	if !isAddressed("chattercode, which file defines key validation?", "chattercode", "Codewright", "") {
		t.Error("plain slug (no @) must count as addressed")
	}
	if !isAddressed("hey Chattercode can you check this", "chattercode", "Codewright", "Chattercode") {
		t.Error("platform display name must count as addressed")
	}
	if isAddressed("I attack the goblin", "chattercode", "Codewright", "Chattercode") {
		t.Error("unrelated message must not be addressed")
	}

	// The 2026-06-10 Vex/Vexa false wake: a cast member "Vex" must not match
	// inside the unrelated name "Vexa Nightbloom" — word boundaries only.
	if isAddressed("@Party Vexa Nightbloom wants to slip past the sentry", "vex", "Vex") {
		t.Error("\"Vex\" must not match inside \"Vexa\"")
	}
	if !isAddressed("Vex, block the door!", "vex", "Vex") {
		t.Error("\"Vex,\" with punctuation boundary must match")
	}
	if !isAddressed("what does Vex think?", "vex", "Vex") {
		t.Error("\"Vex\" followed by space must match")
	}
}

func TestShouldConsider(t *testing.T) {
	rc := &RoomConn{
		persona:   Persona{Slug: "dm-assistant", DisplayName: "DM Assistant"},
		isPersona: func(s string) bool { return s == "DM Assistant" || s == "NPC Ally" },
	}
	cases := []struct {
		wm   wireMessage
		want bool
		why  string
	}{
		{wireMessage{Sender: "DM Assistant", Body: "hi"}, false, "own message"},
		{wireMessage{Sender: "NPC Ally", Body: "hi"}, false, "another persona (humans-only)"},
		{wireMessage{Sender: "michael", Body: "hello"}, true, "human"},
		{wireMessage{Sender: "michael", Body: "   "}, false, "empty body"},
	}
	for _, c := range cases {
		if got := rc.shouldConsider(c.wm); got != c.want {
			t.Errorf("shouldConsider(%+v) = %v, want %v (%s)", c.wm, got, c.want, c.why)
		}
	}
}

func TestBuildTurnZeroFraming(t *testing.T) {
	p := Persona{Slug: "dm-assistant", DisplayName: "DM Assistant", Prompt: "You are a warm theatrical DM."}
	recent := []wireMessage{
		{Sender: "michael", Body: "anyone here?"},
		{Sender: "alice", Body: "just me"},
	}
	trigger := wireMessage{Sender: "michael", Body: "DM Assistant, set the scene"}
	out := buildTurnZeroFraming(p, "tavern", recent, trigger, true, "DM Assistant")

	for _, want := range []string{
		`"DM Assistant"`, "tavern", "warm theatrical DM",
		"michael: DM Assistant, set the scene", "directly addressed", "SILENCE",
		"anyone here?", // recent context included
	} {
		if !strings.Contains(out, want) {
			t.Errorf("turn-zero framing missing %q\n---\n%s", want, out)
		}
	}
	// Matching platform name: no identity-bridge line needed.
	if strings.Contains(out, "you appear under the name") {
		t.Errorf("matching names should not emit the identity bridge:\n%s", out)
	}
}

// The Codewright/Chattercode split: when the platform shows a different name
// than the character, the framing must bridge the identity explicitly.
func TestBuildTurnZeroFraming_PlatformNameBridge(t *testing.T) {
	p := Persona{Slug: "chattercode", DisplayName: "Codewright", Prompt: "You are Codewright."}
	trigger := wireMessage{Sender: "tester", Body: "chattercode, where is the migration runner?"}
	out := buildTurnZeroFraming(p, "Engineering", nil, trigger, true, "Chattercode")
	for _, want := range []string{
		`you appear under the name "Chattercode"`,
		`messages addressed to "Chattercode"`,
		"your own earlier messages",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("identity bridge missing %q\n---\n%s", want, out)
		}
	}
}

func TestBuildTurnZeroFraming_NoPromptFallback(t *testing.T) {
	p := Persona{Slug: "ghost", DisplayName: "Ghost"} // no Prompt
	out := buildTurnZeroFraming(p, "void", nil, wireMessage{Sender: "x", Body: "boo"}, false, "")
	if !strings.Contains(out, "You are Ghost.") {
		t.Errorf("expected character fallback, got:\n%s", out)
	}
	if strings.Contains(out, "directly addressed") {
		t.Errorf("unaddressed turn should not claim direct address:\n%s", out)
	}
	if !strings.Contains(out, "(you just joined") {
		t.Errorf("empty recent should render the just-joined placeholder:\n%s", out)
	}
}

func TestBuildConsultFraming(t *testing.T) {
	out := buildConsultFraming(wireMessage{Sender: "bob", Body: "I open the door"}, true)
	for _, want := range []string{"bob: I open the door", "directly addressed", "SILENCE"} {
		if !strings.Contains(out, want) {
			t.Errorf("consult framing missing %q\n---\n%s", want, out)
		}
	}
}

func TestFormatRecentExcludesTrigger(t *testing.T) {
	trigger := wireMessage{Sender: "michael", Body: "set the scene"}
	recent := []wireMessage{
		{Sender: "alice", Body: "hi"},
		trigger, // must be excluded — it's shown separately
	}
	out := formatRecent(recent, trigger)
	if strings.Contains(out, "set the scene") {
		t.Errorf("formatRecent should exclude the trigger, got:\n%s", out)
	}
	if !strings.Contains(out, "alice: hi") {
		t.Errorf("formatRecent dropped a non-trigger line:\n%s", out)
	}
}

func TestNoteBoundsBuffer(t *testing.T) {
	rc := &RoomConn{}
	for i := 0; i < recentBufferSize+10; i++ {
		rc.note(wireMessage{Sender: "x", Body: string(rune('a' + i%26))})
	}
	if len(rc.recent) != recentBufferSize {
		t.Errorf("recent buffer = %d, want capped at %d", len(rc.recent), recentBufferSize)
	}
}

func TestParseAutojoin(t *testing.T) {
	got := parseAutojoin("dm-assistant@tavern, npc-ally@tavern ,,bad-entry, @noroom, slug@")
	want := []autojoinSpec{
		{Slug: "dm-assistant", Room: "tavern"},
		{Slug: "npc-ally", Room: "tavern"},
	}
	if len(got) != len(want) {
		t.Fatalf("parseAutojoin = %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entry %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestPersonaPredicate(t *testing.T) {
	isP := personaPredicate([]Persona{
		{Slug: "dm-assistant", DisplayName: "DM Assistant"},
		{Slug: "npc-ally", DisplayName: "NPC Ally"},
	})
	for _, s := range []string{"DM Assistant", "dm-assistant", "npc-ally", "NPC Ally"} {
		if !isP(s) {
			t.Errorf("personaPredicate(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"michael", "alice", "human"} {
		if isP(s) {
			t.Errorf("personaPredicate(%q) = true, want false (humans-only gate)", s)
		}
	}
}
