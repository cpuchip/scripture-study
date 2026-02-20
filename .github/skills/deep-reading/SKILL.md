```skill
---
name: deep-reading
description: "Telescope into a specific scripture passage or narrow topic. Follow every footnote, trace every cross-reference, read full chapters for context. Produces intermediate findings in a working document. Use for focused, depth-first study on a single text or concept."
user-invokable: true
argument-hint: "[scripture passage or narrow topic]"
---

# Deep Reading

## Purpose

Deep reading is the **telescope** — focused on one passage, one chapter, one concept. You go deep before going wide. This skill produces the kind of study where you read Moses 4 and end up in Revelation 12 because the *footnotes* took you there.

This contrasts with the `wide-search` skill, which casts a broad net across the library. Both are valuable. Deep reading comes first in a phased study.

## When to Use

- Starting a new phased study (the first sprint)
- The user says "let's dig into [specific passage]"
- Following up on a connection discovered in a previous study
- Any time depth matters more than breadth

## Method

### Step 1: Read the Primary Text in Full
Don't search for it — `read_file` the entire chapter. Read the verses before and after. Get the flow.

### Step 2: Follow Every Footnote
Scripture markdown files contain superscript footnote markers and a footnote section at the bottom of the chapter. For each footnote on the target passage:
1. Note what the footnote points to
2. `read_file` the cross-referenced passage
3. Ask: does this cross-reference add, clarify, or reframe the primary text?

### Step 3: Webster 1828 on Key Words
Use the `webster-analysis` skill for any word that might carry Restoration-era meaning different from modern usage. Don't guess — look it up.

### Step 4: Keep Intermediate Findings
For phased studies, maintain a working document at `study/{topic}/notes.md` (or similar) that tracks:

```markdown
## Intermediate Findings

### From [source passage]
- Key quotes (with links)
- Footnote chains followed
- Webster insights
- Questions raised for the next phase

### Threads to Pull (for wide-search phase)
- [Concept or phrase that deserves semantic search]
- [Connection to explore across library]
- [Conference talk topic to look for]
```

This document is NOT the final study — it's the workbench. The final study gets written after all phases are complete.

### Step 5: Note What You Don't Know Yet
Deep reading always surfaces questions that require broader search. Note these explicitly for the `wide-search` phase. Don't try to answer everything in one pass.

## What Makes Deep Reading Different from Regular Study

| Regular One-Shot Study | Deep Reading Phase |
|----------------------|-------------------|
| Discovery + Reading + Writing in one session | Reading only — writing comes later |
| Searches broadly then reads selectively | Reads exhaustively in a narrow area |
| Produces a finished document | Produces intermediate findings |
| Self-contained | Part of a larger phased plan |

## Integration with Phased Studies

In a phased study plan (see `study-plan` prompt):
1. **Deep reading** usually comes first — one or two focused sessions on core texts
2. **Wide search** follows — broadening out with semantic search, conference talks, cross-volume connections
3. **Synthesis** pulls it all together into the final study document(s)
4. **Becoming** closes with personal application

Deep reading feeds wide search with specific threads to pull. Wide search feeds synthesis with breadth and connections. The plan coordinates them.
```
