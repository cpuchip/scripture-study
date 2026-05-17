package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// mustWrite writes content to repo-root-relative rel under root,
// creating parent directories.
func mustWrite(t *testing.T, root, rel, content string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// ES.5.s1: fs_search must not descend into excluded dirs (node_modules
// etc.) even when one sits nested under an allowed root.
func TestFsSearchSkipsExcludedDirs(t *testing.T) {
	tmp := t.TempDir()
	mustWrite(t, tmp, "docs/real.md", "the NEEDLE token is here")
	mustWrite(t, tmp, "docs/node_modules/pkg/index.md", "the NEEDLE token is here too")
	mustWrite(t, tmp, "docs/target/out.md", "another NEEDLE in build output")

	sb := &sandbox{repoRoot: tmp, allowedGlobs: []string{"docs/**"}, maxReadBytes: 1 << 20}
	fn := makeFsSearch(sb)
	_, out, err := fn(context.Background(), nil, FsSearchInput{Pattern: "NEEDLE"})
	if err != nil {
		t.Fatalf("fs_search returned error: %v", err)
	}
	if out.Count == 0 {
		t.Fatal("expected a match in docs/real.md, got none")
	}
	for _, m := range out.Matches {
		if strings.Contains(m.Path, "node_modules") || strings.Contains(m.Path, "/target/") {
			t.Errorf("fs_search descended into an excluded dir: %s", m.Path)
		}
	}
}

// ES.5.s1: fs_search must honor a cancelled context — return promptly
// instead of churning past the bridge deadline (which invalidated the
// fs-read session in the ES.4 run).
func TestFsSearchHonorsContext(t *testing.T) {
	tmp := t.TempDir()
	for i := 0; i < 80; i++ {
		mustWrite(t, tmp, fmt.Sprintf("docs/f%d.md", i), "needle here\nand more lines\n")
	}
	sb := &sandbox{repoRoot: tmp, allowedGlobs: []string{"docs/**"}, maxReadBytes: 1 << 20}
	fn := makeFsSearch(sb)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled — fs_search must not hang

	done := make(chan struct{})
	go func() {
		_, _, _ = fn(ctx, nil, FsSearchInput{Pattern: "needle"})
		close(done)
	}()
	select {
	case <-done:
		// returned promptly — good
	case <-time.After(5 * time.Second):
		t.Fatal("fs_search did not honor a cancelled context — still running after 5s")
	}
}

// ES.5.s1: a normal search still works and finds matches.
func TestFsSearchFindsMatches(t *testing.T) {
	tmp := t.TempDir()
	mustWrite(t, tmp, "docs/a.md", "first line\nthe ANSWER is 42\nlast line")
	mustWrite(t, tmp, ".mind/b.md", "no match here")

	sb := &sandbox{
		repoRoot:     tmp,
		allowedGlobs: []string{"docs/**", ".mind/*"},
		maxReadBytes: 1 << 20,
	}
	fn := makeFsSearch(sb)
	_, out, err := fn(context.Background(), nil, FsSearchInput{Pattern: "ANSWER"})
	if err != nil {
		t.Fatalf("fs_search error: %v", err)
	}
	if out.Count != 1 {
		t.Fatalf("want 1 match, got %d", out.Count)
	}
	if out.Matches[0].Path != "docs/a.md" || out.Matches[0].Line != 2 {
		t.Errorf("match at wrong location: %+v", out.Matches[0])
	}
	if out.Truncated {
		t.Error("a small search should not report Truncated")
	}
}
