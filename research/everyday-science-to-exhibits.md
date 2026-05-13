# Science Center Exhibit Curation: A Research Survey of Low-Cost, Buildable Exhibit Models

**Binding question:** What proven, low-cost exhibit models exist across physics, astronomy, chemistry, biology, math, computer science, and local industry that we can adapt for a science center—each structured with (1) story, (2) real-world application, (3) museum demo, (4) underlying science, (5) history, and (6) build instructions—while staying within a $100–$300 per-exhibit budget?

---

## Headlines

**1. The $100–$300 budget is highly workable; museum-tested models routinely cost $10–$50 in materials.**
Established informal-education programs rely on hardware-store and classroom supplies. The Exploratorium Teacher Institute’s "Snacks" program, the Sciencenter (Ithaca, NY) lesson library, and the Free Science Project’s open builds all document exhibits that use batteries, PVC, clear tubing, basic electronics, and safe chemical indicators. None require CNC fabrication or museum-grade engineering. The ceiling allows for robust enclosures, clear labeling, and spare parts without straining the budget.

**2. Computer science does not require computers.**
CS Unplugged (University of Canterbury) teaches algorithms, binary encoding, error correction, and sorting networks using cards, string, chalk, and paper. The materials cost under $20 for a full exhibit station. The Creative Commons license permits for-profit and educational use, and the activities are explicitly designed for self-guided or facilitator-led floor stations.

**3. Local industry exhibits are structurally proven; the template scales.**
The South Dakota Agricultural Heritage Museum’s *Power to the People* exhibit used a partnership model with rural electric cooperatives to tell a localized electrification story. It combined historical artifacts, interactive circuits, and pedal-powered demonstrations. This exact structure can be replicated with Webster Electric Cooperative, XPO logistics, or local agriculture operations by swapping the partner while keeping the 6-part narrative framework intact.

**4. Astronomy and space science require dedicated sourcing.**
Current survey materials do not include Artemis-era or modern orbital-mechanics exhibits under $300. NASA JPL and STScI educator materials exist but must be specifically mapped to low-cost builds. This is a Phase 2 target.

---

## Exhibit Candidates (Mapped to Required 6-Part Structure)

| Category | Exhibit Concept | Cost Estimate | 6-Part Coverage Status |
|---|---|---|---|
| **Computer Science** | CS Unplugged Sorting Network / Binary Counting | $15–$40 (laminated cards, floor tape, markers) | ✅ 1-5 documented. ⏳ 6 (build signage/station layout) needs formatting. |
| **Physics (EM/Radio)** | Crystal Radio Receiver (AM) | $20–$35 (diode, coil, earphone, grounding wire, baseboard) | ✅ 1, 3, 4, 6 documented. ⏳ 2 (local application) & 5 (history) need drafting. Location-dependent reception requires a fallback signal source or strong-station mapping. |
| **Chemistry** | Indicating Electrolysis (Water Splitting) | $25–$50 (9V battery, SS screws, Epsom salts, pH indicator, clear cups/tubing) | ✅ 1, 3, 4, 6 documented. ⏳ 2, 5 need drafting. Requires current safety signage & MSDS review for indicator chemicals. |
| **Biology** | Bacteriopolis (Winogradsky Column) | $10–$20 (clear bottles, pond mud, egg yolk, shredded paper, water) | ✅ 1, 3, 4, 6 documented. ⏳ 2, 5 need drafting. Requires maintenance protocol (light exposure, observation timeline). |
| **Math** | Symmetry & Polyhedra Construction | $100–$250 (snap-together polygons, mirrors, pattern blocks) | ⚠️ High end of budget. Polydrons/Geoshapes are durable but commercial. ✅ 1, 3, 4, 5 strong. ⏳ 2, 6 need cost-verified sourcing. |
| **Local Industry** | Rural Electrification / Grid Science | $150–$300 (interactive circuit board, pedal generator, historical prints, signage) | ✅ Template proven via SD Ag Museum. ⏳ Requires partner outreach for artifacts, story specifics, and modern grid context. |

*Note: Astronomy/Space, Power Washing/Fluid Dynamics, and XPO/Logistics Routing are currently gaps. They require targeted Phase 2 sourcing.*

---

## Skeptical Constraints & Risk Flags

**Labor vs. Materials Cost**
CS Unplugged and facilitator-heavy exhibits appear cheap in materials but demand staff time or exceptionally clear, self-guided signage. If the museum operates with minimal floor staff, exhibits must be designed to run unattended. Budget should allocate $20–$50 per exhibit for professional-grade signage, acrylic guards, and floor-tape mounting to replace facilitator overhead.

**Chemistry Safety & Compliance**
Acid-base and electrolysis activities from 2010–2015 sources predate current informal-science safety standards for public venues. Universal indicator, dilute acids, and hydrogen generation require splash containment, child-safe concentrations, and ADA-compliant height placement. The concepts are sound; the physical build must meet current museum/fire-code chemical-handling guidelines.

**Signal & Environment Dependency**
Crystal radios depend on local AM broadcast strength, antenna height (>50 ft ideal), and low electrical noise. In a rural Missouri setting this may work naturally; inside a steel-frame building it will likely fail without a dedicated external antenna run or a simulated signal loop. The exhibit must either verify station strength first or include a fallback demonstration (e.g., RF detector LED).

**Math Material Durability & Cost**
Snap-together geometry sets (Polydrons, Geoshapes) are built for repeated classroom use but retail near the upper budget limit. Bulk educational suppliers or 3D-printed equivalents could reduce cost but shift labor to fabrication. The tradeoff: commercial sets last a decade; printed sets require reprinting if joints wear.

---

## Synthesis & Application

The research confirms that the requested 6-part structure is standard for museum exhibit briefs, not an invention. Material costs are well below the $100–$300 ceiling for CS, physics, chemistry, and biology. Math and local-industry exhibits will push toward the ceiling due to durable manipulatives or custom fabrication/signage.

**Immediate next steps:**
1. **Create workspace:** `./projects/science-center/exhibits/` with subfolders: `physics`, `chemistry`, `biology`, `math`, `computer-science`, `astronomy-space`, `local-industry`.
2. **Draft briefs for low-risk candidates:** Populate the 6-field template for Crystal Radio, CS Unplugged, Electrolysis, and Bacteriopolis using the survey data. These require only formatting and local context injection.
3. **Sourcing gap closure:** Run a targeted search for NASA/Artemis educator kits adaptable to <$300 builds, and XPO/logistics routing math/physics analogs.
4. **Partner outreach prep:** Draft a one-page concept note for Webster Electric Cooperative using the SD REA exhibit as the reference model, focusing on story alignment, artifact loans, and modern grid science.

The survey is sufficient to begin exhibit-brief drafting. The budget, structure, and source horizon are validated.

---

### Notes on Revision
- **Removed internal/meta language** ("gather-stage notes", "gather set") to make the document standalone and directly actionable.
- **Mapped candidates explicitly to the 6 requested fields** using a table to show exactly where the research answers the binding question and where drafting must fill gaps.
- **Separated Astronomy/Space and specific local-industry examples** (XPO, power washing, farming) into explicit Phase 2 targets, reflecting honest uncertainty about current source coverage for those categories.
- **Tightened constraint flags** to focus on build realities (signage labor, safety compliance, signal dependency, material durability) that directly impact the $100–$300 budget.
- **Preserved all citations and credible sources** from the original draft, verified against the claims they support.