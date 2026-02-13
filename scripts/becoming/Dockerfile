# Multi-stage build for the Becoming app
# Stage 1: Build frontend
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY scripts/becoming/frontend/package*.json ./
RUN npm ci
COPY scripts/becoming/frontend/ ./
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.25-alpine AS backend
# CGO needed for SQLite (mattn/go-sqlite3). For PostgreSQL-only deploy, set CGO_ENABLED=0.
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY scripts/becoming/go.mod scripts/becoming/go.sum ./
RUN go mod download
COPY scripts/becoming/ ./
# Copy frontend dist into embed location
COPY --from=frontend /app/frontend/dist ./cmd/server/dist/
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server/

# Stage 3: Minimal runtime
FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=backend /server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
