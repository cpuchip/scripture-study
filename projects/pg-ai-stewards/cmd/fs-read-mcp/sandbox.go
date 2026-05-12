// Path-scope enforcement for fs-read-mcp. Every tool call validates
// candidate paths against the sandbox's allow-list before any
// filesystem syscall that exposes contents.
//
// Threat model: a misbehaving or compromised agent issues fs_read /
// fs_list / fs_search with paths trying to escape the allow-list
// (../.. traversal, symlink escape, absolute-path injection). The
// sandbox treats every path as untrusted input and rejects anything
// that resolves outside the allowed globs.

package main

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

// resolvePath normalizes a candidate path against the sandbox.
//
// Input convention: paths in tool args are repo-root-relative (no
// leading slash). Absolute paths are rejected. After normalization
// the result is checked against the allow-list — if it doesn't match
// at least one allowed glob, the call is rejected.
//
// Returns the absolute path safe for use with os.Open/os.Stat. The
// returned relPath is the canonical repo-root-relative form (forward
// slashes) suitable for logging or for further glob checks.
func (s *sandbox) resolvePath(candidate string) (absPath, relPath string, err error) {
	if candidate == "" {
		return "", "", fmt.Errorf("path is empty")
	}
	if filepath.IsAbs(candidate) {
		return "", "", fmt.Errorf("path must be repo-root-relative, got absolute: %q", candidate)
	}
	// Normalize forward-slash + clean. Clean strips trailing slash and
	// resolves "./" and "../" textually. We re-check absoluteness after
	// joining the root — a candidate like "../etc/passwd" cleans to
	// "../etc/passwd" which after Join lands outside the root.
	clean := filepath.ToSlash(filepath.Clean(candidate))
	if strings.HasPrefix(clean, "../") || clean == ".." {
		return "", "", fmt.Errorf("path escapes repo root: %q", candidate)
	}

	absPath = filepath.Join(s.repoRoot, clean)
	// Even after Clean+Join, a symlink in the path could point outside.
	// We Eval after open in the read path; here we just check that the
	// candidate (pre-symlink-resolution) sits inside the root.
	rootAbs, err := filepath.Abs(s.repoRoot)
	if err != nil {
		return "", "", fmt.Errorf("repo-root not resolvable: %v", err)
	}
	absResolved, err := filepath.Abs(absPath)
	if err != nil {
		return "", "", fmt.Errorf("path not resolvable: %v", err)
	}
	if !strings.HasPrefix(absResolved, rootAbs) {
		return "", "", fmt.Errorf("path resolves outside repo root: %q -> %q", candidate, absResolved)
	}

	if !s.matchesAllowed(clean) {
		return "", "", fmt.Errorf("path %q does not match any --allowed-paths glob", clean)
	}

	return absResolved, clean, nil
}

// matchesAllowed checks whether a repo-root-relative path matches at
// least one allowed glob. Globs follow filepath.Match semantics but
// we add support for "**" (recursive directory wildcard) — common
// expectation for dev tools.
//
// Examples that match ".spec/journal/*":
//   ".spec/journal/foo.md"          ✓
//   ".spec/journal/sub/foo.md"      ✗  (single-level glob)
//
// Examples that match "docs/**":
//   "docs/foo.md"                   ✓
//   "docs/sub/foo.md"               ✓
//   "docs/sub/deep/foo.md"          ✓
func (s *sandbox) matchesAllowed(relPath string) bool {
	for _, g := range s.allowedGlobs {
		if matchGlob(g, relPath) {
			return true
		}
	}
	return false
}

// matchGlob extends path.Match with "**" recursive wildcard.
//
// path.Match (not filepath.Match) is what we want: it always uses
// '/' as separator regardless of host OS, so "*" reliably means
// "any chars except /". filepath.Match on Windows treats '\' as the
// separator and lets '*' match across '/', which is a sandbox bypass.
func matchGlob(pattern, candidate string) bool {
	// Fast path: no "**" — defer to stdlib path.Match.
	if !strings.Contains(pattern, "**") {
		ok, _ := path.Match(pattern, candidate)
		return ok
	}
	// Split pattern around "**". A single "**" matches any sequence
	// of path segments including empty. We handle the most common
	// forms: "prefix/**", "**/suffix", "prefix/**/suffix".
	parts := strings.SplitN(pattern, "**", 2)
	prefix := strings.TrimSuffix(parts[0], "/")
	suffix := strings.TrimPrefix(parts[1], "/")

	if prefix != "" && !strings.HasPrefix(candidate, prefix+"/") && candidate != prefix {
		return false
	}
	if suffix == "" {
		return true // "prefix/**" matches anything under prefix
	}
	// Walk from each possible split of candidate to find suffix match.
	rest := strings.TrimPrefix(candidate, prefix)
	rest = strings.TrimPrefix(rest, "/")
	segs := strings.Split(rest, "/")
	for i := 0; i <= len(segs); i++ {
		tail := strings.Join(segs[i:], "/")
		if ok, _ := path.Match(suffix, tail); ok {
			return true
		}
	}
	return false
}

// Silence unused-import on platforms where filepath is otherwise
// unused in this file (it is used in resolvePath above).
var _ = filepath.Separator
