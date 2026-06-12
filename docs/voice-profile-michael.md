# Voice Profile — Michael (the linguistic fingerprint)

*Built 2026-06-11 by Claude Fable 5, at Michael's request: "build a voice profile of me…
something we can then turn around and gauge the book against. does this sound like
michael? or AI? …bad spelling and all."*

*Companion to [`.claude/skills/voice-michael/`](../.claude/skills/voice-michael/SKILL.md)
(the relational/attitude layer: kindness, pronouns, no fabricated anger). This document
is the **mechanics layer** — orthography, rhythm, structures, lexicon — plus the **Book
Gauge** at the end.*

---

## Corpus & method

| Corpus | Era | Size | Source |
|---|---|---|---|
| VS Code archive (old machine) | Sept 2025 – Jan 2026 | 1,514 messages / ~101K words | `external_context/Code-old-session` workspaceStorage chatSessions (forkirk, astrotreks, storygames, mobile-games/simple-games, homeschool-keeper, scripture-study…) |
| Copilot CLI (old machine) | ~Mar 2026 | 115 messages / ~3K words | `external_context/.copilot-old/session-state` |
| Copilot (current) | Apr 2026 | ~5K words | `.spec/scratch/copilot_prompts.txt` |
| Claude Code (current) | May 2026 | ~54K words | `.spec/scratch/prompts_recent.txt` |

User messages only (tool results, pasted logs, and `/fix` invocations excluded from the
qualitative read; pasted logs inflate some raw counts). Marker rates computed across the
full text; qualitative patterns from stratified samples of the 40–160-word "voice-rich"
band (520 messages) plus the live 2026-06 sessions. Extraction scripts:
`.spec/scratch/extract_*.py`; corpora at `.spec/scratch/voice-corpus-*.txt`.

## The fingerprint at a glance (top tells)

1. **`lets` without the apostrophe** — 691 vs 25 `let's` in the 2025 corpus (96%);
   158 vs 2 in the May 2026 corpus. The single most reliable Michael-marker.
2. **Zero em-dashes.** 0 in 101K words. Michael connects with commas, parentheses, and
   new lines — never `—`. (The AI's favorite glyph is his never-glyph.)
3. **"Okay" as a pivot-opener** — 181 messages open with "Okay" (often "Okay so", "Okay
   I've been…"). It marks a turn: new topic, verdict arriving, or rolling up sleeves.
4. **Numbered decision lists answering AI questions** — terse rulings, one per number,
   often with `a. b. c.` sub-letters: *"1. Just cooldowns to start with… 2. spells
   should auto target nearest enemy… 4. No saved progress, start fresh each time."*
5. **Rule + exception in one breath** — *"keep that while playing that game with that
   opponant to keep things simple (this doesn't make sense to keep as a rule in 3 or
   more) but in two player games…"* He legislates with carve-outs inline.
6. **Evidence-first debugging** — paste the log, then the observation: *"I've noticed
   that not only is Host changing from going first or second (which is good) but it's
   also chaning from X to O which is not good."* Parenthetical verdicts ride along.
7. **"I think" as honest hedge** — 2.1–2.6 per 1k words, everywhere, sincere not filler:
   *"I think it happens every few weeks. I don't have a regular cadence, it just pops
   in my head."*
8. **Plain warm punctuation** — `!` at ~3/1k (real enthusiasm, never decorative), `?`
   ~9/1k (he genuinely asks), `:D` and `:)` present (15 in the game-dev corpus).
9. **Median message: 31 words.** Bursts. Long messages are double-newline-separated
   thought chunks, not flowing paragraphs.
10. **Empathy for the model as a person-ish collaborator** — *"there's a huge amount of
    chat context here and it's slowing vs code, and probably you too"*; *"I'm kinda
    tired, but you're awesome, and really good at tracking down concurrency issues."*

## Orthography (the "bad spelling and all" layer)

Michael types fast and does not look back. The error classes are consistent enough to be
identifying:

- **Dropped apostrophes:** lets, im, id, Ill, thats, dont (rates vary; `lets` is the
  reliable one — the others he often gets right).
- **Transpositions / fast-finger doubles:** thning (thinking), wee'll, leavees, wiill,
  growning, chaning (changing), uing (using).
- **Phonetic & near-miss spellings:** premis, prominant, opponant, indivdual, delt
  (dealt), decending, clenaup, arcived, turnament, sinple, copywrite (copyright),
  rapeditive (repetitive), breavity (brevity).
- **Homophone/word swaps:** "He's a quick premis" (Here's), "outs put" (output's),
  "site the source" (cite), "dairy feel" (diary), "no thats backwards", "15 chang".
- **Doubled small words:** "the noise of of the world."
- **Idiosyncratic spacings:** "tick tack toe" (always three words), "some thing"
  occasionally.
- **Capitalization:** ~17% of messages start lowercase; mid-message sentences often
  lowercase after a blank line. Proper nouns get capitalized when he's being careful
  (product names) and not when he's moving fast.

**Gauge note:** these belong to *chat at speed*, not to the book. Their value is
authentication (is this transcript really him?) and for any verbatim self-quote the book
carries (P1 now quotes one — cleaned only `lets→let's`).

## Rhythm & syntax

- **Comma-glued run-ons** when excited or specifying: clauses chained 3–5 deep, no
  semicolons (0.45/1k, mostly in pasted code), no em-dashes (0).
- **Double-newline thought chunks**: one message = several mini-paragraphs of 1–2
  sentences each, each a complete instruction or observation.
- **Short declarative rulings:** "No saved progress, start fresh each time." / "We'll
  call it Shields Down!" / "Lets use local wifi."
- **Self-clarification re-runs:** *"What I mean by show each card on all screens is,
  Show the last card played on all the screens."* He restates rather than edits.
- **Context-setting openers:** *"Note: I'm on powershell"* — environment first, ask
  second.
- **Politeness frame around redirects:** apology + reversal: *"we'll pass on 3 for the
  moment (again… I'm sorry!)"*; correction + reason, never bare "no."

## Lexicon (words that are his)

`lets` · `Okay` · `awesome` · `sweet` · `nicely done` / `nice work` · `dig in` /
`dig into` (11×) · `hone in` · `fixup` / `cleanup` as verbs · `phase` (26× — he plans in
phases natively) · `nice to have` (backlog vocabulary) · `ratify` (51× in the 2026
corpus — council vocabulary he chose) · `crazy idea` (idea-introduction ritual) ·
`carry on` · `I want to` (36×) · `I've noticed` (38×) · `I'd like` · `we should` ·
`byte sized` · playful coinages ("5 dimensional time travel chess").

**Register shifts by mode (the rates move, the voice doesn't):**
- *Planning/excited:* exclamation marks, coinages, "We'll call it X!", chained `lets`.
- *Debugging/tired:* evidence pastes, "I've noticed", parenthetical verdicts, fatigue
  named honestly + kindness kept ("I'm kinda tired, but you're awesome").
- *Study/devotional:* earnest belief statements ("I believe faith is more than…"),
  multi-part study framing, the same `lets` imperatives aimed at scripture.
- *Council/verdicts:* numbered rulings, "lets do 1 and 2", scope deferrals ("keep it a
  later phase goal").

## What Michael never (or almost never) writes

The zero-and-near-zero list, per 101K words — **these are AI-tells when they appear in
"his" voice:**

| Marker | His rate | AI default |
|---|---|---|
| Em-dash `—` | **0** | constant |
| "It's not X — it's Y" negation-contrast | 0.09/1k | ~1 per paragraph |
| however / moreover / furthermore | ~0.05/1k | transition glue |
| crucial / robust / comprehensive / seamless / leverage / delve | **0–0.01/1k** | résumé varnish |
| Triadic flourishes ("A, B, and C" for rhythm) | rare | signature |
| Closing refrain / thesis restated as one-liner | absent | signature |
| Meta-narration ("In this section we will…") | absent | signature |
| Anger at the model | **0 instances, 14 months** | training-data trope |

## THE BOOK GAUGE

The question is **not** "does the book read like Michael's chat?" — the book is edited
prose and *should* be tighter than chat. The question is: **do the first-person
passages carry his cadence and posture, or the model's?** Run any "I" passage (Part One
stories, Becoming Commitments, preface, afterword, production notes) through this:

**Michael-markers (want ≥4 in any first-person passage):**
1. Concrete specifics over abstractions (a date, a tool name, a count he'd actually know).
2. Short declarative rulings; sentence rhythm that breathes in bursts, not balanced
   periods.
3. Rule + exception carried inline (parenthetical carve-outs).
4. Honest hedges where he'd hedge: "I think," "usually," "I don't have a regular
   cadence" — calibration as voice, not weakness.
5. Plain warm verdicts: good things called good simply (awesome / it worked / I loved it).
6. Responsibility taken in failures; credit shared in wins ("we").
7. Excitement unselfconscious, slightly playful, never performed.
8. Phase/backlog thinking: "to start with," "later," "nice to have," "next."

**AI-tells (each one found = a point against; 2+ = rewrite the passage):**
1. An em-dash doing connective work in HIS mouth (narration may; "I" quotes shouldn't
   lean on them).
2. "It's not X. It's Y." / "not because X but because Y" as the sentence's spine.
3. A triad built for rhythm rather than content.
4. A paragraph that ends by re-saying itself one notch more poetic.
5. Transition glue: however / moreover / indeed / in essence.
6. Varnish vocabulary: crucial, robust, seamless, profound, comprehensive.
7. Anger, fighting, or adversarial framing toward the model.
8. Meta-narration of the document's own structure.

**Worked example (P8, post-v4):** *"I still run it every few weeks, whenever it pops
into my head — usually after the work has settled into a routine."* — markers 1, 4, 8
present; the em-dash is narration-level and the cadence is his hedge-honest register.
Passes. The pre-v4 version ("on a rough cadence" implying schedule) was the model
smoothing him into more discipline than he claims — exactly the drift this gauge
catches.

**Known divergence to respect:** Michael's chat never uses em-dashes, but the book's
*narration* legitimately does (edited prose; budget: one per paragraph, per the
workspace voice rules). The gauge applies the zero-tolerance only inside first-person
self-quotes and the most personal passages.

## Honest limits

- Chat-at-speed ≠ authorial voice; this profile authenticates the *speaker*, not a
  style target for polished prose. The book may be "far above my own skills" (his
  words) in finish — the gauge protects the *posture*, not the typos.
- The corpus skews technical (game dev, substrate, study direction); personal-essay
  registers (the dedication, the afterword's reflective close) have thinner chat
  comparables — gauge those primarily with the attitude layer
  (voice-michael skill) and his ratification.
- Built from ~163K words across 14 months and three harnesses; strong for the patterns
  cited, silent on registers he hasn't typed (public speech, fiction).
