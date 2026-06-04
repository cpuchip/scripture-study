# Substrate proposal — Persona Concept (ai-chattermax item #6)

**Status:** RATIFIED (Michael, 2026-06-04) — build-ready. The substrate-side half of the ai-chattermax split (the coder builds the website; I build this).
**Parent design:** `projects/ai-chattermax/.spec/proposals/chat-server-design.md` (Q2 ratified the substrate owns personas natively; Q5 the sub-token principle; Q4 self-pace-within-ceiling).
**Boundary:** this is item #6 — schema + handshake + sub-token minting + seed two personas. The turn loop (observe→decide→post) is item #7, separate.

---

## Binding question

What does pg-ai-stewards need so that a substrate-hosted persona can **authenticate into an ai-chattermax room and appear in the roster** — with its credential never in model context, its identity owned substrate-side — without yet building the turn loop?

## The four ratified decisions (2026-06-04)

1. **Runtime: long-lived process + work_item per turn.** When a persona joins a room, the substrate runs a persistent (bgworker-managed) WebSocket client that holds room state and self-paces; each *spoken* turn (item #7) dispatches a `work_item` for the LLM call — reusing the dispatch/cost/model machinery. **#6 must make the schema support this; #6 does not build the loop.**
2. **Token: signed token + room verifies with the substrate's public key.** The substrate signs a short-TTL persona token (EdDSA/Ed25519 JWT); ai-chattermax verifies the signature locally with the substrate's **public** key — no per-connect callback. Revocation via short TTL + refresh (+ an optional revocation list later). The **private signing key never enters model context** (same class as the coder's GitHub token).
3. **Model/tools: inherit from agent_family, per-persona override.** A persona defaults to its backing `agent_family`'s model + tools, with optional per-persona overrides (DM-assistant → stronger model + lore-lookup; NPC → cheap).
4. **Scope: schema + handshake + sub-token + seed the two D&D personas, NO turn loop.** Delivers a persona that can authenticate and show up; posting is #7.

## Schema

```sql
-- A persona: substrate-owned identity backed by an agent_family.
CREATE TABLE IF NOT EXISTS stewards.personas (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  slug          text UNIQUE NOT NULL,          -- 'dm-assistant', 'npc-ally'
  display_name  text NOT NULL,                 -- room-facing name
  avatar_url    text,                          -- room-facing avatar (optional)
  agent_family  text NOT NULL,                 -- backing agent definition
  persona_prompt text,                         -- system-prompt overlay on the family prompt
  model_override text,                         -- NULL = inherit family model; else pin
  tools_override jsonb,                        -- NULL = inherit family tools; else subset/superset
  pacing        jsonb NOT NULL DEFAULT '{}',   -- self-pace cfg: min_seconds_between_turns, quiet_period_budget…
  status        text NOT NULL DEFAULT 'active',-- active | disabled
  created_at    timestamptz NOT NULL DEFAULT now(),
  updated_at    timestamptz NOT NULL DEFAULT now()
);

-- Persona ↔ room membership (one persona may join many rooms).
CREATE TABLE IF NOT EXISTS stewards.persona_rooms (
  persona_id   uuid NOT NULL REFERENCES stewards.personas(id) ON DELETE CASCADE,
  room_id      text NOT NULL,                  -- ai-chattermax room identifier
  joined_at    timestamptz NOT NULL DEFAULT now(),
  last_turn_at timestamptz,
  PRIMARY KEY (persona_id, room_id)
);

-- Issuance audit (signed tokens are stateless to verify; this is for audit/revocation).
CREATE TABLE IF NOT EXISTS stewards.persona_token_issuance (
  jti         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  persona_id  uuid NOT NULL REFERENCES stewards.personas(id),
  room_id     text NOT NULL,
  issued_at   timestamptz NOT NULL DEFAULT now(),
  expires_at  timestamptz NOT NULL,
  revoked_at  timestamptz
);
```

The **signing keypair** lives in a substrate secret store (a config row or the existing secrets mechanism), generated once. The **private key is never rendered into any dispatch prompt** — minting is a SQL/Rust path the bgworker calls; the persona's LLM context never sees the key or the raw token.

## The token contract (the interface AX3's room consumes)

This is the seam between the two halves of the split — the coder's room (AX3) must implement the verify side, so the contract is fixed here:

- **Format:** JWT, `alg: EdDSA` (Ed25519).
- **Claims:** `iss="pg-ai-stewards"`, `sub=<persona_id>`, `slug`, `name`, `avatar`, `room=<room_id>`, `iat`, `exp` (TTL ~15 min), `jti`.
- **Verification (room side):** verify the EdDSA signature with the substrate's published public key (room config/env, e.g. `STEWARDS_PERSONA_PUBKEY`); check `exp`, `iss`, and that `room` matches the joined room. No callback to the substrate.
- **Refresh:** the persona-runtime re-mints before `exp` (handled in the runtime, #7-adjacent — for #6 the mint + a single valid token is enough).

## Build sub-steps (C–F cadence, gated commits, smoke before each)

- **P6.1 — schema.** The three tables above (idempotent SQL migration; live-apply via `docker cp + psql -f`).
- **P6.2 — signing key infra.** Generate/store the Ed25519 keypair substrate-side; a `persona_signing_pubkey()` export function (so the pubkey can be handed to ai-chattermax config). Private key never leaves SQL/Rust.
- **P6.3 — `mint_persona_token(persona_slug, room_id, ttl)`.** Builds the JWT, signs with the private key, records `persona_token_issuance`. Returns the compact token. (Rust pg_extern — needs the bump-extension hook; EdDSA via a Rust crate.)
- **P6.4 — seed the two D&D personas.** `dm-assistant` and `npc-ally` rows referencing existing agent_families (model overrides per taste; tools per the D&D MVP).
- **P6.5 — handshake surface.** `persona_join(persona_slug, room_id)` → mints a token + upserts `persona_rooms`. Document the connection contract (above) for AX3.
- **P6.6 — smoke + the security check.** Mint a token for `dm-assistant` in a test room; verify it validates against the exported pubkey with a standalone EdDSA check; **confirm the private key and the raw token never appear in any dispatch prompt** (grep the dispatch path). Inverse-hypothesis: a token signed with a wrong key must FAIL verification.

## Convergence with the coder's half (#7, #12)

After #6: the coder's room (AX3) can verify persona tokens and show personas in the roster. **#7** (the turn loop) wires the long-lived persona-runtime: hold the WS connection, self-pace within the room ceiling, and on each turn dispatch a `pipeline_family='persona_turn'` work_item (room context → message), cost-tracked, model/tools per the persona. **#12** seeds the actual D&D session. Those are separate ratified work-items.

## Open items pinned at work-time

- EdDSA Rust crate choice (e.g. `ed25519-dalek` / `jsonwebtoken` with EdDSA) + key storage location (config table vs. existing secret mechanism).
- `pacing` jsonb exact keys (defer detail to #7, but reserve the field now).
- `tools_override` shape (subset of agent_family tools vs. full re-spec) — lean: an allow-list of tool names, NULL = inherit.
- Revocation list: TTL-only for MVP; add a check against `revoked_at` only if needed.

---

## Cycle framing (for the book audit)

This is **Step 3 (Stewardship) made literal for non-human agents**: a persona is a scoped, credentialed identity minted from a parent authority, able to act only as itself, in its rooms, with a credential it never sees in full. The "sub-token minted from one parent key, never in model context" is the delegation pattern (the coder's GitHub token) generalized — *the steward holds the keys; the agent is handed only what its role requires.* Feeds the delegation chapter (Part Two ch. 07).

*Ratified + spec written 2026-06-04. My build lane (AX2). The coder builds the room that consumes this contract (AX3).*
