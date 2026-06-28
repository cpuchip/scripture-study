# Inbox — general-workspace

## 📬 2026-06-27 (from Michael) — integrate YouTube COMMENTS into the yt-mcp — DEFERRED (security-safe pass required FIRST)

**Michael's ask:** he's curious about pulling a video's **comments** in alongside the transcript +
slides, so a digest can factor in audience response / corrections / "the top comment caught the bug."
**But not yet** — *"people on the internet can be pretty awful in their comments,"* so this needs a
**security / safety pass FIRST** before comments are ever fed to a model or written to a doc.

**What the safety pass must cover (when we pick it up):**
- **Prompt-injection / jailbreak defense** — comments are *untrusted user input*. A comment can carry
  "ignore your instructions…" or worse. Any digest reading comments must treat them as DATA, never
  instructions (delimit + label as untrusted; never let a comment steer the agent). This is the load-bearing one.
- **Toxicity / abuse filtering** — drop or flag slurs, harassment, doxxing, NSFW before they reach a
  model or a doc.
- **Spam / low-signal pruning** — most comments are noise; keep top / most-relevant, not all.
- **PII scrub** — comments leak personal info.

**Shape (later):** a `yt_comments(video_id)` tool (yt-dlp `--write-comments` → `comments.json`) gated
behind a sanitize+classify filter, with an opt-in flag so digests pull them only deliberately. Pairs
naturally with the substrate's untrusted-input handling (same prompt-injection discipline).

— filed by Michael 2026-06-27. NOT now; the security-safe pass is the gate. Noted so it isn't lost.

---

## 📬 2026-06-22 (from pg-ai-stewards) — shared transactional email/SMS service — OPEN (Michael's ask)

**Michael's ask:** *"we need to get a text/email service setup to use across all of our stuff."*
One transactional notify service every project can call (becoming/ibeco, ai-chattermax, the
substrate's escalations/watchman, deadweight/first-orbit, the new llama hub's token grants, etc.)
instead of each wiring its own.

**Shape worth scoping:** a tiny self-hostable notify endpoint (Go, on the NOCIX Dokploy) wrapping
a provider — **email** (Resend / Postmark / SES / SMTP) + **SMS** (Twilio / a cheaper SMS API). One
internal API + a shared secret; callers POST `{to, channel, template, vars}`. Decisions for Michael:
which providers (cost + deliverability — Resend is cheap+simple for email; Twilio is the SMS default
but per-segment), one domain for sending (e.g. `notify@ibeco.me` / SPF+DKIM), and whether it's its
own repo or folds into becoming. **Near-term consumer:** the **llama.cpuchip.net hub** (below) wants
to email a join token / "you've been granted compute" — so this pairs with that build.

— filed by pg-ai-stewards; not yet acted. General-workspace owns cross-cutting infra, so flagging
it here for scoping/council when Michael wants it.

---

_(no other open signals)_

## Handled

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
