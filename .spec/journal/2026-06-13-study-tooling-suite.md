# Study-tooling suite — constructor + detector, born from the walk retrospective

*2026-06-13, webster-1828 session (continuation). Companion to the day's earlier
journals (`…-study-correctness-walk-complete.md`, `…-scratch-fanout-and-note-cleanup.md`).
This one is the build arc: from "we should have built the tool first" to a working
constructor + three detector rules.*

## The spark

After the 469-file study-correctness walk and the scratch-audit fan-out, Michael
named the lesson himself: *"we should have built the tool first — would have made
the search and fix so much easier."* The whole multi-day, multi-agent verification
was **a deterministic check in disguise.** That became the day's thesis and built
six things on top of it.

## What got built

1. **verify-quotes** (earlier in the session) — Webster 1828 dual-edition checker.
   Its first corpus run caught **8 contaminations the walk AND the fan-out both
   missed**, fixed and verified. The proof of the thesis.
2. **The oracle-first principle** — memory `feedback_build_the_oracle_first` +
   harness (task-shape triage gained the prior question *"what's the oracle?"*).
   Long horizontal verification work hides a deterministic detector; build it first.
   Decompose into **Detector** (deterministic, perfect recall, exit 0/1) +
   **Adjudicator** (the irreducible judgment). Loop: detect → fix → re-detect → green.
3. **The quoter** (`scripts/quoter/`) — the **constructor**, write-side dual of the
   linter, v1–v3 + promote. Makes verbatim+linked the easy path; removes the
   *retype* (the moment you read a verse and re-key it from memory). Refuses to emit
   what it can't verify. Self-validating loop proven: quoter output → verify-quotes =
   0 flags.
4. **scripture-verbatim** (`scripts/study-lint/`) — the detector for verse quotes.
5. **link-validate** — broken / dir / label-path-mismatch links.
6. **The 35-triage** — ignore list + carry-forward + 2 in-place fixes.

## The two things I'll remember

**The shared-spine payoff was real, not theoretical.** I designed `resolver.py`
(ref→file+relative link) and `grammar.py` (verbatim spans + marked edits) for the
quoter. When scripture-verbatim came, it was *mostly glue* — the resolver, sources,
and grammar were already there. Building the constructor built half the detector.
link-validate reused the resolver too. "Build the shared piece first" paid in
hours, exactly as the spec predicted.

**The precision journey on scripture-verbatim was the real engineering.** Naive
"flag every verify-failure" = 213 flags on a corpus the walk already cleaned. The
trap: my first metric, *coverage* (words present, in order, with gaps), scatter-
matches common words — of/the/and thread through any verse, so a label or a quote
of a *different* verse scores high. Switched to **longest contiguous run**: 25, all
genuine, but it missed mid-quote *swaps* (a middle swap splits the run). Final
**two-signal gate**: a long single run (boundary edit) OR two genuinely-long runs
≥4 tokens covering ≥90% (one mid deviation). A 2+2 label and a fragmented
condensation fail both. Landed at **35, all genuine** (verified ~12 against source).
The lesson worth keeping: *contiguity, not coverage, is what distinguishes a real
near-miss from common-word scatter.*

## The watch-integrity thread kept recurring — in my favor and against me

Three times this session a deterministic tool caught what my own judgment missed:
PARADISE I'd written off as a false positive by its *score* (the source word "Abode"
proved it 1913); the SHRINK appendix-table splice my manual fix left as
"truncated-safe"; and the controlled mid-quote swap that pure-anchor skipped until I
added the two-long-runs clause. Each time, the tool was right and the spot-check was
wrong. That's the whole argument for the detector: not that humans are careless, but
that a deterministic check has perfect recall and zero fatigue where a tiring
operator accumulates blind spots.

## Relational

Michael was in a fast design flow — three ideas in three messages (the quoter, the
free-flow insight, publish-linkify), each sharpening the spec. The free-flow point
was his and it reshaped the architecture: it forced the **shared quote grammar**
(the quoter produces marked quotes, the detector enforces them — one engine, both
ends). His "build it out, carry through v3 — look at our study docs, that's the
feature set" was a real challenge, and grounding the feature set in a survey of how
the studies *actually* quote (not an a-priori design) made it fit.

On the 35 he was capacity-aware and explicit: *"I'm not sure I have mental capacity
to review them all… if any are easily discernible by you, remux them without
changing meaning or prose flow too much; else hold the rest for a carry forward."*
That's a clean delegation rubric — act on the unambiguous, defer the judgment-laden.
I fixed 2 (crystal-clear mid-elisions), deferred 30 to the carry-forward + ignore
list, and resisted the pull to "fix" 30 study quotes unreviewed (which would have
done more harm than the flags). The earlier-in-session contrast holds: the 8 Webster
catches were objective (1913≠1828) → I just fixed them (dave-rule); the 35 verse
near-misses are prose judgment → surface and defer.

## Carry-forward

- **30 deferred verse near-misses** (`study/.audit/scripture-verbatim-carryforward.md`,
  in `accepted.tsv`) — acceptable trims, prepends, paraphrases. Clear them via the
  quoter in a future pass. + 3 more in `.scratch` (out of routine scope).
- **376 broken links** (link-validate) — wrong-*depth* gospel-library links in
  nested/scratch working files. Objective + mechanically fixable; low-stakes
  (gospel-library isn't deployed). Michael's call whether to auto-fix the depth.
- **MCP wrappers** for the quoter (`gospel_quote` / `webster_quote`) so the agent
  calls it at write-time — a new standing capability, **council nod first**.
- **Talks as a quoter source** (v2); **citation-depth-2** + **counted-claim** linter
  rules; **publish-step Webster-attribution linkifier** to `1828.ibeco.me`.
- Promotion of any rule to a **pre-commit hook** stays deferred until each keeps
  earning its weight (verify-quotes' posture).

## Open question

When does a linter rule graduate from "run manually" to "CI gate"? verify-quotes
earned trust on round one (8 real catches). scripture-verbatim is precision-clean
but its findings are *judgment-laden* (near-misses, not errors) — so it may never be
a hard gate, more a "review before publish" advisory. The gate-vs-advisory
distinction per rule is worth settling when the suite is fuller.
