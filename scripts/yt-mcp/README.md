# YouTube MCP Server

MCP server that downloads YouTube video transcripts (via yt-dlp) and metadata, stores them locally organized by channel, and exposes search/retrieval tools to AI assistants.

## Prerequisites

- Go 1.21+
- [yt-dlp](https://github.com/yt-dlp/yt-dlp) installed and on PATH

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
| `YT_DIR` | Root directory for downloaded transcripts (default: `../../yt`) |
| `YT_STUDY_DIR` | Directory for study/evaluation files (default: `../../study/yt`) |
| `YT_COOKIE_FILE` | Path to cookies.txt for age-restricted content |

## MCP Tools

| Tool | Description |
|------|-------------|
| `yt_download` | Download transcript and metadata for a YouTube video |
| `yt_get` | Retrieve a previously downloaded transcript |
| `yt_list` | List downloaded videos, optionally filtered by channel |
| `yt_search` | Search across downloaded transcripts |

Transcripts are stored as markdown files under `{YT_DIR}/{channel}/{video_id}/`.
