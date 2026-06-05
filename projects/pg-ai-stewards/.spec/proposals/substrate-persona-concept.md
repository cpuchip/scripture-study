# Substrate proposal — Persona Concept (ai-chattermax #6), as a Go SIDECAR

**Status:** ✅ BUILT + verified (2026-06-04). PS.1–PS.6 all shipped in `cmd/persona-host` (commits `fd13f56`→`f529242`, not pushed). Live e2e + log-leak verified. **#6 complete; next is #7 (turn loop), out of scope here.** *(Was: RATIFIED + re-architected, Michael 2026-06-04. Build lane: mine.)*
**Parent design:** `projects/ai-chattermax/.spec/proposals/chat-server-design.md`.
**Supersedes** the in-extension draft (p6-1 reverted). The substrate core stays **general** — no persona/chat/JWT code in the pgrx extension.

---

## The architecture call (revises ratified Q2)

Persona identity + credential minting + room handshake + turn-loop orchestration do **not** belong in the core substrate extension — that would couple a general-purpose substrate to one app most installs never run. They live in an **optional Go sidecar, `cmd/persona-host`**, exactly like the coder lives in `cmd/coder-mcp` (heavy/app-specific logic in a Go sidecar; only thin orchestration in extension SQL).

- **Substrate core (unchanged, general):** offers *cognition per turn* via its existing dispatch (`consult_subagent` / a work_item). It never knows "persona," "room," or "JWT."
- **`cmd/persona-host` sidecar (Go, optional):** the persona registry, EdDSA keypair + JWT minting, the room handshake, and (in #7) the turn loop. A general `pg-ai-stewards` install simply doesn't run this binary.
- **ai-chattermax (the room):** verifies persona tokens with the host's public key.

**One host serves MANY personas** (Michael, 2026-06-04) — not one process per persona. One parent credential, one signing key, N persona sub-tokens. The DM-assistant and NPC-ally are two registry rows in one `cmd/persona-host` process; a third persona is a row, not a deployment.

"pg-ai-stewards **hosts** personas" (Q2's spirit) still holds — the sidecar is a pg-ai-stewards-side service. What changed: **sidecar, not core extension.** Bonus: EdDSA+JWT in Go is stdlib (`crypto/ed25519` + `golang-jwt`), so this needs **no pgrx Rust crate and no extension rebuild.**

## State — `persona_host` schema (sidecar-managed)

Personas live in their **own `persona_host` schema** in the substrate's Postgres (Michael's choice), created + migrated **by the sidecar** (not the extension's migration set), so a core install never sees it.

```sql
CREATE SCHEMA IF NOT EXISTS persona_host;

CREATE TABLE IF NOT EXISTS persona_host.personas (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug text UNIQUE NOT NULL, display_name text NOT NULL, avatar_url text,
    agent_family text NOT NULL,            -- backing substrate agent (resolved at dispatch)
    persona_prompt text,
    model_override text, tools_override jsonb,   -- NULL = inherit agent_family
    pacing jsonb NOT NULL DEFAULT '{}'::jsonb,
    status text NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS persona_host.persona_rooms (
    persona_id uuid NOT NULL REFERENCES persona_host.personas(id) ON DELETE CASCADE,
    room_id text NOT NULL, joined_at timestamptz NOT NULL DEFAULT now(), last_turn_at timestamptz,
    PRIMARY KEY (persona_id, room_id)
);
CREATE TABLE IF NOT EXISTS persona_host.token_issuance (
    jti uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    persona_id uuid NOT NULL REFERENCES persona_host.personas(id),
    room_id text NOT NULL, issued_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL, revoked_at timestamptz
);
CREATE TABLE IF NOT EXISTS persona_host.signing_key (   -- one row; private key never logged/exported
    id int PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    private_key_pem text NOT NULL, public_key_pem text NOT NULL, created_at timestamptz NOT NULL DEFAULT now()
);
```

## The token contract (ai-chattermax verifies this — unchanged)

- **Format:** JWT, `alg: EdDSA` (Ed25519). **Claims:** `iss="pg-ai-stewards"`, `sub=<persona_id>`, `slug`, `name`, `avatar`, `room`, `iat`, `exp` (~15 min), `jti`.
- **Room-side verify:** EdDSA signature against the host's published public key (`STEWARDS_PERSONA_PUBKEY`), check `exp`/`iss`/`room`. No callback.
- The **private key never leaves the sidecar** (stored in `persona_host.signing_key`, never logged, never in any model context — same class as the coder's GitHub token).

## Build sub-steps (`cmd/persona-host`, gated commits, smoke each)

- **PS.1 — sidecar skeleton + `persona_host` schema migration.** A Go `cmd/persona-host` (HTTP + MCP-or-CLI surface), embedded SQL migration that creates the schema on boot, DB conn to the substrate Postgres.
- **PS.2 — signing key.** Generate the Ed25519 keypair on first boot if absent, persist to `persona_host.signing_key`; a `GET /pubkey` (or env export) so ai-chattermax can be configured.
- **PS.3 — `MintToken(personaSlug, roomID, ttl)`.** Go `golang-jwt` EdDSA sign; record `token_issuance`. Private key never logged.
- **PS.4 — persona registry + seed.** CRUD over `persona_host.personas`; seed `dm-assistant` + `npc-ally` referencing existing agent_families.
- **PS.5 — handshake/join.** `JoinRoom(persona, room)` → mint + upsert `persona_rooms`; document the connection contract for ai-chattermax (verify side).
- **PS.6 — smoke + security check.** Mint for `dm-assistant`; verify against the exported pubkey (inverse-hypothesis: a wrong-key signature must FAIL); grep to confirm the private key + raw token never hit logs.

## Convergence (#7, #12)
After #6 the sidecar can mint + a persona can authenticate into a room and appear. **#7** adds the long-lived turn loop: per (persona, room) connection, self-pace within the room ceiling, and on each turn **dispatch to the substrate** (`consult_subagent`/work_item) for the message — cost-tracked, context-engine'd, model/tools per the persona. **#12** seeds the D&D session.

## Open (pin at work-time)
- Sidecar surface: HTTP API vs an MCP server vs CLI (lean HTTP for ai-chattermax + a thin admin CLI).
- DB creds for the sidecar (a scoped role for `persona_host`, not the superuser).
- Where `cmd/persona-host` deploys (its own container next to the substrate; ai-chattermax verifies via the pubkey env).

---

## Cycle framing (book)
**Stewardship (Step 3) made literal for non-human agents, and kept at the right layer.** A persona is a scoped, credentialed identity minted from a parent authority, acting only as itself — and the *machinery* for that lives in an optional sidecar, not bolted onto the general tool. The architecture decision is itself the lesson: delegation means giving each concern its own stewardship boundary; app-specific power doesn't get to colonize the shared substrate. Feeds Part Two ch. 07 (delegation as stewardship).

*Re-architected 2026-06-04. Build: `cmd/persona-host` (me). ✅ BUILT 2026-06-04 — PS.1–PS.6 shipped + live-verified (commits `fd13f56`→`f529242`). Journal: `.spec/journal/2026-06-04-persona-host-ps1-ps6.md`.*
