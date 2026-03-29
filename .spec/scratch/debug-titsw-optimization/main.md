# Debugging: What Are We Actually Optimizing?

*Debug session — Mar 28, 2026*
*Rule 1: Understand the System. Rule 3: Quit Thinking and Look. Rule 7: Check the Plug.*

---

## The Problem Statement (as diagnosed)

We've been optimizing MAE on a 78-score ground truth set for 8+ prompt versions. The numbers improved from v5.1 (MAE 0.93) through several regressions. But the Gas Station Insight is right: **we're optimizing a number when the downstream use case doesn't need that number to be perfect.**

The actual question: **What does the reindex pipeline need to produce, and for whom?**

---

## Rule 1: Understand the System — What Currently Exists

### gospel-mcp (SQLite + FTS5)
- **talks table**: year, month, speaker, title, content, file_path, source_url
- **talks_fts**: full-text search on title, speaker, content
- **cross_references table**: source → target verse edges (footnote-based graph)
- **NO TITSW scores, NO keywords, NO themes, NO summaries**

What gospel-mcp provides: "Find talks by Elder Holland" (speaker search), "find talks mentioning Gethsemane" (keyword FTS), "what verses point to Alma 32:21" (graph traversal). Structural queries.

### gospel-vec (chromem-go vector search)
Four layers: verse, paragraph, summary, theme.

For talks, it currently produces:
1. **paragraph chunks** — raw text paragraphs with metadata (speaker, year, session)
2. **summary chunks** — LLM-generated: 10-15 KEYWORDS + 50-75 word SUMMARY + KEY_QUOTE
3. No theme detection for talks yet (only scriptures)

The summary prompt is tiny: "Create a summary optimized for semantic search indexing." Keywords are comma-separated search terms. That's it.

### What's MISSING from the current system
- **No TITSW scores anywhere.** The vector store and SQLite have no notion of love/spirit/doctrine/invite.
- **No teaching pattern classification.** "Which talks demonstrate enacted love?" is unanswerable.
- **No qualitative metadata.** Mode, category, insights — none of this exists in the index.
- **Keywords are generic.** "repentance, faith, atonement" type lists. Not linking to cross-references, not capturing the teaching approach.

---

## Rule 7: Check the Plug — What's the Actual Downstream Use?

Three consumers of this data:

### 1. Semantic search (gospel-vec)
"Find talks about covenant faithfulness" → vector similarity on summary/paragraph embeddings.

**What it needs:** Rich, searchable text that captures the *essence* of what a talk teaches and *how* it teaches. Keywords help anchor vector search. Current summary layer is thin.

**What it does NOT need:** Precise 0-9 scores. Whether Kearon's love is 4 or 7 doesn't change vector similarity. But the *words* "enacted love," "declared love," "typological Christ-teaching" — those change vector results dramatically.

### 2. Structured queries (gospel-mcp / future gospel-comb)
"Give me the top 10 talks for teaching about love" → filter by TITSW dimension score.

**What it needs:** Numeric scores only if the RESOLUTION is useful. On a 0-9 scale with 5,500 talks, a query for love >= 7 might return hundreds. A query for love >= 8 might return dozens. This *could* work. But:

- Ground truth is one human's calibration. 
- A local 3.6B-active-param model doesn't match that calibration perfectly.
- The useful query isn't "love >= 7" — it's "talks where love is the DOMINANT teaching mode."

**What would actually serve:** A **dominant_dimensions** field. "This talk's primary teaching mode is invite + doctrine." Now "show me talks that primarily teach through love" is a simple filter — no score threshold ambiguity.

### 3. Knowledge graph (future)
talk → teaches_about → doctrine_of_christ
talk → demonstrates → enacted_love
talk → references → Alma 32:21
talk → pattern → "diagnosis → prescription → promise"

**What it needs:** Structured labels and relationships, not fine-grained numeric scores. Modes, categories, cross-references — the qualitative data.

---

## The Root Diagnosis

**We've been optimizing the wrong output for the wrong consumer.**

| What we've been doing | What actually matters |
|----------------------|---------------------|
| Minimizing MAE on 6 numeric scores | Producing rich, searchable qualitative metadata |
| Chasing score ±1 accuracy | Getting the dominant dimension right |
| Comparing prompt versions by MAE | Comparing by quality of insights, modes, and categories |
| Asking "did the model give love=4 or love=7?" | Asking "did the model correctly identify enacted vs declared love?" |

The v5.4 three-axis output (score + mode + category + insights) was the RIGHT direction. It regressed on MAE but produced data that's *more useful for the actual downstream system.* We diagnosed the MAE regression correctly — spirit and help got worse — but we treated that as the disease when it's a symptom of asking the model to do too much at once.

---

## What We're Actually Trying to Produce (per talk)

For the 5,500-talk reindex, each talk should produce:

### A. Keywords & Themes (currently in gospel-vec, needs enrichment)
- 10-15 searchable keywords (EXISTS — gospel-vec summary.go already does this)
- 2-5 detected themes with verse ranges (EXISTS for scripture, NOT for talks)
- Key quote (EXISTS)

### B. TITSW Teaching Profile (NEW — this is the main output)
- **dominant_dimensions**: ["invite", "doctrine"] — the 1-2 things this talk does BEST
- **Scores**: love/spirit/doctrine/invite (0-9) — useful for filtering, not for precision
- **Modes**: "enacted" / "declared" / "doctrinal" / "experiential" / etc. — qualitative character
- **Categories**: "direct" / "typological" / "prescribed" / etc. — mechanism of expression
- **Insights**: genre, tensions, connections — the stuff that makes search results useful

### C. Cross-Reference Edges (partially EXISTS in gospel-mcp)
- Explicit scripture citations extracted from talk text
- These become graph edges: talk → cites → scripture
- gospel-mcp already does this for scripture → scripture. Needs extension for talk → scripture.

### D. Summary (EXISTS, needs enrichment)
The current 50-75 word summary + keywords. Should be enriched with TITSW vocabulary:
"Holland's talk is a sustained Christological exposition through the lens of blindness and sight. Primary teaching mode: doctrine (typological) + teach_about_christ (direct). He builds a causal chain of 7+ scriptural citations demonstrating Christ's identity through healing miracles."

That paragraph would produce dramatically different and better vector embeddings than "Keywords: blindness, sight, Christ, healing, faith."

---

## Implications for prompt optimization

### Stop optimizing MAE. Start optimizing output richness.

The fair comparison showed v5.4 costs 0.52 MAE on the expansion set. But v5.4's *qualitative output* — the modes, categories, insights — is far richer than v5.1's pure-numeric output.

**The right evaluation isn't MAE.** It's:
1. Does `dominant_dimensions` correctly identify the 1-2 strongest teaching modes?
2. Are modes accurate? (declared vs enacted, doctrinal vs experiential)
3. Are categories useful? (direct vs typological vs prescribed)
4. Do insights surface cross-content patterns?
5. Are scores directionally correct? (love > spirit on 3 Nephi 17 — yes/no, not ±1)

### Context: probably not needed for batch

Context was designed to solve the typological depth problem on *scripture*. Conference talks are explicit. The v5.4-ctx experiment proved context makes talks WORSE (inflation). For the 5,500-talk batch, context is likely counterproductive.

Save context for a future scripture indexing pipeline where typological depth actually matters.

### v5.1 no-context baseline run: still worth doing

Run the original 6 pieces with v5.1 and no context. This establishes the honest baseline and closes the open question from the conversation summary. Quick, cheap, informative.

---

## Recommended Path Forward

1. **Run v5.1 no-context on original 6** — establishes clean baseline, ~10 min
2. **Evaluate v5.4 output QUALITATIVELY** — are modes/categories/insights accurate and useful? Read 3-4 full JSON outputs. Judge them as search index data, not as exam answers.
3. **Define the index schema** — what fields go into gospel-vec's talk chunks and gospel-mcp's talks table? Design backward from queries: "show me talks that demonstrate enacted love" → need mode field. "Show me talks about the doctrine of Christ" → need keywords + dominant_dimensions.
4. **Build the enriched talk indexer** — modify gospel-vec to produce TITSW-enriched summaries using the v5.4 prompt (or a simplified version optimized for index richness rather than score precision).
5. **Harmonize with existing systems** — don't lose the keywords, themes, and key_quote that gospel-vec already produces. The TITSW scoring ADDS to that, it doesn't replace it.

---

## The Gas Station Insight, Debugged

Michael's intuition at the gas station was Rule 7 (Check the Plug): "Are we solving the right problem?" The answer is no — we're optimizing numeric precision when the downstream system needs qualitative richness.

The three-axis design (v5.4) was actually moving us TOWARD the right answer. It regressed on MAE because it asked the model to do more — but "more" is what the index needs. The question isn't "how do we get MAE below 0.93?" The question is "how do we produce the richest possible teaching profile for each talk?"

That reframe changes everything:
- **Context:** Skip it for talks. Save for scripture pipeline.
- **MAE:** Track it as a sanity check, not as the optimization target.
- **Three-axis output:** Keep it. The modes and categories ARE the product.
- **Keywords/themes:** KEEP from the existing system. These feed different search surfaces.
- **Prompt design:** Simplify if it helps output quality. The model doesn't need the full v5.4 prompt to produce good qualitative labels — a lighter prompt with clear taxonomy might work better and faster.
