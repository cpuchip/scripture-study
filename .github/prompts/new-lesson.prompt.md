---
name: new-lesson
description: "Start lesson preparation — scaffolds a lesson plan from the template with teaching principles"
agent: lesson
argument-hint: "[topic] [class] [date]"
tools: [read, edit, search, "gospel-engine-v2/*", "webster/*", "search/*"]
---

Prepare a lesson for teaching.

## Setup

1. Read the lesson template for structure: [docs/lesson_template.md](../../docs/lesson_template.md)
2. Check for relevant prior studies in `study/` that could inform this lesson
3. If Come Follow Me, locate the current manual: `gospel-library/eng/manual/come-follow-me-for-home-and-church-old-testament-2026/`

## Scaffold

Create a new file at `lessons/${input:path:subfolder/filename}.md` with:

```markdown
# ${input:topic:Lesson Topic}

**Date:** ${input:date:YYYY-MM-DD}
**Class:** ${input:class:Sunday School / EQ / RS}
**Manual Reference:**
**Time Available:** ${input:time:40 minutes}

---

## Personal Preparation
- [ ] Pray for guidance and for those I will teach
- [ ] Study the assigned material personally
- [ ] Ponder how these principles have blessed my life

---

## Opening (5 min)

**Attention/Introduction:**
<!-- A thought-provoking question, brief personal experience, or scripture -->

---

## Principle 1:

**Scripture(s):**

**Discussion Question(s):**
<!-- "What..." or "How..." questions that allow multiple valid answers -->

**Key Insight:**

---

## Principle 2:

**Scripture(s):**

**Discussion Question(s):**

**Key Insight:**

---

## Invitation to Act (5 min)

**This week, I invite you to:**

---

## Closing Testimony (3 min)
```

## Then Begin

1. **Read the assigned curriculum** from the manual
2. **Cross-reference** additional scriptures and talks
3. **Design 2-3 discussion questions** — "What..." or "How..." not "Did..." or "Is..."
4. Focus on **application** — help learners apply principles, not just cover content

**Remember:** A 20-minute discussion needs 2-3 key scriptures and 1-2 good questions, not an exhaustive cross-reference. The lesson is not a study document.
