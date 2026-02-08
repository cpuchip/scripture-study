# yt-mcp: YouTube Transcript Downloader & Gospel Evaluator

*Created: February 7, 2026*

---

## Vision

A tool to download YouTube video transcripts and evaluate them against scripture and general conference. Works for gospel-centered content (BYU devotionals, seminary, podcast interviews) *and* secular content (news, commentary, lectures). The goal: extract what's good, identify what's off, find what's missing, and surface personal "becoming" commitments — all grounded in the standard works.

The 5-step workflow:
1. **Download** — get the transcript (and metadata) from a YouTube URL
2. **Review** — identify doctrines, principles, scripture/conference references
3. **Cross-reference** — search gospel-mcp and gospel-vec for supporting or contradicting citations
4. **Evaluate** — honest critique against the scriptural standard
5. **Become** — extract personal commitments and action items

---

## Architecture Decision: MCP Server + Copilot Instructions

### What the MCP server does (tools)
- `yt_download` — download transcript + metadata from YouTube URL via yt-dlp
- `yt_list` — list downloaded transcripts (browse `./yt/` directory)
- `yt_get` — retrieve a downloaded transcript's full text + metadata
- `yt_search` — search across downloaded transcripts (keyword/phrase)

### What copilot-instructions handle (workflow)
Steps 2–5 (review, cross-reference, evaluate, become) are *reasoning tasks*, not tool tasks. They use existing tools (gospel-mcp, gospel-vec, webster-mcp, read_file) guided by a new workflow section in copilot-instructions.md. This keeps the MCP server focused and avoids duplicating search logic.

### Why not bake the analysis into the MCP server?
- The AI's reasoning over full context is better than anything we'd hardcode
- gospel-mcp and gospel-vec already exist for cross-referencing
- The evaluation criteria ("in line / out of line / missed something") require judgment, not retrieval
- Keeping the tool simple means it's useful beyond just scripture study

---

## File Layout

```
scripture-study/
├── yt/                                    # Downloaded transcripts (.gitignored — copyrighted material)
│   ├── {channel_slug}/                    # e.g., "book-of-mormon-central"
│   │   ├── {video_id}/                    # e.g., "dQw4w9WgXcQ"
│   │   │   ├── metadata.json             # Title, channel, date, duration, URL, description
│   │   │   ├── cues.json                 # Raw timestamped cues [{begin, end, text}, ...]
│   │   │   └── transcript.md             # Cleaned transcript with clickable timestamp links
│   │   └── ...
│   └── ...
├── study/
│   └── yt/                                # Video evaluation study docs (output of step 4-5)
│       └── {video_id}-{slug}.md          # e.g., "abc123-enoch-and-zion.md"
└── scripts/
    └── yt-mcp/
        ├── main.go                        # CLI entry point + MCP server
        ├── mcp.go                         # MCP JSON-RPC handler (stdin/stdout)
        ├── downloader.go                  # yt-dlp wrapper
        ├── transcript.go                  # TTML → markdown converter (cue merge + timestamp links)
        ├── config.go                      # Configuration
        ├── types.go                       # Data models
        ├── go.mod
        └── docs/
            └── 01_TODO.md                 # This file
```

> **Note:** `yt/` is `.gitignored` — it contains copyrighted material (same policy as `gospel-library/`). The yt-mcp tool itself is the reproducible mechanism: anyone can clone the repo and download transcripts themselves.

### Why `{channel_slug}/{video_id}/` nesting?
- Channel slug groups by source (easy to browse "all BYU devotionals")
- Video ID is the unique key (avoids title collision, enables YouTube URL reconstruction)
- `metadata.json` separates structured data from readable transcript
- `cues.json` preserves raw timestamped cues for precise linking back to the video
- `transcript.md` is the human-readable output with clickable `?t=` links, optimized for both AI and human use

### Transcript Format: Markdown with Clickable Timestamps

The forkirk project stores raw TTML, which is great for timestamp-based quoting but terrible for AI reading. We convert to clean markdown **with inline timestamp links** so you can click and jump straight to that point in the video:

```markdown
# Never Gonna Give You Up

**Channel:** Rick Astley
**Date:** 2009-10-25
**Duration:** 3:33
**URL:** https://www.youtube.com/watch?v=dQw4w9WgXcQ

---

## Transcript

[0:00](https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=0) We're no strangers to love. You know the rules and so do I. A full
commitment's what I'm thinking of. You wouldn't get this from any
other guy.

[0:22](https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=22) I just wanna tell you how I'm feeling. Gotta make you understand.

[0:30](https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=30) Never gonna give you up. Never gonna let you down. Never gonna run
around and desert you...
```

Key decisions:
- **Merge cues into paragraphs** — TTML gives one cue per ~3 words. We merge adjacent cues, inserting paragraph breaks at natural pauses (gaps > 2 seconds or sentence boundaries)
- **Each paragraph starts with a clickable timestamp** — `[M:SS](youtube_url&t=seconds)` links directly to that moment in the video. Click in VS Code preview → opens YouTube at the right spot.
- **Strip duplicate lines** — YouTube auto-subs often repeat lines across cue boundaries
- **Normalize whitespace** — clean up auto-caption artifacts
- **Raw cues preserved in `cues.json`** — for fine-grained timestamp lookups when quoting specific lines
- **Keep it short** — the transcript should fit comfortably in an AI context window

### Timestamp Linking in Study Documents

When quoting a video in an evaluation or study doc, use the same `?t=` pattern:

```markdown
> "You wouldn't get this from any other guy."
> — [Rick Astley, 0:18](https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=18)
```

The `cues.json` file enables precise timestamp lookup. The AI (or a future tool) can find the exact `?t=` value for any quoted text by searching the cue array. This is the same principle as our gospel-library links — click to verify the source.

---

## Tool Specifications

### 1. `yt_download`

Download a YouTube video's transcript and metadata.

```json
{
  "name": "yt_download",
  "description": "Download the English transcript and metadata from a YouTube video using yt-dlp. Saves to ./yt/{channel}/{video_id}/. Returns the transcript text and metadata. Requires yt-dlp to be installed and in PATH.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "url": {
        "type": "string",
        "description": "YouTube video URL (e.g., 'https://www.youtube.com/watch?v=...' or 'https://youtu.be/...')"
      },
      "force": {
        "type": "boolean",
        "description": "Re-download even if transcript already exists locally. Default: false"
      }
    },
    "required": ["url"]
  }
}
```

**Implementation:**

1. Extract video ID from URL (handle youtube.com/watch?v=, youtu.be/, youtube.com/shorts/, etc.)
2. Run `yt-dlp --dump-json --skip-download {url}` to get metadata
3. Extract: `id`, `title`, `channel`, `channel_id`, `upload_date`, `duration`, `description`
4. Generate channel slug from channel name (lowercase, hyphens, strip special chars)
5. Check if `./yt/{channel_slug}/{video_id}/transcript.md` exists — skip if present and `force` is false
6. Run `yt-dlp --write-subs --write-auto-subs --sub-langs "en.*,en" --sub-format ttml --skip-download -o {temp_path} {url}`
7. Parse TTML file → extract raw cues
8. Deduplicate and merge cues into paragraphs (with paragraph-level timestamps)
9. Write `metadata.json`, `cues.json` (raw cues), and `transcript.md` (with `[M:SS](url&t=N)` links)
10. Clean up temp TTML file
11. Return transcript content + metadata summary as MCP response

**yt-dlp flags explained:**
- `--write-subs` — prefer manual/official subtitles if available
- `--write-auto-subs` — fall back to auto-generated if no manual subs
- `--sub-langs "en.*,en"` — English variants (en, en-US, en-GB, etc.)
- `--sub-format ttml` — structured format we can parse (reuse forkirk's `ParseTTMLFile`)
- `--skip-download` — no video file, just subtitles + metadata

**Edge cases:**
- No English subs available → return error with helpful message
- Live stream / premiere → may not have subs yet
- Playlist URL → only process the single video (reject playlists for now)
- Age-restricted → may require cookies (document this)

### 2. `yt_get`

Retrieve a previously downloaded transcript.

```json
{
  "name": "yt_get",
  "description": "Get the full transcript and metadata of a previously downloaded YouTube video. Use after yt_download or yt_list to read the content.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "video_id": {
        "type": "string",
        "description": "YouTube video ID (e.g., 'dQw4w9WgXcQ')"
      },
      "path": {
        "type": "string",
        "description": "Direct path to the transcript directory, if known"
      }
    }
  }
}
```

**Implementation:**
- If `video_id` provided: scan `./yt/*/` for a directory matching the ID
- If `path` provided: read directly
- Return `transcript.md` content + parsed `metadata.json`

### 3. `yt_list`

Browse downloaded transcripts.

```json
{
  "name": "yt_list",
  "description": "List downloaded YouTube transcripts. Can filter by channel. Shows title, date, channel, and video ID for each.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "channel": {
        "type": "string",
        "description": "Filter by channel slug (e.g., 'book-of-mormon-central')"
      },
      "limit": {
        "type": "integer",
        "description": "Maximum results to return (default: 20)"
      }
    }
  }
}
```

**Implementation:**
- Walk `./yt/` directory tree
- Read `metadata.json` from each `{video_id}/` folder
- Sort by date (newest first)
- Return formatted list with title, channel, date, video ID

### 4. `yt_search`

Search across downloaded transcripts.

```json
{
  "name": "yt_search",
  "description": "Search across all downloaded YouTube transcripts for a keyword or phrase. Returns matching excerpts with video context.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "query": {
        "type": "string",
        "description": "Text to search for in transcripts"
      },
      "channel": {
        "type": "string",
        "description": "Filter by channel slug"
      },
      "limit": {
        "type": "integer",
        "description": "Maximum results (default: 10)"
      }
    },
    "required": ["query"]
  }
}
```

**Implementation:**
- Walk `./yt/` directory tree, read each `transcript.md`
- Case-insensitive substring search
- Return matching paragraphs with surrounding context
- Include video metadata (title, channel, date, URL) for each match
- For each match, look up the closest timestamp from `cues.json` and include a clickable `?t=` link

---

## TTML → Markdown Conversion

### Code to Lift from forkirk

The `quoter/ttml.go` file in `../forkirk` has a working TTML parser:
- `ParseTTMLFile(path string) ([]Cue, error)` — parses `<p begin="..." end="...">text</p>` elements
- `parseClockTime(s string) float64` — handles `HH:MM:SS.mmm` and `SS.mmms` formats
- `readInnerText(dec, endLocal)` — extracts text from nested XML elements
- `normalizeWS(s string)` — collapses whitespace

**What we need to add:**
- **Cue merging** — collapse sequential cues into paragraphs based on time gaps
- **Deduplication** — YouTube auto-subs often repeat the same text across overlapping cues
- **Paragraph detection** — insert breaks at natural pauses (>2s gap) or sentence endings
- **Markdown generation** — format the merged text with header, metadata, and clean paragraphs

### Merge Algorithm

```
Input:  [Cue{0.0, 2.5, "We're no"}, Cue{2.5, 5.0, "strangers to love"}, Cue{5.0, 8.0, "you know the rules"}, ...]
Output: Paragraph{Begin: 0.0, Text: "We're no strangers to love. You know the rules..."}

Rules:
1. If next cue starts within 1.5s of previous cue end → same paragraph
2. If gap > 2.0s → new paragraph
3. If text is identical to previous cue → skip (dedup)
4. If text is a prefix of the next cue → skip (YouTube rolling-caption dedup)
5. Sentence-ending punctuation (. ? !) at a gap > 1.0s → new paragraph
6. Each paragraph records the Begin time of its first cue → used for the [M:SS](url&t=N) link
```

### cues.json Format

The raw cues are preserved for fine-grained quoting:

```json
[
  {"begin": 0.0, "end": 2.5, "text": "We're no strangers to love"},
  {"begin": 2.5, "end": 5.0, "text": "you know the rules and so do I"},
  ...
]
```

When the AI quotes a specific line from a transcript, it can scan `cues.json` to find the cue whose text contains the quoted phrase and build a `?t=` link from `floor(begin)`. This gives sub-paragraph precision for citations.

---

## Configuration

```go
type Config struct {
    YTDir       string // Base directory for downloads (default: "./yt")
    YtDlpPath   string // Path to yt-dlp executable (default: "yt-dlp" in PATH)
    StudyDir    string // Where to write evaluation docs (default: "./study/yt")
}
```

- Config from environment variables: `YT_DIR`, `YT_DLP_PATH`, `YT_STUDY_DIR`
- Fallback to sensible defaults relative to workspace root

---

## Copilot Instructions Addition

Add a new workflow section to `.github/copilot-instructions.md`:

```markdown
### Video Evaluation Workflow

For evaluating YouTube videos against the gospel standard:

**Step 1 — Download** (use yt-mcp):
- `yt_download` the transcript from a YouTube URL
- Review the transcript with `yt_get` if needed

**Step 2 — Review** (AI reasoning):
- Read the full transcript
- Identify claims, doctrines, principles, and any scripture/conference references made
- Note the speaker's main thesis or message
- Flag any specific quotes worth examining

**Step 3 — Cross-Reference** (use gospel-mcp, gospel-vec, webster-mcp):
- For each doctrine or principle claimed, search for supporting scriptures
- For messages that aren't direct quotes, find the closest scriptural parallel
- Check conference talks for prophetic statements on the same topics
- Use Webster 1828 if historical word meanings are relevant

**Step 4 — Evaluate** (AI reasoning):
Write an honest assessment:
- **In line:** Messages that align with scripture and prophetic teaching — cite the supporting references
- **Out of line:** Claims that contradict or distort scriptural truth — explain why with references
- **Missed the mark:** Messages that are partially true but miss key context — show what's missing
- **Missed opportunities:** Great points where a powerful scripture or talk would have strengthened the message
- **Overall assessment:** Is this content spiritually nourishing? Would you recommend it?

**Step 5 — Become** (personal application):
- What truth from this video can I apply in my life?
- What warning should I heed?
- Write specific "I will..." commitments with target dates
- Connect commitments to specific scriptures

**Output:** Write evaluation to `study/yt/{video_id}-{slug}.md`
```

---

## Implementation Plan

### Phase 1: Core Download Tool (MVP)

| # | Task | Effort | Files |
|---|------|--------|-------|
| 1 | Project scaffolding — go.mod, main.go, config.go, types.go | Small | 4 files |
| 2 | Port TTML parser from forkirk (ttml.go → transcript.go) | Small | 1 file |
| 3 | Add cue merging / dedup / paragraph detection + timestamp links | Medium | transcript.go |
| 4 | Add cues.json export (raw timestamped cues) | Small | transcript.go |
| 5 | Build yt-dlp wrapper (downloader.go) | Medium | 1 file |
| 6 | Implement `yt_download` tool | Medium | mcp.go + downloader.go |
| 7 | Implement `yt_get` tool | Small | mcp.go |
| 8 | MCP JSON-RPC server (stdin/stdout, same pattern as gospel-vec) | Small | mcp.go |
| 9 | Add `yt/` to .gitignore | Tiny | .gitignore |
| 10 | Test with a real YouTube URL | — | — |

**Deliverable:** `yt-mcp serve` runs as MCP server, `yt_download` takes a URL and produces `./yt/{channel}/{id}/transcript.md` with clickable `?t=` links + `cues.json` for fine-grained quoting

### Phase 2: Browse & Search

| # | Task | Effort | Files |
|---|------|--------|-------|
| 11 | Implement `yt_list` tool | Small | mcp.go |
| 12 | Implement `yt_search` tool (returns matches with `?t=` links) | Medium | mcp.go |

**Deliverable:** Can browse and search across all downloaded transcripts. Search results include clickable timestamp links to the exact moment in the video.

### Phase 3: Workflow Integration

| # | Task | Effort | Files |
|---|------|--------|-------|
| 13 | Add Video Evaluation workflow to copilot-instructions.md | Small | .github/copilot-instructions.md |
| 14 | Create evaluation study template | Small | docs/yt_evaluation_template.md |
| 15 | Integrate `study/yt/` into publish script | Small | scripts/publish/ |
| 16 | Test full 5-step workflow end-to-end | — | — |

**Deliverable:** Complete workflow from URL → download → evaluate → becoming commitments. Evaluation docs publish alongside other study docs.

### Phase 4: Enhancements (Future)

| # | Task | Effort | Notes |
|---|------|--------|-------|
| 17 | Batch download — channel URL or playlist support | Medium | `yt_download_channel` tool |
| 18 | VTT format support (fallback when TTML unavailable) | Small | Parse WebVTT cues into same `[]Cue` model |
| 19 | Vector indexing — feed transcripts into gospel-vec for semantic search | Medium | Cross-tool integration |
| 20 | Auto-detect scripture references in transcript text | Medium | Regex + gospel-mcp verification |
| 21 | YouTube chapter markers — use for section headings | Small | In metadata.json `chapters` field |

---

## Code Reuse from forkirk

### Direct port (copy + adapt)
| forkirk File | yt-mcp File | What to Keep | What to Change |
|-------------|-------------|-------------|----------------|
| `quoter/ttml.go` | `transcript.go` | `ParseTTMLFile()`, `parseClockTime()`, `readInnerText()`, `normalizeWS()` | Remove MongoDB deps, add cue merging, add `?t=` timestamp links, add `cues.json` export |
| `quoter/models.go` | `types.go` | `Cue` struct (Begin, End, Text) | Simplify — no MongoDB, no quote groups. Add `Paragraph` struct (Begin, Text) for merged output |
| `dl_transcripts.bat` | `downloader.go` | yt-dlp flag pattern: `--write-subs --write-auto-subs --sub-format ttml --skip-download` | Wrap in Go `exec.Command`, add `--dump-json`, dynamic output path |

### Do NOT port
- `quoter/importer.go` — MongoDB import logic, not needed
- `quoter/store_mongo.go` — MongoDB storage, not needed
- `quoter/quotes.go` — Quote tagging UI, not needed (our "quoting" is done in study documents)
- `quoter/db.go` — Database layer, not needed

---

## Dependencies

### Required
- **yt-dlp** — installed and in PATH (verified: v2026.02.04 ✅)
- **Go 1.25+** — matches other scripture-study scripts
- No external Go deps for MVP (stdlib only: `encoding/xml`, `encoding/json`, `os/exec`, `bufio`)

### Optional (Phase 4)
- `chromem-go` — if we add vector indexing of transcripts

---

## MCP Server Configuration (for VS Code)

Add to `.vscode/mcp.json` or VS Code MCP settings:

```json
{
  "servers": {
    "yt-mcp": {
      "type": "stdio",
      "command": "${workspaceFolder}/scripts/yt-mcp/yt-mcp.exe",
      "args": ["serve"],
      "env": {
        "YT_DIR": "${workspaceFolder}/yt",
        "YT_STUDY_DIR": "${workspaceFolder}/study/yt"
      }
    }
  }
}
```

---

## Resolved Decisions

1. **VTT fallback** — Phase 4. TTML first (working parser from forkirk). Add VTT when we hit a video that doesn't have TTML.

2. **English language flags** — Punt. Current `--sub-langs "en.*,en"` works. Fix when it doesn't.

3. **Git storage** — `yt/` is `.gitignored`. Copyrighted material, same policy as `gospel-library/`. The tool itself is the reproducible mechanism.

4. **Publish integration** — Yes. `study/yt/` goes through the publish pipeline to `public/study/yt/`. YouTube `?t=` links are already absolute URLs, so no conversion needed (unlike `../gospel-library/` relative links).

---

## Success Criteria

After Phase 1:
- [ ] `yt-mcp serve` starts and responds to MCP requests
- [ ] `yt_download("https://www.youtube.com/watch?v=...")` produces clean `transcript.md`
- [ ] Transcript is readable, paragraph-formatted, deduped, with metadata header
- [ ] Each paragraph has a clickable `[M:SS](url&t=N)` link to that moment in the video
- [ ] `cues.json` preserved for fine-grained timestamp lookups
- [ ] Works for both official subs and auto-generated subs

After Phase 3:
- [ ] Full 5-step workflow works end-to-end in a Copilot session
- [ ] Evaluation doc produced with scripture citations, honest critique, and becoming commitments
- [ ] Video quotes in evaluation docs link to exact timestamps via `?t=`
- [ ] Works equally well for gospel content (BYU devotional) and secular content (news commentary)
- [ ] Cross-references found via gospel-mcp/gospel-vec enrich the evaluation
- [ ] `study/yt/` docs publish to `public/study/yt/` via publish script

---

*"Prove all things; hold fast that which is good."* — [1 Thessalonians 5:21](../../gospel-library/eng/scriptures/nt/1-thes/5.md)
