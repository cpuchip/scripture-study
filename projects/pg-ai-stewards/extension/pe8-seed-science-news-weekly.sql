-- =====================================================================
-- PE.B.8 — seed the science-news-weekly scheduled pipeline
--
-- Companion to ai-news-7am (seeded in pe7-...sql). Where ai-news fires
-- daily and uses the research-summary pipeline (sources_spec with explicit
-- queries), this fires weekly and uses research-write (binding_question
-- only, structured Headlines/Notable/Skeptical/Open Questions output).
--
-- Cadence:   Mondays 13:00 UTC = 7am MT = '0 13 * * 1'
-- Pipeline:  research-write — produces the format Michael loved in
--            research/physics-news-20260503-science-center-roundup.md
--            (commit 220cf35, 2026-05-11). Three stages: gather (tools
--            on, web_search_exa + fetch_url), synthesize (structured
--            output), review (verification).
-- Scope:     Physics + general science (astronomy, chemistry, biology,
--            materials, climate, neuroscience) with science-center
--            exhibit translations.
-- Output:    research/science-news-weekly--YYYY-MM-DD-1300.md
--            (research-write's file_destination_template = research/<slug>.md;
--            scheduled_pipelines_fire appends --YYYY-MM-DD-HHMM to slug.)
-- Missed:    72h fire-one-missed window — if Monday morning is missed,
--            fire any time through Thursday evening, otherwise skip and
--            advance to next Monday. Daily ai-news is 24h; weekly cadence
--            tolerates a few days of slip.
--
-- Idempotent via ON CONFLICT (slug) DO UPDATE. Live-apply via
--   docker cp + psql -f per substrate CLAUDE.md §4.
-- =====================================================================

INSERT INTO stewards.scheduled_pipelines (
    slug, pipeline_family, intent_id, cron_pattern, input_template,
    enabled, missed_window_hours, notes
)
VALUES (
    'science-news-weekly',
    'research-write',
    (SELECT id FROM stewards.intents WHERE slug = 'general-research'),
    '0 13 * * 1',
    jsonb_build_object(
        'binding_question',
            'What significant findings in physics and general science were reported in the past week, '
            || 'and what hands-on exhibits or demonstrations could a community science center build from each result?'
            || E'\n\n'
            || 'Cover physics, astronomy, chemistry, biology, materials science, climate, and neuroscience. '
            || 'Prefer primary sources (institutional press releases, the paper itself, the official announcement) '
            || 'over secondary aggregators; when a finding only appears through a tertiary aggregator, say so and '
            || 'mark it as awaiting confirmation.'
            || E'\n\n'
            || 'For every headline finding, provide a "Science-center translation" — a concrete, buildable exhibit '
            || '(target sub-$500 in materials where possible) that visitors can interact with, with honest caveats '
            || 'when the underlying research requires equipment a small science center cannot reasonably host. '
            || 'Include the physical principle the exhibit teaches, not just the visual hook.'
            || E'\n\n'
            || 'Structure:'
            || E'\n'
            || '  - **Headlines** — 3-5 findings that translate to buildable exhibits, each with a verbatim quote, '
            || 'source link, and exhibit translation'
            || E'\n'
            || '  - **Notable** — 1-3 findings worth a wall panel even if not a full exhibit'
            || E'\n'
            || '  - **Skeptical Takes** — every press-release headline carries a limitation; surface them honestly. '
            || 'Null results, unexplained phenomena, and "we confirmed it but cannot yet explain it" framings teach '
            || 'the scientific method.'
            || E'\n'
            || '  - **Open Questions** — what the sources do not answer'
            || E'\n'
            || '  - **Synthesis** — a short closing paragraph on the week as a whole'
            || E'\n\n'
            || 'Tone: concrete, direct, unadorned. The audience is a community science center planning real exhibits '
            || 'on a real budget; vague enthusiasm wastes their time. Honest "this finding does not exhibit well" '
            || 'is more useful than forced exhibit ideas.'
    ),
    true,
    72,
    'Weekly physics + general-science news with science-center exhibit translations. Mondays 7am MT (13:00 UTC). '
    || 'Pipeline = research-write (Headlines/Notable/Skeptical/Open Questions structure, tools-on gather, qwen review). '
    || 'Format anchor: research/physics-news-20260503-science-center-roundup.md (commit 220cf35). '
    || 'Output materializes to research/science-news-weekly--YYYY-MM-DD-1300.md. '
    || 'Republish to marsfield.org /science-news after manual review — same workflow as the 2026-05-22 first post. '
    || 'Per D-PE4: 72h missed-window allows mid-week catch-up if Monday tick is missed; skip + advance otherwise.'
)
ON CONFLICT (slug) DO UPDATE SET
    pipeline_family     = EXCLUDED.pipeline_family,
    intent_id           = EXCLUDED.intent_id,
    cron_pattern        = EXCLUDED.cron_pattern,
    input_template      = EXCLUDED.input_template,
    enabled             = EXCLUDED.enabled,
    missed_window_hours = EXCLUDED.missed_window_hours,
    notes               = EXCLUDED.notes,
    updated_at          = now();

-- Verify seed + show what next_due_at the BEFORE INSERT trigger picked.
DO $$
DECLARE
    v_row stewards.scheduled_pipelines%ROWTYPE;
BEGIN
    SELECT * INTO v_row
      FROM stewards.scheduled_pipelines
     WHERE slug = 'science-news-weekly';
    IF NOT FOUND THEN
        RAISE EXCEPTION 'science-news-weekly seed failed: row not found after insert';
    END IF;
    RAISE NOTICE 'science-news-weekly seeded: pipeline_family=% cron=% enabled=% next_due_at=% missed_window=%h',
        v_row.pipeline_family, v_row.cron_pattern, v_row.enabled, v_row.next_due_at, v_row.missed_window_hours;
END $$;
