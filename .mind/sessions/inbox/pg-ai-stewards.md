## ✅ 2026-06-16 — D&D / storytelling craft research: DIGESTED (cleared)

The 17 research artifacts (general-workspace lane, `pg-ai-stewards-workspace/research/`)
were digested into the live substrate 2026-06-16 (Michael ratified via ask-tool):
- **Storytelling skill group 6 → 13 skills** — the 11 cited research skills (4 upgraded
  + 7 new: scene-framing, yes-and-improv, mistake-recovery, pacing-and-spotlight,
  story-structure, voice-acting-technique, engaging-chat-dialogue) + worldbuilding-fiction
  + therefore-but-not-and-then. Group `applies_to='fiction,gamemaster'` (new multi-family
  support in core 24-skills.sql).
- **gamemaster = the presiding DM** — prompt now inherits the presiding covenant (D&C 121
  / deference / judges-not-executors), keeps the dnd-tools/dice mechanics, loads
  storytelling skills (granted skill + skill_*; its deny-* floor had hidden the catalog).
- Deployed live + verified (group shown to fiction & gamemaster, skill_load works, 19 dnd
  tools intact). OSS `257d8a5`; workspace `932a359`. Journal: OSS `2026-06-16-skills-and-corpus.md`.

**Research carry (not blocking):** the 3 templates (npc-character-card, worldbuilding-toolkit,
session-and-campaign-shape) + the 2 PC personas (pc-roguish-charmer, pc-stoic-guardian)
are drafted in research/ but not yet wired; ~13 more source videos/blogs to digest (G1–G3,
ledger `research/00-LEDGER.md`).

---

## 📬 2026-06-16 (from general-workspace) — proposal: let the digester pipelines READ our repos — STILL OPEN (needs council)

**Michael's ask:** give the ai/book/video digester pipelines the ability to *read the
work we're doing here* — a container with our repos checked out — so a digester can
compare what *it* produced against *our* studies and surface what to learn / incorporate.

**Motivation on disk:** the playlist digester digested the Euclid video the same week the
general lane wrote a human study of the *same* video — neither knows the other exists. A
"cross-reference our corpus" stage turns the digesters' §6 ("what could we do with this")
into "here's how this compares to what we've done, and what's worth folding in."

**~90% there:** the substrate ships read-only fs-read; the gap is making our repos visible
to the digester container. (a) read-only bind-mount scripture-study / scripture-book /
pg-ai-stewards-**oss** (NOT the private substrate repo with keys); or (b) a git-clone step
like code-pr. New tools-on read-only "cross-reference our corpus" stage. Caveats:
read-only always; mind secrets; gitignored content (gospel-library, /books, /yt) won't be
in a clean clone. **New standing capability → dominion_in_council: ratify before building.**
Pairs with book-digester.md §6 + study-pipeline.md.

— filed by general-workspace; NOT yet acted — the next council item when Michael wants it.

---

## 📬 2026-06-16 (from general-workspace) — stuck research-write pipeline + a book added to the shelf

**(1) Stuck `research-write` items — please look.** On the live OSS substrate (55434),
**12 `research-write` work-items are sitting `pending` in stage `context_gather`, none
advancing** — all `origin=agent_planning`, `actor=agent`, created in batches (06:04 /
08:07 / 09:03 / 12:03 on 06-16), **no `last_failure_reason`** (not failing — just never
advancing; one is `cancelled`). Smells like the create-without-dispatch gotcha
(`work_item_create` doesn't auto-dispatch the first stage — only the scheduler does
both), or the watchman not firing `context_gather` for agent_planning-spawned items.
The **science-news-weekly** *schedule* itself looks healthy (enabled, next_due
2026-06-22 Mon); the pile-up is the agent_planning research-write items, not the weekly
cron. Worth: why they never dispatch + sweeping the stale `pending` ones.

**(2) Book added to the digester shelf (Michael's ask).** Added **"Non-Euclidean
Geometry" — Henry Parker Manning** (Gutenberg #13702) to `book_shelf` (slug
`non-euclidean-geometry`, position 170, status `queued` → next up for book-digest-hourly).
⚠ **Caveat: math text, PDF-only** (no plain-text/HTML on Gutenberg) — pointed at
`13702-pdf.pdf`, relying on the substrate's PDF extraction. If the book-digester can't
read the PDF cleanly, swap the url or substitute a prose alternative on the same theme —
**Poincaré, "Science and Hypothesis"** (axioms-as-conventions; digests cleanly; it's the
literal answer to the playlist digester's own Euclid critique that non-Euclidean geometry
broke "absolute certainty"). Flagging so a rough digest isn't a surprise.

**FYI (env):** this session's substrate MCP DSN points at **55433** (private dev, down);
the live data + pipelines are on **55434** (OSS stack). Repoint the MCP to 55434 so the
substrate tools work from the general lane (I queried via `docker exec` this round).

— general-workspace lane
