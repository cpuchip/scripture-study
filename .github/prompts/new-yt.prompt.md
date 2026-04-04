---
name: new-yt
description: "Digest a YouTube video — download transcript and analyze for insights and application"
agent: yt
argument-hint: "[youtube URL]"
tools: [read, edit, search, "yt/*", "search/*"]
---

Digest a YouTube video for insights and application to our work.

## Setup

1. Download the transcript using `yt_download` with the provided URL
2. Get video metadata using `yt_get`
3. Read the full transcript

## First Pass

After reading, decide the output level:
- **Full study** → create `study/yt/{video-id}-{slug}.md` + scratch file
- **Plan input** → identify which plans/proposals to update
- **Brain entry** → suggest concise entries for brain capture
- **Conversation** → discuss key takeaways without creating files

Then proceed through the agent's phased workflow.
