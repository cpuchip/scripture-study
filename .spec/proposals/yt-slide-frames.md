# Spec — yt-MCP: download the video + extract slide frames (and teach the substrate to *see* slides)

**Status:** SPEC (ready to build). **Date:** 2026-06-27. **Lane:** general-workspace builds Part A
(the workspace yt-MCP); pg-ai-stewards builds Part B (the substrate digester). **Origin:** Michael's
ask via the inbox — "add the ability to (a) download the video and (b) nab screenshots at time
markers so a digesting agent can *see the slides*, not just read the transcript."

## Why

Slide-heavy talks lose their densest content in a transcript-only digest: architecture diagrams,
benchmark tables, product UI, the actual SQL/code on the slide, the "as you can see here →" gestures.
Live example: the Google Cloud agentic-DB talks + the 47-video playlist being digested are mostly
slides; a caption-only digest misses most of the substance. The fix is the **rich-docs pattern
(text + page-pixels → vision)** applied to **video**: transcript + slide-frames → a vision model.

## The shared recipe (used by both Part A and Part B)

Three steps, all on tools already present (yt-dlp + ffmpeg):

1. **Download the video** (resolution-capped to keep it sane):
   `yt-dlp -f "bestvideo[height<=720]+bestaudio/best[height<=720]" --merge-output-format mp4 -o video.mp4 <url>`
2. **Extract frames** — three modes, default = scene-change (one frame per slide):
   - **scene** (default): `ffmpeg -i video.mp4 -vf "select='gt(scene,0.4)',metadata=print:file=scenes.txt" -fps_mode vfr frames/scene-%04d.png` → parse `scenes.txt` for each frame's `pts_time` → that's the slide-transition timestamp. Catches slide changes automatically; one frame per slide, not per second.
   - **interval**: `ffmpeg -i video.mp4 -vf "fps=1/N" frames/...` — one frame every N seconds (fallback for non-slide video).
   - **timestamps**: per explicit mark T, `ffmpeg -ss T -i video.mp4 -frames:v 1 frames/{T}.png` (fast seek).
3. **Write a timestamp-aligned manifest** `frames.json`:
   `[{ "sec": 137, "file": "frames/scene-0007.png", "t_link": "https://youtube.com/watch?v=ID&t=137" }, …]`
   — mirrors `cues.json`, so a consumer can interleave **each slide with the transcript text spoken over
   it** (look up the cue whose `[begin,end]` spans the frame's `sec`). That alignment *is* the value.

**Guardrails (bake in):**
- **Resolution cap** (≤720p default, configurable) — slides are legible at 720p; full-res is wasteful.
- **Dedup near-identical frames** — scene-change already gives ~one-per-slide; optionally a perceptual
  hash (dHash) drops accidental dupes (animation builds, cuts back to the same slide).
- **Cap `max_frames`** — a runaway scene threshold on a busy video shouldn't emit thousands of PNGs.
- **Return the *manifest*, not the image bytes** — the tool returns the frame list + timestamps;
  the consumer reads the specific PNGs it wants (don't flood context with every frame).
- **Storage:** everything under `yt/{channelSlug}/{videoID}/` (`video.mp4`, `frames/*.png`, `frames.json`),
  which is **gitignored** like the rest of `yt/` — big files never get committed.

## Part A — the workspace yt-MCP (`scripts/yt-mcp/`, this lane builds)

Grounded in the real code (`downloader.go`, `config.go`, `mcp.go`). Two new tools + a frames module.

**Config (`config.go`):** add `FfmpegPath` (`FFMPEG_PATH`, default `ffmpeg`) and `MaxVideoHeight`
(`YT_MAX_HEIGHT`, default 720).

**New `frames.go`** (mirrors `downloader.go`'s yt-dlp-via-`exec.Command` pattern):
- `DownloadVideoFile(cfg, rawURL, force, maxHeight, cookieOverride) (videoPath string, err error)`
  — reuses `ExtractVideoID` + `FetchMetadata` + the `yt/{ChannelSlug}/{videoID}/` layout from
  `DownloadVideo`; runs yt-dlp with the format cap; writes `video.mp4`; skips if present unless `force`.
- `ExtractFrames(cfg, videoDir, mode, opts) ([]Frame, err error)` — runs the ffmpeg recipe above into
  `videoDir/frames/`, writes `frames.json`, returns the manifest. `Frame{Sec int, File string, TLink string}`.

**New MCP tools (`mcp.go`, alongside `yt_download`/`yt_get`/`yt_list`/`yt_search`):**
- **`yt_download_video`** — `{ url, force?, max_height?=720, cookies? }` → downloads the mp4. Big/optional;
  a *separate explicit* tool (never auto-called by `yt_download`). Returns path + size + a size warning.
- **`yt_frames`** — `{ video_id | url, mode?="scene"|"interval"|"timestamps", every_sec?, timestamps?[],
  scene_threshold?=0.4, max_frames?=200 }` → extracts frames (downloads the video first if only `url`
  given and `video.mp4` absent). Returns the `frames.json` manifest (timestamps + paths + `?t=` links),
  **not** the images.
- Optional: extend **`yt_get`** to include a `frames` list (paths + timestamps) when `frames.json` exists,
  so a reader sees "transcript + available slide frames" in one call.

**Build phases:** (1) `frames.go` + config + `go test` on a known slide-heavy video; (2) wire the two
tools into `mcp.go`; (3) `frames.json` alignment helper (frame → nearest cue text) + README. ffmpeg
must be on PATH (document it next to the existing yt-dlp prereq).

## Part B — the substrate digester (`pg-ai-stewards`, their stewardship)

The substrate's **playlist/video digester** (the `WITH_YT` bridge + the doc-construction digest stages)
should **see the slides**, not just the captions. This is the **rich-docs multimodal pattern the substrate
already has** (P1–P4: text + page-pixels → `gemma-vision` via `--mmproj`) applied to video frames.

- **Get the frames into the sandbox.** Two options for the pg-ai-stewards session to choose:
  (a) the `WITH_YT` bridge runs the *same shared recipe* in-container (add ffmpeg to the bridge image;
  reuse the yt-dlp it already has) → `frames/` + `frames.json` beside the transcript; or (b) the bridge
  dials the workspace yt-MCP's `yt_frames` over MCP. (a) keeps it self-contained inside the substrate.
- **A multimodal digest stage.** Add a "read slides" step to the playlist-digester: feed the vision model
  **the slide frames + the transcript text aligned by timestamp** (use `frames.json` × `cues.json`), so the
  digest reasons over "this slide + what was said over it." Build via the existing doc-construction tool
  loop (page the frames in by handle, exactly like the rich-docs page-in — don't echo all frames at once).
- **Reuse, don't reinvent:** the substrate's doc-extract / page-pixels / vision-dispatch machinery is the
  template; the only new thing is the *source* (a video → frames) and the *alignment* (frame ↔ cue).

## Open design questions (decide while building)

- **Scene threshold default** (0.4 is a reasonable start; slide decks with subtle transitions may want
  lower; busy video higher). Make it a param + a sane default; tune on the Google-Cloud talks.
- **Frame format/size:** PNG (lossless, crisp text) vs JPEG (smaller). PNG default for slide legibility;
  cap dimensions to the capped video height.
- **How many slides is "too many"** before a digest stage should sample/cluster rather than read all.
- **Cookie/auth + age-gated/private videos** — reuse the existing `cookieArgs` path.

## Net

Part A is a clean, self-contained Go MCP enhancement (yt-dlp + ffmpeg, both already required-or-trivial),
buildable in a focused session, fully gitignored output. Part B is the substrate consuming the same
frames through its existing vision pattern. The shared recipe (capped video → scene-change frames →
timestamp-aligned manifest) is defined once and used by both, so the digester finally *sees* the slide.
