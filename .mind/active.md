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
- **Root UNPUSHED** (Michael pushes root): the stack includes the D&D-holodeck commits (~15, sibling session), Callie + deference + name-sync (`18b31f7`), context tooling (`3b2fab9`), and the session-lanes system (this change). A root push also rebuilds ibeco.me prod — verify after.
- **Session lanes are NEW (2026-06-11):** every open session should claim its lane in `.mind/sessions/` on next reground (the UserPromptSubmit hook will prompt you with your session_id). Statusline now shows `⟨lane⟩ … 📬 N`.

**D&D holodeck (chat.ibeco.me) — machinery complete, table items remain**
- At Michael's table: `/archive` live-proof (admin-gated) · `/char` panel browser eyeball · **THE FIRST REAL CAMPAIGN**. Callie (née Party, she/her) + DM (he/him) live; deference rule active; Theron Nightwind awaits adoption. Journals: `.spec/journal/2026-06-1{0,1}-*.md`, `projects/pg-ai-stewards/.spec/journal/2026-06-1{0,1}-*.md`.
- Sibling-lane follow-ons: DH-5 "character forge" (parked) · CT2 RUN3 (model-driven, after the handle-UX fix) · roster mood UI (#6) · mid-turn pivot (spec'd only).

**Substrate (pg-ai-stewards)**
- Carry-overs indexed in `projects/pg-ai-stewards/.spec/open-items.md` + recent journals; notable: 20 live↔repo function-def mismatches (verify-suite, unclassified) · migrate-manifest design call (Michael) · #136 CT2.4 RUN-2 nod · #139 Spin offload = the named next big build · voice-bridge V0 ratify.

**Webster 1828 remediation — tail**
- Published-works audit walk WITH Michael (legs 2+; leg 1 three-glories done + republished) · ~27 OCR-dropout tier words to hand-add. Spec: `.spec/proposals/webster-1828-data-integrity.md`; carry-over: `projects/1828-illuminated/.spec/carry-over.md`.

**Beyond the Prompt (book)**
- v4 honesty+voicing CHAT WALK in progress with Michael (batches 1–4 + identity items applied; remaining: 4 ground-truth figures, council-verb + Ch 4 title-tense questions, F-37 EPUB list-rendering fix, final rebuild + QR-collision check). State: `projects/scripture-book/.draft/20260609-v4-walk-findings.md` · lane: `book-v4-walk`.

**Studies / series**
- Canon-walk series: **PoGP walk next** (BoM walk complete; scaffold reusable; Strong's MCP live for the Bible walks). Seed: 1 Ne 14 + Rev 1 + Ether 4 as one vision.

**Claude Code tooling**
- Plugin-someday: package the covenant-memory + lanes patterns as a shareable plugin (community + Michael's work env). Memory: `project_claude_code_context_plugin`.
