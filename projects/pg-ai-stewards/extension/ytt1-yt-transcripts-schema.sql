-- =====================================================================
-- Batch YT-T.3 — substrate yt_transcripts + yt_transcript_segments schema
--
-- Per substrate-yt-transcripts proposal §V.3. Two-table model:
--   yt_transcripts (one row per video; TOAST'd full_text + jsonb metadata)
--   yt_transcript_segments (FK + start/end seconds + text per cue)
--
-- D-YTT3 (ratified 2026-05-19): separate segments table with FK for
-- time-range queries. Faster than jsonb-array queries; clean per-segment
-- engram extraction path when needed.
--
-- D-YTT4: no embedding column. Engrams handle long-text vector search
-- (Batch K + L). yt_transcripts.body_tsv tsvector handles full-text
-- substring search.
-- =====================================================================

-- ---------------------------------------------------------------------
-- YT-T.3.a — yt_transcripts (one row per video)
-- ---------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS stewards.yt_transcripts (
    video_id          text PRIMARY KEY,
    channel_slug      text NOT NULL,
    title             text NOT NULL DEFAULT '',
    duration_seconds  int,
    published_at      timestamptz,
    full_text         text NOT NULL DEFAULT '',
    metadata          jsonb NOT NULL DEFAULT '{}'::jsonb,
    imported_at       timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now(),
    body_tsv          tsvector GENERATED ALWAYS AS (
        to_tsvector('english', coalesce(title, '') || ' ' || coalesce(full_text, ''))
    ) STORED
);

CREATE INDEX IF NOT EXISTS yt_transcripts_channel_idx
    ON stewards.yt_transcripts (channel_slug);

CREATE INDEX IF NOT EXISTS yt_transcripts_fts_idx
    ON stewards.yt_transcripts USING gin (body_tsv);

CREATE INDEX IF NOT EXISTS yt_transcripts_metadata_idx
    ON stewards.yt_transcripts USING gin (metadata);

COMMENT ON TABLE stewards.yt_transcripts IS
'YT-T.3: one row per YouTube video the substrate has ingested. full_text auto-TOASTs for long transcripts (Morgan Philpot 3.5h video ~ 50-80k chars). Populated by stewards.import_yt_transcript() from the workspace yt/<channel>/<video_id>/ cache.';

COMMENT ON COLUMN stewards.yt_transcripts.channel_slug IS
'yt-mcp slugified channel name, also the parent directory under /opt/yt/yt/. E.g. "ai-news-strategy-daily-nate-b-jones".';

COMMENT ON COLUMN stewards.yt_transcripts.metadata IS
'yt-dlp --dump-json output, plus any other ingest-time facts. Indexed via GIN for jsonb-path queries.';

-- ---------------------------------------------------------------------
-- YT-T.3.b — yt_transcript_segments (one row per cue)
-- ---------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS stewards.yt_transcript_segments (
    video_id      text NOT NULL REFERENCES stewards.yt_transcripts(video_id) ON DELETE CASCADE,
    segment_idx   int  NOT NULL,
    start_seconds real NOT NULL,
    end_seconds   real NOT NULL,
    text          text NOT NULL,
    PRIMARY KEY (video_id, segment_idx)
);

CREATE INDEX IF NOT EXISTS yt_transcript_segments_time_idx
    ON stewards.yt_transcript_segments (video_id, start_seconds);

COMMENT ON TABLE stewards.yt_transcript_segments IS
'YT-T.3: time-coded segments for each ingested video. Sourced from the cues.json yt-mcp writes alongside transcript.md. Enables time-range queries: "what was said in 9UTrPgjLW7g between minutes 23 and 30".';
