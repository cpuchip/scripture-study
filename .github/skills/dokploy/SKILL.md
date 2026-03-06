---
name: dokploy
description: "Check deployment status, list recent deployments, view app health, and trigger deploys on Dokploy. Use when the user asks about deployments, server status, build logs, or wants to deploy/redeploy an application."
user-invokable: true
argument-hint: "[status | deploy | logs | version]"
---

# Dokploy Deployment Management

Self-hosted Dokploy instance at `https://dokploy.hmslogs.com`. Our own API calls — no third-party MCP needed.

## Configuration

The API key is stored as a **Windows user environment variable**: `DOKPLOY_API_KEY`.

The Dokploy panel URL is `https://dokploy.hmslogs.com`.

## Authentication

Load the key from the Windows user environment, then use the `x-api-key` header:
```powershell
# Load from Windows user env (survives terminal restarts)
$env:DOKPLOY_API_KEY = [System.Environment]::GetEnvironmentVariable("DOKPLOY_API_KEY", "User")
# Make API calls
curl -sk -H "x-api-key: $env:DOKPLOY_API_KEY" "https://dokploy.hmslogs.com/api/<endpoint>"
```

If `$env:DOKPLOY_API_KEY` is already set in the current session, skip the registry read.

## Known Application IDs

| App | Application ID | Domain | Project |
|-----|---------------|--------|---------|
| ibecome (ibeco.me) | `cKp5zaaaQlgBatKIiKN1K` | ibeco.me | NNKeReM683lglA6q0wtdp |
| tinyfarm.store | `72Q1ZEjpJ-cFxOyIIbWKe` | tinyfarm.store | U0ULu49KR5jYUeLJ9wwcO |
| hmslogs.com | `ic5OHUo51DVmQGCo4EayY` | hmslogs.com | W37Si3Bg-3OSdw6pmQEhY |

## API Endpoints

### List all projects
```
GET /api/project.all
```
Returns all projects with their applications, databases, and environments.

**WARNING:** This response includes environment variables (database passwords, OAuth secrets, etc.). Do NOT display raw output to the user. Extract only the fields you need (names, statuses, IDs).

### Get application details
```
GET /api/application.one?applicationId=<id>
```
Key fields: `applicationStatus` (done/running/error/idle), `branch`, `buildType`, `autoDeploy`.

### List deployments
```
GET /api/deployment.all?applicationId=<id>
```
Returns deployments newest-first. Key fields per deployment:
- `deploymentId` — unique ID
- `status` — `done`, `error`, `running`, `queued`
- `title` — commit message
- `description` — "Commit: <hash>"
- `createdAt`, `startedAt`, `finishedAt`
- `errorMessage` — populated on failure
- `logPath` — server-side log file path

### Trigger a deploy
```
POST /api/application.deploy
Content-Type: application/json
Body: {"applicationId": "<id>"}
```
**Ask the user before triggering this.** It kicks off a full Docker rebuild.

### Redeploy (rebuild from same commit)
```
POST /api/application.redeploy
Content-Type: application/json
Body: {"applicationId": "<id>"}
```
Same as deploy but rebuilds without pulling new code.

## Common Workflows

### Quick status check
1. Read `.env` for credentials
2. `GET /api/deployment.all?applicationId=cKp5zaaaQlgBatKIiKN1K` — check latest deployment status
3. `curl -sk https://ibeco.me/version` — verify running version matches expected commit

### After a git push
1. Wait a moment for Dokploy webhook to trigger auto-deploy
2. Check deployment status with the deployment list endpoint
3. Confirm the commit hash in `/version` matches the pushed commit

### Investigating a failed deploy
1. Get deployment list, find the `error` entry
2. Check `errorMessage` field
3. The `logPath` field shows where logs are on the server (not directly accessible via API in current version)
4. Look at the Dockerfile and recent changes for clues

## Security Notes

- **Never display raw `project.all` output** — it contains database passwords, OAuth secrets, and session keys for ALL projects.
- **Load key from Windows env** — `[System.Environment]::GetEnvironmentVariable("DOKPLOY_API_KEY", "User")`. Never hardcode it.
- **Ask before deploying** — `application.deploy` and `application.redeploy` are destructive operations.
- The API key has full admin access. Treat it like a root password.
