package indexer

import (
	"regexp"
	"strings"
	"unicode"
)

var titlePattern = regexp.MustCompile(`^#\s+(.+)$`)

func extractTitle(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if m := titlePattern.FindStringSubmatch(line); m != nil {
			return strings.TrimSpace(m[1])
		}
	}
	return ""
}

func formatTitle(filename string) string {
	name := strings.TrimSuffix(filename, ".md")
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")
	words := strings.Fields(name)
	for i, word := range words {
		if len(word) > 0 {
			runes := []rune(word)
			runes[0] = unicode.ToUpper(runes[0])
			words[i] = string(runes)
		}
	}
	return strings.Join(words, " ")
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		return '-'
	}, s)
	// Collapse multiple dashes
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return strings.Trim(s, "-")
}

// FormatBookName converts directory name to display name.
var bookNameMap = map[string]string{
	// Book of Mormon
	"1-ne": "1 Nephi", "2-ne": "2 Nephi", "jacob": "Jacob", "enos": "Enos",
	"jarom": "Jarom", "omni": "Omni", "w-of-m": "Words of Mormon",
	"mosiah": "Mosiah", "alma": "Alma", "hel": "Helaman",
	"3-ne": "3 Nephi", "4-ne": "4 Nephi", "morm": "Mormon",
	"ether": "Ether", "moro": "Moroni",
	// D&C
	"dc": "D&C",
	// Pearl of Great Price
	"moses": "Moses", "abr": "Abraham", "js-m": "Joseph Smith—Matthew",
	"js-h": "Joseph Smith—History", "a-of-f": "Articles of Faith",
	// Old Testament
	"gen": "Genesis", "ex": "Exodus", "lev": "Leviticus", "num": "Numbers",
	"deut": "Deuteronomy", "josh": "Joshua", "judg": "Judges", "ruth": "Ruth",
	"1-sam": "1 Samuel", "2-sam": "2 Samuel", "1-kgs": "1 Kings", "2-kgs": "2 Kings",
	"1-chr": "1 Chronicles", "2-chr": "2 Chronicles", "ezra": "Ezra",
	"neh": "Nehemiah", "esth": "Esther", "job": "Job", "ps": "Psalms",
	"prov": "Proverbs", "eccl": "Ecclesiastes", "song": "Song of Solomon",
	"isa": "Isaiah", "jer": "Jeremiah", "lam": "Lamentations", "ezek": "Ezekiel",
	"dan": "Daniel", "hosea": "Hosea", "joel": "Joel", "amos": "Amos",
	"obad": "Obadiah", "jonah": "Jonah", "micah": "Micah", "nahum": "Nahum",
	"hab": "Habakkuk", "zeph": "Zephaniah", "hag": "Haggai",
	"zech": "Zechariah", "mal": "Malachi",
	// New Testament
	"matt": "Matthew", "mark": "Mark", "luke": "Luke", "john": "John",
	"acts": "Acts", "rom": "Romans", "1-cor": "1 Corinthians", "2-cor": "2 Corinthians",
	"gal": "Galatians", "eph": "Ephesians", "philip": "Philippians",
	"col": "Colossians", "1-thes": "1 Thessalonians", "2-thes": "2 Thessalonians",
	"1-tim": "1 Timothy", "2-tim": "2 Timothy", "titus": "Titus",
	"philem": "Philemon", "heb": "Hebrews", "james": "James",
	"1-pet": "1 Peter", "2-pet": "2 Peter", "1-jn": "1 John",
	"2-jn": "2 John", "3-jn": "3 John", "jude": "Jude", "rev": "Revelation",
}

func FormatBookName(dirName string) string {
	if name, ok := bookNameMap[dirName]; ok {
		return name
	}
	return formatTitle(dirName)
}
