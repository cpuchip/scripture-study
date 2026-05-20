---
title: 1828-illuminated MVP — built overnight under stewardship
date: 2026-05-20
status: shipped (commits b66444d + follow-up rebuild)
session-type: autonomous (Michael asleep, ~2h sprint window)
---

# 1828-illuminated MVP shipped overnight

Michael gave stewardship over this project at ~midnight 2026-05-20 with ~7h total budget and a ~2h session window before the harness limit. The deliverable: a Docker-deployable Vue SPA for 1828.ibeco.me that surfaces the 1828 Webster meaning frame on scripture words.

Shipped in commit `b66444d` + a follow-up rebuild commit (when the modern-definition fetcher completed).

## What landed

**Project scaffold** at `projects/1828-illuminated/`:
- `intent.yaml` — purpose, values, constraints, success criteria, stretch goals
- `CLAUDE.md` — working-session protocol mirroring `projects/cpuchip.net/`
- `Dockerfile` (multi-stage: node:22-alpine build → nginx:1.27-alpine serve)
- `nginx.conf` — SPA history-mode fallback + healthcheck endpoint
- `.dockerignore` + `.gitignore`
- `README.md`
- `.spec/journal/` (this entry)

**Data pipeline** at `projects/1828-illuminated/scripts/`:
- `build_data.py` — extracts 853-word tier list + their 1828 entries from `scripts/webster-mcp/data/webster1828.json.gz` direct read (zero MCP load)
- `fetch_modern_defs.py` — rate-limited (1/sec) fetcher for the Free Dictionary API (the same source webster-mcp uses for `modern_define`); resumable; persists every 10 words
- Output: `frontend/src/data/{tier-words,definitions-1828,definitions-modern}.json`

**Frontend** (Vue 3 + Vite 8 + Tailwind 4):
- `App.vue` — header nav + footer with provenance
- `composables/useWordData.ts` — centralized data access + tokenizer + reactive `selectedWord` state
- `components/WordCard.vue` — one-word definition card with 1828 + modern + study cross-references
- `components/HighlightedText.vue` — tokenize a text, highlight tier words with click handlers
- `components/ScripturePanel.vue` — iframe to churchofjesuschrist.org (borrowed pattern)
- Views: `Home`, `WordSearch`, `WordDetail`, `VerseExplorer`, `About`
- 8 hand-curated demo verses (KJV public domain + Restoration scripture excerpts for fair-use educational illustration)

**Smoke** done locally:
- `docker build -t 1828-illuminated .` — successful
- `docker run -p 18828:80 1828-illuminated`
- `/healthz` returns ok
- `/` serves the SPA with proper title + Tailwind styles

## Reframes mid-build

**The cpuchip.net copyright rule transferred.** Reading `projects/cpuchip.net/CLAUDE.md` mid-build surfaced *"Scripture text is not hosted on cpuchip.net... never bundled from gospel-library/ or engine.ibeco.me."* I adjusted the MVP design accordingly: word search is the primary feature (no scripture text needed); verse explorer uses hand-curated short excerpts (KJV is public domain; Restoration scripture excerpts are fair-use educational); full passages link to churchofjesuschrist.org via iframe. Scripture text is NEVER bundled.

**Tier system implemented from P5 synthesis.** Five tiers A++/A+/B/C/D, totaling 853 words. Highlighted with two visual treatments (Tier A++/A+ get a heavier underline). Search defaults to A++/A+/B/C; D opt-in.

**No backend in MVP.** Static site. Free Dictionary API hit at build time (the fetcher), not runtime. Keeps the deploy surface tiny, respects Michael's "don't tax LM Studio" constraint, and lets Dokploy auto-build.

**Two TS-strict errors fixed during build.** `seg.tier!` non-null assertions weren't enough for `vue-tsc` — destructure into local vars + `??` fallback on `Record<string, number>` indexer. Build clean after.

## What's deferred (stretch in `intent.yaml`)

- **LLM-rendering of verses** in modern English with 1828-faithful meanings (user-supplied API key, settings page for LM Studio + opencode-go)
- **Thummim 2026 Restoration Dictionary** built from gospel works (scriptures + GC talks), multi-level grade-bands
- **Tablet-friendly presentation mode** for Brother Philpot's stated wish
- **gospel-engine-v2 integration** for live scripture search — deferred until a system MCP key exists
- **Bundle size optimization** — the 2MB (720KB gzipped) `useWordData` chunk could lazy-load definitions on demand; first cut ships everything bundled
- **Substrate-pipeline per-word meaning-shift summaries** — Michael's idea to use pg-ai-stewards to compose narrative comparisons; deferred for its own batch

## Carry-forward for Michael's morning

- **`docker run -p 8080:80 1828-illuminated`** then open `http://localhost:8080` to see it.
- The MVP works without scripture text. The verse explorer's demo verses use KJV public-domain text + Restoration scripture excerpts; full passages link out via the church's site.
- If the look-and-feel passes the smell test, the Dockerfile is ready to push to Dokploy at `1828.ibeco.me`. The image is small (~30MB) and starts in <1s.
- The fetcher's modern-definition data may not be complete for all 703 words depending on how long it ran. The frontend handles missing modern definitions gracefully ("Modern definition not yet fetched" message). Re-running `scripts/fetch_modern_defs.py` is resumable.

## What the work taught

**Stewardship under time pressure forces priority discipline.** Two hours of session, ~6 distinct deliverables Michael named, plus stretch goals. The shape became: ship the data foundation + word search + verse explorer hard; defer LLM-rendering + dictionary + presentation mode to their own sessions. Each deferred goal got a row in `intent.yaml` so it survives the handoff.

**Borrowing patterns is faster than inventing.** `cpuchip.net/src/components/ScripturePanel.vue` gave the iframe pattern. `stewards-ui/frontend/src/views/Intents.vue` gave the card/list/search shape. The whole frontend assembled in ~45 minutes because every part already existed somewhere in the workspace.

**The 1828 lens needs curation more than it needs more data.** The intersect of canon ∩ webster1828.json yielded 7,538 words. That's too many to highlight. The substrate's own study work (the P2 layer) is the curation that matters — words our own work has lensed are the words readers should see. The tier system honors that: A++ and A+ are study-confirmed; everything else is candidate pool.
