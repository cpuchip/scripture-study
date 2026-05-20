-- =====================================================================
-- Batch PE.B.1 — scheduled_pipelines schema + plpgsql cron_next_after
--
-- D-PE3 (ratified 2026-05-19): no hard frequency floor. Cost-cap +
-- bucket caps + quarantine are the safety net.
-- D-PE4: fire one missed run on recovery. missed_window_hours threshold
-- on each row (default 24h); past that, advance without firing.
-- D-PE6: standard 5-field cron with ranges + lists + step values.
--
-- Implementation note (2026-05-19): the cron parser is plpgsql instead
-- of a Rust cron crate. The crate approach would have required a pg
-- rebuild + replay of 86 post-G SQL files; deferred to a dedicated
-- rebuild session. cron_next_after is called once per dispatch (not per
-- tick), so plpgsql performance is fine.
-- =====================================================================

-- ---------------------------------------------------------------------
-- PE.B.1.a — scheduled_pipelines table
-- ---------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS stewards.scheduled_pipelines (
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug                 text UNIQUE NOT NULL,
    pipeline_family      text NOT NULL REFERENCES stewards.pipelines(family) ON DELETE RESTRICT,
    intent_id            uuid NOT NULL REFERENCES stewards.intents(id) ON DELETE RESTRICT,
    cron_pattern         text NOT NULL,
    input_template       jsonb NOT NULL,
    enabled              boolean NOT NULL DEFAULT true,
    missed_window_hours  int    NOT NULL DEFAULT 24,
    last_dispatched_at   timestamptz,
    next_due_at          timestamptz,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    notes                text,
    CONSTRAINT scheduled_pipelines_slug_check CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$')
);

CREATE INDEX IF NOT EXISTS scheduled_pipelines_due_idx
    ON stewards.scheduled_pipelines (next_due_at)
    WHERE enabled = true;

COMMENT ON TABLE stewards.scheduled_pipelines IS
'PE-B: cron-style scheduling for pipeline dispatches. Each row dispatches a new work_item of pipeline_family with input_template each time next_due_at is reached. The scheduled_pipelines_fire() function (called from the 60s watchman tick) scans this table.';

COMMENT ON COLUMN stewards.scheduled_pipelines.cron_pattern IS
'Standard 5-field cron (minute hour day-of-month month day-of-week). Supports literal, *, ranges (1-5), lists (1,3,5), and step values (*/15). Per D-PE6.';

COMMENT ON COLUMN stewards.scheduled_pipelines.missed_window_hours IS
'Per D-PE4: if next_due_at is in the past by less than this many hours, fire one missed run on recovery. Past that, skip the missed runs and advance next_due_at to the next future match. Default 24h.';

COMMENT ON COLUMN stewards.scheduled_pipelines.next_due_at IS
'Materialized by cron_next_after() trigger when cron_pattern is INSERT/UPDATEd, and recomputed by scheduled_pipelines_fire() after each dispatch.';

-- ---------------------------------------------------------------------
-- PE.B.1.b — cron field parser
--
-- Returns the set of valid integers for one cron field. Supports:
--   *                — every value in [p_lo, p_hi]
--   N                — literal integer
--   N-M              — range
--   N,M,...          — comma-separated list of any of the above
--   */N or */N-M     — step value (every Nth in range)
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.cron_field_values(
    p_field text,
    p_lo    int,
    p_hi    int
) RETURNS SETOF int
LANGUAGE plpgsql IMMUTABLE AS $func$
DECLARE
    v_part    text;
    v_step    int;
    v_range   text;
    v_lo      int;
    v_hi      int;
    v_dash    int;
    v_n       int;
BEGIN
    FOR v_part IN
        SELECT trim(t) FROM unnest(string_to_array(p_field, ',')) AS t
    LOOP
        -- Step value: <range>/<n>
        IF v_part ~ '/' THEN
            v_step  := split_part(v_part, '/', 2)::int;
            v_range := split_part(v_part, '/', 1);
            IF v_step <= 0 THEN
                RAISE EXCEPTION 'cron_field_values: step must be > 0 in %', v_part;
            END IF;
        ELSE
            v_step  := 1;
            v_range := v_part;
        END IF;

        -- Resolve range bounds
        IF v_range = '*' THEN
            v_lo := p_lo;
            v_hi := p_hi;
        ELSIF v_range ~ '^[0-9]+-[0-9]+$' THEN
            v_dash := position('-' IN v_range);
            v_lo := substring(v_range FROM 1 FOR v_dash - 1)::int;
            v_hi := substring(v_range FROM v_dash + 1)::int;
        ELSIF v_range ~ '^[0-9]+$' THEN
            v_lo := v_range::int;
            v_hi := v_lo;
        ELSE
            RAISE EXCEPTION 'cron_field_values: unparseable part % (in %)', v_part, p_field;
        END IF;

        IF v_lo < p_lo OR v_hi > p_hi OR v_lo > v_hi THEN
            RAISE EXCEPTION 'cron_field_values: out-of-range [%-%] (allowed [%-%]) in %',
                v_lo, v_hi, p_lo, p_hi, p_field;
        END IF;

        -- Emit values
        v_n := v_lo;
        WHILE v_n <= v_hi LOOP
            RETURN NEXT v_n;
            v_n := v_n + v_step;
        END LOOP;
    END LOOP;
END;
$func$;

-- ---------------------------------------------------------------------
-- PE.B.1.c — cron_next_after(pattern, after) → timestamptz
--
-- Brute-force minute-by-minute search forward from p_after for the next
-- timestamp matching the 5-field cron pattern. Bounded by a hard
-- 366-day horizon to keep pathological patterns from spinning.
--
-- Treats the cron pattern in UTC. dow_field=0..6 (0=Sun..6=Sat) matches
-- PostgreSQL EXTRACT(DOW). day-of-month and day-of-week use OR semantics
-- when either field is restricted (standard cron behavior).
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.cron_next_after(
    p_pattern text,
    p_after   timestamptz
) RETURNS timestamptz
LANGUAGE plpgsql IMMUTABLE AS $func$
DECLARE
    v_parts    text[];
    v_minute   text;
    v_hour     text;
    v_dom      text;
    v_month    text;
    v_dow      text;
    v_t        timestamptz;
    v_horizon  timestamptz;
    v_t_utc    timestamp;
    v_m        int;
    v_h        int;
    v_d        int;
    v_mo       int;
    v_w        int;
    v_dom_unrestricted boolean;
    v_dow_unrestricted boolean;
    v_minute_ok boolean;
    v_hour_ok   boolean;
    v_month_ok  boolean;
    v_dom_ok    boolean;
    v_dow_ok    boolean;
BEGIN
    v_parts := regexp_split_to_array(trim(p_pattern), '\s+');
    IF array_length(v_parts, 1) <> 5 THEN
        RAISE EXCEPTION 'cron_next_after: expected 5-field cron, got %', p_pattern;
    END IF;

    v_minute := v_parts[1];
    v_hour   := v_parts[2];
    v_dom    := v_parts[3];
    v_month  := v_parts[4];
    v_dow    := v_parts[5];

    -- Standard cron semantics: when both dom and dow are restricted
    -- (not *), match if EITHER fires. When one is *, only the other
    -- gates. Implemented by tracking which fields are unrestricted.
    v_dom_unrestricted := (trim(v_dom) = '*');
    v_dow_unrestricted := (trim(v_dow) = '*');

    -- Start at the next minute boundary AFTER p_after (cron fires AT
    -- the minute mark, not in between).
    v_t := date_trunc('minute', p_after) + interval '1 minute';
    v_horizon := p_after + interval '366 days';

    WHILE v_t <= v_horizon LOOP
        v_t_utc := v_t AT TIME ZONE 'UTC';

        v_m  := EXTRACT(MINUTE FROM v_t_utc)::int;
        v_h  := EXTRACT(HOUR   FROM v_t_utc)::int;
        v_d  := EXTRACT(DAY    FROM v_t_utc)::int;
        v_mo := EXTRACT(MONTH  FROM v_t_utc)::int;
        v_w  := EXTRACT(DOW    FROM v_t_utc)::int;

        -- Cheap gates first (minute/hour) to skip-ahead quickly
        v_minute_ok := EXISTS (
            SELECT 1 FROM stewards.cron_field_values(v_minute, 0, 59) WHERE cron_field_values = v_m
        );
        IF NOT v_minute_ok THEN
            v_t := v_t + interval '1 minute';
            CONTINUE;
        END IF;

        v_hour_ok := EXISTS (
            SELECT 1 FROM stewards.cron_field_values(v_hour, 0, 23) WHERE cron_field_values = v_h
        );
        IF NOT v_hour_ok THEN
            v_t := v_t + interval '1 minute';
            CONTINUE;
        END IF;

        v_month_ok := EXISTS (
            SELECT 1 FROM stewards.cron_field_values(v_month, 1, 12) WHERE cron_field_values = v_mo
        );
        IF NOT v_month_ok THEN
            v_t := v_t + interval '1 minute';
            CONTINUE;
        END IF;

        -- Day-of-month + day-of-week OR-semantic
        v_dom_ok := EXISTS (
            SELECT 1 FROM stewards.cron_field_values(v_dom, 1, 31) WHERE cron_field_values = v_d
        );
        v_dow_ok := EXISTS (
            SELECT 1 FROM stewards.cron_field_values(v_dow, 0, 6) WHERE cron_field_values = v_w
        );

        IF v_dom_unrestricted AND v_dow_unrestricted THEN
            -- Both '*' — pass (already gated by minute/hour/month)
            RETURN v_t;
        ELSIF v_dom_unrestricted THEN
            IF v_dow_ok THEN RETURN v_t; END IF;
        ELSIF v_dow_unrestricted THEN
            IF v_dom_ok THEN RETURN v_t; END IF;
        ELSE
            -- Both restricted — OR semantics
            IF v_dom_ok OR v_dow_ok THEN RETURN v_t; END IF;
        END IF;

        v_t := v_t + interval '1 minute';
    END LOOP;

    -- Nothing matched in 366 days — likely an impossible pattern (e.g.
    -- Feb 30). Return NULL so the caller can flag the row.
    RETURN NULL;
END;
$func$;

COMMENT ON FUNCTION stewards.cron_next_after(text, timestamptz) IS
'PE-B: returns the next timestamp >= p_after at which the standard 5-field cron pattern p_pattern fires. Treats p_pattern in UTC. Implements standard cron OR-semantics between day-of-month and day-of-week. Returns NULL if no match within 366 days.';

-- ---------------------------------------------------------------------
-- PE.B.1.d — trigger to materialize next_due_at on INSERT/UPDATE
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.scheduled_pipelines_compute_due()
RETURNS trigger
LANGUAGE plpgsql AS $func$
BEGIN
    -- Only recompute when cron_pattern changes (or on INSERT). Avoids
    -- recomputing every time enabled / input_template / notes change.
    IF TG_OP = 'INSERT'
       OR NEW.cron_pattern IS DISTINCT FROM OLD.cron_pattern
    THEN
        NEW.next_due_at := stewards.cron_next_after(NEW.cron_pattern, now());
    END IF;
    NEW.updated_at := now();
    RETURN NEW;
END;
$func$;

DROP TRIGGER IF EXISTS scheduled_pipelines_compute_due_tg ON stewards.scheduled_pipelines;
CREATE TRIGGER scheduled_pipelines_compute_due_tg
    BEFORE INSERT OR UPDATE ON stewards.scheduled_pipelines
    FOR EACH ROW EXECUTE FUNCTION stewards.scheduled_pipelines_compute_due();

COMMENT ON FUNCTION stewards.scheduled_pipelines_compute_due() IS
'PE-B: BEFORE INSERT/UPDATE trigger on scheduled_pipelines. Recomputes next_due_at via cron_next_after() whenever cron_pattern changes. Always bumps updated_at.';
