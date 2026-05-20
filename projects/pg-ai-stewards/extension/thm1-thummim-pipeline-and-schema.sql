-- =====================================================================
-- Batch THM.1 — Thummim 2026 Restoration Dictionary pipeline + schema
--
-- Per proposal `.spec/proposals/thummim-restoration-dictionary.md`.
-- Defines the generation machinery so the next session can dispatch
-- the v1 corpus (~150-200 words) once D-THM-1..6 are ratified.
--
-- This file ships:
--   1. stewards.thummim_entries table (one row per word; jsonb-keyed levels)
--   2. stewards.pipelines row for 'thummim-define' (3 stages, mirrors
--      research-write's shape but with dictionary-specific prompts)
--   3. stewards.stage_models + pipeline_stage_maturity rows
--
-- This file does NOT:
--   - Dispatch any work_items (cost discipline; Michael's call)
--   - Extend stewards.promote_to_study to handle the new family (one-line
--     change; surface as carry-forward until Michael ratifies that
--     thummim entries should ride the studies+AGE path)
-- =====================================================================

-- ---------------------------------------------------------------------
-- THM.1.a — stewards.thummim_entries
-- ---------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS stewards.thummim_entries (
    word                    text PRIMARY KEY,
    work_item_id            uuid REFERENCES stewards.work_items(id) ON DELETE SET NULL,
    levels                  jsonb NOT NULL DEFAULT '{}'::jsonb,
        -- Shape: {
        --   "elementary":   {"headline": "…", "body": "…"},
        --   "eighth_grade": {"headline": "…", "body": "…"},
        --   "college_plus": {
        --       "headline": "…",
        --       "body": "…",
        --       "key_passages": ["D&C 84:33", "Mosiah 4:14-15", ...],
        --       "conference_refs": ["Cook Apr 2019", "Bednar Oct 2014", ...]
        --   }
        -- }
    webster_1828_compare    text,
    substrate_study         text,
        -- e.g. "study/priesthood-oath-and-covenant.md" for 'obtain'
    generated_at            timestamptz NOT NULL DEFAULT now(),
    updated_at              timestamptz NOT NULL DEFAULT now(),
    body_tsv                tsvector GENERATED ALWAYS AS (
        to_tsvector('english',
            word || ' ' ||
            coalesce(webster_1828_compare, '') || ' ' ||
            coalesce(levels::text, '')
        )
    ) STORED
);

CREATE INDEX IF NOT EXISTS thummim_entries_fts_idx
    ON stewards.thummim_entries USING gin (body_tsv);

CREATE INDEX IF NOT EXISTS thummim_entries_levels_idx
    ON stewards.thummim_entries USING gin (levels);

COMMENT ON TABLE stewards.thummim_entries IS
'THM: Restoration-era dictionary entries. One row per word; multi-level renderings stored as jsonb. Populated by the thummim-define pipeline (substrate-pipelines-expansion + thummim-restoration-dictionary proposals).';

COMMENT ON COLUMN stewards.thummim_entries.levels IS
'jsonb of three grade-level renderings: elementary (age 7-11), eighth_grade (age 13-14), college_plus. Each has headline + body; college_plus additionally has key_passages array + conference_refs array. Schema is intentionally flexible — add more levels (lifelong-learner, scholarly) without ALTER TABLE.';

-- ---------------------------------------------------------------------
-- THM.1.b — 'thummim-define' pipeline
--
-- Three stages mirroring research-write's shape (gather → synthesize →
-- review). Agent family `research` (already has the right tool grants:
-- study_search_text, gospel_get, gospel_search, byu_citations, fetch_url).
--
-- Sabbath disabled — Thummim entries are reference material, not
-- creative study work needing reflection. Atonement enabled — these
-- will run in volume (~150-200 entries), and a cost cap is the safety
-- net for misbehaving prompts.
-- ---------------------------------------------------------------------

INSERT INTO stewards.pipelines (
    family, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath, maturity_ladder,
    auto_materialize_on_verified
)
VALUES (
    'thummim-define',
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
                'Define this word for the Thummim 2026 Restoration Dictionary: **{{input.word}}**' || E'\n\n' ||
                'Binding question: {{input.binding_question}}' || E'\n\n' ||
                'You are gathering evidence from the Restoration corpus for a dictionary entry. The dictionary surfaces what the Restoration ITSELF does with each word — not what 1828 secular English means by it. (Webster 1828 is a companion lens; we cross-reference but do not define from it.)' || E'\n\n' ||
                'Use these tools to gather:' || E'\n' ||
                '  - study_search_text("{{input.word}}", kinds=["scripture"]) — find scripture passages using this word' || E'\n' ||
                '  - study_search_text("{{input.word}}", kinds=["study"]) — find our substrate study work on this word' || E'\n' ||
                '  - gospel_search("{{input.word}}", mode="hybrid") — search the gospel corpus (engine.ibeco.me)' || E'\n' ||
                '  - byu_citations(verse_ref) — for the key scripture passages you find, see which GC talks cite them (this is how you discover conference reinforcement)' || E'\n' ||
                '  - gospel_get(ref) — verify a passage you intend to cite' || E'\n\n' ||
                'Produce a gather brief in markdown with these sections:' || E'\n\n' ||
                '1. **Scripture usage patterns** — group the passages into 2-5 distinct usage senses. For each sense: a short label, 2-4 verbatim passages with refs, and a one-sentence summary of what the word DOES in those passages.' || E'\n' ||
                '2. **GC reinforcement** — 3-8 conference talks where apostles or prophets have built on the scriptural sense. For each: title, speaker, date, and a 1-sentence summary of how they used the word.' || E'\n' ||
                '3. **Comparison to Webster 1828** — what is Webster 1828''s first sense of this word (if any)? Does the Restoration sense reinforce, sharpen, or diverge from it? Don''t infer — quote Webster 1828 directly if you have access; otherwise note "unknown — Webster 1828 not consulted in this stage."' || E'\n' ||
                '4. **Cross-references in our studies** — which of our own substrate studies have lensed this word? (study_search_text on kinds=["study"] should surface these.)' || E'\n\n' ||
                'Quality bar: every passage citation in the brief points to a real verse. If you can''t find at least 3 distinct scriptural usages, say so — that''s data, not failure.'
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
                'Word: **{{input.word}}**' || E'\n\n' ||
                'Gather brief from the previous stage:' || E'\n\n' ||
                '{{stage_results.gather.output}}' || E'\n\n' ||
                'Write the dictionary entry in three grade-level renderings.' || E'\n\n' ||
                '## Output format' || E'\n\n' ||
                'A single JSON object with this exact shape (no markdown wrapping; the substrate will parse this):' || E'\n\n' ||
                '```json' || E'\n' ||
                '{' || E'\n' ||
                '  "word": "{{input.word}}",' || E'\n' ||
                '  "levels": {' || E'\n' ||
                '    "elementary": {' || E'\n' ||
                '      "headline": "...one short sentence...",' || E'\n' ||
                '      "body": "...3-5 sentences, age 7-11 vocabulary; one concrete example..."' || E'\n' ||
                '    },' || E'\n' ||
                '    "eighth_grade": {' || E'\n' ||
                '      "headline": "...one fuller sentence...",' || E'\n' ||
                '      "body": "...5-10 sentences with 2-3 scripture examples; doctrinal context..."' || E'\n' ||
                '    },' || E'\n' ||
                '    "college_plus": {' || E'\n' ||
                '      "headline": "...nuanced one-sentence summary...",' || E'\n' ||
                '      "body": "...full exegesis, 8-15 sentences, multi-passage; engage Webster 1828 comparison where relevant...",' || E'\n' ||
                '      "key_passages": ["D&C 84:33", "Mosiah 4:14-15"],' || E'\n' ||
                '      "conference_refs": ["Cook Apr 2019 Great Love", "Bednar Oct 2014 Spirit of Revelation"]' || E'\n' ||
                '    }' || E'\n' ||
                '  },' || E'\n' ||
                '  "webster_1828_compare": "...one paragraph naming where the Restoration sense aligns with / diverges from Webster 1828..."' || E'\n' ||
                '}' || E'\n' ||
                '```' || E'\n\n' ||
                'Voice notes per level:' || E'\n' ||
                '  - **elementary**: short sentences, concrete words. Imagine reading aloud to a Primary class. No multi-clause sentences.' || E'\n' ||
                '  - **eighth_grade**: classroom voice. Same vocabulary you''d use in a Sunday School class for young men/women. 2-3 scripture examples woven in.' || E'\n' ||
                '  - **college_plus**: full scholarly voice. Engage Webster 1828 directly. Cite passages by reference (D&C 84:33), not paraphrase. Name doctrinal connections; acknowledge ambiguity where present.' || E'\n\n' ||
                'Every quoted scripture must be in the gather brief — do not introduce new citations at this stage. Honest uncertainty is preferred over fabricated precision. If the Restoration usage is genuinely the same as the modern dictionary sense, say so plainly in the webster_1828_compare field.'
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
                'Word: **{{input.word}}**' || E'\n\n' ||
                'Synthesize-stage output (the draft entry as JSON):' || E'\n\n' ||
                '{{stage_results.synthesize.output}}' || E'\n\n' ||
                'Review the draft against five criteria. Tools are DISABLED; review on the text + gather brief alone.' || E'\n\n' ||
                '1. **Citation fidelity** — every scripture ref in `key_passages` matches a passage actually quoted or paraphrased in the body. No invented references.' || E'\n\n' ||
                '2. **Grade-level voice** — elementary is genuinely elementary (no 4-syllable words; short sentences; one example). 8th grade is classroom voice with 2-3 examples. college_plus is scholarly with engagement of Webster 1828.' || E'\n\n' ||
                '3. **No invented GC citations** — conference_refs should match talks that actually exist. If you''re uncertain about a specific talk title, flag it. Better to drop a doubtful citation than invent one.' || E'\n\n' ||
                '4. **Restoration-first definition** — the body defines what the Restoration DOES with the word. It does not import 1828 secular English as the primary meaning. The webster_1828_compare section is where the comparison lives; the main definition stays inside the Restoration corpus.' || E'\n\n' ||
                '5. **Honest uncertainty over fabricated precision** — if the gather brief showed only 2 distinct usages, the entry should not pretend there were 5. If the Restoration sense is genuinely the same as modern, the webster_1828_compare field should say so plainly.' || E'\n\n' ||
                'Return ONE of:' || E'\n' ||
                '(a) The same JSON, verbatim and unchanged. Prefix with "REVIEW: passes" then a blank line then the JSON.' || E'\n' ||
                '(b) A revised JSON object. Prefix with "REVIEW: revised" then a blank line, the revised JSON, and at the end a brief notes section listing what changed and why.'
        )
    ),
    false,  -- sabbath_enabled (reference material, not creative work)
    true,   -- atonement_enabled (~150 entries will run in volume; cap matters)
    'research/dictionary/<slug>.md',
    NULL,
    '["raw","researched","planned","specced","executing","verified"]'::jsonb,
    true    -- auto_materialize_on_verified (entries should auto-write to disk)
)
ON CONFLICT (family) DO UPDATE SET
    stages                       = EXCLUDED.stages,
    sabbath_enabled              = EXCLUDED.sabbath_enabled,
    atonement_enabled            = EXCLUDED.atonement_enabled,
    file_destination_template    = EXCLUDED.file_destination_template,
    file_content_jsonpath        = EXCLUDED.file_content_jsonpath,
    maturity_ladder              = EXCLUDED.maturity_ladder,
    auto_materialize_on_verified = EXCLUDED.auto_materialize_on_verified;

-- Stage models — mirrors research-write's pattern (kimi for gather +
-- synthesize, qwen for review).
INSERT INTO stewards.stage_models (pipeline_family, stage_name, default_model, notes) VALUES
    ('thummim-define', 'gather',     'kimi-k2.6',    'Restoration-corpus gather: study_search_text + gospel_search + byu_citations + gospel_get. Builds the scripture-usage + GC-reinforcement brief.'),
    ('thummim-define', 'synthesize', 'kimi-k2.6',    'Three-grade-level rendering produced as a single JSON object. webster_1828_compare written as a paragraph.'),
    ('thummim-define', 'review',     'qwen3.6-plus', 'Tools-disabled JSON verification — citation fidelity + grade-level voice + no-invented-refs + Restoration-first definition + honest uncertainty.')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE SET
    default_model = EXCLUDED.default_model,
    notes         = EXCLUDED.notes;

-- Maturity rung mapping
INSERT INTO stewards.pipeline_stage_maturity (pipeline_family, stage_name, produces_maturity, notes) VALUES
    ('thummim-define', 'gather',     'researched', 'Corpus evidence gathered; gather brief assembled.'),
    ('thummim-define', 'synthesize', 'planned',    'Three-grade-level JSON drafted. Plan + execute combined here as in research-write.'),
    ('thummim-define', 'review',     'verified',   'Citation + voice + Restoration-first verification complete. Entry ready for materialization + thummim_entries upsert.')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE SET
    produces_maturity = EXCLUDED.produces_maturity,
    notes             = EXCLUDED.notes;
