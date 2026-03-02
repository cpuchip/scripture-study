# Principles — What We've Learned

Core insights extracted from 39 days of study, building, and reflection. These grow over time. Query semantically when relevant topics arise.

---

## Theological Framework

### The Matter Spectrum
Intelligence → Spirit → Element → Glorified Matter. A continuous spectrum, not a binary. "All spirit is matter, but it is more fine or pure" (D&C 131:7-8). This resolves:
- **Hard problem of consciousness:** Intelligence acts through spirit; no explanatory gap
- **Origin of moral law:** Truth is eternal, God works within it because He fully comprehends it
- **Problem of suffering:** Agency requires opposition; suffering is structural, not accidental
- **Why anything exists:** Intelligence and elements are eternal; there was never nothing

*Source: [truth.md](../../study/truth.md), [truth-atonement.md](../../study/truth-atonement.md), [04_observations.md](../../docs/04_observations.md)*

### Atonement as Refinement
Not just legal transaction — a physics of transformation. Christ descended below all things to comprehend the entire spectrum. His light is the energy that purifies. Degrees of glory = the law (light) you've been refined to hold. Subsumes penal substitution without contradicting it.

*Source: [truth-atonement.md](../../study/truth-atonement.md)*

### Becoming Is the Point
Everything in the gospel — Fall, law, opposition, ordinances, covenants, degrees of glory — is oriented toward transformation. Not knowledge for its own sake. Not obedience for its own sake. Becoming like God, full of light, comprehending all things.

*Source: [04_observations.md](../../docs/04_observations.md), [becoming overview](../../becoming/00_overview.md)*

### Gospel Pattern Is Eternal
The gospel was taught to Adam (Moses 6). Each dispensation reveals parts of the same whole. The internal consistency across volumes produced over 13 years by different authors is "genuinely unusual" compared to other theological systems.

*Source: [moses-6-gospel-to-adam.md](../../study/moses-6-gospel-to-adam.md), [04_observations.md](../../docs/04_observations.md)*

---

## Study Methodology

### Two-Phase Workflow
**Phase 1 — Discovery:** Use search tools freely (gospel-mcp, gospel-vec, webster-mcp). Note file paths and references.
**Phase 2 — Deep Reading:** `read_file` every scripture and talk you plan to cite. Follow footnotes. Verify quotes.

**Rule:** Search results are POINTERS, not SOURCES. Never use a search excerpt as a direct quote.

*Source: [01_reflections.md](../../docs/01_reflections.md), learned through mazzaroth and gadianton failures*

### Finding vs. Reading
The core quality problem: tools make finding fast but create shortcuts past deep reading. The degradation path is `read_file → reason → write` (good) degrading to `search → excerpt → write` (bad). The Gadianton-Hinckley case is the clearest example — vector search "found" the talk but AI never read it, missing the most powerful quote.

*Source: [01_reflections.md](../../docs/01_reflections.md), gadianton-robbers study*

### Webster 1828 as Model Tool
The ideal pattern: provides a specific, authoritative result (historical definition). AI reasons about it in context. Output is genuinely enhanced. Tool doesn't replace reading — it complements it. "Obtain" vs "receive" in D&C 84 was the breakthrough example.

*Source: [priesthood-oath-and-covenant.md](../../study/priesthood-oath-and-covenant.md), [01_reflections.md](../../docs/01_reflections.md)*

### Follow the Footnotes
Scriptural footnotes are insights handed to us on a silver platter. Moses 6:35 footnote → John 9:6 parallel. Moroni 7 → Ether 12 → Moroni 10 chain. The best discoveries come from what the scriptural editors already connected.

*Source: [01_reflections.md](../../docs/01_reflections.md) Phase 6*

---

## Collaboration Principles

### Intent Over Instruction
Communicate why, create space for how, evaluate by fruit. Works for AI collaboration, personal growth, and community leadership. The same pattern as personal revelation: God gives intent (Moses 1:39), not procedural checklists.

*Source: [intent engineering study](../../study/ai-responsible-use.md), Feb 24-25 sessions*

### Instruction Minimalism
Instructions bloat kills warmth. The best writing came from sessions with 5-8KB of instructions. By 22KB, the model optimized for compliance rather than engagement. Keep always-on instructions lean (~80 lines). Move workflow rules to agents/skills.

*Source: [05_instruction-refinements.md](../../docs/05_instruction-refinements.md), Feb 14 Opus 4.6 tone discovery*

### Compression Is Curation
Don't dump raw material into context. Curate what matters. Human judgment decides what to keep, what to summarize, what to discard. The AI can amplify that judgment but can't replace it. (Same principle as Nate Jones' memory architecture — applied to our context loading.)

*Source: memory-architecture proposal, reflections doc*

### Stability After Improvement
After making changes to instructions, tools, or architecture — let them prove themselves before changing again. The Feb 20 skills gap review concluded "our changes look good, I'm not going to change anything for a time." Resist the urge to continuously optimize.

*Source: [09_post-skills-quality-review.md](../../docs/09_post-skills-quality-review.md)*

---

## Tool Selection Heuristics

| Need | Tool | Why |
|------|------|-----|
| Exact phrase in scripture | gospel-mcp (keyword/FTS5) | Precise text matching |
| Conceptual/thematic search | gospel-vec (semantic) | Similarity across concepts |
| Historical word meaning | webster-mcp | Authoritative 1828 definitions |
| Conference talk discovery | gospel-vec `search_talks` | Semantic across 54 years |
| Verify a source | `read_file` | Nothing substitutes for reading |
| YouTube transcript | yt-mcp `yt_download` | Download and save to yt/ |
| Dictionary (modern) | webster-mcp `modern_define` | Current usage comparison |
| Track personal practice | becoming-mcp | Daily practice, memorization |

### Tool Selection Failures to Avoid
- **Keyword search for conceptual queries** → zero results accepted instead of switching to semantic search
- **Search result as direct quote** → always read the actual file
- **Directory link instead of file link** → verify the specific file path exists
- **"File not downloaded" without checking** → use `file_search` or `list_dir` to verify
