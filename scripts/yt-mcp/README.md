# YouTube MCP Server

MCP server that downloads YouTube video transcripts (via yt-dlp) and metadata, stores them locally organized by channel, and exposes search/retrieval tools to AI assistants. It can also download the **video** itself and extract **slide/keyframe screenshots** at scene changes, so a digesting agent can *see* the slides — not just read the transcript.

## Prerequisites

- Go 1.21+
- [yt-dlp](https://github.com/yt-dlp/yt-dlp) installed and on PATH — **keep it current** (`pip install -U yt-dlp`). YouTube changes its anti-bot challenge often; a months-old yt-dlp fails *video* downloads with `n challenge solving failed` even though formats list fine. (Transcript downloads are unaffected.)
- [ffmpeg](https://ffmpeg.org/) on PATH — required by `yt_download_video` (merge) and `yt_frames` (frame extraction). If your yt-dlp config has a stale `--ffmpeg-location`, `yt_download_video` overrides it with the ffmpeg it finds on PATH.

## Build

```bash
go build -o yt-mcp.exe .
```

## Run

```bash
./yt-mcp.exe serve
```

## Environment Variables

| Variable | Description |
|---|---|
| `YT_DIR` | Root directory for downloaded transcripts (default: `./yt`) |
| `YT_STUDY_DIR` | Directory for study/evaluation files (default: `./study/yt`) |
| `YT_COOKIE_FILE` | Path to cookies.txt for age-restricted content |
| `FFMPEG_PATH` | ffmpeg executable (default: `ffmpeg`) — for video download + frames |
| `YT_MAX_HEIGHT` | Cap on downloaded video resolution in px (default: `720`) |

## MCP Tools

| Tool | Description |
|------|-------------|
| `yt_download` | Download transcript and metadata for a YouTube video |
| `yt_get` | Retrieve a previously downloaded transcript (notes available slide frames) |
| `yt_list` | List downloaded videos, optionally filtered by channel |
| `yt_search` | Search across downloaded transcripts |
| `yt_download_video` | Download the **video file** (mp4, resolution-capped). Large/optional; never auto-called. |
| `yt_frames` | Extract **slide frames** via ffmpeg → `frames/*.png` + a timestamp-aligned `frames.json`. Modes: `scene` (default, one frame per slide via scene-change detection), `interval` (every N sec), `timestamps` (explicit marks). Returns the manifest, not the images. |
| `yt_slides` | **One-shot study path.** Downloads transcript + video, extracts slide frames (auto-picks **chapters** → **scene** → **interval**), aligns each slide to the narration spoken over it, and writes a readable `slides.md`. The fastest way to *study* a slide talk. |

Transcripts are stored as markdown under `{YT_DIR}/{channel}/{video_id}/` (`transcript.md`,
`cues.json`, `metadata.json`). `yt_download_video` adds `video.mp4`; `yt_frames` adds `frames/*.png`
+ `frames.json`; `yt_slides` also writes `slides.md`. The video + frames are large and stay gitignored
like the rest of `yt/`.

### Seeing the slides

`yt_frames` returns a manifest of `{ sec, file, t_link }` per frame — **not** the image bytes — so the
caller reads only the PNGs it needs. Because each frame's `sec` aligns to the transcript's `cues.json`
timestamps, a digesting agent (or a vision model) can pair **each slide with the words spoken over it**:

```
yt_download_video { url }                         # fetch the mp4 (once)
yt_frames { video_id, mode: "scene" }             # one PNG per slide + frames.json
# → read the frames you care about; align each to the transcript by timestamp
```

**Or just `yt_slides`** — the one-shot that does all of the above and writes the alignment for you:

```
yt_slides { url }                                 # transcript + video + frames + slides.md
# → read slides.md: each slide image interleaved with the narration spoken over it
```

`yt_slides` auto-picks the capture strategy: **chapter markers** in the description (the best signal —
chapters are the slide boundaries) → **scene-change** → **interval** (the fallback for smooth-scroll
screen-shares like Excalidraw, where scene-change under-fires).
