package mcp

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/stuffleberry/i1828/backend/internal/byucitations"
	"github.com/stuffleberry/i1828/backend/internal/httpx"
)

type Service struct {
	byuClient   *byucitations.Client
	engineURL   string
	engineToken string
	httpClient  *http.Client
}

func New(engineURL string, engineToken string) *Service {
	return &Service{
		byuClient:   byucitations.NewClient(),
		engineURL:   strings.TrimRight(engineURL, "/"),
		engineToken: engineToken,
		httpClient:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *Service) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/mcp/citations", s.handleCitations)
	mux.HandleFunc("GET /api/mcp/deep-search", s.handleDeepSearch)
}

func (s *Service) handleCitations(w http.ResponseWriter, r *http.Request) {
	b := r.URL.Query().Get("b")
	cStr := r.URL.Query().Get("c")
	rangeStr := r.URL.Query().Get("r")

	var bookID int
	var chapter int
	var verses string
	var err error

	if b != "" && cStr != "" {
		var ok bool
		bookID, ok = byucitations.AbbrToBookID[b]
		if !ok {
			httpx.WriteError(w, http.StatusBadRequest, "invalid_book", fmt.Sprintf("unknown book abbreviation: %s", b))
			return
		}
		chapter, err = strconv.Atoi(cStr)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid_chapter", "chapter must be an integer")
			return
		}
		verses = rangeStr
	} else {
		ref := r.URL.Query().Get("ref")
		if ref == "" {
			httpx.WriteError(w, http.StatusBadRequest, "missing_params", "either ?b and ?c or ?ref is required")
			return
		}
		bookName, ch, v, err := byucitations.ParseReference(ref)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "parse_error", err.Error())
			return
		}
		bookID = byucitations.BookIDs[bookName]
		chapter = ch
		verses = v
	}

	result, err := s.byuClient.Lookup(bookID, chapter, verses)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "byu_error", err.Error())
		return
	}

	httpx.WriteJSON(w, http.StatusOK, result)
}

func (s *Service) handleDeepSearch(w http.ResponseWriter, r *http.Request) {
	if s.engineURL == "" {
		httpx.WriteError(w, http.StatusServiceUnavailable, "disabled", "gospel-engine integrations not configured")
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		httpx.WriteError(w, http.StatusBadRequest, "missing_query", "?q parameter is required")
		return
	}

	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "20"
	}

	qParams := url.Values{}
	qParams.Set("q", query)
	qParams.Set("limit", limit)
	qParams.Set("mode", "hybrid")

	reqURL := fmt.Sprintf("%s/api/search?%s", s.engineURL, qParams.Encode())
	req, err := http.NewRequestWithContext(r.Context(), "GET", reqURL, nil)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "request_error", err.Error())
		return
	}

	if s.engineToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.engineToken)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "http_error", err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "read_error", err.Error())
		return
	}

	if resp.StatusCode != http.StatusOK {
		httpx.WriteError(w, resp.StatusCode, "engine_error", string(body))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}
