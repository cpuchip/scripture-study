---
name: webster-analysis
description: Look up words in Webster 1828 dictionary and analyze how historical meanings illuminate scripture. Use when studying Restoration-era vocabulary, when a word's historical meaning may differ from modern usage, or when the user asks about word origins.
argument-hint: "[word or phrase to analyze]"
---

# Webster 1828 Analysis

## Why Webster 1828?

Noah Webster published his *American Dictionary of the English Language* in 1828 — two years before the Book of Mormon was published and two years before the Church was organized. It captures the precise meaning of English words as Joseph Smith and his contemporaries understood them.

When a scripture uses a word like "intelligence," "virtue," "charity," or "glory," the 1828 definition often reveals meaning that modern usage has shifted away from.

## Two editions — know which one you're quoting

The webster server carries **two dictionaries** under truthful names:

- **Webster 1828** (`webster_define`) — the genuine 1828 American Dictionary, the Restoration-era authority. This is the edition for scripture word-work.
- **Webster 1913** (`webster1913_define`) — Webster's Revised Unabridged, 85 years later. Useful for tracking how a meaning drifted across the 19th century, NOT a witness of Restoration-era usage.

History note: some digital "Webster 1828" datasets actually serve 1913 text mislabeled as 1828 — this toolkit hit exactly that. Any study quoting "Webster 1828" from before that date needs re-verification. **Verify the edition of a source, not just the quote.**

## When to Use It

- **Restoration-era vocabulary:** Words in the Book of Mormon, D&C, and Pearl of Great Price
- **Biblical King James language:** The KJV was translated in 1611, but Webster 1828 captures meanings closer to how early Saints read those words
- **Words that have narrowed:** "Virtue" (1828: strength, acting power, efficacy) vs. modern (moral excellence only)
- **Words that have shifted:** "Intelligence" (1828: understanding; skill; also "a spiritual being") vs. modern (mental capacity, IQ)
- **Words Joseph Smith emphasized:** Often when a prophet repeats a word, it carries specific weight
- **Semantic drift studies:** Use `define` to see 1828 → 1913 → modern in one call — how a word moved over two centuries

## How to Use It

### Step 1: Look Up the Word
Use the `mcp__webster__webster_define` tool (or `mcp__webster__define` for the full 1828 → 1913 → modern comparison):
```
mcp__webster__webster_define("intelligence")
```

### Step 2: Compare Historical vs. Modern
Note what's different. Is the 1828 meaning broader? Narrower? Entirely different? Did the 1913 already lose it?

### Step 3: Read It Back Into Scripture
Re-read the scripture with the 1828 meaning in mind. Does the passage open up?

**Example (verified against the genuine 1828, 2026-06-09):**
- D&C 93:29 — "Intelligence, or the light of truth, was not created or made"
- Webster 1828: INTELLIGENCE, sense 1 — "Understanding; skill."; sense 4 — "A spiritual being; as a created intelligence. It is believed that the universe is peopled with innumerable superior intelligences."
- Modern: Mental capacity, cognitive ability, IQ
- **Insight:** In 1828, an "intelligence" could be a *being*, not just a faculty. Joseph Smith's readers heard "intelligences" (Abraham 3:22) as a known category of existence — the word itself carried the doctrine.

### Step 4: Check Cross-References
Use the `webster_define` result as a lens to search for other scriptures that use the same word. Does the 1828 meaning illuminate those passages too?

## Patterns to Watch For

| Pattern | Example |
|---------|---------|
| **Broader than modern** | "Virtue" = strength, acting power (Luke 8:46 — "virtue had gone out of me") |
| **More concrete than modern** | "Charity" = "that disposition of heart which inclines men to think favorably of their fellow men, and to do them good" |
| **Theological precision** | "Spirit" sense 5 = "the intelligent, immaterial and immortal part of human beings" |
| **A being, not just a quality** | "Intelligence" sense 4 = "A spiritual being; as a created intelligence" |

## Integration with Studies

When you discover a significant Webster 1828 insight, it often deserves its own section in the study document. Don't just footnote it — let it reshape how you read the passage.

The insight from Webster 1828 is the starting point, not the endpoint. Follow it into the scriptures and see where it leads.

## The Model Tool

Webster 1828 is the *model* MCP tool in this project. It provides a discrete, authoritative answer that you then reason about in context. It doesn't replace deep reading — it enriches it. Every tool in the project should aspire to work this way: return something specific and trustworthy, then get out of the way so the real thinking can happen.
