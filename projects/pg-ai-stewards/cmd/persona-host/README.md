# persona-host

The substrate's optional **persona sidecar** (the `substrate-persona-concept`
proposal, AX2). It owns persona identity, EdDSA/JWT credential minting, and the
ai-chattermax room handshake — kept **out of the core pg-ai-stewards extension**
so a general substrate install never runs it. Like `cmd/coder-mcp`, it's a Go
sidecar; one host serves **many** personas.

## What it is / isn't

- **Is:** a small HTTP service + a `persona_host` schema (which it creates on
  boot) in the substrate's Postgres. It mints short-lived scoped tokens that let
  substrate-hosted personas authenticate into ai-chattermax rooms.
- **Isn't:** in the extension. No pgrx crate, no extension rebuild. The core
  substrate only supplies *cognition per turn* (via its existing dispatch); it
  never knows "persona", "room", or "JWT".

## Run

```
STEWARDS_DSN="postgres://stewards:…@host:5432/stewards" persona-host -addr :8090
persona-host -smoke      # connect → migrate → key → mint/verify → seed → join
```

On boot it: applies the `persona_host` migration (idempotent), generates the
Ed25519 signing key if absent, seeds the default personas, and serves HTTP.

## HTTP surface

| Method | Path | Purpose |
|---|---|---|
| GET | `/healthz` | `{"status":"ok"}` |
| GET | `/pubkey` | the Ed25519 **public** key (PEM) — configure this into ai-chattermax |
| GET | `/personas` | active persona roster (slug, display_name, avatar_url, agent_family) |
| POST | `/join` | `{ "slug", "room" }` → `{ token, persona, room, expires_at }` |

The **private** key never leaves the sidecar — it is stored in
`persona_host.signing_key`, never logged, never served, never placed in any
model context (same class as the coder's GitHub token).

## Token contract (what ai-chattermax verifies)

A persona token is a JWT, `alg: EdDSA` (Ed25519):

| Claim | Value |
|---|---|
| `iss` | `pg-ai-stewards` |
| `sub` | persona id (uuid) |
| `slug`, `name`, `avatar` | persona identity |
| `room` | the room this token authorizes |
| `iat` / `exp` | issued-at / expiry (~15 min) |
| `jti` | issuance id (recorded in `persona_host.token_issuance`) |

### Verify side (ai-chattermax) — no callback

Fetch `/pubkey` once (or set `STEWARDS_PERSONA_PUBKEY`), parse it as a PKIX
Ed25519 public key, then on each persona connection verify the token against it:

```go
pub, _ := parsePublicPEM(stewardsPersonaPubkeyPEM) // x509.ParsePKIXPublicKey
claims := &PersonaClaims{}
tok, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
    if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
        return nil, fmt.Errorf("unexpected alg %v", t.Header["alg"])
    }
    return pub, nil
}, jwt.WithValidMethods([]string{"EdDSA"}), jwt.WithIssuer("pg-ai-stewards"))
// then check claims.Room matches the room the socket is joining.
```

`VerifyToken` in `token.go` is the **reference implementation** — ai-chattermax
mirrors it. EdDSA-only (reject any other alg), issuer-checked, expiry enforced by
the parser, and the `room` claim must match the room being joined.

## Build steps (status)

- **PS.1** ✅ skeleton + `persona_host` schema on boot
- **PS.2** ✅ Ed25519 signing key (generate-on-boot, persist) + `/pubkey`
- **PS.3** ✅ `MintToken` (EdDSA JWT) + issuance recording
- **PS.4** ✅ persona registry + seed `dm-assistant` / `npc-ally`
- **PS.5** ✅ `/join` handshake (mint + `persona_rooms`) + this contract
- **PS.6** ✅ security gate: `-verify-token` ops mode; live end-to-end HTTP
  (healthz → pubkey → personas → join → verify against the *fetched* pubkey);
  tampered token rejected; logs carry no private key / raw token / PEM markers

**#6 (AX2) is complete** — identity, minting, registry, and handshake all work
end-to-end. Next is **#7** (the turn loop), not part of #6.

## Not yet (converge later)

- **#7** the long-lived turn loop: per (persona, room), self-pace within the room
  ceiling and dispatch each turn to the substrate for cognition.
- Deployment: its own container next to the substrate; a scoped DB role for
  `persona_host` (not the superuser); ai-chattermax verifies via the pubkey env.
