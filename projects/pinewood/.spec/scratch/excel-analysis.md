# Excel analysis — 2025 Marshfield Pinewood Derby

Source: `projects/pinewood/excel/2025 Marshfield Pinewood derby.xlsx`

## Sheet structure (confirms the spec)

- **Heats** — `Heat | Lane 1 | Lane 2 | Lane 3` (car number per lane)
- **Scores** — same columns + `Score 1 | Score 2 | Score 3` (1=1st, 2=2nd, 3=3rd)
- **Results** — `Car Number | Total Score | Rank` (lowest total wins)

## Schedule properties (verified by script)

For 25 cars / 50 heats:

| Property | Value |
|---|---|
| Cars | 25 |
| Heats | 50 |
| Races per car | 6 |
| Lane distribution per car | exactly 2 / 2 / 2 |
| Opponent pair frequency | every pair races together exactly **once** (150 unique pairs) |
| Gap between same-car heats | min 2, max 17, avg 8.3 |

This is a **Stearns "perfect-N" chart** — the published standard for fair pinewood racing on a 3-lane track. Properties:

1. Each car runs every lane an equal number of times (lane bias eliminated).
2. Each car faces every other car the same number of times (here: once).
3. Cars get rest between heats.

## Implication for the build

We cannot generate this with a naïve algorithm and expect the same fairness. Two viable paths:

1. **Bundle precomputed charts** for N = 4..32 from the public Stearns/Partridge tables (free, well-known, used by every BSA pack). Lookup by N. Fastest, most reliable.
2. **Generator with constraint solver** (backtracking + opponent-balance + lane-balance scoring). More code, but handles arbitrary N and runoffs of unusual sizes (e.g. 4-car race-off).

Recommended: **both**. Use the table when N is in range; fall back to the solver otherwise. Runoffs (typically 2–5 cars) will use the solver since they're small and fast.

## Anomaly noticed

Heat 20, Score 1 = `` ` `` (a stray backtick instead of a digit). Probably a typo from manual Excel entry. Reinforces the value of digit-only input validation in the entry UI.

## Pointer

Findings consumed by → `projects/pinewood/.spec/proposal.md`
