# Multi-stage build for the Becoming app
# Stage 1: Build frontend
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY scripts/becoming/frontend/package*.json ./
RUN npm ci
COPY scripts/becoming/frontend/ ./
RUN npm run build

# Stage 2: Download scriptures for memorization lookup
# Uses the gospel-library downloader to fetch standard works (~20MB)
FROM golang:1.25-alpine AS scriptures
RUN apk add --no-cache git
WORKDIR /dl
COPY scripts/gospel-library/go.mod scripts/gospel-library/go.sum ./
RUN go mod download
COPY scripts/gospel-library/ ./
RUN CGO_ENABLED=0 go build -o /downloader ./cmd/gospel-downloader/
RUN /downloader -standard -output /gospel-library

# Stage 3: Extract git metadata from build context
FROM alpine:3.21 AS gitinfo
RUN apk add --no-cache git
WORKDIR /repo
COPY .git .git
RUN git rev-parse --short HEAD > /commit_hash && \
    git log -1 --format=%s > /commit_msg

# Stage 4: Build Go binary
FROM golang:1.25-alpine AS backend
# CGO needed for SQLite (mattn/go-sqlite3). For PostgreSQL-only deploy, set CGO_ENABLED=0.
RUN apk add --no-cache gcc musl-dev
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
RUN COMMIT=$(cat /tmp/commit_hash) && \
    MSG=$(cat /tmp/commit_msg) && \
    NOTES="${RELEASE_NOTES:-$MSG}" && \
    CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.CommitHash=${COMMIT} -X 'main.ReleaseNotes=${NOTES}'" \
    -o /server ./cmd/server/

# Stage 5: Minimal runtime
FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=backend /server /server
COPY --from=scriptures /gospel-library/eng/scriptures /scriptures
EXPOSE 8080
ENV BECOMING_SCRIPTURES=/scriptures
ENTRYPOINT ["/server"]
