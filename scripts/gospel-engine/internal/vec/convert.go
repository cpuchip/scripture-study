package vec

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/philippgille/chromem-go"
)

// ConvertProgress reports progress during conversion.
type ConvertProgress struct {
	Collection string
	Total      int
	Extracted  int
	Written    int
}

// ConvertOptions controls the conversion process.
type ConvertOptions struct {
	Verbose    bool
	OnProgress func(ConvertProgress)
}

// ConvertToMmap extracts all documents from a chromem-go Store and writes
// them to .vecf flat files + SQLite vec_docs metadata. This is a one-time
// operation that trades a few minutes of conversion time for instant server
// startup on every subsequent launch.
func ConvertToMmap(store *Store, database *sql.DB, dataDir string, opts ConvertOptions) error {
	start := time.Now()

	// Ensure vec_docs table exists (schema.sql handles this, but be safe)
	if _, err := database.Exec(`DELETE FROM vec_docs`); err != nil {
		// Table might not exist yet if schema wasn't updated
		return fmt.Errorf("clearing vec_docs: %w (run schema migration first)", err)
	}

	// Detect embedding dimension from first available collection
	dim := 0
	collections := listCollections(store)
	if len(collections) == 0 {
		return fmt.Errorf("no collections found in vector store")
	}

	if opts.Verbose {
		fmt.Printf("Found %d collections to convert\n", len(collections))
	}

	for _, name := range collections {
		source, layer := parseCollectionName(name)
		col := store.db.GetCollection(name, store.embed)
		if col == nil || col.Count() == 0 {
			continue
		}

		count := col.Count()
		if opts.Verbose {
			fmt.Printf("  📦 %s: %d documents\n", name, count)
		}

		// Extract all documents using QueryEmbedding with a unit vector.
		// chromem-go normalizes stored embeddings, so we need any unit vector
		// of the correct dimension to run the query.
		if dim == 0 {
			dim = detectDimension(col, store.embed)
			if dim == 0 {
				return fmt.Errorf("could not detect embedding dimension")
			}
			if opts.Verbose {
				fmt.Printf("  Detected embedding dimension: %d\n", dim)
			}
		}

		unitVec := make([]float32, dim)
		unitVec[0] = 1.0 // unit vector along first axis

		results, err := col.QueryEmbedding(context.Background(), unitVec, count, nil, nil)
		if err != nil {
			return fmt.Errorf("extracting %s: %w", name, err)
		}

		if opts.Verbose {
			fmt.Printf("  ✅ Extracted %d documents from %s\n", len(results), name)
		}

		// Sort results by ID for deterministic output
		sort.Slice(results, func(i, j int) bool {
			return results[i].ID < results[j].ID
		})

		// Write .vecf file
		embeddings := make([][]float32, len(results))
		for i, r := range results {
			embeddings[i] = r.Embedding
		}

		vecfPath := filepath.Join(dataDir, name+".vecf")
		if err := WriteVecFile(vecfPath, dim, embeddings); err != nil {
			return fmt.Errorf("writing %s: %w", vecfPath, err)
		}

		if opts.Verbose {
			sizeBytes := vecfHeaderSize + len(results)*dim*4
			fmt.Printf("  💾 Wrote %s (%.1f MB, %d vectors)\n",
				vecfPath, float64(sizeBytes)/1024/1024, len(results))
		}

		// Write metadata to SQLite
		tx, err := database.Begin()
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}

		stmt, err := tx.Prepare(`
			INSERT INTO vec_docs (collection, vec_idx, doc_id, content, source, layer,
				book, chapter, reference, range_text, file_path,
				speaker, position, year, month, session, talk_title)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("prepare insert: %w", err)
		}
		defer stmt.Close()

		for i, r := range results {
			meta := MetadataFromMap(r.Metadata)
			content := r.Content
			if content == "" {
				content = "(empty)"
			}

			_, err := stmt.Exec(
				name, i, r.ID, content,
				string(source), string(layer),
				meta.Book, meta.Chapter, meta.Reference, meta.Range, meta.FilePath,
				meta.Speaker, meta.Position, meta.Year, meta.Month, meta.Session, meta.TalkTitle,
			)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("inserting doc %d (%s): %w", i, r.ID, err)
			}

			if opts.OnProgress != nil && (i+1)%1000 == 0 {
				opts.OnProgress(ConvertProgress{
					Collection: name,
					Total:      count,
					Extracted:  len(results),
					Written:    i + 1,
				})
			}
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit: %w", err)
		}

		if opts.Verbose {
			fmt.Printf("  📝 Wrote %d metadata rows for %s\n", len(results), name)
		}
	}

	if opts.Verbose {
		fmt.Printf("\n✅ Conversion complete in %v\n", time.Since(start).Round(time.Millisecond))
	}

	return nil
}

// listCollections returns all collection names that have data.
func listCollections(store *Store) []string {
	var names []string
	for _, source := range allSources() {
		for _, layer := range []Layer{LayerVerse, LayerParagraph, LayerSummary, LayerTheme} {
			name := collectionName(source, layer)
			col := store.db.GetCollection(name, store.embed)
			if col != nil && col.Count() > 0 {
				names = append(names, name)
			}
		}
	}
	return names
}

// detectDimension gets the embedding dimension by querying a single document.
func detectDimension(col *chromem.Collection, embedFunc chromem.EmbeddingFunc) int {
	// Try dimensions commonly used by embedding models (most likely first)
	for _, dim := range []int{4096, 3072, 1536, 768, 384} {
		unitVec := make([]float32, dim)
		unitVec[0] = 1.0

		results, err := col.QueryEmbedding(context.Background(), unitVec, 1, nil, nil)
		if err == nil && len(results) > 0 && len(results[0].Embedding) > 0 {
			return len(results[0].Embedding)
		}
	}
	return 0
}

// VecFilesExist checks whether mmap-format .vecf files exist in the data directory.
func VecFilesExist(dataDir string) bool {
	entries, err := filepath.Glob(filepath.Join(dataDir, "*.vecf"))
	if err != nil {
		return false
	}
	return len(entries) > 0
}

// collectionSizeMB returns the expected .vecf file size for a collection.
func collectionSizeMB(count, dim int) float64 {
	return float64(vecfHeaderSize+count*dim*4) / (1024 * 1024)
}

// formatSizeMB formats a size in MB.
func formatSizeMB(mb float64) string {
	if mb >= 1024 {
		return fmt.Sprintf("%.1f GB", mb/1024)
	}
	return fmt.Sprintf("%.0f MB", math.Round(mb))
}

// ConvertStats returns projected sizes without actually converting.
func ConvertStats(store *Store) string {
	collections := listCollections(store)
	if len(collections) == 0 {
		return "No vector collections found."
	}

	var sb strings.Builder
	sb.WriteString("Projected .vecf sizes (assuming dim=4096):\n\n")

	totalMB := 0.0
	for _, name := range collections {
		col := store.db.GetCollection(name, store.embed)
		if col == nil {
			continue
		}
		mb := collectionSizeMB(col.Count(), 4096)
		totalMB += mb
		sb.WriteString(fmt.Sprintf("  %s: %d vectors → %s\n", name, col.Count(), formatSizeMB(mb)))
	}

	sb.WriteString(fmt.Sprintf("\n  Total: %s (uncompressed)\n", formatSizeMB(totalMB)))
	sb.WriteString(fmt.Sprintf("  Current gob.gz: ~3.8 GB (compressed)\n"))

	return sb.String()
}
