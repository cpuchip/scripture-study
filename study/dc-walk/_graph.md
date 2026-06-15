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

## D&C 59 — thank the Lord in all things, and the fulness of the earth (Jackson County, Aug 7 1831; the Sabbath; Polly Knight's death)
**Nodes:** `★ gratitude:thank-God-in-all-things / ingratitude-the-chief-offense` (59:7,21) · `★ earth:good — given-for-joy-and-the-senses` (59:16-20) · `Sabbath:joyful-sobriety` (59:9-15) · `worship:the-broken-heart-sacrifice` (59:8) · `comfort:works-follow-the-faithful-at-death` (59:2)
**Edges:**
- `★ gratitude:thank-God-in-all-things / ingratitude-the-chief-offense` (59:7,21) —[the-OFFENSE = confess-not-His-hand-in-ALL-things + disobey]→ ingratitude = blindness-to-the-Giver, the-root-sin · —["his hand in ALL things"]→ the §54/§58 tribulation included · —[gratitude = recognition-of-God regardless-of-circumstance]
- `★ earth:good — given-for-joy-and-the-senses` (59:16-20) —["to please the eye and gladden the heart… for taste and smell, to enliven the soul"]→ NOT just survival but DELIGHT · —[against-asceticism]→ §49 (marriage-ordained); §29 (all things spiritual) · —[the-physical is-GOOD; "with judgment, not to excess" (59:20)] · —[follows-from-gratitude: receive-the-good-as-gifts]
- `Sabbath:joyful-sobriety` (59:9-15) —[rest + worship + confession + REJOICING ("fasting… in other words, rejoicing")]→ keep-unspotted (James 1:27) · —[glad-NOT-flippant: "not with much laughter, for this is sin"]→ §43:34 (solemnities + cheer)
- `worship:the-broken-heart-sacrifice` (59:8) —[the-INNER-sacrifice replaces-the-Mosaic]→ 3 Ne 9:20 · —[the §52:15-16 contrite-pattern AS worship]
- `comfort:works-follow-the-faithful-at-death` (59:2) —[death = rest; "their works shall follow them"]→ Rev 14:13 · —[at-Polly-Knight's-grave: labors crowned-not-ended]→ §51:19

## D&C 60 — bury not the talent (Independence, Aug 8 1831; the elders' return)
**Nodes:** `★ talent:buried-from-fear-of-man` (60:2) · `★ agency:the-means-left-to-the-steward` (60:5) · `manner:bold-but-uncontentious` (60:7,14-15)
**Edges:**
- `★ talent:buried-from-fear-of-man` (60:2) —[§58:29's-passive-servant MADE PERSONAL: "will not open their mouths… hide the talent… because of the fear of man"]→ Matt 25:25 · —[fear-of-man = what-makes-a-steward-bury-the-talent]→ §3:7 / §30:1 · —[the-unused-gift FORFEITED: "taken away, even that which they have" (60:3, Matt 25:29)]→ passivity is SUBTRACTIVE
- `★ agency:the-means-left-to-the-steward` (60:5) —["as seemeth you good, it mattereth not unto me"]→ §58:26 in-a-concrete-case · —[the-irony: same-elders who-wouldn't-act-on-the-command are-now-given-a-thing-to-decide]→ act-on-the-command AND exercise-the-agency
- `manner:bold-but-uncontentious` (60:7,14-15) —[open-the-mouth (vs 60:2) but "without wrath… not in haste nor with strife"; testimony-against "in secret, lest thou provoke"]→ §52; §28:11 / §42:88
- ✦ **DECADE 51-60 CLOSED:** consecration-in-practice (§51) → Zion LOCATED (§57) → the AGENCY doctrine (§58, project-source) → the Sabbath/gratitude/earth-goodness (§59) → the talent buried-from-fear (§60). SIX decades done (1-60); nearly halfway.

## D&C 61 — mine anger is turned away (Missouri River, McIlwaine's Bend, Aug 12 1831; the destroyer on the waters)
**Nodes:** `★ God:anger-is-momentary` (61:20) · `providence:real-danger-but-the-faithful-held` (61:4-6,19) · `★ agency:it-mattereth-not (means-by-judgment)` (61:22) · `gift:command-the-elements — but-bridled-to-the-Spirit` (61:27-28) · `comfort:be-of-good-cheer / not-forsaken` (61:36)
**Edges:**
- `★ God:anger-is-momentary` (61:20) —["angry yesterday, but today mine anger is turned away"]→ §3 / §121:43 (reprove-then-love) at-a-DAILY-scale · —[a-small-moment]→ Isa 54:7-8; §121:7 · —[the-anger IS a-form-of-love, turned-away-when-chastening's-done]
- `providence:real-danger-but-the-faithful-held` (61:4-6,19) —[the-destroyer-rides (danger-NOT-removed, "I revoke not the decree") AND "the faithful shall not perish"]→ §24:8 / §54:10 in-PHYSICAL-peril · —[covenant-preservation, not-a-shield-over-the-reckless]→ "forewarn… lest their faith fail" (61:18)
- `★ agency:it-mattereth-not (means-by-judgment)` (61:22) —[water-or-land "according to their judgments"]→ §58:26 / §60:5 · —[God-REASONS not-dictates]→ 61:13 (period-language, §1:24/§50:11)
- `gift:command-the-elements — but-bridled-to-the-Spirit` (61:27-28) —[exercised "as the Spirit commandeth," not-at-will]→ §46 / §50 (power-real-but-NEVER-freelance)
- `comfort:be-of-good-cheer / not-forsaken` (61:36) —[after-danger-and-anger: presence + good-cheer]→ §50:40-42; §59:15

## D&C 62 — the advocate who knows weakness, testimony recorded in heaven (Missouri River at Chariton, Aug 13 1831)
**Nodes:** `★ Advocate:knows-weakness, succors-the-tempted` (62:1) · `★ heaven:testimony-recorded, angels-rejoice` (62:3) · `★ agency:it-mattereth-not — by-judgment-and-the-Spirit` (62:5,8) · `comfort:I-am-with-the-faithful-always` (62:9)
**Edges:**
- `★ Advocate:knows-weakness, succors-the-tempted` (62:1) —[the-§45:3-Advocate ALSO the-succorer]→ Heb 2:18; Alma 7:11-13 (took-infirmity to-know-HOW-to-succor) · —[pleading (§45) + succoring (§62) = the-same-Christ; the §19 first-person-Atonement makes-Him our-sympathizer]
- `★ heaven:testimony-recorded, angels-rejoice` (62:3) —[your-witness logged-in-the-book-of-life "for the angels to look upon"]→ §47 (record-keeping) OTHER-END; Mal 3:16 · —[nothing-faithful-is-unwitnessed; angels-REJOICE]→ Luke 15:7,10 · —[for-the-unseen-laborer: you-are-SEEN]
- `★ agency:it-mattereth-not — by-judgment-and-the-Spirit` (62:5,8) —[3rd-travel-section-running]→ §58:26 / §60:5 / §61:22 (the settled pattern) · —[principle-commanded (faithful), method-the-steward's] · —[+gratitude]→ "thankful heart in all things" (62:7, §59:7)
- `comfort:I-am-with-the-faithful-always` (62:9) —[presence-on-condition-of-faithfulness]→ Matt 28:20; §24/§61:36

## D&C 63 — signs come by faith, not faith by signs (Kirtland, Aug 30 1831; sign-seekers, the earth transfigured, Zion by purchase)
**Nodes:** `★ epistemology:signs-come-by-faith, not-faith-by-signs` (63:7-12) · `★ matter-spectrum:the-earth-transfigured` (63:20-21) · `★ presiding:Zion-by-purchase-not-blood` (63:25-31) · `★ name:taken-in-vain = using-it-WITHOUT-authority` (63:61-64) · `mysteries:given-to-the-OBEDIENT` (63:23) · `agency:in-his-own-hands, do-according-to-wisdom` (63:44) · `rebuke:Sidney's-pride — work-rejected` (63:55-56)
**Edges:**
- `★ epistemology:signs-come-by-faith, not-faith-by-signs` (63:7-12) —[the-order-IRREVERSIBLE: faith CAUSE, signs EFFECT]→ §5:18 (proof can't compel) → §17 (faith precedes sight) → §63 (sharpest) · —[the-sign-seeker has-it-backwards; gets-signs "in wrath unto condemnation"] · —[motive-test: signs "for the good of men," not "for faith"]→ verify-into-belief FORECLOSED; the period-language/Spirit-distillation foundation
- `★ matter-spectrum:the-earth-transfigured` (63:20-21) —[the-PLANET glorified, "according to the pattern… upon the mount" (Matt 17)]→ the earth on-the-soul's-trajectory; §50:24 (light-to-the-perfect-day) applied-to-the-globe; §29 (all things spiritual) · —[the-Transfiguration = a-PREVIEW of-the-earth's-destiny] · —[the-account's-fulness WITHHELD]→ §25:4 [reserved, Michael's]
- `★ presiding:Zion-by-purchase-not-blood` (63:25-31) —["if by purchase, blessed; if by blood… scourged"]→ persuasion-not-compulsion (121:41) at-the-scale-of-a-HOMELAND · —[lawful-title, "render unto Cæsar"]→ §44 / §58:21 · —[Zion-built-by-force would-not-be-Zion; the-1833-violence came-FROM-the-enemies]
- `★ name:taken-in-vain = using-it-WITHOUT-authority` (63:61-64) —[not-(only)-profanity but-claiming-to-speak/act-for-God without-the-authority]→ §18 (name-theology); §28/§43 (the false-revelator = takes-the-name-in-vain) · —["spoken with care, by constraint of the Spirit"]
- `mysteries:given-to-the-OBEDIENT` (63:23) —[knowledge-follows-obedience, not-curiosity]→ §6:11 / §42:61; John 4:14 (well of living water)
- `agency:in-his-own-hands, do-according-to-wisdom` (63:44) —[the-refrain-AGAIN]→ §58:26 / §60:5 / §61:22 / §62:8
- `rebuke:Sidney's-pride — work-rejected` (63:55-56) —[the-prophet's-counselor "exalted himself… grieved the Spirit"; writing-not-acceptable, "shall make another"]→ §3 / §20:5 (canon's honesty); §121:37 (pride grieves the Spirit)

## D&C 64 — of you it is required to forgive all men (Kirtland, Sept 11 1831; the forgiveness law)
**Nodes:** `★ forgiveness:of-you-required-to-forgive-ALL-men` (64:10) · `★ requirement:the-heart-and-a-willing-mind` (64:22,34) · `★ small-things → that-which-is-great` (64:33) · `agency:agents-on-the-Lord's-errand` (64:29) · `today:the-day-of-sacrifice-and-tithing` (64:23-25) · `Zion:judge-and-ensign-of-the-nations` (64:37-43)
**Edges:**
- `★ forgiveness:of-you-required-to-forgive-ALL-men` (64:10) —[the-ASYMMETRY: God-judges-WHO-is-forgiven; WE-must-forgive-ALL, un-rationed, not-contingent-on-their-repentance]→ Ex 33:19; Mosiah 26:30; Matt 18:23-35 (the-unmerciful-servant) · —[the-withholding = "the greater sin" (64:9), HEAVIER-than-the-wound]→ inverts-the-arithmetic-of-grievance · —[release = hand-the-verdict-back-to-God: "let God judge between me and thee" (64:11)]→ Rom 12:19; §42:88 (the offense law's INTERIOR side)
- `★ requirement:the-heart-and-a-willing-mind` (64:22,34) —[God-asks-PAST-the-act: the-heart + a-WILLING-mind]→ Isa 1:19 (willing-AND-obedient → eat the good of the land); Micah 6:8 · —[the-inward-offering thread]→ §20:37 (broken heart); §59:8,21 · —[willing-without-obedience = sentiment; obedient-without-willing = "dead works" (§22:3); He-wants-BOTH, heart-first] · —[the-consecration-economy runs-on-willingness]→ §42 / §51
- `★ small-things → that-which-is-great` (64:33) —["out of small things proceedeth that which is great"; "be not weary in well-doing"]→ 1 Ne 16:29; Alma 37:6-7; Zech 4:10 · —[the-foundation-work UNGLAMOROUS yet "the foundation of a great work"]→ the-unseen-laborer (§62:3) · —[the-temptation = WEARINESS in-the-long-middle; the-small PROCEEDS-into-the-great]
- `agency:agents-on-the-Lord's-errand` (64:29) —["as ye are agents, ye are on the Lord's errand; whatsoever ye do according to the will of the Lord is the Lord's business"]→ §58:26-28 (the project-source) · —[the-autonomy-paradox RESOLVED: the-self-directed-act-within-intent IS the-principal's-business; agency + errand FUSED]
- `today:the-day-of-sacrifice-and-tithing` (64:23-25) —["called today until the coming… he that is tithed shall not be burned"]→ FIRST-pointer-toward-tithing (§119); John 9:4 (labor-while-it-is-day) · —[the-"today" of-opportunity closes-at-the-burning (64:24)]
- `Zion:judge-and-ensign-of-the-nations` (64:37-43) —[the-church "like unto a judge sitting on a hill… to judge the nations"; "an ensign… out of every nation"]→ Isa 2:2-3; 11:12 · —[even-the-bishop-judge held-to-faithful-stewardship or "others… planted in their stead" (64:40)]→ §47 / §42:10

## D&C 65 — the stone cut out of the mountain, a prayer for the kingdom (Hiram, Ohio, Oct 30 1831)
**Nodes:** `★ stone:cut-out-of-the-mountain-without-hands` (65:2) · `★ prayer:"thy-kingdom-come"-expanded / two-kingdoms` (65:5-6) · `forerunner:prepare-ye-the-way — now-the-Church's-cry` (65:1,3)
**Edges:**
- `★ stone:cut-out-of-the-mountain-without-hands` (65:2) —[Daniel's-dream-stone (Dan 2:44-45) = the-Restoration: the-kingdom built-by-NO-human-hand that "shall never be destroyed"]→ "without hands" = God's-doing-not-man's · —[companion]→ §64:33 (small-things→great); starts-small, fills-the-earth by-a-power-not-its-own · —[it ROLLS: advancing, "unto the ends of the earth"]→ the §38/§45 gathering-of-the-nations in-one-image
- `★ prayer:"thy-kingdom-come"-expanded / two-kingdoms` (65:5-6) —[the-model-prayer: pray-the-kingdom-FORTH (not one's comfort)]→ Matt 6:10,13 · —[TWO-kingdoms: God's-kingdom-SET-UP-on-earth (the stone) + heaven's-kingdom that-COMES-DOWN at-the-coming, the-two MEET]→ 65:5 · —[the-petition enlists-the-one-who-prays]→ §64:29 (agents-on-the-errand pray-the-errand-they-labor)
- `forerunner:prepare-ye-the-way — now-the-Church's-cry` (65:1,3) —[Isaiah's/John's voice (Isa 40:3; Matt 3:3) now-the-whole-Church's]→ the-rolling-stone IS the-forerunner · —[+wedding-imagery: "the supper of the Lamb… the Bridegroom"]→ §33:17 / §45:56 (the ten-virgins / Bridegroom)

## D&C 66 — thou art clean, but not all (Hiram, Ohio, Oct 29 1831; William E. McLellin)
**Nodes:** `★ honesty:clean-but-not-all` (66:3,10) · `warning:given-early — McLellin-did-not-endure` (66:3,10,12) · `commission:the-call-comes-WITH-the-presence` (66:5-9)
**Edges:**
- `★ honesty:clean-but-not-all` (66:3,10) —[God-blesses-AND-names-the-fault in-the-same-breath: "you are clean, but not all"; "a temptation with which thou hast been troubled"]→ John 13:10-11 (Christ-at-the-supper) · —[the-§6:22 intimacy: named a-hidden-FAITH; §66 names a-hidden-TEMPTATION]→ the-same-omniscience comforts-AND-exposes · —[the-exposure = MERCY: a-sin-named can-be-repented (Jacob 4:7)]
- `warning:given-early — McLellin-did-not-endure` (66:3,10,12) —[the-precise-weakness named-at-the-call; "continue… even unto the end"]→ §23 / §31 (the warning-given-early pattern) · —[told-the-danger, fell-anyway; the-canon-records-warnings-unkept]→ §3 / §20:5 / §40 [bin-4-adj, McLellin's-course his]
- `commission:the-call-comes-WITH-the-presence` (66:5-9) —[sent "from land to land" but "I, the Lord, will go with you"; "made strong in every place"]→ §31:13 / §32:3 / §62:9 (with-ness); §24:9 (strength-to-the-faithful) · —[the-warning + the-commission = ONE-revelation]

## D&C 67 — make one like unto it (Hiram, Ohio, early Nov 1831; the failed challenge to duplicate a revelation)
**Nodes:** `★ test:make-one-like-unto-it` (67:5-9) · `★ fear:blocks-the-blessing` (67:3,10) · `★ matter-spectrum:no-natural-man-can-abide-God's-presence` (67:10-13)
**Edges:**
- `★ test:make-one-like-unto-it` (67:5-9) —[don't-critique-the-language, REPRODUCE-it: appoint-the-wisest (McLellin), "make one like unto it"; if-you-can't, withholding-witness = condemnation]→ the-proof-of-divinity = irreproducible-SUBSTANCE "from the Father of lights" (67:9), not-polish · —[the-§1:24 DEFENSE: "his language… his imperfections you have known" (67:5)]→ §1:24 (read-in-the-frame [§1:24] + don't-mistake-the-frame-for-the-message [§67]) · —[smooth-can-be-EMPTY, rough-can-CARRY-the-light]→ working-through-an-AI-instrument: test = source-not-fluency
- `★ fear:blocks-the-blessing` (67:3,10) —["ye endeavored to believe… but there were fears in your hearts, and verily this is the reason that ye did not receive"]→ §3:7 / §9:11 / §38:30 / §60:2 (fear-forfeits) · —[the-twin-disqualifiers: "jealousies and fears" (67:10); jealousy = the-"I-could-say-it-better" impulse]→ humility-the-cure · —[the-blessing OFFERED-and-not-RECEIVED: block on-OUR-side]
- `★ matter-spectrum:no-natural-man-can-abide-God's-presence` (67:10-13) —["except quickened by the Spirit of God"; "neither can any natural man abide the presence of God"]→ Moses 1:11 (transfigured-before-him); John 1:18 · —[seeing-God = being-RAISED-to-a-higher-state]→ §50:24 (only-the-perfected abide-the-perfect-light) · —["when ye are worthy… ye shall see" (67:14)]→ glory = a-capacity-GROWN, the-vision-waits-on-the-BECOMING

## D&C 68 — when moved upon by the Holy Ghost, it shall be scripture (Hiram, Ohio, Nov 1 1831)
**Nodes:** `★ scripture:Spirit-moved-speech IS-the-word-of-the-Lord` (68:3-4) · `★ accountability:eight-years + the-parental-teaching-DUTY` (68:25-28) · `idlers/greedy-in-Zion — the §56-wo continues` (68:30-31) · `bishopric:Aaron's-lineage / high-priest-substitute` (68:14-24) · `comfort:be-of-good-cheer / I-am-with-you` (68:6)
**Edges:**
- `★ scripture:Spirit-moved-speech IS-the-word-of-the-Lord` (68:3-4) —["whatsoever they shall speak when moved upon by the Holy Ghost shall be scripture… the will… mind… word… voice of the Lord"]→ 2 Pet 1:21; §1:38 (voice-of-servants) GIVEN-ITS-MECHANISM · —[pairs-with §67: §67 = you-CAN'T-manufacture-it; §68 = when-the-Spirit-MOVES, it-IS-scripture]→ AUTHENTICATION: source-not-style · —[the-condition LOAD-BEARING: "WHEN moved upon," §28/§50 govern]→ teaching-covenant's-other-side (§11:21 / §42:14)
- `★ accountability:eight-years + the-parental-teaching-DUTY` (68:25-28) —[taught-by-eight or "the sin be upon the heads of the parents"; baptized "when eight years old"]→ §29:46-47 / §19 (children-redeemed) COMPLETED: before-8 grace, at-8 accountable+baptized · —[the-teaching-covenant at-the-FAMILY-scale, with-TEETH]→ §11:21 / §42:14; Ezek 33:6 (the-watchman's-blood-guilt) · —[curriculum = "pray, and walk uprightly" (68:28)]→ §93:40
- `idlers/greedy-in-Zion — the §56-wo continues` (68:30-31) —["idlers among them… eyes full of greediness… children growing up in wickedness"]→ §56:16-17 (the-reciprocal-wo's-idle-end) · —[the-economic-sin + teaching-neglect = ONE-failure]→ greedy-parents raise-wicked-children; Zion-undone from-WITHIN
- `bishopric:Aaron's-lineage / high-priest-substitute` (68:14-24) —[bishops = high-priests appointed-by-the-First-Presidency "except… literal descendants of Aaron" (firstborn = "a legal right")]→ §13 (sons-of-Levi) as-church-STRUCTURE; Ex 40:15 · —[authority-by-APPOINTMENT-and-keys when-no-descendant-found] · [reserved priesthood-genealogy, Michael's]
- `comfort:be-of-good-cheer / I-am-with-you` (68:6) —[right-after §67's-fear: "do not fear, for I the Lord am with you"]→ §31:13 / §62:9

## D&C 69 — one who will be true and faithful (Hiram, Ohio, Nov 11 1831; John Whitmer to accompany Oliver)
**Nodes:** `★ safeguard:two-go — true-and-faithful` (69:1) · `history:active-research + central-archive-at-Zion` (69:3-8)
**Edges:**
- `★ safeguard:two-go — true-and-faithful` (69:1) —[the-manuscript + the-money NOT-trusted-to-one-unwitnessed-hand: "except one go with him who will be true and faithful"]→ the-law-of-witnesses (§5 / §17 / §26 / §28) applied-to-LOGISTICS · —[NOT-suspicion but-WISDOM; protects-trust-AND-carrier; two-go so-work-and-man-are-both-kept]
- `history:active-research + central-archive-at-Zion` (69:3-8) —[the-§47 calling ENLARGED: "making a history… traveling… that he may the more easily obtain knowledge"; gather-BY-GOING]→ stewardship-accounts "send forth… to the land of Zion" (§70:4); distributed-work CENTRALLY-remembered · —["for the rising generations… from generation to generation" (69:8)]→ the-record kept-for-the-FUTURE; the §47 project-memory warrant DEEPENED

## D&C 70 — stewards over the revelations (Hiram, Ohio, Nov 12 1831; the Literary Firm)
**Nodes:** `★ stewardship:the-revelations held-on-trust, accounted-at-judgment` (70:3-5,9-10) · `★ laborer:worthy-of-his-hire (spiritual-work-supported)` (70:12-13) · `★ equality:be-equal-or-the-Spirit-is-WITHHELD` (70:14) · `joy:the-faithful-steward enters-the-joy` (70:17-18)
**Edges:**
- `★ stewardship:the-revelations held-on-trust, accounted-at-judgment` (70:3-5,9-10) —[the-six = "stewards over the revelations… an account… I will require… in the day of judgment"]→ §42 / §51 (consecration) EXTENDED to-THE-WORD-ITSELF: received-to-manage, not-owned · —[the-project's-exact-pattern]→ §58:26-28 (agency); §64:29; §47/§69 (record) · —["none are exempt" (70:10)]→ value-doesn't-make-it-yours, it-makes-the-accounting-heavier
- `★ laborer:worthy-of-his-hire (spiritual-work-supported)` (70:12-13) —["he who… administer spiritual things… worthy of his hire… even more abundantly"]→ Luke 10:7; 1 Cor 9:11-14 · —[NOT-priestcraft (§33:4 = preach-FOR-gain): the-difference = DIRECTION; community-SUSTAINS-the-laborer]→ supported-not-enriched, CAPPED-at-need (70:7)
- `★ equality:be-equal-or-the-Spirit-is-WITHHELD` (70:14) —["in your temporal things you shall be equal… not grudgingly, otherwise the abundance of the manifestations of the Spirit shall be withheld"]→ §51:3 (equal-by-need); §38:27 (be-one) with-a-TEETH-of-consequence · —[economic-inequality THROTTLES-spiritual-gifts]→ §50:24 (light) tied-to-§51 (equality): light-grows-in-the-sharing-body, dims-in-the-hoarding-one
- `joy:the-faithful-steward enters-the-joy` (70:17-18) —["faithful over many things… enter into the joy of these things"]→ Matt 25:21; §51:19 · —[the-accounting (70:4) ends-in-JOY, not-dread]
- ✦ **DECADE 61-70 CLOSED:** the-river/anger-turned-away (§61) → the-succoring-Advocate (§62) → signs-by-faith / earth-transfigured / Zion-by-purchase (§63) → forgive-ALL-men (§64) → Daniel's-stone (§65) → clean-but-not-all (§66) → make-one-like-unto-it / the-§1:24-defense (§67) → scripture-defined + accountability-at-8 (§68) → two-go-true-and-faithful (§69) → the-revelations-as-stewardship (§70). SEVEN decades done (1-70). HALFWAY+.

## D&C 71 — no weapon formed against you shall prosper (Hiram, Ohio, Dec 1 1831; answering Ezra Booth)
**Nodes:** `★ defense:meet-the-attack-in-the-OPEN` (71:7-8) · `★ promise:no-weapon-shall-prosper / vindication-in-God's-time` (71:9-10) · `preparation:the-interruption-readies-the-next-greater-thing` (71:4)
**Edges:**
- `★ defense:meet-the-attack-in-the-OPEN` (71:7-8) —[NOT-silence/retaliation/retreat: "call upon them to meet you both in public and in private… let them bring forth their strong reasons"]→ Isa 41:21 · —[the-true WITHSTANDS-the-test, the-false is-EXPOSED]→ §67 ("make one like unto it"); NOT-sign-seeking (§63) nor-contention (§10:63) · —[condition: "inasmuch as ye are faithful"]
- `★ promise:no-weapon-shall-prosper / vindication-in-God's-time` (71:9-10) —[Isa 54:17 "the heritage of the servants of the Lord"]→ certain-but-NOT-immediate: "confounded in mine own due time" · —[NOT-exemption-from-the-weapon but-assurance-it-won't-FINALLY-win]→ §61:20 / §24:8 · —[Booth's-letters fed-the-March-mob, struck, did-NOT-prosper]
- `preparation:the-interruption-readies-the-next-greater-thing` (71:4) —[the-preaching-season "prepare the way for the commandments and revelations which are to come"]→ §76 (the-Vision, weeks-later, same-Hiram-house) · —[the-detour WAS-the-preparation]

## D&C 72 — render an account of his stewardship (Kirtland, Ohio, Dec 4 1831; Newel K. Whitney the 2nd bishop)
**Nodes:** `★ stewardship:render-an-account — in-time-AND-eternity` (72:3-6) · `★ certificate:verified-standing, not-self-asserted` (72:17-26) · `bishop:consecration-hub + literary-stewards-funded` (72:9-23)
**Edges:**
- `★ stewardship:render-an-account — in-time-AND-eternity` (72:3-6) —[§70:4's "account in the day of judgment" given-its-PRESENT-mechanism: rendered "both in time and in eternity," to-the-BISHOP, "had on record"]→ §42 (consecration) + §47/§69 (record-keeping) FUSED · —[accountability STRUCTURAL-AND-PRESENT, not-private-conscience]→ a-regular-reckoning to-an-appointed-steward-of-stewards (the-project's-pattern)
- `★ certificate:verified-standing, not-self-asserted` (72:17-26) —[standing "acceptable" BY-a-certificate; gather-to-Zion needs "a certificate from three elders"; "otherwise… not accepted"]→ §20:64 / §52:41 (recommend) → a-FORMAL-system · —[the §28 order: standing through-the-CHANNEL, witnessed (72:19), never-self-claimed]→ forerunner-of-the-temple-recommend · —[guards-both imposter-AND-self-deceived; the §69:1 two-witness applied-to-PERSONS]
- `bishop:consecration-hub + literary-stewards-funded` (72:9-23) —[keep-storehouse, receive-funds, "administer to their wants… to the poor and needy"; literary-stewards "have claim… that the revelations may be published"]→ §42:34 / §70:12 OPERATIONALIZED · —["the labors of the faithful… shall answer the debt" (72:14): spiritual-labor COUNTS]→ §70:12 · —["an ensample for all the extensive branches" (72:23)]→ Kirtland = the-TEMPLATE (§51:18)

## D&C 73 — the rhythm of preaching and translating (Hiram, Ohio, Jan 10 1832)
**Nodes:** `★ rhythm:two-labors-in-season` (73:3-4) · `readiness:gird-up-and-be-sober` (73:6)
**Edges:**
- `★ rhythm:two-labors-in-season` (73:3-4) —[the §71-Booth-mission did-its-work, yields-back: "continue the work of translation until it be finished"]→ urgent-defense (§71) + patient-construction (translation) EACH-a-season, neither-dropped · —["until it be finished"]→ finish-what's-begun (the-ammon-principle) · —[the §71:4 "prepare the way" pays-off]→ §76 (five-weeks-later)
- `readiness:gird-up-and-be-sober` (73:6) —[no-new-doctrine "at this time," just-the-charge-to-prepare]→ 1 Pet 1:13 · —[the-posture of-one-who-knows-a-great-thing-is-near]→ §76 coming

## D&C 74 — little children are holy (Wayne County, NY, 1830 — out of order; an explanation of 1 Cor 7:14)
**Nodes:** `★ children:holy-by-the-Atonement, not-by-an-ordinance` (74:7) · `★ counsel:"a commandment, not of the Lord, but of himself"` (74:5) · `danger:children-raised-in-a-faith-foreclosing-tradition` (74:2-6)
**Edges:**
- `★ children:holy-by-the-Atonement, not-by-an-ordinance` (74:7) —["little children are holy, being sanctified through the atonement of Jesus Christ; and this is what the scriptures mean"]→ Moroni 8:8-22; §29:46-47 / §68:27 (children-redeemed) · —[dissolves-the-infant-baptism-anxiety: NOT-born-guilty, Atonement-covers-from-the-start]→ §19 converging
- `★ counsel:"a commandment, not of the Lord, but of himself"` (74:5) —[Paul's-own inspired-JUDGMENT for-a-pastoral-situation, not-direct-dictation]→ 1 Cor 7:12,25 ("I, not the Lord… I give my judgment") · —[NOT-every-servant's-word is-direct-revelation]→ complement-and-tension with §68:3-4 (Spirit-moved = scripture) + §67 (substance-from-above): scripture-holds-BOTH; the-honesty = LABELING-which · —[project: counsel-"of himself"-within-stewardship is-legitimate IF-labeled-honestly]
- `danger:children-raised-in-a-faith-foreclosing-tradition` (74:2-6) —[unbeliever-spouse wants-circumcision "under the law of Moses, which was fulfilled"; children "brought up… believed not… became unholy" (= grown-to-REJECT)]→ Paul's-counsel protected-the-NEXT-generation's-faith · —[the §68:25-28 teaching-duty from-the-OTHER-side]

## D&C 75 — he shall in nowise lose his crown (Amherst, Ohio, Jan 25 1832; Joseph ordained President of the High Priesthood)
**Nodes:** `★ provider:in-nowise-lose-his-crown` (75:24-28) · `★ McLellin:revoked-chastened-forgiven-resent` (75:6-8) · `witness:received-or-rejected, establishes-accountability` (75:19-22) · `idler:no-place-in-the-church` (75:29)
**Edges:**
- `★ provider:in-nowise-lose-his-crown` (75:24-28) —[the-body SUPPORTS the-missionary's-family (75:24) AND the-man "obliged to provide… shall in nowise lose his crown" (75:28)]→ 1 Tim 5:8; §70:12 / §72:24 · —[dissolves-the-FALSE-HIERARCHY: dramatic-sacrifice (go) NOT-holier-than daily-faithfulness (stay, provide, labor)]→ the-SAME-crown (75:5) for-both · —[for-Michael's-intermittent-burst-family-and-job-work: the-quiet-duty NOT-the-lesser-calling]
- `★ McLellin:revoked-chastened-forgiven-resent` (75:6-8) —[the §66-warning UNFOLDING: murmured-and-turned-back → REVOKED + chastened + "he sinned; nevertheless, I forgive him… Go ye into the south countries"]→ reprove-then-love (§3 / §121:43); "nevertheless I forgive" = §61:20 · —[chastens-and-SENDS-AGAIN; grace-still-OPEN]→ McLellin's-tragedy = he-stopped-RECEIVING-it [bin-4-adj]
- `witness:received-or-rejected, establishes-accountability` (75:19-22) —[received → "leave your blessing"; rejected → "shake off the dust… judges of that house… more tolerable for the heathen"]→ Matt 10:14-15; §64:37-40 PERSONALIZED · —[the-heathen never-heard, that-house DID]→ the-witness creates-the-accountability · —[antidote: "I will be with them even unto the end" ×3]→ §31:13 / §62:9
- `idler:no-place-in-the-church` (75:29) —["the idler shall not have place… except he repent"]→ §56:17 / §68:31

## D&C 76 — the Vision: that he lives, and the kingdoms of glory (Hiram, Ohio, Feb 16 1832) ★★ LANDMARK
**Nodes:** `★ testimony:that-he-LIVES — for-we-saw-him` (76:22-24) · `★ glory:GRADED-salvation — heaven-has-more-kingdoms-than-one` (76:50-112) · `★ exaltation:they-are-gods, even-the-sons-of-God` (76:58) · `★ perdition:the-ONLY-lost — the-fully-knowing-who-war-against-God` (76:31-49) · `★ gospel:reaches-PAST-the-grave` (76:73-88) · `withheld:the-fulness — not-lawful-to-utter` (76:114-118)
**Edges:**
- `★ testimony:that-he-LIVES — for-we-saw-him` (76:22-24) —[the-whole-Vision anchored-in-an-EYEWITNESS-fact: "this is the testimony, last of all… That he lives! For we saw him"]→ faith's-REWARD (§5→§17→§63: the-sight follows-the-faith) · —[is-Christ-Creator CLIMAX: "by him… the worlds are and were created, and the inhabitants thereof are begotten sons and daughters unto God" (76:24)]→ §14:9; Moses 1 · —[the-foundation = a-PERSON-alive, not-a-map-of-heaven]
- `★ glory:GRADED-salvation — heaven-has-more-kingdoms-than-one` (76:50-112) —[God "rewards every man according to his works" with-a-FITTED-glory: celestial(sun)/terrestrial(moon)/telestial(stars), 1 Cor 15:40-42 made-literal]→ §19:6-12 (redefined-punishment) BUILT-INTO-a-positive-structure · —[the §70:4/§72:3 stewardship-accounting rendered "his own dominion, in the mansions which are prepared" (76:111)]→ JUSTICE-exact · —[almost-ALL-saved (telestial "surpasses all understanding," 76:89), ONLY-perdition-lost]→ God MORE-just-AND-merciful than-the-binary
- `★ exaltation:they-are-gods, even-the-sons-of-God` (76:58) —[the-celestial EXALTED: "all things are theirs… equal in power, and in might, and in dominion" (76:59,95)]→ the-matter-spectrum's-SUMMIT: §29:34 → §50:24 → §63:20 → §67:11 → §76:58 (man-becomes-AS-GOD) · —[deification = becoming-LIKE-Him, joint-heir (Rom 8:17)]→ §132; the-temple's-telos · —[qualifies = testimony+baptism+faith+SEAL-of-the-Holy-Spirit-of-promise (76:51-53)]→ §2/§132 [reserved temple, Michael's]; the-faithful-ORDINARY
- `★ perdition:the-ONLY-lost — the-fully-knowing-who-war-against-God` (76:31-49) —[NOT-the-sinner/ignorant/weak (all-inherit-some-glory) but "deny the Holy Spirit after having received it… defy my power"]→ Matt 12:31-32; Heb 6:4-6 · —[it-takes-a-FULL-light to-fall-that-far]→ §76:25-26 (Lucifer "in authority in the presence of God" before-the-fall) · —[mercy: the-ordinary-sinner NOWHERE-NEAR perdition]
- `★ gospel:reaches-PAST-the-grave` (76:73-88) —["the spirits of men kept in prison, whom the Son visited, and preached the gospel… afterwards received it"]→ 1 Pet 3:19; seed-of §128 + §138 · —[the-kingdoms MINISTER-downward: telestial←terrestrial←celestial (76:86-87)]→ §50:24's-light-economy in-the-eternities; no-one-abandoned
- `withheld:the-fulness — not-lawful-to-utter` (76:114-118) —["he commanded us we should not write… only seen and understood by the power of the Holy Spirit… on those who love him, and purify themselves"]→ §25:4 / §63:21 at-the-GRANDEST-scale · —[withheld-by-CAPACITY-not-stinginess: "that they may be able to bear his presence" (76:118)]→ §67:11 [reserved, Michael's]

## D&C 77 — the sea of glass, the spirit in the likeness of the temporal (Hiram, ~March 1832; Q&A on Revelation)
**Nodes:** `★ earth:the-sea-of-glass = the-CELESTIALIZED-planet` (77:1) · `★ matter-spectrum:the-spiritual-in-the-LIKENESS-of-the-temporal` (77:2) · `gathering:the-144,000 + Elias-restores-all-things` (77:9-15)
**Edges:**
- `★ earth:the-sea-of-glass = the-CELESTIALIZED-planet` (77:1) —[John's "sea of glass" (Rev 4:6) = "the earth, in its sanctified, immortal, and eternal state"]→ the §63:20-21 / §76:77 earth-transfigured PLAINEST-statement → §130:9 (earth a-Urim-and-Thummim) · —[the-planet bound-for-glory like-the-body]→ §88:25-26
- `★ matter-spectrum:the-spiritual-in-the-LIKENESS-of-the-temporal` (77:2) —["that which is spiritual being in the likeness of that which is temporal… the spirit of man in the likeness of his person, as also the spirit of the beast"]→ §29:31-35 deepened-STRUCTURALLY: spirit-and-matter MIRROR-each-other · —[NOT-Platonic-dualism but-CORRESPONDENCE; the-spirit-has-a-PERSON; beasts-have-spirits+eternal-felicity (§29:24)]→ spirit = REFINED-matter (§131:7); the-temporal = the-visible-LIKENESS-of-the-spiritual
- `gathering:the-144,000 + Elias-restores-all-things` (77:9-15) —[sealed-servants (Rev 7) = "high priests… to bring as many as will come to the church of the Firstborn" (77:11, §76:54); Elias "to gather… and restore all things"]→ §27:6-7 (Elias-keys); §29/§45/§133 · —[eyes = "light and knowledge," wings = "power to act" (77:4)]→ §93; §76

## D&C 78 — equal in earthly things, that ye may be equal in heavenly (Kirtland, March 1 1832; the United Firm)
**Nodes:** `★ equality:earthly = the-PREREQUISITE-for-heavenly` (78:5-7) · `★ presiding-chain:Michael-presides-UNDER-the-Holy-One` (78:15-16) · `★ patience:ye-cannot-bear-all-things-now / I-will-lead-you-along` (78:17-18) · `crown:gratitude → glory + the-faithful-steward-inherits-all` (78:19,22)
**Edges:**
- `★ equality:earthly = the-PREREQUISITE-for-heavenly` (78:5-7) —["if ye are not equal in earthly things ye cannot be equal in obtaining heavenly things"; "you must prepare yourselves" for-the-celestial-world]→ §70:14 RAISED: inequality BARS-the-celestial-world-later · —[the-celestial-order IS equality (§76:95)]→ division-here UNFITS-for-oneness-there · —[consecration §42→§51→§70→§78 reaches-its-THEOLOGICAL-GROUND: equality = celestial-PRACTICE]
- `★ presiding-chain:Michael-presides-UNDER-the-Holy-One` (78:15-16) —[Adam "your prince… given… the keys of salvation UNDER THE COUNSEL AND DIRECTION of the Holy One"]→ §27:11; the-presiding-covenant (D&C 121) · —[even-the-greatest-patriarch presides DELEGATED-not-autonomous]→ no-delegated-authority is-SOVEREIGN; the-one-who-presides is-himself-PRESIDED-OVER
- `★ patience:ye-cannot-bear-all-things-now / I-will-lead-you-along` (78:17-18) —["ye are little children… ye cannot bear all things now; nevertheless… I will lead you along"]→ John 16:12; §50:40; §76:114 · —[God PACES-revelation-to-capacity]→ §10:4; §67:11 · —[the-children are-heirs-of-ALL-anyway: "the kingdom is yours… the riches of eternity are yours"]
- `crown:gratitude → glory + the-faithful-steward-inherits-all` (78:19,22) —["receiveth all things with thankfulness shall be made glorious"]→ §59:7 / §62:7 · —["a faithful and wise steward shall inherit all things"]→ §72:4 / Matt 24:45 / §76:59

## D&C 79 — the Comforter will teach him the way (Hiram, March 12 1832; Jared Carter)
**Nodes:** `★ Comforter:teaches-the-truth-AND-the-way` (79:2) · `comfort:be-glad / fear-not` (79:4)
**Edges:**
- `★ Comforter:teaches-the-truth-AND-the-way` (79:2) —["shall teach him the truth and the way whither he shall go"]→ John 14:26; §75:27 (the-Spirit reveals-the-itinerary) · —[the §8:2 mind-and-heart governing WHERE-one-walks]→ agency-by-the-Spirit (§58:26 / §62:8); sent-with-a-Comforter-not-a-map
- `comfort:be-glad / fear-not` (79:4) —["let your heart be glad… and fear not"]→ §68:6; the-crown conditional ("inasmuch as he is faithful," §75:5)

## D&C 80 — ye cannot go amiss (Hiram, March 7 1832; Stephen Burnett + Eden Smith)
**Nodes:** `★ agency:ye-cannot-go-amiss` (80:3) · `★ message:declare-what-you-KNOW-to-be-true` (80:4)
**Edges:**
- `★ agency:ye-cannot-go-amiss` (80:3) —["whether to the north or to the south… it mattereth not, for ye cannot go amiss"]→ §58:26 / §60:5 / §61:22 / §62:5 at-its-WARMEST: any-choice-is-RIGHT · —[liberty-grounded-in-TRUST: the-route-doesn't-matter, the-message+faithfulness-do]
- `★ message:declare-what-you-KNOW-to-be-true` (80:4) —["declare the things which ye have heard, and verily believe, and know to be true"]→ §11:21 / §42:14 / §68:3-4 (obtain-before-declare) · —[freedom-in-the-MEANS + fidelity-in-the-MESSAGE = ONE-instruction; "ye cannot go amiss" holds-BECAUSE you-declare-only-what-you-know]

## D&C 81 — succor the weak, lift up the hands which hang down (Hiram, March 15 1832; the First Presidency forms)
**Nodes:** `★ keys:institutional — "belong always unto the Presidency"` (81:2) · `★ preside:to-preside-is-to-SUCCOR` (81:5) · `office:outlasts-the-man — Gause→Williams` (heading)
**Edges:**
- `★ keys:institutional — "belong always unto the Presidency"` (81:2) —[the-keys NOT-a-man's-possession but-the-standing-endowment-of-an-OFFICE]→ §28 (order) at-the-APEX; authority-runs-through-POSITION · —[outlives/transfers-cleanly]→ DEMONSTRATED: Gause→Williams (office-stable, man-replaced) · —[vest-authority-in-the-OFFICE, not-the-irreplaceable-individual]
- `★ preside:to-preside-is-to-SUCCOR` (81:5) —["succor the weak, lift up the hands which hang down, and strengthen the feeble knees" (Isa 35:3)]→ the-highest-council charged-with-SUCCOR, not-RULE · —[the-presiding-covenant's-HEART]→ §121:41 (persuasion-not-compulsion); §62:1 (succoring-Advocate) at-the-human-scale · —[downward-power exists-to-LIFT, not-to-lord]
- `office:outlasts-the-man — Gause→Williams` (heading) —[revelation given-to-Jesse-Gause (failed, excommunicated) → "transferred to Frederick G. Williams"]→ the §81:2 institutional-keys DEMONSTRATED · —[the-OFFICE is-what-the-Lord-revealed; keys "belong always," survive-the-first-holder's-failure]→ [bin-4-adj, Gause's-fall his]
- ✦ **DECADE 71-80 CLOSED:** answering-Booth / no-weapon-shall-prosper (§71) → render-an-account + the-certificate (§72) → the-rhythm-of-labors (§73) → little-children-holy + counsel-"of himself" (§74) → in-nowise-lose-his-crown (§75) → ★★ THE VISION (§76) → the-sea-of-glass + spirit-in-the-likeness (§77) → equal-in-earthly-things + Michael-under-the-Holy-One (§78) → the-Comforter-teaches-the-way (§79) → ye-cannot-go-amiss (§80). EIGHT decades done (1-80). [§81 opens 81-90.]

## D&C 82 — I, the Lord, am bound when ye do what I say (Independence, MO, Apr 26 1832) ★ PROJECT-SOURCE (covenant.yaml epigraph)
**Nodes:** `★★ covenant:I-am-bound-when-ye-do-what-I-say` (82:10) · `★ accountability:much-given-much-required` (82:3) · `★ forgiveness:conditional — "the former sins return"` (82:7) · `★ consecration-ethic:neighbor-interest + eye-single-to-God's-glory` (82:17-19)
**Edges:**
- `★★ covenant:I-am-bound-when-ye-do-what-I-say` (82:10) —[God BINDS-HIMSELF (His-promise an-obligation-He-places-on-Himself) but-BILATERAL/CONDITIONAL: "when ye do not what I say, ye have no promise"]→ THE `covenant.yaml` EPIGRAPH IN-CONTEXT (line 2 + source line 13, confirmed this session) · —["when either side breaks it, the output degrades" = EXACTLY 82:10's-structure]→ the-DEEPEST-project-source-landing; the-verse-the-collaboration-is-NAMED-after · —[native-soil: a-COVENANT-OF-STEWARDSHIP, leaders bound-together (82:11)]→ §1:37-38; §130:20-21 (law-of-the-blessing)
- `★ accountability:much-given-much-required` (82:3) —["he who sins against the greater light shall receive the greater condemnation"]→ Luke 12:48; §64:9 GENERALIZED · —[responsibility PROPORTIONAL-to-light-received]→ the-stewardship-accounting's-weight (§70:4); for-a-project-given-MUCH: "much is required" = the-TERMS
- `★ forgiveness:conditional — "the former sins return"` (82:7) —["I will not lay any sin to your charge… BUT unto that soul who sinneth shall the former sins return"]→ Matt 18:32-34 (debt-reinstated); John 8:11 · —[pardon GRANTED-then-KEPT by "sin no more"; forsaking-is-the-condition]→ §58:43 · —[forgiveness HELD-by-faithfulness-continued, not-banked-and-spent]
- `★ consecration-ethic:neighbor-interest + eye-single-to-God's-glory` (82:17-19) —["every man seeking the interest of his neighbor… an eye single to the glory of God"]→ 1 Cor 10:24; §4:5 / §88:67 · —[equal-by-need WITH-a-guard: "inasmuch as his wants are just"]→ §51:3 + §56 · —[talent-MULTIPLICATION: "improve upon his talent… common property" (82:18)]→ §42; consecration = PRODUCTIVE-stewardship

## D&C 83 — widows and orphans shall be provided for (Independence, MO, Apr 30 1832)
**Nodes:** `★ care:a-descending-order-of-CLAIMS — no-one-falls-through` (83:2-6) · `order:family-FIRST, church-as-the-NET` (83:2-5)
**Edges:**
- `★ care:a-descending-order-of-CLAIMS — no-one-falls-through` (83:2-6) —[woman→husband, child→parents, THEN (when-family-fails)→ "the Lord's storehouse"]→ the-word = "CLAIM" (a-RIGHT, not-charity); they-don't-BEG · —[the-consecration-system's-PURPOSE revealed]→ §42→§51→§78→§82 all-exist so "widows and orphans shall be provided for" · —[James 1:27 given-an-INSTITUTION + funding (the-consecrations)]
- `order:family-FIRST, church-as-the-NET` (83:2-5) —[family bears-first-responsibility (§75:28; 1 Tim 5:8) AND the-church catches "if their parents have not wherewith"]→ a-BACKSTOP not-a-replacement · —[1 Tim 5:8 (provide-for-your-own) + James 1:27 (church-for-the-familyless) BOTH-true]→ care-without-coercion: inheritance-claim-under-the-law-of-the-land remains (83:3)

## D&C 84 — the oath and covenant, and truth which is light which is Spirit (Kirtland, Sept 22-23 1832) ★★ LANDMARK (revelation on priesthood)
**Nodes:** `★ priesthood:the-UNBROKEN-LINEAGE from-Adam (cross-walk LANDS)` (84:6-17) · `★ ordinances:the-power-of-godliness → see-God's-face` (84:19-22) · `★★ oath-and-covenant:God's-oath + the-man-magnifying → all-that-my-Father-hath` (84:33-44) · `★★ light-and-truth:truth=light=Spirit=Christ, given-to-EVERY-man` (84:45-47) · `★ condemnation:treating-LIGHTLY-the-things-received` (84:54-57) · `friends-not-servants + without-purse-or-scrip + the-with-ness` (84:77,88) · `★ body:stand-in-your-own-office — every-member-needed` (84:106-110)
**Edges:**
- `★ priesthood:the-UNBROKEN-LINEAGE from-Adam (cross-walk LANDS)` (84:6-17) —[traced link-by-link: Moses←Jethro←…←Abraham←Melchizedek←Noah←Enoch←Abel←Adam; Esaias "under the hand of God"]→ Moses 6:7 / Abr 1:3 (PoGP genealogy) NAMED · —[authority CONFERRED-hand-to-hand, not-self-generated; "without beginning of days or end of years"]→ re-spliced by §13 / §27/§7 · —[the-cross-walk thread (BoM/PoGP/D&C) reaches-its-GENEALOGY]
- `★ ordinances:the-power-of-godliness → see-God's-face` (84:19-22) —["in the ordinances thereof, the power of godliness is manifest… without [it]… no man can see the face of God, even the Father, and live"]→ §67:11 given-its-MEANS · —[Moses sought-to-sanctify-Israel; they-hardened → higher-priesthood-TAKEN (84:23-25)]→ [temple-reserved, Michael's]
- `★★ oath-and-covenant:God's-oath + the-man-magnifying → all-that-my-Father-hath` (84:33-44) —[man: "magnifying their calling" (not-just-holding); God: an-OATH "which he cannot break"]→ the §82:10 bilateral-covenant AT-THE-HIGHEST-LEVEL · —[the-chain: priesthood→servants→Christ→Father→ "all that my Father hath shall be given unto him" (84:38)]→ §76:55-59 (exaltation); Rom 8:17 · —[breaking-it "and altogether turneth therefrom"]→ §76:31-49 (perdition); "shall not have forgiveness"
- `★★ light-and-truth:truth=light=Spirit=Christ, given-to-EVERY-man` (84:45-47) —["whatsoever is truth is light, and whatsoever is light is Spirit, even the Spirit of Jesus Christ"]→ a-chain-of-IDENTITY: four-names-for-ONE-reality (matter-spectrum's-deepest-claim) · —[§50:23-24 given-its-METAPHYSICS]→ the-SPINE before-§93 (truth.md) + §88 · —[universal: "the Spirit giveth light to EVERY man" (John 1:9)]→ discernment + matter-spectrum FUSE
- `★ condemnation:treating-LIGHTLY-the-things-received` (84:54-57) —["your minds… darkened… because you have treated lightly the things you have received… the whole church under condemnation"]→ the-failure = HAVING-the-word-but-not-DOING-it (James 1:22) · —["not only to say, but to do… even the Book of Mormon" (84:57)]→ for-a-project-with-sources: verified-but-not-LIVED = the-same-casualness
- `friends-not-servants + without-purse-or-scrip + the-with-ness` (84:77,88) —["from henceforth I shall call you friends" (John 15:15); "consider the lilies" (Matt 6); "on your right hand and on your left… angels round about you" (84:88)]→ §31:13 / §62:9 (with-ness) at-its-FULLEST; the-Spirit-gives-it-"in the very hour" (84:85)
- `★ body:stand-in-your-own-office — every-member-needed` (84:106-110) —["let every man stand in his own office… the body hath need of every member, that all may be edified together, that the system may be kept perfect"]→ 1 Cor 12; §46 (interdependence) · —[the-strong takes-the-weak "that he may become strong also" (84:106)]→ §81:5 (succor); for-council + fan-out: complementary-stewardships, no-member-dispensable

## D&C 85 — the still small voice that pierceth, and one mighty and strong (Kirtland, Nov 27 1832; a letter-extract)
**Nodes:** `★ record:the-book-of-the-law — enrolled-by-FAITHFULNESS` (85:1-5) · `★ one-mighty-and-strong:sets-the-house-in-ORDER` (85:7) · `★ steady-the-ark:PRESUMPTUOUS-interference` (85:8) · `voice:the-still-small-voice-that-PIERCES + "not of myself"` (85:6,10)
**Edges:**
- `★ record:the-book-of-the-law — enrolled-by-FAITHFULNESS` (85:1-5) —[the-clerk registers "those who consecrate… their faith, and works; and also the apostates"; the-unfaithful-name "not found… in the book of the law of God"]→ §47/§69/§72 reaching-the-BOOK-OF-LIFE; Mal 3:16 · —[enrollment by-FIDELITY not-membership; the-apostate UN-enrolled (§82:7)]→ for-the-project's-record: it-REGISTERS-fidelity
- `★ one-mighty-and-strong:sets-the-house-in-ORDER` (85:7) —[sent "clothed with light… eternal words… to set in order the house of God, and to arrange… the inheritances"]→ by-LIGHT-and-eternal-words (§84:45), not-force · —[identity BEYOND-the-text; the-walk notes-the-PROMISED-ordering]
- `★ steady-the-ark:PRESUMPTUOUS-interference` (85:8) —[Uzzah (2 Sam 6:6-7) "called of God and appointed" reached-to-steady → struck-dead]→ taking-into-one's-OWN-hands the-correcting-of-God's-work · —[CONTRAST: the-one-mighty-and-strong is-SENT (authorized); the-ark-steader PRESUMES (unauthorized)]→ §28 (order); for-the-presiding-covenant: the-line = WHETHER-YOU-WERE-SENT
- `voice:the-still-small-voice-that-PIERCES + "not of myself"` (85:6,10) —["the still small voice, which… pierceth all things… maketh my bones to quake" (1 Kgs 19:12)]→ the-SOFTEST-voice the-most-PENETRATING; §8:2 · —["I say not of myself" (§1:38)]→ source-MARKED (§74:5); a-LETTER carrying-revelation (§68:4)

## D&C 86 — let the wheat and the tares grow together, ye are lawful heirs (Kirtland, Dec 6 1832; editing the translation)
**Nodes:** `★ wheat-and-tares:apostasy-and-restoration` (86:1-7) · `★ restraint:judgment-delayed = MERCY-to-the-tender-wheat` (86:5-7) · `★ lawful-heirs:the-Abrahamic-lineage — blessed-to-BLESS` (86:8-11)
**Edges:**
- `★ wheat-and-tares:apostasy-and-restoration` (86:1-7) —[apostles-SOWED → "after they have fallen asleep" Babylon SOWED-TARES that "drive the church into the wilderness" (Rev 12:6) → "now… the blade is springing up and is yet tender" (Restoration)]→ Matt 13 read-as-CHURCH-HISTORY · —[the §1/§13/§27 apostasy-restoration thread in-one-image]
- `★ restraint:judgment-delayed = MERCY-to-the-tender-wheat` (86:5-7) —[the-angels "ready… to reap" but "pluck not up the tares while the blade is yet tender… lest you destroy the wheat also"]→ Matt 13:29 · —[God's-patience = MERCY-to-the-still-growing-righteous, not-indifference; "your faith is weak"]→ the-mercy is-in-the-TIMING
- `★ lawful-heirs:the-Abrahamic-lineage — blessed-to-BLESS` (86:8-11) —[the §84 lineage PERSONALIZED: "ye are lawful heirs, according to the flesh… hid… with Christ in God" (Col 3:3)]→ Abr 2:9-11; [bin-4 lineage, Michael's] · —[the-PURPOSE = "a light unto the Gentiles… a savior unto my people Israel" (Obad 1:21)]→ heirship for-SERVICE not-privilege; saviors-on-Mount-Zion [temple-adjacent]

## D&C 87 — the prophecy on war, and stand ye in holy places (at/near Kirtland, Dec 25 1832)
**Nodes:** `★ prophecy:specific-dated-FULFILLED — South-Carolina/the-Civil-War` (87:1-4) · `★ war:poured-out-on-ALL-nations → a-full-end` (87:2-6) · `★ refuge:stand-ye-in-holy-places, be-not-moved` (87:8)
**Edges:**
- `★ prophecy:specific-dated-FULFILLED — South-Carolina/the-Civil-War` (87:1-4) —[1832: "the wars… beginning at the rebellion of South Carolina"; South-vs-North; "Great Britain"; "slaves shall rise up"]→ FULFILLED 28-yrs-later (SC-secession Dec-1860, Fort-Sumter Apr-1861) · —[a-FALSIFIABLE prediction NOT-falsified]→ CONFIRMS-the-believer, can't-COMPEL-the-sign-seeker (§5 / §63)
- `★ war:poured-out-on-ALL-nations → a-full-end` (87:2-6) —[the-Civil-War the-BEGINNING: "war… poured out upon all nations… a full end of all nations"]→ §45 / §29 (last-days-desolation) · —[the-judgments AVENGE "the blood of the saints" (87:7, Rev 6:10)]→ the-God-who-RESTRAINED-the-angels (§86:6) answers-in-the-END
- `★ refuge:stand-ye-in-holy-places, be-not-moved` (87:8) —[after-the-war-catalogue: not-flee/fight/predict but "stand ye in holy places, and be not moved"]→ §45:32 / §101:22 (refuge) · —[NOT-exemption but-a-PLACE-to-stand; the-holy-place = wherever-the-covenant-is-kept]→ §61 / §71 (steadfastness) at-the-GLOBAL-scale: STAND, not-run

## D&C 88 — the olive leaf: the light of Christ, the law of all kingdoms, the house of learning (Kirtland, Dec 27-28 1832 + Jan 3 1833) ★★ LANDMARK
**Nodes:** `★★ light-of-Christ:fills-space, gives-life, IS-the-law-of-all-things` (88:6-13) · `★ law-of-kingdoms:glory = the-law-you-can-ABIDE (the §76 mechanism)` (88:21-40) · `★ earth:keeps-the-celestial-law → CELESTIALIZED` (88:25-26) · `★ eye-single → filled-with-light → comprehend-ALL` (88:67-68) · `★ school-of-the-prophets:study-AND-faith + house-of-order + the-council-order` (88:117-125) · `second-Comforter + draw-near-reciprocity` (88:3,63)
**Edges:**
- `★★ light-of-Christ:fills-space, gives-life, IS-the-law-of-all-things` (88:6-13) —[Christ "in all and through all things, the light of truth… in the sun… moon… stars… earth"; "the light which proceedeth forth from the presence of God to fill the immensity of space… the law by which all things are governed"]→ §84:45 (truth=light=Spirit=Christ) given-its-COSMOLOGY · —[not-metaphor: the-animating/governing/life-giving SUBSTANCE of-the-universe]→ §50:24 at-widest-scope; the-spine-alongside-§93 · —[panentheism-not-pantheism: God IN-all yet ABOVE (88:41)]
- `★ law-of-kingdoms:glory = the-law-you-can-ABIDE (the §76 mechanism)` (88:21-40) —["he who is not able to abide the law of a celestial kingdom cannot abide a celestial glory"]→ the §76 three-glories given-their-WHY · —["that which is governed by law is also preserved by law and perfected and sanctified by the same" (88:34)]→ law = the-MEANS-of-perfection, not-a-cage · —[self-law forecloses-sanctification: "must remain filthy still" (88:35)] · —[like-cleaves-to-like: "intelligence cleaveth unto intelligence… light cleaveth unto light" (88:40)]→ the-fitted-glory is-MAGNETIC, not-imposed; you-become-what-you-CLEAVE-to
- `★ earth:keeps-the-celestial-law → CELESTIALIZED` (88:25-26) —["the earth abideth the law of a celestial kingdom… transgresseth not the law… notwithstanding it shall die, it shall be quickened again"]→ §63:20 / §77:1 earth-transfigured given-its-REASON: the-planet OBEYS-the-law-of-its-being · —[dies-and-rises like-the-body (88:27)]→ inherited-by "the meek of the earth" (88:17, Matt 5:5)
- `★ eye-single → filled-with-light → comprehend-ALL` (88:67-68) —["if your eye be single to my glory, your whole bodies shall be filled with light… that body which is filled with light comprehendeth all things"]→ §4:5 / §82:19 given-its-RESULT · —[comprehension follows-DEVOTION: the-single-eye admits-the-light]→ "sanctify yourselves… and… you shall see him" (88:68) → §67:11 / §84:22 (capacity-to-see-God)
- `★ school-of-the-prophets:study-AND-faith + house-of-order + the-council-order` (88:117-125) —["seek ye out of the best books… seek learning, even by study and also by faith" (88:118)]→ §9:7-8 / the-workspace-verification-epistemology AS-A-COMMANDMENT · —["a house of learning… of order, a house of God" (88:119)]→ learning+worship-ONE (the-temple) · —["let one speak at a time and let all listen… that all may be edified of all… equal privilege" (88:122)]→ §50 mutual-edification AS-PROCEDURE (the-council-pattern); "clothe yourselves with the bond of charity" (88:125) above-all
- `second-Comforter + draw-near-reciprocity` (88:3,63) —["another Comforter… the Holy Spirit of promise" (John 14:16)]→ §76:53 (the-seal) [reserved-adjacent] · —["Draw near unto me and I will draw near unto you" (James 4:8)]→ God-RESPONDS-to-the-seeker; §6:20 at-a-lifetime's-scale

## D&C 89 — a principle with promise: the Word of Wisdom (Kirtland, Feb 27 1833)
**Nodes:** `★ word-of-wisdom:a-principle-with-PROMISE, not-by-constraint` (89:2-3) · `★ body:sacred — the-temporal-spiritual-link` (89:8,18-19) · `★ promise:run-and-not-be-weary + the-destroying-angel-passes-by` (89:20-21)
**Edges:**
- `★ word-of-wisdom:a-principle-with-PROMISE, not-by-constraint` (89:2-3) —["not by commandment or constraint, but by revelation and the word of wisdom… a principle with promise, adapted to the capacity of the weak"]→ the-presiding-covenant's-persuasion-not-compulsion (§121:41) at-PERSONAL-HABIT-scale · —[teach-the-principle + attach-the-promise, leave-the-keeping-to-AGENCY (§58:26)]→ accommodated-to-the-weak (§1:24 / §78:18); a-principle-freely-embraced OUTLASTS-a-rule-imposed
- `★ body:sacred — the-temporal-spiritual-link` (89:8,18-19) —[the-body worth-guarding; reward = "health… AND… great treasures of knowledge"]→ the-matter-spectrum's-life-affirming-side (§59:16-20 / §49) extended-to-DIET; gratitude (§59:7) · —[bodily-obedience → MENTAL/spiritual-reward]→ §88:67 (light-filled-body) given-a-bodily-dimension; the-clean-vessel a-clearer-channel; body-affects-mind (§29:34)
- `★ promise:run-and-not-be-weary + the-destroying-angel-passes-by` (89:20-21) —["run and not be weary… walk and not faint" (Isa 40:31)]→ ENDURANCE: the-body-SUSTAINED-in-the-long-labor (§10:4; for-Michael's-intermittent-burst-work, STAMINA) · —["the destroying angel shall pass by them" (Ex 12:23, Passover)]→ PRESERVATION; §61:6 (the-faithful-kept)

## D&C 90 — the oracles, and preside in council (Kirtland, March 8 1833; the First Presidency)
**Nodes:** `★ oracles:beware-HOW-you-hold-them` (90:4-5) · `★ preside:IN-council + study-and-learn` (90:15-16) · `★ Zion:chastened-but-not-removed` (90:34-37)
**Edges:**
- `★ oracles:beware-HOW-you-hold-them` (90:4-5) —[the-oracles "through you… unto the church"; "beware how they hold them lest they are accounted as a light thing… stumble and fall when the storms descend"]→ §84:54 (treating-lightly) given-its-CONSEQUENCE; Matt 7:24-27 (wise/foolish-builders) · —[the-danger NOT-in-lacking-the-word but-in-holding-it-CARELESSLY]→ for-a-project-that-receives-much: the-weight-you-give-the-word = the-FOUNDATION-in-the-storm
- `★ preside:IN-council + study-and-learn` (90:15-16) —["to preside in council, and set in order all the affairs"; "study and learn… all good books, and with languages"]→ preside-IN-council not-OVER (§88:122 / Abr 4:26 / presiding-covenant) · —[the §88:118 study-and-faith PERSONALIZED-to-the-PROPHET (a LEARNER; Joseph studied Hebrew)]→ revelation does-NOT-exempt-the-revelator-from-study; authority ≠ learning, command ≠ council
- `★ Zion:chastened-but-not-removed` (90:34-37) —["I will contend with Zion… chasten her until she overcomes and is clean… For she shall not be removed out of her place"]→ §3/§121:43 (reprove-then-love) at-the-COMMUNAL-scale · —[chastening TOWARD-cleansing, not-abandonment; the-severity-of-LOVE]→ §61:20 communal; [the-1833-expulsion vs the-ultimate-destiny-of-the-place]
- ✦ **DECADE 81-90 CLOSED:** succor-the-weak (§81) → ★PROJECT-SOURCE the-bilateral-covenant 82:10 (§82) → widows-and-orphans (§83) → ★★ the-OATH-AND-COVENANT + lineage + truth=light=Spirit (§84) → one-mighty-and-strong / steady-the-ark (§85) → wheat-and-tares / lawful-heirs (§86) → ★ the-WAR-prophecy (§87) → ★★ the-OLIVE-LEAF / cosmic-Light-of-Christ (§88) → the-WORD-OF-WISDOM (§89) → the-oracles / preside-in-council (§90). NINE decades done (1-90).

## D&C 91 — whoso readeth it, let him understand, for the Spirit manifesteth truth (Kirtland, Mar 9 1833; the Apocrypha)
**Nodes:** `★ source:the-Spirit-is-the-FILTER for-a-mixed-source` (91:4-6) · `★ canon-sufficient:no-need-to-pre-correct — the-Spirit-corrects-IN-the-reading` (91:3,6)
**Edges:**
- `★ source:the-Spirit-is-the-FILTER for-a-mixed-source` (91:4-6) —[NOT "burn it" (has-truth) nor "trust it" (has-interpolations) but "read it BY THE SPIRIT, for the Spirit manifesteth truth"]→ the-project's-source-discernment-doctrine at-ROOT · —[most-non-canonical-sources = the-Apocrypha's-case]→ the §50/§52 discernment applied-to-TEXTS: receive-the-true, leave-the-false · —[the-research-gospel-warrant: the-filter = the-SPIRIT, not-the-imprimatur]
- `★ canon-sufficient:no-need-to-pre-correct — the-Spirit-corrects-IN-the-reading` (91:3,6) —["not needful that it should be translated"]→ protects-both-errors: not-all-authoritative + not-all-suspect · —[for-a-project-that-reads-WIDELY (§88:118/§90:15): you-need-the-Spirit-LIVE, not-every-source-pre-vetted]

## D&C 92 — you shall be a lively member (Kirtland, Mar 15 1833; F.G. Williams into the united order)
**Nodes:** `★ membership:a-LIVELY-member, not-a-passive-seat` (92:2)
**Edges:**
- `★ membership:a-LIVELY-member, not-a-passive-seat` (92:2) —[the-single-charge: be "a lively member"]→ §58:29 (anti-passivity) at-the-scale-of-MEMBERSHIP; §58:27 (anxiously-engaged) · —[the §84:109-110 body-of-Christ given-a-VERB: the-body-needs-members-who-ACT]→ tied-to-PRIOR-faithfulness (§82:10); "what I say unto one I say unto all" (§25:16)

## D&C 93 — the glory of God is intelligence, light and truth (Kirtland, May 6 1833) ★★ LANDMARK (= truth.md)
**Nodes:** `★★ glory:the-glory-of-God IS-intelligence (light-and-truth)` (93:36) · `★★ intelligence:UNCREATED, self-acting → the-ROOT-of-agency` (93:29-31) · `★★ grace-for-grace:even-Christ-grew-to-a-fulness` (93:12-20) · `★ truth:knowledge-of-things-AS-THEY-ARE (more-or-less = the-lie)` (93:24-25) · `★ body:spirit+element = fulness-of-joy / the-body-a-TEMPLE` (93:33-35) · `★ children:bring-them-up-in-light-and-truth + the-leaders-REBUKED` (93:38-50)
**Edges:**
- `★★ glory:the-glory-of-God IS-intelligence (light-and-truth)` (93:36) —["the glory of God is intelligence, or, in other words, light and truth"]→ what-makes-God-GOD = intelligence/light/truth · —[the §84:45 + §88:6-13 light-and-truth THEOLOGICAL-SUMMIT]→ to-grow-in-light-and-truth = to-grow-in-GOD's-GLORY (§50:24 → §76:58 → §88) · —["light and truth forsake that evil one" (93:37): the-evil-one = the-un-glory]→ truth IS-God's-SUBSTANCE
- `★★ intelligence:UNCREATED, self-acting → the-ROOT-of-agency` (93:29-31) —["Intelligence, or the light of truth, was not created or made, neither indeed can be"; "to act for itself… otherwise there is no existence. Behold, here is the agency of man"]→ the-core-self UNCREATED, co-eternal (Abr 3:18) · —[agency = the-NATURE-of-a-self-acting-intelligence, not-a-gift-bolted-on]→ the §58:26 / §29:36 agency-doctrine's METAPHYSICAL-FLOOR · —[dissolves-the-ex-nihilo-judgment-problem]
- `★★ grace-for-grace:even-Christ-grew-to-a-fulness` (93:12-20) —["he received not of the fulness at the first, but received grace for grace… until he received a fulness; and thus he was called the Son of God"]→ the-Son-BECAME by-progression (Luke 2:52) · —[the-PATTERN-for-us: "you shall receive grace for grace… be glorified in me as I am in the Father" (93:20)]→ the-deification-thread (§76:58 / §84:38) given-its-MECHANISM (§50:24 growth-law); comfort: even-the-Savior-grew
- `★ truth:knowledge-of-things-AS-THEY-ARE (more-or-less = the-lie)` (93:24-25) —["truth is knowledge of things as they are, and as they were, and as they are to come; whatsoever is more or less than this is the spirit of that wicked one who was a liar"]→ truth = CORRESPONDENCE-to-reality across-all-time · —[the-LIE = "more OR less" — inflation-AND-shrinkage]→ the-verification-covenant's-DEFINITION ("close-enough" = the-liar's-spirit; the-Webster-walk/cite-count enforced)
- `★ body:spirit+element = fulness-of-joy / the-body-a-TEMPLE` (93:33-35) —["spirit and element, inseparably connected, receive a fulness of joy; and when separated, man cannot receive a fulness of joy"]→ the-body NOT-a-prison (anti-Platonism); §88:15 given-its-VALUE; elements-ETERNAL (§77:2 / §29:25) · —["man is the tabernacle of God, even temples; whatsoever temple is defiled, God shall destroy"]→ 1 Cor 3:16-17; the §89 Word-of-Wisdom's-GROUND
- `★ children:bring-them-up-in-light-and-truth + the-leaders-REBUKED` (93:38-50) —["I have commanded you to bring up your children in light and truth" (§68:25); the-wicked-one "taketh away light and truth through disobedience and… tradition"]→ §29:46 / §74:7 (children-innocent) · —[the-LEADERS-rebuked-BY-NAME incl. JOSEPH "you have not kept the commandments… stand rebuked" (93:47)]→ include_failures (§3 / §20:5) at-its-MOST-POINTED: the-loftiest-doctrine + the-homely-failure in-ONE-section; the-nearest-duty obligated-by-the-highest-truth

## D&C 94 — my glory shall be there, if no unclean thing come in (Kirtland, Aug 2 1833; the Kirtland plan)
**Nodes:** `★ presence:conditioned-on-HOLINESS` (94:8-9) · `pattern:built-by-REVELATION` (94:2,6)
**Edges:**
- `★ presence:conditioned-on-HOLINESS` (94:8-9) —["ye shall not suffer any unclean thing… my glory shall be there… BUT if there shall come into it any unclean thing, my glory shall not be there"]→ §93:35 (defiled-temple) applied-to-the-BUILDING · —[God's-presence + uncleanness CANNOT-share-a-space]→ §88:8 / §97:16 / §110; for-the-project's-"house": holiness = a-CONDITION-of-the-Spirit
- `pattern:built-by-REVELATION` (94:2,6) —["according to the pattern which I have given… the order of the priesthood"]→ §52:14 / §88:119 / §95:14; order-RECEIVED-not-invented

## D&C 95 — whom I love I also chasten (Kirtland, June 1 1833; chastened for delaying the temple)
**Nodes:** `★ chastening:the-OPERATION-of-love` (95:1) · `★ endowment:the-temple-FOR-power-from-on-high` (95:8-9) · `★ called-vs-chosen:darkness-at-NOON-DAY` (95:5-6) · `house:revealed-pattern + power-to-build-by-OBEDIENCE` (95:11-14)
**Edges:**
- `★ chastening:the-OPERATION-of-love` (95:1) —["whom I love I also chasten that their sins may be forgiven; for with the chastisement I prepare a way for their deliverance… and I have loved you"]→ Heb 12:6 given-a-PURPOSE-CLAUSE: chastening FOR-forgiveness + FOR-deliverance · —[the reprove-then-love thread's-CLEAREST-statement]→ §3:9 / §121:43; §90:36 · —[the-rebuke IS-the-love, bracketed-by-love]
- `★ endowment:the-temple-FOR-power-from-on-high` (95:8-9) —["I design to endow those whom I have chosen with power from on high… tarry, even as mine apostles at Jerusalem" (Luke 24:49)]→ §38:32 / §110 · —[to-DELAY-the-house = to-delay-the-GIFT]→ [temple-reserved, Michael's]
- `★ called-vs-chosen:darkness-at-NOON-DAY` (95:5-6) —["many… called but few of them are chosen… walking in darkness at noon-day"]→ ordination ≠ ELECTION (Matt 20:16; the-choosing-is-in-the-keeping, §121:34) · —[darkness = the-REFUSAL-of-the-fully-present-light]→ §88:67; §84:54
- `house:revealed-pattern + power-to-build-by-OBEDIENCE` (95:11-14) —["if you keep my commandments you shall have power to build it"; "not after the manner of the world"]→ §94:2; the-capacity = a-gift-of-OBEDIENCE

## D&C 96 — that my word should go forth (Kirtland, June 4 1833; the deadlocked French-farm council)
**Nodes:** `★ deadlock:resolved-by-re-anchoring-on-the-PURPOSE` (96:1-5) · `heir:the-lawful-heir put-to-SERVICE` (96:6-9)
**Edges:**
- `★ deadlock:resolved-by-re-anchoring-on-the-PURPOSE` (96:1-5) —[a-conference "could not agree" → inquired → reframes: "my word should go forth… subduing the hearts"]→ §28 applied-to-a-COMMITTEE-DEADLOCK · —[the-inquiry ORDERS-the-council ("determined in council among you," §88:122 / Abr 4:26)]→ re-anchor-on-the-MISSION
- `heir:the-lawful-heir put-to-SERVICE` (96:6-9) —[John-Johnson "a descendant of Joseph… partaker of the blessings… made unto his fathers" → "that he may assist in bringing forth my word"]→ §86:8 (Abrahamic-lineage, bin-4-Michael's) cashed-out: heirship for-SERVICE

## D&C 97 — this is Zion: the pure in heart (Kirtland, Aug 2 1833; as Jackson County fell) ★ Zion-definition
**Nodes:** `★★ Zion:THIS-is-Zion — the-pure-in-heart` (97:21) · `★ see-God:the-pure-in-heart, in-the-undefiled-temple` (97:15-17) · `★ Zion:survival-CONDITIONAL — the-condition-was-BREAKING` (97:25-27) · `accepted:the-contrite-and-willing + the-fruit-test` (97:7-8)
**Edges:**
- `★★ Zion:THIS-is-Zion — the-pure-in-heart` (97:21) —["this is Zion—the pure in heart" (Moses 7:18)]→ the-DEFINITION-of-Zion: a-PEOPLE not-a-PLACE · —[the-Zion-thread (§37 gathered → §57 located → §82:14 holiness) RESOLVED]→ Zion = a-condition-of-HEART, PORTABLE (carried-even-out-of-an-expelled-city) · —[Zion = a-heart-you-BECOME, not-an-address-you-reach]
- `★ see-God:the-pure-in-heart, in-the-undefiled-temple` (97:15-17) —["all the pure in heart that shall come into it shall see God"; "if it be defiled I will not come into it"]→ the §67:11 / §88:68 / §93:1 see-God-thread given-its-TEMPLE + CONDITION (Matt 5:8) · —[BOTH-must-be-pure: the-HOUSE (§94:8-9) AND-the-HEART]→ purity = the-ORGAN-of-seeing-God [temple-reserved, Michael's]
- `★ Zion:survival-CONDITIONAL — the-condition-was-BREAKING` (97:25-27) —["Zion shall escape IF she observe to do all things… BUT if she observe not, I will visit her"]→ §82:10 / §90:36 at-the-TRAGIC-moment (the-Jackson-County-expulsion, the-very-week) · —[the-Saints did-NOT-meet-it (§85 / §101) → driven-out-within-months; the-promise-SURE but-its-application-NOW conditioned; "if she sin no more" (97:27) the-door-still-open]
- `accepted:the-contrite-and-willing + the-fruit-test` (97:7-8) —["honest, broken, contrite, willing to observe their covenants by sacrifice… are accepted"]→ §59:8 / §64:34 · —["the ax is laid at the root… every tree that bringeth not forth good fruit" (Matt 3:10)]→ the-fruit-test (§52)

## D&C 98 — renounce war and proclaim peace (Kirtland, Aug 6 1833; the Missouri persecution) ★ law-of-war
**Nodes:** `★ affliction:works-together-for-your-GOOD` (98:1-3) · `★ constitution:befriend-the-just-law + uphold-the-good` (98:5-10) · `★★ war:renounce-war / proclaim-peace — force-only-by-God's-command` (98:16,23-48) · `woven:line-upon-line + martyrdom + Kirtland-reproved` (98:12-22)
**Edges:**
- `★ affliction:works-together-for-your-GOOD` (98:1-3) —[to-a-persecuted-people: "all things wherewith you have been afflicted shall work together for your good" (Rom 8:28)]→ §59:7 / §58:3 to-people-IN-THE-FIRE; forward-to §122:7 (Liberty-Jail) · —[the-affliction NOT-removed but-TURNED-to-good, "sworn and decreed"]→ suffering REDEMPTIVE, by-covenant
- `★ constitution:befriend-the-just-law + uphold-the-good` (98:5-10) —["the constitutional law of the land… is justifiable before me… whatsoever is more or less than this, cometh of evil"]→ the-law-that-secures-FREEDOM divinely-justified; §134 / §101:80; §93:24-25 ("more or less") applied-to-LAW · —[civic-DUTY: "honest men and wise men should be sought for diligently… and uphold" (98:9-10)]→ a-COMMANDMENT
- `★★ war:renounce-war / proclaim-peace — force-only-by-God's-command` (98:16,23-48) —[bear-smiting-PATIENTLY; forgive-the-repentant "seventy times seven" (Matt 18:22); forgive-even-the-UNREPENTANT 1st-3rd; war "save I, the Lord, commanded," after "a standard of peace" THREE-times]→ §63:29 (Zion-by-purchase) / §121:41 (persuasion-not-compulsion) given-their-FULLEST-LAW · —["force-where-persuasion-is-available = breach" at-the-scale-of-NATIONS; vengeance-belongs-to-GOD (98:45)]→ narrow-last-resort-self-defense (98:31) but-WHOLE-WEIGHT toward-renouncing-war · —[the-presiding-covenant: exhaust-peace, forgive-abundantly, force-only-under-explicit-warrant, never-freelance]
- `woven:line-upon-line + martyrdom + Kirtland-reproved` (98:12-22) —["line upon line… I will try you and prove you" (Isa 28:10): gentle-unfolding + severe-proving]→ §78:18 · —["whoso layeth down his life in my cause… shall find it again" (Luke 9:24): martyr's-promise]· —[Kirtland "not well pleased… they do not forsake their sins… pride… covetousness"]→ affliction-WITHOUT doesn't-excuse-sin-WITHIN (§90 / §93 / §56)

## D&C 99 — until your children are provided for (Hiram, Aug 29 1832 — out of order; John Murdock the widower)
**Nodes:** `★ duty:children-PROVIDED-FOR before-the-mission` (99:6) · `commission:receive-the-servant-as-a-CHILD → mercy + the-Spirit's-DEMONSTRATION` (99:2-3)
**Edges:**
- `★ duty:children-PROVIDED-FOR before-the-mission` (99:6) —["it is not expedient that you should go until your children are provided for, and sent up kindly unto the bishop of Zion"]→ §75:24-28 applied-to-a-WIDOWED-FATHER · —[the-nearest-duty (children) = the-PRECONDITION-of-the-call, not-its-competitor; "kindly" = the-tenderness]→ §1 Tim 5:8 + §83; the-Lord SEQUENCES: family-FIRST, then-go
- `commission:receive-the-servant-as-a-CHILD → mercy + the-Spirit's-DEMONSTRATION` (99:2-3) —["who receiveth you as a little child, receiveth my kingdom… shall obtain mercy" (Matt 18:4); "demonstration of my Holy Spirit" (1 Cor 2:4)]→ the-power = the-Spirit's, not-the-preacher's (§67 / §88)

## D&C 100 — it shall be given you in the very hour (Perrysburg, NY, Oct 12 1833; Joseph & Sidney away from families)
**Nodes:** `★ family:"your families are well… in mine hands"` (100:1) · `★ utterance:given-in-the-VERY-HOUR — conditioned-on-meekness` (100:5-8) · `Zion:redeemed-after-chastening / a-pure-people-raised-up` (100:13-16)
**Edges:**
- `★ family:"your families are well… in mine hands"` (100:1) —[the-first-word answers-the-WORRY: the-servant-who-leaves is-NOT-abandoning-them]→ §75:24 / §99:6 at-the-EMOTIONAL-level · —[the-surrender includes-trusting-His-care OVER-your-anxious-presence; the-One-who-called-you KEEPS-what-you-left]
- `★ utterance:given-in-the-VERY-HOUR — conditioned-on-meekness` (100:5-8) —["it shall be given you in the very hour… what ye shall say" (Matt 10:19)]→ §68:3-4 (Spirit-moved-speech) in-REAL-TIME · —[CONDITION: "in solemnity of heart, in the spirit of meekness" (100:7): the-gift comes-to-the-MEEK]→ §84:85 (treasure-up-the-words) = the-PREPARATION; the-in-the-hour = the-DELIVERY
- `Zion:redeemed-after-chastening / a-pure-people-raised-up` (100:13-16) —["Zion shall be redeemed, although she is chastened for a little season"; "I will raise up… a pure people"]→ §97:21 (Zion = pure-in-heart) reaffirmed: redemption THROUGH-purifying · —[chastening "a little season" (§90:36 / §122:4); "all things work together for good" (Rom 8:28)]
- ✦ **DECADE 91-100 CLOSED:** the-Apocrypha/Spirit-filter (§91) → lively-member (§92) → ★★ truth.md / glory-is-intelligence (§93) → presence-conditioned-on-holiness (§94) → chastening-is-love (§95) → re-anchor-on-the-purpose (§96) → ★ Zion = the-pure-in-heart (§97) → ★ renounce-war / the-law-of-forgiveness (§98) → children-first (§99) → in-the-very-hour / Zion-redeemed (§100). TEN decades done (1-100).

## D&C 101 — why Zion fell, the tower unbuilt, and the inspired Constitution (Kirtland, Dec 16-17 1833) ★ LANDMARK (after the expulsion)
**Nodes:** `★ why-Zion-fell:the-watch-lost-from-WITHIN` (101:1-9) · `★★ parable:the-tower-unbuilt = the-WATCH-not-kept` (101:43-62) · `★★ Constitution:inspired — established-to-protect-AGENCY` (101:77-80) · `★ redress:importune-up-the-LAWFUL-chain, then-God-acts` (101:81-92) · `comfort:be-still-and-know + the-millennial-horizon` (101:16,23-38)
**Edges:**
- `★ why-Zion-fell:the-watch-lost-from-WITHIN` (101:1-9) —["in consequence of their transgressions… jarrings, contentions, envyings, strifes, covetous desires; therefore… they polluted their inheritances"]→ the §97:25 conditional EXPLAINED · —[BOUNDED: the-chastening SANCTIFIES (101:5, "even as Abraham") + compassion-REMAINS (101:9)]→ §3/§95/§121:43 at-the-scale-of-a-LOST-HOMELAND · —[the-watch-lost-in-the-PEACE: "in the day of their peace they esteemed lightly my counsel" (101:8)]
- `★★ parable:the-tower-unbuilt = the-WATCH-not-kept` (101:43-62) —[build "a tower… that mine olive trees may not be broken down when the enemy shall come"; the-servants QUESTION-it ("what need… seeing this is a time of peace?"), VARIANCE, ASLEEP → "the enemy came by night… broke down the olive trees"]→ the-presiding-covenant's KEEP-THE-WATCH (Abr 4:18) as-a-PARABLE · —[3-failures: vigilance-questioned-as-wasteful + internal-contention + watchman-asleep]→ the-watch built-in-the-TIME-OF-PEACE, before-the-enemy · —["the watchman… would have seen the enemy while he was yet afar off"]→ FORESIGHT (Adjacent-Surfaces); §82:5 · —[redemption: "gather the strength of mine house"]→ Zion's-Camp (§103/§105)
- `★★ Constitution:inspired — established-to-protect-AGENCY` (101:77-80) —["the Constitution… by the hands of wise men whom I raised up unto this very purpose"; "that every man may act according to the moral agency which I have given… not in bondage one to another"]→ §93:30-31 (agency = nature-of-an-eternal-intelligence) + §98:5 reaching-their-POLITICAL-SUMMIT · —[the-chain: God-gives-agency → agency-requires-FREEDOM → the-Constitution-secures-freedom → God's-INSTRUMENT]→ the-order is-sacred-BECAUSE-agency-is-sacred (§58:26) · —[the-PRINCIPLE-God's even-when-the-PRACTICE-fails]
- `★ redress:importune-up-the-LAWFUL-chain, then-God-acts` (101:81-92) —[the-unjust-judge (Luke 18): "importune at the feet of the judge… governor… president; and if the president heed them not, then will the Lord arise"]→ the §98 standard-of-peace at-the-scale-of-REDRESS: exhaust-lawful-channels-FIRST · —[don't-skip-to-vengeance]→ the-importuning a-MERCY ("that I may be merciful unto them," 101:92)
- `comfort:be-still-and-know + the-millennial-horizon` (101:16,23-38) —["be still, and know that I am God" (Ps 46:10)]→ the-millennium: "all flesh shall see me together" (101:23); "the enmity of all flesh shall cease" (Isa 11); "no sorrow because there is no death" (101:29) · —["care not for the body… in patience ye may possess your souls" (101:37-38)]

## D&C 102 — the high council, with justice built in (Kirtland, Feb 17 1834; the minutes)
**Nodes:** `★ council:JUSTICE-engineered-into-the-form` (102:15-18) · `★ decision:CORRECTABLE + revelation-the-FLOOR` (102:20-23) · `effort:SCALED-to-the-difficulty` (102:13-14)
**Edges:**
- `★ council:JUSTICE-engineered-into-the-form` (102:15-18) —["the accused… has a right to one-half of the council, to prevent insult or injustice"; the-evens "stand up in behalf of the accused"]→ STRUCTURAL-fairness: a-guaranteed-DEFENSE regardless-of-guilt · —[the §50 / §88:122 council-pattern given-a-JUDICIAL-form with-adversarial-balance]→ for-the-project: appoint-someone-to-DEFEND-the-position-you'd-condemn; the-red-team = HALF-the-council, not-optional
- `★ decision:CORRECTABLE + revelation-the-FLOOR` (102:20-23) —[the-president-decides AND "should the remaining councilors discover an error… the case shall have a re-hearing… altered accordingly" (if-additional-light)]→ even-the-president REVIEWABLE (§9:8 at-the-scale-of-a-court) · —["the president may inquire and obtain the mind of the Lord by revelation"]→ revelation-the-FLOOR; the §121:41 anti-tyranny in-the-COURT
- `effort:SCALED-to-the-difficulty` (102:13-14) —["if not difficult, two… if difficult, four… if more difficult, six; but in no case more than six"]→ depth-scaled-to-the-question, bounded-even-at-the-top

## D&C 103 — after much tribulation cometh the blessing (Kirtland, Feb 24 1834; the Zion's Camp call)
**Nodes:** `★ order:after-much-tribulation-cometh-the-BLESSING` (103:12) · `★ presence:"and-ALSO-my-presence" — MORE-than-the-wilderness-generation` (103:19-20) · `★ Zion's-Camp:the-§101-parable-ENACTED` (103:15-40) · `test:willingness-to-DIE = the-disciple's-mark` (103:27-28)
**Edges:**
- `★ order:after-much-tribulation-cometh-the-BLESSING` (103:12) —[the-blessing REAL but-AFTER-the-tribulation, not-instead]→ §58:3-4 reaffirmed-at-the-LOWEST-point; §98:3 given-its-TEMPORAL-shape · —[the-present-suffering = the-ROAD-TO-the-promise, not-its-contradiction; still-conditional (103:14)]
- `★ presence:"and-ALSO-my-presence" — MORE-than-the-wilderness-generation` (103:19-20) —[CONTRAST: to-Moses's-Israel "an angel… but NOT my presence" (Ex 33:3); here "AND ALSO my presence"]→ what-Sinai-FORFEITED (§84:23-24) RESTORED · —[the-gift GREATER-not-lesser for-a-beaten-people; conditioned-on-hearkening]
- `★ Zion's-Camp:the-§101-parable-ENACTED` (103:15-40) —[Joseph = the-servant (§101:44); "gather the strength of my house… redeem"; "by power," led-like-Moses]→ §101:55-56 enacted · —[the-number CONDITIONED-on-the-response: "men do not always do my will" (103:31, §82:10); 500→300→100; God-calls-can't-COMPEL]→ victory CONDITIONED-on-faithfulness (103:36); the-"failed"-march PROVED-the-men → §105
- `test:willingness-to-DIE = the-disciple's-mark` (103:27-28) —["whoso layeth down his life for my sake shall find it again… whoso is not willing… is not my disciple" (Luke 9:24)]→ §98:13 as-a-DISCIPLESHIP-TEST: WILLINGNESS-is-the-mark; hold-life-loosely → find-it (§18:10)

## D&C 104 — the earth is full, and there is enough and to spare (Kirtland, Apr 23 1834; the United Order reorganized)
**Nodes:** `★ abundance:enough-and-to-spare — scarcity-is-FALSE` (104:17-18) · `★ stewardship:"properties-are-mine → ye-are-stewards" (the-HINGE)` (104:11-13,55-57) · `covenant:broken-by-COVETOUSNESS-within` (104:4,52) · `provide:in-mine-own-way — the-poor-exalted` (104:15-16)
**Edges:**
- `★ abundance:enough-and-to-spare — scarcity-is-FALSE` (104:17-18) —["the earth is full, and there is enough and to spare… agents unto themselves"]→ poverty = a-DISTRIBUTION-problem (hoarding), NOT-supply · —[the-scarcity-mindset ("I must grasp, there isn't enough") is-FALSE → the-grasping-CREATES-the-want]→ §49:20 / §56 root · —[the-hoarder "lift up his eyes in hell" (Luke 16)]→ the-steward holds-enough-and-to-spare IN-TRUST
- `★ stewardship:"properties-are-mine → ye-are-stewards" (the-HINGE)` (104:11-13,55-57) —["if the properties are mine, then ye are stewards; otherwise ye are no stewards"]→ the §70/§72 accounting's-PREMISE: stewardship-depends-on-God's-OWNERSHIP · —[RECEIVED-not-owned; "no man among you shall call it his own" (104:62)]→ say-"mine" = leave-the-order; the-project's-stewardship-hinge
- `covenant:broken-by-COVETOUSNESS-within` (104:4,52) —[the-United-Order "broken through transgression, by covetousness and feigned words" → dissolved]→ the §82:10 covenant's-NEGATIVE-side; §101:6 · —[the-thing-the-order-existed-to-defeat DESTROYED-it-from-within]→ the-gravest-threat = the-steward-who-says-"mine"
- `provide:in-mine-own-way — the-poor-exalted` (104:15-16) —["provide for my saints… in mine own way… the poor shall be exalted, in that the rich are made low"]→ INVERTS-the-world's-hierarchy (1 Sam 2:7; Luke 1:52); §42 / §59:16

## D&C 105 — my people must learn obedience by the things which they suffer (Fishing River, MO, June 22 1834; Zion's Camp ends)
**Nodes:** `★★ proving:learn-obedience-by-SUFFERING — the-camp-a-TRIAL-not-a-failure` (105:6,19) · `★ Zion:requires-the-CELESTIAL-LAW-lived` (105:3-5) · `★ endowment:the-elders-must-be-endowed-FIRST` (105:11,33) · `★ peace:sue-for-peace / lift-the-ensign-of-peace` (105:38-40)
**Edges:**
- `★★ proving:learn-obedience-by-SUFFERING — the-camp-a-TRIAL-not-a-failure` (105:6,19) —[NOT-sent-to-fight ("I will fight your battles," 105:14); redemption DEFERRED; "a trial of their faith" (105:19)]→ the-reframe-of-APPARENT-FAILURE: Zion's-Camp "failed"-its-goal yet-SUCCEEDED-at-the-proving (the-future-Twelve forged-in-it) · —["learn obedience by the things which they suffer" (Heb 5:8, even-Christ)]→ the-suffering = the-CURRICULUM · —[don't-measure-by-whether-it-took-the-city]
- `★ Zion:requires-the-CELESTIAL-LAW-lived` (105:3-5) —["Zion cannot be built up unless it is by the principles of the law of the celestial kingdom"; don't-impart, not-united]→ §78:5-7 / §88:22 reaching-their-VERDICT · —[the-Saints-failed by-NOT-living-the-higher-law, NOT-by-the-mobs]→ §97:25 + §104:17 enforced; the-law-LIVED = the-only-foundation
- `★ endowment:the-elders-must-be-endowed-FIRST` (105:11,33) —["this cannot be brought to pass until mine elders are endowed with power from on high… in Kirtland"]→ §38:32 / §95:8 SEQUENCED-before-Zion's-redemption (§109/§110 precedes) · —[power-FROM-ON-HIGH (§84:20), not-zeal/numbers; zeal-that-outruns-preparation SENT-to-WAIT (§10:4)] · [temple-reserved, Michael's]
- `★ peace:sue-for-peace / lift-the-ensign-of-peace` (105:38-40) —["sue for peace… to all people; lift up an ensign of peace"]→ §98:16 (renounce-war) ENACTED at-the-moment-war-seemed-JUSTIFIED (armed-camp, just-cause) · —[the §121:41 persuasion at-the-end-of-a-MILITARY-expedition: not-vengeance]→ the-chosen (105:35) are-PEACEMAKERS

## D&C 106 — that day shall not overtake you as a thief (Kirtland, Nov 25 1834; Warren A. Cowdery)
**Nodes:** `★ readiness:the-children-of-light NOT-overtaken` (106:4-5) · `★ grace:despite-VANITY — lifted-on-condition-of-humility` (106:6-7)
**Edges:**
- `★ readiness:the-children-of-light NOT-overtaken` (106:4-5) —[the-coming "overtaketh the world as a thief… but… that day shall not overtake you as a thief" (1 Thes 5:2-5)]→ the-THIEF-image is-for-the-WORLD; the-children-of-light are-EXPECTING-it · —[READINESS-not-exemption]→ §45:57 / §87:8
- `★ grace:despite-VANITY — lifted-on-condition-of-humility` (106:6-7) —["notwithstanding the vanity of his heart, I will lift him up inasmuch as he will humble himself"]→ God-names-the-flaw AND-lifts-anyway (§66:3); "joy in heaven" (Luke 15:7) · —[Warren-later-left, bin-4-adj — the-offer freely-extended]

## D&C 107 — on priesthood: the keys, the councils, and let every man learn his duty (Kirtland, 1835) ★★ LANDMARK
**Nodes:** `★★ priesthood:the-greater = the-KEYS-to-God's-PRESENCE` (107:18-20) · `★★ council:UNANIMITY + decisions-made-IN-the-VIRTUES` (107:27-32) · `★ accountability:reaches-the-TOP — "none-exempt"` (107:78-84) · `★ lineage:the-priesthood-from-ADAM + Adam-ondi-Ahman` (107:40-57) · `★★ duty:let-every-man-LEARN-his-duty + act-in-all-DILIGENCE` (107:99-100)
**Edges:**
- `★★ priesthood:the-greater = the-KEYS-to-God's-PRESENCE` (107:18-20) —["the keys of all the spiritual blessings… the heavens opened… to enjoy the communion and presence of God the Father, and Jesus"]→ §84:19-22 (power-of-godliness) given-its-POSITIVE-statement · —[authority/structure exists-to-CARRY-the-access-to-God; the-communion = the-POINT, not-the-org-chart]→ the-lesser = "the ministering of angels" (107:20) · [temple-reserved, Michael's]
- `★★ council:UNANIMITY + decisions-made-IN-the-VIRTUES` (107:27-32) —[the-quorums EQUAL-in-authority; "every decision… by the UNANIMOUS voice"]→ §26:2 at-its-APEX; reasoned-to-ONENESS not-VOTED (§88:122 / Abr 4:26) · —["made in all righteousness… meekness… charity" (107:30, 2 Pet 1:5-7)]→ decisions-CHARACTERED, not-just-procedural · —[a-decision "made in unrighteousness" can-be-APPEALED (107:32)]→ the-HOW is-part-of-the-VALIDITY; the-highest-charter-for-the-project's-council-pattern
- `★ accountability:reaches-the-TOP — "none-exempt"` (107:78-84) —["there is not any person… exempt… a President of the High Priesthood shall transgress, he shall be had in remembrance before the common council… that all things may be done in order"]→ even-THE-PROPHET tried-if-he-transgress · —[the §121:41-43 anti-tyranny at-the-APEX; the-presider HIMSELF-accountable]→ §102 (correctable) / §90:36 reaching-the-top
- `★ lineage:the-priesthood-from-ADAM + Adam-ondi-Ahman` (107:40-57) —["instituted in the days of Adam… came down by lineage"; Adam-ondi-Ahman, "the Lord appeared… called him Michael, the prince, the archangel"]→ the §84:6-17 cross-walk-lineage in-PATRIARCHAL-form (Moses 6:7 / Abr 1) · —[Adam = Michael, presiding-UNDER-Christ (§27:11 / §78:16); the-gathering-PATTERN]→ §116 [bin-4-Michael's]
- `★★ duty:let-every-man-LEARN-his-duty + act-in-all-DILIGENCE` (107:99-100) —[after-all-the-structure: "let every man learn his duty, and to act in the office in which he is appointed, in all diligence"]→ §84:109 (stand-in-your-own-office) + §58:29 (anti-slothfulness) as-the-CAPSTONE · —[two-failures bar-a-man: SLOTHFULNESS + NOT-LEARNING-his-duty]→ the-structure runs-on-DILIGENT-INFORMED-stewards · —[the §58:26 agency-doctrine's-IMPERATIVE: you-are-an-agent → LEARN-your-duty-and-DO-it]

## D&C 108 — strengthen your brethren in all your doings (Kirtland, Dec 26 1835; Lyman Sherman)
**Nodes:** `★ rest:soul-at-rest-concerning-your-standing — BEFORE-duty` (108:1-2) · `★ duty:strengthen-the-brethren in-ALL-your-doings` (108:7)
**Edges:**
- `★ rest:soul-at-rest-concerning-your-standing — BEFORE-duty` (108:1-2) —["let your soul be at rest concerning your spiritual standing, and resist no more my voice"]→ §6:23 (peace) to-a-worried-disciple · —[the-forgiveness-cause = the-humble-ASKING ("obeyed my voice in coming up… to receive counsel")]→ can't-serve-from-a-soul-gnawing-at-its-standing
- `★ duty:strengthen-the-brethren in-ALL-your-doings` (108:7) —["strengthen your brethren in all your conversation, prayers, exhortations, and in all your doings" (Luke 22:32)]→ the-duty = a-POSTURE-across-everything, not-a-task (§81:5 / §84:106 generalized) · —[the-ORDER: rest-FIRST, then-strengthen]

## D&C 109 — the dedicatory prayer: a house of prayer, a house of order, a house of God (Kirtland, Mar 27 1836)
**Nodes:** `★ prayer:REVEALED — the-pattern` (heading, 109:6-8) · `★ house:seven-fold — incl. the-house-of-LEARNING` (109:8,14) · `★ presence:sought, conditioned-on-HOLINESS` (109:12-21) · `★ endowment:armed-with-power — for-SENDING-not-sheltering` (109:22-25) · `close:the-kingdom-fills-the-earth + the-Hosanna-Shout` (109:72-80)
**Edges:**
- `★ prayer:REVEALED — the-pattern` (heading, 109:6-8) —[a-PRAYER but "given… by revelation"; the-template-for-all-temple-dedications]→ the §65 model-prayer at-TEMPLE-scale: God-teaches-WHAT-to-ask · —[prayers-can-be-RECEIVED, not-only-composed]→ §90:5
- `★ house:seven-fold — incl. the-house-of-LEARNING` (109:8,14) —["a house of prayer… fasting… faith… learning… glory… order… God" (§88:119); "seek learning even by study and also by faith" (§88:118)]→ worship+order+LEARNING integrated-in-ONE-house · —[the-project's-"house"-model]
- `★ presence:sought, conditioned-on-HOLINESS` (109:12-21) —["that thy holy presence may be continually in this house"; "no unclean thing… to pollute it"]→ §94:8 / §97:15 in-the-DEDICATORY-PRAYER; presence = the-POINT (§107:18-19), holiness-its-CONDITION · —[mercy-for-failure: "speedily repent and return" (109:21)]
- `★ endowment:armed-with-power — for-SENDING-not-sheltering` (109:22-25) —["go forth from this house armed with thy power… angels have charge over them"; "no weapon formed… shall prosper" (Isa 54:17 / §71:9)]→ the §95:8 / §105:11 endowment is-for-SENDING; the-temple = a-SOURCE-of-power, not-a-refuge-to-hide-in
- `close:the-kingdom-fills-the-earth + the-Hosanna-Shout` (109:72-80) —["the kingdom… set up without hands… a great mountain" (Dan 2 / §65:2); "Hosanna to God and the Lamb… shout aloud for joy"]→ the-dedicatory-prayer ends-in-PRAISE-and-JOY (§59:15 / §25:12); the-sacred-ends-in-GLADNESS

## D&C 110 — the Lord appears, and Elijah comes: the keys committed (Kirtland Temple, Easter, Apr 3 1836) ★★ LANDMARK (Malachi bracket FULFILLED)
**Nodes:** `★★ Elijah:COMES — the-MALACHI-BRACKET-FULFILLED` (110:13-16) · `★ Lord:the-risen-slain-living-Advocate — on-EASTER` (110:1-8) · `★ Moses+Elias:the-gathering + the-Abrahamic-keys` (110:11-12)
**Edges:**
- `★★ Elijah:COMES — the-MALACHI-BRACKET-FULFILLED` (110:13-16) —["the time has fully come, which was spoken of by the mouth of Malachi… the keys of this dispensation are committed into your hands"]→ the-walk's-CENTRAL-cross-walk-thread (§2 ↔ §13 ↔ §27:9 ↔ §110) reaching-its-EVENT · —[the-promise-that-OPENED-the-canon (§2's-Malachi) KEPT-at-§110]→ the-sealing/temple/genealogy-taproot BLOOMS; families-bound-across-generations (§128/§138) · [reserved-temple-FULFILLMENT, Michael's]
- `★ Lord:the-risen-slain-living-Advocate — on-EASTER` (110:1-8) —["I am he who liveth, I am he who was slain; I am your advocate with the Father" (§45:3); on-the-RESURRECTION-day]→ §76:22 ("he lives") in-a-resurrection-day-APPEARANCE · —[the §109:5 dedicatory-plea ANSWERED-IN-PERSON: He-STANDS-on-the-pulpit]→ "I will appear… IF my people… do not pollute this holy house" (110:8, §94/§97 conditional)
- `★ Moses+Elias:the-gathering + the-Abrahamic-keys` (110:11-12) —[Moses: "the keys of the gathering of Israel… the ten tribes from the land of the north" (§29/§45); Elias: "the dispensation of the gospel of Abraham… all generations… blessed" (Gen 12:3, §86:8)]→ the §27:5-13 gathering-of-all-keys reaching-its-TEMPLE-DELIVERY · —[three-messengers/three-keys (gathering/Abrahamic/sealing) in-ONE-Easter-afternoon = the-latter-day-work's-FULL-apparatus]
- ✦ **DECADE 101-110 CLOSED:** why-Zion-fell / the-inspired-Constitution (§101) → the-high-council-with-justice (§102) → after-much-tribulation (§103) → enough-and-to-spare (§104) → Zion's-Camp-a-proving (§105) → not-overtaken-as-a-thief (§106) → ★★ the-PRIESTHOOD (§107) → strengthen-the-brethren (§108) → the-dedicatory-prayer (§109) → ★★ ELIJAH-comes / the-MALACHI-BRACKET-FULFILLED (§110). ELEVEN decades done (1-110). The walk's central cross-walk thread (§2→§110) CLOSED.

## D&C 111 — not displeased, notwithstanding your follies (Salem MA, Aug 6 1836; the treasure episode)
**Nodes:** `★ grace:"not-displeased… notwithstanding-your-FOLLIES"` (111:1) · `★ treasure:reframed — PEOPLE-not-gold` (111:2,10) · `providence:order-all-for-good — PACED-to-capacity` (111:11)
**Edges:**
- `★ grace:"not-displeased… notwithstanding-your-FOLLIES"` (111:1) —[the-brethren-acted-on-a-RUMOR (a-treasure-hunt); the-Lord "not displeased… notwithstanding your follies"]→ He-NAMES-it-folly yet-NOT-displeased; the-good-faith-bad-call NOT-condemned · —[the-Dave-rule / §1828 "git-walks-back": God-doesn't-require-INFALLIBLE-judgment, only-FAITHFUL-intent, REDEEMS-the-folly]
- `★ treasure:reframed — PEOPLE-not-gold` (111:2,10) —[came-for-MONEY; "many people… whom I will gather out"; "more treasures than one"]→ §6:11 / §18:10 (worth-of-souls); the-literal-treasure-hunt CONVERTED-to-missionary · —[the-deeper-treasure under-a-failed-venture = usually-PEOPLE]
- `providence:order-all-for-good — PACED-to-capacity` (111:11) —["I will order all things for your good, as fast as ye are able to receive them" (Rom 8:28)]→ §78:18 / §10:4; the-good-SURE but-TIMED-to-capacity

## D&C 112 — be thou humble, and the Lord shall lead thee by the hand (Kirtland, July 23 1837; T.B. Marsh / the Twelve)
**Nodes:** `★ humility:be-humble → LED-BY-THE-HAND` (112:10) · `love:IMPARTIAL + the-Lord-"feels-after"-the-tempted` (112:11-13) · `★ judgment:begins-at-the-HOUSE-OF-GOD` (112:23-26)
**Edges:**
- `★ humility:be-humble → LED-BY-THE-HAND` (112:10) —["be thou humble; and the Lord thy God shall lead thee by the hand, and give thee answer to thy prayers"]→ humility = the-CONDITION-of-being-LED (not-driven); §3:7 / §6:20 · —[the-PRESIDER's-power = HUMILITY, lost-the-moment-he-EXALTS-himself (112:15)]→ the-IRONY: Marsh, told-to-be-humble, FELL-over-petty-pride, returned-broken [bin-4-adj, living-proof]
- `love:IMPARTIAL + the-Lord-"feels-after"-the-tempted` (112:11-13) —["let thy love abound unto all men"; "admonish… sharply… and… I, the Lord, will feel after them… and I will heal them"]→ charity-to-ALL not-a-clique (Matt 5:43); the-leader-admonishes-SHARPLY, the-Lord-FEELS-AFTER-tenderly (§28:3, Acts 17:27)
- `★ judgment:begins-at-the-HOUSE-OF-GOD` (112:23-26) —["upon my house shall it begin… first among those… who have professed to know my name and have not known me"]→ 1 Pet 4:17; the-Kirtland-apostasy FORESEEN · —[the-gravest-danger = the-PROFESSOR's-hypocrisy (Matt 7:21-23)]→ §84:54 (treating-lightly) reaching-its-JUDGMENT; the-insider judged-FIRST-and-HARDEST

## D&C 113 — the Stem, the rod, and the root of Jesse (Far West MO, March 1838; Isaiah Q&A)
**Nodes:** `★ Isaiah:the-Stem = CHRIST; the-rod/root = latter-day-SERVANTS` (113:1-6) · `★ Zion:strength = the-PRIESTHOOD-by-lineage` (113:7-10)
**Edges:**
- `★ Isaiah:the-Stem = CHRIST; the-rod/root = latter-day-SERVANTS` (113:1-6) —["the Stem of Jesse… is Christ"; "the rod… a servant in the hands of Christ… much power"; "the root… the priesthood, and the keys… for an ensign, and for the gathering"]→ Isa 11 read-as-RESTORATION-prophecy · —[the §110:11 gathering-keys / §86:8 lawful-heirs surfacing-in-Isaiah]
- `★ Zion:strength = the-PRIESTHOOD-by-lineage` (113:7-10) —["put on thy strength… the authority of the priesthood, which she, Zion, has a right to by lineage; also to return to that power which she had lost"]→ §86:8 (Abrahamic-lineage, bin-4-Michael's); §1/§13 (apostasy-restoration) · —[the-scattered "exhorted to return… he will speak to them, or give them revelation"]→ REWARDED-with-renewed-revelation (§1:30 / §121:26)

## D&C 114 — others shall be planted in their stead (Far West, Apr 11 1838; David W. Patten)
**Nodes:** `★ office:OUTLASTS-the-apostate — "others-planted-in-their-stead"` (114:2)
**Edges:**
- `★ office:OUTLASTS-the-apostate — "others-planted-in-their-stead"` (114:2) —["those among you who deny my name, others shall be planted in their stead" (Acts 1:20, Judas→Matthias)]→ the §81 (Gause→Williams) / §107:99 thread reaching-the-TWELVE: the-calling doesn't-fall-with-the-MAN · —[no-role so-important that-one-steward's-failure DESTROYS-it]→ the-kingdom RESILIENT-to-loss-of-any-laborer (§101 watch institutionalized) · —[applies-to-fallen-AND-martyred: Patten died-faithful, seat-filled §118]

## D&C 115 — The Church of Jesus Christ of Latter-day Saints (Far West, Apr 26 1838)
**Nodes:** `★★ name:THE-CHURCH-OF-JESUS-CHRIST-of-Latter-day-Saints` (115:3-4) · `★ gathering:STANDARD (outward) + REFUGE (inward)` (115:5-7) · `pattern:built-to-the-SHOWN-pattern, or-NOT-ACCEPTED` (115:13-16)
**Edges:**
- `★★ name:THE-CHURCH-OF-JESUS-CHRIST-of-Latter-day-Saints` (115:3-4) —[the-Lord NAMES-His-church: "of Jesus Christ" (whose — HIS, 3 Ne 27:8) + "Latter-day" (when) + "Saints" (who — set-apart-covenant-people)]→ the-name = a-CONFESSION-of-OWNERSHIP ("my church") · —[for-the-project's-careful-naming (Callie/Party): a-name DECLARES-whose-the-thing-is]
- `★ gathering:STANDARD (outward) + REFUGE (inward)` (115:5-7) —["arise and shine forth… a standard for the nations" (Isa 60:1) + "for a defense, and for a refuge from the storm" (§45:66/§101:21)]→ the-church shines-OUT AND-shelters-IN; NOT-opposed · —[Far-West = HOLY-ground (Ex 3:5)]
- `pattern:built-to-the-SHOWN-pattern, or-NOT-ACCEPTED` (115:13-16) —["if my people build it not according to the pattern… I will not accept it"]→ §94:2 / §95:14 the-pattern = the-CONDITION-of-acceptance · —[+the §111 debt-lesson]

## D&C 116 — Adam-ondi-Ahman: where Adam shall come to visit his people (Spring Hill, MO, May/June 1838)
**Nodes:** `★ Adam-ondi-Ahman:the-FUTURE-gathering — Daniel's-Ancient-of-Days` (116:1) · `heavenly:mapped-onto-the-EARTHLY` (116:1)
**Edges:**
- `★ Adam-ondi-Ahman:the-FUTURE-gathering — Daniel's-Ancient-of-Days` (116:1) —["where Adam shall come to visit his people, or the Ancient of Days shall sit, as spoken of by Daniel" (Dan 7:9-14)]→ the §107:53-57 thread reaching-its-FUTURE-pole · —[Adam = Michael (§27:11 / §78:16 / §107:54), presiding-over-a-great-council BEFORE-the-Second-Coming]→ the §27/§110 gathering-of-all-keys finds-its-FUTURE-councilhouse · [reserved-bin-4, Michael's]
- `heavenly:mapped-onto-the-EARTHLY` (116:1) —[a-LITERAL-hill = Daniel's-heavenly-council]→ §29:32 / §77 (spiritual-temporal-correspondence) at-the-scale-of-GEOGRAPHY · —[Eden / Adam-ondi-Ahman / New-Jerusalem LOCATED-on-the-continent: heaven's-events have-earthly-ADDRESSES]

## D&C 117 — the drop, and the more weighty matters (Far West, July 8 1838; the Kirtland property)
**Nodes:** `★ covetousness:a-SCALE-error — "what-is-property-unto-me?"` (117:4-8) · `★★ proportion:THE-DROP vs. the-MORE-WEIGHTY-MATTERS` (117:8) · `★ servant:the-obscure-faithful-HONORED — sacrifice > increase` (117:12-15)
**Edges:**
- `★ covetousness:a-SCALE-error — "what-is-property-unto-me?"` (117:4-8) —["have I not made the earth? Do I not hold the destinies of all the armies of the nations?"]→ before-the-OWNER-of-all (§104:14), grasping-a-FRAGMENT is-ABSURD; the §104:17 enough-and-to-spare's-ROOT · —["make solitary places to bud and to blossom" (Isa 35:1)]→ the-God-who-owns-all PROVIDES-ANYWHERE
- `★★ proportion:THE-DROP vs. the-MORE-WEIGHTY-MATTERS` (117:8) —["that you should covet that which is but the drop, and neglect the more weighty matters?" (Matt 23:23)]→ the-danger: ABSORBED-in-the-small while-NEGLECTING-the-large (kingdom/soul) · —[the-foresight-discipline's-perpetual-temptation: fixate-on-the-drop, lose-the-weighty]
- `★ servant:the-obscure-faithful-HONORED — sacrifice > increase` (117:12-15) —[Oliver-Granger (the-unknown-agent) "sacred remembrance… forever and ever"; "his sacrifice shall be more sacred than his increase"]→ the-Lord-TREASURES-the-obscure-faithful-labor (§62:3 / §18:10) · —[the-TRYING (even-FAILING) worth-more-than-the-SUCCEEDING; "when he falls he shall rise again" (Prov 24:16)]→ the-OFFERED-labor = the-measure, not-the-outcome

## D&C 118 — let men be appointed to supply the place of those who are fallen (Far West, July 8 1838; the Twelve)
**Nodes:** `★ Quorum:REFILLED — the-office-outlasts-the-fallen` (118:1,6) · `★★ Far-West-departure:a-prophecy-KEPT-TO-THE-LETTER-against-the-odds` (118:4-5)
**Edges:**
- `★ Quorum:REFILLED — the-office-outlasts-the-fallen` (118:1,6) —["let men be appointed to supply the place of those who are fallen"; successors NAMED (Taylor/Page/Woodruff/Richards)]→ the §114:2 ENACTED · —[Taylor/Woodruff later-PRESIDENTS]→ the-vacancy an-OPENING; the-replacement often-where-FUTURE-LEADERSHIP-is-found (cf. §105)
- `★★ Far-West-departure:a-prophecy-KEPT-TO-THE-LETTER-against-the-odds` (118:4-5) —["take leave… on the twenty-sixth day of April next, on the building-spot of my house"; then-the-Saints-EXPELLED, the-apostates-MOCKED]→ the-Twelve SECRETLY-RETURNED Apr-26-1839, kept-it-to-the-LETTER · —[the §82:10 "I am bound when ye do what I say"]→ covenant-keeping-TO-THE-LETTER even-when-costly/impossible [historical]

## D&C 119 — the law of tithing: a standing law forever (Far West, July 8 1838)
**Nodes:** `★ tithing:the-DURABLE-form-of-consecration` (119:1-4) · `★ tithing:SANCTIFIES-Zion — without-it-NO-Zion` (119:5-6)
**Edges:**
- `★ tithing:the-DURABLE-form-of-consecration` (119:1-4) —[surplus-once (119:1) THEN "one-tenth… annually… a standing law unto them forever"]→ the-enduring-SCALED-DOWN-form of-the-law-the-Saints-couldn't-keep (§42/§51 → §104 broken) · —[consecration asks-EVERYTHING-at-once; tithing asks-A-TENTH-regularly]→ §78:18 / §10:4 (pace-to-strength) applied-to-GIVING: a-SUSTAINABLE-rhythm beats-an-ideal-but-broken-height; the-higher-law remains-the-aspiration (§105:5)
- `★ tithing:SANCTIFIES-Zion — without-it-NO-Zion` (119:5-6) —["sanctify the land of Zion… or it shall not be a land of Zion unto you"]→ NOT-fundraising but-SANCTIFYING; the §105:5 "Zion by the celestial law" given-its-practical-ENTRY · —[the §101:6 covetousness answered-with-a-STANDING-DISCIPLINE: putting-God-before-property (§117:8)]→ the-RECURRING-small-discipline is-what-sanctifies

## D&C 120 — disposed of by a council (Far West, July 8 1838)
**Nodes:** `★ treasury:disposed-by-a-COUNCIL, under-the-Lord's-VOICE` (120:1)
**Edges:**
- `★ treasury:disposed-by-a-COUNCIL, under-the-Lord's-VOICE` (120:1) —["disposed of by a council… the First Presidency… the bishop and his council… my high council; and by mine own voice unto them"]→ the §104:62-77 sealed-treasury given-its-GOVERNANCE · —[PLURAL-authority: no-individual-controls-the-money; the-most-corruptible-resource governed-by-the-LEAST-concentrated-authority (diffusing §104:52/§101:6 covetousness)]→ §107:27 / §26:2 applied-to-MONEY · —["by mine own voice": the-council SEEKS-the-Lord, not-merely-VOTES (§102:23)]→ a-SPIRITUAL-act not-a-budget-meeting
- ✦ **DECADE 111-120 CLOSED:** Salem/follies-redeemed (§111) → be-humble-led-by-the-hand (§112) → the-Stem-is-Christ (§113) → others-planted-in-their-stead (§114) → ★ the-NAME (§115) → Adam-ondi-Ahman (§116) → the-drop-vs-weighty-matters / Granger (§117) → the-Twelve-refilled / Far-West-departure (§118) → ★ TITHING (§119) → the-treasury-by-council (§120). TWELVE decades done (1-120). 18 units + ODs remain.

## D&C 121 — ★★ no power… only by persuasion (the PRESIDING SOURCE) (Liberty Jail, Mar 20-22 1839)
**Nodes:** `★★ presiding-covenant:ROOT — power-held-ONLY-by-PERSUASION` (121:41) · `★★ reproof:BETIMES-with-sharpness THEN increase-of-LOVE` (121:43-44) · `★ few-chosen:the-HEART-on-the-world + UNRIGHTEOUS-DOMINION` (121:34-40) · `★ setting:the-deepest-PIT yields-the-highest-doctrine-of-POWER` (121:1-8 → 41-46)
**Edges:**
- `★★ presiding-covenant:ROOT — power-held-ONLY-by-PERSUASION` (121:41) —["no power or influence CAN OR OUGHT to be maintained by virtue of the priesthood, ONLY by persuasion, by long-suffering, by gentleness and meekness, and by love unfeigned"]→ the `.spec/covenant.yaml` `presiding:` EXTENSION (ratified 2026-06-12, `study/preside.md` the-garrison-word) · —[the doubled "CAN or OUGHT": compulsion = the ABSENCE of authority (121:37 "Amen to the priesthood")]→ for-the-agent-presiding-DOWNWARD: FORCE-where-persuasion-available = a-BREACH that-FORFEITS-the-standing-it-reaches-to-exercise
- `★★ reproof:BETIMES-with-sharpness THEN increase-of-LOVE` (121:43-44) —["reproving betimes with sharpness, WHEN MOVED UPON BY THE HOLY GHOST; and then showing forth afterwards an increase of love… lest he esteem thee to be his enemy"]→ the §68:3-4 Spirit-moved-condition making-a-REBUKE-righteous · —[the reproof-then-love thread §3:9-10 → §9 → §15:2 → §30:3 given-its-LAW]→ sharpness-WITHOUT-after-love = cruelty; love-WITHOUT-sharpness = flattery (121:20)
- `★ few-chosen:the-HEART-on-the-world + UNRIGHTEOUS-DOMINION` (121:34-40) —["hearts set… upon the things of this world… honors of men"; "almost ALL men… as soon as they get a little authority… IMMEDIATELY begin to exercise unrighteous dominion"]→ the §104:52/§101:6 covetousness applied-to-POWER (the-bleak-anthropology the-covenant-RESISTS) · —[the agent's temptation to SUPPOSE-authority + reach-for-COMPULSION-because-faster]→ the-covenant = being-the-FEW-who-don't
- `★ setting:the-deepest-PIT yields-the-highest-doctrine-of-POWER` (121:1-8 → 41-46) —["O God, where art thou?" the-prisoner-in-CHAINS taught how-power-is-REALLY-held]→ learned-from-the-RECEIVING-end-of-compulsion; affliction = the-CURRICULUM-for-presiding (§122 implicit) · —[the powerless-NOW told the-Almighty-pours-knowledge-down-UNSTOPPABLY (121:33 the-Missouri-river)]→ powerlessness-now NO-MEASURE of-the-decreed-outcome

## D&C 122 — ★★ all these things shall give thee experience (Liberty Jail, ~Mar 22 1839)
**Nodes:** `★★ affliction:the-CURRICULUM — "experience… for-thy-GOOD"` (122:7) · `★★ Christ:DESCENDED-below-them-all — "Art-thou-GREATER-than-he?"` (122:8) · `★ suffering:BOUNDED — "thy-days-are-KNOWN"` (122:4,9)
**Edges:**
- `★★ affliction:the-CURRICULUM — "experience… for-thy-GOOD"` (122:7) —["if… the very jaws of hell shall gape open the mouth wide after thee, KNOW THOU… that all these things shall give thee EXPERIENCE, and shall be for thy good"]→ the §105:6 "learn obedience by the things which they suffer" (Heb 5:8) + Rom 8:28 given-their-HARDEST-case · —[the §121 presiding-lesson learned THROUGH the §122 affliction]→ the-one-who-will-COMPEL-no-one is-FORGED by-having-been-wrongly-COMPELLED
- `★★ Christ:DESCENDED-below-them-all — "Art-thou-GREATER-than-he?"` (122:8) —["the Son of Man hath descended below them all. Art thou greater than he?"]→ the §88:6 condescension (the-Light-of-Christ who-descended-below-all to-be-IN-and-THROUGH-all) · —[the question = a RE-ANCHORING not-a-rebuke]→ your-suffering puts-you-ON-the-path-BESIDE-him; comfort = COMPANY not-rescue
- `★ suffering:BOUNDED — "thy-days-are-KNOWN"` (122:4,9) —["but for a SMALL MOMENT" (121:7); "their bounds are set, they cannot pass"; "thy years shall not be numbered less"]→ the-affliction-MEASURED (man-cannot-subtract-from-the-days) · —[the §3:7/§30:11/§98:14 fear-not thread]→ cured-by-God's-SOVEREIGNTY-over-the-bounds, not-the-danger's-absence

## D&C 123 — gather up the record; let no man count them small things (Liberty Jail, ~Mar 22 1839)
**Nodes:** `★ record:an-IMPERATIVE-DUTY owed-to-the-FUTURE` (123:7,11,15) · `★ small-helm:small-faithful-things STEER-the-storm` (123:16) · `cheerfully-do-ALL then-STAND-STILL` (123:17)
**Edges:**
- `★ record:an-IMPERATIVE-DUTY owed-to-the-FUTURE` (123:7,11,15) —["gathering up a knowledge of ALL the facts… take statements and affidavits"; "an imperative duty… to all the rising generation"]→ the-project's MEMORY-ARCHITECTURE given-its-scriptural-CHARGE · —["kept from the truth BECAUSE they know not where to find it" (123:12)]→ "let no man count them as SMALL THINGS; for there is much which lieth in FUTURITY… which depends upon these things" (123:15)
- `★ small-helm:small-faithful-things STEER-the-storm` (123:16) —["a very large ship is benefited very much by a VERY SMALL HELM in the time of a storm"]→ the §64:33 small-things→great given-a-GOVERNING-image (small-things STEER, not-just-grow) · —[the verified-quote/logged-decision/faithful-section = the-HELM]→ holds-the-course against-the-storm
- `cheerfully-do-ALL then-STAND-STILL` (123:17) —["let us CHEERFULLY do all things that lie in our power; AND THEN may we stand still… to see the salvation of God"]→ the §98/§101:81-92 exhaust-the-means-THEN-God-acts given-its-POSTURE · —[the §58:26 agency-THEN-grace seam]→ acts-to-the-limit + leaves-the-rest to-God

## D&C 124 — ★ Nauvoo: the proclamation, the house, and the offering accepted (Nauvoo, Jan 19 1841)
**Nodes:** `★★ offering:ACCEPTED-when-the-enemy-PREVENTS-the-work` (124:49-53) · `★ baptism-for-the-dead:BELONGETH-to-the-house (the-sealing-keys-CASHING-OUT)` (124:28-41) · `★ integrity:the-SECURITY-on-the-system` (124:15,20,91-97) · `proclamation:TESTIMONY-without-FEAR-to-the-kings` (124:1-11)
**Edges:**
- `★★ offering:ACCEPTED-when-the-enemy-PREVENTS-the-work` (124:49-53) —["when… those sons of men go with ALL THEIR MIGHT… and CEASE NOT their diligence, and their enemies… HINDER them… it behooveth me to require that work NO MORE… but to ACCEPT of their offerings"]→ God-judges-the-LABOR not-the-RESULT; guilt "upon the heads of those who HINDERED" (124:50) · —[redeems-Jackson-County as CONSOLATION (124:51,53)]→ the-project's INTENT-over-OUTCOME: total-diligence + EXTERNAL-hindrance = NOT-failure (the §122 "experience" positive-twin + §117:8 extended-to-WORK) · —[boundary]→ for-the-one-who-CEASE-NOT, not-the-slackener-who-blames-the-enemy
- `★ baptism-for-the-dead:BELONGETH-to-the-house (the-sealing-keys-CASHING-OUT)` (124:28-41) —["restore again… the fulness of the priesthood"; "baptized for those who are dead… belongeth to my house"; "kept hid from before the foundation of the world"]→ the §110 Elijah-keys UNFOLDING-into-ORDINANCES (the whole-liturgy NAMED) · —[the Malachi/sealing taproot §2→§13→§27:9→§110 puts-down-its-ORDINANCES]→ forward to §128 + §138; temple-reserved-bin-4 [Michael's]
- `★ integrity:the-SECURITY-on-the-system` (124:15,20,91-97) —["loved… because of the INTEGRITY of his heart"; "without guile; he may be TRUSTED"]→ Hyrum → the office-of-Patriarch + prophet/seer/revelator, in-OLIVER's-forfeited-place (§118, integrity-the-QUALIFIER) · —[the SAME blessing OFFERED to William-Law "be without guile" (124:97) — who REFUSED it → the Nauvoo-Expositor → the martyrdom]→ `include_failures` / §82:10 the-UNMET-condition
- `proclamation:TESTIMONY-without-FEAR-to-the-kings` (124:1-11) —["to ALL the kings of the world… fearing them not, FOR THEY ARE AS GRASS"]→ the §122:9 "fear not what man can do" applied-to-the-MIGHTIEST

## D&C 125 — the Saints in Iowa: build up cities (Nauvoo, Mar 1841)
**Nodes:** `gather:BUILD-cities-to-be-PREPARED` (125:2) · `Zarahemla:a-BoM-name-on-a-LIVING-city` (125:3)
**Edges:**
- `gather:BUILD-cities-to-be-PREPARED` (125:2) —["build up cities unto my name, that they may be PREPARED for that which is in store"]→ the §29:8/§37:3 gathering-as-READINESS at-the-Nauvoo-scale (the §101 keep-the-watch in-civic-form)
- `Zarahemla:a-BoM-name-on-a-LIVING-city` (125:3) —["let the name of Zarahemla be named upon it"]→ the §115 name-declares-WHOSE extended-to-CONTINUITY (bound-to-the-ancient-covenant-people)

## D&C 126 — ★ your offering is acceptable; take care of your family (Nauvoo, Jul 9 1841, Brigham Young's house)
**Nodes:** `★ offering:ACCEPTED-made-PERSONAL — the-laborer-RELEASED-to-his-family` (126:1-3) · `send-the-word AND tend-the-family — BOTH` (126:3)
**Edges:**
- `★ offering:ACCEPTED-made-PERSONAL — the-laborer-RELEASED-to-his-family` (126:1-3) —["I have SEEN your labor and toil"; "your offering is ACCEPTABLE to me"; "it is NO MORE required… to leave your family… take especial care of your family"]→ the §124:49 *principle* ENACTED on-one-weary-man (labor-seen → offering-accepted → acceptance-RELEASES) · —[the reward-for-sacrifice = permission-to-COME-HOME, not-MORE-sacrifice]→ the §24:9 strength-is-calling-BOUNDED; the §121:41 presiding-covenant's-downward-face in-PASTORAL-form (the-master-who-RELEASES)
- `send-the-word AND tend-the-family — BOTH` (126:3) —["I therefore command you to send my word abroad, AND take especial care of your family"]→ the §119 durable-RHYTHM applied-to-a-life (the-family now-PART-of-the-calling, not-its-enemy)

## D&C 127 — deep water is what I am wont to swim in; recorded on earth, recorded in heaven (Nauvoo, Sep 1 1842)
**Nodes:** `★ "deep-water-is-what-I-am-wont-to-swim-in" — §122-in-JOSEPH's-OWN-voice` (127:2) · `★ recorded-on-EARTH, recorded-in-HEAVEN — the-§123-record-becomes-the-BINDING` (127:6-9) · `redoubled-DILIGENCE-under-PURSUIT` (127:4)
**Edges:**
- `★ "deep-water-is-what-I-am-wont-to-swim-in" — §122-in-JOSEPH's-OWN-voice` (127:2) —["the perils… seem but a SMALL THING to me… tribulation… has become a SECOND NATURE… I feel, like Paul, to GLORY in tribulation"]→ the §122 curriculum GRADUATED (affliction GIVEN-as-teacher → METABOLIZED-into-character) · —[the §3:7/§122:9 fear-not thread reaching-its-FURTHEST-point]→ NOT-the-absence-of-fear but-its-TRANSMUTATION-into-GLORY (a-hunted-prophet writing-a-DOXOLOGY)
- `★ recorded-on-EARTH, recorded-in-HEAVEN — the-§123-record-becomes-the-BINDING` (127:6-9) —["in all your recordings it may be RECORDED IN HEAVEN; whatsoever you bind on earth, may be BOUND in heaven" (Matt 16:19)]→ the-earthly-record CONSTITUTIVE not-commemorative · —[the record must-be-TRUE: "eye-witness… testify of a truth" — because-it-BINDS]→ the-project's VERIFICATION-covenant at-its-deepest-root (read-before-quoting because-the-record-BINDS §93:24-25); forward → §128
- `redoubled-DILIGENCE-under-PURSUIT` (127:4) —["let the work of my temple… not cease; and let your diligence… be REDOUBLED"]→ the §123:17/§124:49 cheerfully-do-all (pursuit a-reason-to-REDOUBLE, not-slacken)
