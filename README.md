# Scripture Study

AI-assisted scripture study for members of The Church of Jesus Christ of Latter-day Saints. A suite of Go-based tools and AI agent instructions that enable deep, cross-referenced study of the standard works, General Conference talks, manuals, and more — powered by MCP servers and an AI agent of your choice.

> *"Whatever principle of intelligence we attain unto in this life, it will rise with us in the resurrection."* — D&C 130:18-19

## Quick Start

### Prerequisites

| Requirement | Why |
|---|---|
| **Go 1.21+** (with CGO support) | Build all tools. Windows: install [mingw-w64](https://www.mingw-w64.org/) for the C compiler needed by gospel-mcp's SQLite/FTS5. |
| **An AI coding agent** | GitHub Copilot (VS Code), Cursor, Claude Code, OpenCode, Windsurf, etc. |
| **LM Studio _or_ an OpenAI-compatible embeddings API** | Required for gospel-vec (semantic search). Default: `http://localhost:1234/v1` with `text-embedding-qwen3-embedding-4b`. |

> **Model recommendation:** We use **GitHub Copilot with Claude Opus 4.6** for study sessions. Any capable model will work, but Opus-class models handle the cross-referencing depth and source-verification discipline best.

### Step 1 — Clone the Workspace

```bash
git clone https://github.com/cpuchip/scripture-study.git
cd scripture-study
```

### Step 2 — Clone Companion Repos

Several components live in separate git repos but are designed to be cloned into this workspace. The `.gitignore` already ignores these directories.

```bash
# Brain — local second brain (capture, classify, search)
git clone https://github.com/cpuchip/brain.git scripts/brain

# Brain App — Flutter mobile/desktop app
git clone https://github.com/cpuchip/brain-app.git scripts/brain-app

# Chip Voice — TTS engine (markdown → audio, podcast → transcript)
git clone https://github.com/cpuchip/chip-voice.git scripts/chip-voice

# Teaching — interactive teaching content
git clone https://github.com/cpuchip/teaching.git teaching

```

**About private-brain:** The `private-brain/` directory holds your personal brain data — the actual thoughts, journal entries, and captures that brain.exe manages. This folder is gitignored from scripture-study. You can either create your own private repo and clone it there, or just let brain.exe create the directory when you first use it. No public template repo exists — it's your personal data from the start.

All companion repos are optional. The core study workflow (gospel content + MCP servers + agents) works without them.

### Step 3 — Download Gospel Content

**This repository does not contain Church content.** The `gospel-library/` folder is in `.gitignore`. The downloader fetches content from the Church's public Gospel Library API (the same API the official apps use).

```powershell
# Interactive TUI — select what to download
go run .\scripts\gospel-library\cmd\gospel-downloader

# Or grab standard works + latest conference in one shot
go run .\scripts\gospel-library\cmd\gospel-downloader --standard
```

See [scripts/gospel-library/README.md](scripts/gospel-library/README.md) for details.

### Step 4 — Build & Index the MCP Servers

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

# byu-citations — BYU Scripture Citation Index
cd scripts/byu-citations
go build -o byu-citations.exe .
cd ../..

# becoming — personal transformation tracking
cd scripts/becoming/cmd/mcp
go build -o becoming-mcp.exe .
cd ../../../..
```

### Step 5 — Configure MCP Servers

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
    "byu-citations": {
      "command": "<repo>/scripts/byu-citations/byu-citations.exe",
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

### Step 6 — Start Studying

Open the workspace in your AI agent. The `.github/copilot-instructions.md` file (and the agents/skills below) automatically provide context about the project structure, study patterns, and source-verification discipline.

---

## MCP Servers

| Server | Tech | Purpose | MCP Tools |
|---|---|---|---|
| **gospel-mcp** | Go + SQLite/FTS5 | Full-text search over all gospel library content | `gospel_search`, `gospel_get`, `gospel_list` |
| **gospel-vec** | Go + embeddings | Semantic/vector search over scriptures & talks | `search_scriptures`, `search_talks`, `list_books`, `get_talk` |
| **webster-mcp** | Go | Webster 1828 dictionary + Free Dictionary (modern) | `webster_define`, `modern_define`, `webster_search` |
| **yt-mcp** | Go | YouTube transcript download & processing | `yt_download`, `yt_get`, `yt_search`, `yt_list` |
| **search-mcp** | Go | Web search via DuckDuckGo | `web_search` |
| **byu-citations** | Go | BYU Scripture Citation Index — find talks that cite a verse | `byu_citations`, `byu_citations_books`, `byu_citations_bulk` |
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

---

## Companion Repos

This workspace is a multi-repo project. The main scripture-study repo contains the core study workflow, agents, and MCP servers. Companion repos handle separate concerns:

| Repo | Clone Location | Purpose | Required? |
|------|---------------|---------|-----------|
| [brain](https://github.com/cpuchip/brain) | `scripts/brain/` | Local second brain — capture, classify, store, search (Go + SQLite + chromem-go) | Optional |
| [brain-app](https://github.com/cpuchip/brain-app) | `scripts/brain-app/` | Flutter mobile/desktop app for the brain ecosystem | Optional |
| [chip-voice](https://github.com/cpuchip/chip-voice) | `scripts/chip-voice/` | TTS engine — markdown to audio, podcast to transcript (Python + Qwen3-TTS/Kokoro) | Optional |
| [teaching](https://github.com/cpuchip/teaching) | `teaching/` | Interactive teaching content — episode scripts and web presentations | Optional |
| private-brain | `private-brain/` | Personal brain data — create your own private repo or let brain.exe generate it | Optional |

### Brain / Becoming Ecosystem

The brain ecosystem spans three components:

- **brain.exe** (`scripts/brain/`) — Local brain. Captures thoughts via Discord DM, web UI, or relay. Classifies with AI (LM Studio or Copilot SDK). Stores in SQLite with semantic vector search.
- **ibeco.me** (`scripts/becoming/`) — Cloud hub deployed via Dokploy. Connects to brain.exe via WebSocket relay. Provides web UI, practices, journaling, and becoming features. Part of this repo.
- **brain-app** (`scripts/brain-app/`) — Flutter app (Android, Windows; iOS/Mac planned). Connects to brain.exe directly or through the ibeco.me relay.
- **private-brain** (`private-brain/`) — Your personal brain data (markdown files with YAML front matter). Created by brain.exe on first use, or bring your own private repo. Gitignored from scripture-study.

See each component's README for setup: [brain](scripts/brain/README.md), [brain-app](scripts/brain-app/README.md), [chip-voice](scripts/chip-voice/README.md).

---

## AI Agent Framework

This project ships with a complete agent framework under `.github/`. Designed for GitHub Copilot in VS Code, but adaptable to other tools.

### Agents (14)

| Agent | Purpose |
|-------|---------|
| `study` | Deep scripture study — phased writing with externalized memory and critical analysis |
| `lesson` | Lesson planning — phased preparation with scratch files and pedagogy framework |
| `talk` | Sacrament meeting talk preparation |
| `review` | Conference talk analysis for teaching patterns |
| `eval` | YouTube video evaluation — phased evaluation with charitable critical analysis |
| `journal` | Personal reflection, journaling, becoming |
| `plan` | Planning — from idea to spec with critical analysis and creation cycle review |
| `podcast` | Transform studies into shareable podcast/video notes |
| `story` | Weave studies into narrative — emotional arc, pacing, contrast |
| `dev` | MCP server and tool development |
| `ux` | UI/UX expert — design patterns, interaction flows, visual quality |
| `sabbath` | Structured reflection after completed cycles — ending, seeing, declaring |
| `teaching` | Teaching preparation — from study to shareable content with honesty guardrails and the Ben Test |
| `debug` | Systematic debugging — Agans' 9 rules applied to code, tools, and intellectual problems |

### Skills (13)

| Skill | Purpose |
|-------|---------|
| `source-verification` | Read-before-quoting discipline, cite count rule, pre-publish checklist |
| `scripture-linking` | Link format conventions for scriptures and conference talks |
| `webster-analysis` | Webster 1828 word study for Restoration-era vocabulary |
| `deep-reading` | Methodology for close reading of scripture texts |
| `wide-search` | Broad discovery patterns across the corpus |
| `quote-log` | Scratch file format for tracking verified quotes |
| `critical-analysis` | Stress-testing arguments before committing to a narrative |
| `becoming` | Personal transformation — applying what we learn |
| `ben-test` | Calibrated self-assessment — do we practice what we've written? |
| `publish-and-commit` | Study to public HTML pipeline |
| `playwright-cli` | Browser automation for web testing |
| `dokploy` | Deployment status and management |
| `byu-citations` | BYU Scripture Citation Index lookups |

### Prompts (5)

`new-study`, `new-lesson`, `new-eval`, `expound`, `study-plan`

### Using a Different AI Agent

The `.github/` instructions are written for **GitHub Copilot in VS Code**. If you use a different tool:

- **Claude Code / Cursor / Windsurf** — Translate `.github/copilot-instructions.md` into your tool's system prompt or project instructions format. Agent files and skills will need adapting.
- **OpenCode / other CLI agents** — Extract core principles and MCP tool descriptions into your agent's config.
- **Key things to preserve:** source-verification discipline (read before quoting), scripture linking conventions, the session memory architecture at `.spec/memory/`.

---

## Session Memory & Specification

The `.spec/` directory contains the project's memory architecture, covenant, and planning documents:

```
.spec/
├── covenant.yaml              # Bilateral commitment governing collaboration
├── memory/                    # Persistent context across sessions
│   ├── identity.md            # Who we are together
│   ├── preferences.yaml       # Personal context
│   ├── active.md              # Current state — what's in flight
│   ├── decisions.md           # Settled questions
│   └── principles.md          # Enduring insights
├── journal/                   # Session journal entries (YAML)
├── learnings/                 # Named failures → learning entries
├── sabbath/                   # Sabbath reflection records
├── proposals/                 # Feature/workstream proposals
├── scratch/                   # Research provenance — permanent working notes
└── prompts/                   # Reusable system prompts
```

This architecture ensures agents arrive with context rather than as strangers. See `.github/copilot-instructions.md` for the session start/end protocol.

---

## Project Structure

```
scripture-study/
├── intent.yaml                  # Root intent — purpose, values, constraints
├── .github/                     # AI agent framework
│   ├── copilot-instructions.md  # Core principles, project structure, session protocol
│   ├── agents/                  # 14 specialized agents
│   ├── skills/                  # 13 domain skills
│   └── prompts/                 # 5 reusable prompts
├── .spec/                       # Memory, covenant, proposals, journal
│
├── gospel-library/              # Downloaded Church content (NOT in git)
│   └── eng/
│       ├── scriptures/          # Standard works (ot, nt, bofm, dc-testament, pgp)
│       ├── general-conference/  # Conference talks 1971–present
│       └── manual/              # Come Follow Me, handbooks, etc.
│
├── study/                       # Study notes and analysis
│   ├── {topic}.md               # Topic-based studies
│   ├── .scratch/                # Research provenance for studies
│   ├── talks/                   # Conference talk analysis
│   └── yt/                      # YouTube video evaluations
├── lessons/                     # Lesson preparation
├── callings/                    # Calling-specific work
├── journal/                     # Personal journal entries
├── becoming/                    # Personal transformation notes
├── books/                       # Additional texts (Lectures on Faith, etc.)
├── docs/                        # Templates, reflections, work-with-AI guide
│
├── public/                      # Published HTML for sharing
│
├── scripts/                     # All Go tools & companion repos
│   ├── gospel-library/          # Content downloader (TUI)
│   ├── gospel-mcp/              # Full-text search MCP server
│   ├── gospel-vec/              # Semantic search MCP server
│   ├── webster-mcp/             # Webster 1828 dictionary MCP server
│   ├── yt-mcp/                  # YouTube transcript MCP server
│   ├── search-mcp/              # Web search MCP server
│   ├── byu-citations/           # BYU Scripture Citation Index MCP server
│   ├── becoming/                # ibeco.me — cloud hub + becoming (Go + Vue 3)
│   ├── session-journal/         # Session journaling CLI
│   ├── publish/                 # Study → public/ HTML converter
│   ├── brain/                   # brain.exe (separate repo — clone into here)
│   ├── brain-app/               # Flutter app (separate repo — clone into here)
│   └── chip-voice/              # TTS engine (separate repo — clone into here)
│
├── teaching/                    # Teaching content (separate repo — clone into here)
├── private-brain/               # Personal brain data (separate repo — fork & clone)
│
├── go.work                      # Go workspace (all modules)
└── README.md
```

## Templates

| Template | File | Purpose |
|---|---|---|
| Study | `docs/study_template.md` | Personal scripture study sessions |
| Talk | `docs/talk_template.md` | Sacrament meeting talks |
| Lesson | `docs/lesson_template.md` | Sunday School / RS / EQ lessons |
| Evaluation | `docs/yt_evaluation_template.md` | YouTube video evaluation |

## Publishing

```powershell
go run .\scripts\publish\cmd\main.go
```

Converts working markdown from `study/`, `lessons/`, etc. into polished HTML in `public/`.

## Copyright

**Gospel Library content** is © The Church of Jesus Christ of Latter-day Saints. This repository does **not** include or redistribute Church content — it provides tools to download from the Church's public API for personal study.

**Original content** (templates, study notes, scripts) is released under the MIT License.

---

*"Whatever principle of intelligence we attain unto in this life, it will rise with us in the resurrection."* — [D&C 130:18](gospel-library/eng/scriptures/dc-testament/dc/130.md)
