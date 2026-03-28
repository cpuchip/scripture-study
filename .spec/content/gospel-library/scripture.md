# Gospel Library — Scriptures

*Content inventory for model experiment planning.*

---

## Volumes

| Volume | Path | Books | Chapters | Notes |
|--------|------|-------|----------|-------|
| Old Testament | `gospel-library/eng/scriptures/ot/` | 39 (gen, ex, lev, ... mal) | ~929 | Largest volume by chapter count |
| New Testament | `gospel-library/eng/scriptures/nt/` | 27 (matt, mark, ... rev) | ~260 | |
| Book of Mormon | `gospel-library/eng/scriptures/bofm/` | 15 (1-ne, 2-ne, ... moro) | ~239 | Plus: bofm-title.md, introduction.md, explanation.md |
| D&C | `gospel-library/eng/scriptures/dc-testament/dc/` | 1 (sections 1-138) | ~140 | Plus: Official Declarations in `od/`, introduction.md |
| Pearl of Great Price | `gospel-library/eng/scriptures/pgp/` | 5 (moses, abr, js-h, js-m, a-of-f) | ~23 | Smallest volume |

**Estimated total:** ~1,500 chapter files across all volumes

## Study Aids

| Aid | Path | Files | Notes |
|-----|------|-------|-------|
| Topical Guide | `scriptures/tg/` | ~2,500+ | Alphabetical topics, each with definition + scripture refs |
| Bible Dictionary | `scriptures/bd/` | ~1,800 | Alphabetical entries, shorter than TG |
| Guide to the Scriptures | `scriptures/gs/` | Varies | Similar to TG but broader |
| JST | `scriptures/jst/` | Varies | Joseph Smith Translation corrections |
| Harmony of the Gospels | `scriptures/harmony/` | Varies | Cross-referenced gospel accounts |
| Triple Index | `scriptures/triple-index/` | Varies | Combined index |
| Bible Maps/Chronology | `scriptures/bible-maps/`, `bible-chron/` | Varies | Geographic and timeline references |

## Markdown Format

Each scripture chapter file follows this pattern:

```markdown
# Book Chapter

🎧 [Listen to Audio](https://assets.churchofjesuschrist.org/...)

# The First Book of Moses Called Genesis

*Chapter summary text in italics*

**1.** Verse text<sup>[1a](#fn-1a)</sup> with footnote markers...

**2.** Next verse...

## Footnotes
<a id="fn-1a"></a>**1a.** [Cross-reference](../../ot/gen/37.md)
```

### Key characteristics:
- **No YAML frontmatter** — title is in first `#` heading
- **Audio links** — 🎧 CDN links to official Church audio
- **Bold verse numbers** — `**N.**` format, parseable
- **Superscript footnotes** — `<sup>[1a](#fn-1a)</sup>` anchors
- **Cross-references** — relative markdown links between volumes (e.g., `../../ot/gen/37.md`)
- **Chapter summaries** — italicized text before verse 1
- **File size** — typically 5-50KB per chapter

## Digestion Considerations

- **Per-chapter tokens:** ~1,000-10,000 tokens depending on chapter length
- **Full volume fit:** Individual volumes fit easily in 32k context; all scriptures (~1,500 files) would require ~5-15M tokens total
- **Footnotes are gold** — they encode the cross-reference graph that casual reading misses
- **Topical Guide** is a massive structured index — ideal for embedding or RAG retrieval
- **Verse-level chunking** is straightforward thanks to the `**N.**` format
