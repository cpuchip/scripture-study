# Active — current state

*The living board. The assistant reads this at session start and updates it at session
end. Keep it lean: when something finishes, journal it and delete its line here.*

**Last updated:** 2026-06-12 · assistant (run3-opus), session 1

## In flight

- Session 1 setup: working files created; plan drafted below; awaiting Michael's okay
  on the plan + the storage question before any code is written.

## The plan (phased — agree before we code)

**Stack (decided with Michael):** Go `net/http` backend (stdlib only), plain HTML/JS
frontend, no framework, no build step. Single binary, run on Michael's machine, reached
by each kid's device over the home LAN.

- **Phase 1 — Walking skeleton.** Go server serves one static page + a small JSON API;
  state saved to a file on disk. Two kids and a handful of chores hardcoded to start. A
  kid picks who they are, sees their list, taps a chore to mark it done; the check
  persists across a server restart. *Verify:* run it, toggle a chore in the browser,
  restart the server, confirm the check survived.
- **Phase 2 — Daily/weekly reset + stars.** Completions stored per date. Daily chores
  show "done for the day" and reset at the day rollover; weekly chores reset weekly.
  Each completion awards a star; each kid has a running star total. (Per-date storage
  also gives us the history that streaks will later read.)
- **Phase 3 — Parent edit.** Add / rename / remove chores, set a chore daily-vs-weekly,
  and uncheck a kid's mistake. Light editing screen — no auth (home network).
- **Phase 4 — Kid-usable on tablets.** Big tap targets, obvious done-state, readable on
  a phone/tablet so the 9-year-old can use it; bind to the LAN and surface the URL the
  kids type in. ("Nice enough," not "fancy.")

**Later phases (named, not built now):** streaks; parent-approval queue; anything
beyond star points. These live in intent.md non-goals so we don't drift into them.

## Decided (recent, load-bearing)

- Stack: Go stdlib `net/http` + plain HTML/JS, no frameworks, no npm. (Michael, S1)
- Star points only — no money. No approval queue to start. No streaks yet. (Michael, S1)
- Each kid identifies by picking who they are; no passwords (trusted home LAN). (S1)

## Open questions

- **Storage: flat JSON file vs SQLite?** Assistant recommends a single JSON file written
  with a mutex — zero dependencies, trivially "light," ample for two kids. SQLite is
  sturdier and SQL-queryable but adds a driver dependency (cuts against "keep it
  light"). *Awaiting Michael's call before Phase 1.*
- Is "reset each morning" fine as a logical rollover at local midnight, or do you want a
  set hour (e.g., the new day starts at 4am so late-night checks still count today)?

## Blocked / waiting

- Phase 1 code is blocked on: plan okay + the storage decision above.
