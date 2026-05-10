---
description: Start a gospel YouTube video evaluation — downloads transcript and scaffolds an evaluation document
argument-hint: "[youtube URL]"
---

Evaluate a YouTube video against the gospel standard.

**Argument:** `$ARGUMENTS` should be a YouTube URL. If missing, ask for it.

If a more specialized gospel-YouTube workflow exists, consider invoking it via `Agent(subagent_type=yt-gospel, ...)` with the same URL. Otherwise proceed with the steps below.

## Setup

1. Read the evaluation template for structure: [docs/yt_evaluation_template.md](docs/yt_evaluation_template.md)
2. Download the transcript: `mcp__yt__yt_download` with the provided URL
3. Get video metadata: `mcp__yt__yt_get`

## Scaffold

Create a new file at `study/yt/{video-id}-{slug}.md` using the template structure:

```markdown
# Video Evaluation: {title}

**Channel:** {channel}
**Date:** {upload date}
**Duration:** {duration}
**URL:** {youtube link}
**Transcript:** yt/{channel}/{id}/transcript.md

---

## Summary

**Speaker's Main Thesis:**

**Key Claims/Doctrines:**
1.
2.
3.

**Scripture/Conference References Made by Speaker:**
-

---

## Discovery
<!-- Search for supporting/contradicting scriptures and talks -->

## Deep Reading Checklist
<!-- Read EVERY source before writing the evaluation -->
- [ ]

---

## Evaluation

### In Line
<!-- Messages that align with scripture and prophetic teaching -->

### Out of Line
<!-- Claims that contradict or distort scriptural truth -->

### Missed the Mark
<!-- Partially true but missing key context -->

### Missed Opportunities
<!-- Where a powerful scripture would have strengthened the message -->

### Overall Assessment
<!-- Is this content spiritually nourishing? -->

---

## Becoming
<!-- What truth can I apply? What warning should I heed? -->
```

## Then Begin

1. **Read the full transcript** — note the thesis, every scripture cited, every source referenced. Load `quote-log` to capture verbatim quotes with timestamps as you go.
2. **Discovery** — use `mcp__gospel-engine-v2__gospel_search` for supporting and contradicting sources
3. **Deep Reading** — `Read` every cited source AND every one found in discovery. Cite-count rule applies.
4. **Evaluate** — write from verified sources only. The `source-verification`, `discernment-rubric`, and `becoming` skills all apply.
