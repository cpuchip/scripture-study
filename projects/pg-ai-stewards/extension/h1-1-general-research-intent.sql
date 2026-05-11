-- =====================================================================
-- Batch H.1.1 — seed the `general-research` intent
--
-- D-H3 (ratified 2026-05-11) Path B: three new intents total
--   - general-research (this file)     — covers H.1 + H.2 yt-secular
--   - professional-awareness (H.3)     — covers ai-news-summary
--   - creative-fidelity (H.4)          — covers fiction-scene
--
-- Storage: .spec/intents/<slug>.yaml (documentation + future-Rust-parser
-- source). Today seeded directly via SQL because the Rust parser at
-- src/yaml.rs hardcodes slug='scripture-study' and uses the old YAML
-- shape. Rule-of-three triggers the Rust refactor in H.3 when the third
-- new intent joins; this seeder + the YAML doc converge then.
--
-- The intent shape matches stewards.intents columns. scripture_anchor
-- explicitly NULL classifies this intent as low-stakes per the
-- 2026-05-11 §VI amendment for D-F2: master-tier agents can bishop
-- councils convened on a general-research intent.
-- =====================================================================

INSERT INTO stewards.intents (
    slug,
    purpose,
    beneficiary,
    values_hierarchy,
    non_goals,
    scripture_anchor,
    source_file,
    source_yaml_sha,
    updated_at
)
VALUES (
    'general-research',
    'Cast a wider net than scripture-study — gather, summarize, and reason about non-doctrinal sources to inform Michael''s understanding of fields he''s actively working in (AI, engineering, product, education, professional skill acquisition).',
    'Michael primarily; secondary readers of any digest/report',
    jsonb_build_array(
        jsonb_build_object(
            'key', 'credibility-over-volume',
            'description', 'One credible source beats five rumors. Refuse to summarize what can''t be sourced.'
        ),
        jsonb_build_object(
            'key', 'skepticism-as-default',
            'description', 'Treat each claim as needing evidence. Note where the source is opinion vs reporting vs primary documentation.'
        ),
        jsonb_build_object(
            'key', 'recency-matters',
            'description', 'A 2024 take on AI tooling is obsolete. Weight recency where the domain moves fast; flag where a source is older than the topic''s half-life.'
        ),
        jsonb_build_object(
            'key', 'honest-uncertainty',
            'description', '"I couldn''t find a credible source on X" is a valid output. Better than fabrication.'
        ),
        jsonb_build_object(
            'key', 'cross-reference',
            'description', 'Where claims appear in multiple independent sources, say so. Where they appear in only one, say so explicitly.'
        )
    ),
    ARRAY[
        'Doctrinal claims (those go through scripture-study)',
        'Personal recommendations (Michael draws his own conclusions; we summarize)',
        'Source-laundering (rephrasing without attribution)',
        'Speculation framed as reporting'
    ],
    NULL,  -- scripture_anchor: explicitly NULL → low-stakes per D-F2
    '.spec/intents/general-research.yaml',
    encode(sha256('h1-1-seed-2026-05-11'::bytea), 'hex'),  -- placeholder sha; Rust parser will compute real sha on H.3 refactor
    now()
)
ON CONFLICT (slug) DO UPDATE SET
    purpose          = EXCLUDED.purpose,
    beneficiary      = EXCLUDED.beneficiary,
    values_hierarchy = EXCLUDED.values_hierarchy,
    non_goals        = EXCLUDED.non_goals,
    scripture_anchor = EXCLUDED.scripture_anchor,
    source_file      = EXCLUDED.source_file,
    source_yaml_sha  = EXCLUDED.source_yaml_sha,
    updated_at       = now();
