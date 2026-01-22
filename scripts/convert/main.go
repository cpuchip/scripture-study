package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	releaseURL = "https://github.com/beandog/lds-scriptures/archive/refs/tags/2020.12.08.zip"
	zipName    = "lds-scriptures.zip"
)

type Verse struct {
	VolumeTitle     string `json:"volume_title"`
	BookTitle       string `json:"book_title"`
	BookShortTitle  string `json:"book_short_title"`
	ChapterNumber   int    `json:"chapter_number"`
	VerseNumber     int    `json:"verse_number"`
	VerseTitle      string `json:"verse_title"`
	VerseShortTitle string `json:"verse_short_title"`
	ScriptureText   string `json:"scripture_text"`
}

func main() {
	// Get the script directory and workspace root
	scriptDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting working directory: %v\n", err)
		os.Exit(1)
	}

	// Determine paths
	workspaceRoot := filepath.Dir(scriptDir)
	if filepath.Base(scriptDir) == "convert" {
		workspaceRoot = filepath.Dir(filepath.Dir(scriptDir))
	}

	tempDir := filepath.Join(scriptDir, "temp")
	scripturesDir := filepath.Join(workspaceRoot, "scriptures")
	zipPath := filepath.Join(tempDir, zipName)

	fmt.Println("Scripture Converter")
	fmt.Println("===================")
	fmt.Printf("Workspace root: %s\n", workspaceRoot)
	fmt.Printf("Output directory: %s\n", scripturesDir)
	fmt.Println()

	// Create temp directory
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temp directory: %v\n", err)
		os.Exit(1)
	}

	// Step 1: Download the release
	fmt.Println("[1/4] Downloading scriptures release...")
	if err := downloadFile(zipPath, releaseURL); err != nil {
		fmt.Fprintf(os.Stderr, "Error downloading: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("      Downloaded successfully!")

	// Step 2: Extract and find the JSON file
	fmt.Println("[2/4] Extracting JSON file...")
	jsonPath, err := extractJSONFromZip(zipPath, tempDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error extracting: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("      Found: %s\n", filepath.Base(jsonPath))

	// Step 3: Parse and convert to markdown
	fmt.Println("[3/4] Converting to markdown...")
	if err := convertToMarkdown(jsonPath, scripturesDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error converting: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("      Conversion complete!")

	// Step 4: Cleanup
	fmt.Println("[4/4] Cleaning up temporary files...")
	if err := os.RemoveAll(tempDir); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not remove temp directory: %v\n", err)
	}
	fmt.Println("      Cleanup complete!")

	fmt.Println()
	fmt.Println("Done! Scriptures have been converted to markdown.")
	fmt.Printf("Location: %s\n", scripturesDir)
}

func downloadFile(filepath string, url string) error {
	// Check if already downloaded
	if _, err := os.Stat(filepath); err == nil {
		fmt.Println("      Using cached download...")
		return nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func extractJSONFromZip(zipPath, destDir string) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	var jsonFile *zip.File
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, "json/lds-scriptures-json.txt") {
			jsonFile = f
			break
		}
	}

	if jsonFile == nil {
		return "", fmt.Errorf("JSON file not found in archive")
	}

	rc, err := jsonFile.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	destPath := filepath.Join(destDir, "lds-scriptures.json")
	outFile, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, rc)
	if err != nil {
		return "", err
	}

	return destPath, nil
}

func convertToMarkdown(jsonPath, outputDir string) error {
	// Read JSON file
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("reading JSON: %w", err)
	}

	var verses []Verse
	if err := json.Unmarshal(data, &verses); err != nil {
		return fmt.Errorf("parsing JSON: %w", err)
	}

	fmt.Printf("      Loaded %d verses\n", len(verses))

	// Organize verses by volume -> book -> chapter -> verse
	type Chapter struct {
		Number int
		Verses []Verse
	}
	type Book struct {
		Title      string
		ShortTitle string
		Chapters   map[int]*Chapter
	}
	type Volume struct {
		Title string
		Books map[string]*Book
	}

	volumes := make(map[string]*Volume)
	bookOrder := make(map[string][]string) // track book order per volume

	for _, v := range verses {
		// Get or create volume
		vol, ok := volumes[v.VolumeTitle]
		if !ok {
			vol = &Volume{
				Title: v.VolumeTitle,
				Books: make(map[string]*Book),
			}
			volumes[v.VolumeTitle] = vol
		}

		// Get or create book
		book, ok := vol.Books[v.BookTitle]
		if !ok {
			book = &Book{
				Title:      v.BookTitle,
				ShortTitle: v.BookShortTitle,
				Chapters:   make(map[int]*Chapter),
			}
			vol.Books[v.BookTitle] = book
			bookOrder[v.VolumeTitle] = append(bookOrder[v.VolumeTitle], v.BookTitle)
		}

		// Get or create chapter
		ch, ok := book.Chapters[v.ChapterNumber]
		if !ok {
			ch = &Chapter{
				Number: v.ChapterNumber,
				Verses: []Verse{},
			}
			book.Chapters[v.ChapterNumber] = ch
		}

		ch.Verses = append(ch.Verses, v)
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Volume order
	volumeOrder := []string{
		"Old Testament",
		"New Testament",
		"Book of Mormon",
		"Doctrine and Covenants",
		"Pearl of Great Price",
	}

	// Canonical book order within each volume
	canonicalBookOrder := map[string][]string{
		"Old Testament": {
			"Genesis", "Exodus", "Leviticus", "Numbers", "Deuteronomy",
			"Joshua", "Judges", "Ruth", "1 Samuel", "2 Samuel",
			"1 Kings", "2 Kings", "1 Chronicles", "2 Chronicles",
			"Ezra", "Nehemiah", "Esther", "Job", "Psalms",
			"Proverbs", "Ecclesiastes", "Song of Solomon", "Isaiah", "Jeremiah",
			"Lamentations", "Ezekiel", "Daniel", "Hosea", "Joel",
			"Amos", "Obadiah", "Jonah", "Micah", "Nahum",
			"Habakkuk", "Zephaniah", "Haggai", "Zechariah", "Malachi",
		},
		"New Testament": {
			"Matthew", "Mark", "Luke", "John", "Acts",
			"Romans", "1 Corinthians", "2 Corinthians", "Galatians", "Ephesians",
			"Philippians", "Colossians", "1 Thessalonians", "2 Thessalonians",
			"1 Timothy", "2 Timothy", "Titus", "Philemon", "Hebrews",
			"James", "1 Peter", "2 Peter", "1 John", "2 John",
			"3 John", "Jude", "Revelation",
		},
		"Book of Mormon": {
			"1 Nephi", "2 Nephi", "Jacob", "Enos", "Jarom",
			"Omni", "Words of Mormon", "Mosiah", "Alma", "Helaman",
			"3 Nephi", "4 Nephi", "Mormon", "Ether", "Moroni",
		},
		"Doctrine and Covenants": {
			"Doctrine and Covenants",
		},
		"Pearl of Great Price": {
			"Moses", "Abraham", "Joseph Smith--Matthew", "Joseph Smith--History", "Articles of Faith",
		},
	}

	// Write markdown files
	totalFiles := 0
	for _, volTitle := range volumeOrder {
		vol, ok := volumes[volTitle]
		if !ok {
			continue
		}

		// Create volume directory with sanitized name
		volDirName := sanitizeFilename(vol.Title)
		volDir := filepath.Join(outputDir, volDirName)
		if err := os.MkdirAll(volDir, 0755); err != nil {
			return fmt.Errorf("creating volume directory: %w", err)
		}

		// Write each book as a markdown file
		// Use canonical order if available, otherwise use order from data
		booksToWrite := canonicalBookOrder[volTitle]
		if len(booksToWrite) == 0 {
			booksToWrite = bookOrder[volTitle]
		}

		for bookIndex, bookTitle := range booksToWrite {
			book, exists := vol.Books[bookTitle]
			if !exists {
				continue
			}

			// Special handling for Doctrine and Covenants - split by section
			if book.Title == "Doctrine and Covenants" {
				// Sort chapters (sections)
				chapterNums := make([]int, 0, len(book.Chapters))
				for num := range book.Chapters {
					chapterNums = append(chapterNums, num)
				}
				sort.Ints(chapterNums)

				for _, chNum := range chapterNums {
					ch := book.Chapters[chNum]
					var sb strings.Builder
					sb.WriteString(fmt.Sprintf("# Doctrine and Covenants %d\n\n", ch.Number))

					// Sort verses
					sort.Slice(ch.Verses, func(i, j int) bool {
						return ch.Verses[i].VerseNumber < ch.Verses[j].VerseNumber
					})

					for _, verse := range ch.Verses {
						sb.WriteString(fmt.Sprintf("%d. %s\n", verse.VerseNumber, verse.ScriptureText))
					}
					sb.WriteString("\n")

					// Write section file
					fileName := fmt.Sprintf("Section_%03d.md", ch.Number)
					filePath := filepath.Join(volDir, fileName)
					if err := os.WriteFile(filePath, []byte(sb.String()), 0644); err != nil {
						return fmt.Errorf("writing %s: %w", filePath, err)
					}
					totalFiles++
				}
				continue
			}

			// Build markdown content
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("# %s\n\n", book.Title))

			// Sort chapters
			chapterNums := make([]int, 0, len(book.Chapters))
			for num := range book.Chapters {
				chapterNums = append(chapterNums, num)
			}
			sort.Ints(chapterNums)

			for _, chNum := range chapterNums {
				ch := book.Chapters[chNum]
				sb.WriteString(fmt.Sprintf("## Chapter %d\n\n", ch.Number))

				// Sort verses
				sort.Slice(ch.Verses, func(i, j int) bool {
					return ch.Verses[i].VerseNumber < ch.Verses[j].VerseNumber
				})

				for _, verse := range ch.Verses {
					sb.WriteString(fmt.Sprintf("%d. %s\n", verse.VerseNumber, verse.ScriptureText))
				}
				sb.WriteString("\n")
			}

			// Write file with numeric prefix
			sanitizedName := sanitizeFilename(book.Title)
			// Replace spaces with underscores for the filename
			sanitizedName = strings.ReplaceAll(sanitizedName, " ", "_")
			fileName := fmt.Sprintf("%02d_%s.md", bookIndex+1, sanitizedName)
			filePath := filepath.Join(volDir, fileName)
			if err := os.WriteFile(filePath, []byte(sb.String()), 0644); err != nil {
				return fmt.Errorf("writing %s: %w", filePath, err)
			}
			totalFiles++
		}
	}

	fmt.Printf("      Created %d markdown files\n", totalFiles)
	return nil
}

func sanitizeFilename(name string) string {
	// Replace problematic characters
	replacer := strings.NewReplacer(
		" ", "_",
		":", "",
		"/", "-",
		"\\", "-",
		"?", "",
		"*", "",
		"\"", "",
		"<", "",
		">", "",
		"|", "",
	)
	return replacer.Replace(name)
}
