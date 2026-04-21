---
name: byu-citations
description: "Look up which General Conference talks, Journal of Discourses entries, and other sources cite a particular scripture verse using the BYU Scripture Citation Index (scriptures.byu.edu). Use when you want to know who has taught about a verse, how frequently a verse appears in conference, or to discover that a verse has surprisingly little modern commentary."
user-invokable: true
argument-hint: "[scripture reference, e.g. 'D&C 113:6' or 'Alma 32:21']"
---

# BYU Scripture Citation Index

## What It Is

The [BYU Scripture Citation Index](https://scriptures.byu.edu/) tracks which General Conference talks (1942–present), Journal of Discourses entries, Teachings of the Prophet Joseph Smith, and other sources cite each verse of scripture. Our `byu_citations` MCP tool queries this index directly.

## When to Use It

- **During source gathering (Phase 2):** After reading a chapter, use this to discover which conference talks have cited the key verses. This can surface talks you'd never find through keyword search.
- **Confirming silence:** When a verse seems under-discussed, this tool can *authoritatively* confirm how many (or how few) citations exist. "Zero conference citations" is a finding, not an absence.
- **Comparative citation density:** Looking at a cluster of verses (e.g., Isaiah 11:1-10) and checking which ones are heavily cited vs. ignored reveals what the tradition has emphasized and what it hasn't.
- **Finding speakers:** When you want to know who has taught on a specific verse — especially to find modern prophetic commentary.

## How to Use It

### Single Verse Lookup
```
byu_citations("D&C 113:6")
byu_citations("3 Nephi 21:10")
byu_citations("Isaiah 53:5")
```

Accepts standard references, abbreviations, and full book names:
- `"Isa 53:5"` → Isaiah 53:5
- `"Hel 5:12"` → Helaman 5:12
- `"Doctrine and Covenants 93:36"` → D&C 93:36
- `"JS-H 1:19"` → Joseph Smith—History 1:19

### Bulk Lookup
```
byu_citations_bulk("Isaiah 11:1, Isaiah 11:10, D&C 113:1, D&C 113:6")
```

Returns a summary table showing citation counts plus full details for each.

### Book ID List
```
byu_citations_books()
```

Lists all supported books and their BYU internal IDs. Useful for debugging.

## Interpreting Results

Each citation includes:
- **Speaker** — the conference speaker or author
- **Title** — talk or discourse title
- **Reference** — coded as `YEAR-SEASON:PAGE` (e.g., `1989-O:54` = Oct 1989, page 54) or `JD VOL:PAGE` for Journal of Discourses

**Season Codes:**
- `O` = October General Conference
- `A` = April General Conference
- `JD` = Journal of Discourses (19th century)

## Patterns to Watch For

| Pattern | Significance |
|---------|-------------|
| **0 citations** | The verse may be "hiding in plain sight" — prophetically significant but not yet taught. Especially interesting for prophecies. |
| **1-2 citations** | Lightly touched. Check whether the citation is substantive engagement or just a passing reference. |
| **10+ citations** | A well-established teaching verse. The citations show how interpretation has evolved over time. |
| **Only JD citations** | 19th-century engagement only — modern prophets haven't addressed it. This is a significant pattern for prophetic passages. |
| **Recent surge** | Multiple citations in recent years may indicate a verse gaining prophetic emphasis. |

## Integration with Studies

### During Source Gathering
After reading a chapter, query the key verses. The scratch file entry format:

```markdown
### BYU Citation Index — [Verse]
- **Citations found:** [count]
- **Notable speakers:** [list key names and years]
- **Observation:** [what the citation pattern tells you]
```

### In the Study Document
Citation index data belongs in the "Modern Prophets" or "Commentary" sections. The *absence* of citations is as important as their presence:

> "3 Nephi 21:10-11 has exactly one citation in the entire index: John Taylor in the Journal of Discourses (1878). Zero modern conference talks. This is a stunning silence for a verse that contains a direct prophecy of a latter-day servant."

### Cross-Referencing with `gospel_search`
The BYU Citation Index tells you *who cited a verse*. The `gospel_search` semantic mode tells you *who discussed a topic*. They complement each other:
1. Use `byu_citations` to find who cited the specific verse
2. Use `gospel_search` with `mode: "semantic"` to find thematic discussions that may allude to the verse without citing it directly
3. `read_file` the actual talks to verify
