# Doctrine & Covenants Walk — Knowledge Graph

Nodes (people · places · doctrines · covenants · ordinances · offices · revelations) and edges, accreted per section. Pull from this for connections — within the D&C, back into the [BoM walk graph](../bom-walk/_graph.md) and the [PoGP walk graph](../pogp-walk/_graph.md), and into our existing studies.

Edge vocabulary: `cross-ref` · `fulfillment` · `parallel` · `covenant-thread` · `links-to-study` · `cross-walk` (→ BoM / PoGP) · `project-source` (a text the workspace's own covenant/principles are built on) · `historical-setting`.

---

<!-- Sections append below, newest after oldest. Format: a short node/edge block per section. -->

## D&C 1 — the Lord's preface (Nov 1831, Hiram OH)
**Nodes:** `★ doctrine:revelation-accommodated` (1:24, "after the manner of their language") · `★ principle:voice-of-servants=Lord's` (1:38) · `doctrine:weak-break-the-mighty` (1:19,23) · `image:idolatry-as-self` (1:16) · `word:nevertheless` (1:31-32) · `church:only-true-and-living` (1:30, "collectively not individually") · `truth:abideth-forever` (1:39) · `doctrine:universal-voice-of-warning` (1:2,11,34-35)
**Edges:**
- `★ revelation-accommodated` (1:24) —[project-source: scriptural floor of]→ Period-Language-Reading principle · —[=]→ 2 Ne 31:3; Ether 12:39; D&C 67:5 · —[growth-model]→ 1:25-28 ("from time to time")
- `★ voice-of-servants=Lord's` (1:38) —[project-source: delegation/stewardship]→ intent-over-instruction; Abr 2:6 · —[reject-servant=reject-Lord]→ 1:14 · —[→]→ D&C 21:5 (coming)
- `weak-break-the-mighty` (1:19,23) —[cross-ref]→ 1 Cor 1:27; 2 Ne 28:31 (trust-not-arm-of-flesh)
- `idolatry-as-self` (1:16) —[idol's-substance-perishes vs truth-abides 1:39]→ matter-spectrum
- `word:nevertheless` (1:31-32) —[hinge]→ holiness ↔ mercy · —[links-study]→ nevertheless
- `church:only-true-and-living` (1:30) —[living=revelatory]→ Article 9 · —[built-in-humility]→ "collectively not individually"
- **cross-walk:** `Spirit-not-always-strive` (1:33) —[=]→ Moses 8:17; Ether 15:19 (terminal-state thread → now in D&C)
- `truth:abideth-forever` (1:39) —[links-study]→ truth.md

## D&C 2 — Elijah / the promises to the fathers (Sept 1823, Moroni to Joseph)
**Nodes:** `prophecy:Elijah-reveals-the-priesthood` (2:1) · `★ covenant:promises-to-the-fathers` (2:2) · `doctrine:earth-wasted-without-sealing` (2:3) · `seed:temple-sealing-genealogy-theology`
**Edges:**
- `Elijah-reveals-the-priesthood` (2:1) —[changed-from]→ Mal 4:5 ("send you Elijah") · —[fulfilled]→ **D&C 110:13-16** (1836) · —[=sealing-keys]→ Matt 16:19
- `★ promises-to-the-fathers` (2:2) —[=Abrahamic-covenant]→ **Abr 2:9-11** · —[cross-walk]→ **JS—H 1:38-39** (PoGP) · —[expounded]→ D&C 128:18
- `earth-wasted-without-sealing` (2:3) —[earth's-purpose]→ Moses 1:39; Abr 3:24-25 (no seal → no eternal family → creation pointless)
- **cross-walk:** Malachi *changed* here vs *KJV-form* in 3 Ne 25:5-6 (Christ quotes the same)
- ✦ **promises-to-the-fathers thread now spans all 3 walks:** Abr 2 → JS-H 1:38 → D&C 2 → (D&C 110/128 ahead)

## D&C 3 — the lost 116 pages (Harmony, July 1828 — earliest revelation)
**Nodes:** `doctrine:Gods-work-cannot-be-frustrated` (3:1,3) · `God:one-eternal-round` (3:2) · `★ warning:gifts-dont-make-safe` (3:4) · `sin:fear-of-man` (3:7) · `★ pattern:discipline-then-restoration` (3:9-10) · `word:nevertheless` (3:16) · `meta:canon-rebukes-its-prophet`
**Edges:**
- `Gods-work-cannot-be-frustrated` (3:1,3) —[only-mens-fails]→ Acts 5:38; Morm 8:22 · —[provided-2000-yrs-prior]→ D&C 10 (small plates)
- `★ warning:gifts-dont-make-safe` (3:4) —[cross-ref]→ Mosiah 11:19; 1 Cor 10:12 · —[project]→ trust-not-arm-of-flesh (1:19)
- `sin:fear-of-man` (3:7) —[vs]→ 1:38 (voice of God > voice of men)
- `★ pattern:discipline-then-restoration` (3:9-10) —[FIRST-instance-of]→ **D&C 121:43** (reprove w/ sharpness → increase of love) · —[project]→ presiding covenant
- `word:nevertheless` (3:16) —[hinge]→ failure → "my work shall go forth"
- `meta:canon-rebukes-its-prophet` —[=]→ include_failures (covenant) · —[honesty]→ JS-H 1:28

## D&C 4 — desire is the call (Harmony, Feb 1829, to Joseph Smith Sr.)
**Nodes:** `doctrine:desire-is-the-call` (4:3) · `doctrine:character-qualifies` (4:5-6, the 2 Pet 1 ladder) · `single-eye→light` (4:5) · `command:whole-souled-service` (4:2)
**Edges:**
- `desire-is-the-call` (4:3) —[★ entry-point thread]→ Abr 1:2 (covenant begins w/ desire); A of F 13 (seek after these things) — recurring across all 3 walks
- `character-qualifies` (4:5-6) —[=]→ 2 Pet 1:5-7
- `single-eye→light` (4:5) —[forward-link]→ **D&C 88:67** → matter-spectrum
- `whole-souled-service` (4:2) —[=]→ Deut 6:5; Mark 12:30
- `ask-and-receive` (4:7) —[=]→ Matt 7:7; James 1:5 (experimental epistemology)

## D&C 5 — the Three Witnesses; proof can't give faith (Harmony, Mar 1829, Martin Harris)
**Nodes:** `★ epistemology:proof-doesnt-compel-faith` (5:7,18) · `witness:unique-vs-democratic` (5:11-16) · `law-of-witnesses` (5:11-15) · `stewardship:bounded-gift` (5:4) · `condition:humility-not-curiosity` (5:24) · `promise:eternal-life-even-if-slain` (5:22)
**Edges:**
- `★ proof-doesnt-compel-faith` (5:7,18) —[=]→ Luke 16:31 · —[witness=condemnation-not-coercion]→ 5:18 · —[real-witness=Spirit]→ 5:16 · —[thread]→ period-language/Spirit-distillation
- `witness:unique-vs-democratic` — the Three SEE (5:11-14) vs "whosoever believeth" is REBORN (5:16); the common witness is greater
- `law-of-witnesses` (5:11-15) —[=]→ Deut 19:15; 2 Cor 13:1 · —[BoM-walk]→ Ether 5:3-4 · —[fulfilled]→ D&C 17
- `stewardship:bounded-gift` (5:4) —[honor-scope]→ D&C 107:99 (ahead)
- `promise:eternal-life-even-if-slain` (5:22) —[foreshadows]→ D&C 135 (martyrdom)
- `pattern:I-will-provide-means` (5:34) —[=]→ 1 Ne 3:7 · —[fulfilled]→ D&C 6 (Oliver)

## D&C 6 — the witness Oliver already had (Harmony, April 1829, Oliver Cowdery)
**Nodes:** `★ epistemology:witness-is-retrospective-recognition` (6:22-24) · `revelation:peace-to-the-mind` (6:23) · `doctrine:desire-is-the-hinge` (6:8,20) · `image:arms-of-my-love` (6:20) · `Christ:light-shineth-in-darkness` (6:21) · `command:stand-by-Joseph` (6:18-19)
**Edges:**
- `★ witness-is-retrospective-recognition` (6:22-24) —[builds-on]→ D&C 5:7,16 (proof ≠ faith) · —[=peace-to-the-mind]→ 6:23 · —[private-knowledge-proof]→ 6:16,24 (Ps 139:2; 1 Chr 28:9)
- `revelation:peace-to-the-mind` (6:23) —[pairs]→ D&C 8:2 (mind+heart) → D&C 9:8 (burn in bosom) — the how-revelation-feels triad
- `doctrine:desire-is-the-hinge` (6:8,20) —[thread]→ D&C 4:3; 7:8; Abr 1:2
- `image:arms-of-my-love` (6:20) —[links-study]→ divine-love · —[knowing-as-embrace, not exposure]→ 6:16
- `Christ:light-shineth-in-darkness` (6:21) —[=]→ John 1:5 · —[forward]→ matter-spectrum
- `command:stand-by-Joseph` (6:18-19) —[mutual-correction]→ "admonish him… and also receive admonition of him"

## D&C 7 — a man's desire becomes his destiny (Harmony, April 1829; John's translated parchment)
**Nodes:** `doctrine:desire-shapes-destiny` (7:8) · `being:translated` (7:3,6, John) · `office:keys-Peter-James-John` (7:7) · `meta:section-IS-the-parchment`
**Edges:**
- `desire-shapes-destiny` (7:8) —[sharpest-form-of]→ D&C 4:3 / 6:8,20 · —[=]→ Alma 29:4 (BoM walk) · —[two-goods-not-flattened]→ 7:5 (Peter's "good desire")
- `being:translated` (7:3,6) —[=]→ 3 Ne 28:1-12 (Three Nephites, same desire) · —[state-between]→ matter-spectrum; D&C 130:5
- `office:keys-Peter-James-John` (7:7) —[=]→ Matt 16:19; D&C 6:28 · —[fulfilled-weeks-later]→ Melchizedek restoration; JS—H 1:72 · —[cross-walk]→ priesthood-line (Moses 6:7 / Abr 1:3)
- `meta:section-IS-the-parchment` —[translated-record]→ scripture brought to light (like the BoM)

## D&C 8 — "I will tell you in your mind and in your heart" (Harmony, April 1829, Oliver)
**Nodes:** `★ revelation:mind-and-heart` (8:2, project-source) · `revelation:Red-Sea-scale` (8:3) · `★ gift-of-Aaron:the-rod` (8:6-8) · `condition:gift-runs-on-faith` (8:10-11)
**Edges:**
- `★ revelation:mind-and-heart` (8:2) —[project-source: revelation-by-Spirit anchor]→ period-language + Spirit-distillation principle · —[interior→discernment-seam]→ "is this God or me?" · —[pairs]→ D&C 9:8 (the burning)
- `revelation:Red-Sea-scale` (8:3) —[same-Spirit-every-scale]→ Ex 14:16 · —[Moses-parted-sea-BY-revelation]→ matter-spectrum
- `★ gift-of-Aaron:the-rod` (8:6-8) —[Lord-claims-seekers-instrument]→ period-language-practice (cf. 6:24) · —[historical-seam]→ Book of Commandments "working with the rod"; NE folk practice, labeled · —[dignified]→ Aaron's budding rod (Num 17)
- `condition:gift-runs-on-faith` (8:10-11) —[sets-up]→ D&C 9:7-9 (Oliver assumed automatic) · —[=]→ Heb 11:6; Moroni 7:33

## D&C 9 — study it out, then ask (Harmony, April 1829, Oliver's failed translation)
**Nodes:** `★ epistemology:study-it-out-then-confirm` (9:7-8, project-source) · `★ signal:burning-vs-stupor` (9:8-9) · `doctrine:revelation-is-time-bound` (9:11) · `pattern:reprove-then-love` (9:1,3,6,14) · `God:wisdom-withheld` (9:3,6)
**Edges:**
- `★ study-it-out-then-confirm` (9:7-8) —[project-source: cited in covenant.yaml flag_when_wrong]→ the collaboration's verification method · —[=]→ Moroni 10:4-5 · —[completes]→ D&C 8:2 (mind-and-heart → a method) · —[partnership-not-download]→ 9:7 (the named error)
- `★ signal:burning-vs-stupor` (9:8-9) —[burn=warmth-that-stays]→ Luke 24:32 (Emmaus) · —[stupor=numbness-that-fades]→ Webster 1828 (suspension of sensibility); D&C 10:2 · —[negative-signal-is-absence→hard-to-read]→ period-language/Spirit-distillation
- `doctrine:revelation-is-time-bound` (9:11) —[fear-forfeits]→ "the time is past" · —[real-loss-not-condemnation]→ 9:12
- `pattern:reprove-then-love` (9 "my son"×3; 9:14) —[gentlest-form-of]→ D&C 3:9-10 / 121:43 · —[calibrated]→ lesser fault → lesser sharpness than Joseph's §3
- `God:wisdom-withheld` (9:3,6) —[trust-character-not-explanation]→ §10 "a wise purpose" (a withheld reason about to be revealed)
- ✦ **byu_citations:** D&C 9:8 cited 25× in Gen. Conf. 1944→2020 (Lee, Dunn, Scott×2, Nelson, Christofferson…); modern emphasis = the *study-it-out* half Oliver skipped

## D&C 10 — wisdom greater than the cunning of the devil (Harmony, ~April 1829; the 116-pages resolution)
**Nodes:** `★ command:pace-to-strength-and-means` (10:4, project-source) · `★ providence:wise-purpose-redundancy` (10:38-45) · `God:wisdom-greater-than-cunning` (10:43) · `discipline:hold-the-sacred-when-hearts-unknown` (10:37) · `doctrine:minimal-core-of-Christ` (10:67-68) · `Christ:light-shineth-in-darkness` (10:58)
**Edges:**
- `★ pace-to-strength-and-means` (10:4) —[project-source: covenant.yaml not_bypass_process]→ "the process IS the pacing mechanism" · —[direct-lesson-of]→ §3 (Joseph over-reached → lost the pages) · —[=]→ Mosiah 4:27; Ex 18:18
- `★ providence:wise-purpose-redundancy` (10:38-45) —[built-2000-yrs-early]→ W of M 1:3-7; 1 Ne 9:2-6 · —[CLOSES]→ D&C 3:1-3 (work can't be frustrated) · —[loss-yielded-better-book]→ small plates "throw greater views" (10:45)
- `God:wisdom-greater-than-cunning` (10:43) —[win-by-refusing-the-move]→ 10:30 (translate not again) · —[declines-compulsory-route]→ presiding (long game, not strong arm)
- `discipline:hold-the-sacred-when-hearts-unknown` (10:37) —[=]→ D&C 6:12 · —[thread]→ temple-reserve (PoGP facsimiles)
- `doctrine:minimal-core-of-Christ` (10:67-68) —[=]→ 3 Ne 11:39-40; D&C 20:37 · —[both-err]→ "more OR less than this"
- `Christ:light-shineth-in-darkness` (10:58) —[repeats]→ D&C 6:21 (Oliver-cluster signature) · —[other-sheep]→ John 10:16; 3 Ne 15
- ✦ **decade 1-10 CLOSED:** the §3↔§10 lost-pages arc is the bookend; three project-source landings in §§8-10 (8:2, 9:7-9, 10:4)

## D&C 11 — obtain the word before you declare it (Harmony, May 1829, Hyrum Smith)
**Nodes:** `★ teaching:obtain-before-declare` (11:21) · `discernment:fruit-test-of-spirits` (11:12, Micah 6:8) · `honor-scope:dont-self-commission` (11:15) · `liturgy:one-call-issued-identically` (11:1-9) · `promise:power-to-become-sons-of-God` (11:30)
**Edges:**
- `★ obtain-before-declare` (11:21) —[teaching covenant: discovery-not-performance / Ben Test]→ fill the vessel first · —[same-logic]→ D&C 9:8 (study before the ask) · —[=]→ Alma 17:2-3 (searched scriptures THEN went)
- `discernment:fruit-test-of-spirits` (11:12) —[Micah 6:8 → spectrometer]→ Moroni 7:16-17; Matt 7:16 · —[external-check-on]→ D&C 8:2 (is-it-God-or-me → test the direction)
- `honor-scope:dont-self-commission` (11:15) —[=]→ D&C 5:4; 107:99 (ahead) · —[call-to-work ≠ call-to-preach]→ 11:27 vs 11:15
- `liturgy:one-call-issued-identically` (11:1-9 = §§12,14) —[no-bespoke-gospel]→ same white field to every laborer

## D&C 12 — the qualifications of a helper (Harmony, May 1829, Joseph Knight Sr.)
**Nodes:** `doctrine:character-qualifies-the-helper` (12:8) · `stewardship:trustworthy-with-the-entrusted` (12:8) · `desire-is-the-call` (12:7,9)
**Edges:**
- `character-qualifies-the-helper` (12:8) —[same-bar-as]→ D&C 4:5-6 (the laborer) · —[worth-in-the-soul-not-the-storehouse]→ the material benefactor held to the front-line standard
- `stewardship:trustworthy-with-the-entrusted` (12:8) —[=]→ 1 Thes 2:4 · —[project]→ exercise_stewardship; Luke 16:10
- `desire-is-the-call` (12:7,9) —[compressed]→ §4/§6/§7/§11 (desire + heed-with-might = called)

## D&C 13 — the priesthood returns (Susquehanna bank, May 15 1829; John the Baptist)
**Nodes:** `★ doctrine:authority-by-conferral` (13:1 "I confer") · `office:Aaronic-Priesthood` (13:1) · `promise:never-taken-again` (13:1) · `★ temple-limit:sons-of-Levi-offering` (13:1, Mal 3:3) · `principle:fellow-servants` (13:1)
**Edges:**
- `★ authority-by-conferral` (13:1) —[hinge-of-Restoration: apostasy=loss-of-authority]→ remedy = a literal sending · —[re-splices]→ priesthood-line Moses 6:7 / Abr 1:3 (PoGP) · —[answers]→ JS—H (1820 "all wrong" → 1829 restored authority)
- `office:Aaronic-Priesthood` (13:1) —[three-keys]→ ministering of angels + gospel of repentance + baptism by immersion · —[=preparatory]→ D&C 84:26-28 · —[conferred-by]→ John the Baptist under Peter-James-John (§7)
- `promise:never-taken-again` (13:1) —[dispensation-of-the-fulness]→ D&C 27; 128 · —[contrast]→ all prior dispensations ended in apostasy
- `★ temple-limit:sons-of-Levi` (13:1) —[=]→ Mal 3:3; D&C 128:24 · —[Malachi-BRACKET]→ §2 (Elijah/sealing keys) ↔ §13 (Aaronic/lesser keys) frame the whole restoration · —[reserved/forward]→ flag for Michael
- `principle:fellow-servants` (13:1) —[heaven-runs-on-fellowship-not-rank]→ D&C 1:38

## D&C 14 — the Creator calls a witness (Fayette, June 1829, David Whitmer)
**Nodes:** `★ Christology:Christ-is-the-Creator` (14:9) · `witness:participates-in-un-hideable-light` (14:8-9) · `gift:eternal-life-is-greatest` (14:7) · `prophecy:Gentiles→Israel-reversal` (14:10)
**Edges:**
- `★ Christ-is-the-Creator` (14:9) —[=]→ Abr 4:12 (the Gods organizing); Mosiah 4:2; 3 Ne 9:15; John 1:3 · —[is-Christ-God thread]→ BoM + PoGP walks → matter-spectrum · —[source-light]→ "cannot be hid in darkness" (intensifies 6:21/10:58)
- `witness:participates-in-un-hideable-light` (14:8-9) —[the-witness-cannot-be-hid]→ David's never-denied testimony (d.1888 tombstone) [bin-4-adjacent, biography]
- `gift:eternal-life-is-greatest` (14:7) —[=]→ D&C 6:13 · —[rare-gift-serves-supreme-gift]→ witness (means) vs eternal life (end)
- `prophecy:Gentiles→Israel-reversal` (14:10) —[=]→ 1 Ne 13:42; 15:13-20

## D&C 15-16 — the thing of most worth (Fayette, June 1829, John & Peter Whitmer Jr. — VERBATIM TWINS)
**Nodes:** `doctrine:most-worth-is-bringing-souls` (15:6/16:6) · `★ doctrine:most-worth-is-universal-not-bespoke` (15≡16) · `register:sharpness-as-urgency` (15:2/16:2) · `witness:private-knowledge` (15:3/16:3) · `pattern:family-gathered-as-household` (14-15-16)
**Edges:**
- `most-worth-is-bringing-souls` (15:6/16:6) —[worth-to-you = worth-through-you]→ "rest WITH them" · —[seed-of]→ D&C 18:10-16 (worth of souls)
- `★ most-worth-is-universal-not-bespoke` (15≡16, verbatim) —[the-sameness-IS-the-message]→ no hidden unique destiny; the highest call is the common one · —[makes-literal]→ §11 (one call issued identically)
- `register:sharpness-as-urgency` (15:2) —[not-rebuke-here, John is blessed]→ §121:43 sharpness's non-rebuking face
- `witness:private-knowledge` (15:3) —[=]→ D&C 6:16,24 (known → convinced)
- `pattern:family-gathered-as-household` (14-15-16) —[=]→ Smiths, Knights · —[sealing/family]→ §2 · biography: same-call-diverging-lives (David+John left 1838; Peter Jr. faithful to 1836) [bin-4-adjacent]

## D&C 17 — by faith you shall see (Fayette, June 1829; the Three Witnesses' call)
**Nodes:** `★ epistemology:faith-precedes-sight` (17:1-2) · `witnesses:a-shield-for-the-prophet` (17:4) · `participation:same-power-faith-gift` (17:7) · `relics:convergence-of-sacred-history` (17:1) · `faith:the-prophets-of-old` (17:2)
**Edges:**
- `★ faith-precedes-sight` (17:1-2) —[CAPS the arc]→ §5 (proof≠faith) → §6 (retrospective recognition) → §9 (study-then-confirm) → §17 (faith THEN sight) · —[=]→ Ether 12:7-22 · —[saw-BECAUSE-believed]→ "full purpose of heart"
- `witnesses:a-shield-for-the-prophet` (17:4) —[law-of-witnesses]→ Deut 19:15; §5:11-15 · —[disarms-§10-trap]→ no single point of failure · —[human-form-of]→ §10 wise-purpose redundancy
- `participation:same-power-faith-gift` (17:7) —[raised-to-prophets-footing]→ "even as my servant Joseph has seen them" (17:5)
- `relics:convergence` (17:1) —[Jaredite+Lehite+Nephite-in-one-view]→ sword of Laban (1 Ne 4) + interpreters (Ether 3) + Liahona (1 Ne 16)
- `faith:the-prophets-of-old` (17:2) —[=]→ Ether 12; Heb 11 (shared faith, not just shared doctrine)
- ✦ **§5 promise FULFILLED at §17** — the law-of-witnesses thread (§5→§14→§17) lands; the verification arc §5→§17 closes

## D&C 18 — the worth of souls (Fayette, June 1829; the Twelve foreseen)
**Nodes:** `★ doctrine:worth-of-souls = the-Atonement's-cost` (18:10-11) · `math:one-soul-is-worth-a-life` (18:13-16) · `salvation:by-taking-the-name-of-Christ` (18:21-25) · `apostleship:by-desire-and-works` (18:27,37-38) · `office:the-Twelve` (18:26-32, pre-organization)
**Edges:**
- `★ worth-of-souls = the-Atonement's-cost` (18:10-11) —[you-know-worth-by-the-price]→ "suffered the pain of all men" · —[pairs]→ D&C 19:16-18 (cost shown) · —[=]→ Isa 43:4; Luke 15:7
- `math:one-soul-is-worth-a-life` (18:13-16) —[the-floor-dignified]→ "save it be one soul… how great shall be your joy" · —[completes]→ §§15-16
- `salvation:by-taking-the-name-of-Christ` (18:21-25) —[=]→ Acts 4:12; Mosiah 5:8-12 · —[name-at-the-last-day]→ temple/new-name thread [Michael, reserved] · —[look-unto-me]→ D&C 6:36
- `apostleship:by-desire-and-works` (18:27,37-38) —[desire-as-entry-point reaches the highest office]→ §4 → §18 · —[known-by-fruit]→ Matt 7:16; §11:12
- `office:the-Twelve` (18:26-32) —[foreseen-pre-organization]→ realized 1835; D&C 107:23

## D&C 19 — "I, God, have suffered these things for all" (Manchester, summer 1829, Martin Harris)
**Nodes:** `★ mercy:eternal-punishment-is-Gods-punishment` (19:6-12) · `★ Atonement:first-person-Gethsemane` (19:16-19) · `word:exquisite` (19:15) · `doctrine:hell-is-the-withdrawal-of-the-Spirit` (19:20) · `★ structure:the-cost-then-the-farm` (19:16-19 → 26-35) · `response:joy-not-guilt` (19:39)
**Edges:**
- `★ eternal-punishment-is-Gods-punishment` (19:6-12) —[Endless/Eternal = God's NAMES]→ Moses 1:3; 7:35 (PoGP) · —[not-infinite-duration → dissolves Calvinist hell-terror]→ words chosen for weight, "to work upon the hearts" · —[hermeneutical-revelation]→ how to read God's own words (cf. 1:24)
- `★ Atonement:first-person-Gethsemane` (19:16-19) —[only-from-inside]→ vs Luke 22:44 (outside) · —[the-shrinking-left-in]→ "would that I might not… and shrink" · —[pivots-on-NEVERTHELESS]→ 19:19 · —[body-AND-spirit]→ matter-spectrum (the Creator-capacity §14:9 at its redemptive limit)
- `word:exquisite` (19:15) —[Webster 1828: exquiro "to seek out"]→ sought-out/searchingly + highest-degree/very-sensibly-felt → the pain fully perceived
- `doctrine:hell-is-the-withdrawal-of-the-Spirit` (19:20) —[you-have-tasted-the-least-degree]→ §3 / §10:2 (mind darkened) · —[to-be-punished = without-Endless]→ "peace IN ME" (19:23)
- `★ structure:the-cost-then-the-farm` (19:16-19 → 26-35) —[the-juxtaposition-IS-the-sermon]→ after 19:18 the mortgage is unanswerable · —[giving-as-freedom]→ "release thyself from bondage" (19:35)
- `response:joy-not-guilt` (19:39) —[the-right-reading-of-the-Atonement]→ "canst thou read this without rejoicing?" · —[=]→ §6:20 (arms of love) at Atonement scale
- ✦ **§18 ↔ §19:** worth of souls NAMED ↔ the cost SHOWN; links truth-atonement + .scratch-how-is-it-done + nevertheless studies

## D&C 20 — the constitution of the Church (Fayette, April 6 1830; the Articles and Covenants)
**Nodes:** `★ principle:organize-before-building-enacted` (20:68) · `★ honesty:prophets-relapse-in-the-charter` (20:5) · `creed:ecumenical-then-distinct` (20:17-28) · `★ ordinance:sacrament-covenant` (20:77,79) · `offices:defined-by-duty-not-rank` (20:38-59) · `governance:common-consent` (20:65)
**Edges:**
- `★ organize-before-building-enacted` (20:68 "all things in order") —[=]→ D&C 88:119 (principles.md anchor) · —[the-creation-cycle's-Specification-step]→ the Articles and Covenants ARE the spec · —[order=protection]→ uniform ordinances (20:73), fixed doctrine (20:35)
- `★ honesty:prophets-relapse-in-the-charter` (20:5) —[include_failures, most-public-level]→ covenant.yaml teaching ext · —[the §3/§10 failure named again]→ "entangled again in the vanities" · —[grace-claim]→ 20:32 (fall from grace applies to the first elder)
- `creed:ecumenical-then-distinct` (20:17-28) —[shared-ground]→ infinite God/creation/Fall/grace · —[Restoration-distinctives]→ corporeal image (20:18 → Ether 3:15), same-gospel-from-the-beginning (20:26), falling-from-grace (20:32) · —[drafting-posture]→ §18:20 (contend against no church)
- `★ ordinance:sacrament-covenant` (20:77,79) —[weekly-renewal-of]→ §18 name-theology · —[witness / that-we-may-have-His-Spirit]→ §8:2 (indwelling) · —[bilateral]→ D&C 82:10 (covenant.yaml epigraph — administered) · —[=]→ Moroni 4-5 (BoM original)
- `offices:defined-by-duty-not-rank` (20:38-59) —[teacher = pastoral watchcare]→ 20:53-54 (watch over, by strengthening) = Abr 4:18 · —[overlap-and-assist]→ 20:52,57
- `governance:common-consent` (20:65) —[no-ordination-without-the-vote]→ D&C 26:2 · —[persuasion-not-compulsion-encoded]→ D&C 121:41

## D&C 21 — receive his word in all patience and faith (Fayette, April 6 1830; the organizing meeting)
**Nodes:** `★ delegation:prophets-word-as-if-from-mine-own-mouth` (21:5) · `★ posture:in-all-patience-and-faith` (21:5) · `intimacy:the-Lord-saw-his-weeping` (21:7-8) · `principle:authority-received-not-seized` (21:10-11)
**Edges:**
- `★ prophets-word-as-if-from-mine-own-mouth` (21:5) —[=]→ D&C 1:38 (voice-of-servants, now the living prophet) · —[project]→ covenant.yaml (intent through a steward) · —[honor-servant = honor-Lord]→ Ex 16:8; Heb 2:1
- `★ posture:in-all-patience-and-faith` (21:5) —[the-prophet-receives-progressively]→ 21:4; §1:25-28 (line upon line) · —[honest-middle]→ not-infallible (§20:5) / not-just-a-man · —[=covenant-partnership]→ §9:8 flag-when-wrong (a learning steward, not an oracle)
- `intimacy:the-Lord-saw-his-weeping` (21:7-8) —[grief-named-on-the-day-of-power]→ "his weeping for Zion I have seen" · —[private-knowledge]→ §6:22 · —[Enoch-parallel]→ Moses 7:41 (PoGP)
- `principle:authority-received-not-seized` (21:10-11) —[mutual-ordination]→ Joseph ordained BY Oliver · —[=]→ §13 "I confer" (no self-ordination, even at the top) · —[common-consent]→ §20:65
- ✦ **DECADE 11-20 CLOSED:** the Church is *born* — authority (§13) → specification (§20) → the prophet installed (§21); priesthood-line first landing (§13), is-Christ-God landing (§14), the Atonement landmark (§19)

## D&C 22 — a new and everlasting covenant (Manchester, April 16 1830; rebaptism)
**Nodes:** `★ doctrine:authority-not-sincerity` (22:2) · `concept:dead-works` (22:2-3) · `covenant:new-and-everlasting` (22:1) · `posture:seek-not-to-counsel-God` (22:4)
**Edges:**
- `★ authority-not-sincerity` (22:2) —[a-hundred-sincere-baptisms-avail-nothing]→ §13 conferral applied to every member · —[claim-about-ordinance-validity, not-a-judgment-on-people]
- `concept:dead-works` (22:2-3) —[sincere-but-lifeless: form-without-authority]→ Moroni 8:23; Heb 6:1 · —[apostasy-at-ordinance-level]→ §13 repairs
- `posture:seek-not-to-counsel-God` (22:4) —[=]→ Jacob 4:10 · —[receive-His-terms]→ §9

## D&C 23 — five men, "under no condemnation" (Manchester, April 1830)
**Nodes:** `★ grace:clean-standing-precedes-the-call` (23:1,3,4,5) · `calibration:fitted-callings` (23:1-7) · `warning:Oliver-pride-named-early` (23:1)
**Edges:**
- `★ clean-standing-precedes-the-call` ("under no condemnation") —[serve-FROM-acceptance-not-to-earn-it]→ §6:20 (arms of love)
- `calibration:fitted-callings` —[five-different-answers-to-one-desire]→ complement to §15-16 (universal worth ↔ bespoke assignment) · —[incl-"not-yet"]→ 23:4 = §11:15
- `warning:Oliver-pride-named-early` (23:1) —[fault-named-8-yrs-before-the-fall]→ 1838 estrangement [bin-4-adjacent]

## D&C 24 — strength within your calling (Harmony, July 1830; persecution)
**Nodes:** `★ doctrine:strength-is-calling-bounded` (24:9) · `comfort:afflictions-promised-not-removed` (24:8) · `pattern:double-nevertheless` (24:1-2) · `missionary-law:cursing-as-testimony` (24:15-16)
**Edges:**
- `★ strength-is-calling-bounded` (24:9) —[power-supplied-WITH-the-stewardship]→ honor_scope; presiding covenant · —[complement]→ §10:4 (don't run faster than your strength)
- `comfort:afflictions-promised-not-removed` (24:8) —[presence-not-exemption: "I am with thee"]→ Matt 28:20 · —[forward]→ §121-122 (Liberty Jail)
- `pattern:double-nevertheless` (24:1-2) —[deliverance-doesn't-excuse-sin + sin-doesn't-cancel-the-call]→ §3 / §121:43
- `missionary-law:cursing-as-testimony` (24:15-16) —[dust-of-feet = witness left; smiting reserved to God "in mine own due time"]→ force not the elders' to wield (presiding restraint)

## D&C 25 — an elect lady (Harmony, July 1830, Emma Smith)
**Nodes:** `standing:sons-and-daughters` (25:1) · `★ discipline:trust-the-withheld` (25:4) · `★ doctrine:song-of-the-righteous-is-a-prayer` (25:12) · `calling:gifts-and-marriage-as-one-stewardship` (25:5-14) · `★ universalizing:this-is-my-voice-unto-all` (25:16)
**Edges:**
- `standing:sons-and-daughters` (25:1) —[equal-standing-opens-the-womans-revelation]→ John 1:12; elect-lady (2 John 1:1) → Relief Society fulfillment
- `★ discipline:trust-the-withheld` (25:4) —[the-reserved-knowledge-thread]→ §10:37; PoGP facsimiles figs-8-21; Michael's period-language reflection · —[faith-trusts-what's-withheld]→ meekness vs murmur [flag: temple, Michael's]
- `★ song-of-the-righteous-is-a-prayer` (25:12) —[singing-IS-prayer, answered-with-blessing]→ Ps 33:3; Eph 5:19 · —[Emma's-stewardship]→ the 1835 hymnal
- `calling:gifts-and-marriage-as-one-stewardship` (25:5-14) —[comfort + expound + delight]→ refuses the false choice (own calling vs marriage)
- `★ universalizing:this-is-my-voice-unto-all` (25:16) —[particular-words-declared-universal]→ inverts §15-16 (same-words-by-repetition) · —[personal+universal-collapse]→ D&C 1:38

## D&C 26 — by common consent (Harmony, July 1830)
**Nodes:** `★ law:common-consent` (26:2) · `consent:by-prayer-and-faith` (26:2) · `guidance:incremental` (26:1)
**Edges:**
- `★ law:common-consent` (26:2) —[authority-exercised-by-agreement-not-command]→ neither democracy nor autocracy · —[=]→ Mosiah 29:26; §20:65 · —[presiding]→ D&C 121:41 (governing without compulsory means)
- `consent:by-prayer-and-faith` (26:2) —[distributed-spiritual-confirmation]→ §9 (study-it-out-then-confirm) scaled to a church
- `guidance:incremental` (26:1) —[next-step + "then it shall be made known"]→ §25:4 (withheld); §1:25-28 (line upon line)

## D&C 27 — the gathering of the keys + the whole armor of God (Harmony, Aug 1830)
**Nodes:** `sacrament:remembrance-not-substance` (27:2) · `★ priesthood:the-gathering-of-all-keys` (27:5-13) · `★ sealing:Elijah's-turning-of-hearts-keys` (27:9) · `★ armor:Ephesians-6-in-Restoration-voice` (27:15-18) · `supper:the-future-messianic-feast` (27:5)
**Edges:**
- `sacrament:remembrance-not-substance` (27:2) —[eye-single + remembering]→ §4:5 / 88:67 (single eye, matter-spectrum) · —[water-for-wine IS the doctrine: substance adiaphora]→ §20:77 (covenant)
- `★ priesthood:the-gathering-of-all-keys` (27:5-13) —[every-dispensation's-key-holder]→ Moroni/Elias/John/Elijah/patriarchs/Adam/Peter-James-John · —[the-convergence-of]→ PoGP line (Moses 6:7 / Abr 1:3) → §13 (re-splice) → §27 (gathered) · —[why-Restoration-not-revival]→ "gather together in one all things" (Eph 1:10); §10 wise-purpose redundancy at all-history scale
- `★ sealing:Elijah's-turning-of-hearts-keys` (27:9) —[=]→ Mal 4:5-6 (= §2; JS—H 1:38) · —[the-Malachi-bracket's-keystone]→ §2 ↔ §13 ↔ 27:9 · —[family-sealed-not-disconnected]→ "whole earth not smitten with a curse" [temple, Michael's]
- `★ armor:Ephesians-6-in-Restoration-voice` (27:15-18) —[=]→ Eph 6:11-17 · —[the-sword = revelation]→ "sword of my Spirit… and my word which I reveal" (§8:2) · —[defensive/standing]→ "able to STAND" · —[pieces=walk's-threads]→ truth/faith/Spirit-word
- `supper:the-future-messianic-feast` (27:5) —[Christ-drinks-with-all-dispensations-on-the-earth]→ Matt 26:29 · —[sacrament-is-rehearsal]→ §20:75
- ✦ **the priesthood-line + Malachi-sealing threads CONVERGE here** (every key gathered; Elijah's keys named)
