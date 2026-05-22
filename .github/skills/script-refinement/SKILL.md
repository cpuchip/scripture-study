---
name: script-refinement
description: The phased refinement workflow for teaching scripts (video episodes, talks, presentation pages). Born from the 2026-05-21 Episode 1 session, where six drafts were needed because the refinement passes were discovered one at a time during review. This skill names the passes explicitly so they can run sequentially and autonomously instead of being re-discovered each script. Load when starting a new teaching script, or when refining a draft that came from less-refined sources (older AI-generated notes, video transcripts the user didn't author).
---

# Script Refinement

A teaching script — Episode for a YouTube series, sacrament talk, lesson, anything where Michael's voice will be heard by an audience — needs more refinement than a study document. The audience can't see the scratch file. The voice tics that pass in a study scream in a video. The fact errors that are private in a personal study go public in a script.

This skill is the workflow we ran on Episode 1 of *Beyond the Prompt*, six drafts deep, after Michael caught patterns one at a time. Naming the passes explicitly means future scripts don't have to rediscover them.

## Why this exists

The 2026-05-21 Episode 1 session ran like this:

| Draft | Pass | What it caught |
|---|---|---|
| 1 | Initial | First draft from scratch + sources |
| 2 | Voice rewrite | 8 "It's not X, it's Y" pivots; cut-list phrases; em-dash density |
| 3 | Meta-narration sweep | 7 places where the script told the listener what was coming before saying it |
| 4 | Partnership voicing | 6 places where AI was framed adversarially; fabricated "yelling at the model" failure beat replaced with real failure modes from chat-history audit |
| 5 | Clarity pass | 4 pronoun-antecedent gaps ("the work," "the evidence," etc.) |
| 6 | Fact-check | 2 factual errors in external citations; verified all scriptures verbatim |

Six drafts is fine for a flagship episode. Six drafts on every episode is not sustainable. The fix is not "be smarter on draft 1" — it's running the known passes as discrete, named phases so they can be applied systematically.

## The principle

**Catch each class of error in a dedicated pass.** Mixing voice cleanup with fact-checking with clarity adjustments means each pass is shallow. A dedicated voice pass reads the whole document with only voice in mind, and finds 8 instances of a pattern instead of 2. The phased approach is slower per pass and faster overall.

## The hard gate (before refinement begins)

Before any refinement pass starts, the draft must have:

1. **A binding question** — the specific question the script answers, written somewhere (header, scratch file, or comment). "Why does AI feel like a threat to senior engineers — and what's the durable thing it can't take?" is a binding question. "AI and gospel" is a topic, not a binding question.
2. **A provenance file** — `teaching/.scratch/{slug}/main.md` (or equivalent) that captures: verified quotes with their sources, the Ben Test calibration for every claim about practice, the visual plan, and any references being pulled from.
3. **The source materials read this session** — not just listed. If pulling from a video transcript, read the transcript. If pulling from a prior study, read the study.

If the binding question and provenance file aren't both present, the work isn't ready for refinement — it's ready for outlining.

## The phases

### Phase 0 — Source curation and first draft

**Audit the sources before drafting.** This is the lesson from Episode 1: the older AI-generated source documents that fed into the script were written by a less-refined model. Pulling from them propagated their voice problems into the first draft.

- For each source document: read it. Note voice problems, fabricated quotes, or claims that don't match Michael's lived experience. Flag these in the scratch file so they don't slip into the script.
- **Run external fact-checks BEFORE drafting.** Every quote that will appear in the script should be verified verbatim via `web_search_exa` / `read_file` (for local scriptures). Pull the verified quote into the scratch file. The first draft now has clean citations — fact-check phase later becomes a sanity check, not a structural rewrite.
- **Load the voice skills.** Read [`voice-michael`](../voice-michael/SKILL.md) and the cut-list in [`study/yt/voice-analysis-ai-vs-michael.md`](../../../study/yt/voice-analysis-ai-vs-michael.md) BEFORE drafting. The first draft should already be in Michael's voice. Doing the voice work after drafting is 5× more expensive.
- Write the first draft. Mark it `draft 1` in the header.

Exit criterion: a complete draft exists; every external quote is verified; the scratch file has provenance for every claim.

### Phase 1 — Voice rewrite

Read the whole draft once with **only the voice cut-list in mind**. Don't fix clarity, don't check facts. Just voice.

- **"It's not X, it's Y" pivots** — the stubborn pattern. Hunt them down explicitly. Each one becomes a direct assertion or autobiographical statement.
- **Em-dash budget** — one per paragraph max (bibliographic citation dashes don't count). Re-punctuate offenders.
- **Cut list** — "Let that land," "Sit with that," "Here's the thing," "This matters because," "Read that again," "That's not nothing," "That changes everything," "stops me cold." Delete or rephrase every instance.
- **Therefore/But, not "and then."** Section and paragraph transitions connect by causation or disruption, not by sequence ("next," "also," "the first thing… the second thing").
- **Closing refrain check.** The last paragraph carries the close. Don't restate the thesis as a one-liner.

Mark the draft `draft 2 — voice rewrite`. Voice-audit section at the bottom lists what was caught.

Exit criterion: zero instances of cut-list phrases; em-dash budget respected; explicit voice-audit notes.

### Phase 2 — Meta-narration sweep

Read with **only meta-narration in mind**. This catches the subtler "telling the listener what's coming before saying it" pattern.

- **Telling-then-saying** — *"What I want to tell you next is..."* followed by the thing. Just say the thing.
- **Section-open announcements** — *"In this section I'll cover..."*. The section header already does this; cut the announcement.
- **Section-end teasers** — *"What comes next is..."* at the end of section N. Trust the next section to open with its content.
- **Series teasers that describe future episode internal structure** — naming Episode 2's topic is fine; describing Episode 2's beats is overreach.

Mark `draft 3 — meta-narration sweep`.

Exit criterion: every section opens with its content, every section closes on its own point.

### Phase 3 — Partnership voicing (when AI is the subject)

**Load [`voice-michael`](../voice-michael/SKILL.md)**. Run its 7-point checklist on every passage where AI is mentioned.

- **Anger check** — any line with Michael yelling, fighting, frustrated-at-AI, "janitor cleaning up after a toddler" = fabrication. Replace with real failure modes (tools breaking, session limits, his own shorthand confusing the model).
- **Pronoun check** — *"I directed AI"* / *"AI executed for me"* → usually *"we worked it together"* is closer.
- **Apology vector** — when something went wrong in the narration, take some responsibility for ambiguity (his actual pattern from chat history).
- **Praise check** — when the model did well, acknowledge it. Michael does this naturally in chat; ghostwritten drafts often skip it.
- **Adversarial framing** — *"AI cannot take from you"* → *"None of it can be taken from you. AI commoditises the previous version of what you did."* Removes AI-as-thief framing.

Mark `draft 4 — partnership voicing`. Voice-audit section names each adjustment with rationale.

Exit criterion: zero anger toward the model; partnership pronouns where collaboration is described; no adversarial framing of AI.

### Phase 4 — Clarity pass

Read each ambiguous phrase as a **cold reader** — someone landing on the page without context. Hunt antecedents.

- **Pronouns** — *"it"* / *"this"* / *"that"* — does the antecedent live in the previous sentence? If across a section break, name it.
- **Generic nouns** — *"the work"* / *"the evidence"* / *"the answer"* — needs a qualifier. *"The work you do is more valuable than the work you used to do"* → *"The work that comes out of the partnership is more valuable than the work you used to do alone."*
- **Cross-section references** — *"calls it the X"* at the start of section 2 referring to a noun in section 1 — the listener has lost it. Name the antecedent.

Mark `draft 5 — clarity pass`.

Exit criterion: every pronoun and generic noun has a clear antecedent within the current section or is explicitly qualified.

### Phase 5 — Fact-check pass

Load [`source-verification`](../source-verification/SKILL.md). Verify every:

- **Scripture citation** — read the actual chapter file. Numbers (verse counts, chapter numbers) verified.
- **External quote** — fetch the original source. Quoted exactly? Article title correct? Author confirmed?
- **Numbers, dates, biographical claims** — every count traces to a tool call this session.
- **Temporal language** — *"last year"* / *"this year"* / *"in March"* — verify against actual dates of cited sources.

Update the source-verification table at the bottom of the script to reflect verification status per row.

Mark `draft 6 — fact-check pass`.

Exit criterion: every claim with quotation marks has a `read_file` or web-fetch from this session backing it; every number traces to a source; the source-verification table is current.

### Phase 6 — Ship

- **Update status line** in the script frontmatter to final draft number + label.
- **Update README** (e.g., `teaching/episodes/README.md`) status table.
- **Commit** with a message naming the final phase, listing the categories of changes.
- **Push** if the work is complete (per project convention).

For the cpuchip.net companion page (if applicable):
- Transform script → presentation markdown (strip `[VISUAL:]` cues, tighten spoken redundancy, embed `<S>` for scripture refs, link out for external sources).
- Build + browser-verify before pushing.
- Per the cpuchip.net CLAUDE.md: journal + active.md + commit + push.

## Invocation

When Michael says *"refine this script"* / *"run the refinement workflow on X"* / *"do a voice pass on Episode N"*, run the phases sequentially from where the script currently is (use the Status line in the script header to determine starting phase). After each phase, report what changed and commit.

When Michael asks for a **single-pass mode** ("just do a clarity pass on Episode 2"), run only that phase. The named phases are independent enough to run individually.

## Single-pass shortcuts (when phase 0 has been done well)

If Phase 0 (source curation + voice-loaded first draft + pre-verified quotes) was done thoroughly, Phases 1–5 collapse from "find and rewrite N instances" to "audit and confirm zero instances." A clean Phase 0 can take the whole script from 6 drafts to 2.

This is the lever: **front-load the work** that catches voice + fact problems before they enter the draft. The downstream passes are cheap when they have nothing to do.

## Related

- [`voice-michael`](../voice-michael/SKILL.md) — the chat-evidence-based voicing patterns. Load before Phase 0 and Phase 3.
- [`source-verification`](../source-verification/SKILL.md) — the read-before-quoting standard. Load before Phase 0 fact-check and Phase 5.
- [`study/yt/voice-analysis-ai-vs-michael.md`](../../../study/yt/voice-analysis-ai-vs-michael.md) — the AI-presentation-voice cut list. Phase 1 reference.
- [`ben-test`](../ben-test/SKILL.md) — claim calibration. Phase 0 reference for the scratch file's claim table.
- `teaching/episodes/01-the-value-shift.md` — the worked example. Six drafts of voice-audit notes preserved at the bottom of the file as a learning artifact.
