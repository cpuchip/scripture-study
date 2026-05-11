#!/bin/bash
# Phase 5d (Lesson #3 fix): refresh pg_extern function registrations
# after a docker compose build pg.
#
# Why this exists:
#   docker compose build pg + docker compose down/up replaces the
#   container's .so but leaves the pg data volume intact. CREATE
#   EXTENSION is a no-op (extension already installed at version X),
#   so any new #[pg_extern] functions in the .so have no row in
#   pg_proc — calling them fails with "function does not exist".
#
# Strategy (dev-mode):
#   1. Read the bundled SQL inside the container
#      (/usr/share/postgresql/18/extension/pg_ai_stewards--<v>.sql)
#   2. Extract pgrx-generated CREATE FUNCTION blocks (recognizable by
#      `-- pg_ai_stewards::module::name` comment + `CREATE  FUNCTION`)
#   3. Rewrite CREATE FUNCTION → CREATE OR REPLACE FUNCTION
#   4. Apply against the live db with search_path = stewards, public
#
# What this DOESN'T do (intentionally):
#   - Doesn't bump the extension version (no upgrade script written)
#   - Doesn't touch tables/types/seeds (CREATE TABLE in the bundled
#     SQL is non-idempotent for early-phase tables; trying to re-run
#     it fails. New tables ship via separate 5*.sql files applied via
#     the substrate's existing docker cp + psql pattern.)
#   - Doesn't make the new functions extension members in pg_depend
#     (functional but slightly drifty for production; fine for dev)
#
# For new SQL files (5*.sql), use the existing live-migration pattern:
#   docker cp <file>.sql pg-ai-stewards-dev:/tmp/X.sql
#   docker exec pg-ai-stewards-dev psql -U stewards -d stewards -f /tmp/X.sql

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
EXT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
CONTAINER="${STEWARDS_CONTAINER:-pg-ai-stewards-dev}"
# Use MSYS_NO_PATHCONV when invoking docker so Git Bash on Windows
# doesn't rewrite container-side paths into Windows paths.
export MSYS_NO_PATHCONV=1
EXT_PATH_IN_CONTAINER="/usr/share/postgresql/18/extension"

container_running() {
    docker ps --format '{{.Names}}' 2>/dev/null | grep -q "^${CONTAINER}$"
}

if ! container_running; then
    echo "bump-extension: ${CONTAINER} not running; skipping."
    exit 0
fi

# Discover bundled SQL inside the container
BUNDLED_FILE=$(docker exec "${CONTAINER}" bash -c \
    "ls ${EXT_PATH_IN_CONTAINER}/pg_ai_stewards--*.sql 2>/dev/null | grep -E 'pg_ai_stewards--[0-9]+\\.[0-9]+\\.[0-9]+\\.sql$' | sort -V | tail -1")

if [ -z "$BUNDLED_FILE" ]; then
    echo "bump-extension: no bundled SQL found in container" >&2
    exit 1
fi

BUNDLED_VERSION=$(echo "$BUNDLED_FILE" | sed -E 's|.*pg_ai_stewards--([0-9.]+)\.sql$|\1|')

INSTALLED_VERSION=$(docker exec "${CONTAINER}" psql -U stewards -d stewards -tAc \
    "SELECT extversion FROM pg_extension WHERE extname='pg_ai_stewards'" 2>/dev/null || echo "")

if [ -z "$INSTALLED_VERSION" ]; then
    echo "bump-extension: pg_ai_stewards not installed; nothing to refresh."
    exit 0
fi

echo "bump-extension: bundled-in-container=${BUNDLED_VERSION}, installed=${INSTALLED_VERSION}"

# Extract pg_extern CREATE FUNCTION blocks from the bundled SQL.
# Pattern (from pgrx 0.18):
#   /* <begin connected objects> */
#   -- src/<module>.rs:<line>
#   -- pg_ai_stewards::<module>::<name>
#   CREATE  FUNCTION "<name>"(
#       ...args...
#   ) RETURNS <type>
#   ...modifiers...
#   AS 'MODULE_PATHNAME', '<name>_wrapper';
#   /* </end connected objects> */
#
# We grab from `-- pg_ai_stewards::` to the next semicolon-terminated
# line that starts with AS '. Only blocks containing CREATE  FUNCTION
# (two spaces — pgrx signature) are emitted; we skip CREATE OR REPLACE
# blocks because those re-apply safely as part of the extension_sql!
# macro path or via docker cp.

EXTRACT_PY='
import re, sys
src = sys.stdin.read()
# Match each connected-objects block
blocks = re.findall(
    r"/\* <begin connected objects> \*/(.*?)/\* </end connected objects> \*/",
    src, re.DOTALL)
out = []
for b in blocks:
    if not re.search(r"^-- pg_ai_stewards::", b, re.MULTILINE):
        continue
    if not re.search(r"^CREATE  FUNCTION ", b, re.MULTILINE):
        continue
    # Rewrite CREATE  FUNCTION -> CREATE OR REPLACE FUNCTION
    b = re.sub(r"^CREATE  FUNCTION ", "CREATE OR REPLACE FUNCTION ",
               b, flags=re.MULTILINE)
    # Substitute MODULE_PATHNAME (pgrx token expanded by extension
    # machinery; we are running outside it so substitute manually).
    b = b.replace("MODULE_PATHNAME", "$libdir/pg_ai_stewards")
    out.append(b.strip())

if not out:
    sys.stderr.write("bump-extension: no pgrx CREATE FUNCTION blocks found in bundled SQL\n")
    sys.exit(2)

print("SET search_path = stewards, public;")
print("BEGIN;")
for b in out:
    print(b)
print("COMMIT;")
'

EXTRACT_SQL=$(docker exec "${CONTAINER}" cat "${BUNDLED_FILE}" | python3 -c "$EXTRACT_PY")

if [ -z "$EXTRACT_SQL" ]; then
    echo "bump-extension: extraction produced no SQL; bailing." >&2
    exit 1
fi

FN_COUNT=$(echo "$EXTRACT_SQL" | grep -cE "^CREATE OR REPLACE FUNCTION")
echo "bump-extension: refreshing ${FN_COUNT} pg_extern function(s)"

# Apply
echo "$EXTRACT_SQL" | docker exec -i "${CONTAINER}" psql -U stewards -d stewards \
    -v ON_ERROR_STOP=1 > /dev/null

echo "bump-extension: ${FN_COUNT} function(s) refreshed in ${CONTAINER}"
