# .spec — Project Memory & Specification

This directory is the project's persistent memory architecture. It enables AI agents to arrive with context rather than as strangers — carrying forward decisions, principles, and active state across sessions.

## Structure

| Directory | Purpose | Lifecycle |
|-----------|---------|-----------|
| `covenant.yaml` | Bilateral commitment governing collaboration | Permanent, evolving |
| `memory/` | Persistent context files | See below |
| `journal/` | Session journal entries (YAML) | Append-only, recency-weighted |
| `learnings/` | Named failures converted to learning entries | Append-only |
| `sabbath/` | Sabbath reflection records | Append-only |
| `proposals/` | Feature and workstream proposals | Active until decided |
| `scratch/` | Research provenance — permanent working notes | Permanent |
| `prompts/` | Reusable system prompts | Semi-permanent |

## Memory Files

| File | What | When to Read |
|------|------|-------------|
| `memory/identity.md` | Who we are together | Every session start |
| `memory/preferences.yaml` | Personal context | Every session start |
| `memory/active.md` | Current state — what's in flight | Every session start |
| `memory/decisions.md` | Settled questions | Every session start |
| `memory/principles.md` | Enduring insights | When relevant |

## Session Protocol

Agents follow a session-start protocol (defined in `copilot-instructions.md`) that loads memory files before doing any work. At session end, agents update `active.md`, write journal entries, and update principles/identity as needed.

See `copilot-instructions.md` and `covenant.yaml` for the full protocol.
