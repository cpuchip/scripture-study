#!/bin/sh
# Bridge container entrypoint — runs migrations then the bridge daemon.
#
# Migration ledger (h-ledger-1+) requires the repo to be bind-mounted
# at /workspace so stewards-cli can discover extension/*.sql files.
# The docker-compose `bridge` service already mounts ../../..:/workspace:ro
# (the repo root read-only) per the h3-followup-1 fs-read scope work.
#
# Failure modes:
#   - /workspace missing: skip migrations, log warning, start bridge anyway
#     (allows the bridge to come up even without the host repo mounted —
#      operator can run migrations manually via docker exec)
#   - migrate command fails: exit non-zero so the operator sees it.
#     Operator fixes + docker compose restart bridge.
set -e

if [ -d /workspace/projects/pg-ai-stewards/extension ]; then
    echo "bridge-entrypoint: running substrate migrations…"
    /usr/local/bin/stewards-cli migrate --repo-root /workspace
    echo "bridge-entrypoint: migrations done."
else
    echo "bridge-entrypoint: WARNING /workspace not mounted; skipping migrations."
    echo "bridge-entrypoint: run 'docker exec pg-ai-stewards-bridge stewards-cli migrate' manually if needed."
fi

echo "bridge-entrypoint: starting bridge daemon…"
exec /usr/local/bin/stewards-mcp bridge run
