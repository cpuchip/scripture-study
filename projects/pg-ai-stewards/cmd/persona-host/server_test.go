package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
