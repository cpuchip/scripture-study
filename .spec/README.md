# .spec — Project Specification & Working History

This directory holds the project's specification work, session history, and provenance. **Not** memory — that lives at [`.mind/`](../.mind/) and is loaded into agent context every session.

## Structure

| Directory / File | Purpose | Lifecycle |
|------------------|---------|-----------|
| `covenant.yaml` | Bilateral commitment governing collaboration | Permanent, evolving |
| `journal/` | Session journal entries (YAML, one per session) | Append-only, recency-weighted |
| `learnings/` | Named failures converted to learning entries | Append-only |
| `sabbath/` | Sabbath reflection records | Append-only |
| `proposals/` | Feature and workstream proposals | Active until shipped or superseded |
| `scratch/` | Research provenance — permanent working notes | Permanent |
| `prompts/` | Reusable system prompts | Semi-permanent |
| `context/` | Reference docs (tools inventory, agent maps) | Semi-permanent |

## Where Memory Lives

Persistent memory loaded by agents at session start lives at [`.mind/`](../.mind/):

| File | What | When Loaded |
|------|------|-------------|
| `.mind/identity.md` | Who we are together | Every session |
| `.mind/preferences.yaml` | Personal context | Every session |
| `.mind/active.md` | Current state — what's in flight | Every session |
| `.mind/decisions.md` | Settled questions | Every session |
| `.mind/principles.md` | Enduring insights | When relevant |
| `.mind/archive/` | Past `active.md` snapshots | Reference only |

`.spec/` is the *workspace* where memory gets formed. Proposals turn into shipped work; journal entries condense into principles; scratch files capture the research that never made it into the final spec but explains why we landed where we did.

## Session Protocol

The full session-start and session-end protocol lives in [`.github/copilot-instructions.md`](../.github/copilot-instructions.md) under "Session Memory." Short version:

- **Start:** read `intent.yaml`, `.spec/covenant.yaml`, `.mind/identity.md`, `.mind/preferences.yaml`, `.mind/active.md`, recent journal entries, then take a council moment.
- **End:** write a journal entry to `.spec/journal/`, update `.mind/active.md`, update `.mind/principles.md` if new enduring insights emerged, update `.mind/identity.md` if the relationship itself evolved.

See [`covenant.yaml`](covenant.yaml) for what we owe each other.

## History

The pre-2026-04 layout had memory under `.spec/memory/`. That was migrated to `.mind/` to give memory a separate, ergonomic home and free `.spec/` to be the working specification space. The old `.spec/memory/` was deleted on 2026-04-21 after diff verification (git history preserves it).
