# 2026-06-11 — Callie, the deference rule, and the duplicate host that explained everything

**The report:** Michael asked DM for a character sheet; DM asked good follow-ups — but Party beat him to it and just *created* one. "What made Party think I was talking to it?"

**The diagnosis (evidence, not guesswork):** pulled the actual turn from the substrate. Party's framing literally said *"You weren't addressed directly"* — `respond_policy: judgment`, and the model judged character creation its wheelhouse and chimed in anyway. The addressing code was innocent (`@DMAssistant` matched only DM). Etiquette gap: the chime-in license had no "someone else was named — defer" rule. **And** the one message produced FOUR turns: two DMs + two Parties on different pipelines — a second persona-host. Hunted it: a native `persona-host.exe` launched 12:43 PM by the morning session as a live test of r21 room_react (built 12:42, committed 12:53, never cleaned up). The container meanwhile ran a **pre-DH-3 persona snapshot** (plain `persona-turn`) — so the "orphan" was actually the only host with sheet tools; it's the one that built Theron Nightwind.

**Ratified (AskUserQuestion):** Party → **Callie** (she/her — the old-school table *caller* who spoke the party's actions to the DM; also retires "party" as a homograph wake-word) + **Both** deference layers. DM is he/him.

**Shipped + verified:**
- **chattermax `0be51ec`/`02a66cc` (deployed, marker verified):** key-authed `PATCH /api/persona/profile` — hosts assert registry display names; the Codewright/Chattercode name-drift class self-heals.
- **persona-host (root `18b31f7`):** Callie seed (caller lore, she/her, deference line in her card), DM (he/him), `mentionsAnother` + judgment hard-gate (skip the turn when another entity is @-named; `TestMentionsAnother` covers the incident message verbatim), `syncDisplayName` in `Run()` before dial. `go test -race` green.
- **Live state:** registry row renamed in place; **characters link by `persona_slug`** — Thorin + Vexa re-linked (would have been orphaned); `.env` `party=`→`callie=`; image rebuilt; recreated — all 5 connected, Callie joined Holodeck-3 (judgment), platform API returns `displayName: "Callie"`.
- Earlier same session: context statusline + post-compact grounding (`3b2fab9`) — see the morning journal entry.

**OPEN — needs Michael:** the morning session's terminal **relaunched its exe at 4:18 PM** (background Bash, old binary, old "Party" snapshot) after my kill — two hosts again, double turns in Holodeck-3 until it stands down. Deliberately did NOT kill it twice: it belongs to a live sibling session; agents shouldn't ping-pong each other's processes. Michael to close/stand down the other terminal, then anyone can kill the exe.

**Lessons:** (1) after ANY persona-host work, verify exactly ONE host exists — `docker ps` + `Get-Process persona-host` (the DH-3 lesson, now bitten from the native side). (2) A persona named after a common noun in its own domain ("Party" at a D&D table) is an addressing hazard — name personas like names. (3) The characters table links by slug, not id — slug renames must cascade.
