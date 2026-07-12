---
name: study-workflow
description: "The phased scripture-study workflow — binding question first, scratch file always, discovery search before drafting, gap and critical analysis as a hard gate, voice audit before shipping. Load when beginning any scripture study, or when a session realizes mid-stream that it is doing scripture study."
user-invokable: true
---

# Study Workflow

The phased process that keeps a scripture study honest. The study agent
(Abinadi) carries this in full; this skill is the portable version — load it in
any session that turns out to be doing scripture study, so the discipline
travels even when the agent was not invoked.

**Why this exists:** a study written from recall, without the discovery search
and without a scratch file, will look finished and still miss the verse that
would have reframed it. This skill is the checklist that prevents that. It was
written after exactly that failure — see study/.scratch/what-abides.md.

## The principle

**Files are durable, context is not.** Externalize verified quotes and
observations to a scratch file *as you read them* — never hold them in memory
to write all at once.

## The hard gate

Before writing a single line of study draft, the scratch file MUST contain:

1. The binding question.
2. Verified quotes from sources actually read this session.
3. A `## Gap & Critical Analysis` section.

If that section is not in the file, the work did not happen — regardless of
what was said in chat. "This is a tight, focused study, I'll skip the analysis"
is the most reliable way to ship a study that confirms a hypothesis instead of
discovering one.

## The phases

1. **Outline.** State the binding question — the specific question the study
   answers, not the topic. Write it to both `study/{topic}.md` and
   `study/.scratch/{topic}.md`.
2. **Source gathering.** Begin with a discovery search — at least one
   `gospel_search` (semantic or hybrid) on the binding question — *before*
   drafting from recall. Then Read each source; write verified quotes to the
   scratch file after every source. Follow footnotes. Webster 1828 for
   load-bearing words.
3. **Gap analysis.** Read the scratch file against the outline. What is
   under-sourced? What voice is missing?
4. **Critical analysis.** Stress-test before drafting: strongest claims against
   the text, weakest links, missing voices (all five standard works? modern
   prophets?), speculation vs. doctrine, the ring check (does it answer the
   binding question?), the posture check (discovering or confirming?). Write
   these notes into the scratch file's `## Gap & Critical Analysis` section.
5. **Draft.** Read the scratch file — it is the primary source now. Write the
   study. Quotes are already verified; spend context on thinking, not
   re-verifying.
6. **Review and voice audit.** Verify links (`scripture-linking`). Voice audit:
   em-dash budget (one per paragraph), therefore/but not and-then, the cut
   list, no meta-narration, no closing refrain. Stats audit: every number,
   date, and count must trace to a tool call from this session.
7. **Becoming and clean up.** The study lands somewhere personal. Keep the
   scratch file — it is permanent research provenance. Update memory.

## Read before quoting

Every quotation-marks quote comes from a source `Read` this session. Numbers,
dates, and counts are quotes too — verify them the same way. If you have not
read it, paraphrase; do not put quotation marks around it.

## Related

- `quote-log` — the scratch file format.
- `source-verification` — the verification standard.
- `scripture-linking` — link conventions.
- `critical-analysis`, `deep-reading`, `webster-analysis` — phase tools.
- Full agent: `.claude/agents/study.md` (Abinadi).
