---
date: 2026-06-13
title: Scratch-audit fan-out (dogfood) + the correction-note cleanup
lane: webster-1828
type: milestone
---

## What happened

After the study-correctness walk (469/469) and its three post-walk asks, Michael
steered three more things to completion in one long session: the carry-forward
flags, a **multi-agent scratch audit**, and a full **correction-note cleanup**.

## The fan-out — and the distinction it proved

Michael asked whether the scratch-file verification would benefit from "farming
out to a bunch of opus agents." It did — and the *why* is the durable lesson:

- The **note-removal cleanup** is centralizable (one known pattern, deterministic
  to locate) → single careful pass beats fan-out.
- The **scratch verification** is the opposite: independent, judgment-and-tool
  -heavy work, per file, across 62 files — the same profile as the walk itself.
  That's exactly where fan-out earns its keep.

6 Opus subagents, 6 batches, parallel. Each verified every quote against the real
source (gospel-library grep, the repaired Webster MCP, talk files), corrected the
stale, flagged the non-verifiable. I presided per Abraham 4:18 — staged it (wave
1 = 2 agents to validate the spec, then wave 2 = 4), reviewed every report, and
spot-verified the high-stakes catches against source before accepting (image→Matt
22:20, D&C 8:11, Mosiah 18:29, D&C 107:87, transgression, ordain — all confirmed).

**~75 corrections across 23 of 62 files.** Dominant class: Webster 1913-as-1828
(the same contamination the walk fixed in the *studies* but never in scratch).
Plus a class the walk under-checked: **fabricated citations *inside* Webster
entries** (NEVERTHELESS "Heb 12:11", PRESUMPTUOUS "Ps 19:13", STILL "1 Kgs
19:12"…), **confabulated scripture** (D&C 8:11, Mosiah 18:29 — wholly invented in
the scratch; D&C 107 mis-numbered), and 3-4 wrong attributions/titles.

★ **The humbling, valuable part:** the fan-out surfaced one error that had
reached a *study* file (not just scratch) — alma5's "Webster cites Gen 1:27 under
*image*" (genuine = Matt 22:20) + a 1913-flavored *countenance* quote. The walk
had marked that chapter note CLEAN. Root cause: the walk verified alma5's
*scripture* but never ran the Webster MCP on its *image/countenance* quotes.
**Lesson logged:** a study's Webster quotes need the MCP even when its scripture
checks clean. Parallel fresh-eyes-per-file caught what one serial operator (me)
missed — the strongest argument for fan-out on independent-verification work.

## The cleanup — clean is the end state

Michael's call: studies + scratch read clean; `findings.md` + journals are the
durable record (the inline annotations were temporary scaffold). Removed ~136
inline `*(Requoted/Corrected 2026)*` + `*Correction (…)*` notes across the study
tree (tested perl for the bulk; manual for the nested-italic ones; CRLF-aware
blank tidy), and the one "Correction" banner from the published cpuchip morm-8
study. Final scan: zero notes remain. The scratch provenance banners I'd added
earlier in the session were also removed (same clean-preferred call).

## Relational / process notes

- Staged the fan-out (validate-then-scale) rather than blasting all 6 — the
  presiding watch applied to my own delegation. Git as the safety net, as Michael
  named ("git's cheap if we run into issues").
- Honest reporting held: I named that the fan-out caught a class my walk
  under-checked, and that one study error slipped my "CLEAN" verdict. The
  covenant's value isn't a spotless record; it's the true one.

## Carry-forward

- **2 Ne 5:21 / Morm 5:15 / Alma 3:6** — bin-4 curse-question, parked with
  Michael's spiritual reading for post-project review.
- Optional: the without-compulsion line-86 methodology footer (Michael's call).
- The scratch tree is now normalized + clean; findings.md holds the full audit.
