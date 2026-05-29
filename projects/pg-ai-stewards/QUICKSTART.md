# Quickstart

Get the substrate running, verify it, and drive it.

> **Heads-up (pre-extraction):** today the Docker build context reaches into a
> parent monorepo (shared Go workspace + sibling modules), so a bare clone of
> *just* this folder won't build yet. Run from within the parent workspace
> until [.spec/proposals/standalone-extraction.md](.spec/proposals/standalone-extraction.md)
> lands. Everything below assumes that context.

## Prerequisites

- Docker + Docker Compose
- An OpenAI-compatible LLM provider key (e.g. [opencode.ai](https://opencode.ai)
  Zen, Google Gemini, or a local LM Studio / Ollama). At least one is required
  for the bgworker to do real work; embeddings default to a local Ollama
  `nomic-embed-text` model.

## 1. Configure providers

```bash
cp extension/.env.example extension/.env
# edit extension/.env — set at least one provider's API key
```

Providers are configured by env var, naming `STEWARDS_PROVIDER_<NAME>_<FIELD>`
where `<FIELD>` ∈ `BASE_URL | API_KEY | DEFAULT_MODEL | KIND`. Examples ship in
`.env.example` for `opencode_go`, `google_gemini`, `lm_studio`, and `ollama`.

> **Gemini gotcha:** use the OpenAI-compat base URL
> `https://generativelanguage.googleapis.com/v1beta/openai` (the substrate
> POSTs to `{base_url}/chat/completions`; the bare `/v1beta/` path 404s).

## 2. Bring it up

```bash
cd extension
docker compose up -d        # starts three containers:
                            #   pg     — Postgres 18 + extension + bgworker
                            #   ui     — web console (http://127.0.0.1:8080)
                            #   bridge — outbound MCP tool dispatch + migrations
```

The `bridge` entrypoint runs `stewards-cli migrate` against the running `pg`,
applying any pending SQL migrations from `extension/*.sql` (sha-tracked ledger).

## 3. Verify

```bash
# extension loaded?
docker exec pg-ai-stewards-dev psql -U stewards -d stewards -c "SELECT stewards.version();"

# providers loaded from env? (never prints the key)
docker exec pg-ai-stewards-dev psql -U stewards -d stewards -c "SELECT name, base_url, kind, has_api_key FROM stewards.providers_loaded();"

# the model catalog + pricing
docker exec pg-ai-stewards-dev psql -U stewards -d stewards -c "SELECT provider, model FROM stewards.model_pricing ORDER BY provider, model;"
```

Open the web console at **http://127.0.0.1:8080** for the work-items list,
dashboard, scheduled pipelines, models catalog, and the brainstorm form.

## 4. Run something

A brainstorm fans one binding question across multiple "lens" techniques and
synthesizes the results — a quick end-to-end exercise of dispatch + cost
tracking:

```sql
SELECT stewards.start_brainstorm(
    p_binding_question := 'How should we cache provider responses?',
    p_destination      := 'out/brainstorm-cache.md',
    p_lenses           := ARRAY['scamper','six-hats','crazy8s']
);
```

Watch it flow:

```sql
SELECT slug, status, cost_micro_dollars
  FROM stewards.work_items
 WHERE slug LIKE 'brainstorm-%'
 ORDER BY created_at DESC LIMIT 10;
```

## 5. Drive it from an MCP client (Claude Code, etc.)

The `stewards-mcp` server exposes the substrate over MCP — read work items,
inspect runs, dispatch brainstorms, search studies. Point your MCP client at
the `stewards-mcp` binary (stdio transport). Example `.mcp.json` entry:

```json
{
  "mcpServers": {
    "pg-ai-stewards": {
      "command": "/path/to/pg-ai-stewards/bin/stewards-mcp.exe",
      "env": { "STEWARDS_DSN": "postgres://stewards:stewards@localhost:55433/stewards?sslmode=disable" }
    }
  }
}
```

Tools include `work_item_list`, `work_item_show`, `study_search`, `study_get`,
`start_brainstorm`, and the watchman/escalation inspectors. See
`cmd/stewards-mcp/`.

## Operational notes

- **Watchman soak:** a background watchman can run periodic passes. Pause it
  before a rebuild: `UPDATE stewards.watchman_config SET schedule_enabled=false
  WHERE id=1;` (resume with `=true`).
- **SQL-only change:** live-apply with `docker cp` + `psql -f`, no restart.
- **bgworker (Rust) change:** `docker compose build pg && docker compose down &&
  docker compose up -d pg ui`, then `up -d bridge`.
- **Never** `docker compose down -v` — it wipes the data volume.

More conventions in [CONTRIBUTING.md](CONTRIBUTING.md); the runtime map in
[docs/architecture.md](docs/architecture.md).
