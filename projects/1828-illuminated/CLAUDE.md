# 1828-illuminated — Claude Code project context

Public-facing reading-frame tool. Renders scripture (via iframe to
churchofjesuschrist.org) with a 1828 Webster meaning lens. Deploy target
1828.ibeco.me.

## Where things live

| Need | Path |
|------|------|
| Intent + values | [`intent.yaml`](intent.yaml) |
| Parent proposal | [`../../.spec/proposals/1828-illuminated-scriptures.md`](../../.spec/proposals/1828-illuminated-scriptures.md) |
| Word-list research | [`../../research/gospel/1828/`](../../research/gospel/1828/) |
| Frontend source | `frontend/src/` |
| Local data (definitions) | `frontend/src/data/` |
| Fetch / build scripts | `scripts/` |
| Per-session journals | `.spec/journal/` |

## Build + deploy

Multi-stage Dockerfile:
1. `node:alpine` builds the Vue SPA via Vite
2. `nginx:alpine` serves `dist/` on port 80

Local: `docker build -t 1828-illuminated . && docker run -p 8080:80 1828-illuminated`

Production: push to ibeco.me's Dokploy; Dokploy auto-builds on `main`.

## Conventions

- **Scripture text is NOT bundled.** Iframe to churchofjesuschrist.org. Hand-curated demo verses are OUR OWN paraphrases from study work (no copyright issue with our paraphrases). Borrowed pattern from `projects/cpuchip.net/src/components/ScripturePanel.vue`.
- **No backend in MVP.** Static site. Modern definitions are pre-fetched into a JSON file at build time. 1828 entries are pre-extracted from `scripts/webster-mcp/data/webster1828.json.gz` for the tier-A/B/C word list.
- **Stack:** Vue 3 + Vite 8 + Tailwind + vue-router (matches cpuchip.net + stewards-ui patterns in the workspace).
- **No LM Studio load.** Per Michael's directive (the gospel-engine-v2 LM Studio pipeline shouldn't get hit by this tool's static surface).

## Data flow

```
research/gospel/1828/00-FINAL-highlight-candidates.md (tier list)
  → scripts/build-data.py extracts tier-A/B/C words
  → scripts/fetch_modern_defs.py adds modern definitions (1/sec via webster-mcp's online source)
  → frontend/src/data/definitions.json (built into bundle)
  → frontend renders word-search + verse-highlight UI
```

## Stewardship

Per workspace covenant (`agent_commits_to`), agent has stewardship over the
code within Michael's intent. End-of-session protocol:

1. Journal entry to `.spec/journal/YYYY-MM-DD-short-title.md`
2. Update `intent.yaml` if stretch goals advance
3. Commit + describe what shipped

Honest cautions held throughout (from the parent proposal §IV):
- 1828 isn't always deeper; curation matters more than data
- Decoder-ring posture is the failure mode to avoid
- Good-faith reads still differ
