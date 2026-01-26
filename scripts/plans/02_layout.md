# Layout Decision: Scripture Corpus Organization

> **Status:** Draft
> **Created:** 2026-01-25
> **Scope:** Decide how to store scripture content for AI-assisted study and reduce file sprawl.

---

## Current State

We have two parallel corpora:

1. **Legacy corpus** (./scriptures):
   - One file per subbook (e.g., 1 Nephi, Alma, Isaiah)
   - Fewer files, larger context per file

2. **Gospel Library corpus** (./gospel-library/eng/scriptures):
   - One file per chapter (e.g., 1 Nephi 1, 1 Nephi 2, ...)
   - Many more files (~12k total across all content types)
   - Includes additional study helps (Topical Guide, Bible Dictionary, Guide to the Scriptures, Triple Index)

We also have General Conference talks as single files, which is desirable.

---

## Pros & Cons

### One File per Chapter (Gospel Library style)

**Pros**
- Precise linking to a single chapter
- Smaller files are easier for LLMs to ingest without truncation
- Flexible for UI navigation, search results, and file-level notes
- Matches Gospel Library URIs, making links straightforward

**Cons**
- Explosion in file count
- Harder to browse in file explorer
- Study helps are extremely granular (one word per file)
- Increased maintenance and reconversion time

### One File per Subbook (Legacy ./scriptures)

**Pros**
- Far fewer files
- Easy to keep entire subbook in a single context
- Better for long-range study and cross-chapter context
- Minimal file sprawl

**Cons**
- Links to a single chapter require anchors
- Larger files may exceed AI context limits
- Less aligned with Gospel Library URIs

---

## Decision Direction

We should **standardize on Gospel Library corpus only**, and **remove ./scriptures** to eliminate duplication and confusion.

Reasoning:
- Gospel Library corpus is the most accurate source and supports footnotes and structured links.
- Keeping both creates ambiguous references in AI-assisted study.
- We already invested in link localization and crawler fixes.

---

## Mitigations for File Sprawl

### 1. Consolidate Study Helps
These are the largest offenders in file count:
- Topical Guide (TG)
- Bible Dictionary (BD)
- Guide to the Scriptures (GS)
- Triple Index

**Plan:**
- Merge each into a **single consolidated markdown file** (or one per letter).
- Update cross-links to point into those consolidated files using headings/anchors.
- Re-run reconvert after link rules are updated.

### 2. Optional: Consolidate Scriptures by Subbook
We could optionally generate a secondary "combined" view (non-source-of-truth):
- Keep chapter-level files as canonical
- Generate a "combined" file per subbook for reading
- Link chapter references to anchors inside the combined file

This keeps accuracy and linking while providing a readable format.

---

## Proposed Actions

1. **Remove ./scriptures** and standardize on ./gospel-library/eng/scriptures
2. **Consolidate study helps** into fewer files
3. **Update link localization rules** to point into consolidated files
4. **Re-run --reconvert** once link changes are complete

---

## Risks

- Consolidation will require new link resolution rules
- Existing links will break until reconvert is done
- Some cross-links may need custom anchor mapping

---

## Decision Record

**Decision:** Use Gospel Library corpus as canonical, remove ./scriptures
**Rationale:** Accuracy, footnotes, consistent URIs, single source of truth
**Next:** Plan and implement study-help consolidation
