# Webster "1828" Data Integrity — Investigation & Remediation Brief

**Status: OPEN — awaiting a dedicated session.** Created 2026-06-09 by Claude Fable 5
during the scripture-book v4 honesty walk. Michael: *"I kinda feel lied to with the
source of the dictionary I got from git. It's something I want to resolve, and I'll
probably go through all of my published works (study / cpuchip.net / book) to fix
issues there."* This file is the self-sufficient handoff for that session.

---

## What we know (verified 2026-06-09)

**The workspace's "Webster 1828" tooling serves Webster *1913* (Revised Unabridged)
text under the 1828 label.** Caught during the book's v4 audit; verified three ways:

1. **Internal anachronisms.** The tool's *spirit* entry cites **"U. S. Disp."** (the
   United States Dispensatory, first published **1833**), quotes **N. P. Willis** (1830s
   travel writer), Keble, and defines "stannic chloride" — impossible content for an
   1828 text.
2. **Authoritative mismatch.** webstersdictionary1828.com (the standard faithful 1828
   transcription) lacks the quoted phrasing entirely. Genuine 1828 *spirit*: "5. The
   soul of man; the intelligent, immaterial and immortal part of **human beings**."
   "6. An immaterial intelligent substance." The tool's version — "an intelligence
   conceived of apart from any physical organization or embodiment; vital essence,
   force, or energy, as distinct from matter" / "…immortal part of **man**" — is
   **verbatim Webster 1913**.
3. **Second word confirmed.** Genuine 1828 *intelligence* = "1. Understanding; skill.
   2. Notice; information communicated… 3. Commerce of acquaintance… 4. **A spiritual
   being; as a created intelligence.**" The tool returns 1913's "the exercise of the
   understanding" / "the capacity to know or understand" instead.

**The likely chain (to be confirmed):** `scripts/webster-mcp/data/webster1828.json.gz`
← per `scripts/webster-mcp/README.md` line 83: **github.com/ssvivian/WebstersDictionary**
(MIT) — and README line 180 says the dictionary content is under the **Project
Gutenberg License**. Project Gutenberg's Webster etext is the **1913 Revised
Unabridged**, not 1828. So the data was probably honest *Webster-Unabridged-1913* that
got relabeled "1828" somewhere between that repo and our tooling. **Open question for
the investigation: did ssvivian's repo claim 1828, or did we assume it?** (Check their
README/commit history before assigning blame — we may have done the mislabeling
ourselves when we named the file `webster1828.json.gz`.)

**Why nobody caught it:** the 2026-05-29 book fact-check "verified" Webster quotes
*against this tool* — wrong-path verification (same failure shape as the glm-streaming
misdiagnosis, memory `feedback_verify_via_real_path`). A verification tool can itself be
the wrong source.

## Blast radius (what's downstream of the bad label)

| Surface | Exposure | Notes |
|---------|----------|-------|
| `scripts/webster-mcp` (`webster_define`, `define`, `webster_search`, `webster_search_definitions`) | **Root cause.** | 98k+ entries; "1828" in tool descriptions + README. |
| **1828.ibeco.me** / `projects/1828-illuminated/` | **High.** The site's whole identity is "1828"; 853-word tier list + pre-fetched defs built from this data. The book QR-links to `1828.ibeco.me/word/spirit` + `/word/intelligence`. | |
| ***Beyond the Prompt*** (book) | Identified: Ch 0 *spirit* quotes, Ch 1 *intelligence* quotes, Ch 9 "warily" gloss. | Already logged as **F-19 / F-20 / F-21** in `projects/scripture-book/.draft/20260609-v4-walk-findings.md`; fixes ride the book's v4 chat walk (genuine 1828 requotes drafted there — they *improve* both passages). |
| **cpuchip.net published studies** | **Unaudited.** `morm-8-three-glories-reading` leaned heavily on "1828" entries (fine / gain / glory / pollute / apparel; "telestial absent from Webster"); any other published study quoting webster_define as 1828. | |
| `study/` (unpublished workspace studies) | Unaudited; many used webster-analysis over months. | |
| `.github/skills/webster-analysis/` + `.claude/` twin | Skill teaches the tool as 1828. | |
| becoming app / ibeco.me | Check whether any surface quotes "1828" definitions. | |

## The remediation plan (proposed — ratify order with Michael)

1. **Forensics first.** In `scripts/webster-mcp`: inspect `data/webster1828.json.gz`
   provenance (our git history for the data commit, the README claims), then check
   ssvivian/WebstersDictionary's own README — establish *who* introduced "1828."
   Sample ~10 diverse words from the data against webstersdictionary1828.com to
   confirm it's uniformly 1913 (not a mixed set).
2. **Get the genuine 1828.** Known sources to evaluate: webstersdictionary1828.com
   (site, scrape-unfriendly?), the `1828-dictionary` datasets floating on GitHub
   (verify *those* the same way — sample against the site before trusting), or the
   original facsimile text. **Whatever source we adopt, verify it with the same
   anachronism + spot-check method before shipping it.**
3. **Fix webster-mcp honestly:** serve BOTH editions under truthful names —
   `webster_define` → genuine 1828; the current data stays available as 1913 (it's
   still useful — KJV-era it is not, but it's a fine general historical dictionary).
   Update tool descriptions, README, and the webster-analysis skill (both trees).
4. **Rebuild 1828-illuminated** on the genuine data (tier list + pre-fetched defs) and
   redeploy 1828.ibeco.me.
5. **Audit the published works** (Michael wants to walk this together):
   - grep cpuchip.net `content/studies/*` + workspace `study/` for `1828|Webster`;
   - re-verify every quoted definition against genuine 1828;
   - fix + republish (cpuchip.net pushes auto-deploy);
   - the book's F-19/20/21 land via the v4 chat walk in the book repo.
6. **Close the loop:** docs/06 entry update (resolution), memory note, and a learnings
   entry — the durable lesson is *verify the edition of a source, not just the quote*.

## Verification standard for this whole effort

**webstersdictionary1828.com is the authority** for 1828 text until a better facsimile
source is ratified. Any "1828" claim that ships (tool output, site, study, book) must
trace to it — not to our own mirror, which is the thing under repair.

## Cross-references

- `docs/06_tool-use-observance.md` → "June 9, 2026 — webster-mcp serves Webster 1913
  text under the 1828 label" (the original incident log, same evidence).
- `projects/scripture-book/.draft/20260609-v4-walk-findings.md` → F-19, F-20, F-21, SQ-1.
- Memory: `feedback_verify_via_real_path` (the recurring lesson).
