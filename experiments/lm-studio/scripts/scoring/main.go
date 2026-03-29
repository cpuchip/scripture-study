package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

const dbFile = "scoring.db"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "init":
		cmdInit()
	case "import":
		cmdImport(os.Args[2:])
	case "stats":
		cmdStats(os.Args[2:])
	case "compare":
		cmdCompare(os.Args[2:])
	case "gt":
		cmdGT(os.Args[2:])
	case "merge":
		cmdMerge(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `Usage: scoring <command> [args]

Commands:
  init                  Create database and seed ground truth
  import <tag> [dir]    Import result JSON files (default dir: ../results)
  stats <tag>           Show statistics for a prompt version
  compare <tag1,tag2..> Compare multiple versions side-by-side
  gt [set|list]         Manage ground truth scores
    gt list             List all ground truth scores
    gt set <content> <dim> <score> [<hi>]  Set a ground truth score (range if hi given)
  merge <from> <into>   Copy scores from one tag into another (fills gaps)`)
}

// --- Database ---

func openDB() *sql.DB {
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		fatal("open db: %v", err)
	}
	// Enable WAL mode for better concurrent access
	db.Exec("PRAGMA journal_mode=WAL")
	return db
}

func cmdInit() {
	db := openDB()
	defer db.Close()

	schema := `
CREATE TABLE IF NOT EXISTS ground_truth (
    content TEXT NOT NULL,
    dimension TEXT NOT NULL,
    score_lo REAL NOT NULL,
    score_hi REAL NOT NULL,
    PRIMARY KEY (content, dimension)
);

CREATE TABLE IF NOT EXISTS results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tag TEXT NOT NULL,
    content TEXT NOT NULL,
    dimension TEXT NOT NULL,
    score INTEGER NOT NULL,
    model TEXT NOT NULL DEFAULT '',
    timestamp TEXT NOT NULL DEFAULT '',
    tokens_in INTEGER NOT NULL DEFAULT 0,
    tokens_out INTEGER NOT NULL DEFAULT 0,
    ttft_ms INTEGER NOT NULL DEFAULT 0,
    latency_ms INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_results_tag ON results(tag);
CREATE INDEX IF NOT EXISTS idx_results_content ON results(content);
CREATE UNIQUE INDEX IF NOT EXISTS idx_results_unique ON results(tag, content, dimension);
`
	_, err := db.Exec(schema)
	if err != nil {
		fatal("create schema: %v", err)
	}

	// Seed ground truth
	gt := []struct {
		content   string
		dimension string
		lo, hi    float64
	}{
		// Alma 32
		{"alma-32", "teach", 7, 8},
		{"alma-32", "help", 8, 8},
		{"alma-32", "love", 7, 8},
		{"alma-32", "spirit", 3, 4},
		{"alma-32", "doctrine", 8, 9},
		{"alma-32", "invite", 8, 9},
		// Kearon
		{"kearon-receive-his-gift", "teach", 8, 8},
		{"kearon-receive-his-gift", "help", 8, 8},
		{"kearon-receive-his-gift", "love", 4, 4},
		{"kearon-receive-his-gift", "spirit", 3, 3},
		{"kearon-receive-his-gift", "doctrine", 7, 7},
		{"kearon-receive-his-gift", "invite", 8, 8},
		// 3 Nephi 17
		{"3-nephi-17", "teach", 9, 9},
		{"3-nephi-17", "help", 9, 9},
		{"3-nephi-17", "love", 9, 9},
		{"3-nephi-17", "spirit", 8, 8},
		{"3-nephi-17", "doctrine", 4, 4},
		{"3-nephi-17", "invite", 5, 5},
		// D&C 121
		{"dc-121", "teach", 5, 5},
		{"dc-121", "help", 5, 5},
		{"dc-121", "love", 5, 5},
		{"dc-121", "spirit", 4, 4},
		{"dc-121", "doctrine", 8, 8},
		{"dc-121", "invite", 7, 7},
		// Holland
		{"holland-and-now-i-see", "teach", 7, 7},
		{"holland-and-now-i-see", "help", 6, 6},
		{"holland-and-now-i-see", "love", 4, 4},
		{"holland-and-now-i-see", "spirit", 7, 7},
		{"holland-and-now-i-see", "doctrine", 6, 6},
		{"holland-and-now-i-see", "invite", 3, 3},
		// Bednar
		{"bednar-their-own-judges", "teach", 5, 5},
		{"bednar-their-own-judges", "help", 5, 5},
		{"bednar-their-own-judges", "love", 2, 2},
		{"bednar-their-own-judges", "spirit", 3, 3},
		{"bednar-their-own-judges", "doctrine", 9, 9},
		{"bednar-their-own-judges", "invite", 6, 6},
	}

	stmt, err := db.Prepare("INSERT OR REPLACE INTO ground_truth (content, dimension, score_lo, score_hi) VALUES (?, ?, ?, ?)")
	if err != nil {
		fatal("prepare gt: %v", err)
	}
	defer stmt.Close()

	for _, g := range gt {
		_, err := stmt.Exec(g.content, g.dimension, g.lo, g.hi)
		if err != nil {
			fatal("insert gt %s/%s: %v", g.content, g.dimension, err)
		}
	}
	fmt.Printf("Initialized %s with %d ground truth scores\n", dbFile, len(gt))
}

// --- Import ---

type resultJSON struct {
	Tag       string  `json:"tag"`
	Content   string  `json:"content"`
	ModelShort string `json:"model_short"`
	Prompt    string  `json:"prompt"`
	Timestamp string  `json:"timestamp"`
	Response  string  `json:"response"`
	TokensIn  int     `json:"tokens_in"`
	TokensOut int     `json:"tokens_out"`
	TtftMs    int     `json:"ttft_ms"`
	LatencyMs int     `json:"latency_ms"`
}

type modelResponse struct {
	FocusOnChrist struct {
		TeachAboutChrist struct {
			Score int `json:"score"`
		} `json:"teach_about_christ"`
		HelpComeUntoChrist struct {
			Score int `json:"score"`
		} `json:"help_come_unto_christ"`
	} `json:"focus_on_christ"`
	Scores struct {
		Love     int `json:"love"`
		Spirit   int `json:"spirit"`
		Doctrine int `json:"doctrine"`
		Invite   int `json:"invite"`
	} `json:"scores"`
}

func cmdImport(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: scoring import <tag> [results-dir]")
		os.Exit(1)
	}
	tag := args[0]
	dir := filepath.Join("..", "results")
	if len(args) > 1 {
		dir = args[1]
	}

	db := openDB()
	defer db.Close()

	// Tag is stored inside the JSON, so we glob all files and filter
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		fatal("glob: %v", err)
	}

	imported := 0
	skipped := 0
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read %s: %v\n", f, err)
			continue
		}

		var result resultJSON
		if err := json.Unmarshal(data, &result); err != nil {
			continue // not a valid result file
		}

		// Filter by tag
		if result.Tag != tag {
			continue
		}

		// Only import titsw-* prompts (our scoring prompts)
		if !strings.HasPrefix(result.Prompt, "titsw") {
			continue
		}

		// Parse the model response to extract scores
		response := strings.TrimSpace(result.Response)
		// Strip markdown fencing if present
		if strings.HasPrefix(response, "```") {
			lines := strings.Split(response, "\n")
			if len(lines) > 2 {
				response = strings.Join(lines[1:len(lines)-1], "\n")
			}
		}

		var modelResp modelResponse
		if err := json.Unmarshal([]byte(response), &modelResp); err != nil {
			fmt.Fprintf(os.Stderr, "parse response in %s: %v\n", filepath.Base(f), err)
			continue
		}

		// Extract all 6 dimension scores
		scores := map[string]int{
			"teach":    modelResp.FocusOnChrist.TeachAboutChrist.Score,
			"help":     modelResp.FocusOnChrist.HelpComeUntoChrist.Score,
			"love":     modelResp.Scores.Love,
			"spirit":   modelResp.Scores.Spirit,
			"doctrine": modelResp.Scores.Doctrine,
			"invite":   modelResp.Scores.Invite,
		}

		for dim, score := range scores {
			_, err := db.Exec(
				`INSERT OR REPLACE INTO results (tag, content, dimension, score, model, timestamp, tokens_in, tokens_out, ttft_ms, latency_ms)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				result.Tag, result.Content, dim, score, result.ModelShort, result.Timestamp,
				result.TokensIn, result.TokensOut, result.TtftMs, result.LatencyMs,
			)
			if err != nil {
				fmt.Fprintf(os.Stderr, "insert %s/%s/%s: %v\n", result.Tag, result.Content, dim, err)
				continue
			}
		}
		imported++
	}
	fmt.Printf("Imported %d files (tag=%s), skipped %d\n", imported, tag, skipped)
}

// --- Stats ---

type scoreEntry struct {
	content   string
	dimension string
	score     int
	gtLo      float64
	gtHi      float64
}

func loadScores(db *sql.DB, tag string) []scoreEntry {
	rows, err := db.Query(`
		SELECT r.content, r.dimension, r.score, gt.score_lo, gt.score_hi
		FROM results r
		JOIN ground_truth gt ON r.content = gt.content AND r.dimension = gt.dimension
		WHERE r.tag = ?
		ORDER BY r.content, r.dimension
	`, tag)
	if err != nil {
		fatal("query scores: %v", err)
	}
	defer rows.Close()

	var entries []scoreEntry
	for rows.Next() {
		var e scoreEntry
		if err := rows.Scan(&e.content, &e.dimension, &e.score, &e.gtLo, &e.gtHi); err != nil {
			fatal("scan: %v", err)
		}
		entries = append(entries, e)
	}
	return entries
}

func gtMid(lo, hi float64) float64 {
	return (lo + hi) / 2
}

func delta(score int, lo, hi float64) float64 {
	mid := gtMid(lo, hi)
	return float64(score) - mid
}

type stats struct {
	Tag          string
	N            int
	Exact        int     // within ±0.5 of GT midpoint
	Within1      int     // within ±1.5 of GT midpoint
	InflationGe2 int     // delta >= +2
	UnderLe2     int     // delta <= -2
	MAE          float64 // mean absolute error from GT midpoint
}

func calcStats(entries []scoreEntry) stats {
	s := stats{N: len(entries)}
	totalAbsErr := 0.0
	for _, e := range entries {
		d := delta(e.score, e.gtLo, e.gtHi)
		ad := math.Abs(d)
		totalAbsErr += ad
		if ad <= 0.5 {
			s.Exact++
		}
		if ad <= 1.5 {
			s.Within1++
		}
		if d >= 2.0 {
			s.InflationGe2++
		}
		if d <= -2.0 {
			s.UnderLe2++
		}
	}
	if s.N > 0 {
		s.MAE = totalAbsErr / float64(s.N)
	}
	return s
}

func cmdStats(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: scoring stats <tag>")
		os.Exit(1)
	}
	tag := args[0]

	db := openDB()
	defer db.Close()

	entries := loadScores(db, tag)
	if len(entries) == 0 {
		fmt.Fprintf(os.Stderr, "No scores found for tag: %s\n", tag)
		os.Exit(1)
	}

	s := calcStats(entries)

	// Print detail table
	fmt.Printf("\n=== %s (%d scores) ===\n\n", tag, s.N)
	fmt.Printf("%-30s %4s %6s %5s %6s\n", "Content / Dimension", "GT", "Score", "Delta", "")
	fmt.Println(strings.Repeat("-", 60))

	lastContent := ""
	for _, e := range entries {
		label := e.content
		if e.content == lastContent {
			label = ""
		} else if lastContent != "" {
			fmt.Println()
		}
		lastContent = e.content

		gtStr := fmt.Sprintf("%.0f", e.gtLo)
		if e.gtLo != e.gtHi {
			gtStr = fmt.Sprintf("%.0f-%.0f", e.gtLo, e.gtHi)
		}

		d := delta(e.score, e.gtLo, e.gtHi)
		marker := ""
		if math.Abs(d) <= 0.5 {
			marker = "✓"
		} else if d >= 2 {
			marker = "↑↑"
		} else if d <= -2 {
			marker = "↓↓"
		}

		if label != "" {
			fmt.Printf("%-30s\n", label)
		}
		fmt.Printf("  %-28s %4s %5d %+5.1f %s\n", e.dimension, gtStr, e.score, d, marker)
	}

	// Print summary
	fmt.Println()
	fmt.Println(strings.Repeat("=", 40))
	fmt.Printf("MAE:              %.2f\n", s.MAE)
	fmt.Printf("Exact (±0):       %d/%d (%.0f%%)\n", s.Exact, s.N, pct(s.Exact, s.N))
	fmt.Printf("Within ±1:        %d/%d (%.0f%%)\n", s.Within1, s.N, pct(s.Within1, s.N))
	fmt.Printf("Inflation (≥+2):  %d/%d (%.0f%%)\n", s.InflationGe2, s.N, pct(s.InflationGe2, s.N))
	fmt.Printf("Underscoring (≤-2): %d/%d (%.0f%%)\n", s.UnderLe2, s.N, pct(s.UnderLe2, s.N))
}

// --- Compare ---

func cmdCompare(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: scoring compare <tag1,tag2,...>")
		os.Exit(1)
	}
	tags := strings.Split(args[0], ",")

	db := openDB()
	defer db.Close()

	// Collect stats for each tag
	allStats := make([]stats, len(tags))
	for i, tag := range tags {
		entries := loadScores(db, tag)
		allStats[i] = calcStats(entries)
		allStats[i].Tag = tag
	}

	// Print comparison table header
	fmt.Printf("\n%-25s", "Metric")
	for _, s := range allStats {
		fmt.Printf(" %12s", s.Tag)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("-", 25+13*len(tags)))

	// MAE row
	fmt.Printf("%-25s", "MAE")
	bestMAE := math.MaxFloat64
	for _, s := range allStats {
		if s.MAE < bestMAE {
			bestMAE = s.MAE
		}
	}
	for _, s := range allStats {
		val := fmt.Sprintf("%.2f", s.MAE)
		if s.MAE == bestMAE {
			val = fmt.Sprintf("*%.2f*", s.MAE)
		}
		fmt.Printf(" %12s", val)
	}
	fmt.Println()

	// Exact row
	fmt.Printf("%-25s", "Exact (±0)")
	bestExact := 0
	for _, s := range allStats {
		if s.Exact > bestExact {
			bestExact = s.Exact
		}
	}
	for _, s := range allStats {
		val := fmt.Sprintf("%d/%d (%d%%)", s.Exact, s.N, int(pct(s.Exact, s.N)))
		if s.Exact == bestExact {
			val = "*" + val + "*"
		}
		fmt.Printf(" %12s", val)
	}
	fmt.Println()

	// Within ±1
	fmt.Printf("%-25s", "Within ±1")
	bestW1 := 0
	for _, s := range allStats {
		if s.Within1 > bestW1 {
			bestW1 = s.Within1
		}
	}
	for _, s := range allStats {
		val := fmt.Sprintf("%d/%d (%d%%)", s.Within1, s.N, int(pct(s.Within1, s.N)))
		if s.Within1 == bestW1 {
			val = "*" + val + "*"
		}
		fmt.Printf(" %12s", val)
	}
	fmt.Println()

	// Inflation
	fmt.Printf("%-25s", "Inflation (≥+2)")
	bestInf := math.MaxInt32
	for _, s := range allStats {
		if s.InflationGe2 < bestInf {
			bestInf = s.InflationGe2
		}
	}
	for _, s := range allStats {
		val := fmt.Sprintf("%d (%d%%)", s.InflationGe2, int(pct(s.InflationGe2, s.N)))
		if s.InflationGe2 == bestInf {
			val = "*" + val + "*"
		}
		fmt.Printf(" %12s", val)
	}
	fmt.Println()

	// Underscoring
	fmt.Printf("%-25s", "Underscoring (≤-2)")
	bestUnder := math.MaxInt32
	for _, s := range allStats {
		if s.UnderLe2 < bestUnder {
			bestUnder = s.UnderLe2
		}
	}
	for _, s := range allStats {
		val := fmt.Sprintf("%d (%d%%)", s.UnderLe2, int(pct(s.UnderLe2, s.N)))
		if s.UnderLe2 == bestUnder {
			val = "*" + val + "*"
		}
		fmt.Printf(" %12s", val)
	}
	fmt.Println()

	// Per-content breakdown
	fmt.Println()
	fmt.Println("=== Per-Content Breakdown ===")

	// Get all content pieces
	contents := getContents(db)
	dims := []string{"teach", "help", "love", "spirit", "doctrine", "invite"}

	// Header
	fmt.Printf("\n%-25s %-10s %4s", "Content", "Dim", "GT")
	for _, tag := range tags {
		fmt.Printf(" %6s", tag)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("-", 40+7*len(tags)))

	for _, content := range contents {
		for i, dim := range dims {
			label := ""
			if i == 0 {
				label = shortContent(content)
			}

			// Get GT
			var gtLo, gtHi float64
			db.QueryRow("SELECT score_lo, score_hi FROM ground_truth WHERE content=? AND dimension=?", content, dim).Scan(&gtLo, &gtHi)
			gtStr := fmt.Sprintf("%.0f", gtLo)
			if gtLo != gtHi {
				gtStr = fmt.Sprintf("%.0f-%.0f", gtLo, gtHi)
			}

			fmt.Printf("%-25s %-10s %4s", label, dim, gtStr)

			for _, tag := range tags {
				var score int
				err := db.QueryRow("SELECT score FROM results WHERE tag=? AND content=? AND dimension=?", tag, content, dim).Scan(&score)
				if err != nil {
					fmt.Printf(" %6s", "-")
					continue
				}
				d := delta(score, gtLo, gtHi)
				marker := ""
				if math.Abs(d) <= 0.5 {
					marker = "✓"
				}
				fmt.Printf("  %d %s", score, marker)
			}
			fmt.Println()
		}
		fmt.Println()
	}
}

// --- Ground Truth ---

func cmdGT(args []string) {
	if len(args) == 0 || args[0] == "list" {
		cmdGTList()
		return
	}
	if args[0] == "set" {
		cmdGTSet(args[1:])
		return
	}
	fmt.Fprintln(os.Stderr, "Usage: scoring gt [list|set]")
	os.Exit(1)
}

func cmdGTList() {
	db := openDB()
	defer db.Close()

	rows, err := db.Query("SELECT content, dimension, score_lo, score_hi FROM ground_truth ORDER BY content, dimension")
	if err != nil {
		fatal("query gt: %v", err)
	}
	defer rows.Close()

	fmt.Printf("%-30s %-12s %s\n", "Content", "Dimension", "GT")
	fmt.Println(strings.Repeat("-", 52))

	lastContent := ""
	for rows.Next() {
		var content, dim string
		var lo, hi float64
		rows.Scan(&content, &dim, &lo, &hi)

		label := content
		if content == lastContent {
			label = ""
		}
		lastContent = content

		gtStr := fmt.Sprintf("%.0f", lo)
		if lo != hi {
			gtStr = fmt.Sprintf("%.0f-%.0f", lo, hi)
		}
		fmt.Printf("%-30s %-12s %s\n", label, dim, gtStr)
	}
}

func cmdGTSet(args []string) {
	if len(args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: scoring gt set <content> <dimension> <score> [<hi>]")
		os.Exit(1)
	}
	content := args[0]
	dim := args[1]
	var lo, hi float64
	fmt.Sscan(args[2], &lo)
	hi = lo
	if len(args) > 3 {
		fmt.Sscan(args[3], &hi)
	}

	db := openDB()
	defer db.Close()

	_, err := db.Exec("INSERT OR REPLACE INTO ground_truth (content, dimension, score_lo, score_hi) VALUES (?, ?, ?, ?)",
		content, dim, lo, hi)
	if err != nil {
		fatal("set gt: %v", err)
	}
	fmt.Printf("Set GT: %s / %s = %.0f", content, dim, lo)
	if lo != hi {
		fmt.Printf("-%.0f", hi)
	}
	fmt.Println()
}

// --- Merge ---

func cmdMerge(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: scoring merge <from-tag> <into-tag>")
		os.Exit(1)
	}
	from := args[0]
	into := args[1]

	db := openDB()
	defer db.Close()

	res, err := db.Exec(`
		INSERT OR IGNORE INTO results (tag, content, dimension, score, model, timestamp, tokens_in, tokens_out, ttft_ms, latency_ms)
		SELECT ?, content, dimension, score, model, timestamp, tokens_in, tokens_out, ttft_ms, latency_ms
		FROM results WHERE tag = ?
	`, into, from)
	if err != nil {
		fatal("merge: %v", err)
	}
	n, _ := res.RowsAffected()
	fmt.Printf("Merged %d scores from %s into %s\n", n, from, into)
}

// --- Helpers ---

func getContents(db *sql.DB) []string {
	rows, err := db.Query("SELECT DISTINCT content FROM ground_truth ORDER BY content")
	if err != nil {
		return nil
	}
	defer rows.Close()
	var contents []string
	for rows.Next() {
		var c string
		rows.Scan(&c)
		contents = append(contents, c)
	}
	sort.Strings(contents)
	return contents
}

func shortContent(name string) string {
	replacer := strings.NewReplacer(
		"-receive-his-gift", "",
		"-their-own-judges", "",
		"-and-now-i-see", "",
	)
	return replacer.Replace(name)
}

func pct(n, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(n) / float64(total) * 100
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
