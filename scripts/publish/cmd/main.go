// Publish script for scripture-study
// Converts local markdown files with relative gospel-library links
// to public files with absolute Church website URLs.
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	churchBaseURL = "https://www.churchofjesuschrist.org/study"
)

// External URL mappings for books not in gospel-library
var externalBookURLs = map[string]string{
	"lecture-on-faith": "https://lecturesonfaith.com",
}

// Lecture on Faith URL path mappings
var lecturePathMap = map[string]string{
	"00_introduction": "",              // Homepage/introduction
	"00_preface":      "/preface.html", // Preface
	"01_lecture_1":    "/1",            // Lecture First
	"02_lecture_2":    "/2",            // Lecture Second
	"03_lecture_3":    "/3",            // Lecture Third
	"04_lecture_4":    "/4",            // Lecture Fourth
	"05_lecture_5":    "/5",            // Lecture Fifth
	"06_lecture_6":    "/6",            // Lecture Sixth
	"07_lecture_7":    "/7",            // Lecture Seventh
}

var (
	inputDirs     = []string{"study", "lessons", "callings", "journal"}
	outputDir     = flag.String("output", "public", "Output directory for published files")
	verbose       = flag.Bool("v", false, "Verbose output")
	dryRun        = flag.Bool("dry-run", false, "Show what would be done without making changes")
	workspaceRoot string
)

// linkPattern matches markdown links: [text](path)
var linkPattern = regexp.MustCompile(`\[([^\]]*)\]\(([^)]+)\)`)

// versePattern extracts verse references from display text like "Moses 6:59-60" or "1 Nephi 3:7"
var versePattern = regexp.MustCompile(`(?i)(\d+\s+)?([A-Za-z&\-]+)\s+(\d+):(\d+)(?:[–-](\d+))?`)

func main() {
	flag.Parse()

	// Find workspace root
	var err error
	workspaceRoot, err = findWorkspaceRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding workspace root: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Scripture Study Publisher")
	fmt.Println("=========================")
	fmt.Printf("Workspace: %s\n", workspaceRoot)
	fmt.Printf("Output: %s\n", filepath.Join(workspaceRoot, *outputDir))
	if *dryRun {
		fmt.Println("Mode: DRY RUN (no files will be written)")
	}
	fmt.Println()

	// Process each input directory
	totalFiles := 0
	totalConverted := 0

	for _, dir := range inputDirs {
		inputPath := filepath.Join(workspaceRoot, dir)
		outputPath := filepath.Join(workspaceRoot, *outputDir, dir)

		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			if *verbose {
				fmt.Printf("Skipping %s (does not exist)\n", dir)
			}
			continue
		}

		fmt.Printf("Processing %s/...\n", dir)
		files, converted, err := processDirectory(inputPath, outputPath, dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", dir, err)
			continue
		}

		totalFiles += files
		totalConverted += converted
		fmt.Printf("  Processed %d files, %d links converted\n", files, converted)
	}

	fmt.Println()
	fmt.Printf("Total: %d files processed, %d links converted\n", totalFiles, totalConverted)
	if !*dryRun {
		fmt.Printf("Output written to: %s\n", filepath.Join(workspaceRoot, *outputDir))
	}
}

func findWorkspaceRoot() (string, error) {
	// Start from the current directory and look for go.work or known directories
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up to find workspace root
	dir := cwd
	for {
		// Check for go.work (workspace file)
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir, nil
		}
		// Check for gospel-library directory
		if _, err := os.Stat(filepath.Join(dir, "gospel-library")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return cwd, nil
}

func processDirectory(inputPath, outputPath, relDir string) (int, int, error) {
	fileCount := 0
	linkCount := 0

	err := filepath.WalkDir(inputPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Calculate relative path from input directory
		relPath, err := filepath.Rel(inputPath, path)
		if err != nil {
			return err
		}

		// Calculate output path
		outPath := filepath.Join(outputPath, relPath)

		// Process the file
		converted, err := processFile(path, outPath, relDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: error processing %s: %v\n", relPath, err)
			return nil // Continue with other files
		}

		fileCount++
		linkCount += converted

		if *verbose {
			fmt.Printf("  %s (%d links converted)\n", relPath, converted)
		}

		return nil
	})

	return fileCount, linkCount, err
}

func processFile(inputPath, outputPath, sourceDir string) (int, error) {
	// Read the input file
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return 0, err
	}

	// Calculate the directory of the source file for resolving relative paths
	sourceFileDir := filepath.Dir(inputPath)

	// Convert links
	converted := 0
	newContent := linkPattern.ReplaceAllStringFunc(string(content), func(match string) string {
		result, wasConverted := convertLink(match, sourceFileDir)
		if wasConverted {
			converted++
		}
		return result
	})

	if *dryRun {
		return converted, nil
	}

	// Create output directory
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return 0, err
	}

	// Write output file
	if err := os.WriteFile(outputPath, []byte(newContent), 0644); err != nil {
		return 0, err
	}

	return converted, nil
}

func convertLink(match, sourceDir string) (string, bool) {
	// Extract link text and path
	submatches := linkPattern.FindStringSubmatch(match)
	if len(submatches) != 3 {
		return match, false
	}

	linkText := submatches[1]
	linkPath := submatches[2]

	// Skip external links
	if strings.HasPrefix(linkPath, "http://") || strings.HasPrefix(linkPath, "https://") {
		return match, false
	}

	// Skip anchor-only links
	if strings.HasPrefix(linkPath, "#") {
		return match, false
	}

	// Check if this is a books/ link (e.g., lecture-on-faith)
	if strings.Contains(linkPath, "books/") || strings.Contains(linkPath, "lecture-on-faith") {
		externalURL := convertToExternalBookURL(linkPath, linkText, sourceDir)
		if externalURL != "" {
			return fmt.Sprintf("[%s](%s)", linkText, externalURL), true
		}
		return match, false
	}

	// Check if this is a gospel-library link
	if !strings.Contains(linkPath, "gospel-library") {
		return match, false
	}

	// Convert to Church URL
	churchURL := convertToChurchURL(linkPath, linkText, sourceDir)
	if churchURL == "" {
		return match, false
	}

	return fmt.Sprintf("[%s](%s)", linkText, churchURL), true
}

func convertToChurchURL(linkPath, linkText, sourceDir string) string {
	// Resolve relative path to absolute
	var absPath string
	if filepath.IsAbs(linkPath) {
		absPath = linkPath
	} else {
		absPath = filepath.Join(sourceDir, linkPath)
	}

	// Clean the path
	absPath = filepath.Clean(absPath)

	// Find the gospel-library/eng portion
	idx := strings.Index(absPath, "gospel-library")
	if idx == -1 {
		return ""
	}

	// Get the path after gospel-library/eng/
	pathAfterGL := absPath[idx:]
	pathAfterGL = strings.TrimPrefix(pathAfterGL, "gospel-library")
	pathAfterGL = strings.TrimPrefix(pathAfterGL, string(filepath.Separator))
	pathAfterGL = strings.TrimPrefix(pathAfterGL, "eng")
	pathAfterGL = strings.TrimPrefix(pathAfterGL, string(filepath.Separator))

	// Remove .md extension
	pathAfterGL = strings.TrimSuffix(pathAfterGL, ".md")

	// Convert backslashes to forward slashes for URL
	pathAfterGL = strings.ReplaceAll(pathAfterGL, "\\", "/")

	// Build the base URL
	baseURL := churchBaseURL + "/" + pathAfterGL + "?lang=eng"

	// Try to extract verse references from the link text
	verseFragment := extractVerseFragment(linkText, pathAfterGL)
	if verseFragment != "" {
		baseURL += "&id=" + verseFragment + "#" + strings.Split(verseFragment, "-")[0]
	}

	return baseURL
}

func extractVerseFragment(linkText, urlPath string) string {
	// Check if this is a scripture link (contains /scriptures/)
	if !strings.Contains(urlPath, "scriptures/") {
		return ""
	}

	// Try to extract verse references from link text
	// Examples:
	//   "Moses 6:59-60" -> "p59-p60"
	//   "1 Nephi 3:7" -> "p7"
	//   "D&C 93:36" -> "p36"
	//   "John 19:34" -> "p34"

	matches := versePattern.FindStringSubmatch(linkText)
	if len(matches) < 5 {
		return ""
	}

	// matches[4] is the start verse
	startVerse := matches[4]

	// matches[5] is the end verse (if range)
	endVerse := ""
	if len(matches) > 5 {
		endVerse = matches[5]
	}

	if startVerse == "" {
		return ""
	}

	if endVerse != "" && endVerse != "" {
		return fmt.Sprintf("p%s-p%s", startVerse, endVerse)
	}

	return "p" + startVerse
}

// convertToExternalBookURL converts links to books/ directory to external URLs
func convertToExternalBookURL(linkPath, linkText, sourceDir string) string {
	// Resolve relative path to absolute
	var absPath string
	if filepath.IsAbs(linkPath) {
		absPath = linkPath
	} else {
		absPath = filepath.Join(sourceDir, linkPath)
	}

	// Clean the path and normalize separators
	absPath = filepath.Clean(absPath)
	absPath = strings.ReplaceAll(absPath, "\\", "/")

	// Check for lecture-on-faith
	if strings.Contains(absPath, "lecture-on-faith") {
		return convertLectureOnFaithURL(absPath, linkPath)
	}

	// Add more book handlers here as needed
	return ""
}

// convertLectureOnFaithURL converts a lecture-on-faith link to lecturesonfaith.com URL
func convertLectureOnFaithURL(absPath, originalPath string) string {
	baseURL := externalBookURLs["lecture-on-faith"]
	if baseURL == "" {
		return ""
	}

	// Extract anchor if present (e.g., #1, #9, #q14)
	anchor := ""
	if idx := strings.Index(originalPath, "#"); idx != -1 {
		anchor = originalPath[idx+1:] // Get everything after #
	}

	// Remove any anchor from the path before extracting filename
	pathWithoutAnchor := absPath
	if idx := strings.Index(absPath, "#"); idx != -1 {
		pathWithoutAnchor = absPath[:idx]
	}

	// Extract the filename from the path
	filename := filepath.Base(pathWithoutAnchor)
	filename = strings.TrimSuffix(filename, ".md")

	// Look up the URL path
	urlPath, found := lecturePathMap[filename]
	if !found {
		// Default to homepage if not found
		urlPath = ""
	}

	// Build URL: https://lecturesonfaith.com/1
	url := baseURL + urlPath

	// Add anchor for verse/question references
	// lecturesonfaith.com uses #3 for verse 3, #q14 for question 14
	if anchor != "" {
		url += "/#" + anchor
	}

	return url
}
