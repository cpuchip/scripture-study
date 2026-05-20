-- =====================================================================
-- Batch YT-T.4 — stewards.import_yt_transcript(video_id) function
--
-- Reads /opt/yt/yt/<channel>/<video_id>/{metadata.json, cues.json,
-- transcript.md} from the workspace yt/ cache (mounted ro into the pg
-- container by docker-compose), parses them, and upserts into
-- stewards.yt_transcripts + stewards.yt_transcript_segments.
--
-- Channel slug is discovered by scanning /opt/yt/yt/* for whichever
-- directory contains the video_id subdir — caller doesn't need to know.
--
-- Idempotent: ON CONFLICT DO UPDATE on yt_transcripts, DELETE+INSERT
-- on segments. Re-running with the same video_id re-reads + replaces.
-- =====================================================================

CREATE OR REPLACE FUNCTION stewards.import_yt_transcript(p_video_id text)
RETURNS text
LANGUAGE plpgsql
AS $func$
DECLARE
    v_root            text := '/opt/yt/yt';
    v_channel_slug    text;
    v_dir             text;
    v_metadata_path   text;
    v_cues_path       text;
    v_transcript_path text;
    v_metadata_json   jsonb;
    v_cues_json       jsonb;
    v_title           text;
    v_duration        int;
    v_upload_date     text;
    v_published_at    timestamptz;
    v_full_text       text;
    v_segments_count  int := 0;
    v_channel_dirs    text[];
    v_chan            text;
BEGIN
    IF p_video_id IS NULL OR length(p_video_id) = 0 THEN
        RAISE EXCEPTION 'import_yt_transcript: video_id required';
    END IF;

    -- Discover channel_slug by listing /opt/yt/yt/* and finding the
    -- one whose subdir matches p_video_id. ~25 channels in the workspace
    -- today; this scan is instant.
    SELECT array_agg(d) INTO v_channel_dirs FROM pg_ls_dir(v_root) d;
    IF v_channel_dirs IS NULL THEN
        RAISE NOTICE 'import_yt_transcript: %s not readable from pg container', v_root;
        RETURN NULL;
    END IF;

    FOREACH v_chan IN ARRAY v_channel_dirs LOOP
        BEGIN
            IF EXISTS (
                SELECT 1 FROM pg_ls_dir(v_root || '/' || v_chan || '/' || p_video_id) LIMIT 1
            ) THEN
                v_channel_slug := v_chan;
                EXIT;
            END IF;
        EXCEPTION WHEN OTHERS THEN
            -- pg_ls_dir raises if the path doesn't exist or isn't a directory;
            -- treat as "not this channel" and keep scanning.
            CONTINUE;
        END;
    END LOOP;

    IF v_channel_slug IS NULL THEN
        RAISE NOTICE 'import_yt_transcript: video_id % not found under % (have you yt_download''d it?)',
            p_video_id, v_root;
        RETURN NULL;
    END IF;

    v_dir             := v_root || '/' || v_channel_slug || '/' || p_video_id;
    v_metadata_path   := v_dir || '/metadata.json';
    v_cues_path       := v_dir || '/cues.json';
    v_transcript_path := v_dir || '/transcript.md';

    -- ---------------------------------------------------------------
    -- Read metadata.json → jsonb. Required.
    -- ---------------------------------------------------------------
    BEGIN
        v_metadata_json := pg_read_file(v_metadata_path)::jsonb;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'import_yt_transcript: failed to read %: %', v_metadata_path, SQLERRM;
        RETURN NULL;
    END;

    v_title       := coalesce(v_metadata_json->>'title', p_video_id);
    v_duration    := nullif(v_metadata_json->>'duration', '')::int;
    v_upload_date := v_metadata_json->>'upload_date';
    -- yt-dlp upload_date is YYYYMMDD. Convert to timestamptz at midnight UTC.
    IF v_upload_date IS NOT NULL AND length(v_upload_date) = 8 THEN
        v_published_at := to_timestamp(v_upload_date, 'YYYYMMDD') AT TIME ZONE 'UTC';
    END IF;

    -- ---------------------------------------------------------------
    -- Read cues.json → jsonb array of {begin, end, text}. Optional;
    -- fall back to single segment if absent.
    -- ---------------------------------------------------------------
    BEGIN
        v_cues_json := pg_read_file(v_cues_path)::jsonb;
    EXCEPTION WHEN OTHERS THEN
        v_cues_json := NULL;
    END;

    -- ---------------------------------------------------------------
    -- Read transcript.md → text for full_text. If absent, derive
    -- full_text from cues. If both absent, full_text stays empty.
    -- ---------------------------------------------------------------
    BEGIN
        v_full_text := pg_read_file(v_transcript_path);
    EXCEPTION WHEN OTHERS THEN
        IF v_cues_json IS NOT NULL THEN
            SELECT string_agg(c->>'text', ' ' ORDER BY ord)
              INTO v_full_text
              FROM jsonb_array_elements(v_cues_json) WITH ORDINALITY t(c, ord);
        ELSE
            v_full_text := '';
        END IF;
    END;

    -- ---------------------------------------------------------------
    -- UPSERT yt_transcripts. full_text auto-TOASTs over 2KB.
    -- ---------------------------------------------------------------
    INSERT INTO stewards.yt_transcripts (
        video_id, channel_slug, title, duration_seconds, published_at,
        full_text, metadata, imported_at, updated_at
    ) VALUES (
        p_video_id, v_channel_slug, v_title, v_duration, v_published_at,
        coalesce(v_full_text, ''), coalesce(v_metadata_json, '{}'::jsonb), now(), now()
    )
    ON CONFLICT (video_id) DO UPDATE SET
        channel_slug     = EXCLUDED.channel_slug,
        title            = EXCLUDED.title,
        duration_seconds = EXCLUDED.duration_seconds,
        published_at     = EXCLUDED.published_at,
        full_text        = EXCLUDED.full_text,
        metadata         = EXCLUDED.metadata,
        updated_at       = now();

    -- ---------------------------------------------------------------
    -- DELETE + re-INSERT segments. Idempotent.
    -- ---------------------------------------------------------------
    DELETE FROM stewards.yt_transcript_segments WHERE video_id = p_video_id;

    IF v_cues_json IS NOT NULL AND jsonb_typeof(v_cues_json) = 'array' THEN
        INSERT INTO stewards.yt_transcript_segments
            (video_id, segment_idx, start_seconds, end_seconds, text)
        SELECT
            p_video_id,
            (ord - 1)::int,
            coalesce((c->>'begin')::real, 0),
            coalesce((c->>'end')::real, 0),
            coalesce(c->>'text', '')
          FROM jsonb_array_elements(v_cues_json) WITH ORDINALITY t(c, ord);
        GET DIAGNOSTICS v_segments_count = ROW_COUNT;
    ELSIF v_duration IS NOT NULL AND length(coalesce(v_full_text,'')) > 0 THEN
        -- No cues but we have duration + text → single whole-video segment.
        INSERT INTO stewards.yt_transcript_segments
            (video_id, segment_idx, start_seconds, end_seconds, text)
        VALUES (p_video_id, 0, 0, v_duration, v_full_text);
        v_segments_count := 1;
    END IF;

    RAISE NOTICE 'import_yt_transcript: % ingested (channel=%, segments=%, full_text_chars=%)',
        p_video_id, v_channel_slug, v_segments_count, length(coalesce(v_full_text,''));
    RETURN p_video_id;
END;
$func$;

COMMENT ON FUNCTION stewards.import_yt_transcript(text) IS
'YT-T.4: read /opt/yt/yt/<channel>/<video_id>/{metadata.json, cues.json, transcript.md} (workspace yt/ via the pg ro mount) and upsert stewards.yt_transcripts + yt_transcript_segments. Idempotent. Returns video_id on success, NULL on cache-miss (with NOTICE). Channel slug auto-discovered by scanning pg_ls_dir.';
