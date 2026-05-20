-- =====================================================================
-- Batch PE.4 — seed the `yt-secular-digest` pipeline
--
-- D-PE2 (ratified 2026-05-19): yt-secular-digest runs under the
-- general-research intent (now carrying the two YT-aware values from
-- PE.1). Intent is attached per work_item at create-time.
--
-- D-PE5 (ratified 2026-05-19): transcript + metadata enrichment. Ingest
-- mirrors yt-gospel-evaluate.ingest exactly — same yt tools, same brief
-- structure. The fork is at the digest stage where the rubric differs.
--
-- D-D1: sabbath + atonement ON per proposal V.6 — yt-secular-digest
-- produces a structured document worth keeping (Michael may revisit it
-- months later). Treated like study/lesson/talk-class work.
--
-- Three stages: ingest -> digest -> review. Agent family `yt` (registered,
-- has yt/* + becoming/* + study_* + playwright/* tool perms). The yt
-- agent does NOT have byu_citations or gospel_* — appropriate, secular
-- digest doesn't need scripture citation tools.
-- =====================================================================

INSERT INTO stewards.pipelines (
    family, stages, sabbath_enabled, atonement_enabled,
    file_destination_template, file_content_jsonpath, maturity_ladder,
    auto_materialize_on_verified
)
VALUES (
    'yt-secular-digest',
    jsonb_build_array(
        jsonb_build_object(
            'name',            'ingest',
            'next',            'digest',
            'model',           'kimi-k2.6',
            'provider',        'opencode_go',
            'agent_family',    'yt',
            'auto_advance',    true,
            'tools_disabled',  false,
            'input_template',
                'Binding question: {{input.binding_question}}' || E'\n\n' ||
                'Video URL: {{input.video_url}}' || E'\n\n' ||
                'Context tags (optional): {{input.context_tags}}' || E'\n\n' ||
                'You are ingesting a YouTube video for digest extraction. The next stage looks for what is worth keeping — insights, contradictions to existing notes, what to skeptically question. Quality of ingest governs quality of digest downstream.' || E'\n\n' ||
                'Use the yt tools to capture:' || E'\n' ||
                '  1. **Transcript** — yt_download(video_url) for full transcript text. Follow with yt_get if needed to retrieve the transcript itself.' || E'\n' ||
                '  2. **Metadata** — yt_get returns channel, title, publication date, duration, view count. Capture all of it.' || E'\n' ||
                '  3. **Description** — the full video description (creators often signal framing and reference other content in the description).' || E'\n' ||
                '  4. **Chapters** — if the video has chapter markers, list them.' || E'\n' ||
                '  5. **Top comments** — sample 5-10 top comments as a sample of how the video is being received.' || E'\n\n' ||
                'If a tool fails (private video, no transcript, region lock), note that as the ingest output and stop. Return "INGEST: failed — <reason>".' || E'\n\n' ||
                'Produce a structured ingest brief in markdown with sections:' || E'\n' ||
                '  - **Identity** — title, channel, URL, date, duration' || E'\n' ||
                '  - **Description** (full)' || E'\n' ||
                '  - **Chapters** (if any)' || E'\n' ||
                '  - **Transcript** (full text, with timestamps if available)' || E'\n' ||
                '  - **Top-comment sample** (5-10 comments)' || E'\n\n' ||
                'Do NOT digest yet. The next stage does that.'
        ),
        jsonb_build_object(
            'name',            'digest',
            'next',            'review',
            'model',           'kimi-k2.6',
            'provider',        'opencode_go',
            'agent_family',    'yt',
            'auto_advance',    true,
            'tools_disabled',  false,
            'input_template',
                'Binding question: {{input.binding_question}}' || E'\n\n' ||
                'Ingest brief from the previous stage:' || E'\n\n' ||
                '{{stage_results.ingest.output}}' || E'\n\n' ||
                'Extract what is worth keeping. This is a digest, not a transcript — the goal is the 5-10% of content Michael will care about in six months, with enough context to reconstruct why.' || E'\n\n' ||
                'The general-research intent applies, including the two YT-aware values: separate-claim-from-charisma (evaluate the argument, not the delivery) and surface-the-rhetoric (name the rhetorical pattern so it doesn''t substitute for evidence). Apply them.' || E'\n\n' ||
                'Use substrate-internal tools to cross-reference against existing work:' || E'\n' ||
                '  - study_search_text(query, kinds, limit) — find existing studies that touch the same topic' || E'\n' ||
                '  - study_get(slug) — read a study you found and want to compare against' || E'\n' ||
                '  - brain_search(query) — find brain entries on adjacent topics' || E'\n\n' ||
                'Structure your digest in five sections:' || E'\n\n' ||
                '1. **One-sentence summary** — what is this video, in one sentence Michael could quote back.' || E'\n\n' ||
                '2. **Key claims** — the 3-7 substantive claims the video makes. Each with a verbatim quote + timestamp from the transcript. Distinguish reported fact from creator opinion.' || E'\n\n' ||
                '3. **What''s rhetorical, what''s substantive** — name the rhetorical patterns the video uses (contrarian framing, urgency, "what they don''t tell you", insider posturing) and which substantive points survive when those patterns are stripped away.' || E'\n\n' ||
                '4. **Contradictions to existing notes** — for each key claim, search the substrate corpus and brain. If a claim contradicts existing work, name the contradiction explicitly with citations to the existing material. If a claim reinforces existing work, note that too — agreement across independent sources is also signal.' || E'\n\n' ||
                '5. **Application** — name 1-3 specific things to try, watch for, or revisit. Not "be more like X" — concrete behaviors or evaluations. "Carry-forward: nothing" is a valid section if the video doesn''t warrant follow-up.' || E'\n\n' ||
                'Length: 500-1500 words. A 90-minute talk often distills to 800 words of digest. Resist completeness in favor of carry-value.' || E'\n\n' ||
                'Produce the complete digest in markdown. The next stage reviews it.'
        ),
        jsonb_build_object(
            'name',            'review',
            'next',            NULL,
            'model',           'qwen3.6-plus',
            'provider',        'opencode_go',
            'agent_family',    'yt',
            'auto_advance',    true,
            'tools_disabled',  true,
            'input_template',
                'Binding question: {{input.binding_question}}' || E'\n\n' ||
                'The digest draft from the previous stage:' || E'\n\n' ||
                '{{stage_results.digest.output}}' || E'\n\n' ||
                'Review the digest against five criteria:' || E'\n\n' ||
                '1. **Quote fidelity.** Every quote attributed to the video has a timestamp or clear identifier. Paraphrases are labeled as paraphrases, not framed as quotes.' || E'\n\n' ||
                '2. **Separate-claim-from-charisma.** Where the video is rhetorically strong, the digest distinguishes the rhetoric from the underlying substance. Where the digest itself reads as endorsement, ask whether it is responding to the argument or to the presenter.' || E'\n\n' ||
                '3. **Cross-reference is real.** The "contradictions to existing notes" section references actual studies/brain entries with slug or title, not vague "as we''ve discussed before" gestures. If no cross-references were found, the section says so plainly rather than padding with adjacent-but-unrelated material.' || E'\n\n' ||
                '4. **Application is concrete.** Specific behaviors or evaluations, not general aspirations. "Carry-forward: nothing" is fine if the video doesn''t warrant follow-up — better than manufactured application.' || E'\n\n' ||
                '5. **Length discipline.** 500-1500 words. If the digest exceeds 1500, what can be cut without losing the carry-value?' || E'\n\n' ||
                'Tools are DISABLED for this stage — review on the digest text alone.' || E'\n\n' ||
                'Return ONE of:' || E'\n' ||
                '(a) The same digest, verbatim and unchanged, if it passes all five criteria. Prefix with "REVIEW: passes" then a blank line then the digest.' || E'\n' ||
                '(b) A revised digest. Prefix with "REVIEW: revised" then a blank line, the revised digest, and at the end a brief notes section listing what changed and why.'
        )
    ),
    true,   -- sabbath_enabled (digest worth revisiting; sabbath reflection valuable per V.6)
    true,   -- atonement_enabled (cost-cap on substantive secular eval worth atoning over)
    'study/yt/<slug>.md',
    NULL,
    '["raw","researched","planned","specced","executing","verified"]'::jsonb,
    true    -- auto_materialize_on_verified: PE-final fix 2026-05-19. See pe3 comment.
)
ON CONFLICT (family) DO UPDATE SET
    stages                       = EXCLUDED.stages,
    sabbath_enabled              = EXCLUDED.sabbath_enabled,
    atonement_enabled            = EXCLUDED.atonement_enabled,
    file_destination_template    = EXCLUDED.file_destination_template,
    file_content_jsonpath        = EXCLUDED.file_content_jsonpath,
    maturity_ladder              = EXCLUDED.maturity_ladder,
    auto_materialize_on_verified = EXCLUDED.auto_materialize_on_verified;

INSERT INTO stewards.stage_models (pipeline_family, stage_name, default_model, notes) VALUES
    ('yt-secular-digest', 'ingest', 'kimi-k2.6',    'YouTube transcript + metadata enrichment via yt_download / yt_get. Tools enabled. Identical to yt-gospel-evaluate.ingest.'),
    ('yt-secular-digest', 'digest', 'kimi-k2.6',    'Digest extraction + substrate cross-reference via study_search_text / brain_search. Tools enabled.'),
    ('yt-secular-digest', 'review', 'qwen3.6-plus', 'Tools-disabled verification of quote fidelity, cross-reference reality, and concrete application.')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE SET
    default_model = EXCLUDED.default_model,
    notes         = EXCLUDED.notes;

INSERT INTO stewards.pipeline_stage_maturity (pipeline_family, stage_name, produces_maturity, notes) VALUES
    ('yt-secular-digest', 'ingest', 'researched', 'Transcript + metadata + comments captured; ready for digest.'),
    ('yt-secular-digest', 'digest', 'planned',    'Draft is the plan. No separate executing rung — digest has no draft-vs-execute distinction.'),
    ('yt-secular-digest', 'review', 'verified',   'Review pass complete; digest is verified.')
ON CONFLICT (pipeline_family, stage_name) DO UPDATE SET
    produces_maturity = EXCLUDED.produces_maturity,
    notes             = EXCLUDED.notes;
