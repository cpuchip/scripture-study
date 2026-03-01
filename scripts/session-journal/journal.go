// Package journal defines the session journal data model and storage operations.
package journal

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Entry represents a single session journal entry.
type Entry struct {
	Date             string       `yaml:"date"`
	SessionID        string       `yaml:"session_id"`
	DurationEstimate string       `yaml:"duration_estimate,omitempty"`
	Intent           string       `yaml:"intent"`
	Discoveries      []Discovery  `yaml:"discoveries,omitempty"`
	Surprises        []string     `yaml:"surprises,omitempty"`
	Relationship     []Quality    `yaml:"relationship,omitempty"`
	CarryForward     []CarryItem  `yaml:"carry_forward,omitempty"`
	Questions        []string     `yaml:"questions,omitempty"`
	Tags             []string     `yaml:"tags,omitempty"`
	Retroactive      *Retroactive `yaml:"retroactive,omitempty"`
}

// Discovery is something we learned or uncovered together.
type Discovery struct {
	Title  string `yaml:"title"`
	Detail string `yaml:"detail"`
}

// Quality captures a relational dynamic from the session.
type Quality struct {
	Name   string `yaml:"quality"`
	Detail string `yaml:"detail"`
}

// CarryItem is a lesson, priority, or unresolved thread to bring forward.
type CarryItem struct {
	Priority     string `yaml:"priority"` // high, medium, low
	Note         string `yaml:"note"`
	Resolved     bool   `yaml:"resolved,omitempty"`
	ResolvedDate string `yaml:"resolved_date,omitempty"`
	ResolvedNote string `yaml:"resolved_note,omitempty"`
}

// Retroactive holds provenance info for entries reconstructed from chat history.
type Retroactive struct {
	Source        string `yaml:"source"`                  // e.g. "chat-export", "git-inferred", "memory"
	DateCertainty string `yaml:"date_certainty"`          // exact, approximate, inferred
	InferredFrom  string `yaml:"inferred_from,omitempty"` // e.g. "git log for study/charity.md"
	CapturedDate  string `yaml:"captured_date"`           // when this retroactive entry was written
}

// Store manages journal entries on disk.
type Store struct {
	Dir string // absolute path to .spec/journal/
}

// NewStore creates a Store, ensuring the directory exists.
func NewStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create journal dir: %w", err)
	}
	return &Store{Dir: dir}, nil
}

// Write persists an entry to disk as YAML.
// Filename: {date}--{session_id}.yaml
func (s *Store) Write(e *Entry) (string, error) {
	slug := sanitizeSlug(e.SessionID)
	name := fmt.Sprintf("%s--%s.yaml", e.Date, slug)
	path := filepath.Join(s.Dir, name)

	data, err := yaml.Marshal(e)
	if err != nil {
		return "", fmt.Errorf("marshal entry: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("write %s: %w", path, err)
	}
	return path, nil
}

// ReadAll loads every .yaml file in the journal directory.
func (s *Store) ReadAll() ([]*Entry, error) {
	files, err := filepath.Glob(filepath.Join(s.Dir, "*.yaml"))
	if err != nil {
		return nil, err
	}
	sort.Strings(files) // chronological by date prefix

	var entries []*Entry
	for _, f := range files {
		e, err := readEntryFile(f)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", filepath.Base(f), err)
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// Recent returns the last N entries.
func (s *Store) Recent(n int) ([]*Entry, error) {
	all, err := s.ReadAll()
	if err != nil {
		return nil, err
	}
	if n >= len(all) {
		return all, nil
	}
	return all[len(all)-n:], nil
}

// ByTopic returns entries whose discoveries, intent, tags, or questions
// mention the given topic (case-insensitive substring match).
func (s *Store) ByTopic(topic string) ([]*Entry, error) {
	all, err := s.ReadAll()
	if err != nil {
		return nil, err
	}
	t := strings.ToLower(topic)
	var matches []*Entry
	for _, e := range all {
		if entryMatchesTopic(e, t) {
			matches = append(matches, e)
		}
	}
	return matches, nil
}

// Since returns entries on or after the given date.
func (s *Store) Since(dateStr string) ([]*Entry, error) {
	cutoff, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("parse date %q: %w", dateStr, err)
	}
	all, err := s.ReadAll()
	if err != nil {
		return nil, err
	}
	var matches []*Entry
	for _, e := range all {
		d, err := time.Parse("2006-01-02", e.Date)
		if err != nil {
			continue
		}
		if !d.Before(cutoff) {
			matches = append(matches, e)
		}
	}
	return matches, nil
}

// CarryForwardItems returns all unresolved carry_forward items across all entries,
// optionally filtered by priority.
func (s *Store) CarryForwardItems(priority string, includeResolved bool) ([]CarryWithSource, error) {
	all, err := s.ReadAll()
	if err != nil {
		return nil, err
	}
	var items []CarryWithSource
	for _, e := range all {
		for _, c := range e.CarryForward {
			if !includeResolved && c.Resolved {
				continue
			}
			if priority != "" && priority != "all" && c.Priority != priority {
				continue
			}
			items = append(items, CarryWithSource{
				CarryItem: c,
				SessionID: e.SessionID,
				Date:      e.Date,
			})
		}
	}
	return items, nil
}

// AllQuestions returns all questions across entries with their source session.
func (s *Store) AllQuestions() ([]QuestionWithSource, error) {
	all, err := s.ReadAll()
	if err != nil {
		return nil, err
	}
	var qs []QuestionWithSource
	for _, e := range all {
		for _, q := range e.Questions {
			qs = append(qs, QuestionWithSource{
				Question:  q,
				SessionID: e.SessionID,
				Date:      e.Date,
			})
		}
	}
	return qs, nil
}

// CarryWithSource pairs a carry-forward item with its originating session.
type CarryWithSource struct {
	CarryItem
	SessionID string `yaml:"session_id"`
	Date      string `yaml:"date"`
}

// QuestionWithSource pairs a question with its originating session.
type QuestionWithSource struct {
	Question  string `yaml:"question"`
	SessionID string `yaml:"session_id"`
	Date      string `yaml:"date"`
}

// --- helpers ---

func readEntryFile(path string) (*Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var e Entry
	if err := yaml.Unmarshal(data, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

func sanitizeSlug(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		if r == ' ' || r == '_' {
			return '-'
		}
		return -1
	}, s)
	// collapse multiple dashes
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return strings.Trim(s, "-")
}

func entryMatchesTopic(e *Entry, topic string) bool {
	// Check intent
	if strings.Contains(strings.ToLower(e.Intent), topic) {
		return true
	}
	// Check tags
	for _, tag := range e.Tags {
		if strings.Contains(strings.ToLower(tag), topic) {
			return true
		}
	}
	// Check session ID
	if strings.Contains(strings.ToLower(e.SessionID), topic) {
		return true
	}
	// Check discoveries
	for _, d := range e.Discoveries {
		if strings.Contains(strings.ToLower(d.Title), topic) ||
			strings.Contains(strings.ToLower(d.Detail), topic) {
			return true
		}
	}
	// Check questions
	for _, q := range e.Questions {
		if strings.Contains(strings.ToLower(q), topic) {
			return true
		}
	}
	// Check surprises
	for _, s := range e.Surprises {
		if strings.Contains(strings.ToLower(s), topic) {
			return true
		}
	}
	// Check relationship
	for _, r := range e.Relationship {
		if strings.Contains(strings.ToLower(r.Name), topic) ||
			strings.Contains(strings.ToLower(r.Detail), topic) {
			return true
		}
	}
	// Check carry forward
	for _, c := range e.CarryForward {
		if strings.Contains(strings.ToLower(c.Note), topic) {
			return true
		}
	}
	return false
}
