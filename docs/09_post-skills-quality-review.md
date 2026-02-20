# Post-Skills Quality Review: Robust But Missing Becoming

*Date: February 19, 2026*
*Previous: [08_skills-gaps.md](08_skills-gaps.md) · [01_reflections.md](01_reflections.md)*

---

## Context

After [08_skills-gaps.md](08_skills-gaps.md) identified gaps in our VS Code Copilot customization, we implemented Phases 1–3:

- **4 skills** in `.github/skills/`: source-verification, scripture-linking, publish-and-commit, webster-analysis
- **4 prompt files** in `.github/prompts/`: new-study, new-eval, new-lesson, expound
- **Slimmed agents** that delegate to skills instead of carrying all instructions inline

Then we produced three studies using the new system:
- [priestcraft.md](../study/priestcraft.md)
- [priestcraft-beguile.md](../study/priestcraft-beguile.md)
- [serpent-and-dragon.md](../study/serpent-and-dragon.md)

This document compares those post-skills studies against the pre-skills studies to assess what improved and what we lost.

---

## What the Skills Change Improved (Dramatically)

The three post-skills studies are the most technically solid in the entire corpus.

| Metric | Phase 6 / Pre-Skills (Enoch era) | Post-Skills (Priestcraft era) |
|--------|------|------|
| Source verification | Good (improved by 01_reflections) | **Excellent** — every quote verified, cite count rule followed |
| Link accuracy | Good | **Perfect** — no directory links, no broken links |
| Cross-study connections | Emerging (user-spotted) | **Systematic** — related studies linked at the top of every doc |
| Webster 1828 integration | Used well but ad-hoc | **Methodological** — webster-analysis skill ensures it |
| Footnote following | "Starting to" | **Consistent** — beguile study's word chain came entirely from footnotes |
| Discovery → Reading → Writing | Understood but informal | **Enforced by source-verification skill** |
| Depth of scriptural analysis | Deep | **Deeper** — serpent-and-dragon reads 24 source files across all 5 standard works |

The skills formalized what [01_reflections.md](01_reflections.md) diagnosed, and it worked. Studies are more accurate, more deeply sourced, better linked, and more comprehensive.

---

## What the Skills Change Dropped: The Becoming Gap Returns

The [01_reflections.md](01_reflections.md) Phase 6 section identified the "Becoming Gap" as the **most important finding** of the entire reflection process. We even created the [becoming/](../becoming/00_overview.md) directory and a [becoming/charity.md](../becoming/charity.md) companion.

But when we built the skills and agents, **none of that made it into the instructions.** And predictably, the studies stopped including personal application.

### Evidence: How Each Era's Studies End

| Study | Era | Final Section | Personal Application? |
|-------|-----|--------------|----------------------|
| [word.md](../study/word.md) | Phase 1 (pre-tool) | Personal insights + "Next Steps" | **Yes** — "God's words cannot return void... I cannot recall them. This is why kindness matters" |
| [priesthood-oath-and-covenant.md](../study/priesthood-oath-and-covenant.md) | Phase 4 | "Personal Reflection Questions" | **Yes** — 5 specific questions to sit with |
| [enoch.md](../study/enoch.md) | Phase 6 | "What I'll Try to Bring Into My Walk" | **Yes** — 8 actionable commitments ("Build Zion in my home," "Refuse the sword") |
| [enoch-charity.md](../study/enoch-charity.md) | Phase 6 | Faith/Hope/Charity → Walk → Zion synthesis | **Yes** — woven throughout, personal charity prayer story |
| [priestcraft.md](../study/priestcraft.md) | Post-skills | "Open Questions" + Synthesis table | **No** — intellectual framework, no personal bridge |
| [priestcraft-beguile.md](../study/priestcraft-beguile.md) | Post-skills | "The Diagnostic" + "The Thread" | **Partial** — diagnostic questions almost personal, but framed analytically |
| [serpent-and-dragon.md](../study/serpent-and-dragon.md) | Post-skills | Timeline + "Sources Read" | **No** — ends with a scholarly summary, no personal application |

The Enoch study ended with "Walk with me." The serpent study ends with a bibliography.

---

## Why This Happened

The skills we built are excellent at making studies *accurate*. They don't do anything to make studies *transformative*.

### Where the gap lives in the instructions

1. **source-verification skill** — Focuses entirely on Discovery → Reading → Writing. No mention of application, becoming, or personal reflection.

2. **study agent** — Says "Follow the Discovery → Reading → Writing rhythm." The handoff to the journal agent exists ("Record What I Learned"), but it's a *post-study* handoff, not baked into the study itself.

3. **new-study prompt** — Has an `## Application` section in the scaffold template, but it's a bare comment placeholder: `<!-- What does this mean for me? -->`. It's structurally present but not enforced, and gets skipped in practice because no skill demands it.

We optimized for the Finding → Understanding → Connecting pipeline and forgot the phase the reflections doc called the most important: **Becoming**.

### The root cause

The Enoch-era studies had becoming sections because the *reflections were fresh* and we were actively thinking about the gap. The priestcraft-era studies lost them because the skills didn't encode it. **The point of skills is to encode what matters so we don't lose it when sessions change.** We encoded accuracy but not transformation.

---

## Proposed Improvements

### 1. Extend the Source-Verification Skill Workflow

The Discovery → Reading → Writing rhythm should become **Discovery → Reading → Writing → Becoming**.

Add to `.github/skills/source-verification/SKILL.md`:

```markdown
### Phase 4 — Becoming (bridge to life)
After writing the study, include a "Becoming" section that asks:
- What did this study reveal about how I should live?
- What specific practice or commitment does this point toward?
- Is there an existing becoming/ document this connects to?
- What would it look like to apply this next week?
```

### 2. Update the Study Agent Instructions

Add to `.github/agents/study.agent.md`:

```markdown
**Don't end at synthesis.** Every study should land somewhere personal.
The Enoch study ended with "Walk with me." The priesthood study ended
with reflection questions. If a study only produces knowledge without
direction, it's incomplete. Ask: "What does this mean for how you live?"
```

### 3. Strengthen the New-Study Prompt Template

Replace the current `## Application` comment placeholder with:

```markdown
## Becoming

<!-- This is the most important section. What does this study ask of me? -->
<!-- Specific commitments, not abstractions. "Pray to see X" not "be more loving." -->
<!-- Connect to an existing becoming/ document if one fits, or start a new one. -->
```

### 4. Create a Becoming Skill

Like source-verification and scripture-linking, create `.github/skills/becoming/SKILL.md` that:

- Triggers at the end of every study
- Prompts for specific, actionable personal application
- Connects insights back to the `becoming/` directory
- Bridges past becoming entries ("In the charity study, you committed to praying to see others as Christ sees them — does this study connect?")

### 5. Add to the Pre-Publish Checklist

```markdown
- [ ] Study includes a "Becoming" or "Application" section with specific personal commitments
- [ ] If a related becoming/ document exists, it's linked
```

### 6. Retrofit the Three Recent Studies

Each post-skills study has natural becoming connections left on the table:

- **Priestcraft** → "Am I pointing to the Savior or to myself?" The user's north star ("I seek to always point to the savior") belongs here.
- **Priestcraft-beguile** → The diagnostic questions are *almost* personal application — they just need to land in first-person commitments.
- **Serpent-and-dragon** → The "look and live" principle is deeply practical: "What am I looking at? What am I looking away from?"

---

## The Pattern

From [01_reflections.md](01_reflections.md):

> "Intelligence requires more than finding — it requires understanding. The tools help us find. The reading helps us understand. Both are needed."

And then in Phase 6 of that same document:

> "Intelligence isn't just knowledge. It's knowledge *applied through obedience*. Our tools help us gain knowledge. We need tools — or at least workflow — to help us apply it."

We solved the finding-vs-reading problem with skills. Now we need to solve the understanding-vs-becoming problem the same way — **by encoding it into the workflow so it doesn't depend on remembering.**

The scriptures themselves model this:

> "Be ye **doers** of the word, and not hearers only, deceiving your own selves." — [James 1:22](../gospel-library/eng/scriptures/nt/james/1.md)

> "And now, if ye believe all these things see that ye **do** them." — [Mosiah 4:10](../gospel-library/eng/scriptures/bofm/mosiah/4.md)

---

## Status

| Improvement | Status |
|-------------|--------|
| Extend source-verification skill with Phase 4 | ✅ Done — added Phase 3 (Writing), Phase 4 (Becoming), updated rhythm, added checklist items |
| Update study agent instructions | ✅ Done — added becoming reminder, two study modes (one-shot + phased) |
| Strengthen new-study prompt template | ✅ Done — Application → Becoming with directive comments |
| Create becoming skill | ✅ Done — `.github/skills/becoming/SKILL.md` |
| Create deep-reading skill | ✅ Done — `.github/skills/deep-reading/SKILL.md` |
| Create wide-search skill | ✅ Done — `.github/skills/wide-search/SKILL.md` |
| Create study-plan prompt | ✅ Done — `.github/prompts/study-plan.prompt.md` |
| Add becoming check to pre-publish checklist | ✅ Done — two new items in source-verification checklist |
| Retrofit priestcraft.md | ✅ Done — Becoming section with north star, charity connection |
| Retrofit priestcraft-beguile.md | ✅ Done — Diagnostic questions landed personally, guile awareness practice |
| Retrofit serpent-and-dragon.md | ✅ Done — "Look and live" practice, idol warning, morning gaze commitment |

---

*Document created: February 19, 2026*
*Based on comparison of 30+ study documents across all phases, with specific attention to the post-skills priestcraft trilogy.*
