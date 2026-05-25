# Journal Entry: Subdomain Auth Sharing and Redirection Refinements

**Date:** May 24, 2026  
**Focus:** Authentication, Cookie Sharing, and Subdomain Cross-redirection between `ibeco.me` and `1828.ibeco.me`

---

## 1. Context & Problem Statement

We successfully integrated authentication between the `1828 Illuminated` app and the centralized `becoming` account service (`ibeco.me`). However, during initial integration testing, we ran into two critical friction points:

1. **Subdomain Cookie Isolation**: The `becoming` service issued session cookies without specifying a `Domain` attribute. Browsers defaulted cookie scope strictly to the issuing host (`ibeco.me`), preventing the subdomain client `1828.ibeco.me` from reading the session.
2. **Internal vs. External Redirection**:
   - The login/registration frontend pages in the `becoming` app parsed the `redirect` query parameter but passed it directly to Vue Router (`router.replace(redirect)`). When redirecting back to `https://1828.ibeco.me/...`, Vue Router treated the absolute URL as a local path fragment, breaking navigation.
   - For Google Sign-in, the redirect URL parameter was completely ignored; the OAuth callback always redirected to `/today`.

---

## 2. Technical Implementation & Refinements

We designed and implemented a full-circle redirection and cookie sharing solution across the codebase:

### A. Subdomain Cookie Scope (`becoming` service)
We modified the cookie handlers in `becoming` to scope session identifiers to a wildcard parent domain:
- **`handlers.go`**: Added `CookieDomain` string config to the `auth.Handlers` struct. Updated `setSessionCookie` and `clearSessionCookie` to assign the cookie `Domain` property if configured.
- **`main.go`**: Loaded the `COOKIE_DOMAIN` from environment variables and initialized `auth.Handlers`.
- **Environment**: Documented and set `COOKIE_DOMAIN=.ibeco.me` in the local `.env` and `.env.example` configurations. When empty (e.g. localhost development), it defaults to standard single-host cookies.

### B. Vue Router Absolute URL Bypass (`becoming` frontend)
- Updated both [LoginView.vue](file:///c:/Users/cpuch/Documents/code/stuffleberry/scripture-study/scripts/becoming/frontend/src/views/LoginView.vue) and [RegisterView.vue](file:///c:/Users/cpuch/Documents/code/stuffleberry/scripture-study/scripts/becoming/frontend/src/views/RegisterView.vue):
  - Checked if the `redirect` query parameter is an absolute URL (`startsWith('http://')` or `startsWith('https://')`).
  - If absolute, we bypass the internal Vue Router and force a window cutover using `window.location.href = redirect`.
  - If relative, it continues to use SPA route transitions via `router.replace(redirect)`.

### C. OAuth State Mapping for Subdomains (`becoming` backend)
Google OAuth requires a pre-registered static callback URL. To securely redirect users back to the originating subdomain:
- **State Token Binding**: Updated the state structure in [oauth.go](file:///c:/Users/cpuch/Documents/code/stuffleberry/scripture-study/scripts/becoming/internal/auth/oauth.go) to map the random state CSRF token to a `stateData` struct containing the target `redirectURL`.
- **State Retrieval**: In `GoogleCallback`, we lookup the validated token, retrieve the originating redirect URL, and issue a HTTP redirection to it (defaulting to `/today` if none is bound).
- **Forwarding**: Updated the frontend OAuth trigger buttons to pass the current redirect URL: `/auth/google/login?redirect=...`.

### D. Header Sign-In & Sign-Out (1828 Illuminated)
- **Sign In Link**: Wired the navigation header in `App.vue` to compute `signInUrl` using the current active page (`window.location.href`) as the `redirect` parameter, opening it in the same tab for a seamless loop.
- **Backend Logout Endpoint**: Implemented `POST /api/auth/logout` in the 1828 Illuminated backend. It clears the cookie on both the parent domain (`.ibeco.me`) and the host domain (`1828.ibeco.me`) with a max-age of `-1`.
- **Sign Out Button**: Exposed a `Sign Out` button on the header next to the user profile that calls the logout endpoint and updates session reactivity.

---

## 3. Verification & Build Results

Both subprojects compile and build successfully:
- **Becoming Server**: Copied new build files (`npm run build` in `becoming/frontend`) to the embedding directory, and compiled the Go binary (`go build -o server.exe ./cmd/server/`). No errors.
- **1828 Illuminated Client**: Built Vue/TypeScript client assets successfully (`npm run build`). No compilation or typing errors.
