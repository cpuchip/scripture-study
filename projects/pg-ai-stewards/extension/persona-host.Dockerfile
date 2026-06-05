# persona-host — the optional persona SIDECAR (cmd/persona-host). Runs in the
# substrate compose so it reaches the substrate DB (pg:5432) for cognition and
# connects OUT to the ai-chattermax platform gateway (wss://chat.ibeco.me).
# Build context = repo root (../../..); the module is self-contained.

FROM golang:1.26-alpine AS build
WORKDIR /src
COPY projects/pg-ai-stewards/cmd/persona-host/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /persona-host .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /persona-host /usr/local/bin/persona-host
# CHATTERMAX_GATEWAY + CHATTERMAX_PERSONAS come from the compose env_file (.env);
# STEWARDS_DSN points at the internal pg service.
ENTRYPOINT ["/usr/local/bin/persona-host", "-addr", ":8090"]
