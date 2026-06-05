package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

// KeyMaterial is the sidecar's parsed Ed25519 signing keypair plus the
// PEM-encoded public key (safe to publish via /pubkey). The private key lives
// only here and in the persona_host.signing_key row — it is NEVER logged,
// exported over HTTP, or placed in any model context.
type KeyMaterial struct {
	Priv      ed25519.PrivateKey
	Pub       ed25519.PublicKey
	PublicPEM string
}

// Fingerprint is a short, safe-to-log identifier for the public key (the private
// key is never used here).
func (k *KeyMaterial) Fingerprint() string {
	sum := sha256.Sum256(k.Pub)
	return fmt.Sprintf("%x", sum[:6])
}

// LoadOrCreateKey returns the singleton signing key, generating + persisting one
// on first boot. Idempotent and race-safe (see Store.EnsureSigningKey).
func LoadOrCreateKey(ctx context.Context, s *Store) (*KeyMaterial, error) {
	privPEM, pubPEM, err := s.EnsureSigningKey(ctx, generateKeyPEM)
	if err != nil {
		return nil, err
	}
	return parseKeyPEM(privPEM, pubPEM)
}

// generateKeyPEM mints a fresh Ed25519 keypair and PEM-encodes it (PKCS#8
// private, PKIX public).
func generateKeyPEM() (privPEM, pubPEM string, err error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("generate ed25519: %w", err)
	}
	return marshalKeyPEM(priv, pub)
}

func marshalKeyPEM(priv ed25519.PrivateKey, pub ed25519.PublicKey) (privPEM, pubPEM string, err error) {
	pkcs8, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return "", "", fmt.Errorf("marshal private key: %w", err)
	}
	pkix, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", "", fmt.Errorf("marshal public key: %w", err)
	}
	privPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8}))
	pubPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pkix}))
	return privPEM, pubPEM, nil
}

func parseKeyPEM(privPEM, pubPEM string) (*KeyMaterial, error) {
	priv, err := parsePrivatePEM(privPEM)
	if err != nil {
		return nil, err
	}
	pub, err := parsePublicPEM(pubPEM)
	if err != nil {
		return nil, err
	}
	return &KeyMaterial{Priv: priv, Pub: pub, PublicPEM: pubPEM}, nil
}

func parsePrivatePEM(privPEM string) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privPEM))
	if block == nil {
		return nil, errors.New("decode private PEM: no PEM block")
	}
	any, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	priv, ok := any.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is %T, want ed25519", any)
	}
	return priv, nil
}

// parsePublicPEM parses a PKIX public-key PEM into an Ed25519 public key. This is
// exactly what ai-chattermax does with the /pubkey output to verify tokens.
func parsePublicPEM(pubPEM string) (ed25519.PublicKey, error) {
	block, _ := pem.Decode([]byte(pubPEM))
	if block == nil {
		return nil, errors.New("decode public PEM: no PEM block")
	}
	any, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}
	pub, ok := any.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is %T, want ed25519", any)
	}
	return pub, nil
}
