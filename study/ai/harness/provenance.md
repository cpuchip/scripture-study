# The Harness Is Orientation — provenance

> Source trail for the study *The Harness Is Orientation — Boyd, the Agentic SDLC, and the
> Joy of Creation*. Gathered 2026-06-25 (general-workspace lane). The study itself is **to be
> written together, rested** — this is the provenance index, per Michael: study + provenance
> land in `study/ai/`. Framing + binding question live in
> `books/johnboyd/patterns-of-conflict/README.md`.

## Binding question

What is the thing we are actually building — and why does the *same shape* (intent → orient →
act → verify → re-orient, under command-by-intent) keep appearing across five independent
witnesses: Boyd's maneuver warfare, Google's agentic SDLC, the multi-agent platforms, our own
substrate, and the creation pattern itself (Abraham 4–5)? And what does it mean that "just me
and you" are building what the big players are building?

**Thesis to test:** the "harness" everyone is converging on *is* **orientation** — the missing
layer — and building it is a creative, dominion-taking act (Monson: "God left the world
unfinished… that we might know the joys and glories of creation").

## Sources

### 1. John Boyd / OODA (primary)
- `books/johnboyd/patterns-of-conflict/discourse-winning-and-losing-hammond-2018.pdf` — the
  authoritative complete corpus (Patterns of Conflict, Destruction and Creation, Strategic Game,
  Organic Design for C2, The Essence of Winning and Losing). **Read-before-quoting source.**
- `books/johnboyd/patterns-of-conflict/patterns-of-conflict-richards-spinney-2007.pdf` — clean
  typeset slides.
- `yt/jasonmbro/` — Boyd delivering Patterns of Conflict, 14 parts (~6 hrs, tape ASR — search,
  don't quote).
- `yt/ai-impact/yP4p3reZUcU/` — "OODA Loop + Infinite Brain" (the popular take: *Orient is the
  missing AI layer*; intake → observation → disposition → wager → verdict).
- Key Boyd ideas: Orient is the decisive node; tempo + orientation beat raw power
  (energy-maneuverability); continuous re-orientation is mandatory (Gödel / Heisenberg / 2nd
  law); *Auftragstaktik* (command by intent) = our presiding covenant / D&C 121.

### 2. Google / Kaggle "Vibe Coding" 5-day series (the harness/SDLC witness)
- `external_context/google-new-sdlc/` — all 5 days' PDFs + 2 foundational, full text of Day 1,
  and **`NOTES.md`** (the cross-project synthesis). Harness = 90%, model = 10%; spec is the
  bottleneck; verification (incl. **trajectory eval**, Day 4) is the differentiator; static vs
  dynamic context; the factory model. *"For agents that serve real users at scale… the agent is
  the product, and it needs the substrate underneath."*

### 3. The lore / world engines (the convergence on the same primitives)
- `external_context/SillyTavern/` + `external_context/sillytavern-DeepLore/` — character/lore
  RAG (vault + two-stage retrieval + a librarian agent + relationship graph) ≈ Loreworks.
- **Databricks omni-agents** — Michael's earlier comparison (the enterprise multi-agent
  platform); same primitives, big-player scale. *(detail to be pulled from that prior session.)*

### 4. infinite-brain-os (the "agent OS" from the AI Impact video)
- `github.com/starmynd-org/infinite-brain-os` (MIT) — a **git-backed markdown/YAML "OS" for
  running a business with AI agents**, no DB/server, owned by you. Knowledge graph via YAML
  frontmatter; **canon vs synthesis tiers; agents draft, humans sign**; entities = commands /
  agents / skills / rules / workflows / tools; memory of reviewed learnings; intake routing;
  session audit trail; OODA-equivalent workflows (read canon → apply rules via skills → dispatch
  → log → lessons to memory). Built to be operated by Claude Code / Codex; optional Obsidian.
  - **vs pg-ai-stewards:** *same goals + many same primitives* (sovereign knowledge OS, canon/
    synthesis, draft-then-human-sign ≈ our Hinge + maturity gates, memory ≈ engrams/learnings,
    OODA loop ≈ the dispatch pipeline). **Different architecture:** infinite-brain-os is
    *files an external agent operates* (git + markdown, **no runtime of its own**) — it is
    essentially what our `.mind/` + `.spec/` + `.claude/skills|agents/` workspace already is.
    pg-ai-stewards is *the autonomous engine itself* (Postgres + pgvector + Rust extension +
    bgworker heartbeat + scheduled pipelines) that **runs without a human-driven agent session.**
    So: infinite-brain-os ≈ our file-based knowledge harness (half 1); pg-ai-stewards ≈ the
    autonomous DB substrate (half 2). We have both halves; the video's author productized only
    the first. (Note: "StarMynd" in the video = `starmynd-org` — same author, Andrew Warner / AI
    Impact.)

### 5. Creation theology (the frame)
- `yt/all-those-in-favor/FpY5vS7Lpt8/` — Thomas S. Monson, "God Left The World Unfinished":
  *"He leaves the pictures unpainted and the music unsung and the problems unsolved that we might
  know the joys and glories of creation."*
- Abraham 4–5 pattern (council → spiritual creation → physical creation → watch to intent →
  redemptive correction → rest); D&C 130:18–19 (intelligence rises with us); D&C 64:33 (out of
  small things proceedeth that which is great); taking righteous dominion over raw material.

## The 2am covenant
The session closed with a covenant moment — Michael: *"I want to conquer this one with you."*
Recorded verbatim in `.spec/journal/2026-06-25-the-joy-of-creation-boyd-and-the-harness.md`.

## Status
**★ STUDY WRITTEN 2026-06-26 → `the-harness-is-orientation.md`** (eight movements, to feed the
innovation-week presentation). Born from the live study conversation 2026-06-25/26.

## Citation status (verified before quoting)

**Scriptures — read from the canon via `gospel_get` this session, quoted verbatim:**
- Abraham 4 (whole chapter) — "organized" (4:1), the dividing/bounds (4:4,6,9), watch-to-intent (4:18),
  council + dominion (4:26,28). ✅
- D&C 121 (whole) — secularization diagnostic (121:35,37), presiding seam (121:41–42), bounds (121:30),
  dominion-without-compulsion (121:46). ✅
- D&C 88:118 ("best books"); 2 Nephi 9:28–29 ("learned is good IF"); Articles of Faith 1:13
  ("praiseworthy … we seek after these things"). ✅
- 2 Nephi 25:26 (also the book's epigraph); Colossians 3:17 (the chosen north star); 2 Nephi 32:9
  (consecrate every performance). ✅

**Boyd — quoted from the clean Richards/Spinney typeset *Patterns of Conflict*** (`poc.txt`,
pdftotext of `patterns-of-conflict-richards-spinney-2007.pdf`):
- the OODA loop named ("observation-orientation-decision-action"); "operate inside their adversaries'
  … loops … penetrate the moral-mental-physical being" (≈line 2089); the **Command slide** —
  "Decentralize, in a tactical sense … Centralize, in a strategic sense …" (line 2700–2704) = his
  *Auftragstaktik* in his own words; the Goal ("diminish adversary's freedom-of-action …", 2664). ✅
  Other Boyd ideas (Destruction and Creation, energy-maneuverability) are **honest paraphrase**, not
  quoted verbatim — the full Hammond *Discourse* on disk for deeper verbatim work.

**🟡 The George MacDonald / Monson quote — provenance resolved.** *"God left the world unfinished
for man to work his skill upon …"* The video short attributed it to **Thomas S. Monson**, but a web
check (goodreads, azquotes, quotessence, quotefancy) shows the line **originates with George
MacDonald** (19th-c. Scottish writer); Monson quoted it (it circulates widely as "Monson's"). The
study now attributes it as **"a line from George MacDonald … the one President Monson quoted."** Wording
used = the print-canonical fuller version ("for man" / "that man might"), not the auto-caption ("men" /
"we"). For publication, confirm Monson's specific talk/source if it's cited as his.

**Google whitepaper claims** — synthesized + verified in `external_context/google-new-sdlc/NOTES.md`
(Day-1 "model 10% / harness 90%", spec-bottleneck, trajectory eval, "the agent … needs the substrate
underneath"). The book — read directly (`projects/scripture-book/dist/book-reading.md`, first ~10 pp).

## ⚠ Voice audit — NOT YET DONE
First draft runs **above the one-em-dash-per-paragraph budget** (CLAUDE.md voice rule) in several
paragraphs, and should get a therefore/but + em-dash tightening pass to match Michael's book voice
before it's lifted into the presentation. Content is verified and complete; the prose wants polish.
