// Package dict serves Webster 1828, modern definitions (lazy + write-back),
// and tier metadata. Phase 1 ships an empty Service; phase 3 adds handlers.
package dict

import "net/http"

type Service struct {
	db                  any
	modernFetchDailyCap int
}

func New(pool any, modernFetchDailyCap int) *Service {
	return &Service{db: pool, modernFetchDailyCap: modernFetchDailyCap}
}

func (s *Service) Register(mux *http.ServeMux) {
	// Phase 3 will populate: /api/dict/1828/:word, /api/dict/modern/:word,
	// /api/dict/tier/:word, /api/dict/search.
	_ = s
	_ = mux
}
