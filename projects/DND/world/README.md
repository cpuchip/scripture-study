# Sword Coast Campaign World

The campaign is set along the **Sword Coast** of Faerûn, focusing on five cities connected by four major roads. All documents use YAML frontmatter for machine-parsable data (distances, danger levels, services) and prose markdown for narrative content.

## Cities

| City | Population | Danger | Region | File |
|------|-----------|--------|--------|------|
| [Waterdeep](cities/waterdeep.md) | 130,000 | 2 | North | `waterdeep.md` |
| [Neverwinter](cities/neverwinter.md) | 35,000 | 3 | North | `neverwinter.md` |
| [Daggerford](cities/daggerford.md) | 8,000 | 2 | Central | `daggerford.md` |
| [Baldur's Gate](cities/baldurs-gate.md) | 75,000 | 3 | South | `baldurs-gate.md` |
| [Candlekeep](cities/candlekeep.md) | 1,000 | 1 | South | `candlekeep.md` |

## Roads

| Road | Connects | Distance | Days | Danger | File |
|------|----------|----------|------|--------|------|
| [Coast Way](roads/coast-way.md) | Neverwinter ↔ Waterdeep | 100 mi | 3 | 2 | `coast-way.md` |
| [Trade Way](roads/trade-way.md) | Waterdeep ↔ Daggerford ↔ Baldur's Gate | 200 mi | 6 | 3 | `trade-way.md` |
| [Way of the Lion](roads/way-of-the-lion.md) | Daggerford ↔ Candlekeep | 220 mi | 7 | 4 | `way-of-the-lion.md` |
| [Risen Road](roads/risen-road.md) | Baldur's Gate ↔ Candlekeep | 120 mi | 4 | 3 | `risen-road.md` |

## Distance Matrix (miles)

| From \ To | Neverwinter | Waterdeep | Daggerford | Baldur's Gate | Candlekeep |
|-----------|-------------|-----------|------------|---------------|------------|
| Neverwinter | — | 100 | 200 | 300 | 420 |
| Waterdeep | 100 | — | 100 | 200 | 320 |
| Daggerford | 200 | 100 | — | 100 | 220 |
| Baldur's Gate | 300 | 200 | 100 | — | 120 |
| Candlekeep | 420 | 320 | 220 | 120 | — |

## Template Conventions

See [`.spec/templates.md`](../.spec/templates.md) for the YAML frontmatter schemas used across cities, roads, characters, and encounters.

## Cross-Cutting Plot Threads

- **The Five Travelers Prophecy** — Appears at Beregost's temple, Candlekeep's archives, Master Havel's ledger, and the Cormyrean Milestone. The party matches the description.
- **The Shattered Tower** — Magical anomalies, goblin raids, and a sealed lower level that is weakening. Connected to the Sunspring Villa and Candlekeep's Restricted Wing by an unknown symbol.
- **The Blight** — Spreading from the Neverwood, fed by something deliberate. Connected to river corruption at Daggerford and the burial mounds beyond the duchy's walls.
- **The Underground** — Criminal networks in Baldur's Gate planning a movement. Connected to road agent corruption on the Trade Way and bandit information leaks.
- **The Cult of Bane** — Reopened temple in Baldur's Gate slums, cultists at Sunspring Villa, warm wells that should not exist. Growing, patient, waiting.
