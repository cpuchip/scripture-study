---
title: 1828-illuminated scriptures — public-facing reading frame tool
date: 2026-05-20
status: idea recorded — overnight word-list groundwork in progress
workstream: WS7 (teaching / public-facing tooling)
parent: substrate-yt-transcripts arc (the Morgan Philpot evaluation surfaced this)
purpose: >
  Render the standard works with the 1828 Webster meaning frame already
  woven in — so readers SEE the meaning shift instead of having to ask.
  Make the lens that drove our internal study work (obtain vs receive,
  intelligence as substance, charity as gift, etc.) accessible as a
  public reading surface.
---

# 1828-illuminated scriptures

## I. Origin

Brother Morgan Philpot's Marshfield MO talk (2026-05-17, video `9UTrPgjLW7g`) mentioned an "1828 dictionary presentation tool." Michael's note (2026-05-20): a lot of Philpot's talk resonated with substrate work that uses Webster 1828 as a reading lens. He asked: could we combine gospel-engine-v2 and webster-mcp into a public-facing tool that renders scripture with the 1828 frame applied?

This idea touches three existing pieces:
- **webster-mcp** — the substrate's 1828 + modern dictionary tool (already indexed; data at `scripts/webster-mcp/data/webster1828.json.gz`)
- **gospel-engine-v2** — the substrate's search backend, deployed at `engine.ibeco.me` (uses local LM Studio for embeddings)
- **Existing study corpus** — many of our scripture studies invoke Webster 1828 to recover meaning that modern English flattens

The principle "Webster 1828 as Model Tool" is already in `.mind/principles.md`: *"The ideal pattern: provides a specific, authoritative result (historical definition). AI reasons about it in context. Output is genuinely enhanced. Tool doesn't replace reading — it complements it. 'Obtain' vs 'receive' in D&C 84 was the breakthrough example."*

## II. Vision

Not just "click a word, see the 1828 definition" — that's webster-mcp in a webpage. The interesting move is **rendering scripture with the 1828 frame already woven in** so readers see meaning shifts before they have to ask.

Three layers, minimum to ambitious:

### Layer 1 — Illuminated verse
Reader browses or enters a verse. The verse renders with words-that-have-drifted underlined or color-coded. Hover/click reveals the 1828 definition inline. The curation question is the interesting one — which words to flag. Only words whose 1828 meaning differs in a way that changes the reading. *Conversation, prevent, let, charity, soul, intelligence, obtain, receive, suffer, peculiar* are canonical examples.

### Layer 2 — Word study mode
Search a single word, see every scripture occurrence rendered with the 1828 definition pinned. The substrate's AGE graph already has 602 Scripture vertices; this is one cypher query away from being a real cross-reference engine rather than just a definition lookup.

### Layer 3 — Verse-in-context
Pick a verse, get the verse, surrounding context, the 1828 lens for key words, AND the Topical Guide / footnote cross-references the Church study apparatus already provides. 1828 sits alongside existing study aids rather than replacing them.

## III. Deployment direction (per Michael, 2026-05-20)

- **Domain:** `1828.ibeco.me` — borrow the becoming/ibeco.me infrastructure already deployed at NOCIX rather than standing up new hosting.
- **Self-contained 1828 + embedded gospel-engine** is the architectural ideal eventually, but the easier first cut is to deploy as a sibling on ibeco.me and use gospel-engine-v2 as a backend.
- **Long-term:** combine gospel-engine-v2 + webster-mcp surfaces into a multi-referent tool layer.
- **Constraint:** gospel-engine-v2 uses local LM Studio for embeddings. Don't over-tax that pipeline with new tools. 1828 lookups are static (no embedding); only scripture search would hit LM Studio.

## IV. Honest cautions

- **1828 isn't always deeper.** Some words just mean their modern thing. A naive "everything has hidden meaning" frame would mislead. Curation matters more than data.
- **Risk of decoder-ring posture.** Treating 1828 as a hidden-meaning unlock rather than as "the language frame the Restoration scripture was given in." The tool should illuminate, not encode mysteries.
- **Good-faith reads still differ.** "Proper meaning frame" (Michael's phrase) doesn't end the conversation — it opens it. Tool should support that, not foreclose.
- **TTS / auto-caption confidence.** Earlier substrate evaluation work caught what looked like a citation error in Philpot's talk (2 Thess 2:1 vs 2:11) — but the transcript came from Google's auto-captions, which routinely mishear ordinal numbers. The 1828 tool should treat the canon text (which we have verbatim in `gospel-library/`) as ground truth and never inject transcription-based heuristics into the canon-reading surface.

## V. Overnight groundwork (2026-05-19/20, while Michael sleeps)

Michael's directive: produce the word-list of things to highlight as the groundwork for the tool. Four steps:

1. **Search `./study/**`** for Webster references. Compose a markdown table of every word our internal studies have already lensed via 1828, with links back to the studies. Flag cases where 1828 reinforces the Restoration sense AND cases where they blatantly differ based on local grammar (the intelligence study has at least one such case).
2. Use that as the **confirmed-by-existing-studies highlight list**.
3. Extract **unique words across the standard works** (BoM, D&C, PGP, KJV NT + OT) via tokenization.
4. Intersect with `webster1828.json` (local read, no MCP calls — keeps LM Studio untouched per the constraint). Flag candidates where the 1828 entry has linguistic shift signal.

Outputs land in `./research/gospel/1828/` with provenance.

## VI. Carry-forward (after the word list)

- **MVP frontend** — pick 5-10 verses from existing study work (the D&C 84 "obtain vs receive" breakthrough; D&C 130:18-19 on intelligence; 1 Cor 13 on charity; etc.) and hand-render them with the 1828 lens to validate the UX before scaling.
- **Static pre-render vs live lookup** — static for canonical text (fast, indexable, no per-page-view cost); live for free-text or user-entered passages.
- **Topical-Guide + 1828 layered display** — both at once, so the user sees existing study apparatus alongside the new lens.
- **Multi-referent tool surface** (the long-term combine) — `webster_in_context(verse, word)` that returns 1828 def + modern def + appearances in study corpus + AGE Scripture node neighbors. Not built tonight.

## VII. Decision points for future ratification

- **D-1828-1**: domain — 1828.ibeco.me (borrow infra) vs 1828.cpuchip.net (Michael's personal site) vs standalone. *Initial lean per Michael: 1828.ibeco.me for first deploy.*
- **D-1828-2**: pre-render strategy — full canon static at build time, or per-verse on-demand?
- **D-1828-3**: highlight density — only meaning-drifted words, or any word with a 1828 entry (let user hover)?
- **D-1828-4**: source of "modern" definition for comparison — webster-mcp's `modern_define` tool, or a separate corpus?
- **D-1828-5**: integration with study corpus — show "this word lensed in study X" links inline when present?

Not ratifying tonight. Capturing for the next session that picks this up.
