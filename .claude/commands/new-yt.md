---
description: Digest a YouTube video — download transcript and analyze for insights and application
argument-hint: "[youtube URL]"
---

Digest a YouTube video for insights and application to our work.

**Argument:** `$ARGUMENTS` should be a YouTube URL. If missing, ask for it.

If a more specialized YouTube workflow exists, consider invoking it via `Agent(subagent_type=yt, ...)` with the same URL. Otherwise proceed with the steps below.

## Setup

1. Download the transcript: `mcp__yt__yt_download` with the provided URL
2. Get video metadata: `mcp__yt__yt_get`
3. `Read` the full transcript

## First Pass

After reading, decide the output level:
- **Full study** → create `study/yt/{video-id}-{slug}.md` + scratch file (load `quote-log`)
- **Plan input** → identify which plans/proposals to update
- **Brain entry** → suggest concise entries for brain capture
- **Conversation** → discuss key takeaways without creating files

Then proceed through the workflow at the chosen level. The `source-verification` skill applies to any quotes from the transcript — paste verbatim with the timestamp from the actual transcript file, never paraphrase-in-quotes.
