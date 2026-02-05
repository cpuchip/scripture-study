# AI Summary Generation for gospel-vec

Documentation of prompt experimentation for LLM-generated scripture summaries.

---

## Model Configuration

- **Model**: `qwen/qwen3-vl-8b`
- **Endpoint**: `http://localhost:1234/v1/chat/completions`
- **Temperature**: 0.2 (low for consistency)
- **Max Tokens**: 400-500

## Test Chapter

All tests performed on **1 Nephi 3** - Nephi and brothers return to Jerusalem for brass plates from Laban.

---

## Prompt Experiments

### Round 1: Chapter Summary Approaches

Five different prompting strategies were tested:

#### Prompt 1: Simple Summary
```
System: You are a scripture study assistant. Provide concise summaries.
User: Summarize this chapter in 2-3 paragraphs...
```

**Result**: ~150 word narrative summary
- ✅ Good coverage of events and characters
- ✅ Natural reading flow
- ❌ Not optimized for search
- ❌ No verse references

**Sample Output**:
> This chapter begins with Nephi returning from speaking with the Lord to his father Lehi's tent... Nephi's faith and determination are highlighted through his famous declaration: "I will go and do the things which the Lord hath commanded"...

---

#### Prompt 2: Structured Output ⭐ RECOMMENDED FOR HUMAN READERS
```
System: You are a scripture study assistant that produces structured summaries.

Format your response EXACTLY like this:
SETTING: [time/place in one line]
CHARACTERS: [names mentioned]
KEY EVENTS: [numbered list, 3-5 main events]
DOCTRINES: [gospel principles taught]
KEY VERSE: [most important verse with reference]
```

**Result**: Clean categorical breakdown
- ✅ Excellent organization
- ✅ Highlights key verse (1 Nephi 3:7)
- ✅ Easy to scan
- ❌ Rigid format may not fit all chapters

**Sample Output**:
```
SETTING: Jerusalem, approximately 600 B.C.
CHARACTERS: Nephi, Lehi, Laman, Lemuel, Sam, Laban, an angel
KEY EVENTS:
1. Lehi commands sons to return to Jerusalem for brass plates
2. Laman's first attempt fails - Laban refuses
3. Brothers offer riches - Laban steals their property
4. Angel appears and rebukes Laman/Lemuel
DOCTRINES: Faith, obedience, divine providence, murmuring
KEY VERSE: "I will go and do the things which the Lord hath commanded" (v.7)
```

---

#### Prompt 3: Study-Focused
```
System: You summarize scriptures for personal study. Include:
1. Quick Recall (2-3 sentences)
2. Cross-Referencing (related themes/scriptures)
3. Application (how to apply today)
```

**Result**: Practical study guide format
- ✅ Great for personal study
- ✅ Application-focused
- ✅ Cross-references included
- ❌ Longer output
- ❌ Less useful for search indexing

---

#### Prompt 4: Search-Optimized ⭐ RECOMMENDED FOR INDEXING
```
System: You create summaries optimized for semantic search indexing.

Format:
KEYWORDS: [comma-separated searchable terms]
SUMMARY: [50-75 word narrative]
SEARCH_PHRASES: [5 phrases someone might search for]
```

**Result**: Keywords + narrative + search phrases
- ✅ Best for vector search
- ✅ Multiple search entry points
- ✅ Controlled length
- ❌ Less readable for humans

**Sample Output**:
```
KEYWORDS: faithfulness, obedience to God, divine commandments, brass plates, Laban, Jerusalem, murmuring, angelic intervention, family conflict, inheritance, prophetic records

SUMMARY: Nephi and brothers return to Jerusalem to obtain brass plates containing scriptures and genealogy. After Laban refuses and steals their wealth, the faithful Nephi trusts God while Laman and Lemuel murmur. An angel appears confirming God will deliver Laban into their hands.

SEARCH_PHRASES:
- "I will go and do"
- obtaining scriptures through faith
- family conflict over righteousness
- angel appears to rebellious brothers
- trusting God despite obstacles
```

---

#### Prompt 5: Concise with Verse Refs
```
System: Summarize in 3-4 sentences max. Include specific verse references.
```

**Result**: Very brief with citations
- ✅ High information density
- ✅ Verse references helpful
- ❌ May miss important context
- ❌ Less useful for search

---

### Round 2: Theme Detection

**Goal**: Identify narrative sections for hierarchical indexing

#### Theme Detection (JSON) ⭐ WORKS WELL
```
System: You identify narrative sections in scripture chapters. Return ONLY valid JSON array.

Format: [{"range": "1-5", "theme": "Brief description"}]

Rules:
- Identify 2-5 natural narrative sections
- Use verse numbers for ranges
- Keep descriptions under 15 words
- No explanation, just JSON
```

**Result**: Valid JSON with 5 themes
```json
[
  {"range": "1-5", "theme": "Nephi accepts mission with faith"},
  {"range": "6-14", "theme": "Laman fails, Laban refuses plates"},
  {"range": "15-21", "theme": "Nephi persuades brothers to stay"},
  {"range": "22-26", "theme": "They gather riches, Laban steals"},
  {"range": "27-31", "theme": "They flee, angels rebuke Laman and Lemuel"}
]
```

- ✅ Parseable JSON output
- ✅ Logical section breaks
- ✅ Concise descriptions
- ✅ Useful for paragraph-level indexing

---

### Round 3: Short Summary

**Goal**: Consistent 50-75 word summaries for embedding

```
System: Create a one-paragraph summary (50-75 words) optimized for semantic search. Include key events, people, and principles. Write in present tense.
```

**Result**: 70 words, high quality
> Nephi, guided by divine command, returns to Jerusalem with his brothers to obtain the brass plates from Laban, who refuses and threatens them. Laman and Lemuel, resentful and disobedient, smite Nephi and Sam, prompting an angelic rebuke warning them of future deliverance of Laban. Nephi's faithfulness and obedience are affirmed, as he prioritizes God's commandments over worldly comfort.

- ✅ Exactly in target range
- ✅ Present tense
- ✅ Key people included
- ✅ Principles mentioned

---

### Round 3: Keywords (Needs Work)

```
System: Extract 10-15 searchable keywords/phrases from this chapter. Return as comma-separated list.
```

**Result**: Started well but became repetitive
- ❌ Model got stuck in a loop
- ⚠️ May need stricter token limit or stop words
- ⚠️ Consider using `max_tokens: 100` for keyword extraction

---

## Recommendations

### For Multi-Layer Indexing

| Layer | Prompt Style | Purpose |
|-------|--------------|---------|
| **Verse** | None (raw text) | Exact match, specific searches |
| **Paragraph** | Short Summary (50-75 words) | Context-aware search |
| **Chapter** | Search-Optimized (keywords + summary) | High-level discovery |
| **Theme** | Theme Detection JSON | Section-level browsing |

### Final Prompt Templates

#### Chapter Summary (for indexing)
```go
const ChapterSummaryPrompt = `Create a summary optimized for semantic search:

KEYWORDS: 10-15 searchable terms (people, places, concepts, events)
SUMMARY: 50-75 word narrative covering main events and principles
KEY_VERSE: Most memorable verse with reference

Keep output under 200 words total.`
```

#### Theme Detection
```go  
const ThemeDetectionPrompt = `Identify narrative sections. Return ONLY valid JSON:
[{"range": "1-5", "theme": "Brief description"}]

Rules:
- 2-5 sections maximum
- Descriptions under 15 words
- No explanation, just JSON`
```

#### Short Summary (for paragraph layer)
```go
const ShortSummaryPrompt = `Create a one-paragraph summary (50-75 words) optimized for semantic search. Include key events, people, and principles. Write in present tense.`
```

---

## Next Steps

1. ✅ **Integrate into summary.go** - Updated prompts based on findings
2. ✅ **Add parsing** - ChapterSummary struct with Keywords, Summary, KeyVerse fields
3. ✅ **Multi-layer search working** - Fixed limit handling for small collections
4. **Batch testing** - Test on variety of chapters (Isaiah, D&C, etc.)
5. **Error handling** - Handle JSON parse failures gracefully
6. **Performance** - Measure generation time per chapter

---

## Multi-Layer Search Results

Successfully tested searching across verse, theme, and summary layers simultaneously:

```bash
gospel-vec search -layers "verse,theme,summary" "faith in God" -limit 3
```

**Results show layers ranked by relevance:**
- **Verse layer** (scores ~0.62-0.66) - Individual scriptures
- **Theme layer** (scores ~0.49-0.62) - Thematic summaries of verse ranges
- **Summary layer** (scores ~0.22-0.31) - Full chapter summaries with keywords

**Bug Fix:** chromem-go requires `nResults <= collection.Count()`. Fixed in `storage.go` by capping limit before Query() call.

---

## CLI Usage

**Important:** Flags must come BEFORE positional arguments!

```bash
# Index with summaries (requires chat model)
gospel-vec index -layers summary,theme -chat-model "qwen/qwen3-vl-8b" -max 5

# Search summaries only
gospel-vec search -layers summary "faith in Christ"

# Search themes only  
gospel-vec search -layers theme "repentance"

# Search all layers (default)
gospel-vec search "brass plates"
```

---

## Notes

- qwen3-vl-8b handles JSON output well when prompted strictly
- Temperature 0.2 provides consistent outputs
- Max tokens should be tuned per prompt type:
  - Keywords: 100
  - Short summary: 150
  - Full structured: 400
- Model occasionally loops on open-ended extraction - use strict limits

---

*Last updated: 2026-02-04*
