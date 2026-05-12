# =====================================================================
# Bridge image — runs `stewards-mcp bridge run` as a long-lived daemon
# alongside the pg-ai-stewards Postgres container. Bundles the bridge
# itself plus all spawn-target MCP server binaries so the container
# owns the full outbound MCP surface.
#
# Build context: repository root (for access to scripts/ and projects/).
# Compose sets context: ../.. — see projects/pg-ai-stewards/extension/
# docker-compose.yaml service `bridge`.
#
# Phase: bridge-in-docker (2026-05-09). Replaces the host-shell-only
# pattern from 3e.2.b/c v1. The Windows .exe paths in mcp_servers.command
# are migrated to /usr/local/bin/* by 3e2-6-mcp-servers-linux-paths.sql.
# =====================================================================

# ---------------------------------------------------------------------
# Stage 1 — builder. Cross-compile every Go binary the bridge spawns.
# ---------------------------------------------------------------------
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /workspace

# go.work + module sources. We COPY each module independently to avoid
# pulling huge sibling assets (becoming/frontend/node_modules etc.).
COPY go.work go.work.sum ./

# Bridge itself + tools it ships
COPY projects/pg-ai-stewards/cmd/stewards-mcp/ ./projects/pg-ai-stewards/cmd/stewards-mcp/
COPY projects/pg-ai-stewards/cmd/fs-read-mcp/  ./projects/pg-ai-stewards/cmd/fs-read-mcp/
# stewards-cli copied fully (not just go.mod stub) so the migrate
# command can be built into the bridge image and run from the entrypoint.
COPY projects/pg-ai-stewards/cmd/stewards-cli/ ./projects/pg-ai-stewards/cmd/stewards-cli/

# Spawn targets
COPY scripts/fetch-md-mcp/      ./scripts/fetch-md-mcp/
COPY scripts/git-mcp/           ./scripts/git-mcp/
COPY scripts/webster-mcp/       ./scripts/webster-mcp/
COPY scripts/byu-citations/     ./scripts/byu-citations/
COPY scripts/yt-mcp/            ./scripts/yt-mcp/
COPY scripts/search-mcp/        ./scripts/search-mcp/
COPY scripts/gospel-engine-v2/  ./scripts/gospel-engine-v2/

# becoming/ is large; bring only what the cmd/mcp build needs.
COPY scripts/becoming/go.mod         ./scripts/becoming/go.mod
COPY scripts/becoming/go.sum         ./scripts/becoming/go.sum
COPY scripts/becoming/cmd/mcp/       ./scripts/becoming/cmd/mcp/
COPY scripts/becoming/internal/      ./scripts/becoming/internal/

# Other workspace modules referenced by go.work — we don't build them
# but go won't tolerate missing module roots. Stub each with its
# go.mod only (Go's module resolver is satisfied by the manifest).
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
# lectures-on-faith and publish have no go.sum (no external deps);
# COPY pattern handles either presence.
COPY scripts/lectures-on-faith/go.mod                      ./scripts/lectures-on-faith/go.mod
COPY scripts/publish/go.mod                                ./scripts/publish/go.mod
COPY scripts/session-journal/go.mod                        ./scripts/session-journal/go.mod
COPY scripts/session-journal/go.sum                        ./scripts/session-journal/go.sum
COPY scripts/stewards-ui/go.mod                            ./scripts/stewards-ui/go.mod
COPY scripts/stewards-ui/go.sum                            ./scripts/stewards-ui/go.sum
COPY scripts/study-export/go.mod                           ./scripts/study-export/go.mod
COPY scripts/study-export/go.sum                           ./scripts/study-export/go.sum
# stewards-cli is now COPYed in full higher up (so it can be BUILT,
# not just satisfy go.work). Stub COPYs removed.
COPY external_context/tpg/go.mod                           ./external_context/tpg/go.mod
COPY external_context/tpg/go.sum                           ./external_context/tpg/go.sum
COPY experiments/lm-studio/scripts/scoring/go.mod          ./experiments/lm-studio/scripts/scoring/go.mod
COPY experiments/lm-studio/scripts/scoring/go.sum          ./experiments/lm-studio/scripts/scoring/go.sum

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Build the bridge first — its module is workspace-listed.
RUN cd projects/pg-ai-stewards/cmd/stewards-mcp \
    && go build -trimpath -ldflags="-s -w" -o /out/stewards-mcp .

# fs-read-mcp — H.1.7 path-scoped filesystem read MCP for the research agent.
RUN cd projects/pg-ai-stewards/cmd/fs-read-mcp \
    && go build -trimpath -ldflags="-s -w" -o /out/fs-read-mcp .

# stewards-cli — runs the migration ledger on startup (h-ledger-N batch).
# Also useful as an ad-hoc CLI inside the container.
RUN cd projects/pg-ai-stewards/cmd/stewards-cli \
    && go build -trimpath -ldflags="-s -w" -o /out/stewards-cli .

# Spawn targets in workspace
RUN cd scripts/fetch-md-mcp \
    && go build -trimpath -ldflags="-s -w" -o /out/fetch-md-mcp .
RUN cd scripts/git-mcp \
    && go build -trimpath -ldflags="-s -w" -o /out/git-mcp .
RUN cd scripts/webster-mcp \
    && go build -trimpath -ldflags="-s -w" -o /out/webster-mcp ./cmd/webster-mcp
RUN cd scripts/byu-citations \
    && go build -trimpath -ldflags="-s -w" -o /out/byu-citations ./cmd/byu-citations
RUN cd scripts/yt-mcp \
    && go build -trimpath -ldflags="-s -w" -o /out/yt-mcp .
RUN cd scripts/search-mcp \
    && go build -trimpath -ldflags="-s -w" -o /out/search-mcp ./cmd/search-mcp
RUN cd scripts/becoming \
    && go build -trimpath -ldflags="-s -w" -o /out/becoming-mcp ./cmd/mcp

# gospel-engine-v2 is NOT in go.work (it's a submodule with its own
# manifests). Build it outside workspace mode so its go.mod resolves
# without our workspace's `use ()` directive interfering.
RUN cd scripts/gospel-engine-v2 \
    && GOWORK=off go build -trimpath -ldflags="-s -w" -o /out/gospel-mcp ./cmd/gospel-mcp

# ---------------------------------------------------------------------
# Stage 2 — runtime. Slim alpine + binaries + data files.
# ---------------------------------------------------------------------
FROM alpine:3.20

# ca-certificates for HTTPS (gospel-engine, becoming, exa-search HTTP).
# git + github-cli (gh) for git-mcp's spawn targets. chromium for
# fetch-md-mcp v2's JS-rendering path (chromedp finds it via $PATH).
RUN apk add --no-cache ca-certificates tzdata git github-cli chromium

# Binaries
COPY --from=builder /out/stewards-mcp   /usr/local/bin/stewards-mcp
COPY --from=builder /out/fetch-md-mcp   /usr/local/bin/fetch-md-mcp
COPY --from=builder /out/git-mcp        /usr/local/bin/git-mcp
COPY --from=builder /out/webster-mcp    /usr/local/bin/webster-mcp
COPY --from=builder /out/byu-citations  /usr/local/bin/byu-citations
COPY --from=builder /out/yt-mcp         /usr/local/bin/yt-mcp
COPY --from=builder /out/search-mcp     /usr/local/bin/search-mcp
COPY --from=builder /out/becoming-mcp   /usr/local/bin/becoming-mcp
COPY --from=builder /out/gospel-mcp     /usr/local/bin/gospel-mcp
COPY --from=builder /out/fs-read-mcp    /usr/local/bin/fs-read-mcp
COPY --from=builder /out/stewards-cli   /usr/local/bin/stewards-cli

# Data files — webster needs the 1828 dictionary at a known path. The
# mcp_servers seed (3e2-6) points args at /opt/webster/data/.
RUN mkdir -p /opt/webster/data
COPY scripts/webster-mcp/data/webster1828.json.gz /opt/webster/data/webster1828.json.gz

# yt-mcp needs writable workdirs; create empty ones so the binary can
# start. yt-mcp ships disabled in the seed because organic use needs
# host-side YT downloads — operator flips enabled when ready.
RUN mkdir -p /opt/yt/yt /opt/yt/study

# Default DSN points at the compose service name `pg`.
# Tokens are injected via env from compose (.env file at extension/.env).
ENV STEWARDS_DSN="postgres://stewards:stewards@pg:5432/stewards?sslmode=disable"

# Entrypoint script: runs migrations FIRST (substrate auto-current on
# every bridge restart), then execs the bridge daemon. The repo is
# bind-mounted at /workspace by docker-compose so stewards-cli can
# read extension/*.sql for migration discovery.
#
# Failure mode: if migrations fail, the bridge does NOT start —
# substrate would be in an inconsistent state. Log + exit non-zero.
# Operator inspects logs and re-runs after fixing.
COPY projects/pg-ai-stewards/extension/bridge-entrypoint.sh /usr/local/bin/bridge-entrypoint.sh
RUN chmod +x /usr/local/bin/bridge-entrypoint.sh

ENTRYPOINT ["/usr/local/bin/bridge-entrypoint.sh"]
