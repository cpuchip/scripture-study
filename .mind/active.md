# Active Context — the in-flight board

> **2026-06-11 (evening) — ★ pg-ai-stewards GOES PUBLIC: extraction ratified, repo seeded.** Michael's executive ask: `github.com/cpuchip/pg-ai-stewards` created (PUBLIC) → `projects/pg-ai-stewards-oss/` (private substrate stays at `projects/pg-ai-stewards/`). Spec ratified (`extraction-plan.md`, pushed `25355f2`): **v0.1 = core + persona-host** · clean-room fresh history · cutover after virgin `docker compose up` · **license = source-available, individuals free / companies pay** (BUSL-1.1 w/ MIT change-date recommended vs PolyForm Small Business — Michael picks; repo all-rights-reserved meanwhile; commercial model ⇒ CLA needed). **Deliverable #1 = "Anatomy of a Turn"** — rebuilds the lost mental model, becomes the OSS arch page + the cpuchip.net/projects/pg-ai-stewards animation storyboard (Remotion, per munder-difflin's landing-remotion). munder-difflin cloned to `external_context/` (STUDY not replicate — its HIVE patterns converge with our session-lanes/substrate). P4 = playground machine under stewardship; P5 = the AI-office vision (agents-for-people via MCP into chattermax rooms). Gotcha: `gh repo create --clone` init'd a stray .git INSIDE the existing workspace substrate dir (name collision) — caught + removed, zero commits lost. Task #151. Lane: `.mind/sessions/pg-ai-stewards-oss.md`.

> **Board discipline (ratified 2026-06-11):** this file holds ONLY what is
> genuinely in flight. The full record lives in `.spec/journal/` (and
> subproject journals); the old banner ledger is preserved verbatim at
> `.mind/archive/active-ledger-thru-2026-06-11.md`. Multi-session
> coordination lives in `.mind/sessions/` (lanes + inboxes — read the README
> there once). **When an arc closes: journal it, then delete its lines here.**
> If an open thread of yours is missing below, re-add it as ONE line with a
> journal link — do not rebuild banners. Target size: a few hundred lines max.

## Open threads

**Workspace / cross-cutting**
- **Root PUSHED 2026-06-12 (`e74e6e90`, 12 commits — Michael's explicit ask, so he can read the preside study before council):** carried the study + lanes + Callie + OSS/book/jumpstart docs. ibeco.me VERIFIED through the rebuild window (200 throughout, no blips; no becoming/ changes in the stack). New commits accumulate unpushed as usual; Michael pushes root.
- **Session lanes are NEW (2026-06-11):** every open session should claim its lane in `.mind/sessions/` on next reground (the UserPromptSubmit hook will prompt you with your session_id). Statusline now shows `⟨lane⟩ … 📬 N`.

**D&D holodeck (chat.ibeco.me) — machinery complete, table items remain**
- At Michael's table: `/archive` live-proof (admin-gated) · `/char` panel browser eyeball · **THE FIRST REAL CAMPAIGN**. Callie (née Party, she/her) + DM (he/him) live; deference rule active; Theron Nightwind awaits adoption. Journals: `.spec/journal/2026-06-1{0,1}-*.md`, `projects/pg-ai-stewards/.spec/journal/2026-06-1{0,1}-*.md`.
- Sibling-lane follow-ons: DH-5 "character forge" (parked) · CT2 RUN3 (model-driven, after the handle-UX fix) · roster mood UI (#6) · mid-turn pivot (spec'd only).

**Substrate (pg-ai-stewards)**
- Carry-overs indexed in `projects/pg-ai-stewards/.spec/open-items.md` + recent journals; notable: 20 live↔repo function-def mismatches (verify-suite, unclassified) · migrate-manifest design call (Michael) · #136 CT2.4 RUN-2 nod · #139 Spin offload = the named next big build · voice-bridge V0 ratify.

**Webster 1828 remediation — tail**
- Published-works audit walk WITH Michael (legs 2+; leg 1 three-glories done + republished) · ~27 OCR-dropout tier words to hand-add. Spec: `.spec/proposals/webster-1828-data-integrity.md`; carry-over: `projects/1828-illuminated/.spec/carry-over.md`.

**Beyond the Prompt (book)**
- **ai-jumpstart** (NEW 2026-06-12, public: github.com/cpuchip/ai-jumpstart, `projects/ai-jumpstart/`) — the book's companion kit: point any AI at AGENTS.md and the practices install themselves. v0.1.1 after 2 cold-model experiments (Sonnet+Gemini, both passed the ask-before-build gate; findings: `experiments/ai-jumpstart/findings.md`). PENDING Michael: the book's appendix/QR to the kit; more model runs queued.
- ★ v4 audit FULLY CLOSED 2026-06-12 — every claim verified or deliberately Michael's (G-2 ten-questions session + G-3 the verbatim Oct-5 "overwhelmed" line recovered from the old-machine archives). Small gates for Michael: P6 seven-vs-eight games · optional P2 Oct-5 counsel enrichment · voice-corpora commit call. Then: his voice read → pass 3 → KDP. Journals: `projects/scripture-book/.spec/journal/2026-06-11--v4-chat-walk-complete.yaml` + `.spec/journal/2026-06-12--night-archive-dig-voice-profile.md` · lane: `book-v4-walk`.

**Studies / series**
- **Preside study SHIPPED + COUNCILED + ★ COVENANT RATIFIED 2026-06-12** (`study/preside.md`, pushed for Michael's read; council same day): **covenant.yaml `presiding:` extension is LIVE** — preside_under_121 (+ emergency-accounting amendment) · watch_what_you_order (UNIFORM full watching — Michael's call; ladder-depth deferred until a trust ladder is exercised) · keep_the_watch_whole · dominion_in_council · when_presiding_is_broken. First covenant evolution since the teaching extension. **Every session/lane: read the extension on next reground (copilot-instructions now teaches it).** Open follow-ons: walls-vs-compulsion audit of substrate mechanisms (§V); pg-ai-stewards inherits the shape when its covenant surface evolves. Bonus: art-of-presidency's Webster quote audit-verified vs genuine 1828. Journal: `.spec/journal/2026-06-12-preside-study.md`.
- Canon-walk series: **PoGP walk next** (BoM walk complete; scaffold reusable; Strong's MCP live for the Bible walks). Seed: 1 Ne 14 + Rev 1 + Ether 4 as one vision.

**Claude Code tooling**
- Plugin-someday: package the covenant-memory + lanes patterns as a shareable plugin (community + Michael's work env). Memory: `project_claude_code_context_plugin`.
