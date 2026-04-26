# D&D Campaign — Markdown Template Conventions

All campaign documents use YAML frontmatter for machine-parsable data, followed by prose markdown for narrative content.

---

## City Template (`world/cities/`)

```yaml
---
name: City Name
region: Sword Coast North | Sword Coast Central | Sword Coast South | The North
population: number
government: type (e.g., Council of Lords, Duchy, Merchant Patriarchs)
deity_alignment: primary deity or pantheon
notable_factions: [faction1, faction2]
currency: primary currency
danger_level: 1-5 (1 = safe, 5 = lethal)
services:
  inns: [inn_name, inn_name]
  temples: [temple_name]
  shops: [shop_type, shop_type]
  guilds: [guild_name]
distances:
  "CityName": miles_number
  "CityName": miles_number
key_districts: [district1, district2]
secret_locations: [location]  # DM-only knowledge
plot_hooks: [hook_summary]
---
```

**Prose sections (after frontmatter):**
- `## Appearance` — first impressions, atmosphere, sensory details
- `## History` — brief origin, major events
- `## Districts` — neighborhood-by-neighborhood breakdown
- `## Notable NPCs` — key people the party might meet
- `## Sights & Sounds` — immersive detail for roleplay
- `## Secrets` — hidden truths, DM-only

---

## Road Template (`world/roads/`)

```yaml
---
name: Road Name
connects: [CityA, CityB]
total_distance_miles: number
travel_time_days: number
road_type: paved | cobblestone | dirt track | mountain pass
danger_level: 1-5
notable_locations:
  - name: Location Name
    type: tavern | ruin | crossing | village | landmark | ambush_site
    distance_from_start: miles_number
    description: short_summary
    plot_relevance: high | medium | low
hazards: [hazard1, hazard2]
supplies_available: [supply_type]
weather_notes: seasonal_conditions
---
```

**Prose sections (after frontmatter):**
- `## The Route` — overview of the journey
- `## Along the Way` — detailed stops, encounters, landmarks
- `## Dangers` — threats specific to this road
- `## Rumors` — what travelers say about this route

---

## Character Template (`characters/`)

```yaml
---
character_name: Name
gender: Male | Female | Nonbinary
race: Race
class: Class
subclass: Subclass
level: number
background: Background
alignment: Alignment
proficiency_bonus: number
proficiencies:
  armor: [types]
  weapons: [types]
  tools: [types]
  saving_throws: [abilities]
skills: [skill1, skill2]
languages: [lang1, lang2]
equipment: [item1, item2]
xp: number
death_saves:
  successes: 0
  failures: 0
features: [feature1, feature2]
inspiration: 0
notes: freeform_text
---
```

**Prose sections:**
- `## Appearance`
- `## Background`
- `## Motivation`
- `## Personality`
- `## Connections`
- `## Secrets`

---

## Encounter Template (`campaign/encounters/`) — future use

```yaml
---
encounter_name: Name
location: city_or_road_reference
trigger: condition
difficulty: easy | medium | hard | lethal
initiative_order: [combatant1, combatant2]
enemies:
  - name: Enemy Name
    cr: number
    hp: number
    special_abilities: [ability]
rewards:
  xp: number
  loot: [item]
  story: narrative_outcome
---
```

---

## Naming Conventions

- **Filenames:** lowercase, hyphenated, no spaces (`kaelen-ironheart.md`, `coast-way.md`)
- **City names in distances:** quoted strings (`"Waterdeep": 120`)
- **Danger levels:** 1-5 scale, consistent across all documents
- **Distances:** always in miles, always from a named reference point

## Cross-Referencing

Cities reference roads by filename. Roads reference cities by name in `connects:`. Characters reference locations by city name in their backgrounds.
