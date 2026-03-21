# WS-D: Nocix Server Setup + Dokploy Migration

**Status:** In progress — paused at Step 1 (SSH DDoS protection triggered)  
**Server:** `customer@server.ibeco.me`  
**Goal:** Move production web hosting from home Proxmox/Dokploy to Nocix dedicated server in Kansas City.

---

## What We Know

- **OS:** Ubuntu 24.04.4 LTS
- **CPU:** AMD Ryzen 7 3800XT (16 threads)
- **RAM:** 32GB
- **SSD:** `/dev/sda` — 447GB, partitioned as: 1M (BIOS boot), 1G (`/boot`), 16G (swap), 430G (`/`)
- **HDD:** `/dev/sdb` — 2.7TB, **no partitions, not mounted**
- **Users:** `root`, `customer` (uid 1000, sudo group)
- **SSH key already in place:** `customer@precisionChip` key is in `~/.ssh/authorized_keys`
- **Sudoers:** `customer` and `cpuchip` added to NOPASSWD sudoers list ✓

### SSH DDoS Protection Issue
Nocix rate-limits rapid sequential SSH connections and blocks port 22 temporarily. Same behavior seen on the DediOne server earlier. The fix is to keep a **persistent SSH session** rather than opening/closing connections for each command.

**Solution options:**
1. **VS Code Remote - SSH** — opens a persistent SSH tunnel, all terminal work stays in one session. Best for this workflow.
2. **SSH ControlMaster** — add to `~/.ssh/config` to multiplex connections over one socket:
   ```
   Host server.ibeco.me
     ControlMaster auto
     ControlPath ~/.ssh/cm-%r@%h:%p
     ControlPersist 10m
   ```
3. **tmux on server** — connect once, attach/detach without closing the connection.

**Recommended:** Set up VS Code Remote SSH first, then use the integrated terminal for all subsequent steps. This keeps one connection alive for the whole session.

---

## Step 1 — Machine Prep ✅ assessed, ⏳ not executed

### 1a. Setup 2.7TB HDD (`/dev/sdb`)

```bash
# Create GPT partition table and single partition
sudo parted /dev/sdb --script mklabel gpt
sudo parted /dev/sdb --script mkpart primary ext4 0% 100%

# Format as ext4
sudo mkfs.ext4 /dev/sdb1

# Create mount point
sudo mkdir -p /mnt/data

# Get UUID for fstab
sudo blkid /dev/sdb1

# Add to fstab (replace UUID with actual value)
echo "UUID=<uuid-here> /mnt/data ext4 defaults,nofail 0 2" | sudo tee -a /etc/fstab

# Mount
sudo mount -a

# Verify
df -h /mnt/data
```

### 1b. Create `cpuchip` user

```bash
# Create user with home directory
sudo useradd -m -s /bin/bash -G sudo cpuchip

# Copy SSH key from customer
sudo mkdir -p /home/cpuchip/.ssh
sudo cp /home/customer/.ssh/authorized_keys /home/cpuchip/.ssh/
sudo chown -R cpuchip:cpuchip /home/cpuchip/.ssh
sudo chmod 700 /home/cpuchip/.ssh
sudo chmod 600 /home/cpuchip/.ssh/authorized_keys

# Set password (Michael does this himself)
# sudo passwd cpuchip

# Verify login works before continuing
```

### 1c. System prep

```bash
sudo apt update && sudo apt upgrade -y
sudo apt install -y curl git htop ufw
```

### 1d. Firewall baseline (UFW)

```bash
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 3000/tcp   # Dokploy UI — can restrict later
sudo ufw enable
```

---

## Step 2 — Install Dokploy

Dokploy's installer handles Docker installation as well.

```bash
curl -sSL https://dokploy.com/install.sh | sh
```

After install:
- Dokploy UI is available at `http://server.ibeco.me:3000`
- Complete the web setup: create admin account, set domain
- Set Dokploy's data/volumes root to `/mnt/data/dokploy` (use the Settings → Storage path option in the UI)

**Important:** Before installing, create the data directory structure on the HDD:

```bash
sudo mkdir -p /mnt/data/dokploy
sudo mkdir -p /mnt/data/minio
sudo mkdir -p /mnt/data/postgres
```

---

## Step 3 — Migration from Home Dokploy

Dokploy does not have a built-in "export project" migration. The process is per-service:

### 3a. For each service on the home server:

1. **Note the config** — in Dokploy UI: environment variables, exposed ports, domain settings, volume mounts
2. **Export data** (databases, file volumes):
   - For Postgres: `pg_dump` to a file, copy via `scp` or S3
   - For file volumes: `tar` the volume directory, transfer
3. **Recreate in new Dokploy** — add project, add service, paste env vars, set domain
4. **Import data** — restore DB dumps, untar volumes
5. **Test** — verify service works on new server before cutting DNS

### 3b. GitOps / GitHub setup

In Dokploy on new server:
- Settings → GitHub → connect GitHub app (or use deploy key per repo)
- Each service: set Git repository, branch, build command
- Set up automatic deploys on push

### 3c. DNS domains

For each domain:
- Add domain to service in Dokploy
- Dokploy handles Traefik + Let's Encrypt automatically
- Don't point DNS yet — test via `/etc/hosts` trick first:
  ```
  <nocix-server-ip>  yourdomain.com
  ```

---

## Step 4 — MinIO S3 Storage on the HDD

Deploy MinIO as a Dokploy service (Docker Compose):

```yaml
services:
  minio:
    image: minio/minio
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: admin
      MINIO_ROOT_PASSWORD: <set-strong-password>
    volumes:
      - /mnt/data/minio:/data
    ports:
      - "9000:9000"   # S3 API
      - "9001:9001"   # Console
    restart: unless-stopped
```

- Point a subdomain at port 9001 for the console (e.g., `s3-admin.ibeco.me`)
- Point another at port 9000 for the S3 API (e.g., `s3.ibeco.me`)

For cross-site backup replication (home ↔ Nocix), use MinIO's built-in bucket replication:
```bash
mc alias set nocix http://s3.ibeco.me admin <password>
mc alias set home http://s3.home.ibeco.me admin <password>
mc replicate add nocix/mybucket --remote-bucket http://admin:<password>@home-minio:9000/mybucket
```

---

## Step 5 — DNS Cutover via Squarespace

Once tested on the new server:

1. Log into Squarespace → Domains → DNS settings
2. Update A records to point to Nocix server IP
3. TTL: set to 300 (5 min) before the cutover for fast rollback
4. After cutover is confirmed stable, set TTL back to 3600

---

## Open Questions / Next Steps

- [ ] Set up VS Code Remote SSH to `customer@server.ibeco.me` to avoid DDoS protection triggering
- [ ] Set `ControlMaster` in local `~/.ssh/config` as backup
- [ ] Confirm which services are running on home Dokploy (need inventory)
- [ ] Decide: migrate everything at once, or one service at a time?
- [ ] What domains are currently on Squarespace pointing to home server?
- [ ] Does the home Dokploy server currently hold any persistent data (DBs, uploads)?

---

## Notes

- Nocix has known power outage history (July 2023 major event). Home server as backup is the right call.
- The 2.7TB HDD is plenty for MinIO + Docker volumes + backups for the foreseeable future.
- Nocix TOS prohibits open proxies and CDN installations without prior approval — Traefik (Dokploy's built-in reverse proxy) is fine, it's not a CDN.
- Latency: 65ms from Marshfield to KCIX — fine for web hosting, usable for SSH.
