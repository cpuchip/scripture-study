package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dbPath := "data/gospel.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Show specific enriched chapters for verification
	refs := []struct {
		book string
		ch   int
	}{
		{"1-ne", 11},
		{"alma", 32},
		{"gen", 22},
	}

	for _, ref := range refs {
		var kw, summary, keyVerse, christTypes, connections sql.NullString
		err := db.QueryRow(`SELECT enrichment_keywords, enrichment_summary, enrichment_key_verse, enrichment_christ_types, enrichment_connections 
			FROM chapters WHERE book = ? AND chapter = ?`, ref.book, ref.ch).Scan(&kw, &summary, &keyVerse, &christTypes, &connections)
		if err != nil {
			fmt.Printf("\n=== %s %d === ERROR: %v\n", ref.book, ref.ch, err)
			continue
		}
		fmt.Printf("\n=== %s %d ===\n", ref.book, ref.ch)
		fmt.Printf("KEYWORDS: %s\n", kw.String)
		fmt.Printf("SUMMARY: %s\n", summary.String)
		fmt.Printf("KEY_VERSE: %s\n", keyVerse.String)
		fmt.Printf("CHRIST_TYPES: %s\n", christTypes.String)
		fmt.Printf("CONNECTIONS: %s\n", connections.String)
	}
}
