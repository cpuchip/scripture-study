# Research: Custom Desk Project

**Entry ID:** custom-desk-project
**Category:** projects
**Captured:** 2026-04-23
**Tags:** budget, custom, furniture

---

## What This Is About

A custom desk is being made locally — 30" × 72" sit-stand, real wood with a bow cut for monitor proximity. The desk has been commissioned but the deposit has not yet been paid, and action is needed to keep the maker's schedule.

---

## What Already Exists

[WORKSPACE] The March 22, 2026 Sabbath record (`/.spec/sabbath/2026-03-22-sabbath.md`) documents this desk as a completed milestone for that cycle:

> "Custom desk: 30" × 72" sit-stand, real wood, bow cut for monitor proximity. Commissioned locally."

It appears again in the cycle declaration:

> "The desk is real wood, custom-fitted, and serves the work."

And in carry-forward:

> "Desktop swap: new replaces old, old decommissioned (not repurposed)"

[WORKSPACE] No other workspace files (`.spec/`, `becoming/`, `journal/`, `projects/`) contain references to the desk, a deposit, furniture, or the maker. There is no existing spec, contract notes, or financial tracking for this purchase.

[SYNTHESIS] Timeline gap: The desk was "commissioned locally" as of March 22, 2026 — over a month before this capture (April 23, 2026). The deposit has still not been paid. This is a follow-up action that has been delayed.

---

## External Context

[WEB] Web search attempts for custom furniture commissioning guides were largely paywalled or unavailable. The following general knowledge applies:

- **Typical deposit structure:** Custom furniture makers commonly request 30–50% upfront to cover materials. On a $760 desk, that's roughly $228–$380.
- **Deposit purpose:** Protects the maker's time and locks in material costs. Without it, the maker has no obligation to hold your slot or begin sourcing wood.
- **What to confirm before paying:**
  - Final dimensions and species/finish confirmed in writing (or at least text/email)
  - Expected timeline from deposit to delivery
  - What happens if there are material overruns or design changes
  - Delivery/pickup logistics
- **Standing desk specifics:** A sit-stand desk at 30" × 72" with real wood and a custom bow cut is a non-trivial commission. The mechanism (manual crank, electric lift, or static with adjustable feet) should be confirmed if not already settled.

[WEB] For context on value: A 30" × 72" sit-stand desk with solid wood top from commercial makers (e.g., Uplift, Flexispot) typically runs $900–$1,500+ without customization. $760 for a locally commissioned custom piece with a bow cut is competitive pricing, especially for real wood.

---

## Open Questions

1. **What are the exact deposit terms?** How much is due upfront — and is that already agreed, or does it need to be negotiated?
2. **Is there a written agreement?** Email, text, or contract documenting specs (dimensions, wood species, finish, mechanism type, delivery date)?
3. **What's the maker's timeline?** How many weeks from deposit to delivery? Is there a deadline risk if deposit is further delayed?
4. **What is the sit-stand mechanism?** Electric lift, hand crank, or fixed-height with adjustable feet? This affects both cost and long-term use.
5. **What wood species?** This was noted as "real wood" but species (walnut, maple, oak, etc.) affects look, durability, and price.
6. **Delivery or pickup?** How does the finished desk get from maker to home?
7. **Old desk plan?** The Sabbath record says "new replaces old, old decommissioned" — is the old desk gone, or still in place?

---

## Raw Sources

- [WORKSPACE] `/.spec/sabbath/2026-03-22-sabbath.md` — Desk commissioned, specs noted, carry-forward action recorded
- [WEB] Popular Mechanics standing desk roundup (context for market pricing): https://www.popularmechanics.com/home/interior-projects/g39741283/best-standing-desks/
- [WEB] Most custom furniture commissioning guides paywalled (Fine Woodworking, Apartment Therapy, Craftsy) — no usable external sources found

---

## Plan

**Scope:** 1 session (mostly communication + a payment, not dev work)
**Complexity:** Low — this is a logistics/follow-through plan, not a build plan. The hard part is the delay, not the difficulty.

### What to Build

This is a personal project, not a software project. "Build" here means produce the artifacts and actions that get the desk made:

1. A short written spec confirmed with the maker (text/email is fine) — dimensions, wood species, finish, sit-stand mechanism, timeline, deposit amount, delivery method.
2. A paid deposit, with the receipt/confirmation captured.
3. A tracking note in the workspace so this doesn't drift again — likely a single file at `projects/custom-desk.md` (new) recording specs, deposit paid, expected delivery, and old-desk decommission plan.
4. A calendar/reminder anchor for expected delivery week so the next follow-up (delivery logistics, decommissioning the old desk) isn't dropped.

### Phases

1. **Phase 1: Confirm specs in writing** (~30 min)
   - Deliverable: A text/email thread with the maker that captures dimensions (30" × 72"), wood species, finish, sit-stand mechanism, total price ($760), deposit amount, expected timeline, and delivery/pickup plan.
   - Files: none yet — just the conversation.

2. **Phase 2: Pay the deposit** (~15 min)
   - Deliverable: Deposit paid via the maker's preferred method. Receipt or confirmation message saved.
   - Files: optionally a screenshot/PDF in `projects/custom-desk/` if you want a paper trail.

3. **Phase 3: Record the project in the workspace** (~20 min)
   - Deliverable: `projects/custom-desk.md` containing: specs, price breakdown, deposit paid date + amount, balance owed, expected delivery window, maker contact, old-desk decommission plan.
   - Files: `projects/custom-desk.md` (new).

4. **Phase 4: Set the next follow-up** (~10 min)
   - Deliverable: A reminder anchored to the expected delivery window — either a calendar entry, a brain entry with a due date, or a line in `.mind/active.md` under in-flight items.
   - Files: possibly an update to `.mind/active.md` or a new brain entry.

### Scenarios

- When the maker is contacted, then a written confirmation of all specs (dimensions, wood, finish, mechanism, price, timeline) exists in text or email.
- When the deposit is paid, then a dated receipt or confirmation message is captured.
- When `projects/custom-desk.md` is opened a month from now, then everything needed to answer "what's the status of the desk?" is present without having to ask the maker again.
- When the expected delivery week arrives, then a reminder fires (calendar, brain entry, or active.md) so delivery logistics aren't missed.
- When the new desk arrives, then there is a documented plan for the old desk (decommissioned per the Sabbath record — sold, donated, or trashed; not repurposed).

### Decisions Needed

1. **Deposit amount** — Is the maker expecting a specific percentage (30%? 50%? full?), or is this open? Trade-off: paying more upfront secures materials and slot but increases exposure if anything goes sideways with the maker. Resolve by asking directly: "How much would you like for the deposit?"

2. **Sit-stand mechanism** — Electric lift vs hand crank vs fixed-height with adjustable feet. Trade-offs:
   - *Electric:* easiest daily use, most expensive, motor is a long-term failure point.
   - *Hand crank:* cheaper, fully mechanical, friction over time but fixable.
   - *Fixed with feet:* cheapest, requires lifting the whole desk to adjust — fine if you'll only sit-stand occasionally, painful if daily.
   If this isn't already settled with the maker, decide before paying the deposit.

3. **Wood species** — "Real wood" was noted but not specified. Walnut (dark, soft-ish, premium look), maple (light, hard, neutral), oak (mid-tone, hard, classic), cherry (warm, ages dark). Trade-off is mostly aesthetic + slight durability/price differences. Ask the maker what's in stock at the $760 price point.

4. **Where the project record lives** — `projects/custom-desk.md` (workspace) vs a brain entry (becoming pipeline) vs both. Trade-off: workspace file is durable and easy to find; brain entry surfaces in pipeline reviews. Recommend workspace file as primary, with a brain entry only if you want pipeline tracking.

5. **Old desk disposition** — Sabbath record said "decommissioned, not repurposed." Decide now: sell (Marketplace/OfferUp), donate (DI, Habitat ReStore), or trash. Affects timing — selling takes weeks, donating takes a trip, trashing takes a curb.

### Risks

- **Further delay** — A month has already passed since commission. Makers' schedules shift; if the slot is lost, timeline pushes out further. Mitigation: pay the deposit this week.
- **Spec drift** — Without written confirmation, "30 × 72 with a bow cut in real wood" is what *you* remember. The maker may remember something slightly different. Mitigation: text the specs back to him and ask him to confirm before paying.
- **Scope creep at the maker's end** — Custom work invites "while we're at it" additions that inflate cost. Mitigation: lock the spec at $760 in writing; any additions get a separate quote.
- **Old desk lingers** — If the old desk isn't actively decommissioned before the new one arrives, both end up in the room and the "decommission" never happens. Mitigation: pick a disposition path now (Decision #5) and put it on the calendar.
- **No paper trail** — If anything goes wrong (delivery damage, dispute over specs), text messages are the only record. Mitigation: keep the thread; don't delete it.

### Dependencies

- The maker's contact info and availability (assumed in hand — you commissioned this in March).
- Funds available for the deposit (~$228–$380 if 30–50%, possibly different if the maker has his own structure).
- A decision on the sit-stand mechanism and wood species *before* the deposit, if those weren't already locked at commission time.
- Nothing in the workspace blocks this — `projects/` exists as a directory and can hold the new file directly.

### Who Benefits? (Consecration Check)

You. Directly. This is a tool-of-trade purchase — a desk that serves the work you do (study, dev, writing). The Sabbath record already framed it: "real wood, custom-fitted, and serves the work." Naming it honestly: this is a personal upgrade, and that's a legitimate use of resources. Not every project has to serve someone else; some serve the steward so the steward can keep stewarding.

### How Does This Integrate? (Zion Check)

Connects to the March 22 Sabbath cycle, which already declared the desk as a milestone. Following through on the deposit closes a loop that's been open for a month — completing the "carry-forward" item from that cycle. Adding `projects/custom-desk.md` extends an existing pattern (the `projects/` directory) rather than introducing a new convention. The old-desk decommission also closes a Sabbath-declared carry-forward.
