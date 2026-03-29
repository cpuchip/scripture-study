# MCP Tool Inventory

*All available MCP tools in this workspace, organized by server. Reference for all agents.*
*Generated from source code; update when tools change.*

---

## gospel (gospel-mcp)

Full-text search of the gospel library (scriptures, conference talks, manuals, books).

| Tool | Description | Key Params |
|------|-------------|------------|
| `gospel_search` | FTS5 full-text search across all gospel content | `query` (required), `source` (scriptures/conference/manual/etc), `path`, `limit`, `context`, `include_content` |
| `gospel_get` | Retrieve specific content by scripture reference or file path | `reference` (e.g. "D&C 93:36"), `path`, `context`, `include_chapter` |
| `gospel_list` | Browse available content — volumes, years, manuals | `source`, `path`, `depth` |

**MCP name in VS Code:** `gospel`
**Binary:** `scripts/gospel-mcp/gospel-mcp.exe`
**Data:** `scripts/gospel-mcp/gospel.db` (SQLite + FTS5)
**Source:** `scripts/gospel-mcp/internal/mcp/server.go`

---

## gospel-vec

Semantic vector search using chromem-go embeddings. Four layers: verse, paragraph, summary, theme.

| Tool | Description | Key Params |
|------|-------------|------------|
| `search_scriptures` | Semantic similarity search across scriptures + talks | `query` (required), `layers` (verse/paragraph/summary/theme), `limit` |
| `list_books` | List indexed books, optionally by volume | `volume` (bofm/dc/pgp/ot/nt) |
| `get_talk` | Get full text of a conference talk | `speaker`, `year`, `month`, `file_path` |
| `search_talks` | Search conference talks with speaker/year filters | `query` (required), `speaker`, `year_from`, `year_to`, `limit` |

**MCP name in VS Code:** `gospel-vec`
**Binary:** `scripts/gospel-vec/gospel-vec.exe`
**Data:** `scripts/gospel-vec/data/` (.gob.gz vector stores)
**Source:** `scripts/gospel-vec/mcp.go`

**Important:** Results labeled `[AI SUMMARY]` or `[AI THEME]` are NOT direct quotes. Always verify against the source file.

---

## webster

Webster 1828 dictionary + modern dictionary for word studies.

| Tool | Description | Key Params |
|------|-------------|------------|
| `webster_define` | Look up a word in Webster 1828 | `word` (required) |
| `modern_define` | Look up a word in Free Dictionary API (modern) | `word` (required) |
| `define` | Both 1828 AND modern definitions side by side (recommended) | `word` (required) |
| `webster_search` | Search by word pattern in Webster 1828 | `query` (required), `max_results` |
| `webster_search_definitions` | Search within definitions for text | `query` (required), `max_results` |

**MCP name in VS Code:** `webster`
**Binary:** `scripts/webster-mcp/webster-mcp.exe`
**Data:** `scripts/webster-mcp/data/webster1828.json.gz`
**Source:** `scripts/webster-mcp/internal/mcp/server.go`

---

## becoming (ibeco.me)

Practice tracking, journaling, memorization, tasks, notes, and brain relay.

### Practice & Daily Tools
| Tool | Description | Key Params |
|------|-------------|------------|
| `get_today` | Today's daily summary — all practices with logs/schedule | `date` |
| `list_practices` | List all practices with type/category/active status | `type`, `active_only` |
| `log_practice` | Log a practice completion | `practice_id` (required), `date`, `notes`, `value`, `quality` |
| `create_practice` | Create a new practice (memorize/tracker/habit/scheduled) | `name` (required), `type` (required), `description`, `category`, `config` |
| `get_due_cards` | Get memorization cards due for review | `date` |
| `review_card` | Submit a memorization review (spaced repetition) | `practice_id` (required), `quality` (required, 0-5) |

### Tasks & Notes
| Tool | Description | Key Params |
|------|-------------|------------|
| `list_tasks` | List tasks with status/scripture/description | — |
| `create_task` | Create a task from study or prompting | `title` (required), `type` (action/ongoing/reflection) |
| `update_task` | Update task status or details | `id` (required), `status` (active/completed/deferred) |
| `list_notes` | List notes, optionally filtered | `pinned_only` |
| `create_note` | Create a note (insight, cross-reference, thought) | `content` (required), `practice_id`, `task_id`, `pinned` |

### Journal & Reports
| Tool | Description | Key Params |
|------|-------------|------------|
| `get_reflection` | Get daily reflection journal entry | `date` |
| `upsert_reflection` | Create/update daily reflection | `content` (required), `date`, `mood` (1-5) |
| `get_report` | Progress report for date range | `start` (required), `end` (required) |
| `get_today_prompt` | Rotating daily reflection prompt | — |
| `list_corrupted_practices` | Find practices with corrupted data | — |

### Brain (Second Brain Relay)
| Tool | Description | Key Params |
|------|-------------|------------|
| `brain_search` | Search brain entries by text | `query` (required), `category`, `limit` |
| `brain_recent` | Get recent entries, newest first | `category`, `limit` |
| `brain_get` | Get a single entry by UUID | `id` (required) |
| `brain_stats` | Relay status — is brain.exe online? | — |
| `brain_tags` | List all tags with usage counts | — |
| `brain_create` | Create a new brain entry | `title` (required), `body`, `category`, `status`, `tags` |
| `brain_update` | Update an existing entry (partial update safe) | `id` (required), + any field to change |
| `brain_delete` | Delete an entry by UUID | `id` (required) |

Brain categories: `inbox`, `actions`, `projects`, `ideas`, `people`, `study`, `journal`

**MCP name in VS Code:** `becoming`
**Binary:** `scripts/becoming/mcp.exe`
**API base:** `https://ibeco.me` (via `BECOMING_URL` env var)
**Source:** `scripts/becoming/cmd/mcp/main.go`

---

## yt (yt-mcp)

YouTube transcript download and search via yt-dlp.

| Tool | Description | Key Params |
|------|-------------|------------|
| `yt_download` | Download transcript + metadata from a YouTube video | `url` (required), `force`, `cookies` |
| `yt_get` | Get full transcript of a previously downloaded video | `video_id` or `path` |
| `yt_list` | List downloaded transcripts, optionally by channel | `channel`, `limit` |
| `yt_search` | Search across all downloaded transcripts | `query` (required), `channel`, `limit` |

**MCP name in VS Code:** `yt`
**Binary:** `scripts/yt-mcp/yt-mcp.exe`
**Data:** `yt/` (transcripts organized by channel/video_id)
**Source:** `scripts/yt-mcp/mcp.go`
**Requires:** `yt-dlp` in PATH

---

## search (search-mcp)

Web search via DuckDuckGo.

| Tool | Description | Key Params |
|------|-------------|------------|
| `web_search` | General web search | `query` (required), `max_results` |
| `news_search` | Recent news articles | `query` (required), `max_results`, `timelimit` (d/w/m) |
| `instant_answer` | Quick factual answers, definitions | `query` (required) |

**MCP name in VS Code:** `search`
**Binary:** `scripts/search-mcp/search-mcp.exe`
**Source:** `scripts/search-mcp/internal/mcp/server.go`

---

## byu-citations

BYU Scripture Citation Index — who cited what verse in conference.

| Tool | Description | Key Params |
|------|-------------|------------|
| `byu_citations` | Look up who cited a scripture verse | `reference` (required, e.g. "3 Nephi 21:10") |
| `byu_citations_bulk` | Look up citations for multiple references at once | `references` (required, comma-separated) |
| `byu_citations_books` | List all books with BYU Citation Index IDs | — |

**MCP name in VS Code:** `byu-citations`
**Binary:** `scripts/byu-citations/byu-citations.exe`
**Source:** `scripts/byu-citations/internal/mcp/server.go`

---

## exa-search (Remote)

Web search via Exa AI. **This is a remote MCP server — no local binary.**

| Tool | Description | Key Params |
|------|-------------|------------|
| `web_search_exa` | Exa AI web search — neural search with source types | See Exa docs |

**MCP name in VS Code:** `exa-search`
**Type:** Remote (`https://mcp.exa.ai/mcp?tools=web_search_exa`)
**Source:** Cloud-hosted by Exa

**NOTE FOR AGENTS:** This tool is a *deferred tool*. You must use `tool_search_tool_regex` with pattern `exa` to load it before calling it. It will appear as `mcp_exa-search_web_search_exa` in the tool list. Do NOT try to call it without loading it first.

---

## Accessing Deferred MCP Tools

All MCP tools are deferred in VS Code and must be discovered before use:

```
# To find gospel tools:
tool_search_tool_regex → pattern: "gospel_search|gospel_get|gospel_list"

# To find becoming/brain tools:
tool_search_tool_regex → pattern: "mcp_becoming"

# To find exa web search:
tool_search_tool_regex → pattern: "exa"

# To find webster tools:
tool_search_tool_regex → pattern: "webster"

# To find YouTube tools:
tool_search_tool_regex → pattern: "mcp_yt"

# To find BYU citations:
tool_search_tool_regex → pattern: "byu.citation"

# To find search/DuckDuckGo:
tool_search_tool_regex → pattern: "mcp_.*web_search$"
```

The deferred tool names follow the pattern `mcp_{server-name}_{tool-name}`. Check the `availableDeferredTools` list in the system prompt for the exact names.
