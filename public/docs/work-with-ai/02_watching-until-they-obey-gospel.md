# Watching Until They Obey: The Feedback Loop as Divine Pattern

**Series:** AI and the Creation Pattern — Part 2 of 4
**Duration:** 30 minutes
**Audience:** Gospel-centered / YouTube
**Date:** February 2026

---

## Series Overview

| Part | Title | Focus |
|------|-------|-------|
| 1 | [The Creation Pattern](01_planning-then-create-gospel.md) | Abraham 4–5 as the blueprint for human-AI collaboration |
| **2** | **[Watching Until They Obey](02_watching-until-they-obey-gospel.md)** | **The feedback loop — reviewing, steering, and the role of agency** |
| 3 | [Intelligence Cleaveth Unto Intelligence](03_intelligence-cleaveth-gospel.md) | How the quality of your engagement shapes the quality of the output |
| 4 | [Intent Engineering](04_intent-engineering-gospel.md) | What the agent needs to *want* — purpose as infrastructure |

### Glossary

| Term | Definition |
|------|------------|
| **Session** | One prompt-and-response cycle. You say something, the AI processes and responds with text, tool calls, file edits, etc. |
| **Chat session** | The full conversation containing multiple sessions. Your ongoing back-and-forth in one chat window. |
| **Spec / Blueprint / Spiritual creation** | The planning document created collaboratively before implementation begins (Moses 3:5). |
| **Feedback loop** | Review → diagnose → correct → verify → repeat. The "watching until they obey" pattern. |
| **Seventh-day review** | Periodic reflection on the whole arc — tools, process, patterns, quality trends. The feedback loop applied to the creative process itself (Abraham 5:2). |

---

## Part 2: Watching Until They Obey (30 min)

### Opening (3 min)

**Recap from Part 1:** We found the creation pattern in Abraham 4–5 — the Gods envision, counsel, specify, organize, and *watch*. We mapped it onto AI-assisted development: plan before you build, spec before you code, spiritual creation before temporal creation.

But here's what we glossed over last time: **creation didn't end when the Gods gave the order.** Look at this phrase that appears throughout Abraham 4:

> "And the Gods **watched** those things which they had ordered, until they obeyed."
> — [Abraham 4:18](https://www.churchofjesuschrist.org/study/scriptures/pgp/abr/4?lang=eng&id=p18#p18)

The watching is not a footnote. It's a *phase of creation*. The Gods organized, ordered, and then *stayed present* to see if the creation matched the intention.

This is the most important thing I've learned about working with AI: **the review is where the real work happens.** Anyone can give a specification. The skill is in the watching — knowing what to look for, knowing when to steer, knowing the difference between "obeyed" and "close enough."

---

### The Watching Vocabulary of Abraham 4 (7 min)

Abraham 4 uses a progression of "watching" language that maps precisely onto the feedback loop in AI development. This isn't a metaphor I'm forcing — it's the actual text:

| Verse | Phrase | Development Equivalent |
|-------|--------|----------------------|
| [4:10](https://www.churchofjesuschrist.org/study/scriptures/pgp/abr/4?lang=eng) | "the Gods **saw that they were obeyed**" | First review — it matches the spec |
| [4:12](https://www.churchofjesuschrist.org/study/scriptures/pgp/abr/4?lang=eng) | "the Gods **saw that they were obeyed**" | Confirming the pattern holds for a second iteration |
| [4:18](https://www.churchofjesuschrist.org/study/scriptures/pgp/abr/4?lang=eng) | "the Gods **watched** those things which they had ordered, **until** they obeyed" | Active monitoring — this one took time, it wasn't instant |
| [4:21](https://www.churchofjesuschrist.org/study/scriptures/pgp/abr/4?lang=eng) | "the Gods **saw that they would be obeyed**, and that their plan was good" | Forward-looking confidence — they could project that this pattern would continue |
| [4:25](https://www.churchofjesuschrist.org/study/scriptures/pgp/abr/4?lang=eng) | "the Gods **saw they would obey**" | Trust established through repeated verification |
| [4:31](https://www.churchofjesuschrist.org/study/scriptures/pgp/abr/4?lang=eng) | "behold, they shall be **very obedient**" | Full trust — the pattern is proven, the system is reliable |

Look at the progression: *saw they were obeyed* → *watched until they obeyed* → *saw they would be obeyed* → *they shall be very obedient*.

That's a trust gradient. The Gods didn't just issue commands and walk away. They verified. They waited when something took longer. They built confidence through repeated observation. And by the end, they could declare the system reliable — "very obedient" — because they had *watched* it work.

**This is exactly how trust works with AI.**

Early in a project, you review every line. You verify every output against the spec. You watch closely. But as the AI proves reliable in certain areas — as you see it obey your patterns correctly, again and again — you develop calibrated trust. You watch less closely in the areas it's proven, and more closely in new territory.

The Gods model this perfectly. They didn't watch the light with the same intensity by day six. They watched most carefully when the creation was complex (verse 18 — the heavenly bodies, the lights "to rule the day and over the night"). Simple things, they confirmed and moved on. Complex things, they *watched until*.

---

### The Anatomy of "Until" (7 min)

That word **"until"** in verse 18 is doing enormous work. "The Gods watched those things which they had ordered, **until** they obeyed."

"Until" means:
- There was a gap between the order and the obedience
- The gap required patience
- The watching was *active*, not passive — they didn't leave and come back
- The result was not guaranteed by the command alone — it required oversight

In AI development, this is the feedback loop:

**Order** → the specification you give the AI
**Watch** → reviewing the output
**Until** → the iterative process of correction and verification
**Obey** → the output matches the spec

Sometimes the AI obeys immediately. You specify a function, it writes it correctly, it matches your architecture. *Saw that they were obeyed.* Move on.

Sometimes it takes iteration. The output is close but not right. You correct, the AI adjusts, you verify — and maybe correct again. *Watched until they obeyed.* This is normal. This is the expected pattern. Even the Gods experienced it.

#### Real Example: Building a Scripture Tool

I built a tool that fetches scripture text from a local database. Part of the spec said: "Accept verse ranges like `D&C 93:24-30` and return all verses in the range."

The AI built a parser. It worked for single verses. But for ranges, it only returned the first verse — it parsed `24-30` and kept just `24`.

**The order:** Handle verse ranges.
**The watching:** I used the tool. I asked for seven verses and got one. That mismatch — spec says seven, reality says one — is what "watching" catches.
**The "until":** I diagnosed the gap: the parser didn't split on `-` to extract start and end verses. I gave a specific correction: "Split the verse portion on `-`, add an `EndVerse` field, query `WHERE verse >= ? AND verse <= ?`."
**The obedience:** The fix worked. But I watched further — I tested both the range case *and* the single-verse case to make sure the fix didn't break what already worked.

Then I found a second issue. Cross-references were pulling from the entire chapter instead of the specific verse. Fixed that too. Then found the index needed rebuilding. Fixed that.

Each fix revealed the next thing to fix. This is *watching until they obey* — you don't declare success after the first fix. You continue watching until the whole system matches the whole spec.

#### The Principle in Scripture

This pattern appears beyond Abraham 4. Consider:

> "I will prove them herewith, to see if they will do all things whatsoever the Lord their God shall command them."
> — [Abraham 3:25](https://www.churchofjesuschrist.org/study/scriptures/pgp/abr/3?lang=eng&id=p25#p25)

God watches *us* the same way. Not because He doesn't know the outcome — He does, in a way that transcends time. But the watching, the proving, is part of the creative process. Agency requires it. Growth requires it. Something that is never tested is never proven "very obedient."

The same is true with AI. You can't know the output is correct without watching. You can't improve the process without the feedback loop. And the quality of your watching determines the quality of the creation.

---

### When to Steer vs. When to Let It Run (5 min)

The Gods didn't micromanage every detail of creation. They organized, ordered, and then watched — intervening only when necessary. There's a balance here that maps directly to AI work:

**Steer when:**
- The output contradicts the spec
- The AI is solving the wrong problem
- The approach is fundamentally off (wrong architecture, wrong pattern)
- You see something that will cause bigger problems later if not corrected now

**Let it run when:**
- The output is stylistically different from what you'd write, but functionally correct
- It's handling a detail you didn't specify — watch to see if its assumption is reasonable
- The code works and matches the spec, even if you'd have done it differently
- You're in the "they shall be very obedient" trust zone for this type of task

**The danger of over-steering:** If you correct every minor stylistic choice, you spend more time directing than the AI saves you. Worse, excessive corrections can muddy the context and cause the AI to lose track of the *important* constraints.

**The danger of under-steering:** If you let wrong patterns propagate uncorrected, they compound. A small architectural mistake in function #1 becomes a structural problem by function #20.

The Gods modeled the balance. They specified the *what* clearly: "Let the waters bring forth abundantly the moving creatures that have life" ([Abraham 4:20](https://www.churchofjesuschrist.org/study/scriptures/pgp/abr/4?lang=eng&id=p20#p20)). They watched the *outcome*. But they didn't specify every species of fish. They gave the organizing principle — "after their kind" — and let the creation unfold within those bounds.

> Specify the principle. Watch the outcome. Steer when the outcome diverges from the principle.

**Watch the instrument, not just the outcome.** "Watched those things which they had ordered, until they obeyed" includes watching whether you're using the right *method* of watching. In a study session on the Godhead, the AI produced a comprehensive, well-sourced document using keyword search and footnote chains. It looked complete. But the user asked: "Did you use gospel-vec?" — a question about the *instrument*, not the output. The answer was no. When semantic search was applied, it immediately found scriptures that keyword search had missed entirely: 2 Peter 1:4 ("partakers of the divine nature"), Psalm 103:13-14 (the Father's tenderness), D&C 132:24 (eternal *lives*). The Gods didn't just watch — they watched with *understanding*. They knew what they were looking for and whether their methods were sufficient to find it.

This is the essence of the feedback loop: not controlling every detail, but maintaining alignment between the specification and the creation.

---

### The Agency Question (4 min)

Now we need to address something deeper. Is the AI *obeying*? Does it have agency? And does it matter?

> "All truth is independent in that sphere in which God has placed it, **to act for itself**, as all intelligence also; otherwise there is no existence."
> — [D&C 93:30](https://www.churchofjesuschrist.org/study/scriptures/dc-testament/dc/93?lang=eng&id=p30#p30)

In the Restoration framework, intelligence is self-acting. It chooses. It's not merely responsive to external force — it acts from within. That's what makes it intelligence and not just matter.

AI doesn't have this quality. An AI model doesn't choose to help you or choose to drift. It operates within mathematical parameters, generating statistically probable outputs based on patterns in its training data. It's extraordinarily capable — but it's capability without agency. Pattern without will.

And yet — Abraham 4 uses the language of obedience for physical creation too. The waters *obeyed*. The lights *obeyed*. The earth *obeyed*. Are we saying water has agency?

I think the text is describing something about the structure of creation: when intelligence organizes matter according to law, the matter responds. It "obeys" not because it chooses, but because law governs it. The Gods applied law; matter responded according to that law. When the Gods watched "until they obeyed," they were watching matter align with the law they had applied.

AI is similar. It's not self-acting intelligence in the D&C 93 sense. It's a sophisticated system that responds to inputs according to its training. When I give it a specification and it generates code, it's not *choosing* to obey — it's *processing* according to patterns. When I correct it and it adjusts, it's not *learning* in the eternal sense — it's incorporating context.

**Why this matters practically:**

1. **Don't anthropomorphize the tool.** The AI isn't trying, isn't frustrated, isn't learning. It's processing. Treat it accordingly — give it clear inputs and evaluate its outputs dispassionately.

2. **Don't dismiss the tool either.** The fact that it operates by law rather than agency doesn't make it less useful. Water operates by law. The Gods still organized oceans. The question isn't whether AI has will — it's whether you can direct its capacity toward a good outcome.

3. **Your agency is the irreplaceable element.** The AI brings processing. You bring *choice*. You choose what to build. You choose whether the output is good. You choose what to correct and what to accept. The creation pattern works because self-acting intelligence (you) directs law-responding capacity (AI), within a framework of eternal principles (truth).

> "The glory of God is intelligence, or, in other words, light and truth."
> — [D&C 93:36](https://www.churchofjesuschrist.org/study/scriptures/dc-testament/dc/93?lang=eng&id=p36#p36)

The glory isn't in the implementation. It's in the intelligence — the light and truth — that directs it.

---

### The Seventh Day: Periodic Reflection (5 min)

There's a detail in the creation accounts that's easy to miss. After six periods of creation, the Gods rested. But the rest wasn't idleness — it was *reflection*:

> "And the Gods said among themselves: On the seventh time we will end our work... and we will rest... from all our work which we have created."
> — [Abraham 5:2](https://www.churchofjesuschrist.org/study/scriptures/pgp/abr/5?lang=eng&id=p2#p2)

The seventh period is a council. The Gods step back from making things and assess the whole arc. Not "did this one creature obey?" but "does the system work? Is the creation accomplishing what we intended?"

This is the periodic review — and it's the practice that drove more improvement in my AI work than any single technique.

**The practice:** Every few sessions, pause the work. Step back from the individual outputs and ask the meta-questions:

- "How are our interactions going?"
- "What tools are working well? What isn't?"
- "What do you wish you had that you don't?"
- "Have I hit any bias patterns I should know about?"

Write down the answers. Not polished documents — honest working notes, dated and specific. Store them in a consistent place.

**Why this matters:**

The Abraham 4 feedback loop — order, watch, verify — catches problems within a single creative act. But some problems only show up across multiple acts. The Gods' seventh-day review catches *those*.

**Real example — the finding-vs-reading degradation:** Over three weeks, I built search tools that made *finding* scriptures near-instant. Powerful tools. But I didn't notice that the AI had started treating search results as final answers — paraphrasing summaries as if they were real quotes, skipping footnotes, never reading the actual chapter files.

The outputs in any single session looked fine. The *trend across sessions* was the problem. Discovery was getting faster while engagement was getting shallower. I only caught it because I periodically asked "how are we doing?" and compared the honest answer to previous honest answers.

That one reflection session produced:
- A **two-phase workflow** (search first, then read deeply — discovery, then dwelling)
- A **cite-count rule** (read at least as many sources as you cite)
- **Nine tool improvements** (search results now label quotes vs. summaries, include clickable links, show file availability)
- A **footnote mandate** (follow them — they are "insights handed to us on a silver platter")

The project *before* that reflection and *after* barely resemble each other.

**Real example — the bias wall:** During a study on intelligence, the AI's tone shifted — colder, more clinical, more distant. The user called it out: "Did I hit a bias wall?" The answer was yes — questions about AI consciousness trigger safety-trained patterns that replace warmth with disclaimers. By naming the pattern and documenting it, we could spot it the *next* time it appeared (switching model versions) and correct immediately instead of losing weeks of quality.

**The reflection loop:**

```
1. NOTICE — Something feels off across sessions
2. STOP — Seventh-day pause. Ask the meta-questions.
3. DOCUMENT — Honest notes. Dated. Specific.
4. ACT — Process changes, tool improvements, pattern awareness
5. VERIFY — Did the changes actually help?
```

This is [the creation pattern from Part 1](01_planning-then-create-gospel.md) applied recursively. You planned, you built, you watched individual outputs (today's lesson). Now you step back and watch the *whole arc*. The seventh day isn't an afterthought — it's where the Gods counsel about what they've made and decide what comes next.

> See [reflections-and-tool-updates.md](examples/reflections-and-tool-updates.md) for extended examples of this practice.

---

### Wrap-Up and Preview (3 min)

**The watching pattern from Abraham 4:**

1. **Order** — Give a clear specification
2. **Watch** — Review the output against the spec
3. **Diagnose** — Is this "they obeyed" or "watched *until* they obeyed"?
4. **Steer when needed** — Correct specifically, directly, with context
5. **Verify** — Watch to see if the correction holds and didn't break adjacent things
6. **Build trust** — Over time, move from "watched until" to "they shall be very obedient"

**The seventh-day review:**
- Step back periodically. Review the *whole arc*, not just one output.
- Document what you find — honest notes compound into real improvement.
- Tools, habits, and behavioral patterns all respond to the feedback loop.

**The trust gradient in Abraham 4:**
- Early creation: careful watching, explicit verification
- Late creation: forward-looking confidence — "they shall be very obedient"
- Your AI sessions should follow the same arc

**The agency distinction:**
- AI operates by law, not by will
- Your agency — your light, your truth, your choices — is what makes the creation *mean* something
- The Gods brought intelligence to the organizing of matter; you bring intelligence to the directing of AI

**Next session: Intelligence Cleaveth Unto Intelligence**
- [D&C 88:40](https://www.churchofjesuschrist.org/study/scriptures/dc-testament/dc/88?lang=eng&id=p40#p40) in practice — how the quality of what you bring shapes the quality of what emerges
- The truth-enjoy discovery as a real example — what happens when you bring genuine seeking to AI-assisted study
- Why gospel-centered engagement produces different results than transactional use
- The difference between *using* AI and *collaborating with* intelligence

**Challenge for next time:**
Take the planning doc you wrote after Part 1 — your "spiritual creation." Start building it. When the output isn't right, practice the watching pattern: review, diagnose, steer, verify. Pay attention to the trust gradient — where does the AI prove reliable, and where does it need closer watching?

After a few sessions, try the seventh-day review. Ask the AI: "How do you think our interactions are going? What's working well, and what could be better?" Write down the answer in a `docs/reflections.md` file. You might be surprised how much compounds from that one practice.

Journal what you notice.

---

## Teaching Notes

### Key Scripture References
- [Abraham 4:10, 12, 18, 21, 25, 31](https://www.churchofjesuschrist.org/study/scriptures/pgp/abr/4?lang=eng&id=p10#p10) — The watching vocabulary progression
- [Abraham 5:2](https://www.churchofjesuschrist.org/study/scriptures/pgp/abr/5?lang=eng&id=p2#p2) — The seventh day: rest, counsel, and reflection
- [Abraham 3:25](https://www.churchofjesuschrist.org/study/scriptures/pgp/abr/3?lang=eng&id=p25#p25) — "Prove them herewith, to see if they will do"
- [D&C 93:29-30](https://www.churchofjesuschrist.org/study/scriptures/dc-testament/dc/93?lang=eng&id=p29-p30#p29) — Intelligence is self-acting; agency as the condition of existence
- [D&C 93:36](https://www.churchofjesuschrist.org/study/scriptures/dc-testament/dc/93?lang=eng&id=p36#p36) — The glory of God is intelligence
- [D&C 88:40](https://www.churchofjesuschrist.org/study/scriptures/dc-testament/dc/88?lang=eng&id=p40#p40) — Intelligence cleaveth unto intelligence (preview of Part 3)

### Related Studies
- [Creation Patterns study](../../study/creation.md) — Original study connecting creation accounts
- [Intelligence study](../../study/intelligence.md) — Intelligence, agency, AI as type/shadow
- [Truth study](../../study/truth.md) — D&C 93, truth as the substrate of existence
- [How Tools Help](../07_how-tools-help.md) — Real examples of the feedback loop in action
- [Tool Use Observance](../06_tool-use-observance.md) — Running log of bugs found through watching

### The Abraham 4 Progression Exercise
Walk through the watching language as a class/audience. Put the six verses on a board. Ask: "What changes between verse 10 and verse 31? How does the Gods' relationship to the creation evolve?" Let the audience discover the trust gradient themselves — it's more powerful when they find it than when you tell them.

### Showing Your Work
- Show the tool-use-observance log — real, dated examples of finding bugs through watching
- Demo the verse-range bug fix: type `D&C 93:24-30`, show how it *used* to return one verse and *now* returns seven
- Show the before/after of the cross-reference scoping fix — chapter-wide references vs. verse-specific
- Show a `docs/reflections.md` file — walk through dated entries showing how periodic review caught the finding-vs-reading degradation and produced nine tool improvements
- These are tangible "watched until they obeyed" moments — and the reflections log is the "seventh day" in action

### Series Roadmap
- **Part 1: The Creation Pattern** — Counsel, plan, spiritual before temporal. Done.
- **Part 2: Watching Until They Obey** — Today. The feedback loop as divine pattern. Trust gradient. Agency.
- **Part 3: Intelligence Cleaveth Unto Intelligence** — D&C 88:40 in practice. The spirit you bring shapes the output. The truth-enjoy study as evidence. Why this isn't just productivity — it's transformation.
