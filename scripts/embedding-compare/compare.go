package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ---- Compare command ----

func runCompare(tagA, tagB string) error {
	fmt.Printf("Loading embeddings for %q and %q ...\n", tagA, tagB)

	a, err := loadEmbedResult(tagA)
	if err != nil {
		return fmt.Errorf("loading %q: %w", tagA, err)
	}
	b, err := loadEmbedResult(tagB)
	if err != nil {
		return fmt.Errorf("loading %q: %w", tagB, err)
	}

	fmt.Printf("  %s: model=%s, dims=%d, %d docs, %d queries\n", a.Tag, a.Model, a.Dims, len(a.Docs), len(a.Queries))
	fmt.Printf("  %s: model=%s, dims=%d, %d docs, %d queries\n", b.Tag, b.Model, b.Dims, len(b.Docs), len(b.Queries))

	// Warn if same model
	if a.Model == b.Model && a.Dims == b.Dims {
		fmt.Println("\n  ⚠ WARNING: Both tags used the SAME model and dimensions!")
		fmt.Println("  Results will be identical. Swap models in LM Studio between embed runs.")
		fmt.Println()
	}

	// Verify same doc set
	if len(a.Docs) != len(b.Docs) {
		return fmt.Errorf("doc count mismatch: %s has %d, %s has %d", tagA, len(a.Docs), tagB, len(b.Docs))
	}
	if len(a.Queries) != len(b.Queries) {
		return fmt.Errorf("query count mismatch: %s has %d, %s has %d", tagA, len(a.Queries), tagB, len(b.Queries))
	}

	// Build doc index by ID for each set
	aDocIdx := buildDocIndex(a.Docs)
	bDocIdx := buildDocIndex(b.Docs)

	// Separate docs by layer
	layers := []string{"summary", "verse"}
	aByLayer := groupByLayer(a.Docs)
	bByLayer := groupByLayer(b.Docs)

	// Run comparison for each query × each layer
	var allResults []QueryResult
	for i := range a.Queries {
		qText := a.Queries[i].Query.Text
		qCat := a.Queries[i].Query.Category
		aVec := a.Queries[i].Embedding
		bVec := b.Queries[i].Embedding

		for _, layer := range layers {
			aDocs := aByLayer[layer]
			bDocs := bByLayer[layer]
			if len(aDocs) == 0 || len(bDocs) == 0 {
				continue
			}

			aRanked := rankDocs(aVec, aDocs, aDocIdx)
			bRanked := rankDocs(bVec, bDocs, bDocIdx)

			qr := QueryResult{
				Query:    qText,
				Category: qCat,
				Layer:    layer,
			}
			qr.Top1Match = aRanked[0].ID == bRanked[0].ID
			qr.Top5Overlap = overlapCount(aRanked, bRanked, 5)
			qr.Top10Overlap = overlapCount(aRanked, bRanked, 10)
			qr.SpearmanRho = spearmanRho(aRanked, bRanked, min(10, len(aRanked)))

			// Use A's ranking as "relevance truth" and compute NDCG for B
			qr.NDCG10 = ndcg(aRanked, bRanked, 10)

			// Score deltas for top-10
			for j := 0; j < min(10, len(aRanked)); j++ {
				qr.AScores = append(qr.AScores, aRanked[j].Score)
			}
			for j := 0; j < min(10, len(bRanked)); j++ {
				qr.BScores = append(qr.BScores, bRanked[j].Score)
			}

			// Keep top-5 references for the report
			for j := 0; j < min(5, len(aRanked)); j++ {
				qr.ATop5 = append(qr.ATop5, aRanked[j].ID)
			}
			for j := 0; j < min(5, len(bRanked)); j++ {
				qr.BTop5 = append(qr.BTop5, bRanked[j].ID)
			}

			allResults = append(allResults, qr)
		}
	}

	// Compute ground truth precision
	gtResults := computeGroundTruth(a, b, aDocIdx, bDocIdx)

	// Compute perturbation stability
	pertResults := computePerturbationStability(allResults)

	// Generate report
	report := generateReport(a, b, allResults, layers, gtResults, pertResults)

	outPath := filepath.Join("data", "report.md")
	if err := os.MkdirAll("data", 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(outPath, []byte(report), 0o644); err != nil {
		return err
	}

	fmt.Printf("\nReport written to %s\n", outPath)

	// Print summary to stdout
	printSummary(allResults, layers)

	return nil
}

// ---- Types ----

type QueryResult struct {
	Query        string
	Category     string
	Layer        string
	Top1Match    bool
	Top5Overlap  int
	Top10Overlap int
	SpearmanRho  float64
	NDCG10       float64 // Normalized Discounted Cumulative Gain @ 10
	AScores      []float32
	BScores      []float32
	ATop5        []string
	BTop5        []string
}

type RankedDoc struct {
	ID    string
	Score float32
	Rank  int
}

// ---- Core math ----

// dotProduct computes dot product of two float32 slices.
// Vectors should be pre-normalized for this to equal cosine similarity,
// but we normalize here defensively since JSON round-tripping may lose precision.
func dotProduct(a, b []float32) float32 {
	// If dimensions differ, use the shorter one (for cross-dimension comparison)
	n := min(len(a), len(b))
	var sum float32
	for i := 0; i < n; i++ {
		sum += a[i] * b[i]
	}
	return sum
}

// cosineSimilarity computes cosine similarity, handling different dimensions.
func cosineSimilarity(a, b []float32) float32 {
	n := min(len(a), len(b))
	var dot, normA, normB float32
	for i := 0; i < n; i++ {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

// rankDocs computes similarity of a query vector against all docs, returns sorted.
func rankDocs(queryVec []float32, docIDs []string, docIdx map[string][]float32) []RankedDoc {
	ranked := make([]RankedDoc, len(docIDs))
	for i, id := range docIDs {
		docVec := docIdx[id]
		ranked[i] = RankedDoc{
			ID:    id,
			Score: cosineSimilarity(queryVec, docVec),
		}
	}
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Score > ranked[j].Score
	})
	for i := range ranked {
		ranked[i].Rank = i + 1
	}
	return ranked
}

// overlapCount counts how many IDs appear in both A's and B's top-K.
func overlapCount(a, b []RankedDoc, k int) int {
	if k > len(a) {
		k = len(a)
	}
	if k > len(b) {
		k = len(b)
	}
	bSet := make(map[string]bool, k)
	for i := 0; i < k; i++ {
		bSet[b[i].ID] = true
	}
	count := 0
	for i := 0; i < k; i++ {
		if bSet[a[i].ID] {
			count++
		}
	}
	return count
}

// spearmanRho computes Spearman rank correlation for the top-K items.
// Uses ranks within A and B for items that appear in both top-K sets.
func spearmanRho(a, b []RankedDoc, k int) float64 {
	if k > len(a) {
		k = len(a)
	}
	if k > len(b) {
		k = len(b)
	}

	// Build rank maps for all items in both top-K
	aRank := make(map[string]int, k)
	for i := 0; i < k; i++ {
		aRank[a[i].ID] = i + 1
	}
	bRank := make(map[string]int, k)
	for i := 0; i < k; i++ {
		bRank[b[i].ID] = i + 1
	}

	// Union of items in both top-K
	allIDs := make(map[string]bool)
	for id := range aRank {
		allIDs[id] = true
	}
	for id := range bRank {
		allIDs[id] = true
	}

	n := len(allIDs)
	if n < 2 {
		return 1.0
	}

	// For items missing from one list, assign rank k+1 (worst)
	var sumD2 float64
	for id := range allIDs {
		ra, ok := aRank[id]
		if !ok {
			ra = k + 1
		}
		rb, ok := bRank[id]
		if !ok {
			rb = k + 1
		}
		d := float64(ra - rb)
		sumD2 += d * d
	}

	// Spearman: ρ = 1 - (6 * Σd²) / (n * (n² - 1))
	rho := 1.0 - (6.0*sumD2)/(float64(n)*(float64(n*n)-1.0))
	return rho
}

// ndcg computes Normalized Discounted Cumulative Gain at K.
// Uses A's ranking as ground truth relevance: items ranked higher by A get
// higher relevance scores. Measures how well B reproduces A's ranking.
func ndcg(aRanked, bRanked []RankedDoc, k int) float64 {
	if k > len(aRanked) {
		k = len(aRanked)
	}
	if k > len(bRanked) {
		k = len(bRanked)
	}

	// Build relevance map from A's ranking: relevance = k - rank + 1
	// (top-ranked item gets highest relevance)
	relevance := make(map[string]float64, k)
	for i := 0; i < k; i++ {
		relevance[aRanked[i].ID] = float64(k - i)
	}

	// DCG for B's ranking
	var dcg float64
	for i := 0; i < k; i++ {
		rel := relevance[bRanked[i].ID] // 0 if not in A's top-K
		dcg += rel / math.Log2(float64(i+2))
	}

	// Ideal DCG (A's own ranking)
	var idcg float64
	for i := 0; i < k; i++ {
		rel := float64(k - i)
		idcg += rel / math.Log2(float64(i+2))
	}

	if idcg == 0 {
		return 1.0
	}
	return dcg / idcg
}

// ---- Ground truth evaluation ----

type GTResult struct {
	Query      string
	AHits      int // how many expected chapters A found in top-3
	BHits      int
	ATotal     int // total expected
	BTotal     int
	APrecision float64
	BPrecision float64
}

func computeGroundTruth(a, b *EmbedResult, aDocIdx, bDocIdx map[string][]float32) []GTResult {
	gts := groundTruths()
	if len(gts) == 0 {
		return nil
	}

	// Find query embeddings by text
	aQueryMap := make(map[string][]float32)
	for _, q := range a.Queries {
		aQueryMap[q.Query.Text] = q.Embedding
	}
	bQueryMap := make(map[string][]float32)
	for _, q := range b.Queries {
		bQueryMap[q.Query.Text] = q.Embedding
	}

	// Get summary doc IDs
	var summaryIDs []string
	for _, d := range a.Docs {
		if d.Layer == "summary" {
			summaryIDs = append(summaryIDs, d.ID)
		}
	}

	var results []GTResult
	for _, gt := range gts {
		aVec, okA := aQueryMap[gt.Query]
		bVec, okB := bQueryMap[gt.Query]
		if !okA || !okB {
			continue
		}

		aRanked := rankDocs(aVec, summaryIDs, aDocIdx)
		bRanked := rankDocs(bVec, summaryIDs, bDocIdx)

		// Check how many expected chapters appear in top-3
		aHits := countGTHits(aRanked, gt.ExpectedSummary, 3)
		bHits := countGTHits(bRanked, gt.ExpectedSummary, 3)

		results = append(results, GTResult{
			Query:      gt.Query,
			AHits:      aHits,
			BHits:      bHits,
			ATotal:     len(gt.ExpectedSummary),
			BTotal:     len(gt.ExpectedSummary),
			APrecision: float64(aHits) / float64(min(3, len(gt.ExpectedSummary))),
			BPrecision: float64(bHits) / float64(min(3, len(gt.ExpectedSummary))),
		})
	}
	return results
}

func countGTHits(ranked []RankedDoc, expectedChapters []int, k int) int {
	if k > len(ranked) {
		k = len(ranked)
	}
	expected := make(map[string]bool)
	for _, ch := range expectedChapters {
		expected[fmt.Sprintf("1ne-%d-summary", ch)] = true
	}
	count := 0
	for i := 0; i < k; i++ {
		if expected[ranked[i].ID] {
			count++
		}
	}
	return count
}

// ---- Perturbation stability ----
// Compares KJV-phrased queries with their modern-phrased equivalents.
// If both models rank the same chapters similarly regardless of phrasing,
// they're equally robust. If one is more sensitive to phrasing, that matters.

type PerturbResult struct {
	KJVQuery    string
	ModernQuery string
	AOverlap    int // top-5 overlap between KJV and modern results for model A
	BOverlap    int
}

func computePerturbationStability(results []QueryResult) []PerturbResult {
	// KJV and modern-phrasing queries are paired by index within their categories
	kjvResults := make(map[string][]QueryResult)
	modResults := make(map[string][]QueryResult)
	for _, r := range results {
		if r.Layer != "summary" {
			continue
		}
		switch r.Category {
		case "kjv-phrasing":
			kjvResults[r.Query] = append(kjvResults[r.Query], r)
		case "modern-phrasing":
			modResults[r.Query] = append(modResults[r.Query], r)
		}
	}

	// Collect KJV and modern queries in order
	var kjvQueries, modQueries []string
	for _, r := range results {
		if r.Layer == "summary" && r.Category == "kjv-phrasing" {
			if len(kjvQueries) == 0 || kjvQueries[len(kjvQueries)-1] != r.Query {
				kjvQueries = append(kjvQueries, r.Query)
			}
		}
		if r.Layer == "summary" && r.Category == "modern-phrasing" {
			if len(modQueries) == 0 || modQueries[len(modQueries)-1] != r.Query {
				modQueries = append(modQueries, r.Query)
			}
		}
	}

	n := min(len(kjvQueries), len(modQueries))
	var out []PerturbResult
	for i := 0; i < n; i++ {
		kjv := kjvResults[kjvQueries[i]]
		mod := modResults[modQueries[i]]
		if len(kjv) == 0 || len(mod) == 0 {
			continue
		}
		// Compare top-5 overlap between KJV and modern for the same position
		// (these are different queries but same concept — paired by index)
		aOverlap := stringOverlap(kjv[0].ATop5, mod[0].ATop5)
		bOverlap := stringOverlap(kjv[0].BTop5, mod[0].BTop5)
		out = append(out, PerturbResult{
			KJVQuery:    kjvQueries[i],
			ModernQuery: modQueries[i],
			AOverlap:    aOverlap,
			BOverlap:    bOverlap,
		})
	}
	return out
}

func stringOverlap(a, b []string) int {
	set := make(map[string]bool, len(b))
	for _, s := range b {
		set[s] = true
	}
	count := 0
	for _, s := range a {
		if set[s] {
			count++
		}
	}
	return count
}

// ---- Score distribution helpers ----

type ScoreDist struct {
	Mean   float64
	Median float64
	StdDev float64
	Min    float64
	Max    float64
	Spread float64 // Max - Min of top-10
}

// ---- Index helpers ----

func buildDocIndex(docs []DocEmbedding) map[string][]float32 {
	idx := make(map[string][]float32, len(docs))
	for _, d := range docs {
		idx[d.ID] = d.Embedding
	}
	return idx
}

func groupByLayer(docs []DocEmbedding) map[string][]string {
	m := make(map[string][]string)
	for _, d := range docs {
		m[d.Layer] = append(m[d.Layer], d.ID)
	}
	return m
}

// ---- I/O ----

func loadEmbedResult(tag string) (*EmbedResult, error) {
	path := filepath.Join("data", tag, "embeddings.json")
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()

	var result EmbedResult
	if err := json.NewDecoder(f).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding %s: %w", path, err)
	}
	return &result, nil
}

// ---- Report generation ----

func generateReport(a, b *EmbedResult, results []QueryResult, layers []string, gtResults []GTResult, pertResults []PerturbResult) string {
	var sb strings.Builder

	sb.WriteString("# Embedding Comparison Report\n\n")
	sb.WriteString(fmt.Sprintf("**Model A:** %s (tag: %s, dims: %d)\n", a.Model, a.Tag, a.Dims))
	sb.WriteString(fmt.Sprintf("**Model B:** %s (tag: %s, dims: %d)\n", b.Model, b.Tag, b.Dims))
	sb.WriteString(fmt.Sprintf("**Corpus:** 1 Nephi — %d documents, %d queries\n", len(a.Docs), len(a.Queries)))
	sb.WriteString(fmt.Sprintf("**A embed time:** %s | **B embed time:** %s\n\n", a.EmbedTime, b.EmbedTime))

	// Same-model warning
	if a.Model == b.Model && a.Dims == b.Dims {
		sb.WriteString("> ⚠ **WARNING:** Both tags used the SAME model and dimensions. Results will be\n")
		sb.WriteString("> identical. Swap models in LM Studio between embed runs and re-run.\n\n")
	}

	// Overall summary by layer
	sb.WriteString("## Summary by Layer\n\n")
	sb.WriteString("| Layer | Avg Top-5 | Avg Top-10 | Avg ρ | Avg NDCG@10 | Top-1 Agree |\n")
	sb.WriteString("|-------|----------:|-----------:|------:|------------:|------------:|\n")

	for _, layer := range layers {
		layerResults := filterByLayer(results, layer)
		if len(layerResults) == 0 {
			continue
		}

		avgT5 := avgFloat(mapInt(layerResults, func(r QueryResult) int { return r.Top5Overlap }), 5)
		avgT10 := avgFloat(mapInt(layerResults, func(r QueryResult) int { return r.Top10Overlap }), 10)
		avgRho := avgRhoF(layerResults)
		avgNDCG := avgNDCGF(layerResults)
		top1Pct := top1Pct(layerResults)

		sb.WriteString(fmt.Sprintf("| %s | %.1f/5 (%.0f%%) | %.1f/10 (%.0f%%) | %.3f | %.3f | %.0f%% |\n",
			layer,
			avgT5, avgT5/5*100,
			avgT10, avgT10/10*100,
			avgRho,
			avgNDCG,
			top1Pct))
	}

	// Verdict
	sb.WriteString("\n## Verdict\n\n")
	allAvgT10 := avgFloat(mapInt(results, func(r QueryResult) int { return r.Top10Overlap }), 10)
	allAvgRho := avgRhoF(results)
	allAvgNDCG := avgNDCGF(results)

	t10Pct := allAvgT10 / 10 * 100
	rankEquivalent := t10Pct >= 80 && allAvgRho >= 0.85

	sb.WriteString(fmt.Sprintf("- **Overall Top-10 Overlap:** %.1f/10 (%.0f%%) — threshold: ≥80%%\n", allAvgT10, t10Pct))
	sb.WriteString(fmt.Sprintf("- **Overall Spearman ρ:** %.3f — threshold: ≥0.85\n", allAvgRho))
	sb.WriteString(fmt.Sprintf("- **Overall NDCG@10:** %.3f\n", allAvgNDCG))

	// Ground truth summary for verdict
	var aGTPct, bGTPct float64
	if len(gtResults) > 0 {
		var aTotal, bTotal, gtTotal int
		for _, gt := range gtResults {
			aTotal += gt.AHits
			bTotal += gt.BHits
			gtTotal += min(3, gt.ATotal)
		}
		aGTPct = float64(aTotal) / float64(gtTotal) * 100
		bGTPct = float64(bTotal) / float64(gtTotal) * 100
		sb.WriteString(fmt.Sprintf("- **Ground Truth Precision:** A=%.0f%%, B=%.0f%%\n", aGTPct, bGTPct))
	}

	// Perturbation summary for verdict
	var aPertAvg, bPertAvg float64
	if len(pertResults) > 0 {
		var aSum, bSum int
		for _, p := range pertResults {
			aSum += p.AOverlap
			bSum += p.BOverlap
		}
		aPertAvg = float64(aSum) / float64(len(pertResults))
		bPertAvg = float64(bSum) / float64(len(pertResults))
		sb.WriteString(fmt.Sprintf("- **Perturbation Stability:** A=%.1f/5, B=%.1f/5\n", aPertAvg, bPertAvg))
	}
	sb.WriteString("\n")

	if rankEquivalent {
		sb.WriteString("**VERDICT: Models are equivalent.** The smaller model produces search results that are\n")
		sb.WriteString("statistically indistinguishable from the larger model on this corpus. The VRAM savings\n")
		sb.WriteString("are justified.\n")
	} else {
		// Different rankings, but check if the larger model actually wins on quality
		bWinsGT := bGTPct > aGTPct+5         // B needs >5pp better ground truth
		bWinsPert := bPertAvg > aPertAvg+0.5 // B needs >0.5 better perturbation
		bWinsOverall := bWinsGT && bWinsPert

		if bWinsOverall {
			sb.WriteString("**VERDICT: Larger model is meaningfully better.** Model B produces different rankings\n")
			sb.WriteString("AND wins on both ground truth precision and perturbation stability. The VRAM cost\n")
			sb.WriteString("may be justified.\n")
		} else {
			sb.WriteString("**VERDICT: Different but not better.** The models produce different rankings, but the\n")
			sb.WriteString("larger model does not consistently outperform on quality metrics (ground truth precision\n")
			sb.WriteString("and perturbation stability). The smaller model is the better value.\n")
		}
	}

	// Per-category breakdown
	sb.WriteString("\n## By Category\n\n")
	categories := []string{"factual", "thematic", "christological", "cross-reference", "kjv-phrasing", "modern-phrasing", "short", "multi-hop"}
	for _, cat := range categories {
		catResults := filterByCategory(results, cat)
		if len(catResults) == 0 {
			continue
		}
		avgT10 := avgFloat(mapInt(catResults, func(r QueryResult) int { return r.Top10Overlap }), 10)
		avgRho := avgRhoF(catResults)
		avgNDCG := avgNDCGF(catResults)

		sb.WriteString(fmt.Sprintf("### %s (avg top-10: %.0f%%, ρ: %.3f, NDCG: %.3f)\n\n", cat, avgT10/10*100, avgRho, avgNDCG))
		sb.WriteString("| Query | Layer | Top-5 | Top-10 | ρ | NDCG | Top-1 |\n")
		sb.WriteString("|-------|-------|------:|-------:|--:|-----:|------:|\n")

		for _, r := range results {
			if r.Category != cat {
				continue
			}
			top1 := "✗"
			if r.Top1Match {
				top1 = "✓"
			}
			qShort := r.Query
			if len(qShort) > 45 {
				qShort = qShort[:42] + "..."
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %d/5 | %d/10 | %.2f | %.2f | %s |\n",
				qShort, r.Layer, r.Top5Overlap, r.Top10Overlap, r.SpearmanRho, r.NDCG10, top1))
		}
		sb.WriteString("\n")
	}

	// Ground truth precision
	if len(gtResults) > 0 {
		sb.WriteString("## Ground Truth Precision\n\n")
		sb.WriteString("Queries with known correct chapters. Checks if the expected chapters appear in the top-3 summary results.\n\n")
		sb.WriteString(fmt.Sprintf("| Query | Expected | %s hits | %s hits | Same? |\n", a.Tag, b.Tag))
		sb.WriteString("|-------|:--------:|:-------:|:-------:|:-----:|\n")

		var aTotal, bTotal, gtTotal int
		for _, gt := range gtResults {
			same := "✓"
			if gt.AHits != gt.BHits {
				same = "✗"
			}
			qShort := gt.Query
			if len(qShort) > 50 {
				qShort = qShort[:47] + "..."
			}
			sb.WriteString(fmt.Sprintf("| %s | %d | %d/3 | %d/3 | %s |\n",
				qShort, gt.ATotal, gt.AHits, gt.BHits, same))
			aTotal += gt.AHits
			bTotal += gt.BHits
			gtTotal += min(3, gt.ATotal)
		}
		sb.WriteString(fmt.Sprintf("\n**%s overall:** %d/%d (%.0f%%) | **%s overall:** %d/%d (%.0f%%)\n\n",
			a.Tag, aTotal, gtTotal, float64(aTotal)/float64(gtTotal)*100,
			b.Tag, bTotal, gtTotal, float64(bTotal)/float64(gtTotal)*100))
	}

	// Perturbation stability (KJV vs modern phrasing)
	if len(pertResults) > 0 {
		sb.WriteString("## Perturbation Stability (KJV vs Modern Phrasing)\n\n")
		sb.WriteString("Same concept expressed in KJV and modern language. Top-5 overlap shows how\n")
		sb.WriteString("consistently each model retrieves the same chapters regardless of phrasing.\n\n")
		sb.WriteString(fmt.Sprintf("| KJV Query | Modern Query | %s overlap | %s overlap |\n", a.Tag, b.Tag))
		sb.WriteString("|-----------|-------------|:----------:|:----------:|\n")

		var aTotalOverlap, bTotalOverlap int
		for _, p := range pertResults {
			kjvShort := p.KJVQuery
			if len(kjvShort) > 35 {
				kjvShort = kjvShort[:32] + "..."
			}
			modShort := p.ModernQuery
			if len(modShort) > 35 {
				modShort = modShort[:32] + "..."
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %d/5 | %d/5 |\n",
				kjvShort, modShort, p.AOverlap, p.BOverlap))
			aTotalOverlap += p.AOverlap
			bTotalOverlap += p.BOverlap
		}
		n := len(pertResults)
		sb.WriteString(fmt.Sprintf("\n**%s avg:** %.1f/5 | **%s avg:** %.1f/5\n\n",
			a.Tag, float64(aTotalOverlap)/float64(n),
			b.Tag, float64(bTotalOverlap)/float64(n)))

		sb.WriteString("Higher = more robust to phrasing variation. If one model is significantly more\n")
		sb.WriteString("stable, it handles diverse user queries better.\n\n")
	}

	// Score distribution analysis
	sb.WriteString("## Score Distribution\n\n")
	sb.WriteString("Average cosine similarity scores for top-10 results. Higher spread = more\n")
	sb.WriteString("discriminative (clearer separation between relevant and irrelevant).\n\n")
	sb.WriteString(fmt.Sprintf("| Layer | %s mean | %s mean | %s spread | %s spread |\n", a.Tag, b.Tag, a.Tag, b.Tag))
	sb.WriteString("|-------|-------:|-------:|---------:|---------:|\n")
	for _, layer := range layers {
		lr := filterByLayer(results, layer)
		if len(lr) == 0 {
			continue
		}
		aDist := aggregateScores(lr, true)
		bDist := aggregateScores(lr, false)
		sb.WriteString(fmt.Sprintf("| %s | %.4f | %.4f | %.4f | %.4f |\n",
			layer, aDist.Mean, bDist.Mean, aDist.Spread, bDist.Spread))
	}
	sb.WriteString("\n")

	// Detailed top-5 comparison for queries where models disagree most
	sb.WriteString("## Biggest Disagreements (lowest Top-10 overlap)\n\n")
	sorted := make([]QueryResult, len(results))
	copy(sorted, results)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Top10Overlap < sorted[j].Top10Overlap
	})
	for i := 0; i < min(5, len(sorted)); i++ {
		r := sorted[i]
		sb.WriteString(fmt.Sprintf("**%s** (%s, %s) — Top-10 overlap: %d/10\n\n", r.Query, r.Layer, r.Category, r.Top10Overlap))
		sb.WriteString(fmt.Sprintf("  %s top-5: %s\n", a.Tag, strings.Join(r.ATop5, ", ")))
		sb.WriteString(fmt.Sprintf("  %s top-5: %s\n\n", b.Tag, strings.Join(r.BTop5, ", ")))
	}

	return sb.String()
}

func printSummary(results []QueryResult, layers []string) {
	fmt.Println("\n=== SUMMARY ===")
	for _, layer := range layers {
		lr := filterByLayer(results, layer)
		if len(lr) == 0 {
			continue
		}
		avgT10 := avgFloat(mapInt(lr, func(r QueryResult) int { return r.Top10Overlap }), 10)
		avgRho := avgRhoF(lr)
		avgNDCG := avgNDCGF(lr)
		fmt.Printf("  %s: avg top-10 overlap = %.1f/10 (%.0f%%), avg ρ = %.3f, avg NDCG@10 = %.3f\n",
			layer, avgT10, avgT10/10*100, avgRho, avgNDCG)
	}

	allAvgT10 := avgFloat(mapInt(results, func(r QueryResult) int { return r.Top10Overlap }), 10)
	allAvgRho := avgRhoF(results)
	allAvgNDCG := avgNDCGF(results)
	t10Pct := allAvgT10 / 10 * 100

	fmt.Printf("\n  OVERALL: top-10 overlap = %.0f%%, ρ = %.3f, NDCG@10 = %.3f\n", t10Pct, allAvgRho, allAvgNDCG)
	if t10Pct >= 80 && allAvgRho >= 0.85 {
		fmt.Println("  VERDICT: Equivalent — smaller model is fine")
	} else {
		fmt.Println("  VERDICT: Meaningfully different — check report for details")
	}
}

// ---- Aggregation helpers ----

func filterByLayer(results []QueryResult, layer string) []QueryResult {
	var out []QueryResult
	for _, r := range results {
		if r.Layer == layer {
			out = append(out, r)
		}
	}
	return out
}

func mapInt(results []QueryResult, fn func(QueryResult) int) []int {
	out := make([]int, len(results))
	for i, r := range results {
		out[i] = fn(r)
	}
	return out
}

func avgFloat(values []int, maxVal int) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0
	for _, v := range values {
		sum += v
	}
	return float64(sum) / float64(len(values))
}

func avgRhoF(results []QueryResult) float64 {
	if len(results) == 0 {
		return 0
	}
	var sum float64
	for _, r := range results {
		sum += r.SpearmanRho
	}
	return sum / float64(len(results))
}

func top1Pct(results []QueryResult) float64 {
	if len(results) == 0 {
		return 0
	}
	count := 0
	for _, r := range results {
		if r.Top1Match {
			count++
		}
	}
	return float64(count) / float64(len(results)) * 100
}

func avgNDCGF(results []QueryResult) float64 {
	if len(results) == 0 {
		return 0
	}
	var sum float64
	for _, r := range results {
		sum += r.NDCG10
	}
	return sum / float64(len(results))
}

func filterByCategory(results []QueryResult, cat string) []QueryResult {
	var out []QueryResult
	for _, r := range results {
		if r.Category == cat {
			out = append(out, r)
		}
	}
	return out
}

func aggregateScores(results []QueryResult, useA bool) ScoreDist {
	var allScores []float64
	for _, r := range results {
		scores := r.BScores
		if useA {
			scores = r.AScores
		}
		for _, s := range scores {
			allScores = append(allScores, float64(s))
		}
	}
	if len(allScores) == 0 {
		return ScoreDist{}
	}
	sort.Float64s(allScores)

	var sum float64
	for _, s := range allScores {
		sum += s
	}
	mean := sum / float64(len(allScores))

	var variance float64
	for _, s := range allScores {
		d := s - mean
		variance += d * d
	}
	variance /= float64(len(allScores))

	return ScoreDist{
		Mean:   mean,
		Median: allScores[len(allScores)/2],
		StdDev: math.Sqrt(variance),
		Min:    allScores[0],
		Max:    allScores[len(allScores)-1],
		Spread: allScores[len(allScores)-1] - allScores[0],
	}
}
