# Becoming — Deployment Guide

## Prerequisites
- [Fly.io CLI](https://fly.io/docs/flyctl/install/) installed
- Fly.io account created (`fly auth login`)

## First-Time Setup

### 1. Create the Fly app
```bash
cd scripts/becoming
fly apps create becoming-app
```

### 2. Create persistent volume for SQLite
```bash
fly volumes create becoming_data --region slc --size 1
```

### 3. Set secrets
```bash
# Required
fly secrets set BECOMING_DB=/data/becoming.db

# Optional: Google OAuth (omit to disable Google sign-in)
fly secrets set GOOGLE_CLIENT_ID=your-client-id
fly secrets set GOOGLE_CLIENT_SECRET=your-client-secret
fly secrets set GOOGLE_REDIRECT_URL=https://your-domain.fly.dev/auth/google/callback
```

### 4. Deploy
```bash
# From the repo root (so Dockerfile COPY paths resolve)
fly deploy --config scripts/becoming/fly.toml --dockerfile scripts/becoming/Dockerfile
```

### 5. Custom domain (optional)
```bash
fly certs add ibeco.me
fly certs add www.ibeco.me
```
Then configure DNS:
- `A` record → `<fly-app-ipv4>` (shown after `fly certs add`)
- `AAAA` record → `<fly-app-ipv6>`

## Subsequent Deploys
```bash
fly deploy --config scripts/becoming/fly.toml --dockerfile scripts/becoming/Dockerfile
```

## Useful Commands
```bash
fly status                    # App status
fly logs                      # Tail logs
fly ssh console               # SSH into the running machine
fly volumes list              # Check volume status
fly secrets list              # List configured secrets
fly scale count 1             # Ensure at least 1 machine running
```

## SQLite Backup
```bash
# SSH in and copy the database
fly ssh console
cp /data/becoming.db /data/becoming-backup.db
```

## Architecture Notes
- Single machine with attached volume (SQLite needs single-writer)
- Auto-stop/start to minimize costs when idle
- 256MB RAM, shared CPU — sufficient for personal use
- Salt Lake City region for lowest latency
