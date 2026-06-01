# Book of Mormon Walk — Knowledge Graph

The connection index, grown one chapter at a time. Pull from this to trace threads across the Book of Mormon and into our existing studies.

## Node types
`person` · `place` · `doctrine` · `type/symbol` · `prophecy` · `covenant` · `event` · `study-link` (→ our 198 studies, found via `study_search` / `study_similar`)

## Edge types
`cross-ref` · `fulfillment` (prophecy→event) · `parallel` · `type→antitype` · `covenant-thread` · `doctrinal-development` · `links-to-study`

## Conventions
- Stable node ids: `person:lehi`, `doctrine:tender-mercies`, `type:liahona`, `study:give-away-all-my-sins`.
- Edge line: `{from} —[type]→ {to}   (provenance ref; short note)`
- Append as the walk proceeds. Periodic synthesis passes (at book boundaries) surface emergent patterns into `_journal.md`.

## Nodes

### from 1 Nephi 1
- person: `lehi` `nephi` `sariah` `laman` `lemuel` `sam` `zedekiah`
- type/symbol: `pillar-of-fire` (Exodus) · `heavenly-book` (prophetic commission) · `rock`
- doctrine: `tender-mercies` · `deliverance` · `redemption-of-the-world`
- event: `lehi-call-vision`
- prophecy: `jerusalem-destruction` · `babylonian-captivity` · `coming-messiah`

### from 1 Nephi 2
- place: `red-sea` · `valley-of-lemuel` · `river-laman` · `promised-land` (foretold)
- doctrine: `prosperity-covenant` · `soften-the-heart` · `murmuring` · `lamanite-curse` (conditional)
- type/symbol: `lehi-as-abraham` (altar + tent) · `river→righteousness` · `valley→steadfast`
- event: `departure-into-wilderness` · `nephi-call-blessing`

## Edges

### from 1 Nephi 1
- `type:pillar-of-fire` —[cross-ref/type]→ Ex 13:21 (Israel's deliverance)
- `event:lehi-call-vision` —[type-scene]→ Isa 6 · Ezek 2–3 · Rev 10 · Moses 1 · JS–H 1
- `event:lehi-call-vision` —[cross-ref]→ Alma 36:22 (quoted ~verbatim, 2 centuries later)
- `doctrine:tender-mercies` —[verbal-root]→ Ps 145:9 · —[pattern]→ Mosiah 29:20 · —[links-to-study]→ know-god, divine-love
- `prophecy:jerusalem-destruction` —[fulfillment]→ 586 BC (Omni 1:15; 2 Ne 25)

### from 1 Nephi 2
- `person:lehi` —[type/parallel]→ Abraham (Gen 12; Abr 2)
- `doctrine:prosperity-covenant` —[fountainhead]→ 1 Ne 2:20-21 · —[recurs]→ 2 Ne 1:20 · 2 Ne 4:4 · Mosiah 2:22 · Alma 9:13 (to Lehi) · Alma 50:20 · —[OT-root]→ 1 Sam 12:14 · Josh 1:7
- `doctrine:soften-the-heart` —[instance]→ 1 Ne 2:16 · —[links-to-study]→ softening-what-i-cannot-soften
- `person:nephi` —[made]→ ruler-and-teacher (2:22)
- `doctrine:lamanite-curse` —[purpose]→ remembrance (2:24)
