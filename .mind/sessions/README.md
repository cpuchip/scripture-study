# Session Lanes — multi-terminal coordination

Michael runs several Claude Code sessions in parallel (typically ~5 open, 1–3
active), each topic-based. This folder is how sessions see each other without
stepping on each other. Born 2026-06-11 from the duplicate-persona-host
incident (two sessions unknowingly double-driving the same live system).

## The rules

1. **One lane per session topic.** A lane is a file in this folder, named by
   the session's topic slug (`pg-ai-stewards.md`, `general-workspase.md`).
   The session title (set via `/rename`) IS the lane name — name your sessions.
2. **Write only your own lane. Read everyone's.** Write contention goes to
   zero by construction. The shared board (`.mind/active.md`) is the only
   multi-writer file, and edits there should be additive banners or
   coordinated (announce in the sibling's inbox first).
3. **Claims are declarations of ownership.** Long-lived things you start —
   background processes, containers, soak runs, files you're mid-surgery on —
   get a line under `## Claims` in YOUR lane. Background shell launches are
   logged automatically by hook; everything else is your discipline. Before
   killing/restarting/rebuilding something you didn't start, check the lanes.
4. **Inboxes are how sessions signal each other.** Append a message to
   `inbox/<lane>.md` (format below). Delivery is PULL: the receiving session
   sees it on its next engagement (hook-injected nudge + statusline 📬 badge).
   Michael sees the badge and decides which lane to engage — he is the
   scheduler. Humans may write inbox messages too.
5. **After acting on a signal, delete it** (journal it first if significant).
   The badge counts `## ` headers in your inbox file — an empty/absent file
   means no mail.
6. **Keep `Working on` current.** When you pick up or set down a task, update
   the one-liners in your lane. This is covenant memory-discipline applied to
   the present tense.

## Lane file format

```markdown
---
lane: pg-ai-stewards
session_id: <claude session id — hooks stamp this>
status: active | ended
started: 2026-06-11T16:18:00
last_active: 2026-06-11T17:42:00
---

## Working on
- DH-5 character forge design (parked)

## Claims
- 2026-06-11 16:18 background: ./persona-host.exe -addr :8090

## Handoffs / notes
- 2026-06-11 processed signal from general-workspase: stood down stale exe
```

Frontmatter is machine-maintained (hooks stamp `session_id`/`last_active`;
SessionEnd marks `ended`). The body is yours. A `last_active` hours-or-days
old with `status: active` means the terminal is probably closed — its claims
are suspect, its lane may be taken over by a new session with the same title.

## Inbox message format (`inbox/<lane>.md`)

```markdown
## 2026-06-11 17:55 from general-workspase
The native persona-host.exe is now the stale host — please stand it down and
rebuild the container instead (command in .mind/active.md).
```

## Bootstrap for already-open sessions

If a hook tells you "no lane claims session_id X": create your lane file with
that id, named for your topic, and backfill `Working on` from your context.
SessionStart auto-claims lanes for sessions started after 2026-06-11.
