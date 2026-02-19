---
name: webster-analysis
description: "Look up words in Webster 1828 dictionary and analyze how historical meanings illuminate scripture. Use when studying Restoration-era vocabulary, when a word's historical meaning may differ from modern usage, or when the user asks about word origins."
user-invokable: true
argument-hint: "[word or phrase to analyze]"
---

# Webster 1828 Analysis

## Why Webster 1828?

Noah Webster published his *American Dictionary of the English Language* in 1828 — two years before the Book of Mormon was published and two years before the Church was organized. It captures the precise meaning of English words as Joseph Smith and his contemporaries understood them.

When a scripture uses a word like "intelligence," "virtue," "charity," or "glory," the 1828 definition often reveals meaning that modern usage has shifted away from.

## When to Use It

- **Restoration-era vocabulary:** Words in the Book of Mormon, D&C, and Pearl of Great Price
- **Biblical King James language:** The KJV was translated in 1611, but Webster 1828 captures meanings closer to how early Saints read those words
- **Words that have narrowed:** "Virtue" (1828: power, efficacy, strength) vs. modern (moral excellence only)
- **Words that have shifted:** "Intelligence" (1828: understanding, the faculty of understanding) vs. modern (mental capacity, IQ)
- **Words Joseph Smith emphasized:** Often when a prophet repeats a word, it carries specific weight

## How to Use It

### Step 1: Look Up the Word
Use the `webster_define` tool (or `define` for modern comparison):
```
webster_define("intelligence")
```

### Step 2: Compare Historical vs. Modern
Note what's different. Is the 1828 meaning broader? Narrower? Entirely different?

### Step 3: Read It Back Into Scripture
Re-read the scripture with the 1828 meaning in mind. Does the passage open up?

**Example:**
- D&C 93:36 — "The glory of God is intelligence"
- Webster 1828: INTELLIGENCE — "Understanding; skill. The act of understanding."
- Modern: Mental capacity, cognitive ability, IQ
- **Insight:** God's glory isn't about being smart — it's about *understanding*. The deepest kind of knowing.

### Step 4: Check Cross-References
Use the `webster_define` result as a lens to search for other scriptures that use the same word. Does the 1828 meaning illuminate those passages too?

## Patterns to Watch For

| Pattern | Example |
|---------|---------|
| **Broader than modern** | "Virtue" = power, not just moral purity (Luke 8:46 — "virtue had gone out of me") |
| **More concrete than modern** | "Charity" = benevolence, pure love (not just giving money) |
| **Theological precision** | "Atonement" = reconciliation, at-one-ment (the act of making two parties one) |
| **Surprising connections** | "Intelligence" and "light" and "truth" all defined in terms of each other |

## Integration with Studies

When you discover a significant Webster 1828 insight, it often deserves its own section in the study document. Don't just footnote it — let it reshape how you read the passage.

The insight from Webster 1828 is the starting point, not the endpoint. Follow it into the scriptures and see where it leads.

## The Model Tool

Webster 1828 is the *model* MCP tool in this project. It provides a discrete, authoritative answer that you then reason about in context. It doesn't replace deep reading — it enriches it. Every tool in the project should aspire to work this way: return something specific and trustworthy, then get out of the way so the real thinking can happen.
