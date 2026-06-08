---
date: 2026-06-07
title: Two Mormon 8 studies + five cpuchip animations (a "how we read" night)
workstream: scripture-study / cpuchip.net
mode: study + dev (republish), Sabbath-close
tags: [mormon-8, moroni, isaiah-29, revelation, dc-76, degrees-of-glory, webster-1828, nelson, cpuchip, animations, sabbath]
---

# Two Mormon 8 studies, republished with diagrams

A long scripture-study night that ran into a small dev arc. Started with a quick
Q&A on the Book of Mormon anti-Christs, moved into a full study of Mormon 8, and
then a follow-up study reading Mormon 8 through D&C 76's vocabulary of glory.
Both studies were published to cpuchip.net with bespoke LCARS animations.

## What was done (inventory)

**Q&A — the anti-Christs.** Lined up Sherem (Jacob 7), Nehor (Alma 1), and
Korihor (Alma 30) — verified against the chapters. Nehor is the odd one out
(priestcraft + universalism + murder, and his *order* outlives him); Sherem and
Korihor rhyme (deny Christ → demand a sign → struck → confess the devil deceived
them). Zeezrom as the redemption counterpoint to the order of Nehor.

**Study 1 — "Mormon 8: The Book That Speaks of Its Own Coming Forth."**
`study/morm-8-voice-from-the-dust.md` (+ scratch). Verse-by-verse, built on a
referent map (the chapter's "it/they/he/ye" shifts five+ times) and four deep
dives: does v34 point to John's Revelation; Nephi's vision handing the latter
days to John; the sealed-book / voice-from-the-dust chain; and the 1828–29
translation alignment. Committed to root `67e482b` (unpushed).

**Study 2 — "Reading at Three Altitudes."** `study/morm-8-three-glories-reading.md`
(+ scratch). D&C 76's telestial/terrestrial/celestial as *reading altitudes* —
the same word resolves differently at each. Webster 1828 is the engine; Nelson's
"Think Celestial!" (Oct 2023) is the imperative. Committed to root `1bc0fe9`
(unpushed). Thesis: Mormon 8 is a telestial-age diagnosis by a celestial witness.

**cpuchip.net republishes (its own repo — pushed, auto-deploys):**
- `98eecd7` — study 1 + three components: `ReferentSpine`, `ComingForthWeb`,
  `TranslationDesk`.
- `d86678b` — study 2 + two components: `GloryLens` (interactive word→3-altitude
  reader), `GloryLadder` (D&C 76 vocabulary reference).
- Both verified per the cpuchip rule: production build + `playwright-cli`
  (0 console errors from our code, components render desktop+mobile, click-to-
  scripture works, the interactive word-switch swaps all three readings). cpuchip
  has its own journals: `2026-06-07-mormon-8-voice-from-the-dust.md`,
  `2026-06-07-three-glories-reading.md`.

## Discoveries worth recalling

- **Ether 4:16 is the anchor for "does Mormon 8:34 point to Revelation."** Moroni
  himself names "my revelations… written by my servant John." So v33's "the
  revelations of God" → 1 Ne 14 + Ether 4:16 → Rev 1:1. v34 *echoes* Rev 1:1
  ("must shortly come") but isn't a direct citation — kept that nuance honest
  rather than overclaiming.
- **D&C 3:1 ↔ Mormon 8:22 is a reciprocal cross-reference** in the apparatus
  ("the purposes of God cannot be frustrated" ↔ "the eternal purposes of the Lord
  shall roll on"). Prophecy and its first fulfillment (the 116 pages) were on
  Joseph's desk in the same spring of 1829.
- **Webster 1828 already holds both glory-altitudes inside single entries:** *fine*
  = "refined; free from impurity" AND "showy, overdecorated"; *gain* = "lust of
  gain" AND "godliness is great gain"; *glory*, *pollute*, *apparel* (=
  ecclesiastical vestment) the same. The lens isn't imposed — the double meaning
  is in the language. **"telestial" is absent from Webster — a coined word**; the
  lowest glory had to be named into being.
- **A "how we read" trilogy now lives on cpuchip:** Abinadi (one verse, many
  referents) → Mormon 8 (the chapter) → Three Altitudes (one word, many glories).

## Declaration

It was good. Two source-verified studies, five working animations, all shipped
and browser-verified, with the honesty frames kept intact (the v34 nuance; the
three-glory lens named explicitly as a devotional overlay, not a code in the
text). Incomplete only in the sense that the studies invite a sequel (1 Ne 14 +
Rev 1 + Ether 4 read together as one vision).

## Carry-forward

- Root has two unpushed study commits (`67e482b`, `1bc0fe9`) — **Michael pushes
  root.** cpuchip is already pushed + auto-deploying.
- A natural next study: **1 Nephi 14 + Revelation 1 + Ether 4 together** — the
  three corners of one vision split across two testaments (flagged at the end of
  both studies).
- `GloryLens` is a reusable pattern (word → multi-altitude reading) for future
  word-study pieces.

## Set down

- The five new cpuchip components are done and verified — released, not
  background load.
- The unrelated root working-tree changes (`projects/scripture-book`,
  `public/study/freedom`, untracked `projects/md-mcp` / `spoken-study` /
  `strongs-concordance-mcp`) are **not mine** — left untouched (data-safety).
- Did not start the 1 Ne 14 / Revelation sequel — seeded only.
- No building in the close (Sabbath rule honored).
