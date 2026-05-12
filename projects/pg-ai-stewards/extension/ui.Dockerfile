# =====================================================================
# stewards-ui image — local web UI for pg-ai-stewards substrate
# (Phase 3f v1, 2026-05-09).
#
# Multi-stage:
#   1. node:lts-alpine — npm install + vite build → frontend/dist/
#   2. golang:1.26-alpine — go build with embed.FS containing the dist
#   3. alpine:3.20 — slim runtime, ca-certificates, single binary
#
# Build context: repo root (../../..) for access to scripts/ and the
# go.work workspace. See projects/pg-ai-stewards/extension/docker-compose.yaml
# service `ui` (compose sets context).
#
# Image emits one binary at /usr/local/bin/stewards-ui that serves both
# the Vue SPA and the JSON API at /api/* on a single port (default 8080).
# =====================================================================

# ---------------------------------------------------------------------
# Stage 1 — frontend builder. node + npm + vite + tsc.
# ---------------------------------------------------------------------
FROM node:lts-alpine AS frontend

WORKDIR /frontend

# Copy manifests first for layer caching.
COPY scripts/stewards-ui/frontend/package.json ./package.json

# Vite + Tailwind 4 install. No package-lock.json yet (first build);
# subsequent builds get one and we'll switch to `npm ci`.
RUN npm install --no-audit --no-fund

# Now copy the source files and build.
COPY scripts/stewards-ui/frontend/ ./

RUN npm run build

# ---------------------------------------------------------------------
# Stage 2 — Go builder. Embeds the built dist into a static binary.
# ---------------------------------------------------------------------
FROM golang:1.26-alpine AS gobuilder

RUN apk add --no-cache git ca-certificates
WORKDIR /workspace

# Workspace + module sources for stewards-ui.
COPY go.work go.work.sum ./
COPY scripts/stewards-ui/ ./scripts/stewards-ui/

# Stub-out workspace siblings (go workspace mode requires every listed
# module to resolve). Same shape as bridge.Dockerfile.
COPY scripts/brain/go.mod                                  ./scripts/brain/go.mod
COPY scripts/brain/go.sum                                  ./scripts/brain/go.sum
COPY scripts/embedding-compare/go.mod                      ./scripts/embedding-compare/go.mod
COPY scripts/embedding-compare/go.sum                      ./scripts/embedding-compare/go.sum
COPY scripts/gospel-engine/go.mod                          ./scripts/gospel-engine/go.mod
COPY scripts/gospel-engine/go.sum                          ./scripts/gospel-engine/go.sum
COPY scripts/chromem-exp/go.mod                            ./scripts/chromem-exp/go.mod
COPY scripts/chromem-exp/go.sum                            ./scripts/chromem-exp/go.sum
COPY scripts/gospel-library/go.mod                         ./scripts/gospel-library/go.mod
COPY scripts/gospel-library/go.sum                         ./scripts/gospel-library/go.sum
COPY scripts/gospel-mcp/go.mod                             ./scripts/gospel-mcp/go.mod
COPY scripts/gospel-vec/go.mod                             ./scripts/gospel-vec/go.mod
COPY scripts/gospel-vec/go.sum                             ./scripts/gospel-vec/go.sum
# lectures-on-faith and publish have no go.sum (no external deps).
COPY scripts/lectures-on-faith/go.mod                      ./scripts/lectures-on-faith/go.mod
COPY scripts/publish/go.mod                                ./scripts/publish/go.mod
COPY scripts/session-journal/go.mod                        ./scripts/session-journal/go.mod
COPY scripts/session-journal/go.sum                        ./scripts/session-journal/go.sum
COPY scripts/study-export/go.mod                           ./scripts/study-export/go.mod
COPY scripts/study-export/go.sum                           ./scripts/study-export/go.sum
COPY scripts/git-mcp/go.mod                                ./scripts/git-mcp/go.mod
COPY scripts/git-mcp/go.sum                                ./scripts/git-mcp/go.sum
COPY scripts/fetch-md-mcp/go.mod                           ./scripts/fetch-md-mcp/go.mod
COPY scripts/fetch-md-mcp/go.sum                           ./scripts/fetch-md-mcp/go.sum
COPY scripts/webster-mcp/go.mod                            ./scripts/webster-mcp/go.mod
COPY scripts/webster-mcp/go.sum                            ./scripts/webster-mcp/go.sum
COPY scripts/byu-citations/go.mod                          ./scripts/byu-citations/go.mod
COPY scripts/byu-citations/go.sum                          ./scripts/byu-citations/go.sum
# yt-mcp has no go.sum (no external deps).
COPY scripts/yt-mcp/go.mod                                 ./scripts/yt-mcp/go.mod
COPY scripts/search-mcp/go.mod                             ./scripts/search-mcp/go.mod
COPY scripts/search-mcp/go.sum                             ./scripts/search-mcp/go.sum
COPY scripts/becoming/go.mod                               ./scripts/becoming/go.mod
COPY scripts/becoming/go.sum                               ./scripts/becoming/go.sum
COPY projects/pg-ai-stewards/cmd/fs-read-mcp/go.mod        ./projects/pg-ai-stewards/cmd/fs-read-mcp/go.mod
COPY projects/pg-ai-stewards/cmd/fs-read-mcp/go.sum        ./projects/pg-ai-stewards/cmd/fs-read-mcp/go.sum
COPY projects/pg-ai-stewards/cmd/stewards-cli/go.mod       ./projects/pg-ai-stewards/cmd/stewards-cli/go.mod
COPY projects/pg-ai-stewards/cmd/stewards-cli/go.sum       ./projects/pg-ai-stewards/cmd/stewards-cli/go.sum
COPY projects/pg-ai-stewards/cmd/stewards-mcp/go.mod       ./projects/pg-ai-stewards/cmd/stewards-mcp/go.mod
COPY projects/pg-ai-stewards/cmd/stewards-mcp/go.sum       ./projects/pg-ai-stewards/cmd/stewards-mcp/go.sum
COPY external_context/tpg/go.mod                           ./external_context/tpg/go.mod
COPY external_context/tpg/go.sum                           ./external_context/tpg/go.sum
COPY experiments/lm-studio/scripts/scoring/go.mod          ./experiments/lm-studio/scripts/scoring/go.mod
COPY experiments/lm-studio/scripts/scoring/go.sum          ./experiments/lm-studio/scripts/scoring/go.sum

# Replace the stub dist with the freshly-built one from stage 1.
COPY --from=frontend /frontend/dist ./scripts/stewards-ui/frontend/dist

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

RUN cd scripts/stewards-ui \
    && go build -trimpath -ldflags="-s -w" -o /out/stewards-ui .

# ---------------------------------------------------------------------
# Stage 3 — runtime. Slim alpine + ca-certificates.
# ---------------------------------------------------------------------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

COPY --from=gobuilder /out/stewards-ui /usr/local/bin/stewards-ui

# Default DSN points at the compose service name `pg`.
ENV STEWARDS_DSN="postgres://stewards:stewards@pg:5432/stewards?sslmode=disable"

# Bind to 0.0.0.0 inside the container; compose maps to 127.0.0.1
# on the host for local-only access (see docker-compose.yaml).
ENTRYPOINT ["/usr/local/bin/stewards-ui", "--addr", "0.0.0.0:8080"]
EXPOSE 8080
