package urlgen

import (
	"fmt"
	"strings"
)

const baseURL = "https://www.churchofjesuschrist.org/study"

// Scripture generates a URL for a scripture reference.
func Scripture(volume, book string, chapter, verse int) string {
	volumePath := mapVolume(volume)
	if verse > 0 {
		return fmt.Sprintf("%s/scriptures/%s/%s/%d?lang=eng&id=p%d#p%d",
			baseURL, volumePath, book, chapter, verse, verse)
	}
	return fmt.Sprintf("%s/scriptures/%s/%s/%d?lang=eng",
		baseURL, volumePath, book, chapter)
}

// ScriptureChapter generates a URL for a chapter.
func ScriptureChapter(volume, book string, chapter int) string {
	volumePath := mapVolume(volume)
	return fmt.Sprintf("%s/scriptures/%s/%s/%d?lang=eng",
		baseURL, volumePath, book, chapter)
}

// Talk generates a URL for a conference talk.
func Talk(year, month int, filename string) string {
	name := strings.TrimSuffix(filename, ".md")
	return fmt.Sprintf("%s/general-conference/%d/%02d/%s?lang=eng",
		baseURL, year, month, name)
}

// Manual generates a URL for a manual or lesson.
func Manual(collectionID, section string) string {
	if section != "" {
		return fmt.Sprintf("%s/manual/%s/%s?lang=eng", baseURL, collectionID, section)
	}
	return fmt.Sprintf("%s/manual/%s?lang=eng", baseURL, collectionID)
}

func mapVolume(volume string) string {
	// Our internal names match the website paths
	return volume
}
