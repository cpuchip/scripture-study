-- =====================================================================
-- Batch PE.3 — seed the `yt-gospel-evaluate` pipeline
--
-- D-PE2 (ratified 2026-05-19): yt-gospel stays under the scripture-study
-- intent (its discernment rubric fits faith-as-framework / trust-the-
-- discernment). Intent is attached per work_item at work_item_create,
-- not per pipeline — this seed defines stages only.
--
-- D-PE5 (ratified 2026-05-19): transcript + metadata enrichment. Ingest
-- pulls transcript via yt_download AND chapters, full description, top
-- comments via yt_get / yt_list for richer evaluator context.
--
-- D-D1: sabbath + atonement ON — gospel video evaluation is doctrinal
-- work, treated like study-write (sabbath reflection valuable; cost-cap
-- worth atoning over).
--
-- Three stages: ingest -> evaluate -> review. Agent family yt-gospel
-- throughout (registered, has byu_citations + gospel_* + yt/* perms).
-- Review stage tools-off + qwen3.6-plus per research-write convention.
-- =====================================================================

INSERT INTO stewards.pipelines (
    family, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath, maturity_ladder,
    auto_materialize_on_verified
)
VALUES (
    'yt-gospel-evaluate',
    jsonb_build_array(
        jsonb_build_object(
            'name',            'ingest',
            'next',            'evaluate',
            'model',           'kimi-k2.6',
            'provider',        'opencode_go',
            'agent_family',    'yt-gospel',
            'auto_advance',    true,
            'tools_disabled',  false,
            'input_template',
                'Binding question: {{input.binding_question}}' || E'\n\n' ||
                'Video URL: {{input.video_url}}' || E'\n\n' ||
                'You are ingesting a YouTube video for gospel evaluation. The next stage applies the Restoration discernment rubric to your output — quality of ingest governs quality of evaluation downstream.' || E'\n\n' ||
                'Use the yt tools to capture:' || E'\n' ||
                '  1. **Transcript** — call yt_download(video_url) to capture the full transcript. If yt_download returns a path/identifier, follow with yt_get to retrieve the transcript text itself.' || E'\n' ||
                '  2. **Metadata** — yt_get also returns the channel, title, publication date, duration, view count. Capture all of it.' || E'\n' ||
                '  3. **Description** — the full video description (not just the first line) — creators often signal their framing in the description.' || E'\n' ||
                '  4. **Chapters** — if the video has chapter markers, list them; they reveal the creator''s own structuring of the argument.' || E'\n' ||
                '  5. **Top comments** — sample 5-10 top comments. Useful as a witness to how the video is being received (alignment, pushback, confusion).' || E'\n\n' ||
                'If a tool fails (private video, no transcript available, region lock), note that as the ingest output and stop. The evaluate stage cannot run without transcript text — return "INGEST: failed — <reason>" if so.' || E'\n\n' ||
                'Produce a structured ingest brief in markdown with sections:' || E'\n' ||
                '  - **Identity** — title, channel, URL, date, duration' || E'\n' ||
                '  - **Description** (full)' || E'\n' ||
                '  - **Chapters** (if any)' || E'\n' ||
                '  - **Transcript** (the full text, with timestamps if yt_download provided them)' || E'\n' ||
                '  - **Top-comment sample** (5-10 comments, attribution and brief gloss)' || E'\n\n' ||
                'Do NOT evaluate yet. The next stage does that.'
        ),
        jsonb_build_object(
            'name',            'evaluate',
            'next',            'review',
            'model',           'kimi-k2.6',
            'provider',        'opencode_go',
            'agent_family',    'yt-gospel',
            'auto_advance',    true,
            'tools_disabled',  false,
            'input_template',
                'Binding question: {{input.binding_question}}' || E'\n\n' ||
                'Ingest brief from the previous stage:' || E'\n\n' ||
                '{{stage_results.ingest.output}}' || E'\n\n' ||
                'Apply the Restoration discernment standard to this video. The scripture-study intent applies — your active system prompt carries its values; trust-the-discernment and faith-as-framework govern this work.' || E'\n\n' ||
                'Structure your evaluation in five sections:' || E'\n\n' ||
                '1. **Binding question** — restate what we are evaluating and what verdict would look like. Do not begin evaluation until the question is sharp.' || E'\n\n' ||
                '2. **Evidence** — what does the video actually claim? Pull 5-10 verbatim quotes from the transcript that carry the argument (with timestamps where ingest provided them). Distinguish what the creator asserts vs. what they merely report.' || E'\n\n' ||
                '3. **Alignment with canon** — for each substantive claim, check scriptural citation density via byu_citations on the verses cited or alluded to. Where the video cites scripture explicitly, verify against the actual passage (use gospel_get if needed). Where the video gestures at scripture without citation, name what passage it''s drawing on and check whether the use is faithful or strained.' || E'\n\n' ||
                '4. **Witness questions** — what would the Spirit be checking that we cannot? Name 2-4 questions the human reader should hold up against the video that no source-verification tool can answer for them.' || E'\n\n' ||
                '5. **Becoming** — if the video''s claims were applied, what would change in lived practice? Where would those changes be salutary, where would they bend off-track?' || E'\n\n' ||
                'Length: 1200-3000 words. Honest uncertainty is preferred to manufactured verdict. If the video is mostly fine with one or two flaws, say so plainly rather than inflating either the flaws or the praise. The Ben Test applies: if we''re writing flattery we should stop.' || E'\n\n' ||
                'Produce the complete evaluation in markdown. The next stage reviews it.'
        ),
        jsonb_build_object(
            'name',            'review',
            'next',            NULL,
            'model',           'qwen3.6-plus',
            'provider',        'opencode_go',
            'agent_family',    'yt-gospel',
            'auto_advance',    true,
            'tools_disabled',  true,
            'input_template',
                'Binding question: {{input.binding_question}}' || E'\n\n' ||
                'The evaluation draft from the previous stage:' || E'\n\n' ||
                '{{stage_results.evaluate.output}}' || E'\n\n' ||
                'Review the evaluation against five criteria:' || E'\n\n' ||
                '1. **Source faithfulness.** Every transcript quote has a timestamp or clear identifier. Every scripture reference is correctly cited (book chapter:verse). Where the evaluation summarizes the video''s position, it is recognizable as the video''s position, not a steelman or strawman.' || E'\n\n' ||
                '2. **Discernment posture.** The evaluation respects trust-the-discernment — it surfaces questions for the Spirit-witnessing reader without pre-empting the witness itself. No final verdict in the agent''s voice on what the Spirit will say. The witness-questions section actually carries this load.' || E'\n\n' ||
                '3. **Charity + honesty balance.** Apply separate-claim-from-charisma + surface-the-rhetoric (from general-research; carries here too). If the video is rhetorically strong but substantively thin, the evaluation says so. If it is substantively strong but rhetorically off-putting, the evaluation says that. Neither charisma nor abrasiveness obscures the underlying claim.' || E'\n\n' ||
                '4. **Becoming is concrete.** The becoming section names specific changes in practice — not "be more like X" but "consider Y next Sunday school class" or "watch for Z pattern in your own teaching."' || E'\n\n' ||
                '5. **Honest uncertainty.** Where the evidence is mixed, the verdict is mixed. No "but overall" smoothing.' || E'\n\n' ||
                'Tools are DISABLED for this stage. You CANNOT re-fetch the video or re-search citations — review on the evaluation text alone.' || E'\n\n' ||
                'Return ONE of:' || E'\n' ||
                '(a) The same evaluation, verbatim and unchanged, if it passes all five criteria. Prefix with "REVIEW: passes" then a blank line then the evaluation.' || E'\n' ||
                '(b) A revised evaluation. Prefix with "REVIEW: revised" then a blank line, the revised evaluation, and at the end a brief notes section listing what changed and why.'
        )
    ),
    true,   -- sabbath_enabled (gospel evaluation is study-class work; sabbath reflection valuable)
    true,   -- atonement_enabled (cost-cap on doctrinal eval is worth atoning over)
    'study/yt/gospel/<slug>.md',
    NULL,   -- file_content_jsonpath: v1 uses whole stage output
    '["raw","researched","planned","specced","executing","verified"]'::jsonb,
    true    -- auto_materialize_on_verified: PE-final fix 2026-05-19. Required for
            -- on_maturity_verified to render file_destination AND fire promote_to_study
            -- (the PE.5 promotion path is wired inside the auto-materialize block).
            -- Without this, sabbathed verified work_items reached terminal status but
            -- never entered stewards.studies or the AGE graph.
)
ON CONFLICT (family) DO UPDATE SET
    stages                       = EXCLUDED.stages,
    sabbath_enabled              = EXCLUDED.sabbath_enabled,
    atonement_enabled            = EXCLUDED.atonement_enabled,
    file_destination_template    = EXCLUDED.file_destination_template,
    file_content_jsonpath        = EXCLUDED.file_content_jsonpath,
    maturity_ladder              = EXCLUDED.maturity_ladder,
    auto_materialize_on_verified = EXCLUDED.auto_materialize_on_verified;

-- Stage models (mirrors research-write's model assignments).
INSERT INTO stewards.stage_models (pipeline_family, stage_name, default_model, notes) VALUES
    ('yt-gospel-evaluate', 'ingest',   'kimi-k2.6',    'YouTube transcript + metadata enrichment via yt_download / yt_get. Tools enabled.'),
    ('yt-gospel-evaluate', 'evaluate', 'kimi-k2.6',    'Restoration discernment rubric + byu_citations density check + gospel_get verification. Tools enabled.'),
    ('yt-gospel-evaluate', 'review',   'qwen3.6-plus', 'Tools-disabled verification of discernment posture, source faithfulness, and concrete becoming.')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE SET
    default_model = EXCLUDED.default_model,
    notes         = EXCLUDED.notes;

-- Maturity rung mapping (mirrors research-write — synthesize/evaluate IS
-- the draft, no separate executing rung).
INSERT INTO stewards.pipeline_stage_maturity (pipeline_family, stage_name, produces_maturity, notes) VALUES
    ('yt-gospel-evaluate', 'ingest',   'researched', 'Transcript + metadata + comments captured; ready for evaluation.'),
    ('yt-gospel-evaluate', 'evaluate', 'planned',    'Draft is the plan. No separate executing rung — evaluation has no draft-vs-execute distinction.'),
    ('yt-gospel-evaluate', 'review',   'verified',   'Review pass complete; evaluation is verified.')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE SET
    produces_maturity = EXCLUDED.produces_maturity,
    notes             = EXCLUDED.notes;
