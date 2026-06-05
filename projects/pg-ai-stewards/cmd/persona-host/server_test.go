package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fakePersonaStore stubs the read surface so handlers test without a DB.
type fakePersonaStore struct{ personas []Persona }

func (f fakePersonaStore) ListPersonas(_ context.Context) ([]Persona, error) {
	return f.personas, nil
}

func (f fakePersonaStore) PersonaBySlug(_ context.Context, slug string) (*Persona, error) {
	for i := range f.personas {
		if f.personas[i].Slug == slug {
			return &f.personas[i], nil
		}
	}
	return nil, errors.New("persona not found")
}

// testKey builds an in-memory KeyMaterial without touching the DB — it exercises
// the same marshal/parse path the DB-backed key uses.
func testKey(t *testing.T) *KeyMaterial {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	privPEM, pubPEM, err := marshalKeyPEM(priv, pub)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	km, err := parseKeyPEM(privPEM, pubPEM)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return km
}

func TestHealthz(t *testing.T) {
	srv := NewServer(nil, testKey(t))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
}

// TestPubkeyServesParseablePublicKey proves the /pubkey output is exactly what
// ai-chattermax needs: a PEM that parses back to the signing public key.
func TestPubkeyServesParseablePublicKey(t *testing.T) {
	key := testKey(t)
	srv := NewServer(nil, key)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/pubkey", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	pub, err := parsePublicPEM(rr.Body.String())
	if err != nil {
		t.Fatalf("served /pubkey is unparseable: %v", err)
	}
	if !pub.Equal(key.Pub) {
		t.Fatal("served /pubkey does not match the signing public key")
	}
}

func TestPersonasList(t *testing.T) {
	store := fakePersonaStore{personas: []Persona{
		{Slug: "dm-assistant", DisplayName: "DM Assistant", AgentFamily: "fiction"},
		{Slug: "npc-ally", DisplayName: "NPC Ally", AgentFamily: "fiction"},
	}}
	srv := NewServer(store, testKey(t))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/personas", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	var got []personaView
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 2 || got[0].Slug != "dm-assistant" || got[1].Slug != "npc-ally" {
		t.Fatalf("unexpected roster: %+v", got)
	}
}

// TestPubkeyNeverLeaksPrivateKey is a guardrail: the /pubkey response must not
// contain any private-key material.
func TestPubkeyNeverLeaksPrivateKey(t *testing.T) {
	srv := NewServer(nil, testKey(t))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/pubkey", nil))
	body := rr.Body.String()
	for _, bad := range []string{"PRIVATE KEY", "BEGIN PRIVATE", "BEGIN EC PRIVATE", "BEGIN OPENSSH PRIVATE"} {
		if strings.Contains(body, bad) {
			t.Fatalf("/pubkey response leaked %q", bad)
		}
	}
}
