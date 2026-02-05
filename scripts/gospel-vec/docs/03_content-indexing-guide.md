# Gospel-Vec Content Indexing Guide

Notes on indexing different content types from the gospel-library.

---

## Content Types Overview

The `/gospel-library/eng/` directory contains several distinct content types, each with different structures and indexing considerations:

| Content Type | Path Pattern | Structure | Indexing Priority |
|-------------|--------------|-----------|-------------------|
| **Standard Works** | `/scriptures/{volume}/{book}/` | Numbered chapters with verses | ✅ Indexed (BoM, PGP, D&C) |
| **General Conference** | `/general-conference/{year}/{month}/` | Individual talk files | 🔜 High Priority |
| **Manuals** | `/manual/{manual-name}/` | Lesson or chapter files | 🔜 High Priority |
| **Study Aids** | `/scriptures/tg/`, `/scriptures/bd/`, etc. | Dictionary/encyclopedia entries | 🔜 Medium Priority |
| **Magazines** | `/liahona/{year}/{month}/` | Article files | Later |
| **Videos/Broadcasts** | `/video/`, `/broadcasts/` | Transcripts | Later |

---

## Standard Works (Scriptures)

**Status**: ✅ Book of Mormon, Pearl of Great Price, D&C indexed. 🔜 Old Testament, New Testament needed.

### Structure
```
/scriptures/
├── bofm/
│   ├── 1-ne/
│   │   ├── 1.md
│   │   ├── 2.md
│   │   └── ...
│   ├── 2-ne/
│   └── ...
├── ot/
│   ├── gen/
│   ├── ex/
│   └── ...
├── nt/
│   ├── matt/
│   ├── mark/
│   └── ...
├── dc-testament/dc/
│   ├── 1.md
│   ├── 2.md
│   └── ...
└── pgp/
    ├── moses/
    ├── abr/
    └── ...
```

### File Format
- One markdown file per chapter
- Verses numbered with `**X.**` format
- Footnotes at end with `<a id="fn-X">` anchors
- Audio link at top

### Chunking Strategy
- **Verse layer**: Individual verses
- **Paragraph layer**: Groups of ~5 verses (or natural paragraph breaks)
- **Summary layer**: AI-generated chapter summary
- **Theme layer**: AI-identified thematic sections within chapters

### Notes
- Current indexing works well for scriptures
- Book abbreviations: `bofm`, `ot`, `nt`, `dc-testament/dc`, `pgp`
- Chapter numbers without leading zeros: `1.md`, `10.md`, `138.md`

---

## General Conference Talks

**Status**: 🔜 Not yet indexed. HIGH PRIORITY.

### Structure
```
/general-conference/
├── 2024/
│   ├── 04/
│   │   ├── 11oaks.md
│   │   ├── 12larson.md
│   │   ├── 57nelson.md
│   │   └── ...
│   └── 10/
│       └── ...
├── 2023/
└── ...  (back to 1971)
```

### File Naming Convention
- Format: `{session-order}{speaker-name}.md`
- Example: `57nelson.md` = Session 5, talk 7, by President Nelson
- Sessions: 1-6 typically (Saturday AM/PM/Priesthood, Sunday AM/PM/Women's)

### File Format
```markdown
# Talk Title

🎧 [Listen to Audio](url)

# Talk Title

By Speaker Name

Position/Calling

Quote or summary line.

Body text with paragraphs...

## Section Headings (sometimes)

More text...

---

## Footnotes

<a id="fn-1">**1.** Reference text.
```

### Metadata Available
- Talk title
- Speaker name
- Speaker position
- Conference date (from path)
- Session number (from filename)

### Chunking Strategy Recommendations
- **Paragraph layer**: Natural paragraphs
- **Section layer**: H2 headings when present
- **Summary layer**: AI-generated talk summary with key themes
- **Quote layer** (new?): Extract memorable quotes/teachings

### Challenges
- Some talks have no H2 headings
- Speaker positions change over time
- Footnotes reference scriptures (potential for cross-referencing)

### Indexing Notes
- Need to extract speaker name and position as metadata
- Conference date extraction from path: `/general-conference/{year}/{month}/`
- Consider indexing by speaker for "what has President Nelson taught about X"

---

## Come, Follow Me Manuals

**Status**: 🔜 Not yet indexed. HIGH PRIORITY.

### Structure
```
/manual/
├── come-follow-me-for-home-and-church-old-testament-2026/
│   ├── 001-conversion.md
│   ├── 002-using.md
│   ├── 01.md  (week 1 lesson)
│   ├── 02.md  (week 2 lesson)
│   └── ...
├── teaching-in-the-saviors-way-2022/
│   └── ...
└── ...
```

### File Format
```markdown
# Date Range. "Lesson Title": Scripture References

🎧 [Listen to Audio](url)

Metadata lines...

![image]

Date: "Lesson Title"

# Scripture References

Introduction paragraph explaining context...

## Ideas for Learning at Home and at Church

[Scripture Reference]

### Study Prompt Heading (e.g., "As a child of God, I have divine destiny")

Teaching content with questions, scripture references...

See also [links to additional resources]
```

### Metadata Available
- Lesson date range
- Lesson title
- Scripture references being studied
- Year/curriculum type from folder name

### Chunking Strategy Recommendations
- **Lesson layer**: Entire lesson as one document (for overview)
- **Prompt layer**: Each study prompt/section (H3 headings)
- **Summary layer**: AI-generated lesson summary
- **Connection layer**: Link to the scriptures being studied

### Challenges
- Lessons reference specific weeks—temporal relevance
- Mix of study prompts, quotes, and supplementary content
- Links to videos and external resources

### Indexing Notes
- Extract scripture references from title for cross-referencing
- Consider temporal metadata (which week of year)
- Useful for "what does Come Follow Me say about [topic]"

---

## Bible Dictionary & Study Aids

**Status**: 🔜 Not yet indexed. MEDIUM PRIORITY.

### Structure
```
/scriptures/
├── tg/     # Topical Guide (A-Z topics)
├── bd/     # Bible Dictionary (encyclopedia entries)
├── gs/     # Guide to the Scriptures
└── jst/    # Joseph Smith Translation excerpts
```

### Bible Dictionary Format
```markdown
# Entry Title

# **Entry Title**

*Pronunciation or meaning.* Definition and explanation text
with inline scripture references like [Gen. 14:18-20](url).
See also *[Related Entry](url); [Another Entry](url).*
```

### Characteristics
- Varies widely in length (one sentence to multiple paragraphs)
- Inline scripture references
- Cross-references to other entries
- No section headings typically

### Chunking Strategy Recommendations
- **Entry layer**: Entire entry as one document (most entries are short)
- **Summary layer**: Skip for short entries; generate for longer ones
- No paragraph/verse layering needed

### Challenges
- Very short entries (may not embed well)
- Dense with cross-references
- Some entries are outdated (from 1979 LDS Bible)

### Indexing Notes
- Good for definition queries: "what is a Levite"
- Cross-reference extraction could enhance scripture searches
- Consider combining with Topical Guide for topic searches

---

## Topical Guide

### Format
```markdown
# Topic Name

# **Topic Name**

See also [Related Topic](url)

Scripture reference [Book X:Y](url)
Another reference [Book X:Y](url)
...
```

### Characteristics
- Lists of scripture references by topic
- Minimal explanatory text
- Cross-references to related topics

### Chunking Strategy
- Not a good candidate for embedding (just lists of references)
- Better used as a cross-reference index
- Could enhance search results: "Topic Guide says these scriptures are related to X"

---

## Teaching in the Savior's Way

**Status**: 🔜 Not yet indexed. HIGH PRIORITY for lesson preparation.

### Characteristics
- Contains principles for teaching
- Structured around teaching concepts
- Includes examples and application ideas

### Likely useful for queries like:
- "How do I teach by the Spirit?"
- "How do I ask good questions?"
- "Teaching principles for [topic]"

---

## Teachings of Presidents of the Church

**Status**: 🔜 Not yet indexed. LOW-MEDIUM PRIORITY.

### Structure
```
/manual/
├── teachings-joseph-smith/
├── teachings-brigham-young/
├── teachings-john-taylor/
...
└── teachings-russell-m-nelson/   # If available
```

### Characteristics
- Organized by topic/chapter
- Contains quotes from specific prophet
- Historical context notes

### Useful for:
- "What did Joseph Smith teach about [topic]?"
- Historical prophetic teachings

---

## Implementation Priorities

### Phase 1 (Immediate)
1. **Old Testament & New Testament** — Complete standard works coverage
2. **General Conference (recent 10 years)** — Most cited modern teachings

### Phase 2 (Near-term)
3. **Come Follow Me (current year)** — Directly supports weekly study
4. **Teaching in the Savior's Way** — Supports lesson preparation workflow
5. **Bible Dictionary** — Definition/reference queries

### Phase 3 (Future)
6. **General Conference (full archive)** — Historical depth
7. **Teachings of Presidents** — Historical prophetic teachings
8. **Topical Guide** — Cross-reference enhancement (may not need embedding)
9. **Liahona Magazine** — Current articles
10. **Videos/Broadcasts** — If transcripts available

---

## Potential Tool Enhancements

### Metadata Extraction
- **Speaker/Author**: Extract from conference talks, manual credits
- **Date**: Extract from path for temporal context
- **Scripture References**: Extract from inline links for cross-referencing
- **Topic Tags**: Auto-generate or extract from titles

### Search Filters
- Filter by content type (scriptures, talks, manuals)
- Filter by date range
- Filter by speaker (for conference talks)
- Filter by curriculum (for manuals)

### Cross-Reference Integration
- When returning a scripture, also return related TG/BD entries
- When returning a talk, link to referenced scriptures
- Build a citation graph over time

---

## Technical Considerations

### get_chapter Tool Fix
The `get_chapter` tool failed for book="dc", chapter=84. Need to:
1. Document accepted book identifiers
2. Add error message showing valid book names
3. Consider accepting multiple formats (e.g., "dc" OR "D&C" OR "dc-testament/dc")

### Layer Configuration
Different content types may need different layer configurations:

| Content Type | verse | paragraph | summary | theme |
|-------------|-------|-----------|---------|-------|
| Scriptures | ✅ | ✅ | ✅ | ✅ |
| Conference | ❌ | ✅ | ✅ | ✅ |
| Manuals | ❌ | ✅ | ✅ | (prompts) |
| Dictionary | ❌ | ❌ | (long only) | ❌ |

### Embedding Model Considerations
- Short BD entries may not embed distinctively
- Consider minimum text length threshold
- May need to combine very short entries with related context

---

*Last updated: During priesthood study session*
