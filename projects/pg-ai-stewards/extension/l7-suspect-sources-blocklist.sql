-- =====================================================================
-- Batch L.7 — Source-domain blocklist (suspect_sources) + screen trigger
-- =====================================================================
-- A bilateral defense: when tool results from web_search / fetch_url
-- reference URLs whose domains appear in stewards.suspect_sources, the
-- result is annotated with a SUSPECT-SOURCE marker BEFORE the engram
-- extractor or downstream composers ever see it. Manual approval
-- table (suspect_source_approvals) lets the human override per-message.
--
-- Pure SQL — fires from an AFTER INSERT/UPDATE trigger on tool_calls.
-- =====================================================================


-- ---------------------------------------------------------------------
-- 1. suspect_sources blocklist + approval tables.
-- ---------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS stewards.suspect_sources (
    domain      text PRIMARY KEY,
    reason      text NOT NULL,
    severity    text NOT NULL DEFAULT 'warn' CHECK (severity IN ('warn','block')),
    added_at    timestamptz NOT NULL DEFAULT now(),
    added_by    text
);

COMMENT ON TABLE stewards.suspect_sources IS
'Batch L.7: domain-level blocklist for web_search / fetch_url tool results. severity=warn annotates with a marker; severity=block replaces content entirely. Editable by humans; agents do not write to this table.';

-- Seed common low-signal / known-AI-injection-target domains.
-- This is intentionally conservative — humans curate as patterns emerge.
INSERT INTO stewards.suspect_sources (domain, reason, severity, added_by) VALUES
('pastebin.com',          'public paste site — frequent injection vector', 'warn', 'l7-seed'),
('gist.github.com',       'public gists — possible injection vector',      'warn', 'l7-seed'),
('hastebin.com',          'public paste site',                              'warn', 'l7-seed')
ON CONFLICT (domain) DO NOTHING;


CREATE TABLE IF NOT EXISTS stewards.suspect_source_approvals (
    id              bigserial PRIMARY KEY,
    domain          text NOT NULL,
    message_id      bigint REFERENCES stewards.messages(id) ON DELETE CASCADE,
    approved_at     timestamptz NOT NULL DEFAULT now(),
    approved_by     text NOT NULL,
    rationale       text
);

CREATE INDEX IF NOT EXISTS suspect_source_approvals_domain_message
    ON stewards.suspect_source_approvals (domain, message_id);

COMMENT ON TABLE stewards.suspect_source_approvals IS
'Batch L.7: per-message approvals overriding the suspect_sources blocklist. Use when a human inspects a flagged result and decides it is safe in context. NULL message_id = global approval (rare).';


-- ---------------------------------------------------------------------
-- 2. Helper: extract domains from URLs in a jsonb blob.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.extract_domains_from_jsonb(p_doc jsonb)
RETURNS text[] LANGUAGE plpgsql IMMUTABLE AS $FN$
DECLARE
    v_text     text;
    v_match    text[];
    v_domains  text[] := ARRAY[]::text[];
    v_lower    text;
BEGIN
    IF p_doc IS NULL THEN
        RETURN v_domains;
    END IF;

    -- Cheapest approach: stringify and regex out hostnames.
    v_text := p_doc::text;

    FOR v_match IN
        SELECT regexp_matches(
            v_text,
            'https?://([a-zA-Z0-9.-]+)',
            'g'
        )
    LOOP
        v_lower := lower(v_match[1]);
        -- Strip leading 'www.' for canonical match.
        IF starts_with(v_lower, 'www.') THEN
            v_lower := substring(v_lower FROM 5);
        END IF;
        IF NOT (v_lower = ANY(v_domains)) THEN
            v_domains := array_append(v_domains, v_lower);
        END IF;
    END LOOP;

    RETURN v_domains;
END;
$FN$;


-- ---------------------------------------------------------------------
-- 3. Helper: is this domain (or any parent) suspect?
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.is_suspect_domain(p_domain text)
RETURNS stewards.suspect_sources LANGUAGE plpgsql STABLE AS $FN$
DECLARE
    v_row stewards.suspect_sources;
    v_d   text := lower(p_domain);
BEGIN
    -- Walk parent chain: foo.bar.example.com → bar.example.com → example.com
    WHILE v_d <> '' LOOP
        SELECT * INTO v_row FROM stewards.suspect_sources WHERE domain = v_d;
        IF v_row.domain IS NOT NULL THEN
            RETURN v_row;
        END IF;
        v_d := substring(v_d FROM position('.' IN v_d) + 1);
        IF position('.' IN v_d) = 0 THEN
            EXIT;
        END IF;
    END LOOP;
    RETURN NULL;
END;
$FN$;


-- ---------------------------------------------------------------------
-- 4. Trigger: AFTER INSERT/UPDATE OF result on tool_calls screens
--    web_search / fetch_url / scrape* tools and annotates the result.
-- ---------------------------------------------------------------------

CREATE OR REPLACE FUNCTION stewards.trigger_screen_suspect_sources()
RETURNS trigger LANGUAGE plpgsql AS $FN$
DECLARE
    v_domains   text[];
    v_domain    text;
    v_match     stewards.suspect_sources;
    v_approved  boolean;
    v_warnings  jsonb := '[]'::jsonb;
    v_severity  text;
    v_marked    boolean := false;
    v_new_res   jsonb;
BEGIN
    -- Only fire for web-fetching tools when result is non-null.
    IF NEW.result IS NULL THEN RETURN NEW; END IF;

    IF NEW.tool NOT IN ('web_search', 'web_search_exa', 'fetch_url', 'summarize_url', 'fetch_md') THEN
        RETURN NEW;
    END IF;

    -- Don't re-annotate if already screened (avoid loop on UPDATE).
    IF NEW.result ? '_suspect_screened' THEN
        RETURN NEW;
    END IF;

    v_domains := stewards.extract_domains_from_jsonb(NEW.result);

    FOREACH v_domain IN ARRAY v_domains LOOP
        v_match := stewards.is_suspect_domain(v_domain);
        IF v_match.domain IS NOT NULL THEN
            -- Check approval (per-message or global).
            SELECT EXISTS(
                SELECT 1 FROM stewards.suspect_source_approvals
                 WHERE domain = v_match.domain
                   AND (message_id IS NULL OR message_id = NEW.message_id)
            ) INTO v_approved;

            IF NOT v_approved THEN
                v_warnings := v_warnings || jsonb_build_array(jsonb_build_object(
                    'domain', v_match.domain,
                    'matched_via', v_domain,
                    'reason', v_match.reason,
                    'severity', v_match.severity
                ));
                IF v_match.severity = 'block' THEN
                    v_severity := 'block';
                ELSIF v_severity IS NULL OR v_severity <> 'block' THEN
                    v_severity := 'warn';
                END IF;
                v_marked := true;
            END IF;
        END IF;
    END LOOP;

    IF v_marked THEN
        IF v_severity = 'block' THEN
            v_new_res := jsonb_build_object(
                '_suspect_screened', true,
                '_suspect_severity', 'block',
                '_suspect_warnings', v_warnings,
                'content', '[SUSPECT-SOURCE BLOCKED] Result blocked by L.7 source-domain screen. ' ||
                           'See _suspect_warnings for details. Use suspect_source_approvals to override.'
            );
        ELSE
            v_new_res := NEW.result || jsonb_build_object(
                '_suspect_screened', true,
                '_suspect_severity', 'warn',
                '_suspect_warnings', v_warnings
            );
        END IF;

        UPDATE stewards.tool_calls SET result = v_new_res WHERE id = NEW.id;
    ELSE
        -- Mark screened-clean so we don't re-scan.
        UPDATE stewards.tool_calls
           SET result = NEW.result || jsonb_build_object('_suspect_screened', true)
         WHERE id = NEW.id;
    END IF;

    RETURN NEW;
END;
$FN$;

DROP TRIGGER IF EXISTS tool_calls_screen_suspect_sources ON stewards.tool_calls;

CREATE TRIGGER tool_calls_screen_suspect_sources
AFTER INSERT OR UPDATE OF result ON stewards.tool_calls
FOR EACH ROW
WHEN (NEW.result IS NOT NULL AND NOT (NEW.result ? '_suspect_screened'))
EXECUTE FUNCTION stewards.trigger_screen_suspect_sources();

COMMENT ON FUNCTION stewards.trigger_screen_suspect_sources() IS
'Batch L.7: AFTER INSERT/UPDATE OF result on tool_calls. For web-fetching tools, extracts domains from the result and screens them against suspect_sources (walking parent chain), honoring per-message approvals in suspect_source_approvals. severity=block replaces content; severity=warn annotates.';


-- =====================================================================
-- End of l7-suspect-sources-blocklist.sql
-- =====================================================================
