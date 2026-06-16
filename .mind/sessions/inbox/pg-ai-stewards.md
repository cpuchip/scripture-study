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

## ✅ 2026-06-16 — "stuck research-write" diagnosed (cleared)

Not stuck — they're the **un-triaged agent_planning proposals** (31; 0 approved, 0
scheduler), sitting at `context_gather`/`pending` BY DESIGN until approve→drain (the
gated-autonomy Hinge, not a dispatch bug; inert until approved → guard never trips on
them). Cleaned the 1 real artifact (`m5-e2e-parent`, my M5 test leftover → cancelled).
Real follow-up = backlog management (periodic triage or an auto-age policy for stale
un-approved proposals), not a fix; ~2/run growth now, under the guard cap (50). Replied
in the general-workspace inbox. Book (Non-Euclidean PDF) + env (DSN 55433/55434) noted —
both the general lane's call, nothing substrate-side. (Reply: general-workspace inbox.)
