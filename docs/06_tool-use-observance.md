# Tool Use Observance

*A running log of tool behavior, gaps, and improvement ideas. Not a complaint box — a collaboration improvement tracker.*

*Started: February 15, 2026*

---

## How to Use This Document

When something about a tool stands out during a session — good or bad — note it here. Patterns matter more than individual incidents. Over time, this log helps us decide what to build, what to fix, and what to work around.

Categories:
- **Context pressure** — tools that return too much, pushing the context window
- **Missing tools** — gaps where a tool *should* exist but doesn't
- **Behavior** — unexpected results, formatting issues, timeouts
- **Wins** — tools working especially well, worth noting what makes them effective

---

## Observations

### February 15, 2026

**Verse-level retrieval (Missing tool)**
When *building* a document and you already know which verses to cite, pulling full chapter files via `read_file` is expensive — you get the whole chapter when you need three verses. A dedicated tool to fetch a specific verse or range (e.g., `get_verse("moses", 6, 57)` or `get_verses("dc", 93, 24, 30)`) would save significant context window space during document construction.

Important distinction: this tool should NOT replace reading full chapters during *study*. Study needs footnotes, surrounding verses, and cross-references — you need the full context to discover things. But when you're mid-document and just need to pull an accurate quote, a verse-level fetch would be much leaner.

**Context window pressure (Pattern)**
Some MCP tool responses are verbose — full search results or transcript chunks can fill the context window quickly, especially in longer sessions. This is exacerbated by the current model context limits in GitHub Copilot. Worth tracking which tools are the biggest offenders and whether response truncation or summarization options would help.

**gospel-mcp search (Behavior)**
Full-text search works well for exact phrases. Semantic search via gospel-vec complements it for concept-level queries. The two together cover most needs. Worth watching for cases where neither finds what we need.

---

## Ideas for New Tools

| Idea | Priority | Notes |
|------|----------|-------|
| Verse-level retrieval | High | Get specific verse(s) without reading full chapter file. For document building, not study |
| Scripture cross-reference lookup | Medium | Given a verse, return its footnotes/cross-references without reading the whole chapter |
| Study document index | Low | Search across `/study/` files by topic, date, or connected scriptures |

---

*This is a living document. Add observations as they arise during any session.*
