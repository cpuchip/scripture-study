---
name: quote-log
description: "Externalize verified quotes and observations to a scratch file as you read, so they survive context compaction. Use during any study session to build a durable working document alongside source reading. This skill exists because files are durable — context is not."
user-invokable: false
---

# Quote Log

## Why This Exists

Context windows compact. When a study requires reading 15-25 source files, early readings get compressed into summaries during compaction — losing exact wording, verse numbers, and the small observations that make studies rich. Then the agent re-reads to verify what it already read, burning context on redundancy instead of depth.

**The fix:** Write verified quotes to a scratch file *immediately after reading each source*. The file persists across compactions. When it's time to write the study, read the scratch file — not 20 chapters again.

## The Scratch File

Create a working file at `study/.scratch/{topic}.md` at the start of the study, right after the outline.

### Format

```markdown
# {Topic} — Source Log

*Working document. Kept as research provenance — traces how observations were reached.*

---

## Outline

1. [Section heading from Phase 1]
2. [Section heading]
3. ...

---

## Verified Quotes

### [Source Path] — [Book Chapter:Verses]
> "Exact quote from read_file" (v. X)

**Observation:** [What stood out, how it connects, what it might mean]
**Connects to:** [Outline section number or other source]

---

### [Next Source]
> "Quote" (v. X)

**Observation:** ...

---

## Webster 1828

### [Word]
- **Definition X:** "[exact definition]"
- **Study relevance:** [how this illuminates the passage]

---

## Conference Talks

### [Speaker, Title, Date] — [file path]
> "Key quote" 

**Observation:** ...

---

## Threads to Pull
- [ ] [Question or connection to explore]
- [ ] [Gap in the outline that needs more sourcing]

## Cross-Study Connections
- [study/related-study.md] — [brief note on connection]
```

## Rules

1. **Write immediately after reading.** Don't batch. Read a chapter → write the quote and observation → move to the next source. This is the whole point.

2. **Copy exact text.** The quote in the scratch file should be copy-paste from what `read_file` returned. No tidying, no paraphrasing. Tidying happens in the draft phase.

3. **Note the outline connection.** Even a brief "→ Section III" helps during drafting. You're building a map, not just a list.

4. **Track what's missing.** The "Threads to Pull" section is where you note gaps. After reading 10 sources, review this section to decide what else to read — don't re-scan the whole outline from memory.

5. **Don't over-collect.** You don't need every verse from every chapter. Grab the quotes that matter for the study's questions. If a whole chapter is relevant, note the range and pull 2-3 key verses.

## When to Read the Scratch File

- **Before writing the first draft** — read the whole scratch file to get the full picture
- **After compaction** — if you sense context has been compressed, read the scratch file to recover exact quotes
- **During gap analysis** — review Threads to Pull and Outline together to find holes

## Lifecycle

- **Created:** Phase 1 (right after the outline)
- **Populated:** Phase 2 (during all source reading)
- **Consumed:** Phase 4 (during drafting)
- **Kept permanently:** Scratch files are research provenance — they trace how observations and arguments were developed. Published alongside studies.

The scratch file is both scaffolding AND audit trail. It shows how the building was constructed.
