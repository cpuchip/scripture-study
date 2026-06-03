---
name: strongs-analysis
description: "Trace a King James Bible word back to its Hebrew/Greek lemma and sense using Strong's Concordance. Use for KJV word-work — when the English word may flatten the original (one English 'love' over four Greek words; 'soul' over nephesh; 'lovingkindness' over chesed), or when the user asks what an OT/NT word means underneath. The Hebrew/Greek companion to webster-analysis."
user-invokable: true
argument-hint: "[KJV verse reference, or a word/Strong's number to look up]"
---

# Strong's Concordance Analysis (Hebrew/Greek)

## Why Strong's?

The King James Bible is 1611 English laid over Hebrew (Old Testament) and Greek (New Testament). The English often *flattens* the original: one English "love" stands over four distinct Greek words; "soul" carries the Hebrew *nephesh* (a whole living being, not a disembodied part); "lovingkindness" and "mercy" both render *chesed* (covenant loyalty). Strong's Concordance (James Strong, 1890) is the standard bridge — it keys every KJV word to a numbered Hebrew (`H####`) or Greek (`G####`) lemma, usable without knowing the languages.

This is the original-language counterpart to `webster-analysis`. Webster 1828 tells you how an early Saint read the *English* word; Strong's tells you what the *Hebrew or Greek* underneath was. For KJV study, the two together are far richer than either alone.

## When to Use It

- **The canon walks (OT/NT):** the primary use case — tracing a load-bearing word in a chapter note back to its lemma, the way the Book-of-Mormon walk leaned on Webster 1828.
- **Words English flattens:** "love" (G25 *agapáō* / G5368 *philéō*), "soul" (H5315 *nephesh* / G5590 *psychḗ*), "lovingkindness / mercy" (H2617 *chesed*), "world" (G2889 *kósmos* / G165 *aiṓn*).
- **A repeated word that seems to carry weight** — check whether the same English word is actually the same original word across the occurrences (often it isn't).
- **A KJV word that reads oddly** — the original may be more concrete or specific than the English suggests.

## How to Use It (three tools)

### Start at the verse: `strongs_for_verse`
Given a reference, returns the verse text plus its word-by-word tagging — which KJV word carries which Strong's number, with a brief gloss. This is usually the entry point.
```
strongs_for_verse("John 3:16")
  → loved → G25 agapáō (to love);  world → G2889 kósmos;  begotten → G3439 monogenḗs (unique) …
```
The book may be a full name, abbreviation, or alias ("Jn", "Ps", "Song of Solomon", "1 Jn"). **KJV only.**

### Drill into a number: `strongs_define`
Given `H####`/`G####`, returns the lemma + transliteration, Strong's 1890 definition + KJV-usage gloss + derivation, **and** the modern STEPBible (BDB / Abbott-Smith) gloss side by side.
```
strongs_define("H2617")  → chêçêd — kindness; covenant loyalty; "lovingkindness, mercy" …
```

### Reverse lookup: `strongs_search`
Given a KJV English word, returns the Strong's number(s) behind it across the canon (the reverse of `for_verse`).
```
strongs_search("charity")  → G26 agápē, G1654 eleēmosynē …
```

## A worked example

- **1 Corinthians 13** — KJV "charity"
- `strongs_search("charity")` → **G26 *agápē***
- `strongs_define("G26")` → the self-giving, covenant love — the same word behind "God is love" (1 John 4:8) and "God so *loved* the world" (the verb G25 in John 3:16)
- **Insight:** KJV "charity" isn't almsgiving — it's *agápē*, the pure love of Christ (the exact move Moroni 7:47 makes in English). The original ties 1 Cor 13, 1 John 4, and John 3:16 into one thread the English partly hides.

## Patterns to Watch For

| Pattern | Example |
|---------|---------|
| **One English word, several originals** | "love" → G25 *agapáō* vs G5368 *philéō* (John 21 trades them deliberately) |
| **More concrete than the English** | "soul" → *nephesh* / *psychḗ* = the whole living being, not a part |
| **Covenant weight flattened** | "mercy / lovingkindness" → *chesed* (loyal, covenant love) |
| **Tagged function words** | H853 *'eth* (the untranslated direct-object marker) — usually skip in word-work, but it shows the grammar |

## Caveats (important)

- **A gloss is a starting point, not doctrine.** Strong's *glosses*; it does not exegete. A one-word gloss can mislead — read the fuller definition (and the STEPBible layer), and for anything load-bearing, the verse in context.
- **Occasional odd primary senses.** The modern layer picks one primary sense per number; a few read strangely (e.g. a proper-noun sense for a common verb). When a gloss surprises you, that's the signal to read the full `strongs_define` entry, not to trust the one-liner.
- **KJV only.** `strongs_for_verse` does not cover the Restoration text — use `webster_define` + `gospel_get` there.

## Pairing with webster-analysis

For a KJV word, run both: `webster_define` for the 1611/1828 *English* sense the early Saints read, and `strongs_for_verse` / `strongs_define` for the *Hebrew/Greek* underneath. The English shift and the original-language sense often point the same direction — and where they diverge, that gap is usually where the study lives.

## The Model Tool

Like Webster 1828, Strong's returns something discrete and authoritative that you then reason about in context. It enriches deep reading; it doesn't replace it. The insight is the starting point — follow it into the scriptures and see where it leads.
