# AI Fatigue Is Real

A reflection on [Siddhant Khare's article](https://siddhantkhare.com/writing/ai-fatigue-is-real) and what it means for how we work — especially when you've pushed yourself hard and you're feeling it.

> "Come unto me, all ye that labour and are heavy laden, and I will give you rest." — [Matthew 11:28](../../gospel-library/eng/scriptures/nt/matt/11.md)

---

## The Core Problem

Khare names the paradox nobody warned us about:

> "AI reduces the cost of production but increases the cost of coordination, review, and decision-making. And those costs fall entirely on the human."

When each task takes less time, you don't do fewer tasks. You do *more* tasks. Your capacity appears to expand, so the work expands to fill it. The baseline moves. Your manager's expectations adjust. Your *own* expectations adjust.

Before AI, the work itself imposed speed limits. You could only type so fast, think so fast, look things up so fast. Those limits were frustrating — but they were also a governor. **AI removed the governor.** Now the only limit is your cognitive endurance. And most people don't know their limits until they've blown past them.

---

## The Seven Patterns of AI Fatigue

### 1. You Became a Reviewer

Before AI: think → write → test → ship. You were the *creator*.

After AI: prompt → wait → read → evaluate → decide if correct → decide if safe → decide if it matches the architecture → fix → re-prompt → repeat. You became a *reviewer*. A quality inspector on an assembly line that never stops.

**Creating is energizing. Reviewing is draining.** Generative work gives you flow states. Evaluative work gives you decision fatigue. Hundreds of small judgments all day, every day.

### 2. Nondeterminism

Same prompt Monday, different output Tuesday. No stack trace for "the model decided to go a different direction today." Engineers are wired for determinism — same input, same output. AI broke that contract.

> "You are collaborating with a probabilistic system, and your brain is wired for deterministic ones. That mismatch is a constant, low-grade source of stress."

The engineers who handle this best treat AI output like a first draft from a smart but unreliable intern. They *expect* to rewrite 30%. They don't get frustrated when the output is wrong because they never expected it to be *right*. They expected it to be *useful*.

### 3. The FOMO Treadmill

New tools every week. New frameworks every month. Every LinkedIn post: "If you're not using AI agents with sub-agent orchestration in 2026, you're already obsolete." You spend weekends evaluating new tools. You set up a workflow Saturday, see something "better" Monday, start over the next weekend.

Knowledge decay compounds this — prompting best practices shift, model updates change behavior, the workflow you invested two weeks building produces worse results three months later.

### 4. The Prompt Spiral

The AI output is 70% right. You refine the prompt. 75% but broke something. Refine again. 80% but the structure changed. Forty-five minutes later, you could have written the thing from scratch in twenty.

It *feels* productive. You're iterating. Getting closer. But the marginal returns are diminishing and you've lost sight of the actual goal: shipping the feature.

### 5. Perfectionism vs. Probabilistic Output

AI output is never perfect. 70-80% there. Variable names are off, error handling is incomplete, edge cases ignored. For a perfectionist, "almost right" is worse than "completely wrong" — because you spend an hour tweaking instead of throwing away and starting fresh.

> "The engineers who struggle most with AI are often the best engineers. The ones with the highest standards. AI rewards a different skill: the ability to extract value from imperfect output quickly, without getting emotionally invested in making it perfect."

### 6. Thinking Atrophy

Outsource your first-draft thinking to AI long enough and the muscle atrophies. Like GPS and navigation — you stop building mental maps.

> "The struggle is where learning happens. The confusion is where understanding forms. Skip that, and you get faster output but shallower understanding."

### 7. The Comparison Trap

Everyone else's AI workflow posts are highlight reels. Nobody posts "I spent 3 hours trying to get Claude to understand my database schema and eventually gave up." Nobody posts "I'm tired."

---

## What Actually Helped (Khare's Strategies)

These are the concrete changes Khare made:

| Strategy | What It Looks Like |
|----------|-------------------|
| **Time-box AI sessions** | Set a timer. 30 minutes with AI. When it rings, ship what you have or write it yourself. Prevents the prompt spiral *and* the perfectionism trap. |
| **Separate AI time from thinking time** | Morning is for thinking — paper, sketching, reasoning. Afternoon is for AI-assisted execution. Brain gets both exercise and assistance. |
| **Accept 70% from AI** | Stop trying to get perfect output. 70% usable is the bar. Fix the rest yourself. Single biggest reducer of AI-related frustration. |
| **Three-prompt rule** | If AI doesn't get you to 70% usable in three attempts, write it yourself. No exceptions. |
| **Be strategic about the hype cycle** | Use *one* primary tool and know it deeply. Evaluate new tools after they've proved themselves over months, not days. Being informed and being reactive are different things. |
| **Log where AI helps vs. hurts** | Two weeks of tracking: task, used AI (y/n), time spent, satisfaction. Data reveals where AI actually saves time (boilerplate, docs, tests) and where it costs time (architecture, complex debugging, deep context). |
| **Don't review everything** | Focus review energy on what matters: security boundaries, data handling, error paths. Lean on automated tests and static analysis for the rest. |
| **Build on the durable layer** | Tools churn. Problems don't. Invest in understanding the problems underneath the tools. |

---

## What This Means for Us

This article isn't about software engineering — not really. It's about the human cost of working alongside something that never gets tired. A few things I think we should sit with:

### The Governor Metaphor

The most important sentence in the whole article:

> "AI removed the governor. Now the only limit is your cognitive endurance."

This is the Sabbath principle dressed in engineering language. God built rest into the structure of creation — not because He was tired, but because *we need limits*. The seventh day isn't laziness. It's design. When we remove our own governors, we aren't being more productive. We're overriding something God put there on purpose.

See [D&C 59:9-13](../../gospel-library/eng/scriptures/dc-testament/dc/59.md) — the Sabbath commandment isn't about not working. It's about "offering thine oblations" and "confessing thy sins" and going to "the house of prayer." It's *structured rest* — rest with a purpose. The governor isn't just a speed limit. It's a redirect.

### Creating vs. Reviewing

Khare says creating is energizing and reviewing is draining. That's true. But it maps to something deeper.

In [Abraham 4-5](../../gospel-library/eng/scriptures/pgp/abr/4.md), the Gods "organized and formed" — they were creators. They watched things "until they obeyed." They didn't micromanage. They set things in motion and let them become.

When we use AI well, we're in that creative posture — setting direction, watching what emerges, shaping it. When we use it poorly, we become quality inspectors on an assembly line. The *mode of engagement* matters as much as the output. Are we creating, or are we just reviewing?

### Know When to Stop

> "The real skill of the AI era is not prompt engineering. It's knowing when to stop."

This is the "good enough" principle. The Savior sent the seventy out with almost nothing ([Luke 10:4](../../gospel-library/eng/scriptures/nt/luke/10.md)) — no purse, no scrip, no shoes. He didn't optimize their kit. He sent them. Good enough. Go.

Perfectionism dressed as diligence is still perfectionism. Sometimes the study is done at 80%. Sometimes the lesson prep is good enough. Sometimes the MCP server works and you should stop tweaking it.

### The Thinking Atrophy Warning

This one matters most for scripture study. If we always reach for AI first, we stop building the neural pathways that come from struggling with a text ourselves. The footnotes are *right there*. The cross-references are *right there*. The Spirit can work with any mind that's engaged — but it can't work with a mind that's outsourced the engagement.

Khare's solution: first hour of every day, no AI. Think on paper. Sketch by hand. Reason the slow way.

For us, that maps to: *read the chapter yourself first*. Sit with it. Let questions form naturally. *Then* bring AI in to cross-reference, check Hebrew, find connections. The struggle is where revelation happens — not in the AI output, but in the quiet space before you ask for it.

See the Bednar framework in [our earlier AI study](../ai-responsible-use.md) — especially the two questions: Does this invite or impede the Holy Ghost? Does this enlarge or restrict your capacity to live, love, and serve?

---

## Practical Application — When You've Pushed Too Hard

You said you pushed hard last week and you're feeling it. Based on what Khare learned and what I think matters:

**Right now (today/tomorrow):**
- Close some browser tabs. Literally. The visual clutter is cognitive weight.
- If there's a task you keep prompting AI about and it keeps not quite getting there — write it yourself or shelve it. The prompt spiral is real and it's draining you.
- Do something with your hands that has nothing to do with a screen.

**This week:**
- Try the morning-thinking / afternoon-AI split. Even two days of it will tell you something.
- Pick *one* thing you've been trying to get AI to do perfectly and accept 70% on it. Ship it. Move on.
- If you're evaluating any new tools right now, stop. You have tools. They work. The new ones will still be there next month.

**Ongoing:**
- The three-prompt rule is a gift. Adopt it.
- Time-box AI sessions. A 30-minute timer changes the whole dynamic.
- Rest isn't the absence of productivity. It's the thing that makes productivity sustainable.

---

## The Bottom Line

> "If you're tired, it's not because you're doing it wrong. It's because this is genuinely hard. The tool is new, the patterns are still forming, and the industry is pretending that more output equals more value. It doesn't. Sustainable output does."

AI never gets tired. You do. That's not a bug in you — it's a feature. It means you're a person, not a machine. The fatigue is signal, not noise. Listen to it.

> "Take care of your brain. It's the only one you've got, and no AI can replace it."

---

*Source: [AI Fatigue Is Real and Nobody Talks About It](https://siddhantkhare.com/writing/ai-fatigue-is-real) by Siddhant Khare, February 7, 2026.*
