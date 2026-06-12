# Night shift — the archive dig + the voice profile (2026-06-11→12)

Michael handed me the night ("the night is yours steward") with two gifts landing:
the old-machine VS Code/Copilot archives and `external_context/old-code/` (the 2025
repos). Three arcs closed before morning.

## 1. The voice profile (his ask: "build a voice profile of me… bad spelling and all")

Built from four corpora (~163K words, Sept 2025 – May 2026, three harnesses):
**`docs/voice-profile-michael.md`** — the linguistic fingerprint (the `lets` 96%
signature, ZERO em-dashes in 101K words, the typo classes, "Okay" pivots, numbered
rulings, rule+exception inline, evidence-first debugging, empathy-for-the-model) plus
**the Book Gauge**: an 8-marker / 8-tell rubric for "does this first-person passage
sound like Michael or AI?" Both voice-michael skill trees now point at it. Raw corpora
left UNTRACKED (`.spec/scratch/voice-corpus-*.txt`) pending his call — note the May
extracts (`prompts_recent.txt`) are already tracked/public, so precedent exists either
way.

## 2. G-3 fully closed — the book's last unverified sentence, verified verbatim

Copilot session 2025-10-05: "What do you think of this roadmap? I'm feeling a bit
overwhelmed by the amount of work left to do to have a functioning product." The AI's
reply references Milestones 7+8 (the roadmap was EXACTLY eight at that moment) and
preaches P2's own lesson back through time: "treating it as a checklist for launch
will overwhelm you… treat it as your long-term vision." **Every factual claim in
Beyond the Prompt is now verified or deliberately Michael's.** (Findings file updated +
pushed, book repo `74dbfd0`.)

## 3. The archaeology (his ask: "help me sort it out… have fun for me")

**`docs/project-archaeology-2025.md`** — the 2025 timeline attested by git+sessions:
journal-mcp (Sept, the first MCP server) → forkirk (Sept 18–Oct 2, the quoter, set
down) → the Oct 5 roadmap moment → astrotreks (the big bridge sim) → storygames (Oct
22, the daughter's characters) → mobile-games/Shields Down (Oct 25, founding chat with
ten questions) → simple-games (Nov 18–Dec 27, eight wired networked games). Plus the
physical-bridge thread (trek-experiments: ESP32 consoles, KiCad, grant research),
homeschool-keeper, farm-store. Copy gotcha recorded: astrotreks/storybook-web/
trek-experiments came over WITHOUT .git.

## Privacy actions taken (public repos!)

- Scrubbed the co-worker/season context from the book repo's findings file (was pushed
  in an earlier commit; follow-up commit removes it from HEAD — flag to Michael that
  history retains it if he wants a deeper scrub).
- forkirk described neutrally everywhere public; the fuller personal context stays in
  conversation only.
- New voice corpora NOT committed.

## For Michael's morning (small decisions)

1. P6 "seven games" — final wired roster was eight (snake + wizard TD are networked).
   Keep seven (true during the arc described) or update?
2. Optional P2 enrichment: quote the AI's Oct-5 map-not-debt counsel? (His call; the
   chapter stands verified without it.)
3. Commit the raw voice corpora, or keep them local-only?
4. The findings file's git HISTORY still contains the scrubbed sentence (one commit);
   fine to leave, or want a history rewrite on the book repo?
5. Root remains unpushed (his push; includes voice profile + archaeology + skills).
