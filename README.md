# Scripture Study

AI-assisted scripture study for members of The Church of Jesus Christ of Latter-day Saints. A suite of Go-based tools and AI agent instructions that enable deep, cross-referenced study of the standard works, General Conference talks, manuals, and more — powered by MCP servers and an AI agent of your choice.

## Quick Start

### Prerequisites

| Requirement | Why |
|---|---|
| **Go 1.21+** (with CGO support) | Build all tools. Windows: install [mingw-w64](https://www.mingw-w64.org/) for the C compiler needed by gospel-mcp's SQLite/FTS5. |
| **An AI coding agent** | GitHub Copilot (VS Code), Cursor, Claude Code, OpenCode, Windsurf, etc. |
| **LM Studio _or_ an OpenAI-compatible embeddings API** | Required for gospel-vec (semantic search). Default: `http://localhost:1234/v1` with `text-embedding-qwen3-embedding-4b`. |

> **Model recommendation:** We use **GitHub Copilot with Claude Opus 4.6** for study sessions. Any capable model will work, but Opus-class models handle the cross-referencing depth and source-verification discipline best.

### Step 1 — Download Gospel Content

**This repository does not contain Church content.** The `gospel-library/` folder is in `.gitignore`. The downloader fetches content from the Church's public Gospel Library API (the same API the official apps use).

```powershell
# Interactive TUI — select what to download
go run .\scripts\gospel-library\cmd\gospel-downloader

# Or grab standard works + latest conference in one shot
go run .\scripts\gospel-library\cmd\gospel-downloader --standard
```

See [scripts/gospel-library/README.md](scripts/gospel-library/README.md) for details.

### Step 2 — Build & Index the MCP Servers

Build each server from its directory, then index your downloaded content:

```powershell
# gospel-mcp — full-text search (SQLite FTS5)
cd scripts/gospel-mcp
go build -tags "fts5" -o gospel-mcp.exe ./cmd/gospel-mcp
./gospel-mcp.exe index --root ../../
cd ../..

# gospel-vec — semantic/vector search (requires embeddings endpoint running)
cd scripts/gospel-vec
go build -o gospel-vec.exe .
./gospel-vec.exe index
cd ../..

# webster-mcp — Webster 1828 + modern dictionary
cd scripts/webster-mcp
go build -o webster-mcp.exe ./cmd/...
cd ../..

# yt-mcp — YouTube transcript download & processing
cd scripts/yt-mcp
go build -o yt-mcp.exe .
cd ../..

# search-mcp — web search
cd scripts/search-mcp
go build -o search-mcp.exe .
cd ../..

# becoming — personal transformation tracking
cd scripts/becoming/cmd/mcp
go build -o becoming-mcp.exe .
cd ../../../..
```

### Step 3 — Configure MCP Servers

Create `.vscode/mcp.json` (already in `.gitignore`) pointing to your built executables. Example:

```jsonc
{
  "servers": {
    "gospel": {
      "command": "<repo>/scripts/gospel-mcp/gospel-mcp.exe",
      "args": ["serve", "--db", "<repo>/scripts/gospel-mcp/gospel.db"],
      "type": "stdio"
    },
    "gospel-vec": {
      "command": "<repo>/scripts/gospel-vec/gospel-vec.exe",
      "args": ["mcp", "-data", "<repo>/scripts/gospel-vec/data"],
      "type": "stdio"
    },
    "webster": {
      "command": "<repo>/scripts/webster-mcp/webster-mcp.exe",
      "args": ["-dict", "<repo>/scripts/webster-mcp/data/webster1828.json.gz"],
      "type": "stdio"
    },
    "yt": {
      "command": "<repo>/scripts/yt-mcp/yt-mcp.exe",
      "args": ["serve"],
      "env": {
        "YT_DIR": "<repo>/yt",
        "YT_STUDY_DIR": "<repo>/study/yt",
        "YT_COOKIE_FILE": "<repo>/yt/cookies.txt"
      },
      "type": "stdio"
    },
    "search": {
      "command": "<repo>/scripts/search-mcp/search-mcp.exe",
      "type": "stdio"
    },
    "becoming": {
      "command": "<repo>/scripts/becoming/cmd/mcp/becoming-mcp.exe",
      "env": { "BECOMING_URL": "https://your-instance", "BECOMING_TOKEN": "your-token" },
      "type": "stdio"
    }
  }
}
```

Replace `<repo>` with your absolute path. Non-VS Code agents: translate to your tool's MCP config format.

### Step 4 — Start Studying

Open the workspace in your AI agent. The `.github/copilot-instructions.md` file (and the agents/skills below) automatically provide context about the project structure, study patterns, and source-verification discipline.

## MCP Servers

| Server | Tech | Purpose | MCP Tools |
|---|---|---|---|
| **gospel-mcp** | Go + SQLite/FTS5 | Full-text search over all gospel library content | `gospel_search`, `gospel_get`, `gospel_list` |
| **gospel-vec** | Go + embeddings | Semantic/vector search over scriptures & talks | `search_scriptures`, `search_talks`, `list_books`, `get_talk` |
| **webster-mcp** | Go | Webster 1828 dictionary + Free Dictionary (modern) | `webster_define`, `modern_define`, `webster_search` |
| **yt-mcp** | Go | YouTube transcript download & processing | `yt_download`, `yt_get`, `yt_search`, `yt_list` |
| **search-mcp** | Go | Web search | `web_search` |
| **becoming** | Go | Personal transformation tracking (habit/practice logging) | `create_task`, `log_practice`, `get_today`, etc. |
| **session-journal** | Go (CLI) | Session journaling — captures discoveries, carry-forward items | CLI: `read`, `carry`, `write` |

### gospel-vec Environment Variables

gospel-vec defaults to LM Studio at `localhost:1234`. Override with:

| Variable | Default | Description |
|---|---|---|
| `GOSPEL_VEC_EMBEDDING_URL` | `http://localhost:1234/v1` | Embeddings API endpoint |
| `GOSPEL_VEC_EMBEDDING_MODEL` | `text-embedding-qwen3-embedding-4b` | Embedding model name |
| `GOSPEL_VEC_CHAT_URL` | `http://localhost:1234/v1` | Chat endpoint (for summaries) |
| `GOSPEL_VEC_CHAT_MODEL` | _(auto-detected)_ | Chat model name |
| `GOSPEL_VEC_DATA_DIR` | `./data` | Storage directory |

Works with any OpenAI-compatible embeddings API (LM Studio, Ollama, OpenAI, etc.).

## AI Agent Instructions

This project ships with a complete **GitHub Copilot agent framework** under `.github/`:

```
.github/
├── copilot-instructions.md   # Core principles, project structure, session memory
├── agents/                    # 9 specialized agents
│   ├── study.agent.md         # Deep scripture study
│   ├── lesson.agent.md        # Sunday School / EQ / RS lesson prep
│   ├── talk.agent.md          # Sacrament meeting talk preparation
│   ├── review.agent.md        # Conference talk analysis
│   ├── eval.agent.md          # YouTube video evaluation
│   ├── journal.agent.md       # Personal reflection & journaling
│   ├── podcast.agent.md       # Transform studies into podcast notes
│   ├── dev.agent.md           # MCP server & tool development
│   └── ux.agent.md            # UI/UX design patterns
├── prompts/                   # 5 reusable prompts
│   ├── new-study.prompt.md
│   ├── new-lesson.prompt.md
│   ├── new-eval.prompt.md
│   ├── expound.prompt.md
│   └── study-plan.prompt.md
└── skills/                    # 8 domain skills
    ├── source-verification/   # Read-before-quoting, cite count rule
    ├── scripture-linking/     # Link format conventions
    ├── webster-analysis/      # Webster 1828 word study
    ├── deep-reading/          # Deep reading methodology
    ├── wide-search/           # Broad discovery patterns
    ├── publish-and-commit/    # Study → public HTML pipeline
    ├── becoming/              # Personal transformation
    └── playwright-cli/        # Browser automation
```

### Using a Different AI Agent

The `.github/` instructions are written for **GitHub Copilot in VS Code**. If you use a different tool:

- **Claude Code / Cursor / Windsurf** — Translate the `.github/copilot-instructions.md` into your tool's system prompt or project instructions format. Agent files (`.agent.md`) and skills will need to be adapted to your framework's conventions.
- **OpenCode / other CLI agents** — Extract the core principles and MCP tool descriptions into whatever config your agent reads.
- **Key things to preserve:** source-verification discipline (read before quoting), scripture linking conventions, the session memory architecture at `.spec/memory/`.

## Project Structure

```
scripture-study/
├── .github/                     # AI agent instructions, agents, skills, prompts
├── .spec/                       # Memory system, session journal, proposals
│   ├── memory/                  # identity.md, preferences.yaml, active.md, principles.md
│   └── journal/                 # Session journal entries
│
├── gospel-library/              # Downloaded Church content (NOT in git)
│   └── eng/
│       ├── scriptures/          # Standard works (ot, nt, bofm, dc-testament, pgp)
│       ├── general-conference/  # Conference talks 1971–present
│       ├── manual/              # Come Follow Me, handbooks, etc.
│       └── liahona/             # Magazine content
│
├── study/                       # Study notes and analysis
│   ├── {topic}.md               # Topic-based studies
│   ├── talks/                   # Conference talk analysis
│   └── yt/                      # YouTube video evaluations
├── lessons/                     # Lesson preparation
├── callings/                    # Calling-specific work
├── journal/                     # Personal journal entries
├── becoming/                    # Personal transformation notes
├── books/                       # Additional texts (Lectures on Faith, etc.)
│
├── public/                      # Published HTML for sharing
├── docs/                        # Templates, reflections, meta-docs
│
├── scripts/                     # All Go tools
│   ├── gospel-library/          # Content downloader (TUI)
│   ├── gospel-mcp/              # Full-text search MCP server
│   ├── gospel-vec/              # Semantic search MCP server
│   ├── webster-mcp/             # Webster 1828 dictionary MCP server
│   ├── yt-mcp/                  # YouTube transcript MCP server
│   ├── search-mcp/              # Web search MCP server
│   ├── becoming/                # Personal transformation MCP server
│   ├── session-journal/         # Session journaling CLI
│   ├── publish/                 # Study → public/ HTML converter
│   ├── convert/                 # Conversion utilities
│   └── gospel-library/          # Gospel Library downloader
│
├── go.work                      # Go workspace (all modules)
└── README.md
```

## Templates

| Template | File | Purpose |
|---|---|---|
| Study | `docs/study_template.md` | Personal scripture study sessions — spiritual/physical creation pattern |
| Talk | `docs/talk_template.md` | Sacrament meeting talks — based on analysis of 10+ conference talk patterns |
| Lesson | `docs/lesson_template.md` | Sunday School / RS / EQ — Teaching in the Savior's Way framework |
| Evaluation | `docs/yt_evaluation_template.md` | YouTube video evaluation against the gospel standard |

## Publishing

The `public/` directory holds published versions of study documents for external sharing:

```powershell
go run .\scripts\publish\cmd\main.go
```

Converts working markdown from `study/`, `lessons/`, etc. into polished HTML in `public/`.

## Copyright

**Gospel Library content** is © The Church of Jesus Christ of Latter-day Saints. This repository does **not** include or redistribute Church content — it provides tools to download from the Church's public API for personal study.

**Original content** (templates, study notes, scripts) is released under the MIT License.

---

*"Whatever principle of intelligence we attain unto in this life, it will rise with us in the resurrection."* — [D&C 130:18](gospel-library/eng/scriptures/dc-testament/dc/130.md)
