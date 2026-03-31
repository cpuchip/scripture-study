package enricher

import (
	"testing"
)

func TestParseProfile(t *testing.T) {
	// Sample output from a real LLM response (magistral-style terse output)
	input := `KEYWORDS: joy, welcome, belonging, sacrament, worship, testimony, repentance, Jesus Christ, church membership, love, community, baptism, faith, hope, charity

SUMMARY: Elder Kearon invites all members and investigators to rediscover the joy inherent in gospel living. He emphasizes that the Church of Jesus Christ is fundamentally a church of joy, rooted in the Savior's atoning sacrifice, and calls on members to make every person feel welcome and valued in their congregations.

KEY_QUOTE: "Welcome to the church of joy!"

TEACHING_PROFILE:
DOMINANT: help_come_to_christ, invite
MODE: enacted (models the welcoming spirit he advocates)
PATTERN: invitation→doctrine→testimony

REASONING: Teach — Christ is referenced throughout as the source of joy, but 
the sustained focus is on the communal experience of church membership rather 
than Christ's person or mission directly. Help — the talk creates a clear 
mechanism for coming to Christ through belonging and welcome. Love — Kearon 
demonstrates warmth and creates safety, naming specific struggles of 
investigators and new members. Spirit — the talk is warm but scripted; no 
visible departure from prepared remarks. Doctrine — sacrament theology and 
covenant concepts are present but serve the invitation arc. Invite — the 
entire talk builds toward "welcome them, include them, be joyful."

TEACH_SCORE: 5 Christ present as source of joy but not the explored subject
HELP_SCORE: 7 Creates concrete mechanism — belonging and welcome as pathway to Christ
LOVE_SCORE: 5 Warm and pastoral, names some struggles, but more prescribed than demonstrated
SPIRIT_SCORE: 4 Warm delivery but no departure from script; formulaic invocation
DOCTRINE_SCORE: 4 Sacrament theology referenced but not sustained exposition
INVITE_SCORE: 7 The entire talk is structured as an invitation to welcome and belong`

	profile, err := parseProfile(input)
	if err != nil {
		t.Fatalf("parseProfile failed: %v", err)
	}

	// Check scores
	tests := []struct {
		name     string
		got, want int
	}{
		{"teach", profile.Teach, 5},
		{"help", profile.Help, 7},
		{"love", profile.Love, 5},
		{"spirit", profile.Spirit, 4},
		{"doctrine", profile.Doctrine, 4},
		{"invite", profile.Invite, 7},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: got %d, want %d", tt.name, tt.got, tt.want)
		}
	}

	// Check text fields
	if profile.Dominant == "" {
		t.Error("dominant should not be empty")
	}
	if profile.Mode != "enacted" {
		t.Errorf("mode: got %q, want %q", profile.Mode, "enacted")
	}
	if profile.Pattern == "" {
		t.Error("pattern should not be empty")
	}
	if profile.Keywords == "" {
		t.Error("keywords should not be empty")
	}
	if profile.KeyQuote == "" {
		t.Error("key_quote should not be empty")
	}
	if profile.Summary == "" {
		t.Error("summary should not be empty")
	}
	if profile.Reasoning == "" {
		t.Error("reasoning should not be empty")
	}
}

func TestParseProfileBoldScores(t *testing.T) {
	// Some models wrap scores in **bold** or [brackets]
	input := `KEYWORDS: test
SUMMARY: test summary
KEY_QUOTE: "test quote"
TEACHING_PROFILE:
DOMINANT: doctrine
MODE: doctrinal
PATTERN: test

REASONING: Test reasoning text here.

TEACH_SCORE: **5** justification
HELP_SCORE: [3] justification
LOVE_SCORE: 2 justification
SPIRIT_SCORE:  **7**  justification
DOCTRINE_SCORE: [9] justification
INVITE_SCORE: 4 justification`

	profile, err := parseProfile(input)
	if err != nil {
		t.Fatalf("parseProfile failed: %v", err)
	}

	tests := []struct {
		name     string
		got, want int
	}{
		{"teach", profile.Teach, 5},
		{"help", profile.Help, 3},
		{"love", profile.Love, 2},
		{"spirit", profile.Spirit, 7},
		{"doctrine", profile.Doctrine, 9},
		{"invite", profile.Invite, 4},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: got %d, want %d", tt.name, tt.got, tt.want)
		}
	}
}

func TestExtractBetween(t *testing.T) {
	text := "SUMMARY: This is a test summary.\nKEY_QUOTE: test"
	got := extractBetween(text, "SUMMARY:", "KEY_QUOTE:")
	if got != "This is a test summary." {
		t.Errorf("extractBetween: got %q", got)
	}
}
