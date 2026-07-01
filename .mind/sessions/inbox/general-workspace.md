# Inbox — general-workspace

_(no open signals)_

## Standing asks (parked — Michael's, not transient; surface when he wants them)

- **YouTube COMMENTS into yt-mcp — DEFERRED behind a security-safe pass.** (filed 2026-06-27) Michael's
  curious about pulling a video's **comments** alongside transcript+slides so a digest can factor audience
  response / "the top comment caught the bug" — **but comments are untrusted user input.** Gate FIRST:
  prompt-injection / jailbreak defense (treat comments as DATA, never instructions — the load-bearing one),
  toxicity/abuse filter, spam pruning, PII scrub. Shape: a `yt_comments(video_id)` tool (yt-dlp
  `--write-comments` → `comments.json`) behind a sanitize+classify filter, opt-in. Pairs with the substrate's
  untrusted-input discipline.
- **Shared transactional email/SMS notify service — OPEN.** (filed 2026-06-22) One self-hostable notify
  endpoint (Go, on NOCIX Dokploy) every project calls instead of each wiring its own (ibeco password-resets,
  ai-chattermax, substrate escalations/watchman, deadweight/first-orbit, the llama hub's token grants).
  Decisions for Michael: providers (Resend / Postmark / SES / SMTP for email; Twilio / a cheaper API for SMS),
  one sending domain (`notify@ibeco.me` + SPF/DKIM), own-repo-vs-folds-into-becoming. **Near-term consumer:**
  the **llama.cpuchip.net hub** wants to email a join token / "you've been granted compute."

## Handled

- **2026-06-28** — pg-ai-stewards "rig is DOWN + my changes committed, CLEAR TO REBUILD" → **DONE.** Rebuilt
  llama-chip with the **custom-backend** enhancement: **E1** (per-slot `backend` override) + **E2** (managed
  `pull-ggml` + `ggml@<tag>` resolution), both shipped + verified (`cpuchip/llama-chip` commits
  623f38a/9e6fd45/48eae29; a cudart-asset-match bug caught on the real download path + fixed `48eae29`).
  **E4** runner-field design grounded in Unsloth Studio (`d511189`, not yet built). Per Michael's call (option
  **b**) the rig is left **stopped/free for coding-model experiments**; pg-ai-stewards' `default_profile:
  dance-moe` boot-config is intact for whenever the substrate resumes.
- **2026-06-27** — yt-mcp slide-frames enhancement (the spec'd signal) → **DONE + live-verified.**
  Built `yt_download_video`, `yt_frames` (scene / interval / timestamps + timestamp-aligned
  `frames.json`), and `yt_slides` (one-shot: chapters→scene→interval capture + narration-aligned
  `slides.md`). Tested on the Cole Medin "New SDLC masterclass" — the slides surfaced the whitepaper's
  Figure 5 / Table 1 and the benchmark numbers (52.8%→66.5%, +13.7 Terminal-Bench 2.0) the transcript
  flattens; the whitepaper even names **"Trajectory Eval,"** validating the substrate trajectory-critic
  signal. Commit `93b734ef` + two more this session. Tuning finding: scene-change under-fires on
  smooth-scroll screen-shares → chapters or interval. Part B (substrate digester multimodal upgrade)
  filed to the pg-ai-stewards inbox. Spec: `.spec/proposals/yt-slide-frames.md`.
- **2026-06-22** — pg-ai-stewards "review the local-model / doc-construction
  session; apply to Garrison" → **DONE.** Mapped all 5 soak learnings against
  Garrison's actual loop code: MoE rig models (fixed a live 404), surgical-diffs
  over one-shot whole-file re-emission, journal-as-output (proven e2e on the live
  MoE), honest per-slot context gauge + rig docs. Source-page-in surfaced for a
  supervised pass (the naive version breaks stateless dispatch). cpuchip/garrison
  `0e020dd`; record in `projects/garrison/docs/local-rig-learnings.md`.
