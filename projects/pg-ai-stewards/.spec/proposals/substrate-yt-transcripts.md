---
title: Substrate YT transcripts — yt-dlp in bridge + shared cache + native transcript table
date: 2026-05-19
status: ratified — building
parent: open-items.md (post-council-① new batch; wedges before council ②)
purpose: >
  Unblock substrate-side YouTube ingestion. PE-final's yt-gospel-evaluate
  hit a real wall — the bridge container has no yt-dlp and the workspace
  yt/ folder is read-only-mounted, so the agent correctly refused to
  fabricate (good covenant signal, but blocking pipeline runs). This
  batch adds yt-dlp to the bridge, mounts the workspace cache rw, and
  introduces native substrate primitives (yt_transcripts +
  yt_transcript_segments) so long video content participates in the
  substrate the same way studies do.
---

# Substrate YT transcripts

## I. Binding problem

PE-final smoked all three new pipelines end-to-end (council ①, 2026-05-19). yt-gospel-evaluate against Morgan Philpot's `9UTrPgjLW7g` completed but produced a refusal-not-evaluation because:

1. **Bridge has no `yt-dlp`.** `apk add` in bridge.Dockerfile installs ca-certs, tzdata, git, gh, chromium — not yt-dlp.
2. **Workspace `yt/` mounted read-only.** Bridge can READ the existing transcript cache via `/workspace`, but cannot write new downloads back.
3. **No native substrate representation.** Even when transcripts exist on disk, the substrate's `study_search_text` / AGE graph / engrams don't know about YouTube content as a primary kind. Studies that cite YouTube videos point at URIs but the content itself isn't indexed.

The agent honored `read-before-quoting` and refused. Right answer; wrong floor. This batch raises the floor.

## II. Success criteria

1. Re-fire the same Morgan Philpot work_item; this time it produces a real five-section evaluation with verbatim quotes + timestamps + scripture citations (when present).
2. yt-dlp inside the bridge can fetch any public YouTube URL and write the transcript to `/opt/yt/yt/<channel>/<video_id>/` — which is the workspace `yt/` directory, immediately visible to host tooling + git.
3. `stewards.yt_transcripts` + `stewards.yt_transcript_segments` exist and survive container restart. A new substrate function `import_yt_transcript(video_id)` reads the workspace cache and populates them.
4. yt_transcripts is queryable from inside the substrate the same way `studies` is — agent tools can search by video_id, by channel, by text content, by time range.

## III. Constraints and boundaries

**In scope:**
- Bridge Dockerfile: install yt-dlp via apk (alpine 3.18+ has it in community repo).
- docker-compose: workspace `yt/` mounted rw at `/opt/yt/yt`.
- Substrate schema: `yt_transcripts` (one row per video) + `yt_transcript_segments` (FK + time-range per row).
- Substrate function: `import_yt_transcript(video_id)` parses the workspace cache files and populates both tables.
- Agent tool wrapper exposed via stewards-mcp.
- Smoke: re-fire Morgan Philpot evaluation end-to-end.

**Out of scope (deliberately):**
- Auto-ingest hook (yt_download → automatic import_yt_transcript). Defer. Agent calls explicitly for now.
- Modifying yt-gospel/yt-secular pipeline ingest prompts to add the new tool call. Defer. The re-fire happens manually for the smoke; pipeline prompts adapt in a follow-up once the floor is verified.
- pgvector embeddings on transcript content. Per D-YTT4, reuse engrams pattern. Long transcripts feeding the existing K/L engram extraction is sufficient.
- AGE graph nodes for YouTube videos. Could come naturally if `parse_gospel_links` extends to YouTube URIs; that's a separate spec.
- Cookie / authenticated download handling (yt-mcp already has the cookieArgs path; just works if YT_COOKIE_FILE is set).

## IV. Prior art

- **bridge.Dockerfile** — already creates `/opt/yt/yt` and `/opt/yt/study`, anticipating exactly this flip.
- **yt-mcp** (`scripts/yt-mcp/`) — already wraps yt-dlp via `config.go:YtDlpPath` (defaults `"yt-dlp"`, expects it in `$PATH`). Already writes to `cfg.YTDir + "/<channel>/<video_id>/"` (default cwd; here = `/opt/yt/yt`).
- **stewards.import_study** — model for parsing+ingestion+graph-write. import_yt_transcript follows the same shape: read file → upsert table row → optionally graph node → optionally edges.
- **engrams (Batch K)** — long content's existing path through the substrate. yt_transcripts complements engrams, doesn't compete.
- **studies table TOAST** — the substrate already stores 100k+ char study bodies via PostgreSQL's automatic TOAST. yt_transcripts.full_text uses the same mechanism.

## V. Proposed approach

### V.1 Bridge install (YT-T.1)

```dockerfile
RUN apk add --no-cache ca-certificates tzdata git github-cli chromium yt-dlp
```

One added package. Triggers a bridge image rebuild (`docker compose build bridge`). Verified by `docker exec pg-ai-stewards-bridge yt-dlp --version`.

### V.2 Workspace volume mount (YT-T.2)

```yaml
# bridge service in projects/pg-ai-stewards/extension/docker-compose.yaml
volumes:
  - ../../..:/workspace:ro
  - ../../../yt:/opt/yt/yt:rw          # NEW: workspace yt/ is THE cache
```

Bridge yt-mcp writes downloads to `/opt/yt/yt/<channel>/<video_id>/`. This path is the workspace `yt/` folder on the host. The 1.3GB of existing transcripts is immediately visible inside the bridge.

### V.3 Substrate schema (YT-T.3)

```sql
CREATE TABLE IF NOT EXISTS stewards.yt_transcripts (
    video_id          text PRIMARY KEY,           -- YouTube video ID, e.g. '9UTrPgjLW7g'
    channel_slug      text NOT NULL,              -- yt-mcp's channel slug (e.g. 'morganphilpot')
    title             text NOT NULL,
    duration_seconds  int,                        -- NULL if unknown
    published_at      timestamptz,
    full_text         text NOT NULL DEFAULT '',   -- TOAST'd transcript body
    metadata          jsonb NOT NULL DEFAULT '{}',-- yt-dlp info.json (url, uploader, etc.)
    imported_at       timestamptz NOT NULL DEFAULT now(),
    body_tsv          tsvector GENERATED ALWAYS AS (to_tsvector('english', full_text)) STORED
);

CREATE INDEX IF NOT EXISTS yt_transcripts_channel_idx ON stewards.yt_transcripts (channel_slug);
CREATE INDEX IF NOT EXISTS yt_transcripts_fts_idx ON stewards.yt_transcripts USING gin (body_tsv);
CREATE INDEX IF NOT EXISTS yt_transcripts_metadata_idx ON stewards.yt_transcripts USING gin (metadata);

CREATE TABLE IF NOT EXISTS stewards.yt_transcript_segments (
    video_id        text NOT NULL REFERENCES stewards.yt_transcripts(video_id) ON DELETE CASCADE,
    segment_idx     int  NOT NULL,
    start_seconds   real NOT NULL,
    end_seconds     real NOT NULL,
    text            text NOT NULL,
    PRIMARY KEY (video_id, segment_idx)
);

CREATE INDEX IF NOT EXISTS yt_transcript_segments_video_idx ON stewards.yt_transcript_segments (video_id);
CREATE INDEX IF NOT EXISTS yt_transcript_segments_time_idx ON stewards.yt_transcript_segments (video_id, start_seconds);
```

Time-range queries like *"what was said between minute 23 and 30 of `9UTrPgjLW7g`"* become a single indexed lookup. Per D-YTT4 no embedding column — engrams handle that path.

### V.4 import_yt_transcript() (YT-T.4)

```sql
stewards.import_yt_transcript(p_video_id text) RETURNS text
```

Strategy:
1. Glob `/opt/yt/yt/*/p_video_id/` to find the channel directory (channel slug isn't always known by caller).
2. Read `metadata.json` (yt-mcp's metadata format) — title, channel, duration, published_at, full yt-dlp info as jsonb.
3. UPSERT into `yt_transcripts` (`ON CONFLICT (video_id) DO UPDATE`).
4. Read either `transcript.md` (plain text) or the original `.srt` (subtitle file) — yt-mcp keeps SRT in `tmp/` post-process but the parsed transcript is in `transcript.md`. We need to look for whichever format preserves timestamps. Strategy: if `transcript.md` has timestamp markers we can parse, use those; otherwise fall back to a single segment with `start_seconds=0, end_seconds=duration_seconds`.
5. DELETE existing segments for this video_id (clean re-import), then INSERT new ones.
6. RETURN the video_id (or NULL on failure with NOTICE).

Tool wrapper: a thin `import_yt_transcript_tool` exposed in the mcp_servers config so agents can call it the same way they call `study_get_tool`.

### V.5 Smoke and re-fire (YT-T.5)

1. After bridge rebuild + restart: `docker exec pg-ai-stewards-bridge yt-dlp --version` returns a version string.
2. Touch test for rw mount: `docker exec pg-ai-stewards-bridge sh -c "touch /opt/yt/yt/.rw-test && ls /opt/yt/yt/.rw-test"` succeeds; the file appears in workspace `yt/.rw-test` on the host. Cleanup the file.
3. Apply YT-T.3 + YT-T.4 SQL via docker cp + psql.
4. Pre-existing cache smoke: `SELECT stewards.import_yt_transcript('lqiwQiDglGk')` (Nate B Jones Pinecone video). Verify yt_transcripts row populated + segments inserted.
5. Fresh-download smoke: `docker exec pg-ai-stewards-bridge yt-dlp -o '/opt/yt/yt/morganphilpot/9UTrPgjLW7g/transcript' --write-auto-subs --skip-download --sub-format vtt 'https://youtu.be/9UTrPgjLW7g'` — or just let yt-mcp's `yt_download` tool do it. Verify files land in workspace `yt/morganphilpot/9UTrPgjLW7g/`.
6. `SELECT stewards.import_yt_transcript('9UTrPgjLW7g')` — verify substrate ingested.
7. Re-fire the yt-gospel-evaluate work_item with the SAME binding question (new slug `pe-final-yt-gospel-morgan-philpot-rerun`). Verify this time the evaluation has real content + verbatim quotes + timestamps.
8. Verify landed in `stewards.studies` + AGE graph as `kind='gospel-evaluation'`.

## VI. Decisions ratified 2026-05-19

| ID | Choice |
|----|--------|
| D-YTT1 | Mount workspace `yt/` rw at `/opt/yt/yt` — shared cache |
| D-YTT2 | Build now in this batch (no defer) |
| D-YTT3 | Separate `yt_transcript_segments` table with FK |
| D-YTT4 | Reuse engrams pattern — no new pgvector column |

## VII. Acceptance scenarios

1. `docker exec pg-ai-stewards-bridge yt-dlp --version` returns `2025.xx.xx` or similar.
2. New download via yt-mcp from inside bridge writes to workspace `yt/<channel>/<video_id>/` — visible immediately to host.
3. `SELECT * FROM stewards.yt_transcripts WHERE video_id = '9UTrPgjLW7g'` returns a row with the right channel + title + duration.
4. `SELECT count(*) FROM stewards.yt_transcript_segments WHERE video_id = '9UTrPgjLW7g'` returns the actual segment count (Morgan Philpot 3.5h video → probably 500+ segments depending on subtitle granularity).
5. Re-fired yt-gospel-evaluate work_item completes with `kind='gospel-evaluation'` body containing real Philpot quotes with timestamps, not a refusal.

## VIII. Carry-forward

- **Auto-ingest hook** — when bridge's yt-mcp finishes a `yt_download`, automatically call `import_yt_transcript`. Currently manual. Worth a follow-on once the floor is verified.
- **Pipeline ingest prompts** — yt-gospel-evaluate + yt-secular-digest ingest stages could be updated to call the new tool right after `yt_download`. Currently neither does.
- **Embedding population** — when a transcript is imported, if a downstream agent reads it (engrams pattern), embeddings happen naturally. But "find me all video segments that talk about X across the whole corpus" is not yet possible without an explicit engram extraction. If that becomes desired, revisit D-YTT4.
- **AGE Reference nodes** — `parse_gospel_links` could be extended to recognize YouTube URIs (`youtube.com/watch?v=` + `youtu.be/`) and create a typed `:YTTranscript` node + `:CITES` edge from studies that mention them.
