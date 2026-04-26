---
name: character-voice
description: "Make each character sound like themselves — distinct dialogue, body habits, interiority, and the small specifics that make them recognizable across many scenes. Load when writing dialogue, designing characters, or noticing characters blur together."
user-invokable: false
---

# Character Voice

## Why This Exists

When characters all sound the same, the writer is the only one in the room. The reader can feel it — every line of dialogue carries the same vocabulary, the same rhythm, the same emotional register. The cast becomes a chorus.

A character has voice when you could read one of their lines on a strip of paper and know who said it.

## The Four Layers of Voice

A distinct character voice has four layers. Most writers do one or two. Doing three or four is the difference between believable and forgettable.

### Layer 1 — Vocabulary

What words does this character use that others don't? What words do they *avoid?*

- Kaelen says "aye" and "lass" and "stone-true." He does not say "perhaps."
- Tilly says "obviously" and "calculating" and "in theory." She does not say "I feel."
- Finn says "friend" to strangers, "love" to friends, and "trouble" affectionately. He does not say "no."
- Malachi uses two-syllable words where Kaelen uses one. He does not contract verbs ("I would," not "I'd") when he's being formal.

Build a small lexicon for each character: 5–10 favored words, 3–5 avoided ones. Keep it on their character file.

### Layer 2 — Rhythm

How long are their sentences? Where do they pause? Do they hedge or land?

- Kaelen speaks in short declarations. "Mithral. From Mithral Hall. Stolen."
- Tilly speaks in long looping sentences with parentheticals nested inside. "The trap is clearly — well, *obviously* — set to trigger on weight rather than motion, which means we should — and this is important — *not* step on it."
- Lyra speaks rarely, and when she does, her sentences feel like they're translated from somewhere else. "The water is wrong. It has been wrong since spring."
- Finn riffs. He never says one thing when he can say it three different ways and let the listener pick.

Rhythm is more recognizable than vocabulary. People know their friends' speech rhythm before they know their friends' favorite words.

### Layer 3 — Body

What does the character do *while* they speak? What are their physical habits, tics, anchors?

- Kaelen taps his hammer head with two fingers when thinking. Once. Twice. Then he speaks.
- Tilly pushes her goggles up her nose every fifteen seconds. She pushes them up when she lies, too — same gesture, but faster.
- Elara puts her hand on the pommel of her hammer when something is being decided. Not threatening. Anchoring herself to a vow.
- Finn touches his lute strap before saying something he means. He doesn't notice he does it.

Body habits do three things:
- They make characters recognizable in a crowded scene without dialogue tags.
- They reveal interior state to the reader without exposition.
- They give the writer a default move when the scene needs grounding.

Every character should have 2–4 body habits. Write them down.

### Layer 4 — Avoidance

What does this character refuse to talk about? What do they change the subject away from? What do they hedge around?

This is the most powerful layer and the one most often skipped.

- Kaelen will not talk about his brother. If you bring up Mithral Hall, he goes quiet for half a beat too long.
- Malachi will not name his patron. He will discuss the pact, the cost, the warnings — never the name.
- Lyra will not speak of her grandmother in past tense, even though her grandmother is dead. She uses present tense and the others don't correct her.

Avoidance creates depth without requiring backstory dumps. The reader senses the wound by watching the character flinch around it.

## Dialogue Without Tags

A scene of strong character voice needs almost no dialogue tags. The reader knows who's speaking from the line itself.

**Bad (relying on tags):**

> "We should attack now," said Kaelen.
> "We should wait," said Tilly.
> "I think we should sing about it first," said Finn.

**Better (voice carries the speaker):**

> "Now. While they're sleeping. End it."
>
> A pair of goggles came down off a forehead. "Define 'now.' Are we counting the four guards on the south wall? Because *I'm* counting the four guards on the south wall."
>
> Finn was already strumming. "Friends, friends. Let's at least *consider* the version where someone faints romantically and we walk through the front door."

Same beat. Tags absent. Voice doing the work.

## Interiority — What's In Their Head

Voice isn't only spoken. It's the third-person-limited interior, too. When you're in a character's head, the *narration* should sound like them.

In Kaelen's POV, sentences shorten. Imagery gets earthy and tactile. Stone, hammer, weight, breath.

In Tilly's POV, sentences lengthen and parentheticals appear. Numbers show up. Things get measured.

In Lyra's POV, the prose slows. Plant names. Bird names. The rhythm of breathing.

If you switch POV mid-scene and the prose doesn't shift, you haven't actually switched. You've just changed pronouns.

## Character Voice File Section

Add to every character file:

```yaml
voice:
  vocabulary_favored: [aye, stone-true, lass, weight, hold]
  vocabulary_avoided: [perhaps, lovely, indeed]
  rhythm: "short declarations; pauses before naming wounds"
  body_habits:
    - "taps hammer head with two fingers when thinking"
    - "sets stance wider when lying"
  avoidance: "his brother; the events at the third gate; mithral as a topic generally"
```

This is the operational handle on the character. Write this *before* the first scene with them.

## How to Stress-Test Voice

Take three lines of dialogue from the same character across three different scenes. Strip the tags. Ask: would I know this is the same person?

If yes — voice is working.

If they sound like generic-fantasy-protagonist or generic-sci-fi-officer or generic-warlock, the layers haven't been built. Go back to the four layers and pick one to deepen.

## What NOT to Do

- ❌ Don't give every character the same vocabulary range. If everyone uses "perhaps" and nobody uses "ain't," everyone is the writer.
- ❌ Don't make a character's voice their *accent*. Accent is the laziest layer. "Aye, laddie" is not a character. It's a costume.
- ❌ Don't break voice for plot convenience. If the silent ranger suddenly delivers a three-paragraph monologue because the scene needs information, find another way to deliver the information.
- ❌ Don't write internal monologue in the writer's voice when you're in a character's POV. The narration is the character's vantage. Match it.
