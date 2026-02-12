# Becoming App — Phase 3: Authentication & Multi-User Foundation

*Created: February 12, 2026*
*Context: Multi-user foundation required before social features (webeco.me) and hosting*
*Domains: ibeco.me (personal app) + webeco.me (social/group — future)*
*Prerequisite: Phase 2.7 complete (Pillars, Notes, Reflections, Trends)*

---

## Overview

The app is currently single-user with no authentication — it's a local binary serving one SQLite database. To go multi-user, we need to solve three problems simultaneously:

1. **Who is this?** — Authentication (identity)
2. **What's theirs?** — Data isolation (tenancy)
3. **Where does it live?** — Hosting and deployment

These are deeply coupled. The auth strategy determines the data model, the data model determines the hosting requirements, and the hosting environment constrains what auth flows are practical.

### Why Now?

Phase 2.7 completed the solo instrument. Every feature from here forward (shared pillars, group reflections, accountability partners) requires knowing *who* the user is. Auth is the unlock for everything social. And it needs to be right the first time — migrating auth systems after users have data is painful.

---

## Auth Strategy Analysis

### Option A: OAuth Only (Google, Apple, GitHub)

| Pros | Cons |
|------|------|
| No password storage or management | Dependent on third-party availability |
| Users trust Google/Apple sign-in | Can't work offline without a session cache |
| Less security surface area for us | Limited to providers we integrate |
| Mobile-friendly (native SDKs) | Apple Sign-In required if we do App Store |
| Email verified by default | User may have multiple accounts (confusion) |

### Option B: Email + Password (self-managed)

| Pros | Cons |
|------|------|
| Full control, no third-party dependency | Must handle password hashing, reset flows |
| Works anywhere | Phishing/credential stuffing surface |
| Simplest mental model for users | Email verification needed |
| Offline-capable auth | More code to write and maintain |

### Option C: Hybrid (OAuth + optional email/password)

| Pros | Cons |
|------|------|
| Maximum flexibility | Most complex to implement |
| Users choose their preferred method | Account linking complexity |
| Can start with OAuth, add password later | Two codepaths to maintain |
| Best UX — "Sign in with Google" + fallback | |

### Recommendation: **Option C (Hybrid), starting with OAuth**

Start with Google OAuth (largest market share, trusted by target audience). Add Apple if/when App Store is a goal. Keep the door open for email/password by designing the user model to support it. Don't build email/password in Sprint 1 unless there's demand.

**Rationale for Google first:**
- Our target users (gospel-studying Latter-day Saints) overwhelmingly have Google accounts
- Google OAuth is free, well-documented, excellent Go libraries
- Works on mobile browsers (ibeco.me) without native SDKs
- `credentials` package in Go ecosystem is mature

---

## Architecture Decisions

### Decision 1: Session Strategy

| # | Option | Decision |
|---|--------|----------|
| 1 | JWT tokens (stateless) | **No** — JWTs can't be revoked, leak claims to the client, and encourage bad patterns (storing in localStorage). Overkill for our scale. |
| 2 | Server-side sessions (cookie) | **Yes** — HttpOnly secure cookie with session ID. Session stored in SQLite. Simple, secure, revocable. Works with SameSite=Lax for CSRF protection. |
| 3 | Token + refresh (API-style) | **No** — We're not building a public API. The SPA talks to its own backend. Cookies are the right tool. |

**Session details:**
- HttpOnly, Secure, SameSite=Lax cookie named `becoming_session`
- Session stored in `sessions` table with user_id, created_at, expires_at, last_active
- Auto-extend session on activity (sliding window, 30-day max)
- Session cleanup job (delete expired on startup and periodically)

### Decision 2: Data Isolation

| # | Option | Decision |
|---|--------|----------|
| 1 | Shared tables with user_id column | **Yes** — Simple, works with SQLite, easy to query. Every table gets a `user_id INTEGER NOT NULL` column. |
| 2 | Separate SQLite file per user | **No** — Complicates backups, migrations, and cross-user queries (future social features). |
| 3 | Postgres with RLS | **Premature** — If we outgrow SQLite, this is the migration path. But not now. |

**Migration approach:**
- Add `user_id` to all existing tables (practices, logs, tasks, notes, reflections, pillars, etc.)
- Existing data gets assigned to user_id=1 (the current solo user becomes user 1)
- All queries add `WHERE user_id = ?` — enforced at the data layer
- API middleware injects user_id from session before reaching handlers

### Decision 3: OAuth Flow

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

### Decision 4: Hosting

| # | Option | Cost | Fit |
|---|--------|------|-----|
| 1 | **Fly.io** | Free tier: 3 shared VMs, 1GB persistent volume | **Best fit** — single Go binary deploys trivially. SQLite works with persistent volumes. Free HTTPS. Custom domains supported. |
| 2 | Railway | $5/mo hobby | Good, but costs from day one. |
| 3 | DigitalOcean droplet | $6/mo | Full control but more ops work. |
| 4 | Cloudflare Workers + D1 | Free tier generous | Interesting but requires refactoring to Workers runtime (no Go). |
| 5 | Self-hosted (home server) | $0 | Good for development, bad for reliability/HTTPS. |
| 6 | Vercel/Netlify (static) + separate API | Varies | Over-complicates the architecture. Our SPA is embedded in the Go binary. |

**Recommendation: Fly.io**
- Go binary deploys as a Docker container or direct binary
- Persistent volume for SQLite (`becoming.db` on `/data/`)
- Free HTTPS with custom domains (ibeco.me, webeco.me)
- Free tier covers early usage
- Easy `fly deploy` from CI/CD
- Scales to multiple regions if needed later

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

## Database Changes

### New Tables

```sql
-- Users (identity)
CREATE TABLE IF NOT EXISTS users (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    email       TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL DEFAULT '',
    avatar_url  TEXT NOT NULL DEFAULT '',
    provider    TEXT NOT NULL DEFAULT 'google',  -- 'google', 'apple', 'email'
    provider_id TEXT NOT NULL DEFAULT '',         -- OAuth subject ID
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_login  DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_provider ON users(provider, provider_id);

-- Sessions
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

-- OAuth state (CSRF protection for auth flow)
CREATE TABLE IF NOT EXISTS oauth_states (
    state       TEXT PRIMARY KEY,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    redirect_to TEXT NOT NULL DEFAULT '/'
);
```

### Migration of Existing Tables

Every existing table gets a `user_id` column:

```sql
ALTER TABLE practices ADD COLUMN user_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE practice_logs ADD COLUMN user_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE tasks ADD COLUMN user_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE notes ADD COLUMN user_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE reflections ADD COLUMN user_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE prompts ADD COLUMN user_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE pillars ADD COLUMN user_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE practice_pillars ADD COLUMN user_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE task_pillars ADD COLUMN user_id INTEGER NOT NULL DEFAULT 1;

-- Indexes for every query path
CREATE INDEX IF NOT EXISTS idx_practices_user ON practices(user_id);
CREATE INDEX IF NOT EXISTS idx_practice_logs_user ON practice_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_tasks_user ON tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_notes_user ON notes(user_id);
CREATE INDEX IF NOT EXISTS idx_reflections_user ON reflections(user_id);
CREATE INDEX IF NOT EXISTS idx_pillars_user ON pillars(user_id);
```

---

## API Changes

### New Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/auth/google` | Initiate Google OAuth flow |
| GET | `/auth/callback` | Handle OAuth callback, create session |
| POST | `/auth/logout` | Destroy session, clear cookie |
| GET | `/api/me` | Get current user profile (name, email, avatar) |
| PUT | `/api/me` | Update user profile (name) |
| DELETE | `/api/me` | Delete account and all data |

### Middleware

```go
// AuthRequired middleware — placed on all /api/* routes
func AuthRequired(db *DB) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            cookie, err := r.Cookie("becoming_session")
            if err != nil {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }
            session, err := db.GetSession(cookie.Value)
            if err != nil || session.IsExpired() {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }
            // Touch session (sliding window)
            db.TouchSession(session.ID)
            // Inject user_id into context
            ctx := context.WithValue(r.Context(), userIDKey, session.UserID)
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

## Frontend Changes

### Auth Flow

```
┌─────────────────────────────────────────────┐
│                                             │
│           Welcome to Become                 │
│                                             │
│   "Whatever principle of intelligence..."   │
│                     — D&C 130:18            │
│                                             │
│      ┌─────────────────────────────┐        │
│      │  🔵 Sign in with Google     │        │
│      └─────────────────────────────┘        │
│                                             │
│      ┌─────────────────────────────┐        │
│      │  🍎 Sign in with Apple      │        │
│      └─────────────────────────────┘        │
│                                             │
│      (more options coming soon)             │
│                                             │
└─────────────────────────────────────────────┘
```

### Changes Required

1. **Auth guard** — Vue Router navigation guard. If `GET /api/me` returns 401, redirect to `/login`
2. **LoginView.vue** — OAuth buttons, welcome message
3. **User menu** — Top-right avatar/name with dropdown: Profile, Logout
4. **Profile settings** — Name, connected accounts, delete account
5. **API client** — Handle 401 responses globally (redirect to login)

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

### Server Flags

```
-db           Path to SQLite database (existing)
-scriptures   Path to scripture files (existing)
-dev          Development mode — CORS + skip auth (existing, extended)
-google-client-id      Google OAuth client ID (new)
-google-client-secret  Google OAuth client secret (new)
-session-secret        Secret for cookie signing (new)
-base-url              Public URL (e.g., https://ibeco.me) for OAuth callbacks (new)
```

### Environment Variables (for deployment)

```
BECOMING_DB=/data/becoming.db
BECOMING_GOOGLE_CLIENT_ID=xxx
BECOMING_GOOGLE_CLIENT_SECRET=xxx
BECOMING_SESSION_SECRET=xxx
BECOMING_BASE_URL=https://ibeco.me
```

### Google OAuth Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create project "Becoming"
3. Enable "Google Sign-In" API
4. Create OAuth 2.0 credentials (Web Application)
5. Add authorized redirect URIs:
   - `https://ibeco.me/auth/callback`
   - `https://webeco.me/auth/callback`
   - `http://localhost:8080/auth/callback` (development)
6. Save Client ID + Client Secret

---

## Build Order

### Sprint 1: Users & Sessions (Backend foundation)
**Scope:**
- `users`, `sessions`, `oauth_states` tables
- Session CRUD (create, get, touch, delete, cleanup)
- `user_id` column added to all existing tables (migrated)
- AuthRequired middleware
- `/api/me` endpoint
- `-dev` flag extended to auto-login as user_id=1
- **All existing tests still pass** — dev mode means nothing changes for local use

**Estimated: 3-4 hours**

### Sprint 2: Google OAuth (Identity)
**Scope:**
- Google OAuth flow (`/auth/google`, `/auth/callback`)
- User find-or-create on callback
- Session cookie set on successful auth
- `/auth/logout` endpoint
- Go dependency: `golang.org/x/oauth2`

**Estimated: 2-3 hours**

### Sprint 3: Frontend Auth (Gates)
**Scope:**
- LoginView.vue with Google sign-in button
- Vue Router auth guard (check `/api/me`, redirect to `/login`)
- Global 401 handler in api.ts
- User avatar + name in nav bar
- Logout button
- Profile dropdown (basic)

**Estimated: 2-3 hours**

### Sprint 4: Data Isolation (Tenancy)
**Scope:**
- Every DB query function gets `userID` parameter
- Every handler extracts `userID` from context
- Test with two users — data is fully isolated
- Seed prompts become per-user (copy on first login)
- Default pillars onboarding triggers per-user

**Estimated: 4-5 hours** (most tedious — many function signatures change)

### Sprint 5: Deployment (Fly.io)
**Scope:**
- Dockerfile for the Go binary
- `fly.toml` configuration
- Persistent volume for SQLite
- Custom domain setup (ibeco.me, webeco.me)
- HTTPS via Fly.io managed certificates
- Environment variable configuration
- DNS setup for both domains
- Smoke test: sign in with Google on ibeco.me

**Estimated: 2-3 hours**

### Sprint 6: Account Management
**Scope:**
- ProfileView.vue — name editing, connected providers, session list
- Delete account (with confirmation) — cascades to all user data
- Session management — view active sessions, revoke others
- Data export (JSON download of all your practices, logs, notes, reflections)

**Estimated: 2-3 hours**

### Total estimated: ~15-21 hours

---

## Security Considerations

| Concern | Mitigation |
|---------|------------|
| CSRF | SameSite=Lax cookies + verify OAuth state parameter |
| XSS | HttpOnly cookies (JS can't read session token) |
| Session fixation | New session ID on every login |
| Session hijacking | Secure flag (HTTPS only), rotate session on sensitive ops |
| OAuth state replay | Single-use state tokens with 5-minute expiry |
| Data leakage | Every query scoped by user_id. No admin endpoints yet. |
| SQLite concurrency | WAL mode (already enabled). Fly.io single instance for now. |
| Backups | Fly.io volume snapshots + periodic `sqlite3 .backup` to object storage |

---

## Migration Path

### From single-user to multi-user

1. On first run with auth enabled, existing data is assigned to `user_id=1`
2. First Google sign-in creates user_id=1 with that Google account
3. From that point, user_id=1 *is* the original user with all their data
4. New users get fresh databases (user_id=2, 3, ...)
5. `-dev` mode continues to work as before (auto user_id=1, no auth)

### From SQLite to Postgres (future, if needed)

If Becoming outgrows SQLite (concurrent writes from many users on a single Fly.io instance):
1. Switch to Fly.io Postgres (managed)
2. Migrate schema 1:1 (SQLite → Postgres is straightforward)
3. Add connection pooling
4. Enable Row-Level Security for defense-in-depth
5. Scale horizontally (multiple app instances OK with Postgres)

This is a Phase 5+ concern. SQLite on Fly.io handles hundreds of concurrent users in WAL mode with a single writer. We'll know when we outgrow it.

---

## Future Considerations (Not Building Now)

### Apple Sign-In (Phase 3.5)
Required if we submit to the App Store. Similar OAuth flow but with Apple's OIDC quirks (private relay email, name only on first sign-in). Add when App Store is a goal.

### Email/Password (Phase 4+)
For users who don't want OAuth. Requires:
- Bcrypt password hashing
- Email verification flow (send confirmation link)
- Password reset flow (send reset link)
- Rate limiting on login attempts
- Depends on email sending service (Resend, Postmark, SES)

### Two-factor Authentication (Phase 5+)
TOTP (Google Authenticator / Authy) for sensitive accounts. Low priority for a personal practice tracker, but good hygiene if the app stores meaningful personal data.

### Admin Dashboard (Phase 5+)
For monitoring:
- User count, active users, session count
- Storage usage per user
- Error logs
- Feature flags

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
Browser ──► Fly.io ──► Go Binary ──► SQLite (multi-user)
   │            │           │
   │        HTTPS/TLS   go:embed (SPA)
   │            │
ibeco.me    Custom domain
webeco.me   (both → same app)
```

The Go binary stays a single binary with embedded SPA — the architecture doesn't change fundamentally. We add an auth layer in front and a user_id column behind. The deployment wrapper (Fly.io) handles HTTPS, DNS, and persistence.

> "By small and simple things are great things brought to pass." — Alma 37:6

Authentication is the small hinge on which the door to community swings.
