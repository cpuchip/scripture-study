# Active Context — the in-flight board

> **2026-06-11/12 — ★ pg-ai-stewards is PUBLIC: spec ratified, Apache-2.0 FINAL, "Anatomy of a Turn" SHIPPED.** `github.com/cpuchip/pg-ai-stewards` → `projects/pg-ai-stewards-oss/` (private substrate stays at `projects/pg-ai-stewards/`). Ratified: **v0.1 = core + persona-host** · clean-room fresh history · cutover after virgin boot · **Apache-2.0 (`3c43d4e`)** · side-by-side docker dev (`stewards-oss-*`, 55434/8081/8091, own persona keys). **Deliverable #1 landed `0e8c3c9` + order-research update:** `docs/anatomy-of-a-turn.md` — source-verified turn pipeline, 6-beat Remotion storyboard, serial-position section (don't invert; covenant first AND last). P4 = playground machine; P5 = AI-office vision. Next: P1 extraction (task #151). Lane: `.mind/sessions/pg-ai-stewards.md`.

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
- **★ PR.1 SHIPPED + live-verified 2026-06-12:** the presiding covenant now rides every dispatch — `covenants.extensions` jsonb catch-all (anti-silent-drop), parser pass-through, presiding render + `The Watch (echo)` in compose_system_prompt (covenant first AND last, per serial-position research). Live smoke `600f6673` ACK'd with presiding terms in the dispatched payload. +~820 tok/dispatch (watch on gate volume). Journal: `projects/pg-ai-stewards/.spec/journal/2026-06-12-pr1-covenant-extensions.md`. NEW gotchas: MSYS path-mangle ate a `psql -f` (use `MSYS_NO_PATHCONV=1`); ledger naming wart (bridge keys suffix-less vs manual `.sql` rows — feeds the migrate-manifest call).
- Carry-overs indexed in `projects/pg-ai-stewards/.spec/open-items.md` + recent journals; notable: 20 live↔repo function-def mismatches (verify-suite, unclassified) · migrate-manifest design call (Michael) · walls-vs-compulsion audit (preside §V) · trailing-reminder near true context end (proposal-first; provider quirks) · #136 CT2.4 RUN-2 nod · #139 Spin offload = the named next big build · voice-bridge V0 ratify.

**Webster 1828 remediation — tail**
- Published-works audit walk WITH Michael (legs 2+; leg 1 three-glories done + republished) · ~27 OCR-dropout tier words to hand-add. Spec: `.spec/proposals/webster-1828-data-integrity.md`; carry-over: `projects/1828-illuminated/.spec/carry-over.md`.

**Beyond the Prompt (book)**
- **ai-jumpstart** (NEW 2026-06-12, public: github.com/cpuchip/ai-jumpstart, `projects/ai-jumpstart/`) — the book's companion kit: point any AI at AGENTS.md and the practices install themselves. v0.1.1 after 2 cold-model experiments (Sonnet+Gemini, both passed the ask-before-build gate; findings: `experiments/ai-jumpstart/findings.md`). PENDING Michael: the book's appendix/QR to the kit; more model runs queued.
- ★ v4 audit FULLY CLOSED 2026-06-12 — every claim verified or deliberately Michael's (G-2 ten-questions session + G-3 the verbatim Oct-5 "overwhelmed" line recovered from the old-machine archives). Small gates for Michael: P6 seven-vs-eight games · optional P2 Oct-5 counsel enrichment · voice-corpora commit call. Then: his voice read → pass 3 → KDP. Journals: `projects/scripture-book/.spec/journal/2026-06-11--v4-chat-walk-complete.yaml` + `.spec/journal/2026-06-12--night-archive-dig-voice-profile.md` · lane: `book-v4-walk`.

**Studies / series**
- **Preside study SHIPPED + COUNCILED + ★ COVENANT RATIFIED 2026-06-12** (`study/preside.md`, pushed for Michael's read; council same day): **covenant.yaml `presiding:` extension is LIVE** — preside_under_121 (+ emergency-accounting amendment) · watch_what_you_order (UNIFORM full watching — Michael's call; ladder-depth deferred until a trust ladder is exercised) · keep_the_watch_whole · dominion_in_council · when_presiding_is_broken. First covenant evolution since the teaching extension. **Every session/lane: read the extension on next reground (copilot-instructions now teaches it).** Open follow-ons: walls-vs-compulsion audit of substrate mechanisms (§V). (The substrate side LANDED 2026-06-12 — PR.1 carries the presiding extension into every dispatch; see Substrate section.) Bonus: art-of-presidency's Webster quote audit-verified vs genuine 1828. Journal: `.spec/journal/2026-06-12-preside-study.md`.
- Canon-walk series: **PoGP walk next** (BoM walk complete; scaffold reusable; Strong's MCP live for the Bible walks). Seed: 1 Ne 14 + Rev 1 + Ether 4 as one vision.

**Claude Code tooling**
- Plugin-someday: package the covenant-memory + lanes patterns as a shareable plugin (community + Michael's work env). Memory: `project_claude_code_context_plugin`.
