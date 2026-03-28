# Study Documents

*Content inventory for model experiment planning.*

---

## Structure

```
study/
├── *.md                    (44 topic studies)
├── yt/                     (11 YouTube evaluations + voice analysis)
├── talks/                  (4 conference talk analyses)
├── ai/                     (3 AI strategy docs)
├── plan-of-salvation/      (10 phased deep-dive files)
├── stories/                (1 narrative transformation)
├── podcast/                (3 shareable audio/video scripts)
├── atonement/              (4 topical deep-dives)
├── cfm/                    (6 Come Follow Me lesson preps)
├── eq/                     (2 Elder's Quorum lesson preps)
├── teaching-in-the-saviors-way/ (2 pedagogy studies)
└── .scratch/               (scratch files for in-progress studies)
```

**Total:** ~90+ files across all subdirectories

## Document Types

### Topic Studies (44 files, `study/*.md`)
The core output of this project. Written by the study agent using phased methodology.

**Format:**
- Status header (study date, binding question)
- Word etymology (Webster 1828 when relevant)
- Scripture analysis with cross-references
- Synthesis and application
- Genuine questions and open threads
- Links to gospel-library files

**Examples:** agency.md, charity.md, covenants.md, creation.md, intelligence.md, stewardship-pattern.md, only-begotten.md, serpent-and-dragon.md

**Token range:** 5,000-30,000 tokens per study (some are very deep)

### YouTube Evaluations (`study/yt/`, 11 files)
Written by the eval agent. Critical analysis of YouTube content about LDS topics.

**Format:** Phased evaluation with thesis, evidence, problems, and recommendation.

### Talk Analyses (`study/talks/`, 4 files)
Conference talk analysis for teaching patterns. Written by the review agent.

### AI Strategy (`study/ai/`, 3 files)
Meta-documents about AI collaboration: fatigue, multi-agent ideas, staying relevant.

### Lesson Preps (`study/cfm/`, `study/eq/`, 8 files)
Teaching preparation with doctrinal + narrative layers.

## Digestion Considerations

- **Studies are the highest-value original content** in this project
- **Consistent format** thanks to the study agent — parseable sections
- **Cross-references to gospel-library** — studies link to their source scriptures
- **Good candidate for embedding** — would enable "find studies related to X"
- **Scratch files** (`.scratch/`) contain research provenance — lower priority for embedding but valuable for context
- **Total volume is small** — all 90+ files probably fit in a single 262k context window
- **Individual studies fit easily** in any model's context (even 32k)
