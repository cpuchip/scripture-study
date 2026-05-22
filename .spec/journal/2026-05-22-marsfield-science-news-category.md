---
date: 2026-05-22
mode: dev + content + substrate audit
workstream: WS7
project: marsfield.org
title: "marsfield.org — Science News category added + first post live; audit revealed no automated science pipeline exists yet"
status: shipped + verified in production. commit 809a7c9 pushed; Dokploy auto-redeployed; cat-science-news CSS class + science-news routes confirmed in live bundle.
carry_forward:
  - "**Substrate gap: no science-news scheduled pipeline.** Only ai-news-7am exists in `pe7-scheduled-pipelines-fire-and-tick.sql:185`. Michael's recollection was wrong. The physics-news file was a manual NewWork.vue dispatch on 2026-05-11 (commit 220cf35), not a scheduled job. To automate, add a scheduled_pipelines row with pipeline_family=research-summary and a physics-news input_template. Surfaced as a proposal task, not yet built — substrate work warrants its own review."
  - "**Date provenance flag on the physics-news content.** Headline + URLs are May 2025; substrate generated the roundup May 2026. Likely the LLM (4.7, Jan 2026 cutoff) defaulted to year-old physics when asked for 'recent.' For the live pipeline, we'll need a real web-search step (Exa) or the digest will keep pulling stale training-data content."
  - "**Republish workflow validated.** Substrate `research-write` output is a strong source for marsfield.org content with minimal hand-editing. Frontmatter + honesty preamble + h3→h2 promotion was the only transformation. Format candidates for future: yt-secular-digest, research-summary, brainstorm outputs."
  - "**.claude/cache/last-ground.txt is dirty.** Not committed (excluded from `git add`). Local-session artifact. Worth adding to .gitignore."
links:
  - "../../projects/marsfield.org/content/science-news/2026-05-22-physics-roundup-may-2025.md"
  - "../../research/physics-news-20260503-science-center-roundup.md  (source)"
  - "../../projects/pg-ai-stewards/extension/pe7-scheduled-pipelines-fire-and-tick.sql  (where the science-news schedule should land)"
commit: 809a7c9
---

# 2026-05-22 — Science News category live on marsfield.org

## What happened

Michael asked: (a) does the substrate have an automated science-news
pipeline like the ai-news daily digest, and (b) can we take what's been
gathered so far and publish it as a "week of news" post on marsfield.org
under a new `science-news` category, skipping the experimental output
until he reviews it.

The honest answer to (a) was no — surfaced as soon as I had the data.
The answer to (b) needed a finer cut once I traced provenance.

## Substrate audit — what actually exists

One scheduled pipeline: `ai-news-7am`. Defined in
`projects/pg-ai-stewards/extension/pe7-scheduled-pipelines-fire-and-tick.sql:185`.
Fires weekdays 13:00 UTC, dispatches the `research-summary` pipeline_family
with an AI-news input_template, materializes to `study/daily-digest/`. No
science pipeline. No physics pipeline. No general-science pipeline.

`research/physics-news-20260503-science-center-roundup.md` was *not* a
scheduled output. Git history (commit `220cf35`, 2026-05-11) names it as
a one-off `research-write` dispatch through the `NewWork.vue` UI. The
same commit fixed the NewWork.vue path bug discovered during that run.
Michael wrote a custom input prompt; the substrate produced the file;
Michael committed it.

The substrate has the *engine* to run a science-news schedule (PE-B
machinery: `scheduled_pipelines` table + plpgsql cron parser +
dispatcher tick). What it doesn't have is the *seed row* + tuned input
template. That's a small, isolated piece of work — deferred for
Michael's review rather than added unilaterally because substrate
changes deserve a council moment of their own.

## Provenance flag on the content

The physics-news file's headline reads *"May 3–11, 2025"* and every
phys.org URL cited carries `/news/2025-05-...`. The file was generated
2026-05-03. The substrate (or the LLM driving it) appears to have
defaulted to year-old physics when asked for "recent" news, presumably
because Claude Opus 4.7's January 2026 training cutoff puts the
freshest dense reporting in 2025-Q3-Q4. The findings themselves are
real and the exhibit ideas travel; the news cycle they describe is
genuinely stale.

I made this explicit in the post header (the indented note at the top
of the republished file) rather than hiding it. Two reasons: honesty
keeps the audience trust; and it names the gap the eventual scheduled
pipeline needs to fix (real web search, not training-data recall).

## What shipped

Six files in one commit (`809a7c9`):

- `src/composables/useContent.ts` — added `'science-news'` to the
  category modules map.
- `src/router/index.ts` — two routes for `/science-news` and
  `/science-news/:slug`.
- `src/components/LcarsNav.vue` — nav pill, positioned between
  Learning and Research.
- `src/views/CategoryView.vue` — label + description for `science-news`.
- `src/styles/lcars.css` — `.tag-pill.cat-science-news` in peach.
- `content/science-news/2026-05-22-physics-roundup-may-2025.md` —
  the republished physics-news roundup with frontmatter + honesty
  preamble + h3→h2 promotion for revealSections animation.

## Verification

1. `npm run build` clean (same inherited gray-matter eval + chunk-size
   warnings as cpuchip).
2. `docker build -t marsfield-web:dev .` clean.
3. `docker run -d -p 18080:80 marsfield-test` — three new routes
   served 200: `/`, `/science-news`, `/science-news/2026-05-22-physics-roundup-may-2025`.
4. `git push origin main` — Dokploy auto-redeployed within ~45 seconds.
5. Live verification: `https://marsfield.org/` now serves the new
   asset hashes (`index-nd1-5RbB.js`, `index-DQ2o5mTR.css`). The CSS
   bundle contains `cat-science-news`; the JS bundle contains 4
   `science-news` references. Live `/science-news` returns 200.

End-to-end stewardship loop confirmed for marsfield.org:
**edit → build → docker verify → commit → push → live in under a minute.**

## Lessons

- **First instinct was wrong; investigation paid.** When the user said
  "use what we've gathered so far," my first read was "publish the
  three days of AI digest under a science-news label." That would have
  been mislabeled content from day one. Pausing to look in `research/`
  (not just `study/daily-digest/`) surfaced the right artifact. The
  `surface_tensions` covenant did its job — the user's recall about a
  science pipeline existing was wrong, but the artifact they remembered
  *did* exist, and naming both halves let us land on the right move.
- **Date provenance matters in republished AI output.** The substrate
  produced a polished document with a confidently-wrong date span.
  Locally I knew; the public reader would not. Naming it in the post
  header costs ~50 words and prevents a class of credibility failure
  that's specific to LLM-sourced content.
- **The category-add pattern is now five edits.** Documented in
  `projects/marsfield.org/CLAUDE.md` already, exercised here cleanly.
  Composables, router, nav, view labels, pill color. Worth keeping
  this list in CLAUDE.md current — future me will add `events` or
  `exhibits` the same way.
- **Stewardship grant works.** First exercise of the commit+push
  privilege Michael granted earlier in the session. Dokploy picked up
  the push, redeployed within ~45s, and the live bundle confirmed the
  new content. The trust loop is real, observable, and fast.
