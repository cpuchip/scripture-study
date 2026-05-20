# 1828 Illuminated

Public-facing reading-frame tool. Renders scripture (via iframe to churchofjesuschrist.org) with a 1828 Webster meaning lens. Deploy target: **1828.ibeco.me**.

See [intent.yaml](intent.yaml) for purpose + values + constraints. See [CLAUDE.md](CLAUDE.md) for working-session protocol.

## Quick start

```sh
# 1. Build the data bundle (extracts 1828 entries + tier list)
python3 scripts/build_data.py

# 2. Fetch modern definitions (~12 min at 1 word/sec — resumable)
python3 scripts/fetch_modern_defs.py

# 3. Build + run the docker image locally
docker build -t 1828-illuminated .
docker run -p 8080:80 1828-illuminated
# → http://localhost:8080
```

## What it does

- **Word search** — look up any of ~700 curated words; see 1828 and modern definitions side by side; see which substrate studies have lensed each word.
- **Verse explorer** — pick a demo verse (KJV + Restoration scripture excerpts) or paste your own; highlighted words have a meaning-shift between 1828 and modern English.
- **Click any highlighted word** → opens the definition card with 1828, modern, and study cross-references.

## What it does not (yet)

- **Render scripture text directly** — copyright; always defer to the Church's site for full passages.
- **Call any LLM at runtime** — gospel-engine-v2 is not touched; LM Studio is not loaded.
- **Compose verses in modern English with 1828-faithful meanings** — stretch goal; requires a user-supplied LLM API key + settings page.
- **Carry a Restoration-era dictionary** — stretch goal; "Thummim 2026 Restoration Dictionary" working name.

## Data pipeline

```
research/gospel/1828/00-FINAL-highlight-candidates.md (P5 synthesis)
  → scripts/build_data.py
    → frontend/src/data/tier-words.json          (the highlight list)
    → frontend/src/data/definitions-1828.json    (every tier word's 1828 entries)
    → scripts/fetch-wordlist.txt                 (input for the fetcher)
  → scripts/fetch_modern_defs.py
    → frontend/src/data/definitions-modern.json  (added by the fetcher)
  → frontend/src/data/demo-verses.json           (hand-curated)

→ Vite bundle → static dist/ → nginx
```

## Stack

- Vue 3 + Vite 8 + TypeScript
- Tailwind CSS 4
- vue-router 4

## Project tree

```
projects/1828-illuminated/
├── README.md           (this file)
├── intent.yaml         (purpose + values + constraints)
├── CLAUDE.md           (working-session protocol)
├── Dockerfile          (multi-stage: node build → nginx serve)
├── nginx.conf          (SPA-friendly with history-mode fallback)
├── .dockerignore
├── scripts/
│   ├── build_data.py      (tier-words + 1828 defs extraction)
│   ├── fetch_modern_defs.py (modern dictionary fetcher, 1/sec)
│   └── fetch-wordlist.txt
├── frontend/
│   ├── package.json
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── index.html
│   └── src/
│       ├── main.ts
│       ├── App.vue
│       ├── router.ts
│       ├── style.css
│       ├── composables/useWordData.ts
│       ├── components/{WordCard,HighlightedText,ScripturePanel}.vue
│       ├── views/{Home,WordSearch,WordDetail,VerseExplorer,About}.vue
│       └── data/
│           ├── tier-words.json
│           ├── definitions-1828.json
│           ├── definitions-modern.json
│           └── demo-verses.json
└── .spec/journal/      (per-session journals)
```

## Carry-forward

- Stretch goals from [`intent.yaml`](intent.yaml): LLM-rendering settings, Thummim 2026 Restoration Dictionary, presentation mode for tablet.
- The research/gospel/1828/ tier list will benefit from a substrate-pipeline-generated per-word "meaning-shift summary" — currently the user has to read both definitions and infer.
- Optional gospel-engine-v2 integration (search across the corpus) is deferred until a system MCP key is created.
