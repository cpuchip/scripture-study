---
name: wide-search
description: "Cast a broad net across the entire library — semantic search, conference talks, cross-volume connections. Surfaces patterns and relationships that deep reading misses. Use after deep-reading to widen the study, or independently for exploratory research."
user-invokable: true
argument-hint: "[concept or theme to search broadly]"
---

# Wide Search

## Purpose

Wide search is the **wide-angle lens** — scanning across all five standard works, 54 years of conference talks, manuals, and existing studies for connections, patterns, and unexpected resonances. This skill produces the kind of discovery where you search "lifted up" and find it connects Numbers 21, John 3, Helaman 8, 3 Nephi 27, AND a 1978 conference talk you never would have found manually.

This contrasts with the `deep-reading` skill, which telescopes into one passage. Both are valuable. Wide search usually follows deep reading in a phased study.

## When to Use

- Second phase of a phased study (after deep reading surfaced threads to pull)
- The user wants to explore a concept across the whole library
- Looking for conference talk connections to a scriptural theme
- Discovering cross-volume patterns (same idea in OT, NT, BoM, D&C, PoGP)
- Building the "web" of connections that makes a study more than a chapter summary

## Method

### Step 1: Gather Threads from Deep Reading
If this follows a deep-reading phase, start with the "Threads to Pull" from the intermediate findings. These are your search seeds.

If this is standalone, articulate 3-5 search queries before starting:
- What concepts am I looking for?
- What words or phrases might appear?
- What semantic ideas (not just keywords) connect to this topic?

### Step 2: Keyword Search (gospel-mcp)
Use `gospel_search` for specific phrases, names, or terms:
- Exact phrases from scripture ("lifted up," "look and live")
- Key names (Melchizedek, Enoch, Moses)
- Doctrinal terms (priestcraft, atonement, covenant)

**Remember:** These results are *pointers*. Note file paths. Do NOT quote from search results.

### Step 3: Semantic Search (gospel-vec)
Use `search_scriptures` and `search_talks` for conceptual connections:
- Ideas that might be expressed in different words
- Thematic parallels across dispensations
- Conference talks that address the same principle

**Semantic search finds what keyword search misses.** "Serpent as a type of Christ" won't match keyword search, but semantic search will find Numbers 21, John 3, and Helaman 8 together.

### Step 4: Check Existing Studies
Search `study/*.md` for related work. The study corpus is an interconnected web — previous studies may have already traced threads relevant to this topic.

### Step 5: Read and Verify
For every promising discovery, `read_file` the actual source. The `source-verification` skill applies here — search results are pointers, not sources.

### Step 6: Document Connections
Add findings to the working document (if phased study) or compile for synthesis:

```markdown
## Wide Search Findings

### Cross-Volume Pattern: [pattern name]
- OT: [reference + insight]
- NT: [reference + insight]  
- BoM: [reference + insight]
- D&C: [reference + insight]
- PoGP: [reference + insight]

### Conference Talks
- [Speaker, Title, Date](link) — key quote and connection
- [Speaker, Title, Date](link) — key quote and connection

### Unexpected Connections
- [Something you didn't expect to find]

### Connections to Past Studies
- [study/topic.md](link) — how this connects
```

## What Makes Wide Search Different from Regular Discovery

| Regular Discovery (One-Shot) | Wide Search Phase |
|----------------------------|------------------|
| Searches to find what to read | Searches to find *connections* |
| Stops when enough sources found | Keeps going to find the *web* |
| One round of search | Multiple rounds with different query strategies |
| Keyword OR semantic | Both keyword AND semantic, deliberately |
| Results go straight to study | Results go to working notes, then to synthesis |

## The Two-Tool Pattern

**Always use both search tools.** They find different things:

| gospel-mcp (keyword) | gospel-vec (semantic) |
|----------------------|----------------------|
| Finds exact phrases | Finds related concepts |
| Good for: names, specific terms, quoted phrases | Good for: themes, parallels, "verses about X" |
| Misses: paraphrases, different vocabulary | Misses: exact matches, specific rare terms |

A study that only uses keyword search misses entire classes of results. A study that only uses semantic search misses specific critical passages. Use both.

## Integration with Phased Studies

In a phased study plan:
1. **Deep reading** produces focused findings + threads to pull
2. **Wide search** pulls those threads across the full library
3. **Synthesis** weaves deep + wide into the final study document(s)
4. **Becoming** closes with personal application

Wide search feeds synthesis with the breadth and cross-references that make studies like [serpent-and-dragon.md](../../study/serpent-and-dragon.md) comprehensive — 24 source files across all 5 standard works.
