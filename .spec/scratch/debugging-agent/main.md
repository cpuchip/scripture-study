# Debugging Agent — Scratch File

## Binding Problem
When agent output goes wrong, when studies produce contradictions, when tools break, or when a system behaves unexpectedly — we need a *systematic* approach to diagnosis instead of blind retrying, guessing, or panic. Agans' "Debugging: The 9 Indispensable Rules" provides the framework. The question is: how do these rules map to our multi-agent scripture study ecosystem, and what does Moroni teach us about the epistemology of debugging?

---

## The 9 Rules (from the book)

1. **Understand the System** — Read the manual. Know what it's supposed to do before you try to figure out why it doesn't. Don't trust the manual blindly, but know what the designers intended.
2. **Make It Fail** — Reproduce the bug. You can't fix what you can't see. Automate reproduction if possible. Stimulate the failure (create conditions), don't simulate it (guess at the mechanism).
3. **Quit Thinking and Look** — Don't theorize without data. Actually observe what's happening. Instrument the system. See the failure, not just its symptoms.
4. **Divide and Conquer** — Narrow the search space. Binary search for the failure point. Good side / bad side. Split, test, narrow.
5. **Change One Thing at a Time** — Scientific method. Control variables. Use a rifle, not a shotgun. If a change doesn't fix, back it out.
6. **Keep an Audit Trail** — Write it down. What you did, in what order, what happened. The detail you think is irrelevant may be the key.
7. **Check the Plug** — Question your assumptions. The foundation/overhead factors (power, clock, init) get overlooked when you're deep in the details. "Is it plugged in?"
8. **Get a Fresh View** — Ask for help. A differently-biased person sees what you can't. Even explaining the problem to a mannequin (or an AI) can reorganize your thinking.
9. **If You Didn't Fix It, It Ain't Fixed** — Verify the fix. Take it out, see it break again, put it back in. It never "just goes away."

---

## Connections Found

### Moroni 10:4 — The Inverse Hypothesis

Moroni 10:4: "ask God, the Eternal Father, in the name of Christ, if these things are **not** true"

This is the most elegant epistemic move in scripture. Moroni doesn't say "ask if it's true." He says "ask if it's NOT true." This is:
- **Falsification** (Popper): A claim gains strength not by confirmation but by surviving attempts to disprove it.
- **Scientific method**: You don't prove a hypothesis; you fail to reject the null hypothesis.
- **Agans' Rule 9**: "If you didn't fix it, it ain't fixed." You haven't proven the fix works until you've tested it against the failure condition. You prove truth by trying to prove NOT-truth and failing.

Moroni is teaching debugging epistemology 1,400 years before Popper. The pattern: approach with real intent, ask honestly "is this NOT true?", and when the answer doesn't come — when the falsification attempt fails — you know.

This maps to several rules:
- **Rule 3 (Quit Thinking and Look)**: Don't just assume you know. Actually test. Get data.
- **Rule 5 (Change One Thing)**: Isolate the variable. Don't ask about everything at once.
- **Rule 9 (If You Didn't Fix It)**: Don't assume truth without testing the negative.

### Scientific Method = The Debugging Rules

Agans himself says it explicitly in Chapter 7: "This is just scientific method; in order to see the effect of one variable, the scientist tries to control all of the other variables."

The rules ARE the scientific method applied to broken systems:
1. Understand → Literature review
2. Make It Fail → Reproducibility 
3. Quit Thinking and Look → Observation / data collection
4. Divide and Conquer → Hypothesis testing via elimination
5. Change One Thing → Controlled experiment
6. Keep an Audit Trail → Lab notebook / documentation
7. Check the Plug → Verify experimental setup / assumptions
8. Get a Fresh View → Peer review
9. If You Didn't Fix It → Replication / verification

And from our truth.md study (Insight #3): "The Scientific Method Is a Form of Revelation" — "To discover truth through careful observation and testing is to do what the Gods do: comprehend the light."

### Stewardship Pattern Connections

From the stewardship study and covenant:
- **Rule 6 (Keep an Audit Trail)** = Covenant accountability. "Check existing work before making new claims." The agent writes everything down. This is stewardship — faithful record-keeping.
- **Rule 8 (Get a Fresh View)** = Council moment. Abraham 4:26 — "Let us go down and... take counsel among themselves." The debugging rule IS the council pattern: when you're stuck, get another perspective.
- **Rule 1 (Understand the System)** = "Read before quoting — always, everywhere, no exceptions." You can't fix what you don't understand. You can't quote what you haven't read.
- **Rule 7 (Check the Plug)** = Source verification. Are you even running the code you think? Are you quoting what the text actually says?
- **Rule 9 (If You Didn't Fix It)** = "Watched those things which they had ordered until they obeyed" (Abraham 4:18). Don't declare done until you've verified.

### Cross-Agent Applicability

**Study agent**: Rule 3 (look, don't just think) maps to "read before quoting." Rule 4 (divide and conquer) maps to the phased approach — isolate sections, verify each one. Rule 5 (change one thing) maps to testing one interpretation at a time.

**Eval agent**: Rule 3 directly — actually watch the video, don't just evaluate from the transcript. Rule 7 — check the plug: is the claim verifiable? Are the sources real? Rule 9 — if you haven't verified the correction, the correction isn't verified.

**Plan agent**: Rule 5 (change one thing) — don't try to fix everything in one proposal. Phase it. Rule 8 (get a fresh view) — the critical analysis phase IS this rule.

**Dev agent**: All 9 rules apply directly. This is their native domain.

**Lesson agent**: Rule 1 (understand the system) — understand the class, the learners, the context before designing the lesson. Rule 7 (check the plug) — are my foundational assumptions about what they know correct?

### The Principle That "Pops Out"

The unifying principle across all 9 rules is: **reality over narrative.** Every rule is a defense against the human tendency to construct a story about what's happening instead of looking at what's actually happening.

This is the same principle as "read before quoting." The same principle as "the darkness comprehendeth it not" — you can't understand what you refuse to look at. The same principle as Alma 32's experiment — "experiment upon my words" means *test*, not *assume*.

The debugging rules are the operational discipline of comprehension. The Gods comprehended the light *because they looked at it* — "for it was bright" (Abraham 4:4). They didn't theorize about the light. They comprehended it.

---

## Agent Design Decisions

1. **Not a standalone agent for most uses.** Debugging happens IN other workflows. The agent should be available as a mode when things go wrong, but the principles should also be referenceable by other agents.
2. **The agent should handle both technical debugging (tool/code failures) and intellectual debugging (studies that don't hold up, arguments with contradictions, sources that don't verify).**
3. **The 9 rules should be the skeleton.** Each rule becomes a phase or checklist item.
4. **Gospel connections should be woven in, not bolted on.** They're not decorative — they're the same principles expressed in a different sphere.
5. **Handoffs**: to dev (if the fix is a code change), to study (if the debugging reveals a deeper question), to plan (if the debugging reveals a systemic issue needing a proposal).
