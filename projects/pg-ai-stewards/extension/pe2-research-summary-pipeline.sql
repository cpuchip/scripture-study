-- =====================================================================
-- Batch PE.2 — seed the `research-summary` pipeline (daily-digest)
--
-- D-PE1 + D-PE1' (ratified 2026-05-19): pipeline selection itself is a
-- judgment — research-write covers deep-research, research-summary covers
-- daily-digest. Same agent (research), same model assignments as
-- research-write, lighter input templates, sabbath/atonement OFF.
--
-- Intent: general-research (per D-PE2'; reuses existing intent rather
-- than creating professional-awareness). Two YT-aware values were added
-- to general-research in PE.1; this pipeline carries them through its
-- system prompt automatically via the intent block.
--
-- File destination: study/daily-digest/<slug>.md so digests don't bloat
-- the main study/ directory. auto_materialize_on_verified=true since
-- digests are one-off and don't need a separate sabbath gate.
-- =====================================================================

INSERT INTO stewards.pipelines (
    family, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath, maturity_ladder,
    auto_materialize_on_verified
)
VALUES (
    'research-summary',
    jsonb_build_array(
        jsonb_build_object(
            'name',            'gather',
            'next',            'synthesize',
            'model',           'kimi-k2.6',
            'provider',        'opencode_go',
            'agent_family',    'research',
            'auto_advance',    true,
            'tools_disabled',  false,
            'input_template',
                'Binding question: {{input.binding_question}}' || E'\n\n' ||
                'You are gathering items for a DAILY DIGEST that answers the binding question above. This is not a deep research piece — it is a 24-hour news scan.' || E'\n\n' ||
                'Use the tools available (web_search_exa, web_search, fetch_url, yt_*, etc.) to find 4-8 noteworthy items from the last 24 hours that bear on the binding question. Prefer primary sources (official announcements, vendor docs, the paper itself). Secondary reporting only when it adds context the primary source omits.' || E'\n\n' ||
                'For each item kept, capture:' || E'\n' ||
                '  - Title + URL + publication date/time' || E'\n' ||
                '  - One-sentence summary of what shipped or was reported' || E'\n' ||
                '  - A short verbatim quote (1-2 sentences) you might draw on in the synthesis' || E'\n' ||
                '  - Item type: official-release, news-reporting, vendor-blog, opinion-piece, social-media-thread' || E'\n\n' ||
                'The general-research intent applies — apply credibility-over-volume, skepticism-as-default, and surface-the-rhetoric. A loud headline is not evidence of a substantive change; flag rhetorical heat that isn''t backed by a concrete release or document.' || E'\n\n' ||
                'Recency is the whole point of a daily digest: items older than 48 hours need a strong justification to keep. If a story keeps trending on day 3, that itself is the news — note the trending arc, not the original event.' || E'\n\n' ||
                'Produce an items brief — a structured list of every item kept, with the four fields above. The next stage drafts the digest from this brief. Do NOT write the digest yet.'
        ),
        jsonb_build_object(
            'name',            'synthesize',
            'next',            'review',
            'model',           'kimi-k2.6',
            'provider',        'opencode_go',
            'agent_family',    'research',
            'auto_advance',    true,
            'tools_disabled',  false,
            'input_template',
                'Binding question: {{input.binding_question}}' || E'\n\n' ||
                'Items brief from the gather stage:' || E'\n\n' ||
                '{{stage_results.gather.output}}' || E'\n\n' ||
                'Now write the daily digest. Aim for 300-700 words total. This is a scan, not a deep dive — Michael will read it once and move on. If a single item warrants depth, name it and recommend a follow-up deep-research run rather than expanding inline.' || E'\n\n' ||
                'Attribution: every claim has an inline markdown link to the source it came from: [Title](URL). Paraphrase by default; quote verbatim only when you have the source text in front of you in this session.' || E'\n\n' ||
                'Structure (adapt to what the day actually produced):' || E'\n' ||
                '  - **Headlines** — the 1-3 most important items of the day, one short paragraph each' || E'\n' ||
                '  - **Notable** — second-tier items worth knowing, one-line each with link' || E'\n' ||
                '  - **Skeptical takes** — credible dissenting voices on any headline item, if any' || E'\n' ||
                '  - **Carry-forward** — what to watch for tomorrow; any deep-research candidates' || E'\n\n' ||
                'No filler. If a day produced nothing noteworthy, the digest can be three lines: "Slow news day. [link to the one minor thing]. Carry-forward: nothing." Honest emptiness beats manufactured importance.' || E'\n\n' ||
                'Produce the complete digest in markdown. The next stage reviews it.'
        ),
        jsonb_build_object(
            'name',            'review',
            'next',            NULL,
            'model',           'qwen3.6-plus',
            'provider',        'opencode_go',
            'agent_family',    'research',
            'auto_advance',    true,
            'tools_disabled',  true,
            'input_template',
                'Binding question: {{input.binding_question}}' || E'\n\n' ||
                'The digest draft from the previous stage:' || E'\n\n' ||
                '{{stage_results.synthesize.output}}' || E'\n\n' ||
                'Review the digest against four criteria:' || E'\n\n' ||
                '1. **Attribution.** Every claim has an inline link. No claims without a source. If a claim is the synthesizer''s own observation, it is named as such ("These three releases together suggest...") rather than presented as reporting.' || E'\n\n' ||
                '2. **Recency.** Every item is from within the last 24-48 hours, OR the item is explicitly framed as a "still trending" follow-up to an older event.' || E'\n\n' ||
                '3. **Rhetorical inflation.** No headline manufactured from minor news. No urgency that isn''t in the underlying source. Flag any item where the digest''s framing is hotter than the source''s.' || E'\n\n' ||
                '4. **Honest emptiness.** If the day was slow, the digest says so. No padding.' || E'\n\n' ||
                'Tools are DISABLED for this stage. You CANNOT fetch URLs — review on the digest text + its in-line links only.' || E'\n\n' ||
                'Return ONE of:' || E'\n' ||
                '(a) The same digest, verbatim and unchanged, if it passes all four criteria. Prefix with a single line: "REVIEW: passes" then a blank line then the digest.' || E'\n' ||
                '(b) A revised digest. Prefix with "REVIEW: revised" then a blank line, the revised digest, and at the end a brief notes section listing what changed and why.'
        )
    ),
    false,  -- sabbath_enabled: daily-digest is transient; no sabbath reflection
    false,  -- atonement_enabled: digest hitting cost cap means refine the queries, not atone
    'study/daily-digest/<slug>.md',
    NULL,   -- file_content_jsonpath: v1 uses whole stage output
    '["raw","researched","planned","specced","executing","verified"]'::jsonb,
    true    -- auto_materialize_on_verified: digests are one-off; no manual gate
)
ON CONFLICT (family) DO UPDATE SET
    stages                       = EXCLUDED.stages,
    sabbath_enabled              = EXCLUDED.sabbath_enabled,
    atonement_enabled            = EXCLUDED.atonement_enabled,
    file_destination_template    = EXCLUDED.file_destination_template,
    file_content_jsonpath        = EXCLUDED.file_content_jsonpath,
    maturity_ladder              = EXCLUDED.maturity_ladder,
    auto_materialize_on_verified = EXCLUDED.auto_materialize_on_verified;

-- Stage models for the three cells (mirrors research-write to avoid new
-- LLM-cost surface; daily volume is low enough that model choice doesn't
-- meaningfully shift cost vs. simplifying the catalog).
INSERT INTO stewards.stage_models (pipeline_family, stage_name, default_model, notes) VALUES
    ('research-summary', 'gather',     'kimi-k2.6',    'Daily-digest source gather; tools enabled (exa, web_search, fetch_url, yt_*). 24-hour scan, not deep research.'),
    ('research-summary', 'synthesize', 'kimi-k2.6',    'Daily-digest synthesis from gather brief; tools enabled lightly (re-fetch only). 300-700 word target.'),
    ('research-summary', 'review',     'qwen3.6-plus', 'Tools-disabled verification pass; checks attribution + recency + rhetorical inflation + honest emptiness.')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE SET
    default_model = EXCLUDED.default_model,
    notes         = EXCLUDED.notes;

-- Maturity rung mapping. research-summary follows the same shape as
-- research-write — synthesize IS the draft, no separate executing rung.
INSERT INTO stewards.pipeline_stage_maturity (pipeline_family, stage_name, produces_maturity, notes) VALUES
    ('research-summary', 'gather',     'researched', 'Items collected + summarized; ready for synthesis.'),
    ('research-summary', 'synthesize', 'planned',    'Draft is the plan. No separate executing rung — daily-digest has no draft-vs-execute distinction.'),
    ('research-summary', 'review',     'verified',   'Review pass complete; digest is verified and auto-materializes.')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE SET
    produces_maturity = EXCLUDED.produces_maturity,
    notes             = EXCLUDED.notes;
