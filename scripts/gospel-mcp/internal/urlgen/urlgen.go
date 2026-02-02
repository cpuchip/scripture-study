// Package urlgen generates source URLs for churchofjesuschrist.org.
package urlgen

import (
	"fmt"
	"strings"
)

const baseURL = "https://www.churchofjesuschrist.org/study"

// Scripture generates a URL for a scripture reference.
// Example: Scripture("bofm", "mosiah", 3, 19) -> "https://www.churchofjesuschrist.org/study/scriptures/bofm/mosiah/3?lang=eng&id=p19#p19"
func Scripture(volume, book string, chapter, verse int) string {
	// Map our volume names to Church website paths
	volumePath := mapVolume(volume)

	if verse > 0 {
		return fmt.Sprintf("%s/scriptures/%s/%s/%d?lang=eng&id=p%d#p%d",
			baseURL, volumePath, book, chapter, verse, verse)
	}
	return fmt.Sprintf("%s/scriptures/%s/%s/%d?lang=eng",
		baseURL, volumePath, book, chapter)
}

// ScriptureChapter generates a URL for a chapter (no verse).
func ScriptureChapter(volume, book string, chapter int) string {
	volumePath := mapVolume(volume)
	return fmt.Sprintf("%s/scriptures/%s/%s/%d?lang=eng",
		baseURL, volumePath, book, chapter)
}

// Talk generates a URL for a conference talk.
// Example: Talk(2025, 4, "57nelson") -> "https://www.churchofjesuschrist.org/study/general-conference/2025/04/57nelson?lang=eng"
func Talk(year, month int, filename string) string {
	// Remove .md extension if present
	name := strings.TrimSuffix(filename, ".md")
	return fmt.Sprintf("%s/general-conference/%d/%02d/%s?lang=eng",
		baseURL, year, month, name)
}

// Manual generates a URL for a manual or lesson.
// Example: Manual("come-follow-me-for-home-and-church-old-testament-2026", "01") -> "..."
func Manual(collectionID, section string) string {
	if section != "" {
		return fmt.Sprintf("%s/manual/%s/%s?lang=eng",
			baseURL, collectionID, section)
	}
	return fmt.Sprintf("%s/manual/%s?lang=eng", baseURL, collectionID)
}

// Magazine generates a URL for a magazine article.
// Example: Magazine("liahona", 2026, 1, "article-name") -> "..."
func Magazine(magazineName string, year, month int, articleName string) string {
	if articleName != "" {
		return fmt.Sprintf("%s/%s/%d/%02d/%s?lang=eng",
			baseURL, magazineName, year, month, articleName)
	}
	return fmt.Sprintf("%s/%s/%d/%02d?lang=eng",
		baseURL, magazineName, year, month)
}

// TopicalGuide generates a URL for a Topical Guide entry.
func TopicalGuide(entry string) string {
	return fmt.Sprintf("%s/scriptures/tg/%s?lang=eng", baseURL, entry)
}

// BibleDictionary generates a URL for a Bible Dictionary entry.
func BibleDictionary(entry string) string {
	return fmt.Sprintf("%s/scriptures/bd/%s?lang=eng", baseURL, entry)
}

// GuideToScriptures generates a URL for a Guide to the Scriptures entry.
func GuideToScriptures(entry string) string {
	return fmt.Sprintf("%s/scriptures/gs/%s?lang=eng", baseURL, entry)
}

// mapVolume maps our internal volume names to Church website paths.
func mapVolume(volume string) string {
	switch volume {
	case "dc-testament":
		return "dc-testament"
	case "ot":
		return "ot"
	case "nt":
		return "nt"
	case "bofm":
		return "bofm"
	case "pgp":
		return "pgp"
	default:
		return volume
	}
}

// VolumeFromPath extracts the volume from a file path.
// Example: "gospel-library/eng/scriptures/bofm/mosiah/3.md" -> "bofm"
func VolumeFromPath(filePath string) string {
	parts := strings.Split(filePath, "/")
	for i, part := range parts {
		if part == "scriptures" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// BookFromPath extracts the book from a scripture file path.
// Example: "gospel-library/eng/scriptures/bofm/mosiah/3.md" -> "mosiah"
func BookFromPath(filePath string) string {
	parts := strings.Split(filePath, "/")
	for i, part := range parts {
		if part == "scriptures" && i+2 < len(parts) {
			return parts[i+2]
		}
	}
	return ""
}
