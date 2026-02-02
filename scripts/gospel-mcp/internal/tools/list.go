package tools

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ListParams are the parameters for gospel_list.
type ListParams struct {
	Path       string `json:"path"`
	SourceType string `json:"source_type"`
}

// List browses available gospel content.
func (t *Tools) List(args json.RawMessage) (*ListResponse, error) {
	var params ListParams
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("parsing params: %w", err)
	}

	// Default to listing content types
	if params.Path == "" && params.SourceType == "" {
		return t.listContentTypes()
	}

	// Determine what to list based on path structure
	path := strings.Trim(params.Path, "/")
	parts := strings.Split(path, "/")

	// Filter by source type if specified
	if params.SourceType != "" {
		switch params.SourceType {
		case "scriptures":
			return t.listScriptureVolumes()
		case "conference":
			return t.listConferenceYears()
		case "manual":
			return t.listManualCollections("manual")
		case "magazine":
			return t.listManualCollections("magazine")
		}
	}

	// Navigate path
	if len(parts) == 0 || parts[0] == "" {
		return t.listContentTypes()
	}

	switch parts[0] {
	case "scriptures":
		return t.listScriptures(parts[1:])
	case "general-conference", "conference":
		return t.listConference(parts[1:])
	case "manual", "manuals":
		return t.listManuals(parts[1:], "manual")
	case "magazine", "magazines", "liahona", "ensign":
		return t.listManuals(parts[1:], "magazine")
	default:
		// Try to interpret as volume or collection
		return t.tryListPath(parts)
	}
}

func (t *Tools) listContentTypes() (*ListResponse, error) {
	items := []ListItem{
		{Name: "Scriptures", Path: "scriptures", Type: "category"},
		{Name: "General Conference", Path: "general-conference", Type: "category"},
		{Name: "Manuals", Path: "manuals", Type: "category"},
		{Name: "Magazines", Path: "magazines", Type: "category"},
	}

	return &ListResponse{
		Path:  "/",
		Items: items,
		Total: len(items),
	}, nil
}

func (t *Tools) listScriptureVolumes() (*ListResponse, error) {
	rows, err := t.db.Query(`
		SELECT DISTINCT volume FROM scriptures ORDER BY 
		CASE volume 
			WHEN 'ot' THEN 1
			WHEN 'nt' THEN 2
			WHEN 'bofm' THEN 3
			WHEN 'dc-testament' THEN 4
			WHEN 'pgp' THEN 5
		END
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	volumeNames := map[string]string{
		"ot":           "Old Testament",
		"nt":           "New Testament",
		"bofm":         "Book of Mormon",
		"dc-testament": "Doctrine and Covenants",
		"pgp":          "Pearl of Great Price",
	}

	var items []ListItem
	for rows.Next() {
		var volume string
		if err := rows.Scan(&volume); err != nil {
			continue
		}
		name := volumeNames[volume]
		if name == "" {
			name = volume
		}
		items = append(items, ListItem{
			Name: name,
			Path: "scriptures/" + volume,
			Type: "volume",
		})
	}

	return &ListResponse{
		Path:  "scriptures",
		Items: items,
		Total: len(items),
	}, nil
}

func (t *Tools) listScriptures(parts []string) (*ListResponse, error) {
	if len(parts) == 0 {
		return t.listScriptureVolumes()
	}

	volume := parts[0]

	if len(parts) == 1 {
		// List books in volume
		rows, err := t.db.Query(`
			SELECT DISTINCT book, COUNT(DISTINCT chapter) as chapters
			FROM scriptures 
			WHERE volume = ?
			GROUP BY book
			ORDER BY MIN(id)
		`, volume)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var items []ListItem
		for rows.Next() {
			var book string
			var chapters int
			if err := rows.Scan(&book, &chapters); err != nil {
				continue
			}
			items = append(items, ListItem{
				Name:     formatBookName(volume, book),
				Path:     fmt.Sprintf("scriptures/%s/%s", volume, book),
				Type:     "book",
				Chapters: chapters,
			})
		}

		return &ListResponse{
			Path:  "scriptures/" + volume,
			Items: items,
			Total: len(items),
		}, nil
	}

	book := parts[1]

	if len(parts) == 2 {
		// List chapters in book
		rows, err := t.db.Query(`
			SELECT DISTINCT chapter, COUNT(*) as verses
			FROM scriptures 
			WHERE volume = ? AND book = ?
			GROUP BY chapter
			ORDER BY chapter
		`, volume, book)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var items []ListItem
		for rows.Next() {
			var chapter, verses int
			if err := rows.Scan(&chapter, &verses); err != nil {
				continue
			}
			items = append(items, ListItem{
				Name:  fmt.Sprintf("Chapter %d", chapter),
				Path:  fmt.Sprintf("scriptures/%s/%s/%d", volume, book, chapter),
				Type:  "chapter",
				Count: verses,
			})
		}

		return &ListResponse{
			Path:  fmt.Sprintf("scriptures/%s/%s", volume, book),
			Items: items,
			Total: len(items),
		}, nil
	}

	// List verses in chapter
	var chapter int
	fmt.Sscanf(parts[2], "%d", &chapter)

	rows, err := t.db.Query(`
		SELECT verse, SUBSTR(text, 1, 80) as preview
		FROM scriptures 
		WHERE volume = ? AND book = ? AND chapter = ?
		ORDER BY verse
	`, volume, book, chapter)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ListItem
	for rows.Next() {
		var verse int
		var preview string
		if err := rows.Scan(&verse, &preview); err != nil {
			continue
		}
		items = append(items, ListItem{
			Name: fmt.Sprintf("Verse %d: %s...", verse, preview),
			Path: fmt.Sprintf("scriptures/%s/%s/%d/%d", volume, book, chapter, verse),
			Type: "verse",
		})
	}

	return &ListResponse{
		Path:  fmt.Sprintf("scriptures/%s/%s/%d", volume, book, chapter),
		Items: items,
		Total: len(items),
	}, nil
}

func (t *Tools) listConferenceYears() (*ListResponse, error) {
	rows, err := t.db.Query(`
		SELECT DISTINCT year, COUNT(*) as talks
		FROM talks
		GROUP BY year
		ORDER BY year DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ListItem
	for rows.Next() {
		var year, count int
		if err := rows.Scan(&year, &count); err != nil {
			continue
		}
		items = append(items, ListItem{
			Name:  fmt.Sprintf("%d General Conference", year),
			Path:  fmt.Sprintf("general-conference/%d", year),
			Type:  "year",
			Count: count,
		})
	}

	return &ListResponse{
		Path:  "general-conference",
		Items: items,
		Total: len(items),
	}, nil
}

func (t *Tools) listConference(parts []string) (*ListResponse, error) {
	if len(parts) == 0 {
		return t.listConferenceYears()
	}

	var year int
	fmt.Sscanf(parts[0], "%d", &year)

	if len(parts) == 1 {
		// List months in year
		rows, err := t.db.Query(`
			SELECT DISTINCT month, COUNT(*) as talks
			FROM talks
			WHERE year = ?
			GROUP BY month
			ORDER BY month
		`, year)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var items []ListItem
		for rows.Next() {
			var month, count int
			if err := rows.Scan(&month, &count); err != nil {
				continue
			}
			items = append(items, ListItem{
				Name:  fmt.Sprintf("%s %d", monthName(month), year),
				Path:  fmt.Sprintf("general-conference/%d/%02d", year, month),
				Type:  "session",
				Count: count,
			})
		}

		return &ListResponse{
			Path:  fmt.Sprintf("general-conference/%d", year),
			Items: items,
			Total: len(items),
		}, nil
	}

	// List talks in month
	var month int
	fmt.Sscanf(parts[1], "%d", &month)

	rows, err := t.db.Query(`
		SELECT speaker, title, file_path
		FROM talks
		WHERE year = ? AND month = ?
		ORDER BY id
	`, year, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ListItem
	for rows.Next() {
		var speaker, title, filePath string
		if err := rows.Scan(&speaker, &title, &filePath); err != nil {
			continue
		}
		items = append(items, ListItem{
			Name: fmt.Sprintf("%s: %s", speaker, title),
			Path: filePath,
			Type: "talk",
		})
	}

	return &ListResponse{
		Path:  fmt.Sprintf("general-conference/%d/%02d", year, month),
		Items: items,
		Total: len(items),
	}, nil
}

func (t *Tools) listManualCollections(contentType string) (*ListResponse, error) {
	rows, err := t.db.Query(`
		SELECT DISTINCT collection_id, COUNT(*) as sections
		FROM manuals
		WHERE content_type = ?
		GROUP BY collection_id
		ORDER BY collection_id
	`, contentType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ListItem
	for rows.Next() {
		var collectionID string
		var count int
		if err := rows.Scan(&collectionID, &count); err != nil {
			continue
		}
		items = append(items, ListItem{
			Name:  formatCollectionTitle(collectionID),
			Path:  fmt.Sprintf("%ss/%s", contentType, collectionID),
			Type:  contentType,
			Count: count,
		})
	}

	category := "manuals"
	if contentType == "magazine" {
		category = "magazines"
	}

	return &ListResponse{
		Path:  category,
		Items: items,
		Total: len(items),
	}, nil
}

func (t *Tools) listManuals(parts []string, contentType string) (*ListResponse, error) {
	if len(parts) == 0 {
		return t.listManualCollections(contentType)
	}

	collectionID := parts[0]

	// List sections in collection
	rows, err := t.db.Query(`
		SELECT section, title, file_path
		FROM manuals
		WHERE collection_id = ?
		ORDER BY id
	`, collectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ListItem
	for rows.Next() {
		var section, title, filePath string
		if err := rows.Scan(&section, &title, &filePath); err != nil {
			continue
		}
		items = append(items, ListItem{
			Name: title,
			Path: filePath,
			Type: "section",
		})
	}

	return &ListResponse{
		Path:  fmt.Sprintf("%ss/%s", contentType, collectionID),
		Items: items,
		Total: len(items),
	}, nil
}

func (t *Tools) tryListPath(parts []string) (*ListResponse, error) {
	// Try to match volume abbreviations
	volumeMap := map[string]string{
		"ot": "ot", "old-testament": "ot",
		"nt": "nt", "new-testament": "nt",
		"bofm": "bofm", "book-of-mormon": "bofm",
		"dc": "dc-testament", "d&c": "dc-testament", "doctrine-and-covenants": "dc-testament",
		"pgp": "pgp", "pearl-of-great-price": "pgp",
	}

	if volume, ok := volumeMap[parts[0]]; ok {
		newParts := append([]string{volume}, parts[1:]...)
		return t.listScriptures(newParts)
	}

	// Couldn't determine path
	return &ListResponse{
		Path:  strings.Join(parts, "/"),
		Items: nil,
		Total: 0,
	}, nil
}
