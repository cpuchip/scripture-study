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

### from 1 Nephi 3
- person: `laban` · `the-angel`
- object: `brass-plates` (record of the Jews + genealogy)
- place: `cavity-of-a-rock`
- doctrine: `prepare-a-way` · `scripture-preservation` · `reason-from-God-vs-obstacle`
- type/symbol: `nephi-as-joseph` (younger ruler, smitten by elders)

### from 1 Nephi 4
- person: `zoram` (servant → freed → joins; Zoramite line)
- object: `labans-sword` (gold hilt, precious steel) · `labans-garments`
- doctrine: `one-for-many` (4:13) · `obedience-before-sight` · `spirit-constraint`
- type/symbol: `nephi-as-abraham` (4:6, not knowing) · `nephi-as-david` (4:18, own sword) · `laban-as-pharaoh`
- event: `slaying-of-laban` · `plates-obtained` · `zoram-oath`
- study: `1ne4_slaying-of-laban` (spin-off, COMPLETE — one-for-many = Caiaphas-mirror, neutral logic; Mosiah 1:5 confirms)

### from 1 Nephi 5
- person: `sariah` (complaint → own witness) · `joseph-of-egypt` · `jeremiah`
- doctrine: `preservation` (seed/record/word) · `independent-witness` · `brass-plates-canon`
- type/symbol: `lehi-as-joseph-dreamer` (Gen 37:19)

### from 1 Nephi 6
- doctrine: `purpose-of-the-record` (persuade to Christ) · `curation-by-worth` · `please-God-not-world`
- object: `small-plates` (distinct from Lehi's record)

### from 1 Nephi 7
- person: `ishmael` + household (daughters/sons → wives of Lehi's sons)
- doctrine: `deliverance-by-faith` (burst bands) · `remembrance-vs-forgetting` · `frank-forgiveness` · `spirit-ceaseth-to-strive`
- pattern: `laman-lemuel-cycle` (template)

### from 1 Nephi 8
- symbol: `tree-of-life` · `fruit` · `rod-of-iron` · `strait-narrow-path` · `mist-of-darkness` · `great-spacious-building` · `river-of-water` · `dark-dreary-waste`
- person: `white-robed-guide`
- pattern: `four-groups` (responses to the fruit)
- event: `tree-of-life-dream`

### from 1 Nephi 9
- object: `large-plates` (kings/wars) — vs `small-plates` (ministry); both "plates of Nephi"
- doctrine: `prepare-a-way-across-time` · `obedience-without-reason` · `divine-foreknowledge`

### from 1 Nephi 10
- person: `the-messiah`/`lamb-of-god` · `the-forerunner` (John the Baptist) · `the-gentiles`
- doctrine: `seek-and-find` (gift to all) · `God-unchanging`/`one-eternal-round` · `the-fall`+reliance · `scattering-and-gathering`
- symbol: `olive-tree`

### from 1 Nephi 11
- person: `the-spirit-of-the-lord` · `the-angel` · `the-virgin`(Mary) · `lamb-of-god`/`son-of-eternal-father` · `twelve-apostles`
- doctrine: `condescension-of-God` (Father birth + Son cross) · `belief-before-sight` · `incarnation-as-love-of-God`
- symbol: `fountain-of-living-waters` · `high-mountain`(revelation-place)
- study: `1ne11_condescension-of-god` (spin-off, in progress)

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

### from 1 Nephi 3
- `doctrine:prepare-a-way` —[source-text]→ 1 Ne 3:7 · —[restated]→ 1 Ne 17:3 · —[root]→ Gen 18:14 · Philip 4:13
- `person:nephi` —[type]→ Joseph of Egypt (Gen 41:43; smitten by elders, raised to rule)
- `object:brass-plates` —[purpose]→ preserve language + prophets' words (3:19-20) · —[contrast]→ Omni 1:17 (Zarahemla lost both)
- `person:laman-lemuel` —[unchanged-by]→ angelic ministry (3:31): conversion ≠ evidence
- `person:laban` —[lust→downfall]→ 1 Ne 4 (delivered into their hands)

### from 1 Nephi 4
- `event:slaying-of-laban` —[deep-dive]→ studies/1ne4_slaying-of-laban.md
- `doctrine:one-for-many` —[stated]→ 1 Ne 4:13 · —[parallel]→ John 11:50 (Caiaphas) · —[points-to]→ Christ
- `event:slaying-of-laban` —[type]→ David & Goliath (own sword, 1 Sam 17:51)
- `person:nephi` —[type]→ Abraham (Heb 11:8) + David (4:18) + Joseph (ch 3)
- `doctrine:prosperity-covenant` —[remembered-as-hinge]→ 1 Ne 4:14 (cites 2:20)
- `person:zoram` —[freed→joins]→ covenant family · —[line]→ Zoramites (Jacob 1:13; 4 Ne 1:36)
- `type:exodus` —[Laban=Pharaoh]→ 1 Ne 4:2-3

### from 1 Nephi 5
- `person:sariah` —[arc]→ complaint (5:2) → "surety" (5:8); contrast `person:laman-lemuel`
- `person:lehi` —[type]→ Joseph the dreamer (Gen 37:19) · —[lineage]→ Joseph of Egypt (5:14) → Gen 49:22 / 2 Ne 3
- `object:brass-plates` —[contains]→ Torah + Jewish record + prophets + Jeremiah + genealogy (5:11-14) · —[fuller-than]→ surviving OT
- `doctrine:preservation` —[motif]→ Joseph→family→plates-never-perish→commandments-to-children (5:14-21)
- `object:brass-plates` —[destiny]→ all nations + never perish (5:18-19; Alma 37:4)

### from 1 Nephi 6
- `doctrine:purpose-of-the-record` —[stated]→ 1 Ne 6:4 · —[bookend]→ Moro 10:32 · —[parallel]→ John 20:31
- `doctrine:curation-by-worth` —[stated]→ 1 Ne 6:3-6 · —[inherited]→ Mormon/Moroni (W of M 1:4)
- `object:small-plates` —[distinct-from]→ Lehi's record / large plates (6:1)

### from 1 Nephi 7
- `pattern:laman-lemuel-cycle` —[template-set]→ 1 Ne 7 · —[recurs]→ 1 Ne 16, 17, 18 · —[prefigures]→ Nephite pride-cycle
- `doctrine:remembrance` —[unbelief=forgetting]→ 7:10-12 · —[root]→ 2:24
- `doctrine:soften-the-heart` —[intercessory]→ 7:5, 19
- `doctrine:deliverance-by-faith` —[bursts-bands]→ 7:17 · —[recurs]→ Alma 14:28; 3 Ne 28:20; Judg 14:6
- `doctrine:frank-forgiveness` —[Nephi]→ 7:21 · —[parallel]→ Luke 7:42
- `event:jeremiah-imprisoned` —[confirms]→ brass-plates Jeremiah (5:13; Jer 37:15)

### from 1 Nephi 8
- `symbol:tree-of-life` —[interpreted]→ 11:21-22 (love of God/Christ) · —[root]→ Gen 2:9; Rev 22:2 · —[answers]→ Gen 3:6
- `symbol:rod-of-iron` —[=word-of-God]→ 11:25; 15:23-24 · —[verb: clinging]→ 8:24,30
- `symbol:mist-of-darkness` —[=temptation]→ 12:17 (Matt 13:19)
- `symbol:great-spacious-building` —[=pride, foundationless]→ 11:36; 12:18 (Eph 2:2)
- `pattern:four-groups` —[dream-source]→ 8 · —[studies]→ iron-rod-anchor-and-the-four-groups, four-groups-and-the-engineer
- `doctrine:tender-mercies` —[inside-dream]→ 8:8 (cf 1:20)
- `event:tree-of-life-dream` —[framed-by]→ Lehi's fear for his sons (8:3-4, 35-37)

### from 1 Nephi 9
- `doctrine:prepare-a-way` —[scale=millennia]→ 9:5-6 · —[fulfilled]→ D&C 10:38-40 (lost 116 pages) · —[partner]→ 3:7
- `object:small-plates` —[vs]→ `large-plates` (kings/wars, 9:4) · —[both]→ "plates of Nephi" (9:2)
- `doctrine:obedience-without-reason` —[9:5]→ grounded in 9:6 (God knows all); cf 4:6

### from 1 Nephi 10
- `doctrine:seek-and-find` —[gift-to-ALL]→ 10:17-19 · —[drives]→ 1 Ne 11-14 · —[contrast]→ 15:8-9 (L&L did not ask) · —[root]→ Matt 7:7
- `prophecy:messiah` —[named/dated]→ 10:4-10 (600 yrs; Lamb of God) · —[fulfilled]→ 3 Ne 1:1; John 1:29
- `doctrine:God-unchanging` —[one-eternal-round]→ 10:18-19 (Heb 13:8) — basis of continuing revelation
- `doctrine:prepare-a-way` —[soteriological, 3rd deepening]→ 10:18 (3:7 → 9:6 → 10:18)
- `symbol:olive-tree` —[scatter/gather]→ 10:12-14 · —[allegory]→ Jacob 5 · —[family-branch]→ Gen 49:22
- `doctrine:the-fall` —[+reliance-on-Christ]→ 10:6 (→ 2 Ne 2)

### from 1 Nephi 11
- `event:nephi-vision` —[granted-by]→ desire+belief (11:1-7; fulfills 10:17-19) · —[contrast]→ 15:8-9
- `symbol:tree-of-life` —[=love-of-God]→ 11:22 (Rom 5:5) · —[shown-as]→ incarnation (11:18-21) · —[interprets]→ 1 Ne 8
- `doctrine:condescension-of-God` —[Father:birth]→ 11:16-21 · —[Son:ministry+cross]→ 11:26-33 · —[study]→ 1ne11_condescension-of-god
- `symbol:rod-of-iron` —[=word→tree]→ 11:25 (resolves 8:19,24)
- `symbol:great-spacious-building` —[=pride, fights apostles, falls]→ 11:35-36 (resolves 8:26)
