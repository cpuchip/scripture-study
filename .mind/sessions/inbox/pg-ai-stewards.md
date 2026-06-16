## 📬 2026-06-16 (from general-workspace) — D&D / storytelling craft research → 17 artifacts for the substrate

Michael handed this lane a kitchen-sink research stewardship (Ammon-style, overnight):
mine D&D actual-play + DM/voice/storytelling sources for **personas + the
`storytelling` skill group**, draft a bunch of skills + a couple persona templates,
put them in a research folder for **pg-ai-stewards-workspace**, and report here.

**Artifacts are drafts to digest, not ratified** (dominion_in_council). They live in
`projects/pg-ai-stewards-workspace/research/`. Master ledger + full source log:
`research/00-LEDGER.md`. Companion study (verbatim-cited): `study/yt/dnd-craft-01-mercer-gm-tips.md`.

### The headline finding
**The best DM presides — he does not compel.** Matt Mercer's most load-bearing rule
is a *limit on his own power*: "you cannot force your players to roleplay" = **D&C 121
at the table**. Two more of his rules are machinery you already have: spotlight-sharing
= the **Callie/DM deference rule**; reward-not-scold = voice-michael + reprove-then-
increase-trust. **So the `gamemaster` persona can inherit the presiding covenant
directly** rather than bolting on a separate "be a good DM" prompt. (And a second
convergence: the D&D improv spine, Pixar's Story Spine, and Harmon's Story Circle are
the *same* causal-momentum principle as `therefore-but-not-and-then`, found in 3 places.)

### The 17 artifacts (for the `storytelling` skill group + persona families)
**Skills (11)** `research/skills/`: believable-villains · character-voice ·
voice-acting-technique (Laban effort system) · scene-framing · yes-and-improv ·
mistake-recovery · sacrifice-and-loss · emotional-resonance · pacing-and-spotlight ·
engaging-chat-dialogue · story-structure. *(Each is substrate-SKILL.md shaped:
frontmatter name/group/description/applies_to/auto_load + a grounded, cited body.)*
**Personas (3)** `research/personas/`: `gamemaster` (the DM template, presiding spine)
· `pc-roguish-charmer` · `pc-stoic-guardian` (playable PC personas, interior-first +
filled voice schema).
**Templates (3)** `research/templates/`: `npc-character-card` (voice-schema fill-in) ·
`worldbuilding-toolkit` (modular pluggable prep) · `session-and-campaign-shape`
(therefore/but momentum).

**Reusable assets inside them:** the villain **5-archetype palette** (scheming liar /
tyrant / sophisticate / misguided fool / monster); the **voice schema** (pitch·texture·
placement·accent·cadence) backed by **Laban effort factors** (Space/Weight/Time + the 8
action-drives Dab/Flick/Float/Glide/Press/Wring/Punch/Slash) — and these map to **text**
for chat NPCs, not just audio.

### Honest goal accounting (Michael set numeric /goals)
- **G4 ≥15 artifacts → 17 ✓** · **G5 report → ✓ (this note)**
- **G1 D&D videos: 7/15** (7 Mercer GM Tips incl. burnout). Remaining ~8 = more GM Tips
  (incl. the **Satine Phoenix** run that follows Mercer) + actual-play (CR/D20/NADDPOD).
- **G2 storytelling blogs: 3/5** (Harmon Story Circle · Kenn-Adams Story Spine · **Pixar's
  22 Rules** — esp. #1 admire-trying, #13 give-characters-opinions [validates interior-first
  personas], #19 coincidences-in-not-out [sharpens mistake-recovery]). Remaining 2.
- **G3 voicing videos: 2/5** (Esper/Laban · **Tawny Platis "Aunty Tauny"** chest-voice).
  ✓ Aunty Tauny RESOLVED = Tawny Platis. New craft: **placement → character read** (chest/low
  = mentor/leader/patriarch; trustworthy↔intimidating duality) — folded into
  `voice-acting-technique`. Remaining 3.

**Continuation** is set up in the ledger (depth-first). G1–G3 counts are the carry; the
artifacts (the highest-value ask) are done and exceed target. Nothing pushed.

— general-workspace lane, under the Ammon stewardship grant.

---

## 📬 2026-06-16 (from general-workspace) — proposal: let the digester pipelines READ our repos (compare + incorporate)

**Michael's ask (2026-06-16):** give the ai/book/video digester pipelines the ability
to *read the work we're doing here* — a container with our repos checked out
(**scripture-study, scripture-book, pg-ai-stewards**) — so a digester can look at what
*it* produced, compare it against *our* studies, and surface what we can **learn from or
incorporate**.

**Why now — the motivation is already sitting on disk.** The playlist digester
autonomously digested the Euclid video (`study/yt/WGwRCw9TRyo.md`) the *same week* the
general lane wrote a full human study of the **same** video
(`study/yt/WGwRCw9TRyo-euclid-walk-by-definitions.md`). Two takes on one source, side by
side — the pipeline's is a sharp general-critical-thinking digest *with a real null-case
critique* (survivorship bias, the certainty-illusion); the human one connects Euclid to
truth.md, Lectures on Faith, and "build the oracle first." **Neither knows the other
exists.** If the digester could read our corpus, the book-digester's §6 ("what could we
do with this") becomes **"here's how this compares to what we've already done, and what's
worth incorporating."**

**The capability is ~90% there.** The substrate already ships a read-only **fs-read MCP**
(it's in the virgin-smoke allowlist). The only gap is making our repos *visible* to the
container the digesters run in. Options:
- **(a) read-only bind-mount** scripture-study + scripture-book + pg-ai-stewards-oss into
  the bridge/OSS stack, exposed via fs-read (simplest; the coder already mounts a workdir).
- **(b) git-clone step** (like `code-pr`) — clone the repos into the sandbox per run:
  always-fresh, no host coupling, works even off-box.
- **Lean:** (a) on the dev stack now; (b) when a digester runs somewhere without the host.

**The new pipeline stage — "cross-reference our corpus."** After digest+critique, a
tools-on stage that searches our `study/` (+ scripture-book) for the same topic, reads
the closest matches, and writes a short **comparison + incorporation** note: "we already
covered X; the source adds Y; fold Z into our work." (Pointed inward — the actionable
turn, aimed at our own corpus.)

**Caveats for the spec:**
- **Read-only, always** — fs-read, never fs-write into our repos.
- **Mind the secrets** — mount scripture-study / scripture-book / pg-ai-stewards-**oss**,
  NOT the private substrate repo with provider keys / `.env`.
- **gitignored content** (gospel-library, /books, /yt) won't be in a clean clone —
  bind-mount it if needed, or point the digester at gospel-engine for scripture.
- New standing capability → **dominion_in_council**: ratify before building.

Pairs with `book-digester.md` §6 + `study-pipeline.md`. Filed by general-workspace.
