package main

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// tokenIssuer is the JWT `iss` claim. ai-chattermax checks it on verify.
const tokenIssuer = "pg-ai-stewards"

// DefaultTokenTTL is the persona token lifetime — short, because a persona
// re-mints on each (re)join. The room verifies exp on every connection.
const DefaultTokenTTL = 15 * time.Minute

// PersonaClaims is the JWT body ai-chattermax verifies: who the persona is and
// which room the token authorizes.
type PersonaClaims struct {
	Slug   string `json:"slug"`
	Name   string `json:"name"`
	Avatar string `json:"avatar,omitempty"`
	Room   string `json:"room"`
	jwt.RegisteredClaims
}

// issuanceRecorder records a minted token's scope + lifetime and returns its jti.
// *Store satisfies it; tests stub it so minting is unit-testable without a DB.
type issuanceRecorder interface {
	RecordIssuance(ctx context.Context, personaID, roomID string, expiresAt time.Time) (jti string, err error)
}

// Minter signs EdDSA persona tokens and records each issuance.
type Minter struct {
	rec issuanceRecorder
	key *KeyMaterial
	now func() time.Time
}

func NewMinter(rec issuanceRecorder, key *KeyMaterial) *Minter {
	return &Minter{rec: rec, key: key, now: time.Now}
}

// MintToken issues a short-lived EdDSA JWT for (persona, room). It records the
// issuance (jti + scope + expiry) FIRST, then signs. The private key is used
// only to sign — never logged, never returned.
func (m *Minter) MintToken(ctx context.Context, p *Persona, room string, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		ttl = DefaultTokenTTL
	}
	now := m.now()
	exp := now.Add(ttl)

	jti, err := m.rec.RecordIssuance(ctx, p.ID, room, exp)
	if err != nil {
		return "", fmt.Errorf("record issuance: %w", err)
	}

	claims := PersonaClaims{
		Slug:   p.Slug,
		Name:   p.DisplayName,
		Avatar: p.AvatarURL,
		Room:   room,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    tokenIssuer,
			Subject:   p.ID,
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(m.key.Priv)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

// VerifyToken parses + verifies a persona token against an Ed25519 public key.
// This is the REFERENCE implementation ai-chattermax mirrors on the room side:
// EdDSA-only, issuer-checked, expiry enforced by the jwt parser.
func VerifyToken(tokenStr string, pub ed25519.PublicKey) (*PersonaClaims, error) {
	claims := &PersonaClaims{}
	tok, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return pub, nil
	},
		jwt.WithValidMethods([]string{"EdDSA"}),
		jwt.WithIssuer(tokenIssuer),
	)
	if err != nil {
		return nil, err
	}
	if !tok.Valid {
		return nil, errors.New("token invalid")
	}
	return claims, nil
}
