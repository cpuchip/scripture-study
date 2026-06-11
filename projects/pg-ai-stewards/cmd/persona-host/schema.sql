-- persona_host schema — managed by the cmd/persona-host SIDECAR, NOT the core
-- pg-ai-stewards extension. A general substrate install never runs this binary,
-- so this schema only exists where the persona sidecar is deployed. All
-- statements are idempotent (safe to run on every boot).

CREATE SCHEMA IF NOT EXISTS persona_host;

-- One row per persona. agent_family names the backing substrate agent that
-- supplies cognition per turn; model_override/tools_override are NULL to inherit
-- that family, or set to specialize this persona.
CREATE TABLE IF NOT EXISTS persona_host.personas (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug          text UNIQUE NOT NULL,
    display_name  text NOT NULL,
    avatar_url    text,
    agent_family  text NOT NULL,
    persona_prompt text,
    model_override text,
    tools_override jsonb,
    pacing        jsonb NOT NULL DEFAULT '{}'::jsonb,
    status        text NOT NULL DEFAULT 'active',
    -- which substrate pipeline drives this persona's turns (model + tools):
    -- persona-turn (default, kimi), persona-turn-lmstudio, persona-turn-gemini, …
    pipeline      text NOT NULL DEFAULT 'persona-turn',
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);
-- Existing deployments: add the column if the table predates it.
ALTER TABLE persona_host.personas ADD COLUMN IF NOT EXISTS pipeline text NOT NULL DEFAULT 'persona-turn';

-- DH-2 promotion: whether this persona's NEW cast members get their own
-- substrate session by default (true for the Party — every PC is its own
-- mind; false for the DM — scene NPCs stay facets unless promoted).
ALTER TABLE persona_host.personas ADD COLUMN IF NOT EXISTS default_promote boolean NOT NULL DEFAULT false;

-- DH-2 promotion: a persona's named characters. A PROMOTED character has its
-- own substrate session — own memory, own LLM loop — room-agnostic (the mind
-- belongs to the character, not the room); display stays a platform
-- sub-persona. model_override is stored now, applied when per-character model
-- routing lands. The owning persona's connection still posts the lines.
CREATE TABLE IF NOT EXISTS persona_host.characters (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    persona_slug   text NOT NULL,
    name           text NOT NULL,
    session_id     text,
    model_override text,
    prompt         text,
    promoted       boolean NOT NULL DEFAULT false,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_characters_owner_name
    ON persona_host.characters (persona_slug, lower(name));

-- Which rooms a persona has joined (handshake state, PS.5).
CREATE TABLE IF NOT EXISTS persona_host.persona_rooms (
    persona_id   uuid NOT NULL REFERENCES persona_host.personas(id) ON DELETE CASCADE,
    room_id      text NOT NULL,
    joined_at    timestamptz NOT NULL DEFAULT now(),
    last_turn_at timestamptz,
    PRIMARY KEY (persona_id, room_id)
);

-- Audit trail of every minted token (PS.3). The token itself is never stored;
-- only its jti + scope + lifetime, so issuance can be reviewed and revoked.
CREATE TABLE IF NOT EXISTS persona_host.token_issuance (
    jti         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    persona_id  uuid NOT NULL REFERENCES persona_host.personas(id),
    room_id     text NOT NULL,
    issued_at   timestamptz NOT NULL DEFAULT now(),
    expires_at  timestamptz NOT NULL,
    revoked_at  timestamptz
);

-- The sidecar's Ed25519 signing keypair (PS.2). Exactly one row. The private
-- key is generated on first boot and NEVER logged, exported, or placed in any
-- model context — same handling class as the coder's GitHub token.
CREATE TABLE IF NOT EXISTS persona_host.signing_key (
    id              int PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    private_key_pem text NOT NULL,
    public_key_pem  text NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now()
);
