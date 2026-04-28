# Pinewood Derby Scoring — Proposal

**Status:** draft (planning)
**Working dir:** `projects/pinewood/`
**Date:** 2026-04-27

## 1. Binding problem

Marshfield's annual pinewood derby is currently scored in Excel — manual data entry into a spreadsheet, no live display, error-prone, no easy way to handle ties or run-offs. The pack needs an **offline, locally-hosted web app** that:

1. Generates a fair race schedule (each car gets equal lane assignments and races every other car).
2. Lets a scorekeeper enter results with a 10-key numpad as fast as the heats run.
3. Drives a live TV display showing current heat, on-deck heat, and standings — updated the moment a result is entered.
4. Computes standings, detects ties, and lets a leader spin up custom run-off races in seconds.
5. Survives a power blip — every entry is durable, and a flat human-readable backup exists.
6. Feels fun for kids — animation, color, the car's name on the screen, not just a number.

## 2. Success criteria

A scorekeeper can run a 25-car / 50-heat derby start to finish without opening a spreadsheet. Specifically:

- Create a race, enter 25 cars (number + optional name) in under 5 minutes.
- Auto-generated schedule satisfies: every car runs 6 times, exactly 2× per lane, no car races back-to-back, opponent pairings balanced.
- TV display updates within 500 ms of a score entry (websocket).
- Score entry: cursor lands in lane-1 field; typing `3` `Enter` `1` `Enter` `2` `Enter` records lane 1 = 3rd, lane 2 = 1st, lane 3 = 2nd, advances to next heat.
- Any prior heat can be re-opened and corrected; standings recompute.
- A run-off can be created in <30 seconds: tied cars are pre-checked; the operator may add/remove cars; "Start run-off" generates a new sub-race with the same fairness guarantees.
- Every state change is appended to a JSONL log on disk (manual-recovery fallback).

## 3. Constraints and boundaries

**In scope:**

- Single-machine web app, runs offline (no internet required during the event).
- Backend: Go (per request), embedded SQLite (`modernc.org/sqlite` — pure Go, no CGO).
- Frontend: Vue 3 + Vite + Tailwind CSS (per request).
- Realtime: native `gorilla/websocket` (or `nhooyr.io/websocket`) — single broadcaster pattern.
- One binary that serves API, websocket, and the built SPA.
- 3-lane track is the only supported track topology (matches the actual hardware).

**Out of scope (v1):**

- Multi-user auth (it's one operator on one laptop).
- Cloud sync, multi-device admin.
- Photo timing / sensor integration (place is entered by a human watching the finish line).
- Cross-event historical analytics.
- Mobile-first design — TV display must look great, scoring screen needs to work on a laptop with a USB numpad.

**Conventions:**

- Match the existing Go project layout used elsewhere in `scripts/` (cmd/<bin>, internal/<pkg>).
- Single executable; SQLite file lives next to it; JSONL audit log next to that.
- All times stored UTC; display in local time.

## 4. Prior art and related work

- **Stearns / Partridge perfect-N charts** — the published industry standard for pinewood schedules; bundled with most commercial pack-management apps (GrandPrix Race Manager, DerbyNet). The 2025 Marshfield Excel chart matches a Stearns chart for N=25 exactly (verified — see [excel-analysis.md](scratch/excel-analysis.md)).
- **DerbyNet** (open source PHP) is the closest existing tool. We're not using it because: PHP/Apache stack vs. Go, the UI is dated, and the pack wants something they own and can tweak.
- Internal Go conventions from `scripts/gospel-engine-v2/`, `scripts/becoming/`: cmd/internal layout, embedded SQLite, simple HTTP+WS, embedded SPA via `embed.FS`.

## 5. Architecture

```
┌──────────────────────────┐         ┌──────────────────────────┐
│  Vue 3 SPA               │◀────────│  Go binary               │
│  - Home / race list      │  HTTP   │  - REST API              │
│  - Race setup (cars)     │◀────────│  - WebSocket hub         │
│  - Schedule view         │   WS    │  - Schedule generator    │
│  - Score entry (numpad)  │         │  - Scoring engine        │
│  - TV display            │         │  - SQLite (modernc)      │
│  - Run-off builder       │         │  - JSONL audit log       │
└──────────────────────────┘         └──────────────────────────┘
                                                │
                                                ▼
                                     ┌────────────────────────┐
                                     │  derby.db   (SQLite)   │
                                     │  derby.log  (JSONL)    │
                                     └────────────────────────┘
```

### 5.1 Data model (SQLite)

```sql
CREATE TABLE race (
  id            INTEGER PRIMARY KEY,
  name          TEXT NOT NULL,
  created_at    TIMESTAMP NOT NULL,
  status        TEXT NOT NULL,   -- 'setup' | 'racing' | 'complete'
  parent_id     INTEGER REFERENCES race(id),  -- null for main, set for run-off
  lane_count    INTEGER NOT NULL DEFAULT 3
);

CREATE TABLE car (
  id        INTEGER PRIMARY KEY,
  race_id   INTEGER NOT NULL REFERENCES race(id) ON DELETE CASCADE,
  number    INTEGER NOT NULL,
  name      TEXT,
  UNIQUE(race_id, number)
);

CREATE TABLE heat (
  id          INTEGER PRIMARY KEY,
  race_id     INTEGER NOT NULL REFERENCES race(id) ON DELETE CASCADE,
  heat_number INTEGER NOT NULL,    -- 1..N within race
  status      TEXT NOT NULL,       -- 'pending' | 'running' | 'complete'
  UNIQUE(race_id, heat_number)
);

CREATE TABLE heat_slot (
  id        INTEGER PRIMARY KEY,
  heat_id   INTEGER NOT NULL REFERENCES heat(id) ON DELETE CASCADE,
  lane      INTEGER NOT NULL,      -- 1..3
  car_id    INTEGER NOT NULL REFERENCES car(id),
  place     INTEGER,               -- 1, 2, or 3 ; null until entered
  recorded_at TIMESTAMP,
  UNIQUE(heat_id, lane)
);
```

Final score per car = `SUM(heat_slot.place)` over that car's slots within the race. Lowest wins.

### 5.2 JSONL audit log

Append-only file. Every mutation writes one line:

```json
{"ts":"2026-05-15T19:32:14Z","event":"score","race_id":1,"heat":12,"lane":2,"car":17,"place":1}
{"ts":"2026-05-15T19:32:18Z","event":"score","race_id":1,"heat":12,"lane":3,"car":22,"place":3}
```

If the SQLite file ever corrupts mid-event, the log can be replayed or hand-tallied.

### 5.3 Schedule generator

Two-tier:

1. **Lookup table** — bundle precomputed Stearns perfect-N charts as JSON (`internal/schedule/charts/N04.json` … `N32.json`) from the public BSA-circulated tables. Fast, perfect fairness.
2. **Solver fallback** — backtracking generator for N outside the table or for run-offs (typically 2–5 cars). Optimizes for: (a) every lane used equally per car, (b) opponent matchups as balanced as possible, (c) min 1 heat rest between same-car appearances. For run-offs, just run "each car races each lane once" → `N` heats with `N` cars total racing N times if N≥3, or a simple round-robin variant for N=2.

Both paths return the same struct: `[]Heat{ Lane: [LaneCount]CarID }`.

**Run-off heat-count rule of thumb** (proposed defaults — operator can override):

| Cars in run-off | Heats |
|---|---|
| 2 | 3 (best of 3) |
| 3 | 3 (each car each lane once) |
| 4 | 4 |
| 5+ | use solver, target 3 races per car |

### 5.4 Realtime

Single in-process WebSocket hub. On every score insert / heat advance / run-off creation, the server broadcasts a small message:

```json
{"type":"score","heat":12,"lane":2,"place":1}
{"type":"standings","top":[{"car":15,"name":"Lightning","total":6,"rank":1}, ...]}
```

The TV display and scoring screen subscribe to the same channel; React-style reconciliation in Vue keeps them in sync.

## 6. Pages / UI

| Page | Purpose | Notes |
|---|---|---|
| `/` Home | List existing races, "New race" button | Simple table. |
| `/race/:id/setup` | Add cars (number + optional name), generate schedule, start race | Numpad-friendly entry; bulk paste also works. |
| `/race/:id/schedule` | Heat chart — who races when | Highlights current + on-deck. Print-friendly view. |
| `/race/:id/score` | Score entry, one heat at a time | **Numpad UX is the headline feature.** See §6.1. |
| `/race/:id/display` | TV view — current heat, on-deck, top-N standings | Big text, animations, kid-friendly. See §6.2. |
| `/race/:id/results` | Final standings, tie detection, run-off builder | Checkbox list defaults to tied cars; operator adjusts. |
| `/race/:id/heat/:n` | Edit any past heat | Reachable from schedule view; recomputes standings on save. |

### 6.1 Score entry — numpad UX

- Always-visible "current heat" card: `Heat 12 — Lane 1: car 11, Lane 2: car 13, Lane 3: car 16`.
- Three big input boxes labeled by lane, each accepting `1`/`2`/`3` only.
- Cursor auto-advances on Enter (or auto on single-keystroke when valid).
- Each entry POSTs immediately and writes to SQLite + JSONL.
- After all 3 lanes filled, heat marks complete and the next heat loads automatically.
- An "Undo last" button rolls back the most recent entry.
- "Jump to heat #" input lets operator return to any heat to fix mistakes.
- Validation: rejects duplicate places within a heat (no two cars get 1st), with a gentle visual cue, not a modal.

### 6.2 TV display

- Top half: current heat — three lanes side by side, each card shows car number, name, and a slot for place once entered. When all three places land, brief celebratory animation (confetti / car emoji zoom).
- Bottom-left: on-deck heat (smaller card).
- Bottom-right: live top-5 leaderboard. Lowest score = top.
- Subtle background animation (slow gradient or moving track lines) so the screen feels alive between heats.
- Big fonts (60+ pt for car numbers), high contrast, kid-readable from across a gym.
- Keep gimmicks short: no animation should block the next heat from displaying.

## 7. Phased delivery

Each phase ships a usable artifact.

### Phase 1 — Skeleton + schedule (1–2 sessions)

- Go module scaffolding, single binary serving an embedded "hello" SPA.
- SQLite schema migrations.
- Stearns chart loader for N=4..32 (JSON files committed).
- Solver fallback for arbitrary N (small cases especially — needed for run-offs).
- CLI command `pinewood schedule --cars 25` that prints the chart to stdout (used to verify correctness against the 2025 Excel).
- **Verification:** generated N=25 chart matches the 2025 Marshfield chart for the lane-balance and opponent-balance properties (not necessarily heat-for-heat identical).

### Phase 2 — Race CRUD + score entry (1–2 sessions)

- Vue 3 + Vite + Tailwind SPA, embedded into the binary.
- Home, setup, schedule, score-entry pages.
- Numpad-driven score entry with live SQLite writes and JSONL append.
- Heat correction flow.
- **Verification:** can manually enter all 50 heats from the 2025 Excel; resulting standings table matches the Excel `Results` tab.

### Phase 3 — Live TV display (1 session)

- WebSocket hub + broadcast on every mutation.
- TV display page with current/on-deck/leaderboard.
- Light animations (Tailwind transitions + a small sprinkle of confetti on heat completion).
- **Verification:** open display on a second window; entering a score on the operator screen updates the display in <500 ms.

### Phase 4 — Run-offs + polish (1 session)

- Tie detection on the results page.
- Run-off builder with default-checked tied cars and operator-editable selection.
- Run-off race links back to its parent race; results page surfaces both.
- Print-friendly schedule view.
- **Verification:** force a tie in test data; confirm run-off creation, scoring, and parent-race tie resolution.

### Phase 5 (stretch) — Kid delight pass

- Sound effects (toggle-able — operator can mute).
- Per-car custom color / emoji.
- Funny "race name" generator for the run-off ("Showdown at Pack 123").
- Keep additions optional and skippable.

## 8. Verification criteria (overall)

- Reproduce 2025 results: enter the 50 heats from `Scores` tab → standings match `Results` tab exactly.
- Schedule for N=25 satisfies: 6 races/car, 2 per lane, all 300 possible pairs covered (or, for charts that don't cover all pairs, opponent counts within ±1 of each other).
- Heat-edit: change one place in heat 7, confirm only affected cars' totals shift.
- Power-cycle test: kill the server mid-event, restart, confirm state intact and websocket clients reconnect cleanly.
- Run-off: create race-off from a 3-way tie, score it, parent race shows resolved 1st/2nd/3rd.

## 9. Costs and risks

- **Schedule licensing.** Stearns charts are widely circulated and used by free tools (DerbyNet ships them). Worth a 5-minute confirmation that the source we use has no restrictive license. If concerns arise, the solver fallback can produce equivalent charts for any N — slower to build, but unencumbered.
- **WebSocket reconnect.** Browsers will drop the WS if the laptop sleeps. Need a reconnect-with-backoff in the client and a "request full state" sync message on reconnect. Easy but must not be skipped.
- **Single-operator UX.** No undo beyond last action in v1. If the operator types `1 1 1` for places, validation blocks the third entry. Acceptable — the audit log is the safety net.
- **Embedded SQLite + writes during animation.** `modernc.org/sqlite` is fine for ~1 write/sec workload. Not a real risk, just noting.
- **Scope creep into "kid delight."** The right amount of animation is "noticeable but never blocks." Time-box Phase 5.

## 10. Creation Cycle alignment

| Step | Answer |
|---|---|
| Intent | Run a fairer, faster, more fun pinewood derby for the pack. |
| Covenant | Offline-first; operator owns the laptop; data stays local; backup log is the contract. |
| Stewardship | This proposal owned by Michael. Build owned by `dev` agent (proposed). |
| Spiritual creation | This document. |
| Line upon line | 5 phases, each independently shippable; could stop after Phase 3 and still beat Excel. |
| Physical creation | `dev` agent executes Phase 1 first, await review before Phase 2. |
| Review | After each phase: verification step listed above. |
| Atonement | JSONL audit log = recovery path; SQLite is replaceable. |
| Sabbath | Pause between phases; revisit Phase 5 only if the kids would notice. |
| Consecration | Open-sourced under `projects/pinewood/`; usable by any pack. |
| Zion | Serves the pack tonight; the schedule generator + JSONL pattern is reusable for any 3-lane bracket scoring need. |

## 11. Recommendation

**Build.** This is a well-bounded, real-deadline tool with clear value. Start at Phase 1 — schedule generator + CLI verification against the 2025 chart. That single phase de-risks the entire project (schedule fairness is the only place this can quietly fail).

Suggested next action: hand to the `dev` agent with instructions to execute Phase 1 only, returning for review before Phase 2.

## 12. Open questions

1. Does the pack want **single elimination after qualifying**, or is the lowest-total-across-all-heats the actual scoring rule? (The 2025 spreadsheet implies the latter, and the spec says the latter — confirming.)
2. For run-offs of 2 cars on a 3-lane track, do we leave one lane empty or run a 2-of-3-format? (Default proposed: leave lane empty, run each car on each of the 2 lanes used, best total wins.)
3. Should the TV display show car names or just numbers when a name is provided? (Default: both — name above number, number bigger.)
4. Anything about the existing Excel UX (tab two manual entry) that the operator actively wants preserved? (Otherwise we replace it entirely.)
