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
