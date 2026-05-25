package byucitations

import (
	"fmt"
	htmlpkg "html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Citation struct {
	Reference string `json:"reference"`
	Speaker   string `json:"speaker"`
	Title     string `json:"title"`
	TalkID    string `json:"talk_id"`
	RefID     string `json:"ref_id"`
}

type LookupResult struct {
	Scripture string     `json:"scripture"`
	BookID    int        `json:"book_id"`
	Chapter   int        `json:"chapter"`
	Verses    string     `json:"verses"`
	Citations []Citation `json:"citations"`
}

var AbbrToBookID = map[string]int{
	"gen": 101, "ex": 102, "lev": 103, "num": 104, "deut": 105, "josh": 106, "judg": 107, "ruth": 108,
	"1-sam": 109, "2-sam": 110, "1-kgs": 111, "2-kgs": 112, "1-chr": 113, "2-chr": 114, "ezra": 115,
	"neh": 116, "esth": 117, "job": 118, "ps": 119, "prov": 120, "eccl": 121, "song": 122, "isa": 123,
	"jer": 124, "lam": 125, "ezek": 126, "dan": 127, "hosea": 128, "joel": 129, "amos": 130, "obad": 131,
	"jonah": 132, "micah": 133, "nahum": 134, "hab": 135, "zeph": 136, "hag": 137, "zech": 138, "mal": 139,
	"matt": 140, "mark": 141, "luke": 142, "john": 143, "acts": 144, "rom": 145, "1-cor": 146, "2-cor": 147,
	"gal": 148, "eph": 149, "philip": 150, "col": 151, "1-thes": 152, "2-thes": 153, "1-tim": 154, "2-tim": 155,
	"titus": 156, "philem": 157, "heb": 158, "james": 159, "1-pet": 160, "2-pet": 161, "1-jn": 162, "2-jn": 163,
	"3-jn": 164, "jude": 165, "rev": 166, "1-ne": 205, "2-ne": 206, "jacob": 207, "enos": 208, "jarom": 209,
	"omni": 210, "w-of-m": 211, "mosiah": 212, "alma": 213, "hel": 214, "3-ne": 215, "4-ne": 216, "morm": 217,
	"ether": 218, "moro": 219, "dc": 302, "moses": 401, "abr": 402, "js-m": 404, "js-h": 405, "a-of-f": 406,
}

var BookIDs = map[string]int{
	"Genesis": 101, "Exodus": 102, "Leviticus": 103, "Numbers": 104, "Deuteronomy": 105, "Joshua": 106,
	"Judges": 107, "Ruth": 108, "1 Samuel": 109, "2 Samuel": 110, "1 Kings": 111, "2 Kings": 112,
	"1 Chronicles": 113, "2 Chronicles": 114, "Ezra": 115, "Nehemiah": 116, "Esther": 117, "Job": 118,
	"Psalms": 119, "Proverbs": 120, "Ecclesiastes": 121, "Song of Solomon": 122, "Isaiah": 123, "Jeremiah": 124,
	"Lamentations": 125, "Ezekiel": 126, "Daniel": 127, "Hosea": 128, "Joel": 129, "Amos": 130, "Obadiah": 131,
	"Jonah": 132, "Micah": 133, "Nahum": 134, "Habakkuk": 135, "Zephaniah": 136, "Haggai": 137, "Zechariah": 138,
	"Malachi": 139, "Matthew": 140, "Mark": 141, "Luke": 142, "John": 143, "Acts": 144, "Romans": 145,
	"1 Corinthians": 146, "2 Corinthians": 147, "Galatians": 148, "Ephesians": 149, "Philippians": 150,
	"Colossians": 151, "1 Thessalonians": 152, "2 Thessalonians": 153, "1 Timothy": 154, "2 Timothy": 155,
	"Titus": 156, "Philemon": 157, "Hebrews": 158, "James": 159, "1 Peter": 160, "2 Peter": 161, "1 John": 162,
	"2 John": 163, "3 John": 164, "Jude": 165, "Revelation": 166, "1 Nephi": 205, "2 Nephi": 206, "Jacob": 207,
	"Enos": 208, "Jarom": 209, "Omni": 210, "Words of Mormon": 211, "Mosiah": 212, "Alma": 213, "Helaman": 214,
	"3 Nephi": 215, "4 Nephi": 216, "Mormon": 217, "Ether": 218, "Moroni": 219, "D&C": 302, "O.D.": 303,
	"Moses": 401, "Abraham": 402, "JS-M": 404, "JS-H": 405, "Articles of Faith": 406,
}

var sortedBookNames = []string{
	"Song of Solomon", "Words of Mormon", "Articles of Faith", "2 Thessalonians", "1 Thessalonians",
	"2 Corinthians", "1 Corinthians", "2 Chronicles", "1 Chronicles", "Philippians", "Ecclesiastes",
	"Lamentations", "Deuteronomy", "Colossians", "Zephaniah", "Zechariah", "Habakkuk", "Ephesians",
	"Galatians", "Revelation", "2 Timothy", "1 Timothy", "Leviticus", "2 Samuel", "1 Samuel", "Jeremiah",
	"Nehemiah", "Philemon", "2 Nephi", "1 Nephi", "3 Nephi", "4 Nephi", "Proverbs", "2 Kings", "1 Kings",
	"2 Peter", "1 Peter", "Hebrews", "Matthew", "Genesis", "Numbers", "2 John", "3 John", "1 John",
	"Abraham", "Helaman", "Obadiah", "Malachi", "Ezekiel", "Haggai", "Psalms", "Isaiah", "Daniel",
	"Mormon", "Romans", "Moroni", "Mosiah", "Exodus", "Joshua", "Judges", "Esther", "Hosea", "Micah",
	"Nahum", "Jonah", "Ether", "Jacob", "Moses", "James", "Jarom", "Titus", "Luke", "John", "Mark",
	"Joel", "Amos", "Acts", "Jude", "Alma", "Omni", "Enos", "Ezra", "Ruth", "D&C", "Job", "O.D.",
	"JS-M", "JS-H",
}

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) Lookup(bookID int, chapter int, verses string) (*LookupResult, error) {
	url := fmt.Sprintf("https://scriptures.byu.edu/citation_index/citation_ajax/Any/1830/2026/all/s/f/%d/%d?verses=%s",
		bookID, chapter, verses)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching BYU citations: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("BYU API status %d: %s", resp.StatusCode, string(body))
	}

	citations := parseHTML(string(body))

	return &LookupResult{
		Scripture: fmt.Sprintf("Book %d %d:%s", bookID, chapter, verses),
		BookID:    bookID,
		Chapter:   chapter,
		Verses:    verses,
		Citations: citations,
	}, nil
}

func parseHTML(html string) []Citation {
	var citations []Citation

	talkPattern := regexp.MustCompile(`getTalk\('(\d+)',\s*'(\d+)'\)`)
	refPattern := regexp.MustCompile(`class="reference[^"]*"[^>]*>([^<]+)<`)
	titlePattern := regexp.MustCompile(`class="talktitle[^"]*"[^>]*>([^<]+)<`)

	talkMatches := talkPattern.FindAllStringSubmatch(html, -1)
	refMatches := refPattern.FindAllStringSubmatch(html, -1)
	titleMatches := titlePattern.FindAllStringSubmatch(html, -1)

	count := len(talkMatches)
	if len(refMatches) < count {
		count = len(refMatches)
	}
	if len(titleMatches) < count {
		count = len(titleMatches)
	}

	for i := 0; i < count; i++ {
		refText := htmlpkg.UnescapeString(strings.TrimSpace(refMatches[i][1]))
		speaker, reference := parseRefText(refText)

		citations = append(citations, Citation{
			Reference: reference,
			Speaker:   speaker,
			Title:     htmlpkg.UnescapeString(strings.TrimSpace(titleMatches[i][1])),
			TalkID:    talkMatches[i][1],
			RefID:     talkMatches[i][2],
		})
	}

	return citations
}

func parseRefText(text string) (speaker, reference string) {
	parts := strings.SplitN(text, ", ", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1]), strings.TrimSpace(parts[0])
	}
	return text, ""
}

// ParseReference parses human scripture reference (fallback mode)
func ParseReference(input string) (book string, chapter int, verses string, err error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", 0, "", fmt.Errorf("empty reference")
	}

	normalized := normalizeInput(input)

	var matchedBook string
	var remainder string

	for _, book := range sortedBookNames {
		lower := strings.ToLower(book)
		normalizedLower := strings.ToLower(normalized)
		if strings.HasPrefix(normalizedLower, lower) {
			rest := normalized[len(book):]
			if rest == "" || rest[0] == ' ' || rest[0] == ':' || (rest[0] >= '0' && rest[0] <= '9') {
				matchedBook = book
				remainder = strings.TrimSpace(rest)
				break
			}
		}
	}

	if matchedBook == "" {
		return "", 0, "", fmt.Errorf("unrecognized book: %q", input)
	}

	if remainder == "" {
		return "", 0, "", fmt.Errorf("no chapter specified: %q", input)
	}

	chapterVersePattern := regexp.MustCompile(`^(\d+)(?::(.+))?$`)
	m := chapterVersePattern.FindStringSubmatch(remainder)
	if m == nil {
		return "", 0, "", fmt.Errorf("invalid format: %q", remainder)
	}

	_, err = fmt.Sscanf(m[1], "%d", &chapter)
	if err != nil {
		return "", 0, "", fmt.Errorf("invalid chapter %q", m[1])
	}

	if m[2] != "" {
		verses = m[2]
	}

	return matchedBook, chapter, verses, nil
}

func normalizeInput(input string) string {
	s := input
	s = strings.ReplaceAll(s, "—", "-")
	s = strings.ReplaceAll(s, "–", "-")
	s = regexp.MustCompile(`(?i)d\s*&\s*c`).ReplaceAllString(s, "D&C")
	s = regexp.MustCompile(`(?i)doctrine\s+and\s+covenants`).ReplaceAllString(s, "D&C")
	s = regexp.MustCompile(`(?i)js\s*[-—]\s*h`).ReplaceAllString(s, "JS-H")
	s = regexp.MustCompile(`(?i)js\s*[-—]\s*m`).ReplaceAllString(s, "JS-M")
	s = regexp.MustCompile(`(?i)joseph\s+smith[-—]history`).ReplaceAllString(s, "JS-H")
	s = regexp.MustCompile(`(?i)joseph\s+smith[-—]matthew`).ReplaceAllString(s, "JS-M")
	s = regexp.MustCompile(`(?i)articles?\s+of\s+faith`).ReplaceAllString(s, "Articles of Faith")
	s = regexp.MustCompile(`(?i)words?\s+of\s+mormon`).ReplaceAllString(s, "Words of Mormon")
	s = regexp.MustCompile(`(?i)official\s+declarations?`).ReplaceAllString(s, "O.D.")
	s = regexp.MustCompile(`(?i)o\.?\s*d\.?(?:\s|$)`).ReplaceAllString(s, "O.D. ")

	return strings.TrimSpace(s)
}
