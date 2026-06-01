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

## Edges

### from 1 Nephi 1
- `type:pillar-of-fire` —[cross-ref/type]→ Ex 13:21 (Israel's deliverance)
- `event:lehi-call-vision` —[type-scene]→ Isa 6 · Ezek 2–3 · Rev 10 · Moses 1 · JS–H 1
- `event:lehi-call-vision` —[cross-ref]→ Alma 36:22 (quoted ~verbatim, 2 centuries later)
- `doctrine:tender-mercies` —[verbal-root]→ Ps 145:9 · —[pattern]→ Mosiah 29:20 · —[links-to-study]→ know-god, divine-love
- `prophecy:jerusalem-destruction` —[fulfillment]→ 586 BC (Omni 1:15; 2 Ne 25)
