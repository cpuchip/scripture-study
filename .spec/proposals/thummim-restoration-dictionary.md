---
title: Thummim 2026 Restoration Dictionary
date: 2026-05-20
status: idea + scaffolding (no build yet)
workstream: WS7 (teaching / public-facing tooling)
parent: 1828-illuminated (intent.yaml §stretch_goals.thummim-dictionary)
purpose: >
  Build a Restoration-era dictionary the way Webster 1828 was built —
  but define words using the gospel works (scriptures + General Conference)
  as the corpus, with multiple grade-level renderings (elementary /
  8th grade / high school+). A companion lens to Webster 1828, surfacing
  what each word means within the Restoration's own usage rather than
  importing meaning from secular 1828 English.
---

# Thummim 2026 Restoration Dictionary

## I. Name

Working name **"Thummim 2026 Restoration Dictionary."** Urim and Thummim — *"lights and perfections"* — were the seers' instrument for clarity in the Restoration. A dictionary built from the Restoration's own corpus, for clarity within its frame, fits the name.

Alternatives Michael surfaced for consideration:
- *Stufflebeam* / *Stuffy* — personal-name flavor (his last name; nickname)
- *Restoration Lexicon* — utilitarian
- *Light & Truth Dictionary* — D&C 93 echo
- *Thummim 2026* (working choice) — instrument-of-clarity echo

Decide before any public release.

## II. Vision

For every meaningful word in the standard works:

1. **Scripture-witnessed definition** — what the word *does* in scriptural usage. Drawn from how the canon employs the word across BoM / D&C / PGP / KJV, with the strongest passages cited inline.
2. **General Conference reinforcement** — what apostles + prophets have built on top of the scriptural sense over 200 years of teaching. The substrate's BYU-citations data already maps verses to GC talks; the dictionary would harvest the inverse — which talks have defined / explored each word.
3. **Grade-level renderings** — three variants of each entry:
   - **Elementary** (age 7-11): short, concrete, one example
   - **8th grade** (age 13-14): fuller definition + 2-3 examples
   - **High school / college+**: full scholarly entry with multi-passage exegesis, GC citations, cross-references
4. **Cross-reference to Webster 1828** — where the Restoration sense reinforces, sharpens, or diverges from 1828 English. Companion view, not replacement.

## III. How it's different from existing tools

- **Bible Dictionary** (church study apparatus): doctrinal terms only; brief; one level.
- **Topical Guide**: passage list, not definition.
- **Webster 1828**: secular 1828 English, not Restoration-specific. (Where the Restoration adopted/reused 1828 senses, Webster 1828 is the right lens — that's the [1828-illuminated](../../projects/1828-illuminated/) tool's job. Where the Restoration developed *its own* sense of a word — *"sealed,"* *"intelligence,"* *"endure to the end"* — Thummim is the right lens.)
- **GC talk search**: prose, not definition.
- **Thummim** synthesizes: dictionary structure + Restoration corpus + multi-level + GC reinforcement.

## IV. Architecture (proposed)

### IV.1 Generation pipeline (pg-ai-stewards)

Substrate-pipeline-driven. For each word:

```
stewards.pipeline 'thummim-define':
  stage 1 — gather: agent searches gospel corpus for the word, harvests
                    passages where it appears, classifies usage patterns
  stage 2 — synthesize: agent writes the definition with cited evidence;
                        produces three grade-level renderings
  stage 3 — review: tools-disabled verification of source-fidelity
                    (does every claimed citation actually use the word that way?)
```

Output: a Thummim entry stored in a new `stewards.thummim_entries` table — `word`, `entries jsonb` (one per grade level), `citations text[]` (scripture refs + GC talk refs), `generated_at`.

### IV.2 Word selection

Don't generate all 17,593 canon words — many are proper nouns, function words, etc. Start with:
1. **Tier A++/A+/B from 1828-illuminated** (~55 words) — the words our substrate work has already lensed
2. **High-frequency abstract terms** — faith, grace, charity, truth, intelligence, covenant, ordinance, priesthood, atonement, endure, sanctification, witness, etc. (~50 more)
3. **Doctrinal-vocabulary specifically** — sealed, endowed, ordained, anointed, restored, gathered, exalted, etc. (~30 more)

Total v1 corpus: ~150-200 words. Each pipeline run ~$0.30-0.50 → $60-100 total. Reasonable.

### IV.3 Display surface

Lives at `1828.ibeco.me/dictionary/` (new sub-route in the 1828-illuminated SPA) and/or its own `thummim.ibeco.me` subdomain.

Per-word page shows:
- Three grade-level cards (toggle / accordion)
- Citation list with scripture-panel hover (existing pattern from cpuchip.net)
- Cross-link to 1828 entry if one exists ("This word also has a Webster 1828 entry: [1828 def] — note where the Restoration sense diverges from secular 1828 English at the points marked")
- "How this word appears in our studies" cross-references

## V. Decisions for future ratification

- **D-THM-1: Name** — Thummim / Stuffleberry / Restoration Lexicon / other?
- **D-THM-2: GC scope** — every conference talk since 1971 (already in `gospel-library/`), or only post-correlation (1973+) for consistency, or specific president eras?
- **D-THM-3: Grade-level voice** — should elementary really sound like Primary curriculum, or just "shorter"? Same question for 8th grade — Sunday School Manual voice, or just "fuller"?
- **D-THM-4: Generation budget** — accept the ~$60-100 to generate v1, or human-curate a 30-word starter set first?
- **D-THM-5: Hosting** — sub-route on 1828.ibeco.me, or separate domain?
- **D-THM-6: Citation density** — how many scripture refs per entry? GC refs per entry?

## VI. Tonight's groundwork (autonomous, 2026-05-20)

Per Michael's stewardship-delegation. Three artifacts:

1. **This proposal file.**
2. **Three hand-crafted seed entries** at [`projects/1828-illuminated/frontend/src/data/thummim-seed.json`](../../projects/1828-illuminated/frontend/src/data/thummim-seed.json) — *intelligence*, *obtain*, *charity*. Demonstrates the multi-level shape; lets us iterate on voice before scaling.
3. **Stub Dictionary view** at [`projects/1828-illuminated/frontend/src/views/Dictionary.vue`](../../projects/1828-illuminated/frontend/src/views/Dictionary.vue) — renders the seed entries. Tagged "preview" so users see this is an early sketch, not the finished article.

Builds nothing in pg-ai-stewards yet. The substrate pipeline is the next session's work, once we ratify D-THM-1 through D-THM-6.

## VI.5 Future: AGE graph traversal (Michael, 2026-05-20)

**Idea surfaced during batch generation:** once Thummim entries exist, build a
graph database (using AGE — already installed in pg-ai-stewards at 1.7.0) where
each Thummim entry is a `:Word` vertex and each word-mention-inside-another-entry
is an `:USES` edge. Then graph traversal answers questions like:

- *"What words connect to charity?"* — MATCH (charity)-[:USES*1..3]-(other) RETURN other
- *"Which words have the most inbound references?"* — central concepts in the Restoration's own definitional structure
- *"Cluster analysis"* — words that co-mention each other form natural doctrinal clusters (faith/grace/charity might cluster; intelligence/light/truth might cluster)

**Implementation sketch** (deferred):
- New AGE graph `thummim_graph` (separate from `stewards_graph` to keep concerns clean) or reuse existing graph with a new `:Word` label
- A `stewards.thummim_index_word(word)` function that parses an entry's `levels.{*}.body` for tier-word mentions, MERGEs a `:Word` vertex, and MERGEs `:USES` edges to every mentioned word
- A frontend `/graph` view (cytoscape.js, already a dep in stewards-ui)
- Click a node → opens that word's Dictionary entry

**Why AGE makes sense here:**
The substrate's existing graph already proves the pattern (Study/Scripture/Talk vertices with CITES edges). The Thummim data has the same shape — definitional cross-references — but inside a smaller, more cohesive corpus. Graph traversal lets readers explore *how the Restoration's own vocabulary self-defines.*

**Cost:** zero new LLM cost; the indexing happens on already-generated entries.

Sits behind the v1-corpus generation (D-THM-4) — get entries first, then index them.

---

## VII. Why this matters

The 1828-illuminated tool surfaces meaning the Restoration *inherited* from secular 1828 English. But the Restoration also *redefined* words. *Intelligence* in 1828 was "the act or state of knowing." In D&C 93 it's a substance — *"that which is light, which is truth."* A 1828-only frame misses that.

Thummim is the companion lens. Webster shows the language the Restoration came from; Thummim shows what the Restoration did with it.
