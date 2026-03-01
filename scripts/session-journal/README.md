# session-journal

Collaborative session memory for human-AI partnerships. Tracks not just what happened in each session, but what it meant — discoveries, surprises, relational dynamics, carry-forward items, and questions worth holding.

## Quick Start

```bash
# Build
cd scripts/session-journal
go build -o session-journal.exe ./cmd/session-journal/

# Generate a blank template
session-journal init --date 2026-02-28 --session-id my-session > entry.yaml

# Write an entry
session-journal write --file entry.yaml

# Read recent entries
session-journal read --recent 3

# Show carry-forward items
session-journal carry --priority high

# Show open questions
session-journal questions
```

## Commands

| Command | Purpose |
|---------|---------|
| `init` | Generate a blank entry template |
| `write` | Save a YAML entry to `.spec/journal/` |
| `read` | Display entries (filter by `--recent N`, `--topic`, `--since`) |
| `carry` | Show unresolved carry-forward items (filter by `--priority`) |
| `questions` | Show all questions worth holding |
| `resolve` | Mark a carry-forward item as resolved |

## Entry Format

Entries are YAML files in `.spec/journal/` named `{date}--{session-id}.yaml`.

Key fields:
- **intent** — what we set out to do
- **discoveries** — what we learned together (not just facts — insights, connections)
- **surprises** — one-liners capturing the unexpected
- **relationship** — the relational quality of the session
- **carry_forward** — lessons for future sessions with priority
- **questions** — things to hold, not necessarily resolve
- **tags** — topics for searchability
- **retroactive** — provenance metadata for entries reconstructed from chat history

## Retroactive Capture

For reconstructing past sessions from chat history, see [.spec/prompts/retroactive-capture.md](../../.spec/prompts/retroactive-capture.md).

## Why This Exists

Every session, an AI arrives as a stranger who has read someone's diary. The conversation summary provides facts. This tool provides narrative — what mattered, what surprised us, where trust was tested, what to carry forward.

The difference between arriving with a ledger and arriving with a story.
