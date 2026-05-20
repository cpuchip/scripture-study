---
title: 1828-illuminated — overnight iteration 2 (three stretch goals)
date: 2026-05-20
status: shipped — five commits, two autonomous-loop iterations
session-type: autonomous (Michael asleep)
parent: 2026-05-20-mvp-built-overnight.md
---

# Overnight iteration 2 — three stretch goals shipped

Iteration 1 (commits `b66444d`, `2cce3dd`) shipped the MVP. The wakeup fired ~1h later; iteration 2 went after the three single-session stretch goals from `intent.yaml`.

## What shipped in iteration 2

**Commit `ed45def` — Stretch goal #1: LLM-rendering**
- `composables/useLLMSettings.ts` — localStorage-backed reactive settings (provider preset, baseUrl, apiKey, model, temperature, maxTokens). Presets for LM Studio (localhost:1234/v1, no key needed) and OpenCode Go.
- `composables/useLLMRender.ts` — builds a prompt with the verse + every tier word's 1828 first sense, POSTs to `{baseUrl}/chat/completions` (OpenAI-compatible), returns rendered text with `[bracket]` markers on substitutions.
- `views/Settings.vue` (route `/settings`) — three preset buttons, full inputs, advanced (temperature, max tokens), explainer card walking through the data flow.
- `views/VerseExplorer.vue` — "Render in modern English" button (gated; routes to /settings if not configured); inline loading/success/error states; auto-clears on verse change.
- App nav: `⚙` link to /settings.
- **Privacy by design.** API key in localStorage only, requests go browser-to-endpoint, site never sees keys or requests.

**Commit `0aef597` — Stretch goal #2: Presentation mode**
- `views/Present.vue` (route `/present?v=<id>`) — fullscreen tablet-friendly layout, 2xl/4xl typography, big chevron nav buttons for thumb reach, keyboard nav (← → space Esc), fullscreen word card overlay on tap.
- Home page: third tile added alongside Word Search + Verse Explorer.
- Verse Explorer: "📖 Present this verse" button in each demo footer with deep link.
- Per Brother Philpot's stated wish — built for use during teaching.

**Commit `0658e5b` — Stretch goal #3: Thummim Dictionary scaffolding**
- `.spec/proposals/thummim-restoration-dictionary.md` — full proposal with vision, architecture (substrate-pipeline-driven generation), word selection (~150-200 v1 corpus), six unratified decisions (D-THM-1..6).
- `data/thummim-seed.json` — three hand-crafted seed entries (intelligence, obtain, charity) demonstrating multi-level voice (elementary / 8th grade / college+).
- `views/Dictionary.vue` (route `/dictionary`) — PREVIEW-badged stub with word selector, reading-level toggle, key passages, GC reinforcement, compare-to-Webster section per entry.

## Project state — five surfaces total

1. `/` Home — tier vocabulary chips + project state + three feature cards
2. `/word` Word Search — 858-word filterable list with tier filter
3. `/word/:word` Word Detail — direct deep-link per word
4. `/verse` Verse Explorer — 8 demos + paste-your-own + click-to-highlight + LLM-render button
5. `/present?v=<id>` Presentation — fullscreen tablet mode (NEW)
6. `/dictionary` Dictionary — Thummim seed entries with grade-level toggle (NEW, preview)
7. `/about` About — methodology + cautions + provenance
8. `/settings` Settings — LLM endpoint config (NEW)

## What's deferred (carry-forward for next session)

- **Substrate pipeline for Thummim generation.** Proposal exists; D-THM-1..6 await ratification; first run is 1-2 sessions of pg-ai-stewards work.
- **gospel-engine-v2 integration** for live scripture search — needs a system MCP key.
- **Bundle optimization** — `useWordData` chunk is still 2.4MB unminified (841KB gzipped). All data eagerly loaded. Lazy-load on demand is the obvious improvement.
- **More demo verses** — only 8 hand-curated. Could grow as we lens more verses in study work.
- **D-1828-1..5 ratifications** (in proposal §VII): domain, pre-render strategy, highlight density, modern source, study-corpus inline integration.

## Honest finds

- **The Free Dictionary API has 20 archaic words it doesn't carry.** Treated as data, not bugs — those gaps signal the word is sufficiently archaic that mainstream dictionaries skip it.
- **Two TS-strict errors during the LLM-rendering build** — fixed with type-narrowing locals + a cast for the demo-verses indexed access. Build clean now.
- **Bundle warning is real but not blocking.** 841KB gzipped is acceptable on home wifi. Slow connections would benefit from lazy chunks; deferred.
- **The wakeup-as-safety-net pattern worked.** Set a 1h wakeup at iteration start; continued working during the session window; the wakeup is the fallback for when session limit cuts in.

## How to verify when Michael wakes up

```sh
cd projects/1828-illuminated
docker build -t 1828-illuminated .
docker run -p 8080:80 1828-illuminated
# → http://localhost:8080
```

Tabs to try in order:
1. `/` — see the three feature cards + tier-A vocabulary chips
2. `/word` — search "intelligence" or "obtain"
3. `/verse` — switch through the 8 demo verses; click highlighted words; try "Present this verse" on D&C 130:18-19
4. `/dictionary` — see the three seed entries with grade-level toggle
5. `/settings` — see the LLM endpoint config (won't actually call without LM Studio running)
6. `/present?v=dc-130-18-19` — fullscreen the intelligence verse for teaching demo

If something looks off, the `.work/` Python scripts + `frontend/src/data/` JSON files are all reproducible from `research/gospel/1828/`.

Wakeup scheduled for ~02:58 in case the session limit cuts in before then.
