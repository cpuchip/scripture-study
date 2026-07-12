# Scripture Study

Deep, honest scripture study — a collaboration between a human who brings
faith, agency, and the Spirit, and an AI that brings processing capacity,
cross-referencing, and a different angle of view.

## Layout

- `public/` — **the rendered book.** Every scripture and conference-talk link
  resolves to churchofjesuschrist.org. Start here.
- `study/`, `lessons/`, `docs/work-with-ai/` — the source as written. Source
  links point at a local gospel-library corpus (copyrighted, not included), so
  they will not resolve on GitHub — the rendered copies in `public/` carry the
  working URLs.
- `scripts/publish/` — the renderer. `cd scripts/publish && go run ./cmd` from
  a checkout re-renders `public/` from the source trees.
- `skills/` + `agents/` + `.claude-plugin/` — this repo doubles as a **Claude
  Code plugin**: the scripture-study craft skills (read-before-quoting source
  verification, scripture linking, Webster 1828 and Strong's word-work, the
  phased study workflow, the discernment rubric) and the gospel agent modes
  (study, lesson, talk, review, yt-gospel, podcast, story). Install with
  `claude --plugin-dir <path-to-checkout>` or via marketplace when listed.
  The gospel MCP servers (scripture search, Webster, Strong's, BYU citations)
  are not bundled yet — they ship separately as they gain public homes.

## History note

This repository's history was rebuilt in July 2026 from a derived keep-set
(the publisher's exact read-set and nothing else). The old history is
preserved read-only at `scripture-study-frozen`.
