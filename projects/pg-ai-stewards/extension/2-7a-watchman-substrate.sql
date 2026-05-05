-- Phase 2.7a — Watchman substrate (no model dispatch required).
--
-- Live-DB migration. Same pattern as 2-6a/b/c.
--
-- Adds:
--   - stewards.studies.last_consolidated_at column
--     (the existing `updated_at` column already serves as
--     `last_touched_at` because the touch_study() trigger only
--     bumps it on semantic changes to title/body/frontmatter.
--     No new column needed for that side.)
--   - stewards.verdicts table — one row per consolidation pass.
--   - stewards.findings table — drift recommendations + REM
--     synthesis candidates with surface-once-and-stop discipline
--     via acknowledged_at.
--   - stewards.dirty_queue view — docs where updated_at >
--     coalesce(last_consolidated_at, '-infinity'), oldest first.
--   - stewards.record_verdict() — writes a verdict row AND bumps
--     last_consolidated_at in one transaction (single-write rule).
--   - stewards.record_finding() — writes a finding row.
--   - stewards.acknowledge_finding() — marks a finding acknowledged.
--   - stewards.study_history() — verdict + finding timeline for a
--     single doc.
--
-- The schema enforces the anti-loop discipline directly:
--   1. Terminal verdicts never re-enter the queue (the dirty-bit
--      stays satisfied until the doc is touched again).
--   2. Findings have acknowledged_at — the surface-once-and-stop
--      rule is a CHECK on the queue ("don't re-surface findings
--      that already have an open finding row").
--   3. Per-pass token budget is enforced by the bgworker (2.7b),
--      not the schema. The schema only records what was spent.
--
-- This file ships the substrate. The bgworker that automates it
-- is Phase 2.7b — needs Phase 3's model dispatch sidecar to exist.

BEGIN;

-- ============================================================
-- studies.last_consolidated_at
-- ============================================================
ALTER TABLE stewards.studies
    ADD COLUMN IF NOT EXISTS last_consolidated_at timestamptz;

CREATE INDEX IF NOT EXISTS studies_dirty_idx
    ON stewards.studies (updated_at)
    WHERE last_consolidated_at IS NULL
       OR updated_at > last_consolidated_at;

-- ============================================================
-- stewards.verdicts — one row per consolidation pass
-- ============================================================
-- Verdict values:
--   clean      — doc still aligns with current code/spec; no action
--   drift      — doc has drifted; finding row should be written
--   done       — doc represents completed work; archive candidate
--   superseded — doc replaced by another; archive candidate
--   skipped    — pass aborted (token budget, model error, etc.)
--
-- Status terminal-or-not:
--   clean and skipped are NON-terminal (doc may need re-evaluation
--   when touched again). done and superseded are TERMINAL — the
--   doc never re-enters the queue without explicit touch.
--   drift sits in between: surface a finding, don't re-evaluate
--   until the finding is acknowledged or the doc is re-touched.
CREATE TABLE IF NOT EXISTS stewards.verdicts (
    id              bigserial PRIMARY KEY,
    study_id        text NOT NULL
                    REFERENCES stewards.studies(id) ON DELETE CASCADE,
    verdict         text NOT NULL
                    CHECK (verdict IN ('clean', 'drift', 'done',
                                        'superseded', 'skipped')),
    -- Sources for verdict values: the design list above.
    reasoning       text NOT NULL DEFAULT '',
    model           text,           -- e.g. 'claude-haiku-4', 'kimi-k2.6', NULL for human-recorded
    tokens_in       int NOT NULL DEFAULT 0,
    tokens_out      int NOT NULL DEFAULT 0,
    pass_id         text,           -- groups verdicts in one pass run
    actor           text NOT NULL DEFAULT 'system',
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS verdicts_study_idx
    ON stewards.verdicts (study_id, created_at DESC);
CREATE INDEX IF NOT EXISTS verdicts_pass_idx
    ON stewards.verdicts (pass_id, created_at);
CREATE INDEX IF NOT EXISTS verdicts_verdict_idx
    ON stewards.verdicts (verdict);

-- ============================================================
-- stewards.findings — drift recommendations + REM synthesis
-- ============================================================
-- kind:
--   drift      — written from a drift verdict; tells the human
--                "this doc no longer matches reality, here's how"
--   synthesis  — REM-pass output; candidate insight connecting
--                multiple docs; always reviewed before promotion
--
-- acknowledged_at NULL = open (will not be re-surfaced).
-- The surface-once-and-stop rule lives in find_open_findings
-- (the agent only proposes new findings for docs without an
-- open same-kind finding).
CREATE TABLE IF NOT EXISTS stewards.findings (
    id              bigserial PRIMARY KEY,
    study_id        text
                    REFERENCES stewards.studies(id) ON DELETE CASCADE,
    -- study_id is nullable for synthesis findings that span
    -- multiple docs (related_study_ids carries the full set).
    related_study_ids text[] NOT NULL DEFAULT ARRAY[]::text[],
    kind            text NOT NULL CHECK (kind IN ('drift', 'synthesis')),
    severity        text NOT NULL DEFAULT 'medium'
                    CHECK (severity IN ('low', 'medium', 'high')),
    message         text NOT NULL,
    suggested_action text,
    pass_id         text,
    actor           text NOT NULL DEFAULT 'system',
    created_at      timestamptz NOT NULL DEFAULT now(),
    acknowledged_at timestamptz,
    acknowledged_by text,
    resolution      text         -- 'acted', 'dismissed', 'deferred'
);

CREATE INDEX IF NOT EXISTS findings_study_idx
    ON stewards.findings (study_id, created_at DESC);
CREATE INDEX IF NOT EXISTS findings_open_idx
    ON stewards.findings (kind, severity, created_at)
    WHERE acknowledged_at IS NULL;

-- ============================================================
-- View: stewards.dirty_queue
-- Docs that need (re-)consolidation, oldest-touched first.
-- A doc is dirty iff it has been touched since last consolidated,
-- AND has no open drift finding (surface-once-and-stop).
-- ============================================================
CREATE OR REPLACE VIEW stewards.dirty_queue AS
SELECT s.id,
       s.slug,
       s.kind,
       s.title,
       s.updated_at,
       s.last_consolidated_at,
       (s.updated_at - coalesce(s.last_consolidated_at,
                                 'epoch'::timestamptz)) AS dirty_for
  FROM stewards.studies s
 WHERE (s.last_consolidated_at IS NULL
        OR s.updated_at > s.last_consolidated_at)
   AND NOT EXISTS (
       SELECT 1 FROM stewards.findings f
        WHERE f.study_id = s.id
          AND f.kind = 'drift'
          AND f.acknowledged_at IS NULL
   )
 ORDER BY coalesce(s.last_consolidated_at, 'epoch'::timestamptz),
          s.updated_at;

-- ============================================================
-- Function: record_verdict()
-- Writes a verdict row AND bumps last_consolidated_at in one
-- transaction. Single-write rule (same pattern as create_todo /
-- complete_todo from 2.6b).
-- ============================================================
CREATE OR REPLACE FUNCTION stewards.record_verdict(
    p_slug       text,
    p_verdict    text,
    p_reasoning  text DEFAULT '',
    p_model      text DEFAULT NULL,
    p_tokens_in  int  DEFAULT 0,
    p_tokens_out int  DEFAULT 0,
    p_pass_id    text DEFAULT NULL,
    p_actor      text DEFAULT 'system'
) RETURNS bigint
LANGUAGE plpgsql AS $func$
DECLARE
    v_study_id text;
    v_id       bigint;
BEGIN
    SELECT s.id INTO v_study_id
      FROM stewards.studies s
     WHERE s.slug = p_slug;
    IF v_study_id IS NULL THEN
        RAISE EXCEPTION 'record_verdict: no study with slug %', p_slug;
    END IF;

    INSERT INTO stewards.verdicts
        (study_id, verdict, reasoning, model, tokens_in, tokens_out,
         pass_id, actor)
    VALUES
        (v_study_id, p_verdict, p_reasoning, p_model, p_tokens_in,
         p_tokens_out, p_pass_id, p_actor)
    RETURNING id INTO v_id;

    -- Bump last_consolidated_at. Use a direct UPDATE that does NOT
    -- bump updated_at (which would re-dirty the doc immediately).
    -- The touch_study() trigger only bumps updated_at on
    -- title/body/frontmatter changes, so this UPDATE is safe.
    UPDATE stewards.studies
       SET last_consolidated_at = now()
     WHERE id = v_study_id;

    RETURN v_id;
END;
$func$;

-- ============================================================
-- Function: record_finding()
-- ============================================================
CREATE OR REPLACE FUNCTION stewards.record_finding(
    p_slug             text,
    p_kind             text,
    p_message          text,
    p_severity         text   DEFAULT 'medium',
    p_suggested_action text   DEFAULT NULL,
    p_related_slugs    text[] DEFAULT ARRAY[]::text[],
    p_pass_id          text   DEFAULT NULL,
    p_actor            text   DEFAULT 'system'
) RETURNS bigint
LANGUAGE plpgsql AS $func$
DECLARE
    v_study_id     text;
    v_related_ids  text[];
    v_id           bigint;
BEGIN
    SELECT s.id INTO v_study_id
      FROM stewards.studies s
     WHERE s.slug = p_slug;
    -- Note: study_id may be NULL for synthesis findings that span
    -- only related studies. We allow that.

    SELECT array_agg(s.id) INTO v_related_ids
      FROM stewards.studies s
     WHERE s.slug = ANY(p_related_slugs);

    INSERT INTO stewards.findings
        (study_id, related_study_ids, kind, severity, message,
         suggested_action, pass_id, actor)
    VALUES
        (v_study_id, coalesce(v_related_ids, ARRAY[]::text[]),
         p_kind, p_severity, p_message, p_suggested_action,
         p_pass_id, p_actor)
    RETURNING id INTO v_id;

    RETURN v_id;
END;
$func$;

-- ============================================================
-- Function: acknowledge_finding()
-- Marks an open finding acknowledged. Resolutions:
--   'acted'     — human took the suggested action
--   'dismissed' — human disagrees with the finding
--   'deferred'  — valid but not acting now (still leaves queue)
-- ============================================================
CREATE OR REPLACE FUNCTION stewards.acknowledge_finding(
    p_finding_id bigint,
    p_resolution text DEFAULT 'acted',
    p_actor      text DEFAULT 'system'
) RETURNS void
LANGUAGE plpgsql AS $func$
BEGIN
    IF p_resolution NOT IN ('acted', 'dismissed', 'deferred') THEN
        RAISE EXCEPTION 'acknowledge_finding: invalid resolution %',
              p_resolution;
    END IF;

    UPDATE stewards.findings
       SET acknowledged_at = now(),
           acknowledged_by = p_actor,
           resolution      = p_resolution
     WHERE id = p_finding_id
       AND acknowledged_at IS NULL;

    IF NOT FOUND THEN
        RAISE EXCEPTION
            'acknowledge_finding: finding % not found or already acknowledged',
            p_finding_id;
    END IF;
END;
$func$;

-- ============================================================
-- Function: study_history()
-- Returns verdict + finding timeline for a single doc, newest first.
-- ============================================================
CREATE OR REPLACE FUNCTION stewards.study_history(p_slug text)
RETURNS TABLE (
    event_at    timestamptz,
    event_type  text,
    detail      text,
    actor       text,
    extra       jsonb
)
LANGUAGE sql STABLE AS $func$
    WITH s AS (
        SELECT id FROM stewards.studies WHERE slug = p_slug
    )
    SELECT v.created_at,
           ('verdict:' || v.verdict)::text,
           v.reasoning,
           v.actor,
           jsonb_build_object(
               'model',      v.model,
               'tokens_in',  v.tokens_in,
               'tokens_out', v.tokens_out,
               'pass_id',    v.pass_id
           )
      FROM stewards.verdicts v
      JOIN s ON s.id = v.study_id
    UNION ALL
    SELECT f.created_at,
           ('finding:' || f.kind || '/' || f.severity)::text,
           f.message,
           f.actor,
           jsonb_build_object(
               'suggested_action', f.suggested_action,
               'acknowledged_at',  f.acknowledged_at,
               'resolution',       f.resolution,
               'pass_id',          f.pass_id,
               'related',          f.related_study_ids
           )
      FROM stewards.findings f
      JOIN s ON s.id = f.study_id
    ORDER BY 1 DESC;
$func$;

COMMIT;
