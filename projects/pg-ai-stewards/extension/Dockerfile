# projects/pg-ai-stewards/extension/Dockerfile
#
# Multi-stage build for the pg_ai_stewards extension.
# Stage 1: build the extension with cargo-pgrx against PG18 dev headers.
# Stage 2: layer the extension onto pgvector + Apache AGE on PG18.
#
# This deliberately mirrors the probe/ Dockerfile rather than FROM-ing it,
# so the two stacks can evolve independently. If duplication ever bites,
# refactor to a shared Dockerfile.base.

# ============================================================
# Stage 1: builder — Rust + cargo-pgrx + PG18 server-dev
# ============================================================
FROM rust:1-bookworm AS builder

ARG PGRX_VERSION=0.18.0
ENV DEBIAN_FRONTEND=noninteractive

# PG18 dev headers from the official PGDG repository (bookworm).
RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates curl gnupg lsb-release \
        build-essential pkg-config \
        libclang-dev clang \
        libreadline-dev zlib1g-dev libssl-dev \
    && curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc \
       | gpg --dearmor -o /etc/apt/trusted.gpg.d/postgresql.gpg \
    && echo "deb http://apt.postgresql.org/pub/repos/apt bookworm-pgdg main" \
       > /etc/apt/sources.list.d/pgdg.list \
    && apt-get update \
    && apt-get install -y --no-install-recommends \
        postgresql-18 postgresql-server-dev-18 \
    && rm -rf /var/lib/apt/lists/*

RUN cargo install --locked cargo-pgrx --version ${PGRX_VERSION}

# Register the system PG18 with pgrx. --no-run means "don't initdb a
# managed cluster" — we only need cargo-pgrx to know where pg_config is
# so it can package against PG18's headers and lib paths.
RUN cargo pgrx init --pg18 /usr/lib/postgresql/18/bin/pg_config

WORKDIR /build/pg_ai_stewards
COPY Cargo.toml pg_ai_stewards.control ./
COPY src ./src
# The 2-6*/2-7*/3a*.sql files are read at compile time by
# extension_sql_file! macros in src/lib.rs (folded back from live-DB
# migrations as of v0.2.0). They must be present in the build context.
# When adding a new file, also add the corresponding extension_sql_file!
# call in src/lib.rs and update this list.
COPY 2-6a-workstreams.sql 2-6b-todos.sql 2-6c-phases-context.sql \
     2-7a-watchman-substrate.sql 3a-watchman-pass.sql \
     2-7b1-watchman-automation.sql 2-7b2-watchman-scheduler.sql \
     2-7b3-watchman-budget.sql 2-7b4-watchman-soak-prep.sql ./

# `cargo pgrx package` produces ./target/release/pg_ai_stewards-pg18/
# with the directory tree rooted at /, containing the .so, .control,
# and generated .sql under the standard pg_config paths.
RUN cargo pgrx package \
        --pg-config /usr/lib/postgresql/18/bin/pg_config \
        --out-dir /out

# ============================================================
# Stage 2: runtime — pgvector + AGE + pg_ai_stewards on PG18
# ============================================================
FROM pgvector/pgvector:pg18

ARG AGE_REF=release/PG18/1.7.0
ENV DEBIAN_FRONTEND=noninteractive

# Build AGE from source (matches probe/Dockerfile).
RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates curl git \
        build-essential make clang \
        postgresql-server-dev-18 \
        libreadline-dev zlib1g-dev flex bison \
    && git clone --depth 1 --branch ${AGE_REF} https://github.com/apache/age.git /tmp/age \
    && cd /tmp/age \
    && make install \
    && rm -rf /tmp/age \
    && apt-get purge -y --auto-remove build-essential clang make git \
        postgresql-server-dev-18 libreadline-dev zlib1g-dev flex bison \
    && rm -rf /var/lib/apt/lists/*

# Drop in the pgrx-built extension. Stage 1's --out-dir output is rooted
# directly at / (the .so under /usr/lib/postgresql/18/lib/, control +
# generated SQL under /usr/share/postgresql/18/extension/).
COPY --from=builder /out/ /
