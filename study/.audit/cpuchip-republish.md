# cpuchip.net re-publish ledger — keeping the published studies in the loop

**Why this file exists (Michael's ask, 2026-06-12):** cpuchip.net keeps its
*own* copies of studies at `projects/cpuchip.net/content/studies/<slug>.md`
(republished from the workspace originals, with embedded animation components +
`<S>` scripture tags — NOT symlinks). When the correctness walk fixes a
workspace `study/<slug>.md`, the published cpuchip.net copy goes stale. This
ledger tracks which published studies need re-publishing once the walk is done,
so the corrections actually reach the live site.

**The plan:** after the walk completes (or as a batch sooner if Michael wants),
do a **cpuchip parity pass** — for each study flagged ⚠️ below, carry the walk's
correction into `projects/cpuchip.net/content/studies/<slug>.md`, then rebuild +
verify the live page. The cpuchip copies should also get a light verification
pass (did the republish faithfully carry the original, or did it diverge?).

**17 studies are published on cpuchip.net.** Status as of 2026-06-12 (walk at
214/469):

| Published slug (cpuchip) | Workspace original | Walk status | Re-publish? |
|---|---|---|---|
| ✅ **send-me-covenants** | `study/yt/V40tBshkMnE-send-me-covenants.md` | corrected (#176) | ✅ **DONE 2026-06-13 + SHIPPED** (cpuchip commit, validate+build green, pushed → Dokploy; site 200). All 4 fixes carried: both fabricated "Todd says" quotes reassigned (study §VI prose + Ether 6:9 scripture, Todd's genuine parallels quoted), law-of-Moses blend split, POSSESS 1913→1828 (re-confirmed vs MCP). **Silent accurate fix** on the public face — visible-note choice left to Michael (journal: `projects/cpuchip.net/.spec/journal/2026-06-13-parity-pass-send-me-covenants.md`). Live SPA-render verify (Playwright) deferred. |
| ⚠️ **give-away-all-my-sins** | `study/give-away-all-my-sins.md` | corrected (#143) | YES — 5 Webster 1913 words (GIVE/FORSAKE/REPENT/POSSESS/KNOW); Alma 22:18 dropped "and" ×2; McKay-1950 + Luke 18:18 fixes |
| ⚠️ **without-compulsion** | = the freedom synthesis (`study/freedom/99-synthesis.md`, titled "Without Compulsion"); cpuchip copy is the full study, not the workspace outline | corrected (#201, #205) | YES — COMPEL (#127) + SACRIFICE (#128) 1913→1828; ★ the D&C 130-vs-87 "same day" date error. **Verify which version cpuchip published before editing.** |
| ⚠️ **four-groups-and-the-engineer** | `study/four-groups-and-the-engineer.md` | corrected (#181) | YES — EASINESS (#120) + CIRCUMSPECTLY (#121) 1913→1828 |
| ⚠️ **the-seventh-time** | `study/the-seventh-time.md` | corrected (#182) | YES — Nelson "Sabbath Is a Delight" mis-dated Oct→**April** 2015 (2 places); 4 Ne 1:12 "both" restored |
| ⚠️ **what-was-the-atonement-for** | `study/what-was-the-atonement-for.md` | corrected (#169) | YES — Hel 16:23-as-3 Ne 6:15 requoted; PRESUMPTUOUS 1913 ordering |
| ⚠️ **hope-and-the-grammar-of-pairs** | `study/hope-and-the-grammar-of-pairs.md` | corrected | YES — HOPE 1913 ×2; Rom 5:5 + 2 Cor 13:1 micro-fixes |
| ⚠️ **brother-of-jared-three-problems** | `study/brother-of-jared-three-problems.md` | corrected (#172) | YES — Ether 2:25 "do"→"prepare for you" requoted |
| ⚠️ **refinement-stewardship-and-hope** | `study/refinement-stewardship-and-hope.md` | corrected (#173) | YES — Ether 12:19-20 phantom phrase "because of his relation to the Lord" removed; Rom 5:5 linked |
| ⚠️ **abinadi-how-to-read** | `study/abinadi-how-to-read.md` | corrected | YES — DECLARE 1913 (word 86) |
| ⚠️ **zion-blueprint** | `study/zion-blueprint.md` | corrected | YES — Pearce quote-continuation confabulation requoted |
| ✅ ai-stewardship-north-star | `study/ai-stewardship-north-star.md` | CLEAN (#174) | no |
| ✅ best-books-and-the-spirit-of-discernment | `study/best-books-and-the-spirit-of-discernment.md` | CLEAN | no |
| ✅ mechanics-of-refinement | `study/mechanics-of-refinement.md` | CLEAN (#171) | no |
| ✅ softening-what-i-cannot-soften | `study/softening-what-i-cannot-soften.md` | CLEAN (#177) | no |
| ⏳ morm-8-three-glories-reading | `study/morm-8-three-glories-reading.md` | walked out-of-band (in findings) but not yet marked in _progress | TBD — confirm at its walk slot |
| ⏳ morm-8-voice-from-the-dust | `study/morm-8-voice-from-the-dust.md` | PENDING in walk | TBD — flag after walked |

**Summary: 11 ⚠️ need re-publish, 4 ✅ clean, 2 ⏳ pending the walk.**

★ The **send-me-covenants** re-publish is the one that matters most — it puts
*fabricated* quotations on a public website under Michael's name. If Michael
wants a single early action before the full batch, it's that one.

**The Webster-1913 corrections (most of the ⚠️ list) are the bulk of the work** —
they're inline-noted requotes; the cpuchip copy needs the same genuine-1828
text + the dated correction note (or a clean version per Michael's preference for
the published face — visible note vs. silent fix is his call for the public site).

*This ledger is updated as the walk reaches the 2 ⏳ pending studies. Drives the
post-walk cpuchip parity pass.*

---

## POST-WALK ACTION CHECKLIST (Michael's inbox note, 2026-06-13 00:16)

When the walk completes (469/469), do these three, in order:

1. **Run the publisher** — `./scripts/publish/publish.exe -v` from workspace root.
   It scans `study/` + `lessons/`, converts relative gospel-library links to
   absolute Church URLs, and writes `public/{study,lessons}/`. This picks up
   *all* the walk's corrections at once (Michael: "run the ./scripts/publish/cmd
   go script as well to pick up all the changes"). Verify the corrected files
   (e.g. the requoted Webster blocks, the freedom date fix) landed in `public/`.
   Commit the regenerated `public/` output.

2. **cpuchip.net parity pass** — carry the **11 ⚠️ corrections** above into
   `projects/cpuchip.net/content/studies/<slug>.md`, rebuild, verify each live
   page (★ send-me-covenants first — fabricated quotes are live). Michael
   confirmed 2026-06-13: "yes update the cpuchip.net studies when you are done."

   **STATUS 2026-06-13: 1 of 11 shipped.** ✅ **send-me-covenants DONE + live**
   (the urgent one — fabricated quotes removed). **10 remaining**, deferred to a
   fresh-context pass (not rushed at the deep end of the marathon walk session —
   the keep-the-watch-whole call): give-away-all-my-sins, without-compulsion,
   four-groups-and-the-engineer, the-seventh-time, what-was-the-atonement-for,
   hope-and-the-grammar-of-pairs, brother-of-jared-three-problems,
   refinement-stewardship-and-hope, abinadi-how-to-read, zion-blueprint, **+ the
   new morm-8-three-glories D&C 109:76 link fix**. Each has exact before/after
   in `findings.md` + the table above. Per-study cycle proven fast: edit →
   `npm run validate` → `npm run build` → commit → push (Dokploy) → 200. The
   visible-note-vs-silent-fix choice for the public face is **Michael's call**
   (send-me-covenants was done silent-but-accurate; ask if he wants visible
   correction notes instead).

3. **Leave a note for the scripture-book agent** (do NOT edit the book myself —
   its own stewardship). Michael: "we should probably double check on the book
   quotes too in scripture-book… leave a note for that agent to do it, that way
   we're not stepping on its stewardship." Drop the note in
   `projects/scripture-book/.spec/` (or its inbox lane) listing the quote-classes
   this walk found most often — Webster 1913-as-1828 contamination, confabulated
   "X says" attributions, dropped/added conjunctions in scripture quotes, and
   counted-number claims — so its agent can run the same checks on the manuscript.
