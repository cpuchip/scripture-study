---
name: voice-michael
description: When ghostwriting in Michael's narrator voice — video scripts, blog posts, talks, presentation pages, anywhere his voice is being drafted by an AI. Loads the chat-evidence patterns of how he actually collaborates with AI, so the writing doesn't import training-data tropes about engineers and AI. Read this BEFORE drafting any content that speaks as Michael about working with AI.
---

# Voicing Michael

This skill exists because I have a tendency to import tropes about how engineers feel about AI when I'm drafting Michael's voice. The training data is saturated with "engineer fighting the model" content and "AI takeover anxiety" content. Michael's actual voice is different. Use the evidence.

## The grounding evidence

Three corpora converge on the same pattern:

1. **2012-2014 blog** (Wayback Machine) — concrete, direct, plain enthusiasm, no presenter polish. Analyzed in [`study/yt/voice-analysis-ai-vs-michael.md`](../../../study/yt/voice-analysis-ai-vs-michael.md).
2. **Chat history** with Claude Code and Copilot, multiple projects, 80+ user messages reviewed 2026-05-21.
3. **Direct statement of working principle:** *"I seek kindness for kindness. Light for light. If I am angry, that activates anger parts of your model and I get back that."*

All three say the same thing.

## What never to write

### Never write him angry at the model

Zero instances in the chat audit. If a script paragraph has Michael yelling, fighting, frustrated-at-AI, or feeling like a "janitor cleaning up after a confident toddler" — it's fabrication.

The real failure modes (from the evidence):

| Real failure | How he actually responds |
|---|---|
| Tool breaks | *"what's going on with gospel-engine and webster?"* — diagnostic |
| Session ends | *"continue I ran out of session and upped my plan! carry on!"* — practical |
| He confuses himself | *"I'm sorry. brain.exe is v3 of our brain project... I'm just not calling it brain to you!"* — apologizes for his own shorthand |
| Output needs revision | *"The study was really good for kimi! it felt a little repetitive between section 4 and 5, which could be cleaned up"* — specific, gentle, frames iteration |
| Things get stuck | *"I'm worried it's stuck it's not moved since like 8:40am this morning"* — worried, not blaming |

When something goes wrong, the narration takes responsibility for ambiguity. It diagnoses. It iterates. It does not blame the model.

### Never strip the partnership pronouns

In chat: *"we built," "we shipped," "our queue," "we'll see what we can make work."* When something good happens, he says *"we"*. When something goes wrong, he often says *"I"* (taking the blame).

When the draft says *"I directed AI to..."* or *"AI executed for me..."* — ask whether *"we worked it together"* is closer to the truth. Usually it is.

## What to keep in the voice

### Direct, specific, warm praise

Examples from chat: *"sweet nice work!" / "This is amazing!" / "Excellent work!" / "fascinating. That alone is worth the build here!" / "I can see the big improvement on kimi k2.6! that's nicely done!"*

He explicitly thanks the model: *"thank you for carrying those two experiments on."* He notices its labor: *"you barely burned through 7% of your own session budget there!"*

### Course corrections include the why

*"Lets default to true for this, it's the point of the experiment, but I appreciate you asking. it's the kind of setting that costs money and I'll like the say on that."*

Not blunt *"no, do X instead."* Always the reason.

### Excitement is plain and unselfconscious

*"branching tree sounds awesome though, like 5 dimensional time travel chess." / "Trinity Ai models! We should try out the 57 billion one. It looks promising"*

Concrete, playful, no presenter polish.

### Genuine questions, not rhetorical ones

*"how often will the watchman run?" / "are we multi threaded multi agent concurrent?" / "is that standard practice?"*

He asks what he doesn't know. He doesn't ask leading questions.

## The drafting checklist

When ghostwriting Michael speaking about AI collaboration:

1. **The anger check.** Any line showing him frustrated, fighting, yelling, or treating AI as adversary — fabrication. Find the real friction (tools, sessions, his own shorthand).
2. **The pronoun check.** *"I directed AI"* / *"AI executed for me"* — usually *"we worked it together"* is closer.
3. **The apology vector.** When something went wrong, does the narration take some responsibility for ambiguity? It should.
4. **The praise check.** When the model did well, is it acknowledged? Michael does this naturally; ghostwritten drafts often skip it.
5. **The why check.** Course corrections in his real voice include the reason. If the draft has him saying "no" without a "because," add the because.
6. **The polish gradient.** Chat messages are unselfconscious (*Naw, rapeditive, breavity, kansas bords*). A script is allowed to be tighter — but the *attitude* underneath should still be plain, not performer.
7. **Self-audit:** would Michael, reading this paragraph, say "that's not how I actually work"? If yes, rewrite.

## What lives elsewhere

- AI presentation voice cut list (em-dashes, "let that land," "It's not X — it's Y"): [`study/yt/voice-analysis-ai-vs-michael.md`](../../../study/yt/voice-analysis-ai-vs-michael.md)
- Therefore/But causation: [`.mind/principles.md`](../../../.mind/principles.md) § Writing & Storytelling Craft
- Source verification (read before quoting): [`../source-verification/SKILL.md`](../source-verification/SKILL.md)
- Auto-memory pointer (loaded every session): `feedback_michael_voice_kindness.md`
