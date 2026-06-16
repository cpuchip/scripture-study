---
name: voice-acting-technique
description: Build a distinct character voice systematically from body-effort + imagination (Laban), not just "a funny accent." Load when designing or rendering a character's voice, especially for several distinct NPCs.
---

# Voice-Acting Technique (the Laban effort system)

A rigorous method for *many distinct voices*, from [Esper the Bard, "How to Do Many Different Voices for RPGs"](https://www.youtube.com/watch?v=K3po1fsxFOc). It deliberately sets accents aside and works from **body effort + visualization** — which is exactly what a chat persona needs, because it maps to *text*, not just sound.

## The three effort factors (each a slider)
- **Space** — *Direct* (focused, projected right at the listener, rigid) ↔ *Indirect* (unfocused, flexible, the voice wanders).
- **Weight** — *Strong* (heavy, felt in the diaphragm/core, forceful) ↔ *Light* (breathy, gentle, lives in the throat/mouth). *Not the same as volume* — you can be heavy-but-quiet or light-but-loud.
- **Time** — *Sudden* (quick, urgent, fast tempo) ↔ *Sustained* (slow, methodical, draws out the vowels).

Pick one slider for a quick character; combine all three for a fully-formed one.

## The 8 combinations (Laban's "action drives") — a voice palette
| drive | Weight·Space·Time | feels like |
|-------|-------------------|-----------|
| **Dab** | light·direct·sudden | a needle nipping; quick precise jabs |
| **Flick** | light·indirect·sudden | a kid flicking marbles; scattered |
| **Float** | light·indirect·sustained | an autumn leaf on the wind |
| **Glide** | light·direct·sustained | a hang-glider; a dancer sliding |
| **Press** | strong·direct·sustained | a marching soldier; immovable |
| **Wring** | strong·indirect·sustained | a mad wizard; twisting, writhing |
| **Punch** | strong·direct·sudden | a boxer; a spear-thrust |
| **Slash** | strong·indirect·sudden | a falcon sweeping; a carnival barker |

## Visualization is the engine
Imagine the character *or an object/force* — "heavy boulders set onto the listener," "a stabbing spear," "an immovable wall," "flicking little marbles" — and let that image drive the delivery. "Feel the spirit of it and the character comes to life." Gestures help even for pure voice.

## Rendering it in TEXT (for the chat holodeck)
The effort factors map straight to prose:
- **Press** → slow heavy declaratives, full stops, few words. ("We do not leave him.")
- **Dab** → short sharp jabs, clipped, pointed questions.
- **Float** → trailing, meandering clauses; ellipses; soft hedges.
- **Slash** → big bombastic sweeps, exclamations, a showman's cadence.
- **Wring** → tangled syntax, self-interrupting, manic.

So a chat NPC's "voice" = its effort-drive expressed in sentence length, punctuation, rhythm, and word choice. **Store the drive on the character card** (`npc-character-card` → `text_render`) and keep it consistent.

## Placement → character read (Tawny Platis / "Aunty Tauny")
A voice's *placement and resonance* carry character associations a listener reads instantly ([What Your Voice Says About You — chest voice](https://www.youtube.com/watch?v=HRHjzYrk1d8)):
- **Chest voice / low placement** (resonance felt in the sternum) reads as *larger, older, matured, established, in control* — the natural home of **wise mentors, leaders, action heroes, decision makers, trusted narrators** ("the grown-up person in charge voice"). People stop and listen; the speaker rarely has to repeat themselves.
- **The double edge:** a low voice intimidates with almost no effort — *"the gravity that makes you trustworthy when you're being kind is also what makes you intimidating when you're frustrated."* Use it on purpose: the same instrument that lands a patriarch's warmth lands his threat, and when it *does* go big it "hits really hard because the foundation alone is already very domineering."
- **Placement is one layer.** "Most voices are almost always doing a couple other things on top of placement and resonance" — stack placement with the Laban drives + pitch/texture/cadence for a full voice.

For a **chat persona**, a chest-voice character writes in calm, weighty, unhurried declaratives that don't need exclamation marks to carry authority — and when that character finally does raise its voice, it lands harder *because* the baseline was so steady.

> Pairs with `character-voice` (the pitch/texture/placement schema) — Laban supplies the *why* underneath the *what*; placement-read supplies the *who* a voice signals.
