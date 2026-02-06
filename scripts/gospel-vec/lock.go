package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

const lockFileName = "gospel-vec.lock"

// LockInfo stores metadata about the lock holder
type LockInfo struct {
	PID       int       `json:"pid"`
	Command   string    `json:"command"`
	StartedAt time.Time `json:"started_at"`
}

// IndexLock provides file-based mutual exclusion for indexing operations
type IndexLock struct {
	path string
}

// NewIndexLock creates a lock associated with the given data directory
func NewIndexLock(dataDir string) *IndexLock {
	return &IndexLock{
		path: filepath.Join(dataDir, lockFileName),
	}
}

// Acquire attempts to create the lock file. Returns an error if another
// process already holds the lock and is still running.
func (l *IndexLock) Acquire(command string) error {
	// Check for existing lock
	existing, err := l.readLock()
	if err == nil {
		// Lock file exists — check if the process is still alive
		if processExists(existing.PID) {
			return fmt.Errorf(
				"another index is already running (PID %d, command %q, started %s)\n"+
					"   If this is stale, delete %s",
				existing.PID, existing.Command,
				existing.StartedAt.Format("2006-01-02 15:04:05"),
				l.path,
			)
		}
		// Process is gone — stale lock, remove it
		fmt.Printf("🔓 Removing stale lock from PID %d\n", existing.PID)
		os.Remove(l.path)
	}

	// Write new lock file
	info := LockInfo{
		PID:       os.Getpid(),
		Command:   command,
		StartedAt: time.Now(),
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling lock info: %w", err)
	}

	// Use O_CREATE|O_EXCL for atomic create — fails if file already exists
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			// Race: another process grabbed it between our check and create
			return fmt.Errorf("another index process acquired the lock first — try again")
		}
		return fmt.Errorf("creating lock file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		os.Remove(l.path) // clean up on write failure
		return fmt.Errorf("writing lock file: %w", err)
	}

	return nil
}

// Release removes the lock file. Safe to call multiple times.
func (l *IndexLock) Release() {
	os.Remove(l.path)
}

// readLock reads and parses an existing lock file
func (l *IndexLock) readLock() (*LockInfo, error) {
	data, err := os.ReadFile(l.path)
	if err != nil {
		return nil, err
	}
	var info LockInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// processExists checks if a process with the given PID is running.
// On Windows, os.FindProcess always succeeds, so we send signal 0 to probe.
func processExists(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Signal 0 doesn't kill the process — it just checks existence.
	// On Unix this returns nil if alive, error if not.
	// On Windows, FindProcess succeeds for any PID; Signal returns an error
	// for non-existent processes.
	err = p.Signal(os.Signal(syscall.Signal(0)))
	return err == nil
}
