package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// align controls per-column justification in a rendered table.
type align int

const (
	alignLeft align = iota
	alignRight
)

// printTable renders a simple two-space-gutter, column-aligned table to stdout.
// headers and each row should have the same length; aligns may be shorter than
// the column count (missing entries default to left).
func printTable(headers []string, rows [][]string, aligns []align) {
	n := len(headers)
	width := make([]int, n)
	for i, h := range headers {
		width[i] = runeLen(h)
	}
	for _, r := range rows {
		for i := 0; i < n && i < len(r); i++ {
			if l := runeLen(r[i]); l > width[i] {
				width[i] = l
			}
		}
	}
	alignOf := func(i int) align {
		if i < len(aligns) {
			return aligns[i]
		}
		return alignLeft
	}
	printRow(headers, width, alignOf)
	seps := make([]string, n)
	for i := range seps {
		seps[i] = strings.Repeat("-", width[i])
	}
	fmt.Println(strings.Join(seps, "  "))
	for _, r := range rows {
		printRow(r, width, alignOf)
	}
}

func printRow(cells []string, width []int, alignOf func(int) align) {
	out := make([]string, len(width))
	for i := range width {
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		pad := width[i] - runeLen(cell)
		if pad < 0 {
			pad = 0
		}
		if alignOf(i) == alignRight {
			out[i] = strings.Repeat(" ", pad) + cell
		} else {
			out[i] = cell + strings.Repeat(" ", pad)
		}
	}
	fmt.Println(strings.TrimRight(strings.Join(out, "  "), " "))
}

func runeLen(s string) int { return len([]rune(s)) }

// fmtMicro formats micro-dollars (1 dollar = 1,000,000 micro) as $#,###.####
// with four decimal places. Substrate runs are routinely sub-cent, so the usual
// two places would render most rows as $0.00 and hide real spend.
func fmtMicro(micro int64) string {
	neg := micro < 0
	if neg {
		micro = -micro
	}
	whole := micro / 1_000_000
	frac := (micro % 1_000_000) / 100 // 0..9999 → four decimal places
	s := fmt.Sprintf("$%s.%04d", groupThousands(whole), frac)
	if neg {
		s = "-" + s
	}
	return s
}

// fmtTokens renders a token count with thousands separators.
func fmtTokens(n int64) string { return groupThousands(n) }

// groupThousands inserts commas every three digits (handles negatives).
func groupThousands(n int64) string {
	s := fmt.Sprintf("%d", n)
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	if neg {
		return "-" + b.String()
	}
	return b.String()
}

// relTime renders an approximate "Nm/h/d ago" for a past timestamp.
func relTime(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	d := time.Since(t)
	switch {
	case d < 0:
		return "in the future"
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

// summarizeCounts renders a "status N · status N" summary, highest count first.
func summarizeCounts(counts map[string]int) string {
	type kv struct {
		k string
		v int
	}
	pairs := make([]kv, 0, len(counts))
	total := 0
	for k, v := range counts {
		pairs = append(pairs, kv{k, v})
		total += v
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].v != pairs[j].v {
			return pairs[i].v > pairs[j].v
		}
		return pairs[i].k < pairs[j].k
	})
	parts := make([]string, 0, len(pairs))
	for _, p := range pairs {
		parts = append(parts, fmt.Sprintf("%s %d", p.k, p.v))
	}
	return fmt.Sprintf("%d items — %s", total, strings.Join(parts, " · "))
}
