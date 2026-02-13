# Backup & Recovery

## Overview

The Becoming app supports two database backends:
- **SQLite**: Single-file database (local development)
- **PostgreSQL**: Full database server (production on VPS)

Each has different backup strategies.

---

## SQLite Backups (Local Dev)

### Manual Backup

```bash
# Copy the database file (while the app is stopped)
cp becoming.db becoming.db.bak

# OR use SQLite's built-in backup (safe while running — WAL mode)
sqlite3 becoming.db ".backup 'becoming-$(date +%Y%m%d).db'"
```

### Automated Backup (Windows Task Scheduler)

```powershell
# backup-sqlite.ps1 — Schedule with Windows Task Scheduler
$db = "C:\path\to\becoming\becoming.db"
$backupDir = "C:\path\to\backups"
$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"

# Use SQLite VACUUM INTO for a safe online backup
sqlite3 $db "VACUUM INTO '$backupDir\becoming-$timestamp.db'"

# Keep only the last 7 backups
Get-ChildItem "$backupDir\becoming-*.db" |
    Sort-Object -Descending |
    Select-Object -Skip 7 |
    Remove-Item -Force
```

### Export to JSON

The Becoming app has a built-in data export feature:
- **Settings > Export Data** — Downloads all your data as JSON
- **API**: `GET /api/export` — Returns full user data export

---

## PostgreSQL Backups (Production)

### Manual Backup

```bash
# Full database dump
pg_dump -h localhost -U becoming -d becoming -F custom -f becoming-$(date +%Y%m%d).dump

# SQL text dump (human-readable)
pg_dump -h localhost -U becoming -d becoming > becoming-$(date +%Y%m%d).sql
```

### Restore from Backup

```bash
# From custom format
pg_restore -h localhost -U becoming -d becoming --clean becoming-20260101.dump

# From SQL dump
psql -h localhost -U becoming -d becoming < becoming-20260101.sql
```

### Automated Daily Backup (cron)

```bash
#!/bin/bash
# /opt/becoming/backup.sh — Add to crontab: 0 2 * * * /opt/becoming/backup.sh

BACKUP_DIR="/opt/becoming/backups"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
RETENTION_DAYS=30

mkdir -p "$BACKUP_DIR"

# Dump database
PGPASSWORD="${POSTGRES_PASSWORD}" pg_dump \
    -h localhost -U becoming -d becoming \
    -F custom -f "$BACKUP_DIR/becoming-$TIMESTAMP.dump"

# Compress
gzip "$BACKUP_DIR/becoming-$TIMESTAMP.dump"

# Clean up old backups
find "$BACKUP_DIR" -name "becoming-*.dump.gz" -mtime +$RETENTION_DAYS -delete

echo "Backup complete: becoming-$TIMESTAMP.dump.gz"
```

### Dokploy Container Backup

When running with `docker-compose.yml`, the PostgreSQL data is in a Docker volume:

```bash
# Find the volume
docker volume ls | grep pgdata

# Backup via pg_dump inside the container
docker compose exec db pg_dump -U becoming -d becoming -F custom > becoming-backup.dump

# Restore
docker compose exec -T db pg_restore -U becoming -d becoming --clean < becoming-backup.dump
```

---

## Backup Strategy Recommendations

### Home Server (Proxmox)

| What | Frequency | Where | Retention |
|------|-----------|-------|-----------|
| pg_dump | Daily 2am | `/opt/becoming/backups/` | 30 days |
| Proxmox VM snapshot | Weekly | Proxmox storage | 4 weeks |
| Off-site copy | Weekly | Cloud storage (rclone) | 12 weeks |
| Data export (JSON) | Monthly | Download locally | Keep all |

### VPS (dedione.com)

| What | Frequency | Where | Retention |
|------|-----------|-------|-----------|
| pg_dump | Daily 2am | `/opt/becoming/backups/` | 30 days |
| Compressed dump to S3/B2 | Daily | Cloud object storage | 90 days |
| Data export (JSON) | Monthly | Download locally | Keep all |

### Setting up rclone for off-site backups

```bash
# Install rclone
curl https://rclone.org/install.sh | sudo bash

# Configure a remote (S3, Backblaze B2, Google Drive, etc.)
rclone config

# Sync backups to remote
rclone sync /opt/becoming/backups/ remote:becoming-backups/
```

---

## Cloudflare R2 (S3) Backups — Dokploy

Dokploy has built-in S3 backup support, but for database-specific backups you need a
script that dumps PostgreSQL before uploading. Two options:

### Option A: Dokploy Built-in S3 Backups

Dokploy can back up Docker volumes directly to S3-compatible storage:

1. **Settings → S3 Destinations** in Dokploy UI
2. Add a new destination:
   - **Endpoint**: `https://<account-id>.r2.cloudflarestorage.com`
   - **Bucket**: `becoming-backups`
   - **Region**: `auto`
   - **Access Key / Secret Key**: From Cloudflare R2 API tokens
3. Go to your project → **Backups** tab
4. Configure backup schedule for the `pgdata` volume
5. Dokploy will snapshot the volume and upload to R2

> **Pros**: Zero config, managed by Dokploy
> **Cons**: Volume-level backup (not a pg_dump), harder to restore selectively

### Option B: pg_dump + S3 Upload Script (Recommended)

A cron-based script that does a proper `pg_dump` and uploads to R2 via the S3 API.

```bash
#!/bin/bash
# /opt/becoming/backup-s3.sh
# Runs inside the Dokploy host or as a Docker sidecar
# Cron: 0 2 * * * /opt/becoming/backup-s3.sh

set -euo pipefail

# --- Configuration (set in Dokploy environment or .env) ---
: "${POSTGRES_PASSWORD:?Set POSTGRES_PASSWORD}"
: "${BACKUP_S3_ENDPOINT:?Set BACKUP_S3_ENDPOINT (e.g., https://acct.r2.cloudflarestorage.com)}"
: "${BACKUP_S3_BUCKET:?Set BACKUP_S3_BUCKET}"
: "${BACKUP_S3_ACCESS_KEY_ID:?Set BACKUP_S3_ACCESS_KEY_ID}"
: "${BACKUP_S3_SECRET_ACCESS_KEY:?Set BACKUP_S3_SECRET_ACCESS_KEY}"

TIMESTAMP=$(date +%Y%m%d-%H%M%S)
FILENAME="becoming-${TIMESTAMP}.sql.gz"
LOCAL_DIR="/tmp/becoming-backups"
RETENTION_DAYS=90

mkdir -p "$LOCAL_DIR"

echo "[$(date)] Starting backup..."

# 1. Dump PostgreSQL (from the db container)
docker compose -f /opt/becoming/docker-compose.yml exec -T db \
    pg_dump -U becoming -d becoming | gzip > "$LOCAL_DIR/$FILENAME"

SIZE=$(du -h "$LOCAL_DIR/$FILENAME" | cut -f1)
echo "[$(date)] Dump complete: $FILENAME ($SIZE)"

# 2. Upload to Cloudflare R2 via aws CLI (S3-compatible)
AWS_ACCESS_KEY_ID="$BACKUP_S3_ACCESS_KEY_ID" \
AWS_SECRET_ACCESS_KEY="$BACKUP_S3_SECRET_ACCESS_KEY" \
aws s3 cp "$LOCAL_DIR/$FILENAME" \
    "s3://$BACKUP_S3_BUCKET/becoming/$FILENAME" \
    --endpoint-url "$BACKUP_S3_ENDPOINT"

echo "[$(date)] Uploaded to R2: $BACKUP_S3_BUCKET/becoming/$FILENAME"

# 3. Clean up local temp file
rm -f "$LOCAL_DIR/$FILENAME"

# 4. Clean up old R2 backups (keep last RETENTION_DAYS)
# R2 lifecycle rules handle this automatically — configure in Cloudflare dashboard:
#   Object Lifecycle → Delete objects older than 90 days
# Or use: aws s3 ls + delete with --endpoint-url

echo "[$(date)] Backup complete!"
```

### Setup Steps for Cloudflare R2:

1. **Create R2 bucket** in Cloudflare dashboard:
   - Name: `becoming-backups`
   - Location hint: closest region to your VPS

2. **Create API token** (R2 → Manage R2 API Tokens):
   - Permissions: **Object Read & Write**
   - Scope: Just the `becoming-backups` bucket
   - Copy access key ID + secret access key

3. **Configure lifecycle rule** (R2 → Bucket → Settings → Object Lifecycle):
   - Delete objects older than 90 days (automatic cleanup)

4. **Install aws CLI on VPS**:
   ```bash
   apk add aws-cli  # Alpine
   # or
   apt install awscli  # Debian/Ubuntu
   ```

5. **Add cron job on VPS**:
   ```bash
   # Edit crontab
   crontab -e

   # Add daily backup at 2am
   0 2 * * * /opt/becoming/backup-s3.sh >> /var/log/becoming-backup.log 2>&1
   ```

6. **Set environment variables** (in `/opt/becoming/.env` or Dokploy):
   ```
   BACKUP_S3_ENDPOINT=https://<account-id>.r2.cloudflarestorage.com
   BACKUP_S3_BUCKET=becoming-backups
   BACKUP_S3_ACCESS_KEY_ID=<your-access-key>
   BACKUP_S3_SECRET_ACCESS_KEY=<your-secret-key>
   ```

### Restore from R2 Backup

```bash
# Download from R2
aws s3 cp "s3://becoming-backups/becoming/becoming-20260212-020000.sql.gz" /tmp/ \
    --endpoint-url "$BACKUP_S3_ENDPOINT"

# Restore into PostgreSQL container
gunzip -c /tmp/becoming-20260212-020000.sql.gz | \
    docker compose exec -T db psql -U becoming -d becoming
```

---

## Disaster Recovery

### Scenario: Database corruption

1. Stop the app: `docker compose stop app`
2. Restore from latest backup: `docker compose exec -T db pg_restore ...`
3. Start the app: `docker compose start app`

### Scenario: VPS failure

1. Provision new VPS (or use Proxmox home server as fallback)
2. Install Dokploy: `curl -sSL https://dokploy.com/install.sh | sh`
3. Deploy the Becoming app (point to repo)
4. Set environment variables (especially `POSTGRES_PASSWORD`)
5. Restore from latest off-site backup

### Scenario: Migration from SQLite to PostgreSQL

1. Export data from the app: `GET /api/export` (downloads JSON)
2. Set up PostgreSQL (via docker-compose or standalone)
3. Point the app to PostgreSQL: `BECOMING_DB=postgres://...`
4. Run the app (goose migrations will create the schema)
5. Import data (or re-register and re-enter — the export is a safety net)

> **Note**: There is no automated SQLite-to-PostgreSQL data migration tool yet.
> For a future version, a `becoming migrate --from sqlite --to postgres` command
> could be added to the CLI. For now, the JSON export/import is the migration path.

---

## Monitoring

### Check backup freshness

```bash
# Ensure latest backup is < 25 hours old
LATEST=$(ls -t /opt/becoming/backups/ | head -1)
AGE=$(( ($(date +%s) - $(stat -c %Y "/opt/becoming/backups/$LATEST")) / 3600 ))
if [ $AGE -gt 25 ]; then
    echo "WARNING: Latest backup is ${AGE} hours old!"
    # Send alert (email, webhook, etc.)
fi
```

### Check database connectivity

```bash
# Health check endpoint
curl -sf http://localhost:8080/api/auth/providers || echo "App is DOWN"

# PostgreSQL direct check
docker compose exec db pg_isready -U becoming
```
