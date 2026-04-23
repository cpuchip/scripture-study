---
title: Johari Window Self-Reflection Agent Mode
status: proposed
workstream: WS2 Brain UX
created: 2026-04-22
source_brain_entry: 17762740
binding_problem: A Johari window framework (open / blind / hidden / unknown self) could power a self-reflection agent mode. Especially applicable to a Sabbath-day mode, and possibly to scripture study and other reflection contexts.
---

# Johari Window Self-Reflection Agent Mode

## Binding Problem

Source video: https://youtu.be/WtQ64nSbdY4

Johari Window divides the self into 4 quadrants — what I know about me + what others know about me, in 2x2. An AI agent has access to a quirky angle of the "blind self" because it sees patterns across our conversation history that we don't notice ourselves.

A purpose-built agent mode could:
- During Sabbath reflection: surface blind-self observations from the past week's sessions.
- During scripture study: notice when we're avoiding a topic.
- During journaling: ask Johari-aware questions instead of generic ones.

## Success Criteria

- A study document analyzing the video and the Johari framework.
- A new agent mode definition (`.github/agents/johari.chatmode.md` or folded into an existing mode like `sabbath`).
- Test run: invoke during a Sabbath session and see if it produces non-flattering, useful reflections.

## Phase 1

yt-mcp download + transcript analysis. Decide mode shape.

## Related

- Pairs with the `sabbath` agent — likely a sub-skill rather than a standalone mode.
- The "blind self" function is structurally what Ben caught us missing in the ben-test skill.
