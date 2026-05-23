// Scripture lookup via gospel-engine-v2 /api/get.
//
// See scripts/becoming/.spec/proposals/scripture-via-engine.md.
// Soft-dependency design: any failure (network, 5xx, auth, parse)
// returns ErrUnavailable. Callers map to HTTP 503; UI degrades
// gracefully. Process-local LRU(200) cache so a session that hit a
// verse once survives an engine blip.

package engine

import (
	"container/list"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/cpuchip/scripture-study/scripts/becoming/internal/scripture"
)

// ErrUnavailable is the sentinel returned for ANY failure of the
// engine lookup call. Callers map it to HTTP 503.
var ErrUnavailable = errors.New("scripture lookup temporarily unavailable")

// lookupCache holds the LRU. Module-scoped (rather than per-Client)
// because Client is constructed fresh in main.go on each boot and
// callers may create more than one — the cache is keyed by reference,
// not by client, so sharing is safe.
var lookupCache = struct {
	sync.Mutex
	cap   int
	ll    *list.List
	index map[string]*list.Element
}{
	cap:   200,
	ll:    list.New(),
	index: map[string]*list.Element{},
}

type lookupCacheEntry struct {
	key   string
	value *scripture.LookupResult
}

// engineVerseRow mirrors gospel-engine-v2's per-verse JSON shape.
type engineVerseRow struct {
	ID        int64  `json:"id"`
	Volume    string `json:"volume"`
	Book      string `json:"book"`
	Chapter   int    `json:"chapter"`
	Verse     int    `json:"verse"`
	Reference string `json:"reference"`
	Text      string `json:"text"`
	FilePath  string `json:"file_path"`
}

// engineGetResponse mirrors gospel-engine-v2's /api/get envelope.
type engineGetResponse struct {
	ReferenceQuery string           `json:"reference_query"`
	SourceType     string           `json:"source_type"`
	Verses         []engineVerseRow `json:"verses"`
}

// Lookup fetches a scripture reference via /api/get and returns the
// LookupResult shape becoming's handlers have always returned.
//
// Returns ErrUnavailable on ANY soft-dep failure (network, 5xx, auth,
// parse). Returns a regular error for "invalid reference" / "not
// found" so the router can differentiate 404 vs 503.
//
// nil receiver or unconfigured client → ErrUnavailable (so handlers
// can pass a nil/unconfigured client when ENGINE_SERVICE_TOKEN is
// unset and the lookup endpoint cleanly degrades while books + search
// keep working).
func (c *Client) Lookup(ctx context.Context, reference string) (*scripture.LookupResult, error) {
	if c == nil || !c.Configured() {
		return nil, ErrUnavailable
	}

	key := canonLookupRef(reference)
	if cached, ok := lookupCacheGet(key); ok {
		return cached, nil
	}

	url := fmt.Sprintf("%s/api/get?reference=%s", c.BaseURL, queryEscape(reference))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, ErrUnavailable
	}
	req.Header.Set("Authorization", "Bearer "+c.ServiceToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, ErrUnavailable
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusUnauthorized,
		resp.StatusCode == http.StatusForbidden,
		resp.StatusCode >= 500:
		return nil, ErrUnavailable
	case resp.StatusCode == http.StatusBadRequest:
		return nil, fmt.Errorf("invalid reference: %s", reference)
	case resp.StatusCode != http.StatusOK:
		return nil, ErrUnavailable
	}

	var parsed engineGetResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, ErrUnavailable
	}
	if len(parsed.Verses) == 0 {
		return nil, fmt.Errorf("reference not found: %s", reference)
	}

	first := parsed.Verses[0]
	result := &scripture.LookupResult{
		Reference: reference,
		Book:      first.Book,
		Chapter:   first.Chapter,
		Verses:    make([]scripture.Verse, 0, len(parsed.Verses)),
		Path:      first.FilePath,
	}
	for _, v := range parsed.Verses {
		result.Verses = append(result.Verses, scripture.Verse{
			Number:    v.Verse,
			Text:      v.Text,
			Reference: v.Reference,
		})
	}

	lookupCacheSet(key, result)
	return result, nil
}

// ---- helpers ---------------------------------------------------------

func canonLookupRef(ref string) string {
	return strings.Join(strings.Fields(strings.ToLower(ref)), " ")
}

// queryEscape — URL-encode for the query string. Avoids net/url import
// in this hot path; the safe set is RFC-3986 unreserved + a couple
// scripture-ref-friendly chars.
func queryEscape(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == ' ':
			b.WriteByte('+')
		case (c >= 'A' && c <= 'Z') ||
			(c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') ||
			c == '-' || c == '_' || c == '.' || c == '~' || c == ':':
			b.WriteByte(c)
		default:
			fmt.Fprintf(&b, "%%%02X", c)
		}
	}
	return b.String()
}

func lookupCacheGet(key string) (*scripture.LookupResult, bool) {
	lookupCache.Lock()
	defer lookupCache.Unlock()
	el, ok := lookupCache.index[key]
	if !ok {
		return nil, false
	}
	lookupCache.ll.MoveToFront(el)
	return el.Value.(*lookupCacheEntry).value, true
}

func lookupCacheSet(key string, value *scripture.LookupResult) {
	lookupCache.Lock()
	defer lookupCache.Unlock()
	if el, ok := lookupCache.index[key]; ok {
		lookupCache.ll.MoveToFront(el)
		el.Value.(*lookupCacheEntry).value = value
		return
	}
	el := lookupCache.ll.PushFront(&lookupCacheEntry{key: key, value: value})
	lookupCache.index[key] = el
	if lookupCache.ll.Len() > lookupCache.cap {
		old := lookupCache.ll.Back()
		if old != nil {
			lookupCache.ll.Remove(old)
			delete(lookupCache.index, old.Value.(*lookupCacheEntry).key)
		}
	}
}
