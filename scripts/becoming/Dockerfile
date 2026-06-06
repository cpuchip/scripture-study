# Multi-stage build for the Becoming app.
#
# Scripture lookup used to ship a bundled ~20MB markdown tree (former
# `scriptures` stage). 2026-05-23: dropped — the running app fetches
# verses from gospel-engine-v2 (engine.ibeco.me) at runtime via
# engineClient with a small process-local LRU cache. See
# scripts/becoming/.spec/proposals/scripture-via-engine.md.
#
# Soft-dependency posture: when engine is unreachable, the lookup
# endpoint returns 503 cleanly and the UI degrades gracefully —
# books + name-search continue to work (they're metadata-only).
# This shrinks the image by ~20MB AND removes one heavy build stage
# from the deploy path.

# Stage 1: Build frontend
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY scripts/becoming/frontend/package*.json ./
RUN npm ci
COPY scripts/becoming/frontend/ ./
RUN npm run build

# Stage 2: Extract git metadata from build context
FROM alpine:3.21 AS gitinfo
RUN apk add --no-cache git
WORKDIR /repo
COPY .git .git
RUN git rev-parse --short HEAD > /commit_hash && \
    git log -1 --format=%s > /commit_msg

# Stage 3: Build Go binary
FROM golang:1.26-alpine AS backend
WORKDIR /app
COPY scripts/becoming/go.mod scripts/becoming/go.sum ./
RUN go mod download
COPY scripts/becoming/ ./
# Copy frontend dist into embed location
COPY --from=frontend /app/frontend/dist ./cmd/server/dist/
# Git metadata from build context (auto-detected, no build args needed)
COPY --from=gitinfo /commit_hash /commit_msg /tmp/
ARG VERSION=dev
ARG RELEASE_NOTES=""
# NOTES is sanitized (tr -d '\047\042' strips ' and ") before it reaches ldflags:
# an apostrophe in the commit subject (e.g. "Michael's") would close the
# -X 'main.ReleaseNotes=…' grouping that Go's quoted.Split relies on, handing the
# linker garbage flags and failing the build (it dumps `usage: link …`, exit 1).
RUN COMMIT=$(cat /tmp/commit_hash) && \
    MSG=$(cat /tmp/commit_msg) && \
    NOTES=$(printf '%s' "${RELEASE_NOTES:-$MSG}" | tr -d '\047\042') && \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.CommitHash=${COMMIT} -X 'main.ReleaseNotes=${NOTES}'" \
    -o /server ./cmd/server/

# Stage 4: Minimal runtime
FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=backend /server /server
EXPOSE 8080
# Scripture lookup config — engine.ibeco.me at runtime. ENGINE_SERVICE_TOKEN
# is required for the lookup endpoint; without it the app boots cleanly and
# only the lookup path returns 503 (books + search still serve).
ENV ENGINE_URL=https://engine.ibeco.me
# ENGINE_SERVICE_TOKEN injected per-deploy via the Dokploy env UI.
ENTRYPOINT ["/server"]
