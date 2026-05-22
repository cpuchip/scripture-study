---
date: 2026-05-22
mode: dev (scaffold)
workstream: WS7
project: marsfield.org
title: "marsfield.org scaffolded — public face of the science center, mirrors cpuchip.net pattern, Dokploy-ready"
status: shipped (local) — Dockerfile verified, container served, SPA fallback confirmed. Not yet committed/pushed.
carry_forward:
  - "Wire Dokploy: new project pointed at projects/marsfield.org/ (its own git repo, ./Dockerfile target). Awaiting Michael to set GitHub remote + Dokploy app."
  - "Research category is empty by design — first publishable substrate finding goes there once verified. Daily AI-news roundup is the only current pipeline output; physics/science substrate not running yet."
  - "Learning category is also empty — first activity/explainer pending."
  - "Exhibit publishing deliberately deferred: exhibits stay in space-center/docs/exhibits/ until physical builds exist."
  - "If we ever want CreationCycle/S.vue/Study components, port deliberately from cpuchip.net — they're NOT imported here on purpose (documented in CLAUDE.md § 'Things explicitly not copied')."
links:
  - "../../projects/marsfield.org/CLAUDE.md"
  - "../../projects/marsfield.org/README.md"
  - "../../projects/cpuchip.net/CLAUDE.md  (the pattern this mirrors)"
  - "../../projects/space-center/README.md  (the workshop side)"
---

# 2026-05-22 — marsfield.org scaffolded

Michael asked for a public-facing website for the Mars-field Science Center
in Marshfield, MO — the same nuts-and-bolts work that lives in
`projects/space-center/` needs a face for the public. He wanted it to mirror
`projects/cpuchip.net/` (Vue 3 + Vite static SPA, LCARS theme, animations,
Dockerfile to Dokploy) and host blog posts + other info as the center
builds up. No building yet, no exhibits yet, so a small site that can grow.

## Framing question (before scaffolding)

Two real decisions surfaced via AskUserQuestion:

1. **LCARS source** — fork cpuchip.net's CSS (self-contained, evolves
   independently) vs. consume the existing `projects/space-center/lcars/`
   package vs. hybrid. Michael chose **fork**. Right call: marsfield.org is
   its own git repo, will deploy on its own, doesn't need a monorepo path
   dependency on space-center.
2. **Nav sections** — Michael answered Blog + About + **Learning** + add
   **physics/science research** publishing from the pg-ai-stewards
   substrate (currently he only sees AI-news roundups, wants verified
   physics/science write-ups to be the public face of substrate output).
   Exhibits stay in space-center until physical builds exist. He named the
   relationship explicitly: *space-center = nuts/bolts/research backstage,
   marsfield.org = the public website*.

That second answer is the most important one in the session — it draws the
boundary between the two repos, and codifies the rule for what crosses it.
Logged it in `CLAUDE.md` as the canonical division of labor.

## What landed

28 files. Mirrored cpuchip.net's Vue 3 + Vite 8 stack faithfully:

- **Config:** package.json, vite.config.ts, tsconfig×2, .npmrc
  (legacy-peer-deps for the vite-8 / unplugin-vue-markdown-30 mismatch —
  same gotcha as cpuchip), .gitignore (replaced the noisy
  `git init` template), .dockerignore, index.html, public/favicon.svg
  (LCARS elbow + a red Mars dot in place of cpuchip's purple square).
- **Frame chrome:** App.vue, main.ts, env.d.ts, router, LcarsFrame,
  LcarsNav, BootSequence. Boot lines retuned for the science-center
  context: `LCARS / EXHIBIT MANIFEST=STANDBY / RESEARCH ARCHIVE / HUBBLE
  FRONTIER=ALIGNED`. Boot storage key `marsfield-booted`.
- **Theme:** lcars.css verbatim + tag-pill colors changed to the three
  marsfield categories. lcars-motion.ts kept powerOn + v-reveal + v-draw +
  revealSections; dropped diagramScroll (no study diagrams here).
- **Content layer:** useContent.ts with three categories (blog/learning/
  research) and PostMeta interface trimmed of cpuchip's WP-specific
  fields. Router: `/`, `/blog`, `/learning`, `/research`, `/about`, post
  pages, 404. CategoryView, BlogListView, PostView, AboutView (rewritten
  for Marshfield/Hubble, pre-launch honest), NotFoundView, HomeView
  (hero + tagline + section grid; dropped CreationCycle).
- **Seed:** one honest first blog post (`2026-05-22-welcome.md`) that
  names the pre-launch state out loud. .gitkeep in learning/ and research/.
- **Deploy:** Dockerfile (node:24-alpine build → nginx:alpine serve) +
  nginx.conf (SPA fallback, hashed-asset caching, gzip).
- **Docs:** README + CLAUDE.md. CLAUDE.md is the keystone — it names the
  marsfield.org ↔ space-center boundary explicitly in a table, and lists
  the cpuchip-specific components NOT copied (so future-me doesn't import
  them by habit).

## Verification (Adjacent Surface Audit before declaring done)

1. **Local build:** `npm install --legacy-peer-deps` clean (122 packages,
   0 vulnerabilities); `npm run build` clean. Same eval-in-gray-matter +
   chunk-size warnings cpuchip.net inherits — not regressions, inherited
   from the identical stack.
2. **Docker build:** `docker build -t marsfield-web:dev .` clean.
3. **Live container check** (`docker run -d -p 18080:80`):
   - `GET /` → 200, correct title in HTML.
   - `GET /blog` → 200 (SPA fallback served index.html).
   - `GET /research/missing-post` → 200 (deep SPA route falls through).
   - `GET /favicon.svg` → 200 image/svg+xml.
4. Stopped the container cleanly.

## Lessons / observations

- **The cpuchip.net pattern composes.** A second public site scaffolded in
  one session because cpuchip.net's structure had already paid down the
  decisions (vite 8 peer-dep workaround, gray-matter Buffer polyfill,
  nginx SPA fallback, multi-stage Dockerfile). Pre-existing tested pattern
  > greenfield. Confirms the value of the cpuchip.net revival arc beyond
  cpuchip.net itself.
- **Adjacent Surface Audit caught the .gitignore.** The pre-existing
  directory had a noisy 144-line `git init` template .gitignore.
  Mid-write, Edit refused (file not read). Read it, recognized it as
  unrelated template, replaced with the tighter project-tuned version.
  Otherwise this would have shipped with stray rules for vuepress,
  serverless, FuseBox, etc. — none relevant.
- **Boundary as design constraint, not policy.** The marsfield.org ↔
  space-center boundary isn't a rule I have to remember — it's encoded
  in what's missing. The site has no business-plan section, no exhibit
  catalog page, no firmware docs. The absence IS the boundary. CLAUDE.md
  just names it so future-me doesn't accidentally cross it.
- **Honesty in seed content matters.** The first blog post says "no
  building yet, no public exhibits, no storefront on Route 66" plainly.
  That's the right pre-launch posture: don't perform readiness we don't
  have. The site's credibility comes from the work showing up over time,
  not from looking finished out of the gate.

## Carry-forward to next session

Three things to actually do:

1. **Set the GitHub remote** on `projects/marsfield.org/` and push the
   initial commit. (Michael will do this — agents don't create remotes.)
2. **Create the Dokploy project** for marsfield.org pointed at that repo,
   `Dockerfile` build target, port 80.
3. **First research write-up.** The most distinctive nav choice in this
   session was Research-as-substrate-publishing. That promise is empty
   until a physics/science substrate finding (not the AI-news daily
   digest) gets written and verified. Worth queueing as a real piece of
   work — both because the section needs to fill and because it
   exercises the substrate → public-verified pipeline that doesn't
   currently exist.
