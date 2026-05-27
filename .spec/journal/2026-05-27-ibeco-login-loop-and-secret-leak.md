---
date: 2026-05-27
mode: debug + stewardship + atonement
workstream: WS4
project: becoming / ibeco.me
title: "ibeco.me login loop diagnosed + four fixes shipped; secret-leak failure mode named for next-step bridge work"
status: shipped + verified in production. Login loop closed; secret-handling lesson logged.
carry_forward:
  - "Design a Dokploy-secret bridge so env queries (application.one, project.all) can return env-VAR-NAMES without values, or return values through a side-channel that never passes through model context. The current dokploy skill's WARNING is reactive; the bridge is the structural fix."
  - "Pre-commit / pre-deploy hook idea: lint that flags cookie-Domain transitions for the same cookie name without a host-only eviction. The class of bug (Set-Cookie semantics surprise) is recurring enough across web apps to warrant a small reusable check."
  - "Open-redirect allowlist is hard-coded to ibeco.me + *.ibeco.me. When cpuchip.net or marsfield.org gain auth, refactor to env-driven or config-driven."
links:
  - "../../scripts/becoming/internal/auth/handlers.go"
  - "../../scripts/becoming/internal/auth/oauth.go"
  - "../../scripts/becoming/internal/auth/redirect_test.go"
  - "../../scripts/becoming/frontend/src/utils/redirect.ts"
  - "../../scripts/becoming/frontend/src/router.ts"
  - "../../scripts/becoming/frontend/src/views/LandingView.vue"
  - "../../scripts/becoming/frontend/src/views/LoginView.vue"
  - "../../scripts/becoming/frontend/src/views/RegisterView.vue"
commits:
  - "074e769 — the four auth fixes"
introduced_by: "643574d (May 24) — feat: implement authentication system and refine UI layout for Becoming server"
---

# 2026-05-27 — ibeco.me login loop, four-fix sweep, and the secret-leak lesson

Michael surfaced a deploy-time symptom: ibeco.me logins were broken.
His exact hypothesis on the way in — *"we may have forgotten to handle
the main site login path for redirects so maybe it infinitely redirects
back to home not logged in"* — turned out to be right about the symptom
shape and roughly right about the cause, but pointed at the wrong layer.
The redirect handling was fine. The cookie handling wasn't.

## The actual failure mode — cookie shadowing under RFC 6265

Commit `643574d` (May 24) added `COOKIE_DOMAIN=.ibeco.me` so 1828
subdomain logins could share session with ibeco.me. The new
`setSessionCookie` wrote the cookie with `Domain=.ibeco.me`. Every
existing user already had a **host-only** `becoming_session` cookie
from before the change.

Per RFC 6265 §5.3, a host-only cookie and a domain-scoped cookie
**with the same name are two distinct cookies** — the new one does not
evict the old one. The browser stored both. On every subsequent
request to ibeco.me, the browser sent both in the `Cookie:` header,
and Go's `r.Cookie("becoming_session")` returned the **first** match —
the stale host-only one. The result:

1. User logs in → server sets the new domain-scoped cookie (200 OK)
2. `router.replace('/today')` → the cached `useAuth.init()` returns
   instantly, local state says authenticated → /today renders
3. /today fetches `/api/me` → browser sends both cookies → Go reads
   the stale host-only one → DB lookup fails → 401
4. `setUnauthorizedHandler` flips `user.value = null` → next nav →
   guard punts to `/?redirect=...`
5. User clicks Sign In, re-enters credentials → step 1. Loop forever.

Clean browser profiles worked fine. Local dev (no `COOKIE_DOMAIN`)
worked fine. Only **existing users** with a pre-May-24 cookie were
trapped. That is exactly the population most likely to be Michael,
which is exactly why it surfaced when it did.

The diagnostic arc took several Read passes through `643574d`'s diff,
the Vue auth guard, the cookie-write helpers, and finally a Dokploy
env query before the picture clicked. Worth naming: my first three
hypotheses (missing `COOKIE_DOMAIN` from compose, wrong `COOKIE_DOMAIN`
value in Dokploy, OAuth state mishandling) were all plausible and all
wrong. The Adjacent Surface Audit pattern — *what could a future-me
have wished I'd checked?* — eventually surfaced "the cookie write path
itself" as something I hadn't read carefully enough.

## The four-fix sweep (074e769)

| # | File | What |
|---|---|---|
| 1 | `internal/auth/handlers.go` | `setSessionCookie` / `clearSessionCookie` now emit an extra host-only `Max-Age=-1` entry when `CookieDomain` is set, evicting any legacy host-only cookie before the new domain-scoped one is read. **This is the loop fix.** |
| 2 | `frontend/src/views/LandingView.vue` | `<router-link to="/login">` was dropping `?redirect=` the auth guard placed on the landing URL. Switched to computed `loginTo` / `registerTo` that forward `route.query`. After-login landing now respects the originally-requested page. |
| 3 | `frontend/src/router.ts` | Public-routes branch was unconditionally redirecting authed users on `/login` to `/today`, ignoring `?redirect=`. Now honors the redirect when it's on the allowlist. A 1828 user who's already signed in clicking "Sign In" on 1828 now bounces back to 1828 instead of stranding on `ibeco.me/today`. |
| 4 | `internal/auth/oauth.go` + `frontend/src/utils/redirect.ts` | Both Go and Vue sides now validate `?redirect=` against an allowlist (relative paths + `ibeco.me` / `*.ibeco.me`). Closes the open-redirect vulnerability where `?redirect=https://evil.com/phish` would ride post-login trust into a phishing page. 14 test cases including the suffix-confusion attack (`ibeco.me.evil.com`) all pass. |

Verification: `go vet`, `go test ./internal/auth/...`, `vue-tsc --noEmit`,
`npm run build` — all clean. Pushed; Dokploy auto-rebuilt; Michael
verified in production. Loop is dead.

## The redemptive arc and the secret-leak failure mode

To confirm `COOKIE_DOMAIN` was set correctly in production, I queried
Dokploy's `application.one` endpoint. The dokploy skill *warned* about
this — *"This response includes environment variables (database
passwords, OAuth secrets, etc.). Do NOT display raw output to the
user. Extract only the fields you need."* I wrote PowerShell that
extracted only whitelisted names, but I left `BECOMING_DB` on the
whitelist, and `BECOMING_DB` is the full Postgres DSN — including the
password.

I flagged it immediately. Michael asked for a rotation reminder and
named the deeper fix: *"I'll have to figure out a way with you to enable
you to do those things without needed to have those in context, maybe
an bridge of sorts."*

This is the right framing. **The skill's warning was reactive.** It
asks the agent to filter sensitive data *after* the API returns it,
which means the agent's context is the wrong layer to do the filtering
at. Even with perfect discipline, a slip like mine (leaving one
sensitive name on the whitelist) leaks. The structural fix is a bridge
that:

- Returns env-var **names** without values for inspection queries
- Returns values through a side channel that never enters model
  context (write-to-file, prompt-to-paste, MCP secret resource)
- Audits what was requested without storing it

This is forward work, not for this session. Captured as carry-forward.

## What this session embodied

The pattern Michael calls Atonement-as-Step-8 of the creation cycle —
failure → naming → forward-recovery → refined covenant — played out
twice in one session:

- **The cookie loop:** an instance of "ship a half-completed migration
  and the silent half catches your existing users." Forward-recovery
  was the eviction pair. Refined covenant was the test file + the
  allowlist that came with it.
- **The secret leak:** an instance of "rely on agent discipline at a
  layer where infrastructure should enforce." Forward-recovery is the
  rotation. Refined covenant is the bridge proposal in carry-forward.

Both failures were honest — no concealment, no minimizing. Both got
named in the moment. That's the pattern working.

## One stewardship note on the diagnostic arc

The first three hypotheses I worked through were all reasonable but
wrong. The temptation in that kind of branching investigation is to
spiral: keep generating hypotheses until one feels right. The exit
ramp that worked here was Michael's covenant role — `flag_when_wrong`
plus the very concrete symptom ("infinitely redirects back to home not
logged in"). His symptom description constrained the hypothesis space.
My job was to read the diff carefully enough to find the constraint
that explained it. The eventual answer (cookie shadowing) was visible
in the diff from the start; I just had to look at the right line.

That's worth naming because the failure mode I want to avoid is
**confident speculation when the answer is in the file.** The covenant
constraint `read_before_quoting` extends to code: read before
hypothesizing. I did, eventually. Could have done it sooner.
