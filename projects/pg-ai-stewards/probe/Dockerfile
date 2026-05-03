# pg18 + pgvector + apache AGE in one image, bookworm.
# Ours, not Azure's.
FROM pgvector/pgvector:pg18

ARG AGE_REF=release/PG18/1.7.0
ENV DEBIAN_FRONTEND=noninteractive

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
