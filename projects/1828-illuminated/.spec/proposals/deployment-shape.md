---
title: 1828-illuminated — deployment shape
date: 2026-05-20
status: proposed
workstream: WS7
parent: backend-pivot.md
purpose: >
  Specify how the three-container deploy (frontend + backend + db) is
  composed, wired, and shipped to Dokploy at 1828.ibeco.me without
  losing the simplicity of today's single-container deploy or coupling
  to pg-ai-stewards' Postgres lifecycle.
---

# Deployment Shape

## I. Today vs. tomorrow

**Today.** Single image, single container.

```
docker build -t 1828-illuminated .   # multi-stage: node:22-alpine → nginx:1.27-alpine
docker run -p 8080:80 1828-illuminated
```

Dokploy pulls from `main`, auto-builds. Site goes live at `1828.ibeco.me`. No state, no database, no backend, no cross-service coordination.

**Tomorrow.** Three containers under one compose file.

```
docker compose up -d
```

Same Dokploy auto-build hook, but it points at `docker-compose.yaml` instead of a single `Dockerfile`. Dokploy supports both modes; the migration is a project-type change in the Dokploy UI plus a docker-compose.yaml at the project root.

## II. The compose file

Lives at `projects/1828-illuminated/docker-compose.yaml`. Modeled on `scripts/becoming/docker-compose.yml` (proven Dokploy pattern in this workspace).

```yaml
# 1828-illuminated — Dokploy-deploy via Compose
# Three services: frontend (nginx + Vue dist), backend (Go), db (Postgres 17-alpine).
# Volume `pg-data` survives `docker compose down` (named, not anonymous).
# `docker compose down -v` wipes the DB volume — don't run unless reseeding.

services:
  frontend:
    build:
      context: .
      dockerfile: Dockerfile.frontend
    restart: unless-stopped
    ports:
      - "80:80"                 # ← Dokploy's Traefik intercepts and TLS-terminates
    environment:
      BACKEND_UPSTREAM: http://backend:8080
    depends_on:
      backend:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost/healthz"]
      interval: 30s
      timeout: 3s
      retries: 3

  backend:
    build:
      context: .
      dockerfile: Dockerfile.backend
    restart: unless-stopped
    # Internal port; not exposed to host. Frontend reaches it by service name.
    environment:
      DATABASE_URL: postgres://i1828:${POSTGRES_PASSWORD}@db:5432/i1828?sslmode=disable
      LLM_PROVIDER: ${LLM_PROVIDER:-mock}
      LLM_BASE_URL: ${LLM_BASE_URL:-}
      LLM_API_KEY: ${LLM_API_KEY:-}
      LLM_MODEL: ${LLM_MODEL:-}
      LLM_PROXY_ENABLED: ${LLM_PROXY_ENABLED:-false}
      LLM_RATE_PER_IP_PER_MIN: ${LLM_RATE_PER_IP_PER_MIN:-10}
      LLM_GLOBAL_TOKEN_CAP_PER_DAY: ${LLM_GLOBAL_TOKEN_CAP_PER_DAY:-200000}
      MODERN_FETCH_DAILY_CAP: ${MODERN_FETCH_DAILY_CAP:-5000}
      THUMMIM_SYNC_ENABLED: ${THUMMIM_SYNC_ENABLED:-false}
      THUMMIM_SOURCE_URL: ${THUMMIM_SOURCE_URL:-}
    depends_on:
      db:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:8080/api/healthz"]
      interval: 30s
      timeout: 5s
      retries: 5

  db:
    image: postgres:17-alpine
    restart: unless-stopped
    volumes:
      - pg-data:/var/lib/postgresql/data
    environment:
      POSTGRES_DB: i1828
      POSTGRES_USER: i1828
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:?Set POSTGRES_PASSWORD in Dokploy env}
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U i1828"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  pg-data:
    driver: local
```

**Naming choice: `i1828`** for DB name + role. Short, alphanumeric (Postgres rejects leading digits), unambiguous. Distinct from `stewards` (pg-ai-stewards) and `becoming` (becoming-app).

**Why expose only the frontend's port 80?** Backend is reached by the frontend via Docker's internal DNS (`http://backend:8080`); no need to expose it on the host or open it on Traefik. DB is internal-only by default. Smaller attack surface; one route in.

## III. The split Dockerfile (two files, both small)

The current single `Dockerfile` becomes two — `Dockerfile.frontend` (almost unchanged) and `Dockerfile.backend` (new).

### `Dockerfile.frontend`

```dockerfile
# Stage 1 — Node build (unchanged from today)
FROM node:22-alpine AS build
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm install --no-audit --no-fund
COPY frontend/ ./
RUN npm run build

# Stage 2 — nginx serving + reverse-proxy to backend
FROM nginx:1.27-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/frontend/dist /usr/share/nginx/html
EXPOSE 80
HEALTHCHECK --interval=30s --timeout=3s CMD wget -qO- http://localhost/healthz >/dev/null || exit 1
```

The current `nginx.conf` gains one `location` block (the `/api/*` proxy_pass). Otherwise unchanged — SPA fallback, asset cache, healthcheck, gzip all stay.

### `Dockerfile.backend`

```dockerfile
FROM golang:1.23-alpine AS build
WORKDIR /src
# Embedded data (scripture corpus zip, 1828 gz, tier-words seed, modern-defs seed)
COPY backend/ ./
RUN go mod download
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/i1828 ./cmd/server

FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/i1828 /i1828
EXPOSE 8080
USER nonroot:nonroot
HEALTHCHECK --interval=30s --timeout=5s CMD ["/i1828", "healthcheck"]
ENTRYPOINT ["/i1828"]
```

Distroless static — ~5MB base, no shell, no package manager, no apt/apk surface. Adds the ~150-200MB of embedded seed data (scripture zip + 1828 gz + JSON seeds) for a final image around 200-220MB. Acceptable; smaller than `nginx:alpine + Vue dist` already in production.

## IV. The updated nginx.conf

Add one block to today's config:

```nginx
# /api/* → backend (same-origin from the browser's perspective; no CORS)
location /api/ {
    proxy_pass http://backend:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_read_timeout 75s;     # > the LLM_TIMEOUT_SECONDS default of 60s
    proxy_send_timeout 75s;
    # No cache; API responses are dynamic
    proxy_cache off;
}

# /healthz already exists — keep it.
# SPA fallback (try_files $uri $uri/ /index.html) already exists — keep it.
# /assets/ long-cache already exists — keep it.
```

That's the entirety of the nginx-side change. Same origin, no CORS dance, no preflight overhead.

## V. Dokploy migration path

Dokploy supports both project types: **Application** (single Dockerfile) and **Compose** (docker-compose.yaml). Migration is:

1. **Prepare.** Commit the new `Dockerfile.frontend`, `Dockerfile.backend`, `docker-compose.yaml` to `main`. The old `Dockerfile` stays in tree as `Dockerfile.legacy` for one deploy cycle (the rollback escape hatch). The frontend still builds + runs from `Dockerfile.legacy` if Dokploy is pointed at it.
2. **In Dokploy UI.** Create a *new* Compose project pointed at the same repo. Don't delete the existing Application project yet — both coexist briefly on different internal endpoints.
3. **Set env vars in Dokploy.** `POSTGRES_PASSWORD`, `LLM_*`, `MODERN_FETCH_DAILY_CAP`, `THUMMIM_*`. The Dokploy env-var UI generates a `.env` file the compose run picks up.
4. **Deploy the Compose project.** Verify `1828-staging.ibeco.me` (or whichever Dokploy-provided staging URL) serves the site correctly.
5. **Swap domains.** Detach `1828.ibeco.me` from the Application project; attach to the Compose project. DNS doesn't change; Traefik does.
6. **Decommission the Application project** after a 24-hour soak. Delete `Dockerfile.legacy`. Tag a release.

If anything goes wrong at step 4 or 5, the rollback is to re-attach the domain to the Application project. Cost of rollback: ~3 minutes of Dokploy UI work, zero data loss (compose volume is detached, not destroyed).

## VI. Local dev story

```bash
# One-time: create .env from .env.example
cp .env.example .env
# Edit POSTGRES_PASSWORD, LLM_PROVIDER, etc.

# Start everything
docker compose up -d

# Logs
docker compose logs -f backend
docker compose logs -f frontend

# Backend live reload during development (optional — use `air` or similar)
# Frontend live reload — already the existing `npm run dev` workflow,
# pointing at the running backend via VITE_API_BASE_URL=http://localhost/api
```

The `frontend/` directory's `npm run dev` workflow (Vite hot reload on port 5173) keeps working **alongside** the compose stack:
- Vite serves `index.html` + HMR on 5173.
- The frontend's API client uses `VITE_API_BASE_URL` (defaults to `/api` in production builds, `http://localhost/api` in dev). The dev value hits the nginx in the compose stack which proxies to the backend container.
- Result: edit Vue, see it instantly; backend changes still need a `docker compose build backend && docker compose up -d backend`.

For Michael specifically, two `.env` variants:
- `.env.local` — `LLM_BASE_URL=http://host.docker.internal:1234/v1` so the backend container reaches LM Studio on the host.
- `.env.production` — set in Dokploy, points at whatever deploy-time LLM provider is chosen.

## VII. CLAUDE.md / README.md updates after this lands

`projects/1828-illuminated/CLAUDE.md` `## Build + deploy` section gets rewritten:

```markdown
## Build + deploy

Three containers under one compose file:
- frontend  (nginx + Vue static dist)
- backend   (Go binary, embeds scripture corpus + 1828 dict + seed data)
- db        (Postgres 17-alpine, named volume `pg-data`)

Local: `docker compose up -d` after `cp .env.example .env`
Production: Dokploy's Compose project type, env vars in Dokploy UI.
```

The line "**No backend in MVP. Static site.**" gets replaced (the backend-pivot.md line-change manifest covers this). The line "**No LM Studio load.**" gets clarified to say the new backend's LLM proxy is the user's choice; gospel-engine-v2's LM Studio is unaffected.

`README.md` gets a section on local dev with the env-var setup.

## VIII. Backup + disaster recovery

The DB has two kinds of data:

- **Reproducible** — `scripture_books/chapters/verses`, `webster_1828`, `tier_words`, the seed portion of `modern_defs`. All seeded from embedded files in the backend image. A `docker compose down -v && docker compose up -d` rebuilds them in minutes.
- **Non-reproducible** — the *accumulated* portion of `modern_defs` (lazy-fetched words the seed didn't have), `thummim_entries_cache` (snapshot, but the timing of snapshot matters), `verse_highlights_cache` (rebuildable but slow).

The recovery plan:

1. **Weekly `pg_dump`** of the `i1828` database to a workspace-local path. Run by a cronjob *outside* the compose stack — on the host. Or, if Dokploy's host doesn't expose cron, a `cron` service inside the compose file.
2. **Dokploy volume snapshot** (their UI offers volume backups if configured). Belt-and-suspenders.
3. **Disaster recovery rehearsal**, one time: spin up a fresh compose stack, restore from the dump, verify the site works. Document in CLAUDE.md. Required for any system with durable state.

This is the project's first encounter with durable state. The covenant value `read_fully` applies metaphorically to data — Michael should know where his bytes live and how to recover them, not trust Dokploy to do the right thing silently.

## IX. Decisions

| # | Decision | Default | Stakes |
|---|---|---|---|
| **D-DS-1** | Compose project type in Dokploy vs Application + sidecar containers | Compose | Compose is the cleaner pattern; matches becoming-app |
| **D-DS-2** | Postgres-17-alpine vs separately-managed Postgres (e.g. shared with becoming-app) | 17-alpine, dedicated | Isolation; independent lifecycles; modest extra cost |
| **D-DS-3** | Distroless backend image vs alpine | Distroless static | Smaller; no shell surface; reasonable for a Go binary |
| **D-DS-4** | Backup cadence | Weekly pg_dump + Dokploy snapshot | Conservative; modern-def cache is the only loss surface |
| **D-DS-5** | Backup target | Workspace-local path that's gitignored | Recoverable from a re-clone if needed |
| **D-DS-6** | `Dockerfile.legacy` retention period | 1 deploy cycle (~1 week) | Rollback escape hatch |
| **D-DS-7** | Expose backend port to host (debug/dev) | No (container-internal only); add `127.0.0.1:8081:8080` only for local dev profile | Defense in depth |
| **D-DS-8** | Single `.env` vs Dokploy UI env vars in production | Dokploy UI in production; `.env` in local dev | Standard pattern; secrets stay out of git |
| **D-DS-9** | Dokploy "auto-deploy on push to main" enabled | Yes (matches today's posture) | Same as MVP — Michael's existing rhythm |

## X. Verification

After this phase ships:

- `docker compose up -d` from a fresh checkout boots all three services. Healthchecks pass within 60s.
- `curl http://localhost/healthz` returns `ok` (nginx).
- `curl http://localhost/api/healthz` returns `ok` (backend, through proxy).
- `docker compose down` stops the stack; `docker compose up -d` brings it back without data loss.
- `docker compose down -v` wipes the volume — verified once during the disaster-recovery rehearsal, then never again in production.
- A `pg_dump` against the running stack succeeds and produces a restorable artifact.
- Dokploy's Compose project deploys, serves, and rolls back cleanly. `1828.ibeco.me` resolves correctly.

## XI. Risks

- **Dokploy Compose support quirks.** Becoming-app uses this pattern successfully. If Dokploy has a regression, `Dockerfile.legacy` is the rollback. Low risk.
- **Compose `depends_on: service_healthy` race.** Docker Compose v2 supports this; older versions don't. Dokploy uses modern Compose. Verify version on the deploy host.
- **Port 80 contention.** If Dokploy's Traefik already handles port 80, the `ports: 80:80` may need to become `expose: 80` so Traefik proxies in. Becoming-app's compose handles this; check the Dokploy docs at integration time.
- **Distroless debugging.** No shell means no `docker exec sh` into the backend container for live troubleshooting. Mitigate with structured logs, healthcheck command via `/i1828 healthcheck`, and an emergency `Dockerfile.backend.debug` that uses alpine instead of distroless if a deep-debug session is needed.
- **`docker compose down -v` muscle memory.** Real risk — Michael ran into this in pg-ai-stewards. Mitigation: README.md and CLAUDE.md both warn explicitly. Long-term mitigation: a wrapper script (`scripts/safe-down.sh`) that refuses `-v` without an `--i-mean-it` flag.

## XII. Out of scope

- **Kubernetes / k3s.** Compose is the right scale for this project.
- **HA Postgres / replicas.** Single instance is fine for 1828's traffic; if it becomes a problem we'll know.
- **CDN in front.** Dokploy's Traefik is enough; Cloudflare can layer in later if traffic warrants.
- **Multi-region deploy.** N/A.
- **Blue/green deploys.** Dokploy's existing redeploy semantics (build new image, swap, drain old) are sufficient.
