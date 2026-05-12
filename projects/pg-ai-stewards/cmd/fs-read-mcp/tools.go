// Tool handlers for fs-read-mcp.
//
// Three tools: fs_list, fs_read, fs_search. Each wraps a small,
// well-understood stdlib operation behind the sandbox layer.

package main

import (
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func toolError(format string, args ...any) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(format, args...)},
		},
	}
}

func registerTools(srv *mcp.Server, sb *sandbox) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "fs_list",
		Description: "List repo files matching a glob pattern, scoped to the sandbox's " +
			"allowed paths. Patterns are repo-root-relative (e.g., '.spec/journal/*.md', " +
			"'docs/**'). Returns up to 'limit' matching paths sorted lexicographically. " +
			"Use fs_search to find files by content; use fs_list when you know the rough " +
			"directory shape.",
	}, makeFsList(sb))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "fs_read",
		Description: "Read a file from the sandboxed paths. Inputs: 'path' is " +
			"repo-root-relative (e.g., '.spec/proposals/foo.md'). Optional 'max_bytes' " +
			"caps the response (server-side hard cap also applies). Returns the file " +
			"text with frontmatter intact. Reject if path is outside the allow-list or " +
			"the file does not exist.",
	}, makeFsRead(sb))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "fs_search",
		Description: "Regex-search content across files matching an optional glob, scoped " +
			"to the sandbox. Inputs: 'pattern' is a Go regexp. Optional 'path_glob' " +
			"narrows the search to specific files (default: all allowed paths). Returns " +
			"matches with file path, line number, and the matching line. Use to find " +
			"prior work mentioning a topic by name.",
	}, makeFsSearch(sb))
}

// ---------------------------------------------------------------------
// fs_list
// ---------------------------------------------------------------------

type FsListInput struct {
	Glob  string `json:"glob" jsonschema:"glob pattern, repo-root-relative; supports ** for recursive match"`
	Limit int    `json:"limit,omitempty" jsonschema:"max paths to return, default 100, hard cap 500"`
}

type FsListOutput struct {
	Paths []string `json:"paths"`
	Count int      `json:"count"`
}

func makeFsList(sb *sandbox) func(
	ctx context.Context, req *mcp.CallToolRequest, in FsListInput,
) (*mcp.CallToolResult, FsListOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in FsListInput,
	) (*mcp.CallToolResult, FsListOutput, error) {
		if in.Glob == "" {
			return toolError("fs_list: 'glob' is required"), FsListOutput{}, nil
		}
		if in.Limit <= 0 {
			in.Limit = 100
		}
		if in.Limit > 500 {
			in.Limit = 500
		}

		// Reject globs that aren't themselves inside the allow-list.
		// The glob itself doesn't have to literally match the
		// allowed glob (different syntax) — instead we walk and
		// filter at result time, but we sanity-check the glob's
		// directory prefix against the sandbox so we don't traverse
		// the whole filesystem.
		clean := filepath.ToSlash(filepath.Clean(in.Glob))
		if strings.HasPrefix(clean, "/") || strings.HasPrefix(clean, "../") {
			return toolError("fs_list: glob must be repo-root-relative and not escape root: %q", in.Glob),
				FsListOutput{}, nil
		}

		// Walk the directory tree rooted at repo-root, applying the
		// glob and then re-validating each match against the
		// allow-list. The walk is bounded by visiting only paths
		// whose prefix could potentially match the glob's prefix.
		var matches []string
		walkRoot := globWalkRoot(sb.repoRoot, clean)
		err := filepath.WalkDir(walkRoot, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				// Permission errors etc. — skip the entry, continue walking.
				return nil
			}
			if d.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(sb.repoRoot, path)
			if err != nil {
				return nil
			}
			rel = filepath.ToSlash(rel)
			if !matchGlob(clean, rel) {
				return nil
			}
			// Final check: result must sit inside the allow-list.
			if !sb.matchesAllowed(rel) {
				return nil
			}
			matches = append(matches, rel)
			if len(matches) >= in.Limit*2 {
				// Soft early-exit; sort+truncate happens below.
				return filepath.SkipAll
			}
			return nil
		})
		if err != nil {
			return toolError("fs_list: walk error: %v", err), FsListOutput{}, nil
		}

		sort.Strings(matches)
		if len(matches) > in.Limit {
			matches = matches[:in.Limit]
		}
		return nil, FsListOutput{Paths: matches, Count: len(matches)}, nil
	}
}

// globWalkRoot returns the deepest directory under repo-root that is
// guaranteed to contain all matches for the given glob. For
// ".spec/journal/*.md" this is "<repoRoot>/.spec/journal". This trims
// the walk so we don't traverse the whole repo for a tightly-scoped
// pattern.
func globWalkRoot(repoRoot, glob string) string {
	parts := strings.Split(glob, "/")
	var prefixParts []string
	for _, p := range parts {
		if strings.ContainsAny(p, "*?[") {
			break
		}
		prefixParts = append(prefixParts, p)
	}
	if len(prefixParts) == 0 {
		return repoRoot
	}
	return filepath.Join(repoRoot, filepath.Join(prefixParts...))
}

// ---------------------------------------------------------------------
// fs_read
// ---------------------------------------------------------------------

type FsReadInput struct {
	Path     string `json:"path" jsonschema:"repo-root-relative path to read"`
	MaxBytes int    `json:"max_bytes,omitempty" jsonschema:"per-call cap on response size; server-side hard cap also applies"`
}

type FsReadOutput struct {
	Path      string `json:"path"`
	Content   string `json:"content"`
	Bytes     int    `json:"bytes"`
	Truncated bool   `json:"truncated"`
}

func makeFsRead(sb *sandbox) func(
	ctx context.Context, req *mcp.CallToolRequest, in FsReadInput,
) (*mcp.CallToolResult, FsReadOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in FsReadInput,
	) (*mcp.CallToolResult, FsReadOutput, error) {
		if in.Path == "" {
			return toolError("fs_read: 'path' is required"), FsReadOutput{}, nil
		}
		absPath, relPath, err := sb.resolvePath(in.Path)
		if err != nil {
			return toolError("fs_read: %v", err), FsReadOutput{}, nil
		}

		info, err := os.Stat(absPath)
		if err != nil {
			return toolError("fs_read: %v", err), FsReadOutput{}, nil
		}
		if info.IsDir() {
			return toolError("fs_read: path is a directory, use fs_list instead: %q", relPath),
				FsReadOutput{}, nil
		}

		cap := sb.maxReadBytes
		if in.MaxBytes > 0 && in.MaxBytes < cap {
			cap = in.MaxBytes
		}

		data, err := os.ReadFile(absPath)
		if err != nil {
			return toolError("fs_read: %v", err), FsReadOutput{}, nil
		}
		truncated := false
		if len(data) > cap {
			data = data[:cap]
			truncated = true
		}

		return nil, FsReadOutput{
			Path:      relPath,
			Content:   string(data),
			Bytes:     len(data),
			Truncated: truncated,
		}, nil
	}
}

// ---------------------------------------------------------------------
// fs_search
// ---------------------------------------------------------------------

type FsSearchInput struct {
	Pattern      string `json:"pattern" jsonschema:"Go regexp pattern to search for"`
	PathGlob     string `json:"path_glob,omitempty" jsonschema:"optional glob to narrow which files are searched"`
	Limit        int    `json:"limit,omitempty" jsonschema:"max matches to return, default 50, hard cap 200"`
	CaseInsense  bool   `json:"case_insensitive,omitempty" jsonschema:"if true, prepend (?i) to pattern"`
	MaxFileBytes int    `json:"max_file_bytes,omitempty" jsonschema:"skip files larger than this in bytes, default 1MB"`
}

type FsSearchHit struct {
	Path    string `json:"path"`
	Line    int    `json:"line"`
	Content string `json:"content"`
}

type FsSearchOutput struct {
	Matches []FsSearchHit `json:"matches"`
	Count   int           `json:"count"`
}

func makeFsSearch(sb *sandbox) func(
	ctx context.Context, req *mcp.CallToolRequest, in FsSearchInput,
) (*mcp.CallToolResult, FsSearchOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in FsSearchInput,
	) (*mcp.CallToolResult, FsSearchOutput, error) {
		if in.Pattern == "" {
			return toolError("fs_search: 'pattern' is required"), FsSearchOutput{}, nil
		}
		if in.Limit <= 0 {
			in.Limit = 50
		}
		if in.Limit > 200 {
			in.Limit = 200
		}
		if in.MaxFileBytes <= 0 {
			in.MaxFileBytes = 1024 * 1024
		}

		pat := in.Pattern
		if in.CaseInsense {
			pat = "(?i)" + pat
		}
		re, err := regexp.Compile(pat)
		if err != nil {
			return toolError("fs_search: invalid regexp: %v", err), FsSearchOutput{}, nil
		}

		// Build the file list to search. If PathGlob is set, filter
		// to that; else walk all allowed paths.
		var files []string
		if in.PathGlob != "" {
			pg := filepath.ToSlash(filepath.Clean(in.PathGlob))
			files = filesMatchingGlob(sb, pg)
		} else {
			// Walk all files under any allowed glob's prefix.
			seen := make(map[string]struct{})
			for _, g := range sb.allowedGlobs {
				for _, f := range filesMatchingGlob(sb, g) {
					if _, ok := seen[f]; !ok {
						seen[f] = struct{}{}
						files = append(files, f)
					}
				}
			}
		}
		sort.Strings(files)

		var matches []FsSearchHit
	OUTER:
		for _, rel := range files {
			abs := filepath.Join(sb.repoRoot, rel)
			info, err := os.Stat(abs)
			if err != nil || info.IsDir() {
				continue
			}
			if info.Size() > int64(in.MaxFileBytes) {
				continue
			}
			f, err := os.Open(abs)
			if err != nil {
				continue
			}
			sc := bufio.NewScanner(f)
			// Allow longer lines than default 64KB for log-like files.
			sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
			lineNo := 0
			for sc.Scan() {
				lineNo++
				line := sc.Text()
				if !re.MatchString(line) {
					continue
				}
				// Trim very long lines for response economy.
				if len(line) > 400 {
					line = line[:400] + "…"
				}
				matches = append(matches, FsSearchHit{
					Path:    rel,
					Line:    lineNo,
					Content: line,
				})
				if len(matches) >= in.Limit {
					f.Close()
					break OUTER
				}
			}
			f.Close()
		}

		return nil, FsSearchOutput{Matches: matches, Count: len(matches)}, nil
	}
}

// filesMatchingGlob returns the repo-root-relative file paths under
// sb.repoRoot that match the given glob AND sit inside the sandbox's
// allow-list. The glob may itself be one of the allow-list patterns
// or a tighter user-provided narrowing.
func filesMatchingGlob(sb *sandbox, glob string) []string {
	walkRoot := globWalkRoot(sb.repoRoot, glob)
	var out []string
	_ = filepath.WalkDir(walkRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(sb.repoRoot, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if !matchGlob(glob, rel) {
			return nil
		}
		if !sb.matchesAllowed(rel) {
			return nil
		}
		out = append(out, rel)
		return nil
	})
	return out
}
