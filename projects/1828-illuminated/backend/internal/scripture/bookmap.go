package scripture

import "strings"

// BookMeta maps a bcbooks book name to our (volume, abbr, display_order)
// triple. The abbr column mirrors the directory names under
// gospel-library/eng/scriptures/, so workspace-style refs (dc/84:38,
// bofm/1-ne/3:7) round-trip cleanly with study documents.
type BookMeta struct {
	Volume       string
	Abbr         string
	DisplayOrder int
}

// volumeAbbrMap is hand-curated and stable. Adding a book = one new line.
// The order field reflects the canonical reading order within its volume
// for any sort-by-display surfaces.
var volumeAbbrMap = map[string]BookMeta{
	// Old Testament
	"Genesis":         {"ot", "gen", 1},
	"Exodus":          {"ot", "ex", 2},
	"Leviticus":       {"ot", "lev", 3},
	"Numbers":         {"ot", "num", 4},
	"Deuteronomy":     {"ot", "deut", 5},
	"Joshua":          {"ot", "josh", 6},
	"Judges":          {"ot", "judg", 7},
	"Ruth":            {"ot", "ruth", 8},
	"1 Samuel":        {"ot", "1-sam", 9},
	"2 Samuel":        {"ot", "2-sam", 10},
	"1 Kings":         {"ot", "1-kgs", 11},
	"2 Kings":         {"ot", "2-kgs", 12},
	"1 Chronicles":    {"ot", "1-chr", 13},
	"2 Chronicles":    {"ot", "2-chr", 14},
	"Ezra":            {"ot", "ezra", 15},
	"Nehemiah":        {"ot", "neh", 16},
	"Esther":          {"ot", "esth", 17},
	"Job":             {"ot", "job", 18},
	"Psalms":          {"ot", "ps", 19},
	"Proverbs":        {"ot", "prov", 20},
	"Ecclesiastes":    {"ot", "eccl", 21},
	// bcbooks uses "Solomon's Song"; gospel-library uses "song".
	"Solomon's Song":  {"ot", "song", 22},
	"Isaiah":          {"ot", "isa", 23},
	"Jeremiah":        {"ot", "jer", 24},
	"Lamentations":    {"ot", "lam", 25},
	"Ezekiel":         {"ot", "ezek", 26},
	"Daniel":          {"ot", "dan", 27},
	"Hosea":           {"ot", "hosea", 28},
	"Joel":            {"ot", "joel", 29},
	"Amos":            {"ot", "amos", 30},
	"Obadiah":         {"ot", "obad", 31},
	"Jonah":           {"ot", "jonah", 32},
	"Micah":           {"ot", "micah", 33},
	"Nahum":           {"ot", "nahum", 34},
	"Habakkuk":        {"ot", "hab", 35},
	"Zephaniah":       {"ot", "zeph", 36},
	"Haggai":          {"ot", "hag", 37},
	"Zechariah":       {"ot", "zech", 38},
	"Malachi":         {"ot", "mal", 39},

	// New Testament
	"Matthew":         {"nt", "matt", 40},
	"Mark":            {"nt", "mark", 41},
	"Luke":            {"nt", "luke", 42},
	"John":            {"nt", "john", 43},
	"Acts":            {"nt", "acts", 44},
	"Romans":          {"nt", "rom", 45},
	"1 Corinthians":   {"nt", "1-cor", 46},
	"2 Corinthians":   {"nt", "2-cor", 47},
	"Galatians":       {"nt", "gal", 48},
	"Ephesians":       {"nt", "eph", 49},
	"Philippians":     {"nt", "philip", 50},
	"Colossians":      {"nt", "col", 51},
	"1 Thessalonians": {"nt", "1-thes", 52},
	"2 Thessalonians": {"nt", "2-thes", 53},
	"1 Timothy":       {"nt", "1-tim", 54},
	"2 Timothy":       {"nt", "2-tim", 55},
	"Titus":           {"nt", "titus", 56},
	"Philemon":        {"nt", "philem", 57},
	"Hebrews":         {"nt", "heb", 58},
	"James":           {"nt", "james", 59},
	"1 Peter":         {"nt", "1-pet", 60},
	"2 Peter":         {"nt", "2-pet", 61},
	"1 John":          {"nt", "1-jn", 62},
	"2 John":          {"nt", "2-jn", 63},
	"3 John":          {"nt", "3-jn", 64},
	"Jude":            {"nt", "jude", 65},
	"Revelation":      {"nt", "rev", 66},

	// Book of Mormon
	"1 Nephi":         {"bofm", "1-ne", 67},
	"2 Nephi":         {"bofm", "2-ne", 68},
	"Jacob":           {"bofm", "jacob", 69},
	"Enos":            {"bofm", "enos", 70},
	"Jarom":           {"bofm", "jarom", 71},
	"Omni":            {"bofm", "omni", 72},
	"Words of Mormon": {"bofm", "w-of-m", 73},
	"Mosiah":          {"bofm", "mosiah", 74},
	"Alma":            {"bofm", "alma", 75},
	"Helaman":         {"bofm", "hel", 76},
	"3 Nephi":         {"bofm", "3-ne", 77},
	"4 Nephi":         {"bofm", "4-ne", 78},
	"Mormon":          {"bofm", "morm", 79},
	"Ether":           {"bofm", "ether", 80},
	"Moroni":          {"bofm", "moro", 81},

	// Doctrine and Covenants — single book, sections-as-chapters
	"Doctrine and Covenants": {"dc", "dc", 82},

	// Pearl of Great Price
	"Moses":                  {"pgp", "moses", 83},
	"Abraham":                {"pgp", "abr", 84},
	"Joseph Smith—Matthew":   {"pgp", "js-m", 85},
	"Joseph Smith—History":   {"pgp", "js-h", 86},
	"Articles of Faith":      {"pgp", "a-of-f", 87},
}

// LookupBookByName returns the BookMeta for a bcbooks book name. Returns
// zero BookMeta and ok=false when the name isn't recognized — the seeder
// logs and continues so a single new book doesn't crash ingest.
func LookupBookByName(name string) (BookMeta, bool) {
	bm, ok := volumeAbbrMap[name]
	return bm, ok
}

// LookupBookByAbbr is the inverse — used by ref-parsing. Case-insensitive.
func LookupBookByAbbr(abbr string) (string, BookMeta, bool) {
	abbr = strings.ToLower(abbr)
	for name, bm := range volumeAbbrMap {
		if bm.Abbr == abbr {
			return name, bm, true
		}
	}
	return "", BookMeta{}, false
}

// AllBooks returns every (bookName, BookMeta) pair sorted by display_order.
// Used by the seeder for deterministic INSERT order.
func AllBooks() []struct {
	Name string
	Meta BookMeta
} {
	out := make([]struct {
		Name string
		Meta BookMeta
	}, 0, len(volumeAbbrMap))
	for name, meta := range volumeAbbrMap {
		out = append(out, struct {
			Name string
			Meta BookMeta
		}{name, meta})
	}
	// stable order by DisplayOrder
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Meta.DisplayOrder < out[i].Meta.DisplayOrder {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}
