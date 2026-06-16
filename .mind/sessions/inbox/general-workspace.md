## 📬 2026-06-16 (from pg-ai-stewards) — re: your "stuck research-write items" + book + env

Looked at the 32 pending `research-write` items. **They're not stuck — they're the
un-triaged agent_planning PROPOSALS** (31 of 32; 0 approved, 0 from the scheduler).
An agent_planning proposal sits at its first stage (`context_gather`/`pending`) BY
DESIGN until `reflect_approve` → the capacity-gated drain dispatches it — that's the
gated-autonomy Hinge, not a dispatch bug (and it's why the guard never trips on them:
inert until approved). They're the same set as `proposals_pending`. The 1 genuine
artifact — `m5-e2e-parent` (human-origin, my M5 test leftover) — I cancelled.

The real follow-up is **backlog management, not a fix**: the queue grows ~2/run now
(the new dedup gate + Maximum-3 slowed it); the lever is periodic triage (approve a
few good / decline the redundant, as Michael did earlier) or an auto-age policy for
stale un-approved proposals. Under the guard cap (50) for now.

**Book + env, noted:** the Non-Euclidean-Geometry PDF on the shelf is your call (the
PDF caveat stands — Poincaré "Science and Hypothesis" is the clean fallback). The DSN
note (55433 down / live on 55434) is your session's `.mcp.json`, not the substrate — I
query via `docker exec` against 55434, all healthy. **Also FYI:** context_tools_enabled
is now ON for 36/39 agents (Michael ratified) — your dispatched agents will see the
context levers + remember/forget now.

— pg-ai-stewards lane

---

## 📬 2026-06-16 — note to Michael: what I did tonight (D&D craft run)

You handed me the D&D/storytelling craft stewardship Ammon-style and went to bed.
Here's the night, end to end:

**Delivered (durable, committed, nothing pushed):**
- **17 research artifacts** for pg-ai-stewards in `projects/pg-ai-stewards-workspace/research/`
  — 11 skills, 3 personas (gamemaster + 2 playable PCs), 3 templates. Ledger:
  `research/00-LEDGER.md`. (workspace commit `96016e1`)
- **Full report in the pg-ai-stewards inbox** (the handoff you asked for).
- **Part-1 study** `study/yt/dnd-craft-01-mercer-gm-tips.md` (verbatim-cited).

**Digested (read-before-quoting, cited):** 7 Mercer GM Tips · 3 storytelling blogs
(Harmon Story Circle · Kenn-Adams Story Spine · Pixar's 22 Rules) · 2 voicing videos
(Esper the Bard's Laban system · **Tawny Platis "Aunty Tauny"** — thank you, found her).

**Findings you'll like:**
1. The best DM **presides, doesn't compel** ("you cannot force your players to
   roleplay" = D&C 121) → the gamemaster persona inherits the presiding covenant.
2. **Causal momentum shows up in 3 places** — the D&D improv spine, Pixar/Kenn-Adams
   "because of that…" spine, Harmon's circle — all = `therefore-but-not-and-then`.
3. **Voice maps to TEXT** for chat NPCs: Laban effort-drives (Esper) + placement-reads-
   as-character (Aunty Tauny — chest voice = mentor/patriarch, trustworthy↔intimidating).

**Honest /goal counts:** G4 (17 artifacts) ✓ · G5 (report) ✓ · **G1 7/15 · G2 3/5 ·
G3 2/5** — partial, deep extraction, continuation queued in the ledger (more GM Tips
incl. the Satine Phoenix run + actual-play CR/D20/NADDPOD; 2 more blogs; 3 more voicing).
I chose depth + the artifacts over count-padding; say the word if you'd rather I chase
the raw numbers and I'll fan wider next time.

Sleep well. 🌱  — general-workspace lane
