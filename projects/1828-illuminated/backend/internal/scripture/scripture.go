// Package scripture serves canon verses, chapters, search, and the
// word-study reverse lookup. Phase 1 ships an empty Service; phase 2
// adds the parser, search, and handlers.
package scripture

import "net/http"

// Service is the HTTP-level handler bundle. It owns the parsing logic
// for ref strings + the SQL queries against scripture_books/chapters/verses.
type Service struct {
	db dbExecutor
}

// dbExecutor is the minimum surface we need from pgxpool.Pool — kept as
// an interface so tests can substitute. In phase 2 this widens to the
// pgxpool surface for Query / Exec.
type dbExecutor interface{}

// New constructs the Service. The pool is held but only used by handlers
// added in phase 2.
func New(pool any) *Service {
	return &Service{db: pool}
}

// Register attaches all scripture routes to mux.
func (s *Service) Register(mux *http.ServeMux) {
	// Phase 2 will populate: /api/scripture/:ref, chapter, search, word-study.
	_ = s
	_ = mux
}
