---
name: new-yt-gospel
description: "Start a gospel YouTube video evaluation — downloads transcript and scaffolds an evaluation document"
agent: yt-gospel
argument-hint: "[youtube URL]"
tools: [read, edit, search, "gospel-engine-v2/*", "webster/*", "yt/*", "search/*"]
---

Evaluate a YouTube video against the gospel standard.

## Setup

1. Read the evaluation template for structure: [docs/yt_evaluation_template.md](../../docs/yt_evaluation_template.md)
2. Download the transcript using `yt_download` with the provided URL
3. Get video metadata using `yt_get`

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
<!-- read_file EVERY source before writing the evaluation -->
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

1. **Read the full transcript** — note the thesis, every scripture cited, every source referenced
2. **Discovery** — search for supporting and contradicting sources
3. **Deep Reading** — `read_file` every cited source AND every one found in discovery
4. **Evaluate** — write from verified sources only
