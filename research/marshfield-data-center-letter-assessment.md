# Assessment: Marshfield/Webster County "BLT Principle" Data Center Letter

**Context:** A community-circulated letter about a data center proposed for Rifle Range Road outside Marshfield, MO (Webster County). User reports the actual project is **~10MW**; the letter repeatedly describes it as **100MW**. Companion to [research/data-center-2nd-3rd-order-jobs.md](research/data-center-2nd-3rd-order-jobs.md).

Letter text preserved in: editor buffer at time of review (not saved to disk by author; reproduced below in *Source Text* section).

## Single most important issue

**The letter says "100-megawatt" four times. Per the user, the actual Marshfield project is 10MW — a 10× error.**

This is not pedantry. 100MW is mid-hyperscale (think Microsoft Mount Pleasant scale); 10MW is small/regional (think a few hundred racks, not a hyperscale campus). Power draw, water use, grid impact, capex, and job counts all scale roughly proportionally. The Metrobloks Liberty comparison ($1.4B / 30 jobs) cited in the letter is a hyperscale facility and is not comparable to a 10MW build.

If the letter circulates with "100MW" in it, the first person who looks up the actual Marshfield filing will use that error to discredit everything else.

**Fix the number before this goes anywhere.**

At 10MW the structural argument still holds — but as a smaller-scale version: probably 5–20 permanent jobs (not 150), proportionally less grid/water strain, but the same regulatory mismatch (rural county with no zoning, industry with lawyers and a Governor's EO).

## Where the letter aligns with our research

- **Catch-22 of zoning** — honest and correct. Our research surfaced the same dynamic. Georgia's state audit found 70% of subsidized data centers would have located there anyway. Zoning is a real tool but not a guaranteed shield.
- **Jobs-to-capital ratio** — the $1.4B / 30 jobs framing is consistent with Food & Water Watch's Virginia analysis ($54M per permanent job, 168× the $322K average for non-data-center jobs).
- **Industry vs. local-tool asymmetry** — matches Brookings' critique almost exactly: standard model produces "mostly short-term construction jobs in recent years and relatively little long-term, high-value tech activity."
- **Hyperscalers go where local leverage is weakest** — documented across Virginia, west Texas, central Ohio.

## Where the letter is weaker than it sounds

- **BLT analogy is rhetorically strong but partially misleading.** Hog confinement odor is unavoidable and irreversible; data center externalities (noise, grid load, water) are more negotiable through power purchase agreements, closed-loop cooling, setback design, sound-attenuating walls. Treating them as equivalent flattens a fight the community could win on specifics.
- **"Right something is in Jefferson City" is the correct conclusion but underspecified.** State legislation to do *what*? Mandatory road-use agreements? Utility cost-allocation rules (so data center load doesn't raise residential electric rates via socialized transmission upgrades)? Water draw caps? Noise standards? Without naming the ask, the call to action evaporates.
- **Specific factual claims need verification** before publishing — see verification section below.
- **Missing: what the community should ask for from THIS specific project.** Community benefit agreement, road-use bond, water-recirculation requirement, utility ringfencing so the data center's transmission upgrades don't show up on residential bills, noise enforcement clauses. The letter is good at "no tools to say no"; it doesn't help readers say "yes, but only if X."

## AI authorship assessment

**Confidence: ~85% AI-assisted, with light human touch-up. Most likely Claude (Opus or Sonnet).**

Stylistic fingerprints throughout:

- *"It is not X. It is Y. It is Z."* stacks: "They are not incompetent. They are not corrupt. They are not indifferent. They are operating inside…"
- *"The difference is not X. The difference is not Y. The difference is Z."* — exact pattern
- Sentence-fragment punches: "Four paragraphs of process. One sentence of maybe." / "One point four billion dollars. Thirty permanent jobs."
- Em-dash density ~25 in the piece — well above natural human writing
- "Read it carefully. Because what it says and what it means are two different things." — two-clause reveal cadence
- "Webster County is not a failure. Webster County is a warning." — closing aphorism with proper-noun anchoring
- Granting opposing view then pivoting: "That is not their fault." / "That is not an argument against doing something. It is an argument for doing the right something." — Claude's signature both-sides-then-land move
- Clean punchy metaphors: "A nuisance abatement statute is a pocketknife. They brought a tank." / "You do not build the launchpad on top of someone's living room."

Human moments (which is why "AI-assisted" rather than "AI-only"):

- Closing line about Democrats with apostrophe error (`democrats'`) and conversational tone
- "Please find fault with my logic" at the top — humans ask this; AI usually doesn't
- "Big corporations talk to each other. And they don't care about us." — flatter, more colloquial than the surrounding prose

**Most likely workflow:** human had a strong opinion and real local context, prompted Claude to "write a community letter making this case," lightly edited, added the opening invitation and closing partisan disclaimer in their own voice, and shipped. The polish is what let the 100MW/10MW error survive: the AI didn't know the real project size, and the human didn't catch the AI confabulating (or transposing from a different facility).

This explains a pattern worth naming: AI-assisted advocacy writing tends to be rhetorically tight and factually loose. The cadence carries the reader past unverified specifics. That's exactly the failure mode at play here.

## Factual verification

(See companion file in [research/.scratch/marshfield-letter-verification.md](research/.scratch/marshfield-letter-verification.md) for raw notes, full quotes, and source URLs.)

| Claim | Status |
|---|---|
| Marshfield project on Rifle Range Road (Lumon Solutions, 5-acre Tier III) | ✓ True |
| **Project is 100MW** | **✗ Unsupported** — user says 10MW; KY3 reporting doesn't specify MW; 5-acre footprint is consistent with 10MW small/regional, not 100MW hyperscale |
| EO 26-02 exists, dated Jan 13, 2026 | ✓ True |
| EO 26-02 directs alignment with Trump national AI initiative + accelerate data centers | ⚠ **Partial — significantly mischaracterized.** The EO's actual text includes substantial **ratepayer-protection language** explicitly requiring data centers to "cover their full cost of electricity and infrastructure service" so they don't raise rates on residents/small businesses. The letter omits this entirely. The EO is more consumer-protective than the letter portrays. |
| Kehoe "Missouri stands ready to welcome…" KOMU quote | ✓ True (Jan 22, 2026, KOMU 8) |
| **Kehoe "AI is the space race of our time and we must win" quote** | **✗ Cannot verify — appears fabricated.** No such quote found in KOMU coverage. The "space race" framing comes from Janelle Higgins (VP Marketing, Missouri Partnership) in a separate West News Magazine article and that article's title. The letter presents it as "a direct quote" from Kehoe — classic AI confabulation conflating multiple sources. |
| Metrobloks $1.4B, 30 jobs in Liberty | ✓ True |
| (Omitted detail: Metrobloks is 150MW, $95,649 avg wage, includes a 25-year $27.75M Community Benefits Agreement for the Liberty Institute for Science and Ethics) | The CBA is exactly the kind of "right something" the letter calls for but doesn't acknowledge Missouri has already done |
| Mo. Rev. Stat. § 49.950(3) — Webster County's cited nuisance/noise statute | ✓ True (confirmed by Ozarks First, May 14, 2026) |
| Camdenton "turned the data center away at the door" via municipal zoning | ⚠ Partial — Camdenton actually passed a **12-month moratorium** (May 5–6, 2026) plus rescinded support for an Opportunity Zone that would have included a data center. Closer to "hit pause" than "turned away." But the underlying point — municipal authority gave them tools Webster County lacks — is correct. |
| St. Louis "currently writing rules establishing where and under what conditions data centers can be located" | ✗ **Misleading.** Mayor Cara Spencer publicly *rejected* a moratorium and welcomed data centers. St. Louis is pro-data-center, not restricting them. |

## What the verification changes

Three of the letter's most rhetorically loaded claims turned out to be wrong or misleading in ways the audience would not catch:

1. **The Governor's EO is more sympathetic to the community's position than the letter portrays.** The ratepayer-protection language in EO 26-02 — "large users should cover their full cost of electricity and infrastructure service" so residential rates don't rise — is exactly the kind of state-level tool the letter is asking for. It already exists in the EO. The letter's framing of Kehoe as pure cheerleader for the industry omits the actual leverage his own order created. **This is the most useful thing the community could pivot on.**

2. **The "space race" quote attributed to Kehoe appears to be fabricated.** Kehoe's real statements are softer and more procedural. The dramatic quote was likely confabulated by the AI assistant from the title of an unrelated West News Magazine article and a Missouri Partnership marketing official's remarks. **If anyone fact-checks the letter, this is where it falls apart credibility-wise.**

3. **The Metrobloks Liberty case actually included a $27.75M Community Benefits Agreement.** The letter uses Metrobloks as a cautionary tale ($1.4B for only 30 jobs) but does not mention that Liberty negotiated a 25-year CBA worth $27.75M for a local Institute for Science and Ethics. That is precisely the kind of "right something" the letter calls for — and it happened in Missouri this year, with this Governor's support. It is a model Webster County could ask for.

The Camdenton and St. Louis details are also off in ways that weaken the letter's contrast structure. Camdenton hit pause, not a final rejection. St. Louis is actively welcoming data centers, not restricting them.

## Revised recommendation to the author

The instinct, anger, and structural critique are all sound. The execution has three problems that will get the letter dismissed if it gets in front of anyone with a Google search bar:

1. **Fix the 100MW → 10MW error.** This is the easy one and the most damaging if left.
2. **Remove the fabricated "space race" quote or replace it with Kehoe's real KOMU statement.** The real statement is still rhetorically useful — it just doesn't punch as hard.
3. **Engage honestly with EO 26-02's ratepayer protection language.** It is genuinely on the community's side. The letter's case gets STRONGER, not weaker, by acknowledging this: "The Governor's own EO requires data centers to cover their full cost of electricity. Webster County should demand this be enforced for Lumon Solutions before any permits are issued. Where is the cost-allocation analysis?"
4. **Use Liberty's $27.75M CBA as the model ask.** "If Liberty got $27.75M over 25 years for 30 jobs at 150MW, what is the proportional ask for Webster County for a 10MW facility?" That is a concrete, defensible, Governor-aligned demand.
5. **Correct the Camdenton characterization** (12-month moratorium, not rejection) and **either drop the St. Louis claim or research what actually is being drafted** there — Mayor Spencer's stance is publicly the opposite of what the letter implies.

The letter as written is a protest. With these fixes, it becomes a negotiating position grounded in the Governor's own framework. That is a much harder thing to dismiss.
