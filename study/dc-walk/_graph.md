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

## D&C 28 — the order of revelation (Fayette, Sept 1830; Hiram Page's seerstone)
**Nodes:** `★ epistemology:order-as-the-discernment-cross-check` (28:2,11) · `★ contrast:Hirams-stone-vs-Olivers-rod` (28:11 ↔ §8) · `★ correction:private-persuasion-not-public-shaming` (28:11) · `order:channel-AND-consent` (28:2,13) · `succession:another-in-his-stead` (28:7)
**Edges:**
- `★ order-as-the-discernment-cross-check` (28:2,11) —[answers "is it God or me?" STRUCTURALLY]→ §8:2 / §9:8 (felt-signal necessary-not-sufficient) · —[true-distillation-respects-stewardship-bounds; out-of-order = counterfeit's signature]→ the period-language principle's GUARDRAIL
- `★ contrast:Hirams-stone-vs-Olivers-rod` (28:11 ↔ §8:6-8) —[same-folk-instrument, opposite-verdict]→ rod CLAIMED (§8) / stone REJECTED (§28) · —[discriminator = order/stewardship, NOT the object]→ same gift in-order=of-God / out-of-order=deception
- `★ correction:private-persuasion-not-public-shaming` (28:11) —[=]→ Matt 18:15 ("between him and thee alone") · —[Oliver's-role-bounded]→ 28:5-6 · —[presiding]→ D&C 121:41-43
- `order:channel-AND-consent` (28:2,13) —[one-appointed + the-body-confirms]→ safeguard vs anarchy AND tyranny · —[Hiram-fails-both]→ §26
- `succession:another-in-his-stead` (28:7) —[the-channel-is-an-office-not-a-cult]→ the order outlasts the man

## D&C 29 — all things are spiritual (Fayette, Sept 1830; Creation/Fall/agency LANDMARK)
**Nodes:** `★ matter-spectrum:all-things-are-spiritual` (29:34-35) · `★ structure:spiritual→temporal→spiritual-chiasm` (29:31-32) · `★ agency:requires-opposition` (29:39) · `★ rebellion:over-agency` (29:36-37) · `★ death:first=last-both-spiritual` (29:41) · `★ children:redeemed-from-the-foundation` (29:46-47) · `renewal:nothing-lost` (29:25) · `Christ:advocate` (29:5) · `elect:hear-his-voice` (29:7)
**Edges:**
- `★ all-things-are-spiritual` (29:34-35) —[the-spiritual/temporal-divide-is-ours-not-God's]→ truth.md / D&C 93/131 (matter = refined spirit) from the law side · —[every-temporal-law-is-spiritual]→ Word of Wisdom etc. change the spirit
- `★ spiritual→temporal→spiritual-chiasm` (29:31-32) —[creation: spirit→matter / redemption: matter→spirit]→ Moses 3:5 · —[plan-as-narrative]→ spirit descends into matter, returns glorified; §27 "gather in one" = the return
- `★ agency:requires-opposition` (29:39) —[2-Ne-2-in-the-Lord's-voice: "if they never should have bitter they could not know the sweet"]→ the tempter NECESSARY to agency · —[bounds-the-problem-of-evil]→ cost of a world that can produce gods; = §19 economy
- `★ rebellion:over-agency` (29:36-37) —[devil-sought-God's-honor; third-part-followed "because of their agency"]→ Moses 4:1-4; Abr 3:27; Rev 12:4 · —[irony]→ rejecting-the-agency-plan was-itself-an-act-of-agency
- `★ death:first=last-both-spiritual` (29:41) —[the-Fall (separation) = hell ("Depart," 29:28)]→ mortality = the window to escape the first becoming the last
- `★ children:redeemed-from-the-foundation` (29:46-47) —[Atonement-covers-them-before-they-can-choose; Satan-barred-until-accountability]→ Moroni 8 (BoM walk); D&C 68:25 (age 8) · —[the-§19-mercy-for-children: infant-damnation dissolved]
- `renewal:nothing-lost` (29:25 "not one hair, neither mote") —[matter-conserved-and-transfigured-not-annihilated]→ §10 nothing-lost at cosmic scale
- `elect:hear-his-voice` (29:7) —[elect-defined-by-hearing/not-hardening, not-arbitrary-predestination]→ John 10:27

## D&C 30 — fearing man, and not fearing man (Fayette, Sept 1830; the three Whitmer brothers)
**Nodes:** `★ application:§28-cross-check-applied-personally` (30:2, David) · `★ thread:fear-of-man-bookends` (30:1 ↔ 30:11) · `reproof:calibrated-to-cure` (30:3-4) · `order:authority-chain-in-companionship` (30:7)
**Edges:**
- `★ §28-cross-check-applied-personally` (30:2) —[David "persuaded by those whom I have not commanded"]→ §28:11 (Hiram Page, out of order) · —[breached-BOTH-halves]→ didn't-heed-the-appointed + heeded-the-unappointed · —[the-affective-root]→ fear (30:1) + worldliness (30:2) make a heart persuadable by counterfeits (the DEVOTIONAL guard on the period-language principle)
- `★ thread:fear-of-man-bookends` (30:1 ↔ 30:11) —[David feared man / John "not fearing"]→ §3:7 (Joseph's same sin) · —[cured-by-Presence]→ "I am with you" (Matt 28:20; §24:8) — not courage summoned but with-ness relied on
- `reproof:calibrated-to-cure` (30:3-4) —[consequence-fits-the-fault: "left to inquire for yourself"]→ door-not-shut ("until I give… further commandments") · —[=]→ §9 / §23
- `order:authority-chain-in-companionship` (30:7) —[Peter heeds Oliver; Oliver answers only to Joseph]→ §28 order scaled to two men on the road
- ✦ **DECADE 21-30 CLOSED:** the Church finds its feet — order of revelation (§28) + common consent (§26) + the great panorama (§27 keys, §29 Creation/agency); the §28→§30 order-cross-check arc; missions launch (Lamanite mission, §28/30/32)

## D&C 31 — govern your house in meekness (Fayette, Sept 1830, Thomas B. Marsh)
**Nodes:** `★ presiding:govern-your-house-in-meekness` (31:9) · `warning:given-early-at-the-call` (31:9) · `promise:family-faith-comes` (31:2) · `antidote:I-am-with-you` (31:13)
**Edges:**
- `★ presiding:govern-your-house-in-meekness` (31:9) —[=]→ D&C 121:41-42 (no compulsion, only persuasion/meekness) — the presiding covenant in the HOME, its MOST PERSONAL form · —[paired]→ "revile not… physician not combatant" (31:9-10)
- `warning:given-early-at-the-call` (31:9, Marsh's meekness) —[the-fault-line-named-at-the-start]→ §23:1 (Oliver's pride) · —[Marsh's-1838-fall-began-in-a-household-quarrel; returned 1857]→ [bin-4-adjacent biography] · —[grace-outlasts-the-fall]→ 31:2
- `promise:family-faith-comes` (31:2 "nevertheless… the day cometh that they will believe")→ the `nevertheless` hinge
- `antidote:I-am-with-you` (31:13) —[cure-for-affliction/fear = Presence]→ §24:8 / §30:11 (Matt 28:20)

## D&C 32 — the Lamanite mission party completed (Manchester, Oct 1830; Parley P. Pratt + Ziba Peterson)
**Nodes:** `★ order:pretend-to-no-other-revelation` (32:4) · `with-ness:I-will-go-in-their-midst` (32:3) · `posture:meek-and-lowly` (32:1)
**Edges:**
- `★ order:pretend-to-no-other-revelation` (32:4) —[=]→ D&C 28:2 (the order, on the road) · —[declare-the-WRITTEN-word]→ revelation = the Spirit *unfolding* it (§9 study-and-ask), not adding
- `with-ness:I-will-go-in-their-midst` (32:3) —[strongest-early-form]→ §24/§30/§31 · —[as-advocate]→ D&C 29:5; 1 John 2:1 (Christ goes WITH them, not sends them to God)
- `posture:meek-and-lowly` (32:1) —[=]→ Matt 11:29 · —[power-to-declare-rooted-in-lowliness]→ opposite of priestcraft (§33:4)

## D&C 33 — open your mouths and they shall be filled (Fayette, Oct 1830; Ezra Thayre + Northrop Sweet)
**Nodes:** `★ promise:open-your-mouths-and-be-filled` (33:8,10) · `apostasy:priestcraft-as-the-engine` (33:4) · `urgency:eleventh-hour` (33:3) · `readiness:lamps-and-the-Bridegroom` (33:17-18)
**Edges:**
- `★ open-your-mouths-and-be-filled` (33:8,10) —[the-filling-FOLLOWS-the-opening: speak-in-faith, made-ready-in-the-act]→ D&C 24:6; Ex 4:12 · —[spoken-form-of §8:2; the §32:4 guard holds]→ the filling unfolds the WRITTEN word
- `apostasy:priestcraft-as-the-engine` (33:4) —[religion-for-gain/status]→ 2 Ne 26:29 (BoM walk) · —[counter = meekness]→ §31:9 / §32:1
- `urgency:eleventh-hour` (33:3) —[the-LAST-dispensation]→ Matt 20:1-16; Jacob 5:71
- `readiness:lamps-and-the-Bridegroom` (33:17-18) —[oil = Spirit, can't-be-borrowed]→ Matt 25:1-13 · —[§27-supper-from-the-watching-side]

## D&C 34 — believed, and called (Fayette, Nov 4 1830; Orson Pratt, 19)
**Nodes:** `★ ladder:believed-blessed / called-more-blessed` (34:4-5) · `★ John-3:16-personalized` (34:3) · `refrain:light-which-shineth-in-darkness` (34:2) · `antidote:I-am-with-you-until-I-come` (34:11-12)
**Edges:**
- `★ ladder:believed-blessed / called-more-blessed` (34:4-5) —[giving-it-away > keeping-it]→ §15-16/§18 (worth of souls) personalized · —[blessed-believed]→ John 20:29
- `★ John-3:16-personalized` (34:3) —[universal-made-particular: "wherefore YOU are my son"]→ §25:16 · —[become-sons]→ John 1:12; §11:30 / §25:1 (real transformation, §29)
- `refrain:light-which-shineth-in-darkness` (34:2) —[the-Oliver-cluster-signature]→ §6:21 / 10:58 / 11:11; John 1:5 · —[source-light]→ §14:9
- `antidote:I-am-with-you-until-I-come` (34:11-12) —[Presence + a-clock ("I come quickly")]→ §24/§30/§31/§32 · —[the-"quickly"-held-over-Orson's-51-years]

## D&C 35 — the scriptures as they are in mine own bosom (Fayette, Dec 7 1830, Sidney Rigdon; JST scribe)
**Nodes:** `★ scripture:as-they-are-in-mine-own-bosom` (35:20) · `★ recognition:forerunner-life-reframed` (35:3-4) · `★ method:weak-things-thresh-the-nations` (35:13-14) · `watching:over-the-prophet-himself` (35:19) · `oneness:that-we-may-be-one` (35:2)
**Edges:**
- `★ scripture:as-they-are-in-mine-own-bosom` (35:20, the JST) —[recovery-toward-the-divine-source-form]→ the REVERSE of §1:24 (descent into language) · —[period-language-principle: two-directions-on-one-axis]→ read-in-period-language + restore-toward-God's-bosom both seek the word as-God-means-it
- `★ recognition:forerunner-life-reframed` (35:3-4 "thou knewest it not") —[§6-retrospective-recognition at BIOGRAPHICAL scale]→ Sidney's whole restorationist past = the runway; the Spirit distills a life's meaning backward
- `★ method:weak-things-thresh-the-nations` (35:13-14) —[instrument-weak + divine-supply: "their arm shall be my arm"]→ 1 Cor 1:27; D&C 1:19,23 · —[the-weakness-is-the-condition-God-prefers]→ glory His; the harness/supply carries the work
- `watching:over-the-prophet-himself` (35:19) —[even-the-keys-holder-needs-watching]→ Abr 4:18 (watch runs both ways) · —[paired-with-conditional-keys]→ 35:18 ("another will I plant in his stead")
- `oneness:that-we-may-be-one` (35:2) —[becoming-sons → UNITY-with-God]→ John 17:21; §11:30/§34:3 · —[matter-spectrum-destination]→ §29 → §38 (the horizontal demand "be one")

## D&C 36 — every man who embraces it, sent forth (near Fayette, Dec 9 1830, Edward Partridge)
**Nodes:** `★ discernment:the-Spirit-teaches-peaceable-things` (36:2) · `★ ministry:the-open-lay-door` (36:7) · `repentance:thorough-break` (36:6) · `temple:suddenly-come-to-my-temple` (36:8)
**Edges:**
- `★ discernment:the-Spirit-teaches-peaceable-things` (36:2) —[the-Comforter's-curriculum = PEACE, not-spectacle]→ §6:23 / §19:23 (the peaceable signature) · —[agitation/contention = different-signature]→ the discernment thread
- `★ ministry:the-open-lay-door` (36:7 "every man… with singleness of heart") —[only-gate = singleness, not-credentials]→ §4:5 (eye single) · —[§35:13-weak-things-institutionalized]→ a lay priesthood vs priestcraft (§33:4)
- `temple:suddenly-come-to-my-temple` (36:8) —[=]→ Mal 3:1 · —[§27/§35-Elijah/temple-thread]

## D&C 37 — gather to the Ohio (near Fayette, Dec 1830; the FIRST gathering command)
**Nodes:** `★ gathering:becomes-physical-geographic` (37:3) · `gathering:as-garrison-refuge` (37:1) · `★ agency:choose-for-yourself-even-in-obedience` (37:4) · `pastoral:strengthen-before-you-leave` (37:2)
**Edges:**
- `★ gathering:becomes-physical-geographic` (37:3) —[the-§29-promise-gets-boots-and-wagons]→ §29:7-8 / §33:6 (abstract) → §37:3 (relocate) · —[spiritual→temporal logic of §29:32 enacted]→ Zion is a PLACE (→ §38, §57)
- `gathering:as-garrison-refuge` (37:1 "because of the enemy and for your sakes") —[draw-the-laborers-in-from-the-exposed-field]→ §29:8 · —[the-gathering-IS-the-wall]→ presiding-protection (walls lawful)
- `★ agency:choose-for-yourself-even-in-obedience` (37:4) —[divine-command + free-choice]→ §26 (common consent); D&C 121:41; §29:39 · —[the-Restoration's-central-physical-act-is-voluntary]
- `pastoral:strengthen-before-you-leave` (37:2) —[don't-abandon-the-scattered-branch]→ §20:53 (watchcare) in the transition

## D&C 38 — be one, or ye are not mine (Fayette, Jan 2 1831; the conference; LANDMARK)
**Nodes:** `★ Zion:be-one = economic-equality` (38:24-27) · `★ matter-spectrum:the-eternal-present` (38:2) · `★ endowment:first-word` (38:32) · `★ threat:the-enemy-in-secret-chambers-at-home` (38:13,28-29) · `theocracy:God-as-king/lawgiver/watcher` (38:21-22) · `warning:beware-pride-as-the-Nephites` (38:39)
**Edges:**
- `★ Zion:be-one = economic-equality` (38:24-27) —[twelve-sons-parable: robes-vs-rags = injustice]→ NOT mere agreement but equality (38:16,35) · —[belonging-conditioned]→ "if ye are not one ye are NOT MINE" · —[completes-oneness-thread]→ §35:2 (one with God) + §38:27 (one with each other) = ONE fabric · —[Enoch-Zion "no poor among them"]→ Moses 7:18 → consecration (§42, §51)
- `★ matter-spectrum:the-eternal-present` (38:2) —[all-things-present-before-His-eyes, not-foreknowledge-but-an-eternal-now]→ Moses 1:6; D&C 130:7 · —[grounds-foreknowledge-without-erasing-agency: He SEES present, doesn't CAUSE]· —[omniscience-as-watchful-care]→ "mine eyes are upon you" (38:7)
- `★ endowment:first-word` (38:32 "endowed with power from on high") —[=]→ Luke 24:49; D&C 110:9 · —[the-gathering's-purpose: law (§42) + endowment]→ reserved temple thread opens [Michael's] · —[§27 keys / §13 sons-of-Levi converge]
- `★ threat:enemy-in-secret-chambers-at-home` (38:13,28-29) —[secret-combinations turned-on-the-present]→ Ether 8:18-25 (BoM walk) · —[part-of-why-gather: scattered=exposed, gathered+one=fortified]→ §37 garrison
- `theocracy:God-as-king/lawgiver/watcher` (38:21-22) —[presiding/watching at DIVINE scale]→ Abr 4:18 (top of the chain) · —[kingship-that-FREES]→ "a free people"; opposite of tyranny (§26)
- `warning:beware-pride-as-the-Nephites` (38:39) —[riches-test]→ 4 Ne / Hel (pride-cycle, BoM walk) · —[equality = the structural guard against wealth-pride-fall]

## D&C 39 — prepared for a greater work (Fayette, Jan 5 1831, James Covel)
**Nodes:** `affirmation:heart-right-NOW-at-this-time` (39:8) · `★ pattern:fault-named-at-the-call` (39:9) · `offer:prepared-for-a-greater-work` (39:11) · `irony:peace-offered-fear-chosen` (39:6 ↔ 40:2)
**Edges:**
- `affirmation:heart-right-NOW` (39:8) —[genuine-but-time-stamped, not-fixed-sainthood]→ qualified-by-history (39:9) · —[a-heart-right-now-may-not-stay-so]→ §40
- `★ fault-named-at-the-call` (39:9 pride + cares-of-the-world) —[=]→ §23:1 (Oliver) / §31:9 (Marsh) · —[EXACTLY-what-undid-him]→ §40:2 · —[warning ≠ compulsion]→ agency; God's foresight (38:2) doesn't override choice
- `offer:prepared-for-a-greater-work` (39:11) —[the §35:3-Sidney-offer]→ but §40: Sidney-received / Covel-rejected — same offer, opposite responses
- `irony:peace-offered-fear-chosen` (39:6 ↔ 40:2) —[peaceable-things (§36:2) traded for fear-of-persecution]→ §30 (fear-of-man)

## D&C 40 — the word received with gladness, then rejected (Fayette, Jan 6 1831; the §39 sequel)
**Nodes:** `★ parable-of-the-sower:enacted-as-history` (40:2) · `★ covenant:broken` (40:1,3) · `judgment:reserved-to-God` (40:3) · `meta:include_failures-at-scripture-scale` (§39-40)
**Edges:**
- `★ parable-of-the-sower:enacted` (40:2) —[narrated-in-Matt-13's-own-terms: gladness + straightway + persecution + cares-of-the-world]→ Matt 13:20-22 · —[the-gladness-is-NOT-the-test]→ joy-without-root collapses
- `★ covenant:broken` (40:1,3) —[entered-and-broke, not-merely-changed-his-mind]→ D&C 82:10 · —[dissolves-the-promise]→ 39:10
- `judgment:reserved-to-God` (40:3 "it remaineth with me") —[failure-recorded, sentence-withheld]→ §10:37 / §29:30 · —[don't-presume-the-eternal-verdict]
- `meta:include_failures-at-scripture-scale` (§39-40) —[two-revelations-on-a-man-who-failed]→ §3 / §20:5 / §30 (the canon refuses a highlight reel) → covenant.yaml teaching ext
- ✦ **DECADE 31-40 CLOSED:** the gathering to Ohio (§37-38) + Zion as oneness/equality (§38:27) + the endowment's first word (38:32); the early missions (§31-36); the §39-40 sower-enacted pair (include_failures). FOUR decades done (1-40).

## D&C 41 — the disciple is the doer (Kirtland, Feb 4 1831; the first bishop)
**Nodes:** `★ discipleship:the-doer-not-the-professor` (41:5) · `★ office:first-bishop-chosen-for-guilelessness` (41:11) · `polarity:bless/curse-by-hearing` (41:1)
**Edges:**
- `★ discipleship:the-doer-not-the-professor` (41:5) —[=]→ James 1:22; Matt 7:24-27 · —[the-value-is-in-the-doing]→ covenant; §40 (Covel's un-done gladness) · —[bridge-to-§42]→ a law is for DOING
- `★ office:first-bishop-chosen-for-guilelessness` (41:11) —[the-man-who-handles-Zion's-wealth = "no guile" (Nathanael, John 1:47)]→ character-is-the-security-on-the-consecration-system (§42:31-34) · —[anti-priestcraft]→ §33:4
- `polarity:bless/curse-by-hearing` (41:1) —[symmetrical, turns-on-hearkening-and-doing]→ §29:7 (elect hear)

## D&C 42 — the Law (Kirtland, Feb 9 & 23 1831; "the law of the Church"; LANDMARK)
**Nodes:** `★ law-of-consecration:stewardship-under-covenant` (42:30-34) · `★ teaching:if-ye-receive-not-the-Spirit-ye-shall-not-teach` (42:14) · `★ healing:faith-WITHIN-God's-appointment` (42:43-52) · `★ offenses:Matt-18-private-and-proportionate` (42:88-92) · `moral-law:Decalogue-renewed` (42:18-29) · `promise:revelation-upon-revelation / peaceable-things` (42:61)
**Edges:**
- `★ law-of-consecration:stewardship-under-covenant` (42:30-34) —[consecrate-all (deed-unbreakable) → steward-your-need → residue-to-the-poor]→ NOT communism (stewardship kept) / NOT capitalism (no accumulation) · —[the §38:27 "be one" economy]→ Moses 7:18 (Enoch "no poor"); Acts 4:32 · —[the-poor-ARE-Christ]→ Matt 25:40 (42:38) · —[=project's-stewardship-pattern writ large]→ received-not-owned, accountable, surplus-serves (covenant.yaml)
- `★ teaching:if-ye-receive-not-the-Spirit-ye-shall-not-teach` (42:14) —[the-teaching-covenant made LAW]→ obtain-before-declare (§11:21); discovery-not-performance; the Ben Test · —[don't-teach-without-it, forbidden-not-merely-suboptimal]· —[the-Spirit = the-peaceable-one]→ §36:2 (fruit-test)
- `★ healing:faith-WITHIN-God's-appointment` (42:43-52) —[faith AND appointment, not-a-mechanism]→ "not appointed unto death" (42:48; Alma 12:27; D&C 121:25) · —[against-faith-triumphalism]→ for the appointed, death "shall be sweet" (42:46) · —[humane: herbs+medicine+love]→ 42:43 · —[=§9 faith-and-confirmation]→ result confirmed/withheld by God's will, not compelled
- `★ offenses:Matt-18-private-and-proportionate` (42:88-92) —[=]→ Matt 18:15-17; §28:11 (Hiram Page) · —[discipline-MATCHED-to-scope: secret→secret (92), open→open (91)]· —[aim = reconciliation/confession NOT exposure; protects the offender's name]→ §121:43 (reprove-then-love)
- `moral-law:Decalogue-renewed` (42:18-29) —[=]→ Ex 20; Matt 5:21-30 · —[sharpenings]→ lust→denies-the-faith (42:23); fidelity-as-positive-command (42:22)
- `promise:revelation-upon-revelation / peaceable-things` (42:61) —[line-upon-line as ASCENDING gift]→ §1:25-28 · —[true-revelation = joy/peace]→ §36:2 / §39:6 (the discernment refrain)

## D&C 43 — all the day long, but ye would not (Kirtland, Feb 1831; false revelators recur)
**Nodes:** `★ order:standing-law-against-deception` (43:5-6) · `★ agency:all-the-day-long-but-ye-would-not` (43:24-25) · `★ teaching:taught-from-on-high → endowed → teach` (43:15-16) · `posture:solemnities-of-eternity` (43:34-35)
**Edges:**
- `★ order:standing-law-against-deception` (43:5-6) —[§28 generalized because the problem recurred]→ "that you may not be DECEIVED" · —[a-counterfeit-exposed-by-POSITION not-by-out-arguing]→ deception's defense is order (the §28 guardrail codified)
- `★ agency:all-the-day-long-but-ye-would-not` (43:24-25) —[God's-persistence TOTAL (every voice, catastrophic→tender) and STILL doesn't compel]→ §29:39; Matt 23:37 · —[the-calling-is-LOVE → the-refusal-WOUNDS-God]· —[the-two-word-monument-to-agency]
- `★ teaching:taught-from-on-high → endowed → teach` (43:15-16) —[the §42:14-teaching-law given-its-source]→ teach-from-what-heaven-taught-YOU · —[endowment]→ Luke 24:49; §38:32 · —[sanctify→endow→teach]
- `posture:solemnities-of-eternity` (43:34-35) —[weighty-sobriety, against-levity]→ D&C 88:121 · —[joy-AND-gravity]→ §19:39

## D&C 44 — assemble, organize, care for the poor (Kirtland, late Feb 1831)
**Nodes:** `assembly:Spirit-poured-out-in-the-gathering` (44:2) · `★ lawful-walls:organize-under-the-laws-of-man` (44:4-5) · `★ poor:the-condition-of-the-law's-fulfillment` (44:6)
**Edges:**
- `assembly:Spirit-poured-out-in-the-gathering` (44:2) —[blessing-specific-to-convening]→ §43:8 / §41:2 (council pattern); §6:32
- `★ lawful-walls:organize-under-the-laws-of-man` (44:4-5) —[civil-incorporation = a-SHIELD-against-enemies]→ 1 Pet 2:13; D&C 98:5; 134 · —[theocracy (§38:21) + civil-lawfulness]→ the presiding covenant's "walls around the field" (lawful protection)
- `★ poor:the-condition-of-the-law's-fulfillment` (44:6) —[care-the-poor "that all things may be done according to my law"]→ §42:38 / §38:27 · —[neglecting-them = collapse-of-the-law's-purpose]→ James 1:27

## D&C 45 — the Advocate's prayer, and the place of safety (Kirtland, March 7 1831; LANDMARK; Olivet expanded)
**Nodes:** `★ Advocate:the-actual-plea` (45:3-5) · `★ discernment:the-wise-virgins = Spirit-guided + undeceived` (45:57) · `★ Zion:the-place-of-safety / only-non-warring-people` (45:66-71) · `wounds:as-identity` (45:51-52) · `pastoral:signs-as-reassurance-not-terror` (45:35) · `JST:translate-the-NT` (45:60-61)
**Edges:**
- `★ Advocate:the-actual-plea` (45:3-5) —[Christ-pleads-His-OWN-blood, not-our-merits]→ 1 John 2:1; §29:5 / §32:3 · —[our-standing = Christ's-wounds-presented-for-us]→ "spare these my brethren" · —[§19 first-person-Atonement (the cost) → §45 first-person-Advocacy (pleading the cost)]
- `★ discernment:the-wise-virgins = Spirit-guided + undeceived` (45:57) —[the-oil = the-Holy-Spirit-as-guide (accumulated, un-borrowable, §33:17)]→ Matt 25:1-13 · —[readiness = Spirit + NOT-deceived]→ §28/§43 (order against deception); §8:2/§9:8/§36:2 · —[the-period-language/Spirit-distillation-discipline IS the oil]
- `★ Zion:the-place-of-safety / only-non-warring-people` (45:66-71) —[the-gathering-as-refuge at-its-fullest]→ §37 · —[the-§38:27-be-one-people AS the-refuge: oneness = protection]· —[magnet = peace itself; unconquerable-in-peace]→ 45:68,70
- `wounds:as-identity` (45:51-52) —[recognized-BY-the-marks of-suffering]→ Zech 13:6 · —[the-risen-Christ-keeps-the-wounds-as-His-credential]→ §6:37; §19 (Atonement carried into glory)
- `pastoral:signs-as-reassurance-not-terror` (45:35) —[the-unraveling = fulfillment, for-the-faithful]→ "be not troubled"; stand-in-holy-places (45:32)
- `JST:translate-the-NT` (45:60-61) —[the-Bible-re-translation continues]→ §35:20 ("mine own bosom")

## D&C 46 — to every man is given a gift (Kirtland, March 8 1831; the GIFTS / LANDMARK)
**Nodes:** `★ gifts:universal-gifting = a-gift-economy` (46:11-12) · `★ gifts:the-protection-against-deception` (46:7-8) · `★ prayer:asketh-in-the-Spirit = God's-will` (46:30) · `inclusion:welcome-the-seeker` (46:3-6) · `meeting:Spirit-governs-the-form` (46:2)
**Edges:**
- `★ gifts:universal-gifting = a-gift-economy` (46:11-12) —[to-EVERY-man-a-gift, none-has-all (except the head 46:29)]→ interdependence · —[=]→ 1 Cor 12:7; Moroni 10 · —[the §42-consecration-logic on spiritual goods]· —[§38:27 be-one = COMPLEMENTARITY not uniformity]
- `★ gifts:the-protection-against-deception` (46:7-8) —[three-sources: God/men/devils]→ "seek the best gifts… that ye may not be deceived" · —[the-discernment-gifts]→ operations-whether-of-God (46:16); discerning-of-spirits (46:23); leaders' discern-all (46:27, Abr 4:18 watch) · —[the-discernment-thread's-POSITIVE/gift-form]→ §28/§43 (structural order); §45:57 (stakes); the period-language-discipline as a GIFT to seek
- `★ prayer:asketh-in-the-Spirit = God's-will` (46:30) —[the-Spirit-shapes-the-question to-God's-will]→ Rom 8:26; D&C 50:29 · —[resolves]→ §9 / §42:14 / §42:48 (healing-if-not-appointed) · —[prayer = wanting-what-God-wants by-the-Spirit, not-bending-God]
- `inclusion:welcome-the-seeker` (46:3-6) —[don't-cast-out the-earnest, even-non-members; only-exclusion = unrepented-trespass]→ 3 Ne 18:22; §44:6 (open posture)
- `meeting:Spirit-governs-the-form` (46:2) —[the form serves the Spirit]→ §43:8 / §44:2 (council pattern; gifts manifest in the Spirit-led assembly)

## D&C 47 — keep a regular history (Kirtland, March 8 1831, John Whitmer the historian)
**Nodes:** `★ memory:history-keeping-as-sacred-infrastructure` (47:1,4) · `obedience:hesitancy-surrendered-through-the-order` (heading) · `biography:the-historian-who-kept-the-record-and-left`
**Edges:**
- `★ memory:history-keeping-as-sacred-infrastructure` (47:1,4) —[a-calling, by-REVELATION, written-BY-THE-COMFORTER]→ §21:1 / §69 / §85 · —[the-project's `.mind/` + covenant's update_memory]→ §47 = the scriptural warrant (revelation-people = record-people) · —[Spirit-written = testimony not chronicle]
- `obedience:hesitancy-surrendered-through-the-order` (heading) —[John "would rather not" → submits-through-Joseph-the-Seer]→ §28 (the order) · —[reluctance-honestly-surrendered = faithfulness]→ §21:5
- `biography:historian-who-kept-the-record-and-left` —[a-faithful-work outlasts-the-worker's-faithfulness]→ §40/§30 [bin-4-adjacent]

## D&C 48 — share the land, and wait for the place (Kirtland, March 10 1831)
**Nodes:** `★ consecration:share-the-land-at-the-doorstep` (48:2) · `★ withheld:gather-toward-an-unrevealed-Zion` (48:5) · `order:family-by-family-gathering` (48:6)
**Edges:**
- `★ consecration:share-the-land-at-the-doorstep` (48:2) —[the-haves-impart-to-the-landless-arriving-brother]→ §42:30-34; §38:27 · —[Zion-built-first-by-hospitality]→ §44:6
- `★ withheld:gather-toward-an-unrevealed-Zion` (48:5) —[the-place "not yet to be revealed"]→ §25:4 (discipline of the withheld) applied to the MOST consequential unknown · —[act-on-a-withheld-destination in-faith]· —[revealed-through-the-ORDER]→ §28; §57 (Independence)
- `order:family-by-family-gathering` (48:6) —[ordered-settlement by-presidency-and-bishop]→ §42/§44 (the be-one people ORGANIZED)

## D&C 49 — marriage is ordained of God (Kirtland, May 7 1831; the Shakers)
**Nodes:** `★ marriage:ordained-of-God = the-earth's-purpose` (49:15-17) · `★ creation:use-not-waste + the-sin-of-inequality` (49:18-21) · `date:unknown-even-to-angels` (49:7) · `convert:teach-the-received-not-the-syncretic-mix` (49:4)
**Edges:**
- `★ marriage:ordained-of-God = the-earth's-purpose` (49:15-17) —[against-renunciation/celibacy]→ 1 Tim 4:3; Gen 2:24 · —[earth "answers the end of its creation" through family + foreordained "measure of man before the world was made"]→ matter-spectrum anthropology (premortal spirits awaiting bodies, §29/§38:2) · —[Zion = families not celibates]→ §38:27/§42
- `★ creation:use-not-waste + the-sin-of-inequality` (49:18-21) —[animals-for-use, but "wo… that wasteth flesh and hath no need"]→ stewardship not exploitation · —["not given that one man possess above another, wherefore the world lieth in sin"]→ the §42-consecration-ethic on the earth (Acts 4:32); inequality = the world's sin
- `date:unknown-even-to-angels` (49:7) —[rebuke to date-setting / "already occurred"]→ Matt 24:36 · —[discipline-of-the-withheld on-the-supreme-date]→ §25:4; "be not deceived" (49:23)
- `convert:teach-the-received-not-the-syncretic-mix` (49:4) —[reason-from-what-taught, not-from-the-Shaker-past]→ §32:4 · —[half-held-old-belief = the seam]→ Copley's later fall [bin-4-adjacent]

## D&C 50 — that which is of God is light (Kirtland, May 9 1831; the DISCERNMENT CAPSTONE / LANDMARK)
**Nodes:** `★ light:that-which-is-of-God-is-light` (50:23-24) · `★ test:mutual-edification` (50:17-22) · `★ protocol:ask-and-if-not-given-it-is-not-of-God` (50:31-32) · `★ period-language:God-reasons-face-to-face` (50:11-12) · `pastoral:little-children-grow-in-grace-fear-not` (50:40-42) · `oneness:the-Father-and-I-are-one` (50:43)
**Edges:**
- `★ light:that-which-is-of-God-is-light` (50:23-24) —[TEST: edification = light = of God; darkness = not]→ the discernment thread's CAPSTONE · —[GROWTH-LAW: light → more light → "brighter and brighter until the perfect day"]→ matter-spectrum destination (§35:2 oneness; §29; becoming-perfect) · —[the DISCERNMENT + MATTER-SPECTRUM threads FUSE]→ §28/§43/§45:57/§46 + §14:9/§29 (test-of-truth and law-of-growth = the SAME light) · —[period-language capstone: true distillation EDIFIES and GROWS; the fruit is light]
- `★ test:mutual-edification` (50:17-22) —[true-preacher (Spirit) + true-hearer (Spirit) → "understand one another, both edified and rejoice together"]→ the sign = SHARED light not intensity · —[completes]→ §42:14; §46:30 · —[edifies-the-body = true / puffs-up-or-confuses = suspect]
- `★ protocol:ask-and-if-not-given-it-is-not-of-God` (50:31-32) —[don't-receive → ask-the-Father → His-withholding = the-verdict]→ §9 (study-it-out) on spirits; 1 John 4:1-3 · —[humble-discernment]→ "not with railing… nor boasting, lest you be seized" (50:33); §28:11/§42:88
- `★ period-language:God-reasons-face-to-face` (50:11-12) —[=]→ D&C 1:24 (after the manner of their language) · —[God-condescends-to-the-human-level "that you may understand"]→ §35:20 (JST toward His bosom) — speaks DOWN so the light grows UP (50:24)
- `pastoral:little-children-grow-in-grace-fear-not` (50:40-42) —[discernment-ends-in-SECURITY not-anxiety]→ "ye cannot bear all things now… grow in grace"; "I have overcome the world"; "none… shall be lost" · —[a-light-you-grow-INTO, safe-meanwhile]
- `oneness:the-Father-and-I-are-one` (50:43) —[ye-are-in-me-and-I-in-you]→ §35:2 (John 17) — the destination of the growing light
- ✦ **DECADE 41-50 CLOSED:** the LAW (§42) + the gathering-as-refuge (§45) + the gifts (§46) + the DISCERNMENT CAPSTONE (§50 — light = the test AND the growth-law, fusing discernment + matter-spectrum). FIVE decades done (1-50).

## D&C 51 — every man equal according to his wants and needs (Thompson, May 20 1831; consecration in practice)
**Nodes:** `★ consecration:equal-by-need = equity-not-identity` (51:3,9) · `★ deed:protects-both-steward-and-poor` (51:4-5) · `★ withheld:act-upon-this-land-as-for-years` (51:16-17) · `steward:faithful-just-wise → the-joy-of-his-Lord` (51:19)
**Edges:**
- `★ consecration:equal-by-need = equity-not-identity` (51:3,9) —["according to his wants and needs"; "receive alike, that ye may be one"]→ §38:27 (be-one = equity); §46 (gift-economy) · —[not-uniformity, not-coercion: by-covenant under-the-guileless-bishop §41:11]
- `★ deed:protects-both-steward-and-poor` (51:4-5) —[steward-gets-a-legally-secured-deed (his) + the-consecrated-surplus-stays-for-the-poor-if-he-leaves]→ honors private-stewardship AND common-good · —[runs-THROUGH-civil-law]→ §44:4
- `★ withheld:act-upon-this-land-as-for-years` (51:16-17) —[unknown-duration → live-FULLY-in-the-present, build-as-if-permanent]→ §25:4 / §48:5 applied to TIME · —[the-sojourner-who-builds]→ Jer 29:5-7; §45:13 · —[don't-coast-because-temporary]
- `steward:faithful-just-wise → the-joy-of-his-Lord` (51:19) —[=]→ Matt 25:21-23 · —[property = a-training-in-stewardship]→ the project's stewardship covenant

## D&C 52 — a pattern, that ye may not be deceived (Kirtland, June 6 1831; the Missouri call)
**Nodes:** `★ discernment:the-PATTERN` (52:14-19) · `★ fan-out:build-not-on-another's-foundation` (52:33) · `discipleship:the-poor-are-the-mark` (52:40) · `realism:land-of-inheritance = land-of-enemies` (52:42)
**Edges:**
- `★ discernment:the-PATTERN` (52:14-19) —[§50's-principle made-a-CHECKABLE-RUBRIC: contrite + meek + edifieth + fruits-of-praise-and-wisdom + obeys-ordinances]→ "know the spirits in all cases under the whole heavens" · —[the-fruit-test (Matt 7:16; Moroni 7) as a 5-point rubric]· —[humility-markers: a true spirit LOWERS; the false aggrandizes]→ completes §28→§43→§45:57→§46→§50→§52
- `★ fan-out:build-not-on-another's-foundation` (52:33) —[each-laborer-on-his-own-TRACK, no-overlap]→ Rom 15:20 · —[coverage-by-non-overlap]→ the workspace's `fan-out` skill (fresh-ground-per-laborer)
- `discipleship:the-poor-are-the-mark` (52:40) —[neglect-the-poor → "NOT my disciple"]→ §41:5 / §42:38 / §44:6 (absolute, not-relative)
- `realism:land-of-inheritance = land-of-enemies` (52:42) —[the-promised-place not-yet-safe; won-through-opposition]→ §38:28 · —["I will hasten the city in its time"]

## D&C 53 — calling and election, beginning with forsaking the world (Kirtland, June 8 1831, Sidney Gilbert)
**Nodes:** `order:forsake-the-world-FIRST` (53:2) · `withheld:calling-unfolds-with-labor` (53:6) · `endurance:he-only-is-saved-who-endureth` (53:7)
**Edges:**
- `order:forsake-the-world-FIRST` (53:2) —[before-the-office: let-go-of-the-world]→ §49:20 · —[handle-the-kingdom's-money only-after-forsaking-its-worldliness]→ §41:11
- `withheld:calling-unfolds-with-labor` (53:6) —[the-residue "made known… according to your labor"]→ §25:4 on one's OWN calling; line-upon-line §1:25-28
- `endurance:he-only-is-saved-who-endureth` (53:7) —[not-the-calling-but-the-ENDURING]→ Matt 10:22; §5:22 / §14:7 · —[Gilbert endured to HIS end, d.1834]

## D&C 54 — the broken covenant, and mercy for the keepers (Kirtland, June 10 1831; Newel Knight)
**Nodes:** `★ covenant:broken → void-and-of-none-effect` (54:4) · `★ mercy:for-the-keeper-displaced-by-another's-breach` (54:6) · `interim:seek-a-living-like-unto-men` (54:9) · `comfort:patient-in-tribulation → rest` (54:10)
**Edges:**
- `★ covenant:broken → void-and-of-none-effect` (54:4) —[=]→ D&C 82:10 (no promise); §40 (Covel) · —[bilateral: one-side-breaks → binding-lapses]· —[the-breaker-who-displaces-the-innocent → the-MILLSTONE]→ 54:5 (Matt 18:6)
- `★ mercy:for-the-keeper-displaced-by-another's-breach` (54:6) —[the-betrayed-party still-blessed, NOT-punished-for-the-disruption]→ justice-tracks-the-HEART-not-the-circumstance · —[displaced-keep-their-standing; breaker-loses-his]
- `interim:seek-a-living-like-unto-men` (54:9) —[faithful-ordinariness while-the-place-isn't-ready]→ §51:16 in hardship; faith-sustains-THROUGH-the-day-labor
- `comfort:patient-in-tribulation → rest` (54:10) —[presence WITHIN tribulation not-exemption]→ §24:8 / §45:35 · —[sought-me-early → rest]→ Prov 8:17

## D&C 55 — the printer, and books for children (Kirtland, June 14 1831, W.W. Phelps)
**Nodes:** `★ kingdom:invests-in-the-press-and-the-young` (55:4) · `ordinance:the-single-eye-conditions-it` (55:1) · `ordinance:the-contrite-receiver` (55:3)
**Edges:**
- `★ kingdom:invests-in-the-press-and-the-young` (55:4) —[the-first-printer told-to-write-children's-books "that little children may receive instruction"]→ §47 (the record) → §55 (press + children's books) → §88 (School of the Prophets) · —[a-movement-commissioning-children's-books-before-a-city = the-long-game]
- `ordinance:the-single-eye-conditions-it` (55:1) —["if you do with an eye single to my glory, you shall have a remission"]→ §4:5 / §27:2 · —[pure-motive conditions-the-grace not-the-act-alone]
- `ordinance:the-contrite-receiver` (55:3) —[confer-the-Spirit "if they are contrite"]→ §50:19-22; §52:15-16 (authority-confers, contrition-receives)

## D&C 56 — wo to the hoarding rich, and wo to the covetous poor (Kirtland, June 15 1831; Ezra Thayre revoked)
**Nodes:** `★ greed:the-reciprocal-wo (rich AND poor)` (56:16-17) · `★ blessing:the-poor-who-are-PURE-IN-HEART` (56:18) · `revelation:command-and-revoke` (56:3-4) · `root:counsel-in-your-own-ways` (56:14)
**Edges:**
- `★ greed:the-reciprocal-wo (rich AND poor)` (56:16-17) —[rich-sin = HOARDING; poor-sin = COVETING + IDLENESS]→ the SAME sin (the unbroken covetous heart) in two costumes · —[CORRECTS the consecration thread]→ the poor NOT automatically righteous; equality = conversion-of-the-HEART not class-warfare · —[balances §49:20]
- `★ blessing:the-poor-who-are-PURE-IN-HEART` (56:18) —[NOT poverty-as-such, but broken+contrite (Matt 5:3 "poor in spirit")]→ inheritance goes-to-the-HEART not-the-bracket · —[the §52 contrite-pattern applied-to-economics]
- `revelation:command-and-revoke` (56:3-4) —[a-commandment-can-be-WITHDRAWN when-the-one-commanded-won't-obey]→ §3 / §10:2; §53:6 (dark side) · —[the-loss = "answered upon the heads of the rebellious"]
- `root:counsel-in-your-own-ways` (56:14) —[un-pardon-because they-run-it-their-own-way]→ §22:4 (seek not to counsel your God); Jacob 4:10 · —[self-will = the-engine under-both-hoarding-and-coveting]

## D&C 57 — the center place (Jackson County, Missouri, July 20 1831; ZION LOCATED)
**Nodes:** `★ gathering:Zion-LOCATED — the-center-place` (57:2-3) · `honest-commerce:sell-goods-without-fraud` (57:8) · `press:a-founding-institution` (57:11-13)
**Edges:**
- `★ gathering:Zion-LOCATED — the-center-place` (57:2-3) —[the-withholding ENDS at a real address (Independence; the temple lot "not far from the courthouse")]→ §48:5 / §25:4 DISCLOSED · —[the-heavenly-city onto-a-courthouse-lot]→ §29:32 (spiritual→temporal) · —[rewards-the-discipline-of-the-withheld]· —[realism: named not-yet-given; = §52:42's "land of your enemies" (lost 1833)]
- `honest-commerce:sell-goods-without-fraud` (57:8) —[kingdom's-trade held-to-righteousness; profit-serves-the-poor]→ §49:21 / §56 · —[the-marketplace-form-of "be one"]→ §38:27
- `press:a-founding-institution` (57:11-13) —[Zion's-first-institutions: printer + Spirit-checked-editor]→ §47 / §55 (a-covenant-people = a-publishing-people)

## D&C 58 — anxiously engaged: the power is in them (Jackson County, Aug 1 1831; AGENCY LANDMARK / PROJECT-SOURCE)
**Nodes:** `★ agency:it-is-not-meet-that-I-should-command-in-all-things` (58:26-28, PROJECT-SOURCE) · `★ order:after-much-tribulation-come-the-blessings` (58:3-4) · `★ repentance:confess + forsake → remember-no-more` (58:42-43) · `★ presiding:the-judge-is-NOT-a-ruler` (58:20) · `faith-failure:misread-the-revoking` (58:30-33)
**Edges:**
- `★ agency:it-is-not-meet-that-I-should-command-in-all-things` (58:26-28) —[PROJECT-SOURCE: the autonomy/stewardship root]→ `stuffy-in-the-loop` / `dave-rule` / `ammon` / `exercise_stewardship` · —["anxiously engaged… of their own free will"; "the power is in them, wherein they are agents unto themselves"]→ honor-intent-over-literal-command · —[cuts-BOTH-ways]→ over-commanded = slothful (58:26) AND passive-waiting-to-be-commanded = DAMNED (58:29) · —[the-steward HAS the-agency; the-failure-mode is PASSIVITY]
- `★ order:after-much-tribulation-come-the-blessings` (58:3-4) —[glory FOLLOWS tribulation]→ §24:8 / §54:10 deepened into a LAW OF SEQUENCE · —[tribulation = the ROAD not a detour]→ §19/§29 (bitter-before-sweet) as a timeline; → §121:7-8
- `★ repentance:confess + forsake → remember-no-more` (58:42-43) —[two-marks: name-it + leave-it; against-false-repentance]→ Isa 1:18; Jer 31:34 · —[God FORGETS the-forgiven-sin]→ the §19 Atonement's purpose realized
- `★ presiding:the-judge-is-NOT-a-ruler` (58:20) —[unrighteous-dominion-guard built-into-the-granting-of-authority]→ D&C 121:39-46 · —["let God rule him that judgeth"]· —[paired]→ obey-the-laws-of-the-land (58:21-22, Rom 13 / A of F 12 / §134)
- `faith-failure:misread-the-revoking` (58:30-33) —[God-conditioned-it; they-disobeyed; revoked; then-blame-God]→ §56:3-4 / §82:10 misread · —[the-passive-servant becomes-the-doubter]
