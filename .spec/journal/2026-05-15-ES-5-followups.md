---
date: 2026-05-15
mode: build + verify (Emergency Stop, ES.5)
workstream: WS5 (substrate)
project: pg-ai-stewards
title: "ES.5 follow-ups shipped — fs_search ctx fix, PDF extraction via tabula, consult_subagent granted live"
status: ES.5 s1-s3 SHIPPED + verified. s4 is policy (no build). Soak still PAUSED — resuming is well-supported now.
carry_forward:
  - "RESUME THE SOAK — still paused. ES.3 verified live + ES.5.s1/s2 fixed the two ES.4 pain points (fs_search timeout, PDF garbage) + consult_subagent is granted. Resuming is well-supported. UPDATE stewards.watchman_config SET schedule_enabled=true WHERE id=1."
  - "ES.4 re-run for a clean pipeline-to-verified — the first run failed downstream on a transient provider HTTP 500, not ES.3. Optional; the judge path is already verified."
  - "ES.5.s4 judge tool tiers — policy decided (defer with triggers). Tier 2 trigger: a >1M-token doc. Tier 3 trigger: a re-engaged judge observed hitting external-info walls now that consult is live."
  - "ES.3.s5 model-name normalization — still deferred, optional."
  - "PDF extraction quality: tabula returns text with minor intra-word spacing artifacts (glyph-positioning) — fine for the judge; worth knowing."
links:
  - "../../projects/pg-ai-stewards/.spec/proposals/substrate-ES-emergency-stop.md"
  - "../../projects/pg-ai-stewards/cmd/fs-read-mcp/tools.go"
  - "../../scripts/fetch-md-mcp/tools.go"
  - "../../projects/pg-ai-stewards/extension/es10-grant-consult-subagent.sql"
---

# 2026-05-15 — ES.5 follow-ups shipped

Michael: "build ES.5 — gate your steps with git commits at good points,
test the fixes, I think you can use and test these tools too." And the
note worth keeping: "I've been resting in between — time is measured
differently. I rest and think and read between sessions." The rhythm
between commit-groups is real and intended; it's the council and the
watching, not idle time.

## What shipped

| Commit | Phase | What |
|---|---|---|
| `4a3aa7c` | s1 | fs_search honors ctx + skips excluded dirs |
| `028faf7` | s2 | fetch-md-mcp extracts PDF/Office docs via tabula |
| `3f7203d` | s3 | consult_subagent granted to all pipeline agents |

## s1 — the real cause was not what the ratification assumed

The ES.4 fs_search timeout looked like unbounded traversal. It wasn't —
`fs-read`'s allow-list already excludes gospel-library, and the walk is
bounded to it. The real cause: **fs_search never checked `ctx`**. When
the bridge's 60s deadline fired, the search kept churning a leaked
goroutine; the bridge marked the fs-read session invalidated, which
cascaded to the next call. The fix is graceful ctx-honoring (walk +
file loop abort promptly, return partial results with `Truncated`),
plus the ratified dir-skip (defense-in-depth + speed on the 9P mount)
and a file cap. call-timeout 60→120 is headroom, not the fix. Three
unit tests; live-verified `count:5` clean.

Lesson: the ratified fix named a real improvement (scoping) but not the
root cause. Building it surfaced the truth. Worth doing the trace even
when the ratification sounds settled.

## s2 — pure-Go won, and covered more than asked

Researched the pure-Go PDF landscape: `tsawler/tabula` (MIT, pure Go,
RAG-purpose-built) handles PDF + DOCX/XLSX/PPTX/ODT/HTML/EPUB and emits
`.ToMarkdown()` — matching fetch-md's contract. fetch_url now detects a
non-HTML document (PDF by magic bytes, others by extension) and routes
it through tabula instead of mangling it through HTML readability.
tabula stays pure-Go — its Tesseract/CGo OCR path is `-tags ocr` gated.
Live-verified: a real PDF returned `"markdown":"Dumm y PDF fi le"` —
extracted text, not `%PDF…FlateDecode`. PDFs first, as asked; the other
formats came free from the same library.

## s3 — consult_subagent is live

Granted to all 16 pipeline-stage agent families, derived straight from
the pipelines table. The `tool_permission` resolver confirms `allow`.
Re-engagement is now real: any pipeline agent can send a persistent
sub-agent a new question. This also arms the ES.5.s4 Tier-3 trigger —
if a re-engaged judge is observed hitting "I'd need to look that up"
walls, that's the signal to make the consult-judge tool-capable.

## The arc

ES.1 → ES.3 → ES.4 (verify) → ES.5 (follow-ups). The emergency stop is
fully worked through: the bleed class closed, the judge verified live,
the two pain points the live run surfaced both fixed. ~80 commits
across the arc, zero rollbacks. The substrate is stable, and the
tooling under it is cleaner than before the incident. Soak is paused;
resuming it is the clean next move.
