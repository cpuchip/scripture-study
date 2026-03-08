# Phase 3: Authentication, API Tokens & Multi-User

## Goal

Transform Becoming from a single-user local app into a multi-user hosted application with:
1. **Email/password authentication** for browser sessions
2. **API tokens** for programmatic access (AI assistants, scripts, MCP servers)
3. **Data isolation** — every user's data is fully scoped
4. **MCP server** — enabling AI study assistants to interact with your Becoming data

Preserve the current single-binary architecture and local development experience throughout.

## Current State

- Single Go binary with embedded Vue SPA
- SQLite database (WAL mode, foreign keys)
- No authentication — anyone with the URL has full access
- All data is implicitly "user 1"
- Runs locally on `localhost:8080`
- `-dev` flag enables CORS for Vite dev server

---

## Decisions

### Decision 1: Identity Provider Strategy

| # | Option | Pros | Cons |
|---|--------|------|------|
| 1 | Google OAuth only | Simple, no passwords to manage. | Excludes users without Google. No API token story. |
| 2 | **Email + Password first** | Universal. No external dependencies. Works offline. | Password management (hashing, reset flow). |
| 3 | Magic link (email only) | No passwords. | Need email service. Slow login (check email every time). |
| 4 | OAuth first, email later | Covers most users quickly. | Delays the universal option. |
| 5 | **Email/password first, OAuth later** | Start universal. Add convenience later. | Two sprints for full coverage. |

**Decision: Option 5 — Email/password first, Google OAuth added by user when ready**

Rationale:
- Email/password works for everyone, everywhere, immediately
- No external service dependency to start (no Google Cloud Console needed)
- The user will set up Google OAuth credentials on their own timeline and provide them via `.env`
- When `.env` contains Google creds, OAuth endpoints light up automatically
- The `users` table supports both `provider='email'` (with password_hash) and `provider='google'` (with provider_id) from day one

### Decision 2: Session & Token Strategy

This is the key architectural decision. **We need both cookies AND tokens**, for different purposes:

| Auth Method | Used By | Storage | Revocable | Stateless |
|-------------|---------|---------|-----------|-----------|
| **Session cookie** | Browser (SPA) | HttpOnly cookie | Yes (DB lookup) | No |
| **API token** | AI assistants, scripts, MCP | `Authorization: Bearer <token>` | Yes (DB lookup) | No |

#### Why cookies for the browser?

It's not that JWTs are insecure — you use them at work, and they're fine for cross-service auth in microservice architectures where multiple services need to verify identity without a shared session store. JWTs shine when:
- Multiple backends need to verify the same token
- You need stateless verification across services
- Token introspection is expensive

But for Becoming, we have a **single Go binary talking to a single SQLite database**. In that world:
- Statelessness has no advantage (the DB is right there, one query)
- HttpOnly cookies mean JavaScript can **never** read the session token — no XSS can exfiltrate it
- Cookies are sent automatically by the browser — no auth header management in the SPA
- Server-side sessions are instantly revocable (delete the row)

**The tradeoff:** One extra DB query per request (session lookup). At our scale, this is ~0.1ms on SQLite. Negligible.

#### Why API tokens for programmatic access?

Cookies don't work for scripts, AI assistants, or MCP servers — they need a token in a header. API tokens are:
- Generated from the user's profile page
- Stored as bcrypt hashes (like passwords — we never store the raw token)
- Sent as `Authorization: Bearer <token>`
- Scoped with a name and optional permissions
- Revocable from the profile page
- Trackable (`last_used` timestamp)

This is exactly what GitHub Personal Access Tokens, DigitalOcean API tokens, and similar systems do.

#### The middleware handles both:

```go
func AuthRequired(db *DB) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            var userID int64

            // 1. Check for session cookie (browser)
            if cookie, err := r.Cookie("becoming_session"); err == nil {
                if session, err := db.GetSession(cookie.Value); err == nil && !session.IsExpired() {
                    db.TouchSession(session.ID)
                    userID = session.UserID
                }
            }

            // 2. Check for Bearer token (API/MCP)
            if userID == 0 {
                if token := extractBearerToken(r); token != "" {
                    if apiToken, err := db.ValidateAPIToken(token); err == nil {
                        db.TouchAPIToken(apiToken.ID)
                        userID = apiToken.UserID
                    }
                }
            }

            // 3. Dev mode fallback
            if userID == 0 && devMode {
                userID = 1
            }

            if userID == 0 {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }

            ctx := context.WithValue(r.Context(), userIDKey, userID)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### Session Cookie Details

```
Name:     becoming_session
Value:    <random 32-byte hex token>
Path:     /
HttpOnly: true     ← JavaScript cannot read this cookie
Secure:   true     ← Only sent over HTTPS (except localhost)
SameSite: Lax      ← Sent on same-site requests + top-level navigations
MaxAge:   30 days  ← Sliding window (refreshed on activity)
```

### Decision 3: OAuth Flow (When Enabled)

Standard server-side OAuth 2.0 Authorization Code flow. This activates automatically when `BECOMING_GOOGLE_CLIENT_ID` and `BECOMING_GOOGLE_CLIENT_SECRET` are set in `.env`:

```
Browser                    Go Backend                 Google
  │                           │                          │
  ├─── GET /auth/google ─────►│                          │
  │                           ├── redirect ──────────────►│
  │    ◄── 302 to Google ─────┤                          │
  │                           │                          │
  │    (user signs in at Google)                         │
  │                           │                          │
  │    ◄── 302 callback ──────────────────────────────────┤
  ├─── GET /auth/callback ───►│                          │
  │                           ├── exchange code ─────────►│
  │                           │◄── id_token + profile ───┤
  │                           ├── find/create user       │
  │                           ├── create session         │
  │    ◄── Set-Cookie + 302 ──┤                          │
  ├─── (redirected to app) ──►│                          │
```

The login page checks `GET /api/auth/providers` to know which buttons to show. If Google creds aren't configured, the Google button simply doesn't appear.

### Decision 4: Hosting

| # | Option | Cost | Fit |
|---|--------|------|-----|
| 1 | **Dokploy (VPS)** | Self-hosted on VPS via Docker Compose + PostgreSQL | **Chosen** — auto-deploys on push to `main` via GitHub app. Custom domains, TLS via Traefik. |
| 2 | Railway | $5/mo hobby | Good, but costs from day one. |
| 3 | DigitalOcean droplet | $6/mo | Full control but more ops work. |
| 4 | Cloudflare Workers + D1 | Free tier generous | Interesting but requires refactoring to Workers runtime (no Go). |
| 5 | Self-hosted (home server) | $0 | Good for development, bad for reliability/HTTPS. |
| 6 | Vercel/Netlify (static) + separate API | Varies | Over-complicates the architecture. Our SPA is embedded in the Go binary. |

**Chosen: Dokploy on VPS**
- Docker Compose with Go binary + PostgreSQL
- Auto-deploys on push to `main` (Dokploy GitHub app watches the repo)
- TLS via Traefik reverse proxy
- Custom domains (ibeco.me)
- Full control, no vendor lock-in

### Decision 5: Domain Strategy

| Domain | Purpose | Phase |
|--------|---------|-------|
| **ibeco.me** | Personal app — the solo "becoming" experience | Phase 3 |
| **webeco.me** | Social/group features — the community "becoming" | Phase 6+ |

Both point to the same deployed app initially. Routing can differentiate later:
- `ibeco.me` → personal dashboard, practices, reflections
- `webeco.me` → group features, shared pillars, accountability

For Phase 3, both domains serve the same app. The distinction is branding/intent.

---

## Multi-User Audit: Existing Tables

Before building, we audited every table and query for multi-user readiness. Here's what we found:

### Tables That Need user_id

| Table | Current Constraints | Migration Notes |
|-------|-------------------|-----------------|
| `practices` | None blocking | Add `user_id`, index it. Straightforward. |
| `practice_logs` | FK to `practices(id)` | Add `user_id`. Indirectly scoped via practice_id, but `ListLogsByDate(date)` has no user filter — needs one. |
| `tasks` | None blocking | Add `user_id`, index it. `WHERE 1=1` pattern is easy to extend. |
| `notes` | FK to practices/tasks | Add `user_id`. JOIN queries in notes.go need `WHERE n.user_id = ?`. |
| `reflections` | **`UNIQUE(date)`** | **Must become `UNIQUE(user_id, date)`** — currently only one reflection per date globally. |
| `prompts` | None | Add `user_id`. `SeedPrompts` uses `COUNT(*)` globally — needs `WHERE user_id = ?`. |
| `pillars` | None | Add `user_id`. `HasPillars` uses `COUNT(*)` globally — needs `WHERE user_id = ?`. |

### Junction Tables (No user_id Needed)

| Table | Why Safe |
|-------|----------|
| `practice_pillars` | Scoped through `practice_id` → user owns the practice |
| `task_pillars` | Scoped through `task_id` → user owns the task |

These don't need `user_id` because the parent entities are already user-scoped. A user can only link pillars to their own practices/tasks.

### Query Patterns Requiring Changes

| File | Function | Concern |
|------|----------|---------|
| `reports.go` | `GetReport` | Cross-table JOIN: `practices LEFT JOIN practice_logs` — needs `WHERE p.user_id = ?` |
| `memorize.go` | `GetMemorizeQueue` | Complex subqueries on `practice_logs` — needs user_id scoping on both practices and logs |
| `schedule.go` | `GetSchedule` | `MAX(date)` queries on `practice_logs` by `practice_id` — indirectly scoped, but should add explicit user filter |
| `notes.go` | `ListNotes` | LEFT JOIN to practices + tasks for display names — needs `WHERE n.user_id = ?` |
| `reflections.go` | `SeedPrompts` | `COUNT(*)` is global — must become per-user |
| `pillars.go` | `HasPillars` | `COUNT(*)` is global — must become per-user |
| `logs.go` | `ListLogsByDate` | No user filter — queries all logs for a given date |

### Migration Strategy

All existing data gets `user_id = 1` (the DEFAULT). First user to register claims that data. This is safe because:
- SQLite `ALTER TABLE ADD COLUMN` with `DEFAULT 1` doesn't rewrite existing rows
- The `-dev` flag will auto-login as user_id=1, preserving the local dev experience
- New users start fresh (no existing data to conflict with)

---

## Database Changes

### New Tables

```sql
-- Users (identity)
CREATE TABLE IF NOT EXISTS users (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL DEFAULT '',       -- bcrypt hash (empty for OAuth-only users)
    name          TEXT NOT NULL DEFAULT '',
    avatar_url    TEXT NOT NULL DEFAULT '',
    provider      TEXT NOT NULL DEFAULT 'email',  -- 'email', 'google', 'apple'
    provider_id   TEXT NOT NULL DEFAULT '',        -- OAuth subject ID (empty for email users)
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_login    DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_provider
    ON users(provider, provider_id) WHERE provider != 'email';

-- Sessions (browser auth)
CREATE TABLE IF NOT EXISTS sessions (
    id          TEXT PRIMARY KEY,                 -- random 32-byte hex token
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at  DATETIME NOT NULL,
    last_active DATETIME DEFAULT CURRENT_TIMESTAMP,
    user_agent  TEXT NOT NULL DEFAULT '',
    ip_address  TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

-- API tokens (programmatic auth)
CREATE TABLE IF NOT EXISTS api_tokens (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL DEFAULT '',         -- "Copilot scripture study", "backup script"
    token_hash  TEXT NOT NULL,                    -- bcrypt hash of the token
    prefix      TEXT NOT NULL DEFAULT '',         -- first 8 chars for identification (e.g., "bec_a1b2")
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_used   DATETIME,
    expires_at  DATETIME                          -- NULL = never expires
);
CREATE INDEX IF NOT EXISTS idx_api_tokens_user ON api_tokens(user_id);

-- OAuth state (CSRF protection — only needed when Google OAuth is enabled)
CREATE TABLE IF NOT EXISTS oauth_states (
    state       TEXT PRIMARY KEY,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    redirect_to TEXT NOT NULL DEFAULT '/'
);
```

### Migration of Existing Tables

```sql
-- Add user_id to all existing tables (existing data becomes user_id=1)
ALTER TABLE practices ADD COLUMN user_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE practice_logs ADD COLUMN user_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE tasks ADD COLUMN user_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE notes ADD COLUMN user_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE reflections ADD COLUMN user_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE prompts ADD COLUMN user_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE pillars ADD COLUMN user_id INTEGER NOT NULL DEFAULT 1;

-- Indexes for every query path
CREATE INDEX IF NOT EXISTS idx_practices_user ON practices(user_id);
CREATE INDEX IF NOT EXISTS idx_practice_logs_user ON practice_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_tasks_user ON tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_notes_user ON notes(user_id);
CREATE INDEX IF NOT EXISTS idx_reflections_user ON reflections(user_id);
CREATE INDEX IF NOT EXISTS idx_pillars_user ON pillars(user_id);

-- Fix the reflections uniqueness constraint for multi-user
-- SQLite can't ALTER constraints, so we recreate the table:
-- (handled in Go migration code — create new table, copy data, drop old, rename)
-- New constraint: UNIQUE(user_id, date) instead of UNIQUE(date)
```

**Note:** `practice_pillars` and `task_pillars` do NOT get `user_id` — they're scoped through their parent entities.

---

## API Changes

### New Endpoints

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | `/auth/register` | Create account (email + password) | None |
| POST | `/auth/login` | Login (email + password), set session cookie | None |
| POST | `/auth/logout` | Destroy session, clear cookie | Cookie |
| GET | `/auth/google` | Initiate Google OAuth (when configured) | None |
| GET | `/auth/callback` | Handle OAuth callback | None |
| GET | `/api/auth/providers` | List enabled auth methods (`{email: true, google: false}`) | None |
| GET | `/api/me` | Get current user profile | Cookie/Token |
| PUT | `/api/me` | Update user profile (name) | Cookie/Token |
| DELETE | `/api/me` | Delete account and all data | Cookie |
| GET | `/api/tokens` | List API tokens (name, prefix, created, last_used) | Cookie |
| POST | `/api/tokens` | Create API token — returns the raw token ONCE | Cookie |
| DELETE | `/api/tokens/{id}` | Revoke an API token | Cookie |

**Token creation flow:**
1. User clicks "Create API Token" on profile page
2. Enters a name (e.g., "Copilot scripture study")
3. Server generates `bec_<32 random hex chars>`, stores bcrypt hash
4. Raw token is shown ONCE: "Copy this token now — you won't see it again"
5. Token appears in list as `bec_a1b2...` (prefix only) with name and last_used

### Middleware

```go
func AuthRequired(db *DB) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            var userID int64

            // 1. Session cookie (browser)
            if cookie, err := r.Cookie("becoming_session"); err == nil {
                if session, err := db.GetSession(cookie.Value); err == nil && !session.IsExpired() {
                    db.TouchSession(session.ID)
                    userID = session.UserID
                }
            }

            // 2. Bearer token (API/MCP)
            if userID == 0 {
                if token := extractBearerToken(r); token != "" {
                    if apiToken, err := db.ValidateAPIToken(token); err == nil {
                        db.TouchAPIToken(apiToken.ID)
                        userID = apiToken.UserID
                    }
                }
            }

            // 3. Dev mode fallback
            if userID == 0 && devMode {
                userID = 1
            }

            if userID == 0 {
                http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
                return
            }

            ctx := context.WithValue(r.Context(), userIDKey, userID)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### Data Layer Pattern

Every query function gains a `userID int64` parameter:

```go
// Before (single-user):
func (db *DB) ListPractices(activeOnly bool) ([]*Practice, error) {
    rows, err := db.Query("SELECT ... FROM practices WHERE active = ?", active)

// After (multi-user):
func (db *DB) ListPractices(userID int64, activeOnly bool) ([]*Practice, error) {
    rows, err := db.Query("SELECT ... FROM practices WHERE user_id = ? AND active = ?", userID, active)
```

### Handler Pattern

Handlers extract user_id from context:

```go
func listPractices(database *db.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        userID := r.Context().Value(userIDKey).(int64)
        practices, err := database.ListPractices(userID, true)
        // ...
    }
}
```

---

## API Tokens & MCP Server

### The Vision

During a scripture study session, you might say: *"Add 'Moroni 10:5' to my memorize queue under the 'Holy Ghost' category, and create a task to study the context of Moroni 10 this week."*

With an API token and MCP server, I (Copilot) can do that for you directly — no copy-pasting, no switching tabs.

### API Token Usage

```bash
# Example: List your practices
curl -H "Authorization: Bearer bec_a1b2c3d4..." https://ibeco.me/api/practices

# Example: Create a new memorize scripture
curl -X POST -H "Authorization: Bearer bec_a1b2c3d4..." \
     -H "Content-Type: application/json" \
     -d '{"name": "Moroni 10:5", "category": "memorize", "reference": "moro/10"}' \
     https://ibeco.me/api/practices

# Example: Check today's progress
curl -H "Authorization: Bearer bec_a1b2c3d4..." https://ibeco.me/api/today
```

### MCP Server (Phase 3.5)

A lightweight MCP server that wraps the Becoming API. Lives in `scripts/becoming-mcp/` and uses the API token for auth.

**Tools it exposes:**

| Tool | Description |
|------|-------------|
| `becoming_create_practice` | Create a new practice (name, category, active, pillar) |
| `becoming_create_task` | Create a task with due date |
| `becoming_log_practice` | Log a practice for today (with optional value/note) |
| `becoming_get_today` | Get today's summary (practices due, tasks due, streak info) |
| `becoming_get_memorize_queue` | Get scripture memorization queue (SM-2 algorithm) |
| `becoming_add_memorize_scripture` | Add a scripture to the memorize queue |
| `becoming_get_progress` | Get report for date range (streaks, completion rates) |
| `becoming_create_note` | Create a note linked to a practice or task |
| `becoming_list_pillars` | List pillars with their linked practices/tasks |

**Configuration:**

```json
// .vscode/mcp.json (or VS Code settings)
{
  "servers": {
    "becoming": {
      "command": "becoming-mcp",
      "args": ["--api-url", "http://localhost:8080", "--token-file", ".env"]
    }
  }
}
```

**Study session workflow:**

1. We're studying Moroni 10 together
2. I find a verse worth memorizing → call `becoming_add_memorize_scripture`
3. I notice a pattern worth tracking → call `becoming_create_practice`
4. I check your memorization progress → call `becoming_get_memorize_queue`
5. You see all of this reflected in your Becoming app immediately

The MCP server is a thin wrapper — it reads the API token from `.env` or a config file and translates MCP tool calls into HTTP requests to the Becoming API. No business logic in the MCP server itself.

---

## Frontend Changes

### Login Page

```
┌─────────────────────────────────────────────┐
│                                             │
│           Welcome to Become                 │
│                                             │
│   "Whatever principle of intelligence..."   │
│                     — D&C 130:18            │
│                                             │
│      Email:    [____________________]       │
│      Password: [____________________]       │
│                                             │
│      ┌─────────────────────────────┐        │
│      │       Sign In               │        │
│      └─────────────────────────────┘        │
│                                             │
│      Don't have an account? Register        │
│                                             │
│      ─────── or sign in with ───────        │
│                                             │
│      ┌─────────────────────────────┐        │
│      │  🔵 Sign in with Google     │  ← only│
│      └─────────────────────────────┘  shown │
│                                       when  │
│                                     enabled │
└─────────────────────────────────────────────┘
```

### Changes Required

1. **Auth guard** — Vue Router navigation guard. If `GET /api/me` returns 401, redirect to `/login`
2. **LoginView.vue** — Email/password form + conditional OAuth buttons
3. **RegisterView.vue** — Create account form
4. **User menu** — Top-right avatar/name with dropdown: Profile, Tokens, Logout
5. **ProfileView.vue** — Name editing, password change, API token management
6. **TokensView.vue** — List tokens, create new, revoke existing
7. **API client** — Handle 401 responses globally (redirect to login)

### API Token Management UI

```
┌─────────────────────────────────────────────┐
│  API Tokens                                 │
│                                             │
│  These tokens allow external tools to       │
│  access your Becoming data on your behalf.  │
│                                             │
│  ┌─────────────────────────────────────────┐│
│  │ 🔑 Copilot scripture study              ││
│  │    bec_a1b2...  Created Feb 12          ││
│  │    Last used: 2 hours ago    [Revoke]   ││
│  └─────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────┐│
│  │ 🔑 Backup script                        ││
│  │    bec_f3g4...  Created Jan 5           ││
│  │    Last used: never          [Revoke]   ││
│  └─────────────────────────────────────────┘│
│                                             │
│  [+ Create New Token]                       │
│                                             │
└─────────────────────────────────────────────┘
```

### Local Development Mode

The `-dev` flag currently enables CORS for Vite dev server. Extend it to also bypass auth:

```go
if devMode {
    // Skip auth for local development
    // Auto-create/login as user_id=1
}
```

This preserves the current development workflow while auth is being built.

---

## Environment & Config

### .env Pattern

We use `.env` files for secrets and configuration. **`.env.example` is checked into git** with placeholder values. **`.env` is gitignored** and contains real values.

#### .env.example (checked in)
```bash
# Becoming Server Configuration
# Copy this file to .env and fill in real values

# Database path (relative or absolute)
BECOMING_DB=./becoming.db

# Session secret (generate with: openssl rand -hex 32)
BECOMING_SESSION_SECRET=change-me-to-a-random-string

# Base URL for the app (used for OAuth callbacks)
BECOMING_BASE_URL=http://localhost:8080

# Google OAuth (optional — leave empty to disable Google sign-in)
BECOMING_GOOGLE_CLIENT_ID=
BECOMING_GOOGLE_CLIENT_SECRET=

# API token for MCP server (generated from the app's Profile > API Tokens page)
# BECOMING_API_TOKEN=bec_your_token_here
```

#### Server reads .env automatically

```go
// On startup, load .env if present (godotenv or manual)
// Then check environment variables, then fall back to CLI flags
// Priority: env var > CLI flag > default
```

### Server Flags (still supported, overridden by env vars)

```
-db           Path to SQLite database
-scriptures   Path to scripture files
-dev          Development mode — CORS + skip auth
-port         Port to listen on (default 8080)
```

---

## Build Order

### Sprint 1: Users, Sessions & Auth (Backend)
**Scope:**
- `users`, `sessions`, `api_tokens` tables
- `.env` loading (godotenv or manual parser)
- `POST /auth/register` — bcrypt password hash, create user, create session
- `POST /auth/login` — verify password, create session
- `POST /auth/logout` — destroy session
- `GET /api/me` — return user profile
- Session cookie handling (HttpOnly, Secure, SameSite)
- AuthRequired middleware (cookie + Bearer token + dev mode)
- `-dev` flag extended to auto-login as user_id=1
- **All existing features still work** — dev mode means nothing changes locally

**Estimated: 3-4 hours**

### Sprint 2: Frontend Auth (Gates & Forms)
**Scope:**
- LoginView.vue — email/password form
- RegisterView.vue — create account form
- Vue Router auth guard (check `/api/me`, redirect to `/login`)
- Global 401 handler in api.ts
- User name in nav bar + logout button
- Profile dropdown (basic)

**Estimated: 2-3 hours**

### Sprint 3: Data Isolation (Multi-User Tenancy)
**Scope:**
- `user_id` column added to all existing tables (migration)
- Recreate `reflections` table with `UNIQUE(user_id, date)` constraint
- Every DB query function gets `userID` parameter
- Every handler extracts `userID` from context
- `SeedPrompts` and `HasPillars` become per-user
- Test with two users — data is fully isolated

**Estimated: 4-5 hours** (most tedious — many function signatures change)

### Sprint 4: API Tokens
**Scope:**
- `api_tokens` table
- `POST /api/tokens` — generate token, return raw token once, store bcrypt hash
- `GET /api/tokens` — list tokens (prefix, name, created, last_used)
- `DELETE /api/tokens/{id}` — revoke token
- Bearer token validation in AuthRequired middleware
- TokensView.vue — manage tokens from profile page
- Test: `curl -H "Authorization: Bearer bec_..."` works

**Estimated: 2-3 hours**

### Sprint 5: Google OAuth (Optional Identity)
**Scope:**
- Google OAuth flow (`/auth/google`, `/auth/callback`) — only active when env vars set
- `GET /api/auth/providers` — tells frontend which sign-in methods are available
- User find-or-create on callback (link to existing email if match)
- LoginView.vue shows Google button conditionally
- Go dependency: `golang.org/x/oauth2`

**Estimated: 2-3 hours**

### Sprint 6: Deployment (Dokploy)
**Scope:**
- Dockerfile + docker-compose.yml for Go binary + PostgreSQL
- Dokploy GitHub app for auto-deploy on push to `main`
- Custom domain setup (ibeco.me)
- TLS via Traefik reverse proxy
- Environment variable configuration in Dokploy UI
- DNS setup
- Smoke test: register, login, create practice on ibeco.me

**Estimated: 2-3 hours**

### Sprint 7: MCP Server
**Scope:**
- `scripts/becoming-mcp/` — Go binary using the MCP SDK
- Reads API token from env or config file
- Exposes tools: create_practice, create_task, log_practice, get_today, get_memorize_queue, add_memorize_scripture, get_progress, create_note, list_pillars
- VS Code MCP configuration (`.vscode/mcp.json`)
- Test: Copilot creates a practice via MCP tool during study session

**Estimated: 3-4 hours**

### Sprint 8: Account Management & Polish
**Scope:**
- ProfileView.vue — name editing, password change
- Delete account (with confirmation) — cascades to all user data
- Session management — view active sessions, revoke others
- Data export (JSON download of all your practices, logs, notes, reflections)

**Estimated: 2-3 hours**

### Total estimated: ~20-28 hours

---

## Security Considerations

| Concern | Mitigation |
|---------|------------|
| Password storage | bcrypt with cost 12. Never store plaintext. |
| CSRF | SameSite=Lax cookies. POST-only state-changing endpoints. |
| XSS | HttpOnly cookies (JS can't read session token). |
| Session fixation | New session ID on every login. |
| Session hijacking | Secure flag (HTTPS only), rotate session on sensitive ops. |
| API token leakage | Tokens stored as bcrypt hashes. Raw token shown once on creation. `bec_` prefix for easy identification in logs. |
| Brute force | Rate limiting on `/auth/login` (e.g., 5 attempts per minute per IP). |
| OAuth state replay | Single-use state tokens with 5-minute expiry. |
| Data leakage | Every query scoped by user_id. API tokens scoped to their owner. |
| DB concurrency | PostgreSQL in production. SQLite for local dev (WAL mode). |
| Backups | PostgreSQL volume on VPS + periodic `pg_dump` to object storage. |

---

## Migration Path

### From single-user to multi-user

1. On first run with auth enabled, existing data is assigned to `user_id=1`
2. First registration creates the user — if it's the original user, they get user_id=1 and all their existing data
3. New users start fresh (user_id=2, 3, ...)
4. `-dev` mode continues to work as before (auto user_id=1, no auth)

### From SQLite to Postgres (future, if needed)

Production already runs PostgreSQL via Dokploy (Docker Compose). If scaling is needed:
1. Add connection pooling (pgbouncer)
2. Enable Row-Level Security for defense-in-depth
3. Scale horizontally (multiple app instances OK with Postgres)
4. Consider managed PostgreSQL if VPS ops becomes a burden

---

## Future Considerations (Not Building Now)

### Apple Sign-In (Phase 4+)
Required if we submit to the App Store. Similar OAuth flow but with Apple's OIDC quirks (private relay email, name only on first sign-in). Add when App Store is a goal.

### Two-factor Authentication (Phase 5+)
TOTP (Google Authenticator / Authy) for sensitive accounts. Low priority for a personal practice tracker, but good hygiene if the app stores meaningful personal data.

### Admin Dashboard (Phase 5+)
For monitoring:
- User count, active users, session count
- Storage usage per user
- Error logs
- Feature flags

### Token Scopes (Phase 5+)
API tokens currently get full access to the user's data. Future: optional scopes like `read:practices`, `write:practices`, `read:memorize`, etc. Not needed until we have third-party integrations beyond our own MCP server.

---

## How This Changes the Architecture

Before Phase 3:
```
Browser ──► Go Binary ──► SQLite (one user)
                │
            go:embed (SPA)
```

After Phase 3:
```
                                          ┌──► SQLite (multi-user)
Browser ──► Dokploy ──► Go Binary ──────────┤
   │            │           │              └── go:embed (SPA)
   │        HTTPS/TLS      │
   │            │           ▲
ibeco.me    Custom domain   │
webeco.me   (both → same)  │
                            │
Copilot ──► MCP Server ──► API Token ──► /api/* endpoints
Scripts ──► curl/fetch ──► Bearer header ─┘
```

Three paths in, one backend, one database. The Go binary stays a single binary with embedded SPA. We add an auth layer in front, a user_id column behind, and an API token path alongside the session cookie.

> "By small and simple things are great things brought to pass." — Alma 37:6

Authentication is the small hinge on which the door to community swings. API tokens are the bridge that lets AI work alongside us in the becoming.
