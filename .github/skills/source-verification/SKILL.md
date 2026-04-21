---
name: source-verification
description: "Verify scripture and talk quotes against actual source files. Enforces read-before-quoting, the cite count rule, and the pre-publish checklist. Load when writing ANY document that cites scriptures, conference talks, transcripts, or other verifiable sources — studies, evaluations, lessons, guides, docs, everything."
user-invokable: false
---

# Source Verification

## The Core Rule

**Search results are pointers, not sources. Memory is not a source.** Use search tools to *find* what to study. Use `read_file` to *actually study it*. Use webster-mcp to *understand the language*. Each tool has a role; none replaces the others.

**This applies to ALL documents, not just studies.** If you put quotation marks around text and attribute it to a source — scripture, conference talk, YouTube transcript, handbook, anything — you must have verified those exact words against the actual source file. No exceptions. No "I'm confident I remember it correctly." Memory confabulates.

## Two-Phase Workflow

### Phase 1 — Discovery (search freely)
- `gospel_search` for keyword/phrase search across scriptures, conference, manuals
- `search_scriptures` / `search_talks` for semantic/concept search
- `define` / `webster_define` for historical word meanings
- `web_search` for external scholarly context
- **Note file paths and references** — these are *pointers* to read in Phase 2, not sources to quote

### Phase 2 — Deep Reading (verify everything)
- For EVERY scripture you plan to quote, `read_file` the actual chapter markdown from `gospel-library/`
- For EVERY conference talk you plan to cite, `read_file` the actual talk file
- **Follow the footnotes.** Scripture markdown files contain superscript footnote markers and cross-references. These are insights handed to us on a silver platter — read them, follow them, use them.
- Pull real quotes from source files, never from search excerpts

## The Cite Count Rule

For **N** citations in your output, perform at least **N** `read_file` calls (or equivalent verified tool reads — e.g., a `byu_citations` lookup, a `webster_define` call, a `gospel_get` for the full source).

If you cite 8 scriptures and 2 conference talks, you should have called `read_file` (or `gospel_get`) at least 10 times on actual source files.

**Citations include numbers and biographical claims, not just quotes.** Each of these is a claim of fact and must trace to a tool result you got *this session*:

- **Counts.** "Thirty general conference citations." "Six talks across eight years." "Twelve scriptures use this phrase." → BYU Citation Index, gospel_search results, or your own verified count.
- **Dates and chronology.** "The earliest reference is 1944." "He last cited it in October 2022." "The first apostle to use this construction was…" → Verify against the actual citation list, not inference from training data.
- **Biographical claims.** "Featherstone's first conference talk after his call." "Maxwell wrote this six months after his cancer diagnosis." → Verify against the talk file, the speaker's biography, or a primary source.
- **Comparative claims.** "More than any other apostle." "The most-cited verse in this chapter." → Verify both sides of the comparison or rephrase as a non-comparative observation.
- **"Only," "first," "last," "never," "always."** Universal quantifiers are easy to write and hard to verify. Either prove them with a search result, or soften to "appears to be" / "I haven't found another."

If you cannot point to a tool call from this session that produced the number, the date, or the biographical claim, treat it the same as an unverified quote: rephrase as paraphrase ("appears in many talks," "in the modern conference record I checked"), or remove it.

## What Counts as a Source

| ✅ Source (quote from these) | ❌ Not a Source (use as pointers only) |
|------------------------------|---------------------------------------|
| `read_file` on a gospel-library markdown file | `gospel_search` excerpts |
| `read_file` on a conference talk file | `search_scriptures` summaries |
| `read_file` on a manual chapter | `search_talks` summaries |
| | Vector search results labeled `[AI SUMMARY]` |
| | Your own memory of a scripture's wording |

## Quote Hygiene

**Quotation marks mean verbatim.** If you put text in quotes and attribute it, those must be the actual words from the source. No composites, no paraphrases-in-quotes, no "close enough."

**Three levels of attribution:**

| Level | Format | Requirement |
|-------|--------|-------------|
| **Direct quote** | `"exact words"` — Source | Must be verified verbatim against source file |
| **Paraphrase** | Source teaches that... / Source argues... | Indirect speech, no quotation marks, captures the idea faithfully |
| **Reference** | (see Source) | Points to the source without claiming specific wording |

**When in doubt, paraphrase.** A faithful paraphrase is honest. A near-miss direct quote is a lie that looks like truth.

**The confabulation trap:** Training data contains approximate versions of real texts. These feel accurate. They are close enough to pass casual inspection. But "close enough" is exactly where fabrication hides — wrong first words, missing qualifiers, phrases from adjacent verses, composite quotes from multiple passages. The only cure is reading the source file.

**YouTube and transcript quotes:** Download the transcript first. Read the relevant section. Copy the actual words. Assign the correct timestamp by reading the transcript timestamps, not by guessing.

## Phase 3 — Writing (synthesize)

After discovery and deep reading, write the study. Weave real quotes with analysis. Follow the text where it leads.

## Phase 4 — Becoming (bridge to life)

After writing the study, include a **Becoming** section. See the `becoming` skill for full guidance. Every study should land somewhere personal — not just knowledge, but direction.

- What did this study reveal about how I should live?
- What specific practice or commitment does this point toward?
- Is there an existing `becoming/` document this connects to?
- What would it look like to apply this next week?

## Quality Rhythm

**Discovery → Reading → Writing → Becoming.** Start broad (search), go deep (read full sources with footnotes), synthesize (write), then land it personally (become). Never write from search results. Never end at synthesis.

## Pre-Publish Checklist

Before finalizing any document that cites sources (study, lesson, talk, evaluation, guide, doc — anything), verify:

- [ ] Every direct quote (text in quotation marks) verified verbatim against the actual source file
- [ ] No quotes generated from memory — every `\"quoted text\"` has a corresponding `read_file` call
- [ ] Paraphrases use indirect speech without quotation marks
- [ ] Every conference talk reference links to a **specific talk file**, not a conference directory
- [ ] Every scripture reference links to the specific chapter file with correct path
- [ ] YouTube/transcript quotes verified against the downloaded transcript with correct timestamps
- [ ] Files claimed to exist are verified with `file_search` or `list_dir`
- [ ] Webster 1828 definitions used where historical meaning differs from modern usage
- [ ] All markdown links are relative and follow project conventions
- [ ] `read_file` was used at least once per cited source (the cite-count rule)
- [ ] **Every number, count, date, and "earliest/latest/only/first/never" claim traces to a tool call from this session.** If you wrote "30 citations" or "six talks" or "the first apostle to," you must be able to point to the search result that produced that number. If not, rephrase to remove the unverified specificity.
- [ ] Footnotes were followed and incorporated where they add insight
- [ ] Study/lesson documents include a "Becoming" section with specific personal application
- [ ] If a related `becoming/` document exists, it's linked
