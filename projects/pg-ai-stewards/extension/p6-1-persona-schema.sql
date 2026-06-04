-- =====================================================================
-- p6-1 (2026-06-04) — Persona concept #6, sub-step P6.1: the schema.
-- Substrate-owns-personas (chat-server-design Q2). A persona is a
-- substrate-owned identity backed by an agent_family that joins ai-chattermax
-- rooms. This file is schema only — handshake/sub-token (P6.2-P6.5) follow.
-- Idempotent; SQL-only (no rebuild). See .spec/proposals/substrate-persona-concept.md.
-- =====================================================================

-- A persona: substrate-owned identity backed by an agent_family.
CREATE TABLE IF NOT EXISTS stewards.personas (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug           text UNIQUE NOT NULL,           -- 'dm-assistant', 'npc-ally'
    display_name   text NOT NULL,                  -- room-facing name
    avatar_url     text,                           -- room-facing avatar (optional)
    agent_family   text NOT NULL,                  -- backing agent definition
    persona_prompt text,                           -- system-prompt overlay on the family prompt
    model_override text,                           -- NULL = inherit family model; else pin
    tools_override jsonb,                           -- NULL = inherit family tools; else allow-list
    pacing         jsonb NOT NULL DEFAULT '{}'::jsonb,  -- self-pace cfg (P6/#7): min_seconds_between_turns, quiet_period_budget…
    status         text NOT NULL DEFAULT 'active', -- active | disabled
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
);

-- Persona <-> room membership (one persona may join many rooms).
CREATE TABLE IF NOT EXISTS stewards.persona_rooms (
    persona_id   uuid NOT NULL REFERENCES stewards.personas(id) ON DELETE CASCADE,
    room_id      text NOT NULL,                    -- ai-chattermax room identifier
    joined_at    timestamptz NOT NULL DEFAULT now(),
    last_turn_at timestamptz,
    PRIMARY KEY (persona_id, room_id)
);

-- Token issuance audit (signed tokens are stateless to verify; this is for
-- audit + optional revocation — see P6.3 mint_persona_token).
CREATE TABLE IF NOT EXISTS stewards.persona_token_issuance (
    jti         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    persona_id  uuid NOT NULL REFERENCES stewards.personas(id),
    room_id     text NOT NULL,
    issued_at   timestamptz NOT NULL DEFAULT now(),
    expires_at  timestamptz NOT NULL,
    revoked_at  timestamptz
);

CREATE INDEX IF NOT EXISTS idx_persona_token_issuance_persona
    ON stewards.persona_token_issuance (persona_id, issued_at DESC);

COMMENT ON TABLE stewards.personas IS
'Persona concept #6 (P6.1): substrate-owned chat identity backed by an agent_family. ai-chattermax references these via the signed-token handshake (P6.3-P6.5). Turn loop is #7.';
