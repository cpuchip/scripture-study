# Brain Memory Architecture: SQLite + chromem-go

> Replaces: git-backed markdown files in `private-brain/`
> Status: Proposed → Building
> Date: 2026-03-03

## Motivation

The git-backed markdown store works but is write-heavy, read-dumb. You can save a thought and cat a file, but you can't ask "what have I been thinking about related to priesthood?" without grep. There's no semantic memory, no structured query, no cross-session recall.

The second brain needs to be a real memory — shared across models, workspaces, devices, and sessions. SQLite gives us structured queries. chromem-go gives us semantic search. A web UI gives us human eyes on the data.

## Architecture

```
brain.exe
├── internal/store/
│   ├── db.go           # SQLite: entries, tags, audit_log, versions
│   ├── vecstore.go     # chromem-go: semantic index, auto-embed on save
│   ├── types.go        # Entry, AuditRecord (updated — now has ID, embedding status)
│   └── git.go          # KEPT but optional — archive export only
├── internal/web/
│   ├── server.go       # chi router: serves API + embedded frontend
│   ├── api.go          # REST: CRUD entries, search (text + semantic), reclassify
│   ├── handlers.go     # Handler implementations
│   └── dist/           # Embedded SPA (vanilla HTML/JS or Vue, TBD)
├── internal/classifier/ # Unchanged
├── internal/ai/         # Unchanged — also provides embedding func
├── internal/relay/      # Unchanged
├── internal/discord/    # Unchanged
└── internal/config/     # Updated — adds DB path, embedding config, web port
```

## Data Model

### SQLite Schema

```sql
-- Core entries table
CREATE TABLE entries (
    id          TEXT PRIMARY KEY,       -- UUID
    title       TEXT NOT NULL,
    category    TEXT NOT NULL,          -- people, projects, ideas, actions, study, journal, inbox
    body        TEXT NOT NULL,          -- Raw captured text
    confidence  REAL NOT NULL DEFAULT 0.0,
    needs_review BOOLEAN NOT NULL DEFAULT FALSE,
    source      TEXT NOT NULL DEFAULT 'relay',  -- relay, discord, cli, web, app
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Category-specific fields (nullable, populated per category)
    -- People
    person_name   TEXT,
    person_context TEXT,
    follow_ups    TEXT,
    -- Projects
    status        TEXT,    -- active, waiting, blocked, someday, done
    next_action   TEXT,
    -- Ideas
    one_liner     TEXT,
    -- Actions
    due_date      TEXT,
    action_done   BOOLEAN DEFAULT FALSE,
    -- Study
    scripture_refs TEXT,
    insight       TEXT,
    -- Journal
    mood          TEXT,
    gratitude     TEXT
);

-- Tags (many-to-many)
CREATE TABLE tags (
    entry_id TEXT NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    tag      TEXT NOT NULL,
    PRIMARY KEY (entry_id, tag)
);

-- Audit log
CREATE TABLE audit_log (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    entry_id    TEXT REFERENCES entries(id) ON DELETE SET NULL,
    raw_text    TEXT NOT NULL,
    category    TEXT NOT NULL,
    title       TEXT NOT NULL,
    confidence  REAL NOT NULL,
    needs_review BOOLEAN NOT NULL,
    source      TEXT NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Entry versions (simple append-only history)
CREATE TABLE entry_versions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    entry_id    TEXT NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    category    TEXT NOT NULL,
    body        TEXT NOT NULL,
    changed_by  TEXT NOT NULL DEFAULT 'system',  -- system, user, reclassify
    changed_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Embedding status tracking
CREATE TABLE embedding_status (
    entry_id    TEXT PRIMARY KEY REFERENCES entries(id) ON DELETE CASCADE,
    embedded_at DATETIME,
    model       TEXT,         -- embedding model used
    error       TEXT          -- last embedding error, if any
);

-- Indexes
CREATE INDEX idx_entries_category ON entries(category);
CREATE INDEX idx_entries_created ON entries(created_at);
CREATE INDEX idx_entries_needs_review ON entries(needs_review) WHERE needs_review = TRUE;
CREATE INDEX idx_tags_tag ON tags(tag);
CREATE INDEX idx_audit_created ON audit_log(created_at);
```

### chromem-go Vector Store

- **Collection:** `"thoughts"` — one collection for all entries
- **Document ID:** matches `entries.id` (UUID)
- **Content:** `"{title}. {body}"` (title gives the embedding category context)
- **Metadata:** `category`, `source`, `created_at` (for filtering)
- **Persistence:** `NewPersistentDB(dataDir + "/vec", true)` — gob-compressed files
- **Auto-embed on save:** Every `Store.Save()` call triggers `collection.AddDocument()` in the background
- **Re-embed on edit:** Update triggers delete + re-add of the document

### Embedding Function Priority

```go
func chooseEmbedder(cfg *config.Config) chromem.EmbeddingFunc {
    // 1. LM Studio (if available and has embedding endpoint)
    if cfg.LMStudioURL != "" {
        return chromem.NewEmbeddingFuncOpenAICompat(
            cfg.LMStudioURL, "", cfg.EmbeddingModel, nil,
        )
    }
    // 2. Ollama (if OLLAMA_URL is set)
    if cfg.OllamaURL != "" {
        return chromem.NewEmbeddingFuncOllama(cfg.EmbeddingModel, cfg.OllamaURL)
    }
    // 3. OpenAI (if OPENAI_API_KEY is set)
    if os.Getenv("OPENAI_API_KEY") != "" {
        return chromem.NewEmbeddingFuncDefault()
    }
    // 4. Disabled — store works without embeddings, search is text-only
    return nil
}
```

**For us:** LM Studio with an embedding model (same setup as gospel-vec).
**For others:** `ollama pull nomic-embed-text` → set `OLLAMA_URL=http://localhost:11434/api` → done. CPU-only, ~275MB.

## Web UI

Served by brain.exe on a configurable port (default `:8445`). Embedded at compile time.

### Pages

| Route | Purpose |
|-------|---------|
| `/` | Dashboard — recent thoughts, stats, quick capture |
| `/entries` | Browse/filter entries by category, search |
| `/entries/:id` | View/edit single entry |
| `/search` | Semantic search — "find thoughts about X" |
| `/inbox` | Needs-review items for manual reclassification |
| `/archive` | Export to private-brain, bulk operations |

### REST API

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/entries` | List entries (filter: category, tag, date range) |
| GET | `/api/entries/:id` | Get single entry |
| POST | `/api/entries` | Create entry (manual capture from web UI) |
| PUT | `/api/entries/:id` | Update entry (edit title, body, category, fields) |
| DELETE | `/api/entries/:id` | Delete entry |
| POST | `/api/entries/:id/reclassify` | Move to different category |
| GET | `/api/search` | Text search (LIKE queries) |
| GET | `/api/search/semantic` | Semantic search via chromem-go |
| GET | `/api/stats` | Dashboard stats (counts by category, recent activity) |
| POST | `/api/archive` | Export entries to private-brain as markdown |
| GET | `/api/tags` | List all tags with counts |

### Automatic Re-indexing

When an entry is created or updated via the web UI:
1. SQLite row is written/updated
2. Entry version is snapshotted to `entry_versions`
3. chromem-go document is deleted + re-added (re-embedding happens automatically)
4. Audit log entry is written

No manual re-index needed.

## Archive Export (private-brain)

The `private-brain` repo becomes an **archive**, not the live store. Use cases:
- **Pack away finished projects** — export completed ideas/actions as markdown
- **Periodic snapshots** — cron-style export of everything for git-backed backup
- **Readable format** — when you want to browse thoughts as files

Export renders the same YAML front matter + markdown body format that exists today. A `POST /api/archive` call:
1. Reads selected entries from SQLite
2. Renders each as `{category}/{YYYY-MM-DD}-{slug}.md`
3. Writes to the configured `private-brain` directory
4. Optionally commits + pushes via the existing `Git` layer

## Migration

One-time script to import existing `private-brain/` entries:

```
brain migrate --from=./private-brain --db=./brain.db
```

1. Walks each category directory
2. Parses YAML front matter + markdown body
3. INSERTs into `entries` + `tags`
4. Queues embedding generation (batch, with progress bar)
5. Reports: N entries imported, N embedded, N errors

## Config Changes

New environment variables:
```env
BRAIN_DB_PATH=./brain.db           # SQLite database file (default: {BRAIN_DATA_DIR}/brain.db)
BRAIN_VEC_DIR=./vec                # chromem-go persistence dir (default: {BRAIN_DATA_DIR}/vec)
EMBEDDING_MODEL=nomic-embed-text   # Model name for embeddings
EMBEDDING_BACKEND=lmstudio         # lmstudio, ollama, openai, or none
OLLAMA_URL=                        # Ollama API base URL (if using Ollama)
WEB_PORT=8445                      # Web UI port (default: 8445)
WEB_ENABLED=true                   # Enable web UI (default: true)
```

## What Doesn't Change

- **Classifier** — `classifier.Result` stays identical. Storage-agnostic.
- **Relay protocol** — same WebSocket messages, same flow.
- **Discord bot** — same interface, calls `store.Save()` which now writes SQLite instead of files.
- **Mobile app** — connects to relay, doesn't care about storage backend.
- **CLI** — sends thoughts through relay, works the same.
- **AI backends** — LM Studio / Copilot SDK, no changes.

## Build Order

1. **SQLite store** — `db.go` with schema, CRUD, text search. Tests. Wire into `main.go` (git store becomes optional).
2. **chromem-go vector layer** — `vecstore.go` with embed-on-save, semantic search. Tests.
3. **REST API** — `internal/web/` with chi, JSON endpoints. Tests.
4. **Web frontend** — Basic SPA (dashboard, browse, edit, search, inbox). Embedded in binary.
5. **Migration** — Import existing private-brain entries + batch embed.
6. **Archive export** — Wire up the `POST /api/archive` endpoint using the existing `Git` layer.

## Portability Story

For someone else to run this on their machine:

1. `go install github.com/cpuchip/brain@latest` (or clone + build)
2. `ollama pull nomic-embed-text` (for embeddings — optional but recommended)
3. Load a chat model in LM Studio or Ollama (for classification)
4. Create a `.env` with their relay token (or run locally without relay)
5. `brain` — starts the agent. SQLite DB + vec store created automatically.
6. Open `http://localhost:8445` — web UI for browsing/editing

No PostgreSQL. No Docker. No git repo required. One binary + one DB file + optional local models.
