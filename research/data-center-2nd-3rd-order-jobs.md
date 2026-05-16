# Data Center Job Creation — Beyond the 15 On-Site Hires

**Binding question:** When a data center goes in and only employs ~15 local people, what 2nd-order and 3rd-order jobs does it create — including the work made possible by the servers it hosts?

**Source horizon:** Three Exa neural searches across industry-funded impact studies (PwC/Data Center Coalition, Oxford Economics/Google, IMPLAN-based state studies), independent analyses (Brookings, Tailwind Economics, data center trade press), and macroeconomic literature on AI-enabled GDP (arXiv 2601.11196). Sixteen sources reviewed; the strongest are cited inline. Industry-funded sources are noted as such — they tend to publish the high end of multiplier estimates.

## TL;DR

A data center with 15 permanent on-site jobs is the visible tip of four employment tiers:

1. **Direct (on-site):** the 15 — operations technicians, facility engineers, security, logistics.
2. **Indirect (supply chain of the building):** construction trades, equipment manufacturers, utilities, fiber/telecom, professional services. State models put this at roughly **1.6× to 4.5× the direct headcount**, depending on whether construction is included.
3. **Induced (household spending of T1+T2 workers):** restaurants, retail, housing, schools, local services. Roughly **comparable in size to the indirect tier**.
4. **Enabled-by-compute (the servers' downstream):** SaaS companies, AI labs, streaming platforms, enterprises using AI/cloud as production input. Largest tier by far in dollar terms — but **almost none of it lands in the host community**.

The industry-quoted multipliers (5.9 to 7.5 jobs per direct job, nationally) refer to tiers 2 and 3 only. Tier 4 is where most of the value sits and is precisely what the local town does not capture.

## Tier 1 — Direct On-Site Jobs

These are the headline "15 jobs" — although typical large facilities run higher:

- **Typical large data center:** 30–160 permanent staff. The US Chamber of Commerce cites 157 permanent for a typical large facility, against 1,688 construction workers at peak ([Tailwind Economics, Feb 2026](https://tailwindeconomics.com/article/hidden-cost-of-ai-data-centers)).
- **100 MW hyperscale build:** ~150 permanent operations after ~1,500 peak construction ([DC Geeks, Apr 2026](https://dcgeeks.com/data-center-job-market-outlook/)).
- **Microsoft typical data center:** ~50 full-time staff. The Mount Pleasant, WI campus is the larger end: 3,000 construction → 500 operations ([Wisconsin Watch via ibmadison, Mar 2026](https://www.ibmadison.com/business-report/here-s-what-the-data-center-boom-means-for-wisconsin-s-workforce/article_9e5c4a3e-b6e8-4de5-b1de-a3660701335e.html)).
- **Large AI campus:** 300–500 across the full footprint with multiple buildings (DC Geeks).

Roles: data center technician (DCT), critical facility engineer (CFE), critical facility manager (CFM), network operations engineer, security operations, site operations manager, facilities engineers, logistics/inventory staff. Wages are above-average: Ohio reports an industry average of ~$76,000 ([Ohio SRC Interim Report, Jun 2025](https://www.gongwer-oh.com/public/Interim_Report-DC-_2025-SRC-v4.pdf)).

## Tier 2 — Indirect (Supply Chain of the Building)

### Construction phase (temporary, but repeats with each new phase)

Construction is where most of the visible job count lives — and it is dominated by trades:

- Electricians (BLS projects +11% demand 2023→2033, with data center construction as a primary driver)
- Plumbers, pipefitters, HVAC technicians
- Structural steel and iron workers, carpenters, concrete workers, earth drillers
- Commissioning specialists, controls engineers, civil engineers, project managers

US data center construction employment now exceeds **250,000 workers**, concentrated in Virginia, Texas, Arizona, and Ohio (DC Geeks). Construction headcount typically runs ~10× operations during build phases ("construction time, you usually have a lot more jobs — maybe 10 times in magnitude more so than operations" — Liang, in Wisconsin Watch).

### Equipment manufacturing (ongoing)

The servers, networking, power, and cooling gear all need to be built somewhere:

- Server, GPU, and networking hardware manufacturing
- Power equipment: transformers, generators, UPS systems, switchgear
- Cooling equipment: chillers, CRAC/CRAH units, increasingly liquid-cooling systems
- Fiber and cable manufacturing
- Semiconductor foundries upstream — a new 2nm fab runs ~$20–25B per facility ([Lazard Asset Management AI Value Chain](https://www.lazardassetmanagement.com/us/en_us/research-insights/investment-insights/investment-research/beyond-data-center-headlines-ai-value-chain-opportunities))
- DRAM/HBM memory production

The Wisconsin example concretizes this: three Wisconsin manufacturers alone have sold over $1 billion in equipment (motors, generators, cooling systems) to data centers (Wisconsin Watch).

### Operations supply chain (ongoing)

- Utilities — grid operators expanding generation and transmission for data center load. The Kewaunee Power Station nuclear rebuild in Wisconsin is being planned specifically to serve AI/data center demand (Wisconsin Watch).
- Water utility expansion for cooling
- Fiber and telecom carriers
- Spare parts and logistics
- Cybersecurity services
- Engineering consultancies and commissioning specialists
- Industrial gas suppliers (cooling, fire suppression)

### Sectoral breakdown

PwC's Data Center Coalition study reports the distribution of indirect + induced employment by sector ([PwC Impact Study, Sep 2023](https://static1.squarespace.com/static/63a4849eab1c756a1d3e97b1/t/65037be19e1dbf4493d54c6e/1694727143662/DCC-PwC+Impact+Study.pdf)):

- Services (professional, admin, food service, consulting, healthcare, hospitality): >50%
- Wholesale and retail trade: >10%
- Transportation and warehousing: ~9%
- Finance, insurance, real estate, rental and leasing: ~9%
- Manufacturing (servers, fiber, electrical equipment): ~5%
- Information sector (excl. data centers): ~3%
- Construction (new structures, fiber install): ~2%

The dominant share going to services is the giveaway that the model is *mostly capturing induced consumer spending*, not pure supply chain. Real Tier 2 manufacturing/construction is a smaller slice than the multiplier headline suggests.

## Tier 3 — Induced (Household Spending of Tier 1 + Tier 2 Workers)

When data center staff and supply-chain workers spend their paychecks locally, that supports another tranche of jobs: restaurants, retail, grocery, dining, entertainment, recreation, schools, housing markets, healthcare ([Arizona Impact Study, 2021](https://static1.squarespace.com/static/63a4849eab1c756a1d3e97b1/t/663044d8d160b943d48f30bf/1714439385121/The+Impact+of+Data+Centers+on+the+Arizona+Economy.pdf)).

### What the multipliers actually say

| Source | Multiplier | What it measures |
|---|---|---|
| Ohio SRC (state, composite incl. construction) | **2.58** | Each direct job → 1.58 additional Ohio jobs |
| Ohio SRC (state, operations only) | **5.5** | Each on-site operations job → 4.5 additional Ohio jobs |
| PwC national, operations only | **7.5** | Industry-funded national benchmark |
| Arizona (state, excl. construction) | **6.5** | "For every job inside a data center, there are 6.5 jobs created in the Arizona economy" |
| Oxford Economics for Google (national) | **5.9** | Each direct Google data center job supports 4.9 additional US jobs |
| Data Center Coalition (2025) | **>6** | "each direct job…supports more than six jobs elsewhere in the US economy" |

**Concrete state example.** Ohio in 2024 (Ohio SRC Interim Report):

- 36,857 direct (17,300 operations + 19,400 construction)
- 30,357 indirect (supply chain)
- 28,003 induced (household spending)
- **Total: 95,217**
- Output: $26.4B ($12.6B direct, $7.2B indirect, $6.6B induced)

The state honestly notes that the construction-heavy composite multiplier drags downward because construction "is leak-prone (out-of-state steel, turbines, specialty crews)" — meaning the dollars buying construction inputs *do not stay in Ohio* and therefore do not generate Ohio jobs.

## Tier 4 — Enabled-by-Compute (the "Based on the Servers It Hosts" Angle)

This is the tier most economic impact studies miss or understate, and the one the user's question explicitly invokes. The data center is not just a building with a supply chain — it is a production facility whose output is compute. Every job at every company that buys that compute is, in a real sense, enabled by it.

### Direct tenants and customers

- Every SaaS company renting AWS, Azure, GCP, Oracle Cloud capacity
- AI labs training and serving models (OpenAI, Anthropic, xAI, etc.)
- Enterprises that migrated from on-prem to cloud
- Streaming services (Netflix, YouTube, Twitch)
- Financial services using cloud for trading, risk, fraud detection
- Gaming: multiplayer infrastructure, cloud gaming
- Healthcare AI imaging, logistics route optimization, marketing/creative agencies using generative AI

### The macroeconomic case for Tier 4 being large

A 2026 arXiv paper on the AI/data center macro impact ([arXiv:2601.11196](https://arxiv.org/pdf/2601.11196)) argues that the *revenues from selling AI/compute services* are of "the same order of magnitude as capex." In other words, the downstream output economy may be roughly as large as the construction economy that produced the facilities. The paper identifies two channels:

1. Compute sold for final use (consumer subscriptions, exports, government purchases)
2. Compute used as intermediate input by other producers — a car maker running AI inference, a company subscribing to enterprise AI tools for employees

Lazard's value-chain analysis confirms the dollar scale: the four largest hyperscalers (Alphabet, Amazon, Meta, Microsoft) will spend ~$600 billion on data center infrastructure in 2026, up from ~$350 billion in 2025.

### The honest catch: Tier 4 is geographically untethered

This is the crux of the user's intuition — and of community pushback. A data center sits in rural Virginia, central Ohio, west Texas, or eastern Oregon. The jobs *enabled* by its compute sit in San Francisco, Seattle, NYC, Austin, Bangalore, or remote-anywhere. The host community gets:

- 15 to 160 permanent on-site jobs
- A temporary surge of construction work (much of it out-of-state crews)
- Some local supply-chain spillover (electricians, HVAC, security services, food service)
- Strained grid and water infrastructure
- Often substantial tax abatement

And the host community does not get:

- The salaries of the SaaS engineers writing software on that compute
- The salaries of the AI researchers training models on that compute
- The investor returns on the enterprises using that compute as a competitive advantage

This is the structural mismatch driving the recent backlash. Brookings frames it bluntly: "the standard model of data center development has produced mostly short-term construction jobs in recent years and relatively little long-term, high-value tech activity or large-scale employment" ([Brookings, Feb 2026](https://www.brookings.edu/articles/turning-the-data-center-boom-into-long-term-local-prosperity/)).

## The Capital-to-Job Ratio Problem

The clearest single number, from Food & Water Watch's January 2026 analysis of Virginia's economic development data (cited in Tailwind Economics):

> One permanent data center job costs **$54 million in invested capital — 168 times more than the $322,000 average for non–data center jobs.**

Other concretes from the same analysis:

- Typical large data center: 1,688 construction workers (peak) → 157 permanent
- Nationwide permanent data center jobs in 2024: ~23,000
- Anthropic's $50B infrastructure commitment (Nov 2025) → 800 permanent positions
- Virginia tax exemption: projected $1.5M → actual $1.6B in FY2025 (118% YoY)
- Georgia: $2.5B FY2026 projected; state audit found **70% of subsidized projects would have located there anyway** — textbook deadweight loss
- Texas: $130M projection (2023) → $1B (Jan 2025) → $9B cumulative through 2030 projected

This is what makes the "data centers create jobs" framing so contested: at the unit level, they create *very few* per dollar invested. Even with the most generous industry multipliers (7.5 operations-only national), $54M of capex per direct job × 7.5 = roughly $7M per total job supported — still an order of magnitude above the typical industrial development ratio.

## Synthesis

**A data center with 15 permanent on-site jobs probably creates, very roughly:**

- **~20–80 additional indirect jobs** (supply chain — construction trades during build, equipment manufacturers, utilities, services). Mostly diffuse and partially out-of-region.
- **~20–80 additional induced jobs** (household spending of T1+T2 workers). Mostly local.
- **An indeterminate but potentially much larger number of Tier 4 jobs** (at customers and enterprises using the compute). Almost none geographically local.

State-modeled multipliers say 1.6× to 4.5× the direct headcount land as additional jobs in the state — so 15 direct → roughly 25 to 70 additional T2+T3 jobs in-state. National multipliers (5×–7×) capture more T2 because they include out-of-state manufacturing.

**The honest reading:** if "local jobs created" is the metric the community cares about, data centers are among the worst capital-intensive investments in modern economic development on a per-job basis. If "national GDP enabled" is the metric, the picture is dramatically better because of Tier 4 — but Tier 4 accrues to coastal tech labor markets and to hyperscaler shareholders, not to the host town.

This explains both the boosters and the critics. They are measuring different tiers.

## Where the Numbers Should Be Treated Skeptically

1. **Multipliers are modeled, not observed.** IMPLAN and input-output models work from assumed coefficients; they are not headcounts of actual people hired.
2. **Industry-funded studies skew high.** PwC (commissioned by the Data Center Coalition), Oxford Economics (commissioned by Google), and the Consumer Energy Alliance studies all sit at the upper end of the multiplier range.
3. **Construction jobs are real but temporary** and often filled by traveling crews from out of state — they show up in headline counts but do not produce sustained local employment.
4. **The "jobs supported" rhetorical move.** Data center news observers explicitly warn that "Multiplier-based 'jobs supported' are presented as 'jobs created'" — different things ([datacenternews.org, Feb 2026](https://datacenternews.org/data-center-jobs-how-many-jobs-do-data-centers-create/)).
5. **Brookings discloses** that Amazon, Google, Meta, and Microsoft are general donors. Even balanced analyses come with money attached.

## Open Questions

- How much of Tier 4 *could* be steered to host communities through deliberate policy (university R&D partnerships, compute vouchers, testbeds)? Brookings argues this is possible; there is little evidence yet that it works at scale.
- What is the long-run employment outlook once construction phases end? Wisconsin Watch quotes a researcher honestly: "We don't have a firm enough grasp about the indirect effects in the longer term."
- Are nuclear power tie-ins (Kewaunee, others) substantial enough to count as a distinct Tier 2 sub-economy with their own multipliers?
- How should the calculus change for AI-specific facilities (denser ops staffing, more specialized roles) versus traditional colocation?

## Recommended Application

If this question is for a personal-knowledge or framework-building purpose: the four-tier model above is durable and survives the contested-multiplier debate.

If it is for engagement with a specific local project or policy question, the right next step is to find the **project-specific impact study** the developer filed (developers typically commission these for site approvals) and audit it against the framework here — checking whether they bundled construction with permanent, whether they used national multipliers in a state context, and whether they presented "jobs supported" as "jobs created." Those three moves account for most of the inflation in industry job claims.
