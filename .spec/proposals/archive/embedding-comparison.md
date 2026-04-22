# Proposal: Qwen3 Embedding 4B vs 8B Comparison Tool

## Binding Problem

Michael is spending ~5GB VRAM on Qwen3-Embedding-8B (Q4_K_M) when 4B uses ~2.5GB. He's considering removing his second GPU entirely. Before making that hardware decision, he needs data on whether 8B actually buys meaningfully better search results on *our* scripture data — not just MTEB benchmarks. If 4B is within ~5% quality, the VRAM isn't worth it.

## Prior Art

- **MTEB benchmarks** (from `.spec/scratch/lm-studio-model-experiments/main.md`): 4B = 74.60 English, 8B = 75.22 English. That's 0.8% on benchmarks. Multilingual: 69.45 vs 70.58 (~1.6%). Suggestive but not conclusive for domain-specific scripture content.
- **chromem-exp** (`scripts/chromem-exp/`): Already has a `compare` experiment mode using chromem-go + LM Studio. Good reference but uses chromem-go directly, not the vecf/mmap path.
- **gospel-engine embedder** (`internal/vec/embedder.go`): Simple HTTP wrapper around LM Studio `/v1/embeddings`. Model name is configurable via env var.
- **Existing enrichment data**: 1,584 scripture chapters already enriched in `gospel.db` with summaries, keywords, Christ types. 1 Nephi = 22 chapters.
- **Dimension difference**: 4B native = 2560, 8B native = 4096. Both support MRL (configurable dimensions). This alone changes storage cost 1.6x.

## Success Criteria

1. A quantitative comparison showing top-K overlap, rank correlation, and score distributions for both models on the same queries against the same corpus.
2. A clear recommendation: stick with 8B, downgrade to 4B, or "it depends on query type."
3. The tool is reusable if we ever want to test other embedding models.

## Constraints

- Must use 1 Nephi enhanced content (22 chapters of enrichment summaries) as the test corpus.
- Must test at multiple search granularities: summary-level AND verse/paragraph-level.
- Must use LM Studio's `/v1/embeddings` endpoint (same as production).
- Can only run one embedding model at a time in LM Studio — sequential, not parallel.
- Should run in under 10 minutes per model (1 Ne is small).

## Proposed Approach

### Architecture: Standalone Go CLI tool

Location: `scripts/embedding-compare/` (standalone, not inside gospel-engine).

Reuses:
- The LM Studio embedder pattern from `internal/vec/embedder.go`
- The `.vecf` format from gospel-engine for storage
- Test queries hand-picked for our domain

Does NOT reuse:
- gospel-engine's full pipeline (too heavy for a test tool)
- chromem-go (go straight to vecf for simplicity)

### Data Source

Pull from `gospel.db` (the existing enriched database):
- **Summary layer**: 22 enrichment summaries (same content that gets embedded as `scriptures-summary` in production)
- **Verse layer**: All individual verses from 1 Nephi (~600+ verses)
- **Paragraph layer**: The existing paragraph chunks from 1 Nephi

This gives us three granularity levels to see if model size matters more or less at different content lengths.

### Test Queries (20 queries across 4 categories)

**Factual/Specific** (should have clear "right answers"):
1. "Lehi's dream of the tree of life"
2. "brass plates of Laban"
3. "Nephi breaks his bow"
4. "Liahona compass"
5. "ship building in Bountiful"

**Thematic/Conceptual** (harder — tests semantic understanding):
6. "faith and obedience to God"
7. "God's love for his children"
8. "following the prophet"
9. "the power of the word of God"
10. "family conflict and forgiveness"

**Christological** (tests scriptural depth):
11. "types and shadows of Christ"
12. "the Messiah will redeem his people"
13. "Lamb of God"
14. "baptism and remission of sins"
15. "the tree of life as God's love"

**Cross-reference style** (queries that reference other scripture):
16. "Isaiah's prophecy of the last days"
17. "the scattering and gathering of Israel"
18. "the plan of salvation"
19. "priesthood authority"
20. "the Holy Ghost as a guide"

### Metrics

For each query at each granularity:

| Metric | What it measures | Threshold |
|--------|-----------------|-----------|
| **Top-10 Overlap** | How many of the same docs appear in both models' top 10 | ≥8/10 = equivalent |
| **Top-5 Overlap** | Tighter check on the most-relevant results | ≥4/5 = equivalent |
| **Rank Correlation (Spearman ρ)** | Do they rank results in the same order? | ≥0.85 = equivalent |
| **Score Delta** | Average absolute difference in cosine similarity scores | Informational |
| **Top-1 Agreement** | Do they agree on the #1 result? | Informational |

**Overall verdict**: If average Top-10 Overlap ≥ 80% and average Spearman ρ ≥ 0.85, 4B is "equivalent enough" and 8B isn't worth the VRAM.

### Workflow (3 steps)

**Step 1: Embed with 4B** (load 4B in LM Studio, then run)
```
embedding-compare embed --tag=4b --db=../../scripts/gospel-engine/data/gospel.db
```
- Reads 1 Ne summaries, verses, paragraphs from gospel.db
- Embeds each with the currently-loaded LM Studio model
- Embeds all 20 test queries
- Saves everything to `data/4b/` as JSON (embeddings + metadata)
- Records: model name, dimension, embed time per doc

**Step 2: Embed with 8B** (swap model in LM Studio, then run)
```
embedding-compare embed --tag=8b --db=../../scripts/gospel-engine/data/gospel.db
```
- Same process, saves to `data/8b/`

**Step 3: Compare** (no model needed)
```
embedding-compare compare --a=4b --b=8b
```
- Loads both embedding sets
- For each query × each layer: compute top-K, overlap, rank correlation
- Outputs a markdown report to `data/report.md`
- Summary table + per-query breakdown + recommendation

### Why JSON instead of .vecf?

For a 22-chapter + ~600-verse test corpus, JSON is simpler and more debuggable than binary vecf. The vecf format shines at 200K+ vectors with mmap. Here we need human-readable output more than we need performance.

### Dimension Handling

Two comparison modes:
1. **Native dimensions** — 4B at 2560, 8B at 4096. This is what production uses. Apples-to-oranges on dimension but shows the real-world difference.
2. **Matched dimensions** — Both at 2560 (4B's max). Isolates quality from dimension count. Uses LM Studio's `dimensions` parameter.

The tool runs both by default and reports both.

## Phased Delivery

### Phase 1: Build the tool (one session)

- `cmd/main.go` — CLI with `embed` and `compare` commands
- `embed.go` — LM Studio embedder, gospel.db reader, JSON writer
- `compare.go` — Load two sets, compute metrics, generate report
- `queries.go` — The 20 test queries
- `go.mod` — standalone module (only needs `database/sql` + `mattn/go-sqlite3`)

### Phase 2: Run the experiment (manual, ~20 min)

1. Load 4B in LM Studio → run embed → note VRAM usage
2. Load 8B in LM Studio → run embed → note VRAM usage
3. Run compare → read report → decide

### Phase 3: Act on results

- If 4B wins: re-embed gospel-engine with 4B, update config defaults, reclaim VRAM
- If 8B wins by >5%: keep 8B, close the question
- If it's a wash: default to 4B (cheaper resource usage wins ties)

## Costs & Risks

- **Build time**: ~1 session for the tool. It's a focused CLI, not infrastructure.
- **Run time**: ~5-10 min per model for 1 Ne corpus.
- **Risk**: LM Studio's `dimensions` parameter may not work correctly (noted in prior scratch as issue #101). If matched-dimension mode fails, native-only comparison is still valid.
- **VRAM note**: Need to unload one model and load the other between steps. LM Studio handles this.

## Recommendation

**Build it.** This is a small, focused tool that answers a real hardware question with real data. The MTEB numbers suggest 4B is probably fine (0.8% difference), but "probably" isn't a basis for a hardware decision. One session to build, 20 minutes to run, and the question is closed forever.
