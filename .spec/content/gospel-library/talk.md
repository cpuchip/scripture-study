# Gospel Library — Conference Talks

*Content inventory for model experiment planning.*

---

## Structure

```
gospel-library/eng/general-conference/
├── 1971/
│   ├── 04/  (April)
│   └── 10/  (October)
├── ...
└── 2025/
    ├── 04/  (57 talk files)
    └── 10/
```

**Coverage:** 1971-2025 (55 years)
**Conferences per year:** 2 (April + October)
**Talks per conference:** ~50-60
**Estimated total:** ~5,500+ talk files

## File Naming

Files use numbered prefixes: `11oaks.md`, `12larson.md`, `13holland.md`, etc.
The number indicates session order, the suffix is speaker surname.

## Markdown Format

```markdown
# Talk Title

🎧 [Listen to Audio](https://assets.churchofjesuschrist.org/...)

# Talk Title (repeated)

Presented by [Speaker Name]
[Calling/Role]

[Full transcript text with paragraphs]

<sup>[1](#fn-1)</sup>

## Footnotes
<a id="fn-1"></a>**1.** Footnote text with scripture references
```

### Key characteristics:
- **Speaker and role** — always identified after title
- **Full transcripts** — complete text, not summaries
- **Audio links** — same CDN pattern as scriptures
- **Footnotes** — scripture references, though sparser than scripture files
- **File size** — typically 10-100KB per talk (varies widely by length)
- **No YAML frontmatter** — metadata is in heading/byline

## Digestion Considerations

- **Per-talk tokens:** ~2,000-20,000 tokens (most are 5,000-10,000)
- **Full conference fit:** One conference (~50 talks) = ~250K-500K tokens — fits in nemotron-3-nano or qwen3.5-35b context
- **All talks:** ~5,500 talks = ~30-55M tokens total — too large for any single context window
- **Speaker metadata** is parseable from the byline
- **Temporal patterns:** Useful for tracking doctrinal emphasis changes over time
- **Ideal for summarization:** Each talk is a self-contained document
- **gospel-mcp already indexes these** via FTS5 (keyword search)
- **gospel-vec already indexes these** via embeddings (semantic search)
