# TITSW Framework — Source Log

*Working document. Kept as research provenance — traces how observations were reached.*

**Binding Question:** What does each TITSW principle actually look like in practice — and how does a model (or a human) tell the difference between a score of 3 and a score of 7?

**Purpose:** Curate a ~2,500 token context document (Layer 2) that gives a scoring model deep understanding of the TITSW framework. The v2 prompt gives bullet definitions. The framework doc gives *understanding* — what each principle looks like at different levels of quality, with exemplar anchors.

**Target file:** `experiments/lm-studio/scripts/context/01-titsw-framework.md`

---

## Outline

1. Meta-Principle A: Teach About Christ — definition, surface vs typological depth, what differentiation looks like
2. Meta-Principle B: Help Come Unto Christ — definition, knowing about vs experiencing, the receiving/transformation pattern
3. Principle 1: Love — what Christlike love looks like in teaching, safety, seeing divine potential
4. Principle 2: Spirit — teaching BY the Spirit vs ABOUT the Spirit, preparation, responsiveness
5. Principle 3: Doctrine — scripture-grounded, clarity, conversion focus, personal relevance  
6. Principle 4: Invite — specific, escalating, agency-respecting invitations to act

For each: 2-3 lines definition, 1-2 exemplar quotes, what differentiates a 3 from a 7.

---

## Verified Quotes

### [05-teach-about-jesus-christ](https://www.churchofjesuschrist.org/study/manual/teaching-in-the-saviors-way-2022/04-part-1/05-teach-about-jesus-christ?lang=eng) — Meta-Principle A
> "There are many things to teach about in the restored gospel of Jesus Christ—principles, commandments, prophecies, and scripture stories. But all of these are branches of the same tree, for they all have one purpose: to help all people come unto Christ and be perfected in Him" (opening paragraph)

> "'All things,' the Lord declared, 'are created and made to bear record of me' (Moses 6:63; see also 2 Nephi 11:4). With that truth in mind, we can learn to see a multitude of symbols in the scriptures that testify of the Savior."

**Three sub-dimensions defined by the manual:**
1. **Emphasize the Example of Jesus Christ** — "we don't just follow principles—we follow Jesus Christ"
2. **Teach about His Titles, Roles, and Attributes** — "go beyond what He said and did to who He is and what role He desires to play in our lives"
3. **Look for Symbols** — Moses 6:63 principle: "Looking for symbols reveals truths about the Savior in places you might otherwise overlook"

**Key insight for scoring:** The manual explicitly names typological reading ("branches of the same tree," symbols, parallels to Christ's life in prophets' lives). This is the surface-vs-deep dimension. A score of 3 = "mentions Jesus Christ." A score of 7 = "reveals how the topic IS about Christ even when Christ isn't named."

**Connects to:** Outline section 1 (Meta-Principle A). Also connects to gospel-vocab.md Patterns 4+7 (All Things Testify, Types & Shadows).

---

### [06-help-learners-come-unto-christ](https://www.churchofjesuschrist.org/study/manual/teaching-in-the-saviors-way-2022/04-part-1/06-help-learners-come-unto-christ?lang=eng) — Meta-Principle B
> "Nothing you do as a teacher will bless learners more than helping them know Heavenly Father and Jesus Christ and feel Their love"

> "It's good to *know about* the Savior's love, power, and mercy, but we also need to *experience* it."

> "Our ultimate goal in this life is to become more like our Heavenly Father and return to Him. The way we accomplish that goal is by coming unto Jesus Christ"

**Three sub-dimensions defined by the manual:**
1. **Help Learners Recognize the Lord's Love, Power, and Mercy in Their Lives** — move from abstract knowledge to personal experience: "learning about the experiences of Daniel is incomplete if it doesn't inspire us to trust the Lord when we face our own figurative den of lions"
2. **Help Learners Strengthen Their Relationship with Heavenly Father and Jesus Christ** — "how making and keeping covenants binds us to Them"
3. **Help Learners Intentionally Strive to Be More like Jesus Christ** — "becoming like Him happens only as we act in faith... making intentional choices to follow His example and receive His grace"

**Key insight for scoring:** The know-about vs experience distinction is the core differentiator. Score of 3 = "mentions coming to Christ." Score of 7 = "actively closes the gap between knowing about Christ and experiencing His power/love/mercy in the listener's life." The manual's Daniel example is revealing — *incomplete* teaching stays at the narrative level.

**Connects to:** Outline section 2 (Meta-Principle B). Also connects to Becoming skill — that bridge from knowledge to transformation.

---

### [08-love-those-you-teach](https://www.churchofjesuschrist.org/study/manual/teaching-in-the-saviors-way-2022/07-part-2/08-love-those-you-teach?lang=eng) — Principle 1: Love
> "Everything the Savior did throughout His earthly ministry was motivated by love."

> "When the Savior's love is in our hearts, we seek every possible way to help others learn of Christ and come unto Him. Love becomes the motivation for our teaching."

**Five sub-dimensions defined by the manual:**
1. **See divine potential** — "Jesus looked on Zacchaeus's heart and saw an honorable 'son of Abraham'" / "In unpolished fishermen... He saw the future leaders of His Church" / "In the feared persecutor Saul, He saw 'a chosen vessel'"
2. **Know their circumstances** — The woman at the well: "He knew that she had a troubled past... took the physical need that had her immediate interest—life-sustaining water—and connected it with her deeper spiritual needs for 'living water'"
3. **Pray for them by name** — 3 Nephi 17:17: "No one can conceive of the joy which filled our souls at the time we heard [Jesus] pray for us unto the Father"
4. **Create safety** — "He consistently reached out to those who were different... When a woman was accused of adultery, He made her feel safe and inspired her to repent"
5. **Express love** — 3 Nephi 17:3,5-6: "filled with compassion" / Moroni 7:48: "pray unto the Father with all the energy of heart, that ye may be filled with [the pure love of Christ]"

**Key insight for scoring:** Love-as-stated ("I love you" / "God loves you") vs love-as-demonstrated (knowing circumstances, seeing potential, creating safety, altering plans for people). A score of 3 = states God loves people. A score of 7 = demonstrates Christlike love through specificity (naming circumstances, seeing potential, creating safety for the vulnerable). The Zacchaeus/Samaritan woman examples are gold — love that KNOWS the person.

**Connects to:** Outline section 3 (Principle 1). Gospel-vocab Patterns 5+6 (Charity, Love Shed Abroad).

---

### [09-teach-by-the-spirit](https://www.churchofjesuschrist.org/study/manual/teaching-in-the-saviors-way-2022/07-part-2/09-teach-by-the-spirit?lang=eng) — Principle 2: Spirit
> "The Holy Ghost is the true teacher. No mortal teacher, no matter how skilled or experienced, can replace His role in witnessing of truth, testifying of Christ, and changing hearts."

> "But all teachers can be instruments in helping God's children learn by the Spirit."

**Five sub-dimensions defined by the manual:**
1. **Prepare spiritually** — Matt 4:1 (JST): "He was able to draw upon the 'words of life' that He had treasured up for the 'very hour' when He would need them" / "The way to invite the Holy Ghost into your teaching is to invite Him into your life"
2. **Respond to needs in the moment** — The woman who touched His clothes (Luke 8:41-48): "He perceived that in that multitude, someone had approached Him with a specific need" / "be sure that in your haste you don't unintentionally hurry past an urgent need"
3. **Create space for the Spirit** — Matt 16:13-17 (Peter's testimony): "He wanted them to find their answer not from 'flesh and blood' but directly from 'my Father which is in heaven'" / "Something as simple as the arrangement of the chairs in a room... sets a spiritual tone"
4. **Help seek/recognize/act on revelation** — D&C 6:21-24 (Oliver Cowdery): "Did I not speak peace to your mind?" / "One of the greatest gifts you can give as a teacher is to help those you teach progress in this lifelong pursuit of personal revelation"
5. **Bear testimony often** — John 11:23-27 (Martha): "His witness prompted Martha to share her own testimony" / "A testimony of truth is most powerful when it is direct and heartfelt"

**Key insight for scoring:** Teaching ABOUT the Spirit vs teaching BY the Spirit. A score of 3 = mentions the Spirit or quotes about the Spirit. A score of 7 = the teaching itself is an instrument of the Spirit — it creates space for personal witness, responds to real needs in the moment, and the teacher's own testimony invites the listener to feel. The manual's word "instrument" is key — the teacher doesn't replace the Spirit, they make room for Him. The Saul-of-Tarsus touch-His-clothes story is perfect: responsiveness to a need the teacher didn't plan for.

**Connects to:** Outline section 4 (Principle 2).

---

### [10-teach-the-doctrine](https://www.churchofjesuschrist.org/study/manual/teaching-in-the-saviors-way-2022/07-part-2/10-teach-the-doctrine?lang=eng) — Principle 3: Doctrine
> "Doctrine is eternal truth—found in the scriptures and the words of latter-day prophets—that shows us the way to become like our Father in Heaven and return to Him."

> "Did not our heart burn within us," they later reflected, "while he talked with us by the way, and while he opened to us the scriptures?" (Luke 24:27, 32)

> "The central purpose of all scripture is to fill our souls with faith in God the Father and in His Son, Jesus Christ" — Elder Christofferson

**Five sub-dimensions defined by the manual:**
1. **Learn the doctrine yourself** — Jesus's temptation in the wilderness: "He was able to draw upon" scripture He'd learned. D&C 11:21: "seek first to obtain my word"
2. **Teach from the scriptures** — Road to Emmaus (Luke 24): hearts burned as He opened the scriptures. "Be sure that your teaching does not drift away from the scriptures and words of prophets"
3. **Help seek/recognize/understand truths** — Lawyer asking "who is my neighbour?" (Luke 10:25-37): Jesus responded with a parable that led the man to answer his own question. "He rewards the seeker's acts of faith and patience"
4. **Focus on truths that lead to conversion** — Pharisees "looking beyond the mark" (Jacob 4:14) on Sabbath observance. "While there are many truths in the scriptures that can be discussed, it is best to focus on truths of the gospel that lead to conversion and build faith in Jesus Christ"
5. **Help find personal relevance** — Lost sheep, lost coin, prodigal son (Luke 15): "Whatever our circumstances, through His parables, the Savior invites us to find relevance in His teachings—to discover what He wants us to learn and what we may need to change"

**Key insight for scoring:** Doctrine is about scripture-grounding AND making it personally transformative. Score of 3 = quotes scriptures and sounds doctrinal. Score of 7 = opens the scriptures to hearers the way the Emmaus road encounter did — makes hearts burn, connects eternal truth to personal circumstance, leads to conversion not just knowledge. The Pharisees counterexample is critical: doctrinally correct but "looking beyond the mark." Doctrine without relevance is Pharisaic teaching.

**Connects to:** Outline section 5 (Principle 3). Gospel-vocab Patterns 1+4 (Doctrine of Christ, All Things Testify).

---

### [11-invite-diligent-learning](https://www.churchofjesuschrist.org/study/manual/teaching-in-the-saviors-way-2022/07-part-2/11-invite-diligent-learning?lang=eng) — Principle 4: Invite
> "'Come.' 'Come and see.' 'Come, follow me.' 'Go, and do thou likewise.' ... From the beginning of His ministry, the Savior invited His followers to experience for themselves the truths, power, and love that He offered."

> "Learning comes 'by study and also by faith' (D&C 88:118). And faith includes acting for ourselves, not simply being acted upon (see 2 Nephi 2:26)."

> "Our intent ought not to be 'What do I tell them?' Instead, the questions to ask ourselves are 'What can I invite them to do?'" — Elder Bednar

**Five sub-dimensions defined by the manual:**
1. **Help take responsibility for learning** — Brother of Jared (Ether 2-3): Lord gave some instructions, then asked "what will ye that I should do?" / JSH 1:20: "I have learned for myself"
2. **Encourage daily scripture study** — D&C 18:35-36: "It is my voice which speaketh... you can testify that you have heard my voice" / President Nelson: "Daily immersion in the word of God is crucial for spiritual survival"
3. **Invite preparation** — Parable of the sower (Matt 13): "Even the most precious... doctrine is unlikely to change a heart that is unprepared to receive it"
4. **Encourage sharing** — Enoch (Moses 6-7): "I am slow of speech" but "I will give thee utterance" / "give learners opportunities to share with each other what they are learning"
5. **Invite to LIVE it** — Sermon on the Mount conclusion (Matt 7:24): "Whosoever heareth these sayings of mine, and doeth them" / "Living the truth is the quickest path to greater faith, testimony, and conversion" / John 7:17: knowing comes from doing

**Key insight for scoring:** Invite is about specificity and escalation. Score of 3 = generic "we should all do better" / "pray and read scriptures." Score of 7 = specific, escalating invitations tied to the doctrine just taught — "Will you do X this week?" "Will you pray about Y?" The pattern is Come → Come and see → Come follow me → Go and do likewise. Peter walking on water is the archetype: invitation + faith + action + personal experience. The manual explicitly warns against passive learning: "It's not just listening or reading; it's also changing, repenting, and progressing."

**Connects to:** Outline section 6 (Principle 4). Gospel-vocab Pattern 1 (Doctrine of Christ — the invitation IS the doctrine: faith, repent, be baptized, receive the Holy Ghost, endure).

---

## Threads to Pull
- [x] What does the manual say about the RELATIONSHIP between meta-principles and the 4 principles? → Part 1 = WHAT we teach (meta-principles), Part 2 = HOW we teach (4 principles). The "branches of the same tree" framing ties them: everything IS Christ-centered, and the 4 principles are HOW you make it Christ-centered.
- [x] Are there scripture examples from the manual itself? → Yes, extensively. Each sub-dimension has a Christ exemplar (Zacchaeus, woman at the well, Peter's confession, Emmaus road, Peter walking on water, etc.)
- [x] What does "surface vs typological" mean concretely for teach_about_christ scoring? → Ground truth Alma 32 shows it: surface = 1-2 (Christ barely named), informed = 7-8 (the seed IS Christ, the tree IS the tree of life). The manual's Moses 6:63 / "branches of the same tree" principle.
- [x] What's the difference between love-as-stated and love-as-demonstrated? → Kearon "you are beloved" (stated, score 4) vs Brown "painful, especially when it came from people I cared about" (demonstrated through vulnerability, score 5). Manual: Zacchaeus (saw what others didn't), Samaritan woman (knew her thirst was more than physical).

**NEW threads discovered:**
- [ ] How does the ground truth scoring for Kearon spirit=3 map to the framework? → "tightly scripted — no pauses, no visible departure from prepared text. More teaching about the Spirit than teaching by the Spirit." This is the exact distinction the manual makes.
- [x] Elder Brown as scoring exemplar: 5/5 on all dimensions in Michael's analysis. What makes him an 8-9?  → Specificity invites the Spirit ("6:00 a.m. on a Wednesday"), vulnerability creates safety, multi-layered invitation pattern (reflect → act → choose → urgency → promise)

## Overview Study + Ground Truth Observations

### study/teaching-in-the-saviors-way/00_overview.md
Michael's overview confirms the structure: Part 1 = WHAT (two meta-principles), Part 2 = HOW (four teaching principles). His tables map each manual sub-dimension to a Christ exemplar. Key insight he identifies: "It's not enough to *know about* the Savior—we must *experience* His power." This is the Meta-B distinction.

### study/talks/202510-24brown.md — TITSW Scorecard Section
gold standard for what a score of 8-9 looks like across all dimensions:
- **teach_about_christ:** Jamaica as personal Palmyra = symbolic reading of personal experience
- **love:** vulnerability about peer persecution creates safety; names specific people
- **spirit:** sensory specificity invites the Spirit; teaches HOW revelation works
- **doctrine:** 25+ scriptures woven as argument, not proof-texts
- **invite:** five-level escalation pattern (reflect → act → choose → urgency → promise)

### experiments/lm-studio/scripts/references/ground-truth-alma32-kearon.md
surface vs informed scoring gap:
- Alma 32 teach_about_christ: surface 1-2, informed 7-8 (the seed IS Christ)
- Kearon love: 4 (stated love vs demonstrated love)  
- Kearon spirit: 3 (about the Spirit, not by the Spirit)
- Kearon invite: 8 (specific, escalating invitations)
Key lesson: the framework doc should help the model see what DEPTH looks like, not just presence.

## Cross-Study Connections
- [study/.scratch/gospel-vocab.md](gospel-vocab.md) — Pattern 1 (Doctrine of Christ) feeds Meta-Principle A; Pattern 5+6 (Faith/Hope/Charity + Love Shed Abroad) feed Principle 1 (Love)
- [experiments/lm-studio/scripts/references/ground-truth-alma32-kearon.md] — Ground truth scoring shows what surface vs informed scoring looks like

---

## Critical Analysis Notes

### Confirmed Strong
- "Surface vs typological" for teach_about_christ — manual's Moses 6:63, ground truth Alma 32 (1→7-8), and "branches of the same tree" framing all converge
- "Know about vs experience" for help_come_unto_christ — direct quote from manual
- "Stated vs demonstrated" for love — Kearon love=4 vs Samaritan woman exemplar
- "About vs by" for spirit — Kearon spirit=3 validates this distinction

### Needs Qualification
- Doctrine's differentiator ("opens scriptures" vs "quotes scriptures") is the hardest to express concisely in ~4 lines. Use the Pharisee counterexample to anchor it.

### Anti-Inflation Guardrail (CRITICAL)
- Richer definitions must CALIBRATE, not inflate. Include mid-range score exemplars (Kearon love=4, spirit=3) alongside high exemplars (Brown).
- The framework doc needs the same "presence ≠ high score" principle as gospel-vocab.md.
- Score-level anchoring is the key mechanism: concrete examples of what 3 and 7 look like.

### Decisions
- Each principle gets: 2-3 line definition, the key differentiator phrase, score anchor (what 3 looks like / what 7 looks like)
- Meta-principles get slightly more space because they're the framing
- Budget: ~2,500 tokens. Roughly: intro 100, meta-A 300, meta-B 300, love 350, spirit 350, doctrine 350, invite 350, anti-inflation note 100 = ~2,200. Room to breathe.
