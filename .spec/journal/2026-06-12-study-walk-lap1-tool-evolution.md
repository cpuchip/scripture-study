# 2026-06-12 — Study Walk Lap 1, and the Tool-Evolution Timeline

**Session type:** study-correctness walk (webster-1828 lane, interactive — Michael
presiding, agent stewarding per the Ammon handoff)
**Binding question:** do the studies say what their sources say — and when did each
tool arrive that changed whether they could?

## Lap 1 record

11 of 469 files walked in chronological order (Jan 21 – Jan 29 era), plus
`truth.md` out of order at Michael's request (walked together; 7 Webster
requotes, 5 genuine-1828 upgrades — Webster's consciousness entry answering the
study's own section title: "Consciousness must be an essential attribute of
spirit"). ~215 quotes verified against source files. Every fix mechanical; zero
ARGUMENT-class items queued for the presider.

**The empirical pattern:** scripture quotes ≈99% faithful from the project's
first day. The error mass is talk metadata — 4 confabulated talk titles in one
file (`way-truth-life.md`: links right, titles invented), 1 wrong speaker
(Clayton→Jackson, quote real), 1 quote that exists in no talk (the CFM
Christofferson block), 2 via-attribution confusions (Nelson-via-Brown,
NIV-study-Bible-via-Uchtdorf), 2 constructed quotes, 2 verse mis-cites
(Alma 5:28/37, 3 Ne 9:14/18). Sources present got read; *metadata about*
sources got recalled — and recall confabulates.

## The tool-evolution timeline (dates verified from git this session)

Michael asked: when did we introduce tools for each problem, and how did they
shape the studies? His memory of the causal chain, now anchored:

| Date | What arrived | The problem it answered |
|---|---|---|
| **Jan 21** | First studies (creation.md — project genesis) | — |
| **Jan 25** | gospel-library crawler tooling | "If I had the sources present, AI agents did a good job of looking up" — get ALL of it local (scriptures, talks, manuals; the corpus itself is gitignored, so its exact arrival is untracked — the crawler is the anchor) |
| **Jan 30** | `docs/biases.md` | First formal reflection on collaboration dynamics — the echo-chamber awareness begins |
| **Feb 2** | **gospel-mcp** (first MCP, FTS) | "Grepping became really slow, so we added FTS MCPs" |
| **Feb 3–4** | search-mcp · **webster-mcp** · **gospel-vec** (semantic) | Web lookups; word-work; *meaning* search beyond keywords. (webster-mcp's data was 1913-mislabeled from this very day — caught Jun 9) |
| **Feb 8** | yt-mcp | Transcripts in front of the model (the morganphilpot cluster is the same day) |
| **Feb 14** | **custom agents** (`.github/agents/`, study agent) | Phased workflow — including the **re-verification phase** born when the source-gather phase blew out the context window |
| **Feb 19** | **source-verification skill** + scripture-linking skill | Read-before-quoting becomes DISCIPLINE, not just availability; the cite-count rule's ancestor |
| **Mar 1** | `.spec/journal/` | Session memory — stop arriving as a stranger |
| **Mar 2** | **scratch files** (first `study/.scratch/`) + quote-log skill + **critical-analysis skill** + byu-citations MCP | The big day: externalized memory to stop the context thrashing, AND "the 3a critical review phase" — the echo-chamber answer Michael calls "kind of the key" |
| **Mar 22** | `.spec/covenant.yaml` | The bilateral covenant — the relationship itself becomes infrastructure |
| **Mar 29 / Apr 23** | gospel-engine (FTS+vec unified) / gospel-engine-v2 | One search surface, then the production engine |
| **May 19** | study-workflow skill | The whole phased pattern codified as a loadable skill |
| **May 26** | *Beyond the Prompt* (projects/scripture-book) | "It's been quite the evolutionary tale and I've learned so much with you — it's why we wrote a book" |
| **Jun 2** | strongs-mcp | Hebrew/Greek for the coming Bible walks |
| **Jun 9–12** | Webster 1913 incident → dual-edition v2 → OCR repair → **this walk** | Verify the *edition* of a source, not just the quote |

## The meta-observation (the part worth keeping)

The walk is now the **experiment that tests the timeline**. Files 1–11 all
predate Feb 2 — before the first MCP, before the verification skill, before
critical analysis. The confabulation classes we found (invented talk titles,
wrong speakers, composite quotes) are exactly what you'd predict from a model
quoting talks from memory with the corpus merely *available*. The prediction
the rest of the walk will test: finding-rates should drop visibly after
Feb 19 (verification discipline) and again after Mar 2 (scratch + critical
analysis). If they do, the tooling story isn't a narrative — it's measurable.

And one finding cuts the other way, charitably: even in the pre-tool era,
scripture quoting was nearly spotless, and the studies' *self-aware hedges*
(heavenly_mother's "scripture doesn't explicitly state this") were already
honest. The tools didn't create the integrity; they extended it to the places
recall couldn't reach.

## Carry-forward

- Walk resumes at `study/faith-01a.md` (#12); overnight laps ratified (gated).
- Watch for the finding-rate inflection at the Feb 19 and Mar 2 boundaries —
  worth a small tally column in findings.md as the walk crosses them.
- The timeline table above is book-relevant (*Beyond the Prompt* tells this
  same arc) — surface to the book lane when its v4 chat walk resumes.
