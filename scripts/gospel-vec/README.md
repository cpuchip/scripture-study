# Gospel Vec — Semantic Search MCP Server

Vector/semantic search engine for gospel content. Indexes scriptures, conference talks, manuals, and music into a local vector database using embeddings, then serves search via MCP for AI assistant integration.

## Prerequisites

- Go 1.21+
- [LM Studio](https://lmstudio.ai/) or any OpenAI-compatible embeddings API running locally
- Gospel content downloaded via [gospel-library](../gospel-library/README.md)

## Build

```bash
go build -o gospel-vec.exe .
```

## Index

```bash
# Index everything
./gospel-vec.exe index-all

# Or index individually
./gospel-vec.exe index          # Scriptures
./gospel-vec.exe index-talks    # Conference talks
./gospel-vec.exe index-manuals  # Manuals
./gospel-vec.exe index-music    # Hymns & music
```

## Run

```bash
./gospel-vec.exe mcp -data ./data
```

## MCP Tools

| Tool | Description |
|------|-------------|
| `search_scriptures` | Semantic search across indexed scriptures |
| `search_talks` | Semantic search across conference talks |
| `list_books` | List available indexed books |
| `get_talk` | Retrieve a specific talk by path |

## Commands

| Command | Purpose |
|---------|---------|
| `index` | Index scriptures |
| `index-talks` | Index conference talks |
| `index-manuals` | Index manuals |
| `index-music` | Index hymns/music |
| `index-all` | Index everything |
| `search` | CLI search (for testing) |
| `mcp` | Start MCP stdio server |
| `stats` | Show index statistics |
| `config` | Show current configuration |

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `GOSPEL_VEC_EMBEDDING_URL` | `http://localhost:1234/v1` | Embeddings API endpoint |
| `GOSPEL_VEC_EMBEDDING_MODEL` | `text-embedding-qwen3-embedding-4b` | Embedding model name |
| `GOSPEL_VEC_CHAT_URL` | `http://localhost:1234/v1` | Chat endpoint (for summaries) |
| `GOSPEL_VEC_CHAT_MODEL` | _(auto-detected)_ | Chat model name |
| `GOSPEL_VEC_DATA_DIR` | `./data` | Storage directory |

Works with any OpenAI-compatible embeddings API (LM Studio, Ollama, OpenAI, etc.).
