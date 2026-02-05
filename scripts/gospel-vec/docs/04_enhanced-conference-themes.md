# Enhanced Conference Themes & Hierarchical Summarization

Analysis of token capacity for session-level, conference-level, and book-level summarization.

---

## Token Capacity Analysis

### Current LM Studio Context Limits
- **32K tokens**: Standard context window
- **100K tokens**: Extended context (requires swapping between embedding/chat models due to RAM)

### Conference Talk Token Estimates

Using word count × 1.3 as token estimate (conservative for markdown content).

#### April 2025 Conference (34 talks)

| Session | Code | Talks | Est. Tokens | Fits 32K? |
|---------|------|-------|-------------|-----------|
| Saturday Morning | 1 | 8 | ~17,800 | ✅ Yes |
| Saturday Afternoon | 2 | 7 | ~19,800 | ✅ Yes |
| Priesthood Session | 3 | 5 | ~15,700 | ✅ Yes |
| Sunday Morning | 4 | 7 | ~18,500 | ✅ Yes |
| Sunday Afternoon | 5 | 7 | ~16,600 | ✅ Yes |
| **Full Conference** | - | 34 | **~88,300** | ❌ No (need 100K) |

#### Conference Size Over Time

| Year | Talks | Est. Tokens | Notes |
|------|-------|-------------|-------|
| 1975 | 47 | ~118,700 | More talks, shorter each |
| 1985 | 35 | ~77,600 | |
| 1995 | 43 | ~92,100 | |
| 2005 | 38 | ~90,400 | |
| 2015 | 39 | ~92,800 | |
| 2025 | 34 | ~88,300 | Fewer talks, longer each |

**Key Insight**: Conferences have gotten slightly smaller (fewer talks) but talks are longer. Full conference needs ~80K-120K tokens.

---

## Scripture Token Estimates

### Book of Mormon by Book

| Book | Chapters | Est. Tokens | Fits 32K? |
|------|----------|-------------|-----------|
| Alma | 63 | ~144,300 | ❌ No |
| 2 Nephi | 33 | ~60,900 | ❌ No |
| Mosiah | 29 | ~56,400 | ❌ No |
| 3 Nephi | 30 | ~53,000 | ❌ No |
| 1 Nephi | 22 | ~49,000 | ❌ No |
| Helaman | 16 | ~34,200 | ⚠️ Just over |
| Ether | 15 | ~28,300 | ✅ Yes |
| Mormon | 9 | ~17,100 | ✅ Yes |
| Jacob | 7 | ~16,900 | ✅ Yes |
| Moroni | 10 | ~12,100 | ✅ Yes |
| 4 Nephi | 1 | ~3,400 | ✅ Yes |
| Omni | 1 | ~2,500 | ✅ Yes |
| Enos | 1 | ~2,400 | ✅ Yes |
| Words of Mormon | 1 | ~1,700 | ✅ Yes |
| Jarom | 1 | ~1,500 | ✅ Yes |

**Full Book of Mormon**: 239 chapters, **~484,000 tokens** (far exceeds any current context)

### Other Volumes

| Volume | Chapters | Est. Tokens | Fits 32K? |
|--------|----------|-------------|-----------|
| Doctrine & Covenants | 138 | ~253,000 | ❌ No |
| Pearl of Great Price | 16 | ~51,500 | ❌ No (fits 100K) |
| Genesis | 50 | ~72,200 | ❌ No |
| Matthew | 28 | ~51,200 | ❌ No (fits 100K) |

### Manuals

| Manual | Files | Est. Tokens |
|--------|-------|-------------|
| Come Follow Me OT 2026 | 68 | ~130,500 |

---

## Hierarchical Summarization Strategy

Since full conferences and books exceed context limits, we need hierarchical summarization:

### Level 1: Individual Items (Current)
- **Talks**: Paragraph chunks + talk summary
- **Scriptures**: Verse/paragraph chunks + chapter summary

### Level 2: Session/Book Level (Proposed)
- Aggregate Level 1 summaries
- Fits in 32K context
- Extract session themes, doctrinal emphases

**Session Summary Input**: ~5-8 talk summaries × ~300 tokens each = ~1,500-2,400 tokens
**This comfortably fits in 32K!**

### Level 3: Conference/Volume Level (Proposed)
- Aggregate Level 2 summaries
- For conferences: 5 session summaries × ~500 tokens = ~2,500 tokens
- For BoM: 15 book summaries × ~500 tokens = ~7,500 tokens

---

## Proposed Index Enhancements

### New Metadata Fields for Talks

```go
// Enhanced TalkMetadata
type TalkMetadata struct {
    // ... existing fields ...
    
    // New doctrinal fields
    Principles    []string `json:"principles,omitempty"`    // Core principles taught
    Doctrines     []string `json:"doctrines,omitempty"`     // Doctrines referenced
    Programs      []string `json:"programs,omitempty"`      // Church programs/initiatives mentioned
    Invitations   []string `json:"invitations,omitempty"`   // Calls to action
    ScriptureRefs []string `json:"scripture_refs,omitempty"` // All scripture references
}
```

### New Layers

| Layer | Scope | Content |
|-------|-------|---------|
| `talk-summary` | Single talk | Current implementation |
| `session-summary` | 5-8 talks | Themes across a session |
| `conference-summary` | Full conference | Major themes, announcements, trends |
| `book-summary` | Scripture book | Narrative arc, key doctrines |
| `volume-summary` | Full volume | BoM, D&C, etc. overview |

### Example Session Summary Prompt

```
Analyze these General Conference session summaries and identify:
1. THEMES: Common doctrinal themes across talks (3-5)
2. PRINCIPLES: Key gospel principles taught
3. INVITATIONS: Calls to action given to members
4. PROGRAMS: Any church programs or initiatives referenced
5. TRENDS: How these talks relate to current events or church direction

Session: April 2025 Saturday Morning
Talks:
[Insert talk summaries]
```

---

## Benefits of Hierarchical Summarization

### For Conference Talks

1. **Trend Detection**: See how themes shift across years/decades
2. **Program Tracking**: Track when new initiatives are introduced
3. **Doctrinal Emphasis**: Identify which doctrines are being emphasized
4. **Cross-Reference**: Link conference themes to scripture study
5. **Research**: "What did the church emphasize in the 1990s vs 2020s?"

### For Scriptures

1. **Book Overview**: Quick summary of each book's purpose/narrative
2. **Theme Tracking**: See doctrinal themes across the BoM narrative
3. **Study Guides**: Auto-generate high-level study resources
4. **Cross-Volume**: Compare OT themes to BoM fulfillment

---

## Implementation Plan

### Phase 1: Session-Level Summaries (32K compatible)
1. After indexing talks, generate session summaries
2. Store as `conference-session-summary` layer
3. Input: Concatenated talk summaries for a session
4. Output: Session themes, key doctrines, invitations

### Phase 2: Conference-Level Summaries
1. Aggregate session summaries
2. Store as `conference-summary` layer
3. Track year-over-year trends

### Phase 3: Book-Level Scripture Summaries
1. Aggregate chapter summaries per book
2. Store as `book-summary` layer
3. Smaller books (≤32K) can be analyzed directly

### Phase 4: Enhanced Metadata Extraction
1. Update talk parser to extract:
   - Scripture references (already implemented)
   - Doctrines mentioned (keyword matching + LLM)
   - Programs/initiatives
   - Calls to action

---

## Practical Considerations

### RAM/Context Tradeoffs

| Task | Context Needed | Can Parallelize? |
|------|---------------|------------------|
| Talk summary | ~3K per talk | Yes, with embedding |
| Session summary | ~2.5K | Yes, after talk summaries |
| Conference summary | ~2.5K | Yes, after session summaries |
| Direct full-session analysis | ~15-20K | Yes, with 32K model |
| Direct full-conference | ~80-120K | Need 100K+ model |

### Incremental Processing

Since summaries are cached:
1. Index talks with summaries (already running)
2. Generate session summaries (post-processing)
3. Generate conference summaries (post-processing)

Each level only needs to run once per conference.

---

## Questions for Future Exploration

1. **Semantic Drift**: Do the same doctrinal terms mean different things across decades?
2. **Speaker Analysis**: Can we track individual apostles' themes over their ministry?
3. **Correlation**: Do conference themes predict Come Follow Me curriculum topics?
4. **Global vs Local**: Do regional conferences have different emphases?

---

## Related Files

- [00_TODO.md](00_TODO.md) - Active TODO list
- [03_content-indexing-guide.md](03_content-indexing-guide.md) - Content structure guide
- [talk_parser.go](../talk_parser.go) - Current talk parsing implementation
