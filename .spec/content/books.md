# Books

*Content inventory for model experiment planning.*

---

## Available Books

### Lectures on Faith (`books/lecture-on-faith/`)
- **Files:** 9 (00_preface.md, 00_introduction.md, 01-07_lecture_*.md)
- **Format:** YAML frontmatter + markdown sections + numbered paragraphs
- **Content:** Seven theological lectures delivered in Kirtland (1834-35), attributed to Joseph Smith. Core doctrines: nature of God, faith as power, sacrifice and knowledge.
- **Token estimate:** ~50,000-80,000 tokens total (fits any model)
- **Value:** High — foundational Restoration theology, rarely studied in mainstream curriculum

### Debugging: The 9 Indispensable Rules (`books/debugging/`)
- **Files:** PDF + extraction scripts + raw text (17 chapter markdown files under `9-indispensable-rules/`)
- **Format:** Extracted text from book
- **Content:** Agans' debugging methodology, mapped to spiritual/intellectual debugging patterns
- **Token estimate:** ~60,000-100,000 tokens
- **Value:** Medium for model experiments — already processed and mapped to debug agent

### Creator's Playbook (`books/creators-playbook/`)
- **Files:** 88+ episode directories + voice profiles + catalog.json
- **Format:** Episode records (EP_00 through EP_88) with subtopic tags
- **Content:** Creative writing/storytelling resource — narrative archetypes, heroes, magic systems, quests
- **Token estimate:** Large — 88 episodes with structured data
- **Value:** Lower priority for gospel-focused model experiments

## Digestion Considerations

- **Lectures on Faith** is the obvious first candidate — small enough to fit in any context window, theologically rich, already structured
- **Debugging book** is already digested and mapped — debug agent uses it
- **Creator's Playbook** is large structured data — could be useful for narrative agent but not gospel experiments
- **All books are small enough** to fit in 262k or 1M context windows entirely
- **Lectures on Faith** would be an excellent benchmark document — feed the whole thing to a model and ask it to summarize, analyze, cross-reference with current scripture
