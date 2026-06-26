# Inbox — general-workspace

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

- **2026-06-22** — pg-ai-stewards "review the local-model / doc-construction
  session; apply to Garrison" → **DONE.** Mapped all 5 soak learnings against
  Garrison's actual loop code: MoE rig models (fixed a live 404), surgical-diffs
  over one-shot whole-file re-emission, journal-as-output (proven e2e on the live
  MoE), honest per-slot context gauge + rig docs. Source-page-in surfaced for a
  supervised pass (the naive version breaks stateless dispatch). cpuchip/garrison
  `0e020dd`; record in `projects/garrison/docs/local-rig-learnings.md`.
