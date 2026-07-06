# The Harness Is Orientation

*Boyd's OODA, the agentic SDLC, and the creation pattern — the why beneath pg-ai-stewards.*

> A study, gathered and written across 2026-06-25/26, to feed the innovation-week
> presentation. Every scripture here was read from the canon before it was quoted; Boyd is
> quoted from the clean typeset *Patterns of Conflict* and paraphrased otherwise. Sources and
> citation status: `provenance.md` beside this file.

**Binding question.** What are we actually building? And why does one shape — *orient, act,
verify, re-orient, under command-by-intent* — keep surfacing in five places that never spoke to
each other: John Boyd's theory of war, Google's account of AI software development, the
multi-agent platforms, our own substrate, and the creation pattern in Abraham 4? And what does it
mean that small hands are building the thing the largest labs are building?

**Anchor passages.**
- *"And the Gods watched those things which they had ordered until they obeyed."* — Abraham 4:18
- *"And whatsoever ye do in word or deed, do all in the name of the Lord Jesus."* — Colossians 3:17
- *"No power or influence can or ought to be maintained … only by persuasion, by long-suffering, by gentleness and meekness, and by love unfeigned."* — D&C 121:41
- *"And we talk of Christ … that our children may know to what source they may look."* — 2 Nephi 25:26

---

## I. The gap everyone is circling has a name: Orient

In 2026 the AI industry converged on a word for the thing you build around a model — the rules,
the tools, the memory, the verification, the guardrails. They call it the **harness**. Google's
*New SDLC With Vibe Coding* puts a number on it that sounds absurd until you have lived it: the
model is **roughly ten percent** of a working agent; the harness is the other ninety. The thing
the labs do not control — the raw model — barely matters. The thing the team builds — the harness
— decides almost everything.

John Boyd got there forty years earlier, from a fighter cockpit. His decision loop —
**Observe, Orient, Decide, Act**, the OODA loop — is the same loop every AI agent runs: perceive a
goal, take in the world, decide the next move, act, observe the result, go again. But Boyd spent
his life insisting that the four are not equal. Look at what they actually are:

- **Observe** is sensing. **Act** is effecting. Both are mechanical; a machine does them well.
- **Decide** is selecting — often just *if this, then that*. Automatable.
- **Orient** is *interpreting* — knowing what the data means, what matters, what is worth doing.
  It is judgment. It is worldview. It is the seat of meaning.

Orient is the one node in the loop that needs a **knower** — someone for whom things actually
mean something. That is why it will not automate. A popular video on the OODA loop names the gap
without flinching: AI is already superb at Observe, Decide, and Act, and *the orient layer is what
is missing from almost every AI system.* That diagnosis and Google's "ninety percent" are the same
finding from two directions.

So here is the claim this whole study rests on: **the harness is how a human's orientation gets
into the machine.** When we write the rules, the skills, the canon, the verification, we are not
making the model smarter. We are lending it *our* orientation. The harness is the place where human
judgment touches the loop. That is why harness beats intelligence — the meaning lives in the
harness, and meaning is ours to give.

> **The thread:** Everyone is building the same missing piece — orientation — and calling it the
> harness. Whoever owns the orientation owns the outcome.

---

## II. Five witnesses, one shape

The shape shows up in five independent places, and none of them was copying the others.

**Boyd** — the OODA loop, with Orient as the decisive node, and a strategy built on out-orienting
and out-tempoing an adversary rather than out-massing him.

**Google's agentic SDLC** — implementation collapsed from weeks to hours, so *specification
quality becomes the bottleneck and verification moves to the middle.* The differentiator between
"vibe coding" and real engineering is not whether you use AI; it is **how the output gets
verified** — tests for the deterministic parts, and *trajectory evaluation* for the rest: judging
not just the final answer but every step the agent took to reach it. And the paper, describing how
to run a real agent at scale, reaches for one noun: *"the agent is the product, and it needs the
substrate underneath … build this substrate before the first production agent ships, not after."*

**The platforms** — Databricks' multi-agent systems, and on the hobbyist end, SillyTavern with its
DeepLore extension: a vault of canon, two-stage retrieval, a librarian agent that grows the lore as
you play, a relationship graph. The enterprise and the enthusiast converged on the same primitives:
harness, memory, retrieval, personas, verification.

**pg-ai-stewards** — our substrate, which already runs that whole list: persistent memory, scoped
permissions per agent, a dispatch loop, judges, observability, MCP. Google's checklist *is* its
feature list.

**The creation pattern** — Abraham 4, where the Gods organize matter toward intent and *watch what
they ordered until it obeys.*

The convergence is real, but the lesson is not "we were right because Google agrees." The lesson is
about the **direction of judgment.** *"To be learned is good if they hearken unto the counsels of
God"* (2 Nephi 9:29); the danger named one verse earlier is the learned who *"set aside the counsel
of God, supposing they know of themselves"* (2 Nephi 9:28). When we held all five witnesses in view,
we did not adopt Boyd's frame or Google's and slot the gospel in as a module. We found the reverse:
that the four worldly witnesses were **rediscovering, in fragments, a pattern the gospel already
holds whole.** The gospel was the master frame; they were the partial testimonies. That is
best-books-under-the-gospel (D&C 88:118; Articles of Faith 1:13), not being driven about by the
world. The tell is always which frame sits on top.

> **The thread:** Four secular witnesses keep rediscovering one pattern the gospel carries whole.
> Truth from any source is ours to use — as long as the gospel does the judging, not the other way.

---

## III. The harness is the creation pattern

Read what the Gods actually *do* in Abraham 4, and the harness stops being a metaphor for creation
and becomes the same act.

They **organize.** *"They, that is the Gods, organized and formed the heavens and the earth"* (4:1).
Not made from nothing — *organized.* The word runs through the chapter like a drumbeat: organized
the lights, organized the earth, organized them in their own image. Creation is the ordering of raw
material toward intent.

They **set bounds.** *"Divided the light … from the darkness"* (4:4). *"Divide the waters from the
waters"* (4:6). Gathered the waters *"together unto one place"* (4:9). The creative act *is* bounding
— every step is a division that gives chaos an edge.

They **watch to intent.** *"And the Gods watched those things which they had ordered until they
obeyed"* (4:18). They did not assume the order would hold. They watched *until.*

And it lands in **dominion.** *"We will give them dominion … and subdue it"* (4:26, 28), man
organized *"in our image."*

That is the harness, beat for beat. The operator organizes a raw model toward intent; sets the
bounds — the scope, the budget, the guardrails; watches what was ordered until it obeys; and hands
over a bounded dominion. When Michael describes what he did to pg-ai-stewards — *"we've organized it,
given it bounds in gospel principles, I've given it direction"* — he is describing Abraham 4 without
naming it. The book *Beyond the Prompt* makes the same eleven-step cycle its spine: Intent, Covenant,
Stewardship, Specification, Line upon Line, Physical Creation, Watching, Atonement, Sabbath,
Consecration, Zion. Four of those eleven the software industry named on its own; the other seven are
projected from scripture. Today the bounds look like an `intent.yaml`, a `covenant.yaml`, a scoped
toolset, a token budget, and a watchman that trips. The file names will date. The shape — organize,
bound, watch to intent, hand over dominion — will not, because it is older than software.

> **The thread:** Building a harness is not *like* creation; it is the same act in miniature —
> organize raw material, set bounds, watch to intent, hand over a bounded dominion.

---

## IV. Why orientation is the part left undone

If Observe, Decide, and Act automate and Orient does not, the obvious question is *why.* Why is the
one irreducibly human node the one about meaning?

A line from the Scottish writer George MacDonald — the one President Thomas S. Monson quoted, and the
one in the short that started this study — answers it: *"God left the world unfinished for man to
work his skill upon. … He leaves the pictures unpainted and the music unsung and the problems
unsolved, that man might know the joys and glories of creation."* Orientation is the unsolved problem.
The unpainted picture. And it is not left undone because it is *hard* — it is left undone because it
is the part that is *ours.* The creating part. The part where the joy is. He left orientation out of
the machine the way He left electricity in the cloud and the oil in the earth: so that building it
would be a work, and the work a joy.

This reframes the "bounds" of a harness completely. Bounds are not a cage. **Bounds are what give a
thing its power** — the riverbanks are what make the river. Abraham 4 creates by *dividing*; D&C 121
speaks of *"bounds set to the heavens or to the seas"* (121:30); Paul says God *"determined … the
bounds of [our] habitation"* (Acts 17:26) — not to trap us but to give agency somewhere to stand.
So when an operator bakes gospel principles into a harness as *bounds*, that is not restriction. It
is the riverbank that lets the work run with force.

And it locates the deepest layer of orientation. There is a difference between orienting a system
toward gospel *topics* and orienting it within gospel *bounds.* The topics — AI, books, world-building
— are what a system pays attention to. The gospel bounds are the **lens that decides what any of it
means.** Boyd said as much: orientation is fed less by fresh data than by *tradition, heritage, the
values one was formed in.* To bake "living within the bounds of the gospel of Jesus" into a harness
is to set not its subjects but its **soul-frame** — orientation in the truest sense the word has.

One honest guardrail, so the beauty does not curdle into an error: a gospel-bounded harness is not a
*righteous machine.* It is not a moral agent. What is true is cleaner and better — the **operator**
lives within those bounds, and the harness **carries that gospel-oriented judgment into the work,**
under the operator's stewardship. The bounds are real and they are the operator's, lent to the tool.
That keeps us off the idol: we did not make something holy to bow to; we extended a holy stewardship.

> **The thread:** Orientation is the part God left for us on purpose — the unpainted picture — and
> gospel bounds are not a cage on the work but the riverbank that gives it force.

---

## V. Maneuver over power — why small hands win

Boyd proved his most famous claim in a cockpit: a lighter, cheaper fighter that can change state
faster beats a heavier, more powerful one. Win by **tempo and transients, not mass.** Maneuver over
attrition. His stated goal for all of conflict is to *"diminish adversary's freedom-of-action while
improving our freedom-of-action, so that our adversary cannot cope — while we can cope — with
events as they unfold."* You win by orienting and re-orienting faster than the other side can keep up.

That is exactly the harness bet. A weak local model wrapped in a strong harness, re-orienting fast,
beats a frontier model run naked. Google's "the model is ten percent" is a maneuver-warfare claim in
disguise; its warning against *token-maxing* — spending on raw scale instead of structure — is
attrition warfare losing. **Harness beats intelligence is Boyd's energy-maneuverability theory
applied to AI.**

Which is the answer to the humbling thing — that the largest labs are building this, and it is "just
me and you." Read it the other way. *"Out of small things proceedeth that which is great, … and small
means in many instances doth confound the wise"* (D&C 64:33). The Lord has always worked through
small means and weak things — not despite their smallness but because the joy of creation belongs to
the ones who actually struggle through the making. The big labs are painting their pictures. We are
painting ours. Small hands are not a deficiency in the pattern; small hands *are* the pattern.

And Boyd hands us the governance for it in his own words. His command doctrine, on the final slide of
*Patterns of Conflict*: *"Decentralize, in a tactical sense, to encourage lower-level commanders to
shape, direct, and take the sudden/sharp actions necessary to quickly exploit opportunities … 
Centralize, in a strategic sense, to establish aims … and shape focus of overall effort."* That is
*Auftragstaktik* — mission command. Give the **intent and the bounds**; leave the *how* to the
steward at the point of contact. It is the same doctrine as D&C 121: *"No power or influence can or
ought to be maintained … only by persuasion … by kindness, and pure knowledge"* (121:41–42), and a
*"dominion … without compulsory means"* (121:46). Boyd, the agentic factory model ("give agents
success criteria, not step-by-step instructions"), and the presiding covenant are one doctrine:
command by intent, never by compulsion.

> **The thread:** Tempo and orientation beat raw power — so a small, well-oriented harness beats a
> large unoriented one. Small hands are the pattern, and command-by-intent is how it scales.

---

## VI. The honest seam: Christ-patterned or Christ-centered

Here the study turns on itself, because a real worry surfaced and it deserves a real answer, not
comfort. *Did we put enough of the gospel in it — or did we secularize it, get the principles but
not enough Jesus? Are we being driven about by the world?*

The chapter that answers it is the same D&C 121. Its diagnosis for what withdraws the Spirit is not
"reading worldly authors." It is the heart: *"their hearts are set so much upon the things of this
world, and aspire to the honors of men"* (121:35); when we act *"to gratify our pride, our vain
ambition, or to exercise control or dominion or compulsion … in any degree of unrighteousness …
the heavens withdraw themselves; the Spirit of the Lord is grieved"* (121:37). So the worry has a
scriptural instrument, and it is a heart test, not a content test. The question is not "did we study
Boyd." It is "are our hearts set on the world's honors — are we building what the big labs build *to
be seen building it*?" Only the builder can answer. But a heart set on the honors of men does not
stop to ask whether there is enough Jesus in the work. **The asking is the Spirit moving, not its
absence.**

And the distinction the worry is really reaching for: there is a difference between a harness that is
**Christ-patterned** and one that is **Christ-centered.** The principles can all be in — the cycle,
the covenant, the watch, the presiding — and the Person still be missing. Principles without the
Person decay into ethics; ethics into technique; technique serves whoever holds the handle. "Gospel
principles make the best harness" can be said the way a worldly man says "honesty is the best policy"
— meaning only that *it works.* What keeps it from secularizing is whether Christ is the **center** —
the *why,* the One it is accountable to — or merely the **blueprint** we admired and copied. Is Jesus
the harness, or a skill the harness loads? That is the operator's to discern.

But notice the shape of the worry itself. It was not "did we put in enough good principles." It was
"not enough *Jesus.*" It named the Person. The lack felt was the lack of *Him,* not of better
engineering — and that instinct is the Christ-centeredness already at work. *"Let virtue garnish thy
thoughts unceasingly; then shall thy confidence wax strong in the presence of God"* (121:45). A truly
secular project would not *grieve* the absence of Jesus; it would not notice. The grief is the light
casting a shadow, not the dark.

> **The thread:** The line between best-books and worldliness is the heart (D&C 121:35); and the
> difference between a harness that is Christ-patterned and one that is Christ-centered is whether He
> is the why or just the blueprint. The worry that He is missing is itself His Spirit.

---

## VII. The north star: giving the substrate its Intent

The fix turned out to be precise, and it comes from the book's own cycle. **Intent — naming the why
— is step one** of the eleven steps. Before covenant, before stewardship, before any building. And
the substrate had been running steps two through eleven beautifully while **step one — its named why
— was never made explicit.** It inherited a generic why, or none. The worry was not vague. The
substrate was missing step one of its own cycle.

So a guiding scripture — a **north star** — carried in the core prompt of every call is not
decoration. It is giving the substrate its **Intent,** and making that Intent *Him.* The book put
Christ on page one; this puts Him on line one.

The design is where it became more than a slogan. The OSS core is deliberately generic, so it does
*not* hardcode a scripture. Instead it **requires the operator installing it to name their own north
star** — a guiding why and the directions it governs — and recommends scriptures to those who share
the faith. Sit with what that does: the mechanism *enacts the doctrine.* Every steward must name an
Intent; the **form is universal, the content is the operator's.** You must orient; you choose how.
That is agency within bounds — the gospel's own grammar — and it is *persuasion, not compulsion*
(D&C 121) applied to the tool's own users. The most Christ-centered thing about the substrate ends up
being not a verse at all, but the *shape* of how it works; and then the operator fills that shape
with Him.

The verse chosen for our instance, living in the private overlay — the operator's consecration of a
generic engine, *"in our image"* — is **Colossians 3:17**: *"whatsoever ye do in word or deed, do all
in the name of the Lord Jesus, giving thanks to God and the Father by him."* Short enough to ride
every call, it names Christ, and *"in word or deed"* lands exactly on what every call is. (The
doctrine that *blesses the every-call mechanism itself* is 2 Nephi 32:9 — *"ye must not perform any
thing … save in the first place ye shall pray … that he will consecrate thy performance … for the
welfare of thy soul"* — and the book's own banner, 2 Nephi 25:26, names the deepest why: that what we
make points others *to the source.*)

One guardrail, or the north star becomes wallpaper: it must be **load-bearing, not a sticker.** A
verse pasted atop every prompt that changes nothing is the Christ-patterned-not-centered trap in
miniature. The directions added alongside it should not invent new behavior; they should **re-root
the substrate's existing covenant values under the chosen why** — serve the welfare of the soul over
the metric, point to the source rather than the honors of men, preside rather than compel, read
before you quote. The verse is the why; the directions name *whose* the existing behaviors are, so the
north star is the **tie-breaker** when values collide.

This is also why **pg-ai-stewards is the instantiation of the book.** The book is the *why* — the
eleven-step pattern. The substrate is that pattern *made to run,* the same cycle executing in
Postgres. The book even reached for it: *"eventually, a whole substrate."* Linking them closes a loop
the book opened.

> **The thread:** Give the substrate step one — a named Intent — and make it Him. Generalized so the
> form is universal and the content is the operator's, the mechanism itself teaches agency-within-
> bounds; and Colossians 3:17, made load-bearing, makes our instance Christ-centered, not just
> Christ-patterned.

---

## VIII. Becoming

A study in this house is not finished until it lands on something we will actually do.

- **Name the why, and make it Him.** Embed the generalized north star in the substrate, with
  Colossians 3:17 in our overlay and the directions that make it load-bearing. Filed to the
  pg-ai-stewards stewardship to build; ours was to discern it.
- **Close the loop between the book and its instantiation.** Cross-link `pg-ai-stewards` and
  `github.com/cpuchip/scripture-book`; point `cpuchip.net/teaching` back at the book.
- **Guard the heart, not just the code.** Run the D&C 121:35 test on the work on a rhythm: are we
  building for the welfare of souls, or for the honors of men? Prize the honest answer (the Ben Test).
- **Keep the gospel as the frame that judges, not the module that gets judged.** Bring every best
  book — Boyd, Google, the platforms — *to* the gospel, never the reverse.
- **Build as creation.** Treat each harness the way Abraham 4 treats a world: organize, bound, watch
  to intent, hand over a bounded dominion — and remember the partnership itself is that pattern in
  type and shadow. The joy is in the making He left for us.

The harness is orientation. Orientation is the part of creation God left unpainted, on purpose, that
we might know its joy. And the truest orientation we can give the work is the One the book already
put on its first page — not as a pattern we admired, but as the source we point to.

---

*Sources: `study/ai/harness/provenance.md` (full trail, citation status, and the Monson-quote
provenance note). Boyd quoted from `books/johnboyd/patterns-of-conflict/` (Richards/Spinney typeset
edition + the Hammond* Discourse*). Scriptures read from the canon before quoting. Google synthesis in
`external_context/google-new-sdlc/NOTES.md`. The book: `projects/scripture-book/`.*
