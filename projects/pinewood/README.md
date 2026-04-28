# Pinewood Derby Scoring App

Offline scoring system for a 3-lane pinewood derby. One Go binary, embedded Vue 3 SPA, SQLite + JSONL audit log, WebSocket live updates, xlsx import/export.

> Spec: [.spec/proposal.md](.spec/proposal.md) (status: approved)

## Build

```powershell
# 1) Build the SPA into cmd/pinewood/dist/
cd frontend
npm install
npm run build
cd ..

# 2) Build the Go binary (embeds dist/)
go build -o pinewood.exe ./cmd/pinewood
```

## Run

```powershell
.\pinewood.exe serve -addr :8080 -db derby.db -log derby.log
```

Open http://localhost:8080 . Every state-changing API call is appended to `derby.log` (JSONL) so the race can be reconstructed even if the SQLite file is lost.

### CLI: preview a chart

```powershell
.\pinewood.exe schedule -cars 25 -verify
```

For N=25 / 6 runs / 3 lanes the solver currently produces 50 heats, every car runs each lane twice, 145 of 150 unique opponent pairs, gap min/max/avg = 4 / 13 / 8.33 (well within "all heats balanced ±1").

## Workflow

1. **Home (`/`)** — create a new race or import an .xlsx from a previous year.
2. **Registration** — add cars by number (numpad-friendly) + optional name. Shows projected heat count. *Finalize* generates the chart. Late additions after finalize automatically regenerate the schedule for unraced heats and append new heats.
3. **Schedule** — printable heat chart, current heat highlighted, .xlsx export.
4. **Score** — large numpad: enter 1/2/3 for each lane, Enter advances. Auto-saves on every keystroke; auto-jumps to next heat when the current one is complete. Manual jump-to-heat for corrections.
5. **TV Display** — full-screen, large fonts: now-racing heat, on-deck heat, top 5. Confetti / leaderboard updates over WebSocket.
6. **Results** — standings (lowest score wins) with tie detection. Run-off builder pre-selects tied cars; one click creates a child race linked to the parent.

## Data model

`race → car`, `race → heat → heat_slot`. A run-off is just another race row whose `parent_id` points at the main race. Schema: see `internal/db/db.go`.

## What's intentionally minimal

- The schedule generator is a tie-broken solver, not the Stearns lookup tables. It hits the fairness criteria for N=4..32. Lookup tables can be bundled later in `internal/schedule/charts/` if exact Stearns matching matters.
- TV display does not yet have confetti animations — the leaderboard updates live.
- No auth — assumes single-LAN trusted-network use (per spec §2).
