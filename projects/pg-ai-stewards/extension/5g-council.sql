-- =====================================================================
-- Phase 5g (Phase F.1) — Council schema (Zion / cycle step 11)
--
-- Three tables + a few indexes:
--   stewards.councils         — one row per convened council
--   stewards.council_members  — proposer/critic/synthesizer per council
--   stewards.resolutions      — bishop's resolution canonical record
--
-- Plus:
--   - sessions_kind_check extended to add 'council'
--   - one_active_council partial unique index enforces D-F1
--     (1 concurrent council initially)
--
-- Per ratifications:
--   D-F1: 1 concurrent council
--   D-F2: master-tier agents may bishop low-stakes councils
--         (defined as intent.scripture_anchor IS NULL AND
--          values_hierarchy lacks doctrinal/spiritual/discernment;
--          bishop_eligible function lands in F.5)
--   D-F3: all three destinations (resolutions canonical + promote to
--         study OR decisions.md by question type)
--   D-F4: manual + system-suggested convening (watchman flags;
--         human convenes)
-- =====================================================================

-- ---------------------------------------------------------------------
-- (1) sessions_kind_check — add 'council'
-- ---------------------------------------------------------------------

ALTER TABLE stewards.sessions DROP CONSTRAINT IF EXISTS sessions_kind_check;
ALTER TABLE stewards.sessions
    ADD CONSTRAINT sessions_kind_check
    CHECK (kind = ANY (ARRAY['chat','agent','tool','study','dev','gate','sabbath','atonement','council']));

-- ---------------------------------------------------------------------
-- (2) stewards.councils
-- ---------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS stewards.councils (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    intent_id         uuid NOT NULL REFERENCES stewards.intents(id),
    binding_question  text NOT NULL,
    convened_at       timestamptz NOT NULL DEFAULT now(),
    convened_by       text NOT NULL,           -- human name or 'watchman' for system-suggested
    bishop            text NOT NULL,           -- 'human:michael' | 'agent:<family>:<pipeline>:master'
    status            text NOT NULL DEFAULT 'deliberating'
                       CHECK (status IN ('deliberating', 'synthesizing', 'awaiting_bishop',
                                          'resolved', 'dissolved')),
    resolution_id     uuid,                    -- FK declared after resolutions exists
    dissolved_reason  text,
    resolved_at       timestamptz
);

CREATE INDEX IF NOT EXISTS councils_status      ON stewards.councils (status);
CREATE INDEX IF NOT EXISTS councils_convened_at ON stewards.councils (convened_at);

-- D-F1: at most ONE active council (deliberating | synthesizing | awaiting_bishop)
DO $idx$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes WHERE indexname = 'one_active_council'
    ) THEN
        CREATE UNIQUE INDEX one_active_council
            ON stewards.councils ((1))
            WHERE status IN ('deliberating', 'synthesizing', 'awaiting_bishop');
    END IF;
END;
$idx$;

COMMENT ON TABLE stewards.councils IS
'Phase 5g (F.1): one row per convened council. one_active_council partial unique index enforces D-F1 (1 concurrent council initially; lift after a month if real demand).';

-- ---------------------------------------------------------------------
-- (3) stewards.council_members
-- ---------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS stewards.council_members (
    council_id    uuid NOT NULL REFERENCES stewards.councils(id) ON DELETE CASCADE,
    agent_family  text NOT NULL,
    role          text NOT NULL CHECK (role IN ('proposer', 'critic', 'synthesizer')),
    work_id       bigint,             -- the work_queue id of this member's dispatch
    response      text,               -- assistant content when complete
    completed_at  timestamptz,
    PRIMARY KEY (council_id, agent_family, role)
);

CREATE INDEX IF NOT EXISTS council_members_council ON stewards.council_members (council_id);

COMMENT ON TABLE stewards.council_members IS
'Phase 5g (F.1): per-(council, agent_family, role) member. Synthesizer can be the same agent_family as a proposer; role disambiguates. Member key per 2026-05-11 ratification = (council_id, agent_family, role) — model floats per dispatch (steward chooses).';

-- ---------------------------------------------------------------------
-- (4) stewards.resolutions
-- ---------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS stewards.resolutions (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    council_id      uuid REFERENCES stewards.councils(id),
    resolved_at     timestamptz NOT NULL DEFAULT now(),
    resolved_by     text NOT NULL,             -- human name or agent identifier
    text            text NOT NULL,             -- the resolution itself
    promoted_to     text,                      -- 'study/<slug>.md' | '.mind/decisions.md' | NULL
    promoted_at     timestamptz,
    raw_proposal    jsonb                      -- the synthesizer's draft before bishop edits
);

CREATE INDEX IF NOT EXISTS resolutions_council ON stewards.resolutions (council_id);

COMMENT ON TABLE stewards.resolutions IS
'Phase 5g (F.1): canonical resolutions (D-F3). Bishop accept may also promote to study/ or .mind/decisions.md based on question type — file-write is a pending-write pattern (D-F follow-up).';

-- Now wire the FK back from councils.resolution_id → resolutions.id
DO $fk$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
         WHERE conname = 'councils_resolution_id_fkey'
    ) THEN
        ALTER TABLE stewards.councils
            ADD CONSTRAINT councils_resolution_id_fkey
            FOREIGN KEY (resolution_id) REFERENCES stewards.resolutions(id);
    END IF;
END;
$fk$;
