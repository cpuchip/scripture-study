package main

import (
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
)

// Regression guard for the 2026-05-29 brainstorm-run bug. The start_brainstorm
// `models` param was typed json.RawMessage (= []byte), which the MCP SDK's
// schema generator reflects as an ARRAY-OF-INT — so every per-lens override
// failed client-side with InputValidationError, and the per-lens model routing
// could only be reached by calling the SQL function directly via psql.
//
// It must reflect as a free-form OBJECT. map[string]any does; json.RawMessage
// does not. This test fails if the type ever regresses.
func TestStartBrainstormModelsSchemaIsObject(t *testing.T) {
	s, err := jsonschema.For[StartBrainstormInput](nil)
	if err != nil {
		t.Fatalf("schema generation failed: %v", err)
	}
	models, ok := s.Properties["models"]
	if !ok || models == nil {
		t.Fatal("start_brainstorm input schema has no 'models' property")
	}
	if models.Type != "object" {
		t.Fatalf("models schema Type = %q, want \"object\" — json.RawMessage regresses this to \"array\" and breaks per-lens overrides over MCP", models.Type)
	}
	// The bug shape was specifically an array. Guard against it directly so a
	// future []byte-ish regression is caught even if the generator's
	// representation of non-array types shifts.
	if models.Type == "array" {
		t.Fatal("models schema is an array — the json.RawMessage bug has regressed")
	}
}
