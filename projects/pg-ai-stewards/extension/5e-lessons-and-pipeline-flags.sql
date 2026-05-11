-- =====================================================================
-- Phase 5e (Phase D.1) — Atonement + Sabbath substrate
--
-- Three pieces:
--   1. pipelines.sabbath_enabled + atonement_enabled — opt-in flags
--      per pipeline_family (D-D1 ratified opt-out: study/lesson/talk
--      default ON; debug/dev default OFF; D-D2 ratified atonement
--      opt-in initially).
--   2. stewards.lessons — append-only audit ledger mirroring
--      gate_decisions shape (per 2026-05-11 ratification: kind column,
--      single content per row, ratification fields).
--   3. sessions_kind_check extended to add 'sabbath' (mirrors 5c's
--      pattern adding 'gate').
-- =====================================================================

-- ---------------------------------------------------------------------
-- Section 1: pipelines flags
-- ---------------------------------------------------------------------

ALTER TABLE stewards.pipelines
    ADD COLUMN IF NOT EXISTS sabbath_enabled   boolean NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS atonement_enabled boolean NOT NULL DEFAULT false;

COMMENT ON COLUMN stewards.pipelines.sabbath_enabled IS
'Phase 5e (D.1): Sabbath dispatch fires when work_item reaches verified maturity. Per D-D1: study/lesson/talk default ON; new pipelines default OFF until opted in.';
COMMENT ON COLUMN stewards.pipelines.atonement_enabled IS
'Phase 5e (D.1): Atonement dispatch fires when work_item is quarantined. Per D-D2: opt-in initially.';

-- Seed: study/lesson/talk = sabbath ON (per D-D1)
UPDATE stewards.pipelines SET sabbath_enabled = true
 WHERE family IN ('study-write', 'study-write-qwen', 'lesson', 'talk');

-- Atonement opt-in: nothing flipped on by default; humans turn on per pipeline.

-- ---------------------------------------------------------------------
-- Section 2: stewards.lessons — audit ledger (mirrors gate_decisions shape)
-- ---------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS stewards.lessons (
    id              bigserial PRIMARY KEY,
    work_item_id    uuid REFERENCES stewards.work_items(id) ON DELETE CASCADE,
    at              timestamptz NOT NULL DEFAULT now(),
    kind            text NOT NULL CHECK (kind IN
                        ('principle', 'decision', 'lesson', 'sabbath_reflection')),
    content         text NOT NULL,
    raw_response    jsonb,
    ratified_at     timestamptz,
    ratified_by     text,
    promoted_to     text,    -- '.mind/principles.md' | '.mind/decisions.md' | NULL
    work_id         bigint
);

CREATE INDEX IF NOT EXISTS lessons_at         ON stewards.lessons (at);
CREATE INDEX IF NOT EXISTS lessons_work_item  ON stewards.lessons (work_item_id);
CREATE INDEX IF NOT EXISTS lessons_unratified ON stewards.lessons (ratified_at) WHERE ratified_at IS NULL;
CREATE INDEX IF NOT EXISTS lessons_kind       ON stewards.lessons (kind);

COMMENT ON TABLE stewards.lessons IS
'Phase 5e (D.1): append-only ledger of lessons produced by Atonement (kind in principle|decision|lesson) and reflections produced by Sabbath (kind=sabbath_reflection). All rows land unratified; humans curate via Stewards-UI before promotion to .mind/ files (D-D3).';

-- Aggregation view consumed by Phase E's retry composer
CREATE OR REPLACE VIEW stewards.lessons_recent_ratified AS
SELECT l.*, wi.pipeline_family, wi.current_stage
  FROM stewards.lessons l
  JOIN stewards.work_items wi ON wi.id = l.work_item_id
 WHERE l.ratified_at IS NOT NULL
   AND l.kind IN ('lesson', 'principle')
 ORDER BY l.at DESC;

COMMENT ON VIEW stewards.lessons_recent_ratified IS
'Phase 5e (D.1): keyed by pipeline_family + current_stage. Phase E retry composer pulls the last 3 per (pipeline, stage) into retry context.';

-- ---------------------------------------------------------------------
-- Section 3: sessions_kind_check — add 'sabbath' (and 'atonement' for symmetry)
-- ---------------------------------------------------------------------

ALTER TABLE stewards.sessions DROP CONSTRAINT IF EXISTS sessions_kind_check;

ALTER TABLE stewards.sessions
    ADD CONSTRAINT sessions_kind_check
    CHECK (kind = ANY (ARRAY['chat','agent','tool','study','dev','gate','sabbath','atonement']));

-- ---------------------------------------------------------------------
-- Section 4: work_items.sabbath_completed_at — set by apply_sabbath_result;
--            promote_to_study gate checks this in D.5
-- ---------------------------------------------------------------------

ALTER TABLE stewards.work_items
    ADD COLUMN IF NOT EXISTS sabbath_completed_at timestamptz;

COMMENT ON COLUMN stewards.work_items.sabbath_completed_at IS
'Phase 5e (D.1): timestamp Sabbath reflection landed for this work_item. promote_to_study (D.5) refuses if NULL on a sabbath_enabled pipeline.';
