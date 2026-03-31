package mcp

import (
	"fmt"
	"strings"
)

// parsedRef represents a parsed scripture or talk reference.
type parsedRef struct {
	Type     string // "scripture" or "talk" or "unknown"
	Book     string
	Chapter  int
	Verse    int
	EndVerse int // For ranges like 24-30; 0 means single verse
	Speaker  string
}

// parseReference parses a human-readable reference into structured form.
// Supports: "1 Nephi 3:7", "D&C 93:24-30", "Mosiah 4", "Moses 6:57"
func parseReference(ref string) parsedRef {
	ref = strings.TrimSpace(ref)
	ref = strings.ToLower(ref)

	// Map common variations
	ref = strings.ReplaceAll(ref, "doctrine and covenants", "dc")
	ref = strings.ReplaceAll(ref, "d&c", "dc")

	// Check for verse reference pattern: "book chapter:verse"
	parts := strings.Fields(ref)
	if len(parts) >= 2 {
		lastPart := parts[len(parts)-1]
		bookParts := parts[:len(parts)-1]

		if colonIdx := strings.Index(lastPart, ":"); colonIdx > 0 {
			// Has verse — parse chapter:verse or chapter:verse-endverse
			chapterStr := lastPart[:colonIdx]
			verseStr := lastPart[colonIdx+1:]

			var chapter int
			fmt.Sscanf(chapterStr, "%d", &chapter)

			var verse, endVerse int
			if dashIdx := strings.Index(verseStr, "-"); dashIdx > 0 {
				fmt.Sscanf(verseStr[:dashIdx], "%d", &verse)
				fmt.Sscanf(verseStr[dashIdx+1:], "%d", &endVerse)
			} else {
				fmt.Sscanf(verseStr, "%d", &verse)
			}

			book := normalizeBookName(strings.Join(bookParts, " "))
			if book != "" {
				return parsedRef{
					Type:     "scripture",
					Book:     book,
					Chapter:  chapter,
					Verse:    verse,
					EndVerse: endVerse,
				}
			}
		} else {
			// Just chapter number
			var chapter int
			fmt.Sscanf(lastPart, "%d", &chapter)

			if chapter > 0 {
				book := normalizeBookName(strings.Join(bookParts, " "))
				if book != "" {
					return parsedRef{
						Type:    "scripture",
						Book:    book,
						Chapter: chapter,
					}
				}
			}
		}
	}

	// Might be a talk reference with speaker name
	if strings.Contains(ref, ",") || strings.Contains(ref, "elder") ||
		strings.Contains(ref, "president") || strings.Contains(ref, "sister") {
		speaker := ref
		speaker = strings.TrimPrefix(speaker, "elder ")
		speaker = strings.TrimPrefix(speaker, "president ")
		speaker = strings.TrimPrefix(speaker, "sister ")
		return parsedRef{
			Type:    "talk",
			Speaker: speaker,
		}
	}

	return parsedRef{Type: "unknown"}
}

// normalizeBookName maps full book names and variations to abbreviations.
func normalizeBookName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))

	nameMap := map[string]string{
		// Old Testament
		"genesis": "gen", "exodus": "ex", "leviticus": "lev", "numbers": "num",
		"deuteronomy": "deut", "joshua": "josh", "judges": "judg", "ruth": "ruth",
		"1 samuel": "1-sam", "2 samuel": "2-sam", "1 kings": "1-kgs", "2 kings": "2-kgs",
		"1 chronicles": "1-chr", "2 chronicles": "2-chr", "ezra": "ezra", "nehemiah": "neh",
		"esther": "esth", "job": "job", "psalms": "ps", "psalm": "ps", "proverbs": "prov",
		"ecclesiastes": "eccl", "song of solomon": "song", "isaiah": "isa", "jeremiah": "jer",
		"lamentations": "lam", "ezekiel": "ezek", "daniel": "dan", "hosea": "hosea",
		"joel": "joel", "amos": "amos", "obadiah": "obad", "jonah": "jonah", "micah": "micah",
		"nahum": "nahum", "habakkuk": "hab", "zephaniah": "zeph", "haggai": "hag",
		"zechariah": "zech", "malachi": "mal",
		// New Testament
		"matthew": "matt", "mark": "mark", "luke": "luke", "john": "john", "acts": "acts",
		"romans": "rom", "1 corinthians": "1-cor", "2 corinthians": "2-cor",
		"galatians": "gal", "ephesians": "eph", "philippians": "philip",
		"colossians": "col", "1 thessalonians": "1-thes", "2 thessalonians": "2-thes",
		"1 timothy": "1-tim", "2 timothy": "2-tim", "titus": "titus", "philemon": "philem",
		"hebrews": "heb", "james": "james", "1 peter": "1-pet", "2 peter": "2-pet",
		"1 john": "1-jn", "2 john": "2-jn", "3 john": "3-jn", "jude": "jude",
		"revelation": "rev", "revelations": "rev",
		// Book of Mormon
		"1 nephi": "1-ne", "2 nephi": "2-ne", "jacob": "jacob", "enos": "enos",
		"jarom": "jarom", "omni": "omni", "words of mormon": "w-of-m",
		"mosiah": "mosiah", "alma": "alma", "helaman": "hel",
		"3 nephi": "3-ne", "4 nephi": "4-ne", "mormon": "morm",
		"ether": "ether", "moroni": "moro",
		// D&C
		"dc": "dc", "d&c": "dc", "doctrine and covenants": "dc",
		// Pearl of Great Price
		"moses": "moses", "abraham": "abr", "js matthew": "js-m", "js history": "js-h",
		"articles of faith": "a-of-f",
	}

	if abbr, ok := nameMap[name]; ok {
		return abbr
	}

	// Already an abbreviation?
	for _, abbr := range nameMap {
		if name == abbr {
			return abbr
		}
	}

	return ""
}

// formatBookName maps abbreviation back to human-readable name.
func formatBookName(book string) string {
	names := map[string]string{
		"gen": "Genesis", "ex": "Exodus", "lev": "Leviticus", "num": "Numbers", "deut": "Deuteronomy",
		"josh": "Joshua", "judg": "Judges", "ruth": "Ruth", "1-sam": "1 Samuel", "2-sam": "2 Samuel",
		"1-kgs": "1 Kings", "2-kgs": "2 Kings", "1-chr": "1 Chronicles", "2-chr": "2 Chronicles",
		"ezra": "Ezra", "neh": "Nehemiah", "esth": "Esther", "job": "Job", "ps": "Psalms",
		"prov": "Proverbs", "eccl": "Ecclesiastes", "song": "Song of Solomon", "isa": "Isaiah",
		"jer": "Jeremiah", "lam": "Lamentations", "ezek": "Ezekiel", "dan": "Daniel",
		"hosea": "Hosea", "joel": "Joel", "amos": "Amos", "obad": "Obadiah", "jonah": "Jonah",
		"micah": "Micah", "nahum": "Nahum", "hab": "Habakkuk", "zeph": "Zephaniah",
		"hag": "Haggai", "zech": "Zechariah", "mal": "Malachi",
		// New Testament
		"matt": "Matthew", "mark": "Mark", "luke": "Luke", "john": "John", "acts": "Acts",
		"rom": "Romans", "1-cor": "1 Corinthians", "2-cor": "2 Corinthians", "gal": "Galatians",
		"eph": "Ephesians", "philip": "Philippians", "col": "Colossians",
		"1-thes": "1 Thessalonians", "2-thes": "2 Thessalonians", "1-tim": "1 Timothy",
		"2-tim": "2 Timothy", "titus": "Titus", "philem": "Philemon", "heb": "Hebrews",
		"james": "James", "1-pet": "1 Peter", "2-pet": "2 Peter", "1-jn": "1 John",
		"2-jn": "2 John", "3-jn": "3 John", "jude": "Jude", "rev": "Revelation",
		// Book of Mormon
		"1-ne": "1 Nephi", "2-ne": "2 Nephi", "jacob": "Jacob", "enos": "Enos",
		"jarom": "Jarom", "omni": "Omni", "w-of-m": "Words of Mormon", "mosiah": "Mosiah",
		"alma": "Alma", "hel": "Helaman", "3-ne": "3 Nephi", "4-ne": "4 Nephi",
		"morm": "Mormon", "ether": "Ether", "moro": "Moroni",
		// Doctrine and Covenants
		"dc": "D&C",
		// Pearl of Great Price
		"moses": "Moses", "abr": "Abraham", "js-m": "JS—Matthew", "js-h": "JS—History",
		"a-of-f": "Articles of Faith",
	}

	if name, ok := names[book]; ok {
		return name
	}
	return book
}

// formatScriptureRef formats a scripture reference for display.
func formatScriptureRef(book string, chapter, verse int) string {
	if chapter == 0 && verse == 0 {
		// Study aid entry (TG, BD, GS) — format as topic name
		topic := strings.ReplaceAll(book, "-", " ")
		// Title-case each word
		words := strings.Fields(topic)
		for i, w := range words {
			if len(w) > 0 {
				words[i] = strings.ToUpper(w[:1]) + w[1:]
			}
		}
		return strings.Join(words, " ")
	}
	bookName := formatBookName(book)
	if verse > 0 {
		return fmt.Sprintf("%s %d:%d", bookName, chapter, verse)
	}
	return fmt.Sprintf("%s %d", bookName, chapter)
}
