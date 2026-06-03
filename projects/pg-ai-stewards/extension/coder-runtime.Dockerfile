# =====================================================================
# coder-runtime — the sandbox image the substrate's coder spawns per
# work_item (substrate-coding-capability proposal, CC.1).
#
# D-CC4: Go + Node/TypeScript + Python, plus each language's LSP server
# (gopls / typescript-language-server / pyright). Non-root `coder` user.
# Runs idle (`sleep infinity`); the sandbox-manager `docker exec`s build/
# test/run + the coder tools into it, then tears it down (ephemeral).
#
# Built on the HOST docker daemon (the bridge spawns siblings against it
# via the mounted socket):
#   docker build -f projects/pg-ai-stewards/extension/coder-runtime.Dockerfile \
#                -t coder-runtime:latest projects/pg-ai-stewards/extension
# =====================================================================

# Version-safe toolchains: copy from the official images rather than
# pinning download URLs that rot.
FROM golang:1.26-bookworm AS go
FROM node:24-bookworm AS node

FROM debian:bookworm-slim

# Base OS tooling + Python.
RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates git curl openssh-client \
        python3 python3-pip python3-venv \
        build-essential pkg-config \
    && rm -rf /var/lib/apt/lists/*

# Go (from the official image).
COPY --from=go /usr/local/go /usr/local/go
ENV PATH=/usr/local/go/bin:/usr/local/bin:$PATH \
    GOPATH=/home/coder/go \
    GOTOOLCHAIN=local

# Node + npm (from the official image; recreate the npm/npx shims).
COPY --from=node /usr/local/bin/node /usr/local/bin/node
COPY --from=node /usr/local/lib/node_modules /usr/local/lib/node_modules
RUN ln -sf /usr/local/lib/node_modules/npm/bin/npm-cli.js /usr/local/bin/npm \
    && ln -sf /usr/local/lib/node_modules/npm/bin/npx-cli.js /usr/local/bin/npx

# LSP servers (CC.4 uses these; bake them now so the image is built once).
#  - gopls            : Go
#  - typescript + typescript-language-server : JS/TS
#  - pyright          : Python (npm-based, avoids the bookworm pip externally-managed dance)
RUN GOBIN=/usr/local/bin go install golang.org/x/tools/gopls@latest \
    && npm install -g --no-fund --no-audit \
        typescript typescript-language-server pyright \
    && rm -rf /root/.npm /root/go/pkg/mod /root/.cache

# Put `go`/`gofmt` on the standard PATH (/usr/local/bin) so they resolve in
# login shells too — `bash -lc` sources /etc/profile, which resets PATH and
# would otherwise drop the /usr/local/go/bin ENV entry.
RUN ln -sf /usr/local/go/bin/go /usr/local/bin/go \
    && ln -sf /usr/local/go/bin/gofmt /usr/local/bin/gofmt

# Non-root user. The isolation boundary is the container, so inside it the
# coder may act freely; non-root is defense-in-depth, not the main wall.
RUN useradd -m -u 1000 -s /bin/bash coder \
    && mkdir -p /work /home/coder/go \
    && chown -R coder:coder /work /home/coder
USER coder
WORKDIR /work

# Idle by default; the sandbox-manager execs commands in. Overridable.
CMD ["sleep", "infinity"]
