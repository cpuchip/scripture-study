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
	"time"
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

func (f fakePersonaStore) UpsertPersonaRoom(_ context.Context, _, _ string) error { return nil }

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
	srv := NewServer(nil, testKey(t), nil)
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
	srv := NewServer(nil, key, nil)
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
	srv := NewServer(store, testKey(t), nil)
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

func TestJoinRoomMintsVerifiableToken(t *testing.T) {
	key := testKey(t)
	store := fakePersonaStore{personas: []Persona{
		{ID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", Slug: "dm-assistant", DisplayName: "DM Assistant", AgentFamily: "fiction"},
	}}
	minter := &Minter{rec: fakeRec{jti: "join-jti"}, key: key, now: time.Now}
	srv := NewServer(store, key, minter)

	body, _ := json.Marshal(JoinRequest{Slug: "dm-assistant", Room: "tavern"})
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/join", strings.NewReader(string(body))))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body=%s)", rr.Code, rr.Body.String())
	}

	var res JoinResult
	if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if res.Room != "tavern" || res.Persona.Slug != "dm-assistant" {
		t.Fatalf("unexpected join result: %+v", res)
	}
	// The minted token in the response must verify against the host's key, with
	// the room the join requested — exactly ai-chattermax's check.
	claims, err := VerifyToken(res.Token, key.Pub)
	if err != nil {
		t.Fatalf("join token does not verify: %v", err)
	}
	if claims.Room != "tavern" || claims.Slug != "dm-assistant" || claims.Subject != store.personas[0].ID {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestJoinRoomUnknownPersona404(t *testing.T) {
	key := testKey(t)
	store := fakePersonaStore{personas: nil}
	srv := NewServer(store, key, &Minter{rec: fakeRec{jti: "x"}, key: key, now: time.Now})
	body, _ := json.Marshal(JoinRequest{Slug: "ghost", Room: "tavern"})
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/join", strings.NewReader(string(body))))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rr.Code)
	}
}

// TestPubkeyNeverLeaksPrivateKey is a guardrail: the /pubkey response must not
// contain any private-key material.
func TestPubkeyNeverLeaksPrivateKey(t *testing.T) {
	srv := NewServer(nil, testKey(t), nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/pubkey", nil))
	body := rr.Body.String()
	for _, bad := range []string{"PRIVATE KEY", "BEGIN PRIVATE", "BEGIN EC PRIVATE", "BEGIN OPENSSH PRIVATE"} {
		if strings.Contains(body, bad) {
			t.Fatalf("/pubkey response leaked %q", bad)
		}
	}
}
