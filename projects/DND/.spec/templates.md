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
subclass: "—" (chooses at level X)
level: 1
background: Background
alignment: Alignment
proficiency_bonus: 2
hit_dice: 1dX
ability_scores:
  strength: 15
  dexterity: 14
  constitution: 13
  intelligence: 12
  wisdom: 10
  charisma: 8
modifiers:
  strength: 2
  dexterity: 2
  constitution: 1
  intelligence: 1
  wisdom: 0
  charisma: -1
hp: 12
max_hp: 12
ac: 16
initiative: +2
speed: 30
saving_throws:
  strength: 4
  constitution: 3
skill_proficiencies:
  - SkillName (+modifier)
armor_proficiencies: [types]
weapon_proficiencies: [types]
spellcasting_ability: Charisma | Intelligence | Wisdom | "—"
spell_save_dc: 13
spell_attack_bonus: 5
cantrips_known: 0
spells_known: 0
spell_slots:
  level_1: 0
cantrips: []
spells: []
languages: [lang1, lang2]
equipment:
  - Item (+X to hit, YdZ+W damage type)
gold: 15
xp: 0
death_saves:
  successes: 0
  failures: 0
level_1_features: [feature1, feature2]
racial_traits: [trait1, trait2]
notes: freeform_text
---
```

**Level 1 rules (standard array: 15, 14, 13, 12, 10, 8):**
- HP = max hit die + Con modifier
- Proficiency bonus = +2 at level 1
- Subclass chosen at level 3 (except Wizard: level 2, Sorcerer/Cleric: level 1)
- Spell save DC = 8 + proficiency bonus + spellcasting ability modifier
- Spell attack bonus = proficiency bonus + spellcasting ability modifier
- Attack bonus = proficiency bonus + relevant ability modifier (Str for melee, Dex for ranged/finesse)

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
