// Package audit appends every state-changing event to a JSONL file
// for offline recovery if the SQLite DB ever corrupts.
package audit

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type Logger struct {
	mu sync.Mutex
	f  *os.File
}

func Open(path string) (*Logger, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &Logger{f: f}, nil
}

func (l *Logger) Close() error {
	if l == nil || l.f == nil {
		return nil
	}
	return l.f.Close()
}

// Event is what gets written.
type Event struct {
	TS    string                 `json:"ts"`
	Event string                 `json:"event"`
	Data  map[string]interface{} `json:"data,omitempty"`
}

func (l *Logger) Log(event string, data map[string]interface{}) error {
	if l == nil {
		return nil
	}
	e := Event{TS: time.Now().UTC().Format(time.RFC3339Nano), Event: event, Data: data}
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, err := l.f.Write(append(b, '\n')); err != nil {
		return err
	}
	return l.f.Sync()
}
