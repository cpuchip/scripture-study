package vec

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"

	"golang.org/x/exp/mmap"
)

// .vecf file format:
//   Header (16 bytes):
//     bytes 0-3:   magic "VECF"
//     bytes 4-7:   version (uint32 LE, currently 1)
//     bytes 8-11:  dimension (uint32 LE)
//     bytes 12-15: count (uint32 LE)
//   Body:
//     count × dimension × 4 bytes of float32 (little-endian)
//
// Embeddings are stored pre-normalized (unit vectors), so cosine similarity
// reduces to dot product.

const (
	vecfMagic      = "VECF"
	vecfVersion    = 1
	vecfHeaderSize = 16
)

// VecFile provides mmap-backed read access to embedding vectors.
type VecFile struct {
	reader  *mmap.ReaderAt
	dim     int
	count   int
	embSize int // dim * 4 bytes per embedding
	path    string
}

// OpenVecFile memory-maps a .vecf file for reading.
// The mapping is instant — actual I/O happens on-demand via page faults.
func OpenVecFile(path string) (*VecFile, error) {
	r, err := mmap.Open(path)
	if err != nil {
		return nil, fmt.Errorf("mmap open %s: %w", path, err)
	}

	if r.Len() < vecfHeaderSize {
		r.Close()
		return nil, fmt.Errorf("vecf: %s too small (%d bytes)", path, r.Len())
	}

	// Read header
	var hdr [vecfHeaderSize]byte
	if _, err := r.ReadAt(hdr[:], 0); err != nil {
		r.Close()
		return nil, fmt.Errorf("vecf: reading header: %w", err)
	}

	magic := string(hdr[0:4])
	if magic != vecfMagic {
		r.Close()
		return nil, fmt.Errorf("vecf: bad magic %q (expected %q)", magic, vecfMagic)
	}

	version := binary.LittleEndian.Uint32(hdr[4:8])
	if version != vecfVersion {
		r.Close()
		return nil, fmt.Errorf("vecf: unsupported version %d", version)
	}

	dim := int(binary.LittleEndian.Uint32(hdr[8:12]))
	count := int(binary.LittleEndian.Uint32(hdr[12:16]))

	expectedSize := vecfHeaderSize + count*dim*4
	if r.Len() < expectedSize {
		r.Close()
		return nil, fmt.Errorf("vecf: file too small (%d bytes, expected %d for %d vectors of dim %d)",
			r.Len(), expectedSize, count, dim)
	}

	return &VecFile{
		reader:  r,
		dim:     dim,
		count:   count,
		embSize: dim * 4,
		path:    path,
	}, nil
}

// Dim returns the embedding dimension.
func (v *VecFile) Dim() int { return v.dim }

// Count returns the number of embeddings.
func (v *VecFile) Count() int { return v.count }

// Close unmaps the file.
func (v *VecFile) Close() error {
	if v.reader != nil {
		return v.reader.Close()
	}
	return nil
}

// DotProductAt computes the dot product between query and the embedding at
// index idx, reading directly from the mmap'd file. The query vector must
// be pre-normalized (unit vector) for this to equal cosine similarity.
//
// Uses a single reusable buffer to avoid per-call allocations.
func (v *VecFile) DotProductAt(idx int, query []float32, buf []byte) (float32, error) {
	if idx < 0 || idx >= v.count {
		return 0, fmt.Errorf("vecf: index %d out of range [0, %d)", idx, v.count)
	}
	off := int64(vecfHeaderSize) + int64(idx)*int64(v.embSize)
	if _, err := v.reader.ReadAt(buf[:v.embSize], off); err != nil {
		return 0, fmt.Errorf("vecf: reading embedding %d: %w", idx, err)
	}
	return dotProductFromBytes(query, buf[:v.embSize]), nil
}

// TopK scans all embeddings and returns the indices + scores of the top-k
// most similar vectors. The query must be a normalized float32 vector.
func (v *VecFile) TopK(query []float32, k int) ([]int, []float32, error) {
	if len(query) != v.dim {
		return nil, nil, fmt.Errorf("vecf: query dim %d != file dim %d", len(query), v.dim)
	}
	if k <= 0 {
		k = 10
	}
	if k > v.count {
		k = v.count
	}

	// Reusable buffer for reading one embedding at a time
	buf := make([]byte, v.embSize)

	// Simple top-K via partial sort (scan all, keep best K).
	// For 200K vectors this is fast enough — O(n*k) worst case.
	topIdx := make([]int, 0, k)
	topScore := make([]float32, 0, k)
	minScore := float32(-1)

	for i := 0; i < v.count; i++ {
		off := int64(vecfHeaderSize) + int64(i)*int64(v.embSize)
		if _, err := v.reader.ReadAt(buf, off); err != nil {
			return nil, nil, fmt.Errorf("vecf: reading embedding %d: %w", i, err)
		}
		score := dotProductFromBytes(query, buf)

		if len(topIdx) < k {
			topIdx = append(topIdx, i)
			topScore = append(topScore, score)
			if score < minScore || minScore == -1 {
				minScore = score
			}
		} else if score > minScore {
			// Replace the minimum entry
			minI := 0
			for j := 1; j < len(topScore); j++ {
				if topScore[j] < topScore[minI] {
					minI = j
				}
			}
			topIdx[minI] = i
			topScore[minI] = score
			// Recalculate min
			minScore = topScore[0]
			for j := 1; j < len(topScore); j++ {
				if topScore[j] < minScore {
					minScore = topScore[j]
				}
			}
		}
	}

	return topIdx, topScore, nil
}

// dotProductFromBytes computes dot product between a float32 slice and
// a little-endian byte buffer of float32 values.
func dotProductFromBytes(query []float32, embBytes []byte) float32 {
	var sum float32
	for i, q := range query {
		bits := binary.LittleEndian.Uint32(embBytes[i*4:])
		sum += q * math.Float32frombits(bits)
	}
	return sum
}

// normalizeVector normalizes a vector to unit length in-place.
func normalizeVector(v []float32) {
	var norm float32
	for _, x := range v {
		norm += x * x
	}
	norm = float32(math.Sqrt(float64(norm)))
	if norm > 0 {
		for i := range v {
			v[i] /= norm
		}
	}
}

// WriteVecFile writes embeddings to a .vecf file.
// Each embedding must have the same dimension.
func WriteVecFile(path string, dim int, embeddings [][]float32) error {
	f, err := os.Create(path + ".tmp")
	if err != nil {
		return fmt.Errorf("creating vecf file: %w", err)
	}
	defer func() {
		f.Close()
		os.Remove(path + ".tmp")
	}()

	// Write header
	var hdr [vecfHeaderSize]byte
	copy(hdr[0:4], vecfMagic)
	binary.LittleEndian.PutUint32(hdr[4:8], vecfVersion)
	binary.LittleEndian.PutUint32(hdr[8:12], uint32(dim))
	binary.LittleEndian.PutUint32(hdr[12:16], uint32(len(embeddings)))
	if _, err := f.Write(hdr[:]); err != nil {
		return fmt.Errorf("writing header: %w", err)
	}

	// Write embeddings
	buf := make([]byte, dim*4)
	for i, emb := range embeddings {
		if len(emb) != dim {
			return fmt.Errorf("embedding %d: dim %d != expected %d", i, len(emb), dim)
		}
		for j, v := range emb {
			binary.LittleEndian.PutUint32(buf[j*4:], math.Float32bits(v))
		}
		if _, err := f.Write(buf); err != nil {
			return fmt.Errorf("writing embedding %d: %w", i, err)
		}
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("closing vecf file: %w", err)
	}

	return os.Rename(path+".tmp", path)
}
