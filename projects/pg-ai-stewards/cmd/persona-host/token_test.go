package main

import (
	"context"
	"testing"
	"time"
)

// fakeRec stubs the issuance recorder so MintToken is testable without a DB.
type fakeRec struct{ jti string }

func (f fakeRec) RecordIssuance(_ context.Context, _, _ string, _ time.Time) (string, error) {
	return f.jti, nil
}

func samplePersona() *Persona {
	return &Persona{
		ID:          "11111111-1111-1111-1111-111111111111",
		Slug:        "dm-assistant",
		DisplayName: "DM Assistant",
		AvatarURL:   "https://example.test/dm.png",
		AgentFamily: "study",
	}
}

func TestMintAndVerifyRoundTrip(t *testing.T) {
	key := testKey(t)
	m := &Minter{rec: fakeRec{jti: "test-jti-123"}, key: key, now: time.Now}
	p := samplePersona()

	tok, exp, err := m.MintToken(context.Background(), p, "tavern", 15*time.Minute)
	if err != nil {
		t.Fatalf("mint: %v", err)
	}
	if exp.Before(time.Now()) {
		t.Fatalf("expiry %v is already past", exp)
	}

	claims, err := VerifyToken(tok, key.Pub)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if claims.Issuer != tokenIssuer {
		t.Errorf("iss = %q, want %q", claims.Issuer, tokenIssuer)
	}
	if claims.Subject != p.ID {
		t.Errorf("sub = %q, want %q", claims.Subject, p.ID)
	}
	if claims.Slug != p.Slug || claims.Name != p.DisplayName || claims.Room != "tavern" {
		t.Errorf("claims mismatch: %+v", claims)
	}
	if claims.ID != "test-jti-123" {
		t.Errorf("jti = %q, want test-jti-123", claims.ID)
	}
}

// TestVerifyRejectsWrongKey is the inverse hypothesis: a token signed by one key
// must FAIL verification against a different public key.
func TestVerifyRejectsWrongKey(t *testing.T) {
	signer := testKey(t)
	attacker := testKey(t)
	m := &Minter{rec: fakeRec{jti: "j"}, key: signer, now: time.Now}

	tok, _, err := m.MintToken(context.Background(), samplePersona(), "tavern", time.Minute)
	if err != nil {
		t.Fatalf("mint: %v", err)
	}
	if _, err := VerifyToken(tok, attacker.Pub); err == nil {
		t.Fatal("verify MUST fail against the wrong public key, but it passed")
	}
}

// TestVerifyRejectsExpired proves expiry is enforced: mint with a clock in the
// past so the token is already expired, then verify against real time.
func TestVerifyRejectsExpired(t *testing.T) {
	key := testKey(t)
	past := func() time.Time { return time.Now().Add(-time.Hour) }
	m := &Minter{rec: fakeRec{jti: "j"}, key: key, now: past}

	tok, _, err := m.MintToken(context.Background(), samplePersona(), "tavern", time.Minute)
	if err != nil {
		t.Fatalf("mint: %v", err)
	}
	if _, err := VerifyToken(tok, key.Pub); err == nil {
		t.Fatal("expired token MUST be rejected, but it verified")
	}
}
