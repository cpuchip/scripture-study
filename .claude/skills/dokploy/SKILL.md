---
name: dokploy
description: Check deployment status, list recent deployments, view app health, and trigger deploys on Dokploy. Use when the user asks about deployments, server status, build logs, or wants to deploy/redeploy an application.
argument-hint: "[status | deploy | logs | version]"
---

# Dokploy Deployment Management

Michael runs **two** self-hosted Dokploy instances. Pick the right one
based on which app/domain is being asked about — this is the #1 source
of wasted turns ("why is the API returning 502?" usually means you hit
the wrong instance).

## Instances

| Instance | URL | Env var holding API key | Hosts |
|---|---|---|---|
| **NOCIX VPS** (cloud) | `https://server.ibeco.me` | `DOKPLOY_NOCIX_API_KEY` | ibeco.me + sub-domains, cpuchip.net, marsfield.org, 1828 (Compose) |
| **Home NAS** | `https://dokploy.hmslogs.com` | `DOKPLOY_API_KEY` | hmslogs.com + home-network apps (tinyfarm.store etc.) |

The home NAS instance can be flaky / 502 when the home network is
restarting or the NAS is being worked on. The NOCIX VPS is the
always-on production instance.

## Configuration

API keys live in Windows user environment variables (survive terminal
restarts; safer than `.env` files).

```powershell
# NOCIX VPS (server.ibeco.me) — the primary instance
$env:DOKPLOY_NOCIX_API_KEY = [System.Environment]::GetEnvironmentVariable("DOKPLOY_NOCIX_API_KEY", "User")

# Home NAS (dokploy.hmslogs.com)
$env:DOKPLOY_API_KEY       = [System.Environment]::GetEnvironmentVariable("DOKPLOY_API_KEY", "User")
```

If the env var is already set in the current session, skip the
registry read.

> **Note 2026-05-23:** the NOCIX key was previously named
> `DOKPLOY_NOICX_API_KEY` (with the typo `NOICX`). Michael renamed
> it; `DOKPLOY_NOCIX_API_KEY` is now the canonical name. The old
> `NOICX` variable is still set during the transition but should not
> be relied on going forward.

## Authentication

Use the `x-api-key` header against the matching instance:

```powershell
# NOCIX
curl -sk -H "x-api-key: $env:DOKPLOY_NOCIX_API_KEY" "https://server.ibeco.me/api/<endpoint>"

# Home NAS
curl -sk -H "x-api-key: $env:DOKPLOY_API_KEY"       "https://dokploy.hmslogs.com/api/<endpoint>"
```

Hitting the wrong URL with the wrong key returns `401 Unauthorized`.
Hitting the right URL when the instance is down returns `502 Bad
Gateway` (panel + API both); that's an instance-down state, not an
auth problem.

## Known IDs

### NOCIX VPS (server.ibeco.me) — verified 2026-05-22

| Project | Project ID | Service | Service ID | Type | Notes |
|---|---|---|---|---|---|
| `Marsfield.org` | `WZh-QDpkYTTXqb-PdQKEJ` | `web` | `Cvh1jmkE-_TRZC15SCXjk` | Application | marsfield.org |
| `cpuchip.net` | `-ww7musqUc3dspMWdZD85` | `web` | `cNhR0ymKdtVLHIMNuckxS` | Application | cpuchip.net |
| `ibeco.me` | `V5WLEhO8bZxHpVqL7PHTL` | `engine` | `Qo8QSTWShNeqGsPrtw90T` | Application | engine.ibeco.me |
| `ibeco.me` | `V5WLEhO8bZxHpVqL7PHTL` | `web` | `Uu_-qX0ZPdotJ0mQGwn-j` | Application | ibeco.me (becoming-app) |
| `ibeco.me` | `V5WLEhO8bZxHpVqL7PHTL` | `i1828` | `5pWsGxF5yMtOJUACHJyMB` | **Compose** | 1828.ibeco.me |
| `ai-chattermax` | `nGX4hGUUzV0vAJPOYBk4f` | `chattermax` | `B-ER3-TrXHlDr4iCHDOvj` | **Compose** | chat.ibeco.me:8080 (github cpuchip/ai-chattermax, ./docker-compose.yml, domainId `CLLbog6XrdUt8ia6T5LBl`) — created 2026-06-04 |
| `deadweight` | `ilaeCtLXDQrQsP9mlK9rX` | `game` | `l4tkfFkX5GvAvNSlmqR3H` | **Compose** | deadweight.cpuchip.net:8080 (github cpuchip/deadweight-acquisitions-game @ master, ./docker-compose.yml, auto-deploy on push) — created 2026-06-17 |

Project `ibeco.me` also carries 2 project-level Postgres services
(IDs `4xQXSslPNlZl6N1fxv7Zq` + `m6Cfc7WPhAl1eV9Qv25Tf`). The `i1828`
compose runs its OWN internal Postgres inside its compose file —
separate from those project-level DBs.

**Build sources (which repo each app auto-deploys from, `triggerType:"push"`):**

| Service | Repo it builds from | Build path |
|---|---|---|
| `web` (becoming/ibeco.me) | **workspace-root monorepo** `github.com/cpuchip/scripture-study` (main) | `scripts/becoming/` |
| `engine` (engine.ibeco.me) | its own repo `github.com/cpuchip/gospel-engine` | repo root |
| `chattermax` (chat.ibeco.me) | `github.com/cpuchip/ai-chattermax` | `./docker-compose.yml` |
| `cpuchip.net` / `Marsfield.org` | their own repos | repo root |

⚠ **A workspace-ROOT push = an ibeco.me PRODUCTION rebuild** (the `web` app
deploys from the monorepo). "Commit to root, don't push" (Michael pushes) is a
deploy boundary, not just preference. gospel-engine's URL is authoritative in
`.mcp.json` → `GOSPEL_ENGINE_URL=https://engine.ibeco.me` (NOT study.ibeco.me).

### Home NAS (dokploy.hmslogs.com) — stale; verify before trusting

The table below predates the NOCIX migration and is unverified against
the current home-NAS DB. **Re-fetch with `project.all` against
`dokploy.hmslogs.com` when the instance is up, then update.**

| App | Application ID | Domain | Project |
|-----|---------------|--------|---------|
| ibecome (legacy) | `cKp5zaaaQlgBatKIiKN1K` | (was ibeco.me) | `NNKeReM683lglA6q0wtdp` |
| tinyfarm.store | `72Q1ZEjpJ-cFxOyIIbWKe` | tinyfarm.store | `U0ULu49KR5jYUeLJ9wwcO` |
| hmslogs.com | `ic5OHUo51DVmQGCo4EayY` | hmslogs.com | `W37Si3Bg-3OSdw6pmQEhY` |

## API Endpoints

### List all projects
```
GET /api/project.all
```
Returns all projects with their applications, databases, and
environments. Service arrays (`applications`, `compose`, `postgres`,
etc.) live INSIDE `environments[]`, not at the project root — easy
to miss.

**WARNING:** This response includes environment variables (database
passwords, OAuth secrets, etc.). Do NOT display raw output to the
user. Extract only the fields you need (names, statuses, IDs).

### Get application details
```
GET /api/application.one?applicationId=<id>
```
Key fields: `applicationStatus` (done/running/error/idle), `branch`,
`buildType`, `autoDeploy`.

### Get compose details
```
GET /api/compose.one?composeId=<id>
```
Same shape but for Compose services. `composeStatus` instead of
`applicationStatus`. For the 1828 deploy this is the right endpoint.

### List deployments
```
GET /api/deployment.all?applicationId=<id>           # Application type
GET /api/deployment.allByCompose?composeId=<id>      # Compose type — NOT deployment.all (that 400s wanting applicationId)
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
POST /api/application.deploy   Body: {"applicationId": "<id>"}
POST /api/compose.deploy       Body: {"composeId": "<id>"}
```
**Ask the user before triggering this.** It kicks off a full rebuild.

### Redeploy (rebuild from same commit)
```
POST /api/application.redeploy   Body: {"applicationId": "<id>"}
POST /api/compose.redeploy       Body: {"composeId": "<id>"}
```
Same as deploy but rebuilds without pulling new code.

## Common Workflows

### Quick status check (1828.ibeco.me)
1. `GET /api/deployment.allByCompose?composeId=5pWsGxF5yMtOJUACHJyMB` — latest deployments
2. `curl -sk https://1828.ibeco.me/api/healthz` — live health probe

### Quick status check (chat.ibeco.me / ai-chattermax)
1. `GET /api/deployment.allByCompose?composeId=B-ER3-TrXHlDr4iCHDOvj` — latest deployments
2. `curl -s https://chat.ibeco.me/healthz` — live health probe (`{"status":"ok"}`)

### Quick status check (ibeco.me / becoming)
1. `GET /api/deployment.all?applicationId=Uu_-qX0ZPdotJ0mQGwn-j` — latest deployments
2. `curl -sk https://ibeco.me/version` — verify running version matches expected commit

### After a git push
1. Wait a moment for Dokploy webhook to trigger auto-deploy
2. Check deployment status with the deployment list endpoint
3. Confirm the commit hash in `/version` (or the equivalent for the app) matches the pushed commit

### Investigating a failed deploy
1. **`status=error` does NOT mean the site is down.** A failed rebuild leaves the
   prior good container serving — confirm with a live curl (`curl -s https://<domain>/...`)
   before treating it as an outage. The `error` is the latest *build attempt*, not the
   running app. (2026-06-05: ibeco.me showed `error` while still serving + auth working.)
2. Get the deployment list, find the `error` entry. **`errorMessage` is usually EMPTY**
   even on a real failure — don't conclude "transient" from an empty error field.
3. Read the actual build log — that's where the cause is. Dokploy UI → the app →
   Deployments → the failed one → Logs. (`logPath` points at the server-side file but
   isn't API-readable; have Michael paste it if you can't reach it.)
4. Look at the Dockerfile + the triggering commit for clues.

**Known build gotcha — commit subject in ldflags (fixed `2b98b4c`, 2026-06-05).**
The becoming Dockerfile baked the commit SUBJECT into the binary via
`-ldflags "… -X 'main.ReleaseNotes=$MSG'"`. A `'` or `"` in the subject (e.g.
"Michael's expansion") closes the single-quote grouping Go's `quoted.Split` relies
on → the linker gets stray flags and aborts with `usage: link [options] main.o`
(exit 1, fails ~13s in). Because `web` builds from the root monorepo, ANY root commit
message with a quote could trip it. Fixed by sanitizing (`tr -d '\047\042'`). **If a Go
build fails with a linker usage dump, suspect the ldflags value, not the Go code.**

### Refreshing the Known IDs table after a UI-side change
The table above can drift. When in doubt, re-fetch:
```powershell
curl -sk -H "x-api-key: $env:DOKPLOY_NOCIX_API_KEY" "https://server.ibeco.me/api/project.all" > $env:TEMP\proj.json
# inspect with python/jq — pull projectId, applicationId, composeId, name, domain.host
```
Then update this skill if anything changed.

### Creating a new Compose app from scratch (verified 2026-06-04, ai-chattermax)
Each step's required fields were found by POSTing `{}` and reading the Zod error.
```
1. POST /api/project.create   {name, description?}
   -> creates the project + an auto "production" environment.
   The response is lean; re-fetch project.all and read
   environments[].environmentId for the next step.
2. POST /api/compose.create   {name, environmentId, composeType:"docker-compose"}
   -> composeId. appName is auto-generated.
3. POST /api/compose.update   {composeId, sourceType:"github", githubId, owner,
                                repository, branch, composePath:"./docker-compose.yml",
                                triggerType:"push"}
   -> wires the git source + auto-deploy-on-push. Reuse an existing compose's
   githubId (the GitHub App install covers all repos under that owner).
4. POST /api/domain.create    {host, composeId, serviceName:"<compose service name>",
                                domainType:"compose", port:<container port>,
                                https:true, certificateType:"letsencrypt", path:"/"}
   -> chat.ibeco.me etc. (domain.all is 404; list per-parent via
   domain.byComposeId?composeId=). *.ibeco.me wildcard DNS already resolves
   sub-domains to the NOCIX VPS, so only the Dokploy-side domain is needed.
5. POST /api/compose.deploy   {composeId}
   -> builds + goes live. Poll compose.one.composeStatus (idle->done) and curl
   the live /healthz; first deploy issues the Let's Encrypt cert (a few seconds).
```

## Security Notes

- **Never display raw `project.all` output** — it contains database
  passwords, OAuth secrets, and session keys for ALL projects. Extract
  only the fields you need (names, statuses, IDs).
- **Load keys from Windows env** — never hardcode. `NOCIX_API_KEY` and
  `API_KEY` are different keys with different scopes; don't mix them.
- **Ask before deploying** — `application.deploy` / `application.redeploy`
  / `compose.deploy` / `compose.redeploy` are destructive operations.
- Both API keys have full admin access on their instance. Treat them
  like root passwords.
