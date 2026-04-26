---
name: worldbuilding-fiction
description: "Build fictional worlds (D&D, sci-fi, Bridge Simulator) that feel internally consistent and lived-in. Sanderson's Laws, cultural depth without lore-dumping, and the difference between iceberg and infodump. Load when creating a world, faction, city, magic/tech system, or culture."
user-invokable: false
---

# Worldbuilding Fiction

## Why This Exists

A world is believable when it feels *lived in* — when the reader senses there are streets they didn't walk down, debates they didn't hear, histories that happened whether or not the camera was there. A world is *forgettable* when it's a backdrop the writer painted in primary colors and never bothered to age.

The trap on the other end is *lore-dumping* — the writer who knows so much about their world they cannot stop showing it, and every chapter creaks under the weight of explanation. A world that explains itself is not lived-in. A world that *implies* itself is.

## Sanderson's Three Laws (and Why They Apply Beyond Magic)

Brandon Sanderson articulated three laws for magic systems. They generalize to any rule-bound system in fiction — magic, technology, politics, religion.

### First Law

> *An author's ability to solve conflict with magic is directly proportional to how well the reader understands said magic.*

If the reader doesn't understand the rules, magic (or technology, or political maneuver) cannot resolve plot. Resolution by unexplained power feels like the author cheating.

**For Bridge Simulator:** if the science officer reverses the polarity to save the day, the audience has to know — *before that moment* — that polarity is reversible, what it costs, and why this particular reversal is risky. Otherwise it's a deus ex machina with technobabble paint.

**Practical move:** seed the rule three scenes before it resolves anything. Mention it casually. Show its limits. *Then* let it matter.

### Second Law

> *Limitations > Powers. What a magic system cannot do is more interesting than what it can.*

A character who can do anything is a character with no story. A character who can do *one* extraordinary thing, with three things they cannot do, has plot in every direction.

- Lyra's seeds can sense corruption — but only living corruption. They cannot detect undeath, machinery, or political treachery. This is interesting *because* of what they can't do.
- A starship's transporter can move people — but not through shields, not over a certain distance, not while the target is in motion above a threshold. Every limit creates a story possibility.

**Practical move:** for every power, list at least three limits. Make sure the limits are knowable to the reader.

### Third Law

> *Expand what you have before you add something new.*

Most worlds get crowded too fast. The writer keeps inventing new magic schools, new alien species, new factions, instead of going deeper into what's already on the page.

A world with three factions, fully realized, is richer than a world with twelve factions sketched. A magic system with one mechanic, deeply explored, is more memorable than a system with eight schools each given a paragraph.

**Practical move:** before adding a new element to your world, ask: have I exhausted the storytelling potential of what I already have? Almost always: no.

## The Iceberg Principle

Hemingway: *the dignity of movement of an iceberg is due to only one-eighth of it being above water.*

The reader should sense the seven-eighths under the surface without ever seeing it. This is the difference between a world that feels real and a world that has been *explained at*.

**Above the waterline:** what the scene needs the reader to know to feel the moment.

**Below the waterline:** everything else — the history, the economics, the politics three valleys over, the religious schism in the next city. The writer knows it. The reader senses it. It is not stated.

You sense the iceberg in:

- A character glancing at someone and looking away. (There is history there. The writer knows what. The reader doesn't, but feels the weight.)
- A coin with an unfamiliar profile in the marketplace. (Some other realm exists. We don't visit it. We feel it.)
- A street name that nobody comments on but is clearly named for a thing that happened. (The Avenue of Lanterns. The Quiet Steps. The Forgiven Gate.)

**Practical move:** for every scene, write down three pieces of submerged context — things you know that the scene won't state. Their job is to color the scene's surface without surfacing themselves.

## Cultural Depth — The Five-Question Test

Most fictional cultures fail because they're a single trait amplified ("the warrior culture," "the trader culture," "the merchant guild"). Real cultures are bundles of contradiction. Run any culture through these five questions:

1. **What do they eat — and what do they refuse to eat?** (Refusal carries history.)
2. **How do they greet each other? How do strangers? How do enemies who have to be polite?**
3. **What do they sing about? What do their lullabies say?** (Lullabies reveal what a culture wants its children to know first.)
4. **What is shameful that another culture would find normal?** (And vice versa.)
5. **What was true a generation ago that is no longer true now?** (Cultures change. Static cultures are fake.)

You don't need to write all of this into the prose. You need to *know* it. The seven-eighths.

## Politics — The Working Test

A world's politics work if you can answer: *who benefits from the current arrangement, and who is paying the cost?*

If everyone is roughly fine, there is no politics. There is only scenery. Politics begins when there is *someone who is paying* and *someone who is taking*. The factions form along that line, even if they pretend otherwise.

For Bridge Simulator and similar:

- The Federation is funded by member worlds, but starship deployment skews to certain sectors. Which member worlds feel under-protected? Which feel over-policed?
- The trade lanes pass through two systems that never agreed to it. They take a cut. The Federation pretends not to see. Until something breaks.
- Every alliance has dissenters. Name three. Name what they want.

If you can name *who pays* in your world, your politics will write themselves. If you can't, your politics are decorative.

## The Spec File vs. The Prose

Worldbuilding lives in two places. Don't confuse them.

**Spec files** (`world/cities/baldurs-gate.md`, `factions/lords-alliance.md`) are *operational*. They contain everything the writer needs to know to write consistently across many sessions. YAML at the top, prose below. See `projects/DND/.spec/templates.md` for the canonical D&D format.

**Prose** is *experiential*. It contains only what the reader needs in this scene, this chapter, this session. The prose is the eighth above water. The spec is the seven-eighths beneath.

If a spec detail is interesting enough to be in the prose, *show* it (a glance, an object, a habit). If it's not, leave it in the spec. Don't put spec-language ("the Lords' Alliance was formed in 1370 DR by the merger of...") into prose. That's the worldbuilding sin.

## A Workflow for a New World

Building a new world (e.g., a fresh Bridge Simulator setting):

1. **The premise in one sentence.** "What if the Federation's prime directive was being secretly violated by its most decorated captain?" Stop at one sentence. Resist epic.
2. **Three factions.** Name them. Name what each wants. Name who each is taking from. Stop at three.
3. **One central tension.** What is the political/cultural fracture point that will drive most stories in this world? (Federation idealism vs. realpolitik. The colony's secession movement. The two royal houses sharing one throne.)
4. **One rule-bound system.** Magic, technology, religion — pick one. Apply Sanderson's three laws to it. Limits before powers.
5. **Five locations.** Capital, frontier, contested zone, sanctuary, ruin. One sentence each.
6. **Five characters.** A power-holder, a dissenter, a striver, a survivor, a wildcard. Don't fully build yet — just sketch.
7. **The iceberg.** Write down ten things you know about this world that the reader will never be told directly. These are the depth.

That's the foundation. Stop. Don't keep building until the players or readers have shown you what they're curious about. Worldbuilding from the audience's questions is always richer than worldbuilding in a vacuum.

## What NOT to Do

- ❌ Don't write a 30-page "world bible" before writing anything that anyone will read. The bible is a procrastination disguised as preparation.
- ❌ Don't lore-dump in dialogue. ("As you know, the Treaty of Three Crowns was signed during the Long Winter, when...") No one talks like this. Find another way.
- ❌ Don't make every culture have one trait. Real cultures contradict themselves.
- ❌ Don't add a new faction every time the plot needs a new wrinkle. Use the factions you have. Sanderson's Third Law.
- ❌ Don't explain the magic/technology in narration. Show it under pressure. The reader will infer the rules from the constraints.
- ❌ Don't make the world's history exposition. Make it *felt* — through ruins, scars, named streets, a song everyone knows but nobody explains.

## Cross-Reference

When the world is built, characters and villains are populated using `character-voice` and `believable-villains`. Sessions and chapters are written using `emotional-resonance` and (when applicable) `sacrifice-and-loss`. The world is the table the rest of the work sits on.
