# MCP Tool Inventory

*All available MCP tools in this workspace, organized by server. Reference for all agents.*
*Last verified by direct invocation: 2026-04-25.*

---

## Quick Reference: Working Function Names

These are the names you actually call. Verified by successful invocation this session. The function-name prefix is `mcp_{vscode-prefix}_{tool-name}`. The vscode prefix usually matches the server name in `mcp.json` — except VS Code strips trailing version suffixes (e.g. `gospel-engine-v2` → `gospel-engine`).

| Capability | Function Name |
|------------|--------------|
| Gospel search (FTS / semantic / combined) | `mcp_gospel-engine_gospel_search` |
| Gospel get (by reference or path) | `mcp_gospel-engine_gospel_get` |
| Gospel list (browse) | `mcp_gospel-engine_gospel_list` |
| Webster 1828 lookup | `mcp_webster_webster_define` |
| Webster + modern side by side | `mcp_webster_define` |
| Modern dictionary | `mcp_webster_modern_define` |
| Webster word search | `mcp_webster_webster_search` |
| Webster definition search | `mcp_webster_webster_search_definitions` |
| YouTube transcript download | `mcp_yt_yt_download` |
| YouTube transcript get | `mcp_yt_yt_get` |
| YouTube transcript list | `mcp_yt_yt_list` |
| YouTube transcript search | `mcp_yt_yt_search` |
| BYU citation single | `mcp_byu-citations_byu_citations` |
| BYU citation bulk | `mcp_byu-citations_byu_citations_bulk` |
| BYU citation books index | `mcp_byu-citations_byu_citations_books` |
| Brain search | `mcp_becoming_brain_search` |
| Brain recent | `mcp_becoming_brain_recent` |
| Brain get/create/update/delete | `mcp_becoming_brain_get` / `_create` / `_update` / `_delete` |
| Brain stats / tags | `mcp_becoming_brain_stats` / `_tags` |
| Becoming today / practices | `mcp_becoming_get_today` / `_list_practices` / `_log_practice` / `_create_practice` |
| Becoming due cards / review | `mcp_becoming_get_due_cards` / `_review_card` |
| Becoming tasks | `mcp_becoming_list_tasks` / `_create_task` / `_update_task` |
| Becoming notes | `mcp_becoming_list_notes` / `_create_note` |
| Becoming reflections | `mcp_becoming_get_reflection` / `_upsert_reflection` |
| Becoming reports | `mcp_becoming_get_report` / `_get_today_prompt` |
| Web search (Exa neural) | `mcp_exa-search_web_search_exa` |
| Web search (DuckDuckGo) | `mcp_search_web_search` |

---

## Servers Configured (`.vscode/mcp.json`)

7 servers: 6 stdio + 1 remote http.

### gospel-engine-v2 (stdio)

Hosted PG + pgvector backend at `engine.ibeco.me`. Thin MCP client shipped as `gospel-mcp.exe`. **VS Code tool prefix: `gospel-engine` (no `-v2`).**

| Tool | Description | Key Params |
|------|-------------|------------|
| `gospel_search` | Search across all gospel content | `query` (required), `mode` (`keyword` / `semantic` / `combined`), `source`, `path`, `limit`, `context`, `include_content` |
| `gospel_get` | Retrieve content by scripture reference or file path | `reference` (e.g. "D&C 93:36") or `path`, `context`, `include_chapter` |
| `gospel_list` | Browse available content — volumes, years, manuals | `source`, `path`, `depth` |

**Binary:** `scripts/gospel-engine/gospel-mcp.exe`
**Auto-update:** `GOSPEL_AUTO_UPDATE=true` — client updates from engine.ibeco.me on launch.

### webster (stdio)

Webster 1828 dictionary + Free Dictionary API.

| Tool | Description | Key Params |
|------|-------------|------------|
| `webster_define` | Webster 1828 lookup | `word` (required) |
| `modern_define` | Free Dictionary API (modern) | `word` (required) |
| `define` | Both 1828 + modern side by side | `word` (required) |
| `webster_search` | Search Webster by word pattern | `query` (required), `max_results` |
| `webster_search_definitions` | Search within definitions | `query` (required), `max_results` |

**Binary:** `scripts/webster-mcp/webster-mcp.exe`
**Data:** `scripts/webster-mcp/data/webster1828.json.gz`

### yt (stdio)

YouTube transcript download/search via `yt-dlp`.

| Tool | Description | Key Params |
|------|-------------|------------|
| `yt_download` | Download transcript + metadata | `url` (required), `force`, `cookies` |
| `yt_get` | Get full transcript of a downloaded video | `video_id` or `path` |
| `yt_list` | List downloaded transcripts | `channel`, `limit` |
| `yt_search` | Search across all transcripts | `query` (required), `channel`, `limit` |

**Binary:** `scripts/yt-mcp/yt-mcp.exe`
**Data:** `yt/{channel}/{video_id}/`
**Requires:** `yt-dlp` in PATH.

### byu-citations (stdio)

BYU Scripture Citation Index — who cited what verse in conference / Journal of Discourses.

| Tool | Description | Key Params |
|------|-------------|------------|
| `byu_citations` | Citations for one verse | `reference` (e.g. "3 Nephi 21:10") |
| `byu_citations_bulk` | Citations for many verses | `references` (comma-separated) |
| `byu_citations_books` | List all books with their BYU IDs | — |

**Binary:** `scripts/byu-citations/byu-citations.exe`

### becoming (stdio)

Practice tracking, journaling, memorization, tasks, notes, and brain relay (ibeco.me API).

**Practice & Daily**
| Tool | Description | Key Params |
|------|-------------|------------|
| `get_today` | Today's daily summary | `date` |
| `list_practices` | All practices | `type`, `active_only` |
| `log_practice` | Log a practice completion | `practice_id` (required), `date`, `notes`, `value`, `quality` |
| `create_practice` | Create a new practice | `name`, `type` (memorize/tracker/habit/scheduled), `description`, `category`, `config` |
| `get_due_cards` | Memorization cards due | `date` |
| `review_card` | Submit a SR review | `practice_id`, `quality` (0-5) |

**Tasks & Notes**
| Tool | Description |
|------|-------------|
| `list_tasks` / `create_task` / `update_task` | Task CRUD |
| `list_notes` / `create_note` | Notes |

**Journal & Reports**
| Tool | Description |
|------|-------------|
| `get_reflection` / `upsert_reflection` | Daily reflection |
| `get_report` | Progress report by date range |
| `get_today_prompt` | Rotating daily prompt |
| `list_corrupted_practices` | Diagnostic |

**Brain (Second Brain Relay)**
| Tool | Description |
|------|-------------|
| `brain_search` / `brain_recent` / `brain_get` | Read |
| `brain_create` / `brain_update` / `brain_delete` | Write |
| `brain_stats` / `brain_tags` | Meta |

Brain categories: `inbox`, `actions`, `projects`, `ideas`, `people`, `study`, `journal`.

**Binary:** `scripts/becoming/mcp.exe`
**API:** `https://ibeco.me`

### exa-search (remote http)

Exa AI neural web search. **No local binary.**

| Tool | Description | Key Params |
|------|-------------|------------|
| `web_search_exa` | Neural web search with content extraction | `query`, `numResults` |

**URL:** `https://mcp.exa.ai/mcp?tools=web_search_exa`

### search (stdio)

DuckDuckGo web search. Fast, no API key.

| Tool | Description | Key Params |
|------|-------------|------------|
| `web_search` | General web search | `query`, `max_results` |
| `news_search` | Recent news articles | `query`, `max_results`, `timelimit` (d/w/m) |
| `instant_answer` | Quick factual answers | `query` |

**Binary:** `scripts/search-mcp/search-mcp.exe`

---

## Known Gotchas

- **`gospel-engine-v2` → `gospel-engine` in tool names.** VS Code strips the `-v2` suffix when generating function names. The server name in `mcp.json` keeps `-v2`, but every function call uses `mcp_gospel-engine_*`. We have hit this trap multiple times — trust the table at the top of this file.
- **AI summaries are not quotes.** Results labeled `[AI SUMMARY]` or `[AI THEME]` from semantic search are paraphrases generated by the indexing pipeline. Always `read_file` the source before quoting.
- **Brain agent online status.** `brain_stats` returns `agent_online: false` when `brain.exe` is not running locally. Search/recent still work against the relay's queued data.
- **MCP tools are listed as deferred** in the system prompt but can usually be called directly. Try the direct call first; only fall back to `tool_search_tool_regex` if the call fails because the tool isn't loaded.

---

## Legacy Scripts (not registered in `mcp.json`)

Kept on disk as fallback only. Do not call.

- `scripts/gospel-mcp/` — FTS5 only. Superseded by gospel-engine-v2.
- `scripts/gospel-vec/` — chromem-go vector. Superseded by gospel-engine-v2.
- `scripts/gospel-engine/` (the v1, not v2) — local combined backend. Superseded by hosted v2.
