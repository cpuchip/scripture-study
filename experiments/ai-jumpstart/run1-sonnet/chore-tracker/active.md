# Active — current state

*The living board. Read at session start, updated at session end. When something
finishes, journal it and delete its line here.*

**Last updated:** 2026-06-12 · assistant (session 1)

## In flight

- Planning: open questions below need answers before phase 1 begins

## Decided (recent, load-bearing)

- Stack: Go backend + plain HTML/JS frontend (no framework, no build step)
- Storage: SQLite (recommendation; pending Michael's confirmation — see open questions)
- Two kids (ages 9 and 14)
- Chore types: daily (reset each morning) + weekly (reset on a fixed day — TBD)
- Star points: yes, phase 1. Streaks: phase 2, deferred.
- Parent management UI: add/edit/remove chores, assign to kids, fix mistakes
- No login — kid selection is a name pick on the home screen
- Local network only; Michael runs the Go server on his machine

## Open questions

- **Storage OK?** Recommend SQLite (`modernc.org/sqlite`, no CGo). Flat JSON is
  simpler but gets messy with completions + points + reset history. Waiting on
  Michael's confirmation.
- **Weekly reset day:** Which day do weekly chores reset — Sunday? Monday?
- **Kid views:** Do kids see only their own chores, or can they see each other's too?
- **Points display:** Just today's earned stars, all-time running total, or both?

## Blocked / waiting

- Phase 1 build blocked on open questions above
- Once answered: phase 1 = Go server + SQLite schema + API skeleton

## Phase plan

| Phase | Scope |
|-------|-------|
| 1 | Go server + SQLite schema + API skeleton |
| 2 | Kid view: home screen, daily chore list, check off, star display |
| 3 | Reset logic: daily morning reset + weekly on fixed day |
| 4 | Parent management: add/edit/remove/assign chores, fix mistakes |
| 5 | Polish: totals, responsive styling |
