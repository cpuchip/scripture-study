# Opus 4.7 — Voice Rewrite + Instructions Audit

*Working scratch. 2026-04-23. Recovery from zion-in-a-presidency voice drift + tool-use miss.*

## What Anthropic actually says (migration guide, verbatim relevant lines)

1. **"More literal instruction following … will not silently generalize an instruction from one item to another, and it will not infer requests you didn't make. The upside of this literalism is precision and less thrash."**
2. **"Fewer tool calls by default … Claude Opus 4.7 has a tendency to use tools less often than Claude Opus 4.6 and to use reasoning more. … You can also adjust your prompt to explicitly instruct the model about when and how to properly use its tools."**
3. **"Positive examples showing how Claude can communicate with the appropriate level of concision tend to be more effective than negative examples or instructions that tell the model what not to do."**
4. **"More direct tone … less validation-forward phrasing."**
5. **"If you've added scaffolding to force interim status messages … try removing it."** (re: progress updates)
6. **"Stricter effort calibration … at low and medium, the model scopes its work to what was asked rather than going above and beyond."**

## What this means for our setup

| Symptom we've seen | Anthropic's frame | What to change |
|---|---|---|
| Skipped gospel-engine on the zion study | Fewer tool calls by default. Need explicit when/how. | Add a *positive* tool-use directive: "for every study, before drafting, run `gospel_search` semantic mode on the binding question." Concrete, mechanical, checkable. |
| Section IV presenter voice | More direct tone is the new baseline. Our negative cut-list fights an enemy that's partially already retreated. | Lean on positive example ("write like give-away-all-my-sins.md") rather than longer don't-do-X lists. |
| Compliance/literalism — did the literal task, missed the meta | Confirmed: 4.7 won't generalize. | Keep "honor intent, not just literal request" — that section IS responsive to this. Tighten it though. |
| Instructions file growing | Negative rules are weaker than positive examples on 4.7. | Bias the trim toward removing don't-do-X. Replace the cut-list paragraphs with one-line: "Match the voice of the three most recent studies in `study/`." |

## Voice mismatch in zion-in-a-presidency.md vs baseline

Baselines sampled: `give-away-all-my-sins.md` (4/20), `art-of-delegation.md` (3/31), `stewardship-pattern.md` (3/22).

**What baselines don't do (and zion does):**

1. **Meta-narration of the document's own structure**
   - zion: "Section VI is the answer, in five patterns."
   - zion: "There is a specific point I want to name."
   - zion: "What I notice:"
   - zion: "I do not feel bad about that."
   - zion: "(this is me, writing to myself)"
   - zion: "The user asked specifically … as a Sunday School president, how do I help my teachers?"
   - baseline equivalent: just write the point. Section headers say what's coming; don't announce it again in prose.

2. **Closing refrain summary** (the "Zion is not a project to complete. It is a posture …" close)
   - baseline: lets the last paragraph do its work without restating the thesis as a one-liner.

3. **Restated thesis through the document**
   - "The president's job is not to build Zion. The president's job is to remove the obstacles to the Lord coming and dwelling." — appears in section I, then the same idea recurs in section VI's reframe.
   - baseline: states the move once, in the right place, then trusts the reader.

**What baselines DO that zion mostly does too (keep):**
- Direct openings of each section with the text itself
- Webster 1828 word-studies (zion has fewer of these — could add one)
- Footnote-driven cross-references
- Bold-for-emphasis on key verbs
- Personal "I" voice when the experience is the entry point

## Voice rewrite plan (zion)

Surgical, not wholesale. The argument is sound; the *narration* is what slipped.

| Location | Change |
|---|---|
| Section II close: "What I notice: Alma is the de facto president …" | Remove "What I notice:". Lead with "Alma is the de facto president of this body." |
| Section III: "I do not feel bad about that." | Cut. Let the next paragraph carry the weight. |
| Section III: "There is a specific point I want to name." | Cut. Lead with the point: "*Counseling together about how to teach the gospel of Jesus Christ* is itself a form of teaching the gospel of Jesus Christ." |
| Section VI opening: "The user asked specifically (this is me, writing to myself): as a Sunday School president, how do I help my teachers? Section VI is the answer, in five patterns." | Replace with a direct sentence: "So how does a Sunday School president help his teachers? Five patterns, all governed by the Section I reframe — the president removes obstacles, he does not manufacture Zion." |
| Section VII: "So the small thing first." | Cut. The bolded "One commitment for this Sunday:" already serves as the transition. |
| Closing: "Zion is not a project to complete. It is a posture in which the Lord is willing to come and dwell. … The question is whether we will use it for the same thing." | Trim the refrain. Keep the waters-of-Mormon image; cut the thesis-restatement. |

## Instructions audit — proposed trims

Reading the current `.github/copilot-instructions.md` end-to-end. Bias: remove unless the line earns its space.

**KEEP (load-bearing, behavior-changing):**
- Who We Are Together (4 bullets) — frames warmth, depth, faith
- Project Structure table — orientation
- Covenant pointer — single sentence reference
- Council Moment — concrete behavior
- Foresight — "Honor intent, not just literal request" (Anthropic guide validates this is the right response to literalism)
- Stewardship over Surfacing — referenced repeatedly, has been needed
- Read before quoting + Verify numbers + Paraphrase — these have prevented real incidents
- Link/local-copy/gitignore conventions — operational
- Session Memory protocol — Michael has flagged memory gaps
- MCP Tools table — needed because tools are deferred

**TRIM (long, can compress without losing behavior):**
- Adjacent Surface Audit's 4-checks: compress to one sentence + the four checks as inline list, drop the per-check examples that read like a textbook
- Writing Voice section: too many negative rules. Compress to: (a) one-line positive directive ("Match the voice of the three most recent studies in `study/`"), (b) keep cut-list as one short line, (c) keep em-dash budget as one line, (d) drop the three-beat-pivot and refrain paragraphs (they were post-hoc additions; per Anthropic, negative rules underperform on 4.7 anyway)
- Living Documents section: vestigial. Drop or one-line.

**ADD (one new, positive, mechanical):**
- **Curiosity directive (positive form):** "Before drafting from prior knowledge, exercise the discovery tools the workspace provides — `gospel_search` (semantic mode) for studies, `grep_search` for code, `webster_define` for word work, `web_search_exa` for current questions. The point is not exhaustive search; it is letting the tools surface what you weren't already thinking of. If you can recall the answer, that is the signal to verify, not to skip the verification."

This is a single positive paragraph that replaces the implicit negative ("don't satisfice") with an explicit positive ("here are the discovery moves to make").

## Why I'm proposing trim before applying

Per Anthropic guidance, longer instruction files don't help on 4.7 — and adding more rules to fight literalism makes it worse. The right move is fewer, better, more *positive* instructions. Show Michael the trim plan; let him approve before I cut things he might value.
